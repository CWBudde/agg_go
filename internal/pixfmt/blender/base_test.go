package blender

import (
	"math"
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
)

func TestBlenderRGBAGet(t *testing.T) {
	blender := BlenderRGBA[color.Linear, RGBAOrder]{}

	// Test basic Get functionality
	pixel := []basics.Int8u{128, 64, 192, 255} // RGBA: 50%, 25%, 75%, 100%
	cover := basics.Int8u(255)                 // Full coverage

	result := blender.Get(pixel, cover)

	expectedR := 128.0 / 255.0
	expectedG := 64.0 / 255.0
	expectedB := 192.0 / 255.0
	expectedA := 255.0 / 255.0

	if math.Abs(result.R-expectedR) > 0.01 {
		t.Errorf("Get R: expected %f, got %f", expectedR, result.R)
	}
	if math.Abs(result.G-expectedG) > 0.01 {
		t.Errorf("Get G: expected %f, got %f", expectedG, result.G)
	}
	if math.Abs(result.B-expectedB) > 0.01 {
		t.Errorf("Get B: expected %f, got %f", expectedB, result.B)
	}
	if math.Abs(result.A-expectedA) > 0.01 {
		t.Errorf("Get A: expected %f, got %f", expectedA, result.A)
	}
}

func TestBlenderRGBAGetWithCoverage(t *testing.T) {
	blender := BlenderRGBA[color.Linear, RGBAOrder]{}

	// Test Get with partial coverage
	pixel := []basics.Int8u{255, 255, 255, 255} // White pixel
	cover := basics.Int8u(128)                  // 50% coverage

	result := blender.Get(pixel, cover)

	// With 50% coverage, all components should be scaled down by 0.5
	expected := 0.5
	if math.Abs(result.R-expected) > 0.01 {
		t.Errorf("Get with coverage R: expected %f, got %f", expected, result.R)
	}
	if math.Abs(result.G-expected) > 0.01 {
		t.Errorf("Get with coverage G: expected %f, got %f", expected, result.G)
	}
	if math.Abs(result.B-expected) > 0.01 {
		t.Errorf("Get with coverage B: expected %f, got %f", expected, result.B)
	}
	if math.Abs(result.A-expected) > 0.01 {
		t.Errorf("Get with coverage A: expected %f, got %f", expected, result.A)
	}
}

func TestBlenderRGBAGetZeroCoverage(t *testing.T) {
	blender := BlenderRGBA[color.Linear, RGBAOrder]{}

	pixel := []basics.Int8u{255, 255, 255, 255}
	cover := basics.Int8u(0) // Zero coverage

	result := blender.Get(pixel, cover)

	// Zero coverage should return NoColor (transparent black)
	expected := color.NoColor()
	if result.R != expected.R || result.G != expected.G || result.B != expected.B || result.A != expected.A {
		t.Errorf("Get with zero coverage should return NoColor, got %+v", result)
	}
}

func TestBlenderRGBAGetRaw(t *testing.T) {
	blender := BlenderRGBA[color.Linear, RGBAOrder]{}

	pixel := []basics.Int8u{128, 64, 192, 255}

	r, g, b, a := blender.GetRaw(pixel)

	if r != 128 || g != 64 || b != 192 || a != 255 {
		t.Errorf("GetRaw: expected (128, 64, 192, 255), got (%d, %d, %d, %d)", r, g, b, a)
	}
}

func TestBlenderRGBASet(t *testing.T) {
	blender := BlenderRGBA[color.Linear, RGBAOrder]{}

	pixel := make([]basics.Int8u, 4)
	inputColor := color.RGBA{R: 0.5, G: 0.25, B: 0.75, A: 1.0}

	blender.Set(pixel, inputColor)

	expectedR := basics.Int8u(128) // 0.5 * 255 + 0.5 = 128
	expectedG := basics.Int8u(64)  // 0.25 * 255 + 0.5 = 64
	expectedB := basics.Int8u(191) // 0.75 * 255 + 0.5 = 191
	expectedA := basics.Int8u(255) // 1.0 * 255 + 0.5 = 255

	if pixel[0] != expectedR || pixel[1] != expectedG || pixel[2] != expectedB || pixel[3] != expectedA {
		t.Errorf("Set: expected (%d, %d, %d, %d), got (%d, %d, %d, %d)",
			expectedR, expectedG, expectedB, expectedA,
			pixel[0], pixel[1], pixel[2], pixel[3])
	}
}

func TestBlenderRGBASetRaw(t *testing.T) {
	blender := BlenderRGBA[color.Linear, RGBAOrder]{}

	pixel := make([]basics.Int8u, 4)
	blender.SetRaw(pixel, 128, 64, 192, 255)

	if pixel[0] != 128 || pixel[1] != 64 || pixel[2] != 192 || pixel[3] != 255 {
		t.Errorf("SetRaw: expected (128, 64, 192, 255), got (%d, %d, %d, %d)", pixel[0], pixel[1], pixel[2], pixel[3])
	}
}

func TestBlenderRGBARoundTrip(t *testing.T) {
	blender := BlenderRGBA[color.Linear, RGBAOrder]{}

	// Test that Set followed by Get preserves the color (within rounding error)
	originalColor := color.RGBA{R: 0.3, G: 0.6, B: 0.9, A: 0.8}
	pixel := make([]basics.Int8u, 4)

	blender.Set(pixel, originalColor)
	retrievedColor := blender.Get(pixel, 255)

	tolerance := 0.01 // Allow for floating point rounding errors
	if math.Abs(retrievedColor.R-originalColor.R) > tolerance ||
		math.Abs(retrievedColor.G-originalColor.G) > tolerance ||
		math.Abs(retrievedColor.B-originalColor.B) > tolerance ||
		math.Abs(retrievedColor.A-originalColor.A) > tolerance {
		t.Errorf("Round trip failed: original %+v, retrieved %+v", originalColor, retrievedColor)
	}
}

func TestBlenderRGBAColorOrders(t *testing.T) {
	testCases := []struct {
		name     string
		setFunc  func([]basics.Int8u, color.RGBA)
		expected color.ColorOrder
	}{
		{"RGBA", BlenderRGBA[color.Linear, RGBAOrder]{}.Set, color.OrderRGBA},
		{"ARGB", BlenderRGBA[color.Linear, ARGBOrder]{}.Set, color.OrderARGB},
		{"BGRA", BlenderRGBA[color.Linear, BGRAOrder]{}.Set, color.OrderBGRA},
		{"ABGR", BlenderRGBA[color.Linear, ABGROrder]{}.Set, color.OrderABGR},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pixel := make([]basics.Int8u, 4)
			testColor := color.RGBA{R: 0.2, G: 0.4, B: 0.6, A: 0.8}

			tc.setFunc(pixel, testColor)

			// Verify the pixel is stored in the correct order
			expectedR := basics.Int8u(51)  // 0.2 * 255 + 0.5 = 51
			expectedG := basics.Int8u(102) // 0.4 * 255 + 0.5 = 102
			expectedB := basics.Int8u(153) // 0.6 * 255 + 0.5 = 153
			expectedA := basics.Int8u(204) // 0.8 * 255 + 0.5 = 204

			if pixel[tc.expected.R] != expectedR {
				t.Errorf("%s order R component: expected %d at index %d, got %d", tc.name, expectedR, tc.expected.R, pixel[tc.expected.R])
			}
			if pixel[tc.expected.G] != expectedG {
				t.Errorf("%s order G component: expected %d at index %d, got %d", tc.name, expectedG, tc.expected.G, pixel[tc.expected.G])
			}
			if pixel[tc.expected.B] != expectedB {
				t.Errorf("%s order B component: expected %d at index %d, got %d", tc.name, expectedB, tc.expected.B, pixel[tc.expected.B])
			}
			if pixel[tc.expected.A] != expectedA {
				t.Errorf("%s order A component: expected %d at index %d, got %d", tc.name, expectedA, tc.expected.A, pixel[tc.expected.A])
			}
		})
	}
}

func TestBlenderRGBAPreGet(t *testing.T) {
	blender := BlenderRGBAPre[color.Linear, RGBAOrder]{}

	// Test premultiplied Get - create a premultiplied pixel
	originalColor := color.RGBA{R: 0.6, G: 0.4, B: 0.8, A: 0.5}
	premultColor := originalColor
	premultColor.Premultiply()

	pixel := []basics.Int8u{
		basics.Int8u(77),  // 0.3 * 255 + 0.5 = 77
		basics.Int8u(51),  // 0.2 * 255 + 0.5 = 51
		basics.Int8u(102), // 0.4 * 255 + 0.5 = 102
		basics.Int8u(128), // 0.5 * 255 + 0.5 = 128
	}

	result := blender.Get(pixel, 255)

	tolerance := 0.02 // Premultiplied operations have more rounding
	if math.Abs(result.R-originalColor.R) > tolerance ||
		math.Abs(result.G-originalColor.G) > tolerance ||
		math.Abs(result.B-originalColor.B) > tolerance ||
		math.Abs(result.A-originalColor.A) > tolerance {
		t.Errorf("BlenderRGBAPre Get: expected %+v, got %+v", originalColor, result)
	}
}

func TestBlenderRGBAPreSet(t *testing.T) {
	blender := BlenderRGBAPre[color.Linear, RGBAOrder]{}

	pixel := make([]basics.Int8u, 4)
	inputColor := color.RGBA{R: 0.6, G: 0.4, B: 0.8, A: 0.5}

	blender.Set(pixel, inputColor)

	// The Set method should premultiply before storing
	expectedPremult := inputColor
	expectedPremult.Premultiply()

	expectedR := basics.Int8u(77)  // 0.3 * 255 + 0.5 = 77  (0.6 * 0.5)
	expectedG := basics.Int8u(51)  // 0.2 * 255 + 0.5 = 51  (0.4 * 0.5)
	expectedB := basics.Int8u(102) // 0.4 * 255 + 0.5 = 102 (0.8 * 0.5)
	expectedA := basics.Int8u(128) // 0.5 * 255 + 0.5 = 128

	if pixel[0] != expectedR || pixel[1] != expectedG || pixel[2] != expectedB || pixel[3] != expectedA {
		t.Errorf("BlenderRGBAPre Set: expected (%d, %d, %d, %d), got (%d, %d, %d, %d)",
			expectedR, expectedG, expectedB, expectedA,
			pixel[0], pixel[1], pixel[2], pixel[3])
	}
}

func TestBlenderRGBAPreRoundTrip(t *testing.T) {
	blender := BlenderRGBAPre[color.Linear, RGBAOrder]{}

	originalColor := color.RGBA{R: 0.3, G: 0.6, B: 0.9, A: 0.7}
	pixel := make([]basics.Int8u, 4)

	blender.Set(pixel, originalColor)
	retrievedColor := blender.Get(pixel, 255)

	tolerance := 0.02 // Premultiplied operations have more rounding
	if math.Abs(retrievedColor.R-originalColor.R) > tolerance ||
		math.Abs(retrievedColor.G-originalColor.G) > tolerance ||
		math.Abs(retrievedColor.B-originalColor.B) > tolerance ||
		math.Abs(retrievedColor.A-originalColor.A) > tolerance {
		t.Errorf("BlenderRGBAPre round trip failed: original %+v, retrieved %+v", originalColor, retrievedColor)
	}
}

func TestBlenderRGBAPlainRoundTrip(t *testing.T) {
	blender := BlenderRGBAPlain[color.Linear, RGBAOrder]{}

	originalColor := color.RGBA{R: 0.3, G: 0.6, B: 0.9, A: 0.7}
	pixel := make([]basics.Int8u, 4)

	blender.Set(pixel, originalColor)
	retrievedColor := blender.Get(pixel, 255)

	tolerance := 0.01
	if math.Abs(retrievedColor.R-originalColor.R) > tolerance ||
		math.Abs(retrievedColor.G-originalColor.G) > tolerance ||
		math.Abs(retrievedColor.B-originalColor.B) > tolerance ||
		math.Abs(retrievedColor.A-originalColor.A) > tolerance {
		t.Errorf("BlenderRGBAPlain round trip failed: original %+v, retrieved %+v", originalColor, retrievedColor)
	}
}

