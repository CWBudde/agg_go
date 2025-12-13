package font

import "agg_go/internal/basics"

// IntegerPathStorage defines the common interface for integer-based path storage.
// Both PathStorageInteger[int16] and PathStorageInteger[int32] implement this.
// This eliminates the need for interface{} when handling dual-precision paths.
//
// The interface uses int64 parameters which are wide enough to accommodate both
// int16 and int32 values without loss of precision. Implementations convert to
// their specific integer type as needed.
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

// SerializedScanlinesAdaptor defines the common interface for scanline adaptors.
// Both SerializedScanlinesAdaptorAA and SerializedScanlinesAdaptorBin implement this.
// This eliminates the need for interface{} in text rendering.
//
// The interface provides read-only access to serialized scanline data which can
// be used for rendering glyphs. Different implementations may store anti-aliased
// or binary scanline data, but both expose the same interface for accessing the
// raw data and bounds.
type SerializedScanlinesAdaptor interface {
	Bounds() basics.Rect[int]
	Data() []byte
}
