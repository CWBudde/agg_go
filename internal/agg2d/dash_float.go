// Package agg2d float dash-pattern support (L5, Phase 4.7). Float twin of the
// 8-bit dash methods in stroke.go: AddDash/RemoveAllDashes/DashStart/
// GetDashStart/NoDashes. The conv_dash converter and all dash math are
// color-agnostic and reused verbatim; only the stroke-pipeline rebuild lives
// here so it threads through the float convCurve/convStroke fields.
package agg2d

import (
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/conv"
	"github.com/cwbudde/agg_go/internal/path"
)

// AddDash appends one dash-gap pair to the current dash pattern. The first call
// installs the dash converter into the stroke pipeline (Path -> Curve -> Dash ->
// Stroke); subsequent calls accumulate segments. Mirrors the 8-bit Agg2D.AddDash.
func (a *Agg2DFloat) AddDash(dashLen, gapLen float64) {
	if a.convDash == nil {
		a.initializeDashing()
	}
	if a.convDash != nil {
		a.convDash.AddDash(dashLen, gapLen)
	}
}

// RemoveAllDashes clears every dash segment, returning to solid stroking. The
// dash converter stays in the pipeline; renderStroke bypasses it when
// NumDashes() == 0 (matching the 8-bit path).
func (a *Agg2DFloat) RemoveAllDashes() {
	if a.convDash != nil {
		a.convDash.RemoveAllDashes()
	}
}

// DashStart sets the dash-phase offset along the line.
func (a *Agg2DFloat) DashStart(offset float64) {
	if a.convDash != nil {
		a.convDash.DashStart(offset)
	}
}

// GetDashStart returns the current dash-phase offset (0 when no dash converter
// is installed).
func (a *Agg2DFloat) GetDashStart() float64 {
	if a.convDash != nil {
		return a.convDash.GetDashStart()
	}
	return 0.0
}

// NoDashes is the AGG C++-style alias for RemoveAllDashes.
func (a *Agg2DFloat) NoDashes() {
	a.RemoveAllDashes()
}

// initializeDashing rebuilds the stroke pipeline to insert a conv_dash stage
// between the curve converter and the stroke converter, preserving the current
// stroke state. Mirrors the 8-bit Agg2D.initializeDashing.
func (a *Agg2DFloat) initializeDashing() {
	if a.convDash != nil {
		return
	}

	// Preserve the current stroke state before recreating the converters.
	width := a.lineWidth
	lineCap := a.lineCap
	lineJoin := a.lineJoin
	miterLimit, innerMiterLimit, approximationScale, shorten := 4.0, 1.01, 1.0, 0.0
	if a.convStroke != nil {
		miterLimit = a.convStroke.MiterLimit()
		innerMiterLimit = a.convStroke.InnerMiterLimit()
		approximationScale = a.convStroke.ApproximationScale()
		shorten = a.convStroke.Shorten()
	}

	// Insert the dash converter on top of the curve converter, then stroke the
	// dashed output: Path -> Curve -> Dash -> Stroke.
	pathAdapter := path.NewPathStorageStlVertexSourceAdapter(a.path)
	a.convCurve = conv.NewConvCurve(pathAdapter)
	a.convDash = conv.NewConvDash(a.convCurve)
	a.convStroke = conv.NewConvStroke(a.convDash)

	// Restore the stroke attributes onto the new stroke converter.
	a.convStroke.SetWidth(width)
	a.convStroke.SetLineCap(basics.LineCap(lineCap))
	a.convStroke.SetLineJoin(basics.LineJoin(lineJoin))
	a.convStroke.SetMiterLimit(miterLimit)
	a.convStroke.SetInnerMiterLimit(innerMiterLimit)
	a.convStroke.SetApproximationScale(approximationScale)
	a.convStroke.SetShorten(shorten)
}
