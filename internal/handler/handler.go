// Package handler handles commands and executions of resp
package handler

import (
	"bytes"
	"fmt"
	"io"

	"github.com/Arush71/redis-server/internal/helpers"
	"github.com/Arush71/redis-server/internal/storage"
)

var nullBulkStr = []byte("$-1\r\n")

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
			return writeToConn(nullBulkStr, connWrite)
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
		start, ok := helpers.ParseInt(data[2])
		stop, ok2 := helpers.ParseInt(data[3])
		if !ok || !ok2 {
			return writeErrorToConn(helpers.ErrNotInt, connWrite)
		}
		list, err := storage.LRange(string(data[1]), start, stop)
		if err != nil {
			return writeErrorToConn(err, connWrite)
		}
		return writeBulkArray(list, connWrite)
	case "LPUSH":
		if len(data) < 3 {
			return writeErrorToConn(helpers.ErrWrongArgCount, connWrite)
		}
		response, err := storage.LPush(string(data[1]), data[2:]...)
		if err != nil {
			return writeErrorToConn(err, connWrite)
		}
		return writeInteger(response, connWrite)

	case "LLEN":
		if len(data) != 2 {
			return writeErrorToConn(helpers.ErrWrongArgCount, connWrite)
		}
		response, err := storage.LLEN(string(data[1]))
		if err != nil {
			return writeErrorToConn(err, connWrite)
		}
		return writeInteger(response, connWrite)

	case "LPOP":
		if len(data) < 2 || len(data) > 3 {
			return writeErrorToConn(helpers.ErrWrongArgCount, connWrite)
		}
		var count *int64
		if len(data) == 3 {
			num, ok := helpers.ParsePositiveInt(data[2])
			if !ok {
				return writeErrorToConn(helpers.ErrNotPosInt, connWrite)
			}
			count = &num
		}
		response, ok, err := storage.LPOP(string(data[1]), count)
		if err != nil {
			return writeErrorToConn(err, connWrite)
		}
		if !ok {
			return writeToConn(nullBulkStr, connWrite)
		}
		if count != nil {
			return writeBulkArray(response, connWrite)
		}
		return writeBulk(response[0], connWrite)
	case "BLPOP":
		if len(data) < 3 {
			return writeErrorToConn(helpers.ErrWrongArgCount, connWrite)
		}
		timeout, err := helpers.ParsePositiveFloat(data[len(data)-1])
		if err != nil {
			return writeErrorToConn(err, connWrite)
		}
		result, err := storage.BLPOP(timeout, data[1:len(data)-1])
		if err == nil && result == nil {
			return writeNullArr(connWrite)
		}
		if err != nil {
			return writeErrorToConn(err, connWrite)
		}
		return writeBulkArray(result, connWrite)
	default:
		return writeToConn(fmt.Appendf(nil, "-Err unknown command '%s'\r\n", command), connWrite)
	}
}
