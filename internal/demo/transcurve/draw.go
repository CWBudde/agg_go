// Package transcurve provides Go-idiomatic equivalents of AGG's trans_curve1.cpp
// (Draw, single B-spline rail) and trans_curve2.cpp (DrawDouble, two rails).
//
// Both use the embedded GSV vector font rather than a platform TrueType/FreeType
// backend, but preserve the core demo behavior: a text paragraph transformed
// along the rail(s) with configurable subdivision, X-scale preservation, fixed
// base length/height, and animated control points. The functions are shared
// verbatim by the standalone examples and the WASM demos to keep standalone/web
// output identical.
//
// C++ references: ../agg-2.6/agg-src/examples/trans_curve1.cpp and
// ../agg-2.6/agg-src/examples/trans_curve2.cpp.
package transcurve

import (
	"math"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/conv"
	"github.com/cwbudde/agg_go/internal/ctrl/polygon"
	"github.com/cwbudde/agg_go/internal/gsv"
	"github.com/cwbudde/agg_go/internal/transform"
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

// Double-path defaults mirror AGG's trans_curve2.cpp: two B-spline "rails"
// (6 control points each), base_length(1140), base_height(30), and the trailing
// space that the C++ text literal carries.
const (
	DefaultDoubleBaseLength = 1140.0
	DefaultDoubleBaseHeight = 30.0
	DefaultDoubleText       = DefaultText + " "
)

var (
	DefaultPoints1 = [ControlPointCount * 2]float64{60, 40, 170, 130, 230, 270, 370, 330, 430, 470, 550, 550}
	DefaultPoints2 = [ControlPointCount * 2]float64{40, 60, 150, 170, 210, 290, 350, 350, 410, 490, 530, 570}
)

type AnimationState struct {
	DX [ControlPointCount]float64
	DY [ControlPointCount]float64
}

type Config struct {
	Points          [ControlPointCount * 2]float64
	NumIntermediate float64
	Close           bool
	PreserveXScale  bool
	FixedLength     bool
	BaseLength      float64
	Text            string
	OffsetX         float64
	OffsetY         float64
	TextHeight      float64
	TextStrokeWidth float64
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

	transformedText := conv.NewConvTransform(&segmentatorAdapter{source: segm}, tcurve)

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

// DoubleConfig drives DrawDouble, the faithful port of AGG's trans_curve2.cpp:
// a text paragraph warped between two interactive B-spline rails via
// transform.TransDoublePath. It is the two-rail counterpart of Config.
type DoubleConfig struct {
	Points1         [ControlPointCount * 2]float64
	Points2         [ControlPointCount * 2]float64
	NumIntermediate float64
	PreserveXScale  bool
	FixedLength     bool
	BaseLength      float64
	BaseHeight      float64
	Text            string
	OffsetX         float64
	OffsetY         float64
	TextHeight      float64
	TextStrokeWidth float64
}

// DrawDouble renders the double-path (two-rail) transform demo. It mirrors
// AGG's trans_curve2.cpp on_draw(): the GSV vector-font paragraph is segmented
// and transformed between two B-spline rails, then the rails, control polygons,
// and drag handles are drawn on top.
//
// C++ reference: ../agg-2.6/agg-src/examples/trans_curve2.cpp.
func DrawDouble(ctx *agg.Context, cfg DoubleConfig) {
	ctx.Clear(agg.White)

	a := ctx.GetAgg2D()
	a.ResetTransformations()

	if cfg.NumIntermediate < 1 {
		cfg.NumIntermediate = 200
	}
	if cfg.BaseLength <= 0 {
		cfg.BaseLength = DefaultDoubleBaseLength
	}
	if cfg.BaseHeight <= 0 {
		cfg.BaseHeight = DefaultDoubleBaseHeight
	}
	if cfg.Text == "" {
		cfg.Text = DefaultDoubleText
	}
	if cfg.TextHeight <= 0 {
		cfg.TextHeight = DefaultTextHeight
	}
	if cfg.TextStrokeWidth <= 0 {
		cfg.TextStrokeWidth = 1.0
	}

	poly1 := polygon.NewSimplePolygonVertexSource(cfg.Points1[:], ControlPointCount, false, false)
	poly2 := polygon.NewSimplePolygonVertexSource(cfg.Points2[:], ControlPointCount, false, false)

	bspline1 := conv.NewConvBSpline(poly1)
	bspline2 := conv.NewConvBSpline(poly2)
	bspline1.SetInterpolationStep(1.0 / cfg.NumIntermediate)
	bspline2.SetInterpolationStep(1.0 / cfg.NumIntermediate)

	tcurve := transform.NewTransDoublePath()
	tcurve.SetPreserveXScale(cfg.PreserveXScale)
	if cfg.FixedLength {
		tcurve.SetBaseLength(cfg.BaseLength)
	} else {
		tcurve.SetBaseLength(0)
	}
	tcurve.SetBaseHeight(cfg.BaseHeight)
	tcurve.AddPaths(bspline1, bspline2, 0, 0)

	text := gsv.NewGSVText()
	text.SetFlip(true)
	text.SetSize(cfg.TextHeight, 0)
	text.SetStartPoint(0, 3)
	text.SetText(cfg.Text)

	outline := gsv.NewGSVTextOutline(text)
	outline.SetWidth(cfg.TextStrokeWidth)

	segm := conv.NewConvSegmentator(outline)
	segm.ApproximationScale(3.0)

	transformedText := conv.NewConvTransform(&segmentatorAdapter{source: segm}, tcurve)

	a.FillColor(agg.Black)
	a.NoLine()
	if appendPath(a, transformedText, cfg.OffsetX, cfg.OffsetY) {
		a.DrawPath(agg.FillOnly)
	}

	a.LineColor(agg.NewColor(170, 50, 20, 100))
	a.LineWidth(2.0)
	a.NoFill()
	if appendPath(a, bspline1, cfg.OffsetX, cfg.OffsetY) {
		a.DrawPath(agg.StrokeOnly)
	}
	if appendPath(a, bspline2, cfg.OffsetX, cfg.OffsetY) {
		a.DrawPath(agg.StrokeOnly)
	}

	a.LineColor(agg.NewColor(0, 76, 128, 120))
	a.LineWidth(1.0)
	a.NoFill()
	if appendPath(a, poly1, cfg.OffsetX, cfg.OffsetY) {
		a.DrawPath(agg.StrokeOnly)
	}
	if appendPath(a, poly2, cfg.OffsetX, cfg.OffsetY) {
		a.DrawPath(agg.StrokeOnly)
	}

	for i := 0; i < ControlPointCount; i++ {
		drawHandle(ctx, cfg.Points1[i*2]+cfg.OffsetX, cfg.Points1[i*2+1]+cfg.OffsetY)
		drawHandle(ctx, cfg.Points2[i*2]+cfg.OffsetX, cfg.Points2[i*2+1]+cfg.OffsetY)
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

type segmentatorAdapter struct {
	source *conv.ConvSegmentator
}

func (a *segmentatorAdapter) Rewind(pathID uint) {
	a.source.Rewind(pathID)
}

func (a *segmentatorAdapter) Vertex() (x, y float64, cmd basics.PathCommand) {
	x, y, raw := a.source.Vertex()
	return x, y, basics.PathCommand(raw)
}
