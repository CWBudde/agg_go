package conv

import (
	"agg_go/internal/vcgen"
)

// ConvSegmentator is a path converter that segments lines into equal-length pieces.
// This is equivalent to conv_segmentator<VertexSource> in the C++ AGG implementation.
//
// The segmentator divides each line segment into smaller pieces based on the
// approximation scale. This is useful for creating evenly-spaced points along paths
// or for preparing paths for further processing that requires uniform sampling.
type ConvSegmentator struct {
	adaptor *ConvAdaptorVPGen[*vcgen.VPGenSegmentator]
}

// NewConvSegmentator creates a new segmentator converter
func NewConvSegmentator(vs VertexSource) *ConvSegmentator {
	segmentator := vcgen.NewVPGenSegmentator()
	adaptor := NewConvAdaptorVPGen(vs, segmentator)

	return &ConvSegmentator{
		adaptor: adaptor,
	}
}

// Attach sets a new vertex source for the converter
func (c *ConvSegmentator) Attach(vs VertexSource) {
	c.adaptor.Attach(vs)
}

// ApproximationScale sets the approximation scale factor.
// Higher values create more segments per unit length.
func (c *ConvSegmentator) ApproximationScale(scale float64) {
	c.adaptor.VPGen().ApproximationScale(scale)
}

// GetApproximationScale returns the current approximation scale
func (c *ConvSegmentator) GetApproximationScale() float64 {
	return c.adaptor.VPGen().GetApproximationScale()
}

// Rewind resets the converter to start reading from the beginning
func (c *ConvSegmentator) Rewind(pathID uint) {
	c.adaptor.Rewind(pathID)
}

// Vertex returns the next segmented vertex from the path
func (c *ConvSegmentator) Vertex() (x, y float64, cmd uint32) {
	fx, fy, fcmd := c.adaptor.Vertex()
	return fx, fy, uint32(fcmd)
}
