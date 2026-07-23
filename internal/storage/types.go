package storage

import (
	"errors"
	"sync"
	"time"
)

var ErrEOF = errors.New("eof while trying to execute the request")

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

type listnerManager struct {
	table listnerTable
	mu    sync.Mutex
}
type Storage struct {
	mu       sync.Mutex
	storage  storage
	listners *listnerManager
}

type (
	listenChType struct {
		key   string
		value []byte
	}
	listenCh chan listenChType
	waiter   struct {
		listenCh listenCh
		claimed  bool
	}
	listnerTable map[string][]*waiter
)

func (l *listnerManager) EnqueueListners(waiter *waiter, keys [][]byte) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i := range len(keys) {
		key := string(keys[i])
		l.table[key] = append(l.table[key], waiter)
	}
}

func (l *listnerManager) TryDequeueListners(key string, list *List) {
	l.mu.Lock()
	defer l.mu.Unlock()
	waiters := l.table[key]
	if len(waiters) == 0 {
		return
	}
	i := 0
	for ; i < len(waiters); i++ {
		if !waiters[i].claimed {
			data, ok := list.popFront()
			if !ok {
				break
			}
			waiters[i].listenCh <- listenChType{
				key:   key,
				value: data,
			}
			waiters[i].claimed = true
		}
		waiters[i] = nil
	}
	waiters = waiters[i:]
	if len(waiters) == 0 {
		delete(l.table, key)
		return
	}
	l.table[key] = waiters
}

func (l *listnerManager) removeKeys(keys [][]byte, waiterToDel *waiter) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i := range len(keys) {
		key := string(keys[i])
		waiters, ok := l.table[key]
		if !ok {
			continue
		}
		for j := range waiters {
			if waiters[j] != waiterToDel {
				continue
			}
			if j < len(waiters)-1 {
				copy(waiters[j:], waiters[j+1:])
			}
			waiters[len(waiters)-1] = nil
			waiters = waiters[:len(waiters)-1]
			break
		}
		if len(waiters) == 0 {
			delete(l.table, key)
			continue
		}
		l.table[key] = waiters
	}
}

func (l *List) popFront() ([]byte, bool) {
	if len(l.value) == 0 {
		return nil, false
	}
	item := l.value[0]
	l.value[0], l.value = nil, l.value[1:]
	return item, true
}
