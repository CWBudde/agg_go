package vcgen

import (
	"math"
	"testing"

	"agg_go/internal/array"
	"agg_go/internal/basics"
)

func TestVCGenVertexSequence_Basic(t *testing.T) {
	gen := NewVCGenVertexSequence()

	// Add a simple path
	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	gen.AddVertex(10, 0, basics.PathCmdLineTo)
	gen.AddVertex(10, 10, basics.PathCmdLineTo)
	gen.AddVertex(0, 10, basics.PathCmdLineTo)

	gen.Rewind(0)

	// Check first vertex
	x, y, cmd := gen.Vertex()
	if x != 0 || y != 0 || cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected (0,0,MoveTo), got (%f,%f,%v)", x, y, cmd)
	}

	// Check second vertex
	x, y, cmd = gen.Vertex()
	if x != 10 || y != 0 || cmd != basics.PathCmdLineTo {
		t.Errorf("Expected (10,0,LineTo), got (%f,%f,%v)", x, y, cmd)
	}

	// Check third vertex
	x, y, cmd = gen.Vertex()
	if x != 10 || y != 10 || cmd != basics.PathCmdLineTo {
		t.Errorf("Expected (10,10,LineTo), got (%f,%f,%v)", x, y, cmd)
	}

	// Check fourth vertex
	x, y, cmd = gen.Vertex()
	if x != 0 || y != 10 || cmd != basics.PathCmdLineTo {
		t.Errorf("Expected (0,10,LineTo), got (%f,%f,%v)", x, y, cmd)
	}

	// Should reach end
	x, y, cmd = gen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected PathCmdStop, got %v", cmd)
	}
}

func TestVCGenVertexSequence_MoveTo(t *testing.T) {
	gen := NewVCGenVertexSequence()

	// MoveTo should modify the last vertex (starting with empty)
	gen.AddVertex(5, 5, basics.PathCmdMoveTo)
	gen.AddVertex(10, 10, basics.PathCmdMoveTo) // This should replace the previous
	gen.AddVertex(20, 20, basics.PathCmdLineTo)

	gen.Rewind(0)

	// Should start with the final MoveTo position
	x, y, cmd := gen.Vertex()
	if x != 10 || y != 10 || cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected (10,10,MoveTo), got (%f,%f,%v)", x, y, cmd)
	}

	x, y, cmd = gen.Vertex()
	if x != 20 || y != 20 || cmd != basics.PathCmdLineTo {
		t.Errorf("Expected (20,20,LineTo), got (%f,%f,%v)", x, y, cmd)
	}
}

func TestVCGenVertexSequence_PathFlags(t *testing.T) {
	gen := NewVCGenVertexSequence()

	// Add path with close flag
	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	gen.AddVertex(10, 0, basics.PathCmdLineTo)
	gen.AddVertex(0, 10, basics.PathCmdLineTo)
	gen.AddVertex(0, 0, basics.PathCmdEndPoly|basics.PathFlagClose) // Close the path

	gen.Rewind(0)

	// The sequence should be processed with close flag
	vertexCount := 0
	for {
		_, _, cmd := gen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertexCount++
	}

	// Should have received all vertices
	if vertexCount == 0 {
		t.Error("Should have received some vertices from closed path")
	}
}

func TestVCGenVertexSequence_Shortening(t *testing.T) {
	gen := NewVCGenVertexSequence()
	gen.SetShorten(5.0) // Shorten by 5 units total (2.5 from each end)

	// Create a straight line path longer than shortening amount
	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	gen.AddVertex(20, 0, basics.PathCmdLineTo) // 20 unit long line

	gen.Rewind(0)

	// Get all vertices
	vertices := make([]array.VertexDistCmd, 0)
	for {
		x, y, cmd := gen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertices = append(vertices, array.VertexDistCmd{X: x, Y: y, Cmd: cmd})
	}

	// Should still have vertices (path is longer than shortening)
	if len(vertices) == 0 {
		t.Error("Shortened path should still have vertices")
	}

	// First vertex should be moved inward
	if len(vertices) > 0 && vertices[0].X <= 0 {
		t.Error("First vertex should be shortened from start")
	}
}

func TestVCGenVertexSequence_TooMuchShortening(t *testing.T) {
	gen := NewVCGenVertexSequence()
	gen.SetShorten(50.0) // Shorten more than path length

	// Create a short path
	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	gen.AddVertex(10, 0, basics.PathCmdLineTo) // 10 unit long line

	gen.Rewind(0)

	// Should have no vertices (path completely shortened)
	x, y, cmd := gen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Over-shortened path should return Stop immediately, got %v at (%f,%f)", cmd, x, y)
	}
}

func TestVCGenVertexSequence_RemoveAll(t *testing.T) {
	gen := NewVCGenVertexSequence()

	// Add some vertices
	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	gen.AddVertex(10, 10, basics.PathCmdLineTo)

	// Remove all
	gen.RemoveAll()

	gen.Rewind(0)

	// Should be empty
	x, y, cmd := gen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("After RemoveAll, should return Stop, got %v at (%f,%f)", cmd, x, y)
	}
}

func TestVCGenVertexSequence_MultipleRewinds(t *testing.T) {
	gen := NewVCGenVertexSequence()

	gen.AddVertex(5, 5, basics.PathCmdMoveTo)
	gen.AddVertex(15, 15, basics.PathCmdLineTo)

	// First iteration
	gen.Rewind(0)
	x1, y1, cmd1 := gen.Vertex()

	// Second iteration should produce same results
	gen.Rewind(0)
	x2, y2, cmd2 := gen.Vertex()

	if x1 != x2 || y1 != y2 || cmd1 != cmd2 {
		t.Errorf("Multiple rewinds should produce same results: (%f,%f,%v) vs (%f,%f,%v)",
			x1, y1, cmd1, x2, y2, cmd2)
	}
}

func TestVCGenVertexSequence_EmptySequence(t *testing.T) {
	gen := NewVCGenVertexSequence()

	gen.Rewind(0)
	x, y, cmd := gen.Vertex()

	if cmd != basics.PathCmdStop {
		t.Errorf("Empty sequence should return Stop, got %v at (%f,%f)", cmd, x, y)
	}
}

func TestVCGenVertexSequence_SingleVertex(t *testing.T) {
	gen := NewVCGenVertexSequence()

	gen.AddVertex(100, 200, basics.PathCmdMoveTo)

	gen.Rewind(0)

	// Should get the single vertex
	x, y, cmd := gen.Vertex()
	if x != 100 || y != 200 || cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected (100,200,MoveTo), got (%f,%f,%v)", x, y, cmd)
	}

	// Then stop
	x, y, cmd = gen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("After single vertex, should return Stop, got %v", cmd)
	}
}

func TestVCGenVertexSequence_PrepareSrcCaching(t *testing.T) {
	gen := NewVCGenVertexSequence()
	gen.SetShorten(1.0)

	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	gen.AddVertex(10, 0, basics.PathCmdLineTo)

	// First rewind should prepare source
	gen.Rewind(0)
	gen.Vertex() // This triggers PrepareSrc internally

	// Get ready state (should be true after first access)
	if !gen.ready {
		t.Error("Generator should be ready after first vertex access")
	}

	// Second rewind should use cached preparation
	gen.Rewind(0)
	firstX, firstY, firstCmd := gen.Vertex()

	gen.Rewind(0)
	secondX, secondY, secondCmd := gen.Vertex()

	// Results should be identical (cached)
	if firstX != secondX || firstY != secondY || firstCmd != secondCmd {
		t.Errorf("Cached preparation should give same results: (%f,%f,%v) vs (%f,%f,%v)",
			firstX, firstY, firstCmd, secondX, secondY, secondCmd)
	}
}

func TestVCGenVertexSequence_NonVertexCommands(t *testing.T) {
	gen := NewVCGenVertexSequence()

	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	gen.AddVertex(10, 10, basics.PathCmdLineTo)
	// Add non-vertex command (should be stored as flags)
	gen.AddVertex(0, 0, basics.PathCmdEndPoly|basics.PathFlagClose)

	gen.Rewind(0)

	// Should still process vertex commands normally
	_, _, cmd := gen.Vertex()
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("Non-vertex commands shouldn't affect vertex processing, got %v", cmd)
	}
}

func TestVCGenVertexSequence_ShortenValue(t *testing.T) {
	gen := NewVCGenVertexSequence()

	// Test default shorten value
	if gen.Shorten() != 0.0 {
		t.Errorf("Default shorten should be 0.0, got %f", gen.Shorten())
	}

	// Test setting and getting shorten value
	gen.SetShorten(2.5)
	if gen.Shorten() != 2.5 {
		t.Errorf("Shorten should be 2.5, got %f", gen.Shorten())
	}

	// Setting shorten should mark as not ready
	gen.ready = true
	gen.SetShorten(3.0)
	if gen.ready {
		t.Error("Setting shorten should mark generator as not ready")
	}
}

// Test edge cases and error conditions
func TestVCGenVertexSequence_EdgeCases(t *testing.T) {
	gen := NewVCGenVertexSequence()

	// Test with very small coordinates
	gen.AddVertex(1e-10, 1e-10, basics.PathCmdMoveTo)
	gen.AddVertex(2e-10, 2e-10, basics.PathCmdLineTo)

	gen.Rewind(0)
	_, _, cmd := gen.Vertex()
	if cmd != basics.PathCmdMoveTo {
		t.Error("Should handle very small coordinates")
	}

	// Test with very large coordinates
	gen.RemoveAll()
	gen.AddVertex(1e10, 1e10, basics.PathCmdMoveTo)
	gen.AddVertex(2e10, 2e10, basics.PathCmdLineTo)

	gen.Rewind(0)
	_, _, cmd = gen.Vertex()
	if cmd != basics.PathCmdMoveTo {
		t.Error("Should handle very large coordinates")
	}

	// Test with NaN coordinates (should not crash)
	gen.RemoveAll()
	gen.AddVertex(math.NaN(), math.NaN(), basics.PathCmdMoveTo)
	gen.AddVertex(1, 1, basics.PathCmdLineTo)

	gen.Rewind(0)
	// Should not crash even with NaN
	gen.Vertex()
}

// Benchmark tests
func BenchmarkVCGenVertexSequence_AddVertices(b *testing.B) {
	gen := NewVCGenVertexSequence()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.RemoveAll()
		for j := 0; j < 100; j++ {
			cmd := basics.PathCmdLineTo
			if j == 0 {
				cmd = basics.PathCmdMoveTo
			}
			gen.AddVertex(float64(j), float64(j*2), cmd)
		}
	}
}

func BenchmarkVCGenVertexSequence_IterateVertices(b *testing.B) {
	gen := NewVCGenVertexSequence()

	// Setup path
	for i := 0; i < 100; i++ {
		cmd := basics.PathCmdLineTo
		if i == 0 {
			cmd = basics.PathCmdMoveTo
		}
		gen.AddVertex(float64(i), float64(i*2), cmd)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Rewind(0)
		for {
			_, _, cmd := gen.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
		}
	}
}

func BenchmarkVCGenVertexSequence_WithShortening(b *testing.B) {
	gen := NewVCGenVertexSequence()
	gen.SetShorten(5.0)

	// Setup longer path for shortening
	for i := 0; i < 100; i++ {
		cmd := basics.PathCmdLineTo
		if i == 0 {
			cmd = basics.PathCmdMoveTo
		}
		gen.AddVertex(float64(i*10), float64(i*10), cmd)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Rewind(0)
		for {
			_, _, cmd := gen.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
		}
	}
}
