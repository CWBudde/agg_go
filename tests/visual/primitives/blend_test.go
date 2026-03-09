package primitives

import (
	"image"
	"os"
	"testing"

	agg "github.com/MeKo-Christian/agg_go"
)

// TestBlendModes runs visual tests for blend mode compositing.
func TestBlendModes(t *testing.T) {
	runner := getTestRunner()
	tests := getBlendTests()

	suite := runner.RunTestSuite("blends", tests)

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
		t.Fatalf("%d blend mode tests failed", failed)
	}
}

// TestGenerateBlendReferences generates reference images for blend mode tests.
// Use: GENERATE_REFERENCES=1 go test -run TestGenerateBlendReferences ./tests/visual/primitives/
func TestGenerateBlendReferences(t *testing.T) {
	if os.Getenv("GENERATE_REFERENCES") != "1" {
		t.Skip("Skipping reference generation (set GENERATE_REFERENCES=1 to enable)")
	}

	runner := getTestRunner()
	tests := getBlendTests()

	if err := runner.CreateReferenceImages(tests, "primitives"); err != nil {
		t.Fatalf("Failed to create reference images: %v", err)
	}

	t.Logf("Generated %d reference images for blend mode tests", len(tests))
}

func getBlendTests() map[string]func() (image.Image, error) {
	return map[string]func() (image.Image, error){
		"blend_src_over":   testBlendSrcOver,
		"blend_multiply":   testBlendMultiply,
		"blend_screen":     testBlendScreen,
		"blend_overlay":    testBlendOverlay,
		"blend_darken":     testBlendDarken,
		"blend_lighten":    testBlendLighten,
		"blend_difference": testBlendDifference,
		"blend_xor":        testBlendXor,
		"blend_add":        testBlendAdd,
		"global_alpha":     testGlobalAlpha,
	}
}

// blendScene is a helper that draws a base (red circle) then a foreground (blue
// rectangle) with the specified blend mode applied to the foreground layer.
func blendScene(mode agg.BlendMode) (image.Image, error) {
	ctx := agg.NewContext(200, 160)
	ctx.Clear(agg.White)

	// Base layer: red circle, drawn with normal SrcOver
	ctx.SetBlendMode(agg.BlendSrcOver)
	ctx.SetColor(agg.Red)
	ctx.FillCircle(80, 80, 65)

	// Foreground: semi-transparent blue rectangle with the given blend mode
	ctx.SetBlendMode(mode)
	ctx.SetColor(agg.RGBA(0.0, 0.3, 1.0, 0.7))
	ctx.FillRectangle(60, 40, 100, 90)

	// Reset blend mode
	ctx.SetBlendMode(agg.BlendSrcOver)

	return ctx.GetImage().ToGoImage(), nil
}

func testBlendSrcOver() (image.Image, error)    { return blendScene(agg.BlendSrcOver) }
func testBlendMultiply() (image.Image, error)   { return blendScene(agg.BlendMultiply) }
func testBlendScreen() (image.Image, error)     { return blendScene(agg.BlendScreen) }
func testBlendOverlay() (image.Image, error)    { return blendScene(agg.BlendOverlay) }
func testBlendDarken() (image.Image, error)     { return blendScene(agg.BlendDarken) }
func testBlendLighten() (image.Image, error)    { return blendScene(agg.BlendLighten) }
func testBlendDifference() (image.Image, error) { return blendScene(agg.BlendDifference) }
func testBlendXor() (image.Image, error)        { return blendScene(agg.BlendXor) }
func testBlendAdd() (image.Image, error)        { return blendScene(agg.BlendAdd) }

// testGlobalAlpha renders a checkerboard-like scene to verify that SetGlobalAlpha
// scales the alpha of every subsequent draw call uniformly.
func testGlobalAlpha() (image.Image, error) {
	ctx := agg.NewContext(240, 180)
	ctx.Clear(agg.Black)

	// Draw a bright background pattern at full opacity
	ctx.SetGlobalAlpha(1.0)
	colors := []agg.Color{agg.Red, agg.Green, agg.Blue, agg.Yellow}
	for i, c := range colors {
		x := float64((i % 2) * 120)
		y := float64((i / 2) * 90)
		ctx.SetColor(c)
		ctx.FillRectangle(x, y, 120, 90)
	}

	// Overlay a white rectangle at half global alpha — should be faint.
	ctx.SetGlobalAlpha(0.35)
	ctx.SetColor(agg.White)
	ctx.FillRectangle(60, 45, 120, 90)

	ctx.SetGlobalAlpha(1.0)
	return ctx.GetImage().ToGoImage(), nil
}
