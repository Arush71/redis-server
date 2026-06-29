// Package server is the server
package server

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log/slog"
	"net"

	"github.com/Arush71/redis-server/internal/handler"
	"github.com/Arush71/redis-server/internal/resp"
	"github.com/Arush71/redis-server/internal/storage"
)

type Server struct {
	logger     *slog.Logger
	listenAddr string
	ln         net.Listener
	ctx        context.Context
	cancel     context.CancelFunc
	storage    *storage.Storage
}

func NewServer(listenAddr string, logger *slog.Logger, storage *storage.Storage) *Server {
	s := &Server{
		listenAddr: listenAddr,
		logger:     logger,
		storage:    storage,
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()
	s.ln = ln
	go s.acceptLoop()
	<-s.ctx.Done()
	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			s.logger.Error("accept connection error", "err", err)
			continue
		}
		go s.manageConnection(conn)
	}
}

func (s *Server) manageConnection(conn net.Conn) {
	defer conn.Close()
	bufReader := bufio.NewReader(conn)
	for {
		s.logger.Debug("entering the loop")
		value, err := resp.Parse(bufReader, s.logger)
		if err != nil {
			if errors.Is(err, io.EOF) {
				s.logger.Debug("connection closed on eof")
				return
			}
			if errors.Is(err, resp.ErrProtocol) {
				s.logger.Debug("protocol err")
				return
			}
			s.logger.Error("unknown err on parse", "error", err)
			return
		}
		arr, ok := value.(resp.Array)
		if !ok {
			s.logger.Error("protocol errror on wrong array")
			return
		}
		data := make([][]byte, 0, len(arr.Value))
		for _, v := range arr.Value {
			bulkStr, ok := v.(resp.BulkString)
			if !ok {
				s.logger.Error("protocol errror on wrong array value not bulk")
				return
			}
			data = append(data, bulkStr.Value)
		}
		if err := handler.HandleReqData(data, conn, s.storage); err != nil {
			s.logger.Error("handler error, failed to write")
			return
		}
	}
}
