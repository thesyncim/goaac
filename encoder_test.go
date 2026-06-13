package aac

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
)

var encoderAPILenSink int

func TestEncoderRawVector(t *testing.T) {
	enc := newTestEncoder(t, TransportRaw)
	defer enc.Close()
	var pcm [2 * encoderSamplesPerFrame]int16
	fillEncoderSmoothPCM(pcm[:], 2)
	var storage [512]byte

	out, info, err := enc.EncodeRawInto(storage[:0], pcm[:])
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 341 || info.PayloadBytes != 341 || info.OutputBytes != 341 {
		t.Fatalf("raw lengths = len %d info %+v", len(out), info)
	}
	if info.Transport != TransportRaw || info.SampleRate != 48000 || info.Channels != 2 || info.BitRate != 128000 {
		t.Fatalf("raw info = %+v", info)
	}
	if !info.QuantizationDone || info.ChannelElements != 1 || info.GlobalFillExtensions != 1 {
		t.Fatalf("raw encode state = %+v", info)
	}
	if got, want := sha256Hex(out), "86738e2a79887cb24c6c5897dc78acdf3fdd8d6d79dc14cf5412dfc368b60641"; got != want {
		t.Fatalf("raw sha256 = %s, want %s", got, want)
	}
}

func TestEncoderRawMonoVector(t *testing.T) {
	enc, err := NewEncoder(EncoderOptions{
		Config: Config{
			ObjectType:    AOTAACLC,
			SampleRate:    48000,
			ChannelConfig: 1,
		},
		BitRate:   64000,
		Transport: TransportRaw,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer enc.Close()
	var pcm [encoderSamplesPerFrame]int16
	fillEncoderSmoothPCM(pcm[:], 1)
	var storage [512]byte

	out, info, err := enc.EncodeRawInto(storage[:0], pcm[:])
	if err != nil {
		t.Fatal(err)
	}
	if info.Transport != TransportRaw || info.SampleRate != 48000 || info.Channels != 1 || info.BitRate != 64000 {
		t.Fatalf("mono raw info = %+v", info)
	}
	if info.ChannelElements != 1 || info.PayloadBytes != len(out) || info.PayloadBytes == 0 {
		t.Fatalf("mono raw lengths/state = len %d info %+v", len(out), info)
	}
	if got, want := sha256Hex(out), "1b844a1a68e3540e53c5b4185d2596b65ccaba4eb7a3062905d5fc35efb4ca65"; got != want {
		t.Fatalf("mono raw sha256 = %s, want %s; len=%d info=%+v", got, want, len(out), info)
	}
}

func TestEncoderADTSVector(t *testing.T) {
	enc := newTestEncoder(t, TransportADTS)
	defer enc.Close()
	var pcm [2 * encoderSamplesPerFrame]int16
	fillEncoderSmoothPCM(pcm[:], 2)
	var storage [512]byte

	out, info, err := enc.Encode(storage[:0], pcm[:])
	if err != nil {
		t.Fatal(err)
	}
	if info.Transport != TransportADTS || info.ADTSHeaderBytes != 7 || info.PayloadBytes != 334 || info.TransportStaticBits != 56 {
		t.Fatalf("ADTS info = %+v", info)
	}
	h, err := ParseADTSHeader(out)
	if err != nil {
		t.Fatal(err)
	}
	if h.FrameLength != len(out) || h.SampleRate != 48000 || h.Channels != 2 || h.BufferFullness != 149 {
		t.Fatalf("ADTS header = %+v len=%d", h, len(out))
	}
	if got, want := sha256Hex(out[info.ADTSHeaderBytes:]), "e31b33b5a166df7b387c8d3ca682a78e02f1bf161536f751cfadc96b093c0296"; got != want {
		t.Fatalf("ADTS payload sha256 = %s, want %s", got, want)
	}
}

func TestEncoderFlushADTSVector(t *testing.T) {
	enc := newTestEncoder(t, TransportADTS)
	defer enc.Close()
	var pcm [2 * encoderSamplesPerFrame]int16
	fillEncoderSmoothPCM(pcm[:], 2)

	stream, info, err := enc.EncodeADTSFrameInto(nil, pcm[:])
	if err != nil {
		t.Fatal(err)
	}
	payloadBytes := info.PayloadBytes
	flushFrames := 0
	for {
		before := len(stream)
		var more bool
		stream, info, more, err = enc.FlushFrameInto(stream)
		if err != nil {
			t.Fatalf("flush frame %d: %v", flushFrames, err)
		}
		if !more {
			break
		}
		if info.InputSamples != 0 || info.Transport != TransportADTS || info.ADTSHeaderBytes != ADTSHeaderSize {
			t.Fatalf("flush frame %d info = %+v", flushFrames, info)
		}
		if info.OutputBytes != len(stream)-before || info.PayloadBytes <= 0 {
			t.Fatalf("flush frame %d lengths = before %d after %d info %+v", flushFrames, before, len(stream), info)
		}
		payloadBytes += info.PayloadBytes
		flushFrames++
	}
	if flushFrames != 2 {
		t.Fatalf("flush frames = %d, want 2", flushFrames)
	}
	frames, err := SplitADTSFrames(stream)
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 3 {
		t.Fatalf("ADTS frames after flush = %d, want 3", len(frames))
	}
	if gotPayload := len(stream) - len(frames)*ADTSHeaderSize; gotPayload != payloadBytes {
		t.Fatalf("payload bytes = %d, want encoder sum %d", gotPayload, payloadBytes)
	}
	if got, want := sha256Hex(stream), "f14dd64a7f69bbf60f5e7b292df01032469019b42be4b1386828b3dff5cc985a"; got != want {
		t.Fatalf("flushed ADTS sha256 = %s, want %s; len=%d payload=%d", got, want, len(stream), payloadBytes)
	}
	got, _, err := enc.EncodeADTSFrameInto(stream, pcm[:])
	if !errors.Is(err, ErrClosed) {
		t.Fatalf("encode after flush err = %v, want ErrClosed", err)
	}
	if !bytes.Equal(got, stream) {
		t.Fatalf("encode after flush changed dst: got len %d, want %d", len(got), len(stream))
	}
}

func TestEncoderADTSMultiFrameTransitionRoundTrip(t *testing.T) {
	enc := newTestEncoder(t, TransportADTS)
	defer enc.Close()
	var frame [2 * encoderSamplesPerFrame]int16
	var stream []byte
	var payloadBytes int

	for i := 0; i < 6; i++ {
		fillEncoderTransitionPCM(frame[:], 2, i)
		before := len(stream)
		var info EncodedFrameInfo
		var err error
		stream, info, err = enc.EncodeADTSFrameInto(stream, frame[:])
		if err != nil {
			t.Fatalf("encode frame %d: %v", i, err)
		}
		if !info.QuantizationDone || info.Transport != TransportADTS || info.ADTSHeaderBytes != ADTSHeaderSize {
			t.Fatalf("frame %d info = %+v", i, info)
		}
		if info.OutputBytes != len(stream)-before || info.PayloadBytes <= 0 {
			t.Fatalf("frame %d lengths = before %d after %d info %+v", i, before, len(stream), info)
		}
		payloadBytes += info.PayloadBytes
	}
	flushFrames := 0
	for {
		before := len(stream)
		var info EncodedFrameInfo
		var more bool
		var err error
		stream, info, more, err = enc.FlushFrameInto(stream)
		if err != nil {
			t.Fatalf("flush frame %d: %v", flushFrames, err)
		}
		if !more {
			break
		}
		if info.InputSamples != 0 || !info.QuantizationDone || info.Transport != TransportADTS || info.ADTSHeaderBytes != ADTSHeaderSize {
			t.Fatalf("flush frame %d info = %+v", flushFrames, info)
		}
		if info.OutputBytes != len(stream)-before || info.PayloadBytes <= 0 {
			t.Fatalf("flush frame %d lengths = before %d after %d info %+v", flushFrames, before, len(stream), info)
		}
		payloadBytes += info.PayloadBytes
		flushFrames++
	}
	if flushFrames != 2 {
		t.Fatalf("flush frames = %d, want 2", flushFrames)
	}

	frames, err := SplitADTSFrames(stream)
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 8 {
		t.Fatalf("ADTS frames = %d, want 8", len(frames))
	}
	if gotPayload := len(stream) - len(frames)*ADTSHeaderSize; gotPayload != payloadBytes {
		t.Fatalf("payload bytes = %d, want encoder sum %d", gotPayload, payloadBytes)
	}

	dec, err := NewADTSDecoder()
	if err != nil {
		t.Fatal(err)
	}
	defer dec.Close()
	var pcm []int16
	for i, frame := range frames {
		before := len(pcm)
		var info FrameInfo
		pcm, info, err = dec.DecodeADTSFrameInto(pcm, frame.Data)
		if err != nil {
			t.Fatalf("decode frame %d: %v", i, err)
		}
		if info.SampleRate != 48000 || info.Channels != 2 || info.InputBytes != frame.Header.FrameLength {
			t.Fatalf("decode frame %d info = %+v header=%+v", i, info, frame.Header)
		}
		if i == 0 && info.OutputSamples != 0 {
			t.Fatalf("first decode frame output = %d, want decoder delay", info.OutputSamples)
		}
		if i > 1 && len(pcm) == before {
			t.Fatalf("decode frame %d produced no PCM after warmup", i)
		}
	}
	if len(pcm) == 0 {
		t.Fatal("roundtrip produced no PCM")
	}
	if got, want := sha256Hex(stream), "6e203648c4472d40570bb69e5a4ca0effc264214b8b1d2540c8bc47ecb94bcaf"; got != want {
		t.Fatalf("transition stream sha256 = %s, want %s; len=%d payload=%d", got, want, len(stream), payloadBytes)
	}
	if got, want := sha256Int16(pcm), "4be31c47c811a91e45375f4e2b6300e59117530c73edfeef2daea82370966e85"; got != want {
		t.Fatalf("transition PCM sha256 = %s, want %s; samples=%d", got, want, len(pcm))
	}
}

func TestEncoderRTMPVector(t *testing.T) {
	enc := newTestEncoder(t, TransportRaw)
	defer enc.Close()
	var pcm [2 * encoderSamplesPerFrame]int16
	fillEncoderSmoothPCM(pcm[:], 2)

	seq, err := enc.AppendRTMPSequenceHeader(nil)
	if err != nil {
		t.Fatal(err)
	}
	seqTag, err := ParseRTMPAudioMessage(seq)
	if err != nil {
		t.Fatal(err)
	}
	if seqTag.SoundFormat != FLVSoundFormatAAC || seqTag.AACPacketType != FLVAACPacketTypeSequenceHeader {
		t.Fatalf("sequence header tag = %+v", seqTag)
	}
	if cfg, err := ParseAudioSpecificConfig(seqTag.Payload); err != nil || cfg.SampleRate != 48000 || cfg.Channels != 2 {
		t.Fatalf("sequence config = %+v err=%v", cfg, err)
	}

	msg, info, err := enc.EncodeRTMPMessageInto(nil, pcm[:])
	if err != nil {
		t.Fatal(err)
	}
	if info.PayloadBytes != 341 || info.OutputBytes != 343 {
		t.Fatalf("RTMP info = %+v", info)
	}
	rawTag, err := ParseRTMPAudioMessage(msg)
	if err != nil {
		t.Fatal(err)
	}
	if rawTag.SoundFormat != FLVSoundFormatAAC || rawTag.AACPacketType != FLVAACPacketTypeRaw {
		t.Fatalf("raw tag = %+v", rawTag)
	}
	if got, want := sha256Hex(rawTag.Payload), "86738e2a79887cb24c6c5897dc78acdf3fdd8d6d79dc14cf5412dfc368b60641"; got != want {
		t.Fatalf("RTMP payload sha256 = %s, want %s", got, want)
	}
}

func TestEncoderRTMPFlushFrames(t *testing.T) {
	enc := newTestEncoder(t, TransportRaw)
	defer enc.Close()
	var pcm [2 * encoderSamplesPerFrame]int16
	fillEncoderSmoothPCM(pcm[:], 2)

	msg, _, err := enc.EncodeRTMPMessageInto(nil, pcm[:])
	if err != nil {
		t.Fatal(err)
	}
	flushFrames := 0
	for {
		before := len(msg)
		var info EncodedFrameInfo
		var more bool
		msg, info, more, err = enc.FlushRTMPMessageInto(msg)
		if err != nil {
			t.Fatalf("RTMP flush frame %d: %v", flushFrames, err)
		}
		if !more {
			break
		}
		if info.InputSamples != 0 || info.Transport != TransportRaw || info.PayloadBytes <= 0 || info.OutputBytes != len(msg)-before {
			t.Fatalf("RTMP flush frame %d info = %+v", flushFrames, info)
		}
		tag, err := ParseRTMPAudioMessage(msg[before:])
		if err != nil {
			t.Fatalf("parse RTMP flush frame %d: %v", flushFrames, err)
		}
		if tag.SoundFormat != FLVSoundFormatAAC || tag.AACPacketType != FLVAACPacketTypeRaw || len(tag.Payload) != info.PayloadBytes {
			t.Fatalf("RTMP flush frame %d tag = %+v info=%+v", flushFrames, tag, info)
		}
		flushFrames++
	}
	if flushFrames != 2 {
		t.Fatalf("RTMP flush frames = %d, want 2", flushFrames)
	}
	dst := []byte{1, 2, 3}
	got, _, err := enc.EncodeRTMPMessageInto(dst, pcm[:])
	if !errors.Is(err, ErrClosed) {
		t.Fatalf("RTMP encode after flush err = %v, want ErrClosed", err)
	}
	if !bytes.Equal(got, dst) {
		t.Fatalf("RTMP encode after flush changed dst: got %v, want %v", got, dst)
	}
}

func TestEncoderRejectsInvalid(t *testing.T) {
	_, err := NewEncoder(EncoderOptions{
		Config:  Config{ObjectType: AOTAACLC, SampleRate: 48000, Channels: 6},
		BitRate: 320000,
	})
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Fatalf("6-channel err = %v, want ErrUnsupportedFormat", err)
	}
	_, err = NewEncoder(EncoderOptions{
		Config:  Config{ObjectType: AOTAACLC, SampleRate: 48000, ChannelConfig: 2},
		BitRate: -1,
	})
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("bitrate err = %v, want ErrInvalidConfig", err)
	}

	enc := newTestEncoder(t, TransportRaw)
	if _, _, err = enc.EncodeRawInto(nil, make([]int16, 8)); !errors.Is(err, ErrInvalidFrame) {
		t.Fatalf("short PCM err = %v, want ErrInvalidFrame", err)
	}
	if err = enc.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, err = enc.EncodeRawInto(nil, make([]int16, 2*encoderSamplesPerFrame)); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed err = %v, want ErrClosed", err)
	}
}

func TestEncoderRawAllocs(t *testing.T) {
	enc := newTestEncoder(t, TransportRaw)
	defer enc.Close()
	var pcm [2 * encoderSamplesPerFrame]int16
	fillEncoderSmoothPCM(pcm[:], 2)
	var storage [512]byte

	allocs := testing.AllocsPerRun(100, func() {
		out, info, err := enc.EncodeRawInto(storage[:0], pcm[:])
		if err != nil {
			t.Fatal(err)
		}
		encoderAPILenSink = len(out) + info.PayloadBytes
	})
	if allocs != 0 {
		t.Fatalf("allocs = %.2f, want 0", allocs)
	}
}

func newTestEncoder(t *testing.T, transport Transport) *Encoder {
	t.Helper()
	enc, err := NewEncoder(EncoderOptions{
		Config: Config{
			ObjectType:    AOTAACLC,
			SampleRate:    48000,
			ChannelConfig: 2,
		},
		BitRate:   128000,
		Transport: transport,
	})
	if err != nil {
		t.Fatal(err)
	}
	return enc
}

func fillEncoderSmoothPCM(dst []int16, channels int) {
	for i := 0; i < encoderSamplesPerFrame; i++ {
		for ch := 0; ch < channels; ch++ {
			v := ((i + ch*7) % 64) - 32
			dst[i*channels+ch] = int16(v / 8)
		}
	}
}

func fillEncoderTransitionPCM(dst []int16, channels int, frame int) {
	clear(dst)
	switch frame {
	case 0, 5:
		return
	case 1, 4:
		fillEncoderSmoothPCM(dst, channels)
	case 2:
		for i := 0; i < encoderSamplesPerFrame; i++ {
			for ch := 0; ch < channels; ch++ {
				v := 0
				if i >= 736 && i < 832 {
					if (i+ch)&1 == 0 {
						v = 24000
					} else {
						v = -24000
					}
				}
				dst[i*channels+ch] = int16(v)
			}
		}
	default:
		for i := 0; i < encoderSamplesPerFrame; i++ {
			for ch := 0; ch < channels; ch++ {
				v := ((i*3 + ch*11) % 96) - 48
				dst[i*channels+ch] = int16(v * 32)
			}
		}
	}
}

func sha256Int16(samples []int16) string {
	h := sha256.New()
	var b [2]byte
	for _, sample := range samples {
		u := uint16(sample)
		b[0] = byte(u)
		b[1] = byte(u >> 8)
		_, _ = h.Write(b[:])
	}
	return hex.EncodeToString(h.Sum(nil))
}
