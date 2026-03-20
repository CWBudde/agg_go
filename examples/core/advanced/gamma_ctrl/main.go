// Package main ports AGG's gamma_ctrl.cpp demo.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	icol "github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	gammactrl "github.com/MeKo-Christian/agg_go/internal/ctrl/gamma"
	"github.com/MeKo-Christian/agg_go/internal/gsv"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

const (
	gammaCtrlWidth  = 500
	gammaCtrlHeight = 400
)

type demo struct{}

type rasterVertexSourceAdapter struct {
	src simpleVertexSource
}

type simpleVertexSource interface {
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
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

func renderStrokeEllipse(a *agg.Agg2D, cx, cy, rx, ry float64, steps uint32, width float64, color agg.Color) {
	ras := a.GetInternalRasterizer()
	ellipse := shapes.NewEllipseWithParams(cx, cy, rx, ry, steps, false)
	stroke := conv.NewConvStroke(&ellipseSource{ellipse: ellipse})
	stroke.SetWidth(width)

	ras.Reset()
	ras.AddPath(&rasterVertexSourceAdapter{src: stroke}, 0)
	a.RenderRasterizerWithColor(color)
}

func renderSimpleSource(a *agg.Agg2D, src simpleVertexSource, color agg.Color) {
	ras := a.GetInternalRasterizer()
	ras.Reset()
	ras.AddPath(&rasterVertexSourceAdapter{src: src}, 0)
	a.RenderRasterizerWithColor(color)
}

func drawArrowPair(a *agg.Agg2D, angle float64) {
	const (
		cx = 400.0
		cy = 130.0
	)

	rotate := func(x, y float64) (float64, float64) {
		s, c := math.Sincos(angle)
		return cx + x*c - y*s, cy + x*s + y*c
	}

	red := agg.NewColor(128, 0, 0, 255)
	a.FillColor(red)
	a.NoLine()
	a.ResetPath()

	p0x, p0y := rotate(30, -1)
	p1x, p1y := rotate(60, 0)
	p2x, p2y := rotate(30, 1)
	a.MoveTo(p0x, p0y)
	a.LineTo(p1x, p1y)
	a.LineTo(p2x, p2y)
	a.ClosePolygon()

	p3x, p3y := rotate(27, -1)
	p4x, p4y := rotate(10, 0)
	p5x, p5y := rotate(27, 1)
	a.MoveTo(p3x, p3y)
	a.LineTo(p4x, p4y)
	a.LineTo(p5x, p5y)
	a.ClosePolygon()

	a.DrawPath(agg.FillOnly)
}

func (d *demo) Render(img *agg.Image) {
	ctx := agg.NewContextForImage(img)
	ctx.Clear(agg.White)
	a := ctx.GetAgg2D()
	a.ResetTransformations()

	gc := gammactrl.NewGammaCtrl(10, 10, 300, 200, false)
	gc.SetTextSize(10.0, 12.0)
	gc.Values(1.0, 1.0, 1.0, 1.0)
	renderCtrl(a, gc)

	ras := a.GetInternalRasterizer()
	ras.SetGamma(gc.Y)

	eWidth := float64(gammaCtrlWidth)/2.0 - 10.0
	eCenter := float64(gammaCtrlWidth) / 2.0
	rows := []struct {
		cy    float64
		rx    float64
		ry    float64
		width float64
		color agg.Color
	}{
		{220, eWidth, 15.0, 2.0, agg.NewColor(0, 0, 0x66, 255)},
		{260, eWidth, 15.0, 2.0, agg.NewColor(0, 0, 0x66, 255)},
		{300, eWidth, 15.0, 2.0, agg.NewColor(0, 0, 0x66, 255)},
		{340, eWidth, 15.5, 1.0, agg.NewColor(192, 192, 192, 255)},
		{380, eWidth, 15.5, 0.4, agg.NewColor(127, 127, 127, 255)},
		{420, eWidth, 15.5, 0.1, agg.NewColor(0, 0, 0, 255)},
	}

	for _, row := range rows {
		renderStrokeEllipse(a, eCenter, row.cy, row.rx, row.ry, 100, row.width, row.color)
		renderStrokeEllipse(a, eCenter, row.cy, 11.0, 11.0, 100, row.width, row.color)
	}

	text := gsv.NewGSVText()
	text.SetText("Text 2345")
	text.SetSize(50, 20)
	text.SetStartPoint(320, 10)
	textOutline := gsv.NewGSVTextOutline(text)
	textOutline.SetWidth(2.0)
	textOutline.SetTransform(transform.NewTransAffineSkewing(0.15, 0.0))
	renderSimpleSource(a, textOutline, agg.NewColor(0, 128, 0, 255))

	for i := 0; i < 35; i++ {
		drawArrowPair(a, float64(i)/35.0*math.Pi*2.0)
	}
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Anti-Aliasing Gamma Correction",
		Width:  gammaCtrlWidth,
		Height: gammaCtrlHeight,
	}, &demo{})
}
