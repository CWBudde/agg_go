package conv

import (
	"encoding/binary"
	"testing"

	"agg_go/internal/basics"
)

// TestRGB555AccuracyVsCpp tests RGB555 conversion accuracy against expected C++ AGG results.
// This ensures the Go implementation matches the C++ bit manipulation exactly.
func TestRGB555AccuracyVsCpp(t *testing.T) {
	conv := NewColorConvRGB555ToRGB24()

	// Test cases with known C++ AGG results
	testCases := []struct {
		name     string
		rgb555   uint16   // Input RGB555 value
		expected [3]uint8 // Expected RGB24 output
	}{
		{"Black", 0x0000, [3]uint8{0x00, 0x00, 0x00}},
		{"White", 0x7FFF, [3]uint8{0xF8, 0xF8, 0xF8}},
		{"Red", 0x7C00, [3]uint8{0xF8, 0x00, 0x00}},
		{"Green", 0x03E0, [3]uint8{0x00, 0xF8, 0x00}},
		{"Blue", 0x001F, [3]uint8{0x00, 0x00, 0xF8}},
		{"Mid Gray", 0x39CE, [3]uint8{0x70, 0x70, 0x70}},     // Approx mid-level
		{"Bright Red", 0x7800, [3]uint8{0xF0, 0x00, 0x00}},   // Near-max red
		{"Bright Green", 0x0380, [3]uint8{0x00, 0xE0, 0x00}}, // Mid-level green
		{"Bright Blue", 0x001C, [3]uint8{0x00, 0x00, 0xE0}},  // Near-max blue
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create source buffer with RGB555 data
			src := make([]basics.Int8u, 2)
			binary.LittleEndian.PutUint16(src, tc.rgb555)

			// Create destination buffer
			dst := make([]basics.Int8u, 3)

			// Perform conversion
			conv.CopyRow(dst, src, 1)

			// Verify exact match with C++ behavior
			if dst[0] != tc.expected[0] || dst[1] != tc.expected[1] || dst[2] != tc.expected[2] {
				t.Errorf("RGB555 0x%04X: expected [%02X,%02X,%02X], got [%02X,%02X,%02X]",
					tc.rgb555, tc.expected[0], tc.expected[1], tc.expected[2],
					dst[0], dst[1], dst[2])

				// Show bit analysis for debugging
				rgb := tc.rgb555
				expectedR := uint8((rgb >> 7) & 0xF8)
				expectedG := uint8((rgb >> 2) & 0xF8)
				expectedB := uint8((rgb << 3) & 0xF8)
				t.Logf("Bit analysis: R=0x%02X G=0x%02X B=0x%02X", expectedR, expectedG, expectedB)
			}
		})
	}
}

// TestRGB565AccuracyVsCpp tests RGB565 conversion accuracy against expected C++ AGG results.
func TestRGB565AccuracyVsCpp(t *testing.T) {
	conv := NewColorConvRGB565ToRGB24()

	testCases := []struct {
		name     string
		rgb565   uint16   // Input RGB565 value
		expected [3]uint8 // Expected RGB24 output
	}{
		{"Black", 0x0000, [3]uint8{0x00, 0x00, 0x00}},
		{"White", 0xFFFF, [3]uint8{0xF8, 0xFC, 0xF8}},
		{"Red", 0xF800, [3]uint8{0xF8, 0x00, 0x00}},
		{"Green", 0x07E0, [3]uint8{0x00, 0xFC, 0x00}},
		{"Blue", 0x001F, [3]uint8{0x00, 0x00, 0xF8}},
		{"Yellow", 0xFFE0, [3]uint8{0xF8, 0xFC, 0x00}},   // Red + Green
		{"Cyan", 0x07FF, [3]uint8{0x00, 0xFC, 0xF8}},     // Green + Blue
		{"Magenta", 0xF81F, [3]uint8{0xF8, 0x00, 0xF8}},  // Red + Blue
		{"Mid Gray", 0x7BEF, [3]uint8{0x78, 0x7C, 0x78}}, // Approx mid-level
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create source buffer with RGB565 data
			src := make([]basics.Int8u, 2)
			binary.LittleEndian.PutUint16(src, tc.rgb565)

			// Create destination buffer
			dst := make([]basics.Int8u, 3)

			// Perform conversion
			conv.CopyRow(dst, src, 1)

			// Verify exact match with C++ behavior
			if dst[0] != tc.expected[0] || dst[1] != tc.expected[1] || dst[2] != tc.expected[2] {
				t.Errorf("RGB565 0x%04X: expected [%02X,%02X,%02X], got [%02X,%02X,%02X]",
					tc.rgb565, tc.expected[0], tc.expected[1], tc.expected[2],
					dst[0], dst[1], dst[2])

				// Show bit analysis for debugging
				rgb := tc.rgb565
				expectedR := uint8((rgb >> 8) & 0xF8)
				expectedG := uint8((rgb >> 3) & 0xFC)
				expectedB := uint8((rgb << 3) & 0xF8)
				t.Logf("Bit analysis: R=0x%02X G=0x%02X B=0x%02X", expectedR, expectedG, expectedB)
			}
		})
	}
}

// TestColorConvRowConversions tests individual row conversions for accuracy.
func TestColorConvRowConversions(t *testing.T) {
	// Test RGB565 to RGB24 conversion at the row level
	converter := NewColorConvRGB565ToRGB24()

	// Test data: Red, Green, Blue, White in RGB565
	testPixels := []uint16{0xF800, 0x07E0, 0x001F, 0xFFFF}
	src := make([]basics.Int8u, 8) // 4 pixels * 2 bytes each

	for i, pixel := range testPixels {
		offset := i * 2
		binary.LittleEndian.PutUint16(src[offset:], pixel)
	}

	// Convert
	dst := make([]basics.Int8u, 12) // 4 pixels * 3 bytes each
	converter.CopyRow(dst, src, 4)

	// Verify results
	expectedResults := [][3]uint8{
		{0xF8, 0x00, 0x00}, // Red
		{0x00, 0xFC, 0x00}, // Green
		{0x00, 0x00, 0xF8}, // Blue
		{0xF8, 0xFC, 0xF8}, // White
	}

	for i, expected := range expectedResults {
		offset := i * 3
		actual := [3]uint8{dst[offset], dst[offset+1], dst[offset+2]}
		if actual != expected {
			t.Errorf("Pixel %d: expected %v, got %v", i, expected, actual)
		}
	}
}

// TestColorConvSameBehavior tests that identical formats use fast path correctly.
func TestColorConvSameBehavior(t *testing.T) {
	// Test that same-format conversion uses memmove path (ColorConvSame)
	conv := NewColorConvSame(3) // 3 bytes per pixel (RGB24)

	src := []basics.Int8u{0xFF, 0x80, 0x40, 0x20, 0x10, 0x08}
	dst := make([]basics.Int8u, 6)

	conv.CopyRow(dst, src, 2) // 2 pixels

	for i := 0; i < 6; i++ {
		if dst[i] != src[i] {
			t.Errorf("Index %d: expected 0x%02X, got 0x%02X", i, src[i], dst[i])
		}
	}
}
