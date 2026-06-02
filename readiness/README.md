# readiness — toolchain & aeb smoke tests for the aether-builder container

Two staged probes that confirm an `aether-build` container image is sound.
Both compile + run; if they pass, the container can build real work.

## Test 1 — bare Aether toolchain (`hello.ae`)

```sh
ae build hello.ae -o hello && ./hello      # → aether-ready
```

Proves aetherc + gcc + libaether + stdlib work in the container. The
minimal "does the compiler exist and function" probe.

## Test 2 — aeb multi-target DAG (`aebproj/`)

A small, **pure-Aether** (no cross-language deps) hierarchy:

```
aebproj/                      .aeb         project root marker
  greeter/  greeter_main.ae   .build.ae    leaf target (has main → runnable)
  app/      app.ae            .build.ae    root target: build.dep(greeter) + imports greeting()
```

```sh
cd aebproj && aeb app/.build.ae             # builds greeter, then app (topo order)
./target/build/app/bin/program              # → hello-from-greeter
```

Proves aeb itself drives a multi-node DAG in the container: the
`build.dep("greeter/.build.ae")` edge, topo-sort, a cross-dir module import
(`import greeter.greeter_main`), and the `target/<buildtype>/<dir>` layout.
No rust/go/java — deliberately, so the Aether-only toolchain image suffices
(the google-monorepo-sim's aether targets dep on rust, which a pure-Aether
image can't build; this harness is the dependency-free readiness equivalent).

## Why here

`aether-build` already vendors host headers from this repo at a pinned
commit; co-locating the readiness probes means the same single source
acquisition carries the smoke tests. Run them after an image bootstrap to
confirm the toolchain (Test 1) and aeb (Test 2) are functional before
trusting the image with real builds.
