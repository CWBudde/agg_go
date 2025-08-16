package color

import (
	"testing"

	"agg_go/internal/basics"
)

// Test Gray8 basic functionality
func TestGray8Basic(t *testing.T) {
	// Test constructor
	g := NewGray8[Linear](128)
	if g.V != 128 {
		t.Errorf("Expected V=128, got V=%d", g.V)
	}
	if g.A != Gray8BaseMask {
		t.Errorf("Expected A=%d, got A=%d", Gray8BaseMask, g.A)
	}

	// Test constructor with alpha
	g2 := NewGray8WithAlpha[Linear](64, 200)
	if g2.V != 64 || g2.A != 200 {
		t.Errorf("Expected V=64, A=200, got V=%d, A=%d", g2.V, g2.A)
	}
}

func TestGray8Methods(t *testing.T) {
	g := NewGray8WithAlpha[Linear](128, 200)

	// Test Clear
	g.Clear()
	if g.V != 0 || g.A != 0 {
		t.Errorf("Clear() failed: V=%d, A=%d", g.V, g.A)
	}

	// Test Transparent
	g = NewGray8WithAlpha[Linear](128, 200)
	g.Transparent()
	if g.V != 128 || g.A != 0 {
		t.Errorf("Transparent() failed: V=%d, A=%d", g.V, g.A)
	}

	// Test Opacity
	g = NewGray8WithAlpha[Linear](128, 200)
	g.Opacity(0.5)
	expected := basics.Int8u(0.5*float64(Gray8BaseMask) + 0.5)
	if g.A != expected {
		t.Errorf("Opacity(0.5) failed: expected A=%d, got A=%d", expected, g.A)
	}

	// Test IsTransparent
	g.A = 0
	if !g.IsTransparent() {
		t.Error("IsTransparent() should return true for A=0")
	}

	// Test IsOpaque
	g.A = Gray8BaseMask
	if !g.IsOpaque() {
		t.Error("IsOpaque() should return true for A=255")
	}
}

func TestGray8Arithmetic(t *testing.T) {
	// Test Multiply
	result := Gray8Multiply(128, 128)
	expected := basics.Int8u(64) // 128*128/256 ≈ 64
	if result != expected {
		t.Errorf("Gray8Multiply(128, 128) expected %d, got %d", expected, result)
	}

	// Test Lerp
	result = Gray8Lerp(0, 255, 128) // 50% interpolation
	expected = basics.Int8u(127)    // Should be approximately 127-128
	if result < 126 || result > 129 {
		t.Errorf("Gray8Lerp(0, 255, 128) expected ~127, got %d", result)
	}

	// Test Prelerp
	result = Gray8Prelerp(100, 50, 128)
	// p + q - multiply(p, a) = 100 + 50 - multiply(100, 128)
	mulResult := Gray8Multiply(100, 128)
	expected = 100 + 50 - mulResult
	if result != expected {
		t.Errorf("Gray8Prelerp(100, 50, 128) expected %d, got %d", expected, result)
	}
}

func TestGray8Premultiply(t *testing.T) {
	g := NewGray8WithAlpha[Linear](200, 128)
	originalV := g.V

	g.Premultiply()

	// V should be reduced by alpha
	expectedV := Gray8Multiply(originalV, 128)
	if g.V != expectedV {
		t.Errorf("Premultiply() expected V=%d, got V=%d", expectedV, g.V)
	}

	// Alpha should remain unchanged
	if g.A != 128 {
		t.Errorf("Premultiply() should not change alpha: got A=%d", g.A)
	}
}

func TestGray8Demultiply(t *testing.T) {
	g := NewGray8WithAlpha[Linear](100, 128)
	g.Premultiply()
	premultipliedV := g.V

	g.Demultiply()

	// V should be increased back (approximately)
	if g.V <= premultipliedV {
		t.Errorf("Demultiply() should increase V: premult=%d, demult=%d", premultipliedV, g.V)
	}
}

func TestGray8Gradient(t *testing.T) {
	g1 := NewGray8WithAlpha[Linear](0, 0)
	g2 := NewGray8WithAlpha[Linear](255, 255)

	// 50% interpolation
	result := g1.Gradient(g2, 0.5)

	// Should be approximately halfway
	if result.V < 125 || result.V > 130 {
		t.Errorf("Gradient V expected ~127, got %d", result.V)
	}
	if result.A < 125 || result.A > 130 {
		t.Errorf("Gradient A expected ~127, got %d", result.A)
	}
}

func TestGray8Add(t *testing.T) {
	g := NewGray8WithAlpha[Linear](100, 100)
	c := NewGray8WithAlpha[Linear](50, 50)

	g.Add(c, 255) // Full coverage

	// Values should be added (with clamping)
	expectedV := basics.Int8u(150)
	expectedA := basics.Int8u(150)

	if g.V != expectedV {
		t.Errorf("Add() V expected %d, got %d", expectedV, g.V)
	}
	if g.A != expectedA {
		t.Errorf("Add() A expected %d, got %d", expectedA, g.A)
	}
}

func TestGray8Conversion(t *testing.T) {
	// Test conversion from RGBA
	rgba := NewRGBA(0.5, 0.5, 0.5, 0.8)
	gray := ConvertGray8FromRGBA[Linear](rgba)

	// Should be approximately 50% gray with 80% alpha
	// Allow for rounding differences
	if gray.V < 127 || gray.V > 128 {
		t.Errorf("ConvertGray8FromRGBA V expected ~128, got %d", gray.V)
	}
	if gray.A < 203 || gray.A > 205 {
		t.Errorf("ConvertGray8FromRGBA A expected ~204, got %d", gray.A)
	}

	// Test conversion to RGBA
	rgba2 := gray.ConvertToRGBA()

	expectedR := float64(gray.V) / 255.0
	expectedG := expectedR
	expectedB := expectedR
	expectedAlpha := float64(gray.A) / 255.0

	tolerance := 0.01
	if abs(rgba2.R-expectedR) > tolerance {
		t.Errorf("ConvertToRGBA R expected %.3f, got %.3f", expectedR, rgba2.R)
	}
	if abs(rgba2.G-expectedG) > tolerance {
		t.Errorf("ConvertToRGBA G expected %.3f, got %.3f", expectedG, rgba2.G)
	}
	if abs(rgba2.B-expectedB) > tolerance {
		t.Errorf("ConvertToRGBA B expected %.3f, got %.3f", expectedB, rgba2.B)
	}
	if abs(rgba2.A-expectedAlpha) > tolerance {
		t.Errorf("ConvertToRGBA A expected %.3f, got %.3f", expectedAlpha, rgba2.A)
	}
}

func TestGray8Constants(t *testing.T) {
	if Gray8EmptyValue() != 0 {
		t.Errorf("Gray8EmptyValue() expected 0, got %d", Gray8EmptyValue())
	}

	if Gray8FullValue() != Gray8BaseMask {
		t.Errorf("Gray8FullValue() expected %d, got %d", Gray8BaseMask, Gray8FullValue())
	}
}

func TestGray16Basic(t *testing.T) {
	g := NewGray16[Linear](32768)
	if g.V != 32768 {
		t.Errorf("Expected V=32768, got V=%d", g.V)
	}
	if g.A != Gray16BaseMask {
		t.Errorf("Expected A=%d, got A=%d", Gray16BaseMask, g.A)
	}

	g.Clear()
	if g.V != 0 || g.A != 0 {
		t.Errorf("Clear() failed: V=%d, A=%d", g.V, g.A)
	}
}

func TestGray32Basic(t *testing.T) {
	g := NewGray32[Linear](0.5)
	if g.V != 0.5 {
		t.Errorf("Expected V=0.5, got V=%f", g.V)
	}
	if g.A != 1.0 {
		t.Errorf("Expected A=1.0, got A=%f", g.A)
	}

	g.Clear()
	if g.V != 0.0 || g.A != 0.0 {
		t.Errorf("Clear() failed: V=%f, A=%f", g.V, g.A)
	}
}

// Helper function for floating point comparison
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
