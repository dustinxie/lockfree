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
	stack struct {
		count uint64
		head  *node
	}
)

// NewStack creates a new stack
func NewStack() *stack {
	var empty interface{}
	return &stack{
		head: &node{val: unsafe.Pointer(&empty)},
	}
}

func (s *stack) Len() int {
	return int(atomic.LoadUint64(&s.count))
}

func (s *stack) Push(v interface{}) {
	n := node{
		val: unsafe.Pointer(&v),
	}
	headAddr := (*unsafe.Pointer)(unsafe.Pointer(&s.head))
	for {
		head := atomic.LoadPointer(headAddr)
		n.nxt = head
		if casAddr(headAddr, head, unsafe.Pointer(&n)) {
			atomic.AddUint64(&s.count, 1)
			return
		}
	}
}

func (s *stack) Pop() interface{} {
	headAddr := (*unsafe.Pointer)(unsafe.Pointer(&s.head))
	for {
		head := (*node)(atomic.LoadPointer(headAddr))
		n := head.next()
		if n == nil {
			return nil
		}
		if casAddr(headAddr, unsafe.Pointer(head), unsafe.Pointer(n)) {
			atomic.AddUint64(&s.count, ^uint64(0))
			return *(*interface{})(head.value())
		}
	}
}

func (s *stack) Peek() interface{} {
	return *(*interface{})(s.head.value())
}
