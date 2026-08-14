package filesystem

import (
	"io"
	"os"
	"path/filepath"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2loader/asset"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2loader/asset/types"
)

// static check that Source implements AssetSource
var _ asset.Source = &Source{}

// Source represents an asset source which is a normal directory on the host file system
type Source struct {
	Root string
}

// Type returns the type of this asset source
func (s *Source) Type() types.SourceType {
	return types.AssetSourceFileSystem
}

// Open opens a file with the given sub-path within the Root dir of the file system source
func (s *Source) Open(subPath string) (io.ReadSeeker, error) {
	return os.Open(s.fullPath(subPath))
}

// Exists returns true if the file exists
//
// Correction (août 2026): this used to check os.IsExist(err), which
// classifies "already exists" errors from creation calls (os.Mkdir/
// os.OpenFile with O_EXCL) -- not whether os.Stat succeeded. os.Stat
// returns a nil error on success, and os.IsExist(nil) is always false,
// so this always reported an existing file as missing. Found while
// building a filesystem-backed asset source for generated placeholder
// sprites (no real Diablo II MPQ exists in this environment): every file
// written to disk was rejected as "not found" by the composite loader.
func (s *Source) Exists(subPath string) bool {
	_, err := os.Stat(s.fullPath(subPath))
	return err == nil
}

func (s *Source) fullPath(subPath string) string {
	return filepath.Clean(filepath.Join(s.Root, subPath))
}

// Path returns the Root dir of this file system source
func (s *Source) Path() string {
	return s.Root
}

// String returns the path
func (s *Source) String() string {
	return s.Path()
}
