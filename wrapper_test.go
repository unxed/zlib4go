package zlib_wasm

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"testing"
)

// Проверка: сжатие в памяти через Wasm -> распаковка через Go stdlib
func TestCompressDecompressGoStd(t *testing.T) {
	input := []byte("hello world! this is a test of the wasm2go transpiled zlib wrapper. bit-for-bit compatibility is highly valued.")

	compressed, err := Compress(input, 6)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}

	r, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("Failed to create std zlib reader: %v", err)
	}
	defer r.Close()

	decompressed, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}

	if !bytes.Equal(input, decompressed) {
		t.Errorf("Decompressed data mismatch. Expected %q, got %q", input, decompressed)
	}
}

// Проверка: распаковка в памяти через Wasm -> сжатых через Go stdlib данных
func TestDecompressGoStdCompressed(t *testing.T) {
	input := []byte("Some data that will be compressed using standard library and decompressed via Wasm.")
	
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	w.Write(input)
	w.Close()

	decompressed, err := Decompress(buf.Bytes())
	if err != nil {
		t.Fatalf("Decompress failed: %v", err)
	}

	if !bytes.Equal(input, decompressed) {
		t.Errorf("Mismatch. Got: %s", string(decompressed))
	}
}

// Проверка: сквозной тест потоков (Wasm Writer -> Wasm Reader)
func TestStreamingReaderWriter(t *testing.T) {
	input := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.")

	// Сжимаем поток
	var compressedBuf bytes.Buffer
	w, err := NewWriterLevel(&compressedBuf, 9, false)
	if err != nil {
		t.Fatalf("NewWriterLevel failed: %v", err)
	}

	// Важно: записываем данные в писатель!
	chunkSize := 10
	for i := 0; i < len(input); i += chunkSize {
		end := i + chunkSize
		if end > len(input) { end = len(input) }
		w.Write(input[i:end])
	}
	w.Close()

	// Распаковываем поток
	reader, err := NewReader(&compressedBuf)
	if err != nil {
		t.Fatalf("NewReader failed: %v", err)
	}

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if !bytes.Equal(input, decompressed) {
		t.Errorf("Streaming mismatch.\nExpected: %s\nGot:      %s", string(input), string(decompressed))
	}
}

// Побитовая проверка: сжатие разных уровней один-в-один совпадает с Python
func TestBitForBitWithPython(t *testing.T) {
	_, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 not found, skipping bit-for-bit validation")
	}

	input := "This is some test data that we will compress and compare with Python's zlib to verify bit-for-bit equivalence!"

	for _, lv := range []int{1, 6, 9} {
		var buf bytes.Buffer
		w, _ := NewWriterLevel(&buf, lv, false)
		w.Write([]byte(input))
		w.Close()
		compWasm := buf.Bytes()

		script := fmt.Sprintf("import zlib; print(zlib.compress(b%q, %d).hex())", input, lv)
		cmd := exec.Command("python3", "-c", script)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Run()

		pythonHex := strings.TrimSpace(out.String())
		wasmHex := hex.EncodeToString(compWasm)

		if pythonHex != wasmHex {
			t.Errorf("Mismatch at level %d!\nPython: %s\nWasm:   %s", lv, pythonHex, wasmHex)
		}
	}
}
