package conv

import (
	"agg_go/internal/basics"
)

// ConvUnclosePolygon removes closing flags from polygon endpoints, converting
// closed polygons to open polylines. This is the inverse operation of ConvClosePolygon.
//
// This is equivalent to AGG's conv_unclose_polygon template class.
type ConvUnclosePolygon struct {
	source VertexSource
}

// NewConvUnclosePolygon creates a new polygon unclosing converter.
func NewConvUnclosePolygon(source VertexSource) *ConvUnclosePolygon {
	return &ConvUnclosePolygon{
		source: source,
	}
}

// Attach attaches a new vertex source to the converter.
func (c *ConvUnclosePolygon) Attach(source VertexSource) {
	c.source = source
}

// Rewind rewinds the vertex source to the beginning of the specified path.
func (c *ConvUnclosePolygon) Rewind(pathID uint) {
	c.source.Rewind(pathID)
}

// Vertex returns the next vertex from the source, removing close flags from EndPoly commands.
func (c *ConvUnclosePolygon) Vertex() (x, y float64, cmd basics.PathCommand) {
	x, y, cmd = c.source.Vertex()

	// If this is an EndPoly command, remove the close flag
	if basics.IsEndPoly(cmd) {
		cmd &= ^basics.PathFlagClose
	}

	return x, y, cmd
}
