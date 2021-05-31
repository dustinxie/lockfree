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
	"container/list"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewStack(t *testing.T) {
	req := require.New(t)

	// test 4 threads
	s := NewStack()
	m := NewHashMap()
	wg := sync.WaitGroup{}
	wg.Add(4)
	for i := 0; i < 4; i++ {
		go func(start, end int) {
			for i := start; i < end; i++ {
				s.Push(i)
			}
			for i := start; i < end; i++ {
				m.Set(s.Pop(), nil)
			}
			wg.Done()
		}(i*10000, (i+1)*10000)
	}
	wg.Wait()
	req.Equal(0, s.Len())
	req.Nil(s.Pop())
	req.Equal(40000, m.Len())
	for i := 0; i < 40000; i++ {
		v, ok := m.Get(i)
		req.Nil(v)
		req.True(ok)
	}
}

func BenchmarkLockfreeStack(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := NewStack()
		wg := sync.WaitGroup{}
		wg.Add(10)
		for i := 0; i < 10; i++ {
			go func(start, end int) {
				for i := start; i < end; i++ {
					s.Push(i)
				}
				for i := start; i < end; i++ {
					s.Pop()
				}
				wg.Done()
			}(i*10000, (i+1)*10000)
		}
		wg.Wait()
	}
}

func BenchmarkStackAndRWMutex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		q := list.New()
		lock := sync.RWMutex{}
		wg := sync.WaitGroup{}
		wg.Add(10)
		for i := 0; i < 10; i++ {
			go func(start, end int) {
				for i := start; i < end; i++ {
					lock.Lock()
					q.PushFront(i)
					lock.Unlock()
				}
				for i := start; i < end; i++ {
					lock.Lock()
					q.Front()
					lock.Unlock()
				}
				wg.Done()
			}(i*10000, (i+1)*10000)
		}
		wg.Wait()
	}
}
