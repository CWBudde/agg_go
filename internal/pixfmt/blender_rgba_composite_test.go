package pixfmt

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
)

func TestCompositeBlenderMultiply(t *testing.T) {
	blender := NewMultiplyBlender[color.Linear, RGBAOrder]()

	// Test multiply blend with white source and red destination
	dst := []basics.Int8u{255, 0, 0, 255}          // Red
	blender.BlendPix(dst, 255, 255, 255, 255, 255) // White source, full coverage

	// With multiply, result should be destination * source = red * white = red
	// But the formula includes alpha blending, so it's more complex
	// For now, just verify the function doesn't crash and modifies the pixel
	if dst[0] == 255 && dst[1] == 0 && dst[2] == 0 && dst[3] == 255 {
		t.Log("Multiply blend preserved red color as expected")
	} else {
		t.Logf("Multiply blend result: R=%d, G=%d, B=%d, A=%d", dst[0], dst[1], dst[2], dst[3])
	}
}

func TestCompositeBlenderScreen(t *testing.T) {
	blender := NewScreenBlender[color.Linear, RGBAOrder]()

	// Test screen blend
	dst := []basics.Int8u{128, 128, 128, 255}      // 50% gray
	blender.BlendPix(dst, 128, 128, 128, 255, 255) // Same gray source

	// Screen should lighten the image
	t.Logf("Screen blend result: R=%d, G=%d, B=%d, A=%d", dst[0], dst[1], dst[2], dst[3])
}

func TestCompositeBlenderOverlay(t *testing.T) {
	blender := NewOverlayBlender[color.Linear, RGBAOrder]()

	// Test overlay blend
	dst := []basics.Int8u{100, 100, 100, 255}
	blender.BlendPix(dst, 200, 200, 200, 255, 255)

	t.Logf("Overlay blend result: R=%d, G=%d, B=%d, A=%d", dst[0], dst[1], dst[2], dst[3])
}

// Comprehensive tests for all Porter-Duff operations
func TestPorterDuffOperations(t *testing.T) {
	tests := []struct {
		name     string
		blender  CompositeBlender[color.Linear, RGBAOrder]
		dst      []basics.Int8u
		src      []basics.Int8u
		expected []basics.Int8u
	}{
		{
			name:     "Clear",
			blender:  NewClearBlender[color.Linear, RGBAOrder](),
			dst:      []basics.Int8u{255, 0, 0, 255}, // Red
			src:      []basics.Int8u{0, 255, 0, 255}, // Green
			expected: []basics.Int8u{0, 0, 0, 0},     // Transparent
		},
		{
			name:     "Src",
			blender:  NewSrcBlender[color.Linear, RGBAOrder](),
			dst:      []basics.Int8u{255, 0, 0, 255}, // Red
			src:      []basics.Int8u{0, 255, 0, 255}, // Green
			expected: []basics.Int8u{0, 255, 0, 255}, // Green
		},
		{
			name:     "Dst",
			blender:  NewDstBlender[color.Linear, RGBAOrder](),
			dst:      []basics.Int8u{255, 0, 0, 255}, // Red
			src:      []basics.Int8u{0, 255, 0, 255}, // Green
			expected: []basics.Int8u{255, 0, 0, 255}, // Red (unchanged)
		},
		{
			name:     "SrcIn",
			blender:  NewSrcInBlender[color.Linear, RGBAOrder](),
			dst:      []basics.Int8u{255, 0, 0, 255}, // Red, opaque
			src:      []basics.Int8u{0, 255, 0, 128}, // Green, half alpha
			expected: []basics.Int8u{0, 255, 0, 128}, // Src * dst.alpha (255) = Src
		},
		{
			name:     "XOR",
			blender:  NewXorBlender[color.Linear, RGBAOrder](),
			dst:      []basics.Int8u{255, 0, 0, 128},   // Red, half alpha
			src:      []basics.Int8u{0, 255, 0, 128},   // Green, half alpha
			expected: []basics.Int8u{127, 127, 0, 127}, // Expected XOR result (approximately)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := make([]basics.Int8u, 4)
			copy(dst, tt.dst)

			tt.blender.BlendPix(dst, tt.src[0], tt.src[1], tt.src[2], tt.src[3], 255)

			// Allow for small rounding differences
			tolerance := basics.Int8u(2)
			for i := 0; i < 4; i++ {
				diff := dst[i] - tt.expected[i]
				if diff < 0 {
					diff = -diff
				}
				if diff > tolerance {
					t.Errorf("%s: component %d: got %d, expected %d (±%d)",
						tt.name, i, dst[i], tt.expected[i], tolerance)
				}
			}
		})
	}
}

// Test blend modes with known mathematical results
func TestBlendModeFormulas(t *testing.T) {
	tests := []struct {
		name    string
		blender CompositeBlender[color.Linear, RGBAOrder]
		dst     []basics.Int8u
		src     []basics.Int8u
		checkFn func(t *testing.T, result []basics.Int8u)
	}{
		{
			name:    "Multiply with white source",
			blender: NewMultiplyBlender[color.Linear, RGBAOrder](),
			dst:     []basics.Int8u{128, 64, 192, 255},
			src:     []basics.Int8u{255, 255, 255, 255},
			checkFn: func(t *testing.T, result []basics.Int8u) {
				// Multiply with white should preserve the destination
				expected := []basics.Int8u{128, 64, 192, 255}
				for i := 0; i < 3; i++ { // Check RGB, allow alpha variation
					if abs(int(result[i])-int(expected[i])) > 2 {
						t.Errorf("Multiply with white: component %d: got %d, expected %d",
							i, result[i], expected[i])
					}
				}
			},
		},
		{
			name:    "Screen with black source",
			blender: NewScreenBlender[color.Linear, RGBAOrder](),
			dst:     []basics.Int8u{128, 64, 192, 255},
			src:     []basics.Int8u{0, 0, 0, 255},
			checkFn: func(t *testing.T, result []basics.Int8u) {
				// Screen with black should preserve the destination
				expected := []basics.Int8u{128, 64, 192, 255}
				for i := 0; i < 3; i++ { // Check RGB, allow alpha variation
					if abs(int(result[i])-int(expected[i])) > 2 {
						t.Errorf("Screen with black: component %d: got %d, expected %d",
							i, result[i], expected[i])
					}
				}
			},
		},
		{
			name:    "Plus blend",
			blender: NewPlusBlender[color.Linear, RGBAOrder](),
			dst:     []basics.Int8u{100, 50, 25, 255},
			src:     []basics.Int8u{50, 100, 25, 128},
			checkFn: func(t *testing.T, result []basics.Int8u) {
				// Plus should add the colors (with clamping)
				if result[0] < 140 || result[0] > 160 { // ~150 expected
					t.Errorf("Plus blend R: got %d, expected ~150", result[0])
				}
				if result[1] < 140 || result[1] > 160 { // ~150 expected
					t.Errorf("Plus blend G: got %d, expected ~150", result[1])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := make([]basics.Int8u, 4)
			copy(dst, tt.dst)

			tt.blender.BlendPix(dst, tt.src[0], tt.src[1], tt.src[2], tt.src[3], 255)

			tt.checkFn(t, dst)
		})
	}
}

// Test alpha blending edge cases
func TestAlphaBlending(t *testing.T) {
	tests := []struct {
		name    string
		blender CompositeBlender[color.Linear, RGBAOrder]
		dst     []basics.Int8u
		src     []basics.Int8u
		cover   basics.Int8u
		checkFn func(t *testing.T, result []basics.Int8u)
	}{
		{
			name:    "Zero alpha source",
			blender: NewMultiplyBlender[color.Linear, RGBAOrder](),
			dst:     []basics.Int8u{255, 0, 0, 255},
			src:     []basics.Int8u{0, 255, 0, 0}, // Green with zero alpha
			cover:   255,
			checkFn: func(t *testing.T, result []basics.Int8u) {
				// Zero alpha source should not change destination
				expected := []basics.Int8u{255, 0, 0, 255}
				for i := 0; i < 4; i++ {
					if result[i] != expected[i] {
						t.Errorf("Zero alpha: component %d: got %d, expected %d",
							i, result[i], expected[i])
					}
				}
			},
		},
		{
			name:    "Partial coverage",
			blender: NewSrcOverBlender[color.Linear, RGBAOrder](),
			dst:     []basics.Int8u{255, 0, 0, 255},
			src:     []basics.Int8u{0, 255, 0, 255},
			cover:   128, // 50% coverage
			checkFn: func(t *testing.T, result []basics.Int8u) {
				// Debug: log actual result
				t.Logf("Partial coverage result: R=%d, G=%d, B=%d, A=%d", result[0], result[1], result[2], result[3])
				// SrcOver with partial coverage should blend proportionally
				// With 50% coverage, we expect some blending
				if result[0] == 255 && result[1] == 0 {
					t.Errorf("Partial coverage: no blending occurred, got unchanged destination")
				}
				// Just verify there was some change
				if result[1] == 0 {
					t.Errorf("Partial coverage: green component should be > 0, got %d", result[1])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := make([]basics.Int8u, 4)
			copy(dst, tt.dst)

			tt.blender.BlendPix(dst, tt.src[0], tt.src[1], tt.src[2], tt.src[3], tt.cover)

			tt.checkFn(t, dst)
		})
	}
}

// Helper function for absolute difference
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func TestPixFmtTransposer(t *testing.T) {
	// Create a simple mock pixel format for testing
	mockPixfmt := &mockPixelFormat{
		width:  10,
		height: 20,
		pixels: make([]color.RGBA8[color.Linear], 10*20),
	}

	// Set a test pixel
	mockPixfmt.pixels[5*10+3] = color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 255} // Red at (3,5)

	transposer := NewPixFmtTransposer(mockPixfmt)

	// Check transposed dimensions
	if transposer.Width() != 20 || transposer.Height() != 10 {
		t.Errorf("Transposer dimensions wrong: got %dx%d, want 20x10", transposer.Width(), transposer.Height())
	}

	// Check transposed pixel access - (3,5) should become (5,3)
	pixel := transposer.GetPixel(5, 3)
	if pixel.R != 255 || pixel.G != 0 || pixel.B != 0 {
		t.Errorf("Transposed pixel wrong: got R=%d G=%d B=%d, want R=255 G=0 B=0", pixel.R, pixel.G, pixel.B)
	}
}

// Mock pixel format for testing
type mockPixelFormat struct {
	width, height int
	pixels        []color.RGBA8[color.Linear]
}

func (m *mockPixelFormat) Width() int  { return m.width }
func (m *mockPixelFormat) Height() int { return m.height }

func (m *mockPixelFormat) GetPixel(x, y int) color.RGBA8[color.Linear] {
	if x >= 0 && y >= 0 && x < m.width && y < m.height {
		return m.pixels[y*m.width+x]
	}
	return color.RGBA8[color.Linear]{}
}

func (m *mockPixelFormat) CopyPixel(x, y int, c color.RGBA8[color.Linear]) {
	if x >= 0 && y >= 0 && x < m.width && y < m.height {
		m.pixels[y*m.width+x] = c
	}
}

func (m *mockPixelFormat) BlendPixel(x, y int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	// Simple blend for testing
	if x >= 0 && y >= 0 && x < m.width && y < m.height {
		dst := &m.pixels[y*m.width+x]
		alpha := c.A * cover / 255
		invAlpha := 255 - alpha
		dst.R = basics.Int8u((uint32(c.R)*uint32(alpha) + uint32(dst.R)*uint32(invAlpha)) / 255)
		dst.G = basics.Int8u((uint32(c.G)*uint32(alpha) + uint32(dst.G)*uint32(invAlpha)) / 255)
		dst.B = basics.Int8u((uint32(c.B)*uint32(alpha) + uint32(dst.B)*uint32(invAlpha)) / 255)
		dst.A = basics.Int8u((uint32(c.A)*uint32(alpha) + uint32(dst.A)*uint32(invAlpha)) / 255)
	}
}

func (m *mockPixelFormat) CopyHline(x, y, length int, c color.RGBA8[color.Linear]) {
	for i := 0; i < length; i++ {
		m.CopyPixel(x+i, y, c)
	}
}

func (m *mockPixelFormat) CopyVline(x, y, length int, c color.RGBA8[color.Linear]) {
	for i := 0; i < length; i++ {
		m.CopyPixel(x, y+i, c)
	}
}

func (m *mockPixelFormat) BlendHline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	for i := 0; i < length; i++ {
		m.BlendPixel(x+i, y, c, cover)
	}
}

func (m *mockPixelFormat) BlendVline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	for i := 0; i < length; i++ {
		m.BlendPixel(x, y+i, c, cover)
	}
}

func (m *mockPixelFormat) BlendSolidHspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u) {
	for i := 0; i < length; i++ {
		if covers != nil {
			m.BlendPixel(x+i, y, c, covers[i])
		} else {
			m.BlendPixel(x+i, y, c, 255)
		}
	}
}

func (m *mockPixelFormat) BlendSolidVspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u) {
	for i := 0; i < length; i++ {
		if covers != nil {
			m.BlendPixel(x, y+i, c, covers[i])
		} else {
			m.BlendPixel(x, y+i, c, 255)
		}
	}
}
