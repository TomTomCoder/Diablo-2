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

	dc6 := BuildFlatColorDC6(width, height, frames, 5)

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

// TestBuildFlatColorDC6SurvivesMarshalAndLoad proves the generated DC6 is
// a genuinely valid file, not just a valid in-memory struct: Marshal it
// to bytes and Load those bytes back.
func TestBuildFlatColorDC6SurvivesMarshalAndLoad(t *testing.T) {
	dc6 := BuildFlatColorDC6(2, 2, 1, 9)

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

func TestBuildMinimalCOFHasOneLayerOneDirection(t *testing.T) {
	cof := BuildMinimalCOF(d2enum.CompositeTypeTorso, d2enum.WeaponClassHandToHand, 3, 10)

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

// TestBuildMinimalCOFSurvivesMarshalAndUnmarshal proves the generated COF
// is a genuinely valid file: Marshal it and Unmarshal the bytes back via
// this codebase's own d2cof parser.
func TestBuildMinimalCOFSurvivesMarshalAndUnmarshal(t *testing.T) {
	original := BuildMinimalCOF(d2enum.CompositeTypeTorso, d2enum.WeaponClassHandToHand, 4, 10)

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
