# Parity Ledger

| Upstream path/function | Local path/function | Status | Parity proof | Known deviations |
| --- | --- | --- | --- | --- |
| `libfaad/*.c` LC decoder slice | `internal/faad2ccgo/faad2_*.go` | generated translation | native FAAD2 oracle byte comparison | Generated for `darwin/arm64` and `linux/arm64`; source macros are `LC_ONLY_DECODER` and `DISABLE_SBR`. |
| `libfaad/decoder.c:NeAACDecOpen` | `decoder_purego.go:newPureDecoder` | wrapped | integration test | Go wrapper fixes output to interleaved S16 and disables implicit SBR upsampling. |
| `libfaad/decoder.c:NeAACDecInit` | `decoder_purego.go:initADTS` | wrapped | integration test | Public ADTS path rejects frames with multiple raw data blocks before calling FAAD2. |
| `libfaad/decoder.c:NeAACDecInit2` | `decoder_purego.go:initRaw` | wrapped | raw config tests plus compile gate | Public raw path accepts AAC-LC AudioSpecificConfig only. |
| `libfaad/decoder.c:NeAACDecDecode` | `decoder_purego.go:decode` | wrapped | native FAAD2 oracle byte comparison | Samples are copied out of FAAD2-owned scratch before the next decode call. |
| `libfaad/error.c:NeAACDecGetErrorMessage` | `decoder_purego.go:frameError` | wrapped | integration error path coverage | Error text is surfaced with the FAAD2 code. |
| `include/neaacdec.h` output structs | `internal/faad2ccgo.NeAACDecFrameInfo` | generated translation | compile gate | Field names use generated `F...` prefixes. |
| MPEG-4 AudioSpecificConfig parsing | `config.go:ParseAudioSpecificConfig` | hand-translated subset | unit tests | AAC-LC/SBR/PS detection only; unsupported extensions are rejected by the public facade. |
| ADTS fixed/variable header parsing | `adts.go:ParseADTSHeader` | hand-translated subset | unit tests, integration fixture | Rejects multiple raw data blocks because the public decoder is one access unit at a time. |
| ADTS header emission | `adts.go:AppendADTSHeader` | hand-translated subset | unit tests | Writes CRC-less ADTS headers only. |
