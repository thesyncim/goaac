# AAC-LC Encoder Reference

## Upstream Pin

- Project: Fraunhofer FDK AAC Codec Library for Android
- Repository: `https://github.com/mstorsjo/fdk-aac`
- Release tag: `v2.0.3`
- Commit: `716f4394641d53f0d79c9ddac3fa93b03a49f278`
- Local checkout: `third_party/fdk-aac`
- Reference CLI: `aac-enc`

FDK-AAC is the encoder source truth for the Go AAC-LC encoder track. The
initial parity target is AAC-LC (`AOT_AAC_LC`, value `2`) with ADTS and raw
access-unit output. HE-AAC/SBR/PS, AAC-LD, AAC-ELD, USAC, and surround modes are
out of scope until AAC-LC stereo/mono parity is proven.

The FDK source is C++ and is not linked or loaded by the Go runtime. It is kept
as a pinned oracle and translation source.

## Oracle Build

Build the pinned native encoder oracle:

```sh
./scripts/build_fdkaac_oracle.sh
```

The script builds `third_party/fdk-aac/aac-enc` into `dist/fdk-aac-oracle` with
Release flags, static library output, and `BUILD_PROGRAMS=ON`.

Use the oracle for AAC-LC ADTS output:

```sh
dist/fdk-aac-oracle/aac-enc -t 2 -a 1 -r 96000 input.wav output.aac
```

The expected FDK encoder initialization sequence, from
`libAACenc/include/aacenc_lib.h` and `aac-enc.c`, is:

- `aacEncOpen`
- `aacEncoder_SetParam(AACENC_AOT, AOT_AAC_LC)`
- `aacEncoder_SetParam(AACENC_SAMPLERATE, sampleRate)`
- `aacEncoder_SetParam(AACENC_CHANNELMODE, mode)`
- `aacEncoder_SetParam(AACENC_CHANNELORDER, 1)`
- `aacEncoder_SetParam(AACENC_BITRATE, bitrate)` or `AACENC_BITRATEMODE`
- `aacEncoder_SetParam(AACENC_TRANSMUX, TT_MP4_ADTS)` for ADTS fixtures
- `aacEncoder_SetParam(AACENC_AFTERBURNER, 1)`
- `aacEncEncode(... nil ...)` initialization
- `aacEncInfo`
- repeated `aacEncEncode`
- `aacEncClose`

## Go Parity Plan

The encoder port must be source-shaped first and optimized second:

- `libFDK`: fixed-point primitives, bit buffers, MDCT/FFT, math tables.
- `libMpegTPEnc`: ASC, ADTS, and raw access-unit transport emission.
- `libAACenc`: block switching, psychoacoustic model, TNS, MS stereo, bit
  counting, quantization, rate control, and bitstream coding.
- Public facade: append-style `EncodeInto` APIs for S16 PCM, raw access units,
  ADTS, and RTMP/FLV AAC packet payloads.

Parity is not complete until pinned PCM fixtures encoded by the Go encoder match
the FDK oracle at the selected semantic boundary. For the strict source-shaped
path that means byte-identical ADTS/raw AAC output for fixed inputs and encoder
settings. Decode-roundtrip PCM hashes are an additional quality guard, not a
substitute for encoder parity.

## License Note

The FDK-AAC software license grants copyright redistribution rights but does not
grant patent rights. Keep the full upstream notice in source distributions and
do not claim patent coverage from this repository.
