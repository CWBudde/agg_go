// Go-idiomatic equivalent of AGG's trans_curve1_ft.cpp.
//
// This variant uses the FreeType outline backend when available. If the
// `freetype` build tag is not enabled or no suitable italic serif font is found,
// it still renders the guide curve and shows a fallback note.
package main

import (
	"os"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/polygon"
	"github.com/MeKo-Christian/agg_go/internal/demo/transcurve"
	"github.com/MeKo-Christian/agg_go/internal/font"
	"github.com/MeKo-Christian/agg_go/internal/font/freetype"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

const (
	width  = 600
	height = 600
)

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	ctx.Clear(agg.White)

	a := ctx.GetAgg2D()
	a.ResetTransformations()

	tcurve, bspline, poly := buildCurve()
	drawTextAlongCurve(a, tcurve)
	drawCurve(ctx, a, bspline, poly)
}

func buildCurve() (*transform.TransSinglePath, *conv.ConvBSpline, *polygon.SimplePolygonVertexSource) {
	poly := polygon.NewSimplePolygonVertexSource(
		transcurve.DefaultPoints[:],
		transcurve.ControlPointCount,
		false,
		false,
	)

	bspline := conv.NewConvBSpline(poly)
	bspline.SetInterpolationStep(1.0 / 200.0)

	tcurve := transform.NewTransSinglePath()
	tcurve.SetPreserveXScale(true)
	tcurve.SetBaseLength(transcurve.DefaultBaseLength)
	tcurve.AddPath(bspline, 0)

	return tcurve, bspline, poly
}

func drawTextAlongCurve(a *agg.Agg2D, tcurve *transform.TransSinglePath) {
	fontPath := findFreetypeFont()
	engine, err := freetype.NewFontEngineFreetype(false, 32)
	if err != nil || fontPath == "" {
		drawFallbackMessage(a, err, fontPath == "")
		return
	}
	defer func() { _ = engine.Close() }()

	if err := engine.LoadFont(fontPath, 0, freetype.GlyphRenderingOutline, nil); err != nil {
		drawFallbackMessage(a, err, false)
		return
	}

	engine.SetHinting(false)
	engine.SetFlipY(true)
	engine.SetHeight(transcurve.DefaultTextHeight)

	fcm := font.NewFontCacheManager(engine, 32)

	x, y := 0.0, 3.0
	firstGlyph := true
	var prevGlyphIndex uint

	for _, r := range transcurve.DefaultText {
		glyph := fcm.Glyph(uint(r))
		if glyph == nil {
			continue
		}
		if x > tcurve.TotalLength() {
			break
		}

		if !firstGlyph {
			fcm.AddKerning(&x, &y, prevGlyphIndex, glyph.GlyphIndex)
		}
		fcm.InitEmbeddedAdaptors(glyph, x, y)

		if glyph.DataType == font.GlyphDataOutline {
			curves := conv.NewConvCurve(path.NewPathStorageStlVertexSourceAdapter(fcm.PathAdaptor()))
			curves.SetApproximationScale(2.0)

			segm := conv.NewConvSegmentator(curves)
			segm.ApproximationScale(3.0)

			transformed := conv.NewConvTransform(&segmentatorAdapter{source: segm}, tcurve)
			a.FillColor(agg.Black)
			a.NoLine()
			if appendPath(a, transformed) {
				a.DrawPath(agg.FillOnly)
			}
		}

		x += glyph.AdvanceX
		y += glyph.AdvanceY
		prevGlyphIndex = glyph.GlyphIndex
		firstGlyph = false
	}
}

func drawFallbackMessage(a *agg.Agg2D, fontErr error, missingFont bool) {
	a.FillColor(agg.Black)
	a.NoLine()
	a.FontGSV(10)
	a.Text(14, 20, "trans_curve1_ft: FreeType outline text unavailable", false, 0, 0)
	if missingFont {
		a.Text(14, 36, "No italic serif TTF found in common system locations.", false, 0, 0)
		return
	}
	if fontErr != nil {
		a.Text(14, 36, fontErr.Error(), false, 0, 0)
	}
}

func drawCurve(ctx *agg.Context, a *agg.Agg2D, bspline *conv.ConvBSpline, poly *polygon.SimplePolygonVertexSource) {
	a.LineColor(agg.NewColor(170, 50, 20, 100))
	a.LineWidth(2.0)
	a.NoFill()
	if appendPath(a, bspline) {
		a.DrawPath(agg.StrokeOnly)
	}

	a.LineColor(agg.NewColor(0, 76, 128, 120))
	a.LineWidth(1.0)
	a.NoFill()
	if appendPath(a, poly) {
		a.DrawPath(agg.StrokeOnly)
	}

	for i := 0; i < transcurve.ControlPointCount; i++ {
		drawHandle(ctx, transcurve.DefaultPoints[i*2], transcurve.DefaultPoints[i*2+1])
	}
}

func appendPath(a *agg.Agg2D, src conv.VertexSource) bool {
	a.ResetPath()
	src.Rewind(0)

	hasVertices := false
	for {
		x, y, cmd := src.Vertex()
		switch {
		case basics.IsStop(cmd):
			return hasVertices
		case basics.IsMoveTo(cmd):
			a.MoveTo(x, y)
			hasVertices = true
		case basics.IsLineTo(cmd):
			a.LineTo(x, y)
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

func findFreetypeFont() string {
	candidates := []string{
		"/usr/share/fonts/truetype/dejavu/DejaVuSerif-Italic.ttf",
		"/usr/share/fonts/TTF/DejaVuSerif-Italic.ttf",
		"/usr/share/fonts/truetype/liberation2/LiberationSerif-Italic.ttf",
		"/usr/share/fonts/truetype/liberation/LiberationSerif-Italic.ttf",
		"/usr/share/fonts/liberation/LiberationSerif-Italic.ttf",
		"/usr/share/fonts/truetype/noto/NotoSerif-Italic.ttf",
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Trans Curve 1 FreeType",
		Width:  width,
		Height: height,
	}, &demo{})
}
