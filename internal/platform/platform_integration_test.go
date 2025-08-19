package platform

import (
	"testing"
)

// TestPlatformSupportIntegration tests integration between PlatformSupport and backends
func TestPlatformSupportIntegration(t *testing.T) {
	// Create platform support
	ps := NewPlatformSupport(PixelFormatRGBA32, false)
	if ps == nil {
		t.Fatal("Failed to create PlatformSupport")
	}

	// Test initialization
	err := ps.Init(800, 600, WindowResize)
	if err != nil {
		t.Fatalf("Failed to initialize PlatformSupport: %v", err)
	}

	// Verify properties
	if ps.Width() != 800 || ps.Height() != 600 {
		t.Errorf("Expected dimensions (800,600), got (%d,%d)", ps.Width(), ps.Height())
	}

	if ps.Format() != PixelFormatRGBA32 {
		t.Errorf("Expected format RGBA32, got %s", ps.Format().String())
	}

	if ps.BPP() != 32 {
		t.Errorf("Expected 32 BPP, got %d", ps.BPP())
	}

	// Test buffer operations
	windowBuffer := ps.WindowBuffer()
	if windowBuffer == nil {
		t.Fatal("Window buffer is nil")
	}

	if windowBuffer.Width() != 800 || windowBuffer.Height() != 600 {
		t.Errorf("Buffer dimensions don't match window: buffer(%d,%d) vs window(%d,%d)",
			windowBuffer.Width(), windowBuffer.Height(), ps.Width(), ps.Height())
	}

	// Test image buffer operations
	if !ps.CreateImage(0, 200, 150) {
		t.Error("Failed to create image buffer 0")
	}

	imageBuffer := ps.ImageBuffer(0)
	if imageBuffer == nil {
		t.Fatal("Image buffer 0 is nil")
	}

	if imageBuffer.Width() != 200 || imageBuffer.Height() != 150 {
		t.Errorf("Image buffer dimensions incorrect: expected (200,150), got (%d,%d)",
			imageBuffer.Width(), imageBuffer.Height())
	}

	// Test image copy operations
	ps.CopyWindowToImage(1)
	imageBuffer1 := ps.ImageBuffer(1)
	if imageBuffer1 == nil {
		t.Fatal("Image buffer 1 is nil after copy")
	}

	if imageBuffer1.Width() != ps.Width() || imageBuffer1.Height() != ps.Height() {
		t.Errorf("Copied image buffer dimensions incorrect")
	}

	// Test invalid image indices
	if ps.ImageBuffer(-1) != nil {
		t.Error("Expected nil for invalid negative image index")
	}

	if ps.ImageBuffer(maxImages) != nil {
		t.Error("Expected nil for image index >= maxImages")
	}
}

// TestPixelFormatProperties tests pixel format properties
func TestPixelFormatProperties(t *testing.T) {
	testCases := []struct {
		format      PixelFormat
		expectedBPP int
		name        string
	}{
		{PixelFormatBW, 1, "BW"},
		{PixelFormatGray8, 8, "Gray8"},
		{PixelFormatRGB565, 16, "RGB565"},
		{PixelFormatRGB24, 24, "RGB24"},
		{PixelFormatRGBA32, 32, "RGBA32"},
		{PixelFormatRGB48, 48, "RGB48"},
		{PixelFormatRGBA64, 64, "RGBA64"},
		{PixelFormatRGB96, 96, "RGB96"},
		{PixelFormatRGBA128, 128, "RGBA128"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.format.BPP() != tc.expectedBPP {
				t.Errorf("Format %s: expected %d BPP, got %d",
					tc.format.String(), tc.expectedBPP, tc.format.BPP())
			}

			// Test string representation
			if tc.format.String() == "" {
				t.Errorf("Format %d has empty string representation", tc.format)
			}
		})
	}

	// Test undefined format
	if PixelFormatUndefined.BPP() != 0 {
		t.Errorf("Undefined format should have 0 BPP, got %d", PixelFormatUndefined.BPP())
	}
}

// TestEventHandling tests the event handling system
func TestEventHandling(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGBA32, false)

	// Track event calls
	var (
		initCalled   = false
		resizeCalled = false
		drawCalled   = false
		idleCalled   = false
		mouseCalled  = false
		keyCalled    = false
	)

	// Set up event handlers
	ps.SetOnInit(func() { initCalled = true })
	ps.SetOnResize(func(w, h int) { resizeCalled = true })
	ps.SetOnDraw(func() { drawCalled = true })
	ps.SetOnIdle(func() { idleCalled = true })
	ps.SetOnMouseMove(func(x, y int, flags InputFlags) { mouseCalled = true })
	ps.SetOnKey(func(x, y int, key KeyCode, flags InputFlags) { keyCalled = true })

	// Initialize (should trigger OnInit)
	err := ps.Init(800, 600, 0)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	if !initCalled {
		t.Error("OnInit was not called during initialization")
	}

	// Trigger events manually for testing
	ps.TriggerResize(1024, 768)
	if !resizeCalled {
		t.Error("OnResize was not called")
	}

	ps.TriggerDraw()
	if !drawCalled {
		t.Error("OnDraw was not called")
	}

	ps.TriggerIdle()
	if !idleCalled {
		t.Error("OnIdle was not called")
	}

	ps.TriggerMouseMove(100, 200, MouseLeft)
	if !mouseCalled {
		t.Error("OnMouseMove was not called")
	}

	ps.TriggerKey(50, 100, KeyEscape, KbdCtrl)
	if !keyCalled {
		t.Error("OnKey was not called")
	}
}

// TestInputFlagsIntegration tests the input flag system in integration context
func TestInputFlagsIntegration(t *testing.T) {
	// Test combined flags with platform events
	combined := MouseLeft | KbdShift
	if !combined.HasMouseLeft() {
		t.Error("Combined flags should have mouse left")
	}
	if !combined.HasShift() {
		t.Error("Combined flags should have shift")
	}
}

// TestKeyCodes tests the key code system
func TestKeyCodes(t *testing.T) {
	testCases := []struct {
		key      KeyCode
		expected string
	}{
		{KeyEscape, "Escape"},
		{KeyF1, "F1"},
		{KeyUp, "Up"},
		{KeyKP0, "KP0"},
		{KeyCode('A'), "'A'"},
		{KeyCode(' '), "' '"},
		{KeyCode(999), "Unknown(999)"},
	}

	for _, tc := range testCases {
		if tc.key.String() != tc.expected {
			t.Errorf("Key %d: expected '%s', got '%s'",
				int(tc.key), tc.expected, tc.key.String())
		}
	}

	// Test key type checking
	if !KeyF5.IsFunctionKey() {
		t.Error("F5 should be recognized as function key")
	}

	if !KeyUp.IsArrowKey() {
		t.Error("Up arrow should be recognized as arrow key")
	}

	if !KeyKP5.IsKeypadKey() {
		t.Error("KP5 should be recognized as keypad key")
	}

	if !KeyHome.IsNavigationKey() {
		t.Error("Home should be recognized as navigation key")
	}

	if !KeyCode('Z').IsASCII() {
		t.Error("'Z' should be recognized as ASCII")
	}
}

// TestWindowFlagsIntegration tests the window flag system in integration context
func TestWindowFlagsIntegration(t *testing.T) {
	// Test flag usage with platform support
	combined := WindowResize | WindowKeepAspectRatio

	ps := NewPlatformSupport(PixelFormatRGBA32, false)
	err := ps.Init(800, 600, combined)
	if err != nil {
		t.Errorf("Failed to initialize with combined window flags: %v", err)
	}

	if ps.WindowFlags() != combined {
		t.Error("Window flags not preserved during initialization")
	}
}

// TestBufferOperations tests buffer manipulation operations
func TestBufferOperations(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGBA32, false)
	err := ps.Init(400, 300, 0)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Test window buffer
	windowBuffer := ps.WindowBuffer()
	if windowBuffer.Buf() == nil {
		t.Error("Window buffer data is nil")
	}

	expectedSize := 400 * 300 * 4 // RGBA32
	if len(windowBuffer.Buf()) != expectedSize {
		t.Errorf("Expected buffer size %d, got %d", expectedSize, len(windowBuffer.Buf()))
	}

	// Test image buffer creation with different sizes
	sizes := []struct{ w, h int }{
		{100, 100},
		{200, 150},
		{50, 75},
	}

	for i, size := range sizes {
		if !ps.CreateImage(i, size.w, size.h) {
			t.Errorf("Failed to create image %d with size (%d,%d)", i, size.w, size.h)
			continue
		}

		imgBuffer := ps.ImageBuffer(i)
		if imgBuffer == nil {
			t.Errorf("Image buffer %d is nil", i)
			continue
		}

		if imgBuffer.Width() != size.w || imgBuffer.Height() != size.h {
			t.Errorf("Image %d: expected size (%d,%d), got (%d,%d)",
				i, size.w, size.h, imgBuffer.Width(), imgBuffer.Height())
		}

		expectedImgSize := size.w * size.h * 4 // RGBA32
		if len(imgBuffer.Buf()) != expectedImgSize {
			t.Errorf("Image %d: expected buffer size %d, got %d",
				i, expectedImgSize, len(imgBuffer.Buf()))
		}
	}

	// Test copy operations
	ps.CopyWindowToImage(5)
	imgBuffer5 := ps.ImageBuffer(5)
	if imgBuffer5 == nil {
		t.Error("Failed to copy window to image 5")
	} else if imgBuffer5.Width() != ps.Width() || imgBuffer5.Height() != ps.Height() {
		t.Error("Copied image dimensions don't match window")
	}

	// Test image to image copy
	ps.CopyImageToImage(6, 0)
	imgBuffer6 := ps.ImageBuffer(6)
	if imgBuffer6 == nil {
		t.Error("Failed to copy image 0 to image 6")
	} else {
		imgBuffer0 := ps.ImageBuffer(0)
		if imgBuffer6.Width() != imgBuffer0.Width() ||
			imgBuffer6.Height() != imgBuffer0.Height() {
			t.Error("Copied image dimensions don't match source")
		}
	}

	// Test copy from window to image and back
	ps.CopyImageToWindow(0)
	// This should work without error (though effect is not visible in test)
}

// TestEdgeCases tests edge cases and error conditions
func TestEdgeCases(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGBA32, false)

	// Test operations before initialization
	windowBuffer := ps.WindowBuffer()
	if windowBuffer == nil {
		t.Error("Window buffer should not be nil before init")
	}

	// Test small size initialization (should work)
	err := ps.Init(1, 1, 0)
	if err != nil {
		t.Errorf("Failed to initialize with minimal size: %v", err)
	}

	// Test proper initialization
	err = ps.Init(100, 100, 0)
	if err != nil {
		t.Fatalf("Failed to initialize with valid size: %v", err)
	}

	// Test multiple initializations (should work)
	err = ps.Init(200, 200, 0)
	if err != nil {
		t.Errorf("Failed to reinitialize: %v", err)
	}

	// Test very large sizes (should work but use lots of memory)
	err = ps.Init(10000, 10000, 0)
	if err != nil {
		// This is acceptable - system may not have enough memory
		t.Logf("Large size initialization failed (acceptable): %v", err)
	}
}

// BenchmarkPlatformSupport benchmarks platform support operations
func BenchmarkPlatformSupport(b *testing.B) {
	ps := NewPlatformSupport(PixelFormatRGBA32, false)
	ps.Init(800, 600, 0)

	b.Run("WindowBufferAccess", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ps.WindowBuffer()
		}
	})

	b.Run("CreateImage", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ps.CreateImage(0, 100, 100)
		}
	})

	b.Run("CopyOperations", func(b *testing.B) {
		ps.CreateImage(0, 100, 100)
		ps.CreateImage(1, 100, 100)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ps.CopyImageToImage(1, 0)
		}
	})
}
