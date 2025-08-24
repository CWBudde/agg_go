package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/pixfmt/blender"
)

// PixFmtRGBA64 represents a 64-bit RGBA pixel format (16-bit per channel)
type PixFmtRGBA64[B any] struct {
	buf    *buffer.RenderingBufferU8
	stride int
}

// NewPixFmtRGBA64 creates a new 64-bit RGBA pixel format
func NewPixFmtRGBA64[B any](buf *buffer.RenderingBufferU8) *PixFmtRGBA64[B] {
	return &PixFmtRGBA64[B]{
		buf:    buf,
		stride: 8, // 4 channels * 2 bytes per channel
	}
}

// Width returns the width of the pixel format
func (pf *PixFmtRGBA64[B]) Width() int {
	return pf.buf.Width()
}

// Height returns the height of the pixel format
func (pf *PixFmtRGBA64[B]) Height() int {
	return pf.buf.Height()
}

// RowData returns a pointer to the row data at the given y coordinate
func (pf *PixFmtRGBA64[B]) RowData(y int) []basics.Int8u {
	return pf.buf.Row(y)
}

// MakePix creates a pixel pointer at the given coordinates
func (pf *PixFmtRGBA64[B]) MakePix(x, y int) []basics.Int16u {
	row := pf.buf.Row(y)
	// Convert to 16-bit slice
	ptr := x * pf.stride
	result := make([]basics.Int16u, 4)
	for i := 0; i < 4; i++ {
		result[i] = basics.Int16u(row[ptr+i*2]) | (basics.Int16u(row[ptr+i*2+1]) << 8)
	}
	return result
}

// Pixel returns the color at the given coordinates
func (pf *PixFmtRGBA64[B]) Pixel(x, y int) color.RGBA16[color.Linear] {
	if x >= 0 && y >= 0 && x < pf.Width() && y < pf.Height() {
		pix := pf.MakePix(x, y)
		return color.RGBA16[color.Linear]{
			R: pix[0], G: pix[1], B: pix[2], A: pix[3],
		}
	}
	return color.RGBA16[color.Linear]{}
}

// CopyPixel copies a pixel without blending
func (pf *PixFmtRGBA64[B]) CopyPixel(x, y int, c color.RGBA16[color.Linear]) {
	if x >= 0 && y >= 0 && x < pf.Width() && y < pf.Height() {
		row := pf.buf.Row(y)
		ptr := x * pf.stride
		// Write 16-bit values as little-endian bytes
		row[ptr+0] = basics.Int8u(c.R)
		row[ptr+1] = basics.Int8u(c.R >> 8)
		row[ptr+2] = basics.Int8u(c.G)
		row[ptr+3] = basics.Int8u(c.G >> 8)
		row[ptr+4] = basics.Int8u(c.B)
		row[ptr+5] = basics.Int8u(c.B >> 8)
		row[ptr+6] = basics.Int8u(c.A)
		row[ptr+7] = basics.Int8u(c.A >> 8)
	}
}

// BlendPixel blends a single pixel using the blender
func (pf *PixFmtRGBA64[B]) BlendPixel(x, y int, c color.RGBA16[color.Linear], cover basics.Int8u) {
	if x >= 0 && y >= 0 && x < pf.Width() && y < pf.Height() && !c.IsTransparent() {
		row := pf.buf.Row(y)
		ptr := x * pf.stride

		// Convert cover from 8-bit to 16-bit
		cover16 := basics.Int16u(cover) | (basics.Int16u(cover) << 8)

		var bl B
		switch any(bl).(type) {
		case blender.BlenderRGBA16:
			blender.BlendRGBA16Pixel(row[ptr:], c, cover16, blender.BlenderRGBA16{})
		case blender.BlenderRGBA16Pre:
			blender.BlendRGBA16Pixel(row[ptr:], c, cover16, blender.BlenderRGBA16Pre{})
		case blender.BlenderRGBA16Plain:
			blender.BlendRGBA16Pixel(row[ptr:], c, cover16, blender.BlenderRGBA16Plain{})
		default:
			// Default to standard blending
			blender.BlendRGBA16Pixel(row[ptr:], c, cover16, blender.BlenderRGBA16{})
		}
	}
}

// CopyHline copies a horizontal line of pixels
func (pf *PixFmtRGBA64[B]) CopyHline(x, y, length int, c color.RGBA16[color.Linear]) {
	if y >= 0 && y < pf.Height() && length > 0 {
		x1 := max(0, x)
		x2 := min(x+length, pf.Width())
		if x1 < x2 {
			row := pf.buf.Row(y)
			for i := x1; i < x2; i++ {
				ptr := i * pf.stride
				row[ptr+0] = basics.Int8u(c.R)
				row[ptr+1] = basics.Int8u(c.R >> 8)
				row[ptr+2] = basics.Int8u(c.G)
				row[ptr+3] = basics.Int8u(c.G >> 8)
				row[ptr+4] = basics.Int8u(c.B)
				row[ptr+5] = basics.Int8u(c.B >> 8)
				row[ptr+6] = basics.Int8u(c.A)
				row[ptr+7] = basics.Int8u(c.A >> 8)
			}
		}
	}
}

// BlendHline blends a horizontal line of pixels
func (pf *PixFmtRGBA64[B]) BlendHline(x, y, length int, c color.RGBA16[color.Linear], cover basics.Int8u) {
	if y >= 0 && y < pf.Height() && length > 0 && !c.IsTransparent() {
		x1 := max(0, x)
		x2 := min(x+length, pf.Width())
		if x1 < x2 {
			row := pf.buf.Row(y)
			cover16 := basics.Int16u(cover) | (basics.Int16u(cover) << 8)

			var bl B
			switch any(bl).(type) {
			case blender.BlenderRGBA16:
				for i := x1; i < x2; i++ {
					blender.BlendRGBA16Pixel(row[i*pf.stride:], c, cover16, blender.BlenderRGBA16{})
				}
			case blender.BlenderRGBA16Pre:
				for i := x1; i < x2; i++ {
					blender.BlendRGBA16Pixel(row[i*pf.stride:], c, cover16, blender.BlenderRGBA16Pre{})
				}
			case blender.BlenderRGBA16Plain:
				for i := x1; i < x2; i++ {
					blender.BlendRGBA16Pixel(row[i*pf.stride:], c, cover16, blender.BlenderRGBA16Plain{})
				}
			default:
				for i := x1; i < x2; i++ {
					blender.BlendRGBA16Pixel(row[i*pf.stride:], c, cover16, blender.BlenderRGBA16{})
				}
			}
		}
	}
}

// BlendSolidHspan blends a horizontal span with varying coverage
func (pf *PixFmtRGBA64[B]) BlendSolidHspan(x, y, length int, c color.RGBA16[color.Linear], covers []basics.Int8u) {
	if y >= 0 && y < pf.Height() && length > 0 && !c.IsTransparent() {
		x1 := max(0, x)
		x2 := min(x+length, pf.Width())
		if x1 < x2 {
			row := pf.buf.Row(y)
			coverOffset := x1 - x

			var bl B
			switch any(bl).(type) {
			case blender.BlenderRGBA16:
				for i := x1; i < x2; i++ {
					if covers[coverOffset] > 0 {
						cover16 := basics.Int16u(covers[coverOffset]) | (basics.Int16u(covers[coverOffset]) << 8)
						blender.BlendRGBA16Pixel(row[i*pf.stride:], c, cover16, blender.BlenderRGBA16{})
					}
					coverOffset++
				}
			case blender.BlenderRGBA16Pre:
				for i := x1; i < x2; i++ {
					if covers[coverOffset] > 0 {
						cover16 := basics.Int16u(covers[coverOffset]) | (basics.Int16u(covers[coverOffset]) << 8)
						blender.BlendRGBA16Pixel(row[i*pf.stride:], c, cover16, blender.BlenderRGBA16Pre{})
					}
					coverOffset++
				}
			case blender.BlenderRGBA16Plain:
				for i := x1; i < x2; i++ {
					if covers[coverOffset] > 0 {
						cover16 := basics.Int16u(covers[coverOffset]) | (basics.Int16u(covers[coverOffset]) << 8)
						blender.BlendRGBA16Pixel(row[i*pf.stride:], c, cover16, blender.BlenderRGBA16Plain{})
					}
					coverOffset++
				}
			default:
				for i := x1; i < x2; i++ {
					if covers[coverOffset] > 0 {
						cover16 := basics.Int16u(covers[coverOffset]) | (basics.Int16u(covers[coverOffset]) << 8)
						blender.BlendRGBA16Pixel(row[i*pf.stride:], c, cover16, blender.BlenderRGBA16{})
					}
					coverOffset++
				}
			}
		}
	}
}

// RGBA16 Blender definitions

// Concrete RGBA64 pixel format types for different component orders
type (
	PixFmtRGBA64Linear = PixFmtRGBA64[blender.BlenderRGBA16]
	PixFmtRGBA64Pre    = PixFmtRGBA64[blender.BlenderRGBA16Pre]
	PixFmtRGBA64Plain  = PixFmtRGBA64[blender.BlenderRGBA16Plain]

	// Different byte orders (would need separate implementations)
	PixFmtARGB64Linear = PixFmtRGBA64[blender.BlenderRGBA16] // TODO: Implement ARGB order
	PixFmtABGR64Linear = PixFmtRGBA64[blender.BlenderRGBA16] // TODO: Implement ABGR order
	PixFmtBGRA64Linear = PixFmtRGBA64[blender.BlenderRGBA16] // TODO: Implement BGRA order
)

// Helper functions to create specific RGBA64 pixel formats
func NewPixFmtRGBA64Linear(buf *buffer.RenderingBufferU8) *PixFmtRGBA64Linear {
	return NewPixFmtRGBA64[blender.BlenderRGBA16](buf)
}

func NewPixFmtRGBA64Pre(buf *buffer.RenderingBufferU8) *PixFmtRGBA64Pre {
	return NewPixFmtRGBA64[blender.BlenderRGBA16Pre](buf)
}

func NewPixFmtRGBA64Plain(buf *buffer.RenderingBufferU8) *PixFmtRGBA64Plain {
	return NewPixFmtRGBA64[blender.BlenderRGBA16Plain](buf)
}

// Utility functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
