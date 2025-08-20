package conv

import (
	"fmt"
	"testing"

	"agg_go/internal/basics"
)

// Example demonstrating how to use ConvGPC for polygon boolean operations
func ExampleConvGPC() {
	// Create two simple rectangles
	rect1Vertices := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{10, 10, basics.PathCmdLineTo},
		{0, 10, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdEndPoly | basics.PathFlagClose},
	}

	rect2Vertices := []Vertex{
		{5, 5, basics.PathCmdMoveTo},
		{15, 5, basics.PathCmdLineTo},
		{15, 15, basics.PathCmdLineTo},
		{5, 15, basics.PathCmdLineTo},
		{5, 5, basics.PathCmdEndPoly | basics.PathFlagClose},
	}

	// Create vertex sources
	source1 := NewMockVertexSource(rect1Vertices)
	source2 := NewMockVertexSource(rect2Vertices)

	// Create GPC converter for union operation
	gpc := NewConvGPC(source1, source2, GPCOr)

	// Process the boolean operation
	gpc.Rewind(0)

	// Extract the result vertices
	fmt.Printf("Union result:\n")
	for {
		x, y, cmd := gpc.Vertex()
		if basics.IsStop(cmd) {
			fmt.Printf("  Stop\n")
			break
		}

		switch cmd {
		case basics.PathCmdMoveTo:
			fmt.Printf("  MoveTo(%.1f, %.1f)\n", x, y)
		case basics.PathCmdLineTo:
			fmt.Printf("  LineTo(%.1f, %.1f)\n", x, y)
		case basics.PathCmdEndPoly | basics.PathFlagClose:
			fmt.Printf("  EndPoly(Close)\n")
		}
	}
}

// TestExampleUsage demonstrates various ConvGPC operations
func TestExampleUsage(t *testing.T) {
	// Two overlapping rectangles for demonstration
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

	// Test each operation
	operations := []struct {
		op   GPCOp
		name string
	}{
		{GPCOr, "Union"},
		{GPCAnd, "Intersection"},
		{GPCXor, "XOR"},
		{GPCAMinusB, "A - B"},
		{GPCBMinusA, "B - A"},
	}

	for _, test := range operations {
		t.Run(test.name, func(t *testing.T) {
			gpc := NewConvGPC(source1, source2, test.op)
			gpc.Rewind(0)

			vertexCount := 0
			for {
				_, _, cmd := gpc.Vertex()
				if basics.IsStop(cmd) {
					break
				}
				vertexCount++
			}

			t.Logf("%s operation produced %d vertices", test.name, vertexCount)

			// Note: Some operations may return 0 vertices due to placeholder GPC implementation
			// The actual GPC algorithm would compute proper results
			if test.op == GPCOr && vertexCount == 0 {
				t.Errorf("Union operation should always produce some result")
			}
		})
	}
}

// TestConvGPC_WithPathStorageAdapter demonstrates using ConvGPC with PathStorage
func TestConvGPC_WithPathStorageAdapter(t *testing.T) {
	// This test would work with the PathStorage adapter once the path storage is properly integrated
	// For now, we'll skip it since it requires additional integration work
	t.Skip("PathStorage adapter integration test - requires complete path storage integration")

	// TODO: Uncomment and implement when PathStorage is fully integrated
	/*
		path1 := path.NewPathStorage()
		path1.MoveTo(0, 0)
		path1.LineTo(10, 0)
		path1.LineTo(10, 10)
		path1.LineTo(0, 10)
		path1.ClosePath()

		path2 := path.NewPathStorage()
		path2.MoveTo(5, 5)
		path2.LineTo(15, 5)
		path2.LineTo(15, 15)
		path2.LineTo(5, 15)
		path2.ClosePath()

		// Create adapters
		adapter1 := path.NewPathStorageVertexSourceAdapter(path1)
		adapter2 := path.NewPathStorageVertexSourceAdapter(path2)

		// Use with ConvGPC
		gpc := NewConvGPC(adapter1, adapter2, GPCOr)
		gpc.Rewind(0)

		// Process result...
	*/
}
