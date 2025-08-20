// Package conv provides color conversion utilities for AGG.
// This package implements color space and pixel format conversions
// corresponding to agg_color_conv.h from the original AGG library.
package conv

import (
	"agg_go/internal/basics"
)

// RenderingBuffer interface represents the minimum required interface
// for color conversion operations, compatible with existing buffer types.
// This matches the C++ AGG rendering_buffer interface.
type RenderingBuffer interface {
	Width() int
	Height() int
	RowPtr(x, y, length int) []basics.Int8u
}

// CopyRowFunctor defines the interface for row copying operations.
// Implementations of this interface handle the actual pixel format conversion.
type CopyRowFunctor interface {
	CopyRow(dst, src []basics.Int8u, width int)
}

// ColorConv performs color conversion between two rendering buffers using
// the provided copy row functor. This is the main conversion function.
// It corresponds to the template function color_conv() in AGG.
func ColorConv(dst, src RenderingBuffer, copyRowFunctor CopyRowFunctor) {
	width := src.Width()
	height := src.Height()

	// Limit conversion to the smaller buffer dimensions
	if dst.Width() < width {
		width = dst.Width()
	}
	if dst.Height() < height {
		height = dst.Height()
	}

	if width > 0 && height > 0 {
		for y := 0; y < height; y++ {
			// Match C++ AGG buffer access pattern exactly
			// C++ uses: copy_row_functor(dst->row_ptr(0, y, width), src->row_ptr(y), width)
			dstRow := dst.RowPtr(0, y, width)
			srcRow := src.RowPtr(0, y, width) // Use consistent interface
			if dstRow != nil && srcRow != nil {
				copyRowFunctor.CopyRow(dstRow, srcRow, width)
			}
		}
	}
}

// ColorConvRow performs color conversion on a single row of pixels.
// This function is useful for converting individual scanlines.
func ColorConvRow(dst, src []basics.Int8u, width int, copyRowFunctor CopyRowFunctor) {
	copyRowFunctor.CopyRow(dst, src, width)
}

// ColorConvSame implements a copy row functor that performs a direct memory copy.
// This is used when source and destination formats are identical.
// The BPP parameter specifies bytes per pixel.
type ColorConvSame struct {
	BPP int // Bytes per pixel
}

// NewColorConvSame creates a new same-format copy functor.
func NewColorConvSame(bpp int) *ColorConvSame {
	return &ColorConvSame{BPP: bpp}
}

// CopyRow implements direct memory copy for identical pixel formats.
func (c *ColorConvSame) CopyRow(dst, src []basics.Int8u, width int) {
	if width <= 0 {
		return
	}

	bytesToCopy := width * c.BPP
	if len(dst) < bytesToCopy || len(src) < bytesToCopy {
		return
	}

	copy(dst[:bytesToCopy], src[:bytesToCopy])
}

// Generic pixel converter interface for format-specific conversions.
// This replaces the C++ template conv_pixel<DstFormat, SrcFormat>.
type PixelConverter interface {
	ConvertPixel(dst, src []basics.Int8u)
}

// Generic row converter that uses a pixel converter for individual pixels.
// This replaces the C++ template conv_row<DstFormat, SrcFormat>.
type ConvRow struct {
	PixelConverter PixelConverter
	DstPixWidth    int // Destination pixel width in bytes
	SrcPixWidth    int // Source pixel width in bytes
}

// NewConvRow creates a new generic row converter.
func NewConvRow(pixelConverter PixelConverter, dstPixWidth, srcPixWidth int) *ConvRow {
	return &ConvRow{
		PixelConverter: pixelConverter,
		DstPixWidth:    dstPixWidth,
		SrcPixWidth:    srcPixWidth,
	}
}

// CopyRow converts a row of pixels using the pixel converter.
func (c *ConvRow) CopyRow(dst, src []basics.Int8u, width int) {
	if width <= 0 || c.PixelConverter == nil {
		return
	}

	dstOffset := 0
	srcOffset := 0

	for i := 0; i < width; i++ {
		if dstOffset+c.DstPixWidth <= len(dst) && srcOffset+c.SrcPixWidth <= len(src) {
			c.PixelConverter.ConvertPixel(
				dst[dstOffset:dstOffset+c.DstPixWidth],
				src[srcOffset:srcOffset+c.SrcPixWidth],
			)
		}
		dstOffset += c.DstPixWidth
		srcOffset += c.SrcPixWidth
	}
}

// Convert provides a high-level conversion function that automatically
// creates the appropriate row converter and performs the conversion.
// This corresponds to the template function convert() in AGG.
func Convert(dst, src RenderingBuffer, dstPixWidth, srcPixWidth int, pixelConverter PixelConverter) {
	// Use ColorConvSame for identical formats
	if dstPixWidth == srcPixWidth && pixelConverter == nil {
		copyFunctor := NewColorConvSame(dstPixWidth)
		ColorConv(dst, src, copyFunctor)
		return
	}

	// Use generic row converter for different formats
	if pixelConverter != nil {
		rowConverter := NewConvRow(pixelConverter, dstPixWidth, srcPixWidth)
		ColorConv(dst, src, rowConverter)
	}
}
