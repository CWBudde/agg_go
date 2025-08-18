package transform

import (
	"math"
	"testing"
)

const perspectiveTestEpsilon = 1e-12

func TestNewTransPerspective(t *testing.T) {
	p := NewTransPerspective()

	// Should be identity matrix
	if !p.IsIdentity(perspectiveTestEpsilon) {
		t.Error("NewTransPerspective should create identity matrix")
	}

	// Check specific values
	if p.SX != 1.0 || p.SY != 1.0 || p.W2 != 1.0 {
		t.Error("NewTransPerspective diagonal should be 1.0")
	}

	if p.SHY != 0.0 || p.W0 != 0.0 || p.SHX != 0.0 || p.W1 != 0.0 || p.TX != 0.0 || p.TY != 0.0 {
		t.Error("NewTransPerspective off-diagonal should be 0.0")
	}
}

func TestNewTransPerspectiveFromValues(t *testing.T) {
	sx, shy, w0 := 2.0, 0.5, 0.1
	shx, sy, w1 := 0.3, 1.5, 0.2
	tx, ty, w2 := 10.0, 20.0, 1.0

	p := NewTransPerspectiveFromValues(sx, shy, w0, shx, sy, w1, tx, ty, w2)

	if p.SX != sx || p.SHY != shy || p.W0 != w0 ||
		p.SHX != shx || p.SY != sy || p.W1 != w1 ||
		p.TX != tx || p.TY != ty || p.W2 != w2 {
		t.Error("NewTransPerspectiveFromValues should set all values correctly")
	}
}

func TestNewTransPerspectiveFromArray(t *testing.T) {
	m := [9]float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0}
	p := NewTransPerspectiveFromArray(m)

	if p.SX != 1.0 || p.SHY != 2.0 || p.W0 != 3.0 ||
		p.SHX != 4.0 || p.SY != 5.0 || p.W1 != 6.0 ||
		p.TX != 7.0 || p.TY != 8.0 || p.W2 != 9.0 {
		t.Error("NewTransPerspectiveFromArray should set values from array")
	}
}

func TestNewTransPerspectiveFromAffine(t *testing.T) {
	a := NewTransAffineFromValues(2.0, 0.5, 0.3, 1.5, 10.0, 20.0)
	p := NewTransPerspectiveFromAffine(a)

	if p.SX != a.SX || p.SHY != a.SHY || p.SHX != a.SHX ||
		p.SY != a.SY || p.TX != a.TX || p.TY != a.TY {
		t.Error("Should copy affine values correctly")
	}

	if p.W0 != 0.0 || p.W1 != 0.0 || p.W2 != 1.0 {
		t.Error("Should set perspective values for affine conversion")
	}
}

func TestPerspectiveReset(t *testing.T) {
	p := NewTransPerspectiveFromValues(1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0)
	p.Reset()

	if !p.IsIdentity(perspectiveTestEpsilon) {
		t.Error("Reset should create identity matrix")
	}
}

func TestSquareToQuad(t *testing.T) {
	// Test parallelogram case (affine)
	p := NewTransPerspective()
	quad := [8]float64{0.0, 0.0, 2.0, 0.0, 2.0, 1.0, 0.0, 1.0} // rectangle scaled by 2 in x

	if !p.SquareToQuad(quad) {
		t.Error("SquareToQuad should succeed for simple rectangle")
	}

	// Test transformation of unit square corners
	x, y := 0.0, 0.0
	p.Transform(&x, &y)
	if math.Abs(x-0.0) > perspectiveTestEpsilon || math.Abs(y-0.0) > perspectiveTestEpsilon {
		t.Errorf("(0,0) should map to (0,0), got (%f,%f)", x, y)
	}

	x, y = 1.0, 0.0
	p.Transform(&x, &y)
	if math.Abs(x-2.0) > perspectiveTestEpsilon || math.Abs(y-0.0) > perspectiveTestEpsilon {
		t.Errorf("(1,0) should map to (2,0), got (%f,%f)", x, y)
	}
}

func TestSquareToQuadSingular(t *testing.T) {
	p := NewTransPerspective()
	// Degenerate quadrilateral (all points on a line)
	quad := [8]float64{0.0, 0.0, 1.0, 0.0, 2.0, 0.0, 3.0, 0.0}

	if p.SquareToQuad(quad) {
		t.Error("SquareToQuad should fail for degenerate quadrilateral")
	}
}

func TestQuadToSquare(t *testing.T) {
	p := NewTransPerspective()
	quad := [8]float64{0.0, 0.0, 2.0, 0.0, 2.0, 1.0, 0.0, 1.0}

	if !p.QuadToSquare(quad) {
		t.Error("QuadToSquare should succeed")
	}

	// Test that quad corners map to unit square
	x, y := 0.0, 0.0
	p.Transform(&x, &y)
	if math.Abs(x-0.0) > perspectiveTestEpsilon || math.Abs(y-0.0) > perspectiveTestEpsilon {
		t.Errorf("(0,0) should map to (0,0), got (%f,%f)", x, y)
	}

	x, y = 2.0, 0.0
	p.Transform(&x, &y)
	if math.Abs(x-1.0) > perspectiveTestEpsilon || math.Abs(y-0.0) > perspectiveTestEpsilon {
		t.Errorf("(2,0) should map to (1,0), got (%f,%f)", x, y)
	}
}

func TestQuadToQuad(t *testing.T) {
	p := NewTransPerspective()
	src := [8]float64{0.0, 0.0, 1.0, 0.0, 1.0, 1.0, 0.0, 1.0} // unit square
	dst := [8]float64{0.0, 0.0, 2.0, 0.0, 2.0, 1.0, 0.0, 1.0} // scaled rectangle

	if !p.QuadToQuad(src, dst) {
		t.Error("QuadToQuad should succeed")
	}

	// Test that source corners map to destination corners
	x, y := 1.0, 0.0
	p.Transform(&x, &y)
	if math.Abs(x-2.0) > perspectiveTestEpsilon || math.Abs(y-0.0) > perspectiveTestEpsilon {
		t.Errorf("(1,0) should map to (2,0), got (%f,%f)", x, y)
	}
}

func TestRectToQuad(t *testing.T) {
	p := NewTransPerspective()
	quad := [8]float64{10.0, 10.0, 20.0, 10.0, 20.0, 20.0, 10.0, 20.0}

	if !p.RectToQuad(0.0, 0.0, 1.0, 1.0, quad) {
		t.Error("RectToQuad should succeed")
	}

	// Test corners
	x, y := 0.0, 0.0
	p.Transform(&x, &y)
	if math.Abs(x-10.0) > perspectiveTestEpsilon || math.Abs(y-10.0) > perspectiveTestEpsilon {
		t.Errorf("(0,0) should map to (10,10), got (%f,%f)", x, y)
	}
}

func TestQuadToRect(t *testing.T) {
	p := NewTransPerspective()
	quad := [8]float64{10.0, 10.0, 20.0, 10.0, 20.0, 20.0, 10.0, 20.0}

	if !p.QuadToRect(quad, 0.0, 0.0, 1.0, 1.0) {
		t.Error("QuadToRect should succeed")
	}

	// Test corners
	x, y := 10.0, 10.0
	p.Transform(&x, &y)
	if math.Abs(x-0.0) > perspectiveTestEpsilon || math.Abs(y-0.0) > perspectiveTestEpsilon {
		t.Errorf("(10,10) should map to (0,0), got (%f,%f)", x, y)
	}
}

func TestPerspectiveInvert(t *testing.T) {
	// Create a non-trivial transformation
	p := NewTransPerspectiveFromValues(2.0, 0.5, 0.1, 0.3, 1.5, 0.2, 10.0, 20.0, 1.0)
	original := *p

	if !p.Invert() {
		t.Error("Invert should succeed for non-singular matrix")
	}

	// Test that p * original = identity
	p.Multiply(&original)
	if !p.IsIdentity(1e-10) {
		t.Error("Inverted matrix times original should be identity")
	}
}

func TestInvertSingular(t *testing.T) {
	// Create singular matrix (determinant = 0)
	p := NewTransPerspectiveFromValues(1.0, 2.0, 0.0, 2.0, 4.0, 0.0, 0.0, 0.0, 0.0)

	if p.Invert() {
		t.Error("Invert should fail for singular matrix")
	}
}

func TestPerspectiveMultiply(t *testing.T) {
	p1 := NewTransPerspectiveFromValues(2.0, 0.0, 0.0, 0.0, 2.0, 0.0, 10.0, 20.0, 1.0) // scale 2x, translate
	p2 := NewTransPerspectiveFromValues(1.0, 0.0, 0.0, 0.0, 1.0, 0.0, 5.0, 10.0, 1.0)  // translate

	p1.Multiply(p2)

	// Test combined transformation
	x, y := 0.0, 0.0
	p1.Transform(&x, &y)
	if math.Abs(x-15.0) > perspectiveTestEpsilon || math.Abs(y-30.0) > perspectiveTestEpsilon {
		t.Errorf("Combined transformation failed, got (%f,%f)", x, y)
	}
}

func TestPerspectiveTransform(t *testing.T) {
	// Simple scaling transformation
	p := NewTransPerspectiveFromValues(2.0, 0.0, 0.0, 0.0, 3.0, 0.0, 0.0, 0.0, 1.0)

	x, y := 1.0, 1.0
	p.Transform(&x, &y)

	if math.Abs(x-2.0) > perspectiveTestEpsilon || math.Abs(y-3.0) > perspectiveTestEpsilon {
		t.Errorf("Transform failed, expected (2,3), got (%f,%f)", x, y)
	}
}

func TestTransformPerspective(t *testing.T) {
	// Test with actual perspective (w != 1)
	p := NewTransPerspectiveFromValues(1.0, 0.0, 0.1, 0.0, 1.0, 0.2, 0.0, 0.0, 1.0)

	x, y := 1.0, 1.0
	p.Transform(&x, &y)

	// With w = 1*0.1 + 1*0.2 + 1 = 1.3
	// Expected: x = 1/1.3, y = 1/1.3
	expectedX := 1.0 / 1.3
	expectedY := 1.0 / 1.3

	if math.Abs(x-expectedX) > perspectiveTestEpsilon || math.Abs(y-expectedY) > perspectiveTestEpsilon {
		t.Errorf("Perspective transform failed, expected (%f,%f), got (%f,%f)", expectedX, expectedY, x, y)
	}
}

func TestTransformAffine(t *testing.T) {
	p := NewTransPerspectiveFromValues(2.0, 0.5, 0.1, 0.3, 1.5, 0.2, 10.0, 20.0, 1.0)

	x, y := 1.0, 1.0
	p.TransformAffine(&x, &y)

	// Should apply only affine part: [2.0, 0.5; 0.3, 1.5] * [1,1] + [10,20]
	expectedX := 2.0*1.0 + 0.3*1.0 + 10.0 // 12.3
	expectedY := 0.5*1.0 + 1.5*1.0 + 20.0 // 22.0

	if math.Abs(x-expectedX) > perspectiveTestEpsilon || math.Abs(y-expectedY) > perspectiveTestEpsilon {
		t.Errorf("TransformAffine failed, expected (%f,%f), got (%f,%f)", expectedX, expectedY, x, y)
	}
}

func TestPerspectiveTransform2x2(t *testing.T) {
	p := NewTransPerspectiveFromValues(2.0, 0.5, 0.1, 0.3, 1.5, 0.2, 10.0, 20.0, 1.0)

	x, y := 1.0, 1.0
	p.Transform2x2(&x, &y)

	// Should apply only 2x2 part: [2.0, 0.5; 0.3, 1.5] * [1,1]
	expectedX := 2.0*1.0 + 0.3*1.0 // 2.3
	expectedY := 0.5*1.0 + 1.5*1.0 // 2.0

	if math.Abs(x-expectedX) > perspectiveTestEpsilon || math.Abs(y-expectedY) > perspectiveTestEpsilon {
		t.Errorf("Transform2x2 failed, expected (%f,%f), got (%f,%f)", expectedX, expectedY, x, y)
	}
}

func TestPerspectiveInverseTransform(t *testing.T) {
	p := NewTransPerspectiveFromValues(2.0, 0.0, 0.0, 0.0, 3.0, 0.0, 10.0, 20.0, 1.0)

	// Transform a point
	x1, y1 := 1.0, 1.0
	p.Transform(&x1, &y1)

	// Then inverse transform it back
	p.InverseTransform(&x1, &y1)

	if math.Abs(x1-1.0) > perspectiveTestEpsilon || math.Abs(y1-1.0) > perspectiveTestEpsilon {
		t.Errorf("InverseTransform failed, expected (1,1), got (%f,%f)", x1, y1)
	}
}

func TestStoreTo(t *testing.T) {
	p := NewTransPerspectiveFromValues(1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0)
	m := make([]float64, 9)
	p.StoreTo(m)

	expected := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0}
	for i, v := range expected {
		if math.Abs(m[i]-v) > perspectiveTestEpsilon {
			t.Errorf("StoreTo failed at index %d, expected %f, got %f", i, v, m[i])
		}
	}
}

func TestLoadFrom(t *testing.T) {
	p := NewTransPerspective()
	m := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0}
	p.LoadFrom(m)

	if p.SX != 1.0 || p.SHY != 2.0 || p.W0 != 3.0 ||
		p.SHX != 4.0 || p.SY != 5.0 || p.W1 != 6.0 ||
		p.TX != 7.0 || p.TY != 8.0 || p.W2 != 9.0 {
		t.Error("LoadFrom failed to set values correctly")
	}
}

func TestPerspectiveDeterminant(t *testing.T) {
	// Test identity matrix
	p := NewTransPerspective()
	det := p.Determinant()
	if math.Abs(det-1.0) > perspectiveTestEpsilon {
		t.Errorf("Identity determinant should be 1.0, got %f", det)
	}

	// Test known matrix
	p = NewTransPerspectiveFromValues(2.0, 0.0, 0.0, 0.0, 2.0, 0.0, 0.0, 0.0, 1.0)
	det = p.Determinant()
	if math.Abs(det-4.0) > perspectiveTestEpsilon {
		t.Errorf("Scale matrix determinant should be 4.0, got %f", det)
	}
}

func TestPerspectiveIsValid(t *testing.T) {
	p := NewTransPerspective()
	if !p.IsValid(perspectiveTestEpsilon) {
		t.Error("Identity matrix should be valid")
	}

	// Create invalid matrix
	p.SX = 0.0
	if p.IsValid(perspectiveTestEpsilon) {
		t.Error("Matrix with zero SX should be invalid")
	}
}

func TestPerspectiveIsIdentity(t *testing.T) {
	p := NewTransPerspective()
	if !p.IsIdentity(perspectiveTestEpsilon) {
		t.Error("New matrix should be identity")
	}

	p.TX = 0.1
	if p.IsIdentity(perspectiveTestEpsilon) {
		t.Error("Matrix with translation should not be identity")
	}
}

func TestPerspectiveIsEqual(t *testing.T) {
	p1 := NewTransPerspectiveFromValues(1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0)
	p2 := NewTransPerspectiveFromValues(1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0)
	p3 := NewTransPerspectiveFromValues(1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.1)

	if !p1.IsEqual(p2, perspectiveTestEpsilon) {
		t.Error("Identical matrices should be equal")
	}

	if p1.IsEqual(p3, perspectiveTestEpsilon) {
		t.Error("Different matrices should not be equal")
	}
}

func TestScaleFactor(t *testing.T) {
	// Test uniform scaling
	p := NewTransPerspectiveFromValues(2.0, 0.0, 0.0, 0.0, 2.0, 0.0, 0.0, 0.0, 1.0)
	scale := p.ScaleFactor()
	if math.Abs(scale-2.0) > 1e-9 {
		t.Errorf("Scale factor should be 2.0, got %f", scale)
	}
}

func TestTranslation(t *testing.T) {
	p := NewTransPerspectiveFromValues(1.0, 0.0, 0.0, 0.0, 1.0, 0.0, 10.0, 20.0, 1.0)
	tx, ty := p.Translation()
	if tx != 10.0 || ty != 20.0 {
		t.Errorf("Translation should be (10,20), got (%f,%f)", tx, ty)
	}
}

func TestScalingAbs(t *testing.T) {
	p := NewTransPerspectiveFromValues(3.0, 0.0, 0.0, 0.0, 4.0, 0.0, 0.0, 0.0, 1.0)
	sx, sy := p.ScalingAbs()
	if math.Abs(sx-3.0) > perspectiveTestEpsilon || math.Abs(sy-4.0) > perspectiveTestEpsilon {
		t.Errorf("ScalingAbs should be (3,4), got (%f,%f)", sx, sy)
	}
}

func TestIteratorX(t *testing.T) {
	// Test with identity transformation
	p := NewTransPerspective()
	it := p.NewIteratorX(0.0, 0.0, 1.0)

	// Should start at origin
	if math.Abs(it.X-0.0) > perspectiveTestEpsilon || math.Abs(it.Y-0.0) > perspectiveTestEpsilon {
		t.Errorf("Iterator should start at (0,0), got (%f,%f)", it.X, it.Y)
	}

	// Next step should be at (1,0)
	it.Next()
	if math.Abs(it.X-1.0) > perspectiveTestEpsilon || math.Abs(it.Y-0.0) > perspectiveTestEpsilon {
		t.Errorf("Iterator next should be at (1,0), got (%f,%f)", it.X, it.Y)
	}
}

func TestIteratorXWithTransform(t *testing.T) {
	// Test with scaling transformation
	p := NewTransPerspectiveFromValues(2.0, 0.0, 0.0, 0.0, 2.0, 0.0, 0.0, 0.0, 1.0)
	it := p.NewIteratorX(0.0, 0.0, 1.0)

	// Should start at origin
	if math.Abs(it.X-0.0) > perspectiveTestEpsilon || math.Abs(it.Y-0.0) > perspectiveTestEpsilon {
		t.Errorf("Iterator should start at (0,0), got (%f,%f)", it.X, it.Y)
	}

	// Next step should be at (2,0) due to 2x scaling
	it.Next()
	if math.Abs(it.X-2.0) > perspectiveTestEpsilon || math.Abs(it.Y-0.0) > perspectiveTestEpsilon {
		t.Errorf("Iterator next should be at (2,0), got (%f,%f)", it.X, it.Y)
	}
}

func TestOperationsChaining(t *testing.T) {
	p := NewTransPerspective()

	// Chain multiple operations
	p.Translate(10.0, 20.0).Scale(2.0).Translate(-5.0, -10.0)

	// Test final transformation
	x, y := 0.0, 0.0
	p.Transform(&x, &y)

	// Expected: translate(10,20) -> scale(2) -> translate(-5,-10)
	// (0,0) -> (10,20) -> (20,40) -> (15,30)
	if math.Abs(x-15.0) > perspectiveTestEpsilon || math.Abs(y-30.0) > perspectiveTestEpsilon {
		t.Errorf("Chained operations failed, expected (15,30), got (%f,%f)", x, y)
	}
}
