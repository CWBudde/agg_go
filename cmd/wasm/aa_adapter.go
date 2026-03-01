package main

import (
	"agg_go/internal/path"
)

// AAAdapter bridges the gap between PathStorageStl (uint Rewind)
// and Rasterizer's expected VertexSource (uint32 Rewind + pointer Vertex).
type AAAdapter struct {
	ps *path.PathStorageStl
}

func (a *AAAdapter) Rewind(pathID uint32) {
	a.ps.Rewind(uint(pathID))
}

func (a *AAAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ps.NextVertex()
	*x = vx
	*y = vy
	return cmd
}
