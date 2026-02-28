package agg2d

import (
	"math"
	"testing"

	"agg_go/internal/transform"
)

// TestColorGradient tests the Color.Gradient method for color interpolation
func TestColorGradient(t *testing.T) {
	// Test basic interpolation
	red := NewColorRGB(255, 0, 0)
	blue := NewColorRGB(0, 0, 255)

	// Test edge cases
	t.Run("k=0 returns first color", func(t *testing.T) {
		result := red.Gradient(blue, 0.0)
		if result != red {
			t.Errorf("Expected %v, got %v", red, result)
		}
	})

	t.Run("k=1 returns second color", func(t *testing.T) {
		result := red.Gradient(blue, 1.0)
		if result != blue {
			t.Errorf("Expected %v, got %v", blue, result)
		}
	})

	t.Run("k=0.5 returns midpoint", func(t *testing.T) {
		result := red.Gradient(blue, 0.5)
		expected := NewColorRGB(127, 0, 127) // Midpoint between red and blue
		// Allow for small rounding differences
		if abs(int(result.R())-int(expected.R())) > 1 ||
			abs(int(result.G())-int(expected.G())) > 1 ||
			abs(int(result.B())-int(expected.B())) > 1 {
			t.Errorf("Expected ~%v, got %v", expected, result)
		}
	})

	t.Run("k<0 clamps to first color", func(t *testing.T) {
		result := red.Gradient(blue, -0.5)
		if result != red {
			t.Errorf("Expected %v, got %v", red, result)
		}
	})

	t.Run("k>1 clamps to second color", func(t *testing.T) {
		result := red.Gradient(blue, 1.5)
		if result != blue {
			t.Errorf("Expected %v, got %v", blue, result)
		}
	})

	// Test with alpha channel
	t.Run("alpha interpolation", func(t *testing.T) {
		transparent := NewColor(255, 0, 0, 0) // Transparent red
		opaque := NewColor(255, 0, 0, 255)    // Opaque red

		result := transparent.Gradient(opaque, 0.5)
		expectedAlpha := uint8(127)
		if abs(int(result.A())-int(expectedAlpha)) > 1 {
			t.Errorf("Expected alpha ~%d, got %d", expectedAlpha, result.A())
		}
	})
}

// TestLinearGradientSetup tests linear gradient setup methods
func TestLinearGradientSetup(t *testing.T) {
	agg2d := NewAgg2D()

	// Setup a test rendering buffer
	width, height := 100, 100
	buf := make([]uint8, width*height*4)
	agg2d.Attach(buf, width, height, width*4)

	t.Run("FillLinearGradient basic setup", func(t *testing.T) {
		red := NewColorRGB(255, 0, 0)
		blue := NewColorRGB(0, 0, 255)

		// Set up horizontal gradient
		agg2d.FillLinearGradient(0, 0, 100, 0, red, blue, 1.0)

		// Check gradient flag is set
		if agg2d.fillGradientFlag != Linear {
			t.Errorf("Expected Linear gradient flag, got %v", agg2d.fillGradientFlag)
		}

		// Check gradient array has correct start and end colors
		if agg2d.fillGradient[0] != red {
			t.Errorf("Expected gradient start color %v, got %v", red, agg2d.fillGradient[0])
		}
		if agg2d.fillGradient[255] != blue {
			t.Errorf("Expected gradient end color %v, got %v", blue, agg2d.fillGradient[255])
		}

		// Check gradient distance
		expectedDistance := 100.0
		if math.Abs(agg2d.fillGradientD2-expectedDistance) > 0.01 {
			t.Errorf("Expected gradient distance %v, got %v", expectedDistance, agg2d.fillGradientD2)
		}
	})

	t.Run("LineLinearGradient basic setup", func(t *testing.T) {
		green := NewColorRGB(0, 255, 0)
		yellow := NewColorRGB(255, 255, 0)

		// Set up diagonal gradient
		agg2d.LineLinearGradient(0, 0, 50, 50, green, yellow, 1.0)

		// Check gradient flag is set
		if agg2d.lineGradientFlag != Linear {
			t.Errorf("Expected Linear gradient flag, got %v", agg2d.lineGradientFlag)
		}

		// Check gradient array has correct start and end colors
		if agg2d.lineGradient[0] != green {
			t.Errorf("Expected gradient start color %v, got %v", green, agg2d.lineGradient[0])
		}
		lineEndColor := agg2d.lineGradient[255]
		if abs(int(lineEndColor.R())-int(yellow.R())) > 1 ||
			abs(int(lineEndColor.G())-int(yellow.G())) > 1 ||
			abs(int(lineEndColor.B())-int(yellow.B())) > 1 ||
			abs(int(lineEndColor.A())-int(yellow.A())) > 1 {
			t.Errorf("Expected line gradient end color near %v, got %v", yellow, lineEndColor)
		}

		// Check gradient distance (diagonal: sqrt(50^2 + 50^2))
		expectedDistance := math.Sqrt(50*50 + 50*50)
		if math.Abs(agg2d.lineGradientD2-expectedDistance) > 0.01 {
			t.Errorf("Expected gradient distance %v, got %v", expectedDistance, agg2d.lineGradientD2)
		}
	})

	t.Run("Profile parameter effects", func(t *testing.T) {
		red := NewColorRGB(255, 0, 0)
		blue := NewColorRGB(0, 0, 255)

		// Test sharp profile (0.5)
		agg2d.FillLinearGradient(0, 0, 100, 0, red, blue, 0.5)

		// With profile 0.5, more of the gradient should be solid colors
		// startGradient = 128 - int(0.5 * 127) = 128 - 63 = 65
		// endGradient = 128 + int(0.5 * 127) = 128 + 63 = 191

		// Check that colors before index 65 are solid red
		if agg2d.fillGradient[64] != red {
			t.Errorf("Expected solid red at index 64, got %v", agg2d.fillGradient[64])
		}

		// Check that colors after index 191 are solid blue
		if agg2d.fillGradient[192] != blue {
			t.Errorf("Expected solid blue at index 192, got %v", agg2d.fillGradient[192])
		}
	})
}

// TestRadialGradientSetup tests radial gradient setup methods
func TestRadialGradientSetup(t *testing.T) {
	agg2d := NewAgg2D()

	// Setup a test rendering buffer
	width, height := 100, 100
	buf := make([]uint8, width*height*4)
	agg2d.Attach(buf, width, height, width*4)

	t.Run("FillRadialGradient basic setup", func(t *testing.T) {
		white := NewColorRGB(255, 255, 255)
		black := NewColorRGB(0, 0, 0)

		// Set up radial gradient at center
		agg2d.FillRadialGradient(50, 50, 25, white, black, 1.0)

		// Check gradient flag is set
		if agg2d.fillGradientFlag != Radial {
			t.Errorf("Expected Radial gradient flag, got %v", agg2d.fillGradientFlag)
		}

		// Check gradient array has correct start and end colors
		if agg2d.fillGradient[0] != white {
			t.Errorf("Expected gradient center color %v, got %v", white, agg2d.fillGradient[0])
		}
		if agg2d.fillGradient[255] != black {
			t.Errorf("Expected gradient edge color %v, got %v", black, agg2d.fillGradient[255])
		}

		// Check gradient radius
		expectedRadius := 25.0
		if math.Abs(agg2d.fillGradientD2-expectedRadius) > 0.01 {
			t.Errorf("Expected gradient radius %v, got %v", expectedRadius, agg2d.fillGradientD2)
		}
	})

	t.Run("LineRadialGradient basic setup", func(t *testing.T) {
		cyan := NewColorRGB(0, 255, 255)
		magenta := NewColorRGB(255, 0, 255)

		// Set up radial gradient
		agg2d.LineRadialGradient(25, 75, 40, cyan, magenta, 1.0)

		// Check gradient flag is set
		if agg2d.lineGradientFlag != Radial {
			t.Errorf("Expected Radial gradient flag, got %v", agg2d.lineGradientFlag)
		}

		// Check gradient array has correct start and end colors
		if agg2d.lineGradient[0] != cyan {
			t.Errorf("Expected gradient center color %v, got %v", cyan, agg2d.lineGradient[0])
		}
		lineEdgeColor := agg2d.lineGradient[255]
		if abs(int(lineEdgeColor.R())-int(magenta.R())) > 1 ||
			abs(int(lineEdgeColor.G())-int(magenta.G())) > 1 ||
			abs(int(lineEdgeColor.B())-int(magenta.B())) > 1 ||
			abs(int(lineEdgeColor.A())-int(magenta.A())) > 1 {
			t.Errorf("Expected gradient edge color near %v, got %v", magenta, lineEdgeColor)
		}
	})
}

// TestRadialGradientMultiStop tests multi-stop radial gradients
func TestRadialGradientMultiStop(t *testing.T) {
	agg2d := NewAgg2D()

	// Setup a test rendering buffer
	width, height := 100, 100
	buf := make([]uint8, width*height*4)
	agg2d.Attach(buf, width, height, width*4)

	t.Run("FillRadialGradientMultiStop", func(t *testing.T) {
		red := NewColorRGB(255, 0, 0)
		green := NewColorRGB(0, 255, 0)
		blue := NewColorRGB(0, 0, 255)

		agg2d.FillRadialGradientMultiStop(50, 50, 30, red, green, blue)

		// Check gradient flag is set
		if agg2d.fillGradientFlag != Radial {
			t.Errorf("Expected Radial gradient flag, got %v", agg2d.fillGradientFlag)
		}

		// Check color progression
		// Index 0: should be red (center)
		if agg2d.fillGradient[0] != red {
			t.Errorf("Expected red at center, got %v", agg2d.fillGradient[0])
		}

		// Index 127: should be close to green (middle)
		middleColor := agg2d.fillGradient[127]
		if abs(int(middleColor.G())-255) > 50 { // Allow some tolerance
			t.Errorf("Expected green-ish at middle, got %v", middleColor)
		}

		// Index 255: should be blue (edge)
		if agg2d.fillGradient[255] != blue {
			t.Errorf("Expected blue at edge, got %v", agg2d.fillGradient[255])
		}
	})

	t.Run("LineRadialGradientMultiStop", func(t *testing.T) {
		yellow := NewColorRGB(255, 255, 0)
		cyan := NewColorRGB(0, 255, 255)
		magenta := NewColorRGB(255, 0, 255)

		agg2d.LineRadialGradientMultiStop(25, 25, 50, yellow, cyan, magenta)

		// Check gradient flag is set
		if agg2d.lineGradientFlag != Radial {
			t.Errorf("Expected Radial gradient flag, got %v", agg2d.lineGradientFlag)
		}

		// Check start and end colors
		if agg2d.lineGradient[0] != yellow {
			t.Errorf("Expected yellow at center, got %v", agg2d.lineGradient[0])
		}
		if agg2d.lineGradient[255] != magenta {
			t.Errorf("Expected magenta at edge, got %v", agg2d.lineGradient[255])
		}
	})
}

// TestGradientPositionMethods tests position-only gradient setup methods
func TestGradientPositionMethods(t *testing.T) {
	agg2d := NewAgg2D()

	// Setup a test rendering buffer
	width, height := 100, 100
	buf := make([]uint8, width*height*4)
	agg2d.Attach(buf, width, height, width*4)

	// First setup a gradient with colors
	red := NewColorRGB(255, 0, 0)
	blue := NewColorRGB(0, 0, 255)
	agg2d.FillRadialGradient(25, 25, 20, red, blue, 1.0)

	// Store original gradient array
	originalGradient := agg2d.fillGradient

	t.Run("FillRadialGradientPos preserves colors", func(t *testing.T) {
		// Change position and radius
		agg2d.FillRadialGradientPos(75, 75, 35)

		// Check that gradient array is unchanged
		for i := 0; i < 256; i++ {
			if agg2d.fillGradient[i] != originalGradient[i] {
				t.Errorf("Gradient color changed at index %d", i)
				break
			}
		}

		// Check that radius was updated
		expectedRadius := 35.0
		if math.Abs(agg2d.fillGradientD2-expectedRadius) > 0.01 {
			t.Errorf("Expected gradient radius %v, got %v", expectedRadius, agg2d.fillGradientD2)
		}
	})

	t.Run("LineRadialGradientPos preserves colors", func(t *testing.T) {
		// Setup line gradient first
		green := NewColorRGB(0, 255, 0)
		yellow := NewColorRGB(255, 255, 0)
		agg2d.LineRadialGradient(10, 10, 15, green, yellow, 1.0)

		// Store original gradient array
		originalLineGradient := agg2d.lineGradient

		// Change position and radius
		agg2d.LineRadialGradientPos(90, 90, 45)

		// Check that gradient array is unchanged
		for i := 0; i < 256; i++ {
			if agg2d.lineGradient[i] != originalLineGradient[i] {
				t.Errorf("Line gradient color changed at index %d", i)
				break
			}
		}

		// Check that radius was updated
		expectedRadius := 45.0
		if math.Abs(agg2d.lineGradientD2-expectedRadius) > 0.01 {
			t.Errorf("Expected line gradient radius %v, got %v", expectedRadius, agg2d.lineGradientD2)
		}
	})
}

func TestRadialGradientUsesWorldToScreenTransform(t *testing.T) {
	agg2d := NewAgg2D()

	width, height := 100, 100
	buf := make([]uint8, width*height*4)
	agg2d.Attach(buf, width, height, width*4)

	agg2d.Scale(2.0, 3.0)
	agg2d.Translate(10.0, -4.0)

	centerX, centerY := 5.0, 7.0
	radius := 10.0
	expectedCenterX, expectedCenterY := centerX, centerY
	agg2d.WorldToScreen(&expectedCenterX, &expectedCenterY)
	expectedRadius := agg2d.WorldToScreenScalar(radius)

	agg2d.FillRadialGradient(centerX, centerY, radius, White, Black, 1.0)
	if math.Abs(agg2d.fillGradientD2-expectedRadius) > 1e-9 {
		t.Fatalf("fill radial radius mismatch: got %v, want %v", agg2d.fillGradientD2, expectedRadius)
	}
	if math.Abs(agg2d.fillGradientMatrix.TX+expectedCenterX) > 1e-9 {
		t.Fatalf("fill radial matrix TX mismatch: got %v, want %v", agg2d.fillGradientMatrix.TX, -expectedCenterX)
	}
	if math.Abs(agg2d.fillGradientMatrix.TY+expectedCenterY) > 1e-9 {
		t.Fatalf("fill radial matrix TY mismatch: got %v, want %v", agg2d.fillGradientMatrix.TY, -expectedCenterY)
	}

	agg2d.LineRadialGradient(centerX, centerY, radius, White, Black, 1.0)
	if math.Abs(agg2d.lineGradientD2-expectedRadius) > 1e-9 {
		t.Fatalf("line radial radius mismatch: got %v, want %v", agg2d.lineGradientD2, expectedRadius)
	}
	if math.Abs(agg2d.lineGradientMatrix.TX+expectedCenterX) > 1e-9 {
		t.Fatalf("line radial matrix TX mismatch: got %v, want %v", agg2d.lineGradientMatrix.TX, -expectedCenterX)
	}
	if math.Abs(agg2d.lineGradientMatrix.TY+expectedCenterY) > 1e-9 {
		t.Fatalf("line radial matrix TY mismatch: got %v, want %v", agg2d.lineGradientMatrix.TY, -expectedCenterY)
	}
}

func TestLinearGradientMatrixUsesCurrentTransform(t *testing.T) {
	agg2d := NewAgg2D()

	width, height := 100, 100
	buf := make([]uint8, width*height*4)
	agg2d.Attach(buf, width, height, width*4)

	agg2d.Scale(2.0, 3.0)
	agg2d.Translate(4.0, -6.0)

	x1, y1 := 3.0, 5.0
	x2, y2 := 13.0, 5.0

	agg2d.FillLinearGradient(x1, y1, x2, y2, White, Black, 1.0)
	expectedFill := transform.NewTransAffine()
	expectedFill.Rotate(math.Atan2(y2-y1, x2-x1))
	expectedFill.Translate(x1, y1)
	expectedFill.Multiply(agg2d.transform)
	expectedFill.Invert()
	assertAffineApproxEqual(t, agg2d.fillGradientMatrix, expectedFill, 1e-9)

	agg2d.LineLinearGradient(x1, y1, x2, y2, White, Black, 1.0)
	expectedLine := transform.NewTransAffine()
	expectedLine.Rotate(math.Atan2(y2-y1, x2-x1))
	expectedLine.Translate(x1, y1)
	expectedLine.Multiply(agg2d.transform)
	expectedLine.Invert()
	assertAffineApproxEqual(t, agg2d.lineGradientMatrix, expectedLine, 1e-9)
}

func assertAffineApproxEqual(t *testing.T, got, want *transform.TransAffine, eps float64) {
	t.Helper()
	if math.Abs(got.SX-want.SX) > eps ||
		math.Abs(got.SHY-want.SHY) > eps ||
		math.Abs(got.SHX-want.SHX) > eps ||
		math.Abs(got.SY-want.SY) > eps ||
		math.Abs(got.TX-want.TX) > eps ||
		math.Abs(got.TY-want.TY) > eps {
		t.Fatalf("affine mismatch: got %+v, want %+v", *got, *want)
	}
}

// Helper function for absolute value of integers
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// BenchmarkColorGradient benchmarks the color interpolation performance
func BenchmarkColorGradient(b *testing.B) {
	red := NewColorRGB(255, 0, 0)
	blue := NewColorRGB(0, 0, 255)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = red.Gradient(blue, 0.5)
	}
}

// BenchmarkLinearGradientSetup benchmarks linear gradient setup performance
func BenchmarkLinearGradientSetup(b *testing.B) {
	agg2d := NewAgg2D()
	width, height := 100, 100
	buf := make([]uint8, width*height*4)
	agg2d.Attach(buf, width, height, width*4)

	red := NewColorRGB(255, 0, 0)
	blue := NewColorRGB(0, 0, 255)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agg2d.FillLinearGradient(0, 0, 100, 100, red, blue, 1.0)
	}
}

// BenchmarkRadialGradientSetup benchmarks radial gradient setup performance
func BenchmarkRadialGradientSetup(b *testing.B) {
	agg2d := NewAgg2D()
	width, height := 100, 100
	buf := make([]uint8, width*height*4)
	agg2d.Attach(buf, width, height, width*4)

	white := NewColorRGB(255, 255, 255)
	black := NewColorRGB(0, 0, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agg2d.FillRadialGradient(50, 50, 25, white, black, 1.0)
	}
}
