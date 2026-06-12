#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "usage: scripts/oracle_ffmpeg_pcm.sh input.aac output.s16le" >&2
  exit 2
fi

ffmpeg -hide_banner -loglevel error \
  -i "$1" \
  -f s16le \
  -acodec pcm_s16le \
  "$2"
