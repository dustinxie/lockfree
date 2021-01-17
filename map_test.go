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
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewHashMap(t *testing.T) {
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

	m := NewHashMap()
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

	// test 4 threads
	wg := sync.WaitGroup{}
	wg.Add(4)
	for i := 0; i < 4; i++ {
		go func(start, end int) {
			for i := start; i < end; i++ {
				m.Set(i, i*i)
			}
			for i := start; i < end; i++ {
				v, ok := m.Get(i)
				req.True(ok)
				req.NotNil(v)
				req.Equal(i*i, v.(int))
			}
			for i := start; i < end; i++ {
				m.Del(i)
			}
			wg.Done()
		}(i*10000, (i+1)*10000)
	}
	wg.Wait()
	m.(*hmap).info()
}

func BenchmarkLockfreeHashMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := NewHashMap()
		wg := sync.WaitGroup{}
		wg.Add(10)
		for i := 0; i < 10; i++ {
			go func(start, end int) {
				for i := start; i < end; i++ {
					m.Set(i, i*i)
				}
				for i := start; i < end; i++ {
					v, ok := m.Get(i)
					if !ok {
						b.Error("key not exist")
					}
					if v == nil || v.(int) != i*i {
						b.Error("key not match")
					}
				}
				for i := start; i < end; i++ {
					m.Del(i)
				}
				wg.Done()
			}(i*10000, (i+1)*10000)
		}
		wg.Wait()
	}
}

func BenchmarkMapAndRWMutex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := make(map[int]int)
		lock := sync.RWMutex{}
		wg := sync.WaitGroup{}
		wg.Add(10)
		for i := 0; i < 10; i++ {
			go func(start, end int) {
				for i := start; i < end; i++ {
					lock.Lock()
					m[i] = i * i
					lock.Unlock()
				}
				for i := start; i < end; i++ {
					lock.RLock()
					v, ok := m[i]
					lock.RUnlock()
					if !ok {
						b.Error("key not exist")
					}
					if v != i*i {
						b.Error("key not match")
					}
				}
				for i := start; i < end; i++ {
					lock.Lock()
					delete(m, i)
					lock.Unlock()
				}
				wg.Done()
			}(i*10000, (i+1)*10000)
		}
		wg.Wait()
	}
}

func BenchmarkSyncMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := sync.Map{}
		wg := sync.WaitGroup{}
		wg.Add(10)
		for i := 0; i < 10; i++ {
			go func(start, end int) {
				for i := start; i < end; i++ {
					m.Store(i, i*i)
				}
				for i := start; i < end; i++ {
					v, ok := m.Load(i)
					if !ok {
						b.Error("key not exist")
					}
					if v == nil || v.(int) != i*i {
						b.Error("key not match")
					}
				}
				for i := start; i < end; i++ {
					m.Delete(i)
				}
				wg.Done()
			}(i*10000, (i+1)*10000)
		}
		wg.Wait()
	}
}
