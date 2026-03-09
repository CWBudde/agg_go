package interactivepolygon

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/polygon"
)

// State stores the interactive polygon control state across frames.
type State struct {
	poly *polygon.PolygonCtrl[color.RGBA]
}

// NewState creates a 5-point interactive polygon initialized relative to canvas size.
func NewState(canvasW, canvasH float64) *State {
	p := polygon.NewDefaultPolygonCtrl(5, 6.0)
	p.SetLineWidth(1.5)
	p.SetClose(true)
	p.SetInPolygonCheck(true)

	p.SetXn(0, canvasW*0.18)
	p.SetYn(0, canvasH*0.20)
	p.SetXn(1, canvasW*0.78)
	p.SetYn(1, canvasH*0.18)
	p.SetXn(2, canvasW*0.84)
	p.SetYn(2, canvasH*0.58)
	p.SetXn(3, canvasW*0.52)
	p.SetYn(3, canvasH*0.84)
	p.SetXn(4, canvasW*0.20)
	p.SetYn(4, canvasH*0.72)

	return &State{poly: p}
}

func (s *State) MouseDown(x, y float64) bool {
	if s == nil || s.poly == nil {
		return false
	}
	return s.poly.OnMouseButtonDown(x, y)
}

func (s *State) MouseMove(x, y float64, leftPressed bool) bool {
	if s == nil || s.poly == nil {
		return false
	}
	return s.poly.OnMouseMove(x, y, leftPressed)
}

func (s *State) MouseUp(x, y float64) bool {
	if s == nil || s.poly == nil {
		return false
	}
	return s.poly.OnMouseButtonUp(x, y)
}

func (s *State) Draw(ctx *agg.Context) {
	if s == nil || s.poly == nil {
		return
	}

	a := ctx.GetAgg2D()
	a.ResetTransformations()
	ctx.Clear(agg.White)

	// Polygon fill and outline.
	a.ResetPath()
	a.MoveTo(s.poly.Xn(0), s.poly.Yn(0))
	for i := uint(1); i < s.poly.NumPoints(); i++ {
		a.LineTo(s.poly.Xn(i), s.poly.Yn(i))
	}
	a.ClosePolygon()
	a.FillColor(agg.RGBA(0.12, 0.45, 0.80, 0.18))
	a.LineColor(agg.RGBA(0.05, 0.25, 0.45, 0.95))
	a.LineWidth(s.poly.LineWidth())
	a.DrawPath(agg.FillAndStroke)

	// Vertex handles.
	r := s.poly.PointRadius()
	for i := uint(0); i < s.poly.NumPoints(); i++ {
		x, y := s.poly.Xn(i), s.poly.Yn(i)
		ctx.SetColor(agg.RGBA(0.85, 0.2, 0.15, 0.80))
		ctx.FillCircle(x, y, r)
		ctx.SetColor(agg.Black)
		ctx.SetLineWidth(1.0)
		ctx.DrawCircle(x, y, r)
	}
}
