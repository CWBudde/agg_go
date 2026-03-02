package main

import (
	"agg_go/internal/path"
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
