package conv

import (
	"agg_go/internal/basics"
	"agg_go/internal/vcgen"
)

// ConvContour is a path converter that creates contour outlines.
// This is a port of AGG's conv_contour template class which creates offset curves
// (parallel lines) at a fixed distance from the original path.
// Unlike stroke which creates outlined shapes with caps, contour creates simple
// offset paths at the specified width.
type ConvContour struct {
	*ConvAdaptorVCGen
	generator *vcgen.VCGenContour
}

// NewConvContour creates a new contour converter with the specified vertex source
func NewConvContour(source VertexSource) *ConvContour {
	generator := vcgen.NewVCGenContour()
	adaptor := NewConvAdaptorVCGen(source, generator)

	return &ConvContour{
		ConvAdaptorVCGen: adaptor,
		generator:        generator,
	}
}

// Generator returns the underlying vertex generator for direct access
func (c *ConvContour) Generator() *vcgen.VCGenContour {
	return c.generator
}

// Width sets the contour width
// Positive values create outer contours for CCW paths, negative for CW paths
func (c *ConvContour) Width(w float64) {
	c.generator.Width(w)
}

// GetWidth returns the current contour width
func (c *ConvContour) GetWidth() float64 {
	return c.generator.GetWidth()
}

// LineJoin sets the line join style for contour vertices
func (c *ConvContour) LineJoin(lj basics.LineJoin) {
	c.generator.LineJoin(lj)
}

// GetLineJoin returns the current line join style
func (c *ConvContour) GetLineJoin() basics.LineJoin {
	return c.generator.GetLineJoin()
}

// InnerJoin sets the inner join style for contour vertices
func (c *ConvContour) InnerJoin(ij basics.InnerJoin) {
	c.generator.InnerJoin(ij)
}

// GetInnerJoin returns the current inner join style
func (c *ConvContour) GetInnerJoin() basics.InnerJoin {
	return c.generator.GetInnerJoin()
}

// MiterLimit sets the miter limit for join calculations
func (c *ConvContour) MiterLimit(ml float64) {
	c.generator.MiterLimit(ml)
}

// GetMiterLimit returns the current miter limit
func (c *ConvContour) GetMiterLimit() float64 {
	return c.generator.GetMiterLimit()
}

// MiterLimitTheta sets the miter limit by angle (in radians)
func (c *ConvContour) MiterLimitTheta(t float64) {
	c.generator.MiterLimitTheta(t)
}

// InnerMiterLimit sets the inner miter limit for join calculations
func (c *ConvContour) InnerMiterLimit(ml float64) {
	c.generator.InnerMiterLimit(ml)
}

// GetInnerMiterLimit returns the current inner miter limit
func (c *ConvContour) GetInnerMiterLimit() float64 {
	return c.generator.GetInnerMiterLimit()
}

// ApproximationScale sets the approximation scale for curve rendering
func (c *ConvContour) ApproximationScale(as float64) {
	c.generator.ApproximationScale(as)
}

// GetApproximationScale returns the current approximation scale
func (c *ConvContour) GetApproximationScale() float64 {
	return c.generator.GetApproximationScale()
}

// AutoDetectOrientation sets whether to automatically detect polygon orientation
// When enabled, the contour width direction will be determined by the polygon area
func (c *ConvContour) AutoDetectOrientation(v bool) {
	c.generator.AutoDetectOrientation(v)
}

// GetAutoDetectOrientation returns the current auto-detect orientation setting
func (c *ConvContour) GetAutoDetectOrientation() bool {
	return c.generator.GetAutoDetectOrientation()
}
