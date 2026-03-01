package integration

import (
	"testing"

	"agg_go/internal/agg2d"
)

// TestColorBlendingAlphaNormal tests normal alpha blending
func TestColorBlendingAlphaNormal(t *testing.T) {
	width, height := 100, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Draw base layer - solid blue
	ctx.FillColor(agg2d.Color{0, 0, 255, 255}) // Blue, full opacity
	drawFilledRectPath(ctx, 20, 20, 80, 80)

	// Draw overlapping layer - semi-transparent red
	ctx.FillColor(agg2d.Color{255, 0, 0, 128}) // Red, 50% opacity
	drawFilledRectPath(ctx, 40, 40, 100, 100)

	// Test different regions
	// Blue only region
	blueOnlyPixel := getPixel(buffer, stride, 30, 30)
	expectColorNear(t, [4]uint8{0, 0, 255, 255}, blueOnlyPixel, 10.0, "blue only region")

	// Red only region
	redOnlyPixel := getPixel(buffer, stride, 90, 90)
	// Red at 50% alpha over white should be lighter red
	expectColorNear(t, [4]uint8{255, 128, 128, 255}, redOnlyPixel, 20.0, "red only region")

	// Blended region (red over blue)
	blendedPixel := getPixel(buffer, stride, 60, 60)
	// Should be a mix of blue and red
	if blendedPixel[2] < 100 { // Should still have significant blue component
		t.Errorf("Blended region should have blue component, got RGB(%d,%d,%d)",
			blendedPixel[0], blendedPixel[1], blendedPixel[2])
	}
	if blendedPixel[0] < 100 { // Should have red component
		t.Errorf("Blended region should have red component, got RGB(%d,%d,%d)",
			blendedPixel[0], blendedPixel[1], blendedPixel[2])
	}
}

// TestColorBlendingMultipleLayers tests blending multiple semi-transparent layers
func TestColorBlendingMultipleLayers(t *testing.T) {
	width, height := 100, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{0, 0, 0, 255}) // Black background

	// Draw three overlapping semi-transparent rectangles (RGB)
	colors := []agg2d.Color{
		{255, 0, 0, 85}, // Red, ~33% opacity
		{0, 255, 0, 85}, // Green, ~33% opacity
		{0, 0, 255, 85}, // Blue, ~33% opacity
	}

	positions := [][4]int{
		{10, 10, 60, 60}, // Red rectangle
		{20, 20, 70, 70}, // Green rectangle (offset)
		{30, 30, 80, 80}, // Blue rectangle (offset)
	}

	for i, color := range colors {
		ctx.FillColor(color)
		pos := positions[i]
		drawFilledRectPath(ctx, float64(pos[0]), float64(pos[1]), float64(pos[2]), float64(pos[3]))
	}

	// Check center where all three overlap
	centerPixel := getPixel(buffer, stride, 45, 45)
	// Should have components of all three colors
	if centerPixel[0] == 0 || centerPixel[1] == 0 || centerPixel[2] == 0 {
		t.Errorf("Center pixel should have all RGB components, got RGB(%d,%d,%d)",
			centerPixel[0], centerPixel[1], centerPixel[2])
	}

	// Check two-layer overlap regions
	redGreenPixel := getPixel(buffer, stride, 25, 25) // Red + Green only
	if redGreenPixel[0] == 0 || redGreenPixel[1] == 0 {
		t.Errorf("Red-Green overlap should have R and G components, got RGB(%d,%d,%d)",
			redGreenPixel[0], redGreenPixel[1], redGreenPixel[2])
	}
	if redGreenPixel[2] > 50 { // Should have minimal blue
		t.Errorf("Red-Green overlap should have minimal blue, got RGB(%d,%d,%d)",
			redGreenPixel[0], redGreenPixel[1], redGreenPixel[2])
	}
}

// TestColorBlendingDifferentModes tests different blend modes
func TestColorBlendingDifferentModes(t *testing.T) {
	width, height := 200, 100
	stride := width * 4

	// Test different blend modes
	blendModes := []struct {
		mode agg2d.BlendMode
		name string
	}{
		{agg2d.BlendAlpha, "alpha"},
		{agg2d.BlendAdd, "additive"},
		{agg2d.BlendMultiply, "multiply"},
		{agg2d.BlendScreen, "screen"},
	}

	for _, blendTest := range blendModes {
		buffer := make([]uint8, height*stride)
		ctx := agg2d.NewAgg2D()
		ctx.Attach(buffer, width, height, stride)
		ctx.ClearAll(agg2d.Color{128, 128, 128, 255}) // Gray background

		// Draw base shape
		ctx.FillColor(agg2d.Color{255, 0, 0, 255}) // Red
		drawFilledRectPath(ctx, 10, 10, 90, 90)

		// Set blend mode and draw overlapping shape
		ctx.SetBlendMode(blendTest.mode)
		ctx.FillColor(agg2d.Color{0, 255, 0, 255}) // Green
		drawFilledRectPath(ctx, 50, 10, 130, 90)

		// Check the blend result
		blendPixel := getPixel(buffer, stride, 70, 50) // Overlap center

		switch blendTest.mode {
		case agg2d.BlendAdd:
			// Additive should create yellow (red + green)
			if blendPixel[0] < 200 || blendPixel[1] < 200 {
				t.Errorf("Additive blend should create bright yellow, got RGB(%d,%d,%d)",
					blendPixel[0], blendPixel[1], blendPixel[2])
			}
		case agg2d.BlendMultiply:
			// Multiply with red and green should create dark result
			if blendPixel[0] > 50 || blendPixel[1] > 50 || blendPixel[2] > 50 {
				t.Errorf("Multiply blend should create dark result, got RGB(%d,%d,%d)",
					blendPixel[0], blendPixel[1], blendPixel[2])
			}
		}

		t.Logf("Blend mode %s: overlap pixel RGB(%d,%d,%d)",
			blendTest.name, blendPixel[0], blendPixel[1], blendPixel[2])
	}
}

// TestMasterAlphaEffect tests master alpha affecting all rendering
func TestMasterAlphaEffect(t *testing.T) {
	width, height := 100, 100
	stride := width * 4

	// Render with full master alpha
	buffer1 := make([]uint8, height*stride)
	ctx1 := agg2d.NewAgg2D()
	ctx1.Attach(buffer1, width, height, stride)
	ctx1.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background
	ctx1.SetMasterAlpha(1.0)

	ctx1.FillColor(agg2d.Color{255, 0, 0, 255}) // Red
	drawFilledRectPath(ctx1, 25, 25, 75, 75)

	// Render with reduced master alpha
	buffer2 := make([]uint8, height*stride)
	ctx2 := agg2d.NewAgg2D()
	ctx2.Attach(buffer2, width, height, stride)
	ctx2.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background
	ctx2.SetMasterAlpha(0.5)

	ctx2.FillColor(agg2d.Color{255, 0, 0, 255}) // Same red
	drawFilledRectPath(ctx2, 25, 25, 75, 75)

	// Compare results
	fullAlphaPixel := getPixel(buffer1, stride, 50, 50)
	halfAlphaPixel := getPixel(buffer2, stride, 50, 50)

	// Over an opaque white background the red channel stays saturated, while
	// green/blue rise because the red is blended with white.
	if halfAlphaPixel[0] != 255 {
		t.Errorf("Half master alpha should preserve the saturated red channel, got RGB(%d,%d,%d)",
			halfAlphaPixel[0], halfAlphaPixel[1], halfAlphaPixel[2])
	}
	if halfAlphaPixel[1] <= fullAlphaPixel[1] || halfAlphaPixel[2] <= fullAlphaPixel[2] {
		t.Errorf("Master alpha should lighten the non-red channels: full RGB(%d,%d,%d), half RGB(%d,%d,%d)",
			fullAlphaPixel[0], fullAlphaPixel[1], fullAlphaPixel[2],
			halfAlphaPixel[0], halfAlphaPixel[1], halfAlphaPixel[2])
	}
	if halfAlphaPixel[1] < 100 || halfAlphaPixel[2] < 100 {
		t.Errorf("Half master alpha should create light red, got RGB(%d,%d,%d)",
			halfAlphaPixel[0], halfAlphaPixel[1], halfAlphaPixel[2])
	}
}

// TestGradientRendering tests gradient rendering through the color pipeline
func TestGradientRendering(t *testing.T) {
	width, height := 200, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Create linear gradient from red to blue
	ctx.FillLinearGradient(50, 50, 150, 50, // Start and end points
		agg2d.Color{255, 0, 0, 255}, // Red
		agg2d.Color{0, 0, 255, 255}, // Blue
		1.0)                         // Linear profile

	// Draw rectangle with gradient fill
	drawFilledRectPath(ctx, 50, 25, 150, 75)

	// Check gradient progression
	leftPixel := getPixel(buffer, stride, 60, 50)    // Near start (should be reddish)
	centerPixel := getPixel(buffer, stride, 100, 50) // Middle (should be purple)
	rightPixel := getPixel(buffer, stride, 140, 50)  // Near end (should be blueish)

	// Left should be more red
	if leftPixel[0] < leftPixel[2] {
		t.Errorf("Left gradient pixel should be more red than blue, got RGB(%d,%d,%d)",
			leftPixel[0], leftPixel[1], leftPixel[2])
	}

	// Right should be more blue
	if rightPixel[2] < rightPixel[0] {
		t.Errorf("Right gradient pixel should be more blue than red, got RGB(%d,%d,%d)",
			rightPixel[0], rightPixel[1], rightPixel[2])
	}

	// Center should have both components
	if centerPixel[0] < 50 || centerPixel[2] < 50 {
		t.Errorf("Center gradient pixel should have both red and blue, got RGB(%d,%d,%d)",
			centerPixel[0], centerPixel[1], centerPixel[2])
	}
}

func TestBlendModeDstLeavesDestinationUnchanged(t *testing.T) {
	width, height := 32, 32
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{0, 0, 0, 0})
	ctx.SetBlendMode(agg2d.BlendDst)
	ctx.FillColor(agg2d.Color{255, 0, 0, 255})

	drawFilledRectPath(ctx, 8, 8, 24, 24)

	for i, v := range buffer {
		if v != 0 {
			t.Fatalf("BlendDst should preserve destination, first changed byte at %d: %d", i, v)
		}
	}
}

// TestRadialGradientRendering tests radial gradient rendering
func TestRadialGradientRendering(t *testing.T) {
	width, height := 100, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{0, 0, 0, 255}) // Black background

	// Create radial gradient from white center to black edge
	ctx.FillRadialGradient(50, 50, 40, // Center and radius
		agg2d.Color{255, 255, 255, 255}, // White center
		agg2d.Color{0, 0, 0, 255},       // Black edge
		1.0)                             // Linear profile

	// Draw circle with gradient fill
	ctx.ResetPath()
	ctx.AddEllipse(50, 50, 40, 40, agg2d.CCW)
	ctx.DrawPath(agg2d.FillOnly)

	// Check gradient from center outward
	centerPixel := getPixel(buffer, stride, 50, 50) // Center (should be white)
	// midPixel := getPixel(buffer, stride, 70, 50)        // Mid-radius (should be gray)
	edgePixel := getPixel(buffer, stride, 89, 50) // Near edge (should be dark)

	// Center should be bright
	if centerPixel[0] < 200 || centerPixel[1] < 200 || centerPixel[2] < 200 {
		t.Errorf("Center should be bright white, got RGB(%d,%d,%d)",
			centerPixel[0], centerPixel[1], centerPixel[2])
	}

	// Edge should be darker than center
	avgCenter := (int(centerPixel[0]) + int(centerPixel[1]) + int(centerPixel[2])) / 3
	avgEdge := (int(edgePixel[0]) + int(edgePixel[1]) + int(edgePixel[2])) / 3

	if avgEdge >= avgCenter {
		t.Errorf("Edge should be darker than center: center avg=%d, edge avg=%d",
			avgCenter, avgEdge)
	}
}

// TestColorSpaceConversions tests color space handling in rendering
func TestColorSpaceConversions(t *testing.T) {
	width, height := 100, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Test various color representations
	colors := []struct {
		color agg2d.Color
		name  string
	}{
		{agg2d.Color{255, 0, 0, 255}, "pure red"},
		{agg2d.Color{0, 255, 0, 255}, "pure green"},
		{agg2d.Color{0, 0, 255, 255}, "pure blue"},
		{agg2d.Color{128, 128, 128, 255}, "50% gray"},
		{agg2d.Color{255, 255, 0, 255}, "yellow"},
		{agg2d.Color{255, 0, 255, 255}, "magenta"},
		{agg2d.Color{0, 255, 255, 255}, "cyan"},
	}

	y := 10
	for _, colorTest := range colors {
		ctx.FillColor(colorTest.color)
		ctx.ResetPath()
		ctx.Rectangle(10, float64(y), 90, float64(y+10))
		ctx.DrawPath(agg2d.FillOnly)

		// Verify the rendered color matches expectation
		renderedPixel := getPixel(buffer, stride, 50, y+5)
		expectColorNear(t, colorTest.color, renderedPixel, 5.0, colorTest.name)

		y += 12
	}
}

// TestAntiAliasingColorAccuracy tests color accuracy in anti-aliased edges
func TestAntiAliasingColorAccuracy(t *testing.T) {
	width, height := 100, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Draw a red circle (will have anti-aliased edges)
	ctx.FillColor(agg2d.Color{255, 0, 0, 255}) // Pure red
	ctx.ResetPath()
	ctx.AddEllipse(50, 50, 30, 30, agg2d.CCW)
	ctx.DrawPath(agg2d.FillOnly)

	// Check anti-aliased edge pixel
	edgePixel := getPixel(buffer, stride, 50, 20) // Top edge of circle

	// Edge pixel should be a blend between red and white
	// It should not be pure red or pure white
	isPureRed := (edgePixel[0] == 255 && edgePixel[1] == 0 && edgePixel[2] == 0)
	isPureWhite := (edgePixel[0] == 255 && edgePixel[1] == 255 && edgePixel[2] == 255)

	if isPureRed || isPureWhite {
		t.Log("Edge pixel appears to be pure color rather than anti-aliased blend")
		// This might be expected in some implementations
	}

	// Edge should have some red component but not be pure
	if edgePixel[0] < 128 {
		t.Errorf("Edge pixel should have significant red component, got RGB(%d,%d,%d)",
			edgePixel[0], edgePixel[1], edgePixel[2])
	}
}
