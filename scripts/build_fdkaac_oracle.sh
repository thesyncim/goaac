#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
src="$root/third_party/fdk-aac"
build_dir="${FDKAAC_BUILD_DIR:-$root/dist/fdk-aac-oracle}"

if [ ! -f "$src/libAACenc/include/aacenc_lib.h" ]; then
  cat >&2 <<EOF
missing FDK-AAC source at $src
run: git submodule update --init --recursive third_party/fdk-aac
EOF
  exit 1
fi

jobs="${JOBS:-}"
if [ -z "$jobs" ]; then
  if command -v getconf >/dev/null 2>&1; then
    jobs="$(getconf _NPROCESSORS_ONLN 2>/dev/null || true)"
  fi
  if [ -z "$jobs" ] && command -v sysctl >/dev/null 2>&1; then
    jobs="$(sysctl -n hw.ncpu 2>/dev/null || true)"
  fi
  jobs="${jobs:-2}"
fi

cmake \
  -S "$src" \
  -B "$build_dir" \
  -DCMAKE_BUILD_TYPE=Release \
  -DBUILD_SHARED_LIBS=OFF \
  -DBUILD_PROGRAMS=ON \
  -DFDK_AAC_INSTALL_CMAKE_CONFIG_MODULE=OFF \
  -DFDK_AAC_INSTALL_PKGCONFIG_MODULE=OFF

cmake --build "$build_dir" --target aac-enc --config Release --parallel "$jobs"

printf '%s\n' "$build_dir/aac-enc"
