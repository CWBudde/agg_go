package conv

import "github.com/MeKo-Christian/agg_go/internal/rasterizer"

// RasterizerVertexSourceAdapter bridges a converter-style VertexSource to the
// rasterizer AddPath contract while preserving all path commands.
type RasterizerVertexSourceAdapter struct {
	source VertexSource
}

// NewRasterizerVertexSourceAdapter wraps a converter-style vertex source for rasterizer use.
func NewRasterizerVertexSourceAdapter(source VertexSource) *RasterizerVertexSourceAdapter {
	return &RasterizerVertexSourceAdapter{source: source}
}

// Rewind rewinds the wrapped source to the requested path.
func (a *RasterizerVertexSourceAdapter) Rewind(pathID uint32) {
	a.source.Rewind(uint(pathID))
}

// Vertex forwards the next vertex and command unchanged.
func (a *RasterizerVertexSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.source.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

var _ rasterizer.VertexSource = (*RasterizerVertexSourceAdapter)(nil)
