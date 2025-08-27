package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/order"
	"agg_go/internal/pixfmt/blender"
)

// ==============================================================================
// RGB48 (16-bit per channel) Pixel Formats
// ==============================================================================

// PixFmtAlphaBlendRGB48 represents RGB pixel format with 16-bit components (6 bytes per pixel)
type PixFmtAlphaBlendRGB48[
	B blender.RGB48Blender[S, O],
	S color.Space,
	O order.RGBOrder,
] struct {
	rbuf     *buffer.RenderingBufferU16
	blender  B
	category PixFmtRGBTag
}

// NewPixFmtAlphaBlendRGB48 creates a new RGB48 pixel format
func NewPixFmtAlphaBlendRGB48[
	B blender.RGB48Blender[S, O],
	S color.Space,
	O order.RGBOrder,
](rbuf *buffer.RenderingBufferU16, blender B) *PixFmtAlphaBlendRGB48[B, S, O] {
	return &PixFmtAlphaBlendRGB48[B, S, O]{
		rbuf:    rbuf,
		blender: blender,
	}
}

// Width returns the buffer width
func (pf *PixFmtAlphaBlendRGB48[B, CS, O]) Width() int {
	return pf.rbuf.Width()
}

// Height returns the buffer height
func (pf *PixFmtAlphaBlendRGB48[B, CS, O]) Height() int {
	return pf.rbuf.Height()
}

// PixWidth returns bytes per pixel (6 for RGB48)
func (pf *PixFmtAlphaBlendRGB48[B, CS, O]) PixWidth() int {
	return 6
}

// GetPixel returns the pixel at the given coordinates
func (pf *PixFmtAlphaBlendRGB48[B, CS, O]) GetPixel(x, y int) color.RGB16[CS] {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return color.RGB16[CS]{}
	}

	row := buffer.RowU16(pf.rbuf, y)
	pixelOffset := x * 3 // 3 components per pixel
	if pixelOffset+2 >= len(row) {
		return color.RGB16[CS]{}
	}

	order := blender.GetRGBColorOrder[O]()
	return color.RGB16[CS]{
		R: row[pixelOffset+order.R],
		G: row[pixelOffset+order.G],
		B: row[pixelOffset+order.B],
	}
}

// CopyPixel copies a pixel without blending
func (pf *PixFmtAlphaBlendRGB48[B, CS, O]) CopyPixel(x, y int, c color.RGB16[CS]) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}

	row := buffer.RowU16(pf.rbuf, y)
	pixelOffset := x * 3
	if pixelOffset+2 >= len(row) {
		return
	}

	order := blender.GetRGBColorOrder[O]()
	row[pixelOffset+order.R] = c.R
	row[pixelOffset+order.G] = c.G
	row[pixelOffset+order.B] = c.B
}

// BlendPixel blends a pixel with the given alpha and coverage
func (pf *PixFmtAlphaBlendRGB48[B, CS, O]) BlendPixel(x, y int, c color.RGB16[CS], alpha, cover basics.Int16u) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}

	row := buffer.RowU16(pf.rbuf, y)
	pixelOffset := x * 3
	if pixelOffset+2 >= len(row) {
		return
	}

	// Direct blending call - no type assertion needed with proper constraints
	pf.blender.BlendPix(row[pixelOffset:pixelOffset+3], c.R, c.G, c.B, alpha, cover)
}

// Clear clears the entire buffer with the given color
func (pf *PixFmtAlphaBlendRGB48[B, CS, O]) Clear(c color.RGB16[CS]) {
	for y := 0; y < pf.Height(); y++ {
		row := buffer.RowU16(pf.rbuf, y)
		for x := 0; x < pf.Width(); x++ {
			pixelOffset := x * 3
			if pixelOffset+2 < len(row) {
				order := blender.GetRGBColorOrder[O]()
				row[pixelOffset+order.R] = c.R
				row[pixelOffset+order.G] = c.G
				row[pixelOffset+order.B] = c.B
			}
		}
	}
}

// Concrete RGB48 pixel format types
type (
	PixFmtRGB48Linear = PixFmtAlphaBlendRGB48[blender.BlenderRGB48Linear, color.Linear, color.RGB24Order]
	PixFmtBGR48Linear = PixFmtAlphaBlendRGB48[blender.BlenderBGR48Linear, color.Linear, color.BGR24Order]
	PixFmtRGB48SRGB   = PixFmtAlphaBlendRGB48[blender.BlenderRGB48SRGB, color.SRGB, color.RGB24Order]
	PixFmtBGR48SRGB   = PixFmtAlphaBlendRGB48[blender.BlenderBGR48SRGB, color.SRGB, color.BGR24Order]
)

// Constructor functions for RGB48 pixel formats
func NewPixFmtRGB48Linear(rbuf *buffer.RenderingBufferU16) *PixFmtRGB48Linear {
	return NewPixFmtAlphaBlendRGB48[blender.BlenderRGB48Linear, color.Linear, color.RGB24Order](rbuf, blender.BlenderRGB48Linear{})
}

func NewPixFmtBGR48Linear(rbuf *buffer.RenderingBufferU16) *PixFmtBGR48Linear {
	return NewPixFmtAlphaBlendRGB48[blender.BlenderBGR48Linear, color.Linear, color.BGR24Order](rbuf, blender.BlenderBGR48Linear{})
}

func NewPixFmtRGB48SRGB(rbuf *buffer.RenderingBufferU16) *PixFmtRGB48SRGB {
	return NewPixFmtAlphaBlendRGB48[blender.BlenderRGB48SRGB, color.SRGB, color.RGB24Order](rbuf, blender.BlenderRGB48SRGB{})
}

func NewPixFmtBGR48SRGB(rbuf *buffer.RenderingBufferU16) *PixFmtBGR48SRGB {
	return NewPixFmtAlphaBlendRGB48[blender.BlenderBGR48SRGB, color.SRGB, color.BGR24Order](rbuf, blender.BlenderBGR48SRGB{})
}

// ==============================================================================
// Premultiplied RGB Pixel Format Types
// ==============================================================================

// RGB48 premultiplied variants
type (
	PixFmtRGB48Pre     = PixFmtAlphaBlendRGB48[blender.BlenderRGB48PreLinear, color.Linear, color.RGB24Order]
	PixFmtBGR48Pre     = PixFmtAlphaBlendRGB48[blender.BlenderBGR48PreLinear, color.Linear, color.BGR24Order]
	PixFmtRGB48PreSRGB = PixFmtAlphaBlendRGB48[blender.BlenderRGB48PreSRGB, color.SRGB, color.RGB24Order]
	PixFmtBGR48PreSRGB = PixFmtAlphaBlendRGB48[blender.BlenderBGR48PreSRGB, color.SRGB, color.BGR24Order]
)

// Constructor functions for premultiplied RGB48 formats
func NewPixFmtRGB48Pre(rbuf *buffer.RenderingBufferU16) *PixFmtRGB48Pre {
	return NewPixFmtAlphaBlendRGB48[blender.BlenderRGB48PreLinear, color.Linear, color.RGB24Order](rbuf, blender.BlenderRGB48PreLinear{})
}

func NewPixFmtBGR48Pre(rbuf *buffer.RenderingBufferU16) *PixFmtBGR48Pre {
	return NewPixFmtAlphaBlendRGB48[blender.BlenderBGR48PreLinear, color.Linear, color.BGR24Order](rbuf, blender.BlenderBGR48PreLinear{})
}

func NewPixFmtRGB48PreSRGB(rbuf *buffer.RenderingBufferU16) *PixFmtRGB48PreSRGB {
	return NewPixFmtAlphaBlendRGB48[blender.BlenderRGB48PreSRGB, color.SRGB, color.RGB24Order](rbuf, blender.BlenderRGB48PreSRGB{})
}

func NewPixFmtBGR48PreSRGB(rbuf *buffer.RenderingBufferU16) *PixFmtBGR48PreSRGB {
	return NewPixFmtAlphaBlendRGB48[blender.BlenderBGR48PreSRGB, color.SRGB, color.BGR24Order](rbuf, blender.BlenderBGR48PreSRGB{})
}
