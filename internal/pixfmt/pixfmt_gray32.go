package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/pixfmt/blender"
)

// PixFmtAlphaBlendGray32 implements alpha blending for 32-bit float grayscale pixel formats
type PixFmtAlphaBlendGray32[B any, CS any] struct {
	rbuf     *buffer.RenderingBufferF32
	blender  B
	category PixFmtGrayTag
}

// Gray32PixelType represents a 32-bit float grayscale pixel
type Gray32PixelType struct {
	V float32 // Grayscale value
}

// Set sets the grayscale value
func (p *Gray32PixelType) Set(v float32) {
	p.V = v
}

// NewPixFmtAlphaBlendGray32 creates a new 32-bit float grayscale pixel format
func NewPixFmtAlphaBlendGray32[B any, CS any](rbuf *buffer.RenderingBufferF32, blender B) *PixFmtAlphaBlendGray32[B, CS] {
	return &PixFmtAlphaBlendGray32[B, CS]{
		rbuf:    rbuf,
		blender: blender,
	}
}

// Basic properties
func (pf *PixFmtAlphaBlendGray32[B, CS]) Width() int {
	return pf.rbuf.Width()
}

func (pf *PixFmtAlphaBlendGray32[B, CS]) Height() int {
	return pf.rbuf.Height()
}

func (pf *PixFmtAlphaBlendGray32[B, CS]) PixWidth() int {
	return 4 // 4 bytes per pixel for 32-bit float grayscale
}

func (pf *PixFmtAlphaBlendGray32[B, CS]) Stride() int {
	return pf.rbuf.Stride()
}

// RowPtr returns a pointer to the pixel data for the given row
func (pf *PixFmtAlphaBlendGray32[B, CS]) RowPtr(y int) []float32 {
	return buffer.RowF32(pf.rbuf, y)
}

// PixPtr returns a pointer to the specific pixel
func (pf *PixFmtAlphaBlendGray32[B, CS]) PixPtr(x, y int) *float32 {
	row := buffer.RowF32(pf.rbuf, y)
	if x >= 0 && x < len(row) {
		return &row[x]
	}
	return nil
}

// MakePix creates a pixel value from grayscale components
func (pf *PixFmtAlphaBlendGray32[B, CS]) MakePix(v float32) Gray32PixelType {
	return Gray32PixelType{V: v}
}

// Core pixel operations

// CopyPixel copies a pixel without blending
func (pf *PixFmtAlphaBlendGray32[B, CS]) CopyPixel(x, y int, c color.Gray32[CS]) {
	if InBounds(x, y, pf.Width(), pf.Height()) {
		pixel := pf.PixPtr(x, y)
		if pixel != nil {
			*pixel = c.V
		}
	}
}

// BlendPixel blends a pixel with coverage
func (pf *PixFmtAlphaBlendGray32[B, CS]) BlendPixel(x, y int, c color.Gray32[CS], cover basics.Int8u) {
	if InBounds(x, y, pf.Width(), pf.Height()) && c.A > 0.0 {
		pixel := pf.PixPtr(x, y)
		if pixel != nil {
			if blender, ok := any(pf.blender).(blender.Gray32Blender); ok {
				// Convert Int8u cover to float32 (0-255 -> 0.0-1.0)
				floatCover := float32(cover) / 255.0
				blender.BlendPix(pixel, c.V, c.A, floatCover)
			}
		}
	}
}

// GetPixel gets the pixel color at the specified coordinates
func (pf *PixFmtAlphaBlendGray32[B, CS]) GetPixel(x, y int) color.Gray32[CS] {
	if InBounds(x, y, pf.Width(), pf.Height()) {
		pixel := pf.PixPtr(x, y)
		if pixel != nil {
			return color.NewGray32WithAlpha[CS](*pixel, 1.0) // Full alpha
		}
	}
	return color.Gray32[CS]{}
}

// Pixel returns the pixel at the given coordinates (alias for GetPixel to satisfy interface)
func (pf *PixFmtAlphaBlendGray32[B, CS]) Pixel(x, y int) color.Gray32[CS] {
	return pf.GetPixel(x, y)
}

// Line operations

// CopyHline copies a horizontal line
func (pf *PixFmtAlphaBlendGray32[B, CS]) CopyHline(x, y, length int, c color.Gray32[CS]) {
	if y < 0 || y >= pf.Height() || length <= 0 {
		return
	}

	x = Max(0, x)
	if x+length > pf.Width() {
		length = pf.Width() - x
	}

	row := pf.RowPtr(y)
	blender.CopyGray32Hline(row, x, length, c)
}

// BlendHline blends a horizontal line with coverage
func (pf *PixFmtAlphaBlendGray32[B, CS]) BlendHline(x, y, length int, c color.Gray32[CS], cover basics.Int8u) {
	if y < 0 || y >= pf.Height() || length <= 0 || c.A == 0.0 {
		return
	}

	x = Max(0, x)
	if x+length > pf.Width() {
		length = pf.Width() - x
	}

	row := pf.RowPtr(y)
	if blender, ok := any(pf.blender).(blender.Gray32Blender); ok {
		floatCover := float32(cover) / 255.0
		for i := 0; i < length; i++ {
			if c.A > 0.0 {
				blender.BlendPix(&row[x+i], c.V, c.A, floatCover)
			}
		}
	}
}

// CopyVline copies a vertical line
func (pf *PixFmtAlphaBlendGray32[B, CS]) CopyVline(x, y1, y2 int, c color.Gray32[CS]) {
	if x < 0 || x >= pf.Width() {
		return
	}

	if y1 > y2 {
		y1, y2 = y2, y1
	}

	y1 = Max(0, y1)
	y2 = Min(pf.Height()-1, y2)

	for y := y1; y <= y2; y++ {
		pf.CopyPixel(x, y, c)
	}
}

// BlendVline blends a vertical line with coverage
func (pf *PixFmtAlphaBlendGray32[B, CS]) BlendVline(x, y1, y2 int, c color.Gray32[CS], cover basics.Int8u) {
	if x < 0 || x >= pf.Width() || c.A == 0.0 {
		return
	}

	if y1 > y2 {
		y1, y2 = y2, y1
	}

	y1 = Max(0, y1)
	y2 = Min(pf.Height()-1, y2)

	for y := y1; y <= y2; y++ {
		pf.BlendPixel(x, y, c, cover)
	}
}

// Rectangle operations

// CopyBar copies a filled rectangle
func (pf *PixFmtAlphaBlendGray32[B, CS]) CopyBar(x1, y1, x2, y2 int, c color.Gray32[CS]) {
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
		pf.CopyHline(x1, y, x2, c)
	}
}

// BlendBar blends a filled rectangle with coverage
func (pf *PixFmtAlphaBlendGray32[B, CS]) BlendBar(x1, y1, x2, y2 int, c color.Gray32[CS], cover basics.Int8u) {
	if c.A == 0.0 {
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
		pf.BlendHline(x1, y, x2, c, cover)
	}
}

// Span operations

// BlendSolidHspan blends a horizontal span with solid color and variable coverage
func (pf *PixFmtAlphaBlendGray32[B, CS]) BlendSolidHspan(x, y, length int, c color.Gray32[CS], covers []basics.Int8u) {
	if y < 0 || y >= pf.Height() || c.A == 0.0 || length <= 0 {
		return
	}

	x1 := Max(0, x)
	x2 := Min(pf.Width()-1, x+length-1)

	if x1 <= x2 {
		row := pf.RowPtr(y)
		coverOffset := Max(0, -x)
		effectiveLength := x2 - x1 + 1

		if blender, ok := any(pf.blender).(blender.Gray32Blender); ok {
			// Blend each pixel with its corresponding coverage
			for i := 0; i < effectiveLength; i++ {
				coverIndex := coverOffset + i
				if coverIndex < len(covers) && covers[coverIndex] > 0 && c.A > 0.0 {
					// Convert Int8u cover to float32 (0-255 -> 0.0-1.0)
					floatCover := float32(covers[coverIndex]) / 255.0
					blender.BlendPix(&row[x1+i], c.V, c.A, floatCover)
				}
			}
		}
	}
}

// BlendSolidVspan blends a vertical span with solid color and variable coverage
func (pf *PixFmtAlphaBlendGray32[B, CS]) BlendSolidVspan(x, y, length int, c color.Gray32[CS], covers []basics.Int8u) {
	if x < 0 || x >= pf.Width() || c.A == 0.0 || length <= 0 {
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

// CopyColorHspan copies a horizontal span of colors
func (pf *PixFmtAlphaBlendGray32[B, CS]) CopyColorHspan(x, y, length int, colors []color.Gray32[CS]) {
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
func (pf *PixFmtAlphaBlendGray32[B, CS]) BlendColorHspan(x, y, length int, colors []color.Gray32[CS], covers []basics.Int8u, cover basics.Int8u) {
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
		if c.A == 0.0 {
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
func (pf *PixFmtAlphaBlendGray32[B, CS]) CopyColorVspan(x, y, length int, colors []color.Gray32[CS]) {
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
func (pf *PixFmtAlphaBlendGray32[B, CS]) BlendColorVspan(x, y, length int, colors []color.Gray32[CS], covers []basics.Int8u, cover basics.Int8u) {
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
		if c.A == 0.0 {
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

// Clear fills the entire buffer with a color (sets alpha to 0)
func (pf *PixFmtAlphaBlendGray32[B, CS]) Clear(c color.Gray32[CS]) {
	for y := 0; y < pf.Height(); y++ {
		row := pf.RowPtr(y)
		for x := 0; x < pf.Width(); x++ {
			row[x] = c.V
		}
	}
}

// Fill fills the entire buffer with a color (ignores alpha)
func (pf *PixFmtAlphaBlendGray32[B, CS]) Fill(c color.Gray32[CS]) {
	pf.Clear(c)
}

// Concrete pixel format types
type (
	PixFmtGray32     = PixFmtAlphaBlendGray32[blender.BlenderGray32Linear, color.Linear]
	PixFmtSGray32    = PixFmtAlphaBlendGray32[blender.BlenderGray32SRGB, color.SRGB]
	PixFmtGray32Pre  = PixFmtAlphaBlendGray32[blender.BlenderGray32PreLinear, color.Linear]
	PixFmtSGray32Pre = PixFmtAlphaBlendGray32[blender.BlenderGray32PreSRGB, color.SRGB]
)

// Constructor functions for concrete types
func NewPixFmtGray32(rbuf *buffer.RenderingBufferF32) *PixFmtGray32 {
	return NewPixFmtAlphaBlendGray32[blender.BlenderGray32Linear, color.Linear](rbuf, blender.BlenderGray32Linear{})
}

func NewPixFmtSGray32(rbuf *buffer.RenderingBufferF32) *PixFmtSGray32 {
	return NewPixFmtAlphaBlendGray32[blender.BlenderGray32SRGB, color.SRGB](rbuf, blender.BlenderGray32SRGB{})
}

func NewPixFmtGray32Pre(rbuf *buffer.RenderingBufferF32) *PixFmtGray32Pre {
	return NewPixFmtAlphaBlendGray32[blender.BlenderGray32PreLinear, color.Linear](rbuf, blender.BlenderGray32PreLinear{})
}

func NewPixFmtSGray32Pre(rbuf *buffer.RenderingBufferF32) *PixFmtSGray32Pre {
	return NewPixFmtAlphaBlendGray32[blender.BlenderGray32PreSRGB, color.SRGB](rbuf, blender.BlenderGray32PreSRGB{})
}
