# lockfree
Golang lock-free concurrent Hashmap

[![LICENSE](https://img.shields.io/badge/License-Apache%202.0-turquise.svg)](LICENSE)
[![Go version](https://img.shields.io/badge/Go-1.14.4-turquise.svg)]()
[![Go Report card](https://goreportcard.com/badge/github.com/dustinxie/lockfree)](https://goreportcard.com/report/github.com/dustinxie/lockfree)
[![Go Reference](https://pkg.go.dev/badge/github.com/dustinxie/lockfree.svg)](https://pkg.go.dev/github.com/dustinxie/lockfree)
---
## Features
- Safe concurrent access by multiple threads/go-routines
- Allow different key types in the same map
- 3x faster than `sync.Map`

## How to use
```go
package anyname

import "github.com/dustinxie/lockfree"

func main() {
	m := lockfree.NewHashMap()
	
	// set
	m.Set(1, "one")
	m.Set("one", 1)
	
	// get
	s, ok := m.Get(1)     // s.(string) = "one", ok = true
	i, ok := m.Get("one") // i.(int) = 1, ok = true
	
	// delete
	m.Del(1)
	m.Del("one")
	
	// can have multiple go-routines/threads call m.Set/Get/Del
}
```

### BucketSizeOption
You can specify a preferred average bucket size when creating the map, a smaller
size gives faster access speed with more memory. While a larger size costs less 
memory but gives slower access speed. See the benchmark section below for details.

```
func main() {
	m := lockfree.NewHashMap(BucketSizeOption(16))
}
```
The default bucket size is 24 if a BucketSizeOption is not set.

### for k, v := range
Since this is a concurrent hashmap, you'll need to call `Lock()` before doing a
range operation. And remember to call `Unlock()` afterwards.
```
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
 
## Benchmark
The benchmark program starts 10 go-routines, each would `Set()` 10,000 keys,
`Get()` these 10,000 keys, and finally `Del()` them, for a total of 100,000
keys, and 300,000 map operations.

Here's the timing and memory allocation data (2.2GHz Intel Core i7, 16GB
1600MHz DDR3 RAM):

```
BenchmarkLockfreeHashMap-8      27      42189873 ns/op      9935506 B/op     703336 allocs/op
BenchmarkMapAndRWMutex-8         9     115930416 ns/op      5752789 B/op       3916 allocs/op
BenchmarkSyncMap-8               8     127551298 ns/op     13686902 B/op     503735 allocs/op
```
with bucket size = 16 (faster time, more memory)
```
BenchmarkLockfreeHashMap-8      27      41867046 ns/op     10279584 B/op     707433 allocs/op
```
Couple of observations:
1. The lockfree hashmap is 3x faster than `sync.Map`, and costs 37% less memory
2. The decrease in time (and increase in memory) of using bucket size 16 vs. 24
is very minimal, less than 4%
3. Golang's native map + RWMutex to synchronize access is even slightly faster
than `sync.Map`, and costs least amount of memory

## Tech details
For technical details of the implementation, click [here](/technical.md)
