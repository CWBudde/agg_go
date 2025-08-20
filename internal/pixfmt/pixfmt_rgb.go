package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
)

// RGB pixel type for internal operations (24-bit, no alpha)
type RGBPixelType struct {
	R, G, B basics.Int8u
}

// Set sets all RGB components
func (p *RGBPixelType) Set(r, g, b basics.Int8u) {
	p.R, p.G, p.B = r, g, b
}

// SetColor sets from a color type
func (p *RGBPixelType) SetColor(c color.RGB8[color.Linear]) {
	p.R, p.G, p.B = c.R, c.G, c.B
}

// GetColor returns as color type
func (p *RGBPixelType) GetColor() color.RGB8[color.Linear] {
	return color.RGB8[color.Linear]{R: p.R, G: p.G, B: p.B}
}

// PixFmtAlphaBlendRGB represents the main RGB pixel format with alpha blending
// This is a 24-bit format (3 bytes per pixel) without alpha channel storage
type PixFmtAlphaBlendRGB[B any, CS any, O any] struct {
	rbuf     *buffer.RenderingBufferU8
	blender  B
	category PixFmtRGBTag
}

// NewPixFmtAlphaBlendRGB creates a new RGB pixel format
func NewPixFmtAlphaBlendRGB[B any, CS any, O any](rbuf *buffer.RenderingBufferU8, blender B) *PixFmtAlphaBlendRGB[B, CS, O] {
	return &PixFmtAlphaBlendRGB[B, CS, O]{
		rbuf:    rbuf,
		blender: blender,
	}
}

// Width returns the buffer width
func (pf *PixFmtAlphaBlendRGB[B, CS, O]) Width() int {
	return pf.rbuf.Width()
}

// Height returns the buffer height
func (pf *PixFmtAlphaBlendRGB[B, CS, O]) Height() int {
	return pf.rbuf.Height()
}

// PixWidth returns bytes per pixel (3 for RGB)
func (pf *PixFmtAlphaBlendRGB[B, CS, O]) PixWidth() int {
	return 3
}

// GetPixel returns the pixel at the given coordinates
func (pf *PixFmtAlphaBlendRGB[B, CS, O]) GetPixel(x, y int) color.RGB8[CS] {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return color.RGB8[CS]{}
	}

	row := buffer.RowU8(pf.rbuf, y)
	pixelOffset := x * 3
	if pixelOffset+2 >= len(row) {
		return color.RGB8[CS]{}
	}

	// Use the order type parameter to get correct color order
	order := getRGBColorOrder[O]()
	return color.RGB8[CS]{
		R: row[pixelOffset+order.R],
		G: row[pixelOffset+order.G],
		B: row[pixelOffset+order.B],
	}
}

// CopyPixel copies a pixel without blending
func (pf *PixFmtAlphaBlendRGB[B, CS, O]) CopyPixel(x, y int, c color.RGB8[CS]) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}

	row := buffer.RowU8(pf.rbuf, y)
	pixelOffset := x * 3
	if pixelOffset+2 >= len(row) {
		return
	}

	// Use the order type parameter to set correct color order
	order := getRGBColorOrder[O]()
	row[pixelOffset+order.R] = c.R
	row[pixelOffset+order.G] = c.G
	row[pixelOffset+order.B] = c.B
}

// BlendPixel blends a pixel with the given alpha and coverage
func (pf *PixFmtAlphaBlendRGB[B, CS, O]) BlendPixel(x, y int, c color.RGB8[CS], alpha, cover basics.Int8u) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}

	row := buffer.RowU8(pf.rbuf, y)
	pixelOffset := x * 3
	if pixelOffset+2 >= len(row) {
		return
	}

	// Use interface assertion for blending
	if blender, ok := interface{}(pf.blender).(RGBBlender); ok {
		blender.BlendPix(row[pixelOffset:pixelOffset+3], c.R, c.G, c.B, alpha, cover)
	}
}

// BlendPixelRGBA blends an RGBA pixel (ignores alpha channel for storage)
func (pf *PixFmtAlphaBlendRGB[B, CS, O]) BlendPixelRGBA(x, y int, c color.RGBA8[CS], cover basics.Int8u) {
	pf.BlendPixel(x, y, color.RGB8[CS]{R: c.R, G: c.G, B: c.B}, c.A, cover)
}

// CopyHline copies a horizontal line
func (pf *PixFmtAlphaBlendRGB[B, CS, O]) CopyHline(x1, y, x2 int, c color.RGB8[CS]) {
	if y < 0 || y >= pf.Height() {
		return
	}

	x1 = ClampX(x1, pf.Width())
	x2 = ClampX(x2, pf.Width())
	if x1 > x2 {
		x1, x2 = x2, x1
	}

	row := buffer.RowU8(pf.rbuf, y)
	for x := x1; x <= x2; x++ {
		pixelOffset := x * 3
		if pixelOffset+2 < len(row) {
			row[pixelOffset] = c.R
			row[pixelOffset+1] = c.G
			row[pixelOffset+2] = c.B
		}
	}
}

// BlendHline blends a horizontal line
func (pf *PixFmtAlphaBlendRGB[B, CS, O]) BlendHline(x1, y, x2 int, c color.RGB8[CS], alpha, cover basics.Int8u) {
	if y < 0 || y >= pf.Height() {
		return
	}

	x1 = ClampX(x1, pf.Width())
	x2 = ClampX(x2, pf.Width())
	if x1 > x2 {
		x1, x2 = x2, x1
	}

	row := buffer.RowU8(pf.rbuf, y)
	if blender, ok := interface{}(pf.blender).(RGBBlender); ok {
		for x := x1; x <= x2; x++ {
			pixelOffset := x * 3
			if pixelOffset+2 < len(row) {
				blender.BlendPix(row[pixelOffset:pixelOffset+3], c.R, c.G, c.B, alpha, cover)
			}
		}
	}
}

// CopyVline copies a vertical line
func (pf *PixFmtAlphaBlendRGB[B, CS, O]) CopyVline(x, y1, y2 int, c color.RGB8[CS]) {
	if x < 0 || x >= pf.Width() {
		return
	}

	y1 = ClampY(y1, pf.Height())
	y2 = ClampY(y2, pf.Height())
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	for y := y1; y <= y2; y++ {
		pf.CopyPixel(x, y, c)
	}
}

// BlendVline blends a vertical line
func (pf *PixFmtAlphaBlendRGB[B, CS, O]) BlendVline(x, y1, y2 int, c color.RGB8[CS], alpha, cover basics.Int8u) {
	if x < 0 || x >= pf.Width() {
		return
	}

	y1 = ClampY(y1, pf.Height())
	y2 = ClampY(y2, pf.Height())
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	for y := y1; y <= y2; y++ {
		pf.BlendPixel(x, y, c, alpha, cover)
	}
}

// CopyBar copies a filled rectangle
func (pf *PixFmtAlphaBlendRGB[B, CS, O]) CopyBar(x1, y1, x2, y2 int, c color.RGB8[CS]) {
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
func (pf *PixFmtAlphaBlendRGB[B, CS, O]) BlendBar(x1, y1, x2, y2 int, c color.RGB8[CS], alpha, cover basics.Int8u) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if x1 > x2 {
		x1, x2 = x2, x1
	}

	for y := y1; y <= y2; y++ {
		pf.BlendHline(x1, y, x2, c, alpha, cover)
	}
}

// BlendSolidHspan blends a horizontal span with varying coverage
func (pf *PixFmtAlphaBlendRGB[B, CS, O]) BlendSolidHspan(x, y, length int, c color.RGB8[CS], alpha basics.Int8u, covers []basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 {
		return
	}

	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}

	row := buffer.RowU8(pf.rbuf, y)
	if blender, ok := interface{}(pf.blender).(RGBBlender); ok {
		if covers == nil {
			// Uniform coverage
			for i := 0; i < length; i++ {
				pixelOffset := (x + i) * 3
				if pixelOffset+2 < len(row) {
					blender.BlendPix(row[pixelOffset:pixelOffset+3], c.R, c.G, c.B, alpha, 255)
				}
			}
		} else {
			// Varying coverage
			for i := 0; i < length && i < len(covers); i++ {
				if covers[i] > 0 {
					pixelOffset := (x + i) * 3
					if pixelOffset+2 < len(row) {
						blender.BlendPix(row[pixelOffset:pixelOffset+3], c.R, c.G, c.B, alpha, covers[i])
					}
				}
			}
		}
	}
}

// BlendSolidVspan blends a vertical span with varying coverage
func (pf *PixFmtAlphaBlendRGB[B, CS, O]) BlendSolidVspan(x, y, length int, c color.RGB8[CS], alpha basics.Int8u, covers []basics.Int8u) {
	if x < 0 || x >= pf.Width() || length <= 0 {
		return
	}

	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}

	if covers == nil {
		// Uniform coverage
		for i := 0; i < length; i++ {
			pf.BlendPixel(x, y+i, c, alpha, 255)
		}
	} else {
		// Varying coverage
		for i := 0; i < length && i < len(covers); i++ {
			if covers[i] > 0 {
				pf.BlendPixel(x, y+i, c, alpha, covers[i])
			}
		}
	}
}

// Clear clears the entire buffer with the given color
func (pf *PixFmtAlphaBlendRGB[B, CS, O]) Clear(c color.RGB8[CS]) {
	for y := 0; y < pf.Height(); y++ {
		row := buffer.RowU8(pf.rbuf, y)
		for x := 0; x < pf.Width(); x++ {
			pixelOffset := x * 3
			if pixelOffset+2 < len(row) {
				row[pixelOffset] = c.R
				row[pixelOffset+1] = c.G
				row[pixelOffset+2] = c.B
			}
		}
	}
}

// Fill is an alias for Clear (fills entire buffer)
func (pf *PixFmtAlphaBlendRGB[B, CS, O]) Fill(c color.RGB8[CS]) {
	pf.Clear(c)
}

// CopyFrom copies from another RGB pixel format
func (pf *PixFmtAlphaBlendRGB[B, CS, O]) CopyFrom(src *PixFmtAlphaBlendRGB[B, CS, O], srcX, srcY, dstX, dstY, width, height int) {
	// Clamp source and destination rectangles
	if srcX < 0 {
		width += srcX
		dstX -= srcX
		srcX = 0
	}
	if srcY < 0 {
		height += srcY
		dstY -= srcY
		srcY = 0
	}
	if dstX < 0 {
		width += dstX
		srcX -= dstX
		dstX = 0
	}
	if dstY < 0 {
		height += dstY
		srcY -= dstY
		dstY = 0
	}

	if srcX+width > src.Width() {
		width = src.Width() - srcX
	}
	if srcY+height > src.Height() {
		height = src.Height() - srcY
	}
	if dstX+width > pf.Width() {
		width = pf.Width() - dstX
	}
	if dstY+height > pf.Height() {
		height = pf.Height() - dstY
	}

	if width <= 0 || height <= 0 {
		return
	}

	// Copy pixel by pixel
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := src.GetPixel(srcX+x, srcY+y)
			pf.CopyPixel(dstX+x, dstY+y, pixel)
		}
	}
}

// Concrete RGB pixel format types for different color orders
type (
	PixFmtRGB24  = PixFmtAlphaBlendRGB[BlenderRGB24, color.Linear, color.RGB24Order]
	PixFmtBGR24  = PixFmtAlphaBlendRGB[BlenderBGR24, color.Linear, color.BGR24Order]
	PixFmtSRGB24 = PixFmtAlphaBlendRGB[BlenderRGB24SRGB, color.SRGB, color.RGB24Order]
	PixFmtSBGR24 = PixFmtAlphaBlendRGB[BlenderBGR24SRGB, color.SRGB, color.BGR24Order]
)

// Constructor functions for RGB24 pixel formats
func NewPixFmtRGB24(rbuf *buffer.RenderingBufferU8) *PixFmtRGB24 {
	return NewPixFmtAlphaBlendRGB[BlenderRGB24, color.Linear, color.RGB24Order](rbuf, BlenderRGB24{})
}

func NewPixFmtBGR24(rbuf *buffer.RenderingBufferU8) *PixFmtBGR24 {
	return NewPixFmtAlphaBlendRGB[BlenderBGR24, color.Linear, color.BGR24Order](rbuf, BlenderBGR24{})
}

func NewPixFmtSRGB24(rbuf *buffer.RenderingBufferU8) *PixFmtSRGB24 {
	return NewPixFmtAlphaBlendRGB[BlenderRGB24SRGB, color.SRGB, color.RGB24Order](rbuf, BlenderRGB24SRGB{})
}

func NewPixFmtSBGR24(rbuf *buffer.RenderingBufferU8) *PixFmtSBGR24 {
	return NewPixFmtAlphaBlendRGB[BlenderBGR24SRGB, color.SRGB, color.BGR24Order](rbuf, BlenderBGR24SRGB{})
}

// ==============================================================================
// RGB48 (16-bit per channel) Pixel Formats
//==============================================================================

// PixFmtAlphaBlendRGB48 represents RGB pixel format with 16-bit components (6 bytes per pixel)
type PixFmtAlphaBlendRGB48[B any, CS any, O any] struct {
	rbuf     *buffer.RenderingBufferU16
	blender  B
	category PixFmtRGBTag
}

// NewPixFmtAlphaBlendRGB48 creates a new RGB48 pixel format
func NewPixFmtAlphaBlendRGB48[B any, CS any, O any](rbuf *buffer.RenderingBufferU16, blender B) *PixFmtAlphaBlendRGB48[B, CS, O] {
	return &PixFmtAlphaBlendRGB48[B, CS, O]{
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

	order := getRGBColorOrder[O]()
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

	order := getRGBColorOrder[O]()
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

	// Use interface assertion for blending
	if blender, ok := interface{}(pf.blender).(RGB48Blender); ok {
		blender.BlendPix(row[pixelOffset:pixelOffset+3], c.R, c.G, c.B, alpha, cover)
	}
}

// Clear clears the entire buffer with the given color
func (pf *PixFmtAlphaBlendRGB48[B, CS, O]) Clear(c color.RGB16[CS]) {
	for y := 0; y < pf.Height(); y++ {
		row := buffer.RowU16(pf.rbuf, y)
		for x := 0; x < pf.Width(); x++ {
			pixelOffset := x * 3
			if pixelOffset+2 < len(row) {
				order := getRGBColorOrder[O]()
				row[pixelOffset+order.R] = c.R
				row[pixelOffset+order.G] = c.G
				row[pixelOffset+order.B] = c.B
			}
		}
	}
}

// Concrete RGB48 pixel format types
type (
	PixFmtRGB48Linear = PixFmtAlphaBlendRGB48[BlenderRGB48Linear, color.Linear, color.RGB24Order]
	PixFmtBGR48Linear = PixFmtAlphaBlendRGB48[BlenderBGR48Linear, color.Linear, color.BGR24Order]
	PixFmtRGB48SRGB   = PixFmtAlphaBlendRGB48[BlenderRGB48SRGB, color.SRGB, color.RGB24Order]
	PixFmtBGR48SRGB   = PixFmtAlphaBlendRGB48[BlenderBGR48SRGB, color.SRGB, color.BGR24Order]
)

// Constructor functions for RGB48 pixel formats
func NewPixFmtRGB48Linear(rbuf *buffer.RenderingBufferU16) *PixFmtRGB48Linear {
	return NewPixFmtAlphaBlendRGB48[BlenderRGB48Linear, color.Linear, color.RGB24Order](rbuf, BlenderRGB48Linear{})
}

func NewPixFmtBGR48Linear(rbuf *buffer.RenderingBufferU16) *PixFmtBGR48Linear {
	return NewPixFmtAlphaBlendRGB48[BlenderBGR48Linear, color.Linear, color.BGR24Order](rbuf, BlenderBGR48Linear{})
}

func NewPixFmtRGB48SRGB(rbuf *buffer.RenderingBufferU16) *PixFmtRGB48SRGB {
	return NewPixFmtAlphaBlendRGB48[BlenderRGB48SRGB, color.SRGB, color.RGB24Order](rbuf, BlenderRGB48SRGB{})
}

func NewPixFmtBGR48SRGB(rbuf *buffer.RenderingBufferU16) *PixFmtBGR48SRGB {
	return NewPixFmtAlphaBlendRGB48[BlenderBGR48SRGB, color.SRGB, color.BGR24Order](rbuf, BlenderBGR48SRGB{})
}

// ==============================================================================
// RGB96 (32-bit float per channel) Pixel Formats
//==============================================================================

// PixFmtAlphaBlendRGB96 represents RGB pixel format with 32-bit float components (12 bytes per pixel)
type PixFmtAlphaBlendRGB96[B any, CS any, O any] struct {
	rbuf     *buffer.RenderingBufferF32
	blender  B
	category PixFmtRGBTag
}

// NewPixFmtAlphaBlendRGB96 creates a new RGB96 pixel format
func NewPixFmtAlphaBlendRGB96[B any, CS any, O any](rbuf *buffer.RenderingBufferF32, blender B) *PixFmtAlphaBlendRGB96[B, CS, O] {
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

	order := getRGBColorOrder[O]()
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

	order := getRGBColorOrder[O]()
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

	// Use interface assertion for blending
	if blender, ok := interface{}(pf.blender).(RGB96Blender); ok {
		blender.BlendPix(row[pixelOffset:pixelOffset+3], c.R, c.G, c.B, alpha, cover)
	}
}

// Clear clears the entire buffer with the given color
func (pf *PixFmtAlphaBlendRGB96[B, CS, O]) Clear(c color.RGB32[CS]) {
	for y := 0; y < pf.Height(); y++ {
		row := buffer.RowF32(pf.rbuf, y)
		for x := 0; x < pf.Width(); x++ {
			pixelOffset := x * 3
			if pixelOffset+2 < len(row) {
				order := getRGBColorOrder[O]()
				row[pixelOffset+order.R] = c.R
				row[pixelOffset+order.G] = c.G
				row[pixelOffset+order.B] = c.B
			}
		}
	}
}

// Concrete RGB96 pixel format types
type (
	PixFmtRGB96Linear = PixFmtAlphaBlendRGB96[BlenderRGB96Linear, color.Linear, color.RGB24Order]
	PixFmtBGR96Linear = PixFmtAlphaBlendRGB96[BlenderBGR96Linear, color.Linear, color.BGR24Order]
	PixFmtRGB96SRGB   = PixFmtAlphaBlendRGB96[BlenderRGB96SRGB, color.SRGB, color.RGB24Order]
	PixFmtBGR96SRGB   = PixFmtAlphaBlendRGB96[BlenderBGR96SRGB, color.SRGB, color.BGR24Order]
)

// Constructor functions for RGB96 pixel formats
func NewPixFmtRGB96Linear(rbuf *buffer.RenderingBufferF32) *PixFmtRGB96Linear {
	return NewPixFmtAlphaBlendRGB96[BlenderRGB96Linear, color.Linear, color.RGB24Order](rbuf, BlenderRGB96Linear{})
}

func NewPixFmtBGR96Linear(rbuf *buffer.RenderingBufferF32) *PixFmtBGR96Linear {
	return NewPixFmtAlphaBlendRGB96[BlenderBGR96Linear, color.Linear, color.BGR24Order](rbuf, BlenderBGR96Linear{})
}

func NewPixFmtRGB96SRGB(rbuf *buffer.RenderingBufferF32) *PixFmtRGB96SRGB {
	return NewPixFmtAlphaBlendRGB96[BlenderRGB96SRGB, color.SRGB, color.RGB24Order](rbuf, BlenderRGB96SRGB{})
}

func NewPixFmtBGR96SRGB(rbuf *buffer.RenderingBufferF32) *PixFmtBGR96SRGB {
	return NewPixFmtAlphaBlendRGB96[BlenderBGR96SRGB, color.SRGB, color.BGR24Order](rbuf, BlenderBGR96SRGB{})
}

// ==============================================================================
// RGBX32/XRGB32 (RGB with padding byte) Pixel Formats
//==============================================================================

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
type PixFmtAlphaBlendRGBX32[B any, CS any, O any] struct {
	rbuf     *buffer.RenderingBufferU8
	blender  B
	category PixFmtRGBTag
}

// NewPixFmtAlphaBlendRGBX32 creates a new RGBX32 pixel format
func NewPixFmtAlphaBlendRGBX32[B any, CS any, O any](rbuf *buffer.RenderingBufferU8, blender B) *PixFmtAlphaBlendRGBX32[B, CS, O] {
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
	if blender, ok := any(pf.blender).(RGBBlender); ok {
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
	PixFmtRGBX32 = PixFmtAlphaBlendRGBX32[BlenderRGB24, color.Linear, RGBX32Order]
	PixFmtXRGB32 = PixFmtAlphaBlendRGBX32[BlenderRGB24, color.Linear, XRGB32Order]
	PixFmtBGRX32 = PixFmtAlphaBlendRGBX32[BlenderBGR24, color.Linear, BGRX32Order]
	PixFmtXBGR32 = PixFmtAlphaBlendRGBX32[BlenderBGR24, color.Linear, XBGR32Order]

	PixFmtSRGBX32 = PixFmtAlphaBlendRGBX32[BlenderRGB24SRGB, color.SRGB, RGBX32Order]
	PixFmtSXRGB32 = PixFmtAlphaBlendRGBX32[BlenderRGB24SRGB, color.SRGB, XRGB32Order]
	PixFmtSBGRX32 = PixFmtAlphaBlendRGBX32[BlenderBGR24SRGB, color.SRGB, BGRX32Order]
	PixFmtSXBGR32 = PixFmtAlphaBlendRGBX32[BlenderBGR24SRGB, color.SRGB, XBGR32Order]
)

// Constructor functions for RGBX32 pixel formats
func NewPixFmtRGBX32(rbuf *buffer.RenderingBufferU8) *PixFmtRGBX32 {
	return NewPixFmtAlphaBlendRGBX32[BlenderRGB24, color.Linear, RGBX32Order](rbuf, BlenderRGB24{})
}

func NewPixFmtXRGB32(rbuf *buffer.RenderingBufferU8) *PixFmtXRGB32 {
	return NewPixFmtAlphaBlendRGBX32[BlenderRGB24, color.Linear, XRGB32Order](rbuf, BlenderRGB24{})
}

func NewPixFmtBGRX32(rbuf *buffer.RenderingBufferU8) *PixFmtBGRX32 {
	return NewPixFmtAlphaBlendRGBX32[BlenderBGR24, color.Linear, BGRX32Order](rbuf, BlenderBGR24{})
}

func NewPixFmtXBGR32(rbuf *buffer.RenderingBufferU8) *PixFmtXBGR32 {
	return NewPixFmtAlphaBlendRGBX32[BlenderBGR24, color.Linear, XBGR32Order](rbuf, BlenderBGR24{})
}

func NewPixFmtSRGBX32(rbuf *buffer.RenderingBufferU8) *PixFmtSRGBX32 {
	return NewPixFmtAlphaBlendRGBX32[BlenderRGB24SRGB, color.SRGB, RGBX32Order](rbuf, BlenderRGB24SRGB{})
}

func NewPixFmtSXRGB32(rbuf *buffer.RenderingBufferU8) *PixFmtSXRGB32 {
	return NewPixFmtAlphaBlendRGBX32[BlenderRGB24SRGB, color.SRGB, XRGB32Order](rbuf, BlenderRGB24SRGB{})
}

func NewPixFmtSBGRX32(rbuf *buffer.RenderingBufferU8) *PixFmtSBGRX32 {
	return NewPixFmtAlphaBlendRGBX32[BlenderBGR24SRGB, color.SRGB, BGRX32Order](rbuf, BlenderBGR24SRGB{})
}

func NewPixFmtSXBGR32(rbuf *buffer.RenderingBufferU8) *PixFmtSXBGR32 {
	return NewPixFmtAlphaBlendRGBX32[BlenderBGR24SRGB, color.SRGB, XBGR32Order](rbuf, BlenderBGR24SRGB{})
}

// ==============================================================================
// Premultiplied RGB Pixel Format Types
//==============================================================================

// RGB24 premultiplied variants
type (
	PixFmtRGB24Pre  = PixFmtAlphaBlendRGB[BlenderRGB24Pre, color.Linear, color.RGB24Order]
	PixFmtBGR24Pre  = PixFmtAlphaBlendRGB[BlenderBGR24Pre, color.Linear, color.BGR24Order]
	PixFmtSRGB24Pre = PixFmtAlphaBlendRGB[BlenderRGB24PreSRGB, color.SRGB, color.RGB24Order]
	PixFmtSBGR24Pre = PixFmtAlphaBlendRGB[BlenderBGR24PreSRGB, color.SRGB, color.BGR24Order]
)

// RGB48 premultiplied variants
type (
	PixFmtRGB48Pre     = PixFmtAlphaBlendRGB48[BlenderRGB48PreLinear, color.Linear, color.RGB24Order]
	PixFmtBGR48Pre     = PixFmtAlphaBlendRGB48[BlenderBGR48PreLinear, color.Linear, color.BGR24Order]
	PixFmtRGB48PreSRGB = PixFmtAlphaBlendRGB48[BlenderRGB48PreSRGB, color.SRGB, color.RGB24Order]
	PixFmtBGR48PreSRGB = PixFmtAlphaBlendRGB48[BlenderBGR48PreSRGB, color.SRGB, color.BGR24Order]
)

// RGB96 premultiplied variants
type (
	PixFmtRGB96Pre     = PixFmtAlphaBlendRGB96[BlenderRGB96PreLinear, color.Linear, color.RGB24Order]
	PixFmtBGR96Pre     = PixFmtAlphaBlendRGB96[BlenderBGR96PreLinear, color.Linear, color.BGR24Order]
	PixFmtRGB96PreSRGB = PixFmtAlphaBlendRGB96[BlenderRGB96PreSRGB, color.SRGB, color.RGB24Order]
	PixFmtBGR96PreSRGB = PixFmtAlphaBlendRGB96[BlenderBGR96PreSRGB, color.SRGB, color.BGR24Order]
)

// RGBX32 premultiplied variants
type (
	PixFmtRGBX32Pre = PixFmtAlphaBlendRGBX32[BlenderRGB24Pre, color.Linear, RGBX32Order]
	PixFmtXRGB32Pre = PixFmtAlphaBlendRGBX32[BlenderRGB24Pre, color.Linear, XRGB32Order]
	PixFmtBGRX32Pre = PixFmtAlphaBlendRGBX32[BlenderBGR24Pre, color.Linear, BGRX32Order]
	PixFmtXBGR32Pre = PixFmtAlphaBlendRGBX32[BlenderBGR24Pre, color.Linear, XBGR32Order]

	PixFmtSRGBX32Pre = PixFmtAlphaBlendRGBX32[BlenderRGB24PreSRGB, color.SRGB, RGBX32Order]
	PixFmtSXRGB32Pre = PixFmtAlphaBlendRGBX32[BlenderRGB24PreSRGB, color.SRGB, XRGB32Order]
	PixFmtSBGRX32Pre = PixFmtAlphaBlendRGBX32[BlenderBGR24PreSRGB, color.SRGB, BGRX32Order]
	PixFmtSXBGR32Pre = PixFmtAlphaBlendRGBX32[BlenderBGR24PreSRGB, color.SRGB, XBGR32Order]
)

// Constructor functions for premultiplied RGB24 formats
func NewPixFmtRGB24Pre(rbuf *buffer.RenderingBufferU8) *PixFmtRGB24Pre {
	return NewPixFmtAlphaBlendRGB[BlenderRGB24Pre, color.Linear, color.RGB24Order](rbuf, BlenderRGB24Pre{})
}

func NewPixFmtBGR24Pre(rbuf *buffer.RenderingBufferU8) *PixFmtBGR24Pre {
	return NewPixFmtAlphaBlendRGB[BlenderBGR24Pre, color.Linear, color.BGR24Order](rbuf, BlenderBGR24Pre{})
}

func NewPixFmtSRGB24Pre(rbuf *buffer.RenderingBufferU8) *PixFmtSRGB24Pre {
	return NewPixFmtAlphaBlendRGB[BlenderRGB24PreSRGB, color.SRGB, color.RGB24Order](rbuf, BlenderRGB24PreSRGB{})
}

func NewPixFmtSBGR24Pre(rbuf *buffer.RenderingBufferU8) *PixFmtSBGR24Pre {
	return NewPixFmtAlphaBlendRGB[BlenderBGR24PreSRGB, color.SRGB, color.BGR24Order](rbuf, BlenderBGR24PreSRGB{})
}

// Constructor functions for premultiplied RGB48 formats
func NewPixFmtRGB48Pre(rbuf *buffer.RenderingBufferU16) *PixFmtRGB48Pre {
	return NewPixFmtAlphaBlendRGB48[BlenderRGB48PreLinear, color.Linear, color.RGB24Order](rbuf, BlenderRGB48PreLinear{})
}

func NewPixFmtBGR48Pre(rbuf *buffer.RenderingBufferU16) *PixFmtBGR48Pre {
	return NewPixFmtAlphaBlendRGB48[BlenderBGR48PreLinear, color.Linear, color.BGR24Order](rbuf, BlenderBGR48PreLinear{})
}

func NewPixFmtRGB48PreSRGB(rbuf *buffer.RenderingBufferU16) *PixFmtRGB48PreSRGB {
	return NewPixFmtAlphaBlendRGB48[BlenderRGB48PreSRGB, color.SRGB, color.RGB24Order](rbuf, BlenderRGB48PreSRGB{})
}

func NewPixFmtBGR48PreSRGB(rbuf *buffer.RenderingBufferU16) *PixFmtBGR48PreSRGB {
	return NewPixFmtAlphaBlendRGB48[BlenderBGR48PreSRGB, color.SRGB, color.BGR24Order](rbuf, BlenderBGR48PreSRGB{})
}

// Constructor functions for premultiplied RGB96 formats
func NewPixFmtRGB96Pre(rbuf *buffer.RenderingBufferF32) *PixFmtRGB96Pre {
	return NewPixFmtAlphaBlendRGB96[BlenderRGB96PreLinear, color.Linear, color.RGB24Order](rbuf, BlenderRGB96PreLinear{})
}

func NewPixFmtBGR96Pre(rbuf *buffer.RenderingBufferF32) *PixFmtBGR96Pre {
	return NewPixFmtAlphaBlendRGB96[BlenderBGR96PreLinear, color.Linear, color.BGR24Order](rbuf, BlenderBGR96PreLinear{})
}

func NewPixFmtRGB96PreSRGB(rbuf *buffer.RenderingBufferF32) *PixFmtRGB96PreSRGB {
	return NewPixFmtAlphaBlendRGB96[BlenderRGB96PreSRGB, color.SRGB, color.RGB24Order](rbuf, BlenderRGB96PreSRGB{})
}

func NewPixFmtBGR96PreSRGB(rbuf *buffer.RenderingBufferF32) *PixFmtBGR96PreSRGB {
	return NewPixFmtAlphaBlendRGB96[BlenderBGR96PreSRGB, color.SRGB, color.BGR24Order](rbuf, BlenderBGR96PreSRGB{})
}

// Constructor functions for premultiplied RGBX32 formats
func NewPixFmtRGBX32Pre(rbuf *buffer.RenderingBufferU8) *PixFmtRGBX32Pre {
	return NewPixFmtAlphaBlendRGBX32[BlenderRGB24Pre, color.Linear, RGBX32Order](rbuf, BlenderRGB24Pre{})
}

func NewPixFmtXRGB32Pre(rbuf *buffer.RenderingBufferU8) *PixFmtXRGB32Pre {
	return NewPixFmtAlphaBlendRGBX32[BlenderRGB24Pre, color.Linear, XRGB32Order](rbuf, BlenderRGB24Pre{})
}

func NewPixFmtBGRX32Pre(rbuf *buffer.RenderingBufferU8) *PixFmtBGRX32Pre {
	return NewPixFmtAlphaBlendRGBX32[BlenderBGR24Pre, color.Linear, BGRX32Order](rbuf, BlenderBGR24Pre{})
}

func NewPixFmtXBGR32Pre(rbuf *buffer.RenderingBufferU8) *PixFmtXBGR32Pre {
	return NewPixFmtAlphaBlendRGBX32[BlenderBGR24Pre, color.Linear, XBGR32Order](rbuf, BlenderBGR24Pre{})
}
