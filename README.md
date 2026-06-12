# goaac

`goaac` is a pure-Go AAC-LC decoder package. The decoder core is a source-shaped
Go translation of FAAD2 `2.11.2`, generated from the pinned C source and wrapped
with a small Go API for ADTS and raw AAC-LC access units.

The runtime path does not link to FFmpeg, FAAD2, or any other native library.
Generated decoder files are currently checked in for `darwin/arm64` and
`linux/arm64`; other targets fail with an explicit unsupported-target error.

## Requirements

- Go 1.25+
- Supported runtime target: `darwin/arm64` or `linux/arm64`
- Tests: `ffmpeg` to synthesize AAC-LC fixtures and a C compiler for the native
  FAAD2 oracle

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

The integration test synthesizes an AAC-LC ADTS fixture with `ffmpeg`, builds a
small native FAAD2 oracle from `third_party/faad2`, and byte-compares S16 PCM
from the Go port against the C reference.

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

See `docs/reference.md` and `docs/parity-ledger.md` for the source-truth record.
