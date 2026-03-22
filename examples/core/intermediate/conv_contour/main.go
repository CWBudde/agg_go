// Package main ports AGG's conv_contour.cpp demo.
//
// The original example is an interactive contour/orientation tool. This Go
// port keeps the same glyph path, contour pipeline, default control values,
// and control layout, but renders them as a static frame.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/checkbox"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/rbox"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

const (
	frameWidth  = 440
	frameHeight = 330
)

type ctrlIface interface {
	NumPaths() uint
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
	Color(pathID uint) color.RGBA
}

type ctrlVS struct{ ctrl ctrlIface }

func (a *ctrlVS) Rewind(pathID uint32) { a.ctrl.Rewind(uint(pathID)) }
func (a *ctrlVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

func linearToSRGB(v float64) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	if v <= 0.0031308 {
		return uint8(v * 12.92 * 255.0)
	}
	return uint8((1.055*math.Pow(v, 1.0/2.4) - 0.055) * 255.0)
}

func renderCtrl(a *agg.Agg2D, c ctrlIface) {
	ras := a.GetInternalRasterizer()
	adapter := &ctrlVS{ctrl: c}
	for pathID := uint(0); pathID < c.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(adapter, uint32(pathID))
		col := c.Color(pathID)
		a.RenderRasterizerWithColor(agg.NewColor(
			linearToSRGB(col.R),
			linearToSRGB(col.G),
			linearToSRGB(col.B),
			linearToSRGB(col.A),
		))
	}
}

type demo struct{}

func newDemo() *demo { return &demo{} }

func composePath(closeMode int) *path.PathStorageStl {
	var flag basics.PathFlag
	switch closeMode {
	case 1:
		flag = basics.PathFlagsCW
	case 2:
		flag = basics.PathFlagsCCW
	default:
		flag = basics.PathFlagsNone
	}

	ps := path.NewPathStorageStl()
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

	return ps
}

func renderGlyph(a *agg.Agg2D, closeMode int, width float64, autoDetect bool) {
	pathStorage := composePath(closeMode)
	adapter := path.NewPathStorageStlVertexSourceAdapter(pathStorage)
	mtx := transform.NewTransAffineFromValues(4.0, 0, 0, 4.0, 150.0, 100.0)
	trans := conv.NewConvTransform(adapter, mtx)
	curve := conv.NewConvCurve(trans)
	contour := conv.NewConvContour(curve)
	contour.Width(width)
	contour.AutoDetectOrientation(autoDetect)

	a.ResetPath()
	contour.Rewind(0)
	for {
		x, y, cmd := contour.Vertex()
		switch {
		case basics.IsStop(cmd):
			goto done
		case basics.IsMoveTo(cmd):
			a.MoveTo(x, y)
		case basics.IsEndPoly(cmd):
			if basics.IsClose(uint32(cmd)) {
				a.ClosePolygon()
			}
		case basics.IsVertex(cmd):
			a.LineTo(x, y)
		}
	}
done:
	a.FillColor(agg.Black)
	a.NoLine()
	a.DrawPath(agg.FillOnly)
}

func (d *demo) Render(img *agg.Image) {
	ctx := agg.NewContextForImage(img)
	ctx.Clear(agg.White)

	a := ctx.GetAgg2D()
	a.ResetTransformations()

	// Match the original C++ example defaults.
	renderGlyph(a, 0, 0.0, false)

	closeCtrl := rbox.NewDefaultRboxCtrl(10, 10, 130, 80, false)
	closeCtrl.SetTextSize(7.5, 0)
	closeCtrl.SetTextThickness(1.0)
	_ = closeCtrl.AddItem("Close")
	_ = closeCtrl.AddItem("Close CW")
	_ = closeCtrl.AddItem("Close CCW")
	closeCtrl.SetCurItem(0)

	widthCtrl := slider.NewSliderCtrl(140, 14, 430, 22, false)
	widthCtrl.SetRange(-100.0, 100.0)
	widthCtrl.SetValue(0.0)
	widthCtrl.SetLabel("Width=%1.2f")

	autoDetectCtrl := checkbox.NewDefaultCheckboxCtrl(140, 30, "Autodetect orientation if not defined", false)
	autoDetectCtrl.SetChecked(false)

	renderCtrl(a, closeCtrl)
	renderCtrl(a, widthCtrl)
	renderCtrl(a, autoDetectCtrl)

	flipImageVertically(img)
}

func flipImageVertically(img *agg.Image) {
	w := img.Width()
	h := img.Height()
	stride := w * 4
	if len(img.Data) < stride*h {
		return
	}
	tmp := make([]uint8, stride)
	for y := 0; y < h/2; y++ {
		top := y * stride
		bottom := (h - 1 - y) * stride
		copy(tmp, img.Data[top:top+stride])
		copy(img.Data[top:top+stride], img.Data[bottom:bottom+stride])
		copy(img.Data[bottom:bottom+stride], tmp)
	}
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Conv Contour",
		Width:  frameWidth,
		Height: frameHeight,
	}, newDemo())
}
