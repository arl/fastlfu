package fastlfu

/* https://arxiv.org/pdf/2110.11602.pdf
 *
 */

type T string

type V int

type set map[T]struct{}

type Cache struct {
	bykey    map[T]*lfuItem
	freqhead *frequencyNode
}

func NewCache() *Cache {
	c := &Cache{
		bykey:    make(map[T]*lfuItem),
		freqhead: newFrequencyNode(),
	}

	return c
}

// routine called GET-LFU-ITEM in the paper
func (c *Cache) evictLFU() (T, *lfuItem) {
	for k := range c.freqhead.next.items {
		return k, c.bykey[k]
	}
	panic("the cache is empty")
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
	parent *frequencyNode
}

type frequencyNode struct {
	next, prev *frequencyNode
	items      set
	value      float64
}

type frequencyList struct {
	first, last *frequencyNode
}

// newFrequencyNode creates a new frequency node with an access frequency value of 0
func newFrequencyNode() *frequencyNode {
	n := &frequencyNode{
		items: make(set),
	}
	n.prev = n
	n.next = n
	return n
}

// s/getNewNode/newNode
func getNewNode(v float64, prev, next *frequencyNode) *frequencyNode {
	nn := newFrequencyNode()
	nn.value = v
	nn.prev = prev
	nn.next = next
	prev.next = nn
	next.prev = nn
	return nn
}

func unlink(n *frequencyNode) {
	n.prev.next = n.next
	n.next.prev = n.prev
}
