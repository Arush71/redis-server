package resp

// In future for parse visit:https: github.com/tidwall/redcon/blob/bc9875b4b00f22ed2575ea7602ca83d3eeef88c8/resp.go
import (
	"bufio"
	"bytes"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data         []byte
	numBytesRead int
	pos          int
}

func (cr *chunkReader) Read(p []byte) (int, error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := min(cr.pos+cr.numBytesRead, len(cr.data))
	n := copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n
	return n, nil
}

func MustBulkStringArray(t *testing.T, value RespValue) [][]byte {
	require.IsType(t, Array{}, value)
	arr := value.(Array)
	data := make([][]byte, 0, len(arr.Value))
	for _, v := range arr.Value {
		require.IsType(t, BulkString{}, v)
		bulkStr := v.(BulkString)
		data = append(data, bulkStr.Value)
	}
	return data
}

func TestParse(t *testing.T) {
	loger := slog.Default()
	t.Run("echo command", func(t *testing.T) {
		data := []byte("*2\r\n$4\r\necho\r\n$11\r\nhello World\r\n")
		expected := [][]byte{
			[]byte("echo"),
			[]byte("hello World"),
		}
		// one shot read
		a := assert.New(t)
		r := require.New(t)
		bufRead := bufio.NewReader(bytes.NewReader(data))
		respVal, err := Parse(bufRead, loger)
		r.NoError(err)
		got := MustBulkStringArray(t, respVal)
		a.Equal(expected, got)

		// Test: chunk reading
		bufRead = bufio.NewReader(&chunkReader{
			data:         data,
			numBytesRead: 1,
		})
		respVal, err = Parse(bufRead, loger)
		r.NoError(err)
		got = MustBulkStringArray(t, respVal)
		a.Equal(expected, got)
	})

	t.Run("ping command", func(t *testing.T) {
		data := []byte("*1\r\n$4\r\nping\r\n")
		expected := [][]byte{
			[]byte("ping"),
		}
		// one shot read
		a := assert.New(t)
		r := require.New(t)
		bufRead := bufio.NewReader(bytes.NewReader(data))
		respVal, err := Parse(bufRead, loger)
		r.NoError(err)
		got := MustBulkStringArray(t, respVal)
		a.Equal(expected, got)

		// Test: chunk reading
		bufRead = bufio.NewReader(&chunkReader{
			data:         data,
			numBytesRead: 1,
		})
		respVal, err = Parse(bufRead, loger)
		r.NoError(err)
		got = MustBulkStringArray(t, respVal)
		a.Equal(expected, got)
	})

	t.Run("empty array", func(t *testing.T) {
		data := []byte("*0\r\n")
		expected := Array{Value: []RespValue{}}
		// one shot read
		a := assert.New(t)
		r := require.New(t)
		bufRead := bufio.NewReader(bytes.NewReader(data))
		respVal, err := Parse(bufRead, loger)
		r.NoError(err)
		a.Equal(expected, respVal)

		// Test: chunk reading
		bufRead = bufio.NewReader(&chunkReader{
			data:         data,
			numBytesRead: 1,
		})
		respVal, err = Parse(bufRead, loger)
		r.NoError(err)
		a.Equal(expected, respVal)
	})

	t.Run("invalid array", func(t *testing.T) {
		data := []byte("*abc\r\n")
		// one shot read
		r := require.New(t)
		bufRead := bufio.NewReader(bytes.NewReader(data))
		_, err := Parse(bufRead, loger)
		r.ErrorIs(err, ErrProtocol)

		// Test: chunk reading
		bufRead = bufio.NewReader(&chunkReader{
			data:         data,
			numBytesRead: 1,
		})
		_, err = Parse(bufRead, loger)
		r.ErrorIs(err, ErrProtocol)
	})
	t.Run("invalid bulk length", func(t *testing.T) {
		data := []byte("*1\r\n$abc\r\nhello\r\n")
		// one shot read
		r := require.New(t)
		bufRead := bufio.NewReader(bytes.NewReader(data))
		_, err := Parse(bufRead, loger)
		r.ErrorIs(err, ErrProtocol)

		// Test: chunk reading
		bufRead = bufio.NewReader(&chunkReader{
			data:         data,
			numBytesRead: 1,
		})
		_, err = Parse(bufRead, loger)
		r.ErrorIs(err, ErrProtocol)
	})

	t.Run("missing crlf", func(t *testing.T) {
		data := []byte("*1\r\n$4\nPing\n")
		// one shot read
		r := require.New(t)
		bufRead := bufio.NewReader(bytes.NewReader(data))
		_, err := Parse(bufRead, loger)
		r.ErrorIs(err, ErrProtocol)

		// Test: chunk reading
		bufRead = bufio.NewReader(&chunkReader{
			data:         data,
			numBytesRead: 1,
		})
		_, err = Parse(bufRead, loger)
		r.ErrorIs(err, ErrProtocol)
	})

	t.Run("unexpected eof", func(t *testing.T) {
		data := []byte("*2\r\n$5\r\nhello\r\n$5\r\nwo")
		// one shot read
		r := require.New(t)
		bufRead := bufio.NewReader(bytes.NewReader(data))
		_, err := Parse(bufRead, loger)
		r.ErrorIs(err, io.ErrUnexpectedEOF)

		// Test: chunk reading
		bufRead = bufio.NewReader(&chunkReader{
			data:         data,
			numBytesRead: 1,
		})
		_, err = Parse(bufRead, loger)
		r.ErrorIs(err, io.ErrUnexpectedEOF)
	})
}
