package transform

import (
	"math"
	"testing"
)

const viewportTestEpsilon = 1e-10

func TestNewTransViewport(t *testing.T) {
	v := NewTransViewport()

	// Test default world viewport
	wx1, wy1, wx2, wy2 := v.GetWorldViewport()
	if wx1 != 0.0 || wy1 != 0.0 || wx2 != 1.0 || wy2 != 1.0 {
		t.Errorf("Default world viewport should be (0,0)-(1,1), got (%g,%g)-(%g,%g)", wx1, wy1, wx2, wy2)
	}

	// Test default device viewport
	dx1, dy1, dx2, dy2 := v.GetDeviceViewport()
	if dx1 != 0.0 || dy1 != 0.0 || dx2 != 1.0 || dy2 != 1.0 {
		t.Errorf("Default device viewport should be (0,0)-(1,1), got (%g,%g)-(%g,%g)", dx1, dy1, dx2, dy2)
	}

	// Test default aspect ratio and alignment
	if v.AspectRatio() != AspectRatioStretch {
		t.Error("Default aspect ratio should be AspectRatioStretch")
	}

	if v.AlignX() != 0.5 || v.AlignY() != 0.5 {
		t.Errorf("Default alignment should be (0.5, 0.5), got (%g, %g)", v.AlignX(), v.AlignY())
	}

	// Test validity
	if !v.IsValid() {
		t.Error("Default viewport should be valid")
	}
}

func TestViewportSettersGetters(t *testing.T) {
	v := NewTransViewport()

	// Test world viewport setting
	v.WorldViewport(10.0, 20.0, 110.0, 120.0)
	wx1, wy1, wx2, wy2 := v.GetWorldViewport()
	if wx1 != 10.0 || wy1 != 20.0 || wx2 != 110.0 || wy2 != 120.0 {
		t.Errorf("World viewport setter/getter failed, expected (10,20)-(110,120), got (%g,%g)-(%g,%g)", wx1, wy1, wx2, wy2)
	}

	// Test device viewport setting
	v.DeviceViewport(0.0, 0.0, 800.0, 600.0)
	dx1, dy1, dx2, dy2 := v.GetDeviceViewport()
	if dx1 != 0.0 || dy1 != 0.0 || dx2 != 800.0 || dy2 != 600.0 {
		t.Errorf("Device viewport setter/getter failed, expected (0,0)-(800,600), got (%g,%g)-(%g,%g)", dx1, dy1, dx2, dy2)
	}

	// Test aspect ratio setting
	v.PreserveAspectRatio(0.25, 0.75, AspectRatioMeet)
	if v.AlignX() != 0.25 || v.AlignY() != 0.75 || v.AspectRatio() != AspectRatioMeet {
		t.Errorf("Aspect ratio setting failed, expected (0.25, 0.75, Meet), got (%g, %g, %v)", v.AlignX(), v.AlignY(), v.AspectRatio())
	}
}

func TestBasicTransformation(t *testing.T) {
	v := NewTransViewport()
	v.WorldViewport(0.0, 0.0, 100.0, 100.0)
	v.DeviceViewport(0.0, 0.0, 800.0, 600.0)

	// Test scale factors
	expectedScaleX := 800.0 / 100.0 // 8.0
	expectedScaleY := 600.0 / 100.0 // 6.0

	if math.Abs(v.ScaleX()-expectedScaleX) > viewportTestEpsilon {
		t.Errorf("ScaleX failed, expected %g, got %g", expectedScaleX, v.ScaleX())
	}

	if math.Abs(v.ScaleY()-expectedScaleY) > viewportTestEpsilon {
		t.Errorf("ScaleY failed, expected %g, got %g", expectedScaleY, v.ScaleY())
	}

	// Test coordinate transformation
	worldX, worldY := 50.0, 25.0
	v.Transform(&worldX, &worldY)

	expectedDeviceX := 50.0 * 8.0 // 400.0
	expectedDeviceY := 25.0 * 6.0 // 150.0

	if math.Abs(worldX-expectedDeviceX) > viewportTestEpsilon || math.Abs(worldY-expectedDeviceY) > viewportTestEpsilon {
		t.Errorf("Transform failed, expected (%g,%g), got (%g,%g)", expectedDeviceX, expectedDeviceY, worldX, worldY)
	}

	// Test inverse transformation
	deviceX, deviceY := 400.0, 150.0
	v.InverseTransform(&deviceX, &deviceY)

	if math.Abs(deviceX-50.0) > viewportTestEpsilon || math.Abs(deviceY-25.0) > viewportTestEpsilon {
		t.Errorf("InverseTransform failed, expected (50,25), got (%g,%g)", deviceX, deviceY)
	}
}

func TestAspectRatioMeet(t *testing.T) {
	v := NewTransViewport()
	v.WorldViewport(0.0, 0.0, 100.0, 100.0)          // Square world
	v.DeviceViewport(0.0, 0.0, 200.0, 100.0)         // 2:1 rectangle device
	v.PreserveAspectRatio(0.5, 0.5, AspectRatioMeet) // Center alignment

	// With AspectRatioMeet, the square should fit entirely within the rectangle
	// The limiting factor is the Y dimension (device height 100), so scale should be 1.0
	// The world should be centered in X, using only part of device width

	_, actualWorldY1, _, actualWorldY2 := v.GetWorldViewportActual()

	// Y should remain unchanged
	if math.Abs(actualWorldY1-0.0) > viewportTestEpsilon || math.Abs(actualWorldY2-100.0) > viewportTestEpsilon {
		t.Errorf("AspectRatioMeet Y bounds should be unchanged, got (%g,%g)", actualWorldY1, actualWorldY2)
	}

	if v.ScaleX() != v.ScaleY() {
		t.Errorf("AspectRatioMeet should have equal scales, got X=%g, Y=%g", v.ScaleX(), v.ScaleY())
	}
}

func TestAspectRatioSlice(t *testing.T) {
	v := NewTransViewport()
	v.WorldViewport(0.0, 0.0, 100.0, 100.0)           // Square world
	v.DeviceViewport(0.0, 0.0, 200.0, 100.0)          // 2:1 rectangle device
	v.PreserveAspectRatio(0.5, 0.5, AspectRatioSlice) // Center alignment

	// With AspectRatioSlice, the square should fill the entire device viewport
	// The limiting factor is the X dimension (device width 200), so scale should be 2.0
	// Part of the world will be cropped in Y

	if v.ScaleX() != v.ScaleY() {
		t.Errorf("AspectRatioSlice should have equal scales, got X=%g, Y=%g", v.ScaleX(), v.ScaleY())
	}

	// Scale should be determined by the larger dimension ratio
	expectedScale := 200.0 / 100.0 // 2.0 (limited by X)
	if math.Abs(v.ScaleX()-expectedScale) > viewportTestEpsilon {
		t.Errorf("AspectRatioSlice scale should be %g, got %g", expectedScale, v.ScaleX())
	}
}

func TestAspectRatioStretch(t *testing.T) {
	v := NewTransViewport()
	v.WorldViewport(0.0, 0.0, 100.0, 100.0)  // Square world
	v.DeviceViewport(0.0, 0.0, 200.0, 100.0) // 2:1 rectangle device
	v.PreserveAspectRatio(0.5, 0.5, AspectRatioStretch)

	// With AspectRatioStretch, scales can be different
	expectedScaleX := 200.0 / 100.0 // 2.0
	expectedScaleY := 1.0           // 100.0 / 100.0

	if math.Abs(v.ScaleX()-expectedScaleX) > viewportTestEpsilon {
		t.Errorf("AspectRatioStretch ScaleX should be %g, got %g", expectedScaleX, v.ScaleX())
	}

	if math.Abs(v.ScaleY()-expectedScaleY) > viewportTestEpsilon {
		t.Errorf("AspectRatioStretch ScaleY should be %g, got %g", expectedScaleY, v.ScaleY())
	}

	// World viewport should remain unchanged
	actualWorldX1, actualWorldY1, actualWorldX2, actualWorldY2 := v.GetWorldViewportActual()
	if math.Abs(actualWorldX1-0.0) > viewportTestEpsilon || math.Abs(actualWorldY1-0.0) > viewportTestEpsilon ||
		math.Abs(actualWorldX2-100.0) > viewportTestEpsilon || math.Abs(actualWorldY2-100.0) > viewportTestEpsilon {
		t.Errorf("AspectRatioStretch should not change world bounds, got (%g,%g)-(%g,%g)", actualWorldX1, actualWorldY1, actualWorldX2, actualWorldY2)
	}
}

func TestInvalidViewport(t *testing.T) {
	v := NewTransViewport()

	// Test zero-width world viewport
	v.WorldViewport(100.0, 100.0, 100.0, 200.0)
	if v.IsValid() {
		t.Error("Zero-width world viewport should be invalid")
	}

	// Test zero-height device viewport
	v.WorldViewport(0.0, 0.0, 100.0, 100.0) // Reset to valid
	v.DeviceViewport(0.0, 100.0, 200.0, 100.0)
	if v.IsValid() {
		t.Error("Zero-height device viewport should be invalid")
	}
}

func TestScaleOnlyTransformations(t *testing.T) {
	v := NewTransViewport()
	v.WorldViewport(10.0, 20.0, 110.0, 120.0) // Offset world bounds
	v.DeviceViewport(0.0, 0.0, 200.0, 100.0)

	// Test scale-only transformation
	x, y := 5.0, 3.0
	v.TransformScaleOnly(&x, &y)

	expectedX := 5.0 * v.ScaleX()
	expectedY := 3.0 * v.ScaleY()

	if math.Abs(x-expectedX) > viewportTestEpsilon || math.Abs(y-expectedY) > viewportTestEpsilon {
		t.Errorf("TransformScaleOnly failed, expected (%g,%g), got (%g,%g)", expectedX, expectedY, x, y)
	}

	// Test inverse scale-only transformation
	v.InverseTransformScaleOnly(&x, &y)

	if math.Abs(x-5.0) > viewportTestEpsilon || math.Abs(y-3.0) > viewportTestEpsilon {
		t.Errorf("InverseTransformScaleOnly failed, expected (5,3), got (%g,%g)", x, y)
	}
}

func TestAffineConversion(t *testing.T) {
	v := NewTransViewport()
	v.WorldViewport(0.0, 0.0, 100.0, 100.0)
	v.DeviceViewport(10.0, 20.0, 110.0, 80.0)

	// Get the equivalent affine transformation
	affine := v.ToAffine()

	// Test that the affine transformation produces the same result
	worldX, worldY := 50.0, 25.0

	// Transform using viewport
	viewportX, viewportY := worldX, worldY
	v.Transform(&viewportX, &viewportY)

	// Transform using affine
	affineX, affineY := worldX, worldY
	affine.Transform(&affineX, &affineY)

	if math.Abs(viewportX-affineX) > viewportTestEpsilon || math.Abs(viewportY-affineY) > viewportTestEpsilon {
		t.Errorf("Affine conversion mismatch, viewport: (%g,%g), affine: (%g,%g)", viewportX, viewportY, affineX, affineY)
	}

	// Test scale-only affine conversion
	scaleAffine := v.ToAffineScaleOnly()
	if math.Abs(scaleAffine.SX-v.ScaleX()) > viewportTestEpsilon || math.Abs(scaleAffine.SY-v.ScaleY()) > viewportTestEpsilon {
		t.Errorf("Scale-only affine conversion failed, expected scales (%g,%g), got (%g,%g)", v.ScaleX(), v.ScaleY(), scaleAffine.SX, scaleAffine.SY)
	}
}

func TestDeviceOffsets(t *testing.T) {
	v := NewTransViewport()
	v.WorldViewport(10.0, 20.0, 110.0, 120.0)
	v.DeviceViewport(100.0, 200.0, 300.0, 400.0)

	// Test by transforming the world origin
	worldX, worldY := 10.0, 20.0 // World origin
	v.Transform(&worldX, &worldY)

	// The transformed world origin should be at the device origin
	expectedDeviceX := 100.0
	expectedDeviceY := 200.0

	if math.Abs(worldX-expectedDeviceX) > viewportTestEpsilon || math.Abs(worldY-expectedDeviceY) > viewportTestEpsilon {
		t.Errorf("Device offset calculation failed, world origin maps to (%g,%g), expected (%g,%g)", worldX, worldY, expectedDeviceX, expectedDeviceY)
	}
}

func TestSerialization(t *testing.T) {
	v1 := NewTransViewport()
	v1.WorldViewport(10.0, 20.0, 110.0, 120.0)
	v1.DeviceViewport(0.0, 0.0, 800.0, 600.0)
	v1.PreserveAspectRatio(0.25, 0.75, AspectRatioMeet)

	// Serialize
	data := make([]byte, v1.ByteSize())
	err := v1.Serialize(data)
	if err != nil {
		t.Fatalf("Serialization failed: %v", err)
	}

	// Deserialize
	v2 := NewTransViewport()
	err = v2.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialization failed: %v", err)
	}

	// Compare viewports
	wx1, wy1, wx2, wy2 := v1.GetWorldViewport()
	wx1_2, wy1_2, wx2_2, wy2_2 := v2.GetWorldViewport()

	if math.Abs(wx1-wx1_2) > viewportTestEpsilon || math.Abs(wy1-wy1_2) > viewportTestEpsilon ||
		math.Abs(wx2-wx2_2) > viewportTestEpsilon || math.Abs(wy2-wy2_2) > viewportTestEpsilon {
		t.Error("World viewport not preserved during serialization")
	}

	dx1, dy1, dx2, dy2 := v1.GetDeviceViewport()
	dx1_2, dy1_2, dx2_2, dy2_2 := v2.GetDeviceViewport()

	if math.Abs(dx1-dx1_2) > viewportTestEpsilon || math.Abs(dy1-dy1_2) > viewportTestEpsilon ||
		math.Abs(dx2-dx2_2) > viewportTestEpsilon || math.Abs(dy2-dy2_2) > viewportTestEpsilon {
		t.Error("Device viewport not preserved during serialization")
	}

	if v1.AlignX() != v2.AlignX() || v1.AlignY() != v2.AlignY() || v1.AspectRatio() != v2.AspectRatio() {
		t.Error("Aspect ratio settings not preserved during serialization")
	}

	if v1.IsValid() != v2.IsValid() {
		t.Error("Validity flag not preserved during serialization")
	}
}

func TestSerializationErrors(t *testing.T) {
	v := NewTransViewport()

	// Test insufficient buffer for serialization
	smallData := make([]byte, 10)
	err := v.Serialize(smallData)
	if err == nil {
		t.Error("Expected error for insufficient serialization buffer")
	}

	// Test insufficient data for deserialization
	err = v.Deserialize(smallData)
	if err == nil {
		t.Error("Expected error for insufficient deserialization data")
	}
}
