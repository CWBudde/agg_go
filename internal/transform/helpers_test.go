package transform

import (
	"math"
	"testing"
)

func TestNewTransAffineRotation(t *testing.T) {
	angle := math.Pi / 4 // 45 degrees
	m := NewTransAffineRotation(angle)

	expectedCos := math.Cos(angle)
	expectedSin := math.Sin(angle)

	if math.Abs(m.SX-expectedCos) > testEpsilon ||
		math.Abs(m.SHY-expectedSin) > testEpsilon ||
		math.Abs(m.SHX-(-expectedSin)) > testEpsilon ||
		math.Abs(m.SY-expectedCos) > testEpsilon {
		t.Error("NewTransAffineRotation should create correct rotation matrix")
	}

	if m.TX != 0.0 || m.TY != 0.0 {
		t.Error("Rotation matrix should have zero translation")
	}
}

func TestNewTransAffineScaling(t *testing.T) {
	scale := 2.5
	m := NewTransAffineScaling(scale)

	if m.SX != scale || m.SY != scale {
		t.Error("NewTransAffineScaling should create uniform scaling matrix")
	}

	if m.SHX != 0.0 || m.SHY != 0.0 || m.TX != 0.0 || m.TY != 0.0 {
		t.Error("Scaling matrix should have zero shear and translation")
	}
}

func TestNewTransAffineScalingXY(t *testing.T) {
	sx, sy := 2.0, 3.0
	m := NewTransAffineScalingXY(sx, sy)

	if m.SX != sx || m.SY != sy {
		t.Error("NewTransAffineScalingXY should create non-uniform scaling matrix")
	}

	if m.SHX != 0.0 || m.SHY != 0.0 || m.TX != 0.0 || m.TY != 0.0 {
		t.Error("Scaling matrix should have zero shear and translation")
	}
}

func TestNewTransAffineTranslation(t *testing.T) {
	tx, ty := 10.0, 20.0
	m := NewTransAffineTranslation(tx, ty)

	if m.TX != tx || m.TY != ty {
		t.Error("NewTransAffineTranslation should create translation matrix")
	}

	if m.SX != 1.0 || m.SY != 1.0 || m.SHX != 0.0 || m.SHY != 0.0 {
		t.Error("Translation matrix should have identity scaling and no shear")
	}
}

func TestNewTransAffineSkewing(t *testing.T) {
	sx, sy := math.Pi/6, math.Pi/4 // 30 and 45 degrees
	m := NewTransAffineSkewing(sx, sy)

	expectedShx := math.Tan(sx)
	expectedShy := math.Tan(sy)

	if math.Abs(m.SHX-expectedShx) > testEpsilon || math.Abs(m.SHY-expectedShy) > testEpsilon {
		t.Error("NewTransAffineSkewing should create correct skew matrix")
	}

	if m.SX != 1.0 || m.SY != 1.0 || m.TX != 0.0 || m.TY != 0.0 {
		t.Error("Skewing matrix should have unit scaling and no translation")
	}
}

func TestNewTransAffineLineSegment(t *testing.T) {
	// Test line from (0,0) to (3,4) with distance 1
	m := NewTransAffineLineSegment(0.0, 0.0, 3.0, 4.0, 1.0)

	// Apply to point (1, 0) - should map to end of line segment
	x, y := 1.0, 0.0
	m.Transform(&x, &y)

	// Line segment has length 5, so scaling factor should be 5
	// After rotation and translation, (1,0) should map to (3,4)
	if math.Abs(x-3.0) > testEpsilon || math.Abs(y-4.0) > testEpsilon {
		t.Errorf("NewTransAffineLineSegment failed: got (%f, %f), expected (3, 4)", x, y)
	}

	// Test with zero distance
	m = NewTransAffineLineSegment(0.0, 0.0, 3.0, 4.0, 0.0)
	x, y = 1.0, 0.0
	m.Transform(&x, &y)

	// Should only apply rotation and translation (no scaling)
	expectedAngle := math.Atan2(4.0, 3.0)
	expectedX := math.Cos(expectedAngle)
	expectedY := math.Sin(expectedAngle)

	if math.Abs(x-expectedX) > testEpsilon || math.Abs(y-expectedY) > testEpsilon {
		t.Error("NewTransAffineLineSegment with zero distance should not scale")
	}
}

func TestNewTransAffineReflectionUnit(t *testing.T) {
	// Test reflection across X-axis (unit vector (1, 0))
	m := NewTransAffineReflectionUnit(1.0, 0.0)

	// Point (1, 1) should reflect to (1, -1)
	x, y := 1.0, 1.0
	m.Transform(&x, &y)

	if math.Abs(x-1.0) > testEpsilon || math.Abs(y-(-1.0)) > testEpsilon {
		t.Errorf("Reflection across X-axis failed: got (%f, %f), expected (1, -1)", x, y)
	}

	// Test reflection across Y-axis (unit vector (0, 1))
	m = NewTransAffineReflectionUnit(0.0, 1.0)

	// Point (1, 1) should reflect to (-1, 1)
	x, y = 1.0, 1.0
	m.Transform(&x, &y)

	if math.Abs(x-(-1.0)) > testEpsilon || math.Abs(y-1.0) > testEpsilon {
		t.Errorf("Reflection across Y-axis failed: got (%f, %f), expected (-1, 1)", x, y)
	}
}

func TestNewTransAffineReflection(t *testing.T) {
	// Test reflection across 45-degree line (y = x)
	angle := math.Pi / 4
	m := NewTransAffineReflection(angle)

	// Point (1, 0) should reflect to (0, 1)
	x, y := 1.0, 0.0
	m.Transform(&x, &y)

	if math.Abs(x-0.0) > testEpsilon || math.Abs(y-1.0) > testEpsilon {
		t.Errorf("Reflection across 45° line failed: got (%f, %f), expected (0, 1)", x, y)
	}
}

func TestNewTransAffineReflectionXY(t *testing.T) {
	// Test reflection across line through origin with direction vector (1, 1)
	m := NewTransAffineReflectionXY(1.0, 1.0)

	// Should be same as reflection across 45-degree line
	// Point (1, 0) should reflect to (0, 1)
	x, y := 1.0, 0.0
	m.Transform(&x, &y)

	if math.Abs(x-0.0) > testEpsilon || math.Abs(y-1.0) > testEpsilon {
		t.Errorf("Reflection across (1,1) vector failed: got (%f, %f), expected (0, 1)", x, y)
	}

	// Test with zero vector - should return identity
	m = NewTransAffineReflectionXY(0.0, 0.0)
	if !m.IsIdentity(testEpsilon) {
		t.Error("Reflection with zero vector should return identity matrix")
	}
}

func TestRotateAround(t *testing.T) {
	m := NewTransAffine()
	angle := math.Pi / 2 // 90 degrees
	cx, cy := 1.0, 1.0   // Center of rotation

	m.RotateAround(angle, cx, cy)

	// Point (2, 1) should rotate around (1, 1) to (1, 2)
	x, y := 2.0, 1.0
	m.Transform(&x, &y)

	if math.Abs(x-1.0) > testEpsilon || math.Abs(y-2.0) > testEpsilon {
		t.Errorf("RotateAround failed: got (%f, %f), expected (1, 2)", x, y)
	}

	// Center point should remain unchanged
	x, y = cx, cy
	m.Transform(&x, &y)

	if math.Abs(x-cx) > testEpsilon || math.Abs(y-cy) > testEpsilon {
		t.Error("Center point should remain unchanged during rotation around it")
	}
}

func TestScaleAround(t *testing.T) {
	m := NewTransAffine()
	scale := 2.0
	cx, cy := 1.0, 1.0 // Center of scaling

	m.ScaleAround(scale, cx, cy)

	// Point (3, 3) should scale around (1, 1) to (5, 5)
	// Vector from center: (2, 2) -> scaled: (4, 4) -> final: (5, 5)
	x, y := 3.0, 3.0
	m.Transform(&x, &y)

	if math.Abs(x-5.0) > testEpsilon || math.Abs(y-5.0) > testEpsilon {
		t.Errorf("ScaleAround failed: got (%f, %f), expected (5, 5)", x, y)
	}

	// Center point should remain unchanged
	x, y = cx, cy
	m.Transform(&x, &y)

	if math.Abs(x-cx) > testEpsilon || math.Abs(y-cy) > testEpsilon {
		t.Error("Center point should remain unchanged during scaling around it")
	}
}

func TestScaleAroundXY(t *testing.T) {
	m := NewTransAffine()
	sx, sy := 2.0, 3.0
	cx, cy := 1.0, 1.0 // Center of scaling

	m.ScaleAroundXY(sx, sy, cx, cy)

	// Point (3, 2) should scale around (1, 1) to (5, 4)
	// Vector from center: (2, 1) -> scaled: (4, 3) -> final: (5, 4)
	x, y := 3.0, 2.0
	m.Transform(&x, &y)

	if math.Abs(x-5.0) > testEpsilon || math.Abs(y-4.0) > testEpsilon {
		t.Errorf("ScaleAroundXY failed: got (%f, %f), expected (5, 4)", x, y)
	}

	// Center point should remain unchanged
	x, y = cx, cy
	m.Transform(&x, &y)

	if math.Abs(x-cx) > testEpsilon || math.Abs(y-cy) > testEpsilon {
		t.Error("Center point should remain unchanged during scaling around it")
	}
}

func TestNewTransAffineRotateAround(t *testing.T) {
	angle := math.Pi / 2
	cx, cy := 1.0, 1.0

	m := NewTransAffineRotateAround(angle, cx, cy)

	// Should be equivalent to manual RotateAround
	expected := NewTransAffine()
	expected.RotateAround(angle, cx, cy)

	if !m.IsEqual(expected, testEpsilon) {
		t.Error("NewTransAffineRotateAround should be equivalent to RotateAround")
	}
}

func TestNewTransAffineScaleAround(t *testing.T) {
	scale := 2.0
	cx, cy := 1.0, 1.0

	m := NewTransAffineScaleAround(scale, cx, cy)

	// Should be equivalent to manual ScaleAround
	expected := NewTransAffine()
	expected.ScaleAround(scale, cx, cy)

	if !m.IsEqual(expected, testEpsilon) {
		t.Error("NewTransAffineScaleAround should be equivalent to ScaleAround")
	}
}

func TestNewTransAffineScaleAroundXY(t *testing.T) {
	sx, sy := 2.0, 3.0
	cx, cy := 1.0, 1.0

	m := NewTransAffineScaleAroundXY(sx, sy, cx, cy)

	// Should be equivalent to manual ScaleAroundXY
	expected := NewTransAffine()
	expected.ScaleAroundXY(sx, sy, cx, cy)

	if !m.IsEqual(expected, testEpsilon) {
		t.Error("NewTransAffineScaleAroundXY should be equivalent to ScaleAroundXY")
	}
}

func TestComplexTransformationChaining(t *testing.T) {
	// Test complex transformation: scale around point, then rotate around different point
	m := NewTransAffine()

	// First: scale by 2 around (1, 1)
	m.ScaleAround(2.0, 1.0, 1.0)

	// Then: rotate 90° around (0, 0)
	m.RotateAround(math.Pi/2, 0.0, 0.0)

	// Test point (2, 1)
	// After scale around (1,1): (2,1) -> (3,1)  [vector (1,0) becomes (2,0)]
	// After rotate 90° around (0,0): (3,1) -> (-1,3)
	x, y := 2.0, 1.0
	m.Transform(&x, &y)

	if math.Abs(x-(-1.0)) > testEpsilon || math.Abs(y-3.0) > testEpsilon {
		t.Errorf("Complex transformation chaining failed: got (%f, %f), expected (-1, 3)", x, y)
	}
}

func TestSpecializedConstructorConsistency(t *testing.T) {
	// Test that specialized constructors are consistent with manual construction

	// Rotation
	angle := math.Pi / 3
	rot1 := NewTransAffineRotation(angle)
	rot2 := NewTransAffine()
	rot2.Rotate(angle)

	if !rot1.IsEqual(rot2, testEpsilon) {
		t.Error("NewTransAffineRotation should be consistent with Rotate method")
	}

	// Translation
	tx, ty := 10.0, 20.0
	trans1 := NewTransAffineTranslation(tx, ty)
	trans2 := NewTransAffine()
	trans2.Translate(tx, ty)

	if !trans1.IsEqual(trans2, testEpsilon) {
		t.Error("NewTransAffineTranslation should be consistent with Translate method")
	}

	// Scaling
	sx, sy := 2.0, 3.0
	scale1 := NewTransAffineScalingXY(sx, sy)
	scale2 := NewTransAffine()
	scale2.ScaleXY(sx, sy)

	if !scale1.IsEqual(scale2, testEpsilon) {
		t.Error("NewTransAffineScalingXY should be consistent with ScaleXY method")
	}
}

func BenchmarkNewTransAffineRotation(b *testing.B) {
	angle := math.Pi / 4

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewTransAffineRotation(angle)
	}
}

func BenchmarkRotateAround(b *testing.B) {
	m := NewTransAffine()
	angle := math.Pi / 4
	cx, cy := 1.0, 1.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Reset()
		m.RotateAround(angle, cx, cy)
	}
}
