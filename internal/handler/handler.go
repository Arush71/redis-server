// Package handler handles commands and executions of resp
package handler

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/Arush71/redis-server/internal/helpers"
	"github.com/Arush71/redis-server/internal/storage"
)

func writeToConn(data []byte, connWrite io.Writer) error {
	_, err := connWrite.Write(data)
	return err
}

func writeBulk(data []byte, connWrite io.Writer) error {
	buf := make([]byte, 0, 32)
	buf = append(buf, '$')
	buf = append(buf, strconv.Itoa(len(data))...)
	buf = append(buf, '\r', '\n')
	buf = append(buf, data...)
	buf = append(buf, '\r', '\n')
	return writeToConn(buf, connWrite)
}

var ErrWrongArgCount = []byte("-ERR wrong number of arguments\r\n")

func HandleReqData(data [][]byte, connWrite io.Writer, storage *storage.Storage) error {
	command := string(bytes.ToUpper(data[0]))
	switch command {
	case "PING":
		if len(data) != 1 {
			return writeToConn(ErrWrongArgCount, connWrite)
		}
		return writeToConn([]byte("+PONG\r\n"), connWrite)
	case "ECHO":
		if len(data) != 2 {
			return writeToConn(ErrWrongArgCount, connWrite)
		}
		fmt.Println("the data is", string(data[1]))
		return writeBulk(data[1], connWrite)
	case "SET":
		if len(data) != 3 && len(data) != 5 {
			return writeToConn(ErrWrongArgCount, connWrite)
		}
		var px int64 = -1
		if len(data) == 5 {
			var ok bool
			px, ok = helpers.ParsePositiveInt(data[4])
			if !bytes.EqualFold(data[3], []byte("PX")) || !ok {
				return writeToConn(ErrWrongArgCount, connWrite)
			}
		}
		storage.Set(string(data[1]), data[2], px)
		return writeToConn([]byte("+OK\r\n"), connWrite)
	case "GET":
		if len(data) != 2 {
			return writeToConn(ErrWrongArgCount, connWrite)
		}
		value, ok := storage.Get(string(data[1]))
		if !ok {
			return writeToConn([]byte("$-1\r\n"), connWrite)
		}
		return writeBulk(value, connWrite)
	default:
		return writeToConn(fmt.Appendf(nil, "-Err unknown command '%s'\r\n", command), connWrite)
	}
}
