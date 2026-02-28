package rasterizer

import (
	"agg_go/internal/basics"
)

// Maximum coordinate value for polygon clipping
const PolyMaxCoord = (1 << 30) - 1

// Conv defines the conversion policy interface (mirrors AGG's ras_conv_* "static" API)
type Conv[C basics.CoordType] interface {
	// MulDiv returns round(a*b/c) for the coordinate type
	MulDiv(a, b, c float64) C
	Xi(v C) int
	Yi(v C) int
	Upscale(v float64) C
	Downscale(v int) C
}

// IntConv: internal coord is int, upscale = round(v * poly_subpixel_scale)
// Equivalent to AGG's ras_conv_int struct.
type IntConv struct{}

// MulDiv performs multiplication and division with rounding
func (IntConv) MulDiv(a, b, c float64) int {
	return basics.IRound(a * b / c)
}

// Xi converts input X coordinate (no transformation for integer converter)
func (IntConv) Xi(v int) int {
	return v
}

// Yi converts input Y coordinate (no transformation for integer converter)
func (IntConv) Yi(v int) int {
	return v
}

// Upscale converts double coordinate to subpixel integer coordinate
func (IntConv) Upscale(v float64) int {
	return basics.IRound(v * basics.PolySubpixelScale)
}

// Downscale converts subpixel integer coordinate back to integer coordinate
func (IntConv) Downscale(v int) int {
	return v / basics.PolySubpixelScale
}

// IntSatConv provides saturated integer coordinate conversion.
// Equivalent to AGG's ras_conv_int_sat struct.
type IntSatConv struct{}

// MulDiv performs multiplication and division with saturation
func (IntSatConv) MulDiv(a, b, c float64) int {
	sat := basics.NewSaturationInt(PolyMaxCoord)
	return sat.IRound(a * b / c)
}

// Xi converts input X coordinate (no transformation)
func (IntSatConv) Xi(v int) int {
	return v
}

// Yi converts input Y coordinate (no transformation)
func (IntSatConv) Yi(v int) int {
	return v
}

// Upscale converts double coordinate to subpixel integer with saturation
func (IntSatConv) Upscale(v float64) int {
	sat := basics.NewSaturationInt(PolyMaxCoord)
	return sat.IRound(v * basics.PolySubpixelScale)
}

// Downscale converts subpixel integer coordinate back to integer coordinate
func (IntSatConv) Downscale(v int) int {
	return v / basics.PolySubpixelScale
}

// Int3xConv provides 3x integer coordinate conversion for sub-pixel rendering.
// Equivalent to AGG's ras_conv_int_3x struct.
type Int3xConv struct{}

// MulDiv performs multiplication and division with rounding
func (Int3xConv) MulDiv(a, b, c float64) int {
	return basics.IRound(a * b / c)
}

// Xi converts input X coordinate with 3x scaling
func (Int3xConv) Xi(v int) int {
	return v * 3
}

// Yi converts input Y coordinate (no transformation)
func (Int3xConv) Yi(v int) int {
	return v
}

// Upscale converts double coordinate to subpixel integer coordinate
func (Int3xConv) Upscale(v float64) int {
	return basics.IRound(v * basics.PolySubpixelScale)
}

// Downscale converts subpixel integer coordinate back to integer coordinate
func (Int3xConv) Downscale(v int) int {
	return v / basics.PolySubpixelScale
}

// DblConv: internal coord is float64, upscale is identity
// Equivalent to AGG's ras_conv_dbl struct.
type DblConv struct{}

// MulDiv performs floating point multiplication and division
func (DblConv) MulDiv(a, b, c float64) float64 {
	return a * b / c
}

// Xi converts input X coordinate to subpixel integer
func (DblConv) Xi(v float64) int {
	return basics.IRound(v * basics.PolySubpixelScale)
}

// Yi converts input Y coordinate to subpixel integer
func (DblConv) Yi(v float64) int {
	return basics.IRound(v * basics.PolySubpixelScale)
}

// Upscale for double converter is pass-through
func (DblConv) Upscale(v float64) float64 {
	return v
}

// Downscale converts subpixel integer back to double coordinate
func (DblConv) Downscale(v int) float64 {
	return float64(v) / basics.PolySubpixelScale
}

// Dbl3xConv provides 3x double precision coordinate conversion.
// Equivalent to AGG's ras_conv_dbl_3x struct.
type Dbl3xConv struct{}

// MulDiv performs floating point multiplication and division
func (Dbl3xConv) MulDiv(a, b, c float64) float64 {
	return a * b / c
}

// Xi converts input X coordinate to subpixel integer with 3x scaling
func (Dbl3xConv) Xi(v float64) int {
	return basics.IRound(v * basics.PolySubpixelScale * 3)
}

// Yi converts input Y coordinate to subpixel integer
func (Dbl3xConv) Yi(v float64) int {
	return basics.IRound(v * basics.PolySubpixelScale)
}

// Upscale for double converter is pass-through
func (Dbl3xConv) Upscale(v float64) float64 {
	return v
}

// Downscale converts subpixel integer back to double coordinate
func (Dbl3xConv) Downscale(v int) float64 {
	return float64(v) / basics.PolySubpixelScale
}

// LineSink defines the interface implemented by the rasterizer/cell-sink
type LineSink interface {
	Line(x1, y1, x2, y2 int)
}

// Rect represents a clipping rectangle for generic coordinate types
type Rect[C basics.CoordType] struct {
	X1, Y1, X2, Y2 C
}

// Normalize ensures the rectangle coordinates are in the correct order
func (r *Rect[C]) Normalize() {
	if r.X2 < r.X1 {
		r.X1, r.X2 = r.X2, r.X1
	}
	if r.Y2 < r.Y1 {
	}
}

// Clipping flags (like AGG)
const (
	ClpX1 = 1
	ClpX2 = 2
	ClpY1 = 4
	ClpY2 = 8
)

func clippingFlags[C basics.CoordType](x, y C, rc Rect[C]) uint {
	var f uint
	if x < rc.X1 {
		f |= ClpX1
	} else if x > rc.X2 {
		f |= ClpX2
	}
	if y < rc.Y1 {
		f |= ClpY1
	} else if y > rc.Y2 {
		f |= ClpY2
	}
	return f
}

func clippingFlagsY[C basics.CoordType](y C, rc Rect[C]) uint {
	if y < rc.Y1 {
		return ClpY1
	}
	if y > rc.Y2 {
		return ClpY2
	}
	return 0
}

// RasterizerSlClip implements the scanline clipping rasterizer.
// Equivalent to AGG's rasterizer_sl_clip<Conv> template class.
type RasterizerSlClip[C basics.CoordType, V Conv[C]] struct {
	conv     V
	clipBox  Rect[C]
	x1, y1   C
	f1       uint
	clipping bool
}

// NewRasterizerSlClip creates a new scanline clipping rasterizer
func NewRasterizerSlClip[C basics.CoordType, V Conv[C]](conv V) *RasterizerSlClip[C, V] {
	return &RasterizerSlClip[C, V]{
		conv:     conv,
		clipping: false,
	}
}

// ResetClipping disables clipping
func (r *RasterizerSlClip[C, V]) ResetClipping() {
	r.clipping = false
}

// ClipBox sets the clipping rectangle and enables clipping
func (r *RasterizerSlClip[C, V]) ClipBox(x1, y1, x2, y2 C) {
	r.clipBox = Rect[C]{X1: x1, Y1: y1, X2: x2, Y2: y2}
	r.clipBox.Normalize()
	r.clipping = true
}

// MoveTo sets the current position
func (r *RasterizerSlClip[C, V]) MoveTo(x1, y1 C) {
	r.x1 = x1
	r.y1 = y1
	if r.clipping {
		r.f1 = clippingFlags(x1, y1, r.clipBox)
	}
}

// lineClipY implements Y-axis clipping for lines
func (r *RasterizerSlClip[C, V]) lineClipY(
	sink LineSink,
	x1, y1, x2, y2 C,
	f1, f2 uint,
) {
	f1 &= (ClpY1 | ClpY2)
	f2 &= (ClpY1 | ClpY2)
	if (f1 | f2) == 0 {
		// fully visible
		sink.Line(r.conv.Xi(x1), r.conv.Yi(y1), r.conv.Xi(x2), r.conv.Yi(y2))
		return
	}
	if f1 == f2 {
		// invisible by Y
		return
	}
	tx1, ty1 := x1, y1
	tx2, ty2 := x2, y2

	if (f1 & ClpY1) != 0 {
		tx1 = x1 + r.conv.MulDiv(float64(r.clipBox.Y1-y1), float64(x2-x1), float64(y2-y1))
		ty1 = r.clipBox.Y1
	}
	if (f1 & ClpY2) != 0 {
		tx1 = x1 + r.conv.MulDiv(float64(r.clipBox.Y2-y1), float64(x2-x1), float64(y2-y1))
		ty1 = r.clipBox.Y2
	}
	if (f2 & ClpY1) != 0 {
		tx2 = x1 + r.conv.MulDiv(float64(r.clipBox.Y1-y1), float64(x2-x1), float64(y2-y1))
		ty2 = r.clipBox.Y1
	}
	if (f2 & ClpY2) != 0 {
		tx2 = x1 + r.conv.MulDiv(float64(r.clipBox.Y2-y1), float64(x2-x1), float64(y2-y1))
		ty2 = r.clipBox.Y2
	}
	sink.Line(r.conv.Xi(tx1), r.conv.Yi(ty1), r.conv.Xi(tx2), r.conv.Yi(ty2))
}

// LineTo draws a line from the current position to (x2, y2) with clipping
func (r *RasterizerSlClip[C, V]) LineTo(sink LineSink, x2, y2 C) {
	if r.clipping {
		f2 := clippingFlags(x2, y2, r.clipBox)

		if (r.f1&(ClpY1|ClpY2)) == (f2&(ClpY1|ClpY2)) && (r.f1&(ClpY1|ClpY2)) != 0 {
			// invisible by Y
			r.x1, r.y1, r.f1 = x2, y2, f2
			return
		}
		x1, y1 := r.x1, r.y1
		f1 := r.f1

		// Handle X clipping cases
		switch ((f1 & (ClpX1 | ClpX2)) << 1) | (f2 & (ClpX1 | ClpX2)) {
		case 0:
			r.lineClipY(sink, x1, y1, x2, y2, f1, f2)
		case 1: // x2 > clip.x2
			y3 := y1 + r.conv.MulDiv(float64(r.clipBox.X2-x1), float64(y2-y1), float64(x2-x1))
			f3 := clippingFlagsY(y3, r.clipBox)
			r.lineClipY(sink, x1, y1, r.clipBox.X2, y3, f1, f3)
			r.lineClipY(sink, r.clipBox.X2, y3, r.clipBox.X2, y2, f3, f2)
		case 2: // x1 > clip.x2
			y3 := y1 + r.conv.MulDiv(float64(r.clipBox.X2-x1), float64(y2-y1), float64(x2-x1))
			f3 := clippingFlagsY(y3, r.clipBox)
			r.lineClipY(sink, r.clipBox.X2, y3, x2, y2, f3, f2)
		case 3: // both > x2
			r.lineClipY(sink, r.clipBox.X2, y1, r.clipBox.X2, y2, f1, f2)
		case 4: // x2 < clip.x1
			y3 := y1 + r.conv.MulDiv(float64(r.clipBox.X1-x1), float64(y2-y1), float64(x2-x1))
			f3 := clippingFlagsY(y3, r.clipBox)
			r.lineClipY(sink, x1, y1, r.clipBox.X1, y3, f1, f3)
			r.lineClipY(sink, r.clipBox.X1, y3, r.clipBox.X1, y2, f3, f2)
		case 6: // x1 > x2 && x2 < x1
			y3 := y1 + r.conv.MulDiv(float64(r.clipBox.X2-x1), float64(y2-y1), float64(x2-x1))
			y4 := y1 + r.conv.MulDiv(float64(r.clipBox.X1-x1), float64(y2-y1), float64(x2-x1))
			f3 := clippingFlagsY(y3, r.clipBox)
			f4 := clippingFlagsY(y4, r.clipBox)
			r.lineClipY(sink, r.clipBox.X2, y1, r.clipBox.X2, y3, f1, f3)
			r.lineClipY(sink, r.clipBox.X2, y3, r.clipBox.X1, y4, f3, f4)
			r.lineClipY(sink, r.clipBox.X1, y4, r.clipBox.X1, y2, f4, f2)
		case 8: // x1 < clip.x1
			y3 := y1 + r.conv.MulDiv(float64(r.clipBox.X1-x1), float64(y2-y1), float64(x2-x1))
			f3 := clippingFlagsY(y3, r.clipBox)
			r.lineClipY(sink, r.clipBox.X1, y1, r.clipBox.X1, y3, f1, f3)
			r.lineClipY(sink, r.clipBox.X1, y3, x2, y2, f3, f2)
		case 9: // x1 < x1 && x2 > x2
			y3 := y1 + r.conv.MulDiv(float64(r.clipBox.X1-x1), float64(y2-y1), float64(x2-x1))
			y4 := y1 + r.conv.MulDiv(float64(r.clipBox.X2-x1), float64(y2-y1), float64(x2-x1))
			f3 := clippingFlagsY(y3, r.clipBox)
			f4 := clippingFlagsY(y4, r.clipBox)
			r.lineClipY(sink, r.clipBox.X1, y1, r.clipBox.X1, y3, f1, f3)
			r.lineClipY(sink, r.clipBox.X1, y3, r.clipBox.X2, y4, f3, f4)
			r.lineClipY(sink, r.clipBox.X2, y4, r.clipBox.X2, y2, f4, f2)
		case 12: // both < x1
			r.lineClipY(sink, r.clipBox.X1, y1, r.clipBox.X1, y2, f1, f2)
		}
		r.f1 = f2
	} else {
		sink.Line(r.conv.Xi(r.x1), r.conv.Yi(r.y1), r.conv.Xi(x2), r.conv.Yi(y2))
	}
	r.x1 = x2
	r.y1 = y2
}

// RasterizerSlNoClip provides a no-clipping implementation.
// Equivalent to AGG's rasterizer_sl_no_clip class.
type RasterizerSlNoClip struct {
	x1, y1 int
}

// NewRasterizerSlNoClip creates a new no-clip rasterizer
func NewRasterizerSlNoClip() *RasterizerSlNoClip {
	return &RasterizerSlNoClip{}
}

// ResetClipping does nothing for no-clip implementation
func (r *RasterizerSlNoClip) ResetClipping() {}

// ClipBox does nothing for no-clip implementation
func (r *RasterizerSlNoClip) ClipBox(_, _, _, _ int) {}

// MoveTo sets the current position
func (r *RasterizerSlNoClip) MoveTo(x1, y1 int) {
	r.x1, r.y1 = x1, y1
}

// LineTo draws a line from the current position to (x2, y2)
func (r *RasterizerSlNoClip) LineTo(sink LineSink, x2, y2 int) {
	sink.Line(r.x1, r.y1, x2, y2)
	r.x1, r.y1 = x2, y2
}

// Type aliases for backward compatibility and convenience.
// These match the naming convention used in C++ AGG (ras_conv_*).
type (
	RasConvInt    = IntConv
	RasConvIntSat = IntSatConv
	RasConvInt3x  = Int3xConv
	RasConvDbl    = DblConv
	RasConvDbl3x  = Dbl3xConv
)
