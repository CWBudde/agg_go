package conv

import (
	"math"

	"agg_go/internal/basics"
	"agg_go/internal/transform"
)

// MarkerLocator interface defines how to locate and retrieve markers along a path
type MarkerLocator interface {
	// Rewind resets the marker locator to start from marker at given index
	Rewind(markerIndex uint)

	// Vertex returns the next vertex pair defining a marker position and direction
	// Returns two consecutive points: the marker position and direction vector
	Vertex() (x, y float64, cmd basics.PathCommand)
}

// MarkerShapes interface defines how to generate marker geometry
type MarkerShapes interface {
	// Rewind resets the marker shapes to start from shape at given index
	Rewind(shapeIndex uint)

	// Vertex returns the next vertex of the current marker shape
	Vertex() (x, y float64, cmd basics.PathCommand)
}

// MarkerStatus represents the state of the marker converter
type MarkerStatus int

const (
	MarkerStatusInitial MarkerStatus = iota
	MarkerStatusMarkers
	MarkerStatusPolygon
	MarkerStatusStop
)

// ConvMarker places markers along paths with automatic orientation
// This is a port of AGG's conv_marker template class
type ConvMarker struct {
	markerLocator *MarkerLocator
	markerShapes  *MarkerShapes
	transform     *transform.TransAffine
	mtx           *transform.TransAffine
	status        MarkerStatus
	marker        uint
	numMarkers    uint
}

// NewConvMarker creates a new marker converter
func NewConvMarker(markerLocator MarkerLocator, markerShapes MarkerShapes) *ConvMarker {
	return &ConvMarker{
		markerLocator: &markerLocator,
		markerShapes:  &markerShapes,
		transform:     transform.NewTransAffine(),
		mtx:           transform.NewTransAffine(),
		status:        MarkerStatusInitial,
		marker:        0,
		numMarkers:    1,
	}
}

// Transform returns a reference to the transformation matrix
func (c *ConvMarker) Transform() *transform.TransAffine {
	return c.transform
}

// Rewind resets the marker converter to start processing
func (c *ConvMarker) Rewind(pathID uint) {
	c.status = MarkerStatusInitial
	c.marker = 0
	c.numMarkers = 1
}

// Vertex returns the next vertex in the marker sequence
func (c *ConvMarker) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = basics.PathCmdMoveTo
	var x1, y1, x2, y2 float64

	for !basics.IsStop(cmd) {
		switch c.status {
		case MarkerStatusInitial:
			if c.numMarkers == 0 {
				cmd = basics.PathCmdStop
				break
			}
			(*c.markerLocator).Rewind(c.marker)
			c.marker++
			c.numMarkers = 0 // Reset count for this marker locator position
			c.status = MarkerStatusMarkers

		case MarkerStatusMarkers:
			// Get first point of marker position/direction pair
			x1, y1, cmd = (*c.markerLocator).Vertex()
			if basics.IsStop(cmd) {
				c.status = MarkerStatusInitial
				break
			}

			// Get second point of marker position/direction pair
			x2, y2, cmd = (*c.markerLocator).Vertex()
			if basics.IsStop(cmd) {
				c.status = MarkerStatusInitial
				break
			}

			c.numMarkers++

			// Calculate marker transformation matrix (matching C++ order)
			// Start with the base transformation
			*c.mtx = *c.transform

			// Multiply by rotation transformation based on direction vector
			angle := math.Atan2(y2-y1, x2-x1)
			rotMatrix := transform.NewTransAffineRotation(angle)
			c.mtx.Multiply(rotMatrix)

			// Multiply by translation to marker position
			transMatrix := transform.NewTransAffineTranslation(x1, y1)
			c.mtx.Multiply(transMatrix)

			// Start reading marker shape
			(*c.markerShapes).Rewind(c.marker - 1)
			c.status = MarkerStatusPolygon

		case MarkerStatusPolygon:
			// Get next vertex from marker shape
			x, y, cmd = (*c.markerShapes).Vertex()
			if basics.IsStop(cmd) {
				cmd = basics.PathCmdMoveTo
				c.status = MarkerStatusMarkers
				break
			}

			// Transform the vertex by the marker transformation matrix
			c.mtx.Transform(&x, &y)
			return x, y, cmd

		case MarkerStatusStop:
			cmd = basics.PathCmdStop
			break
		}
	}

	return x, y, cmd
}
