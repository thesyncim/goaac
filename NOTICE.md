# Notices

This repository uses FFmpeg as the pinned native AAC decoder reference and links
to FFmpeg libraries dynamically via `pkg-config`.

FFmpeg source is available at `https://git.ffmpeg.org/ffmpeg.git`. The pinned
reference checkout is tag `n8.1.1`, commit
`239f2c733de417201d7ad3b3b8b0d9b63285b2b1`.

FFmpeg is distributed under LGPL/GPL terms depending on build configuration.
Applications using this package should comply with the license terms of the
FFmpeg libraries they link at build and runtime.
