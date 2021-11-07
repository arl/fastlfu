package fastlfu

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func keyFrom(i int) T {
	return T("key" + strconv.Itoa(i))
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
		v := c.Fetch(key)
		fmt.Println(v)
		if v != V(i) {
			t.Errorf("Fetch(%q) = %v, want %v", key, v, i)
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
		var ss []string
		for k := range s {
			ss = append(ss, string(k))
		}
		sort.Strings(ss)
		fmt.Printf("\tfreq=%d -> {%s}\n", freq, strings.Join(ss, ", "))
	})
}

func (c *Cache) forEachFrequency(f func(freq int, s set)) {
	cur := c.freqhead.next
	// TODO(arl) don't know if the second check is normal (meaning that the linked list is circular)
	for cur != nil && cur.next != c.freqhead.next {
		f(int(cur.value), cur.items)
		cur = cur.next
	}
}
