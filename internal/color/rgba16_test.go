package color

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/gamma"
)

func TestRGBA16ApplyGamma(t *testing.T) {
	lut := gamma.NewGammaLUT16WithGamma(2.0)

	c := NewRGBA16[Linear](32768, 16384, 49152, 65535)
	original := c

	c.ApplyGammaDir(lut)

	if c.A != original.A {
		t.Errorf("ApplyGammaDir changed alpha: got %d, expected %d", c.A, original.A)
	}
	if c.R == original.R && c.G == original.G && c.B == original.B {
		t.Error("ApplyGammaDir should change RGB values")
	}

	c.ApplyGammaInv(lut)

	tolerance := basics.Int16u(1000)
	if abs16u(c.R, original.R) > tolerance ||
		abs16u(c.G, original.G) > tolerance ||
		abs16u(c.B, original.B) > tolerance {
		t.Errorf("ApplyGammaInv didn't restore original within tolerance: got (%d,%d,%d), expected (%d,%d,%d)",
			c.R, c.G, c.B, original.R, original.G, original.B)
	}
}
