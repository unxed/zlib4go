
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
