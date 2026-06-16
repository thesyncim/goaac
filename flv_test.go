package aac

import (
	"errors"
	"testing"
)

const testFLVAACHeader = byte(FLVSoundFormatAAC<<4) |
	byte(FLVSoundRate44100Hz<<2) |
	byte(FLVSoundSize16Bit<<1) |
	byte(FLVSoundTypeStereo)

func TestParseFLVAACMessagePayloads(t *testing.T) {
	cfg := Config{ObjectType: AOTAACLC, SampleRate: 44100, ChannelConfig: 2}
	seq := appendTestFLVAACSequenceHeader(t, []byte{0xee}, cfg)
	if seq[0] != 0xee {
		t.Fatalf("sequence header prefix = %#x, want 0xee", seq[0])
	}
	tag, err := ParseRTMPAudioMessage(seq[1:])
	if err != nil {
		t.Fatal(err)
	}
	if tag.SoundFormat != FLVSoundFormatAAC || tag.AACPacketType != FLVAACPacketTypeSequenceHeader {
		t.Fatalf("sequence header tag = %+v, want AAC sequence header", tag)
	}
	parsed, err := ParseAudioSpecificConfig(tag.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.ObjectType != AOTAACLC || parsed.SampleRate != 44100 || parsed.Channels != 2 {
		t.Fatalf("sequence header config = %+v, want AAC-LC 44100 Hz/2 ch", parsed)
	}

	raw := appendTestFLVAACRawTag([]byte{0xdd}, []byte{1, 2, 3})
	if raw[0] != 0xdd {
		t.Fatalf("raw tag prefix = %#x, want 0xdd", raw[0])
	}
	tag, err = ParseFLVAudioTag(raw[1:])
	if err != nil {
		t.Fatal(err)
	}
	if tag.SoundFormat != FLVSoundFormatAAC || tag.AACPacketType != FLVAACPacketTypeRaw {
		t.Fatalf("raw tag = %+v, want AAC raw", tag)
	}
	if len(tag.Payload) != 3 || tag.Payload[0] != 1 || tag.Payload[1] != 2 || tag.Payload[2] != 3 {
		t.Fatalf("raw payload = %v, want [1 2 3]", tag.Payload)
	}
}

func TestParseFLVAudioTagAAC(t *testing.T) {
	data := []byte{testFLVAACHeader, byte(FLVAACPacketTypeRaw), 0x11, 0x22, 0x33}
	tag, err := ParseFLVAudioTag(data)
	if err != nil {
		t.Fatal(err)
	}
	if tag.SoundFormat != FLVSoundFormatAAC {
		t.Fatalf("sound format = %s, want %s", tag.SoundFormat, FLVSoundFormatAAC)
	}
	if tag.SoundRate != FLVSoundRate44100Hz || tag.SoundRate.Hertz() != 44100 {
		t.Fatalf("sound rate = %d/%d Hz, want 3/44100 Hz", tag.SoundRate, tag.SoundRate.Hertz())
	}
	if tag.SoundSize != FLVSoundSize16Bit {
		t.Fatalf("sound size = %d, want %d", tag.SoundSize, FLVSoundSize16Bit)
	}
	if tag.SoundType != FLVSoundTypeStereo {
		t.Fatalf("sound type = %d, want %d", tag.SoundType, FLVSoundTypeStereo)
	}
	if tag.AACPacketType != FLVAACPacketTypeRaw {
		t.Fatalf("AAC packet type = %s, want %s", tag.AACPacketType, FLVAACPacketTypeRaw)
	}
	if len(tag.Payload) != 3 {
		t.Fatalf("payload len = %d, want 3", len(tag.Payload))
	}
	tag.Payload[0] = 0x99
	if data[2] != 0x99 {
		t.Fatal("payload does not alias input")
	}

	rtmpTag, err := ParseRTMPAudioMessage(data)
	if err != nil {
		t.Fatal(err)
	}
	if rtmpTag.SoundFormat != tag.SoundFormat || rtmpTag.AACPacketType != tag.AACPacketType {
		t.Fatalf("RTMP alias parse = %+v, want matching FLV parse %+v", rtmpTag, tag)
	}
}

func TestParseFLVAudioTagControls(t *testing.T) {
	if _, err := ParseFLVAudioTag(nil); !errors.Is(err, ErrNeedMoreData) {
		t.Fatalf("empty tag err = %v, want ErrNeedMoreData", err)
	}
	if _, err := ParseFLVAudioTag([]byte{testFLVAACHeader}); !errors.Is(err, ErrNeedMoreData) {
		t.Fatalf("short AAC tag err = %v, want ErrNeedMoreData", err)
	}

	data := []byte{byte(FLVSoundFormatMP3 << 4), 0xaa, 0xbb}
	tag, err := ParseFLVAudioTag(data)
	if err != nil {
		t.Fatal(err)
	}
	if tag.SoundFormat != FLVSoundFormatMP3 {
		t.Fatalf("sound format = %s, want %s", tag.SoundFormat, FLVSoundFormatMP3)
	}
	if len(tag.Payload) != 2 {
		t.Fatalf("payload len = %d, want 2", len(tag.Payload))
	}
	tag.Payload[0] = 0xcc
	if data[1] != 0xcc {
		t.Fatal("non-AAC payload does not alias input")
	}
}

func TestParseFLVAudioTagAllocs(t *testing.T) {
	data := []byte{testFLVAACHeader, byte(FLVAACPacketTypeRaw), 1, 2, 3, 4}
	allocs := testing.AllocsPerRun(1000, func() {
		tag, err := ParseFLVAudioTag(data)
		if err != nil {
			t.Fatal(err)
		}
		if tag.SoundFormat != FLVSoundFormatAAC || len(tag.Payload) != 4 {
			t.Fatal("unexpected parsed tag")
		}
	})
	if allocs != 0 {
		t.Fatalf("ParseFLVAudioTag allocations = %v, want 0", allocs)
	}
}

func TestFLVAACDecoderVectors(t *testing.T) {
	m := loadTestVectorManifest(t)
	for _, v := range m.Vectors {
		t.Run(v.Name, func(t *testing.T) {
			frames, err := SplitADTSFrames(readTestVector(t, v.File))
			if err != nil {
				t.Fatal(err)
			}
			if len(frames) == 0 {
				t.Fatal("no frames")
			}

			dec := NewRTMPAACDecoder()
			defer dec.Close()

			prefix := []int16{-11, 22}
			pcm := append([]int16(nil), prefix...)
			seq := flvAACSequenceHeaderFromADTS(t, frames[0].Header)
			pcm, info, err := dec.DecodeRTMPMessage(pcm, seq)
			if err != nil {
				t.Fatal(err)
			}
			if !info.SequenceHeader {
				t.Fatal("sequence header info flag is false")
			}
			if info.InputBytes != len(seq) {
				t.Fatalf("sequence header input bytes = %d, want %d", info.InputBytes, len(seq))
			}
			if info.OutputSamples != 0 || !equalInt16s(pcm, prefix) {
				t.Fatalf("sequence header changed PCM: samples=%d pcm=%v", info.OutputSamples, pcm)
			}
			assertVectorConfig(t, v, info.Config)
			assertVectorConfig(t, v, dec.Config())

			for i, frame := range frames {
				tag := flvAACRawTagFromADTSFrame(t, frame)
				before := len(pcm)
				pcm, info, err = dec.DecodeTag(pcm, tag)
				if err != nil {
					t.Fatalf("FLV raw frame %d: %v", i, err)
				}
				if info.SequenceHeader {
					t.Fatalf("raw frame %d reported sequence header", i)
				}
				if info.Tag.SoundFormat != FLVSoundFormatAAC || info.Tag.AACPacketType != FLVAACPacketTypeRaw {
					t.Fatalf("raw frame %d tag = %+v, want AAC raw", i, info.Tag)
				}
				if info.InputBytes != len(tag) {
					t.Fatalf("raw frame %d input bytes = %d, want %d", i, info.InputBytes, len(tag))
				}
				if info.Frame.Transport != TransportRaw {
					t.Fatalf("raw frame %d transport = %s, want %s", i, info.Frame.Transport, TransportRaw)
				}
				if info.Frame.InputBytes != frame.Header.PayloadLength {
					t.Fatalf("raw frame %d payload bytes = %d, want %d", i, info.Frame.InputBytes, frame.Header.PayloadLength)
				}
				if info.OutputSamples != len(pcm)-before || info.Frame.OutputSamples != info.OutputSamples {
					t.Fatalf("raw frame %d output samples = wrapper %d frame %d appended %d", i, info.OutputSamples, info.Frame.OutputSamples, len(pcm)-before)
				}
				if i == 0 && info.OutputSamples != 0 {
					t.Fatalf("first raw frame output samples = %d, want decoder-delay transition with 0", info.OutputSamples)
				}
				if i > 0 && info.OutputSamples == 0 {
					t.Fatalf("raw frame %d output samples = 0, want PCM after decoder-delay transition", i)
				}
			}
			if !equalInt16s(pcm[:len(prefix)], prefix) {
				t.Fatalf("DecodeTag modified prefix: got %v, want %v", pcm[:len(prefix)], prefix)
			}
			assertPCMHash(t, v, pcm[len(prefix):])
			assertVectorConfig(t, v, dec.Config())
		})
	}
}

func TestFLVAACDecoderControlsAndTransitions(t *testing.T) {
	dec := NewFLVAACDecoder()
	dst := []int16{1, 2, 3}

	raw := []byte{testFLVAACHeader, byte(FLVAACPacketTypeRaw), 0}
	got, info, err := dec.DecodeTag(dst, raw)
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("raw before sequence header err = %v, want ErrInvalidConfig", err)
	}
	if !equalInt16s(got, dst) {
		t.Fatalf("raw before sequence header changed dst: got %v, want %v", got, dst)
	}
	if info.Tag.AACPacketType != FLVAACPacketTypeRaw || info.InputBytes != len(raw) {
		t.Fatalf("raw before sequence header info = %+v", info)
	}

	mp3 := []byte{byte(FLVSoundFormatMP3 << 4), 0}
	got, _, err = dec.DecodeTag(dst, mp3)
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Fatalf("non-AAC err = %v, want ErrUnsupportedFormat", err)
	}
	if !equalInt16s(got, dst) {
		t.Fatalf("non-AAC changed dst: got %v, want %v", got, dst)
	}

	badPacketType := []byte{testFLVAACHeader, 2}
	got, _, err = dec.DecodeTag(dst, badPacketType)
	if !errors.Is(err, ErrInvalidFLV) {
		t.Fatalf("bad AAC packet type err = %v, want ErrInvalidFLV", err)
	}
	if !equalInt16s(got, dst) {
		t.Fatalf("bad AAC packet type changed dst: got %v, want %v", got, dst)
	}

	stereo := Config{ObjectType: AOTAACLC, SampleRate: 44100, ChannelConfig: 2}
	mono := Config{ObjectType: AOTAACLC, SampleRate: 48000, ChannelConfig: 1}
	got, info, err = dec.DecodeTag(dst, flvAACSequenceHeaderFromConfig(t, stereo))
	if err != nil {
		t.Fatal(err)
	}
	if !info.SequenceHeader || !equalInt16s(got, dst) {
		t.Fatalf("stereo sequence header info=%+v got=%v", info, got)
	}
	if cfg := dec.Config(); cfg.SampleRate != 44100 || cfg.Channels != 2 {
		t.Fatalf("stereo config = %+v, want 44100 Hz/2 ch", cfg)
	}

	got, info, err = dec.DecodeTag(dst, flvAACSequenceHeaderFromConfig(t, mono))
	if err != nil {
		t.Fatal(err)
	}
	if !info.SequenceHeader || !equalInt16s(got, dst) {
		t.Fatalf("mono sequence header info=%+v got=%v", info, got)
	}
	if cfg := dec.Config(); cfg.SampleRate != 48000 || cfg.Channels != 1 {
		t.Fatalf("mono config = %+v, want 48000 Hz/1 ch", cfg)
	}

	got, _, err = dec.DecodeTag(dst, []byte{testFLVAACHeader, byte(FLVAACPacketTypeSequenceHeader), 0xff})
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("bad sequence header err = %v, want ErrInvalidConfig", err)
	}
	if !equalInt16s(got, dst) {
		t.Fatalf("bad sequence header changed dst: got %v, want %v", got, dst)
	}
	if cfg := dec.Config(); cfg.SampleRate != 48000 || cfg.Channels != 1 {
		t.Fatalf("bad sequence header changed config to %+v", cfg)
	}
}

func TestFLVAACDecoderClose(t *testing.T) {
	dec := NewFLVAACDecoder()
	if err := dec.Close(); err != nil {
		t.Fatal(err)
	}
	if err := dec.Close(); err != nil {
		t.Fatalf("second close = %v, want nil", err)
	}
	dst := []int16{7, 8}
	got, _, err := dec.DecodeTag(dst, []byte{testFLVAACHeader, byte(FLVAACPacketTypeRaw)})
	if !errors.Is(err, ErrClosed) {
		t.Fatalf("DecodeTag after close err = %v, want ErrClosed", err)
	}
	if !equalInt16s(got, dst) {
		t.Fatalf("DecodeTag after close changed dst: got %v, want %v", got, dst)
	}
}

func flvAACSequenceHeaderFromADTS(t *testing.T, h ADTSHeader) []byte {
	t.Helper()
	return flvAACSequenceHeaderFromConfig(t, h.Config())
}

func flvAACSequenceHeaderFromConfig(t *testing.T, cfg Config) []byte {
	t.Helper()
	return appendTestFLVAACSequenceHeader(t, nil, cfg)
}

func appendTestFLVAACSequenceHeader(t *testing.T, dst []byte, cfg Config) []byte {
	t.Helper()
	normalized, err := normalizeRawConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	extra := buildAudioSpecificConfig(normalized)
	dst = append(dst, testFLVAACHeader, byte(FLVAACPacketTypeSequenceHeader))
	return append(dst, extra...)
}

func appendTestFLVAACRawTag(dst, accessUnit []byte) []byte {
	dst = append(dst, testFLVAACHeader, byte(FLVAACPacketTypeRaw))
	return append(dst, accessUnit...)
}

func flvAACRawTagFromADTSFrame(t *testing.T, frame ADTSFrame) []byte {
	t.Helper()
	payload := frame.Payload()
	if payload == nil {
		t.Fatal("invalid ADTS frame bounds")
	}
	tag := make([]byte, 2+len(payload))
	tag[0] = testFLVAACHeader
	tag[1] = byte(FLVAACPacketTypeRaw)
	copy(tag[2:], payload)
	return tag
}
