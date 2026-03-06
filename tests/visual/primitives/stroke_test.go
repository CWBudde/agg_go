// Package primitives contains visual tests for basic geometric primitives.
package primitives

import (
	"image"
	"os"
	"testing"

	agg "agg_go"
)

// TestStrokeStyles runs visual tests for stroke-specific rendering behavior.
func TestStrokeStyles(t *testing.T) {
	runner := getTestRunner()
	tests := getStrokeTests()

	suite := runner.RunTestSuite("strokes", tests)

	failed := 0
	for _, result := range suite.Results {
		if result.Error != nil {
			t.Errorf("Test %s failed with error: %v", result.Name, result.Error)
			failed++
		} else if !result.Passed {
			t.Errorf("Test %s failed: %d/%d pixels different",
				result.Name, result.Comparison.DifferentPixels, result.Comparison.TotalPixels)
			failed++
		}
	}

	t.Logf("%s", runner.GetTestSummary(suite))

	if failed > 0 {
		t.Fatalf("%d stroke tests failed", failed)
	}
}

// TestGenerateStrokeReferences generates reference images for stroke tests.
// Use: GENERATE_REFERENCES=1 go test -run TestGenerateStrokeReferences ./tests/visual/primitives/
func TestGenerateStrokeReferences(t *testing.T) {
	if os.Getenv("GENERATE_REFERENCES") != "1" {
		t.Skip("Skipping reference generation (set GENERATE_REFERENCES=1 to enable)")
	}

	runner := getTestRunner()
	tests := getStrokeTests()

	if err := runner.CreateReferenceImages(tests, "primitives"); err != nil {
		t.Fatalf("Failed to create reference images: %v", err)
	}

	t.Logf("Generated %d reference images for stroke tests", len(tests))
}

func getStrokeTests() map[string]func() (image.Image, error) {
	return map[string]func() (image.Image, error){
		"dashed_round_cap_comparison":   testDashedRoundCapComparison,
		"line_cap_style_comparison":     testLineCapStyleComparison,
		"line_join_style_comparison":    testLineJoinStyleComparison,
		"dash_pattern_variants":         testDashPatternVariants,
		"dash_offset_phase_comparison":  testDashOffsetPhaseComparison,
		"miter_limit_comparison":        testMiterLimitComparison,
		"stroke_width_ramp":             testStrokeWidthRamp,
		"subpixel_stroke_alignment":     testSubpixelStrokeAlignment,
		"closed_path_stroke_comparison": testClosedPathStrokeComparison,
	}
}

func testDashedRoundCapComparison() (image.Image, error) {
	ctx := agg.NewContext(260, 140)
	ctx.Clear(agg.White)

	drawDashedLine := func(y float64, c agg.Color, cap agg.LineCap) {
		ctx.SetColor(c)
		ctx.BeginPath()
		ctx.SetLineWidth(16)
		ctx.SetLineCap(cap)
		ctx.ClearDashes()
		ctx.AddDash(26, 14)
		ctx.MoveTo(30, y)
		ctx.LineTo(230, y)
		ctx.Stroke()
	}

	drawDashedLine(45, agg.DarkGray, agg.CapButt)
	drawDashedLine(95, agg.Black, agg.CapRound)

	return ctx.GetImage().ToGoImage(), nil
}

func testLineCapStyleComparison() (image.Image, error) {
	ctx := agg.NewContext(320, 180)
	ctx.Clear(agg.White)

	ctx.SetLineWidth(18)
	ctx.SetLineJoin(agg.JoinRound)
	ctx.ClearDashes()

	draw := func(y float64, c agg.Color, cap agg.LineCap) {
		ctx.SetColor(c)
		ctx.SetLineCap(cap)
		ctx.BeginPath()
		ctx.MoveTo(55, y)
		ctx.LineTo(265, y)
		ctx.Stroke()
	}

	draw(45, agg.Red, agg.CapButt)
	draw(90, agg.Green, agg.CapSquare)
	draw(135, agg.Blue, agg.CapRound)

	return ctx.GetImage().ToGoImage(), nil
}

func testLineJoinStyleComparison() (image.Image, error) {
	ctx := agg.NewContext(360, 220)
	ctx.Clear(agg.White)

	draw := func(x float64, c agg.Color, join agg.LineJoin) {
		ctx.SetColor(c)
		ctx.SetLineWidth(16)
		ctx.SetLineCap(agg.CapButt)
		ctx.SetLineJoin(join)
		ctx.ClearDashes()
		ctx.BeginPath()
		ctx.MoveTo(x, 170)
		ctx.LineTo(x+40, 50)
		ctx.LineTo(x+80, 170)
		ctx.Stroke()
	}

	draw(30, agg.Red, agg.JoinMiter)
	draw(140, agg.Green, agg.JoinRound)
	draw(250, agg.Blue, agg.JoinBevel)

	return ctx.GetImage().ToGoImage(), nil
}

func testDashPatternVariants() (image.Image, error) {
	ctx := agg.NewContext(340, 200)
	ctx.Clear(agg.White)

	draw := func(y float64, c agg.Color, pattern []float64) {
		ctx.SetColor(c)
		ctx.SetLineWidth(12)
		ctx.SetLineCap(agg.CapButt)
		ctx.SetLineJoin(agg.JoinMiter)
		ctx.SetDashPattern(pattern)
		ctx.BeginPath()
		ctx.MoveTo(30, y)
		ctx.LineTo(310, y)
		ctx.Stroke()
	}

	draw(45, agg.Black, []float64{28, 10})
	draw(100, agg.DarkGray, []float64{18, 8, 6, 8})
	draw(155, agg.Blue, []float64{6, 10})
	ctx.ClearDashes()

	return ctx.GetImage().ToGoImage(), nil
}

func testDashOffsetPhaseComparison() (image.Image, error) {
	ctx := agg.NewContext(340, 180)
	ctx.Clear(agg.White)

	draw := func(y float64, c agg.Color, offset float64) {
		ctx.SetColor(c)
		ctx.SetLineWidth(12)
		ctx.SetLineCap(agg.CapRound)
		ctx.SetLineJoin(agg.JoinRound)
		ctx.ClearDashes()
		ctx.AddDash(24, 12)
		ctx.SetDashOffset(offset)
		ctx.BeginPath()
		ctx.MoveTo(30, y)
		ctx.LineTo(310, y)
		ctx.Stroke()
	}

	draw(50, agg.Red, 0)
	draw(90, agg.Green, 8)
	draw(130, agg.Blue, 16)
	ctx.SetDashOffset(0)
	ctx.ClearDashes()

	return ctx.GetImage().ToGoImage(), nil
}

func testMiterLimitComparison() (image.Image, error) {
	ctx := agg.NewContext(360, 220)
	ctx.Clear(agg.White)

	draw := func(x float64, c agg.Color, miterLimit float64) {
		ctx.SetColor(c)
		ctx.SetLineWidth(18)
		ctx.SetLineCap(agg.CapButt)
		ctx.SetLineJoin(agg.JoinMiter)
		ctx.SetMiterLimit(miterLimit)
		ctx.ClearDashes()
		ctx.BeginPath()
		ctx.MoveTo(x, 185)
		ctx.LineTo(x+45, 35)
		ctx.LineTo(x+90, 185)
		ctx.Stroke()
	}

	draw(35, agg.Black, 1.2)
	draw(140, agg.DarkGray, 2.5)
	draw(245, agg.Red, 8.0)
	ctx.SetMiterLimit(4.0)

	return ctx.GetImage().ToGoImage(), nil
}

func testStrokeWidthRamp() (image.Image, error) {
	ctx := agg.NewContext(300, 220)
	ctx.Clear(agg.White)

	widths := []float64{1, 2, 4, 7, 11, 16}
	colors := []agg.Color{agg.Black, agg.DarkGray, agg.Blue, agg.Green, agg.Orange, agg.Red}

	for i, w := range widths {
		y := float64(20 + i*32)
		ctx.SetColor(colors[i])
		ctx.SetLineCap(agg.CapRound)
		ctx.SetLineJoin(agg.JoinRound)
		ctx.SetLineWidth(w)
		ctx.ClearDashes()
		ctx.BeginPath()
		ctx.MoveTo(30, y)
		ctx.LineTo(270, y)
		ctx.Stroke()
	}

	return ctx.GetImage().ToGoImage(), nil
}

func testSubpixelStrokeAlignment() (image.Image, error) {
	ctx := agg.NewContext(320, 200)
	ctx.Clear(agg.White)

	ctx.SetLineWidth(2.5)
	ctx.SetLineCap(agg.CapButt)
	ctx.SetLineJoin(agg.JoinMiter)
	ctx.ClearDashes()

	for i := 0; i < 8; i++ {
		ctx.SetColor(agg.Black)
		y := 20.5 + float64(i)*20.25
		ctx.BeginPath()
		ctx.MoveTo(20.5, y)
		ctx.LineTo(300.5, y+0.4)
		ctx.Stroke()
	}

	return ctx.GetImage().ToGoImage(), nil
}

func testClosedPathStrokeComparison() (image.Image, error) {
	ctx := agg.NewContext(340, 220)
	ctx.Clear(agg.White)

	drawPolygon := func(offsetX float64, c agg.Color, join agg.LineJoin) {
		ctx.SetColor(c)
		ctx.SetLineWidth(14)
		ctx.SetLineCap(agg.CapRound)
		ctx.SetLineJoin(join)
		ctx.ClearDashes()
		ctx.BeginPath()
		ctx.MoveTo(offsetX+40, 170)
		ctx.LineTo(offsetX+90, 45)
		ctx.LineTo(offsetX+140, 170)
		ctx.LineTo(offsetX+65, 95)
		ctx.LineTo(offsetX+115, 95)
		ctx.ClosePath()
		ctx.Stroke()
	}

	drawPolygon(0, agg.Blue, agg.JoinMiter)
	drawPolygon(170, agg.Green, agg.JoinRound)

	return ctx.GetImage().ToGoImage(), nil
}
