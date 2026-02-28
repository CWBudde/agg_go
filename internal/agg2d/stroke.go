// Package agg provides advanced stroke attributes for the AGG2D high-level interface.
// This implements Phase 5: Advanced Rendering - Stroke Attributes from the AGG 2.6 C++ library.
package agg2d

import (
	"agg_go/internal/conv"
	"agg_go/internal/path"
)

// MiterLimit sets the miter limit for line joins.
// The miter limit determines when miter joins are converted to bevel joins.
// A higher value allows sharper corners, lower values create more bevels.
// Default value is typically 4.0, which corresponds to 28.96 degrees.
// This matches the C++ Agg2D behavior, though not explicitly exposed in original API.
func (agg2d *Agg2D) MiterLimit(ml float64) {
	if agg2d.convStroke != nil {
		agg2d.convStroke.SetMiterLimit(ml)
	}
}

// GetMiterLimit returns the current miter limit.
func (agg2d *Agg2D) GetMiterLimit() float64 {
	if agg2d.convStroke != nil {
		return agg2d.convStroke.MiterLimit()
	}
	return 4.0 // Default AGG miter limit
}

// MiterLimitTheta sets the miter limit from an angle in radians.
// This is a convenience method that converts an angle to the corresponding miter limit.
// Smaller angles result in higher miter limits (sharper corners allowed).
func (agg2d *Agg2D) MiterLimitTheta(theta float64) {
	if agg2d.convStroke != nil {
		agg2d.convStroke.SetMiterLimitTheta(theta)
	}
}

// InnerMiterLimit sets the inner miter limit for inner corners.
// This controls the behavior of inner (concave) corners in stroked paths.
func (agg2d *Agg2D) InnerMiterLimit(ml float64) {
	if agg2d.convStroke != nil {
		agg2d.convStroke.SetInnerMiterLimit(ml)
	}
}

// GetInnerMiterLimit returns the current inner miter limit.
func (agg2d *Agg2D) GetInnerMiterLimit() float64 {
	if agg2d.convStroke != nil {
		return agg2d.convStroke.InnerMiterLimit()
	}
	return 1.01 // Default AGG inner miter limit
}

// DashPattern manages dash patterns for stroked lines.
// These methods provide high-level access to dash functionality.

// AddDash adds a dash pattern to the current stroke.
// dashLen: length of the dash segment
// gapLen: length of the gap between dashes
// Multiple calls accumulate dash patterns to create complex dash sequences.
// This matches the functionality available in AGG's conv_dash.
func (agg2d *Agg2D) AddDash(dashLen, gapLen float64) {
	// Initialize dash converter if not already present
	if agg2d.convDash == nil {
		// We need to modify the current stroke pipeline to include dashing
		agg2d.initializeDashing()
	}

	if agg2d.convDash != nil {
		agg2d.convDash.AddDash(dashLen, gapLen)
	}
}

// RemoveAllDashes clears all dash patterns, returning to solid line rendering.
func (agg2d *Agg2D) RemoveAllDashes() {
	if agg2d.convDash != nil {
		agg2d.convDash.RemoveAllDashes()
	}
	// Note: Keep dash converter for potential future use, just clear patterns
}

// DashStart sets the starting offset for the dash pattern.
// This allows animation or positioning of the dash pattern along the line.
// offset: distance along the line where the dash pattern begins
func (agg2d *Agg2D) DashStart(offset float64) {
	if agg2d.convDash != nil {
		agg2d.convDash.DashStart(offset)
	}
}

// GetDashStart returns the current dash start offset.
func (agg2d *Agg2D) GetDashStart() float64 {
	if agg2d.convDash != nil {
		return agg2d.convDash.GetDashStart()
	}
	return 0.0
}

// Shorten sets the path shortening distance for strokes.
// This shortens the path by the specified amount at both ends.
// Useful for creating gaps between connected path segments.
func (agg2d *Agg2D) Shorten(s float64) {
	if agg2d.convStroke != nil {
		agg2d.convStroke.SetShorten(s)
	}
}

// GetShorten returns the current path shortening distance.
func (agg2d *Agg2D) GetShorten() float64 {
	if agg2d.convStroke != nil {
		return agg2d.convStroke.Shorten()
	}
	return 0.0
}

// Private helper methods

// initializeDashing sets up the dashing pipeline if not already initialized.
// This modifies the rendering pipeline to: Path -> Curve -> Dash -> Stroke
func (agg2d *Agg2D) initializeDashing() {
	if agg2d.convDash != nil {
		return // Already initialized
	}

	// Create dash converter that operates on the curve converter
	pathAdapter := path.NewPathStorageStlVertexSourceAdapter(agg2d.path)
	agg2d.convCurve = conv.NewConvCurve(pathAdapter)
	agg2d.convDash = conv.NewConvDash(agg2d.convCurve)

	// Recreate stroke converter to operate on dashed output
	agg2d.convStroke = conv.NewConvStroke(agg2d.convDash)

	// Restore current line attributes
	agg2d.convStroke.SetWidth(agg2d.lineWidth)
	agg2d.LineCap(agg2d.lineCap)
	agg2d.LineJoin(agg2d.lineJoin)
}

// ApproximationScale sets the approximation scale for curved segments.
// This affects the quality vs. performance trade-off for curve rendering.
// Higher values produce smoother curves but require more computation.
func (agg2d *Agg2D) ApproximationScale(scale float64) {
	if agg2d.convCurve != nil {
		agg2d.convCurve.SetApproximationScale(scale)
	}
	if agg2d.convStroke != nil {
		agg2d.convStroke.SetApproximationScale(scale)
	}
}

// GetApproximationScale returns the current approximation scale.
func (agg2d *Agg2D) GetApproximationScale() float64 {
	if agg2d.convStroke != nil {
		return agg2d.convStroke.ApproximationScale()
	}
	return 1.0 // Default scale
}

// StrokeAttributes represents a complete set of stroke attributes.
// This allows saving and restoring stroke state.
type StrokeAttributes struct {
	Width              float64
	MiterLimit         float64
	InnerMiterLimit    float64
	Cap                LineCap
	Join               LineJoin
	DashStart          float64
	DashPattern        []float64 // Dash pattern array
	DashOffset         float64   // Dash offset
	PathShorten        float64   // Path shortening
	Shorten            float64
	ApproximationScale float64
}

// GetStrokeAttributes returns the current complete stroke attributes.
func (agg2d *Agg2D) GetStrokeAttributes() StrokeAttributes {
	// Get dash pattern from conv_dash if available
	var dashPattern []float64
	if agg2d.convDash != nil {
		dashPattern = agg2d.getDashPattern()
	}

	return StrokeAttributes{
		Width:              agg2d.lineWidth,
		MiterLimit:         agg2d.GetMiterLimit(),
		InnerMiterLimit:    agg2d.GetInnerMiterLimit(),
		Cap:                agg2d.lineCap,
		Join:               agg2d.lineJoin,
		DashStart:          agg2d.GetDashStart(),
		DashPattern:        dashPattern,
		DashOffset:         agg2d.GetDashStart(), // DashStart is the offset
		PathShorten:        agg2d.GetShorten(),
		Shorten:            agg2d.GetShorten(),
		ApproximationScale: agg2d.GetApproximationScale(),
	}
}

// SetStrokeAttributes sets all stroke attributes at once.
// This is useful for quickly switching between different stroke styles.
func (agg2d *Agg2D) SetStrokeAttributes(attrs StrokeAttributes) {
	agg2d.LineWidth(attrs.Width)
	agg2d.MiterLimit(attrs.MiterLimit)
	agg2d.InnerMiterLimit(attrs.InnerMiterLimit)
	agg2d.LineCap(attrs.Cap)
	agg2d.LineJoin(attrs.Join)
	agg2d.DashStart(attrs.DashStart)
	agg2d.Shorten(attrs.Shorten)
	agg2d.ApproximationScale(attrs.ApproximationScale)
}

// GetLineWidth is now defined in rendering.go to avoid duplication

// NoDashes is an alias for RemoveAllDashes() for compatibility.
// This matches AGG C++ naming conventions used in some examples.
func (agg2d *Agg2D) NoDashes() {
	agg2d.RemoveAllDashes()
}

// getDashPattern returns the current dash pattern array.
// This is a helper method for GetStrokeAttributes.
func (agg2d *Agg2D) getDashPattern() []float64 {
	if agg2d.convDash == nil {
		return nil
	}

	// For now, return empty slice as we don't have direct access to the pattern
	// In a full implementation, conv_dash would expose its pattern
	// This would require extending the conv.ConvDash interface
	return []float64{}
}
