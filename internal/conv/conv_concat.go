package conv

import (
	"agg_go/internal/basics"
)

// ConvConcat concatenates vertices from two vertex sources sequentially.
// This converter outputs all vertices from the first source, then all vertices
// from the second source. It's typically used to combine paths with markers
// such as arrowheads.
//
// This is equivalent to AGG's conv_concat template class.
type ConvConcat[VS1, VS2 VertexSource] struct {
	source1 VS1
	source2 VS2
	status  int // 0 = reading from source1, 1 = reading from source2, 2 = done
}

// NewConvConcat creates a new path concatenation converter.
func NewConvConcat[VS1, VS2 VertexSource](source1 VS1, source2 VS2) *ConvConcat[VS1, VS2] {
	return &ConvConcat[VS1, VS2]{
		source1: source1,
		source2: source2,
		status:  2, // Start in stopped state
	}
}

// Attach1 attaches a new first vertex source to the converter.
func (c *ConvConcat[VS1, VS2]) Attach1(source VS1) {
	c.source1 = source
}

// Attach2 attaches a new second vertex source to the converter.
func (c *ConvConcat[VS1, VS2]) Attach2(source VS2) {
	c.source2 = source
}

// Rewind rewinds both vertex sources to the beginning of the specified paths.
// The first source uses the provided pathID, the second source always uses pathID 0.
func (c *ConvConcat[VS1, VS2]) Rewind(pathID uint) {
	c.source1.Rewind(pathID)
	c.source2.Rewind(0)
	c.status = 0
}

// Vertex returns the next vertex from the concatenated sources.
// It first returns all vertices from source1, then all vertices from source2.
func (c *ConvConcat[VS1, VS2]) Vertex() (x, y float64, cmd basics.PathCommand) {
	// If reading from source1
	if c.status == 0 {
		x, y, cmd = c.source1.Vertex()
		if !basics.IsStop(cmd) {
			return x, y, cmd
		}
		c.status = 1
	}

	// If reading from source2
	if c.status == 1 {
		x, y, cmd = c.source2.Vertex()
		if !basics.IsStop(cmd) {
			return x, y, cmd
		}
		c.status = 2
	}

	// Done with both sources
	return 0, 0, basics.PathCmdStop
}
