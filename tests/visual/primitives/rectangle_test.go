// Package primitives contains visual tests for basic geometric primitives.
package primitives

import (
	"image"
	"math"
	"os"
	"path/filepath"
	"testing"

	agg "agg_go"
	"github.com/MeKo-Christian/agg_go/tests/visual/framework"
)

// TestRectangles runs all rectangle-related visual tests.
func TestRectangles(t *testing.T) {
	// Get the test runner
	runner := getTestRunner()

	// Define all rectangle tests
	tests := getRectangleTests()

	// Run the test suite
	suite := runner.RunTestSuite("rectangles", tests)

	// Check results
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

	// Print summary
	t.Logf("%s", runner.GetTestSummary(suite))

	if failed > 0 {
		t.Fatalf("%d rectangle tests failed", failed)
	}
}

// TestGenerateRectangleReferences generates reference images for rectangle tests.
// This test is meant to be run manually when setting up or updating references.
// Use: go test -run TestGenerateRectangleReferences ./tests/visual/primitives/
func TestGenerateRectangleReferences(t *testing.T) {
	if os.Getenv("GENERATE_REFERENCES") != "1" {
		t.Skip("Skipping reference generation (set GENERATE_REFERENCES=1 to enable)")
	}

	runner := getTestRunner()
	tests := getRectangleTests()

	if err := runner.CreateReferenceImages(tests, "primitives"); err != nil {
		t.Fatalf("Failed to create reference images: %v", err)
	}

	t.Logf("Generated %d reference images for rectangle tests", len(tests))
}

// getTestRunner creates a configured test runner for rectangle tests.
func getTestRunner() *framework.TestRunner {
	// Find the project root (where go.mod is located)
	projectRoot := findProjectRoot()
	visualTestDir := filepath.Join(projectRoot, "tests", "visual")

	return framework.NewTestRunner(visualTestDir)
}

// getRectangleTests returns a map of all rectangle test functions.
func getRectangleTests() map[string]func() (image.Image, error) {
	return map[string]func() (image.Image, error){
		"filled_rectangle_basic":        testFilledRectangleBasic,
		"outlined_rectangle_basic":      testOutlinedRectangleBasic,
		"rectangle_with_thick_stroke":   testRectangleWithThickStroke,
		"small_rectangle":               testSmallRectangle,
		"large_rectangle":               testLargeRectangle,
		"rectangle_subpixel_position":   testRectangleSubpixelPosition,
		"rectangle_different_colors":    testRectangleDifferentColors,
		"multiple_rectangles":           testMultipleRectangles,
		"rectangle_transparency":        testRectangleTransparency,
		"overlapping_rectangles":        testOverlappingRectangles,
		"hairline_rectangle_outline":    testHairlineRectangleOutline,
		"rectangle_negative_dimensions": testRectangleNegativeDimensions,
		"clipped_rectangles":            testClippedRectangles,
		"rounded_rectangle_fill":        testRoundedRectangleFill,
		"rounded_rectangle_outline":     testRoundedRectangleOutline,
		"transformed_rectangles_rotate": testTransformedRectanglesRotate,
		"transformed_rectangles_scale":  testTransformedRectanglesScale,
		"nested_transform_rectangles":   testNestedTransformRectangles,
		"thin_stroke_rectangle_grid":    testThinStrokeRectangleGrid,
	}
}

// Test functions for different rectangle scenarios

func testFilledRectangleBasic() (image.Image, error) {
	ctx := agg.NewContext(200, 150)
	ctx.Clear(agg.White)
	ctx.SetColor(agg.Blue)
	ctx.FillRectangle(50, 40, 100, 70)
	return ctx.GetImage().ToGoImage(), nil
}

func testOutlinedRectangleBasic() (image.Image, error) {
	ctx := agg.NewContext(200, 150)
	ctx.Clear(agg.White)
	ctx.SetColor(agg.Red)
	ctx.DrawRectangle(50, 40, 100, 70)
	return ctx.GetImage().ToGoImage(), nil
}

func testRectangleWithThickStroke() (image.Image, error) {
	ctx := agg.NewContext(200, 150)
	ctx.Clear(agg.White)
	ctx.SetColor(agg.Green)

	// Draw thick outlined rectangle using thick line function
	// Since there's no direct thick rectangle, draw as path
	x, y, w, h := 50.0, 40.0, 100.0, 70.0
	thickness := 5.0

	// Draw thick outline by drawing four thick lines
	ctx.DrawThickLine(x, y, x+w, y, thickness)     // top
	ctx.DrawThickLine(x+w, y, x+w, y+h, thickness) // right
	ctx.DrawThickLine(x+w, y+h, x, y+h, thickness) // bottom
	ctx.DrawThickLine(x, y+h, x, y, thickness)     // left

	return ctx.GetImage().ToGoImage(), nil
}

func testSmallRectangle() (image.Image, error) {
	ctx := agg.NewContext(100, 100)
	ctx.Clear(agg.White)
	ctx.SetColor(agg.Purple)
	ctx.FillRectangle(40, 40, 20, 20)
	return ctx.GetImage().ToGoImage(), nil
}

func testLargeRectangle() (image.Image, error) {
	ctx := agg.NewContext(400, 300)
	ctx.Clear(agg.White)
	ctx.SetColor(agg.Orange)
	ctx.FillRectangle(20, 20, 360, 260)
	return ctx.GetImage().ToGoImage(), nil
}

func testRectangleSubpixelPosition() (image.Image, error) {
	ctx := agg.NewContext(200, 150)
	ctx.Clear(agg.White)
	ctx.SetColor(agg.Cyan)
	// Use sub-pixel positioning
	ctx.FillRectangle(50.5, 40.5, 100, 70)
	return ctx.GetImage().ToGoImage(), nil
}

func testRectangleDifferentColors() (image.Image, error) {
	ctx := agg.NewContext(300, 200)
	ctx.Clear(agg.White)

	// Draw rectangles with different colors
	ctx.SetColor(agg.Red)
	ctx.FillRectangle(20, 20, 60, 60)

	ctx.SetColor(agg.Green)
	ctx.FillRectangle(100, 20, 60, 60)

	ctx.SetColor(agg.Blue)
	ctx.FillRectangle(180, 20, 60, 60)

	ctx.SetColor(agg.Yellow)
	ctx.FillRectangle(60, 100, 60, 60)

	ctx.SetColor(agg.Magenta)
	ctx.FillRectangle(140, 100, 60, 60)

	return ctx.GetImage().ToGoImage(), nil
}

func testMultipleRectangles() (image.Image, error) {
	ctx := agg.NewContext(250, 200)
	ctx.Clear(agg.White)

	// Draw a grid of rectangles
	colors := []agg.Color{agg.Red, agg.Green, agg.Blue, agg.Yellow, agg.Cyan, agg.Magenta}
	colorIndex := 0

	for row := 0; row < 3; row++ {
		for col := 0; col < 4; col++ {
			x := float64(col*60 + 10)
			y := float64(row*60 + 10)
			ctx.SetColor(colors[colorIndex%len(colors)])
			ctx.FillRectangle(x, y, 50, 50)
			colorIndex++
		}
	}

	return ctx.GetImage().ToGoImage(), nil
}

func testRectangleTransparency() (image.Image, error) {
	ctx := agg.NewContext(200, 150)
	ctx.Clear(agg.White)

	// Draw semi-transparent rectangles
	ctx.SetColor(agg.RGBA(1.0, 0.0, 0.0, 0.5)) // Semi-transparent red
	ctx.FillRectangle(30, 30, 80, 60)

	ctx.SetColor(agg.RGBA(0.0, 0.0, 1.0, 0.7)) // Semi-transparent blue
	ctx.FillRectangle(70, 50, 80, 60)

	return ctx.GetImage().ToGoImage(), nil
}

func testOverlappingRectangles() (image.Image, error) {
	ctx := agg.NewContext(250, 200)
	ctx.Clear(agg.White)

	// Draw overlapping opaque rectangles
	ctx.SetColor(agg.Red)
	ctx.FillRectangle(50, 50, 80, 80)

	ctx.SetColor(agg.Green)
	ctx.FillRectangle(90, 70, 80, 80)

	ctx.SetColor(agg.Blue)
	ctx.FillRectangle(70, 90, 80, 80)

	return ctx.GetImage().ToGoImage(), nil
}

func testHairlineRectangleOutline() (image.Image, error) {
	ctx := agg.NewContext(240, 180)
	ctx.Clear(agg.White)
	ctx.SetColor(agg.Black)
	// Hairline width 0 can be backend-dependent; use a very thin explicit width
	// so the visual test remains meaningful and stable across environments.
	ctx.SetLineWidth(0.6)

	for i := 0; i < 5; i++ {
		off := float64(i * 12)
		ctx.DrawRectangle(20+off, 20+off, 180-2*off, 120-2*off)
	}

	ctx.SetLineWidth(1.0)
	return ctx.GetImage().ToGoImage(), nil
}

func testRectangleNegativeDimensions() (image.Image, error) {
	ctx := agg.NewContext(280, 200)
	ctx.Clear(agg.White)

	ctx.SetColor(agg.Red)
	ctx.FillRectangle(200, 40, -90, 70)

	ctx.SetColor(agg.Blue)
	ctx.FillRectangle(90, 150, 120, -60)

	ctx.SetColor(agg.Black)
	ctx.SetLineWidth(2)
	ctx.DrawRectangle(250, 170, -160, -120)

	return ctx.GetImage().ToGoImage(), nil
}

func testClippedRectangles() (image.Image, error) {
	ctx := agg.NewContext(300, 220)
	ctx.Clear(agg.White)

	a := ctx.GetAgg2D()
	a.ClipBox(60, 40, 240, 180)

	ctx.SetColor(agg.Red)
	ctx.FillRectangle(10, 10, 160, 160)

	ctx.SetColor(agg.Blue)
	ctx.FillRectangle(140, 70, 150, 130)

	ctx.SetColor(agg.Black)
	ctx.SetLineWidth(3)
	ctx.DrawRectangle(40, 30, 220, 160)

	// Reset clip to full canvas for border indicator.
	a.ClipBox(0, 0, 300, 220)
	ctx.SetColor(agg.DarkGray)
	ctx.SetLineWidth(1)
	ctx.DrawRectangle(60, 40, 180, 140)

	return ctx.GetImage().ToGoImage(), nil
}

func testRoundedRectangleFill() (image.Image, error) {
	ctx := agg.NewContext(300, 220)
	ctx.Clear(agg.White)

	ctx.SetColor(agg.Cyan)
	ctx.FillRoundedRectangle(25, 25, 110, 80, 6)

	ctx.SetColor(agg.Green)
	ctx.FillRoundedRectangle(155, 30, 120, 70, 20)

	ctx.SetColor(agg.Orange)
	ctx.FillRoundedRectangle(70, 120, 170, 75, 34)

	return ctx.GetImage().ToGoImage(), nil
}

func testRoundedRectangleOutline() (image.Image, error) {
	ctx := agg.NewContext(320, 220)
	ctx.Clear(agg.White)

	ctx.SetLineWidth(4)

	ctx.SetColor(agg.Red)
	ctx.DrawRoundedRectangle(25, 25, 120, 90, 10)

	ctx.SetColor(agg.Blue)
	ctx.DrawRoundedRectangle(170, 30, 125, 80, 22)

	ctx.SetColor(agg.Black)
	ctx.DrawRoundedRectangle(60, 125, 210, 70, 30)

	return ctx.GetImage().ToGoImage(), nil
}

func testTransformedRectanglesRotate() (image.Image, error) {
	ctx := agg.NewContext(320, 220)
	ctx.Clear(agg.White)

	ctx.SetColor(agg.DarkGray)
	ctx.FillRectangle(135, 85, 50, 50)

	angles := []float64{-25, -10, 15, 35}
	colors := []agg.Color{agg.Red, agg.Green, agg.Blue, agg.Orange}
	for i, angle := range angles {
		ctx.PushTransform()
		ctx.Translate(160, 110)
		ctx.RotateDegrees(angle)
		ctx.Translate(-160, -110)
		ctx.SetColor(colors[i])
		ctx.FillRectangle(120, 90, 80, 40)
		ctx.PopTransform()
	}

	return ctx.GetImage().ToGoImage(), nil
}

func testTransformedRectanglesScale() (image.Image, error) {
	ctx := agg.NewContext(320, 220)
	ctx.Clear(agg.White)

	for i := 0; i < 5; i++ {
		s := 0.6 + float64(i)*0.18
		ctx.PushTransform()
		ctx.Translate(160, 110)
		ctx.Scale(s, 1.0+(float64(i)*0.1))
		ctx.Translate(-160, -110)
		ctx.SetColor(agg.RGBA(0.0, 0.2+float64(i)*0.15, 0.8-float64(i)*0.12, 0.75))
		ctx.FillRectangle(95, 70, 130, 80)
		ctx.PopTransform()
	}

	return ctx.GetImage().ToGoImage(), nil
}

func testNestedTransformRectangles() (image.Image, error) {
	ctx := agg.NewContext(320, 240)
	ctx.Clear(agg.White)
	ctx.SetLineWidth(2.0)

	ctx.PushTransform()
	ctx.Translate(160, 120)
	for i := 0; i < 8; i++ {
		theta := float64(i) * (math.Pi / 8)
		ctx.PushTransform()
		ctx.Rotate(theta)
		ctx.SetColor(agg.RGBA(0.1+float64(i)*0.1, 0.3, 0.9-float64(i)*0.09, 0.7))
		ctx.DrawRectangle(20, -18, 95, 36)
		ctx.PopTransform()
	}
	ctx.PopTransform()

	return ctx.GetImage().ToGoImage(), nil
}

func testThinStrokeRectangleGrid() (image.Image, error) {
	ctx := agg.NewContext(300, 220)
	ctx.Clear(agg.White)
	ctx.SetColor(agg.Black)

	widths := []float64{0.25, 0.5, 0.75, 1.0, 1.5, 2.0}
	idx := 0
	for row := 0; row < 2; row++ {
		for col := 0; col < 3; col++ {
			w := widths[idx]
			idx++
			x := float64(20 + col*95)
			y := float64(25 + row*95)
			ctx.SetLineWidth(w)
			ctx.DrawRectangle(x+0.5, y+0.5, 70, 70)
		}
	}

	return ctx.GetImage().ToGoImage(), nil
}

// findProjectRoot searches for the project root directory (containing go.mod).
func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		panic("Could not get working directory")
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached filesystem root
		}
		dir = parent
	}

	panic("Could not find project root (go.mod)")
}
