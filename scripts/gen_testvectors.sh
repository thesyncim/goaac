#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
outdir="$root/testdata/aaclc"
tmp="$(mktemp -d "${TMPDIR:-/tmp}/goaac-vectors.XXXXXX")"
trap 'rm -rf "$tmp"' EXIT

ffmpeg_bin="${FFMPEG:-ffmpeg}"
cc_bin="${CC:-cc}"
go_bin="${GO:-go}"
pcm_out_dir="${PCM_OUT_DIR:-}"

mkdir -p "$outdir"
if [ -n "$pcm_out_dir" ]; then
  mkdir -p "$pcm_out_dir"
fi

cat >"$tmp/faad2_oracle.c" <<'C'
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
C

"$cc_bin" \
  -std=c99 \
  -O2 \
  -ffp-contract=off \
  -DLC_ONLY_DECODER \
  -DDISABLE_SBR \
  -DHAVE_INTTYPES_H=1 \
  -DHAVE_MEMCPY=1 \
  -DHAVE_STRING_H=1 \
  -DHAVE_STRINGS_H=1 \
  -DHAVE_SYS_STAT_H=1 \
  -DHAVE_SYS_TYPES_H=1 \
  -DHAVE_LRINTF=1 \
  -DPACKAGE_VERSION='"2.11.2"' \
  -DFAAD2_VERSION='"2.11.2"' \
  -I "$root/third_party/faad2/include" \
  -I "$root/third_party/faad2/libfaad" \
  -o "$tmp/faad2-oracle" \
  "$tmp/faad2_oracle.c" \
  "$root/third_party/faad2/libfaad/bits.c" \
  "$root/third_party/faad2/libfaad/cfft.c" \
  "$root/third_party/faad2/libfaad/common.c" \
  "$root/third_party/faad2/libfaad/decoder.c" \
  "$root/third_party/faad2/libfaad/drc.c" \
  "$root/third_party/faad2/libfaad/error.c" \
  "$root/third_party/faad2/libfaad/filtbank.c" \
  "$root/third_party/faad2/libfaad/huffman.c" \
  "$root/third_party/faad2/libfaad/is.c" \
  "$root/third_party/faad2/libfaad/mdct.c" \
  "$root/third_party/faad2/libfaad/mp4.c" \
  "$root/third_party/faad2/libfaad/ms.c" \
  "$root/third_party/faad2/libfaad/output.c" \
  "$root/third_party/faad2/libfaad/pns.c" \
  "$root/third_party/faad2/libfaad/pulse.c" \
  "$root/third_party/faad2/libfaad/specrec.c" \
  "$root/third_party/faad2/libfaad/syntax.c" \
  "$root/third_party/faad2/libfaad/tns.c" \
  -lm

"$ffmpeg_bin" -hide_banner -nostdin -loglevel error -y -bitexact \
  -f lavfi -i "sine=frequency=997:duration=0.10:sample_rate=44100" \
  -ac 2 -c:a aac -profile:a aac_low -b:a 96k -f adts \
  "$outdir/sine_44100_stereo_96k.aac"

"$ffmpeg_bin" -hide_banner -nostdin -loglevel error -y -bitexact \
  -f lavfi -i "sine=frequency=523.25:duration=0.12:sample_rate=48000" \
  -ac 1 -c:a aac -profile:a aac_low -b:a 64k -f adts \
  "$outdir/sine_48000_mono_64k.aac"

"$ffmpeg_bin" -hide_banner -nostdin -loglevel error -y -bitexact \
  -f lavfi -i "sine=frequency=440:duration=0.12:sample_rate=32000" \
  -f lavfi -i "sine=frequency=1760:duration=0.12:sample_rate=32000" \
  -filter_complex "[0:a][1:a]amerge=inputs=2[a]" -map "[a]" \
  -c:a aac -profile:a aac_low -b:a 80k -f adts \
  "$outdir/dual_tone_32000_stereo_80k.aac"

"$ffmpeg_bin" -hide_banner -nostdin -loglevel error -y -bitexact \
  -f lavfi -i "anullsrc=r=22050:cl=mono:d=0.08" \
  -c:a aac -profile:a aac_low -b:a 32k -f adts \
  "$outdir/silence_22050_mono_32k.aac"

for f in "$outdir"/*.aac; do
  "$tmp/faad2-oracle" "$f" "$tmp/$(basename "${f%.aac}").pcm"
  if [ -n "$pcm_out_dir" ]; then
    cp "$tmp/$(basename "${f%.aac}").pcm" "$pcm_out_dir/"
  fi
done

cat >"$tmp/write_manifest.go" <<'GO'
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	aac "github.com/thesyncim/goaac"
)

type manifest struct {
	Version     int      `json:"version"`
	Source     string   `json:"source"`
	Oracle     string   `json:"oracle"`
	Vectors    []vector `json:"vectors"`
}

type vector struct {
	Name        string `json:"name"`
	File        string `json:"file"`
	Description string `json:"description"`
	SampleRate  int    `json:"sample_rate"`
	Channels    int    `json:"channels"`
	Frames      int    `json:"frames"`
	InputBytes  int    `json:"input_bytes"`
	InputSHA256 string `json:"input_sha256"`
	PCMBytes    int    `json:"pcm_bytes"`
	PCMSHA256   string `json:"pcm_sha256"`
}

func main() {
	if len(os.Args) != 3 {
		panic("usage: write_manifest outdir pcmdir")
	}
	outdir, pcmdir := os.Args[1], os.Args[2]
	fixtures := []struct {
		name string
		desc string
	}{
		{"sine_44100_stereo_96k", "997 Hz sine, 44.1 kHz stereo, AAC-LC ADTS, 96 kbps"},
		{"sine_48000_mono_64k", "523.25 Hz sine, 48 kHz mono, AAC-LC ADTS, 64 kbps"},
		{"dual_tone_32000_stereo_80k", "440 Hz left and 1760 Hz right tones, 32 kHz stereo, AAC-LC ADTS, 80 kbps"},
		{"silence_22050_mono_32k", "Digital silence, 22.05 kHz mono, AAC-LC ADTS, 32 kbps"},
	}
	m := manifest{
		Version: 1,
		Source: "Generated by scripts/gen_testvectors.sh with FFmpeg AAC-LC ADTS inputs.",
		Oracle: "PCM SHA-256 values are little-endian signed 16-bit output from pinned FAAD2 2.11.2 built with LC_ONLY_DECODER and DISABLE_SBR, decoded one ADTS frame per NeAACDecDecode call.",
	}
	for _, fixture := range fixtures {
		file := fixture.name + ".aac"
		input, err := os.ReadFile(filepath.Join(outdir, file))
		if err != nil {
			panic(err)
		}
		frames, err := aac.SplitADTSFrames(input)
		if err != nil {
			panic(err)
		}
		if len(frames) == 0 {
			panic(fmt.Sprintf("%s: no frames", file))
		}
		pcm, err := os.ReadFile(filepath.Join(pcmdir, fixture.name+".pcm"))
		if err != nil {
			panic(err)
		}
		inputHash := sha256.Sum256(input)
		pcmHash := sha256.Sum256(pcm)
		h := frames[0].Header
		m.Vectors = append(m.Vectors, vector{
			Name:        fixture.name,
			File:        file,
			Description: fixture.desc,
			SampleRate:  h.SampleRate,
			Channels:    h.Channels,
			Frames:      len(frames),
			InputBytes:  len(input),
			InputSHA256: hex.EncodeToString(inputHash[:]),
			PCMBytes:    len(pcm),
			PCMSHA256:   hex.EncodeToString(pcmHash[:]),
		})
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(m); err != nil {
		panic(err)
	}
}
GO

(cd "$root" && "$go_bin" run "$tmp/write_manifest.go" "$outdir" "$tmp") >"$outdir/manifest.json"
