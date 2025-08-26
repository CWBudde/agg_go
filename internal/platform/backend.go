package platform

import (
	"fmt"

	"agg_go/internal/buffer"
	types "agg_go/internal/platform/types"
)

// PlatformBackend defines the interface that platform-specific implementations must satisfy.
// This allows the platform support system to work with different windowing systems
// like X11, SDL2, Win32, etc.
type PlatformBackend interface {
	// Window lifecycle
	Init(width, height int, flags WindowFlags) error
	Destroy() error
	Run() int

	// Window properties
	SetCaption(caption string)
	GetCaption() string
	SetWindowSize(width, height int) error
	GetWindowSize() (width, height int)

	// Buffer management
	UpdateWindow(buffer *buffer.RenderingBuffer[uint8]) error
	CreateImageSurface(width, height int) (types.ImageSurface, error)
	DestroyImageSurface(surface types.ImageSurface) error

	// Event handling
	PollEvents() bool
	WaitEvent() bool
	ForceRedraw()

	// Timing
	GetTicks() uint32
	Delay(ms uint32)

	// Image operations (platform-specific format support)
	LoadImage(filename string) (types.ImageSurface, error)
	SaveImage(surface types.ImageSurface, filename string) error
	GetImageExtension() string

	// Platform-specific data access (for advanced users)
	GetNativeHandle() types.NativeHandle
}

// BackendType represents the type of platform backend
type BackendType int

const (
	BackendMock BackendType = iota
	BackendX11
	BackendSDL2
	BackendWin32
	BackendMacOS
)

// String returns the string representation of the backend type
func (bt BackendType) String() string {
	switch bt {
	case BackendMock:
		return "Mock"
	case BackendX11:
		return "X11"
	case BackendSDL2:
		return "SDL2"
	case BackendWin32:
		return "Win32"
	case BackendMacOS:
		return "macOS"
	default:
		return fmt.Sprintf("Unknown(%d)", int(bt))
	}
}

// BackendFactory creates platform-specific backends based on build tags and runtime environment
type BackendFactory interface {
	CreateBackend(backendType BackendType, format PixelFormat, flipY bool) (PlatformBackend, error)
	GetAvailableBackends() []BackendType
	GetDefaultBackend() BackendType
}

// MockImageSurface implements types.ImageSurface for testing
type MockImageSurface struct {
	width  int
	height int
	data   []byte
}

func (m *MockImageSurface) GetWidth() int   { return m.width }
func (m *MockImageSurface) GetHeight() int  { return m.height }
func (m *MockImageSurface) GetData() []byte { return m.data }
func (m *MockImageSurface) IsValid() bool   { return m.data != nil }

// MockNativeHandle implements types.NativeHandle for testing
type MockNativeHandle struct {
	backendType string
	valid       bool
}

func (m *MockNativeHandle) GetType() string { return m.backendType }
func (m *MockNativeHandle) IsValid() bool   { return m.valid }

// MockBackend provides a basic implementation for testing and headless operation
type MockBackend struct {
	caption       string
	width         int
	height        int
	format        PixelFormat
	flipY         bool
	initialized   bool
	eventCallback EventCallback
	startTicks    uint32
}

// NewMockBackend creates a new mock backend for testing
func NewMockBackend(format PixelFormat, flipY bool) *MockBackend {
	return &MockBackend{
		caption:    "Mock Window",
		format:     format,
		flipY:      flipY,
		startTicks: 0, // Will be set on Init
	}
}

// SetEventCallback sets the event callback handler
func (m *MockBackend) SetEventCallback(callback EventCallback) {
	m.eventCallback = callback
}

// Init initializes the mock backend
func (m *MockBackend) Init(width, height int, flags WindowFlags) error {
	m.width = width
	m.height = height
	m.initialized = true
	m.startTicks = 1000 // Mock start time

	if m.eventCallback != nil {
		m.eventCallback.OnInit()
	}
	return nil
}

// Destroy cleans up the mock backend
func (m *MockBackend) Destroy() error {
	if m.eventCallback != nil {
		m.eventCallback.OnDestroy()
	}
	m.initialized = false
	return nil
}

// Run starts the mock event loop (just returns immediately)
func (m *MockBackend) Run() int {
	return 0
}

// SetCaption sets the window caption
func (m *MockBackend) SetCaption(caption string) {
	m.caption = caption
}

// GetCaption returns the window caption
func (m *MockBackend) GetCaption() string {
	return m.caption
}

// SetWindowSize sets the window size
func (m *MockBackend) SetWindowSize(width, height int) error {
	oldWidth, oldHeight := m.width, m.height
	m.width = width
	m.height = height

	if m.eventCallback != nil && (width != oldWidth || height != oldHeight) {
		m.eventCallback.OnResize(width, height)
	}
	return nil
}

// GetWindowSize returns the current window size
func (m *MockBackend) GetWindowSize() (width, height int) {
	return m.width, m.height
}

// UpdateWindow updates the window display (no-op for mock)
func (m *MockBackend) UpdateWindow(buffer *buffer.RenderingBuffer[uint8]) error {
	return nil
}

// CreateImageSurface creates a mock image surface
func (m *MockBackend) CreateImageSurface(width, height int) (types.ImageSurface, error) {
	// Create a mock surface with proper interface implementation
	data := make([]byte, width*height*4) // Assume RGBA32
	return &MockImageSurface{
		width:  width,
		height: height,
		data:   data,
	}, nil
}

// DestroyImageSurface destroys a mock image surface
func (m *MockBackend) DestroyImageSurface(surface types.ImageSurface) error {
	// Nothing to do for mock - Go GC will handle cleanup
	return nil
}

// PollEvents polls for events (mock implementation)
func (m *MockBackend) PollEvents() bool {
	// No events in mock implementation
	return false
}

// WaitEvent waits for events (mock implementation)
func (m *MockBackend) WaitEvent() bool {
	// No events in mock implementation
	return false
}

// ForceRedraw forces a window redraw
func (m *MockBackend) ForceRedraw() {
	if m.eventCallback != nil {
		m.eventCallback.OnDraw()
	}
}

// GetTicks returns mock ticks
func (m *MockBackend) GetTicks() uint32 {
	return m.startTicks + 1000 // Mock elapsed time
}

// Delay provides a mock delay
func (m *MockBackend) Delay(ms uint32) {
	// No actual delay in mock
}

// LoadImage loads a mock image
func (m *MockBackend) LoadImage(filename string) (types.ImageSurface, error) {
	// Return a mock 100x100 RGBA surface
	data := make([]byte, 100*100*4)
	return &MockImageSurface{
		width:  100,
		height: 100,
		data:   data,
	}, nil
}

// SaveImage saves a mock image
func (m *MockBackend) SaveImage(surface types.ImageSurface, filename string) error {
	return nil
}

// GetImageExtension returns the mock image extension
func (m *MockBackend) GetImageExtension() string {
	return ".bmp"
}

// GetNativeHandle returns the mock native handle
func (m *MockBackend) GetNativeHandle() types.NativeHandle {
	return &MockNativeHandle{
		backendType: "mock",
		valid:       true,
	}
}

// DefaultBackendFactory provides the default backend factory implementation
type DefaultBackendFactory struct{}

// CreateBackend creates a backend of the specified type
func (f *DefaultBackendFactory) CreateBackend(backendType BackendType, format PixelFormat, flipY bool) (PlatformBackend, error) {
	switch backendType {
	case BackendMock:
		return NewMockBackend(format, flipY), nil
	case BackendX11:
		return NewX11Backend(format, flipY)
	case BackendSDL2:
		return NewSDL2Backend(format, flipY)
	default:
		// Fall back to mock backend for unsupported types
		return NewMockBackend(format, flipY), nil
	}
}

// GetAvailableBackends returns the list of available backends
func (f *DefaultBackendFactory) GetAvailableBackends() []BackendType {
	backends := []BackendType{BackendMock}

	// Add backends based on build tags
	if isX11Available() {
		backends = append(backends, BackendX11)
	}
	if isSDL2Available() {
		backends = append(backends, BackendSDL2)
	}

	return backends
}

// GetDefaultBackend returns the default backend for the current platform
func (f *DefaultBackendFactory) GetDefaultBackend() BackendType {
	// Prefer SDL2 if available, then X11, finally mock
	if isSDL2Available() {
		return BackendSDL2
	}
	if isX11Available() {
		return BackendX11
	}
	return BackendMock
}

// Global backend factory instance
var backendFactory BackendFactory = &DefaultBackendFactory{}

// SetBackendFactory allows replacing the default backend factory
func SetBackendFactory(factory BackendFactory) {
	backendFactory = factory
}

// GetBackendFactory returns the current backend factory
func GetBackendFactory() BackendFactory {
	return backendFactory
}
