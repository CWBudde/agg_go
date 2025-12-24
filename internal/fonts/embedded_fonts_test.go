package fonts

import (
	"bytes"
	"testing"
)

// Test font header validation
func TestGSE4x6FontHeader(t *testing.T) {
	font := GetGSE4x6()

	if len(font) < 4 {
		t.Fatalf("Font data too short: %d bytes", len(font))
	}

	// Test header values
	height := font[0]
	baseline := font[1]
	startChar := font[2]
	numChars := font[3]

	if height != 6 {
		t.Errorf("Expected height 6, got %d", height)
	}

	if baseline != 0 {
		t.Errorf("Expected baseline 0, got %d", baseline)
	}

	if startChar != 32 {
		t.Errorf("Expected start char 32 (space), got %d", startChar)
	}

	if numChars != 96 {
		t.Errorf("Expected 96 characters, got %d", numChars)
	}
}

// Test character offset table integrity
func TestGSE4x6OffsetTable(t *testing.T) {
	font := GetGSE4x6()

	numChars := int(font[3])
	offsetTableStart := 4
	offsetTableEnd := offsetTableStart + numChars*2

	if len(font) < offsetTableEnd {
		t.Fatalf("Font data too short for offset table: %d bytes", len(font))
	}

	// Test that offsets are increasing (monotonic)
	prevOffset := uint16(0)
	for i := 0; i < numChars; i++ {
		offsetPos := offsetTableStart + i*2
		// Little-endian 16-bit offset
		offset := uint16(font[offsetPos]) | (uint16(font[offsetPos+1]) << 8)

		if i > 0 && offset < prevOffset {
			t.Errorf("Offset table not monotonic at char %d: offset %d < previous %d",
				i, offset, prevOffset)
		}
		prevOffset = offset
	}
}

// Test bitmap data extraction for specific characters
func TestGSE4x6CharacterBitmaps(t *testing.T) {
	font := GetGSE4x6()

	// Test space character (ASCII 32, index 0)
	spaceOffset := getCharacterOffset(font, 32)
	if font[spaceOffset] != 4 {
		t.Errorf("Space character should be 4 pixels wide, got %d", font[spaceOffset])
	}

	// Test 'A' character (ASCII 65, index 33)
	aOffset := getCharacterOffset(font, 65)
	aWidth := font[aOffset]
	if aWidth != 4 {
		t.Errorf("'A' character should be 4 pixels wide, got %d", aWidth)
	}

	// Test that 'A' has the expected bitmap pattern
	expectedA := []byte{0x40, 0xa0, 0xe0, 0xa0, 0xa0, 0x00}
	actualA := font[aOffset+1 : aOffset+1+6] // width byte + 6 rows

	if !bytes.Equal(expectedA, actualA) {
		t.Errorf("'A' bitmap mismatch. Expected %v, got %v", expectedA, actualA)
	}
}

// Test font data immutability
func TestGSE4x6Immutability(t *testing.T) {
	font1 := GetGSE4x6()
	font2 := GetGSE4x6()

	// Verify they're different slices but same data
	if &font1[0] == &font2[0] {
		t.Error("Font data should return copies, not the same slice")
	}

	if !bytes.Equal(font1, font2) {
		t.Error("Multiple calls should return identical data")
	}

	// Modify one and ensure the other is unchanged
	originalByte := font1[0]
	font1[0] = 255

	if font2[0] != originalByte {
		t.Error("Modifying one font slice affected another")
	}
}

// Test character range coverage
func TestGSE4x6CharacterRange(t *testing.T) {
	font := GetGSE4x6()

	// Test that we can extract bitmaps for all characters in range
	for char := 32; char < 128; char++ {
		offset := getCharacterOffset(font, char)
		if offset >= len(font) {
			t.Errorf("Character %d (0x%02x) offset %d exceeds font data length %d",
				char, char, offset, len(font))
		}

		width := font[offset]
		if width == 0 || width > 16 {
			t.Errorf("Character %d (0x%02x) has invalid width %d", char, char, width)
		}
	}
}

// Test deprecated function backward compatibility
func TestSimple4x6FontBackwardCompatibility(t *testing.T) {
	oldFont := GetSimple4x6Font()

	if len(oldFont) < 4 {
		t.Fatalf("Deprecated font data too short: %d bytes", len(oldFont))
	}

	// Test basic header
	if oldFont[0] != 6 {
		t.Errorf("Deprecated font height should be 6, got %d", oldFont[0])
	}

	if oldFont[2] != 65 {
		t.Errorf("Deprecated font should start at 'A' (65), got %d", oldFont[2])
	}

	if oldFont[3] != 3 {
		t.Errorf("Deprecated font should have 3 chars, got %d", oldFont[3])
	}
}

// Benchmark font data access
func BenchmarkGetGSE4x6(b *testing.B) {
	for i := 0; i < b.N; i++ {
		font := GetGSE4x6()
		_ = font[0] // Access first byte to prevent optimization
	}
}

// Benchmark character lookup
func BenchmarkCharacterLookup(b *testing.B) {
	font := GetGSE4x6()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		char := 65 + (i % 26) // Cycle through A-Z
		offset := getCharacterOffset(font, char)
		_ = font[offset] // Access width byte
	}
}

// Helper function to get character offset from font data
func getCharacterOffset(font []byte, char int) int {
	startChar := int(font[2])
	numChars := int(font[3])

	if char < startChar || char >= startChar+numChars {
		return -1 // Character not in font
	}

	charIndex := char - startChar
	offsetTableStart := 4
	offsetPos := offsetTableStart + charIndex*2

	// Read little-endian 16-bit offset
	offset := int(font[offsetPos]) | (int(font[offsetPos+1]) << 8)
	return offsetTableStart + numChars*2 + offset
}

// Test helper function itself
func TestGetCharacterOffset(t *testing.T) {
	font := GetGSE4x6()

	// Test space character (first in range)
	spaceOffset := getCharacterOffset(font, 32)
	expectedOffset := 4 + 96*2 + 0 // header + offset table + first char offset
	if spaceOffset != expectedOffset {
		t.Errorf("Space offset should be %d, got %d", expectedOffset, spaceOffset)
	}

	// Test character outside range
	invalidOffset := getCharacterOffset(font, 31) // Below range
	if invalidOffset != -1 {
		t.Errorf("Character outside range should return -1, got %d", invalidOffset)
	}

	invalidOffset = getCharacterOffset(font, 128) // Above range
	if invalidOffset != -1 {
		t.Errorf("Character outside range should return -1, got %d", invalidOffset)
	}
}

// Test font width calculation for strings
func TestStringWidth(t *testing.T) {
	font := GetGSE4x6()

	// Calculate width of "ABC"
	totalWidth := 0
	for _, char := range "ABC" {
		offset := getCharacterOffset(font, int(char))
		if offset >= 0 {
			totalWidth += int(font[offset])
		}
	}

	// Each character in GSE4x6 is 4 pixels wide
	expectedWidth := 4 * 3
	if totalWidth != expectedWidth {
		t.Errorf("Width of 'ABC' should be %d, got %d", expectedWidth, totalWidth)
	}
}

// Test font data integrity against corruption
func TestFontDataIntegrity(t *testing.T) {
	font := GetGSE4x6()

	// Test that font has reasonable size (should be around 670+ bytes)
	if len(font) < 600 {
		t.Errorf("Font data seems too small: %d bytes", len(font))
	}

	if len(font) > 1000 {
		t.Errorf("Font data seems too large: %d bytes", len(font))
	}

	// Test that all character widths are reasonable
	numChars := int(font[3])
	dataStart := 4 + numChars*2

	pos := dataStart
	for i := 0; i < numChars && pos < len(font); i++ {
		width := int(font[pos])
		if width > 16 || width == 0 {
			t.Errorf("Character %d has unreasonable width %d", i, width)
		}

		// Skip to next character (width byte + height rows)
		pos += 1 + 6 // GSE4x6 has height of 6
	}
}

// Test GSE5x7 font header validation
func TestGSE5x7FontHeader(t *testing.T) {
	font := GetGSE5x7()

	if len(font) < 4 {
		t.Fatalf("GSE5x7 font data too short: %d bytes", len(font))
	}

	// Test header values
	height := font[0]
	baseline := font[1]
	startChar := font[2]
	numChars := font[3]

	if height != 7 {
		t.Errorf("Expected height 7, got %d", height)
	}

	if baseline != 0 {
		t.Errorf("Expected baseline 0, got %d", baseline)
	}

	if startChar != 32 {
		t.Errorf("Expected start char 32 (space), got %d", startChar)
	}

	if numChars != 96 {
		t.Errorf("Expected 96 characters, got %d", numChars)
	}
}

// Test GSE5x7 character bitmaps
func TestGSE5x7CharacterBitmaps(t *testing.T) {
	font := GetGSE5x7()

	// Test 'A' character (ASCII 65, index 33)
	aOffset := getCharacterOffset(font, 65)
	aWidth := font[aOffset]
	if aWidth != 5 {
		t.Errorf("'A' character should be 5 pixels wide, got %d", aWidth)
	}

	// Test that 'A' has the expected bitmap pattern (5x7)
	expectedA := []byte{0x00, 0x20, 0x50, 0x88, 0xf8, 0x88, 0x00}
	actualA := font[aOffset+1 : aOffset+1+7] // width byte + 7 rows

	if !bytes.Equal(expectedA, actualA) {
		t.Errorf("'A' bitmap mismatch. Expected %v, got %v", expectedA, actualA)
	}
}

// Test all font accessor functions
func TestAllFontAccessors(t *testing.T) {
	fonts := map[string]func() []byte{
		"GSE4x6":             GetGSE4x6,
		"GSE5x7":             GetGSE5x7,
		"GSE4x8":             GetGSE4x8,
		"GSE5x9":             GetGSE5x9,
		"GSE6x9":             GetGSE6x9,
		"GSE6x12":            GetGSE6x12,
		"GSE7x11":            GetGSE7x11,
		"GSE7x11Bold":        GetGSE7x11Bold,
		"GSE7x15":            GetGSE7x15,
		"GSE7x15Bold":        GetGSE7x15Bold,
		"GSE8x16":            GetGSE8x16,
		"GSE8x16Bold":        GetGSE8x16Bold,
		"MCS5x10Mono":        GetMCS5x10Mono,
		"MCS5x11Mono":        GetMCS5x11Mono,
		"MCS6x10Mono":        GetMCS6x10Mono,
		"MCS6x11Mono":        GetMCS6x11Mono,
		"MCS7x12MonoHigh":    GetMCS7x12MonoHigh,
		"MCS7x12MonoLow":     GetMCS7x12MonoLow,
		"MCS11Prop":          GetMCS11Prop,
		"MCS11PropCondensed": GetMCS11PropCondensed,
		"MCS12Prop":          GetMCS12Prop,
		"MCS13Prop":          GetMCS13Prop,
		"Verdana12":          GetVerdana12,
		"Verdana12Bold":      GetVerdana12Bold,
		"Verdana13":          GetVerdana13,
		"Verdana13Bold":      GetVerdana13Bold,
		"Verdana14":          GetVerdana14,
		"Verdana14Bold":      GetVerdana14Bold,
		"Verdana16":          GetVerdana16,
		"Verdana16Bold":      GetVerdana16Bold,
		"Verdana17":          GetVerdana17,
		"Verdana17Bold":      GetVerdana17Bold,
		"Verdana18":          GetVerdana18,
		"Verdana18Bold":      GetVerdana18Bold,
	}

	for name, getter := range fonts {
		t.Run(name, func(t *testing.T) {
			font := getter()
			if len(font) < 4 {
				t.Errorf("%s font data too short: %d bytes", name, len(font))
			}

			// Basic validation
			height := font[0]
			startChar := font[2]
			numChars := font[3]

			if height == 0 || height > 32 {
				t.Errorf("%s has invalid height: %d", name, height)
			}

			if startChar != 32 {
				t.Errorf("%s should start at space (32), got %d", name, startChar)
			}

			// MCS proportional fonts have 128 chars, others have 96
			expectedChars := byte(96)
			if len(name) >= 3 && name[:3] == "MCS" {
				// Check if it's a proportional font (ends with "Prop" or "PropCondensed")
				if len(name) >= 4 && name[len(name)-4:] == "Prop" {
					expectedChars = 128
				} else if len(name) >= 13 && name[len(name)-13:] == "PropCondensed" {
					expectedChars = 128
				}
			}

			if numChars != expectedChars {
				t.Errorf("%s should have %d chars, got %d", name, expectedChars, numChars)
			}
		})
	}
}

// Test MCS5x10Mono font header validation
func TestMCS5x10MonoFontHeader(t *testing.T) {
	font := GetMCS5x10Mono()

	if len(font) < 4 {
		t.Fatalf("MCS5x10Mono font data too short: %d bytes", len(font))
	}

	// Test header values
	height := font[0]
	baseline := font[1]
	startChar := font[2]
	numChars := font[3]

	if height != 10 {
		t.Errorf("Expected height 10, got %d", height)
	}

	if baseline != 2 {
		t.Errorf("Expected baseline 2, got %d", baseline)
	}

	if startChar != 32 {
		t.Errorf("Expected start char 32 (space), got %d", startChar)
	}

	if numChars != 96 {
		t.Errorf("Expected 96 characters, got %d", numChars)
	}
}

// Test MCS5x10Mono character bitmaps
func TestMCS5x10MonoCharacterBitmaps(t *testing.T) {
	font := GetMCS5x10Mono()

	// Test 'A' character (ASCII 65, index 33)
	aOffset := getCharacterOffset(font, 65)
	aWidth := font[aOffset]
	if aWidth != 5 {
		t.Errorf("'A' character should be 5 pixels wide, got %d", aWidth)
	}

	// Test that 'A' has the expected bitmap pattern (5x10)
	expectedA := []byte{0x00, 0x70, 0x88, 0x88, 0x88, 0x88, 0xF8, 0x88, 0x88, 0x00}
	actualA := font[aOffset+1 : aOffset+1+10] // width byte + 10 rows

	if !bytes.Equal(expectedA, actualA) {
		t.Errorf("'A' bitmap mismatch. Expected %v, got %v", expectedA, actualA)
	}
}

// Test MCS font dimensions
func TestMCSFontDimensions(t *testing.T) {
	testCases := []struct {
		name             string
		getter           func() []byte
		expectedHeight   int
		expectedBaseline int
	}{
		{"MCS5x10Mono", GetMCS5x10Mono, 10, 2},
		{"MCS5x11Mono", GetMCS5x11Mono, 11, 3},         // actual font data
		{"MCS6x10Mono", GetMCS6x10Mono, 10, 3},         // actual font data
		{"MCS7x12MonoHigh", GetMCS7x12MonoHigh, 12, 3}, // newly implemented
		{"MCS7x12MonoLow", GetMCS7x12MonoLow, 12, 4},   // newly implemented
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			font := tc.getter()
			if len(font) < 4 {
				t.Fatalf("%s font data too short: %d bytes", tc.name, len(font))
			}

			height := int(font[0])
			baseline := int(font[1])

			if height != tc.expectedHeight {
				t.Errorf("%s: expected height %d, got %d", tc.name, tc.expectedHeight, height)
			}

			if baseline != tc.expectedBaseline {
				t.Errorf("%s: expected baseline %d, got %d", tc.name, tc.expectedBaseline, baseline)
			}
		})
	}
}

// Test MCS font immutability
func TestMCSFontImmutability(t *testing.T) {
	font1 := GetMCS5x10Mono()
	font2 := GetMCS5x10Mono()

	// Verify they're different slices but same data
	if &font1[0] == &font2[0] {
		t.Error("MCS font data should return copies, not the same slice")
	}

	if !bytes.Equal(font1, font2) {
		t.Error("Multiple MCS font calls should return identical data")
	}

	// Modify one and ensure the other is unchanged
	originalByte := font1[0]
	font1[0] = 255

	if font2[0] != originalByte {
		t.Error("Modifying one MCS font slice affected another")
	}
}

// Test Verdana12 font header validation
func TestVerdana12FontHeader(t *testing.T) {
	font := GetVerdana12()

	if len(font) < 4 {
		t.Fatalf("Verdana12 font data too short: %d bytes", len(font))
	}

	// Test header values
	height := font[0]
	baseline := font[1]
	startChar := font[2]
	numChars := font[3]

	if height != 12 {
		t.Errorf("Expected height 12, got %d", height)
	}

	if baseline != 3 {
		t.Errorf("Expected baseline 3, got %d", baseline)
	}

	if startChar != 32 {
		t.Errorf("Expected start char 32 (space), got %d", startChar)
	}

	if numChars != 96 {
		t.Errorf("Expected 96 characters, got %d", numChars)
	}
}

// Test Verdana12 character bitmaps
func TestVerdana12CharacterBitmaps(t *testing.T) {
	font := GetVerdana12()

	// Test 'A' character (ASCII 65, index 33) - matches C++ AGG verdana12
	aOffset := getCharacterOffset(font, 65)
	if aOffset < 0 || aOffset >= len(font) {
		t.Fatalf("Invalid offset for 'A': %d (font length: %d)", aOffset, len(font))
	}

	aWidth := font[aOffset]
	if aWidth != 8 {
		t.Errorf("'A' character should be 8 pixels wide, got %d", aWidth)
	}

	// Test that 'A' has the expected bitmap pattern (8x12) - matches C++ AGG verdana12
	expectedA := []byte{0x00, 0x00, 0x00, 0x18, 0x18, 0x24, 0x24, 0x7E, 0x42, 0x42, 0x00, 0x00}
	if aOffset+1+12 > len(font) {
		t.Fatalf("Not enough data for 'A' bitmap at offset %d", aOffset)
	}
	actualA := font[aOffset+1 : aOffset+1+12] // width byte + 12 rows

	if !bytes.Equal(expectedA, actualA) {
		t.Errorf("'A' bitmap mismatch. Expected %v, got %v", expectedA, actualA)
	}
}

// Test Verdana font dimensions
func TestVerdanaFontDimensions(t *testing.T) {
	testCases := []struct {
		name             string
		getter           func() []byte
		expectedHeight   int
		expectedBaseline int
	}{
		{"Verdana12", GetVerdana12, 12, 3},
		// Note: Other Verdana fonts are placeholders returning Verdana12 data for now
		{"Verdana13", GetVerdana13, 13, 3}, // actual font data
		{"Verdana14", GetVerdana14, 14, 3}, // actual font data
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			font := tc.getter()
			if len(font) < 4 {
				t.Fatalf("%s font data too short: %d bytes", tc.name, len(font))
			}

			height := int(font[0])
			baseline := int(font[1])

			if height != tc.expectedHeight {
				t.Errorf("%s: expected height %d, got %d", tc.name, tc.expectedHeight, height)
			}

			if baseline != tc.expectedBaseline {
				t.Errorf("%s: expected baseline %d, got %d", tc.name, tc.expectedBaseline, baseline)
			}
		})
	}
}

// Test Verdana font immutability
func TestVerdanaFontImmutability(t *testing.T) {
	font1 := GetVerdana12()
	font2 := GetVerdana12()

	// Verify they're different slices but same data
	if &font1[0] == &font2[0] {
		t.Error("Verdana font data should return copies, not the same slice")
	}

	if !bytes.Equal(font1, font2) {
		t.Error("Multiple Verdana font calls should return identical data")
	}

	// Modify one and ensure the other is unchanged
	originalByte := font1[0]
	font1[0] = 255

	if font2[0] != originalByte {
		t.Error("Modifying one Verdana font slice affected another")
	}
}

// Test newly implemented MCS7x12 fonts
func TestMCS7x12Fonts(t *testing.T) {
	// Test MCS7x12MonoHigh
	fontHigh := GetMCS7x12MonoHigh()
	if len(fontHigh) < 4 {
		t.Fatalf("MCS7x12MonoHigh font data too short: %d bytes", len(fontHigh))
	}
	if fontHigh[0] != 12 { // height
		t.Errorf("MCS7x12MonoHigh: expected height 12, got %d", fontHigh[0])
	}
	if fontHigh[1] != 3 { // baseline
		t.Errorf("MCS7x12MonoHigh: expected baseline 3, got %d", fontHigh[1])
	}
	if fontHigh[2] != 32 { // start char
		t.Errorf("MCS7x12MonoHigh: expected start char 32, got %d", fontHigh[2])
	}
	if fontHigh[3] != 96 { // num chars
		t.Errorf("MCS7x12MonoHigh: expected 96 chars, got %d", fontHigh[3])
	}

	// Test MCS7x12MonoLow
	fontLow := GetMCS7x12MonoLow()
	if len(fontLow) < 4 {
		t.Fatalf("MCS7x12MonoLow font data too short: %d bytes", len(fontLow))
	}
	if fontLow[0] != 12 { // height
		t.Errorf("MCS7x12MonoLow: expected height 12, got %d", fontLow[0])
	}
	if fontLow[1] != 4 { // baseline
		t.Errorf("MCS7x12MonoLow: expected baseline 4, got %d", fontLow[1])
	}
	if fontLow[2] != 32 { // start char
		t.Errorf("MCS7x12MonoLow: expected start char 32, got %d", fontLow[2])
	}
	if fontLow[3] != 96 { // num chars
		t.Errorf("MCS7x12MonoLow: expected 96 chars, got %d", fontLow[3])
	}

	// Test that they have different content (different baseline)
	if bytes.Equal(fontHigh, fontLow) {
		t.Error("MCS7x12MonoHigh and MCS7x12MonoLow should have different data")
	}

	// Test immutability
	font1 := GetMCS7x12MonoHigh()
	font2 := GetMCS7x12MonoHigh()
	if &font1[0] == &font2[0] {
		t.Error("MCS7x12MonoHigh should return copies, not the same slice")
	}
	if !bytes.Equal(font1, font2) {
		t.Error("Multiple MCS7x12MonoHigh calls should return identical data")
	}
}

// Test font families comprehensive coverage
func TestFontFamiliesCoverage(t *testing.T) {
	// Test that we have fonts from all three families
	gseFonts := []func() []byte{GetGSE4x6, GetGSE5x7}
	mcsFonts := []func() []byte{GetMCS5x10Mono, GetMCS11Prop}
	verdanaFonts := []func() []byte{GetVerdana12, GetVerdana16}

	families := map[string][]func() []byte{
		"GSE":     gseFonts,
		"MCS":     mcsFonts,
		"Verdana": verdanaFonts,
	}

	for family, fonts := range families {
		for i, getter := range fonts {
			font := getter()
			if len(font) < 4 {
				t.Errorf("%s family font %d is too short: %d bytes", family, i, len(font))
			}

			// MCS proportional fonts have 128 chars, others have 96 (ASCII 32-127)
			numChars := int(font[3])
			expectedChars := 96
			// MCS family: index 0 is MCS5x10Mono (96 chars), index 1 is MCS11Prop (128 chars)
			if family == "MCS" && i == 1 {
				expectedChars = 128
			}
			if numChars != expectedChars {
				t.Errorf("%s family font %d should have %d chars, got %d", family, i, expectedChars, numChars)
			}
		}
	}
}
