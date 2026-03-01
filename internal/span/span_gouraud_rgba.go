package span

import (
	"math"

	"agg_go/internal/basics"
)

// SubpixelShift defines the subpixel precision shift (4 bits = 16 subpixels per pixel).
const SubpixelShift = 4

// SubpixelScale is the subpixel scale factor (1 << SubpixelShift).
const SubpixelScale = 1 << SubpixelShift

// RGBACalc performs RGBA color interpolation calculations for one triangle edge.
type RGBACalc struct {
	x1    float64 // Start x coordinate
	y1    float64 // Start y coordinate
	dx    float64 // Delta x
	invDy float64 // 1/dy for fast division
	r1    int     // Start red
	g1    int     // Start green
	b1    int     // Start blue
	a1    int     // Start alpha
	dr    int     // Delta red
	dg    int     // Delta green
	db    int     // Delta blue
	da    int     // Delta alpha
	r     int     // Current red
	g     int     // Current green
	b     int     // Current blue
	a     int     // Current alpha
	x     int     // Current x (subpixel)
}

// RGBAColor represents an RGBA color with integer components.
type RGBAColor struct {
	R, G, B, A int
}

// SpanGouraudRGBA implements RGBA Gouraud shading for triangles.
// This provides smooth color interpolation with subpixel accuracy.
// It's equivalent to AGG's span_gouraud_rgba template class.
type SpanGouraudRGBA struct {
	*SpanGouraud[RGBAColor]          // Embed base Gouraud functionality
	swap                    bool     // Triangle orientation flag
	y2                      int      // Middle vertex Y coordinate
	rgba1                   RGBACalc // Edge interpolator 1
	rgba2                   RGBACalc // Edge interpolator 2
	rgba3                   RGBACalc // Edge interpolator 3
}

// NewSpanGouraudRGBA creates a new RGBA Gouraud span generator.
func NewSpanGouraudRGBA() *SpanGouraudRGBA {
	return &SpanGouraudRGBA{
		SpanGouraud: NewSpanGouraud[RGBAColor](),
	}
}

// NewSpanGouraudRGBAWithTriangle creates a new RGBA Gouraud span generator with initial triangle.
func NewSpanGouraudRGBAWithTriangle(c1, c2, c3 RGBAColor, x1, y1, x2, y2, x3, y3, d float64) *SpanGouraudRGBA {
	sg := NewSpanGouraudRGBA()
	sg.Colors(c1, c2, c3)
	sg.Triangle(x1, y1, x2, y2, x3, y3, d)
	return sg
}

// init initializes an RGBA edge interpolator.
func (rc *RGBACalc) init(c1, c2 CoordType[RGBAColor]) {
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
func (rc *RGBACalc) calc(y float64) {
	k := (y - rc.y1) * rc.invDy
	if k < 0.0 {
		k = 0.0
	}
	if k > 1.0 {
		k = 1.0
	}

	rc.r = rc.r1 + basics.IRound(float64(rc.dr)*k)
	rc.g = rc.g1 + basics.IRound(float64(rc.dg)*k)
	rc.b = rc.b1 + basics.IRound(float64(rc.db)*k)
	rc.a = rc.a1 + basics.IRound(float64(rc.da)*k)
	rc.x = basics.IRound((rc.x1 + rc.dx*k) * SubpixelScale)
}

// Prepare prepares the span generator for rendering by setting up edge interpolators.
func (sg *SpanGouraudRGBA) Prepare() {
	coord := sg.ArrangeVertices()

	sg.y2 = int(coord[1].Y)

	// Determine triangle orientation
	sg.swap = basics.CrossProduct(coord[0].X, coord[0].Y,
		coord[2].X, coord[2].Y,
		coord[1].X, coord[1].Y) < 0.0

	// Initialize edge interpolators
	sg.rgba1.init(coord[0], coord[2])
	sg.rgba2.init(coord[0], coord[1])
	sg.rgba3.init(coord[1], coord[2])
}

// Generate generates a span of interpolated colors.
func (sg *SpanGouraudRGBA) Generate(span []RGBAColor, x, y int, length uint) {
	sg.rgba1.calc(float64(y))
	pc1 := &sg.rgba1
	pc2 := &sg.rgba2

	if y <= sg.y2 {
		// Bottom part of triangle (first subtriangle)
		sg.rgba2.calc(float64(y) + sg.rgba2.invDy)
	} else {
		// Upper part of triangle (second subtriangle)
		sg.rgba3.calc(float64(y) - sg.rgba3.invDy)
		pc2 = &sg.rgba3
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

	// Create DDA interpolators for each color component
	r := NewGouraudDDAInterpolator(pc1.r, pc2.r, uint(nlen), 14)
	g := NewGouraudDDAInterpolator(pc1.g, pc2.g, uint(nlen), 14)
	b := NewGouraudDDAInterpolator(pc1.b, pc2.b, uint(nlen), 14)
	a := NewGouraudDDAInterpolator(pc1.a, pc2.a, uint(nlen), 14)

	// Calculate starting point with subpixel accuracy
	start := pc1.x - (x << SubpixelShift)
	
	// Safety: prevent massive underflow if start is negative or too small
	if start < 0 {
		// If start is negative, it means we're beginning before the span's x
		// We'll skip the "beginning part" logic
		start = 0
	}
	
	r.Sub(uint(start))
	g.Sub(uint(start))
	b.Sub(uint(start))
	a.Sub(uint(start))
	nlen += start

	spanIdx := 0
	len := int(length)

	// Beginning part - check for overflow since we rolled back interpolators
	// Added hard limit to prevent infinite loops if start is corrupted
	for len > 0 && start > 0 && spanIdx < int(length) {
		vr := r.Y()
		vg := g.Y()
		vb := b.Y()
		va := a.Y()

		// Clamp values to valid range [0, 255]
		if vr < 0 {
			vr = 0
		} else if vr > 255 {
			vr = 255
		}
		if vg < 0 {
			vg = 0
		} else if vg > 255 {
			vg = 255
		}
		if vb < 0 {
			vb = 0
		} else if vb > 255 {
			vb = 255
		}
		if va < 0 {
			va = 0
		} else if va > 255 {
			va = 255
		}

		span[spanIdx] = RGBAColor{R: vr, G: vg, B: vb, A: va}

		r.Add(SubpixelScale)
		g.Add(SubpixelScale)
		b.Add(SubpixelScale)
		a.Add(SubpixelScale)
		nlen -= SubpixelScale
		start -= SubpixelScale
		spanIdx++
		len--
	}

	// Middle part - no overflow checking needed
	for len > 0 && nlen > 0 && spanIdx < int(length) {
		span[spanIdx] = RGBAColor{
			R: r.Y(),
			G: g.Y(),
			B: b.Y(),
			A: a.Y(),
		}

		r.Add(SubpixelScale)
		g.Add(SubpixelScale)
		b.Add(SubpixelScale)
		a.Add(SubpixelScale)
		nlen -= SubpixelScale
		spanIdx++
		len--
	}

	// Ending part - check for overflow again
	for len > 0 && spanIdx < int(length) {
		vr := r.Y()
		vg := g.Y()
		vb := b.Y()
		va := a.Y()

		// Clamp values to valid range [0, 255]
		if vr < 0 {
			vr = 0
		} else if vr > 255 {
			vr = 255
		}
		if vg < 0 {
			vg = 0
		} else if vg > 255 {
			vg = 255
		}
		if vb < 0 {
			vb = 0
		} else if vb > 255 {
			vb = 255
		}
		if va < 0 {
			va = 0
		} else if va > 255 {
			va = 255
		}

		span[spanIdx] = RGBAColor{R: vr, G: vg, B: vb, A: va}

		r.Add(SubpixelScale)
		g.Add(SubpixelScale)
		b.Add(SubpixelScale)
		a.Add(SubpixelScale)
		spanIdx++
		len--
	}
}
