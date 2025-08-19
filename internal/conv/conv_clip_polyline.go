package conv

import (
	"agg_go/internal/vcgen"
)

// ConvClipPolyline is a converter that clips polylines using the Liang-Barsky algorithm.
// This is equivalent to conv_clip_polyline<VertexSource> in the C++ AGG implementation.
type ConvClipPolyline struct {
	*ConvAdaptorVPGen[*vcgen.VPGenClipPolyline]
}

// NewConvClipPolyline creates a new polyline clipping converter
func NewConvClipPolyline(source VertexSource) *ConvClipPolyline {
	vpgen := vcgen.NewVPGenClipPolyline()
	adaptor := NewConvAdaptorVPGen(source, vpgen)
	return &ConvClipPolyline{adaptor}
}

// ClipBox sets the clipping rectangle
func (c *ConvClipPolyline) ClipBox(x1, y1, x2, y2 float64) {
	c.VPGen().ClipBox(x1, y1, x2, y2)
}

// X1 returns the left edge of the clipping box
func (c *ConvClipPolyline) X1() float64 {
	return c.VPGen().X1()
}

// Y1 returns the bottom edge of the clipping box
func (c *ConvClipPolyline) Y1() float64 {
	return c.VPGen().Y1()
}

// X2 returns the right edge of the clipping box
func (c *ConvClipPolyline) X2() float64 {
	return c.VPGen().X2()
}

// Y2 returns the top edge of the clipping box
func (c *ConvClipPolyline) Y2() float64 {
	return c.VPGen().Y2()
}
