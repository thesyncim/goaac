package aac

import (
	"errors"
	"fmt"
	"io"
)

type ADTSReader struct {
	r          io.Reader
	frameIndex int
	offset     int64
}

// NewADTSReader creates a frame reader for an ADTS byte stream.
func NewADTSReader(r io.Reader) *ADTSReader {
	return &ADTSReader{r: r}
}

// FrameIndex returns the zero-based index of the next frame to be read.
func (r *ADTSReader) FrameIndex() int {
	if r == nil {
		return 0
	}
	return r.frameIndex
}

// Offset returns the byte offset of the next frame in the stream.
func (r *ADTSReader) Offset() int64 {
	if r == nil {
		return 0
	}
	return r.offset
}

// ReadFrame reads one complete ADTS frame and appends its bytes to dst. The
// returned ADTSFrame.Data slice aliases the returned buffer region.
func (r *ADTSReader) ReadFrame(dst []byte) (ADTSFrame, error) {
	if r == nil || r.r == nil {
		return ADTSFrame{}, fmt.Errorf("%w: nil ADTS reader", ErrInvalidADTS)
	}
	startLen := len(dst)
	var header [ADTSHeaderSize]byte
	n, err := io.ReadFull(r.r, header[:])
	if err != nil {
		if errors.Is(err, io.EOF) && n == 0 {
			return ADTSFrame{}, io.EOF
		}
		return ADTSFrame{}, adtsFrameError(r.frameIndex, r.offset, ErrNeedMoreData)
	}
	frameIndex := r.frameIndex
	frameOffset := r.offset
	h, err := parseADTSHeaderPrefix(header[:])
	if err != nil {
		return ADTSFrame{}, adtsFrameError(frameIndex, frameOffset, err)
	}
	dst = append(dst, header[:]...)
	remaining := h.FrameLength - ADTSHeaderSize
	if remaining < 0 {
		return ADTSFrame{}, adtsFrameError(frameIndex, frameOffset, ErrInvalidADTS)
	}
	payloadStart := len(dst)
	dst = growBytes(dst, remaining)
	if _, err := io.ReadFull(r.r, dst[payloadStart:]); err != nil {
		return ADTSFrame{}, adtsFrameError(frameIndex, frameOffset, ErrNeedMoreData)
	}
	h, err = ParseADTSHeader(dst[startLen:])
	if err != nil {
		return ADTSFrame{}, adtsFrameError(frameIndex, frameOffset, err)
	}
	frame := ADTSFrame{Header: h, Data: dst[startLen:], Index: frameIndex, Offset: frameOffset}
	r.frameIndex++
	r.offset += int64(h.FrameLength)
	return frame, nil
}

// DecodeADTSReader decodes all ADTS frames from r and returns newly allocated
// interleaved signed 16-bit PCM samples.
func DecodeADTSReader(r io.Reader) ([]int16, Config, error) {
	return DecodeADTSReaderInto(nil, r)
}

// DecodeADTSReaderInto decodes all ADTS frames from r and appends PCM to dst.
func DecodeADTSReaderInto(dst []int16, r io.Reader) ([]int16, Config, error) {
	ar := NewADTSReader(r)
	dec, err := NewADTSDecoder()
	if err != nil {
		return dst, Config{}, err
	}
	defer dec.Close()

	var frameBuf []byte
	for {
		frameBuf = frameBuf[:0]
		frame, err := ar.ReadFrame(frameBuf)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return dst, Config{}, err
		}
		frameBuf = frame.Data
		dst, _, err = dec.DecodeADTSFrameInto(dst, frame.Data)
		if err != nil {
			return dst, Config{}, adtsFrameError(frame.Index, frame.Offset, err)
		}
	}
	return dst, dec.Config(), nil
}

func growBytes(dst []byte, n int) []byte {
	if n <= 0 {
		return dst
	}
	need := len(dst) + n
	if need <= cap(dst) {
		return dst[:need]
	}
	next := make([]byte, need)
	copy(next, dst)
	return next
}
