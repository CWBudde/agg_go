package agg2d

import "testing"

// compositeFillScene clears to an opaque background, selects a blend mode, and
// fills an opaque rectangle. Opaque content keeps premultiplied == straight, so
// the composite pipeline's premultiplied buffer interpretation lines up with the
// straight clear/readback on both the 8-bit and float paths.
func compositeFillScene(mode BlendMode) func(parityTarget) {
	return func(g parityTarget) {
		g.ClearAll(NewColor(180, 120, 90, 255))
		g.SetBlendMode(mode)
		g.FillColor(NewColor(140, 200, 80, 255))
		g.ResetPath()
		g.MoveTo(4, 4)
		g.LineTo(28, 4)
		g.LineTo(28, 28)
		g.LineTo(4, 28)
		g.ClosePolygon()
		g.DrawPath(FillOnly)
	}
}

// TestParityCompositeFill checks the float composite pipeline matches the 8-bit
// composite pipeline for a representative set of blend modes (opaque content).
func TestParityCompositeFill(t *testing.T) {
	const w, h = 32, 32
	modes := []struct {
		name string
		mode BlendMode
	}{
		{"Multiply", BlendMultiply},
		{"Screen", BlendScreen},
		{"Darken", BlendDarken},
		{"Lighten", BlendLighten},
		{"Difference", BlendDifference},
		{"Exclusion", BlendExclusion},
		{"Overlay", BlendOverlay},
		{"HardLight", BlendHardLight},
		{"Plus", BlendAdd},
		{"SrcOver", BlendSrcOver},
	}

	samples := [][2]int{{8, 8}, {16, 16}, {20, 12}, {12, 24}}
	const tol = 2

	for _, m := range modes {
		scene := compositeFillScene(m.mode)
		buf := render8bit(w, h, scene)
		img := renderFloat(w, h, scene)
		for _, p := range samples {
			c8 := pixel8(buf, w, p[0], p[1])
			cf := pixelFloatAsU8(img, p[0], p[1])
			if d := maxChanDiff(c8, cf); d > tol {
				t.Errorf("%s mismatch at (%d,%d): 8bit=%v float=%v maxdiff=%d (tol=%d)",
					m.name, p[0], p[1], c8, cf, d, tol)
			}
		}
	}
}

// TestCompositeFloatBlendModeRoundTrip verifies the blend-mode state setters and
// that the composite operator actually changes output (Clear blanks the fill).
func TestCompositeFloatBlendModeRoundTrip(t *testing.T) {
	a := NewAgg2DFloat()
	img := NewImageFloatEmpty(16, 16)
	a.AttachImageFloat(img)

	if a.GetBlendMode() != BlendAlpha {
		t.Fatalf("default blend mode = %v, want BlendAlpha", a.GetBlendMode())
	}
	a.SetBlendMode(BlendMultiply)
	if a.GetBlendMode() != BlendMultiply {
		t.Fatalf("blend mode = %v after set, want BlendMultiply", a.GetBlendMode())
	}

	// Clear composite op must wipe the fill to transparent regardless of source.
	a.ClearAll(NewColor(200, 100, 50, 255))
	a.SetBlendMode(BlendClear)
	a.FillColor(NewColor(255, 255, 255, 255))
	a.ResetPath()
	a.MoveTo(2, 2)
	a.LineTo(14, 2)
	a.LineTo(14, 14)
	a.LineTo(2, 14)
	a.ClosePolygon()
	a.DrawPath(FillOnly)

	c := img.GetPixel(8, 8)
	if c.R != 0 || c.G != 0 || c.B != 0 || c.A != 0 {
		t.Fatalf("clear composite at (8,8) = %v, want zero", c)
	}
}

// TestCompositeFloatImageBlend verifies a transformed image honors a non-default
// blend mode by routing through the composite premultiplied renderer, matching
// the 8-bit path for opaque content.
func TestCompositeFloatImageBlend(t *testing.T) {
	const w, h = 48, 48
	src8 := makeSourceImage8(8, 8)
	srcF := NewImageFloatFromImage8(src8)

	a8 := NewAgg2D()
	dst8 := make([]uint8, w*h*4)
	a8.Attach(dst8, w, h, w*4)
	a8.ClearAll(NewColor(180, 120, 90, 255))
	a8.SetBlendMode(BlendMultiply)
	if err := a8.TransformImageSimple(src8, 8, 8, 40, 40); err != nil {
		t.Fatalf("8bit TransformImageSimple: %v", err)
	}

	aF := NewAgg2DFloat()
	imgF := NewImageFloatEmpty(w, h)
	aF.AttachImageFloat(imgF)
	aF.ClearAll(NewColor(180, 120, 90, 255))
	aF.SetBlendMode(BlendMultiply)
	if err := aF.TransformImageFloatSimple(srcF, 8, 8, 40, 40); err != nil {
		t.Fatalf("float TransformImageFloatSimple: %v", err)
	}

	const tol = 3
	for _, p := range [][2]int{{16, 16}, {24, 24}, {30, 20}} {
		c8 := pixel8(dst8, w, p[0], p[1])
		cf := pixelFloatAsU8(imgF, p[0], p[1])
		if d := maxChanDiff(c8, cf); d > tol {
			t.Errorf("composite image blend mismatch at (%d,%d): 8bit=%v float=%v maxdiff=%d (tol=%d)",
				p[0], p[1], c8, cf, d, tol)
		}
	}
}
