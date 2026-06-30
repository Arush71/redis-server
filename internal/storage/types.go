package storage

import "time"

type redisValue interface {
	Value()
}

type String struct {
	value []byte
}

type List struct {
	value [][]byte
}

func (List) Value()   {}
func (String) Value() {}

type item struct {
	value     redisValue
	expiresAt *time.Time
}

type storage map[string]item
