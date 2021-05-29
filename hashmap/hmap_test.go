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

package hashmap

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHmap(t *testing.T) {
	req := require.New(t)

	tests := []struct {
		k, v interface{}
	}{
		{1, "1"},
		{2, "2"},
		{3, "3"},
		{"4", 4},
		{"5", 5},
		{"6", 6},
		{"a", []byte("a")},
		{"b", []byte("b")},
		{"c", []byte("c")},
	}

	m := New()
	for i := range tests {
		m.Set(tests[i].k, tests[i].v)
	}
	req.Equal(len(tests), m.Len())
	for i := range tests {
		v, ok := m.Get(tests[i].k)
		req.True(ok)
		req.Equal(tests[i].v, v)
	}

	// test non-existence
	nxTests := []interface{}{4, "7", "d"}
	for i := range nxTests {
		v, ok := m.Get(nxTests[i])
		req.False(ok)
		req.Nil(v)
	}

	// test delete
	m.Del(tests[6].k)
	req.Equal(len(tests)-1, m.Len())
	for i := range tests {
		v, ok := m.Get(tests[i].k)
		if i != 6 {
			req.True(ok)
			req.Equal(tests[i].v, v)
		} else {
			req.False(ok)
			req.Nil(v)
		}
	}

	// add another 10000 keys
	for i := 4; i < 10004; i++ {
		m.Set(i, i*i)
	}

	// test Range()
	m.Lock()
	var (
		total, match int
	)
	for k, v, ok := m.Next(); ok; k, v, ok = m.Next() {
		for i := range tests {
			if k == tests[i].k {
				req.Equal(tests[i].v, v)
				match++
			}
		}
		total++
	}
	k, v, ok := m.Next()
	req.False(ok)
	req.Nil(k)
	req.Nil(v)
	m.Unlock()
	req.Equal(len(tests)-1, match)
	req.Equal(10000+len(tests)-1, total)
	m.info()
}
