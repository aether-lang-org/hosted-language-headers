// Compiled by TinyGo to a c-shared .so (libgreet.so) IN the toolchain
// container (Phase 1), loaded by the Aether program ON the host (Phase 2).
// Mirrors aether contrib/host/tinygo/examples/greet.go.
package main

import "C"

//export Answer
func Answer() int32 { return 42 }

//export Add
func Add(a, b int32) int32 { return a + b }

func main() {} // c-shared still requires main()
