// Package outline provides anti-aliased outline rendering functionality.
// This file contains comprehensive tests for line profile AA implementation.
package outline

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

// TestGammaFunction is a test implementation of GammaFunction.
type TestGammaFunction struct {
	power float64
}

func (tgf *TestGammaFunction) Call(x float64) float64 {
	return math.Pow(x, tgf.power)
}

// LinearTestGamma implements a linear gamma function for testing.
type LinearTestGamma struct{}

func (ltg *LinearTestGamma) Call(x float64) float64 {
	return x
}

// InverseTestGamma implements an inverse gamma function for testing.
type InverseTestGamma struct{}

func (itg *InverseTestGamma) Call(x float64) float64 {
	return 1.0 - x
}

// TestLineProfileAABasicCreation tests basic profile creation.
func TestLineProfileAABasicCreation(t *testing.T) {
	t.Run("Default constructor", func(t *testing.T) {
		profile := NewLineProfileAA()

		if profile == nil {
			t.Fatal("NewLineProfileAA should not return nil")
		}

		if profile.GetMinWidth() != 1.0 {
			t.Errorf("Expected default min width 1.0, got %f", profile.GetMinWidth())
		}

		if profile.GetSmootherWidth() != 1.0 {
			t.Errorf("Expected default smoother width 1.0, got %f", profile.GetSmootherWidth())
		}

		if profile.SubpixelWidth() != 0 {
			t.Error("Expected initial subpixel width to be 0")
		}

		if profile.ProfileSize() != 0 {
			t.Error("Expected initial profile size to be 0")
		}
	})

	t.Run("Constructor with gamma function", func(t *testing.T) {
		gamma := &LinearTestGamma{}
		width := 2.5

		profile := NewLineProfileAAWithGamma(width, gamma)

		if profile == nil {
			t.Fatal("NewLineProfileAAWithGamma should not return nil")
		}

		if profile.SubpixelWidth() == 0 {
			t.Error("Profile should have non-zero subpixel width after construction with width")
		}

		if profile.ProfileSize() == 0 {
			t.Error("Profile should have non-zero size after construction with width")
		}
	})
}

// TestLineProfileAAWidthSetting tests width configuration.
func TestLineProfileAAWidthSetting(t *testing.T) {
	profile := NewLineProfileAA()

	testCases := []struct {
		name        string
		width       float64
		expectSize  bool
		expectWidth bool
	}{
		{
			name:        "Positive width",
			width:       2.0,
			expectSize:  true,
			expectWidth: true,
		},
		{
			name:        "Small positive width",
			width:       0.5,
			expectSize:  true,
			expectWidth: true,
		},
		{
			name:        "Very small width",
			width:       0.1,
			expectSize:  true,
			expectWidth: true,
		},
		{
			name:        "Zero width",
			width:       0.0,
			expectSize:  true,
			expectWidth: true,
		},
		{
			name:        "Negative width",
			width:       -1.0,
			expectSize:  true,
			expectWidth: true,
		},
		{
			name:        "Large width",
			width:       100.0,
			expectSize:  true,
			expectWidth: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			profile.Width(tc.width)

			if tc.expectSize && profile.ProfileSize() == 0 {
				t.Error("Profile size should be non-zero after setting width")
			}

			if tc.expectWidth && profile.SubpixelWidth() == 0 {
				t.Error("Subpixel width should be non-zero after setting width")
			}
		})
	}
}

// TestLineProfileAAMinAndSmootherWidth tests min and smoother width setters.
func TestLineProfileAAMinAndSmootherWidth(t *testing.T) {
	profile := NewLineProfileAA()

	t.Run("MinWidth setter", func(t *testing.T) {
		testValues := []float64{0.5, 1.0, 2.0, 5.0, 0.1, 10.0}

		for _, val := range testValues {
			profile.MinWidth(val)
			if profile.GetMinWidth() != val {
				t.Errorf("Expected min width %f, got %f", val, profile.GetMinWidth())
			}
		}
	})

	t.Run("SmootherWidth setter", func(t *testing.T) {
		testValues := []float64{0.5, 1.0, 2.0, 5.0, 0.1, 10.0}

		for _, val := range testValues {
			profile.SmootherWidth(val)
			if profile.GetSmootherWidth() != val {
				t.Errorf("Expected smoother width %f, got %f", val, profile.GetSmootherWidth())
			}
		}
	})

	t.Run("Combined min and smoother width effects", func(t *testing.T) {
		profile.MinWidth(3.0)
		profile.SmootherWidth(1.5)
		profile.Width(2.0) // Less than min width

		// Should still create a valid profile
		if profile.ProfileSize() == 0 {
			t.Error("Profile should be created even with width less than min width")
		}
	})
}

// TestLineProfileAAGammaFunctions tests various gamma function implementations.
func TestLineProfileAAGammaFunctions(t *testing.T) {
	profile := NewLineProfileAA()

	testGammaFunctions := []struct {
		name  string
		gamma GammaFunction
	}{
		{
			name:  "Linear gamma",
			gamma: &LinearTestGamma{},
		},
		{
			name:  "Power gamma 0.5",
			gamma: &TestGammaFunction{power: 0.5},
		},
		{
			name:  "Power gamma 2.0",
			gamma: &TestGammaFunction{power: 2.0},
		},
		{
			name:  "Power gamma 2.2",
			gamma: &TestGammaFunction{power: 2.2},
		},
		{
			name:  "Inverse gamma",
			gamma: &InverseTestGamma{},
		},
	}

	for _, tc := range testGammaFunctions {
		t.Run(tc.name, func(t *testing.T) {
			profile.SetGamma(tc.gamma)
			profile.Width(2.0)

			// Should not panic and should produce valid values
			centerValue := profile.Value(0)
			if centerValue < 0 || centerValue > 255 {
				t.Errorf("Center value should be in range [0, 255], got %d", centerValue)
			}

			// Test edge values
			edgeValue := profile.Value(SubpixelScale)
			if edgeValue < 0 || edgeValue > 255 {
				t.Errorf("Edge value should be in range [0, 255], got %d", edgeValue)
			}
		})
	}
}

// TestLineProfileAAValueRetrieval tests value retrieval at different distances.
func TestLineProfileAAValueRetrieval(t *testing.T) {
	profile := NewLineProfileAA()
	profile.Width(2.0)

	t.Run("Center and nearby values", func(t *testing.T) {
		centerValue := profile.Value(0)
		if centerValue == 0 {
			t.Error("Center value should be non-zero for a 2.0 width line")
		}

		// Test values at small distances from center
		distances := []int{1, -1, 10, -10, 50, -50}
		for _, dist := range distances {
			value := profile.Value(dist)
			if value < 0 || value > 255 {
				t.Errorf("Value at distance %d should be in range [0, 255], got %d", dist, value)
			}
		}
	})

	t.Run("Boundary values", func(t *testing.T) {
		// Test values at profile boundaries
		profileSize := profile.ProfileSize()

		// Test near boundaries
		if profileSize > SubpixelScale*4 {
			minDist := -(SubpixelScale * 2)
			maxDist := profileSize - (SubpixelScale * 2) - 1

			minValue := profile.Value(minDist)
			maxValue := profile.Value(maxDist)

			if minValue < 0 || minValue > 255 {
				t.Errorf("Min boundary value should be in range [0, 255], got %d", minValue)
			}

			if maxValue < 0 || maxValue > 255 {
				t.Errorf("Max boundary value should be in range [0, 255], got %d", maxValue)
			}
		}
	})

	t.Run("Out of bounds values", func(t *testing.T) {
		profileSize := profile.ProfileSize()

		// Test values outside the profile range
		outOfBoundsDistances := []int{
			-(SubpixelScale * 3),
			profileSize,
			profileSize + 100,
			-1000,
			1000,
		}

		for _, dist := range outOfBoundsDistances {
			value := profile.Value(dist)
			if value != 0 {
				t.Errorf("Out of bounds value at distance %d should be 0, got %d", dist, value)
			}
		}
	})
}

// TestLineProfileAAConstants tests the constants used in the profile.
func TestLineProfileAAConstants(t *testing.T) {
	t.Run("Subpixel constants", func(t *testing.T) {
		if SubpixelShift != 8 {
			t.Errorf("Expected SubpixelShift to be 8, got %d", SubpixelShift)
		}

		if SubpixelScale != 256 {
			t.Errorf("Expected SubpixelScale to be 256, got %d", SubpixelScale)
		}

		if SubpixelMask != 255 {
			t.Errorf("Expected SubpixelMask to be 255, got %d", SubpixelMask)
		}
	})

	t.Run("AA constants", func(t *testing.T) {
		if AAShift != 8 {
			t.Errorf("Expected AAShift to be 8, got %d", AAShift)
		}

		if AAScale != 256 {
			t.Errorf("Expected AAScale to be 256, got %d", AAScale)
		}

		if AAMask != 255 {
			t.Errorf("Expected AAMask to be 255, got %d", AAMask)
		}
	})
}

// TestLineProfileAADifferentWidthScenarios tests various width scenarios.
func TestLineProfileAADifferentWidthScenarios(t *testing.T) {
	scenarios := []struct {
		name          string
		width         float64
		minWidth      float64
		smootherWidth float64
	}{
		{
			name:          "Width equals smoother width",
			width:         1.0,
			minWidth:      1.0,
			smootherWidth: 1.0,
		},
		{
			name:          "Width less than smoother width",
			width:         0.5,
			minWidth:      1.0,
			smootherWidth: 1.0,
		},
		{
			name:          "Width greater than smoother width",
			width:         3.0,
			minWidth:      1.0,
			smootherWidth: 1.0,
		},
		{
			name:          "Large smoother width",
			width:         2.0,
			minWidth:      1.0,
			smootherWidth: 3.0,
		},
		{
			name:          "Small min width",
			width:         0.8,
			minWidth:      0.5,
			smootherWidth: 1.0,
		},
		{
			name:          "Large min width",
			width:         1.5,
			minWidth:      3.0,
			smootherWidth: 1.0,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			profile := NewLineProfileAA()
			profile.MinWidth(scenario.minWidth)
			profile.SmootherWidth(scenario.smootherWidth)
			profile.Width(scenario.width)

			if profile.ProfileSize() == 0 {
				t.Error("Profile should have non-zero size")
			}

			if profile.SubpixelWidth() < 0 {
				t.Error("Subpixel width should not be negative")
			}

			// Test that center value is reasonable
			centerValue := profile.Value(0)
			if centerValue < 0 || centerValue > 255 {
				t.Errorf("Center value should be in range [0, 255], got %d", centerValue)
			}
		})
	}
}

// TestLineProfileAAProfileGeneration tests the internal profile generation.
func TestLineProfileAAProfileGeneration(t *testing.T) {
	profile := NewLineProfileAA()

	t.Run("Profile symmetry", func(t *testing.T) {
		profile.Width(3.0)

		// Test that profile is symmetric around center
		testDistances := []int{1, 2, 5, 10, 20, 50}

		for _, dist := range testDistances {
			posValue := profile.Value(dist)
			negValue := profile.Value(-dist)

			if posValue != negValue {
				t.Errorf("Profile should be symmetric: Value(%d)=%d != Value(%d)=%d",
					dist, posValue, -dist, negValue)
			}
		}
	})

	t.Run("Profile monotonicity", func(t *testing.T) {
		profile.Width(4.0)

		// Generally, values should decrease as we move away from center
		// (though this may not be strictly true for all gamma functions)
		centerValue := profile.Value(0)

		// Test a few points - values should generally decrease
		checkDistances := []int{SubpixelScale / 4, SubpixelScale / 2, SubpixelScale}

		lastValue := int(centerValue)
		for _, dist := range checkDistances {
			currentValue := int(profile.Value(dist))

			// Allow for some flexibility in monotonicity due to rounding and gamma
			if currentValue > lastValue+10 {
				t.Errorf("Profile values should generally decrease with distance: "+
					"Value(%d)=%d > previous value %d", dist, currentValue, lastValue)
			}

			lastValue = currentValue
		}
	})
}

// TestLineProfileAAEdgeCases tests edge cases and boundary conditions.
func TestLineProfileAAEdgeCases(t *testing.T) {
	t.Run("Zero and negative widths", func(t *testing.T) {
		profile := NewLineProfileAA()

		// Zero width
		profile.Width(0.0)
		if profile.ProfileSize() == 0 {
			t.Error("Profile should be created even with zero width")
		}

		// Negative width (should be treated as zero)
		profile.Width(-5.0)
		if profile.ProfileSize() == 0 {
			t.Error("Profile should be created even with negative width")
		}

		// Test that values are still reasonable
		centerValue := profile.Value(0)
		if centerValue < 0 || centerValue > 255 {
			t.Errorf("Center value should be valid even for edge case width, got %d", centerValue)
		}
	})

	t.Run("Very large widths", func(t *testing.T) {
		profile := NewLineProfileAA()
		profile.Width(1000.0)

		if profile.ProfileSize() == 0 {
			t.Error("Profile should handle large widths")
		}

		if profile.SubpixelWidth() <= 0 {
			t.Error("Subpixel width should be positive for large width")
		}

		// Should still produce valid values
		centerValue := profile.Value(0)
		if centerValue <= 0 {
			t.Error("Large width profile should produce non-zero center value")
		}
	})

	t.Run("Extreme min/smoother width combinations", func(t *testing.T) {
		profile := NewLineProfileAA()

		// Very small min width
		profile.MinWidth(0.01)
		profile.SmootherWidth(0.01)
		profile.Width(2.0)

		if profile.ProfileSize() == 0 {
			t.Error("Profile should handle very small min/smoother widths")
		}

		// Very large min width
		profile.MinWidth(1000.0)
		profile.SmootherWidth(100.0)
		profile.Width(50.0)

		if profile.ProfileSize() == 0 {
			t.Error("Profile should handle very large min width")
		}
	})
}

// TestLineProfileAAValueTypeConsistency tests ValueType consistency.
func TestLineProfileAAValueTypeConsistency(t *testing.T) {
	profile := NewLineProfileAA()
	profile.Width(2.0)

	// ValueType should be equivalent to basics.Int8u
	var valueType ValueType = 255
	var int8u basics.Int8u = 255

	if valueType != ValueType(int8u) {
		t.Error("ValueType should be compatible with basics.Int8u")
	}

	// Test that profile values fit in ValueType range
	testDistances := []int{-100, -10, -1, 0, 1, 10, 100, SubpixelScale}

	for _, dist := range testDistances {
		value := profile.Value(dist)
		if value > ValueType(255) {
			t.Errorf("Profile value %d at distance %d exceeds ValueType range", value, dist)
		}
	}
}

// BenchmarkLineProfileAACreation benchmarks profile creation.
func BenchmarkLineProfileAACreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		profile := NewLineProfileAA()
		profile.Width(2.0)
	}
}

// BenchmarkLineProfileAAValueAccess benchmarks value access.
func BenchmarkLineProfileAAValueAccess(b *testing.B) {
	profile := NewLineProfileAA()
	profile.Width(3.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = profile.Value(i % (SubpixelScale * 2))
	}
}

// BenchmarkLineProfileAAGammaApplication benchmarks gamma function application.
func BenchmarkLineProfileAAGammaApplication(b *testing.B) {
	gamma := &TestGammaFunction{power: 2.2}

	for i := 0; i < b.N; i++ {
		profile := NewLineProfileAAWithGamma(2.0, gamma)
		_ = profile.Value(0)
	}
}
