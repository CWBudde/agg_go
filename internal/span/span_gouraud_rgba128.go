package span

import (
	"math"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/color"
)

// SpanGouraudRGBA128 is the float (128-bit, 4 x float32) twin of SpanGouraudRGBA.
// It implements AGG's span_gouraud_rgba algorithm for the float rgba color type:
// the triangle geometry and the subpixel horizontal position handling are
// identical to the 8-bit generator (precision-independent), while the color
// components are interpolated in straight float space ([0,1]) rather than as
// integer 0-255 components.
//
// Colors are clamped to [0,1] in the begin/end (anti-aliasing) portions of each
// scanline, mirroring the integer generator's clampRGBAComponent behavior.
type SpanGouraudRGBA128 struct {
	*SpanGouraud[color.RGBA32[color.Linear]]      // Embed color-agnostic base
	swap                                     bool // Triangle orientation flag
	y2                                       int  // Middle vertex Y coordinate
	rgba1                                    rgba128Calc
	rgba2                                    rgba128Calc
	rgba3                                    rgba128Calc
}

// rgba128Calc performs float RGBA color interpolation for one triangle edge.
// The x position is interpolated in integer subpixels (matching the 8-bit edge
// calc); the color components are interpolated as floats.
type rgba128Calc struct {
	x1    float64 // Start x coordinate (biased by -0.5)
	y1    float64 // Start y coordinate (biased by -0.5)
	dx    float64 // Delta x
	invDy float64 // 1/dy for fast division
	r1    float32 // Start red
	g1    float32 // Start green
	b1    float32 // Start blue
	a1    float32 // Start alpha
	dr    float32 // Delta red
	dg    float32 // Delta green
	db    float32 // Delta blue
	da    float32 // Delta alpha
	r     float32 // Current red
	g     float32 // Current green
	b     float32 // Current blue
	a     float32 // Current alpha
	x     int     // Current x (subpixel)
}

// NewSpanGouraudRGBA128 creates a new float RGBA Gouraud span generator.
func NewSpanGouraudRGBA128() *SpanGouraudRGBA128 {
	return &SpanGouraudRGBA128{
		SpanGouraud: NewSpanGouraud[color.RGBA32[color.Linear]](),
	}
}

// NewSpanGouraudRGBA128WithTriangle creates a float RGBA Gouraud span generator
// with an initial triangle.
func NewSpanGouraudRGBA128WithTriangle(c1, c2, c3 color.RGBA32[color.Linear], x1, y1, x2, y2, x3, y3, d float64) *SpanGouraudRGBA128 {
	sg := NewSpanGouraudRGBA128()
	sg.Colors(c1, c2, c3)
	sg.Triangle(x1, y1, x2, y2, x3, y3, d)
	return sg
}

func clampUnit128(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// init initializes a float RGBA edge interpolator.
func (rc *rgba128Calc) init(c1, c2 CoordType[color.RGBA32[color.Linear]]) {
	rc.x1 = c1.X - 0.5
	rc.y1 = c1.Y - 0.5
	rc.dx = c2.X - c1.X

	dy := c2.Y - c1.Y
	if math.Abs(dy) < 1e-5 {
		rc.invDy = 1e5
	} else {
		rc.invDy = 1.0 / dy
	}

	rc.r1 = c1.Color.R
	rc.g1 = c1.Color.G
	rc.b1 = c1.Color.B
	rc.a1 = c1.Color.A

	rc.dr = c2.Color.R - rc.r1
	rc.dg = c2.Color.G - rc.g1
	rc.db = c2.Color.B - rc.b1
	rc.da = c2.Color.A - rc.a1
}

// calc calculates interpolated values for a given Y coordinate.
func (rc *rgba128Calc) calc(y float64) {
	k := (y - rc.y1) * rc.invDy
	if k < 0.0 {
		k = 0.0
	}
	if k > 1.0 {
		k = 1.0
	}
	kf := float32(k)

	rc.r = rc.r1 + rc.dr*kf
	rc.g = rc.g1 + rc.dg*kf
	rc.b = rc.b1 + rc.db*kf
	rc.a = rc.a1 + rc.da*kf
	rc.x = basics.IRound((rc.x1 + rc.dx*k) * SubpixelScale)
}

// Prepare sets up the edge interpolators. Mirrors SpanGouraudRGBA.Prepare.
func (sg *SpanGouraudRGBA128) Prepare() {
	coord := sg.ArrangeVertices()

	sg.y2 = int(coord[1].Y)

	sg.swap = basics.CrossProduct(coord[0].X, coord[0].Y,
		coord[2].X, coord[2].Y,
		coord[1].X, coord[1].Y) < 0.0

	sg.rgba1.init(coord[0], coord[2])
	sg.rgba2.init(coord[0], coord[1])
	sg.rgba3.init(coord[1], coord[2])
}

// Generate generates a span of interpolated float colors. Mirrors
// SpanGouraudRGBA.Generate, interpolating color components linearly in float
// space across the subpixel horizontal extent.
func (sg *SpanGouraudRGBA128) Generate(spanColors []color.RGBA32[color.Linear], x, y int, length uint) {
	sg.rgba1.calc(float64(y))
	pc1 := &sg.rgba1
	pc2 := &sg.rgba2

	if y <= sg.y2 {
		sg.rgba2.calc(float64(y) + sg.rgba2.invDy)
	} else {
		sg.rgba3.calc(float64(y) - sg.rgba3.invDy)
		pc2 = &sg.rgba3
	}

	if sg.swap {
		pc1, pc2 = pc2, pc1
	}

	nlen := basics.Abs(pc2.x - pc1.x)
	if nlen <= 0 {
		nlen = 1
	}

	// Float linear color interpolation across nlen subpixel units. step is the
	// per-subpixel-unit increment; SubpixelScale units advance one pixel.
	invNlen := 1.0 / float32(nlen)
	rStep := (pc2.r - pc1.r) * invNlen
	gStep := (pc2.g - pc1.g) * invNlen
	bStep := (pc2.b - pc1.b) * invNlen
	aStep := (pc2.a - pc1.a) * invNlen

	r, g, b, a := pc1.r, pc1.g, pc1.b, pc1.a

	// Subpixel start offset relative to the requested pixel x. Roll the
	// interpolators forward/back so the first emitted pixel samples the correct
	// position, matching the integer generator's start handling.
	start := pc1.x - (x << SubpixelShift)
	if start >= 0 {
		// Interpolators are to the right of x: roll them back.
		off := float32(start)
		r -= rStep * off
		g -= gStep * off
		b -= bStep * off
		a -= aStep * off
	} else {
		off := float32(-start)
		r += rStep * off
		g += gStep * off
		b += bStep * off
		a += aStep * off
	}
	nlen += start

	pixStepR := rStep * SubpixelScale
	pixStepG := gStep * SubpixelScale
	pixStepB := bStep * SubpixelScale
	pixStepA := aStep * SubpixelScale

	i := 0
	lim := int(length)

	// Beginning part (rolled-back interpolators may exceed [0,1]) while start>0.
	for lim > 0 && start > 0 {
		spanColors[i] = color.RGBA32[color.Linear]{
			R: clampUnit128(r), G: clampUnit128(g), B: clampUnit128(b), A: clampUnit128(a),
		}
		r += pixStepR
		g += pixStepG
		b += pixStepB
		a += pixStepA
		nlen -= SubpixelScale
		start -= SubpixelScale
		i++
		lim--
	}

	// Middle part: no clamping needed while nlen > 0.
	for lim > 0 && nlen > 0 {
		spanColors[i] = color.RGBA32[color.Linear]{R: r, G: g, B: b, A: a}
		r += pixStepR
		g += pixStepG
		b += pixStepB
		a += pixStepA
		nlen -= SubpixelScale
		i++
		lim--
	}

	// Ending part: interpolators may overflow again (anti-aliasing pixels).
	for lim > 0 {
		spanColors[i] = color.RGBA32[color.Linear]{
			R: clampUnit128(r), G: clampUnit128(g), B: clampUnit128(b), A: clampUnit128(a),
		}
		r += pixStepR
		g += pixStepG
		b += pixStepB
		a += pixStepA
		i++
		lim--
	}
}
