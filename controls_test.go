package aac

import (
	"bytes"
	"errors"
	"testing"
)

func TestNewOptionsControlPlane(t *testing.T) {
	adts, err := New(Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer adts.Close()
	if got := adts.Transport(); got != TransportADTS {
		t.Fatalf("auto transport without config = %s, want %s", got, TransportADTS)
	}

	raw, err := New(Options{
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
		t.Fatalf("auto transport with raw config = %s, want %s", got, TransportRaw)
	}

	if _, err := New(Options{Transport: Transport(99)}); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("unknown transport err = %v, want ErrInvalidConfig", err)
	}
	if _, err := New(Options{Transport: TransportRaw}); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("raw without config err = %v, want ErrInvalidConfig", err)
	}
}

func TestDecoderCloseTransitions(t *testing.T) {
	adts, err := New(Options{Transport: TransportADTS})
	if err != nil {
		t.Fatal(err)
	}
	if err := adts.Close(); err != nil {
		t.Fatal(err)
	}
	if err := adts.Close(); err != nil {
		t.Fatalf("second close = %v, want nil", err)
	}
	dst := []int16{7, 8}
	got, _, err := adts.Decode(dst, nil)
	if !errors.Is(err, ErrClosed) {
		t.Fatalf("Decode after close err = %v, want ErrClosed", err)
	}
	if !equalInt16s(got, dst) {
		t.Fatalf("Decode after close changed dst: got %v, want %v", got, dst)
	}

	raw := newRawTestDecoder(t)
	if err := raw.Close(); err != nil {
		t.Fatal(err)
	}
	got, _, err = raw.DecodeRawInto(dst, []byte{0})
	if !errors.Is(err, ErrClosed) {
		t.Fatalf("DecodeRawInto after close err = %v, want ErrClosed", err)
	}
	if !equalInt16s(got, dst) {
		t.Fatalf("DecodeRawInto after close changed dst: got %v, want %v", got, dst)
	}
}

func TestDecoderTransportMisusePreservesDst(t *testing.T) {
	adts, err := New(Options{Transport: TransportADTS})
	if err != nil {
		t.Fatal(err)
	}
	defer adts.Close()

	dst := []int16{1, 2, 3}
	got, _, err := adts.DecodeRawInto(dst, []byte{0})
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("ADTS decoder raw err = %v, want ErrInvalidConfig", err)
	}
	if !equalInt16s(got, dst) {
		t.Fatalf("ADTS decoder raw misuse changed dst: got %v, want %v", got, dst)
	}

	raw := newRawTestDecoder(t)
	defer raw.Close()
	frame, err := AppendADTSHeader(nil, Config{ObjectType: AOTAACLC, SampleRate: 44100, ChannelConfig: 2}, 0)
	if err != nil {
		t.Fatal(err)
	}
	got, _, err = raw.DecodeADTSFrameInto(dst, frame)
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("raw decoder ADTS err = %v, want ErrInvalidConfig", err)
	}
	if !equalInt16s(got, dst) {
		t.Fatalf("raw decoder ADTS misuse changed dst: got %v, want %v", got, dst)
	}
}

func TestEncoderTransportMisusePreservesDst(t *testing.T) {
	var pcm [2 * encoderSamplesPerFrame]int16
	fillEncoderSmoothPCM(pcm[:], 2)
	dst := []byte{1, 2, 3}

	adts := newTestEncoder(t, TransportADTS)
	defer adts.Close()
	got, _, err := adts.EncodeRawInto(dst, pcm[:])
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("ADTS encoder raw err = %v, want ErrInvalidConfig", err)
	}
	if !bytes.Equal(got, dst) {
		t.Fatalf("ADTS encoder raw misuse changed dst: got %v, want %v", got, dst)
	}
	got, _, err = adts.EncodeRTMPMessageInto(dst, pcm[:])
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("ADTS encoder RTMP err = %v, want ErrInvalidConfig", err)
	}
	if !bytes.Equal(got, dst) {
		t.Fatalf("ADTS encoder RTMP misuse changed dst: got %v, want %v", got, dst)
	}
	got, _, _, _, err = adts.EncodeRTMPSamplesInto(dst, pcm[:])
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("ADTS encoder RTMP samples err = %v, want ErrInvalidConfig", err)
	}
	if !bytes.Equal(got, dst) {
		t.Fatalf("ADTS encoder RTMP samples misuse changed dst: got %v, want %v", got, dst)
	}

	raw := newTestEncoder(t, TransportRaw)
	defer raw.Close()
	got, _, err = raw.EncodeADTSFrameInto(dst, pcm[:])
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("raw encoder ADTS err = %v, want ErrInvalidConfig", err)
	}
	if !bytes.Equal(got, dst) {
		t.Fatalf("raw encoder ADTS misuse changed dst: got %v, want %v", got, dst)
	}
}

func TestEncoderSamplesControls(t *testing.T) {
	var pcm [2 * encoderSamplesPerFrame]int16
	fillEncoderSmoothPCM(pcm[:], 2)
	dst := []byte{1, 2, 3}

	raw := newTestEncoder(t, TransportRaw)
	defer raw.Close()
	got, _, consumed, ready, err := raw.EncodeSamplesInto(dst, []int16{1})
	if !errors.Is(err, ErrInvalidFrame) {
		t.Fatalf("misaligned samples err = %v, want ErrInvalidFrame", err)
	}
	if consumed != 0 || ready {
		t.Fatalf("misaligned samples consumed=%d ready=%v", consumed, ready)
	}
	if !bytes.Equal(got, dst) {
		t.Fatalf("misaligned samples changed dst: got %v, want %v", got, dst)
	}

	got, _, consumed, ready, err = raw.EncodeSamplesInto(dst, pcm[:128])
	if err != nil {
		t.Fatal(err)
	}
	if consumed != 128 || ready || !bytes.Equal(got, dst) {
		t.Fatalf("partial samples got=%v consumed=%d ready=%v", got, consumed, ready)
	}
	got, _, err = raw.EncodeRawInto(dst, pcm[:])
	if !errors.Is(err, ErrInvalidFrame) {
		t.Fatalf("exact encode with buffered samples err = %v, want ErrInvalidFrame", err)
	}
	if !bytes.Equal(got, dst) {
		t.Fatalf("exact encode with buffered samples changed dst: got %v, want %v", got, dst)
	}

	got, _, more, err := raw.FlushFrameInto(dst)
	if err != nil {
		t.Fatal(err)
	}
	if !more || !bytes.Equal(got[:len(dst)], dst) {
		t.Fatalf("flush after partial more=%v got prefix=%v want %v", more, got[:len(dst)], dst)
	}
	got, _, consumed, ready, err = raw.EncodeSamplesInto(dst, pcm[:])
	if !errors.Is(err, ErrClosed) {
		t.Fatalf("samples after flush err = %v, want ErrClosed", err)
	}
	if consumed != 0 || ready {
		t.Fatalf("samples after flush consumed=%d ready=%v", consumed, ready)
	}
	if !bytes.Equal(got, dst) {
		t.Fatalf("samples after flush changed dst: got %v, want %v", got, dst)
	}

	mixed := newTestEncoder(t, TransportRaw)
	defer mixed.Close()
	got, _, consumed, ready, err = mixed.EncodeRTMPSamplesInto(dst, pcm[:128])
	if err != nil {
		t.Fatal(err)
	}
	if consumed != 128 || ready || !bytes.Equal(got, dst) {
		t.Fatalf("partial RTMP samples got=%v consumed=%d ready=%v", got, consumed, ready)
	}
	got, _, consumed, ready, err = mixed.EncodeSamplesInto(dst, pcm[128:])
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("mixed output samples err = %v, want ErrInvalidConfig", err)
	}
	if consumed != 0 || ready {
		t.Fatalf("mixed output samples consumed=%d ready=%v", consumed, ready)
	}
	if !bytes.Equal(got, dst) {
		t.Fatalf("mixed output samples changed dst: got %v, want %v", got, dst)
	}
	got, _, more, err = mixed.FlushFrameInto(dst)
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("mixed output flush err = %v, want ErrInvalidConfig", err)
	}
	if more || !bytes.Equal(got, dst) {
		t.Fatalf("mixed output flush got=%v more=%v, want dst/no frame", got, more)
	}
	got, _, more, err = mixed.FlushRTMPMessageInto(dst)
	if err != nil {
		t.Fatal(err)
	}
	if !more || !bytes.Equal(got[:len(dst)], dst) {
		t.Fatalf("correct RTMP flush after mismatch more=%v got prefix=%v want %v", more, got[:len(dst)], dst)
	}
}

func TestDecoderConfigReturnsCopy(t *testing.T) {
	dec := newRawTestDecoder(t)
	defer dec.Close()

	cfg := dec.Config()
	if len(cfg.ExtraData) == 0 {
		t.Fatal("raw config ExtraData is empty")
	}
	cfg.ExtraData[0] ^= 0xff

	again := dec.Config()
	if bytes.Equal(cfg.ExtraData, again.ExtraData) {
		t.Fatal("Config returned alias to decoder ExtraData")
	}
}

func TestADTSReaderTruncatedFrameDoesNotAdvance(t *testing.T) {
	header, err := AppendADTSHeader(nil, Config{ObjectType: AOTAACLC, SampleRate: 44100, ChannelConfig: 2}, 4)
	if err != nil {
		t.Fatal(err)
	}
	reader := NewADTSReader(bytes.NewReader(append(header, 1, 2)))
	_, err = reader.ReadFrame(nil)
	if !errors.Is(err, ErrNeedMoreData) {
		t.Fatalf("truncated frame err = %v, want ErrNeedMoreData", err)
	}
	if reader.FrameIndex() != 0 || reader.Offset() != 0 {
		t.Fatalf("truncated frame advanced reader: index=%d offset=%d", reader.FrameIndex(), reader.Offset())
	}
}

func TestADTSFrameMetadataTransitions(t *testing.T) {
	m := loadTestVectorManifest(t)
	v := m.Vectors[0]
	frames, err := SplitADTSFrames(readTestVector(t, v.File))
	if err != nil {
		t.Fatal(err)
	}
	dec, err := New(Options{Transport: TransportADTS})
	if err != nil {
		t.Fatal(err)
	}
	defer dec.Close()
	if cfg := dec.Config(); cfg.SampleRate != 0 || cfg.Channels != 0 {
		t.Fatalf("initial ADTS config = %+v, want zero value", cfg)
	}

	var pcm []int16
	for i, frame := range frames {
		before := len(pcm)
		var info FrameInfo
		pcm, info, err = dec.Decode(pcm, frame.Data)
		if err != nil {
			t.Fatalf("frame %d: %v", i, err)
		}
		if info.Transport != TransportADTS {
			t.Fatalf("frame %d transport = %s, want %s", i, info.Transport, TransportADTS)
		}
		if info.InputBytes != frame.Header.FrameLength {
			t.Fatalf("frame %d input bytes = %d, want %d", i, info.InputBytes, frame.Header.FrameLength)
		}
		if info.OutputSamples != len(pcm)-before {
			t.Fatalf("frame %d output samples = %d, want %d", i, info.OutputSamples, len(pcm)-before)
		}
		if info.Config.SampleRate != v.SampleRate || info.Config.Channels != v.Channels {
			t.Fatalf("frame %d config = %+v, want %d Hz/%d ch", i, info.Config, v.SampleRate, v.Channels)
		}
		if i == 0 && info.OutputSamples != 0 {
			t.Fatalf("first frame output samples = %d, want decoder-delay transition with 0", info.OutputSamples)
		}
		if i > 0 && info.OutputSamples == 0 {
			t.Fatalf("frame %d output samples = 0, want PCM after decoder-delay transition", i)
		}
	}
	assertPCMHash(t, v, pcm)
}

func newRawTestDecoder(t *testing.T) *Decoder {
	t.Helper()
	dec, err := New(Options{
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
	return dec
}
