package renderer

import (
	"testing"

	"agg_go/internal/basics"
)

// MockPixelFormat implements PixelFormatInterface for testing
type MockPixelFormat struct {
	width, height int
	pixels        map[[2]int]interface{}
}

func NewMockPixelFormat(width, height int) *MockPixelFormat {
	return &MockPixelFormat{
		width:  width,
		height: height,
		pixels: make(map[[2]int]interface{}),
	}
}

func (m *MockPixelFormat) Width() int    { return m.width }
func (m *MockPixelFormat) Height() int   { return m.height }
func (m *MockPixelFormat) PixWidth() int { return 4 }

func (m *MockPixelFormat) CopyPixel(x, y int, c interface{}) {
	m.pixels[[2]int{x, y}] = c
}

func (m *MockPixelFormat) BlendPixel(x, y int, c interface{}, cover basics.Int8u) {
	m.pixels[[2]int{x, y}] = c
}

func (m *MockPixelFormat) Pixel(x, y int) interface{} {
	if pixel, exists := m.pixels[[2]int{x, y}]; exists {
		return pixel
	}
	return nil
}

func (m *MockPixelFormat) CopyHline(x, y, length int, c interface{}) {
	for i := 0; i < length; i++ {
		m.pixels[[2]int{x + i, y}] = c
	}
}

func (m *MockPixelFormat) BlendHline(x, y, length int, c interface{}, cover basics.Int8u) {
	for i := 0; i < length; i++ {
		m.pixels[[2]int{x + i, y}] = c
	}
}

func (m *MockPixelFormat) CopyVline(x, y, length int, c interface{}) {
	for i := 0; i < length; i++ {
		m.pixels[[2]int{x, y + i}] = c
	}
}

func (m *MockPixelFormat) BlendVline(x, y, length int, c interface{}, cover basics.Int8u) {
	for i := 0; i < length; i++ {
		m.pixels[[2]int{x, y + i}] = c
	}
}

func (m *MockPixelFormat) BlendSolidHspan(x, y, length int, c interface{}, covers []basics.Int8u) {
	for i := 0; i < length; i++ {
		m.pixels[[2]int{x + i, y}] = c
	}
}

func (m *MockPixelFormat) BlendSolidVspan(x, y, length int, c interface{}, covers []basics.Int8u) {
	for i := 0; i < length; i++ {
		m.pixels[[2]int{x, y + i}] = c
	}
}

func (m *MockPixelFormat) CopyColorHspan(x, y, length int, colors []interface{}) {
	for i := 0; i < length && i < len(colors); i++ {
		m.pixels[[2]int{x + i, y}] = colors[i]
	}
}

func (m *MockPixelFormat) BlendColorHspan(x, y, length int, colors []interface{}, covers []basics.Int8u, cover basics.Int8u) {
	for i := 0; i < length && i < len(colors); i++ {
		m.pixels[[2]int{x + i, y}] = colors[i]
	}
}

func (m *MockPixelFormat) CopyColorVspan(x, y, length int, colors []interface{}) {
	for i := 0; i < length && i < len(colors); i++ {
		m.pixels[[2]int{x, y + i}] = colors[i]
	}
}

func (m *MockPixelFormat) BlendColorVspan(x, y, length int, colors []interface{}, covers []basics.Int8u, cover basics.Int8u) {
	for i := 0; i < length && i < len(colors); i++ {
		m.pixels[[2]int{x, y + i}] = colors[i]
	}
}

func (m *MockPixelFormat) CopyFrom(src interface{}, dstX, dstY, srcX, srcY, length int) {
	// Mock implementation
}

// MockColorType implements ColorTypeInterface for testing
type MockColorType struct{}

func (m MockColorType) NoColor() interface{} {
	return nil
}

func TestRendererBaseConstruction(t *testing.T) {
	// Test default constructor
	renderer := NewRendererBase[*MockPixelFormat, MockColorType]()
	if renderer == nil {
		t.Fatal("NewRendererBase returned nil")
	}

	// Test constructor with pixel format
	pixfmt := NewMockPixelFormat(100, 100)
	renderer2 := NewRendererBaseWithPixfmt[*MockPixelFormat, MockColorType](pixfmt)
	if renderer2 == nil {
		t.Fatal("NewRendererBaseWithPixfmt returned nil")
	}

	if renderer2.Width() != 100 || renderer2.Height() != 100 {
		t.Errorf("Expected 100x100, got %dx%d", renderer2.Width(), renderer2.Height())
	}
}

func TestRendererBaseClipping(t *testing.T) {
	pixfmt := NewMockPixelFormat(100, 100)
	renderer := NewRendererBaseWithPixfmt[*MockPixelFormat, MockColorType](pixfmt)

	// Test clipping box
	if !renderer.ClipBox(10, 10, 50, 50) {
		t.Error("ClipBox should return true for valid box")
	}

	if renderer.Xmin() != 10 || renderer.Ymin() != 10 ||
		renderer.Xmax() != 50 || renderer.Ymax() != 50 {
		t.Errorf("Expected clipping box (10,10,50,50), got (%d,%d,%d,%d)",
			renderer.Xmin(), renderer.Ymin(), renderer.Xmax(), renderer.Ymax())
	}

	// Test point in box
	if !renderer.InBox(25, 25) {
		t.Error("Point (25,25) should be in clipping box")
	}

	if renderer.InBox(5, 5) {
		t.Error("Point (5,5) should not be in clipping box")
	}

	// Test reset clipping
	renderer.ResetClipping(true)
	if renderer.Xmin() != 0 || renderer.Ymin() != 0 ||
		renderer.Xmax() != 99 || renderer.Ymax() != 99 {
		t.Error("ResetClipping(true) should set to full buffer bounds")
	}

	renderer.ResetClipping(false)
	if renderer.InBox(50, 50) {
		t.Error("After ResetClipping(false), no points should be in box")
	}
}

func TestRendererBasePixelOperations(t *testing.T) {
	pixfmt := NewMockPixelFormat(100, 100)
	renderer := NewRendererBaseWithPixfmt[*MockPixelFormat, MockColorType](pixfmt)

	// Test pixel operations with clipping
	renderer.ClipBox(10, 10, 50, 50)

	// This should work (inside clipping box)
	renderer.CopyPixel(25, 25, "red")
	pixel := renderer.Pixel(25, 25)
	if pixel != "red" {
		t.Errorf("Expected 'red', got %v", pixel)
	}

	// This should be ignored (outside clipping box)
	renderer.CopyPixel(5, 5, "blue")
	pixel = renderer.Pixel(5, 5)
	if pixel != nil {
		t.Errorf("Expected nil (NoColor), got %v", pixel)
	}
}

func TestRendererBaseLineOperations(t *testing.T) {
	pixfmt := NewMockPixelFormat(100, 100)
	renderer := NewRendererBaseWithPixfmt[*MockPixelFormat, MockColorType](pixfmt)

	// Test horizontal line
	renderer.CopyHline(10, 20, 30, "green")

	// Check a few pixels along the line
	if renderer.Pixel(15, 20) != "green" {
		t.Error("Horizontal line pixel should be 'green'")
	}
	if renderer.Pixel(25, 20) != "green" {
		t.Error("Horizontal line pixel should be 'green'")
	}

	// Test vertical line
	renderer.CopyVline(40, 10, 30, "blue")

	// Check a few pixels along the line
	if renderer.Pixel(40, 15) != "blue" {
		t.Error("Vertical line pixel should be 'blue'")
	}
	if renderer.Pixel(40, 25) != "blue" {
		t.Error("Vertical line pixel should be 'blue'")
	}
}

func TestRendererBaseRectangleOperations(t *testing.T) {
	pixfmt := NewMockPixelFormat(100, 100)
	renderer := NewRendererBaseWithPixfmt[*MockPixelFormat, MockColorType](pixfmt)

	// Test rectangle
	renderer.CopyBar(10, 10, 20, 20, "yellow")

	// Check corners and center
	if renderer.Pixel(10, 10) != "yellow" {
		t.Error("Rectangle corner should be 'yellow'")
	}
	if renderer.Pixel(15, 15) != "yellow" {
		t.Error("Rectangle center should be 'yellow'")
	}
	if renderer.Pixel(20, 20) != "yellow" {
		t.Error("Rectangle corner should be 'yellow'")
	}
}
