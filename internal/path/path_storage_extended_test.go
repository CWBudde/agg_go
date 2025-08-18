package path

import (
	"testing"

	"agg_go/internal/basics"
)

// Tests for STL-based vertex storage
func TestVertexStlStorage(t *testing.T) {
	t.Run("BasicOperations", func(t *testing.T) {
		vss := NewVertexStlStorage[float64]()

		// Test initial state
		if vss.TotalVertices() != 0 {
			t.Errorf("Expected 0 vertices, got %d", vss.TotalVertices())
		}

		// Add vertices
		vss.AddVertex(10.0, 20.0, uint32(basics.PathCmdMoveTo))
		vss.AddVertex(30.0, 40.0, uint32(basics.PathCmdLineTo))

		if vss.TotalVertices() != 2 {
			t.Errorf("Expected 2 vertices, got %d", vss.TotalVertices())
		}

		// Check last vertex
		x, y, cmd := vss.LastVertex()
		if x != 30.0 || y != 40.0 || cmd != uint32(basics.PathCmdLineTo) {
			t.Errorf("Expected last vertex (30, 40, LineTo), got (%f, %f, %d)", x, y, cmd)
		}

		// Check specific vertices
		x, y, cmd = vss.Vertex(0)
		if x != 10.0 || y != 20.0 || cmd != uint32(basics.PathCmdMoveTo) {
			t.Errorf("Expected vertex 0 (10, 20, MoveTo), got (%f, %f, %d)", x, y, cmd)
		}

		x, y, cmd = vss.Vertex(1)
		if x != 30.0 || y != 40.0 || cmd != uint32(basics.PathCmdLineTo) {
			t.Errorf("Expected vertex 1 (30, 40, LineTo), got (%f, %f, %d)", x, y, cmd)
		}
	})

	t.Run("CapacityManagement", func(t *testing.T) {
		vss := NewVertexStlStorageWithCapacity[float64](100)

		// Test initial capacity
		if vss.Capacity() < 100 {
			t.Errorf("Expected capacity >= 100, got %d", vss.Capacity())
		}

		// Add some vertices
		for i := 0; i < 10; i++ {
			vss.AddVertex(float64(i), float64(i*2), uint32(basics.PathCmdLineTo))
		}

		// Test reserve
		initialCap := vss.Capacity()
		vss.Reserve(200)
		if vss.Capacity() < 200 {
			t.Errorf("Expected capacity >= 200 after reserve, got %d", vss.Capacity())
		}

		// Test shrink (capacity should reduce to match length)
		vss.Shrink()
		if vss.Capacity() != int(vss.TotalVertices()) {
			t.Errorf("Expected capacity to match vertex count after shrink, got cap=%d, vertices=%d",
				vss.Capacity(), vss.TotalVertices())
		}

		// Verify vertices are still intact after shrink
		for i := 0; i < 10; i++ {
			x, y, _ := vss.Vertex(uint(i))
			expectedX, expectedY := float64(i), float64(i*2)
			if x != expectedX || y != expectedY {
				t.Errorf("Vertex %d corrupted after shrink: expected (%f, %f), got (%f, %f)",
					i, expectedX, expectedY, x, y)
			}
		}

		// Use the initialCap to avoid unused variable error
		_ = initialCap
	})

	t.Run("EdgeCases", func(t *testing.T) {
		vss := NewVertexStlStorage[float64]()

		// Test operations on empty storage
		x, y, cmd := vss.LastVertex()
		if x != 0.0 || y != 0.0 || cmd != 0 {
			t.Errorf("Expected empty storage to return zeros, got (%f, %f, %d)", x, y, cmd)
		}

		x, y, cmd = vss.PrevVertex()
		if x != 0.0 || y != 0.0 || cmd != 0 {
			t.Errorf("Expected empty storage PrevVertex to return zeros, got (%f, %f, %d)", x, y, cmd)
		}

		// Test out-of-bounds access
		x, y, cmd = vss.Vertex(999)
		if x != 0.0 || y != 0.0 || cmd != 0 {
			t.Errorf("Expected out-of-bounds access to return zeros, got (%f, %f, %d)", x, y, cmd)
		}
	})
}

func TestPathStorageStl(t *testing.T) {
	t.Run("BasicPathOperations", func(t *testing.T) {
		path := NewPathStorageStl()

		// Create a simple triangle
		path.MoveTo(0.0, 0.0)
		path.LineTo(10.0, 0.0)
		path.LineTo(5.0, 10.0)
		path.ClosePolygon(basics.PathFlagsNone)

		if path.TotalVertices() != 4 { // MoveTo, LineTo, LineTo, EndPoly
			t.Errorf("Expected 4 vertices, got %d", path.TotalVertices())
		}

		// Test path iteration
		path.Rewind(0)
		x, y, cmd := path.NextVertex()
		if x != 0.0 || y != 0.0 || !basics.IsMoveTo(basics.PathCommand(cmd)) {
			t.Errorf("First vertex should be MoveTo(0,0), got (%f, %f, %d)", x, y, cmd)
		}

		x, y, cmd = path.NextVertex()
		if x != 10.0 || y != 0.0 || !basics.IsLineTo(basics.PathCommand(cmd)) {
			t.Errorf("Second vertex should be LineTo(10,0), got (%f, %f, %d)", x, y, cmd)
		}
	})

	t.Run("WithInitialCapacity", func(t *testing.T) {
		path := NewPathStorageStlWithCapacity(1000)

		// Add many vertices to test capacity
		for i := 0; i < 500; i++ {
			if i == 0 {
				path.MoveTo(float64(i), float64(i))
			} else {
				path.LineTo(float64(i), float64(i))
			}
		}

		if path.TotalVertices() != 500 {
			t.Errorf("Expected 500 vertices, got %d", path.TotalVertices())
		}
	})
}

func TestLargePathsCrossingBlockBoundaries(t *testing.T) {
	t.Run("BlockStorage", func(t *testing.T) {
		path := NewPathStorage()

		// Add more than 256 vertices to cross block boundaries
		const numVertices = 300
		for i := 0; i < numVertices; i++ {
			if i == 0 {
				path.MoveTo(float64(i), float64(i))
			} else {
				path.LineTo(float64(i), float64(i))
			}
		}

		if path.TotalVertices() != numVertices {
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
					i, float64(i), float64(i), expectedCmd, x, y, cmd)
			}
		}
	})

	t.Run("StlStorage", func(t *testing.T) {
		path := NewPathStorageStl()

		// Add many vertices to test slice growth
		const numVertices = 1000
		for i := 0; i < numVertices; i++ {
			if i == 0 {
				path.MoveTo(float64(i), float64(i))
			} else {
				path.LineTo(float64(i), float64(i))
			}
		}

		if path.TotalVertices() != numVertices {
			t.Errorf("Expected %d vertices, got %d", numVertices, path.TotalVertices())
		}

		// Test path iteration with large path
		path.Rewind(0)
		count := 0
		for {
			_, _, cmd := path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			count++
		}

		if count != numVertices {
			t.Errorf("Expected to iterate over %d vertices, got %d", numVertices, count)
		}
	})
}

func TestComplexCurveOperations(t *testing.T) {
	t.Run("BezierCurves", func(t *testing.T) {
		path := NewPathStorage()

		path.MoveTo(0.0, 0.0)
		path.Curve4(10.0, 0.0, 10.0, 10.0, 0.0, 10.0) // Cubic Bezier
		path.Curve3(5.0, 15.0, 0.0, 20.0)             // Quadratic Bezier

		// Should have vertices for move, curve4 (3 vertices), curve3 (2 vertices)
		expectedVertices := 1 + 3 + 2
		if path.TotalVertices() != uint(expectedVertices) {
			t.Errorf("Expected %d vertices for curve path, got %d", expectedVertices, path.TotalVertices())
		}
	})

	t.Run("ArcOperations", func(t *testing.T) {
		path := NewPathStorage()

		path.MoveTo(0.0, 0.0)
		path.ArcTo(10.0, 10.0, 0.0, false, false, 20.0, 0.0)

		// Arc should generate multiple vertices
		if path.TotalVertices() < 2 {
			t.Errorf("Arc should generate multiple vertices, got %d", path.TotalVertices())
		}
	})
}

func TestPolyAdaptorsExtended(t *testing.T) {
	t.Run("PlainAdaptor", func(t *testing.T) {
		// Create coordinate data: triangle
		coords := []float64{0.0, 0.0, 10.0, 0.0, 5.0, 10.0}
		adaptor := NewPolyPlainAdaptorWithData(coords, 3, true)

		vertices := make([]VertexD, 0)
		adaptor.Rewind(0)

		for {
			x, y, cmd := adaptor.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			vertices = append(vertices, NewVertexD(x, y, cmd))
		}

		// Should have 4 vertices: 3 triangle points + close command
		if len(vertices) != 4 {
			t.Errorf("Expected 4 vertices from closed triangle, got %d", len(vertices))
		}

		// Check first vertex is MoveTo
		if !basics.IsMoveTo(basics.PathCommand(vertices[0].Cmd)) {
			t.Errorf("First vertex should be MoveTo, got %d", vertices[0].Cmd)
		}

		// Check remaining vertices are LineTo
		for i := 1; i < 3; i++ {
			if !basics.IsLineTo(basics.PathCommand(vertices[i].Cmd)) {
				t.Errorf("Vertex %d should be LineTo, got %d", i, vertices[i].Cmd)
			}
		}
	})

	t.Run("LineAdaptor", func(t *testing.T) {
		adaptor := NewLineAdaptor()
		adaptor.Init(0.0, 0.0, 10.0, 10.0)

		vertices := make([]VertexD, 0)
		adaptor.Rewind(0)

		for {
			x, y, cmd := adaptor.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			vertices = append(vertices, NewVertexD(x, y, cmd))
		}

		// Should have exactly 2 vertices
		if len(vertices) != 2 {
			t.Errorf("Expected 2 vertices from line, got %d", len(vertices))
		}

		// Check coordinates
		if vertices[0].X != 0.0 || vertices[0].Y != 0.0 {
			t.Errorf("First vertex should be (0,0), got (%f,%f)", vertices[0].X, vertices[0].Y)
		}

		if vertices[1].X != 10.0 || vertices[1].Y != 10.0 {
			t.Errorf("Second vertex should be (10,10), got (%f,%f)", vertices[1].X, vertices[1].Y)
		}
	})
}

func TestF32Precision(t *testing.T) {
	t.Run("BlockStorageF32", func(t *testing.T) {
		path := NewPathStorageF32()

		// Add vertices with precision that would matter for float32 vs float64
		path.MoveTo(0.123456789, 0.987654321)
		path.LineTo(1.123456789, 1.987654321)

		x, y, _ := path.LastVertex()
		// Should be rounded to float32 precision
		expectedX := float64(float32(1.123456789))
		expectedY := float64(float32(1.987654321))

		if x != expectedX || y != expectedY {
			t.Errorf("F32 storage should limit precision: expected (%f, %f), got (%f, %f)",
				expectedX, expectedY, x, y)
		}
	})

	t.Run("StlStorageF32", func(t *testing.T) {
		path := NewPathStorageStlF32()

		// Similar test for STL storage with F32
		path.MoveTo(0.123456789, 0.987654321)
		path.LineTo(1.123456789, 1.987654321)

		if path.TotalVertices() != 2 {
			t.Errorf("Expected 2 vertices, got %d", path.TotalVertices())
		}
	})
}

func TestStorageIntegration(t *testing.T) {
	t.Run("CopyBetweenStorageTypes", func(t *testing.T) {
		// Create path with block storage
		blockPath := NewPathStorage()
		blockPath.MoveTo(0.0, 0.0)
		blockPath.LineTo(10.0, 10.0)
		blockPath.LineTo(20.0, 0.0)
		blockPath.ClosePolygon(basics.PathFlagsNone)

		// Create STL storage and copy vertices
		stlPath := NewPathStorageStl()
		for i := uint(0); i < blockPath.TotalVertices(); i++ {
			x, y, cmd := blockPath.Vertex(i)
			// Use the underlying storage's AddVertex method
			stlPath.vertices.AddVertex(x, y, cmd)
		}

		// Verify they have same vertices
		if blockPath.TotalVertices() != stlPath.TotalVertices() {
			t.Errorf("Vertex count mismatch: block=%d, stl=%d",
				blockPath.TotalVertices(), stlPath.TotalVertices())
		}

		for i := uint(0); i < blockPath.TotalVertices(); i++ {
			bx, by, bcmd := blockPath.Vertex(i)
			sx, sy, scmd := stlPath.Vertex(i)
			if bx != sx || by != sy || bcmd != scmd {
				t.Errorf("Vertex %d mismatch: block=(%f,%f,%d), stl=(%f,%f,%d)",
					i, bx, by, bcmd, sx, sy, scmd)
			}
		}
	})

	t.Run("MemoryEfficiency", func(t *testing.T) {
		// This test is mainly to ensure both storage types work with large datasets
		blockPath := NewPathStorage()
		stlPath := NewPathStorageStl()

		// Add a substantial number of vertices
		const numVertices = 2000
		for i := 0; i < numVertices; i++ {
			x, y := float64(i%100), float64((i/100)%100)
			if i == 0 || i%100 == 0 {
				blockPath.MoveTo(x, y)
				stlPath.MoveTo(x, y)
			} else {
				blockPath.LineTo(x, y)
				stlPath.LineTo(x, y)
			}
		}

		// Verify both have same data
		if blockPath.TotalVertices() != stlPath.TotalVertices() {
			t.Errorf("Large dataset vertex count mismatch: block=%d, stl=%d",
				blockPath.TotalVertices(), stlPath.TotalVertices())
		}

		// Spot check some vertices
		for i := 0; i < 10; i++ {
			idx := uint(i * 200) // Check every 200th vertex
			if idx >= blockPath.TotalVertices() {
				break
			}

			bx, by, bcmd := blockPath.Vertex(idx)
			sx, sy, scmd := stlPath.Vertex(idx)
			if bx != sx || by != sy || bcmd != scmd {
				t.Errorf("Large dataset vertex %d mismatch: block=(%f,%f,%d), stl=(%f,%f,%d)",
					idx, bx, by, bcmd, sx, sy, scmd)
			}
		}
	})
}

// Benchmark tests for storage performance comparison
func BenchmarkStorageComparison(b *testing.B) {
	b.Run("BlockStorage/AddVertices", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			path := NewPathStorage()
			for j := 0; j < 1000; j++ {
				path.LineTo(float64(j), float64(j))
			}
		}
	})

	b.Run("StlStorage/AddVertices", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			path := NewPathStorageStl()
			for j := 0; j < 1000; j++ {
				path.LineTo(float64(j), float64(j))
			}
		}
	})

	b.Run("StlStorageWithCapacity/AddVertices", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			path := NewPathStorageStlWithCapacity(1000)
			for j := 0; j < 1000; j++ {
				path.LineTo(float64(j), float64(j))
			}
		}
	})

	b.Run("BlockStorage/Iteration", func(b *testing.B) {
		path := NewPathStorage()
		for i := 0; i < 1000; i++ {
			path.LineTo(float64(i), float64(i))
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			path.Rewind(0)
			count := 0
			for {
				_, _, cmd := path.NextVertex()
				if basics.IsStop(basics.PathCommand(cmd)) {
					break
				}
				count++
			}
		}
	})

	b.Run("StlStorage/Iteration", func(b *testing.B) {
		path := NewPathStorageStl()
		for i := 0; i < 1000; i++ {
			path.LineTo(float64(i), float64(i))
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			path.Rewind(0)
			count := 0
			for {
				_, _, cmd := path.NextVertex()
				if basics.IsStop(basics.PathCommand(cmd)) {
					break
				}
				count++
			}
		}
	})

	b.Run("BlockStorage/RandomAccess", func(b *testing.B) {
		path := NewPathStorage()
		for i := 0; i < 1000; i++ {
			path.LineTo(float64(i), float64(i))
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for j := 0; j < 100; j++ {
				idx := uint(j * 10) // Access every 10th vertex
				_, _, _ = path.Vertex(idx)
			}
		}
	})

	b.Run("StlStorage/RandomAccess", func(b *testing.B) {
		path := NewPathStorageStl()
		for i := 0; i < 1000; i++ {
			path.LineTo(float64(i), float64(i))
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for j := 0; j < 100; j++ {
				idx := uint(j * 10) // Access every 10th vertex
				_, _, _ = path.Vertex(idx)
			}
		}
	})
}
