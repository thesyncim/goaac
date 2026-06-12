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

## Regenerate

```sh
./scripts/gen_testvectors.sh
CGO_ENABLED=0 go test ./...
```

Set `PCM_OUT_DIR=/tmp/pcm` when regenerating if you want to keep the reference
PCM files for manual inspection. PCM files are not committed because the manifest
hashes are enough for the automated checks.
