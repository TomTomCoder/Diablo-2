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
//
// Covers 8 directions and three real player animation modes (Neutral,
// Walk, Attack1) rather than just one static pose -- a placeholder
// character still needs to actually move and attack, not just stand
// still, even with plain color-block art.
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
		directions  = 8
		frames      = 4
		speed       = 10
	)

	modes := []d2enum.PlayerAnimationMode{
		d2enum.PlayerAnimationModeTownNeutral,
		d2enum.PlayerAnimationModeWalk,
		d2enum.PlayerAnimationModeAttack1,
	}

	palette := placeholdergen.BuildPalette(map[byte][3]byte{1: {200, 40, 40}})
	writePlaceholderFile(t, root, d2resource.PaletteUnits, palette)

	assetManager, err := NewAssetManager(d2util.LogLevelNone)
	if err != nil {
		t.Fatalf("NewAssetManager failed: %v", err)
	}

	if err = assetManager.AddSource(root, types.AssetSourceFileSystem); err != nil {
		t.Fatalf("AddSource failed: %v", err)
	}

	// Records.Animation.Data holds every mode's entry -- a real Composite
	// swaps modes against the same shared AnimData.d2 the whole game
	// loaded once at startup, not a per-mode file.
	animData, err := placeholdergen.BuildMinimalAnimData(animKey(token, modes[0], weaponClass), frames, speed)
	if err != nil {
		t.Fatalf("BuildMinimalAnimData failed: %v", err)
	}

	for _, mode := range modes[1:] {
		if err = animData.AddEntry(animKey(token, mode, weaponClass)); err != nil {
			t.Fatalf("AddEntry failed: %v", err)
		}

		animData.PushRecord(animKey(token, mode, weaponClass))

		record := animData.GetRecord(animKey(token, mode, weaponClass))
		record.SetFramesPerDirection(frames)
		record.SetSpeed(speed)
	}

	assetManager.Records.Animation.Data = animData

	for _, mode := range modes {
		writeGeneratedMode(t, root, token, mode, weaponClass, layerType, directions, frames, speed)
	}

	composite, err := assetManager.LoadComposite(d2enum.ObjectTypePlayer, token, d2resource.PaletteUnits)
	if err != nil {
		t.Fatalf("LoadComposite failed: %v", err)
	}

	for _, mode := range modes {
		if err := composite.SetMode(mode, weaponClass.String()); err != nil {
			t.Fatalf("SetMode(%s) failed against generated placeholder assets: %v", mode.String(), err)
		}

		if composite.GetFrameCount() != frames {
			t.Errorf("mode %s: expected GetFrameCount() %d, got %d", mode.String(), frames, composite.GetFrameCount())
		}
	}
}

// animKey mirrors Composite.createMode's own lookup key exactly:
// strings.ToUpper(token + animationMode.String() + weaponClass).
func animKey(token string, mode d2enum.PlayerAnimationMode, weaponClass d2enum.WeaponClass) string {
	return strings.ToUpper(token + mode.String() + weaponClass.String())
}

// writeGeneratedMode builds and writes the COF + single-layer DC6 for one
// (token, mode, weaponClass) combination at the exact paths
// Composite.createMode/loadCompositeLayer expect.
func writeGeneratedMode(
	t *testing.T, root, token string, mode d2enum.PlayerAnimationMode, weaponClass d2enum.WeaponClass,
	layerType d2enum.CompositeType, directions, frames, speed int,
) {
	t.Helper()

	cof := placeholdergen.BuildMinimalCOF(layerType, weaponClass, frames, directions, speed)
	cofPath := "/data/global/chars/" + token + "/COF/" + token + mode.String() + weaponClass.String() + ".COF"
	writePlaceholderFile(t, root, cofPath, cof.Marshal())

	//nolint:gosec // frames/directions here are always small, fixed test constants
	dc6 := placeholdergen.BuildFlatColorDC6(16, 16, uint32(frames), uint32(directions), 1)
	// layerValue defaults to "lit" whenever Composite has no equipment set
	// for this layer type -- see Composite.createMode's own fallback.
	dc6Path := "/data/global/chars/" + token + "/" + layerType.String() + "/" +
		token + layerType.String() + "lit" + mode.String() + weaponClass.String() + ".dc6"
	writePlaceholderFile(t, root, dc6Path, dc6.Marshal())
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
