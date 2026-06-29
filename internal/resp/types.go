package resp

type RespValue interface {
	respValue()
}

type BulkString struct {
	Value []byte
}

type Array struct {
	Value []RespValue
}

func (Array) respValue()      {}
func (BulkString) respValue() {}
