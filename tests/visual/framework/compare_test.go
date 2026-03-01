package framework

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func solidRGBA(width, height int, c color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, c)
		}
	}
	return img
}

func TestCompareImagesMaxDifferentPixelsThreshold(t *testing.T) {
	ref := solidRGBA(2, 2, color.RGBA{255, 0, 0, 255})
	gen := solidRGBA(2, 2, color.RGBA{255, 0, 0, 255})
	gen.SetRGBA(0, 0, color.RGBA{0, 0, 255, 255})

	result := CompareImages(ref, gen, ComparisonOptions{
		ExactMatch:         true,
		MaxDifferentPixels: 1,
	})
	if !result.Passed {
		t.Fatalf("expected comparison to pass with one differing pixel below threshold")
	}
	if result.DifferentPixels != 1 {
		t.Fatalf("DifferentPixels=%d, want 1", result.DifferentPixels)
	}
	if result.DifferentRatio != 0.25 {
		t.Fatalf("DifferentRatio=%v, want 0.25", result.DifferentRatio)
	}

	result = CompareImages(ref, gen, ComparisonOptions{
		ExactMatch:         true,
		MaxDifferentPixels: 0,
	})
	if result.Passed {
		t.Fatal("expected exact comparison with no pixel threshold to fail")
	}
}

func TestCompareImagesMaxDifferentRatioThreshold(t *testing.T) {
	ref := solidRGBA(4, 1, color.RGBA{255, 255, 255, 255})
	gen := solidRGBA(4, 1, color.RGBA{255, 255, 255, 255})
	gen.SetRGBA(0, 0, color.RGBA{0, 0, 0, 255})

	result := CompareImages(ref, gen, ComparisonOptions{
		ExactMatch:        true,
		MaxDifferentRatio: 0.25,
	})
	if !result.Passed {
		t.Fatal("expected comparison to pass at exact ratio threshold")
	}

	result = CompareImages(ref, gen, ComparisonOptions{
		ExactMatch:        true,
		MaxDifferentRatio: 0.20,
	})
	if result.Passed {
		t.Fatal("expected comparison to fail above ratio threshold")
	}
}

func TestNewTestRunnerReadsEnvironmentThresholds(t *testing.T) {
	t.Setenv("VISUAL_DIFF_TOLERANCE", "3")
	t.Setenv("VISUAL_MAX_DIFFERENT_PIXELS", "7")
	t.Setenv("VISUAL_MAX_DIFFERENT_RATIO", "0.125")
	t.Setenv("VISUAL_IGNORE_ALPHA", "true")
	t.Setenv("VISUAL_GENERATE_DIFFS", "false")

	runner := NewTestRunner("/tmp/visual")
	if runner.ComparisonOptions.ExactMatch {
		t.Fatal("expected ExactMatch to be disabled when VISUAL_DIFF_TOLERANCE is set")
	}
	if runner.ComparisonOptions.Tolerance != 3 {
		t.Fatalf("Tolerance=%d, want 3", runner.ComparisonOptions.Tolerance)
	}
	if runner.ComparisonOptions.MaxDifferentPixels != 7 {
		t.Fatalf("MaxDifferentPixels=%d, want 7", runner.ComparisonOptions.MaxDifferentPixels)
	}
	if runner.ComparisonOptions.MaxDifferentRatio != 0.125 {
		t.Fatalf("MaxDifferentRatio=%v, want 0.125", runner.ComparisonOptions.MaxDifferentRatio)
	}
	if !runner.ComparisonOptions.IgnoreAlpha {
		t.Fatal("expected IgnoreAlpha=true from environment")
	}
	if runner.ComparisonOptions.GenerateDiffImage {
		t.Fatal("expected GenerateDiffImage=false from environment")
	}
}

func TestGenerateReportIncludesDifferentRatio(t *testing.T) {
	dir := t.TempDir()
	reporter := NewReportGenerator(dir)
	suite := &TestSuite{
		Name: "ratio",
		Results: []TestResult{
			{
				Name: "example",
				Comparison: &ComparisonResult{
					Passed:          false,
					DifferentPixels: 1,
					DifferentRatio:  0.25,
					TotalPixels:     4,
				},
			},
		},
	}

	if err := reporter.GenerateReport(suite); err != nil {
		t.Fatalf("GenerateReport failed: %v", err)
	}

	reportPath := filepath.Join(dir, "ratio_report.html")
	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) failed: %v", reportPath, err)
	}
	if want := "Different ratio: 25.00%"; !contains(string(data), want) {
		t.Fatalf("report missing %q", want)
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
