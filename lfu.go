package fastlfu

/* https://arxiv.org/pdf/2110.11602.pdf */

type set[K comparable] map[K]struct{}

// a freqNode is a node in the 'frequency list', it holds the items having the
// same frequency (i.e. items with the same number of accesses).
type freqNode[T comparable] struct {
	next, prev *freqNode[T] // fequency list neighbour nodes.
	items      set[T]       // items
	freq       uint64       // number of accesses
}

// newNode creates a new frequency node and inserts it between prev and freq.
func newNode[T comparable](v uint64, prev, next *freqNode[T]) *freqNode[T] {
	n := &freqNode[T]{
		items: make(set[T]),
		freq:  v,
		prev:  prev,
		next:  next,
	}
	prev.next = n
	next.prev = n
	return n
}

// unlink unlinks n from its own frequency list.
func (n *freqNode[T]) unlink() {
	n.prev.next = n.next
	n.next.prev = n.prev
}

type Cache[K comparable, V any] struct {
	bykey    map[K]*lfuItem[K, V]
	freqhead *freqNode[K]
}

type lfuItem[K comparable, V any] struct {
	data   V
	parent *freqNode[K] // points back to the first node in the frequency list containing this lfuItem.
}

func NewCache[K comparable, V any]() *Cache[K, V] {
	// Initialize the first frequency list.
	node := &freqNode[K]{
		items: make(set[K]),
	}
	node.prev = node
	node.next = node

	return &Cache[K, V]{
		bykey:    make(map[K]*lfuItem[K, V]),
		freqhead: node,
	}
}

// Len returns the number of elements in the cache.
func (c *Cache[K, V]) Len() int {
	return len(c.bykey)
}

// Evict evicts a single item from the cache, randomly chosen among the list of
// least frequently used items, and returns that item and a boolean equals to
// true. If the cache is empty and no item can be evicted, Evict returns the
// zero-value of T and false.
func (c *Cache[K, V]) Evict() (K, bool) {
	for k := range c.freqhead.next.items {
		item := c.bykey[k]
		if len(c.freqhead.next.items) == 1 {
			// No other elements having the current frequency
			item.parent.unlink()
		}
		delete(c.bykey, k)
		delete(c.freqhead.next.items, k)
		return k, true
	}

	var k K
	return k, false
}

// EvictMultiple evicts up to n items from the cache, randomly chosen among the
// least frequently used items, and returns the number of items actually
// evicted.
func (c *Cache[K, V]) EvictMultiple(n int) int {
	evicted := 0

	cur := c.freqhead.next
	for evicted < n {
		for k := range c.freqhead.next.items {
			item := c.bykey[k]
			item.parent.unlink()
			delete(c.bykey, k)
			evicted++
		}
		cur = cur.next
		if cur == nil || cur.next == c.freqhead.next {
			break
		}
	}

	return evicted
}

/*
01 if key in lfu cache.bykey then
02 throw Exception(”Key already exists”)
03
04 freq ← lfu cache.freq head.next
05 if freq.value does not equal 1 then
06 freq ← GET-NEW-NODE(1, lfu cache.freq head, freq)
07
08 freq.items.add(key)
09 lfu cache.bykey[key] ← NEW-LFU-ITEM(value, freq)
*/
func (c *Cache[K, V]) Insert(key K, value V) {
	_, ok := c.bykey[key]
	if ok {
		// TODO(arl) we shouldn't panic but probably just Fetch the item, and replace its value.
		panic("Insert: key already exists")
	}

	freq := c.freqhead.next
	if freq.freq != 1 {
		freq = newNode(1, c.freqhead, freq)
	}

	freq.items[key] = struct{}{}
	c.bykey[key] = &lfuItem[K, V]{
		data:   value,
		parent: freq,
	}
}

/*
01 tmp ← lfu cache.bykey[key]
02 if tmp equals null then
03 throw Exception(”No such key”)
04 freq ← tmp.parent
05 next freq ← freq.next
06
07 if next freq equals lfu cache.freq head or
08 next freq.value does not equal freq.value + 1 then
08 next freq ← GET-NEW-NODE(freq.value + 1, freq, next freq)
09 next freq.items.add(key)
10 tmp.parent ← next freq
11
12 freq.items.remove(key)
13 if freq.items.length equals 0 then
14 DELETE-NODE(freq)
15 return tmp.data
*/

// Fetch fetches the value associated with key and returns it, with true, and
// increments its access frequency. However if there's no such key in the cache,
// it returns the zero value of the value type and false.
func (c *Cache[K, V]) Fetch(key K) (V, bool) {
	tmp := c.bykey[key]
	if tmp == nil {
		var v V
		return v, false
	}
	freq := tmp.parent
	nextFreq := freq.next

	if nextFreq == c.freqhead || nextFreq.freq != freq.freq+1 {
		nextFreq = newNode(freq.freq+1, freq, nextFreq)
	}
	nextFreq.items[key] = struct{}{}
	tmp.parent = nextFreq

	delete(freq.items, key)
	if len(freq.items) == 0 {
		freq.unlink()
	}

	return tmp.data, true
}
