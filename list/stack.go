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
	for {
		tip := atomic.LoadPointer(&s.head.nxt)
		n.nxt = tip
		if s.head.casNext(tip, unsafe.Pointer(&n)) {
			atomic.AddUint64(&s.count, 1)
			return
		}
	}
}

func (s *stack) Pop() interface{} {
	for {
		tip := s.head.next()
		if tip == nil {
			return nil
		}
		nn := tip.next()
		if s.head.casNext(unsafe.Pointer(tip), unsafe.Pointer(nn)) {
			atomic.AddUint64(&s.count, ^uint64(0))
			return *(*interface{})(tip.value())
		}
	}
}

func (s *stack) Peek() interface{} {
	tip := s.head.next()
	if tip != nil {
		return *(*interface{})(tip.value())
	}
	return nil
}
