// Package primitives — anti-aliasing edge-coverage tests.
//
// These tests isolate and quantify the AA edge-pixel quality difference between
// the Go AGG reimplementation and the reference C++ matplotlib AGG backend.
//
// Background
// ----------
// When drawing thin lines (~1.15 px) over bright viridis-like backgrounds
// (green/yellow), matplotlib C++ AGG produces more "halo" pixels around the
// line than Go AGG.  Quantitatively (from the unstructured_showcase parity
// analysis):
//
//	At 19% brightness threshold: Go ≈ 1514, Ref ≈ 1525  (ratio 0.99 — identical)
//	At 31% brightness threshold: Go ≈ 2024, Ref ≈ 3824  (ratio 0.53 — 2× off)
//
// The 19% match confirms that fully-covered line-core pixels are identical in
// both implementations.  The 31% discrepancy is entirely in the low-coverage
// (16–94%) edge-pixel zone, where matplotlib produces wider/softer AA halos.
//
// TestThinLineAAEdgeCoverage renders the exact scenario and logs the statistics
// so that before/after comparisons are easy when fixing the AA quality.
//
// TestThinLineCoverageDistribution uses the low-level rasterizer to dump the
// raw coverage values per scanline for a single 1.15 px line, providing a
// direct view into the AA algorithm output.
package primitives

import (
	"fmt"
	"image"
	"testing"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	aggpath "github.com/cwbudde/agg_go/internal/path"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
)

// ─────────────────────────────────────────────────────────────────────────────
// TestThinLineAAEdgeCoverage
// ─────────────────────────────────────────────────────────────────────────────

// TestThinLineAAEdgeCoverage renders dark contour-like lines over bright viridis
// backgrounds and reports per-threshold dark-pixel counts.
//
// This test does NOT assert specific counts — it characterises the current
// behaviour so that before/after improvements are immediately visible in the log.
// Once the AA quality is improved to match matplotlib C++ AGG, expected counts
// can be added via t.Fatalf.
func TestThinLineAAEdgeCoverage(t *testing.T) {
	const (
		w = 320
		h = 480
	)

	ctx := agg.NewContext(w, h)
	ctx.Clear(agg.White)

	ink := agg.RGBA(20.0/255, 31.0/255, 46.0/255, 242.0/255)

	viridisColors := [5][3]uint8{
		{68, 1, 84},    // 0.00 dark purple
		{59, 82, 139},  // 0.25 blue
		{33, 145, 140}, // 0.50 teal
		{94, 201, 98},  // 0.75 green
		{253, 231, 37}, // 1.00 yellow
	}
	lineWidths := []float64{0.5, 1.0, 1.15, 2.0}

	bandH := float64(h) / float64(len(viridisColors))
	for bi, vc := range viridisColors {
		y0 := float64(bi) * bandH
		y1 := y0 + bandH
		ctx.SetColor(agg.NewColor(vc[0], vc[1], vc[2], 255))
		ctx.BeginPath()
		ctx.MoveTo(0, y0)
		ctx.LineTo(float64(w), y0)
		ctx.LineTo(float64(w), y1)
		ctx.LineTo(0, y1)
		ctx.ClosePath()
		ctx.Fill()

		ctx.SetColor(ink)
		for wi, lw := range lineWidths {
			lineY := y0 + bandH*float64(wi+1)/float64(len(lineWidths)+1)
			ctx.SetLineWidth(lw)
			ctx.BeginPath()
			ctx.MoveTo(8, lineY)
			ctx.LineTo(float64(w)-8, lineY)
			ctx.Stroke()
		}
	}

	img := ctx.GetImage().ToGoImage()

	thresholds := []struct {
		name  string
		limit uint32
	}{
		{"12%", 0x2000},
		{"19%", 0x3000},
		{"25%", 0x4000},
		{"31%", 0x5000},
	}

	t.Log("=== Dark-pixel counts per viridis band (x=8..312) ===")
	t.Log("(expected matplotlib C++ AGG counts in parentheses where known)")
	for bi, vc := range viridisColors {
		y0 := int(float64(bi) * bandH)
		y1 := int(float64(bi+1) * bandH)
		counts := countDarkPerThreshold(img, 8, w-8, y0, y1, thresholds)
		t.Logf("  band %d (%3d,%3d,%3d): %s", bi, vc[0], vc[1], vc[2], formatCounts(thresholds, counts))
	}

	t.Log("")
	t.Log("=== Dark-pixel counts: full scene (x=8..312, y=0..480) ===")
	totalCounts := countDarkPerThreshold(img, 8, w-8, 0, h, thresholds)
	for i, th := range thresholds {
		t.Logf("  threshold %s: %d dark pixels", th.name, totalCounts[i])
	}

	// Characterise single 1.15 px line over each background.
	t.Log("")
	t.Log("=== Single 1.15 px line only (each band's 3rd line, wi=2) ===")
	for bi, vc := range viridisColors {
		lineY := float64(bi)*bandH + bandH*3/5
		y0 := int(lineY) - 3
		y1 := int(lineY) + 4
		if y0 < 0 {
			y0 = 0
		}
		if y1 > h {
			y1 = h
		}
		counts := countDarkPerThreshold(img, 8, w-8, y0, y1, thresholds)
		t.Logf("  band %d (%3d,%3d,%3d) lineY=%.1f y=%d..%d: %s",
			bi, vc[0], vc[1], vc[2], lineY, y0, y1,
			formatCounts(thresholds, counts))
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// TestThinLineCoverageDistribution
// ─────────────────────────────────────────────────────────────────────────────

// TestThinLineCoverageDistribution uses the low-level rasterizer to extract the
// raw scanline coverage values for a single horizontal 1.15 px line.
//
// This shows exactly what coverage values the AA algorithm produces BEFORE they
// are composited against any background — the root input to the blending step.
// Comparing this output to the C++ AGG reference reveals whether the difference
// is in the rasterizer or in the compositing/blending step.
//
// Run with -v to see the full coverage tables.
func TestThinLineCoverageDistribution(t *testing.T) {
	const (
		imgW     = 200
		imgH     = 20
		lineW    = 1.15 // stroke width in pixels
		lineXMin = 10.0
		lineXMax = 190.0
	)
	lineYF := 10.3 // sub-pixel line centre

	// Build a stroke path for the line.
	ps := aggpath.NewPathStorageStl()
	ps.MoveTo(lineXMin, lineYF)
	ps.LineTo(lineXMax, lineYF)

	stroke := conv.NewConvStroke(aaStlPathVS{ps: ps})
	stroke.SetWidth(lineW)
	stroke.SetLineCap(basics.SquareCap)

	// Rasterize.
	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	ras.ClipBox(0, 0, float64(imgW), float64(imgH))

	stroke.Rewind(0)
	for {
		x, y, cmd := stroke.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		ras.AddVertex(x, y, uint32(cmd))
	}

	if !ras.RewindScanlines() {
		t.Skip("no scanlines produced for the test line")
	}

	sl := scanline.NewScanlineU8()
	sl.Reset(ras.MinX(), ras.MaxX())

	type spanRecord struct {
		y      int
		x      int
		covers []uint8
	}
	var records []spanRecord

	for ras.SweepScanline(sl) {
		y := sl.Y()
		it := sl.BeginIterator()
		for i, n := 0, sl.NumSpans(); i < n; i++ {
			sp := it.GetSpan()
			numPix := sp.Len
			solid := numPix < 0
			if solid {
				numPix = -numPix
			}
			rec := spanRecord{y: y, x: sp.X, covers: make([]uint8, numPix)}
			for j := 0; j < numPix; j++ {
				if solid {
					rec.covers[j] = sp.Covers[0]
				} else {
					rec.covers[j] = sp.Covers[j]
				}
			}
			records = append(records, rec)
			if i < n-1 {
				it.Next()
			}
		}
	}

	// Print a concise cross-section near x=50.
	t.Logf("Line: y=%.2f, width=%.2f px, x=%.1f..%.1f", lineYF, lineW, lineXMin, lineXMax)
	t.Log("Coverage values near x=50 (y, x_start, covers):")
	for _, rec := range records {
		if rec.x > 60 || rec.x+len(rec.covers) < 45 {
			continue
		}
		end := len(rec.covers)
		if end > 15 {
			end = 15
		}
		t.Logf("  y=%2d x=%d %v", rec.y, rec.x, rec.covers[:end])
	}

	// ── Render over dark-purple and dump pixel values ───────────────────────
	t.Log("")
	t.Log("Rendered pixel values over dark-purple background (68,1,84):")

	buf := make([]uint8, imgW*imgH*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, imgW, imgH, imgW*4)
	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	rb := renderer.NewRendererBaseWithPixfmt(pf)
	rb.Clear(color.RGBA8[color.Linear]{R: 68, G: 1, B: 84, A: 255})

	ps2 := aggpath.NewPathStorageStl()
	ps2.MoveTo(lineXMin, lineYF)
	ps2.LineTo(lineXMax, lineYF)
	stk2 := conv.NewConvStroke(aaStlPathVS{ps: ps2})
	stk2.SetWidth(lineW)
	stk2.SetLineCap(basics.SquareCap)

	ras2 := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	ras2.ClipBox(0, 0, float64(imgW), float64(imgH))
	stk2.Rewind(0)
	for {
		x, y, cmd := stk2.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		ras2.AddVertex(x, y, uint32(cmd))
	}

	sl2 := scanline.NewScanlineU8()
	renscan.RenderScanlinesAASolid(ras2, sl2, rb,
		color.RGBA8[color.Linear]{R: 20, G: 31, B: 46, A: 242})

	// Print cross-section rows around the line centre.
	yLow := int(lineYF) - 2
	yHigh := int(lineYF) + 3
	if yLow < 0 {
		yLow = 0
	}
	if yHigh >= imgH {
		yHigh = imgH - 1
	}

	t.Logf("Cross-section at x=50 (background (68,1,84), ink (20,31,46,α=242)):")
	for y := yLow; y <= yHigh; y++ {
		row := rbuf.Row(y)
		off := 50 * 4
		r, g, b := row[off+0], row[off+1], row[off+2]
		// Estimate coverage via the B channel: B = 84*(1-α)+46*α → α=(84-B)/(84-46)
		alphaEst := (float64(84) - float64(b)) / float64(84-46)
		if alphaEst < 0 {
			alphaEst = 0
		}
		if alphaEst > 1 {
			alphaEst = 1
		}
		t.Logf("  y=%2d: R=%3d G=%3d B=%3d  estimated_coverage=%.3f",
			y, r, g, b, alphaEst)
	}

	t.Log("")
	t.Log("NOTE: Compare these values with C++ AGG output.")
	t.Log("Go AGG typically produces a sharper transition (fewer intermediate-coverage rows)")
	t.Log("than C++ AGG which produces softer transitions spread over more rows/pixels.")
}

// ─────────────────────────────────────────────────────────────────────────────
// TestThinLineOverBrightBackground
// ─────────────────────────────────────────────────────────────────────────────

// TestThinLineOverBrightBackground renders 1.15 px lines over viridis green/purple
// backgrounds and saves the output images for visual inspection.
func TestThinLineOverBrightBackground(t *testing.T) {
	runner := getTestRunner()

	tests := map[string]func() (image.Image, error){
		"thin_line_1px_over_viridis_green":    thinLineOverGreen(1.0),
		"thin_line_115px_over_viridis_green":  thinLineOverGreen(1.15),
		"thin_line_2px_over_viridis_green":    thinLineOverGreen(2.0),
		"thin_line_115px_over_viridis_purple": thinLineOverPurple(1.15),
	}

	suite := runner.RunTestSuite("antialiasing", tests)

	// Always-passing: this test saves images and logs stats only.
	// Once a C++ reference is available, compare against it.
	for _, result := range suite.Results {
		if result.Error != nil {
			t.Logf("Warning: %s: %v", result.Name, result.Error)
		}
	}

	t.Log(runner.GetTestSummary(suite))
	t.Log("Images saved to tests/visual/output/antialiasing_*.png")
	t.Log("Compare with C++ matplotlib AGG reference to assess AA quality.")
}

func thinLineOverGreen(lw float64) func() (image.Image, error) {
	return func() (image.Image, error) {
		const imgW, imgH = 200, 40
		ctx := agg.NewContext(imgW, imgH)
		ctx.Clear(agg.NewColor(94, 201, 98, 255)) // viridis green
		ctx.SetColor(agg.RGBA(20.0/255, 31.0/255, 46.0/255, 242.0/255))
		ctx.SetLineWidth(lw)
		ctx.BeginPath()
		ctx.MoveTo(10, float64(imgH)/2+0.3)
		ctx.LineTo(float64(imgW)-10, float64(imgH)/2+0.3)
		ctx.Stroke()
		return ctx.GetImage().ToGoImage(), nil
	}
}

func thinLineOverPurple(lw float64) func() (image.Image, error) {
	return func() (image.Image, error) {
		const imgW, imgH = 200, 40
		ctx := agg.NewContext(imgW, imgH)
		ctx.Clear(agg.NewColor(68, 1, 84, 255)) // viridis dark purple
		ctx.SetColor(agg.RGBA(20.0/255, 31.0/255, 46.0/255, 242.0/255))
		ctx.SetLineWidth(lw)
		ctx.BeginPath()
		ctx.MoveTo(10, float64(imgH)/2+0.3)
		ctx.LineTo(float64(imgW)-10, float64(imgH)/2+0.3)
		ctx.Stroke()
		return ctx.GetImage().ToGoImage(), nil
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// aaStlPathVS wraps PathStorageStl as a conv.VertexSource for conv.NewConvStroke.
type aaStlPathVS struct{ ps *aggpath.PathStorageStl }

func (s aaStlPathVS) Rewind(id uint) { s.ps.Rewind(id) }
func (s aaStlPathVS) Vertex() (float64, float64, basics.PathCommand) {
	x, y, cmd := s.ps.NextVertex()
	return x, y, basics.PathCommand(cmd)
}

func countDarkPerThreshold(img *image.RGBA, x0, x1, y0, y1 int, thresholds []struct {
	name  string
	limit uint32
}) []int {
	counts := make([]int, len(thresholds))
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			c := img.RGBAAt(x, y)
			r := uint32(c.R) << 8
			g := uint32(c.G) << 8
			b := uint32(c.B) << 8
			for ti, th := range thresholds {
				if r < th.limit && g < th.limit && b < th.limit {
					counts[ti]++
				}
			}
		}
	}
	return counts
}

func formatCounts(thresholds []struct {
	name  string
	limit uint32
}, counts []int) string {
	s := ""
	for i, th := range thresholds {
		if i > 0 {
			s += "  "
		}
		s += fmt.Sprintf("%s=%d", th.name, counts[i])
	}
	return s
}

func min2(a, b int) int {
	if a < b {
		return a
	}
	return b
}
