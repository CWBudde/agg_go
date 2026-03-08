// Package main ports AGG's line_thickness.cpp demo.
//
// It renders variable-width anti-aliased lines, applies slight blur, and saves PNG.
package main

import (
	"fmt"
	"math"

	agg "agg_go"
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/effects"
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

func main() {
	const (
		w          = 640
		h          = 480
		factor     = 1.0
		blurRadius = 1.5
		mono       = true
		invert     = false
		out        = "line_thickness.png"
	)

	ctx := agg.NewContext(w, h)
	a := ctx.GetAgg2D()
	a.ResetTransformations()

	clr1 := agg.RGBA(1, 1, 1, 1)
	clr2 := agg.RGBA(0, 0, 0, 1)
	if !mono {
		clr1 = agg.RGBA(1, 0, 1, 1)
		clr2 = agg.RGBA(0, 1, 0, 1)
	}
	foreground := clr2
	background := clr1
	if invert {
		foreground = clr1
		background = clr2
	}

	ctx.Clear(background)
	ctx.SetColor(foreground)

	for i := 0; i < 20; i++ {
		a.LineWidth(factor * 0.3 * float64(i+1))
		a.ResetPath()
		a.MoveTo(20+30*float64(i), 310)
		a.LineTo(40+30*float64(i), 460)
		a.DrawPath(agg.StrokeOnly)
	}

	for i := 0; i < 40; i++ {
		ang := float64(i) * math.Pi / 20.0
		a.LineWidth(factor)
		a.ResetPath()
		a.MoveTo(320+20*math.Sin(ang), 180+20*math.Cos(ang))
		a.LineTo(320+100*math.Sin(ang), 180+100*math.Cos(ang))
		a.DrawPath(agg.StrokeOnly)
	}

	effects.ApplySlightBlurFull(&imagePixFmtAdapter{img: ctx.GetImage()}, blurRadius)

	if err := ctx.GetImage().SaveToPNG(out); err != nil {
		fmt.Printf("error writing %s: %v\n", out, err)
		return
	}
	fmt.Printf("wrote %s (%dx%d)\n", out, w, h)
}
