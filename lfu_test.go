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

	ek1, enode1 := c.evictLFU()
	unlink(enode1.parent)
	delete(c.bykey, ek1)
	fmt.Println("evicted", ek1)

	ek2, enode2 := c.evictLFU()
	unlink(enode2.parent)
	delete(c.bykey, ek2)
	fmt.Println("evicted", ek2)
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
		evicted, item := c.evictLFU()
		if evicted != keyFrom(i) {
			t.Fatalf("evicted %+v, want %+v", evicted, keyFrom(i))
		}

		// TODO(arl) -> this should probably be done in evictLFU if we kepe this public API
		// untested
		unlink(item.parent)
		delete(c.bykey, evicted)

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

/*
func TestEvict(t *testing.T) {
	c := NewCache()

	for nitems := 1; nitems <= 10; nitems++ {
		t.Log("~~~ nitems =", nitems, "~~~")
		c.Insert(keyFrom(nitems), V(nitems))
		for nfetches := 1; nfetches <= nitems; nfetches++ {
			key := keyFrom(nfetches)
			got := c.Fetch(key)
			if got != V(nfetches) {
				t.Errorf("Fetch(%q) = %v, want %v [nitems=%d]", key, got, nfetches, nitems)
			}
			t.Logf("[nitems=%d] Fetch(%q) = %v", nitems, key, got)
		}

		// key, item := c.evictLFU()
		// fmt.Printf("key=%v item=%+v", key, item)
		// c.debug("just created")

		for nevicts := 0; nevicts < nitems; nevicts++ {
			k, _ := c.evictLFU()
			t.Logf("[nitems=%d] evicted -> %v %v/%v", nitems, k, nevicts, nitems)
		}
	}

		// c.debug("just created")

		// c.Insert(keyFrom(1), 1)
		// c.debug("inserted key1/1")

		// key, item := c.evictLFU()
		// fmt.Printf("key=%v item=%+v", key, item)
		// c.debug("just created")
}
*/

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
