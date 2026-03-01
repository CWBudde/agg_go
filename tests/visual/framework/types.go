// Package framework provides the core infrastructure for visual testing.
// This includes image comparison, test execution, and result reporting.
package framework

import (
	"image"
)

// ComparisonResult represents the outcome of comparing two images.
type ComparisonResult struct {
	// Passed indicates if the images match within tolerance
	Passed bool
	// DifferentPixels is the count of pixels that differ
	DifferentPixels int
	// DifferentRatio is the fraction of pixels that differ (0.0-1.0)
	DifferentRatio float64
	// TotalPixels is the total number of pixels compared
	TotalPixels int
	// MaxDifference is the maximum color channel difference found (0-255)
	MaxDifference uint8
	// AverageDifference is the average difference across all channels
	AverageDifference float64
	// DiffImage contains a visual representation of differences (optional)
	DiffImage *image.RGBA
}

// TestResult represents the result of a single visual test.
type TestResult struct {
	// Name is the test name/identifier
	Name string
	// Passed indicates if the test passed
	Passed bool
	// ReferencePath is the path to the reference image
	ReferencePath string
	// GeneratedPath is the path to the generated test image
	GeneratedPath string
	// DiffPath is the path to the difference image (if test failed)
	DiffPath string
	// Comparison contains detailed comparison results
	Comparison *ComparisonResult
	// Error contains any error that occurred during testing
	Error error
}

// TestSuite represents a collection of test results.
type TestSuite struct {
	// Name is the test suite name
	Name string
	// Results contains all test results
	Results []TestResult
	// StartTime when the test suite started
	StartTime string
	// Duration how long the tests took to run
	Duration string
}

// ComparisonOptions configures how images are compared.
type ComparisonOptions struct {
	// ExactMatch requires pixel-perfect matching (default: true)
	ExactMatch bool
	// Tolerance allows for small differences in color values (0-255)
	// Only used when ExactMatch is false
	Tolerance uint8
	// MaxDifferentPixels allows a bounded number of differing pixels to pass.
	// A zero value requires no differing pixels unless MaxDifferentRatio is set.
	MaxDifferentPixels int
	// MaxDifferentRatio allows a bounded fraction of differing pixels to pass.
	// A zero value requires no differing pixels unless MaxDifferentPixels is set.
	MaxDifferentRatio float64
	// GenerateDiffImage creates a visual diff image for failures
	GenerateDiffImage bool
	// IgnoreAlpha ignores alpha channel differences
	IgnoreAlpha bool
}

// DefaultComparisonOptions returns sensible defaults for image comparison.
func DefaultComparisonOptions() ComparisonOptions {
	return ComparisonOptions{
		ExactMatch:         true,
		Tolerance:          0,
		MaxDifferentPixels: 0,
		MaxDifferentRatio:  0,
		GenerateDiffImage:  true,
		IgnoreAlpha:        false,
	}
}
