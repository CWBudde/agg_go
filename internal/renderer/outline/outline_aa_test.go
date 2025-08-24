// Package outline provides anti-aliased outline rendering functionality.
// This file contains comprehensive tests for the renderer_outline_aa implementation.
package outline

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/primitives"
)

// MockOutlineRenderer implements OutlineRenderer for testing line interpolators.
type MockOutlineRenderer struct {
	width, height  int
	subpixelWidth  int
	hspanCalls     []HSpanCall
	vspanCalls     []VSpanCall
	coverCallCount int
	coverValues    []int
}

// MockBaseRenderer implements BaseRendererInterface for testing the main renderer.
type MockBaseRenderer struct {
	width, height int
	hspanCalls    []HSpanCall
	vspanCalls    []VSpanCall
}

type HSpanCall struct {
	x, y, length int
	color        TestColor
	covers       []basics.CoverType
}

type VSpanCall struct {
	x, y, length int
	color        TestColor
	covers       []basics.CoverType
}

func NewMockOutlineRenderer(width, height, subpixelWidth int) *MockOutlineRenderer {
	return &MockOutlineRenderer{
		width:         width,
		height:        height,
		subpixelWidth: subpixelWidth,
		coverValues:   make([]int, 1000), // Preset cover values
	}
}

func NewMockBaseRenderer(width, height int) *MockBaseRenderer {
	return &MockBaseRenderer{
		width:  width,
		height: height,
	}
}

// MockOutlineRenderer methods (OutlineRenderer interface)
func (m *MockOutlineRenderer) Width() int         { return m.width }
func (m *MockOutlineRenderer) Height() int        { return m.height }
func (m *MockOutlineRenderer) SubpixelWidth() int { return m.subpixelWidth }

func (m *MockOutlineRenderer) Cover(d int) int {
	m.coverCallCount++
	if d >= 0 && d < len(m.coverValues) {
		return m.coverValues[d]
	}
	return 128 // Default cover value
}

func (m *MockOutlineRenderer) BlendSolidHSpan(x, y, length int, covers []basics.CoverType) {
	// Make a copy of covers slice to avoid reference issues
	coversCopy := make([]basics.CoverType, len(covers))
	copy(coversCopy, covers)

	m.hspanCalls = append(m.hspanCalls, HSpanCall{
		x: x, y: y, length: length, color: TestColor{}, covers: coversCopy,
	})
}

func (m *MockOutlineRenderer) BlendSolidVSpan(x, y, length int, covers []basics.CoverType) {
	// Make a copy of covers slice to avoid reference issues
	coversCopy := make([]basics.CoverType, len(covers))
	copy(coversCopy, covers)

	m.vspanCalls = append(m.vspanCalls, VSpanCall{
		x: x, y: y, length: length, color: TestColor{}, covers: coversCopy,
	})
}

// MockBaseRenderer methods (BaseRendererInterface)
func (m *MockBaseRenderer) Width() int  { return m.width }
func (m *MockBaseRenderer) Height() int { return m.height }

func (m *MockBaseRenderer) BlendSolidHSpan(x, y, length int, color TestColor, covers []basics.CoverType) {
	// Make a copy of covers slice to avoid reference issues
	coversCopy := make([]basics.CoverType, len(covers))
	copy(coversCopy, covers)

	m.hspanCalls = append(m.hspanCalls, HSpanCall{
		x: x, y: y, length: length, color: color, covers: coversCopy,
	})
}

func (m *MockBaseRenderer) BlendSolidVSpan(x, y, length int, color TestColor, covers []basics.CoverType) {
	// Make a copy of covers slice to avoid reference issues
	coversCopy := make([]basics.CoverType, len(covers))
	copy(coversCopy, covers)

	m.vspanCalls = append(m.vspanCalls, VSpanCall{
		x: x, y: y, length: length, color: color, covers: coversCopy,
	})
}

// Simple color type for testing
type TestColor struct {
	R, G, B, A basics.Int8u
}

// TestDistanceInterpolators tests all distance interpolator classes.
func TestDistanceInterpolators(t *testing.T) {
	t.Run("DistanceInterpolator0", func(t *testing.T) {
		di := NewDistanceInterpolator0(0, 0, 100, 100, 50, 50)

		initialDist := di.Dist()
		di.IncX()
		newDist := di.Dist()

		// Distance should change after increment
		if initialDist == newDist {
			t.Error("Distance should change after IncX()")
		}
	})

	t.Run("DistanceInterpolator00", func(t *testing.T) {
		di := NewDistanceInterpolator00(50, 50, 0, 0, 100, 100, 25, 25)

		dist1 := di.Dist1()
		dist2 := di.Dist2()

		di.IncX()

		newDist1 := di.Dist1()
		newDist2 := di.Dist2()

		// Both distances should change
		if dist1 == newDist1 {
			t.Error("Dist1 should change after IncX()")
		}
		if dist2 == newDist2 {
			t.Error("Dist2 should change after IncX()")
		}
	})

	t.Run("DistanceInterpolator1", func(t *testing.T) {
		di := NewDistanceInterpolator1(0, 0, 100, 100, 50, 50)

		initialDist := di.Dist()

		// Test all movement directions
		di.IncX()
		di.DecX() // Should return to initial

		if di.Dist() != initialDist {
			t.Error("IncX followed by DecX should return to initial distance")
		}

		di.IncY()
		di.DecY() // Should return to initial

		if di.Dist() != initialDist {
			t.Error("IncY followed by DecY should return to initial distance")
		}
	})

	t.Run("DistanceInterpolator2", func(t *testing.T) {
		di := NewDistanceInterpolator2Start(0, 0, 100, 100, 10, 20, 50, 50)

		initialDist := di.Dist()
		initialDistStart := di.DistStart()

		di.IncX()

		if di.Dist() == initialDist {
			t.Error("Dist should change after IncX()")
		}
		// Only check if the start deltas are non-zero
		if di.DYStart() != 0 && di.DistStart() == initialDistStart {
			t.Error("DistStart should change after IncX() when start deltas are non-zero")
		}
	})

	t.Run("DistanceInterpolator3", func(t *testing.T) {
		di := NewDistanceInterpolator3(0, 0, 100, 100, 10, 20, 90, 80, 50, 50)

		initialDist := di.Dist()
		initialDistStart := di.DistStart()
		initialDistEnd := di.DistEnd()

		di.IncX()

		if di.Dist() == initialDist {
			t.Error("Dist should change after IncX()")
		}
		// Only check if the deltas are non-zero
		if di.DYStart() != 0 && di.DistStart() == initialDistStart {
			t.Error("DistStart should change after IncX() when start deltas are non-zero")
		}
		if di.DYEnd() != 0 && di.DistEnd() == initialDistEnd {
			t.Error("DistEnd should change after IncX() when end deltas are non-zero")
		}
	})
}

// TestLineProfileAA tests the line profile functionality.
func TestLineProfileAA(t *testing.T) {
	t.Run("BasicCreation", func(t *testing.T) {
		profile := NewLineProfileAA()

		if profile.GetMinWidth() != 1.0 {
			t.Errorf("Expected min width 1.0, got %f", profile.GetMinWidth())
		}

		if profile.GetSmootherWidth() != 1.0 {
			t.Errorf("Expected smoother width 1.0, got %f", profile.GetSmootherWidth())
		}
	})

	t.Run("WidthSetting", func(t *testing.T) {
		profile := NewLineProfileAA()
		profile.Width(2.0)

		if profile.SubpixelWidth() == 0 {
			t.Error("Subpixel width should be set after Width() call")
		}

		if profile.ProfileSize() == 0 {
			t.Error("Profile size should be greater than 0 after Width() call")
		}
	})

	t.Run("ValueRetrieval", func(t *testing.T) {
		profile := NewLineProfileAA()
		profile.Width(1.0)

		// Value at distance 0 (center) should be non-zero
		centerValue := profile.Value(0)
		if centerValue == 0 {
			t.Error("Center value should be non-zero")
		}
	})

	t.Run("GammaFunction", func(t *testing.T) {
		profile := NewLineProfileAA()

		// Simple linear gamma function for testing
		linearGamma := &LinearGammaFunction{}
		profile.SetGamma(linearGamma)
		profile.Width(1.0)

		// Should not panic and should produce valid values
		value := profile.Value(SubpixelScale)
		if value < 0 {
			t.Error("Profile value should not be negative")
		}
	})
}

// LinearGammaFunction implements a simple linear gamma function for testing.
type LinearGammaFunction struct{}

func (lgf *LinearGammaFunction) Call(x float64) float64 {
	return x // Linear gamma (no correction)
}

// TestLineInterpolators tests the line interpolator functionality.
func TestLineInterpolators(t *testing.T) {
	renderer := NewMockOutlineRenderer(800, 600, 256)

	// Set up some reasonable cover values
	for i := 0; i < len(renderer.coverValues); i++ {
		if i < 128 {
			renderer.coverValues[i] = 255 - i*2 // Decreasing coverage
		} else {
			renderer.coverValues[i] = 0
		}
	}

	// Create a test line
	lp := primitives.NewLineParameters(
		100<<primitives.LineSubpixelShift,
		100<<primitives.LineSubpixelShift,
		200<<primitives.LineSubpixelShift,
		200<<primitives.LineSubpixelShift,
		100)

	t.Run("LineInterpolatorAA0", func(t *testing.T) {
		li := NewLineInterpolatorAA0(renderer, &lp)

		if li.Count() <= 0 {
			t.Error("Line interpolator should have steps to process")
		}

		initialSpanCalls := len(renderer.vspanCalls) + len(renderer.hspanCalls)

		// Process a few steps
		stepCount := 0
		for li.Vertical() && li.StepVer() && stepCount < 5 {
			stepCount++
		}
		for !li.Vertical() && li.StepHor() && stepCount < 5 {
			stepCount++
		}

		finalSpanCalls := len(renderer.vspanCalls) + len(renderer.hspanCalls)

		if finalSpanCalls <= initialSpanCalls {
			t.Error("Line interpolator should have made span calls")
		}
	})

	t.Run("LineInterpolatorAA1", func(t *testing.T) {
		sx := 90 << primitives.LineSubpixelShift
		sy := 90 << primitives.LineSubpixelShift

		li := NewLineInterpolatorAA1(renderer, &lp, sx, sy)

		if li.Count() <= 0 {
			t.Error("Line interpolator should have steps to process")
		}

		// Should be able to process at least one step without panicking
		if li.Vertical() {
			li.StepVer()
		} else {
			li.StepHor()
		}
	})
}

// TestRendererOutlineAA tests the main renderer functionality.
func TestRendererOutlineAA(t *testing.T) {
	renderer := NewMockBaseRenderer(800, 600)
	profile := NewLineProfileAA()
	profile.Width(2.0)

	outlineRenderer := NewRendererOutlineAA[*MockBaseRenderer, TestColor](renderer, profile)

	// Set test color
	testColor := TestColor{R: 255, G: 0, B: 0, A: 255}
	outlineRenderer.Color(testColor)

	t.Run("BasicProperties", func(t *testing.T) {
		if outlineRenderer.SubpixelWidth() != profile.SubpixelWidth() {
			t.Error("Outline renderer subpixel width should match profile")
		}

		if outlineRenderer.GetColor() != testColor {
			t.Error("Color should be set correctly")
		}

		if outlineRenderer.GetProfile() != profile {
			t.Error("Profile should be set correctly")
		}
	})

	t.Run("ClippingOperations", func(t *testing.T) {
		outlineRenderer.ResetClipping()
		outlineRenderer.ClipBox(0, 0, 800, 600)

		// Should not panic when calling clipping operations
	})

	t.Run("Line0Rendering", func(t *testing.T) {
		lp := primitives.NewLineParameters(
			100<<primitives.LineSubpixelShift,
			100<<primitives.LineSubpixelShift,
			200<<primitives.LineSubpixelShift,
			100<<primitives.LineSubpixelShift,
			100)

		initialSpanCalls := len(renderer.hspanCalls) + len(renderer.vspanCalls)

		outlineRenderer.Line0(&lp)

		finalSpanCalls := len(renderer.hspanCalls) + len(renderer.vspanCalls)

		if finalSpanCalls <= initialSpanCalls {
			t.Error("Line0 rendering should have made span calls")
		}
	})

	t.Run("Line1Rendering", func(t *testing.T) {
		lp := primitives.NewLineParameters(
			100<<primitives.LineSubpixelShift,
			100<<primitives.LineSubpixelShift,
			200<<primitives.LineSubpixelShift,
			150<<primitives.LineSubpixelShift,
			100)

		sx := 90 << primitives.LineSubpixelShift
		sy := 90 << primitives.LineSubpixelShift

		initialSpanCalls := len(renderer.hspanCalls) + len(renderer.vspanCalls)

		outlineRenderer.Line1(&lp, sx, sy)

		finalSpanCalls := len(renderer.hspanCalls) + len(renderer.vspanCalls)

		if finalSpanCalls <= initialSpanCalls {
			t.Error("Line1 rendering should have made span calls")
		}
	})
}

// TestCoverageCalculation tests coverage calculation accuracy.
func TestCoverageCalculation(t *testing.T) {
	profile := NewLineProfileAA()
	profile.Width(2.0)

	renderer := NewMockBaseRenderer(800, 600)
	outlineRenderer := NewRendererOutlineAA[*MockBaseRenderer, TestColor](renderer, profile)

	// Test different distance values
	distances := []int{0, SubpixelScale / 4, SubpixelScale / 2, SubpixelScale, SubpixelScale * 2}

	for _, dist := range distances {
		coverage := outlineRenderer.Cover(dist)

		if coverage < 0 {
			t.Errorf("Coverage should not be negative for distance %d, got %d", dist, coverage)
		}

		if coverage > 255 {
			t.Errorf("Coverage should not exceed 255 for distance %d, got %d", dist, coverage)
		}
	}
}

// TestEdgeCases tests various edge cases and boundary conditions.
func TestEdgeCases(t *testing.T) {
	t.Run("ZeroWidthLine", func(t *testing.T) {
		profile := NewLineProfileAA()
		profile.Width(0.0)

		// Should not panic
		if profile.ProfileSize() <= 0 {
			t.Error("Profile should have non-zero size even for zero width")
		}
	})

	t.Run("VeryLargeLine", func(t *testing.T) {
		renderer := NewMockBaseRenderer(2000, 2000)
		profile := NewLineProfileAA()
		profile.Width(10.0)

		outlineRenderer := NewRendererOutlineAA[*MockBaseRenderer, TestColor](renderer, profile)

		// Create a line longer than LineMaxLength to test subdivision
		lp := primitives.NewLineParameters(
			0, 0,
			primitives.LineMaxLength+100,
			primitives.LineMaxLength+100,
			primitives.LineMaxLength+100)

		// Should handle line subdivision without panicking
		outlineRenderer.Line0(&lp)
	})

	t.Run("NegativeCoordinates", func(t *testing.T) {
		renderer := NewMockBaseRenderer(800, 600)
		profile := NewLineProfileAA()
		profile.Width(2.0)

		outlineRenderer := NewRendererOutlineAA[*MockBaseRenderer, TestColor](renderer, profile)

		lp := primitives.NewLineParameters(
			-100<<primitives.LineSubpixelShift,
			-100<<primitives.LineSubpixelShift,
			100<<primitives.LineSubpixelShift,
			100<<primitives.LineSubpixelShift,
			200)

		// Should handle negative coordinates without panicking
		outlineRenderer.Line0(&lp)
	})
}
