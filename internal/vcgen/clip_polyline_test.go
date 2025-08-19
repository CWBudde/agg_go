package vcgen

import (
	"testing"

	"agg_go/internal/basics"
)

func TestVPGenClipPolyline_Basic(t *testing.T) {
	vpgen := NewVPGenClipPolyline()
	vpgen.ClipBox(0, 0, 10, 10)

	// Test line completely inside clipping box
	vpgen.Reset()
	vpgen.MoveTo(2, 2)
	vpgen.LineTo(8, 8)

	// Should get MoveTo(2,2) + LineTo(8,8) for a line completely inside
	x, y, cmd := vpgen.Vertex()
	if cmd != basics.PathCmdMoveTo || x != 2 || y != 2 {
		t.Errorf("Expected move_to (2,2), got %v (%f,%f)", cmd, x, y)
	}

	x, y, cmd = vpgen.Vertex()
	if cmd != basics.PathCmdLineTo || x != 8 || y != 8 {
		t.Errorf("Expected line_to (8,8), got %v (%f,%f)", cmd, x, y)
	}

	x, y, cmd = vpgen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected stop, got %v", cmd)
	}
}

func TestVPGenClipPolyline_CompletelyOutside(t *testing.T) {
	vpgen := NewVPGenClipPolyline()
	vpgen.ClipBox(0, 0, 10, 10)

	// Test line completely outside clipping box
	vpgen.Reset()
	vpgen.MoveTo(20, 20)
	vpgen.LineTo(30, 30)

	x, y, cmd := vpgen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected stop for completely clipped line, got %v (%f,%f)", cmd, x, y)
	}
}

func TestVPGenClipPolyline_PartialClipping(t *testing.T) {
	vpgen := NewVPGenClipPolyline()
	vpgen.ClipBox(0, 0, 10, 10)

	// Test line crossing right edge (inside->outside)
	vpgen.Reset()
	vpgen.MoveTo(5, 5)
	vpgen.LineTo(15, 5)

	// Should get MoveTo(5,5) + LineTo(10,5) - MoveTo because this is first segment after MoveTo
	x, y, cmd := vpgen.Vertex()
	if cmd != basics.PathCmdMoveTo || x != 5 || y != 5 {
		t.Errorf("Expected move_to (5,5), got %v (%f,%f)", cmd, x, y)
	}

	x, y, cmd = vpgen.Vertex()
	if cmd != basics.PathCmdLineTo || x != 10 || y != 5 {
		t.Errorf("Expected clipped line_to (10,5), got %v (%f,%f)", cmd, x, y)
	}

	x, y, cmd = vpgen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected stop, got %v", cmd)
	}
}

func TestVPGenClipPolyline_BothEndpointsClipped(t *testing.T) {
	vpgen := NewVPGenClipPolyline()
	vpgen.ClipBox(0, 0, 10, 10)

	// Test line crossing from left to right
	vpgen.Reset()
	vpgen.MoveTo(-5, 5)
	vpgen.LineTo(15, 5)

	x, y, cmd := vpgen.Vertex()
	if cmd != basics.PathCmdMoveTo || x != 0 || y != 5 {
		t.Errorf("Expected clipped move_to (0,5), got %v (%f,%f)", cmd, x, y)
	}

	x, y, cmd = vpgen.Vertex()
	if cmd != basics.PathCmdLineTo || x != 10 || y != 5 {
		t.Errorf("Expected clipped line_to (10,5), got %v (%f,%f)", cmd, x, y)
	}

	x, y, cmd = vpgen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected stop, got %v", cmd)
	}
}

func TestVPGenClipPolyline_MultipleSegments(t *testing.T) {
	vpgen := NewVPGenClipPolyline()
	vpgen.ClipBox(0, 0, 10, 10)

	// Test polyline with multiple segments
	vpgen.Reset()
	vpgen.MoveTo(2, 2)
	vpgen.LineTo(8, 2) // inside -> inside

	// First segment should produce move_to + line_to
	x, y, cmd := vpgen.Vertex()
	if cmd != basics.PathCmdMoveTo || x != 2 || y != 2 {
		t.Errorf("Expected move_to (2,2), got %v (%f,%f)", cmd, x, y)
	}

	x, y, cmd = vpgen.Vertex()
	if cmd != basics.PathCmdLineTo || x != 8 || y != 2 {
		t.Errorf("Expected line_to (8,2), got %v (%f,%f)", cmd, x, y)
	}

	x, y, cmd = vpgen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected stop after first segment, got %v", cmd)
	}

	// Second segment: inside -> outside (no moveTo flag set, so just line_to)
	vpgen.LineTo(15, 2)

	x, y, cmd = vpgen.Vertex()
	if cmd != basics.PathCmdLineTo || x != 10 || y != 2 {
		t.Errorf("Expected clipped line_to (10,2), got %v (%f,%f)", cmd, x, y)
	}

	x, y, cmd = vpgen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected stop after second segment, got %v", cmd)
	}

	// Third segment: outside -> inside (should generate move_to + line_to)
	vpgen.LineTo(5, 2)

	x, y, cmd = vpgen.Vertex()
	if cmd != basics.PathCmdMoveTo || x != 10 || y != 2 {
		t.Errorf("Expected move_to (10,2), got %v (%f,%f)", cmd, x, y)
	}

	x, y, cmd = vpgen.Vertex()
	if cmd != basics.PathCmdLineTo || x != 5 || y != 2 {
		t.Errorf("Expected line_to (5,2), got %v (%f,%f)", cmd, x, y)
	}

	x, y, cmd = vpgen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected stop after third segment, got %v", cmd)
	}
}

func TestVPGenClipPolyline_EdgeCases(t *testing.T) {
	vpgen := NewVPGenClipPolyline()
	vpgen.ClipBox(0, 0, 10, 10)

	// Test line on boundary
	vpgen.Reset()
	vpgen.MoveTo(0, 5)
	vpgen.LineTo(10, 5)

	// Should get MoveTo + LineTo for line on boundary
	x, y, cmd := vpgen.Vertex()
	if cmd != basics.PathCmdMoveTo || x != 0 || y != 5 {
		t.Errorf("Expected move_to (0,5), got %v (%f,%f)", cmd, x, y)
	}

	x, y, cmd = vpgen.Vertex()
	if cmd != basics.PathCmdLineTo || x != 10 || y != 5 {
		t.Errorf("Expected line_to (10,5), got %v (%f,%f)", cmd, x, y)
	}

	x, y, cmd = vpgen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected stop, got %v", cmd)
	}
}

func TestVPGenClipPolyline_ZeroLengthSegment(t *testing.T) {
	vpgen := NewVPGenClipPolyline()
	vpgen.ClipBox(0, 0, 10, 10)

	// Test zero-length segment (same start and end point)
	vpgen.Reset()
	vpgen.MoveTo(5, 5)
	vpgen.LineTo(5, 5)

	// Should get MoveTo + LineTo even for zero-length segment
	x, y, cmd := vpgen.Vertex()
	if cmd != basics.PathCmdMoveTo || x != 5 || y != 5 {
		t.Errorf("Expected move_to (5,5), got %v (%f,%f)", cmd, x, y)
	}

	x, y, cmd = vpgen.Vertex()
	if cmd != basics.PathCmdLineTo || x != 5 || y != 5 {
		t.Errorf("Expected line_to (5,5), got %v (%f,%f)", cmd, x, y)
	}

	x, y, cmd = vpgen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected stop, got %v", cmd)
	}
}

func TestVPGenClipPolyline_ClipBoxAccessors(t *testing.T) {
	vpgen := NewVPGenClipPolyline()
	vpgen.ClipBox(1, 2, 8, 9)

	if vpgen.X1() != 1 {
		t.Errorf("Expected X1=1, got %f", vpgen.X1())
	}
	if vpgen.Y1() != 2 {
		t.Errorf("Expected Y1=2, got %f", vpgen.Y1())
	}
	if vpgen.X2() != 8 {
		t.Errorf("Expected X2=8, got %f", vpgen.X2())
	}
	if vpgen.Y2() != 9 {
		t.Errorf("Expected Y2=9, got %f", vpgen.Y2())
	}
}

func TestVPGenClipPolyline_AutoCloseUnclose(t *testing.T) {
	vpgen := NewVPGenClipPolyline()

	// Polylines should not auto-close but should auto-unclose
	if vpgen.AutoClose() {
		t.Error("Polylines should not auto-close")
	}
	if !vpgen.AutoUnclose() {
		t.Error("Polylines should auto-unclose")
	}
}

func TestVPGenClipPolyline_ClipBoxNormalization(t *testing.T) {
	vpgen := NewVPGenClipPolyline()
	// Set clip box with reversed coordinates
	vpgen.ClipBox(10, 10, 0, 0)

	// Should be normalized to (0,0,10,10)
	if vpgen.X1() != 0 || vpgen.Y1() != 0 || vpgen.X2() != 10 || vpgen.Y2() != 10 {
		t.Errorf("Expected normalized clip box (0,0,10,10), got (%f,%f,%f,%f)",
			vpgen.X1(), vpgen.Y1(), vpgen.X2(), vpgen.Y2())
	}
}
