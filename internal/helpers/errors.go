package helpers

type RespError struct {
	msg string
}

func (e *RespError) Error() string {
	return e.msg
}

func (e *RespError) RespError() []byte {
	return []byte("-" + e.msg + "\r\n")
}

var (
	ErrWrongArgCount = &RespError{msg: "ERR wrong number of arguments"}
	ErrWrongType     = &RespError{msg: "WRONGTYPE Operation against a key holding the wrong kind of value"}
)
