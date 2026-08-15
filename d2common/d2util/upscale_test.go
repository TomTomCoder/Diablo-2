package d2util

import "testing"

func TestScale2xUniformImageStaysUniform(t *testing.T) {
	src := []byte{7, 7, 7, 7, 7, 7} // 3x2, one solid color
	dst, w, h := Scale2x(src, 3, 2)

	if w != 6 || h != 4 {
		t.Fatalf("expected 6x4, got %dx%d", w, h)
	}

	for i, v := range dst {
		if v != 7 {
			t.Errorf("pixel %d: expected 7 (unchanged solid color), got %d", i, v)
		}
	}
}

func TestScale2xHorizontalStripReplicatesEachColumn(t *testing.T) {
	// A single row with no vertical neighbours to disagree with: Scale2x
	// degrades to simple nearest-neighbour column doubling here, since
	// every corner condition needs a genuine top/bottom difference to fire.
	src := []byte{1, 2, 3}
	dst, w, h := Scale2x(src, 3, 1)

	want := []byte{
		1, 1, 2, 2, 3, 3,
		1, 1, 2, 2, 3, 3,
	}

	if w != 6 || h != 2 {
		t.Fatalf("expected 6x2, got %dx%d", w, h)
	}

	for i, v := range dst {
		if v != want[i] {
			t.Errorf("pixel %d: expected %d, got %d", i, want[i], v)
		}
	}
}

func TestScale2xIsolatedPixelEnlargesToASolidBlockNotACross(t *testing.T) {
	// 3x3, background 0 with a single foreground pixel dead center. Every
	// one of Scale2x's four corner rules requires a same-side pair (e.g.
	// top==left) that disagrees with the opposite side to round a corner;
	// here every neighbour of the center pixel is background, so all four
	// rules stay silent and the pixel just doubles in place -- it must NOT
	// smear into a plus-shaped cross across the block boundary.
	src := []byte{
		0, 0, 0,
		0, 1, 0,
		0, 0, 0,
	}
	dst, w, h := Scale2x(src, 3, 3)

	if w != 6 || h != 6 {
		t.Fatalf("expected 6x6, got %dx%d", w, h)
	}

	total := 0
	for _, v := range dst {
		total += int(v)
	}

	if total != 4 {
		t.Fatalf("expected exactly 4 foreground pixels (one solid 2x2 block), got sum %d", total)
	}

	for _, pos := range []int{2*w + 2, 2*w + 3, 3*w + 2, 3*w + 3} {
		if dst[pos] != 1 {
			t.Errorf("expected the center 2x2 block (index %d) to be foreground, got %d", pos, dst[pos])
		}
	}
}
