package agg

import "testing"

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
