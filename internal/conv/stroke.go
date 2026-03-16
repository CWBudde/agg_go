package conv

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/vcgen"
)

// ConvStroke is the Go equivalent of AGG's conv_stroke. It wraps VCGenStroke
// behind ConvAdaptorVCGen so any vertex source can be expanded into a stroked
// outline.
type ConvStroke struct {
	*ConvAdaptorVCGen
	strokeGen *vcgen.VCGenStroke
}

// NewConvStroke creates a stroke converter without terminal markers.
func NewConvStroke(source VertexSource) *ConvStroke {
	strokeGen := vcgen.NewVCGenStroke()
	adaptor := NewConvAdaptorVCGen(source, strokeGen)

	return &ConvStroke{
		ConvAdaptorVCGen: adaptor,
		strokeGen:        strokeGen,
	}
}

// NewConvStrokeWithMarkers creates a stroke converter with terminal-marker
// support.
func NewConvStrokeWithMarkers(source VertexSource, markers Markers) *ConvStroke {
	strokeGen := vcgen.NewVCGenStroke()
	adaptor := NewConvAdaptorVCGenWithMarkers(source, strokeGen, markers)

	return &ConvStroke{
		ConvAdaptorVCGen: adaptor,
		strokeGen:        strokeGen,
	}
}

// SetLineCap sets the cap style used on open-path ends.
func (cs *ConvStroke) SetLineCap(lc basics.LineCap) {
	cs.strokeGen.SetLineCap(lc)
}

// LineCap returns the current cap style.
func (cs *ConvStroke) LineCap() basics.LineCap {
	return cs.strokeGen.LineCap()
}

// SetLineJoin sets the outer join style.
func (cs *ConvStroke) SetLineJoin(lj basics.LineJoin) {
	cs.strokeGen.SetLineJoin(lj)
}

// LineJoin returns the current outer join style.
func (cs *ConvStroke) LineJoin() basics.LineJoin {
	return cs.strokeGen.LineJoin()
}

// SetInnerJoin sets the inner join style used on self-intersections and narrow
// turns.
func (cs *ConvStroke) SetInnerJoin(ij basics.InnerJoin) {
	cs.strokeGen.SetInnerJoin(ij)
}

// InnerJoin returns the current inner join style.
func (cs *ConvStroke) InnerJoin() basics.InnerJoin {
	return cs.strokeGen.InnerJoin()
}

// SetWidth sets the stroke width in source-space units.
func (cs *ConvStroke) SetWidth(w float64) {
	cs.strokeGen.SetWidth(w)
}

// Width returns the current stroke width.
func (cs *ConvStroke) Width() float64 {
	return cs.strokeGen.Width()
}

// SetMiterLimit sets the outer miter limit.
func (cs *ConvStroke) SetMiterLimit(ml float64) {
	cs.strokeGen.SetMiterLimit(ml)
}

// MiterLimit returns the current outer miter limit.
func (cs *ConvStroke) MiterLimit() float64 {
	return cs.strokeGen.MiterLimit()
}

// SetMiterLimitTheta derives the outer miter limit from an angle in radians.
func (cs *ConvStroke) SetMiterLimitTheta(t float64) {
	cs.strokeGen.SetMiterLimitTheta(t)
}

// SetInnerMiterLimit sets the inner miter limit.
func (cs *ConvStroke) SetInnerMiterLimit(ml float64) {
	cs.strokeGen.SetInnerMiterLimit(ml)
}

// InnerMiterLimit returns the current inner miter limit.
func (cs *ConvStroke) InnerMiterLimit() float64 {
	return cs.strokeGen.InnerMiterLimit()
}

// SetApproximationScale controls how finely round joins and caps are tessellated.
func (cs *ConvStroke) SetApproximationScale(as float64) {
	cs.strokeGen.SetApproximationScale(as)
}

// ApproximationScale returns the current curve approximation scale.
func (cs *ConvStroke) ApproximationScale() float64 {
	return cs.strokeGen.ApproximationScale()
}

// SetShorten trims both ends of open paths before stroking.
func (cs *ConvStroke) SetShorten(s float64) {
	cs.strokeGen.SetShorten(s)
}

// Shorten returns the current end-trimming amount.
func (cs *ConvStroke) Shorten() float64 {
	return cs.strokeGen.Shorten()
}

// Generator returns the underlying VCGenStroke for advanced tuning.
func (cs *ConvStroke) Generator() *vcgen.VCGenStroke {
	return cs.strokeGen
}
