package effects

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
)

// mockPixelFormat is a simple mock implementation of PixFmtInterface for testing
type mockPixelFormat struct {
	width  int
	height int
	pixels [][]color.RGBA8[color.Linear]
}

func newMockPixelFormat(width, height int) *mockPixelFormat {
	pixels := make([][]color.RGBA8[color.Linear], height)
	for i := range pixels {
		pixels[i] = make([]color.RGBA8[color.Linear], width)
	}
	return &mockPixelFormat{
		width:  width,
		height: height,
		pixels: pixels,
	}
}

func (m *mockPixelFormat) Width() int {
	return m.width
}

func (m *mockPixelFormat) Height() int {
	return m.height
}

func (m *mockPixelFormat) GetPixel(x, y int) color.RGBA8[color.Linear] {
	if x < 0 || x >= m.width || y < 0 || y >= m.height {
		return color.RGBA8[color.Linear]{}
	}
	return m.pixels[y][x]
}

func (m *mockPixelFormat) CopyPixel(x, y int, c color.RGBA8[color.Linear]) {
	if x < 0 || x >= m.width || y < 0 || y >= m.height {
		return
	}
	m.pixels[y][x] = c
}

func TestSimpleStackBlur(t *testing.T) {
	// Create a simple test image
	pixels := make([][]color.RGBA8[color.Linear], 5)
	for i := range pixels {
		pixels[i] = make([]color.RGBA8[color.Linear], 5)
	}

	// Fill with a simple pattern - vertical line in the middle
	white := color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255}
	black := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255}

	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			if x == 2 {
				pixels[y][x] = white
			} else {
				pixels[y][x] = black
			}
		}
	}

	// Apply horizontal blur with radius 1
	blur := NewSimpleStackBlur()
	blur.BlurHorizontal(pixels, 1)

	// Check that the blur has spread the white line
	// The center should still be white, but neighboring pixels should be gray
	centerPixel := pixels[2][2]
	if centerPixel.R == 0 || centerPixel.G == 0 || centerPixel.B == 0 {
		t.Error("Center pixel should not be black after blur")
	}

	// The pixel next to the center should have some white mixed in
	leftPixel := pixels[2][1]
	if leftPixel.R == 0 && leftPixel.G == 0 && leftPixel.B == 0 {
		t.Error("Left pixel should have some blur effect")
	}
}

func TestSimpleStackBlurVertical(t *testing.T) {
	// Create a simple test image
	pixels := make([][]color.RGBA8[color.Linear], 5)
	for i := range pixels {
		pixels[i] = make([]color.RGBA8[color.Linear], 5)
	}

	// Fill with a simple pattern - horizontal line in the middle
	white := color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255}
	black := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255}

	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			if y == 2 {
				pixels[y][x] = white
			} else {
				pixels[y][x] = black
			}
		}
	}

	// Apply vertical blur with radius 1
	blur := NewSimpleStackBlur()
	blur.BlurVertical(pixels, 1)

	// Check that the blur has spread the white line vertically
	centerPixel := pixels[2][2]
	if centerPixel.R == 0 || centerPixel.G == 0 || centerPixel.B == 0 {
		t.Error("Center pixel should not be black after blur")
	}

	// The pixel above the center should have some white mixed in
	topPixel := pixels[1][2]
	if topPixel.R == 0 && topPixel.G == 0 && topPixel.B == 0 {
		t.Error("Top pixel should have some blur effect")
	}
}

func TestSimpleRecursiveBlur(t *testing.T) {
	// Create a simple test image
	pixels := make([][]color.RGBA8[color.Linear], 5)
	for i := range pixels {
		pixels[i] = make([]color.RGBA8[color.Linear], 5)
	}

	// Fill with a checkerboard pattern
	white := color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255}
	black := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255}

	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			if (x+y)%2 == 0 {
				pixels[y][x] = white
			} else {
				pixels[y][x] = black
			}
		}
	}

	// Apply recursive blur
	blur := NewSimpleRecursiveBlur()
	blur.BlurHorizontal(pixels, 2.0)

	// The blur should smooth out the checkerboard pattern
	// We can't test exact values, but we can check that it doesn't crash
	if blur == nil {
		t.Error("SimpleRecursiveBlur creation failed")
	}
}

func TestStackBlurRadius(t *testing.T) {
	pixels := make([][]color.RGBA8[color.Linear], 3)
	for i := range pixels {
		pixels[i] = make([]color.RGBA8[color.Linear], 3)
	}

	blur := NewSimpleStackBlur()

	// Test with zero radius (should be a no-op)
	blur.BlurHorizontal(pixels, 0)

	// Test with negative radius (should be a no-op)
	blur.BlurHorizontal(pixels, -1)

	// Test with large radius
	blur.BlurHorizontal(pixels, 100)
}

func TestRecursiveBlurRadius(t *testing.T) {
	pixels := make([][]color.RGBA8[color.Linear], 3)
	for i := range pixels {
		pixels[i] = make([]color.RGBA8[color.Linear], 3)
	}

	blur := NewSimpleRecursiveBlur()

	// Test with very small radius (should be a no-op)
	blur.BlurHorizontal(pixels, 0.1)

	// Test with normal radius
	blur.BlurHorizontal(pixels, 2.0)
}

func TestEmptyImage(t *testing.T) {
	// Test with empty image
	var pixels [][]color.RGBA8[color.Linear]

	blur := NewSimpleStackBlur()
	blur.BlurHorizontal(pixels, 1) // Should not crash

	recursiveBlur := NewSimpleRecursiveBlur()
	recursiveBlur.BlurHorizontal(pixels, 1.0) // Should not crash
}

func BenchmarkSimpleStackBlur(b *testing.B) {
	// Create a test image
	size := 100
	pixels := make([][]color.RGBA8[color.Linear], size)
	for i := range pixels {
		pixels[i] = make([]color.RGBA8[color.Linear], size)
		for j := range pixels[i] {
			pixels[i][j] = color.RGBA8[color.Linear]{
				R: uint8(i % 256),
				G: uint8(j % 256),
				B: uint8((i + j) % 256),
				A: 255,
			}
		}
	}

	blur := NewSimpleStackBlur()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		blur.BlurHorizontal(pixels, 5)
	}
}

func BenchmarkSimpleRecursiveBlur(b *testing.B) {
	// Create a test image
	size := 100
	pixels := make([][]color.RGBA8[color.Linear], size)
	for i := range pixels {
		pixels[i] = make([]color.RGBA8[color.Linear], size)
		for j := range pixels[i] {
			pixels[i][j] = color.RGBA8[color.Linear]{
				R: uint8(i % 256),
				G: uint8(j % 256),
				B: uint8((i + j) % 256),
				A: 255,
			}
		}
	}

	blur := NewSimpleRecursiveBlur()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		blur.BlurHorizontal(pixels, 3.0)
	}
}

func TestSlightBlur(t *testing.T) {
	// Create a 5x5 test image with a single white pixel in the center
	img := newMockPixelFormat(5, 5)

	// Fill with black pixels
	black := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255}
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			img.CopyPixel(x, y, black)
		}
	}

	// Set center pixel to white
	white := color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255}
	img.CopyPixel(2, 2, white)

	// Apply slight blur with default radius
	bounds := basics.RectI{X1: 0, Y1: 0, X2: 4, Y2: 4}
	ApplySlightBlur(img, bounds, 1.33)

	// The center pixel should still be relatively bright but not pure white
	center := img.GetPixel(2, 2)
	if center.R == 0 || center.G == 0 || center.B == 0 {
		t.Error("Center pixel should not be black after blur")
	}
	if center.R == 255 && center.G == 255 && center.B == 255 {
		t.Error("Center pixel should not be pure white after blur")
	}

	// Adjacent pixels should have some blur effect (not pure black)
	right := img.GetPixel(3, 2)
	if right.R == 0 && right.G == 0 && right.B == 0 {
		t.Error("Adjacent pixel should have some blur effect")
	}

	// Corner pixels should be less affected
	corner := img.GetPixel(0, 0)
	if corner.R > 50 || corner.G > 50 || corner.B > 50 {
		t.Error("Corner pixel should not be heavily affected by blur")
	}
}

func TestSlightBlurRadius(t *testing.T) {
	// Test with zero radius (should be no-op)
	img := newMockPixelFormat(3, 3)
	bounds := basics.RectI{X1: 0, Y1: 0, X2: 2, Y2: 2}

	// This should not crash
	ApplySlightBlur(img, bounds, 0)

	// Test with negative radius (should be no-op)
	ApplySlightBlur(img, bounds, -1.0)

	// Test with very small radius
	ApplySlightBlur(img, bounds, 0.1)
}

func TestSlightBlurFull(t *testing.T) {
	// Test the convenience function for full image blur
	img := newMockPixelFormat(3, 3)

	// Fill with a simple pattern
	white := color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255}
	img.CopyPixel(1, 1, white)

	// This should not crash
	ApplySlightBlurFull(img, 1.0)

	// Verify the pixel has been blurred
	center := img.GetPixel(1, 1)
	if center.R == 255 && center.G == 255 && center.B == 255 {
		t.Error("Center pixel should be blurred and not pure white")
	}
}

func TestSlightBlurSmallImage(t *testing.T) {
	// Test with very small images (should handle edge cases)
	img := newMockPixelFormat(1, 1)
	bounds := basics.RectI{X1: 0, Y1: 0, X2: 0, Y2: 0}

	// This should not crash - too small to blur
	ApplySlightBlur(img, bounds, 1.0)

	// Test with 2x2 image
	img2 := newMockPixelFormat(2, 2)
	bounds2 := basics.RectI{X1: 0, Y1: 0, X2: 1, Y2: 1}

	// This should not crash - still too small to blur effectively
	ApplySlightBlur(img2, bounds2, 1.0)
}
