//go:build x11
// +build x11

package platform

import (
	"fmt"

	"agg_go/internal/platform/x11"
)

// Availability when X11 is built
const (
	x11Available  = true
	sdl2Available = false // Will be overridden if both are built
)

// newX11Backend creates a new X11 backend
func newX11Backend(format PixelFormat, flipY bool) (PlatformBackend, error) {
	backend, err := x11.NewX11BackendImpl(format, flipY)
	if err != nil {
		return nil, fmt.Errorf("failed to create X11 backend: %w", err)
	}
	return backend, nil
}

// newSDL2Backend stub for X11-only builds
func newSDL2Backend(format PixelFormat, flipY bool) (PlatformBackend, error) {
	return nil, fmt.Errorf("SDL2 backend not available in this build")
}
