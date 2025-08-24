// Package platform provides interfaces for platform-specific functionality in AGG.
// This package defines interfaces for backend implementations, event handling, and platform abstraction.
package platform

import "agg_go/internal/platform/types"

// EventCallbackSetter defines the interface for platform backends that support event callbacks.
// This interface replaces the duck-typing pattern used in platform backend code.
type EventCallbackSetter interface {
	// SetEventCallback sets the callback function for platform events
	SetEventCallback(callback types.EventCallback)
}

// ImageSurface defines the interface for platform-specific image surfaces.
// This replaces the use of interface{} in image-related operations.
type ImageSurface interface {
	// GetWidth returns the width of the image surface
	GetWidth() int

	// GetHeight returns the height of the image surface
	GetHeight() int

	// GetData returns the raw pixel data (may be platform-specific format)
	GetData() []byte

	// IsValid returns true if the surface is valid and ready for use
	IsValid() bool
}

// NativeHandle defines the interface for platform-specific native handles.
// This replaces the use of interface{} in GetNativeHandle operations.
type NativeHandle interface {
	// GetType returns a string identifying the type of native handle
	GetType() string

	// IsValid returns true if the handle is valid
	IsValid() bool
}

// RawEventHandler defines the interface for handling platform-specific raw event handlers.
// This provides type safety for event handler parameters in OnPostDraw.
type RawEventHandler interface {
	// GetBackendType returns the type of backend this handler is for
	GetBackendType() string

	// IsValid returns true if the handler is valid
	IsValid() bool
}

// Backend defines the core interface for platform backends.
// This interface provides a common abstraction for different platform implementations
// such as SDL2, X11, or other windowing systems.
type Backend interface {
	// Initialize initializes the backend with the given parameters
	Initialize(width, height int, title string) error

	// Shutdown shuts down the backend and releases resources
	Shutdown() error

	// Update updates the backend (processes events, refreshes display, etc.)
	Update() error
}

// Compile-time interface checks
// These verify that platform backends implement the expected interfaces

// Note: Actual backend implementations would add these checks in their own files:
// var _ EventCallbackSetter = (*SDL2Backend)(nil)
// var _ EventCallbackSetter = (*X11Backend)(nil)
// var _ Backend = (*SDL2Backend)(nil)
// var _ Backend = (*X11Backend)(nil)
