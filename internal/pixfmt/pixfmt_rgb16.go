package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/pixfmt/blender"
)

// ==============================================================================
// RGB48 (16-bit per channel) Pixel Formats
// ==============================================================================

// PixFmtAlphaBlendRGB48 represents RGB pixel format with 16-bit components (6 bytes per pixel)
type PixFmtAlphaBlendRGB48[
	S color.Space,
	B blender.RGB48Blender[S],
] struct {
	rbuf    *buffer.RenderingBufferU16
	blender B
}

// NewPixFmtAlphaBlendRGB48 creates a new RGB48 pixel format
func NewPixFmtAlphaBlendRGB48[
	S color.Space,
	B blender.RGB48Blender[S],
](rbuf *buffer.RenderingBufferU16, blender B) *PixFmtAlphaBlendRGB48[S, B] {
	return &PixFmtAlphaBlendRGB48[S, B]{
		rbuf:    rbuf,
		blender: blender,
	}
}

// Width returns the buffer width
func (pf *PixFmtAlphaBlendRGB48[S, B]) Width() int {
	return pf.rbuf.Width()
}

// Height returns the buffer height
func (pf *PixFmtAlphaBlendRGB48[S, B]) Height() int {
	return pf.rbuf.Height()
}

// PixWidth returns bytes per pixel (6 for RGB48)
func (pf *PixFmtAlphaBlendRGB48[S, B]) PixWidth() int {
	return 6
}

// GetPixel returns the pixel at the given coordinates
func (pf *PixFmtAlphaBlendRGB48[S, B]) GetPixel(x, y int) color.RGB16[S] {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return color.RGB16[S]{}
	}

	row := buffer.RowU16(pf.rbuf, y)
	pixelOffset := x * 3 // 3 components per pixel
	if pixelOffset+2 >= len(row) {
		return color.RGB16[S]{}
	}

	r, g, b := pf.blender.GetPlain(row[pixelOffset : pixelOffset+3])
	return color.RGB16[S]{R: r, G: g, B: b}
}

// CopyPixel copies a pixel without blending
func (pf *PixFmtAlphaBlendRGB48[S, B]) CopyPixel(x, y int, c color.RGB16[S]) {
	if !InBounds(x, y, pf.Width(), pf.Height()) {
		return
	}

	row := buffer.RowU16(pf.rbuf, y)
	pixelOffset := x * 3
	if pixelOffset+2 >= len(row) {
		return
	}

	pf.blender.SetPlain(row[pixelOffset:pixelOffset+3], c.R, c.G, c.B)
}

// BlendPixel blends a pixel with the given alpha and coverage
func (pf *PixFmtAlphaBlendRGB48[S, B]) BlendPixel(x, y int, c color.RGB16[S], alpha, cover basics.Int16u) {
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
func (pf *PixFmtAlphaBlendRGB48[S, B]) Clear(c color.RGB16[S]) {
	for y := 0; y < pf.Height(); y++ {
		row := buffer.RowU16(pf.rbuf, y)
		for x := 0; x < pf.Width(); x++ {
			pixelOffset := x * 3
			if pixelOffset+2 < len(row) {
				pf.blender.SetPlain(row[pixelOffset:pixelOffset+3], c.R, c.G, c.B)
			}
		}
	}
}

// Concrete RGB48 pixel format types
type (
	PixFmtRGB48Linear = PixFmtAlphaBlendRGB48[color.Linear, blender.BlenderRGB48LinearRGB]
	PixFmtBGR48Linear = PixFmtAlphaBlendRGB48[color.Linear, blender.BlenderRGB48LinearBGR]
	PixFmtRGB48SRGB   = PixFmtAlphaBlendRGB48[color.SRGB, blender.BlenderRGB48SRGBRGB]
	PixFmtBGR48SRGB   = PixFmtAlphaBlendRGB48[color.SRGB, blender.BlenderRGB48SRGBBGR]
)

// Constructor functions for RGB48 pixel formats
func NewPixFmtRGB48Linear(rbuf *buffer.RenderingBufferU16) *PixFmtRGB48Linear {
	return NewPixFmtAlphaBlendRGB48[color.Linear](rbuf, blender.BlenderRGB48LinearRGB{})
}

func NewPixFmtBGR48Linear(rbuf *buffer.RenderingBufferU16) *PixFmtBGR48Linear {
	return NewPixFmtAlphaBlendRGB48[color.Linear](rbuf, blender.BlenderRGB48LinearBGR{})
}

func NewPixFmtRGB48SRGB(rbuf *buffer.RenderingBufferU16) *PixFmtRGB48SRGB {
	return NewPixFmtAlphaBlendRGB48[color.SRGB](rbuf, blender.BlenderRGB48SRGBRGB{})
}

func NewPixFmtBGR48SRGB(rbuf *buffer.RenderingBufferU16) *PixFmtBGR48SRGB {
	return NewPixFmtAlphaBlendRGB48[color.SRGB](rbuf, blender.BlenderRGB48SRGBBGR{})
}

// ==============================================================================
// Premultiplied RGB Pixel Format Types
// ==============================================================================

// RGB48 premultiplied variants
type (
	PixFmtRGB48Pre     = PixFmtAlphaBlendRGB48[color.Linear, blender.BlenderRGB48PreLinearRGB]
	PixFmtBGR48Pre     = PixFmtAlphaBlendRGB48[color.Linear, blender.BlenderRGB48PreLinearBGR]
	PixFmtRGB48PreSRGB = PixFmtAlphaBlendRGB48[color.SRGB, blender.BlenderRGB48PreSRGBRGB]
	PixFmtBGR48PreSRGB = PixFmtAlphaBlendRGB48[color.SRGB, blender.BlenderRGB48PreSRGBBGR]
)

// Constructor functions for premultiplied RGB48 formats
func NewPixFmtRGB48Pre(rbuf *buffer.RenderingBufferU16) *PixFmtRGB48Pre {
	return NewPixFmtAlphaBlendRGB48[color.Linear](rbuf, blender.BlenderRGB48PreLinearRGB{})
}

func NewPixFmtBGR48Pre(rbuf *buffer.RenderingBufferU16) *PixFmtBGR48Pre {
	return NewPixFmtAlphaBlendRGB48[color.Linear](rbuf, blender.BlenderRGB48PreLinearBGR{})
}

func NewPixFmtRGB48PreSRGB(rbuf *buffer.RenderingBufferU16) *PixFmtRGB48PreSRGB {
	return NewPixFmtAlphaBlendRGB48[color.SRGB](rbuf, blender.BlenderRGB48PreSRGBRGB{})
}

func NewPixFmtBGR48PreSRGB(rbuf *buffer.RenderingBufferU16) *PixFmtBGR48PreSRGB {
	return NewPixFmtAlphaBlendRGB48[color.SRGB](rbuf, blender.BlenderRGB48PreSRGBBGR{})
}

// Pixel returns the pixel at the given coordinates (alias for GetPixel)
func (pf *PixFmtAlphaBlendRGB48[S, B]) Pixel(x, y int) color.RGB16[S] {
	return pf.GetPixel(x, y)
}

func (pf *PixFmtAlphaBlendRGB48[S, B]) BlendPixelRGBA(x, y int, c color.RGBA16[S], cover basics.Int16u) {
	pf.BlendPixel(x, y, color.RGB16[S]{R: c.R, G: c.G, B: c.B}, c.A, cover)
}

func (pf *PixFmtAlphaBlendRGB48[S, B]) CopyHline(x1, y, x2 int, c color.RGB16[S]) {
	if y < 0 || y >= pf.Height() {
		return
	}
	x1 = ClampX(x1, pf.Width())
	x2 = ClampX(x2, pf.Width())
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	row := buffer.RowU16(pf.rbuf, y)
	// Check if we can use fast path
	if ro, ok := any(pf.blender).(blender.RawRGB48Order); ok {
		ir, ig, ib := ro.IdxR(), ro.IdxG(), ro.IdxB()
		for x := x1; x <= x2; x++ {
			off := x * 3
			if off+2 >= len(row) {
				break
			}
			row[off+ir], row[off+ig], row[off+ib] = c.R, c.G, c.B
		}
	} else {
		for x := x1; x <= x2; x++ {
			off := x * 3
			if off+2 >= len(row) {
				break
			}
			pf.blender.SetPlain(row[off:off+3], c.R, c.G, c.B)
		}
	}
}

func (pf *PixFmtAlphaBlendRGB48[S, B]) BlendHline(x1, y, x2 int, c color.RGB16[S], alpha, cover basics.Int16u) {
	if y < 0 || y >= pf.Height() {
		return
	}
	x1 = ClampX(x1, pf.Width())
	x2 = ClampX(x2, pf.Width())
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	row := buffer.RowU16(pf.rbuf, y)
	for x := x1; x <= x2; x++ {
		off := x * 3
		if off+2 >= len(row) {
			break
		}
		pf.blender.BlendPix(row[off:off+3], c.R, c.G, c.B, alpha, cover)
	}
}

func (pf *PixFmtAlphaBlendRGB48[S, B]) CopyVline(x, y1, y2 int, c color.RGB16[S]) {
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

func (pf *PixFmtAlphaBlendRGB48[S, B]) BlendVline(x, y1, y2 int, c color.RGB16[S], alpha, cover basics.Int16u) {
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

func (pf *PixFmtAlphaBlendRGB48[S, B]) CopyBar(x1, y1, x2, y2 int, c color.RGB16[S]) {
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

func (pf *PixFmtAlphaBlendRGB48[S, B]) BlendBar(x1, y1, x2, y2 int, c color.RGB16[S], alpha, cover basics.Int16u) {
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

func (pf *PixFmtAlphaBlendRGB48[S, B]) BlendSolidHspan(x, y, length int, c color.RGB16[S], alpha basics.Int16u, covers []basics.Int16u) {
	if y < 0 || y >= pf.Height() || length <= 0 {
		return
	}
	x = ClampX(x, pf.Width())
	if x+length > pf.Width() {
		length = pf.Width() - x
	}
	row := buffer.RowU16(pf.rbuf, y)
	if covers == nil {
		for i := 0; i < length; i++ {
			off := (x + i) * 3
			if off+2 >= len(row) {
				break
			}
			pf.blender.BlendPix(row[off:off+3], c.R, c.G, c.B, alpha, 65535)
		}
		return
	}
	for i := 0; i < length && i < len(covers); i++ {
		if covers[i] == 0 {
			continue
		}
		off := (x + i) * 3
		if off+2 >= len(row) {
			break
		}
		pf.blender.BlendPix(row[off:off+3], c.R, c.G, c.B, alpha, covers[i])
	}
}

func (pf *PixFmtAlphaBlendRGB48[S, B]) BlendSolidVspan(x, y, length int, c color.RGB16[S], alpha basics.Int16u, covers []basics.Int16u) {
	if x < 0 || x >= pf.Width() || length <= 0 {
		return
	}
	y = ClampY(y, pf.Height())
	if y+length > pf.Height() {
		length = pf.Height() - y
	}
	if covers == nil {
		for i := 0; i < length; i++ {
			pf.BlendPixel(x, y+i, c, alpha, 65535)
		}
		return
	}
	for i := 0; i < length && i < len(covers); i++ {
		if covers[i] != 0 {
			pf.BlendPixel(x, y+i, c, alpha, covers[i])
		}
	}
}

// CopyColorHspan copies a horizontal span of colors
func (pf *PixFmtAlphaBlendRGB48[S, B]) CopyColorHspan(x, y, length int, colors []color.RGB16[S]) {
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
func (pf *PixFmtAlphaBlendRGB48[S, B]) BlendColorHspan(x, y, length int, colors []color.RGB16[S], covers []basics.Int16u, alpha, cover basics.Int16u) {
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

		cvr := cover
		if covers != nil && i < len(covers) {
			cvr = covers[i]
		}
		if cvr > 0 {
			pf.BlendPixel(x+i, y, c, alpha, cvr)
		}
	}
}

// CopyColorVspan copies a vertical span of colors
func (pf *PixFmtAlphaBlendRGB48[S, B]) CopyColorVspan(x, y, length int, colors []color.RGB16[S]) {
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
func (pf *PixFmtAlphaBlendRGB48[S, B]) BlendColorVspan(x, y, length int, colors []color.RGB16[S], covers []basics.Int16u, alpha, cover basics.Int16u) {
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

		cvr := cover
		if covers != nil && i < len(covers) {
			cvr = covers[i]
		}
		if cvr > 0 {
			pf.BlendPixel(x, y+i, c, alpha, cvr)
		}
	}
}

// Fill is an alias for Clear (fills entire buffer)
func (pf *PixFmtAlphaBlendRGB48[S, B]) Fill(c color.RGB16[S]) {
	pf.Clear(c)
}

// CopyFrom copies from another RGB pixel format
func (pf *PixFmtAlphaBlendRGB48[S, B]) CopyFrom(src *PixFmtAlphaBlendRGB48[S, B], srcX, srcY, dstX, dstY, width, height int) {
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

	rows := make([][]color.RGB16[S], height)
	for y := 0; y < height; y++ {
		row := make([]color.RGB16[S], width)
		for x := 0; x < width; x++ {
			row[x] = src.GetPixel(srcX+x, srcY+y)
		}
		rows[y] = row
	}

	for y, row := range rows {
		for x, pixel := range row {
			pf.CopyPixel(dstX+x, dstY+y, pixel)
		}
	}
}

// ==============================================================================
// AGG Compatibility: Whole-Buffer Utilities
// ==============================================================================

// Premultiply converts the entire buffer to premultiplied alpha format.
// For RGB formats without stored alpha, this is a no-op since there's no alpha channel.
func (pf *PixFmtAlphaBlendRGB48[S, B]) Premultiply() {
	// RGB formats don't store alpha, so premultiplying is not applicable
	// This is a no-op for RGB formats
}

// Demultiply converts the entire buffer from premultiplied alpha format.
// For RGB formats without stored alpha, this is a no-op since there's no alpha channel.
func (pf *PixFmtAlphaBlendRGB48[S, B]) Demultiply() {
	// RGB formats don't store alpha, so demultiplying is not applicable
	// This is a no-op for RGB formats
}

// ApplyGammaDir applies gamma correction to the entire buffer using the forward direction.
func (pf *PixFmtAlphaBlendRGB48[S, B]) ApplyGammaDir(gamma func(basics.Int16u) basics.Int16u) {
	for y := 0; y < pf.Height(); y++ {
		row := buffer.RowU16(pf.rbuf, y)
		for x := 0; x < pf.Width(); x++ {
			off := x * 3
			if off+2 >= len(row) {
				break
			}

			// Use fast path if available
			if ro, ok := any(pf.blender).(blender.RawRGB48Order); ok {
				ir, ig, ib := ro.IdxR(), ro.IdxG(), ro.IdxB()
				row[off+ir] = gamma(row[off+ir])
				row[off+ig] = gamma(row[off+ig])
				row[off+ib] = gamma(row[off+ib])
			} else {
				// Use blender interface
				r, g, b := pf.blender.GetPlain(row[off : off+3])
				r, g, b = gamma(r), gamma(g), gamma(b)
				pf.blender.SetPlain(row[off:off+3], r, g, b)
			}
		}
	}
}

// ApplyGammaInv applies inverse gamma correction to the entire buffer.
func (pf *PixFmtAlphaBlendRGB48[S, B]) ApplyGammaInv(gamma func(basics.Int16u) basics.Int16u) {
	pf.ApplyGammaDir(gamma) // Same implementation for RGB formats
}
