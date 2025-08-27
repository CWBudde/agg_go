package pixfmt

import (
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/order"
	"agg_go/internal/pixfmt/blender"
)

// ==============================================================================
// RGB96 (32-bit float per channel) Pixel Formats
// ==============================================================================

// PixFmtAlphaBlendRGB96 represents RGB pixel format with 32-bit float components (12 bytes per pixel)
type PixFmtAlphaBlendRGB96[
	B blender.RGB96Blender[S, O],
	S color.Space,
	O order.RGBOrder,
] struct {
	rbuf     *buffer.RenderingBufferF32
	blender  B
	category PixFmtRGBTag
}

// NewPixFmtAlphaBlendRGB96 creates a new RGB96 pixel format
func NewPixFmtAlphaBlendRGB96[B blender.RGB96Blender[S, O], S color.Space, O order.RGBOrder](rbuf *buffer.RenderingBufferF32, blender B) *PixFmtAlphaBlendRGB96[B, S, O] {
	return &PixFmtAlphaBlendRGB96[B, S, O]{
		rbuf:    rbuf,
		blender: blender,
	}
}

// Width returns the buffer width
func (pf *PixFmtAlphaBlendRGB96[B, CS, O]) Width() int {
	return pf.rbuf.Width()
}

// Height returns the buffer height
func (pf *PixFmtAlphaBlendRGB96[B, CS, O]) Height() int {
	return pf.rbuf.Height()
}

// PixWidth returns bytes per pixel (12 for RGB96)
func (pf *PixFmtAlphaBlendRGB96[B, CS, O]) PixWidth() int {
	return 12
}

// GetPixel returns the pixel at the given coordinates
func (pf *PixFmtAlphaBlendRGB96[B, CS, O]) GetPixel(x, y int) color.RGB32[CS] {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return color.RGB32[CS]{}
	}

	row := buffer.RowF32(pf.rbuf, y)
	pixelOffset := x * 3 // 3 components per pixel
	if pixelOffset+2 >= len(row) {
		return color.RGB32[CS]{}
	}

	var order O
	return color.RGB32[CS]{
		R: row[pixelOffset+order.IdxR()],
		G: row[pixelOffset+order.IdxG()],
		B: row[pixelOffset+order.IdxB()],
	}
}

// CopyPixel copies a pixel without blending
func (pf *PixFmtAlphaBlendRGB96[B, CS, O]) CopyPixel(x, y int, c color.RGB32[CS]) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}

	row := buffer.RowF32(pf.rbuf, y)
	pixelOffset := x * 3
	if pixelOffset+2 >= len(row) {
		return
	}

	var order O
	row[pixelOffset+order.IdxR()] = c.R
	row[pixelOffset+order.IdxG()] = c.G
	row[pixelOffset+order.IdxB()] = c.B
}

// BlendPixel blends a pixel with the given alpha and coverage
func (pf *PixFmtAlphaBlendRGB96[B, CS, O]) BlendPixel(x, y int, c color.RGB32[CS], alpha, cover float32) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}

	row := buffer.RowF32(pf.rbuf, y)
	pixelOffset := x * 3
	if pixelOffset+2 >= len(row) {
		return
	}

	// Direct blending call - no type assertion needed with proper constraints
	pf.blender.BlendPix(row[pixelOffset:pixelOffset+3], c.R, c.G, c.B, alpha, cover)
}

// Clear clears the entire buffer with the given color
func (pf *PixFmtAlphaBlendRGB96[B, CS, O]) Clear(c color.RGB32[CS]) {
	for y := 0; y < pf.Height(); y++ {
		row := buffer.RowF32(pf.rbuf, y)
		for x := 0; x < pf.Width(); x++ {
			pixelOffset := x * 3
			if pixelOffset+2 < len(row) {
				var order O
				row[pixelOffset+order.IdxR()] = c.R
				row[pixelOffset+order.IdxG()] = c.G
				row[pixelOffset+order.IdxB()] = c.B
			}
		}
	}
}

// Concrete RGB96 pixel format types
type (
	PixFmtRGB96Linear = PixFmtAlphaBlendRGB96[blender.BlenderRGB96Linear, color.Linear, order.RGB]
	PixFmtBGR96Linear = PixFmtAlphaBlendRGB96[blender.BlenderBGR96Linear, color.Linear, order.BGR]
	PixFmtRGB96SRGB   = PixFmtAlphaBlendRGB96[blender.BlenderRGB96SRGB, color.SRGB, order.RGB]
	PixFmtBGR96SRGB   = PixFmtAlphaBlendRGB96[blender.BlenderBGR96SRGB, color.SRGB, order.BGR]
)

// Constructor functions for RGB96 pixel formats
func NewPixFmtRGB96Linear(rbuf *buffer.RenderingBufferF32) *PixFmtRGB96Linear {
	return NewPixFmtAlphaBlendRGB96[blender.BlenderRGB96Linear, color.Linear, order.RGB](rbuf, blender.BlenderRGB96Linear{})
}

func NewPixFmtBGR96Linear(rbuf *buffer.RenderingBufferF32) *PixFmtBGR96Linear {
	return NewPixFmtAlphaBlendRGB96[blender.BlenderBGR96Linear, color.Linear, order.BGR](rbuf, blender.BlenderBGR96Linear{})
}

func NewPixFmtRGB96SRGB(rbuf *buffer.RenderingBufferF32) *PixFmtRGB96SRGB {
	return NewPixFmtAlphaBlendRGB96[blender.BlenderRGB96SRGB, color.SRGB, order.RGB](rbuf, blender.BlenderRGB96SRGB{})
}

func NewPixFmtBGR96SRGB(rbuf *buffer.RenderingBufferF32) *PixFmtBGR96SRGB {
	return NewPixFmtAlphaBlendRGB96[blender.BlenderBGR96SRGB, color.SRGB, order.BGR](rbuf, blender.BlenderBGR96SRGB{})
}

// ==============================================================================
// Premultiplied RGB Pixel Format Types
// ==============================================================================

// RGB96 premultiplied variants
type (
	PixFmtRGB96Pre     = PixFmtAlphaBlendRGB96[blender.BlenderRGB96PreLinear, color.Linear, order.RGB]
	PixFmtBGR96Pre     = PixFmtAlphaBlendRGB96[blender.BlenderBGR96PreLinear, color.Linear, order.BGR]
	PixFmtRGB96PreSRGB = PixFmtAlphaBlendRGB96[blender.BlenderRGB96PreSRGB, color.SRGB, order.RGB]
	PixFmtBGR96PreSRGB = PixFmtAlphaBlendRGB96[blender.BlenderBGR96PreSRGB, color.SRGB, order.BGR]
)

// Constructor functions for premultiplied RGB96 formats
func NewPixFmtRGB96Pre(rbuf *buffer.RenderingBufferF32) *PixFmtRGB96Pre {
	return NewPixFmtAlphaBlendRGB96[blender.BlenderRGB96PreLinear, color.Linear, order.RGB](rbuf, blender.BlenderRGB96PreLinear{})
}

func NewPixFmtBGR96Pre(rbuf *buffer.RenderingBufferF32) *PixFmtBGR96Pre {
	return NewPixFmtAlphaBlendRGB96[blender.BlenderBGR96PreLinear, color.Linear, order.BGR](rbuf, blender.BlenderBGR96PreLinear{})
}

func NewPixFmtRGB96PreSRGB(rbuf *buffer.RenderingBufferF32) *PixFmtRGB96PreSRGB {
	return NewPixFmtAlphaBlendRGB96[blender.BlenderRGB96PreSRGB, color.SRGB, order.RGB](rbuf, blender.BlenderRGB96PreSRGB{})
}

func NewPixFmtBGR96PreSRGB(rbuf *buffer.RenderingBufferF32) *PixFmtBGR96PreSRGB {
	return NewPixFmtAlphaBlendRGB96[blender.BlenderBGR96PreSRGB, color.SRGB, order.BGR](rbuf, blender.BlenderBGR96PreSRGB{})
}

// Constructor functions for premultiplied RGBX32 formats
func NewPixFmtRGBX32Pre(rbuf *buffer.RenderingBufferU8) *PixFmtRGBX32Pre {
	return NewPixFmtAlphaBlendRGBX32[blender.BlenderRGB24Pre, color.Linear, order.RGBX32](rbuf, blender.BlenderRGB24Pre{})
}

func NewPixFmtXRGB32Pre(rbuf *buffer.RenderingBufferU8) *PixFmtXRGB32Pre {
	return NewPixFmtAlphaBlendRGBX32[blender.BlenderRGB24Pre, color.Linear, order.XRGB32](rbuf, blender.BlenderRGB24Pre{})
}

func NewPixFmtBGRX32Pre(rbuf *buffer.RenderingBufferU8) *PixFmtBGRX32Pre {
	return NewPixFmtAlphaBlendRGBX32[blender.BlenderBGR24Pre, color.Linear, order.BGRX32](rbuf, blender.BlenderBGR24Pre{})
}

func NewPixFmtXBGR32Pre(rbuf *buffer.RenderingBufferU8) *PixFmtXBGR32Pre {
	return NewPixFmtAlphaBlendRGBX32[blender.BlenderBGR24Pre, color.Linear, order.XBGR32](rbuf, blender.BlenderBGR24Pre{})
}
