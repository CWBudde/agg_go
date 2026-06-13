// Package agg2d float viewport + coordinate-mapping methods (L5/breadth).
// Color-agnostic delegations mirrored onto Agg2DFloat; the bodies match
// transform.go / utilities.go / rendering.go because they only manipulate the
// shared world transform, clip box, and resample policy. The package-level
// viewportTransform helper is reused verbatim.
package agg2d

import (
	"math"

	"github.com/cwbudde/agg_go/internal/transform"
)

// Viewport sets up a viewport transformation mapping world coordinates onto a
// screen rectangle. Mirrors transform.go.
func (a *Agg2DFloat) Viewport(worldX1, worldY1, worldX2, worldY2,
	screenX1, screenY1, screenX2, screenY2 float64, opt ViewportOption,
) {
	if vt := viewportTransform(worldX1, worldY1, worldX2, worldY2, screenX1, screenY1, screenX2, screenY2, opt); vt != nil {
		a.Affine(vt)
	}
}

// GetViewportTransform calculates the viewport transform without applying it.
func (a *Agg2DFloat) GetViewportTransform(worldX1, worldY1, worldX2, worldY2,
	screenX1, screenY1, screenX2, screenY2 float64, opt ViewportOption,
) *transform.TransAffine {
	return viewportTransform(worldX1, worldY1, worldX2, worldY2, screenX1, screenY1, screenX2, screenY2, opt)
}

// GetScaling returns the average overall scaling factor of the current
// transform. Mirrors transform.go.
func (a *Agg2DFloat) GetScaling() float64 {
	sx, sy := a.transform.GetScaling()
	return (sx + sy) * 0.5
}

// WorldToScreenDistance transforms a distance from world to screen units
// (accounts for scaling, ignores translation). Mirrors transform.go.
func (a *Agg2DFloat) WorldToScreenDistance(worldDistance float64) float64 {
	return worldDistance * a.GetScaling()
}

// ScreenToWorldDistance transforms a distance from screen to world units.
// Returns false if the transform is not invertible. Mirrors transform.go.
func (a *Agg2DFloat) ScreenToWorldDistance(screenDistance float64) (worldDistance float64, ok bool) {
	scaling := a.GetScaling()
	if scaling == 0 {
		return 0, false
	}
	return screenDistance / scaling, true
}

// AlignPoint snaps a world point to pixel-centre boundaries for crisp
// rendering. Mirrors utilities.go.
func (a *Agg2DFloat) AlignPoint(x, y *float64) {
	if x == nil || y == nil {
		return
	}
	a.WorldToScreen(x, y)
	*x = math.Floor(*x) + 0.5
	*y = math.Floor(*y) + 0.5
	a.ScreenToWorld(x, y)
}

// InBox reports whether a world point lies inside the current clip box.
// Mirrors utilities.go.
func (a *Agg2DFloat) InBox(worldX, worldY float64) bool {
	a.WorldToScreen(&worldX, &worldY)
	if a.renBase != nil {
		return a.renBase.rendererBase().InBox(int(worldX), int(worldY))
	}
	return int(worldX) >= int(a.clipBox.X1) && int(worldX) <= int(a.clipBox.X2) &&
		int(worldY) >= int(a.clipBox.Y1) && int(worldY) <= int(a.clipBox.Y2)
}

// AffineImageResamplePolicy controls how affine image transforms choose between
// direct filtered spans and the affine resampler. Mirrors rendering.go.
func (a *Agg2DFloat) AffineImageResamplePolicy(policy AffineImageResamplePolicy) {
	a.affineImageResamplePolicy = policy
}

// GetAffineImageResamplePolicy returns the current affine image resample policy.
func (a *Agg2DFloat) GetAffineImageResamplePolicy() AffineImageResamplePolicy {
	return a.affineImageResamplePolicy
}
