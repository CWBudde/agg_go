package conv

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
)

// Status tracks the internal state machine of ConvAdaptorVCGen while it
// alternates between accumulating a source subpath and emitting generated
// vertices.
type Status int

const (
	StatusInitial Status = iota
	StatusAccumulate
	StatusGenerate
)

// VertexSource is the common path-source contract used across AGG's converter
// layer.
type VertexSource interface {
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
}

// VertexGenerator is the contract implemented by vcgen-style generators such as
// stroke, dash, contour, and B-spline.
type VertexGenerator interface {
	RemoveAll()
	AddVertex(x, y float64, cmd basics.PathCommand)
	PrepareSrc()
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
}

// NullMarkers is the default no-op terminal-marker implementation.
type NullMarkers struct{}

// RemoveAll is a no-op.
func (m *NullMarkers) RemoveAll() {}

// AddVertex is a no-op.
func (m *NullMarkers) AddVertex(x, y float64, cmd basics.PathCommand) {}

// PrepareSrc is a no-op.
func (m *NullMarkers) PrepareSrc() {}

// Rewind is a no-op.
func (m *NullMarkers) Rewind(pathID uint) {}

// Vertex always reports stop.
func (m *NullMarkers) Vertex() (x, y float64, cmd basics.PathCommand) {
	return 0, 0, basics.PathCmdStop
}

// Markers is the optional terminal-marker contract used by converters such as
// stroked paths with arrowheads.
type Markers interface {
	RemoveAll()
	AddVertex(x, y float64, cmd basics.PathCommand)
	PrepareSrc()
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
}

// ConvAdaptorVCGen is the Go equivalent of AGG's conv_adaptor_vcgen. It
// accumulates one source subpath, feeds it into a vcgen-style generator, and
// then exposes the generated outline as a new VertexSource.
type ConvAdaptorVCGen struct {
	source    VertexSource
	generator VertexGenerator
	markers   Markers
	status    Status
	lastCmd   basics.PathCommand
	startX    float64
	startY    float64
}

// NewConvAdaptorVCGen creates a vcgen adaptor without terminal markers.
func NewConvAdaptorVCGen(source VertexSource, generator VertexGenerator) *ConvAdaptorVCGen {
	return &ConvAdaptorVCGen{
		source:    source,
		generator: generator,
		markers:   &NullMarkers{},
		status:    StatusInitial,
	}
}

// NewConvAdaptorVCGenWithMarkers creates a vcgen adaptor with terminal-marker
// support.
func NewConvAdaptorVCGenWithMarkers(source VertexSource, generator VertexGenerator, markers Markers) *ConvAdaptorVCGen {
	return &ConvAdaptorVCGen{
		source:    source,
		generator: generator,
		markers:   markers,
		status:    StatusInitial,
	}
}

// Attach replaces the wrapped source.
func (c *ConvAdaptorVCGen) Attach(source VertexSource) {
	c.source = source
}

// Generator returns the wrapped vcgen object.
func (c *ConvAdaptorVCGen) Generator() VertexGenerator {
	return c.generator
}

// Markers returns the terminal-marker implementation.
func (c *ConvAdaptorVCGen) Markers() Markers {
	return c.markers
}

// GetGenerator is an alias for Generator.
func (c *ConvAdaptorVCGen) GetGenerator() VertexGenerator {
	return c.generator
}

// GetMarkers is an alias for Markers.
func (c *ConvAdaptorVCGen) GetMarkers() Markers {
	return c.markers
}

// Rewind resets the adaptor to the start of the requested path.
func (c *ConvAdaptorVCGen) Rewind(pathID uint) {
	c.source.Rewind(pathID)
	c.status = StatusInitial
}

// Vertex advances the adaptor state machine and returns the next generated
// vertex. The logic mirrors AGG's conv_adaptor_vcgen flow: accumulate one
// subpath, rewind the generator, then drain generated output.
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
				c.status = StatusAccumulate
				continue
			}
			done = true
		}
	}
	return x, y, cmd
}
