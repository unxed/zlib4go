package zlib_wasm

import (
	"bytes"
	"compress/zlib"
	"io"
	"testing"
)

// getBenchData генерирует тестовые данные (около ~2.5 MB),
// содержащие повторения, чтобы сжатие было эффективным.
func getBenchData() []byte {
	return bytes.Repeat([]byte("This is a test string for zlib compression benchmarking. It contains some redundant data to allow compression to work effectively. "), 20000)
}

// === Benchmark: Wasm (zlib4go) ===

func BenchmarkCompressWasm(b *testing.B) {
	data := getBenchData()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Compress(data, 6)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecompressWasm(b *testing.B) {
	data := getBenchData()
	compressed, _ := Compress(data, 6)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Decompress(compressed)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// === Benchmark: Go Stdlib (compress/zlib) ===

func BenchmarkCompressStdlib(b *testing.B) {
	data := getBenchData()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		w, _ := zlib.NewWriterLevel(&buf, 6)
		w.Write(data)
		w.Close()
	}
}

func BenchmarkDecompressStdlib(b *testing.B) {
	data := getBenchData()
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