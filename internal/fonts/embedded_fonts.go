// Package fonts provides embedded font data for AGG text rendering.
// This package contains simple bitmap fonts that can be used for testing
// and basic text rendering functionality.
package fonts

// Simple4x6Font is a basic 4x6 pixel bitmap font for testing.
// Font format: [height, baseline, start_char, num_chars, offset_table..., bitmap_data...]
var Simple4x6Font = []byte{
	6,  // height
	5,  // baseline
	65, // start char ('A')
	3,  // num chars (A, B, C)

	// Character offset table (2 bytes per character, little endian)
	// A (65)
	0x00, 0x00, // offset 0
	// B (66)
	0x07, 0x00, // offset 7
	// C (67)
	0x0E, 0x00, // offset 14

	// Bitmap data starts here
	// A (width=3, 6 rows, 1 byte per row since width <= 8)
	3, 0x60, 0xA0, 0xE0, 0xA0, 0xA0, 0x00,
	// B (width=3, 6 rows, 1 byte per row)
	3, 0xC0, 0xA0, 0xC0, 0xA0, 0xC0, 0x00,
	// C (width=3, 6 rows, 1 byte per row)
	3, 0x60, 0x80, 0x80, 0x80, 0x60, 0x00,
}

// GetSimple4x6Font returns the simple 4x6 bitmap font data
func GetSimple4x6Font() []byte {
	return Simple4x6Font
}
