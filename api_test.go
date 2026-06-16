package aac

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestNewOptionsSelectsTransport(t *testing.T) {
	adts, err := New(Options{Transport: TransportADTS})
	if err != nil {
		t.Fatal(err)
	}
	defer adts.Close()
	if got := adts.Transport(); got != TransportADTS {
		t.Fatalf("transport = %s, want %s", got, TransportADTS)
	}

	raw, err := New(Options{
		Transport: TransportRaw,
		Config: Config{
			ObjectType:    AOTAACLC,
			SampleRate:    44100,
			ChannelConfig: 2,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	if got := raw.Transport(); got != TransportRaw {
		t.Fatalf("transport = %s, want %s", got, TransportRaw)
	}
}

func TestADTSReaderReadFrame(t *testing.T) {
	cfg := Config{ObjectType: AOTAACLC, SampleRate: 44100, ChannelConfig: 2}
	frame, err := appendTestADTSHeader(nil, cfg, 4)
	if err != nil {
		t.Fatal(err)
	}
	frame = append(frame, 1, 2, 3, 4)

	reader := NewADTSReader(bytes.NewReader(append(frame, frame...)))
	buf := make([]byte, 0, len(frame))
	got, err := reader.ReadFrame(buf)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.Data, frame) {
		t.Fatalf("frame bytes = %x, want %x", got.Data, frame)
	}
	if got.Header.FrameLength != len(frame) {
		t.Fatalf("frame length = %d, want %d", got.Header.FrameLength, len(frame))
	}
	if reader.FrameIndex() != 1 {
		t.Fatalf("frame index = %d, want 1", reader.FrameIndex())
	}
	if reader.Offset() != int64(len(frame)) {
		t.Fatalf("offset = %d, want %d", reader.Offset(), len(frame))
	}
	got, err = reader.ReadFrame(got.Data[:0])
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.Data, frame) {
		t.Fatalf("second frame bytes = %x, want %x", got.Data, frame)
	}
	_, err = reader.ReadFrame(got.Data[:0])
	if !errors.Is(err, io.EOF) {
		t.Fatalf("err = %v, want EOF", err)
	}
}

func TestDecodeADTSFrameIntoRejectsRawDecoder(t *testing.T) {
	raw, err := New(Options{
		Transport: TransportRaw,
		Config: Config{
			ObjectType:    AOTAACLC,
			SampleRate:    44100,
			ChannelConfig: 2,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()

	frame, err := appendTestADTSHeader(nil, Config{ObjectType: AOTAACLC, SampleRate: 44100, ChannelConfig: 2}, 0)
	if err != nil {
		t.Fatal(err)
	}
	dst := []int16{1, 2, 3}
	got, _, err := raw.DecodeADTSFrameInto(dst, frame)
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("err = %v, want ErrInvalidConfig", err)
	}
	if len(got) != len(dst) {
		t.Fatalf("dst length = %d, want %d", len(got), len(dst))
	}
}

func TestADTSReaderNil(t *testing.T) {
	_, err := NewADTSReader(nil).ReadFrame(nil)
	if !errors.Is(err, ErrInvalidADTS) {
		t.Fatalf("err = %v, want ErrInvalidADTS", err)
	}
	if NewADTSReader(nil).FrameIndex() != 0 {
		t.Fatal("nil reader frame index changed")
	}
}
