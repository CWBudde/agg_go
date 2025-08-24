// Package outline provides anti-aliased outline rendering functionality.
// This file contains comprehensive tests for distance interpolator implementations.
package outline

import (
	"testing"

	"agg_go/internal/primitives"
)

// TestDistanceInterpolator0Comprehensive tests all DistanceInterpolator0 functionality.
func TestDistanceInterpolator0Comprehensive(t *testing.T) {
	testCases := []struct {
		name       string
		x1, y1     int
		x2, y2     int
		x, y       int
		expectDist bool // whether we expect a distance calculation
	}{
		{
			name: "Basic horizontal line",
			x1:   0, y1: 0,
			x2: 100, y2: 0,
			x: 50, y: 25,
			expectDist: true,
		},
		{
			name: "Basic vertical line",
			x1:   0, y1: 0,
			x2: 0, y2: 100,
			x: 25, y: 50,
			expectDist: true,
		},
		{
			name: "Diagonal line",
			x1:   0, y1: 0,
			x2: 100, y2: 100,
			x: 25, y: 75,
			expectDist: true,
		},
		{
			name: "Zero length line",
			x1:   50, y1: 50,
			x2: 50, y2: 50,
			x: 50, y: 50,
			expectDist: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			di := NewDistanceInterpolator0(tc.x1, tc.y1, tc.x2, tc.y2, tc.x, tc.y)

			// Test IncX multiple times
			distChanged := false
			for i := 0; i < 5; i++ {
				prevDist := di.Dist()
				di.IncX()
				newDist := di.Dist()

				// Check if distance changed
				if prevDist != newDist {
					distChanged = true
				}
			}

			// Verify the distance method works - should change if dy != 0
			if tc.y1 != tc.y2 && !distChanged {
				t.Error("Distance should have changed after multiple IncX() calls when dy != 0")
			}
		})
	}
}

// TestDistanceInterpolator00Comprehensive tests all DistanceInterpolator00 functionality.
func TestDistanceInterpolator00Comprehensive(t *testing.T) {
	testCases := []struct {
		name   string
		xc, yc int // center
		x1, y1 int // first point
		x2, y2 int // second point
		x, y   int // test point
	}{
		{
			name: "Basic triangle setup",
			xc:   50, yc: 50,
			x1: 0, y1: 0,
			x2: 100, y2: 100,
			x: 25, y: 25,
		},
		{
			name: "Right angle setup",
			xc:   0, yc: 0,
			x1: 100, y1: 0,
			x2: 0, y2: 100,
			x: 50, y: 50,
		},
		{
			name: "Collinear points",
			xc:   50, yc: 50,
			x1: 0, y1: 0,
			x2: 100, y2: 100,
			x: 50, y: 50,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			di := NewDistanceInterpolator00(tc.xc, tc.yc, tc.x1, tc.y1, tc.x2, tc.y2, tc.x, tc.y)

			initialDist1 := di.Dist1()
			initialDist2 := di.Dist2()

			// Test multiple IncX calls
			for i := 0; i < 3; i++ {
				di.IncX()
			}

			finalDist1 := di.Dist1()
			finalDist2 := di.Dist2()

			// Both distances should be accessible
			_ = finalDist1
			_ = finalDist2

			// If dy values are non-zero, distances should change
			if tc.y1 != tc.yc && finalDist1 == initialDist1 {
				t.Error("Dist1 should change after IncX() when dy1 != 0")
			}
			if tc.y2 != tc.yc && finalDist2 == initialDist2 {
				t.Error("Dist2 should change after IncX() when dy2 != 0")
			}
		})
	}
}

// TestDistanceInterpolator1Comprehensive tests all DistanceInterpolator1 functionality.
func TestDistanceInterpolator1Comprehensive(t *testing.T) {
	di := NewDistanceInterpolator1(0, 0, 100, 100, 50, 50)

	t.Run("Basic operations", func(t *testing.T) {
		initialDist := di.Dist()

		// Test DX and DY accessors
		dx := di.DX()
		dy := di.DY()

		if dx == 0 && dy == 0 {
			t.Error("DX and DY should not both be zero for this test case")
		}

		// Test all increment/decrement operations
		di.IncX()

		di.DecX()
		afterDecX := di.Dist()

		if afterDecX != initialDist {
			t.Error("IncX followed by DecX should return to initial distance")
		}

		di.IncY()

		di.DecY()
		afterDecY := di.Dist()

		if afterDecY != initialDist {
			t.Error("IncY followed by DecY should return to initial distance")
		}
	})

	t.Run("WithDelta operations", func(t *testing.T) {
		// Test IncXWithDY with different dy values
		testDYValues := []int{-1, 0, 1, 5, -3}

		for _, dyVal := range testDYValues {
			di = NewDistanceInterpolator1(0, 0, 100, 100, 50, 50) // Reset

			di.IncXWithDY(dyVal)
			distAfterInc := di.Dist()

			di.DecXWithDY(dyVal)
			distAfterDec := di.Dist()

			// The exact values depend on internal calculations, but operations should be consistent
			_ = distAfterInc
			_ = distAfterDec
		}

		// Test IncYWithDX with different dx values
		testDXValues := []int{-1, 0, 1, 5, -3}

		for _, dxVal := range testDXValues {
			di = NewDistanceInterpolator1(0, 0, 100, 100, 50, 50) // Reset

			di.IncYWithDX(dxVal)
			distAfterInc := di.Dist()

			di.DecYWithDX(dxVal)
			distAfterDec := di.Dist()

			// The exact values depend on internal calculations, but operations should be consistent
			_ = distAfterInc
			_ = distAfterDec
		}
	})
}

// TestDistanceInterpolator2Comprehensive tests all DistanceInterpolator2 functionality.
func TestDistanceInterpolator2Comprehensive(t *testing.T) {
	t.Run("Start interpolator", func(t *testing.T) {
		di := NewDistanceInterpolator2Start(0, 0, 100, 100, 10, 20, 50, 50)

		// Test all getter methods
		_ = di.Dist()
		_ = di.DistStart()
		_ = di.DistEnd() // Alias for DistStart
		_ = di.DX()
		_ = di.DY()
		_ = di.DXStart()
		_ = di.DYStart()
		_ = di.DXEnd() // Alias for DXStart
		_ = di.DYEnd() // Alias for DYStart

		// Test all movement operations
		initialDist := di.Dist()
		initialDistStart := di.DistStart()

		di.IncX()
		di.DecX() // Should return to initial values

		if di.Dist() != initialDist {
			t.Error("IncX followed by DecX should return to initial dist")
		}
		if di.DistStart() != initialDistStart {
			t.Error("IncX followed by DecX should return to initial distStart")
		}

		di.IncY()
		di.DecY() // Should return to initial values

		if di.Dist() != initialDist {
			t.Error("IncY followed by DecY should return to initial dist")
		}
		if di.DistStart() != initialDistStart {
			t.Error("IncY followed by DecY should return to initial distStart")
		}
	})

	t.Run("End interpolator", func(t *testing.T) {
		di := NewDistanceInterpolator2End(0, 0, 100, 100, 90, 80, 50, 50)

		// Test all getter methods
		_ = di.Dist()
		_ = di.DistStart() // Actually distEnd for this constructor
		_ = di.DistEnd()   // Same as DistStart
		_ = di.DX()
		_ = di.DY()
		_ = di.DXStart() // Actually dxEnd for this constructor
		_ = di.DYStart() // Actually dyEnd for this constructor
		_ = di.DXEnd()   // Same as DXStart
		_ = di.DYEnd()   // Same as DYStart

		// Test all movement operations with delta
		testDeltas := []int{-2, -1, 0, 1, 2}

		for _, delta := range testDeltas {
			di = NewDistanceInterpolator2End(0, 0, 100, 100, 90, 80, 50, 50) // Reset

			di.IncXWithDY(delta)
			_ = di.Dist()
			_ = di.DistStart()

			di = NewDistanceInterpolator2End(0, 0, 100, 100, 90, 80, 50, 50) // Reset

			di.DecXWithDY(delta)
			_ = di.Dist()
			_ = di.DistStart()

			di = NewDistanceInterpolator2End(0, 0, 100, 100, 90, 80, 50, 50) // Reset

			di.IncYWithDX(delta)
			_ = di.Dist()
			_ = di.DistStart()

			di = NewDistanceInterpolator2End(0, 0, 100, 100, 90, 80, 50, 50) // Reset

			di.DecYWithDX(delta)
			_ = di.Dist()
			_ = di.DistStart()
		}
	})
}

// TestDistanceInterpolator3Comprehensive tests all DistanceInterpolator3 functionality.
func TestDistanceInterpolator3Comprehensive(t *testing.T) {
	di := NewDistanceInterpolator3(0, 0, 100, 100, 10, 20, 90, 80, 50, 50)

	t.Run("Getter methods", func(t *testing.T) {
		// Test all getter methods
		_ = di.Dist()
		_ = di.DistStart()
		_ = di.DistEnd()
		_ = di.DX()
		_ = di.DY()
		_ = di.DXStart()
		_ = di.DYStart()
		_ = di.DXEnd()
		_ = di.DYEnd()
	})

	t.Run("Basic movement operations", func(t *testing.T) {
		initialDist := di.Dist()
		initialDistStart := di.DistStart()
		initialDistEnd := di.DistEnd()

		di.IncX()
		di.DecX() // Should return to initial values

		if di.Dist() != initialDist {
			t.Error("IncX followed by DecX should return to initial dist")
		}
		if di.DistStart() != initialDistStart {
			t.Error("IncX followed by DecX should return to initial distStart")
		}
		if di.DistEnd() != initialDistEnd {
			t.Error("IncX followed by DecX should return to initial distEnd")
		}

		di.IncY()
		di.DecY() // Should return to initial values

		if di.Dist() != initialDist {
			t.Error("IncY followed by DecY should return to initial dist")
		}
		if di.DistStart() != initialDistStart {
			t.Error("IncY followed by DecY should return to initial distStart")
		}
		if di.DistEnd() != initialDistEnd {
			t.Error("IncY followed by DecY should return to initial distEnd")
		}
	})

	t.Run("Delta movement operations", func(t *testing.T) {
		testDeltas := []int{-2, -1, 0, 1, 2}

		for _, delta := range testDeltas {
			di = NewDistanceInterpolator3(0, 0, 100, 100, 10, 20, 90, 80, 50, 50) // Reset

			di.IncXWithDY(delta)
			_ = di.Dist()
			_ = di.DistStart()
			_ = di.DistEnd()

			di = NewDistanceInterpolator3(0, 0, 100, 100, 10, 20, 90, 80, 50, 50) // Reset

			di.DecXWithDY(delta)
			_ = di.Dist()
			_ = di.DistStart()
			_ = di.DistEnd()

			di = NewDistanceInterpolator3(0, 0, 100, 100, 10, 20, 90, 80, 50, 50) // Reset

			di.IncYWithDX(delta)
			_ = di.Dist()
			_ = di.DistStart()
			_ = di.DistEnd()

			di = NewDistanceInterpolator3(0, 0, 100, 100, 10, 20, 90, 80, 50, 50) // Reset

			di.DecYWithDX(delta)
			_ = di.Dist()
			_ = di.DistStart()
			_ = di.DistEnd()
		}
	})
}

// TestDistanceInterpolatorEdgeCases tests edge cases for all interpolators.
func TestDistanceInterpolatorEdgeCases(t *testing.T) {
	t.Run("Zero coordinates", func(t *testing.T) {
		// Test with all zero coordinates
		di0 := NewDistanceInterpolator0(0, 0, 0, 0, 0, 0)
		di0.IncX()
		_ = di0.Dist()

		di00 := NewDistanceInterpolator00(0, 0, 0, 0, 0, 0, 0, 0)
		di00.IncX()
		_ = di00.Dist1()
		_ = di00.Dist2()

		di1 := NewDistanceInterpolator1(0, 0, 0, 0, 0, 0)
		di1.IncX()
		di1.DecX()
		di1.IncY()
		di1.DecY()
		di1.IncXWithDY(0)
		di1.DecXWithDY(0)
		di1.IncYWithDX(0)
		di1.DecYWithDX(0)
		_ = di1.Dist()
		_ = di1.DX()
		_ = di1.DY()

		di2start := NewDistanceInterpolator2Start(0, 0, 0, 0, 0, 0, 0, 0)
		di2start.IncX()
		di2start.DecX()
		di2start.IncY()
		di2start.DecY()
		di2start.IncXWithDY(0)
		di2start.DecXWithDY(0)
		di2start.IncYWithDX(0)
		di2start.DecYWithDX(0)

		di2end := NewDistanceInterpolator2End(0, 0, 0, 0, 0, 0, 0, 0)
		di2end.IncX()
		di2end.DecX()
		di2end.IncY()
		di2end.DecY()
		di2end.IncXWithDY(0)
		di2end.DecXWithDY(0)
		di2end.IncYWithDX(0)
		di2end.DecYWithDX(0)

		di3 := NewDistanceInterpolator3(0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
		di3.IncX()
		di3.DecX()
		di3.IncY()
		di3.DecY()
		di3.IncXWithDY(0)
		di3.DecXWithDY(0)
		di3.IncYWithDX(0)
		di3.DecYWithDX(0)
		_ = di3.Dist()
		_ = di3.DistStart()
		_ = di3.DistEnd()
		_ = di3.DX()
		_ = di3.DY()
		_ = di3.DXStart()
		_ = di3.DYStart()
		_ = di3.DXEnd()
		_ = di3.DYEnd()
	})

	t.Run("Large coordinates", func(t *testing.T) {
		// Test with large coordinates to check for overflow handling
		large := 1000000

		di0 := NewDistanceInterpolator0(-large, -large, large, large, 0, 0)
		di0.IncX()
		_ = di0.Dist()

		di1 := NewDistanceInterpolator1(-large, -large, large, large, 0, 0)
		di1.IncXWithDY(1)
		di1.DecXWithDY(1)
		di1.IncYWithDX(1)
		di1.DecYWithDX(1)

		di2 := NewDistanceInterpolator2Start(-large, -large, large, large, -large/2, -large/2, 0, 0)
		di2.IncXWithDY(1)
		di2.DecXWithDY(1)
		di2.IncYWithDX(1)
		di2.DecYWithDX(1)

		di3 := NewDistanceInterpolator3(-large, -large, large, large, -large/2, -large/2, large/2, large/2, 0, 0)
		di3.IncXWithDY(1)
		di3.DecXWithDY(1)
		di3.IncYWithDX(1)
		di3.DecYWithDX(1)
	})

	t.Run("LineMR coordinate handling", func(t *testing.T) {
		// Test with coordinates that will exercise LineMR function
		coords := []int{
			0,
			primitives.LineSubpixelScale / 2,
			primitives.LineSubpixelScale,
			primitives.LineSubpixelScale * 2,
			-primitives.LineSubpixelScale,
		}

		for _, coord := range coords {
			di0 := NewDistanceInterpolator0(coord, coord, coord+100, coord+100, coord+50, coord+50)
			di0.IncX()
			_ = di0.Dist()

			di00 := NewDistanceInterpolator00(coord+50, coord+50, coord, coord, coord+100, coord+100, coord+25, coord+25)
			di00.IncX()
			_ = di00.Dist1()
			_ = di00.Dist2()
		}
	})
}
