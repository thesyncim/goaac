# goaac

`goaac` is a Go AAC-LC decoder package with a production native decode path and
source-shaped parsing for MPEG-4 AudioSpecificConfig and ADTS streams.

The decoder is pinned to FFmpeg `n8.1.1` as the C reference source. The current
production path dynamically links FFmpeg's native AAC decoder through cgo, while
the pure Go parser code tracks the upstream bitstream helpers closely enough to
serve as the first parity layer.

## Requirements

- Go 1.22+
- `pkg-config`
- FFmpeg development libraries: `libavcodec`, `libavutil`, `libswresample`

On macOS with Homebrew:

```sh
brew install ffmpeg pkg-config
```

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

## CLI

```sh
go run ./cmd/goaac-decode input.aac output.wav
```

## Validation

```sh
go test ./...
```

The integration test synthesizes an AAC-LC ADTS fixture with `ffmpeg`, decodes it
with this package, and byte-compares the resulting S16 PCM against an FFmpeg CLI
oracle.

## Reference

- Upstream: `https://git.ffmpeg.org/ffmpeg.git`
- Tag: `n8.1.1`
- Commit: `239f2c733de417201d7ad3b3b8b0d9b63285b2b1`
- Local checkout: `third_party/ffmpeg`

See `docs/reference.md` and `docs/parity-ledger.md` for the source-truth record.
