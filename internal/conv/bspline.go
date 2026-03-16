package conv

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/vcgen"
)

// ConvBSpline is the Go equivalent of AGG's conv_bspline. It converts the
// source polyline into a sampled B-spline curve using VCGenBSpline.
type ConvBSpline struct {
	*ConvAdaptorVCGen
	generator *vcgen.VCGenBSpline
}

// NewConvBSpline creates a B-spline converter.
func NewConvBSpline(source VertexSource) *ConvBSpline {
	generator := vcgen.NewVCGenBSpline()
	adaptor := NewConvAdaptorVCGen(source, generator)

	return &ConvBSpline{
		ConvAdaptorVCGen: adaptor,
		generator:        generator,
	}
}

// SetInterpolationStep sets the distance between generated spline samples.
func (c *ConvBSpline) SetInterpolationStep(step float64) {
	c.generator.SetInterpolationStep(step)
}

// InterpolationStep returns the current spline sampling step.
func (c *ConvBSpline) InterpolationStep() float64 {
	return c.generator.InterpolationStep()
}

// Rewind resets the converter to the start of the requested path.
func (c *ConvBSpline) Rewind(pathID uint) {
	c.ConvAdaptorVCGen.Rewind(pathID)
}

// Vertex returns the next sampled spline vertex.
func (c *ConvBSpline) Vertex() (x, y float64, cmd basics.PathCommand) {
	return c.ConvAdaptorVCGen.Vertex()
}
