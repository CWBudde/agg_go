package curves

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

func TestCurveApproximationMethod(t *testing.T) {
	if CurveInc != 0 {
		t.Error("Expected CurveInc to be 0")
	}
	if CurveDiv != 1 {
		t.Error("Expected CurveDiv to be 1")
	}
}

func TestCurve3Inc(t *testing.T) {
	t.Run("Basic functionality", func(t *testing.T) {
		curve := NewCurve3Inc()

		// Test defaults
		if curve.ApproximationMethod() != CurveInc {
			t.Error("Expected approximation method to be CurveInc")
		}
		if curve.ApproximationScale() != 1.0 {
			t.Error("Expected default scale to be 1.0")
		}
		if curve.AngleTolerance() != 0.0 {
			t.Error("Expected angle tolerance to be 0.0 for incremental curves")
		}
		if curve.CuspLimit() != 0.0 {
			t.Error("Expected cusp limit to be 0.0 for incremental curves")
		}

		// Test initialization
		curve.Init(0, 0, 50, 100, 100, 0)

		// Test vertex iteration
		curve.Rewind(0)

		// First vertex should be MoveTo start point
		x, y, cmd := curve.Vertex()
		if cmd != basics.PathCmdMoveTo {
			t.Errorf("Expected first vertex to be MoveTo, got %v", cmd)
		}
		if x != 0 || y != 0 {
			t.Errorf("Expected first vertex at (0,0), got (%f,%f)", x, y)
		}

		// Iterate through curve
		vertexCount := 1
		var lastX, lastY float64
		for {
			x, y, cmd = curve.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
			if cmd != basics.PathCmdLineTo {
				t.Errorf("Expected LineTo command, got %v", cmd)
			}
			lastX, lastY = x, y
			vertexCount++
		}

		// Should have generated multiple vertices
		if vertexCount < 4 {
			t.Errorf("Expected at least 4 vertices, got %d", vertexCount)
		}

		// Last vertex should be at end point (100, 0)
		if lastX != 100 || lastY != 0 {
			t.Errorf("Expected last vertex at (100,0), got (%f,%f)", lastX, lastY)
		}
	})

	t.Run("With constructor points", func(t *testing.T) {
		curve := NewCurve3IncWithPoints(0, 0, 50, 100, 100, 0)
		curve.Rewind(0)

		x, y, cmd := curve.Vertex()
		if cmd != basics.PathCmdMoveTo || x != 0 || y != 0 {
			t.Error("Constructor with points failed")
		}
	})

	t.Run("Scale setting", func(t *testing.T) {
		curve := NewCurve3Inc()
		curve.SetApproximationScale(2.0)

		if curve.ApproximationScale() != 2.0 {
			t.Error("Failed to set approximation scale")
		}

		// Higher scale should produce more vertices
		curve.Init(0, 0, 50, 100, 100, 0)
		curve.Rewind(0)

		vertexCount := 0
		for {
			_, _, cmd := curve.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
			vertexCount++
		}

		// Should have more vertices with higher scale
		if vertexCount < 4 {
			t.Errorf("Expected more vertices with higher scale, got %d", vertexCount)
		}
	})
}

func TestCurve3Div(t *testing.T) {
	t.Run("Basic functionality", func(t *testing.T) {
		curve := NewCurve3Div()

		// Test defaults
		if curve.ApproximationMethod() != CurveDiv {
			t.Error("Expected approximation method to be CurveDiv")
		}
		if curve.ApproximationScale() != 1.0 {
			t.Error("Expected default scale to be 1.0")
		}
		if curve.AngleTolerance() != 0.0 {
			t.Error("Expected default angle tolerance to be 0.0")
		}

		// Test initialization
		curve.Init(0, 0, 50, 100, 100, 0)

		// Test vertex iteration
		curve.Rewind(0)

		// First vertex should be MoveTo start point
		x, y, cmd := curve.Vertex()
		if cmd != basics.PathCmdMoveTo {
			t.Errorf("Expected first vertex to be MoveTo, got %v", cmd)
		}
		if x != 0 || y != 0 {
			t.Errorf("Expected first vertex at (0,0), got (%f,%f)", x, y)
		}

		// Iterate through curve
		vertexCount := 1
		var lastX, lastY float64
		for {
			x, y, cmd = curve.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
			if cmd != basics.PathCmdLineTo {
				t.Errorf("Expected LineTo command, got %v", cmd)
			}
			lastX, lastY = x, y
			vertexCount++
		}

		// Should have generated vertices
		if vertexCount < 3 {
			t.Errorf("Expected at least 3 vertices, got %d", vertexCount)
		}

		// Last vertex should be at end point (100, 0)
		if lastX != 100 || lastY != 0 {
			t.Errorf("Expected last vertex at (100,0), got (%f,%f)", lastX, lastY)
		}
	})

	t.Run("Angle tolerance", func(t *testing.T) {
		curve := NewCurve3Div()
		curve.SetAngleTolerance(0.1)

		if curve.AngleTolerance() != 0.1 {
			t.Error("Failed to set angle tolerance")
		}

		// Test with angle tolerance
		curve.Init(0, 0, 50, 100, 100, 0)
		curve.Rewind(0)

		vertexCount := 0
		for {
			_, _, cmd := curve.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
			vertexCount++
		}

		if vertexCount < 2 {
			t.Error("Expected vertices with angle tolerance")
		}
	})
}

func TestCurve4Inc(t *testing.T) {
	t.Run("Basic functionality", func(t *testing.T) {
		curve := NewCurve4Inc()

		// Test cubic curve
		curve.Init(0, 0, 33, 100, 66, 100, 100, 0)
		curve.Rewind(0)

		// First vertex should be MoveTo start point
		x, y, cmd := curve.Vertex()
		if cmd != basics.PathCmdMoveTo {
			t.Errorf("Expected first vertex to be MoveTo, got %v", cmd)
		}
		if x != 0 || y != 0 {
			t.Errorf("Expected first vertex at (0,0), got (%f,%f)", x, y)
		}

		// Iterate through curve
		vertexCount := 1
		var lastX, lastY float64
		for {
			x, y, cmd = curve.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
			lastX, lastY = x, y
			vertexCount++
		}

		// Should have generated multiple vertices
		if vertexCount < 4 {
			t.Errorf("Expected at least 4 vertices, got %d", vertexCount)
		}

		// Last vertex should be at end point (100, 0)
		if lastX != 100 || lastY != 0 {
			t.Errorf("Expected last vertex at (100,0), got (%f,%f)", lastX, lastY)
		}
	})

	t.Run("With constructor points", func(t *testing.T) {
		curve := NewCurve4IncWithPoints(0, 0, 33, 100, 66, 100, 100, 0)
		curve.Rewind(0)

		x, y, cmd := curve.Vertex()
		if cmd != basics.PathCmdMoveTo || x != 0 || y != 0 {
			t.Error("Constructor with points failed")
		}
	})
}

func TestCurve4Div(t *testing.T) {
	t.Run("Basic functionality", func(t *testing.T) {
		curve := NewCurve4Div()

		// Test cubic curve
		curve.Init(0, 0, 33, 100, 66, 100, 100, 0)
		curve.Rewind(0)

		// First vertex should be MoveTo start point
		x, y, cmd := curve.Vertex()
		if cmd != basics.PathCmdMoveTo {
			t.Errorf("Expected first vertex to be MoveTo, got %v", cmd)
		}
		if x != 0 || y != 0 {
			t.Errorf("Expected first vertex at (0,0), got (%f,%f)", x, y)
		}

		// Iterate through curve
		vertexCount := 1
		var lastX, lastY float64
		for {
			x, y, cmd = curve.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
			lastX, lastY = x, y
			vertexCount++
		}

		// Should have generated vertices
		if vertexCount < 3 {
			t.Errorf("Expected at least 3 vertices, got %d", vertexCount)
		}

		// Last vertex should be at end point (100, 0)
		if lastX != 100 || lastY != 0 {
			t.Errorf("Expected last vertex at (100,0), got (%f,%f)", lastX, lastY)
		}
	})

	t.Run("Cusp limit", func(t *testing.T) {
		curve := NewCurve4Div()
		curve.SetCuspLimit(0.1)

		if math.Abs(curve.CuspLimit()-0.1) > 1e-10 {
			t.Error("Failed to set cusp limit")
		}
	})
}

func TestCurve4Points(t *testing.T) {
	t.Run("Basic functionality", func(t *testing.T) {
		cp := NewCurve4Points(0, 0, 33, 100, 66, 100, 100, 0)

		// Test accessors
		if cp.At(0) != 0 || cp.At(1) != 0 {
			t.Error("Failed to get first point")
		}
		if cp.At(6) != 100 || cp.At(7) != 0 {
			t.Error("Failed to get last point")
		}

		// Test setters
		cp.Set(2, 50)
		if cp.At(2) != 50 {
			t.Error("Failed to set control point")
		}

		// Test Init
		cp.Init(10, 10, 20, 20, 30, 30, 40, 40)
		if cp.At(0) != 10 || cp.At(1) != 10 {
			t.Error("Failed to init control points")
		}
	})
}

func TestConversionFunctions(t *testing.T) {
	t.Run("Catrom to Bezier", func(t *testing.T) {
		// Test Catmull-Rom to Bezier conversion
		bezier := CatromToBezier(0, 0, 0, 100, 100, 100, 100, 0)

		// First point should be second control point
		if bezier.At(0) != 0 || bezier.At(1) != 100 {
			t.Errorf("Expected first point (0,100), got (%f,%f)", bezier.At(0), bezier.At(1))
		}

		// Last point should be third control point
		if bezier.At(6) != 100 || bezier.At(7) != 100 {
			t.Errorf("Expected last point (100,100), got (%f,%f)", bezier.At(6), bezier.At(7))
		}
	})

	t.Run("Catrom to Bezier with points", func(t *testing.T) {
		cp := NewCurve4Points(0, 0, 0, 100, 100, 100, 100, 0)
		bezier := CatromToBezierPoints(cp)

		if bezier.At(0) != 0 || bezier.At(1) != 100 {
			t.Error("Points conversion failed")
		}
	})

	t.Run("UBSpline to Bezier", func(t *testing.T) {
		bezier := UBSplineToBezier(0, 0, 50, 100, 100, 100, 150, 0)

		// Should produce valid Bezier control points
		if bezier.At(0) < 0 || bezier.At(0) > 150 {
			t.Error("UBSpline conversion produced invalid first point")
		}
	})

	t.Run("UBSpline to Bezier with points", func(t *testing.T) {
		cp := NewCurve4Points(0, 0, 50, 100, 100, 100, 150, 0)
		bezier := UBSplineToBezierPoints(cp)

		if bezier.At(0) < 0 || bezier.At(0) > 150 {
			t.Error("UBSpline points conversion failed")
		}
	})

	t.Run("Hermite to Bezier", func(t *testing.T) {
		bezier := HermiteToBezier(0, 0, 100, 0, 100, 0, 0, 0)

		// First point should be first control point
		if bezier.At(0) != 0 || bezier.At(1) != 0 {
			t.Error("Hermite conversion failed for first point")
		}

		// Last point should be second control point
		if bezier.At(6) != 100 || bezier.At(7) != 0 {
			t.Error("Hermite conversion failed for last point")
		}
	})

	t.Run("Hermite to Bezier with points", func(t *testing.T) {
		cp := NewCurve4Points(0, 0, 100, 0, 100, 0, 0, 0)
		bezier := HermiteToBezierPoints(cp)

		if bezier.At(0) != 0 || bezier.At(1) != 0 {
			t.Error("Hermite points conversion failed")
		}
	})
}

func TestCurve3UnifiedInterface(t *testing.T) {
	t.Run("Default method", func(t *testing.T) {
		curve := NewCurve3()

		// Should default to CurveDiv
		if curve.ApproximationMethod() != CurveDiv {
			t.Error("Expected default method to be CurveDiv")
		}

		curve.Init(0, 0, 50, 100, 100, 0)
		curve.Rewind(0)

		x, y, cmd := curve.Vertex()
		if cmd != basics.PathCmdMoveTo || x != 0 || y != 0 {
			t.Error("Failed to get first vertex from unified interface")
		}
	})

	t.Run("Switch methods", func(t *testing.T) {
		curve := NewCurve3()
		curve.Init(0, 0, 50, 100, 100, 0)

		// Test with Div method
		curve.SetApproximationMethod(CurveDiv)
		curve.Rewind(0)

		divVertexCount := 0
		for {
			_, _, cmd := curve.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
			divVertexCount++
		}

		// Test with Inc method
		curve.SetApproximationMethod(CurveInc)
		curve.Init(0, 0, 50, 100, 100, 0) // Need to reinitialize with new method
		curve.Rewind(0)

		incVertexCount := 0
		for {
			_, _, cmd := curve.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
			incVertexCount++
		}

		// Both should produce vertices
		if divVertexCount < 2 || incVertexCount < 2 {
			t.Error("Both approximation methods should produce vertices")
		}
	})

	t.Run("With constructor points", func(t *testing.T) {
		curve := NewCurve3WithPoints(0, 0, 50, 100, 100, 0)
		curve.Rewind(0)

		x, y, cmd := curve.Vertex()
		if cmd != basics.PathCmdMoveTo || x != 0 || y != 0 {
			t.Error("Constructor with points failed")
		}
	})
}

func TestCurve4UnifiedInterface(t *testing.T) {
	t.Run("Default method", func(t *testing.T) {
		curve := NewCurve4()

		// Should default to CurveDiv
		if curve.ApproximationMethod() != CurveDiv {
			t.Error("Expected default method to be CurveDiv")
		}

		curve.Init(0, 0, 33, 100, 66, 100, 100, 0)
		curve.Rewind(0)

		x, y, cmd := curve.Vertex()
		if cmd != basics.PathCmdMoveTo || x != 0 || y != 0 {
			t.Error("Failed to get first vertex from unified interface")
		}
	})

	t.Run("With constructor points", func(t *testing.T) {
		curve := NewCurve4WithPoints(0, 0, 33, 100, 66, 100, 100, 0)
		curve.Rewind(0)

		x, y, cmd := curve.Vertex()
		if cmd != basics.PathCmdMoveTo || x != 0 || y != 0 {
			t.Error("Constructor with points failed")
		}
	})

	t.Run("With control points", func(t *testing.T) {
		cp := NewCurve4Points(0, 0, 33, 100, 66, 100, 100, 0)
		curve := NewCurve4WithControlPoints(cp)
		curve.Rewind(0)

		x, y, cmd := curve.Vertex()
		if cmd != basics.PathCmdMoveTo || x != 0 || y != 0 {
			t.Error("Constructor with control points failed")
		}

		// Test InitWithControlPoints
		newCP := NewCurve4Points(10, 10, 20, 20, 30, 30, 40, 40)
		curve.InitWithControlPoints(newCP)
		curve.Rewind(0)

		x, y, cmd = curve.Vertex()
		if cmd != basics.PathCmdMoveTo || x != 10 || y != 10 {
			t.Error("InitWithControlPoints failed")
		}
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("Straight line quadratic", func(t *testing.T) {
		// Test with collinear control points (straight line)
		curve := NewCurve3Div()
		curve.Init(0, 0, 50, 0, 100, 0)
		curve.Rewind(0)

		vertices := []basics.Point[float64]{}
		for {
			x, y, cmd := curve.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
			vertices = append(vertices, basics.Point[float64]{X: x, Y: y})
		}

		// Should still produce start and end points
		if len(vertices) < 2 {
			t.Error("Straight line should produce at least 2 vertices")
		}

		// First and last should be correct
		if vertices[0].X != 0 || vertices[0].Y != 0 {
			t.Error("First vertex incorrect for straight line")
		}
		if vertices[len(vertices)-1].X != 100 || vertices[len(vertices)-1].Y != 0 {
			t.Error("Last vertex incorrect for straight line")
		}
	})

	t.Run("Straight line cubic", func(t *testing.T) {
		// Test with collinear control points (straight line)
		curve := NewCurve4Div()
		curve.Init(0, 0, 33, 0, 66, 0, 100, 0)
		curve.Rewind(0)

		vertices := []basics.Point[float64]{}
		for {
			x, y, cmd := curve.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
			vertices = append(vertices, basics.Point[float64]{X: x, Y: y})
		}

		// Should still produce start and end points
		if len(vertices) < 2 {
			t.Error("Straight line should produce at least 2 vertices")
		}

		if vertices[0].X != 0 || vertices[0].Y != 0 {
			t.Error("First vertex incorrect for straight line")
		}
		if vertices[len(vertices)-1].X != 100 || vertices[len(vertices)-1].Y != 0 {
			t.Error("Last vertex incorrect for straight line")
		}
	})

	t.Run("Zero length curve", func(t *testing.T) {
		// Test with all points at same location
		curve := NewCurve3Inc()
		curve.Init(50, 50, 50, 50, 50, 50)
		curve.Rewind(0)

		x, y, cmd := curve.Vertex()
		if cmd != basics.PathCmdMoveTo || x != 50 || y != 50 {
			t.Error("Zero length curve should start correctly")
		}

		// Should still produce end point
		_, _, cmd = curve.Vertex()
		if cmd != basics.PathCmdLineTo {
			t.Error("Zero length curve should produce LineTo")
		}
	})
}

func TestReset(t *testing.T) {
	t.Run("Curve3 reset", func(t *testing.T) {
		curve := NewCurve3()
		curve.Init(0, 0, 50, 100, 100, 0)
		curve.Rewind(0)

		// Get first vertex
		curve.Vertex()

		// Reset and reinitialize
		curve.Reset()
		curve.Init(0, 0, 50, 100, 100, 0) // Need to reinitialize after reset
		curve.Rewind(0)

		// Should be able to iterate again
		x, y, cmd := curve.Vertex()
		if cmd != basics.PathCmdMoveTo || x != 0 || y != 0 {
			t.Error("Reset failed for Curve3")
		}
	})

	t.Run("Curve4 reset", func(t *testing.T) {
		curve := NewCurve4()
		curve.Init(0, 0, 33, 100, 66, 100, 100, 0)
		curve.Rewind(0)

		// Get first vertex
		curve.Vertex()

		// Reset and reinitialize
		curve.Reset()
		curve.Init(0, 0, 33, 100, 66, 100, 100, 0) // Need to reinitialize after reset
		curve.Rewind(0)

		// Should be able to iterate again
		x, y, cmd := curve.Vertex()
		if cmd != basics.PathCmdMoveTo || x != 0 || y != 0 {
			t.Error("Reset failed for Curve4")
		}
	})
}
