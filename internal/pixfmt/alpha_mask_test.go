package pixfmt

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/buffer"
)

// TestOneComponentMaskU8 tests the OneComponentMaskU8 mask function
func TestOneComponentMaskU8(t *testing.T) {
	mask := OneComponentMaskU8{}

	// Test with valid data
	data := []basics.Int8u{128, 255, 0, 64}
	result := mask.Calculate(data)
	if result != 128 {
		t.Errorf("Expected 128, got %d", result)
	}

	// Test with empty data
	result = mask.Calculate([]basics.Int8u{})
	if result != 0 {
		t.Errorf("Expected 0 for empty data, got %d", result)
	}

	// Test with single element
	result = mask.Calculate([]basics.Int8u{200})
	if result != 200 {
		t.Errorf("Expected 200, got %d", result)
	}
}

// TestRGBToGrayMaskU8 tests the RGB to grayscale conversion mask function
func TestRGBToGrayMaskU8(t *testing.T) {
	mask := NewRGBToGrayMaskU8(0, 1, 2)

	// Test with RGB data: R=255, G=0, B=0 (red)
	data := []basics.Int8u{255, 0, 0}
	result := mask.Calculate(data)
	expected := basics.Int8u((255*77 + 0*150 + 0*29) >> 8) // Should be around 75
	if result != expected {
		t.Errorf("Red conversion: expected %d, got %d", expected, result)
	}

	// Test with RGB data: R=0, G=255, B=0 (green)
	data = []basics.Int8u{0, 255, 0}
	result = mask.Calculate(data)
	expected = basics.Int8u((0*77 + 255*150 + 0*29) >> 8) // Should be around 149
	if result != expected {
		t.Errorf("Green conversion: expected %d, got %d", expected, result)
	}

	// Test with RGB data: R=0, G=0, B=255 (blue)
	data = []basics.Int8u{0, 0, 255}
	result = mask.Calculate(data)
	expected = basics.Int8u((0*77 + 0*150 + 255*29) >> 8) // Should be around 28
	if result != expected {
		t.Errorf("Blue conversion: expected %d, got %d", expected, result)
	}

	// Test with white (all 255)
	data = []basics.Int8u{255, 255, 255}
	result = mask.Calculate(data)
	expected = basics.Int8u((255*77 + 255*150 + 255*29) >> 8) // Should be around 255
	if result != expected {
		t.Errorf("White conversion: expected %d, got %d", expected, result)
	}

	// Test with insufficient data
	result = mask.Calculate([]basics.Int8u{100, 150})
	if result != 0 {
		t.Errorf("Expected 0 for insufficient data, got %d", result)
	}
}

// TestBGRToGrayMaskU8 tests BGR to grayscale conversion
func TestBGRToGrayMaskU8(t *testing.T) {
	mask := NewRGBToGrayMaskU8(2, 1, 0) // BGR order

	// Test with BGR data: B=255, G=0, R=0 (blue in BGR format)
	data := []basics.Int8u{255, 0, 0}
	result := mask.Calculate(data)
	expected := basics.Int8u((0*77 + 0*150 + 255*29) >> 8) // Blue component
	if result != expected {
		t.Errorf("BGR blue conversion: expected %d, got %d", expected, result)
	}
}

// TestAlphaMaskU8BasicOperations tests basic pixel operations
func TestAlphaMaskU8BasicOperations(t *testing.T) {
	// Create a 4x4 mask buffer with some test data
	width, height := 4, 4
	maskData := make([]basics.Int8u, width*height)

	// Fill with a gradient pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			maskData[y*width+x] = basics.Int8u((x + y) * 32)
		}
	}

	rbuf := buffer.NewRenderingBufferU8WithData(maskData, width, height, width)
	mask := NewAlphaMaskU8(1, 0, OneComponentMaskU8{})
	mask.Attach(rbuf)

	// Test Pixel method
	pixel := mask.Pixel(1, 1)
	expected := basics.Int8u((1 + 1) * 32) // x=1, y=1 -> (1+1)*32 = 64
	if pixel != expected {
		t.Errorf("Pixel(1,1): expected %d, got %d", expected, pixel)
	}

	// Test out of bounds
	pixel = mask.Pixel(-1, 0)
	if pixel != 0 {
		t.Errorf("Out of bounds pixel should be 0, got %d", pixel)
	}

	pixel = mask.Pixel(0, -1)
	if pixel != 0 {
		t.Errorf("Out of bounds pixel should be 0, got %d", pixel)
	}

	pixel = mask.Pixel(width, 0)
	if pixel != 0 {
		t.Errorf("Out of bounds pixel should be 0, got %d", pixel)
	}

	pixel = mask.Pixel(0, height)
	if pixel != 0 {
		t.Errorf("Out of bounds pixel should be 0, got %d", pixel)
	}
}

// TestAlphaMaskU8CombinePixel tests the CombinePixel method
func TestAlphaMaskU8CombinePixel(t *testing.T) {
	// Create a simple 2x2 mask with known values
	width, height := 2, 2
	maskData := []basics.Int8u{255, 128, 64, 0} // Full, half, quarter, none

	rbuf := buffer.NewRenderingBufferU8WithData(maskData, width, height, width)
	mask := NewAlphaMaskU8(1, 0, OneComponentMaskU8{})
	mask.Attach(rbuf)

	// Test combination with full mask (255)
	result := mask.CombinePixel(0, 0, 100)
	expected := basics.Int8u((CoverFull + 100*255) >> CoverShift)
	if result != expected {
		t.Errorf("CombinePixel with full mask: expected %d, got %d", expected, result)
	}

	// Test combination with half mask (128)
	result = mask.CombinePixel(1, 0, 100)
	expected = basics.Int8u((CoverFull + 100*128) >> CoverShift)
	if result != expected {
		t.Errorf("CombinePixel with half mask: expected %d, got %d", expected, result)
	}

	// Test combination with zero mask (0)
	result = mask.CombinePixel(1, 1, 100)
	expected = basics.Int8u((CoverFull + 100*0) >> CoverShift)
	if result != expected {
		t.Errorf("CombinePixel with zero mask: expected %d, got %d", expected, result)
	}
}

// TestAlphaMaskU8FillHspan tests horizontal span filling
func TestAlphaMaskU8FillHspan(t *testing.T) {
	// Create a 5x3 mask with gradient
	width, height := 5, 3
	maskData := make([]basics.Int8u, width*height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			maskData[y*width+x] = basics.Int8u(x * 50) // 0, 50, 100, 150, 200
		}
	}

	rbuf := buffer.NewRenderingBufferU8WithData(maskData, width, height, width)
	mask := NewAlphaMaskU8(1, 0, OneComponentMaskU8{})
	mask.Attach(rbuf)

	// Test normal span
	dst := make([]basics.Int8u, 3)
	mask.FillHspan(1, 1, dst, 3) // x=1, y=1, length=3
	expected := []basics.Int8u{50, 100, 150}
	for i := 0; i < 3; i++ {
		if dst[i] != expected[i] {
			t.Errorf("FillHspan[%d]: expected %d, got %d", i, expected[i], dst[i])
		}
	}

	// Test span with negative x
	dst = make([]basics.Int8u, 3)
	mask.FillHspan(-1, 1, dst, 3)       // x=-1, y=1, length=3
	expected = []basics.Int8u{0, 0, 50} // First two should be 0, third should be maskData[1*5+0] = 0
	for i := 0; i < 3; i++ {
		if dst[i] != expected[i] {
			t.Errorf("FillHspan negative x[%d]: expected %d, got %d", i, expected[i], dst[i])
		}
	}

	// Test span going beyond width
	dst = make([]basics.Int8u, 3)
	mask.FillHspan(3, 1, dst, 3)           // x=3, y=1, length=3 (goes beyond width=5)
	expected = []basics.Int8u{150, 200, 0} // Last should be 0
	for i := 0; i < 3; i++ {
		if dst[i] != expected[i] {
			t.Errorf("FillHspan beyond width[%d]: expected %d, got %d", i, expected[i], dst[i])
		}
	}

	// Test span with out of bounds y
	dst = make([]basics.Int8u, 3)
	mask.FillHspan(0, -1, dst, 3) // y=-1
	for i := 0; i < 3; i++ {
		if dst[i] != 0 {
			t.Errorf("FillHspan out of bounds y[%d]: expected 0, got %d", i, dst[i])
		}
	}
}

// TestAlphaMaskU8CombineHspan tests horizontal span combination
func TestAlphaMaskU8CombineHspan(t *testing.T) {
	// Create a simple 3x1 mask
	width, height := 3, 1
	maskData := []basics.Int8u{255, 128, 64} // Full, half, quarter

	rbuf := buffer.NewRenderingBufferU8WithData(maskData, width, height, width)
	mask := NewAlphaMaskU8(1, 0, OneComponentMaskU8{})
	mask.Attach(rbuf)

	// Test combination
	dst := []basics.Int8u{100, 100, 100}
	mask.CombineHspan(0, 0, dst, 3)

	expected := []basics.Int8u{
		basics.Int8u((CoverFull + 100*255) >> CoverShift),
		basics.Int8u((CoverFull + 100*128) >> CoverShift),
		basics.Int8u((CoverFull + 100*64) >> CoverShift),
	}

	for i := 0; i < 3; i++ {
		if dst[i] != expected[i] {
			t.Errorf("CombineHspan[%d]: expected %d, got %d", i, expected[i], dst[i])
		}
	}
}

// TestAlphaMaskU8FillVspan tests vertical span filling
func TestAlphaMaskU8FillVspan(t *testing.T) {
	// Create a 3x5 mask
	width, height := 3, 5
	maskData := make([]basics.Int8u, width*height)

	// Fill column 1 with gradient values
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			maskData[y*width+x] = basics.Int8u(y * 50) // 0, 50, 100, 150, 200
		}
	}

	rbuf := buffer.NewRenderingBufferU8WithData(maskData, width, height, width)
	mask := NewAlphaMaskU8(1, 0, OneComponentMaskU8{})
	mask.Attach(rbuf)

	// Test normal vertical span
	dst := make([]basics.Int8u, 3)
	mask.FillVspan(1, 1, dst, 3) // x=1, y=1, length=3
	expected := []basics.Int8u{50, 100, 150}
	for i := 0; i < 3; i++ {
		if dst[i] != expected[i] {
			t.Errorf("FillVspan[%d]: expected %d, got %d", i, expected[i], dst[i])
		}
	}

	// Test span with out of bounds x
	dst = make([]basics.Int8u, 3)
	mask.FillVspan(-1, 0, dst, 3) // x=-1
	for i := 0; i < 3; i++ {
		if dst[i] != 0 {
			t.Errorf("FillVspan out of bounds x[%d]: expected 0, got %d", i, dst[i])
		}
	}
}

// TestAlphaMaskU8CombineVspan tests vertical span combination
func TestAlphaMaskU8CombineVspan(t *testing.T) {
	// Create a simple 1x3 mask
	width, height := 1, 3
	maskData := []basics.Int8u{255, 128, 64} // Full, half, quarter

	rbuf := buffer.NewRenderingBufferU8WithData(maskData, width, height, width)
	mask := NewAlphaMaskU8(1, 0, OneComponentMaskU8{})
	mask.Attach(rbuf)

	// Test combination
	dst := []basics.Int8u{100, 100, 100}
	mask.CombineVspan(0, 0, dst, 3)

	expected := []basics.Int8u{
		basics.Int8u((CoverFull + 100*255) >> CoverShift),
		basics.Int8u((CoverFull + 100*128) >> CoverShift),
		basics.Int8u((CoverFull + 100*64) >> CoverShift),
	}

	for i := 0; i < 3; i++ {
		if dst[i] != expected[i] {
			t.Errorf("CombineVspan[%d]: expected %d, got %d", i, expected[i], dst[i])
		}
	}
}

// TestAMaskNoClipU8 tests the no-clip variant
func TestAMaskNoClipU8(t *testing.T) {
	// Create a 3x3 mask
	width, height := 3, 3
	maskData := make([]basics.Int8u, width*height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			maskData[y*width+x] = basics.Int8u((x + y) * 40)
		}
	}

	rbuf := buffer.NewRenderingBufferU8WithData(maskData, width, height, width)
	mask := NewAMaskNoClipU8(1, 0, OneComponentMaskU8{})
	mask.Attach(rbuf)

	// Test basic pixel access (no bounds checking in this variant)
	pixel := mask.Pixel(1, 1)
	expected := basics.Int8u((1 + 1) * 40) // 80
	if pixel != expected {
		t.Errorf("NoClip Pixel(1,1): expected %d, got %d", expected, pixel)
	}

	// Test combine pixel
	result := mask.CombinePixel(0, 0, 100)
	maskVal := basics.Int8u((0 + 0) * 40) // 0
	expected = basics.Int8u((CoverFull + 100*int(maskVal)) >> CoverShift)
	if result != expected {
		t.Errorf("NoClip CombinePixel: expected %d, got %d", expected, result)
	}

	// Test horizontal span
	dst := make([]basics.Int8u, 2)
	mask.FillHspan(0, 1, dst, 2)
	expected0 := basics.Int8u((0 + 1) * 40) // 40
	expected1 := basics.Int8u((1 + 1) * 40) // 80
	if dst[0] != expected0 || dst[1] != expected1 {
		t.Errorf("NoClip FillHspan: expected [%d, %d], got [%d, %d]", expected0, expected1, dst[0], dst[1])
	}
}

// TestAlphaMaskU8WithRGB24 tests RGB24 component masks
func TestAlphaMaskU8WithRGB24(t *testing.T) {
	// Create RGB24 data: 2x1 image with (255,128,64) and (0,100,200)
	width, height := 2, 1
	rgbData := []basics.Int8u{
		255, 128, 64, // Pixel 0: R=255, G=128, B=64
		0, 100, 200, // Pixel 1: R=0, G=100, B=200
	}

	rbuf := buffer.NewRenderingBufferU8WithData(rgbData, width, height, width*3)

	// Test R component mask
	maskR := NewAlphaMaskU8(3, 0, OneComponentMaskU8{})
	maskR.Attach(rbuf)

	if maskR.Pixel(0, 0) != 255 {
		t.Errorf("RGB24 R component: expected 255, got %d", maskR.Pixel(0, 0))
	}
	if maskR.Pixel(1, 0) != 0 {
		t.Errorf("RGB24 R component: expected 0, got %d", maskR.Pixel(1, 0))
	}

	// Test G component mask
	maskG := NewAlphaMaskU8(3, 1, OneComponentMaskU8{})
	maskG.Attach(rbuf)

	if maskG.Pixel(0, 0) != 128 {
		t.Errorf("RGB24 G component: expected 128, got %d", maskG.Pixel(0, 0))
	}
	if maskG.Pixel(1, 0) != 100 {
		t.Errorf("RGB24 G component: expected 100, got %d", maskG.Pixel(1, 0))
	}

	// Test B component mask
	maskB := NewAlphaMaskU8(3, 2, OneComponentMaskU8{})
	maskB.Attach(rbuf)

	if maskB.Pixel(0, 0) != 64 {
		t.Errorf("RGB24 B component: expected 64, got %d", maskB.Pixel(0, 0))
	}
	if maskB.Pixel(1, 0) != 200 {
		t.Errorf("RGB24 B component: expected 200, got %d", maskB.Pixel(1, 0))
	}
}

// TestAlphaMaskU8WithRGBA32 tests RGBA32 component masks
func TestAlphaMaskU8WithRGBA32(t *testing.T) {
	// Create RGBA32 data: 1x1 image with (255,128,64,32)
	width, height := 1, 1
	rgbaData := []basics.Int8u{255, 128, 64, 32} // R=255, G=128, B=64, A=32

	rbuf := buffer.NewRenderingBufferU8WithData(rgbaData, width, height, width*4)

	// Test A component mask
	maskA := NewAlphaMaskU8(4, 3, OneComponentMaskU8{})
	maskA.Attach(rbuf)

	if maskA.Pixel(0, 0) != 32 {
		t.Errorf("RGBA32 A component: expected 32, got %d", maskA.Pixel(0, 0))
	}
}

// TestRGBToGrayAlphaMask tests RGB to grayscale conversion mask
func TestRGBToGrayAlphaMask(t *testing.T) {
	// Create RGB24 data with pure colors
	width, height := 3, 1
	rgbData := []basics.Int8u{
		255, 0, 0, // Pure red
		0, 255, 0, // Pure green
		0, 0, 255, // Pure blue
	}

	rbuf := buffer.NewRenderingBufferU8WithData(rgbData, width, height, width*3)
	mask := NewAlphaMaskU8(3, 0, NewRGBToGrayMaskU8(0, 1, 2))
	mask.Attach(rbuf)

	// Test red conversion
	redGray := mask.Pixel(0, 0)
	expectedRed := basics.Int8u((255 * 77) >> 8) // Around 75
	if redGray != expectedRed {
		t.Errorf("Red to gray: expected %d, got %d", expectedRed, redGray)
	}

	// Test green conversion
	greenGray := mask.Pixel(1, 0)
	expectedGreen := basics.Int8u((255 * 150) >> 8) // Around 147
	if greenGray != expectedGreen {
		t.Errorf("Green to gray: expected %d, got %d", expectedGreen, greenGray)
	}

	// Test blue conversion
	blueGray := mask.Pixel(2, 0)
	expectedBlue := basics.Int8u((255 * 29) >> 8) // Around 28
	if blueGray != expectedBlue {
		t.Errorf("Blue to gray: expected %d, got %d", expectedBlue, blueGray)
	}
}

// TestPredefinedConstructors tests the predefined constructor functions
func TestPredefinedConstructors(t *testing.T) {
	// Test grayscale constructor
	maskGray := NewAlphaMaskGray8()
	if maskGray == nil {
		t.Error("NewAlphaMaskGray8 returned nil")
	}

	// Test no-clip grayscale constructor
	maskNoClipGray := NewAMaskNoClipGray8()
	if maskNoClipGray == nil {
		t.Error("NewAMaskNoClipGray8 returned nil")
	}

	// Test RGB24 to gray constructor
	maskRGBGray := NewAlphaMaskRGB24Gray()
	if maskRGBGray == nil {
		t.Error("NewAlphaMaskRGB24Gray returned nil")
	}

	// Test RGB24 component constructors
	maskR := NewAlphaMaskRGB24R()
	if maskR == nil {
		t.Error("NewAlphaMaskRGB24R returned nil")
	}

	maskG := NewAlphaMaskRGB24G()
	if maskG == nil {
		t.Error("NewAlphaMaskRGB24G returned nil")
	}

	maskB := NewAlphaMaskRGB24B()
	if maskB == nil {
		t.Error("NewAlphaMaskRGB24B returned nil")
	}
}

// TestNilRenderingBuffer tests behavior with nil rendering buffer
func TestNilRenderingBuffer(t *testing.T) {
	mask := NewAlphaMaskU8(1, 0, OneComponentMaskU8{})
	// Don't attach any buffer

	// All operations should return 0 or handle gracefully
	if mask.Pixel(0, 0) != 0 {
		t.Error("Pixel should return 0 with nil buffer")
	}

	if mask.CombinePixel(0, 0, 100) != 0 {
		t.Error("CombinePixel should return 0 with nil buffer")
	}

	dst := make([]basics.Int8u, 3)
	mask.FillHspan(0, 0, dst, 3)
	// Should not crash, dst should remain unchanged

	mask.CombineHspan(0, 0, dst, 3)
	// Should not crash

	mask.FillVspan(0, 0, dst, 3)
	// Should not crash

	mask.CombineVspan(0, 0, dst, 3)
	// Should not crash
}

// TestInvalidSpanParameters tests span operations with invalid parameters
func TestInvalidSpanParameters(t *testing.T) {
	width, height := 2, 2
	maskData := []basics.Int8u{100, 150, 200, 250}

	rbuf := buffer.NewRenderingBufferU8WithData(maskData, width, height, width)
	mask := NewAlphaMaskU8(1, 0, OneComponentMaskU8{})
	mask.Attach(rbuf)

	// Test with numPix <= 0
	dst := make([]basics.Int8u, 3)
	originalDst := make([]basics.Int8u, 3)
	copy(originalDst, dst)

	mask.FillHspan(0, 0, dst, 0)
	// dst should be unchanged
	for i := range dst {
		if dst[i] != originalDst[i] {
			t.Errorf("FillHspan with numPix=0 should not modify dst[%d]", i)
		}
	}

	mask.FillHspan(0, 0, dst, -1)
	// dst should be unchanged
	for i := range dst {
		if dst[i] != originalDst[i] {
			t.Errorf("FillHspan with numPix=-1 should not modify dst[%d]", i)
		}
	}

	// Test with dst slice too small
	smallDst := make([]basics.Int8u, 1)
	mask.FillHspan(0, 0, smallDst, 3) // numPix > len(dst)
	// Should not crash
}

// BenchmarkAlphaMaskU8Pixel benchmarks the Pixel method
func BenchmarkAlphaMaskU8Pixel(b *testing.B) {
	width, height := 100, 100
	maskData := make([]basics.Int8u, width*height)
	for i := range maskData {
		maskData[i] = basics.Int8u(i % 256)
	}

	rbuf := buffer.NewRenderingBufferU8WithData(maskData, width, height, width)
	mask := NewAlphaMaskU8(1, 0, OneComponentMaskU8{})
	mask.Attach(rbuf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mask.Pixel(i%width, (i/width)%height)
	}
}

// BenchmarkAlphaMaskU8FillHspan benchmarks the FillHspan method
func BenchmarkAlphaMaskU8FillHspan(b *testing.B) {
	width, height := 1000, 100
	maskData := make([]basics.Int8u, width*height)
	for i := range maskData {
		maskData[i] = basics.Int8u(i % 256)
	}

	rbuf := buffer.NewRenderingBufferU8WithData(maskData, width, height, width)
	mask := NewAlphaMaskU8(1, 0, OneComponentMaskU8{})
	mask.Attach(rbuf)

	dst := make([]basics.Int8u, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mask.FillHspan(0, i%height, dst, len(dst))
	}
}

// BenchmarkAMaskNoClipU8Pixel benchmarks the no-clip variant
func BenchmarkAMaskNoClipU8Pixel(b *testing.B) {
	width, height := 100, 100
	maskData := make([]basics.Int8u, width*height)
	for i := range maskData {
		maskData[i] = basics.Int8u(i % 256)
	}

	rbuf := buffer.NewRenderingBufferU8WithData(maskData, width, height, width)
	mask := NewAMaskNoClipU8(1, 0, OneComponentMaskU8{})
	mask.Attach(rbuf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mask.Pixel(i%width, (i/width)%height)
	}
}
