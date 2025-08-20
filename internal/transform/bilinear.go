package transform

// Compile-time interface checks
var _ Transformer = (*TransBilinear)(nil)
var _ InverseTransformer = (*TransBilinear)(nil)

// TransBilinear represents a bilinear transformation for quadrilateral mapping
type TransBilinear struct {
	mtx   [4][2]float64 // Transformation matrix coefficients
	valid bool          // Whether the transformation is valid
}

// NewTransBilinear creates a new invalid bilinear transformation
func NewTransBilinear() *TransBilinear {
	return &TransBilinear{valid: false}
}

// NewTransBilinearQuadToQuad creates a bilinear transformation from arbitrary quadrilateral to quadrilateral
func NewTransBilinearQuadToQuad(src, dst [8]float64) *TransBilinear {
	tb := NewTransBilinear()
	tb.QuadToQuad(src, dst)
	return tb
}

// NewTransBilinearRectToQuad creates a bilinear transformation from rectangle to quadrilateral
func NewTransBilinearRectToQuad(x1, y1, x2, y2 float64, quad [8]float64) *TransBilinear {
	tb := NewTransBilinear()
	tb.RectToQuad(x1, y1, x2, y2, quad)
	return tb
}

// NewTransBilinearQuadToRect creates a bilinear transformation from quadrilateral to rectangle
func NewTransBilinearQuadToRect(quad [8]float64, x1, y1, x2, y2 float64) *TransBilinear {
	tb := NewTransBilinear()
	tb.QuadToRect(quad, x1, y1, x2, y2)
	return tb
}

// QuadToQuad sets the transformation using two arbitrary quadrilaterals
// Each quadrilateral is represented as [x0,y0, x1,y1, x2,y2, x3,y3]
func (tb *TransBilinear) QuadToQuad(src, dst [8]float64) {
	var left [4][4]float64
	var right [4][2]float64

	for i := 0; i < 4; i++ {
		ix := i * 2
		iy := ix + 1
		left[i][0] = 1.0
		left[i][1] = src[ix] * src[iy]
		left[i][2] = src[ix]
		left[i][3] = src[iy]

		right[i][0] = dst[ix]
		right[i][1] = dst[iy]
	}

	tb.valid = solve4x2(&left, right, &tb.mtx)
}

// RectToQuad sets the direct transformation from rectangle to quadrilateral
// Matches C++ AGG implementation: src[0,6]=x1, src[2,4]=x2, src[1,3]=y1, src[5,7]=y2
func (tb *TransBilinear) RectToQuad(x1, y1, x2, y2 float64, quad [8]float64) {
	src := [8]float64{
		x1, y1, // 0,1: bottom-left
		x2, y1, // 2,3: bottom-right
		x2, y2, // 4,5: top-right
		x1, y2, // 6,7: top-left
	}
	tb.QuadToQuad(src, quad)
}

// QuadToRect sets the reverse transformation from quadrilateral to rectangle
// Matches C++ AGG implementation: dst[0,6]=x1, dst[2,4]=x2, dst[1,3]=y1, dst[5,7]=y2
func (tb *TransBilinear) QuadToRect(quad [8]float64, x1, y1, x2, y2 float64) {
	dst := [8]float64{
		x1, y1, // 0,1: bottom-left
		x2, y1, // 2,3: bottom-right
		x2, y2, // 4,5: top-right
		x1, y2, // 6,7: top-left
	}
	tb.QuadToQuad(quad, dst)
}

// IsValid returns true if the transformation matrix was computed successfully
func (tb *TransBilinear) IsValid() bool {
	return tb.valid
}

// Transform applies the bilinear transformation to a point (implements Transformer interface)
func (tb *TransBilinear) Transform(x, y *float64) {
	if !tb.valid {
		return
	}

	tx := *x
	ty := *y
	xy := tx * ty
	*x = tb.mtx[0][0] + tb.mtx[1][0]*xy + tb.mtx[2][0]*tx + tb.mtx[3][0]*ty
	*y = tb.mtx[0][1] + tb.mtx[1][1]*xy + tb.mtx[2][1]*tx + tb.mtx[3][1]*ty
}

// TransformValues applies the bilinear transformation to a point and returns new coordinates
// Extension: This method is not present in the original C++ AGG implementation.
// It provides a convenient value-returning variant of the pointer-based Transform method.
func (tb *TransBilinear) TransformValues(x, y float64) (float64, float64) {
	if !tb.valid {
		return x, y
	}

	xy := x * y
	newX := tb.mtx[0][0] + tb.mtx[1][0]*xy + tb.mtx[2][0]*x + tb.mtx[3][0]*y
	newY := tb.mtx[0][1] + tb.mtx[1][1]*xy + tb.mtx[2][1]*x + tb.mtx[3][1]*y
	return newX, newY
}

// InverseTransform applies the inverse bilinear transformation (implements extended Transformer interface)
// Extension: This method is not present in the original C++ AGG implementation.
// It provides inverse transformation capabilities using Newton-Raphson iteration.
func (tb *TransBilinear) InverseTransform(x, y *float64) {
	if !tb.valid {
		return
	}

	dx, dy := *x, *y
	sx, sy := tb.InverseTransformValues(dx, dy)
	*x = sx
	*y = sy
}

// InverseTransformValues finds the source coordinates that would transform to the given destination coordinates
// Extension: This method is not present in the original C++ AGG implementation.
// Uses Newton-Raphson iteration to solve the inverse bilinear transformation.
func (tb *TransBilinear) InverseTransformValues(dx, dy float64) (float64, float64) {
	if !tb.valid {
		return dx, dy
	}

	// Newton-Raphson parameters
	const maxIterations = 10
	const epsilon = 1e-10

	// Initial guess - use the destination point as starting point
	x, y := dx, dy

	for i := 0; i < maxIterations; i++ {
		// Current transformation result
		fx, fy := tb.TransformValues(x, y)

		// Calculate residual (how far we are from target)
		residualX := fx - dx
		residualY := fy - dy

		// Check convergence
		if residualX*residualX+residualY*residualY < epsilon*epsilon {
			return x, y
		}

		// Calculate Jacobian matrix elements
		// F(x,y) = [mtx[0][0] + mtx[1][0]*x*y + mtx[2][0]*x + mtx[3][0]*y,
		//           mtx[0][1] + mtx[1][1]*x*y + mtx[2][1]*x + mtx[3][1]*y]
		//
		// ∂F/∂x = [mtx[1][0]*y + mtx[2][0], mtx[1][1]*y + mtx[2][1]]
		// ∂F/∂y = [mtx[1][0]*x + mtx[3][0], mtx[1][1]*x + mtx[3][1]]

		j11 := tb.mtx[1][0]*y + tb.mtx[2][0] // ∂fx/∂x
		j12 := tb.mtx[1][0]*x + tb.mtx[3][0] // ∂fx/∂y
		j21 := tb.mtx[1][1]*y + tb.mtx[2][1] // ∂fy/∂x
		j22 := tb.mtx[1][1]*x + tb.mtx[3][1] // ∂fy/∂y

		// Calculate determinant of Jacobian
		det := j11*j22 - j12*j21
		if det*det < epsilon*epsilon {
			// Jacobian is singular, cannot continue
			return dx, dy
		}

		// Solve J * delta = -residual using Cramer's rule
		deltaX := (-residualX*j22 + residualY*j12) / det
		deltaY := (residualX*j21 - residualY*j11) / det

		// Update estimate
		x += deltaX
		y += deltaY
	}

	// If we didn't converge, return the best estimate we have
	return x, y
}

// IteratorX provides efficient iteration along horizontal lines
// Note: The C++ AGG version has public x,y fields and operator++.
// This Go version uses methods for better encapsulation, which is idiomatic Go.
type IteratorX struct {
	x, y float64 // Current transformed coordinates
	incX float64 // X increment per step
	incY float64 // Y increment per step
}

// NewIteratorX creates a new iterator starting at (tx, ty) with given step size
// This corresponds to the begin(x, y, step) method in C++ AGG.
func (tb *TransBilinear) NewIteratorX(tx, ty, step float64) *IteratorX {
	if !tb.valid {
		return &IteratorX{x: tx, y: ty}
	}

	return &IteratorX{
		incX: tb.mtx[1][0]*step*ty + tb.mtx[2][0]*step,
		incY: tb.mtx[1][1]*step*ty + tb.mtx[2][1]*step,
		x:    tb.mtx[0][0] + tb.mtx[1][0]*tx*ty + tb.mtx[2][0]*tx + tb.mtx[3][0]*ty,
		y:    tb.mtx[0][1] + tb.mtx[1][1]*tx*ty + tb.mtx[2][1]*tx + tb.mtx[3][1]*ty,
	}
}

// X returns the current x coordinate
func (it *IteratorX) X() float64 {
	return it.x
}

// Y returns the current y coordinate
func (it *IteratorX) Y() float64 {
	return it.y
}

// Next advances the iterator to the next position
func (it *IteratorX) Next() {
	it.x += it.incX
	it.y += it.incY
}

// ToMatrix returns the transformation matrix as a 4x2 array
// Extension: This method is not present in the original C++ AGG implementation.
// This is useful for debugging and interfacing with span interpolators.
func (tb *TransBilinear) ToMatrix() [4][2]float64 {
	return tb.mtx
}

// NewTransBilinearFromTransformer attempts to create a TransBilinear from another transformer
// by sampling it at 4 corner points and fitting a bilinear transformation.
// Extension: This function is not present in the original C++ AGG implementation.
// This is useful for integrating with span interpolators.
func NewTransBilinearFromTransformer(transformer Transformer, srcQuad [8]float64) *TransBilinear {
	// Transform the source quadrilateral using the provided transformer
	var dstQuad [8]float64
	for i := 0; i < 4; i++ {
		x, y := srcQuad[i*2], srcQuad[i*2+1]
		transformer.Transform(&x, &y)
		dstQuad[i*2] = x
		dstQuad[i*2+1] = y
	}

	return NewTransBilinearQuadToQuad(srcQuad, dstQuad)
}

// NewTransBilinearFromRect creates a bilinear transformation by sampling a transformer over a rectangle
// Extension: This function is not present in the original C++ AGG implementation.
func NewTransBilinearFromRect(transformer Transformer, x1, y1, x2, y2 float64) *TransBilinear {
	// Create a rectangle as source quadrilateral
	srcQuad := [8]float64{
		x1, y1, // bottom-left
		x2, y1, // bottom-right
		x2, y2, // top-right
		x1, y2, // top-left
	}

	return NewTransBilinearFromTransformer(transformer, srcQuad)
}
