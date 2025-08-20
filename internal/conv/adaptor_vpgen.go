package conv

import (
	"agg_go/internal/basics"
)

// VPGen interface defines the contract for vertex processor generators
// This matches the expected interface from the C++ AGG implementation
type VPGen interface {
	// Reset resets the vertex processor state
	Reset()

	// MoveTo starts a new path at the given coordinates
	MoveTo(x, y float64)

	// LineTo adds a line segment to the current path
	LineTo(x, y float64)

	// Vertex returns the next processed vertex
	Vertex() (x, y float64, cmd basics.PathCommand)

	// AutoClose returns true if polygons should be automatically closed
	AutoClose() bool

	// AutoUnclose returns true if polygons should be automatically unclosed
	AutoUnclose() bool
}

// ConvAdaptorVPGen is a generic adaptor that connects vertex sources with vertex processor generators
// This is equivalent to conv_adaptor_vpgen<VertexSource, VPGen> in the C++ implementation
type ConvAdaptorVPGen[VPG VPGen] struct {
	source    VertexSource
	vpgen     VPG
	startX    float64
	startY    float64
	polyFlags basics.PathCommand
	vertices  int
}

// NewConvAdaptorVPGen creates a new vertex processor adaptor
func NewConvAdaptorVPGen[VPG VPGen](source VertexSource, vpgen VPG) *ConvAdaptorVPGen[VPG] {
	return &ConvAdaptorVPGen[VPG]{
		source: source,
		vpgen:  vpgen,
	}
}

// Attach sets a new vertex source
func (c *ConvAdaptorVPGen[VPG]) Attach(source VertexSource) {
	c.source = source
}

// VPGen returns a reference to the vertex processor generator
func (c *ConvAdaptorVPGen[VPG]) VPGen() VPG {
	return c.vpgen
}

// Rewind resets the adaptor to start reading from the beginning of the path
func (c *ConvAdaptorVPGen[VPG]) Rewind(pathID uint) {
	c.source.Rewind(pathID)
	c.vpgen.Reset()
	c.startX = 0
	c.startY = 0
	c.polyFlags = 0
	c.vertices = 0
}

// Vertex returns the next vertex from the processed path
func (c *ConvAdaptorVPGen[VPG]) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = basics.PathCmdStop

	for {
		x, y, cmd = c.vpgen.Vertex()
		if !basics.IsStop(cmd) {
			break
		}

		// Handle polygon flags and auto-unclose
		if c.polyFlags != 0 && !c.vpgen.AutoUnclose() {
			x = 0.0
			y = 0.0
			cmd = c.polyFlags
			c.polyFlags = 0
			break
		}

		// Handle negative vertex count (closing state)
		if c.vertices < 0 {
			if c.vertices < -1 {
				c.vertices = 0
				return 0, 0, basics.PathCmdStop
			}
			c.vpgen.MoveTo(c.startX, c.startY)
			c.vertices = 1
			continue
		}

		// Get next vertex from source
		tx, ty, cmd := c.source.Vertex()
		if basics.IsVertex(cmd) {
			if basics.IsMoveTo(cmd) {
				// Handle auto-close for move_to
				if c.vpgen.AutoClose() && c.vertices > 2 {
					c.vpgen.LineTo(c.startX, c.startY)
					c.polyFlags = basics.PathCmdEndPoly | basics.PathCommand(basics.PathFlagsClose)
					c.startX = tx
					c.startY = ty
					c.vertices = -1
					continue
				}
				c.vpgen.MoveTo(tx, ty)
				c.startX = tx
				c.startY = ty
				c.vertices = 1
			} else {
				// Line or curve command
				c.vpgen.LineTo(tx, ty)
				c.vertices++
			}
		} else {
			if basics.IsEndPoly(cmd) {
				c.polyFlags = cmd
				if basics.IsClosed(uint32(cmd)) || c.vpgen.AutoClose() {
					if c.vpgen.AutoClose() {
						c.polyFlags |= basics.PathCommand(basics.PathFlagsClose)
					}
					if c.vertices > 2 {
						c.vpgen.LineTo(c.startX, c.startY)
					}
					c.vertices = 0
				}
			} else {
				// path_cmd_stop
				if c.vpgen.AutoClose() && c.vertices > 2 {
					c.vpgen.LineTo(c.startX, c.startY)
					c.polyFlags = basics.PathCmdEndPoly | basics.PathCommand(basics.PathFlagsClose)
					c.vertices = -2
					continue
				}
				break
			}
		}
	}

	return x, y, cmd
}
