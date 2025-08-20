package polygon

import (
	"math"

	"agg_go/internal/array"
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/conv"
	"agg_go/internal/ctrl"
	"agg_go/internal/shapes"
)

// PolygonCtrl implements an interactive polygon control with draggable points.
// This corresponds to AGG's polygon_ctrl_impl class.
type PolygonCtrl struct {
	*ctrl.BaseCtrl

	// Polygon data
	polygon        *array.PodArray[float64] // coordinate array [x1, y1, x2, y2, ...]
	numPoints      uint
	node           int     // currently selected node (-1 if none)
	edge           int     // currently selected edge (-1 if none)
	pointRadius    float64 // radius for point hit testing
	status         uint    // rendering state
	dx, dy         float64 // drag offset
	inPolygonCheck bool    // whether to enable point-in-polygon checking

	// Rendering components
	vs      *SimplePolygonVertexSource
	stroke  *conv.ConvStroke
	ellipse *shapes.Ellipse

	// Color for single-color rendering
	lineColor color.RGBA
}

// NewPolygonCtrl creates a new polygon control with the specified number of points.
// numPoints: number of polygon vertices
// pointRadius: radius for point hit testing and rendering (default: 5.0)
func NewPolygonCtrl(numPoints uint, pointRadius float64) *PolygonCtrl {
	if pointRadius <= 0 {
		pointRadius = 5.0
	}

	polygon := array.NewPodArrayWithSize[float64](int(numPoints * 2))

	vs := NewSimplePolygonVertexSource(polygon.Data(), numPoints, false, true)
	stroke := conv.NewConvStroke(vs)
	stroke.SetWidth(1.0)

	ctrl := &PolygonCtrl{
		BaseCtrl:       ctrl.NewBaseCtrl(0, 0, 1, 1, false),
		polygon:        polygon,
		numPoints:      numPoints,
		node:           -1,
		edge:           -1,
		pointRadius:    pointRadius,
		status:         0,
		dx:             0.0,
		dy:             0.0,
		inPolygonCheck: true,
		vs:             vs,
		stroke:         stroke,
		ellipse:        shapes.NewEllipse(),
		lineColor:      color.NewRGBA(0.0, 0.0, 0.0, 1.0),
	}

	return ctrl
}

// Polygon Management Methods

// NumPoints returns the number of points in the polygon.
func (p *PolygonCtrl) NumPoints() uint {
	return p.numPoints
}

// Xn returns the X coordinate of point n.
func (p *PolygonCtrl) Xn(n uint) float64 {
	if n >= p.numPoints {
		return 0.0
	}
	return p.polygon.At(int(n * 2))
}

// Yn returns the Y coordinate of point n.
func (p *PolygonCtrl) Yn(n uint) float64 {
	if n >= p.numPoints {
		return 0.0
	}
	return p.polygon.At(int(n*2 + 1))
}

// SetXn sets the X coordinate of point n.
func (p *PolygonCtrl) SetXn(n uint, x float64) {
	if n < p.numPoints {
		p.polygon.Set(int(n*2), x)
	}
}

// SetYn sets the Y coordinate of point n.
func (p *PolygonCtrl) SetYn(n uint, y float64) {
	if n < p.numPoints {
		p.polygon.Set(int(n*2+1), y)
	}
}

// PolygonData returns the raw polygon coordinate array.
func (p *PolygonCtrl) PolygonData() []float64 {
	return p.polygon.Data()
}

// Configuration Methods

// SetLineWidth sets the width of the polygon stroke.
func (p *PolygonCtrl) SetLineWidth(w float64) {
	p.stroke.SetWidth(w)
}

// LineWidth returns the current stroke width.
func (p *PolygonCtrl) LineWidth() float64 {
	return p.stroke.Width()
}

// SetPointRadius sets the radius for point hit testing.
func (p *PolygonCtrl) SetPointRadius(r float64) {
	p.pointRadius = r
}

// PointRadius returns the current point radius.
func (p *PolygonCtrl) PointRadius() float64 {
	return p.pointRadius
}

// SetInPolygonCheck enables or disables point-in-polygon checking.
func (p *PolygonCtrl) SetInPolygonCheck(f bool) {
	p.inPolygonCheck = f
}

// InPolygonCheck returns whether point-in-polygon checking is enabled.
func (p *PolygonCtrl) InPolygonCheck() bool {
	return p.inPolygonCheck
}

// SetClose sets whether the polygon should be closed.
func (p *PolygonCtrl) SetClose(f bool) {
	p.vs.Close(f)
}

// Close returns whether the polygon is closed.
func (p *PolygonCtrl) Close() bool {
	return p.vs.IsClose()
}

// SetLineColor sets the line color for rendering.
func (p *PolygonCtrl) SetLineColor(c color.RGBA) {
	p.lineColor = c
}

// LineColor returns the current line color.
func (p *PolygonCtrl) LineColor() color.RGBA {
	return p.lineColor
}

// Mouse Interaction Methods

// OnMouseButtonDown handles mouse button press events.
func (p *PolygonCtrl) OnMouseButtonDown(x, y float64) bool {
	p.node = -1
	p.edge = -1
	p.InverseTransformXY(&x, &y)

	// Check if clicking on a point
	for i := uint(0); i < p.numPoints; i++ {
		dx := x - p.Xn(i)
		dy := y - p.Yn(i)
		if math.Sqrt(dx*dx+dy*dy) <= p.pointRadius {
			p.dx = dx
			p.dy = dy
			p.node = int(i)
			return true
		}
	}

	// Check if clicking on an edge
	for i := uint(0); i < p.numPoints; i++ {
		if p.checkEdge(i, x, y) {
			p.dx = x
			p.dy = y
			p.edge = int(i)
			return true
		}
	}

	// Check if clicking inside polygon (if enabled)
	if p.inPolygonCheck && p.pointInPolygon(x, y) {
		p.dx = x
		p.dy = y
		p.node = int(p.numPoints) // special value for whole polygon
		return true
	}

	return false
}

// OnMouseButtonUp handles mouse button release events.
func (p *PolygonCtrl) OnMouseButtonUp(x, y float64) bool {
	result := p.node >= 0 || p.edge >= 0
	p.node = -1
	p.edge = -1
	return result
}

// OnMouseMove handles mouse move events.
func (p *PolygonCtrl) OnMouseMove(x, y float64, buttonPressed bool) bool {
	if !buttonPressed {
		return false
	}

	p.InverseTransformXY(&x, &y)

	if p.node == int(p.numPoints) {
		// Move entire polygon
		dx := x - p.dx
		dy := y - p.dy
		for i := uint(0); i < p.numPoints; i++ {
			p.SetXn(i, p.Xn(i)+dx)
			p.SetYn(i, p.Yn(i)+dy)
		}
		p.dx = x
		p.dy = y
		return true
	}

	if p.edge >= 0 {
		// Move edge (both vertices)
		dx := x - p.dx
		dy := y - p.dy
		i := uint(p.edge)
		j := (i + p.numPoints - 1) % p.numPoints
		p.SetXn(i, p.Xn(i)+dx)
		p.SetYn(i, p.Yn(i)+dy)
		p.SetXn(j, p.Xn(j)+dx)
		p.SetYn(j, p.Yn(j)+dy)
		p.dx = x
		p.dy = y
		return true
	}

	if p.node >= 0 && p.node < int(p.numPoints) {
		// Move single point
		p.SetXn(uint(p.node), x-p.dx)
		p.SetYn(uint(p.node), y-p.dy)
		return true
	}

	return false
}

// OnArrowKeys handles arrow key events for point adjustment.
func (p *PolygonCtrl) OnArrowKeys(left, right, down, up bool) bool {
	if p.node < 0 || p.node >= int(p.numPoints) {
		return false
	}

	dx, dy := 0.0, 0.0
	if left {
		dx = -1.0
	}
	if right {
		dx = 1.0
	}
	if down {
		dy = 1.0
	}
	if up {
		dy = -1.0
	}

	if dx != 0.0 || dy != 0.0 {
		i := uint(p.node)
		p.SetXn(i, p.Xn(i)+dx)
		p.SetYn(i, p.Yn(i)+dy)
		return true
	}

	return false
}

// Vertex Source Interface

// NumPaths returns the number of rendering paths (1 for polygon stroke).
func (p *PolygonCtrl) NumPaths() uint {
	return 1
}

// Rewind resets the vertex generation state.
func (p *PolygonCtrl) Rewind(pathID uint) {
	p.status = 0
	p.stroke.Rewind(0)
}

// Vertex returns the next vertex for rendering.
func (p *PolygonCtrl) Vertex() (x, y float64, cmd basics.PathCommand) {
	r := p.pointRadius

	if p.status == 0 {
		// First render the polygon stroke
		x, y, cmd = p.stroke.Vertex()
		if cmd != basics.PathCmdStop {
			p.TransformXY(&x, &y)
			return x, y, cmd
		}

		// Then start rendering control points
		if p.node >= 0 && p.node == int(p.status) {
			r *= 1.2 // Highlight selected point
		}
		p.ellipse.Init(p.Xn(p.status), p.Yn(p.status), r, r, 32, false)
		p.status++
	}

	// Render current control point
	cmd = p.ellipse.Vertex(&x, &y)
	if cmd != basics.PathCmdStop {
		p.TransformXY(&x, &y)
		return x, y, cmd
	}

	// Move to next control point
	if p.status >= p.numPoints {
		return 0, 0, basics.PathCmdStop
	}

	if p.node >= 0 && p.node == int(p.status) {
		r *= 1.2 // Highlight selected point
	}
	p.ellipse.Init(p.Xn(p.status), p.Yn(p.status), r, r, 32, false)
	p.status++

	cmd = p.ellipse.Vertex(&x, &y)
	if cmd != basics.PathCmdStop {
		p.TransformXY(&x, &y)
	}
	return x, y, cmd
}

// Color returns the color for the specified path.
func (p *PolygonCtrl) Color(pathID uint) interface{} {
	return p.lineColor
}

// Helper Methods

// checkEdge checks if a point is near an edge of the polygon.
func (p *PolygonCtrl) checkEdge(i uint, x, y float64) bool {
	n1 := i
	n2 := (i + p.numPoints - 1) % p.numPoints
	x1 := p.Xn(n1)
	y1 := p.Yn(n1)
	x2 := p.Xn(n2)
	y2 := p.Yn(n2)

	dx := x2 - x1
	dy := y2 - y1

	if math.Sqrt(dx*dx+dy*dy) <= 0.0000001 {
		return false
	}

	x3 := x
	y3 := y
	x4 := x3 - dy
	y4 := y3 + dx

	den := (y4-y3)*(x2-x1) - (x4-x3)*(y2-y1)
	if math.Abs(den) < 0.0000001 {
		return false
	}

	u1 := ((x4-x3)*(y1-y3) - (y4-y3)*(x1-x3)) / den

	xi := x1 + u1*(x2-x1)
	yi := y1 + u1*(y2-y1)

	dx = xi - x
	dy = yi - y

	return u1 > 0.0 && u1 < 1.0 && math.Sqrt(dx*dx+dy*dy) <= p.pointRadius
}

// pointInPolygon determines if a point is inside the polygon using ray casting.
func (p *PolygonCtrl) pointInPolygon(x, y float64) bool {
	if p.numPoints < 3 {
		return false
	}

	inside := false
	j := p.numPoints - 1

	for i := uint(0); i < p.numPoints; i++ {
		xi, yi := p.Xn(i), p.Yn(i)
		xj, yj := p.Xn(j), p.Yn(j)

		if ((yi > y) != (yj > y)) && (x < (xj-xi)*(y-yi)/(yj-yi)+xi) {
			inside = !inside
		}
		j = i
	}

	return inside
}
