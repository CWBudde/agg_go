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
		"dashed_round_cap_comparison": testDashedRoundCapComparison,
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
