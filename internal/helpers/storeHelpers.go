package helpers

func DeduplicateKeys(keys [][]byte) [][]byte {
	store := make(map[string]struct{})
	new := make([][]byte, 0, len(keys))
	for _, v := range keys {
		value := string(v)
		if _, ok := store[value]; ok {
			continue
		}
		new = append(new, v)
		store[value] = struct{}{}
	}
	return new
}
