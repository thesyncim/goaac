# AAC-LC Reference Source

## Upstream Pin

- Project: FAAD2
- Repository: `https://github.com/knik0/faad2`
- Release tag: `2.11.2`
- Commit: `673a22a3c7c33e96e2ff7aae7c4d2bc190dfbf92`
- Local checkout: `third_party/faad2`

FAAD2 is used as the standalone C reference because its decoder core is compact,
well-known, and directly buildable as an oracle without pulling in a media
framework. The local port is generated with `LC_ONLY_DECODER` and `DISABLE_SBR`
so the checked-in decoder surface is AAC-LC only.

FFmpeg remains a fixture generator in tests, not a runtime dependency or decode
oracle.

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
