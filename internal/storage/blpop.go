package storage

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/Arush71/redis-server/internal/helpers"
)

func (s *Storage) BLPOP(ctx context.Context, timeout float64, keys [][]byte, conn io.Writer, logger *slog.Logger) ([][]byte, error) {
	keys = helpers.DeduplicateKeys(keys)
	s.mu.Lock()
	for i := range len(keys) {
		key := string(keys[i])
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
	}
	listnerCh := make(listenCh, 1)
	waiterToWait := &waiter{
		claimed:  false,
		listenCh: listnerCh,
	}
	s.listners.EnqueueListners(waiterToWait, keys)
	s.mu.Unlock()
	defer s.listners.removeKeys(keys, waiterToWait)
	var timerCh <-chan time.Time
	if timeout > 0 {
		d := time.Duration(float64(time.Second) * timeout)
		timerCh = time.After(d)
	}
	select {
	case <-ctx.Done():
		return nil, ErrEOF
	case item := <-listnerCh:
		return [][]byte{[]byte(item.key), item.value}, nil
	case <-timerCh:
		return nil, nil
	}
}
