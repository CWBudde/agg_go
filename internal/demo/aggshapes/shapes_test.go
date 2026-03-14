package aggshapes

import (
	"testing"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/path"
)

func TestMakeArrows(t *testing.T) {
	ps := path.NewPathStorageStl()
	MakeArrows(ps)

	if got := ps.TotalVertices(); got != 32 {
		t.Fatalf("expected 32 vertices for four closed arrow polygons, got %d", got)
	}

	ps.Rewind(0)
	moveCount := 0
	endPolyCount := 0
	minX, minY := 1e9, 1e9
	maxX, maxY := -1e9, -1e9

	for {
		x, y, cmd := ps.NextVertex()
		pc := basics.PathCommand(cmd)
		if basics.IsStop(pc) {
			break
		}
		if basics.IsMoveTo(pc) {
			moveCount++
		}
		if basics.IsEndPoly(pc) {
			endPolyCount++
			continue
		}
		if x < minX {
			minX = x
		}
		if y < minY {
			minY = y
		}
		if x > maxX {
			maxX = x
		}
		if y > maxY {
			maxY = y
		}
	}

	if moveCount != 4 {
		t.Fatalf("expected 4 move_to commands, got %d", moveCount)
	}
	if endPolyCount != 4 {
		t.Fatalf("expected 4 end_poly commands, got %d", endPolyCount)
	}
	if minX != 1252.6 || minY != 1204.4 || maxX != 1393.0 || maxY != 1344.8 {
		t.Fatalf("unexpected bounds: got [%0.1f,%0.1f]-[%0.1f,%0.1f]", minX, minY, maxX, maxY)
	}
}

func TestMakeGBPolyPopulatesPath(t *testing.T) {
	ps := path.NewPathStorageStl()
	MakeGBPoly(ps)
	if got := ps.TotalVertices(); got == 0 {
		t.Fatal("expected Great Britain polygon path to contain vertices")
	}
}
