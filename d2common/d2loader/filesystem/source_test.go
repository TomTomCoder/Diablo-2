package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSourceExistsFindsARealFile is a regression test: Exists used to
// check os.IsExist(err) against os.Stat's result -- os.IsExist classifies
// "already exists" errors from creation calls (os.Mkdir/os.OpenFile with
// O_EXCL), not whether a stat succeeded. os.Stat returns a nil error on
// success, and os.IsExist(nil) is always false, so Exists always reported
// a real, present file as missing. Found while building a filesystem-
// backed asset source for generated placeholder sprites (no real Diablo
// II MPQ exists in this environment) -- every written file was rejected
// as "not found".
func TestSourceExistsFindsARealFile(t *testing.T) {
	root := t.TempDir()

	if err := os.WriteFile(filepath.Join(root, "present.txt"), []byte("hi"), 0o600); err != nil {
		t.Fatalf("test setup: WriteFile failed: %v", err)
	}

	source := &Source{Root: root}

	if !source.Exists("present.txt") {
		t.Error("expected Exists to find a file that's actually there")
	}
}

func TestSourceExistsFalseForAMissingFile(t *testing.T) {
	source := &Source{Root: t.TempDir()}

	if source.Exists("nope.txt") {
		t.Error("expected Exists to report a missing file as absent")
	}
}
