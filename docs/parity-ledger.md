# Parity Ledger

| Upstream path/function | Local path/function | Status | Parity proof | Known deviations |
| --- | --- | --- | --- | --- |
| `libavcodec/mpeg4audio.c:ff_mpeg4audio_get_config_gb` | `config.go:ParseAudioSpecificConfig` | translated subset | unit tests | AAC-LC/SBR/PS detection only; ALS and sync-extension scanning are not part of the AAC-LC production gate yet. |
| `libavcodec/mpeg4audio.c:get_object_type` | `config.go:readObjectType` | translated | unit tests | None for escaped object type syntax. |
| `libavcodec/mpeg4audio.c:get_sample_rate` | `config.go:readSampleRate` | translated | unit tests | None for standard table and explicit 24-bit rate. |
| `libavcodec/mpeg4audio.c:ff_mpeg4audio_channels` | `config.go:mpeg4AudioChannels` | copied | unit tests | None. |
| `libavcodec/mpeg4audio_sample_rates.c` | `config.go:mpeg4AudioSampleRates` | copied | unit tests | None. |
| `libavcodec/adts_header.c:ff_adts_header_parse` | `adts.go:ParseADTSHeader` | translated subset | unit tests, integration fixture | Rejects multiple raw data blocks for now because the public decoder path is framed one access unit at a time. |
| `libavformat/adtsenc.c:adts_write_frame_header` | `adts.go:AppendADTSHeader` | translated subset | unit tests | Writes CRC-less ADTS headers only. |
| `libavcodec/aac/aacdec.c:aac_decode_frame` | `ffmpeg_bridge.c:aaclc_decode` | wrapped native reference | integration byte oracle | Native FFmpeg path is used directly; pure Go spectral decode is not claimed yet. |
| `libswresample` S16 conversion | `ffmpeg_bridge.c:convert_frame` | wrapped native reference | integration byte oracle | Output is fixed to interleaved S16 PCM. |
