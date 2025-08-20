package span

import (
	"math"
	"testing"

	"agg_go/internal/transform"
)

// MockTransformer is a simple transformer for testing
type MockTransformer struct {
	scaleX, scaleY   float64
	offsetX, offsetY float64
}

func NewMockTransformer(scaleX, scaleY, offsetX, offsetY float64) *MockTransformer {
	return &MockTransformer{
		scaleX: scaleX, scaleY: scaleY,
		offsetX: offsetX, offsetY: offsetY,
	}
}

func (m *MockTransformer) Transform(x, y *float64) {
	*x = *x*m.scaleX + m.offsetX
	*y = *y*m.scaleY + m.offsetY
}

// RotationTransformer rotates coordinates by the given angle
type RotationTransformer struct {
	sin, cos float64
}

func NewRotationTransformer(angleRad float64) *RotationTransformer {
	return &RotationTransformer{
		sin: math.Sin(angleRad),
		cos: math.Cos(angleRad),
	}
}

func (r *RotationTransformer) Transform(x, y *float64) {
	oldX := *x
	*x = oldX*r.cos - (*y)*r.sin
	*y = oldX*r.sin + (*y)*r.cos
}

func TestSpanInterpolatorTrans_BasicFunctionality(t *testing.T) {
	t.Run("IdentityTransform", func(t *testing.T) {
		// Identity transform (scale 1, offset 0)
		transformer := NewMockTransformer(1.0, 1.0, 0.0, 0.0)
		interp := NewSpanInterpolatorTrans(transformer)

		// Begin at (0, 0)
		interp.Begin(0, 0, 10)

		x, y := interp.Coordinates()
		if x != 0 || y != 0 {
			t.Errorf("Identity transform at origin: got (%d, %d), want (0, 0)", x, y)
		}

		// Advance and check
		interp.Next()
		x, y = interp.Coordinates()
		expectedX := 1 << defaultSubpixelShift // 1.0 * 256
		if x != expectedX || y != 0 {
			t.Errorf("After Next(): got (%d, %d), want (%d, 0)", x, y, expectedX)
		}
	})

	t.Run("ScaleTransform", func(t *testing.T) {
		// Scale by 2x in both dimensions
		transformer := NewMockTransformer(2.0, 2.0, 0.0, 0.0)
		interp := NewSpanInterpolatorTrans(transformer)

		interp.Begin(1, 1, 5)

		x, y := interp.Coordinates()
		expectedX := 2 << defaultSubpixelShift // 2.0 * 256
		expectedY := 2 << defaultSubpixelShift // 2.0 * 256
		if x != expectedX || y != expectedY {
			t.Errorf("Scale 2x at (1,1): got (%d, %d), want (%d, %d)", x, y, expectedX, expectedY)
		}

		// Advance to x=2
		interp.Next()
		x, y = interp.Coordinates()
		expectedX = 4 << defaultSubpixelShift // 4.0 * 256 (x=2 scaled by 2)
		expectedY = 2 << defaultSubpixelShift // 2.0 * 256 (y=1 scaled by 2)
		if x != expectedX || y != expectedY {
			t.Errorf("Scale 2x after Next(): got (%d, %d), want (%d, %d)", x, y, expectedX, expectedY)
		}
	})

	t.Run("TranslationTransform", func(t *testing.T) {
		// Translate by (10, 20)
		transformer := NewMockTransformer(1.0, 1.0, 10.0, 20.0)
		interp := NewSpanInterpolatorTrans(transformer)

		interp.Begin(0, 0, 5)

		x, y := interp.Coordinates()
		expectedX := 10 << defaultSubpixelShift // 10.0 * 256
		expectedY := 20 << defaultSubpixelShift // 20.0 * 256
		if x != expectedX || y != expectedY {
			t.Errorf("Translation (10,20) at origin: got (%d, %d), want (%d, %d)", x, y, expectedX, expectedY)
		}

		// Advance
		interp.Next()
		x, y = interp.Coordinates()
		expectedX = 11 << defaultSubpixelShift // 11.0 * 256 (x=1 + 10)
		expectedY = 20 << defaultSubpixelShift // 20.0 * 256 (y=0 + 20)
		if x != expectedX || y != expectedY {
			t.Errorf("Translation after Next(): got (%d, %d), want (%d, %d)", x, y, expectedX, expectedY)
		}
	})
}

func TestSpanInterpolatorTrans_RotationTransform(t *testing.T) {
	t.Run("90DegreeRotation", func(t *testing.T) {
		// 90-degree rotation: (x,y) -> (-y,x)
		transformer := NewRotationTransformer(math.Pi / 2)
		interp := NewSpanInterpolatorTrans(transformer)

		// Start at (1, 0) - should rotate to (0, 1)
		interp.Begin(1, 0, 5)

		x, y := interp.Coordinates()
		expectedX := 0                         // 0.0 * 256
		expectedY := 1 << defaultSubpixelShift // 1.0 * 256

		// Allow small floating point errors
		if abs(x-expectedX) > 1 || abs(y-expectedY) > 1 {
			t.Errorf("90° rotation of (1,0): got (%d, %d), want (~%d, ~%d)", x, y, expectedX, expectedY)
		}

		// Advance to (2, 0) - should rotate to (0, 2)
		interp.Next()
		x, y = interp.Coordinates()
		expectedX = 0                         // 0.0 * 256
		expectedY = 2 << defaultSubpixelShift // 2.0 * 256

		if abs(x-expectedX) > 1 || abs(y-expectedY) > 1 {
			t.Errorf("90° rotation of (2,0): got (%d, %d), want (~%d, ~%d)", x, y, expectedX, expectedY)
		}
	})

	t.Run("45DegreeRotation", func(t *testing.T) {
		// 45-degree rotation
		transformer := NewRotationTransformer(math.Pi / 4)
		interp := NewSpanInterpolatorTrans(transformer)

		// Start at (1, 1)
		interp.Begin(1, 1, 3)

		x, y := interp.Coordinates()
		// (1,1) rotated 45° should be approximately (0, √2)
		expectedY := int(math.Sqrt(2) * float64(defaultSubpixelScale))

		// Check that rotation occurred (y should be positive and larger)
		if y <= defaultSubpixelScale || abs(x) > 10 {
			t.Errorf("45° rotation of (1,1): got (%d, %d), expected x≈0, y≈%d", x, y, expectedY)
		}
	})
}

func TestSpanInterpolatorTrans_AffineTransform(t *testing.T) {
	t.Run("AffineTransformation", func(t *testing.T) {
		// Create an affine transformation (scale + rotate + translate)
		affine := transform.NewTransAffine()
		affine.ScaleXY(2.0, 1.5)
		affine.Rotate(math.Pi / 6) // 30 degrees
		affine.Translate(10, 20)

		interp := NewSpanInterpolatorTrans(affine)
		interp.Begin(0, 0, 5)

		// Just verify that transformation is applied (exact values depend on complex math)
		x, y := interp.Coordinates()
		if x == 0 && y == 0 {
			t.Error("Affine transform should not result in (0,0) for non-zero translation")
		}

		// Verify coordinates change when advancing
		prevX, prevY := x, y
		interp.Next()
		x, y = interp.Coordinates()

		if x == prevX && y == prevY {
			t.Error("Coordinates should change after Next() with affine transform")
		}
	})
}

func TestSpanInterpolatorTrans_CustomSubpixelShift(t *testing.T) {
	t.Run("SubpixelShift4", func(t *testing.T) {
		transformer := NewMockTransformer(1.0, 1.0, 0.0, 0.0)
		interp := NewSpanInterpolatorTransWithShift(transformer, 4) // 2^4 = 16

		if interp.SubpixelShift() != 4 {
			t.Errorf("SubpixelShift: got %d, want 4", interp.SubpixelShift())
		}

		if interp.SubpixelScale() != 16 {
			t.Errorf("SubpixelScale: got %d, want 16", interp.SubpixelScale())
		}

		interp.Begin(1, 1, 5)
		x, y := interp.Coordinates()
		expectedX := 1 << 4 // 1.0 * 16
		expectedY := 1 << 4 // 1.0 * 16

		if x != expectedX || y != expectedY {
			t.Errorf("Custom subpixel shift: got (%d, %d), want (%d, %d)", x, y, expectedX, expectedY)
		}
	})
}

func TestSpanInterpolatorTrans_TransformerAccess(t *testing.T) {
	t.Run("GetSetTransformer", func(t *testing.T) {
		transformer1 := NewMockTransformer(1.0, 1.0, 0.0, 0.0)
		transformer2 := NewMockTransformer(2.0, 2.0, 5.0, 5.0)

		interp := NewSpanInterpolatorTrans(transformer1)

		// Verify initial transformer
		if interp.Transformer() != transformer1 {
			t.Error("Initial transformer not set correctly")
		}

		// Change transformer
		interp.SetTransformer(transformer2)
		if interp.Transformer() != transformer2 {
			t.Error("Transformer not changed correctly")
		}

		// Verify new transformer is used
		interp.Begin(0, 0, 5)
		x, y := interp.Coordinates()
		expectedX := 5 << defaultSubpixelShift // offset 5
		expectedY := 5 << defaultSubpixelShift // offset 5

		if x != expectedX || y != expectedY {
			t.Errorf("New transformer not applied: got (%d, %d), want (%d, %d)", x, y, expectedX, expectedY)
		}
	})
}

func TestSpanInterpolatorTrans_EdgeCases(t *testing.T) {
	t.Run("NegativeCoordinates", func(t *testing.T) {
		transformer := NewMockTransformer(1.0, 1.0, 0.0, 0.0)
		interp := NewSpanInterpolatorTrans(transformer)

		interp.Begin(-5, -3, 5)
		x, y := interp.Coordinates()
		expectedX := -5 << defaultSubpixelShift
		expectedY := -3 << defaultSubpixelShift

		if x != expectedX || y != expectedY {
			t.Errorf("Negative coordinates: got (%d, %d), want (%d, %d)", x, y, expectedX, expectedY)
		}
	})

	t.Run("ZeroLength", func(t *testing.T) {
		transformer := NewMockTransformer(2.0, 2.0, 0.0, 0.0)
		interp := NewSpanInterpolatorTrans(transformer)

		// Length parameter should be ignored
		interp.Begin(1, 1, 0)
		x, y := interp.Coordinates()
		expectedX := 2 << defaultSubpixelShift
		expectedY := 2 << defaultSubpixelShift

		if x != expectedX || y != expectedY {
			t.Errorf("Zero length: got (%d, %d), want (%d, %d)", x, y, expectedX, expectedY)
		}
	})

	t.Run("LargeCoordinates", func(t *testing.T) {
		transformer := NewMockTransformer(1.0, 1.0, 0.0, 0.0)
		interp := NewSpanInterpolatorTrans(transformer)

		largeX, largeY := 100000.0, 200000.0
		interp.Begin(largeX, largeY, 5)

		x, y := interp.Coordinates()
		expectedX := int(largeX * float64(defaultSubpixelScale))
		expectedY := int(largeY * float64(defaultSubpixelScale))

		if x != expectedX || y != expectedY {
			t.Errorf("Large coordinates: got (%d, %d), want (%d, %d)", x, y, expectedX, expectedY)
		}
	})
}

func TestSpanInterpolatorTrans_ConstructorVariants(t *testing.T) {
	t.Run("NewSpanInterpolatorTransAtPoint", func(t *testing.T) {
		transformer := NewMockTransformer(1.0, 1.0, 10.0, 20.0)
		interp := NewSpanInterpolatorTransAtPoint(transformer, 5, 7, 10)

		// Should already be at the starting point
		x, y := interp.Coordinates()
		expectedX := 15 << defaultSubpixelShift // 5 + 10
		expectedY := 27 << defaultSubpixelShift // 7 + 20

		if x != expectedX || y != expectedY {
			t.Errorf("Constructor with point: got (%d, %d), want (%d, %d)", x, y, expectedX, expectedY)
		}
	})
}

func TestSpanInterpolatorTrans_SpanIteration(t *testing.T) {
	t.Run("MultiplePixelSpan", func(t *testing.T) {
		// Test iterating across multiple pixels
		transformer := NewMockTransformer(1.0, 1.0, 0.0, 0.0)
		interp := NewSpanInterpolatorTrans(transformer)

		startX, startY := 10.0, 5.0
		interp.Begin(startX, startY, 5)

		// Collect coordinates for 5 pixels
		coords := make([][2]int, 5)
		for i := 0; i < 5; i++ {
			x, y := interp.Coordinates()
			coords[i] = [2]int{x, y}
			if i < 4 {
				interp.Next()
			}
		}

		// Verify x advances by 1.0 each step, y stays constant
		for i := 0; i < 5; i++ {
			expectedX := int((startX + float64(i)) * float64(defaultSubpixelScale))
			expectedY := int(startY * float64(defaultSubpixelScale))

			if coords[i][0] != expectedX || coords[i][1] != expectedY {
				t.Errorf("Pixel %d: got (%d, %d), want (%d, %d)",
					i, coords[i][0], coords[i][1], expectedX, expectedY)
			}
		}
	})
}

// Benchmark tests
func BenchmarkSpanInterpolatorTrans_IdentityTransform(b *testing.B) {
	transformer := NewMockTransformer(1.0, 1.0, 0.0, 0.0)
	interp := NewSpanInterpolatorTrans(transformer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		interp.Begin(0, 0, 1000)
		for j := 0; j < 1000; j++ {
			interp.Next()
			interp.Coordinates()
		}
	}
}

func BenchmarkSpanInterpolatorTrans_AffineTransform(b *testing.B) {
	affine := transform.NewTransAffine()
	affine.ScaleXY(1.5, 1.2)
	affine.Rotate(0.1)
	affine.Translate(100, 200)

	interp := NewSpanInterpolatorTrans(affine)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		interp.Begin(0, 0, 1000)
		for j := 0; j < 1000; j++ {
			interp.Next()
			interp.Coordinates()
		}
	}
}
