// Package main ports AGG's rasterizer_compound.cpp demo.
//
// This is a close port of the original interactive example: the layered
// compound rasterizer scene, the alpha sliders, the width slider, and the
// invert-order checkbox are all preserved.
package main

import (
	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/examples/shared/lowlevelrunner"
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	icol "github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	ctrlbase "github.com/cwbudde/agg_go/internal/ctrl"
	checkboxctrl "github.com/cwbudde/agg_go/internal/ctrl/checkbox"
	sliderctrl "github.com/cwbudde/agg_go/internal/ctrl/slider"
	"github.com/cwbudde/agg_go/internal/path"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/shapes"
	"github.com/cwbudde/agg_go/internal/span"
	"github.com/cwbudde/agg_go/internal/transform"
)

const (
	frameWidth  = 440
	frameHeight = 330
)

type compoundNoClip struct{ x1, y1 float64 }

func (c *compoundNoClip) ResetClipping()                 {}
func (c *compoundNoClip) ClipBox(x1, y1, x2, y2 float64) {}
func (c *compoundNoClip) MoveTo(x, y float64)            { c.x1, c.y1 = x, y }
func (c *compoundNoClip) LineTo(outline *rasterizer.RasterizerCellsAAStyled, x, y float64) {
	outline.Line(
		basics.IRound(c.x1*basics.PolySubpixelScale), basics.IRound(c.y1*basics.PolySubpixelScale),
		basics.IRound(x*basics.PolySubpixelScale), basics.IRound(y*basics.PolySubpixelScale),
	)
	c.x1, c.y1 = x, y
}

type rcStyleHandler struct {
	styles []icol.RGBA8[icol.Linear]
}

func (h *rcStyleHandler) IsSolid(style int) bool { return true }
func (h *rcStyleHandler) Color(style int) icol.RGBA8[icol.Linear] {
	if style < 0 || style >= len(h.styles) {
		return icol.RGBA8[icol.Linear]{}
	}
	return h.styles[style]
}
func (h *rcStyleHandler) GenerateSpan(colors []icol.RGBA8[icol.Linear], x, y, length, style int) {}

// srgba8 mirrors the C++ demo's agg::srgba8 literals: assigning an srgba8 to
// the linear color_type performs an sRGB -> linear decode.
func srgba8(r, g, b, a uint8) icol.RGBA8[icol.Linear] {
	return icol.ConvertRGBA8SRGBToLinear(icol.RGBA8[icol.SRGB]{R: r, G: g, B: b, A: a})
}

type rcConvVertexSource interface {
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
}

type rcConvVSAdapter struct {
	vs rcConvVertexSource
}

func (a *rcConvVSAdapter) Rewind(pathID uint32) { a.vs.Rewind(uint(pathID)) }
func (a *rcConvVSAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.vs.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

type rcEllipseConvAdapter struct {
	ell *shapes.Ellipse
}

func (a *rcEllipseConvAdapter) Rewind(pathID uint) { a.ell.Rewind(uint32(pathID)) }
func (a *rcEllipseConvAdapter) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = a.ell.Vertex(&x, &y)
	return x, y, cmd
}

func composeCompoundPath(ps *path.PathStorageStl) {
	ps.RemoveAll()
	ps.MoveTo(28.47, 6.45)
	ps.Curve3(21.58, 1.12, 19.82, 0.29)
	ps.Curve3(17.19, -0.93, 14.21, -0.93)
	ps.Curve3(9.57, -0.93, 6.57, 2.25)
	ps.Curve3(3.56, 5.42, 3.56, 10.60)
	ps.Curve3(3.56, 13.87, 5.03, 16.26)
	ps.Curve3(7.03, 19.58, 11.99, 22.51)
	ps.Curve3(16.94, 25.44, 28.47, 29.64)
	ps.LineTo(28.47, 31.40)
	ps.Curve3(28.47, 38.09, 26.34, 40.58)
	ps.Curve3(24.22, 43.07, 20.17, 43.07)
	ps.Curve3(17.09, 43.07, 15.28, 41.41)
	ps.Curve3(13.43, 39.75, 13.43, 37.60)
	ps.LineTo(13.53, 34.77)
	ps.Curve3(13.53, 32.52, 12.38, 31.30)
	ps.Curve3(11.23, 30.08, 9.38, 30.08)
	ps.Curve3(7.57, 30.08, 6.42, 31.35)
	ps.Curve3(5.27, 32.62, 5.27, 34.81)
	ps.Curve3(5.27, 39.01, 9.57, 42.53)
	ps.Curve3(13.87, 46.04, 21.63, 46.04)
	ps.Curve3(27.59, 46.04, 31.40, 44.04)
	ps.Curve3(34.28, 42.53, 35.64, 39.31)
	ps.Curve3(36.52, 37.21, 36.52, 30.71)
	ps.LineTo(36.52, 15.53)
	ps.Curve3(36.52, 9.13, 36.77, 7.69)
	ps.Curve3(37.01, 6.25, 37.57, 5.76)
	ps.Curve3(38.13, 5.27, 38.87, 5.27)
	ps.Curve3(39.65, 5.27, 40.23, 5.62)
	ps.Curve3(41.26, 6.25, 44.19, 9.18)
	ps.LineTo(44.19, 6.45)
	ps.Curve3(38.72, -0.88, 33.74, -0.88)
	ps.Curve3(31.35, -0.88, 29.93, 0.78)
	ps.Curve3(28.52, 2.44, 28.47, 6.45)
	ps.ClosePolygon(basics.PathFlagsNone)

	ps.MoveTo(28.47, 9.62)
	ps.LineTo(28.47, 26.66)
	ps.Curve3(21.09, 23.73, 18.95, 22.51)
	ps.Curve3(15.09, 20.36, 13.43, 18.02)
	ps.Curve3(11.77, 15.67, 11.77, 12.89)
	ps.Curve3(11.77, 9.38, 13.87, 7.06)
	ps.Curve3(15.97, 4.74, 18.70, 4.74)
	ps.Curve3(22.41, 4.74, 28.47, 9.62)
	ps.ClosePolygon(basics.PathFlagsNone)
}

type demo struct {
	widthCtrl  *sliderctrl.SliderCtrl
	alpha1Ctrl *sliderctrl.SliderCtrl
	alpha2Ctrl *sliderctrl.SliderCtrl
	alpha3Ctrl *sliderctrl.SliderCtrl
	alpha4Ctrl *sliderctrl.SliderCtrl
	invertCtrl *checkboxctrl.CheckboxCtrl[icol.RGBA]
	controls   []ctrlbase.Ctrl[icol.RGBA]
}

type controlVertexSourceAdapter struct {
	ctrl ctrlbase.Ctrl[icol.RGBA]
}

func (a *controlVertexSourceAdapter) Rewind(pathID uint32) { a.ctrl.Rewind(uint(pathID)) }
func (a *controlVertexSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

// ctrlColor converts a control's float color the way the C++ demo's
// slider_ctrl<color_type> does: rgba -> rgba8 is a plain *255 quantization
// with no colorspace conversion.
func ctrlColor(c icol.RGBA) icol.RGBA8[icol.Linear] {
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
	return icol.RGBA8[icol.Linear]{R: clamp(c.R), G: clamp(c.G), B: clamp(c.B), A: clamp(c.A)}
}

type bgRasterizer = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

type plainRenBase = renderer.RendererBase[*pixfmt.PixFmtRGBA32[icol.Linear], icol.RGBA8[icol.Linear]]

// renderCtrl is the Go equivalent of C++ agg::render_ctrl.
func renderCtrl(ras *bgRasterizer, sl *scanline.ScanlineU8, rb *plainRenBase, c ctrlbase.Ctrl[icol.RGBA]) {
	for pathID := uint(0); pathID < c.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(&controlVertexSourceAdapter{ctrl: c}, uint32(pathID))
		renscan.RenderScanlinesAASolid(ras, sl, rb, ctrlColor(c.Color(pathID)))
	}
}

func newDemo() *demo {
	width := sliderctrl.NewSliderCtrl(190, 5, 430, 12, false)
	width.SetRange(-20.0, 50.0)
	width.SetValue(10.0)
	width.SetLabel("Width=%1.2f")

	alpha1 := sliderctrl.NewSliderCtrl(5, 5, 180, 12, false)
	alpha1.SetRange(0.0, 1.0)
	alpha1.SetValue(1.0)
	alpha1.SetLabel("Alpha1=%1.3f")

	alpha2 := sliderctrl.NewSliderCtrl(5, 25, 180, 32, false)
	alpha2.SetRange(0.0, 1.0)
	alpha2.SetValue(1.0)
	alpha2.SetLabel("Alpha2=%1.3f")

	alpha3 := sliderctrl.NewSliderCtrl(5, 45, 180, 52, false)
	alpha3.SetRange(0.0, 1.0)
	alpha3.SetValue(1.0)
	alpha3.SetLabel("Alpha3=%1.3f")

	alpha4 := sliderctrl.NewSliderCtrl(5, 65, 180, 72, false)
	alpha4.SetRange(0.0, 1.0)
	alpha4.SetValue(1.0)
	alpha4.SetLabel("Alpha4=%1.3f")

	invert := checkboxctrl.NewDefaultCheckboxCtrl(190, 25, "Invert Z-Order", false)

	return &demo{
		widthCtrl:  width,
		alpha1Ctrl: alpha1,
		alpha2Ctrl: alpha2,
		alpha3Ctrl: alpha3,
		alpha4Ctrl: alpha4,
		invertCtrl: invert,
		controls:   []ctrlbase.Ctrl[icol.RGBA]{width, alpha1, alpha2, alpha3, alpha4, invert},
	}
}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()
	rbuf := buffer.NewRenderingBufferU8WithData(img.Data, w, h, img.Stride())

	// Mirror the C++ demo's two renderer bases over the same buffer:
	// a plain one for background and controls, a premultiplied one for the
	// compound pass.
	pf := pixfmt.NewPixFmtRGBA32Linear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[icol.Linear], icol.RGBA8[icol.Linear]](pf)

	pfPre := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	renBasePre := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32Pre[icol.Linear], icol.RGBA8[icol.Linear]](pfPre)

	// Background gradient: yellow -> cyan. The C++ demo interpolates
	// srgba8 values in sRGB space and stores them decoded to linear.
	yellow := icol.RGBA8[icol.SRGB]{R: 255, G: 255, B: 0, A: 255}
	cyan := icol.RGBA8[icol.SRGB]{R: 0, G: 255, B: 255, A: 255}
	gradient := make([]icol.RGBA8[icol.Linear], w)
	for x := 0; x < w; x++ {
		k := basics.URound(float64(x) / float64(w) * 255.0)
		gradient[x] = icol.ConvertRGBA8SRGBToLinear(yellow.Gradient(cyan, basics.Int8u(k)))
	}
	for y := 0; y < h; y++ {
		renBase.CopyColorHspan(0, y, w, gradient)
	}

	// Background triangles.
	bgRas := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	bgSl := scanline.NewScanlineU8()

	bgPath := path.NewPathStorageStl()
	bgPath.MoveTo(0, 0)
	bgPath.LineTo(float64(w), 0)
	bgPath.LineTo(float64(w), float64(h))
	bgPath.ClosePolygon(basics.PathFlagsNone)
	bgRas.Reset()
	bgRas.AddPath(&rcConvVSAdapter{vs: path.NewPathStorageStlVertexSourceAdapter(bgPath)}, 0)
	renscan.RenderScanlinesAASolid(bgRas, bgSl, renBase, srgba8(0, 100, 0, 255))

	bgPath2 := path.NewPathStorageStl()
	bgPath2.MoveTo(0, 0)
	bgPath2.LineTo(0, float64(h))
	bgPath2.LineTo(float64(w), 0)
	bgPath2.ClosePolygon(basics.PathFlagsNone)
	bgRas.Reset()
	bgRas.AddPath(&rcConvVSAdapter{vs: path.NewPathStorageStlVertexSourceAdapter(bgPath2)}, 0)
	renscan.RenderScanlinesAASolid(bgRas, bgSl, renBase, srgba8(0, 100, 100, 255))

	// Compound scene.
	ps := path.NewPathStorageStl()
	composeCompoundPath(ps)
	psAdapter := path.NewPathStorageStlVertexSourceAdapter(ps)

	mtx := transform.NewTransAffine()
	mtx.Multiply(transform.NewTransAffineScaling(4.0))
	mtx.Multiply(transform.NewTransAffineTranslation(150, 100))
	transPath := conv.NewConvTransform(psAdapter, mtx)
	curve := conv.NewConvCurve(transPath)
	stroke := conv.NewConvStroke(curve)
	stroke.SetWidth(d.widthCtrl.Value())

	ell := shapes.NewEllipseWithParams(220.0, 180.0, 120.0, 10.0, 128, false)
	ellTrans := conv.NewConvTransform(&rcEllipseConvAdapter{ell: ell}, transform.NewTransAffine())
	ellStroke := conv.NewConvStroke(ellTrans)
	ellStroke.SetWidth(d.widthCtrl.Value() * 0.5)

	styles := []icol.RGBA8[icol.Linear]{
		srgba8(0, 0, 255, 255),
		srgba8(143, 90, 6, 255),
		srgba8(51, 0, 151, 255),
		srgba8(255, 0, 108, 255),
	}
	styles[3].Opacity(d.alpha1Ctrl.Value())
	styles[2].Opacity(d.alpha2Ctrl.Value())
	styles[1].Opacity(d.alpha3Ctrl.Value())
	styles[0].Opacity(d.alpha4Ctrl.Value())
	for i := range styles {
		styles[i].Premultiply()
	}

	rasc := rasterizer.NewRasterizerCompoundAA(&compoundNoClip{})
	if d.invertCtrl.IsChecked() {
		rasc.LayerOrder(basics.LayerInverse)
	} else {
		rasc.LayerOrder(basics.LayerDirect)
	}
	rasc.Styles(3, -1)
	rasc.AddPath(&rcConvVSAdapter{vs: ellStroke}, 0)
	rasc.Styles(2, -1)
	rasc.AddPath(&rcConvVSAdapter{vs: ellTrans}, 0)
	rasc.Styles(1, -1)
	rasc.AddPath(&rcConvVSAdapter{vs: stroke}, 0)
	rasc.Styles(0, -1)
	rasc.AddPath(&rcConvVSAdapter{vs: curve}, 0)

	slAA := scanline.NewScanlineU8()
	alloc := span.NewSpanAllocator[icol.RGBA8[icol.Linear]]()
	styleHandler := &rcStyleHandler{styles: styles}
	renscan.RenderScanlinesCompoundLayered[icol.RGBA8[icol.Linear], *icol.RGBA8[icol.Linear]](
		rasc, slAA, renBasePre, alloc, styleHandler,
	)

	for _, ctrl := range d.controls {
		renderCtrl(bgRas, bgSl, renBase, ctrl)
	}
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if !btn.Left {
		return false
	}
	fx, fy := float64(x), float64(y)
	for _, ctrl := range d.controls {
		if ctrl.OnMouseButtonDown(fx, fy) {
			return true
		}
	}
	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	redraw := false
	for _, ctrl := range d.controls {
		if ctrl.OnMouseMove(fx, fy, btn.Left) {
			redraw = true
		}
	}
	return redraw
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	redraw := false
	for _, ctrl := range d.controls {
		if ctrl.OnMouseButtonUp(fx, fy) {
			redraw = true
		}
	}
	return redraw
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:                 "Rasterizer Compound",
		Width:                 frameWidth,
		Height:                frameHeight,
		FlipY:                 true,
		EncodeLinearRGBToSRGB: true,
	}, newDemo())
}
