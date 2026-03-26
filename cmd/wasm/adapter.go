package main

import (
	"math"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
)

// pathSourceAdapter bridges PathStorageStl (uint Rewind) to the rasterizer's
// VertexSource interface (uint32 Rewind + pointer-based Vertex).
type pathSourceAdapter struct {
	ps *path.PathStorageStl
}

func (a *pathSourceAdapter) Rewind(pathID uint32) {
	a.ps.Rewind(uint(pathID))
}

func (a *pathSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ps.NextVertex()
	*x = vx
	*y = vy
	return cmd
}

// ellipseVS adapts shapes.Ellipse to the rasterizer's VertexSource interface.
type ellipseVS struct{ ell *shapes.Ellipse }

func (a *ellipseVS) Rewind(pathID uint32) { a.ell.Rewind(pathID) }
func (a *ellipseVS) Vertex(x, y *float64) uint32 { return uint32(a.ell.Vertex(x, y)) }

// ellipseConvVS adapts shapes.Ellipse to the conv.VertexSource interface
// (Rewind(uint), Vertex() returning (x, y float64, cmd basics.PathCommand)).
type ellipseConvVS struct{ ell *shapes.Ellipse }

func (a *ellipseConvVS) Rewind(pathID uint) { a.ell.Rewind(uint32(pathID)) }
func (a *ellipseConvVS) Vertex() (x, y float64, cmd basics.PathCommand) {
	var vx, vy float64
	c := a.ell.Vertex(&vx, &vy)
	return vx, vy, c
}

// getLionBoundingRect computes the axis-aligned bounding box of all vertices in
// the lion path data. It is shared by multiple demo files.
func getLionBoundingRect(ld *liondemo.LionData) (x1, y1, x2, y2 float64) {
	x1, y1, x2, y2 = math.MaxFloat64, math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64
	for idx := uint(0); idx < ld.Path.TotalVertices(); idx++ {
		x, y, cmd := ld.Path.Vertex(idx)
		if basics.IsVertex(basics.PathCommand(cmd)) {
			if x < x1 {
				x1 = x
			}
			if y < y1 {
				y1 = y
			}
			if x > x2 {
				x2 = x
			}
			if y > y2 {
				y2 = y
			}
		}
	}
	return
}
