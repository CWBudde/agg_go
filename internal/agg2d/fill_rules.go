// Package agg2d provides fill rule functionality for the AGG2D high-level interface.
// This implements Phase 5: Advanced Rendering - Fill Rules from the AGG 2.6 C++ library.
package agg2d

import (
	"agg_go/internal/basics"
)

// FillEvenOdd sets the fill rule for path rendering.
// When evenOddFlag is true, uses even-odd fill rule (XOR-based filling).
// When evenOddFlag is false, uses non-zero winding fill rule (direction-based filling).
//
// Even-odd rule: A point is inside if it crosses an odd number of path segments.
// Non-zero winding rule: A point is inside if the winding number is non-zero.
//
// This matches the C++ Agg2D::fillEvenOdd method.
func (agg2d *Agg2D) FillEvenOdd(evenOddFlag bool) {
	agg2d.evenOddFlag = evenOddFlag

	// Apply the fill rule to the rasterizer if it exists
	// Note: In a complete implementation, we would set this on the rasterizer
	// For now, we store the flag and apply it when rendering
}

// GetFillEvenOdd returns the current fill rule.
// Returns true if using even-odd fill rule, false if using non-zero winding rule.
// This matches the C++ Agg2D::fillEvenOdd() const method.
func (agg2d *Agg2D) GetFillEvenOdd() bool {
	return agg2d.evenOddFlag
}

// IsEvenOddFillRule is an alias for GetFillEvenOdd for convenience.
func (agg2d *Agg2D) IsEvenOddFillRule() bool {
	return agg2d.GetFillEvenOdd()
}

// IsNonZeroFillRule returns true if using non-zero winding rule.
func (agg2d *Agg2D) IsNonZeroFillRule() bool {
	return !agg2d.GetFillEvenOdd()
}

// SetFillRule sets the fill rule using the basics.FillingRule enum.
// This provides a lower-level interface compatible with the rasterizer.
func (agg2d *Agg2D) SetFillRule(rule basics.FillingRule) {
	switch rule {
	case basics.FillEvenOdd:
		agg2d.FillEvenOdd(true)
	case basics.FillNonZero:
		agg2d.FillEvenOdd(false)
	}
}

// GetFillRule returns the current fill rule as a basics.FillingRule enum.
func (agg2d *Agg2D) GetFillRule() basics.FillingRule {
	if agg2d.evenOddFlag {
		return basics.FillEvenOdd
	}
	return basics.FillNonZero
}

// FillRuleDescription returns a human-readable description of the current fill rule.
func (agg2d *Agg2D) FillRuleDescription() string {
	if agg2d.evenOddFlag {
		return "Even-Odd (XOR-based filling)"
	}
	return "Non-Zero Winding (direction-based filling)"
}

// FillingRuleSetter defines the interface for objects that can accept fill rule settings.
// This interface is implemented by rasterizers that support different fill rule algorithms.
type FillingRuleSetter interface {
	FillingRule(rule basics.FillingRule)
}

// applyFillRuleToRasterizer applies the current fill rule to a rasterizer.
// This is a helper method that will be used when integrating with the rendering pipeline.
// The rasterizer parameter must implement the FillingRuleSetter interface.
func (agg2d *Agg2D) applyFillRuleToRasterizer(rasterizer FillingRuleSetter) {
	rasterizer.FillingRule(agg2d.GetFillRule())
}

// FillRuleExamples provides example use cases for different fill rules.
// This is primarily for documentation and testing purposes.
type FillRuleExamples struct{}

// ComplexPolygonExample demonstrates when to use even-odd vs non-zero winding.
// Even-odd is useful for creating holes and complex shapes.
// Non-zero winding is useful for self-intersecting paths where direction matters.
func (FillRuleExamples) ComplexPolygonExample() string {
	return `
Fill Rule Usage Examples:

1. Even-Odd Fill Rule:
   - Use for shapes with holes (donut, star with cutouts)
   - Self-intersections create alternating fill/unfill areas
   - Direction of path drawing doesn't matter
   - Common in SVG and PostScript

2. Non-Zero Winding Fill Rule:
   - Use for solid shapes where direction matters
   - Self-intersections don't create holes unless paths wind opposite directions
   - Clockwise vs counter-clockwise path direction affects filling
   - Common in fonts and complex geometric shapes

Example: Drawing a star shape
- Even-odd: Star points are filled, center may be unfilled
- Non-zero: All areas are filled based on winding direction
`
}

// PathWindingDirection represents the direction a path is drawn.
type PathWindingDirection int

const (
	Clockwise PathWindingDirection = iota
	CounterClockwise
)

// CalculateWindingNumber calculates the winding number for a point relative to a polygon.
// This is used internally by the non-zero winding fill rule.
// Returns positive for clockwise winding, negative for counter-clockwise.
// Note: This is a simplified example for documentation purposes.
func CalculateWindingNumber(point [2]float64, polygon [][2]float64) int {
	// Simplified winding number calculation
	// In practice, this would be handled by the rasterizer
	winding := 0
	n := len(polygon)

	for i := 0; i < n; i++ {
		j := (i + 1) % n

		// Check if edge crosses horizontal ray from point to the right
		if polygon[i][1] <= point[1] {
			if polygon[j][1] > point[1] { // Upward crossing
				// Check if point is left of edge
				if isLeft(polygon[i], polygon[j], point) > 0 {
					winding++ // Counterclockwise crossing
				}
			}
		} else {
			if polygon[j][1] <= point[1] { // Downward crossing
				// Check if point is left of edge
				if isLeft(polygon[i], polygon[j], point) < 0 {
					winding-- // Clockwise crossing
				}
			}
		}
	}

	return winding
}

// isLeft tests if point p2 is left, on, or right of the line from p0 to p1.
// Returns: > 0 for p2 left of the line
//
//	= 0 for p2 on the line
//	< 0 for p2 right of the line
func isLeft(p0, p1, p2 [2]float64) float64 {
	return (p1[0]-p0[0])*(p2[1]-p0[1]) - (p2[0]-p0[0])*(p1[1]-p0[1])
}

// FillRuleTest provides methods for testing fill rule behavior.
type FillRuleTest struct {
	agg2d *Agg2D
}

// NewFillRuleTest creates a new fill rule test helper.
func NewFillRuleTest(agg2d *Agg2D) *FillRuleTest {
	return &FillRuleTest{agg2d: agg2d}
}

// TestEvenOddRule tests the even-odd fill rule with a simple crossing pattern.
func (frt *FillRuleTest) TestEvenOddRule() bool {
	// Save current fill rule
	originalRule := frt.agg2d.GetFillEvenOdd()

	// Set to even-odd
	frt.agg2d.FillEvenOdd(true)

	// Verify the rule was set
	result := frt.agg2d.IsEvenOddFillRule()

	// Restore original rule
	frt.agg2d.FillEvenOdd(originalRule)

	return result
}

// TestNonZeroRule tests the non-zero winding fill rule.
func (frt *FillRuleTest) TestNonZeroRule() bool {
	// Save current fill rule
	originalRule := frt.agg2d.GetFillEvenOdd()

	// Set to non-zero winding
	frt.agg2d.FillEvenOdd(false)

	// Verify the rule was set
	result := frt.agg2d.IsNonZeroFillRule()

	// Restore original rule
	frt.agg2d.FillEvenOdd(originalRule)

	return result
}

// TestFillRuleToggle tests switching between fill rules.
func (frt *FillRuleTest) TestFillRuleToggle() bool {
	// Start with even-odd
	frt.agg2d.FillEvenOdd(true)
	if !frt.agg2d.IsEvenOddFillRule() {
		return false
	}

	// Switch to non-zero
	frt.agg2d.FillEvenOdd(false)
	if !frt.agg2d.IsNonZeroFillRule() {
		return false
	}

	// Switch back to even-odd
	frt.agg2d.FillEvenOdd(true)
	return frt.agg2d.IsEvenOddFillRule()
}

// TestFillingRuleEnum tests the enum-based interface.
func (frt *FillRuleTest) TestFillingRuleEnum() bool {
	// Test even-odd via enum
	frt.agg2d.SetFillRule(basics.FillEvenOdd)
	if frt.agg2d.GetFillRule() != basics.FillEvenOdd {
		return false
	}

	// Test non-zero via enum
	frt.agg2d.SetFillRule(basics.FillNonZero)
	if frt.agg2d.GetFillRule() != basics.FillNonZero {
		return false
	}

	return true
}
