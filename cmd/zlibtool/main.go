package main

import (
	"fmt"
	"flag"
	"io"
	"log"
	"os"

	"github.com/unxed/zlib4go" 
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: zlibtool [-d] [-l level]\n\n")
		fmt.Fprintf(os.Stderr, "zlibtool compresses or decompresses stdin to stdout using Wasm-based zlib.\n\n")
		flag.PrintDefaults()
	}

	decompress := flag.Bool("d", false, "Decompress input (default is compression)")
	level := flag.Int("l", -1, "Compression level (-1 to 9)")
	flag.Parse()

	if *decompress {
		// Decompression mode
		reader, err := zlib_wasm.NewReader(os.Stdin)
		if err != nil {
			log.Fatalf("Failed to initialize Reader: %v", err)
		}
		defer reader.Close()

		_, err = io.Copy(os.Stdout, reader)
		if err != nil {
			log.Fatalf("Decompression failed: %v", err)
		}
	} else {
		// Compression mode
		writer, err := zlib_wasm.NewWriterLevel(os.Stdout, *level)
		if err != nil {
			log.Fatalf("Failed to initialize Writer: %v", err)
		}

		_, err = io.Copy(writer, os.Stdin)
		if err != nil {
			log.Fatalf("Compression failed: %v", err)
		}

		err = writer.Close()
		if err != nil {
			log.Fatalf("Failed to close stream: %v", err)
		}
	}
}
