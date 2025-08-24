package platform

import (
	"testing"
)

func TestNewRenderingContext(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGBA32, false)
	ps.Init(800, 600, 0)

	rc := NewRenderingContext(ps)

	if rc.PlatformSupport() != ps {
		t.Error("RenderingContext should return the same platform support instance")
	}

	if rc.WindowBuffer() == nil {
		t.Error("Window buffer should not be nil")
	}

	if rc.ResizeTransform() == nil {
		t.Error("Resize transform should not be nil")
	}
}

func TestSetupResizeTransform(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGBA32, false)
	ps.Init(400, 300, WindowKeepAspectRatio)

	rc := NewRenderingContext(ps)

	// Test aspect ratio preserving resize
	rc.SetupResizeTransform(800, 600)

	// The transformation should scale uniformly
	// Original: 400x300, New: 800x600
	// Scale should be min(800/400, 600/300) = min(2.0, 2.0) = 2.0
	transform := rc.ResizeTransform()
	if transform == nil {
		t.Fatal("Resize transform should not be nil")
	}

	// Test non-aspect ratio preserving resize
	ps2 := NewPlatformSupport(PixelFormatRGBA32, false)
	ps2.Init(400, 300, 0) // No aspect ratio preservation

	rc2 := NewRenderingContext(ps2)
	rc2.SetupResizeTransform(800, 600)

	transform2 := rc2.ResizeTransform()
	if transform2 == nil {
		t.Fatal("Resize transform should not be nil")
	}
}

func TestTransformPoint(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGBA32, false)
	ps.Init(100, 100, 0)

	rc := NewRenderingContext(ps)
	rc.SetupResizeTransform(200, 200) // 2x scaling

	// Test forward transformation
	x, y := rc.TransformPoint(50, 50)
	if x != 100 || y != 100 {
		t.Errorf("TransformPoint(50, 50): expected (100, 100), got (%f, %f)", x, y)
	}

	// Test inverse transformation
	ix, iy := rc.InverseTransformPoint(100, 100)
	if ix != 50 || iy != 50 {
		t.Errorf("InverseTransformPoint(100, 100): expected (50, 50), got (%f, %f)", ix, iy)
	}
}

func TestClearWindow(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGBA32, false)
	ps.Init(10, 10, 0)

	rc := NewRenderingContext(ps)

	// Clear with red color
	rc.ClearWindow(255, 0, 0, 255)

	// Check a few pixels
	r, g, b, a, ok := rc.GetPixel(0, 0)
	if !ok {
		t.Fatal("GetPixel should succeed")
	}
	if r != 255 || g != 0 || b != 0 || a != 255 {
		t.Errorf("Pixel (0,0): expected RGBA(255,0,0,255), got RGBA(%d,%d,%d,%d)", r, g, b, a)
	}

	r, g, b, a, ok = rc.GetPixel(5, 5)
	if !ok {
		t.Fatal("GetPixel should succeed")
	}
	if r != 255 || g != 0 || b != 0 || a != 255 {
		t.Errorf("Pixel (5,5): expected RGBA(255,0,0,255), got RGBA(%d,%d,%d,%d)", r, g, b, a)
	}
}

func TestClearImage(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGBA32, false)
	ps.Init(100, 100, 0)

	rc := NewRenderingContext(ps)

	// Create an image buffer
	if !ps.CreateImage(0, 50, 50) {
		t.Fatal("Failed to create image buffer")
	}

	// Clear the image with blue color
	rc.ClearImage(0, 0, 0, 255, 255)

	// Test that image was cleared (we can't directly access image pixels through RC,
	// but we can test that the method doesn't crash)
}

func TestGetSetPixel(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGBA32, false)
	ps.Init(100, 100, 0)

	rc := NewRenderingContext(ps)

	// Test setting a pixel
	if !rc.SetPixel(50, 50, 128, 64, 192, 255) {
		t.Fatal("SetPixel should succeed")
	}

	// Test getting the pixel back
	r, g, b, a, ok := rc.GetPixel(50, 50)
	if !ok {
		t.Fatal("GetPixel should succeed")
	}
	if r != 128 || g != 64 || b != 192 || a != 255 {
		t.Errorf("Pixel: expected RGBA(128,64,192,255), got RGBA(%d,%d,%d,%d)", r, g, b, a)
	}

	// Test bounds checking
	if rc.SetPixel(-1, 50, 255, 255, 255, 255) {
		t.Error("SetPixel should fail for negative x")
	}

	if rc.SetPixel(50, -1, 255, 255, 255, 255) {
		t.Error("SetPixel should fail for negative y")
	}

	if rc.SetPixel(100, 50, 255, 255, 255, 255) {
		t.Error("SetPixel should fail for x >= width")
	}

	if rc.SetPixel(50, 100, 255, 255, 255, 255) {
		t.Error("SetPixel should fail for y >= height")
	}

	// Test getting out of bounds
	_, _, _, _, ok = rc.GetPixel(-1, 50)
	if ok {
		t.Error("GetPixel should fail for negative x")
	}

	_, _, _, _, ok = rc.GetPixel(100, 50)
	if ok {
		t.Error("GetPixel should fail for x >= width")
	}
}

func TestBlendPixel(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGBA32, false)
	ps.Init(100, 100, 0)

	rc := NewRenderingContext(ps)

	// Set a base pixel (white)
	rc.SetPixel(50, 50, 255, 255, 255, 255)

	// Blend with transparent red (alpha = 0)
	if !rc.BlendPixel(50, 50, 255, 0, 0, 0) {
		t.Fatal("BlendPixel should succeed")
	}

	// Should remain white
	r, g, b, a, ok := rc.GetPixel(50, 50)
	if !ok {
		t.Fatal("GetPixel should succeed")
	}
	if r != 255 || g != 255 || b != 255 || a != 255 {
		t.Errorf("After transparent blend: expected RGBA(255,255,255,255), got RGBA(%d,%d,%d,%d)", r, g, b, a)
	}

	// Blend with opaque red (alpha = 255)
	if !rc.BlendPixel(50, 50, 255, 0, 0, 255) {
		t.Fatal("BlendPixel should succeed")
	}

	// Should be red
	r, g, b, a, ok = rc.GetPixel(50, 50)
	if !ok {
		t.Fatal("GetPixel should succeed")
	}
	if r != 255 || g != 0 || b != 0 || a != 255 {
		t.Errorf("After opaque blend: expected RGBA(255,0,0,255), got RGBA(%d,%d,%d,%d)", r, g, b, a)
	}

	// Test semi-transparent blending
	rc.SetPixel(60, 60, 255, 255, 255, 255) // White
	rc.BlendPixel(60, 60, 0, 0, 0, 128)     // Semi-transparent black

	r, g, b, a, ok = rc.GetPixel(60, 60)
	if !ok {
		t.Fatal("GetPixel should succeed")
	}
	// Should be some shade of gray
	if r == 255 || g == 255 || b == 255 {
		t.Error("After semi-transparent blend, pixel should not be pure white")
	}
	if r == 0 || g == 0 || b == 0 {
		t.Error("After semi-transparent blend, pixel should not be pure black")
	}
}

func TestDrawLine(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGBA32, false)
	ps.Init(100, 100, 0)

	rc := NewRenderingContext(ps)
	rc.ClearWindow(0, 0, 0, 255) // Black background

	// Draw a horizontal line
	rc.DrawLine(10, 50, 90, 50, 255, 255, 255, 255) // White line

	// Check that some pixels on the line are white
	r, g, b, _, ok := rc.GetPixel(10, 50)
	if !ok {
		t.Fatal("GetPixel should succeed")
	}
	if r != 255 || g != 255 || b != 255 {
		t.Error("Start point of line should be white")
	}

	r, g, b, _, ok = rc.GetPixel(50, 50)
	if !ok {
		t.Fatal("GetPixel should succeed")
	}
	if r != 255 || g != 255 || b != 255 {
		t.Error("Middle point of line should be white")
	}

	r, g, b, _, ok = rc.GetPixel(90, 50)
	if !ok {
		t.Fatal("GetPixel should succeed")
	}
	if r != 255 || g != 255 || b != 255 {
		t.Error("End point of line should be white")
	}
}

func TestDrawRectangle(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGBA32, false)
	ps.Init(100, 100, 0)

	rc := NewRenderingContext(ps)
	rc.ClearWindow(0, 0, 0, 255) // Black background

	// Draw a rectangle outline
	rc.DrawRectangle(20, 20, 40, 30, 255, 0, 0, 255) // Red rectangle

	// Check corners
	r, g, b, _, ok := rc.GetPixel(20, 20)
	if !ok {
		t.Fatal("GetPixel should succeed")
	}
	if r != 255 || g != 0 || b != 0 {
		t.Error("Top-left corner should be red")
	}

	r, g, b, _, ok = rc.GetPixel(59, 20) // 20 + 40 - 1
	if !ok {
		t.Fatal("GetPixel should succeed")
	}
	if r != 255 || g != 0 || b != 0 {
		t.Error("Top-right corner should be red")
	}

	// Check that interior is still black
	r, g, b, _, ok = rc.GetPixel(30, 30)
	if !ok {
		t.Fatal("GetPixel should succeed")
	}
	if r != 0 || g != 0 || b != 0 {
		t.Error("Interior should remain black")
	}
}

func TestFillRectangle(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGBA32, false)
	ps.Init(100, 100, 0)

	rc := NewRenderingContext(ps)
	rc.ClearWindow(0, 0, 0, 255) // Black background

	// Fill a rectangle
	rc.FillRectangle(20, 20, 40, 30, 0, 255, 0, 255) // Green rectangle

	// Check that interior is green
	r, g, b, _, ok := rc.GetPixel(30, 30)
	if !ok {
		t.Fatal("GetPixel should succeed")
	}
	if r != 0 || g != 255 || b != 0 {
		t.Error("Interior should be green")
	}

	// Check that outside is still black
	r, g, b, _, ok = rc.GetPixel(10, 10)
	if !ok {
		t.Fatal("GetPixel should succeed")
	}
	if r != 0 || g != 0 || b != 0 {
		t.Error("Outside should remain black")
	}
}

func TestDrawCircle(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGBA32, false)
	ps.Init(100, 100, 0)

	rc := NewRenderingContext(ps)
	rc.ClearWindow(0, 0, 0, 255) // Black background

	// Draw a circle
	rc.DrawCircle(50, 50, 20, 255, 255, 0, 255) // Yellow circle

	// Check that some points on the circle are yellow
	// We can't easily test exact circle points without implementing the algorithm,
	// but we can check that the method doesn't crash and some pixels change
}

func TestGetBufferInfo(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGBA32, false)
	ps.Init(800, 600, 0)

	rc := NewRenderingContext(ps)

	width, height, stride, bpp, format := rc.GetBufferInfo()

	if width != 800 {
		t.Errorf("Expected width 800, got %d", width)
	}

	if height != 600 {
		t.Errorf("Expected height 600, got %d", height)
	}

	if stride != 800*4 { // 4 bytes per pixel for RGBA32
		t.Errorf("Expected stride %d, got %d", 800*4, stride)
	}

	if bpp != 32 {
		t.Errorf("Expected bpp 32, got %d", bpp)
	}

	if format != PixelFormatRGBA32 {
		t.Errorf("Expected format RGBA32, got %v", format)
	}
}

func TestGetImageInfo(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGB24, false)
	ps.Init(400, 300, 0)

	rc := NewRenderingContext(ps)

	// Test non-existent image
	_, _, _, _, _, ok := rc.GetImageInfo(0)
	if ok {
		t.Error("GetImageInfo should return false for non-existent image")
	}

	// Create an image
	if !ps.CreateImage(0, 200, 150) {
		t.Fatal("Failed to create image")
	}

	width, height, _, bpp, format, ok := rc.GetImageInfo(0)
	if !ok {
		t.Fatal("GetImageInfo should succeed for existing image")
	}

	if width != 200 {
		t.Errorf("Expected width 200, got %d", width)
	}

	if height != 150 {
		t.Errorf("Expected height 150, got %d", height)
	}

	if bpp != 24 {
		t.Errorf("Expected bpp 24, got %d", bpp)
	}

	if format != PixelFormatRGB24 {
		t.Errorf("Expected format RGB24, got %v", format)
	}
}

func TestValidateBufferAccess(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGBA32, false)
	ps.Init(100, 100, 0)

	rc := NewRenderingContext(ps)

	// Test valid coordinates
	if !rc.ValidateBufferAccess(0, 0) {
		t.Error("(0,0) should be valid")
	}

	if !rc.ValidateBufferAccess(99, 99) {
		t.Error("(99,99) should be valid")
	}

	if !rc.ValidateBufferAccess(50, 50) {
		t.Error("(50,50) should be valid")
	}

	// Test invalid coordinates
	if rc.ValidateBufferAccess(-1, 0) {
		t.Error("(-1,0) should be invalid")
	}

	if rc.ValidateBufferAccess(0, -1) {
		t.Error("(0,-1) should be invalid")
	}

	if rc.ValidateBufferAccess(100, 50) {
		t.Error("(100,50) should be invalid")
	}

	if rc.ValidateBufferAccess(50, 100) {
		t.Error("(50,100) should be invalid")
	}
}

func TestStatistics(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGBA32, false)
	ps.Init(800, 600, 0)

	rc := NewRenderingContext(ps)

	stats := rc.Statistics()

	// Check basic window info
	if stats.WindowWidth != 800 {
		t.Errorf("Expected window_width 800, got %v", stats.WindowWidth)
	}

	if stats.WindowHeight != 600 {
		t.Errorf("Expected window_height 600, got %v", stats.WindowHeight)
	}

	if stats.PixelFormat != "RGBA32" {
		t.Errorf("Expected pixel_format RGBA32, got %v", stats.PixelFormat)
	}

	if stats.BPP != 32 {
		t.Errorf("Expected bpp 32, got %v", stats.BPP)
	}

	if stats.FlipY != false {
		t.Errorf("Expected flip_y false, got %v", stats.FlipY)
	}

	// Check buffer size
	expectedSize := 800 * 600 * 4 // width * height * 4 bytes per pixel
	if stats.WindowBufferSize != expectedSize {
		t.Errorf("Expected window_buffer_size %d, got %v", expectedSize, stats.WindowBufferSize)
	}

	// Check image buffer count
	if stats.ActiveImageBuffers != 0 {
		t.Errorf("Expected active_image_buffers 0, got %v", stats.ActiveImageBuffers)
	}

	// Create some image buffers
	ps.CreateImage(0, 100, 100)
	ps.CreateImage(2, 50, 50)

	stats = rc.Statistics()
	if stats.ActiveImageBuffers != 2 {
		t.Errorf("Expected active_image_buffers 2, got %v", stats.ActiveImageBuffers)
	}

	// Test resize transform info
	rc.SetupResizeTransform(1600, 1200) // Scale by 2
	stats = rc.Statistics()

	if stats.HasResizeTransform != true {
		t.Error("Expected has_resize_transform to be true after setup")
	}

	// Check that scale values are present
	if stats.ResizeScaleX == 0 {
		t.Error("resize_scale_x should be present when transform is not identity")
	}
}

func TestDifferentPixelFormats(t *testing.T) {
	formats := []PixelFormat{
		PixelFormatRGBA32,
		PixelFormatBGRA32,
		PixelFormatRGB24,
		PixelFormatBGR24,
		PixelFormatGray8,
	}

	for _, format := range formats {
		t.Run(format.String(), func(t *testing.T) {
			ps := NewPlatformSupport(format, false)
			ps.Init(50, 50, 0)

			rc := NewRenderingContext(ps)

			// Clear with a color - should not crash
			rc.ClearWindow(128, 64, 192, 255)

			// Try to set and get a pixel - may not work perfectly for all formats
			// but should not crash
			rc.SetPixel(25, 25, 255, 128, 64, 255)
			_, _, _, _, _ = rc.GetPixel(25, 25)
		})
	}
}
