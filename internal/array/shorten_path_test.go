package array

import (
	"testing"
)

// TestShortenPath tests the ShortenPath function
func TestShortenPath(t *testing.T) {
	// Create a simple path with 3 vertices
	vs := NewVertexSequence[VertexDist]()
	vs.Add(VertexDist{X: 0, Y: 0, Dist: 10.0})  // Distance to next: 10
	vs.Add(VertexDist{X: 10, Y: 0, Dist: 10.0}) // Distance to next: 10
	vs.Add(VertexDist{X: 20, Y: 0, Dist: 0.0})  // End vertex

	// Shorten by 5 units
	ShortenPath(vs, 5.0, false)

	// Should adjust the last vertex position
	if vs.Size() != 3 {
		t.Errorf("Expected 3 vertices after shortening, got %d", vs.Size())
	}

	// Last vertex should be moved back 5 units along the line
	last := vs.Get(vs.Size() - 1)
	expectedX := 15.0 // 20 - 5
	if last.X != expectedX {
		t.Errorf("Expected last vertex X to be %f, got %f", expectedX, last.X)
	}
}

// TestShortenPathRemoveVertices tests shortening that removes entire vertices
func TestShortenPathRemoveVertices(t *testing.T) {
	vs := NewVertexSequence[VertexDist]()
	vs.Add(VertexDist{X: 0, Y: 0, Dist: 10.0})  // Distance to next: 10
	vs.Add(VertexDist{X: 10, Y: 0, Dist: 10.0}) // Distance to next: 10
	vs.Add(VertexDist{X: 20, Y: 0, Dist: 0.0})  // End vertex

	// Shorten by 15 units (more than last segment distance)
	ShortenPath(vs, 15.0, false)

	// Should remove the last vertex and adjust the second-to-last
	if vs.Size() != 2 {
		t.Errorf("Expected 2 vertices after shortening by 15 units, got %d", vs.Size())
	}

	// Second vertex should be moved back 5 units (15 - 10 from removed vertex)
	if vs.Size() >= 2 {
		second := vs.Get(vs.Size() - 1)
		expectedX := 5.0 // 10 - 5
		if second.X != expectedX {
			t.Errorf("Expected second vertex X to be %f, got %f", expectedX, second.X)
		}
	}
}

// TestShortenPathRemoveAll tests shortening that removes all vertices
func TestShortenPathRemoveAll(t *testing.T) {
	vs := NewVertexSequence[VertexDist]()
	vs.Add(VertexDist{X: 0, Y: 0, Dist: 10.0})
	vs.Add(VertexDist{X: 10, Y: 0, Dist: 0.0})

	// Shorten by more than total path length
	ShortenPath(vs, 25.0, false)

	// Should remove all vertices
	if vs.Size() != 0 {
		t.Errorf("Expected 0 vertices after shortening by 25 units, got %d", vs.Size())
	}
}

// TestShortenPathNoShortening tests shortening by 0
func TestShortenPathNoShortening(t *testing.T) {
	vs := NewVertexSequence[VertexDist]()
	vs.Add(VertexDist{X: 0, Y: 0, Dist: 10.0})
	vs.Add(VertexDist{X: 10, Y: 0, Dist: 0.0})

	initialSize := vs.Size()
	initialLast := vs.Get(vs.Size() - 1)

	// Shorten by 0 units (should not change anything)
	ShortenPath(vs, 0.0, false)

	if vs.Size() != initialSize {
		t.Errorf("Expected size to remain %d, got %d", initialSize, vs.Size())
	}

	last := vs.Get(vs.Size() - 1)
	if last.X != initialLast.X || last.Y != initialLast.Y {
		t.Errorf("Expected last vertex to remain (%f,%f), got (%f,%f)",
			initialLast.X, initialLast.Y, last.X, last.Y)
	}
}

// TestShortenPathClosed tests shortening on closed paths
func TestShortenPathClosed(t *testing.T) {
	vs := NewVertexSequence[VertexDist]()
	vs.Add(VertexDist{X: 0, Y: 0, Dist: 10.0})
	vs.Add(VertexDist{X: 10, Y: 0, Dist: 10.0})
	vs.Add(VertexDist{X: 10, Y: 10, Dist: 10.0})
	vs.Add(VertexDist{X: 0, Y: 10, Dist: 10.0}) // Distance back to start

	// Calculate distances properly
	vs.Close(true)

	initialSize := vs.Size()
	t.Logf("Initial size: %d", initialSize)
	for i := 0; i < vs.Size(); i++ {
		v := vs.Get(i)
		t.Logf("Vertex %d: (%f, %f) dist=%f", i, v.X, v.Y, v.Dist)
	}

	// Shorten closed path by 5 units
	ShortenPath(vs, 5.0, true)

	t.Logf("After shortening size: %d", vs.Size())
	for i := 0; i < vs.Size(); i++ {
		v := vs.Get(i)
		t.Logf("Vertex %d: (%f, %f) dist=%f", i, v.X, v.Y, v.Dist)
	}

	// Should adjust the path but maintain closure semantics
	if vs.Size() == 0 {
		t.Error("Expected vertices to remain after shortening closed path")
	}

	// Last vertex should be adjusted - moved back 5 units from (0,10) toward (10,10)
	last := vs.Get(vs.Size() - 1)
	if last.X != 5.0 { // Should be moved back 5 units along X axis
		t.Errorf("Expected last vertex X to be 5.0, got %f", last.X)
	}
	if last.Y != 10.0 {
		t.Errorf("Expected last vertex Y to be 10.0, got %f", last.Y)
	}
}

// TestShortenPathSingleVertex tests behavior with single vertex
func TestShortenPathSingleVertex(t *testing.T) {
	vs := NewVertexSequence[VertexDist]()
	vs.Add(VertexDist{X: 0, Y: 0, Dist: 0.0})

	// Try to shorten with only one vertex
	ShortenPath(vs, 5.0, false)

	// Should not change anything since we need at least 2 vertices
	if vs.Size() != 1 {
		t.Errorf("Expected size to remain 1, got %d", vs.Size())
	}
}
