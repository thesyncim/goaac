# AAC-LC Test Vectors

The committed vector suite lives in `testdata/aaclc`.

Each `.aac` file is an ADTS AAC-LC elementary stream. `manifest.json` records:

- input byte length and SHA-256
- decoded PCM byte length and SHA-256
- sample rate, channel count, and ADTS frame count
- the oracle contract used to produce the PCM hashes

The PCM hashes are for little-endian signed 16-bit interleaved samples produced
by pinned FAAD2 `2.11.2`, built with:

```sh
-O2 -ffp-contract=off -DLC_ONLY_DECODER -DDISABLE_SBR
```

`-ffp-contract=off` is intentional. The Go port evaluates the translated FAAD2
float path as explicit `float32` operations, so the C oracle disables fused
floating-point contraction to match that source-level contract.

## What The Tests Prove

The committed vector tests verify every vector through:

- ADTS frame parsing and input SHA-256 checks
- complete-stream `DecodeADTS`
- streaming `DecodeADTSReader`
- reusable-buffer `DecodeADTSInto`
- frame-by-frame `Decoder.Decode`
- raw AAC payload decode with `TransportRaw`
- RTMP/FLV AAC sequence-header plus raw-packet decode with `FLVAACDecoder`
- zero-allocation `ParseFLVAudioTag` inspection
- byte-exact PCM SHA-256 comparison against the FAAD2 oracle output

The RTMP/FLV checks synthesize AAC sequence-header and raw AAC audio-message
payloads from the committed ADTS frames. That proves the production streaming
shape used by RTMP ingest produces the same PCM hashes as the ADTS and raw
access-unit APIs.

The live integration test still synthesizes a fresh AAC-LC fixture with FFmpeg
and builds the FAAD2 oracle locally when `ffmpeg` and a C compiler are available.

Encoder bring-up also has primitive oracle vectors under `internal/fdkaac`.
Those vectors are checked against pinned FDK-AAC v2.0.3 probes for writer-side
bitstream emission, AAC-LC transport headers, fixed-point arithmetic, and
multiply-add, scale-vector, complex multiply, and short FFT primitives. They
prove exact low-level behavior before the full AAC-LC encoder core is exposed.
The radix-2 FFT bring-up additionally has source-derived vectors for
`scramble` and the arm64 `SINETABLE_16BIT` `dit_fft` loop. Those vectors are
exercised by the `internal/fdkaac` wasm test binary and still need a fresh
native probe once local execution of newly built Mach-O binaries is available
again.

The FDK encoder bandwidth/channel-control subset is covered with source-shaped
vectors for `FDKaacEnc_DetermineBandWidth`, `FDKaacEnc_GetVBRBitrate`,
`FDKaacEnc_AdjustVBRBitrateMode`, `FDKaacEnc_DetermineEncoderMode`, and
`FDKaacEnc_InitChannelMapping`, and `FDKaacEnc_InitElementBits`. Tests verify
CBR/VBR table selection, proposed-bandwidth clipping, low-delay fixed-point
interpolation, effective channel handling with LFE excluded, channel-mode
selection, MPEG/WAV/WG4 channel order, relative-bit shares, 7.1 aliases,
per-element bitrate, max-bit budgets, LFE reservoir exclusion, malformed
controls, no-cgo compilation, and zero steady-state allocation.

The FDK `aacenc.cpp` encoder config and QC-budget setup subset is covered with
source-shaped vectors for default config initialization, frame-bit and bitrate
math, low/high bitrate limiting, ancillary-rate validation, CBR and VBR
multi-element QC-budget construction, transport bit-reservoir signaling, and
the handoff from prepared config into `FDKaacEnc_QCInit`. Tests also verify
malformed controls, no-cgo compilation, and zero steady-state allocation.

The FDK `SineTable512` ROM used by the 64/128/256/512-point radix FFT path is
also locked by source-derived sample entries, an FNV-1a checksum, output hashes
for all four supported lengths, an unsupported-length transition test, and a
zero-allocation guard.

The FDK `fft` dispatcher subset is covered for lengths 2, 4, 8, 64, 128, 256,
and 512. Tests verify exact short-kernel outputs, radix FFT hashes,
scale-factor increments, unsupported-length transitions, and zero steady-state
allocation.

The FDK `SineTable1024` ROM used by radix-2 DCT-IV sine twiddle selection is
locked by source-derived sample entries, an FNV-1a checksum, a relation check
against `SineTable512`, and a zero-allocation guard.

The AAC-LC window slope ROM subset is covered for 1024/960 and 128/120
sine/KBD windows. Tests verify source-derived checksums, sample entries, FDK
selector transitions including `shape & 1`, unsupported-length rejection, and
zero steady-state allocation.

The FDK radix-2 DCT-IV subset is covered for AAC-LC lengths 128 and 1024.
Tests verify source-derived output hashes, table selection for sine window
twiddles and the 1024-entry sine table, exact scale-factor increments,
unsupported-length rejection for the missing mixed-radix paths, and zero
steady-state allocation.

The FDK AAC-LC MDCT transform subset is covered for radix-2 long and short
transforms. Tests verify source-derived output hashes for long and short MDCT
blocks, long/start/short/stop transition hashes with sine/KBD window-state
changes, exact scale-factor increments, unsupported mixed-radix rejection, and
zero steady-state allocation.

The FDK AAC-LC block switching subset is covered with source-derived state
hashes for attack/no-attack frame transitions, IIR-filtered and unfiltered
window energies, LFE and low-delay controls, stereo common-window grouping
sync, unsupported-state rejection, and zero steady-state allocation.

The FDK log-data subset is covered with source-derived vectors for zero,
negative, and positive fixed-point energies through `CalcLdData` and
`LdDataVector`, including in-place vector use, invalid vector transitions, and
zero steady-state allocation.

The FDK inverse-LD subset is covered with source-derived vectors for
`CalcInvLdData`, including the three exponent lookup tables, interpolation,
underflow, saturation, negative fractional inputs, zero, positive inputs, and
zero steady-state allocation.

The FDK square-root subset is covered with source-derived vectors for the
generic fixed-point `invSqrtNorm2` table/interpolation path and `sqrtFixp`.
Tests also verify invalid negative/nil transitions and zero steady-state
allocation.

The FDK AAC-LC band-energy subset is covered with source-derived hashes for
scalefactor-band max-scale scans, pre-shift energy checks, long-window
energy/log-data scaling, short-window band energies, and mid/side band energies
with log-data output enabled. Tests also verify invalid offset/scale/output
transitions and zero steady-state allocation.

The FDK scalefactor delta bit-count subset is covered with a source-derived
Huffman length table hash, selected deltas across the legal `[-60, 60]` range,
invalid delta transitions, and zero steady-state allocation.

The FDK spectral Huffman bit-count subset is covered with mechanically derived
ROM checksums for the spectral length tables and source-derived vectors for
zero spectra, LAV category transitions, escape lengths, short-band zero-tail
counts, fixed-codebook recounts, invalid controls, and zero steady-state
allocation.

The FDK Huffman codeword emission subset is covered with mechanically derived
codeword ROM checksums and source-derived byte vectors for spectral codebooks
0-11, pair-width codebooks, escape extensions, scalefactor deltas, invalid
controls, and zero steady-state allocation.

The FDK dynamic noiseless bit-count subset is covered with source-derived
long-window, ER-VCB11, short-grouped, PNS, and intensity vectors. Tests verify
greedy section merges, section layout hashes, Huffman/side-info/scalefactor/noise
totals, malformed controls, and zero steady-state allocation.

The FDK channel bitstream writer helper subset is covered with source-derived
byte vectors for global gain, long and short ICS info, section data,
scalefactor/noise data, spectral payload emission, MS mask signalling, TNS
present flags, long and short TNS payload data, and one-bit pulse/gain-control
placeholders. Extension writer vectors cover raw extension payload syntax, GA
fill-element wrapping, data-stream elements, ER en-bloc payloads, and ELD SBR
direct payloads. Tests also verify a combined long-channel payload sequence,
invalid writer/control transitions, and zero steady-state allocation.

The FDK AAC-LC channel element writer subset is covered with source-derived
sequence-table pins for SCE and CPE normal syntax, plus byte vectors for a mono
SCE and a common-window stereo CPE assembled through the lower-level channel
payload writers. Tests also verify minimum static-count vectors for SCE TNS
suppression and CPE MS suppression, malformed bit-count error returns, invalid
element/control transitions, and zero steady-state allocation.

The FDK AAC-LC raw AU bitstream writer subset is covered with source-shaped
vectors for `FDKaacEnc_WriteBitstream` and `FDKaacEnc_ByteAlignment`. Tests
verify top-level SCE raw payload bytes, element-associated extension payload
ordering, global extension payload ordering, the synthetic fill-data entry,
`ID_END` emission, explicit alignment bits, written-bit mismatch errors,
malformed controls, no-cgo compilation, and zero steady-state allocation.

The FDK AAC-LC short grouping subset is covered with source-derived grouped
offset and min-SNR vectors, saturating threshold/energy/MS/spread-energy sums,
max-SFB scan and zero-spectrum floor checks, active and inactive-band
spectrum-regroup hashes, invalid grouping/control transitions, and zero
steady-state allocation.

The FDK AAC-LC threshold spreading subset is covered with a source-derived
bidirectional spreading vector over scalefactor-band energy/threshold data,
invalid input transitions, and zero steady-state allocation.

The FDK AAC-LC pre-echo control subset is covered with source-derived vectors
for initialization, disabled-control copy-through, current-scale-larger
limiting, and previous-scale-larger limiting, plus invalid state/shift
transitions and zero steady-state allocation.

The FDK AAC-LC threshold advance subset is covered with source-derived
long-window and short-window vectors for clip-energy scaling, threshold
spreading, PCM quantization floors, FDK's per-window `energyShift` mutation,
start/stop pre-echo history transitions, spread-energy preparation, invalid
control transitions, and zero steady-state allocation.

The FDK AAC-LC MS stereo subset is covered with source-derived MS-none,
MS-some, and MS-all promotion vectors. Tests verify per-band mask and digest
decisions, intensity-stereo exclusion handling, mid/side spectrum rewrites,
energy/threshold/LD/spread updates, invalid controls, and zero steady-state
allocation.

The FDK AAC-LC psychoacoustic output handoff subset is covered with
source-derived long and short vectors for output field selection, grouping-mask
generation, group-length copying, grouped energy/spread-energy copying, invalid
state transitions, and zero steady-state allocation.

The FDK AAC-LC form-factor subset is covered with source-derived long and
grouped-short vectors for `FDKaacEnc_CalcFormFactor`, including inactive-band
`-1.0` sentinels, malformed channel/output/offset/spectrum transitions, and
zero steady-state allocation.

The FDK AAC-LC relevant-lines subset is covered with source-derived long and
grouped-short vectors for `FDKaacEnc_calcSfbRelevantLines`, including
output-clearing behavior, active-band energy/threshold gating, inactive-band
zeros, malformed offset/data transitions, and zero steady-state allocation.

The FDK AAC-LC single-band PE subset is covered with source-derived vectors for
`FDKaacEnc_countSingleScfBits` and `FDKaacEnc_calcSingleSpecPe`, including both
spectral-PE branches, out-of-range delta transitions through the scalefactor
Huffman length helper, and zero steady-state allocation.

The FDK AAC-LC multi-band PE diff subset is covered with source-derived vectors
for `FDKaacEnc_countScfBitsDiff` and `FDKaacEnc_calcSpecPeDiff`, including
previous/next relevant-band linking, skipped `FDK_INT_MIN` bands, lazy
`sfbConstPePart` cache-fill mutation, invalid transitions, and zero
steady-state allocation.

The FDK AAC-LC improve-SCF subset is covered with source-derived vectors for
`FDKaacEnc_improveScf`. Tests verify accepted upward refinement, accepted
downward refinement, minimum-scale stopping, dead-zone quantization,
quantized-spectrum side effects, malformed controls, and zero steady-state
allocation.

The FDK AAC-LC single-SCF assimilation subset is covered with source-derived
vectors for `FDKaacEnc_assimilateSingleScf`. Tests verify neighbouring-scale
constraints, restart-on-success rescans, already-checked minimum-scale controls,
no-op flat-scale controls, quantized-spectrum side effects, malformed controls,
and zero steady-state allocation.

The FDK AAC-LC multi-SCF assimilation subset is covered with source-derived
vectors for `FDKaacEnc_assimilateMultipleScf`. Tests verify regional shared-scale
lowering, threshold rejection, minimum-scale rejection, irrelevant-band no-op
control, quantized and temporary-spectrum side effects, malformed controls, and
zero steady-state allocation.

The FDK AAC-LC multi-SCF assimilation-2 subset is covered with source-derived
vectors for `FDKaacEnc_FDKaacEnc_assimilateMultipleScf2`. Tests verify coarser
regional refinement, finer regional refinement, no-quant scale lowering,
threshold rejection, irrelevant-band no-op control, malformed controls, and zero
steady-state allocation.

The FDK AAC-LC channel scale-factor estimation subset is covered with
source-derived vectors for `FDKaacEnc_EstimateScaleFactorsChannel`. Tests verify
initial scale-factor estimation, max-quant lower bounds, inverse-quant
refinement, single/multi-SCF orchestration, dead-zone quantization, all-zero
spectrum clearing, grouped inactive-band transitions, malformed controls, and
zero steady-state allocation.

The FDK AAC-LC multi-channel scale-factor wrapper is covered with
source-derived vectors for `FDKaacEnc_EstimateScaleFactors` and the owned
`QC_OUT_CHANNEL` scale-factor fields. Tests verify channel-owned `quantSpec`,
`scf`, and `globalGain` mutation, active/all-zero channel iteration, shared
caller-owned scratch reuse, malformed controls, and zero steady-state
allocation.

The FDK AAC-LC line-PE subset is covered with source-derived vectors for
`CalcLdInt`, `fMultI`, `FDKaacEnc_prepareSfbPe`, and `FDKaacEnc_calcSfbPe`.
Tests verify active-line preparation, high/low PE branches, inactive and
intensity-book transitions, malformed controls, and zero steady-state
allocation.

The FDK AAC-LC min-SNR adaptation subset is covered with source-derived
vectors for `FDKaacEnc_calcThreshExp` and `FDKaacEnc_adaptMinSnr`. Tests verify
threshold exponent scratch output, default `MINSNR_ADAPT_PARAM` constants,
low-energy min-SNR reduction, zero-energy no-op behavior, malformed controls,
and zero steady-state allocation.

The FDK AAC-LC avoid-hole initialization subset is covered with source-derived
vectors for `FDKaacEnc_initAvoidHoleFlag`. Tests verify long-window spread
attenuation and peak/valley min-SNR edits, short-window spread scaling with
inactive grouped bands preserved, stereo MS min-SNR/spread coupling, avoid-hole
flag states, malformed controls, no-cgo compilation, and zero steady-state
allocation.

The FDK AAC-LC CBR threshold-reduction subset is covered with source-derived
vectors for `FDKaacEnc_FDKaacEnc_calcPeNoAH` and
`FDKaacEnc_reduceThresholdsCBR`. Tests verify scaled no-AH PE summation,
AH-active skips, NO_AH bypass, AH activation and min-SNR clamping, 29 dB
minimum-ratio limiting, zero-reduction no-op behavior, malformed controls,
no-cgo compilation, and zero steady-state allocation.

The FDK AAC-LC VBR threshold-reduction subset is covered with source-derived
`invInt` and `invSqrt4` table hashes, a `FDKaacEnc_calcChaosMeasure` vector,
and long/short-window vectors for `FDKaacEnc_reduceThresholdsVBR`. Tests verify
chaos smoothing/history, no-active chaos fallback, AH-active skips, AH
activation and min-SNR clamping, short-window group scaling, malformed
controls, no-cgo compilation, and zero steady-state allocation.

The FDK AAC-LC correct-threshold subset is covered with source-shaped positive
and negative PE-delta vectors for `FDKaacEnc_correctThresh`. Tests verify
threshold correction, PE-factor scratch, active-line LD scratch, AH-inactive
clamp controls, malformed prepared-state controls, no-cgo compilation, and zero
steady-state allocation.

The FDK AAC-LC min-SNR PE-reduction subset is covered with a source-shaped
vector for `FDKaacEnc_reduceMinSnr`. Tests verify descending high-SFB traversal,
min-SNR relaxation to 1 dB, threshold replacement, PE-band replacement,
global/channel PE updates, no-op behavior, malformed controls, no-cgo
compilation, and zero steady-state allocation.

The FDK AAC-LC allow-more-holes subset is covered with source-shaped M/S and
energy-fallback vectors for `FDKaacEnc_allowMoreHoles`. Tests verify
inactive-band PE clearing, lower-energy side-channel hole opening, low-energy
high-band erasure, threshold replacement, no-op behavior, malformed controls,
no-cgo compilation, and zero steady-state allocation.

The FDK AAC-LC AH reset subset is covered with a source-shaped vector for
`FDKaacEnc_resetAHFlags`. Tests verify `AH_ACTIVE` to `AH_INACTIVE` transitions
inside active SFB ranges, inactive-band preservation, malformed controls,
no-cgo compilation, and zero steady-state allocation.

The FDK AAC-LC PE threshold-adjustment subset is covered with source-derived
vectors for normalized Schur division, `FDKaacEnc_preparePe`,
`FDKaacEnc_calcWeighting`, `FDKaacEnc_calcPe`, and `FDKaacEnc_peCalculation`.
Tests verify long-window weighting-patch execution, weighted energy/threshold
mutation, short-window patch-state transition, malformed controls, and zero
steady-state allocation.

The FDK AAC-LC threshold-adjustment wrapper subset is covered with CBR and VBR
vectors for `FDKaacEnc_AdjustThresholds`. Tests compare the wrapper branch
wiring against the already-ported inner CBR/VBR threshold adapters, verify
post-adjustment threshold unweighting, malformed mapping/control rejection, a
two-element global inter-CBR reduction vector, no-cgo compilation, and zero
steady-state allocation.

The FDK AAC-LC QC prepare wrapper subset is covered with a source-shaped SCE
vector for `FDKaacEnc_QCMainPrepare`. Tests compare the wrapper against the
direct form-factor, PE-calculation, and static channel-element dry-run sequence,
verify the nil-QC static side-info bit count, malformed controls, no-cgo
compilation, and zero steady-state allocation.

The FDK AAC-LC QC lifecycle subset is covered with source-shaped vectors for
`FDKaacEnc_QCNew`, `FDKaacEnc_QCOutNew`, and `FDKaacEnc_QCOutInit`. Tests verify
fresh fixed-state pointer linking, 5.1 WAV channel-to-element traversal,
handoff into `FDKaacEnc_QCInit`, malformed controls, no-cgo compilation, and
zero steady-state allocation.

The FDK AAC-LC QC initialization subset is covered with source-shaped vectors
for `FDKaacEnc_QCInit`. Tests verify CBR kernel budget/reservoir transfer, VBR
multi-element relative-bit traversal, VBR quality-factor selection, FF reservoir
preservation and reset controls, element-bit initialization, threshold-state
seeding, malformed controls, no-cgo compilation, and zero steady-state
allocation.

The FDK AAC-LC frame bitrate-padding subset is covered with source-shaped
vectors for `FDKaacEnc_calcFrameLen`, `FDKaacEnc_framePadding`, and
`FDKaacEnc_AdjustBitrate`. Tests verify integer and modulo frame-byte math,
repeated padding-rest transitions, prepared QC-kernel handoff, malformed
controls, no-cgo compilation, and zero steady-state allocation.

The FDK AAC-LC QC quantize/reduce-loop subset is covered with a source-shaped
`FDKaacEnc_calcMaxValueInSfb` vector, `FDKaacEnc_reduceBitConsumption` normal
and max-iteration vectors, a source-shaped `FDKaacEnc_crashRecovery` vector,
and an SCE `FDKaacEnc_QCMain` body vector. Tests compare the wrapper against
the direct scale-factor estimation, spectrum quantization, max-value scan,
dynamic noiseless bit count, `dynBitsLast` seeding, and frame dynamic-bit
summing sequence. The crash-recovery vector verifies high-band trimming to
zero spectrum, TNS/tool clearing, static side-info recounting, dynamic-budget
rebasing, retry scheduling, malformed controls, no-cgo compilation, and zero
steady-state allocation. A constraint-reset re-entry vector verifies the next
quantize pass increments gain and iteration state exactly once after the frame
decision clears the constraint flags.

The FDK AAC-LC QC fill/finalize/reservoir accounting subset is covered with
source-shaped vectors for `FDKaacEnc_getTotalConsumedBits`,
`FDKaacEnc_updateFillBits`, `FDKaacEnc_FinalizeBitConsumption`, and
`FDKaacEnc_updateBitres`. Tests verify SCE/CPE/DSE total-bit summing, VBR and
CBR fill-bit math, fill-extension sizing, byte alignment, static transport
header correction, CBR/VBR reservoir updates, malformed controls, no-cgo
compilation, and zero steady-state allocation.

The FDK AAC-LC QC bit-distribution/reservoir redistribution subset is covered
with source-shaped vectors for `FDKaacEnc_distributeElementDynBits`,
`FDKaacEnc_BitResRedistribution`, and `FDKaacEnc_prepareBitDistribution`.
Tests verify per-element dynamic-bit rounding compensation, one-subframe
reservoir level redistribution, PE grant handoff through the existing
bit-reservoir helper, minimal static bit-demand accounting, the low-reservoir
negative dynamic-grant transition, low/high reservoir error returns, malformed
controls, no-cgo compilation, and zero steady-state allocation.

The FDK AAC-LC QC one-frame control wrapper subset is covered with
source-shaped vectors for `FDKaacEnc_QCMain` setup, repeated quantize/count
staging, constraint reset, and post-quantization decision logic. Tests verify
the CBR/full-reservoir mode switch, reservoir redistribution, bit distribution,
threshold adjustment, quantize/count handoff, under-budget exit, over-budget
saving, dynamic-bit overshoot, emergency iteration spending, the pinned no-op
`checkMinFrameBitsDemand` behavior, malformed controls, no-cgo compilation,
and zero steady-state allocation.

The FDK AAC-LC bit-reservoir PE distribution subset is covered with
source-derived vectors for `FDKaacEnc_bitresCalcBitFac`,
`FDKaacEnc_DistributeBits`, and the high/low bit-reservoir PE-correction
helpers. Tests verify long/short bit-factor transitions, full/reduced/disabled
reservoir modes, negative dynamic-grant zero-PE behavior, malformed controls,
no-cgo compilation, and zero steady-state
allocation.

The FDK AAC-LC quantization subset is covered with mechanically derived ROM
table checksums and source-derived vectors for `FDKaacEnc_quantizeLines`,
`FDKaacEnc_invQuantizeLines`, `FDKaacEnc_QuantizeSpectrum`,
`FDKaacEnc_calcSfbDist`, and `FDKaacEnc_calcSfbQuantEnergyAndDist`. Tests
verify grouped quantization, quantized energy/distortion feedback, invalid
controls, and zero steady-state allocation.

## Regenerate

```sh
./scripts/gen_testvectors.sh
CGO_ENABLED=0 go test ./...
```

Set `PCM_OUT_DIR=/tmp/pcm` when regenerating if you want to keep the reference
PCM files for manual inspection. PCM files are not committed because the manifest
hashes are enough for the automated checks.
