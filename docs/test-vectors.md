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

The FDK AAC-LC short grouping subset is covered with source-derived grouped
offset and min-SNR vectors, saturating threshold/energy/MS/spread-energy sums,
max-SFB scan and zero-spectrum floor checks, spectrum-regroup hashes, invalid
grouping/control transitions, and zero steady-state allocation.

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

The FDK AAC-LC line-PE subset is covered with source-derived vectors for
`CalcLdInt`, `fMultI`, `FDKaacEnc_prepareSfbPe`, and `FDKaacEnc_calcSfbPe`.
Tests verify active-line preparation, high/low PE branches, inactive and
intensity-book transitions, malformed controls, and zero steady-state
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
