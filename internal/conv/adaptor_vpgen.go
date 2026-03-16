package conv

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
)

// VPGen is the contract implemented by vpgen-style streaming processors such as
// clip_polygon, clip_polyline, and segmentator.
type VPGen interface {
	Reset()
	MoveTo(x, y float64)
	LineTo(x, y float64)
	Vertex() (x, y float64, cmd basics.PathCommand)
	AutoClose() bool
	AutoUnclose() bool
}

// ConvAdaptorVPGen is the Go equivalent of AGG's conv_adaptor_vpgen. It streams
// source vertices through a vpgen processor while preserving AGG's auto-close
// and end_poly handling.
type ConvAdaptorVPGen[VPG VPGen] struct {
	source    VertexSource
	vpgen     VPG
	startX    float64
	startY    float64
	polyFlags basics.PathCommand
	vertices  int
}

// NewConvAdaptorVPGen creates a vpgen adaptor.
func NewConvAdaptorVPGen[VPG VPGen](source VertexSource, vpgen VPG) *ConvAdaptorVPGen[VPG] {
	return &ConvAdaptorVPGen[VPG]{
		source: source,
		vpgen:  vpgen,
	}
}

// Attach replaces the wrapped source.
func (c *ConvAdaptorVPGen[VPG]) Attach(source VertexSource) {
	c.source = source
}

// VPGen returns the wrapped vpgen object.
func (c *ConvAdaptorVPGen[VPG]) VPGen() VPG {
	return c.vpgen
}

// Rewind resets both the source and the vpgen state for a new path walk.
func (c *ConvAdaptorVPGen[VPG]) Rewind(pathID uint) {
	c.source.Rewind(pathID)
	c.vpgen.Reset()
	c.startX = 0
	c.startY = 0
	c.polyFlags = 0
	c.vertices = 0
}

// Vertex advances the streaming conversion state machine, following AGG's
// conv_adaptor_vpgen logic for auto-closing polygons and replaying end_poly
// flags after vpgen output is drained.
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
