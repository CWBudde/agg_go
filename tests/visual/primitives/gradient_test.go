package primitives

import (
	"image"
	"math"
	"os"
	"testing"

	agg "github.com/MeKo-Christian/agg_go"
)

// TestGradients runs visual tests for gradient rendering.
func TestGradients(t *testing.T) {
	runner := getTestRunner()
	tests := getGradientTests()

	suite := runner.RunTestSuite("gradients", tests)

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
		t.Fatalf("%d gradient tests failed", failed)
	}
}

// TestGenerateGradientReferences generates reference images for gradient tests.
// Use: GENERATE_REFERENCES=1 go test -run TestGenerateGradientReferences ./tests/visual/primitives/
func TestGenerateGradientReferences(t *testing.T) {
	if os.Getenv("GENERATE_REFERENCES") != "1" {
		t.Skip("Skipping reference generation (set GENERATE_REFERENCES=1 to enable)")
	}

	runner := getTestRunner()
	tests := getGradientTests()

	if err := runner.CreateReferenceImages(tests, "primitives"); err != nil {
		t.Fatalf("Failed to create reference images: %v", err)
	}

	t.Logf("Generated %d reference images for gradient tests", len(tests))
}

func getGradientTests() map[string]func() (image.Image, error) {
	return map[string]func() (image.Image, error){
		"linear_gradient_horizontal":     testLinearGradientHorizontal,
		"linear_gradient_vertical":       testLinearGradientVertical,
		"linear_gradient_diagonal":       testLinearGradientDiagonal,
		"linear_gradient_narrow_profile": testLinearGradientNarrowProfile,
		"radial_gradient_centered":       testRadialGradientCentered,
		"radial_gradient_off_center":     testRadialGradientOffCenter,
		"radial_gradient_multi_stop":     testRadialGradientMultiStop,
		"gradient_on_triangle":           testGradientOnTriangle,
		"multiple_gradient_fills":        testMultipleGradientFills,
		"radial_gradient_transparency":   testRadialGradientTransparency,
	}
}

// testLinearGradientHorizontal renders a rectangle with a left-to-right red→blue gradient.
func testLinearGradientHorizontal() (image.Image, error) {
	ctx := agg.NewContext(240, 120)
	ctx.Clear(agg.White)
	ctx.SetLinearGradient(20, 0, 220, 0, agg.Red, agg.Blue)
	ctx.FillRectangle(20, 20, 200, 80)
	return ctx.GetImage().ToGoImage(), nil
}

// testLinearGradientVertical renders a rectangle with a top-to-bottom white→black gradient.
func testLinearGradientVertical() (image.Image, error) {
	ctx := agg.NewContext(200, 200)
	ctx.Clear(agg.White)
	ctx.SetLinearGradient(0, 20, 0, 180, agg.White, agg.Black)
	ctx.FillRectangle(20, 20, 160, 160)
	return ctx.GetImage().ToGoImage(), nil
}

// testLinearGradientDiagonal renders a diagonal cyan→magenta gradient on a square.
func testLinearGradientDiagonal() (image.Image, error) {
	ctx := agg.NewContext(200, 200)
	ctx.Clear(agg.White)
	ctx.SetLinearGradient(20, 20, 180, 180, agg.Cyan, agg.Magenta)
	ctx.FillRectangle(20, 20, 160, 160)
	return ctx.GetImage().ToGoImage(), nil
}

// testLinearGradientNarrowProfile renders a gradient with a sharp (narrow) profile,
// producing a hard-edged color band rather than a smooth ramp.
func testLinearGradientNarrowProfile() (image.Image, error) {
	ctx := agg.NewContext(240, 120)
	ctx.Clear(agg.White)
	// profile > 1 concentrates the gradient toward center, < 1 widens it
	ctx.SetLinearGradientWithProfile(20, 0, 220, 0, agg.Red, agg.Blue, 0.25)
	ctx.FillRectangle(20, 20, 200, 80)
	return ctx.GetImage().ToGoImage(), nil
}

// testRadialGradientCentered renders a filled circle with a yellow-center→transparent radial gradient.
func testRadialGradientCentered() (image.Image, error) {
	ctx := agg.NewContext(200, 200)
	ctx.Clear(agg.DarkGray)
	ctx.SetRadialGradient(100, 100, 80, agg.Yellow, agg.Transparent)
	ctx.FillCircle(100, 100, 80)
	return ctx.GetImage().ToGoImage(), nil
}

// testRadialGradientOffCenter places the center of the gradient off-center
// relative to the circle, creating an asymmetric illumination effect.
func testRadialGradientOffCenter() (image.Image, error) {
	ctx := agg.NewContext(240, 200)
	ctx.Clear(agg.Black)
	ctx.SetRadialGradientWithProfile(80, 70, 100, agg.White, agg.Black, 1.0)
	ctx.FillCircle(120, 100, 100)
	return ctx.GetImage().ToGoImage(), nil
}

// testRadialGradientMultiStop renders a three-color radial gradient (green→white→red).
func testRadialGradientMultiStop() (image.Image, error) {
	ctx := agg.NewContext(200, 200)
	ctx.Clear(agg.White)
	ctx.SetRadialGradientMultiStop(100, 100, 80, agg.Green, agg.White, agg.Red)
	ctx.FillCircle(100, 100, 80)
	return ctx.GetImage().ToGoImage(), nil
}

// testGradientOnTriangle fills a triangle with a diagonal gradient to verify
// that gradient fills work on non-rectangular paths.
func testGradientOnTriangle() (image.Image, error) {
	ctx := agg.NewContext(200, 200)
	ctx.Clear(agg.White)
	ctx.SetLinearGradient(40, 160, 160, 40, agg.Orange, agg.Purple)

	a := ctx.GetAgg2D()
	a.ResetPath()
	a.MoveTo(100, 20)
	a.LineTo(180, 170)
	a.LineTo(20, 170)
	a.ClosePolygon()
	a.DrawPath(agg.FillOnly)

	return ctx.GetImage().ToGoImage(), nil
}

// testMultipleGradientFills draws several shapes each with their own gradient fill.
func testMultipleGradientFills() (image.Image, error) {
	ctx := agg.NewContext(320, 200)
	ctx.Clear(agg.White)

	// Left: horizontal gradient rectangle
	ctx.SetLinearGradient(20, 0, 120, 0, agg.Red, agg.Yellow)
	ctx.FillRectangle(20, 40, 100, 120)

	// Middle: vertical gradient circle
	ctx.SetLinearGradient(0, 40, 0, 160, agg.Blue, agg.Cyan)
	ctx.FillCircle(200, 100, 55)

	// Right: radial gradient
	ctx.SetRadialGradient(280, 100, 40, agg.Green, agg.White)
	ctx.FillCircle(280, 100, 40)

	return ctx.GetImage().ToGoImage(), nil
}

// testRadialGradientTransparency draws a semi-transparent radial gradient over
// a colored background to verify alpha compositing with gradient fills.
func testRadialGradientTransparency() (image.Image, error) {
	ctx := agg.NewContext(200, 200)
	ctx.Clear(agg.Blue)

	// Draw a grid of lines as background pattern
	ctx.SetColor(agg.RGBA(1.0, 1.0, 1.0, 0.3))
	for i := 0; i < 10; i++ {
		x := float64(i * 20)
		ctx.DrawThickLine(x, 0, x, 200, 1)
		ctx.DrawThickLine(0, x, 200, x, 1)
	}

	// Radial gradient from opaque white center to transparent
	ctx.SetRadialGradient(100, 100, 75,
		agg.RGBA(1.0, 1.0, 0.0, 0.9),
		agg.RGBA(1.0, 0.5, 0.0, 0.0))
	// Draw a star to show gradient clipping to path
	a := ctx.GetAgg2D()
	a.ResetPath()
	for i := 0; i < 10; i++ {
		angle := float64(i)*math.Pi/5 - math.Pi/2
		var r float64
		if i%2 == 0 {
			r = 75
		} else {
			r = 35
		}
		x := 100 + r*math.Cos(angle)
		y := 100 + r*math.Sin(angle)
		if i == 0 {
			a.MoveTo(x, y)
		} else {
			a.LineTo(x, y)
		}
	}
	a.ClosePolygon()
	a.DrawPath(agg.FillOnly)

	return ctx.GetImage().ToGoImage(), nil
}
