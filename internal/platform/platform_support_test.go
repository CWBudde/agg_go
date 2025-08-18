package platform

import (
	"testing"
	"time"
)

func TestNewPlatformSupport(t *testing.T) {
	tests := []struct {
		name   string
		format PixelFormat
		flipY  bool
	}{
		{"RGBA32 normal", PixelFormatRGBA32, false},
		{"RGBA32 flipped", PixelFormatRGBA32, true},
		{"RGB24 normal", PixelFormatRGB24, false},
		{"Gray8 flipped", PixelFormatGray8, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewPlatformSupport(tt.format, tt.flipY)
			
			if ps.Format() != tt.format {
				t.Errorf("Expected format %v, got %v", tt.format, ps.Format())
			}
			
			if ps.FlipY() != tt.flipY {
				t.Errorf("Expected flipY %v, got %v", tt.flipY, ps.FlipY())
			}
			
			expectedBPP := tt.format.BPP()
			if ps.BPP() != expectedBPP {
				t.Errorf("Expected BPP %d, got %d", expectedBPP, ps.BPP())
			}
			
			if ps.WaitMode() != false {
				t.Error("Expected initial wait mode to be false")
			}
			
			if ps.GetCaption() != "AGG Application" {
				t.Errorf("Expected default caption 'AGG Application', got '%s'", ps.GetCaption())
			}
		})
	}
}

func TestPixelFormat(t *testing.T) {
	tests := []struct {
		format      PixelFormat
		expectedBPP int
		expectedStr string
	}{
		{PixelFormatRGBA32, 32, "RGBA32"},
		{PixelFormatRGB24, 24, "RGB24"},
		{PixelFormatGray8, 8, "Gray8"},
		{PixelFormatBGRA32, 32, "BGRA32"},
		{PixelFormatGray16, 16, "Gray16"},
		{PixelFormatUndefined, 0, "Undefined"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedStr, func(t *testing.T) {
			if tt.format.BPP() != tt.expectedBPP {
				t.Errorf("Expected BPP %d, got %d", tt.expectedBPP, tt.format.BPP())
			}
			
			if tt.format.String() != tt.expectedStr {
				t.Errorf("Expected string '%s', got '%s'", tt.expectedStr, tt.format.String())
			}
		})
	}
}

func TestPlatformSupportInit(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGBA32, false)
	
	width, height := 800, 600
	flags := WindowResize | WindowKeepAspectRatio
	
	err := ps.Init(width, height, flags)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	
	if ps.Width() != width {
		t.Errorf("Expected width %d, got %d", width, ps.Width())
	}
	
	if ps.Height() != height {
		t.Errorf("Expected height %d, got %d", height, ps.Height())
	}
	
	if ps.InitialWidth() != width {
		t.Errorf("Expected initial width %d, got %d", width, ps.InitialWidth())
	}
	
	if ps.InitialHeight() != height {
		t.Errorf("Expected initial height %d, got %d", height, ps.InitialHeight())
	}
	
	if ps.WindowFlags() != flags {
		t.Errorf("Expected flags %v, got %v", flags, ps.WindowFlags())
	}
	
	// Check that window buffer is properly initialized
	buf := ps.WindowBuffer()
	if buf.Width() != width {
		t.Errorf("Buffer width: expected %d, got %d", width, buf.Width())
	}
	
	if buf.Height() != height {
		t.Errorf("Buffer height: expected %d, got %d", height, buf.Height())
	}
	
	expectedStride := width * ps.BPP() / 8
	if buf.Stride() != expectedStride {
		t.Errorf("Buffer stride: expected %d, got %d", expectedStride, buf.Stride())
	}
}

func TestCaption(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGB24, false)
	
	newCaption := "Test Application"
	ps.Caption(newCaption)
	
	if ps.GetCaption() != newCaption {
		t.Errorf("Expected caption '%s', got '%s'", newCaption, ps.GetCaption())
	}
}

func TestWaitMode(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGB24, false)
	
	// Initial state
	if ps.WaitMode() != false {
		t.Error("Expected initial wait mode to be false")
	}
	
	// Set to true
	ps.SetWaitMode(true)
	if ps.WaitMode() != true {
		t.Error("Expected wait mode to be true after setting")
	}
	
	// Set back to false
	ps.SetWaitMode(false)
	if ps.WaitMode() != false {
		t.Error("Expected wait mode to be false after setting back")
	}
}

func TestImageBuffers(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGBA32, false)
	ps.Init(400, 300, 0)
	
	// Test creating image
	if !ps.CreateImage(0, 200, 150) {
		t.Fatal("Failed to create image 0")
	}
	
	buf := ps.ImageBuffer(0)
	if buf == nil {
		t.Fatal("Image buffer 0 is nil")
	}
	
	if buf.Width() != 200 {
		t.Errorf("Image width: expected 200, got %d", buf.Width())
	}
	
	if buf.Height() != 150 {
		t.Errorf("Image height: expected 150, got %d", buf.Height())
	}
	
	// Test creating image with default dimensions
	if !ps.CreateImage(1, 0, 0) {
		t.Fatal("Failed to create image 1 with default dimensions")
	}
	
	buf1 := ps.ImageBuffer(1)
	if buf1.Width() != ps.Width() {
		t.Errorf("Image 1 width: expected %d, got %d", ps.Width(), buf1.Width())
	}
	
	if buf1.Height() != ps.Height() {
		t.Errorf("Image 1 height: expected %d, got %d", ps.Height(), buf1.Height())
	}
	
	// Test invalid image index
	if ps.ImageBuffer(-1) != nil {
		t.Error("Expected nil for negative image index")
	}
	
	if ps.ImageBuffer(maxImages) != nil {
		t.Error("Expected nil for image index >= maxImages")
	}
}

func TestImageCopyOperations(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGBA32, false)
	ps.Init(100, 100, 0)
	
	// Create an image buffer
	if !ps.CreateImage(0, 50, 50) {
		t.Fatal("Failed to create image 0")
	}
	
	// Test copy window to image
	ps.CopyWindowToImage(1)
	buf1 := ps.ImageBuffer(1)
	if buf1.Width() != ps.Width() {
		t.Errorf("Copied image width: expected %d, got %d", ps.Width(), buf1.Width())
	}
	
	// Test copy image to image
	ps.CopyImageToImage(2, 0)
	buf2 := ps.ImageBuffer(2)
	if buf2.Width() != 50 {
		t.Errorf("Copied image width: expected 50, got %d", buf2.Width())
	}
	
	// Test copy image to window
	ps.CopyImageToWindow(0) // Should not crash
}

func TestTimer(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGB24, false)
	
	ps.StartTimer()
	time.Sleep(10 * time.Millisecond)
	elapsed := ps.ElapsedTime()
	
	if elapsed < 5 { // Should be at least 5ms
		t.Errorf("Expected elapsed time >= 5ms, got %fms", elapsed)
	}
	
	if elapsed > 100 { // Should be less than 100ms
		t.Errorf("Expected elapsed time < 100ms, got %fms", elapsed)
	}
}

func TestEventHandlers(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGB24, false)
	
	var initCalled bool
	var resizeCalled bool
	var mouseMoveCalled bool
	var mouseDownCalled bool
	var mouseUpCalled bool
	var keyCalled bool
	var idleCalled bool
	var drawCalled bool
	
	// Set event handlers
	ps.SetOnInit(func() { initCalled = true })
	ps.SetOnResize(func(w, h int) { resizeCalled = true })
	ps.SetOnMouseMove(func(x, y int, flags InputFlags) { mouseMoveCalled = true })
	ps.SetOnMouseDown(func(x, y int, flags InputFlags) { mouseDownCalled = true })
	ps.SetOnMouseUp(func(x, y int, flags InputFlags) { mouseUpCalled = true })
	ps.SetOnKey(func(x, y int, key KeyCode, flags InputFlags) { keyCalled = true })
	ps.SetOnIdle(func() { idleCalled = true })
	ps.SetOnDraw(func() { drawCalled = true })
	ps.SetOnCtrlChange(func() { /* no-op for test */ })
	
	// Initialize (should trigger OnInit)
	ps.Init(100, 100, 0)
	if !initCalled {
		t.Error("OnInit handler was not called during Init")
	}
	
	// Trigger events
	ps.TriggerResize(200, 150)
	if !resizeCalled {
		t.Error("OnResize handler was not called")
	}
	
	ps.TriggerMouseMove(50, 50, MouseLeft)
	if !mouseMoveCalled {
		t.Error("OnMouseMove handler was not called")
	}
	
	ps.TriggerMouseDown(50, 50, MouseLeft)
	if !mouseDownCalled {
		t.Error("OnMouseDown handler was not called")
	}
	
	ps.TriggerMouseUp(50, 50, MouseLeft)
	if !mouseUpCalled {
		t.Error("OnMouseUp handler was not called")
	}
	
	ps.TriggerKey(50, 50, KeyEscape, KbdCtrl)
	if !keyCalled {
		t.Error("OnKey handler was not called")
	}
	
	ps.TriggerIdle()
	if !idleCalled {
		t.Error("OnIdle handler was not called")
	}
	
	ps.TriggerDraw()
	if !drawCalled {
		t.Error("OnDraw handler was not called")
	}
	
	// Force redraw should also trigger draw
	drawCalled = false
	ps.ForceRedraw()
	if !drawCalled {
		t.Error("ForceRedraw did not trigger OnDraw handler")
	}
}

func TestTriggerResize(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGBA32, false)
	ps.Init(100, 100, 0)
	
	var resizeWidth, resizeHeight int
	ps.SetOnResize(func(w, h int) {
		resizeWidth = w
		resizeHeight = h
	})
	
	// Trigger resize
	newWidth, newHeight := 200, 150
	ps.TriggerResize(newWidth, newHeight)
	
	// Check that dimensions were updated
	if ps.Width() != newWidth {
		t.Errorf("Width after resize: expected %d, got %d", newWidth, ps.Width())
	}
	
	if ps.Height() != newHeight {
		t.Errorf("Height after resize: expected %d, got %d", newHeight, ps.Height())
	}
	
	// Check that handler received correct values
	if resizeWidth != newWidth {
		t.Errorf("Resize handler width: expected %d, got %d", newWidth, resizeWidth)
	}
	
	if resizeHeight != newHeight {
		t.Errorf("Resize handler height: expected %d, got %d", newHeight, resizeHeight)
	}
	
	// Check that window buffer was resized
	buf := ps.WindowBuffer()
	if buf.Width() != newWidth || buf.Height() != newHeight {
		t.Errorf("Buffer size after resize: expected %dx%d, got %dx%d", 
			newWidth, newHeight, buf.Width(), buf.Height())
	}
}

func TestLoadSaveImage(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGB24, false)
	
	// These are stub implementations, so they should return success
	// but not actually load/save files
	if !ps.LoadImage(0, "test.bmp") {
		t.Error("LoadImage should return true (stub implementation)")
	}
	
	if !ps.SaveImage(0, "test.bmp") {
		t.Error("SaveImage should return true (stub implementation)")
	}
}

func TestImageExtension(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGB24, false)
	
	ext := ps.ImageExtension()
	if ext != ".bmp" {
		t.Errorf("Expected image extension '.bmp', got '%s'", ext)
	}
}

func TestRun(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGB24, false)
	
	var drawCalled bool
	ps.SetOnDraw(func() { drawCalled = true })
	
	result := ps.Run()
	if result != 0 {
		t.Errorf("Expected Run() to return 0, got %d", result)
	}
	
	if !drawCalled {
		t.Error("Run() should call OnDraw handler")
	}
}

func TestMessage(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGB24, false)
	
	// This should not crash (stub implementation)
	ps.Message("Test message")
}

func TestWindowFlags(t *testing.T) {
	tests := []struct {
		name     string
		flags    WindowFlags
		expected string
	}{
		{"No flags", 0, ""},
		{"Resize only", WindowResize, "WindowResize"},
		{"Multiple flags", WindowResize | WindowKeepAspectRatio, "WindowResize|WindowKeepAspectRatio"},
		{"All flags", WindowResize | WindowHWBuffer | WindowKeepAspectRatio | WindowProcessAllKeys, 
			"WindowResize|WindowHWBuffer|WindowKeepAspectRatio|WindowProcessAllKeys"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that flags can be set and retrieved
			ps := NewPlatformSupport(PixelFormatRGB24, false)
			ps.Init(100, 100, tt.flags)
			
			if ps.WindowFlags() != tt.flags {
				t.Errorf("Expected flags %v, got %v", tt.flags, ps.WindowFlags())
			}
		})
	}
}