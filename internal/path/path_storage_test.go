package path

import (
	"testing"

	"agg_go/internal/basics"
)

func TestVertexBlockStorage(t *testing.T) {
	t.Run("BasicOperations", func(t *testing.T) {
		vbs := NewVertexBlockStorage[float64]()

		// Test initial state
		if vbs.TotalVertices() != 0 {
			t.Errorf("Expected 0 vertices, got %d", vbs.TotalVertices())
		}

		// Add vertices
		vbs.AddVertex(10.0, 20.0, uint32(basics.PathCmdMoveTo))
		vbs.AddVertex(30.0, 40.0, uint32(basics.PathCmdLineTo))

		if vbs.TotalVertices() != 2 {
			t.Errorf("Expected 2 vertices, got %d", vbs.TotalVertices())
		}

		// Check last vertex
		x, y, cmd := vbs.LastVertex()
		if x != 30.0 || y != 40.0 || cmd != uint32(basics.PathCmdLineTo) {
			t.Errorf("Expected last vertex (30, 40, LineTo), got (%f, %f, %d)", x, y, cmd)
		}

		// Check specific vertices
		x, y, cmd = vbs.Vertex(0)
		if x != 10.0 || y != 20.0 || cmd != uint32(basics.PathCmdMoveTo) {
			t.Errorf("Expected vertex 0 (10, 20, MoveTo), got (%f, %f, %d)", x, y, cmd)
		}

		x, y, cmd = vbs.Vertex(1)
		if x != 30.0 || y != 40.0 || cmd != uint32(basics.PathCmdLineTo) {
			t.Errorf("Expected vertex 1 (30, 40, LineTo), got (%f, %f, %d)", x, y, cmd)
		}
	})

	t.Run("ModifyOperations", func(t *testing.T) {
		vbs := NewVertexBlockStorage[float64]()
		vbs.AddVertex(10.0, 20.0, uint32(basics.PathCmdMoveTo))
		vbs.AddVertex(30.0, 40.0, uint32(basics.PathCmdLineTo))

		// Modify vertex coordinates
		vbs.ModifyVertex(0, 15.0, 25.0)
		x, y, cmd := vbs.Vertex(0)
		if x != 15.0 || y != 25.0 || cmd != uint32(basics.PathCmdMoveTo) {
			t.Errorf("Expected modified vertex (15, 25, MoveTo), got (%f, %f, %d)", x, y, cmd)
		}

		// Modify command
		vbs.ModifyCommand(1, uint32(basics.PathCmdCurve3))
		cmd = vbs.Command(1)
		if cmd != uint32(basics.PathCmdCurve3) {
			t.Errorf("Expected command Curve3, got %d", cmd)
		}

		// Modify vertex and command
		vbs.ModifyVertexAndCommand(1, 50.0, 60.0, uint32(basics.PathCmdCurve4))
		x, y, cmd = vbs.Vertex(1)
		if x != 50.0 || y != 60.0 || cmd != uint32(basics.PathCmdCurve4) {
			t.Errorf("Expected modified vertex (50, 60, Curve4), got (%f, %f, %d)", x, y, cmd)
		}
	})

	t.Run("SwapVertices", func(t *testing.T) {
		vbs := NewVertexBlockStorage[float64]()
		vbs.AddVertex(10.0, 20.0, uint32(basics.PathCmdMoveTo))
		vbs.AddVertex(30.0, 40.0, uint32(basics.PathCmdLineTo))

		vbs.SwapVertices(0, 1)

		x, y, cmd := vbs.Vertex(0)
		if x != 30.0 || y != 40.0 || cmd != uint32(basics.PathCmdLineTo) {
			t.Errorf("Expected swapped vertex 0 (30, 40, LineTo), got (%f, %f, %d)", x, y, cmd)
		}

		x, y, cmd = vbs.Vertex(1)
		if x != 10.0 || y != 20.0 || cmd != uint32(basics.PathCmdMoveTo) {
			t.Errorf("Expected swapped vertex 1 (10, 20, MoveTo), got (%f, %f, %d)", x, y, cmd)
		}
	})

	t.Run("BlockAllocation", func(t *testing.T) {
		// Test with small block size to trigger multiple blocks
		vbs := NewVertexBlockStorageWithParams[float64](2, 4) // 4 vertices per block

		// Add more vertices than fit in one block
		for i := 0; i < 10; i++ {
			vbs.AddVertex(float64(i), float64(i*2), uint32(basics.PathCmdLineTo))
		}

		if vbs.TotalVertices() != 10 {
			t.Errorf("Expected 10 vertices, got %d", vbs.TotalVertices())
		}

		// Verify all vertices are stored correctly
		for i := 0; i < 10; i++ {
			x, y, cmd := vbs.Vertex(uint(i))
			expectedX := float64(i)
			expectedY := float64(i * 2)
			if x != expectedX || y != expectedY || cmd != uint32(basics.PathCmdLineTo) {
				t.Errorf("Vertex %d: expected (%f, %f, LineTo), got (%f, %f, %d)",
					i, expectedX, expectedY, x, y, cmd)
			}
		}
	})

	t.Run("CopyConstructor", func(t *testing.T) {
		vbs1 := NewVertexBlockStorage[float64]()
		vbs1.AddVertex(10.0, 20.0, uint32(basics.PathCmdMoveTo))
		vbs1.AddVertex(30.0, 40.0, uint32(basics.PathCmdLineTo))

		vbs2 := NewVertexBlockStorageFromCopy(vbs1)

		if vbs2.TotalVertices() != vbs1.TotalVertices() {
			t.Errorf("Expected %d vertices in copy, got %d", vbs1.TotalVertices(), vbs2.TotalVertices())
		}

		for i := uint(0); i < vbs1.TotalVertices(); i++ {
			x1, y1, cmd1 := vbs1.Vertex(i)
			x2, y2, cmd2 := vbs2.Vertex(i)
			if x1 != x2 || y1 != y2 || cmd1 != cmd2 {
				t.Errorf("Vertex %d mismatch: original (%f, %f, %d), copy (%f, %f, %d)",
					i, x1, y1, cmd1, x2, y2, cmd2)
			}
		}
	})

	t.Run("RemoveAllAndFreeAll", func(t *testing.T) {
		vbs := NewVertexBlockStorage[float64]()
		vbs.AddVertex(10.0, 20.0, uint32(basics.PathCmdMoveTo))
		vbs.AddVertex(30.0, 40.0, uint32(basics.PathCmdLineTo))

		vbs.RemoveAll()
		if vbs.TotalVertices() != 0 {
			t.Errorf("Expected 0 vertices after RemoveAll, got %d", vbs.TotalVertices())
		}

		// Add vertices again to test that storage still works
		vbs.AddVertex(50.0, 60.0, uint32(basics.PathCmdMoveTo))
		if vbs.TotalVertices() != 1 {
			t.Errorf("Expected 1 vertex after adding again, got %d", vbs.TotalVertices())
		}

		vbs.FreeAll()
		if vbs.TotalVertices() != 0 {
			t.Errorf("Expected 0 vertices after FreeAll, got %d", vbs.TotalVertices())
		}
	})
}

func TestPathBase(t *testing.T) {
	t.Run("BasicPathOperations", func(t *testing.T) {
		path := NewPathStorage()

		// Test move_to and line_to
		path.MoveTo(10.0, 20.0)
		path.LineTo(30.0, 40.0)
		path.LineTo(50.0, 60.0)

		if path.TotalVertices() != 3 {
			t.Errorf("Expected 3 vertices, got %d", path.TotalVertices())
		}

		// Check vertices
		x, y, cmd := path.Vertex(0)
		if x != 10.0 || y != 20.0 || cmd != uint32(basics.PathCmdMoveTo) {
			t.Errorf("Expected vertex 0 (10, 20, MoveTo), got (%f, %f, %d)", x, y, cmd)
		}

		x, y, cmd = path.Vertex(1)
		if x != 30.0 || y != 40.0 || cmd != uint32(basics.PathCmdLineTo) {
			t.Errorf("Expected vertex 1 (30, 40, LineTo), got (%f, %f, %d)", x, y, cmd)
		}
	})

	t.Run("RelativeCommands", func(t *testing.T) {
		path := NewPathStorage()

		path.MoveTo(10.0, 20.0)
		path.LineRel(5.0, 10.0) // Should result in (15, 30)
		path.HLineRel(10.0)     // Should result in (25, 30)
		path.VLineRel(5.0)      // Should result in (25, 35)

		// Check results
		x, y, _ := path.Vertex(1)
		if x != 15.0 || y != 30.0 {
			t.Errorf("Expected LineRel result (15, 30), got (%f, %f)", x, y)
		}

		x, y, _ = path.Vertex(2)
		if x != 25.0 || y != 30.0 {
			t.Errorf("Expected HLineRel result (25, 30), got (%f, %f)", x, y)
		}

		x, y, _ = path.Vertex(3)
		if x != 25.0 || y != 35.0 {
			t.Errorf("Expected VLineRel result (25, 35), got (%f, %f)", x, y)
		}
	})

	t.Run("CurveCommands", func(t *testing.T) {
		path := NewPathStorage()

		path.MoveTo(0.0, 0.0)
		path.Curve3(10.0, 10.0, 20.0, 0.0) // Quadratic curve

		if path.TotalVertices() != 3 {
			t.Errorf("Expected 3 vertices, got %d", path.TotalVertices())
		}

		// Check curve vertices
		x, y, cmd := path.Vertex(1)
		if x != 10.0 || y != 10.0 || cmd != uint32(basics.PathCmdCurve3) {
			t.Errorf("Expected curve3 control (10, 10, Curve3), got (%f, %f, %d)", x, y, cmd)
		}

		x, y, cmd = path.Vertex(2)
		if x != 20.0 || y != 0.0 || cmd != uint32(basics.PathCmdCurve3) {
			t.Errorf("Expected curve3 end (20, 0, Curve3), got (%f, %f, %d)", x, y, cmd)
		}

		// Test cubic curve
		path.RemoveAll()
		path.MoveTo(0.0, 0.0)
		path.Curve4(5.0, 10.0, 15.0, 10.0, 20.0, 0.0)

		if path.TotalVertices() != 4 {
			t.Errorf("Expected 4 vertices for cubic curve, got %d", path.TotalVertices())
		}
	})

	t.Run("PolygonOperations", func(t *testing.T) {
		path := NewPathStorage()

		path.MoveTo(0.0, 0.0)
		path.LineTo(10.0, 0.0)
		path.LineTo(10.0, 10.0)
		path.LineTo(0.0, 10.0)
		path.ClosePolygon(basics.PathFlagsNone)

		// Check end_poly command
		lastCmd := path.vertices.LastCommand()
		if !basics.IsEndPoly(basics.PathCommand(lastCmd)) {
			t.Errorf("Expected EndPoly command, got %d", lastCmd)
		}

		if !basics.IsClose(lastCmd) {
			t.Errorf("Expected Close flag to be set")
		}
	})

	t.Run("PathTransformation", func(t *testing.T) {
		path := NewPathStorage()

		path.MoveTo(10.0, 20.0)
		path.LineTo(30.0, 40.0)

		// Test translation
		path.TranslateAllPaths(5.0, 10.0)

		x, y, _ := path.Vertex(0)
		if x != 15.0 || y != 30.0 {
			t.Errorf("Expected translated vertex 0 (15, 30), got (%f, %f)", x, y)
		}

		x, y, _ = path.Vertex(1)
		if x != 35.0 || y != 50.0 {
			t.Errorf("Expected translated vertex 1 (35, 50), got (%f, %f)", x, y)
		}

		// Test flip
		path.RemoveAll()
		path.MoveTo(10.0, 20.0)
		path.LineTo(30.0, 40.0)

		path.FlipX(0.0, 100.0) // Flip horizontally between 0 and 100
		x, y, _ = path.Vertex(0)
		if x != 90.0 || y != 20.0 { // 100 - 10 + 0 = 90
			t.Errorf("Expected flipped vertex 0 (90, 20), got (%f, %f)", x, y)
		}
	})

	t.Run("StartNewPath", func(t *testing.T) {
		path := NewPathStorage()

		path.MoveTo(10.0, 20.0)
		path.LineTo(30.0, 40.0)

		pathID := path.StartNewPath()
		if pathID != 3 { // Should be after the stop command
			t.Errorf("Expected path ID 3, got %d", pathID)
		}

		// Check that stop command was added
		_, _, cmd := path.Vertex(2)
		if cmd != uint32(basics.PathCmdStop) {
			t.Errorf("Expected Stop command, got %d", cmd)
		}
	})

	t.Run("VertexSourceInterface", func(t *testing.T) {
		path := NewPathStorage()

		path.MoveTo(10.0, 20.0)
		path.LineTo(30.0, 40.0)

		// Test VertexSource interface
		path.Rewind(0)

		x, y, cmd := path.NextVertex()
		if x != 10.0 || y != 20.0 || cmd != uint32(basics.PathCmdMoveTo) {
			t.Errorf("Expected first vertex (10, 20, MoveTo), got (%f, %f, %d)", x, y, cmd)
		}

		x, y, cmd = path.NextVertex()
		if x != 30.0 || y != 40.0 || cmd != uint32(basics.PathCmdLineTo) {
			t.Errorf("Expected second vertex (30, 40, LineTo), got (%f, %f, %d)", x, y, cmd)
		}

		x, y, cmd = path.NextVertex()
		if cmd != uint32(basics.PathCmdStop) {
			t.Errorf("Expected Stop command, got %d", cmd)
		}
	})
}

func TestPolyAdaptors(t *testing.T) {
	t.Run("PolyPlainAdaptor", func(t *testing.T) {
		// Test data: square
		data := []float64{0.0, 0.0, 10.0, 0.0, 10.0, 10.0, 0.0, 10.0}
		adaptor := NewPolyPlainAdaptorWithData(data, 4, true) // closed polygon

		adaptor.Rewind(0)

		// First vertex should be MoveTo
		x, y, cmd := adaptor.NextVertex()
		if x != 0.0 || y != 0.0 || cmd != uint32(basics.PathCmdMoveTo) {
			t.Errorf("Expected first vertex (0, 0, MoveTo), got (%f, %f, %d)", x, y, cmd)
		}

		// Remaining vertices should be LineTo
		for i := 1; i < 4; i++ {
			x, y, cmd := adaptor.NextVertex()
			expectedX := data[i*2]
			expectedY := data[i*2+1]
			if x != expectedX || y != expectedY || cmd != uint32(basics.PathCmdLineTo) {
				t.Errorf("Expected vertex %d (%f, %f, LineTo), got (%f, %f, %d)",
					i, expectedX, expectedY, x, y, cmd)
			}
		}

		// Should get EndPoly with Close flag
		x, y, cmd = adaptor.NextVertex()
		if cmd != uint32(basics.PathCmdEndPoly)|uint32(basics.PathFlagsClose) {
			t.Errorf("Expected EndPoly|Close, got %d", cmd)
		}

		// Should get Stop
		x, y, cmd = adaptor.NextVertex()
		if cmd != uint32(basics.PathCmdStop) {
			t.Errorf("Expected Stop, got %d", cmd)
		}
	})

	t.Run("LineAdaptor", func(t *testing.T) {
		line := NewLineAdaptorWithCoords(10.0, 20.0, 30.0, 40.0)

		line.Rewind(0)

		// First vertex should be MoveTo
		x, y, cmd := line.NextVertex()
		if x != 10.0 || y != 20.0 || cmd != uint32(basics.PathCmdMoveTo) {
			t.Errorf("Expected start point (10, 20, MoveTo), got (%f, %f, %d)", x, y, cmd)
		}

		// Second vertex should be LineTo
		x, y, cmd = line.NextVertex()
		if x != 30.0 || y != 40.0 || cmd != uint32(basics.PathCmdLineTo) {
			t.Errorf("Expected end point (30, 40, LineTo), got (%f, %f, %d)", x, y, cmd)
		}

		// Should get Stop
		x, y, cmd = line.NextVertex()
		if cmd != uint32(basics.PathCmdStop) {
			t.Errorf("Expected Stop, got %d", cmd)
		}
	})

	t.Run("PolyContainerAdaptor", func(t *testing.T) {
		// Create simple vertex container
		vertices := SimpleVertexContainer{
			NewSimpleVertex(0.0, 0.0),
			NewSimpleVertex(10.0, 0.0),
			NewSimpleVertex(10.0, 10.0),
		}

		adaptor := NewPolyContainerAdaptorWithData(vertices, false)
		adaptor.Rewind(0)

		// First vertex should be MoveTo
		x, y, cmd := adaptor.NextVertex()
		if x != 0.0 || y != 0.0 || cmd != uint32(basics.PathCmdMoveTo) {
			t.Errorf("Expected first vertex (0, 0, MoveTo), got (%f, %f, %d)", x, y, cmd)
		}

		// Check remaining vertices
		for i := 1; i < vertices.Size(); i++ {
			x, y, cmd := adaptor.NextVertex()
			v := vertices.At(i)
			if x != v.GetX() || y != v.GetY() || cmd != uint32(basics.PathCmdLineTo) {
				t.Errorf("Expected vertex %d (%f, %f, LineTo), got (%f, %f, %d)",
					i, v.GetX(), v.GetY(), x, y, cmd)
			}
		}

		// Should get Stop (not closed)
		x, y, cmd = adaptor.NextVertex()
		if cmd != uint32(basics.PathCmdStop) {
			t.Errorf("Expected Stop, got %d", cmd)
		}
	})
}

func TestConcatAndJoinPath(t *testing.T) {
	t.Run("ConcatPath", func(t *testing.T) {
		path1 := NewPathStorage()
		path1.MoveTo(0.0, 0.0)
		path1.LineTo(10.0, 10.0)

		path2 := NewPathStorage()
		path2.MoveTo(20.0, 20.0)
		path2.LineTo(30.0, 30.0)

		// Concatenate path2 to path1
		path1.ConcatPath(path2, 0)

		if path1.TotalVertices() != 4 {
			t.Errorf("Expected 4 vertices after concat, got %d", path1.TotalVertices())
		}

		// Check that MoveTo command is preserved
		x, y, cmd := path1.Vertex(2)
		if x != 20.0 || y != 20.0 || cmd != uint32(basics.PathCmdMoveTo) {
			t.Errorf("Expected concatenated MoveTo (20, 20), got (%f, %f, %d)", x, y, cmd)
		}
	})

	t.Run("JoinPath", func(t *testing.T) {
		path1 := NewPathStorage()
		path1.MoveTo(0.0, 0.0)
		path1.LineTo(10.0, 10.0)

		path2 := NewPathStorage()
		path2.MoveTo(20.0, 20.0) // This should become LineTo when joined
		path2.LineTo(30.0, 30.0)

		// Join path2 to path1
		path1.JoinPath(path2)

		if path1.TotalVertices() != 4 {
			t.Errorf("Expected 4 vertices after join, got %d", path1.TotalVertices())
		}

		// Check that MoveTo was converted to LineTo
		x, y, cmd := path1.Vertex(2)
		if x != 20.0 || y != 20.0 || cmd != uint32(basics.PathCmdLineTo) {
			t.Errorf("Expected joined LineTo (20, 20), got (%f, %f, %d)", x, y, cmd)
		}
	})

	t.Run("ConcatPoly", func(t *testing.T) {
		path := NewPathStorage()
		path.MoveTo(0.0, 0.0)

		// Concatenate a triangle
		triangle := []float64{10.0, 0.0, 5.0, 8.66, 0.0, 0.0}
		path.ConcatPoly(triangle, 3, true)

		// Should have 1 + 3 + 1 = 5 vertices (original MoveTo + 3 triangle vertices + EndPoly)
		if path.TotalVertices() != 5 {
			t.Errorf("Expected 5 vertices after ConcatPoly, got %d", path.TotalVertices())
		}

		// Check that EndPoly was added
		lastCmd := path.vertices.LastCommand()
		if !basics.IsEndPoly(basics.PathCommand(lastCmd)) {
			t.Errorf("Expected EndPoly command, got %d", lastCmd)
		}
	})
}

func TestPathMath(t *testing.T) {
	t.Run("DistanceCalculation", func(t *testing.T) {
		// Test that arc rejection works based on distance
		path := NewPathStorage()
		path.MoveTo(0.0, 0.0)

		// Arc with identical endpoints should be ignored
		path.ArcTo(10.0, 10.0, 0.0, false, false, 0.0, 0.0)

		// Should still have only the MoveTo
		if path.TotalVertices() != 1 {
			t.Errorf("Expected 1 vertex (arc with identical endpoints should be ignored), got %d", path.TotalVertices())
		}

		// Arc with very small radii should become a line
		path.ArcTo(1e-31, 1e-31, 0.0, false, false, 10.0, 10.0)

		// Should have LineTo command added
		if path.TotalVertices() != 2 {
			t.Errorf("Expected 2 vertices (small radii arc becomes line), got %d", path.TotalVertices())
		}

		x, y, cmd := path.Vertex(1)
		if x != 10.0 || y != 10.0 || cmd != uint32(basics.PathCmdLineTo) {
			t.Errorf("Expected LineTo (10, 10), got (%f, %f, %d)", x, y, cmd)
		}
	})

	t.Run("SmoothCurves", func(t *testing.T) {
		path := NewPathStorage()

		// Create a path with curves to test smooth curve functionality
		path.MoveTo(0.0, 0.0)
		path.Curve3(10.0, 10.0, 20.0, 0.0)
		path.Curve3Smooth(40.0, 0.0) // Should reflect control point

		if path.TotalVertices() != 5 { // MoveTo + 2 Curve3 + 2 Curve3
			t.Errorf("Expected 5 vertices, got %d", path.TotalVertices())
		}

		// Test smooth curve4
		path.RemoveAll()
		path.MoveTo(0.0, 0.0)
		path.Curve4(5.0, 10.0, 15.0, 10.0, 20.0, 0.0)
		path.Curve4Smooth(35.0, 10.0, 40.0, 0.0)

		if path.TotalVertices() != 7 { // MoveTo + 3 Curve4 + 3 Curve4
			t.Errorf("Expected 7 vertices, got %d", path.TotalVertices())
		}
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("EmptyPath", func(t *testing.T) {
		path := NewPathStorage()

		// Test operations on empty path
		if path.LastX() != 0.0 || path.LastY() != 0.0 {
			t.Errorf("Expected (0, 0) for empty path last position, got (%f, %f)", path.LastX(), path.LastY())
		}

		_, _, cmd := path.LastVertex()
		if cmd != uint32(basics.PathCmdStop) {
			t.Errorf("Expected Stop command for empty path, got %d", cmd)
		}

		// Test vertex source on empty path
		path.Rewind(0)
		_, _, cmd = path.NextVertex()
		if cmd != uint32(basics.PathCmdStop) {
			t.Errorf("Expected Stop command from empty path vertex source, got %d", cmd)
		}
	})

	t.Run("RelativeFromEmpty", func(t *testing.T) {
		path := NewPathStorage()

		// Relative commands from empty path should treat (0,0) as origin
		path.MoveRel(10.0, 20.0)
		x, y, _ := path.Vertex(0)
		if x != 10.0 || y != 20.0 {
			t.Errorf("Expected MoveRel from empty (10, 20), got (%f, %f)", x, y)
		}

		path.LineRel(5.0, 10.0)
		x, y, _ = path.Vertex(1)
		if x != 15.0 || y != 30.0 {
			t.Errorf("Expected LineRel result (15, 30), got (%f, %f)", x, y)
		}
	})

	t.Run("VeryLargePath", func(t *testing.T) {
		path := NewPathStorage()

		// Create a path with many vertices to test memory management
		numVertices := 1000
		for i := 0; i < numVertices; i++ {
			if i == 0 {
				path.MoveTo(float64(i), float64(i))
			} else {
				path.LineTo(float64(i), float64(i))
			}
		}

		if int(path.TotalVertices()) != numVertices {
			t.Errorf("Expected %d vertices, got %d", numVertices, path.TotalVertices())
		}

		// Verify all vertices are correct
		for i := 0; i < numVertices; i++ {
			x, y, cmd := path.Vertex(uint(i))
			expectedCmd := basics.PathCmdLineTo
			if i == 0 {
				expectedCmd = basics.PathCmdMoveTo
			}

			if x != float64(i) || y != float64(i) || cmd != uint32(expectedCmd) {
				t.Errorf("Vertex %d: expected (%f, %f, %d), got (%f, %f, %d)",
					i, float64(i), float64(i), uint32(expectedCmd), x, y, cmd)
			}
		}
	})
}

func BenchmarkVertexBlockStorage(b *testing.B) {
	vbs := NewVertexBlockStorage[float64]()

	b.Run("AddVertex", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			vbs.AddVertex(float64(i), float64(i*2), uint32(basics.PathCmdLineTo))
		}
	})

	b.Run("AccessVertex", func(b *testing.B) {
		// First add some vertices
		for i := 0; i < 1000; i++ {
			vbs.AddVertex(float64(i), float64(i*2), uint32(basics.PathCmdLineTo))
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, _ = vbs.Vertex(uint(i % 1000))
		}
	})
}

func BenchmarkPathOperations(b *testing.B) {
	b.Run("PathCreation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			path := NewPathStorage()
			path.MoveTo(0.0, 0.0)
			path.LineTo(10.0, 10.0)
			path.LineTo(20.0, 0.0)
			path.ClosePolygon(basics.PathFlagsNone)
		}
	})

	b.Run("PathIteration", func(b *testing.B) {
		path := NewPathStorage()
		for i := 0; i < 1000; i++ {
			if i == 0 {
				path.MoveTo(float64(i), float64(i))
			} else {
				path.LineTo(float64(i), float64(i))
			}
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			path.Rewind(0)
			for {
				_, _, cmd := path.NextVertex()
				if basics.IsStop(basics.PathCommand(cmd)) {
					break
				}
			}
		}
	})
}
