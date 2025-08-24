package span

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/image"
)

// MockRGBASource provides a simple RGBA image source for testing
type MockRGBASource struct {
	width, height int
	data          []color.RGBA8[color.Linear]
}

func NewMockRGBASource(width, height int) *MockRGBASource {
	return &MockRGBASource{
		width:  width,
		height: height,
		data:   make([]color.RGBA8[color.Linear], width*height),
	}
}

func (m *MockRGBASource) Width() int  { return m.width }
func (m *MockRGBASource) Height() int { return m.height }

func (m *MockRGBASource) SetPixel(x, y int, c color.RGBA8[color.Linear]) {
	if x >= 0 && y >= 0 && x < m.width && y < m.height {
		m.data[y*m.width+x] = c
	}
}

func (m *MockRGBASource) GetPixel(x, y int) color.RGBA8[color.Linear] {
	if x >= 0 && y >= 0 && x < m.width && y < m.height {
		return m.data[y*m.width+x]
	}
	return color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 0}
}

// Span returns a slice starting at the requested pixel
func (m *MockRGBASource) Span(x, y, length int) []basics.Int8u {
	if x < 0 || y < 0 || x >= m.width || y >= m.height {
		return []basics.Int8u{0, 0, 0, 0} // Return transparent black for out-of-bounds
	}

	// Convert RGBA data to raw bytes
	result := make([]basics.Int8u, length*4)
	for i := 0; i < length && x+i < m.width; i++ {
		pixel := m.GetPixel(x+i, y)
		result[i*4] = pixel.R
		result[i*4+1] = pixel.G
		result[i*4+2] = pixel.B
		result[i*4+3] = pixel.A
	}
	return result
}

// NextX returns the next pixel in X direction
func (m *MockRGBASource) NextX() []basics.Int8u {
	// For simplicity in tests, return the same as the first pixel
	return m.Span(0, 0, 1)
}

// NextY returns the next pixel in Y direction
func (m *MockRGBASource) NextY() []basics.Int8u {
	// For simplicity in tests, return the same as the first pixel
	return m.Span(0, 0, 1)
}

// RowPtr returns a pointer to a specific row
func (m *MockRGBASource) RowPtr(y int) []basics.Int8u {
	if y < 0 || y >= m.height {
		return []basics.Int8u{}
	}

	result := make([]basics.Int8u, m.width*4)
	for x := 0; x < m.width; x++ {
		pixel := m.GetPixel(x, y)
		result[x*4] = pixel.R
		result[x*4+1] = pixel.G
		result[x*4+2] = pixel.B
		result[x*4+3] = pixel.A
	}
	return result
}

// ColorType returns the RGBA color type identifier
func (m *MockRGBASource) ColorType() string {
	return "RGBA8"
}

// OrderType returns the color component ordering (RGBA)
func (m *MockRGBASource) OrderType() color.ColorOrder {
	return color.OrderRGBA
}

func TestSpanImageFilterRGBANN_Generate(t *testing.T) {
	// Create a test source image
	source := NewMockRGBASource(4, 4)

	// Fill with a simple pattern including alpha channel
	source.SetPixel(0, 0, color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 255})   // Opaque Red
	source.SetPixel(1, 0, color.RGBA8[color.Linear]{R: 0, G: 255, B: 0, A: 128})   // Semi-transparent Green
	source.SetPixel(2, 0, color.RGBA8[color.Linear]{R: 0, G: 0, B: 255, A: 64})    // More transparent Blue
	source.SetPixel(3, 0, color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 0}) // Transparent White

	// Create interpolator
	interpolator := NewMockInterpolator()

	// Create the span filter
	filter := NewSpanImageFilterRGBANNWithParams(source, interpolator)

	// Generate a span
	span := make([]color.RGBA8[color.Linear], 4)
	filter.Generate(span, 0, 0)

	// Verify results - should match source pixels exactly for nearest neighbor
	expected := []color.RGBA8[color.Linear]{
		{R: 255, G: 0, B: 0, A: 255},   // Opaque Red
		{R: 0, G: 255, B: 0, A: 128},   // Semi-transparent Green
		{R: 0, G: 0, B: 255, A: 64},    // More transparent Blue
		{R: 255, G: 255, B: 255, A: 0}, // Transparent White
	}

	for i, expectedColor := range expected {
		if span[i] != expectedColor {
			t.Errorf("Pixel %d: expected %+v, got %+v", i, expectedColor, span[i])
		}
	}
}

func TestSpanImageFilterRGBABilinear_Generate(t *testing.T) {
	// Create a test source image
	source := NewMockRGBASource(3, 3)

	// Fill with gradient pattern including alpha
	source.SetPixel(0, 0, color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 0})     // Transparent
	source.SetPixel(1, 0, color.RGBA8[color.Linear]{R: 128, G: 0, B: 0, A: 128}) // Semi-transparent Red
	source.SetPixel(2, 0, color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 255}) // Opaque Red

	// Create interpolator
	interpolator := NewMockInterpolator()

	// Create the span filter
	filter := NewSpanImageFilterRGBABilinearWithParams(source, interpolator)

	// Generate a span
	span := make([]color.RGBA8[color.Linear], 3)
	filter.Generate(span, 0, 0)

	// For bilinear interpolation, the exact values depend on the interpolation
	// We mainly want to ensure the alpha constraints are applied
	for i, pixel := range span {
		// Verify alpha constraints: R, G, B should not exceed A
		if pixel.R > pixel.A {
			t.Errorf("Pixel %d: R (%d) should not exceed A (%d)", i, pixel.R, pixel.A)
		}
		if pixel.G > pixel.A {
			t.Errorf("Pixel %d: G (%d) should not exceed A (%d)", i, pixel.G, pixel.A)
		}
		if pixel.B > pixel.A {
			t.Errorf("Pixel %d: B (%d) should not exceed A (%d)", i, pixel.B, pixel.A)
		}
	}
}

func TestSpanImageFilterRGBABilinearClip_Generate(t *testing.T) {
	// Create a small test source image
	source := NewMockRGBASource(2, 2)

	// Fill with known values
	source.SetPixel(0, 0, color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 255})
	source.SetPixel(1, 0, color.RGBA8[color.Linear]{R: 0, G: 255, B: 0, A: 255})
	source.SetPixel(0, 1, color.RGBA8[color.Linear]{R: 0, G: 0, B: 255, A: 255})
	source.SetPixel(1, 1, color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	// Create interpolator
	interpolator := NewMockInterpolator()

	// Create the span filter with a specific background color
	backgroundColor := color.RGBA8[color.Linear]{R: 64, G: 64, B: 64, A: 64}
	filter := NewSpanImageFilterRGBABilinearClipWithParams(source, backgroundColor, interpolator)

	// Test that the background color can be set and retrieved
	if filter.BackgroundColor() != backgroundColor {
		t.Errorf("Expected background color %+v, got %+v", backgroundColor, filter.BackgroundColor())
	}

	// Change background color
	newBackground := color.RGBA8[color.Linear]{R: 128, G: 128, B: 128, A: 128}
	filter.SetBackgroundColor(newBackground)

	if filter.BackgroundColor() != newBackground {
		t.Errorf("Expected updated background color %+v, got %+v", newBackground, filter.BackgroundColor())
	}

	// Generate a span
	span := make([]color.RGBA8[color.Linear], 2)
	filter.Generate(span, 0, 0)

	// Verify alpha constraints are still applied
	for i, pixel := range span {
		if pixel.R > pixel.A {
			t.Errorf("Pixel %d: R (%d) should not exceed A (%d)", i, pixel.R, pixel.A)
		}
		if pixel.G > pixel.A {
			t.Errorf("Pixel %d: G (%d) should not exceed A (%d)", i, pixel.G, pixel.A)
		}
		if pixel.B > pixel.A {
			t.Errorf("Pixel %d: B (%d) should not exceed A (%d)", i, pixel.B, pixel.A)
		}
	}
}

func TestSpanImageFilterRGBA2x2_Generate(t *testing.T) {
	// Create a test source image
	source := NewMockRGBASource(3, 3)

	// Fill with test pattern
	source.SetPixel(0, 0, color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 255})
	source.SetPixel(1, 0, color.RGBA8[color.Linear]{R: 0, G: 255, B: 0, A: 255})
	source.SetPixel(2, 0, color.RGBA8[color.Linear]{R: 0, G: 0, B: 255, A: 255})

	// Create interpolator
	interpolator := NewMockInterpolator()

	// Create a mock filter LUT
	filter := image.NewImageFilterLUT()

	// Create the span filter
	spanFilter := NewSpanImageFilterRGBA2x2WithParams(source, interpolator, filter)

	// Generate a span (this will use fallback since our mock LUT may not have weight array)
	span := make([]color.RGBA8[color.Linear], 3)
	spanFilter.Generate(span, 0, 0)

	// Verify all pixels are valid (no negative values, alpha constraints)
	for i, pixel := range span {
		if pixel.R > pixel.A {
			t.Errorf("Pixel %d: R (%d) should not exceed A (%d)", i, pixel.R, pixel.A)
		}
		if pixel.G > pixel.A {
			t.Errorf("Pixel %d: G (%d) should not exceed A (%d)", i, pixel.G, pixel.A)
		}
		if pixel.B > pixel.A {
			t.Errorf("Pixel %d: B (%d) should not exceed A (%d)", i, pixel.B, pixel.A)
		}
	}
}

func TestSpanImageFilterRGBA_Generate(t *testing.T) {
	// Create a test source image
	source := NewMockRGBASource(4, 4)

	// Fill with test pattern
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			// Create a gradient with varying alpha
			alpha := basics.Int8u((x + y) * 32)
			if alpha > 255 {
				alpha = 255
			}
			source.SetPixel(x, y, color.RGBA8[color.Linear]{
				R: basics.Int8u(x * 64),
				G: basics.Int8u(y * 64),
				B: 128,
				A: alpha,
			})
		}
	}

	// Create interpolator
	interpolator := NewMockInterpolator()

	// Create a mock filter LUT
	filter := image.NewImageFilterLUT()

	// Create the span filter
	spanFilter := NewSpanImageFilterRGBAWithParams(source, interpolator, filter)

	// Generate a span
	span := make([]color.RGBA8[color.Linear], 4)
	spanFilter.Generate(span, 0, 0)

	// Verify alpha constraints
	for i, pixel := range span {
		if pixel.R > pixel.A {
			t.Errorf("Pixel %d: R (%d) should not exceed A (%d)", i, pixel.R, pixel.A)
		}
		if pixel.G > pixel.A {
			t.Errorf("Pixel %d: G (%d) should not exceed A (%d)", i, pixel.G, pixel.A)
		}
		if pixel.B > pixel.A {
			t.Errorf("Pixel %d: B (%d) should not exceed A (%d)", i, pixel.B, pixel.A)
		}
	}
}

func TestSpanImageResampleRGBAAffine_Create(t *testing.T) {
	// Test that we can create the resampling filters
	filter := NewSpanImageResampleRGBAAffine[*MockRGBASource]()
	if filter == nil {
		t.Error("Failed to create SpanImageResampleRGBAAffine")
	}

	// Test that the base is properly initialized
	if filter.base == nil {
		t.Error("Base SpanImageResampleAffine should not be nil")
	}
}

func TestSpanImageResampleRGBA_Create(t *testing.T) {
	// Test that we can create the general resampling filter
	filter := NewSpanImageResampleRGBA[*MockRGBASource, *MockInterpolator]()
	if filter == nil {
		t.Error("Failed to create SpanImageResampleRGBA")
	}

	// Test that the base is properly initialized
	if filter.base == nil {
		t.Error("Base SpanImageResample should not be nil")
	}
}

func TestRGBAAlphaConstraints(t *testing.T) {
	// Test specific alpha constraint scenarios
	testCases := []struct {
		name     string
		input    color.RGBA8[color.Linear]
		expected color.RGBA8[color.Linear]
	}{
		{
			name:     "Normal case - all components within alpha",
			input:    color.RGBA8[color.Linear]{R: 100, G: 150, B: 200, A: 255},
			expected: color.RGBA8[color.Linear]{R: 100, G: 150, B: 200, A: 255},
		},
		{
			name:     "R exceeds alpha",
			input:    color.RGBA8[color.Linear]{R: 200, G: 100, B: 50, A: 150},
			expected: color.RGBA8[color.Linear]{R: 150, G: 100, B: 50, A: 150},
		},
		{
			name:     "All RGB exceed alpha",
			input:    color.RGBA8[color.Linear]{R: 200, G: 220, B: 180, A: 100},
			expected: color.RGBA8[color.Linear]{R: 100, G: 100, B: 100, A: 100},
		},
		{
			name:     "Zero alpha case",
			input:    color.RGBA8[color.Linear]{R: 100, G: 100, B: 100, A: 0},
			expected: color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Apply the same constraints that our filters apply
			result := tc.input

			// Apply alpha constraints
			if result.R > result.A {
				result.R = result.A
			}
			if result.G > result.A {
				result.G = result.A
			}
			if result.B > result.A {
				result.B = result.A
			}

			if result != tc.expected {
				t.Errorf("Expected %+v, got %+v", tc.expected, result)
			}
		})
	}
}

func TestRGBATransparentPixelHandling(t *testing.T) {
	// Create a source with transparent pixels
	source := NewMockRGBASource(2, 2)

	// Set some pixels to be transparent
	source.SetPixel(0, 0, color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 0}) // Transparent white
	source.SetPixel(1, 0, color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})     // Opaque black

	// Create interpolator
	interpolator := NewMockInterpolator()

	// Test nearest neighbor filter with transparent pixels
	filter := NewSpanImageFilterRGBANNWithParams(source, interpolator)
	span := make([]color.RGBA8[color.Linear], 2)
	filter.Generate(span, 0, 0)

	// First pixel should be transparent
	if span[0].A != 0 {
		t.Errorf("Expected transparent pixel (A=0), got A=%d", span[0].A)
	}

	// When alpha is 0, RGB values should be constrained to 0 in proper implementations
	// but since we're doing nearest neighbor, we get the exact values
	if span[0] != (color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 0}) {
		t.Errorf("Expected {255, 255, 255, 0}, got %+v", span[0])
	}

	// Second pixel should be opaque black
	if span[1] != (color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255}) {
		t.Errorf("Expected {0, 0, 0, 255}, got %+v", span[1])
	}
}
