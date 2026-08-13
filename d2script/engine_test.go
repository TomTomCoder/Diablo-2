package d2script

import "testing"

// hello.wasm imports "devil"."recordCall" and calls it once with 42 -- see
// testdata/wasmfixture/main.go. This is a real, compiled WASM module, not
// a mock: the test only passes if the host<->guest wiring genuinely works.
const helloWasmPath = "testdata/hello.wasm"

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

	if err := engine.RunScript(helloWasmPath); err != nil {
		t.Fatalf("RunScript: %v", err)
	}

	if recorded != 42 {
		t.Errorf("expected the guest module to call recordCall(42), got %d", recorded)
	}
}

func TestScriptEngineMissingFileErrors(t *testing.T) {
	engine := CreateScriptEngine()
	defer engine.Close() //nolint:errcheck // test cleanup

	if err := engine.RunScript("testdata/does-not-exist.wasm"); err == nil {
		t.Error("expected an error for a missing script file")
	}
}
