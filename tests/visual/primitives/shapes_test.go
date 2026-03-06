// Package primitives contains visual tests for basic geometric primitives.
package primitives

import (
	"image"
	"os"
	"testing"

	agg "agg_go"
)

// TestShapes runs visual tests for non-rectangle primitive rendering.
func TestShapes(t *testing.T) {
	runner := getTestRunner()
	tests := getShapeTests()

	suite := runner.RunTestSuite("shapes", tests)

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
		t.Fatalf("%d shape tests failed", failed)
	}
}

// TestGenerateShapeReferences generates reference images for shape tests.
func TestGenerateShapeReferences(t *testing.T) {
	if os.Getenv("GENERATE_REFERENCES") != "1" {
		t.Skip("Skipping reference generation (set GENERATE_REFERENCES=1 to enable)")
	}

	runner := getTestRunner()
	tests := getShapeTests()

	if err := runner.CreateReferenceImages(tests, "primitives"); err != nil {
		t.Fatalf("Failed to create reference images: %v", err)
	}

	t.Logf("Generated %d reference images for shape tests", len(tests))
}

func getShapeTests() map[string]func() (image.Image, error) {
	return map[string]func() (image.Image, error){
		"filled_circle_basic":      testFilledCircleBasic,
		"outlined_circle_thick":    testOutlinedCircleThick,
		"circle_subpixel_position": testCircleSubpixelPosition,
		"concentric_circles":       testConcentricCircles,
		"ellipse_fill_and_outline": testEllipseFillAndOutline,
		"crossed_lines_caps":       testCrossedLinesCaps,
		"triangle_fill_stroke":     testTriangleFillStroke,
		"star_path_fill":           testStarPathFill,
	}
}

func testFilledCircleBasic() (image.Image, error) {
	ctx := agg.NewContext(220, 180)
	ctx.Clear(agg.White)
	ctx.SetColor(agg.Blue)
	ctx.FillCircle(110, 90, 55)
	return ctx.GetImage().ToGoImage(), nil
}

func testOutlinedCircleThick() (image.Image, error) {
	ctx := agg.NewContext(240, 200)
	ctx.Clear(agg.White)
	ctx.SetColor(agg.Red)
	ctx.SetLineWidth(10)
	ctx.DrawCircle(120, 100, 60)
	return ctx.GetImage().ToGoImage(), nil
}

func testCircleSubpixelPosition() (image.Image, error) {
	ctx := agg.NewContext(220, 180)
	ctx.Clear(agg.White)
	ctx.SetColor(agg.Cyan)
	ctx.FillCircle(110.5, 90.5, 52.25)
	return ctx.GetImage().ToGoImage(), nil
}

func testConcentricCircles() (image.Image, error) {
	ctx := agg.NewContext(260, 220)
	ctx.Clear(agg.White)

	colors := []agg.Color{agg.Red, agg.Orange, agg.Yellow, agg.Green, agg.Blue}
	radius := 95.0
	for _, c := range colors {
		ctx.SetColor(c)
		ctx.SetLineWidth(8)
		ctx.DrawCircle(130, 110, radius)
		radius -= 16
	}

	return ctx.GetImage().ToGoImage(), nil
}

func testEllipseFillAndOutline() (image.Image, error) {
	ctx := agg.NewContext(300, 220)
	ctx.Clear(agg.White)

	ctx.SetColor(agg.RGBA(0.1, 0.5, 1.0, 0.7))
	ctx.FillEllipse(95, 110, 70, 45)

	ctx.SetColor(agg.Black)
	ctx.SetLineWidth(4)
	ctx.DrawEllipse(95, 110, 70, 45)

	ctx.SetColor(agg.RGBA(0.9, 0.2, 0.2, 0.65))
	ctx.FillEllipse(210, 110, 55, 75)

	ctx.SetColor(agg.DarkGray)
	ctx.SetLineWidth(3)
	ctx.DrawEllipse(210, 110, 55, 75)

	return ctx.GetImage().ToGoImage(), nil
}

func testCrossedLinesCaps() (image.Image, error) {
	ctx := agg.NewContext(300, 220)
	ctx.Clear(agg.White)
	ctx.SetLineWidth(18)
	ctx.ClearDashes()

	draw := func(x1, y1, x2, y2 float64, c agg.Color, cap agg.LineCap) {
		ctx.SetColor(c)
		ctx.SetLineCap(cap)
		ctx.BeginPath()
		ctx.MoveTo(x1, y1)
		ctx.LineTo(x2, y2)
		ctx.Stroke()
	}

	draw(35, 40, 265, 180, agg.Red, agg.CapButt)
	draw(35, 180, 265, 40, agg.Blue, agg.CapRound)
	draw(150, 20, 150, 200, agg.Green, agg.CapSquare)

	return ctx.GetImage().ToGoImage(), nil
}

func testTriangleFillStroke() (image.Image, error) {
	ctx := agg.NewContext(260, 220)
	ctx.Clear(agg.White)

	ctx.SetColor(agg.RGBA(0.9, 0.8, 0.1, 0.8))
	ctx.BeginPath()
	ctx.MoveTo(130, 25)
	ctx.LineTo(220, 185)
	ctx.LineTo(40, 185)
	ctx.ClosePath()
	ctx.Fill()

	ctx.SetColor(agg.Black)
	ctx.SetLineWidth(5)
	ctx.BeginPath()
	ctx.MoveTo(130, 25)
	ctx.LineTo(220, 185)
	ctx.LineTo(40, 185)
	ctx.ClosePath()
	ctx.Stroke()

	return ctx.GetImage().ToGoImage(), nil
}

func testStarPathFill() (image.Image, error) {
	ctx := agg.NewContext(280, 240)
	ctx.Clear(agg.White)

	ctx.SetColor(agg.Magenta)
	ctx.BeginPath()
	ctx.MoveTo(140, 25)
	ctx.LineTo(169, 95)
	ctx.LineTo(245, 95)
	ctx.LineTo(183, 140)
	ctx.LineTo(206, 215)
	ctx.LineTo(140, 168)
	ctx.LineTo(74, 215)
	ctx.LineTo(97, 140)
	ctx.LineTo(35, 95)
	ctx.LineTo(111, 95)
	ctx.ClosePath()
	ctx.Fill()

	ctx.SetColor(agg.Black)
	ctx.SetLineWidth(2)
	ctx.BeginPath()
	ctx.MoveTo(140, 25)
	ctx.LineTo(169, 95)
	ctx.LineTo(245, 95)
	ctx.LineTo(183, 140)
	ctx.LineTo(206, 215)
	ctx.LineTo(140, 168)
	ctx.LineTo(74, 215)
	ctx.LineTo(97, 140)
	ctx.LineTo(35, 95)
	ctx.LineTo(111, 95)
	ctx.ClosePath()
	ctx.Stroke()

	return ctx.GetImage().ToGoImage(), nil
}
