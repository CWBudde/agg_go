package conv

import (
	"agg_go/internal/basics"
	"agg_go/internal/vcgen"
)

// ConvShortenPath is a vertex converter that shortens paths by removing
// segments from the end of the path. This is useful for creating effects
// like arrow heads or ensuring paths don't overlap with markers.
//
// This is a port of AGG's conv_shorten_path template class.
type ConvShortenPath struct {
	*ConvAdaptorVCGen
	vcgen *vcgen.VCGenVertexSequence
}

// NewConvShortenPath creates a new path shortening converter
func NewConvShortenPath(source VertexSource) *ConvShortenPath {
	vcgen := vcgen.NewVCGenVertexSequence()
	adaptor := NewConvAdaptorVCGen(source, vcgen)

	return &ConvShortenPath{
		ConvAdaptorVCGen: adaptor,
		vcgen:            vcgen,
	}
}

// SetShorten sets the path shortening distance
// A positive value shortens the path from the end
func (c *ConvShortenPath) SetShorten(s float64) {
	c.vcgen.SetShorten(s)
}

// Shorten returns the current path shortening distance
func (c *ConvShortenPath) Shorten() float64 {
	return c.vcgen.Shorten()
}

// Attach attaches a new vertex source to the converter
func (c *ConvShortenPath) Attach(source VertexSource) {
	c.ConvAdaptorVCGen.Attach(source)
}

// Rewind rewinds the converter to start processing from the beginning
func (c *ConvShortenPath) Rewind(pathID uint) {
	c.ConvAdaptorVCGen.Rewind(pathID)
}

// Vertex returns the next vertex in the shortened path
func (c *ConvShortenPath) Vertex() (x, y float64, cmd basics.PathCommand) {
	return c.ConvAdaptorVCGen.Vertex()
}
