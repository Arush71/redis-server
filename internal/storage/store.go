// Package storage is responsible for in-memory storage
package storage

import (
	"sync"
	"time"

	"github.com/Arush71/redis-server/internal/helpers"
)

type Storage struct {
	mu      sync.RWMutex
	storage storage
}

func InitStorage() *Storage {
	return &Storage{
		storage: make(storage),
		mu:      sync.RWMutex{},
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
	return str.value, ok
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
	it.value = arr
	s.storage[key] = it
	return len(arr.value), nil
}

func (s *Storage) LRange(key string, start int64, stop int64) ([][]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	it, ok := s.storage[key]
	if !ok {
		return [][]byte{}, nil
	}
	list, ok := it.value.(List)
	if !ok {
		return nil, helpers.ErrWrongType
	}
	lenList := int64(len(list.value))
	if start >= lenList || start > stop {
		return [][]byte{}, nil
	}
	if stop >= lenList {
		stop = lenList - 1
	}
	// TODO: solve the potential concurrency problem and complete the list function.
	return list.value[start : stop+1], nil
}
