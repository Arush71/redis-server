package storage

import "github.com/Arush71/redis-server/internal/helpers"

func (s *Storage) XADD(streamKey string, id string, values [][]byte) (string, error) {
	timeID, seqID, err := helpers.ParseStreamID(id)
	if err != nil {
		return "", err
	}
	streamMap := make(streamStorageType, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		streamMap[string(values[i])] = values[i+1]
	}
	newEntry := streamEntry{
		id:            streamID{},
		streamStorage: streamMap,
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	it, ok := s.storage[streamKey]
	if ok {
		stream, ok := it.value.(Stream)
		if !ok {
			return "", helpers.ErrWrongType
		}
		if len(stream.value) > 0 {
			lastID := stream.value[len(stream.value)-1].id
			if seqID == nil {
				// TODO:do the auto-generate seq.
			} else {
				var num uint64 = 0
				seqID = &num
			}
			if lastID.time > newEntry.id.time || (lastID.time == newEntry.id.time && lastID.seq >= newEntry.id.seq) {
				return "", helpers.ErrInvalidIdAppend
			}
		}
		stream.value = append(stream.value, newEntry)
		it.value = stream
		s.storage[streamKey] = it
		return id, nil
	}
	streamEntries := make([]streamEntry, 1)
	streamEntries[0] = newEntry
	s.storage[streamKey] = item{
		expiresAt: nil,
		value:     Stream{value: streamEntries},
	}
	return id, nil
}
