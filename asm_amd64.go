//go:build amd64 && !noasm

package zlib_wasm

import "unsafe"

//go:noescape
func updateHashChain(memBase unsafe.Pointer, pos, endPos, windowOffset, headOffset, prevOffset int32, masks uint64, shift, hash int32) int32
