// Package storage is responsible for in-memory storage
package storage

import (
	"sync"
	"time"

	"github.com/Arush71/redis-server/internal/helpers"
)

type Storage struct {
	mu       sync.Mutex
	storage  storage
	listners listner
}

var emptyList = [][]byte{}

func InitStorage() *Storage {
	return &Storage{
		storage:  make(storage),
		mu:       sync.Mutex{},
		listners: make(listner),
	}
}

func (s *Storage) Set(key string, val []byte, px int64) {
	var expiresAt *time.Time
	if px > 0 {
		t := time.Now().Add(time.Duration(px) * time.Millisecond)
		expiresAt = &t
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.storage[key] = item{
		value:     String{value: val},
		expiresAt: expiresAt,
	}
}

func (s *Storage) Get(key string) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	it, ok := s.storage[key]
	if ok && it.expiresAt != nil && time.Now().After(*it.expiresAt) {
		delete(s.storage, key)
		return nil, false
	}
	str, ok := it.value.(String)
	if !ok {
		return nil, false
	}
	return helpers.CopyBytes(str.value), ok
}

func (s *Storage) RPush(key string, value ...[]byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	it, ok := s.storage[key]
	var arr List
	if ok {
		arr, ok = it.value.(List)
		if !ok {
			return 0, helpers.ErrWrongType
		}
	} else {
		it = item{
			expiresAt: nil,
		}
	}
	arr.value = append(arr.value, value...)
	maxLen := len(arr.value)
	s.handleListAppendLocked(arr, key, it)
	return maxLen, nil
}

func (s *Storage) LRange(key string, start int64, stop int64) ([][]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	it, ok := s.storage[key]
	if !ok {
		return emptyList, nil
	}
	list, ok := it.value.(List)
	if !ok {
		return nil, helpers.ErrWrongType
	}
	lenList := int64(len(list.value))
	if stop < 0 {
		stop += lenList
		if stop < 0 {
			return emptyList, nil
		}
	}
	if start < 0 {
		start += lenList
		if start < 0 {
			start = 0
		}
	}
	if start >= lenList || start > stop {
		return emptyList, nil
	}
	if stop >= lenList {
		stop = lenList - 1
	}
	return helpers.CopyList(list.value[start : stop+1]), nil
}

func (s *Storage) LPush(key string, value ...[]byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var arr List
	it, ok := s.storage[key]
	if !ok {
		it = item{}
	} else {
		arr, ok = it.value.(List)
		if !ok {
			return 0, helpers.ErrWrongType
		}
	}
	newSlice := make([][]byte, len(arr.value)+len(value))
	for i, v := range value {
		newSlice[len(value)-1-i] = v
	}
	copy(newSlice[len(value):], arr.value)
	arr.value = newSlice
	maxLen := len(arr.value)
	s.handleListAppendLocked(arr, key, it)
	return maxLen, nil
}

func (s *Storage) LLEN(key string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	it, ok := s.storage[key]
	if !ok {
		return 0, nil
	}
	list, ok := it.value.(List)
	if !ok {
		return 0, helpers.ErrWrongType
	}
	return len(list.value), nil
}

func (s *Storage) LPOP(key string, count *int64) ([][]byte, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	it, ok := s.storage[key]
	if !ok {
		return nil, false, nil
	}
	list, ok := it.value.(List)
	if !ok {
		return nil, true, helpers.ErrWrongType
	}
	if len(list.value) == 0 {
		delete(s.storage, key)
		return nil, false, nil
	}
	var popCount int64 = 1
	if count != nil {
		popCount = min(*count, int64(len(list.value)))
	}

	// Gonna shallow copy since the these elements would be effectively deleted
	firstElms := helpers.CopyListShallow(list.value[0:popCount])
	copy(list.value[0:], list.value[popCount:])
	newLen := int64(len(list.value)) - popCount
	for k := newLen; k < int64(len(list.value)); k++ {
		list.value[k] = nil
	}
	list.value = list.value[:newLen]

	if len(list.value) == 0 {
		delete(s.storage, key)
	} else {
		it.value = list
		s.storage[key] = it
	}
	return firstElms, true, nil
}

func (s *Storage) BLPOP(key string, timeout float64) ([][]byte, error) {
	s.mu.Lock()
	it, ok := s.storage[key]
	if ok {
		list, ok := it.value.(List)
		if !ok {
			s.mu.Unlock()
			return nil, helpers.ErrWrongType
		}
		item, ok := list.popFront()
		if ok {
			if len(list.value) == 0 {
				delete(s.storage, key)
			} else {
				it.value = list
				s.storage[key] = it
			}
			s.mu.Unlock()
			return [][]byte{[]byte(key), item}, nil
		}
	}
	listenCh := make(chan []byte, 1)
	s.listners[key] = listenCh
	s.mu.Unlock()
	item := <-listenCh
	return [][]byte{[]byte(key), item}, nil
}
