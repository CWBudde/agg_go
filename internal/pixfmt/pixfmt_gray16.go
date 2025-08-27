package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/pixfmt/blender"
)

// PixFmtAlphaBlendGray16 implements alpha blending for 16-bit grayscale pixel formats
type PixFmtAlphaBlendGray16[B any, CS color.Space] struct {
	rbuf     *buffer.RenderingBufferU16
	blender  B
	category PixFmtGrayTag
}

// Gray16PixelType represents a 16-bit grayscale pixel
type Gray16PixelType struct {
	V basics.Int16u // Grayscale value
}

// Set sets the grayscale value
func (p *Gray16PixelType) Set(v basics.Int16u) {
	p.V = v
}

// NewPixFmtAlphaBlendGray16 creates a new 16-bit grayscale pixel format
func NewPixFmtAlphaBlendGray16[B any, CS color.Space](rbuf *buffer.RenderingBufferU16, blender B) *PixFmtAlphaBlendGray16[B, CS] {
	return &PixFmtAlphaBlendGray16[B, CS]{
		rbuf:    rbuf,
		blender: blender,
	}
}

// Basic properties
func (pf *PixFmtAlphaBlendGray16[B, CS]) Width() int {
	return pf.rbuf.Width()
}

func (pf *PixFmtAlphaBlendGray16[B, CS]) Height() int {
	return pf.rbuf.Height()
}

func (pf *PixFmtAlphaBlendGray16[B, CS]) PixWidth() int {
	return 2 // 2 bytes per pixel for 16-bit grayscale
}

func (pf *PixFmtAlphaBlendGray16[B, CS]) Stride() int {
	return pf.rbuf.Stride()
}

// RowPtr returns a pointer to the pixel data for the given row
func (pf *PixFmtAlphaBlendGray16[B, CS]) RowPtr(y int) []basics.Int16u {
	return buffer.RowU16(pf.rbuf, y)
}

// PixPtr returns a pointer to the specific pixel
func (pf *PixFmtAlphaBlendGray16[B, CS]) PixPtr(x, y int) *basics.Int16u {
	row := buffer.RowU16(pf.rbuf, y)
	if x >= 0 && x < len(row) {
		return &row[x]
	}
	return nil
}

// MakePix creates a pixel value from grayscale components
func (pf *PixFmtAlphaBlendGray16[B, CS]) MakePix(v basics.Int16u) Gray16PixelType {
	return Gray16PixelType{V: v}
}

// Core pixel operations

// CopyPixel copies a pixel without blending
func (pf *PixFmtAlphaBlendGray16[B, CS]) CopyPixel(x, y int, c color.Gray16[CS]) {
	if InBounds(x, y, pf.Width(), pf.Height()) {
		pixel := pf.PixPtr(x, y)
		if pixel != nil {
			*pixel = c.V
		}
	}
}

// BlendPixel blends a pixel with coverage
func (pf *PixFmtAlphaBlendGray16[B, CS]) BlendPixel(x, y int, c color.Gray16[CS], cover basics.Int16u) {
	if InBounds(x, y, pf.Width(), pf.Height()) && c.A > 0 {
		pixel := pf.PixPtr(x, y)
		if pixel != nil {
			if blender, ok := any(pf.blender).(blender.BlenderGray16Linear); ok {
				blender.BlendPix(pixel, c.V, c.A, cover)
			}
		}
	}
}

// GetPixel gets the pixel color at the specified coordinates
func (pf *PixFmtAlphaBlendGray16[B, CS]) GetPixel(x, y int) color.Gray16[CS] {
	if InBounds(x, y, pf.Width(), pf.Height()) {
		pixel := pf.PixPtr(x, y)
		if pixel != nil {
			return color.NewGray16WithAlpha[CS](*pixel, 65535) // Full alpha
		}
	}
	return color.Gray16[CS]{}
}

// Pixel returns the pixel at the given coordinates (alias for GetPixel to satisfy interface)
func (pf *PixFmtAlphaBlendGray16[B, CS]) Pixel(x, y int) color.Gray16[CS] {
	return pf.GetPixel(x, y)
}

// Line operations

// CopyHline copies a horizontal line
func (pf *PixFmtAlphaBlendGray16[B, CS]) CopyHline(x, y, length int, c color.Gray16[CS]) {
	if y < 0 || y >= pf.Height() || length <= 0 {
		return
	}

	x = Max(0, x)
	if x+length > pf.Width() {
		length = pf.Width() - x
	}

	row := pf.RowPtr(y)
	blender.CopyGray16Hline(row, x, length, c)
}

// BlendHline blends a horizontal line with coverage
func (pf *PixFmtAlphaBlendGray16[B, CS]) BlendHline(x, y, length int, c color.Gray16[CS], cover basics.Int16u) {
	if y < 0 || y >= pf.Height() || length <= 0 || c.A == 0 {
		return
	}

	x = Max(0, x)
	if x+length > pf.Width() {
		length = pf.Width() - x
	}

	row := pf.RowPtr(y)
	if blender, ok := any(pf.blender).(blender.BlenderGray16Linear); ok {
		for i := 0; i < length; i++ {
			if c.A > 0 {
				blender.BlendPix(&row[x+i], c.V, c.A, cover)
			}
		}
	}
}

// CopyVline copies a vertical line
func (pf *PixFmtAlphaBlendGray16[B, CS]) CopyVline(x, y, length int, c color.Gray16[CS]) {
	if x < 0 || x >= pf.Width() || length <= 0 {
		return
	}

	y = Max(0, y)
	if y+length > pf.Height() {
		length = pf.Height() - y
	}

	for i := 0; i < length; i++ {
		pf.CopyPixel(x, y+i, c)
	}
}

// BlendVline blends a vertical line with coverage
func (pf *PixFmtAlphaBlendGray16[B, CS]) BlendVline(x, y, length int, c color.Gray16[CS], cover basics.Int16u) {
	if x < 0 || x >= pf.Width() || length <= 0 || c.A == 0 {
		return
	}

	y = Max(0, y)
	if y+length > pf.Height() {
		length = pf.Height() - y
	}

	for i := 0; i < length; i++ {
		pf.BlendPixel(x, y+i, c, cover)
	}
}

// Rectangle operations

// CopyBar copies a filled rectangle
func (pf *PixFmtAlphaBlendGray16[B, CS]) CopyBar(x1, y1, x2, y2 int, c color.Gray16[CS]) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	x1 = Max(0, x1)
	y1 = Max(0, y1)
	x2 = Min(pf.Width()-1, x2)
	y2 = Min(pf.Height()-1, y2)

	for y := y1; y <= y2; y++ {
		pf.CopyHline(x1, y, x2-x1+1, c)
	}
}

// BlendBar blends a filled rectangle with coverage
func (pf *PixFmtAlphaBlendGray16[B, CS]) BlendBar(x1, y1, x2, y2 int, c color.Gray16[CS], cover basics.Int16u) {
	if c.A == 0 {
		return
	}

	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	x1 = Max(0, x1)
	y1 = Max(0, y1)
	x2 = Min(pf.Width()-1, x2)
	y2 = Min(pf.Height()-1, y2)

	for y := y1; y <= y2; y++ {
		pf.BlendHline(x1, y, x2-x1+1, c, cover)
	}
}

// Span operations

// BlendSolidHspan blends a horizontal span with solid color and variable coverage
func (pf *PixFmtAlphaBlendGray16[B, CS]) BlendSolidHspan(x, y, length int, c color.Gray16[CS], covers []basics.Int16u) {
	if y < 0 || y >= pf.Height() || c.A == 0 || length <= 0 {
		return
	}

	x1 := Max(0, x)
	x2 := Min(pf.Width()-1, x+length-1)

	if x1 <= x2 {
		row := pf.RowPtr(y)
		coverOffset := Max(0, -x)
		effectiveLength := x2 - x1 + 1

		if blender, ok := any(pf.blender).(blender.BlenderGray16Linear); ok {
			// Blend each pixel with its corresponding coverage
			for i := 0; i < effectiveLength; i++ {
				coverIndex := coverOffset + i
				if coverIndex < len(covers) && covers[coverIndex] > 0 && c.A > 0 {
					blender.BlendPix(&row[x1+i], c.V, c.A, covers[coverIndex])
				}
			}
		}
	}
}

// BlendSolidVspan blends a vertical span with solid color and variable coverage
func (pf *PixFmtAlphaBlendGray16[B, CS]) BlendSolidVspan(x, y, length int, c color.Gray16[CS], covers []basics.Int16u) {
	if x < 0 || x >= pf.Width() || c.A == 0 || length <= 0 {
		return
	}

	y1 := Max(0, y)
	y2 := Min(pf.Height()-1, y+length-1)
	coverOffset := Max(0, -y)

	for i, currentY := 0, y1; currentY <= y2; i, currentY = i+1, currentY+1 {
		if coverOffset+i < len(covers) && covers[coverOffset+i] > 0 {
			pf.BlendPixel(x, currentY, c, covers[coverOffset+i])
		}
	}
}

// Color span operations

// CopyColorHspan copies a horizontal span of colors
func (pf *PixFmtAlphaBlendGray16[B, CS]) CopyColorHspan(x, y, length int, colors []color.Gray16[CS]) {
	if y < 0 || y >= pf.Height() || length <= 0 || len(colors) == 0 {
		return
	}

	x = Max(0, x)
	if x+length > pf.Width() {
		length = pf.Width() - x
	}

	for i := 0; i < length; i++ {
		colorIdx := i % len(colors)
		pf.CopyPixel(x+i, y, colors[colorIdx])
	}
}

// BlendColorHspan blends a horizontal span of colors
func (pf *PixFmtAlphaBlendGray16[B, CS]) BlendColorHspan(x, y, length int, colors []color.Gray16[CS], covers []basics.Int8u, cover basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || len(colors) == 0 {
		return
	}

	x = Max(0, x)
	if x+length > pf.Width() {
		length = pf.Width() - x
	}

	for i := 0; i < length; i++ {
		colorIdx := i % len(colors)
		c := colors[colorIdx]
		if c.A == 0 {
			continue
		}

		cvr := basics.Int16u(cover)
		if covers != nil && i < len(covers) {
			cvr = basics.Int16u(covers[i])
		}
		pf.BlendPixel(x+i, y, c, cvr)
	}
}

// CopyColorVspan copies a vertical span of colors
func (pf *PixFmtAlphaBlendGray16[B, CS]) CopyColorVspan(x, y, length int, colors []color.Gray16[CS]) {
	if x < 0 || x >= pf.Width() || length <= 0 || len(colors) == 0 {
		return
	}

	y = Max(0, y)
	if y+length > pf.Height() {
		length = pf.Height() - y
	}

	for i := 0; i < length; i++ {
		colorIdx := i % len(colors)
		pf.CopyPixel(x, y+i, colors[colorIdx])
	}
}

// BlendColorVspan blends a vertical span of colors
func (pf *PixFmtAlphaBlendGray16[B, CS]) BlendColorVspan(x, y, length int, colors []color.Gray16[CS], covers []basics.Int8u, cover basics.Int8u) {
	if x < 0 || x >= pf.Width() || length <= 0 || len(colors) == 0 {
		return
	}

	y = Max(0, y)
	if y+length > pf.Height() {
		length = pf.Height() - y
	}

	for i := 0; i < length; i++ {
		colorIdx := i % len(colors)
		c := colors[colorIdx]
		if c.A == 0 {
			continue
		}

		cvr := basics.Int16u(cover)
		if covers != nil && i < len(covers) {
			cvr = basics.Int16u(covers[i])
		}
		pf.BlendPixel(x, y+i, c, cvr)
	}
}

// Clear operations

// Clear fills the entire buffer with a color (sets alpha to 0)
func (pf *PixFmtAlphaBlendGray16[B, CS]) Clear(c color.Gray16[CS]) {
	for y := 0; y < pf.Height(); y++ {
		row := pf.RowPtr(y)
		for x := 0; x < pf.Width(); x++ {
			row[x] = c.V
		}
	}
}

// Fill fills the entire buffer with a color (ignores alpha)
func (pf *PixFmtAlphaBlendGray16[B, CS]) Fill(c color.Gray16[CS]) {
	pf.Clear(c)
}

// Concrete pixel format types
type (
	PixFmtGray16     = PixFmtAlphaBlendGray16[blender.BlenderGray16Linear, color.Linear]
	PixFmtSGray16    = PixFmtAlphaBlendGray16[blender.BlenderGray16SRGB, color.SRGB]
	PixFmtGray16Pre  = PixFmtAlphaBlendGray16[blender.BlenderGray16PreLinear, color.Linear]
	PixFmtSGray16Pre = PixFmtAlphaBlendGray16[blender.BlenderGray16PreSRGB, color.SRGB]
)

// Constructor functions for concrete types
func NewPixFmtGray16(rbuf *buffer.RenderingBufferU16) *PixFmtGray16 {
	return NewPixFmtAlphaBlendGray16[blender.BlenderGray16Linear, color.Linear](rbuf, blender.BlenderGray16Linear{})
}

func NewPixFmtSGray16(rbuf *buffer.RenderingBufferU16) *PixFmtSGray16 {
	return NewPixFmtAlphaBlendGray16[blender.BlenderGray16SRGB, color.SRGB](rbuf, blender.BlenderGray16SRGB{})
}

func NewPixFmtGray16Pre(rbuf *buffer.RenderingBufferU16) *PixFmtGray16Pre {
	return NewPixFmtAlphaBlendGray16[blender.BlenderGray16PreLinear, color.Linear](rbuf, blender.BlenderGray16PreLinear{})
}

func NewPixFmtSGray16Pre(rbuf *buffer.RenderingBufferU16) *PixFmtSGray16Pre {
	return NewPixFmtAlphaBlendGray16[blender.BlenderGray16PreSRGB, color.SRGB](rbuf, blender.BlenderGray16PreSRGB{})
}
