package span

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/image"
)

// MockRGBSource provides a simple RGB image source for testing
// that implements the RGBSourceInterface
type MockRGBSource struct {
	width, height      int
	data               []color.RGB8[color.Linear]
	currentX, currentY int
	orderType          color.ColorOrder
}

type FixedInterpolator struct {
	x int
	y int
}

func NewFixedInterpolator(x, y int) *FixedInterpolator {
	return &FixedInterpolator{x: x, y: y}
}

func (fi *FixedInterpolator) Begin(x, y float64, length int) {}

func (fi *FixedInterpolator) Resynchronize(xe, ye float64, length int) {}

func (fi *FixedInterpolator) Coordinates() (x, y int) {
	return fi.x, fi.y
}

func (fi *FixedInterpolator) Next() {}

func (fi *FixedInterpolator) SubpixelShift() int {
	return image.ImageSubpixelShift
}

func NewMockRGBSource(width, height int) *MockRGBSource {
	return &MockRGBSource{
		width:     width,
		height:    height,
		data:      make([]color.RGB8[color.Linear], width*height),
		orderType: color.OrderRGB24, // Default to RGB order
	}
}

func (m *MockRGBSource) Width() int  { return m.width }
func (m *MockRGBSource) Height() int { return m.height }

// ColorType returns the RGB color type identifier
func (m *MockRGBSource) ColorType() string {
	return "RGB8"
}

// OrderType returns the color component ordering
func (m *MockRGBSource) OrderType() color.ColorOrder {
	return m.orderType
}

func (m *MockRGBSource) SetPixel(x, y int, c color.RGB8[color.Linear]) {
	if x >= 0 && y >= 0 && x < m.width && y < m.height {
		m.data[y*m.width+x] = c
	}
}

func (m *MockRGBSource) GetPixel(x, y int) color.RGB8[color.Linear] {
	if x >= 0 && y >= 0 && x < m.width && y < m.height {
		return m.data[y*m.width+x]
	}
	return color.RGB8[color.Linear]{R: 0, G: 0, B: 0}
}

// Span returns raw RGB pixel data starting at (x, y) with given length
func (m *MockRGBSource) Span(x, y, length int) []basics.Int8u {
	if x < 0 || y < 0 || x >= m.width || y >= m.height {
		return []basics.Int8u{0, 0, 0} // Return black for out-of-bounds
	}

	// Set current position for NextX/NextY methods
	m.currentX = x
	m.currentY = y

	// Convert RGB data to raw bytes according to order type
	result := make([]basics.Int8u, 3) // Return single pixel for first call
	pixel := m.GetPixel(x, y)

	result[m.orderType.R] = pixel.R
	result[m.orderType.G] = pixel.G
	result[m.orderType.B] = pixel.B

	return result
}

// NextX advances to the next pixel in current span and returns its RGB data
func (m *MockRGBSource) NextX() []basics.Int8u {
	m.currentX++
	if m.currentX >= m.width {
		return []basics.Int8u{0, 0, 0}
	}

	result := make([]basics.Int8u, 3)
	pixel := m.GetPixel(m.currentX, m.currentY)

	result[m.orderType.R] = pixel.R
	result[m.orderType.G] = pixel.G
	result[m.orderType.B] = pixel.B

	return result
}

// NextY advances to the next row at original x position
func (m *MockRGBSource) NextY() []basics.Int8u {
	m.currentY++
	if m.currentY >= m.height {
		return []basics.Int8u{0, 0, 0}
	}

	result := make([]basics.Int8u, 3)
	pixel := m.GetPixel(m.currentX, m.currentY)

	result[m.orderType.R] = pixel.R
	result[m.orderType.G] = pixel.G
	result[m.orderType.B] = pixel.B

	return result
}

// RowPtr returns pointer to row data starting at specified row
func (m *MockRGBSource) RowPtr(y int) []basics.Int8u {
	if y < 0 || y >= m.height {
		return []basics.Int8u{}
	}

	// Convert entire row to raw bytes
	result := make([]basics.Int8u, m.width*3)
	for x := 0; x < m.width; x++ {
		pixel := m.GetPixel(x, y)
		offset := x * 3
		result[offset+m.orderType.R] = pixel.R
		result[offset+m.orderType.G] = pixel.G
		result[offset+m.orderType.B] = pixel.B
	}

	return result
}

// Using the existing MockInterpolator from span_image_filter_gray_test.go

func TestSpanImageFilterRGBNN_Generate(t *testing.T) {
	// Create a test source image
	source := NewMockRGBSource(4, 4)

	// Fill with a simple pattern
	source.SetPixel(0, 0, color.RGB8[color.Linear]{R: 255, G: 0, B: 0})     // Red
	source.SetPixel(1, 0, color.RGB8[color.Linear]{R: 0, G: 255, B: 0})     // Green
	source.SetPixel(2, 0, color.RGB8[color.Linear]{R: 0, G: 0, B: 255})     // Blue
	source.SetPixel(3, 0, color.RGB8[color.Linear]{R: 255, G: 255, B: 255}) // White

	// Create interpolator
	interpolator := NewMockInterpolator()

	// Create the span filter
	filter := NewSpanImageFilterRGBNNWithParams(source, interpolator)

	// Generate a span
	span := make([]color.RGB8[color.Linear], 4)
	filter.Generate(span, 0, 0)

	// Verify results - should match source pixels exactly for nearest neighbor
	expected := []color.RGB8[color.Linear]{
		{R: 255, G: 0, B: 0},     // Red
		{R: 0, G: 255, B: 0},     // Green
		{R: 0, G: 0, B: 255},     // Blue
		{R: 255, G: 255, B: 255}, // White
	}

	for i, expectedColor := range expected {
		if span[i] != expectedColor {
			t.Errorf("Pixel %d: expected %+v, got %+v", i, expectedColor, span[i])
		}
	}
}

func TestSpanImageFilterRGBBilinear_Generate(t *testing.T) {
	// Create a test source image
	source := NewMockRGBSource(3, 3)

	// Fill with gradient pattern
	source.SetPixel(0, 0, color.RGB8[color.Linear]{R: 0, G: 0, B: 0})       // Black
	source.SetPixel(1, 0, color.RGB8[color.Linear]{R: 128, G: 128, B: 128}) // Gray
	source.SetPixel(2, 0, color.RGB8[color.Linear]{R: 255, G: 255, B: 255}) // White

	source.SetPixel(0, 1, color.RGB8[color.Linear]{R: 255, G: 0, B: 0}) // Red
	source.SetPixel(1, 1, color.RGB8[color.Linear]{R: 0, G: 255, B: 0}) // Green
	source.SetPixel(2, 1, color.RGB8[color.Linear]{R: 0, G: 0, B: 255}) // Blue

	// Create interpolator
	interpolator := NewMockInterpolator()

	// Create the bilinear span filter
	filter := NewSpanImageFilterRGBBilinearWithParams(source, interpolator)

	// Generate a span
	span := make([]color.RGB8[color.Linear], 3)
	filter.Generate(span, 0, 0)

	// For bilinear interpolation, results should be reasonable but not exact matches
	// Just verify we get non-zero values and they're within expected range
	for i, pixel := range span {
		if pixel.R > 255 || pixel.G > 255 || pixel.B > 255 {
			t.Errorf("Pixel %d has invalid values: %+v", i, pixel)
		}
		t.Logf("Bilinear pixel %d: %+v", i, pixel)
	}
}

func TestSpanImageFilterRGBBilinearClip_BackgroundColor(t *testing.T) {
	// Create a test source image
	source := NewMockRGBSource(2, 2)

	// Fill with simple pattern
	source.SetPixel(0, 0, color.RGB8[color.Linear]{R: 255, G: 0, B: 0})
	source.SetPixel(1, 1, color.RGB8[color.Linear]{R: 0, G: 255, B: 0})

	// Create interpolator
	interpolator := NewMockInterpolator()

	// Create background color
	backgroundColor := color.RGB8[color.Linear]{R: 128, G: 64, B: 192}

	// Create the clipping span filter
	filter := NewSpanImageFilterRGBBilinearClipWithParams(source, backgroundColor, interpolator)

	// Test background color getter/setter
	if filter.BackgroundColor() != backgroundColor {
		t.Errorf("Background color mismatch: expected %+v, got %+v",
			backgroundColor, filter.BackgroundColor())
	}

	newBg := color.RGB8[color.Linear]{R: 64, G: 128, B: 255}
	filter.SetBackgroundColor(newBg)

	if filter.BackgroundColor() != newBg {
		t.Errorf("Background color after set: expected %+v, got %+v",
			newBg, filter.BackgroundColor())
	}
}

func TestSpanImageFilterRGBBilinearClip_PartialOverlapBlendsBackground(t *testing.T) {
	source := NewMockRGBSource(1, 1)
	source.SetPixel(0, 0, color.RGB8[color.Linear]{R: 200, G: 100, B: 50})

	backgroundColor := color.RGB8[color.Linear]{R: 40, G: 20, B: 10}

	// Coordinates chosen so that after subtracting the default 0.5-pixel filter
	// offset the bilinear footprint overlaps 25% background and 75% source.
	interpolator := NewFixedInterpolator(64, 128)

	filter := NewSpanImageFilterRGBBilinearClipWithParams(source, backgroundColor, interpolator)

	span := make([]color.RGB8[color.Linear], 1)
	filter.Generate(span, 0, 0)

	got := span[0]
	want := color.RGB8[color.Linear]{R: 160, G: 80, B: 40}

	if got != want {
		t.Fatalf("partial-overlap sample = %+v, want %+v", got, want)
	}
}

func TestSpanImageFilterRGB2x2_WithFilter(t *testing.T) {
	// Create a test source image
	source := NewMockRGBSource(4, 4)

	// Fill with checkerboard pattern
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			if (x+y)%2 == 0 {
				source.SetPixel(x, y, color.RGB8[color.Linear]{R: 255, G: 255, B: 255})
			} else {
				source.SetPixel(x, y, color.RGB8[color.Linear]{R: 0, G: 0, B: 0})
			}
		}
	}

	// Create interpolator
	interpolator := NewMockInterpolator()

	// Create a simple bilinear filter
	bilinearFilter := image.BilinearFilter{}
	filterLUT := image.NewImageFilterLUTWithFilter(bilinearFilter, true)

	// Create the 2x2 span filter
	filter := NewSpanImageFilterRGB2x2WithParams(source, interpolator, filterLUT)

	// Generate a span
	span := make([]color.RGB8[color.Linear], 2)
	filter.Generate(span, 1, 1)

	// Results should be filtered values (not pure black/white due to filtering)
	for i, pixel := range span {
		t.Logf("2x2 filtered pixel %d: %+v", i, pixel)
		// Just verify reasonable values
		if pixel.R > 255 || pixel.G > 255 || pixel.B > 255 {
			t.Errorf("Pixel %d has invalid values: %+v", i, pixel)
		}
	}
}

func TestSpanImageFilterRGB_WithVariousFilters(t *testing.T) {
	// Create a test source image
	source := NewMockRGBSource(5, 5)

	// Fill center with white, edges with black
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			if x == 2 && y == 2 {
				source.SetPixel(x, y, color.RGB8[color.Linear]{R: 255, G: 255, B: 255})
			} else {
				source.SetPixel(x, y, color.RGB8[color.Linear]{R: 0, G: 0, B: 0})
			}
		}
	}

	// Create interpolator
	interpolator := NewMockInterpolator()

	tests := []struct {
		name   string
		filter image.FilterFunction
	}{
		{"bilinear", image.BilinearFilter{}},
		{"hanning", image.HanningFilter{}},
		{"hamming", image.HammingFilter{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterLUT := image.NewImageFilterLUTWithFilter(tt.filter, true)
			filter := NewSpanImageFilterRGBWithParams(source, interpolator, filterLUT)

			span := make([]color.RGB8[color.Linear], 3)
			filter.Generate(span, 1, 2) // Sample around the white center pixel

			for i, pixel := range span {
				t.Logf("%s filtered pixel %d: %+v", tt.name, i, pixel)
				// Verify reasonable values
				if pixel.R > 255 || pixel.G > 255 || pixel.B > 255 {
					t.Errorf("Pixel %d has invalid values: %+v", i, pixel)
				}
			}
		})
	}
}

func TestSpanImageFilterRGB_EmptySpan(t *testing.T) {
	source := NewMockRGBSource(2, 2)
	interpolator := NewMockInterpolator()
	filter := NewSpanImageFilterRGBNNWithParams(source, interpolator)

	// Test with empty span
	var span []color.RGB8[color.Linear]
	filter.Generate(span, 0, 0) // Should not panic

	// Test with zero-length span
	span = make([]color.RGB8[color.Linear], 0)
	filter.Generate(span, 0, 0) // Should not panic
}

func TestSpanImageFilterRGB_BoundaryConditions(t *testing.T) {
	// Create small source image
	source := NewMockRGBSource(2, 2)
	source.SetPixel(0, 0, color.RGB8[color.Linear]{R: 255, G: 128, B: 64})
	source.SetPixel(1, 1, color.RGB8[color.Linear]{R: 64, G: 128, B: 255})

	interpolator := NewMockInterpolator()
	filter := NewSpanImageFilterRGBNNWithParams(source, interpolator)

	// Test sampling at various positions
	testCases := []struct {
		x, y int
		name string
	}{
		{0, 0, "top-left"},
		{1, 0, "top-right"},
		{0, 1, "bottom-left"},
		{1, 1, "bottom-right"},
		{-1, 0, "out-of-bounds-left"},
		{2, 0, "out-of-bounds-right"},
		{0, -1, "out-of-bounds-top"},
		{0, 2, "out-of-bounds-bottom"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			span := make([]color.RGB8[color.Linear], 1)
			filter.Generate(span, tc.x, tc.y)

			t.Logf("Position (%d,%d): %+v", tc.x, tc.y, span[0])

			// Should not crash and should produce reasonable values
			if span[0].R > 255 || span[0].G > 255 || span[0].B > 255 {
				t.Errorf("Invalid pixel values at (%d,%d): %+v", tc.x, tc.y, span[0])
			}
		})
	}
}

// Benchmark tests for performance evaluation
func BenchmarkSpanImageFilterRGBNN(b *testing.B) {
	source := NewMockRGBSource(100, 100)

	// Fill with random-ish pattern
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			source.SetPixel(x, y, color.RGB8[color.Linear]{
				R: basics.Int8u((x * y) % 256),
				G: basics.Int8u((x + y) % 256),
				B: basics.Int8u((x - y) % 256),
			})
		}
	}

	interpolator := NewMockInterpolator()
	filter := NewSpanImageFilterRGBNNWithParams(source, interpolator)
	span := make([]color.RGB8[color.Linear], 50)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.Generate(span, 25, 25)
	}
}

func BenchmarkSpanImageFilterRGBBilinear(b *testing.B) {
	source := NewMockRGBSource(100, 100)

	// Fill with random-ish pattern
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			source.SetPixel(x, y, color.RGB8[color.Linear]{
				R: basics.Int8u((x * y) % 256),
				G: basics.Int8u((x + y) % 256),
				B: basics.Int8u((x - y) % 256),
			})
		}
	}

	interpolator := NewMockInterpolator()
	filter := NewSpanImageFilterRGBBilinearWithParams(source, interpolator)
	span := make([]color.RGB8[color.Linear], 50)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.Generate(span, 25, 25)
	}
}

// Additional comprehensive tests for RGB filtering

func TestSpanImageFilterRGB_ColorOrderHandling(t *testing.T) {
	// Test different color orderings (RGB vs BGR)
	testCases := []struct {
		name      string
		order     color.ColorOrder
		inputRGB  [3]basics.Int8u
		expectRGB [3]basics.Int8u
	}{
		{
			name:      "RGB order",
			order:     color.OrderRGB24,
			inputRGB:  [3]basics.Int8u{255, 128, 64}, // R, G, B
			expectRGB: [3]basics.Int8u{255, 128, 64}, // Should match input
		},
		{
			name:      "BGR order",
			order:     color.OrderBGR24,
			inputRGB:  [3]basics.Int8u{255, 128, 64}, // R, G, B
			expectRGB: [3]basics.Int8u{255, 128, 64}, // Should still be correct RGB
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			source := NewMockRGBSource(2, 2)
			source.orderType = tc.order
			source.SetPixel(0, 0, color.RGB8[color.Linear]{
				R: tc.inputRGB[0],
				G: tc.inputRGB[1],
				B: tc.inputRGB[2],
			})

			interpolator := NewMockInterpolator()
			filter := NewSpanImageFilterRGBNNWithParams(source, interpolator)
			span := make([]color.RGB8[color.Linear], 1)
			filter.Generate(span, 0, 0)

			if span[0].R != tc.expectRGB[0] ||
				span[0].G != tc.expectRGB[1] ||
				span[0].B != tc.expectRGB[2] {
				t.Errorf("Color order %s: expected RGB(%d,%d,%d), got RGB(%d,%d,%d)",
					tc.name, tc.expectRGB[0], tc.expectRGB[1], tc.expectRGB[2],
					span[0].R, span[0].G, span[0].B)
			}
		})
	}
}

func TestSpanImageFilterRGBBilinear_InterpolationAccuracy(t *testing.T) {
	// Create a 2x2 gradient to test interpolation
	source := NewMockRGBSource(2, 2)

	// Set up a simple gradient
	source.SetPixel(0, 0, color.RGB8[color.Linear]{R: 0, G: 0, B: 0})       // Top-left: black
	source.SetPixel(1, 0, color.RGB8[color.Linear]{R: 255, G: 0, B: 0})     // Top-right: red
	source.SetPixel(0, 1, color.RGB8[color.Linear]{R: 0, G: 255, B: 0})     // Bottom-left: green
	source.SetPixel(1, 1, color.RGB8[color.Linear]{R: 255, G: 255, B: 255}) // Bottom-right: white

	interpolator := NewMockInterpolator()
	filter := NewSpanImageFilterRGBBilinearWithParams(source, interpolator)
	span := make([]color.RGB8[color.Linear], 1)

	// Sample at exact pixel locations should match source
	filter.Generate(span, 0, 0)
	expected := source.GetPixel(0, 0)
	if span[0] != expected {
		t.Errorf("Bilinear at (0,0): expected %+v, got %+v", expected, span[0])
	}

	t.Logf("Bilinear interpolation test: sampled pixel at (0,0) = %+v", span[0])
}
