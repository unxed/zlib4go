//go:build !amd64 || noasm

package zlib_wasm

import "unsafe"

func updateHashChain(memBase unsafe.Pointer, pos, endPos, windowOffset, headOffset, prevOffset int32, masks uint64, shift, hash int32) int32 {
	mem := (*[0x7fffffff]byte)(memBase)
	mem16 := (*[0x3fffffff]uint16)(memBase)
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
