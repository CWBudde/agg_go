// Package imagefilters2 ports AGG's image_filters2.cpp demo.
package imagefilters2

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	iagg2d "github.com/MeKo-Christian/agg_go/internal/agg2d"
	imgacc "github.com/MeKo-Christian/agg_go/internal/image"
)

const (
	baseWidth  = 500.0
	baseHeight = 340.0
)

type FilterSpec struct {
	Name           string
	VariableRadius bool
}

var filterSpecs = [...]FilterSpec{
	{Name: "simple (NN)"},
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

type State struct {
	Radius    float64
	FilterIdx int
	Normalize bool
}

func DefaultState() State {
	return State{
		Radius:    4.0,
		FilterIdx: 1,
		Normalize: true,
	}
}

func FilterSpecs() []FilterSpec {
	out := make([]FilterSpec, len(filterSpecs))
	copy(out, filterSpecs[:])
	return out
}

func (s *State) Clamp() {
	if s.Radius < 2.0 {
		s.Radius = 2.0
	}
	if s.Radius > 8.0 {
		s.Radius = 8.0
	}
	if s.FilterIdx < 0 {
		s.FilterIdx = 0
	}
	if s.FilterIdx >= len(filterSpecs) {
		s.FilterIdx = len(filterSpecs) - 1
	}
}

func (s State) UsesRadius() bool {
	return filterSpecs[s.FilterIdx].VariableRadius
}

func Draw(ctx *agg.Context, st State) {
	st.Clamp()
	ctx.Clear(agg.White)

	scale, offX, offY := fitFrame(ctx.GetImage().Width(), ctx.GetImage().Height())
	mapX := func(x float64) float64 { return offX + x*scale }
	mapY := func(y float64) float64 { return offY + y*scale }

	srcImg := createSourceImage()
	drawPreview(ctx, mapX, mapY, scale)
	drawDestinationImage(ctx, srcImg, st, mapX, mapY)
	drawGraph(ctx, st, mapX, mapY, scale)
	drawFrames(ctx, mapX, mapY, scale)
}

func drawPreview(ctx *agg.Context, mapX, mapY func(float64) float64, scale float64) {
	cell := 18.0
	x0 := 18.0
	y0 := 40.0
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			ctx.SetColor(sourceColors[y*4+x])
			ctx.FillRectangle(
				mapX(x0+float64(x)*cell),
				mapY(y0+float64(y)*cell),
				cell*scale,
				cell*scale,
			)
		}
	}

	ctx.SetColor(agg.RGBA(0, 0, 0, 0.2))
	ctx.SetLineWidth(max(0.75, scale))
	for i := 0; i <= 4; i++ {
		x := x0 + float64(i)*cell
		ctx.DrawLine(mapX(x), mapY(y0), mapX(x), mapY(y0+4*cell))
		y := y0 + float64(i)*cell
		ctx.DrawLine(mapX(x0), mapY(y), mapX(x0+4*cell), mapY(y))
	}
}

func drawDestinationImage(ctx *agg.Context, srcImg *agg.Image, st State, mapX, mapY func(float64) float64) {
	dst := ctx.GetImage().ToInternalImage()
	a := iagg2d.NewAgg2D()
	a.AttachImage(dst)

	if st.FilterIdx == 0 {
		a.ImageFilter(iagg2d.NoFilter)
	} else {
		a.SetImageFilterLUT(imgacc.NewImageFilterLUTWithFilter(newFilter(st.FilterIdx, st.Radius), st.Normalize))
	}

	par := []float64{
		mapX(200), mapY(40),
		mapX(500), mapY(40),
		mapX(500), mapY(340),
	}
	_ = a.TransformImageParallelogramSimple(srcImg.ToInternalImage(), par)
}

func drawFrames(ctx *agg.Context, mapX, mapY func(float64) float64, scale float64) {
	ctx.SetColor(agg.RGBA(0, 0, 0, 0.35))
	ctx.SetLineWidth(max(0.9, 1.2*scale))
	ctx.DrawRectangle(mapX(18), mapY(40), 72*scale, 72*scale)
	ctx.DrawRectangle(mapX(200), mapY(40), 300*scale, 300*scale)
}

func drawGraph(ctx *agg.Context, st State, mapX, mapY func(float64) float64, scale float64) {
	xStart := 5.0
	xEnd := 195.0
	yStart := 235.0
	yEnd := baseHeight - 5.0
	ys := yStart + (yEnd-yStart)/6.0

	for i := 0; i <= 16; i++ {
		x := xStart + (xEnd-xStart)*float64(i)/16.0
		alpha := 100.0 / 255.0
		if i == 8 {
			alpha = 1.0
		}
		strokeLine(ctx, max(0.75, 0.8*scale), agg.RGBA(0, 0, 0, alpha), mapX(x+0.5), mapY(yStart), mapX(x+0.5), mapY(yEnd))
	}
	strokeLine(ctx, max(0.75, 0.8*scale), agg.Black, mapX(xStart), mapY(ys), mapX(xEnd), mapY(ys))

	if st.FilterIdx == 0 {
		return
	}

	filter := newFilter(st.FilterIdx, st.Radius)
	lut := imgacc.NewImageFilterLUTWithFilter(filter, st.Normalize)
	weights := lut.WeightArray()

	radius := lut.Radius()
	n := int(radius * 256.0 * 2.0)
	if n < 2 {
		n = 2
	}
	dx := (xEnd - xStart) * radius / 8.0
	dy := yEnd - ys
	xs := (xEnd+xStart)/2.0 - (float64(lut.Diameter()) * (xEnd - xStart) / 32.0)
	nn := lut.Diameter() * 256

	pts := make([]point, 0, nn)
	pts = append(pts, point{
		x: mapX(xs + 0.5),
		y: mapY(ys + dy*float64(weights[0])/float64(imgacc.ImageFilterScale)),
	})
	for i := 1; i < nn; i++ {
		x := xs + dx*float64(i)/float64(n) + 0.5
		y := ys + dy*float64(weights[i])/float64(imgacc.ImageFilterScale)
		pts = append(pts, point{x: mapX(x), y: mapY(y)})
	}
	strokePolyline(ctx, max(0.75, 1.1*scale), agg.RGBA(0.39, 0.0, 0.0, 1.0), pts)
}

type point struct {
	x, y float64
}

func strokeLine(ctx *agg.Context, width float64, clr agg.Color, x1, y1, x2, y2 float64) {
	ctx.SetColor(clr)
	ctx.SetLineWidth(width)
	ctx.DrawLine(x1, y1, x2, y2)
}

func strokePolyline(ctx *agg.Context, width float64, clr agg.Color, pts []point) {
	if len(pts) < 2 {
		return
	}
	ctx.SetColor(clr)
	ctx.SetLineWidth(width)
	ctx.BeginPath()
	ctx.MoveTo(pts[0].x, pts[0].y)
	for i := 1; i < len(pts); i++ {
		ctx.LineTo(pts[i].x, pts[i].y)
	}
	ctx.Stroke()
}

func fitFrame(w, h int) (scale, offX, offY float64) {
	sx := float64(w) / baseWidth
	sy := float64(h) / baseHeight
	scale = math.Min(sx, sy)
	if scale > 1.0 {
		scale = 1.0
	}
	if scale <= 0 {
		scale = 1.0
	}
	offX = (float64(w) - baseWidth*scale) * 0.5
	offY = (float64(h) - baseHeight*scale) * 0.5
	return scale, offX, offY
}

var sourceColors = [...]agg.Color{
	agg.Green, agg.Red, agg.White, agg.Blue,
	agg.Blue, agg.Black, agg.White, agg.White,
	agg.White, agg.White, agg.Red, agg.Blue,
	agg.Red, agg.White, agg.Black, agg.Green,
}

func createSourceImage() *agg.Image {
	img := agg.CreateImage(4, 4)
	imgCtx := agg.NewContextForImage(img)
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			imgCtx.SetColor(sourceColors[y*4+x])
			imgCtx.FillRectangle(float64(x), float64(y), 1, 1)
		}
	}
	return img
}

type absFilter struct {
	base imgacc.FilterFunction
}

func (f absFilter) Radius() float64 {
	return f.base.Radius()
}

func (f absFilter) CalcWeight(x float64) float64 {
	return f.base.CalcWeight(math.Abs(x))
}

func newFilter(idx int, radius float64) imgacc.FilterFunction {
	switch idx {
	case 1:
		return absFilter{base: imgacc.BilinearFilter{}}
	case 2:
		return absFilter{base: imgacc.BicubicFilter{}}
	case 3:
		return absFilter{base: imgacc.Spline16Filter{}}
	case 4:
		return absFilter{base: imgacc.Spline36Filter{}}
	case 5:
		return absFilter{base: imgacc.HanningFilter{}}
	case 6:
		return absFilter{base: imgacc.HammingFilter{}}
	case 7:
		return absFilter{base: imgacc.HermiteFilter{}}
	case 8:
		return absFilter{base: imgacc.NewKaiserFilter(0)}
	case 9:
		return absFilter{base: imgacc.QuadricFilter{}}
	case 10:
		return absFilter{base: imgacc.CatromFilter{}}
	case 11:
		return absFilter{base: imgacc.GaussianFilter{}}
	case 12:
		return absFilter{base: imgacc.BesselFilter{}}
	case 13:
		return absFilter{base: imgacc.NewMitchellFilter(1.0/3.0, 1.0/3.0)}
	case 14:
		return absFilter{base: imgacc.NewSincFilter(radius)}
	case 15:
		return absFilter{base: imgacc.NewLanczosFilter(radius)}
	case 16:
		return absFilter{base: imgacc.NewBlackmanFilter(radius)}
	default:
		return absFilter{base: imgacc.BilinearFilter{}}
	}
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
