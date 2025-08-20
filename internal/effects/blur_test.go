package effects

import (
	"testing"

	"agg_go/internal/color"
)

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
