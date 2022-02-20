package fastlfu

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

type lfuItem[K comparable, V any] struct {
	data   V
	parent *freqNode[K] // points back to the first node in the frequency list containing this lfuItem.
}

type Cache[K comparable, V any] struct {
	bykey    map[K]*lfuItem[K, V]
	freqhead *freqNode[K]
	maxed    bool
}

func New[K comparable, V any]() *Cache[K, V] {
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

func NewMaxed[K comparable, V any]() *Cache[K, V] {
	c := New[K, V]()
	c.maxed = true
	return c
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

// Insert a key value pair in the cache. If key already is in the cache, its
// value is replaced by value, but its access frequency isn't changed.
func (c *Cache[K, V]) Insert(key K, value V) {
	item, ok := c.bykey[key]
	if ok {
		item.data = value
		return
	}

	// If we're a maxed cache, we need to evict before inserting a new pair.
	if c.maxed {
		c.Evict()
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

// Fetch fetches the value associated with a key and returns it, with true, and
// increments its access frequency. However if there's no such key in the cache,
// it returns the zero value of the value type and false.
func (c *Cache[K, V]) Fetch(key K) (V, bool) {
	item := c.bykey[key]
	if item == nil {
		var v V
		return v, false
	}
	freq := item.parent
	nextFreq := freq.next

	if nextFreq == c.freqhead || nextFreq.freq != freq.freq+1 {
		nextFreq = newNode(freq.freq+1, freq, nextFreq)
	}
	nextFreq.items[key] = struct{}{}
	item.parent = nextFreq

	delete(freq.items, key)
	if len(freq.items) == 0 {
		freq.unlink()
	}

	return item.data, true
}
