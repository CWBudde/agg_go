package conv

import (
	"agg_go/internal/basics"
	"agg_go/internal/curves"
	"agg_go/internal/path"
)

// ConvCurveInteger is a specialized curve converter for integer path storage
// This corresponds to AGG's conv_curve<path_storage_integer<T>> pattern
type ConvCurveInteger[T ~int16 | ~int32 | ~int64] struct {
	storage   *path.PathStorageInteger[T]
	adapter   *path.PathStorageIntegerAdapter[T]
	convCurve *ConvCurve
}

// NewConvCurveInteger creates a new curve converter for integer path storage
func NewConvCurveInteger[T ~int16 | ~int32 | ~int64](storage *path.PathStorageInteger[T]) *ConvCurveInteger[T] {
	adapter := path.NewPathStorageIntegerAdapter(storage)
	convCurve := NewConvCurve(adapter)

	return &ConvCurveInteger[T]{
		storage:   storage,
		adapter:   adapter,
		convCurve: convCurve,
	}
}

// Storage returns the underlying PathStorageInteger
func (cci *ConvCurveInteger[T]) Storage() *path.PathStorageInteger[T] {
	return cci.storage
}

// ConvCurve returns the underlying ConvCurve converter
func (cci *ConvCurveInteger[T]) ConvCurve() *ConvCurve {
	return cci.convCurve
}

// ApproximationMethod returns the current approximation method
func (cci *ConvCurveInteger[T]) ApproximationMethod() curves.CurveApproximationMethod {
	return cci.convCurve.ApproximationMethod()
}

// SetApproximationMethod sets the approximation method for curve conversion
func (cci *ConvCurveInteger[T]) SetApproximationMethod(method curves.CurveApproximationMethod) {
	cci.convCurve.SetApproximationMethod(method)
}

// ApproximationScale returns the current approximation scale
func (cci *ConvCurveInteger[T]) ApproximationScale() float64 {
	return cci.convCurve.ApproximationScale()
}

// SetApproximationScale sets the approximation scale for curve conversion
func (cci *ConvCurveInteger[T]) SetApproximationScale(scale float64) {
	cci.convCurve.SetApproximationScale(scale)
}

// AngleTolerance returns the current angle tolerance
func (cci *ConvCurveInteger[T]) AngleTolerance() float64 {
	return cci.convCurve.AngleTolerance()
}

// SetAngleTolerance sets the angle tolerance for curve conversion
func (cci *ConvCurveInteger[T]) SetAngleTolerance(tolerance float64) {
	cci.convCurve.SetAngleTolerance(tolerance)
}

// CuspLimit returns the current cusp limit
func (cci *ConvCurveInteger[T]) CuspLimit() float64 {
	return cci.convCurve.CuspLimit()
}

// SetCuspLimit sets the cusp limit for curve conversion
func (cci *ConvCurveInteger[T]) SetCuspLimit(limit float64) {
	cci.convCurve.SetCuspLimit(limit)
}

// Rewind rewinds the curve converter to start from the beginning
func (cci *ConvCurveInteger[T]) Rewind(pathID uint) {
	cci.convCurve.Rewind(pathID)
}

// Vertex returns the next vertex, converting curves to line segments
func (cci *ConvCurveInteger[T]) Vertex() (x, y float64, cmd basics.PathCommand) {
	return cci.convCurve.Vertex()
}

// RemoveAll removes all vertices from the underlying path storage
func (cci *ConvCurveInteger[T]) RemoveAll() {
	cci.storage.RemoveAll()
}

// MoveTo adds a move_to command with the given coordinates
func (cci *ConvCurveInteger[T]) MoveTo(x, y T) {
	cci.storage.MoveTo(x, y)
}

// LineTo adds a line_to command with the given coordinates
func (cci *ConvCurveInteger[T]) LineTo(x, y T) {
	cci.storage.LineTo(x, y)
}

// Curve3 adds a quadratic Bézier curve with control point and end point
func (cci *ConvCurveInteger[T]) Curve3(xCtrl, yCtrl, xTo, yTo T) {
	cci.storage.Curve3(xCtrl, yCtrl, xTo, yTo)
}

// Curve4 adds a cubic Bézier curve with two control points and end point
func (cci *ConvCurveInteger[T]) Curve4(xCtrl1, yCtrl1, xCtrl2, yCtrl2, xTo, yTo T) {
	cci.storage.Curve4(xCtrl1, yCtrl1, xCtrl2, yCtrl2, xTo, yTo)
}

// ClosePolygon closes the current polygon
func (cci *ConvCurveInteger[T]) ClosePolygon() {
	cci.storage.ClosePolygon()
}

// Size returns the number of vertices in the path
func (cci *ConvCurveInteger[T]) Size() uint32 {
	return cci.storage.Size()
}

// BoundingRect calculates the bounding rectangle of the path
func (cci *ConvCurveInteger[T]) BoundingRect() basics.Rect[float64] {
	return cci.storage.BoundingRect()
}
