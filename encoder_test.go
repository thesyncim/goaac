package aac

import (
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
	if info.Transport != TransportADTS || info.ADTSHeaderBytes != 7 || info.PayloadBytes != 341 {
		t.Fatalf("ADTS info = %+v", info)
	}
	h, err := ParseADTSHeader(out)
	if err != nil {
		t.Fatal(err)
	}
	if h.FrameLength != len(out) || h.SampleRate != 48000 || h.Channels != 2 {
		t.Fatalf("ADTS header = %+v len=%d", h, len(out))
	}
	if got, want := sha256Hex(out[info.ADTSHeaderBytes:]), "86738e2a79887cb24c6c5897dc78acdf3fdd8d6d79dc14cf5412dfc368b60641"; got != want {
		t.Fatalf("ADTS payload sha256 = %s, want %s", got, want)
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
