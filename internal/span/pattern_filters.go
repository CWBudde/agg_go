// Package span provides pattern filter functionality for AGG.
// This implements a port of AGG's agg_pattern_filters_rgba.h functionality.
package span

import (
	"agg_go/internal/color"
	"agg_go/internal/primitives"
)

// PatternFilterNN implements nearest neighbor pattern filtering.
// This is a port of AGG's pattern_filter_nn template struct.
type PatternFilterNN[ColorT any] struct{}

// NewPatternFilterNN creates a new nearest neighbor pattern filter.
func NewPatternFilterNN[ColorT any]() *PatternFilterNN[ColorT] {
	return &PatternFilterNN[ColorT]{}
}

// Dilation returns the filter dilation (0 for nearest neighbor).
func (pf *PatternFilterNN[ColorT]) Dilation() int {
	return 0
}

// PixelLowRes performs low-resolution pixel sampling (direct pixel access).
func (pf *PatternFilterNN[ColorT]) PixelLowRes(buf [][]ColorT, p *ColorT, x, y int) {
	if y >= 0 && y < len(buf) && x >= 0 && x < len(buf[y]) {
		*p = buf[y][x]
	}
}

// PixelHighRes performs high-resolution pixel sampling with subpixel coordinates.
func (pf *PatternFilterNN[ColorT]) PixelHighRes(buf [][]ColorT, p *ColorT, x, y int) {
	xLr := x >> primitives.LineSubpixelShift
	yLr := y >> primitives.LineSubpixelShift

	if yLr >= 0 && yLr < len(buf) && xLr >= 0 && xLr < len(buf[yLr]) {
		*p = buf[yLr][xLr]
	}
}

// PatternFilterBilinearRGBA8 implements bilinear pattern filtering for RGBA8 colors.
// This is a port of AGG's pattern_filter_bilinear_rgba template struct.
type PatternFilterBilinearRGBA8[CS any] struct{}

// NewPatternFilterBilinearRGBA8 creates a new bilinear RGBA8 pattern filter.
func NewPatternFilterBilinearRGBA8[CS any]() *PatternFilterBilinearRGBA8[CS] {
	return &PatternFilterBilinearRGBA8[CS]{}
}

// Dilation returns the filter dilation (1 for bilinear).
func (pf *PatternFilterBilinearRGBA8[CS]) Dilation() int {
	return 1
}

// PixelLowRes performs low-resolution pixel sampling (direct pixel access).
func (pf *PatternFilterBilinearRGBA8[CS]) PixelLowRes(buf [][]color.RGBA8[CS], p *color.RGBA8[CS], x, y int) {
	if y >= 0 && y < len(buf) && x >= 0 && x < len(buf[y]) {
		*p = buf[y][x]
	}
}

// PixelHighRes performs high-resolution bilinear pixel sampling for RGBA8 colors.
func (pf *PatternFilterBilinearRGBA8[CS]) PixelHighRes(buf [][]color.RGBA8[CS], p *color.RGBA8[CS], x, y int) {
	var r, g, b, a int

	// Extract subpixel coordinates
	xLr := x >> primitives.LineSubpixelShift
	yLr := y >> primitives.LineSubpixelShift

	x &= primitives.LineSubpixelMask
	y &= primitives.LineSubpixelMask

	// Bounds check - ensure we have valid access to all four corners
	if yLr < 0 || yLr >= len(buf) || xLr < 0 || xLr >= len(buf[yLr]) {
		*p = color.RGBA8[CS]{}
		return
	}

	// Sample 4 pixels for bilinear interpolation

	// Top-left pixel
	weight := (primitives.LineSubpixelScale - x) * (primitives.LineSubpixelScale - y)
	ptr := buf[yLr][xLr]
	r += weight * int(ptr.R)
	g += weight * int(ptr.G)
	b += weight * int(ptr.B)
	a += weight * int(ptr.A)

	// Top-right pixel
	if xLr+1 < len(buf[yLr]) {
		weight = x * (primitives.LineSubpixelScale - y)
		ptr = buf[yLr][xLr+1]
		r += weight * int(ptr.R)
		g += weight * int(ptr.G)
		b += weight * int(ptr.B)
		a += weight * int(ptr.A)
	}

	// Bottom-left pixel
	if yLr+1 < len(buf) && xLr < len(buf[yLr+1]) {
		weight = (primitives.LineSubpixelScale - x) * y
		ptr = buf[yLr+1][xLr]
		r += weight * int(ptr.R)
		g += weight * int(ptr.G)
		b += weight * int(ptr.B)
		a += weight * int(ptr.A)
	}

	// Bottom-right pixel
	if yLr+1 < len(buf) && xLr+1 < len(buf[yLr+1]) {
		weight = x * y
		ptr = buf[yLr+1][xLr+1]
		r += weight * int(ptr.R)
		g += weight * int(ptr.G)
		b += weight * int(ptr.B)
		a += weight * int(ptr.A)
	}

	// Downshift to get final values (divide by LineSubpixelScale^2)
	shift := primitives.LineSubpixelShift * 2
	p.R = uint8(r >> shift)
	p.G = uint8(g >> shift)
	p.B = uint8(b >> shift)
	p.A = uint8(a >> shift)
}

// PatternFilterBilinearRGBA16 implements bilinear pattern filtering for RGBA16 colors.
type PatternFilterBilinearRGBA16[CS any] struct{}

// NewPatternFilterBilinearRGBA16 creates a new bilinear RGBA16 pattern filter.
func NewPatternFilterBilinearRGBA16[CS any]() *PatternFilterBilinearRGBA16[CS] {
	return &PatternFilterBilinearRGBA16[CS]{}
}

// Dilation returns the filter dilation (1 for bilinear).
func (pf *PatternFilterBilinearRGBA16[CS]) Dilation() int {
	return 1
}

// PixelLowRes performs low-resolution pixel sampling (direct pixel access).
func (pf *PatternFilterBilinearRGBA16[CS]) PixelLowRes(buf [][]color.RGBA16[CS], p *color.RGBA16[CS], x, y int) {
	if y >= 0 && y < len(buf) && x >= 0 && x < len(buf[y]) {
		*p = buf[y][x]
	}
}

// PixelHighRes performs high-resolution bilinear pixel sampling for RGBA16 colors.
func (pf *PatternFilterBilinearRGBA16[CS]) PixelHighRes(buf [][]color.RGBA16[CS], p *color.RGBA16[CS], x, y int) {
	var r, g, b, a int

	// Extract subpixel coordinates
	xLr := x >> primitives.LineSubpixelShift
	yLr := y >> primitives.LineSubpixelShift

	x &= primitives.LineSubpixelMask
	y &= primitives.LineSubpixelMask

	// Bounds check - ensure we have valid access to all four corners
	if yLr < 0 || yLr >= len(buf) || xLr < 0 || xLr >= len(buf[yLr]) {
		*p = color.RGBA16[CS]{}
		return
	}

	// Sample 4 pixels for bilinear interpolation

	// Top-left pixel
	weight := (primitives.LineSubpixelScale - x) * (primitives.LineSubpixelScale - y)
	ptr := buf[yLr][xLr]
	r += weight * int(ptr.R)
	g += weight * int(ptr.G)
	b += weight * int(ptr.B)
	a += weight * int(ptr.A)

	// Top-right pixel
	if xLr+1 < len(buf[yLr]) {
		weight = x * (primitives.LineSubpixelScale - y)
		ptr = buf[yLr][xLr+1]
		r += weight * int(ptr.R)
		g += weight * int(ptr.G)
		b += weight * int(ptr.B)
		a += weight * int(ptr.A)
	}

	// Bottom-left pixel
	if yLr+1 < len(buf) && xLr < len(buf[yLr+1]) {
		weight = (primitives.LineSubpixelScale - x) * y
		ptr = buf[yLr+1][xLr]
		r += weight * int(ptr.R)
		g += weight * int(ptr.G)
		b += weight * int(ptr.B)
		a += weight * int(ptr.A)
	}

	// Bottom-right pixel
	if yLr+1 < len(buf) && xLr+1 < len(buf[yLr+1]) {
		weight = x * y
		ptr = buf[yLr+1][xLr+1]
		r += weight * int(ptr.R)
		g += weight * int(ptr.G)
		b += weight * int(ptr.B)
		a += weight * int(ptr.A)
	}

	// Downshift to get final values (divide by LineSubpixelScale^2)
	shift := primitives.LineSubpixelShift * 2
	p.R = uint16(r >> shift)
	p.G = uint16(g >> shift)
	p.B = uint16(b >> shift)
	p.A = uint16(a >> shift)
}

// PatternFilterBilinearRGBA32 implements bilinear pattern filtering for RGBA32 colors.
type PatternFilterBilinearRGBA32[CS any] struct{}

// NewPatternFilterBilinearRGBA32 creates a new bilinear RGBA32 pattern filter.
func NewPatternFilterBilinearRGBA32[CS any]() *PatternFilterBilinearRGBA32[CS] {
	return &PatternFilterBilinearRGBA32[CS]{}
}

// Dilation returns the filter dilation (1 for bilinear).
func (pf *PatternFilterBilinearRGBA32[CS]) Dilation() int {
	return 1
}

// PixelLowRes performs low-resolution pixel sampling (direct pixel access).
func (pf *PatternFilterBilinearRGBA32[CS]) PixelLowRes(buf [][]color.RGBA32[CS], p *color.RGBA32[CS], x, y int) {
	if y >= 0 && y < len(buf) && x >= 0 && x < len(buf[y]) {
		*p = buf[y][x]
	}
}

// PixelHighRes performs high-resolution bilinear pixel sampling for RGBA32 colors.
func (pf *PatternFilterBilinearRGBA32[CS]) PixelHighRes(buf [][]color.RGBA32[CS], p *color.RGBA32[CS], x, y int) {
	var r, g, b, a float32

	// Extract subpixel coordinates
	xLr := x >> primitives.LineSubpixelShift
	yLr := y >> primitives.LineSubpixelShift

	xf := float32(x&primitives.LineSubpixelMask) / float32(primitives.LineSubpixelScale)
	yf := float32(y&primitives.LineSubpixelMask) / float32(primitives.LineSubpixelScale)

	// Bounds check - ensure we have valid access to all four corners
	if yLr < 0 || yLr >= len(buf) || xLr < 0 || xLr >= len(buf[yLr]) {
		*p = color.RGBA32[CS]{}
		return
	}

	// Sample 4 pixels for bilinear interpolation

	// Top-left pixel
	weight := (1.0 - xf) * (1.0 - yf)
	ptr := buf[yLr][xLr]
	r += weight * ptr.R
	g += weight * ptr.G
	b += weight * ptr.B
	a += weight * ptr.A

	// Top-right pixel
	if xLr+1 < len(buf[yLr]) {
		weight = xf * (1.0 - yf)
		ptr = buf[yLr][xLr+1]
		r += weight * ptr.R
		g += weight * ptr.G
		b += weight * ptr.B
		a += weight * ptr.A
	}

	// Bottom-left pixel
	if yLr+1 < len(buf) && xLr < len(buf[yLr+1]) {
		weight = (1.0 - xf) * yf
		ptr = buf[yLr+1][xLr]
		r += weight * ptr.R
		g += weight * ptr.G
		b += weight * ptr.B
		a += weight * ptr.A
	}

	// Bottom-right pixel
	if yLr+1 < len(buf) && xLr+1 < len(buf[yLr+1]) {
		weight = xf * yf
		ptr = buf[yLr+1][xLr+1]
		r += weight * ptr.R
		g += weight * ptr.G
		b += weight * ptr.B
		a += weight * ptr.A
	}

	p.R = r
	p.G = g
	p.B = b
	p.A = a
}

// Type aliases matching C++ AGG naming
type PatternFilterNNRGBA8 = PatternFilterNN[color.RGBA8[color.Linear]]
type PatternFilterNNRGBA16 = PatternFilterNN[color.RGBA16[color.Linear]]
type PatternFilterNNRGBA32 = PatternFilterNN[color.RGBA32[color.Linear]]
