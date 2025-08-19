package conv

import (
	"agg_go/internal/basics"
	"agg_go/internal/vcgen"
)

// ConvStroke converts paths to stroked outlines using the adaptor pattern
type ConvStroke struct {
	*ConvAdaptorVCGen
	strokeGen *vcgen.VCGenStroke
}

// NewConvStroke creates a new stroke converter
func NewConvStroke(source VertexSource) *ConvStroke {
	strokeGen := vcgen.NewVCGenStroke()
	adaptor := NewConvAdaptorVCGen(source, strokeGen)

	return &ConvStroke{
		ConvAdaptorVCGen: adaptor,
		strokeGen:        strokeGen,
	}
}

// NewConvStrokeWithMarkers creates a new stroke converter with markers
func NewConvStrokeWithMarkers(source VertexSource, markers Markers) *ConvStroke {
	strokeGen := vcgen.NewVCGenStroke()
	adaptor := NewConvAdaptorVCGenWithMarkers(source, strokeGen, markers)

	return &ConvStroke{
		ConvAdaptorVCGen: adaptor,
		strokeGen:        strokeGen,
	}
}

// SetLineCap sets the line cap style
func (cs *ConvStroke) SetLineCap(lc basics.LineCap) {
	cs.strokeGen.SetLineCap(lc)
}

// LineCap returns the current line cap style
func (cs *ConvStroke) LineCap() basics.LineCap {
	return cs.strokeGen.LineCap()
}

// SetLineJoin sets the line join style
func (cs *ConvStroke) SetLineJoin(lj basics.LineJoin) {
	cs.strokeGen.SetLineJoin(lj)
}

// LineJoin returns the current line join style
func (cs *ConvStroke) LineJoin() basics.LineJoin {
	return cs.strokeGen.LineJoin()
}

// SetInnerJoin sets the inner join style
func (cs *ConvStroke) SetInnerJoin(ij basics.InnerJoin) {
	cs.strokeGen.SetInnerJoin(ij)
}

// InnerJoin returns the current inner join style
func (cs *ConvStroke) InnerJoin() basics.InnerJoin {
	return cs.strokeGen.InnerJoin()
}

// SetWidth sets the stroke width
func (cs *ConvStroke) SetWidth(w float64) {
	cs.strokeGen.SetWidth(w)
}

// Width returns the current stroke width
func (cs *ConvStroke) Width() float64 {
	return cs.strokeGen.Width()
}

// SetMiterLimit sets the miter limit
func (cs *ConvStroke) SetMiterLimit(ml float64) {
	cs.strokeGen.SetMiterLimit(ml)
}

// MiterLimit returns the current miter limit
func (cs *ConvStroke) MiterLimit() float64 {
	return cs.strokeGen.MiterLimit()
}

// SetMiterLimitTheta sets the miter limit from an angle in radians
func (cs *ConvStroke) SetMiterLimitTheta(t float64) {
	cs.strokeGen.SetMiterLimitTheta(t)
}

// SetInnerMiterLimit sets the inner miter limit
func (cs *ConvStroke) SetInnerMiterLimit(ml float64) {
	cs.strokeGen.SetInnerMiterLimit(ml)
}

// InnerMiterLimit returns the current inner miter limit
func (cs *ConvStroke) InnerMiterLimit() float64 {
	return cs.strokeGen.InnerMiterLimit()
}

// SetApproximationScale sets the approximation scale for curves
func (cs *ConvStroke) SetApproximationScale(as float64) {
	cs.strokeGen.SetApproximationScale(as)
}

// ApproximationScale returns the current approximation scale
func (cs *ConvStroke) ApproximationScale() float64 {
	return cs.strokeGen.ApproximationScale()
}

// SetShorten sets the path shortening amount
func (cs *ConvStroke) SetShorten(s float64) {
	cs.strokeGen.SetShorten(s)
}

// Shorten returns the current path shortening amount
func (cs *ConvStroke) Shorten() float64 {
	return cs.strokeGen.Shorten()
}

// Generator returns the underlying stroke generator for advanced access
func (cs *ConvStroke) Generator() *vcgen.VCGenStroke {
	return cs.strokeGen
}
