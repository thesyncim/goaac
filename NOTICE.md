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

This repository does not include an AAC encoder and does not provide, claim, or
imply any patent license or patent non-infringement guarantee. Users and
redistributors are responsible for evaluating patent obligations in their own
jurisdictions and use cases.
