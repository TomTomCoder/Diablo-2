package d2dc6

import (
	"testing"
)

func TestDC6New(t *testing.T) {
	dc6 := New()

	if dc6 == nil {
		t.Error("d2dc6.New() method returned nil")
	}
}

func getExampleDC6() *DC6 {
	exampleDC6 := &DC6{
		Version:            6,
		Flags:              1,
		Encoding:           0,
		Termination:        []byte{238, 238, 238, 238},
		Directions:         1,
		FramesPerDirection: 1,
		FramePointers:      []uint32{56},
		Frames: []*DC6Frame{
			{
				Flipped:    0,
				Width:      32,
				Height:     26,
				OffsetX:    45,
				OffsetY:    24,
				Unknown:    0,
				NextBlock:  50,
				Length:     10,
				FrameData:  []byte{2, 23, 34, 128, 53, 64, 39, 43, 123, 12},
				Terminator: []byte{2, 8, 5},
			},
		},
	}

	return exampleDC6
}

func TestDC6Unmarshal(t *testing.T) {
	exampleDC6 := getExampleDC6()

	data := exampleDC6.Marshal()

	extractedDC6, err := Load(data)
	if err != nil {
		t.Error(err)
	}

	if exampleDC6.Version != extractedDC6.Version ||
		len(exampleDC6.Frames) != len(extractedDC6.Frames) ||
		exampleDC6.Frames[0].NextBlock != extractedDC6.Frames[0].NextBlock {
		t.Fatal("encoded and decoded DC6 isn't the same")
	}
}

// TestEncodeFrameRoundTripsThroughDecodeFrame is the core correctness
// check for EncodeFrame: build a small indexed image mixing transparent
// (0) and opaque pixels, encode it, decode it back via the DC6's own
// DecodeFrame, and confirm the pixels survive exactly.
func TestEncodeFrameRoundTripsThroughDecodeFrame(t *testing.T) {
	const width, height = 5, 3

	pixels := []byte{
		0, 1, 2, 0, 0,
		3, 3, 3, 3, 3,
		0, 0, 5, 0, 0,
	}

	encoded := EncodeFrame(pixels, width, height)

	dc6 := &DC6{
		Directions:         1,
		FramesPerDirection: 1,
		Frames: []*DC6Frame{
			{Width: width, Height: height, FrameData: encoded},
		},
	}

	decoded := dc6.DecodeFrame(0)

	if len(decoded) != len(pixels) {
		t.Fatalf("expected %d decoded pixels, got %d", len(pixels), len(decoded))
	}

	for i := range pixels {
		if decoded[i] != pixels[i] {
			t.Errorf("pixel %d: expected %d, got %d", i, pixels[i], decoded[i])
		}
	}
}

// TestEncodeFrameAllTransparent covers a row with no opaque pixels at all
// (a single transparent run per row, no opaque-run byte ever emitted).
func TestEncodeFrameAllTransparent(t *testing.T) {
	const width, height = 4, 2
	pixels := make([]byte, width*height)

	encoded := EncodeFrame(pixels, width, height)

	dc6 := &DC6{Frames: []*DC6Frame{{Width: width, Height: height, FrameData: encoded}}}
	decoded := dc6.DecodeFrame(0)

	for i, p := range decoded {
		if p != 0 {
			t.Errorf("pixel %d: expected transparent (0), got %d", i, p)
		}
	}
}

// TestEncodeFrameRunLongerThanMaxRunLength covers a row whose opaque run
// exceeds maxRunLength (0x7f) -- EncodeFrame must split it into multiple
// runs rather than overflowing a single run-length byte.
func TestEncodeFrameRunLongerThanMaxRunLength(t *testing.T) {
	const width, height = 200, 1
	pixels := make([]byte, width*height)

	for i := range pixels {
		pixels[i] = 7 // all opaque, same color, one long run
	}

	encoded := EncodeFrame(pixels, width, height)

	dc6 := &DC6{Frames: []*DC6Frame{{Width: width, Height: height, FrameData: encoded}}}
	decoded := dc6.DecodeFrame(0)

	if len(decoded) != len(pixels) {
		t.Fatalf("expected %d decoded pixels, got %d", len(pixels), len(decoded))
	}

	for i, p := range decoded {
		if p != 7 {
			t.Errorf("pixel %d: expected 7, got %d", i, p)
		}
	}
}

// TestEncodeFrameSurvivesFullFileMarshalAndLoad goes one step further than
// the DecodeFrame-only round-trip above: build a complete *DC6 (matching
// New()/Load()'s real header fields, not just a bare Frames slice),
// Marshal it to actual file bytes, Load those bytes back, and decode --
// proving an EncodeFrame-built frame survives being written to and read
// from a real .DC6 file, not just held in memory.
func TestEncodeFrameSurvivesFullFileMarshalAndLoad(t *testing.T) {
	const width, height = 3, 2

	pixels := []byte{
		0, 9, 0,
		4, 4, 4,
	}

	encoded := EncodeFrame(pixels, width, height)

	original := &DC6{
		Version:            6,
		Flags:              1,
		Encoding:           1,
		Termination:        []byte{238, 238, 238, 238},
		Directions:         1,
		FramesPerDirection: 1,
		FramePointers:      []uint32{0},
		Frames: []*DC6Frame{
			{
				Width:      width,
				Height:     height,
				Length:     uint32(len(encoded)),
				FrameData:  encoded,
				Terminator: []byte{238, 238, 238},
			},
		},
	}

	loaded, err := Load(original.Marshal())
	if err != nil {
		t.Fatalf("Load(Marshal()) failed: %v", err)
	}

	decoded := loaded.DecodeFrame(0)
	if len(decoded) != len(pixels) {
		t.Fatalf("expected %d decoded pixels, got %d", len(pixels), len(decoded))
	}

	for i := range pixels {
		if decoded[i] != pixels[i] {
			t.Errorf("pixel %d: expected %d, got %d", i, pixels[i], decoded[i])
		}
	}
}

func TestDC6Clone(t *testing.T) {
	exampleDC6 := getExampleDC6()
	clonedDC6 := exampleDC6.Clone()

	if exampleDC6.Termination[0] != clonedDC6.Termination[0] ||
		len(exampleDC6.Frames) != len(clonedDC6.Frames) ||
		exampleDC6.Frames[0].NextBlock != clonedDC6.Frames[0].NextBlock {
		t.Fatal("cloned dc6 isn't equal to original")
	}
}
