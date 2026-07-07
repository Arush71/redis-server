package storage

// NOTE: This function should only be called with the write lock acquired.
func (s *Storage) handleListAppendLocked(list List, key string, itemToAppend item) {
	if len(list.value) == 0 {
		return
	}
	ch, ok := s.listners[key]
	if ok {
		firstElm, _ := list.popFront()
		ch <- firstElm
		delete(s.listners, key)
		if len(list.value) == 0 {
			delete(s.storage, key)
			return
		}
	}
	itemToAppend.value = list
	s.storage[key] = itemToAppend
}
