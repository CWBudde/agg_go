package renderer

import (
	"testing"

	"agg_go/internal/basics"
)

// Use MockColorType from renderer_base_test.go

// TestNewRendererMClip tests the constructor
func TestNewRendererMClip(t *testing.T) {
	pixfmt := NewMockPixelFormat(100, 80)
	renderer := NewRendererMClip[*MockPixelFormat, MockColorType](pixfmt)

	if renderer == nil {
		t.Fatal("NewRendererMClip should not return nil")
	}

	if renderer.Width() != 100 {
		t.Errorf("Expected width 100, got %d", renderer.Width())
	}

	if renderer.Height() != 80 {
		t.Errorf("Expected height 80, got %d", renderer.Height())
	}

	// Check initial bounding box
	bounds := renderer.BoundingClipBox()
	expected := basics.RectI{X1: 0, Y1: 0, X2: 99, Y2: 79}
	if bounds != expected {
		t.Errorf("Expected initial bounds %+v, got %+v", expected, bounds)
	}
}

// TestClipBoxManagement tests clip box management methods
func TestClipBoxManagement(t *testing.T) {
	pixfmt := NewMockPixelFormat(100, 100)
	renderer := NewRendererMClip[*MockPixelFormat, MockColorType](pixfmt)

	// Initially, no clip boxes should be added
	renderer.FirstClipBox()

	// Add some clip boxes
	renderer.AddClipBox(10, 10, 30, 30)
	renderer.AddClipBox(50, 50, 80, 80)
	renderer.AddClipBox(20, 60, 40, 90)

	// Test FirstClipBox and NextClipBox iteration
	renderer.FirstClipBox()

	// Should be able to advance to next clip box
	if !renderer.NextClipBox() {
		t.Error("Expected to have a second clip box")
	}

	// Should be able to advance to third clip box
	if !renderer.NextClipBox() {
		t.Error("Expected to have a third clip box")
	}

	// Should not be able to advance further
	if renderer.NextClipBox() {
		t.Error("Should not have more than three clip boxes")
	}

	// Test bounding box calculation
	bounds := renderer.BoundingClipBox()
	expectedBounds := basics.RectI{X1: 10, Y1: 10, X2: 80, Y2: 90}
	if bounds != expectedBounds {
		t.Errorf("Expected bounds %+v, got %+v", expectedBounds, bounds)
	}

	// Test reset clipping
	renderer.ResetClipping(true)
	bounds = renderer.BoundingClipBox()
	expectedBounds = renderer.ClipBox() // Should be reset to full size
	if bounds != expectedBounds {
		t.Errorf("Expected bounds to reset to %+v, got %+v", expectedBounds, bounds)
	}
}

// TestClipBoxBoundaryConditions tests edge cases for clip box management
func TestClipBoxBoundaryConditions(t *testing.T) {
	pixfmt := NewMockPixelFormat(100, 100)
	renderer := NewRendererMClip[*MockPixelFormat, MockColorType](pixfmt)

	// Test adding clip box outside surface bounds (should be clipped)
	renderer.AddClipBox(90, 90, 200, 200)
	renderer.FirstClipBox()

	// The clip box should be clipped to surface bounds
	clipBox := renderer.ClipBox()
	if clipBox.X2 >= 100 || clipBox.Y2 >= 100 {
		t.Errorf("Clip box should be clipped to surface bounds, got %+v", clipBox)
	}

	// Test adding invalid clip box (should be normalized)
	renderer.ResetClipping(true)
	renderer.AddClipBox(30, 30, 10, 10) // x2 < x1, y2 < y1
	renderer.FirstClipBox()

	clipBox = renderer.ClipBox()
	if clipBox.X1 > clipBox.X2 || clipBox.Y1 > clipBox.Y2 {
		t.Errorf("Invalid clip box should be normalized, got %+v", clipBox)
	}

	// Test adding clip box completely outside surface (should be rejected)
	renderer.ResetClipping(true)
	renderer.AddClipBox(150, 150, 200, 200)
	renderer.FirstClipBox()
	// This should result in no valid clip boxes
}

// TestPixelOperations tests pixel-level rendering methods
func TestPixelOperations(t *testing.T) {
	pixfmt := NewMockPixelFormat(100, 100)
	renderer := NewRendererMClip[*MockPixelFormat, MockColorType](pixfmt)

	// Add two non-overlapping clip boxes
	renderer.AddClipBox(10, 10, 30, 30)
	renderer.AddClipBox(50, 50, 80, 80)

	testColor := "red"

	// Test CopyPixel in first clip box
	renderer.CopyPixel(20, 20, testColor)
	if pixfmt.Pixel(20, 20) != testColor {
		t.Errorf("Expected pixel at (20,20) to be %v, got %v", testColor, pixfmt.Pixel(20, 20))
	}

	// Test CopyPixel in second clip box
	renderer.CopyPixel(60, 60, testColor)
	if pixfmt.Pixel(60, 60) != testColor {
		t.Errorf("Expected pixel at (60,60) to be %v, got %v", testColor, pixfmt.Pixel(60, 60))
	}

	// Test CopyPixel outside all clip boxes (should not affect pixel format)
	renderer.CopyPixel(5, 5, testColor)
	if pixfmt.Pixel(5, 5) == testColor {
		t.Error("Pixel outside clip boxes should not be affected")
	}

	// Test BlendPixel
	renderer.BlendPixel(25, 25, "blue", 128)
	if pixfmt.Pixel(25, 25) != "blue" {
		t.Errorf("Expected pixel at (25,25) to be blue after blending, got %v", pixfmt.Pixel(25, 25))
	}

	// Test Pixel method
	pixel := renderer.Pixel(20, 20)
	if pixel != testColor {
		t.Errorf("Expected Pixel(20,20) to return %v, got %v", testColor, pixel)
	}

	// Test Pixel method outside clip boxes (should return NoColor)
	pixel = renderer.Pixel(5, 5)
	expected := MockColorType{}.NoColor()
	if pixel != expected {
		t.Errorf("Expected Pixel outside clip boxes to return %v, got %v", expected, pixel)
	}
}

// TestLineOperations tests line rendering methods
func TestLineOperations(t *testing.T) {
	pixfmt := NewMockPixelFormat(100, 100)
	renderer := NewRendererMClip[*MockPixelFormat, MockColorType](pixfmt)

	// Add a clip box
	renderer.AddClipBox(20, 20, 60, 60)

	testColor := "green"

	// Test CopyHline
	renderer.CopyHline(25, 30, 45, testColor)
	// Verify pixels in the line are set
	for x := 25; x <= 45; x++ {
		if pixfmt.Pixel(x, 30) != testColor {
			t.Errorf("Expected pixel at (%d,30) to be %v, got %v", x, testColor, pixfmt.Pixel(x, 30))
		}
	}

	// Test CopyVline
	renderer.CopyVline(35, 25, 45, testColor)
	// Verify pixels in the line are set
	for y := 25; y <= 45; y++ {
		if pixfmt.Pixel(35, y) != testColor {
			t.Errorf("Expected pixel at (35,%d) to be %v, got %v", y, testColor, pixfmt.Pixel(35, y))
		}
	}

	// Test BlendHline
	renderer.BlendHline(30, 35, 50, "purple", 200)
	for x := 30; x <= 50; x++ {
		if pixfmt.Pixel(x, 35) != "purple" {
			t.Errorf("Expected pixel at (%d,35) to be purple after blending, got %v", x, pixfmt.Pixel(x, 35))
		}
	}

	// Test BlendVline
	renderer.BlendVline(40, 30, 50, "orange", 150)
	for y := 30; y <= 50; y++ {
		if pixfmt.Pixel(40, y) != "orange" {
			t.Errorf("Expected pixel at (40,%d) to be orange after blending, got %v", y, pixfmt.Pixel(40, y))
		}
	}
}

// TestBarOperations tests rectangle/bar rendering methods
func TestBarOperations(t *testing.T) {
	pixfmt := NewMockPixelFormat(100, 100)
	renderer := NewRendererMClip[*MockPixelFormat, MockColorType](pixfmt)

	// Add overlapping clip boxes to test multi-region rendering
	renderer.AddClipBox(10, 10, 40, 40)
	renderer.AddClipBox(30, 30, 60, 60)

	testColor := "yellow"

	// Test CopyBar
	renderer.CopyBar(20, 20, 50, 50, testColor)

	// Verify some pixels in the overlapping region
	if pixfmt.Pixel(35, 35) != testColor {
		t.Errorf("Expected pixel at (35,35) to be %v, got %v", testColor, pixfmt.Pixel(35, 35))
	}

	// Test BlendBar
	renderer.BlendBar(15, 15, 35, 35, "cyan", 180)
	if pixfmt.Pixel(25, 25) != "cyan" {
		t.Errorf("Expected pixel at (25,25) to be cyan after bar blend, got %v", pixfmt.Pixel(25, 25))
	}
}

// TestSpanOperations tests span rendering methods
func TestSpanOperations(t *testing.T) {
	pixfmt := NewMockPixelFormat(100, 100)
	renderer := NewRendererMClip[*MockPixelFormat, MockColorType](pixfmt)

	renderer.AddClipBox(20, 20, 70, 70)

	testColor := "magenta"
	covers := []basics.Int8u{255, 200, 150, 100, 50}

	// Test BlendSolidHspan
	renderer.BlendSolidHspan(30, 30, 5, testColor, covers)
	for i := 0; i < 5; i++ {
		if pixfmt.Pixel(30+i, 30) != testColor {
			t.Errorf("Expected pixel at (%d,30) to be %v, got %v", 30+i, testColor, pixfmt.Pixel(30+i, 30))
		}
	}

	// Test BlendSolidVspan
	renderer.BlendSolidVspan(40, 40, 5, testColor, covers)
	for i := 0; i < 5; i++ {
		if pixfmt.Pixel(40, 40+i) != testColor {
			t.Errorf("Expected pixel at (40,%d) to be %v, got %v", 40+i, testColor, pixfmt.Pixel(40, 40+i))
		}
	}

	// Test CopyColorHspan
	colors := []interface{}{"red", "green", "blue", "yellow", "purple"}
	renderer.CopyColorHspan(25, 25, 5, colors)
	for i, expectedColor := range colors {
		if pixfmt.Pixel(25+i, 25) != expectedColor {
			t.Errorf("Expected pixel at (%d,25) to be %v, got %v", 25+i, expectedColor, pixfmt.Pixel(25+i, 25))
		}
	}

	// Test BlendColorHspan
	renderer.BlendColorHspan(50, 50, 5, colors, covers, 128)
	for i, expectedColor := range colors {
		if pixfmt.Pixel(50+i, 50) != expectedColor {
			t.Errorf("Expected pixel at (%d,50) to be %v, got %v", 50+i, expectedColor, pixfmt.Pixel(50+i, 50))
		}
	}

	// Test BlendColorVspan
	renderer.BlendColorVspan(55, 25, 5, colors, covers, 128)
	for i, expectedColor := range colors {
		if pixfmt.Pixel(55, 25+i) != expectedColor {
			t.Errorf("Expected pixel at (55,%d) to be %v, got %v", 25+i, expectedColor, pixfmt.Pixel(55, 25+i))
		}
	}
}

// TestBufferCopyOperations tests buffer copy methods
func TestBufferCopyOperations(t *testing.T) {
	pixfmt := NewMockPixelFormat(100, 100)
	renderer := NewRendererMClip[*MockPixelFormat, MockColorType](pixfmt)

	renderer.AddClipBox(10, 10, 50, 50)

	// Test Clear
	renderer.Clear("white")
	// Note: Clear should work on the entire surface, not just clip regions

	// Test CopyFrom (basic test - implementation may be simplified)
	srcBuffer := "dummy_source"
	rect := &basics.RectI{X1: 0, Y1: 0, X2: 20, Y2: 20}
	renderer.CopyFrom(srcBuffer, rect, 15, 15)
	// This mainly tests that the method doesn't panic and iterates through clip boxes

	// Test BlendFrom (basic test - implementation may be simplified)
	renderer.BlendFrom(srcBuffer, rect, 25, 25, 128)
	// This mainly tests that the method doesn't panic and iterates through clip boxes
}

// TestMultipleClipBoxIteration tests the iteration mechanism with multiple clip boxes
func TestMultipleClipBoxIteration(t *testing.T) {
	pixfmt := NewMockPixelFormat(100, 100)
	renderer := NewRendererMClip[*MockPixelFormat, MockColorType](pixfmt)

	// Add multiple clip boxes
	renderer.AddClipBox(10, 10, 30, 30) // Box 1
	renderer.AddClipBox(40, 40, 60, 60) // Box 2
	renderer.AddClipBox(70, 70, 90, 90) // Box 3

	// Test that pixel operations affect pixels in each clip box
	testColor := "multi_test"

	// This pixel should only appear in box 1 - test coordinate (20, 20)
	renderer.CopyPixel(20, 20, testColor)
	if pixfmt.Pixel(20, 20) != testColor {
		t.Error("Pixel in first clip box should be set")
	}

	// This pixel should only appear in box 2 - test coordinate (50, 50)
	renderer.CopyPixel(50, 50, testColor)
	if pixfmt.Pixel(50, 50) != testColor {
		t.Error("Pixel in second clip box should be set")
	}

	// This pixel should only appear in box 3 - test coordinate (80, 80)
	renderer.CopyPixel(80, 80, testColor)
	if pixfmt.Pixel(80, 80) != testColor {
		t.Error("Pixel in third clip box should be set")
	}

	// Test pixel outside all boxes should not be affected
	renderer.CopyPixel(5, 5, testColor)
	if pixfmt.Pixel(5, 5) == testColor {
		t.Error("Pixel outside all clip boxes should not be affected")
	}

	renderer.CopyPixel(35, 35, testColor) // Between box 1 and 2
	if pixfmt.Pixel(35, 35) == testColor {
		t.Error("Pixel between clip boxes should not be affected")
	}
}
