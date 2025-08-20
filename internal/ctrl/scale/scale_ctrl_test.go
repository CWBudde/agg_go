package scale

import (
	"math"
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
)

func TestNewScaleCtrl(t *testing.T) {
	scale := NewScaleCtrl(10, 20, 200, 50, false)

	// Test basic properties
	if scale.X1() != 10 || scale.Y1() != 20 || scale.X2() != 200 || scale.Y2() != 50 {
		t.Errorf("Expected bounds (10, 20, 200, 50), got (%.1f, %.1f, %.1f, %.1f)",
			scale.X1(), scale.Y1(), scale.X2(), scale.Y2())
	}

	// Test default values
	if scale.Value1() != 0.3 || scale.Value2() != 0.7 {
		t.Errorf("Expected default values (0.3, 0.7), got (%.2f, %.2f)",
			scale.Value1(), scale.Value2())
	}

	// Test minimum delta
	if scale.MinDelta() != 0.01 {
		t.Errorf("Expected default MinDelta 0.01, got %.3f", scale.MinDelta())
	}

	// Test border extra calculation for horizontal control
	expectedExtra := (50 - 20) / 2.0 // (y2 - y1) / 2
	if scale.borderExtra != expectedExtra {
		t.Errorf("Expected borderExtra %.1f for horizontal control, got %.1f",
			expectedExtra, scale.borderExtra)
	}
}

func TestScaleCtrlVertical(t *testing.T) {
	// Create vertical scale control (height > width)
	scale := NewScaleCtrl(10, 20, 40, 200, false)

	// Test border extra calculation for vertical control
	expectedExtra := (40 - 10) / 2.0 // (x2 - x1) / 2
	if scale.borderExtra != expectedExtra {
		t.Errorf("Expected borderExtra %.1f for vertical control, got %.1f",
			expectedExtra, scale.borderExtra)
	}
}

func TestValueConstraints(t *testing.T) {
	scale := NewScaleCtrl(0, 0, 100, 20, false)
	scale.SetMinDelta(0.1) // 10% minimum distance

	// Test Value1 constraints
	scale.SetValue1(-0.5) // Should clamp to 0
	if scale.Value1() != 0.0 {
		t.Errorf("Expected Value1 to clamp to 0.0, got %.2f", scale.Value1())
	}

	// Test clamping with minD constraint - when value2=0.7, max value1 is 0.6
	scale.SetValue1(1.5)                             // Should be constrained by minD
	expectedMax := scale.Value2() - scale.MinDelta() // 0.7 - 0.1 = 0.6
	if math.Abs(scale.Value1()-expectedMax) > 0.001 {
		t.Errorf("Expected Value1 to be constrained to %.2f, got %.2f", expectedMax, scale.Value1())
	}

	// Test minimum distance enforcement - reset values first
	scale.SetValue1(0.2)        // Set to a low value first
	scale.SetValue2(0.5)        // Now this should work
	scale.SetValue1(0.45)       // Should be adjusted to maintain minD
	expectedValue1 := 0.5 - 0.1 // value2 - minD
	if math.Abs(scale.Value1()-expectedValue1) > 0.001 {
		t.Errorf("Expected Value1 %.2f to maintain minD, got %.2f",
			expectedValue1, scale.Value1())
	}

	// Test Value2 constraints
	scale.SetValue1(0.3)
	scale.SetValue2(0.35)       // Should be adjusted to maintain minD
	expectedValue2 := 0.3 + 0.1 // value1 + minD
	if math.Abs(scale.Value2()-expectedValue2) > 0.001 {
		t.Errorf("Expected Value2 %.2f to maintain minD, got %.2f",
			expectedValue2, scale.Value2())
	}
}

func TestMove(t *testing.T) {
	scale := NewScaleCtrl(0, 0, 100, 20, false)
	scale.SetValue1(0.3)
	scale.SetValue2(0.7)

	// Test normal move
	scale.Move(0.1)
	if math.Abs(scale.Value1()-0.4) > 0.001 || math.Abs(scale.Value2()-0.8) > 0.001 {
		t.Errorf("Expected values (0.4, 0.8) after move, got (%.2f, %.2f)",
			scale.Value1(), scale.Value2())
	}

	// Test move with upper bound constraint
	scale.Move(0.3) // This should hit the upper bound
	if scale.Value2() != 1.0 {
		t.Errorf("Expected Value2 to clamp to 1.0, got %.2f", scale.Value2())
	}

	// Test move with lower bound constraint
	scale.SetValue1(0.2)
	scale.SetValue2(0.4)
	scale.Move(-0.3) // This should hit the lower bound
	if scale.Value1() != 0.0 {
		t.Errorf("Expected Value1 to clamp to 0.0, got %.2f", scale.Value1())
	}
}

func TestColorMethods(t *testing.T) {
	scale := NewScaleCtrl(0, 0, 100, 20, false)

	// Test color setting and retrieval
	testColor := color.NewRGBA(1.0, 0.0, 0.0, 1.0) // Red
	scale.BackgroundColor(testColor)

	bgColor := scale.Color(0).(color.RGBA)
	if bgColor.R != 1.0 || bgColor.G != 0.0 || bgColor.B != 0.0 || bgColor.A != 1.0 {
		t.Errorf("Expected background color (1.0, 0.0, 0.0, 1.0), got (%.1f, %.1f, %.1f, %.1f)",
			bgColor.R, bgColor.G, bgColor.B, bgColor.A)
	}

	// Test pointers color affects both pointers
	scale.PointersColor(testColor)
	pointer1Color := scale.Color(2).(color.RGBA)
	pointer2Color := scale.Color(3).(color.RGBA)

	if pointer1Color != testColor || pointer2Color != testColor {
		t.Error("PointersColor should set both pointer colors")
	}
}

func TestInRect(t *testing.T) {
	scale := NewScaleCtrl(10, 20, 100, 50, false)

	// Test points inside bounds
	if !scale.InRect(50, 35) {
		t.Error("Point (50, 35) should be inside bounds")
	}

	// Test points outside bounds
	if scale.InRect(5, 35) {
		t.Error("Point (5, 35) should be outside bounds")
	}
	if scale.InRect(50, 15) {
		t.Error("Point (50, 15) should be outside bounds")
	}
}

func TestMouseInteractionHorizontal(t *testing.T) {
	scale := NewScaleCtrl(0, 0, 200, 20, false)
	scale.SetValue1(0.3)
	scale.SetValue2(0.7)

	// Calculate pointer positions
	pointer1X := scale.xs1 + (scale.xs2-scale.xs1)*scale.Value1()
	pointerY := (scale.ys1 + scale.ys2) / 2.0

	// Test clicking on pointer 1
	result := scale.OnMouseButtonDown(pointer1X, pointerY)
	if !result {
		t.Error("Should successfully click on pointer 1")
	}
	if scale.moveWhat != MoveValue1 {
		t.Errorf("Expected MoveValue1, got %v", scale.moveWhat)
	}

	// Test dragging pointer 1
	newX := scale.xs1 + (scale.xs2-scale.xs1)*0.2 // Move to 20%
	result = scale.OnMouseMove(newX, pointerY, true)
	if !result {
		t.Error("Mouse move should return true when dragging")
	}

	// Value should be approximately 0.2
	if math.Abs(scale.Value1()-0.2) > 0.05 {
		t.Errorf("Expected Value1 around 0.2, got %.3f", scale.Value1())
	}

	// Test ending drag
	scale.OnMouseButtonUp(newX, pointerY)
	if scale.moveWhat != MoveNothing {
		t.Error("Expected MoveNothing after mouse button up")
	}
}

func TestMouseInteractionSlider(t *testing.T) {
	scale := NewScaleCtrl(0, 0, 200, 20, false)
	scale.SetValue1(0.3)
	scale.SetValue2(0.7)

	// Click in the middle of the slider bar
	sliderX := scale.xs1 + (scale.xs2-scale.xs1)*0.5 // Middle of range
	sliderY := (scale.ys1 + scale.ys2) / 2.0

	result := scale.OnMouseButtonDown(sliderX, sliderY)
	if !result {
		t.Error("Should successfully click on slider bar")
	}
	if scale.moveWhat != MoveSlider {
		t.Errorf("Expected MoveSlider, got %v", scale.moveWhat)
	}

	// Test dragging the entire slider
	originalRange := scale.Value2() - scale.Value1()
	newX := scale.xs1 + (scale.xs2-scale.xs1)*0.1 // Move range to start at 10%

	result = scale.OnMouseMove(newX, sliderY, true)
	if !result {
		t.Error("Mouse move should return true when dragging slider")
	}

	// Range should be preserved
	newRange := scale.Value2() - scale.Value1()
	if math.Abs(newRange-originalRange) > 0.01 {
		t.Errorf("Expected range %.3f to be preserved, got %.3f",
			originalRange, newRange)
	}
}

func TestVertexGeneration(t *testing.T) {
	scale := NewScaleCtrl(0, 0, 100, 20, false)

	// Test NumPaths
	if scale.NumPaths() != 5 {
		t.Errorf("Expected 5 paths, got %d", scale.NumPaths())
	}

	// Test background path (path 0)
	scale.Rewind(0)
	_, _, cmd := scale.Vertex()
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected first vertex to be MoveTo, got %v", cmd)
	}

	// Should generate 4 vertices for rectangle
	vertexCount := 1 // Already got the first vertex
	for {
		_, _, cmd = scale.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertexCount++
		if vertexCount > 10 { // Safety check
			t.Fatal("Too many vertices generated")
		}
	}

	if vertexCount != 4 {
		t.Errorf("Expected 4 vertices for background rectangle, got %d", vertexCount)
	}

	// Test pointer path (ellipse)
	scale.Rewind(2)
	vertexCount = 0
	for {
		_, _, cmd = scale.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertexCount++
		if vertexCount > 100 { // Safety check for ellipse
			t.Fatal("Too many vertices generated for ellipse")
		}
	}

	if vertexCount == 0 {
		t.Error("Expected vertices for pointer ellipse")
	}
}

func TestVertexGenerationVertical(t *testing.T) {
	// Test vertical scale control
	scale := NewScaleCtrl(0, 0, 20, 100, false) // Tall and narrow
	scale.SetValue1(0.2)
	scale.SetValue2(0.8)

	// Test slider path in vertical orientation
	scale.Rewind(4)
	x1, _, cmd1 := scale.Vertex()
	if cmd1 != basics.PathCmdMoveTo {
		t.Error("Expected first slider vertex to be MoveTo")
	}

	// In vertical mode, slider should span horizontally across the control
	expectedMinX := scale.X1() - scale.borderExtra/2.0
	if math.Abs(x1-expectedMinX) > 1.0 {
		t.Errorf("Expected slider X coordinate near %.1f, got %.1f", expectedMinX, x1)
	}
}

func TestBorderThickness(t *testing.T) {
	scale := NewScaleCtrl(0, 0, 100, 20, false)

	originalXS1 := scale.xs1

	// Change border thickness
	scale.BorderThickness(5.0, 2.0)

	if scale.borderThickness != 5.0 {
		t.Errorf("Expected borderThickness 5.0, got %.1f", scale.borderThickness)
	}
	if scale.borderExtra != 2.0 {
		t.Errorf("Expected borderExtra 2.0, got %.1f", scale.borderExtra)
	}

	// Inner bounds should have changed
	if scale.xs1 == originalXS1 {
		t.Error("Inner bounds should have changed after BorderThickness update")
	}
}

func TestResize(t *testing.T) {
	scale := NewScaleCtrl(0, 0, 100, 30, false) // horizontal: borderExtra = 15
	originalBorderExtra := scale.borderExtra

	// Resize to make it vertical with different aspect ratio
	scale.Resize(0, 0, 20, 100) // vertical: borderExtra = 10

	if scale.X2() != 20 || scale.Y2() != 100 {
		t.Errorf("Expected bounds (0, 0, 20, 100), got (%.1f, %.1f, %.1f, %.1f)",
			scale.X1(), scale.Y1(), scale.X2(), scale.Y2())
	}

	// Border extra should change for new orientation
	if scale.borderExtra == originalBorderExtra {
		t.Error("Border extra should change when orientation changes")
	}

	expectedBorderExtra := (20 - 0) / 2.0 // (x2 - x1) / 2 for vertical = 10
	if math.Abs(scale.borderExtra-expectedBorderExtra) > 0.001 {
		t.Errorf("Expected borderExtra %.1f for vertical orientation, got %.1f",
			expectedBorderExtra, scale.borderExtra)
	}
}

func TestArrowKeys(t *testing.T) {
	scale := NewScaleCtrl(0, 0, 100, 20, false)

	// Arrow keys are not implemented, should return false
	result := scale.OnArrowKeys(true, false, false, false)
	if result {
		t.Error("Arrow keys should return false (not implemented)")
	}
}
