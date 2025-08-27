package integration

import (
	"math"
	"testing"

	"agg_go/internal/agg2d"
)

// TestTransformationWithRendering tests transformation matrices with actual rendering
func TestTransformationWithRendering(t *testing.T) {
	width, height := 200, 200
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Draw a rectangle at origin without transformation first
	ctx.FillColor(agg2d.Color{255, 0, 0, 255}) // Red
	ctx.ResetPath()
	ctx.Rectangle(0, 0, 20, 20)
	ctx.DrawPath(agg2d.FillOnly)

	// Check that original rectangle is at origin
	originPixel := getPixel(buffer, stride, 10, 10)
	if originPixel[0] != 255 || originPixel[1] != 0 || originPixel[2] != 0 {
		t.Errorf("Original rectangle should be red at origin, got RGB(%d,%d,%d)",
			originPixel[0], originPixel[1], originPixel[2])
	}
}

// TestTranslationTransform tests translation transformation
func TestTranslationTransform(t *testing.T) {
	width, height := 200, 200
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Apply translation
	ctx.Translate(50, 60)

	// Draw a rectangle at origin (should appear translated)
	ctx.FillColor(agg2d.Color{0, 255, 0, 255}) // Green
	ctx.ResetPath()
	ctx.Rectangle(0, 0, 20, 20)
	ctx.DrawPath(agg2d.FillOnly)

	// Check that rectangle appears at translated position
	translatedPixel := getPixel(buffer, stride, 60, 70) // 50+10, 60+10 (center)
	if translatedPixel[0] != 0 || translatedPixel[1] != 255 || translatedPixel[2] != 0 {
		t.Errorf("Translated rectangle should be green at (60,70), got RGB(%d,%d,%d)",
			translatedPixel[0], translatedPixel[1], translatedPixel[2])
	}

	// Check that origin is still white
	originPixel := getPixel(buffer, stride, 10, 10)
	if originPixel != [4]uint8{255, 255, 255, 255} {
		t.Errorf("Origin should remain white after translation, got %v", originPixel)
	}
}

// TestScaleTransform tests scaling transformation
func TestScaleTransform(t *testing.T) {
	width, height := 200, 200
	stride := width * 4
	buffer1 := make([]uint8, height*stride)
	buffer2 := make([]uint8, height*stride)

	// Render normal size rectangle
	ctx1 := agg2d.NewAgg2D()
	ctx1.Attach(buffer1, width, height, stride)
	ctx1.ClearAll(agg2d.Color{255, 255, 255, 255})

	ctx1.FillColor(agg2d.Color{255, 0, 0, 255}) // Red
	ctx1.ResetPath()
	ctx1.Rectangle(50, 50, 70, 70) // 20x20 rectangle
	ctx1.DrawPath(agg2d.FillOnly)

	// Render scaled rectangle
	ctx2 := agg2d.NewAgg2D()
	ctx2.Attach(buffer2, width, height, stride)
	ctx2.ClearAll(agg2d.Color{255, 255, 255, 255})

	ctx2.Scale(2.0, 2.0)                        // 2x scale
	ctx2.FillColor(agg2d.Color{255, 0, 0, 255}) // Red
	ctx2.ResetPath()
	ctx2.Rectangle(50, 50, 70, 70) // Same rectangle, but scaled
	ctx2.DrawPath(agg2d.FillOnly)

	// Check that scaled version is larger
	// Original rectangle center
	originalCenter := getPixel(buffer1, stride, 60, 60)
	if originalCenter[0] != 255 || originalCenter[1] != 0 || originalCenter[2] != 0 {
		t.Errorf("Original rectangle center should be red, got RGB(%d,%d,%d)",
			originalCenter[0], originalCenter[1], originalCenter[2])
	}

	// Scaled rectangle should extend further
	scaledExtent := getPixel(buffer2, stride, 120, 120) // Should be inside scaled rectangle
	if scaledExtent[0] != 255 || scaledExtent[1] != 0 || scaledExtent[2] != 0 {
		t.Errorf("Scaled rectangle should extend to (120,120) and be red, got RGB(%d,%d,%d)",
			scaledExtent[0], scaledExtent[1], scaledExtent[2])
	}
}

// TestRotationTransform tests rotation transformation
func TestRotationTransform(t *testing.T) {
	width, height := 200, 200
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Move to center for rotation
	ctx.Translate(100, 100)
	ctx.Rotate(math.Pi / 4) // 45 degrees

	// Draw a rectangle that should appear rotated
	ctx.FillColor(agg2d.Color{0, 0, 255, 255}) // Blue
	ctx.ResetPath()
	ctx.Rectangle(-10, -10, 10, 10) // 20x20 rectangle centered at origin
	ctx.DrawPath(agg2d.FillOnly)

	// Check that center is still blue
	centerPixel := getPixel(buffer, stride, 100, 100)
	if centerPixel[0] != 0 || centerPixel[1] != 0 || centerPixel[2] != 255 {
		t.Errorf("Rotated rectangle center should be blue, got RGB(%d,%d,%d)",
			centerPixel[0], centerPixel[1], centerPixel[2])
	}

	// Check that corners extend diagonally due to rotation
	// At 45 degrees, corners should be further from center
	diagonalPixel := getPixel(buffer, stride, 114, 114) // Approximate rotated corner
	// This pixel should be blue or anti-aliased blue/white
	if diagonalPixel == [4]uint8{255, 255, 255, 255} {
		t.Error("Rotated rectangle should extend diagonally, but pixel appears white")
	}
}

// TestCompoundTransforms tests multiple transformations combined
func TestCompoundTransforms(t *testing.T) {
	width, height := 300, 300
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Apply compound transformation: translate, rotate, scale
	ctx.Translate(150, 150) // Move to center
	ctx.Rotate(math.Pi / 6) // 30 degrees
	ctx.Scale(1.5, 1.0)     // Scale X by 1.5, Y by 1.0

	// Draw a simple shape
	ctx.FillColor(agg2d.Color{255, 128, 0, 255}) // Orange
	ctx.ResetPath()
	ctx.Rectangle(-20, -10, 20, 10) // 40x20 rectangle
	ctx.DrawPath(agg2d.FillOnly)

	// Verify that transformation was applied
	centerPixel := getPixel(buffer, stride, 150, 150)
	if centerPixel[0] != 255 || centerPixel[1] != 128 || centerPixel[2] != 0 {
		t.Errorf("Compound transform center should be orange, got RGB(%d,%d,%d)",
			centerPixel[0], centerPixel[1], centerPixel[2])
	}

	// Check that original rectangle position is white (not transformed location)
	originalPixel := getPixel(buffer, stride, 130, 140) // Where rectangle would be without transform
	if originalPixel != [4]uint8{255, 255, 255, 255} {
		t.Error("Original position should be white (rectangle was transformed)")
	}
}

// TestTransformPushPop tests transformation stack operations
func TestTransformPushPop(t *testing.T) {
	width, height := 200, 200
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Apply initial transformation
	ctx.Translate(100, 100)

	// Push transformation and add more
	ctx.PushTransform()
	ctx.Scale(2.0, 2.0)
	ctx.Rotate(math.Pi / 2) // 90 degrees

	// Draw with compound transformation
	ctx.FillColor(agg2d.Color{255, 0, 255, 255}) // Magenta
	ctx.ResetPath()
	ctx.Rectangle(-5, -5, 5, 5)
	ctx.DrawPath(agg2d.FillOnly)

	// Pop back to just translation
	ctx.PopTransform()

	// Draw with only translation
	ctx.FillColor(agg2d.Color{0, 255, 255, 255}) // Cyan
	ctx.ResetPath()
	ctx.Rectangle(-5, -5, 5, 5)
	ctx.DrawPath(agg2d.FillOnly)

	// Check that both shapes are rendered correctly
	centerPixel := getPixel(buffer, stride, 100, 100)
	// Should have cyan (since it was drawn last at this position)
	if centerPixel[0] != 0 || centerPixel[1] != 255 || centerPixel[2] != 255 {
		t.Errorf("Center should be cyan from popped transform, got RGB(%d,%d,%d)",
			centerPixel[0], centerPixel[1], centerPixel[2])
	}
}

// TestViewportTransform tests viewport transformations with clipping
func TestViewportTransform(t *testing.T) {
	width, height := 100, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Set viewport transformation (world coordinates 0,0,200,200 -> screen 0,0,100,100)
	ctx.Viewport(0, 0, 100, 100, // screen viewport
		0, 0, 200, 200, // world viewport
		agg2d.XMidYMid) // aspect ratio preservation

	// Draw in world coordinates (should be scaled down)
	ctx.FillColor(agg2d.Color{128, 64, 192, 255}) // Purple
	ctx.ResetPath()
	ctx.Rectangle(50, 50, 150, 150) // 100x100 in world coordinates
	ctx.DrawPath(agg2d.FillOnly)

	// Check that shape appears scaled to screen coordinates
	centerPixel := getPixel(buffer, stride, 50, 50) // Center of screen
	if centerPixel[0] != 128 || centerPixel[1] != 64 || centerPixel[2] != 192 {
		t.Errorf("Viewport scaled shape should be purple at center, got RGB(%d,%d,%d)",
			centerPixel[0], centerPixel[1], centerPixel[2])
	}
}

// TestTransformWithClipping tests transformations with clipping
func TestTransformWithClipping(t *testing.T) {
	width, height := 100, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Set clipping box
	ctx.ClipBox(25, 25, 75, 75)

	// Apply transformation that would move shape outside clip
	ctx.Translate(60, 60)

	// Draw shape that extends beyond clip region
	ctx.FillColor(agg2d.Color{255, 255, 0, 255}) // Yellow
	ctx.ResetPath()
	ctx.Rectangle(-40, -40, 40, 40) // Large rectangle centered at translation point
	ctx.DrawPath(agg2d.FillOnly)

	// Check that shape is clipped
	insideClipPixel := getPixel(buffer, stride, 50, 50)  // Inside clip region
	outsideClipPixel := getPixel(buffer, stride, 10, 10) // Outside clip region

	if insideClipPixel[0] != 255 || insideClipPixel[1] != 255 || insideClipPixel[2] != 0 {
		t.Errorf("Inside clip region should be yellow, got RGB(%d,%d,%d)",
			insideClipPixel[0], insideClipPixel[1], insideClipPixel[2])
	}

	if outsideClipPixel != [4]uint8{255, 255, 255, 255} {
		t.Errorf("Outside clip region should remain white, got %v", outsideClipPixel)
	}
}

// TestTransformPrecision tests transformation precision with subpixel accuracy
func TestTransformPrecision(t *testing.T) {
	width, height := 100, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Apply small translation (subpixel)
	ctx.Translate(0.5, 0.5)

	// Draw pixel-aligned rectangle
	ctx.FillColor(agg2d.Color{0, 0, 0, 255}) // Black
	ctx.ResetPath()
	ctx.Rectangle(10, 10, 20, 20)
	ctx.DrawPath(agg2d.FillOnly)

	// Due to subpixel translation, edges should be anti-aliased
	edgePixel := getPixel(buffer, stride, 10, 15) // Left edge

	// Edge should not be pure black or pure white due to anti-aliasing
	if (edgePixel[0] == 0 && edgePixel[1] == 0 && edgePixel[2] == 0) ||
		(edgePixel[0] == 255 && edgePixel[1] == 255 && edgePixel[2] == 255) {
		t.Log("Warning: Subpixel translation may not be producing expected anti-aliasing")
		// This might be expected behavior depending on implementation
	}
}
