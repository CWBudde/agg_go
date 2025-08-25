package rasterizer

import (
	"agg_go/internal/basics"
)

// Maximum coordinate value for polygon clipping
const PolyMaxCoord = (1 << 30) - 1

// ConverterInterface defines the interface for coordinate conversion
type ConverterInterface interface {
	MulDiv(a, b, c float64) float64
	Xi(v any) int
	Yi(v any) int
	Upscale(v float64) any
	Downscale(v int) any
}

// RasConvInt provides integer coordinate conversion for rasterization.
// Equivalent to AGG's ras_conv_int struct.
type RasConvInt struct{}

// CoordType returns the coordinate type used by this converter
type CoordType = int

// MulDiv performs multiplication and division with rounding
func (RasConvInt) MulDiv(a, b, c float64) float64 {
	return float64(basics.IRound(a * b / c))
}

// Xi converts input X coordinate (no transformation for integer converter)
func (RasConvInt) Xi(v any) int {
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return basics.IRound(val)
	default:
		return 0
	}
}

// Yi converts input Y coordinate (no transformation for integer converter)
func (RasConvInt) Yi(v any) int {
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return basics.IRound(val)
	default:
		return 0
	}
}

// Upscale converts double coordinate to subpixel integer coordinate
func (RasConvInt) Upscale(v float64) any {
	return basics.IRound(v * basics.PolySubpixelScale)
}

// Downscale converts subpixel integer coordinate back to integer (no-op)
func (RasConvInt) Downscale(v int) any { return v }

// RasConvIntSat provides saturated integer coordinate conversion.
// Equivalent to AGG's ras_conv_int_sat struct.
type RasConvIntSat struct{}

// MulDiv performs multiplication and division with saturation
func (RasConvIntSat) MulDiv(a, b, c float64) float64 {
	sat := basics.NewSaturation[int](PolyMaxCoord)
	return float64(sat.IRound(a * b / c))
}

// Xi converts input X coordinate (no transformation)
func (RasConvIntSat) Xi(v any) int {
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return basics.IRound(val)
	default:
		return 0
	}
}

// Yi converts input Y coordinate (no transformation)
func (RasConvIntSat) Yi(v any) int {
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return basics.IRound(val)
	default:
		return 0
	}
}

// Upscale converts double coordinate to subpixel integer with saturation
func (RasConvIntSat) Upscale(v float64) any {
	sat := basics.NewSaturation[int](PolyMaxCoord)
	return sat.IRound(v * basics.PolySubpixelScale)
}

// Downscale converts subpixel integer coordinate back to integer (no-op)
func (RasConvIntSat) Downscale(v int) any { return v }

// RasConvInt3x provides 3x integer coordinate conversion for sub-pixel rendering.
// Equivalent to AGG's ras_conv_int_3x struct.
type RasConvInt3x struct{}

// MulDiv performs multiplication and division with rounding
func (RasConvInt3x) MulDiv(a, b, c float64) float64 {
	return float64(basics.IRound(a * b / c))
}

// Xi converts input X coordinate with 3x scaling
func (RasConvInt3x) Xi(v any) int {
	switch val := v.(type) {
	case int:
		return val * 3
	case float64:
		return basics.IRound(val) * 3
	default:
		return 0
	}
}

// Yi converts input Y coordinate (no transformation)
func (RasConvInt3x) Yi(v any) int {
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return basics.IRound(val)
	default:
		return 0
	}
}

// Upscale converts double coordinate to subpixel integer coordinate
func (RasConvInt3x) Upscale(v float64) any {
	return basics.IRound(v * basics.PolySubpixelScale)
}

// Downscale converts subpixel integer coordinate back to integer (no-op)
func (RasConvInt3x) Downscale(v int) any { return v }

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
func (RasConvDbl) Xi(v any) int {
	switch val := v.(type) {
	case float64:
		return basics.IRound(val * basics.PolySubpixelScale)
	case int:
		return basics.IRound(float64(val) * basics.PolySubpixelScale)
	default:
		return 0
	}
}

// Yi converts input Y coordinate to subpixel integer
func (RasConvDbl) Yi(v any) int {
	switch val := v.(type) {
	case float64:
		return basics.IRound(val * basics.PolySubpixelScale)
	case int:
		return basics.IRound(float64(val) * basics.PolySubpixelScale)
	default:
		return 0
	}
}

// Upscale for double converter is pass-through
func (RasConvDbl) Upscale(v float64) any { return v }

// Downscale converts subpixel integer back to double coordinate
func (RasConvDbl) Downscale(v int) any {
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
func (RasConvDbl3x) Xi(v any) int {
	switch val := v.(type) {
	case float64:
		return basics.IRound(val * basics.PolySubpixelScale * 3)
	case int:
		return basics.IRound(float64(val) * basics.PolySubpixelScale * 3)
	default:
		return 0
	}
}

// Yi converts input Y coordinate to subpixel integer
func (RasConvDbl3x) Yi(v any) int {
	switch val := v.(type) {
	case float64:
		return basics.IRound(val * basics.PolySubpixelScale)
	case int:
		return basics.IRound(float64(val) * basics.PolySubpixelScale)
	default:
		return 0
	}
}

// Upscale for double converter is pass-through
func (RasConvDbl3x) Upscale(v float64) any { return v }

// Downscale converts subpixel integer back to double coordinate
func (RasConvDbl3x) Downscale(v int) any {
	return float64(v) / basics.PolySubpixelScale
}

// RasterizerInterface defines the interface that rasterizers must implement
type RasterizerInterface interface {
	Line(x1, y1, x2, y2 int)
}

// ClipInterface defines the interface for clipping implementations
type ClipInterface interface {
	ResetClipping()
	ClipBox(x1, y1, x2, y2 float64)
	MoveTo(x, y float64)
	LineTo(outline RasterizerInterface, x, y float64)
}

// RasterizerSlClip implements the scanline clipping rasterizer.
// Equivalent to AGG's rasterizer_sl_clip<Conv> template class.
type RasterizerSlClip[Conv any] struct {
	clipBox  basics.Rect[float64]
	x1, y1   float64
	f1       uint32
	clipping bool
}

// NewRasterizerSlClip creates a new scanline clipping rasterizer
func NewRasterizerSlClip[Conv any]() *RasterizerSlClip[Conv] {
	return &RasterizerSlClip[Conv]{
		clipBox:  basics.Rect[float64]{X1: 0, Y1: 0, X2: 0, Y2: 0},
		clipping: false,
	}
}

// ResetClipping disables clipping
func (r *RasterizerSlClip[Conv]) ResetClipping() {
	r.clipping = false
}

// ClipBox sets the clipping rectangle and enables clipping
func (r *RasterizerSlClip[Conv]) ClipBox(x1, y1, x2, y2 float64) {
	r.clipBox = basics.Rect[float64]{X1: x1, Y1: y1, X2: x2, Y2: y2}
	// Normalize the rectangle
	if r.clipBox.X1 > r.clipBox.X2 {
		r.clipBox.X1, r.clipBox.X2 = r.clipBox.X2, r.clipBox.X1
	}
	if r.clipBox.Y1 > r.clipBox.Y2 {
		r.clipBox.Y1, r.clipBox.Y2 = r.clipBox.Y2, r.clipBox.Y1
	}
	r.clipping = true
}

// MoveTo sets the current position
func (r *RasterizerSlClip[Conv]) MoveTo(x1, y1 float64) {
	r.x1 = x1
	r.y1 = y1
	if r.clipping {
		r.f1 = basics.ClippingFlags(x1, y1, r.clipBox)
	}
}

// lineClipY implements Y-axis clipping for lines
func (r *RasterizerSlClip[Conv]) lineClipY(
	ras RasterizerInterface,
	x1, y1, x2, y2 float64,
	f1, f2 uint32,
) {
	conv := *new(Conv)

	f1 &= 10 // Keep only Y flags (8 + 2)
	f2 &= 10

	if (f1 | f2) == 0 {
		// Fully visible
		switch c := any(conv).(type) {
		case RasConvInt:
			ras.Line(c.Xi(x1), c.Yi(y1), c.Xi(x2), c.Yi(y2))
		case RasConvIntSat:
			ras.Line(c.Xi(x1), c.Yi(y1), c.Xi(x2), c.Yi(y2))
		case RasConvInt3x:
			ras.Line(c.Xi(x1), c.Yi(y1), c.Xi(x2), c.Yi(y2))
		case RasConvDbl:
			ras.Line(c.Xi(x1), c.Yi(y1), c.Xi(x2), c.Yi(y2))
		case RasConvDbl3x:
			ras.Line(c.Xi(x1), c.Yi(y1), c.Xi(x2), c.Yi(y2))
		}
		return
	}

	if f1 == f2 {
		// Invisible by Y
		return
	}

	tx1, ty1 := x1, y1
	tx2, ty2 := x2, y2

	if f1&8 != 0 { // y1 < clip.y1
		switch c := any(conv).(type) {
		case RasConvInt:
			tx1 = x1 + c.MulDiv(r.clipBox.Y1-y1, x2-x1, y2-y1)
		case RasConvIntSat:
			tx1 = x1 + c.MulDiv(r.clipBox.Y1-y1, x2-x1, y2-y1)
		case RasConvInt3x:
			tx1 = x1 + c.MulDiv(r.clipBox.Y1-y1, x2-x1, y2-y1)
		case RasConvDbl:
			tx1 = x1 + c.MulDiv(r.clipBox.Y1-y1, x2-x1, y2-y1)
		case RasConvDbl3x:
			tx1 = x1 + c.MulDiv(r.clipBox.Y1-y1, x2-x1, y2-y1)
		}
		ty1 = r.clipBox.Y1
	}

	if f1&2 != 0 { // y1 > clip.y2
		switch c := any(conv).(type) {
		case RasConvInt:
			tx1 = x1 + c.MulDiv(r.clipBox.Y2-y1, x2-x1, y2-y1)
		case RasConvIntSat:
			tx1 = x1 + c.MulDiv(r.clipBox.Y2-y1, x2-x1, y2-y1)
		case RasConvInt3x:
			tx1 = x1 + c.MulDiv(r.clipBox.Y2-y1, x2-x1, y2-y1)
		case RasConvDbl:
			tx1 = x1 + c.MulDiv(r.clipBox.Y2-y1, x2-x1, y2-y1)
		case RasConvDbl3x:
			tx1 = x1 + c.MulDiv(r.clipBox.Y2-y1, x2-x1, y2-y1)
		}
		ty1 = r.clipBox.Y2
	}

	if f2&8 != 0 { // y2 < clip.y1
		switch c := any(conv).(type) {
		case RasConvInt:
			tx2 = x1 + c.MulDiv(r.clipBox.Y1-y1, x2-x1, y2-y1)
		case RasConvIntSat:
			tx2 = x1 + c.MulDiv(r.clipBox.Y1-y1, x2-x1, y2-y1)
		case RasConvInt3x:
			tx2 = x1 + c.MulDiv(r.clipBox.Y1-y1, x2-x1, y2-y1)
		case RasConvDbl:
			tx2 = x1 + c.MulDiv(r.clipBox.Y1-y1, x2-x1, y2-y1)
		case RasConvDbl3x:
			tx2 = x1 + c.MulDiv(r.clipBox.Y1-y1, x2-x1, y2-y1)
		}
		ty2 = r.clipBox.Y1
	}

	if f2&2 != 0 { // y2 > clip.y2
		switch c := any(conv).(type) {
		case RasConvInt:
			tx2 = x1 + c.MulDiv(r.clipBox.Y2-y1, x2-x1, y2-y1)
		case RasConvIntSat:
			tx2 = x1 + c.MulDiv(r.clipBox.Y2-y1, x2-x1, y2-y1)
		case RasConvInt3x:
			tx2 = x1 + c.MulDiv(r.clipBox.Y2-y1, x2-x1, y2-y1)
		case RasConvDbl:
			tx2 = x1 + c.MulDiv(r.clipBox.Y2-y1, x2-x1, y2-y1)
		case RasConvDbl3x:
			tx2 = x1 + c.MulDiv(r.clipBox.Y2-y1, x2-x1, y2-y1)
		}
		ty2 = r.clipBox.Y2
	}

	switch c := any(conv).(type) {
	case RasConvInt:
		ras.Line(c.Xi(tx1), c.Yi(ty1), c.Xi(tx2), c.Yi(ty2))
	case RasConvIntSat:
		ras.Line(c.Xi(tx1), c.Yi(ty1), c.Xi(tx2), c.Yi(ty2))
	case RasConvInt3x:
		ras.Line(c.Xi(tx1), c.Yi(ty1), c.Xi(tx2), c.Yi(ty2))
	case RasConvDbl:
		ras.Line(c.Xi(tx1), c.Yi(ty1), c.Xi(tx2), c.Yi(ty2))
	case RasConvDbl3x:
		ras.Line(c.Xi(tx1), c.Yi(ty1), c.Xi(tx2), c.Yi(ty2))
	}
}

// LineTo draws a line from the current position to (x2, y2) with clipping
func (r *RasterizerSlClip[Conv]) LineTo(outline RasterizerInterface, x2, y2 float64) {
	ras := outline
	conv := *new(Conv)

	if r.clipping {
		f2 := basics.ClippingFlags(x2, y2, r.clipBox)

		if (r.f1&10) == (f2&10) && (r.f1&10) != 0 {
			// Invisible by Y
			r.x1 = x2
			r.y1 = y2
			r.f1 = f2
			return
		}

		x1, y1 := r.x1, r.y1
		f1 := r.f1
		var y3, y4 float64
		var f3, f4 uint32

		// Handle X clipping cases
		switch ((f1 & 5) << 1) | (f2 & 5) {
		case 0: // Visible by X
			r.lineClipY(ras, x1, y1, x2, y2, f1, f2)

		case 1: // x2 > clip.x2
			switch c := any(conv).(type) {
			case RasConvInt:
				y3 = y1 + c.MulDiv(r.clipBox.X2-x1, y2-y1, x2-x1)
			case RasConvIntSat:
				y3 = y1 + c.MulDiv(r.clipBox.X2-x1, y2-y1, x2-x1)
			case RasConvInt3x:
				y3 = y1 + c.MulDiv(r.clipBox.X2-x1, y2-y1, x2-x1)
			case RasConvDbl:
				y3 = y1 + c.MulDiv(r.clipBox.X2-x1, y2-y1, x2-x1)
			case RasConvDbl3x:
				y3 = y1 + c.MulDiv(r.clipBox.X2-x1, y2-y1, x2-x1)
			}
			f3 = basics.ClippingFlagsY(y3, r.clipBox)
			r.lineClipY(ras, x1, y1, r.clipBox.X2, y3, f1, f3)
			r.lineClipY(ras, r.clipBox.X2, y3, r.clipBox.X2, y2, f3, f2)

		case 2: // x1 > clip.x2
			switch c := any(conv).(type) {
			case RasConvInt:
				y3 = y1 + c.MulDiv(r.clipBox.X2-x1, y2-y1, x2-x1)
			case RasConvIntSat:
				y3 = y1 + c.MulDiv(r.clipBox.X2-x1, y2-y1, x2-x1)
			case RasConvInt3x:
				y3 = y1 + c.MulDiv(r.clipBox.X2-x1, y2-y1, x2-x1)
			case RasConvDbl:
				y3 = y1 + c.MulDiv(r.clipBox.X2-x1, y2-y1, x2-x1)
			case RasConvDbl3x:
				y3 = y1 + c.MulDiv(r.clipBox.X2-x1, y2-y1, x2-x1)
			}
			f3 = basics.ClippingFlagsY(y3, r.clipBox)
			r.lineClipY(ras, r.clipBox.X2, y1, r.clipBox.X2, y3, f1, f3)
			r.lineClipY(ras, r.clipBox.X2, y3, x2, y2, f3, f2)

		case 3: // x1 > clip.x2 && x2 > clip.x2
			r.lineClipY(ras, r.clipBox.X2, y1, r.clipBox.X2, y2, f1, f2)

		case 4: // x2 < clip.x1
			switch c := any(conv).(type) {
			case RasConvInt:
				y3 = y1 + c.MulDiv(r.clipBox.X1-x1, y2-y1, x2-x1)
			case RasConvIntSat:
				y3 = y1 + c.MulDiv(r.clipBox.X1-x1, y2-y1, x2-x1)
			case RasConvInt3x:
				y3 = y1 + c.MulDiv(r.clipBox.X1-x1, y2-y1, x2-x1)
			case RasConvDbl:
				y3 = y1 + c.MulDiv(r.clipBox.X1-x1, y2-y1, x2-x1)
			case RasConvDbl3x:
				y3 = y1 + c.MulDiv(r.clipBox.X1-x1, y2-y1, x2-x1)
			}
			f3 = basics.ClippingFlagsY(y3, r.clipBox)
			r.lineClipY(ras, x1, y1, r.clipBox.X1, y3, f1, f3)
			r.lineClipY(ras, r.clipBox.X1, y3, r.clipBox.X1, y2, f3, f2)

		case 6: // x1 > clip.x2 && x2 < clip.x1
			switch c := any(conv).(type) {
			case RasConvInt:
				y3 = y1 + c.MulDiv(r.clipBox.X2-x1, y2-y1, x2-x1)
				y4 = y1 + c.MulDiv(r.clipBox.X1-x1, y2-y1, x2-x1)
			case RasConvIntSat:
				y3 = y1 + c.MulDiv(r.clipBox.X2-x1, y2-y1, x2-x1)
				y4 = y1 + c.MulDiv(r.clipBox.X1-x1, y2-y1, x2-x1)
			case RasConvInt3x:
				y3 = y1 + c.MulDiv(r.clipBox.X2-x1, y2-y1, x2-x1)
				y4 = y1 + c.MulDiv(r.clipBox.X1-x1, y2-y1, x2-x1)
			case RasConvDbl:
				y3 = y1 + c.MulDiv(r.clipBox.X2-x1, y2-y1, x2-x1)
				y4 = y1 + c.MulDiv(r.clipBox.X1-x1, y2-y1, x2-x1)
			case RasConvDbl3x:
				y3 = y1 + c.MulDiv(r.clipBox.X2-x1, y2-y1, x2-x1)
				y4 = y1 + c.MulDiv(r.clipBox.X1-x1, y2-y1, x2-x1)
			}
			f3 = basics.ClippingFlagsY(y3, r.clipBox)
			f4 = basics.ClippingFlagsY(y4, r.clipBox)
			r.lineClipY(ras, r.clipBox.X2, y1, r.clipBox.X2, y3, f1, f3)
			r.lineClipY(ras, r.clipBox.X2, y3, r.clipBox.X1, y4, f3, f4)
			r.lineClipY(ras, r.clipBox.X1, y4, r.clipBox.X1, y2, f4, f2)

		case 8: // x1 < clip.x1
			switch c := any(conv).(type) {
			case RasConvInt:
				y3 = y1 + c.MulDiv(r.clipBox.X1-x1, y2-y1, x2-x1)
			case RasConvIntSat:
				y3 = y1 + c.MulDiv(r.clipBox.X1-x1, y2-y1, x2-x1)
			case RasConvInt3x:
				y3 = y1 + c.MulDiv(r.clipBox.X1-x1, y2-y1, x2-x1)
			case RasConvDbl:
				y3 = y1 + c.MulDiv(r.clipBox.X1-x1, y2-y1, x2-x1)
			case RasConvDbl3x:
				y3 = y1 + c.MulDiv(r.clipBox.X1-x1, y2-y1, x2-x1)
			}
			f3 = basics.ClippingFlagsY(y3, r.clipBox)
			// TODO: Fix clipping for boundary-crossing lines
			// Test expectation: single line from intersection to end point
			// Current behavior: draws 2 lines (vertical boundary segment + clipped segment)
			// The first lineClipY call may be drawing an unwanted boundary line
			r.lineClipY(ras, r.clipBox.X1, y1, r.clipBox.X1, y3, f1, f3)
			r.lineClipY(ras, r.clipBox.X1, y3, x2, y2, f3, f2)

		case 9: // x1 < clip.x1 && x2 > clip.x2
			switch c := any(conv).(type) {
			case RasConvInt:
				y3 = y1 + c.MulDiv(r.clipBox.X1-x1, y2-y1, x2-x1)
				y4 = y1 + c.MulDiv(r.clipBox.X2-x1, y2-y1, x2-x1)
			case RasConvIntSat:
				y3 = y1 + c.MulDiv(r.clipBox.X1-x1, y2-y1, x2-x1)
				y4 = y1 + c.MulDiv(r.clipBox.X2-x1, y2-y1, x2-x1)
			case RasConvInt3x:
				y3 = y1 + c.MulDiv(r.clipBox.X1-x1, y2-y1, x2-x1)
				y4 = y1 + c.MulDiv(r.clipBox.X2-x1, y2-y1, x2-x1)
			case RasConvDbl:
				y3 = y1 + c.MulDiv(r.clipBox.X1-x1, y2-y1, x2-x1)
				y4 = y1 + c.MulDiv(r.clipBox.X2-x1, y2-y1, x2-x1)
			case RasConvDbl3x:
				y3 = y1 + c.MulDiv(r.clipBox.X1-x1, y2-y1, x2-x1)
				y4 = y1 + c.MulDiv(r.clipBox.X2-x1, y2-y1, x2-x1)
			}
			f3 = basics.ClippingFlagsY(y3, r.clipBox)
			f4 = basics.ClippingFlagsY(y4, r.clipBox)
			r.lineClipY(ras, r.clipBox.X1, y1, r.clipBox.X1, y3, f1, f3)
			r.lineClipY(ras, r.clipBox.X1, y3, r.clipBox.X2, y4, f3, f4)
			r.lineClipY(ras, r.clipBox.X2, y4, r.clipBox.X2, y2, f4, f2)

		case 12: // x1 < clip.x1 && x2 < clip.x1
			// TODO: Fix rasterizer clipping boundary detection logic
			// Test expectation: when both points are completely outside the same boundary
			// (e.g., both points left of clipBox.X1), no lines should be drawn.
			// Current behavior: draws a vertical line along the boundary.
			// Need to compare with AGG C++ implementation to determine correct behavior.
			r.lineClipY(ras, r.clipBox.X1, y1, r.clipBox.X1, y2, f1, f2)
		}
		r.f1 = f2
	} else {
		switch c := any(conv).(type) {
		case RasConvInt:
			ras.Line(c.Xi(r.x1), c.Yi(r.y1), c.Xi(x2), c.Yi(y2))
		case RasConvIntSat:
			ras.Line(c.Xi(r.x1), c.Yi(r.y1), c.Xi(x2), c.Yi(y2))
		case RasConvInt3x:
			ras.Line(c.Xi(r.x1), c.Yi(r.y1), c.Xi(x2), c.Yi(y2))
		case RasConvDbl:
			ras.Line(c.Xi(r.x1), c.Yi(r.y1), c.Xi(x2), c.Yi(y2))
		case RasConvDbl3x:
			ras.Line(c.Xi(r.x1), c.Yi(r.y1), c.Xi(x2), c.Yi(y2))
		}
	}
	r.x1 = x2
	r.y1 = y2
}

// RasterizerSlNoClip provides a no-clipping implementation.
// Equivalent to AGG's rasterizer_sl_no_clip class.
type RasterizerSlNoClip struct {
	x1, y1     float64
	rasterizer RasterizerInterface
}

// NewRasterizerSlNoClip creates a new no-clip rasterizer
func NewRasterizerSlNoClip(rasterizer RasterizerInterface) *RasterizerSlNoClip {
	return &RasterizerSlNoClip{
		x1:         0,
		y1:         0,
		rasterizer: rasterizer,
	}
}

// ResetClipping does nothing for no-clip implementation
func (r *RasterizerSlNoClip) ResetClipping() {
	// No-op
}

// ClipBox does nothing for no-clip implementation
func (r *RasterizerSlNoClip) ClipBox(x1, y1, x2, y2 float64) {
	// No-op
}

// MoveTo sets the current position
func (r *RasterizerSlNoClip) MoveTo(x1, y1 float64) {
	r.x1 = x1
	r.y1 = y1
}

// LineTo draws a line from the current position to (x2, y2)
func (r *RasterizerSlNoClip) LineTo(outline RasterizerInterface, x2, y2 float64) {
	// Apply coordinate conversion just like RasConvDbl.Xi() and Yi() do
	// This matches the C++ ras_conv_dbl::upscale() behavior
	x1i := basics.IRound(r.x1 * basics.PolySubpixelScale)
	y1i := basics.IRound(r.y1 * basics.PolySubpixelScale)
	x2i := basics.IRound(x2 * basics.PolySubpixelScale)
	y2i := basics.IRound(y2 * basics.PolySubpixelScale)
	outline.Line(x1i, y1i, x2i, y2i)
	r.x1 = x2
	r.y1 = y2
}

// Type aliases for convenience
type (
	RasterizerSlClipInt    = *RasterizerSlClip[RasConvInt]
	RasterizerSlClipIntSat = *RasterizerSlClip[RasConvIntSat]
	RasterizerSlClipInt3x  = *RasterizerSlClip[RasConvInt3x]
	RasterizerSlClipDbl    = *RasterizerSlClip[RasConvDbl]
	RasterizerSlClipDbl3x  = *RasterizerSlClip[RasConvDbl3x]
)
