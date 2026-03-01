// Package framework provides test execution and result management for visual tests.
package framework

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// TestRunner manages the execution of visual tests.
type TestRunner struct {
	// BaseDir is the base directory for visual tests
	BaseDir string
	// ReferencesDir is the directory containing reference images
	ReferencesDir string
	// OutputDir is the directory for generated test images
	OutputDir string
	// DiffsDir is the directory for difference images
	DiffsDir string
	// ReportsDir is the directory for HTML reports
	ReportsDir string
	// ComparisonOptions configures how images are compared
	ComparisonOptions ComparisonOptions
}

// NewTestRunner creates a new test runner with default configuration.
func NewTestRunner(baseDir string) *TestRunner {
	runner := &TestRunner{
		BaseDir:           baseDir,
		ReferencesDir:     filepath.Join(baseDir, "reference"),
		OutputDir:         filepath.Join(baseDir, "output"),
		DiffsDir:          filepath.Join(baseDir, "diffs"),
		ReportsDir:        filepath.Join(baseDir, "reports"),
		ComparisonOptions: DefaultComparisonOptions(),
	}
	runner.applyEnvironmentOptions()
	return runner
}

// RunVisualTest executes a single visual test.
// The testFunc should generate an image and return it.
func (r *TestRunner) RunVisualTest(testName string, testFunc func() (image.Image, error)) TestResult {
	result := TestResult{
		Name: testName,
	}

	// Generate the test image
	generatedImage, err := testFunc()
	if err != nil {
		result.Error = fmt.Errorf("test function failed: %v", err)
		return result
	}

	// Save generated image
	result.GeneratedPath = filepath.Join(r.OutputDir, testName+".png")
	if err := SaveImage(generatedImage, result.GeneratedPath); err != nil {
		result.Error = fmt.Errorf("failed to save generated image: %v", err)
		return result
	}

	// Look for reference image
	result.ReferencePath = r.findReferenceImage(testName)
	if result.ReferencePath == "" {
		result.Error = fmt.Errorf("no reference image found for test %s", testName)
		return result
	}

	// Compare with reference
	comparison, err := CompareImageFiles(result.ReferencePath, result.GeneratedPath, r.ComparisonOptions)
	if err != nil {
		result.Error = fmt.Errorf("comparison failed: %v", err)
		return result
	}

	result.Comparison = comparison
	result.Passed = comparison.Passed

	// Save diff image if test failed and diff was generated
	if !comparison.Passed && comparison.DiffImage != nil {
		result.DiffPath = filepath.Join(r.DiffsDir, testName+"_diff.png")
		if err := SaveDiffImage(comparison.DiffImage, result.DiffPath); err != nil {
			// Don't fail the test for diff save errors, but note it
			fmt.Printf("Warning: failed to save diff image for %s: %v\n", testName, err)
		}
	}

	return result
}

// RunTestSuite executes multiple visual tests and returns a test suite result.
func (r *TestRunner) RunTestSuite(suiteName string, tests map[string]func() (image.Image, error)) *TestSuite {
	startTime := time.Now()

	suite := &TestSuite{
		Name:      suiteName,
		StartTime: startTime.Format(time.RFC3339),
		Results:   make([]TestResult, 0, len(tests)),
	}

	// Ensure output directories exist
	r.ensureDirectories()

	// Run each test
	for testName, testFunc := range tests {
		fmt.Printf("Running visual test: %s\n", testName)
		result := r.RunVisualTest(testName, testFunc)
		suite.Results = append(suite.Results, result)

		if result.Passed {
			fmt.Printf("  ✓ PASS\n")
		} else {
			fmt.Printf("  ✗ FAIL")
			if result.Error != nil {
				fmt.Printf(" - %v", result.Error)
			} else if result.Comparison != nil {
				fmt.Printf(" - %d/%d pixels different", result.Comparison.DifferentPixels, result.Comparison.TotalPixels)
			}
			fmt.Println()
		}
	}

	suite.Duration = time.Since(startTime).String()

	// Generate HTML report
	reportGen := NewReportGenerator(r.ReportsDir)
	if err := reportGen.GenerateReport(suite); err != nil {
		fmt.Printf("Warning: failed to generate HTML report: %v\n", err)
	}

	return suite
}

// CreateReferenceImages generates reference images from test functions.
// This should be used when setting up tests or updating references.
func (r *TestRunner) CreateReferenceImages(tests map[string]func() (image.Image, error), category string) error {
	referenceDir := filepath.Join(r.ReferencesDir, category)
	if err := os.MkdirAll(referenceDir, 0o755); err != nil {
		return fmt.Errorf("failed to create reference directory: %v", err)
	}

	for testName, testFunc := range tests {
		fmt.Printf("Creating reference image: %s\n", testName)

		// Generate the image
		img, err := testFunc()
		if err != nil {
			return fmt.Errorf("failed to generate reference image for %s: %v", testName, err)
		}

		// Save as reference
		referencePath := filepath.Join(referenceDir, testName+".png")
		if err := SaveImage(img, referencePath); err != nil {
			return fmt.Errorf("failed to save reference image for %s: %v", testName, err)
		}

		fmt.Printf("  Saved: %s\n", referencePath)
	}

	return nil
}

// GetTestSummary returns a summary of test results.
func (r *TestRunner) GetTestSummary(suite *TestSuite) string {
	passed := 0
	failed := 0
	errors := 0

	for _, result := range suite.Results {
		if result.Error != nil {
			errors++
		} else if result.Passed {
			passed++
		} else {
			failed++
		}
	}

	total := len(suite.Results)
	return fmt.Sprintf("Visual Tests: %d total, %d passed, %d failed, %d errors", total, passed, failed, errors)
}

// findReferenceImage searches for a reference image for the given test name.
func (r *TestRunner) findReferenceImage(testName string) string {
	// Try different possible locations
	possibilities := []string{
		filepath.Join(r.ReferencesDir, testName+".png"),
		filepath.Join(r.ReferencesDir, "primitives", testName+".png"),
		filepath.Join(r.ReferencesDir, "shapes", testName+".png"),
		filepath.Join(r.ReferencesDir, "basic", testName+".png"),
	}

	for _, path := range possibilities {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// ensureDirectories creates necessary directories if they don't exist.
func (r *TestRunner) ensureDirectories() {
	dirs := []string{r.OutputDir, r.DiffsDir, r.ReportsDir}
	for _, dir := range dirs {
		os.MkdirAll(dir, 0o755)
	}
}

// CleanOutputs removes generated test outputs (useful for cleanup).
func (r *TestRunner) CleanOutputs() error {
	dirs := []string{r.OutputDir, r.DiffsDir, r.ReportsDir}

	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("failed to clean directory %s: %v", dir, err)
		}
		// Recreate empty directory
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to recreate directory %s: %v", dir, err)
		}
	}

	return nil
}

func (r *TestRunner) applyEnvironmentOptions() {
	if value, ok := envUint8("VISUAL_DIFF_TOLERANCE"); ok {
		r.ComparisonOptions.ExactMatch = false
		r.ComparisonOptions.Tolerance = value
	}
	if value, ok := envInt("VISUAL_MAX_DIFFERENT_PIXELS"); ok {
		r.ComparisonOptions.MaxDifferentPixels = value
	}
	if value, ok := envFloat64("VISUAL_MAX_DIFFERENT_RATIO"); ok {
		r.ComparisonOptions.MaxDifferentRatio = value
	}
	if value, ok := envBool("VISUAL_IGNORE_ALPHA"); ok {
		r.ComparisonOptions.IgnoreAlpha = value
	}
	if value, ok := envBool("VISUAL_GENERATE_DIFFS"); ok {
		r.ComparisonOptions.GenerateDiffImage = value
	}
}

func envInt(name string) (int, bool) {
	raw, ok := os.LookupEnv(name)
	if !ok || raw == "" {
		return 0, false
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	return value, true
}

func envUint8(name string) (uint8, bool) {
	raw, ok := os.LookupEnv(name)
	if !ok || raw == "" {
		return 0, false
	}
	value, err := strconv.ParseUint(raw, 10, 8)
	if err != nil {
		return 0, false
	}
	return uint8(value), true
}

func envFloat64(name string) (float64, bool) {
	raw, ok := os.LookupEnv(name)
	if !ok || raw == "" {
		return 0, false
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, false
	}
	return value, true
}

func envBool(name string) (bool, bool) {
	raw, ok := os.LookupEnv(name)
	if !ok || raw == "" {
		return false, false
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, false
	}
	return value, true
}

// ListTests returns a list of available reference tests.
func (r *TestRunner) ListTests() ([]string, error) {
	var tests []string

	err := filepath.Walk(r.ReferencesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".png") {
			// Convert path to test name
			relPath, err := filepath.Rel(r.ReferencesDir, path)
			if err != nil {
				return err
			}
			testName := strings.TrimSuffix(relPath, ".png")
			testName = strings.ReplaceAll(testName, string(filepath.Separator), "_")
			tests = append(tests, testName)
		}

		return nil
	})

	return tests, err
}
