package server

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"

	"github.com/Arush71/redis-server/internal/handler"
	"github.com/Arush71/redis-server/internal/resp"
)

type commandCh chan [][]byte

func (c *Server) manageRead(conn net.Conn, ctx context.Context, cancel context.CancelFunc, cmdCh commandCh) {
	bufReader := bufio.NewReader(conn)
	defer cancel()
	for {
		value, err := resp.Parse(bufReader, c.logger)
		if err != nil {
			if errors.Is(err, io.EOF) {
				c.logger.Debug("connection closed on eof")
				return
			}
			if errors.Is(err, resp.ErrProtocol) {
				c.logger.Debug("protocol err")
				return
			}
			c.logger.Error("unknown err on parse", "error", err)
			return
		}
		arr, ok := value.(resp.Array)
		if !ok {
			c.logger.Error("protocol errror on wrong array")
			return
		}
		data := make([][]byte, 0, len(arr.Value))
		for _, v := range arr.Value {
			bulkStr, ok := v.(resp.BulkString)
			if !ok {
				c.logger.Error("protocol errror on wrong array value not bulk")
				return
			}
			data = append(data, bulkStr.Value)
		}
		select {
		case cmdCh <- data:
		case <-ctx.Done():
			return
		}
	}
}

func (s *Server) manageConnection(conn net.Conn) {
	defer conn.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cmdCh := make(commandCh, 8)
	go s.manageRead(conn, ctx, cancel, cmdCh)
	for {
		select {
		case data := <-cmdCh:
			if err := handler.HandleReqData(data, conn, s.storage, s.logger, ctx); err != nil {
				s.logger.Error("handler error, failed to write")
				cancel()
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
