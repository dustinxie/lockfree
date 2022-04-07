# lockfree
Golang lock-free concurrent Hashmap

[![LICENSE](https://img.shields.io/badge/License-Apache%202.0-turquise.svg)](LICENSE)
[![Go version](https://img.shields.io/badge/Go-1.14.4-turquise.svg)]()
[![Go Report card](https://goreportcard.com/badge/github.com/dustinxie/lockfree)](https://goreportcard.com/report/github.com/dustinxie/lockfree)
[![Go Reference](https://pkg.go.dev/badge/github.com/dustinxie/lockfree.svg)](https://pkg.go.dev/github.com/dustinxie/lockfree)
---
# Table of Contents
- [Overview](#overview)
  * [Hashmap](#hashmap)
  * [Queue](#queue)
  * [Stack](#stack)
- [Benchmark](#benchmark)

# Overview
Golang's native data structures (such as map, List) are not designed to be
thread-safe at first place. A common solution is to use mutex to synchronize
access to data that are shared by multiple threads. This is however not very
efficient due to multiple threads contending a single read/write mutex and in
general will lead to performance loss in timing.

A more efficient alternative is to leverage the [CAS](https://en.wikipedia.org/wiki/Compare-and-swap)
technique to remove (or significantly reduce) the use of mutex. Several data
structures are implemented using this idea.

Benchmark results show that the CAS solution is generally 2~3x times faster as
compared to solution using mutex.

For technical details of the implementation, click [here](/technical.md) for an
explanation of the concurrent hashmap.

## Hashmap
- can be concurrently accessed
- allows different key types in the same map
```go
package anyname

import "github.com/dustinxie/lockfree"

func main() {
	m := lockfree.NewHashMap()
	
	// set
	m.Set(1, "one")
	m.Set("one", 1)
	size := m.Len() // size = 2
	
	// get
	s, ok := m.Get(1)     // s.(string) = "one", ok = true
	i, ok := m.Get("one") // i.(int) = 1, ok = true
	
	// delete
	m.Del(1)
	m.Del("one")
	size = m.Len() // map is empty, size = 0
	
	// can have multiple threads/go-routines call m.Set/Get/Del
}
```

### BucketSizeOption
You can specify a preferred average bucket size when creating the map, a smaller
size gives faster access speed with more memory. While a larger size costs less 
memory but gives slower access speed. See the benchmark section below for details.

```
import (
    "github.com/dustinxie/lockfree"
    "github.com/dustinxie/lockfree/hashmap"
)

func main() {
	m := lockfree.NewHashMap(hashmap.BucketSizeOption(16))
}
```
The default bucket size is 24 if a BucketSizeOption is not set.

### for k, v := range
Since this is a concurrent hashmap, you'll need to call `Lock()` before doing a
range operation. And remember to call `Unlock()` afterwards.
```go
func rangeMap(m HashMap) error {
	f := func(interface{}, interface{}) error {
		// define what to do for each entry in the map
		return
	}

	m.Lock()
	for k, v, ok := m.Next(); ok; k, v, ok = m.Next() {
	    if err := f(k, v); err != nil {
	        // unlock the map before return, otherwise it will deadlock
	        m.Unlock()
	        return err
	    } 
	}
	m.Unlock()
	return nil
}
```
A shortcut:
```go
m.Iterate(func(_k interface{}, _v interface{}) error {
	k := _k.(string)
	v := _v.(int64)
	// do something
	return nil
})
```

## Queue
- FIFO list that can be concurrently accessed
- can put different data types into the queue
```go
package anyname

import "github.com/dustinxie/lockfree"

func main() {
	q := lockfree.NewQueue()
	
	// add to the queue
	q.Enque(1)
	q.Enque("one")
	size := q.Len() // size = 2
	
	// remove from queue
	i := q.Deque() // i.(int) = 1
	s := q.Deque() // s.(string) = "one"
	
	size = q.Len()  // queue is empty, size = 0
	s = q.Deque() // queue is empty, s = nil
	
	// can have multiple threads/go-routines call q.Enque/Deque
}
```

## Stack
- LIFO list that can be concurrently accessed
- can put different data types into the stack
```go
package anyname

import "github.com/dustinxie/lockfree"

func main() {
	s := lockfree.NewStack()
	
	// add to the stack
	s.Push(1)
	s.Push("one")
	size := s.Len() // size = 2
	
	// peek the top of stack
	str := s.Pop() // str.(string) = "one"
	
	// remove from stack
	str = s.Pop() // str.(string) = "one"
	i := s.Pop()   // i.(int) = 1
	
	size = s.Len() // queue is empty, size = 0
	str = s.Pop()  // queue is empty, s = nil
	
	// can have multiple threads/go-routines call s.Push/Pop/Peek
}
```

# Benchmark
The benchmark program starts 10 go-routines, each would perform a certain set
of tasks concurrently. Tests are run on a machine with following config:

| Hardware | Spec |
| --- | --- |
| CPU | Intel Core i7-4770HQ @ 2.20GHz |
| Core | 4 |
| L2 cache (per Core) | 256kB |
| L3 cache | 6MB |
| RAM | 16GB 1600MHz DDR3 RAM |
| Storage | 256GB SSD |

## Benchmark Hashmap
Task set for each concurrent thread is to `Set()` 10,000 keys, `Get()` these
10,000 keys, and finally `Del()` all these keys.

The total workload (10 threads) amounts to 100,000 keys, and 300,000 map
operations.

Here's the timing and memory allocation data comparison between the lockfree
map, `sync.Map` and golang's map + RWMutex:
```
BenchmarkLockfreeHashMap-8      27       42189873 ns/op      9935506 B/op     703336 allocs/op
BenchmarkMapAndRWMutex-8         9      115930416 ns/op      5752789 B/op       3916 allocs/op
BenchmarkSyncMap-8               8      127551298 ns/op     13686902 B/op     503735 allocs/op
```
with bucket size = 16 (faster time, more memory)
```
BenchmarkLockfreeHashMap-8      27       41867046 ns/op     10279584 B/op     707433 allocs/op
```
Couple of observations:
1. The lockfree hashmap is 3x as fast as `sync.Map`, and costs 37% less memory
2. The decrease in time (and increase in memory) of using bucket size 16 vs. 24
is very minimal, less than 4%
3. Golang's native map + RWMutex to synchronize access is even slightly faster
than `sync.Map`, and costs least amount of memory

## Benchmark Queue
Task set for each concurrent thread is to `Enque()` 10,000 items, then `Deque()`
these 10,000 items.

The total workload (10 threads) amounts to 100,000 items, and 200,000 queue
operations.

Here's the timing and memory allocation data comparison between the lockfree
queue and golang's list + RWMutex:

```
BenchmarkLockfreeQueue-8          42      24980347 ns/op      3999350 B/op     299750 allocs/op
BenchmarkQueueAndRWMutex-8        27      43929601 ns/op      5599024 B/op     199750 allocs/op
```
The lockfree queue has better performance in both timing and memory.

## Benchmark Stack
Task set for each concurrent thread is to `Push()` 10,000 items, then `Pop()`
these 10,000 items.

The total workload (10 threads) amounts to 100,000 items, and 200,000 stack
operations.

Here's the timing and memory allocation data comparison between the lockfree
stack and golang's list + RWMutex:

```
BenchmarkLockfreeStack-8          34      31400698 ns/op      3999327 B/op     299751 allocs/op
BenchmarkStackAndRWMutex-8        28      43869583 ns/op      5598690 B/op     199750 allocs/op
```
The lockfree stack has better performance in both timing and memory.
