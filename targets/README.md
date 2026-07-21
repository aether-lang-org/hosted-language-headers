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
  x86_64-windows/python/pyconfig.h       # captured from python.org 3.12.8 on Windows
```

Ruby's configured header is `<ruby/config.h>` and is reached from a SEPARATE
include root (`ruby-arch/`, mirrored by the shared tree). So ruby overlays
provide their own `ruby-arch/` root that goes on `-I` *before* the shared one —
no shared-tree removal needed (different `-I` dir, first match wins), unlike
python where the shared `pyconfig.h` had to be deleted. A target may also carry
platform-only headers that don't exist in the Debian-captured shared tree:

```
  x86_64-windows/ruby-arch/ruby/config.h                        # the win config
  x86_64-windows/ruby-win/ruby/win32.h                          # Windows-only
  x86_64-windows/ruby-win/ruby/internal/intern/select/win32.h   # Windows-only
```

When a target's language VERSION differs from the shared tree (Debian ruby is
3.1; FreeBSD 15 ships ruby 3.4), don't mix a new arch `config.h` onto old
portable headers — capture the target's WHOLE header tree as a self-contained
overlay instead. `x86_64-freebsd/ruby/` is the full ruby-3.4 tree from .204
(`ruby.h` + `ruby/` + the arch root `amd64-freebsd15/ruby/config.h`), consumed
with `-Itargets/x86_64-freebsd/ruby/amd64-freebsd15 -Itargets/x86_64-freebsd/ruby`
and NOT `-I<repo>/ruby`. The dlopen model makes this correct: the runtime IS
3.4, so the headers should be 3.4 too.

## Consuming (a cross-build)

Put the target overlay on the include path BEFORE the shared tree:

```
zig cc -target x86_64-macos-none \
  -I<repo>/targets/x86_64-macos/python \
  -I<repo>/python  <runtime -I…>  -c aether_host_python.c
```

The overlay's `pyconfig.h` resolves first; the shared `Python.h` (which
`#include "pyconfig.h"`) finds it there. No `-D` hand-patching, no dispatcher.

Ruby (note the two overlay roots ahead of the shared `ruby/`):

```
zig cc -target x86_64-windows-gnu \
  -I<repo>/targets/x86_64-windows/ruby-arch \
  -I<repo>/targets/x86_64-windows/ruby-win \
  -I<repo>/ruby  <runtime -I…>  -c aether_host_ruby.c
```

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
| x86_64-linux-gnu | ✅ (Debian 12) | ✅ shared ruby-arch | — | — | — |
| x86_64-macos     | ✅ (3.12, py.org) RUNS | needs ruby-3.1-dev capture | — | — | — |
| x86_64-freebsd   | ✅ (3.12.13) RUNS | ✅ (3.4.9, full tree) RUNS | — | — | — |
| x86_64-windows   | ✅ (3.12.8) RUNS | ✅ (3.1.4 ucrt) RUNS | — | — | — |

RUNS = embedded interpreter proven executing inside a Linux-cross-built binary
on a real machine of that target (macOS Hackintosh KVM; Windows 11 VM; FreeBSD 15 box .204).

python proven end-to-end: cross-built on Linux, embedded Python RUNS on real
macOS 15 (via the overlay's config). Others follow the same capture pattern.
