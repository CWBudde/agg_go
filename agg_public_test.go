package agg

import (
	"math"
	"testing"
)

func TestAgg2DPublicWrappers(t *testing.T) {
	a := NewAgg2D()
	buf := make([]uint8, 16*16*4)
	a.Attach(buf, 16, 16, 16*4)

	x1, y1, x2, y2 := 1.0, 2.0, 10.0, 12.0
	a.ClipBox(x1, y1, x2, y2)
	gx1, gy1, gx2, gy2 := a.GetClipBox()
	if gx1 != x1 || gy1 != y1 || gx2 != x2 || gy2 != y2 {
		t.Fatalf("GetClipBox() = (%v, %v, %v, %v), want (%v, %v, %v, %v)", gx1, gy1, gx2, gy2, x1, y1, x2, y2)
	}

	tr := Translation(3, 4)
	a.SetTransformations(tr)
	got := a.GetTransformations()
	if got == nil || got.AffineMatrix != tr.AffineMatrix {
		t.Fatalf("GetTransformations() = %#v, want %#v", got, tr)
	}

	a.PushTransform()
	a.Translate(5, 6)
	if !a.PopTransform() {
		t.Fatal("PopTransform() = false, want true")
	}
	got = a.GetTransformations()
	if got == nil || got.AffineMatrix != tr.AffineMatrix {
		t.Fatalf("transform after PopTransform() = %#v, want %#v", got, tr)
	}

	a.LineCap(CapSquare)
	if a.GetLineCap() != CapSquare {
		t.Fatalf("GetLineCap() = %v, want %v", a.GetLineCap(), CapSquare)
	}
	a.LineJoin(JoinBevel)
	if a.GetLineJoin() != JoinBevel {
		t.Fatalf("GetLineJoin() = %v, want %v", a.GetLineJoin(), JoinBevel)
	}

	a.ImageBlendMode(BlendMultiply)
	if a.GetImageBlendMode() != BlendMultiply {
		t.Fatalf("GetImageBlendMode() = %v, want %v", a.GetImageBlendMode(), BlendMultiply)
	}
	a.ImageFilter(NoFilter)
	if a.GetImageFilter() != NoFilter {
		t.Fatalf("GetImageFilter() = %v, want %v", a.GetImageFilter(), NoFilter)
	}
	a.ImageFilter(Bicubic)
	if a.GetImageFilter() != Bicubic {
		t.Fatalf("GetImageFilter() = %v, want %v", a.GetImageFilter(), Bicubic)
	}
	a.AffineImageResamplePolicy(AffineImageResamplePreferFiltered)
	if a.GetAffineImageResamplePolicy() != AffineImageResamplePreferFiltered {
		t.Fatalf("GetAffineImageResamplePolicy() = %v, want %v", a.GetAffineImageResamplePolicy(), AffineImageResamplePreferFiltered)
	}
	a.FillColorRGBA(90, 91, 92, 93)
	if gotColor := a.GetFillColor(); gotColor != (Color{R: 90, G: 91, B: 92, A: 93}) {
		t.Fatalf("GetFillColor() = %#v", gotColor)
	}
	a.LineColorRGBA(40, 41, 42, 43)
	if gotColor := a.GetLineColor(); gotColor != (Color{R: 40, G: 41, B: 42, A: 43}) {
		t.Fatalf("GetLineColor() = %#v", gotColor)
	}
	blendColor := Color{R: 10, G: 20, B: 30, A: 40}
	a.ImageBlendColor(blendColor)
	if gotColor := a.GetImageBlendColor(); gotColor != blendColor {
		t.Fatalf("GetImageBlendColor() = %#v, want %#v", gotColor, blendColor)
	}
	a.ImageBlendColorRGBA(11, 21, 31, 41)
	if gotColor := a.GetImageBlendColor(); gotColor != (Color{R: 11, G: 21, B: 31, A: 41}) {
		t.Fatalf("ImageBlendColorRGBA() stored %#v", gotColor)
	}

	a.AntiAliasGamma(2.0)
	if gotGamma := a.GetAntiAliasGamma(); gotGamma != 2.0 {
		t.Fatalf("GetAntiAliasGamma() = %v, want 2.0", gotGamma)
	}

	a.ClearAllRGBA(1, 2, 3, 4)
	if got := buf[:4]; got[0] != 1 || got[1] != 2 || got[2] != 3 || got[3] != 4 {
		t.Fatalf("ClearAllRGBA() pixel = %#v, want [1 2 3 4]", got)
	}

	a.MoveTo(1, 1)
	a.MoveRel(1, 0)
	a.LineTo(4, 4)
	a.HorLineTo(6)
	a.VerLineTo(7)
	a.ArcTo(2, 2, 0.5, false, true, 8, 8)
	a.QuadricCurveTo(8, 9, 10, 11)
	a.QuadricCurveRel(1, 1, 2, 2)
	a.QuadricCurveToSmooth(12, 13)
	a.QuadricCurveRelSmooth(1, 1)
	a.CubicCurveTo(1, 2, 3, 4, 5, 6)
	a.CubicCurveRel(1, 1, 2, 2, 3, 3)
	a.CubicCurveToSmooth(7, 8, 9, 10)
	a.CubicCurveRelSmooth(1, 2, 3, 4)
	a.DrawPathNoTransform(StrokeOnly)

	a.Triangle(1, 1, 5, 1, 3, 4)
	a.RoundedRectXY(1, 1, 8, 8, 2, 3)
	a.RoundedRectVariableRadii(1, 1, 8, 8, 2, 2, 3, 3)
	a.Arc(5, 5, 3, 2, 0, 1.5)
	a.Star(5, 5, 2, 4, 0.3, 5)
	a.Curve(0, 0, 2, 3, 4, 5)
	a.Curve4(0, 0, 2, 3, 4, 5, 6, 7)
	a.Polygon([]float64{1, 1, 4, 1, 4, 4}, 3)
	a.Polyline([]float64{1, 1, 2, 2, 3, 3}, 3)

	a.FillRadialGradient(2, 2, 3, Red, Blue, 1.0)
	a.FillRadialGradientPos(4, 5, 6)
	if got := a.FillGradientD2(); math.Abs(got-6) > 1e-9 {
		t.Fatalf("FillRadialGradientPos() radius = %v, want 6", got)
	}

	a.LineRadialGradient(2, 2, 3, Red, Blue, 1.0)
	a.LineRadialGradientPos(7, 8, 9)
	if got := a.LineGradientD2(); math.Abs(got-9) > 1e-9 {
		t.Fatalf("LineRadialGradientPos() radius = %v, want 9", got)
	}

	a.ResetTransformations()
	a.Parallelogram(10, 20, 14, 20, 10, 26)
	got = a.GetTransformations()
	want := [6]float64{4, 0, 0, 6, 10, 20}
	if got == nil || got.AffineMatrix != want {
		t.Fatalf("Parallelogram() = %#v, want %#v", got, want)
	}
}

func TestAgg2DCompatibilityShims(t *testing.T) {
	a := NewAgg2D()
	buf := make([]uint8, 16*16*4)
	a.Attach(buf, 16, 16, 16*4)

	a.ClipBox(2, 3, 12, 14)
	if got := a.GetClipBoxRect(); got != (RectD{X1: 2, Y1: 3, X2: 12, Y2: 14}) {
		t.Fatalf("GetClipBoxRect() = %#v", got)
	}

	a.ResetTransformations()
	a.ParallelogramFromRect(10, 20, 14, 26, []float64{10, 20, 14, 20, 10, 26})
	got := a.GetTransformations()
	want := [6]float64{1, 0, 0, 1, 0, 0}
	if got == nil || got.AffineMatrix != want {
		t.Fatalf("ParallelogramFromRect(identity) = %#v, want %#v", got, want)
	}

	if Deg2RadFunc(180) != math.Pi {
		t.Fatalf("Deg2RadFunc(180) = %v", Deg2RadFunc(180))
	}
	if math.Abs(Rad2DegFunc(math.Pi)-180.0) > 1e-12 {
		t.Fatalf("Rad2DegFunc(pi) = %v", Rad2DegFunc(math.Pi))
	}

	imgBuf := []uint8{100, 50, 25, 128}
	img := Image{}
	img.Attach(imgBuf, 1, 1, 4)
	if img.Width() != 1 || img.Height() != 1 || img.Stride() != 4 {
		t.Fatalf("zero-value Image.Attach() produced invalid image: %dx%d stride=%d", img.Width(), img.Height(), img.Stride())
	}
	if err := img.Premultiply(); err != nil {
		t.Fatalf("Premultiply() failed: %v", err)
	}
	if got := img.Data; got[0] != 50 || got[1] != 25 || got[2] != 12 || got[3] != 128 {
		t.Fatalf("Premultiply() pixel = %#v", got)
	}
	if err := img.Demultiply(); err != nil {
		t.Fatalf("Demultiply() failed: %v", err)
	}
	if got := img.Data; got[3] != 128 {
		t.Fatalf("Demultiply() alpha = %d, want 128", got[3])
	}

	if ImageFilterBlackman144 != ImageFilterBlackman {
		t.Fatalf("ImageFilterBlackman144 = %v, want %v", ImageFilterBlackman144, ImageFilterBlackman)
	}
	if Blackman144 != Blackman {
		t.Fatalf("Blackman144 = %v, want %v", Blackman144, Blackman)
	}
	if FilterBlackman144 != FilterBlackman {
		t.Fatalf("FilterBlackman144 = %v, want %v", FilterBlackman144, FilterBlackman)
	}

	a.ResetPath()
	a.MoveTo(1, 1)
	a.LineTo(8, 1)
	a.LineTo(8, 8)
	a.ClosePolygon()
	a.DrawPathDefault()

	a.ResetPath()
	a.MoveTo(2, 2)
	a.LineTo(6, 2)
	a.LineTo(6, 6)
	a.ClosePolygon()
	a.DrawPathNoTransformDefault()

	a.ViewportDefault(0, 0, 10, 10, 0, 0, 20, 20)
	sx, sy := 5.0, 5.0
	a.WorldToScreen(&sx, &sy)
	if math.Abs(sx-10.0) > 1e-9 || math.Abs(sy-10.0) > 1e-9 {
		t.Fatalf("ViewportDefault worldToScreen(5,5) = (%v,%v), want (10,10)", sx, sy)
	}

	if err := a.BlendImageDefaultAlpha(&img, 0, 0, 1, 1, 1, 1); err != nil {
		t.Fatalf("BlendImageDefaultAlpha() failed: %v", err)
	}
	if err := a.BlendImageSimpleDefaultAlpha(&img, 2, 2); err != nil {
		t.Fatalf("BlendImageSimpleDefaultAlpha() failed: %v", err)
	}
}

func TestFillRadialGradientStops(t *testing.T) {
	const w, h = 64, 64
	buf := make([]uint8, w*h*4)
	a := NewAgg2D()
	a.Attach(buf, w, h, w*4)
	a.ClearAll(NewColor(0, 0, 0, 0))

	// Wet-edge profile: transparent centre → opaque ring at 70% → transparent edge.
	stops := []GradientStop{
		{Position: 0.0, Color: NewColor(255, 0, 0, 0)},   // transparent centre
		{Position: 0.7, Color: NewColor(255, 0, 0, 255)}, // opaque ring
		{Position: 1.0, Color: NewColor(255, 0, 0, 0)},   // transparent edge
	}
	cx, cy, r := 32.0, 32.0, 28.0

	a.NoLine()
	a.Translate(cx, cy)
	a.ResetPath()
	a.AddEllipse(0, 0, r, r, CCW)
	a.FillRadialGradientStops(0, 0, r, stops)
	a.DrawPath(FillOnly)

	// Centre pixel (32,32) must be transparent (alpha ≈ 0).
	centreIdx := (32*w + 32) * 4
	if buf[centreIdx+3] > 20 {
		t.Errorf("centre alpha = %d, want ≈ 0 (transparent centre)", buf[centreIdx+3])
	}

	// Mid-ring pixel at roughly 70% radius should be mostly opaque.
	ringX := int(cx + r*0.7)
	ringIdx := (32*w + ringX) * 4
	if buf[ringIdx+3] < 200 {
		t.Errorf("ring pixel alpha = %d at x=%d, want > 200 (opaque ring)", buf[ringIdx+3], ringX)
	}

	// Corner pixel far outside the ellipse must be untouched.
	outerIdx := 0
	if buf[outerIdx+3] != 0 {
		t.Errorf("outer pixel alpha = %d, want 0", buf[outerIdx+3])
	}
}

func TestFillLinearGradientStops(t *testing.T) {
	const w, h = 60, 16
	buf := make([]uint8, w*h*4)
	a := NewAgg2D()
	a.Attach(buf, w, h, w*4)
	a.ClearAll(NewColor(0, 0, 0, 255))

	// Four-stop ramp: red → green → blue → red. The interior green and blue
	// stops only appear if every stop is fed into the gradient LUT.
	stops := []GradientStop{
		{Position: 0.0, Color: NewColor(255, 0, 0, 255)},
		{Position: 1.0 / 3.0, Color: NewColor(0, 255, 0, 255)},
		{Position: 2.0 / 3.0, Color: NewColor(0, 0, 255, 255)},
		{Position: 1.0, Color: NewColor(255, 0, 0, 255)},
	}

	a.NoLine()
	a.ResetPath()
	a.MoveTo(0, 0)
	a.LineTo(w, 0)
	a.LineTo(w, h)
	a.LineTo(0, h)
	a.ClosePolygon()
	a.FillLinearGradientStops(0, 8, w, 8, stops)
	a.DrawPath(FillOnly)

	// Green stop sits at x = 60/3 = 20; blue stop at x = 40.
	greenIdx := (8*w + 20) * 4
	if buf[greenIdx+1] <= buf[greenIdx] || buf[greenIdx+1] <= buf[greenIdx+2] {
		t.Errorf("green stop missing at x=20: rgb=(%d,%d,%d)", buf[greenIdx], buf[greenIdx+1], buf[greenIdx+2])
	}
	blueIdx := (8*w + 40) * 4
	if buf[blueIdx+2] <= buf[blueIdx] || buf[blueIdx+2] <= buf[blueIdx+1] {
		t.Errorf("blue stop missing at x=40: rgb=(%d,%d,%d)", buf[blueIdx], buf[blueIdx+1], buf[blueIdx+2])
	}
}
