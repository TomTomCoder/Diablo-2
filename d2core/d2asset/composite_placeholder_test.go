package d2asset

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2enum"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2loader/asset/types"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2resource"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2util"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2asset/placeholdergen"
)

// TestLoadCompositeWithGeneratedPlaceholderAssets is, as far as a repo-wide
// search could confirm, the first test in this codebase to exercise the
// full real rendering pipeline end to end: COF -> layer DC6 -> palette ->
// AssetManager.LoadComposite -> Composite.SetMode -- without any real
// Diablo II file. Every byte comes from placeholdergen, built entirely by
// this session's own code (see that package's doc comment for why: no
// real D2 install exists in this environment, and no image-generation
// tool is available either).
//
// This doesn't prove what the result *looks like* (this environment has
// no display -- same limitation as every other client-rendering change
// this session), only that the pipeline accepts hand-authored minimal
// data without erroring, which nothing in this repo had verified before
// (see ROADMAP.md's own investigation: no test anywhere imports d2cof/
// d2dcc together with d2asset.Composite, and none calls LoadComposite).
func TestLoadCompositeWithGeneratedPlaceholderAssets(t *testing.T) {
	root := t.TempDir()

	// A real D2 install's own AnimData.d2 already has an entry for every
	// stock hero token/mode/weapon combination -- reusing an existing
	// stock token (Sorceress, "SO"/"SOR") here means the *real* game
	// would never need this generated AnimData at all. It exists purely
	// so this test can exercise the pipeline without any real D2 file.
	const (
		token       = "SO"
		layerType   = d2enum.CompositeTypeTorso
		weaponClass = d2enum.WeaponClassHandToHand
		mode        = d2enum.PlayerAnimationModeTownNeutral
		frames      = 4
		speed       = 10
	)

	palette := placeholdergen.BuildPalette(map[byte][3]byte{1: {200, 40, 40}})
	writePlaceholderFile(t, root, d2resource.PaletteUnits, palette)

	cof := placeholdergen.BuildMinimalCOF(layerType, weaponClass, frames, speed)
	cofPath := "/data/global/chars/" + token + "/COF/" + token + mode.String() + weaponClass.String() + ".COF"
	writePlaceholderFile(t, root, cofPath, cof.Marshal())

	dc6 := placeholdergen.BuildFlatColorDC6(16, 16, frames, 1)
	// layerValue defaults to "lit" whenever Composite has no equipment set
	// for this layer type -- see Composite.createMode's own fallback.
	dc6Path := "/data/global/chars/" + token + "/" + layerType.String() + "/" +
		token + layerType.String() + "lit" + mode.String() + weaponClass.String() + ".dc6"
	writePlaceholderFile(t, root, dc6Path, dc6.Marshal())

	// Composite.createMode looks this up via
	// strings.ToUpper(token+animationMode.String()+weaponClass) -- the key
	// must match that exactly, case included.
	animKey := strings.ToUpper(token + mode.String() + weaponClass.String())

	animData, err := placeholdergen.BuildMinimalAnimData(animKey, frames, speed)
	if err != nil {
		t.Fatalf("BuildMinimalAnimData failed: %v", err)
	}

	assetManager, err := NewAssetManager(d2util.LogLevelNone)
	if err != nil {
		t.Fatalf("NewAssetManager failed: %v", err)
	}

	if err = assetManager.AddSource(root, types.AssetSourceFileSystem); err != nil {
		t.Fatalf("AddSource failed: %v", err)
	}

	assetManager.Records.Animation.Data = animData

	composite, err := assetManager.LoadComposite(d2enum.ObjectTypePlayer, token, d2resource.PaletteUnits)
	if err != nil {
		t.Fatalf("LoadComposite failed: %v", err)
	}

	if err := composite.SetMode(mode, weaponClass.String()); err != nil {
		t.Fatalf("SetMode failed against generated placeholder assets: %v", err)
	}

	if composite.GetFrameCount() != frames {
		t.Errorf("expected GetFrameCount() %d, got %d", frames, composite.GetFrameCount())
	}
}

// writePlaceholderFile writes data at root+subPath, creating parent
// directories as needed. subPath uses forward slashes regardless of host
// OS, matching how d2asset's own path constants (d2resource.*) and
// Composite's fmt.Sprintf-built paths are always written.
func writePlaceholderFile(t *testing.T, root, subPath string, data []byte) {
	t.Helper()

	fullPath := filepath.Join(root, filepath.FromSlash(subPath))

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o750); err != nil {
		t.Fatalf("failed to create directories for %s: %v", fullPath, err)
	}

	if err := os.WriteFile(fullPath, data, 0o600); err != nil {
		t.Fatalf("failed to write %s: %v", fullPath, err)
	}
}
