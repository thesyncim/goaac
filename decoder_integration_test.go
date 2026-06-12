package aac

import (
	"bytes"
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDecodeADTSMatchesFAAD2Oracle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg CLI not available")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler not available")
	}
	tmp := t.TempDir()
	input := filepath.Join(tmp, "sine.aac")
	oracle := filepath.Join(tmp, "oracle.pcm")

	run(t, ffmpeg,
		"-hide_banner", "-loglevel", "error",
		"-f", "lavfi",
		"-i", "sine=frequency=997:duration=0.10:sample_rate=44100",
		"-ac", "2",
		"-c:a", "aac",
		"-profile:a", "aac_low",
		"-b:a", "96k",
		"-f", "adts",
		input,
	)
	data, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}
	pcm, cfg, err := DecodeADTS(data)
	if err != nil {
		t.Fatal(err)
	}
	streamPCM, streamCfg, err := DecodeADTSReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(int16sToLE(streamPCM), int16sToLE(pcm)) {
		t.Fatalf("streaming PCM mismatch: got %d samples, want %d samples", len(streamPCM), len(pcm))
	}
	if !reflect.DeepEqual(streamCfg, cfg) {
		t.Fatalf("stream config = %+v, want %+v", streamCfg, cfg)
	}
	if cfg.ObjectType != AOTAACLC || cfg.SampleRate != 44100 || cfg.Channels != 2 {
		t.Fatalf("decoded config = %+v", cfg)
	}
	if len(pcm) == 0 {
		t.Fatal("decoded no PCM")
	}

	buildFAAD2Oracle(t, cc, filepath.Join(tmp, "faad2-oracle"))
	run(t, filepath.Join(tmp, "faad2-oracle"), input, oracle)

	want, err := os.ReadFile(oracle)
	if err != nil {
		t.Fatal(err)
	}
	got := int16sToLE(pcm)
	if !bytes.Equal(got, want) {
		t.Fatalf("PCM mismatch: got %d bytes, want %d bytes", len(got), len(want))
	}
}

func buildFAAD2Oracle(t *testing.T, cc, out string) {
	t.Helper()
	helper := filepath.Join(t.TempDir(), "faad2_oracle.c")
	if err := os.WriteFile(helper, []byte(faad2OracleSource), 0o600); err != nil {
		t.Fatal(err)
	}
	args := []string{
		"-std=c99",
		"-O2",
		"-ffp-contract=off",
		"-DLC_ONLY_DECODER",
		"-DDISABLE_SBR",
		"-DHAVE_INTTYPES_H=1",
		"-DHAVE_MEMCPY=1",
		"-DHAVE_STRING_H=1",
		"-DHAVE_STRINGS_H=1",
		"-DHAVE_SYS_STAT_H=1",
		"-DHAVE_SYS_TYPES_H=1",
		"-DHAVE_LRINTF=1",
		"-DPACKAGE_VERSION=\"2.11.2\"",
		"-DFAAD2_VERSION=\"2.11.2\"",
		"-I", "third_party/faad2/include",
		"-I", "third_party/faad2/libfaad",
		"-o", out,
		helper,
	}
	args = append(args, faad2OracleSources()...)
	args = append(args, "-lm")
	run(t, cc, args...)
}

func faad2OracleSources() []string {
	return []string{
		"third_party/faad2/libfaad/bits.c",
		"third_party/faad2/libfaad/cfft.c",
		"third_party/faad2/libfaad/common.c",
		"third_party/faad2/libfaad/decoder.c",
		"third_party/faad2/libfaad/drc.c",
		"third_party/faad2/libfaad/error.c",
		"third_party/faad2/libfaad/filtbank.c",
		"third_party/faad2/libfaad/huffman.c",
		"third_party/faad2/libfaad/is.c",
		"third_party/faad2/libfaad/mdct.c",
		"third_party/faad2/libfaad/mp4.c",
		"third_party/faad2/libfaad/ms.c",
		"third_party/faad2/libfaad/output.c",
		"third_party/faad2/libfaad/pns.c",
		"third_party/faad2/libfaad/pulse.c",
		"third_party/faad2/libfaad/specrec.c",
		"third_party/faad2/libfaad/syntax.c",
		"third_party/faad2/libfaad/tns.c",
	}
}

func run(t *testing.T, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, out)
	}
}

func int16sToLE(samples []int16) []byte {
	out := make([]byte, len(samples)*2)
	for i, s := range samples {
		binary.LittleEndian.PutUint16(out[i*2:], uint16(s))
	}
	return out
}

const faad2OracleSource = `
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include "neaacdec.h"

static int read_file(const char *path, unsigned char **data, unsigned long *size) {
    FILE *f = fopen(path, "rb");
    long n;
    if (!f) return 1;
    if (fseek(f, 0, SEEK_END) != 0) return 1;
    n = ftell(f);
    if (n < 0) return 1;
    if (fseek(f, 0, SEEK_SET) != 0) return 1;
    *data = (unsigned char *)malloc((size_t)n);
    if (!*data) return 1;
    if (fread(*data, 1, (size_t)n, f) != (size_t)n) return 1;
    fclose(f);
    *size = (unsigned long)n;
    return 0;
}

static unsigned long adts_frame_length(const unsigned char *data, unsigned long size) {
    unsigned long n;
    if (size < 7) return 0;
    if (data[0] != 0xff || (data[1] & 0xf0) != 0xf0) return 0;
    n = (((unsigned long)data[3] & 0x03) << 11) |
        ((unsigned long)data[4] << 3) |
        ((unsigned long)data[5] >> 5);
    if (n < 7 || n > size) return 0;
    return n;
}

int main(int argc, char **argv) {
    unsigned char *data = NULL;
    unsigned long size = 0, samplerate = 0;
    unsigned char channels = 0;
    unsigned long off = 0;
    NeAACDecHandle dec;
    NeAACDecConfigurationPtr cfg;
    FILE *out;
    unsigned long init_size;

    if (argc != 3) return 2;
    if (read_file(argv[1], &data, &size) != 0) return 3;
    out = fopen(argv[2], "wb");
    if (!out) return 4;

    dec = NeAACDecOpen();
    if (!dec) return 5;
    cfg = NeAACDecGetCurrentConfiguration(dec);
    cfg->defObjectType = LC;
    cfg->outputFormat = FAAD_FMT_16BIT;
    cfg->dontUpSampleImplicitSBR = 1;
    if (!NeAACDecSetConfiguration(dec, cfg)) return 6;
    init_size = adts_frame_length(data, size);
    if (init_size == 0) init_size = size;
    if (NeAACDecInit(dec, data, init_size, &samplerate, &channels) < 0) return 7;

    while (off < size) {
        NeAACDecFrameInfo info;
        unsigned long chunk = adts_frame_length(data + off, size - off);
        void *samples;
        if (chunk == 0) chunk = size - off;
        samples = NeAACDecDecode(dec, &info, data + off, chunk);
        if (info.error != 0) {
            fprintf(stderr, "faad2 decode error %u: %s\n", info.error, NeAACDecGetErrorMessage(info.error));
            return 8;
        }
        if (samples && info.samples) {
            fwrite(samples, sizeof(int16_t), (size_t)info.samples, out);
        }
        if (info.bytesconsumed == 0) break;
        off += chunk;
    }

    NeAACDecClose(dec);
    fclose(out);
    free(data);
    return 0;
}
`
