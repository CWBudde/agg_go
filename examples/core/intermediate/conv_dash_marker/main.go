package main

import (
	"math"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/examples/shared/lowlevelrunner"
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	"github.com/cwbudde/agg_go/internal/ctrl/checkbox"
	"github.com/cwbudde/agg_go/internal/ctrl/rbox"
	"github.com/cwbudde/agg_go/internal/ctrl/slider"
	"github.com/cwbudde/agg_go/internal/path"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/shapes"
	"github.com/cwbudde/agg_go/internal/vcgen"
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

// plainBase is the renderer base type used throughout the demo. C++
// conv_dash_marker renders everything (fills, strokes, controls) through a
// single renderer_base<pixfmt> (plain, non-premultiplied) with color_type
// rgba8 (linear); the framebuffer is sRGB-encoded at save time.
type plainBase = *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]]

type rasterizerAA = *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

// renderSolid rasterizes the currently added path(s) and blends them with a
// solid color, mirroring agg::render_scanlines_aa_solid.
func renderSolid(ras rasterizerAA, sl *scanline.ScanlineU8, renBase plainBase, col color.RGBA8[color.Linear]) {
	if !ras.RewindScanlines() {
		return
	}
	sl.Reset(ras.MinX(), ras.MaxX())
	for ras.SweepScanline(sl) {
		y := sl.Y()
		for _, spanData := range sl.Spans() {
			if spanData.Len > 0 {
				renBase.BlendSolidHspan(int(spanData.X), y, int(spanData.Len), col, spanData.Covers)
			}
		}
	}
}

func renderControl(ras rasterizerAA, sl *scanline.ScanlineU8, renBase plainBase, ctrl control) {
	adapter := &controlPathAdapter{rewindFn: ctrl.Rewind, vertexFn: ctrl.Vertex}
	for pathID := uint(0); pathID < ctrl.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(adapter, uint32(pathID))
		renderSolid(ras, sl, renBase, rgbaToRGBA8(ctrl.Color(pathID)))
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

	d.capCtrl = rbox.NewDefaultRboxCtrl(10, 10, 130, 80, false)
	_ = d.capCtrl.AddItem("Butt Cap")
	_ = d.capCtrl.AddItem("Square Cap")
	_ = d.capCtrl.AddItem("Round Cap")
	d.capCtrl.SetCurItem(0)

	d.widthCtrl = slider.NewSliderCtrl(140, 14, 280, 22, false)
	d.widthCtrl.SetRange(0, 10)
	d.widthCtrl.SetValue(3)
	d.widthCtrl.SetLabel("Width=%1.2f")

	d.smoothCtrl = slider.NewSliderCtrl(290, 14, 490, 22, false)
	d.smoothCtrl.SetRange(0, 2)
	d.smoothCtrl.SetValue(1)
	d.smoothCtrl.SetLabel("Smooth=%1.2f")

	d.closeCtrl = checkbox.NewDefaultCheckboxCtrl(140, 30, "Close Polygons", false)
	d.evenOddCtrl = checkbox.NewDefaultCheckboxCtrl(290, 30, "Even-Odd Fill", false)

	d.controls = []control{d.capCtrl, d.widthCtrl, d.smoothCtrl, d.closeCtrl, d.evenOddCtrl}
	return d
}

func mapPoint(x, y float64) (float64, float64)   { return x, y }
func unmapPoint(x, y float64) (float64, float64) { return x, y }

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

func (d *demo) Render(img *agg.Image) {
	ps := d.buildPath()
	rawSrc := &pathToConvSource{ps: ps}

	rbuf := buffer.NewRenderingBufferU8WithData(img.Data, img.Width(), img.Height(), img.Stride())
	pf := pixfmt.NewPixFmtRGBA32Linear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]](pf)
	renBase.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	// C++ applies the even-odd rule (when enabled) to all four scene passes.
	sceneRule := basics.FillNonZero
	if d.evenOddCtrl.IsChecked() {
		sceneRule = basics.FillEvenOdd
	}

	// (1) Raw path fill — agg::rgba(0.7, 0.5, 0.1, 0.5) as linear bytes.
	ras.Reset()
	ras.FillingRule(sceneRule)
	ras.AddPath(&convToRasSource{src: rawSrc}, 0)
	renderSolid(ras, sl, renBase, color.RGBA8[color.Linear]{R: 179, G: 128, B: 26, A: 128})

	// (2) Smoothed polygon fill — agg::rgba(0.1, 0.5, 0.7, 0.1).
	// C++ feeds conv_smooth_poly1 (raw curve4 commands) directly to the
	// rasterizer, which connects the bezier control points with straight lines.
	// Wrapping in a curve converter would flatten the spline and shrink the fill.
	smoothFill := conv.NewConvSmoothPoly1(rawSrc)
	smoothFill.SetSmoothValue(d.smoothCtrl.Value())
	ras.Reset()
	ras.FillingRule(sceneRule)
	ras.AddPath(&convToRasSource{src: smoothFill}, 0)
	renderSolid(ras, sl, renBase, color.RGBA8[color.Linear]{R: 26, G: 128, B: 179, A: 26})

	// (3) Smoothed outline stroke — agg::rgba(0.0, 0.6, 0.0, 0.8).
	smoothOutline := conv.NewConvSmoothPoly1(rawSrc)
	smoothOutline.SetSmoothValue(d.smoothCtrl.Value())
	greenStroke := conv.NewConvStroke(smoothOutline)
	greenStroke.SetWidth(1.0)
	ras.Reset()
	ras.FillingRule(sceneRule)
	ras.AddPath(&convToRasSource{src: greenStroke}, 0)
	renderSolid(ras, sl, renBase, color.RGBA8[color.Linear]{R: 0, G: 153, B: 0, A: 204})

	// (4) Dashed stroke + arrowhead markers — agg::rgba(0, 0, 0).
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

	ras.Reset()
	ras.FillingRule(sceneRule)
	ras.AddPath(&convToRasSource{src: stroke}, 0)
	ras.AddPath(&convToRasSource{src: arrow}, 0)
	renderSolid(ras, sl, renBase, color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})

	// Controls use the non-zero rule.
	ras.FillingRule(basics.FillNonZero)
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

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
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

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	if d.activeControl != nil {
		return d.activeControl.OnMouseMove(float64(x), float64(y), btn.Left)
	}
	if btn.Left {
		return d.handleSceneMouseMove(float64(x), float64(y))
	}
	d.idx = -1
	return false
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
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
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:                 "Conv Dash Marker",
		Width:                 frameWidth,
		Height:                frameHeight,
		EncodeLinearRGBToSRGB: true,
		FlipY:                 true,
	}, newDemo())
}
