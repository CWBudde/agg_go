package pixfmt

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
)

// MockPixFmt implements a simple pixel format for testing
type MockPixFmt struct {
	width, height int
	pixels        [][]color.RGBA8[color.Linear]
}

func NewMockPixFmt(width, height int) *MockPixFmt {
	pixels := make([][]color.RGBA8[color.Linear], height)
	for y := 0; y < height; y++ {
		pixels[y] = make([]color.RGBA8[color.Linear], width)
	}
	return &MockPixFmt{
		width:  width,
		height: height,
		pixels: pixels,
	}
}

func (m *MockPixFmt) Width() int  { return m.width }
func (m *MockPixFmt) Height() int { return m.height }

func (m *MockPixFmt) GetPixel(x, y int) color.RGBA8[color.Linear] {
	if x >= 0 && x < m.width && y >= 0 && y < m.height {
		return m.pixels[y][x]
	}
	return color.RGBA8[color.Linear]{}
}

func (m *MockPixFmt) CopyPixel(x, y int, c color.RGBA8[color.Linear]) {
	if x >= 0 && x < m.width && y >= 0 && y < m.height {
		m.pixels[y][x] = c
	}
}

func (m *MockPixFmt) BlendPixel(x, y int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	if x >= 0 && x < m.width && y >= 0 && y < m.height {
		if cover == basics.CoverFull {
			m.pixels[y][x] = c
		} else {
			// Simple blending for test purposes
			existing := m.pixels[y][x]
			existing.AddWithCover(c, cover)
			m.pixels[y][x] = existing
		}
	}
}

func (m *MockPixFmt) CopyHline(x, y, length int, c color.RGBA8[color.Linear]) {
	for i := 0; i < length; i++ {
		m.CopyPixel(x+i, y, c)
	}
}

func (m *MockPixFmt) CopyVline(x, y, length int, c color.RGBA8[color.Linear]) {
	for i := 0; i < length; i++ {
		m.CopyPixel(x, y+i, c)
	}
}

func (m *MockPixFmt) BlendHline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	for i := 0; i < length; i++ {
		m.BlendPixel(x+i, y, c, cover)
	}
}

func (m *MockPixFmt) BlendVline(x, y, length int, c color.RGBA8[color.Linear], cover basics.Int8u) {
	for i := 0; i < length; i++ {
		m.BlendPixel(x, y+i, c, cover)
	}
}

func (m *MockPixFmt) BlendSolidHspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u) {
	for i := 0; i < length; i++ {
		cover := basics.Int8u(basics.CoverFull)
		if covers != nil && i < len(covers) {
			cover = covers[i]
		}
		m.BlendPixel(x+i, y, c, cover)
	}
}

func (m *MockPixFmt) BlendSolidVspan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.Int8u) {
	for i := 0; i < length; i++ {
		cover := basics.Int8u(basics.CoverFull)
		if covers != nil && i < len(covers) {
			cover = covers[i]
		}
		m.BlendPixel(x, y+i, c, cover)
	}
}

// MockRenderBuffer implements a rendering buffer interface for testing CopyFrom
type MockRenderBuffer struct {
	width, height int
	data          []basics.Int8u
}

type BlendSrcWrapper struct {
	pixels [][]color.RGBA8[color.Linear]
	w      int
	h      int
}

func (b *BlendSrcWrapper) GetPixel(x, y int) color.RGBA8[color.Linear] {
	if y < 0 || y >= b.h || x < 0 || x >= b.w {
		return color.RGBA8[color.Linear]{}
	}
	return b.pixels[y][x]
}

func (b *BlendSrcWrapper) Width() int  { return b.w }
func (b *BlendSrcWrapper) Height() int { return b.h }

func NewMockRenderBuffer(width, height int) *MockRenderBuffer {
	return &MockRenderBuffer{
		width:  width,
		height: height,
		data:   make([]basics.Int8u, width*height*4), // 4 bytes per pixel (RGBA)
	}
}

func (m *MockRenderBuffer) Width() int  { return m.width }
func (m *MockRenderBuffer) Height() int { return m.height }

func (m *MockRenderBuffer) RowData(y int) []basics.Int8u {
	if y >= 0 && y < m.height {
		start := y * m.width * 4
		end := start + m.width*4
		return m.data[start:end]
	}
	return nil
}

func (m *MockRenderBuffer) SetPixel(x, y int, c color.RGBA8[color.Linear]) {
	if x >= 0 && x < m.width && y >= 0 && y < m.height {
		offset := (y*m.width + x) * 4
		m.data[offset] = basics.Int8u(c.R)
		m.data[offset+1] = basics.Int8u(c.G)
		m.data[offset+2] = basics.Int8u(c.B)
		m.data[offset+3] = basics.Int8u(c.A)
	}
}

func TestPixFmtTransposer_BasicFunctionality(t *testing.T) {
	// Create a 3x2 mock pixel format
	mock := NewMockPixFmt(3, 2)
	transposer := NewPixFmtTransposer(mock)

	// Test dimensions are transposed
	if transposer.Width() != 2 {
		t.Errorf("Expected transposed width 2, got %d", transposer.Width())
	}
	if transposer.Height() != 3 {
		t.Errorf("Expected transposed height 3, got %d", transposer.Height())
	}

	// Test pixel operations with transposed coordinates
	testColor := color.RGBA8[color.Linear]{R: 255, G: 128, B: 64, A: 255}

	// Set a pixel in transposed coordinates (1, 2) should go to (2, 1) in original
	transposer.CopyPixel(1, 2, testColor)

	// Check that it was stored in the correct location in the original format
	stored := mock.GetPixel(2, 1)
	if stored != testColor {
		t.Errorf("Expected stored pixel %+v, got %+v", testColor, stored)
	}

	// Check that we can read it back through the transposer
	retrieved := transposer.GetPixel(1, 2)
	if retrieved != testColor {
		t.Errorf("Expected retrieved pixel %+v, got %+v", testColor, retrieved)
	}
}

func TestPixFmtTransposer_LineOperations(t *testing.T) {
	mock := NewMockPixFmt(4, 3)
	transposer := NewPixFmtTransposer(mock)

	testColor := color.RGBA8[color.Linear]{R: 200, G: 100, B: 50, A: 255}

	// Test horizontal line becomes vertical line in original
	transposer.CopyHline(1, 0, 2, testColor)

	// This should have created a vertical line from (0, 1) to (0, 2) in original
	if mock.GetPixel(0, 1) != testColor {
		t.Errorf("Expected pixel at (0,1) in original to be %+v, got %+v", testColor, mock.GetPixel(0, 1))
	}
	if mock.GetPixel(0, 2) != testColor {
		t.Errorf("Expected pixel at (0,2) in original to be %+v, got %+v", testColor, mock.GetPixel(0, 2))
	}

	// Test vertical line becomes horizontal line in original
	testColor2 := color.RGBA8[color.Linear]{R: 50, G: 150, B: 200, A: 255}
	transposer.CopyVline(0, 1, 2, testColor2)

	// This should have created a horizontal line from (1, 0) to (2, 0) in original
	if mock.GetPixel(1, 0) != testColor2 {
		t.Errorf("Expected pixel at (1,0) in original to be %+v, got %+v", testColor2, mock.GetPixel(1, 0))
	}
	if mock.GetPixel(2, 0) != testColor2 {
		t.Errorf("Expected pixel at (2,0) in original to be %+v, got %+v", testColor2, mock.GetPixel(2, 0))
	}
}

func TestPixFmtTransposer_BlendOperations(t *testing.T) {
	mock := NewMockPixFmt(3, 3)
	transposer := NewPixFmtTransposer(mock)

	// Set up initial color
	initialColor := color.RGBA8[color.Linear]{R: 100, G: 100, B: 100, A: 128}
	mock.CopyPixel(1, 2, initialColor)

	// Blend a color at transposed coordinates
	blendColor := color.RGBA8[color.Linear]{R: 200, G: 0, B: 0, A: 255}
	transposer.BlendPixel(2, 1, blendColor, basics.Int8u(128)) // 50% blend

	// Check that the pixel was blended correctly in the original coordinates
	result := mock.GetPixel(1, 2)
	// The result should be different from the initial color (some blending should have occurred)
	if result.R == initialColor.R && result.G == initialColor.G && result.B == initialColor.B {
		t.Errorf("Expected blending to occur, but pixel unchanged: %+v", result)
	}
}

func TestPixFmtTransposer_SpanOperations(t *testing.T) {
	mock := NewMockPixFmt(4, 3)
	transposer := NewPixFmtTransposer(mock)

	testColor := color.RGBA8[color.Linear]{R: 255, G: 255, B: 0, A: 255}
	covers := []basics.Int8u{255, 128, 64}

	// Test horizontal span becomes vertical span in original
	transposer.BlendSolidHspan(1, 0, 3, testColor, covers)

	// This should have created a vertical span from (0, 1) to (0, 3) in original
	// with varying coverage
	pixel1 := mock.GetPixel(0, 1)
	pixel2 := mock.GetPixel(0, 2)
	pixel3 := mock.GetPixel(0, 3)

	// The pixels should have different intensities based on coverage
	if pixel1 == pixel2 || pixel2 == pixel3 {
		t.Errorf("Expected different coverage levels, but got same colors")
	}
}

func TestPixFmtTransposer_CopyFrom_Fallback(t *testing.T) {
	mock := NewMockPixFmt(3, 3)
	transposer := NewPixFmtTransposer(mock)

	// Create source rendering buffer
	src := NewMockRenderBuffer(3, 2)

	// Set some test data in the source
	testColor1 := color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 255}
	testColor2 := color.RGBA8[color.Linear]{R: 0, G: 255, B: 0, A: 255}

	src.SetPixel(0, 0, testColor1)
	src.SetPixel(1, 0, testColor2)

	// Copy from source to transposer (should use fallback since MockPixFmt doesn't implement CopyFrom)
	transposer.CopyFrom(src, 0, 0, 0, 0, 2)

	// Check that pixels were copied correctly
	copied1 := transposer.GetPixel(0, 0)
	copied2 := transposer.GetPixel(1, 0)

	if copied1 != testColor1 {
		t.Errorf("Expected copied pixel 1 to be %+v, got %+v", testColor1, copied1)
	}
	if copied2 != testColor2 {
		t.Errorf("Expected copied pixel 2 to be %+v, got %+v", testColor2, copied2)
	}
}

func TestPixFmtTransposer_CopyFrom_BoundsChecking(t *testing.T) {
	mock := NewMockPixFmt(2, 2)
	transposer := NewPixFmtTransposer(mock)
	src := NewMockRenderBuffer(2, 2)

	// Test copying beyond source bounds (should not crash)
	transposer.CopyFrom(src, 0, 0, 1, 0, 3) // length 3 but source width is 2

	// Test copying to negative coordinates (should not crash)
	transposer.CopyFrom(src, -1, 0, 0, 0, 2)

	// Test copying from negative source coordinates (should not crash)
	transposer.CopyFrom(src, 0, 0, -1, 0, 2)

	// These should complete without panicking
}

func TestPixFmtTransposer_BlendFrom_Fallback(t *testing.T) {
	mock := NewMockPixFmt(3, 3)
	transposer := NewPixFmtTransposer(mock)

	src := &BlendSrcWrapper{
		w: 3,
		h: 2,
		pixels: [][]color.RGBA8[color.Linear]{
			{
				{R: 10, G: 20, B: 30, A: 255},
				{R: 40, G: 50, B: 60, A: 255},
				{R: 70, G: 80, B: 90, A: 255},
			},
			{
				{R: 15, G: 25, B: 35, A: 255},
				{R: 45, G: 55, B: 65, A: 255},
				{R: 75, G: 85, B: 95, A: 255},
			},
		},
	}

	transposer.BlendFrom(src, 0, 1, 1, 0, 2, basics.CoverFull)

	if got := mock.GetPixel(1, 0); got != src.pixels[0][1] {
		t.Fatalf("expected underlying pixel (1,0) to be %+v, got %+v", src.pixels[0][1], got)
	}
	if got := mock.GetPixel(1, 1); got != src.pixels[0][2] {
		t.Fatalf("expected underlying pixel (1,1) to be %+v, got %+v", src.pixels[0][2], got)
	}
}

func TestPixFmtTransposer_Attach(t *testing.T) {
	transposer := &PixFmtTransposer{}

	mock1 := NewMockPixFmt(2, 3)
	mock2 := NewMockPixFmt(4, 5)

	// Test attaching first format
	transposer.Attach(mock1)
	if transposer.Width() != 3 || transposer.Height() != 2 {
		t.Errorf("Expected dimensions (3,2) after attaching mock1, got (%d,%d)",
			transposer.Width(), transposer.Height())
	}

	// Test attaching second format
	transposer.Attach(mock2)
	if transposer.Width() != 5 || transposer.Height() != 4 {
		t.Errorf("Expected dimensions (5,4) after attaching mock2, got (%d,%d)",
			transposer.Width(), transposer.Height())
	}
}

func TestPixFmtTransposer_CoordinateTransposition(t *testing.T) {
	// Test the utility functions
	x, y, w, h := TransposeCoords(10, 20, 100, 200)
	if x != 20 || y != 10 || w != 200 || h != 100 {
		t.Errorf("TransposeCoords(10,20,100,200) = (%d,%d,%d,%d), expected (20,10,200,100)", x, y, w, h)
	}

	// Test IsTransposed
	if !IsTransposed(100, 200, 200, 100) {
		t.Error("IsTransposed should return true for properly transposed dimensions")
	}

	if IsTransposed(100, 200, 100, 200) {
		t.Error("IsTransposed should return false for same dimensions")
	}
}

func TestPixFmtTransposer_ColorSpanOperations(t *testing.T) {
	mock := NewMockPixFmt(4, 3)
	transposer := NewPixFmtTransposer(mock)

	// Test BlendColorHspan fallback (since MockPixFmt doesn't have BlendColorVspan)
	colors := []color.RGBA8[color.Linear]{
		{R: 255, G: 0, B: 0, A: 255},
		{R: 0, G: 255, B: 0, A: 255},
		{R: 0, G: 0, B: 255, A: 255},
	}
	covers := []basics.Int8u{255, 128, 64}

	transposer.BlendColorHspan(0, 1, 3, colors, covers, basics.CoverFull)

	// Verify the colors were blended correctly with transposed coordinates
	// The horizontal span in the transposer should be a vertical span in the original
	result1 := mock.GetPixel(1, 0)
	result2 := mock.GetPixel(1, 1)
	result3 := mock.GetPixel(1, 2)

	// Check that different colors were applied (exact blending depends on implementation)
	if result1 == result2 || result2 == result3 {
		t.Error("Expected different colors from BlendColorHspan")
	}
}
