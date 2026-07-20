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

type (
	listenChType struct {
		key   string
		value []byte
	}
	listenCh     chan listenChType
	listnerTable map[string][]listenCh
)

func (l *listnerManager) EnqueueListners(listnerCh listenCh, keys [][]byte) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i := range len(keys) {
		key := string(keys[i])
		l.table[key] = append(l.table[key], listnerCh)
	}
}

func (l *listnerManager) TryDequeueListners(key string, list *List) {
	l.mu.Lock()
	defer l.mu.Unlock()
	listners := l.table[key]
	if len(listners) == 0 {
		return
	}
	toLoop := min(len(listners), len(list.value))
	for i := range toLoop {
		data, ok := list.popFront()
		if !ok {
			return
		}
		listners[i] <- listenChType{
			key:   key,
			value: data,
		}
		listners[i] = nil
	}
	listners = listners[toLoop:]
	if len(listners) == 0 {
		delete(l.table, key)
		return
	}
	l.table[key] = listners
}

func (l *listnerManager) removeKeys(keys [][]byte, listnerChToDel listenCh) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i := range len(keys) {
		key := string(keys[i])
		listenerChs, ok := l.table[key]
		if !ok {
			continue
		}
		for j := range listenerChs {
			if listenerChs[j] != listnerChToDel {
				continue
			}
			if j < len(listenerChs)-1 {
				copy(listenerChs[j:], listenerChs[j+1:])
			}
			listenerChs[len(listenerChs)-1] = nil
			listenerChs = listenerChs[:len(listenerChs)-1]
			break
		}
		if len(listenerChs) == 0 {
			delete(l.table, key)
			continue
		}
		l.table[key] = listenerChs
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
