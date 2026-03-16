// Package conv implements AGG's row-oriented pixel-format conversion helpers.
// It is the Go equivalent of agg_color_conv.h plus the rgb8/rgb16 conversion
// header families.
package conv

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
)

// RenderingBuffer is the minimal buffer contract required by row converters.
type RenderingBuffer interface {
	Width() int
	Height() int
	RowPtr(x, y, length int) []basics.Int8u
}

// CopyRowFunctor is the AGG-style row-conversion functor contract.
type CopyRowFunctor interface {
	CopyRow(dst, src []basics.Int8u, width int)
}

// ColorConv is the main AGG-style buffer-to-buffer row conversion entry point.
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

// ColorConvRow applies one row-conversion functor to one row.
func ColorConvRow(dst, src []basics.Int8u, width int, copyRowFunctor CopyRowFunctor) {
	copyRowFunctor.CopyRow(dst, src, width)
}

// ColorConvSame is the AGG fast path for identical source and destination
// formats.
type ColorConvSame struct {
	BPP int // Bytes per pixel
}

// NewColorConvSame creates a same-format row copier.
func NewColorConvSame(bpp int) *ColorConvSame {
	return &ColorConvSame{BPP: bpp}
}

// CopyRow performs a direct byte copy.
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

// PixelConverter is the per-pixel converter contract used by the generic row adapter.
type PixelConverter interface {
	ConvertPixel(dst, src []basics.Int8u)
}

// ConvRow is the generic row adapter around a per-pixel converter.
type ConvRow struct {
	PixelConverter PixelConverter
	DstPixWidth    int // Destination pixel width in bytes
	SrcPixWidth    int // Source pixel width in bytes
}

// NewConvRow creates a generic row converter.
func NewConvRow(pixelConverter PixelConverter, dstPixWidth, srcPixWidth int) *ConvRow {
	return &ConvRow{
		PixelConverter: pixelConverter,
		DstPixWidth:    dstPixWidth,
		SrcPixWidth:    srcPixWidth,
	}
}

// CopyRow converts one row pixel-by-pixel.
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

// Convert is the high-level entry point that selects same-format copy or
// generic row conversion.
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
