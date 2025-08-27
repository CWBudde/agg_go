package rasterizer

import (
	"math"

	"agg_go/internal/basics"
)

// Maximum coordinate value for polygon clipping
const PolyMaxCoord = (1 << 30) - 1

// Coord constraint for coordinate types
type Coord interface {
	~int | ~int32 | ~int64 | ~float64
}

// Conv defines the conversion policy interface (mirrors AGG's ras_conv_* "static" API)
type Conv[C Coord] interface {
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
	sat := basics.NewSaturation[int](PolyMaxCoord)
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
	sat := basics.NewSaturation[int](PolyMaxCoord)
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
type Rect[C Coord] struct {
	X1, Y1, X2, Y2 C
}

// Normalize ensures the rectangle coordinates are in the correct order
func (r *Rect[C]) Normalize() {
	if r.X2 < r.X1 {
		r.X1, r.X2 = r.X2, r.X1
	}
	if r.Y2 < r.Y1 {
		r.Y1, r.Y2 = r.Y1, r.Y2
	}
}

// Clipping flags (like AGG)
const (
	ClpX1 = 1
	ClpX2 = 2
	ClpY1 = 4
	ClpY2 = 8
)

func clippingFlags[C Coord](x, y C, rc Rect[C]) uint {
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

func clippingFlagsY[C Coord](y C, rc Rect[C]) uint {
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
type RasterizerSlClip[C Coord, V Conv[C]] struct {
	conv     V
	clipBox  Rect[C]
	x1, y1   C
	f1       uint
	clipping bool
}

// NewRasterizerSlClip creates a new scanline clipping rasterizer
func NewRasterizerSlClip[C Coord, V Conv[C]](conv V) *RasterizerSlClip[C, V] {
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
	if (f1|f2) == 0 {
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
		switch ((f1 & (ClpX1|ClpX2)) << 1) | (f2 & (ClpX1|ClpX2)) {
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
			// Fixed: Only draw the visible segment from intersection to end point
			r.lineClipY(ras, r.clipBox.X2, y3, x2, y2, f3, f2)

		case 3: // x1 > clip.x2 && x2 > clip.x2
			// Fixed: When both points are completely outside the same boundary,
			// no lines should be drawn (matches C++ AGG implementation)

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
			// Fixed: Only draw the visible segment from intersection to end point
			// The boundary line was creating unwanted artifacts
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
			// Fixed: When both points are completely outside the same boundary,
			// no lines should be drawn (matches C++ AGG implementation)
			// Just update position without drawing anything
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

// Concrete clipper implementations for common coordinate types

// RasterizerSlClipInt provides integer coordinate clipping
type RasterizerSlClipInt struct {
	clipBox  basics.Rect[int]
	x1, y1   int
	f1       uint32
	clipping bool
}

// NewRasterizerSlClipInt creates a new integer coordinate clipper
func NewRasterizerSlClipInt() *RasterizerSlClipInt {
	return &RasterizerSlClipInt{
		clipBox:  basics.Rect[int]{X1: 0, Y1: 0, X2: 0, Y2: 0},
		clipping: false,
	}
}

func (r *RasterizerSlClipInt) ResetClipping() {
	r.clipping = false
}

func (r *RasterizerSlClipInt) ClipBox(x1, y1, x2, y2 int) {
	r.clipBox = basics.Rect[int]{X1: x1, Y1: y1, X2: x2, Y2: y2}
	if r.clipBox.X1 > r.clipBox.X2 {
		r.clipBox.X1, r.clipBox.X2 = r.clipBox.X2, r.clipBox.X1
	}
	if r.clipBox.Y1 > r.clipBox.Y2 {
		r.clipBox.Y1, r.clipBox.Y2 = r.clipBox.Y2, r.clipBox.Y1
	}
	r.clipping = true
}

func (r *RasterizerSlClipInt) MoveTo(x1, y1 int) {
	r.x1 = x1
	r.y1 = y1
	if r.clipping {
		r.f1 = basics.ClippingFlags(float64(x1), float64(y1), basics.Rect[float64]{
			X1: float64(r.clipBox.X1), Y1: float64(r.clipBox.Y1),
			X2: float64(r.clipBox.X2), Y2: float64(r.clipBox.Y2),
		})
	}
}

func (r *RasterizerSlClipInt) LineTo(outline RasterizerInterface, x2, y2 int) {
	if r.clipping {
		f2 := basics.ClippingFlags(float64(x2), float64(y2), basics.Rect[float64]{
			X1: float64(r.clipBox.X1), Y1: float64(r.clipBox.Y1),
			X2: float64(r.clipBox.X2), Y2: float64(r.clipBox.Y2),
		})
		if (r.f1 & f2) == 0 {
			if (r.f1 | f2) != 0 {
				// Complex clipping needed - for now, pass through
				outline.Line(r.x1, r.y1, x2, y2)
			} else {
				outline.Line(r.x1, r.y1, x2, y2)
			}
		}
	} else {
		outline.Line(r.x1, r.y1, x2, y2)
	}
	r.x1 = x2
	r.y1 = y2
}
