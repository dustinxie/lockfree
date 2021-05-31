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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStack(t *testing.T) {
	req := require.New(t)

	s := NewStack()
	req.Equal(0, s.Len())
	req.Nil(s.Pop())
	req.Nil(s.Peek())

	tests := []interface{}{"a", 1, "b", 2, "c", 3, "d", 4}
	size := len(tests)
	for i, item := range tests {
		s.Push(item)
		req.Equal(i+1, s.Len())
	}
	req.Equal(size, s.Len())

	for i := range tests {
		req.Equal(tests[size-1-i], s.Pop())
		req.Equal(size-1-i, s.Len())
	}
	req.Equal(0, s.Len())
	req.Nil(s.Peek())
	req.Nil(s.Pop())
}
