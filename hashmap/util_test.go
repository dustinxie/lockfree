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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHash(t *testing.T) {
	req := require.New(t)

	tests := []struct {
		key  interface{}
		hash uint64
	}{
		// same value but diff type yields diff hash
		{uint8(16), 0x405e0156dd8344e7},
		{int8(16), 0xf3a679dd47011da5},
		{uint16(16), 0x2b5c34d98d93a618},
		{int16(16), 0x1ef38a49efb7a317},
		{uint32(16), 0x41fa2a8ac5fc042},
		{int32(16), 0xc0657c65c573a378},
		{uint64(16), 16},
		{int64(16), 0xf085e8eea6d7b547},
		{uint(16), 0x3cd2e914767dd151},
		{int(16), 0x4bbe4f327568c4f2},
		// []byte with same content as int yields diff hash
		{[]byte{16}, 0xc546af9a38a18681},
		{[]byte{16, 0}, 0x77553048d67374f0},
		{[]byte{16, 0, 0, 0}, 0x4aa15dc72df00989},
		{[]byte{16, 0, 0, 0, 0, 0, 0, 0}, 0x22a267105c467397},
		// string with same content as int yields diff hash
		{string([]byte{16}), 0xa75cd46d609bb55e},
		{string([]byte{16, 0}), 0x80106b95cd538e05},
		{string([]byte{16, 0, 0, 0}), 0x13750e44c2505c76},
		{string([]byte{16, 0, 0, 0, 0, 0, 0, 0}), 0xeed0936aebcf1fac},
		// string with same byte content yields diff hash
		{[]byte{0x10, 0x32, 0x54, 0x76}, 0x3c0db94b667c1e27},
		{string([]byte{0x10, 0x32, 0x54, 0x76}), 0x62b494eed1cdac58},
		// struct that implements Hash64 interface
		{testHash64{value: 16}, testHash64{value: 16}.Sum64()},
	}

	h := &hmap{}
	for _, test := range tests {
		req.Equal(test.hash, h.hash(test.key))
	}
}

type testHash64 struct {
	value int
}

func (th testHash64) Sum64() uint64 {
	return uint64(th.value*th.value%65535)<<33 + 1
}
