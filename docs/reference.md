# AAC-LC Reference Source

## Upstream Pin

- Project: FFmpeg
- Repository: `https://git.ffmpeg.org/ffmpeg.git`
- Release tag: `n8.1.1`
- Tag object: `150ba6ddfabb5c433bb2fb3ee546d2a96e59066d`
- Commit: `239f2c733de417201d7ad3b3b8b0d9b63285b2b1`
- Release page: `https://ffmpeg.org/download.html`
- Local checkout: `third_party/ffmpeg`

FFmpeg is used here because its native AAC decoder is actively maintained,
widely deployed, and performance-oriented. The reference pin is treated as part
of the behavioral spec along with libavcodec/libavutil/libswresample ABI
behavior.

## Decoder Files

- `libavcodec/aac/aacdec.c`
- `libavcodec/aac/aacdec.h`
- `libavcodec/aac/aacdec_float.c`
- `libavcodec/aac/aacdec_fixed.c`
- `libavcodec/aac/aacdec_tab.c`
- `libavcodec/aactab.c`
- `libavcodec/mpeg4audio.c`
- `libavcodec/mpeg4audio_sample_rates.c`
- `libavcodec/adts_header.c`
- `libavcodec/adts_parser.c`

## Local Production Slice

The current production slice links FFmpeg's AAC decoder dynamically and exposes a
small Go API for:

- AudioSpecificConfig validation for AAC-LC.
- ADTS header parsing and frame splitting.
- ADTS AAC-LC decode to interleaved S16 PCM.
- Raw AAC access-unit decode when supplied with AAC-LC AudioSpecificConfig.

## Intentional Scope

Supported now:

- AAC-LC profile.
- ADTS streams with one raw data block per frame.
- Raw AAC access units when configured with AAC-LC AudioSpecificConfig.
- Interleaved signed 16-bit PCM output.

Not yet claimed as pure Go parity:

- Spectral decode, Huffman decode, prediction, TNS, MS/IS stereo, PNS, IMDCT,
  and channel coupling.
- HE-AAC SBR/PS profiles.
- LATM/LOAS transport.
- ADTS frames carrying multiple raw data blocks.
