# Per-target configured headers

The top-level trees (`python/`, `ruby/`, `lua/`, `perl/`, `js/`) carry the
**portable** header declarations — identical across every Unix-y target. But a
few files are **configured per platform** at the language's `./configure` /
packaging time (see `../PLATFORM.md`): `pyconfig.h`, `luaconf.h`,
`ruby/config.h`, `duk_config.h`, perl's `config.h`. Those encode
`sizeof(long)`, which syscalls exist, `HAVE_STRLCPY`, `HAVE_ALLOCA_H`,
endianness — real per-OS/arch facts. A Debian config does NOT describe macOS.

This directory holds those configured files **per target triple**, captured
from a real machine of that platform. The shared trees deliberately DO NOT ship
a configured header (e.g. `python/` has no `pyconfig.h`) — it comes only from
the target overlay, so it can never be wrong-for-target.

## Layout

```
targets/<triple>/<lang>/<configured-header>
  x86_64-macos/python/pyconfig.h        # captured from python.org 3.12 on macOS 15
  x86_64-freebsd/python/pyconfig.h       # captured from FreeBSD 15 python 3.12
  x86_64-linux-gnu/python/pyconfig.h     # the real Debian 12 config (was the dispatcher)
```

## Consuming (a cross-build)

Put the target overlay on the include path BEFORE the shared tree:

```
zig cc -target x86_64-macos-none \
  -I<repo>/targets/x86_64-macos/python \
  -I<repo>/python  <runtime -I…>  -c aether_host_python.c
```

The overlay's `pyconfig.h` resolves first; the shared `Python.h` (which
`#include "pyconfig.h"`) finds it there. No `-D` hand-patching, no dispatcher.

## Why overlay dirs, not branch-per-platform

`../README.md` originally suggested a git branch per target. That's fine for a
single-target native build, but a **cross-build tool needs all targets at once**
— it can't check out three branches. Overlay dirs give every target
simultaneously with a tiny footprint (only the divergent config files, a few KB
each), and the diff between targets is visible in one tree — which is itself the
per-arch documentation `PLATFORM.md` wanted.

## Capturing a new target's config

On a machine running that target, with the language's `-dev`/headers installed:

```
# python: find the real (non-dispatcher) pyconfig.h
find / -name pyconfig.h -path '*python3.*' 2>/dev/null   # the configured one
cp <that>  targets/<triple>/python/pyconfig.h
```

Do the same for luaconf.h / ruby/config.h / duk_config.h / perl config.h.
The configured header must be SELF-CONTAINED (a real config, not Debian's
multiarch dispatcher which `#error`s off its host arch).

## Status

| target | python | ruby | lua | perl | js |
|---|---|---|---|---|---|
| x86_64-linux-gnu | ✅ (Debian 12) | — | — | — | — |
| x86_64-macos     | ✅ (3.12, py.org) | needs ruby-3.1-dev capture | — | — | — |
| x86_64-freebsd   | ✅ (3.12) | needs capture | — | — | — |
| x86_64-windows   | needs Windows py config | needs capture | — | — | — |

python proven end-to-end: cross-built on Linux, embedded Python RUNS on real
macOS 15 (via the overlay's config). Others follow the same capture pattern.
