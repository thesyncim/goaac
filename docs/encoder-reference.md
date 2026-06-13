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
| `libAACenc/src/block_switch.cpp:FDKaacEnc_InitBlockSwitching`, `FDKaacEnc_BlockSwitching`, `FDKaacEnc_SyncBlockSwitching` | `internal/fdkaac/block_switch.go:FDKaacEncInitBlockSwitching`, `FDKaacEncBlockSwitching`, `FDKaacEncSyncBlockSwitching` | Source-derived attack-detection vectors for AAC-LC mono transitions, LFE and low-delay controls, stereo common-window grouping sync vectors, state hashes for energies/IIR/grouping, unsupported-state tests, and zero-allocation guard under the `internal/fdkaac` wasm test binary. |
| `libFDK/include/fixpoint_math.h:fLog2`, `CalcLdData`; `libFDK/src/fixpoint_math.cpp:LdDataVector` | `internal/fdkaac/fixlog.go` | Source-derived log-data vectors for zero/negative and positive fixed-point energies, vector/in-place tests, invalid vector transition tests, and zero-allocation guard under the `internal/fdkaac` wasm test binary. |
| `libFDK/include/fixpoint_math.h:CalcInvLdData`; `libFDK/src/fixpoint_math.cpp:exp2_tab_long`, `exp2w_tab_long`, `exp2x_tab_long` | `internal/fdkaac/fixlog.go` | Source-derived inverse-LD64 exponent vectors covering underflow, saturation, negative fractional, zero, and positive cases, ROM edge checks, and zero-allocation guard under the `internal/fdkaac` wasm test binary. |
| `libFDK/include/fixpoint_math.h:invSqrtNorm2`, `sqrtFixp`; `libFDK/src/FDK_tools_rom.cpp:invSqrtTab` | `internal/fdkaac/fixsqrt.go` | Source-derived fixed-point inverse-square-root and square-root vectors from the generic FDK table/interpolation path, invalid input tests, and zero-allocation guard under the `internal/fdkaac` wasm test binary. |
| `libAACenc/src/band_nrg.cpp:FDKaacEnc_CalcSfbMaxScaleSpec`, `FDKaacEnc_CheckBandEnergyOptim`, `FDKaacEnc_CalcBandEnergyOptimLong`, `FDKaacEnc_CalcBandEnergyOptimShort`, `FDKaacEnc_CalcBandNrgMSOpt` | `internal/fdkaac/band_nrg.go` | Source-derived scalefactor-band max-scale, check-energy, long-window energy/log-data, short-window energy, and mid/side energy/log-data hashes, invalid offset/scale/output transition tests, and zero-allocation guard under the `internal/fdkaac` wasm test binary. |
| `libAACenc/src/bit_cnt.h:FDKaacEnc_bitCountScalefactorDelta`; `libAACenc/src/aacEnc_rom.cpp:FDKaacEnc_huff_ltabscf` | `internal/fdkaac/bit_cnt.go` | Source-derived scalefactor Huffman length table checksum, selected delta vectors across `[-60, 60]`, invalid delta transition tests, and zero-allocation guard under the `internal/fdkaac` wasm test binary. |
| `libAACenc/src/grp_data.cpp:FDKaacEnc_groupShortData` | `internal/fdkaac/group_short.go` | Source-derived grouped-offset, min-SNR, saturating threshold/energy/MS/spread-energy, max-SFB, zero-spectrum floor, and spectrum-regroup vectors, invalid transition tests, and zero-allocation guard under the `internal/fdkaac` wasm test binary. |
| `libAACenc/src/spreading.cpp:FDKaacEnc_SpreadingMax` | `internal/fdkaac/spreading.go` | Source-derived bidirectional scalefactor-band spreading vectors, invalid input transition tests, and zero-allocation guard under the `internal/fdkaac` wasm test binary. |
| `libAACenc/src/pre_echo_control.cpp:FDKaacEnc_InitPreEchoControl`, `FDKaacEnc_PreEchoControl` | `internal/fdkaac/pre_echo.go` | Source-derived pre-echo initialization, disabled-control, current-scale-larger, and previous-scale-larger vectors, invalid state/shift tests, and zero-allocation guard under the `internal/fdkaac` wasm test binary. |
| `libAACenc/src/psy_main.cpp` "Advance thresholds" block using `THR_SHIFTBITS`, `PCM_QUANT_THR_SCALE`, `FDKaacEnc_SpreadingMax`, and `FDKaacEnc_PreEchoControl` | `internal/fdkaac/threshold.go` | Source-derived long/short threshold-advance vectors covering clip-energy scaling, threshold spreading, PCM quantization floors, start/stop pre-echo history transitions, spread-energy preparation, invalid transition tests, and zero-allocation guard under the `internal/fdkaac` wasm test binary. |
| `libAACenc/src/ms_stereo.cpp:FDKaacEnc_MsStereoProcessing` | `internal/fdkaac/ms_stereo.go` | Source-derived MS-none, MS-some with intensity-stereo exclusions, and MS-all promotion vectors covering mask/digest decisions, spectrum mid/side rewrites, energy/threshold/LD/spread updates, invalid transition tests, and zero-allocation guard under the `internal/fdkaac` wasm test binary. |
| `libAACenc/src/psy_main.cpp` "build output" block | `internal/fdkaac/psy_output.go` | Source-derived long and short output-handoff vectors covering `sfbCnt`, `sfbPerGroup`, `maxSfbPerGroup`, window sequence/shape, grouping-mask generation, group-length copy, grouped energy/spread-energy copies, invalid transition tests, and zero-allocation guard under the `internal/fdkaac` wasm test binary. |
| `libAACenc/src/sf_estim.cpp:FDKaacEnc_CalcFormFactor` | `internal/fdkaac/sf_estim.go:FDKaacEncCalcFormFactor` | Source-derived long and grouped-short form-factor LD-data vectors, inactive-band `-1.0` handling, malformed channel/output/offset/spectrum transition tests, and zero-allocation guard under the `internal/fdkaac` wasm test binary. |
| `libAACenc/src/sf_estim.cpp:FDKaacEnc_calcSfbRelevantLines` | `internal/fdkaac/sf_estim.go:FDKaacEncCalcSfbRelevantLines` | Source-derived long and grouped-short relevant-line vectors covering output clearing, active-band energy/threshold gating, inactive-band zeros, malformed offset/data transition tests, and zero-allocation guard under the `internal/fdkaac` wasm test binary. |
| `libAACenc/src/sf_estim.cpp:FDKaacEnc_countSingleScfBits`, `FDKaacEnc_calcSingleSpecPe` | `internal/fdkaac/sf_estim.go:FDKaacEncCountSingleScfBits`, `FDKaacEncCalcSingleSpecPe` | Source-derived single-band scalefactor-bit and spectral-PE vectors covering both PE branches, out-of-range delta transitions through the Huffman length helper, and zero-allocation guard under the `internal/fdkaac` wasm test binary. |
| `libAACenc/src/sf_estim.cpp:FDKaacEnc_countScfBitsDiff`, `FDKaacEnc_calcSpecPeDiff` | `internal/fdkaac/sf_estim.go:FDKaacEncCountScfBitsDiff`, `FDKaacEncCalcSpecPeDiff` | Source-derived multi-band scalefactor-bit and spectral-PE delta vectors covering previous/next relevant-band linking, skipped `FDK_INT_MIN` bands, lazy `sfbConstPePart` cache fill, high/low PE branches, invalid transition tests, and zero-allocation guard under the `internal/fdkaac` wasm test binary. |
| `libAACenc/src/sf_estim.cpp:FDKaacEnc_improveScf` | `internal/fdkaac/sf_estim.go:FDKaacEncImproveScf` | Source-derived scale-factor improvement vectors covering accepted upward refinement, accepted downward refinement, minimum-scale stop, dead-zone quantization, quantized-spectrum side effects, invalid transition tests, and zero-allocation guard under the `internal/fdkaac` wasm test binary. |
| `libAACenc/src/sf_estim.cpp:FDKaacEnc_assimilateSingleScf` | `internal/fdkaac/sf_estim.go:FDKaacEncAssimilateSingleScf` | Source-derived single-band assimilation vectors covering neighbouring-scale constraints, restart-on-success rescans, already-checked minimum-scale controls, no-op flat-scale control, quantized-spectrum side effects, invalid transition tests, and zero-allocation guard under the `internal/fdkaac` wasm test binary. |
| `libAACenc/src/sf_estim.cpp:FDKaacEnc_assimilateMultipleScf` | `internal/fdkaac/sf_estim.go:FDKaacEncAssimilateMultipleScf` | Source-derived multi-band assimilation vectors covering regional shared-scale lowering, threshold rejection, minimum-scale rejection, irrelevant-band no-op control, quantized/temp-spectrum side effects, invalid transition tests, and zero-allocation guard under the `internal/fdkaac` wasm test binary. |
| `libAACenc/src/sf_estim.cpp:FDKaacEnc_FDKaacEnc_assimilateMultipleScf2` | `internal/fdkaac/sf_estim.go:FDKaacEncAssimilateMultipleScf2` | Source-derived second-pass multi-band assimilation vectors covering coarser regional refinement, finer regional refinement, no-quant scale lowering, threshold rejection, irrelevant-band no-op control, invalid transition tests, and zero-allocation guard under the `internal/fdkaac` wasm test binary. |
| `libFDK/src/fixpoint_math.cpp:CalcLdInt`, `libFDK/include/fixpoint_math.h:fMultI`; `libAACenc/src/line_pe.cpp:FDKaacEnc_prepareSfbPe`, `FDKaacEnc_calcSfbPe` | `internal/fdkaac/fixlog.go:CalcLdInt`, `internal/fdkaac/fixpoint.go:FMultI`, `internal/fdkaac/line_pe.go` | Source-derived integer-LD and fractional-times-int vectors, line-PE channel vectors covering active-line preparation, high/low PE branches, inactive/intensity-book transitions, malformed input tests, and zero-allocation guard under the `internal/fdkaac` wasm test binary. |
| `libAACenc/src/aacEnc_rom.cpp` quantizer ROM tables; `libAACenc/src/quantize.cpp:FDKaacEnc_quantizeLines`, `FDKaacEnc_invQuantizeLines`, `FDKaacEnc_QuantizeSpectrum`, `FDKaacEnc_calcSfbDist`, `FDKaacEnc_calcSfbQuantEnergyAndDist` | `internal/fdkaac/quant_tables.go`, `internal/fdkaac/quantize.go` | Mechanically derived quantizer ROM checksums, source-derived quantize/inverse-quantize vectors, grouped spectrum quantization vectors, SFB distortion and quantized energy/distortion vectors, invalid transition tests, and zero-allocation guard under the `internal/fdkaac` wasm test binary. |
| `libMpegTPEnc/src/tpenc_asc.cpp:transportEnc_writeASC`, `writeAot`, `writeSampleRate`, AAC-LC `transportEnc_writeGASpecificConfig`; `libMpegTPEnc/src/tpenc_adts.cpp:adtsWrite_Init`, `adtsWrite_EncodeHeader` | `internal/fdkaac/transport.go` | Native FDK v2.0.3 byte-vector probe for AAC-LC ASC and CRC-less ADTS headers, unsupported-branch tests, and scratch-based zero-allocation guard. |
