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
	"github.com/dustinxie/lockfree/hashmap"
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

		// call this before for k, v := range map
		Lock()

		// call this after for k, v := range map
		Unlock()

		// returns next <k, v> in the map
		Next() (interface{}, interface{}, bool)
	}
)

// NewHashMap creates a new hashmap
func NewHashMap(opts ...hashmap.Option) HashMap {
	return hashmap.New(opts...)
}
