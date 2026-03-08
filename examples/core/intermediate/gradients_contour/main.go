// Package main ports AGG's gradients_contour.cpp demo.
//
// Demonstrates contour-based gradients using a Distance Transform: colours
// follow the outline of an arbitrary path. Four polygon shapes (Star, Great
// Britain, Spiral, Glyph) and four gradient modes (Contour, Auto-Contour,
// Conic Angle, Flat Fill) can be cycled interactively.
package main

import (
	"math"

	agg "agg_go"
	"agg_go/examples/shared/demorunner"
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/conv"
	"agg_go/internal/demo/aggshapes"
	"agg_go/internal/path"
	"agg_go/internal/span"
	"agg_go/internal/transform"
)

// --- Adapter types (bridge different vertex-source interfaces) ---

type convVS struct{ vs conv.VertexSource }

func (a *convVS) Rewind(id uint32) { a.vs.Rewind(uint(id)) }
func (a *convVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.vs.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

type pathConvVS struct{ ps *path.PathStorage }

func (a *pathConvVS) Rewind(id uint) { a.ps.Rewind(id) }
func (a *pathConvVS) NextVertex() (x, y float64, cmd uint32) {
	return a.ps.NextVertex()
}

// convAsPath wraps any conv.VertexSource as a path.VertexSource.
type convAsPath struct{ vs conv.VertexSource }

func (a *convAsPath) Rewind(id uint) { a.vs.Rewind(id) }
func (a *convAsPath) NextVertex() (x, y float64, cmd uint32) {
	vx, vy, c := a.vs.Vertex()
	return vx, vy, uint32(c)
}

type pathAsConv struct{ ps *path.PathStorage }

func (a *pathAsConv) Rewind(id uint) { a.ps.Rewind(id) }
func (a *pathAsConv) Vertex() (x, y float64, c basics.PathCommand) {
	vx, vy, cmd := a.ps.NextVertex()
	return vx, vy, basics.PathCommand(cmd)
}

type stlConvVS struct{ ps *path.PathStorageStl }

func (a *stlConvVS) Rewind(id uint) { a.ps.Rewind(id) }
func (a *stlConvVS) Vertex() (x, y float64, c basics.PathCommand) {
	vx, vy, cmd := a.ps.NextVertex()
	return vx, vy, basics.PathCommand(cmd)
}

type pathStorageRasVS struct{ ps *path.PathStorage }

func (a *pathStorageRasVS) Rewind(id uint32) { a.ps.Rewind(uint(id)) }
func (a *pathStorageRasVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ps.NextVertex()
	*x, *y = vx, vy
	return cmd
}

// --- Span generator ---

type contourSpanGen struct {
	interp    *span.SpanInterpolatorLinear[*transform.TransAffine]
	calcFunc  func(x, y, d2 int) int
	reflect   bool
	colors    []color.RGBA8[color.Linear]
	d1scaled  int
	d2scaled  int
	downscale int
}

func (g *contourSpanGen) Prepare() {}

func (g *contourSpanGen) Generate(colors []color.RGBA8[color.Linear], x, y, length int) {
	g.interp.Begin(float64(x)+0.5, float64(y)+0.5, length)
	nColors := len(g.colors)
	dRange := g.d2scaled - g.d1scaled
	if dRange < 1 {
		dRange = 1
	}
	for i := 0; i < length; i++ {
		ix, iy := g.interp.Coordinates()
		d := g.calcFunc(ix>>g.downscale, iy>>g.downscale, g.d2scaled)
		if g.reflect {
			d2 := g.d2scaled * 2
			d = d % d2
			if d < 0 {
				d += d2
			}
			if d >= g.d2scaled {
				d = d2 - d
			}
		}
		ci := ((d - g.d1scaled) * nColors) / dRange
		if ci < 0 {
			ci = 0
		} else if ci >= nColors {
			ci = nColors - 1
		}
		colors[i] = g.colors[ci]
		g.interp.Next()
	}
}

// --- Spiral vertex source ---

type spiral struct {
	x, y, r1, r2, step   float64
	angle, currR, da, dr float64
	started              bool
}

func newSpiral(x, y, r1, r2, step float64) *spiral {
	return &spiral{
		x: x, y: y, r1: r1, r2: r2, step: step,
		da: basics.Deg2RadF(4.0), dr: step / 90.0,
	}
}

func (s *spiral) Rewind(_ uint) { s.angle = 0; s.currR = s.r1; s.started = false }
func (s *spiral) Vertex() (x, y float64, cmd basics.PathCommand) {
	if s.currR > s.r2 {
		return 0, 0, basics.PathCmdStop
	}
	x = s.x + math.Cos(s.angle)*s.currR
	y = s.y + math.Sin(s.angle)*s.currR
	s.currR += s.dr
	s.angle += s.da
	if !s.started {
		s.started = true
		return x, y, basics.PathCmdMoveTo
	}
	return x, y, basics.PathCmdLineTo
}

// --- Color palette (2-colour default: firebrick→yellow) ---

func buildLUT() []color.RGBA8[color.Linear] {
	type stop struct {
		t       float64
		r, g, b uint8
	}
	stops := []stop{
		{0.0, 178, 34, 34},
		{1.0, 255, 255, 0},
	}
	const size = 1024
	lut := make([]color.RGBA8[color.Linear], size)
	lerp := func(a, b uint8, t float64) uint8 {
		return uint8(float64(a)*(1-t) + float64(b)*t + 0.5)
	}
	for i := range lut {
		t := float64(i) / float64(size-1)
		j := 0
		for j < len(stops)-2 && t > stops[j+1].t {
			j++
		}
		dt := stops[j+1].t - stops[j].t
		u := 0.0
		if dt > 0 {
			u = (t - stops[j].t) / dt
		}
		if u < 0 {
			u = 0
		}
		if u > 1 {
			u = 1
		}
		lut[i] = color.RGBA8[color.Linear]{
			R: lerp(stops[j].r, stops[j+1].r, u),
			G: lerp(stops[j].g, stops[j+1].g, u),
			B: lerp(stops[j].b, stops[j+1].b, u),
			A: 255,
		}
	}
	return lut
}

// --- Star path ---

func buildStar() *path.PathStorage {
	ps := path.NewPathStorage()
	ps.MoveTo(12, 40)
	ps.LineTo(52, 40)
	ps.LineTo(72, 6)
	ps.LineTo(92, 40)
	ps.LineTo(132, 40)
	ps.LineTo(112, 76)
	ps.LineTo(132, 112)
	ps.LineTo(92, 112)
	ps.LineTo(72, 148)
	ps.LineTo(52, 112)
	ps.LineTo(12, 112)
	ps.LineTo(32, 76)
	ps.ClosePolygon(0)
	return ps
}

// --- Bounding box ---

func boundingRect(vs conv.VertexSource) (x1, y1, x2, y2 float64, ok bool) {
	vs.Rewind(0)
	first := true
	for {
		x, y, cmd := vs.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		if basics.IsVertex(cmd) {
			if first {
				x1, y1, x2, y2 = x, y, x, y
				first = false
			} else {
				if x < x1 {
					x1 = x
				}
				if y < y1 {
					y1 = y
				}
				if x > x2 {
					x2 = x
				}
				if y > y2 {
					y2 = y
				}
			}
		}
	}
	return x1, y1, x2, y2, !first
}

// --- Demo ---

type demo struct {
	polygon  int // 0=Star,1=GB,2=Spiral,3=Glyph
	gradient int // 0=Contour,1=AutoContour,2=Conic,3=Flat
}

func (d *demo) mainVS() conv.VertexSource {
	switch d.polygon {
	case 1:
		gbPS := path.NewPathStorageStl()
		aggshapes.MakeGBPoly(gbPS)
		return &stlConvVS{ps: gbPS}
	case 2:
		sp := newSpiral(0, 0, 10, 150, 30)
		stroke := conv.NewConvStroke(sp)
		stroke.SetWidth(22.0)
		return stroke
	case 3:
		glyph := path.NewPathStorage()
		buildGlyph(glyph)
		return conv.NewConvCurve(&pathAsConv{ps: glyph})
	default:
		return &pathAsConv{ps: buildStar()}
	}
}

func (d *demo) Render(ctx *agg.Context) {
	ctx.Clear(agg.White)
	a := ctx.GetAgg2D()
	a.ResetTransformations()

	w := float64(ctx.GetImage().Width())
	h := float64(ctx.GetImage().Height())

	vs := d.mainVS()
	x1, y1, x2, y2, ok := boundingRect(vs)
	if !ok {
		return
	}

	margin := 120.0
	scaleX := (w - margin) / (x2 - x1)
	scaleY := (h - margin) / (y2 - y1)
	scale := scaleX
	if scaleY < scale {
		scale = scaleY
	}
	offsetX := (w - scale*(x2-x1)) / 2
	offsetY := (h - scale*(y2-y1)) / 2

	shapeToScreen := transform.NewTransAffine()
	shapeToScreen.Multiply(transform.NewTransAffineTranslation(-x1, -y1))
	shapeToScreen.Multiply(transform.NewTransAffineScaling(scale))
	shapeToScreen.Multiply(transform.NewTransAffineTranslation(offsetX, offsetY))

	colors := buildLUT()
	ras := a.GetInternalRasterizer()

	switch d.gradient {
	case 0, 1:
		gc := span.NewGradientContour()
		gc.SetFrame(0)
		gc.SetD1(0)
		gc.SetD2(512)

		contourPath := path.NewPathStorage()
		if d.gradient == 0 {
			star := buildStar()
			contourPath.ConcatPath(&pathConvVS{ps: star}, 0)
		} else {
			shapeToScaled := transform.NewTransAffine()
			shapeToScaled.Multiply(transform.NewTransAffineTranslation(-x1, -y1))
			shapeToScaled.Multiply(transform.NewTransAffineScaling(scale))
			vs.Rewind(0)
			scaledT := conv.NewConvTransform(vs, shapeToScaled)
			contourPath.ConcatPath(&convAsPath{vs: scaledT}, 0)
		}

		gc.ContourCreate(contourPath)

		gradMtx := transform.NewTransAffineTranslation(-offsetX, -offsetY)
		interp := span.NewSpanInterpolatorLinearDefault(gradMtx)
		downscale := interp.SubpixelShift() - span.GradientSubpixelShift
		if downscale < 0 {
			downscale = 0
		}
		d2s := basics.IRound(100.0 * float64(span.GradientSubpixelScale))
		spanGen := &contourSpanGen{
			interp:    interp,
			calcFunc:  func(x, y, d2 int) int { return gc.Calculate(x, y, d2) },
			reflect:   true,
			colors:    colors,
			d1scaled:  0,
			d2scaled:  d2s,
			downscale: downscale,
		}

		vs.Rewind(0)
		shapeT := conv.NewConvTransform(vs, shapeToScreen)
		ras.Reset()
		ras.AddPath(&convVS{vs: shapeT}, 0)
		a.RenderScanlinesAAWithSpanGen(ras, spanGen)

	case 2:
		cx, cy := w/2, h/2
		conicMtx := transform.NewTransAffineTranslation(-cx, -cy)
		interp := span.NewSpanInterpolatorLinearDefault(conicMtx)
		downscale := interp.SubpixelShift() - span.GradientSubpixelShift
		if downscale < 0 {
			downscale = 0
		}
		d2s := basics.IRound(100.0 * float64(span.GradientSubpixelScale))
		calcFunc := func(x, y, d2 int) int {
			res := math.Atan2(float64(y), float64(x))
			if res < 0 {
				v := math.Abs(1600 - math.Round(math.Abs(res)*float64(d2)/math.Pi/2))
				return int(math.Abs(v))
			}
			return basics.IRound(res * float64(d2) / math.Pi / 2)
		}
		spanGen := &contourSpanGen{
			interp: interp, calcFunc: calcFunc, reflect: false,
			colors: colors, d1scaled: 0, d2scaled: d2s, downscale: downscale,
		}
		vs.Rewind(0)
		shapeT := conv.NewConvTransform(vs, shapeToScreen)
		ras.Reset()
		ras.AddPath(&convVS{vs: shapeT}, 0)
		a.RenderScanlinesAAWithSpanGen(ras, spanGen)

	case 3:
		vs.Rewind(0)
		shapeT := conv.NewConvTransform(vs, shapeToScreen)
		ras.Reset()
		ras.AddPath(&convVS{vs: shapeT}, 0)
		a.RenderRasterizerWithColor(agg.RGBA(0, 0.6, 0, 1.0))
	}
}

func (d *demo) OnMouseDown(x, y int, btn demorunner.Buttons) bool { return false }
func (d *demo) OnMouseMove(x, y int, btn demorunner.Buttons) bool { return false }
func (d *demo) OnMouseUp(x, y int, btn demorunner.Buttons) bool   { return false }

// --- Glyph path ---

func buildGlyph(ps *path.PathStorage) {
	ps.MoveTo(28.47, 6.45)
	ps.Curve3(21.58, 1.12, 19.82, 0.29)
	ps.Curve3(17.19, -0.93, 14.21, -0.93)
	ps.Curve3(9.57, -0.93, 6.57, 2.25)
	ps.Curve3(3.56, 5.42, 3.56, 10.60)
	ps.Curve3(3.56, 13.87, 5.03, 16.26)
	ps.Curve3(6.51, 18.66, 9.15, 19.79)
	ps.Curve3(11.78, 20.93, 15.80, 21.58)
	ps.Curve3(18.24, 21.95, 27.42, 23.11)
	ps.Curve3(27.42, 24.42, 27.42, 24.78)
	ps.Curve3(27.42, 27.27, 26.54, 28.30)
	ps.Curve3(24.79, 30.56, 21.04, 30.56)
	ps.Curve3(17.84, 30.56, 16.18, 29.40)
	ps.Curve3(14.52, 28.25, 13.82, 25.61)
	ps.LineTo(6.35, 26.59)
	ps.Curve3(7.25, 29.51, 8.79, 31.20)
	ps.Curve3(10.34, 32.89, 13.17, 33.81)
	ps.Curve3(16.01, 34.73, 20.00, 34.73)
	ps.Curve3(23.97, 34.73, 26.64, 33.93)
	ps.Curve3(29.31, 33.13, 30.64, 31.85)
	ps.Curve3(31.97, 30.57, 32.54, 28.62)
	ps.Curve3(32.84, 27.44, 32.84, 24.79)
	ps.LineTo(32.84, 16.09)
	ps.Curve3(32.84, 11.21, 33.08, 9.89)
	ps.Curve3(33.33, 8.57, 34.00, 7.40)
	ps.LineTo(26.16, 7.40)
	ps.Curve3(25.65, 8.57, 28.47, 6.45)
	ps.ClosePolygon(0)
	ps.MoveTo(27.42, 18.05)
	ps.Curve3(19.70, 16.72, 18.29, 15.92)
	ps.Curve3(16.53, 14.93, 15.67, 13.52)
	ps.Curve3(14.81, 12.10, 14.81, 10.44)
	ps.Curve3(14.81, 7.86, 16.48, 6.18)
	ps.Curve3(18.15, 4.51, 20.76, 4.51)
	ps.Curve3(23.08, 4.51, 25.09, 5.71)
	ps.Curve3(27.10, 6.92, 27.42, 8.51)
	ps.LineTo(27.42, 18.05)
}

func main() {
	d := &demo{polygon: 0, gradient: 1}
	demorunner.Run(demorunner.Config{
		Title:  "Gradients Contour (Distance Transform)",
		Width:  800,
		Height: 600,
	}, d)
}
