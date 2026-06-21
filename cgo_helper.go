//go:build cgobench

package zlib_wasm

/*
#cgo LDFLAGS: -lz
#include <zlib.h>
#include <stdlib.h>
*/
import "C"

import (
	"testing"
	"unsafe"
)

func runCGOCompressBench(b *testing.B, data []byte) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Формула размера буфера по умолчанию для zlib
		destLen := C.uLong(len(data) + len(data)/1000 + 12)
		dest := C.malloc(C.size_t(destLen))

		ret := C.compress((*C.Bytef)(dest), &destLen, (*C.Bytef)(unsafe.Pointer(&data[0])), C.uLong(len(data)))
		if ret != C.Z_OK {
			C.free(dest)
			b.Fatalf("CGO compress failed with code: %d", ret)
		}

		C.free(dest)
	}
}

func runCGODecompressBench(b *testing.B, data []byte) {
	// Подготавливаем сжатые данные заранее
	compLen := C.uLong(len(data) + len(data)/1000 + 12)
	compDest := C.malloc(C.size_t(compLen))
	C.compress((*C.Bytef)(compDest), &compLen, (*C.Bytef)(unsafe.Pointer(&data[0])), C.uLong(len(data)))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		destLen := C.uLong(len(data))
		dest := C.malloc(C.size_t(destLen))

		ret := C.uncompress((*C.Bytef)(dest), &destLen, (*C.Bytef)(compDest), compLen)
		if ret != C.Z_OK {
			C.free(dest)
			b.Fatalf("CGO uncompress failed with code: %d", ret)
		}

		C.free(dest)
	}
	C.free(compDest)
}