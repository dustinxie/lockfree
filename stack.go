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

package lockfree

import (
	"github.com/dustinxie/lockfree/list"
)

type (
	// Stack is a LIFO list
	Stack interface {
		// length of stack
		Len() int

		// add an item to the stack
		Push(interface{})

		// remove an item from the stack
		Pop() interface{}

		// return (but not remove) the top item on the stack
		Peek() interface{}
	}
)

// NewStack creates a new stack
func NewStack() Stack {
	return list.NewStack()
}
