# Memory allocation and Performance in Golang maps

Answer to a [question](https://stackoverflow.com/questions/65258003/memory-allocation-of-mapintinterface-vs-mapintstruct) (stackoverflow) related to memory allocation for `map[int]interface{}` and `map[int]struct{}` in Golang. 

Medium post: [Memory Allocation and Performance in Golang Maps
](https://ricardoerikson.medium.com/memory-allocation-and-performance-in-golang-maps-b267b5ad9217)

# TL;DR

The internal design of maps in Golang is highly optimized for performance and memory management. Maps keep track of keys and values that can hold pointers. If the entries in a bucket can't hold pointers, maps just create overflow buckets to avoid unnecessary overhead with GC, which results in more allocations (the case of `map[int]struct{}`).

# Long answer

We need to understand map initialization and map structure. Then, we will analyze the benchmarks.

## Map initialization

There are two methods for map initialization:

- `make(map[int]string)` when we have no idea about how many entries will be added.
- `make(map[int]string, hint)` when we have an idea about how many entries will be added. `hint` is an estimate of the initial capacity.

Maps are mutable and they will grow on-demand, no matter which initialization method we choose. However, the second method pre-allocates memory for at least `hint` entries, which results in increased performance.

## Map structure

A map in Go is a hash table that stores its key/value pairs into buckets. Each bucket is an array that holds up to 8 entries. The default number of buckets is 1. Once the number of entries across each bucket reaches an average load of buckets (aka load factor), the map gets bigger by doubling the number of buckets. Every time a map grows, it allocates memory for the new-coming entries. In practice, every time the load of the buckets reaches 6.5 or more, the map grows.

Behind the scenes, a map is a pointer to the `hmap` struct. There is also the `maptype` struct, which holds some information about the type of the `map`. The source code for map can be found here:

[https://github.com/golang/go/blob/master/src/runtime/map.go](https://github.com/golang/go/blob/master/src/runtime/map.go)

Below you can find some insights on how to hack the `map` type and how to see a map growing:
- [https://hackernoon.com/some-insights-on-maps-in-golang-rm5v3ywh](https://hackernoon.com/some-insights-on-maps-in-golang-rm5v3ywh)
- [https://play.golang.org/p/NaoC8fkmy9x](https://play.golang.org/p/NaoC8fkmy9x)

**One important thing to note is that maps keep track of keys and values that can hold pointers. If the entries in a bucket can't hold any pointers, the bucket is marked as containing no pointers and maps just create overflow buckets (which means more memory allocations). This avoids unnecessary overhead with GC. See this [comment in the `mapextra` struct (line 132)](https://github.com/golang/go/blob/682a1d2176b02337460aeede0ff9e49429525195/src/runtime/map.go#L132) and this [post](https://www.komu.engineer/blogs/go-gc-maps) for reference.**

## Benchmarks

An empty struct `struct{}` has no fields and cannot hold any pointers. As a result, the buckets in the empty struct case will be marked as *containing no pointers* and we can expect more memory allocations for a map of type `map[int]struct{}` as it grows. On the other hand, `interface{}` can hold any values, including pointers. Map buckets keep track of the size of memory prefix holding pointers ([`ptrdata` field, line 33](https://github.com/golang/go/blob/cd99385ff4a4b7534c71bb92420da6f462c5598e/src/runtime/type.go#L33)) to decide if more overflow buckets will be created ([map.go, line 265](https://github.com/golang/go/blob/b634f5d97a6e65f19057c00ed2095a1a872c7fa8/src/runtime/map.go#L265)). Refer to this [link](https://play.golang.org/p/_-QKWu1GBnr) to see the size of memory prefix holding all pointers for `map[int]struct{}` and `map[int]interface{}`.

The difference between the two benchmarks (`Benchmark_EmptyStruct` and `Benchmark_Interface`) is clear when we see the CPU profile. `Benchmark_Interface` does not have the `(*hmap) createOverflow` method that results in an additional memory allocation flow:

### Benchmark_EmptyStruct CPU profile

![](https://raw.githubusercontent.com/ricardoerikson/benchmark-golang-maps/main/map_empty_struct_cpu_profile.png)
*Benchmark_EmptyStruct CPU profile [[png](https://raw.githubusercontent.com/ricardoerikson/benchmark-golang-maps/main/map_empty_struct_cpu_profile.png), [svg](https://raw.githubusercontent.com/ricardoerikson/benchmark-golang-maps/main/map_empty_struct_cpu_profile.svg)]*

### Benchmark_Interface CPU profile

![](https://raw.githubusercontent.com/ricardoerikson/benchmark-golang-maps/main/map_interface_cpu_profile.png)
*Benchmark_Interface CPU profile [[png](https://raw.githubusercontent.com/ricardoerikson/benchmark-golang-maps/main/map_interface_cpu_profile.png), [svg](https://raw.githubusercontent.com/ricardoerikson/benchmark-golang-maps/main/map_interface_cpu_profile.svg)]*

I customized the tests to pass the number of entries and the map's initial capacity (hint). Here are the results of the executions. Results are basically the same when there are few entries or when the initial capacity is greater than the number of entries. If you have many entries with an initial capacity of 0, you will get quite a different number for allocations.

| Benchmark   | Entries | InitialCapacity |      Speed |  Bytes/op | Allocations/op |
| ----------- | ------: | --------------: | ---------: | --------: | -------------: |
| EmptyStruct |       7 |               0 |  115 ns/op |    0 B/op |    0 allocs/op |
| Interface   |       7 |               0 | 94.8 ns/op |    0 B/op |    0 allocs/op |
| EmptyStruct |       8 |               0 |  114 ns/op |    0 B/op |    0 allocs/op |
| Interface   |       8 |               0 |  110 ns/op |    0 B/op |    0 allocs/op |
| EmptyStruct |       9 |               0 |  339 ns/op |  160 B/op |    1 allocs/op |
| Interface   |       9 |               0 |  439 ns/op |  416 B/op |    1 allocs/op |
| EmptyStruct |      16 |              16 |  444 ns/op |  324 B/op |    1 allocs/op |
| Interface   |      16 |              16 |  586 ns/op |  902 B/op |    1 allocs/op |
| EmptyStruct |      16 |              32 |  448 ns/op |  640 B/op |    1 allocs/op |
| Interface   |      16 |              32 |  724 ns/op | 1792 B/op |    1 allocs/op |
| EmptyStruct |      16 |             100 |  634 ns/op | 1440 B/op |    2 allocs/op |
| Interface   |      16 |             100 | 1241 ns/op | 4128 B/op |    2 allocs/op |
| EmptyStruct |     100 |               0 | 5339 ns/op | 3071 B/op |   17 allocs/op |
| Interface   |     100 |               0 | 6524 ns/op | 7824 B/op |    7 allocs/op |
| EmptyStruct |     100 |             128 | 2665 ns/op | 3109 B/op |    2 allocs/op |
| Interface   |     100 |             128 | 3938 ns/op | 8224 B/op |    2 allocs/op |

## Conclusion

There's nothing wrong with the benchmark methodology. It's all related to map optimization for performance and memory management. The benchmarks show the average value per iteration. Maps of type `map[int]interface{}` are slower because they suffer performance degradation when GC scans the buckets that can hold pointers. Maps of type `map[int]struct{}` use less memory because they, in fact, use less memory (`Test_EmptyStructValueSize` shows that `struct{}{}` has zero size). Despite `nil` being the zero value for `interface{}`, this type requires some space to store ANY value (`Test_NilInterfaceValueSize` test shows the size of an `interface{}` holding a `nil` value is not zero). Finally, the empty struct benchmark allocations are higher because the type `map[int]struct{}` requires more overflow buckets (for perfomance optimization) since its buckets don't hold any pointers.

# CPU Profiling Procedure

(Benchmark_Interface only)

```
$ go test -cpuprofile interface.prof -benchmem -bench=^Benchmark_Interface$ -run=^$
$ go tool pprof interface.prof
(pprof) web mallocgc
```
