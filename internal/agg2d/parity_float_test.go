package agg2d

import "testing"

// parityTarget is the common method subset shared verbatim by *Agg2D and
// *Agg2DFloat, letting one scene drive both the 8-bit and float pipelines.
type parityTarget interface {
	ClearAll(Color)
	FillColor(Color)
	LineColor(Color)
	LineWidth(float64)
	ResetPath()
	MoveTo(x, y float64)
	LineTo(x, y float64)
	ClosePolygon()
	DrawPath(DrawPathFlag)
	Rectangle(x1, y1, x2, y2 float64)
	FillLinearGradient(x1, y1, x2, y2 float64, c1, c2 Color, profile float64)
}

var (
	_ parityTarget = (*Agg2D)(nil)
	_ parityTarget = (*Agg2DFloat)(nil)
)

// render8bit runs scene into a straight-RGBA8 buffer and returns it.
func render8bit(w, h int, scene func(parityTarget)) []uint8 {
	a := NewAgg2D()
	buf := make([]uint8, w*h*4)
	a.Attach(buf, w, h, w*4)
	scene(a)
	return buf
}

// renderFloat runs scene into a float image and returns it.
func renderFloat(w, h int, scene func(parityTarget)) *ImageFloat {
	a := NewAgg2DFloat()
	img := NewImageFloatEmpty(w, h)
	a.AttachImageFloat(img)
	scene(a)
	return img
}

// pixel8 reads a straight-RGBA8 pixel as [4]int.
func pixel8(buf []uint8, w, x, y int) [4]int {
	o := (y*w + x) * 4
	return [4]int{int(buf[o]), int(buf[o+1]), int(buf[o+2]), int(buf[o+3])}
}

// pixelFloatAsU8 reads a float pixel scaled to [0,255] as [4]int.
func pixelFloatAsU8(img *ImageFloat, x, y int) [4]int {
	c := img.GetPixel(x, y)
	r := func(v float32) int { return int(v*255 + 0.5) }
	return [4]int{r(c.R), r(c.G), r(c.B), r(c.A)}
}

func maxChanDiff(a, b [4]int) int {
	m := 0
	for i := range a {
		d := a[i] - b[i]
		if d < 0 {
			d = -d
		}
		if d > m {
			m = d
		}
	}
	return m
}

// TestParitySolidFill: an opaque solid fill must be pixel-identical (interior)
// between the 8-bit and float pipelines.
func TestParitySolidFill(t *testing.T) {
	const w, h = 24, 24
	scene := func(g parityTarget) {
		g.ClearAll(NewColor(0, 0, 0, 0))
		g.FillColor(NewColor(200, 100, 50, 255))
		g.ResetPath()
		g.MoveTo(4, 4)
		g.LineTo(20, 4)
		g.LineTo(20, 20)
		g.LineTo(4, 20)
		g.ClosePolygon()
		g.DrawPath(FillOnly)
	}
	buf := render8bit(w, h, scene)
	img := renderFloat(w, h, scene)

	for _, p := range [][2]int{{8, 8}, {12, 12}, {16, 16}, {10, 15}} {
		c8 := pixel8(buf, w, p[0], p[1])
		cf := pixelFloatAsU8(img, p[0], p[1])
		if d := maxChanDiff(c8, cf); d > 1 {
			t.Errorf("solid fill mismatch at (%d,%d): 8bit=%v float=%v maxdiff=%d", p[0], p[1], c8, cf, d)
		}
	}
}

// TestParityLinearGradient: a linear gradient fill must match within a small
// tolerance (8-bit LUT interpolation vs float LUT interpolation + rounding).
func TestParityLinearGradient(t *testing.T) {
	const w, h = 40, 16
	scene := func(g parityTarget) {
		g.ClearAll(NewColor(0, 0, 0, 0))
		g.FillLinearGradient(0, 0, 40, 0, NewColor(255, 0, 0, 255), NewColor(0, 0, 255, 255), 1.0)
		g.ResetPath()
		g.MoveTo(0, 0)
		g.LineTo(40, 0)
		g.LineTo(40, 16)
		g.LineTo(0, 16)
		g.ClosePolygon()
		g.DrawPath(FillOnly)
	}
	buf := render8bit(w, h, scene)
	img := renderFloat(w, h, scene)

	const tol = 3
	for _, x := range []int{5, 15, 25, 35} {
		c8 := pixel8(buf, w, x, 8)
		cf := pixelFloatAsU8(img, x, 8)
		if d := maxChanDiff(c8, cf); d > tol {
			t.Errorf("gradient mismatch at x=%d: 8bit=%v float=%v maxdiff=%d (tol=%d)", x, c8, cf, d, tol)
		}
	}
}

// TestParityStroke: a stroked shape must match the 8-bit pipeline within AA tol.
func TestParityStroke(t *testing.T) {
	const w, h = 24, 24
	scene := func(g parityTarget) {
		g.ClearAll(NewColor(0, 0, 0, 0))
		g.LineColor(NewColor(0, 200, 0, 255))
		g.FillColor(NewColor(0, 0, 0, 0))
		g.LineWidth(3)
		g.Rectangle(5, 5, 19, 19)
	}
	buf := render8bit(w, h, scene)
	img := renderFloat(w, h, scene)

	// Sample points on the stroked border.
	for _, p := range [][2]int{{5, 12}, {19, 12}, {12, 5}, {12, 19}} {
		c8 := pixel8(buf, w, p[0], p[1])
		cf := pixelFloatAsU8(img, p[0], p[1])
		if d := maxChanDiff(c8, cf); d > 2 {
			t.Errorf("stroke mismatch at (%d,%d): 8bit=%v float=%v maxdiff=%d", p[0], p[1], c8, cf, d)
		}
	}
}
