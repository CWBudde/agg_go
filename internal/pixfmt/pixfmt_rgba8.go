package pixfmt

import (
	"unsafe"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/order"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt/blender"
	"github.com/MeKo-Christian/agg_go/internal/simd"
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
type PixFmtAlphaBlendRGBA[S color.Space, B blender.RGBABlender[S]] struct {
	rbuf     *buffer.RenderingBufferU8
	blender  B
	category PixFmtRGBATag
}

// NewPixFmtAlphaBlendRGBA creates a new RGBA pixel format
func NewPixFmtAlphaBlendRGBA[S color.Space, B blender.RGBABlender[S]](rbuf *buffer.RenderingBufferU8, b B) *PixFmtAlphaBlendRGBA[S, B] {
	return &PixFmtAlphaBlendRGBA[S, B]{rbuf: rbuf, blender: b}
}

// Width returns the buffer width
func (pf *PixFmtAlphaBlendRGBA[S, B]) Width() int {
	return pf.rbuf.Width()
}

// Height returns the buffer height
func (pf *PixFmtAlphaBlendRGBA[S, B]) Height() int {
	return pf.rbuf.Height()
}

// PixWidth returns bytes per pixel (4 for RGBA)
func (pf *PixFmtAlphaBlendRGBA[S, B]) PixWidth() int {
	return 4
}

// GetPixel returns the pixel at the given coordinates
func (pf *PixFmtAlphaBlendRGBA[S, B]) GetPixel(x, y int) color.RGBA8[S] {
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
func (pf *PixFmtAlphaBlendRGBA[S, B]) Pixel(x, y int) color.RGBA8[S] {
	return pf.GetPixel(x, y)
}

// RowData returns the raw row bytes for transfer operations.
func (pf *PixFmtAlphaBlendRGBA[S, B]) RowData(y int) []basics.Int8u {
	if y < 0 || y >= pf.Height() {
		return nil
	}
	return buffer.RowU8(pf.rbuf, y)
}

// CopyPixel copies a pixel without blending
func (pf *PixFmtAlphaBlendRGBA[S, B]) CopyPixel(x, y int, c color.RGBA8[S]) {
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
func (pf *PixFmtAlphaBlendRGBA[S, B]) BlendPixel(x, y int, c color.RGBA8[S], cover basics.Int8u) {
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
func (pf *PixFmtAlphaBlendRGBA[S, B]) CopyHline(x, y, length int, c color.RGBA8[S]) {
	if y < 0 || y >= pf.Height() || length <= 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}

	row := buffer.RowU8(pf.rbuf, y)
	off := x * 4
	if ro, ok := any(pf.blender).(blender.RawRGBAOrder); ok {
		var pixel [4]byte
		pixel[ro.IdxR()] = c.R
		pixel[ro.IdxG()] = c.G
		pixel[ro.IdxB()] = c.B
		pixel[ro.IdxA()] = c.A
		simd.FillRGBA(row[off:], pixel[0], pixel[1], pixel[2], pixel[3], length)
	} else {
		// Safe path: use blender SetPlain
		for i := 0; i < length; i++ {
			p := off + i*4
			if p+3 < len(row) {
				pf.blender.SetPlain(row[p:p+4], c.R, c.G, c.B, c.A)
			}
		}
	}
}

// BlendHline blends a horizontal line with uniform coverage.
//
// Mirrors C++ AGG's blend_hline optimizations: opaque+full-cover → direct
// copy, full-cover → blend with c.a, partial cover → blend with scaled alpha.
func (pf *PixFmtAlphaBlendRGBA[S, B]) BlendHline(x, y, length int, c color.RGBA8[S], cover basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || c.IsTransparent() {
		return
	}

	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}

	row := buffer.RowU8(pf.rbuf, y)

	// Fast path: resolve order and blend mode once.
	if fb, ok := any(pf.blender).(blender.RGBAFastBlender); ok {
		ir, ig, ib, ia := fb.IdxR(), fb.IdxG(), fb.IdxB(), fb.IdxA()

		spanEnd := (x+length-1)*4 + 3
		if spanEnd >= len(row) {
			goto slow
		}
		_ = row[spanEnd]

		// Mask indices so the compiler knows they are in [0,3] (BCE).
		ir, ig, ib, ia = ir&3, ig&3, ib&3, ia&3

		// SIMD fast path: standard RGBA byte order — handles all coverage cases
		// (opaque+full-cover routes to FillRGBA inside BlendHlineRGBA).
		if ir == 0 && ig == 1 && ib == 2 && ia == 3 {
			simd.BlendHlineRGBA(row[x*4:], c.R, c.G, c.B, c.A, cover, length, fb.PremulSrc())
			return
		}

		// Scalar path for non-RGBA byte-order blenders (e.g., BGRA).
		// C++ optimization: opaque + full cover → direct copy.
		if c.IsOpaque() && cover == basics.CoverFull {
			p := x * 4
			for range length {
				px := row[p : p+4 : p+4] // BCE
				px[ir] = c.R
				px[ig] = c.G
				px[ib] = c.B
				px[ia] = c.A
				p += 4
			}
			return
		}

		if fb.PremulSrc() {
			blendHlinePre(row, x, length, c.R, c.G, c.B, c.A, ir, ig, ib, ia, cover)
		} else {
			blendHlinePlain(row, x, length, c.R, c.G, c.B, c.A, ir, ig, ib, ia, cover)
		}
		return
	}

slow:
	for i := 0; i < length; i++ {
		pixelOffset := (x + i) * 4
		if pixelOffset+3 < len(row) {
			pf.blender.BlendPix(row[pixelOffset:pixelOffset+4], c.R, c.G, c.B, c.A, cover)
		}
	}
}

// blendHlinePlain implements the plain→premul blend loop for uniform coverage.
func blendHlinePlain(
	row []byte, x, length int,
	sr, sg, sb, sa byte,
	ir, ig, ib, ia int,
	cover byte,
) {
	// Pre-compute the effective alpha once (same for every pixel).
	alpha := color.RGBA8MultCover(sa, cover)
	if alpha == 0 {
		return
	}
	ir, ig, ib, ia = ir&3, ig&3, ib&3, ia&3 // BCE: constrain to [0,3]
	p := x * 4
	for range length {
		px := row[p : p+4 : p+4] // BCE: one check per pixel, not four
		px[ir] = color.RGBA8Lerp(px[ir], sr, alpha)
		px[ig] = color.RGBA8Lerp(px[ig], sg, alpha)
		px[ib] = color.RGBA8Lerp(px[ib], sb, alpha)
		px[ia] = color.RGBA8Prelerp(px[ia], alpha, alpha)
		p += 4
	}
}

// blendHlinePre implements the premul→premul blend loop for uniform coverage.
func blendHlinePre(
	row []byte, x, length int,
	sr, sg, sb, sa byte,
	ir, ig, ib, ia int,
	cover byte,
) {
	r, g, b, a := sr, sg, sb, sa
	if cover != 255 {
		r = color.RGBA8MultCover(r, cover)
		g = color.RGBA8MultCover(g, cover)
		b = color.RGBA8MultCover(b, cover)
		a = color.RGBA8MultCover(a, cover)
	}
	if a == 0 && r == 0 && g == 0 && b == 0 {
		return
	}
	ir, ig, ib, ia = ir&3, ig&3, ib&3, ia&3 // BCE: constrain to [0,3]
	p := x * 4
	for range length {
		px := row[p : p+4 : p+4] // BCE: one check per pixel, not four
		px[ir] = color.RGBA8Prelerp(px[ir], r, a)
		px[ig] = color.RGBA8Prelerp(px[ig], g, a)
		px[ib] = color.RGBA8Prelerp(px[ib], b, a)
		px[ia] = color.RGBA8Prelerp(px[ia], a, a)
		p += 4
	}
}

// CopyVline copies a vertical line
func (pf *PixFmtAlphaBlendRGBA[S, B]) CopyVline(x, y, length int, c color.RGBA8[S]) {
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
func (pf *PixFmtAlphaBlendRGBA[S, B]) BlendVline(x, y, length int, c color.RGBA8[S], cover basics.Int8u) {
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

func (pf *PixFmtAlphaBlendRGBA[S, B]) CopyBar(x1, y1, x2, y2 int, c color.RGBA8[S]) {
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

func (pf *PixFmtAlphaBlendRGBA[S, B]) BlendBar(x1, y1, x2, y2 int, c color.RGBA8[S], cover basics.Int8u) {
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

// BlendSolidHspan blends a horizontal span with varying coverage.
//
// This is the single hottest path in AA rendering (~43% cumulative CPU in the
// lion benchmark). The fast path resolves channel order and blend formula once,
// eliminates per-pixel bounds checks, and inlines the color arithmetic —
// mirroring C++ AGG's template monomorphization of blend_pix.
func (pf *PixFmtAlphaBlendRGBA[S, B]) BlendSolidHspan(x, y, length int, c color.RGBA8[S], covers []basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || c.IsTransparent() {
		return
	}

	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}

	row := buffer.RowU8(pf.rbuf, y)

	// Fast path: resolve order and blend mode once, then loop with inlined math.
	if fb, ok := any(pf.blender).(blender.RGBAFastBlender); ok {
		ir, ig, ib, ia := fb.IdxR(), fb.IdxG(), fb.IdxB(), fb.IdxA()

		// Single upfront bounds check for the entire span (BCE).
		spanEnd := (x+length-1)*4 + 3
		if spanEnd >= len(row) {
			goto slow
		}
		_ = row[spanEnd] // prove to the compiler the span is in bounds

		if ir == 0 && ig == 1 && ib == 2 && ia == 3 && covers != nil {
			simd.BlendSolidHspanRGBA(row[x*4:], covers[:length], c.R, c.G, c.B, c.A, fb.PremulSrc())
			return
		}

		if fb.PremulSrc() {
			blendSolidHspanPre(row, x, length, c.R, c.G, c.B, c.A, ir, ig, ib, ia, c.IsOpaque(), covers)
		} else {
			blendSolidHspanPlain(row, x, length, c.R, c.G, c.B, c.A, ir, ig, ib, ia, c.IsOpaque(), covers)
		}
		return
	}

slow:
	if covers == nil {
		for i := 0; i < length; i++ {
			pixelOffset := (x + i) * 4
			if pixelOffset+3 < len(row) {
				pf.blender.BlendPix(row[pixelOffset:pixelOffset+4], c.R, c.G, c.B, c.A, 255)
			}
		}
	} else {
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

// blendSolidHspanPlain implements the plain→premul blend loop (blender_rgba).
// Channel indices and bounds are resolved by the caller.
func blendSolidHspanPlain(
	row []byte, x, length int,
	sr, sg, sb, sa byte,
	ir, ig, ib, ia int,
	opaque bool,
	covers []byte,
) {
	ir, ig, ib, ia = ir&3, ig&3, ib&3, ia&3 // BCE: constrain to [0,3]
	p := x * 4
	if covers == nil {
		// Uniform full coverage — use c.a directly as alpha.
		for range length {
			px := row[p : p+4 : p+4] // BCE: one check per pixel, not four
			px[ir] = color.RGBA8Lerp(px[ir], sr, sa)
			px[ig] = color.RGBA8Lerp(px[ig], sg, sa)
			px[ib] = color.RGBA8Lerp(px[ib], sb, sa)
			px[ia] = color.RGBA8Prelerp(px[ia], sa, sa)
			p += 4
		}
		return
	}
	covers = covers[:length] // BCE: prove covers in bounds
	for i := range length {
		cv := covers[i]
		if cv == 0 {
			p += 4
			continue
		}
		px := row[p : p+4 : p+4] // BCE: one check per pixel, not four
		// C++ optimization: opaque color + full coverage → direct copy.
		if opaque && cv == basics.CoverFull {
			px[ir] = sr
			px[ig] = sg
			px[ib] = sb
			px[ia] = sa
			p += 4
			continue
		}
		alpha := color.RGBA8MultCover(sa, cv)
		if alpha == 0 {
			p += 4
			continue
		}
		px[ir] = color.RGBA8Lerp(px[ir], sr, alpha)
		px[ig] = color.RGBA8Lerp(px[ig], sg, alpha)
		px[ib] = color.RGBA8Lerp(px[ib], sb, alpha)
		px[ia] = color.RGBA8Prelerp(px[ia], alpha, alpha)
		p += 4
	}
}

// blendSolidHspanPre implements the premul→premul blend loop (blender_rgba_pre).
// Channel indices and bounds are resolved by the caller.
func blendSolidHspanPre(
	row []byte, x, length int,
	sr, sg, sb, sa byte,
	ir, ig, ib, ia int,
	opaque bool,
	covers []byte,
) {
	ir, ig, ib, ia = ir&3, ig&3, ib&3, ia&3 // BCE: constrain to [0,3]
	p := x * 4
	if covers == nil {
		// Uniform full coverage — no scaling needed.
		for range length {
			px := row[p : p+4 : p+4] // BCE: one check per pixel, not four
			px[ir] = color.RGBA8Prelerp(px[ir], sr, sa)
			px[ig] = color.RGBA8Prelerp(px[ig], sg, sa)
			px[ib] = color.RGBA8Prelerp(px[ib], sb, sa)
			px[ia] = color.RGBA8Prelerp(px[ia], sa, sa)
			p += 4
		}
		return
	}
	covers = covers[:length] // BCE: prove covers in bounds
	for i := range length {
		cv := covers[i]
		if cv == 0 {
			p += 4
			continue
		}
		px := row[p : p+4 : p+4] // BCE: one check per pixel, not four
		// C++ optimization: opaque color + full coverage → direct copy.
		if opaque && cv == basics.CoverFull {
			px[ir] = sr
			px[ig] = sg
			px[ib] = sb
			px[ia] = sa
			p += 4
			continue
		}
		r, g, b, a := sr, sg, sb, sa
		if cv != 255 {
			r = color.RGBA8MultCover(r, cv)
			g = color.RGBA8MultCover(g, cv)
			b = color.RGBA8MultCover(b, cv)
			a = color.RGBA8MultCover(a, cv)
		}
		if a == 0 && r == 0 && g == 0 && b == 0 {
			p += 4
			continue
		}
		px[ir] = color.RGBA8Prelerp(px[ir], r, a)
		px[ig] = color.RGBA8Prelerp(px[ig], g, a)
		px[ib] = color.RGBA8Prelerp(px[ib], b, a)
		px[ia] = color.RGBA8Prelerp(px[ia], a, a)
		p += 4
	}
}

// BlendSolidVspan blends a vertical span with varying coverage
func (pf *PixFmtAlphaBlendRGBA[S, B]) BlendSolidVspan(x, y, length int, c color.RGBA8[S], covers []basics.Int8u) {
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
func (pf *PixFmtAlphaBlendRGBA[S, B]) CopyColorHspan(x, y, length int, colors []color.RGBA8[S]) {
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
func (pf *PixFmtAlphaBlendRGBA[S, B]) BlendColorHspan(x, y, length int, colors []color.RGBA8[S], covers []basics.Int8u, cover basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || len(colors) == 0 {
		return
	}

	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	if length > len(colors) {
		length = len(colors)
	}

	row := buffer.RowU8(pf.rbuf, y)

	if fb, ok := any(pf.blender).(blender.RGBAFastBlender); ok {
		ir, ig, ib, ia := fb.IdxR(), fb.IdxG(), fb.IdxB(), fb.IdxA()

		spanEnd := (x+length-1)*4 + 3
		if spanEnd >= len(row) {
			goto slow
		}
		_ = row[spanEnd]

		// SIMD fast path: RGBA byte order and per-pixel covers provided.
		if ir == 0 && ig == 1 && ib == 2 && ia == 3 && covers != nil && len(covers) >= length {
			srcBytes := unsafe.Slice((*byte)(unsafe.Pointer(&colors[0])), length*4)
			simd.BlendColorHspanRGBA(row[x*4:], srcBytes, covers[:length], length, fb.PremulSrc())
			return
		}

		// Fast scalar for RGBA order with nil/uniform covers, or any non-RGBA blender.
		ir, ig, ib, ia = ir&3, ig&3, ib&3, ia&3
		p := x * 4
		premul := fb.PremulSrc()
		for i := range length {
			c := colors[i]
			if c.A == 0 {
				p += 4
				continue
			}
			cvr := cover
			if covers != nil && i < len(covers) {
				cvr = covers[i]
			}
			if cvr == 0 {
				p += 4
				continue
			}
			px := row[p : p+4 : p+4]
			if premul {
				sr := color.RGBA8MultCover(c.R, cvr)
				sg := color.RGBA8MultCover(c.G, cvr)
				sb := color.RGBA8MultCover(c.B, cvr)
				sa := color.RGBA8MultCover(c.A, cvr)
				if sa == 0 && sr == 0 && sg == 0 && sb == 0 {
					p += 4
					continue
				}
				px[ir] = color.RGBA8Prelerp(px[ir], sr, sa)
				px[ig] = color.RGBA8Prelerp(px[ig], sg, sa)
				px[ib] = color.RGBA8Prelerp(px[ib], sb, sa)
				px[ia] = color.RGBA8Prelerp(px[ia], sa, sa)
			} else {
				alpha := color.RGBA8MultCover(c.A, cvr)
				if alpha == 0 {
					p += 4
					continue
				}
				px[ir] = color.RGBA8Lerp(px[ir], c.R, alpha)
				px[ig] = color.RGBA8Lerp(px[ig], c.G, alpha)
				px[ib] = color.RGBA8Lerp(px[ib], c.B, alpha)
				px[ia] = color.RGBA8Prelerp(px[ia], alpha, alpha)
			}
			p += 4
		}
		return
	}

slow:
	for i := 0; i < length; i++ {
		c := colors[i]
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
func (pf *PixFmtAlphaBlendRGBA[S, B]) CopyColorVspan(x, y, length int, colors []color.RGBA8[S]) {
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
func (pf *PixFmtAlphaBlendRGBA[S, B]) BlendColorVspan(x, y, length int, colors []color.RGBA8[S], covers []basics.Int8u, cover basics.Int8u) {
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

// Clear clears the entire buffer with the given color.
// Uses bulk copy for contiguous buffers with positive stride.
func (pf *PixFmtAlphaBlendRGBA[S, B]) Clear(c color.RGBA8[S]) {
	w, h := pf.Width(), pf.Height()
	if w <= 0 || h <= 0 {
		return
	}

	stride := pf.rbuf.Stride()
	rowBytes := w * 4

	// Fast path: contiguous buffer with positive stride — fill first row,
	// then replicate via copy() which the compiler maps to memcpy.
	if stride >= rowBytes {
		buf := pf.rbuf.Buf()
		if len(buf) < h*stride {
			goto slow
		}

		// Resolve channel order once and fill the first pixel.
		var pixel [4]byte
		if ro, ok := any(pf.blender).(blender.RawRGBAOrder); ok {
			pixel[ro.IdxR()] = c.R
			pixel[ro.IdxG()] = c.G
			pixel[ro.IdxB()] = c.B
			pixel[ro.IdxA()] = c.A
		} else {
			pixel[0], pixel[1], pixel[2], pixel[3] = c.R, c.G, c.B, c.A
		}

		row0 := buf[:rowBytes]
		simd.FillRGBA(row0, pixel[0], pixel[1], pixel[2], pixel[3], w)

		// Replicate first row to all subsequent rows.
		for y := 1; y < h; y++ {
			copy(buf[y*stride:y*stride+rowBytes], row0)
		}
		return
	}

slow:
	// Fallback for negative stride (bottom-up) or unusual layouts.
	if ro, ok := any(pf.blender).(blender.RawRGBAOrder); ok {
		var pixel [4]byte
		pixel[ro.IdxR()] = c.R
		pixel[ro.IdxG()] = c.G
		pixel[ro.IdxB()] = c.B
		pixel[ro.IdxA()] = c.A
		for y := 0; y < h; y++ {
			row := buffer.RowU8(pf.rbuf, y)
			simd.FillRGBA(row, pixel[0], pixel[1], pixel[2], pixel[3], w)
		}
		return
	}

	for y := 0; y < h; y++ {
		row := buffer.RowU8(pf.rbuf, y)
		for x := 0; x < w; x++ {
			off := x * 4
			if off+3 >= len(row) {
				break
			}
			pf.blender.SetPlain(row[off:off+4], c.R, c.G, c.B, c.A)
		}
	}
}

// Fill is an alias for Clear (fills entire buffer)
func (pf *PixFmtAlphaBlendRGBA[S, B]) Fill(c color.RGBA8[S]) {
	pf.Clear(c)
}

// CopyFrom copies a single scanline from another rendering buffer.
func (pf *PixFmtAlphaBlendRGBA[S, B]) CopyFrom(src interface {
	RowData(y int) []basics.Int8u
	Width() int
	Height() int
}, xdst, ydst, xsrc, ysrc, length int,
) {
	if ydst < 0 || ydst >= pf.Height() || ysrc < 0 || ysrc >= src.Height() || length <= 0 {
		return
	}

	if xsrc < 0 {
		length += xsrc
		xdst -= xsrc
		xsrc = 0
	}
	if xdst < 0 {
		length += xdst
		xsrc -= xdst
		xdst = 0
	}
	if xsrc+length > src.Width() {
		length = src.Width() - xsrc
	}
	if xdst+length > pf.Width() {
		length = pf.Width() - xdst
	}
	if length <= 0 {
		return
	}

	srcRow := src.RowData(ysrc)
	if srcRow == nil {
		return
	}

	bytesPerPixel := detectBytesPerPixel(src, ysrc)
	if bytesPerPixel == 4 {
		dstRow := pf.RowData(ydst)
		if dstRow == nil {
			return
		}
		copy(dstRow[xdst*4:(xdst+length)*4], srcRow[xsrc*4:(xsrc+length)*4])
		return
	}

	for i := 0; i < length; i++ {
		srcColor, ok := decodeRGBA8FromRowData(srcRow, bytesPerPixel, xsrc+i)
		if !ok {
			continue
		}
		pf.CopyPixel(xdst+i, ydst, color.RGBA8[S](srcColor))
	}
}

type rgbaBlendFromSource[S color.Space] interface {
	GetPixel(x, y int) color.RGBA8[S]
	Width() int
	Height() int
}

// BlendFrom blends a single scanline from another RGBA surface.
func (pf *PixFmtAlphaBlendRGBA[S, B]) BlendFrom(src rgbaBlendFromSource[S], xdst, ydst, xsrc, ysrc, length int, cover basics.Int8u) {
	if ydst < 0 || ydst >= pf.Height() || ysrc < 0 || ysrc >= src.Height() || length <= 0 {
		return
	}

	if xsrc < 0 {
		length += xsrc
		xdst -= xsrc
		xsrc = 0
	}
	if xdst < 0 {
		length += xdst
		xsrc -= xdst
		xdst = 0
	}
	if xsrc+length > src.Width() {
		length = src.Width() - xsrc
	}
	if xdst+length > pf.Width() {
		length = pf.Width() - xdst
	}
	if length <= 0 {
		return
	}

	if rowSrc, ok := src.(interface{ RowData(y int) []basics.Int8u }); ok {
		srcRow := rowSrc.RowData(ysrc)
		if srcRow != nil {
			bytesPerPixel := 4
			if bytesPerPixel == 4 {
				if dstRow := pf.RowData(ydst); dstRow != nil && cover == basics.CoverFull {
					start := 0
					end := length
					step := 1
					if xdst > xsrc {
						start = length - 1
						end = -1
						step = -1
					}
					for i := start; i != end; i += step {
						srcOff := (xsrc + i) * 4
						dstOff := (xdst + i) * 4
						copy(dstRow[dstOff:dstOff+4], srcRow[srcOff:srcOff+4])
					}
					return
				}
				for i := 0; i < length; i++ {
					srcOff := (xsrc + i) * 4
					c := color.RGBA8[S]{
						R: srcRow[srcOff+0],
						G: srcRow[srcOff+1],
						B: srcRow[srcOff+2],
						A: srcRow[srcOff+3],
					}
					if cover == basics.CoverFull && c.A == 255 {
						pf.CopyPixel(xdst+i, ydst, c)
						continue
					}
					pf.BlendPixel(xdst+i, ydst, c, cover)
				}
				return
			}
		}
	}

	start := 0
	end := length
	step := 1
	if xdst > xsrc {
		start = length - 1
		end = -1
		step = -1
	}

	for i := start; i != end; i += step {
		c := src.GetPixel(xsrc+i, ysrc)
		if cover == basics.CoverFull && c.A == 255 {
			pf.CopyPixel(xdst+i, ydst, c)
			continue
		}
		pf.BlendPixel(xdst+i, ydst, c, cover)
	}
}

// BlendFromColor blends a single color through a grayscale source row used as coverage.
func (pf *PixFmtAlphaBlendRGBA[S, B]) BlendFromColor(src interface {
	RowData(y int) []basics.Int8u
	Width() int
	Height() int
}, c color.RGBA8[S], xdst, ydst, xsrc, ysrc, length int, cover basics.Int8u,
) {
	if ydst < 0 || ydst >= pf.Height() || ysrc < 0 || ysrc >= src.Height() || length <= 0 || c.IsTransparent() {
		return
	}

	if xsrc < 0 {
		length += xsrc
		xdst -= xsrc
		xsrc = 0
	}
	if xdst < 0 {
		length += xdst
		xsrc -= xdst
		xdst = 0
	}
	if xsrc+length > src.Width() {
		length = src.Width() - xsrc
	}
	if xdst+length > pf.Width() {
		length = pf.Width() - xdst
	}
	if length <= 0 {
		return
	}

	srcRow := src.RowData(ysrc)
	if srcRow == nil {
		return
	}

	bytesPerPixel := detectBytesPerPixel(src, ysrc)
	for i := 0; i < length; i++ {
		srcOffset := (xsrc + i) * bytesPerPixel
		if srcOffset < 0 || srcOffset >= len(srcRow) {
			continue
		}
		scaledCover := color.RGBA8MultCover(srcRow[srcOffset], cover)
		if scaledCover == 0 {
			continue
		}
		if scaledCover == basics.CoverFull && c.A == 255 {
			pf.CopyPixel(xdst+i, ydst, c)
			continue
		}
		pf.BlendPixel(xdst+i, ydst, c, scaledCover)
	}
}

// BlendFromLUT blends colors from a lookup table indexed by a grayscale source row.
func (pf *PixFmtAlphaBlendRGBA[S, B]) BlendFromLUT(src interface {
	RowData(y int) []basics.Int8u
	Width() int
	Height() int
}, colorLUT []color.RGBA8[S], xdst, ydst, xsrc, ysrc, length int, cover basics.Int8u,
) {
	if ydst < 0 || ydst >= pf.Height() || ysrc < 0 || ysrc >= src.Height() || length <= 0 || len(colorLUT) == 0 {
		return
	}

	if xsrc < 0 {
		length += xsrc
		xdst -= xsrc
		xsrc = 0
	}
	if xdst < 0 {
		length += xdst
		xsrc -= xdst
		xdst = 0
	}
	if xsrc+length > src.Width() {
		length = src.Width() - xsrc
	}
	if xdst+length > pf.Width() {
		length = pf.Width() - xdst
	}
	if length <= 0 {
		return
	}

	srcRow := src.RowData(ysrc)
	if srcRow == nil {
		return
	}

	bytesPerPixel := detectBytesPerPixel(src, ysrc)
	for i := 0; i < length; i++ {
		srcOffset := (xsrc + i) * bytesPerPixel
		if srcOffset < 0 || srcOffset >= len(srcRow) {
			continue
		}
		lutIndex := int(srcRow[srcOffset])
		if lutIndex >= len(colorLUT) {
			continue
		}
		lutColor := colorLUT[lutIndex]
		if lutColor.IsTransparent() {
			continue
		}
		if cover == basics.CoverFull && lutColor.A == 255 {
			pf.CopyPixel(xdst+i, ydst, lutColor)
			continue
		}
		pf.BlendPixel(xdst+i, ydst, lutColor, cover)
	}
}

// Premultiply converts the entire buffer from plain to premultiplied alpha.
// For standard RGBA byte-order blenders the hot path delegates to SIMD.
func (pf *PixFmtAlphaBlendRGBA[S, B]) Premultiply() {
	w := pf.Width()
	for y := 0; y < pf.Height(); y++ {
		row := buffer.RowU8(pf.rbuf, y)
		if ro, ok := any(pf.blender).(blender.RawRGBAOrder); ok {
			ir, ig, ib, ia := ro.IdxR(), ro.IdxG(), ro.IdxB(), ro.IdxA()
			// SIMD fast path: standard R,G,B,A byte order.
			if ir == 0 && ig == 1 && ib == 2 && ia == 3 {
				simd.PremultiplyRGBA(row, w)
				continue
			}
			// Non-standard order (e.g. BGRA): scalar loop with direct indexing.
			ir, ig, ib, ia = ir&3, ig&3, ib&3, ia&3
			for x := 0; x < w; x++ {
				off := x * 4
				if off+3 < len(row) {
					r, g, b, a := row[off+ir], row[off+ig], row[off+ib], row[off+ia]
					row[off+ir] = color.RGBA8Multiply(r, a)
					row[off+ig] = color.RGBA8Multiply(g, a)
					row[off+ib] = color.RGBA8Multiply(b, a)
				}
			}
		} else {
			for x := 0; x < w; x++ {
				off := x * 4
				if off+3 < len(row) {
					r, g, b, a := pf.blender.GetPlain(row[off : off+4])
					pf.blender.SetPlain(row[off:off+4], color.RGBA8Multiply(r, a), color.RGBA8Multiply(g, a), color.RGBA8Multiply(b, a), a)
				}
			}
		}
	}
}

// Demultiply converts the entire buffer from premultiplied to plain alpha.
// For standard RGBA byte-order blenders the hot path delegates to SIMD.
func (pf *PixFmtAlphaBlendRGBA[S, B]) Demultiply() {
	w := pf.Width()
	for y := 0; y < pf.Height(); y++ {
		row := buffer.RowU8(pf.rbuf, y)
		if ro, ok := any(pf.blender).(blender.RawRGBAOrder); ok {
			ir, ig, ib, ia := ro.IdxR(), ro.IdxG(), ro.IdxB(), ro.IdxA()
			// SIMD fast path: standard R,G,B,A byte order.
			if ir == 0 && ig == 1 && ib == 2 && ia == 3 {
				simd.DemultiplyRGBA(row, w)
				continue
			}
			// Non-standard order: scalar loop.
			ir, ig, ib, ia = ir&3, ig&3, ib&3, ia&3
			for x := 0; x < w; x++ {
				off := x * 4
				if off+3 < len(row) {
					a := row[off+ia]
					if a > 0 {
						row[off+ir] = demul8(row[off+ir], a)
						row[off+ig] = demul8(row[off+ig], a)
						row[off+ib] = demul8(row[off+ib], a)
					} else {
						row[off+ir] = 0
						row[off+ig] = 0
						row[off+ib] = 0
					}
				}
			}
		} else {
			for x := 0; x < w; x++ {
				off := x * 4
				if off+3 < len(row) {
					r, g, b, a := pf.blender.GetPlain(row[off : off+4])
					pf.blender.SetPlain(row[off:off+4], demul8(r, a), demul8(g, a), demul8(b, a), a)
				}
			}
		}
	}
}

// Concrete RGBA pixel format types for different color orders
type (
	// Premultiplied framebuffer, plain src blending (AGG blender_rgba)
	PixFmtRGBA32[S color.Space] = PixFmtAlphaBlendRGBA[S, blender.BlenderRGBA8[S, order.RGBA]]
	PixFmtBGRA32[S color.Space] = PixFmtAlphaBlendRGBA[S, blender.BlenderRGBA8[S, order.BGRA]]
	PixFmtARGB32[S color.Space] = PixFmtAlphaBlendRGBA[S, blender.BlenderRGBA8[S, order.ARGB]]
	PixFmtABGR32[S color.Space] = PixFmtAlphaBlendRGBA[S, blender.BlenderRGBA8[S, order.ABGR]]

	// Premultiplied framebuffer, premul-style blending (AGG blender_rgba_pre)
	PixFmtRGBA32Pre[S color.Space] = PixFmtAlphaBlendRGBA[S, blender.BlenderRGBA8Pre[S, order.RGBA]]
	PixFmtBGRA32Pre[S color.Space] = PixFmtAlphaBlendRGBA[S, blender.BlenderRGBA8Pre[S, order.BGRA]]
	PixFmtARGB32Pre[S color.Space] = PixFmtAlphaBlendRGBA[S, blender.BlenderRGBA8Pre[S, order.ARGB]]
	PixFmtABGR32Pre[S color.Space] = PixFmtAlphaBlendRGBA[S, blender.BlenderRGBA8Pre[S, order.ABGR]]

	// Plain framebuffer (AGG blender_rgba_plain)
	PixFmtRGBA32Plain[S color.Space] = PixFmtAlphaBlendRGBA[S, blender.BlenderRGBA8Plain[S, order.RGBA]]
	PixFmtBGRA32Plain[S color.Space] = PixFmtAlphaBlendRGBA[S, blender.BlenderRGBA8Plain[S, order.BGRA]]
	PixFmtARGB32Plain[S color.Space] = PixFmtAlphaBlendRGBA[S, blender.BlenderRGBA8Plain[S, order.ARGB]]
	PixFmtABGR32Plain[S color.Space] = PixFmtAlphaBlendRGBA[S, blender.BlenderRGBA8Plain[S, order.ABGR]]
)

//////////////////////////////////////////////////////////////////////////////////////
// Constructors
//////////////////////////////////////////////////////////////////////////////////////

// Constructors for RGBA pixel formats
func NewPixFmtRGBA32[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtRGBA32[S] {
	return NewPixFmtAlphaBlendRGBA[S](rbuf, blender.BlenderRGBA8[S, order.RGBA]{})
}

func NewPixFmtBGRA32[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtBGRA32[S] {
	return NewPixFmtAlphaBlendRGBA[S](rbuf, blender.BlenderRGBA8[S, order.BGRA]{})
}

func NewPixFmtARGB32[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtARGB32[S] {
	return NewPixFmtAlphaBlendRGBA[S](rbuf, blender.BlenderRGBA8[S, order.ARGB]{})
}

func NewPixFmtABGR32[S color.Space](rbuf *buffer.RenderingBufferU8) *PixFmtABGR32[S] {
	return NewPixFmtAlphaBlendRGBA[S](rbuf, blender.BlenderRGBA8[S, order.ABGR]{})
}

// Constructors for RGBA pixel formats (premultiplied)
func NewPixFmtRGBA32Pre[S color.Space](r *buffer.RenderingBufferU8) *PixFmtRGBA32Pre[S] {
	return NewPixFmtAlphaBlendRGBA[S](r, blender.BlenderRGBA8Pre[S, order.RGBA]{})
}

func NewPixFmtBGRA32Pre[S color.Space](r *buffer.RenderingBufferU8) *PixFmtBGRA32Pre[S] {
	return NewPixFmtAlphaBlendRGBA[S](r, blender.BlenderRGBA8Pre[S, order.BGRA]{})
}

func NewPixFmtARGB32Pre[S color.Space](r *buffer.RenderingBufferU8) *PixFmtARGB32Pre[S] {
	return NewPixFmtAlphaBlendRGBA[S](r, blender.BlenderRGBA8Pre[S, order.ARGB]{})
}

func NewPixFmtABGR32Pre[S color.Space](r *buffer.RenderingBufferU8) *PixFmtABGR32Pre[S] {
	return NewPixFmtAlphaBlendRGBA[S](r, blender.BlenderRGBA8Pre[S, order.ABGR]{})
}

// Constructors for RGBA pixel formats (plain)
func NewPixFmtRGBA32Plain[S color.Space](r *buffer.RenderingBufferU8) *PixFmtRGBA32Plain[S] {
	return NewPixFmtAlphaBlendRGBA[S](r, blender.BlenderRGBA8Plain[S, order.RGBA]{})
}

func NewPixFmtBGRA32Plain[S color.Space](r *buffer.RenderingBufferU8) *PixFmtBGRA32Plain[S] {
	return NewPixFmtAlphaBlendRGBA[S](r, blender.BlenderRGBA8Plain[S, order.BGRA]{})
}

func NewPixFmtARGB32Plain[S color.Space](r *buffer.RenderingBufferU8) *PixFmtARGB32Plain[S] {
	return NewPixFmtAlphaBlendRGBA[S](r, blender.BlenderRGBA8Plain[S, order.ARGB]{})
}

func NewPixFmtABGR32Plain[S color.Space](r *buffer.RenderingBufferU8) *PixFmtABGR32Plain[S] {
	return NewPixFmtAlphaBlendRGBA[S](r, blender.BlenderRGBA8Plain[S, order.ABGR]{})
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

// ==============================================================================

// ApplyGammaDir applies gamma correction to the entire buffer using the forward direction.
func (pf *PixFmtAlphaBlendRGBA[S, B]) ApplyGammaDir(gamma func(basics.Int8u) basics.Int8u) {
	for y := 0; y < pf.Height(); y++ {
		row := buffer.RowU8(pf.rbuf, y)
		for x := 0; x < pf.Width(); x++ {
			off := x * 4
			if off+3 >= len(row) {
				break
			}

			// Use fast path if available
			if ro, ok := any(pf.blender).(blender.RawRGBAOrder); ok {
				ir, ig, ib := ro.IdxR(), ro.IdxG(), ro.IdxB()
				// Apply gamma only to RGB channels, not alpha
				row[off+ir] = gamma(row[off+ir])
				row[off+ig] = gamma(row[off+ig])
				row[off+ib] = gamma(row[off+ib])
			} else {
				// Use blender interface
				r, g, b, a := pf.blender.GetPlain(row[off : off+4])
				r, g, b = gamma(r), gamma(g), gamma(b)
				pf.blender.SetPlain(row[off:off+4], r, g, b, a)
			}
		}
	}
}

// ApplyGammaInv applies inverse gamma correction to the entire buffer.
func (pf *PixFmtAlphaBlendRGBA[S, B]) ApplyGammaInv(gamma func(basics.Int8u) basics.Int8u) {
	pf.ApplyGammaDir(gamma) // Same implementation for RGBA formats
}

////////////////////////////////////////////////////////////////////////////////
// Helpers
////////////////////////////////////////////////////////////////////////////////

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
