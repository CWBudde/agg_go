package transform

// Transformer is the common forward-transform contract shared by AGG's
// trans_* classes and span interpolators.
type Transformer interface {
	Transform(x, y *float64)
}

// InverseTransformer extends Transformer with inverse mapping support.
type InverseTransformer interface {
	Transformer
	InverseTransform(x, y *float64)
}

// TransformerGetter exposes a currently configured transformer.
type TransformerGetter interface {
	Transformer() Transformer
}

// TransformerSetter allows callers to replace a configured transformer.
type TransformerSetter interface {
	SetTransformer(transformer Transformer)
}

// TransformerAccessor combines getter and setter access.
type TransformerAccessor interface {
	TransformerGetter
	TransformerSetter
}
