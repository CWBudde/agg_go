package conv

import (
	"agg_go/internal/vcgen"
)

// ConvDash is a dash converter that uses VCGenDash as the vertex generator.
// This is a port of AGG's conv_dash template class.
type ConvDash struct {
	*ConvAdaptorVCGen
	dashGen *vcgen.VCGenDash
}

// NewConvDash creates a new dash converter with the specified vertex source
func NewConvDash(source VertexSource) *ConvDash {
	dashGen := vcgen.NewVCGenDash()
	conv := &ConvDash{
		ConvAdaptorVCGen: NewConvAdaptorVCGen(source, dashGen),
		dashGen:          dashGen,
	}
	return conv
}

// NewConvDashWithMarkers creates a new dash converter with markers
func NewConvDashWithMarkers(source VertexSource, markers Markers) *ConvDash {
	dashGen := vcgen.NewVCGenDash()
	conv := &ConvDash{
		ConvAdaptorVCGen: NewConvAdaptorVCGenWithMarkers(source, dashGen, markers),
		dashGen:          dashGen,
	}
	return conv
}

// RemoveAllDashes removes all dash patterns
func (c *ConvDash) RemoveAllDashes() {
	c.dashGen.RemoveAllDashes()
}

// AddDash adds a dash pattern (dash length + gap length)
func (c *ConvDash) AddDash(dashLen, gapLen float64) {
	c.dashGen.AddDash(dashLen, gapLen)
}

// DashStart sets the dash start offset
func (c *ConvDash) DashStart(ds float64) {
	c.dashGen.DashStart(ds)
}

// Shorten sets the path shortening distance
func (c *ConvDash) Shorten(s float64) {
	c.dashGen.Shorten(s)
}

// GetShorten returns the current path shortening distance
func (c *ConvDash) GetShorten() float64 {
	return c.dashGen.GetShorten()
}

// DashGenerator returns the underlying dash generator for direct access
func (c *ConvDash) DashGenerator() *vcgen.VCGenDash {
	return c.dashGen
}
