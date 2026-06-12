// Package main ports AGG's conv_contour.cpp demo.
//
// The original example is an interactive contour/orientation tool built on
// AGG_BGR24 (linear color_type): everything renders in linear space and the
// platform encodes linear->sRGB when saving. This port mirrors that with
// linear pixfmts plus EncodeLinearRGBToSRGB in the runner config.
package main

import (
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
	"github.com/cwbudde/agg_go/internal/path"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/transform"
)

const (
	frameWidth  = 440
	frameHeight = 330
)

type rasType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

type renBase = renderer.RendererBase[*pixfmt.PixFmtRGBA32[icol.Linear], icol.RGBA8[icol.Linear]]

// convRasAdapter wraps a conv.VertexSource into the rasterizer interface.
type convRasAdapter struct{ src conv.VertexSource }

func (a *convRasAdapter) Rewind(id uint32) { a.src.Rewind(uint(id)) }
func (a *convRasAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// ctrlRasAdapter wraps a ctrl.Ctrl into the rasterizer interface.
type ctrlRasAdapter struct{ ctrl ctrlbase.Ctrl[icol.RGBA] }

func (a *ctrlRasAdapter) Rewind(id uint32) { a.ctrl.Rewind(uint(id)) }
func (a *ctrlRasAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// ctrlColor converts a control's float color the way the C++ demo does:
// rgba -> rgba8 is a plain *255 quantization with no colorspace conversion.
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

// renderCtrl is the Go equivalent of C++ agg::render_ctrl.
func renderCtrl(ras *rasType, sl *scanline.ScanlineP8, rb *renBase, c ctrlbase.Ctrl[icol.RGBA]) {
	for pathID := uint(0); pathID < c.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(&ctrlRasAdapter{ctrl: c}, uint32(pathID))
		renscan.RenderScanlinesAASolid(ras, sl, rb, ctrlColor(c.Color(pathID)))
	}
}

func composePath(ps *path.PathStorageStl, closeMode int) {
	var flag basics.PathFlag
	switch closeMode {
	case 1:
		flag = basics.PathFlagsCW
	case 2:
		flag = basics.PathFlagsCCW
	default:
		flag = basics.PathFlagsNone
	}

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
	ps.ClosePolygon(flag)

	ps.MoveTo(28.47, 9.62)
	ps.LineTo(28.47, 26.66)
	ps.Curve3(21.09, 23.73, 18.95, 22.51)
	ps.Curve3(15.09, 20.36, 13.43, 18.02)
	ps.Curve3(11.77, 15.67, 11.77, 12.89)
	ps.Curve3(11.77, 9.38, 13.87, 7.06)
	ps.Curve3(15.97, 4.74, 18.70, 4.74)
	ps.Curve3(22.41, 4.74, 28.47, 9.62)
	ps.ClosePolygon(flag)
}

type demo struct {
	closeCtrl      *rbox.RboxCtrl[icol.RGBA]
	widthCtrl      *slider.SliderCtrl
	autoDetectCtrl *checkbox.CheckboxCtrl[icol.RGBA]
	controls       []ctrlbase.Ctrl[icol.RGBA]
}

func newDemo() *demo {
	// C++ ctrl coordinates (flip_y=true, so ctrl flipY=false):
	//   m_close      (10, 10, 130, 80)
	//   m_width      (140, 14, 430, 22)
	//   m_auto_detect(140, 30, "...")
	closeCtrl := rbox.NewDefaultRboxCtrl(10, 10, 130, 80, false)
	_ = closeCtrl.AddItem("Close")
	_ = closeCtrl.AddItem("Close CW")
	_ = closeCtrl.AddItem("Close CCW")
	closeCtrl.SetCurItem(0)

	widthCtrl := slider.NewSliderCtrl(140, 14, 430, 22, false)
	widthCtrl.SetRange(-100.0, 100.0)
	widthCtrl.SetValue(0.0)
	widthCtrl.SetLabel("Width=%1.2f")

	autoDetectCtrl := checkbox.NewDefaultCheckboxCtrl(140, 30, "Autodetect orientation if not defined", false)

	return &demo{
		closeCtrl:      closeCtrl,
		widthCtrl:      widthCtrl,
		autoDetectCtrl: autoDetectCtrl,
		controls:       []ctrlbase.Ctrl[icol.RGBA]{closeCtrl, widthCtrl, autoDetectCtrl},
	}
}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()
	rbuf := buffer.NewRenderingBufferU8WithData(img.Data, w, h, img.Stride())

	pf := pixfmt.NewPixFmtRGBA32Linear(rbuf)
	rb := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[icol.Linear], icol.RGBA8[icol.Linear]](pf)
	rb.Clear(icol.RGBA8[icol.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineP8()

	mtx := transform.NewTransAffine()
	mtx.Multiply(transform.NewTransAffineScaling(4.0))
	mtx.Multiply(transform.NewTransAffineTranslation(150, 100))

	ps := path.NewPathStorageStl()
	composePath(ps, d.closeCtrl.CurItem())

	trans := conv.NewConvTransform(path.NewPathStorageStlVertexSourceAdapter(ps), mtx)
	curve := conv.NewConvCurve(trans)
	contour := conv.NewConvContour(curve)
	contour.Width(d.widthCtrl.Value())
	contour.AutoDetectOrientation(d.autoDetectCtrl.IsChecked())

	ras.Reset()
	ras.AddPath(&convRasAdapter{src: contour}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, rb, icol.RGBA8[icol.Linear]{R: 0, G: 0, B: 0, A: 255})

	for _, c := range d.controls {
		renderCtrl(ras, sl, rb, c)
	}
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if !btn.Left {
		return false
	}
	fx, fy := float64(x), float64(y)
	for _, c := range d.controls {
		if c.OnMouseButtonDown(fx, fy) {
			return true
		}
	}
	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	redraw := false
	for _, c := range d.controls {
		if c.OnMouseMove(fx, fy, btn.Left) {
			redraw = true
		}
	}
	return redraw
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	redraw := false
	for _, c := range d.controls {
		if c.OnMouseButtonUp(fx, fy) {
			redraw = true
		}
	}
	return redraw
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:                 "Conv Contour",
		Width:                 frameWidth,
		Height:                frameHeight,
		FlipY:                 true,
		EncodeLinearRGBToSRGB: true,
	}, newDemo())
}
