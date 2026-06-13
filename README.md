# goaac

`goaac` is a pure-Go AAC-LC package. The decoder core is a source-shaped Go
translation of FAAD2 `2.11.2`; the encoder track is a source-shaped FDK-AAC
`v2.0.3` port with a Go API for mono/stereo AAC-LC raw access units, ADTS
frames, and RTMP/FLV AAC audio-message bodies.

The runtime path does not link to FFmpeg, FAAD2, FDK-AAC, or any other native
library.
Generated decoder files are currently checked in for `darwin/arm64` and
`linux/arm64`; other targets fail with an explicit unsupported-target error.

## Requirements

- Go 1.25+
- Supported runtime target: `darwin/arm64` or `linux/arm64`
- Normal tests use committed AAC-LC vectors and need no native decoder
- Vector regeneration and live oracle synthesis need `ffmpeg` and a C compiler

## Decode ADTS AAC-LC

```go
data, err := os.ReadFile("input.aac")
if err != nil {
    panic(err)
}

pcm, cfg, err := aac.DecodeADTS(data)
if err != nil {
    panic(err)
}

fmt.Println(cfg.SampleRate, cfg.Channels, len(pcm))
```

The returned PCM is interleaved signed 16-bit native-endian samples.

## Encode AAC-LC

The encoder accepts one 1024-sample interleaved S16 PCM frame per call. Use raw
AAC access units for MP4/FLV/RTMP muxers and ADTS when you need self-framed AAC:

```go
enc, err := aac.NewEncoder(aac.EncoderOptions{
    Config: aac.Config{
        ObjectType:    aac.AOTAACLC,
        SampleRate:    48000,
        ChannelConfig: 2,
    },
    BitRate:   128000,
    Transport: aac.TransportRaw,
})
if err != nil {
    panic(err)
}
defer enc.Close()

var au []byte
au, info, err := enc.EncodeRawInto(au[:0], pcm1024InterleavedS16)
if err != nil {
    panic(err)
}
fmt.Println(info.PayloadBytes)
```

For ADTS:

```go
enc, err := aac.NewEncoder(aac.EncoderOptions{
    Config:    aac.Config{ObjectType: aac.AOTAACLC, SampleRate: 48000, ChannelConfig: 2},
    BitRate:   128000,
    Transport: aac.TransportADTS,
})
frame, info, err := enc.EncodeADTSFrameInto(nil, pcm1024InterleavedS16)
fmt.Println(len(frame), info.ADTSHeaderBytes)
```

For RTMP/FLV AAC audio payloads, send one sequence header, then raw messages:

```go
seq, err := enc.AppendRTMPSequenceHeader(nil)
msg, info, err := enc.EncodeRTMPMessageInto(nil, pcm1024InterleavedS16)
_ = seq
_ = msg
_ = info
```

## Decoder API

Use `New` when you want explicit transport selection and reusable output
buffers:

```go
dec, err := aac.New(aac.Options{Transport: aac.TransportADTS})
if err != nil {
    panic(err)
}
defer dec.Close()

var pcm []int16
for _, frame := range frames {
    var info aac.FrameInfo
    pcm, info, err = dec.Decode(pcm, frame.Data)
    if err != nil {
        panic(err)
    }
    fmt.Println(info.SampleRate, info.Channels, info.OutputSamples)
}
```

For raw AAC-LC access units, pass a raw config:

```go
dec, err := aac.New(aac.Options{
    Transport: aac.TransportRaw,
    Config: aac.Config{
        ObjectType:    aac.AOTAACLC,
        SampleRate:    44100,
        ChannelConfig: 2,
    },
})
```

## Decode RTMP/FLV AAC-LC

RTMP audio messages carry FLV audio tag bodies. Pass each AAC audio message
payload to `FLVAACDecoder` in stream order:

```go
dec := aac.NewRTMPAACDecoder()
defer dec.Close()

var pcm []int16
for _, payload := range rtmpAudioPayloads {
    var info aac.FLVAACFrameInfo
    pcm, info, err = dec.DecodeRTMPMessage(pcm, payload)
    if err != nil {
        panic(err)
    }
    if info.SequenceHeader {
        fmt.Println(info.Config.SampleRate, info.Config.Channels)
        continue
    }
    fmt.Println(info.OutputSamples)
}
```

The first AAC sequence-header tag initializes or reconfigures the raw AAC-LC
decoder from MPEG-4 AudioSpecificConfig. Raw AAC tags append interleaved S16 PCM
to the caller-owned buffer. `ParseFLVAudioTag` is also exposed when callers need
zero-copy inspection before decode.

For streaming ADTS input, avoid loading the whole file:

```go
f, err := os.Open("input.aac")
if err != nil {
    panic(err)
}
defer f.Close()

pcm, cfg, err := aac.DecodeADTSReader(f)
```

## CLI

```sh
go run ./cmd/goaac-decode input.aac output.wav
```

## Validation

```sh
CGO_ENABLED=0 go test ./...
CGO_ENABLED=0 go vet -unsafeptr=false -unreachable=false ./...
CGO_ENABLED=0 go build ./cmd/goaac-decode
```

The normal test suite checks committed ADTS AAC-LC decoder vectors in
`testdata/aaclc`
against S16 PCM SHA-256 values generated from the pinned FAAD2 reference. The
same fixtures are decoded through ADTS, raw access-unit, and RTMP/FLV AAC tag
paths. The live integration test can also synthesize a fresh fixture with
`ffmpeg`, build a small native FAAD2 oracle from `third_party/faad2`, and
byte-compare the Go decoder output against it.

Encoder tests pin the pure-Go FDK-shaped raw access-unit SHA-256 for a
deterministic stereo S16 frame, verify ADTS and RTMP/FLV wrappers carry the same
payload, check invalid control transitions, and enforce zero allocations on the
initialized raw encode hot path.

To regenerate the committed vectors:

```sh
./scripts/gen_testvectors.sh
CGO_ENABLED=0 go test ./...
```

## Regenerating The Port

```sh
CCGO=/path/to/ccgo ./scripts/gen_faad2ccgo.sh
```

The default generation targets are `darwin/arm64` and `linux/arm64`. Additional
targets can be passed as `GOOS/GOARCH` arguments when the local C compiler setup
can provide a matching configuration.

## Reference

- Upstream: `https://github.com/knik0/faad2`
- Tag: `2.11.2`
- Commit: `673a22a3c7c33e96e2ff7aae7c4d2bc190dfbf92`
- Local checkout: `third_party/faad2`

The encoder track uses pinned FDK-AAC `v2.0.3` as its quality and parity
reference. Build the native encoder oracle with:

```sh
./scripts/build_fdkaac_oracle.sh
```

See `docs/reference.md`, `docs/encoder-reference.md`, and
`docs/parity-ledger.md` for the source-truth record.
