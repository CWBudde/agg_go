package conv

import (
	"encoding/binary"
	"testing"

	"agg_go/internal/basics"
)

func TestColorConvGray16ToGray8(t *testing.T) {
	conv := NewColorConvGray16ToGray8()

	// Create test data: 16-bit grayscale values
	src := make([]basics.Int8u, 6)                 // 3 pixels
	binary.LittleEndian.PutUint16(src[0:], 0x1234) // Gray value 0x1234 → high byte 0x12
	binary.LittleEndian.PutUint16(src[2:], 0x5678) // Gray value 0x5678 → high byte 0x56
	binary.LittleEndian.PutUint16(src[4:], 0xFF00) // Gray value 0xFF00 → high byte 0xFF

	dst := make([]basics.Int8u, 3)
	conv.CopyRow(dst, src, 3)

	// Should extract high bytes
	expected := []basics.Int8u{0x12, 0x56, 0xFF}
	for i, exp := range expected {
		if dst[i] != exp {
			t.Errorf("At index %d: expected %d, got %d", i, exp, dst[i])
		}
	}
}

func TestColorConvRGB24RGB48(t *testing.T) {
	conv := NewColorConvRGB24ToRGB48()

	src := []basics.Int8u{
		255, 128, 64, // R=255, G=128, B=64
		100, 50, 25, // R=100, G=50, B=25
	}
	dst := make([]basics.Int8u, 12) // 2 pixels × 6 bytes each

	conv.CopyRow(dst, src, 2)

	// Each 8-bit value should become 16-bit by: (value << 8) | value
	expected16 := []uint16{
		0xFFFF, 0x8080, 0x4040, // First pixel
		0x6464, 0x3232, 0x1919, // Second pixel
	}

	for i, exp := range expected16 {
		actual := binary.LittleEndian.Uint16(dst[i*2:])
		if actual != exp {
			t.Errorf("At 16-bit index %d: expected 0x%04X, got 0x%04X", i, exp, actual)
		}
	}
}

func TestColorConvRGB48RGB24(t *testing.T) {
	conv := NewColorConvRGB48ToRGB24()

	// Create RGB48 test data
	src := make([]basics.Int8u, 12) // 2 pixels × 6 bytes each

	// First pixel: R=0x1234, G=0x5678, B=0x9ABC
	binary.LittleEndian.PutUint16(src[0:], 0x1234)
	binary.LittleEndian.PutUint16(src[2:], 0x5678)
	binary.LittleEndian.PutUint16(src[4:], 0x9ABC)

	// Second pixel: R=0xFF00, G=0x8000, B=0x4000
	binary.LittleEndian.PutUint16(src[6:], 0xFF00)
	binary.LittleEndian.PutUint16(src[8:], 0x8000)
	binary.LittleEndian.PutUint16(src[10:], 0x4000)

	dst := make([]basics.Int8u, 6)
	conv.CopyRow(dst, src, 2)

	// Should extract high bytes
	expected := []basics.Int8u{
		0x12, 0x56, 0x9A, // First pixel high bytes
		0xFF, 0x80, 0x40, // Second pixel high bytes
	}

	for i, exp := range expected {
		if dst[i] != exp {
			t.Errorf("At index %d: expected 0x%02X, got 0x%02X", i, exp, dst[i])
		}
	}
}

func TestColorConvRGBAAAToRGB24(t *testing.T) {
	conv := NewColorConvRGBAAAToRGB24Std()

	// Create 10-bit AAA format test data
	src := make([]basics.Int8u, 4)

	// RRRRRRRRRRGGGGGGGGGGBBBBBBBBBBAA format
	// Let's create: R=1023(max), G=512, B=0, A=0
	// Binary: 1111111111 1000000000 0000000000 00
	rgbAAA := uint32(0x3FF<<22) | uint32(0x200<<12) | uint32(0x000<<2) | uint32(0x00)
	binary.LittleEndian.PutUint32(src, rgbAAA)

	dst := make([]basics.Int8u, 3)
	conv.CopyRow(dst, src, 1)

	// Should extract 10-bit values and convert to 8-bit by taking high 8 bits
	// R: (rgbAAA >> 22) & 0x3FF → 1023 → high 8 bits
	// G: (rgbAAA >> 12) & 0x3FF → 512 → high 8 bits
	// B: (rgbAAA >> 2) & 0x3FF → 0 → high 8 bits

	expectedR := basics.Int8u(rgbAAA >> 22)
	expectedG := basics.Int8u(rgbAAA >> 12)
	expectedB := basics.Int8u(rgbAAA >> 2)

	if dst[0] != expectedR {
		t.Errorf("Red: expected %d, got %d", expectedR, dst[0])
	}
	if dst[1] != expectedG {
		t.Errorf("Green: expected %d, got %d", expectedG, dst[1])
	}
	if dst[2] != expectedB {
		t.Errorf("Blue: expected %d, got %d", expectedB, dst[2])
	}
}

func TestColorConvRGBBBAToRGB24(t *testing.T) {
	conv := NewColorConvRGBBBAToRGB24Std()

	// Create 10-bit BBA format test data
	src := make([]basics.Int8u, 4)

	// RRRRRRRRRRRRGGGGGGGGGGBBBBBBBBAA format (12+11+9 bits)
	// R=4095(max), G=1024, B=256, A=0
	rgbBBA := uint32(0xFFF<<20) | uint32(0x400<<9) | uint32(0x100<<2) | uint32(0x00)
	binary.LittleEndian.PutUint32(src, rgbBBA)

	dst := make([]basics.Int8u, 3)
	conv.CopyRow(dst, src, 1)

	expectedR := basics.Int8u(rgbBBA >> 24)
	expectedG := basics.Int8u(rgbBBA >> 13)
	expectedB := basics.Int8u(rgbBBA >> 2)

	if dst[0] != expectedR {
		t.Errorf("Red: expected %d, got %d", expectedR, dst[0])
	}
	if dst[1] != expectedG {
		t.Errorf("Green: expected %d, got %d", expectedG, dst[1])
	}
	if dst[2] != expectedB {
		t.Errorf("Blue: expected %d, got %d", expectedB, dst[2])
	}
}

func TestColorConvBGRABBToRGB24(t *testing.T) {
	conv := NewColorConvBGRABBToRGB24Std()

	// Create 10-bit ABB format test data
	src := make([]basics.Int8u, 4)

	// BBBBBBBBBBGGGGGGGGGGAARRRRRRRRR format
	// B=1023, G=512, A=256, R=128
	bgrABB := uint32(0x3FF<<22) | uint32(0x200<<12) | uint32(0x100<<3) | uint32(0x80>>3)
	binary.LittleEndian.PutUint32(src, bgrABB)

	dst := make([]basics.Int8u, 3)
	conv.CopyRow(dst, src, 1)

	// Note: this converter expects BGR format input but outputs RGB
	expectedR := basics.Int8u(bgrABB >> 3)  // R from low bits
	expectedG := basics.Int8u(bgrABB >> 14) // G from middle
	expectedB := basics.Int8u(bgrABB >> 24) // B from high bits

	if dst[0] != expectedR {
		t.Errorf("Red: expected %d, got %d", expectedR, dst[0])
	}
	if dst[1] != expectedG {
		t.Errorf("Green: expected %d, got %d", expectedG, dst[1])
	}
	if dst[2] != expectedB {
		t.Errorf("Blue: expected %d, got %d", expectedB, dst[2])
	}
}

func TestColorConvRGBA64RGBA32(t *testing.T) {
	conv := NewColorConvRGBA64ToRGBA32()

	// Create RGBA64 test data
	src := make([]basics.Int8u, 8) // 1 pixel × 8 bytes

	// R=0x1234, G=0x5678, B=0x9ABC, A=0xDEF0
	binary.LittleEndian.PutUint16(src[0:], 0x1234)
	binary.LittleEndian.PutUint16(src[2:], 0x5678)
	binary.LittleEndian.PutUint16(src[4:], 0x9ABC)
	binary.LittleEndian.PutUint16(src[6:], 0xDEF0)

	dst := make([]basics.Int8u, 4)
	conv.CopyRow(dst, src, 1)

	// Should extract high bytes in same order
	expected := []basics.Int8u{0x12, 0x56, 0x9A, 0xDE}
	for i, exp := range expected {
		if dst[i] != exp {
			t.Errorf("At index %d: expected 0x%02X, got 0x%02X", i, exp, dst[i])
		}
	}
}

func TestColorConvRGBA64ChannelReordering(t *testing.T) {
	// Test ARGB64 to BGRA32 conversion (channels 3,2,1,0)
	conv := NewColorConvARGB64ToBGRA32()

	src := make([]basics.Int8u, 8)
	// A=0x10AA, R=0x20BB, G=0x30CC, B=0x40DD
	binary.LittleEndian.PutUint16(src[0:], 0x10AA) // A
	binary.LittleEndian.PutUint16(src[2:], 0x20BB) // R
	binary.LittleEndian.PutUint16(src[4:], 0x30CC) // G
	binary.LittleEndian.PutUint16(src[6:], 0x40DD) // B

	dst := make([]basics.Int8u, 4)
	conv.CopyRow(dst, src, 1)

	// Should reorder: A,R,G,B → B,G,R,A (channels 3,2,1,0)
	expected := []basics.Int8u{0x40, 0x30, 0x20, 0x10} // High bytes of B,G,R,A
	for i, exp := range expected {
		if dst[i] != exp {
			t.Errorf("At index %d: expected 0x%02X, got 0x%02X", i, exp, dst[i])
		}
	}
}

func TestColorConvRGB24RGBA64(t *testing.T) {
	conv := NewColorConvRGB24ToRGBA64()

	src := []basics.Int8u{255, 128, 64} // R=255, G=128, B=64
	dst := make([]basics.Int8u, 8)

	conv.CopyRow(dst, src, 1)

	// Each 8-bit value should become 16-bit: (value << 8) | value
	// Alpha should be 0xFFFF (65535)
	expectedValues := []uint16{0xFFFF, 0x8080, 0x4040, 0xFFFF} // R,G,B,A

	for i, exp := range expectedValues {
		actual := binary.LittleEndian.Uint16(dst[i*2:])
		if actual != exp {
			t.Errorf("At component %d: expected 0x%04X, got 0x%04X", i, exp, actual)
		}
	}
}

func TestColorConvRGB24RGBA64ChannelReordering(t *testing.T) {
	// Test BGR24 to ARGB64 conversion
	conv := NewColorConvBGR24ToARGB64()

	src := []basics.Int8u{64, 128, 255} // B=64, G=128, R=255 (BGR format)
	dst := make([]basics.Int8u, 8)

	conv.CopyRow(dst, src, 1)

	// Should map BGR24 to ARGB64: B,G,R → A,R,G,B with A=0xFFFF
	// Expecting: A=0xFFFF, R=0xFFFF, G=0x8080, B=0x4040
	expectedValues := []uint16{0xFFFF, 0xFFFF, 0x8080, 0x4040} // A,R,G,B

	for i, exp := range expectedValues {
		actual := binary.LittleEndian.Uint16(dst[i*2:])
		if actual != exp {
			t.Errorf("At component %d: expected 0x%04X, got 0x%04X", i, exp, actual)
		}
	}
}

func TestColorConvRGB24Gray16(t *testing.T) {
	conv := NewColorConvRGB24ToGray16()

	src := []basics.Int8u{
		255, 0, 0, // Pure red
		0, 255, 0, // Pure green
		0, 0, 255, // Pure blue
		255, 255, 255, // White
		0, 0, 0, // Black
	}
	dst := make([]basics.Int8u, 10) // 5 pixels × 2 bytes each

	conv.CopyRow(dst, src, 5)

	// ITU-R BT.601 weights for 16-bit: 77*R + 150*G + 29*B
	expectedValues := []uint16{
		77 * 255,                  // Pure red
		150 * 255,                 // Pure green
		29 * 255,                  // Pure blue
		77*255 + 150*255 + 29*255, // White
		0,                         // Black
	}

	for i, exp := range expectedValues {
		actual := binary.LittleEndian.Uint16(dst[i*2:])
		if actual != exp {
			t.Errorf("At pixel %d: expected %d, got %d", i, exp, actual)
		}
	}
}

func TestColorConvBGR24Gray16(t *testing.T) {
	conv := NewColorConvBGR24ToGray16()

	// BGR format: Blue, Green, Red
	src := []basics.Int8u{
		0, 0, 255, // Pure red (in BGR)
		0, 255, 0, // Pure green (in BGR)
		255, 0, 0, // Pure blue (in BGR)
	}
	dst := make([]basics.Int8u, 6) // 3 pixels × 2 bytes each

	conv.CopyRow(dst, src, 3)

	// Should produce same results as RGB since we account for BGR ordering
	expectedValues := []uint16{
		77 * 255,  // Red
		150 * 255, // Green
		29 * 255,  // Blue
	}

	for i, exp := range expectedValues {
		actual := binary.LittleEndian.Uint16(dst[i*2:])
		if actual != exp {
			t.Errorf("At pixel %d: expected %d, got %d", i, exp, actual)
		}
	}
}

func TestColorConv16BitEdgeCases(t *testing.T) {
	// Test zero width
	conv := NewColorConvRGB24ToRGB48()
	src := []basics.Int8u{1, 2, 3}
	dst := make([]basics.Int8u, 6)
	conv.CopyRow(dst, src, 0)

	// Should not modify destination
	for i, val := range dst {
		if val != 0 {
			t.Errorf("Zero width: dst[%d] should remain 0, got %d", i, val)
		}
	}

	// Test insufficient buffer
	conv2 := NewColorConvGray16ToGray8()
	shortDst := make([]basics.Int8u, 1)
	shortSrc := make([]basics.Int8u, 2)
	binary.LittleEndian.PutUint16(shortSrc, 0x1234)

	conv2.CopyRow(shortDst, shortSrc, 2) // Needs 4 bytes source, 2 bytes dest

	// Should not crash - may or may not convert depending on bounds check
}

func TestAllRGB48Conversions(t *testing.T) {
	testCases := []struct {
		name   string
		conv   CopyRowFunctor
		swapRB bool
	}{
		{"RGB24ToRGB48", NewColorConvRGB24ToRGB48(), false},
		{"BGR24ToBGR48", NewColorConvBGR24ToBGR48(), false},
		{"RGB24ToBGR48", NewColorConvRGB24ToBGR48(), true},
		{"BGR24ToRGB48", NewColorConvBGR24ToRGB48(), true},
	}

	src := []basics.Int8u{100, 150, 200} // Test RGB values

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dst := make([]basics.Int8u, 6)
			tc.conv.CopyRow(dst, src, 1)

			// Check that values were expanded properly
			r16 := binary.LittleEndian.Uint16(dst[0:])
			g16 := binary.LittleEndian.Uint16(dst[2:])
			b16 := binary.LittleEndian.Uint16(dst[4:])

			expectedR := uint16(src[0])<<8 | uint16(src[0])
			expectedG := uint16(src[1])<<8 | uint16(src[1])
			expectedB := uint16(src[2])<<8 | uint16(src[2])

			if tc.swapRB {
				// R and B should be swapped
				if r16 != expectedB {
					t.Errorf("Expected R=B=%04X, got %04X", expectedB, r16)
				}
				if b16 != expectedR {
					t.Errorf("Expected B=R=%04X, got %04X", expectedR, b16)
				}
			} else {
				// R and B should be same
				if r16 != expectedR {
					t.Errorf("Expected R=%04X, got %04X", expectedR, r16)
				}
				if b16 != expectedB {
					t.Errorf("Expected B=%04X, got %04X", expectedB, b16)
				}
			}

			// G should always be the same
			if g16 != expectedG {
				t.Errorf("Expected G=%04X, got %04X", expectedG, g16)
			}
		})
	}
}

func BenchmarkColorConvGray16ToGray8(b *testing.B) {
	conv := NewColorConvGray16ToGray8()
	src := make([]basics.Int8u, 1920*2) // Full HD width, 16-bit
	dst := make([]basics.Int8u, 1920)   // Full HD width, 8-bit

	// Fill with test pattern
	for i := 0; i < 1920; i++ {
		binary.LittleEndian.PutUint16(src[i*2:], uint16(i%65536))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conv.CopyRow(dst, src, 1920)
	}
}

func BenchmarkColorConvRGB24RGB48(b *testing.B) {
	conv := NewColorConvRGB24ToRGB48()
	src := make([]basics.Int8u, 1920*3) // Full HD width RGB24
	dst := make([]basics.Int8u, 1920*6) // Full HD width RGB48

	// Fill with test pattern
	for i := range src {
		src[i] = basics.Int8u(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conv.CopyRow(dst, src, 1920)
	}
}

func BenchmarkColorConvRGBA64RGBA32(b *testing.B) {
	conv := NewColorConvRGBA64ToRGBA32()
	src := make([]basics.Int8u, 1920*8) // Full HD width RGBA64
	dst := make([]basics.Int8u, 1920*4) // Full HD width RGBA32

	// Fill with test pattern
	for i := 0; i < 1920*4; i++ {
		val := uint16(i % 65536)
		binary.LittleEndian.PutUint16(src[i*2:], val)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conv.CopyRow(dst, src, 1920)
	}
}
