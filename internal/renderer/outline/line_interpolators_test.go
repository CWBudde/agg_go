// Package outline provides anti-aliased outline rendering functionality.
// This file contains comprehensive tests for line interpolator implementations.
package outline

import (
	"testing"

	"agg_go/internal/primitives"
)

// TestLineInterpolatorAABaseComprehensive tests the base functionality.
func TestLineInterpolatorAABaseComprehensive(t *testing.T) {
	renderer := NewMockOutlineRenderer(800, 600, 256)

	// Set up reasonable cover values
	for i := 0; i < len(renderer.coverValues); i++ {
		if i < 200 {
			renderer.coverValues[i] = 255 - i // Decreasing coverage
		} else {
			renderer.coverValues[i] = 0
		}
	}

	testCases := []struct {
		name     string
		x1, y1   int
		x2, y2   int
		len      int
		vertical bool
	}{
		{
			name:     "Horizontal line",
			x1:       100 << primitives.LineSubpixelShift,
			y1:       100 << primitives.LineSubpixelShift,
			x2:       200 << primitives.LineSubpixelShift,
			y2:       100 << primitives.LineSubpixelShift,
			len:      100,
			vertical: false,
		},
		{
			name:     "Vertical line",
			x1:       100 << primitives.LineSubpixelShift,
			y1:       100 << primitives.LineSubpixelShift,
			x2:       100 << primitives.LineSubpixelShift,
			y2:       200 << primitives.LineSubpixelShift,
			len:      100,
			vertical: true,
		},
		{
			name:     "Diagonal line",
			x1:       50 << primitives.LineSubpixelShift,
			y1:       50 << primitives.LineSubpixelShift,
			x2:       150 << primitives.LineSubpixelShift,
			y2:       80 << primitives.LineSubpixelShift,
			len:      100, // More horizontal than vertical
			vertical: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lp := primitives.NewLineParameters(tc.x1, tc.y1, tc.x2, tc.y2, tc.len)
			base := NewLineInterpolatorAABase(renderer, &lp)

			// Test basic properties
			if base.Vertical() != tc.vertical {
				t.Errorf("Expected vertical=%v, got %v", tc.vertical, base.Vertical())
			}

			if base.Count() <= 0 {
				t.Error("Count should be positive")
			}

			if base.Width() != renderer.SubpixelWidth() {
				t.Errorf("Expected width=%d, got %d", renderer.SubpixelWidth(), base.Width())
			}
		})
	}
}

// TestLineInterpolatorAA0Comprehensive tests LineInterpolatorAA0 functionality.
func TestLineInterpolatorAA0Comprehensive(t *testing.T) {
	renderer := NewMockOutlineRenderer(800, 600, 512)

	// Set up cover values with a gradient
	for i := 0; i < len(renderer.coverValues); i++ {
		if i < 300 {
			renderer.coverValues[i] = 255 - (i * 255 / 300)
		} else {
			renderer.coverValues[i] = 0
		}
	}

	testConfigs := []struct {
		name       string
		x1, y1     int
		x2, y2     int
		len        int
		expectVert bool
	}{
		{
			name:       "Short horizontal line",
			x1:         100 << primitives.LineSubpixelShift,
			y1:         100 << primitives.LineSubpixelShift,
			x2:         150 << primitives.LineSubpixelShift,
			y2:         100 << primitives.LineSubpixelShift,
			len:        50,
			expectVert: false,
		},
		{
			name:       "Short vertical line",
			x1:         100 << primitives.LineSubpixelShift,
			y1:         100 << primitives.LineSubpixelShift,
			x2:         100 << primitives.LineSubpixelShift,
			y2:         150 << primitives.LineSubpixelShift,
			len:        50,
			expectVert: true,
		},
		{
			name:       "Diagonal line short",
			x1:         50 << primitives.LineSubpixelShift,
			y1:         50 << primitives.LineSubpixelShift,
			x2:         100 << primitives.LineSubpixelShift,
			y2:         80 << primitives.LineSubpixelShift,
			len:        58, // approx
			expectVert: false,
		},
	}

	for _, config := range testConfigs {
		t.Run(config.name, func(t *testing.T) {
			lp := primitives.NewLineParameters(config.x1, config.y1, config.x2, config.y2, config.len)
			li := NewLineInterpolatorAA0(renderer, &lp)

			spanCallsBefore := len(renderer.vspanCalls) + len(renderer.hspanCalls)

			// Process all steps
			stepCount := 0
			maxSteps := li.Count() + 5 // Safety margin

			if li.Vertical() {
				for li.StepVer() && stepCount < maxSteps {
					stepCount++
				}
			} else {
				for li.StepHor() && stepCount < maxSteps {
					stepCount++
				}
			}

			spanCallsAfter := len(renderer.vspanCalls) + len(renderer.hspanCalls)

			if spanCallsAfter <= spanCallsBefore {
				t.Error("AA0 interpolator should have made span calls")
			}

			if stepCount == 0 {
				t.Error("Should have processed at least one step")
			}
		})
	}
}

// TestLineInterpolatorAA1Comprehensive tests LineInterpolatorAA1 functionality.
func TestLineInterpolatorAA1Comprehensive(t *testing.T) {
	renderer := NewMockOutlineRenderer(800, 600, 512)

	// Set up cover values
	for i := 0; i < len(renderer.coverValues); i++ {
		if i < 400 {
			renderer.coverValues[i] = 200 - (i * 200 / 400)
		} else {
			renderer.coverValues[i] = 0
		}
	}

	testConfigs := []struct {
		name   string
		x1, y1 int
		x2, y2 int
		sx, sy int // start cap coordinates
		len    int
	}{
		{
			name: "Horizontal with start cap",
			x1:   100 << primitives.LineSubpixelShift,
			y1:   100 << primitives.LineSubpixelShift,
			x2:   200 << primitives.LineSubpixelShift,
			y2:   100 << primitives.LineSubpixelShift,
			sx:   90 << primitives.LineSubpixelShift,
			sy:   100 << primitives.LineSubpixelShift,
			len:  100,
		},
		{
			name: "Vertical with start cap",
			x1:   100 << primitives.LineSubpixelShift,
			y1:   100 << primitives.LineSubpixelShift,
			x2:   100 << primitives.LineSubpixelShift,
			y2:   200 << primitives.LineSubpixelShift,
			sx:   100 << primitives.LineSubpixelShift,
			sy:   90 << primitives.LineSubpixelShift,
			len:  100,
		},
		{
			name: "Diagonal with start cap",
			x1:   100 << primitives.LineSubpixelShift,
			y1:   100 << primitives.LineSubpixelShift,
			x2:   150 << primitives.LineSubpixelShift,
			y2:   130 << primitives.LineSubpixelShift,
			sx:   95 << primitives.LineSubpixelShift,
			sy:   95 << primitives.LineSubpixelShift,
			len:  58,
		},
	}

	for _, config := range testConfigs {
		t.Run(config.name, func(t *testing.T) {
			lp := primitives.NewLineParameters(config.x1, config.y1, config.x2, config.y2, config.len)
			li := NewLineInterpolatorAA1(renderer, &lp, config.sx, config.sy)

			spanCallsBefore := len(renderer.vspanCalls) + len(renderer.hspanCalls)

			// Process all steps
			stepCount := 0
			maxSteps := li.Count() + 10 // Safety margin for start cap processing

			if li.Vertical() {
				for li.StepVer() && stepCount < maxSteps {
					stepCount++
				}
			} else {
				for li.StepHor() && stepCount < maxSteps {
					stepCount++
				}
			}

			spanCallsAfter := len(renderer.vspanCalls) + len(renderer.hspanCalls)

			if spanCallsAfter <= spanCallsBefore {
				t.Error("AA1 interpolator should have made span calls")
			}
		})
	}
}

// TestLineInterpolatorAA2Comprehensive tests LineInterpolatorAA2 functionality.
func TestLineInterpolatorAA2Comprehensive(t *testing.T) {
	renderer := NewMockOutlineRenderer(800, 600, 512)

	// Set up cover values for end cap testing
	for i := 0; i < len(renderer.coverValues); i++ {
		if i < 300 {
			renderer.coverValues[i] = 150 - (i * 150 / 300)
		} else {
			renderer.coverValues[i] = 0
		}
	}

	testConfigs := []struct {
		name   string
		x1, y1 int
		x2, y2 int
		ex, ey int // end cap coordinates
		len    int
	}{
		{
			name: "Horizontal with end cap",
			x1:   100 << primitives.LineSubpixelShift,
			y1:   100 << primitives.LineSubpixelShift,
			x2:   200 << primitives.LineSubpixelShift,
			y2:   100 << primitives.LineSubpixelShift,
			ex:   210 << primitives.LineSubpixelShift,
			ey:   100 << primitives.LineSubpixelShift,
			len:  100,
		},
		{
			name: "Vertical with end cap",
			x1:   100 << primitives.LineSubpixelShift,
			y1:   100 << primitives.LineSubpixelShift,
			x2:   100 << primitives.LineSubpixelShift,
			y2:   200 << primitives.LineSubpixelShift,
			ex:   100 << primitives.LineSubpixelShift,
			ey:   210 << primitives.LineSubpixelShift,
			len:  100,
		},
		{
			name: "Diagonal with end cap",
			x1:   100 << primitives.LineSubpixelShift,
			y1:   100 << primitives.LineSubpixelShift,
			x2:   150 << primitives.LineSubpixelShift,
			y2:   130 << primitives.LineSubpixelShift,
			ex:   155 << primitives.LineSubpixelShift,
			ey:   135 << primitives.LineSubpixelShift,
			len:  58,
		},
	}

	for _, config := range testConfigs {
		t.Run(config.name, func(t *testing.T) {
			lp := primitives.NewLineParameters(config.x1, config.y1, config.x2, config.y2, config.len)
			li := NewLineInterpolatorAA2(renderer, &lp, config.ex, config.ey)

			// Process all steps
			stepCount := 0
			maxSteps := li.Count() + li.maxExtent + 10 // Account for extent adjustment

			continueProcessing := true
			if li.Vertical() {
				for continueProcessing && stepCount < maxSteps {
					continueProcessing = li.StepVer()
					stepCount++
				}
			} else {
				for continueProcessing && stepCount < maxSteps {
					continueProcessing = li.StepHor()
					stepCount++
				}
			}

			// AA2 may not always make span calls if end cap conditions aren't met
			if stepCount == 0 {
				t.Error("Should have processed at least one step")
			}
		})
	}
}

// TestLineInterpolatorAA3Comprehensive tests LineInterpolatorAA3 functionality.
func TestLineInterpolatorAA3Comprehensive(t *testing.T) {
	renderer := NewMockOutlineRenderer(800, 600, 512)

	// Set up cover values for both cap testing
	for i := 0; i < len(renderer.coverValues); i++ {
		if i < 200 {
			renderer.coverValues[i] = 100 - (i * 100 / 200)
		} else {
			renderer.coverValues[i] = 0
		}
	}

	testConfigs := []struct {
		name   string
		x1, y1 int
		x2, y2 int
		sx, sy int // start cap coordinates
		ex, ey int // end cap coordinates
		len    int
	}{
		{
			name: "Horizontal with both caps",
			x1:   100 << primitives.LineSubpixelShift,
			y1:   100 << primitives.LineSubpixelShift,
			x2:   200 << primitives.LineSubpixelShift,
			y2:   100 << primitives.LineSubpixelShift,
			sx:   90 << primitives.LineSubpixelShift,
			sy:   100 << primitives.LineSubpixelShift,
			ex:   210 << primitives.LineSubpixelShift,
			ey:   100 << primitives.LineSubpixelShift,
			len:  100,
		},
		{
			name: "Vertical with both caps",
			x1:   100 << primitives.LineSubpixelShift,
			y1:   100 << primitives.LineSubpixelShift,
			x2:   100 << primitives.LineSubpixelShift,
			y2:   200 << primitives.LineSubpixelShift,
			sx:   100 << primitives.LineSubpixelShift,
			sy:   90 << primitives.LineSubpixelShift,
			ex:   100 << primitives.LineSubpixelShift,
			ey:   210 << primitives.LineSubpixelShift,
			len:  100,
		},
		{
			name: "Diagonal with both caps",
			x1:   100 << primitives.LineSubpixelShift,
			y1:   100 << primitives.LineSubpixelShift,
			x2:   140 << primitives.LineSubpixelShift,
			y2:   120 << primitives.LineSubpixelShift,
			sx:   95 << primitives.LineSubpixelShift,
			sy:   98 << primitives.LineSubpixelShift,
			ex:   145 << primitives.LineSubpixelShift,
			ey:   122 << primitives.LineSubpixelShift,
			len:  45,
		},
	}

	for _, config := range testConfigs {
		t.Run(config.name, func(t *testing.T) {
			lp := primitives.NewLineParameters(config.x1, config.y1, config.x2, config.y2, config.len)
			li := NewLineInterpolatorAA3(renderer, &lp, config.sx, config.sy, config.ex, config.ey)

			// Process all steps
			stepCount := 0
			maxSteps := li.Count() + li.maxExtent + 15 // Account for both extent adjustments

			continueProcessing := true
			if li.Vertical() {
				for continueProcessing && stepCount < maxSteps {
					continueProcessing = li.StepVer()
					stepCount++
				}
			} else {
				for continueProcessing && stepCount < maxSteps {
					continueProcessing = li.StepHor()
					stepCount++
				}
			}

			// AA3 may not always make span calls if cap conditions aren't met
			if stepCount == 0 {
				t.Error("Should have processed at least one step")
			}
		})
	}
}

// TestLineInterpolatorEdgeCases tests edge cases for line interpolators.
func TestLineInterpolatorEdgeCases(t *testing.T) {
	renderer := NewMockOutlineRenderer(800, 600, 256)

	// Set minimal cover values
	for i := 0; i < 100; i++ {
		renderer.coverValues[i] = 50
	}

	t.Run("Zero length line", func(t *testing.T) {
		x := 100 << primitives.LineSubpixelShift
		y := 100 << primitives.LineSubpixelShift
		lp := primitives.NewLineParameters(x, y, x, y, 0)

		// Should not panic
		li0 := NewLineInterpolatorAA0(renderer, &lp)
		if li0.Count() < 0 {
			t.Error("Count should not be negative")
		}

		li1 := NewLineInterpolatorAA1(renderer, &lp, x-10, y-10)
		if li1.Count() < 0 {
			t.Error("Count should not be negative")
		}

		li2 := NewLineInterpolatorAA2(renderer, &lp, x+10, y+10)
		if li2.Count() < 0 {
			t.Error("Count should not be negative")
		}

		li3 := NewLineInterpolatorAA3(renderer, &lp, x-10, y-10, x+10, y+10)
		if li3.Count() < 0 {
			t.Error("Count should not be negative")
		}
	})

	t.Run("Single pixel line", func(t *testing.T) {
		x1 := 100 << primitives.LineSubpixelShift
		y1 := 100 << primitives.LineSubpixelShift
		x2 := (100 + 1) << primitives.LineSubpixelShift
		y2 := 100 << primitives.LineSubpixelShift
		lp := primitives.NewLineParameters(x1, y1, x2, y2, 1)

		li0 := NewLineInterpolatorAA0(renderer, &lp)
		// Process single step if available
		if li0.Count() > 0 {
			if !li0.Vertical() {
				li0.StepHor()
			} else {
				li0.StepVer()
			}
		}

		li1 := NewLineInterpolatorAA1(renderer, &lp, x1-5, y1)
		if li1.Count() > 0 {
			if !li1.Vertical() {
				li1.StepHor()
			} else {
				li1.StepVer()
			}
		}
	})

	t.Run("Large coordinate values", func(t *testing.T) {
		large := 10000
		x1 := large << primitives.LineSubpixelShift
		y1 := large << primitives.LineSubpixelShift
		x2 := (large + 100) << primitives.LineSubpixelShift
		y2 := (large + 100) << primitives.LineSubpixelShift
		lp := primitives.NewLineParameters(x1, y1, x2, y2, 141)

		// Should not panic with large coordinates
		li0 := NewLineInterpolatorAA0(renderer, &lp)
		if li0.Count() > 0 {
			// Process a few steps
			for i := 0; i < 3 && ((!li0.Vertical() && li0.StepHor()) || (li0.Vertical() && li0.StepVer())); i++ {
			}
		}
	})

	t.Run("Negative coordinates", func(t *testing.T) {
		x1 := -100 << primitives.LineSubpixelShift
		y1 := -100 << primitives.LineSubpixelShift
		x2 := -50 << primitives.LineSubpixelShift
		y2 := -80 << primitives.LineSubpixelShift
		lp := primitives.NewLineParameters(x1, y1, x2, y2, 54)

		// Should handle negative coordinates gracefully
		li0 := NewLineInterpolatorAA0(renderer, &lp)
		if li0.Count() > 0 {
			if !li0.Vertical() {
				li0.StepHor()
			} else {
				li0.StepVer()
			}
		}

		li3 := NewLineInterpolatorAA3(renderer, &lp, x1-10, y1-10, x2+10, y2+10)
		if li3.Count() > 0 {
			if !li3.Vertical() {
				li3.StepHor()
			} else {
				li3.StepVer()
			}
		}
	})
}

// TestDistanceInterpolatorInterface tests the interface compliance.
func TestDistanceInterpolatorInterface(t *testing.T) {
	// Test that our distance interpolators implement the interface correctly
	var di DistanceInterpolatorInterface

	di = NewDistanceInterpolator1(0, 0, 100, 100, 50, 50)
	_ = di.Dist()
	di.IncXWithDY(1)
	di.DecXWithDY(1)
	di.IncYWithDX(1)
	di.DecYWithDX(1)

	// DistanceInterpolator2 should also work as the interface
	di2 := NewDistanceInterpolator2Start(0, 0, 100, 100, 10, 20, 50, 50)
	di = di2
	_ = di.Dist()
	di.IncXWithDY(1)
	di.DecXWithDY(1)
	di.IncYWithDX(1)
	di.DecYWithDX(1)

	// DistanceInterpolator3 should also work as the interface
	di3 := NewDistanceInterpolator3(0, 0, 100, 100, 10, 20, 90, 80, 50, 50)
	di = di3
	_ = di.Dist()
	di.IncXWithDY(1)
	di.DecXWithDY(1)
	di.IncYWithDX(1)
	di.DecYWithDX(1)
}

// TestMaxHalfWidthConstant tests the MaxHalfWidth constant usage.
func TestMaxHalfWidthConstant(t *testing.T) {
	if MaxHalfWidth != 64 {
		t.Errorf("Expected MaxHalfWidth to be 64, got %d", MaxHalfWidth)
	}

	// Test that interpolators work with maximum width
	renderer := NewMockOutlineRenderer(800, 600, MaxHalfWidth*8)

	// Fill cover values for high width
	for i := 0; i < len(renderer.coverValues); i++ {
		renderer.coverValues[i] = 100
	}

	x1 := 100 << primitives.LineSubpixelShift
	y1 := 100 << primitives.LineSubpixelShift
	x2 := 200 << primitives.LineSubpixelShift
	y2 := 100 << primitives.LineSubpixelShift
	lp := primitives.NewLineParameters(x1, y1, x2, y2, 100)

	// Should handle maximum width gracefully
	li := NewLineInterpolatorAA0(renderer, &lp)
	if li.Count() > 0 {
		li.StepHor()
	}
}
