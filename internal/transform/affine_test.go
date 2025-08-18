package transform

import (
	"math"
	"testing"
)

const testEpsilon = 1e-10

func TestNewTransAffine(t *testing.T) {
	m := NewTransAffine()

	if !m.IsIdentity(testEpsilon) {
		t.Error("NewTransAffine should create identity matrix")
	}

	if m.SX != 1.0 || m.SY != 1.0 {
		t.Error("Identity matrix should have unit scaling")
	}

	if m.SHX != 0.0 || m.SHY != 0.0 {
		t.Error("Identity matrix should have zero shearing")
	}

	if m.TX != 0.0 || m.TY != 0.0 {
		t.Error("Identity matrix should have zero translation")
	}
}

func TestNewTransAffineFromValues(t *testing.T) {
	m := NewTransAffineFromValues(2.0, 0.5, 1.0, 3.0, 10.0, 20.0)

	if m.SX != 2.0 || m.SHY != 0.5 || m.SHX != 1.0 || m.SY != 3.0 || m.TX != 10.0 || m.TY != 20.0 {
		t.Error("NewTransAffineFromValues should set correct values")
	}
}

func TestNewTransAffineFromArray(t *testing.T) {
	values := [6]float64{2.0, 0.5, 1.0, 3.0, 10.0, 20.0}
	m := NewTransAffineFromArray(values)

	if m.SX != 2.0 || m.SHY != 0.5 || m.SHX != 1.0 || m.SY != 3.0 || m.TX != 10.0 || m.TY != 20.0 {
		t.Error("NewTransAffineFromArray should set correct values")
	}
}

func TestReset(t *testing.T) {
	m := NewTransAffineFromValues(2.0, 0.5, 1.0, 3.0, 10.0, 20.0)
	m.Reset()

	if !m.IsIdentity(testEpsilon) {
		t.Error("Reset should restore identity matrix")
	}
}

func TestTranslate(t *testing.T) {
	m := NewTransAffine()
	m.Translate(10.0, 20.0)

	if m.TX != 10.0 || m.TY != 20.0 {
		t.Error("Translate should set translation values")
	}

	// Test cumulative translation
	m.Translate(5.0, 8.0)
	if m.TX != 15.0 || m.TY != 28.0 {
		t.Error("Translate should accumulate translation values")
	}
}

func TestRotate(t *testing.T) {
	m := NewTransAffine()
	angle := math.Pi / 4 // 45 degrees
	m.Rotate(angle)

	expectedCos := math.Cos(angle)
	expectedSin := math.Sin(angle)

	if math.Abs(m.SX-expectedCos) > testEpsilon ||
		math.Abs(m.SHY-expectedSin) > testEpsilon ||
		math.Abs(m.SHX-(-expectedSin)) > testEpsilon ||
		math.Abs(m.SY-expectedCos) > testEpsilon {
		t.Error("Rotate should create correct rotation matrix")
	}
}

func TestScale(t *testing.T) {
	m := NewTransAffine()
	m.Scale(2.0)

	if m.SX != 2.0 || m.SY != 2.0 {
		t.Error("Scale should set uniform scaling")
	}

	// Test that other components are scaled too
	m.Reset()
	m.Translate(10.0, 20.0) // Set translation
	m.Scale(2.0)

	if m.TX != 20.0 || m.TY != 40.0 {
		t.Error("Scale should scale all matrix components including translation")
	}
}

func TestScaleXY(t *testing.T) {
	m := NewTransAffine()
	m.ScaleXY(2.0, 3.0)

	if m.SX != 2.0 || m.SY != 3.0 {
		t.Error("ScaleXY should set non-uniform scaling")
	}
}

func TestMultiply(t *testing.T) {
	// Test: translation * rotation
	m1 := NewTransAffineTranslation(10.0, 20.0)
	m2 := NewTransAffineRotation(math.Pi / 2) // 90 degrees

	result := m1.Copy()
	result.Multiply(m2)

	// Apply to point (1, 0)
	x, y := 1.0, 0.0
	result.Transform(&x, &y)

	// The order is: translate first, then rotate
	// (1, 0) -> translate -> (11, 20) -> rotate 90° -> (-20, 11)
	if math.Abs(x-(-20.0)) > testEpsilon || math.Abs(y-11.0) > testEpsilon {
		t.Errorf("Matrix multiplication failed: got (%f, %f), expected (-20, 11)", x, y)
	}
}

func TestInvert(t *testing.T) {
	// Test inversion of translation
	m := NewTransAffineTranslation(10.0, 20.0)
	original := m.Copy()
	m.Invert()

	if math.Abs(m.TX-(-10.0)) > testEpsilon || math.Abs(m.TY-(-20.0)) > testEpsilon {
		t.Error("Invert of translation should negate translation")
	}

	// Test that multiply by inverse gives identity
	m.Multiply(original)
	if !m.IsIdentity(testEpsilon) {
		t.Error("Matrix multiplied by its inverse should be identity")
	}
}

func TestDeterminant(t *testing.T) {
	// Identity matrix should have determinant 1
	m := NewTransAffine()
	if math.Abs(m.Determinant()-1.0) > testEpsilon {
		t.Error("Identity matrix should have determinant 1")
	}

	// Scaling matrix determinant should be product of scales
	m = NewTransAffineScalingXY(2.0, 3.0)
	if math.Abs(m.Determinant()-6.0) > testEpsilon {
		t.Error("Scaling matrix determinant should be product of scales")
	}
}

func TestTransform(t *testing.T) {
	// Test translation
	m := NewTransAffineTranslation(10.0, 20.0)
	x, y := 5.0, 8.0
	m.Transform(&x, &y)

	if math.Abs(x-15.0) > testEpsilon || math.Abs(y-28.0) > testEpsilon {
		t.Error("Translation transform failed")
	}

	// Test scaling
	m = NewTransAffineScalingXY(2.0, 3.0)
	x, y = 5.0, 8.0
	m.Transform(&x, &y)

	if math.Abs(x-10.0) > testEpsilon || math.Abs(y-24.0) > testEpsilon {
		t.Error("Scaling transform failed")
	}

	// Test rotation (90 degrees)
	m = NewTransAffineRotation(math.Pi / 2)
	x, y = 1.0, 0.0
	m.Transform(&x, &y)

	if math.Abs(x-0.0) > testEpsilon || math.Abs(y-1.0) > testEpsilon {
		t.Errorf("Rotation transform failed: got (%f, %f), expected (0, 1)", x, y)
	}
}

func TestTransform2x2(t *testing.T) {
	// Test that translation is ignored
	m := NewTransAffineFromValues(2.0, 0.0, 0.0, 3.0, 100.0, 200.0)
	x, y := 5.0, 8.0
	m.Transform2x2(&x, &y)

	if math.Abs(x-10.0) > testEpsilon || math.Abs(y-24.0) > testEpsilon {
		t.Error("Transform2x2 should ignore translation")
	}
}

func TestInverseTransform(t *testing.T) {
	m := NewTransAffineTranslation(10.0, 20.0)
	x, y := 15.0, 28.0
	m.InverseTransform(&x, &y)

	if math.Abs(x-5.0) > testEpsilon || math.Abs(y-8.0) > testEpsilon {
		t.Error("InverseTransform failed")
	}

	// Test round-trip: transform then inverse transform
	m = NewTransAffineFromValues(2.0, 0.5, 1.0, 3.0, 10.0, 20.0)
	originalX, originalY := 5.0, 8.0
	x, y = originalX, originalY

	m.Transform(&x, &y)
	m.InverseTransform(&x, &y)

	if math.Abs(x-originalX) > testEpsilon || math.Abs(y-originalY) > testEpsilon {
		t.Error("Round-trip transform should return original values")
	}
}

func TestIsValid(t *testing.T) {
	m := NewTransAffine()
	if !m.IsValid(testEpsilon) {
		t.Error("Identity matrix should be valid")
	}

	// Degenerate matrix (zero scaling)
	m.SX = 0.0
	if m.IsValid(testEpsilon) {
		t.Error("Matrix with zero SX should be invalid")
	}
}

func TestIsIdentity(t *testing.T) {
	m := NewTransAffine()
	if !m.IsIdentity(testEpsilon) {
		t.Error("NewTransAffine should create identity matrix")
	}

	m.Translate(0.1, 0.0)
	if m.IsIdentity(testEpsilon) {
		t.Error("Translated matrix should not be identity")
	}
}

func TestIsEqual(t *testing.T) {
	m1 := NewTransAffineFromValues(1.0, 2.0, 3.0, 4.0, 5.0, 6.0)
	m2 := NewTransAffineFromValues(1.0, 2.0, 3.0, 4.0, 5.0, 6.0)
	m3 := NewTransAffineFromValues(1.1, 2.0, 3.0, 4.0, 5.0, 6.0)

	if !m1.IsEqual(m2, testEpsilon) {
		t.Error("Identical matrices should be equal")
	}

	if m1.IsEqual(m3, testEpsilon) {
		t.Error("Different matrices should not be equal")
	}
}

func TestGetRotation(t *testing.T) {
	angle := math.Pi / 3 // 60 degrees
	m := NewTransAffineRotation(angle)
	extractedAngle := m.GetRotation()

	if math.Abs(extractedAngle-angle) > testEpsilon {
		t.Errorf("GetRotation failed: got %f, expected %f", extractedAngle, angle)
	}
}

func TestGetTranslation(t *testing.T) {
	tx, ty := 10.0, 20.0
	m := NewTransAffineTranslation(tx, ty)
	extractedTx, extractedTy := m.GetTranslation()

	if math.Abs(extractedTx-tx) > testEpsilon || math.Abs(extractedTy-ty) > testEpsilon {
		t.Error("GetTranslation failed")
	}
}

func TestGetScaling(t *testing.T) {
	sx, sy := 2.0, 3.0
	m := NewTransAffineScalingXY(sx, sy)
	extractedSx, extractedSy := m.GetScaling()

	if math.Abs(extractedSx-sx) > testEpsilon || math.Abs(extractedSy-sy) > testEpsilon {
		t.Error("GetScaling failed")
	}
}

func TestGetScalingAbs(t *testing.T) {
	// Test with shear to verify it calculates magnitude correctly
	m := NewTransAffineFromValues(3.0, 0.0, 4.0, 0.0, 0.0, 0.0) // Creates shearing
	sx, sy := m.GetScalingAbs()

	expectedSx := math.Sqrt(3.0*3.0 + 4.0*4.0) // 5.0
	expectedSy := 0.0

	if math.Abs(sx-expectedSx) > testEpsilon || math.Abs(sy-expectedSy) > testEpsilon {
		t.Errorf("GetScalingAbs failed: got (%f, %f), expected (%f, %f)", sx, sy, expectedSx, expectedSy)
	}
}

func TestFlipX(t *testing.T) {
	m := NewTransAffineFromValues(1.0, 2.0, 3.0, 4.0, 5.0, 6.0)
	m.FlipX()

	if m.SX != -1.0 || m.SHY != -2.0 || m.TX != -5.0 {
		t.Error("FlipX should negate SX, SHY, and TX")
	}

	if m.SHX != 3.0 || m.SY != 4.0 || m.TY != 6.0 {
		t.Error("FlipX should not change SHX, SY, and TY")
	}
}

func TestFlipY(t *testing.T) {
	m := NewTransAffineFromValues(1.0, 2.0, 3.0, 4.0, 5.0, 6.0)
	m.FlipY()

	if m.SHX != -3.0 || m.SY != -4.0 || m.TY != -6.0 {
		t.Error("FlipY should negate SHX, SY, and TY")
	}

	if m.SX != 1.0 || m.SHY != 2.0 || m.TX != 5.0 {
		t.Error("FlipY should not change SX, SHY, and TX")
	}
}

func TestStoreToLoadFrom(t *testing.T) {
	original := NewTransAffineFromValues(1.0, 2.0, 3.0, 4.0, 5.0, 6.0)
	array := make([]float64, 6)

	original.StoreTo(array)

	loaded := NewTransAffine()
	loaded.LoadFrom(array)

	if !original.IsEqual(loaded, testEpsilon) {
		t.Error("StoreTo/LoadFrom should preserve matrix values")
	}
}

func TestCopy(t *testing.T) {
	original := NewTransAffineFromValues(1.0, 2.0, 3.0, 4.0, 5.0, 6.0)
	copied := original.Copy()

	if !original.IsEqual(copied, testEpsilon) {
		t.Error("Copy should create identical matrix")
	}

	// Modify original and ensure copy is unchanged
	original.Translate(10.0, 20.0)
	if original.IsEqual(copied, testEpsilon) {
		t.Error("Copy should be independent of original")
	}
}

func TestPremultiply(t *testing.T) {
	// Test: premultiply rotation before translation
	m := NewTransAffineTranslation(10.0, 20.0)
	rotation := NewTransAffineRotation(math.Pi / 2) // 90 degrees

	m.Premultiply(rotation)

	// Apply to point (1, 0)
	x, y := 1.0, 0.0
	m.Transform(&x, &y)

	// Premultiply means rotation * translation
	// (1, 0) -> rotate 90° -> (0, 1) -> translate -> (10, 21)
	if math.Abs(x-10.0) > testEpsilon || math.Abs(y-21.0) > testEpsilon {
		t.Errorf("Premultiply failed: got (%f, %f), expected (10, 21)", x, y)
	}
}

func TestMultiplyInv(t *testing.T) {
	m1 := NewTransAffineTranslation(10.0, 20.0)
	m2 := NewTransAffineTranslation(5.0, 8.0)

	result := m1.Copy()
	result.MultiplyInv(m2)

	// Should be equivalent to m1 * m2.Inverse()
	expected := m1.Copy()
	m2Inv := m2.Copy().Invert()
	expected.Multiply(m2Inv)

	if !result.IsEqual(expected, testEpsilon) {
		t.Error("MultiplyInv should be equivalent to multiplying by inverse")
	}
}

func TestPremultiplyInv(t *testing.T) {
	m1 := NewTransAffineTranslation(10.0, 20.0)
	m2 := NewTransAffineTranslation(5.0, 8.0)

	result := m1.Copy()
	result.PremultiplyInv(m2)

	// Should be equivalent to m2.Inverse() * m1
	expected := m2.Copy().Invert()
	expected.Multiply(m1)

	if !result.IsEqual(expected, testEpsilon) {
		t.Error("PremultiplyInv should be equivalent to premultiplying by inverse")
	}
}

func TestTransformationChaining(t *testing.T) {
	// Test complex transformation chain: scale -> rotate -> translate
	m := NewTransAffine()
	m.ScaleXY(2.0, 3.0)
	m.Rotate(math.Pi / 2) // 90 degrees
	m.Translate(10.0, 20.0)

	// Apply to point (1, 1)
	x, y := 1.0, 1.0
	m.Transform(&x, &y)

	// Manual calculation:
	// 1. Scale: (1, 1) -> (2, 3)
	// 2. Rotate 90°: (2, 3) -> (-3, 2)
	// 3. Translate: (-3, 2) -> (7, 22)

	if math.Abs(x-7.0) > testEpsilon || math.Abs(y-22.0) > testEpsilon {
		t.Errorf("Transformation chaining failed: got (%f, %f), expected (7, 22)", x, y)
	}
}

func TestParlToParl(t *testing.T) {
	// Test parallelogram to parallelogram transformation
	src := [6]float64{0, 0, 1, 0, 1, 1} // Unit square (3 corners)
	dst := [6]float64{0, 0, 2, 0, 2, 2} // 2x2 square

	m := NewTransAffine()
	m.ParlToParl(src, dst)

	// Transform the three corners of source parallelogram
	testPoints := [][2]float64{{0, 0}, {1, 0}, {1, 1}}
	expectedPoints := [][2]float64{{0, 0}, {2, 0}, {2, 2}}

	for i, point := range testPoints {
		x, y := point[0], point[1]
		m.Transform(&x, &y)

		if math.Abs(x-expectedPoints[i][0]) > testEpsilon || math.Abs(y-expectedPoints[i][1]) > testEpsilon {
			t.Errorf("ParlToParl failed for point %d: got (%f, %f), expected (%f, %f)",
				i, x, y, expectedPoints[i][0], expectedPoints[i][1])
		}
	}
}

func TestRectToParl(t *testing.T) {
	// Test rectangle to parallelogram transformation
	parl := [6]float64{0, 0, 2, 0, 2, 2} // 2x2 square

	m := NewTransAffine()
	m.RectToParl(0, 0, 1, 1, parl) // Unit square to 2x2 square

	// Test corner transformation
	x, y := 1.0, 1.0 // Top-right corner of unit square
	m.Transform(&x, &y)

	if math.Abs(x-2.0) > testEpsilon || math.Abs(y-2.0) > testEpsilon {
		t.Errorf("RectToParl failed: got (%f, %f), expected (2, 2)", x, y)
	}
}

func TestEqual(t *testing.T) {
	m1 := NewTransAffineFromValues(1.0, 2.0, 3.0, 4.0, 5.0, 6.0)
	m2 := NewTransAffineFromValues(1.0, 2.0, 3.0, 4.0, 5.0, 6.0)
	m3 := NewTransAffineFromValues(1.1, 2.0, 3.0, 4.0, 5.0, 6.0)

	if !m1.Equal(m2) {
		t.Error("Equal should return true for identical matrices")
	}

	if m1.Equal(m3) {
		t.Error("Equal should return false for different matrices")
	}
}

func TestNotEqual(t *testing.T) {
	m1 := NewTransAffineFromValues(1.0, 2.0, 3.0, 4.0, 5.0, 6.0)
	m2 := NewTransAffineFromValues(1.0, 2.0, 3.0, 4.0, 5.0, 6.0)
	m3 := NewTransAffineFromValues(1.1, 2.0, 3.0, 4.0, 5.0, 6.0)

	if m1.NotEqual(m2) {
		t.Error("NotEqual should return false for identical matrices")
	}

	if !m1.NotEqual(m3) {
		t.Error("NotEqual should return true for different matrices")
	}
}

func TestMultiplyBy(t *testing.T) {
	m1 := NewTransAffineTranslation(10.0, 20.0)
	m2 := NewTransAffineScaling(2.0)

	result := m1.MultiplyBy(m2)

	// Original should be unchanged
	if !m1.IsEqual(NewTransAffineTranslation(10.0, 20.0), testEpsilon) {
		t.Error("MultiplyBy should not modify original matrix")
	}

	// Result should be m1 * m2
	expected := NewTransAffineTranslation(10.0, 20.0)
	expected.Multiply(m2)

	if !result.IsEqual(expected, testEpsilon) {
		t.Error("MultiplyBy should return correct multiplication result")
	}
}

func TestDivideBy(t *testing.T) {
	m1 := NewTransAffineTranslation(10.0, 20.0)
	m2 := NewTransAffineScaling(2.0)

	result := m1.DivideBy(m2)

	// Should be equivalent to m1 * m2.Inverse()
	expected := m1.Copy()
	expected.MultiplyInv(m2)

	if !result.IsEqual(expected, testEpsilon) {
		t.Error("DivideBy should be equivalent to multiplying by inverse")
	}
}

func TestInverse(t *testing.T) {
	m := NewTransAffineFromValues(2.0, 0.5, 1.0, 3.0, 10.0, 20.0)
	original := m.Copy()

	inverse := m.Inverse()

	// Original should be unchanged
	if !m.IsEqual(original, testEpsilon) {
		t.Error("Inverse should not modify original matrix")
	}

	// Multiply original by inverse should give identity
	result := original.MultiplyBy(inverse)
	if !result.IsIdentity(testEpsilon) {
		t.Error("Matrix multiplied by its inverse should be identity")
	}
}

func TestNumericalStability(t *testing.T) {
	// Test with very small but non-zero values
	m := NewTransAffineFromValues(1e-15, 0, 0, 1e-15, 0, 0)

	if m.IsValid(1e-14) {
		t.Error("Very small scaling should be considered invalid with appropriate epsilon")
	}

	// Test with large values
	m = NewTransAffineFromValues(1e10, 0, 0, 1e10, 1e10, 1e10)
	x, y := 1.0, 1.0
	m.Transform(&x, &y)

	if math.IsInf(x, 0) || math.IsInf(y, 0) || math.IsNaN(x) || math.IsNaN(y) {
		t.Error("Large value transformation should not produce infinite or NaN results")
	}
}

func BenchmarkTransform(b *testing.B) {
	m := NewTransAffineFromValues(2.0, 0.5, 1.0, 3.0, 10.0, 20.0)
	x, y := 5.0, 8.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Transform(&x, &y)
	}
}

func BenchmarkMultiply(b *testing.B) {
	m1 := NewTransAffineFromValues(2.0, 0.5, 1.0, 3.0, 10.0, 20.0)
	m2 := NewTransAffineFromValues(1.5, 0.2, 0.8, 2.5, 5.0, 15.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := m1.Copy()
		result.Multiply(m2)
	}
}

func BenchmarkInvert(b *testing.B) {
	m := NewTransAffineFromValues(2.0, 0.5, 1.0, 3.0, 10.0, 20.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := m.Copy()
		result.Invert()
	}
}
