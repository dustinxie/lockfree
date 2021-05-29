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
	"math"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
)

func TestBucket(t *testing.T) {
	req := require.New(t)

	b := newBucket(0, 0)
	req.Nil(b.fence.next())
	b.fence.linkTo(newFence())
	req.Equal(&b.fence, b.last())

	tests := []struct {
		hash uint64
		k, v interface{}
	}{
		{1, "1", 1},
		{10, "2", 2},
		{10, "3", 3},
		{10, "4", 4},
		{20, "5", 5},
		{math.MaxUint64, "6", 6},
		{math.MaxUint64, "7", 7},
	}

	for i := range tests {
		req.True(b.upsert(&hashNode{
			hash: tests[i].hash,
			key:  unsafe.Pointer(&tests[i].k),
			val:  unsafe.Pointer(&tests[i].v),
		}))
	}

	req.EqualValues(len(tests), b.count)
	last := b.last()
	req.Equal(tests[len(tests)-1].hash, last.hash)
	req.Equal(tests[len(tests)-1].k, *(*interface{})(last.key))
	req.Equal(tests[len(tests)-1].v, *(*interface{})(last.val))
	for i := range tests {
		v, ok := b.get(tests[i].k, tests[i].hash)
		req.True(ok)
		req.Equal(tests[i].v, v)
	}

	searchTests := []struct {
		hash   uint64
		k, v   interface{}
		curr   int
		next   int
		insert bool
	}{
		{0, "0", 0, -1, 0, true},
		{1, "1", 11, -1, 0, false},
		{1, "collision", 12, 0, 1, true},
		{3, "new", 13, 0, 1, true},
		{10, "2", 22, 0, 1, false},
		{10, "3", 33, 1, 2, false},
		{10, "4", 44, 2, 3, false},
		{10, "collision", 15, 3, 4, true},
		{11, "new", 16, 3, 4, true},
		{20, "5", 55, 3, 4, false},
		{20, "collision", 23, 4, 5, true},
		{27, "new", 27, 4, 5, true},
		{math.MaxUint64, "6", 66, 4, 5, false},
		{math.MaxUint64, "7", 77, 5, 6, false},
		{math.MaxUint64, "new", 88, 6, -1, true},
	}

	// test search
	node := hashNode{}
	for i := range searchTests {
		node.hash = searchTests[i].hash
		node.key = unsafe.Pointer(&searchTests[i].k)
		curr, next, insert := b.search(&node)

		if c := searchTests[i].curr; c == -1 {
			req.True(isFence(curr))
		} else {
			req.Equal(tests[c].hash, curr.hash)
			req.Equal(tests[c].k, *(*interface{})(curr.key))
		}

		if n := searchTests[i].next; n == -1 {
			req.True(isFence(next))
		} else {
			req.Equal(tests[n].hash, next.hash)
			req.Equal(tests[n].k, *(*interface{})(next.key))
		}
		req.Equal(searchTests[i].insert, insert)
	}

	for i := range searchTests {
		req.Equal(searchTests[i].insert, b.upsert(&hashNode{
			hash: searchTests[i].hash,
			key:  unsafe.Pointer(&searchTests[i].k),
			val:  unsafe.Pointer(&searchTests[i].v),
		}))
	}

	// test pivot
	splitTests := []struct {
		hash  uint64
		curr  int
		count uint32
	}{
		{0, -1, 0},
		{1, 0, 1},
		{3, 2, 3},
		{10, 3, 4},
		{11, 7, 8},
		{20, 8, 9},
		{27, 10, 11},
		{math.MaxUint64, 11, 12},
	}

	for _, v := range splitTests {
		curr, _, count := b.pivot(v.hash)
		req.Equal(v.count, count)
		if count == 0 {
			req.True(isFence(curr))
		} else {
			req.Equal(searchTests[v.curr].hash, curr.hash)
			req.Equal(searchTests[v.curr].k, *(*interface{})(curr.key))
		}
	}

	// test split
	pivot := splitTests[7].hash
	b1 := b.split(pivot)
	req.Equal(splitTests[7].count, b.count)
	req.Equal(uint32(len(searchTests))-b.count, b1.count)

	// test delete
	node = hashNode{
		hash: searchTests[3].hash,
		key:  unsafe.Pointer(&searchTests[2].k),
	}
	req.False(b.del(&node))
	node.hash = searchTests[2].hash
	req.True(b.del(&node))
	req.Equal(splitTests[7].count-1, b.count)

	// final count
	var (
		v  interface{}
		ok bool
	)
	for i := range searchTests {
		if hash := searchTests[i].hash; hash < pivot {
			v, ok = b.get(searchTests[i].k, hash)
		} else {
			v, ok = b1.get(searchTests[i].k, hash)
		}
		if i != 2 {
			req.True(ok)
			req.Equal(searchTests[i].v, v)
		} else {
			req.False(ok)
			req.Nil(v)
		}
	}
}
