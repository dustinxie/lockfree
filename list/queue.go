// Copyright 2021 dustinxie
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

package list

import (
	"sync/atomic"
	"unsafe"
)

type (
	queue struct {
		count      uint64
		head, tail *node
	}
)

// NewQueue creates a new queue
func NewQueue() *queue {
	empty := node{}
	return &queue{
		head: &empty,
		tail: &empty,
	}
}

func (q *queue) Len() int {
	return int(atomic.LoadUint64(&q.count))
}

func (q *queue) Enque(v interface{}) {
	n := node{
		val: unsafe.Pointer(&v),
	}
	tailAddr := (*unsafe.Pointer)(unsafe.Pointer(&q.tail))
	for {
		tail := (*node)(atomic.LoadPointer(tailAddr))
		if tail.casNext(nil, unsafe.Pointer(&n)) {
			atomic.StorePointer(tailAddr, unsafe.Pointer(&n))
			atomic.AddUint64(&q.count, 1)
			return
		}
	}
}

func (q *queue) Deque() interface{} {
	headAddr := (*unsafe.Pointer)(unsafe.Pointer(&q.head))
	for {
		head := atomic.LoadPointer(headAddr)
		n := (*node)(head).next()
		if n == nil {
			return nil
		}
		if casAddr(headAddr, head, unsafe.Pointer(n)) {
			atomic.AddUint64(&q.count, ^uint64(0))
			return *(*interface{})(n.value())
		}
	}
}
