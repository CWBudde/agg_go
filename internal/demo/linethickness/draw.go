// Package linethickness ports AGG's line_thickness.cpp demo.
package linethickness

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/effects"
)

const (
	BaseWidth  = 640.0
	BaseHeight = 480.0
)

type State struct {
	Thickness float64
	Blur      float64
	Mono      bool
	Invert    bool
}

func DefaultState() State {
	return State{
		Thickness: 1.0,
		Blur:      1.5,
		Mono:      true,
		Invert:    false,
	}
}

func (s *State) Clamp() {
	if s.Thickness < 0 {
		s.Thickness = 0
	}
	if s.Thickness > 5 {
		s.Thickness = 5
	}
	if s.Blur < 0 {
		s.Blur = 0
	}
	if s.Blur > 2 {
		s.Blur = 2
	}
}

func Draw(ctx *agg.Context, st State) {
	st.Clamp()

	fg, bg := colorsForState(st)
	ctx.Clear(bg)
	ctx.SetColor(fg)
	ctx.SetLineCap(agg.CapButt)

	scale, offX, offY := fitFrame(ctx.Width(), ctx.Height())
	mapX := func(x float64) float64 { return offX + x*scale }
	linesTop, wheelCenterY := verticalLayout()
	mapY := func(y float64) float64 { return offY + y*scale }

	a := ctx.GetAgg2D()
	a.ResetTransformations()
	a.LineCap(agg.CapButt)

	for i := 0; i < 20; i++ {
		a.LineWidth(st.Thickness * 0.3 * float64(i+1) * scale)
		a.ResetPath()
		a.MoveTo(mapX(20+30*float64(i)), mapY(linesTop+150))
		a.LineTo(mapX(40+30*float64(i)), mapY(linesTop))
		a.DrawPath(agg.StrokeOnly)
	}

	for i := 0; i < 40; i++ {
		ang := float64(i) * math.Pi / 20.0
		a.LineWidth(st.Thickness * scale)
		a.ResetPath()
		a.MoveTo(mapX(320+20*math.Sin(ang)), mapY(wheelCenterY-20*math.Cos(ang)))
		a.LineTo(mapX(320+100*math.Sin(ang)), mapY(wheelCenterY-100*math.Cos(ang)))
		a.DrawPath(agg.StrokeOnly)
	}

	if st.Blur > 0 {
		effects.ApplySlightBlurFull(&imagePixFmtAdapter{img: ctx.GetImage()}, st.Blur)
	}
}

func verticalLayout() (linesTop, wheelCenterY float64) {
	const (
		linesHeight = 150.0
		wheelRadius = 100.0
		wheelHeight = wheelRadius * 2.0
	)

	remaining := BaseHeight - linesHeight - wheelHeight
	spaceUnit := remaining / 2.0
	linesTop = spaceUnit * 0.5
	wheelTop := linesTop + linesHeight + spaceUnit
	wheelCenterY = wheelTop + wheelRadius
	return linesTop, wheelCenterY
}

func colorsForState(st State) (foreground, background agg.Color) {
	clr1 := agg.RGBA(1, 1, 1, 1)
	clr2 := agg.RGBA(0, 0, 0, 1)
	if !st.Mono {
		clr1 = agg.RGBA(1, 0, 1, 1)
		clr2 = agg.RGBA(0, 1, 0, 1)
	}
	foreground = clr2
	background = clr1
	if st.Invert {
		foreground = clr1
		background = clr2
	}
	return foreground, background
}

func fitFrame(w, h int) (scale, offX, offY float64) {
	sx := float64(w) / BaseWidth
	sy := float64(h) / BaseHeight
	scale = math.Min(sx, sy)
	if scale > 1.0 {
		scale = 1.0
	}
	if scale <= 0 {
		scale = 1.0
	}
	offX = (float64(w) - BaseWidth*scale) * 0.5
	offY = (float64(h) - BaseHeight*scale) * 0.5
	return scale, offX, offY
}

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
