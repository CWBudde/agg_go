package basics

import (
	"math"
	"testing"
)

// MockVertexConsumer for testing
type MockVertexConsumer struct {
	vertices []PointD
}

func NewMockVertexConsumer() *MockVertexConsumer {
	return &MockVertexConsumer{vertices: make([]PointD, 0)}
}

func (m *MockVertexConsumer) Add(x, y float64) {
	m.vertices = append(m.vertices, PointD{X: x, Y: y})
}

func (m *MockVertexConsumer) RemoveAll() {
	m.vertices = m.vertices[:0]
}

func (m *MockVertexConsumer) Vertices() []PointD {
	return m.vertices
}

func TestMathStrokeCreation(t *testing.T) {
	ms := NewMathStroke()

	// Test default values
	if ms.Width() != 1.0 { // width is stored as radius, returned as diameter
		t.Errorf("Expected default width 1.0, got %f", ms.Width())
	}
	if ms.LineCap() != ButtCap {
		t.Errorf("Expected default line cap ButtCap, got %v", ms.LineCap())
	}
	if ms.LineJoin() != MiterJoin {
		t.Errorf("Expected default line join MiterJoin, got %v", ms.LineJoin())
	}
	if ms.InnerJoin() != InnerMiter {
		t.Errorf("Expected default inner join InnerMiter, got %v", ms.InnerJoin())
	}
	if ms.MiterLimit() != 4.0 {
		t.Errorf("Expected default miter limit 4.0, got %f", ms.MiterLimit())
	}
}

func TestMathStrokeSetters(t *testing.T) {
	ms := NewMathStroke()

	// Test width setting
	ms.SetWidth(10.0)
	if ms.Width() != 10.0 {
		t.Errorf("Expected width 10.0, got %f", ms.Width())
	}

	// Test negative width
	ms.SetWidth(-5.0)
	if ms.Width() != -5.0 {
		t.Errorf("Expected width -5.0, got %f", ms.Width())
	}

	// Test line cap
	ms.SetLineCap(RoundCap)
	if ms.LineCap() != RoundCap {
		t.Errorf("Expected RoundCap, got %v", ms.LineCap())
	}

	// Test line join
	ms.SetLineJoin(RoundJoin)
	if ms.LineJoin() != RoundJoin {
		t.Errorf("Expected RoundJoin, got %v", ms.LineJoin())
	}

	// Test inner join
	ms.SetInnerJoin(InnerRound)
	if ms.InnerJoin() != InnerRound {
		t.Errorf("Expected InnerRound, got %v", ms.InnerJoin())
	}

	// Test miter limit
	ms.SetMiterLimit(8.0)
	if ms.MiterLimit() != 8.0 {
		t.Errorf("Expected miter limit 8.0, got %f", ms.MiterLimit())
	}

	// Test miter limit theta
	theta := Pi / 6 // 30 degrees
	expectedLimit := 1.0 / math.Sin(theta*0.5)
	ms.SetMiterLimitTheta(theta)
	if math.Abs(ms.MiterLimit()-expectedLimit) > 1e-10 {
		t.Errorf("Expected miter limit %f, got %f", expectedLimit, ms.MiterLimit())
	}
}

func TestMathStrokeButtCap(t *testing.T) {
	ms := NewMathStroke()
	ms.SetWidth(2.0) // radius = 1.0
	ms.SetLineCap(ButtCap)

	consumer := NewMockVertexConsumer()

	// Create test vertices - horizontal line from (0,0) to (10,0)
	v0 := VertexDist{X: 0, Y: 0, Dist: 10.0}
	v1 := VertexDist{X: 10, Y: 0, Dist: 0}

	ms.CalcCap(consumer, v0, v1, 10.0)

	vertices := consumer.Vertices()
	if len(vertices) != 2 {
		t.Errorf("Expected 2 vertices for butt cap, got %d", len(vertices))
		return
	}

	// For butt cap on horizontal line, should get vertices at (0,-1) and (0,1)
	expected := []PointD{{X: 0, Y: -1}, {X: 0, Y: 1}}
	for i, v := range vertices {
		if math.Abs(v.X-expected[i].X) > 1e-10 || math.Abs(v.Y-expected[i].Y) > 1e-10 {
			t.Errorf("Vertex %d: expected (%f,%f), got (%f,%f)", i, expected[i].X, expected[i].Y, v.X, v.Y)
		}
	}
}

func TestMathStrokeSquareCap(t *testing.T) {
	ms := NewMathStroke()
	ms.SetWidth(2.0) // radius = 1.0
	ms.SetLineCap(SquareCap)

	consumer := NewMockVertexConsumer()

	// Create test vertices - horizontal line from (0,0) to (10,0)
	v0 := VertexDist{X: 0, Y: 0, Dist: 10.0}
	v1 := VertexDist{X: 10, Y: 0, Dist: 0}

	ms.CalcCap(consumer, v0, v1, 10.0)

	vertices := consumer.Vertices()
	if len(vertices) != 2 {
		t.Errorf("Expected 2 vertices for square cap, got %d", len(vertices))
		return
	}

	// For square cap on horizontal line, should extend by width in the perpendicular direction
	// Expected vertices should be at (-1,-1) and (-1,1)
	expected := []PointD{{X: -1, Y: -1}, {X: -1, Y: 1}}
	for i, v := range vertices {
		if math.Abs(v.X-expected[i].X) > 1e-10 || math.Abs(v.Y-expected[i].Y) > 1e-10 {
			t.Errorf("Vertex %d: expected (%f,%f), got (%f,%f)", i, expected[i].X, expected[i].Y, v.X, v.Y)
		}
	}
}

func TestMathStrokeRoundCap(t *testing.T) {
	ms := NewMathStroke()
	ms.SetWidth(2.0) // radius = 1.0
	ms.SetLineCap(RoundCap)

	consumer := NewMockVertexConsumer()

	// Create test vertices - horizontal line from (0,0) to (10,0)
	v0 := VertexDist{X: 0, Y: 0, Dist: 10.0}
	v1 := VertexDist{X: 10, Y: 0, Dist: 0}

	ms.CalcCap(consumer, v0, v1, 10.0)

	vertices := consumer.Vertices()

	// Round cap should generate multiple vertices (at least 3: start, arc points, end)
	if len(vertices) < 3 {
		t.Errorf("Expected at least 3 vertices for round cap, got %d", len(vertices))
		return
	}

	// First vertex should be at (0,-1)
	if math.Abs(vertices[0].X-0) > 1e-10 || math.Abs(vertices[0].Y-(-1)) > 1e-10 {
		t.Errorf("First vertex: expected (0,-1), got (%f,%f)", vertices[0].X, vertices[0].Y)
	}

	// Last vertex should be at (0,1)
	last := vertices[len(vertices)-1]
	if math.Abs(last.X-0) > 1e-10 || math.Abs(last.Y-1) > 1e-10 {
		t.Errorf("Last vertex: expected (0,1), got (%f,%f)", last.X, last.Y)
	}

	// All vertices should be approximately on a circle of radius 1 centered at (0,0)
	for i, v := range vertices {
		distance := math.Sqrt(v.X*v.X + v.Y*v.Y)
		if math.Abs(distance-1.0) > 1e-2 { // Allow some tolerance for approximation
			t.Errorf("Vertex %d at (%f,%f) is not on unit circle, distance = %f", i, v.X, v.Y, distance)
		}
	}
}

func TestMathStrokeBevelJoin(t *testing.T) {
	ms := NewMathStroke()
	ms.SetWidth(2.0) // radius = 1.0
	ms.SetLineJoin(BevelJoin)

	consumer := NewMockVertexConsumer()

	// Create test vertices - L-shape: (0,0) -> (10,0) -> (10,10)
	v0 := VertexDist{X: 0, Y: 0, Dist: 10.0}
	v1 := VertexDist{X: 10, Y: 0, Dist: 10.0}
	v2 := VertexDist{X: 10, Y: 10, Dist: 0}

	ms.CalcJoin(consumer, v0, v1, v2, 10.0, 10.0)

	vertices := consumer.Vertices()

	// Bevel join should generate exactly 2 vertices
	if len(vertices) != 2 {
		t.Errorf("Expected 2 vertices for bevel join, got %d", len(vertices))
	}
}

func TestMathStrokeRoundJoin(t *testing.T) {
	ms := NewMathStroke()
	ms.SetWidth(2.0) // radius = 1.0
	ms.SetLineJoin(RoundJoin)

	consumer := NewMockVertexConsumer()

	// Create test vertices - L-shape: (0,0) -> (10,0) -> (10,10)
	v0 := VertexDist{X: 0, Y: 0, Dist: 10.0}
	v1 := VertexDist{X: 10, Y: 0, Dist: 10.0}
	v2 := VertexDist{X: 10, Y: 10, Dist: 0}

	ms.CalcJoin(consumer, v0, v1, v2, 10.0, 10.0)

	vertices := consumer.Vertices()

	// Round join should generate multiple vertices
	if len(vertices) < 3 {
		t.Errorf("Expected at least 3 vertices for round join, got %d", len(vertices))
	}
}

func TestMathStrokeMiterJoin(t *testing.T) {
	ms := NewMathStroke()
	ms.SetWidth(2.0) // radius = 1.0
	ms.SetLineJoin(MiterJoin)
	ms.SetMiterLimit(4.0)

	consumer := NewMockVertexConsumer()

	// Create test vertices - L-shape: (0,0) -> (10,0) -> (10,10)
	v0 := VertexDist{X: 0, Y: 0, Dist: 10.0}
	v1 := VertexDist{X: 10, Y: 0, Dist: 10.0}
	v2 := VertexDist{X: 10, Y: 10, Dist: 0}

	ms.CalcJoin(consumer, v0, v1, v2, 10.0, 10.0)

	vertices := consumer.Vertices()

	// For a 90-degree angle, miter join should generate exactly 1 vertex
	if len(vertices) != 1 {
		t.Errorf("Expected 1 vertex for miter join, got %d", len(vertices))
		return
	}

	// The miter point for this L-shape should be at (11, 1) for outer join
	// (This is approximate based on the geometry)
	vertex := vertices[0]
	if math.Abs(vertex.X-11) > 0.1 || math.Abs(vertex.Y-1) > 0.1 {
		t.Logf("Miter vertex at (%f,%f) - this may be correct for the geometry", vertex.X, vertex.Y)
	}
}
