package conv

import (
	"agg_go/internal/basics"
	"encoding/binary"
	"testing"
)

func TestColorConvRGB24(t *testing.T) {
	conv := NewColorConvRGB24()

	// Test data: RGB pixels
	src := []basics.Int8u{
		255, 0, 128, // Red=255, Green=0, Blue=128
		64, 192, 32, // Red=64, Green=192, Blue=32
	}
	dst := make([]basics.Int8u, 6)

	conv.CopyRow(dst, src, 2)

	// Should swap R and B channels
	expected := []basics.Int8u{
		128, 0, 255, // Blue=128, Green=0, Red=255
		32, 192, 64, // Blue=32, Green=192, Red=64
	}

	for i, exp := range expected {
		if dst[i] != exp {
			t.Errorf("At index %d: expected %d, got %d", i, exp, dst[i])
		}
	}
}

func TestColorConvRGBA32(t *testing.T) {
	// Test ARGB32 to ABGR32 conversion (0,3,2,1)
	conv := NewColorConvARGB32ToABGR32()

	src := []basics.Int8u{
		255, 128, 64, 32, // A=255, R=128, G=64, B=32
		200, 100, 50, 25, // A=200, R=100, G=50, B=25
	}
	dst := make([]basics.Int8u, 8)

	conv.CopyRow(dst, src, 2)

	// Should remap: A→A, R→B, G→G, B→R
	expected := []basics.Int8u{
		255, 32, 64, 128, // A=255, B=32, G=64, R=128
		200, 25, 50, 100, // A=200, B=25, G=50, R=100
	}

	for i, exp := range expected {
		if dst[i] != exp {
			t.Errorf("At index %d: expected %d, got %d", i, exp, dst[i])
		}
	}
}

func TestColorConvRGB24RGBA32(t *testing.T) {
	// Test RGB24 to RGBA32 conversion
	conv := NewColorConvRGB24ToRGBA32()

	src := []basics.Int8u{
		255, 128, 64, // R=255, G=128, B=64
		100, 50, 25, // R=100, G=50, B=25
	}
	dst := make([]basics.Int8u, 8)

	conv.CopyRow(dst, src, 2)

	// Should add full alpha (255) at the end
	expected := []basics.Int8u{
		255, 128, 64, 255, // R=255, G=128, B=64, A=255
		100, 50, 25, 255, // R=100, G=50, B=25, A=255
	}

	for i, exp := range expected {
		if dst[i] != exp {
			t.Errorf("At index %d: expected %d, got %d", i, exp, dst[i])
		}
	}
}

func TestColorConvRGBA32RGB24(t *testing.T) {
	// Test RGBA32 to RGB24 conversion
	conv := NewColorConvRGBA32ToRGB24()

	src := []basics.Int8u{
		255, 128, 64, 200, // R=255, G=128, B=64, A=200
		100, 50, 25, 150, // R=100, G=50, B=25, A=150
	}
	dst := make([]basics.Int8u, 6)

	conv.CopyRow(dst, src, 2)

	// Should drop alpha channel
	expected := []basics.Int8u{
		255, 128, 64, // R=255, G=128, B=64
		100, 50, 25, // R=100, G=50, B=25
	}

	for i, exp := range expected {
		if dst[i] != exp {
			t.Errorf("At index %d: expected %d, got %d", i, exp, dst[i])
		}
	}
}

func TestColorConvRGB555RGB24(t *testing.T) {
	conv := NewColorConvRGB555ToRGB24()

	// Create RGB555 test data
	src := make([]basics.Int8u, 4) // 2 pixels

	// First pixel: RGB555 = 0x7C1F (R=31, G=0, B=31 in 5-bit)
	binary.LittleEndian.PutUint16(src[0:], 0x7C1F)
	// Second pixel: RGB555 = 0x03E0 (R=0, G=31, B=0 in 5-bit)
	binary.LittleEndian.PutUint16(src[2:], 0x03E0)

	dst := make([]basics.Int8u, 6)
	conv.CopyRow(dst, src, 2)

	// RGB555 values should be expanded to 8-bit
	// R=31 (5-bit) → ~248 (8-bit), G=0 → 0, B=31 → ~248
	// Check approximate values due to bit expansion
	if dst[0] < 240 || dst[0] > 255 {
		t.Errorf("Expected red ~248, got %d", dst[0])
	}
	if dst[1] != 0 {
		t.Errorf("Expected green 0, got %d", dst[1])
	}
	if dst[2] < 240 || dst[2] > 255 {
		t.Errorf("Expected blue ~248, got %d", dst[2])
	}

	// Second pixel: pure green
	if dst[3] != 0 {
		t.Errorf("Expected red 0, got %d", dst[3])
	}
	if dst[4] < 240 || dst[4] > 255 {
		t.Errorf("Expected green ~248, got %d", dst[4])
	}
	if dst[5] != 0 {
		t.Errorf("Expected blue 0, got %d", dst[5])
	}
}

func TestColorConvRGB24RGB555(t *testing.T) {
	conv := NewColorConvRGB24ToRGB555()

	src := []basics.Int8u{
		248, 0, 248, // Near-maximum R and B
		0, 248, 0, // Near-maximum G
	}
	dst := make([]basics.Int8u, 4)

	conv.CopyRow(dst, src, 2)

	// First pixel should be close to 0x7C1F
	rgb555_1 := binary.LittleEndian.Uint16(dst[0:])
	if (rgb555_1 & 0x7C00) == 0 { // Red bits should be set
		t.Errorf("Expected red bits to be set in RGB555")
	}
	if (rgb555_1 & 0x001F) == 0 { // Blue bits should be set
		t.Errorf("Expected blue bits to be set in RGB555")
	}

	// Second pixel should be close to 0x03E0
	rgb555_2 := binary.LittleEndian.Uint16(dst[2:])
	if (rgb555_2 & 0x03E0) == 0 { // Green bits should be set
		t.Errorf("Expected green bits to be set in RGB555")
	}
}

func TestColorConvRGB565RGB24(t *testing.T) {
	conv := NewColorConvRGB565ToRGB24()

	// Create RGB565 test data
	src := make([]basics.Int8u, 4) // 2 pixels

	// First pixel: RGB565 = 0xF81F (R=31, G=0, B=31 in 565)
	binary.LittleEndian.PutUint16(src[0:], 0xF81F)
	// Second pixel: RGB565 = 0x07E0 (R=0, G=63, B=0 in 565)
	binary.LittleEndian.PutUint16(src[2:], 0x07E0)

	dst := make([]basics.Int8u, 6)
	conv.CopyRow(dst, src, 2)

	// RGB565 values should be expanded to 8-bit
	// R=31 (5-bit) → ~248, G=0 → 0, B=31 (5-bit) → ~248
	if dst[0] < 240 || dst[0] > 255 {
		t.Errorf("Expected red ~248, got %d", dst[0])
	}
	if dst[1] != 0 {
		t.Errorf("Expected green 0, got %d", dst[1])
	}
	if dst[2] < 240 || dst[2] > 255 {
		t.Errorf("Expected blue ~248, got %d", dst[2])
	}

	// Second pixel: pure green (6-bit max → ~252)
	if dst[3] != 0 {
		t.Errorf("Expected red 0, got %d", dst[3])
	}
	if dst[4] < 250 || dst[4] > 255 {
		t.Errorf("Expected green ~252, got %d", dst[4])
	}
	if dst[5] != 0 {
		t.Errorf("Expected blue 0, got %d", dst[5])
	}
}

func TestColorConvRGB24RGB565(t *testing.T) {
	conv := NewColorConvRGB24ToRGB565()

	src := []basics.Int8u{
		248, 0, 248, // Near-maximum R and B
		0, 252, 0, // Near-maximum G
	}
	dst := make([]basics.Int8u, 4)

	conv.CopyRow(dst, src, 2)

	// First pixel should be close to 0xF81F
	rgb565_1 := binary.LittleEndian.Uint16(dst[0:])
	if (rgb565_1 & 0xF800) == 0 { // Red bits should be set
		t.Errorf("Expected red bits to be set in RGB565")
	}
	if (rgb565_1 & 0x001F) == 0 { // Blue bits should be set
		t.Errorf("Expected blue bits to be set in RGB565")
	}

	// Second pixel should be close to 0x07E0
	rgb565_2 := binary.LittleEndian.Uint16(dst[2:])
	if (rgb565_2 & 0x07E0) == 0 { // Green bits should be set
		t.Errorf("Expected green bits to be set in RGB565")
	}
}

func TestColorConvRGB24Gray8(t *testing.T) {
	conv := NewColorConvRGB24ToGray8()

	src := []basics.Int8u{
		255, 0, 0, // Pure red
		0, 255, 0, // Pure green
		0, 0, 255, // Pure blue
		255, 255, 255, // White
		0, 0, 0, // Black
	}
	dst := make([]basics.Int8u, 5)

	conv.CopyRow(dst, src, 5)

	// ITU-R BT.601 weights: R=0.299, G=0.587, B=0.114
	// Using integer math: (77*R + 150*G + 29*B) >> 8

	// Pure red: (77*255 + 150*0 + 29*0) >> 8 = 19635 >> 8 = 76
	if dst[0] != 76 {
		t.Errorf("Red to gray: expected 76, got %d", dst[0])
	}

	// Pure green: (77*0 + 150*255 + 29*0) >> 8 = 38250 >> 8 = 149
	if dst[1] != 149 {
		t.Errorf("Green to gray: expected 149, got %d", dst[1])
	}

	// Pure blue: (77*0 + 150*0 + 29*255) >> 8 = 7395 >> 8 = 28
	if dst[2] != 28 {
		t.Errorf("Blue to gray: expected 28, got %d", dst[2])
	}

	// White: (77*255 + 150*255 + 29*255) >> 8 = 65280 >> 8 = 255
	if dst[3] != 255 {
		t.Errorf("White to gray: expected 255, got %d", dst[3])
	}

	// Black: all zeros
	if dst[4] != 0 {
		t.Errorf("Black to gray: expected 0, got %d", dst[4])
	}
}

func TestColorConvBGR24ToGray8(t *testing.T) {
	conv := NewColorConvBGR24ToGray8()

	// BGR format: Blue, Green, Red
	src := []basics.Int8u{
		0, 0, 255, // Pure red (in BGR)
		0, 255, 0, // Pure green (in BGR)
		255, 0, 0, // Pure blue (in BGR)
	}
	dst := make([]basics.Int8u, 3)

	conv.CopyRow(dst, src, 3)

	// Should produce same results as RGB24ToGray8 since we swapped input
	if dst[0] != 76 { // Red
		t.Errorf("BGR red to gray: expected 76, got %d", dst[0])
	}
	if dst[1] != 149 { // Green
		t.Errorf("BGR green to gray: expected 149, got %d", dst[1])
	}
	if dst[2] != 28 { // Blue
		t.Errorf("BGR blue to gray: expected 28, got %d", dst[2])
	}
}

func TestColorConvEdgeCases(t *testing.T) {
	conv := NewColorConvRGB24()

	// Test zero width
	src := []basics.Int8u{1, 2, 3}
	dst := make([]basics.Int8u, 3)
	conv.CopyRow(dst, src, 0)

	// Should not modify destination
	for i, val := range dst {
		if val != 0 {
			t.Errorf("Zero width: dst[%d] should remain 0, got %d", i, val)
		}
	}

	// Test insufficient buffer
	shortDst := make([]basics.Int8u, 2)
	conv.CopyRow(shortDst, src, 1) // Needs 3 bytes for 1 RGB pixel

	// Should not crash and not modify anything
	for i, val := range shortDst {
		if val != 0 {
			t.Errorf("Insufficient buffer: dst[%d] should remain 0, got %d", i, val)
		}
	}
}

func TestAllRGBA32Conversions(t *testing.T) {
	// Test all common RGBA32 permutations
	testCases := []struct {
		name    string
		conv    CopyRowFunctor
		mapping [4]int // Expected channel mapping
	}{
		{"ARGB32ToABGR32", NewColorConvARGB32ToABGR32(), [4]int{0, 3, 2, 1}},
		{"ARGB32ToBGRA32", NewColorConvARGB32ToBGRA32(), [4]int{3, 2, 1, 0}},
		{"ARGB32ToRGBA32", NewColorConvARGB32ToRGBA32(), [4]int{1, 2, 3, 0}},
		{"RGBA32ToARGB32", NewColorConvRGBA32ToARGB32(), [4]int{3, 0, 1, 2}},
		{"RGBA32ToBGRA32", NewColorConvRGBA32ToBGRA32(), [4]int{2, 1, 0, 3}},
	}

	src := []basics.Int8u{10, 20, 30, 40} // A=10, R=20, G=30, B=40 (or similar)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dst := make([]basics.Int8u, 4)
			tc.conv.CopyRow(dst, src, 1)

			for i, expectedSrcIdx := range tc.mapping {
				if dst[i] != src[expectedSrcIdx] {
					t.Errorf("Channel %d: expected src[%d]=%d, got %d",
						i, expectedSrcIdx, src[expectedSrcIdx], dst[i])
				}
			}
		})
	}
}

func BenchmarkColorConvRGB24(b *testing.B) {
	conv := NewColorConvRGB24()
	src := make([]basics.Int8u, 1920*3) // Full HD width
	dst := make([]basics.Int8u, 1920*3)

	// Fill with test pattern
	for i := range src {
		src[i] = basics.Int8u(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conv.CopyRow(dst, src, 1920)
	}
}

func BenchmarkColorConvRGBA32(b *testing.B) {
	conv := NewColorConvARGB32ToRGBA32()
	src := make([]basics.Int8u, 1920*4) // Full HD width RGBA
	dst := make([]basics.Int8u, 1920*4)

	// Fill with test pattern
	for i := range src {
		src[i] = basics.Int8u(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conv.CopyRow(dst, src, 1920)
	}
}

func BenchmarkColorConvRGB24Gray8(b *testing.B) {
	conv := NewColorConvRGB24ToGray8()
	src := make([]basics.Int8u, 1920*3)
	dst := make([]basics.Int8u, 1920)

	// Fill with test pattern
	for i := range src {
		src[i] = basics.Int8u(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conv.CopyRow(dst, src, 1920)
	}
}
