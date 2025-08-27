package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/order"
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
type PixFmtAlphaBlendRGBA[
	B blender.RGBABlender[S, O],
	S color.Space,
	O order.RGBAOrder,
] struct {
	rbuf     *buffer.RenderingBufferU8
	blender  B
	category PixFmtRGBATag
}

// NewPixFmtAlphaBlendRGBA creates a new RGBA pixel format
func NewPixFmtAlphaBlendRGBA[
	B blender.RGBABlender[S, O],
	S color.Space,
	O order.RGBAOrder,
](rbuf *buffer.RenderingBufferU8, b B) *PixFmtAlphaBlendRGBA[B, S, O] {
	return &PixFmtAlphaBlendRGBA[B, S, O]{rbuf: rbuf, blender: b}
}

// Width returns the buffer width
func (pf *PixFmtAlphaBlendRGBA[B, S, O]) Width() int {
	return pf.rbuf.Width()
}

// Height returns the buffer height
func (pf *PixFmtAlphaBlendRGBA[B, S, O]) Height() int {
	return pf.rbuf.Height()
}

// PixWidth returns bytes per pixel (4 for RGBA)
func (pf *PixFmtAlphaBlendRGBA[B, S, O]) PixWidth() int {
	return 4
}

// GetPixel returns the pixel at the given coordinates
func (pf *PixFmtAlphaBlendRGBA[B, S, O]) GetPixel(x, y int) color.RGBA8[S] {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return color.RGBA8[S]{}
	}
	row := buffer.RowU8(pf.rbuf, y)
	off := x * 4
	if off+3 >= len(row) {
		return color.RGBA8[S]{}
	}

	r, g, b, a := pf.blender.GetPlain(row[off : off+4])
	return color.RGBA8[S]{R: r, G: g, B: b, A: a}
}

// Pixel returns the pixel at the given coordinates (alias for GetPixel)
func (pf *PixFmtAlphaBlendRGBA[B, S, O]) Pixel(x, y int) color.RGBA8[S] {
	return pf.GetPixel(x, y)
}

// CopyPixel copies a pixel without blending
func (pf *PixFmtAlphaBlendRGBA[B, S, O]) CopyPixel(x, y int, c color.RGBA8[S]) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}

	row := buffer.RowU8(pf.rbuf, y)
	off := x * 4
	if off+3 >= len(row) {
		return
	}

	pf.blender.SetPlain(row[off:off+4], c.R, c.G, c.B, c.A)
}

// BlendPixel blends a pixel with the given coverage
func (pf *PixFmtAlphaBlendRGBA[B, S, O]) BlendPixel(x, y int, c color.RGBA8[S], cover basics.Int8u) {
	if !InBounds(x, y, pf.Width(), pf.Height()) || c.IsTransparent() {
		return
	}

	row := buffer.RowU8(pf.rbuf, y)
	off := x * 4
	if off+3 >= len(row) {
		return
	}

	// Direct blending call - no type assertion needed with proper constraints
	pf.blender.BlendPix(row[off:off+4], c.R, c.G, c.B, c.A, cover)
}

// CopyHline copies a horizontal line
func (pf *PixFmtAlphaBlendRGBA[B, S, O]) CopyHline(x, y, length int, c color.RGBA8[S]) {
	if y < 0 || y >= pf.Height() || length <= 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}

	ir, ig, ib, ia := idxs[O]()
	row := buffer.RowU8(pf.rbuf, y)
	p := x * 4
	for i := 0; i < length; i++ {
		if p+3 < len(row) {
			row[p+ir] = c.R
			row[p+ig] = c.G
			row[p+ib] = c.B
			row[p+ia] = c.A
		}
		p += 4
	}
}

// BlendHline blends a horizontal line
func (pf *PixFmtAlphaBlendRGBA[B, S, O]) BlendHline(x, y, length int, c color.RGBA8[S], cover basics.Int8u) {
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
func (pf *PixFmtAlphaBlendRGBA[B, S, O]) CopyVline(x, y, length int, c color.RGBA8[S]) {
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
func (pf *PixFmtAlphaBlendRGBA[B, S, O]) BlendVline(x, y, length int, c color.RGBA8[S], cover basics.Int8u) {
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

func (pf *PixFmtAlphaBlendRGBA[B, S, O]) CopyBar(x1, y1, x2, y2 int, c color.RGBA8[S]) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	length := x2 - x1 + 1
	for y := y1; y <= y2; y++ {
		pf.CopyHline(x1, y, length, c)
	}
}

func (pf *PixFmtAlphaBlendRGBA[B, S, O]) BlendBar(x1, y1, x2, y2 int, c color.RGBA8[S], cover basics.Int8u) {
	if c.IsTransparent() {
		return
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	length := x2 - x1 + 1
	for y := y1; y <= y2; y++ {
		pf.BlendHline(x1, y, length, c, cover)
	}
}

// BlendSolidHspan blends a horizontal span with varying coverage
func (pf *PixFmtAlphaBlendRGBA[B, S, O]) BlendSolidHspan(x, y, length int, c color.RGBA8[S], covers []basics.Int8u) {
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
func (pf *PixFmtAlphaBlendRGBA[B, S, O]) BlendSolidVspan(x, y, length int, c color.RGBA8[S], covers []basics.Int8u) {
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
func (pf *PixFmtAlphaBlendRGBA[B, S, O]) CopyColorHspan(x, y, length int, colors []color.RGBA8[S]) {
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
func (pf *PixFmtAlphaBlendRGBA[B, S, O]) BlendColorHspan(x, y, length int, colors []color.RGBA8[S], covers []basics.Int8u, cover basics.Int8u) {
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
func (pf *PixFmtAlphaBlendRGBA[B, S, O]) CopyColorVspan(x, y, length int, colors []color.RGBA8[S]) {
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
func (pf *PixFmtAlphaBlendRGBA[B, S, O]) BlendColorVspan(x, y, length int, colors []color.RGBA8[S], covers []basics.Int8u, cover basics.Int8u) {
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
func (pf *PixFmtAlphaBlendRGBA[B, S, O]) Clear(c color.RGBA8[S]) {
	for y := 0; y < pf.Height(); y++ {
		row := buffer.RowU8(pf.rbuf, y)
		for x := 0; x < pf.Width(); x++ {
			off := x * 4
			if off+3 >= len(row) {
				break
			}
			pf.blender.SetPlain(row[off:off+4], c.R, c.G, c.B, c.A)
		}
	}
}

// Fill is an alias for Clear (fills entire buffer)
func (pf *PixFmtAlphaBlendRGBA[B, S, O]) Fill(c color.RGBA8[S]) {
	pf.Clear(c)
}

// Concrete RGBA pixel format types for different color orders
type (
	// Premultiplied framebuffer, plain src blending (AGG blender_rgba)
	PixFmtRGBA32[S color.Space] = PixFmtAlphaBlendRGBA[blender.BlenderRGBA8[S, order.RGBA], S, order.RGBA]
	PixFmtBGRA32[S color.Space] = PixFmtAlphaBlendRGBA[blender.BlenderRGBA8[S, order.BGRA], S, order.BGRA]
	PixFmtARGB32[S color.Space] = PixFmtAlphaBlendRGBA[blender.BlenderRGBA8[S, order.ARGB], S, order.ARGB]
	PixFmtABGR32[S color.Space] = PixFmtAlphaBlendRGBA[blender.BlenderRGBA8[S, order.ABGR], S, order.ABGR]

	// Premultiplied framebuffer, premul-style blending (AGG blender_rgba_pre)
	PixFmtRGBA32Pre[S color.Space] = PixFmtAlphaBlendRGBA[blender.BlenderRGBA8Pre[S, order.RGBA], S, order.RGBA]
	PixFmtBGRA32Pre[S color.Space] = PixFmtAlphaBlendRGBA[blender.BlenderRGBA8Pre[S, order.BGRA], S, order.BGRA]
	PixFmtARGB32Pre[S color.Space] = PixFmtAlphaBlendRGBA[blender.BlenderRGBA8Pre[S, order.ARGB], S, order.ARGB]
	PixFmtABGR32Pre[S color.Space] = PixFmtAlphaBlendRGBA[blender.BlenderRGBA8Pre[S, order.ABGR], S, order.ABGR]

	// Plain framebuffer (AGG blender_rgba_plain)
	PixFmtRGBA32Plain[S color.Space] = PixFmtAlphaBlendRGBA[blender.BlenderRGBA8Plain[S, order.RGBA], S, order.RGBA]
	PixFmtBGRA32Plain[S color.Space] = PixFmtAlphaBlendRGBA[blender.BlenderRGBA8Plain[S, order.BGRA], S, order.BGRA]
	PixFmtARGB32Plain[S color.Space] = PixFmtAlphaBlendRGBA[blender.BlenderRGBA8Plain[S, order.ARGB], S, order.ARGB]
	PixFmtABGR32Plain[S color.Space] = PixFmtAlphaBlendRGBA[blender.BlenderRGBA8Plain[S, order.ABGR], S, order.ABGR]
)

//////////////////////////////////////////////////////////////////////////////////////
// Constructors
//////////////////////////////////////////////////////////////////////////////////////

// Constructors for RGBA pixel formats
func NewPixFmtRGBA32[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtRGBA32[S] {
	return NewPixFmtAlphaBlendRGBA[blender.BlenderRGBA8[S, order.RGBA], S, order.RGBA](rbuf, blender.BlenderRGBA8[S, order.RGBA]{})
}

func NewPixFmtBGRA32[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtBGRA32[S] {
	return NewPixFmtAlphaBlendRGBA[blender.BlenderRGBA8[S, order.BGRA], S, order.BGRA](rbuf, blender.BlenderRGBA8[S, order.BGRA]{})
}

func NewPixFmtARGB32[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtARGB32[S] {
	return NewPixFmtAlphaBlendRGBA[blender.BlenderRGBA8[S, order.ARGB], S, order.ARGB](rbuf, blender.BlenderRGBA8[S, order.ARGB]{})
}

func NewPixFmtABGR32[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtABGR32[S] {
	return NewPixFmtAlphaBlendRGBA[blender.BlenderRGBA8[S, order.ABGR], S, order.ABGR](rbuf, blender.BlenderRGBA8[S, order.ABGR]{})
}

// Constructors for RGBA pixel formats (premultiplied)
func NewPixFmtRGBA32Pre[S color.Space](r *buffer.RenderingBufferU8) *PixFmtRGBA32Pre[S] {
	return NewPixFmtAlphaBlendRGBA[blender.BlenderRGBA8Pre[S, order.RGBA], S, order.RGBA](r, blender.BlenderRGBA8Pre[S, order.RGBA]{})
}

func NewPixFmtBGRA32Pre[S color.Space](r *buffer.RenderingBufferU8) *PixFmtBGRA32Pre[S] {
	return NewPixFmtAlphaBlendRGBA[blender.BlenderRGBA8Pre[S, order.BGRA], S, order.BGRA](r, blender.BlenderRGBA8Pre[S, order.BGRA]{})
}

func NewPixFmtARGB32Pre[S color.Space](r *buffer.RenderingBufferU8) *PixFmtARGB32Pre[S] {
	return NewPixFmtAlphaBlendRGBA[blender.BlenderRGBA8Pre[S, order.ARGB], S, order.ARGB](r, blender.BlenderRGBA8Pre[S, order.ARGB]{})
}

func NewPixFmtABGR32Pre[S color.Space](r *buffer.RenderingBufferU8) *PixFmtABGR32Pre[S] {
	return NewPixFmtAlphaBlendRGBA[blender.BlenderRGBA8Pre[S, order.ABGR], S, order.ABGR](r, blender.BlenderRGBA8Pre[S, order.ABGR]{})
}

// Constructors for RGBA pixel formats (plain)
func NewPixFmtRGBA32Plain[S color.Space](r *buffer.RenderingBufferU8) *PixFmtRGBA32Plain[S] {
	return NewPixFmtAlphaBlendRGBA[blender.BlenderRGBA8Plain[S, order.RGBA], S, order.RGBA](r, blender.BlenderRGBA8Plain[S, order.RGBA]{})
}

func NewPixFmtBGRA32Plain[S color.Space](r *buffer.RenderingBufferU8) *PixFmtBGRA32Plain[S] {
	return NewPixFmtAlphaBlendRGBA[blender.BlenderRGBA8Plain[S, order.BGRA], S, order.BGRA](r, blender.BlenderRGBA8Plain[S, order.BGRA]{})
}

func NewPixFmtARGB32Plain[S color.Space](r *buffer.RenderingBufferU8) *PixFmtARGB32Plain[S] {
	return NewPixFmtAlphaBlendRGBA[blender.BlenderRGBA8Plain[S, order.ARGB], S, order.ARGB](r, blender.BlenderRGBA8Plain[S, order.ARGB]{})
}

func NewPixFmtABGR32Plain[S color.Space](r *buffer.RenderingBufferU8) *PixFmtABGR32Plain[S] {
	return NewPixFmtAlphaBlendRGBA[blender.BlenderRGBA8Plain[S, order.ABGR], S, order.ABGR](r, blender.BlenderRGBA8Plain[S, order.ABGR]{})
}

// Constructors for RGBA pixel formats (linear)

func NewPixFmtRGBA32Linear(r *buffer.RenderingBufferU8) *PixFmtRGBA32[color.Linear] {
	return NewPixFmtRGBA32[color.Linear](r)
}

func NewPixFmtBGRA32Linear(r *buffer.RenderingBufferU8) *PixFmtBGRA32[color.Linear] {
	return NewPixFmtBGRA32[color.Linear](r)
}

func NewPixFmtARGB32Linear(r *buffer.RenderingBufferU8) *PixFmtARGB32[color.Linear] {
	return NewPixFmtARGB32[color.Linear](r)
}

func NewPixFmtABGR32Linear(r *buffer.RenderingBufferU8) *PixFmtABGR32[color.Linear] {
	return NewPixFmtABGR32[color.Linear](r)
}

// Constructors for RGBA pixel formats (linear, premultiplied)

func NewPixFmtRGBA32PreLinear(r *buffer.RenderingBufferU8) *PixFmtRGBA32Pre[color.Linear] {
	return NewPixFmtRGBA32Pre[color.Linear](r)
}

func NewPixFmtBGRA32PreLinear(r *buffer.RenderingBufferU8) *PixFmtBGRA32Pre[color.Linear] {
	return NewPixFmtBGRA32Pre[color.Linear](r)
}

func NewPixFmtARGB32PreLinear(r *buffer.RenderingBufferU8) *PixFmtARGB32Pre[color.Linear] {
	return NewPixFmtARGB32Pre[color.Linear](r)
}

func NewPixFmtABGR32PreLinear(r *buffer.RenderingBufferU8) *PixFmtABGR32Pre[color.Linear] {
	return NewPixFmtABGR32Pre[color.Linear](r)
}

// Constructors for RGBA pixel formats (linear, plain)

func NewPixFmtRGBA32PlainLinear(r *buffer.RenderingBufferU8) *PixFmtRGBA32Plain[color.Linear] {
	return NewPixFmtRGBA32Plain[color.Linear](r)
}

func NewPixFmtBGRA32PlainLinear(r *buffer.RenderingBufferU8) *PixFmtBGRA32Plain[color.Linear] {
	return NewPixFmtBGRA32Plain[color.Linear](r)
}

func NewPixFmtARGB32PlainLinear(r *buffer.RenderingBufferU8) *PixFmtARGB32Plain[color.Linear] {
	return NewPixFmtARGB32Plain[color.Linear](r)
}

func NewPixFmtABGR32PlainLinear(r *buffer.RenderingBufferU8) *PixFmtABGR32Plain[color.Linear] {
	return NewPixFmtABGR32Plain[color.Linear](r)
}

////////////////////////////////////////////////////////////////////////////////
// Helpers
////////////////////////////////////////////////////////////////////////////////

func idxs[O order.RGBAOrder]() (ir, ig, ib, ia int) {
	var o O
	return o.IdxR(), o.IdxG(), o.IdxB(), o.IdxA()
}

func premul(r, g, b, a basics.Int8u) (pr, pg, pb, pa basics.Int8u) {
	if a == 255 || a == 0 {
		return r * (a / 255), g * (a / 255), b * (a / 255), a // fast paths; see below for exact multiply
	}
	return color.RGBA8Multiply(r, a), color.RGBA8Multiply(g, a),
		color.RGBA8Multiply(b, a), a
}

// classic rounded demul: x*255/a
func demul8(x, a basics.Int8u) basics.Int8u {
	if a == 0 {
		return 0
	}
	if a == 255 {
		return x
	}
	return basics.Int8u((uint32(x)*255 + uint32(a)/2) / uint32(a))
}
