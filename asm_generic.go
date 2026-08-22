//go:build !amd64 || noasm

package zlib_wasm

import "unsafe"

func updateHashChain(memBase unsafe.Pointer, pos, endPos, windowOffset, headOffset, prevOffset int32, masks uint64, shift, hash int32) int32 {
	// Use slices instead of giant array types: the latter are rejected by the
	// compiler on 32-bit targets even though the backing window is addressed
	// only at runtime.
	mem := unsafe.Slice((*byte)(memBase), 0x7fffffff)
	mem16 := unsafe.Slice((*uint16)(memBase), 0x3fffffff)
	wMask := int32(masks)
	hMask := int32(masks >> 32)
	for pos < endPos {
		b := int32(mem[windowOffset+pos])
		hash = ((hash << shift) ^ b) & hMask
		oldHead := mem16[(headOffset>>1)+hash]
		mem16[(prevOffset>>1)+(pos&wMask)] = oldHead
		mem16[(headOffset>>1)+hash] = uint16(pos)
		pos++
	}
	return hash
}
