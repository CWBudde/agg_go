package transform

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
)

// boundedFloat generates float64 values in [-1e6, 1e6] — a range that avoids
// float64 overflow when values are combined in matrix arithmetic.
type boundedFloat struct{ V float64 }

func (boundedFloat) Generate(r *rand.Rand, _ int) reflect.Value {
	v := (r.Float64()*2 - 1) * 1e6
	return reflect.ValueOf(boundedFloat{v})
}

// Property: translating a point then applying the inverse transform returns the original point.
func TestPropertyTranslateInverseRoundTrip(t *testing.T) {
	err := quick.Check(func(tx, ty, px, py boundedFloat) bool {
		m := NewTransAffine().Translate(tx.V, ty.V)
		x, y := px.V, py.V
		m.Transform(&x, &y)
		m.InverseTransform(&x, &y)
		return math.Abs(x-px.V) < 1e-9 && math.Abs(y-py.V) < 1e-9
	}, nil)
	if err != nil {
		t.Error(err)
	}
}

// Property: rotating a point then applying the inverse transform returns the original point.
// Tolerance scales with point magnitude to allow for accumulated floating-point error.
func TestPropertyRotateInverseRoundTrip(t *testing.T) {
	err := quick.Check(func(angle, px, py boundedFloat) bool {
		m := NewTransAffine().Rotate(angle.V)
		x, y := px.V, py.V
		m.Transform(&x, &y)
		m.InverseTransform(&x, &y)
		eps := 1e-9 * (1 + math.Hypot(px.V, py.V))
		return math.Abs(x-px.V) < eps && math.Abs(y-py.V) < eps
	}, nil)
	if err != nil {
		t.Error(err)
	}
}

// Property: multiplying any matrix by the identity matrix leaves it unchanged.
func TestPropertyMultiplyByIdentityIsNoOp(t *testing.T) {
	err := quick.Check(func(sx, shy, shx, sy, tx, ty boundedFloat) bool {
		m := NewTransAffineFromValues(sx.V, shy.V, shx.V, sy.V, tx.V, ty.V)
		id := NewTransAffine()
		result := m.Copy().Multiply(id)
		return result.IsEqual(m, 1e-10)
	}, nil)
	if err != nil {
		t.Error(err)
	}
}

// Property: translation composition is associative — (A*B)*C == A*(B*C).
func TestPropertyTranslationCompositionAssociativity(t *testing.T) {
	err := quick.Check(func(ax, ay, bx, by, cx, cy boundedFloat) bool {
		a := NewTransAffine().Translate(ax.V, ay.V)
		b := NewTransAffine().Translate(bx.V, by.V)
		c := NewTransAffine().Translate(cx.V, cy.V)

		// (a*b)*c
		abc1 := a.Copy().Multiply(b).Multiply(c)
		// a*(b*c)
		abc2 := a.Copy().Multiply(b.Copy().Multiply(c))
		return abc1.IsEqual(abc2, 1e-9)
	}, nil)
	if err != nil {
		t.Error(err)
	}
}

// Property: a scale+translate matrix composed with its inverse gives the identity.
func TestPropertyInverseGivesIdentity(t *testing.T) {
	err := quick.Check(func(scale, tx, ty boundedFloat) bool {
		if math.Abs(scale.V) < 1e-3 {
			return true // skip near-singular matrices
		}
		m := NewTransAffine().Scale(scale.V).Translate(tx.V, ty.V)
		inv := m.Copy().Invert()
		product := m.Copy().Multiply(inv)
		return product.IsIdentity(1e-8)
	}, nil)
	if err != nil {
		t.Error(err)
	}
}
