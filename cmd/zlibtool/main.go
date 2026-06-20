package main

import (
	"flag"
	"io"
	"log"
	"os"

	// Убедитесь, что этот импорт соответствует названию модуля из go.mod
	// Если go.mod: "module zlib4go", то импорт такой:
	"github.com/unxed/zlib4go" 
)

func main() {
	// Флаги запуска
	decompress := flag.Bool("d", false, "Распаковать поток (по умолчанию сжимает)")
	level := flag.Int("l", -1, "Уровень сжатия (от -1 до 9)")
	flag.Parse()

	if *decompress {
		// Режим распаковки
		reader, err := zlib_wasm.NewReader(os.Stdin)
		if err != nil {
			log.Fatalf("Ошибка инициализации Reader: %v", err)
		}
		defer reader.Close()

		_, err = io.Copy(os.Stdout, reader)
		if err != nil {
			log.Fatalf("Ошибка распаковки: %v", err)
		}
	} else {
		// Режим сжатия
		writer, err := zlib_wasm.NewWriterLevel(os.Stdout, *level, false)
		if err != nil {
			log.Fatalf("Ошибка инициализации Writer: %v", err)
		}
		
		_, err = io.Copy(writer, os.Stdin)
		if err != nil {
			log.Fatalf("Ошибка сжатия: %v", err)
		}
		
		err = writer.Close()
		if err != nil {
			log.Fatalf("Ошибка закрытия потока: %v", err)
		}
	}
}
