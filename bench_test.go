package fastlfu

import (
	"testing"
)

func BenchmarkInsert(b *testing.B) {
	c := New[int, int]()
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c.Insert(n, n)
	}
}

func benchmarkFetch(nitems int, hit bool) func(b *testing.B) {
	return func(b *testing.B) {
		var key int
		if hit {
			key = 0
		} else {
			key = nitems
		}

		c := New[int, int]()
		for i := 0; i < nitems; i++ {
			c.Insert(i, i)
		}

		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			v, ok := c.Fetch(key)
			_, _ = v, ok
			if ok != hit {
				b.Fatalf("Fetch(%v) = (%v, %t), want %t", key, v, ok, hit)
			}
		}
	}
}

func BenchmarkFetch(b *testing.B) {
	b.Run("fetch=hit/items=10", benchmarkFetch(10, true))
	b.Run("fetch=hit/items=100", benchmarkFetch(100, true))
	b.Run("fetch=hit/items=1000", benchmarkFetch(1000, true))
	b.Run("fetch=hit/items=10000", benchmarkFetch(10000, true))
	b.Run("fetch=hit/items=100000", benchmarkFetch(100000, true))

	b.Run("fetch=miss/items=10", benchmarkFetch(10, false))
	b.Run("fetch=miss/items=100", benchmarkFetch(100, false))
	b.Run("fetch=miss/items=1000", benchmarkFetch(1000, false))
	b.Run("fetch=miss/items=10000", benchmarkFetch(10000, false))
	b.Run("fetch=miss/items=100000", benchmarkFetch(100000, false))
}

func BenchmarkEvict(b *testing.B) {
	c := New[int, int]()
	for i := 0; i < b.N; i++ {
		c.Insert(i, i)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		v, ok := c.Evict()
		_, _ = v, ok
		if !ok {
			b.Fatalf("Evict() = false, want true")
		}
	}
}

var sink interface{}

func BenchmarkFetchLastNodeItem(b *testing.B) {
	// Test the case where the fetched item is the last on its frequency node,
	// and the next (+1) node doesn't exist.
	c := New[int, string]()
	c.Insert(1, "foo")

	b.ReportAllocs()
	b.ResetTimer()

	var (
		val string
		ok  bool
	)
	for n := 0; n < b.N; n++ {
		val, ok = c.Fetch(1)
	}

	sink, sink = val, ok
}
