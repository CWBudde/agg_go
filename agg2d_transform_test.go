package agg

import (
	"math"
	"testing"

	"agg_go/internal/transform"
)

// Helper function to compare floats with tolerance
func floatEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) < tolerance
}

// Helper function to create a test Agg2D instance
func createTestAgg2D() *Agg2D {
	agg2d := &Agg2D{}
	agg2d.transform = transform.NewTransAffine()
	return agg2d
}

func TestTransformations_GetSet(t *testing.T) {
	agg2d := createTestAgg2D()

	// Test identity transformation
	tr := agg2d.GetTransformations()
	expected := [6]float64{1.0, 0.0, 0.0, 1.0, 0.0, 0.0}
	if tr.AffineMatrix != expected {
		t.Errorf("Expected identity matrix %v, got %v", expected, tr.AffineMatrix)
	}

	// Test setting transformations
	newTr := &Transformations{
		AffineMatrix: [6]float64{2.0, 0.5, 0.3, 1.5, 10.0, 20.0},
	}
	agg2d.SetTransformations(newTr)

	// Verify the transformation was set
	result := agg2d.GetTransformations()
	if result.AffineMatrix != newTr.AffineMatrix {
		t.Errorf("Expected %v, got %v", newTr.AffineMatrix, result.AffineMatrix)
	}
}

func TestBasicTransformations(t *testing.T) {
	const tolerance = 1e-10

	t.Run("Translate", func(t *testing.T) {
		agg2d := createTestAgg2D()
		agg2d.Translate(10.0, 20.0)

		x, y := agg2d.GetTranslation()
		if !floatEqual(x, 10.0, tolerance) || !floatEqual(y, 20.0, tolerance) {
			t.Errorf("Expected translation (10, 20), got (%f, %f)", x, y)
		}
	})

	t.Run("Scale", func(t *testing.T) {
		agg2d := createTestAgg2D()
		agg2d.Scale(2.0, 3.0)

		// Verify scaling
		tr := agg2d.GetTransformations()
		if !floatEqual(tr.AffineMatrix[0], 2.0, tolerance) || !floatEqual(tr.AffineMatrix[3], 3.0, tolerance) {
			t.Errorf("Expected scale factors (2, 3), got (%f, %f)", tr.AffineMatrix[0], tr.AffineMatrix[3])
		}
	})

	t.Run("UniformScale", func(t *testing.T) {
		agg2d := createTestAgg2D()
		agg2d.UniformScale(2.5)

		// Verify uniform scaling
		tr := agg2d.GetTransformations()
		if !floatEqual(tr.AffineMatrix[0], 2.5, tolerance) || !floatEqual(tr.AffineMatrix[3], 2.5, tolerance) {
			t.Errorf("Expected uniform scale 2.5, got (%f, %f)", tr.AffineMatrix[0], tr.AffineMatrix[3])
		}
	})

	t.Run("Rotate", func(t *testing.T) {
		agg2d := createTestAgg2D()
		angle := math.Pi / 4 // 45 degrees

		agg2d.Rotate(angle)

		rotation := agg2d.GetRotation()
		if !floatEqual(rotation, angle, tolerance) {
			t.Errorf("Expected rotation %f, got %f", angle, rotation)
		}
	})

	t.Run("Skew", func(t *testing.T) {
		agg2d := createTestAgg2D()
		agg2d.Skew(0.1, 0.2)

		// Verify skew was applied (transformation matrix should be modified)
		tr := agg2d.GetTransformations()
		if floatEqual(tr.AffineMatrix[1], 0.0, tolerance) && floatEqual(tr.AffineMatrix[2], 0.0, tolerance) {
			t.Error("Expected skew transformation to modify shear components")
		}
	})
}

func TestAdvancedTransformations(t *testing.T) {
	const tolerance = 1e-9

	t.Run("RotateAround", func(t *testing.T) {
		agg2d := createTestAgg2D()
		centerX, centerY := 100.0, 200.0
		angle := math.Pi / 2 // 90 degrees

		agg2d.RotateAround(centerX, centerY, angle)

		// Transform the center point - should remain unchanged
		x, y := centerX, centerY
		agg2d.WorldToScreen(&x, &y)
		if !floatEqual(x, centerX, tolerance) || !floatEqual(y, centerY, tolerance) {
			t.Errorf("Center point should remain unchanged: expected (%f, %f), got (%f, %f)",
				centerX, centerY, x, y)
		}
	})

	t.Run("ScaleAround", func(t *testing.T) {
		agg2d := createTestAgg2D()
		centerX, centerY := 50.0, 75.0
		scaleX, scaleY := 2.0, 3.0

		agg2d.ScaleAround(centerX, centerY, scaleX, scaleY)

		// Transform the center point - should remain unchanged
		x, y := centerX, centerY
		agg2d.WorldToScreen(&x, &y)
		if !floatEqual(x, centerX, tolerance) || !floatEqual(y, centerY, tolerance) {
			t.Errorf("Center point should remain unchanged: expected (%f, %f), got (%f, %f)",
				centerX, centerY, x, y)
		}
	})

	t.Run("FlipHorizontal", func(t *testing.T) {
		agg2d := createTestAgg2D()
		axisX := 100.0

		agg2d.FlipHorizontal(axisX)

		// Point on the axis should remain unchanged
		x, y := axisX, 50.0
		agg2d.WorldToScreen(&x, &y)
		if !floatEqual(x, axisX, tolerance) {
			t.Errorf("Point on flip axis should remain unchanged: expected x=%f, got x=%f", axisX, x)
		}

		// Point to the right should flip to the left
		rightX, rightY := axisX+10.0, 50.0
		agg2d.WorldToScreen(&rightX, &rightY)
		expectedX := axisX - 10.0
		if !floatEqual(rightX, expectedX, tolerance) {
			t.Errorf("Horizontal flip failed: expected x=%f, got x=%f", expectedX, rightX)
		}
	})

	t.Run("FlipVertical", func(t *testing.T) {
		agg2d := createTestAgg2D()
		axisY := 200.0

		agg2d.FlipVertical(axisY)

		// Point on the axis should remain unchanged
		x, y := 50.0, axisY
		agg2d.WorldToScreen(&x, &y)
		if !floatEqual(y, axisY, tolerance) {
			t.Errorf("Point on flip axis should remain unchanged: expected y=%f, got y=%f", axisY, y)
		}

		// Point above should flip below
		topX, topY := 50.0, axisY+10.0
		agg2d.WorldToScreen(&topX, &topY)
		expectedY := axisY - 10.0
		if !floatEqual(topY, expectedY, tolerance) {
			t.Errorf("Vertical flip failed: expected y=%f, got y=%f", expectedY, topY)
		}
	})
}

func TestTransformStack(t *testing.T) {
	agg2d := createTestAgg2D()

	// Initial state
	if agg2d.GetTransformStackDepth() != 0 {
		t.Error("Initial stack depth should be 0")
	}

	// Push a transformation
	agg2d.Scale(2.0, 2.0)
	agg2d.Translate(10.0, 20.0)
	agg2d.PushTransform()

	if agg2d.GetTransformStackDepth() != 1 {
		t.Error("Stack depth should be 1 after push")
	}

	// Modify current transformation
	agg2d.Rotate(math.Pi / 4)
	agg2d.Scale(0.5, 0.5)

	// Pop transformation
	success := agg2d.PopTransform()
	if !success {
		t.Error("PopTransform should succeed")
	}

	if agg2d.GetTransformStackDepth() != 0 {
		t.Error("Stack depth should be 0 after pop")
	}

	// Verify original transformation was restored
	x, y := agg2d.GetTranslation()
	const tolerance = 1e-10
	if !floatEqual(x, 10.0, tolerance) || !floatEqual(y, 20.0, tolerance) {
		t.Errorf("Expected restored translation (10, 20), got (%f, %f)", x, y)
	}

	// Test pop from empty stack
	success = agg2d.PopTransform()
	if success {
		t.Error("PopTransform should fail on empty stack")
	}
}

func TestViewportTransformations(t *testing.T) {
	const tolerance = 1e-9

	t.Run("Anisotropic", func(t *testing.T) {
		agg2d := createTestAgg2D()

		// Map world (0,0)-(100,100) to screen (0,0)-(200,400)
		agg2d.Viewport(0, 0, 100, 100, 0, 0, 200, 400, Anisotropic)

		// Test corner points
		x, y := 0.0, 0.0
		agg2d.WorldToScreen(&x, &y)
		if !floatEqual(x, 0, tolerance) || !floatEqual(y, 0, tolerance) {
			t.Errorf("Expected (0,0) -> (0,0), got (%f,%f)", x, y)
		}

		x, y = 100.0, 100.0
		agg2d.WorldToScreen(&x, &y)
		if !floatEqual(x, 200, tolerance) || !floatEqual(y, 400, tolerance) {
			t.Errorf("Expected (100,100) -> (200,400), got (%f,%f)", x, y)
		}

		x, y = 50.0, 50.0
		agg2d.WorldToScreen(&x, &y)
		if !floatEqual(x, 100, tolerance) || !floatEqual(y, 200, tolerance) {
			t.Errorf("Expected (50,50) -> (100,200), got (%f,%f)", x, y)
		}
	})

	t.Run("XMidYMid", func(t *testing.T) {
		agg2d := createTestAgg2D()

		// Map world (0,0)-(100,100) to screen (0,0)-(300,200) with center alignment
		agg2d.Viewport(0, 0, 100, 100, 0, 0, 300, 200, XMidYMid)

		// World center should map to screen center
		x, y := 50.0, 50.0
		agg2d.WorldToScreen(&x, &y)
		expectedX, expectedY := 150.0, 100.0 // Screen center
		if !floatEqual(x, expectedX, tolerance) || !floatEqual(y, expectedY, tolerance) {
			t.Errorf("Expected center (50,50) -> (%f,%f), got (%f,%f)",
				expectedX, expectedY, x, y)
		}
	})
}

func TestParallelogramTransformations(t *testing.T) {
	const tolerance = 1e-9

	agg2d := createTestAgg2D()

	// Define a parallelogram: (0,0) -> (10,0) -> (15,10) -> (5,10)
	agg2d.Parallelogram(0, 0, 10, 0, 5, 10)

	// Test unit square corners
	x, y := 0.0, 0.0
	agg2d.WorldToScreen(&x, &y)
	if !floatEqual(x, 0, tolerance) || !floatEqual(y, 0, tolerance) {
		t.Errorf("Expected (0,0) -> (0,0), got (%f,%f)", x, y)
	}

	x, y = 1.0, 0.0
	agg2d.WorldToScreen(&x, &y)
	if !floatEqual(x, 10, tolerance) || !floatEqual(y, 0, tolerance) {
		t.Errorf("Expected (1,0) -> (10,0), got (%f,%f)", x, y)
	}

	x, y = 0.0, 1.0
	agg2d.WorldToScreen(&x, &y)
	if !floatEqual(x, 5, tolerance) || !floatEqual(y, 10, tolerance) {
		t.Errorf("Expected (0,1) -> (5,10), got (%f,%f)", x, y)
	}
}

func TestWorldScreenConversions(t *testing.T) {
	const tolerance = 1e-9

	agg2d := createTestAgg2D()
	agg2d.Scale(2.0, 3.0)
	agg2d.Translate(10.0, 20.0)

	t.Run("Point transformation", func(t *testing.T) {
		worldX, worldY := 5.0, 8.0
		screenX, screenY := worldX, worldY
		agg2d.WorldToScreen(&screenX, &screenY)

		// Verify round-trip conversion
		backWorldX, backWorldY := screenX, screenY
		agg2d.ScreenToWorld(&backWorldX, &backWorldY)
		ok := true
		if !ok {
			t.Error("ScreenToWorld should succeed")
		}

		if !floatEqual(backWorldX, worldX, tolerance) || !floatEqual(backWorldY, worldY, tolerance) {
			t.Errorf("Round-trip failed: (%f,%f) -> (%f,%f) -> (%f,%f)",
				worldX, worldY, screenX, screenY, backWorldX, backWorldY)
		}
	})

	t.Run("Distance transformation", func(t *testing.T) {
		worldDistance := 10.0
		screenDistance := agg2d.WorldToScreenDistance(worldDistance)

		backWorldDistance, ok := agg2d.ScreenToWorldDistance(screenDistance)
		if !ok {
			t.Error("ScreenToWorldDistance should succeed")
		}

		if !floatEqual(backWorldDistance, worldDistance, tolerance) {
			t.Errorf("Distance round-trip failed: %f -> %f -> %f",
				worldDistance, screenDistance, backWorldDistance)
		}
	})

	t.Run("Rectangle transformation", func(t *testing.T) {
		worldX1, worldY1, worldX2, worldY2 := 1.0, 2.0, 5.0, 8.0
		screenX1, screenY1, screenX2, screenY2 := agg2d.WorldToScreenRect(worldX1, worldY1, worldX2, worldY2)

		backWorldX1, backWorldY1, backWorldX2, backWorldY2, ok := agg2d.ScreenToWorldRect(screenX1, screenY1, screenX2, screenY2)
		if !ok {
			t.Error("ScreenToWorldRect should succeed")
		}

		// For axis-aligned transformations, rectangles should map back exactly
		if agg2d.IsAxisAligned() {
			if !floatEqual(backWorldX1, worldX1, tolerance) || !floatEqual(backWorldY1, worldY1, tolerance) ||
				!floatEqual(backWorldX2, worldX2, tolerance) || !floatEqual(backWorldY2, worldY2, tolerance) {
				t.Errorf("Rectangle round-trip failed: (%f,%f,%f,%f) -> (%f,%f,%f,%f)",
					worldX1, worldY1, worldX2, worldY2, backWorldX1, backWorldY1, backWorldX2, backWorldY2)
			}
		}
	})
}

func TestVectorTransformations(t *testing.T) {
	const tolerance = 1e-9

	agg2d := createTestAgg2D()
	agg2d.Scale(2.0, 3.0)
	agg2d.Rotate(math.Pi / 6) // 30 degrees

	vectorX, vectorY := 1.0, 0.0
	screenVectorX, screenVectorY := agg2d.TransformVector(vectorX, vectorY)

	backVectorX, backVectorY, ok := agg2d.InverseTransformVector(screenVectorX, screenVectorY)
	if !ok {
		t.Error("InverseTransformVector should succeed")
	}

	if !floatEqual(backVectorX, vectorX, tolerance) || !floatEqual(backVectorY, vectorY, tolerance) {
		t.Errorf("Vector round-trip failed: (%f,%f) -> (%f,%f) -> (%f,%f)",
			vectorX, vectorY, screenVectorX, screenVectorY, backVectorX, backVectorY)
	}
}

func TestTransformationQueries(t *testing.T) {
	agg2d := createTestAgg2D()

	// Test identity queries
	if !agg2d.IsIdentity() {
		t.Error("New transformation should be identity")
	}

	if !agg2d.IsTranslationOnly() {
		t.Error("Identity transformation should be translation-only")
	}

	if !agg2d.IsAxisAligned() {
		t.Error("Identity transformation should be axis-aligned")
	}

	if !agg2d.HasUniformScaling() {
		t.Error("Identity transformation should have uniform scaling")
	}

	// Test after scaling
	agg2d.Scale(2.0, 2.0)
	if agg2d.IsIdentity() {
		t.Error("Scaled transformation should not be identity")
	}

	if !agg2d.HasUniformScaling() {
		t.Error("Uniform scaling should be detected")
	}

	// Test after non-uniform scaling
	agg2d.Scale(1.0, 2.0)
	if agg2d.HasUniformScaling() {
		t.Error("Non-uniform scaling should be detected")
	}

	// Test after rotation
	agg2d.Rotate(math.Pi / 4)
	if agg2d.IsAxisAligned() {
		t.Error("Rotated transformation should not be axis-aligned")
	}

	// Test determinant and validity
	det := agg2d.Determinant()
	if det == 0 {
		t.Error("Valid transformation should have non-zero determinant")
	}

	if !agg2d.IsValid() {
		t.Error("Non-singular transformation should be valid")
	}
}

func TestDecomposeTransform(t *testing.T) {
	const tolerance = 1e-6 // Slightly larger tolerance for decomposition

	agg2d := createTestAgg2D()

	// Apply known transformations
	scaleX, scaleY := 2.0, 3.0
	rotation := math.Pi / 6 // 30 degrees
	translateX, translateY := 10.0, 20.0

	agg2d.Scale(scaleX, scaleY)
	agg2d.Rotate(rotation)
	agg2d.Translate(translateX, translateY)

	// Decompose
	components := agg2d.DecomposeTransform()

	// Check translation (should be exact)
	if !floatEqual(components.TranslateX, translateX, tolerance) ||
		!floatEqual(components.TranslateY, translateY, tolerance) {
		t.Errorf("Translation decomposition failed: expected (%f,%f), got (%f,%f)",
			translateX, translateY, components.TranslateX, components.TranslateY)
	}

	// Check rotation (should be close)
	if !floatEqual(components.Rotation, rotation, tolerance) {
		t.Errorf("Rotation decomposition failed: expected %f, got %f",
			rotation, components.Rotation)
	}

	// Check scaling (order matters due to rotation)
	expectedScaleX := scaleX
	expectedScaleY := scaleY
	if !floatEqual(components.ScaleX, expectedScaleX, tolerance) ||
		!floatEqual(components.ScaleY, expectedScaleY, tolerance) {
		t.Errorf("Scale decomposition failed: expected (%f,%f), got (%f,%f)",
			expectedScaleX, expectedScaleY, components.ScaleX, components.ScaleY)
	}
}

func TestInvertTransform(t *testing.T) {
	const tolerance = 1e-10

	agg2d := createTestAgg2D()
	agg2d.Scale(2.0, 3.0)
	agg2d.Rotate(math.Pi / 4)
	agg2d.Translate(10.0, 20.0)

	inverse := agg2d.InvertTransform()
	if inverse == nil {
		t.Error("InvertTransform should succeed for valid transformation")
	}

	// Test by transforming a point and then applying inverse
	originalX, originalY := 5.0, 8.0
	transformedX, transformedY := originalX, originalY
	agg2d.WorldToScreen(&transformedX, &transformedY)

	// Apply inverse transformation
	backX, backY := transformedX, transformedY
	inverse.Transform(&backX, &backY)

	if !floatEqual(backX, originalX, tolerance) || !floatEqual(backY, originalY, tolerance) {
		t.Errorf("Inverse transformation failed: (%f,%f) -> (%f,%f) -> (%f,%f)",
			originalX, originalY, transformedX, transformedY, backX, backY)
	}
}

// Benchmark tests
func BenchmarkWorldToScreen(b *testing.B) {
	agg2d := createTestAgg2D()
	agg2d.Scale(2.0, 3.0)
	agg2d.Rotate(math.Pi / 6)
	agg2d.Translate(100.0, 200.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x, y := float64(i%1000), float64(i%500)
		agg2d.WorldToScreen(&x, &y)
	}
}

func BenchmarkTransformStack(b *testing.B) {
	agg2d := createTestAgg2D()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agg2d.PushTransform()
		agg2d.Scale(1.1, 1.1)
		agg2d.Rotate(0.01)
		agg2d.PopTransform()
	}
}

func BenchmarkViewportTransform(b *testing.B) {
	agg2d := createTestAgg2D()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agg2d.ResetTransform()
		agg2d.Viewport(0, 0, 1000, 1000, 0, 0, 800, 600, XMidYMid)
	}
}
