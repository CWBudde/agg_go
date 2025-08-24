package span

import (
	"testing"

	"agg_go/internal/color"
)

// TestColorForSpan is a simple color type for testing span converters
type TestColorForSpan struct {
	R, G, B, A uint8
	Name       string
}

func TestSpanConverter_Basic(t *testing.T) {
	// Create a simple test with the new generic approach
	testColor := TestColorForSpan{R: 255, G: 128, B: 64, A: 255, Name: "orange"}

	// Create generator and converter
	gen := NewSolidSpanGenerator[TestColorForSpan](testColor)
	alphaConv := NewAlphaConverterSpan[TestColorForSpan](0.5)

	// Create span converter
	spanConv := NewSpanConverter(gen, alphaConv)

	// Test preparation
	spanConv.Prepare()

	// Test generation
	colors := make([]TestColorForSpan, 3)
	spanConv.Generate(colors, 10, 20, 3)

	// Verify all colors were set
	for i, c := range colors {
		if c.Name != testColor.Name {
			t.Errorf("Expected color[%d].Name to be '%s', got '%s'", i, testColor.Name, c.Name)
		}
		if c.R != testColor.R || c.G != testColor.G || c.B != testColor.B {
			t.Errorf("Expected color[%d] RGB to be (%d,%d,%d), got (%d,%d,%d)",
				i, testColor.R, testColor.G, testColor.B, c.R, c.G, c.B)
		}
	}
}

func TestBrightnessAlphaConverter(t *testing.T) {
	// Test the brightness-alpha converter with RGBA8 colors
	brightColor := color.NewRGBA8[color.SRGB](255, 255, 255, 255) // Bright white
	darkColor := color.NewRGBA8[color.SRGB](50, 50, 50, 255)      // Dark gray

	// Create brightness converter with linear alpha mapping
	alphaArray := make([]uint8, 768)
	for i := 0; i < 768; i++ {
		alphaArray[i] = uint8(i * 255 / 767) // Linear mapping
	}

	converter := NewBrightnessAlphaConverter[color.RGBA8[color.SRGB]](alphaArray)

	// Test with bright color
	brightColors := []color.RGBA8[color.SRGB]{brightColor}
	converter.Generate(brightColors, 0, 0, 1)

	// Bright color should have high alpha
	if brightColors[0].A < 200 {
		t.Errorf("Expected bright color to have high alpha, got %d", brightColors[0].A)
	}

	// Test with dark color
	darkColors := []color.RGBA8[color.SRGB]{darkColor}
	converter.Generate(darkColors, 0, 0, 1)

	// Dark color should have low alpha
	if darkColors[0].A > 100 {
		t.Errorf("Expected dark color to have low alpha, got %d", darkColors[0].A)
	}
}

func TestAlphaConverter(t *testing.T) {
	// Test the alpha converter
	originalColor := color.NewRGBA8[color.SRGB](255, 128, 64, 200)

	// Create alpha converter with 0.5 alpha
	converter := NewAlphaConverterSpan[color.RGBA8[color.SRGB]](0.5)

	// Apply conversion
	colors := []color.RGBA8[color.SRGB]{originalColor}
	converter.Generate(colors, 0, 0, 1)

	// Check that alpha was reduced
	expectedAlpha := uint8(float64(200) * 0.5)
	if colors[0].A != expectedAlpha {
		t.Errorf("Expected alpha to be %d, got %d", expectedAlpha, colors[0].A)
	}

	// RGB should remain unchanged
	if colors[0].R != 255 || colors[0].G != 128 || colors[0].B != 64 {
		t.Errorf("Expected RGB to remain unchanged, got (%d,%d,%d)", colors[0].R, colors[0].G, colors[0].B)
	}
}
