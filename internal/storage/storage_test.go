package storage

import (
	"strconv"
	"testing"
)

func BenchmarkStorageGet(b *testing.B) {
	s := InitStorage()
	s.Set("foo", []byte("bar"), -1)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = s.Get("foo")
	}
}

func BenchmarkStorageSet(b *testing.B) {
	s := InitStorage()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Set(strconv.Itoa(i), []byte("bar"), -1)
	}
}

func BenchmarkStorageMixed(b *testing.B) {
	s := InitStorage()
	s.Set("foo", []byte("bar"), -1)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Set("foo", []byte("bar"), -1)
		_, _ = s.Get("foo")
	}
}

func BenchmarkStorageReadParallel(b *testing.B) {
	s := InitStorage()
	s.Set("foo", []byte("bar"), -1)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = s.Get("foo")
		}
	})
}

func BenchmarkStorageWriteParallel(b *testing.B) {
	s := InitStorage()

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := strconv.Itoa(i)
			s.Set(key, []byte("bar"), -1)
			i++
		}
	})
}

func BenchmarkStorageMixedParallel(b *testing.B) {
	s := InitStorage()

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := strconv.Itoa(i)

			s.Set(key, []byte("bar"), -1)
			_, _ = s.Get(key)

			i++
		}
	})
}
