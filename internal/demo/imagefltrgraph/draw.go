// Package imagefltrgraph ports AGG's image_fltr_graph.cpp demo.
//
// It draws the raw filter function (red), an unnormalized discrete sum (green),
// and normalized LUT weights (blue) over the same axis.
package imagefltrgraph

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/image"
)

const (
	baseWidth  = 780.0
	baseHeight = 300.0
	numFilters = 16
)

type FilterSpec struct {
	Name           string
	VariableRadius bool
}

var filterSpecs = [numFilters]FilterSpec{
	{Name: "bilinear"},
	{Name: "bicubic"},
	{Name: "spline16"},
	{Name: "spline36"},
	{Name: "hanning"},
	{Name: "hamming"},
	{Name: "hermite"},
	{Name: "kaiser"},
	{Name: "quadric"},
	{Name: "catrom"},
	{Name: "gaussian"},
	{Name: "bessel"},
	{Name: "mitchell"},
	{Name: "sinc", VariableRadius: true},
	{Name: "lanczos", VariableRadius: true},
	{Name: "blackman", VariableRadius: true},
}

func FilterSpecs() []FilterSpec {
	out := make([]FilterSpec, 0, len(filterSpecs))
	out = append(out, filterSpecs[:]...)
	return out
}

type State struct {
	Radius  float64
	Enabled [numFilters]bool
}

func DefaultState() State {
	s := State{Radius: 4.0}
	s.Enabled[0] = true
	return s
}

func (s *State) Clamp() {
	if s.Radius < 2.0 {
		s.Radius = 2.0
	}
	if s.Radius > 8.0 {
		s.Radius = 8.0
	}
}

func (s *State) SetMask(mask uint32) {
	for i := range s.Enabled {
		s.Enabled[i] = ((mask >> i) & 1) != 0
	}
}

func (s State) Mask() uint32 {
	var mask uint32
	for i, on := range s.Enabled {
		if on {
			mask |= 1 << i
		}
	}
	return mask
}

func Draw(ctx *agg.Context, st State) {
	st.Clamp()
	ctx.Clear(agg.White)

	scale, offX, offY := fitFrame(ctx.GetImage().Width(), ctx.GetImage().Height())
	mapX := func(x float64) float64 { return offX + x*scale }
	mapY := func(y float64) float64 { return offY + y*scale }

	xStart := 125.0
	xEnd := baseWidth - 15.0
	yStart := 10.0
	yEnd := baseHeight - 10.0
	xCenter := (xStart + xEnd) * 0.5
	ys := yStart + (yEnd-yStart)/6.0

	// Vertical reference grid + center axis.
	for i := 0; i <= 16; i++ {
		x := xStart + (xEnd-xStart)*float64(i)/16.0
		alpha := 100.0 / 255.0
		if i == 8 {
			alpha = 1.0
		}
		strokeLine(ctx, 1.0*scale, agg.RGBA(0, 0, 0, alpha), mapX(x+0.5), mapY(yStart), mapX(x+0.5), mapY(yEnd))
	}
	strokeLine(ctx, 1.0*scale, agg.Black, mapX(xStart), mapY(ys), mapX(xEnd), mapY(ys))

	for i := 0; i < numFilters; i++ {
		if !st.Enabled[i] {
			continue
		}

		filter := newFilter(i, st.Radius)
		radius := filter.Radius()
		dy := yEnd - ys
		dx := (xEnd - xStart) * radius / 8.0
		n := int(radius * 256.0 * 2.0)
		if n < 2 {
			n = 2
		}
		xs := (xEnd+xStart)/2.0 - (radius * (xEnd - xStart) / 16.0)

		// Raw continuous filter function (red).
		rawPts := make([]point, 0, n)
		rawPts = append(rawPts, point{
			x: mapX(xs + 0.5),
			y: mapY(ys + dy*filter.CalcWeight(-radius)),
		})
		for j := 1; j < n; j++ {
			x := xs + dx*float64(j)/float64(n) + 0.5
			y := ys + dy*filter.CalcWeight(float64(j)/256.0-radius)
			rawPts = append(rawPts, point{x: mapX(x), y: mapY(y)})
		}
		strokePolyline(ctx, 1.5*scale, agg.RGBA(0.5, 0, 0, 1), rawPts)

		// Unnormalized discrete sum response (green), preserved from C++ logic.
		// Note: the `||` condition is intentionally kept for parity with upstream.
		ir := int(math.Ceil(radius) + 0.1)
		sumPts := make([]point, 0, 256)
		for xint := 0; xint < 256; xint++ {
			sum := 0.0
			for xfract := -ir; xfract < ir; xfract++ {
				xf := float64(xint)/256.0 + float64(xfract)
				if xf >= -radius || xf <= radius {
					sum += filter.CalcWeight(xf)
				}
			}
			x := xCenter + ((-128.0+float64(xint))/128.0)*radius*(xEnd-xStart)/16.0
			y := ys + sum*256.0 - 256.0
			sumPts = append(sumPts, point{x: mapX(x), y: mapY(y)})
		}
		strokePolyline(ctx, 1.5*scale, agg.RGBA(0, 0.5, 0, 1), sumPts)

		// Normalized LUT weights (blue).
		lut := image.NewImageFilterLUTWithFilter(filter, true)
		weights := lut.WeightArray()
		xsLUT := (xEnd+xStart)/2.0 - (float64(lut.Diameter()) * (xEnd - xStart) / 32.0)
		nn := lut.Diameter() * 256
		nBase := float64(n)
		lutPts := make([]point, 0, nn)
		for j := 0; j < nn; j++ {
			x := xsLUT + dx*float64(j)/nBase + 0.5
			y := ys + dy*float64(weights[j])/float64(image.ImageFilterScale)
			lutPts = append(lutPts, point{x: mapX(x), y: mapY(y)})
		}
		strokePolyline(ctx, 1.5*scale, agg.RGBA(0, 0, 0.5, 1), lutPts)
	}
}

type point struct {
	x, y float64
}

func strokeLine(ctx *agg.Context, width float64, clr agg.Color, x1, y1, x2, y2 float64) {
	if width < 0.75 {
		width = 0.75
	}
	ctx.SetColor(clr)
	ctx.SetLineWidth(width)
	ctx.DrawLine(x1, y1, x2, y2)
}

func strokePolyline(ctx *agg.Context, width float64, clr agg.Color, pts []point) {
	if len(pts) < 2 {
		return
	}
	if width < 0.75 {
		width = 0.75
	}
	ctx.SetColor(clr)
	ctx.SetLineWidth(width)
	a := ctx.GetAgg2D()
	a.ResetPath()
	a.MoveTo(pts[0].x, pts[0].y)
	for i := 1; i < len(pts); i++ {
		a.LineTo(pts[i].x, pts[i].y)
	}
	a.DrawPath(agg.StrokeOnly)
}

func fitFrame(w, h int) (scale, offX, offY float64) {
	fw := float64(w)
	fh := float64(h)
	sx := fw / baseWidth
	sy := fh / baseHeight
	scale = math.Min(sx, sy)
	if scale > 1.0 {
		scale = 1.0
	}
	if scale <= 0 {
		scale = 1.0
	}
	offX = (fw - baseWidth*scale) * 0.5
	offY = (fh - baseHeight*scale) * 0.5
	return scale, offX, offY
}

type filterEval interface {
	image.FilterFunction
	SetRadius(r float64)
}

type constFilter struct {
	fn image.FilterFunction
}

func (f *constFilter) Radius() float64 { return f.fn.Radius() }

func (f *constFilter) SetRadius(_ float64) {}

func (f *constFilter) CalcWeight(x float64) float64 {
	return f.fn.CalcWeight(math.Abs(x))
}

type varFilter struct {
	radius  float64
	factory func(r float64) image.FilterFunction
	fn      image.FilterFunction
}

func newVarFilter(factory func(r float64) image.FilterFunction) *varFilter {
	f := &varFilter{factory: factory}
	f.SetRadius(2.0)
	return f
}

func (f *varFilter) Radius() float64 { return f.fn.Radius() }

func (f *varFilter) SetRadius(r float64) {
	f.radius = r
	f.fn = f.factory(r)
}

func (f *varFilter) CalcWeight(x float64) float64 {
	return f.fn.CalcWeight(math.Abs(x))
}

func newFilter(index int, radius float64) filterEval {
	switch index {
	case 0:
		return &constFilter{fn: image.BilinearFilter{}}
	case 1:
		return &constFilter{fn: image.BicubicFilter{}}
	case 2:
		return &constFilter{fn: image.Spline16Filter{}}
	case 3:
		return &constFilter{fn: image.Spline36Filter{}}
	case 4:
		return &constFilter{fn: image.HanningFilter{}}
	case 5:
		return &constFilter{fn: image.HammingFilter{}}
	case 6:
		return &constFilter{fn: image.HermiteFilter{}}
	case 7:
		return &constFilter{fn: image.NewKaiserFilter(6.33)}
	case 8:
		return &constFilter{fn: image.QuadricFilter{}}
	case 9:
		return &constFilter{fn: image.CatromFilter{}}
	case 10:
		return &constFilter{fn: image.GaussianFilter{}}
	case 11:
		return &constFilter{fn: image.BesselFilter{}}
	case 12:
		return &constFilter{fn: image.NewMitchellFilter(1.0/3.0, 1.0/3.0)}
	case 13:
		f := newVarFilter(func(r float64) image.FilterFunction { return image.NewSincFilter(r) })
		f.SetRadius(radius)
		return f
	case 14:
		f := newVarFilter(func(r float64) image.FilterFunction { return image.NewLanczosFilter(r) })
		f.SetRadius(radius)
		return f
	case 15:
		f := newVarFilter(func(r float64) image.FilterFunction { return image.NewBlackmanFilter(r) })
		f.SetRadius(radius)
		return f
	default:
		return &constFilter{fn: image.BilinearFilter{}}
	}
}
