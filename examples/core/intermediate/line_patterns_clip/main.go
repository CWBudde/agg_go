// Package main ports AGG's line_patterns_clip.cpp demo.
//
// Renders a polyline with a chain-link image pattern using two-pass clipping:
// the first pass draws outside a 50-pixel border (unclipped), then a
// semi-transparent white overlay fades it, and the second pass draws the
// crisp clipped result inside the border. Matches line_patterns_clip.cpp
// with flip_y=true and default scale_x=1, start_x=0.
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	icolor "github.com/MeKo-Christian/agg_go/internal/color"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	sliderctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	"github.com/MeKo-Christian/agg_go/internal/demo/linepatterns"
	"github.com/MeKo-Christian/agg_go/internal/order"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt/blender"
	"github.com/MeKo-Christian/agg_go/internal/primitives"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	outline "github.com/MeKo-Christian/agg_go/internal/renderer/outline"
)

// default polygon polyline matching C++ m_line1 initial state.
var defaultPoints = [][2]float64{
	{20, 20},
	{480, 480},
	{440, 20},
	{40, 460},
	{100, 300},
}

// -- pattern source ---------------------------------------------------------

type chainPatternSrc struct {
	img linepatterns.PatternImage
}

func (s *chainPatternSrc) Width() float64  { return float64(s.img.Width) }
func (s *chainPatternSrc) Height() float64 { return float64(s.img.Height) }
func (s *chainPatternSrc) Pixel(x, y int) icolor.RGBA {
	if x < 0 || y < 0 || x >= s.img.Width || y >= s.img.Height {
		return icolor.NewRGBA(0, 0, 0, 0)
	}
	p := s.img.Pixels[y*s.img.Width+x]
	r := uint8((p >> 16) & 0xFF)
	g := uint8((p >> 8) & 0xFF)
	b := uint8(p & 0xFF)
	c := icolor.NewRGBAFromRGBA8(r, g, b, linepatterns.BrightnessToAlpha(int(r)+int(g)+int(b)))
	return c
}

// -- renderer-base adapter for the image outline renderer -------------------

type renBaseType = renderer.RendererBase[
	*pixfmt.PixFmtAlphaBlendRGBA[icolor.Linear, blender.BlenderRGBA8Pre[icolor.Linear, order.RGBA]],
	icolor.RGBA8[icolor.Linear],
]

type imgBaseAdaptor struct{ rb *renBaseType }

func clampF(v float64) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	return uint8(v*255 + 0.5)
}

func rgbaToRGBA8Pre(c icolor.RGBA) icolor.RGBA8[icolor.Linear] {
	out := icolor.RGBA8[icolor.Linear]{R: clampF(c.R), G: clampF(c.G), B: clampF(c.B), A: clampF(c.A)}
	out.Premultiply()
	return out
}

func (a *imgBaseAdaptor) BlendColorHSpan(x, y, length int, colors []icolor.RGBA, _ []basics.CoverType) {
	buf := make([]icolor.RGBA8[icolor.Linear], length)
	for i := range buf {
		buf[i] = rgbaToRGBA8Pre(colors[i])
	}
	a.rb.BlendColorHspan(x, y, length, buf, nil, basics.CoverFull)
}

func (a *imgBaseAdaptor) BlendColorVSpan(x, y, length int, colors []icolor.RGBA, _ []basics.CoverType) {
	buf := make([]icolor.RGBA8[icolor.Linear], length)
	for i := range buf {
		buf[i] = rgbaToRGBA8Pre(colors[i])
	}
	a.rb.BlendColorVspan(x, y, length, buf, nil, basics.CoverFull)
}

// -- solid base adapter for the AA outline renderer -------------------------

type solidBaseAdaptor struct{ rb *renBaseType }

func (a *solidBaseAdaptor) Width() int  { return a.rb.Width() }
func (a *solidBaseAdaptor) Height() int { return a.rb.Height() }
func (a *solidBaseAdaptor) BlendSolidHSpan(x, y, length int, c icolor.RGBA8[icolor.Linear], covers []basics.CoverType) {
	a.rb.BlendSolidHspan(x, y, length, c, covers)
}
func (a *solidBaseAdaptor) BlendSolidVSpan(x, y, length int, c icolor.RGBA8[icolor.Linear], covers []basics.CoverType) {
	a.rb.BlendSolidVspan(x, y, length, c, covers)
}

// -- outline image renderer adapter -----------------------------------------

type imgOutlineAdaptor struct{ ren *outline.RendererOutlineImage }

func (a *imgOutlineAdaptor) AccurateJoinOnly() bool              { return a.ren.AccurateJoinOnly() }
func (a *imgOutlineAdaptor) Color(_ icolor.RGBA8[icolor.Linear]) {}
func (a *imgOutlineAdaptor) Pie(x, y, x1, y1, x2, y2 int)        { a.ren.Pie(x, y, x1, y1, x2, y2) }
func (a *imgOutlineAdaptor) Semidot(cmp func(int) bool, x, y, x1, y1 int) {
	a.ren.Semidot(cmp, x, y, x1, y1)
}
func (a *imgOutlineAdaptor) Line0(lp primitives.LineParameters) { a.ren.Line0(&lp) }
func (a *imgOutlineAdaptor) Line1(lp primitives.LineParameters, sx, sy int) {
	a.ren.Line1(&lp, sx, sy)
}
func (a *imgOutlineAdaptor) Line2(lp primitives.LineParameters, ex, ey int) {
	a.ren.Line2(&lp, ex, ey)
}
func (a *imgOutlineAdaptor) Line3(lp primitives.LineParameters, sx, sy, ex, ey int) {
	a.ren.Line3(&lp, sx, sy, ex, ey)
}

// -- solid AA outline renderer adapter --------------------------------------

type solidOutlineAdaptor struct {
	ren *outline.RendererOutlineAA[*solidBaseAdaptor, icolor.RGBA8[icolor.Linear]]
}

func (a *solidOutlineAdaptor) AccurateJoinOnly() bool              { return a.ren.AccurateJoinOnly() }
func (a *solidOutlineAdaptor) Color(c icolor.RGBA8[icolor.Linear]) { a.ren.Color(c) }
func (a *solidOutlineAdaptor) Pie(x, y, x1, y1, x2, y2 int)        { a.ren.Pie(x, y, x1, y1, x2, y2) }
func (a *solidOutlineAdaptor) Semidot(cmp func(int) bool, x, y, x1, y1 int) {
	a.ren.Semidot(cmp, x, y, x1, y1)
}
func (a *solidOutlineAdaptor) Line0(lp primitives.LineParameters) { a.ren.Line0(&lp) }
func (a *solidOutlineAdaptor) Line1(lp primitives.LineParameters, sx, sy int) {
	a.ren.Line1(&lp, sx, sy)
}
func (a *solidOutlineAdaptor) Line2(lp primitives.LineParameters, ex, ey int) {
	a.ren.Line2(&lp, ex, ey)
}
func (a *solidOutlineAdaptor) Line3(lp primitives.LineParameters, sx, sy, ex, ey int) {
	a.ren.Line3(&lp, sx, sy, ex, ey)
}

// -- path source adapter ----------------------------------------------------

type psAdaptor struct{ ps *path.PathStorageStl }

func (a *psAdaptor) Rewind(pathID uint32) { a.ps.Rewind(uint(pathID)) }
func (a *psAdaptor) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ps.NextVertex()
	*x, *y = vx, vy
	return cmd
}

// -- ctrl rendering helpers (mirrors idea.cpp pattern) ----------------------

type ctrlIface interface {
	NumPaths() uint
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
	Color(pathID uint) icolor.RGBA
}

type ctrlVSAdaptor struct{ c ctrlIface }

func (a *ctrlVSAdaptor) Rewind(id uint32) { a.c.Rewind(uint(id)) }
func (a *ctrlVSAdaptor) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.c.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

func renderCtrl(ag *agg.Agg2D, ctrl ctrlIface) {
	ras := ag.GetInternalRasterizer()
	vs := &ctrlVSAdaptor{ctrl}
	for id := uint(0); id < ctrl.NumPaths(); id++ {
		ras.Reset()
		ras.AddPath(vs, uint32(id))
		c := ctrl.Color(id)
		toU8 := func(v float64) uint8 {
			if v <= 0 {
				return 0
			}
			if v >= 1 {
				return 255
			}
			return uint8(v*255 + 0.5)
		}
		ag.RenderRasterizerWithColor(agg.NewColor(toU8(c.R), toU8(c.G), toU8(c.B), toU8(c.A)))
	}
}

// ensure ctrlbase is used (for the ctrl.BaseCtrl dependency in slider).
var _ = ctrlbase.NewBaseCtrl

// -- demo -------------------------------------------------------------------

type demo struct{}

func buildPath(pts [][2]float64) *path.PathStorageStl {
	ps := path.NewPathStorageStl()
	ps.MoveTo(pts[0][0], pts[0][1])
	for i := 1; i < len(pts); i++ {
		ps.LineTo(pts[i][0], pts[i][1])
	}
	return ps
}

func (d *demo) Render(img *agg.Image) {
	w := img.Width()
	h := img.Height()

	// ----- low-level pipeline over the RGBA premultiplied buffer -----
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, w, h, w*4)
	pf := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[
		*pixfmt.PixFmtAlphaBlendRGBA[icolor.Linear, blender.BlenderRGBA8Pre[icolor.Linear, order.RGBA]],
		icolor.RGBA8[icolor.Linear]](pf)

	renBase.Clear(icolor.RGBA8[icolor.Linear]{R: 128, G: 191, B: 217, A: 255})

	// Pattern: chain.bmp via BrightnessToAlpha (matches C++ pattern_src_brightness_to_alpha).
	patSrc := &chainPatternSrc{img: linepatterns.Images[0]}
	filter := outline.NewPatternFilterRGBAAdapter()
	pattern := outline.NewLineImagePattern(filter)
	pattern.Create(patSrc)

	// Image-pattern outline renderer.
	imgBA := &imgBaseAdaptor{rb: renBase}
	renImg := outline.NewRendererOutlineImage(imgBA, pattern)
	renImg.SetScaleX(1.0)
	renImg.SetStartX(0.0)
	rasImg := rasterizer.NewRasterizerOutlineAA[*imgOutlineAdaptor, icolor.RGBA8[icolor.Linear]](
		&imgOutlineAdaptor{ren: renImg})
	rasImg.SetRoundCap(true)

	// Solid AA-profile line renderer (dark blue, width=8, smoother=10).
	profile := outline.NewLineProfileAA()
	profile.SmootherWidth(10.0)
	profile.Width(8.0)
	solidBA := &solidBaseAdaptor{rb: renBase}
	renLine := outline.NewRendererOutlineAA[*solidBaseAdaptor, icolor.RGBA8[icolor.Linear]](solidBA, profile)
	renLine.Color(icolor.RGBA8[icolor.Linear]{R: 0, G: 0, B: 127, A: 255})
	rasLine := rasterizer.NewRasterizerOutlineAA[*solidOutlineAdaptor, icolor.RGBA8[icolor.Linear]](
		&solidOutlineAdaptor{ren: renLine})
	rasLine.SetRoundCap(true)

	// Clip box: slightly wider than the raster clip so caps draw correctly.
	const clipPad = 9.0
	renImg.ClipBox(50-clipPad, 50-clipPad, float64(w)-50+clipPad, float64(h)-50+clipPad)
	renLine.ClipBox(50-clipPad, 50-clipPad, float64(w)-50+clipPad, float64(h)-50+clipPad)

	// ---- First pass: draw polyline outside the 50-px border ----
	ps := buildPath(defaultPoints)
	rasLine.AddPath(&psAdaptor{ps}, 0)
	rasImg.AddPath(&psAdaptor{ps}, 0)

	// Semi-transparent white overlay (alpha=200) fades the unclipped lines.
	white := icolor.RGBA8[icolor.Linear]{R: 255, G: 255, B: 255, A: 255}
	renBase.BlendBar(0, 0, w-1, h-1, white, 200)

	// Raster clip and clear interior to white.
	renBase.ClipBox(50, 50, w-50, h-50)
	renBase.CopyBar(0, 0, w-1, h-1, white)

	// ---- Second pass: draw again, now clipped to the inner box ----
	ps = buildPath(defaultPoints)
	rasLine.AddPath(&psAdaptor{ps}, 0)
	rasImg.AddPath(&psAdaptor{ps}, 0)

	renBase.ResetClipping(true)

	// ----- Agg2D: polygon-ctrl guide + slider controls -----
	ctx := agg.NewContextForImage(img)
	a := ctx.GetAgg2D()
	a.ResetTransformations()

	// Polygon guide (matches C++ polygon_ctrl rendering: line + point circles).
	a.LineColor(agg.NewColor(0, 77, 128, 76))
	a.LineWidth(1.0)
	a.NoFill()
	a.ResetPath()
	a.MoveTo(defaultPoints[0][0], defaultPoints[0][1])
	for i := 1; i < len(defaultPoints); i++ {
		a.LineTo(defaultPoints[i][0], defaultPoints[i][1])
	}
	a.DrawPath(agg.StrokeOnly)
	a.NoLine()
	a.FillColor(agg.NewColor(0, 77, 128, 128))
	for _, pt := range defaultPoints {
		a.FillCircle(pt[0], pt[1], 5.0)
	}

	// Sliders at y=5..12 matching C++ source coords (flip_y=true equivalent;
	// the demorunner flips the PNG so y=5 appears at the visual bottom).
	scaleSlider := sliderctrl.NewSliderCtrl(5, 5, 240, 12, false)
	scaleSlider.SetLabel("Scale X=%.2f")
	scaleSlider.SetRange(0.2, 3.0)
	scaleSlider.SetValue(1.0)

	startSlider := sliderctrl.NewSliderCtrl(250, 5, float64(w)-5, 12, false)
	startSlider.SetLabel("Start X=%.2f")
	startSlider.SetRange(0.0, 10.0)
	startSlider.SetValue(0.0)

	renderCtrl(a, scaleSlider)
	renderCtrl(a, startSlider)
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Line Patterns Clip",
		Width:  500,
		Height: 500,
	}, &demo{})
}
