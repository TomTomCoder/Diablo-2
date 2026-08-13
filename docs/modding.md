# Modding Devil (WASM scripts)

Devil's scripting layer (`d2script`) runs sandboxed WebAssembly modules
instead of interpreting a scripting language from source text. A script is
a compiled `.wasm` binary that:

- **Imports** host functions Devil's engine exposes (see "Available host
  functions" below) to read or affect game state.
- **Exports** a `run` function, called once by the engine when the script
  is loaded.

That's the entire contract. There's no `eval`, no arbitrary code string
execution, and no access to anything the engine hasn't explicitly exported
as a host function -- a script can only do what it's given the tools to do.

## Writing a script

Any language that compiles to WASM works. The simplest path, since it
needs no extra toolchain, is Go itself:

```go
//go:build wasip1

package main

// Import a host function Devil's engine exposes. The module name is
// always "devil"; see "Available host functions" for what's defined.
//go:wasmimport devil mapEngineCount
func mapEngineCount() uint32

func run() {
	_ = mapEngineCount()
	// ... do something with it
}

func main() {} // required by the wasip1 target, unused by the engine
```

Compile it with the stock Go toolchain -- no tinygo, no clang, no wasm-pack:

```sh
GOOS=wasip1 GOARCH=wasm go build -o myscript.wasm .
```

Rust, TinyGo, and AssemblyScript work too, using their own equivalent of
`//go:wasmimport` (import declarations targeting the `devil` module).

## Loading a script

```go
engine := d2script.CreateScriptEngine()
defer engine.Close()

engine.AddFunction("mapEngineCount", func() uint32 {
	return uint32(len(gameServer.mapEngines))
})

if err := engine.RunScript("myscript.wasm"); err != nil {
	// ...
}
```

`AddFunction` must be called for every host function a script might import
*before* the first `RunScript` call -- the host module is instantiated
lazily on first use, and later additions don't take effect.

## Available host functions

| Name | Signature | Notes |
|------|-----------|-------|
| `mapEngineCount` | `func() uint32` | Number of active map engines. A placeholder proving the mechanism works -- see below. |

This list is intentionally short. `d2interface`'s entire surface is *not*
exposed, on purpose (ROADMAP.md Phase 3: "Exposer une API minimale et
stable aux scripts"): every host function is a deliberate, narrow
decision, not a byproduct of convenience. Expect this table to grow as
real content (dialogue triggers, quest state, loot events) needs it --
see ROADMAP.md Phase 4/5 for what's planned.

## Known limitations (as of this writing)

- **No quest/dialogue API yet.** The functions needed to trigger a line of
  dialogue or advance a quest state (e.g. Sage Wyn's lines, ROADMAP.md
  Phase 4) don't exist yet -- only the loading mechanism does.
- **Numeric/memory types only.** Unlike the previous JS-based engine (which
  could pass arbitrary Go values into scripts via reflection), a WASM host
  function's parameters and results are limited to numeric types
  (`uint32`, `int64`, `float64`, ...). Passing structured data (a string, a
  list) means writing it into the guest's linear memory and passing a
  pointer/length pair -- there's no example of this yet in this codebase.
- **No save/load for script state.** Scripts are re-instantiated fresh each
  time `RunScript` runs; nothing persists script-side state across runs.
