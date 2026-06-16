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

Supported:

- AAC-LC profile.
- ADTS streams with one raw data block per frame.
- Raw AAC access units configured with AAC-LC AudioSpecificConfig.
- RTMP/FLV AAC audio-message body parsing and decode.
- Interleaved signed 16-bit PCM output.
- Pure-Go runtime on generated targets: `darwin/arm64` and `linux/arm64`.

Not claimed:

- AAC encoding.
- HE-AAC SBR/PS profiles.
- AAC Main, SSR, LTP, LD, DRM, or error-resilience profiles.
- LATM/LOAS transport.
- ADTS frames carrying multiple raw data blocks.
- Generated support for targets that are not checked in under
  `internal/faad2ccgo`.

## Patent And Licensing Stance

This repository does not provide, claim, or imply any patent license or patent
non-infringement guarantee. The production code intentionally remains
decoder-only and does not carry an AAC encoder port or encoder source checkout.
Users and redistributors are responsible for evaluating patent obligations in
their own jurisdictions and use cases.
