# fastLFU implementation


:warning: `fastlfu` is still under developement, the API might not be stable yet, you should not use in production :exclamation:

This is a Go 1.18 implementation of the **fastLFU** (Least Frequently Used)
cache eviction scheme, described in a 2021 [paper](https://arxiv.org/pdf/2110.11602v1.pdf)
by Dhruv Matani, Ketan Shah and Anirban Mitra called _An O(1) algorithm for implementing the LFU cache eviction scheme_.

Here's the paper introduction: 

> Cache eviction algorithms are used widely in operating systems, databases and other systems that use caches to speed up execution by caching data that is used by the application. There are many policies such as MRU (Most Recently Used), MFU (Most Frequently Used), LRU (Least Recently Used) and LFU (Least Frequently Used) which each have their advantages and drawbacks and are hence used in specific scenarios. By far, the most widely used algorithm is LRU, both for its O(1) speed of operation as well as its close resemblance to the kind of behaviour that is expected by most applications. The LFU algorithm also has behaviour desirable by many real world workloads. However, in many places, the LRU algorithm is is preferred over the LFU algorithm because of its lower run time complexity of O(1) versus O(logn). We present here an LFU cache eviction algorithm that has a runtime complexity of O(1) for all of its operations, which include insertion, access and deletion(eviction).

## Usage

Install the `fastlfu` module:

```
go get github.com/arl/fastlfu
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
    v1, ok1 := c.Fetch(0) // "foo", true
    v2, ok1 := c.Fetch(1) // "bar", true
```

If we call `Fetch` with a key that is not present, we get the zero-value of the
value type, and false:

```go
    v, ok := c.Fetch(12345678) // "", false
```

After these calls, the `(0, "foo")` and `(1, "bar")` key-value pairs have been
more frequently used than `(2, "baz").`

So if we require the cache to evict the least frequently used value, it's going
to be `"baz"`.  `Evict` also returns a boolean indicating if an eviction has
been performed. 

```go
    v, ok := c.Evict(0) // "baz", true
```

NOTE: `Evict` is guaranteed to succeed unless the `Cache` is empty.

If we call `Evict` again, we can't predict they key-value pair that is going to
be evicted because the remaining pairs have the same frequency.

```go
    v, ok := c.Evict(0) // either ("foo", true) or ("bar", true)
```

## Performance

TODO

## LICENSE

MIT License

Copyright (c) 2022 Aurélien Rainone

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
