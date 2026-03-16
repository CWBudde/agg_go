package conv

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/vcgen"
)

// ConvContour is the Go equivalent of AGG's conv_contour. It wraps
// VCGenContour to generate one offset contour from the source path instead of a
// full stroked outline.
type ConvContour struct {
	*ConvAdaptorVCGen
	generator *vcgen.VCGenContour
}

// NewConvContour creates a contour converter.
func NewConvContour(source VertexSource) *ConvContour {
	generator := vcgen.NewVCGenContour()
	adaptor := NewConvAdaptorVCGen(source, generator)

	return &ConvContour{
		ConvAdaptorVCGen: adaptor,
		generator:        generator,
	}
}

// Generator returns the underlying VCGenContour.
func (c *ConvContour) Generator() *vcgen.VCGenContour {
	return c.generator
}

// Width sets the contour offset distance.
func (c *ConvContour) Width(w float64) {
	c.generator.Width(w)
}

// GetWidth returns the current contour offset distance.
func (c *ConvContour) GetWidth() float64 {
	return c.generator.GetWidth()
}

// LineJoin sets the join style used at contour corners.
func (c *ConvContour) LineJoin(lj basics.LineJoin) {
	c.generator.LineJoin(lj)
}

// GetLineJoin returns the current contour join style.
func (c *ConvContour) GetLineJoin() basics.LineJoin {
	return c.generator.GetLineJoin()
}

// InnerJoin sets the fallback style for problematic inner corners.
func (c *ConvContour) InnerJoin(ij basics.InnerJoin) {
	c.generator.InnerJoin(ij)
}

// GetInnerJoin returns the current inner join style.
func (c *ConvContour) GetInnerJoin() basics.InnerJoin {
	return c.generator.GetInnerJoin()
}

// MiterLimit sets the outer miter limit.
func (c *ConvContour) MiterLimit(ml float64) {
	c.generator.MiterLimit(ml)
}

// GetMiterLimit returns the current outer miter limit.
func (c *ConvContour) GetMiterLimit() float64 {
	return c.generator.GetMiterLimit()
}

// MiterLimitTheta derives the outer miter limit from an angle in radians.
func (c *ConvContour) MiterLimitTheta(t float64) {
	c.generator.MiterLimitTheta(t)
}

// InnerMiterLimit sets the inner miter limit.
func (c *ConvContour) InnerMiterLimit(ml float64) {
	c.generator.InnerMiterLimit(ml)
}

// GetInnerMiterLimit returns the current inner miter limit.
func (c *ConvContour) GetInnerMiterLimit() float64 {
	return c.generator.GetInnerMiterLimit()
}

// ApproximationScale controls how finely round joins are tessellated.
func (c *ConvContour) ApproximationScale(as float64) {
	c.generator.ApproximationScale(as)
}

// GetApproximationScale returns the current approximation scale.
func (c *ConvContour) GetApproximationScale() float64 {
	return c.generator.GetApproximationScale()
}

// AutoDetectOrientation controls whether contour direction follows the detected
// polygon winding, mirroring AGG's contour generator behavior.
func (c *ConvContour) AutoDetectOrientation(v bool) {
	c.generator.AutoDetectOrientation(v)
}

// GetAutoDetectOrientation returns the current orientation-detection mode.
func (c *ConvContour) GetAutoDetectOrientation() bool {
	return c.generator.GetAutoDetectOrientation()
}
