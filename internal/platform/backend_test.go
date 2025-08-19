package platform

import (
	"testing"

	"agg_go/internal/buffer"
)

// TestMockBackend tests the mock backend implementation
func TestMockBackend(t *testing.T) {
	// Create mock backend
	backend := NewMockBackend(PixelFormatRGBA32, false)
	if backend == nil {
		t.Fatal("Failed to create mock backend")
	}

	// Test initial state
	if backend.GetCaption() != "Mock Window" {
		t.Errorf("Expected caption 'Mock Window', got '%s'", backend.GetCaption())
	}

	width, height := backend.GetWindowSize()
	if width != 0 || height != 0 {
		t.Errorf("Expected initial size (0,0), got (%d,%d)", width, height)
	}

	// Test initialization
	err := backend.Init(800, 600, WindowResize)
	if err != nil {
		t.Fatalf("Failed to initialize mock backend: %v", err)
	}

	width, height = backend.GetWindowSize()
	if width != 800 || height != 600 {
		t.Errorf("Expected size (800,600) after init, got (%d,%d)", width, height)
	}

	// Test caption setting
	backend.SetCaption("Test Window")
	if backend.GetCaption() != "Test Window" {
		t.Errorf("Expected caption 'Test Window', got '%s'", backend.GetCaption())
	}

	// Test window resizing
	err = backend.SetWindowSize(1024, 768)
	if err != nil {
		t.Errorf("Failed to resize window: %v", err)
	}

	width, height = backend.GetWindowSize()
	if width != 1024 || height != 768 {
		t.Errorf("Expected size (1024,768) after resize, got (%d,%d)", width, height)
	}

	// Test buffer operations
	testBuffer := buffer.NewRenderingBuffer[uint8]()
	data := make([]uint8, 1024*768*4) // RGBA32
	testBuffer.Attach(data, 1024, 768, 1024*4)

	err = backend.UpdateWindow(testBuffer)
	if err != nil {
		t.Errorf("Failed to update window: %v", err)
	}

	// Test image surface operations
	surface, err := backend.CreateImageSurface(100, 100)
	if err != nil {
		t.Errorf("Failed to create image surface: %v", err)
	}

	err = backend.DestroyImageSurface(surface)
	if err != nil {
		t.Errorf("Failed to destroy image surface: %v", err)
	}

	// Test timing functions
	ticks := backend.GetTicks()
	if ticks == 0 {
		t.Error("Expected non-zero ticks")
	}

	// Test image operations
	_, err = backend.LoadImage("test.bmp")
	if err != nil {
		t.Errorf("Failed to load image: %v", err)
	}

	err = backend.SaveImage(surface, "test.bmp")
	if err != nil {
		t.Errorf("Failed to save image: %v", err)
	}

	if backend.GetImageExtension() != ".bmp" {
		t.Errorf("Expected image extension '.bmp', got '%s'", backend.GetImageExtension())
	}

	// Test event polling
	if backend.PollEvents() != false {
		t.Error("Expected PollEvents to return false for mock backend (no events)")
	}

	if backend.WaitEvent() != false {
		t.Error("Expected WaitEvent to return false for mock backend (no events)")
	}

	// Test cleanup
	err = backend.Destroy()
	if err != nil {
		t.Errorf("Failed to destroy backend: %v", err)
	}
}

// TestBackendFactory tests the backend factory functionality
func TestBackendFactory(t *testing.T) {
	factory := GetBackendFactory()
	if factory == nil {
		t.Fatal("Failed to get backend factory")
	}

	// Test available backends
	backends := factory.GetAvailableBackends()
	if len(backends) == 0 {
		t.Error("Expected at least one available backend")
	}

	// Mock should always be available
	foundMock := false
	for _, backend := range backends {
		if backend == BackendMock {
			foundMock = true
			break
		}
	}
	if !foundMock {
		t.Error("Mock backend should always be available")
	}

	// Test default backend
	defaultBackend := factory.GetDefaultBackend()
	if defaultBackend < BackendMock || defaultBackend > BackendMacOS {
		t.Errorf("Invalid default backend type: %v", defaultBackend)
	}

	// Test creating backends
	mockBackend, err := factory.CreateBackend(BackendMock, PixelFormatRGBA32, false)
	if err != nil {
		t.Errorf("Failed to create mock backend: %v", err)
	}
	if mockBackend == nil {
		t.Error("Created backend is nil")
	}

	// Test creating unsupported backend (should fall back to mock)
	macBackend, err := factory.CreateBackend(BackendMacOS, PixelFormatRGBA32, false)
	if err != nil {
		t.Errorf("Failed to create fallback backend: %v", err)
	}
	if macBackend == nil {
		t.Error("Fallback backend is nil")
	}

	// Clean up
	if mockBackend != nil {
		mockBackend.Destroy()
	}
	if macBackend != nil {
		macBackend.Destroy()
	}
}

// TestBackendTypes tests the backend type enumeration
func TestBackendTypes(t *testing.T) {
	types := []BackendType{
		BackendMock, BackendX11, BackendSDL2, BackendWin32, BackendMacOS,
	}

	expectedNames := []string{
		"Mock", "X11", "SDL2", "Win32", "macOS",
	}

	for i, backendType := range types {
		name := backendType.String()
		if name != expectedNames[i] {
			t.Errorf("Expected backend type %d to be named '%s', got '%s'",
				backendType, expectedNames[i], name)
		}
	}

	// Test unknown backend type
	unknownType := BackendType(999)
	expected := "Unknown(999)"
	actual := unknownType.String()
	if actual != expected {
		t.Errorf("Expected unknown backend type to return '%s', got '%s'",
			expected, actual)
	}
}

// MockEventCallback implements EventCallback for testing
type MockEventCallback struct {
	initCalled     bool
	destroyCalled  bool
	resizeWidth    int
	resizeHeight   int
	idleCalled     bool
	mouseMoveX     int
	mouseMoveY     int
	mouseMoveFlags InputFlags
	mouseDownX     int
	mouseDownY     int
	mouseDownFlags InputFlags
	mouseUpX       int
	mouseUpY       int
	mouseUpFlags   InputFlags
	keyX           int
	keyY           int
	keyCode        KeyCode
	keyFlags       InputFlags
	drawCalled     bool
	postDrawCalled bool
}

func (m *MockEventCallback) OnInit() {
	m.initCalled = true
}

func (m *MockEventCallback) OnDestroy() {
	m.destroyCalled = true
}

func (m *MockEventCallback) OnResize(width, height int) {
	m.resizeWidth = width
	m.resizeHeight = height
}

func (m *MockEventCallback) OnIdle() {
	m.idleCalled = true
}

func (m *MockEventCallback) OnMouseMove(x, y int, flags InputFlags) {
	m.mouseMoveX = x
	m.mouseMoveY = y
	m.mouseMoveFlags = flags
}

func (m *MockEventCallback) OnMouseButtonDown(x, y int, flags InputFlags) {
	m.mouseDownX = x
	m.mouseDownY = y
	m.mouseDownFlags = flags
}

func (m *MockEventCallback) OnMouseButtonUp(x, y int, flags InputFlags) {
	m.mouseUpX = x
	m.mouseUpY = y
	m.mouseUpFlags = flags
}

func (m *MockEventCallback) OnKey(x, y int, key KeyCode, flags InputFlags) {
	m.keyX = x
	m.keyY = y
	m.keyCode = key
	m.keyFlags = flags
}

func (m *MockEventCallback) OnDraw() {
	m.drawCalled = true
}

func (m *MockEventCallback) OnPostDraw(rawHandler interface{}) {
	m.postDrawCalled = true
}

// TestMockBackendEventCallbacks tests event callback functionality
func TestMockBackendEventCallbacks(t *testing.T) {
	backend := NewMockBackend(PixelFormatRGBA32, false)
	callback := &MockEventCallback{}

	// Set callback
	backend.SetEventCallback(callback)

	// Test init callback
	err := backend.Init(800, 600, WindowResize)
	if err != nil {
		t.Fatalf("Failed to initialize backend: %v", err)
	}

	if !callback.initCalled {
		t.Error("OnInit callback was not called")
	}

	// Test resize callback
	err = backend.SetWindowSize(1024, 768)
	if err != nil {
		t.Errorf("Failed to resize window: %v", err)
	}

	if callback.resizeWidth != 1024 || callback.resizeHeight != 768 {
		t.Errorf("OnResize callback not called with correct dimensions: got (%d,%d)",
			callback.resizeWidth, callback.resizeHeight)
	}

	// Test force redraw (should call OnDraw)
	backend.ForceRedraw()
	if !callback.drawCalled {
		t.Error("OnDraw callback was not called by ForceRedraw")
	}

	// Test destroy callback
	err = backend.Destroy()
	if err != nil {
		t.Errorf("Failed to destroy backend: %v", err)
	}

	if !callback.destroyCalled {
		t.Error("OnDestroy callback was not called")
	}
}

// TestPixelFormatConversion tests basic pixel format handling
func TestPixelFormatConversion(t *testing.T) {
	formats := []PixelFormat{
		PixelFormatRGBA32, PixelFormatBGRA32, PixelFormatRGB24,
		PixelFormatBGR24, PixelFormatGray8, PixelFormatRGB565,
	}

	expectedBPP := []int{32, 32, 24, 24, 8, 16}

	for i, format := range formats {
		backend := NewMockBackend(format, false)

		err := backend.Init(100, 100, 0)
		if err != nil {
			t.Errorf("Failed to initialize backend with format %s: %v", format.String(), err)
			continue
		}

		// Create test buffer with appropriate size
		bpp := expectedBPP[i]
		stride := 100 * bpp / 8
		data := make([]uint8, stride*100)
		testBuffer := buffer.NewRenderingBuffer[uint8]()
		testBuffer.Attach(data, 100, 100, stride)

		// Test buffer update (should handle format conversion internally)
		err = backend.UpdateWindow(testBuffer)
		if err != nil {
			t.Errorf("Failed to update window with format %s: %v", format.String(), err)
		}

		backend.Destroy()
	}
}

// BenchmarkMockBackend benchmarks the mock backend performance
func BenchmarkMockBackend(b *testing.B) {
	backend := NewMockBackend(PixelFormatRGBA32, false)
	backend.Init(800, 600, 0)
	defer backend.Destroy()

	// Create test buffer
	data := make([]uint8, 800*600*4)
	testBuffer := buffer.NewRenderingBuffer[uint8]()
	testBuffer.Attach(data, 800, 600, 800*4)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		backend.UpdateWindow(testBuffer)
	}
}
