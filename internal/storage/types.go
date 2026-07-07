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

type listner map[string]chan []byte

func (l *List) popFront() ([]byte, bool) {
	if len(l.value) == 0 {
		return nil, false
	}
	item := l.value[0]
	l.value[0], l.value = nil, l.value[1:]
	return item, true
}
