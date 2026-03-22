// Port of AGG C++ image_filters2.cpp.
//
// The standalone example renders the full original composition:
// - filter response graph in the upper-left
// - radio buttons and checkbox in the lower-left
// - transformed 4x4 source image on the right
//
// The C++ demo runs with flip_y=true, so the final image is vertically flipped
// before saving to match the original window orientation.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	iagg2d "github.com/MeKo-Christian/agg_go/internal/agg2d"
	icol "github.com/MeKo-Christian/agg_go/internal/color"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/checkbox"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/rbox"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	imgacc "github.com/MeKo-Christian/agg_go/internal/image"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
)

const (
	frameWidth  = 500
	frameHeight = 340
)

type demo struct {
	state imageFilters2State
}

type imageFilters2State struct {
	filterIdx int
	radius    float64
	normalize bool
}

func defaultState() imageFilters2State {
	return imageFilters2State{
		filterIdx: 1,
		radius:    4.0,
		normalize: true,
	}
}

func (s *imageFilters2State) clamp() {
	if s.radius < 2.0 {
		s.radius = 2.0
	}
	if s.radius > 8.0 {
		s.radius = 8.0
	}
	if s.filterIdx < 0 {
		s.filterIdx = 0
	}
	if s.filterIdx > 16 {
		s.filterIdx = 16
	}
}

func newDemo() *demo {
	return &demo{state: defaultState()}
}

func (d *demo) Render(img *agg.Image) {
	d.state.clamp()

	ctx := agg.NewContextForImage(img)
	ctx.Clear(agg.White)

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	srcImg := createSourceImage()
	drawDestinationImage(ctx.GetImage(), srcImg, d.state)
	drawGraph(ctx, d.state)
	drawControls(ctx, d.state)

	flipImageY(img)
}

func drawDestinationImage(dstImg, srcImg *agg.Image, st imageFilters2State) {
	internal := iagg2d.NewAgg2D()
	internal.AttachImage(dstImg.ToInternalImage())
	switch st.filterIdx {
	case 0:
		internal.ImageFilter(iagg2d.NoFilter)
	case 1:
		internal.ImageFilter(iagg2d.Bilinear)
	case 2:
		internal.ImageFilter(iagg2d.Bicubic)
	case 3:
		internal.ImageFilter(iagg2d.Spline16)
	case 4:
		internal.ImageFilter(iagg2d.Spline36)
	case 5:
		internal.ImageFilter(iagg2d.Hanning)
	case 6:
		internal.ImageFilter(iagg2d.Hamming)
	case 7:
		internal.ImageFilter(iagg2d.Hermite)
	case 8:
		internal.ImageFilter(iagg2d.Kaiser)
	case 9:
		internal.ImageFilter(iagg2d.Quadric)
	case 10:
		internal.ImageFilter(iagg2d.Catrom)
	case 11:
		internal.ImageFilter(iagg2d.Gaussian)
	case 12:
		internal.ImageFilter(iagg2d.Bessel)
	case 13:
		internal.ImageFilter(iagg2d.Mitchell)
	case 14:
		internal.SetImageFilterRadius(iagg2d.Sinc, st.radius)
	case 15:
		internal.SetImageFilterRadius(iagg2d.Lanczos, st.radius)
	case 16:
		internal.SetImageFilterRadius(iagg2d.Blackman, st.radius)
	default:
		internal.ImageFilter(iagg2d.Bilinear)
	}
	par := []float64{
		200.5, 40.5,
		500.5, 40.5,
		500.5, 340.5,
	}
	_ = internal.TransformImageParallelogramSimple(srcImg.ToInternalImage(), par)
}

func drawGraph(ctx *agg.Context, st imageFilters2State) {
	xStart := 5.0
	xEnd := 195.0
	yStart := 235.0
	yEnd := float64(frameHeight) - 5.0
	ys := yStart + (yEnd-yStart)/6.0

	for i := 0; i <= 16; i++ {
		x := xStart + (xEnd-xStart)*float64(i)/16.0
		alpha := 100.0 / 255.0
		if i == 8 {
			alpha = 1.0
		}
		strokeLine(ctx, max(0.75, 0.8), agg.RGBA(0, 0, 0, alpha), x+0.5, yStart, x+0.5, yEnd)
	}
	strokeLine(ctx, max(0.75, 0.8), agg.Black, xStart, ys, xEnd, ys)

	if st.filterIdx == 0 {
		return
	}

	filter := newFilter(st.filterIdx, st.radius)
	lut := imgacc.NewImageFilterLUTWithFilter(filter, st.normalize)
	weights := lut.WeightArray()

	radius := lut.Radius()
	n := int(radius * 256.0 * 2.0)
	if n < 2 {
		n = 2
	}
	dx := (xEnd - xStart) * radius / 8.0
	dy := yEnd - ys
	xs := (xEnd+xStart)/2.0 - (float64(lut.Diameter())*(xEnd-xStart))/32.0
	nn := lut.Diameter() * 256

	pts := make([]point, 0, nn)
	pts = append(pts, point{
		x: xs + 0.5,
		y: ys + dy*float64(weights[0])/float64(imgacc.ImageFilterScale),
	})
	for i := 1; i < nn; i++ {
		x := xs + dx*float64(i)/float64(n) + 0.5
		y := ys + dy*float64(weights[i])/float64(imgacc.ImageFilterScale)
		pts = append(pts, point{x: x, y: y})
	}
	strokePolyline(ctx, max(0.75, 1.1), agg.RGBA(0.39, 0.0, 0.0, 1.0), pts)
}

func drawControls(ctx *agg.Context, st imageFilters2State) {
	agg2d := ctx.GetAgg2D()
	ras := agg2d.GetInternalRasterizer()

	radius := slider.NewSliderCtrl(115, 5, 495, 11, false)
	radius.SetLabel("Filter Radius=%.3f")
	radius.SetRange(2.0, 8.0)
	radius.SetValue(st.radius)

	filters := rbox.NewDefaultRboxCtrl(0, 0, 110, 210, false)
	filters.SetBorderWidth(0, 0)
	filters.SetBackgroundColor(icol.NewRGBA(0.0, 0.0, 0.0, 0.1))
	filters.SetTextSize(6.0, 0)
	filters.SetTextThickness(0.85)
	filters.AddItem("simple (NN)")
	filters.AddItem("bilinear")
	filters.AddItem("bicubic")
	filters.AddItem("spline16")
	filters.AddItem("spline36")
	filters.AddItem("hanning")
	filters.AddItem("hamming")
	filters.AddItem("hermite")
	filters.AddItem("kaiser")
	filters.AddItem("quadric")
	filters.AddItem("catrom")
	filters.AddItem("gaussian")
	filters.AddItem("bessel")
	filters.AddItem("mitchell")
	filters.AddItem("sinc")
	filters.AddItem("lanczos")
	filters.AddItem("blackman")
	filters.SetCurItem(st.filterIdx)

	normalize := checkbox.NewDefaultCheckboxCtrl(8, 215, "Normalize Filter", false)
	normalize.SetTextSize(7.5, 0)
	normalize.SetChecked(st.normalize)

	if st.filterIdx >= 14 {
		renderCtrl(agg2d, ras, radius)
	}
	renderCtrl(agg2d, ras, filters)
	renderCtrl(agg2d, ras, normalize)
}

type point struct {
	x, y float64
}

type ctrlVertexSource struct {
	ctrl ctrlbase.Ctrl[icol.RGBA]
}

func (a *ctrlVertexSource) Rewind(pathID uint32) {
	a.ctrl.Rewind(uint(pathID))
}

func (a *ctrlVertexSource) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

func renderCtrl(a *agg.Agg2D, ras interface {
	Reset()
	AddPath(vs rasterizer.VertexSource, pathID uint32)
}, c ctrlbase.Ctrl[icol.RGBA]) {
	for pathID := uint(0); pathID < c.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(&ctrlVertexSource{ctrl: c}, uint32(pathID))
		col := c.Color(pathID)
		a.RenderRasterizerWithColor(agg.NewColor(
			uint8(math.Round(col.R*255.0)),
			uint8(math.Round(col.G*255.0)),
			uint8(math.Round(col.B*255.0)),
			uint8(math.Round(col.A*255.0)),
		))
	}
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

var sourceColors = [...]agg.Color{
	agg.Green, agg.Red, agg.White, agg.Blue,
	agg.Blue, agg.Black, agg.White, agg.White,
	agg.White, agg.White, agg.Red, agg.Blue,
	agg.Red, agg.White, agg.Black, agg.Green,
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

func flipImageY(img *agg.Image) {
	if img == nil {
		return
	}
	w, h := img.Width(), img.Height()
	if w == 0 || h == 0 {
		return
	}
	stride := w * 4
	row := make([]byte, stride)
	for y := 0; y < h/2; y++ {
		top := y * stride
		bottom := (h - 1 - y) * stride
		copy(row, img.Data[top:top+stride])
		copy(img.Data[top:top+stride], img.Data[bottom:bottom+stride])
		copy(img.Data[bottom:bottom+stride], row)
	}
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Image Filters 2",
		Width:  frameWidth,
		Height: frameHeight,
		FlipY:  true,
	}, newDemo())
}
