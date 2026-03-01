package main

import (
	"agg_go/internal/path"
)

// WasmAdapter bridges the gap between PathStorageStl (uint Rewind)
// and Rasterizer's expected VertexSource (uint32 Rewind + pointer Vertex).
type WasmAdapter struct {
	ps *path.PathStorageStl
}

func (a *WasmAdapter) Rewind(pathID uint32) {
	a.ps.Rewind(uint(pathID))
}

func (a *WasmAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ps.NextVertex()
	*x = vx
	*y = vy
	return cmd
}

// GouraudAdapter bridges SpanGouraudRGBA (uint Rewind)
// to Rasterizer's expected VertexSource (uint32 Rewind + pointer Vertex).
type GouraudAdapter struct {
	sg VertexSourceUint
}

type VertexSourceUint interface {
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd uint32)
}

func (a *GouraudAdapter) Rewind(pathID uint32) {
	a.sg.Rewind(uint(pathID))
}

func (a *GouraudAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.sg.Vertex()
	*x = vx
	*y = vy
	return cmd
}
