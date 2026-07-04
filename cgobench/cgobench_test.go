//go:build cgobench

package cgobench

import (
	"bytes"
	"compress/zlib"
	"io"
	"testing"

	zlib_wasm "github.com/unxed/zlib4go"
)

func getBenchData() []byte {
	return bytes.Repeat([]byte("This is a test string for zlib compression benchmarking. It contains some redundant data to allow compression to work effectively. "), 20000)
}

func Benchmark1_CompressCGO(b *testing.B) {
	data := getBenchData()
	RunCGOCompressBench(b, data)
}

func Benchmark1_CompressStdlib(b *testing.B) {
	data := getBenchData()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		w, _ := zlib.NewWriterLevel(&buf, 6)
		w.Write(data)
		w.Close()
	}
}

func Benchmark1_CompressWasm(b *testing.B) {
	data := getBenchData()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := zlib_wasm.Compress(data, 6)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark2_DecompressCGO(b *testing.B) {
	data := getBenchData()
	RunCGODecompressBench(b, data)
}

func Benchmark2_DecompressStdlib(b *testing.B) {
	data := getBenchData()
	b.SetBytes(int64(len(data)))
	var buf bytes.Buffer
	w, _ := zlib.NewWriterLevel(&buf, 6)
	w.Write(data)
	w.Close()
	compressed := buf.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r, err := zlib.NewReader(bytes.NewReader(compressed))
		if err != nil {
			b.Fatal(err)
		}
		_, err = io.ReadAll(r)
		r.Close()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark2_DecompressWasm(b *testing.B) {
	data := getBenchData()
	compressed, _ := zlib_wasm.Compress(data, 6)
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := zlib_wasm.Decompress(compressed)
		if err != nil {
			b.Fatal(err)
		}
	}
}