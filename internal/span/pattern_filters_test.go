// Package span provides pattern filter functionality tests for AGG.
package span

import (
	"testing"

	"agg_go/internal/color"
	"agg_go/internal/primitives"
)

// Test data for pattern filters
func createRGBA8TestPattern() [][]color.RGBA8[color.Linear] {
	// Create a simple 4x4 test pattern
	pattern := make([][]color.RGBA8[color.Linear], 4)
	for i := range pattern {
		pattern[i] = make([]color.RGBA8[color.Linear], 4)
	}

	// Fill with gradient pattern
	pattern[0][0] = color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255}
	pattern[0][1] = color.RGBA8[color.Linear]{R: 85, G: 0, B: 0, A: 255}
	pattern[0][2] = color.RGBA8[color.Linear]{R: 170, G: 0, B: 0, A: 255}
	pattern[0][3] = color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 255}

	pattern[1][0] = color.RGBA8[color.Linear]{R: 0, G: 85, B: 0, A: 255}
	pattern[1][1] = color.RGBA8[color.Linear]{R: 85, G: 85, B: 0, A: 255}
	pattern[1][2] = color.RGBA8[color.Linear]{R: 170, G: 85, B: 0, A: 255}
	pattern[1][3] = color.RGBA8[color.Linear]{R: 255, G: 85, B: 0, A: 255}

	pattern[2][0] = color.RGBA8[color.Linear]{R: 0, G: 170, B: 0, A: 255}
	pattern[2][1] = color.RGBA8[color.Linear]{R: 85, G: 170, B: 0, A: 255}
	pattern[2][2] = color.RGBA8[color.Linear]{R: 170, G: 170, B: 0, A: 255}
	pattern[2][3] = color.RGBA8[color.Linear]{R: 255, G: 170, B: 0, A: 255}

	pattern[3][0] = color.RGBA8[color.Linear]{R: 0, G: 255, B: 0, A: 255}
	pattern[3][1] = color.RGBA8[color.Linear]{R: 85, G: 255, B: 0, A: 255}
	pattern[3][2] = color.RGBA8[color.Linear]{R: 170, G: 255, B: 0, A: 255}
	pattern[3][3] = color.RGBA8[color.Linear]{R: 255, G: 255, B: 0, A: 255}

	return pattern
}

func createRGBA16TestPattern() [][]color.RGBA16[color.Linear] {
	// Create a simple 3x3 test pattern
	pattern := make([][]color.RGBA16[color.Linear], 3)
	for i := range pattern {
		pattern[i] = make([]color.RGBA16[color.Linear], 3)
	}

	// Fill with test values
	pattern[0][0] = color.RGBA16[color.Linear]{R: 0, G: 0, B: 0, A: 65535}
	pattern[0][1] = color.RGBA16[color.Linear]{R: 32768, G: 0, B: 0, A: 65535}
	pattern[0][2] = color.RGBA16[color.Linear]{R: 65535, G: 0, B: 0, A: 65535}

	pattern[1][0] = color.RGBA16[color.Linear]{R: 0, G: 32768, B: 0, A: 65535}
	pattern[1][1] = color.RGBA16[color.Linear]{R: 32768, G: 32768, B: 0, A: 65535}
	pattern[1][2] = color.RGBA16[color.Linear]{R: 65535, G: 32768, B: 0, A: 65535}

	pattern[2][0] = color.RGBA16[color.Linear]{R: 0, G: 65535, B: 0, A: 65535}
	pattern[2][1] = color.RGBA16[color.Linear]{R: 32768, G: 65535, B: 0, A: 65535}
	pattern[2][2] = color.RGBA16[color.Linear]{R: 65535, G: 65535, B: 0, A: 65535}

	return pattern
}

func createRGBA32TestPattern() [][]color.RGBA32[color.Linear] {
	// Create a simple 2x2 test pattern
	pattern := make([][]color.RGBA32[color.Linear], 2)
	for i := range pattern {
		pattern[i] = make([]color.RGBA32[color.Linear], 2)
	}

	pattern[0][0] = color.RGBA32[color.Linear]{R: 0.0, G: 0.0, B: 0.0, A: 1.0}
	pattern[0][1] = color.RGBA32[color.Linear]{R: 1.0, G: 0.0, B: 0.0, A: 1.0}
	pattern[1][0] = color.RGBA32[color.Linear]{R: 0.0, G: 1.0, B: 0.0, A: 1.0}
	pattern[1][1] = color.RGBA32[color.Linear]{R: 1.0, G: 1.0, B: 0.0, A: 1.0}

	return pattern
}

func TestPatternFilterNN_Creation(t *testing.T) {
	// Test creation of nearest neighbor filters
	filter8 := NewPatternFilterNN[color.RGBA8[color.Linear]]()
	if filter8 == nil {
		t.Fatal("Failed to create RGBA8 nearest neighbor filter")
	}

	filter16 := NewPatternFilterNN[color.RGBA16[color.Linear]]()
	if filter16 == nil {
		t.Fatal("Failed to create RGBA16 nearest neighbor filter")
	}

	filter32 := NewPatternFilterNN[color.RGBA32[color.Linear]]()
	if filter32 == nil {
		t.Fatal("Failed to create RGBA32 nearest neighbor filter")
	}

	// Test dilation values
	if filter8.Dilation() != 0 {
		t.Errorf("Expected dilation 0 for NN filter, got %d", filter8.Dilation())
	}
}

func TestPatternFilterNN_PixelLowRes(t *testing.T) {
	pattern := createRGBA8TestPattern()
	filter := NewPatternFilterNN[color.RGBA8[color.Linear]]()

	// Test valid coordinates
	var result color.RGBA8[color.Linear]
	filter.PixelLowRes(pattern, &result, 1, 1)

	expected := pattern[1][1]
	if result != expected {
		t.Errorf("Expected %+v, got %+v", expected, result)
	}

	// Test bounds - should not panic but not modify result
	originalResult := result
	filter.PixelLowRes(pattern, &result, 10, 10)
	if result != originalResult {
		t.Error("Out-of-bounds access should not modify result")
	}
}

func TestPatternFilterNN_PixelHighRes(t *testing.T) {
	pattern := createRGBA8TestPattern()
	filter := NewPatternFilterNN[color.RGBA8[color.Linear]]()

	// Test high-res coordinates (subpixel precision)
	var result color.RGBA8[color.Linear]

	// Test coordinates that map to pixel (1,1) in low resolution
	x := (1 << primitives.LineSubpixelShift) + (primitives.LineSubpixelScale / 2)
	y := (1 << primitives.LineSubpixelShift) + (primitives.LineSubpixelScale / 2)

	filter.PixelHighRes(pattern, &result, x, y)

	expected := pattern[1][1]
	if result != expected {
		t.Errorf("Expected %+v, got %+v", expected, result)
	}
}

func TestPatternFilterBilinearRGBA8_Creation(t *testing.T) {
	filter := NewPatternFilterBilinearRGBA8[color.Linear]()
	if filter == nil {
		t.Fatal("Failed to create bilinear RGBA8 filter")
	}

	if filter.Dilation() != 1 {
		t.Errorf("Expected dilation 1 for bilinear filter, got %d", filter.Dilation())
	}
}

func TestPatternFilterBilinearRGBA8_PixelLowRes(t *testing.T) {
	pattern := createRGBA8TestPattern()
	filter := NewPatternFilterBilinearRGBA8[color.Linear]()

	var result color.RGBA8[color.Linear]
	filter.PixelLowRes(pattern, &result, 2, 2)

	expected := pattern[2][2]
	if result != expected {
		t.Errorf("Expected %+v, got %+v", expected, result)
	}
}

func TestPatternFilterBilinearRGBA8_PixelHighRes(t *testing.T) {
	pattern := createRGBA8TestPattern()
	filter := NewPatternFilterBilinearRGBA8[color.Linear]()

	// Test exact pixel coordinate (should match low-res)
	var result color.RGBA8[color.Linear]
	x := 1 << primitives.LineSubpixelShift
	y := 1 << primitives.LineSubpixelShift
	filter.PixelHighRes(pattern, &result, x, y)

	// Should be the same as pattern[1][1]
	expected := pattern[1][1]
	if result != expected {
		t.Errorf("Exact coordinates: expected %+v, got %+v", expected, result)
	}

	// Test interpolation - halfway between four pixels
	x = (1 << primitives.LineSubpixelShift) + (primitives.LineSubpixelScale / 2)
	y = (1 << primitives.LineSubpixelShift) + (primitives.LineSubpixelScale / 2)
	filter.PixelHighRes(pattern, &result, x, y)

	// Should be average of pattern[1][1], pattern[1][2], pattern[2][1], pattern[2][2]
	// Each weight is (primitives.LineSubpixelScale/2)^2 = primitives.LineSubpixelScale^2/4
	// Total weight is primitives.LineSubpixelScale^2
	// After downshifting by (primitives.LineSubpixelShift * 2), we divide by primitives.LineSubpixelScale^2

	expectedR := (int(pattern[1][1].R) + int(pattern[1][2].R) + int(pattern[2][1].R) + int(pattern[2][2].R)) / 4
	expectedG := (int(pattern[1][1].G) + int(pattern[1][2].G) + int(pattern[2][1].G) + int(pattern[2][2].G)) / 4
	expectedB := (int(pattern[1][1].B) + int(pattern[1][2].B) + int(pattern[2][1].B) + int(pattern[2][2].B)) / 4
	expectedA := (int(pattern[1][1].A) + int(pattern[1][2].A) + int(pattern[2][1].A) + int(pattern[2][2].A)) / 4

	// Allow some tolerance for integer rounding
	tolerance := uint8(2)
	if absIntPF(int(result.R)-expectedR) > int(tolerance) ||
		absIntPF(int(result.G)-expectedG) > int(tolerance) ||
		absIntPF(int(result.B)-expectedB) > int(tolerance) ||
		absIntPF(int(result.A)-expectedA) > int(tolerance) {
		t.Errorf("Interpolation test: expected ~(%d,%d,%d,%d), got (%d,%d,%d,%d)",
			expectedR, expectedG, expectedB, expectedA,
			result.R, result.G, result.B, result.A)
	}
}

func TestPatternFilterBilinearRGBA8_BoundsChecking(t *testing.T) {
	pattern := createRGBA8TestPattern()
	filter := NewPatternFilterBilinearRGBA8[color.Linear]()

	var result color.RGBA8[color.Linear]
	originalResult := result

	// Test out-of-bounds coordinates
	filter.PixelHighRes(pattern, &result, -100, -100)
	if result != originalResult {
		t.Error("Out-of-bounds should return zero color")
	}

	// Test coordinates near the edge
	x := (len(pattern[0]) - 1) << primitives.LineSubpixelShift
	y := (len(pattern) - 1) << primitives.LineSubpixelShift
	filter.PixelHighRes(pattern, &result, x, y)

	// Should get the corner pixel
	expected := pattern[len(pattern)-1][len(pattern[0])-1]
	if result != expected {
		t.Errorf("Edge case: expected %+v, got %+v", expected, result)
	}
}

func TestPatternFilterBilinearRGBA16(t *testing.T) {
	pattern := createRGBA16TestPattern()
	filter := NewPatternFilterBilinearRGBA16[color.Linear]()

	var result color.RGBA16[color.Linear]

	// Test exact pixel sampling
	x := 1 << primitives.LineSubpixelShift
	y := 1 << primitives.LineSubpixelShift
	filter.PixelHighRes(pattern, &result, x, y)

	expected := pattern[1][1]
	if result != expected {
		t.Errorf("RGBA16 exact sampling: expected %+v, got %+v", expected, result)
	}
}

func TestPatternFilterBilinearRGBA32(t *testing.T) {
	pattern := createRGBA32TestPattern()
	filter := NewPatternFilterBilinearRGBA32[color.Linear]()

	var result color.RGBA32[color.Linear]

	// Test bilinear interpolation at center of 2x2 pattern
	x := (0 << primitives.LineSubpixelShift) + (primitives.LineSubpixelScale / 2)
	y := (0 << primitives.LineSubpixelShift) + (primitives.LineSubpixelScale / 2)
	filter.PixelHighRes(pattern, &result, x, y)

	// Should be average of all four pixels: (0.5, 0.5, 0, 1)
	expectedR := float32(0.5)
	expectedG := float32(0.5)
	expectedB := float32(0.0)
	expectedA := float32(1.0)

	tolerance := float32(0.01)
	if abs32(result.R-expectedR) > tolerance ||
		abs32(result.G-expectedG) > tolerance ||
		abs32(result.B-expectedB) > tolerance ||
		abs32(result.A-expectedA) > tolerance {
		t.Errorf("RGBA32 interpolation: expected (~%.2f,~%.2f,~%.2f,~%.2f), got (%.2f,%.2f,%.2f,%.2f)",
			expectedR, expectedG, expectedB, expectedA,
			result.R, result.G, result.B, result.A)
	}
}

func TestTypeAliases(t *testing.T) {
	// Test that type aliases work correctly
	filter1 := NewPatternFilterNN[color.RGBA8[color.Linear]]()
	if filter1.Dilation() != 0 {
		t.Error("Type alias PatternFilterNNRGBA8 not working correctly")
	}

	filter2 := NewPatternFilterNN[color.RGBA16[color.Linear]]()
	if filter2.Dilation() != 0 {
		t.Error("Type alias PatternFilterNNRGBA16 not working correctly")
	}

	filter3 := NewPatternFilterNN[color.RGBA32[color.Linear]]()
	if filter3.Dilation() != 0 {
		t.Error("Type alias PatternFilterNNRGBA32 not working correctly")
	}
}

func TestSubpixelConstants(t *testing.T) {
	// Verify our understanding of the subpixel constants
	if primitives.LineSubpixelShift != 8 {
		t.Errorf("Expected LineSubpixelShift=8, got %d", primitives.LineSubpixelShift)
	}

	if primitives.LineSubpixelScale != 256 {
		t.Errorf("Expected LineSubpixelScale=256, got %d", primitives.LineSubpixelScale)
	}

	if primitives.LineSubpixelMask != 255 {
		t.Errorf("Expected LineSubpixelMask=255, got %d", primitives.LineSubpixelMask)
	}
}

func TestBilinearWeights(t *testing.T) {
	// Test that bilinear weights sum to LineSubpixelScale^2
	x := primitives.LineSubpixelScale / 3
	y := primitives.LineSubpixelScale / 4

	w1 := (primitives.LineSubpixelScale - x) * (primitives.LineSubpixelScale - y)
	w2 := x * (primitives.LineSubpixelScale - y)
	w3 := (primitives.LineSubpixelScale - x) * y
	w4 := x * y

	totalWeight := w1 + w2 + w3 + w4
	expectedTotal := primitives.LineSubpixelScale * primitives.LineSubpixelScale

	if totalWeight != expectedTotal {
		t.Errorf("Bilinear weights don't sum correctly: got %d, expected %d", totalWeight, expectedTotal)
	}
}

// Helper functions
func absIntPF(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func abs32(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
