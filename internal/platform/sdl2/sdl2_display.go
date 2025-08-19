package sdl2

import (
	"fmt"

	"agg_go/internal/buffer"
	"agg_go/internal/platform/types"
	"github.com/veandco/go-sdl2/sdl"
)

// copyBufferToSurface copies the AGG rendering buffer to SDL2 surface, handling pixel format conversion
func (s *SDL2Backend) copyBufferToSurface(buffer *buffer.RenderingBuffer[uint8]) error {
	if buffer == nil || buffer.Buf() == nil {
		return fmt.Errorf("invalid buffer")
	}

	bufWidth := buffer.Width()
	bufHeight := buffer.Height()
	bufStride := buffer.Stride()

	// Ensure buffer dimensions match window dimensions
	if bufWidth != s.width || bufHeight != s.height {
		return fmt.Errorf("buffer dimensions (%dx%d) don't match window (%dx%d)",
			bufWidth, bufHeight, s.width, s.height)
	}

	// Lock surface for pixel access
	err := s.surface.Lock()
	if err != nil {
		return fmt.Errorf("failed to lock SDL surface: %w", err)
	}
	defer s.surface.Unlock()

	// Get surface pixel data
	surfacePixels := s.surface.Pixels()
	surfacePitch := s.surface.Pitch

	// Get source buffer data
	srcData := buffer.Buf()

	// Perform pixel format conversion based on AGG format
	switch s.format {
	case types.PixelFormatRGBA32:
		return s.copyRGBA32ToSurface(srcData, surfacePixels, bufStride, int(surfacePitch))
	case types.PixelFormatBGRA32:
		return s.copyBGRA32ToSurface(srcData, surfacePixels, bufStride, int(surfacePitch))
	case types.PixelFormatARGB32:
		return s.copyARGB32ToSurface(srcData, surfacePixels, bufStride, int(surfacePitch))
	case types.PixelFormatABGR32:
		return s.copyABGR32ToSurface(srcData, surfacePixels, bufStride, int(surfacePitch))
	case types.PixelFormatRGB24:
		return s.copyRGB24ToSurface(srcData, surfacePixels, bufStride, int(surfacePitch))
	case types.PixelFormatBGR24:
		return s.copyBGR24ToSurface(srcData, surfacePixels, bufStride, int(surfacePitch))
	case types.PixelFormatGray8:
		return s.copyGray8ToSurface(srcData, surfacePixels, bufStride, int(surfacePitch))
	case types.PixelFormatRGB565:
		return s.copyRGB565ToSurface(srcData, surfacePixels, bufStride, int(surfacePitch))
	case types.PixelFormatRGB555:
		return s.copyRGB555ToSurface(srcData, surfacePixels, bufStride, int(surfacePitch))
	default:
		// For unsupported formats, do a raw copy
		return s.copyRawToSurface(srcData, surfacePixels, bufStride, int(surfacePitch))
	}
}

// copyRGBA32ToSurface converts RGBA32 to SDL2 surface format
func (s *SDL2Backend) copyRGBA32ToSurface(src []byte, dst []byte, srcStride, dstStride int) error {
	for y := 0; y < s.height; y++ {
		srcRow := y * srcStride
		dstRow := y * dstStride

		for x := 0; x < s.width; x++ {
			srcPixel := srcRow + x*4
			dstPixel := dstRow + x*4

			if srcPixel+3 < len(src) && dstPixel+3 < len(dst) {
				// Copy RGBA directly (SDL2 RGBA32 format)
				dst[dstPixel+0] = src[srcPixel+0] // R
				dst[dstPixel+1] = src[srcPixel+1] // G
				dst[dstPixel+2] = src[srcPixel+2] // B
				dst[dstPixel+3] = src[srcPixel+3] // A
			}
		}
	}
	return nil
}

// copyBGRA32ToSurface converts BGRA32 to SDL2 surface format
func (s *SDL2Backend) copyBGRA32ToSurface(src []byte, dst []byte, srcStride, dstStride int) error {
	for y := 0; y < s.height; y++ {
		srcRow := y * srcStride
		dstRow := y * dstStride

		for x := 0; x < s.width; x++ {
			srcPixel := srcRow + x*4
			dstPixel := dstRow + x*4

			if srcPixel+3 < len(src) && dstPixel+3 < len(dst) {
				// Copy BGRA directly
				dst[dstPixel+0] = src[srcPixel+0] // B
				dst[dstPixel+1] = src[srcPixel+1] // G
				dst[dstPixel+2] = src[srcPixel+2] // R
				dst[dstPixel+3] = src[srcPixel+3] // A
			}
		}
	}
	return nil
}

// copyARGB32ToSurface converts ARGB32 to SDL2 surface format
func (s *SDL2Backend) copyARGB32ToSurface(src []byte, dst []byte, srcStride, dstStride int) error {
	for y := 0; y < s.height; y++ {
		srcRow := y * srcStride
		dstRow := y * dstStride

		for x := 0; x < s.width; x++ {
			srcPixel := srcRow + x*4
			dstPixel := dstRow + x*4

			if srcPixel+3 < len(src) && dstPixel+3 < len(dst) {
				// Convert ARGB to the surface format
				dst[dstPixel+0] = src[srcPixel+1] // R (from A-R-G-B)
				dst[dstPixel+1] = src[srcPixel+2] // G
				dst[dstPixel+2] = src[srcPixel+3] // B
				dst[dstPixel+3] = src[srcPixel+0] // A
			}
		}
	}
	return nil
}

// copyABGR32ToSurface converts ABGR32 to SDL2 surface format
func (s *SDL2Backend) copyABGR32ToSurface(src []byte, dst []byte, srcStride, dstStride int) error {
	for y := 0; y < s.height; y++ {
		srcRow := y * srcStride
		dstRow := y * dstStride

		for x := 0; x < s.width; x++ {
			srcPixel := srcRow + x*4
			dstPixel := dstRow + x*4

			if srcPixel+3 < len(src) && dstPixel+3 < len(dst) {
				// Convert ABGR to surface format
				dst[dstPixel+0] = src[srcPixel+3] // R (from A-B-G-R)
				dst[dstPixel+1] = src[srcPixel+2] // G
				dst[dstPixel+2] = src[srcPixel+1] // B
				dst[dstPixel+3] = src[srcPixel+0] // A
			}
		}
	}
	return nil
}

// copyRGB24ToSurface converts RGB24 to SDL2 surface format
func (s *SDL2Backend) copyRGB24ToSurface(src []byte, dst []byte, srcStride, dstStride int) error {
	for y := 0; y < s.height; y++ {
		srcRow := y * srcStride
		dstRow := y * dstStride

		for x := 0; x < s.width; x++ {
			srcPixel := srcRow + x*3
			dstPixel := dstRow + x*3

			if srcPixel+2 < len(src) && dstPixel+2 < len(dst) {
				// Copy RGB directly
				dst[dstPixel+0] = src[srcPixel+0] // R
				dst[dstPixel+1] = src[srcPixel+1] // G
				dst[dstPixel+2] = src[srcPixel+2] // B
			}
		}
	}
	return nil
}

// copyBGR24ToSurface converts BGR24 to SDL2 surface format
func (s *SDL2Backend) copyBGR24ToSurface(src []byte, dst []byte, srcStride, dstStride int) error {
	for y := 0; y < s.height; y++ {
		srcRow := y * srcStride
		dstRow := y * dstStride

		for x := 0; x < s.width; x++ {
			srcPixel := srcRow + x*3
			dstPixel := dstRow + x*3

			if srcPixel+2 < len(src) && dstPixel+2 < len(dst) {
				// Copy BGR directly
				dst[dstPixel+0] = src[srcPixel+0] // B
				dst[dstPixel+1] = src[srcPixel+1] // G
				dst[dstPixel+2] = src[srcPixel+2] // R
			}
		}
	}
	return nil
}

// copyGray8ToSurface converts Gray8 to SDL2 surface format
func (s *SDL2Backend) copyGray8ToSurface(src []byte, dst []byte, srcStride, dstStride int) error {
	// Convert grayscale to RGB (3 bytes per pixel in surface)
	for y := 0; y < s.height; y++ {
		srcRow := y * srcStride
		dstRow := y * dstStride

		for x := 0; x < s.width; x++ {
			srcPixel := srcRow + x
			dstPixel := dstRow + x*3

			if srcPixel < len(src) && dstPixel+2 < len(dst) {
				gray := src[srcPixel]
				dst[dstPixel+0] = gray // R
				dst[dstPixel+1] = gray // G
				dst[dstPixel+2] = gray // B
			}
		}
	}
	return nil
}

// copyRGB565ToSurface converts RGB565 to SDL2 surface format
func (s *SDL2Backend) copyRGB565ToSurface(src []byte, dst []byte, srcStride, dstStride int) error {
	for y := 0; y < s.height; y++ {
		srcRow := y * srcStride
		dstRow := y * dstStride

		for x := 0; x < s.width; x++ {
			srcPixel := srcRow + x*2
			dstPixel := dstRow + x*2

			if srcPixel+1 < len(src) && dstPixel+1 < len(dst) {
				// Copy RGB565 directly (16-bit)
				dst[dstPixel+0] = src[srcPixel+0]
				dst[dstPixel+1] = src[srcPixel+1]
			}
		}
	}
	return nil
}

// copyRGB555ToSurface converts RGB555 to SDL2 surface format
func (s *SDL2Backend) copyRGB555ToSurface(src []byte, dst []byte, srcStride, dstStride int) error {
	for y := 0; y < s.height; y++ {
		srcRow := y * srcStride
		dstRow := y * dstStride

		for x := 0; x < s.width; x++ {
			srcPixel := srcRow + x*2
			dstPixel := dstRow + x*2

			if srcPixel+1 < len(src) && dstPixel+1 < len(dst) {
				// Copy RGB555 directly (16-bit)
				dst[dstPixel+0] = src[srcPixel+0]
				dst[dstPixel+1] = src[srcPixel+1]
			}
		}
	}
	return nil
}

// copyRawToSurface does a raw memory copy (fallback for unsupported formats)
func (s *SDL2Backend) copyRawToSurface(src []byte, dst []byte, srcStride, dstStride int) error {
	minStride := srcStride
	if dstStride < srcStride {
		minStride = dstStride
	}

	for y := 0; y < s.height; y++ {
		srcRow := y * srcStride
		dstRow := y * dstStride

		if srcRow+minStride <= len(src) && dstRow+minStride <= len(dst) {
			copy(dst[dstRow:dstRow+minStride], src[srcRow:srcRow+minStride])
		}
	}
	return nil
}

// CreateImageSurface creates an SDL2 image surface
func (s *SDL2Backend) CreateImageSurface(width, height int) (interface{}, error) {
	surface, err := sdl.CreateRGBSurface(
		0, int32(width), int32(height), int32(s.bpp),
		s.rmask, s.gmask, s.bmask, s.amask)
	if err != nil {
		return nil, fmt.Errorf("failed to create SDL2 surface: %w", err)
	}

	// Find an empty slot in imageSurfaces array
	for i := range s.imageSurfaces {
		if s.imageSurfaces[i] == nil {
			s.imageSurfaces[i] = surface
			return &SDL2ImageSurface{
				surface: surface,
				index:   i,
			}, nil
		}
	}

	// If no slots available, still return the surface but don't track it
	return &SDL2ImageSurface{
		surface: surface,
		index:   -1,
	}, nil
}

// DestroyImageSurface destroys an SDL2 image surface
func (s *SDL2Backend) DestroyImageSurface(surface interface{}) error {
	if imageSurface, ok := surface.(*SDL2ImageSurface); ok {
		if imageSurface.index >= 0 && imageSurface.index < len(s.imageSurfaces) {
			s.imageSurfaces[imageSurface.index] = nil
		}
		if imageSurface.surface != nil {
			imageSurface.surface.Free()
		}
	}
	return nil
}

// GetTicks returns the current tick count
func (s *SDL2Backend) GetTicks() uint32 {
	return sdl.GetTicks()
}

// Delay provides a delay using SDL2
func (s *SDL2Backend) Delay(ms uint32) {
	sdl.Delay(ms)
}

// LoadImage loads an image file using SDL2_image
func (s *SDL2Backend) LoadImage(filename string) (interface{}, error) {
	// Note: This requires SDL2_image library
	// For now, we'll create a basic BMP loader or return a placeholder

	// Try to load as BMP (SDL2 built-in support)
	surface, err := sdl.LoadBMP(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to load BMP: %w", err)
	}
	if surface == nil {
		// Create a placeholder surface if loading fails
		return s.CreateImageSurface(100, 100)
	}

	// Convert to our pixel format if needed
	convertedSurface, err := surface.Convert(s.surface.Format, 0)
	surface.Free() // Free the original surface

	if err != nil {
		return nil, fmt.Errorf("failed to convert loaded image: %w", err)
	}

	return &SDL2ImageSurface{
		surface: convertedSurface,
		index:   -1, // Not tracked in array
	}, nil
}

// SaveImage saves an image to file
func (s *SDL2Backend) SaveImage(surface interface{}, filename string) error {
	if imageSurface, ok := surface.(*SDL2ImageSurface); ok && imageSurface.surface != nil {
		return imageSurface.surface.SaveBMP(filename)
	}
	return fmt.Errorf("invalid surface for saving")
}

// GetImageExtension returns the preferred image extension
func (s *SDL2Backend) GetImageExtension() string {
	return ".bmp"
}

// GetNativeHandle returns the native SDL2 handles
func (s *SDL2Backend) GetNativeHandle() interface{} {
	return &SDL2NativeHandle{
		window:   s.window,
		renderer: s.renderer,
		texture:  s.texture,
		surface:  s.surface,
	}
}

// SDL2ImageSurface represents an image surface in SDL2
type SDL2ImageSurface struct {
	surface *sdl.Surface
	index   int // Index in the imageSurfaces array, or -1 if not tracked
}

// SDL2NativeHandle provides access to native SDL2 handles
type SDL2NativeHandle struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
	surface  *sdl.Surface
}
