
# zlib4go

A Pure Go wrapper for `zlib` using WebAssembly. This package provides bit-for-bit compatibility with the original C zlib implementation without requiring CGO.

## Features
- **No CGO:** Uses WebAssembly (transpiled to Go via `wasm2go`) for portability.
- **Compatibility:** Fully compatible with standard zlib (RFC 1950).
- **Streaming:** Provides `io.Reader` and `io.Writer` implementations.

## Installation

```bash
go get github.com/unxed/zlib4go
```

## Usage

### Simple Compression/Decompression
```go
import "github.com/unxed/zlib4go"

data := []byte("hello world")

// Compress
compressed, _ := zlib_wasm.Compress(data, 6)

// Decompress
decompressed, _ := zlib_wasm.Decompress(compressed)
```

### Streaming
```go
// Writing (Compressing)
w, _ := zlib_wasm.NewWriterLevel(outputStream, 9)
w.Write(data)
w.Close()

// Reading (Decompressing)
r, _ := zlib_wasm.NewReader(inputStream)
io.Copy(destination, r)
r.Close()
```
## Benchmarks

To run performance benchmarks comparing Wasm `zlib4go`, Go standard library `compress/zlib`, and vanilla C `zlib` (via CGO):

```bash
# Run pure Go benchmarks (zlib4go Wasm vs Standard Library)
go test -bench=. -benchmem

# Run all benchmarks including vanilla C zlib (requires CGO and libz)
go test -tags cgobench -bench=. -benchmem
```

## Compilation (Internal)
The underlying WebAssembly core was compiled from zlib source using:

```bash
$WASI_SDK_PATH/bin/clang -O3 -nostartfiles \
  -Wl,--no-entry \
  -Wl,--export=compress \
  -Wl,--export=uncompress \
  -Wl,--export=compressBound \
  -Wl,--export=deflateInit2_ \
  -Wl,--export=deflate \
  -Wl,--export=deflateEnd \
  -Wl,--export=inflateInit2_ \
  -Wl,--export=inflate \
  -Wl,--export=inflateEnd \
  -Wl,--export=malloc \
  -Wl,--export=free \
  -o zlib.wasm \
  adler32.c crc32.c deflate.c infback.c inffast.c inflate.c inftrees.c trees.c zutil.c compress.c uncompr.c
```
## Performance Benchmarks

Below is a comparison of `zlib4go` (Wasm), Go standard library (`compress/zlib`), and native C `zlib` (via CGO) on a ~2.5 MB redundant text payload (Intel i5-6300U @ 2.40GHz, Linux amd64):

| Operation / Engine | `zlib4go` (Wasm) | Go Stdlib (`compress/zlib`) | C `zlib` (CGO) |
| --- | --- | --- | --- |
| **Compress (level 6)** | **20.9 ms** (100% / Baseline) | 10.7 ms (**95% faster** / ~1.9x) | 13.8 ms (**51% faster** / ~1.5x) |
| **Decompress** | **10.2 ms** (100% / Baseline) | 11.4 ms (**11% slower** / ~0.9x) | 1.4 ms (**7.2x faster** / ~7.2x) |
| **Memory Allocations** | **~288 KB / 1 alloc** | ~846 KB / 28 allocs (**193% more memory**) | ~8 B / 1 alloc |

### Takeaways
- **Memory Efficiency:** Thanks to aggressive `sync.Pool` caching, `zlib4go` produces only **1 heap allocation** per operation (the result slice). It consumes nearly **3x less RAM** during compression than Go's standard library, significantly reducing garbage collector pressure.
- **Decompression:** Performance is **11% faster** than Go's standard library while remaining entirely portable, thread-safe, and dependency-free.
- **CGO vs Wasm:** Pure Go Wasm-based compression is now only **1.5x slower than raw CGO**, making it a highly competitive choice for serverless, scratch containers, and CGO-disabled environments.
