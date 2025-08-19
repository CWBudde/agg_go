package vcgen

import (
	"testing"

	"agg_go/internal/basics"
)

func TestVPGenClipPolygon_Creation(t *testing.T) {
	vpgen := NewVPGenClipPolygon()

	// Test default clipping box (0, 0, 1, 1)
	if vpgen.X1() != 0 || vpgen.Y1() != 0 || vpgen.X2() != 1 || vpgen.Y2() != 1 {
		t.Errorf("Expected default clip box (0,0,1,1), got (%f,%f,%f,%f)",
			vpgen.X1(), vpgen.Y1(), vpgen.X2(), vpgen.Y2())
	}

	// Test auto close/unclose behavior
	if !vpgen.AutoClose() {
		t.Error("Expected AutoClose() to return true")
	}
	if vpgen.AutoUnclose() {
		t.Error("Expected AutoUnclose() to return false")
	}
}

func TestVPGenClipPolygon_ClipBox(t *testing.T) {
	vpgen := NewVPGenClipPolygon()

	// Test setting clip box
	vpgen.ClipBox(10, 20, 100, 200)
	if vpgen.X1() != 10 || vpgen.Y1() != 20 || vpgen.X2() != 100 || vpgen.Y2() != 200 {
		t.Errorf("Expected clip box (10,20,100,200), got (%f,%f,%f,%f)",
			vpgen.X1(), vpgen.Y1(), vpgen.X2(), vpgen.Y2())
	}

	// Test with inverted coordinates (should normalize)
	vpgen.ClipBox(100, 200, 10, 20)
	if vpgen.X1() != 10 || vpgen.Y1() != 20 || vpgen.X2() != 100 || vpgen.Y2() != 200 {
		t.Errorf("Expected normalized clip box (10,20,100,200), got (%f,%f,%f,%f)",
			vpgen.X1(), vpgen.Y1(), vpgen.X2(), vpgen.Y2())
	}
}

func TestVPGenClipPolygon_ClippingFlags(t *testing.T) {
	vpgen := NewVPGenClipPolygon()
	vpgen.ClipBox(10, 20, 100, 200)

	testCases := []struct {
		x, y         float64
		expectedFlag uint32
		description  string
	}{
		{50, 100, 0, "inside"},
		{200, 100, 1, "right of clip box"},
		{50, 300, 2, "above clip box"},
		{200, 300, 3, "right and above"},
		{5, 100, 4, "left of clip box"},
		{5, 300, 6, "left and above"},
		{50, 10, 8, "below clip box"},
		{200, 10, 9, "right and below"},
		{5, 10, 12, "left and below"},
	}

	for _, tc := range testCases {
		flag := vpgen.clippingFlags(tc.x, tc.y)
		if flag != tc.expectedFlag {
			t.Errorf("Point (%f,%f) %s: expected flag %d, got %d",
				tc.x, tc.y, tc.description, tc.expectedFlag, flag)
		}
	}
}

// Test VPGenClipPolygon through the intended usage pattern
func TestVPGenClipPolygon_ThroughAdaptor(t *testing.T) {
	// This test uses VPGenClipPolygon the way it's intended to be used
	// through ConvAdaptorVPGen, not directly
	vpgen := NewVPGenClipPolygon()
	vpgen.ClipBox(0, 0, 100, 100)

	// Simulate what ConvAdaptorVPGen does for a simple line inside the clip box
	vpgen.MoveTo(20, 30) // Point inside - should be stored
	vpgen.LineTo(80, 70) // Point inside - should be stored

	// Now get vertices
	x, y, cmd := vpgen.Vertex()
	if cmd == basics.PathCmdStop {
		t.Error("Expected vertex, got stop")
	}

	// The implementation stores only one vertex at a time
	// According to the C++ code, LineTo overwrites the MoveTo vertex
	// when both are inside, so we should get the LineTo endpoint
	if x != 80 || y != 70 {
		t.Errorf("Expected (80,70), got (%f,%f)", x, y)
	}

	// Next call should return stop
	_, _, cmd = vpgen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected stop after first vertex, got %v", cmd)
	}
}

func TestVPGenClipPolygon_PointOutside(t *testing.T) {
	vpgen := NewVPGenClipPolygon()
	vpgen.ClipBox(0, 0, 100, 100)

	// Move to point outside
	vpgen.MoveTo(200, 200)

	// Should have no vertices to return
	_, _, cmd := vpgen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected stop for point outside, got %v", cmd)
	}
}

func TestVPGenClipPolygon_LineClipping(t *testing.T) {
	vpgen := NewVPGenClipPolygon()
	vpgen.ClipBox(0, 0, 100, 100)

	// Line from inside to outside that should be clipped
	vpgen.MoveTo(50, 50)   // inside
	vpgen.LineTo(150, 150) // outside

	// Should get clipped vertices from Liang-Barsky
	vertexCount := 0
	for {
		x, y, cmd := vpgen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertexCount++

		// Clipped vertices should be within bounds
		if x < 0 || x > 100 || y < 0 || y > 100 {
			t.Errorf("Clipped vertex (%f,%f) is outside clip box", x, y)
		}

		// Safety check
		if vertexCount > 10 {
			t.Fatal("Too many vertices - infinite loop?")
		}
	}

	// Should have generated some clipped vertices
	if vertexCount == 0 {
		t.Error("Expected clipped vertices, got none")
	}
}

func TestVPGenClipPolygon_Reset(t *testing.T) {
	vpgen := NewVPGenClipPolygon()
	vpgen.ClipBox(0, 0, 100, 100)

	// Set up some state
	vpgen.MoveTo(50, 50)

	// Reset
	vpgen.Reset()

	// Should have no vertices to return after reset
	_, _, cmd := vpgen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected stop after reset, got %v", cmd)
	}
}

func TestVPGenClipPolygon_ClippedLineSegment(t *testing.T) {
	vpgen := NewVPGenClipPolygon()
	vpgen.ClipBox(0, 0, 100, 100)

	// Line from outside to outside that passes through clip box
	vpgen.MoveTo(-50, 50) // outside left
	vpgen.LineTo(150, 50) // outside right

	// Should get clipped segment
	vertices := []struct {
		x, y float64
		cmd  basics.PathCommand
	}{}
	for {
		x, y, cmd := vpgen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertices = append(vertices, struct {
			x, y float64
			cmd  basics.PathCommand
		}{x, y, cmd})
		if len(vertices) > 10 { // Safety
			break
		}
	}

	if len(vertices) == 0 {
		t.Error("Expected clipped line segment, got no vertices")
	}

	// All vertices should be on the clipping boundary or inside
	for _, v := range vertices {
		if v.x < 0 || v.x > 100 || v.y < 0 || v.y > 100 {
			t.Errorf("Vertex (%f,%f) is outside clip bounds", v.x, v.y)
		}
	}
}
