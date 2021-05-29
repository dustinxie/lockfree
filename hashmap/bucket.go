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

package hashmap

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

type bucket struct {
	sync.RWMutex
	count uint32
	fence hashNode // dummy hashNode that marks beginning of a bucket
}

func newBucket(count uint32, hash uint64) *bucket {
	return &bucket{
		count: count,
		fence: hashNode{hash: hash},
	}
}

func (b *bucket) size() uint32 {
	return atomic.LoadUint32(&b.count)
}

func (b *bucket) get(key interface{}, hash uint64) (interface{}, bool) {
	b.RLock()
	defer b.RUnlock()
	// running into the next fence hashNode means we exhausted all nodes in this bucket
	for curr := b.fence.next(); !isFence(curr); curr = curr.next() {
		if hash == curr.hash && key == *(*interface{})(curr.key) {
			return *(*interface{})(curr.value()), true
		}
	}
	return nil, false
}

// last return the last node in the bucket
func (b *bucket) last() *hashNode {
	curr := &b.fence
	for next := curr.next(); !isFence(next); {
		curr = next
		next = next.next()
	}
	return curr
}

func (b *bucket) upsert(node *hashNode) bool {
	b.RLock()
	defer b.RUnlock()
	for {
		curr, next, insert := b.search(node)
		if insert {
			node.linkTo(next)
			// insert the new hashNode, curr --> node --> next
			if curr.casNext(node.nxt, unsafe.Pointer(node)) {
				atomic.AddUint32(&b.count, 1)
				return true
			}
		} else {
			val := next.value()
			// update the new value
			if next.casValue(val, node.val) {
				return false
			}
		}
	}
}

func (b *bucket) del(node *hashNode) bool {
	b.Lock()
	defer b.Unlock()
	curr, next, insert := b.search(node)
	if insert {
		return false
	}
	curr.nxt = nil
	curr.nxt = next.nxt
	atomic.AddUint32(&b.count, ^uint32(0))
	return true
}

// search finds the position to insert or update the key
func (b *bucket) search(node *hashNode) (*hashNode, *hashNode, bool) {
	var (
		hash          = node.hash
		curr, next, _ = b.pivot(hash)
	)
	for ; hash == next.hash && !isFence(next); curr, next = next, next.next() {
		if *(*interface{})(node.key) == *(*interface{})(next.key) {
			// next is the node to be updated or deleted
			return curr, next, false
		}
	}
	return curr, next, true
}

// pivot returns the node with hash < input, and number of such nodes
func (b *bucket) pivot(hash uint64) (*hashNode, *hashNode, uint32) {
	var (
		curr  = &b.fence
		next  = curr.next()
		count uint32
	)
	for ; hash > next.hash; count++ {
		curr = next
		next = next.next()
	}
	return curr, next, count
}

// split breaks the bucket at the given hash, and returns the new bucket
func (b *bucket) split(hash uint64) *bucket {
	b.Lock()
	curr, next, count := b.pivot(hash)
	b1 := newBucket(b.count-count, hash)
	b1.fence.linkTo(next)
	b.count = count
	curr.linkTo(&b1.fence)
	b.Unlock()
	return b1
}

// merge merges 2 buckets into 1
func (b *bucket) merge(b1 *bucket) {
	b.Lock()
	b1.Lock()
	b.count += b1.count
	b.last().linkTo(b1.fence.next())
	b1.Unlock()
	b.Unlock()
}
