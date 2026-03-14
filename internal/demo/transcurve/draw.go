// Package transcurve provides a Go-idiomatic equivalent of AGG's trans_curve1.cpp.
//
// It uses the embedded GSV vector font rather than a platform FreeType backend,
// but preserves the core demo behavior: text transformed along a B-spline path
// with configurable subdivision, closure, X-scale preservation, fixed base
// length, and animated control points.
package transcurve

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/polygon"
	"github.com/MeKo-Christian/agg_go/internal/gsv"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

const (
	ControlPointCount = 6
	DefaultBaseLength = 1120.0
	DefaultTextHeight = 40.0
)

var DefaultPoints = [ControlPointCount * 2]float64{
	50, 50,
	170, 130,
	230, 270,
	370, 330,
	430, 470,
	550, 550,
}

const DefaultText = "Anti-Grain Geometry is designed as a set of loosely coupled algorithms and class templates united with a common idea, so that all the components can be easily combined. Also, the template based design allows you to replace any part of the library without the necessity to modify a single byte in the existing code."

type AnimationState struct {
	DX [ControlPointCount]float64
	DY [ControlPointCount]float64
}

type Config struct {
	Points           [ControlPointCount * 2]float64
	NumIntermediate  float64
	Close            bool
	PreserveXScale   bool
	FixedLength      bool
	BaseLength       float64
	Text             string
	OffsetX          float64
	OffsetY          float64
	TextHeight       float64
	TextStrokeWidth  float64
}

func NewAnimationState() AnimationState {
	var anim AnimationState
	for i := 0; i < ControlPointCount; i++ {
		anim.DX[i] = (math.Mod(float64(i*1234), 10.0) - 5.0) * 0.5
		anim.DY[i] = (math.Mod(float64(i*5678), 10.0) - 5.0) * 0.5
	}
	return anim
}

func AnimatePoints(points *[ControlPointCount * 2]float64, anim *AnimationState, width, height float64) {
	for i := 0; i < ControlPointCount; i++ {
		points[i*2] += anim.DX[i]
		points[i*2+1] += anim.DY[i]
		if points[i*2] < 0 || points[i*2] > width {
			anim.DX[i] = -anim.DX[i]
			points[i*2] += anim.DX[i]
		}
		if points[i*2+1] < 0 || points[i*2+1] > height {
			anim.DY[i] = -anim.DY[i]
			points[i*2+1] += anim.DY[i]
		}
	}
}

func Draw(ctx *agg.Context, cfg Config) {
	ctx.Clear(agg.White)

	a := ctx.GetAgg2D()
	a.ResetTransformations()

	if cfg.NumIntermediate < 1 {
		cfg.NumIntermediate = 1
	}
	if cfg.BaseLength <= 0 {
		cfg.BaseLength = DefaultBaseLength
	}
	if cfg.Text == "" {
		cfg.Text = DefaultText
	}
	if cfg.TextHeight <= 0 {
		cfg.TextHeight = DefaultTextHeight
	}
	if cfg.TextStrokeWidth <= 0 {
		cfg.TextStrokeWidth = 1.0
	}

	poly := polygon.NewSimplePolygonVertexSource(cfg.Points[:], ControlPointCount, false, cfg.Close)
	bspline := conv.NewConvBSpline(poly)
	bspline.SetInterpolationStep(1.0 / cfg.NumIntermediate)

	tcurve := transform.NewTransSinglePath()
	tcurve.SetPreserveXScale(cfg.PreserveXScale)
	if cfg.FixedLength {
		tcurve.SetBaseLength(cfg.BaseLength)
	} else {
		tcurve.SetBaseLength(0)
	}
	tcurve.AddPath(bspline, 0)

	text := gsv.NewGSVText()
	text.SetFlip(true)
	text.SetSize(cfg.TextHeight, 0)
	text.SetStartPoint(0, 3)
	text.SetText(cfg.Text)

	outline := gsv.NewGSVTextOutline(text)
	outline.SetWidth(cfg.TextStrokeWidth)

	segm := conv.NewConvSegmentator(outline)
	segm.ApproximationScale(3.0)

	transformedText := conv.NewConvTransform(segm, tcurve)

	a.FillColor(agg.Black)
	a.NoLine()
	if appendPath(a, transformedText, cfg.OffsetX, cfg.OffsetY) {
		a.DrawPath(agg.FillOnly)
	}

	a.LineColor(agg.NewColor(170, 50, 20, 100))
	a.LineWidth(2.0)
	a.NoFill()
	if appendPath(a, bspline, cfg.OffsetX, cfg.OffsetY) {
		a.DrawPath(agg.StrokeOnly)
	}

	a.LineColor(agg.NewColor(0, 76, 128, 120))
	a.LineWidth(1.0)
	a.NoFill()
	if appendPath(a, poly, cfg.OffsetX, cfg.OffsetY) {
		a.DrawPath(agg.StrokeOnly)
	}

	for i := 0; i < ControlPointCount; i++ {
		drawHandle(ctx, cfg.Points[i*2]+cfg.OffsetX, cfg.Points[i*2+1]+cfg.OffsetY)
	}
}

func appendPath(a *agg.Agg2D, src conv.VertexSource, offsetX, offsetY float64) bool {
	a.ResetPath()
	src.Rewind(0)

	hasVertices := false
	for {
		x, y, cmd := src.Vertex()
		switch {
		case basics.IsStop(cmd):
			return hasVertices
		case basics.IsMoveTo(cmd):
			a.MoveTo(x+offsetX, y+offsetY)
			hasVertices = true
		case basics.IsLineTo(cmd):
			a.LineTo(x+offsetX, y+offsetY)
			hasVertices = true
		case basics.IsEndPoly(cmd):
			if basics.IsClosed(uint32(cmd)) {
				a.ClosePolygon()
			}
		}
	}
}

func drawHandle(ctx *agg.Context, x, y float64) {
	ctx.SetColor(agg.RGBA(0.8, 0.2, 0.1, 0.6))
	ctx.FillCircle(x, y, 5)
	ctx.SetColor(agg.Black)
	ctx.DrawCircle(x, y, 5)
}
