package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/checkbox"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/rbox"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	"github.com/MeKo-Christian/agg_go/internal/order"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt/blender"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
	"github.com/MeKo-Christian/agg_go/internal/vcgen"
)

const (
	frameWidth  = 500
	frameHeight = 330
)

type control interface {
	InRect(x, y float64) bool
	OnMouseButtonDown(x, y float64) bool
	OnMouseButtonUp(x, y float64) bool
	OnMouseMove(x, y float64, buttonPressed bool) bool
	NumPaths() uint
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
	Color(pathID uint) color.RGBA
}

type controlPathAdapter struct {
	rewindFn func(pathID uint)
	vertexFn func() (x, y float64, cmd basics.PathCommand)
}

func (a *controlPathAdapter) Rewind(pathID uint32) { a.rewindFn(uint(pathID)) }

func (a *controlPathAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.vertexFn()
	*x = vx
	*y = vy
	return uint32(cmd)
}

type pathToConvSource struct{ ps *path.PathStorageStl }

func (a *pathToConvSource) Rewind(pathID uint) { a.ps.Rewind(pathID) }
func (a *pathToConvSource) Vertex() (x, y float64, cmd basics.PathCommand) {
	vx, vy, c := a.ps.NextVertex()
	return vx, vy, basics.PathCommand(c)
}

type convToRasSource struct{ src conv.VertexSource }

func (a *convToRasSource) Rewind(pathID uint32) { a.src.Rewind(uint(pathID)) }
func (a *convToRasSource) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

type rasScanlineAdapter struct{ sl *scanline.ScanlineU8 }

func (a *rasScanlineAdapter) ResetSpans()                 { a.sl.ResetSpans() }
func (a *rasScanlineAdapter) AddCell(x int, cover uint32) { a.sl.AddCell(x, uint(cover)) }
func (a *rasScanlineAdapter) AddSpan(x, length int, cover uint32) {
	a.sl.AddSpan(x, length, uint(cover))
}
func (a *rasScanlineAdapter) Finalize(y int) { a.sl.Finalize(y) }
func (a *rasScanlineAdapter) NumSpans() int  { return a.sl.NumSpans() }

func rgbaToRGBA8(c color.RGBA) color.RGBA8[color.Linear] {
	clamp := func(v float64) uint8 {
		if v <= 0 {
			return 0
		}
		if v >= 1 {
			return 255
		}
		return uint8(v*255 + 0.5)
	}
	return color.RGBA8[color.Linear]{
		R: clamp(c.R),
		G: clamp(c.G),
		B: clamp(c.B),
		A: clamp(c.A),
	}
}

func renderControl(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	sl *scanline.ScanlineU8,
	renBase *renderer.RendererBase[*pixfmt.PixFmtAlphaBlendRGBA[color.Linear, blender.BlenderRGBA8Pre[color.Linear, order.RGBA]], color.RGBA8[color.Linear]],
	ctrl control,
) {
	adapter := &controlPathAdapter{rewindFn: ctrl.Rewind, vertexFn: ctrl.Vertex}
	for pathID := uint(0); pathID < ctrl.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(adapter, uint32(pathID))
		col := rgbaToRGBA8(ctrl.Color(pathID))
		if !ras.RewindScanlines() {
			continue
		}
		sl.Reset(ras.MinX(), ras.MaxX())
		for ras.SweepScanline(&rasScanlineAdapter{sl: sl}) {
			y := sl.Y()
			for _, spanData := range sl.Spans() {
				if spanData.Len > 0 {
					renBase.BlendSolidHspan(int(spanData.X), y, int(spanData.Len), col, spanData.Covers)
				}
			}
		}
	}
}

type demo struct {
	x [3]float64
	y [3]float64

	dx  float64
	dy  float64
	idx int

	capCtrl       *rbox.RboxCtrl[color.RGBA]
	widthCtrl     *slider.SliderCtrl
	smoothCtrl    *slider.SliderCtrl
	closeCtrl     *checkbox.CheckboxCtrl[color.RGBA]
	evenOddCtrl   *checkbox.CheckboxCtrl[color.RGBA]
	controls      []control
	activeControl control
}

func newDemo() *demo {
	d := &demo{
		x:   [3]float64{157, 469, 243},
		y:   [3]float64{60, 170, 310},
		idx: -1,
	}

	d.capCtrl = rbox.NewDefaultRboxCtrl(10, 250, 130, 320, false)
	_ = d.capCtrl.AddItem("Butt Cap")
	_ = d.capCtrl.AddItem("Square Cap")
	_ = d.capCtrl.AddItem("Round Cap")
	d.capCtrl.SetCurItem(0)

	d.widthCtrl = slider.NewSliderCtrl(140, 308, 280, 316, false)
	d.widthCtrl.SetRange(0, 10)
	d.widthCtrl.SetValue(3)
	d.widthCtrl.SetLabel("Width=%1.2f")

	d.smoothCtrl = slider.NewSliderCtrl(290, 308, 490, 316, false)
	d.smoothCtrl.SetRange(0, 2)
	d.smoothCtrl.SetValue(1)
	d.smoothCtrl.SetLabel("Smooth=%1.2f")

	d.closeCtrl = checkbox.NewDefaultCheckboxCtrl(140, 286, "Close Polygons", false)
	d.evenOddCtrl = checkbox.NewDefaultCheckboxCtrl(290, 286, "Even-Odd Fill", false)

	d.controls = []control{d.capCtrl, d.widthCtrl, d.smoothCtrl, d.closeCtrl, d.evenOddCtrl}
	return d
}

func mapPoint(x, y float64) (float64, float64)   { return x, float64(frameHeight) - y }
func unmapPoint(x, y float64) (float64, float64) { return x, float64(frameHeight) - y }

func (d *demo) buildPath() *path.PathStorageStl {
	cx := (d.x[0] + d.x[1] + d.x[2]) / 3
	cy := (d.y[0] + d.y[1] + d.y[2]) / 3

	ps := path.NewPathStorageStl()
	x, y := mapPoint(d.x[0], d.y[0])
	ps.MoveTo(x, y)
	x, y = mapPoint(d.x[1], d.y[1])
	ps.LineTo(x, y)
	x, y = mapPoint(cx, cy)
	ps.LineTo(x, y)
	x, y = mapPoint(d.x[2], d.y[2])
	ps.LineTo(x, y)
	if d.closeCtrl.IsChecked() {
		ps.ClosePolygon(basics.PathFlagsNone)
	}

	x, y = mapPoint((d.x[0]+d.x[1])/2, (d.y[0]+d.y[1])/2)
	ps.MoveTo(x, y)
	x, y = mapPoint((d.x[1]+d.x[2])/2, (d.y[1]+d.y[2])/2)
	ps.LineTo(x, y)
	x, y = mapPoint((d.x[2]+d.x[0])/2, (d.y[2]+d.y[0])/2)
	ps.LineTo(x, y)
	if d.closeCtrl.IsChecked() {
		ps.ClosePolygon(basics.PathFlagsNone)
	}

	return ps
}

func addVertexSourcePath(a *agg.Agg2D, src conv.VertexSource) {
	src.Rewind(0)
	for {
		x, y, cmd := src.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		switch {
		case basics.IsMoveTo(cmd):
			a.MoveTo(x, y)
		case basics.IsLineTo(cmd):
			a.LineTo(x, y)
		case basics.IsClosed(uint32(cmd)):
			a.ClosePolygon()
		}
	}
}

func (d *demo) drawHandles(ctx *agg.Context) {
	for i := 0; i < 3; i++ {
		x, y := mapPoint(d.x[i], d.y[i])
		ctx.SetColor(agg.RGBA(0.8, 0.2, 0.1, 0.6))
		ctx.FillCircle(x, y, 5)
		ctx.SetColor(agg.Black)
		ctx.DrawCircle(x, y, 5)
	}
}

func (d *demo) Render(ctx *agg.Context) {
	ctx.Clear(agg.White)
	a := ctx.GetAgg2D()
	a.ResetTransformations()

	ps := d.buildPath()
	rawSrc := &pathToConvSource{ps: ps}
	a.FillEvenOdd(d.evenOddCtrl.IsChecked())

	a.ResetPath()
	addVertexSourcePath(a, rawSrc)
	a.FillColor(agg.RGBA(0.7, 0.5, 0.1, 0.5))
	a.NoLine()
	a.DrawPath(agg.FillOnly)

	smoothFill := conv.NewConvSmoothPoly1Curve(rawSrc)
	smoothFill.SetSmoothValue(d.smoothCtrl.Value())
	a.ResetPath()
	addVertexSourcePath(a, smoothFill)
	a.FillColor(agg.RGBA(0.1, 0.5, 0.7, 0.1))
	a.NoLine()
	a.DrawPath(agg.FillOnly)

	smoothOutline := conv.NewConvSmoothPoly1(rawSrc)
	smoothOutline.SetSmoothValue(d.smoothCtrl.Value())
	greenStroke := conv.NewConvStroke(smoothOutline)
	greenStroke.SetWidth(1.0)

	imgData := ctx.GetImage().Data
	rbuf := buffer.NewRenderingBufferU8WithData(imgData, frameWidth, frameHeight, frameWidth*4)
	pf := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtAlphaBlendRGBA[color.Linear, blender.BlenderRGBA8Pre[color.Linear, order.RGBA]], color.RGBA8[color.Linear]](pf)
	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	if d.evenOddCtrl.IsChecked() {
		ras.FillingRule(basics.FillEvenOdd)
	} else {
		ras.FillingRule(basics.FillNonZero)
	}
	sl := scanline.NewScanlineU8()
	ras.AddPath(&convToRasSource{src: greenStroke}, 0)
	green := color.RGBA8[color.Linear]{R: 0, G: 153, B: 0, A: 204}
	if ras.RewindScanlines() {
		sl.Reset(ras.MinX(), ras.MaxX())
		for ras.SweepScanline(&rasScanlineAdapter{sl: sl}) {
			y := sl.Y()
			for _, span := range sl.Spans() {
				if span.Len > 0 {
					renBase.BlendSolidHspan(int(span.X), y, int(span.Len), green, span.Covers)
				}
			}
		}
	}
	ras.Reset()

	curve := conv.NewConvSmoothPoly1Curve(rawSrc)
	curve.SetSmoothValue(d.smoothCtrl.Value())
	markers := vcgen.NewVCGenMarkersTerm()
	dash := conv.NewConvDashWithMarkers(curve, markers)
	dash.AddDash(20, 5)
	dash.AddDash(5, 5)
	dash.AddDash(5, 5)
	dash.DashStart(10)

	stroke := conv.NewConvStroke(dash)
	stroke.SetWidth(d.widthCtrl.Value())
	switch d.capCtrl.CurItem() {
	case 1:
		stroke.SetLineCap(basics.SquareCap)
	case 2:
		stroke.SetLineCap(basics.RoundCap)
	default:
		stroke.SetLineCap(basics.ButtCap)
	}

	k := math.Pow(d.widthCtrl.Value(), 0.7)
	ah := shapes.NewArrowhead()
	ah.Head(4*k, 4*k, 3*k, 2*k)
	if !d.closeCtrl.IsChecked() {
		ah.Tail(1*k, 1.5*k, 3*k, 5*k)
	}
	arrow := conv.NewConvMarker(markers, &arrowheadShapes{ah: ah})

	ras.AddPath(&convToRasSource{src: stroke}, 0)
	ras.AddPath(&convToRasSource{src: arrow}, 0)
	black := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255}
	if ras.RewindScanlines() {
		sl.Reset(ras.MinX(), ras.MaxX())
		for ras.SweepScanline(&rasScanlineAdapter{sl: sl}) {
			y := sl.Y()
			for _, span := range sl.Spans() {
				if span.Len > 0 {
					renBase.BlendSolidHspan(int(span.X), y, int(span.Len), black, span.Covers)
				}
			}
		}
	}

	a.FillEvenOdd(false)
	ras.FillingRule(basics.FillNonZero)

	d.drawHandles(ctx)

	for _, ctrl := range d.controls {
		renderControl(ras, sl, renBase, ctrl)
	}
}

type arrowheadShapes struct{ ah *shapes.Arrowhead }

func (a *arrowheadShapes) Rewind(shapeIndex uint) { a.ah.Rewind(uint32(shapeIndex)) }
func (a *arrowheadShapes) Vertex() (x, y float64, cmd basics.PathCommand) {
	var vx, vy float64
	c := a.ah.Vertex(&vx, &vy)
	return vx, vy, c
}

func pointInTriangle(ax, ay, bx, by, cx, cy, px, py float64) bool {
	d1 := (px-bx)*(ay-by) - (ax-bx)*(py-by)
	d2 := (px-cx)*(by-cy) - (bx-cx)*(py-cy)
	d3 := (px-ax)*(cy-ay) - (cx-ax)*(py-ay)
	hasNeg := (d1 < 0) || (d2 < 0) || (d3 < 0)
	hasPos := (d1 > 0) || (d2 > 0) || (d3 > 0)
	return !hasNeg || !hasPos
}

func (d *demo) handleSceneMouseDown(x, y float64) bool {
	x, y = unmapPoint(x, y)
	d.idx = -1
	for i := 0; i < 3; i++ {
		if math.Hypot(x-d.x[i], y-d.y[i]) < 20 {
			d.dx = x - d.x[i]
			d.dy = y - d.y[i]
			d.idx = i
			return true
		}
	}
	if pointInTriangle(d.x[0], d.y[0], d.x[1], d.y[1], d.x[2], d.y[2], x, y) {
		d.dx = x - d.x[0]
		d.dy = y - d.y[0]
		d.idx = 3
		return true
	}
	return false
}

func (d *demo) handleSceneMouseMove(x, y float64) bool {
	x, y = unmapPoint(x, y)
	if d.idx == 3 {
		dx := x - d.dx
		dy := y - d.dy
		d.x[1] -= d.x[0] - dx
		d.y[1] -= d.y[0] - dy
		d.x[2] -= d.x[0] - dx
		d.y[2] -= d.y[0] - dy
		d.x[0] = dx
		d.y[0] = dy
		return true
	}
	if d.idx >= 0 {
		d.x[d.idx] = x - d.dx
		d.y[d.idx] = y - d.dy
		return true
	}
	return false
}

func (d *demo) OnMouseDown(x, y int, btn demorunner.Buttons) bool {
	if !btn.Left {
		return false
	}
	for _, ctrl := range d.controls {
		if ctrl.InRect(float64(x), float64(y)) && ctrl.OnMouseButtonDown(float64(x), float64(y)) {
			d.activeControl = ctrl
			return true
		}
	}
	d.activeControl = nil
	return d.handleSceneMouseDown(float64(x), float64(y))
}

func (d *demo) OnMouseMove(x, y int, btn demorunner.Buttons) bool {
	if d.activeControl != nil {
		return d.activeControl.OnMouseMove(float64(x), float64(y), btn.Left)
	}
	if btn.Left {
		return d.handleSceneMouseMove(float64(x), float64(y))
	}
	d.idx = -1
	return false
}

func (d *demo) OnMouseUp(x, y int, btn demorunner.Buttons) bool {
	_ = btn
	redraw := false
	if d.activeControl != nil {
		redraw = d.activeControl.OnMouseButtonUp(float64(x), float64(y))
		d.activeControl = nil
	}
	d.idx = -1
	return redraw
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Conv Dash Marker",
		Width:  frameWidth,
		Height: frameHeight,
	}, newDemo())
}
