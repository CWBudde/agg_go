package platform

// Backend availability detection based on build tags

// isX11Available returns true if X11 backend is available (determined at build time)
func isX11Available() bool {
	return x11Available
}

// isSDL2Available returns true if SDL2 backend is available (determined at build time)
func isSDL2Available() bool {
	return sdl2Available
}

// NewX11Backend creates a new X11 backend (implemented in x11 build tag files)
func NewX11Backend(format PixelFormat, flipY bool) (PlatformBackend, error) {
	if !isX11Available() {
		return NewMockBackend(format, flipY), nil
	}
	return newX11Backend(format, flipY)
}

// NewSDL2Backend creates a new SDL2 backend (implemented in sdl2 build tag files)
func NewSDL2Backend(format PixelFormat, flipY bool) (PlatformBackend, error) {
	if !isSDL2Available() {
		return NewMockBackend(format, flipY), nil
	}
	return newSDL2Backend(format, flipY)
}

// Functions newX11Backend and newSDL2Backend are implemented in platform-specific files with build tags
