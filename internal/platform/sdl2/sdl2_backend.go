package sdl2

import (
	"fmt"
	"unsafe"

	"agg_go/internal/buffer"
	"agg_go/internal/platform/types"
	"github.com/veandco/go-sdl2/sdl"
)

// SDL2Backend implements PlatformBackend for SDL2
type SDL2Backend struct {
	// SDL objects
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
	surface  *sdl.Surface

	// Window properties
	caption string
	width   int
	height  int
	format  types.PixelFormat
	flipY   bool
	bpp     int

	// SDL format information
	pixelFormat uint32
	rmask       uint32
	gmask       uint32
	bmask       uint32
	amask       uint32

	// Event handling
	eventCallback types.EventCallback

	// State flags
	initialized bool
	shouldClose bool
	startTicks  uint32

	// Image surfaces for the max_images functionality
	imageSurfaces [16]*sdl.Surface
}

// NewSDL2BackendImpl creates a new SDL2 backend implementation
func NewSDL2BackendImpl(format types.PixelFormat, flipY bool) (*SDL2Backend, error) {
	backend := &SDL2Backend{
		caption: "AGG SDL2 Window",
		format:  format,
		flipY:   flipY,
		bpp:     format.BPP(),
	}

	// Initialize SDL2 pixel format settings
	err := backend.initPixelFormat()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize pixel format: %w", err)
	}

	return backend, nil
}

// Init initializes the SDL2 backend
func (s *SDL2Backend) Init(width, height int, flags types.WindowFlags) error {
	if s.initialized {
		return fmt.Errorf("SDL2 backend already initialized")
	}

	s.width = width
	s.height = height

	// Initialize SDL2
	err := sdl.Init(sdl.INIT_VIDEO)
	if err != nil {
		return fmt.Errorf("failed to initialize SDL2: %w", err)
	}

	// Convert AGG window flags to SDL2 flags
	windowFlags := s.convertWindowFlags(flags)

	// Create window
	s.window, err = sdl.CreateWindow(
		s.caption,
		sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED,
		int32(width), int32(height),
		windowFlags)
	if err != nil {
		s.cleanup()
		return fmt.Errorf("failed to create SDL2 window: %w", err)
	}

	// Create renderer
	s.renderer, err = sdl.CreateRenderer(s.window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		// Fall back to software rendering
		s.renderer, err = sdl.CreateRenderer(s.window, -1, sdl.RENDERER_SOFTWARE)
		if err != nil {
			s.cleanup()
			return fmt.Errorf("failed to create SDL2 renderer: %w", err)
		}
	}

	// Create texture for the rendering buffer
	s.texture, err = s.renderer.CreateTexture(
		s.pixelFormat,
		sdl.TEXTUREACCESS_STREAMING,
		int32(width), int32(height))
	if err != nil {
		s.cleanup()
		return fmt.Errorf("failed to create SDL2 texture: %w", err)
	}

	// Create surface for CPU-side rendering
	s.surface, err = sdl.CreateRGBSurface(
		0, int32(width), int32(height), int32(s.bpp),
		s.rmask, s.gmask, s.bmask, s.amask)
	if err != nil {
		s.cleanup()
		return fmt.Errorf("failed to create SDL2 surface: %w", err)
	}

	s.initialized = true
	s.startTicks = sdl.GetTicks()

	// Trigger init callback
	if s.eventCallback != nil {
		s.eventCallback.OnInit()
	}

	return nil
}

// initPixelFormat initializes SDL2 pixel format settings based on AGG pixel format
func (s *SDL2Backend) initPixelFormat() error {
	switch s.format {
	case types.PixelFormatRGB24:
		s.pixelFormat = sdl.PIXELFORMAT_RGB24
		s.rmask = 0xFF0000
		s.gmask = 0x00FF00
		s.bmask = 0x0000FF
		s.amask = 0
		s.bpp = 24

	case types.PixelFormatBGR24:
		s.pixelFormat = sdl.PIXELFORMAT_BGR24
		s.rmask = 0x0000FF
		s.gmask = 0x00FF00
		s.bmask = 0xFF0000
		s.amask = 0

	case types.PixelFormatRGBA32:
		s.pixelFormat = uint32(sdl.PIXELFORMAT_RGBA32)
		s.rmask = 0x000000FF
		s.gmask = 0x0000FF00
		s.bmask = 0x00FF0000
		s.amask = 0xFF000000
		s.bpp = 32

	case types.PixelFormatBGRA32:
		s.pixelFormat = uint32(sdl.PIXELFORMAT_BGRA32)
		s.rmask = 0x00FF0000
		s.gmask = 0x0000FF00
		s.bmask = 0x000000FF
		s.amask = 0xFF000000
		s.bpp = 32

	case types.PixelFormatARGB32:
		s.pixelFormat = uint32(sdl.PIXELFORMAT_ARGB32)
		s.rmask = 0x0000FF00
		s.gmask = 0x00FF0000
		s.bmask = 0xFF000000
		s.amask = 0x000000FF
		s.bpp = 32

	case types.PixelFormatABGR32:
		s.pixelFormat = uint32(sdl.PIXELFORMAT_ABGR32)
		s.rmask = 0xFF000000
		s.gmask = 0x00FF0000
		s.bmask = 0x0000FF00
		s.amask = 0x000000FF
		s.bpp = 32

	case types.PixelFormatGray8:
		// SDL2 doesn't have native grayscale, so we'll use RGB24 and convert
		s.pixelFormat = sdl.PIXELFORMAT_RGB24
		s.rmask = 0xFF0000
		s.gmask = 0x00FF00
		s.bmask = 0x0000FF
		s.amask = 0
		s.bpp = 8 // Keep original BPP for conversion

	case types.PixelFormatRGB565:
		s.pixelFormat = sdl.PIXELFORMAT_RGB565
		s.rmask = 0xF800
		s.gmask = 0x07E0
		s.bmask = 0x001F
		s.amask = 0
		s.bpp = 16

	case types.PixelFormatRGB555:
		s.pixelFormat = sdl.PIXELFORMAT_RGB555
		s.rmask = 0x7C00
		s.gmask = 0x03E0
		s.bmask = 0x001F
		s.amask = 0
		s.bpp = 16

	default:
		// Default to RGBA32 for unknown formats
		s.pixelFormat = uint32(sdl.PIXELFORMAT_RGBA32)
		s.rmask = 0x000000FF
		s.gmask = 0x0000FF00
		s.bmask = 0x00FF0000
		s.amask = 0xFF000000
		s.bpp = 32
	}

	return nil
}

// convertWindowFlags converts AGG window flags to SDL2 window flags
func (s *SDL2Backend) convertWindowFlags(flags types.WindowFlags) uint32 {
	var sdlFlags uint32 = sdl.WINDOW_SHOWN

	if flags&types.WindowResize != 0 {
		sdlFlags |= sdl.WINDOW_RESIZABLE
	}

	// Note: SDL2 doesn't have direct equivalent for HW_BUFFER or KEEP_ASPECT_RATIO
	// These would be handled differently in SDL2

	return sdlFlags
}

// Destroy cleans up SDL2 resources
func (s *SDL2Backend) Destroy() error {
	if !s.initialized {
		return nil
	}

	if s.eventCallback != nil {
		s.eventCallback.OnDestroy()
	}

	s.cleanup()
	s.initialized = false
	return nil
}

// cleanup performs the actual resource cleanup
func (s *SDL2Backend) cleanup() {
	// Clean up image surfaces
	for i := range s.imageSurfaces {
		if s.imageSurfaces[i] != nil {
			s.imageSurfaces[i].Free()
			s.imageSurfaces[i] = nil
		}
	}

	if s.surface != nil {
		s.surface.Free()
		s.surface = nil
	}

	if s.texture != nil {
		s.texture.Destroy()
		s.texture = nil
	}

	if s.renderer != nil {
		s.renderer.Destroy()
		s.renderer = nil
	}

	if s.window != nil {
		s.window.Destroy()
		s.window = nil
	}

	sdl.Quit()
}

// Run starts the SDL2 event loop
func (s *SDL2Backend) Run() int {
	if !s.initialized {
		return 1
	}

	for !s.shouldClose {
		if !s.PollEvents() {
			break
		}
	}

	return 0
}

// SetCaption sets the window caption
func (s *SDL2Backend) SetCaption(caption string) {
	s.caption = caption
	if s.initialized && s.window != nil {
		s.window.SetTitle(caption)
	}
}

// GetCaption returns the window caption
func (s *SDL2Backend) GetCaption() string {
	return s.caption
}

// SetWindowSize sets the window size
func (s *SDL2Backend) SetWindowSize(width, height int) error {
	if !s.initialized {
		return fmt.Errorf("SDL2 backend not initialized")
	}

	oldWidth, oldHeight := s.width, s.height
	s.width = width
	s.height = height

	// Resize window
	s.window.SetSize(int32(width), int32(height))

	// Recreate texture and surface for new size
	if s.texture != nil {
		s.texture.Destroy()
	}

	var err error
	s.texture, err = s.renderer.CreateTexture(
		s.pixelFormat,
		sdl.TEXTUREACCESS_STREAMING,
		int32(width), int32(height))
	if err != nil {
		return fmt.Errorf("failed to recreate texture: %w", err)
	}

	if s.surface != nil {
		s.surface.Free()
	}

	s.surface, err = sdl.CreateRGBSurface(
		0, int32(width), int32(height), int32(s.bpp),
		s.rmask, s.gmask, s.bmask, s.amask)
	if err != nil {
		return fmt.Errorf("failed to recreate surface: %w", err)
	}

	// Trigger resize callback
	if s.eventCallback != nil && (width != oldWidth || height != oldHeight) {
		s.eventCallback.OnResize(width, height)
	}

	return nil
}

// GetWindowSize returns the current window size
func (s *SDL2Backend) GetWindowSize() (width, height int) {
	return s.width, s.height
}

// UpdateWindow updates the window display with the rendering buffer
func (s *SDL2Backend) UpdateWindow(buffer *buffer.RenderingBuffer[uint8]) error {
	if !s.initialized || s.surface == nil || s.texture == nil {
		return fmt.Errorf("SDL2 backend not properly initialized")
	}

	// Copy buffer data to SDL surface
	err := s.copyBufferToSurface(buffer)
	if err != nil {
		return fmt.Errorf("failed to copy buffer to surface: %w", err)
	}

	// Update texture from surface
	err = s.texture.Update(nil, unsafe.Pointer(&s.surface.Pixels()[0]), int(s.surface.Pitch))
	if err != nil {
		return fmt.Errorf("failed to update texture: %w", err)
	}

	// Clear renderer and copy texture
	s.renderer.Clear()
	s.renderer.Copy(s.texture, nil, nil)
	s.renderer.Present()

	return nil
}

// SetEventCallback sets the event callback handler
func (s *SDL2Backend) SetEventCallback(callback types.EventCallback) {
	s.eventCallback = callback
}
