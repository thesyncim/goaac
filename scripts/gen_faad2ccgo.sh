#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ccgo_bin="${CCGO:-ccgo}"

targets=("$@")
if [ "${#targets[@]}" -eq 0 ]; then
  targets=("darwin/arm64" "linux/arm64")
fi

sources=(
  "$root/third_party/faad2/libfaad/bits.c"
  "$root/third_party/faad2/libfaad/cfft.c"
  "$root/third_party/faad2/libfaad/common.c"
  "$root/third_party/faad2/libfaad/decoder.c"
  "$root/third_party/faad2/libfaad/drc.c"
  "$root/third_party/faad2/libfaad/error.c"
  "$root/third_party/faad2/libfaad/filtbank.c"
  "$root/third_party/faad2/libfaad/huffman.c"
  "$root/third_party/faad2/libfaad/is.c"
  "$root/third_party/faad2/libfaad/mdct.c"
  "$root/third_party/faad2/libfaad/mp4.c"
  "$root/third_party/faad2/libfaad/ms.c"
  "$root/third_party/faad2/libfaad/output.c"
  "$root/third_party/faad2/libfaad/pns.c"
  "$root/third_party/faad2/libfaad/pulse.c"
  "$root/third_party/faad2/libfaad/specrec.c"
  "$root/third_party/faad2/libfaad/syntax.c"
  "$root/third_party/faad2/libfaad/tns.c"
)

mkdir -p "$root/internal/faad2ccgo"

for target in "${targets[@]}"; do
  goos="${target%/*}"
  goarch="${target#*/}"
  out="$root/internal/faad2ccgo/faad2_${goos}_${goarch}.go"
  echo "generating $out"
  "$ccgo_bin" \
    --goos "$goos" \
    --goarch "$goarch" \
    --package-name faad2ccgo \
    -ignore-unsupported-alignment \
    -o "$out" \
    -I "$root/third_party/faad2/include" \
    -I "$root/third_party/faad2/libfaad" \
    -DLC_ONLY_DECODER \
    -DDISABLE_SBR \
    -DHAVE_INTTYPES_H=1 \
    -DHAVE_MEMCPY=1 \
    -DHAVE_STRING_H=1 \
    -DHAVE_STRINGS_H=1 \
    -DHAVE_SYS_STAT_H=1 \
    -DHAVE_SYS_TYPES_H=1 \
    -DHAVE_LRINTF=1 \
    -DPACKAGE_VERSION='"2.11.2"' \
    -DFAAD2_VERSION='"2.11.2"' \
    -O2 \
    "${sources[@]}"
done
