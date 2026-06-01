# hosted-language-headers

Pre-captured C headers for the languages [Aether](https://github.com/aether-lang-org/aether)'s
`contrib.host.*` bridges embed. Used by the
[`aether-builder` Docker/Podman image](https://github.com/aether-lang-org/aether/tree/main/tools/docker)
to avoid installing `python3-dev` / `liblua5.4-dev` / `libperl-dev`
/ `ruby-dev` / `duktape-dev` in every image build.

## Layout

```
python/         Python 3.11 headers (Python.h + everything pulled by it)
lua/            Lua 5.4 headers (lua.h, lauxlib.h, lualib.h, luaconf.h)
perl/           Perl 5.36 CORE/ headers (EXTERN.h, perl.h, ~80 files)
ruby/           Ruby 3.1.2 portable headers (ruby.h and tree)
ruby-arch/      Ruby 3.1.2 platform-specific headers (ruby/config.h)
js/             duktape 2.7.0 headers (duktape.h + duk_config.h)
```

Sizes (committed):

| Tree      | Size    |
|-----------|---------|
| python    | ~1.3 MB |
| lua       | ~50 KB  |
| perl      | ~10 MB  |
| ruby      | ~2 MB   |
| ruby-arch | ~1 MB   |
| js        | ~200 KB |
| **total** | **~14 MB** |

## Per-OS / per-arch

This branch (`main`) carries the headers for **linux-x86_64-glibc**
(captured from Debian 12 bookworm). See `PLATFORM.md` for the
package-version snapshot and which files are platform-sensitive.

When another target is needed (linux-arm64, linux-musl, macOS, …)
make a new branch named after the target — `linux-arm64-glibc`,
etc. — and populate from a machine running that platform. The
configured `*config.h` files diverge per arch; the rest of each
tree should be nearly identical, and the diff between branches
documents the per-arch surface.

## Usage

The Aether builder image clones this repo at build time and
copies subtrees into `/opt/aether/include/<lang>/`. Consumers
that don't use Docker can `git clone` this repo and point their
toolchain at `--include-dir <path-to-this-repo>/python` (or
whichever language).

## Licensing

Each language's headers are licensed under that language's
own license:

| Language | License |
|----------|---------|
| Python   | Python Software Foundation License (PSFL) |
| Lua      | MIT |
| Perl     | Artistic License / GPL (dual) |
| Ruby     | BSD-2-Clause + Ruby License (dual) |
| duktape  | MIT |

All five permit redistribution of source files with attribution.
The headers' upstream copyright / license notices are preserved
inside the files themselves (the `Copyright …` blocks at the top
of `Python.h`, `lua.h`, `EXTERN.h`, `ruby.h`, `duktape.h`); no
extra `LICENSE` files are added at the repo level.

## How the headers were captured

```bash
sudo apt install -y --no-install-recommends \
    python3-dev liblua5.4-dev libperl-dev ruby-dev duktape-dev

# Python
PY_INCDIR="$(python3 -c 'import sysconfig; print(sysconfig.get_path("include"))')"
cp -r "$PY_INCDIR/." python/

# Lua
cp /usr/include/lua5.4/*.h lua/

# Perl
PERL_INCLUDE="$(perl -MConfig -e 'print $Config{archlibexp}')/CORE"
cp -r "$PERL_INCLUDE/." perl/

# Ruby (two trees — portable + arch-specific)
RUBY_HDRDIR="$(ruby -rrbconfig -e 'print RbConfig::CONFIG["rubyhdrdir"]')"
RUBY_ARCHHDRDIR="$(ruby -rrbconfig -e 'print RbConfig::CONFIG["rubyarchhdrdir"]')"
cp -r "$RUBY_HDRDIR/." ruby/
cp -r "$RUBY_ARCHHDRDIR/." ruby-arch/

# duktape
cp /usr/include/duktape.h /usr/include/duk_config.h js/
```

## What's NOT here

- **Java**. `contrib.host.java` doesn't use C headers — it ships
  a prebuilt `aether-sandbox.jar`. End users running a binary that
  uses contrib.host.java need a JRE installed on their target host;
  the JAR is built upstream with a JDK that supports the Foreign
  Function & Memory API (Java 21+).
- **TinyGo, Tcl, Go** — `contrib/host/{tinygo,tcl,go}` exist in
  Aether but aren't currently driven by this header-capture flow.
- **Runtime shared libraries** (`libpython3.so`, `libruby.so`,
  `liblua5.4.so`, etc.) — those come from the user's target-host
  package manager. The aether-builder image installs the
  runtime-only variants when a `WITH_<LANG>=1` build arg is set;
  the user's deployment host installs them separately.
