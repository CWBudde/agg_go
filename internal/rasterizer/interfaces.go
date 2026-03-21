package rasterizer

// VertexSource is the minimal AGG-style vertex-source contract consumed by
// rasterizers and outline renderers.
type VertexSource interface {
	Rewind(pathID uint32)
	Vertex(x, y *float64) uint32
}

// ScanlineInterface is the span-accumulation contract expected during scanline
// sweeping. The rasterizer writes coverage data; the renderer reads it back.
// Deprecated: prefer scanline.Scanline for new code.
type ScanlineInterface interface {
	ResetSpans()
	AddCell(x int, cover uint)
	AddSpan(x, len int, cover uint)
	Finalize(y int)
	NumSpans() int
}
