package storage

// NOTE: This function should only be called with the write lock acquired of the storage.
func (s *Storage) handleListAppendLocked(list List, key string, itemToAppend item) {
	if len(list.value) == 0 {
		return
	}
	s.listners.TryDequeueListners(key, &list)
	if len(list.value) == 0 {
		delete(s.storage, key)
		return
	}
	itemToAppend.value = list
	s.storage[key] = itemToAppend
}
