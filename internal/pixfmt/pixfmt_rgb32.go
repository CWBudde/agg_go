package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/pixfmt/blender"
)

// ==============================================================================
// RGB96 (32-bit float per channel) Pixel Formats
// ==============================================================================

// PixFmtAlphaBlendRGB96 represents RGB pixel format with 32-bit float components (12 bytes per pixel)
type PixFmtAlphaBlendRGB96[B blender.RGB96Blender, CS any, O any] struct {
	rbuf     *buffer.RenderingBufferF32
	blender  B
	category PixFmtRGBTag
}

// NewPixFmtAlphaBlendRGB96 creates a new RGB96 pixel format
func NewPixFmtAlphaBlendRGB96[B blender.RGB96Blender, CS any, O any](rbuf *buffer.RenderingBufferF32, blender B) *PixFmtAlphaBlendRGB96[B, CS, O] {
	return &PixFmtAlphaBlendRGB96[B, CS, O]{
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

	order := blender.GetRGBColorOrder[O]()
	return color.RGB32[CS]{
		R: row[pixelOffset+order.R],
		G: row[pixelOffset+order.G],
		B: row[pixelOffset+order.B],
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

	order := blender.GetRGBColorOrder[O]()
	row[pixelOffset+order.R] = c.R
	row[pixelOffset+order.G] = c.G
	row[pixelOffset+order.B] = c.B
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
				order := blender.GetRGBColorOrder[O]()
				row[pixelOffset+order.R] = c.R
				row[pixelOffset+order.G] = c.G
				row[pixelOffset+order.B] = c.B
			}
		}
	}
}

// Concrete RGB96 pixel format types
type (
	PixFmtRGB96Linear = PixFmtAlphaBlendRGB96[blender.BlenderRGB96Linear, color.Linear, color.RGB24Order]
	PixFmtBGR96Linear = PixFmtAlphaBlendRGB96[blender.BlenderBGR96Linear, color.Linear, color.BGR24Order]
	PixFmtRGB96SRGB   = PixFmtAlphaBlendRGB96[blender.BlenderRGB96SRGB, color.SRGB, color.RGB24Order]
	PixFmtBGR96SRGB   = PixFmtAlphaBlendRGB96[blender.BlenderBGR96SRGB, color.SRGB, color.BGR24Order]
)

// Constructor functions for RGB96 pixel formats
func NewPixFmtRGB96Linear(rbuf *buffer.RenderingBufferF32) *PixFmtRGB96Linear {
	return NewPixFmtAlphaBlendRGB96[blender.BlenderRGB96Linear, color.Linear, color.RGB24Order](rbuf, blender.BlenderRGB96Linear{})
}

func NewPixFmtBGR96Linear(rbuf *buffer.RenderingBufferF32) *PixFmtBGR96Linear {
	return NewPixFmtAlphaBlendRGB96[blender.BlenderBGR96Linear, color.Linear, color.BGR24Order](rbuf, blender.BlenderBGR96Linear{})
}

func NewPixFmtRGB96SRGB(rbuf *buffer.RenderingBufferF32) *PixFmtRGB96SRGB {
	return NewPixFmtAlphaBlendRGB96[blender.BlenderRGB96SRGB, color.SRGB, color.RGB24Order](rbuf, blender.BlenderRGB96SRGB{})
}

func NewPixFmtBGR96SRGB(rbuf *buffer.RenderingBufferF32) *PixFmtBGR96SRGB {
	return NewPixFmtAlphaBlendRGB96[blender.BlenderBGR96SRGB, color.SRGB, color.BGR24Order](rbuf, blender.BlenderBGR96SRGB{})
}

// ==============================================================================
// RGBX32/XRGB32 (RGB with padding byte) Pixel Formats
// ==============================================================================

// These formats store RGB in 4 bytes with one padding byte
// RGBX32: RGB + padding byte (RGBX ordering)
// XRGB32: padding byte + RGB (XRGB ordering)
// They reuse the RGB24 blenders but with step=4 instead of step=3

// Color order for RGBX32/XRGB32 formats
type (
	RGBX32Order struct{} // RGB components at offsets 0,1,2, padding at 3
	XRGB32Order struct{} // Padding at 0, RGB components at offsets 1,2,3
	BGRX32Order struct{} // BGR components at offsets 0,1,2, padding at 3
	XBGR32Order struct{} // Padding at 0, BGR components at offsets 1,2,3
)

// Color ordering for 32-bit RGB formats with padding
var (
	OrderRGBX32 = color.ColorOrder{R: 0, G: 1, B: 2, A: -1} // Padding at byte 3
	OrderXRGB32 = color.ColorOrder{R: 1, G: 2, B: 3, A: -1} // Padding at byte 0
	OrderBGRX32 = color.ColorOrder{R: 2, G: 1, B: 0, A: -1} // Padding at byte 3
	OrderXBGR32 = color.ColorOrder{R: 3, G: 2, B: 1, A: -1} // Padding at byte 0
)

// Helper function to get RGB color order for 32-bit padded formats
func getRGB32ColorOrder[O any]() color.ColorOrder {
	var order color.ColorOrder
	switch any(*new(O)).(type) {
	case RGBX32Order:
		order = OrderRGBX32
	case XRGB32Order:
		order = OrderXRGB32
	case BGRX32Order:
		order = OrderBGRX32
	case XBGR32Order:
		order = OrderXBGR32
	default:
		// Default to RGBX order
		order = OrderRGBX32
	}
	return order
}

// PixFmtAlphaBlendRGBX32 represents 32-bit RGB with padding (4 bytes per pixel)
type PixFmtAlphaBlendRGBX32[B blender.RGBBlender, CS any, O any] struct {
	rbuf     *buffer.RenderingBufferU8
	blender  B
	category PixFmtRGBTag
}

// NewPixFmtAlphaBlendRGBX32 creates a new RGBX32 pixel format
func NewPixFmtAlphaBlendRGBX32[B blender.RGBBlender, CS any, O any](rbuf *buffer.RenderingBufferU8, blender B) *PixFmtAlphaBlendRGBX32[B, CS, O] {
	return &PixFmtAlphaBlendRGBX32[B, CS, O]{
		rbuf:    rbuf,
		blender: blender,
	}
}

// Width returns the buffer width
func (pf *PixFmtAlphaBlendRGBX32[B, CS, O]) Width() int {
	return pf.rbuf.Width()
}

// Height returns the buffer height
func (pf *PixFmtAlphaBlendRGBX32[B, CS, O]) Height() int {
	return pf.rbuf.Height()
}

// PixWidth returns bytes per pixel (4 for RGBX32)
func (pf *PixFmtAlphaBlendRGBX32[B, CS, O]) PixWidth() int {
	return 4
}

// GetPixel returns the pixel at the given coordinates
func (pf *PixFmtAlphaBlendRGBX32[B, CS, O]) GetPixel(x, y int) color.RGB8[CS] {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return color.RGB8[CS]{}
	}

	row := buffer.RowU8(pf.rbuf, y)
	pixelOffset := x * 4 // 4 bytes per pixel
	if pixelOffset+3 >= len(row) {
		return color.RGB8[CS]{}
	}

	order := getRGB32ColorOrder[O]()
	return color.RGB8[CS]{
		R: row[pixelOffset+order.R],
		G: row[pixelOffset+order.G],
		B: row[pixelOffset+order.B],
	}
}

// CopyPixel copies a pixel without blending
func (pf *PixFmtAlphaBlendRGBX32[B, CS, O]) CopyPixel(x, y int, c color.RGB8[CS]) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}

	row := buffer.RowU8(pf.rbuf, y)
	pixelOffset := x * 4
	if pixelOffset+3 >= len(row) {
		return
	}

	order := getRGB32ColorOrder[O]()
	row[pixelOffset+order.R] = c.R
	row[pixelOffset+order.G] = c.G
	row[pixelOffset+order.B] = c.B
	// Leave padding byte untouched
}

// BlendPixel blends a pixel with the given alpha and coverage
func (pf *PixFmtAlphaBlendRGBX32[B, CS, O]) BlendPixel(x, y int, c color.RGB8[CS], alpha, cover basics.Int8u) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}

	row := buffer.RowU8(pf.rbuf, y)
	pixelOffset := x * 4
	if pixelOffset+3 >= len(row) {
		return
	}

	// Create a temporary 3-byte array for the RGB components
	order := getRGB32ColorOrder[O]()
	rgb := []basics.Int8u{
		row[pixelOffset+order.R],
		row[pixelOffset+order.G],
		row[pixelOffset+order.B],
	}

	// Use interface assertion for blending on the RGB components
	if blender, ok := any(pf.blender).(blender.RGBBlender); ok {
		blender.BlendPix(rgb, c.R, c.G, c.B, alpha, cover)

		// Copy back the blended RGB values
		row[pixelOffset+order.R] = rgb[0]
		row[pixelOffset+order.G] = rgb[1]
		row[pixelOffset+order.B] = rgb[2]
		// Leave padding byte untouched
	}
}

// Clear clears the entire buffer with the given color
func (pf *PixFmtAlphaBlendRGBX32[B, CS, O]) Clear(c color.RGB8[CS]) {
	for y := 0; y < pf.Height(); y++ {
		row := buffer.RowU8(pf.rbuf, y)
		for x := 0; x < pf.Width(); x++ {
			pixelOffset := x * 4
			if pixelOffset+3 < len(row) {
				order := getRGB32ColorOrder[O]()
				row[pixelOffset+order.R] = c.R
				row[pixelOffset+order.G] = c.G
				row[pixelOffset+order.B] = c.B
				// Leave padding byte untouched
			}
		}
	}
}

// Concrete RGBX32 pixel format types
type (
	PixFmtRGBX32 = PixFmtAlphaBlendRGBX32[blender.BlenderRGB24, color.Linear, RGBX32Order]
	PixFmtXRGB32 = PixFmtAlphaBlendRGBX32[blender.BlenderRGB24, color.Linear, XRGB32Order]
	PixFmtBGRX32 = PixFmtAlphaBlendRGBX32[blender.BlenderBGR24, color.Linear, BGRX32Order]
	PixFmtXBGR32 = PixFmtAlphaBlendRGBX32[blender.BlenderBGR24, color.Linear, XBGR32Order]

	PixFmtSRGBX32 = PixFmtAlphaBlendRGBX32[blender.BlenderRGB24SRGB, color.SRGB, RGBX32Order]
	PixFmtSXRGB32 = PixFmtAlphaBlendRGBX32[blender.BlenderRGB24SRGB, color.SRGB, XRGB32Order]
	PixFmtSBGRX32 = PixFmtAlphaBlendRGBX32[blender.BlenderBGR24SRGB, color.SRGB, BGRX32Order]
	PixFmtSXBGR32 = PixFmtAlphaBlendRGBX32[blender.BlenderBGR24SRGB, color.SRGB, XBGR32Order]
)

// Constructor functions for RGBX32 pixel formats
func NewPixFmtRGBX32(rbuf *buffer.RenderingBufferU8) *PixFmtRGBX32 {
	return NewPixFmtAlphaBlendRGBX32[blender.BlenderRGB24, color.Linear, RGBX32Order](rbuf, blender.BlenderRGB24{})
}

func NewPixFmtXRGB32(rbuf *buffer.RenderingBufferU8) *PixFmtXRGB32 {
	return NewPixFmtAlphaBlendRGBX32[blender.BlenderRGB24, color.Linear, XRGB32Order](rbuf, blender.BlenderRGB24{})
}

func NewPixFmtBGRX32(rbuf *buffer.RenderingBufferU8) *PixFmtBGRX32 {
	return NewPixFmtAlphaBlendRGBX32[blender.BlenderBGR24, color.Linear, BGRX32Order](rbuf, blender.BlenderBGR24{})
}

func NewPixFmtXBGR32(rbuf *buffer.RenderingBufferU8) *PixFmtXBGR32 {
	return NewPixFmtAlphaBlendRGBX32[blender.BlenderBGR24, color.Linear, XBGR32Order](rbuf, blender.BlenderBGR24{})
}

func NewPixFmtSRGBX32(rbuf *buffer.RenderingBufferU8) *PixFmtSRGBX32 {
	return NewPixFmtAlphaBlendRGBX32[blender.BlenderRGB24SRGB, color.SRGB, RGBX32Order](rbuf, blender.BlenderRGB24SRGB{})
}

func NewPixFmtSXRGB32(rbuf *buffer.RenderingBufferU8) *PixFmtSXRGB32 {
	return NewPixFmtAlphaBlendRGBX32[blender.BlenderRGB24SRGB, color.SRGB, XRGB32Order](rbuf, blender.BlenderRGB24SRGB{})
}

func NewPixFmtSBGRX32(rbuf *buffer.RenderingBufferU8) *PixFmtSBGRX32 {
	return NewPixFmtAlphaBlendRGBX32[blender.BlenderBGR24SRGB, color.SRGB, BGRX32Order](rbuf, blender.BlenderBGR24SRGB{})
}

func NewPixFmtSXBGR32(rbuf *buffer.RenderingBufferU8) *PixFmtSXBGR32 {
	return NewPixFmtAlphaBlendRGBX32[blender.BlenderBGR24SRGB, color.SRGB, XBGR32Order](rbuf, blender.BlenderBGR24SRGB{})
}

// ==============================================================================
// Premultiplied RGB Pixel Format Types
// ==============================================================================

// RGB96 premultiplied variants
type (
	PixFmtRGB96Pre     = PixFmtAlphaBlendRGB96[blender.BlenderRGB96PreLinear, color.Linear, color.RGB24Order]
	PixFmtBGR96Pre     = PixFmtAlphaBlendRGB96[blender.BlenderBGR96PreLinear, color.Linear, color.BGR24Order]
	PixFmtRGB96PreSRGB = PixFmtAlphaBlendRGB96[blender.BlenderRGB96PreSRGB, color.SRGB, color.RGB24Order]
	PixFmtBGR96PreSRGB = PixFmtAlphaBlendRGB96[blender.BlenderBGR96PreSRGB, color.SRGB, color.BGR24Order]
)

// Constructor functions for premultiplied RGB96 formats
func NewPixFmtRGB96Pre(rbuf *buffer.RenderingBufferF32) *PixFmtRGB96Pre {
	return NewPixFmtAlphaBlendRGB96[blender.BlenderRGB96PreLinear, color.Linear, color.RGB24Order](rbuf, blender.BlenderRGB96PreLinear{})
}

func NewPixFmtBGR96Pre(rbuf *buffer.RenderingBufferF32) *PixFmtBGR96Pre {
	return NewPixFmtAlphaBlendRGB96[blender.BlenderBGR96PreLinear, color.Linear, color.BGR24Order](rbuf, blender.BlenderBGR96PreLinear{})
}

func NewPixFmtRGB96PreSRGB(rbuf *buffer.RenderingBufferF32) *PixFmtRGB96PreSRGB {
	return NewPixFmtAlphaBlendRGB96[blender.BlenderRGB96PreSRGB, color.SRGB, color.RGB24Order](rbuf, blender.BlenderRGB96PreSRGB{})
}

func NewPixFmtBGR96PreSRGB(rbuf *buffer.RenderingBufferF32) *PixFmtBGR96PreSRGB {
	return NewPixFmtAlphaBlendRGB96[blender.BlenderBGR96PreSRGB, color.SRGB, color.BGR24Order](rbuf, blender.BlenderBGR96PreSRGB{})
}

// Constructor functions for premultiplied RGBX32 formats
func NewPixFmtRGBX32Pre(rbuf *buffer.RenderingBufferU8) *PixFmtRGBX32Pre {
	return NewPixFmtAlphaBlendRGBX32[blender.BlenderRGB24Pre, color.Linear, RGBX32Order](rbuf, blender.BlenderRGB24Pre{})
}

func NewPixFmtXRGB32Pre(rbuf *buffer.RenderingBufferU8) *PixFmtXRGB32Pre {
	return NewPixFmtAlphaBlendRGBX32[blender.BlenderRGB24Pre, color.Linear, XRGB32Order](rbuf, blender.BlenderRGB24Pre{})
}

func NewPixFmtBGRX32Pre(rbuf *buffer.RenderingBufferU8) *PixFmtBGRX32Pre {
	return NewPixFmtAlphaBlendRGBX32[blender.BlenderBGR24Pre, color.Linear, BGRX32Order](rbuf, blender.BlenderBGR24Pre{})
}

func NewPixFmtXBGR32Pre(rbuf *buffer.RenderingBufferU8) *PixFmtXBGR32Pre {
	return NewPixFmtAlphaBlendRGBX32[blender.BlenderBGR24Pre, color.Linear, XBGR32Order](rbuf, blender.BlenderBGR24Pre{})
}
