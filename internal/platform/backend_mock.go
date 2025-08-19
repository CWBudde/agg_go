//go:build !x11 && !sdl2
// +build !x11,!sdl2

package platform

import "fmt"

// Default availability when no specific backends are built
const (
	x11Available  = false
	sdl2Available = false
)

// Default stub implementations that return mock backend
func newX11Backend(format PixelFormat, flipY bool) (PlatformBackend, error) {
	return nil, fmt.Errorf("X11 backend not available in this build")
}

func newSDL2Backend(format PixelFormat, flipY bool) (PlatformBackend, error) {
	return nil, fmt.Errorf("SDL2 backend not available in this build")
}
