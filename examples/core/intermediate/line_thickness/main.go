// Package main ports AGG's line_thickness.cpp demo.
//
// It renders variable-width anti-aliased lines, applies slight blur, and saves PNG.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/effects"
)

type imagePixFmtAdapter struct {
	img *agg.Image
}

func (p *imagePixFmtAdapter) Width() int  { return p.img.Width() }
func (p *imagePixFmtAdapter) Height() int { return p.img.Height() }

func (p *imagePixFmtAdapter) GetPixel(x, y int) color.RGBA8[color.Linear] {
	if x < 0 || y < 0 || x >= p.img.Width() || y >= p.img.Height() {
		return color.RGBA8[color.Linear]{}
	}
	i := (y*p.img.Width() + x) * 4
	return color.RGBA8[color.Linear]{
		R: basics.Int8u(p.img.Data[i+0]),
		G: basics.Int8u(p.img.Data[i+1]),
		B: basics.Int8u(p.img.Data[i+2]),
		A: basics.Int8u(p.img.Data[i+3]),
	}
}

func (p *imagePixFmtAdapter) CopyPixel(x, y int, c color.RGBA8[color.Linear]) {
	if x < 0 || y < 0 || x >= p.img.Width() || y >= p.img.Height() {
		return
	}
	i := (y*p.img.Width() + x) * 4
	p.img.Data[i+0] = uint8(c.R)
	p.img.Data[i+1] = uint8(c.G)
	p.img.Data[i+2] = uint8(c.B)
	p.img.Data[i+3] = uint8(c.A)
}

const (
	ltFactor     = 1.0
	ltBlurRadius = 1.5
	ltMono       = true
	ltInvert     = false
)

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	a := ctx.GetAgg2D()
	a.ResetTransformations()

	clr1 := agg.RGBA(1, 1, 1, 1)
	clr2 := agg.RGBA(0, 0, 0, 1)
	if !ltMono {
		clr1 = agg.RGBA(1, 0, 1, 1)
		clr2 = agg.RGBA(0, 1, 0, 1)
	}
	foreground := clr2
	background := clr1
	if ltInvert {
		foreground = clr1
		background = clr2
	}

	ctx.Clear(background)
	ctx.SetColor(foreground)

	for i := 0; i < 20; i++ {
		a.LineWidth(ltFactor * 0.3 * float64(i+1))
		a.ResetPath()
		a.MoveTo(20+30*float64(i), 310)
		a.LineTo(40+30*float64(i), 460)
		a.DrawPath(agg.StrokeOnly)
	}

	for i := 0; i < 40; i++ {
		ang := float64(i) * math.Pi / 20.0
		a.LineWidth(ltFactor)
		a.ResetPath()
		a.MoveTo(320+20*math.Sin(ang), 180+20*math.Cos(ang))
		a.LineTo(320+100*math.Sin(ang), 180+100*math.Cos(ang))
		a.DrawPath(agg.StrokeOnly)
	}

	effects.ApplySlightBlurFull(&imagePixFmtAdapter{img: ctx.GetImage()}, ltBlurRadius)
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Line Thickness",
		Width:  640,
		Height: 480,
	}, &demo{})
}
