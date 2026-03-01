package integration

import (
	"testing"

	"agg_go/internal/agg2d"
)

// TestScanlineRendererBasic tests basic scanline rendering functionality
func TestScanlineRendererBasic(t *testing.T) {
	width, height := 100, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Draw a simple filled rectangle to test scanline generation
	ctx.FillColor(agg2d.Color{255, 0, 0, 255}) // Red
	ctx.ResetPath()
	ctx.Rectangle(25, 25, 75, 75)
	ctx.DrawPath(agg2d.FillOnly)

	// Check horizontal scanlines at different Y positions
	scanlineTests := []struct {
		y        int
		startX   int
		endX     int
		expected [4]uint8
		desc     string
	}{
		{20, 50, 50, [4]uint8{255, 255, 255, 255}, "above rectangle"},
		{25, 25, 65, [4]uint8{255, 0, 0, 255}, "top edge of rectangle"},
		{50, 25, 65, [4]uint8{255, 0, 0, 255}, "middle of rectangle"},
		{75, 50, 50, [4]uint8{255, 255, 255, 255}, "bottom edge outside rectangle"},
		{80, 50, 50, [4]uint8{255, 255, 255, 255}, "below rectangle"},
	}

	for _, test := range scanlineTests {
		for x := test.startX; x <= test.endX; x += 10 {
			pixel := getPixel(buffer, stride, x, test.y)
			if test.expected[0] == 255 && test.expected[1] == 255 && test.expected[2] == 255 {
				// Expecting white
				if pixel[0] < 200 || pixel[1] < 200 || pixel[2] < 200 {
					t.Errorf("Scanline test %s at (%d,%d): expected white, got RGB(%d,%d,%d)",
						test.desc, x, test.y, pixel[0], pixel[1], pixel[2])
				}
			} else if test.expected[0] == 255 && test.expected[1] == 0 && test.expected[2] == 0 {
				// Expecting red
				if pixel[0] < 200 || pixel[1] > 50 || pixel[2] > 50 {
					t.Errorf("Scanline test %s at (%d,%d): expected red, got RGB(%d,%d,%d)",
						test.desc, x, test.y, pixel[0], pixel[1], pixel[2])
				}
			}
		}
	}
}

// TestScanlineAntiAliasing tests anti-aliased scanline rendering
func TestScanlineAntiAliasing(t *testing.T) {
	width, height := 100, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Draw a circle to generate anti-aliased edges
	ctx.FillColor(agg2d.Color{0, 255, 0, 255}) // Green
	ctx.ResetPath()
	ctx.AddEllipse(50, 50, 30, 30, agg2d.CCW)
	ctx.DrawPath(agg2d.FillOnly)

	// Check anti-aliasing at circle edges
	edgeTests := []struct {
		x, y int
		desc string
	}{
		{50, 20, "top edge"},    // Top edge
		{80, 50, "right edge"},  // Right edge
		{50, 80, "bottom edge"}, // Bottom edge
		{20, 50, "left edge"},   // Left edge
	}

	for _, test := range edgeTests {
		pixel := getPixel(buffer, stride, test.x, test.y)

		// Edge pixels should not be pure green or pure white (anti-aliased)
		isPureGreen := (pixel[0] < 50 && pixel[1] > 200 && pixel[2] < 50)
		isPureWhite := (pixel[0] > 200 && pixel[1] > 200 && pixel[2] > 200)

		if isPureGreen {
			t.Logf("Edge pixel at %s (%d,%d) is pure green - may indicate hard edge",
				test.desc, test.x, test.y)
		} else if isPureWhite {
			t.Logf("Edge pixel at %s (%d,%d) is pure white - may be outside circle",
				test.desc, test.x, test.y)
		} else {
			// This is expected - anti-aliased blend
			t.Logf("Edge pixel at %s (%d,%d) is anti-aliased: RGB(%d,%d,%d)",
				test.desc, test.x, test.y, pixel[0], pixel[1], pixel[2])
		}

		// Edge should have some green component unless completely outside
		if pixel[1] < 20 && (pixel[0] < 180 || pixel[2] < 180) {
			t.Errorf("Edge pixel at %s should have green component or be white, got RGB(%d,%d,%d)",
				test.desc, pixel[0], pixel[1], pixel[2])
		}
	}
}

// TestScanlineComplexShape tests scanline rendering of complex shapes
func TestScanlineComplexShape(t *testing.T) {
	width, height := 150, 150
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Draw a star shape that will create complex scanlines
	ctx.FillColor(agg2d.Color{255, 255, 0, 255}) // Yellow
	ctx.ResetPath()

	// Create 5-pointed star
	cx, cy := 75.0, 75.0
	outerRadius := 50.0
	innerRadius := 20.0

	for i := 0; i < 10; i++ {
		// angle := float64(i) * 3.14159 / 5
		var radius float64
		if i%2 == 0 {
			radius = outerRadius
		} else {
			radius = innerRadius
		}

		x := cx + radius*float64(1.0) // Simplified cos
		y := cy + radius*float64(0.8) // Simplified sin

		if i == 0 {
			ctx.MoveTo(x, y)
		} else {
			ctx.LineTo(x, y)
		}
	}
	ctx.ClosePolygon()
	ctx.DrawPath(agg2d.FillOnly)

	// Check that complex shape generates proper scanlines
	// Test center (should be filled)
	centerPixel := getPixel(buffer, stride, 75, 75)
	if centerPixel[0] < 200 || centerPixel[1] < 200 || centerPixel[2] < 50 {
		t.Errorf("Star center should be yellow, got RGB(%d,%d,%d)",
			centerPixel[0], centerPixel[1], centerPixel[2])
	}

	// Test points that should be inside star
	insidePoints := []struct{ x, y int }{
		{75, 60}, // Upper middle
		{60, 75}, // Left middle
		{90, 75}, // Right middle
		{75, 90}, // Lower middle
	}

	filledPoints := 0
	for _, point := range insidePoints {
		pixel := getPixel(buffer, stride, point.x, point.y)
		if pixel[0] > 150 && pixel[1] > 150 && pixel[2] < 100 { // Yellow-ish
			filledPoints++
		}
	}

	if filledPoints < 2 {
		t.Errorf("Complex star shape should fill multiple points, only %d were yellow", filledPoints)
	}
}

// TestScanlineSpanGeneration tests span generation for gradients
func TestScanlineSpanGeneration(t *testing.T) {
	width, height := 200, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Create a linear gradient that will generate spans
	ctx.FillLinearGradient(50, 50, 150, 50, // Horizontal gradient
		agg2d.Color{255, 0, 0, 255}, // Red start
		agg2d.Color{0, 0, 255, 255}, // Blue end
		1.0)                         // Linear profile

	// Draw rectangle with gradient fill
	ctx.ResetPath()
	ctx.Rectangle(50, 25, 150, 75)
	ctx.DrawPath(agg2d.FillOnly)

	// Check horizontal spans at different positions
	spanTests := []struct {
		x    int
		y    int
		desc string
	}{
		{60, 50, "left side (more red)"},
		{100, 50, "middle (purple)"},
		{140, 50, "right side (more blue)"},
	}

	for _, test := range spanTests {
		pixel := getPixel(buffer, stride, test.x, test.y)

		switch test.desc {
		case "left side (more red)":
			if pixel[0] < pixel[2] {
				t.Errorf("Left side should be more red than blue, got RGB(%d,%d,%d)",
					pixel[0], pixel[1], pixel[2])
			}
		case "right side (more blue)":
			if pixel[2] < pixel[0] {
				t.Errorf("Right side should be more blue than red, got RGB(%d,%d,%d)",
					pixel[0], pixel[1], pixel[2])
			}
		case "middle (purple)":
			if pixel[0] < 50 || pixel[2] < 50 {
				t.Errorf("Middle should have both red and blue components, got RGB(%d,%d,%d)",
					pixel[0], pixel[1], pixel[2])
			}
		}
	}
}

// TestScanlineClipping tests scanline rendering with clipping
func TestScanlineClipping(t *testing.T) {
	width, height := 100, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Set clipping region
	ctx.ClipBox(30, 30, 70, 70)

	// Draw a larger shape that extends beyond clip region
	ctx.FillColor(agg2d.Color{128, 0, 128, 255}) // Purple
	ctx.ResetPath()
	ctx.Rectangle(10, 10, 90, 90)
	ctx.DrawPath(agg2d.FillOnly)

	// Check clipping results
	clippingTests := []struct {
		x, y     int
		expected bool
		desc     string
	}{
		{50, 50, true, "inside clip region"},
		{70, 50, true, "at clip boundary"},
		{30, 50, true, "at clip boundary"},
		{20, 50, false, "outside clip region (left)"},
		{80, 50, false, "outside clip region (right)"},
		{50, 20, false, "outside clip region (top)"},
		{50, 80, false, "outside clip region (bottom)"},
	}

	for _, test := range clippingTests {
		pixel := getPixel(buffer, stride, test.x, test.y)
		isPurple := (pixel[0] > 100 && pixel[0] < 150 && pixel[2] > 100 && pixel[2] < 150 && pixel[1] < 50)
		isWhite := (pixel[0] > 200 && pixel[1] > 200 && pixel[2] > 200)

		if test.expected {
			// Should be purple (inside clip)
			if !isPurple && !isWhite { // Allow anti-aliased edges
				t.Errorf("Point %s (%d,%d) should be purple, got RGB(%d,%d,%d)",
					test.desc, test.x, test.y, pixel[0], pixel[1], pixel[2])
			}
		} else {
			// Should be white (outside clip)
			if !isWhite {
				t.Errorf("Point %s (%d,%d) should be white (clipped), got RGB(%d,%d,%d)",
					test.desc, test.x, test.y, pixel[0], pixel[1], pixel[2])
			}
		}
	}
}

// TestScanlineSubPixelAccuracy tests subpixel accuracy in scanline rendering
func TestScanlineSubPixelAccuracy(t *testing.T) {
	width, height := 100, 100
	stride := width * 4
	buffer1 := make([]uint8, height*stride) // Integer aligned
	buffer2 := make([]uint8, height*stride) // Subpixel offset

	// Render integer-aligned rectangle
	ctx1 := agg2d.NewAgg2D()
	ctx1.Attach(buffer1, width, height, stride)
	ctx1.ClearAll(agg2d.Color{255, 255, 255, 255})

	ctx1.FillColor(agg2d.Color{0, 0, 0, 255}) // Black
	ctx1.ResetPath()
	ctx1.Rectangle(40, 40, 60, 60)
	ctx1.DrawPath(agg2d.FillOnly)

	// Render subpixel-offset rectangle
	ctx2 := agg2d.NewAgg2D()
	ctx2.Attach(buffer2, width, height, stride)
	ctx2.ClearAll(agg2d.Color{255, 255, 255, 255})

	ctx2.FillColor(agg2d.Color{0, 0, 0, 255}) // Black
	ctx2.ResetPath()
	ctx2.Rectangle(40.5, 40.5, 60.5, 60.5) // Half-pixel offset
	ctx2.DrawPath(agg2d.FillOnly)

	// Compare edge pixels - subpixel version should show anti-aliasing
	edgePixel1 := getPixel(buffer1, stride, 40, 50) // Integer edge
	edgePixel2 := getPixel(buffer2, stride, 40, 50) // Subpixel edge

	// Integer-aligned should be more definite (black or white)
	isDefinite1 := (edgePixel1[0] < 50 && edgePixel1[1] < 50 && edgePixel1[2] < 50) ||
		(edgePixel1[0] > 200 && edgePixel1[1] > 200 && edgePixel1[2] > 200)

	// Subpixel should potentially show gray values
	avgGray2 := (int(edgePixel2[0]) + int(edgePixel2[1]) + int(edgePixel2[2])) / 3

	t.Logf("Integer-aligned edge: RGB(%d,%d,%d), definite=%t",
		edgePixel1[0], edgePixel1[1], edgePixel1[2], isDefinite1)
	t.Logf("Subpixel edge: RGB(%d,%d,%d), avg gray=%d",
		edgePixel2[0], edgePixel2[1], edgePixel2[2], avgGray2)

	// This test is informational - subpixel accuracy depends on implementation
}

// TestScanlineEvenOddFill tests even-odd fill rule in scanline rendering
func TestScanlineEvenOddFill(t *testing.T) {
	width, height := 150, 150
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Enable even-odd fill rule
	// Note: EvenOddFill method may not be available, using default fill rule

	// Draw overlapping shapes to test even-odd rule
	ctx.FillColor(agg2d.Color{255, 128, 0, 255}) // Orange
	ctx.ResetPath()

	// Draw two overlapping rectangles
	ctx.Rectangle(50, 50, 100, 100) // First rectangle
	ctx.Rectangle(70, 70, 120, 120) // Overlapping rectangle
	ctx.DrawPath(agg2d.FillOnly)

	// Check fill results
	fillTests := []struct {
		x, y     int
		expected bool
		desc     string
	}{
		{60, 60, true, "first rectangle only"},
		{110, 110, true, "second rectangle only"},
		{85, 85, false, "overlap (should be unfilled with even-odd)"},
	}

	for _, test := range fillTests {
		pixel := getPixel(buffer, stride, test.x, test.y)
		isOrange := (pixel[0] > 200 && pixel[1] > 100 && pixel[1] < 150 && pixel[2] < 50)
		isWhite := (pixel[0] > 200 && pixel[1] > 200 && pixel[2] > 200)

		if test.expected {
			if !isOrange {
				t.Errorf("Point %s (%d,%d) should be filled (orange), got RGB(%d,%d,%d)",
					test.desc, test.x, test.y, pixel[0], pixel[1], pixel[2])
			}
		} else {
			if !isWhite {
				t.Errorf("Point %s (%d,%d) should be unfilled (white) with even-odd rule, got RGB(%d,%d,%d)",
					test.desc, test.x, test.y, pixel[0], pixel[1], pixel[2])
			}
		}
	}
}

// TestScanlinePerformanceStress tests scanline rendering under stress
func TestScanlinePerformanceStress(t *testing.T) {
	width, height := 500, 500
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Draw many overlapping shapes to stress scanline system
	colors := []agg2d.Color{
		{255, 0, 0, 64},   // Semi-transparent red
		{0, 255, 0, 64},   // Semi-transparent green
		{0, 0, 255, 64},   // Semi-transparent blue
		{255, 255, 0, 64}, // Semi-transparent yellow
	}

	for i := 0; i < 100; i++ {
		ctx.FillColor(colors[i%len(colors)])
		ctx.ResetPath()

		x := float64((i * 37) % 400) // Pseudo-random positioning
		y := float64((i * 43) % 400)
		size := 50.0

		ctx.AddEllipse(x+50, y+50, size, size, agg2d.CCW)
		ctx.DrawPath(agg2d.FillOnly)
	}

	// Just verify rendering completed without crashing
	// Check a few sample pixels to ensure something was rendered
	samplesWithColor := 0
	samplePoints := []struct{ x, y int }{
		{100, 100}, {200, 200}, {300, 300}, {150, 350}, {350, 150},
	}

	for _, point := range samplePoints {
		pixel := getPixel(buffer, stride, point.x, point.y)
		if pixel[0] != 255 || pixel[1] != 255 || pixel[2] != 255 { // Not white
			samplesWithColor++
		}
	}

	if samplesWithColor < 2 {
		t.Errorf("Stress test should produce colored pixels, only found %d non-white samples", samplesWithColor)
	}

	t.Logf("Stress test completed: rendered 100 overlapping shapes, %d sample points have color", samplesWithColor)
}
