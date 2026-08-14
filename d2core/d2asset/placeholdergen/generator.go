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

// BuildFlatColorDC6 builds a DC6 with directions*framesPerDirection
// frames, each a solid colorIndex-filled width x height rectangle on a
// transparent (index 0) background. Frames are stored direction-major
// (all of direction 0's frames, then direction 1's, ...), matching real
// DC6 files -- the wrapping d2interface.Animation is what maps
// (direction, frame) to an index into this flat list.
//
// Correction (août 2026): directions used to be hardcoded to 1. A real
// character needs 8 (or 16/32/64) directions to look like anything but a
// single frozen pose from one angle; since every frame here is identical
// anyway (a flat color block), adding more directions costs nothing but
// repeating the same encoded bytes.
func BuildFlatColorDC6(width, height, framesPerDirection, directions uint32, colorIndex byte) *d2dc6.DC6 {
	pixels := make([]byte, width*height)
	for i := range pixels {
		pixels[i] = colorIndex
	}

	encoded := d2dc6.EncodeFrame(pixels, width, height)
	totalFrames := directions * framesPerDirection

	dc6 := &d2dc6.DC6{
		Version:            6,
		Flags:              1,
		Encoding:           1,
		Termination:        []byte{0xEE, 0xEE, 0xEE, 0xEE},
		Directions:         directions,
		FramesPerDirection: framesPerDirection,
		FramePointers:      make([]uint32, totalFrames),
		Frames:             make([]*d2dc6.DC6Frame, totalFrames),
	}

	for i := uint32(0); i < totalFrames; i++ {
		// FrameData is shared across every frame deliberately -- every
		// frame looks identical (a static color block) regardless of
		// direction or position in the cycle, so there's no reason to
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

// BuildMinimalCOF builds a COF with exactly one layer (layerType),
// `directions` directions, and framesPerDirection frames, using
// weaponClass for that single layer.
//
// ponytail: one layer is the floor this codebase's own d2cof parser and
// d2asset.Composite.createMode actually require (no fixed real-D2 layer
// count is enforced) -- this isn't attempting a faithful multi-layer
// reproduction, just the minimum that renders as *something*. Direction
// count, unlike layer count, is passed through rather than hardcoded --
// see BuildFlatColorDC6's own correction note for why 1 direction alone
// is too thin for real movement.
func BuildMinimalCOF(layerType d2enum.CompositeType, weaponClass d2enum.WeaponClass, framesPerDirection, directions, speed int) *d2cof.COF {
	cof := d2cof.New()
	cof.NumberOfDirections = directions
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

	cof.Priority = make([][][]d2enum.CompositeType, directions)

	for dir := 0; dir < directions; dir++ {
		cof.Priority[dir] = make([][]d2enum.CompositeType, framesPerDirection)

		for frame := 0; frame < framesPerDirection; frame++ {
			cof.Priority[dir][frame] = []d2enum.CompositeType{layerType}
		}
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
