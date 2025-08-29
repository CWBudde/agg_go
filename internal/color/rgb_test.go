package color

import (
	"testing"
)

func TestRGB(t *testing.T) {
	// Test NewRGB
	rgb := NewRGB(0.5, 0.7, 0.3)
	if rgb.R != 0.5 || rgb.G != 0.7 || rgb.B != 0.3 {
		t.Errorf("NewRGB failed: got %v, want {0.5, 0.7, 0.3}", rgb)
	}

	// Test Clear
	rgb.Clear()
	if rgb.R != 0 || rgb.G != 0 || rgb.B != 0 {
		t.Errorf("Clear failed: got %v, want {0, 0, 0}", rgb)
	}

	// Test Scale
	rgb = NewRGB(0.5, 0.6, 0.4)
	scaled := rgb.Scale(2.0)
	if scaled.R != 1.0 || scaled.G != 1.2 || scaled.B != 0.8 {
		t.Errorf("Scale failed: got %v, want {1.0, 1.2, 0.8}", scaled)
	}

	// Test Add
	rgb1 := NewRGB(0.3, 0.4, 0.5)
	rgb2 := NewRGB(0.2, 0.3, 0.1)
	sum := rgb1.Add(rgb2)
	if sum.R != 0.5 || sum.G != 0.7 || sum.B != 0.6 {
		t.Errorf("Add failed: got %v, want {0.5, 0.7, 0.6}", sum)
	}

	// Test Gradient
	rgb1 = NewRGB(0.0, 0.0, 0.0)
	rgb2 = NewRGB(1.0, 1.0, 1.0)
	mid := rgb1.Gradient(rgb2, 0.5)
	if mid.R != 0.5 || mid.G != 0.5 || mid.B != 0.5 {
		t.Errorf("Gradient failed: got %v, want {0.5, 0.5, 0.5}", mid)
	}

	// Test ToRGBA
	rgb = NewRGB(0.3, 0.6, 0.9)
	rgba := rgb.ToRGBA()
	if rgba.R != 0.3 || rgba.G != 0.6 || rgba.B != 0.9 || rgba.A != 1.0 {
		t.Errorf("ToRGBA failed: got %v, want {0.3, 0.6, 0.9, 1.0}", rgba)
	}
}
