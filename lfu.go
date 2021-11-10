package fastlfu

/* https://arxiv.org/pdf/2110.11602.pdf
 *
 */

type T string

type V int

type set map[T]struct{}

type Cache struct {
	bykey    map[T]*lfuItem
	freqhead *freqNode
}

type lfuItem struct {
	data   V
	parent *freqNode // points back to the first node in the frequency list containing this lfuItem.
}

func NewCache() *Cache {
	// Initialize the first frequency list.
	node := &freqNode{
		items: make(set),
	}
	node.prev = node
	node.next = node

	return &Cache{
		bykey:    make(map[T]*lfuItem),
		freqhead: node,
	}
}

// Evict evicts a single item from the cache, randomly chosen among the list of
// least frequently used items, and returns that item and a boolean equals to
// true. If the cache is empty and no item can be evicted, Evict returns the
// zero-value of T and false.
func (c *Cache) Evict() (T, bool) {
	for k := range c.freqhead.next.items {
		item := c.bykey[k]
		item.parent.unlink()
		delete(c.bykey, k)
		return k, true
	}

	var t T
	return t, false
}

// EvictMultiple evicts up to n items from the cache, randomly chosen among the
// least frequently used items, and returns the number of items actually
// evicted.
func (c *Cache) EvictMultiple(n int) int {
	evicted := 0

	cur := c.freqhead.next
	for evicted < n {
		for k := range c.freqhead.next.items {
			item := c.bykey[k]
			item.parent.unlink()
			delete(c.bykey, k)
			evicted++
		}
		if cur == nil || cur.next == c.freqhead.next {
			break
		}
		cur = cur.next
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
func (c *Cache) Insert(key T, value V) {
	_, ok := c.bykey[key]
	if ok {
		// TODO(arl) we shouldn't panic but probably just Fetch the item, and replace its value.
		panic("Insert: key already exists")
	}

	freq := c.freqhead.next
	if freq.value != 1 {
		freq = newNode(1, c.freqhead, freq)
	}

	freq.items[key] = struct{}{}
	c.bykey[key] = &lfuItem{
		data:   value,
		parent: freq,
	}
}

// Fetch fetches an element from the LFU cache, simultaneously incrementing its
// frequency.

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

// Fetch ...  TODO(arl) document
func (c *Cache) Fetch(key T) V {
	tmp := c.bykey[key]
	if tmp == nil {
		// TODO(arl) we shouldn't panic and return (_, false) instead.
		panic("no such key")
	}
	freq := tmp.parent
	nextFreq := freq.next

	if nextFreq == c.freqhead || nextFreq.value != freq.value+1 {
		nextFreq = newNode(freq.value+1, freq, nextFreq)
	}
	nextFreq.items[key] = struct{}{}
	tmp.parent = nextFreq

	delete(freq.items, key)
	if len(freq.items) == 0 {
		freq.unlink()
	}

	return tmp.data
}

// a freqNode is a node in the 'frequency list', it holds the items having the
// same frequency (i.e. items with the same number of accesses).
type freqNode struct {
	next, prev *freqNode // fequency list neighbour nodes.
	items      set       // items
	value      float64   // frequency value. TODO(arl) should be an integer
}

// newNode creates a new frequency node and inserts it between prev and freq.
func newNode(v float64, prev, next *freqNode) *freqNode {
	n := &freqNode{
		items: make(set),
		value: v,
		prev:  prev,
		next:  next,
	}
	prev.next = n
	next.prev = n
	return n
}

// unlink unlinks n from its own frequency list.
func (n *freqNode) unlink() {
	n.prev.next = n.next
	n.next.prev = n.prev
}
