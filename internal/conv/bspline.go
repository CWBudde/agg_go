package conv

import (
	"agg_go/internal/basics"
	"agg_go/internal/vcgen"
)

// ConvBSpline converts path vertices to smooth B-spline curves
// This is equivalent to agg::conv_bspline in the original AGG library
type ConvBSpline struct {
	*ConvAdaptorVCGen
	generator *vcgen.VCGenBSpline
}

// NewConvBSpline creates a new B-spline converter
func NewConvBSpline(source VertexSource) *ConvBSpline {
	generator := vcgen.NewVCGenBSpline()
	adaptor := NewConvAdaptorVCGen(source, generator)

	return &ConvBSpline{
		ConvAdaptorVCGen: adaptor,
		generator:        generator,
	}
}

// SetInterpolationStep sets the interpolation step size
// Smaller values create smoother curves with more vertices
func (c *ConvBSpline) SetInterpolationStep(step float64) {
	c.generator.SetInterpolationStep(step)
}

// InterpolationStep returns the current interpolation step
func (c *ConvBSpline) InterpolationStep() float64 {
	return c.generator.InterpolationStep()
}

// Rewind rewinds the B-spline converter
func (c *ConvBSpline) Rewind(pathID uint) {
	c.ConvAdaptorVCGen.Rewind(pathID)
}

// Vertex returns the next vertex in the B-spline approximation
func (c *ConvBSpline) Vertex() (x, y float64, cmd basics.PathCommand) {
	return c.ConvAdaptorVCGen.Vertex()
}
