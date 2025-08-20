// Package transform provides transformation interfaces and utilities for AGG.
package transform

// Transformer interface for applying coordinate transformations.
// This interface abstracts the transformation operation, allowing different
// transformation types (affine, perspective, etc.) to be used interchangeably.
type Transformer interface {
	// Transform applies the transformation to the given point coordinates.
	// The x and y coordinates are modified in place.
	Transform(x, y *float64)
}

// InverseTransformer interface for transformations that support inverse operations.
// This interface extends Transformer with the ability to reverse transformations.
type InverseTransformer interface {
	Transformer
	// InverseTransform applies the inverse transformation to the given point coordinates.
	// The x and y coordinates are modified in place.
	InverseTransform(x, y *float64)
}

// TransformerGetter defines the interface for objects that can return their transformer.
// This interface replaces duck-typing patterns for accessing transformers.
type TransformerGetter interface {
	// Transformer returns the current transformer
	Transformer() Transformer
}

// TransformerSetter defines the interface for objects that can have their transformer set.
// This interface replaces duck-typing patterns for setting transformers.
type TransformerSetter interface {
	// SetTransformer sets the transformer
	SetTransformer(transformer Transformer)
}

// TransformerAccessor combines getter and setter for transformer access.
// This interface is useful for objects that both get and set transformers.
type TransformerAccessor interface {
	TransformerGetter
	TransformerSetter
}

// Compile-time interface checks
// These verify that transformer types implement the expected interfaces

// Note: Actual transformer implementations should add checks like:
// var _ Transformer = (*Affine)(nil)
// var _ InverseTransformer = (*Affine)(nil)
// var _ Transformer = (*Perspective)(nil)
// var _ InverseTransformer = (*Perspective)(nil)
// var _ TransformerGetter = (*SomeInterpolator)(nil)
// var _ TransformerSetter = (*SomeInterpolator)(nil)
