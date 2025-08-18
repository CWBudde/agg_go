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

// Rewind rewinds the converter
func (c *ConvAdaptorVCGen) Rewind(pathID uint) {
	c.source.Rewind(pathID)
	c.status = StatusInitial
}

// Vertex returns the next vertex in the converted sequence
func (c *ConvAdaptorVCGen) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = basics.PathCmdStop

	for {
		switch c.status {
		case StatusInitial:
			c.markers.RemoveAll()
			c.startX, c.startY, c.lastCmd = c.source.Vertex()
			c.status = StatusAccumulate

		case StatusAccumulate:
			if c.lastCmd == basics.PathCmdStop {
				c.status = StatusGenerate
				c.generator.PrepareSrc()
				c.generator.Rewind(0)
				c.markers.PrepareSrc()
				c.markers.Rewind(0)
				continue
			}

			c.generator.AddVertex(c.startX, c.startY, c.lastCmd)
			c.markers.AddVertex(c.startX, c.startY, c.lastCmd)

			c.startX, c.startY, c.lastCmd = c.source.Vertex()
			continue

		case StatusGenerate:
			x, y, cmd = c.generator.Vertex()
			if cmd != basics.PathCmdStop {
				return x, y, cmd
			}

			x, y, cmd = c.markers.Vertex()
			if cmd != basics.PathCmdStop {
				return x, y, cmd
			}

			c.status = StatusInitial
			return 0, 0, basics.PathCmdStop
		}
		break
	}

	return 0, 0, basics.PathCmdStop
}
