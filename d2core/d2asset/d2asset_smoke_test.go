package d2asset

import (
	"testing"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2util"
)

// TestNewAssetManagerSmoke is the project's smoke test: it boots the asset
// manager (loader + record manager + caches) the same way d2app.Create does,
// without opening a window or reading MPQ game files, and fails if that
// wiring panics or errors.
//
// ponytail: this is as close to "run the engine headlessly" as CI can get
// without the user's own copyrighted Diablo II data files and a display
// server (ebiten needs a real/virtual GPU surface). Exercising the full
// render loop end-to-end still has to be done manually against real game
// files — see ROADMAP.md Phase 0.
func TestNewAssetManagerSmoke(t *testing.T) {
	for _, level := range []d2util.LogLevel{d2util.LogLevelDefault, d2util.LogLevelUnspecified - 1} {
		manager, err := NewAssetManager(level)
		if err != nil {
			t.Fatalf("NewAssetManager(%v) returned error: %v", level, err)
		}

		if manager == nil {
			t.Fatalf("NewAssetManager(%v) returned a nil manager with no error", level)
		}

		if manager.Records == nil {
			t.Errorf("NewAssetManager(%v): Records manager not initialized", level)
		}

		if manager.Loader == nil {
			t.Errorf("NewAssetManager(%v): Loader not initialized", level)
		}
	}
}
