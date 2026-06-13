package agg2d

import (
	"math"
	"testing"

	"github.com/cwbudde/agg_go/internal/transform"
)

// transformStackScene is the method subset needed to exercise the transform
// stack + affine-matrix accessors across both pipelines. All of these manipulate
// only the shared world transform (a *transform.TransAffine) and are color-
// agnostic, so the float twin must produce bit-identical matrices to the 8-bit
// oracle after the same sequence of operations.
type transformStackScene interface {
	ResetTransformations()
	Translate(x, y float64)
	Scale(sx, sy float64)
	Rotate(angle float64)
	Affine(tr *transform.TransAffine)
	AffineFromMatrix(tr *Transformations)
	GetTransformations() *Transformations
	SetTransformations(tr *Transformations)
	PushTransform()
	PopTransform() bool
	PushTransformations()
	PopTransformations()
	GetTransformStackDepth() int
}

var (
	_ transformStackScene = (*Agg2D)(nil)
	_ transformStackScene = (*Agg2DFloat)(nil)
)

func matricesEqual(a, b [6]float64) bool {
	for i := range a {
		if math.Abs(a[i]-b[i]) > 1e-9 {
			return false
		}
	}
	return true
}

// runTransformStackSeq performs a fixed sequence of transform-stack operations
// and returns the resulting affine matrix plus the final stack depth.
func runTransformStackSeq(s transformStackScene) ([6]float64, int) {
	s.ResetTransformations()
	s.Translate(10, 20)
	s.PushTransform()       // save T(10,20)
	s.Scale(2, 3)           // now T(10,20) * S(2,3)
	s.Rotate(0.5)           // ... * R(0.5)
	s.PushTransformations() // save the scaled+rotated state (alias)
	s.Affine(transform.NewTransAffineFromValues(1, 0.2, 0.1, 1, 5, 5))
	s.PopTransformations() // restore scaled+rotated state
	s.AffineFromMatrix(&Transformations{AffineMatrix: [6]float64{1, 0, 0, 1, 3, 4}})
	s.PopTransform() // restore T(10,20)
	m := s.GetTransformations()
	return m.AffineMatrix, s.GetTransformStackDepth()
}

func TestParityTransformStackSequence(t *testing.T) {
	a8 := NewAgg2D()
	af := NewAgg2DFloat()

	m8, depth8 := runTransformStackSeq(a8)
	mF, depthF := runTransformStackSeq(af)

	if !matricesEqual(m8, mF) {
		t.Errorf("transform-stack sequence matrix: float=%v 8bit=%v", mF, m8)
	}
	if depth8 != depthF {
		t.Errorf("transform-stack depth: float=%d 8bit=%d", depthF, depth8)
	}
	if depthF != 0 {
		t.Errorf("expected balanced push/pop to leave depth 0, got %d", depthF)
	}
}

func TestParityTransformStackGetSetRoundTrip(t *testing.T) {
	af := NewAgg2DFloat()
	af.ResetTransformations()
	af.Translate(7, 8)
	af.Scale(1.5, 2.5)
	saved := af.GetTransformations()

	// Mutate, then restore via SetTransformations.
	af.Rotate(1.0)
	af.SetTransformations(saved)
	got := af.GetTransformations()
	if !matricesEqual(saved.AffineMatrix, got.AffineMatrix) {
		t.Errorf("Get/SetTransformations round-trip: got %v want %v", got.AffineMatrix, saved.AffineMatrix)
	}
}

func TestParityTransformPopEmptyReturnsFalse(t *testing.T) {
	af := NewAgg2DFloat()
	if af.PopTransform() {
		t.Error("PopTransform on empty stack should return false")
	}
	af.PushTransform()
	if !af.PopTransform() {
		t.Error("PopTransform after PushTransform should return true")
	}
	if af.PopTransform() {
		t.Error("PopTransform after emptying stack should return false")
	}
}
