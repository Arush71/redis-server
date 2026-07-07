package handler

import (
	"fmt"
	"io"
	"strconv"

	"github.com/Arush71/redis-server/internal/helpers"
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

func writeErrorToConn(err error, connWrite io.Writer) error {
	if err == nil {
		return nil
	}
	if err, ok := err.(*helpers.RespError); ok {
		return writeToConn(err.RespError(), connWrite)
	}
	return writeToConn(fmt.Appendf(nil, "-Err %s\r\n", err.Error()), connWrite)
}

func writeInteger(value int, connWrite io.Writer) error {
	buf := make([]byte, 0, 12)
	buf = append(buf, ':')
	buf = append(buf, strconv.Itoa(value)...)
	buf = append(buf, '\r', '\n')
	return writeToConn(buf, connWrite)
}

func writeBulkArray(data [][]byte, connWrite io.Writer) error {
	buf := make([]byte, 0, 64+len(data)*8)
	buf = append(buf, '*')
	buf = append(buf, strconv.Itoa(len(data))...)
	buf = append(buf, '\r', '\n')
	for _, v := range data {
		buf = append(buf, '$')
		buf = append(buf, strconv.Itoa(len(v))...)
		buf = append(buf, '\r', '\n')
		buf = append(buf, v...)
		buf = append(buf, '\r', '\n')
	}
	return writeToConn(buf, connWrite)
}
