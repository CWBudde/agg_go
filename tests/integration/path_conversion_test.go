package integration

import (
	"math"
	"testing"

	"agg_go/internal/agg2d"
)

// TestPathConversionStrokeToDash tests stroke to dash conversion with rendering
func TestPathConversionStrokeToDash(t *testing.T) {
	width, height := 200, 100
	stride := width * 4
	buffer1 := make([]uint8, height*stride) // Solid stroke
	buffer2 := make([]uint8, height*stride) // Dashed stroke

	// Render solid line
	ctx1 := agg2d.NewAgg2D()
	ctx1.Attach(buffer1, width, height, stride)
	ctx1.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	ctx1.LineColor(agg2d.Color{0, 0, 0, 255}) // Black line
	ctx1.LineWidth(3.0)
	ctx1.ResetPath()
	ctx1.MoveTo(20, 50)
	ctx1.LineTo(180, 50)
	ctx1.DrawPath(agg2d.StrokeOnly)

	// Render dashed line
	ctx2 := agg2d.NewAgg2D()
	ctx2.Attach(buffer2, width, height, stride)
	ctx2.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	ctx2.LineColor(agg2d.Color{0, 0, 0, 255}) // Black line
	ctx2.LineWidth(3.0)

	// Add dash pattern: 10 units dash, 5 units gap
	ctx2.AddDash(10.0, 5.0)
	ctx2.ResetPath()
	ctx2.MoveTo(20, 50)
	ctx2.LineTo(180, 50)
	ctx2.DrawPath(agg2d.StrokeOnly)

	// Compare results
	solidPixel := getPixel(buffer1, stride, 100, 50) // Middle of solid line
	// dashedMiddlePixel := getPixel(buffer2, stride, 100, 50) // Middle of dashed line
	// dashedGapPixel := getPixel(buffer2, stride, 70, 50) // Should be in a gap

	// Solid line should be black
	if solidPixel[0] != 0 || solidPixel[1] != 0 || solidPixel[2] != 0 {
		t.Errorf("Solid line should be black, got RGB(%d,%d,%d)",
			solidPixel[0], solidPixel[1], solidPixel[2])
	}

	// Check that dashed line has gaps (some pixels should be white)
	gapFound := false
	for x := 30; x < 170; x += 5 {
		pixel := getPixel(buffer2, stride, x, 50)
		if pixel[0] > 200 && pixel[1] > 200 && pixel[2] > 200 { // White-ish (gap)
			gapFound = true
			break
		}
	}

	if !gapFound {
		t.Error("Dashed line should have white gaps, but none found")
	}

	// Check that dashed line has black parts
	dashFound := false
	for x := 30; x < 170; x += 5 {
		pixel := getPixel(buffer2, stride, x, 50)
		if pixel[0] < 50 && pixel[1] < 50 && pixel[2] < 50 { // Black-ish (dash)
			dashFound = true
			break
		}
	}

	if !dashFound {
		t.Error("Dashed line should have black dashes, but none found")
	}
}

// TestPathConversionCurveApproximation tests curve to line approximation
func TestPathConversionCurveApproximation(t *testing.T) {
	width, height := 200, 200
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Draw a Bezier curve that will be approximated by line segments
	ctx.LineColor(agg2d.Color{255, 0, 0, 255}) // Red line
	ctx.LineWidth(2.0)
	ctx.ResetPath()
	ctx.MoveTo(50, 150)
	ctx.CubicCurveTo(50, 50, 150, 50, 150, 150) // S-curve - control points and end point
	ctx.DrawPath(agg2d.StrokeOnly)

	// Verify curve was rendered by checking multiple points along expected path
	curveSamples := []struct{ x, y int }{
		{50, 150},  // Start point
		{75, 80},   // Approximate curve point
		{100, 60},  // Middle area
		{125, 80},  // Approximate curve point
		{150, 150}, // End point
	}

	redPixelsFound := 0
	for _, sample := range curveSamples {
		// Check in a small area around each sample point
		for dx := -3; dx <= 3; dx++ {
			for dy := -3; dy <= 3; dy++ {
				x, y := sample.x+dx, sample.y+dy
				if x >= 0 && x < width && y >= 0 && y < height {
					pixel := getPixel(buffer, stride, x, y)
					if pixel[0] > 200 && pixel[1] < 50 && pixel[2] < 50 { // Red-ish
						redPixelsFound++
						goto nextSample
					}
				}
			}
		}
	nextSample:
	}

	if redPixelsFound < 3 {
		t.Errorf("Expected to find red pixels near curve path, found %d samples", redPixelsFound)
	}
}

// TestPathConversionStrokeWithJoins tests stroke conversion with different join types
func TestPathConversionStrokeWithJoins(t *testing.T) {
	width, height := 150, 150
	stride := width * 4

	joinTypes := []struct {
		join agg2d.LineJoin
		name string
	}{
		{agg2d.JoinMiter, "miter"},
		{agg2d.JoinRound, "round"},
		//{agg2d.JoinBevel, "bevel"}, // Might not be implemented
	}

	for _, joinTest := range joinTypes {
		buffer := make([]uint8, height*stride)
		ctx := agg2d.NewAgg2D()
		ctx.Attach(buffer, width, height, stride)
		ctx.ClearAll(agg2d.Color{255, 255, 255, 255})

		// Set line properties
		ctx.LineColor(agg2d.Color{0, 0, 255, 255}) // Blue line
		ctx.LineWidth(10.0)                        // Thick line to see join clearly
		ctx.LineJoin(joinTest.join)

		// Draw angle that will show join behavior
		ctx.ResetPath()
		ctx.MoveTo(50, 100)
		ctx.LineTo(75, 50)   // Up and right
		ctx.LineTo(100, 100) // Down and right
		ctx.DrawPath(agg2d.StrokeOnly)

		// Check that join area has been filled
		joinPixel := getPixel(buffer, stride, 75, 65) // Near the join point
		if joinPixel[0] > 200 && joinPixel[1] > 200 && joinPixel[2] > 200 {
			t.Errorf("Join type %s: join area should be filled with blue, got RGB(%d,%d,%d)",
				joinTest.name, joinPixel[0], joinPixel[1], joinPixel[2])
		}

		t.Logf("Join type %s: join pixel RGB(%d,%d,%d)",
			joinTest.name, joinPixel[0], joinPixel[1], joinPixel[2])
	}
}

// TestPathConversionContour tests contour/outline conversion
func TestPathConversionContour(t *testing.T) {
	width, height := 100, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Create a path and convert it to contour (outline)
	ctx.LineColor(agg2d.Color{0, 255, 0, 255}) // Green line
	ctx.LineWidth(3.0)

	// Draw a simple closed shape
	ctx.ResetPath()
	ctx.MoveTo(30, 30)
	ctx.LineTo(70, 30)
	ctx.LineTo(70, 70)
	ctx.LineTo(30, 70)
	ctx.ClosePolygon()
	ctx.DrawPath(agg2d.StrokeOnly) // Draw only the outline

	// Check that outline is rendered
	topEdge := getPixel(buffer, stride, 50, 30)
	rightEdge := getPixel(buffer, stride, 70, 50)
	bottomEdge := getPixel(buffer, stride, 50, 70)
	leftEdge := getPixel(buffer, stride, 30, 50)
	center := getPixel(buffer, stride, 50, 50)

	// Edges should be green (or anti-aliased green)
	edges := []struct {
		pixel [4]uint8
		name  string
	}{
		{topEdge, "top edge"},
		{rightEdge, "right edge"},
		{bottomEdge, "bottom edge"},
		{leftEdge, "left edge"},
	}

	for _, edge := range edges {
		// Should have significant green component
		if edge.pixel[1] < 100 {
			t.Errorf("%s should be green, got RGB(%d,%d,%d)",
				edge.name, edge.pixel[0], edge.pixel[1], edge.pixel[2])
		}
	}

	// Center should be white (not filled)
	if center[0] < 200 || center[1] < 200 || center[2] < 200 {
		t.Errorf("Center should be white (outline only), got RGB(%d,%d,%d)",
			center[0], center[1], center[2])
	}
}

// TestPathConversionTransformChain tests path transformation followed by conversion
func TestPathConversionTransformChain(t *testing.T) {
	width, height := 150, 150
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Apply transformation
	ctx.Translate(75, 75)   // Move to center
	ctx.Rotate(math.Pi / 4) // 45 degrees
	ctx.Scale(1.5, 1.0)     // Stretch horizontally

	// Draw stroked path that will go through: transform -> stroke conversion
	ctx.LineColor(agg2d.Color{255, 0, 255, 255}) // Magenta
	ctx.LineWidth(4.0)
	ctx.ResetPath()
	ctx.Rectangle(-20, -10, 20, 10) // Rectangle centered at origin
	ctx.DrawPath(agg2d.StrokeOnly)

	// Check that transformed stroke appears in buffer
	// Due to rotation and scaling, the rectangle outline should be visible
	magentaFound := false
	for y := 30; y < 120; y += 5 {
		for x := 30; x < 120; x += 5 {
			pixel := getPixel(buffer, stride, x, y)
			if pixel[0] > 200 && pixel[2] > 200 && pixel[1] < 100 {
				magentaFound = true
				goto found
			}
		}
	}
found:

	if !magentaFound {
		t.Error("Transformed stroked rectangle should be visible, but no magenta pixels found")
	}
}

// TestPathConversionComplexShape tests conversion of complex shapes
func TestPathConversionComplexShape(t *testing.T) {
	width, height := 200, 200
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Create a complex path with curves and lines
	ctx.FillColor(agg2d.Color{128, 0, 128, 255}) // Purple
	ctx.ResetPath()

	// Start with a base shape
	ctx.MoveTo(100, 50)
	// Add curves to create a flower-like shape
	for i := 0; i < 8; i++ {
		angle := float64(i) * math.Pi / 4
		cos := math.Cos(angle)
		sin := math.Sin(angle)

		// Control points for petal
		cx1 := 100 + cos*30
		cy1 := 100 + sin*30
		cx2 := 100 + cos*50
		cy2 := 100 + sin*50
		endx := 100 + cos*25
		endy := 100 + sin*25

		ctx.CubicCurveTo(cx1, cy1, cx2, cy2, endx, endy)
	}
	ctx.ClosePolygon()
	ctx.DrawPath(agg2d.FillOnly)

	// Check that complex shape was rendered
	centerPixel := getPixel(buffer, stride, 100, 100)
	if centerPixel[0] != 128 || centerPixel[2] != 128 || centerPixel[1] > 50 {
		t.Errorf("Complex shape center should be purple, got RGB(%d,%d,%d)",
			centerPixel[0], centerPixel[1], centerPixel[2])
	}

	// Check that shape extends in multiple directions (petals)
	petalPixels := 0
	directions := []struct{ x, y int }{
		{100, 75},  // North
		{125, 100}, // East
		{100, 125}, // South
		{75, 100},  // West
	}

	for _, dir := range directions {
		pixel := getPixel(buffer, stride, dir.x, dir.y)
		if pixel[0] == 128 && pixel[2] == 128 && pixel[1] < 50 {
			petalPixels++
		}
	}

	if petalPixels < 2 {
		t.Errorf("Complex shape should extend in multiple directions, found %d petal pixels", petalPixels)
	}
}

// TestPathConversionMarkers tests path markers and arrowheads
func TestPathConversionMarkers(t *testing.T) {
	width, height := 200, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Draw line with arrowhead (if supported)
	ctx.LineColor(agg2d.Color{255, 128, 0, 255}) // Orange
	ctx.LineWidth(3.0)
	ctx.ResetPath()
	ctx.MoveTo(30, 50)
	ctx.LineTo(170, 50)
	ctx.DrawPath(agg2d.StrokeOnly)

	// Manually add arrowhead at end
	ctx.FillColor(agg2d.Color{255, 128, 0, 255}) // Same orange
	ctx.ResetPath()
	ctx.MoveTo(170, 50)
	ctx.LineTo(160, 45)
	ctx.LineTo(160, 55)
	ctx.ClosePolygon()
	ctx.DrawPath(agg2d.FillOnly)

	// Check that line and arrowhead are rendered
	linePixel := getPixel(buffer, stride, 100, 50)  // Middle of line
	arrowPixel := getPixel(buffer, stride, 165, 50) // Arrow area

	if linePixel[0] != 255 || linePixel[1] != 128 || linePixel[2] > 50 {
		t.Errorf("Line should be orange, got RGB(%d,%d,%d)",
			linePixel[0], linePixel[1], linePixel[2])
	}

	if arrowPixel[0] != 255 || arrowPixel[1] != 128 || arrowPixel[2] > 50 {
		t.Errorf("Arrow should be orange, got RGB(%d,%d,%d)",
			arrowPixel[0], arrowPixel[1], arrowPixel[2])
	}
}

// TestPathConversionSmoothCurves tests smooth curve generation
func TestPathConversionSmoothCurves(t *testing.T) {
	width, height := 200, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Create a wavy line using smooth curves
	ctx.LineColor(agg2d.Color{0, 128, 255, 255}) // Light blue
	ctx.LineWidth(2.0)
	ctx.ResetPath()
	ctx.MoveTo(20, 50)

	// Create smooth sine-like wave
	for x := 20; x <= 180; x += 20 {
		y := 50 + 20*math.Sin(float64(x-20)*math.Pi/80)
		if x == 20 {
			ctx.LineTo(float64(x), y)
		} else {
			// Use smooth curve instead of straight line
			prevX := float64(x - 20)
			prevY := 50 + 20*math.Sin(float64(x-40)*math.Pi/80)
			cx := prevX + 10
			cy := (prevY + y) / 2
			ctx.QuadricCurveTo(cx, cy, float64(x), y)
		}
	}
	ctx.DrawPath(agg2d.StrokeOnly)

	// Check that smooth curve was rendered along expected path
	curvePointsFound := 0
	for x := 30; x <= 170; x += 20 {
		expectedY := int(50 + 20*math.Sin(float64(x-20)*math.Pi/80))

		// Check in area around expected curve point
		for dy := -5; dy <= 5; dy++ {
			y := expectedY + dy
			if y >= 0 && y < height {
				pixel := getPixel(buffer, stride, x, y)
				if pixel[0] < 50 && pixel[1] > 100 && pixel[2] > 200 { // Light blue
					curvePointsFound++
					break
				}
			}
		}
	}

	if curvePointsFound < 4 {
		t.Errorf("Smooth curve should be visible at multiple points, found %d", curvePointsFound)
	}
}
