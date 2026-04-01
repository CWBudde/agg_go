// Port of AGG C++ lion_outline.cpp.
//
// This version matches the original interactive demo more closely:
//   - the lion is rendered either through the scanline stroke pipeline or the
//     outline AA rasterizer
//   - a width slider controls the stroke thickness
//   - a checkbox toggles between the two rendering paths
//   - left-drag rotates/scales the lion, right-drag skews it
package main

import (
	"math"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/examples/shared/lowlevelrunner"
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	ctrlbase "github.com/cwbudde/agg_go/internal/ctrl"
	checkboxctrl "github.com/cwbudde/agg_go/internal/ctrl/checkbox"
	sliderctrl "github.com/cwbudde/agg_go/internal/ctrl/slider"
	liondemo "github.com/cwbudde/agg_go/internal/demo/lion"
	"github.com/cwbudde/agg_go/internal/path"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/primitives"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	outline "github.com/cwbudde/agg_go/internal/renderer/outline"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/transform"
)

const (
	frameWidth  = 512
	frameHeight = 512

	defaultOutlineWidth = 1.0
	defaultAngle        = 0.0
	defaultScale        = 1.0
	defaultSkewX        = 0.0
	defaultSkewY        = 0.0
)

type rasterizerType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

type renBaseType = renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]]

type demo struct {
	lion        liondemo.LionData
	baseDX      float64
	baseDY      float64
	angle       float64
	scale       float64
	skewX       float64
	skewY       float64
	width       int
	height      int
	widthCtrl   *sliderctrl.SliderCtrl
	scanlineCtl *checkboxctrl.CheckboxCtrl[color.RGBA]
	controlDrag bool
}

type lionColorView struct {
	data *liondemo.LionData
}

func (v lionColorView) GetColor(index int) color.RGBA8[color.Linear] {
	return v.data.Colors[index]
}

func newDemo() *demo {
	ld := liondemo.Parse()
	pathVS := path.NewPathStorageStlVertexSourceAdapter(ld.Path)
	bounds, ok := basics.BoundingRect[float64](pathVS, basics.SliceGetID(ld.PathIdx), 0, uint(ld.NPaths))
	if !ok {
		panic("lion_outline: bounding rect not found")
	}

	widthCtrl := sliderctrl.NewSliderCtrl(5, 5, 150, 12, false)
	widthCtrl.SetRange(0.0, 4.0)
	widthCtrl.SetValue(defaultOutlineWidth)
	widthCtrl.SetLabel("Width %3.2f")

	scanlineCtrl := checkboxctrl.NewDefaultCheckboxCtrl(160, 5, "Use Scanline Rasterizer", false)
	scanlineCtrl.SetChecked(false)

	return &demo{
		lion:        ld,
		baseDX:      (bounds.X2 - bounds.X1) * 0.5,
		baseDY:      (bounds.Y2 - bounds.Y1) * 0.5,
		angle:       defaultAngle,
		scale:       defaultScale,
		skewX:       defaultSkewX,
		skewY:       defaultSkewY,
		width:       frameWidth,
		height:      frameHeight,
		widthCtrl:   widthCtrl,
		scanlineCtl: scanlineCtrl,
	}
}

func (d *demo) OnInit() {
	d.angle = defaultAngle
	d.scale = defaultScale
	d.skewX = defaultSkewX
	d.skewY = defaultSkewY
	d.widthCtrl.SetValue(defaultOutlineWidth)
	d.scanlineCtl.SetChecked(false)
	d.controlDrag = false
}

func (d *demo) Render(img *agg.Image) {
	d.width = img.Width()
	d.height = img.Height()

	rbuf := buffer.NewRenderingBufferU8WithData(img.Data, d.width, d.height, img.Stride())
	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	rb := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]](pf)
	rb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineP8()

	mtx := transform.NewTransAffine()
	mtx.Multiply(transform.NewTransAffineTranslation(-d.baseDX, -d.baseDY))
	mtx.Multiply(transform.NewTransAffineScaling(d.scale))
	mtx.Multiply(transform.NewTransAffineRotation(d.angle + math.Pi))
	mtx.Multiply(transform.NewTransAffineSkewing(d.skewX/1000.0, d.skewY/1000.0))
	mtx.Multiply(transform.NewTransAffineTranslation(float64(d.width)/2, float64(d.height)/2))

	pathVS := path.NewPathStorageStlVertexSourceAdapter(d.lion.Path)
	transVS := conv.NewConvTransform(pathVS, mtx)

	if d.scanlineCtl.IsChecked() {
		stroke := conv.NewConvStroke(transVS)
		stroke.SetWidth(d.widthCtrl.Value())
		stroke.SetLineJoin(basics.RoundJoin)

		strokeVS := conv.NewRasterizerVertexSourceAdapter(stroke)
		renscan.RenderAllPaths(ras, sl, renscan.NewRendererScanlineAASolidWithRenderer(rb), strokeVS, lionColorView{data: &d.lion}, &d.lion, d.lion.NPaths)
	} else {
		profile := outline.NewLineProfileAA()
		profile.Width(d.widthCtrl.Value() * mtx.GetScale())

		outlineBase := &outlineBaseAdapter{rb: rb}
		renOutline := outline.NewRendererOutlineAA[*outlineBaseAdapter, color.RGBA8[color.Linear]](outlineBase, profile)
		rasOutline := rasterizer.NewRasterizerOutlineAA[*outlineAAAdapter, color.RGBA8[color.Linear]](&outlineAAAdapter{ren: renOutline})

		outlineVS := conv.NewRasterizerVertexSourceAdapter(transVS)
		rasOutline.RenderAllPaths(outlineVS, &d.lion, &d.lion, d.lion.NPaths)
	}

	renderCtrl(ras, sl, rb, d.widthCtrl)
	renderCtrl(ras, sl, rb, d.scanlineCtl)
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	d.controlDrag = false

	if btn.Left {
		if d.widthCtrl.OnMouseButtonDown(fx, fy) {
			d.controlDrag = true
			return true
		}
		if d.scanlineCtl.OnMouseButtonDown(fx, fy) {
			return true
		}

		d.transform(d.width, d.height, fx, fy)
		return true
	}

	if btn.Right {
		d.skewX = fx
		d.skewY = fy
		return true
	}

	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)

	if d.controlDrag {
		return d.widthCtrl.OnMouseMove(fx, fy, btn.Left)
	}

	if btn.Left {
		d.transform(d.width, d.height, fx, fy)
		return true
	}

	if btn.Right {
		d.skewX = fx
		d.skewY = fy
		return true
	}

	return false
}

func (d *demo) OnMouseUp(x, y int, _ lowlevelrunner.Buttons) bool {
	if d.controlDrag {
		d.controlDrag = false
		return d.widthCtrl.OnMouseButtonUp(float64(x), float64(y))
	}
	return false
}

func (d *demo) transform(width, height int, x, y float64) {
	x -= float64(width) / 2
	y -= float64(height) / 2
	d.angle = math.Atan2(y, x)
	d.scale = math.Sqrt(x*x+y*y) / 100.0
}

type outlineBaseAdapter struct {
	rb *renBaseType
}

func (a *outlineBaseAdapter) Width() int  { return a.rb.Width() }
func (a *outlineBaseAdapter) Height() int { return a.rb.Height() }

func (a *outlineBaseAdapter) BlendSolidHSpan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.CoverType) {
	convCovers := make([]basics.Int8u, len(covers))
	for i := range covers {
		convCovers[i] = basics.Int8u(covers[i])
	}
	a.rb.BlendSolidHspan(x, y, length, c, convCovers)
}

func (a *outlineBaseAdapter) BlendSolidVSpan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.CoverType) {
	convCovers := make([]basics.Int8u, len(covers))
	for i := range covers {
		convCovers[i] = basics.Int8u(covers[i])
	}
	a.rb.BlendSolidVspan(x, y, length, c, convCovers)
}

type outlineAAAdapter struct {
	ren *outline.RendererOutlineAA[*outlineBaseAdapter, color.RGBA8[color.Linear]]
}

func (a *outlineAAAdapter) AccurateJoinOnly() bool            { return a.ren.AccurateJoinOnly() }
func (a *outlineAAAdapter) Color(c color.RGBA8[color.Linear]) { a.ren.Color(c) }

//nolint:gocritic // Interface compatibility requires a by-value parameter here.
func (a *outlineAAAdapter) Line0(lp primitives.LineParameters) {
	a.ren.Line0(&lp)
}

//nolint:gocritic // Interface compatibility requires a by-value parameter here.
func (a *outlineAAAdapter) Line1(lp primitives.LineParameters, sx, sy int) {
	a.ren.Line1(&lp, sx, sy)
}

//nolint:gocritic // Interface compatibility requires a by-value parameter here.
func (a *outlineAAAdapter) Line2(lp primitives.LineParameters, ex, ey int) {
	a.ren.Line2(&lp, ex, ey)
}

//nolint:gocritic // Interface compatibility requires a by-value parameter here.
func (a *outlineAAAdapter) Line3(lp primitives.LineParameters, sx, sy, ex, ey int) {
	a.ren.Line3(&lp, sx, sy, ex, ey)
}

func (a *outlineAAAdapter) Pie(x, y, x1, y1, x2, y2 int) { a.ren.Pie(x, y, x1, y1, x2, y2) }
func (a *outlineAAAdapter) Semidot(cmp func(int) bool, x, y, x1, y1 int) {
	a.ren.Semidot(cmp, x, y, x1, y1)
}

type ctrlVertexSourceAdapter struct {
	ctrl ctrlbase.Ctrl[color.RGBA]
}

func (a *ctrlVertexSourceAdapter) Rewind(pathID uint32) { a.ctrl.Rewind(uint(pathID)) }
func (a *ctrlVertexSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

func renderCtrl(
	ras *rasterizerType,
	sl *scanline.ScanlineP8,
	rb *renBaseType,
	ctrl ctrlbase.Ctrl[color.RGBA],
) {
	for pathID := uint(0); pathID < ctrl.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(&ctrlVertexSourceAdapter{ctrl: ctrl}, uint32(pathID))
		renscan.RenderScanlinesAASolid(ras, sl, rb, rgbaToLinear(ctrl.Color(pathID)))
	}
}

func rgbaToLinear(c color.RGBA) color.RGBA8[color.Linear] {
	clamp := func(v float64) uint8 {
		switch {
		case v <= 0:
			return 0
		case v >= 1:
			return 255
		default:
			return uint8(v*255.0 + 0.5)
		}
	}
	return color.RGBA8[color.Linear]{
		R: clamp(c.R),
		G: clamp(c.G),
		B: clamp(c.B),
		A: clamp(c.A),
	}
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:                 "Lion Outline",
		Width:                 frameWidth,
		Height:                frameHeight,
		FlipY:                 true,
		EncodeLinearRGBToSRGB: true,
	}, newDemo())
}
