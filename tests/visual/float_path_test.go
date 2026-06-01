package visual

import (
	"image"
	"path/filepath"
	"testing"

	agg "github.com/cwbudde/agg_go"
)

// Float Agg2D path visual/demo hook (PLAN.md §4.5).
//
// This renders one composed scene through the public FLOAT path
// (agg.ContextFloat / Agg2DFloat -> ImageFloat.ToRGBA) and the SAME scene
// through the existing 8-bit path (agg.Context -> Image.ToGoImage), then asserts
// the two agree within a documented tolerance. The 8-bit render is the oracle,
// so this needs NO new reference PNGs and cannot disturb the 8-bit baseline.
//
// The scene is fully OPAQUE on purpose: 8-bit ToGoImage exports straight bytes
// while float ToRGBA exports premultiplied bytes; for opaque pixels (a == 255 /
// 1.0) premultiplied == straight, so the boundary conversion is identity and the
// comparison is apples-to-apples. Anti-aliased edges over an opaque background
// stay opaque, so the whole image remains opaque.
//
// Tolerances mirror the proven cross-precision envelopes in
// internal/agg2d/parity_float_test.go: solid fill ~1, gradient/AA ~4.

// drawFloatScene paints the shared scene through the float context.
func drawFloatScene(ctx *agg.ContextFloat) {
	g := ctx.GetAgg2D()
	drawSharedScene(
		func(c agg.Color) { ctx.Clear(c) },
		g.FillColor, g.FillLinearGradient, g.ResetPath, g.MoveTo, g.LineTo,
		g.ClosePolygon, g.DrawPath, g.FillCircle,
		ctx.Width(), ctx.Height(),
	)
}

// drawScene8 paints the shared scene through the 8-bit context.
func drawScene8(ctx *agg.Context) {
	g := ctx.GetAgg2D()
	drawSharedScene(
		func(c agg.Color) { ctx.Clear(c) },
		g.FillColor, g.FillLinearGradient, g.ResetPath, g.MoveTo, g.LineTo,
		g.ClosePolygon, g.DrawPath, g.FillCircle,
		ctx.Width(), ctx.Height(),
	)
}

// drawSharedScene encodes the scene once against the method subset shared by
// both pipelines, so the float and 8-bit renders are byte-for-byte the same
// drawing program.
func drawSharedScene(
	clearAll func(agg.Color),
	fillColor func(agg.Color),
	fillLinearGradient func(x1, y1, x2, y2 float64, c1, c2 agg.Color, profile float64),
	resetPath func(),
	moveTo func(x, y float64),
	lineTo func(x, y float64),
	closePolygon func(),
	drawPath func(agg.DrawPathFlag),
	fillCircle func(cx, cy, radius float64),
	w, h int,
) {
	opaque := func(r, g, b uint8) agg.Color { return agg.Color{R: r, G: g, B: b, A: 255} }

	// Opaque white background.
	clearAll(opaque(255, 255, 255))

	// Horizontal (Y-invariant) opaque linear gradient across the full canvas.
	fillLinearGradient(0, 0, float64(w), 0, opaque(220, 40, 40), opaque(40, 40, 220), 1.0)
	resetPath()
	moveTo(0, 0)
	lineTo(float64(w), 0)
	lineTo(float64(w), float64(h))
	lineTo(0, float64(h))
	closePolygon()
	drawPath(agg.FillOnly)

	// Centered opaque green circle on top.
	fillColor(opaque(20, 160, 60))
	fillCircle(float64(w)/2, float64(h)/2, float64(h)/3)
}

func TestFloatPathVisualParity(t *testing.T) {
	const w, h = 120, 80

	// Float path.
	cf := agg.NewContextFloat(w, h)
	drawFloatScene(cf)
	imgF := cf.GetImage().ToRGBA()

	// 8-bit oracle.
	c8 := agg.NewContext(w, h)
	drawScene8(c8)
	img8 := c8.GetImage().ToGoImage()

	// Save the float render for manual inspection.
	out := filepath.Join("output", "float_path.png")
	if err := savePNG(out, imgF.Pix, w, h); err != nil {
		t.Fatalf("failed to save float render: %v", err)
	}
	t.Logf("float render written to %s", out)

	// Compare interior sample points, skipping shape edges where AA differences
	// (and tiny rounding) are expected. Each sample documents its region.
	type sample struct {
		x, y int
		tol  int
		what string
	}
	cx, cy := w/2, h/2
	samples := []sample{
		{6, h / 2, 4, "gradient left edge band"},
		{w / 4, 8, 4, "gradient upper-left"},
		{3 * w / 4, h - 8, 4, "gradient lower-right"},
		{w - 6, h / 2, 4, "gradient right edge band"},
		{cx, cy, 1, "circle center (solid)"},
		{cx - 8, cy + 6, 1, "circle interior (solid)"},
	}

	for _, s := range samples {
		if d := maxRGBADiff(imgF, img8, s.x, s.y); d > s.tol {
			f := imgF.RGBAAt(s.x, s.y)
			e := img8.RGBAAt(s.x, s.y)
			t.Errorf("%s at (%d,%d): float=%v 8bit=%v maxdiff=%d (tol=%d)",
				s.what, s.x, s.y, f, e, d, s.tol)
		}
	}
}

// TestFloatTextVisualParity renders GSV stroke text through the public float and
// 8-bit paths and asserts whole-frame parity. The 8-bit render is the oracle, so
// no reference PNG is needed. GSV is cgo-free, so this runs without FreeType.
func TestFloatTextVisualParity(t *testing.T) {
	const w, h = 160, 48

	drawText := func(fillColor func(agg.Color), fontGSV func(float64),
		text func(x, y float64, s string, roundOff bool, dx, dy float64),
		clear func(agg.Color),
	) {
		clear(agg.Color{R: 255, G: 255, B: 255, A: 255})
		fontGSV(22)
		fillColor(agg.Color{R: 10, G: 30, B: 120, A: 255})
		text(8, 32, "Float!", false, 0, 0)
	}

	cf := agg.NewContextFloat(w, h)
	gf := cf.GetAgg2D()
	drawText(gf.FillColor, gf.FontGSV, gf.Text, func(c agg.Color) { cf.Clear(c) })
	imgF := cf.GetImage().ToRGBA()

	c8 := agg.NewContext(w, h)
	g8 := c8.GetAgg2D()
	drawText(g8.FillColor, g8.FontGSV, g8.Text, func(c agg.Color) { c8.Clear(c) })
	img8 := c8.GetImage().ToGoImage()

	out := filepath.Join("output", "float_text.png")
	if err := savePNG(out, imgF.Pix, w, h); err != nil {
		t.Fatalf("failed to save float text render: %v", err)
	}
	t.Logf("float text render written to %s", out)

	maxDiff, ink := 0, 0
	for y := range h {
		for x := range w {
			if d := maxRGBADiff(imgF, img8, x, y); d > maxDiff {
				maxDiff = d
			}
			c := img8.RGBAAt(x, y)
			if c.R < 250 || c.G < 250 || c.B < 250 {
				ink++
			}
		}
	}
	if ink < 50 {
		t.Fatalf("GSV text drew too little ink in 8-bit oracle: %d pixels", ink)
	}
	if maxDiff > 2 {
		t.Errorf("float vs 8-bit GSV text max channel diff = %d (tol 2)", maxDiff)
	}
}

// maxRGBADiff returns the largest per-channel absolute difference between two
// *image.RGBA at (x,y).
func maxRGBADiff(a, b *image.RGBA, x, y int) int {
	ca := a.RGBAAt(x, y)
	cb := b.RGBAAt(x, y)
	diff := func(p, q uint8) int {
		d := int(p) - int(q)
		if d < 0 {
			return -d
		}
		return d
	}
	m := diff(ca.R, cb.R)
	for _, d := range []int{diff(ca.G, cb.G), diff(ca.B, cb.B), diff(ca.A, cb.A)} {
		if d > m {
			m = d
		}
	}
	return m
}
