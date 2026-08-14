// Package placeholdergen generates minimal, valid COF/DC6/palette/
// AnimData assets from scratch -- not extracted from any real Diablo II
// installation, and not derived from any third-party art. It exists
// because this development environment has no real Diablo II MPQ files
// to render against (confirmed while investigating the DC6/MPQ codec
// this session), and no image-generation tool is available either, but
// d2asset.Composite's full COF-compositing pipeline (the only path
// Player/NPC rendering ever goes through -- there is no simpler flat-
// sprite path) is otherwise tolerant of hand-authored, minimal data: a
// COF with exactly one layer and one direction is structurally legal to
// this codebase's own d2cof parser, and a palette is just 256 raw BGR
// triplets.
//
// Placeholder assets built here are deliberately plain (a solid color
// block, not humanoid pixel art): this environment can't visually verify
// what anything looks like, so the point is proving the pipeline renders
// *something* at all, not aiming for a specific look.
package placeholdergen

import (
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2enum"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2fileformats/d2animdata"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2fileformats/d2cof"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2fileformats/d2dc6"
)

// paletteSize is the fixed number of entries in a DC6/DAT palette --
// mirrors d2dat's own unexported numColors constant, which can't be
// imported from here.
const paletteSize = 256

// BuildPalette returns a 768-byte palette.dat (256 BGR triplets, matching
// d2common/d2fileformats/d2dat.Load's own byte layout). colors maps a
// palette index to an (R, G, B) triplet; index 0 is conventionally left
// black/unused since DC6 frames never draw it (it's the transparent
// index -- see d2dc6.DecodeFrame/EncodeFrame). Any index not present in
// colors stays black.
func BuildPalette(colors map[byte][3]byte) []byte {
	data := make([]byte, paletteSize*3)

	for idx, rgb := range colors {
		offset := int(idx) * 3
		data[offset] = rgb[2]   // B
		data[offset+1] = rgb[1] // G
		data[offset+2] = rgb[0] // R
	}

	return data
}

// BuildFlatColorDC6 builds a DC6 with framesPerDirection identical
// frames, each a solid colorIndex-filled width x height rectangle on a
// transparent (index 0) background. One direction only -- see
// BuildMinimalCOF's own doc comment for why that's structurally fine.
func BuildFlatColorDC6(width, height, framesPerDirection uint32, colorIndex byte) *d2dc6.DC6 {
	pixels := make([]byte, width*height)
	for i := range pixels {
		pixels[i] = colorIndex
	}

	encoded := d2dc6.EncodeFrame(pixels, width, height)

	dc6 := &d2dc6.DC6{
		Version:            6,
		Flags:              1,
		Encoding:           1,
		Termination:        []byte{0xEE, 0xEE, 0xEE, 0xEE},
		Directions:         1,
		FramesPerDirection: framesPerDirection,
		FramePointers:      make([]uint32, framesPerDirection),
		Frames:             make([]*d2dc6.DC6Frame, framesPerDirection),
	}

	for i := uint32(0); i < framesPerDirection; i++ {
		// FrameData is shared across frames deliberately -- every frame
		// looks identical (a static color block), so there's no reason to
		// repeat the encode.
		dc6.Frames[i] = &d2dc6.DC6Frame{
			Width:  width,
			Height: height,
			//nolint:gosec // a placeholder frame's encoded length never
			// remotely approaches uint32's range
			Length:     uint32(len(encoded)),
			FrameData:  encoded,
			Terminator: []byte{0xEE, 0xEE, 0xEE},
		}
	}

	return dc6
}

// BuildMinimalCOF builds a COF with exactly one layer (layerType), one
// direction, and framesPerDirection frames, using weaponClass for that
// single layer.
//
// ponytail: one layer/one direction is the floor this codebase's own
// d2cof parser and d2asset.Composite.createMode actually require (no
// fixed real-D2 layer count or direction count is enforced) -- this
// isn't attempting a faithful multi-layer/8-direction reproduction, just
// the minimum that renders as *something*.
func BuildMinimalCOF(layerType d2enum.CompositeType, weaponClass d2enum.WeaponClass, framesPerDirection, speed int) *d2cof.COF {
	cof := d2cof.New()
	cof.NumberOfDirections = 1
	cof.FramesPerDirection = framesPerDirection
	cof.NumberOfLayers = 1
	cof.Speed = speed
	cof.CofLayers = []d2cof.CofLayer{
		{
			Type:        layerType,
			WeaponClass: weaponClass,
		},
	}
	cof.CompositeLayers = map[d2enum.CompositeType]int{layerType: 0}
	cof.AnimationFrames = make([]d2enum.AnimationFrame, framesPerDirection)

	cof.Priority = make([][][]d2enum.CompositeType, 1)
	cof.Priority[0] = make([][]d2enum.CompositeType, framesPerDirection)

	for frame := 0; frame < framesPerDirection; frame++ {
		cof.Priority[0][frame] = []d2enum.CompositeType{layerType}
	}

	return cof
}

// BuildMinimalAnimData builds a synthetic AnimationData with a single
// entry, keyed exactly the way d2asset.Composite.createMode looks it up:
// strings.ToUpper(token + animationMode.String() + weaponClass.String()).
// Real gameplay never needs this -- a real Diablo II install's own
// AnimData.d2 already has entries for every stock hero token/mode/weapon
// combination -- this exists purely so this environment's own tests can
// exercise the full composite pipeline without any real D2 file at all.
func BuildMinimalAnimData(key string, framesPerDirection uint32, speed uint16) (*d2animdata.AnimationData, error) {
	animData := d2animdata.New()

	if err := animData.AddEntry(key); err != nil {
		return nil, err
	}

	animData.PushRecord(key)

	record := animData.GetRecord(key)
	record.SetFramesPerDirection(framesPerDirection)
	record.SetSpeed(speed)

	return animData, nil
}
