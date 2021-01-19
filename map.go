// Copyright 2020 dustinxie
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lockfree

import (
	"crypto/rand"
	"encoding/binary"
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	bucketOverflow  = 32
	bucketUnderflow = 10
)

type (
	// HashMap is a map[key]value
	HashMap interface {
		// len(map)
		Len() int

		// v, ok := map[key]
		Get(key interface{}) (interface{}, bool)

		// map[key] = value
		Set(key, value interface{})

		// delete(map, key)
		Del(key interface{})
	}

	// Hash64 returns 64-bit hash
	Hash64 interface {
		Sum64() uint64
	}

	hmap struct {
		sync.RWMutex
		B       uint32    // log_2 of number of buckets (can hold up to loadFactor * 2^B items)
		count   uint64    // number of items in the map
		k0, k1  uint64    // hash seed
		buckets []*bucket // array of 2^B Buckets
	}
)

// NewHashMap creates a new hashmap
func NewHashMap() HashMap {
	h := hmap{
		buckets: make([]*bucket, 1),
	}

	// generate 2 random seeds
	binary.Read(rand.Reader, binary.BigEndian, &h.k0)
	binary.Read(rand.Reader, binary.BigEndian, &h.k1)

	// create the very first bucket
	h.buckets[0] = newBucket(0, 0)
	h.buckets[0].fence.linkTo(newFence())
	return &h
}

func (h *hmap) Len() int {
	return int(atomic.LoadUint64(&h.count))
}

func (h *hmap) Get(key interface{}) (interface{}, bool) {
	hash := h.hash(key)
	return h.getBucket(hash).get(key, hash)
}

func (h *hmap) Set(key, value interface{}) {
	hash := h.hash(key)
	node := hashNode{
		hash: hash,
		key:  unsafe.Pointer(&key),
		val:  unsafe.Pointer(&value),
	}
	if h.getBucket(hash).upsert(&node) {
		atomic.AddUint64(&h.count, 1)
	}

	if h.isOverflow() {
		h.expand()
	}
}

func (h *hmap) isOverflow() bool {
	return atomic.LoadUint64(&h.count)>>atomic.LoadUint32(&h.B) > bucketOverflow
}

func (h *hmap) Del(key interface{}) {
	hash := h.hash(key)
	node := hashNode{
		hash: hash,
		key:  unsafe.Pointer(&key),
	}
	if h.getBucket(hash).del(&node) {
		atomic.AddUint64(&h.count, ^uint64(0))
	}

	if h.isUnderflow() {
		h.shrink()
	}
}

func (h *hmap) isUnderflow() bool {
	B := atomic.LoadUint32(&h.B)
	return B > 5 && (atomic.LoadUint64(&h.count)>>B) < bucketUnderflow
}

func (h *hmap) getBucket(hash uint64) *bucket {
	h.RLock()
	b := h.buckets[hash>>(64-h.B)]
	h.RUnlock()
	return b
}

func (h *hmap) expand() {
	h.Lock()
	defer h.Unlock()
	if !h.isOverflow() {
		return
	}

	// double the buckets list
	h.buckets = append(h.buckets, h.buckets...)

	// move i-th item to 2i-th position --> [00, x, 01, x, 10, x, 11, x]
	// then split the buckets
	// [00, x, 01, x, 10, x, 11, x] --> [000, 001, 010, 011, 100, 101, 110, 111]
	atomic.AddUint32(&h.B, 1)
	for i := len(h.buckets)/2 - 1; i >= 0; i-- {
		if i != 0 {
			h.buckets[2*i] = nil
			h.buckets[2*i] = h.buckets[i]
		}
		h.buckets[2*i+1] = nil
		h.buckets[2*i+1] = h.buckets[2*i].split(uint64(2*i+1) << (64 - h.B))
	}
}

func (h *hmap) shrink() {
	h.Lock()
	defer h.Unlock()
	if !h.isUnderflow() {
		return
	}

	// merge the buckets
	// [000, 001, 010, 011, 100, 101, 110, 111] --> [00, x, 01, x, 10, x, 11, x]
	// then halve the list
	// [00, x, 01, x, 10, x, 11, x] --> [00, 01, 10, 11]
	half := len(h.buckets) / 2
	for i := 0; i < half; i++ {
		h.buckets[2*i].merge(h.buckets[2*i+1])
		h.buckets[2*i+1] = nil
		if i != 0 {
			h.buckets[i] = nil
			h.buckets[i] = h.buckets[2*i]
		}
	}
	atomic.AddUint32(&h.B, ^uint32(0))
	h.buckets = h.buckets[:half]
}

func (h *hmap) info() {
	var count, min, max uint32
	min = 1<<32 - 1
	for i := 0; i < (1 << h.B); i++ {
		b := h.buckets[i]
		count += b.count
		if b.count < min {
			min = b.count
		}
		if b.count > max {
			max = b.count
		}
	}
	println("++==========================")
	println("|| total key count =", h.count)
	println("|| bucket number =", 1<<h.B)
	println("|| key per bucket =", h.count>>h.B)
	println("|| total key count =", count)
	println("|| min keys per bucket =", min)
	println("|| max keys per bucket =", max)
	println("++==========================")
}
