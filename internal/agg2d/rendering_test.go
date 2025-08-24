package agg

import (
	"testing"
)

// TestRenderingComponents verifies that rendering components are initialized
func TestRenderingComponents(t *testing.T) {
	agg2d := NewAgg2D()

	// Create a test buffer
	width, height := 100, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)

	// Attach buffer - this should initialize rendering components
	agg2d.Attach(buffer, width, height, stride)

	// Verify rendering components are initialized
	if agg2d.rasterizer == nil {
		t.Error("Expected rasterizer to be initialized")
	}

	if agg2d.scanline == nil {
		t.Error("Expected scanline to be initialized")
	}

	if agg2d.pixfmt == nil {
		t.Error("Expected pixfmt to be initialized")
	}

	// Test ClearAll functionality
	red := NewColorRGB(255, 0, 0)
	agg2d.ClearAll(red)

	// Verify buffer is filled with red
	for i := 0; i < len(buffer); i += 4 {
		if buffer[i] != 255 || buffer[i+1] != 0 || buffer[i+2] != 0 || buffer[i+3] != 255 {
			t.Errorf("Expected red pixel at offset %d, got RGBA(%d, %d, %d, %d)",
				i, buffer[i], buffer[i+1], buffer[i+2], buffer[i+3])
		}
	}
}

// TestBasicDrawing tests basic drawing functionality
func TestBasicDrawing(t *testing.T) {
	agg2d := NewAgg2D()

	// Create a test buffer
	width, height := 100, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)

	agg2d.Attach(buffer, width, height, stride)

	// Clear with white background
	agg2d.ClearAll(White)

	// Set fill color to blue
	agg2d.FillColor(Blue)

	// Try to draw a simple rectangle (even if rendering isn't fully functional)
	agg2d.Rectangle(10, 10, 50, 50)
	agg2d.DrawPath(FillOnly)

	// At this stage, we're just testing that the API calls don't crash
	// The actual rendering may not work due to incomplete implementation
	t.Log("Basic drawing API calls completed without crashing")
}

// TestGradientAPI tests gradient API functionality
func TestGradientAPI(t *testing.T) {
	agg2d := NewAgg2D()

	// Create a test buffer
	width, height := 100, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)

	agg2d.Attach(buffer, width, height, stride)

	// Test gradient setup (API only for now)
	agg2d.FillLinearGradient(0, 0, 100, 100, Red, Blue, 1.0)

	// Draw a shape with gradient
	agg2d.Rectangle(20, 20, 80, 80)
	agg2d.DrawPath(FillOnly)

	// At this stage, gradients may not render but API should not crash
	t.Log("Gradient API calls completed without crashing")
}

// TestRenderingTransformations tests transformation functionality
func TestRenderingTransformations(t *testing.T) {
	agg2d := NewAgg2D()

	// Test basic transformations
	agg2d.Translate(10, 20)
	agg2d.Scale(2.0, 1.5)
	agg2d.Rotate(45.0 * Pi / 180.0) // 45 degrees in radians

	// Test transformation stack
	agg2d.PushTransform()
	agg2d.Scale(0.5, 0.5)

	success := agg2d.PopTransform()
	if !success {
		t.Error("Expected PopTransform to succeed")
	}

	// Test coordinate transformation
	x, y := 100.0, 200.0
	agg2d.WorldToScreen(&x, &y)

	t.Logf("Transformed coordinates: (%f, %f)", x, y)
}
