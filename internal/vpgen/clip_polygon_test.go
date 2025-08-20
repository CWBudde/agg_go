package vpgen

import (
	"testing"

	"agg_go/internal/basics"
)

func TestVPGenClipPolygon_Basic(t *testing.T) {
	vpgen := NewVPGenClipPolygon()

	// Test default clip box
	if vpgen.X1() != 0 || vpgen.Y1() != 0 || vpgen.X2() != 1 || vpgen.Y2() != 1 {
		t.Errorf("Default clip box incorrect: got (%v,%v,%v,%v), want (0,0,1,1)",
			vpgen.X1(), vpgen.Y1(), vpgen.X2(), vpgen.Y2())
	}

	// Test AutoClose/AutoUnclose
	if !vpgen.AutoClose() {
		t.Error("AutoClose should return true for polygon clipping")
	}
	if vpgen.AutoUnclose() {
		t.Error("AutoUnclose should return false for polygon clipping")
	}
}

func TestVPGenClipPolygon_SetClipBox(t *testing.T) {
	vpgen := NewVPGenClipPolygon()
	vpgen.SetClipBox(10, 20, 100, 200)

	if vpgen.X1() != 10 || vpgen.Y1() != 20 || vpgen.X2() != 100 || vpgen.Y2() != 200 {
		t.Errorf("SetClipBox failed: got (%v,%v,%v,%v), want (10,20,100,200)",
			vpgen.X1(), vpgen.Y1(), vpgen.X2(), vpgen.Y2())
	}

	// Test normalization
	vpgen.SetClipBox(200, 100, 20, 10)
	if vpgen.X1() != 20 || vpgen.Y1() != 10 || vpgen.X2() != 200 || vpgen.Y2() != 100 {
		t.Errorf("Clip box normalization failed: got (%v,%v,%v,%v), want (20,10,200,100)",
			vpgen.X1(), vpgen.Y1(), vpgen.X2(), vpgen.Y2())
	}
}

func TestVPGenClipPolygon_ClippingFlags(t *testing.T) {
	vpgen := NewVPGenClipPolygon()
	vpgen.SetClipBox(0, 0, 100, 100)

	tests := []struct {
		name     string
		x, y     float64
		expected uint32
	}{
		{"center", 50, 50, 0},         // inside
		{"right", 150, 50, 1},         // right
		{"top", 50, 150, 2},           // top
		{"top-right", 150, 150, 3},    // top-right
		{"left", -50, 50, 4},          // left
		{"top-left", -50, 150, 6},     // top-left
		{"bottom", 50, -50, 8},        // bottom
		{"bottom-right", 150, -50, 9}, // bottom-right
		{"bottom-left", -50, -50, 12}, // bottom-left
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vpgen.clippingFlags(tt.x, tt.y)
			if result != tt.expected {
				t.Errorf("clippingFlags(%v, %v) = %v, want %v", tt.x, tt.y, result, tt.expected)
			}
		})
	}
}

func TestVPGenClipPolygon_InsideTriangle(t *testing.T) {
	vpgen := NewVPGenClipPolygon()
	vpgen.SetClipBox(0, 0, 100, 100)

	// Triangle completely inside clip box
	vpgen.Reset()
	vpgen.MoveTo(25, 25)

	// First vertex (MoveTo)
	x, y, cmd := vpgen.Vertex()
	if cmd != basics.PathCmdMoveTo || x != 25 || y != 25 {
		t.Errorf("First vertex: got (%v, %v, %v), want (25, 25, MoveTo)", x, y, cmd)
	}

	// Process second line
	vpgen.LineTo(75, 25)

	// Second vertex (LineTo)
	x, y, cmd = vpgen.Vertex()
	if cmd != basics.PathCmdLineTo || x != 75 || y != 25 {
		t.Errorf("Second vertex: got (%v, %v, %v), want (75, 25, LineTo)", x, y, cmd)
	}

	// Process third line
	vpgen.LineTo(50, 75)

	// Third vertex (LineTo)
	x, y, cmd = vpgen.Vertex()
	if cmd != basics.PathCmdLineTo || x != 50 || y != 75 {
		t.Errorf("Third vertex: got (%v, %v, %v), want (50, 75, LineTo)", x, y, cmd)
	}

	// End
	x, y, cmd = vpgen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Should end with Stop command, got %v", cmd)
	}
}

func TestVPGenClipPolygon_OutsideTriangle(t *testing.T) {
	vpgen := NewVPGenClipPolygon()
	vpgen.SetClipBox(0, 0, 100, 100)

	// Triangle completely outside clip box
	vpgen.Reset()
	vpgen.MoveTo(-25, -25)
	vpgen.LineTo(-75, -25)
	vpgen.LineTo(-50, -75)

	// Should produce no vertices
	x, y, cmd := vpgen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Outside triangle should produce no vertices, got (%v, %v, %v)", x, y, cmd)
	}
}

func TestVPGenClipPolygon_ClippedLine(t *testing.T) {
	vpgen := NewVPGenClipPolygon()
	vpgen.SetClipBox(0, 0, 100, 100)

	// Line crossing the clip box
	vpgen.Reset()
	vpgen.MoveTo(-50, 50)
	vpgen.LineTo(150, 50)

	// Should produce clipped line
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
	}

	// Should have exactly 2 vertices (entry and exit points)
	if len(vertices) != 2 {
		t.Errorf("Expected 2 clipped vertices, got %d", len(vertices))
	}

	if len(vertices) >= 2 {
		// First vertex should be at left edge
		if vertices[0].x != 0 || vertices[0].y != 50 || vertices[0].cmd != basics.PathCmdMoveTo {
			t.Errorf("First clipped vertex: got (%v, %v, %v), want (0, 50, MoveTo)",
				vertices[0].x, vertices[0].y, vertices[0].cmd)
		}

		// Second vertex should be at right edge
		if vertices[1].x != 100 || vertices[1].y != 50 || vertices[1].cmd != basics.PathCmdLineTo {
			t.Errorf("Second clipped vertex: got (%v, %v, %v), want (100, 50, LineTo)",
				vertices[1].x, vertices[1].y, vertices[1].cmd)
		}
	}
}

func TestVPGenClipPolygon_Reset(t *testing.T) {
	vpgen := NewVPGenClipPolygon()
	vpgen.SetClipBox(0, 0, 100, 100)

	// Add a vertex
	vpgen.MoveTo(50, 50)

	// Reset should clear state
	vpgen.Reset()

	// Should produce no vertices after reset
	x, y, cmd := vpgen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("After reset should produce Stop, got (%v, %v, %v)", x, y, cmd)
	}
}

func TestVPGenClipPolygon_MultipleSubsequentCalls(t *testing.T) {
	vpgen := NewVPGenClipPolygon()
	vpgen.SetClipBox(0, 0, 100, 100)

	// First path
	vpgen.Reset()
	vpgen.MoveTo(25, 25)
	vpgen.Vertex() // MoveTo

	vpgen.LineTo(75, 75)
	vpgen.Vertex() // LineTo

	// Second path
	vpgen.Reset()
	vpgen.MoveTo(10, 10)

	// Should get fresh vertices
	x, y, cmd := vpgen.Vertex()
	if cmd != basics.PathCmdMoveTo || x != 10 || y != 10 {
		t.Errorf("Second path first vertex: got (%v, %v, %v), want (10, 10, MoveTo)", x, y, cmd)
	}

	vpgen.LineTo(90, 90)
	x, y, cmd = vpgen.Vertex()
	if cmd != basics.PathCmdLineTo || x != 90 || y != 90 {
		t.Errorf("Second path second vertex: got (%v, %v, %v), want (90, 90, LineTo)", x, y, cmd)
	}
}
