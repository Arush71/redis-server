package helpers

func CopyBytes(b []byte) []byte {
	target := make([]byte, len(b))
	copy(target, b)
	return target
}

func CopyList(list [][]byte) [][]byte {
	target := make([][]byte, len(list))
	copy(target, list)
	for i, v := range target {
		target[i] = CopyBytes(v)
	}
	return target
}

func CopyListShallow(list [][]byte) [][]byte {
	target := make([][]byte, len(list))
	copy(target, list)
	return target
}
