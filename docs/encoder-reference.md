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

## Ported Slices

| Upstream path/function | Local path/function | Proof |
| --- | --- | --- |
| `libFDK/include/FDK_bitstream.h:FDKwriteBits`, `FDKwriteEscapedValue`, `FDKsyncCache`, `FDKbyteAlign`, `FDKfetchBuffer`; `libFDK/src/FDK_bitbuffer.cpp:FDK_put`, `FDK_Fetch` | `internal/fdkaac/bitstream.go` | Unit vectors checked against a native FDK v2.0.3 probe, ring-buffer wrap tests, partial-byte fetch tests, and steady-state allocation guard. |
| `libFDK/include/common_fix.h`, `fixmul.h`, `clz.h`, `scale.h` fixed-point constants, multiply helpers, headroom helpers, and saturating shifts | `internal/fdkaac/fixpoint.go` | Native FDK v2.0.3 arm64 vector probe for signed multiply variants, pow2 helpers, `fixnorm*`, `scaleValue`, `scaleValueSaturate`, normal and alt saturating shifts, plus zero-allocation guard. |
| `libFDK/include/fixmadd.h`, `libFDK/include/arm/fixmadd_arm.h` multiply-add/subtract helpers and pow2-add helpers | `internal/fdkaac/fixmadd.go` | Native FDK v2.0.3 arm64 vector probe for `fixmadd*`, `fixmsub*`, bit-exact div2 helpers, `fixpadd*`, plus zero-allocation guard. |
| `libFDK/src/scale.cpp`, `libFDK/include/scale.h`, `libFDK/src/arm/scale_arm.cpp` scale-vector, saturated scale-vector, factor-scale, and scalefactor helpers | `internal/fdkaac/scale.go` | Native FDK v2.0.3 arm64 vector probe for in-place and dst/src DBL/SGL/PCM scaling, saturated DBL/SGL scaling, factor scaling, scalefactor scans, plus zero-allocation guard. |
| `libFDK/include/cplx_mul.h`, `libFDK/include/arm/cplx_mul_arm.h` complex fixed-point multiply overloads | `internal/fdkaac/cplxmul.go` | Native FDK v2.0.3 arm64 vector probe for `cplxMultDiv2`, `cplxMultSubDiv2`, and `cplxMult` DBL/SGL overloads, including the arm64 combined-product div2 path and zero-allocation guard. |
| `libFDK/include/fft.h:fft_4`, `fft_8` fixed short FFT kernels | `internal/fdkaac/fft.go` | Native FDK v2.0.3 arm64 vector probe for in-place length-4 and length-8 complex FFT outputs, plus zero-allocation guard. |
| `libFDK/include/scramble.h:scramble`, `libFDK/src/arm/fft_rad2_arm.cpp:dit_fft` radix-2 FFT path | `internal/fdkaac/fft.go` | Source-derived vectors from the pinned arm64 `SINETABLE_16BIT` implementation for bit reversal and length-8/16 radix loops, exercised by the `internal/fdkaac` wasm test binary plus compile/static checks. Fresh native execution was blocked by the local Mach-O loader before `main`. |
| `libFDK/src/FDK_tools_rom.cpp:SineTable512`; `libFDK/src/fft.cpp` 64/128/256/512-point `dit_fft` call sites | `internal/fdkaac/rom.go:SineTable512`, `internal/fdkaac/fft.go:DITFFT512` | Source-derived `SINETABLE_16BIT` ROM checksum and sample-index tests, source-derived length 64/128/256/512 radix FFT hashes, unsupported-length transition test, and zero-allocation guard under the `internal/fdkaac` wasm test binary. |
| `libFDK/src/fft.cpp:fft2`, selected `fft` dispatcher branches and scale-factor constants | `internal/fdkaac/fft.go:FFT2`, `FFT` | Source-derived length 2 vectors, exact dispatcher vectors for 2/4/8, source-derived radix hashes for 64/128/256/512, scale-factor transition tests, unsupported-length tests, and zero-allocation guard. |
| `libFDK/src/FDK_tools_rom.cpp:SineTable1024`; `libFDK/src/dct.cpp:dct_getTables` radix-2 sine twiddle selection | `internal/fdkaac/rom.go:SineTable1024` | Source-derived `SINETABLE_16BIT` ROM checksum and sample-index tests, relation test proving `SineTable512` equals every second 1024 entry, and zero-allocation guard. |
| `libFDK/src/FDK_tools_rom.cpp` AAC-LC sine/KBD window slope tables; `FDKgetWindowSlope` AAC-LC length selection | `internal/fdkaac/window.go` | Source-derived `WINDOWTABLE_16BIT` checksums and sample-index tests for 1024/960 and 128/120 sine/KBD tables, selector transition tests including `shape & 1`, unsupported-length tests, and zero-allocation guard. |
| `libFDK/src/dct.cpp:dct_getTables` radix-2 table selection, `dct_IV` | `internal/fdkaac/dct.go:dctGetTables`, `DCTIV` | Source-derived radix-2 DCT-IV output hashes for AAC-LC lengths 128 and 1024, table-selection tests, scale-factor transition tests, unsupported-length tests, and zero-allocation guard under the `internal/fdkaac` wasm test binary. |
| `libFDK/src/mdct.cpp:mdct_block`; `libAACenc/src/transform.cpp:FDKaacEnc_Transform_Real` | `internal/fdkaac/mdct.go:MDCTBlock`, `FDKaacEncTransformReal` | Source-derived MDCT output hashes for AAC-LC long and short windows, long/start/short/stop transition hashes with sine/KBD state changes, scale-factor tests, unsupported mixed-radix controls, and zero-allocation guard under the `internal/fdkaac` wasm test binary. |
| `libMpegTPEnc/src/tpenc_asc.cpp:transportEnc_writeASC`, `writeAot`, `writeSampleRate`, AAC-LC `transportEnc_writeGASpecificConfig`; `libMpegTPEnc/src/tpenc_adts.cpp:adtsWrite_Init`, `adtsWrite_EncodeHeader` | `internal/fdkaac/transport.go` | Native FDK v2.0.3 byte-vector probe for AAC-LC ASC and CRC-less ADTS headers, unsupported-branch tests, and scratch-based zero-allocation guard. |
