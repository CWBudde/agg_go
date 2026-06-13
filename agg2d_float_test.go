package agg

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/color"
)

func feqf(a, b float32) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= 1e-6
}

func TestPublicAgg2DFloatSolidFill(t *testing.T) {
	img := NewImageFloat(20, 20)
	a := NewAgg2DFloat()
	a.AttachImage(img)
	a.ClearAll(NewColor(0, 0, 0, 0))

	a.FillColor(NewColor(255, 0, 0, 255))
	a.ResetPath()
	a.MoveTo(2, 2)
	a.LineTo(18, 2)
	a.LineTo(18, 18)
	a.LineTo(2, 18)
	a.ClosePolygon()
	a.DrawPath(FillOnly)

	r, g, b, al := img.GetPixelFloat(10, 10)
	if !feqf(r, 1.0) || !feqf(g, 0.0) || !feqf(b, 0.0) || !feqf(al, 1.0) {
		t.Fatalf("center = {%v,%v,%v,%v}, want opaque red", r, g, b, al)
	}
	_, _, _, ca := img.GetPixelFloat(0, 0)
	if ca != 0 {
		t.Fatalf("corner alpha = %v, want 0", ca)
	}
}

func TestPublicAgg2DFloatGradientAndImageOps(t *testing.T) {
	img := NewImageFloat(40, 10)
	a := NewAgg2DFloat()
	a.AttachImage(img)
	a.ClearAll(NewColor(0, 0, 0, 0))

	a.FillLinearGradient(0, 0, 40, 0, NewColor(255, 0, 0, 255), NewColor(0, 0, 255, 255), 1.0)
	a.ResetPath()
	a.MoveTo(0, 0)
	a.LineTo(40, 0)
	a.LineTo(40, 10)
	a.LineTo(0, 10)
	a.ClosePolygon()
	a.DrawPath(FillOnly)

	lr, _, _, _ := img.GetPixelFloat(3, 5)
	_, _, rb, _ := img.GetPixelFloat(36, 5)
	if lr <= 0 || rb <= 0 {
		t.Fatalf("gradient endpoints not rendered: lr=%v rb=%v", lr, rb)
	}

	// CopyImage onto a fresh target.
	dst := NewImageFloat(60, 20)
	b := NewAgg2DFloat()
	b.AttachImage(dst)
	b.ClearAll(NewColor(0, 0, 0, 0))
	b.CopyImage(img, 5, 5)
	_, _, _, a2 := dst.GetPixelFloat(20, 8)
	if a2 <= 0 {
		t.Fatalf("copied region alpha = %v, want > 0", a2)
	}
}

func TestPublicAgg2DFloatBoundaryToRGBA(t *testing.T) {
	img := NewImageFloat(1, 1)
	img.SetPixelFloat(0, 0, 1.0, 0.5, 0.0, 0.5)
	rgba := img.ToRGBA()
	got := rgba.RGBAAt(0, 0)
	// Go image.RGBA is premultiplied: a=128, r=round(1*0.5*255)=128
	if got.A < 126 || got.A > 130 || got.R < 126 || got.R > 130 {
		t.Fatalf("ToRGBA pixel = %+v, want ~{128,64,0,128}", got)
	}
}

func TestPublicAgg2DFloatTransformAndFillMode(t *testing.T) {
	img := NewImageFloat(24, 24)
	a := NewAgg2DFloat()
	a.AttachImage(img)
	a.ClearAll(NewColor(0, 0, 0, 0))

	// Translate then fill a small rect; verify it moved.
	a.FillColor(NewColor(255, 0, 0, 255))
	a.Translate(10, 10)
	a.ResetPath()
	a.MoveTo(0, 0)
	a.LineTo(6, 0)
	a.LineTo(6, 6)
	a.LineTo(0, 6)
	a.ClosePolygon()
	a.DrawPath(FillOnly)
	if _, _, _, al := img.GetPixelFloat(13, 13); al <= 0 {
		t.Fatalf("translated fill missing at (13,13): alpha=%v", al)
	}
	if _, _, _, al := img.GetPixelFloat(3, 3); al != 0 {
		t.Fatalf("untranslated location should be empty: alpha=%v", al)
	}

	// Fill-mode toggles are reachable from the public API.
	a.FillEvenOdd(true)
	if !a.GetFillEvenOdd() {
		t.Fatal("GetFillEvenOdd should be true after FillEvenOdd(true)")
	}
	a.NoFill()
	a.NoLine()
}

func TestPublicAgg2DFloatTransformImage(t *testing.T) {
	// Opaque source with a recognizable solid color.
	src := NewImageFloat(8, 8)
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			src.SetPixelFloat(x, y, 0.0, 0.8, 0.2, 1.0)
		}
	}

	dst := NewImageFloat(40, 40)
	a := NewAgg2DFloat()
	a.AttachImage(dst)
	a.ClearAll(NewColor(0, 0, 0, 0))

	// Affine scale the 8x8 source into a 24x24 destination rectangle.
	if err := a.TransformImageSimple(src, 8, 8, 32, 32); err != nil {
		t.Fatalf("TransformImageSimple: %v", err)
	}
	if _, g, _, al := dst.GetPixelFloat(20, 20); g <= 0 || al <= 0 {
		t.Fatalf("transformed region missing at (20,20): g=%v a=%v", g, al)
	}
	if _, _, _, al := dst.GetPixelFloat(2, 2); al != 0 {
		t.Fatalf("outside transformed region should be empty: alpha=%v", al)
	}

	// Perspective quad maps the source to a non-affine quadrangle.
	dst2 := NewImageFloat(40, 40)
	b := NewAgg2DFloat()
	b.AttachImage(dst2)
	b.ClearAll(NewColor(0, 0, 0, 0))
	quad := [8]float64{6, 8, 34, 5, 32, 36, 9, 31}
	if err := b.TransformImageQuadSimple(src, quad); err != nil {
		t.Fatalf("TransformImageQuadSimple: %v", err)
	}
	if _, _, _, al := dst2.GetPixelFloat(20, 20); al <= 0 {
		t.Fatalf("perspective region missing at (20,20): alpha=%v", al)
	}
}

func TestPublicContextFloat(t *testing.T) {
	ctx := NewContextFloat(20, 20)
	ctx.Clear(NewColor(0, 0, 0, 0))
	ctx.SetColor(NewColor(0, 0, 255, 255))
	ctx.FillRectangle(2, 2, 16, 16)

	img := ctx.GetImage()
	_, _, bb, ba := img.GetPixelFloat(10, 10)
	if !feqf(bb, 1.0) || ba <= 0 {
		t.Fatalf("context fill center = blue? b=%v a=%v", bb, ba)
	}
	if ctx.Width() != 20 || ctx.Height() != 20 {
		t.Fatalf("ctx dims = %dx%d, want 20x20", ctx.Width(), ctx.Height())
	}
}

func TestPublicAgg2DFloatBlendMode(t *testing.T) {
	img := NewImageFloat(24, 24)
	a := NewAgg2DFloat()
	a.AttachImage(img)

	if a.GetBlendMode() != BlendAlpha {
		t.Fatalf("default blend mode = %v, want BlendAlpha", a.GetBlendMode())
	}

	// Opaque background; Multiply an opaque fill over it. Premultiplied == straight
	// for opaque content, so the result is the component-wise product.
	a.ClearAll(NewColor(255, 128, 64, 255)) // ~(1.0, 0.502, 0.251)
	a.SetBlendMode(BlendMultiply)
	if a.GetBlendMode() != BlendMultiply {
		t.Fatalf("blend mode = %v after set, want BlendMultiply", a.GetBlendMode())
	}
	a.FillColor(NewColor(128, 128, 128, 255)) // ~0.502
	a.ResetPath()
	a.MoveTo(4, 4)
	a.LineTo(20, 4)
	a.LineTo(20, 20)
	a.LineTo(4, 20)
	a.ClosePolygon()
	a.DrawPath(FillOnly)

	r, g, b, al := img.GetPixelFloat(12, 12)
	near := func(v, want float32) bool {
		d := v - want
		if d < 0 {
			d = -d
		}
		return d <= 0.01
	}
	// r = 1.0*0.502, g = 0.502*0.502, b = 0.251*0.502
	if !near(r, 0.502) || !near(g, 0.252) || !near(b, 0.126) || !near(al, 1.0) {
		t.Fatalf("multiply center = {%v,%v,%v,%v}, want ~{0.502,0.252,0.126,1}", r, g, b, al)
	}
}

// TestPublicAgg2DFloatText draws GSV stroke text through the public float
// surface and verifies ink lands in the expected region.
func TestPublicAgg2DFloatText(t *testing.T) {
	img := NewImageFloat(96, 32)
	a := NewAgg2DFloat()
	a.AttachImage(img)

	a.ClearAll(NewColor(255, 255, 255, 255))
	a.FontGSV(18)
	if h := a.FontHeight(); h != 18 {
		t.Fatalf("FontHeight = %v, want 18", h)
	}
	if w := a.TextWidth("Hi!"); w <= 0 {
		t.Fatalf("TextWidth = %v, want > 0", w)
	}
	a.FillColor(NewColor(0, 0, 0, 255))
	a.Text(4, 22, "Hi!", false, 0, 0)

	ink := 0
	for y := range 32 {
		for x := range 96 {
			if _, _, _, al := img.GetPixelFloat(x, y); al > 0 {
				r, _, _, _ := img.GetPixelFloat(x, y)
				if r < 1.0 { // darkened by the black stroke
					ink++
				}
			}
		}
	}
	if ink < 20 {
		t.Fatalf("public float GSV text rendered too little ink: %d pixels", ink)
	}
}

// TestPublicAgg2DFloatShapes drives the Shapes-group convenience methods through
// the public float surface and the public 8-bit surface (the oracle), then
// asserts whole-frame parity. The scene is fully opaque so float-premultiplied
// and 8-bit-straight exports are apples-to-apples.
func TestPublicAgg2DFloatShapes(t *testing.T) {
	const w, h = 120, 100

	scene := func(
		clear func(Color), fill func(Color), line func(Color), lw func(float64),
		roundedRect func(x1, y1, x2, y2, r float64),
		star func(cx, cy, r1, r2, startAngle float64, numRays int),
		arc func(cx, cy, rx, ry, start, sweep float64),
		polygon func(xy []float64, n int),
		curve4 func(x1, y1, x2, y2, x3, y3, x4, y4 float64),
	) {
		clear(NewColor(255, 255, 255, 255))
		fill(NewColor(60, 160, 90, 255))
		line(NewColor(10, 10, 10, 255))
		lw(1.5)
		roundedRect(8, 8, 60, 44, 10)
		fill(NewColor(240, 200, 30, 255))
		star(86, 28, 8, 18, 0, 5)
		line(NewColor(180, 30, 30, 255))
		lw(2)
		arc(34, 74, 24, 16, 0.2, 3.6)
		fill(NewColor(40, 120, 200, 255))
		polygon([]float64{80, 60, 112, 78, 86, 96}, 3)
		line(NewColor(20, 160, 120, 255))
		lw(2)
		curve4(6, 92, 22, 60, 50, 60, 66, 92)
	}

	// Float path.
	imgF := NewImageFloat(w, h)
	af := NewAgg2DFloat()
	af.AttachImage(imgF)
	scene(af.ClearAll, af.FillColor, af.LineColor, af.LineWidth,
		af.RoundedRect, af.Star, af.Arc, af.Polygon, af.Curve4)
	rgbaF := imgF.ToRGBA()

	// 8-bit oracle.
	buf := make([]uint8, w*h*4)
	a8 := NewAgg2D()
	a8.Attach(buf, w, h, w*4)
	scene(a8.ClearAll, a8.FillColor, a8.LineColor, a8.LineWidth,
		a8.RoundedRect, a8.Star, a8.Arc, a8.Polygon, a8.Curve4)
	img8 := NewImage(buf, w, h, w*4).ToGoImage()

	maxDiff, ink := 0, 0
	for y := range h {
		for x := range w {
			cf := rgbaF.RGBAAt(x, y)
			c8 := img8.RGBAAt(x, y)
			for _, d := range []int{
				absInt(int(cf.R) - int(c8.R)),
				absInt(int(cf.G) - int(c8.G)),
				absInt(int(cf.B) - int(c8.B)),
				absInt(int(cf.A) - int(c8.A)),
			} {
				if d > maxDiff {
					maxDiff = d
				}
			}
			if c8.R < 250 || c8.G < 250 || c8.B < 250 {
				ink++
			}
		}
	}
	if ink < 100 {
		t.Fatalf("public shapes drew too little ink in 8-bit oracle: %d pixels", ink)
	}
	if maxDiff > 2 {
		t.Errorf("public float vs 8-bit shapes max channel diff = %d (tol 2)", maxDiff)
	}
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// TestPublicAgg2DFloatCurves drives the curve-command and relative path methods
// through the public float surface and the public 8-bit oracle, asserting
// whole-frame parity. The smooth-curve reflection depends on shared control-point
// tracking, so any drift in that state would surface as a pixel diff here.
func TestPublicAgg2DFloatCurves(t *testing.T) {
	const w, h = 120, 100

	scene := func(
		clear func(Color), line func(Color), lw func(float64),
		reset func(), moveTo func(x, y float64),
		quadTo func(xc, yc, xt, yt float64), quadSmooth func(xt, yt float64),
		cubicTo func(xc1, yc1, xc2, yc2, xt, yt float64), cubicSmooth func(xc2, yc2, xt, yt float64),
		horRel func(dx float64), verRel func(dy float64),
		draw func(DrawPathFlag),
	) {
		clear(NewColor(255, 255, 255, 255))
		line(NewColor(20, 60, 180, 255))
		lw(2)

		// Smooth quadratic chain.
		reset()
		moveTo(10, 50)
		quadTo(30, 10, 50, 50)
		quadSmooth(90, 50)
		draw(StrokeOnly)

		// Smooth cubic chain.
		reset()
		moveTo(10, 80)
		cubicTo(25, 55, 45, 55, 55, 80)
		cubicSmooth(85, 105, 100, 80)
		draw(StrokeOnly)

		// Relative hor/ver box.
		reset()
		moveTo(15, 15)
		horRel(40)
		verRel(20)
		horRel(-40)
		verRel(-20)
		draw(StrokeOnly)
	}

	imgF := NewImageFloat(w, h)
	af := NewAgg2DFloat()
	af.AttachImage(imgF)
	scene(af.ClearAll, af.LineColor, af.LineWidth, af.ResetPath, af.MoveTo,
		af.QuadricCurveTo, af.QuadricCurveToSmooth, af.CubicCurveTo, af.CubicCurveToSmooth,
		af.HorLineRel, af.VerLineRel, af.DrawPath)
	rgbaF := imgF.ToRGBA()

	buf := make([]uint8, w*h*4)
	a8 := NewAgg2D()
	a8.Attach(buf, w, h, w*4)
	scene(a8.ClearAll, a8.LineColor, a8.LineWidth, a8.ResetPath, a8.MoveTo,
		a8.QuadricCurveTo, a8.QuadricCurveToSmooth, a8.CubicCurveTo, a8.CubicCurveToSmooth,
		a8.HorLineRel, a8.VerLineRel, a8.DrawPath)
	img8 := NewImage(buf, w, h, w*4).ToGoImage()

	maxDiff, ink := 0, 0
	for y := range h {
		for x := range w {
			cf := rgbaF.RGBAAt(x, y)
			c8 := img8.RGBAAt(x, y)
			for _, d := range []int{
				absInt(int(cf.R) - int(c8.R)),
				absInt(int(cf.G) - int(c8.G)),
				absInt(int(cf.B) - int(c8.B)),
				absInt(int(cf.A) - int(c8.A)),
			} {
				if d > maxDiff {
					maxDiff = d
				}
			}
			if c8.R < 250 || c8.G < 250 || c8.B < 250 {
				ink++
			}
		}
	}
	if ink < 80 {
		t.Fatalf("public curves drew too little ink in 8-bit oracle: %d pixels", ink)
	}
	if maxDiff > 2 {
		t.Errorf("public float vs 8-bit curves max channel diff = %d (tol 2)", maxDiff)
	}
}

// TestPublicAgg2DFloatDashedStrokes drives the dashed-stroke methods through the
// public float surface and the public 8-bit oracle, asserting whole-frame parity.
// Covers a multi-segment dash pattern with a phase offset plus the NoDashes()
// solid-fallback branch.
func TestPublicAgg2DFloatDashedStrokes(t *testing.T) {
	const w, h = 140, 120

	scene := func(
		clear func(Color), line func(Color), lw func(float64), cap func(LineCap),
		reset func(), moveTo func(x, y float64), lineTo func(x, y float64),
		addDash func(dashLen, gapLen float64), dashStart func(offset float64),
		noDashes func(), draw func(DrawPathFlag),
	) {
		clear(NewColor(255, 255, 255, 255))
		line(NewColor(20, 60, 180, 255))
		lw(3)
		cap(CapButt)

		// Dashed polyline with a two-segment pattern and a phase offset.
		addDash(12, 6)
		addDash(4, 6)
		dashStart(3)
		reset()
		moveTo(12, 20)
		lineTo(128, 20)
		lineTo(128, 60)
		lineTo(12, 60)
		draw(StrokeOnly)

		// Solid fallback after NoDashes.
		noDashes()
		line(NewColor(200, 40, 40, 255))
		reset()
		moveTo(12, 95)
		lineTo(128, 95)
		draw(StrokeOnly)
	}

	imgF := NewImageFloat(w, h)
	af := NewAgg2DFloat()
	af.AttachImage(imgF)
	scene(af.ClearAll, af.LineColor, af.LineWidth, af.LineCap, af.ResetPath,
		af.MoveTo, af.LineTo, af.AddDash, af.DashStart, af.NoDashes, af.DrawPath)
	rgbaF := imgF.ToRGBA()

	buf := make([]uint8, w*h*4)
	a8 := NewAgg2D()
	a8.Attach(buf, w, h, w*4)
	scene(a8.ClearAll, a8.LineColor, a8.LineWidth, a8.LineCap, a8.ResetPath,
		a8.MoveTo, a8.LineTo, a8.AddDash, a8.DashStart, a8.NoDashes, a8.DrawPath)
	img8 := NewImage(buf, w, h, w*4).ToGoImage()

	maxDiff, ink := 0, 0
	for y := range h {
		for x := range w {
			cf := rgbaF.RGBAAt(x, y)
			c8 := img8.RGBAAt(x, y)
			for _, d := range []int{
				absInt(int(cf.R) - int(c8.R)),
				absInt(int(cf.G) - int(c8.G)),
				absInt(int(cf.B) - int(c8.B)),
				absInt(int(cf.A) - int(c8.A)),
			} {
				if d > maxDiff {
					maxDiff = d
				}
			}
			if c8.R < 250 || c8.G < 250 || c8.B < 250 {
				ink++
			}
		}
	}
	if ink < 80 {
		t.Fatalf("public dashes drew too little ink in 8-bit oracle: %d pixels", ink)
	}
	if maxDiff > 2 {
		t.Errorf("public float vs 8-bit dashes max channel diff = %d (tol 2)", maxDiff)
	}

	// GetDashStart round-trips through the public surface.
	af.DashStart(9)
	if got := af.GetDashStart(); got != 9 {
		t.Errorf("public GetDashStart = %v, want 9", got)
	}
}

// TestPublicAgg2DFloatGradientVariants drives the gradient-variant setters and
// accessors through the public float surface and the public 8-bit oracle,
// asserting whole-frame parity for a multi-stop radial fill plus accessor
// readbacks.
func TestPublicAgg2DFloatGradientVariants(t *testing.T) {
	const w, h = 64, 64

	fillScene := func(
		clear func(Color),
		stops func(x, y, r float64, s []GradientStop),
		reset func(), moveTo func(x, y float64), lineTo func(x, y float64),
		closePoly func(), draw func(DrawPathFlag),
	) {
		clear(NewColor(255, 255, 255, 255))
		stops(32, 32, 30, []GradientStop{
			{Position: 0.0, Color: NewColor(255, 0, 0, 255)},
			{Position: 0.5, Color: NewColor(0, 200, 0, 255)},
			{Position: 1.0, Color: NewColor(0, 0, 255, 255)},
		})
		reset()
		moveTo(2, 2)
		lineTo(62, 2)
		lineTo(62, 62)
		lineTo(2, 62)
		closePoly()
		draw(FillOnly)
	}

	imgF := NewImageFloat(w, h)
	af := NewAgg2DFloat()
	af.AttachImage(imgF)
	fillScene(af.ClearAll, af.FillRadialGradientStops, af.ResetPath, af.MoveTo, af.LineTo, af.ClosePolygon, af.DrawPath)
	rgbaF := imgF.ToRGBA()

	buf := make([]uint8, w*h*4)
	a8 := NewAgg2D()
	a8.Attach(buf, w, h, w*4)
	fillScene(a8.ClearAll, a8.FillRadialGradientStops, a8.ResetPath, a8.MoveTo, a8.LineTo, a8.ClosePolygon, a8.DrawPath)
	img8 := NewImage(buf, w, h, w*4).ToGoImage()

	maxDiff, ink := 0, 0
	for y := range h {
		for x := range w {
			cf := rgbaF.RGBAAt(x, y)
			c8 := img8.RGBAAt(x, y)
			for _, d := range []int{
				absInt(int(cf.R) - int(c8.R)),
				absInt(int(cf.G) - int(c8.G)),
				absInt(int(cf.B) - int(c8.B)),
				absInt(int(cf.A) - int(c8.A)),
			} {
				if d > maxDiff {
					maxDiff = d
				}
			}
			if c8.R < 250 || c8.G < 250 || c8.B < 250 {
				ink++
			}
		}
	}
	if ink < 100 {
		t.Fatalf("public gradient variants drew too little ink in 8-bit oracle: %d pixels", ink)
	}
	if maxDiff > 3 {
		t.Errorf("public float vs 8-bit gradient variants max channel diff = %d (tol 3)", maxDiff)
	}

	// Accessor parity through the public surface.
	af.LineRadialGradientMultiStop(10, 10, 12, NewColor(0, 0, 0, 255), NewColor(128, 128, 128, 255), NewColor(255, 255, 255, 255))
	a8.LineRadialGradientMultiStop(10, 10, 12, NewColor(0, 0, 0, 255), NewColor(128, 128, 128, 255), NewColor(255, 255, 255, 255))
	af.FillRadialGradientPos(20, 20, 25)
	a8.FillRadialGradientPos(20, 20, 25)
	af.LineRadialGradientPos(40, 40, 18)
	a8.LineRadialGradientPos(40, 40, 18)

	if af.FillGradientFlag() != a8.FillGradientFlag() {
		t.Errorf("public FillGradientFlag float=%v 8bit=%v", af.FillGradientFlag(), a8.FillGradientFlag())
	}
	if af.LineGradientFlag() != a8.LineGradientFlag() {
		t.Errorf("public LineGradientFlag float=%v 8bit=%v", af.LineGradientFlag(), a8.LineGradientFlag())
	}
	if af.FillGradientD1() != a8.FillGradientD1() || af.FillGradientD2() != a8.FillGradientD2() {
		t.Errorf("public FillGradient D1/D2 float=(%v,%v) 8bit=(%v,%v)",
			af.FillGradientD1(), af.FillGradientD2(), a8.FillGradientD1(), a8.FillGradientD2())
	}
	if af.LineGradientD1() != a8.LineGradientD1() || af.LineGradientD2() != a8.LineGradientD2() {
		t.Errorf("public LineGradient D1/D2 float=(%v,%v) 8bit=(%v,%v)",
			af.LineGradientD1(), af.LineGradientD2(), a8.LineGradientD1(), a8.LineGradientD2())
	}
}

// TestPublicAgg2DFloatViewportCoordinateMapping exercises the public viewport +
// coordinate-mapping surface of Agg2DFloat against the 8-bit Agg2D oracle. The
// math is color-agnostic, so scalar/bool results must match exactly and a
// viewport-mapped fill must land on the same pixels.
func TestPublicAgg2DFloatViewportCoordinateMapping(t *testing.T) {
	const w, h = 60, 60

	// Render-parity: a viewport mapping must produce identical pixels.
	a8 := NewAgg2D()
	buf := make([]uint8, w*h*4)
	a8.Attach(buf, w, h, w*4)
	af := NewAgg2DFloat()
	imgF := NewImageFloat(w, h)
	af.AttachImage(imgF)

	for _, s := range []interface {
		ClearAll(Color)
		FillColor(Color)
		Viewport(float64, float64, float64, float64, float64, float64, float64, float64, ViewportOption)
		ViewportDefault(float64, float64, float64, float64, float64, float64, float64, float64)
		ResetPath()
		MoveTo(float64, float64)
		LineTo(float64, float64)
		ClosePolygon()
		DrawPath(DrawPathFlag)
	}{a8, af} {
		s.ClearAll(NewColor(255, 255, 255, 255))
		s.Viewport(0, 0, 10, 10, 5, 5, 55, 55, Anisotropic)
		s.FillColor(NewColor(200, 40, 40, 255))
		s.ResetPath()
		s.MoveTo(2, 2)
		s.LineTo(8, 2)
		s.LineTo(8, 8)
		s.LineTo(2, 8)
		s.ClosePolygon()
		s.DrawPath(FillOnly)
	}

	rgbaF := imgF.ToRGBA()
	maxDiff, ink := 0, 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := (y*w + x) * 4
			c8 := []int{int(buf[i]), int(buf[i+1]), int(buf[i+2]), int(buf[i+3])}
			cf := []int{int(rgbaF.Pix[i]), int(rgbaF.Pix[i+1]), int(rgbaF.Pix[i+2]), int(rgbaF.Pix[i+3])}
			for k := 0; k < 4; k++ {
				if d := absInt(c8[k] - cf[k]); d > maxDiff {
					maxDiff = d
				}
			}
			if buf[i] < 250 || buf[i+1] < 250 || buf[i+2] < 250 {
				ink++
			}
		}
	}
	if ink < 50 {
		t.Fatalf("public viewport drew too little ink in 8-bit oracle: %d pixels", ink)
	}
	if maxDiff > 2 {
		t.Errorf("public float vs 8-bit viewport max channel diff = %d (tol 2)", maxDiff)
	}

	// Scalar/bool accessor parity after an identical ViewportDefault setup.
	a8b := NewAgg2D()
	buf2 := make([]uint8, 50*50*4)
	a8b.Attach(buf2, 50, 50, 50*4)
	afb := NewAgg2DFloat()
	afb.AttachImage(NewImageFloat(50, 50))
	a8b.ViewportDefault(0, 0, 100, 100, 0, 0, 50, 50)
	afb.ViewportDefault(0, 0, 100, 100, 0, 0, 50, 50)

	for _, d := range []float64{0, 1, 7.5, 100} {
		if math.Abs(a8b.WorldToScreenDistance(d)-afb.WorldToScreenDistance(d)) > 1e-9 {
			t.Errorf("public WorldToScreenDistance(%v): float=%v 8bit=%v", d, afb.WorldToScreenDistance(d), a8b.WorldToScreenDistance(d))
		}
		got8, ok8 := a8b.ScreenToWorldDistance(d)
		gotF, okF := afb.ScreenToWorldDistance(d)
		if ok8 != okF || math.Abs(got8-gotF) > 1e-9 {
			t.Errorf("public ScreenToWorldDistance(%v): float=(%v,%v) 8bit=(%v,%v)", d, gotF, okF, got8, ok8)
		}
	}

	for _, p := range [][2]float64{{0, 0}, {3.3, 7.7}, {49, 49}} {
		x8, y8 := p[0], p[1]
		a8b.AlignPoint(&x8, &y8)
		xF, yF := p[0], p[1]
		afb.AlignPoint(&xF, &yF)
		if math.Abs(x8-xF) > 1e-9 || math.Abs(y8-yF) > 1e-9 {
			t.Errorf("public AlignPoint(%v): float=(%v,%v) 8bit=(%v,%v)", p, xF, yF, x8, y8)
		}
	}

	for _, p := range [][2]float64{{0, 0}, {25, 25}, {49, 49}, {-5, 5}, {60, 60}} {
		if a8b.InBox(p[0], p[1]) != afb.InBox(p[0], p[1]) {
			t.Errorf("public InBox(%v): float=%v 8bit=%v", p, afb.InBox(p[0], p[1]), a8b.InBox(p[0], p[1]))
		}
	}

	afb.AffineImageResamplePolicy(AffineImageResamplePreferFiltered)
	a8b.AffineImageResamplePolicy(AffineImageResamplePreferFiltered)
	if afb.GetAffineImageResamplePolicy() != a8b.GetAffineImageResamplePolicy() {
		t.Errorf("public GetAffineImageResamplePolicy: float=%v 8bit=%v",
			afb.GetAffineImageResamplePolicy(), a8b.GetAffineImageResamplePolicy())
	}
}

// TestPublicAgg2DFloatTransformStack exercises the public transform-stack +
// affine-matrix surface of Agg2DFloat against the 8-bit Agg2D oracle. The math
// is color-agnostic, so the resulting matrix and stack depth must match exactly.
func TestPublicAgg2DFloatTransformStack(t *testing.T) {
	type stackScene interface {
		ResetTransformations()
		Translate(x, y float64)
		Scale(sx, sy float64)
		Rotate(angle float64)
		Affine(tr *Transformations)
		GetTransformations() *Transformations
		SetTransformations(tr *Transformations)
		PushTransform()
		PopTransform() bool
	}

	run := func(s stackScene) [6]float64 {
		s.ResetTransformations()
		s.Translate(10, 20)
		s.PushTransform()
		s.Scale(2, 3)
		s.Rotate(0.5)
		s.Affine(&Transformations{AffineMatrix: [6]float64{1, 0.2, 0.1, 1, 5, 5}})
		s.PopTransform()
		s.Affine(&Transformations{AffineMatrix: [6]float64{1, 0, 0, 1, 3, 4}})
		return s.GetTransformations().AffineMatrix
	}

	a8 := NewAgg2D()
	af := NewAgg2DFloat()
	m8 := run(a8)
	mF := run(af)

	for i := range m8 {
		if math.Abs(m8[i]-mF[i]) > 1e-9 {
			t.Fatalf("public transform-stack matrix mismatch at [%d]: float=%v 8bit=%v", i, mF, m8)
		}
	}

	// Get/Set round-trip through the public surface.
	af.ResetTransformations()
	af.Translate(7, 8)
	af.Scale(1.5, 2.5)
	saved := af.GetTransformations()
	af.Rotate(1.0)
	af.SetTransformations(saved)
	got := af.GetTransformations()
	for i := range saved.AffineMatrix {
		if math.Abs(saved.AffineMatrix[i]-got.AffineMatrix[i]) > 1e-9 {
			t.Fatalf("public Get/SetTransformations round-trip mismatch at [%d]: got %v want %v", i, got.AffineMatrix, saved.AffineMatrix)
		}
	}

	// PopTransform on an empty stack reports false through the public surface.
	if af.PopTransform() {
		t.Error("public PopTransform on empty stack should return false")
	}
	af.PushTransform()
	if !af.PopTransform() {
		t.Error("public PopTransform after PushTransform should return true")
	}
}

// publicStateScene is the root-level (package agg) method subset exercised by
// the state-accessor parity test. Both the 8-bit Agg2D and the float twin
// implement it, letting one driver pin the float public surface to the 8-bit
// oracle.
type publicStateScene interface {
	FillColorRGBA(r, g, b, a uint8)
	LineColorRGBA(r, g, b, a uint8)
	GetFillColor() Color
	GetLineColor() Color
	LineCap(LineCap)
	LineJoin(LineJoin)
	GetLineCap() LineCap
	GetLineJoin() LineJoin
	MiterLimit(ml float64)
	GetMiterLimit() float64
	ClipBox(x1, y1, x2, y2 float64)
	GetClipBox() (x1, y1, x2, y2 float64)
	GetClipBoxRect() RectD
	ImageFilter(ImageFilter)
	GetImageFilter() ImageFilter
	ImageResample(ImageResample)
	GetImageResample() ImageResample
	SetImageFilterRadius(ft ImageFilter, radius float64)
	AntiAliasGamma(float64)
	GetAntiAliasGamma() float64
	FillEvenOdd(bool)
	IsEvenOddFillRule() bool
	IsNonZeroFillRule() bool
	FillRuleDescription() string
	ResetStyle()
}

var (
	_ publicStateScene = (*Agg2D)(nil)
	_ publicStateScene = (*Agg2DFloat)(nil)
)

func TestPublicAgg2DFloatStateAccessors(t *testing.T) {
	type snap struct {
		fill, line Color
		cap        LineCap
		join       LineJoin
		miter      float64
		cx1, cy1   float64
		cx2, cy2   float64
		rect       RectD
		filter     ImageFilter
		resample   ImageResample
		gamma      float64
		eo, nz     bool
		desc       string
	}
	run := func(s publicStateScene) snap {
		s.FillColorRGBA(10, 20, 30, 40)
		s.LineColorRGBA(50, 60, 70, 80)
		s.LineCap(CapSquare)
		s.LineJoin(JoinBevel)
		s.MiterLimit(7.5)
		s.ClipBox(5, 6, 95, 96)
		s.ImageFilter(Bicubic)
		s.ImageResample(ResampleAlways)
		s.AntiAliasGamma(1.7)
		s.FillEvenOdd(true)
		x1, y1, x2, y2 := s.GetClipBox()
		return snap{
			fill:     s.GetFillColor(),
			line:     s.GetLineColor(),
			cap:      s.GetLineCap(),
			join:     s.GetLineJoin(),
			miter:    s.GetMiterLimit(),
			cx1:      x1,
			cy1:      y1,
			cx2:      x2,
			cy2:      y2,
			rect:     s.GetClipBoxRect(),
			filter:   s.GetImageFilter(),
			resample: s.GetImageResample(),
			gamma:    s.GetAntiAliasGamma(),
			eo:       s.IsEvenOddFillRule(),
			nz:       s.IsNonZeroFillRule(),
			desc:     s.FillRuleDescription(),
		}
	}

	if s8, sF := run(NewAgg2D()), run(NewAgg2DFloat()); s8 != sF {
		t.Errorf("public state accessor mismatch:\n  8bit = %+v\n float = %+v", s8, sF)
	}

	// SetImageFilterRadius updates the filter id through the public surface.
	af := NewAgg2DFloat()
	af.SetImageFilterRadius(FilterLanczos, 3.0)
	if got := af.GetImageFilter(); got != FilterLanczos {
		t.Errorf("public GetImageFilter after radius = %v, want FilterLanczos", got)
	}

	// ResetStyle returns defaults identically for both pipelines.
	reset := func(s publicStateScene) (Color, Color, LineCap, LineJoin, float64, bool) {
		s.FillColorRGBA(1, 1, 1, 1)
		s.LineColorRGBA(2, 2, 2, 2)
		s.LineCap(CapSquare)
		s.MiterLimit(9)
		s.FillEvenOdd(true)
		s.ResetStyle()
		return s.GetFillColor(), s.GetLineColor(), s.GetLineCap(), s.GetLineJoin(), s.GetMiterLimit(), s.IsEvenOddFillRule()
	}
	f8, l8, c8, j8, m8, e8 := reset(NewAgg2D())
	fF, lF, cF, jF, mF, eF := reset(NewAgg2DFloat())
	if f8 != fF || l8 != lF || c8 != cF || j8 != jF || m8 != mF || e8 != eF {
		t.Errorf("public ResetStyle mismatch: 8bit=(%v %v %v %v %v %v) float=(%v %v %v %v %v %v)",
			f8, l8, c8, j8, m8, e8, fF, lF, cF, jF, mF, eF)
	}
}

func TestPublicAgg2DFloatClearClipBoxRGBA(t *testing.T) {
	const w, h = 16, 16
	af := NewAgg2DFloat()
	img := NewImageFloat(w, h)
	af.AttachImage(img)
	af.ClearClipBoxRGBA(200, 100, 50, 255)

	a8 := NewAgg2D()
	buf := make([]uint8, w*h*4)
	a8.Attach(buf, w, h, w*4)
	a8.ClearClipBoxRGBA(200, 100, 50, 255)

	rgba := img.ToRGBA()
	for y := 0; y < h; y += 5 {
		for x := 0; x < w; x += 5 {
			o := (y*w + x) * 4
			for c := 0; c < 4; c++ {
				if absInt(int(rgba.Pix[o+c])-int(buf[o+c])) > 2 {
					t.Errorf("ClearClipBoxRGBA pixel (%d,%d) ch %d: float=%d 8bit=%d", x, y, c, rgba.Pix[o+c], buf[o+c])
				}
			}
		}
	}
}

// publicAliasScene exercises the C++-style accessor aliases that the 8-bit
// public surface exposes (MasterAlpha/BlendMode/ImageBlendMode/ImageBlendColor/
// ImageBlendColorRGBA). The float twin must offer the same alias spelling.
type publicAliasScene interface {
	MasterAlpha(float64)
	GetMasterAlpha() float64
	BlendMode(BlendMode)
	GetBlendMode() BlendMode
	ImageBlendMode(BlendMode)
	GetImageBlendMode() BlendMode
	ImageBlendColor(Color)
	GetImageBlendColor() Color
	ImageBlendColorRGBA(r, g, b, a uint8)
}

var (
	_ publicAliasScene = (*Agg2D)(nil)
	_ publicAliasScene = (*Agg2DFloat)(nil)
)

func TestPublicAgg2DFloatAccessorAliases(t *testing.T) {
	type snap struct {
		master float64
		blend  BlendMode
		imode  BlendMode
		icolor Color
	}
	run := func(s publicAliasScene) snap {
		s.MasterAlpha(0.5)
		s.BlendMode(BlendMultiply)
		s.ImageBlendMode(BlendDarken)
		s.ImageBlendColorRGBA(11, 22, 33, 44)
		return snap{
			master: s.GetMasterAlpha(),
			blend:  s.GetBlendMode(),
			imode:  s.GetImageBlendMode(),
			icolor: s.GetImageBlendColor(),
		}
	}
	if s8, sF := run(NewAgg2D()), run(NewAgg2DFloat()); s8 != sF {
		t.Errorf("public accessor-alias mismatch:\n  8bit = %+v\n float = %+v", s8, sF)
	}

	// ImageBlendColor (Color form) round-trips through the float surface.
	af := NewAgg2DFloat()
	af.ImageBlendColor(NewColor(1, 2, 3, 4))
	if got := af.GetImageBlendColor(); got != NewColor(1, 2, 3, 4) {
		t.Errorf("ImageBlendColor round-trip = %v, want {1 2 3 4}", got)
	}
}

// TestPublicAgg2DFloatImageConvenience exercises the public float-twin image
// convenience wrappers: whole-image float-dst copy/blend, the default-alpha
// spellings, and that they land pixels at the rounded destination.
func TestPublicAgg2DFloatImageConvenience(t *testing.T) {
	src := NewImageFloat(4, 4)
	for y := range 4 {
		for x := range 4 {
			src.SetPixelFloat(x, y, 1.0, 0.0, 0.0, 1.0)
		}
	}

	// CopyImageSimple.
	dst := NewImageFloat(16, 16)
	a := NewAgg2DFloat()
	a.AttachImage(dst)
	a.ClearAll(NewColor(0, 0, 0, 0))
	if err := a.CopyImageSimple(src, 3, 3); err != nil {
		t.Fatalf("CopyImageSimple: %v", err)
	}
	if r, _, _, al := dst.GetPixelFloat(4, 4); r <= 0 || al <= 0 {
		t.Fatalf("CopyImageSimple pixel(4,4) = r=%v a=%v, want opaque red", r, al)
	}

	// BlendImageSimple.
	dst2 := NewImageFloat(16, 16)
	b := NewAgg2DFloat()
	b.AttachImage(dst2)
	b.ClearAll(NewColor(0, 0, 0, 0))
	if err := b.BlendImageSimple(src, 3, 3, 255); err != nil {
		t.Fatalf("BlendImageSimple: %v", err)
	}
	if _, _, _, al := dst2.GetPixelFloat(5, 5); al <= 0 {
		t.Fatalf("BlendImageSimple pixel(5,5) alpha = %v, want > 0", al)
	}

	// BlendImageDefaultAlpha (integer dst).
	dst3 := NewImageFloat(16, 16)
	c := NewAgg2DFloat()
	c.AttachImage(dst3)
	c.ClearAll(NewColor(0, 0, 0, 0))
	c.BlendImageDefaultAlpha(src, 3, 3)
	if _, _, _, al := dst3.GetPixelFloat(5, 5); al <= 0 {
		t.Fatalf("BlendImageDefaultAlpha pixel(5,5) alpha = %v, want > 0", al)
	}

	// BlendImageSimpleDefaultAlpha.
	dst4 := NewImageFloat(16, 16)
	d := NewAgg2DFloat()
	d.AttachImage(dst4)
	d.ClearAll(NewColor(0, 0, 0, 0))
	if err := d.BlendImageSimpleDefaultAlpha(src, 3, 3); err != nil {
		t.Fatalf("BlendImageSimpleDefaultAlpha: %v", err)
	}
	if _, _, _, al := dst4.GetPixelFloat(5, 5); al <= 0 {
		t.Fatalf("BlendImageSimpleDefaultAlpha pixel(5,5) alpha = %v, want > 0", al)
	}

	// Nil source is an error on the float-dst forms.
	if err := a.CopyImageSimple(nil, 0, 0); err == nil {
		t.Fatal("CopyImageSimple(nil) should error")
	}
}

// TestPublicAgg2DFloatSaveImagePPM verifies the public float-twin PPM exporter
// writes a valid binary P6 file with the correct dimensions.
func TestPublicAgg2DFloatSaveImagePPM(t *testing.T) {
	const w, h = 5, 3
	img := NewImageFloat(w, h)
	a := NewAgg2DFloat()
	a.AttachImage(img)
	a.ClearAll(NewColor(10, 20, 30, 255))

	path := filepath.Join(t.TempDir(), "float.ppm")
	if err := a.SaveImagePPM(path); err != nil {
		t.Fatalf("SaveImagePPM: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read ppm: %v", err)
	}
	header := fmt.Sprintf("P6\n%d %d\n255\n", w, h)
	if !bytes.HasPrefix(raw, []byte(header)) {
		t.Fatalf("ppm header = %q, want prefix %q", raw[:min(len(raw), 20)], header)
	}
	if len(raw) != len(header)+w*h*3 {
		t.Fatalf("ppm size = %d, want %d", len(raw), len(header)+w*h*3)
	}

	// A nil target errors.
	if err := NewAgg2DFloat().SaveImagePPM(path); err == nil {
		t.Fatal("SaveImagePPM with no attached buffer should error")
	}
}

// TestPublicAgg2DFloatDrawPathDefaults verifies the default-mode DrawPath
// convenience wrappers render (fill + stroke) on the public float surface.
func TestPublicAgg2DFloatDrawPathDefaults(t *testing.T) {
	img := NewImageFloat(20, 20)
	a := NewAgg2DFloat()
	a.AttachImage(img)
	a.ClearAll(NewColor(0, 0, 0, 0))
	a.FillColor(NewColor(255, 0, 0, 255))
	a.LineColor(NewColor(0, 0, 255, 255))
	a.LineWidth(1.0)

	a.ResetPath()
	a.MoveTo(4, 4)
	a.LineTo(15, 4)
	a.LineTo(15, 15)
	a.LineTo(4, 15)
	a.ClosePolygon()
	a.DrawPathDefault()

	if _, _, _, al := img.GetPixelFloat(9, 9); al <= 0 {
		t.Fatalf("DrawPathDefault interior alpha = %v, want > 0", al)
	}

	// DrawPathNoTransformDefault renders even with a world transform set.
	img2 := NewImageFloat(20, 20)
	b := NewAgg2DFloat()
	b.AttachImage(img2)
	b.ClearAll(NewColor(0, 0, 0, 0))
	b.FillColor(NewColor(0, 255, 0, 255))
	b.Translate(100, 100) // would push the path off-buffer if applied
	b.ResetPath()
	b.MoveTo(4, 4)
	b.LineTo(15, 4)
	b.LineTo(15, 15)
	b.LineTo(4, 15)
	b.ClosePolygon()
	b.DrawPathNoTransformDefault()

	if _, _, _, al := img2.GetPixelFloat(9, 9); al <= 0 {
		t.Fatalf("DrawPathNoTransformDefault interior alpha = %v, want > 0 (transform should be ignored)", al)
	}
}

// publicFloatSpanGen is a trivial constant-color float span generator for the
// public-surface RenderScanlinesAAWithSpanGen test.
type publicFloatSpanGen struct {
	c color.RGBA32[color.Linear]
}

func (g *publicFloatSpanGen) Prepare() {}

func (g *publicFloatSpanGen) Generate(span []color.RGBA32[color.Linear], x, y, length int) {
	for i := 0; i < length && i < len(span); i++ {
		span[i] = g.c
	}
}

// TestPublicAgg2DFloatRasterizerEscapeHatches exercises GetInternalRasterizer +
// RenderRasterizerWithColor / ScanlineRender / RenderScanlinesAAWithSpanGen on
// the public float surface.
func TestPublicAgg2DFloatRasterizerEscapeHatches(t *testing.T) {
	addTri := func(ras interface {
		Reset()
		AddVertex(x, y float64, cmd uint32)
	},
	) {
		ras.Reset()
		ras.AddVertex(2, 2, uint32(basics.PathCmdMoveTo))
		ras.AddVertex(14, 2, uint32(basics.PathCmdLineTo))
		ras.AddVertex(8, 14, uint32(basics.PathCmdLineTo))
		ras.AddVertex(0, 0, uint32(basics.PathCmdEndPoly)|uint32(basics.PathFlagsClose))
	}

	// RenderRasterizerWithColor.
	img := NewImageFloat(16, 16)
	a := NewAgg2DFloat()
	a.AttachImage(img)
	a.ClearAll(NewColor(0, 0, 0, 0))
	if a.GetInternalRasterizer() == nil {
		t.Fatal("GetInternalRasterizer returned nil")
	}
	addTri(a.GetInternalRasterizer())
	a.RenderRasterizerWithColor(NewColor(255, 0, 0, 255))
	if r, _, _, al := img.GetPixelFloat(8, 6); r <= 0 || al <= 0 {
		t.Fatalf("RenderRasterizerWithColor pixel(8,6) = r=%v a=%v, want opaque red", r, al)
	}

	// RenderScanlinesAAWithSpanGen.
	img2 := NewImageFloat(16, 16)
	b := NewAgg2DFloat()
	b.AttachImage(img2)
	b.ClearAll(NewColor(0, 0, 0, 0))
	ras := b.GetInternalRasterizer()
	addTri(ras)
	gen := &publicFloatSpanGen{c: color.NewRGBA32[color.Linear](0.0, 0.0, 1.0, 1.0)}
	b.RenderScanlinesAAWithSpanGen(ras, gen)
	if _, _, bl, al := img2.GetPixelFloat(8, 6); bl <= 0 || al <= 0 {
		t.Fatalf("RenderScanlinesAAWithSpanGen pixel(8,6) = b=%v a=%v, want opaque blue", bl, al)
	}
}

// TestPublicAgg2DFloatGouraudTriangle renders a three-color Gouraud triangle
// through the public float surface and the 8-bit oracle, asserting whole-frame
// parity. The float twin interpolates colors in float space while the 8-bit
// oracle interpolates in integer 0-255 space, so a small per-pixel tolerance is
// allowed.
func TestPublicAgg2DFloatGouraudTriangle(t *testing.T) {
	const w, h = 80, 80

	scene := func(
		clear func(Color),
		gouraud func(x1, y1, x2, y2, x3, y3 float64, c1, c2, c3 Color, d float64),
	) {
		clear(NewColor(0, 0, 0, 255))
		gouraud(8, 8, 72, 8, 40, 72,
			NewColor(255, 0, 0, 255),
			NewColor(0, 255, 0, 255),
			NewColor(0, 0, 255, 255), 0)
	}

	// Float path.
	imgF := NewImageFloat(w, h)
	af := NewAgg2DFloat()
	af.AttachImage(imgF)
	scene(af.ClearAll, af.GouraudTriangle)
	rgbaF := imgF.ToRGBA()

	// 8-bit oracle.
	buf := make([]uint8, w*h*4)
	a8 := NewAgg2D()
	a8.Attach(buf, w, h, w*4)
	scene(a8.ClearAll, a8.GouraudTriangle)
	img8 := NewImage(buf, w, h, w*4).ToGoImage()

	maxDiff, ink := 0, 0
	for y := range h {
		for x := range w {
			cf := rgbaF.RGBAAt(x, y)
			c8 := img8.RGBAAt(x, y)
			for _, d := range []int{
				absInt(int(cf.R) - int(c8.R)),
				absInt(int(cf.G) - int(c8.G)),
				absInt(int(cf.B) - int(c8.B)),
				absInt(int(cf.A) - int(c8.A)),
			} {
				if d > maxDiff {
					maxDiff = d
				}
			}
			if c8.R > 10 || c8.G > 10 || c8.B > 10 {
				ink++
			}
		}
	}
	if ink < 500 {
		t.Fatalf("Gouraud triangle drew too little ink in 8-bit oracle: %d pixels", ink)
	}
	if maxDiff > 2 {
		t.Errorf("public float vs 8-bit Gouraud max channel diff = %d (tol 2)", maxDiff)
	}
}
