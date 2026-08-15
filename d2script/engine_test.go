package d2script

import (
	"context"
	"testing"

	"github.com/tetratelabs/wazero/api"
)

// hello.wasm imports "devil"."recordCall" and "devil"."showText", and
// calls each once (42, and a fixed narration string respectively) -- see
// testdata/wasmfixture/main.go. This is a real, compiled WASM module, not
// a mock: the test only passes if the host<->guest wiring genuinely works.
const helloWasmPath = "testdata/hello.wasm"

// helloWasmNarrationText must match testdata/wasmfixture/main.go's own
// literal exactly -- there's no way to assert this generically, since
// it's the guest's own hardcoded string, not something the host chooses.
const helloWasmNarrationText = "Hello from Devil's narrative script"

func TestScriptEngineRunsGuestModuleAndCallsHostFunction(t *testing.T) {
	engine := CreateScriptEngine()
	defer func() {
		if err := engine.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	}()

	var recorded int32

	engine.AddFunction("recordCall", func(n int32) {
		recorded = n
	})
	engine.AddFunction("showText", func(_ context.Context, m api.Module, ptr, length uint32) {})

	if err := engine.RunScript(helloWasmPath); err != nil {
		t.Fatalf("RunScript: %v", err)
	}

	if recorded != 42 {
		t.Errorf("expected the guest module to call recordCall(42), got %d", recorded)
	}
}

// TestScriptEngineReadsStringFromGuestMemory is a real end-to-end proof of
// ReadString: the guest module (testdata/wasmfixture/main.go) writes a
// real Go string into its own linear memory and passes only a
// (pointer, length) pair across the ABI boundary -- the host function
// registered here must reconstruct the exact original string via
// ReadString, using the api.Module wazero hands it (see AddFunction's own
// doc comment on this kind of host function signature).
func TestScriptEngineReadsStringFromGuestMemory(t *testing.T) {
	engine := CreateScriptEngine()
	defer func() {
		if err := engine.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	}()

	var (
		narrated string
		readOk   bool
	)

	engine.AddFunction("recordCall", func(n int32) {})
	engine.AddFunction("showText", func(_ context.Context, m api.Module, ptr, length uint32) {
		narrated, readOk = ReadString(m, ptr, length)
	})

	if err := engine.RunScript(helloWasmPath); err != nil {
		t.Fatalf("RunScript: %v", err)
	}

	if !readOk {
		t.Fatal("expected ReadString to succeed on the guest's own valid pointer/length")
	}

	if narrated != helloWasmNarrationText {
		t.Errorf("expected the exact guest string %q, got %q", helloWasmNarrationText, narrated)
	}
}

func TestScriptEngineMissingFileErrors(t *testing.T) {
	engine := CreateScriptEngine()
	defer engine.Close() //nolint:errcheck // test cleanup

	if err := engine.RunScript("testdata/does-not-exist.wasm"); err == nil {
		t.Error("expected an error for a missing script file")
	}
}
