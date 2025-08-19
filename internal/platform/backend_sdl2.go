//go:build sdl2
// +build sdl2

package platform

import (
	"fmt"

	"agg_go/internal/platform/sdl2"
)

// Availability when SDL2 is built
const (
	x11Available  = false // Will be overridden if both are built
	sdl2Available = true
)

// newX11Backend stub for SDL2-only builds
func newX11Backend(format PixelFormat, flipY bool) (PlatformBackend, error) {
	return nil, fmt.Errorf("X11 backend not available in this build")
}

// newSDL2Backend creates a new SDL2 backend
func newSDL2Backend(format PixelFormat, flipY bool) (PlatformBackend, error) {
	backend, err := sdl2.NewSDL2BackendImpl(format, flipY)
	if err != nil {
		return nil, fmt.Errorf("failed to create SDL2 backend: %w", err)
	}
	return backend, nil
}
