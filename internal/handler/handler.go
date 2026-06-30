// Package handler handles commands and executions of resp
package handler

import (
	"bytes"
	"fmt"
	"io"

	"github.com/Arush71/redis-server/internal/helpers"
	"github.com/Arush71/redis-server/internal/storage"
)

func HandleReqData(data [][]byte, connWrite io.Writer, storage *storage.Storage) error {
	command := string(bytes.ToUpper(data[0]))
	switch command {
	case "PING":
		if len(data) != 1 {
			return writeErrorToConn(helpers.ErrWrongArgCount, connWrite)
		}
		return writeToConn([]byte("+PONG\r\n"), connWrite)
	case "ECHO":
		if len(data) != 2 {
			return writeErrorToConn(helpers.ErrWrongArgCount, connWrite)
		}
		fmt.Println("the data is", string(data[1]))
		return writeBulk(data[1], connWrite)
	case "SET":
		if len(data) != 3 && len(data) != 5 {
			return writeErrorToConn(helpers.ErrWrongArgCount, connWrite)
		}
		var px int64 = -1
		if len(data) == 5 {
			var ok bool
			px, ok = helpers.ParsePositiveInt(data[4])
			if !bytes.EqualFold(data[3], []byte("PX")) || !ok {
				return writeErrorToConn(helpers.ErrWrongArgCount, connWrite)
			}
		}
		storage.Set(string(data[1]), data[2], px)
		return writeToConn([]byte("+OK\r\n"), connWrite)
	case "GET":
		if len(data) != 2 {
			return writeErrorToConn(helpers.ErrWrongArgCount, connWrite)
		}
		value, ok := storage.Get(string(data[1]))
		if !ok {
			return writeToConn([]byte("$-1\r\n"), connWrite)
		}
		return writeBulk(value, connWrite)
	case "RPUSH":
		if len(data) < 3 {
			return writeErrorToConn(helpers.ErrWrongArgCount, connWrite)
		}
		response, err := storage.RPush(string(data[1]), data[2:]...)
		if err != nil {
			return writeErrorToConn(err, connWrite)
		}
		return writeInteger(response, connWrite)
	case "LRANGE":
		if len(data) != 4 {
			return writeErrorToConn(helpers.ErrWrongArgCount, connWrite)
		}
		list, err := storage.LRange(string(data[1]), data[2], data[3])
		if err != nil {
			return writeErrorToConn(err, connWrite)
		}
		fallthrough
	default:
		return writeToConn(fmt.Appendf(nil, "-Err unknown command '%s'\r\n", command), connWrite)
	}
}
