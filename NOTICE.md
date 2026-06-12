# Notices

This repository contains a pure-Go translation of FAAD2 `2.11.2` for AAC-LC
decode. The translated files under `internal/faad2ccgo` are generated from the
pinned FAAD2 C source in `third_party/faad2`.

FAAD2 is distributed under the GNU General Public License, version 2 or later.
The required FAAD2 notice is:

Code from FAAD2 is copyright (c) Nero AG, www.nero.com

The original FAAD2 source is available at `https://github.com/knik0/faad2`.
The pinned checkout is tag `2.11.2`, commit
`673a22a3c7c33e96e2ff7aae7c4d2bc190dfbf92`.

FFmpeg is used by tests only to synthesize AAC-LC ADTS fixtures. It is not linked
or loaded by the decoder runtime.

This repository also contains a pinned FDK-AAC `v2.0.3` source checkout under
`third_party/fdk-aac` for encoder reference and native oracle generation.
FDK-AAC is distributed under the "Software License for The Fraunhofer FDK AAC
Codec Library for Android"; see `third_party/fdk-aac/NOTICE` for the complete
license text. That license does not grant patent rights.
