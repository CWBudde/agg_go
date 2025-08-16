package span

import (
	"testing"
)

func TestSolidSpanGenerator(t *testing.T) {
	color := "red"
	gen := NewSolidSpanGenerator(color)

	t.Run("basic generation", func(t *testing.T) {
		colors := make([]interface{}, 10)
		gen.Generate(colors, 0, 0, 10)

		for i, c := range colors {
			if c != color {
				t.Errorf("Expected %v at index %d, got %v", color, i, c)
			}
		}
	})

	t.Run("color getter", func(t *testing.T) {
		if gen.Color() != color {
			t.Errorf("Expected %v, got %v", color, gen.Color())
		}
	})

	t.Run("color setter", func(t *testing.T) {
		newColor := "blue"
		gen.SetColor(newColor)

		if gen.Color() != newColor {
			t.Errorf("Expected %v, got %v", newColor, gen.Color())
		}

		colors := make([]interface{}, 5)
		gen.Generate(colors, 0, 0, 5)

		for i, c := range colors {
			if c != newColor {
				t.Errorf("Expected %v at index %d, got %v", newColor, i, c)
			}
		}
	})

	t.Run("prepare does nothing", func(t *testing.T) {
		// Should not panic or error
		gen.Prepare()
	})

	t.Run("zero length generation", func(t *testing.T) {
		colors := make([]interface{}, 0)
		gen.Generate(colors, 0, 0, 0)
		// Should not panic
	})
}
