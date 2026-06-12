package outline

import (
	"testing"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/primitives"
)

// gridBaseRenderer accumulates blended covers into a pixel grid.
type gridBaseRenderer struct {
	w, h  int
	cover [][]basics.CoverType
}

func newGridBaseRenderer(w, h int) *gridBaseRenderer {
	g := &gridBaseRenderer{w: w, h: h, cover: make([][]basics.CoverType, h)}
	for i := range g.cover {
		g.cover[i] = make([]basics.CoverType, w)
	}
	return g
}

func (g *gridBaseRenderer) Width() int  { return g.w }
func (g *gridBaseRenderer) Height() int { return g.h }

func (g *gridBaseRenderer) BlendSolidHSpan(x, y, length int, _ TestColor, covers []basics.CoverType) {
	for i := 0; i < length; i++ {
		if x+i >= 0 && x+i < g.w && y >= 0 && y < g.h && covers[i] > g.cover[y][x+i] {
			g.cover[y][x+i] = covers[i]
		}
	}
}

func (g *gridBaseRenderer) BlendSolidVSpan(x, y, length int, _ TestColor, covers []basics.CoverType) {
	for i := 0; i < length; i++ {
		if x >= 0 && x < g.w && y+i >= 0 && y+i < g.h && covers[i] > g.cover[y+i][x] {
			g.cover[y+i][x] = covers[i]
		}
	}
}

// TestLine3NoStaleDistanceGap is a regression test for a port bug where the
// LineInterpolatorAA1/2/3 step methods read DistStart/DistEnd BEFORE
// stepHorBase/stepVerBase advanced the distance interpolator (C++ calls
// step_hor_base(m_di) first). The stale distances gated out the first pixel
// column at line starts, leaving a one-pixel hole between a round cap and the
// line body (found via the alpha_mask2 demo, RMSE parity work).
func TestLine3NoStaleDistanceGap(t *testing.T) {
	const scale = primitives.LineSubpixelScale
	grid := newGridBaseRenderer(60, 30)

	profile := NewLineProfileAA()
	profile.Width(5.0)
	ren := NewRendererOutlineAA[*gridBaseRenderer, TestColor](grid, profile)
	ren.Color(TestColor{R: 0, G: 0, B: 0, A: 255})

	// Horizontal line (10,15)-(50,15), width 5, as the rasterizer would emit it.
	x1 := 10*scale + scale/2
	y1 := 15*scale + scale/2
	x2 := 50*scale + scale/2
	y2 := y1
	lp := primitives.NewLineParameters(x1, y1, x2, y2, int(basics.URound(40*scale)))
	ren.Line3(&lp, x1+(y2-y1), y1-(x2-x1), x2+(y2-y1), y2-(x2-x1))

	// The full-cover body of a width-5 line spans rows 13..17; every body row
	// must be covered in the first line column (x=10). With the stale reads,
	// rows 15..17 at x=10 received zero cover. (The final column x=50 is
	// correctly left to the end-cap semidot: dist_end > 0 gates it out.)
	for y := 13; y <= 17; y++ {
		if grid.cover[y][10] == 0 {
			t.Errorf("line body pixel (10,%d) has zero cover — first-column gap at line start", y)
		}
	}
	// The body row must be continuous up to the column before the endpoint.
	for x := 10; x <= 49; x++ {
		if grid.cover[15][x] == 0 {
			t.Errorf("line body pixel (%d,15) has zero cover — hole in line body", x)
		}
	}
}
