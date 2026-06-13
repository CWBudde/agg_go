package agg2d

import "testing"

// stateAccessorScene is the method subset exercised by the state-accessor /
// RGBA-setter / alias parity tests. Every one of these manipulates or reads
// color-agnostic style state (colors, caps/joins, miter limit, clip box, image
// filter/resample, fill rule, gamma), so the float twin must produce results
// identical to the 8-bit oracle after the same operations.
type stateAccessorScene interface {
	// RGBA convenience setters.
	FillColorRGBA(r, g, b, a uint8)
	LineColorRGBA(r, g, b, a uint8)
	ClearAllRGBA(r, g, b, a uint8)
	ClearClipBoxRGBA(r, g, b, a uint8)

	// Getters.
	GetFillColor() Color
	GetLineColor() Color
	GetLineCap() LineCap
	GetLineJoin() LineJoin
	GetMiterLimit() float64
	GetClipBox() (x1, y1, x2, y2 float64)
	GetImageFilter() ImageFilter
	GetImageResample() ImageResample
	GetAntiAliasGamma() float64

	// Setters that the getters read back.
	LineCap(LineCap)
	LineJoin(LineJoin)
	MiterLimit(ml float64)
	ImageFilter(f ImageFilter)
	ImageResample(r ImageResample)
	SetImageFilterRadius(f ImageFilter, radius float64)
	SetAntiAliasGamma(gamma float64)
	ClipBox(x1, y1, x2, y2 float64)

	// Fill-rule queries.
	FillEvenOdd(bool)
	IsEvenOddFillRule() bool
	IsNonZeroFillRule() bool
	FillRuleDescription() string

	// Whole-style reset and clip clear.
	ResetStyle()
	ClearClipBox(c Color)
}

var (
	_ stateAccessorScene = (*Agg2D)(nil)
	_ stateAccessorScene = (*Agg2DFloat)(nil)
)

// applyStateAccessorSeq drives a fixed sequence of style mutations and reads
// back every accessor, returning a comparable snapshot of the style state.
type stateSnapshot struct {
	fill, line     Color
	cap            LineCap
	join           LineJoin
	miter          float64
	clipX1, clipY1 float64
	clipX2, clipY2 float64
	filter         ImageFilter
	resample       ImageResample
	gamma          float64
	evenOdd        bool
	nonZero        bool
	ruleDesc       string
}

func applyStateAccessorSeq(s stateAccessorScene) stateSnapshot {
	s.FillColorRGBA(10, 20, 30, 40)
	s.LineColorRGBA(50, 60, 70, 80)
	s.LineCap(CapSquare)
	s.LineJoin(JoinBevel)
	s.MiterLimit(7.5)
	s.ClipBox(5, 6, 95, 96)
	s.ImageFilter(Bicubic)
	s.ImageResample(ResampleAlways)
	s.SetAntiAliasGamma(1.7)
	s.FillEvenOdd(true)

	x1, y1, x2, y2 := s.GetClipBox()
	return stateSnapshot{
		fill:     s.GetFillColor(),
		line:     s.GetLineColor(),
		cap:      s.GetLineCap(),
		join:     s.GetLineJoin(),
		miter:    s.GetMiterLimit(),
		clipX1:   x1,
		clipY1:   y1,
		clipX2:   x2,
		clipY2:   y2,
		filter:   s.GetImageFilter(),
		resample: s.GetImageResample(),
		gamma:    s.GetAntiAliasGamma(),
		evenOdd:  s.IsEvenOddFillRule(),
		nonZero:  s.IsNonZeroFillRule(),
		ruleDesc: s.FillRuleDescription(),
	}
}

func TestParityStateAccessorReadback(t *testing.T) {
	a8 := NewAgg2D()
	af := NewAgg2DFloat()

	s8 := applyStateAccessorSeq(a8)
	sF := applyStateAccessorSeq(af)

	if s8 != sF {
		t.Errorf("state snapshot mismatch:\n  8bit = %+v\n float = %+v", s8, sF)
	}
}

func TestParityStateAccessorFloatValues(t *testing.T) {
	// Independent of the 8-bit oracle: verify the float twin reads back exactly
	// what was set (the 8-bit parity test already pins these to AGG semantics).
	af := NewAgg2DFloat()
	af.FillColorRGBA(1, 2, 3, 4)
	if got := af.GetFillColor(); got != (Color{1, 2, 3, 4}) {
		t.Errorf("GetFillColor = %v, want {1 2 3 4}", got)
	}
	af.LineColorRGBA(5, 6, 7, 8)
	if got := af.GetLineColor(); got != (Color{5, 6, 7, 8}) {
		t.Errorf("GetLineColor = %v, want {5 6 7 8}", got)
	}
	af.MiterLimit(3.25)
	if got := af.GetMiterLimit(); got != 3.25 {
		t.Errorf("GetMiterLimit = %v, want 3.25", got)
	}
	if !af.IsNonZeroFillRule() || af.IsEvenOddFillRule() {
		t.Error("default fill rule should be non-zero")
	}
	af.FillEvenOdd(true)
	if !af.IsEvenOddFillRule() || af.IsNonZeroFillRule() {
		t.Error("after FillEvenOdd(true) fill rule should be even-odd")
	}

	// SetImageFilterRadius for a radius-based filter must update the filter id.
	af.SetImageFilterRadius(Lanczos, 3.0)
	if got := af.GetImageFilter(); got != Lanczos {
		t.Errorf("GetImageFilter after SetImageFilterRadius = %v, want Lanczos", got)
	}
}

// TestParityResetStyle verifies ResetStyle returns style to defaults identically
// for both pipelines.
func TestParityResetStyle(t *testing.T) {
	reset := func(s stateAccessorScene) stateSnapshot {
		// Dirty the style first, then reset.
		s.FillColorRGBA(1, 1, 1, 1)
		s.LineColorRGBA(2, 2, 2, 2)
		s.LineCap(CapSquare)
		s.LineJoin(JoinBevel)
		s.MiterLimit(9)
		s.FillEvenOdd(true)
		s.ResetStyle()
		return stateSnapshot{
			fill:     s.GetFillColor(),
			line:     s.GetLineColor(),
			cap:      s.GetLineCap(),
			join:     s.GetLineJoin(),
			miter:    s.GetMiterLimit(),
			evenOdd:  s.IsEvenOddFillRule(),
			nonZero:  s.IsNonZeroFillRule(),
			ruleDesc: s.FillRuleDescription(),
		}
	}

	if r8, rF := reset(NewAgg2D()), reset(NewAgg2DFloat()); r8 != rF {
		t.Errorf("ResetStyle mismatch:\n  8bit = %+v\n float = %+v", r8, rF)
	}
}

// TestParityClearClipBoxRender verifies ClearClipBoxRGBA paints the buffer with
// the requested color identically in both pipelines.
func TestParityClearClipBoxRender(t *testing.T) {
	const w, h = 16, 16
	a8 := NewAgg2D()
	buf := make([]uint8, w*h*4)
	a8.Attach(buf, w, h, w*4)
	a8.ClearClipBoxRGBA(200, 100, 50, 255)

	af := NewAgg2DFloat()
	img := NewImageFloatEmpty(w, h)
	af.AttachImageFloat(img)
	af.ClearClipBoxRGBA(200, 100, 50, 255)

	for y := 0; y < h; y += 5 {
		for x := 0; x < w; x += 5 {
			p8 := pixel8(buf, w, x, y)
			pF := pixelFloatAsU8(img, x, y)
			if d := maxChanDiff(p8, pF); d > 2 {
				t.Errorf("ClearClipBox pixel (%d,%d): 8bit=%v float=%v diff=%d", x, y, p8, pF, d)
			}
		}
	}
}
