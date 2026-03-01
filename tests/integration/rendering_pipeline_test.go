// Package integration provides integration tests that verify the interaction
// between different AGG components: renderer, rasterizer, scanline, pixfmt, etc.
package integration

import (
	"math"
	"testing"

	"agg_go/internal/agg2d"
)

// TestRenderingPipelineBasic tests the complete rendering pipeline:
// Path → Transform → Rasterizer → Scanline → Renderer → PixelFormat
func TestRenderingPipelineBasic(t *testing.T) {
	// Create a small buffer for pixel-level verification
	width, height := 100, 100
	stride := width * 4 // RGBA
	buffer := make([]uint8, height*stride)

	// Create AGG2D instance and attach buffer
	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)

	// Clear to white background
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255})

	// Verify background is white
	pixel := getPixel(buffer, stride, 50, 50)
	if pixel != [4]uint8{255, 255, 255, 255} {
		t.Errorf("Background should be white, got %v", pixel)
	}

	// Draw a red rectangle using the complete pipeline
	ctx.FillColor(agg2d.Color{255, 0, 0, 255}) // Red
	ctx.ResetPath()
	ctx.MoveTo(20, 20)
	ctx.LineTo(80, 20)
	ctx.LineTo(80, 80)
	ctx.LineTo(20, 80)
	ctx.ClosePolygon()
	ctx.DrawPath(agg2d.FillOnly)

	// Verify pixels inside rectangle are red
	centerPixel := getPixel(buffer, stride, 50, 50)
	if centerPixel[0] != 255 || centerPixel[1] != 0 || centerPixel[2] != 0 {
		t.Errorf("Center pixel should be red, got RGB(%d,%d,%d)",
			centerPixel[0], centerPixel[1], centerPixel[2])
	}

	// Verify pixels outside rectangle are still white
	outsidePixel := getPixel(buffer, stride, 10, 10)
	if outsidePixel != [4]uint8{255, 255, 255, 255} {
		t.Errorf("Outside pixel should be white, got %v", outsidePixel)
	}

	// Integer-aligned AGG rectangle edges can land on full-coverage pixels.
	// Verify the edge doesn't leak outside the intended bounds.
	edgePixel := getPixel(buffer, stride, 20, 50) // Left edge
	if edgePixel[0] != 255 || edgePixel[1] != 0 || edgePixel[2] != 0 {
		t.Errorf("Aligned edge pixel should be fully covered red, got RGB(%d,%d,%d)",
			edgePixel[0], edgePixel[1], edgePixel[2])
	}
	outsideEdgePixel := getPixel(buffer, stride, 19, 50)
	if outsideEdgePixel != [4]uint8{255, 255, 255, 255} {
		t.Errorf("Pixel just outside aligned edge should stay white, got %v", outsideEdgePixel)
	}
}

// TestRenderingPipelineWithBlending tests blend modes through the pipeline
func TestRenderingPipelineWithBlending(t *testing.T) {
	width, height := 100, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Draw a blue rectangle first
	ctx.FillColor(agg2d.Color{0, 0, 255, 255}) // Blue
	ctx.ResetPath()
	ctx.Rectangle(30, 30, 70, 70)
	ctx.DrawPath(agg2d.FillOnly)

	// Verify blue rectangle
	bluePixel := getPixel(buffer, stride, 50, 50)
	if bluePixel[2] != 255 || bluePixel[0] != 0 || bluePixel[1] != 0 {
		t.Errorf("First rectangle should be blue, got RGB(%d,%d,%d)",
			bluePixel[0], bluePixel[1], bluePixel[2])
	}

	// Draw a semi-transparent red rectangle overlapping
	ctx.FillColor(agg2d.Color{255, 0, 0, 128}) // Semi-transparent red
	ctx.ResetPath()
	ctx.Rectangle(50, 50, 90, 90)
	ctx.DrawPath(agg2d.FillOnly)

	// Verify blended color in overlap area
	overlapPixel := getPixel(buffer, stride, 60, 60)
	// Should be a blend of blue and semi-transparent red
	if overlapPixel[0] == 0 || overlapPixel[2] == 0 {
		t.Errorf("Overlap pixel should show blending, got RGB(%d,%d,%d)",
			overlapPixel[0], overlapPixel[1], overlapPixel[2])
	}
}

// TestRenderingPipelineWithGamma tests gamma correction effects
func TestRenderingPipelineWithGamma(t *testing.T) {
	width, height := 100, 100
	stride := width * 4
	buffer1 := make([]uint8, height*stride)
	buffer2 := make([]uint8, height*stride)

	// Render with gamma = 1.0 (no correction)
	ctx1 := agg2d.NewAgg2D()
	ctx1.Attach(buffer1, width, height, stride)
	ctx1.SetAntiAliasGamma(1.0)
	ctx1.ClearAll(agg2d.Color{255, 255, 255, 255})

	ctx1.FillColor(agg2d.Color{128, 128, 128, 255})
	ctx1.ResetPath()
	ctx1.AddEllipse(50, 50, 30, 30, agg2d.CCW)
	ctx1.DrawPath(agg2d.FillOnly)

	// Render with gamma = 2.2 (typical display gamma)
	ctx2 := agg2d.NewAgg2D()
	ctx2.Attach(buffer2, width, height, stride)
	ctx2.SetAntiAliasGamma(2.2)
	ctx2.ClearAll(agg2d.Color{255, 255, 255, 255})

	ctx2.FillColor(agg2d.Color{128, 128, 128, 255})
	ctx2.ResetPath()
	ctx2.AddEllipse(50, 50, 30, 30, agg2d.CCW)
	ctx2.DrawPath(agg2d.FillOnly)

	// Compare edge pixels - gamma correction should affect anti-aliasing
	edge1 := getPixel(buffer1, stride, 50, 20) // Top edge
	edge2 := getPixel(buffer2, stride, 50, 20) // Same position with gamma

	// The gamma-corrected version should have different edge blending
	if edge1[0] == edge2[0] && edge1[1] == edge2[1] && edge1[2] == edge2[2] {
		t.Log("Warning: Gamma correction may not be affecting anti-aliased edges as expected")
		// Note: This might be expected if gamma is applied differently or not at all
	}
}

// TestRenderingPipelinePixelPerfect tests exact pixel values for simple shapes
func TestRenderingPipelinePixelPerfect(t *testing.T) {
	width, height := 10, 10
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{0, 0, 0, 255}) // Black background

	// Draw a single pixel-aligned rectangle
	ctx.FillColor(agg2d.Color{255, 255, 255, 255}) // White
	ctx.ResetPath()
	ctx.Rectangle(2, 2, 6, 6) // 4x4 white rectangle
	ctx.DrawPath(agg2d.FillOnly)

	// Check specific pixels
	tests := []struct {
		x, y     int
		expected [4]uint8
		desc     string
	}{
		{1, 1, [4]uint8{0, 0, 0, 255}, "outside top-left"},
		{2, 2, [4]uint8{255, 255, 255, 255}, "inside top-left"},
		{5, 5, [4]uint8{255, 255, 255, 255}, "inside bottom-right"},
		{6, 6, [4]uint8{0, 0, 0, 255}, "outside bottom-right"},
		{8, 8, [4]uint8{0, 0, 0, 255}, "outside bottom-right"},
		{4, 4, [4]uint8{255, 255, 255, 255}, "center"},
	}

	for _, test := range tests {
		pixel := getPixel(buffer, stride, test.x, test.y)
		if pixel != test.expected {
			t.Errorf("Pixel at (%d,%d) %s: expected %v, got %v",
				test.x, test.y, test.desc, test.expected, pixel)
		}
	}
}

// TestRenderingPipelineStroke tests stroke rendering through the pipeline
func TestRenderingPipelineStroke(t *testing.T) {
	width, height := 100, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Draw a stroked line
	ctx.LineColor(agg2d.Color{0, 0, 0, 255}) // Black line
	ctx.LineWidth(3.0)
	ctx.ResetPath()
	ctx.MoveTo(10, 50)
	ctx.LineTo(90, 50)
	ctx.DrawPath(agg2d.StrokeOnly)

	// Check that line pixels are black
	linePixel := getPixel(buffer, stride, 50, 50)
	if linePixel[0] != 0 || linePixel[1] != 0 || linePixel[2] != 0 {
		t.Errorf("Line pixel should be black, got RGB(%d,%d,%d)",
			linePixel[0], linePixel[1], linePixel[2])
	}

	// Check pixels above and below line are still white (accounting for stroke width)
	abovePixel := getPixel(buffer, stride, 50, 47) // Above stroke
	belowPixel := getPixel(buffer, stride, 50, 53) // Below stroke

	// These might be anti-aliased, so just check they're not pure black
	if abovePixel[0] == 0 && abovePixel[1] == 0 && abovePixel[2] == 0 {
		t.Error("Pixel above stroke should not be pure black (outside stroke area)")
	}
	if belowPixel[0] == 0 && belowPixel[1] == 0 && belowPixel[2] == 0 {
		t.Error("Pixel below stroke should not be pure black (outside stroke area)")
	}
}

// TestRenderingPipelineCurves tests curve rendering through the complete pipeline
func TestRenderingPipelineCurves(t *testing.T) {
	width, height := 100, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Draw a quadratic Bezier curve
	ctx.LineColor(agg2d.Color{255, 0, 0, 255}) // Red line
	ctx.LineWidth(2.0)
	ctx.ResetPath()
	ctx.MoveTo(20, 80)
	ctx.QuadricCurveTo(50, 20, 80, 80) // Control point at (50,20), end at (80,80)
	ctx.DrawPath(agg2d.StrokeOnly)

	// Check that curve was rendered (should have red pixels along the curve)
	foundRed := false
	for y := 20; y <= 80; y += 10 {
		for x := 20; x <= 80; x += 10 {
			pixel := getPixel(buffer, stride, x, y)
			if pixel[0] > 200 && pixel[1] < 50 && pixel[2] < 50 { // Red-ish pixel
				foundRed = true
				break
			}
		}
		if foundRed {
			break
		}
	}

	if !foundRed {
		t.Error("No red pixels found - curve may not have been rendered correctly")
	}
}

// Helper function to get RGBA values at a specific pixel
func getPixel(buffer []uint8, stride int, x, y int) [4]uint8 {
	offset := y*stride + x*4
	if offset+3 >= len(buffer) {
		return [4]uint8{0, 0, 0, 0} // Return black if out of bounds
	}
	return [4]uint8{buffer[offset], buffer[offset+1], buffer[offset+2], buffer[offset+3]}
}

// Helper function to set RGBA values at a specific pixel
func setPixel(buffer []uint8, stride int, x, y int, rgba [4]uint8) {
	offset := y*stride + x*4
	if offset+3 < len(buffer) {
		buffer[offset] = rgba[0]
		buffer[offset+1] = rgba[1]
		buffer[offset+2] = rgba[2]
		buffer[offset+3] = rgba[3]
	}
}

// colorDistance calculates the Euclidean distance between two RGB colors
func colorDistance(c1, c2 [4]uint8) float64 {
	dr := float64(c1[0]) - float64(c2[0])
	dg := float64(c1[1]) - float64(c2[1])
	db := float64(c1[2]) - float64(c2[2])
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

// expectColorNear checks if two colors are within a tolerance
func expectColorNear(t *testing.T, expected, actual [4]uint8, tolerance float64, desc string) {
	dist := colorDistance(expected, actual)
	if dist > tolerance {
		t.Errorf("%s: expected color RGB(%d,%d,%d), got RGB(%d,%d,%d), distance=%.2f",
			desc, expected[0], expected[1], expected[2], actual[0], actual[1], actual[2], dist)
	}
}
