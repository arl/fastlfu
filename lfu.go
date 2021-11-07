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

func NewCache() *Cache {
	c := &Cache{
		bykey:    make(map[T]*lfuItem),
		freqhead: newFreqNode(),
	}

	return c
}

// Evict evicts a single item from the list containing the least frequently used
// items, and returns that item and a boolean set to true. If the cache is empty
// and no item can be evicted, Evict returns the zero-value of T and false.
func (c *Cache) Evict() (T, bool) {
	for k := range c.freqhead.next.items {
		item := c.bykey[k]
		unlink(item.parent)
		delete(c.bykey, k)
		return k, true
	}

	var t T
	return t, false
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
		panic("Insert: key already exists")
	}

	freq := c.freqhead.next
	if freq.value != 1 {
		freq = getNewNode(1, c.freqhead, freq)
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
func (c *Cache) Fetch(key T) V {
	tmp := c.bykey[key]
	if tmp == nil {
		panic("no such key")
	}
	freq := tmp.parent
	nextFreq := freq.next

	if nextFreq == c.freqhead || nextFreq.value != freq.value+1 {
		nextFreq = getNewNode(freq.value+1, freq, nextFreq)
	}
	nextFreq.items[key] = struct{}{}
	tmp.parent = nextFreq

	delete(freq.items, key)
	if len(freq.items) == 0 {
		unlink(freq)
	}

	return tmp.data
}

type lfuItem struct {
	data   V
	parent *freqNode
}

type freqNode struct {
	next, prev *freqNode
	items      set
	value      float64
}

// newFreqNode creates a new frequency node with an access frequency value of 0
func newFreqNode() *freqNode {
	n := &freqNode{
		items: make(set),
	}
	n.prev = n
	n.next = n
	return n
}

// s/getNewNode/newNode
func getNewNode(v float64, prev, next *freqNode) *freqNode {
	nn := newFreqNode()
	nn.value = v
	nn.prev = prev
	nn.next = next
	prev.next = nn
	next.prev = nn
	return nn
}

func unlink(n *freqNode) {
	n.prev.next = n.next
	n.next.prev = n.prev
}
