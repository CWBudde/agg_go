package conv

import (
	"agg_go/internal/basics"
)

// ConvClosePolygon ensures that polygons are properly closed by automatically
// adding the close flag to EndPoly commands. This converter tracks polygon
// sequences and inserts EndPoly commands with close flags when needed.
//
// This is equivalent to AGG's conv_close_polygon template class.
type ConvClosePolygon struct {
	source VertexSource
	cmd    [2]basics.PathCommand // Command buffer
	x      [2]float64            // X coordinate buffer
	y      [2]float64            // Y coordinate buffer
	vertex int                   // Current buffer index (0, 1, or 2+)
	lineTo bool                  // Whether we've seen line_to commands
}

// NewConvClosePolygon creates a new polygon closing converter.
func NewConvClosePolygon(source VertexSource) *ConvClosePolygon {
	return &ConvClosePolygon{
		source: source,
		vertex: 2, // Start with no buffered commands
	}
}

// Attach attaches a new vertex source to the converter.
func (c *ConvClosePolygon) Attach(source VertexSource) {
	c.source = source
}

// Rewind rewinds the vertex source to the beginning of the specified path.
func (c *ConvClosePolygon) Rewind(pathID uint) {
	c.source.Rewind(pathID)
	c.vertex = 2
	c.lineTo = false
}

// Vertex returns the next vertex from the source, adding close flags to EndPoly commands as needed.
func (c *ConvClosePolygon) Vertex() (x, y float64, cmd basics.PathCommand) {
	for {
		// If we have buffered commands to return
		if c.vertex < 2 {
			x = c.x[c.vertex]
			y = c.y[c.vertex]
			cmd = c.cmd[c.vertex]
			c.vertex++
			return x, y, cmd
		}

		// Get next vertex from source
		x, y, cmd = c.source.Vertex()

		// Handle EndPoly - add close flag
		if basics.IsEndPoly(cmd) {
			cmd |= basics.PathFlagClose
			return x, y, cmd
		}

		// Handle Stop - if we've seen line_to commands, insert EndPoly|Close first
		if basics.IsStop(cmd) {
			if c.lineTo {
				c.cmd[0] = basics.PathCmdEndPoly | basics.PathFlagClose
				c.cmd[1] = basics.PathCmdStop
				c.x[0] = 0.0
				c.y[0] = 0.0
				c.x[1] = x
				c.y[1] = y
				c.vertex = 0
				c.lineTo = false
				continue
			}
			return x, y, cmd
		}

		// Handle MoveTo - if we've seen line_to commands, insert EndPoly|Close first
		if basics.IsMoveTo(cmd) {
			if c.lineTo {
				c.cmd[0] = basics.PathCmdEndPoly | basics.PathFlagClose
				c.cmd[1] = cmd
				c.x[0] = 0.0
				c.y[0] = 0.0
				c.x[1] = x
				c.y[1] = y
				c.vertex = 0
				c.lineTo = false
				continue
			}
			return x, y, cmd
		}

		// Handle any vertex command - set line_to flag
		if basics.IsVertex(cmd) {
			c.lineTo = true
		}

		return x, y, cmd
	}
}
