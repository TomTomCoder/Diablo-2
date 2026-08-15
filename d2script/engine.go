// Package d2script hosts Devil's scripting/modding layer: sandboxed WASM
// modules, replacing the previous otto-based JavaScript interpreter (see
// ROADMAP.md Phase 3). Scripts are compiled ahead of time to WASM -- from
// any language that targets it (Go via `GOOS=wasip1 GOARCH=wasm go build`,
// Rust, TinyGo, AssemblyScript) -- rather than interpreted from source
// text. That's both the modern approach for embeddable scripting and a
// real sandboxing improvement over otto: a WASM guest can only call the
// specific host functions it's explicitly given (see AddFunction) and can
// never reach arbitrary Go/process state the way otto's Eval() could.
package d2script

import (
	"context"
	"fmt"
	"os"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// hostModuleName is the module name Devil's own host functions (see
// AddFunction) are exported under; a guest WASM module imports them by
// this name (e.g. Go's `//go:wasmimport devil funcName`).
const hostModuleName = "devil"

// ScriptEngine hosts and runs sandboxed WASM script modules.
type ScriptEngine struct {
	runtime     wazero.Runtime
	hostBuilder wazero.HostModuleBuilder
	ctx         context.Context
	hostReady   bool
}

// CreateScriptEngine creates the script engine and returns a pointer to it.
func CreateScriptEngine() *ScriptEngine {
	ctx := context.Background()
	runtime := wazero.NewRuntime(ctx)

	// WASI gives guest modules (e.g. ones compiled from Go) the minimal
	// syscall surface the toolchain expects at startup -- it does not on
	// its own grant filesystem/network access beyond what ModuleConfig
	// below explicitly wires up (nothing, today).
	wasi_snapshot_preview1.MustInstantiate(ctx, runtime)

	return &ScriptEngine{
		runtime:     runtime,
		hostBuilder: runtime.NewHostModuleBuilder(hostModuleName),
		ctx:         ctx,
	}
}

// AddFunction registers a Go function under name, callable from guest WASM
// modules that import it from the "devil" host module. Must be called
// before the first RunScript -- once the host module is instantiated
// (lazily, on first RunScript), later AddFunction calls have no effect.
//
// Unlike otto's arbitrary interface{} values, WASM function signatures are
// limited to numeric/memory types -- fn's signature must be one wazero's
// reflection-based WithFunc supports (numeric params/results, optionally a
// leading context.Context and/or api.Module).
func (s *ScriptEngine) AddFunction(name string, fn interface{}) {
	s.hostBuilder = s.hostBuilder.NewFunctionBuilder().WithFunc(fn).Export(name)
}

// ensureHostReady instantiates the host module on first use. Idempotent.
func (s *ScriptEngine) ensureHostReady() error {
	if s.hostReady {
		return nil
	}

	if _, err := s.hostBuilder.Instantiate(s.ctx); err != nil {
		return fmt.Errorf("could not instantiate host module: %w", err)
	}

	s.hostReady = true

	return nil
}

// RunScript loads the compiled WASM module at path, instantiates it, and
// -- if present -- calls its exported "run" function.
func (s *ScriptEngine) RunScript(path string) error {
	if err := s.ensureHostReady(); err != nil {
		return err
	}

	wasmBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read script file: %w", err)
	}

	mod, err := s.runtime.InstantiateWithConfig(s.ctx, wasmBytes,
		wazero.NewModuleConfig().WithStdout(os.Stdout).WithStderr(os.Stderr))
	if err != nil {
		return fmt.Errorf("could not instantiate script module: %w", err)
	}

	defer mod.Close(s.ctx) //nolint:errcheck // best-effort cleanup

	run := mod.ExportedFunction("run")
	if run == nil {
		return nil
	}

	if _, err := run.Call(s.ctx); err != nil {
		return fmt.Errorf("script run() failed: %w", err)
	}

	return nil
}

// Close releases the engine's runtime resources. Callers should defer this
// once they're done running scripts.
func (s *ScriptEngine) Close() error {
	return s.runtime.Close(s.ctx)
}
