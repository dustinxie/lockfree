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
	"math"
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	bucketOverflow  = 256
	bucketUnderflow = 85
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
		B:       1, // start from 2 buckets
		buckets: make([]*bucket, 3),
	}

	// generate 2 random seeds
	binary.Read(rand.Reader, binary.BigEndian, &h.k0)
	binary.Read(rand.Reader, binary.BigEndian, &h.k1)

	// init buckets
	for i := 1; i < len(h.buckets); i++ {
		h.buckets[i] = newBucket(0)
		h.buckets[i].fence.linkTo(newFence())
	}
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

	if (atomic.LoadUint64(&h.count) >> atomic.LoadUint32(&h.B)) > bucketOverflow {
		h.expand()
	}
}

func (h *hmap) Del(key interface{}) {
	hash := h.hash(key)
	if h.getBucket(hash).del(key, hash) {
		atomic.AddUint64(&h.count, ^uint64(0))
	}
}

func (h *hmap) getBucket(hash uint64) *bucket {
	h.RLock()
	start := (1 << h.B) - 1
	b := h.buckets[start+int(hash>>(64-h.B))]
	h.RUnlock()
	return b
}

func (h *hmap) expand() {
	h.Lock()
	start := (1 << h.B) - 1
	pivot := uint64(math.MaxUint64)>>(h.B+1) + 1
	for i := 0; i < (1 << h.B); i++ {
		b := h.buckets[start+i]
		h.buckets = append(h.buckets, b, b.split(pivot*(2*uint64(i)+1)))
	}
	atomic.AddUint32(&h.B, 1)
	h.Unlock()
}
