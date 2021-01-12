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
	"fmt"
	"reflect"
	"unsafe"

	"github.com/dchest/siphash"
)

const (
	// intSize is the size in bytes of an int or uint value
	intSize = (32 << (^uint(0) >> 63)) >> 3
)

// 64-bit hash provides 2^32 collision-resistance, which suffices for most use-case
func (h *hmap) hash(key interface{}) uint64 {
	switch v := key.(type) {
	case uint8:
		return memhash(h.k0, h.k1, unsafe.Pointer(&v), 1)
	case int8:
		return memhash(h.k0, h.k1-1, unsafe.Pointer(&v), 1)
	case uint16:
		return memhash(h.k0, h.k1, unsafe.Pointer(&v), 2)
	case int16:
		return memhash(h.k0, h.k1-1, unsafe.Pointer(&v), 2)
	case uint32:
		return memhash(h.k0, h.k1, unsafe.Pointer(&v), 4)
	case int32:
		return memhash(h.k0, h.k1-1, unsafe.Pointer(&v), 4)
	case uint64:
		return v
	case int64:
		return memhash(h.k0, h.k1, unsafe.Pointer(&v), 8)
	case uint:
		return memhash(h.k0, h.k1+1, unsafe.Pointer(&v), intSize)
	case int:
		return memhash(h.k0, h.k1+2, unsafe.Pointer(&v), intSize)
	case []byte:
		return siphash.Hash(h.k0, h.k1, v)
	case string:
		hdr := (*reflect.StringHeader)(unsafe.Pointer(&v))
		sh := reflect.SliceHeader{
			Data: hdr.Data,
			Len:  hdr.Len,
			Cap:  hdr.Len,
		}
		return siphash.Hash(h.k0, h.k1-1, *(*[]byte)(unsafe.Pointer(&sh)))
	default:
		if h, ok := v.(Hash64); ok {
			return h.Sum64()
		}
		panic(fmt.Errorf("unsupported key type %T", v))
	}
}

// memhash computes the hash of 'size' bytes of memory at addr
func memhash(k0, k1 uint64, addr unsafe.Pointer, size int) uint64 {
	sh := reflect.SliceHeader{
		Data: uintptr(addr),
		Len:  size,
		Cap:  size,
	}
	return siphash.Hash(k0, k1, *(*[]byte)(unsafe.Pointer(&sh)))
}
