package aac

import "io"

type bitReader struct {
	data []byte
	bit  int
}

func newBitReader(data []byte) bitReader {
	return bitReader{data: data}
}

func (r *bitReader) bitsLeft() int {
	return len(r.data)*8 - r.bit
}

func (r *bitReader) readBits(n uint) (uint32, error) {
	if n == 0 {
		return 0, nil
	}
	if n > 32 || int(n) > r.bitsLeft() {
		return 0, io.ErrUnexpectedEOF
	}
	var v uint32
	for i := uint(0); i < n; i++ {
		byteIndex := r.bit >> 3
		shift := 7 - (r.bit & 7)
		v = (v << 1) | uint32((r.data[byteIndex]>>shift)&1)
		r.bit++
	}
	return v, nil
}

type bitWriter struct {
	buf []byte
	bit int
}

func (w *bitWriter) writeBits(n uint, v uint32) {
	for i := int(n) - 1; i >= 0; i-- {
		if w.bit&7 == 0 {
			w.buf = append(w.buf, 0)
		}
		if (v>>uint(i))&1 != 0 {
			w.buf[len(w.buf)-1] |= 1 << uint(7-(w.bit&7))
		}
		w.bit++
	}
}

func (w *bitWriter) bytes() []byte {
	out := make([]byte, len(w.buf))
	copy(out, w.buf)
	return out
}
