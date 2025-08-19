package conv

import (
	"agg_go/internal/basics"
	"agg_go/internal/vcgen"
)

// ConvMarkerAdaptor is an adaptor for custom marker systems
// This is a port of AGG's conv_marker_adaptor template class
type ConvMarkerAdaptor struct {
	*ConvAdaptorVCGen
	vcgen *vcgen.VCGenVertexSequence
}

// NewConvMarkerAdaptor creates a new marker adaptor converter
func NewConvMarkerAdaptor(source VertexSource) *ConvMarkerAdaptor {
	vcgen := vcgen.NewVCGenVertexSequence()
	adaptor := NewConvAdaptorVCGen(source, vcgen)

	return &ConvMarkerAdaptor{
		ConvAdaptorVCGen: adaptor,
		vcgen:            vcgen,
	}
}

// NewConvMarkerAdaptorWithMarkers creates a new marker adaptor converter with custom markers
func NewConvMarkerAdaptorWithMarkers(source VertexSource, markers Markers) *ConvMarkerAdaptor {
	vcgen := vcgen.NewVCGenVertexSequence()
	adaptor := NewConvAdaptorVCGenWithMarkers(source, vcgen, markers)

	return &ConvMarkerAdaptor{
		ConvAdaptorVCGen: adaptor,
		vcgen:            vcgen,
	}
}

// SetShorten sets the path shortening distance
// This affects how markers are positioned relative to the path endpoints
func (c *ConvMarkerAdaptor) SetShorten(s float64) {
	c.vcgen.SetShorten(s)
}

// Shorten returns the current path shortening distance
func (c *ConvMarkerAdaptor) Shorten() float64 {
	return c.vcgen.Shorten()
}

// Attach attaches a new vertex source to the converter
func (c *ConvMarkerAdaptor) Attach(source VertexSource) {
	c.ConvAdaptorVCGen.Attach(source)
}

// Rewind rewinds the converter to start processing from the beginning
func (c *ConvMarkerAdaptor) Rewind(pathID uint) {
	c.ConvAdaptorVCGen.Rewind(pathID)
}

// Vertex returns the next vertex in the processed path with marker support
func (c *ConvMarkerAdaptor) Vertex() (x, y float64, cmd basics.PathCommand) {
	return c.ConvAdaptorVCGen.Vertex()
}
