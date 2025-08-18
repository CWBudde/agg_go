package array

import (
	"testing"
)

func TestNewVertexSequence(t *testing.T) {
	vs := NewVertexSequence[LineAAVertex]()
	if vs == nil {
		t.Fatalf("NewVertexSequence returned nil")
	}
	if vs.Size() != 0 {
		t.Errorf("New vertex sequence size = %d, want 0", vs.Size())
	}
}

func TestNewVertexSequenceWithScale(t *testing.T) {
	scale := NewBlockScale(4)
	vs := NewVertexSequenceWithScale[LineAAVertex](scale)
	if vs == nil {
		t.Fatalf("NewVertexSequenceWithScale returned nil")
	}
	if vs.Size() != 0 {
		t.Errorf("New vertex sequence size = %d, want 0", vs.Size())
	}
}

func TestLineAAVertexValidate(t *testing.T) {
	v1 := NewLineAAVertex(0, 0)
	v2 := NewLineAAVertex(1000, 1000) // Far enough to pass validation
	v3 := NewLineAAVertex(1, 1)       // Too close, should fail validation

	// Test validation - vertices far apart should validate
	if !v1.Validate(v2) {
		t.Errorf("Expected distant vertices to validate")
	}

	// Test validation - vertices close together should not validate
	if v1.Validate(v3) {
		t.Errorf("Expected close vertices to not validate")
	}
}

func TestLineAAVertexCalculateDistance(t *testing.T) {
	v1 := NewLineAAVertex(0, 0)
	v2 := NewLineAAVertex(3, 4) // 3-4-5 triangle, distance = 5

	v1.CalculateDistance(v2)

	// Distance should be approximately 5 (rounded)
	if v1.Len != 5 {
		t.Errorf("CalculateDistance result = %d, want 5", v1.Len)
	}
}

func TestVertexSequenceAdd(t *testing.T) {
	vs := NewVertexSequence[LineAAVertex]()

	// Add first vertex
	v1 := NewLineAAVertex(0, 0)
	vs.Add(v1)
	if vs.Size() != 1 {
		t.Errorf("Size after first add = %d, want 1", vs.Size())
	}

	// Add second vertex
	v2 := NewLineAAVertex(1000, 1000)
	vs.Add(v2)
	if vs.Size() != 2 {
		t.Errorf("Size after second add = %d, want 2", vs.Size())
	}

	// Add third vertex that's far from second
	v3 := NewLineAAVertex(2000, 2000)
	vs.Add(v3)
	if vs.Size() != 3 {
		t.Errorf("Size after third add = %d, want 3", vs.Size())
	}
}

func TestVertexSequenceAddWithFiltering(t *testing.T) {
	vs := NewVertexSequence[LineAAVertex]()

	// Add first vertex
	v1 := NewLineAAVertex(0, 0)
	vs.Add(v1)

	// Add second vertex far away
	v2 := NewLineAAVertex(1000, 1000)
	vs.Add(v2)

	// Add third vertex close to second (should be filtered)
	v3 := NewLineAAVertex(1001, 1001)
	vs.Add(v3)

	// The close vertex might cause filtering of the previous vertex
	// The exact behavior depends on the validation logic
	if vs.Size() < 2 {
		t.Errorf("Expected at least 2 vertices after filtering, got %d", vs.Size())
	}
}

func TestVertexSequenceGet(t *testing.T) {
	vs := NewVertexSequence[LineAAVertex]()

	v1 := NewLineAAVertex(10, 20)
	v2 := NewLineAAVertex(1000, 2000)

	vs.Add(v1)
	vs.Add(v2)

	retrieved1 := vs.Get(0)
	if retrieved1.X != 10 || retrieved1.Y != 20 {
		t.Errorf("Get(0) = (%d, %d), want (10, 20)", retrieved1.X, retrieved1.Y)
	}

	retrieved2 := vs.Get(1)
	if retrieved2.X != 1000 || retrieved2.Y != 2000 {
		t.Errorf("Get(1) = (%d, %d), want (1000, 2000)", retrieved2.X, retrieved2.Y)
	}
}

func TestVertexSequenceModifyLast(t *testing.T) {
	vs := NewVertexSequence[LineAAVertex]()

	v1 := NewLineAAVertex(0, 0)
	v2 := NewLineAAVertex(1000, 1000)

	vs.Add(v1)
	vs.Add(v2)

	if vs.Size() != 2 {
		t.Fatalf("Expected 2 vertices before modify, got %d", vs.Size())
	}

	// Modify the last vertex
	v3 := NewLineAAVertex(2000, 2000)
	vs.ModifyLast(v3)

	// Size should remain the same
	if vs.Size() != 2 {
		t.Errorf("Size after ModifyLast = %d, want 2", vs.Size())
	}

	// Last vertex should be the new one
	last := vs.Get(vs.Size() - 1)
	if last.X != 2000 || last.Y != 2000 {
		t.Errorf("Modified last vertex = (%d, %d), want (2000, 2000)", last.X, last.Y)
	}
}

func TestVertexSequenceClose(t *testing.T) {
	vs := NewVertexSequence[LineAAVertex]()

	// Add several vertices
	vs.Add(NewLineAAVertex(0, 0))
	vs.Add(NewLineAAVertex(1000, 0))
	vs.Add(NewLineAAVertex(1000, 1000))
	vs.Add(NewLineAAVertex(0, 1000))

	originalSize := vs.Size()

	// Close as a polygon
	vs.Close(true)

	// The close operation may filter some vertices
	finalSize := vs.Size()
	if finalSize > originalSize {
		t.Errorf("Close increased size from %d to %d", originalSize, finalSize)
	}
}

func TestVertexSequenceRemoveAll(t *testing.T) {
	vs := NewVertexSequence[LineAAVertex]()

	vs.Add(NewLineAAVertex(0, 0))
	vs.Add(NewLineAAVertex(1000, 1000))

	if vs.Size() == 0 {
		t.Fatalf("Expected vertices before RemoveAll")
	}

	vs.RemoveAll()

	if vs.Size() != 0 {
		t.Errorf("Size after RemoveAll = %d, want 0", vs.Size())
	}
}

func TestVertexDist(t *testing.T) {
	v1 := NewVertexDist(0.0, 0.0)
	v2 := NewVertexDist(3.0, 4.0)     // Distance = 5.0
	v3 := NewVertexDist(1e-15, 1e-15) // Very close, below epsilon (1e-14)

	// Test validation - distant vertices should validate
	if !v1.Validate(v2) {
		t.Errorf("Expected distant VertexDist to validate")
	}

	// Test validation - very close vertices should not validate
	if v1.Validate(v3) {
		t.Errorf("Expected close VertexDist to not validate")
	}
}

func TestVertexDistCalculateDistance(t *testing.T) {
	v1 := NewVertexDist(0.0, 0.0)
	v2 := NewVertexDist(3.0, 4.0) // 3-4-5 triangle

	v1.CalculateDistance(v2)

	// Distance should be 5.0
	if v1.Dist != 5.0 {
		t.Errorf("CalculateDistance result = %f, want 5.0", v1.Dist)
	}
}

func TestVertexDistCalculateDistanceSmall(t *testing.T) {
	v1 := NewVertexDist(0.0, 0.0)
	v2 := NewVertexDist(1e-15, 1e-15) // Very small distance

	v1.CalculateDistance(v2)

	// For very small distances, it should be set to 1/epsilon
	expectedDist := 1.0 / 1e-14 // VertexDistEpsilon = 1e-14
	if v1.Dist != expectedDist {
		t.Errorf("CalculateDistance for small distance = %f, want %f", v1.Dist, expectedDist)
	}
}

func TestVertexSequenceWithVertexDist(t *testing.T) {
	vs := NewVertexSequence[VertexDist]()

	// Add vertices with sufficient distance
	vs.Add(NewVertexDist(0.0, 0.0))
	vs.Add(NewVertexDist(1.0, 1.0))
	vs.Add(NewVertexDist(2.0, 2.0))

	if vs.Size() != 3 {
		t.Errorf("VertexSequence with VertexDist size = %d, want 3", vs.Size())
	}

	// Test getting vertices
	v0 := vs.Get(0)
	if v0.X != 0.0 || v0.Y != 0.0 {
		t.Errorf("Get(0) = (%f, %f), want (0.0, 0.0)", v0.X, v0.Y)
	}
}

// Test type constraint - this should compile if VertexFilter is implemented correctly
func TestVertexFilterConstraint(t *testing.T) {
	// Test that LineAAVertex implements VertexFilter
	var _ VertexFilter = LineAAVertex{}

	// Test that VertexDist implements VertexFilter
	var _ VertexFilter = VertexDist{}
}

func BenchmarkVertexSequenceAdd(b *testing.B) {
	vs := NewVertexSequence[LineAAVertex]()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Add vertices in a way that they won't be filtered
		vs.Add(NewLineAAVertex(i*1000, i*1000))
	}
}

func BenchmarkLineAAVertexValidate(b *testing.B) {
	v1 := NewLineAAVertex(0, 0)
	v2 := NewLineAAVertex(1000, 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v1.Validate(v2)
	}
}

func BenchmarkVertexDistValidate(b *testing.B) {
	v1 := NewVertexDist(0.0, 0.0)
	v2 := NewVertexDist(1.0, 1.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v1.Validate(v2)
	}
}
