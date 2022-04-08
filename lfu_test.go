package fastlfu

import (
	"fmt"
	"reflect"
	"testing"
)

func TestFastLFU(t *testing.T) {
	c := New[int, int]()

	for i := 0; i < 10; i++ {
		c.Insert(i, i)
	}

	for i := 0; i < 10; i++ {
		v, ok := c.Fetch(i)
		if v != i || !ok {
			t.Errorf("Fetch(%q) = (%v, %t), want (%v, true)", i, v, ok, i)
		}
	}

	for i := 0; i < 7; i++ {
		c.Fetch(i)
	}

	c.debugln("before first evict")

	// TODO convert in propoer unit-tests
	ev1, ok1 := c.Evict()
	if testing.Verbose() {
		fmt.Printf("evicted? %t, item = %+v\n", ok1, ev1)
	}

	ev2, ok2 := c.Evict()
	if testing.Verbose() {
		fmt.Printf("evicted? %t, item = %+v\n", ok2, ev2)
	}

	ev3, ok3 := c.Evict()
	if testing.Verbose() {
		fmt.Printf("evicted? %t, item = %+v\n", ok3, ev3)
	}
}

func testEvict(t *testing.T, nitems int) {
	c := New[int, int]()

	for i := 0; i < nitems; i++ {
		c.Insert(i, i)
	}

	c.debugln("after insertions")

	// We want to force eviction of the ith element
	for i := 0; i < nitems; i++ {
		t.Log("in this loop, we want to evict", i)
		// We fetch all elements but the ith element
		for f := 0; f < nitems; f++ {
			if f != i {
				c.Fetch(f)
				c.debugf("after fetch(%s)", f)
			}
		}
		c.debugf("before eviction. (should evict %s)", i)
		evicted, ok := c.Evict()
		if evicted != i || !ok {
			t.Fatalf("Evict() = (%+v, %t), want (%+v, %t)", evicted, ok, i, true)
		}

		c.debugln("after successful eviction of", i)

		// We now reinsert the evicted element, and we artifically fetch it a
		// number a times so that it gets the same frequency as every other
		// element in the cache.
		c.Insert(i, i)

		tot := 1
		for j := 0; j <= i; j++ {
			tot += j
		}
		for j := 0; j < tot; j++ {
			c.Fetch(i)
		}

		c.debugln("after re-insertion of", i)
	}
}

func TestInsert(t *testing.T) {
	const marker = "-reinserted"
	items := map[int]string{
		0: "A",
		1: "B",
		2: "C",
		3: "D",
		4: "E",
	}
	c := New[int, string]()
	for k, v := range items {
		c.Insert(k, v)
	}

	// Change the value before fetching.
	for k := 0; k < 4; k++ {
		c.Insert(k, items[k]+marker)
		vf, ok := c.Fetch(k)
		want := items[k] + marker
		if want != vf {
			t.Errorf("Fetch(%v) = (%v, %t), want (%v, %t)", k, vf, ok, want, ok)
		}
	}

	// Ensure we evict the only key we hadn't fetched.
	evicted, ok := c.Evict()
	if evicted != 4 {
		t.Errorf("evicted = (%v, %t), want (4, ok)", evicted, ok)
	}
}

func TestEvict(t *testing.T) {
	testEvict(t, 100)
}

func testEvictSameFrequencies(nitems int) func(t *testing.T) {
	return func(t *testing.T) {
		c := New[int, int]()
		for i := 0; i < nitems; i++ {
			c.Insert(i, i)
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
	wantItems   map[int]int // expected content in cache after all evictions
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
			wantItems:   map[int]int{1: 3, 2: 3, 3: 3, 4: 2},
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
			wantItems:   map[int]int{1: 3, 2: 3, 3: 3},
		},
		{
			name:        "try evict on empty cache",
			freqs:       map[int]int{},
			nevictions:  1,
			wantEvicted: 0,
			wantItems:   map[int]int{},
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
			wantItems:   map[int]int{1: 3, 2: 3, 3: 3, 4: 2, 5: 1},
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
			wantItems:   map[int]int{},
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
			wantItems:   map[int]int{},
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

	c := New[int, int]()
	fillCache(c, tt.freqs, tt.freqs)
	evicted := c.EvictMultiple(tt.nevictions)

	if evicted != tt.wantEvicted {
		t.Errorf("items evicted = %d, want %d", evicted, tt.wantEvicted)
	}

	if got := c.items(); !reflect.DeepEqual(got, tt.wantItems) {
		t.Errorf("got:\n%+v\n\nwant:\n%+v\n", got, tt.wantItems)
	}
}

func (c *Cache[K, V]) debugln(a ...interface{}) {
	if !testing.Verbose() {
		return
	}

	a = append(a, ":")
	fmt.Println(a...)
	c.debug()
	fmt.Println()
}

func (c *Cache[K, V]) debugf(format string, a ...interface{}) {
	if !testing.Verbose() {
		return
	}

	fmt.Printf(format+" :\n", a...)
	c.debug()
	fmt.Println()
}

func (c *Cache[K, V]) debug() {
	c.forEachFrequency(func(freq int, s set[K]) {
		var sl []K
		for k := range s {
			sl = append(sl, k)
		}
		fmt.Printf("\tfreq=%d -> {%+v}\n", freq, sl)
	})
}

func (c *Cache[K, V]) forEachFrequency(f func(freq int, s set[K])) {
	cur := c.freqhead.next
	// TODO(arl) should the frequency linked list really be circular?
	for cur != nil && cur.next != c.freqhead.next {
		f(int(cur.freq), cur.items)
		cur = cur.next
	}
}

func (c *Cache[K, V]) items() map[K]V {
	m := make(map[K]V)
	for k, v := range c.bykey {
		m[k] = v.data
	}
	return m
}

// items is a map of key, which values are the key access frequency.
func fillCache[K comparable, V any](c *Cache[K, V], items map[K]V, freqs map[K]int) {
	for k, v := range items {
		c.Insert(k, v)

		freq := freqs[k]
		if freq < 1 {
			panic("an item can't have a frequency < 1")
		}
		for i := 1; i < freq; i++ {
			c.Fetch(k)
		}
	}
}

func TestMaxedCache(t *testing.T) {
	freqs := map[int]int{
		1: 3,
		2: 3,
		3: 3,
		4: 2,
		5: 1,
	}
	c := NewMaxed[int, int](5)
	fillCache(c, freqs, freqs)

	if got := c.Len(); got != 5 {
		t.Fatalf("c.Len() = %d, want 5", got)
	}

	c.Insert(1, 0)
	if got := c.Len(); got != 5 {
		t.Fatalf("c.Len() = %d, want 5", got)
	}

	c.Insert(6, 0)
	if got := c.Len(); got != 5 {
		t.Fatalf("c.Len() = %d, want 5", got)
	}
}

func TestFetchLastNodeItem(t *testing.T) {
	// Test the case where the fetched item is the last on its frequency node,
	// and the next (+1) node doesn't exist.
	c := New[int, string]()
	c.Insert(1, "foo")
	if val, ok := c.Fetch(1); val != "foo" || !ok {
		t.Errorf("fetched(1) -> (%v, %v), want (%v, %v)", val, ok, "foo", true)
	}

	if val, ok := c.Fetch(1); val != "foo" || !ok {
		t.Errorf("fetched(1) -> (%v, %v), want (%v, %v)", val, ok, "foo", true)
	}

	if val, ok := c.Fetch(1); val != "foo" || !ok {
		t.Errorf("fetched(1) -> (%v, %v), want (%v, %v)", val, ok, "foo", true)
	}
}
