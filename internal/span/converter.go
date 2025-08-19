// Package span provides span conversion functionality for AGG rendering.
// This file implements the span converter pipeline for two-stage span processing.
package span

import (
	"reflect"
)

// SpanConverterInterface defines the interface for span converters.
// A span converter takes an existing span and applies some transformation to it.
type SpanConverterInterface interface {
	// Prepare is called before rendering begins
	Prepare()

	// Generate applies conversion to the provided colors array for the given span
	Generate(colors []interface{}, x, y, len int)
}

// SpanConverter implements a pipeline for span processing that chains
// a span generator with a span converter in a two-stage process.
// This corresponds to AGG's span_converter<SpanGenerator, SpanConverter> template class.
type SpanConverter[SG SpanGenerator, SC SpanConverterInterface] struct {
	spanGen SG // The span generator
	spanCnv SC // The span converter
}

// NewSpanConverter creates a new span converter with the given generator and converter.
func NewSpanConverter[SG SpanGenerator, SC SpanConverterInterface](spanGen SG, spanCnv SC) *SpanConverter[SG, SC] {
	return &SpanConverter[SG, SC]{
		spanGen: spanGen,
		spanCnv: spanCnv,
	}
}

// NewSpanConverterEmpty creates a new empty span converter.
// Components must be attached later using AttachGenerator and AttachConverter.
func NewSpanConverterEmpty[SG SpanGenerator, SC SpanConverterInterface]() *SpanConverter[SG, SC] {
	var nilGen SG
	var nilCnv SC
	return &SpanConverter[SG, SC]{
		spanGen: nilGen,
		spanCnv: nilCnv,
	}
}

// AttachGenerator attaches or changes the span generator.
func (sc *SpanConverter[SG, SC]) AttachGenerator(spanGen SG) {
	sc.spanGen = spanGen
}

// AttachConverter attaches or changes the span converter.
func (sc *SpanConverter[SG, SC]) AttachConverter(spanCnv SC) {
	sc.spanCnv = spanCnv
}

// Prepare initializes both the generator and converter components.
func (sc *SpanConverter[SG, SC]) Prepare() {
	if !reflect.ValueOf(sc.spanGen).IsNil() {
		sc.spanGen.Prepare()
	}
	if !reflect.ValueOf(sc.spanCnv).IsNil() {
		sc.spanCnv.Prepare()
	}
}

// Generate performs two-stage span processing:
// 1. First, the generator fills the colors array
// 2. Then, the converter applies its transformation to the colors
func (sc *SpanConverter[SG, SC]) Generate(colors []interface{}, x, y, len int) {
	// Stage 1: Generate colors using the span generator
	if !reflect.ValueOf(sc.spanGen).IsNil() {
		sc.spanGen.Generate(colors, x, y, len)
	}

	// Stage 2: Apply conversion to the generated colors
	if !reflect.ValueOf(sc.spanCnv).IsNil() {
		sc.spanCnv.Generate(colors, x, y, len)
	}
}

// AlphaConverterSpan is a simple span converter that applies alpha blending.
// This serves as an example of a span converter implementation.
type AlphaConverterSpan struct {
	alpha float64 // Alpha value (0.0 to 1.0)
}

// NewAlphaConverterSpan creates a new alpha converter with the given alpha value.
func NewAlphaConverterSpan(alpha float64) *AlphaConverterSpan {
	if alpha < 0.0 {
		alpha = 0.0
	}
	if alpha > 1.0 {
		alpha = 1.0
	}
	return &AlphaConverterSpan{
		alpha: alpha,
	}
}

// SetAlpha sets the alpha value for conversion.
func (ac *AlphaConverterSpan) SetAlpha(alpha float64) {
	if alpha < 0.0 {
		alpha = 0.0
	}
	if alpha > 1.0 {
		alpha = 1.0
	}
	ac.alpha = alpha
}

// Alpha returns the current alpha value.
func (ac *AlphaConverterSpan) Alpha() float64 {
	return ac.alpha
}

// Prepare is called before rendering begins.
// For alpha conversion, no preparation is needed.
func (ac *AlphaConverterSpan) Prepare() {
	// Nothing to prepare for alpha conversion
}

// Generate applies alpha blending to the colors array.
// This is a simplified implementation that works with any color type.
func (ac *AlphaConverterSpan) Generate(colors []interface{}, x, y, len int) {
	// This is a simplified alpha application - in a real implementation,
	// you would need to handle specific color types and their alpha channels
	// For now, we just store the alpha as metadata or apply it conceptually

	// In practice, this would depend on the color type:
	// - For RGBA types, multiply the alpha channel
	// - For RGB types, might need to convert to RGBA first
	// - The exact implementation depends on the color system being used

	// For this generic implementation, we don't modify the colors directly
	// as we don't know their specific type, but in a real converter,
	// you would cast to the appropriate color type and modify accordingly
	_ = len // Use the parameters to avoid unused variable warnings
	_ = x
	_ = y
}
