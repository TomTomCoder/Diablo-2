package d2script

import "github.com/tetratelabs/wazero/api"

// ReadString reads a UTF-8 string of byteCount bytes at offset in m's
// linear memory -- the standard way a host function receives a string
// argument from a WASM guest module (guests can only pass numeric
// pointer/length pairs across the ABI boundary, not string values
// directly; see AddFunction's own doc comment). This is the first piece
// of Devil's own narration API (ROADMAP.md's "API de script pour la
// narration", Phase 3): any future host function a script calls to
// display narration/quest text will receive its string argument this way.
// Deliberately generic -- no narration/quest-specific host function is
// added here, since what a script should actually be able to trigger
// (and when) is still real narrative content this session doesn't
// invent (see ROADMAP.md).
//
// Returns ("", false) if the read is out of the guest's own memory
// bounds -- the same shape api.Memory.Read itself already reports,
// passed straight through rather than panicking on a malformed
// pointer/length pair from the guest.
func ReadString(m api.Module, offset, byteCount uint32) (string, bool) {
	buf, ok := m.Memory().Read(offset, byteCount)
	if !ok {
		return "", false
	}

	return string(buf), true
}
