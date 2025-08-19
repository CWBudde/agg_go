// Package conv provides path conversion utilities for AGG.
package conv

import (
	"agg_go/internal/basics"
	"agg_go/internal/transform"
)

// ConvTransform applies coordinate transformations to vertex sources.
// It's a generic converter that takes any VertexSource and any Transformer,
// applying the transformation to vertex coordinates while preserving path commands.
//
// This is equivalent to AGG's conv_transform template class.
type ConvTransform[VS VertexSource, T transform.Transformer] struct {
	source      VS
	transformer T
}

// NewConvTransform creates a new transform converter with the given source and transformer.
func NewConvTransform[VS VertexSource, T transform.Transformer](source VS, transformer T) *ConvTransform[VS, T] {
	return &ConvTransform[VS, T]{
		source:      source,
		transformer: transformer,
	}
}

// Attach attaches a new vertex source to the converter.
func (ct *ConvTransform[VS, T]) Attach(source VS) {
	ct.source = source
}

// SetTransformer sets a new transformer for the converter.
// This is an idiomatic Go method name for setting the transformer.
func (ct *ConvTransform[VS, T]) SetTransformer(transformer T) {
	ct.transformer = transformer
}

// Transformer sets a new transformer for the converter.
// This method name matches the AGG C++ API exactly.
func (ct *ConvTransform[VS, T]) Transformer(transformer T) {
	ct.transformer = transformer
}

// Rewind rewinds the vertex source to the beginning of the specified path.
func (ct *ConvTransform[VS, T]) Rewind(pathID uint) {
	ct.source.Rewind(pathID)
}

// Vertex returns the next vertex from the source with transformation applied.
// Only vertex coordinates are transformed; path commands pass through unchanged.
func (ct *ConvTransform[VS, T]) Vertex() (x, y float64, cmd basics.PathCommand) {
	x, y, cmd = ct.source.Vertex()
	if basics.IsVertex(cmd) {
		ct.transformer.Transform(&x, &y)
	}
	return x, y, cmd
}
