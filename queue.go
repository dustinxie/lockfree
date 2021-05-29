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
	// Queue is a FIFO list
	Queue interface {
		// length of queue
		Len() int

		// add an item to the queue
		Enque(interface{})

		// remove an item from the queue
		Deque() interface{}
	}
)

// NewQueue creates a new queue
func NewQueue() Queue {
	return list.NewQueue()
}
