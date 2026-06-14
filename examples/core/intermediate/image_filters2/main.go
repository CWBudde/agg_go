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

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/examples/shared/lowlevelrunner"
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	icol "github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	ctrlbase "github.com/cwbudde/agg_go/internal/ctrl"
	"github.com/cwbudde/agg_go/internal/ctrl/checkbox"
	"github.com/cwbudde/agg_go/internal/ctrl/rbox"
	"github.com/cwbudde/agg_go/internal/ctrl/slider"
	imgacc "github.com/cwbudde/agg_go/internal/image"
	"github.com/cwbudde/agg_go/internal/path"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/span"
	"github.com/cwbudde/agg_go/internal/transform"
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
}

// imagePixFmt is a minimal straight-alpha RGBA32 pixel format over the 4x4
// source buffer, matching the C++ pixfmt used for img_pixf.
type imagePixFmt struct {
	rbuf *buffer.RenderingBufferU8
}

func (p imagePixFmt) Width() int    { return p.rbuf.Width() }
func (p imagePixFmt) Height() int   { return p.rbuf.Height() }
func (p imagePixFmt) PixWidth() int { return 4 }
func (p imagePixFmt) PixPtr(x, y int) []basics.Int8u {
	row := buffer.RowU8(p.rbuf, y)
	return row[x*4:]
}

// imageCloneSource adapts ImageAccessorClone to the span RGBA source interface,
// mirroring C++ image_accessor_clone<pixfmt> (clamp-to-edge sampling).
type imageCloneSource struct {
	accessor *imgacc.ImageAccessorClone[imagePixFmt]
	ipf      *imagePixFmt
}

func (s *imageCloneSource) Width() int                      { return s.ipf.Width() }
func (s *imageCloneSource) Height() int                     { return s.ipf.Height() }
func (s *imageCloneSource) ColorType() string               { return "RGBA8" }
func (s *imageCloneSource) OrderType() icol.ColorOrder      { return icol.OrderRGBA }
func (s *imageCloneSource) Span(x, y, l int) []basics.Int8u { return s.accessor.Span(x, y, l) }
func (s *imageCloneSource) NextX() []basics.Int8u           { return s.accessor.NextX() }
func (s *imageCloneSource) NextY() []basics.Int8u           { return s.accessor.NextY() }
func (s *imageCloneSource) RowPtr(y int) []basics.Int8u     { return s.ipf.PixPtr(0, y) }

// spanImageGenerator is the common shape of the span_image_filter generators.
type spanImageGenerator interface {
	Generate(span []icol.RGBA8[icol.Linear], x, y int)
}

// spanGenAdapter bridges a span_image_filter generator to the renderer's
// SpanGeneratorInterface (which carries an explicit length).
type spanGenAdapter struct{ gen spanImageGenerator }

func (a *spanGenAdapter) Prepare() {}

func (a *spanGenAdapter) Generate(colors []icol.RGBA8[icol.Linear], x, y, length int) {
	if length > len(colors) {
		length = len(colors)
	}
	if length <= 0 {
		return
	}
	a.gen.Generate(colors[:length], x, y)
}

// drawDestinationImage scales the 4x4 source image into the destination
// parallelogram, faithfully mirroring C++ image_filters2.cpp: the general
// span_image_filter_rgba LUT filter (or span_image_filter_rgba_nn for the
// "simple (NN)" case) fed by an image_accessor_clone and rendered through
// render_scanlines_aa onto the straight-alpha destination.
func drawDestinationImage(dstImg, srcImg *agg.Image, st imageFilters2State) {
	dstRbuf := buffer.NewRenderingBufferU8WithData(dstImg.Data, dstImg.Width(), dstImg.Height(), dstImg.Stride())
	pf := pixfmt.NewPixFmtRGBA32Linear(dstRbuf)
	rb := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[icol.Linear], icol.RGBA8[icol.Linear]](pf)

	imgRbuf := buffer.NewRenderingBufferU8WithData(srcImg.Data, srcImg.Width(), srcImg.Height(), srcImg.Stride())
	ipf := imagePixFmt{rbuf: imgRbuf}
	accessor := imgacc.NewImageAccessorClone(&ipf)
	source := &imageCloneSource{accessor: accessor, ipf: &ipf}

	// trans_affine(para, 0,0,4,4): parallelogram (screen) -> source rectangle.
	para := [6]float64{200, 40, 500, 40, 500, 340}
	imgMtx := transform.NewTransAffineParlToRect(para, 0, 0, 4, 4)
	interp := span.NewSpanInterpolatorLinear[*transform.TransAffine](imgMtx, 8)

	alloc := span.NewSpanAllocator[icol.RGBA8[icol.Linear]]()

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	ras.Reset()
	ras.MoveToD(200, 40)
	ras.LineToD(500, 40)
	ras.LineToD(500, 340)
	ras.LineToD(200, 340)

	sg := buildImageSpanGenerator(source, interp, st)
	renscan.RenderScanlinesAA[icol.RGBA8[icol.Linear]](ras, sl, rb, alloc, &spanGenAdapter{gen: sg})
}

func buildImageSpanGenerator(
	source *imageCloneSource,
	interp *span.SpanInterpolatorLinear[*transform.TransAffine],
	st imageFilters2State,
) spanImageGenerator {
	if st.filterIdx == 0 {
		return span.NewSpanImageFilterRGBANNWithParams[*imageCloneSource, *span.SpanInterpolatorLinear[*transform.TransAffine]](
			source, interp,
		)
	}
	return span.NewSpanImageFilterRGBAWithParams[*imageCloneSource, *span.SpanInterpolatorLinear[*transform.TransAffine]](
		source,
		interp,
		imgacc.NewImageFilterLUTWithFilter(newFilter(st.filterIdx, st.radius), st.normalize),
	)
}

// strokeVS adapts a conv.ConvStroke to the rasterizer's VertexSource interface.
type strokeVS struct{ s *conv.ConvStroke }

func (g *strokeVS) Rewind(pathID uint32) { g.s.Rewind(uint(pathID)) }
func (g *strokeVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := g.s.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// srgbaLinear decodes an AGG srgba8 literal into the linear color_type the
// renderer blends in, mirroring C++'s implicit srgba8 -> rgba8 conversion when
// a color is handed to render_scanlines_aa_solid (RGB sRGB-decoded, alpha kept).
func srgbaLinear(r, g, b, a uint8) icol.RGBA8[icol.Linear] {
	return icol.ConvertRGBA8SRGBToLinear(icol.RGBA8[icol.SRGB]{R: r, G: g, B: b, A: a})
}

// drawGraph renders the filter-response graph (grid + axis + response curve)
// with the low-level pipeline used by C++ image_filters2.cpp: a single
// conv_stroke (width 0.8, butt caps) feeding render_scanlines_aa_solid.
func drawGraph(ctx *agg.Context, st imageFilters2State) {
	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8WithData(img.Data, img.Width(), img.Height(), img.Stride())
	pf := pixfmt.NewPixFmtRGBA32Linear(rbuf)
	rb := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[icol.Linear], icol.RGBA8[icol.Linear]](pf)

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	p := path.NewPathStorageStl()
	stroke := conv.NewConvStroke(path.NewPathStorageStlVertexSourceAdapter(p))
	stroke.SetWidth(0.8)
	strokeSrc := &strokeVS{s: stroke}

	xStart := 5.0
	xEnd := 195.0
	yStart := 235.0
	yEnd := float64(frameHeight) - 5.0
	ys := yStart + (yEnd-yStart)/6.0

	for i := 0; i <= 16; i++ {
		x := xStart + (xEnd-xStart)*float64(i)/16.0
		p.RemoveAll()
		p.MoveTo(x+0.5, yStart)
		p.LineTo(x+0.5, yEnd)
		ras.Reset()
		ras.AddPath(strokeSrc, 0)
		a := uint8(100)
		if i == 8 {
			a = 255
		}
		renscan.RenderScanlinesAASolid[icol.RGBA8[icol.Linear]](ras, sl, rb, srgbaLinear(0, 0, 0, a))
	}

	p.RemoveAll()
	p.MoveTo(xStart, ys)
	p.LineTo(xEnd, ys)
	ras.Reset()
	ras.AddPath(strokeSrc, 0)
	renscan.RenderScanlinesAASolid[icol.RGBA8[icol.Linear]](ras, sl, rb, srgbaLinear(0, 0, 0, 255))

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

	p.RemoveAll()
	p.MoveTo(xs+0.5, ys+dy*float64(weights[0])/float64(imgacc.ImageFilterScale))
	for i := 1; i < nn; i++ {
		p.LineTo(
			xs+dx*float64(i)/float64(n)+0.5,
			ys+dy*float64(weights[i])/float64(imgacc.ImageFilterScale),
		)
	}
	ras.Reset()
	ras.AddPath(strokeSrc, 0)
	renscan.RenderScanlinesAASolid[icol.RGBA8[icol.Linear]](ras, sl, rb, srgbaLinear(100, 0, 0, 255))
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
}, c ctrlbase.Ctrl[icol.RGBA],
) {
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

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:                 "Image Filters 2",
		Width:                 frameWidth,
		Height:                frameHeight,
		FlipY:                 true,
		EncodeLinearRGBToSRGB: true,
	}, newDemo())
}
