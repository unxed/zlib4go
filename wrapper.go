package zlib_wasm

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Compress compresses the given byte slice using the specified compression level (-1 to 9).
func Compress(data []byte, level int) ([]byte, error) {
	m := New()
	sourceLen := int32(len(data))
	bound := _compressBound(sourceLen)

	srcPtr := m.Xmalloc(sourceLen)
	destPtr := m.Xmalloc(bound)
	destLenPtr := m.Xmalloc(4)
	
	if srcPtr == 0 || destPtr == 0 || destLenPtr == 0 {
		return nil, fmt.Errorf("zlib: allocation failed")
	}
	defer m.Xfree(srcPtr)
	defer m.Xfree(destPtr)
	defer m.Xfree(destLenPtr)

	copy(m.memory[srcPtr:], data)
	binary.LittleEndian.PutUint32(m.memory[destLenPtr:], uint32(bound))

	// By default, the C implementation uses level 6.
	ret := m.Xcompress(destPtr, destLenPtr, srcPtr, sourceLen)
	if ret != 0 {
		return nil, fmt.Errorf("zlib: compression failed with code %d", ret)
	}

	actualSize := binary.LittleEndian.Uint32(m.memory[destLenPtr:])
	result := make([]byte, actualSize)
	copy(result, m.memory[destPtr:destPtr+int32(actualSize)])
	return result, nil
}

// Decompress decompresses a zlib-compressed buffer.
func Decompress(data []byte) ([]byte, error) {
	m := New()
	sourceLen := int32(len(data))
	destLen := sourceLen * 4
	if destLen < 1024 { destLen = 1024 }

	srcPtr := m.Xmalloc(sourceLen)
	destLenPtr := m.Xmalloc(4)
	if srcPtr == 0 || destLenPtr == 0 {
		return nil, fmt.Errorf("zlib: allocation failed")
	}
	defer m.Xfree(srcPtr)
	defer m.Xfree(destLenPtr)

	copy(m.memory[srcPtr:], data)

	for {
		destPtr := m.Xmalloc(destLen)
		if destPtr == 0 { return nil, fmt.Errorf("zlib: allocation failed") }
		binary.LittleEndian.PutUint32(m.memory[destLenPtr:], uint32(destLen))

		ret := m.Xuncompress(destPtr, destLenPtr, srcPtr, sourceLen)
		if ret == -5 { // Z_BUF_ERROR: destination buffer is too small
			m.Xfree(destPtr)
			destLen *= 2
			continue
		}

		if ret != 0 {
			m.Xfree(destPtr)
			return nil, fmt.Errorf("zlib: decompression failed with code %d", ret)
		}

		actualSize := binary.LittleEndian.Uint32(m.memory[destLenPtr:])
		result := make([]byte, actualSize)
		copy(result, m.memory[destPtr:destPtr+int32(actualSize)])
		m.Xfree(destPtr)
		return result, nil
	}
}

// Writer implements an io.WriteCloser that compresses data to an underlying io.Writer.
type Writer struct {
	m         *Module
	w         io.Writer
	strmPtr   int32
	inBufPtr  int32
	outBufPtr int32
	bufSize   int32
	closed    bool
}

func NewWriterLevel(w io.Writer, level int) (*Writer, error) {
	m := New()
	strmPtr := m.Xmalloc(56)
	if strmPtr == 0 { return nil, fmt.Errorf("zlib: malloc fail") }
	for i := int32(0); i < 56; i++ { m.memory[strmPtr+i] = 0 }

	ver := "1.3.2.1\x00"
	verPtr := m.Xmalloc(int32(len(ver)))
	copy(m.memory[verPtr:], ver)
	defer m.Xfree(verPtr)

	ret := m.XdeflateInit2_(strmPtr, int32(level), 8, 15, 8, 0, verPtr, 56)
	if ret != 0 { return nil, fmt.Errorf("deflateInit failed: %d", ret) }

	bs := int32(16384)
	return &Writer{m: m, w: w, strmPtr: strmPtr, inBufPtr: m.Xmalloc(bs), outBufPtr: m.Xmalloc(bs), bufSize: bs}, nil
}

func (w *Writer) Write(p []byte) (int, error) {
	if w.closed { return 0, io.ErrClosedPipe }
	off := 0
	for off < len(p) {
		chunk := int32(len(p) - off)
		if chunk > w.bufSize { chunk = w.bufSize }
		copy(w.m.memory[w.inBufPtr:], p[off:off+int(chunk)])
		binary.LittleEndian.PutUint32(w.m.memory[w.strmPtr+0:], uint32(w.inBufPtr))
		binary.LittleEndian.PutUint32(w.m.memory[w.strmPtr+4:], uint32(chunk))
		for {
			binary.LittleEndian.PutUint32(w.m.memory[w.strmPtr+12:], uint32(w.outBufPtr))
			binary.LittleEndian.PutUint32(w.m.memory[w.strmPtr+16:], uint32(w.bufSize))
			if w.m.Xdeflate(w.strmPtr, 0) != 0 { return off, fmt.Errorf("deflate error") }
			written := w.bufSize - int32(binary.LittleEndian.Uint32(w.m.memory[w.strmPtr+16:]))
			if written > 0 { w.w.Write(w.m.memory[w.outBufPtr : w.outBufPtr+written]) }
			if binary.LittleEndian.Uint32(w.m.memory[w.strmPtr+16:]) > 0 { break }
		}
		off += int(chunk)
	}
	return off, nil
}

func (w *Writer) Close() error {
	if w.closed { return nil }
	w.closed = true
	for {
		binary.LittleEndian.PutUint32(w.m.memory[w.strmPtr+12:], uint32(w.outBufPtr))
		binary.LittleEndian.PutUint32(w.m.memory[w.strmPtr+16:], uint32(w.bufSize))
		ret := w.m.Xdeflate(w.strmPtr, 4)
		written := w.bufSize - int32(binary.LittleEndian.Uint32(w.m.memory[w.strmPtr+16:]))
		if written > 0 { w.w.Write(w.m.memory[w.outBufPtr : w.outBufPtr+written]) }
		if ret == 1 { break }
	}
	w.m.XdeflateEnd(w.strmPtr); w.m.Xfree(w.strmPtr); w.m.Xfree(w.inBufPtr); w.m.Xfree(w.outBufPtr)
	return nil
}

// Reader implements an io.ReadCloser that decompresses data from an underlying io.Reader.
type Reader struct {
	m *Module; r io.Reader; strmPtr, inBufPtr, outBufPtr, bufSize int32
	inBuf []byte; eof, closed bool
}

func NewReader(r io.Reader) (*Reader, error) {
	m := New(); strmPtr := m.Xmalloc(56)
	for i := int32(0); i < 56; i++ { m.memory[strmPtr+i] = 0 }
	ver := "1.3.2.1\x00"; vPtr := m.Xmalloc(int32(len(ver)))
	copy(m.memory[vPtr:], ver); defer m.Xfree(vPtr)
	if m.XinflateInit2_(strmPtr, 15, vPtr, 56) != 0 { return nil, fmt.Errorf("inflateInit failed") }
	bs := int32(16384)
	return &Reader{m: m, r: r, strmPtr: strmPtr, inBufPtr: m.Xmalloc(bs), outBufPtr: m.Xmalloc(bs), bufSize: bs, inBuf: make([]byte, bs)}, nil
}

func (rd *Reader) Read(p []byte) (int, error) {
	if rd.closed { return 0, io.ErrClosedPipe }
	total := 0
	for total < len(p) {
		availIn := binary.LittleEndian.Uint32(rd.m.memory[rd.strmPtr+4:])
		if availIn == 0 && !rd.eof {
			n, err := rd.r.Read(rd.inBuf)
			if n > 0 {
				copy(rd.m.memory[rd.inBufPtr:], rd.inBuf[:n])
				binary.LittleEndian.PutUint32(rd.m.memory[rd.strmPtr+0:], uint32(rd.inBufPtr))
				binary.LittleEndian.PutUint32(rd.m.memory[rd.strmPtr+4:], uint32(n))
			}
			if err != nil { 
				if err == io.EOF { rd.eof = true } else { return total, err }
			}
		}
		if binary.LittleEndian.Uint32(rd.m.memory[rd.strmPtr+4:]) == 0 && rd.eof { return total, io.EOF }
		
		destL := int32(len(p) - total)
		if destL > rd.bufSize { destL = rd.bufSize }
		binary.LittleEndian.PutUint32(rd.m.memory[rd.strmPtr+12:], uint32(rd.outBufPtr))
		binary.LittleEndian.PutUint32(rd.m.memory[rd.strmPtr+16:], uint32(destL))
		
		ret := rd.m.Xinflate(rd.strmPtr, 0)
		written := destL - int32(binary.LittleEndian.Uint32(rd.m.memory[rd.strmPtr+16:]))
		if written > 0 {
			copy(p[total:], rd.m.memory[rd.outBufPtr : rd.outBufPtr+written])
			total += int(written)
		}
		if ret == 1 { rd.eof = true; return total, nil }
		if ret != 0 { return total, fmt.Errorf("inflate error %d", ret) }
		if written == 0 { break }
	}
	return total, nil
}

func (rd *Reader) Close() error {
	if !rd.closed {
		rd.closed = true
		rd.m.XinflateEnd(rd.strmPtr)
		rd.m.Xfree(rd.strmPtr); rd.m.Xfree(rd.inBufPtr); rd.m.Xfree(rd.outBufPtr)
	}
	return nil
}