//go:build x11 && sdl2
// +build x11,sdl2

package platform

import (
	"fmt"

	"agg_go/internal/platform/sdl2"
	"agg_go/internal/platform/x11"
)

// Availability when both X11 and SDL2 are built
const (
	x11Available  = true
	sdl2Available = true
)

// newX11Backend creates a new X11 backend
func newX11Backend(format PixelFormat, flipY bool) (PlatformBackend, error) {
	backend, err := x11.NewX11BackendImpl(format, flipY)
	if err != nil {
		return nil, fmt.Errorf("failed to create X11 backend: %w", err)
	}
	return backend, nil
}

// newSDL2Backend creates a new SDL2 backend
func newSDL2Backend(format PixelFormat, flipY bool) (PlatformBackend, error) {
	backend, err := sdl2.NewSDL2BackendImpl(format, flipY)
	if err != nil {
		return nil, fmt.Errorf("failed to create SDL2 backend: %w", err)
	}
	return backend, nil
}
