package rasterizer

// VertexSource interface for vertex generators
type VertexSource interface {
	Rewind(pathID uint32)
	Vertex(x, y *float64) uint32
}

// ScanlineInterface defines the interface for scanline objects
type ScanlineInterface interface {
	ResetSpans()
	AddCell(x int, cover uint32)
	AddSpan(x, len int, cover uint32)
	Finalize(y int)
	NumSpans() int
}
