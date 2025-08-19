package conv

import (
	"testing"

	"agg_go/internal/basics"
)

// MockRenderingBuffer implements RenderingBuffer for testing
type MockRenderingBuffer struct {
	width, height int
	data          [][]basics.Int8u
}

func NewMockRenderingBuffer(width, height, bytesPerPixel int) *MockRenderingBuffer {
	data := make([][]basics.Int8u, height)
	for y := range data {
		data[y] = make([]basics.Int8u, width*bytesPerPixel)
	}
	return &MockRenderingBuffer{
		width:  width,
		height: height,
		data:   data,
	}
}

func (m *MockRenderingBuffer) Width() int {
	return m.width
}

func (m *MockRenderingBuffer) Height() int {
	return m.height
}

func (m *MockRenderingBuffer) RowPtr(x, y, length int) []basics.Int8u {
	if y < 0 || y >= m.height || x < 0 {
		return nil
	}
	row := m.data[y]
	end := x + length
	if end > len(row) {
		end = len(row)
	}
	return row[x:end]
}

func (m *MockRenderingBuffer) RowPtrConst(y int) []basics.Int8u {
	if y < 0 || y >= m.height {
		return nil
	}
	return m.data[y]
}

func (m *MockRenderingBuffer) SetPixel(x, y int, value basics.Int8u) {
	if y >= 0 && y < m.height && x >= 0 && x < len(m.data[y]) {
		m.data[y][x] = value
	}
}

func (m *MockRenderingBuffer) GetPixel(x, y int) basics.Int8u {
	if y >= 0 && y < m.height && x >= 0 && x < len(m.data[y]) {
		return m.data[y][x]
	}
	return 0
}

// MockCopyRowFunctor for testing
type MockCopyRowFunctor struct {
	CallCount int
	LastWidth int
}

func (m *MockCopyRowFunctor) CopyRow(dst, src []basics.Int8u, width int) {
	m.CallCount++
	m.LastWidth = width
	// Simple copy for testing
	if len(dst) >= width && len(src) >= width {
		copy(dst[:width], src[:width])
	}
}

func TestColorConv(t *testing.T) {
	// Create test buffers
	src := NewMockRenderingBuffer(4, 3, 1)
	dst := NewMockRenderingBuffer(4, 3, 1)

	// Fill source buffer with test pattern
	for y := 0; y < 3; y++ {
		for x := 0; x < 4; x++ {
			src.SetPixel(x, y, basics.Int8u(y*10+x))
		}
	}

	// Test conversion
	functor := &MockCopyRowFunctor{}
	ColorConv(dst, src, functor)

	// Verify functor was called for each row
	if functor.CallCount != 3 {
		t.Errorf("Expected 3 calls, got %d", functor.CallCount)
	}

	if functor.LastWidth != 4 {
		t.Errorf("Expected last width 4, got %d", functor.LastWidth)
	}

	// Verify data was copied
	for y := 0; y < 3; y++ {
		for x := 0; x < 4; x++ {
			expected := basics.Int8u(y*10 + x)
			actual := dst.GetPixel(x, y)
			if actual != expected {
				t.Errorf("At (%d,%d): expected %d, got %d", x, y, expected, actual)
			}
		}
	}
}

func TestColorConvDifferentSizes(t *testing.T) {
	// Test with different buffer sizes - should use minimum dimensions
	src := NewMockRenderingBuffer(5, 4, 1)
	dst := NewMockRenderingBuffer(3, 2, 1)

	// Fill source
	for y := 0; y < 4; y++ {
		for x := 0; x < 5; x++ {
			src.SetPixel(x, y, basics.Int8u(100+y*10+x))
		}
	}

	functor := &MockCopyRowFunctor{}
	ColorConv(dst, src, functor)

	// Should only process 2 rows (dst height) with width 3 (dst width)
	if functor.CallCount != 2 {
		t.Errorf("Expected 2 calls, got %d", functor.CallCount)
	}

	if functor.LastWidth != 3 {
		t.Errorf("Expected last width 3, got %d", functor.LastWidth)
	}
}

func TestColorConvRow(t *testing.T) {
	src := []basics.Int8u{1, 2, 3, 4, 5}
	dst := make([]basics.Int8u, 5)

	functor := &MockCopyRowFunctor{}
	ColorConvRow(dst, src, 5, functor)

	if functor.CallCount != 1 {
		t.Errorf("Expected 1 call, got %d", functor.CallCount)
	}

	for i := 0; i < 5; i++ {
		if dst[i] != src[i] {
			t.Errorf("At index %d: expected %d, got %d", i, src[i], dst[i])
		}
	}
}

func TestColorConvSame(t *testing.T) {
	tests := []struct {
		name     string
		bpp      int
		width    int
		srcData  []basics.Int8u
		expected []basics.Int8u
	}{
		{
			name:     "1 BPP",
			bpp:      1,
			width:    3,
			srcData:  []basics.Int8u{1, 2, 3},
			expected: []basics.Int8u{1, 2, 3},
		},
		{
			name:     "3 BPP RGB",
			bpp:      3,
			width:    2,
			srcData:  []basics.Int8u{255, 128, 0, 64, 192, 32},
			expected: []basics.Int8u{255, 128, 0, 64, 192, 32},
		},
		{
			name:     "4 BPP RGBA",
			bpp:      4,
			width:    2,
			srcData:  []basics.Int8u{255, 128, 0, 255, 64, 192, 32, 128},
			expected: []basics.Int8u{255, 128, 0, 255, 64, 192, 32, 128},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := NewColorConvSame(tt.bpp)
			dst := make([]basics.Int8u, len(tt.srcData))

			conv.CopyRow(dst, tt.srcData, tt.width)

			for i, expected := range tt.expected {
				if dst[i] != expected {
					t.Errorf("At index %d: expected %d, got %d", i, expected, dst[i])
				}
			}
		})
	}
}

func TestColorConvSameEdgeCases(t *testing.T) {
	conv := NewColorConvSame(3)

	// Test zero width
	dst := make([]basics.Int8u, 6)
	src := []basics.Int8u{1, 2, 3, 4, 5, 6}
	conv.CopyRow(dst, src, 0)

	// Destination should remain unchanged (zeros)
	for i, val := range dst {
		if val != 0 {
			t.Errorf("Expected dst[%d] to remain 0, got %d", i, val)
		}
	}

	// Test insufficient destination buffer
	shortDst := make([]basics.Int8u, 2)
	conv.CopyRow(shortDst, src, 2) // Would need 6 bytes (2*3 BPP)

	// Should not crash and not copy anything
	for i, val := range shortDst {
		if val != 0 {
			t.Errorf("Expected shortDst[%d] to remain 0, got %d", i, val)
		}
	}
}

// MockPixelConverter for testing ConvRow
type MockPixelConverter struct {
	CallCount int
}

func (m *MockPixelConverter) ConvertPixel(dst, src []basics.Int8u) {
	m.CallCount++
	// Simple test conversion: invert first byte
	if len(dst) > 0 && len(src) > 0 {
		dst[0] = 255 - src[0]
	}
	// Copy remaining bytes
	for i := 1; i < len(dst) && i < len(src); i++ {
		dst[i] = src[i]
	}
}

func TestConvRow(t *testing.T) {
	pixelConverter := &MockPixelConverter{}
	conv := NewConvRow(pixelConverter, 3, 3) // 3 BPP for both src and dst

	src := []basics.Int8u{100, 150, 200, 50, 75, 25}
	dst := make([]basics.Int8u, 6)

	conv.CopyRow(dst, src, 2) // 2 pixels

	if pixelConverter.CallCount != 2 {
		t.Errorf("Expected 2 pixel conversions, got %d", pixelConverter.CallCount)
	}

	// First pixel: inverted R, same G, B
	if dst[0] != 155 { // 255 - 100
		t.Errorf("Expected dst[0] = 155, got %d", dst[0])
	}
	if dst[1] != 150 {
		t.Errorf("Expected dst[1] = 150, got %d", dst[1])
	}
	if dst[2] != 200 {
		t.Errorf("Expected dst[2] = 200, got %d", dst[2])
	}

	// Second pixel: inverted R, same G, B
	if dst[3] != 205 { // 255 - 50
		t.Errorf("Expected dst[3] = 205, got %d", dst[3])
	}
	if dst[4] != 75 {
		t.Errorf("Expected dst[4] = 75, got %d", dst[4])
	}
	if dst[5] != 25 {
		t.Errorf("Expected dst[5] = 25, got %d", dst[5])
	}
}

func TestConvRowDifferentPixelSizes(t *testing.T) {
	pixelConverter := &MockPixelConverter{}
	conv := NewConvRow(pixelConverter, 4, 3) // 3→4 BPP conversion

	src := []basics.Int8u{100, 150, 200} // 1 RGB pixel
	dst := make([]basics.Int8u, 4)       // 1 RGBA pixel

	conv.CopyRow(dst, src, 1) // 1 pixel

	if pixelConverter.CallCount != 1 {
		t.Errorf("Expected 1 pixel conversion, got %d", pixelConverter.CallCount)
	}
}

func TestConvert(t *testing.T) {
	// Test same format conversion (uses ColorConvSame)
	src := NewMockRenderingBuffer(2, 2, 3)
	dst := NewMockRenderingBuffer(2, 2, 3)

	// Fill source with test pattern by setting raw bytes
	testData := []basics.Int8u{255, 0, 128, 64, 192, 32, 16, 240, 80, 160, 120, 200}
	for i, val := range testData {
		y := i / 6
		x := i % 6
		src.SetPixel(x, y, val)
	}

	Convert(dst, src, 3, 3, nil)

	// Verify data was copied
	for i, expected := range testData {
		y := i / 6
		x := i % 6
		actual := dst.GetPixel(x, y)
		if actual != expected {
			t.Errorf("At byte index %d (%d,%d): expected %d, got %d", i, x, y, expected, actual)
		}
	}
}

func TestConvertWithPixelConverter(t *testing.T) {
	// Test different format conversion (uses ConvRow)
	src := NewMockRenderingBuffer(1, 1, 1)
	dst := NewMockRenderingBuffer(1, 1, 1)

	src.SetPixel(0, 0, 100)

	pixelConverter := &MockPixelConverter{}
	Convert(dst, src, 1, 1, pixelConverter)

	if pixelConverter.CallCount != 1 {
		t.Errorf("Expected 1 pixel conversion, got %d", pixelConverter.CallCount)
	}

	// Should be inverted
	if dst.GetPixel(0, 0) != 155 { // 255 - 100
		t.Errorf("Expected 155, got %d", dst.GetPixel(0, 0))
	}
}

func BenchmarkColorConvSame(b *testing.B) {
	conv := NewColorConvSame(3)
	src := make([]basics.Int8u, 1920*3) // Full HD width RGB
	dst := make([]basics.Int8u, 1920*3)

	// Fill source with pattern
	for i := range src {
		src[i] = basics.Int8u(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conv.CopyRow(dst, src, 1920)
	}
}

func BenchmarkColorConv(b *testing.B) {
	width, height := 640, 480
	src := NewMockRenderingBuffer(width, height, 3)
	dst := NewMockRenderingBuffer(width, height, 3)

	// Fill source with pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width*3; x++ {
			src.SetPixel(x, y, basics.Int8u((x+y)%256))
		}
	}

	conv := NewColorConvSame(3)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ColorConv(dst, src, conv)
	}
}
