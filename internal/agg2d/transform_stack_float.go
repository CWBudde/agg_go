// Package agg2d float transform-stack + affine-matrix methods (L5/breadth).
// Color-agnostic delegations mirrored onto Agg2DFloat; the bodies match
// transform.go because they only read/write the shared world transform and the
// push/pop stack. The Affine(*transform.TransAffine) method already lives in
// transform_float.go.
package agg2d

import "github.com/cwbudde/agg_go/internal/transform"

// GetTransformations returns the current world transform as a Transformations
// matrix. Mirrors transform.go.
func (a *Agg2DFloat) GetTransformations() *Transformations {
	return &Transformations{
		AffineMatrix: [6]float64{
			a.transform.SX, a.transform.SHY, a.transform.SHX,
			a.transform.SY, a.transform.TX, a.transform.TY,
		},
	}
}

// SetTransformations replaces the world transform from a Transformations matrix.
// Mirrors transform.go.
func (a *Agg2DFloat) SetTransformations(tr *Transformations) {
	a.transform.SX = tr.AffineMatrix[0]
	a.transform.SHY = tr.AffineMatrix[1]
	a.transform.SHX = tr.AffineMatrix[2]
	a.transform.SY = tr.AffineMatrix[3]
	a.transform.TX = tr.AffineMatrix[4]
	a.transform.TY = tr.AffineMatrix[5]
	a.updateApproximationScales()
}

// AffineFromMatrix post-multiplies the world transform by a Transformations
// matrix. Mirrors transform.go.
func (a *Agg2DFloat) AffineFromMatrix(tr *Transformations) {
	affine := transform.NewTransAffineFromValues(
		tr.AffineMatrix[0], tr.AffineMatrix[1], tr.AffineMatrix[2],
		tr.AffineMatrix[3], tr.AffineMatrix[4], tr.AffineMatrix[5],
	)
	a.Affine(affine)
}

// PushTransformations is an alias for PushTransform to match the C++ API.
func (a *Agg2DFloat) PushTransformations() { a.PushTransform() }

// PopTransformations is an alias for PopTransform to match the C++ API.
func (a *Agg2DFloat) PopTransformations() { a.PopTransform() }

// PushTransform saves a copy of the current world transform. Mirrors
// transform.go.
func (a *Agg2DFloat) PushTransform() {
	if a.transformStack == nil {
		a.transformStack = &TransformStack{stack: make([]*transform.TransAffine, 0)}
	}
	transformCopy := transform.NewTransAffineFromValues(
		a.transform.SX, a.transform.SHY, a.transform.SHX,
		a.transform.SY, a.transform.TX, a.transform.TY,
	)
	a.transformStack.stack = append(a.transformStack.stack, transformCopy)
}

// PopTransform restores the most recently saved world transform; returns false
// if the stack is empty. Mirrors transform.go.
func (a *Agg2DFloat) PopTransform() bool {
	if a.transformStack == nil || len(a.transformStack.stack) == 0 {
		return false
	}
	stack := a.transformStack.stack
	lastIndex := len(stack) - 1
	saved := stack[lastIndex]
	a.transformStack.stack = stack[:lastIndex]

	a.transform.SX = saved.SX
	a.transform.SHY = saved.SHY
	a.transform.SHX = saved.SHX
	a.transform.SY = saved.SY
	a.transform.TX = saved.TX
	a.transform.TY = saved.TY

	a.updateApproximationScales()
	return true
}

// GetTransformStackDepth returns the number of saved transforms. Mirrors
// transform.go.
func (a *Agg2DFloat) GetTransformStackDepth() int {
	if a.transformStack == nil {
		return 0
	}
	return len(a.transformStack.stack)
}
