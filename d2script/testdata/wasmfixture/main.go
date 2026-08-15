// Command wasmfixture is d2script's test fixture: a minimal guest module
// that imports two host functions ("devil"."recordCall" and
// "devil"."showText") and calls each once, proving the host<->guest
// wiring actually works end to end -- for both a plain numeric argument
// and a string passed as a (pointer, length) pair, the standard way a
// WASM guest hands a string to its host (see d2script.ReadString's own
// doc comment). Not part of the main module's build (see the build
// constraint below) -- it's compiled ahead of time into ../hello.wasm.
//
// To rebuild after changing this file:
//
//	GOOS=wasip1 GOARCH=wasm go build -o ../hello.wasm .
//
//go:build wasip1

package main

import "unsafe"

//go:wasmimport devil recordCall
func recordCall(n int32)

//go:wasmimport devil showText
func showText(ptr, length uint32)

func main() {
	recordCall(42)

	text := "Hello from Devil's narrative script"
	ptr := uint32(uintptr(unsafe.Pointer(unsafe.StringData(text))))
	showText(ptr, uint32(len(text)))
}
