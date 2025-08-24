package gamma

import (
	"fmt"
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/pixfmt"
)

const testEpsilon = 1e-9

func TestApplyGammaDirectRGB_BasicFunctionality(t *testing.T) {
	// Test with gamma = 2.2 (common display gamma)
	gamma := NewSimpleGammaLut(2.2)
	applicator := NewApplyGammaDirectRGB[color.Linear, color.RGB24Order](gamma)

	// Test pixel with mid-range values
	pixel := []basics.Int8u{128, 64, 192}
	original := make([]basics.Int8u, len(pixel))
	copy(original, pixel)

	applicator.Apply(pixel)

	// Gamma correction should reduce mid-range values (2.2 > 1.0)
	if pixel[0] >= original[0] || pixel[1] >= original[1] || pixel[2] >= original[2] {
		t.Errorf("Gamma 2.2 should reduce mid-range values, got %v from %v", pixel, original)
	}

	// Values should be different from original
	for i := 0; i < 3; i++ {
		if pixel[i] == original[i] {
			t.Errorf("Channel %d unchanged after gamma correction: %d", i, pixel[i])
		}
	}
}

func TestApplyGammaInverseRGB_BasicFunctionality(t *testing.T) {
	// Test with gamma = 2.2
	gamma := NewSimpleGammaLut(2.2)
	applicator := NewApplyGammaInverseRGB[color.Linear, color.RGB24Order](gamma)

	// Test pixel with mid-range values
	pixel := []basics.Int8u{128, 64, 192}
	original := make([]basics.Int8u, len(pixel))
	copy(original, pixel)

	applicator.Apply(pixel)

	// Inverse gamma correction should increase mid-range values (1/2.2 < 1.0)
	if pixel[0] <= original[0] || pixel[1] <= original[1] || pixel[2] <= original[2] {
		t.Errorf("Inverse gamma 2.2 should increase mid-range values, got %v from %v", pixel, original)
	}

	// Values should be different from original
	for i := 0; i < 3; i++ {
		if pixel[i] == original[i] {
			t.Errorf("Channel %d unchanged after inverse gamma correction: %d", i, pixel[i])
		}
	}
}

func TestApplyGammaRGB_RoundTrip(t *testing.T) {
	gammaValues := []float64{1.8, 2.2, 2.4}
	testPixels := [][]basics.Int8u{
		{0, 0, 0},       // Black
		{255, 255, 255}, // White
		{128, 128, 128}, // Mid gray
		{64, 128, 192},  // Mixed values
		{255, 0, 128},   // High contrast
	}

	for _, gammaVal := range gammaValues {
		t.Run(fmt.Sprintf("Gamma%.1f", gammaVal), func(t *testing.T) {
			gamma := NewSimpleGammaLut(gammaVal)
			dirApp := NewApplyGammaDirectRGB[color.Linear, color.RGB24Order](gamma)
			invApp := NewApplyGammaInverseRGB[color.Linear, color.RGB24Order](gamma)

			for _, original := range testPixels {
				// Make copies for testing
				directPixel := make([]basics.Int8u, len(original))
				copy(directPixel, original)

				// Apply direct gamma correction
				dirApp.Apply(directPixel)

				// Apply inverse gamma correction
				invApp.Apply(directPixel)

				// Check round-trip accuracy (allowing for quantization errors)
				for i := 0; i < 3; i++ {
					diff := int(directPixel[i]) - int(original[i])
					if diff < 0 {
						diff = -diff
					}
					// Allow up to 2 levels difference due to quantization
					if diff > 2 {
						t.Errorf("Round-trip error for gamma %.1f, channel %d: %d -> %d (diff: %d)",
							gammaVal, i, original[i], directPixel[i], diff)
					}
				}
			}
		})
	}
}

func TestApplyGammaRGB_LinearGamma(t *testing.T) {
	gamma := NewLinearGammaLut()
	dirApp := NewApplyGammaDirectRGB[color.Linear, color.RGB24Order](gamma)
	invApp := NewApplyGammaInverseRGB[color.Linear, color.RGB24Order](gamma)

	testPixels := [][]basics.Int8u{
		{0, 0, 0},
		{128, 64, 192},
		{255, 255, 255},
	}

	for _, original := range testPixels {
		// Test direct application
		directPixel := make([]basics.Int8u, len(original))
		copy(directPixel, original)
		dirApp.Apply(directPixel)

		for i := 0; i < 3; i++ {
			if directPixel[i] != original[i] {
				t.Errorf("Linear gamma should not change values: channel %d, %d != %d",
					i, directPixel[i], original[i])
			}
		}

		// Test inverse application
		inversePixel := make([]basics.Int8u, len(original))
		copy(inversePixel, original)
		invApp.Apply(inversePixel)

		for i := 0; i < 3; i++ {
			if inversePixel[i] != original[i] {
				t.Errorf("Linear inverse gamma should not change values: channel %d, %d != %d",
					i, inversePixel[i], original[i])
			}
		}
	}
}

func TestApplyGammaRGB_ColorOrdering(t *testing.T) {
	gamma := NewSimpleGammaLut(2.2)

	// Test RGB ordering
	rgbApp := NewApplyGammaDirectRGB[color.Linear, color.RGB24Order](gamma)
	rgbPixel := []basics.Int8u{100, 150, 200} // R=100, G=150, B=200
	rgbOriginal := make([]basics.Int8u, len(rgbPixel))
	copy(rgbOriginal, rgbPixel)
	rgbApp.Apply(rgbPixel)

	// Test BGR ordering
	bgrApp := NewApplyGammaDirectRGB[color.Linear, color.BGR24Order](gamma)
	bgrPixel := []basics.Int8u{200, 150, 100} // B=200, G=150, R=100 (same RGB values, different order)
	bgrOriginal := make([]basics.Int8u, len(bgrPixel))
	copy(bgrOriginal, bgrPixel)
	bgrApp.Apply(bgrPixel)

	// After gamma correction:
	// RGB pixel: [gamma(100), gamma(150), gamma(200)]
	// BGR pixel: [gamma(200), gamma(150), gamma(100)]
	// So: rgbPixel[0] should equal bgrPixel[2], rgbPixel[2] should equal bgrPixel[0]
	if rgbPixel[0] != bgrPixel[2] {
		t.Errorf("RGB R-channel and BGR B-channel should have same gamma correction: %d != %d",
			rgbPixel[0], bgrPixel[2])
	}
	if rgbPixel[2] != bgrPixel[0] {
		t.Errorf("RGB B-channel and BGR R-channel should have same gamma correction: %d != %d",
			rgbPixel[2], bgrPixel[0])
	}
	if rgbPixel[1] != bgrPixel[1] {
		t.Errorf("G-channel should be same for both orderings: %d != %d",
			rgbPixel[1], bgrPixel[1])
	}
}

func TestApplyGammaRGB_ShortBuffer(t *testing.T) {
	gamma := NewSimpleGammaLut(2.2)
	dirApp := NewApplyGammaDirectRGB[color.Linear, color.RGB24Order](gamma)

	// Test with buffer too short (should not panic)
	shortBuffer := []basics.Int8u{100, 150} // Only 2 bytes
	original := make([]basics.Int8u, len(shortBuffer))
	copy(original, shortBuffer)

	// Should not panic and should not modify the buffer
	dirApp.Apply(shortBuffer)

	for i, val := range shortBuffer {
		if val != original[i] {
			t.Errorf("Short buffer should not be modified: index %d, %d != %d", i, val, original[i])
		}
	}
}

func TestApplyGammaRGB_EdgeCases(t *testing.T) {
	gamma := NewSimpleGammaLut(2.2)
	dirApp := NewApplyGammaDirectRGB[color.Linear, color.RGB24Order](gamma)
	invApp := NewApplyGammaInverseRGB[color.Linear, color.RGB24Order](gamma)

	// Test edge values
	edgeCases := [][]basics.Int8u{
		{0, 0, 0},       // All black
		{255, 255, 255}, // All white
		{0, 255, 0},     // Pure green
		{255, 0, 255},   // Magenta
		{1, 1, 1},       // Near black
		{254, 254, 254}, // Near white
	}

	for i, original := range edgeCases {
		t.Run(fmt.Sprintf("EdgeCase%d", i), func(t *testing.T) {
			// Test direct gamma
			directPixel := make([]basics.Int8u, len(original))
			copy(directPixel, original)
			dirApp.Apply(directPixel)

			// Test inverse gamma
			inversePixel := make([]basics.Int8u, len(original))
			copy(inversePixel, original)
			invApp.Apply(inversePixel)

			// Ensure values are in valid range
			for j := 0; j < 3; j++ {
				if directPixel[j] > 255 {
					t.Errorf("Direct gamma produced invalid value: %d > 255", directPixel[j])
				}
				if inversePixel[j] > 255 {
					t.Errorf("Inverse gamma produced invalid value: %d > 255", inversePixel[j])
				}
			}
		})
	}
}

// Integration tests for PixFmtRGBGamma wrapper functionality
func TestPixFmtRGBGamma_Integration(t *testing.T) {
	// Create a rendering buffer for testing
	width, height := 10, 10
	bufferData := make([]basics.Int8u, width*height*3)

	// Fill with test pattern
	for i := 0; i < width*height; i++ {
		bufferData[i*3+0] = basics.Int8u(i % 256)       // R
		bufferData[i*3+1] = basics.Int8u((i * 2) % 256) // G
		bufferData[i*3+2] = basics.Int8u((i * 3) % 256) // B
	}

	// Create rendering buffer and pixel format
	rbuf := buffer.NewRenderingBufferU8WithData(bufferData, width, height, width*3)
	pixfmt := pixfmt.NewPixFmtRGB24(rbuf)

	// Test gamma wrapper creation
	gammaFormat := NewPixFmtRGB24Gamma(pixfmt, 2.2)

	// Test basic properties
	if gammaFormat.Width() != width {
		t.Errorf("Width should be %d, got %d", width, gammaFormat.Width())
	}
	if gammaFormat.Height() != height {
		t.Errorf("Height should be %d, got %d", height, gammaFormat.Height())
	}
	if gammaFormat.PixWidth() != 3 {
		t.Errorf("PixWidth should be 3 for RGB24, got %d", gammaFormat.PixWidth())
	}
}

func TestPixFmtRGBGamma_ConcreteTypes(t *testing.T) {
	// Create a dummy rendering buffer for testing
	bufferData := make([]basics.Int8u, 10*10*3)
	rbuf := buffer.NewRenderingBufferU8WithData(bufferData, 10, 10, 10*3)

	// Test that all the concrete type definitions work
	testCases := []struct {
		name   string
		create func() interface{}
	}{
		{"RGB24Gamma", func() interface{} {
			pixfmt := pixfmt.NewPixFmtRGB24(rbuf)
			return NewPixFmtRGB24Gamma(pixfmt, 2.2)
		}},
		{"RGB24GammaLinear", func() interface{} {
			pixfmt := pixfmt.NewPixFmtRGB24(rbuf)
			return NewPixFmtRGB24GammaLinear(pixfmt)
		}},
		{"BGR24Gamma", func() interface{} {
			pixfmt := pixfmt.NewPixFmtBGR24(rbuf)
			return NewPixFmtBGR24Gamma(pixfmt, 2.2)
		}},
		{"BGR24GammaLinear", func() interface{} {
			pixfmt := pixfmt.NewPixFmtBGR24(rbuf)
			return NewPixFmtBGR24GammaLinear(pixfmt)
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Constructor for %s panicked: %v", tc.name, r)
				}
			}()
			obj := tc.create()
			if obj == nil {
				t.Errorf("Constructor for %s returned nil", tc.name)
			}
		})
	}
}

func TestPixFmtRGBGamma_ApplyMethods(t *testing.T) {
	// Test the ApplyGammaDirect and ApplyGammaInverse methods
	// These should work on pixel formats that support ForEachPixel

	// Create a simple test buffer
	buffer := []basics.Int8u{100, 150, 200, 50, 75, 225}
	original := make([]basics.Int8u, len(buffer))
	copy(original, buffer)

	// Create a mock pixel format that supports ForEachPixel
	pixfmt := &mockPixFmtWithForEach{buffer: buffer}
	gammaFormat := NewPixFmtRGBGamma[*mockPixFmtWithForEach](pixfmt, NewSimpleGammaLut(2.2))

	// Apply direct gamma - should modify the buffer
	gammaFormat.ApplyGammaDirect()

	// Verify buffer was modified
	modified := false
	for i, val := range buffer {
		if val != original[i] {
			modified = true
			break
		}
	}
	if !modified {
		t.Error("ApplyGammaDirect should have modified the buffer")
	}

	// Apply inverse gamma - should restore closer to original
	gammaFormat.ApplyGammaInverse()

	// Check if values are closer to original (allowing for quantization error)
	for i := 0; i < len(buffer); i++ {
		diff := int(buffer[i]) - int(original[i])
		if diff < 0 {
			diff = -diff
		}
		if diff > 3 { // Allow some quantization error
			t.Errorf("Round-trip error too large at index %d: %d -> %d (diff: %d)",
				i, original[i], buffer[i], diff)
		}
	}
}

// Mock pixel format that supports ForEachPixel for testing
type mockPixFmtWithForEach struct {
	buffer []basics.Int8u
}

func (m *mockPixFmtWithForEach) ForEachPixel(fn func([]basics.Int8u)) {
	for i := 0; i < len(m.buffer); i += 3 {
		if i+2 < len(m.buffer) {
			pixel := m.buffer[i : i+3]
			fn(pixel)
		}
	}
}

// Benchmark tests for performance
func BenchmarkApplyGammaDirectRGB(b *testing.B) {
	gamma := NewSimpleGammaLut(2.2)
	app := NewApplyGammaDirectRGB[color.Linear, color.RGB24Order](gamma)
	pixel := []basics.Int8u{128, 128, 128}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.Apply(pixel)
	}
}

func BenchmarkApplyGammaInverseRGB(b *testing.B) {
	gamma := NewSimpleGammaLut(2.2)
	app := NewApplyGammaInverseRGB[color.Linear, color.RGB24Order](gamma)
	pixel := []basics.Int8u{128, 128, 128}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.Apply(pixel)
	}
}
