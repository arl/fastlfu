# fastLFU cache


:warning: `fastlfu` is still under developement, the API might not be stable yet, you should not use in production :exclamation:

This is a Go 1.18 implementation of the **fastLFU** (Least Frequently Used)
cache eviction scheme, described in a 2021 [paper](https://arxiv.org/pdf/2110.11602v1.pdf)
by Dhruv Matani, Ketan Shah and Anirban Mitra called _An O(1) algorithm for implementing the LFU cache eviction scheme_.

Here's an extract from the paper introduction: 

> The LFU algorithm has behaviour desirable by many real world workloads. However, in many places, the LRU algorithm is is preferred over the LFU algorithm because of its lower run time complexity of O(1) versus O(logn). We present here an LFU cache eviction algorithm that has a runtime complexity of O(1) for all of its operations, which include insertion, access and deletion(eviction).

## Usage

Install the `fastlfu` module:

```
go get github.com/arl/fastlfu@latest
```

Let's create a `Cache` object where keys are `uint64` and values are `string`s:

```go
    c := fastlfu.New[uint64, string]()
```

Add a bunch of key-value pairs:

```go
    c.Insert(0, "foo")
    c.Insert(1, "bar")
    c.Insert(2, "baz")
```

Each time we fetch a value associated from a key, we increment the key
frequency:

```go
    v1, ok1 := c.Fetch(0)
    // returns "foo", true

    v2, ok1 := c.Fetch(1)
    // returns "bar", true
```

If we call `Fetch` with a key that is not present, we get the zero-value of the
value type, and false:

```go
    v, ok := c.Fetch(12345678)
    // returns "", false
```

After these calls, the `(0, "foo")` and `(1, "bar")` key-value pairs have been
more frequently used than `(2, "baz").`

So if we require the cache to evict the least frequently used value, it's going
to be `"baz"`.  `Evict` also returns a boolean indicating if any eviction has
been performed. 

```go
    v, ok := c.Evict(0)
    // returns "baz", true
```

NOTE: `Evict` is guaranteed to succeed unless the `Cache` is empty.

If we call `Evict` again, we can't predict they key-value pair that is going to
be evicted because the remaining pairs have the same frequency.

```go
    v, ok := c.Evict(0)
    // returns either ("foo", true) or ("bar", true)
```

## Performance

TODO

## [MIT LICENSE](./LICENSE.md)
