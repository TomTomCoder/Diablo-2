package d2util

// Scale2x upscales a rectangular pixel grid 2x using the Scale2x (a.k.a.
// AdvMAME2x) algorithm: https://www.scale2x.it/algorithm -- public domain,
// no copyrighted art or invented content involved.
//
// ponytail: chosen over HQx/xBRZ (the two algorithms ROADMAP.md names as
// examples) because it decides each output pixel purely by comparing
// neighbour values for equality, never blends colors. DC6's DecodeFrame
// returns a flat grid of *palette indices*, not RGB -- an algorithm that
// blends would have to look up colors, blend, then re-quantize back to
// the nearest palette entry (lossy, and needs the palette threaded
// through). Scale2x needs none of that, so it runs directly and losslessly
// on the same []byte DecodeFrame already produces.
//
// src is row-major, width*height bytes, one byte per pixel (a palette
// index, or any other single-byte pixel format). Edge pixels are treated
// as bordered by copies of themselves (clamped), matching the reference
// algorithm's handling of image edges.
func Scale2x(src []byte, width, height int) (dst []byte, dstWidth, dstHeight int) {
	dstWidth, dstHeight = width*2, height*2
	dst = make([]byte, dstWidth*dstHeight)

	at := func(x, y int) byte {
		if x < 0 {
			x = 0
		}

		if x >= width {
			x = width - 1
		}

		if y < 0 {
			y = 0
		}

		if y >= height {
			y = height - 1
		}

		return src[y*width+x]
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			b := at(x, y-1)
			d := at(x-1, y)
			e := at(x, y)
			f := at(x+1, y)
			h := at(x, y+1)

			e0 := cornerPixel(d, b, h, f, d, e)
			e1 := cornerPixel(b, f, d, h, f, e)
			e2 := cornerPixel(d, h, b, f, d, e)
			e3 := cornerPixel(h, f, b, d, f, e)

			ox, oy := x*2, y*2
			dst[oy*dstWidth+ox] = e0
			dst[oy*dstWidth+ox+1] = e1
			dst[(oy+1)*dstWidth+ox] = e2
			dst[(oy+1)*dstWidth+ox+1] = e3
		}
	}

	return dst, dstWidth, dstHeight
}

// cornerPixel is Scale2x's shared corner-rounding rule: two neighbours
// agreeing with each other while each also disagrees with the far side of
// the other means the pixel between them is a diagonal edge, so the
// candidate takes over from e. Factored out of Scale2x itself purely to
// keep that loop's branching (4 independent instances of this same rule)
// under the cyclomatic-complexity lint threshold -- the four call sites
// below are the four corner formulas from the reference algorithm, just
// written positionally instead of spelled out with their own if-statements.
func cornerPixel(x, y, notForX, notForY, candidate, e byte) byte {
	if x == y && x != notForX && y != notForY {
		return candidate
	}

	return e
}
