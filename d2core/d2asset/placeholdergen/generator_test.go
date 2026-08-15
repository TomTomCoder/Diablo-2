package placeholdergen

import (
	"testing"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2enum"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2fileformats/d2cof"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2fileformats/d2dc6"
)

func TestBuildPaletteEncodesBGRTriplets(t *testing.T) {
	data := BuildPalette(map[byte][3]byte{
		1: {255, 0, 0}, // red
		2: {0, 255, 0}, // green
	})

	if len(data) != paletteSize*3 {
		t.Fatalf("expected %d bytes, got %d", paletteSize*3, len(data))
	}

	// index 1: R=255,G=0,B=0 -> stored as B,G,R = 0,0,255
	if data[3] != 0 || data[4] != 0 || data[5] != 255 {
		t.Errorf("expected index 1 to be BGR(0,0,255), got (%d,%d,%d)", data[3], data[4], data[5])
	}

	// index 2: R=0,G=255,B=0 -> stored as B,G,R = 0,255,0
	if data[6] != 0 || data[7] != 255 || data[8] != 0 {
		t.Errorf("expected index 2 to be BGR(0,255,0), got (%d,%d,%d)", data[6], data[7], data[8])
	}

	// an untouched index (e.g. 0) stays black.
	if data[0] != 0 || data[1] != 0 || data[2] != 0 {
		t.Errorf("expected index 0 to stay black, got (%d,%d,%d)", data[0], data[1], data[2])
	}
}

func TestBuildFlatColorDC6DecodesToASolidRectangle(t *testing.T) {
	const width, height, frames = 4, 3, 2

	dc6 := BuildFlatColorDC6(width, height, frames, 1, 5)

	if dc6.Directions != 1 || dc6.FramesPerDirection != frames {
		t.Fatalf("expected 1 direction / %d frames, got %d/%d", frames, dc6.Directions, dc6.FramesPerDirection)
	}

	if len(dc6.Frames) != frames {
		t.Fatalf("expected %d frames, got %d", frames, len(dc6.Frames))
	}

	for f := 0; f < frames; f++ {
		pixels := dc6.DecodeFrame(f)
		if len(pixels) != width*height {
			t.Fatalf("frame %d: expected %d pixels, got %d", f, width*height, len(pixels))
		}

		for i, p := range pixels {
			if p != 5 {
				t.Errorf("frame %d, pixel %d: expected color index 5, got %d", f, i, p)
			}
		}
	}
}

// TestBuildFlatColorDC6MultipleDirections is a regression test:
// Directions used to be hardcoded to 1. With 8 directions and 3 frames
// each, the DC6 must carry 24 total frames, all still decodable.
func TestBuildFlatColorDC6MultipleDirections(t *testing.T) {
	const width, height, frames, directions = 4, 4, 3, 8

	dc6 := BuildFlatColorDC6(width, height, frames, directions, 7)

	if dc6.Directions != directions || dc6.FramesPerDirection != frames {
		t.Fatalf("expected %d directions / %d frames, got %d/%d", directions, frames, dc6.Directions, dc6.FramesPerDirection)
	}

	if len(dc6.Frames) != directions*frames {
		t.Fatalf("expected %d total frames, got %d", directions*frames, len(dc6.Frames))
	}

	for f := 0; f < directions*frames; f++ {
		pixels := dc6.DecodeFrame(f)
		for i, p := range pixels {
			if p != 7 {
				t.Errorf("frame %d, pixel %d: expected color index 7, got %d", f, i, p)
			}
		}
	}
}

// TestBuildFlatColorDC6SurvivesMarshalAndLoad proves the generated DC6 is
// a genuinely valid file, not just a valid in-memory struct: Marshal it
// to bytes and Load those bytes back.
func TestBuildFlatColorDC6SurvivesMarshalAndLoad(t *testing.T) {
	dc6 := BuildFlatColorDC6(2, 2, 1, 1, 9)

	loaded, err := d2dc6.Load(dc6.Marshal())
	if err != nil {
		t.Fatalf("Load(Marshal()) failed: %v", err)
	}

	pixels := loaded.DecodeFrame(0)
	for i, p := range pixels {
		if p != 9 {
			t.Errorf("pixel %d: expected color index 9, got %d", i, p)
		}
	}
}

func TestBuildMinimalCOFHasOneLayer(t *testing.T) {
	cof := BuildMinimalCOF(d2enum.CompositeTypeTorso, d2enum.WeaponClassHandToHand, 3, 1, 10)

	if cof.NumberOfLayers != 1 || cof.NumberOfDirections != 1 || cof.FramesPerDirection != 3 {
		t.Fatalf("expected 1 layer/1 direction/3 frames, got %d/%d/%d",
			cof.NumberOfLayers, cof.NumberOfDirections, cof.FramesPerDirection)
	}

	if len(cof.CofLayers) != 1 || cof.CofLayers[0].Type != d2enum.CompositeTypeTorso {
		t.Fatalf("expected a single Torso layer, got %v", cof.CofLayers)
	}

	if len(cof.Priority) != 1 || len(cof.Priority[0]) != 3 || len(cof.Priority[0][0]) != 1 {
		t.Fatalf("expected Priority shaped [1][3][1], got %v", cof.Priority)
	}
}

// TestBuildMinimalCOFMultipleDirections is a regression test for the same
// reason as TestBuildFlatColorDC6MultipleDirections: Priority must be
// shaped [directions][frames][1 layer], not hardcoded to one direction.
func TestBuildMinimalCOFMultipleDirections(t *testing.T) {
	const directions, frames = 8, 3

	cof := BuildMinimalCOF(d2enum.CompositeTypeTorso, d2enum.WeaponClassHandToHand, frames, directions, 10)

	if cof.NumberOfDirections != directions {
		t.Fatalf("expected %d directions, got %d", directions, cof.NumberOfDirections)
	}

	if len(cof.Priority) != directions {
		t.Fatalf("expected Priority to have %d directions, got %d", directions, len(cof.Priority))
	}

	for dir := 0; dir < directions; dir++ {
		if len(cof.Priority[dir]) != frames || len(cof.Priority[dir][0]) != 1 {
			t.Fatalf("direction %d: expected Priority shaped [%d][1], got %v", dir, frames, cof.Priority[dir])
		}
	}
}

// TestBuildMinimalCOFSurvivesMarshalAndUnmarshal proves the generated COF
// is a genuinely valid file: Marshal it and Unmarshal the bytes back via
// this codebase's own d2cof parser.
func TestBuildMinimalCOFSurvivesMarshalAndUnmarshal(t *testing.T) {
	original := BuildMinimalCOF(d2enum.CompositeTypeTorso, d2enum.WeaponClassHandToHand, 4, 1, 10)

	loaded, err := d2cof.Unmarshal(original.Marshal())
	if err != nil {
		t.Fatalf("Unmarshal(Marshal()) failed: %v", err)
	}

	if loaded.NumberOfLayers != 1 || loaded.NumberOfDirections != 1 || loaded.FramesPerDirection != 4 {
		t.Fatalf("expected 1 layer/1 direction/4 frames after round-trip, got %d/%d/%d",
			loaded.NumberOfLayers, loaded.NumberOfDirections, loaded.FramesPerDirection)
	}

	if len(loaded.CofLayers) != 1 || loaded.CofLayers[0].Type != d2enum.CompositeTypeTorso {
		t.Fatalf("expected a single Torso layer after round-trip, got %v", loaded.CofLayers)
	}

	if loaded.CofLayers[0].WeaponClass != d2enum.WeaponClassHandToHand {
		t.Errorf("expected WeaponClassHandToHand after round-trip, got %v", loaded.CofLayers[0].WeaponClass)
	}
}

func TestBuildMinimalAnimDataProducesARetrievableRecord(t *testing.T) {
	animData, err := BuildMinimalAnimData("SOTNHTH", 6, 256)
	if err != nil {
		t.Fatalf("BuildMinimalAnimData failed: %v", err)
	}

	record := animData.GetRecord("SOTNHTH")
	if record == nil {
		t.Fatal("expected a retrievable record for key SOTNHTH")
	}

	if record.FramesPerDirection() != 6 {
		t.Errorf("expected FramesPerDirection 6, got %d", record.FramesPerDirection())
	}
}
