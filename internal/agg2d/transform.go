// Package agg provides transformation functionality for the AGG2D high-level interface.
// This implements Phase 5: Advanced Rendering - Transformations from the AGG 2.6 C++ library.
package agg2d

import (
	"math"

	"agg_go/internal/transform"
)

// Transformations represents a complete transformation matrix.
// This matches the C++ Agg2D::Transformations structure.
type Transformations struct {
	AffineMatrix [6]float64 // [sx, shy, shx, sy, tx, ty]
}

// NewTransformations creates a new identity transformation.
func NewTransformations() *Transformations {
	return &Transformations{
		AffineMatrix: [6]float64{1.0, 0.0, 0.0, 1.0, 0.0, 0.0}, // Identity matrix
	}
}

// GetTransformations returns the current transformation matrix.
// This matches the C++ Agg2D::transformations() const method.
func (agg2d *Agg2D) GetTransformations() *Transformations {
	// Extract matrix components from the internal TransAffine
	sx := agg2d.transform.SX
	shy := agg2d.transform.SHY
	shx := agg2d.transform.SHX
	sy := agg2d.transform.SY
	tx := agg2d.transform.TX
	ty := agg2d.transform.TY

	return &Transformations{
		AffineMatrix: [6]float64{sx, shy, shx, sy, tx, ty},
	}
}

// SetTransformations sets the transformation matrix from a Transformations struct.
// This matches the C++ Agg2D::transformations(const Transformations& tr) method.
func (agg2d *Agg2D) SetTransformations(tr *Transformations) {
	// Set matrix components directly
	agg2d.transform.SX = tr.AffineMatrix[0]
	agg2d.transform.SHY = tr.AffineMatrix[1]
	agg2d.transform.SHX = tr.AffineMatrix[2]
	agg2d.transform.SY = tr.AffineMatrix[3]
	agg2d.transform.TX = tr.AffineMatrix[4]
	agg2d.transform.TY = tr.AffineMatrix[5]

	// Update approximation scales for converters
	agg2d.updateApproximationScales()
}

// Affine applies an affine transformation to the current transformation matrix.
// This multiplies the current matrix by the provided transformation.
// This matches the C++ Agg2D::affine(const Affine& tr) method.
func (agg2d *Agg2D) Affine(tr *transform.TransAffine) {
	agg2d.transform.Multiply(tr)
	agg2d.updateApproximationScales()
}

// AffineFromMatrix applies an affine transformation from a Transformations struct.
// This matches the C++ Agg2D::affine(const Transformations& tr) method.
func (agg2d *Agg2D) AffineFromMatrix(tr *Transformations) {
	affine := transform.NewTransAffineFromValues(
		tr.AffineMatrix[0], tr.AffineMatrix[1], tr.AffineMatrix[2],
		tr.AffineMatrix[3], tr.AffineMatrix[4], tr.AffineMatrix[5])
	agg2d.Affine(affine)
}

// Rotate applies a rotation transformation.
// angle: rotation angle in radians (positive = counter-clockwise)
// This matches the C++ Agg2D::rotate(double angle) method.
func (agg2d *Agg2D) Rotate(angle float64) {
	agg2d.transform.Rotate(angle)
	agg2d.updateApproximationScales()
}

// Scale applies a scaling transformation.
// sx: horizontal scale factor
// sy: vertical scale factor
// This matches the C++ Agg2D::scale(double sx, double sy) method.
func (agg2d *Agg2D) Scale(sx, sy float64) {
	agg2d.transform.ScaleXY(sx, sy)
	agg2d.updateApproximationScales()
}

// UniformScale applies uniform scaling (same factor for both axes).
// s: scale factor for both horizontal and vertical
func (agg2d *Agg2D) UniformScale(s float64) {
	agg2d.Scale(s, s)
}

// Skew applies a skewing transformation.
// sx: horizontal skew factor (radians)
// sy: vertical skew factor (radians)
// This matches the C++ Agg2D::skew(double sx, double sy) method.
func (agg2d *Agg2D) Skew(sx, sy float64) {
	// Create skew transformation: [1, tan(sy), tan(sx), 1, 0, 0]
	skewTransform := transform.NewTransAffineFromValues(
		1.0, math.Tan(sy), math.Tan(sx), 1.0, 0.0, 0.0)
	agg2d.Affine(skewTransform)
}

// Translate applies a translation transformation.
// x: horizontal translation
// y: vertical translation
// This matches the C++ Agg2D::translate(double x, double y) method.
func (agg2d *Agg2D) Translate(x, y float64) {
	agg2d.transform.Translate(x, y)
	// Translation doesn't affect approximation scale
}

// worldToScreenScalar converts a scalar distance through the transformation.
// This is used internally to calculate approximation scales.
func (agg2d *Agg2D) worldToScreenScalar(scalar float64) float64 {
	// Transform two points separated by scalar distance
	x1, y1 := 0.0, 0.0
	x2, y2 := scalar, scalar

	agg2d.transform.Transform(&x1, &y1)
	agg2d.transform.Transform(&x2, &y2)

	// Calculate the distance between transformed points
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx+dy*dy) / math.Sqrt(2.0) // Normalize for diagonal
}

// Transform stack management

// PushTransform saves the current transformation state.
func (agg2d *Agg2D) PushTransform() {
	if agg2d.transformStack == nil {
		agg2d.transformStack = &TransformStack{
			stack: make([]*transform.TransAffine, 0),
		}
	}

	// Create a copy of the current transformation
	copy := transform.NewTransAffineFromValues(
		agg2d.transform.SX, agg2d.transform.SHY, agg2d.transform.SHX,
		agg2d.transform.SY, agg2d.transform.TX, agg2d.transform.TY)

	agg2d.transformStack.stack = append(agg2d.transformStack.stack, copy)
}

// PopTransform restores the most recently saved transformation state.
// Returns false if no transformation was saved.
func (agg2d *Agg2D) PopTransform() bool {
	if agg2d.transformStack == nil || len(agg2d.transformStack.stack) == 0 {
		return false
	}

	// Get the last saved transformation
	stack := agg2d.transformStack.stack
	lastIndex := len(stack) - 1
	savedTransform := stack[lastIndex]

	// Remove from stack
	agg2d.transformStack.stack = stack[:lastIndex]

	// Restore the transformation
	agg2d.transform.SX = savedTransform.SX
	agg2d.transform.SHY = savedTransform.SHY
	agg2d.transform.SHX = savedTransform.SHX
	agg2d.transform.SY = savedTransform.SY
	agg2d.transform.TX = savedTransform.TX
	agg2d.transform.TY = savedTransform.TY

	agg2d.updateApproximationScales()
	return true
}

// GetTransformStackDepth returns the number of saved transformations.
func (agg2d *Agg2D) GetTransformStackDepth() int {
	if agg2d.transformStack == nil {
		return 0
	}
	return len(agg2d.transformStack.stack)
}

// Transformation queries and utilities

// GetScaling returns the overall scaling factor of the current transformation.
func (agg2d *Agg2D) GetScaling() float64 {
	sx, sy := agg2d.transform.GetScaling()
	// Return average scaling factor
	return (sx + sy) / 2.0
}

// GetRotation returns the rotation component of the current transformation in radians.
func (agg2d *Agg2D) GetRotation() float64 {
	return agg2d.transform.GetRotation()
}

// GetTranslation returns the translation components of the current transformation.
func (agg2d *Agg2D) GetTranslation() (x, y float64) {
	return agg2d.transform.TX, agg2d.transform.TY
}

// IsIdentity returns true if the current transformation is the identity matrix.
func (agg2d *Agg2D) IsIdentity() bool {
	return agg2d.transform.IsIdentity(1e-10)
}

// IsTranslationOnly returns true if the transformation only includes translation.
func (agg2d *Agg2D) IsTranslationOnly() bool {
	return agg2d.transform.SX == 1.0 && agg2d.transform.SY == 1.0 &&
		agg2d.transform.SHX == 0.0 && agg2d.transform.SHY == 0.0
}

// Determinant returns the determinant of the transformation matrix.
// Zero determinant indicates a singular (non-invertible) transformation.
func (agg2d *Agg2D) Determinant() float64 {
	return agg2d.transform.Determinant()
}

// IsValid returns true if the transformation is valid (non-singular).
func (agg2d *Agg2D) IsValid() bool {
	return agg2d.transform.IsValid(1e-10)
}

// Convenient transformation builders

// RotateAround applies rotation around a specific point.
// centerX, centerY: point to rotate around
// angle: rotation angle in radians
func (agg2d *Agg2D) RotateAround(centerX, centerY, angle float64) {
	agg2d.Translate(-centerX, -centerY)
	agg2d.Rotate(angle)
	agg2d.Translate(centerX, centerY)
}

// ScaleAround applies scaling around a specific point.
// centerX, centerY: point to scale around
// sx, sy: scale factors
func (agg2d *Agg2D) ScaleAround(centerX, centerY, sx, sy float64) {
	agg2d.Translate(-centerX, -centerY)
	agg2d.Scale(sx, sy)
	agg2d.Translate(centerX, centerY)
}

// UniformScaleAround applies uniform scaling around a specific point.
func (agg2d *Agg2D) UniformScaleAround(centerX, centerY, scale float64) {
	agg2d.ScaleAround(centerX, centerY, scale, scale)
}

// FlipHorizontal flips the coordinate system horizontally around a vertical axis.
func (agg2d *Agg2D) FlipHorizontal(axisX float64) {
	agg2d.Translate(-axisX, 0)
	agg2d.Scale(-1.0, 1.0)
	agg2d.Translate(axisX, 0)
}

// FlipVertical flips the coordinate system vertically around a horizontal axis.
func (agg2d *Agg2D) FlipVertical(axisY float64) {
	agg2d.Translate(0, -axisY)
	agg2d.Scale(1.0, -1.0)
	agg2d.Translate(0, axisY)
}

// Matrix decomposition and analysis

// DecomposeTransform decomposes the transformation matrix into its basic components.
type TransformComponents struct {
	ScaleX     float64 // Horizontal scaling
	ScaleY     float64 // Vertical scaling
	Rotation   float64 // Rotation in radians
	TranslateX float64 // Horizontal translation
	TranslateY float64 // Vertical translation
	SkewX      float64 // Horizontal skew
	SkewY      float64 // Vertical skew
}

// DecomposeTransform analyzes the current transformation and extracts a
// canonical set of affine components for this Go extension API.
func (agg2d *Agg2D) DecomposeTransform() *TransformComponents {
	tx := agg2d.transform.TX
	ty := agg2d.transform.TY
	rotation := agg2d.transform.GetRotation()

	// Remove rotation so the remaining 2x2 matrix captures scale and skew.
	linear := transform.NewTransAffineFromValues(
		agg2d.transform.SX, agg2d.transform.SHY,
		agg2d.transform.SHX, agg2d.transform.SY,
		0.0, 0.0,
	)
	linear.Multiply(transform.NewTransAffineRotation(-rotation))

	scaleX := linear.SX
	scaleY := linear.SY

	skewX := 0.0
	if scaleY != 0.0 {
		skewX = math.Atan2(linear.SHX, scaleY)
	}

	skewY := 0.0
	if scaleX != 0.0 {
		skewY = math.Atan2(linear.SHY, scaleX)
	}

	return &TransformComponents{
		ScaleX:     scaleX,
		ScaleY:     scaleY,
		Rotation:   rotation,
		TranslateX: tx,
		TranslateY: ty,
		SkewX:      skewX,
		SkewY:      skewY,
	}
}

// InvertTransform returns the inverse of the current transformation.
// Returns nil if the transformation is not invertible.
func (agg2d *Agg2D) InvertTransform() *transform.TransAffine {
	inverse := transform.NewTransAffineFromValues(
		agg2d.transform.SX, agg2d.transform.SHY, agg2d.transform.SHX,
		agg2d.transform.SY, agg2d.transform.TX, agg2d.transform.TY)

	if inverse.Invert() != nil {
		return inverse
	}
	return nil
}

// Utility functions for angle conversion

// Viewport transformations

func viewportTransform(worldX1, worldY1, worldX2, worldY2,
	screenX1, screenY1, screenX2, screenY2 float64, opt ViewportOption,
) *transform.TransAffine {
	worldWidth := worldX2 - worldX1
	worldHeight := worldY2 - worldY1
	screenWidth := screenX2 - screenX1
	screenHeight := screenY2 - screenY1

	if worldWidth == 0 || worldHeight == 0 || screenWidth == 0 || screenHeight == 0 {
		return nil
	}

	scaleX := screenWidth / worldWidth
	scaleY := screenHeight / worldHeight

	var finalScaleX, finalScaleY, offsetX, offsetY float64

	switch opt {
	case Anisotropic:
		finalScaleX = scaleX
		finalScaleY = scaleY
		offsetX = screenX1 - worldX1*scaleX
		offsetY = screenY1 - worldY1*scaleY
	default:
		uniformScale := scaleX
		if scaleY < uniformScale {
			uniformScale = scaleY
		}

		finalScaleX = uniformScale
		finalScaleY = uniformScale

		alignedScreenWidth := worldWidth * uniformScale
		alignedScreenHeight := worldHeight * uniformScale

		var alignX, alignY float64
		switch opt {
		case XMinYMin, XMinYMid, XMinYMax:
			alignX = screenX1
		case XMidYMin, XMidYMid, XMidYMax:
			alignX = screenX1 + (screenWidth-alignedScreenWidth)/2
		case XMaxYMin, XMaxYMid, XMaxYMax:
			alignX = screenX2 - alignedScreenWidth
		default:
			alignX = screenX1 + (screenWidth-alignedScreenWidth)/2
		}

		switch opt {
		case XMinYMin, XMidYMin, XMaxYMin:
			alignY = screenY1
		case XMinYMid, XMidYMid, XMaxYMid:
			alignY = screenY1 + (screenHeight-alignedScreenHeight)/2
		case XMinYMax, XMidYMax, XMaxYMax:
			alignY = screenY2 - alignedScreenHeight
		default:
			alignY = screenY1 + (screenHeight-alignedScreenHeight)/2
		}

		offsetX = alignX - worldX1*uniformScale
		offsetY = alignY - worldY1*uniformScale
	}

	return transform.NewTransAffineFromValues(finalScaleX, 0, 0, finalScaleY, offsetX, offsetY)
}

// Viewport sets up a viewport transformation.
// This maps world coordinates to screen coordinates within a viewport.
// worldX1, worldY1, worldX2, worldY2: world coordinate bounds
// screenX1, screenY1, screenX2, screenY2: screen coordinate bounds
// opt: viewport scaling option (how to handle aspect ratio differences)
// This matches the C++ Agg2D::viewport method.
func (agg2d *Agg2D) Viewport(worldX1, worldY1, worldX2, worldY2,
	screenX1, screenY1, screenX2, screenY2 float64, opt ViewportOption,
) {
	if viewportTransform := viewportTransform(worldX1, worldY1, worldX2, worldY2, screenX1, screenY1, screenX2, screenY2, opt); viewportTransform != nil {
		agg2d.Affine(viewportTransform)
	}
}

// GetViewportTransform calculates the viewport transformation matrix without applying it.
// Returns the transformation that would map the given world coordinates to screen coordinates.
func (agg2d *Agg2D) GetViewportTransform(worldX1, worldY1, worldX2, worldY2,
	screenX1, screenY1, screenX2, screenY2 float64, opt ViewportOption,
) *transform.TransAffine {
	return viewportTransform(worldX1, worldY1, worldX2, worldY2, screenX1, screenY1, screenX2, screenY2, opt)
}

// ResetTransform is a Go compatibility alias for ResetTransformations.
func (agg2d *Agg2D) ResetTransform() {
	agg2d.ResetTransformations()
}

// Parallelogram transformations

// Parallelogram applies a transformation that maps a unit square to an arbitrary parallelogram.
// x1, y1: first corner of parallelogram
// x2, y2: second corner of parallelogram
// x3, y3: third corner of parallelogram
// The fourth corner is calculated automatically to complete the parallelogram.
// This matches the C++ Agg2D::parallelogram method.
func (agg2d *Agg2D) Parallelogram(x1, y1, x2, y2, x3, y3 float64) {
	// Calculate the parallelogram transformation matrix
	// The unit square (0,0)-(1,0)-(1,1)-(0,1) maps to (x1,y1)-(x2,y2)-(x3,y3)-(x1+x3-x2, y1+y3-y2)

	// Matrix elements for mapping unit square to parallelogram:
	// [x2-x1  x3-x1  x1]
	// [y2-y1  y3-y1  y1]
	// [0      0      1 ]

	sx := x2 - x1  // Horizontal basis vector
	shx := x3 - x1 // Horizontal skew vector
	sy := y3 - y1  // Vertical skew vector
	shy := y2 - y1 // Vertical basis vector
	tx := x1       // Translation X
	ty := y1       // Translation Y

	parallelogramTransform := transform.NewTransAffineFromValues(sx, shy, shx, sy, tx, ty)
	agg2d.Affine(parallelogramTransform)
}

// ParallelogramFromRect applies a transformation that maps a rectangle to a parallelogram.
// rectX1, rectY1, rectX2, rectY2: source rectangle bounds
// x1, y1, x2, y2, x3, y3: target parallelogram corners
func (agg2d *Agg2D) ParallelogramFromRect(rectX1, rectY1, rectX2, rectY2,
	x1, y1, x2, y2, x3, y3 float64,
) {
	// First translate and scale from rectangle to unit square
	rectWidth := rectX2 - rectX1
	rectHeight := rectY2 - rectY1

	if rectWidth == 0 || rectHeight == 0 {
		return // Invalid rectangle
	}

	// Transform to map rectangle to unit square
	agg2d.Translate(-rectX1, -rectY1)
	agg2d.Scale(1.0/rectWidth, 1.0/rectHeight)

	// Then apply parallelogram transformation
	agg2d.Parallelogram(x1, y1, x2, y2, x3, y3)
}

// GetParallelogramTransform calculates the parallelogram transformation matrix without applying it.
func (agg2d *Agg2D) GetParallelogramTransform(x1, y1, x2, y2, x3, y3 float64) *transform.TransAffine {
	sx := x2 - x1
	shx := x3 - x1
	sy := y3 - y1
	shy := y2 - y1
	tx := x1
	ty := y1

	return transform.NewTransAffineFromValues(sx, shy, shx, sy, tx, ty)
}

// Enhanced world-to-screen conversion methods

// WorldToScreen transforms a point from world coordinates to screen coordinates.
// This applies the current transformation matrix to the point.
// WorldToScreen method is already defined in agg2d.go

// ScreenToWorld method is already defined in agg2d.go

// WorldToScreenDistance transforms a distance from world coordinates to screen coordinates.
// This accounts for scaling but ignores translation.
func (agg2d *Agg2D) WorldToScreenDistance(worldDistance float64) float64 {
	return worldDistance * agg2d.GetScaling()
}

// ScreenToWorldDistance transforms a distance from screen coordinates to world coordinates.
// Returns false if the transformation is not invertible.
func (agg2d *Agg2D) ScreenToWorldDistance(screenDistance float64) (worldDistance float64, ok bool) {
	scaling := agg2d.GetScaling()
	if scaling == 0 {
		return 0, false
	}
	return screenDistance / scaling, true
}

// WorldToScreenRect transforms a rectangle from world coordinates to screen coordinates.
// Note: If the transformation includes rotation or skew, the result will be the bounding box.
func (agg2d *Agg2D) WorldToScreenRect(worldX1, worldY1, worldX2, worldY2 float64) (screenX1, screenY1, screenX2, screenY2 float64) {
	// Transform all four corners
	corners := [4][2]float64{
		{worldX1, worldY1},
		{worldX2, worldY1},
		{worldX2, worldY2},
		{worldX1, worldY2},
	}

	// Initialize with first corner
	screenX1, screenY1 = corners[0][0], corners[0][1]
	agg2d.WorldToScreen(&screenX1, &screenY1)
	screenX2, screenY2 = screenX1, screenY1

	// Find bounding box of all corners
	for i := 1; i < 4; i++ {
		x, y := corners[i][0], corners[i][1]
		agg2d.WorldToScreen(&x, &y)
		if x < screenX1 {
			screenX1 = x
		}
		if x > screenX2 {
			screenX2 = x
		}
		if y < screenY1 {
			screenY1 = y
		}
		if y > screenY2 {
			screenY2 = y
		}
	}

	return screenX1, screenY1, screenX2, screenY2
}

// ScreenToWorldRect transforms a rectangle from screen coordinates to world coordinates.
// Returns false if the transformation is not invertible.
func (agg2d *Agg2D) ScreenToWorldRect(screenX1, screenY1, screenX2, screenY2 float64) (worldX1, worldY1, worldX2, worldY2 float64, ok bool) {
	// Transform all four corners using existing ScreenToWorld method
	corners := [4][2]float64{
		{screenX1, screenY1},
		{screenX2, screenY1},
		{screenX2, screenY2},
		{screenX1, screenY2},
	}

	// Initialize with first corner
	worldX1, worldY1 = corners[0][0], corners[0][1]
	agg2d.ScreenToWorld(&worldX1, &worldY1)
	worldX2, worldY2 = worldX1, worldY1

	// Find bounding box of all corners
	for i := 1; i < 4; i++ {
		x, y := corners[i][0], corners[i][1]
		agg2d.ScreenToWorld(&x, &y)
		if x < worldX1 {
			worldX1 = x
		}
		if x > worldX2 {
			worldX2 = x
		}
		if y < worldY1 {
			worldY1 = y
		}
		if y > worldY2 {
			worldY2 = y
		}
	}

	return worldX1, worldY1, worldX2, worldY2, true
}

// TransformVector transforms a vector (direction) from world to screen coordinates.
// Unlike point transformation, this ignores translation.
func (agg2d *Agg2D) TransformVector(vectorX, vectorY float64) (screenVectorX, screenVectorY float64) {
	// Transform the vector by applying only the linear part of the transformation
	screenVectorX = agg2d.transform.SX*vectorX + agg2d.transform.SHX*vectorY
	screenVectorY = agg2d.transform.SHY*vectorX + agg2d.transform.SY*vectorY
	return screenVectorX, screenVectorY
}

// InverseTransformVector transforms a vector from screen to world coordinates.
// Returns false if the transformation is not invertible.
func (agg2d *Agg2D) InverseTransformVector(screenVectorX, screenVectorY float64) (vectorX, vectorY float64, ok bool) {
	det := agg2d.Determinant()
	if det == 0 {
		return 0, 0, false
	}

	// Calculate inverse of 2x2 linear part
	sx := agg2d.transform.SX
	shy := agg2d.transform.SHY
	shx := agg2d.transform.SHX
	sy := agg2d.transform.SY

	invDet := 1.0 / det
	invSX := sy * invDet
	invSHY := -shy * invDet
	invSHX := -shx * invDet
	invSY := sx * invDet

	vectorX = invSX*screenVectorX + invSHX*screenVectorY
	vectorY = invSHY*screenVectorX + invSY*screenVectorY
	return vectorX, vectorY, true
}

// GetTransformationBounds calculates the bounds of a transformed rectangle.
// This is useful for determining clipping regions after transformation.
func (agg2d *Agg2D) GetTransformationBounds(x1, y1, x2, y2 float64) (minX, minY, maxX, maxY float64) {
	return agg2d.WorldToScreenRect(x1, y1, x2, y2)
}

// IsAxisAligned returns true if the current transformation preserves axis alignment.
// This means the transformation only includes scaling and translation (no rotation or skew).
func (agg2d *Agg2D) IsAxisAligned() bool {
	return agg2d.transform.SHX == 0.0 && agg2d.transform.SHY == 0.0
}

// HasUniformScaling returns true if the transformation has uniform scaling in both axes.
func (agg2d *Agg2D) HasUniformScaling() bool {
	return math.Abs(agg2d.transform.SX-agg2d.transform.SY) < 1e-10 &&
		agg2d.transform.SHX == 0.0 && agg2d.transform.SHY == 0.0
}
