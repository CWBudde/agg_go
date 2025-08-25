package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/pixfmt/blender"
)

// PixFmtRGBA64 represents a 64-bit RGBA pixel format (16-bit per channel) with byte order support
type PixFmtRGBA64[B interface {
	BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int16u)
}, O any] struct {
	buf    *buffer.RenderingBufferU8
	stride int
}

// NewPixFmtRGBA64 creates a new 64-bit RGBA pixel format
func NewPixFmtRGBA64[B interface {
	BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int16u)
}, O any](buf *buffer.RenderingBufferU8) *PixFmtRGBA64[B, O] {
	return &PixFmtRGBA64[B, O]{
		buf:    buf,
		stride: 8, // 4 channels * 2 bytes per channel
	}
}

// Width returns the width of the pixel format
func (pf *PixFmtRGBA64[B, O]) Width() int {
	return pf.buf.Width()
}

// Height returns the height of the pixel format
func (pf *PixFmtRGBA64[B, O]) Height() int {
	return pf.buf.Height()
}

// RowData returns a pointer to the row data at the given y coordinate
func (pf *PixFmtRGBA64[B, O]) RowData(y int) []basics.Int8u {
	return pf.buf.Row(y)
}

// MakePix creates a pixel pointer at the given coordinates
func (pf *PixFmtRGBA64[B, O]) MakePix(x, y int) []basics.Int16u {
	row := pf.buf.Row(y)
	order := blender.GetColorOrder[O]()

	// Convert to 16-bit slice respecting byte order
	ptr := x * pf.stride
	result := make([]basics.Int16u, 4)

	// Read each channel in correct order
	result[0] = basics.Int16u(row[ptr+order.R*2]) | (basics.Int16u(row[ptr+order.R*2+1]) << 8) // R
	result[1] = basics.Int16u(row[ptr+order.G*2]) | (basics.Int16u(row[ptr+order.G*2+1]) << 8) // G
	result[2] = basics.Int16u(row[ptr+order.B*2]) | (basics.Int16u(row[ptr+order.B*2+1]) << 8) // B
	result[3] = basics.Int16u(row[ptr+order.A*2]) | (basics.Int16u(row[ptr+order.A*2+1]) << 8) // A

	return result
}

// Pixel returns the color at the given coordinates
func (pf *PixFmtRGBA64[B, O]) Pixel(x, y int) color.RGBA16[color.Linear] {
	if x >= 0 && y >= 0 && x < pf.Width() && y < pf.Height() {
		pix := pf.MakePix(x, y)
		return color.RGBA16[color.Linear]{
			R: pix[0], G: pix[1], B: pix[2], A: pix[3],
		}
	}
	return color.RGBA16[color.Linear]{}
}

// CopyPixel copies a pixel without blending
func (pf *PixFmtRGBA64[B, O]) CopyPixel(x, y int, c color.RGBA16[color.Linear]) {
	if x >= 0 && y >= 0 && x < pf.Width() && y < pf.Height() {
		row := pf.buf.Row(y)
		order := blender.GetColorOrder[O]()
		ptr := x * pf.stride

		// Write 16-bit values as little-endian bytes respecting byte order
		row[ptr+order.R*2] = basics.Int8u(c.R)
		row[ptr+order.R*2+1] = basics.Int8u(c.R >> 8)
		row[ptr+order.G*2] = basics.Int8u(c.G)
		row[ptr+order.G*2+1] = basics.Int8u(c.G >> 8)
		row[ptr+order.B*2] = basics.Int8u(c.B)
		row[ptr+order.B*2+1] = basics.Int8u(c.B >> 8)
		row[ptr+order.A*2] = basics.Int8u(c.A)
		row[ptr+order.A*2+1] = basics.Int8u(c.A >> 8)
	}
}

// BlendPixel blends a single pixel using the blender
func (pf *PixFmtRGBA64[B, O]) BlendPixel(x, y int, c color.RGBA16[color.Linear], cover basics.Int8u) {
	if x >= 0 && y >= 0 && x < pf.Width() && y < pf.Height() && !c.IsTransparent() {
		row := pf.buf.Row(y)
		ptr := x * pf.stride

		// Convert cover from 8-bit to 16-bit
		cover16 := basics.Int16u(cover) | (basics.Int16u(cover) << 8)

		// Use the generic blender interface
		var bl B
		blender.BlendRGBA16Pixel(row[ptr:], c, cover16, bl)
	}
}

// CopyHline copies a horizontal line of pixels
func (pf *PixFmtRGBA64[B, O]) CopyHline(x, y, length int, c color.RGBA16[color.Linear]) {
	if y >= 0 && y < pf.Height() && length > 0 {
		x1 := max(0, x)
		x2 := min(x+length, pf.Width())
		if x1 < x2 {
			row := pf.buf.Row(y)
			order := blender.GetColorOrder[O]()
			for i := x1; i < x2; i++ {
				ptr := i * pf.stride
				row[ptr+order.R*2] = basics.Int8u(c.R)
				row[ptr+order.R*2+1] = basics.Int8u(c.R >> 8)
				row[ptr+order.G*2] = basics.Int8u(c.G)
				row[ptr+order.G*2+1] = basics.Int8u(c.G >> 8)
				row[ptr+order.B*2] = basics.Int8u(c.B)
				row[ptr+order.B*2+1] = basics.Int8u(c.B >> 8)
				row[ptr+order.A*2] = basics.Int8u(c.A)
				row[ptr+order.A*2+1] = basics.Int8u(c.A >> 8)
			}
		}
	}
}

// BlendHline blends a horizontal line of pixels
func (pf *PixFmtRGBA64[B, O]) BlendHline(x, y, length int, c color.RGBA16[color.Linear], cover basics.Int8u) {
	if y >= 0 && y < pf.Height() && length > 0 && !c.IsTransparent() {
		x1 := max(0, x)
		x2 := min(x+length, pf.Width())
		if x1 < x2 {
			row := pf.buf.Row(y)
			cover16 := basics.Int16u(cover) | (basics.Int16u(cover) << 8)

			var bl B
			for i := x1; i < x2; i++ {
				blender.BlendRGBA16Pixel(row[i*pf.stride:], c, cover16, bl)
			}
		}
	}
}

// BlendSolidHspan blends a horizontal span with varying coverage
func (pf *PixFmtRGBA64[B, O]) BlendSolidHspan(x, y, length int, c color.RGBA16[color.Linear], covers []basics.Int8u) {
	if y >= 0 && y < pf.Height() && length > 0 && !c.IsTransparent() {
		x1 := max(0, x)
		x2 := min(x+length, pf.Width())
		if x1 < x2 {
			row := pf.buf.Row(y)
			coverOffset := x1 - x

			var bl B
			for i := x1; i < x2; i++ {
				if covers[coverOffset] > 0 {
					cover16 := basics.Int16u(covers[coverOffset]) | (basics.Int16u(covers[coverOffset]) << 8)
					blender.BlendRGBA16Pixel(row[i*pf.stride:], c, cover16, bl)
				}
				coverOffset++
			}
		}
	}
}

// RGBA16 Blender definitions

// Concrete RGBA64 pixel format types for different component orders
type (
	// RGBA order formats
	PixFmtRGBA64Linear = PixFmtRGBA64[blender.BlenderRGBA16Linear, blender.RGBAOrder]
	PixFmtRGBA64Pre    = PixFmtRGBA64[blender.BlenderRGBA16PreLinear, blender.RGBAOrder]
	PixFmtRGBA64Plain  = PixFmtRGBA64[blender.BlenderRGBA16PlainLinear, blender.RGBAOrder]

	// ARGB order formats
	PixFmtARGB64Linear = PixFmtRGBA64[blender.BlenderARGB16Linear, blender.ARGBOrder]
	PixFmtARGB64Pre    = PixFmtRGBA64[blender.BlenderARGB16PreLinear, blender.ARGBOrder]
	PixFmtARGB64Plain  = PixFmtRGBA64[blender.BlenderARGB16PlainLinear, blender.ARGBOrder]

	// ABGR order formats
	PixFmtABGR64Linear = PixFmtRGBA64[blender.BlenderABGR16Linear, blender.ABGROrder]
	PixFmtABGR64Pre    = PixFmtRGBA64[blender.BlenderABGR16PreLinear, blender.ABGROrder]
	PixFmtABGR64Plain  = PixFmtRGBA64[blender.BlenderABGR16PlainLinear, blender.ABGROrder]

	// BGRA order formats
	PixFmtBGRA64Linear = PixFmtRGBA64[blender.BlenderBGRA16Linear, blender.BGRAOrder]
	PixFmtBGRA64Pre    = PixFmtRGBA64[blender.BlenderBGRA16PreLinear, blender.BGRAOrder]
	PixFmtBGRA64Plain  = PixFmtRGBA64[blender.BlenderBGRA16PlainLinear, blender.BGRAOrder]
)

// Helper functions to create specific RGBA64 pixel formats

// RGBA order constructors
func NewPixFmtRGBA64Linear(buf *buffer.RenderingBufferU8) *PixFmtRGBA64Linear {
	return NewPixFmtRGBA64[blender.BlenderRGBA16Linear, blender.RGBAOrder](buf)
}

func NewPixFmtRGBA64Pre(buf *buffer.RenderingBufferU8) *PixFmtRGBA64Pre {
	return NewPixFmtRGBA64[blender.BlenderRGBA16PreLinear, blender.RGBAOrder](buf)
}

func NewPixFmtRGBA64Plain(buf *buffer.RenderingBufferU8) *PixFmtRGBA64Plain {
	return NewPixFmtRGBA64[blender.BlenderRGBA16PlainLinear, blender.RGBAOrder](buf)
}

// ARGB order constructors
func NewPixFmtARGB64Linear(buf *buffer.RenderingBufferU8) *PixFmtARGB64Linear {
	return NewPixFmtRGBA64[blender.BlenderARGB16Linear, blender.ARGBOrder](buf)
}

func NewPixFmtARGB64Pre(buf *buffer.RenderingBufferU8) *PixFmtARGB64Pre {
	return NewPixFmtRGBA64[blender.BlenderARGB16PreLinear, blender.ARGBOrder](buf)
}

func NewPixFmtARGB64Plain(buf *buffer.RenderingBufferU8) *PixFmtARGB64Plain {
	return NewPixFmtRGBA64[blender.BlenderARGB16PlainLinear, blender.ARGBOrder](buf)
}

// ABGR order constructors
func NewPixFmtABGR64Linear(buf *buffer.RenderingBufferU8) *PixFmtABGR64Linear {
	return NewPixFmtRGBA64[blender.BlenderABGR16Linear, blender.ABGROrder](buf)
}

func NewPixFmtABGR64Pre(buf *buffer.RenderingBufferU8) *PixFmtABGR64Pre {
	return NewPixFmtRGBA64[blender.BlenderABGR16PreLinear, blender.ABGROrder](buf)
}

func NewPixFmtABGR64Plain(buf *buffer.RenderingBufferU8) *PixFmtABGR64Plain {
	return NewPixFmtRGBA64[blender.BlenderABGR16PlainLinear, blender.ABGROrder](buf)
}

// BGRA order constructors
func NewPixFmtBGRA64Linear(buf *buffer.RenderingBufferU8) *PixFmtBGRA64Linear {
	return NewPixFmtRGBA64[blender.BlenderBGRA16Linear, blender.BGRAOrder](buf)
}

func NewPixFmtBGRA64Pre(buf *buffer.RenderingBufferU8) *PixFmtBGRA64Pre {
	return NewPixFmtRGBA64[blender.BlenderBGRA16PreLinear, blender.BGRAOrder](buf)
}

func NewPixFmtBGRA64Plain(buf *buffer.RenderingBufferU8) *PixFmtBGRA64Plain {
	return NewPixFmtRGBA64[blender.BlenderBGRA16PlainLinear, blender.BGRAOrder](buf)
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
