// Package span provides gradient span generation functionality for AGG.
// This file implements image-based gradient functionality that samples colors from an image.
package span

import (
	"agg_go/internal/color"
)

// OneColorFunction is a simple color function that holds a single mutable color.
// This matches AGG's one_color_function template class.
type OneColorFunction[ColorT any] struct {
	color ColorT
}

// NewOneColorFunction creates a new single-color function.
func NewOneColorFunction[ColorT any]() *OneColorFunction[ColorT] {
	return &OneColorFunction[ColorT]{}
}

// Size always returns 1 for single color functions.
func (ocf *OneColorFunction[ColorT]) Size() int {
	return 1
}

// ColorAt returns the single color (index is ignored).
func (ocf *OneColorFunction[ColorT]) ColorAt(index int) ColorT {
	return ocf.color
}

// Color returns a pointer to the mutable color.
func (ocf *OneColorFunction[ColorT]) Color() *ColorT {
	return &ocf.color
}

// GradientImageRGBA8 provides image-based gradient functionality for RGBA8 colors.
// It samples colors from an image buffer using coordinate wrapping for tiling.
// This is a port of AGG's gradient_image template class, specialized for RGBA8.
type GradientImageRGBA8 struct {
	buffer      []color.RGBA8[color.SRGB]                  // Image buffer
	allocWidth  int                                        // Allocated buffer width
	allocHeight int                                        // Allocated buffer height
	width       int                                        // Current image width
	height      int                                        // Current image height
	colorFunc   *OneColorFunction[color.RGBA8[color.SRGB]] // Color function for results
}

// NewGradientImageRGBA8 creates a new image-based gradient for RGBA8 colors.
func NewGradientImageRGBA8() *GradientImageRGBA8 {
	return &GradientImageRGBA8{
		colorFunc: NewOneColorFunction[color.RGBA8[color.SRGB]](),
	}
}

// ImageCreate creates or resizes the image buffer.
// Returns the buffer slice if successful, nil if allocation failed.
func (gi *GradientImageRGBA8) ImageCreate(width, height int) []color.RGBA8[color.SRGB] {
	// Only reallocate if we need more space
	if width > gi.allocWidth || height > gi.allocHeight {
		// Allocate new buffer
		gi.buffer = make([]color.RGBA8[color.SRGB], width*height)
		if gi.buffer != nil {
			gi.allocWidth = width
			gi.allocHeight = height
		} else {
			gi.allocWidth = 0
			gi.allocHeight = 0
		}
	}

	if gi.buffer != nil {
		gi.width = width
		gi.height = height

		// Clear the buffer to transparent black
		for i := range gi.buffer {
			gi.buffer[i] = color.RGBA8[color.SRGB]{}
		}

		return gi.buffer
	}
	return nil
}

// ImageBuffer returns the current image buffer.
func (gi *GradientImageRGBA8) ImageBuffer() []color.RGBA8[color.SRGB] {
	return gi.buffer
}

// ImageWidth returns the current image width.
func (gi *GradientImageRGBA8) ImageWidth() int {
	return gi.width
}

// ImageHeight returns the current image height.
func (gi *GradientImageRGBA8) ImageHeight() int {
	return gi.height
}

// ImageStride returns the stride (bytes per row) of the image.
func (gi *GradientImageRGBA8) ImageStride() int {
	return gi.allocWidth * 4 // 4 bytes per RGBA pixel
}

// Calculate samples a color from the image at the given coordinates.
// This implements the GradientFunction interface.
// Coordinates are in gradient subpixel precision and use wrapping/tiling.
func (gi *GradientImageRGBA8) Calculate(x, y, d2 int) int {
	if gi.buffer != nil && gi.width > 0 && gi.height > 0 {
		// Convert from subpixel to pixel coordinates
		px := x >> GradientSubpixelShift
		py := y >> GradientSubpixelShift

		// Wrap coordinates for tiling behavior
		px = px % gi.width
		if px < 0 {
			px += gi.width
		}

		py = py % gi.height
		if py < 0 {
			py += gi.height
		}

		// Sample pixel from buffer (use allocated width for stride, like C++ AGG)
		pixel := gi.buffer[py*gi.allocWidth+px]

		// Update the color function's color directly (no type conversion needed)
		*gi.colorFunc.Color() = pixel
	} else {
		// No buffer available - set color to transparent black
		*gi.colorFunc.Color() = color.RGBA8[color.SRGB]{}
	}

	// Gradient functions typically return distance, but image gradients
	// use the color sampling side effect, so return 0
	return 0
}

// ColorFunction returns the color function used for gradient colors.
func (gi *GradientImageRGBA8) ColorFunction() *OneColorFunction[color.RGBA8[color.SRGB]] {
	return gi.colorFunc
}

// AllocWidth returns the allocated buffer width (for debugging).
func (gi *GradientImageRGBA8) AllocWidth() int {
	return gi.allocWidth
}

// AllocHeight returns the allocated buffer height (for debugging).
func (gi *GradientImageRGBA8) AllocHeight() int {
	return gi.allocHeight
}

// GradientImage provides image-based gradient functionality.
// It samples colors from an image buffer using coordinate wrapping for tiling.
// This is a port of AGG's gradient_image template class.
type GradientImage[ColorT any] struct {
	buffer      []color.RGBA8[color.SRGB] // Image buffer
	allocWidth  int                       // Allocated buffer width
	allocHeight int                       // Allocated buffer height
	width       int                       // Current image width
	height      int                       // Current image height
	colorFunc   *OneColorFunction[ColorT] // Color function for results
}

// NewGradientImage creates a new image-based gradient.
func NewGradientImage[ColorT any]() *GradientImage[ColorT] {
	return &GradientImage[ColorT]{
		colorFunc: NewOneColorFunction[ColorT](),
	}
}

// ImageCreate creates or resizes the image buffer.
// Returns the buffer slice if successful, nil if allocation failed.
func (gi *GradientImage[ColorT]) ImageCreate(width, height int) []color.RGBA8[color.SRGB] {
	// Only reallocate if we need more space
	if width > gi.allocWidth || height > gi.allocHeight {
		// Allocate new buffer
		gi.buffer = make([]color.RGBA8[color.SRGB], width*height)
		if gi.buffer != nil {
			gi.allocWidth = width
			gi.allocHeight = height
		} else {
			gi.allocWidth = 0
			gi.allocHeight = 0
		}
	}

	if gi.buffer != nil {
		gi.width = width
		gi.height = height

		// Clear the buffer to transparent black
		for i := range gi.buffer {
			gi.buffer[i] = color.RGBA8[color.SRGB]{}
		}

		return gi.buffer
	}
	return nil
}

// ImageBuffer returns the current image buffer.
func (gi *GradientImage[ColorT]) ImageBuffer() []color.RGBA8[color.SRGB] {
	return gi.buffer
}

// ImageWidth returns the current image width.
func (gi *GradientImage[ColorT]) ImageWidth() int {
	return gi.width
}

// ImageHeight returns the current image height.
func (gi *GradientImage[ColorT]) ImageHeight() int {
	return gi.height
}

// ImageStride returns the stride (bytes per row) of the image.
func (gi *GradientImage[ColorT]) ImageStride() int {
	return gi.allocWidth * 4 // 4 bytes per RGBA pixel
}

// Calculate samples a color from the image at the given coordinates.
// This implements the GradientFunction interface.
// Coordinates are in gradient subpixel precision and use wrapping/tiling.
func (gi *GradientImage[ColorT]) Calculate(x, y, d2 int) int {
	if gi.buffer != nil && gi.width > 0 && gi.height > 0 {
		// Convert from subpixel to pixel coordinates
		px := x >> GradientSubpixelShift
		py := y >> GradientSubpixelShift

		// Wrap coordinates for tiling behavior
		px = px % gi.width
		if px < 0 {
			px += gi.width
		}

		py = py % gi.height
		if py < 0 {
			py += gi.height
		}

		// Sample pixel from buffer (use allocated width for stride, like C++ AGG)
		pixel := gi.buffer[py*gi.allocWidth+px]

		// Update the color function's color based on the sampled pixel
		// Use type conversion through any to handle the generic ColorT
		colorPtr := gi.colorFunc.Color()
		*colorPtr = any(pixel).(ColorT)
	} else {
		// No buffer available - set color to transparent black
		colorPtr := gi.colorFunc.Color()
		*colorPtr = any(color.RGBA8[color.SRGB]{}).(ColorT)
	}

	// Gradient functions typically return distance, but image gradients
	// use the color sampling side effect, so return 0
	return 0
}

// ColorFunction returns the color function used for gradient colors.
func (gi *GradientImage[ColorT]) ColorFunction() *OneColorFunction[ColorT] {
	return gi.colorFunc
}
