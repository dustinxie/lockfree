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
	node struct {
		val unsafe.Pointer
		nxt unsafe.Pointer
	}
)

func (n *node) value() unsafe.Pointer {
	return atomic.LoadPointer(&n.val)
}

func (n *node) next() *node {
	return (*node)(atomic.LoadPointer(&n.nxt))
}

func (n *node) casNext(expected, target unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(&n.nxt, expected, target)
}

func casAddr(addr *unsafe.Pointer, expected, target unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(addr, expected, target)
}
