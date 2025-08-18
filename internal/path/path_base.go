package path

import (
	"math"

	"agg_go/internal/basics"
	"agg_go/internal/bezierarc"
)

// PathBase is the main path container that stores vertices with their commands.
// A path consists of contours separated by "move_to" commands.
// This is a direct port of AGG's path_base template class.
type PathBase[VertexContainer VertexStorageInterface] struct {
	vertices VertexContainer
	iterator uint
}

// VertexStorageInterface defines the interface that vertex storage must implement.
type VertexStorageInterface interface {
	RemoveAll()
	FreeAll()
	AddVertex(x, y float64, cmd uint32)
	ModifyVertex(idx uint, x, y float64)
	ModifyVertexAndCommand(idx uint, x, y float64, cmd uint32)
	ModifyCommand(idx uint, cmd uint32)
	SwapVertices(v1, v2 uint)
	LastCommand() uint32
	LastVertex() (x, y float64, cmd uint32)
	PrevVertex() (x, y float64, cmd uint32)
	LastX() float64
	LastY() float64
	TotalVertices() uint
	Vertex(idx uint) (x, y float64, cmd uint32)
	Command(idx uint) uint32
}

// NewPathBase creates a new path base.
func NewPathBase[VertexContainer VertexStorageInterface](vertices VertexContainer) *PathBase[VertexContainer] {
	return &PathBase[VertexContainer]{
		vertices: vertices,
		iterator: 0,
	}
}

// RemoveAll removes all vertices but keeps allocated memory.
func (pb *PathBase[VC]) RemoveAll() {
	pb.vertices.RemoveAll()
	pb.iterator = 0
}

// FreeAll removes all vertices and deallocates memory.
func (pb *PathBase[VC]) FreeAll() {
	pb.vertices.FreeAll()
	pb.iterator = 0
}

// StartNewPath starts a new path and returns the path ID.
func (pb *PathBase[VC]) StartNewPath() uint {
	if !basics.IsStop(basics.PathCommand(pb.vertices.LastCommand())) {
		pb.vertices.AddVertex(0.0, 0.0, uint32(basics.PathCmdStop))
	}
	return pb.vertices.TotalVertices()
}

// MoveTo adds a move_to command.
func (pb *PathBase[VC]) MoveTo(x, y float64) {
	pb.vertices.AddVertex(x, y, uint32(basics.PathCmdMoveTo))
}

// MoveRel adds a relative move_to command.
func (pb *PathBase[VC]) MoveRel(dx, dy float64) {
	pb.relToAbs(&dx, &dy)
	pb.vertices.AddVertex(dx, dy, uint32(basics.PathCmdMoveTo))
}

// LineTo adds a line_to command.
func (pb *PathBase[VC]) LineTo(x, y float64) {
	pb.vertices.AddVertex(x, y, uint32(basics.PathCmdLineTo))
}

// LineRel adds a relative line_to command.
func (pb *PathBase[VC]) LineRel(dx, dy float64) {
	pb.relToAbs(&dx, &dy)
	pb.vertices.AddVertex(dx, dy, uint32(basics.PathCmdLineTo))
}

// HLineTo adds a horizontal line to the specified X coordinate.
func (pb *PathBase[VC]) HLineTo(x float64) {
	pb.vertices.AddVertex(x, pb.LastY(), uint32(basics.PathCmdLineTo))
}

// HLineRel adds a relative horizontal line.
func (pb *PathBase[VC]) HLineRel(dx float64) {
	dy := 0.0
	pb.relToAbs(&dx, &dy)
	pb.vertices.AddVertex(dx, dy, uint32(basics.PathCmdLineTo))
}

// VLineTo adds a vertical line to the specified Y coordinate.
func (pb *PathBase[VC]) VLineTo(y float64) {
	pb.vertices.AddVertex(pb.LastX(), y, uint32(basics.PathCmdLineTo))
}

// VLineRel adds a relative vertical line.
func (pb *PathBase[VC]) VLineRel(dy float64) {
	dx := 0.0
	pb.relToAbs(&dx, &dy)
	pb.vertices.AddVertex(dx, dy, uint32(basics.PathCmdLineTo))
}

// ArcTo adds an elliptical arc from current position to (x, y).
func (pb *PathBase[VC]) ArcTo(rx, ry, angle float64, largeArcFlag, sweepFlag bool, x, y float64) {
	if pb.vertices.TotalVertices() > 0 && basics.IsVertex(basics.PathCommand(pb.vertices.LastCommand())) {
		const epsilon = 1e-30
		x0, y0, _ := pb.vertices.LastVertex()

		rx = math.Abs(rx)
		ry = math.Abs(ry)

		// Ensure radii are valid
		if rx < epsilon || ry < epsilon {
			pb.LineTo(x, y)
			return
		}

		if basics.CalcDistance(x0, y0, x, y) < epsilon {
			// If endpoints are identical, omit the arc segment
			return
		}

		// Create bezier arc
		bezierArc := bezierarc.NewBezierArcSVGWithParams(x0, y0, rx, ry, angle, largeArcFlag, sweepFlag, x, y)
		if bezierArc.RadiiOk() {
			// Convert bezier arc vertices to path
			bezierArc.Rewind(0)
			for {
				var x, y float64
				cmd := bezierArc.Vertex(&x, &y)
				if basics.IsStop(cmd) {
					break
				}
				cmdUint := uint32(cmd)
				if basics.IsMoveTo(cmd) {
					cmdUint = uint32(basics.PathCmdLineTo) // Convert first moveto to lineto since we're joining
				}
				pb.vertices.AddVertex(x, y, cmdUint)
			}
		} else {
			pb.LineTo(x, y)
		}
	} else {
		pb.MoveTo(x, y)
	}
}

// ArcRel adds a relative elliptical arc.
func (pb *PathBase[VC]) ArcRel(rx, ry, angle float64, largeArcFlag, sweepFlag bool, dx, dy float64) {
	pb.relToAbs(&dx, &dy)
	pb.ArcTo(rx, ry, angle, largeArcFlag, sweepFlag, dx, dy)
}

// Curve3 adds a quadratic Bezier curve (2 control points).
func (pb *PathBase[VC]) Curve3(xCtrl, yCtrl, xTo, yTo float64) {
	pb.vertices.AddVertex(xCtrl, yCtrl, uint32(basics.PathCmdCurve3))
	pb.vertices.AddVertex(xTo, yTo, uint32(basics.PathCmdCurve3))
}

// Curve3Rel adds a relative quadratic Bezier curve.
func (pb *PathBase[VC]) Curve3Rel(dxCtrl, dyCtrl, dxTo, dyTo float64) {
	pb.relToAbs(&dxCtrl, &dyCtrl)
	pb.relToAbs(&dxTo, &dyTo)
	pb.vertices.AddVertex(dxCtrl, dyCtrl, uint32(basics.PathCmdCurve3))
	pb.vertices.AddVertex(dxTo, dyTo, uint32(basics.PathCmdCurve3))
}

// Curve3Smooth adds a smooth quadratic Bezier curve (reflects previous control point).
func (pb *PathBase[VC]) Curve3Smooth(xTo, yTo float64) {
	x0, y0, _ := pb.vertices.LastVertex()
	if basics.IsVertex(basics.PathCommand(pb.vertices.LastCommand())) {
		xCtrl, yCtrl, cmd := pb.vertices.PrevVertex()
		if basics.IsCurve(basics.PathCommand(cmd)) {
			xCtrl = x0 + x0 - xCtrl
			yCtrl = y0 + y0 - yCtrl
		} else {
			xCtrl = x0
			yCtrl = y0
		}
		pb.Curve3(xCtrl, yCtrl, xTo, yTo)
	}
}

// Curve3SmoothRel adds a relative smooth quadratic Bezier curve.
func (pb *PathBase[VC]) Curve3SmoothRel(dxTo, dyTo float64) {
	pb.relToAbs(&dxTo, &dyTo)
	pb.Curve3Smooth(dxTo, dyTo)
}

// Curve4 adds a cubic Bezier curve (3 control points).
func (pb *PathBase[VC]) Curve4(xCtrl1, yCtrl1, xCtrl2, yCtrl2, xTo, yTo float64) {
	pb.vertices.AddVertex(xCtrl1, yCtrl1, uint32(basics.PathCmdCurve4))
	pb.vertices.AddVertex(xCtrl2, yCtrl2, uint32(basics.PathCmdCurve4))
	pb.vertices.AddVertex(xTo, yTo, uint32(basics.PathCmdCurve4))
}

// Curve4Rel adds a relative cubic Bezier curve.
func (pb *PathBase[VC]) Curve4Rel(dxCtrl1, dyCtrl1, dxCtrl2, dyCtrl2, dxTo, dyTo float64) {
	pb.relToAbs(&dxCtrl1, &dyCtrl1)
	pb.relToAbs(&dxCtrl2, &dyCtrl2)
	pb.relToAbs(&dxTo, &dyTo)
	pb.vertices.AddVertex(dxCtrl1, dyCtrl1, uint32(basics.PathCmdCurve4))
	pb.vertices.AddVertex(dxCtrl2, dyCtrl2, uint32(basics.PathCmdCurve4))
	pb.vertices.AddVertex(dxTo, dyTo, uint32(basics.PathCmdCurve4))
}

// Curve4Smooth adds a smooth cubic Bezier curve (reflects previous control point).
func (pb *PathBase[VC]) Curve4Smooth(xCtrl2, yCtrl2, xTo, yTo float64) {
	x0, y0, _ := pb.LastVertex()
	if basics.IsVertex(basics.PathCommand(pb.vertices.LastCommand())) {
		xCtrl1, yCtrl1, cmd := pb.PrevVertex()
		if basics.IsCurve(basics.PathCommand(cmd)) {
			xCtrl1 = x0 + x0 - xCtrl1
			yCtrl1 = y0 + y0 - yCtrl1
		} else {
			xCtrl1 = x0
			yCtrl1 = y0
		}
		pb.Curve4(xCtrl1, yCtrl1, xCtrl2, yCtrl2, xTo, yTo)
	}
}

// Curve4SmoothRel adds a relative smooth cubic Bezier curve.
func (pb *PathBase[VC]) Curve4SmoothRel(dxCtrl2, dyCtrl2, dxTo, dyTo float64) {
	pb.relToAbs(&dxCtrl2, &dyCtrl2)
	pb.relToAbs(&dxTo, &dyTo)
	pb.Curve4Smooth(dxCtrl2, dyCtrl2, dxTo, dyTo)
}

// EndPoly adds an end_poly command with optional flags.
func (pb *PathBase[VC]) EndPoly(flags basics.PathFlag) {
	if basics.IsVertex(basics.PathCommand(pb.vertices.LastCommand())) {
		pb.vertices.AddVertex(0.0, 0.0, uint32(basics.PathCmdEndPoly)|uint32(flags))
	}
}

// ClosePolygon closes the current polygon with optional flags.
func (pb *PathBase[VC]) ClosePolygon(flags basics.PathFlag) {
	pb.EndPoly(basics.PathFlagsClose | flags)
}

// Vertices returns the vertex container.
func (pb *PathBase[VC]) Vertices() VC {
	return pb.vertices
}

// TotalVertices returns the total number of vertices.
func (pb *PathBase[VC]) TotalVertices() uint {
	return pb.vertices.TotalVertices()
}

// relToAbs converts relative coordinates to absolute.
func (pb *PathBase[VC]) relToAbs(x, y *float64) {
	if pb.vertices.TotalVertices() > 0 {
		x2, y2, cmd := pb.vertices.LastVertex()
		if basics.IsVertex(basics.PathCommand(cmd)) {
			*x += x2
			*y += y2
		}
	}
}

// LastVertex returns the coordinates and command of the last vertex.
func (pb *PathBase[VC]) LastVertex() (x, y float64, cmd uint32) {
	return pb.vertices.LastVertex()
}

// PrevVertex returns the coordinates and command of the previous vertex.
func (pb *PathBase[VC]) PrevVertex() (x, y float64, cmd uint32) {
	return pb.vertices.PrevVertex()
}

// LastX returns the X coordinate of the last vertex.
func (pb *PathBase[VC]) LastX() float64 {
	return pb.vertices.LastX()
}

// LastY returns the Y coordinate of the last vertex.
func (pb *PathBase[VC]) LastY() float64 {
	return pb.vertices.LastY()
}

// Vertex returns the coordinates and command of the vertex at the given index.
func (pb *PathBase[VC]) Vertex(idx uint) (x, y float64, cmd uint32) {
	return pb.vertices.Vertex(idx)
}

// Command returns the command of the vertex at the given index.
func (pb *PathBase[VC]) Command(idx uint) uint32 {
	return pb.vertices.Command(idx)
}

// ModifyVertex modifies the coordinates of an existing vertex.
func (pb *PathBase[VC]) ModifyVertex(idx uint, x, y float64) {
	pb.vertices.ModifyVertex(idx, x, y)
}

// ModifyVertexAndCommand modifies both coordinates and command of an existing vertex.
func (pb *PathBase[VC]) ModifyVertexAndCommand(idx uint, x, y float64, cmd uint32) {
	pb.vertices.ModifyVertexAndCommand(idx, x, y, cmd)
}

// ModifyCommand modifies the command of an existing vertex.
func (pb *PathBase[VC]) ModifyCommand(idx uint, cmd uint32) {
	pb.vertices.ModifyCommand(idx, cmd)
}

// Rewind implements the VertexSource interface.
func (pb *PathBase[VC]) Rewind(pathID uint) {
	pb.iterator = pathID
}

// NextVertex implements the VertexSource interface.
func (pb *PathBase[VC]) NextVertex() (x, y float64, cmd uint32) {
	if pb.iterator >= pb.vertices.TotalVertices() {
		return 0, 0, uint32(basics.PathCmdStop)
	}
	x, y, cmd = pb.vertices.Vertex(pb.iterator)
	pb.iterator++
	return
}

// ConcatPath concatenates another path to this one.
func (pb *PathBase[VC]) ConcatPath(vs VertexSource, pathID uint) {
	vs.Rewind(pathID)
	for {
		x, y, cmd := vs.NextVertex()
		if basics.IsStop(basics.PathCommand(cmd)) {
			break
		}
		pb.vertices.AddVertex(x, y, cmd)
	}
}

// JoinPath joins another path to this one (continuous drawing).
func (pb *PathBase[VC]) JoinPath(vs VertexSource) {
	vs.Rewind(0)
	x, y, cmd := vs.NextVertex()
	if !basics.IsStop(basics.PathCommand(cmd)) {
		if basics.IsVertex(basics.PathCommand(cmd)) {
			x0, y0, cmd0 := pb.LastVertex()
			if basics.IsVertex(basics.PathCommand(cmd0)) {
				if basics.CalcDistance(x, y, x0, y0) > basics.VertexDistEpsilon {
					if basics.IsMoveTo(basics.PathCommand(cmd)) {
						cmd = uint32(basics.PathCmdLineTo)
					}
					pb.vertices.AddVertex(x, y, cmd)
				}
			} else {
				if basics.IsStop(basics.PathCommand(cmd0)) {
					cmd = uint32(basics.PathCmdMoveTo)
				} else {
					if basics.IsMoveTo(basics.PathCommand(cmd)) {
						cmd = uint32(basics.PathCmdLineTo)
					}
				}
				pb.vertices.AddVertex(x, y, cmd)
			}
		}
		for {
			x, y, cmd := vs.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			if basics.IsMoveTo(basics.PathCommand(cmd)) {
				cmd = uint32(basics.PathCmdLineTo)
			}
			pb.vertices.AddVertex(x, y, cmd)
		}
	}
}

// ConcatPoly concatenates a polygon from coordinate array.
func (pb *PathBase[VC]) ConcatPoly(data []float64, numPoints uint, closed bool) {
	adaptor := NewPolyPlainAdaptorWithData(data, numPoints, closed)
	pb.ConcatPath(adaptor, 0)
}

// JoinPoly joins a polygon from coordinate array.
func (pb *PathBase[VC]) JoinPoly(data []float64, numPoints uint, closed bool) {
	adaptor := NewPolyPlainAdaptorWithData(data, numPoints, closed)
	pb.JoinPath(adaptor)
}

// Translate translates all vertices in a path by (dx, dy).
func (pb *PathBase[VC]) Translate(dx, dy float64, pathID uint) {
	numVer := pb.vertices.TotalVertices()
	for i := pathID; i < numVer; i++ {
		x, y, cmd := pb.vertices.Vertex(i)
		if basics.IsStop(basics.PathCommand(cmd)) {
			break
		}
		if basics.IsVertex(basics.PathCommand(cmd)) {
			pb.vertices.ModifyVertex(i, x+dx, y+dy)
		}
	}
}

// TranslateAllPaths translates all vertices in all paths by (dx, dy).
func (pb *PathBase[VC]) TranslateAllPaths(dx, dy float64) {
	numVer := pb.vertices.TotalVertices()
	for i := uint(0); i < numVer; i++ {
		x, y, cmd := pb.vertices.Vertex(i)
		if basics.IsVertex(basics.PathCommand(cmd)) {
			pb.vertices.ModifyVertex(i, x+dx, y+dy)
		}
	}
}

// FlipX flips all vertices horizontally between x1 and x2.
func (pb *PathBase[VC]) FlipX(x1, x2 float64) {
	numVer := pb.vertices.TotalVertices()
	for i := uint(0); i < numVer; i++ {
		x, y, cmd := pb.vertices.Vertex(i)
		if basics.IsVertex(basics.PathCommand(cmd)) {
			pb.vertices.ModifyVertex(i, x2-x+x1, y)
		}
	}
}

// FlipY flips all vertices vertically between y1 and y2.
func (pb *PathBase[VC]) FlipY(y1, y2 float64) {
	numVer := pb.vertices.TotalVertices()
	for i := uint(0); i < numVer; i++ {
		x, y, cmd := pb.vertices.Vertex(i)
		if basics.IsVertex(basics.PathCommand(cmd)) {
			pb.vertices.ModifyVertex(i, x, y2-y+y1)
		}
	}
}

// PathStorage is the standard path storage type using VertexBlockStorage with float64.
type PathStorage = PathBase[*VertexBlockStorage[float64]]

// NewPathStorage creates a new path storage.
func NewPathStorage() *PathStorage {
	return NewPathBase(NewVertexBlockStorage[float64]())
}
