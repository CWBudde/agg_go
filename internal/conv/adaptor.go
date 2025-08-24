package conv

import (
	"agg_go/internal/basics"
)

// Status represents the state of the conv_adaptor_vcgen
type Status int

const (
	StatusInitial Status = iota
	StatusAccumulate
	StatusGenerate
	StatusMarkers
)

// VertexSource interface for providing vertices
type VertexSource interface {
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
}

// VertexGenerator interface for generating vertices from accumulated path data
type VertexGenerator interface {
	RemoveAll()
	AddVertex(x, y float64, cmd basics.PathCommand)
	PrepareSrc()
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
}

// NullMarkers is a default marker implementation that does nothing
type NullMarkers struct{}

// RemoveAll removes all markers (no-op)
func (m *NullMarkers) RemoveAll() {}

// AddVertex adds a vertex marker (no-op)
func (m *NullMarkers) AddVertex(x, y float64, cmd basics.PathCommand) {}

// PrepareSrc prepares source for marker processing (no-op)
func (m *NullMarkers) PrepareSrc() {}

// Rewind rewinds the marker iterator (no-op)
func (m *NullMarkers) Rewind(pathID uint) {}

// Vertex returns the next marker vertex (always stops)
func (m *NullMarkers) Vertex() (x, y float64, cmd basics.PathCommand) {
	return 0, 0, basics.PathCmdStop
}

// Markers interface for terminal markers
type Markers interface {
	RemoveAll()
	AddVertex(x, y float64, cmd basics.PathCommand)
	PrepareSrc()
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
}

// ConvAdaptorVCGen is the base class for vertex converter generators
type ConvAdaptorVCGen struct {
	source    VertexSource
	generator VertexGenerator
	markers   Markers
	status    Status
	lastCmd   basics.PathCommand
	startX    float64
	startY    float64
}

// NewConvAdaptorVCGen creates a new converter adaptor with vertex generator
func NewConvAdaptorVCGen(source VertexSource, generator VertexGenerator) *ConvAdaptorVCGen {
	return &ConvAdaptorVCGen{
		source:    source,
		generator: generator,
		markers:   &NullMarkers{},
		status:    StatusInitial,
	}
}

// NewConvAdaptorVCGenWithMarkers creates a new converter adaptor with vertex generator and markers
func NewConvAdaptorVCGenWithMarkers(source VertexSource, generator VertexGenerator, markers Markers) *ConvAdaptorVCGen {
	return &ConvAdaptorVCGen{
		source:    source,
		generator: generator,
		markers:   markers,
		status:    StatusInitial,
	}
}

// Attach attaches a new vertex source
func (c *ConvAdaptorVCGen) Attach(source VertexSource) {
	c.source = source
}

// Generator returns the vertex generator
func (c *ConvAdaptorVCGen) Generator() VertexGenerator {
	return c.generator
}

// Markers returns the markers
func (c *ConvAdaptorVCGen) Markers() Markers {
	return c.markers
}

// GetGenerator returns the vertex generator (read-only access)
func (c *ConvAdaptorVCGen) GetGenerator() VertexGenerator {
	return c.generator
}

// GetMarkers returns the markers (read-only access)
func (c *ConvAdaptorVCGen) GetMarkers() Markers {
	return c.markers
}

// Rewind rewinds the converter
func (c *ConvAdaptorVCGen) Rewind(pathID uint) {
	c.source.Rewind(pathID)
	c.status = StatusInitial
}

// Vertex returns the next vertex in the converted sequence
func (c *ConvAdaptorVCGen) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = basics.PathCmdStop
	done := false

	for !done {
		switch c.status {
		case StatusInitial:
			c.markers.RemoveAll()
			c.startX, c.startY, c.lastCmd = c.source.Vertex()
			c.status = StatusAccumulate
			fallthrough

		case StatusAccumulate:
			if basics.IsStop(c.lastCmd) {
				return 0, 0, basics.PathCmdStop
			}

			c.generator.RemoveAll()
			c.generator.AddVertex(c.startX, c.startY, basics.PathCmdMoveTo)
			c.markers.AddVertex(c.startX, c.startY, basics.PathCmdMoveTo)

			for {
				x, y, cmd = c.source.Vertex()
				if basics.IsVertex(cmd) {
					c.lastCmd = cmd
					if basics.IsMoveTo(cmd) {
						c.startX = x
						c.startY = y
						break
					}
					c.generator.AddVertex(x, y, cmd)
					c.markers.AddVertex(x, y, basics.PathCmdLineTo)
				} else {
					if basics.IsStop(cmd) {
						c.lastCmd = basics.PathCmdStop
						break
					}
					if basics.IsEndPoly(cmd) {
						c.generator.AddVertex(x, y, cmd)
						break
					}
				}
			}
			c.generator.Rewind(0)
			c.status = StatusGenerate

		case StatusGenerate:
			x, y, cmd = c.generator.Vertex()
			if basics.IsStop(cmd) {
				// After generator finishes, prepare and output markers
				c.markers.PrepareSrc()
				c.markers.Rewind(0)
				c.status = StatusMarkers
				continue // Continue to markers state
			}
			done = true

		case StatusMarkers:
			x, y, cmd = c.markers.Vertex()
			if basics.IsStop(cmd) {
				// After markers finish, read next command from source for next accumulate cycle
				c.startX, c.startY, c.lastCmd = c.source.Vertex()
				c.status = StatusAccumulate
				continue // Continue to next iteration of while loop
			}
			done = true
		}
	}
	return x, y, cmd
}
