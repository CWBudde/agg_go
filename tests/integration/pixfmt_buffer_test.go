package integration

import (
	"testing"

	"agg_go/internal/agg2d"
)

// TestPixelFormatBufferBasic tests basic pixel format and buffer interaction
func TestPixelFormatBufferBasic(t *testing.T) {
	width, height := 100, 100
	stride := width * 4 // RGBA format
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)

	// Test buffer initialization
	ctx.ClearAll(agg2d.Color{128, 64, 192, 255}) // Purple background

	// Verify buffer was properly initialized
	for y := 0; y < height; y += 10 {
		for x := 0; x < width; x += 10 {
			pixel := getPixel(buffer, stride, x, y)
			if pixel[0] != 128 || pixel[1] != 64 || pixel[2] != 192 || pixel[3] != 255 {
				t.Errorf("Buffer initialization failed at (%d,%d): expected RGBA(128,64,192,255), got RGBA(%d,%d,%d,%d)",
					x, y, pixel[0], pixel[1], pixel[2], pixel[3])
				return // Stop after first failure to avoid spam
			}
		}
	}
}

// TestPixelFormatStride tests different buffer strides
func TestPixelFormatStride(t *testing.T) {
	width, height := 50, 50

	// Test with normal stride (width * 4)
	normalStride := width * 4
	normalBuffer := make([]uint8, height*normalStride)

	ctx1 := agg2d.NewAgg2D()
	ctx1.Attach(normalBuffer, width, height, normalStride)
	ctx1.ClearAll(agg2d.Color{255, 0, 0, 255})  // Red
	ctx1.FillColor(agg2d.Color{0, 255, 0, 255}) // Green
	drawFilledRectPath(ctx1, 10, 10, 40, 40)

	// Test with padded stride (width * 4 + 16 for alignment)
	paddedStride := width*4 + 16
	paddedBuffer := make([]uint8, height*paddedStride)

	ctx2 := agg2d.NewAgg2D()
	ctx2.Attach(paddedBuffer, width, height, paddedStride)
	ctx2.ClearAll(agg2d.Color{255, 0, 0, 255})  // Red
	ctx2.FillColor(agg2d.Color{0, 255, 0, 255}) // Green
	drawFilledRectPath(ctx2, 10, 10, 40, 40)

	// Compare results - they should be identical
	for y := 0; y < height; y += 5 {
		for x := 0; x < width; x += 5 {
			normalPixel := getPixelWithStride(normalBuffer, normalStride, x, y)
			paddedPixel := getPixelWithStride(paddedBuffer, paddedStride, x, y)

			if normalPixel != paddedPixel {
				t.Errorf("Stride difference at (%d,%d): normal=%v, padded=%v",
					x, y, normalPixel, paddedPixel)
			}
		}
	}
}

// TestPixelFormatDifferentSizes tests various buffer sizes
func TestPixelFormatDifferentSizes(t *testing.T) {
	sizes := []struct {
		width, height int
		desc          string
	}{
		{1, 1, "1x1 minimum"},
		{32, 32, "32x32 small"},
		{100, 50, "100x50 rectangular"},
		{256, 256, "256x256 medium"},
		{1000, 1, "1000x1 wide"},
		{1, 1000, "1x1000 tall"},
	}

	for _, size := range sizes {
		stride := size.width * 4
		buffer := make([]uint8, size.height*stride)

		ctx := agg2d.NewAgg2D()
		ctx.Attach(buffer, size.width, size.height, stride)

		// Clear and draw something simple
		ctx.ClearAll(agg2d.Color{100, 100, 100, 255}) // Gray

		if size.width > 10 && size.height > 10 {
			ctx.FillColor(agg2d.Color{255, 255, 255, 255}) // White
			drawFilledRectPath(ctx, 2, 2, float64(size.width-2), float64(size.height-2))

			// Check center pixel
			centerPixel := getPixel(buffer, stride, size.width/2, size.height/2)
			if centerPixel[0] < 200 || centerPixel[1] < 200 || centerPixel[2] < 200 {
				t.Errorf("Size %s: center should be white, got RGB(%d,%d,%d)",
					size.desc, centerPixel[0], centerPixel[1], centerPixel[2])
			}
		}

		t.Logf("Successfully tested buffer size %s (%dx%d)", size.desc, size.width, size.height)
	}
}

// TestPixelFormatClipping tests buffer boundary clipping
func TestPixelFormatClipping(t *testing.T) {
	width, height := 50, 50
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Draw shapes that extend beyond buffer boundaries
	ctx.FillColor(agg2d.Color{255, 0, 0, 255}) // Red

	testShapes := []struct {
		shape func()
		desc  string
	}{
		{func() {
			drawFilledRectPath(ctx, -10, -10, 20, 20) // Extends beyond top-left
		}, "top-left overflow"},
		{func() {
			drawFilledRectPath(ctx, 40, 40, 70, 70) // Extends beyond bottom-right
		}, "bottom-right overflow"},
		{func() {
			drawFilledRectPath(ctx, -10, 20, 70, 40) // Extends beyond left and right
		}, "horizontal overflow"},
	}

	for _, test := range testShapes {
		// Clear buffer before each test
		ctx.ClearAll(agg2d.Color{255, 255, 255, 255})

		// Draw shape that should be clipped
		test.shape()

		// Verify that only valid pixels are affected
		hasValidRedPixels := false
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				pixel := getPixel(buffer, stride, x, y)
				if pixel[0] > 200 && pixel[1] < 50 && pixel[2] < 50 { // Red pixel
					hasValidRedPixels = true
				}
			}
		}

		if !hasValidRedPixels {
			t.Errorf("Clipping test %s: should have some red pixels within buffer bounds", test.desc)
		}

		t.Logf("Clipping test %s: completed without buffer overflow", test.desc)
	}
}

// TestPixelFormatAlphaChannel tests alpha channel handling
func TestPixelFormatAlphaChannel(t *testing.T) {
	width, height := 100, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background with full alpha

	// Test different alpha values
	alphaTests := []struct {
		alpha uint8
		color agg2d.Color
		desc  string
	}{
		{255, agg2d.Color{255, 0, 0, 255}, "full opacity red"},
		{128, agg2d.Color{0, 255, 0, 128}, "half opacity green"},
		{64, agg2d.Color{0, 0, 255, 64}, "quarter opacity blue"},
		{0, agg2d.Color{255, 255, 0, 0}, "fully transparent yellow"},
	}

	for i, test := range alphaTests {
		x := (i%2)*40 + 10
		y := (i/2)*40 + 10

		ctx.FillColor(test.color)
		drawFilledRectPath(ctx, float64(x), float64(y), float64(x+30), float64(y+30))

		// Check rendered alpha
		pixel := getPixel(buffer, stride, x+15, y+15) // Center of rectangle

		switch test.alpha {
		case 255:
			// Full opacity should show pure color
			if pixel[0] < 200 || pixel[1] > 50 || pixel[2] > 50 {
				t.Errorf("Full opacity red should be pure red, got RGBA(%d,%d,%d,%d)",
					pixel[0], pixel[1], pixel[2], pixel[3])
			}
		case 128:
			// Half opacity should blend with background
			if pixel[1] < 100 { // Should have green component
				t.Errorf("Half opacity green should show blending, got RGBA(%d,%d,%d,%d)",
					pixel[0], pixel[1], pixel[2], pixel[3])
			}
		case 0:
			// Fully transparent should show background
			if pixel[0] < 200 || pixel[1] < 200 || pixel[2] < 200 {
				t.Errorf("Fully transparent should show white background, got RGBA(%d,%d,%d,%d)",
					pixel[0], pixel[1], pixel[2], pixel[3])
			}
		}

		t.Logf("Alpha test %s: pixel RGBA(%d,%d,%d,%d)",
			test.desc, pixel[0], pixel[1], pixel[2], pixel[3])
	}
}

// TestPixelFormatDirectAccess tests direct pixel manipulation vs. rendering
func TestPixelFormatDirectAccess(t *testing.T) {
	width, height := 100, 100
	stride := width * 4
	buffer1 := make([]uint8, height*stride) // Rendered
	buffer2 := make([]uint8, height*stride) // Direct manipulation

	// Render using AGG
	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer1, width, height, stride)
	ctx.ClearAll(agg2d.Color{0, 0, 0, 255})        // Black background
	ctx.FillColor(agg2d.Color{255, 255, 255, 255}) // White
	drawFilledRectPath(ctx, 25, 25, 75, 75)

	// Direct manipulation
	for i := range buffer2 {
		buffer2[i] = 0 // Initialize to black
	}

	// Set alpha channel
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			offset := y*stride + x*4 + 3 // Alpha channel
			buffer2[offset] = 255
		}
	}

	// Draw white rectangle directly
	for y := 25; y < 75; y++ {
		for x := 25; x < 75; x++ {
			setPixel(buffer2, stride, x, y, [4]uint8{255, 255, 255, 255})
		}
	}

	// Compare results
	differences := 0
	for y := 0; y < height; y += 5 {
		for x := 0; x < width; x += 5 {
			pixel1 := getPixel(buffer1, stride, x, y)
			pixel2 := getPixel(buffer2, stride, x, y)

			if pixel1 != pixel2 {
				differences++
				if differences == 1 { // Log first difference
					t.Logf("First difference at (%d,%d): rendered=%v, direct=%v",
						x, y, pixel1, pixel2)
				}
			}
		}
	}

	// Allow for some differences due to anti-aliasing
	maxAllowedDifferences := (width / 5) * (height / 5) / 10 // 10% tolerance
	if differences > maxAllowedDifferences {
		t.Errorf("Too many differences between rendered and direct manipulation: %d > %d",
			differences, maxAllowedDifferences)
	}
}

// TestPixelFormatColorAccuracy tests color accuracy in pixel format
func TestPixelFormatColorAccuracy(t *testing.T) {
	width, height := 100, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{255, 255, 255, 255}) // White background

	// Test precise color reproduction
	testColors := []agg2d.Color{
		{0, 0, 0, 255},       // Black
		{255, 255, 255, 255}, // White
		{255, 0, 0, 255},     // Red
		{0, 255, 0, 255},     // Green
		{0, 0, 255, 255},     // Blue
		{128, 128, 128, 255}, // Gray
		{255, 128, 64, 255},  // Orange
		{64, 192, 128, 255},  // Teal
	}

	for i, color := range testColors {
		x := (i % 4) * 25
		y := (i / 4) * 25

		ctx.FillColor(color)
		drawFilledRectPath(ctx, float64(x), float64(y), float64(x+20), float64(y+20))

		// Check color accuracy
		pixel := getPixel(buffer, stride, x+10, y+10)
		colorDistance := float64(0)
		for c := 0; c < 3; c++ { // RGB only
			diff := float64(pixel[c]) - float64(color[c])
			colorDistance += diff * diff
		}
		colorDistance = colorDistance / 3 // Average

		if colorDistance > 25.0 { // Allow small tolerance for anti-aliasing
			t.Errorf("Color accuracy test failed for RGBA(%d,%d,%d,%d): got RGBA(%d,%d,%d,%d), distance=%.2f",
				color[0], color[1], color[2], color[3],
				pixel[0], pixel[1], pixel[2], pixel[3], colorDistance)
		}
	}
}

// TestPixelFormatMemoryLayout tests pixel format memory layout
func TestPixelFormatMemoryLayout(t *testing.T) {
	width, height := 10, 10
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{0, 0, 0, 255}) // Black background

	// Set a specific pixel to red
	ctx.FillColor(agg2d.Color{255, 0, 0, 255})
	drawFilledRectPath(ctx, 5, 5, 6, 6) // Single pixel rectangle

	// Check memory layout directly
	pixelOffset := 5*stride + 5*4 // Row 5, Column 5, RGBA format

	if len(buffer) <= pixelOffset+3 {
		t.Fatalf("Buffer too small: len=%d, needed offset=%d", len(buffer), pixelOffset+3)
	}

	r := buffer[pixelOffset]
	g := buffer[pixelOffset+1]
	b := buffer[pixelOffset+2]
	a := buffer[pixelOffset+3]

	t.Logf("Direct memory access at pixel (5,5): RGBA(%d,%d,%d,%d)", r, g, b, a)

	// Verify RGBA layout
	if r < 200 || g > 50 || b > 50 || a != 255 {
		t.Errorf("Memory layout test failed: expected red pixel, got RGBA(%d,%d,%d,%d)",
			r, g, b, a)
	}

	// Cross-check with helper function
	helperPixel := getPixel(buffer, stride, 5, 5)
	if [4]uint8{r, g, b, a} != helperPixel {
		t.Errorf("Helper function mismatch: direct=%v, helper=%v",
			[4]uint8{r, g, b, a}, helperPixel)
	}
}

// Helper function to get pixel with custom stride
func getPixelWithStride(buffer []uint8, stride int, x, y int) [4]uint8 {
	offset := y*stride + x*4
	if offset+3 >= len(buffer) {
		return [4]uint8{0, 0, 0, 0}
	}
	return [4]uint8{buffer[offset], buffer[offset+1], buffer[offset+2], buffer[offset+3]}
}

// TestPixelFormatLargeBuffer tests handling of large buffers
func TestPixelFormatLargeBuffer(t *testing.T) {
	// Test with a moderately large buffer to ensure no overflow issues
	width, height := 1000, 1000
	stride := width * 4

	// This creates a 4MB buffer - reasonable for testing
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)

	// Clear to a specific pattern
	ctx.ClearAll(agg2d.Color{50, 100, 150, 255}) // Blue-ish

	// Draw something in the middle
	ctx.FillColor(agg2d.Color{255, 255, 0, 255}) // Yellow
	drawFilledRectPath(ctx, 450, 450, 550, 550)  // 100x100 rectangle in center

	// Check corners and center
	corners := []struct{ x, y int }{
		{0, 0},                  // Top-left
		{width - 1, 0},          // Top-right
		{0, height - 1},         // Bottom-left
		{width - 1, height - 1}, // Bottom-right
		{500, 500},              // Center (yellow rectangle)
	}

	for _, corner := range corners {
		pixel := getPixel(buffer, stride, corner.x, corner.y)

		if corner.x == 500 && corner.y == 500 {
			// Center should be yellow
			if pixel[0] < 200 || pixel[1] < 200 || pixel[2] > 50 {
				t.Errorf("Large buffer center should be yellow, got RGB(%d,%d,%d)",
					pixel[0], pixel[1], pixel[2])
			}
		} else {
			// Corners should be blue-ish background
			if pixel[0] != 50 || pixel[1] != 100 || pixel[2] != 150 {
				t.Errorf("Large buffer corner (%d,%d) should be background color, got RGB(%d,%d,%d)",
					corner.x, corner.y, pixel[0], pixel[1], pixel[2])
			}
		}
	}

	t.Logf("Successfully handled large buffer: %dx%d (%.1fMB)", width, height, float64(len(buffer))/(1024*1024))
}
