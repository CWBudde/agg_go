// Based on the original AGG examples: trans_curve2.cpp.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/polygon"
	"github.com/MeKo-Christian/agg_go/internal/gsv"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

var (
	transCurve2Points1  = [12]float64{60, 40, 170, 130, 230, 270, 370, 330, 430, 470, 550, 550}
	transCurve2Points2  = [12]float64{40, 60, 150, 170, 210, 290, 350, 350, 410, 490, 530, 570}
	transCurve2Selected = -1
	transCurve2Animate  = false
	transCurve2DX1      [6]float64
	transCurve2DY1      [6]float64
	transCurve2DX2      [6]float64
	transCurve2DY2      [6]float64

	transCurve2NumPoints      = 200.0
	transCurve2FixedLen       = true
	transCurve2PreserveXScale = true
)

const (
	transCurve2RefW    = 600.0
	transCurve2RefH    = 600.0
	transCurve2Text    = "Anti-Grain Geometry is designed as a set of loosely coupled algorithms and class templates united with a common idea, so that all the components can be easily combined. Also, the template based design allows you to replace any part of the library without the necessity to modify a single byte in the existing code. "
	transCurve2BaseLen = 1140.0
)

func transCurve2FrameOffset() (float64, float64) {
	return (float64(width) - transCurve2RefW) * 0.5, (float64(height) - transCurve2RefH) * 0.5
}

func initTransCurve2Demo() {
	for i := 0; i < 6; i++ {
		transCurve2DX1[i] = (math.Mod(float64(i*1234+1), 10.0) - 5.0) * 0.5
		transCurve2DY1[i] = (math.Mod(float64(i*5678+2), 10.0) - 5.0) * 0.5
		transCurve2DX2[i] = (math.Mod(float64(i*1234+3), 10.0) - 5.0) * 0.5
		transCurve2DY2[i] = (math.Mod(float64(i*5678+4), 10.0) - 5.0) * 0.5
	}
}

type transDoubleAdapter struct {
	source *conv.ConvBSpline
}

func (a *transDoubleAdapter) Rewind(id uint) { a.source.Rewind(id) }
func (a *transDoubleAdapter) Vertex() (float64, float64, basics.PathCommand) {
	return a.source.Vertex()
}

func drawTransCurve2Demo() {
	initTransCurve2Demo()
	offX, offY := transCurve2FrameOffset()

	if transCurve2Animate {
		for i := 0; i < 6; i++ {
			moveTransCurve2Point(&transCurve2Points1[i*2], &transCurve2Points1[i*2+1], &transCurve2DX1[i], &transCurve2DY1[i])
			moveTransCurve2Point(&transCurve2Points2[i*2], &transCurve2Points2[i*2+1], &transCurve2DX2[i], &transCurve2DY2[i])
			d := math.Sqrt((transCurve2Points1[i*2]-transCurve2Points2[i*2])*(transCurve2Points1[i*2]-transCurve2Points2[i*2]) + (transCurve2Points1[i*2+1]-transCurve2Points2[i*2+1])*(transCurve2Points1[i*2+1]-transCurve2Points2[i*2+1]))
			if d > 28.28 {
				transCurve2Points2[i*2] = transCurve2Points1[i*2] + (transCurve2Points2[i*2]-transCurve2Points1[i*2])*28.28/d
				transCurve2Points2[i*2+1] = transCurve2Points1[i*2+1] + (transCurve2Points2[i*2+1]-transCurve2Points1[i*2+1])*28.28/d
			}
		}
	}

	ctx.Clear(agg.White)
	a := ctx.GetAgg2D()
	a.ResetTransformations()

	// 1. Build guide B-splines from the two control-point polygons.
	poly1 := polygon.NewSimplePolygonVertexSource(transCurve2Points1[:], 6, false, false)
	poly2 := polygon.NewSimplePolygonVertexSource(transCurve2Points2[:], 6, false, false)

	bs1 := conv.NewConvBSpline(poly1)
	bs2 := conv.NewConvBSpline(poly2)
	bs1.SetInterpolationStep(1.0 / transCurve2NumPoints)
	bs2.SetInterpolationStep(1.0 / transCurve2NumPoints)

	// 2. Set up the double-path transformer.
	tcurve := transform.NewTransDoublePath()
	tcurve.SetPreserveXScale(transCurve2PreserveXScale)
	if transCurve2FixedLen {
		tcurve.SetBaseLength(transCurve2BaseLen)
	} else {
		tcurve.SetBaseLength(0)
	}
	tcurve.SetBaseHeight(30.0)
	tcurve.AddPaths(&transDoubleAdapter{bs1}, &transDoubleAdapter{bs2}, 0, 0)

	// 3. Render text transformed along the double path.
	text := gsv.NewGSVText()
	text.SetFlip(true)
	text.SetSize(40.0, 0)
	text.SetStartPoint(0, 3)
	text.SetText(transCurve2Text)

	outline := gsv.NewGSVTextOutline(text)
	outline.SetWidth(1.0)

	segm := conv.NewConvSegmentator(outline)
	segm.ApproximationScale(3.0)

	transformedText := conv.NewConvTransform(
		&segmentatorAdapter2{source: segm},
		tcurve,
	)

	a.FillColor(agg.Black)
	a.NoLine()
	transCurve2AppendPath(a, transformedText, offX, offY)
	a.DrawPath(agg.FillOnly)

	// 4. Draw guide curves.
	a.LineColor(agg.NewColor(170, 50, 20, 100))
	a.LineWidth(2.0)
	a.NoFill()
	transCurve2AppendPath(a, bs1, offX, offY)
	a.DrawPath(agg.StrokeOnly)
	transCurve2AppendPath(a, bs2, offX, offY)
	a.DrawPath(agg.StrokeOnly)

	// 5. Draw polygon guide lines.
	a.LineColor(agg.NewColor(0, 76, 128, 120))
	a.LineWidth(1.0)
	a.NoFill()
	transCurve2AppendPath(a, poly1, offX, offY)
	a.DrawPath(agg.StrokeOnly)
	transCurve2AppendPath(a, poly2, offX, offY)
	a.DrawPath(agg.StrokeOnly)

	// 6. Draw interactive handles.
	for i := 0; i < 6; i++ {
		drawHandle(transCurve2Points1[i*2]+offX, transCurve2Points1[i*2+1]+offY)
		drawHandle(transCurve2Points2[i*2]+offX, transCurve2Points2[i*2+1]+offY)
	}
}

// transCurve2AppendPath feeds vertices from src into the Agg2D path builder.
func transCurve2AppendPath(a *agg.Agg2D, src conv.VertexSource, offsetX, offsetY float64) {
	a.ResetPath()
	src.Rewind(0)
	for {
		x, y, cmd := src.Vertex()
		switch {
		case basics.IsStop(cmd):
			return
		case basics.IsMoveTo(cmd):
			a.MoveTo(x+offsetX, y+offsetY)
		case basics.IsLineTo(cmd):
			a.LineTo(x+offsetX, y+offsetY)
		case basics.IsEndPoly(cmd):
			if basics.IsClosed(uint32(cmd)) {
				a.ClosePolygon()
			}
		}
	}
}

// segmentatorAdapter2 bridges ConvSegmentator to the conv.VertexSourceCommander
// interface required by conv.NewConvTransform.
type segmentatorAdapter2 struct {
	source *conv.ConvSegmentator
}

func (a *segmentatorAdapter2) Rewind(pathID uint) { a.source.Rewind(pathID) }
func (a *segmentatorAdapter2) Vertex() (x, y float64, cmd basics.PathCommand) {
	x, y, raw := a.source.Vertex()
	return x, y, basics.PathCommand(raw)
}

func moveTransCurve2Point(x, y, dx, dy *float64) {
	*x += *dx
	*y += *dy
	if *x < 0 || *x > transCurve2RefW {
		*dx = -*dx
	}
	if *y < 0 || *y > transCurve2RefH {
		*dy = -*dy
	}
}

func handleTransCurve2MouseDown(x, y float64) bool {
	offX, offY := transCurve2FrameOffset()
	x -= offX
	y -= offY
	transCurve2Selected = -1
	for i := 0; i < 6; i++ {
		if math.Sqrt((x-transCurve2Points1[i*2])*(x-transCurve2Points1[i*2])+(y-transCurve2Points1[i*2+1])*(y-transCurve2Points1[i*2+1])) < 15 {
			transCurve2Selected = i
			return true
		}
		if math.Sqrt((x-transCurve2Points2[i*2])*(x-transCurve2Points2[i*2])+(y-transCurve2Points2[i*2+1])*(y-transCurve2Points2[i*2+1])) < 15 {
			transCurve2Selected = i + 6
			return true
		}
	}
	return false
}

func handleTransCurve2MouseMove(x, y float64) bool {
	offX, offY := transCurve2FrameOffset()
	x -= offX
	y -= offY
	if transCurve2Selected != -1 {
		if transCurve2Selected < 6 {
			transCurve2Points1[transCurve2Selected*2] = x
			transCurve2Points1[transCurve2Selected*2+1] = y
		} else {
			idx := transCurve2Selected - 6
			transCurve2Points2[idx*2] = x
			transCurve2Points2[idx*2+1] = y
		}
		return true
	}
	return false
}

func handleTransCurve2MouseUp() {
	transCurve2Selected = -1
}

func toggleTransCurve2Animate() {
	transCurve2Animate = !transCurve2Animate
}

func setTransCurve2NumPoints(v float64) {
	transCurve2NumPoints = v
}

func setTransCurve2FixedLen(v bool) {
	transCurve2FixedLen = v
}

func setTransCurve2PreserveXScale(v bool) {
	transCurve2PreserveXScale = v
}
