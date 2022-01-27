package fastlfu

import (
	"fmt"
	"reflect"
	"testing"
)

func keyFrom(i int) T {
	return T(i)
}

func TestFastLFU(t *testing.T) {
	c := NewCache()

	for i := 0; i < 10; i++ {
		key := keyFrom(i)
		val := V(i)
		c.Insert(key, val)
	}

	for i := 0; i < 10; i++ {
		key := keyFrom(i)
		v, ok := c.Fetch(key)
		if v != V(i) || !ok {
			t.Errorf("Fetch(%q) = (%v, %t), want (%v, true)", key, v, ok, i)
		}
	}

	for i := 0; i < 7; i++ {
		key := keyFrom(i)
		c.Fetch(key)
	}

	c.debugln("before first evict")

	// TODO(arl) convert in propoer unit-tests
	ev1, ok1 := c.Evict()
	fmt.Printf("evicted? %t, item = %+v\n", ok1, ev1)

	ev2, ok2 := c.Evict()
	fmt.Printf("evicted? %t, item = %+v\n", ok2, ev2)

	ev3, ok3 := c.Evict()
	fmt.Printf("evicted? %t, item = %+v\n", ok3, ev3)
}

func testEvict(t *testing.T, nitems int) {
	c := NewCache()

	for i := 0; i < nitems; i++ {
		c.Insert(keyFrom(i), V(i))
	}

	c.debugln("after insertions")

	// We want to force eviction of the ith element
	for i := 0; i < nitems; i++ {
		t.Log("in this loop, we want to evict", keyFrom(i))
		// We fetch all elements but the ith element
		for f := 0; f < nitems; f++ {
			if f != i {
				k := keyFrom(f)
				c.Fetch(k)
				c.debugf("after fetch(%s)", k)
			}
		}
		c.debugf("before eviction. (should evict %s)", keyFrom(i))
		evicted, ok := c.Evict()
		if evicted != keyFrom(i) || !ok {
			t.Fatalf("Evict() = (%+v, %t), want (%+v, %t)", evicted, ok, keyFrom(i), true)
		}

		c.debugln("after successful eviction of", keyFrom(i))

		// We now reinsert the evicted element, and we artifically fetch it a
		// number a times so that it gets the same frequency as every other
		// element in the cache.
		c.Insert(keyFrom(i), V(i))

		tot := 1
		for j := 0; j <= i; j++ {
			tot += j
		}
		for j := 0; j < tot; j++ {
			c.Fetch(keyFrom(i))
		}

		c.debugln("after re-insertion of", keyFrom(i))
	}
}

func TestEvict(t *testing.T) {
	testEvict(t, 100)
}

func testEvictSameFrequencies(nitems int) func(t *testing.T) {
	return func(t *testing.T) {
		c := NewCache()
		for i := 0; i < nitems; i++ {
			c.Insert(keyFrom(i), V(i))
		}
		// We can successfully evict nitems times
		for i := 0; i < nitems; i++ {
			if evc, ok := c.Evict(); !ok {
				t.Errorf("%dth Evict() -> (%v, %v), want (_, true)", i, evc, ok)
			}
		}

		// No more evictions possible
		if evc, ok := c.Evict(); ok {
			t.Errorf("last Evict() -> (%v, %v), want (_, false)", evc, ok)
		}
	}
}

func TestEvictSameFrequencies(t *testing.T) {
	t.Run("1", testEvictSameFrequencies(1))
	t.Run("10", testEvictSameFrequencies(10))
	t.Run("100", testEvictSameFrequencies(100))
	t.Run("1000", testEvictSameFrequencies(1000))
}

type evictMultipleTestCase struct {
	name        string
	freqs       map[int]int // state the cache should be for the test (key=>frequency)
	nevictions  int         // number of evictions to perform with EvictMultiple
	wantItems   []int       // the keys we want in the cache after evcitions
	wantEvicted int         // number of actual evictions performed by EvictMultiple
}

func TestEvictMultiple(t *testing.T) {
	tests := []evictMultipleTestCase{
		{
			name: "evict 1 key",
			freqs: map[int]int{
				1: 3,
				2: 3,
				3: 3,
				4: 2,
				5: 1,
			},
			nevictions:  1,
			wantEvicted: 1,
			wantItems:   []int{1, 2, 3, 4},
		},
		{
			name: "evict 2 keys",
			freqs: map[int]int{
				1: 3,
				2: 3,
				3: 3,
				4: 2,
				5: 1,
			},
			nevictions:  2,
			wantEvicted: 2,
			wantItems:   []int{1, 2, 3},
		},
		{
			name:        "try evict on empty cache",
			freqs:       map[int]int{},
			nevictions:  1,
			wantEvicted: 0,
			wantItems:   []int{},
		},
		{
			name: "evict nothing",
			freqs: map[int]int{
				1: 3,
				2: 3,
				3: 3,
				4: 2,
				5: 1,
			},
			nevictions:  0,
			wantEvicted: 0,
			wantItems:   []int{1, 2, 3, 4, 5},
		},
		{
			name: "evict everything",
			freqs: map[int]int{
				1: 3,
				2: 3,
				3: 3,
				4: 2,
				5: 1,
			},
			nevictions:  5,
			wantEvicted: 5,
			wantItems:   []int{},
		},
		{
			name: "evict everything",
			freqs: map[int]int{
				1: 3,
				2: 3,
				3: 3,
				4: 2,
				5: 1,
			},
			nevictions:  10,
			wantEvicted: 5,
			wantItems:   []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testEvictMultiple(t, tt)
		})
	}
}

func testEvictMultiple(t *testing.T, tt evictMultipleTestCase) {
	t.Helper()

	c := buildCache(tt.freqs)
	evicted := c.EvictMultiple(tt.nevictions)

	if evicted != tt.wantEvicted {
		t.Errorf("items evicted = %d, want %d", evicted, tt.wantEvicted)
	}

	want := make(map[T]V)
	for _, i := range tt.wantItems {
		want[keyFrom(i)] = V(i)
	}
	if got := c.items(); !reflect.DeepEqual(got, want) {
		t.Errorf("got:\n%+v\n\nwant:\n%+v\n", got, want)
	}
}

func (c *Cache) debugln(a ...interface{}) {
	if !testing.Verbose() {
		return
	}

	a = append(a, ":")
	fmt.Println(a...)
	debug(c)
	fmt.Println()
}

func (c *Cache) debugf(format string, a ...interface{}) {
	if !testing.Verbose() {
		return
	}

	fmt.Printf(format+" :\n", a...)
	debug(c)
	fmt.Println()
}

func debug(c *Cache) {
	c.forEachFrequency(func(freq int, s set) {
		var sl []T
		for k := range s {
			sl = append(sl, k)
		}
		fmt.Printf("\tfreq=%d -> {%+v}\n", freq, sl)
	})
}

func (c *Cache) forEachFrequency(f func(freq int, s set)) {
	cur := c.freqhead.next
	// TODO(arl) should the frequency linked list really be circular?
	for cur != nil && cur.next != c.freqhead.next {
		f(int(cur.freq), cur.items)
		cur = cur.next
	}
}

func (c *Cache) items() map[T]V {
	m := make(map[T]V)
	for k, v := range c.bykey {
		m[k] = v.data
	}
	return m
}

// items is a map key is the cache key and value frequency.
func buildCache(items map[int]int) *Cache {
	c := NewCache()
	for k, v := range items {
		skey := keyFrom(k)
		c.Insert(skey, V(k))
		if v < 1 {
			panic("an item can't have a frequency < 1")
		}
		for i := 1; i < v; i++ {
			c.Fetch(skey)
		}
	}
	return c
}

func BenchmarkInsert(b *testing.B) {
	c := NewCache()
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c.Insert(T(n), V(n))
	}
}

func benchmarkFetch(nitems int, hit bool) func(b *testing.B) {
	return func(b *testing.B) {
		var key T
		if hit {
			key = T(0)
		} else {
			key = T(nitems)
		}

		c := NewCache()
		for i := 0; i < nitems; i++ {
			c.Insert(T(i), V(i))
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
	b.Run("items=10/fetch=hit", benchmarkFetch(10, true))
	b.Run("items=10/fetch=miss", benchmarkFetch(10, false))
	b.Run("items=100/fetch=hit", benchmarkFetch(100, true))
	b.Run("items=100/fetch=miss", benchmarkFetch(100, false))
	b.Run("items=1000/fetch=hit", benchmarkFetch(1000, true))
	b.Run("items=1000/fetch=miss", benchmarkFetch(1000, false))
	b.Run("items=10000/fetch=hit", benchmarkFetch(10000, true))
	b.Run("items=10000/fetch=miss", benchmarkFetch(10000, false))
	b.Run("items=100000/fetch=hit", benchmarkFetch(100000, true))
	b.Run("items=100000/fetch=miss", benchmarkFetch(100000, false))
}

func BenchmarkEvict(b *testing.B) {
	c := NewCache()
	for i := 0; i < b.N; i++ {
		c.Insert(T(i), V(i))
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
