package vpgen

import (
	"testing"

	"agg_go/internal/basics"
)

func TestVPGenClipPolyline_Basic(t *testing.T) {
	vpgen := NewVPGenClipPolyline()

	// Test default clip box
	if vpgen.X1() != 0 || vpgen.Y1() != 0 || vpgen.X2() != 1 || vpgen.Y2() != 1 {
		t.Errorf("Default clip box incorrect: got (%v,%v,%v,%v), want (0,0,1,1)",
			vpgen.X1(), vpgen.Y1(), vpgen.X2(), vpgen.Y2())
	}

	// Test AutoClose/AutoUnclose
	if vpgen.AutoClose() {
		t.Error("AutoClose should return false for polyline clipping")
	}
	if !vpgen.AutoUnclose() {
		t.Error("AutoUnclose should return true for polyline clipping")
	}
}

func TestVPGenClipPolyline_SetClipBox(t *testing.T) {
	vpgen := NewVPGenClipPolyline()
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

func TestVPGenClipPolyline_InsideLine(t *testing.T) {
	vpgen := NewVPGenClipPolyline()
	vpgen.SetClipBox(0, 0, 100, 100)

	// Line completely inside clip box
	vpgen.Reset()
	vpgen.MoveTo(25, 25)
	vpgen.LineTo(75, 75)

	// First vertex (MoveTo)
	x, y, cmd := vpgen.Vertex()
	if cmd != basics.PathCmdMoveTo || x != 25 || y != 25 {
		t.Errorf("First vertex: got (%v, %v, %v), want (25, 25, MoveTo)", x, y, cmd)
	}

	// Second vertex (LineTo)
	x, y, cmd = vpgen.Vertex()
	if cmd != basics.PathCmdLineTo || x != 75 || y != 75 {
		t.Errorf("Second vertex: got (%v, %v, %v), want (75, 75, LineTo)", x, y, cmd)
	}

	// End
	x, y, cmd = vpgen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Should end with Stop command, got %v", cmd)
	}
}

func TestVPGenClipPolyline_OutsideLine(t *testing.T) {
	vpgen := NewVPGenClipPolyline()
	vpgen.SetClipBox(0, 0, 100, 100)

	// Line completely outside clip box
	vpgen.Reset()
	vpgen.MoveTo(-25, -25)
	vpgen.LineTo(-75, -75)

	// Should produce no vertices
	x, y, cmd := vpgen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Outside line should produce no vertices, got (%v, %v, %v)", x, y, cmd)
	}
}

func TestVPGenClipPolyline_CrossingLine(t *testing.T) {
	vpgen := NewVPGenClipPolyline()
	vpgen.SetClipBox(0, 0, 100, 100)

	// Line crossing the clip box horizontally
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

	// Should have exactly 2 vertices (MoveTo at clip entry, LineTo at clip exit)
	if len(vertices) != 2 {
		t.Errorf("Expected 2 clipped vertices, got %d", len(vertices))
	}

	if len(vertices) >= 2 {
		// First vertex should be MoveTo at left edge
		if vertices[0].cmd != basics.PathCmdMoveTo || vertices[0].x != 0 || vertices[0].y != 50 {
			t.Errorf("First clipped vertex: got (%v, %v, %v), want (0, 50, MoveTo)",
				vertices[0].x, vertices[0].y, vertices[0].cmd)
		}

		// Second vertex should be LineTo at right edge
		if vertices[1].cmd != basics.PathCmdLineTo || vertices[1].x != 100 || vertices[1].y != 50 {
			t.Errorf("Second clipped vertex: got (%v, %v, %v), want (100, 50, LineTo)",
				vertices[1].x, vertices[1].y, vertices[1].cmd)
		}
	}
}

func TestVPGenClipPolyline_ContinuousPath(t *testing.T) {
	vpgen := NewVPGenClipPolyline()
	vpgen.SetClipBox(0, 0, 100, 100)

	// Path with multiple connected segments, some inside, some outside
	vpgen.Reset()
	vpgen.MoveTo(50, 50) // Start inside
	vpgen.LineTo(75, 75) // Inside line

	// Collect first segment
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

	// Should have MoveTo and LineTo
	if len(vertices) != 2 {
		t.Errorf("First segment: expected 2 vertices, got %d", len(vertices))
	}

	// Continue with line going outside
	vpgen.LineTo(150, 150) // Goes outside

	// Collect second segment
	vertices = []struct {
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

	// Should have only LineTo (clipped to edge)
	if len(vertices) != 1 {
		t.Errorf("Second segment: expected 1 vertex, got %d", len(vertices))
	}

	if len(vertices) >= 1 && vertices[0].cmd != basics.PathCmdLineTo {
		t.Errorf("Second segment should be LineTo, got %v", vertices[0].cmd)
	}
}

func TestVPGenClipPolyline_Reset(t *testing.T) {
	vpgen := NewVPGenClipPolyline()
	vpgen.SetClipBox(0, 0, 100, 100)

	// Add vertices
	vpgen.MoveTo(50, 50)
	vpgen.LineTo(75, 75)

	// Reset should clear state
	vpgen.Reset()

	// Should produce no vertices after reset
	x, y, cmd := vpgen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("After reset should produce Stop, got (%v, %v, %v)", x, y, cmd)
	}
}

func TestVPGenClipPolyline_PartiallyClippedStart(t *testing.T) {
	vpgen := NewVPGenClipPolyline()
	vpgen.SetClipBox(0, 0, 100, 100)

	// Line starting outside, ending inside
	vpgen.Reset()
	vpgen.MoveTo(-50, 50)
	vpgen.LineTo(50, 50)

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

	// Should have MoveTo at clip entry and LineTo at end point
	if len(vertices) != 2 {
		t.Errorf("Expected 2 vertices for partially clipped start, got %d", len(vertices))
	}

	if len(vertices) >= 2 {
		if vertices[0].cmd != basics.PathCmdMoveTo || vertices[0].x != 0 {
			t.Errorf("First vertex should be MoveTo at x=0, got (%v, %v, %v)",
				vertices[0].x, vertices[0].y, vertices[0].cmd)
		}
		if vertices[1].cmd != basics.PathCmdLineTo || vertices[1].x != 50 {
			t.Errorf("Second vertex should be LineTo at x=50, got (%v, %v, %v)",
				vertices[1].x, vertices[1].y, vertices[1].cmd)
		}
	}
}

func TestVPGenClipPolyline_PartiallyClippedEnd(t *testing.T) {
	vpgen := NewVPGenClipPolyline()
	vpgen.SetClipBox(0, 0, 100, 100)

	// Line starting inside, ending outside
	vpgen.Reset()
	vpgen.MoveTo(50, 50)
	vpgen.LineTo(150, 50)

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

	// Should have MoveTo at start and LineTo at clip exit
	if len(vertices) != 2 {
		t.Errorf("Expected 2 vertices for partially clipped end, got %d", len(vertices))
	}

	if len(vertices) >= 2 {
		if vertices[0].cmd != basics.PathCmdMoveTo || vertices[0].x != 50 {
			t.Errorf("First vertex should be MoveTo at x=50, got (%v, %v, %v)",
				vertices[0].x, vertices[0].y, vertices[0].cmd)
		}
		if vertices[1].cmd != basics.PathCmdLineTo || vertices[1].x != 100 {
			t.Errorf("Second vertex should be LineTo at x=100, got (%v, %v, %v)",
				vertices[1].x, vertices[1].y, vertices[1].cmd)
		}
	}
}
