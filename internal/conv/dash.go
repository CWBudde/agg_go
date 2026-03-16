package conv

import (
	"github.com/MeKo-Christian/agg_go/internal/vcgen"
)

// ConvDash is the Go equivalent of AGG's conv_dash. It wraps VCGenDash behind
// ConvAdaptorVCGen so a source path is rewritten as a dashed path before any
// later stroke or contour stage.
type ConvDash struct {
	*ConvAdaptorVCGen
	dashGen *vcgen.VCGenDash
}

// NewConvDash creates a dash converter without terminal markers.
func NewConvDash(source VertexSource) *ConvDash {
	dashGen := vcgen.NewVCGenDash()
	conv := &ConvDash{
		ConvAdaptorVCGen: NewConvAdaptorVCGen(source, dashGen),
		dashGen:          dashGen,
	}
	return conv
}

// NewConvDashWithMarkers creates a dash converter with terminal-marker support.
func NewConvDashWithMarkers(source VertexSource, markers Markers) *ConvDash {
	dashGen := vcgen.NewVCGenDash()
	conv := &ConvDash{
		ConvAdaptorVCGen: NewConvAdaptorVCGenWithMarkers(source, dashGen, markers),
		dashGen:          dashGen,
	}
	return conv
}

// RemoveAllDashes clears the dash pattern, returning to solid-line behavior.
func (c *ConvDash) RemoveAllDashes() {
	c.dashGen.RemoveAllDashes()
}

// AddDash appends one dash-gap pair to the repeating pattern.
func (c *ConvDash) AddDash(dashLen, gapLen float64) {
	c.dashGen.AddDash(dashLen, gapLen)
}

// DashStart sets the phase offset into the repeating dash pattern.
func (c *ConvDash) DashStart(ds float64) {
	c.dashGen.DashStart(ds)
}

// GetDashStart returns the current phase offset.
func (c *ConvDash) GetDashStart() float64 {
	return c.dashGen.GetDashStart()
}

// Shorten trims both ends of open paths before dash generation.
func (c *ConvDash) Shorten(s float64) {
	c.dashGen.Shorten(s)
}

// GetShorten returns the current end-trimming amount.
func (c *ConvDash) GetShorten() float64 {
	return c.dashGen.GetShorten()
}

// DashGenerator returns the underlying VCGenDash.
func (c *ConvDash) DashGenerator() *vcgen.VCGenDash {
	return c.dashGen
}

// NumDashes returns the number of stored dash elements.
func (c *ConvDash) NumDashes() uint {
	return c.dashGen.NumDashes()
}
