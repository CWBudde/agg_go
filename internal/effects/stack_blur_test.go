package effects

import (
	"testing"
	"unsafe"

	"agg_go/internal/basics"
	"agg_go/internal/color"
)

// ──────────────────────────────────────────────────────────────────────────
// SimpleStackBlur.Blur (full 2D path)
// ──────────────────────────────────────────────────────────────────────────

func TestSimpleStackBlurBoth(t *testing.T) {
	pixels := make([][]color.RGBA8[color.Linear], 7)
	for i := range pixels {
		pixels[i] = make([]color.RGBA8[color.Linear], 7)
	}
	white := color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255}
	pixels[3][3] = white

	sb := NewSimpleStackBlur()
	sb.Blur(pixels, 1)

	// After full blur the white dot should have spread in both axes
	center := pixels[3][3]
	if center.R == 255 && center.G == 255 && center.B == 255 {
		// centre may still be brightest, but neighbours should be non-zero
	}
	neighbor := pixels[3][4]
	if neighbor.R == 0 && neighbor.G == 0 && neighbor.B == 0 {
		t.Error("horizontal neighbour should have blur contribution")
	}
	neighbor2 := pixels[4][3]
	if neighbor2.R == 0 && neighbor2.G == 0 && neighbor2.B == 0 {
		t.Error("vertical neighbour should have blur contribution")
	}
}

// ──────────────────────────────────────────────────────────────────────────
// StackBlurCalcRGBA.CalcPix
// ──────────────────────────────────────────────────────────────────────────

func TestCalcPix(t *testing.T) {
	calc := StackBlurCalcRGBA[uint32]{r: 255, g: 128, b: 64, a: 255}
	var result color.RGBA8[color.Linear]
	calc.CalcPix(&result, 1)
	if result.R != 255 || result.G != 128 || result.B != 64 || result.A != 255 {
		t.Errorf("CalcPix/1 = %+v, want {255,128,64,255}", result)
	}

	calc2 := StackBlurCalcRGBA[uint32]{r: 200, g: 100, b: 50, a: 200}
	var result2 color.RGBA8[color.Linear]
	calc2.CalcPix(&result2, 2)
	if result2.R != 100 || result2.G != 50 || result2.B != 25 || result2.A != 100 {
		t.Errorf("CalcPix/2 = %+v, want {100,50,25,100}", result2)
	}
}

// ──────────────────────────────────────────────────────────────────────────
// Mock images for StackBlurGray8 / StackBlurRGB24 / StackBlurRGBA32
// ──────────────────────────────────────────────────────────────────────────

type mockGrayImage struct {
	w, h int
	data []basics.Int8u
}

func newMockGrayImage(w, h int) *mockGrayImage {
	return &mockGrayImage{w: w, h: h, data: make([]basics.Int8u, w*h)}
}
func (m *mockGrayImage) Width() int  { return m.w }
func (m *mockGrayImage) Height() int { return m.h }
func (m *mockGrayImage) Stride() int { return m.w }
func (m *mockGrayImage) PixPtr(x, y int) *basics.Int8u {
	return &m.data[y*m.w+x]
}

func (m *mockGrayImage) NextPixPtr(ptr *basics.Int8u) *basics.Int8u {
	base := uintptr(unsafe.Pointer(&m.data[0]))
	cur := uintptr(unsafe.Pointer(ptr))
	idx := int(cur-base) + 1
	if idx >= len(m.data) {
		return &m.data[len(m.data)-1]
	}
	return &m.data[idx]
}

func (m *mockGrayImage) PixPtrOffset(ptr *basics.Int8u, offset int) *basics.Int8u {
	base := uintptr(unsafe.Pointer(&m.data[0]))
	cur := uintptr(unsafe.Pointer(ptr))
	idx := int(cur-base) + offset
	if idx < 0 {
		idx = 0
	}
	if idx >= len(m.data) {
		idx = len(m.data) - 1
	}
	return &m.data[idx]
}

func TestStackBlurGray8(t *testing.T) {
	img := newMockGrayImage(5, 5)
	// Set a bright centre pixel
	img.data[2*5+2] = 255

	StackBlurGray8(img, 1, 1)

	// Centre should still have some brightness
	if img.data[2*5+2] == 0 {
		t.Error("centre pixel should not be zero after blur")
	}
	// Adjacent pixel should have received some brightness
	if img.data[2*5+3] == 0 {
		t.Error("adjacent pixel should have blur contribution")
	}
}

func TestStackBlurGray8NoRadius(t *testing.T) {
	img := newMockGrayImage(4, 4)
	img.data[1*4+1] = 200
	// radius 0 means no horizontal or vertical pass - must not panic
	StackBlurGray8(img, 0, 0)
	if img.data[1*4+1] != 200 {
		t.Error("zero radius should leave image unchanged")
	}
}

// mockRGBPixel is a *[3]byte-style pointer used by RGBImageInterface
type mockRGBImage struct {
	w, h int
	data [][3]byte
}

func newMockRGBImage(w, h int) *mockRGBImage {
	return &mockRGBImage{w: w, h: h, data: make([][3]byte, w*h)}
}
func (m *mockRGBImage) Width() int  { return m.w }
func (m *mockRGBImage) Height() int { return m.h }
func (m *mockRGBImage) PixPtr(x, y int) *[3]byte {
	return &m.data[y*m.w+x]
}

func (m *mockRGBImage) NextPixPtr(ptr *[3]byte) *[3]byte {
	base := uintptr(unsafe.Pointer(&m.data[0]))
	cur := uintptr(unsafe.Pointer(ptr))
	idx := int((cur - base) / 3)
	if idx+1 >= len(m.data) {
		return &m.data[len(m.data)-1]
	}
	return &m.data[idx+1]
}

func (m *mockRGBImage) GetRGB(ptr *[3]byte) color.RGB8[color.Linear] {
	return color.RGB8[color.Linear]{R: ptr[0], G: ptr[1], B: ptr[2]}
}

func (m *mockRGBImage) SetRGB(ptr *[3]byte, rgb color.RGB8[color.Linear]) {
	ptr[0] = rgb.R
	ptr[1] = rgb.G
	ptr[2] = rgb.B
}

func TestStackBlurRGB24(t *testing.T) {
	img := newMockRGBImage(5, 5)
	// Bright centre pixel
	img.data[2*5+2] = [3]byte{255, 0, 0}

	StackBlurRGB24(img, 1, 1)

	// Centre should still be reddish
	if img.data[2*5+2][0] == 0 {
		t.Error("centre red channel should not be zero after blur")
	}
}

func TestStackBlurRGB24NoRadius(t *testing.T) {
	img := newMockRGBImage(4, 4)
	img.data[0] = [3]byte{100, 100, 100}
	StackBlurRGB24(img, 0, 0)
	// No modification expected
}

// mockRGBAImage implements RGBAImageInterface[*[4]byte]
type mockRGBAImage struct {
	w, h int
	data [][4]byte
}

func newMockRGBAImage(w, h int) *mockRGBAImage {
	return &mockRGBAImage{w: w, h: h, data: make([][4]byte, w*h)}
}
func (m *mockRGBAImage) Width() int  { return m.w }
func (m *mockRGBAImage) Height() int { return m.h }
func (m *mockRGBAImage) PixPtr(x, y int) *[4]byte {
	return &m.data[y*m.w+x]
}

func (m *mockRGBAImage) NextPixPtr(ptr *[4]byte) *[4]byte {
	base := uintptr(unsafe.Pointer(&m.data[0]))
	cur := uintptr(unsafe.Pointer(ptr))
	idx := int((cur - base) / 4)
	if idx+1 >= len(m.data) {
		return &m.data[len(m.data)-1]
	}
	return &m.data[idx+1]
}

func (m *mockRGBAImage) GetRGBA(ptr *[4]byte) color.RGBA8[color.Linear] {
	return color.RGBA8[color.Linear]{R: ptr[0], G: ptr[1], B: ptr[2], A: ptr[3]}
}

func (m *mockRGBAImage) SetRGBA(ptr *[4]byte, rgba color.RGBA8[color.Linear]) {
	ptr[0] = rgba.R
	ptr[1] = rgba.G
	ptr[2] = rgba.B
	ptr[3] = rgba.A
}

func TestStackBlurRGBA32(t *testing.T) {
	img := newMockRGBAImage(5, 5)
	img.data[2*5+2] = [4]byte{255, 128, 0, 255}

	StackBlurRGBA32(img, 1, 1)

	// Centre should still have some red
	if img.data[2*5+2][0] == 0 {
		t.Error("centre red channel should not be zero after blur")
	}
}

func TestStackBlurRGBA32NoRadius(t *testing.T) {
	img := newMockRGBAImage(4, 4)
	img.data[0] = [4]byte{50, 50, 50, 255}
	StackBlurRGBA32(img, 0, 0)
	// No modification expected - must not panic
}
