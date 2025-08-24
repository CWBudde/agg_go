// Package span provides span conversion functionality for AGG rendering.
// This file implements the span converter pipeline for two-stage span processing.
package span

import (
	"reflect"

	"agg_go/internal/basics"
	"agg_go/internal/color"
)

// SpanConverterInterface defines the interface for span converters.
// A span converter takes an existing span and applies some transformation to it.
type SpanConverterInterface[C SpanColorType] interface {
	// Prepare is called before rendering begins
	Prepare()

	// Generate applies conversion to the provided colors array for the given span
	Generate(colors []C, x, y, len int)
}

// SpanConverter implements a pipeline for span processing that chains
// a span generator with a span converter in a two-stage process.
// This corresponds to AGG's span_converter<SpanGenerator, SpanConverter> template class.
type SpanConverter[C SpanColorType, SG SpanGenerator[C], SC SpanConverterInterface[C]] struct {
	spanGen SG // The span generator
	spanCnv SC // The span converter
}

// NewSpanConverter creates a new span converter with the given generator and converter.
func NewSpanConverter[C SpanColorType, SG SpanGenerator[C], SC SpanConverterInterface[C]](spanGen SG, spanCnv SC) *SpanConverter[C, SG, SC] {
	return &SpanConverter[C, SG, SC]{
		spanGen: spanGen,
		spanCnv: spanCnv,
	}
}

// NewSpanConverterEmpty creates a new empty span converter.
// Components must be attached later using AttachGenerator and AttachConverter.
func NewSpanConverterEmpty[C SpanColorType, SG SpanGenerator[C], SC SpanConverterInterface[C]]() *SpanConverter[C, SG, SC] {
	var nilGen SG
	var nilCnv SC
	return &SpanConverter[C, SG, SC]{
		spanGen: nilGen,
		spanCnv: nilCnv,
	}
}

// AttachGenerator attaches or changes the span generator.
func (sc *SpanConverter[C, SG, SC]) AttachGenerator(spanGen SG) {
	sc.spanGen = spanGen
}

// AttachConverter attaches or changes the span converter.
func (sc *SpanConverter[C, SG, SC]) AttachConverter(spanCnv SC) {
	sc.spanCnv = spanCnv
}

// Prepare initializes both the generator and converter components.
func (sc *SpanConverter[C, SG, SC]) Prepare() {
	// Check if span generator is not zero value using reflection
	if !reflect.ValueOf(sc.spanGen).IsZero() {
		sc.spanGen.Prepare()
	}
	// Check if span converter is not zero value using reflection
	if !reflect.ValueOf(sc.spanCnv).IsZero() {
		sc.spanCnv.Prepare()
	}
}

// Generate performs two-stage span processing:
// 1. First, the generator fills the colors array
// 2. Then, the converter applies its transformation to the colors
func (sc *SpanConverter[C, SG, SC]) Generate(colors []C, x, y, len int) {
	// Stage 1: Generate colors using the span generator
	if !reflect.ValueOf(sc.spanGen).IsZero() {
		sc.spanGen.Generate(colors, x, y, len)
	}

	// Stage 2: Apply conversion to the generated colors
	if !reflect.ValueOf(sc.spanCnv).IsZero() {
		sc.spanCnv.Generate(colors, x, y, len)
	}
}

// AlphaConverterSpan is a span converter that applies alpha blending.
// This implementation works with any color type that has an alpha channel.
type AlphaConverterSpan[C SpanColorType] struct {
	alpha float64 // Alpha value (0.0 to 1.0)
}

// NewAlphaConverterSpan creates a new alpha converter with the given alpha value.
func NewAlphaConverterSpan[C SpanColorType](alpha float64) *AlphaConverterSpan[C] {
	if alpha < 0.0 {
		alpha = 0.0
	}
	if alpha > 1.0 {
		alpha = 1.0
	}
	return &AlphaConverterSpan[C]{
		alpha: alpha,
	}
}

// SetAlpha sets the alpha value for conversion.
func (ac *AlphaConverterSpan[C]) SetAlpha(alpha float64) {
	if alpha < 0.0 {
		alpha = 0.0
	}
	if alpha > 1.0 {
		alpha = 1.0
	}
	ac.alpha = alpha
}

// Alpha returns the current alpha value.
func (ac *AlphaConverterSpan[C]) Alpha() float64 {
	return ac.alpha
}

// Prepare is called before rendering begins.
// For alpha conversion, no preparation is needed.
func (ac *AlphaConverterSpan[C]) Prepare() {
	// Nothing to prepare for alpha conversion
}

// Generate applies alpha blending to the colors array.
// This implementation uses type assertion to handle different color types.
func (ac *AlphaConverterSpan[C]) Generate(colors []C, x, y, len int) {
	// Apply alpha to each color in the span
	for i := 0; i < len; i++ {
		ac.applyAlphaToColor(&colors[i])
	}
}

// applyAlphaToColor applies alpha to a single color using type assertion.
// This handles the common color types used in AGG.
func (ac *AlphaConverterSpan[C]) applyAlphaToColor(c *C) {
	// Use type assertion to handle different color types from the AGG color package
	switch color := any(c).(type) {
	case *color.RGBA8[color.SRGB]:
		// For 8-bit RGBA with SRGB colorspace
		color.A = basics.Int8u(float64(color.A) * ac.alpha)
	case *color.RGBA8[color.Linear]:
		// For 8-bit RGBA with Linear colorspace
		color.A = basics.Int8u(float64(color.A) * ac.alpha)
	case *color.RGBA:
		// For floating-point RGBA
		color.A = color.A * ac.alpha
	default:
		// For color types without alpha channel, alpha conversion is not possible
		// This matches AGG's behavior - alpha conversion only works with RGBA types
	}
}

// BrightnessAlphaConverter implements brightness-based alpha conversion.
// This is a port of the span_conv_brightness_alpha class from the AGG C++ examples.
// It calculates alpha based on the brightness (luminance) of the color.
type BrightnessAlphaConverter[C SpanColorType] struct {
	alphaArray []basics.Int8u // Alpha lookup table based on brightness
}

// NewBrightnessAlphaConverter creates a new brightness-alpha converter.
// alphaArray should be a 768-element array (256 * 3) for full brightness range.
func NewBrightnessAlphaConverter[C SpanColorType](alphaArray []basics.Int8u) *BrightnessAlphaConverter[C] {
	if len(alphaArray) != 768 { // 256 * 3 for full RGB brightness range
		// Create a default linear alpha array if not provided correctly
		alphaArray = make([]basics.Int8u, 768)
		for i := 0; i < 768; i++ {
			alphaArray[i] = basics.Int8u(i * 255 / 767) // Linear mapping
		}
	}

	return &BrightnessAlphaConverter[C]{
		alphaArray: alphaArray,
	}
}

// Prepare is called before rendering begins.
func (bac *BrightnessAlphaConverter[C]) Prepare() {
	// Nothing to prepare for brightness-alpha conversion
}

// Generate applies brightness-based alpha conversion to the colors array.
// This matches the C++ implementation from the image_alpha.cpp example.
func (bac *BrightnessAlphaConverter[C]) Generate(colors []C, x, y, len int) {
	for i := 0; i < len; i++ {
		bac.applyBrightnessAlpha(&colors[i])
	}
}

// applyBrightnessAlpha applies brightness-based alpha to a single color.
func (bac *BrightnessAlphaConverter[C]) applyBrightnessAlpha(c *C) {
	switch color := any(c).(type) {
	case *color.RGBA8[color.SRGB]:
		// Calculate brightness from RGB components
		brightness := int(color.R) + int(color.G) + int(color.B)
		// Map to alpha array index (matches C++ algorithm)
		index := brightness * 767 / (3 * 255) // 767 = array_size - 1, 255 = full_value
		if index >= len(bac.alphaArray) {
			index = len(bac.alphaArray) - 1
		}
		// Set alpha based on brightness lookup
		color.A = bac.alphaArray[index]

	case *color.RGBA8[color.Linear]:
		// Same algorithm for Linear colorspace
		brightness := int(color.R) + int(color.G) + int(color.B)
		index := brightness * 767 / (3 * 255)
		if index >= len(bac.alphaArray) {
			index = len(bac.alphaArray) - 1
		}
		color.A = bac.alphaArray[index]

	case *color.RGBA:
		// For floating-point RGBA, scale to 8-bit range for calculation
		brightness := int(color.R*255) + int(color.G*255) + int(color.B*255)
		index := brightness * 767 / (3 * 255)
		if index >= len(bac.alphaArray) {
			index = len(bac.alphaArray) - 1
		}
		// Convert back to floating-point alpha
		color.A = float64(bac.alphaArray[index]) / 255.0

	default:
		// For color types without alpha channel, brightness conversion is not possible
	}
}
