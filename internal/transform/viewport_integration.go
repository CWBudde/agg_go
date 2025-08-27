// Package transform provides viewport integration functionality for AGG.
// This implements integration helpers for viewport transformations with rendering and path processing.
package transform

import (
	"agg_go/internal/basics"
)

// ViewportVertexSource wraps a vertex source to apply viewport transformations.
// This allows seamless integration of viewport transformations with path processing.
type ViewportVertexSource struct {
	source   basics.VertexSource
	viewport *TransViewport
}

// NewViewportVertexSource creates a new viewport-transformed vertex source.
func NewViewportVertexSource(source basics.VertexSource, viewport *TransViewport) *ViewportVertexSource {
	return &ViewportVertexSource{
		source:   source,
		viewport: viewport,
	}
}

// Rewind implements the VertexSource interface.
func (vvs *ViewportVertexSource) Rewind(pathID uint) {
	vvs.source.Rewind(pathID)
}

// Vertex implements the VertexSource interface.
// It retrieves vertices from the wrapped source and applies viewport transformation.
func (vvs *ViewportVertexSource) Vertex() (x, y float64, cmd basics.PathCommand) {
	x, y, cmd = vvs.source.Vertex()

	// Only transform coordinate commands, not path control commands
	if basics.IsVertex(cmd) {
		vvs.viewport.Transform(&x, &y)
	}

	return x, y, cmd
}

// SetViewport changes the viewport used for transformations.
func (vvs *ViewportVertexSource) SetViewport(viewport *TransViewport) {
	vvs.viewport = viewport
}

// GetViewport returns the current viewport.
func (vvs *ViewportVertexSource) GetViewport() *TransViewport {
	return vvs.viewport
}

// ViewportRenderer provides viewport-aware rendering capabilities.
// This is a helper for integrating viewports with various renderer types.
type ViewportRenderer struct {
	viewport *TransViewport
}

// NewViewportRenderer creates a new viewport renderer helper.
func NewViewportRenderer(viewport *TransViewport) *ViewportRenderer {
	return &ViewportRenderer{
		viewport: viewport,
	}
}

// TransformPoint applies the viewport transformation to a single point.
func (vr *ViewportRenderer) TransformPoint(x, y *float64) {
	vr.viewport.Transform(x, y)
}

// TransformBounds applies the viewport transformation to a bounding rectangle.
func (vr *ViewportRenderer) TransformBounds(x1, y1, x2, y2 *float64) {
	vr.viewport.Transform(x1, y1)
	vr.viewport.Transform(x2, y2)

	// Ensure bounds are ordered correctly after transformation
	if *x1 > *x2 {
		*x1, *x2 = *x2, *x1
	}
	if *y1 > *y2 {
		*y1, *y2 = *y2, *y1
	}
}

// GetTransformMatrix returns the viewport transformation as an affine matrix.
func (vr *ViewportRenderer) GetTransformMatrix() *TransAffine {
	return vr.viewport.ToAffine()
}

// SetViewport changes the viewport used by this renderer helper.
func (vr *ViewportRenderer) SetViewport(viewport *TransViewport) {
	vr.viewport = viewport
}

// ViewportUtilities provides common utility functions for viewport operations.
type ViewportUtilities struct{}

// GetViewportUtils returns a shared instance of viewport utilities.
func GetViewportUtils() *ViewportUtilities {
	return &ViewportUtilities{}
}

// CalculateOptimalZoom calculates the optimal zoom factor to fit content within the viewport.
// contentBounds: [minX, minY, maxX, maxY] of the content to fit
// viewportBounds: [minX, minY, maxX, maxY] of the viewport
// Returns the zoom factor and center point for optimal fitting.
func (vu *ViewportUtilities) CalculateOptimalZoom(contentBounds, viewportBounds [4]float64) (zoomFactor, centerX, centerY float64) {
	contentWidth := contentBounds[2] - contentBounds[0]
	contentHeight := contentBounds[3] - contentBounds[1]
	viewportWidth := viewportBounds[2] - viewportBounds[0]
	viewportHeight := viewportBounds[3] - viewportBounds[1]

	if contentWidth <= 0 || contentHeight <= 0 {
		return 1.0, 0.0, 0.0
	}

	// Calculate scale factors for both dimensions
	scaleX := viewportWidth / contentWidth
	scaleY := viewportHeight / contentHeight

	// Use the smaller scale to ensure content fits entirely
	zoomFactor = scaleX
	if scaleY < scaleX {
		zoomFactor = scaleY
	}

	// Calculate center point of content
	centerX = (contentBounds[0] + contentBounds[2]) * 0.5
	centerY = (contentBounds[1] + contentBounds[3]) * 0.5

	return zoomFactor, centerX, centerY
}

// InterpolateViewports creates smooth transitions between two viewports.
// t should be in the range [0.0, 1.0], where 0.0 returns viewport1 and 1.0 returns viewport2.
func (vu *ViewportUtilities) InterpolateViewports(viewport1, viewport2 *TransViewport, t float64) *TransViewport {
	// Clamp t to valid range
	if t < 0.0 {
		t = 0.0
	} else if t > 1.0 {
		t = 1.0
	}

	// Get world and device bounds from both viewports
	w1x1, w1y1, w1x2, w1y2 := viewport1.GetWorldViewport()
	d1x1, d1y1, d1x2, d1y2 := viewport1.GetDeviceViewport()

	w2x1, w2y1, w2x2, w2y2 := viewport2.GetWorldViewport()
	d2x1, d2y1, d2x2, d2y2 := viewport2.GetDeviceViewport()

	// Interpolate world bounds
	wx1 := w1x1 + (w2x1-w1x1)*t
	wy1 := w1y1 + (w2y1-w1y1)*t
	wx2 := w1x2 + (w2x2-w1x2)*t
	wy2 := w1y2 + (w2y2-w1y2)*t

	// Interpolate device bounds
	dx1 := d1x1 + (d2x1-d1x1)*t
	dy1 := d1y1 + (d2y1-d1y1)*t
	dx2 := d1x2 + (d2x2-d1x2)*t
	dy2 := d1y2 + (d2y2-d1y2)*t

	// Create interpolated viewport
	result := NewTransViewport()
	result.WorldViewport(wx1, wy1, wx2, wy2)
	result.DeviceViewport(dx1, dy1, dx2, dy2)

	// Interpolate aspect ratio and alignment settings
	// For simplicity, we'll use viewport1's settings when t < 0.5, viewport2's otherwise
	if t < 0.5 {
		result.PreserveAspectRatio(viewport1.AlignX(), viewport1.AlignY(), viewport1.AspectRatio())
	} else {
		result.PreserveAspectRatio(viewport2.AlignX(), viewport2.AlignY(), viewport2.AspectRatio())
	}

	return result
}

// IsPointInViewport checks if a world coordinate point is visible in the viewport.
func (vu *ViewportUtilities) IsPointInViewport(viewport *TransViewport, worldX, worldY float64) bool {
	deviceX, deviceY := worldX, worldY
	viewport.Transform(&deviceX, &deviceY)

	dx1, dy1, dx2, dy2 := viewport.GetDeviceViewport()

	return deviceX >= dx1 && deviceX <= dx2 && deviceY >= dy1 && deviceY <= dy2
}

// GetVisibleWorldBounds returns the world coordinate bounds that are visible in the viewport.
func (vu *ViewportUtilities) GetVisibleWorldBounds(viewport *TransViewport) (minX, minY, maxX, maxY float64) {
	dx1, dy1, dx2, dy2 := viewport.GetDeviceViewport()

	// Transform device corners to world coordinates
	minX, minY = dx1, dy1
	maxX, maxY = dx2, dy2

	viewport.InverseTransform(&minX, &minY)
	viewport.InverseTransform(&maxX, &maxY)

	// Ensure proper ordering
	if minX > maxX {
		minX, maxX = maxX, minX
	}
	if minY > maxY {
		minY, maxY = maxY, minY
	}

	return minX, minY, maxX, maxY
}

// ViewportAnimation provides smooth viewport animations.
type ViewportAnimation struct {
	startViewport *TransViewport
	endViewport   *TransViewport
	duration      float64 // Animation duration in seconds
	elapsed       float64 // Elapsed time in seconds
	utils         *ViewportUtilities
}

// NewViewportAnimation creates a new viewport animation.
func NewViewportAnimation(start, end *TransViewport, duration float64) *ViewportAnimation {
	return &ViewportAnimation{
		startViewport: start,
		endViewport:   end,
		duration:      duration,
		elapsed:       0.0,
		utils:         GetViewportUtils(),
	}
}

// Update updates the animation by the given time step (in seconds).
// Returns the current interpolated viewport and whether the animation is complete.
func (va *ViewportAnimation) Update(deltaTime float64) (*TransViewport, bool) {
	va.elapsed += deltaTime

	if va.elapsed >= va.duration {
		va.elapsed = va.duration
		return va.endViewport, true
	}

	// Calculate interpolation parameter with easing
	t := va.elapsed / va.duration
	// Apply smooth easing (smoothstep function)
	t = t * t * (3.0 - 2.0*t)

	return va.utils.InterpolateViewports(va.startViewport, va.endViewport, t), false
}

// Reset resets the animation to the beginning.
func (va *ViewportAnimation) Reset() {
	va.elapsed = 0.0
}

// IsComplete returns true if the animation has finished.
func (va *ViewportAnimation) IsComplete() bool {
	return va.elapsed >= va.duration
}

// GetProgress returns the animation progress as a value between 0.0 and 1.0.
func (va *ViewportAnimation) GetProgress() float64 {
	if va.duration <= 0.0 {
		return 1.0
	}

	progress := va.elapsed / va.duration
	if progress > 1.0 {
		progress = 1.0
	}

	return progress
}
