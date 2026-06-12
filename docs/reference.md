# AAC-LC Reference Source

## Decoder Upstream Pin

- Project: FAAD2
- Repository: `https://github.com/knik0/faad2`
- Release tag: `2.11.2`
- Commit: `673a22a3c7c33e96e2ff7aae7c4d2bc190dfbf92`
- Local checkout: `third_party/faad2`

FAAD2 is used as the standalone C reference because its decoder core is compact,
well-known, and directly buildable as an oracle without pulling in a media
framework. The local port is generated with `LC_ONLY_DECODER` and `DISABLE_SBR`
so the checked-in decoder surface is AAC-LC only.

FFmpeg remains a fixture generator for committed vectors and live integration
tests, not a runtime dependency or decode oracle. Oracle PCM hashes come from
the pinned FAAD2 source compiled with `-ffp-contract=off` to match the generated
Go port's explicit `float32` evaluation.

## Decoder Files

- `libfaad/bits.c`
- `libfaad/cfft.c`
- `libfaad/common.c`
- `libfaad/decoder.c`
- `libfaad/drc.c`
- `libfaad/error.c`
- `libfaad/filtbank.c`
- `libfaad/huffman.c`
- `libfaad/is.c`
- `libfaad/mdct.c`
- `libfaad/mp4.c`
- `libfaad/ms.c`
- `libfaad/output.c`
- `libfaad/pns.c`
- `libfaad/pulse.c`
- `libfaad/specrec.c`
- `libfaad/syntax.c`
- `libfaad/tns.c`

## Local Production Slice

Supported now:

- AAC-LC profile.
- ADTS streams with one raw data block per frame.
- Raw AAC access units when configured with AAC-LC AudioSpecificConfig.
- Interleaved signed 16-bit PCM output.
- Pure-Go runtime on generated targets: `darwin/arm64` and `linux/arm64`.

Not claimed:

- HE-AAC SBR/PS profiles.
- AAC Main, SSR, LTP, LD, DRM, or error-resilience profiles.
- LATM/LOAS transport.
- ADTS frames carrying multiple raw data blocks.
- Generated support for targets that are not checked in under
  `internal/faad2ccgo`.

## Encoder Upstream Pin

- Project: Fraunhofer FDK AAC Codec Library for Android
- Repository: `https://github.com/mstorsjo/fdk-aac`
- Release tag: `v2.0.3`
- Commit: `716f4394641d53f0d79c9ddac3fa93b03a49f278`
- Local checkout: `third_party/fdk-aac`

FDK-AAC is the source truth for the AAC-LC encoder track. It is used as a pinned
native oracle and source-shaped translation reference only; the Go runtime must
remain pure Go and must not link or load FDK-AAC.

See `docs/encoder-reference.md` for the encoder module map, oracle command, and
completion gates.
