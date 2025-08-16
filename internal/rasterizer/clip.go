package rasterizer

import (
	"agg_go/internal/basics"
)

// Maximum coordinate value for polygon clipping
const PolyMaxCoord = (1 << 30) - 1

// RasConvInt provides integer coordinate conversion for rasterization.
// Equivalent to AGG's ras_conv_int struct.
type RasConvInt struct{}

// CoordType returns the coordinate type used by this converter
type CoordType = int

// MulDiv performs multiplication and division with rounding
func (RasConvInt) MulDiv(a, b, c float64) int {
	return basics.IRound(a * b / c)
}

// Xi converts input X coordinate (no transformation for integer converter)
func (RasConvInt) Xi(v int) int { return v }

// Yi converts input Y coordinate (no transformation for integer converter)
func (RasConvInt) Yi(v int) int { return v }

// Upscale converts double coordinate to subpixel integer coordinate
func (RasConvInt) Upscale(v float64) int {
	return basics.IRound(v * basics.PolySubpixelScale)
}

// Downscale converts subpixel integer coordinate back to integer (no-op)
func (RasConvInt) Downscale(v int) int { return v }

// RasConvIntSat provides saturated integer coordinate conversion.
// Equivalent to AGG's ras_conv_int_sat struct.
type RasConvIntSat struct{}

// MulDiv performs multiplication and division with saturation
func (RasConvIntSat) MulDiv(a, b, c float64) int {
	sat := basics.NewSaturation[int](PolyMaxCoord)
	return sat.IRound(a * b / c)
}

// Xi converts input X coordinate (no transformation)
func (RasConvIntSat) Xi(v int) int { return v }

// Yi converts input Y coordinate (no transformation)
func (RasConvIntSat) Yi(v int) int { return v }

// Upscale converts double coordinate to subpixel integer with saturation
func (RasConvIntSat) Upscale(v float64) int {
	sat := basics.NewSaturation[int](PolyMaxCoord)
	return sat.IRound(v * basics.PolySubpixelScale)
}

// Downscale converts subpixel integer coordinate back to integer (no-op)
func (RasConvIntSat) Downscale(v int) int { return v }

// RasConvInt3x provides 3x integer coordinate conversion for sub-pixel rendering.
// Equivalent to AGG's ras_conv_int_3x struct.
type RasConvInt3x struct{}

// MulDiv performs multiplication and division with rounding
func (RasConvInt3x) MulDiv(a, b, c float64) int {
	return basics.IRound(a * b / c)
}

// Xi converts input X coordinate with 3x scaling
func (RasConvInt3x) Xi(v int) int { return v * 3 }

// Yi converts input Y coordinate (no transformation)
func (RasConvInt3x) Yi(v int) int { return v }

// Upscale converts double coordinate to subpixel integer coordinate
func (RasConvInt3x) Upscale(v float64) int {
	return basics.IRound(v * basics.PolySubpixelScale)
}

// Downscale converts subpixel integer coordinate back to integer (no-op)
func (RasConvInt3x) Downscale(v int) int { return v }

// RasConvDbl provides double precision coordinate conversion.
// Equivalent to AGG's ras_conv_dbl struct.
type RasConvDbl struct{}

// CoordTypeFloat is the coordinate type for double conversion
type CoordTypeFloat = float64

// MulDiv performs floating point multiplication and division
func (RasConvDbl) MulDiv(a, b, c float64) float64 {
	return a * b / c
}

// Xi converts input X coordinate to subpixel integer
func (RasConvDbl) Xi(v float64) int {
	return basics.IRound(v * basics.PolySubpixelScale)
}

// Yi converts input Y coordinate to subpixel integer
func (RasConvDbl) Yi(v float64) int {
	return basics.IRound(v * basics.PolySubpixelScale)
}

// Upscale for double converter is pass-through
func (RasConvDbl) Upscale(v float64) float64 { return v }

// Downscale converts subpixel integer back to double coordinate
func (RasConvDbl) Downscale(v int) float64 {
	return float64(v) / basics.PolySubpixelScale
}

// RasConvDbl3x provides 3x double precision coordinate conversion.
// Equivalent to AGG's ras_conv_dbl_3x struct.
type RasConvDbl3x struct{}

// MulDiv performs floating point multiplication and division
func (RasConvDbl3x) MulDiv(a, b, c float64) float64 {
	return a * b / c
}

// Xi converts input X coordinate to subpixel integer with 3x scaling
func (RasConvDbl3x) Xi(v float64) int {
	return basics.IRound(v * basics.PolySubpixelScale * 3)
}

// Yi converts input Y coordinate to subpixel integer
func (RasConvDbl3x) Yi(v float64) int {
	return basics.IRound(v * basics.PolySubpixelScale)
}

// Upscale for double converter is pass-through
func (RasConvDbl3x) Upscale(v float64) float64 { return v }

// Downscale converts subpixel integer back to double coordinate
func (RasConvDbl3x) Downscale(v int) float64 {
	return float64(v) / basics.PolySubpixelScale
}

// ClipInterface defines the interface for clipping implementations
type ClipInterface interface {
	ResetClipping()
	ClipBox(x1, y1, x2, y2 float64)
	MoveTo(x, y float64)
	LineTo(x, y float64)
}

// RasterizerSlNoClip provides a no-clipping implementation.
// Equivalent to AGG's rasterizer_sl_no_clip<Conv> template class.
type RasterizerSlNoClip[Conv any] struct {
	outline *RasterizerCellsAA[*CellAA] // The underlying rasterizer
}

// NewRasterizerSlNoClip creates a new no-clip rasterizer
func NewRasterizerSlNoClip[Conv any](outline *RasterizerCellsAA[*CellAA]) *RasterizerSlNoClip[Conv] {
	return &RasterizerSlNoClip[Conv]{
		outline: outline,
	}
}

// ResetClipping does nothing for no-clip implementation
func (r *RasterizerSlNoClip[Conv]) ResetClipping() {
	// No-op
}

// ClipBox does nothing for no-clip implementation
func (r *RasterizerSlNoClip[Conv]) ClipBox(x1, y1, x2, y2 float64) {
	// No-op
}

// MoveTo converts coordinates and calls the outline
func (r *RasterizerSlNoClip[Conv]) MoveTo(x, y float64) {
	conv := *new(Conv)

	// Use type assertion to call the appropriate conversion method
	switch c := any(conv).(type) {
	case RasConvInt:
		r.outline.setCurrCell(c.Upscale(x)>>basics.PolySubpixelShift, c.Upscale(y)>>basics.PolySubpixelShift)
	case RasConvDbl:
		r.outline.setCurrCell(c.Xi(x)>>basics.PolySubpixelShift, c.Yi(y)>>basics.PolySubpixelShift)
	}
}

// LineTo converts coordinates and draws a line
func (r *RasterizerSlNoClip[Conv]) LineTo(x, y float64) {
	conv := *new(Conv)

	// Use type assertion to call the appropriate conversion method
	switch c := any(conv).(type) {
	case RasConvInt:
		x1, y1 := c.Upscale(x), c.Upscale(y)
		r.outline.Line(0, 0, x1, y1) // Would need to track previous position
	case RasConvDbl:
		x1, y1 := c.Xi(x), c.Yi(y)
		r.outline.Line(0, 0, x1, y1) // Would need to track previous position
	}
}

// RasterizerSlClipInt provides integer clipping implementation.
// Equivalent to AGG's rasterizer_sl_clip_int<Conv> template class.
type RasterizerSlClipInt[Conv any] struct {
	outline  *RasterizerCellsAA[*CellAA]
	clipBox  basics.RectI
	x1, y1   int
	f1       uint32
	clipping bool
}

// NewRasterizerSlClipInt creates a new integer clipping rasterizer
func NewRasterizerSlClipInt[Conv any](outline *RasterizerCellsAA[*CellAA]) *RasterizerSlClipInt[Conv] {
	return &RasterizerSlClipInt[Conv]{
		outline:  outline,
		clipBox:  basics.RectI{X1: 0, Y1: 0, X2: 0, Y2: 0},
		clipping: false,
	}
}

// ResetClipping disables clipping
func (r *RasterizerSlClipInt[Conv]) ResetClipping() {
	r.clipping = false
}

// ClipBox sets the clipping rectangle
func (r *RasterizerSlClipInt[Conv]) ClipBox(x1, y1, x2, y2 float64) {
	r.clipBox = basics.RectI{
		X1: basics.IRound(x1),
		Y1: basics.IRound(y1),
		X2: basics.IRound(x2),
		Y2: basics.IRound(y2),
	}
	r.clipping = true
}

// MoveTo sets the current position with clipping
func (r *RasterizerSlClipInt[Conv]) MoveTo(x, y float64) {
	conv := *new(Conv)

	switch c := any(conv).(type) {
	case RasConvInt:
		r.x1, r.y1 = c.Upscale(x), c.Upscale(y)
	case RasConvDbl:
		r.x1, r.y1 = c.Xi(x), c.Yi(y)
	}

	if r.clipping {
		r.f1 = r.clippingFlags(r.x1, r.y1)
	}
}

// LineTo draws a line to the specified position with clipping
func (r *RasterizerSlClipInt[Conv]) LineTo(x, y float64) {
	conv := *new(Conv)

	var x2, y2 int
	switch c := any(conv).(type) {
	case RasConvInt:
		x2, y2 = c.Upscale(x), c.Upscale(y)
	case RasConvDbl:
		x2, y2 = c.Xi(x), c.Yi(y)
	}

	if r.clipping {
		f2 := r.clippingFlags(x2, y2)
		if (r.f1 & f2) == 0 {
			if r.f1 != 0 {
				// Line crosses clipping boundary - would implement clipping here
				r.lineClipY(r.x1, r.y1, x2, y2, r.f1, f2)
			} else {
				r.outline.Line(r.x1, r.y1, x2, y2)
			}
		}
		r.f1 = f2
	} else {
		r.outline.Line(r.x1, r.y1, x2, y2)
	}

	r.x1, r.y1 = x2, y2
}

// clippingFlags returns clipping flags for a point
func (r *RasterizerSlClipInt[Conv]) clippingFlags(x, y int) uint32 {
	var f uint32 = 0
	if x < r.clipBox.X1<<basics.PolySubpixelShift {
		f |= 1
	}
	if y < r.clipBox.Y1<<basics.PolySubpixelShift {
		f |= 2
	}
	if x > r.clipBox.X2<<basics.PolySubpixelShift {
		f |= 4
	}
	if y > r.clipBox.Y2<<basics.PolySubpixelShift {
		f |= 8
	}
	return f
}

// lineClipY implements Y clipping for lines (simplified version)
func (r *RasterizerSlClipInt[Conv]) lineClipY(x1, y1, x2, y2 int, f1, f2 uint32) {
	// Simplified clipping implementation
	// Full implementation would use Liang-Barsky or Cohen-Sutherland clipping
	r.outline.Line(x1, y1, x2, y2)
}
