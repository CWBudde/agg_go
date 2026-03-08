package main

import (
	"math"

	agg "agg_go"
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/effects"
)

// Port of AGG C++ line_thickness.cpp.
//
// Web variant keeps controls outside AGG widgets: parameters are controlled
// via JS/URL query params.
var (
	lineThicknessFactor = 1.0 // C++ slider1, range [0..5]
	lineThicknessBlur   = 1.5 // C++ slider2, range [0..2]
	lineThicknessMono   = true
	lineThicknessInvert = false
)

func setLineThicknessFactor(v float64) {
	if v < 0 {
		v = 0
	}
	if v > 5 {
		v = 5
	}
	lineThicknessFactor = v
}

func setLineThicknessBlur(v float64) {
	if v < 0 {
		v = 0
	}
	if v > 2 {
		v = 2
	}
	lineThicknessBlur = v
}

func setLineThicknessMono(v bool) { lineThicknessMono = v }

func setLineThicknessInvert(v bool) { lineThicknessInvert = v }

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

func drawLineThicknessScene(ctx *agg.Context, factor, blurRadius float64, mono, invert bool) {
	a := ctx.GetAgg2D()
	a.ResetTransformations()

	// Keep AGG's original 640x480 framing centered in the web canvas.
	const baseW, baseH = 640.0, 480.0
	offX := (float64(ctx.GetImage().Width()) - baseW) * 0.5
	offY := (float64(ctx.GetImage().Height()) - baseH) * 0.5

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

	// Row of straight lines.
	for i := 0; i < 20; i++ {
		a.LineWidth(factor * 0.3 * float64(i+1))
		a.ResetPath()
		a.MoveTo(offX+20+30*float64(i), offY+310)
		a.LineTo(offX+40+30*float64(i), offY+460)
		a.DrawPath(agg.StrokeOnly)
	}

	// Wheel of lines.
	for i := 0; i < 40; i++ {
		ang := float64(i) * math.Pi / 20.0
		a.LineWidth(factor)
		a.ResetPath()
		a.MoveTo(offX+320+20*math.Sin(ang), offY+180+20*math.Cos(ang))
		a.LineTo(offX+320+100*math.Sin(ang), offY+180+100*math.Cos(ang))
		a.DrawPath(agg.StrokeOnly)
	}

	if blurRadius > 0 {
		effects.ApplySlightBlurFull(&imagePixFmtAdapter{img: ctx.GetImage()}, blurRadius)
	}
}

func drawLineThicknessDemo() {
	drawLineThicknessScene(ctx, lineThicknessFactor, lineThicknessBlur, lineThicknessMono, lineThicknessInvert)
}
