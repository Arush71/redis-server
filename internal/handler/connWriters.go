package handler

import (
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

func writeSimpleStr(data string, connWrite io.Writer) error {
	buf := make([]byte, 0, len(data)+3)
	buf = append(buf, '+')
	buf = append(buf, []byte(data)...)
	buf = append(buf, '\r', '\n')
	return writeToConn(buf, connWrite)
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
	if err == storage.ErrEOF {
		return err
	}
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

var (
	nullArr  = []byte("*-1\r\n")
	emptyArr = []byte("*0\r\n")
)

func writeNullArr(connWrite io.Writer) error {
	return writeToConn(nullArr, connWrite)
}

func writeEmptyArr(connWrite io.Writer) error {
	return writeToConn(emptyArr, connWrite)
}

func writeXrange(connWrite io.Writer, data []storage.StreamEntry) error {
	buf := make([]byte, 0, len(data)+24)
	buf = append(buf, '*')
	buf = append(buf, strconv.Itoa(len(data))...)
	buf = append(buf, '\r', '\n')
	for _, v := range data {
		buf = append(buf, '*', '2', '\r', '\n')
		buf = append(buf, '$')
		idStr := fmt.Sprintf("%d-%d", v.Id.Time, v.Id.Seq)
		buf = append(buf, strconv.Itoa(len(idStr))...)
		buf = append(buf, '\r', '\n')
		buf = append(buf, idStr...)
		buf = append(buf, '\r', '\n')
		buf = append(buf, '*')
		buf = append(buf, strconv.Itoa(len(v.StreamStorage)*2)...)
		buf = append(buf, '\r', '\n')
		// loop
		for _, v2 := range v.StreamStorage {
			buf = append(buf, '$')
			buf = append(buf, strconv.Itoa(len(v2.Field))...)
			buf = append(buf, '\r', '\n')
			buf = append(buf, v2.Field...)
			buf = append(buf, '\r', '\n')
			// value
			buf = append(buf, '$')
			buf = append(buf, strconv.Itoa(len(v2.Value))...)
			buf = append(buf, '\r', '\n')
			buf = append(buf, v2.Value...)
			buf = append(buf, '\r', '\n')
		}
	}
	return writeToConn(buf, connWrite)
}

func writeXread(connWrite io.Writer, data []storage.StreamReadResults) error {
	buf := make([]byte, 0, len(data)+24)
	buf = append(buf, '*')
	buf = append(buf, strconv.Itoa(len(data))...)
	buf = append(buf, '\r', '\n')
	for _, v := range data {
		buf = append(buf, '*', '2', '\r', '\n')
		// firt element
		buf = append(buf, '$')
		buf = append(buf, strconv.Itoa(len(v.StreamKey))...)
		buf = append(buf, '\r', '\n')
		buf = append(buf, v.StreamKey...)
		buf = append(buf, '\r', '\n')
		// end of first element

		// 2nd element
		buf = append(buf, '*', '1', '\r', '\n')
		// end of 2nd element

		// 2nd element's element
		buf = append(buf, '*', '2', '\r', '\n')
		// end of 2nd element's element

		// 2 elemetnt's first
		// TODO: finish xread resp parsing.
		buf = append(buf, '$')
		idStr := fmt.Sprintf("%d-%d", v.Id.Time, v.Id.Seq)
		buf = append(buf, strconv.Itoa(len(idStr))...)
		buf = append(buf, '\r', '\n')
		buf = append(buf, idStr...)
		buf = append(buf, '\r', '\n')
		buf = append(buf, '*')
		buf = append(buf, strconv.Itoa(len(v.StreamStorage)*2)...)
		buf = append(buf, '\r', '\n')
		// loop
		for _, v2 := range v.StreamStorage {
			buf = append(buf, '$')
			buf = append(buf, strconv.Itoa(len(v2.Field))...)
			buf = append(buf, '\r', '\n')
			buf = append(buf, v2.Field...)
			buf = append(buf, '\r', '\n')
			// value
			buf = append(buf, '$')
			buf = append(buf, strconv.Itoa(len(v2.Value))...)
			buf = append(buf, '\r', '\n')
			buf = append(buf, v2.Value...)
			buf = append(buf, '\r', '\n')
		}
	}
	return writeToConn(buf, connWrite)
}
