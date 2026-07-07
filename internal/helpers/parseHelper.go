// Package helpers helps with stuff
package helpers

import "strconv"

func ParsePositiveInt(d []byte) (int64, bool) {
	if len(d) == 0 {
		return -1, false
	}
	num, err := strconv.ParseInt(string(d), 10, 64)
	if err != nil || num < 0 {
		return -1, false
	}
	return num, true
}

func ParseInt(d []byte) (int64, bool) {
	if len(d) == 0 {
		return 0, false
	}
	num, err := strconv.ParseInt(string(d), 10, 64)
	if err != nil {
		return 0, false
	}
	return num, true
}

func ParsePositiveFloat(d []byte) (float64, error) {
	if len(d) == 0 {
		return 0, ErrTimeoutNotFloat
	}
	num, err := strconv.ParseFloat(string(d), 64)
	if err != nil {
		return 0, ErrTimeoutNotFloat
	}
	if num < 0 {
		return 0, ErrTimeoutNegative
	}
	return num, nil
}
