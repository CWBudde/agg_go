package color

import (
	"math"

	"github.com/MeKo-Christian/agg_go/internal/basics"
)

// ColorOrder identifies channel positions inside packed pixel layouts.
type ColorOrder struct {
	R, G, B, A int
}

// Common packed-channel orders used by pixel formats and row converters.
var (
	OrderRGB  = ColorOrder{R: 0, G: 1, B: 2, A: 3}
	OrderBGR  = ColorOrder{R: 2, G: 1, B: 0, A: 3}
	OrderRGBA = ColorOrder{R: 0, G: 1, B: 2, A: 3}
	OrderARGB = ColorOrder{R: 1, G: 2, B: 3, A: 0}
	OrderABGR = ColorOrder{R: 3, G: 2, B: 1, A: 0}
	OrderBGRA = ColorOrder{R: 2, G: 1, B: 0, A: 3}
)

// RGBA is the floating-point base color type used by AGG-style color math.
// Most fixed-width color families provide ConvertToRGBA/ConvertFromRGBA around
// this representation.
type RGBA struct {
	R, G, B, A float64
}

const (
	rgba8FloatScale  = 1.0 / 255.0
	rgba16FloatScale = 1.0 / 65535.0
)

// NewRGBA creates a floating-point RGBA color in normalized 0..1 units.
func NewRGBA(r, g, b, a float64) RGBA {
	return RGBA{R: r, G: g, B: b, A: a}
}

// NewRGBAFromRGBA8 expands 8-bit RGBA channels into normalized floats.
func NewRGBAFromRGBA8(r, g, b, a basics.Int8u) RGBA {
	return RGBA{
		R: float64(r) * rgba8FloatScale,
		G: float64(g) * rgba8FloatScale,
		B: float64(b) * rgba8FloatScale,
		A: float64(a) * rgba8FloatScale,
	}
}

// NewRGBAFromGray8 expands 8-bit grayscale plus alpha into normalized RGBA.
func NewRGBAFromGray8(v, a basics.Int8u) RGBA {
	return RGBA{
		R: float64(v) * rgba8FloatScale,
		G: float64(v) * rgba8FloatScale,
		B: float64(v) * rgba8FloatScale,
		A: float64(a) * rgba8FloatScale,
	}
}

// NewRGBAFromGray16 expands 16-bit grayscale plus alpha into normalized RGBA.
func NewRGBAFromGray16(v, a basics.Int16u) RGBA {
	return RGBA{
		R: float64(v) * rgba16FloatScale,
		G: float64(v) * rgba16FloatScale,
		B: float64(v) * rgba16FloatScale,
		A: float64(a) * rgba16FloatScale,
	}
}

// Clear resets the color to transparent black.
func (c *RGBA) Clear() *RGBA {
	c.R, c.G, c.B, c.A = 0, 0, 0, 0
	return c
}

// Transparent preserves RGB but clears alpha.
func (c *RGBA) Transparent() *RGBA {
	c.A = 0
	return c
}

// Opacity clamps and sets the alpha channel.
func (c *RGBA) Opacity(a float64) *RGBA {
	switch {
	case a < 0:
		c.A = 0
	case a > 1:
		c.A = 1
	default:
		c.A = a
	}
	return c
}

// GetOpacity returns the current alpha value.
func (c RGBA) GetOpacity() float64 {
	return c.A
}

// Premultiply converts the RGB channels to premultiplied-alpha form.
func (c *RGBA) Premultiply() *RGBA {
	if c.A <= 0 {
		c.R, c.G, c.B = 0, 0, 0
	} else if c.A < 1 {
		c.R *= c.A
		c.G *= c.A
		c.B *= c.A
	}
	return c
}

// PremultiplyAlpha premultiplies RGB using a caller-provided target alpha.
func (c *RGBA) PremultiplyAlpha(a float64) *RGBA {
	if c.A <= 0 || a <= 0 {
		c.R, c.G, c.B, c.A = 0, 0, 0, 0
	} else {
		a /= c.A
		c.R *= a
		c.G *= a
		c.B *= a
		c.A = a
	}
	return c
}

// Demultiply converts premultiplied RGB back to straight-alpha form.
func (c *RGBA) Demultiply() *RGBA {
	if c.A == 0 {
		c.R, c.G, c.B = 0, 0, 0
	} else if c.A > 0 && c.A < 1 {
		inv := 1.0 / c.A
		c.R *= inv
		c.G *= inv
		c.B *= inv
	}
	return c
}

// Gradient linearly interpolates between the receiver and c2.
func (c RGBA) Gradient(c2 RGBA, k float64) RGBA {
	return RGBA{
		R: c.R + k*(c2.R-c.R),
		G: c.G + k*(c2.G-c.G),
		B: c.B + k*(c2.B-c.B),
		A: c.A + k*(c2.A-c.A),
	}
}

// Add adds channel values component-wise.
func (c RGBA) Add(c2 RGBA) RGBA {
	return RGBA{
		R: c.R + c2.R,
		G: c.G + c2.G,
		B: c.B + c2.B,
		A: c.A + c2.A,
	}
}

// Subtract subtracts channel values component-wise.
func (c RGBA) Subtract(c2 RGBA) RGBA {
	return RGBA{
		R: c.R - c2.R,
		G: c.G - c2.G,
		B: c.B - c2.B,
		A: c.A - c2.A,
	}
}

// Multiply multiplies channel values component-wise.
func (c RGBA) Multiply(c2 RGBA) RGBA {
	return RGBA{
		R: c.R * c2.R,
		G: c.G * c2.G,
		B: c.B * c2.B,
		A: c.A * c2.A,
	}
}

// Scale multiplies all channels by k.
func (c RGBA) Scale(k float64) RGBA {
	return RGBA{
		R: c.R * k,
		G: c.G * k,
		B: c.B * k,
		A: c.A * k,
	}
}

// AddAssign performs in-place component-wise addition.
func (c *RGBA) AddAssign(c2 RGBA) *RGBA {
	c.R += c2.R
	c.G += c2.G
	c.B += c2.B
	c.A += c2.A
	return c
}

// MultiplyAssign scales the color in place.
func (c *RGBA) MultiplyAssign(k float64) *RGBA {
	c.R *= k
	c.G *= k
	c.B *= k
	c.A *= k
	return c
}

// NoColor returns transparent black.
func NoColor() RGBA {
	return RGBA{0, 0, 0, 0}
}

// FromWavelength reproduces AGG's visible-spectrum helper for wavelengths in
// roughly the 380-780nm range.
func FromWavelength(wl, gamma float64) RGBA {
	t := RGBA{0.0, 0.0, 0.0, 1.0}

	switch {
	case wl >= 380.0 && wl <= 440.0:
		t.R = -1.0 * (wl - 440.0) / (440.0 - 380.0)
		t.B = 1.0
	case wl >= 440.0 && wl <= 490.0:
		t.G = (wl - 440.0) / (490.0 - 440.0)
		t.B = 1.0
	case wl >= 490.0 && wl <= 510.0:
		t.G = 1.0
		t.B = -1.0 * (wl - 510.0) / (510.0 - 490.0)
	case wl >= 510.0 && wl <= 580.0:
		t.R = (wl - 510.0) / (580.0 - 510.0)
		t.G = 1.0
	case wl >= 580.0 && wl <= 645.0:
		t.R = 1.0
		t.G = -1.0 * (wl - 645.0) / (645.0 - 580.0)
	case wl >= 645.0 && wl <= 780.0:
		t.R = 1.0
	}

	s := 1.0
	if wl > 700.0 {
		s = 0.3 + 0.7*(780.0-wl)/(780.0-700.0)
	} else if wl < 420.0 {
		s = 0.3 + 0.7*(wl-380.0)/(420.0-380.0)
	}

	t.R = math.Pow(t.R*s, gamma)
	t.G = math.Pow(t.G*s, gamma)
	t.B = math.Pow(t.B*s, gamma)
	return t
}

// NewRGBAFromWavelength is the constructor-style alias for FromWavelength.
func NewRGBAFromWavelength(wl, gamma float64) RGBA {
	return FromWavelength(wl, gamma)
}

// RGBAPre creates a premultiplied floating-point RGBA color.
func RGBAPre(r, g, b, a float64) RGBA {
	c := NewRGBA(r, g, b, a)
	c.Premultiply()
	return c
}
