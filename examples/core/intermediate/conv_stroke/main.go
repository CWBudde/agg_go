// Port of AGG C++ conv_stroke.cpp example.
//
// This matches the original interactive "Line Join" demo: one stroked path,
// one dashed overlay, a filled base path, two rbox controls, two sliders, and
// draggable triangle points.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	icol "github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	rboxctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/rbox"
	sliderctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	"github.com/MeKo-Christian/agg_go/internal/path"
)

const (
	frameWidth  = 500
	frameHeight = 330
)

type demo struct {
	points    [3][2]float64
	selected  int
	dragDX    float64
	dragDY    float64
	joinCtrl  *rboxctrl.RboxCtrl[icol.RGBA]
	capCtrl   *rboxctrl.RboxCtrl[icol.RGBA]
	widthCtrl *sliderctrl.SliderCtrl
	miterCtrl *sliderctrl.SliderCtrl
	controls  []ctrlbase.Ctrl[icol.RGBA]
}

type ctrlVertexSourceAdapter struct {
	ctrl ctrlbase.Ctrl[icol.RGBA]
}

func (a *ctrlVertexSourceAdapter) Rewind(pathID uint32) {
	a.ctrl.Rewind(uint(pathID))
}

func (a *ctrlVertexSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
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
	for pathID := uint(0); pathID < c.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(&ctrlVertexSourceAdapter{ctrl: c}, uint32(pathID))
		a.RenderRasterizerWithColor(toAggColor(c.Color(pathID)))
	}
}

func newDemo() *demo {
	join := rboxctrl.NewDefaultRboxCtrl(10, 10, 133, 80, false)
	join.SetTextSize(7.5, 0)
	join.SetTextThickness(1.0)
	join.AddItem("Miter Join")
	join.AddItem("Miter Join Revert")
	join.AddItem("Round Join")
	join.AddItem("Bevel Join")
	join.SetCurItem(2)

	capCtrl := rboxctrl.NewDefaultRboxCtrl(10, 90, 133, 170, false)
	capCtrl.AddItem("Butt Cap")
	capCtrl.AddItem("Square Cap")
	capCtrl.AddItem("Round Cap")
	capCtrl.SetCurItem(2)

	width := sliderctrl.NewSliderCtrl(140, 14, 490, 22, false)
	width.SetRange(3.0, 40.0)
	width.SetValue(20.0)
	width.SetLabel("Width=%1.2f")

	miter := sliderctrl.NewSliderCtrl(140, 34, 490, 42, false)
	miter.SetRange(1.0, 10.0)
	miter.SetValue(4.0)
	miter.SetLabel("Miter Limit=%1.2f")

	return &demo{
		points: [3][2]float64{
			{57 + 100, 60},
			{369 + 100, 170},
			{143 + 100, 310},
		},
		selected:  -1,
		joinCtrl:  join,
		capCtrl:   capCtrl,
		widthCtrl: width,
		miterCtrl: miter,
		controls:  []ctrlbase.Ctrl[icol.RGBA]{join, capCtrl, width, miter},
	}
}

func (d *demo) Render(img *agg.Image) {
	ctx := agg.NewContextForImage(img)
	ctx.Clear(agg.White)

	a := ctx.GetAgg2D()
	a.ResetTransformations()

	joinStyles := []basics.LineJoin{
		basics.MiterJoin,
		basics.MiterJoinRevert,
		basics.RoundJoin,
		basics.BevelJoin,
	}
	capStyles := []basics.LineCap{
		basics.ButtCap,
		basics.SquareCap,
		basics.RoundCap,
	}

	join := joinStyles[d.joinCtrl.CurItem()]
	capStyle := capStyles[d.capCtrl.CurItem()]
	strokeWidth := d.widthCtrl.Value()
	miterLimit := d.miterCtrl.Value()

	buildPath := func() {
		a.ResetPath()
		a.MoveTo(d.points[0][0], d.points[0][1])
		a.LineTo((d.points[0][0]+d.points[1][0])/2, (d.points[0][1]+d.points[1][1])/2)
		a.LineTo(d.points[1][0], d.points[1][1])
		a.LineTo(d.points[2][0], d.points[2][1])
		a.LineTo(d.points[2][0], d.points[2][1]) // Numerical stability check from the C++ demo.

		a.MoveTo((d.points[0][0]+d.points[1][0])/2, (d.points[0][1]+d.points[1][1])/2)
		a.LineTo((d.points[1][0]+d.points[2][0])/2, (d.points[1][1]+d.points[2][1])/2)
		a.LineTo((d.points[2][0]+d.points[0][0])/2, (d.points[2][1]+d.points[0][1])/2)
		a.ClosePolygon()
	}

	// (1) Wide stroked path with the selected join/cap style.
	buildPath()
	a.LineJoin(agg.LineJoin(join))
	a.LineCap(agg.LineCap(capStyle))
	a.MiterLimit(miterLimit)
	a.LineWidth(strokeWidth)
	a.LineColor(agg.NewColor(204, 178, 153, 255))
	a.NoFill()
	a.DrawPath(agg.StrokeOnly)

	// (2) Thin outline of the raw path in black.
	buildPath()
	a.LineJoin(agg.JoinMiter)
	a.LineCap(agg.CapButt)
	a.LineWidth(1.5)
	a.LineColor(agg.Black)
	a.DrawPath(agg.StrokeOnly)

	// (3) Dashed overlay on the wide stroke.
	// C++ pipeline: path → conv_stroke(wide) → conv_dash → conv_stroke(thin).
	// The dashes follow the outline of the wide stroke, not the center path.
	{
		ps := path.NewPathStorageStl()
		ps.MoveTo(d.points[0][0], d.points[0][1])
		ps.LineTo((d.points[0][0]+d.points[1][0])/2, (d.points[0][1]+d.points[1][1])/2)
		ps.LineTo(d.points[1][0], d.points[1][1])
		ps.LineTo(d.points[2][0], d.points[2][1])
		ps.LineTo(d.points[2][0], d.points[2][1])

		ps.MoveTo((d.points[0][0]+d.points[1][0])/2, (d.points[0][1]+d.points[1][1])/2)
		ps.LineTo((d.points[1][0]+d.points[2][0])/2, (d.points[1][1]+d.points[2][1])/2)
		ps.LineTo((d.points[2][0]+d.points[0][0])/2, (d.points[2][1]+d.points[0][1])/2)
		ps.ClosePolygon(0)

		psAdapter := path.NewPathStorageStlVertexSourceAdapter(ps)

		// Wide stroke (same as step 1).
		wideStroke := conv.NewConvStroke(psAdapter)
		wideStroke.SetLineJoin(join)
		wideStroke.SetLineCap(capStyle)
		wideStroke.SetMiterLimit(miterLimit)
		wideStroke.SetWidth(strokeWidth)

		// Dash the stroked outline.
		dashedStroke := conv.NewConvDash(wideStroke)
		dashedStroke.AddDash(20.0, strokeWidth/2.5)

		// Thin stroke of the dashed outline.
		thinStroke := conv.NewConvStroke(dashedStroke)
		thinStroke.SetMiterLimit(4.0)
		thinStroke.SetWidth(strokeWidth / 5.0)
		thinStroke.SetLineCap(capStyle)
		thinStroke.SetLineJoin(join)

		ras := a.GetInternalRasterizer()
		ras.Reset()
		// Feed vertices from the low-level pipeline into the rasterizer.
		thinStroke.Rewind(0)
		for {
			x, y, cmd := thinStroke.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
			ras.AddVertex(x, y, uint32(cmd))
		}
		a.RenderRasterizerWithColor(agg.NewColor(0, 0, 77, 255))
	}

	// (4) Semi-transparent fill of the raw path.
	buildPath()
	a.FillColor(agg.NewColor(0, 0, 0, 51))
	a.NoLine()
	a.DrawPath(agg.FillOnly)

	for _, ctrl := range d.controls {
		renderCtrl(a, ctrl)
	}
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if !btn.Left {
		return false
	}

	fx, fy := float64(x), float64(y)
	for _, ctrl := range d.controls {
		if ctrl.OnMouseButtonDown(fx, fy) {
			return true
		}
	}

	for i := 0; i < len(d.points); i++ {
		dx := fx - d.points[i][0]
		dy := fy - d.points[i][1]
		if math.Sqrt(dx*dx+dy*dy) < 20.0 {
			d.selected = i
			d.dragDX = dx
			d.dragDY = dy
			return true
		}
	}

	if pointInTriangle(
		d.points[0][0], d.points[0][1],
		d.points[1][0], d.points[1][1],
		d.points[2][0], d.points[2][1],
		fx, fy,
	) {
		d.selected = 3
		d.dragDX = fx - d.points[0][0]
		d.dragDY = fy - d.points[0][1]
		return true
	}

	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	redraw := false

	for _, ctrl := range d.controls {
		if ctrl.OnMouseMove(fx, fy, btn.Left) {
			redraw = true
		}
	}

	if d.selected == -1 || !btn.Left {
		return redraw
	}

	if d.selected == 3 {
		newX := fx - d.dragDX
		newY := fy - d.dragDY
		shiftX := newX - d.points[0][0]
		shiftY := newY - d.points[0][1]
		for i := range d.points {
			d.points[i][0] += shiftX
			d.points[i][1] += shiftY
		}
		return true
	}

	d.points[d.selected][0] = fx - d.dragDX
	d.points[d.selected][1] = fy - d.dragDY
	return true
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	redraw := false

	for _, ctrl := range d.controls {
		if ctrl.OnMouseButtonUp(fx, fy) {
			redraw = true
		}
	}

	if d.selected != -1 {
		d.selected = -1
		redraw = true
	}

	return redraw
}

func pointInTriangle(x1, y1, x2, y2, x3, y3, px, py float64) bool {
	sign := func(ax, ay, bx, by, px, py float64) float64 {
		return (px-bx)*(ay-by) - (ax-bx)*(py-by)
	}
	d1 := sign(x1, y1, x2, y2, px, py)
	d2 := sign(x2, y2, x3, y3, px, py)
	d3 := sign(x3, y3, x1, y1, px, py)
	hasNeg := d1 < 0 || d2 < 0 || d3 < 0
	hasPos := d1 > 0 || d2 > 0 || d3 > 0
	return !hasNeg || !hasPos
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:                 "Conv Stroke",
		Width:                 frameWidth,
		Height:                frameHeight,
		FlipY:                 true,
		EncodeLinearRGBToSRGB: true,
	}, newDemo())
}
