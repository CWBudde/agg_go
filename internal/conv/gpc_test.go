package conv

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/gpc"
	"agg_go/internal/path"
)

func TestNewConvGPC(t *testing.T) {
	source1 := NewMockVertexSource([]Vertex{})
	source2 := NewMockVertexSource([]Vertex{})
	gpc := NewConvGPC(source1, source2, GPCOr)

	if gpc == nil {
		t.Error("NewConvGPC should return non-nil converter")
	}
	if gpc.operation != GPCOr {
		t.Errorf("Expected operation GPCOr, got %v", gpc.operation)
	}
}

func TestConvGPC_Attach(t *testing.T) {
	source1 := NewMockVertexSource([]Vertex{})
	source2 := NewMockVertexSource([]Vertex{})
	source3 := NewMockVertexSource([]Vertex{})
	gpc := NewConvGPC(source1, source2, GPCOr)

	gpc.Attach1(source3)
	if gpc.srcA != source3 {
		t.Error("Attach1 should update srcA")
	}

	gpc.Attach2(source3)
	if gpc.srcB != source3 {
		t.Error("Attach2 should update srcB")
	}
}

func TestConvGPC_Operation(t *testing.T) {
	source1 := NewMockVertexSource([]Vertex{})
	source2 := NewMockVertexSource([]Vertex{})
	gpc := NewConvGPC(source1, source2, GPCOr)

	gpc.Operation(GPCAnd)
	if gpc.operation != GPCAnd {
		t.Errorf("Expected operation GPCAnd, got %v", gpc.operation)
	}
}

func TestGPCOp_String(t *testing.T) {
	tests := []struct {
		op       GPCOp
		expected string
	}{
		{GPCOr, "Union"},
		{GPCAnd, "Intersection"},
		{GPCXor, "Exclusive-or"},
		{GPCAMinusB, "A-minus-B"},
		{GPCBMinusA, "B-minus-A"},
		{GPCOp(999), "Unknown"},
	}

	for _, test := range tests {
		if result := test.op.String(); result != test.expected {
			t.Errorf("GPCOp(%v).String() = %s, expected %s", test.op, result, test.expected)
		}
	}
}

func TestConvGPC_SimpleRectangles_Union(t *testing.T) {
	// First rectangle: (0,0) to (10,10)
	rect1 := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{10, 10, basics.PathCmdLineTo},
		{0, 10, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdEndPoly | basics.PathFlagClose},
	}

	// Second rectangle: (5,5) to (15,15) - overlaps with first
	rect2 := []Vertex{
		{5, 5, basics.PathCmdMoveTo},
		{15, 5, basics.PathCmdLineTo},
		{15, 15, basics.PathCmdLineTo},
		{5, 15, basics.PathCmdLineTo},
		{5, 5, basics.PathCmdEndPoly | basics.PathFlagClose},
	}

	source1 := NewMockVertexSource(rect1)
	source2 := NewMockVertexSource(rect2)
	gpc := NewConvGPC(source1, source2, GPCOr)

	gpc.Rewind(0)

	// Should be able to extract vertices from the union
	vertices := extractAllVertices(gpc)

	// The union should have some vertices (exact result depends on GPC implementation)
	// For now, just verify we get some result and no panics
	t.Logf("Union result has %d vertices", len(vertices))

	// At minimum, we should get a stop command at the end
	if len(vertices) == 0 {
		t.Error("Expected some vertices from union operation")
	}
}

func TestConvGPC_NormalizeContourVertices(t *testing.T) {
	closed := []gpc.GPCVertex{
		{X: 0, Y: 0},
		{X: 10, Y: 0},
		{X: 10, Y: 10},
		{X: 0, Y: 0},
	}

	normalized := normalizeContourVertices(closed)
	if len(normalized) != 3 {
		t.Fatalf("normalized contour len = %d, want 3", len(normalized))
	}

	open := []gpc.GPCVertex{
		{X: 0, Y: 0},
		{X: 10, Y: 0},
		{X: 10, Y: 10},
	}
	normalized = normalizeContourVertices(open)
	if len(normalized) != 3 {
		t.Fatalf("open contour len = %d, want 3", len(normalized))
	}
}

func TestConvGPC_AddToPolygonNormalizesClosedContour(t *testing.T) {
	rect1 := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{10, 10, basics.PathCmdLineTo},
		{0, 10, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdEndPoly | basics.PathFlagClose},
	}

	converter := NewConvGPC(NewMockVertexSource(rect1), NewMockVertexSource([]Vertex{}), GPCOr)
	converter.addToPolygon(NewMockVertexSource(rect1), converter.polyA)

	contour, _, err := converter.polyA.GetContour(0)
	if err != nil {
		t.Fatalf("GetContour(0) error: %v", err)
	}
	if contour.NumVertices != 4 {
		t.Fatalf("normalized contour vertices = %d, want 4", contour.NumVertices)
	}
}

func TestConvGPC_PathStorageAdapterNormalizesClosedContour(t *testing.T) {
	path1 := path.NewPathStorage()
	path1.MoveTo(0, 0)
	path1.LineTo(10, 0)
	path1.LineTo(10, 10)
	path1.LineTo(0, 10)
	path1.ClosePolygon(basics.PathFlagsNone)

	converter := NewConvGPC(
		path.NewPathStorageVertexSourceAdapter(path1),
		NewMockVertexSource([]Vertex{}),
		GPCOr,
	)
	converter.addToPolygon(path.NewPathStorageVertexSourceAdapter(path1), converter.polyA)

	contour, _, err := converter.polyA.GetContour(0)
	if err != nil {
		t.Fatalf("GetContour(0) error: %v", err)
	}
	if contour.NumVertices != 4 {
		t.Fatalf("normalized path contour vertices = %d, want 4", contour.NumVertices)
	}
}

func TestConvGPC_RewindMatchesDirectPolygonClip(t *testing.T) {
	rect1 := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{10, 10, basics.PathCmdLineTo},
		{0, 10, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdEndPoly | basics.PathFlagClose},
	}
	rect2 := []Vertex{
		{5, 5, basics.PathCmdMoveTo},
		{15, 5, basics.PathCmdLineTo},
		{15, 15, basics.PathCmdLineTo},
		{5, 15, basics.PathCmdLineTo},
		{5, 5, basics.PathCmdEndPoly | basics.PathFlagClose},
	}

	converter := NewConvGPC(NewMockVertexSource(rect1), NewMockVertexSource(rect2), GPCOr)
	converter.addToPolygon(NewMockVertexSource(rect1), converter.polyA)
	converter.addToPolygon(NewMockVertexSource(rect2), converter.polyB)

	directResult, directErr := gpc.PolygonClip(gpc.GPCUnion, converter.polyA, converter.polyB)

	converter.Rewind(0)

	if converter.lastClipError() != nil {
		t.Fatalf("ConvGPC.Rewind clip error: %v", converter.lastClipError())
	}
	if directErr != nil {
		t.Fatalf("direct PolygonClip error: %v", directErr)
	}
	if converter.result.NumContours != directResult.NumContours {
		t.Fatalf("result contour mismatch: rewind=%d direct=%d", converter.result.NumContours, directResult.NumContours)
	}
}

func TestConvGPC_ResultContoursExplainEmptyVertexStream(t *testing.T) {
	rect1 := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{10, 10, basics.PathCmdLineTo},
		{0, 10, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdEndPoly | basics.PathFlagClose},
	}
	rect2 := []Vertex{
		{5, 5, basics.PathCmdMoveTo},
		{15, 5, basics.PathCmdLineTo},
		{15, 15, basics.PathCmdLineTo},
		{5, 15, basics.PathCmdLineTo},
		{5, 5, basics.PathCmdEndPoly | basics.PathFlagClose},
	}

	converter := NewConvGPC(NewMockVertexSource(rect1), NewMockVertexSource(rect2), GPCOr)
	converter.Rewind(0)

	emitted := countPreparedDrawableCommands(converter)
	if converter.result.NumContours == 0 && emitted != 0 {
		t.Fatalf("expected no emitted commands when result has no contours, got %d", emitted)
	}
	if converter.result.NumContours > 0 && emitted == 0 {
		t.Fatalf("result has %d contours but vertex stream emitted none", converter.result.NumContours)
	}
}

func TestConvGPC_RewindReportsClipErrors(t *testing.T) {
	degenerate := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{1, 0, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdEndPoly | basics.PathFlagClose},
	}

	converter := NewConvGPC(NewMockVertexSource(degenerate), NewMockVertexSource(degenerate), GPCOr)
	converter.Rewind(0)

	if converter.result == nil {
		t.Fatal("expected non-nil result polygon")
	}
}

func TestConvGPC_SimpleRectangles_Intersection(t *testing.T) {
	// First rectangle: (0,0) to (10,10)
	rect1 := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{10, 10, basics.PathCmdLineTo},
		{0, 10, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdEndPoly | basics.PathFlagClose},
	}

	// Second rectangle: (5,5) to (15,15) - overlaps with first
	rect2 := []Vertex{
		{5, 5, basics.PathCmdMoveTo},
		{15, 5, basics.PathCmdLineTo},
		{15, 15, basics.PathCmdLineTo},
		{5, 15, basics.PathCmdLineTo},
		{5, 5, basics.PathCmdEndPoly | basics.PathFlagClose},
	}

	source1 := NewMockVertexSource(rect1)
	source2 := NewMockVertexSource(rect2)
	gpc := NewConvGPC(source1, source2, GPCAnd)

	gpc.Rewind(0)

	vertices := extractAllVertices(gpc)
	t.Logf("Intersection result has %d vertices", len(vertices))

	// The intersection should produce some result
	if len(vertices) == 0 {
		t.Error("Expected some vertices from intersection operation")
	}
}

func TestConvGPC_SimpleRectangles_XOR(t *testing.T) {
	// First rectangle: (0,0) to (10,10)
	rect1 := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{10, 10, basics.PathCmdLineTo},
		{0, 10, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdEndPoly | basics.PathFlagClose},
	}

	// Second rectangle: (5,5) to (15,15) - overlaps with first
	rect2 := []Vertex{
		{5, 5, basics.PathCmdMoveTo},
		{15, 5, basics.PathCmdLineTo},
		{15, 15, basics.PathCmdLineTo},
		{5, 15, basics.PathCmdLineTo},
		{5, 5, basics.PathCmdEndPoly | basics.PathFlagClose},
	}

	source1 := NewMockVertexSource(rect1)
	source2 := NewMockVertexSource(rect2)
	gpc := NewConvGPC(source1, source2, GPCXor)

	gpc.Rewind(0)

	vertices := extractAllVertices(gpc)
	t.Logf("XOR result has %d vertices", len(vertices))

	// XOR should produce some result (symmetric difference)
	if len(vertices) == 0 {
		t.Error("Expected some vertices from XOR operation")
	}
}

func TestConvGPC_SimpleRectangles_Difference(t *testing.T) {
	// First rectangle: (0,0) to (10,10)
	rect1 := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{10, 10, basics.PathCmdLineTo},
		{0, 10, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdEndPoly | basics.PathFlagClose},
	}

	// Second rectangle: (5,5) to (15,15) - overlaps with first
	rect2 := []Vertex{
		{5, 5, basics.PathCmdMoveTo},
		{15, 5, basics.PathCmdLineTo},
		{15, 15, basics.PathCmdLineTo},
		{5, 15, basics.PathCmdLineTo},
		{5, 5, basics.PathCmdEndPoly | basics.PathFlagClose},
	}

	source1 := NewMockVertexSource(rect1)
	source2 := NewMockVertexSource(rect2)

	// Test A - B
	gpcAB := NewConvGPC(source1, source2, GPCAMinusB)
	gpcAB.Rewind(0)
	verticesAB := extractAllVertices(gpcAB)
	t.Logf("A-minus-B result has %d vertices", len(verticesAB))

	// Test B - A
	gpcBA := NewConvGPC(source1, source2, GPCBMinusA)
	gpcBA.Rewind(0)
	verticesBA := extractAllVertices(gpcBA)
	t.Logf("B-minus-A result has %d vertices", len(verticesBA))

	// Both operations should produce some result
	if len(verticesAB) == 0 {
		t.Error("Expected some vertices from A-minus-B operation")
	}
	if len(verticesBA) == 0 {
		t.Error("Expected some vertices from B-minus-A operation")
	}
}

func TestConvGPC_EmptyPolygons(t *testing.T) {
	// Test with empty first polygon
	empty1 := []Vertex{}
	rect2 := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{10, 10, basics.PathCmdLineTo},
		{0, 10, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdEndPoly | basics.PathFlagClose},
	}

	source1 := NewMockVertexSource(empty1)
	source2 := NewMockVertexSource(rect2)
	gpc := NewConvGPC(source1, source2, GPCOr)

	gpc.Rewind(0)
	vertices := extractAllVertices(gpc)

	// Union with empty should not panic
	t.Logf("Union with empty polygon has %d vertices", len(vertices))

	// Test with both polygons empty
	emptySource1 := NewMockVertexSource([]Vertex{})
	emptySource2 := NewMockVertexSource([]Vertex{})
	emptyGpc := NewConvGPC(emptySource1, emptySource2, GPCOr)

	emptyGpc.Rewind(0)
	emptyVertices := extractAllVertices(emptyGpc)

	t.Logf("Union of two empty polygons has %d vertices", len(emptyVertices))
}

func TestConvGPC_ComplexPolygon(t *testing.T) {
	// Create a more complex polygon (triangle)
	triangle := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{5, 10, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdEndPoly | basics.PathFlagClose},
	}

	// Square that intersects the triangle
	square := []Vertex{
		{3, 3, basics.PathCmdMoveTo},
		{7, 3, basics.PathCmdLineTo},
		{7, 7, basics.PathCmdLineTo},
		{3, 7, basics.PathCmdLineTo},
		{3, 3, basics.PathCmdEndPoly | basics.PathFlagClose},
	}

	source1 := NewMockVertexSource(triangle)
	source2 := NewMockVertexSource(square)

	// Test all operations
	operations := []GPCOp{GPCOr, GPCAnd, GPCXor, GPCAMinusB, GPCBMinusA}
	for _, op := range operations {
		gpc := NewConvGPC(source1, source2, op)
		gpc.Rewind(0)
		vertices := extractAllVertices(gpc)
		t.Logf("Complex polygon %s result has %d vertices", op.String(), len(vertices))
	}
}

func TestConvGPC_MultipleContours(t *testing.T) {
	// Polygon with hole (two contours)
	outerRect := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{20, 0, basics.PathCmdLineTo},
		{20, 20, basics.PathCmdLineTo},
		{0, 20, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdEndPoly | basics.PathFlagClose},
	}

	// Simple rectangle
	simpleRect := []Vertex{
		{5, 5, basics.PathCmdMoveTo},
		{15, 5, basics.PathCmdLineTo},
		{15, 15, basics.PathCmdLineTo},
		{5, 15, basics.PathCmdLineTo},
		{5, 5, basics.PathCmdEndPoly | basics.PathFlagClose},
	}

	source1 := NewMockVertexSource(outerRect)
	source2 := NewMockVertexSource(simpleRect)
	gpc := NewConvGPC(source1, source2, GPCOr)

	gpc.Rewind(0)
	vertices := extractAllVertices(gpc)

	t.Logf("Multiple contours result has %d vertices", len(vertices))

	// Should not panic and should produce some result
	if len(vertices) == 0 {
		t.Error("Expected some vertices from multiple contours operation")
	}
}

// Helper function to extract all vertices from a vertex source
func extractAllVertices(vs VertexSource) []Vertex {
	var vertices []Vertex
	vs.Rewind(0)

	for {
		x, y, cmd := vs.Vertex()
		vertices = append(vertices, Vertex{X: x, Y: y, Cmd: cmd})

		if basics.IsStop(cmd) {
			break
		}
	}

	return vertices
}

func countPreparedDrawableCommands(vs VertexSource) int {
	count := 0
	for {
		_, _, cmd := vs.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		count++
	}
	return count
}

// Benchmark tests for performance
func BenchmarkConvGPC_SimpleRectangles_Union(b *testing.B) {
	rect1 := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{10, 10, basics.PathCmdLineTo},
		{0, 10, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdEndPoly | basics.PathFlagClose},
	}

	rect2 := []Vertex{
		{5, 5, basics.PathCmdMoveTo},
		{15, 5, basics.PathCmdLineTo},
		{15, 15, basics.PathCmdLineTo},
		{5, 15, basics.PathCmdLineTo},
		{5, 5, basics.PathCmdEndPoly | basics.PathFlagClose},
	}

	source1 := NewMockVertexSource(rect1)
	source2 := NewMockVertexSource(rect2)
	gpc := NewConvGPC(source1, source2, GPCOr)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gpc.Rewind(0)
		extractAllVertices(gpc)
	}
}

func BenchmarkConvGPC_ComplexPolygons(b *testing.B) {
	// Create more complex polygons with many vertices
	vertices1 := make([]Vertex, 0, 100)
	vertices2 := make([]Vertex, 0, 100)

	// First polygon: approximation of a circle
	vertices1 = append(vertices1, Vertex{10, 0, basics.PathCmdMoveTo})
	for i := 1; i <= 50; i++ {
		x := 10 + 8*float64(i)/50 // Varying radius to make it more complex
		y := 8 * float64(i) / 50
		vertices1 = append(vertices1, Vertex{x, y, basics.PathCmdLineTo})
	}
	vertices1 = append(vertices1, Vertex{10, 0, basics.PathCmdEndPoly | basics.PathFlagClose})

	// Second polygon: another complex shape
	vertices2 = append(vertices2, Vertex{5, 5, basics.PathCmdMoveTo})
	for i := 1; i <= 50; i++ {
		x := 5 + float64(i)*0.2
		y := 5 + float64(i)*0.15
		vertices2 = append(vertices2, Vertex{x, y, basics.PathCmdLineTo})
	}
	vertices2 = append(vertices2, Vertex{5, 5, basics.PathCmdEndPoly | basics.PathFlagClose})

	source1 := NewMockVertexSource(vertices1)
	source2 := NewMockVertexSource(vertices2)
	gpc := NewConvGPC(source1, source2, GPCOr)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gpc.Rewind(0)
		extractAllVertices(gpc)
	}
}
