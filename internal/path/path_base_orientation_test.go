package path

import (
	"testing"

	"github.com/MeKo-Christian/agg_go/internal/basics"
)

func TestPerceivePolygonOrientation(t *testing.T) {
	t.Run("CW triangle", func(t *testing.T) {
		// In AGG's y-down coordinate system, a clockwise polygon has
		// negative shoelace area. Vertices: (0,0)->(10,10)->(10,0)
		// Area = 0*10 - 0*10 + 10*0 - 10*10 + 10*0 - 0*0 = -100 < 0 => CW
		ps := NewPathStorage()
		ps.MoveTo(0, 0)
		ps.LineTo(10, 10)
		ps.LineTo(10, 0)
		got := ps.PerceivePolygonOrientation(0, 3)
		if got != basics.PathFlagsCW {
			t.Errorf("expected PathFlagsCW (%d), got %d", basics.PathFlagsCW, got)
		}
	})

	t.Run("CCW triangle", func(t *testing.T) {
		// Vertices: (0,0)->(10,0)->(10,10)
		// Area = 0*0 - 0*10 + 10*10 - 0*10 + 10*0 - 10*0 = 100 > 0 => CCW
		ps := NewPathStorage()
		ps.MoveTo(0, 0)
		ps.LineTo(10, 0)
		ps.LineTo(10, 10)
		got := ps.PerceivePolygonOrientation(0, 3)
		if got != basics.PathFlagsCCW {
			t.Errorf("expected PathFlagsCCW (%d), got %d", basics.PathFlagsCCW, got)
		}
	})

	t.Run("CW square", func(t *testing.T) {
		// Vertices: (0,0)->(0,10)->(10,10)->(10,0) => CW (negative area)
		ps := NewPathStorage()
		ps.MoveTo(0, 0)
		ps.LineTo(0, 10)
		ps.LineTo(10, 10)
		ps.LineTo(10, 0)
		got := ps.PerceivePolygonOrientation(0, 4)
		if got != basics.PathFlagsCW {
			t.Errorf("expected PathFlagsCW (%d), got %d", basics.PathFlagsCW, got)
		}
	})
}

func TestInvertPolygonRange(t *testing.T) {
	t.Run("reverse triangle vertices", func(t *testing.T) {
		ps := NewPathStorage()
		ps.MoveTo(0, 0)   // idx 0
		ps.LineTo(10, 0)  // idx 1
		ps.LineTo(10, 10) // idx 2

		ps.InvertPolygonRange(0, 3)

		// After inversion: commands shift left, then SwapVertices swaps
		// both coords AND cmds. Net effect: coordinates reversed, commands
		// restored to original order [MoveTo, LineTo, LineTo].

		x0, y0, cmd0 := ps.Vertex(0)
		x1, y1, cmd1 := ps.Vertex(1)
		x2, y2, cmd2 := ps.Vertex(2)

		if x0 != 10 || y0 != 10 {
			t.Errorf("vertex 0: expected (10,10), got (%v,%v)", x0, y0)
		}
		if basics.PathCommand(cmd0) != basics.PathCmdMoveTo {
			t.Errorf("cmd 0: expected MoveTo (%d), got %d", basics.PathCmdMoveTo, cmd0)
		}

		if x1 != 10 || y1 != 0 {
			t.Errorf("vertex 1: expected (10,0), got (%v,%v)", x1, y1)
		}
		if basics.PathCommand(cmd1) != basics.PathCmdLineTo {
			t.Errorf("cmd 1: expected LineTo (%d), got %d", basics.PathCmdLineTo, cmd1)
		}

		if x2 != 0 || y2 != 0 {
			t.Errorf("vertex 2: expected (0,0), got (%v,%v)", x2, y2)
		}
		if basics.PathCommand(cmd2) != basics.PathCmdLineTo {
			t.Errorf("cmd 2: expected LineTo (%d), got %d", basics.PathCmdLineTo, cmd2)
		}
	})

	t.Run("inversion changes orientation", func(t *testing.T) {
		ps := NewPathStorage()
		ps.MoveTo(0, 0)
		ps.LineTo(10, 10)
		ps.LineTo(10, 0)
		before := ps.PerceivePolygonOrientation(0, 3)
		ps.InvertPolygonRange(0, 3)
		after := ps.PerceivePolygonOrientation(0, 3)
		if before == after {
			t.Errorf("expected orientation to change after invert, both are %d", before)
		}
	})
}

func TestArrangePolygonOrientation(t *testing.T) {
	t.Run("CCW polygon asked for CW is inverted", func(t *testing.T) {
		ps := NewPathStorage()
		// CCW triangle
		ps.MoveTo(0, 0)
		ps.LineTo(10, 0)
		ps.LineTo(10, 10)
		ps.EndPoly(basics.PathFlagsClose | basics.PathFlagsCCW)

		end := ps.ArrangePolygonOrientation(0, basics.PathFlagsCW)
		// Should return index past the EndPoly command
		if end != 4 {
			t.Errorf("expected end=4, got %d", end)
		}
		got := ps.PerceivePolygonOrientation(0, 3)
		if got != basics.PathFlagsCW {
			t.Errorf("expected CW after arrange, got %d", got)
		}
	})

	t.Run("CW polygon asked for CW stays same", func(t *testing.T) {
		ps := NewPathStorage()
		// CW triangle
		ps.MoveTo(0, 0)
		ps.LineTo(10, 10)
		ps.LineTo(10, 0)
		ps.EndPoly(basics.PathFlagsClose | basics.PathFlagsCW)

		x0, y0, _ := ps.Vertex(0)
		x1, y1, _ := ps.Vertex(1)

		ps.ArrangePolygonOrientation(0, basics.PathFlagsCW)

		ax0, ay0, _ := ps.Vertex(0)
		ax1, ay1, _ := ps.Vertex(1)
		if x0 != ax0 || y0 != ay0 || x1 != ax1 || y1 != ay1 {
			t.Error("vertices changed when orientation already matched")
		}
	})

	t.Run("PathFlagsNone is no-op", func(t *testing.T) {
		ps := NewPathStorage()
		ps.MoveTo(0, 0)
		ps.LineTo(10, 0)
		ps.LineTo(10, 10)
		end := ps.ArrangePolygonOrientation(0, basics.PathFlagsNone)
		if end != 0 {
			t.Errorf("expected start returned for PathFlagsNone, got %d", end)
		}
	})
}

func TestArrangeOrientations(t *testing.T) {
	t.Run("stops at Stop command", func(t *testing.T) {
		ps := NewPathStorage()
		// First path: one CCW polygon + stop
		ps.MoveTo(0, 0)
		ps.LineTo(10, 0)
		ps.LineTo(10, 10)
		ps.EndPoly(basics.PathFlagsClose | basics.PathFlagsCCW)
		ps.Vertices().AddVertex(0, 0, uint32(basics.PathCmdStop))
		// Second path: also CCW
		ps.MoveTo(20, 20)
		ps.LineTo(30, 20)
		ps.LineTo(30, 30)
		ps.EndPoly(basics.PathFlagsClose | basics.PathFlagsCCW)

		// ArrangeOrientations processes first path and returns past the Stop
		end := ps.ArrangeOrientations(0, basics.PathFlagsCW)
		if end != 5 {
			t.Errorf("expected end=5 (past stop), got %d", end)
		}

		// First polygon should be CW now
		got := ps.PerceivePolygonOrientation(0, 3)
		if got != basics.PathFlagsCW {
			t.Errorf("expected CW, got %d", got)
		}

		// Second polygon should be untouched (still CCW)
		got2 := ps.PerceivePolygonOrientation(5, 8)
		if got2 != basics.PathFlagsCCW {
			t.Errorf("second polygon should still be CCW, got %d", got2)
		}
	})
}

func TestArrangeOrientationsAllPaths(t *testing.T) {
	t.Run("two paths both arranged", func(t *testing.T) {
		ps := NewPathStorage()
		// First path: CCW triangle
		ps.MoveTo(0, 0)
		ps.LineTo(10, 0)
		ps.LineTo(10, 10)
		ps.EndPoly(basics.PathFlagsClose | basics.PathFlagsCCW)
		// Stop between paths
		ps.Vertices().AddVertex(0, 0, uint32(basics.PathCmdStop))
		// Second path: also CCW
		ps.MoveTo(20, 20)
		ps.LineTo(30, 20)
		ps.LineTo(30, 30)
		ps.EndPoly(basics.PathFlagsClose | basics.PathFlagsCCW)

		// Arrange all to CW
		ps.ArrangeOrientationsAllPaths(basics.PathFlagsCW)

		// Check first polygon is now CW
		got1 := ps.PerceivePolygonOrientation(0, 3)
		if got1 != basics.PathFlagsCW {
			t.Errorf("first polygon: expected CW, got %d", got1)
		}

		// Check second polygon is now CW
		got2 := ps.PerceivePolygonOrientation(5, 8)
		if got2 != basics.PathFlagsCW {
			t.Errorf("second polygon: expected CW, got %d", got2)
		}
	})

	t.Run("PathFlagsNone is no-op", func(t *testing.T) {
		ps := NewPathStorage()
		ps.MoveTo(0, 0)
		ps.LineTo(10, 0)
		ps.LineTo(10, 10)
		x0, y0, _ := ps.Vertex(0)
		ps.ArrangeOrientationsAllPaths(basics.PathFlagsNone)
		ax0, ay0, _ := ps.Vertex(0)
		if x0 != ax0 || y0 != ay0 {
			t.Error("PathFlagsNone should be a no-op")
		}
	})
}
