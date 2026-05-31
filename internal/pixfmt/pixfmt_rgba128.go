package pixfmt

import (
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/pixfmt/blender"
)

// PixFmtAlphaBlendRGBA128 is the float (128-bit, 4 x float32) RGBA pixel format.
//
// It is the structural twin of PixFmtAlphaBlendGray32 (float gray) and the float
// analogue of the 8-bit PixFmtAlphaBlendRGBA. It binds a float32 rendering buffer
// to an RGBA128 blender, so channel ordering and premultiplied-vs-plain source
// semantics come from the blender while the pixfmt supplies AGG's span-oriented
// drawing operations. It pairs with the float color color.RGBA32 and mirrors C++
// AGG's pixfmt_rgba128. Coverage arrives as basics.Int8u (0..255) per the
// renderer.PixelFormat contract and is normalised to [0,1] for the float blender.
type PixFmtAlphaBlendRGBA128[B blender.RGBA128Blender[CS], CS color.Space] struct {
	rbuf    *buffer.RenderingBufferF32
	blender B
}

// NewPixFmtAlphaBlendRGBA128 creates an RGBA128 pixfmt over rbuf using blender b.
func NewPixFmtAlphaBlendRGBA128[B blender.RGBA128Blender[CS], CS color.Space](rbuf *buffer.RenderingBufferF32, b B) *PixFmtAlphaBlendRGBA128[B, CS] {
	return &PixFmtAlphaBlendRGBA128[B, CS]{rbuf: rbuf, blender: b}
}

// Basic properties

func (pf *PixFmtAlphaBlendRGBA128[B, CS]) Width() int  { return pf.rbuf.Width() }
func (pf *PixFmtAlphaBlendRGBA128[B, CS]) Height() int { return pf.rbuf.Height() }

// PixWidth returns the storage width of one pixel in bytes (4 x float32).
func (pf *PixFmtAlphaBlendRGBA128[B, CS]) PixWidth() int { return 16 }

func (pf *PixFmtAlphaBlendRGBA128[B, CS]) Stride() int { return pf.rbuf.Stride() }

// RowPtr returns the float32 slice backing row y.
func (pf *PixFmtAlphaBlendRGBA128[B, CS]) RowPtr(y int) []float32 {
	return buffer.RowF32(pf.rbuf, y)
}

// pixelSlice returns the 4-float slice for pixel (x,y), or nil if out of range.
func (pf *PixFmtAlphaBlendRGBA128[B, CS]) pixelSlice(x, y int) []float32 {
	row := buffer.RowF32(pf.rbuf, y)
	off := x * 4
	if off < 0 || off+4 > len(row) {
		return nil
	}
	return row[off : off+4]
}

// Pixel operations

// GetPixel reads a pixel back through the blender's plain-color accessor.
func (pf *PixFmtAlphaBlendRGBA128[B, CS]) GetPixel(x, y int) color.RGBA32[CS] {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return color.RGBA32[CS]{}
	}
	px := pf.pixelSlice(x, y)
	if px == nil {
		return color.RGBA32[CS]{}
	}
	r, g, b, a := pf.blender.GetPlain(px)
	return color.RGBA32[CS]{R: r, G: g, B: b, A: a}
}

// Pixel is an alias for GetPixel (satisfies renderer.PixelFormat).
func (pf *PixFmtAlphaBlendRGBA128[B, CS]) Pixel(x, y int) color.RGBA32[CS] {
	return pf.GetPixel(x, y)
}

// CopyPixel writes c directly, bypassing alpha compositing.
func (pf *PixFmtAlphaBlendRGBA128[B, CS]) CopyPixel(x, y int, c color.RGBA32[CS]) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}
	if px := pf.pixelSlice(x, y); px != nil {
		pf.blender.SetPlain(px, c.R, c.G, c.B, c.A)
	}
}

// BlendPixel applies the blender's per-pixel compositing rule with cover.
func (pf *PixFmtAlphaBlendRGBA128[B, CS]) BlendPixel(x, y int, c color.RGBA32[CS], cover basics.Int8u) {
	if !InBounds(x, y, pf.Width(), pf.Height()) || c.A <= 0 {
		return
	}
	if px := pf.pixelSlice(x, y); px != nil {
		pf.blender.BlendPix(px, c.R, c.G, c.B, c.A, coverToF32(cover))
	}
}

// Line operations

// CopyHline writes a solid horizontal run without blending.
func (pf *PixFmtAlphaBlendRGBA128[B, CS]) CopyHline(x, y, length int, c color.RGBA32[CS]) {
	if y < 0 || y >= pf.Height() || length <= 0 {
		return
	}
	var ok bool
	x, length, _, ok = clipSpan(x, length, pf.Width())
	if !ok {
		return
	}
	row := pf.RowPtr(y)
	for i := range length {
		p := (x + i) * 4
		pf.blender.SetPlain(row[p:p+4], c.R, c.G, c.B, c.A)
	}
}

// BlendHline blends a horizontal line with uniform coverage.
func (pf *PixFmtAlphaBlendRGBA128[B, CS]) BlendHline(x, y, length int, c color.RGBA32[CS], cover basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || c.A <= 0 {
		return
	}
	var ok bool
	x, length, _, ok = clipSpan(x, length, pf.Width())
	if !ok {
		return
	}
	row := pf.RowPtr(y)
	fc := coverToF32(cover)
	for i := range length {
		p := (x + i) * 4
		pf.blender.BlendPix(row[p:p+4], c.R, c.G, c.B, c.A, fc)
	}
}

// CopyVline writes a solid vertical run of the given length without blending.
func (pf *PixFmtAlphaBlendRGBA128[B, CS]) CopyVline(x, y, length int, c color.RGBA32[CS]) {
	if x < 0 || x >= pf.Width() || length <= 0 {
		return
	}
	for i := range length {
		pf.CopyPixel(x, y+i, c)
	}
}

// BlendVline blends a vertical run of the given length with uniform coverage.
func (pf *PixFmtAlphaBlendRGBA128[B, CS]) BlendVline(x, y, length int, c color.RGBA32[CS], cover basics.Int8u) {
	if x < 0 || x >= pf.Width() || length <= 0 || c.A <= 0 {
		return
	}
	for i := range length {
		pf.BlendPixel(x, y+i, c, cover)
	}
}

// Rectangle operations

// CopyBar copies a filled rectangle with inclusive corners (x1,y1)-(x2,y2).
func (pf *PixFmtAlphaBlendRGBA128[B, CS]) CopyBar(x1, y1, x2, y2 int, c color.RGBA32[CS]) {
	x1, y1, x2, y2 = normalizeBar(x1, y1, x2, y2)
	x1 = Max(0, x1)
	y1 = Max(0, y1)
	x2 = Min(pf.Width()-1, x2)
	y2 = Min(pf.Height()-1, y2)
	for y := y1; y <= y2; y++ {
		pf.CopyHline(x1, y, x2-x1+1, c)
	}
}

// BlendBar blends a filled rectangle with inclusive corners and uniform coverage.
func (pf *PixFmtAlphaBlendRGBA128[B, CS]) BlendBar(x1, y1, x2, y2 int, c color.RGBA32[CS], cover basics.Int8u) {
	if c.A <= 0 {
		return
	}
	x1, y1, x2, y2 = normalizeBar(x1, y1, x2, y2)
	x1 = Max(0, x1)
	y1 = Max(0, y1)
	x2 = Min(pf.Width()-1, x2)
	y2 = Min(pf.Height()-1, y2)
	for y := y1; y <= y2; y++ {
		pf.BlendHline(x1, y, x2-x1+1, c, cover)
	}
}

// Span operations

// BlendSolidHspan blends a solid color along a horizontal span with per-pixel coverage.
func (pf *PixFmtAlphaBlendRGBA128[B, CS]) BlendSolidHspan(x, y, length int, c color.RGBA32[CS], covers []basics.Int8u) {
	if y < 0 || y >= pf.Height() || c.A <= 0 || length <= 0 {
		return
	}
	var skip int
	var ok bool
	x, length, skip, ok = clipSpan(x, length, pf.Width())
	if !ok || skip >= len(covers) {
		return
	}
	covers = covers[skip:]
	if length > len(covers) {
		length = len(covers)
	}
	row := pf.RowPtr(y)
	for i := range length {
		if covers[i] > 0 {
			p := (x + i) * 4
			pf.blender.BlendPix(row[p:p+4], c.R, c.G, c.B, c.A, coverToF32(covers[i]))
		}
	}
}

// BlendSolidVspan blends a solid color along a vertical span with per-pixel coverage.
func (pf *PixFmtAlphaBlendRGBA128[B, CS]) BlendSolidVspan(x, y, length int, c color.RGBA32[CS], covers []basics.Int8u) {
	if x < 0 || x >= pf.Width() || c.A <= 0 || length <= 0 {
		return
	}
	var skip int
	var ok bool
	y, length, skip, ok = clipSpan(y, length, pf.Height())
	if !ok || skip >= len(covers) {
		return
	}
	covers = covers[skip:]
	if length > len(covers) {
		length = len(covers)
	}
	for i := range length {
		if covers[i] > 0 {
			pf.BlendPixel(x, y+i, c, covers[i])
		}
	}
}

// CopyColorHspan copies a horizontal span of colors.
func (pf *PixFmtAlphaBlendRGBA128[B, CS]) CopyColorHspan(x, y, length int, colors []color.RGBA32[CS]) {
	if y < 0 || y >= pf.Height() || length <= 0 || len(colors) == 0 {
		return
	}
	var skip int
	var ok bool
	x, length, skip, ok = clipSpan(x, length, pf.Width())
	if !ok || skip >= len(colors) {
		return
	}
	colors = colors[skip:]
	if length > len(colors) {
		length = len(colors)
	}
	for i := range length {
		pf.CopyPixel(x+i, y, colors[i])
	}
}

// BlendColorHspan blends a horizontal span of colors with optional per-pixel
// coverage (covers == nil uses the uniform cover).
func (pf *PixFmtAlphaBlendRGBA128[B, CS]) BlendColorHspan(x, y, length int, colors []color.RGBA32[CS], covers []basics.Int8u, cover basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || len(colors) == 0 {
		return
	}
	var skip int
	var ok bool
	x, length, skip, ok = clipSpan(x, length, pf.Width())
	if !ok || skip >= len(colors) {
		return
	}
	colors = colors[skip:]
	if covers != nil {
		if skip >= len(covers) {
			return
		}
		covers = covers[skip:]
		if length > len(covers) {
			length = len(covers)
		}
	}
	if length > len(colors) {
		length = len(colors)
	}
	for i := range length {
		c := colors[i]
		if c.A <= 0 {
			continue
		}
		cvr := cover
		if covers != nil && i < len(covers) {
			cvr = covers[i]
		}
		pf.BlendPixel(x+i, y, c, cvr)
	}
}

// CopyColorVspan copies a vertical span of colors.
func (pf *PixFmtAlphaBlendRGBA128[B, CS]) CopyColorVspan(x, y, length int, colors []color.RGBA32[CS]) {
	if x < 0 || x >= pf.Width() || length <= 0 || len(colors) == 0 {
		return
	}
	var skip int
	var ok bool
	y, length, skip, ok = clipSpan(y, length, pf.Height())
	if !ok || skip >= len(colors) {
		return
	}
	colors = colors[skip:]
	if length > len(colors) {
		length = len(colors)
	}
	for i := range length {
		pf.CopyPixel(x, y+i, colors[i])
	}
}

// BlendColorVspan blends a vertical span of colors with optional per-pixel coverage.
func (pf *PixFmtAlphaBlendRGBA128[B, CS]) BlendColorVspan(x, y, length int, colors []color.RGBA32[CS], covers []basics.Int8u, cover basics.Int8u) {
	if x < 0 || x >= pf.Width() || length <= 0 || len(colors) == 0 {
		return
	}
	var skip int
	var ok bool
	y, length, skip, ok = clipSpan(y, length, pf.Height())
	if !ok || skip >= len(colors) {
		return
	}
	colors = colors[skip:]
	if covers != nil {
		if skip >= len(covers) {
			return
		}
		covers = covers[skip:]
		if length > len(covers) {
			length = len(covers)
		}
	}
	if length > len(colors) {
		length = len(colors)
	}
	for i := range length {
		c := colors[i]
		if c.A <= 0 {
			continue
		}
		cvr := cover
		if covers != nil && i < len(covers) {
			cvr = covers[i]
		}
		pf.BlendPixel(x, y+i, c, cvr)
	}
}

// Clear operations

// Clear writes c to every pixel (no blending).
func (pf *PixFmtAlphaBlendRGBA128[B, CS]) Clear(c color.RGBA32[CS]) {
	for y := range pf.Height() {
		row := pf.RowPtr(y)
		for x := range pf.Width() {
			p := x * 4
			pf.blender.SetPlain(row[p:p+4], c.R, c.G, c.B, c.A)
		}
	}
}

// Fill is an alias for Clear (matches the gray/rgba pixfmt convention).
func (pf *PixFmtAlphaBlendRGBA128[B, CS]) Fill(c color.RGBA32[CS]) {
	pf.Clear(c)
}

// coverToF32 normalises an 8-bit coverage value to [0,1].
func coverToF32(cover basics.Int8u) float32 {
	return float32(cover) / 255.0
}

// normalizeBar orders the rectangle corners so x1<=x2 and y1<=y2.
func normalizeBar(x1, y1, x2, y2 int) (int, int, int, int) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	return x1, y1, x2, y2
}

// Concrete pixel format types (Linear and sRGB, RGBA channel order).
type (
	PixFmtRGBA128      = PixFmtAlphaBlendRGBA128[blender.BlenderRGBA128Linear, color.Linear]
	PixFmtRGBA128Pre   = PixFmtAlphaBlendRGBA128[blender.BlenderRGBA128PreLinear, color.Linear]
	PixFmtRGBA128Plain = PixFmtAlphaBlendRGBA128[blender.BlenderRGBA128PlainLinear, color.Linear]

	PixFmtSRGBA128      = PixFmtAlphaBlendRGBA128[blender.BlenderRGBA128SRGB, color.SRGB]
	PixFmtSRGBA128Pre   = PixFmtAlphaBlendRGBA128[blender.BlenderRGBA128PreSRGB, color.SRGB]
	PixFmtSRGBA128Plain = PixFmtAlphaBlendRGBA128[blender.BlenderRGBA128PlainSRGB, color.SRGB]
)

// Constructors for the concrete types.

func NewPixFmtRGBA128(rbuf *buffer.RenderingBufferF32) *PixFmtRGBA128 {
	return NewPixFmtAlphaBlendRGBA128[blender.BlenderRGBA128Linear, color.Linear](rbuf, blender.BlenderRGBA128Linear{})
}

func NewPixFmtRGBA128Pre(rbuf *buffer.RenderingBufferF32) *PixFmtRGBA128Pre {
	return NewPixFmtAlphaBlendRGBA128[blender.BlenderRGBA128PreLinear, color.Linear](rbuf, blender.BlenderRGBA128PreLinear{})
}

func NewPixFmtRGBA128Plain(rbuf *buffer.RenderingBufferF32) *PixFmtRGBA128Plain {
	return NewPixFmtAlphaBlendRGBA128[blender.BlenderRGBA128PlainLinear, color.Linear](rbuf, blender.BlenderRGBA128PlainLinear{})
}

func NewPixFmtSRGBA128(rbuf *buffer.RenderingBufferF32) *PixFmtSRGBA128 {
	return NewPixFmtAlphaBlendRGBA128[blender.BlenderRGBA128SRGB, color.SRGB](rbuf, blender.BlenderRGBA128SRGB{})
}

func NewPixFmtSRGBA128Pre(rbuf *buffer.RenderingBufferF32) *PixFmtSRGBA128Pre {
	return NewPixFmtAlphaBlendRGBA128[blender.BlenderRGBA128PreSRGB, color.SRGB](rbuf, blender.BlenderRGBA128PreSRGB{})
}

func NewPixFmtSRGBA128Plain(rbuf *buffer.RenderingBufferF32) *PixFmtSRGBA128Plain {
	return NewPixFmtAlphaBlendRGBA128[blender.BlenderRGBA128PlainSRGB, color.SRGB](rbuf, blender.BlenderRGBA128PlainSRGB{})
}
