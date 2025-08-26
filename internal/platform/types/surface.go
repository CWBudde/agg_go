package types

// ImageSurface defines the interface for platform-specific image surfaces.
// This provides a common abstraction for image data across different backend implementations.
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

// NativeHandle provides a common interface for platform-specific native handles.
// Implementations represent underlying system resources (e.g., window handles).
type NativeHandle interface {
    // GetType returns a string identifying the type of native handle
    GetType() string

    // IsValid returns true if the handle is valid
    IsValid() bool
}

