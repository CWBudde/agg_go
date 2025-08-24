package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/pixfmt/blender"
)

// RGBA pixel type for internal operations
type RGBAPixelType struct {
	R, G, B, A basics.Int8u
}

// Set sets all RGBA components
func (p *RGBAPixelType) Set(r, g, b, a basics.Int8u) {
	p.R, p.G, p.B, p.A = r, g, b, a
}

// SetColor sets from a color type
func (p *RGBAPixelType) SetColor(c color.RGBA8[color.Linear]) {
	p.R, p.G, p.B, p.A = c.R, c.G, c.B, c.A
}

// GetColor returns as color type
func (p *RGBAPixelType) GetColor() color.RGBA8[color.Linear] {
	return color.RGBA8[color.Linear]{R: p.R, G: p.G, B: p.B, A: p.A}
}

// PixFmtAlphaBlendRGBA represents the main RGBA pixel format with alpha blending
type PixFmtAlphaBlendRGBA[B blender.RGBABlender, CS any] struct {
	rbuf     *buffer.RenderingBufferU8
	blender  B
	category PixFmtRGBATag
}

// NewPixFmtAlphaBlendRGBA creates a new RGBA pixel format
func NewPixFmtAlphaBlendRGBA[B blender.RGBABlender, CS any](rbuf *buffer.RenderingBufferU8, blender B) *PixFmtAlphaBlendRGBA[B, CS] {
	return &PixFmtAlphaBlendRGBA[B, CS]{
		rbuf:    rbuf,
		blender: blender,
	}
}

// Width returns the buffer width
func (pf *PixFmtAlphaBlendRGBA[B, CS]) Width() int {
	return pf.rbuf.Width()
}

// Height returns the buffer height
func (pf *PixFmtAlphaBlendRGBA[B, CS]) Height() int {
	return pf.rbuf.Height()
}

// PixWidth returns bytes per pixel (4 for RGBA)
func (pf *PixFmtAlphaBlendRGBA[B, CS]) PixWidth() int {
	return 4
}

// GetPixel returns the pixel at the given coordinates
func (pf *PixFmtAlphaBlendRGBA[B, CS]) GetPixel(x, y int) color.RGBA8[CS] {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return color.RGBA8[CS]{}
	}

	row := buffer.RowU8(pf.rbuf, y)
	pixelOffset := x * 4
	if pixelOffset+3 >= len(row) {
		return color.RGBA8[CS]{}
	}

	// Assume RGBA order for now - we'll make this configurable later
	return color.RGBA8[CS]{
		R: row[pixelOffset],
		G: row[pixelOffset+1],
		B: row[pixelOffset+2],
		A: row[pixelOffset+3],
	}
}

// Pixel returns the pixel at the given coordinates (alias for GetPixel)
func (pf *PixFmtAlphaBlendRGBA[B, CS]) Pixel(x, y int) color.RGBA8[CS] {
	return pf.GetPixel(x, y)
}

// CopyPixel copies a pixel without blending
func (pf *PixFmtAlphaBlendRGBA[B, CS]) CopyPixel(x, y int, c color.RGBA8[CS]) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}

	row := buffer.RowU8(pf.rbuf, y)
	pixelOffset := x * 4
	if pixelOffset+3 >= len(row) {
		return
	}

	// Assume RGBA order for now - we'll make this configurable later
	row[pixelOffset] = c.R
	row[pixelOffset+1] = c.G
	row[pixelOffset+2] = c.B
	row[pixelOffset+3] = c.A
}

// BlendPixel blends a pixel with the given coverage
func (pf *PixFmtAlphaBlendRGBA[B, CS]) BlendPixel(x, y int, c color.RGBA8[CS], cover basics.Int8u) {
	if !InBounds(x, y, pf.Width(), pf.Height()) || c.IsTransparent() {
		return
	}

	row := buffer.RowU8(pf.rbuf, y)
	pixelOffset := x * 4
	if pixelOffset+3 >= len(row) {
		return
	}

	// Direct blending call - no type assertion needed with proper constraints
	pf.blender.BlendPix(row[pixelOffset:pixelOffset+4], c.R, c.G, c.B, c.A, cover)
}

// CopyHline copies a horizontal line
func (pf *PixFmtAlphaBlendRGBA[B, CS]) CopyHline(x, y, length int, c color.RGBA8[CS]) {
	if y < 0 || y >= pf.Height() || length <= 0 {
		return
	}

	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}

	row := buffer.RowU8(pf.rbuf, y)
	for i := 0; i < length; i++ {
		pixelOffset := (x + i) * 4
		if pixelOffset+3 < len(row) {
			row[pixelOffset] = c.R
			row[pixelOffset+1] = c.G
			row[pixelOffset+2] = c.B
			row[pixelOffset+3] = c.A
		}
	}
}

// BlendHline blends a horizontal line
func (pf *PixFmtAlphaBlendRGBA[B, CS]) BlendHline(x, y, length int, c color.RGBA8[CS], cover basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || c.IsTransparent() {
		return
	}

	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}

	row := buffer.RowU8(pf.rbuf, y)
	for i := 0; i < length; i++ {
		pixelOffset := (x + i) * 4
		if pixelOffset+3 < len(row) {
			pf.blender.BlendPix(row[pixelOffset:pixelOffset+4], c.R, c.G, c.B, c.A, cover)
		}
	}
}

// CopyVline copies a vertical line
func (pf *PixFmtAlphaBlendRGBA[B, CS]) CopyVline(x, y, length int, c color.RGBA8[CS]) {
	if x < 0 || x >= pf.Width() || length <= 0 {
		return
	}

	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}

	for i := 0; i < length; i++ {
		pf.CopyPixel(x, y+i, c)
	}
}

// BlendVline blends a vertical line
func (pf *PixFmtAlphaBlendRGBA[B, CS]) BlendVline(x, y, length int, c color.RGBA8[CS], cover basics.Int8u) {
	if x < 0 || x >= pf.Width() || length <= 0 || c.IsTransparent() {
		return
	}

	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}

	for i := 0; i < length; i++ {
		pf.BlendPixel(x, y+i, c, cover)
	}
}

// CopyBar copies a filled rectangle
func (pf *PixFmtAlphaBlendRGBA[B, CS]) CopyBar(x1, y1, x2, y2 int, c color.RGBA8[CS]) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if x1 > x2 {
		x1, x2 = x2, x1
	}

	for y := y1; y <= y2; y++ {
		pf.CopyHline(x1, y, x2, c)
	}
}

// BlendBar blends a filled rectangle
func (pf *PixFmtAlphaBlendRGBA[B, CS]) BlendBar(x1, y1, x2, y2 int, c color.RGBA8[CS], cover basics.Int8u) {
	if c.IsTransparent() {
		return
	}

	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if x1 > x2 {
		x1, x2 = x2, x1
	}

	for y := y1; y <= y2; y++ {
		pf.BlendHline(x1, y, x2, c, cover)
	}
}

// BlendSolidHspan blends a horizontal span with varying coverage
func (pf *PixFmtAlphaBlendRGBA[B, CS]) BlendSolidHspan(x, y, length int, c color.RGBA8[CS], covers []basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || c.IsTransparent() {
		return
	}

	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}

	row := buffer.RowU8(pf.rbuf, y)
	if covers == nil {
		// Uniform coverage
		for i := 0; i < length; i++ {
			pixelOffset := (x + i) * 4
			if pixelOffset+3 < len(row) {
				pf.blender.BlendPix(row[pixelOffset:pixelOffset+4], c.R, c.G, c.B, c.A, 255)
			}
		}
	} else {
		// Varying coverage
		for i := 0; i < length && i < len(covers); i++ {
			if covers[i] > 0 {
				pixelOffset := (x + i) * 4
				if pixelOffset+3 < len(row) {
					pf.blender.BlendPix(row[pixelOffset:pixelOffset+4], c.R, c.G, c.B, c.A, covers[i])
				}
			}
		}
	}
}

// BlendSolidVspan blends a vertical span with varying coverage
func (pf *PixFmtAlphaBlendRGBA[B, CS]) BlendSolidVspan(x, y, length int, c color.RGBA8[CS], covers []basics.Int8u) {
	if x < 0 || x >= pf.Width() || length <= 0 || c.IsTransparent() {
		return
	}

	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}

	if covers == nil {
		// Uniform coverage
		for i := 0; i < length; i++ {
			pf.BlendPixel(x, y+i, c, 255)
		}
	} else {
		// Varying coverage
		for i := 0; i < length && i < len(covers); i++ {
			if covers[i] > 0 {
				pf.BlendPixel(x, y+i, c, covers[i])
			}
		}
	}
}

// CopyColorHspan copies a horizontal span of colors
func (pf *PixFmtAlphaBlendRGBA[B, CS]) CopyColorHspan(x, y, length int, colors []color.RGBA8[CS]) {
	if y < 0 || y >= pf.Height() || length <= 0 || len(colors) == 0 {
		return
	}

	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}

	for i := 0; i < length; i++ {
		colorIdx := i % len(colors)
		pf.CopyPixel(x+i, y, colors[colorIdx])
	}
}

// BlendColorHspan blends a horizontal span of colors
func (pf *PixFmtAlphaBlendRGBA[B, CS]) BlendColorHspan(x, y, length int, colors []color.RGBA8[CS], covers []basics.Int8u, cover basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || len(colors) == 0 {
		return
	}

	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}

	for i := 0; i < length; i++ {
		colorIdx := i % len(colors)
		c := colors[colorIdx]
		if c.IsTransparent() {
			continue
		}

		cvr := cover
		if covers != nil && i < len(covers) {
			cvr = covers[i]
		}
		pf.BlendPixel(x+i, y, c, cvr)
	}
}

// CopyColorVspan copies a vertical span of colors
func (pf *PixFmtAlphaBlendRGBA[B, CS]) CopyColorVspan(x, y, length int, colors []color.RGBA8[CS]) {
	if x < 0 || x >= pf.Width() || length <= 0 || len(colors) == 0 {
		return
	}

	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}

	for i := 0; i < length; i++ {
		colorIdx := i % len(colors)
		pf.CopyPixel(x, y+i, colors[colorIdx])
	}
}

// BlendColorVspan blends a vertical span of colors
func (pf *PixFmtAlphaBlendRGBA[B, CS]) BlendColorVspan(x, y, length int, colors []color.RGBA8[CS], covers []basics.Int8u, cover basics.Int8u) {
	if x < 0 || x >= pf.Width() || length <= 0 || len(colors) == 0 {
		return
	}

	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}

	for i := 0; i < length; i++ {
		colorIdx := i % len(colors)
		c := colors[colorIdx]
		if c.IsTransparent() {
			continue
		}

		cvr := cover
		if covers != nil && i < len(covers) {
			cvr = covers[i]
		}
		pf.BlendPixel(x, y+i, c, cvr)
	}
}

// Clear clears the entire buffer with the given color
func (pf *PixFmtAlphaBlendRGBA[B, CS]) Clear(c color.RGBA8[CS]) {
	for y := 0; y < pf.Height(); y++ {
		row := buffer.RowU8(pf.rbuf, y)
		for x := 0; x < pf.Width(); x++ {
			pixelOffset := x * 4
			if pixelOffset+3 < len(row) {
				row[pixelOffset] = c.R
				row[pixelOffset+1] = c.G
				row[pixelOffset+2] = c.B
				row[pixelOffset+3] = c.A
			}
		}
	}
}

// Fill is an alias for Clear (fills entire buffer)
func (pf *PixFmtAlphaBlendRGBA[B, CS]) Fill(c color.RGBA8[CS]) {
	pf.Clear(c)
}

// Concrete RGBA pixel format types for different color orders
type (
	PixFmtRGBA32 = PixFmtAlphaBlendRGBA[blender.BlenderRGBA8, color.Linear]
	PixFmtARGB32 = PixFmtAlphaBlendRGBA[blender.BlenderARGB8, color.Linear]
	PixFmtBGRA32 = PixFmtAlphaBlendRGBA[blender.BlenderBGRA8, color.Linear]
	PixFmtABGR32 = PixFmtAlphaBlendRGBA[blender.BlenderABGR8, color.Linear]
)

// Constructor functions
func NewPixFmtRGBA32(rbuf *buffer.RenderingBufferU8) *PixFmtRGBA32 {
	return NewPixFmtAlphaBlendRGBA[blender.BlenderRGBA8, color.Linear](rbuf, blender.BlenderRGBA8{})
}

func NewPixFmtARGB32(rbuf *buffer.RenderingBufferU8) *PixFmtARGB32 {
	return NewPixFmtAlphaBlendRGBA[blender.BlenderARGB8, color.Linear](rbuf, blender.BlenderARGB8{})
}

func NewPixFmtBGRA32(rbuf *buffer.RenderingBufferU8) *PixFmtBGRA32 {
	return NewPixFmtAlphaBlendRGBA[blender.BlenderBGRA8, color.Linear](rbuf, blender.BlenderBGRA8{})
}

func NewPixFmtABGR32(rbuf *buffer.RenderingBufferU8) *PixFmtABGR32 {
	return NewPixFmtAlphaBlendRGBA[blender.BlenderABGR8, color.Linear](rbuf, blender.BlenderABGR8{})
}
