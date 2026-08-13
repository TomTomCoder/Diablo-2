// Command wasmfixture is d2script's test fixture: a minimal guest module
// that imports a host function ("devil"."recordCall") and calls it once
// with a fixed value, proving the host<->guest wiring actually works end
// to end. Not part of the main module's build (see the build constraint
// below) -- it's compiled ahead of time into ../hello.wasm.
//
// To rebuild after changing this file:
//
//	GOOS=wasip1 GOARCH=wasm go build -o ../hello.wasm .
//
//go:build wasip1

package main

//go:wasmimport devil recordCall
func recordCall(n int32)

func main() {
	recordCall(42)
}
