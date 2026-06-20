
```
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
