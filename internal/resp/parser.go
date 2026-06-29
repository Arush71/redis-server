// Package resp is responsible for adhering to the RESP protocol
package resp

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log/slog"

	"github.com/Arush71/redis-server/internal/helpers"
)

var (
	CRLF    = []byte("\r\n")
	crlfLen = int64(len(CRLF))
)

var ErrProtocol = errors.New("protocol error")

func Parse(bufReader *bufio.Reader, logger *slog.Logger) (RespValue, error) {
	logger.Debug("entering parsing now")
	readBytes, err := bufReader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	logger.Debug("got no err on the read till \n")
	readBytes, found := bytes.CutSuffix(readBytes, CRLF)
	if !found || len(readBytes) < 1 {
		return nil, ErrProtocol
	}
	typeOfByte := readBytes[0]
	readBytes = readBytes[1:]
	switch typeOfByte {
	case '$':
		bytesToRead, ok := helpers.ParsePositiveInt(readBytes)
		if !ok {
			return nil, ErrProtocol
		}
		data := make([]byte, bytesToRead+crlfLen)
		_, err := io.ReadFull(bufReader, data)
		if err != nil {
			return nil, err
		}
		logger.Debug("got no err on the read full of bulk")
		data, found = bytes.CutSuffix(data, CRLF)
		if !found {
			return nil, ErrProtocol
		}
		return BulkString{Value: data}, nil
	case '*':
		logger.Debug("entering the array part")
		numOfelements, ok := helpers.ParsePositiveInt(readBytes)
		if !ok {
			return nil, ErrProtocol
		}
		logger.Debug("the array is gonna loop for", "loop", numOfelements)
		arr := make([]RespValue, 0, numOfelements)
		for range numOfelements {
			value, err := Parse(bufReader, logger)
			if err != nil {
				return Array{Value: arr}, err
			}
			arr = append(arr, value)
		}
		logger.Debug("len of items post loop", "len", len(arr))
		return Array{Value: arr}, nil
	default:
		return nil, ErrProtocol
	}
}
