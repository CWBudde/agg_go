package transform

import (
	"math"
	"testing"
)

const warpMagnifierTestEpsilon = 1e-10

func TestNewTransWarpMagnifier(t *testing.T) {
	m := NewTransWarpMagnifier()

	// Test default values
	if m.CenterX() != 0.0 || m.CenterY() != 0.0 {
		t.Errorf("Default center should be (0,0), got (%g,%g)", m.CenterX(), m.CenterY())
	}

	if m.Magnification() != 1.0 {
		t.Errorf("Default magnification should be 1.0, got %g", m.Magnification())
	}

	if m.Radius() != 1.0 {
		t.Errorf("Default radius should be 1.0, got %g", m.Radius())
	}

	// Test Center() method
	x, y := m.Center()
	if x != 0.0 || y != 0.0 {
		t.Errorf("Center() should return (0,0), got (%g,%g)", x, y)
	}
}

func TestNewTransWarpMagnifierWithParams(t *testing.T) {
	xc, yc, magn, radius := 10.0, 20.0, 2.5, 15.0
	m := NewTransWarpMagnifierWithParams(xc, yc, magn, radius)

	if m.CenterX() != xc || m.CenterY() != yc {
		t.Errorf("Center should be (%g,%g), got (%g,%g)", xc, yc, m.CenterX(), m.CenterY())
	}

	if m.Magnification() != magn {
		t.Errorf("Magnification should be %g, got %g", magn, m.Magnification())
	}

	if m.Radius() != radius {
		t.Errorf("Radius should be %g, got %g", radius, m.Radius())
	}
}

func TestSettersAndGetters(t *testing.T) {
	m := NewTransWarpMagnifier()

	// Test SetCenter
	m.SetCenter(5.0, 10.0)
	if m.CenterX() != 5.0 || m.CenterY() != 10.0 {
		t.Errorf("SetCenter failed, expected (5,10), got (%g,%g)", m.CenterX(), m.CenterY())
	}

	x, y := m.Center()
	if x != 5.0 || y != 10.0 {
		t.Errorf("Center() failed after SetCenter, expected (5,10), got (%g,%g)", x, y)
	}

	// Test SetMagnification
	m.SetMagnification(3.0)
	if m.Magnification() != 3.0 {
		t.Errorf("SetMagnification failed, expected 3.0, got %g", m.Magnification())
	}

	// Test SetRadius
	m.SetRadius(25.0)
	if m.Radius() != 25.0 {
		t.Errorf("SetRadius failed, expected 25.0, got %g", m.Radius())
	}
}

func TestTransformInsideRadius(t *testing.T) {
	m := NewTransWarpMagnifier()
	m.SetCenter(100.0, 100.0)
	m.SetMagnification(2.0)
	m.SetRadius(50.0)

	// Test point at center (should be unchanged)
	x, y := 100.0, 100.0
	m.Transform(&x, &y)
	if math.Abs(x-100.0) > warpMagnifierTestEpsilon || math.Abs(y-100.0) > warpMagnifierTestEpsilon {
		t.Errorf("Point at center should be unchanged, got (%g,%g)", x, y)
	}

	// Test point inside radius (should be magnified)
	x, y = 120.0, 130.0 // distance = sqrt(20^2 + 30^2) = sqrt(1300) ≈ 36.06 < 50
	originalX, originalY := x, y
	m.Transform(&x, &y)

	// Expected: center + (original - center) * magnification
	expectedX := 100.0 + (originalX-100.0)*2.0 // 100 + 20*2 = 140
	expectedY := 100.0 + (originalY-100.0)*2.0 // 100 + 30*2 = 160

	if math.Abs(x-expectedX) > warpMagnifierTestEpsilon || math.Abs(y-expectedY) > warpMagnifierTestEpsilon {
		t.Errorf("Inside radius transform failed, expected (%g,%g), got (%g,%g)", expectedX, expectedY, x, y)
	}
}

func TestTransformOutsideRadius(t *testing.T) {
	m := NewTransWarpMagnifier()
	m.SetCenter(0.0, 0.0)
	m.SetMagnification(2.0)
	m.SetRadius(10.0)

	// Test point outside radius
	x, y := 20.0, 0.0 // distance = 20 > 10
	m.Transform(&x, &y)

	// Expected calculation: r = 20, mult = (20 + 10*(2-1)) / 20 = 30/20 = 1.5
	// So x should be 0 + 20 * 1.5 = 30
	expectedX := 30.0
	expectedY := 0.0

	if math.Abs(x-expectedX) > warpMagnifierTestEpsilon || math.Abs(y-expectedY) > warpMagnifierTestEpsilon {
		t.Errorf("Outside radius transform failed, expected (%g,%g), got (%g,%g)", expectedX, expectedY, x, y)
	}
}

func TestTransformAtBoundary(t *testing.T) {
	m := NewTransWarpMagnifier()
	m.SetCenter(0.0, 0.0)
	m.SetMagnification(2.0)
	m.SetRadius(10.0)

	// Test point exactly at radius boundary
	x, y := 10.0, 0.0 // distance = exactly 10
	m.Transform(&x, &y)

	// At the boundary, the transformation should be continuous
	// Using outside formula: mult = (10 + 10*(2-1)) / 10 = 20/10 = 2.0
	expectedX := 20.0
	expectedY := 0.0

	if math.Abs(x-expectedX) > warpMagnifierTestEpsilon || math.Abs(y-expectedY) > warpMagnifierTestEpsilon {
		t.Errorf("Boundary transform failed, expected (%g,%g), got (%g,%g)", expectedX, expectedY, x, y)
	}
}

func TestWarpMagnifierInverseTransform(t *testing.T) {
	m := NewTransWarpMagnifier()
	m.SetCenter(50.0, 50.0)
	m.SetMagnification(3.0)
	m.SetRadius(20.0)

	testCases := []struct {
		name string
		x, y float64
	}{
		{"Center point", 50.0, 50.0},
		{"Inside radius", 60.0, 55.0},
		{"Outside radius", 100.0, 80.0},
		{"At boundary", 70.0, 50.0}, // distance = 20 (exactly at radius)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			originalX, originalY := tc.x, tc.y

			// Apply transform then inverse transform
			x, y := originalX, originalY
			m.Transform(&x, &y)
			m.InverseTransform(&x, &y)

			if math.Abs(x-originalX) > warpMagnifierTestEpsilon || math.Abs(y-originalY) > warpMagnifierTestEpsilon {
				t.Errorf("Round-trip failed for %s: original (%g,%g), final (%g,%g)", tc.name, originalX, originalY, x, y)
			}
		})
	}
}

func TestInverseTransformSpecificCases(t *testing.T) {
	m := NewTransWarpMagnifier()
	m.SetCenter(0.0, 0.0)
	m.SetMagnification(2.0)
	m.SetRadius(10.0)

	// Test inverse transform for point inside magnified radius
	x, y := 15.0, 0.0 // This is < radius * magnification (10 * 2 = 20)
	m.InverseTransform(&x, &y)

	expectedX := 15.0 / 2.0 // 7.5
	expectedY := 0.0

	if math.Abs(x-expectedX) > warpMagnifierTestEpsilon || math.Abs(y-expectedY) > warpMagnifierTestEpsilon {
		t.Errorf("Inverse transform inside magnified radius failed, expected (%g,%g), got (%g,%g)", expectedX, expectedY, x, y)
	}

	// Test inverse transform for point outside magnified radius
	x, y = 30.0, 0.0 // This is > radius * magnification (20)
	m.InverseTransform(&x, &y)

	// rnew = 30 - 10*(2-1) = 30 - 10 = 20
	// x = 0 + 20 * 30 / 30 = 20
	expectedX = 20.0
	expectedY = 0.0

	if math.Abs(x-expectedX) > warpMagnifierTestEpsilon || math.Abs(y-expectedY) > warpMagnifierTestEpsilon {
		t.Errorf("Inverse transform outside magnified radius failed, expected (%g,%g), got (%g,%g)", expectedX, expectedY, x, y)
	}
}

func TestEdgeCases(t *testing.T) {
	// Test with magnification = 1.0 (no magnification)
	t.Run("No magnification", func(t *testing.T) {
		m := NewTransWarpMagnifier()
		m.SetMagnification(1.0)

		x, y := 5.0, 10.0
		originalX, originalY := x, y
		m.Transform(&x, &y)

		if math.Abs(x-originalX) > warpMagnifierTestEpsilon || math.Abs(y-originalY) > warpMagnifierTestEpsilon {
			t.Errorf("No magnification should leave coordinates unchanged, got (%g,%g)", x, y)
		}
	})

	// Test with zero radius
	t.Run("Zero radius", func(t *testing.T) {
		m := NewTransWarpMagnifier()
		m.SetRadius(0.0)
		m.SetMagnification(2.0)

		x, y := 1.0, 1.0
		m.Transform(&x, &y)

		// With zero radius, all points are outside, so formula applies
		// r = sqrt(2), mult = (sqrt(2) + 0*(2-1)) / sqrt(2) = 1.0
		// So coordinates should be unchanged
		if math.Abs(x-1.0) > warpMagnifierTestEpsilon || math.Abs(y-1.0) > warpMagnifierTestEpsilon {
			t.Errorf("Zero radius should not change coordinates, got (%g,%g)", x, y)
		}
	})

	// Test with shrinking magnification (< 1.0)
	t.Run("Shrinking magnification", func(t *testing.T) {
		m := NewTransWarpMagnifier()
		m.SetCenter(0.0, 0.0)
		m.SetMagnification(0.5)
		m.SetRadius(10.0)

		x, y := 5.0, 0.0 // Inside radius
		m.Transform(&x, &y)

		expectedX := 5.0 * 0.5 // 2.5
		if math.Abs(x-expectedX) > warpMagnifierTestEpsilon {
			t.Errorf("Shrinking magnification failed, expected %g, got %g", expectedX, x)
		}
	})
}

func TestContinuityAtBoundary(t *testing.T) {
	m := NewTransWarpMagnifier()
	m.SetCenter(0.0, 0.0)
	m.SetMagnification(3.0)
	m.SetRadius(5.0)

	// Test points very close to the boundary from both sides
	epsilon := 1e-8

	// Point just inside radius
	x1, y1 := 5.0-epsilon, 0.0
	m.Transform(&x1, &y1)

	// Point just outside radius
	x2, y2 := 5.0+epsilon, 0.0
	m.Transform(&x2, &y2)

	// The transformation should be continuous at the boundary
	if math.Abs(x1-x2) > 1e-6 {
		t.Errorf("Transformation not continuous at boundary: inside=%g, outside=%g, diff=%g", x1, x2, math.Abs(x1-x2))
	}
}

func BenchmarkWarpMagnifierTransform(b *testing.B) {
	m := NewTransWarpMagnifier()
	m.SetCenter(100.0, 100.0)
	m.SetMagnification(2.0)
	m.SetRadius(50.0)

	x, y := 120.0, 130.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx, ty := x, y
		m.Transform(&tx, &ty)
	}
}

func BenchmarkWarpMagnifierInverseTransform(b *testing.B) {
	m := NewTransWarpMagnifier()
	m.SetCenter(100.0, 100.0)
	m.SetMagnification(2.0)
	m.SetRadius(50.0)

	x, y := 140.0, 160.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx, ty := x, y
		m.InverseTransform(&tx, &ty)
	}
}
