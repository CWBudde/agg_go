package font

import "github.com/MeKo-Christian/agg_go/internal/basics"

// IntegerPathStorage abstracts the integer outline storages used by the
// FreeType-backed engines for 16-bit and 32-bit glyph coordinate paths.
type IntegerPathStorage interface {
	// Core path commands
	RemoveAll()
	MoveTo64(x, y int64)
	LineTo64(x, y int64)
	Curve3_64(xCtrl, yCtrl, xTo, yTo int64)
	Curve4_64(xCtrl1, yCtrl1, xCtrl2, yCtrl2, xTo, yTo int64)
	ClosePolygon()

	// Query methods
	Size() uint32
	ByteSize() uint32
	Vertex(idx uint32) (float64, float64, basics.PathCommand)
}

// SerializedScanlinesAdaptor is the common read-only view over serialized AA and
// binary glyph scanlines.
type SerializedScanlinesAdaptor interface {
	Bounds() basics.Rect[int]
	Data() []byte
}
