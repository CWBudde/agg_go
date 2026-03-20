// Package main ports AGG's gradients.cpp demo.
//
// The image is rendered in a flipped work buffer and copied with y-flip,
// matching the C++ original's flip_y=true coordinate system. This places
// the gradient circle top-left, spline controls at the bottom, and the
// radio-button selector on the left.
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	icol "github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	gammactrl "github.com/MeKo-Christian/agg_go/internal/ctrl/gamma"
	rboxctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/rbox"
	splinectrl "github.com/MeKo-Christian/agg_go/internal/ctrl/spline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
	"github.com/MeKo-Christian/agg_go/internal/span"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

const (
	gradientsWidth  = 512
	gradientsHeight = 400
	gradCenterX     = 350.0
	gradCenterY     = 280.0
)

type demo struct{}

type simpleVertexSource interface {
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
}

type rasterVertexSourceAdapter struct {
	src simpleVertexSource
}

func (a *rasterVertexSourceAdapter) Rewind(pathID uint32) {
	a.src.Rewind(uint(pathID))
}

func (a *rasterVertexSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

type ellipseSource struct {
	ellipse *shapes.Ellipse
}

func (s *ellipseSource) Rewind(pathID uint) {
	s.ellipse.Rewind(uint32(pathID))
}

func (s *ellipseSource) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = s.ellipse.Vertex(&x, &y)
	return x, y, cmd
}

type gradientColorFunction struct {
	colors  []icol.RGBA8[icol.Linear]
	profile []uint8
}

func (g *gradientColorFunction) Size() int { return 256 }

func (g *gradientColorFunction) ColorAt(index int) icol.RGBA8[icol.Linear] {
	return g.colors[g.profile[index]]
}

func toAggColor(c icol.RGBA) agg.Color {
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
	return agg.NewColor(clamp(c.R), clamp(c.G), clamp(c.B), clamp(c.A))
}

func renderCtrl(a *agg.Agg2D, c ctrlbase.Ctrl[icol.RGBA]) {
	ras := a.GetInternalRasterizer()
	for i := uint(0); i < c.NumPaths(); i++ {
		ras.Reset()
		ras.AddPath(&rasterVertexSourceAdapter{src: c}, uint32(i))
		a.RenderRasterizerWithColor(toAggColor(c.Color(i)))
	}
}

func buildColorProfile(
	splineR, splineG, splineB, splineA *splinectrl.SplineCtrl[icol.RGBA],
) []icol.RGBA8[icol.Linear] {
	colors := make([]icol.RGBA8[icol.Linear], 256)
	for i := range 256 {
		colors[i] = icol.RGBA8[icol.Linear]{
			R: uint8(splineR.Spline()[i]*255.0 + 0.5),
			G: uint8(splineG.Spline()[i]*255.0 + 0.5),
			B: uint8(splineB.Spline()[i]*255.0 + 0.5),
			A: uint8(splineA.Spline()[i]*255.0 + 0.5),
		}
	}
	return colors
}

func newGammaControl() *gammactrl.GammaCtrl {
	gc := gammactrl.NewGammaCtrl(10.0, 10.0, 200.0, 165.0, false)
	gc.SetTextSize(8.0, 0.0)
	return gc
}

func newSplineControls() (*splinectrl.SplineCtrl[icol.RGBA], *splinectrl.SplineCtrl[icol.RGBA], *splinectrl.SplineCtrl[icol.RGBA], *splinectrl.SplineCtrl[icol.RGBA]) {
	splineR := splinectrl.NewSplineCtrlRGBA(210, 10, 460, 45, 6, false)
	splineG := splinectrl.NewSplineCtrlRGBA(210, 50, 460, 85, 6, false)
	splineB := splinectrl.NewSplineCtrlRGBA(210, 90, 460, 125, 6, false)
	splineA := splinectrl.NewSplineCtrlRGBA(210, 130, 460, 165, 6, false)

	splineR.SetBackgroundColor(icol.NewRGBA(1.0, 0.8, 0.8, 1.0))
	splineG.SetBackgroundColor(icol.NewRGBA(0.8, 1.0, 0.8, 1.0))
	splineB.SetBackgroundColor(icol.NewRGBA(0.8, 0.8, 1.0, 1.0))
	splineA.SetBackgroundColor(icol.NewRGBA(1.0, 1.0, 1.0, 1.0))

	splineR.BorderWidth(1.0, 2.0)
	splineG.BorderWidth(1.0, 2.0)
	splineB.BorderWidth(1.0, 2.0)
	splineA.BorderWidth(1.0, 2.0)

	for i := 0; i < 6; i++ {
		x := float64(i) / 5.0
		y := 1.0 - x
		splineR.SetPoint(uint(i), x, y)
		splineG.SetPoint(uint(i), x, y)
		splineB.SetPoint(uint(i), x, y)
		splineA.SetPoint(uint(i), x, 1.0)
	}

	return splineR, splineG, splineB, splineA
}

func newRboxControl() *rboxctrl.RboxCtrl[icol.RGBA] {
	rbox := rboxctrl.NewDefaultRboxCtrl(10.0, 180.0, 200.0, 300.0, false)
	rbox.SetBorderWidth(2.0, 2.0)
	rbox.AddItem("Circular")
	rbox.AddItem("Diamond")
	rbox.AddItem("Linear")
	rbox.AddItem("XY")
	rbox.AddItem("sqrt(XY)")
	rbox.AddItem("Conic")
	rbox.SetCurItem(0)
	return rbox
}

// copyFlipY copies src to dst with vertical flip (y=0 at bottom → y=0 at top).
func copyFlipY(src, dst []uint8, w, h int) {
	stride := w * 4
	for y := 0; y < h; y++ {
		srcOff := (h - 1 - y) * stride
		dstOff := y * stride
		copy(dst[dstOff:dstOff+stride], src[srcOff:srcOff+stride])
	}
}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()

	// Work buffer: y=0 at bottom (flip_y=true convention), copied with y-flip.
	// Control coordinates and center_x/center_y match the C++ original directly.
	workBuf := make([]uint8, w*h*4)
	workImg := agg.NewImage(workBuf, w, h, w*4)
	ctx := agg.NewContextForImage(workImg)
	ctx.Clear(agg.Black)

	a := ctx.GetAgg2D()
	a.ResetTransformations()

	profileCtrl := newGammaControl()
	splineR, splineG, splineB, splineA := newSplineControls()
	rboxCtrl := newRboxControl()

	renderCtrl(a, profileCtrl)
	renderCtrl(a, splineR)
	renderCtrl(a, splineG)
	renderCtrl(a, splineB)
	renderCtrl(a, splineA)
	renderCtrl(a, rboxCtrl)

	colors := buildColorProfile(splineR, splineG, splineB, splineA)
	colorFunc := &gradientColorFunction{
		colors:  colors,
		profile: profileCtrl.Gamma(),
	}

	mtx1 := transform.NewTransAffine()
	mtx1.Multiply(transform.NewTransAffineScaling(1.0))
	mtx1.Multiply(transform.NewTransAffineRotation(0.0))
	mtx1.Multiply(transform.NewTransAffineTranslation(gradCenterX, gradCenterY))

	mtxG1 := transform.NewTransAffine()
	mtxG1.Multiply(transform.NewTransAffineScaling(1.0))
	mtxG1.Multiply(transform.NewTransAffineScalingXY(1.0, 1.0))
	mtxG1.Multiply(transform.NewTransAffineRotation(0.0))
	mtxG1.Multiply(transform.NewTransAffineTranslation(gradCenterX, gradCenterY))
	mtxG1.Invert()

	ellipse := shapes.NewEllipseWithParams(0, 0, 110, 110, 64, false)
	ellipsePath := conv.NewConvTransform(&ellipseSource{ellipse: ellipse}, mtx1)

	var gradientFunc span.GradientFunction = span.GradientRadial{}
	switch rboxCtrl.CurItem() {
	case 1:
		gradientFunc = span.GradientDiamond{}
	case 2:
		gradientFunc = span.GradientLinearX{}
	case 3:
		gradientFunc = span.GradientXY{}
	case 4:
		gradientFunc = span.GradientSqrtXY{}
	case 5:
		gradientFunc = span.GradientConic{}
	}

	interpolator := span.NewSpanInterpolatorLinearDefault(mtxG1)
	spanGen := span.NewSpanGradient(interpolator, gradientFunc, colorFunc, 0, 150)

	ras := a.GetInternalRasterizer()
	ras.Reset()
	ras.AddPath(&rasterVertexSourceAdapter{src: ellipsePath}, 0)
	a.RenderScanlinesAAWithSpanGen(ras, spanGen)

	copyFlipY(workBuf, img.Data, w, h)
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "AGG gradients with Mach bands compensation",
		Width:  gradientsWidth,
		Height: gradientsHeight,
	}, &demo{})
}
