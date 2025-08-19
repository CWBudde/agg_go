package span

import (
	"math"

	"agg_go/internal/basics"
)

// GrayCalc performs grayscale color interpolation calculations for one triangle edge.
type GrayCalc struct {
	x1    float64 // Start x coordinate
	y1    float64 // Start y coordinate
	dx    float64 // Delta x
	invDy float64 // 1/dy for fast division
	v1    int     // Start value (grayscale)
	a1    int     // Start alpha
	dv    int     // Delta value
	da    int     // Delta alpha
	v     int     // Current value
	a     int     // Current alpha
	x     int     // Current x (subpixel)
}

// GrayColor represents a grayscale color with alpha.
type GrayColor struct {
	V int // Value (grayscale intensity)
	A int // Alpha
}

// SpanGouraudGray implements grayscale Gouraud shading for triangles.
// This provides smooth grayscale interpolation optimized for single-channel output.
// It's equivalent to AGG's span_gouraud_gray template class.
type SpanGouraudGray struct {
	*SpanGouraud[GrayColor]          // Embed base Gouraud functionality
	swap                    bool     // Triangle orientation flag
	y2                      int      // Middle vertex Y coordinate
	c1                      GrayCalc // Edge interpolator 1
	c2                      GrayCalc // Edge interpolator 2
	c3                      GrayCalc // Edge interpolator 3
}

// NewSpanGouraudGray creates a new grayscale Gouraud span generator.
func NewSpanGouraudGray() *SpanGouraudGray {
	return &SpanGouraudGray{
		SpanGouraud: NewSpanGouraud[GrayColor](),
	}
}

// NewSpanGouraudGrayWithTriangle creates a new grayscale Gouraud span generator with initial triangle.
func NewSpanGouraudGrayWithTriangle(c1, c2, c3 GrayColor, x1, y1, x2, y2, x3, y3, d float64) *SpanGouraudGray {
	sg := NewSpanGouraudGray()
	sg.Colors(c1, c2, c3)
	sg.Triangle(x1, y1, x2, y2, x3, y3, d)
	return sg
}

// init initializes a grayscale edge interpolator.
func (gc *GrayCalc) init(c1, c2 CoordType[GrayColor]) {
	gc.x1 = c1.X - 0.5
	gc.y1 = c1.Y - 0.5
	gc.dx = c2.X - c1.X

	dy := c2.Y - c1.Y
	if math.Abs(dy) < 1e-10 {
		gc.invDy = 1e10
	} else {
		gc.invDy = 1.0 / dy
	}

	gc.v1 = c1.Color.V
	gc.a1 = c1.Color.A
	gc.dv = c2.Color.V - gc.v1
	gc.da = c2.Color.A - gc.a1
}

// calc calculates interpolated values for a given Y coordinate.
func (gc *GrayCalc) calc(y float64) {
	k := (y - gc.y1) * gc.invDy
	if k < 0.0 {
		k = 0.0
	}
	if k > 1.0 {
		k = 1.0
	}

	gc.v = gc.v1 + basics.IRound(float64(gc.dv)*k)
	gc.a = gc.a1 + basics.IRound(float64(gc.da)*k)
	gc.x = basics.IRound((gc.x1 + gc.dx*k) * SubpixelScale)
}

// Prepare prepares the span generator for rendering by setting up edge interpolators.
func (sg *SpanGouraudGray) Prepare() {
	coord := sg.ArrangeVertices()

	sg.y2 = int(coord[1].Y)

	// Determine triangle orientation
	sg.swap = basics.CrossProduct(coord[0].X, coord[0].Y,
		coord[2].X, coord[2].Y,
		coord[1].X, coord[1].Y) < 0.0

	// Initialize edge interpolators
	sg.c1.init(coord[0], coord[2])
	sg.c2.init(coord[0], coord[1])
	sg.c3.init(coord[1], coord[2])
}

// Generate generates a span of interpolated grayscale colors.
func (sg *SpanGouraudGray) Generate(span []GrayColor, x, y int, length uint) {
	sg.c1.calc(float64(y))
	pc1 := &sg.c1
	pc2 := &sg.c2

	if y < sg.y2 {
		// Bottom part of triangle (first subtriangle)
		sg.c2.calc(float64(y) + sg.c2.invDy)
	} else {
		// Upper part of triangle (second subtriangle)
		sg.c3.calc(float64(y) - sg.c3.invDy)
		pc2 = &sg.c3
	}

	if sg.swap {
		// Triangle is clockwise, swap the controlling structures
		pc1, pc2 = pc2, pc1
	}

	// Get horizontal length with subpixel accuracy
	nlen := basics.Abs(pc2.x - pc1.x)
	if nlen <= 0 {
		nlen = 1
	}

	// Create DDA interpolators for value and alpha
	v := NewGouraudDDAInterpolator(pc1.v, pc2.v, uint(nlen), 14)
	a := NewGouraudDDAInterpolator(pc1.a, pc2.a, uint(nlen), 14)

	// Calculate starting point with subpixel accuracy
	start := pc1.x - (x << SubpixelShift)
	v.Sub(uint(start))
	a.Sub(uint(start))
	nlen += start

	spanIdx := 0
	len := int(length)

	// Beginning part - check for overflow since we rolled back interpolators
	for len > 0 && start > 0 {
		vv := v.Y()
		va := a.Y()

		// Clamp values to valid range [0, 255]
		if vv < 0 {
			vv = 0
		} else if vv > 255 {
			vv = 255
		}
		if va < 0 {
			va = 0
		} else if va > 255 {
			va = 255
		}

		span[spanIdx] = GrayColor{V: vv, A: va}

		v.Add(SubpixelScale)
		a.Add(SubpixelScale)
		nlen -= SubpixelScale
		start -= SubpixelScale
		spanIdx++
		len--
	}

	// Middle part - no overflow checking needed
	for len > 0 && nlen > 0 {
		span[spanIdx] = GrayColor{
			V: v.Y(),
			A: a.Y(),
		}

		v.Add(SubpixelScale)
		a.Add(SubpixelScale)
		nlen -= SubpixelScale
		spanIdx++
		len--
	}

	// Ending part - check for overflow again
	for len > 0 {
		vv := v.Y()
		va := a.Y()

		// Clamp values to valid range [0, 255]
		if vv < 0 {
			vv = 0
		} else if vv > 255 {
			vv = 255
		}
		if va < 0 {
			va = 0
		} else if va > 255 {
			va = 255
		}

		span[spanIdx] = GrayColor{V: vv, A: va}

		v.Add(SubpixelScale)
		a.Add(SubpixelScale)
		spanIdx++
		len--
	}
}
