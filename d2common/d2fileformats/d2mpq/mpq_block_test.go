package d2mpq

import "testing"

// TestCalculateEncryptionSeedPlainEncryptedUsesFilenameHashOnly is a
// regression test: calculateEncryptionSeed used to only ever be called for
// FileFixKey blocks (see CreateStream) -- a plain FileEncrypted block (no
// FileFixKey, the common case for real-world MPQ archives, confirmed by
// reading a real third-party archive where every single file hit this)
// never got a seed calculated at all, leaving EncryptionSeed at its zero
// value and making every read of such a file fail. The fix is at both call
// sites (CreateStream now calls this for FileEncrypted, not just
// FileFixKey) and here: a plain encrypted block's seed is the filename
// hash, unmodified.
func TestCalculateEncryptionSeedPlainEncryptedUsesFilenameHashOnly(t *testing.T) {
	block := &Block{
		Flags:                FileEncrypted,
		FilePosition:         12345,
		UncompressedFileSize: 6789,
	}

	block.calculateEncryptionSeed(`data\global\items\amass.dc6`)

	want := hashString("amass.dc6", 3)
	if block.EncryptionSeed != want {
		t.Errorf("expected EncryptionSeed %d (plain filename hash), got %d", want, block.EncryptionSeed)
	}
}

// TestCalculateEncryptionSeedFixKeyAdjustsByPositionAndSize covers the
// other half: FileFixKey additionally folds in FilePosition/
// UncompressedFileSize, so its seed must differ from the plain case above
// even for the same filename.
func TestCalculateEncryptionSeedFixKeyAdjustsByPositionAndSize(t *testing.T) {
	block := &Block{
		Flags:                FileEncrypted | FileFixKey,
		FilePosition:         12345,
		UncompressedFileSize: 6789,
	}

	block.calculateEncryptionSeed(`data\global\items\amass.dc6`)

	baseHash := hashString("amass.dc6", 3)
	want := (baseHash + block.FilePosition) ^ block.UncompressedFileSize

	if block.EncryptionSeed != want {
		t.Errorf("expected EncryptionSeed %d (position/size-adjusted), got %d", want, block.EncryptionSeed)
	}

	if block.EncryptionSeed == baseHash {
		t.Error("expected FileFixKey's seed to differ from the plain filename hash")
	}
}

// TestCalculateEncryptionSeedStripsDirectoryFromFilename confirms only the
// base filename (after the last "\") feeds the hash, matching real MPQ
// paths using backslashes.
func TestCalculateEncryptionSeedStripsDirectoryFromFilename(t *testing.T) {
	withPath := &Block{Flags: FileEncrypted}
	withPath.calculateEncryptionSeed(`data\global\items\amass.dc6`)

	baseNameOnly := &Block{Flags: FileEncrypted}
	baseNameOnly.calculateEncryptionSeed(`amass.dc6`)

	if withPath.EncryptionSeed != baseNameOnly.EncryptionSeed {
		t.Errorf("expected the directory prefix to be stripped before hashing, got %d vs %d",
			withPath.EncryptionSeed, baseNameOnly.EncryptionSeed)
	}
}
