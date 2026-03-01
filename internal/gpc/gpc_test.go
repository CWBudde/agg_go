package gpc

import (
	"bytes"
	"math"
	"strings"
	"testing"
	"time"
)

// Test constants (removed unused testEpsilon)

// Helper function to create a simple rectangle polygon
func createRectanglePolygon(x1, y1, x2, y2 float64) *GPCPolygon {
	polygon := NewGPCPolygon()
	contour := NewGPCVertexList(4)
	contour.AddVertex(x1, y1)
	contour.AddVertex(x2, y1)
	contour.AddVertex(x2, y2)
	contour.AddVertex(x1, y2)
	polygon.AddContour(contour, false)
	return polygon
}

// Helper function to create a simple triangle polygon
func createTrianglePolygon(x1, y1, x2, y2, x3, y3 float64) *GPCPolygon {
	polygon := NewGPCPolygon()
	contour := NewGPCVertexList(3)
	contour.AddVertex(x1, y1)
	contour.AddVertex(x2, y2)
	contour.AddVertex(x3, y3)
	polygon.AddContour(contour, false)
	return polygon
}

// Helper function to create a polygon with a hole
func createPolygonWithHole() *GPCPolygon {
	polygon := NewGPCPolygon()

	// Outer contour (counter-clockwise)
	outer := NewGPCVertexList(4)
	outer.AddVertex(0, 0)
	outer.AddVertex(10, 0)
	outer.AddVertex(10, 10)
	outer.AddVertex(0, 10)
	polygon.AddContour(outer, false)

	// Inner hole (clockwise)
	inner := NewGPCVertexList(4)
	inner.AddVertex(3, 3)
	inner.AddVertex(3, 7)
	inner.AddVertex(7, 7)
	inner.AddVertex(7, 3)
	polygon.AddContour(inner, true)

	return polygon
}

func TestGPCOp_String(t *testing.T) {
	tests := []struct {
		op       GPCOp
		expected string
	}{
		{GPCDiff, "Difference"},
		{GPCInt, "Intersection"},
		{GPCXor, "Exclusive-or"},
		{GPCUnion, "Union"},
		{GPCOp(999), "Unknown"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			if got := test.op.String(); got != test.expected {
				t.Errorf("GPCOp.String() = %v, want %v", got, test.expected)
			}
		})
	}
}

func TestGPCVertex_Equal(t *testing.T) {
	tests := []struct {
		name     string
		v1       GPCVertex
		v2       GPCVertex
		expected bool
	}{
		{"Identical vertices", GPCVertex{1.0, 2.0}, GPCVertex{1.0, 2.0}, true},
		{"Within epsilon", GPCVertex{1.0, 2.0}, GPCVertex{1.0 + Epsilon/2, 2.0}, true},
		{"Outside epsilon X", GPCVertex{1.0, 2.0}, GPCVertex{1.0 + Epsilon*2, 2.0}, false},
		{"Outside epsilon Y", GPCVertex{1.0, 2.0}, GPCVertex{1.0, 2.0 + Epsilon*2}, false},
		{"Different vertices", GPCVertex{1.0, 2.0}, GPCVertex{3.0, 4.0}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.v1.Equal(test.v2); got != test.expected {
				t.Errorf("GPCVertex.Equal() = %v, want %v", got, test.expected)
			}
		})
	}
}

func TestGPCVertex_String(t *testing.T) {
	v := GPCVertex{1.234567, 2.345678}
	expected := "(1.234567, 2.345678)"
	if got := v.String(); got != expected {
		t.Errorf("GPCVertex.String() = %v, want %v", got, expected)
	}
}

func TestNewGPCVertexList(t *testing.T) {
	capacity := 10
	vl := NewGPCVertexList(capacity)

	if vl == nil {
		t.Fatal("NewGPCVertexList() returned nil")
	}

	if vl.NumVertices != 0 {
		t.Errorf("NumVertices = %v, want 0", vl.NumVertices)
	}

	if cap(vl.Vertices) != capacity {
		t.Errorf("Vertices capacity = %v, want %v", cap(vl.Vertices), capacity)
	}
}

func TestGPCVertexList_AddVertex(t *testing.T) {
	vl := NewGPCVertexList(5)

	vl.AddVertex(1.0, 2.0)
	if vl.NumVertices != 1 {
		t.Errorf("NumVertices = %v, want 1", vl.NumVertices)
	}

	vertex, err := vl.GetVertex(0)
	if err != nil {
		t.Fatalf("GetVertex(0) error: %v", err)
	}

	expected := GPCVertex{1.0, 2.0}
	if !vertex.Equal(expected) {
		t.Errorf("GetVertex(0) = %v, want %v", vertex, expected)
	}
}

func TestGPCVertexList_GetVertex(t *testing.T) {
	vl := NewGPCVertexList(5)
	vl.AddVertex(1.0, 2.0)
	vl.AddVertex(3.0, 4.0)

	// Valid index
	vertex, err := vl.GetVertex(1)
	if err != nil {
		t.Errorf("GetVertex(1) error: %v", err)
	}
	expected := GPCVertex{3.0, 4.0}
	if !vertex.Equal(expected) {
		t.Errorf("GetVertex(1) = %v, want %v", vertex, expected)
	}

	// Invalid indices
	_, err = vl.GetVertex(-1)
	if err == nil {
		t.Error("GetVertex(-1) should return error")
	}

	_, err = vl.GetVertex(2)
	if err == nil {
		t.Error("GetVertex(2) should return error")
	}
}

func TestGPCVertexList_Clear(t *testing.T) {
	vl := NewGPCVertexList(5)
	vl.AddVertex(1.0, 2.0)
	vl.AddVertex(3.0, 4.0)

	vl.Clear()
	if vl.NumVertices != 0 {
		t.Errorf("NumVertices after Clear() = %v, want 0", vl.NumVertices)
	}

	if len(vl.Vertices) != 0 {
		t.Errorf("len(Vertices) after Clear() = %v, want 0", len(vl.Vertices))
	}
}

func TestNewGPCPolygon(t *testing.T) {
	polygon := NewGPCPolygon()

	if polygon == nil {
		t.Fatal("NewGPCPolygon() returned nil")
	}

	if polygon.NumContours != 0 {
		t.Errorf("NumContours = %v, want 0", polygon.NumContours)
	}

	if !polygon.IsEmpty() {
		t.Error("New polygon should be empty")
	}
}

func TestGPCPolygon_AddContour(t *testing.T) {
	polygon := NewGPCPolygon()
	contour := NewGPCVertexList(3)
	contour.AddVertex(0, 0)
	contour.AddVertex(1, 0)
	contour.AddVertex(0, 1)

	err := polygon.AddContour(contour, false)
	if err != nil {
		t.Errorf("AddContour() error: %v", err)
	}

	if polygon.NumContours != 1 {
		t.Errorf("NumContours = %v, want 1", polygon.NumContours)
	}

	if polygon.IsEmpty() {
		t.Error("Polygon should not be empty after adding contour")
	}

	// Test adding nil contour
	err = polygon.AddContour(nil, false)
	if err == nil {
		t.Error("AddContour(nil) should return error")
	}

	// Test adding contour with insufficient vertices
	badContour := NewGPCVertexList(2)
	badContour.AddVertex(0, 0)
	badContour.AddVertex(1, 0)
	err = polygon.AddContour(badContour, false)
	if err == nil {
		t.Error("AddContour() with < 3 vertices should return error")
	}
}

func TestGPCPolygon_GetContour(t *testing.T) {
	polygon := createTrianglePolygon(0, 0, 1, 0, 0, 1)

	// Valid index
	contour, isHole, err := polygon.GetContour(0)
	if err != nil {
		t.Errorf("GetContour(0) error: %v", err)
	}

	if contour == nil {
		t.Error("GetContour(0) returned nil contour")
		return
	}

	if isHole {
		t.Error("GetContour(0) should not be a hole")
	}

	if contour.NumVertices != 3 {
		t.Errorf("GetContour(0) NumVertices = %v, want 3", contour.NumVertices)
	}

	// Invalid indices
	_, _, err = polygon.GetContour(-1)
	if err == nil {
		t.Error("GetContour(-1) should return error")
	}

	_, _, err = polygon.GetContour(1)
	if err == nil {
		t.Error("GetContour(1) should return error")
	}
}

func TestGPCPolygon_Validate(t *testing.T) {
	// Valid polygon
	validPolygon := createTrianglePolygon(0, 0, 1, 0, 0, 1)
	err := validPolygon.Validate()
	if err != nil {
		t.Errorf("Validate() on valid polygon error: %v", err)
	}

	// Create invalid polygon with mismatched counts
	invalidPolygon := &GPCPolygon{
		NumContours: 2,
		Contours:    []*GPCVertexList{NewGPCVertexList(3)}, // Only 1 contour
		Hole:        []bool{false, false},                  // But 2 hole flags
	}

	err = invalidPolygon.Validate()
	if err == nil {
		t.Error("Validate() on invalid polygon should return error")
	}
}

func TestNewGPCTristrip(t *testing.T) {
	tristrip := NewGPCTristrip()

	if tristrip == nil {
		t.Fatal("NewGPCTristrip() returned nil")
	}

	if tristrip.NumStrips != 0 {
		t.Errorf("NumStrips = %v, want 0", tristrip.NumStrips)
	}

	if !tristrip.IsEmpty() {
		t.Error("New tristrip should be empty")
	}
}

func TestGPCTristrip_AddStrip(t *testing.T) {
	tristrip := NewGPCTristrip()
	strip := NewGPCVertexList(3)
	strip.AddVertex(0, 0)
	strip.AddVertex(1, 0)
	strip.AddVertex(0, 1)

	err := tristrip.AddStrip(strip)
	if err != nil {
		t.Errorf("AddStrip() error: %v", err)
	}

	if tristrip.NumStrips != 1 {
		t.Errorf("NumStrips = %v, want 1", tristrip.NumStrips)
	}

	if tristrip.IsEmpty() {
		t.Error("Tristrip should not be empty after adding strip")
	}

	// Test adding nil strip
	err = tristrip.AddStrip(nil)
	if err == nil {
		t.Error("AddStrip(nil) should return error")
	}
}

func TestGPCTristrip_GetStrip(t *testing.T) {
	tristrip := NewGPCTristrip()
	strip := NewGPCVertexList(3)
	strip.AddVertex(0, 0)
	strip.AddVertex(1, 0)
	strip.AddVertex(0, 1)
	tristrip.AddStrip(strip)

	// Valid index
	retrievedStrip, err := tristrip.GetStrip(0)
	if err != nil {
		t.Errorf("GetStrip(0) error: %v", err)
	}

	if retrievedStrip != strip {
		t.Error("GetStrip(0) returned wrong strip")
	}

	// Invalid indices
	_, err = tristrip.GetStrip(-1)
	if err == nil {
		t.Error("GetStrip(-1) should return error")
	}

	_, err = tristrip.GetStrip(1)
	if err == nil {
		t.Error("GetStrip(1) should return error")
	}
}

func TestWritePolygon(t *testing.T) {
	polygon := createTrianglePolygon(0, 0, 1, 0, 0.5, 1)

	tests := []struct {
		name           string
		writeHoleFlags bool
		expected       string
	}{
		{
			name:           "Without hole flags",
			writeHoleFlags: false,
			expected:       "1\n3\n0.000000 0.000000\n1.000000 0.000000\n0.500000 1.000000\n",
		},
		{
			name:           "With hole flags",
			writeHoleFlags: true,
			expected:       "1\n0 3\n0.000000 0.000000\n1.000000 0.000000\n0.500000 1.000000\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WritePolygon(&buf, polygon, test.writeHoleFlags)
			if err != nil {
				t.Errorf("WritePolygon() error: %v", err)
			}

			if got := buf.String(); got != test.expected {
				t.Errorf("WritePolygon() = %q, want %q", got, test.expected)
			}
		})
	}

	// Test with nil polygon
	var buf bytes.Buffer
	err := WritePolygon(&buf, nil, false)
	if err == nil {
		t.Error("WritePolygon(nil) should return error")
	}
}

func TestAddContour(t *testing.T) {
	polygon := NewGPCPolygon()
	contour := NewGPCVertexList(3)
	contour.AddVertex(0, 0)
	contour.AddVertex(1, 0)
	contour.AddVertex(0, 1)

	err := AddContour(polygon, contour, false)
	if err != nil {
		t.Errorf("AddContour() error: %v", err)
	}

	if polygon.NumContours != 1 {
		t.Errorf("NumContours = %v, want 1", polygon.NumContours)
	}

	// Test with nil polygon
	err = AddContour(nil, contour, false)
	if err == nil {
		t.Error("AddContour(nil, ...) should return error")
	}
}

func TestPolygonClip(t *testing.T) {
	subject := createRectanglePolygon(0, 0, 4, 4)
	clip := createRectanglePolygon(2, 2, 6, 6)

	tests := []struct {
		name      string
		operation GPCOp
	}{
		{"Union", GPCUnion},
		{"Intersection", GPCInt},
		{"Difference", GPCDiff},
		{"Exclusive-or", GPCXor},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := PolygonClip(test.operation, subject, clip)
			if err != nil {
				t.Errorf("PolygonClip() error: %v", err)
			}

			if result == nil {
				t.Error("PolygonClip() returned nil result")
			}

			// Placeholder tests since the algorithm is not fully implemented
			// The actual tests would verify the geometric correctness of the results
		})
	}

	// Test with nil polygons
	_, err := PolygonClip(GPCUnion, nil, clip)
	if err == nil {
		t.Error("PolygonClip(nil, ...) should return error")
	}

	_, err = PolygonClip(GPCUnion, subject, nil)
	if err == nil {
		t.Error("PolygonClip(..., nil) should return error")
	}
}

func TestTristripClip(t *testing.T) {
	subject := createTrianglePolygon(0, 0, 2, 0, 1, 2)
	clip := createTrianglePolygon(1, 0, 3, 0, 2, 2)

	result, err := TristripClip(GPCInt, subject, clip)
	if err != nil {
		t.Errorf("TristripClip() error: %v", err)
	}

	if result == nil {
		t.Error("TristripClip() returned nil result")
	}

	// Test with nil polygons
	_, err = TristripClip(GPCUnion, nil, clip)
	if err == nil {
		t.Error("TristripClip(nil, ...) should return error")
	}
}

func TestPolygonToTristrip(t *testing.T) {
	polygon := createTrianglePolygon(0, 0, 2, 0, 1, 2)

	result, err := PolygonToTristrip(polygon)
	if err != nil {
		t.Errorf("PolygonToTristrip() error: %v", err)
	}

	if result == nil {
		t.Error("PolygonToTristrip() returned nil result")
	}

	// Test with nil polygon
	_, err = PolygonToTristrip(nil)
	if err == nil {
		t.Error("PolygonToTristrip(nil) should return error")
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test eq function
	if !eq(1.0, 1.0+Epsilon/2) {
		t.Error("eq() should return true for values within epsilon")
	}

	if eq(1.0, 1.0+Epsilon*2) {
		t.Error("eq() should return false for values outside epsilon")
	}

	// Test isClockwise function
	clockwiseVertices := []GPCVertex{
		{0, 0}, {0, 1}, {1, 1}, {1, 0}, // Square vertices in clockwise order
	}

	counterClockwiseVertices := []GPCVertex{
		{0, 0}, {1, 0}, {1, 1}, {0, 1}, // Square vertices in counter-clockwise order
	}

	if !isClockwise(clockwiseVertices) {
		t.Error("isClockwise() should return true for clockwise vertices")
	}

	if isClockwise(counterClockwiseVertices) {
		t.Error("isClockwise() should return false for counter-clockwise vertices")
	}

	// Test validateContourWinding function
	contour := NewGPCVertexList(4)
	for _, v := range clockwiseVertices {
		contour.AddVertex(v.X, v.Y)
	}

	if !validateContourWinding(contour, true) {
		t.Error("validateContourWinding() should return true for correct clockwise winding")
	}

	if validateContourWinding(contour, false) {
		t.Error("validateContourWinding() should return false for incorrect winding")
	}
}

func TestComplexPolygonOperations(t *testing.T) {
	// Test with polygon containing holes
	polygonWithHole := createPolygonWithHole()

	err := polygonWithHole.Validate()
	if err != nil {
		t.Errorf("Polygon with hole validation error: %v", err)
	}

	if polygonWithHole.NumContours != 2 {
		t.Errorf("Polygon with hole should have 2 contours, got %d", polygonWithHole.NumContours)
	}

	// Verify hole flags
	_, isHole0, _ := polygonWithHole.GetContour(0)
	_, isHole1, _ := polygonWithHole.GetContour(1)

	if isHole0 {
		t.Error("First contour should not be a hole")
	}

	if !isHole1 {
		t.Error("Second contour should be a hole")
	}
}

func TestEdgeCases(t *testing.T) {
	// Test with very small polygons
	tinyTriangle := createTrianglePolygon(0, 0, Epsilon/2, 0, 0, Epsilon/2)
	err := tinyTriangle.Validate()
	if err != nil {
		t.Errorf("Tiny triangle validation error: %v", err)
	}

	// Test with degenerate cases
	degenerateContour := NewGPCVertexList(3)
	degenerateContour.AddVertex(0, 0)
	degenerateContour.AddVertex(0, 0) // Duplicate vertex
	degenerateContour.AddVertex(1, 1)

	degeneratePolygon := NewGPCPolygon()
	err = degeneratePolygon.AddContour(degenerateContour, false)
	if err != nil {
		t.Errorf("Adding degenerate contour error: %v", err)
	}
}

func TestNumericalPrecision(t *testing.T) {
	// Test with high precision coordinates
	highPrecision := createTrianglePolygon(
		math.Pi, math.E,
		math.Sqrt(2), math.Sqrt(3),
		math.Log(2), math.Log(10),
	)

	err := highPrecision.Validate()
	if err != nil {
		t.Errorf("High precision polygon validation error: %v", err)
	}

	// Test floating point edge cases
	largeCoords := createTrianglePolygon(1e10, 1e10, 1e10+1, 1e10, 1e10, 1e10+1)
	err = largeCoords.Validate()
	if err != nil {
		t.Errorf("Large coordinates polygon validation error: %v", err)
	}
}

func TestReadPolygon(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		readHoleFlags bool
		wantContours  int
		wantHoles     []bool
		wantVertices  []int
	}{
		{
			name:          "Without hole flags",
			input:         "1\n3\n0.000000 0.000000\n1.000000 0.000000\n0.500000 1.000000\n",
			readHoleFlags: false,
			wantContours:  1,
			wantHoles:     []bool{false},
			wantVertices:  []int{3},
		},
		{
			name:          "With hole flags",
			input:         "2\n0 4\n0 0\n10 0\n10 10\n0 10\n1 4\n3 3\n3 7\n7 7\n7 3\n",
			readHoleFlags: true,
			wantContours:  2,
			wantHoles:     []bool{false, true},
			wantVertices:  []int{4, 4},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			polygon, err := ReadPolygon(strings.NewReader(test.input), test.readHoleFlags)
			if err != nil {
				t.Fatalf("ReadPolygon() error: %v", err)
			}

			if polygon.NumContours != test.wantContours {
				t.Fatalf("NumContours = %d, want %d", polygon.NumContours, test.wantContours)
			}

			for i := 0; i < test.wantContours; i++ {
				contour, isHole, err := polygon.GetContour(i)
				if err != nil {
					t.Fatalf("GetContour(%d) error: %v", i, err)
				}
				if isHole != test.wantHoles[i] {
					t.Fatalf("contour %d hole = %v, want %v", i, isHole, test.wantHoles[i])
				}
				if contour.NumVertices != test.wantVertices[i] {
					t.Fatalf("contour %d vertex count = %d, want %d", i, contour.NumVertices, test.wantVertices[i])
				}
			}
		})
	}
}

func TestReadPolygonErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		readHoleFlags bool
	}{
		{
			name:          "Bad contour count",
			input:         "not-a-number\n",
			readHoleFlags: false,
		},
		{
			name:          "Negative contour count",
			input:         "-1\n",
			readHoleFlags: false,
		},
		{
			name:          "Bad hole flag",
			input:         "1\n2 3\n0 0\n1 0\n0 1\n",
			readHoleFlags: true,
		},
		{
			name:          "Too few vertices",
			input:         "1\n2\n0 0\n1 0\n",
			readHoleFlags: false,
		},
		{
			name:          "Truncated vertices",
			input:         "1\n3\n0 0\n1 0\n",
			readHoleFlags: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := ReadPolygon(strings.NewReader(test.input), test.readHoleFlags); err == nil {
				t.Fatal("ReadPolygon() error = nil, want non-nil")
			}
		})
	}

	if _, err := ReadPolygon(nil, false); err == nil {
		t.Fatal("ReadPolygon(nil) error = nil, want non-nil")
	}
}

// TestComplexGeometricOperations tests the full GPC implementation with real geometric operations
func TestComplexGeometricOperations(t *testing.T) {
	// Create two overlapping rectangles for testing
	createRect := func(x1, y1, x2, y2 float64) *GPCPolygon {
		poly := NewGPCPolygon()
		contour := NewGPCVertexList(4)
		contour.AddVertex(x1, y1) // Bottom-left
		contour.AddVertex(x2, y1) // Bottom-right
		contour.AddVertex(x2, y2) // Top-right
		contour.AddVertex(x1, y2) // Top-left
		poly.AddContour(contour, false)
		return poly
	}

	rect1 := createRect(0, 0, 10, 10)
	rect2 := createRect(5, 5, 15, 15)

	tests := []struct {
		name        string
		operation   GPCOp
		minContours int
	}{
		{
			name:        "Union",
			operation:   GPCUnion,
			minContours: 1,
		},
		{
			name:        "Intersection",
			operation:   GPCInt,
			minContours: 1,
		},
		{
			name:        "Difference",
			operation:   GPCDiff,
			minContours: 1,
		},
		{
			name:        "XOR",
			operation:   GPCXor,
			minContours: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := PolygonClip(tt.operation, rect1, rect2)
			if err != nil {
				t.Errorf("PolygonClip() error = %v", err)
				return
			}

			if result == nil {
				t.Error("PolygonClip() returned nil result")
				return
			}

			if result.NumContours < tt.minContours {
				t.Fatalf("PolygonClip(%s) contours = %d, want at least %d", tt.name, result.NumContours, tt.minContours)
			}

			t.Logf("Operation %s completed successfully with %d contours", tt.name, result.NumContours)
		})
	}
}

func TestPolygonClipDebugOverlappingRectangles(t *testing.T) {
	rect1 := createRectanglePolygon(0, 0, 10, 10)
	rect2 := createRectanglePolygon(5, 5, 15, 15)

	result, err := PolygonClip(GPCUnion, rect1, rect2)
	if err != nil {
		t.Fatalf("PolygonClip() error = %v", err)
	}

	t.Logf(
		"debug: local_min_adds=%d out_poly_nil=%v max_raw_vertices=%d counted_contours=%d result_contours=%d boundary_adds=(L:%d R:%d) intersect_adds=(L:%d R:%d) boundary_vclass=%v intersect_vclass=%v",
		lastPolygonClipDebugInfo.LocalMinAdds,
		lastPolygonClipDebugInfo.OutPolyWasNil,
		lastPolygonClipDebugInfo.MaxRawVertices,
		lastPolygonClipDebugInfo.CountedContours,
		result.NumContours,
		lastPolygonClipDebugInfo.BoundaryAddsLeft,
		lastPolygonClipDebugInfo.BoundaryAddsRight,
		lastPolygonClipDebugInfo.IntersectAddsLeft,
		lastPolygonClipDebugInfo.IntersectAddsRight,
		lastPolygonClipDebugInfo.BoundaryVClass,
		lastPolygonClipDebugInfo.IntersectVClass,
	)
}

// TestGPCPerformanceBenchmark provides basic performance testing
func TestGPCPerformanceBenchmark(t *testing.T) {
	// Create simple polygons for performance testing
	createCircleApprox := func(cx, cy, radius float64, segments int) *GPCPolygon {
		poly := NewGPCPolygon()
		contour := NewGPCVertexList(segments)

		for i := 0; i < segments; i++ {
			angle := 2 * math.Pi * float64(i) / float64(segments)
			x := cx + radius*math.Cos(angle)
			y := cy + radius*math.Sin(angle)
			contour.AddVertex(x, y)
		}

		poly.AddContour(contour, false)
		return poly
	}

	circle1 := createCircleApprox(0, 0, 10, 16)
	circle2 := createCircleApprox(5, 5, 10, 16)

	start := time.Now()
	result, err := PolygonClip(GPCUnion, circle1, circle2)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Performance test failed with error: %v", err)
		return
	}

	if result == nil {
		t.Error("Performance test returned nil result")
		return
	}

	t.Logf("Performance test: Union of two 16-sided polygons completed in %v", elapsed)
	t.Logf("Result has %d contours", result.NumContours)

	// Performance should be reasonable (under 1 second for simple polygons)
	if elapsed > time.Second {
		t.Logf("Warning: Performance may be suboptimal (%v > 1s)", elapsed)
	}
}

// TestPolygonToTristripWithHoles tests the improved PolygonToTristrip implementation with holes
func TestPolygonToTristripWithHoles(t *testing.T) {
	// Create a polygon with a hole
	polygon := createPolygonWithHole()

	// Convert to tristrip using the new implementation
	result, err := PolygonToTristrip(polygon)
	if err != nil {
		t.Errorf("PolygonToTristrip() error: %v", err)
		return
	}

	if result == nil {
		t.Error("PolygonToTristrip() returned nil result")
		return
	}

	t.Logf("PolygonToTristrip with holes produced %d triangle strips", result.NumStrips)
}

// TestCompleteScanlineAlgorithm tests the complete GPC scanline algorithm
func TestCompleteScanlineAlgorithm(t *testing.T) {
	// Create two overlapping rectangles
	rect1 := createRectanglePolygon(0, 0, 10, 10)
	rect2 := createRectanglePolygon(5, 5, 15, 15)

	tests := []struct {
		name           string
		operation      GPCOp
		expectContours bool
	}{
		{"Union", GPCUnion, true},
		{"Intersection", GPCInt, true},
		{"Difference", GPCDiff, true},
		{"XOR", GPCXor, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := PolygonClip(tt.operation, rect1, rect2)
			if err != nil {
				t.Errorf("PolygonClip(%s) error: %v", tt.name, err)
				return
			}

			if result == nil {
				t.Error("PolygonClip() returned nil result")
				return
			}

			// For most operations, we should get some result
			if tt.expectContours {
				t.Logf("Operation %s produced %d contours", tt.name, result.NumContours)
			}

			// Validate the result structure
			if err := result.Validate(); err != nil {
				t.Errorf("Result validation failed for %s: %v", tt.name, err)
			}
		})
	}
}

// TestTristripClipWithComplexPolygons tests tristrip clipping with more complex scenarios
func TestTristripClipWithComplexPolygons(t *testing.T) {
	// Create a complex polygon (hexagon)
	hexagon := NewGPCPolygon()
	vertices := NewGPCVertexList(6)
	for i := 0; i < 6; i++ {
		angle := 2 * math.Pi * float64(i) / 6
		x := 5 + 3*math.Cos(angle)
		y := 5 + 3*math.Sin(angle)
		vertices.AddVertex(x, y)
	}
	hexagon.AddContour(vertices, false)

	// Create a square for clipping
	square := createRectanglePolygon(3, 3, 7, 7)

	// Test tristrip clipping
	result, err := TristripClip(GPCInt, hexagon, square)
	if err != nil {
		t.Errorf("TristripClip() error: %v", err)
		return
	}

	if result == nil {
		t.Error("TristripClip() returned nil result")
		return
	}

	t.Logf("TristripClip produced %d triangle strips", result.NumStrips)
}

// TestEdgeCasesForGPCOperations tests edge cases in GPC operations
func TestEdgeCasesForGPCOperations(t *testing.T) {
	// Test with empty polygons
	empty := NewGPCPolygon()
	rect := createRectanglePolygon(0, 0, 5, 5)

	// Union with empty polygon
	result, err := PolygonClip(GPCUnion, empty, rect)
	if err != nil {
		t.Errorf("Union with empty polygon error: %v", err)
	} else if result == nil || result.NumContours == 0 {
		t.Error("Union with empty polygon should return the non-empty polygon")
	}

	// Intersection with empty polygon
	result, err = PolygonClip(GPCInt, rect, empty)
	if err != nil {
		t.Errorf("Intersection with empty polygon error: %v", err)
	} else if result == nil || result.NumContours != 0 {
		t.Error("Intersection with empty polygon should return empty")
	}

	// Difference with empty polygon
	result, err = PolygonClip(GPCDiff, rect, empty)
	if err != nil {
		t.Errorf("Difference with empty polygon error: %v", err)
	} else if result == nil || result.NumContours == 0 {
		t.Error("Difference with empty polygon should return the subject polygon")
	}
}

// TestPolygonWithMultipleHoles tests polygons with multiple holes
func TestPolygonWithMultipleHoles(t *testing.T) {
	polygon := NewGPCPolygon()

	// Large outer contour
	outer := NewGPCVertexList(4)
	outer.AddVertex(0, 0)
	outer.AddVertex(20, 0)
	outer.AddVertex(20, 20)
	outer.AddVertex(0, 20)
	polygon.AddContour(outer, false)

	// First hole
	hole1 := NewGPCVertexList(4)
	hole1.AddVertex(2, 2)
	hole1.AddVertex(2, 8)
	hole1.AddVertex(8, 8)
	hole1.AddVertex(8, 2)
	polygon.AddContour(hole1, true)

	// Second hole
	hole2 := NewGPCVertexList(4)
	hole2.AddVertex(12, 12)
	hole2.AddVertex(12, 18)
	hole2.AddVertex(18, 18)
	hole2.AddVertex(18, 12)
	polygon.AddContour(hole2, true)

	// Test validation
	err := polygon.Validate()
	if err != nil {
		t.Errorf("Polygon with multiple holes validation failed: %v", err)
	}

	if polygon.NumContours != 3 {
		t.Errorf("Expected 3 contours, got %d", polygon.NumContours)
	}

	// Test conversion to tristrip
	result, err := PolygonToTristrip(polygon)
	if err != nil {
		t.Errorf("PolygonToTristrip() with multiple holes error: %v", err)
	} else {
		t.Logf("Polygon with multiple holes converted to %d triangle strips", result.NumStrips)
	}
}

// Benchmark tests will be in a separate file
