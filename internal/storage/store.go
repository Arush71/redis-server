// Package storage is responsible for in-memory storage
package storage

import (
	"sync"
	"time"
)

type item struct {
	value     []byte
	ExpiresAt *time.Time
}

type storage map[string]item

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
	s.mu.Lock()
	defer s.mu.Unlock()
	var expiresAt *time.Time
	if px >= 0 {
		t := time.Now().Add(time.Duration(px) * time.Millisecond)
		expiresAt = &t
	}
	s.storage[key] = item{
		value:     val,
		ExpiresAt: expiresAt,
	}
}

func (s *Storage) Get(key string) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, ok := s.storage[key]
	if ok && val.ExpiresAt != nil && time.Now().After(*val.ExpiresAt) {
		delete(s.storage, key)
		return nil, false
	}
	return val.value, ok
}
