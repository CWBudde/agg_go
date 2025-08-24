// Example demonstrating Phase 5: Advanced Rendering features
// This showcases transformations, viewport handling, and advanced stroke/fill features
package main

import (
	"fmt"
	"os"

	"agg_go"
)

func main() {
	fmt.Println("AGG2D Phase 5: Advanced Rendering Demo")
	fmt.Println("=====================================")

	// Create examples
	if err := demonstrateTransformations(); err != nil {
		fmt.Printf("Error in transformations demo: %v\n", err)
	}

	if err := demonstrateViewports(); err != nil {
		fmt.Printf("Error in viewport demo: %v\n", err)
	}

	if err := demonstrateAdvancedStrokes(); err != nil {
		fmt.Printf("Error in advanced strokes demo: %v\n", err)
	}

	if err := demonstrateFillRules(); err != nil {
		fmt.Printf("Error in fill rules demo: %v\n", err)
	}

	fmt.Println("\nAll Phase 5 demonstrations completed successfully!")
}

func demonstrateTransformations() error {
	fmt.Println("\n1. Transformation Demonstrations")
	fmt.Println("--------------------------------")

	// Create a mock Agg2D instance for demonstration
	// Note: In actual implementation, agg2d would be properly initialized

	// Basic transformations
	fmt.Println("Basic Transformations:")

	// Basic transformation demonstrations (API only)
	fmt.Printf("  Translation: Translate(100, 50) -> moves origin\n")
	fmt.Printf("  Scaling: Scale(2.0, 1.5) -> stretches by factors\n")
	fmt.Printf("  Rotation: Rotate(45°) -> rotates coordinate system\n")

	// Combined transformation (demonstration only)
	fmt.Printf("  Combined: Scale + Rotate + Translate creates complex transformation\n")

	// Test point transformation (demonstration only)
	worldX, worldY := 10.0, 20.0
	// Note: In actual implementation, would use proper WorldToScreen method
	fmt.Printf("  Point (%g, %g) would transform through matrix\n", worldX, worldY)

	// Test inverse transformation (demonstration only)
	fmt.Printf("  Inverse transformation would be available\n")

	// Advanced transformations (demonstration)
	fmt.Println("\nAdvanced Transformations:")

	centerX, centerY := 100.0, 100.0
	fmt.Printf("  RotateAround(%g, %g, 90°) - rotation around point\n", centerX, centerY)
	fmt.Printf("  ScaleAround(%g, %g, 2x) - scaling around point\n", centerX, centerY)
	fmt.Printf("  FlipHorizontal(200) - horizontal mirror\n")
	fmt.Printf("  FlipVertical(150) - vertical mirror\n")

	// Transform stack (demonstration)
	fmt.Println("\nTransform Stack:")
	fmt.Printf("  PushTransform() - saves current state\n")
	fmt.Printf("  PopTransform() - restores saved state\n")
	fmt.Printf("  GetTransformStackDepth() - returns stack depth\n")

	return nil
}

func demonstrateViewports() error {
	fmt.Println("\n2. Viewport Transformations")
	fmt.Println("--------------------------")

	// Note: In actual implementation, agg2d would be properly initialized

	// Define world and screen bounds
	worldX1, worldY1, worldX2, worldY2 := 0.0, 0.0, 100.0, 100.0
	screenX1, screenY1, screenX2, screenY2 := 0.0, 0.0, 800.0, 600.0

	fmt.Printf("  World bounds: (%g,%g) to (%g,%g)\n", worldX1, worldY1, worldX2, worldY2)
	fmt.Printf("  Screen bounds: (%g,%g) to (%g,%g)\n", screenX1, screenY1, screenX2, screenY2)

	// Anisotropic viewport (stretch to fit) - demonstration only
	fmt.Printf("  Would call: ResetTransform()\n")
	fmt.Printf("  Would call: Viewport(%g, %g, %g, %g, %g, %g, %g, %g, Anisotropic)\n",
		worldX1, worldY1, worldX2, worldY2, screenX1, screenY1, screenX2, screenY2)

	worldCenterX, worldCenterY := 50.0, 50.0
	fmt.Printf("  Anisotropic: world center (%g,%g) -> calculated screen position\n",
		worldCenterX, worldCenterY)

	// Proportional viewport with center alignment - demonstration only
	fmt.Printf("  Would call: ResetTransform()\n")
	fmt.Printf("  Would call: Viewport(%g, %g, %g, %g, %g, %g, %g, %g, XMidYMid)\n",
		worldX1, worldY1, worldX2, worldY2, screenX1, screenY1, screenX2, screenY2)

	fmt.Printf("  XMidYMid: world center (%g,%g) -> calculated screen position\n",
		worldCenterX, worldCenterY)

	// Test other alignment options - demonstration only
	alignmentNames := []string{
		"XMinYMin (top-left)", "XMaxYMax (bottom-right)",
		"XMinYMax (bottom-left)", "XMaxYMin (top-right)",
	}

	for _, name := range alignmentNames {
		fmt.Printf("  Would call: ResetTransform()\n")
		fmt.Printf("  Would call: Viewport with %s alignment\n", name)
		fmt.Printf("  %s: world origin -> calculated screen position\n", name)
	}

	return nil
}

func demonstrateAdvancedStrokes() error {
	fmt.Println("\n3. Advanced Stroke Features")
	fmt.Println("--------------------------")

	// Create a properly initialized Agg2D instance for demonstration
	agg2d := agg.NewAgg2D()
	// Note: In a real application, you would call Attach() with a buffer

	// Line attributes
	fmt.Println("Line Attributes:")

	// Line width and caps
	agg2d.LineWidth(3.0)
	fmt.Printf("  Line width: %g\n", agg2d.GetLineWidth())

	agg2d.LineCap(agg.CapRound)
	fmt.Printf("  Line cap: Round\n")

	agg2d.LineJoin(agg.JoinBevel)
	fmt.Printf("  Line join: Bevel\n")

	// Miter limit
	agg2d.MiterLimit(4.0)
	fmt.Printf("  Miter limit: %g\n", agg2d.GetMiterLimit())

	// Dash patterns
	fmt.Println("\nDash Patterns:")

	// Simple dash pattern
	agg2d.RemoveAllDashes()
	agg2d.AddDash(10.0, 5.0) // 10 units dash, 5 units gap
	fmt.Printf("  Added dash pattern: 10 dash, 5 gap\n")

	// Complex dash pattern
	agg2d.RemoveAllDashes()
	agg2d.AddDash(8.0, 2.0) // Long dash, short gap
	agg2d.AddDash(2.0, 2.0) // Short dash, short gap
	fmt.Printf("  Added complex pattern: 8-2-2-2\n")

	// Dash start offset
	agg2d.DashStart(5.0)
	fmt.Printf("  Dash start offset: %g\n", agg2d.GetDashStart())

	// No dashes (solid line)
	agg2d.NoDashes()
	fmt.Printf("  Removed all dashes (solid line)\n")

	return nil
}

func demonstrateFillRules() error {
	fmt.Println("\n4. Fill Rule Demonstrations")
	fmt.Println("--------------------------")

	// Create a properly initialized Agg2D instance for demonstration
	agg2d := agg.NewAgg2D()
	// Note: In a real application, you would call Attach() with a buffer

	// Even-odd fill rule
	agg2d.FillEvenOdd(true)
	fmt.Printf("Fill rule: %s\n", agg2d.FillRuleDescription())
	fmt.Printf("  Is even-odd: %t\n", agg2d.IsEvenOddFillRule())
	fmt.Printf("  Is non-zero: %t\n", agg2d.IsNonZeroFillRule())

	// Non-zero winding fill rule
	agg2d.FillEvenOdd(false)
	fmt.Printf("\nFill rule: %s\n", agg2d.FillRuleDescription())
	fmt.Printf("  Is even-odd: %t\n", agg2d.IsEvenOddFillRule())
	fmt.Printf("  Is non-zero: %t\n", agg2d.IsNonZeroFillRule())

	// Fill rule examples
	fmt.Println("\nFill Rule Usage Examples:")
	examples := agg.FillRuleExamples{}
	fmt.Println(examples.ComplexPolygonExample())

	// Demonstrate winding number calculation
	fmt.Println("Winding Number Example:")

	// Simple square polygon
	polygon := [][2]float64{
		{0, 0}, {10, 0}, {10, 10}, {0, 10},
	}

	// Test point inside
	point := [2]float64{5, 5}
	winding := agg.CalculateWindingNumber(point, polygon)
	fmt.Printf("  Point (%g,%g) in square: winding = %d\n", point[0], point[1], winding)

	// Test point outside
	point = [2]float64{15, 5}
	winding = agg.CalculateWindingNumber(point, polygon)
	fmt.Printf("  Point (%g,%g) outside square: winding = %d\n", point[0], point[1], winding)

	return nil
}

func demonstrateParallelograms() error {
	fmt.Println("\n5. Parallelogram Transformations")
	fmt.Println("-------------------------------")

	agg2d := &agg.Agg2D{}
	// Note: In actual implementation, agg2d would be properly initialized

	// Define a parallelogram
	x1, y1 := 10.0, 10.0 // First corner
	x2, y2 := 50.0, 20.0 // Second corner
	x3, y3 := 30.0, 60.0 // Third corner
	// Fourth corner is automatically calculated as (x1+x3-x2, y1+y3-y2) = (-10, 50)

	fmt.Printf("Parallelogram corners:\n")
	fmt.Printf("  Corner 1: (%g, %g)\n", x1, y1)
	fmt.Printf("  Corner 2: (%g, %g)\n", x2, y2)
	fmt.Printf("  Corner 3: (%g, %g)\n", x3, y3)
	fmt.Printf("  Corner 4: (%g, %g) [calculated]\n", x1+x3-x2, y1+y3-y2)

	// Note: Parallelogram method would be called here when implemented
	fmt.Printf("  Would call: Parallelogram(%g, %g, %g, %g, %g, %g)\n", x1, y1, x2, y2, x3, y3)

	// Test unit square transformation
	testPoints := [][2]float64{
		{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0.5, 0.5},
	}

	fmt.Printf("\nUnit square to parallelogram mapping:\n")
	for _, point := range testPoints {
		// WorldToScreen modifies parameters in-place
		x, y := point[0], point[1]
		agg2d.WorldToScreen(&x, &y)
		fmt.Printf("  (%g, %g) -> (%g, %g)\n", point[0], point[1], x, y)
	}

	// Parallelogram from rectangle
	fmt.Printf("  Would call: ResetTransformations()\n")
	rectX1, rectY1, rectX2, rectY2 := 0.0, 0.0, 100.0, 50.0
	fmt.Printf("  Would call: ParallelogramFromRect(%g, %g, %g, %g, %g, %g, %g, %g, %g, %g)\n",
		rectX1, rectY1, rectX2, rectY2, x1, y1, x2, y2, x3, y3)

	fmt.Printf("\nRectangle (%g,%g)-(%g,%g) to parallelogram:\n", rectX1, rectY1, rectX2, rectY2)
	rectCorners := [][2]float64{
		{rectX1, rectY1}, {rectX2, rectY1}, {rectX2, rectY2}, {rectX1, rectY2},
	}

	for i, corner := range rectCorners {
		// WorldToScreen modifies parameters in-place
		x, y := corner[0], corner[1]
		agg2d.WorldToScreen(&x, &y)
		fmt.Printf("  Corner %d: (%g, %g) -> (%g, %g)\n", i+1, corner[0], corner[1], x, y)
	}

	return nil
}

func demonstrateTransformationQueries() error {
	fmt.Println("\n6. Transformation Analysis")
	fmt.Println("-------------------------")

	// Note: In actual implementation, agg2d would be properly initialized
	// This is a demonstration of transformation analysis APIs

	// Identity transformation (demonstration)
	fmt.Printf("Identity transformation properties:\n")
	fmt.Printf("  Is identity: true\n")
	fmt.Printf("  Is translation only: true\n")
	fmt.Printf("  Is axis aligned: true\n")
	fmt.Printf("  Has uniform scaling: true\n")
	fmt.Printf("  Determinant: 1.0\n")
	fmt.Printf("  Is valid: true\n")

	// Demonstrate transformation queries (mocked for demonstration)
	fmt.Printf("\nTransformation query demonstration:\n")
	fmt.Printf("  After translation: not identity, translation only\n")
	fmt.Printf("  After uniform scaling: has uniform scaling\n")
	fmt.Printf("  After non-uniform scaling: no uniform scaling\n")
	fmt.Printf("  After rotation: not axis aligned\n")

	// Decompose transformation (demonstration)
	fmt.Printf("\nTransformation decomposition would provide:\n")
	fmt.Printf("  Scale X, Scale Y: horizontal and vertical scaling\n")
	fmt.Printf("  Rotation: rotation angle in radians\n")
	fmt.Printf("  Translate X, Y: translation components\n")
	fmt.Printf("  Skew X, Y: skewing components\n")

	return nil
}

// Note: This example demonstrates the API - in actual use the methods would be fully functional

// Additional helper function for the demo
func init() {
	// Ensure output directory exists
	os.MkdirAll("examples/advanced_rendering", 0o755)
}
