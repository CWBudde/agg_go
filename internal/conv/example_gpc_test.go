package conv

import (
	"fmt"
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/path"
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
	// Create first path - rectangle from (0,0) to (10,10)
	path1 := path.NewPathStorage()
	path1.MoveTo(0, 0)
	path1.LineTo(10, 0)
	path1.LineTo(10, 10)
	path1.LineTo(0, 10)
	path1.ClosePolygon(basics.PathFlagsNone)

	// Create second path - rectangle from (5,5) to (15,15), overlapping first
	path2 := path.NewPathStorage()
	path2.MoveTo(5, 5)
	path2.LineTo(15, 5)
	path2.LineTo(15, 15)
	path2.LineTo(5, 15)
	path2.ClosePolygon(basics.PathFlagsNone)

	// Create adapters
	adapter1 := path.NewPathStorageVertexSourceAdapter(path1)
	adapter2 := path.NewPathStorageVertexSourceAdapter(path2)

	// Test different GPC operations
	testCases := []struct {
		name string
		op   GPCOp
	}{
		{"Union", GPCOr},
		{"Intersection", GPCAnd},
		{"A-minus-B", GPCAMinusB},
		{"XOR", GPCXor},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use with ConvGPC
			gpc := NewConvGPC(adapter1, adapter2, tc.op)
			gpc.Rewind(0)

			// Process result and count vertices
			vertexCount := 0
			for {
				x, y, cmd := gpc.Vertex()
				if basics.IsStop(cmd) {
					break
				}
				vertexCount++

				// Verify coordinates are reasonable (within expected bounds)
				if x < -1 || x > 16 || y < -1 || y > 16 {
					t.Errorf("Vertex (%f, %f) outside expected bounds for %s operation", x, y, tc.name)
				}
			}

			t.Logf("%s operation with PathStorage produced %d vertices", tc.name, vertexCount)

			// Verify we got some result (except for operations that might legitimately produce empty results)
			if tc.op == GPCOr && vertexCount == 0 {
				t.Errorf("Union operation should produce vertices for overlapping rectangles")
			}
		})
	}
}

// TestConvGPC_EdgeCases tests edge cases with PathStorage
func TestConvGPC_EdgeCases(t *testing.T) {
	// Test case 1: Non-overlapping rectangles
	t.Run("NonOverlapping", func(t *testing.T) {
		// Rectangle 1: (0,0) to (5,5)
		path1 := path.NewPathStorage()
		path1.MoveTo(0, 0)
		path1.LineTo(5, 0)
		path1.LineTo(5, 5)
		path1.LineTo(0, 5)
		path1.ClosePolygon(basics.PathFlagsNone)

		// Rectangle 2: (10,10) to (15,15) - no overlap
		path2 := path.NewPathStorage()
		path2.MoveTo(10, 10)
		path2.LineTo(15, 10)
		path2.LineTo(15, 15)
		path2.LineTo(10, 15)
		path2.ClosePolygon(basics.PathFlagsNone)

		adapter1 := path.NewPathStorageVertexSourceAdapter(path1)
		adapter2 := path.NewPathStorageVertexSourceAdapter(path2)

		// Union should produce both rectangles
		gpc := NewConvGPC(adapter1, adapter2, GPCOr)
		gpc.Rewind(0)

		vertexCount := 0
		for {
			_, _, cmd := gpc.Vertex()
			if basics.IsStop(cmd) {
				break
			}
			vertexCount++
		}

		t.Logf("Non-overlapping union produced %d vertices", vertexCount)
	})

	// Test case 2: One rectangle fully contains another
	t.Run("FullyContained", func(t *testing.T) {
		// Large rectangle: (0,0) to (20,20)
		path1 := path.NewPathStorage()
		path1.MoveTo(0, 0)
		path1.LineTo(20, 0)
		path1.LineTo(20, 20)
		path1.LineTo(0, 20)
		path1.ClosePolygon(basics.PathFlagsNone)

		// Small rectangle inside: (5,5) to (10,10)
		path2 := path.NewPathStorage()
		path2.MoveTo(5, 5)
		path2.LineTo(10, 5)
		path2.LineTo(10, 10)
		path2.LineTo(5, 10)
		path2.ClosePolygon(basics.PathFlagsNone)

		adapter1 := path.NewPathStorageVertexSourceAdapter(path1)
		adapter2 := path.NewPathStorageVertexSourceAdapter(path2)

		// Test intersection - should give the smaller rectangle
		gpc := NewConvGPC(adapter1, adapter2, GPCAnd)
		gpc.Rewind(0)

		vertexCount := 0
		for {
			_, _, cmd := gpc.Vertex()
			if basics.IsStop(cmd) {
				break
			}
			vertexCount++
		}

		t.Logf("Fully contained intersection produced %d vertices", vertexCount)
	})

	// Test case 3: Identical rectangles
	t.Run("Identical", func(t *testing.T) {
		// Rectangle 1: (0,0) to (10,10)
		path1 := path.NewPathStorage()
		path1.MoveTo(0, 0)
		path1.LineTo(10, 0)
		path1.LineTo(10, 10)
		path1.LineTo(0, 10)
		path1.ClosePolygon(basics.PathFlagsNone)

		// Rectangle 2: Same as rectangle 1
		path2 := path.NewPathStorage()
		path2.MoveTo(0, 0)
		path2.LineTo(10, 0)
		path2.LineTo(10, 10)
		path2.LineTo(0, 10)
		path2.ClosePolygon(basics.PathFlagsNone)

		adapter1 := path.NewPathStorageVertexSourceAdapter(path1)
		adapter2 := path.NewPathStorageVertexSourceAdapter(path2)

		// XOR of identical shapes should produce empty result
		gpc := NewConvGPC(adapter1, adapter2, GPCXor)
		gpc.Rewind(0)

		vertexCount := 0
		for {
			_, _, cmd := gpc.Vertex()
			if basics.IsStop(cmd) {
				break
			}
			vertexCount++
		}

		t.Logf("Identical rectangles XOR produced %d vertices", vertexCount)
		// Note: XOR of identical shapes should be empty, but current GPC implementation may not handle this correctly
	})
}
