//go:build cgobench

package zlib_wasm

import (
	"testing"
)

// === Benchmark: Vanilla C (zlib via CGO) ===

func BenchmarkCompressCGO(b *testing.B) {
	data := getBenchData()
	runCGOCompressBench(b, data)
}

func BenchmarkDecompressCGO(b *testing.B) {
	data := getBenchData()
	runCGODecompressBench(b, data)
}