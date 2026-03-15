package conv

import (
	"testing"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
	"github.com/MeKo-Christian/agg_go/internal/vcgen"
)

func TestConvMarkerWithTerminalMarkersMultipleSubpaths(t *testing.T) {
	markers := vcgen.NewVCGenMarkersTerm()
	markers.AddVertex(0, 0, basics.PathCmdMoveTo)
	markers.AddVertex(10, 0, basics.PathCmdLineTo)
	markers.AddVertex(20, 0, basics.PathCmdMoveTo)
	markers.AddVertex(30, 0, basics.PathCmdLineTo)

	ah := shapes.NewArrowhead()
	ah.Head(4, 4, 3, 2)
	ah.Tail(1, 1.5, 3, 5)

	cm := NewConvMarker(markers, &testArrowheadShapes{ah: ah})
	cm.Rewind(0)

	moveCount := 0
	for i := 0; i < 200; i++ {
		_, _, cmd := cm.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		if basics.IsMoveTo(cmd) {
			moveCount++
		}
	}

	if moveCount != 4 {
		t.Fatalf("expected 4 marker polygons, got %d", moveCount)
	}
}

type testArrowheadShapes struct{ ah *shapes.Arrowhead }

func (a *testArrowheadShapes) Rewind(shapeIndex uint) { a.ah.Rewind(uint32(shapeIndex)) }
func (a *testArrowheadShapes) Vertex() (x, y float64, cmd basics.PathCommand) {
	var vx, vy float64
	c := a.ah.Vertex(&vx, &vy)
	return vx, vy, c
}
