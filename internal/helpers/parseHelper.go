// Package helpers helps with stuff
package helpers

import (
	"math"
	"strconv"
	"strings"

	"github.com/Arush71/redis-server/internal/models"
)

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

func ParseStreamID(d string) (*uint64, *uint64, error) {
	if d == "*" {
		return nil, nil, nil
	}
	before, after, found := strings.Cut(d, "-")
	if !found {
		return nil, nil, ErrStreamIdParse
	}
	v1, err := strconv.ParseUint(before, 10, 64)
	if err != nil {
		return nil, nil, ErrStreamIdParse
	}
	if after == "*" {
		return &v1, nil, nil
	}
	v2, err := strconv.ParseUint(after, 10, 64)
	if err != nil {
		return nil, nil, ErrStreamIdParse
	}
	if v1 == 0 && v2 == 0 {
		return nil, nil, ErrStreamIdZero
	}
	return &v1, &v2, nil
}

func ParseStreamIDsRange(start string, end string) (*models.StreamID, *models.StreamID, error) {
	// start parse
	newStart := &models.StreamID{}
	if start == "-" {
		newStart.Time = 0
		newStart.Seq = 0
	} else {
		before, after, found := strings.Cut(start, "-")
		if !found {
			before = start
			newStart.Seq = 0
		} else {
			seq, err := strconv.ParseUint(after, 10, 64)
			if err != nil {
				return nil, nil, ErrStreamIdParse
			}
			newStart.Seq = seq
		}
		time, err := strconv.ParseUint(before, 10, 64)
		if err != nil {
			return nil, nil, ErrStreamIdParse
		}
		newStart.Time = time
	}

	// end parse
	newEnd := &models.StreamID{}
	if end == "+" {
		newEnd.Time = math.MaxUint64
		newEnd.Seq = math.MaxUint64
	} else {
		before, after, found := strings.Cut(end, "-")
		if !found {
			before = end
			newEnd.Seq = math.MaxUint64
		} else {
			seq, err := strconv.ParseUint(after, 10, 64)
			if err != nil {
				return nil, nil, ErrStreamIdParse
			}
			newEnd.Seq = seq
		}
		time, err := strconv.ParseUint(before, 10, 64)
		if err != nil {
			return nil, nil, ErrStreamIdParse
		}
		newEnd.Time = time
	}

	return newStart, newEnd, nil
}
