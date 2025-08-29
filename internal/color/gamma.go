// Package color provides color types and conversion functions for AGG.
// This package implements RGBA, grayscale, and color space conversions.
package color

import (
	"agg_go/internal/basics"
)

// lut8Like defines the minimal interface for 8-bit gamma lookup tables.
// It maps 8-bit channel values (0–255) to corrected 8-bit values.
// Used for both direct (linearization) and inverse (de-linearization) gamma.
type lut8Like interface {
	Dir(basics.Int8u) basics.Int8u // Direct gamma correction
	Inv(basics.Int8u) basics.Int8u // Inverse gamma correction
}

// lut16Like defines the minimal interface for 16-bit gamma lookup tables.
// It uses an 8-bit input (0–255) for Dir() and produces a 16-bit output (0–65535).
// Inv() reverses this mapping: from 16-bit back to 8-bit.
type lut16Like interface {
	Dir(basics.Int8u) basics.Int16u // Direct gamma correction: 8-bit → 16-bit
	Inv(basics.Int16u) basics.Int8u // Inverse gamma correction: 16-bit → 8-bit
}

// lut32Like defines the minimal interface for floating-point gamma functions.
// It operates directly on normalized float32 values in the range [0,1].
// Typically implemented by functional gamma models instead of lookup tables.
type lut32Like interface {
	DirFloat(v float32) float32 // Direct gamma correction
	InvFloat(v float32) float32 // Inverse gamma correction
}
