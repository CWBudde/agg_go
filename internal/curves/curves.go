package curves

import (
	"math"

	"agg_go/internal/array"
	"agg_go/internal/basics"
)

// Constants from AGG curves implementation
const (
	CurveDistanceEpsilon            = 1e-30
	CurveCollinearityEpsilon        = 1e-30
	CurveAngleToleranceEpsilon      = 0.01
	CurveRecursionLimit        uint = 32
)

// CurveApproximationMethod defines the curve approximation algorithm
type CurveApproximationMethod uint

const (
	CurveInc CurveApproximationMethod = iota // Incremental approximation
	CurveDiv                                 // Recursive subdivision
)

// Curve3Inc implements incremental quadratic Bezier curve approximation
type Curve3Inc struct {
	numSteps int
	step     int
	scale    float64
	startX   float64
	startY   float64
	endX     float64
	endY     float64
	fx       float64
	fy       float64
	dfx      float64
	dfy      float64
	ddfx     float64
	ddfy     float64
	savedFx  float64
	savedFy  float64
	savedDfx float64
	savedDfy float64
}

// NewCurve3Inc creates a new incremental quadratic curve
func NewCurve3Inc() *Curve3Inc {
	return &Curve3Inc{
		numSteps: 0,
		step:     0,
		scale:    1.0,
	}
}

// NewCurve3IncWithPoints creates a new incremental quadratic curve with initial points
func NewCurve3IncWithPoints(x1, y1, x2, y2, x3, y3 float64) *Curve3Inc {
	curve := &Curve3Inc{
		numSteps: 0,
		step:     0,
		scale:    1.0,
	}
	curve.Init(x1, y1, x2, y2, x3, y3)
	return curve
}

// Reset resets the curve iterator
func (c *Curve3Inc) Reset() {
	c.numSteps = 0
	c.step = -1
}

// Init initializes the curve with control points
func (c *Curve3Inc) Init(x1, y1, x2, y2, x3, y3 float64) {
	c.startX = x1
	c.startY = y1
	c.endX = x3
	c.endY = y3

	dx1 := x2 - x1
	dy1 := y2 - y1
	dx2 := x3 - x2
	dy2 := y3 - y2

	length := math.Sqrt(dx1*dx1+dy1*dy1) + math.Sqrt(dx2*dx2+dy2*dy2)

	c.numSteps = int(basics.URound(length * 0.25 * c.scale))

	if c.numSteps < 4 {
		c.numSteps = 4
	}

	subdivideStep := 1.0 / float64(c.numSteps)
	subdivideStep2 := subdivideStep * subdivideStep

	tmpx := (x1 - x2*2.0 + x3) * subdivideStep2
	tmpy := (y1 - y2*2.0 + y3) * subdivideStep2

	c.savedFx = x1
	c.fx = x1
	c.savedFy = y1
	c.fy = y1

	c.savedDfx = tmpx + (x2-x1)*(2.0*subdivideStep)
	c.dfx = c.savedDfx
	c.savedDfy = tmpy + (y2-y1)*(2.0*subdivideStep)
	c.dfy = c.savedDfy

	c.ddfx = tmpx * 2.0
	c.ddfy = tmpy * 2.0

	c.step = c.numSteps
}

// ApproximationMethod returns the approximation method (always CurveInc)
func (c *Curve3Inc) ApproximationMethod() CurveApproximationMethod {
	return CurveInc
}

// SetApproximationMethod is a no-op for incremental curves
func (c *Curve3Inc) SetApproximationMethod(CurveApproximationMethod) {
	// No-op for incremental curves
}

// ApproximationScale returns the current approximation scale
func (c *Curve3Inc) ApproximationScale() float64 {
	return c.scale
}

// SetApproximationScale sets the approximation scale
func (c *Curve3Inc) SetApproximationScale(s float64) {
	c.scale = s
}

// AngleTolerance returns the angle tolerance (always 0 for incremental)
func (c *Curve3Inc) AngleTolerance() float64 {
	return 0.0
}

// SetAngleTolerance is a no-op for incremental curves
func (c *Curve3Inc) SetAngleTolerance(float64) {
	// No-op for incremental curves
}

// CuspLimit returns the cusp limit (always 0 for incremental)
func (c *Curve3Inc) CuspLimit() float64 {
	return 0.0
}

// SetCuspLimit is a no-op for incremental curves
func (c *Curve3Inc) SetCuspLimit(float64) {
	// No-op for incremental curves
}

// Rewind rewinds the curve iterator
func (c *Curve3Inc) Rewind(pathID uint) {
	if c.numSteps == 0 {
		c.step = -1
		return
	}
	c.step = c.numSteps
	c.fx = c.savedFx
	c.fy = c.savedFy
	c.dfx = c.savedDfx
	c.dfy = c.savedDfy
}

// Vertex returns the next vertex in the curve approximation
func (c *Curve3Inc) Vertex() (x, y float64, cmd basics.PathCommand) {
	if c.step < 0 {
		return 0, 0, basics.PathCmdStop
	}
	if c.step == c.numSteps {
		x = c.startX
		y = c.startY
		c.step--
		return x, y, basics.PathCmdMoveTo
	}
	if c.step == 0 {
		x = c.endX
		y = c.endY
		c.step--
		return x, y, basics.PathCmdLineTo
	}

	c.fx += c.dfx
	c.fy += c.dfy
	c.dfx += c.ddfx
	c.dfy += c.ddfy
	c.step--
	return c.fx, c.fy, basics.PathCmdLineTo
}

// Curve3Div implements recursive subdivision quadratic Bezier curve approximation
type Curve3Div struct {
	approximationScale float64
	angleTolerance     float64
	count              int
	points             *array.PodBVector[basics.Point[float64]]
}

// NewCurve3Div creates a new recursive subdivision quadratic curve
func NewCurve3Div() *Curve3Div {
	return &Curve3Div{
		approximationScale: 1.0,
		angleTolerance:     0.0,
		count:              0,
		points:             array.NewPodBVector[basics.Point[float64]](),
	}
}

// NewCurve3DivWithPoints creates a new recursive subdivision quadratic curve with initial points
func NewCurve3DivWithPoints(x1, y1, x2, y2, x3, y3 float64) *Curve3Div {
	curve := &Curve3Div{
		approximationScale: 1.0,
		angleTolerance:     0.0,
		count:              0,
		points:             array.NewPodBVector[basics.Point[float64]](),
	}
	curve.Init(x1, y1, x2, y2, x3, y3)
	return curve
}

// Reset resets the curve iterator
func (c *Curve3Div) Reset() {
	c.points.RemoveAll()
	c.count = 0
}

// Init initializes the curve with control points
func (c *Curve3Div) Init(x1, y1, x2, y2, x3, y3 float64) {
	c.points.RemoveAll()
	distanceToleranceSquare := 0.5 / c.approximationScale
	distanceToleranceSquare *= distanceToleranceSquare
	c.bezier(x1, y1, x2, y2, x3, y3, distanceToleranceSquare)
	c.count = 0
}

// ApproximationMethod returns the approximation method (always CurveDiv)
func (c *Curve3Div) ApproximationMethod() CurveApproximationMethod {
	return CurveDiv
}

// SetApproximationMethod is a no-op for subdivision curves
func (c *Curve3Div) SetApproximationMethod(CurveApproximationMethod) {
	// No-op for subdivision curves
}

// ApproximationScale returns the current approximation scale
func (c *Curve3Div) ApproximationScale() float64 {
	return c.approximationScale
}

// SetApproximationScale sets the approximation scale
func (c *Curve3Div) SetApproximationScale(s float64) {
	c.approximationScale = s
}

// AngleTolerance returns the angle tolerance
func (c *Curve3Div) AngleTolerance() float64 {
	return c.angleTolerance
}

// SetAngleTolerance sets the angle tolerance
func (c *Curve3Div) SetAngleTolerance(a float64) {
	c.angleTolerance = a
}

// CuspLimit returns the cusp limit (always 0 for quadratic)
func (c *Curve3Div) CuspLimit() float64 {
	return 0.0
}

// SetCuspLimit is a no-op for quadratic curves
func (c *Curve3Div) SetCuspLimit(float64) {
	// No-op for quadratic curves
}

// Rewind rewinds the curve iterator
func (c *Curve3Div) Rewind(pathID uint) {
	c.count = 0
}

// Vertex returns the next vertex in the curve approximation
func (c *Curve3Div) Vertex() (x, y float64, cmd basics.PathCommand) {
	if c.count >= c.points.Size() {
		return 0, 0, basics.PathCmdStop
	}

	p := c.points.At(c.count)
	c.count++

	if c.count == 1 {
		return p.X, p.Y, basics.PathCmdMoveTo
	}
	return p.X, p.Y, basics.PathCmdLineTo
}

// bezier generates the curve points using recursive subdivision
func (c *Curve3Div) bezier(x1, y1, x2, y2, x3, y3, distanceToleranceSquare float64) {
	c.points.Add(basics.Point[float64]{X: x1, Y: y1})
	c.recursiveBezier(x1, y1, x2, y2, x3, y3, 0, distanceToleranceSquare)
	c.points.Add(basics.Point[float64]{X: x3, Y: y3})
}

// recursiveBezier performs recursive subdivision
func (c *Curve3Div) recursiveBezier(x1, y1, x2, y2, x3, y3 float64, level uint, distanceToleranceSquare float64) {
	if level > CurveRecursionLimit {
		return
	}

	// Calculate all the mid-points of the line segments
	x12 := (x1 + x2) / 2
	y12 := (y1 + y2) / 2
	x23 := (x2 + x3) / 2
	y23 := (y2 + y3) / 2
	x123 := (x12 + x23) / 2
	y123 := (y12 + y23) / 2

	dx := x3 - x1
	dy := y3 - y1
	d := math.Abs((x2-x3)*dy - (y2-y3)*dx)

	if d > CurveCollinearityEpsilon {
		// Regular case
		if d*d <= distanceToleranceSquare*(dx*dx+dy*dy) {
			// If the curvature doesn't exceed the distance_tolerance value
			// we tend to finish subdivisions.
			if c.angleTolerance < CurveAngleToleranceEpsilon {
				c.points.Add(basics.Point[float64]{X: x123, Y: y123})
				return
			}

			// Angle & Cusp Condition
			da := math.Abs(math.Atan2(y3-y2, x3-x2) - math.Atan2(y2-y1, x2-x1))
			if da >= basics.Pi {
				da = 2*basics.Pi - da
			}

			if da < c.angleTolerance {
				// Finally we can stop the recursion
				c.points.Add(basics.Point[float64]{X: x123, Y: y123})
				return
			}
		}
	} else {
		// Collinear case
		da := dx*dx + dy*dy
		if da == 0 {
			d = basics.CalcSqDistance(x1, y1, x2, y2)
		} else {
			d = ((x2-x1)*dx + (y2-y1)*dy) / da
			if d > 0 && d < 1 {
				// Simple collinear case, 1---2---3
				// We can leave just two endpoints
				return
			}
			switch {
			case d <= 0:
				d = basics.CalcSqDistance(x2, y2, x1, y1)
			case d >= 1:
				d = basics.CalcSqDistance(x2, y2, x3, y3)
			default:
				d = basics.CalcSqDistance(x2, y2, x1+d*dx, y1+d*dy)
			}
		}
		if d < distanceToleranceSquare {
			c.points.Add(basics.Point[float64]{X: x2, Y: y2})
			return
		}
	}

	// Continue subdivision
	c.recursiveBezier(x1, y1, x12, y12, x123, y123, level+1, distanceToleranceSquare)
	c.recursiveBezier(x123, y123, x23, y23, x3, y3, level+1, distanceToleranceSquare)
}

// Curve4Inc implements incremental cubic Bezier curve approximation
type Curve4Inc struct {
	numSteps  int
	step      int
	scale     float64
	startX    float64
	startY    float64
	endX      float64
	endY      float64
	fx        float64
	fy        float64
	dfx       float64
	dfy       float64
	ddfx      float64
	ddfy      float64
	dddfx     float64
	dddfy     float64
	savedFx   float64
	savedFy   float64
	savedDfx  float64
	savedDfy  float64
	savedDdfx float64
	savedDdfy float64
}

// NewCurve4Inc creates a new incremental cubic curve
func NewCurve4Inc() *Curve4Inc {
	return &Curve4Inc{
		numSteps: 0,
		step:     0,
		scale:    1.0,
	}
}

// NewCurve4IncWithPoints creates a new incremental cubic curve with initial points
func NewCurve4IncWithPoints(x1, y1, x2, y2, x3, y3, x4, y4 float64) *Curve4Inc {
	curve := &Curve4Inc{
		numSteps: 0,
		step:     0,
		scale:    1.0,
	}
	curve.Init(x1, y1, x2, y2, x3, y3, x4, y4)
	return curve
}

// Reset resets the curve iterator
func (c *Curve4Inc) Reset() {
	c.numSteps = 0
	c.step = -1
}

// Init initializes the curve with control points
func (c *Curve4Inc) Init(x1, y1, x2, y2, x3, y3, x4, y4 float64) {
	c.startX = x1
	c.startY = y1
	c.endX = x4
	c.endY = y4

	dx1 := x2 - x1
	dy1 := y2 - y1
	dx2 := x3 - x2
	dy2 := y3 - y2
	dx3 := x4 - x3
	dy3 := y4 - y3

	length := (math.Sqrt(dx1*dx1+dy1*dy1) +
		math.Sqrt(dx2*dx2+dy2*dy2) +
		math.Sqrt(dx3*dx3+dy3*dy3)) * 0.25 * c.scale

	c.numSteps = int(basics.URound(length))

	if c.numSteps < 4 {
		c.numSteps = 4
	}

	subdivideStep := 1.0 / float64(c.numSteps)
	subdivideStep2 := subdivideStep * subdivideStep
	subdivideStep3 := subdivideStep * subdivideStep * subdivideStep

	pre1 := 3.0 * subdivideStep
	pre2 := 3.0 * subdivideStep2
	pre4 := 6.0 * subdivideStep2
	pre5 := 6.0 * subdivideStep3

	tmp1x := x1 - x2*2.0 + x3
	tmp1y := y1 - y2*2.0 + y3

	tmp2x := (x2-x3)*3.0 - x1 + x4
	tmp2y := (y2-y3)*3.0 - y1 + y4

	c.savedFx = x1
	c.fx = x1
	c.savedFy = y1
	c.fy = y1

	c.savedDfx = (x2-x1)*pre1 + tmp1x*pre2 + tmp2x*subdivideStep3
	c.dfx = c.savedDfx
	c.savedDfy = (y2-y1)*pre1 + tmp1y*pre2 + tmp2y*subdivideStep3
	c.dfy = c.savedDfy

	c.savedDdfx = tmp1x*pre4 + tmp2x*pre5
	c.ddfx = c.savedDdfx
	c.savedDdfy = tmp1y*pre4 + tmp2y*pre5
	c.ddfy = c.savedDdfy

	c.dddfx = tmp2x * pre5
	c.dddfy = tmp2y * pre5

	c.step = c.numSteps
}

// ApproximationMethod returns the approximation method (always CurveInc)
func (c *Curve4Inc) ApproximationMethod() CurveApproximationMethod {
	return CurveInc
}

// SetApproximationMethod is a no-op for incremental curves
func (c *Curve4Inc) SetApproximationMethod(CurveApproximationMethod) {
	// No-op for incremental curves
}

// ApproximationScale returns the current approximation scale
func (c *Curve4Inc) ApproximationScale() float64 {
	return c.scale
}

// SetApproximationScale sets the approximation scale
func (c *Curve4Inc) SetApproximationScale(s float64) {
	c.scale = s
}

// AngleTolerance returns the angle tolerance (always 0 for incremental)
func (c *Curve4Inc) AngleTolerance() float64 {
	return 0.0
}

// SetAngleTolerance is a no-op for incremental curves
func (c *Curve4Inc) SetAngleTolerance(float64) {
	// No-op for incremental curves
}

// CuspLimit returns the cusp limit (always 0 for incremental)
func (c *Curve4Inc) CuspLimit() float64 {
	return 0.0
}

// SetCuspLimit is a no-op for incremental curves
func (c *Curve4Inc) SetCuspLimit(float64) {
	// No-op for incremental curves
}

// Rewind rewinds the curve iterator
func (c *Curve4Inc) Rewind(pathID uint) {
	if c.numSteps == 0 {
		c.step = -1
		return
	}
	c.step = c.numSteps
	c.fx = c.savedFx
	c.fy = c.savedFy
	c.dfx = c.savedDfx
	c.dfy = c.savedDfy
	c.ddfx = c.savedDdfx
	c.ddfy = c.savedDdfy
}

// Vertex returns the next vertex in the curve approximation
func (c *Curve4Inc) Vertex() (x, y float64, cmd basics.PathCommand) {
	if c.step < 0 {
		return 0, 0, basics.PathCmdStop
	}
	if c.step == c.numSteps {
		x = c.startX
		y = c.startY
		c.step--
		return x, y, basics.PathCmdMoveTo
	}

	if c.step == 0 {
		x = c.endX
		y = c.endY
		c.step--
		return x, y, basics.PathCmdLineTo
	}

	c.fx += c.dfx
	c.fy += c.dfy
	c.dfx += c.ddfx
	c.dfy += c.ddfy
	c.ddfx += c.dddfx
	c.ddfy += c.dddfy

	c.step--
	return c.fx, c.fy, basics.PathCmdLineTo
}

// Curve4Div implements recursive subdivision cubic Bezier curve approximation
type Curve4Div struct {
	approximationScale float64
	angleTolerance     float64
	cuspLimit          float64
	count              int
	points             *array.PodBVector[basics.Point[float64]]
}

// NewCurve4Div creates a new recursive subdivision cubic curve
func NewCurve4Div() *Curve4Div {
	return &Curve4Div{
		approximationScale: 1.0,
		angleTolerance:     0.0,
		cuspLimit:          0.0,
		count:              0,
		points:             array.NewPodBVector[basics.Point[float64]](),
	}
}

// NewCurve4DivWithPoints creates a new recursive subdivision cubic curve with initial points
func NewCurve4DivWithPoints(x1, y1, x2, y2, x3, y3, x4, y4 float64) *Curve4Div {
	curve := &Curve4Div{
		approximationScale: 1.0,
		angleTolerance:     0.0,
		cuspLimit:          0.0,
		count:              0,
		points:             array.NewPodBVector[basics.Point[float64]](),
	}
	curve.Init(x1, y1, x2, y2, x3, y3, x4, y4)
	return curve
}

// Reset resets the curve iterator
func (c *Curve4Div) Reset() {
	c.points.RemoveAll()
	c.count = 0
}

// Init initializes the curve with control points
func (c *Curve4Div) Init(x1, y1, x2, y2, x3, y3, x4, y4 float64) {
	c.points.RemoveAll()
	distanceToleranceSquare := 0.5 / c.approximationScale
	distanceToleranceSquare *= distanceToleranceSquare
	c.bezier(x1, y1, x2, y2, x3, y3, x4, y4, distanceToleranceSquare)
	c.count = 0
}

// ApproximationMethod returns the approximation method (always CurveDiv)
func (c *Curve4Div) ApproximationMethod() CurveApproximationMethod {
	return CurveDiv
}

// SetApproximationMethod is a no-op for subdivision curves
func (c *Curve4Div) SetApproximationMethod(CurveApproximationMethod) {
	// No-op for subdivision curves
}

// ApproximationScale returns the current approximation scale
func (c *Curve4Div) ApproximationScale() float64 {
	return c.approximationScale
}

// SetApproximationScale sets the approximation scale
func (c *Curve4Div) SetApproximationScale(s float64) {
	c.approximationScale = s
}

// AngleTolerance returns the angle tolerance
func (c *Curve4Div) AngleTolerance() float64 {
	return c.angleTolerance
}

// SetAngleTolerance sets the angle tolerance
func (c *Curve4Div) SetAngleTolerance(a float64) {
	c.angleTolerance = a
}

// CuspLimit returns the cusp limit
func (c *Curve4Div) CuspLimit() float64 {
	if c.cuspLimit == 0.0 {
		return 0.0
	}
	return basics.Pi - c.cuspLimit
}

// SetCuspLimit sets the cusp limit
func (c *Curve4Div) SetCuspLimit(v float64) {
	if v == 0.0 {
		c.cuspLimit = 0.0
	} else {
		c.cuspLimit = basics.Pi - v
	}
}

// Rewind rewinds the curve iterator
func (c *Curve4Div) Rewind(pathID uint) {
	c.count = 0
}

// Vertex returns the next vertex in the curve approximation
func (c *Curve4Div) Vertex() (x, y float64, cmd basics.PathCommand) {
	if c.count >= c.points.Size() {
		return 0, 0, basics.PathCmdStop
	}

	p := c.points.At(c.count)
	c.count++

	if c.count == 1 {
		return p.X, p.Y, basics.PathCmdMoveTo
	}
	return p.X, p.Y, basics.PathCmdLineTo
}

// bezier generates the curve points using recursive subdivision
func (c *Curve4Div) bezier(x1, y1, x2, y2, x3, y3, x4, y4, distanceToleranceSquare float64) {
	c.points.Add(basics.Point[float64]{X: x1, Y: y1})
	c.recursiveBezier(x1, y1, x2, y2, x3, y3, x4, y4, 0, distanceToleranceSquare)
	c.points.Add(basics.Point[float64]{X: x4, Y: y4})
}

// recursiveBezier performs recursive subdivision for cubic curves
func (c *Curve4Div) recursiveBezier(x1, y1, x2, y2, x3, y3, x4, y4 float64, level uint, distanceToleranceSquare float64) {
	if level > CurveRecursionLimit {
		return
	}

	// Calculate all the mid-points of the line segments
	x12 := (x1 + x2) / 2
	y12 := (y1 + y2) / 2
	x23 := (x2 + x3) / 2
	y23 := (y2 + y3) / 2
	x34 := (x3 + x4) / 2
	y34 := (y3 + y4) / 2
	x123 := (x12 + x23) / 2
	y123 := (y12 + y23) / 2
	x234 := (x23 + x34) / 2
	y234 := (y23 + y34) / 2
	x1234 := (x123 + x234) / 2
	y1234 := (y123 + y234) / 2

	// Try to approximate the full cubic curve by a single straight line
	dx := x4 - x1
	dy := y4 - y1

	d2 := math.Abs((x2-x4)*dy - (y2-y4)*dx)
	d3 := math.Abs((x3-x4)*dy - (y3-y4)*dx)

	var da1, da2, k float64

	switch (func() int {
		result := 0
		if d2 > CurveCollinearityEpsilon {
			result += 1
		}
		if d3 > CurveCollinearityEpsilon {
			result += 2
		}
		return result
	})() {
	case 0:
		// All collinear OR p1==p4
		k = dx*dx + dy*dy
		if k == 0 {
			d2 = basics.CalcSqDistance(x1, y1, x2, y2)
			d3 = basics.CalcSqDistance(x4, y4, x3, y3)
		} else {
			k = 1 / k
			da1 = x2 - x1
			da2 = y2 - y1
			d2 = k * (da1*dx + da2*dy)
			da1 = x3 - x1
			da2 = y3 - y1
			d3 = k * (da1*dx + da2*dy)
			if d2 > 0 && d2 < 1 && d3 > 0 && d3 < 1 {
				// Simple collinear case, 1---2---3---4
				// We can leave just two endpoints
				return
			}
			switch {
			case d2 <= 0:
				d2 = basics.CalcSqDistance(x2, y2, x1, y1)
			case d2 >= 1:
				d2 = basics.CalcSqDistance(x2, y2, x4, y4)
			default:
				d2 = basics.CalcSqDistance(x2, y2, x1+d2*dx, y1+d2*dy)
			}

			switch {
			case d3 <= 0:
				d3 = basics.CalcSqDistance(x3, y3, x1, y1)
			case d3 >= 1:
				d3 = basics.CalcSqDistance(x3, y3, x4, y4)
			default:
				d3 = basics.CalcSqDistance(x3, y3, x1+d3*dx, y1+d3*dy)
			}
		}
		if d2 > d3 {
			if d2 < distanceToleranceSquare {
				c.points.Add(basics.Point[float64]{X: x2, Y: y2})
				return
			}
		} else {
			if d3 < distanceToleranceSquare {
				c.points.Add(basics.Point[float64]{X: x3, Y: y3})
				return
			}
		}

	case 1:
		// p1,p2,p4 are collinear, p3 is significant
		if d3*d3 <= distanceToleranceSquare*(dx*dx+dy*dy) {
			if c.angleTolerance < CurveAngleToleranceEpsilon {
				c.points.Add(basics.Point[float64]{X: x23, Y: y23})
				return
			}

			// Angle Condition
			da1 = math.Abs(math.Atan2(y4-y3, x4-x3) - math.Atan2(y3-y2, x3-x2))
			if da1 >= basics.Pi {
				da1 = 2*basics.Pi - da1
			}

			if da1 < c.angleTolerance {
				c.points.Add(basics.Point[float64]{X: x2, Y: y2})
				c.points.Add(basics.Point[float64]{X: x3, Y: y3})
				return
			}

			if c.cuspLimit != 0.0 {
				if da1 > c.cuspLimit {
					c.points.Add(basics.Point[float64]{X: x3, Y: y3})
					return
				}
			}
		}

	case 2:
		// p1,p3,p4 are collinear, p2 is significant
		if d2*d2 <= distanceToleranceSquare*(dx*dx+dy*dy) {
			if c.angleTolerance < CurveAngleToleranceEpsilon {
				c.points.Add(basics.Point[float64]{X: x23, Y: y23})
				return
			}

			// Angle Condition
			da1 = math.Abs(math.Atan2(y3-y2, x3-x2) - math.Atan2(y2-y1, x2-x1))
			if da1 >= basics.Pi {
				da1 = 2*basics.Pi - da1
			}

			if da1 < c.angleTolerance {
				c.points.Add(basics.Point[float64]{X: x2, Y: y2})
				c.points.Add(basics.Point[float64]{X: x3, Y: y3})
				return
			}

			if c.cuspLimit != 0.0 {
				if da1 > c.cuspLimit {
					c.points.Add(basics.Point[float64]{X: x2, Y: y2})
					return
				}
			}
		}

	case 3:
		// Regular case
		if (d2+d3)*(d2+d3) <= distanceToleranceSquare*(dx*dx+dy*dy) {
			// If the curvature doesn't exceed the distance_tolerance value
			// we tend to finish subdivisions.
			if c.angleTolerance < CurveAngleToleranceEpsilon {
				c.points.Add(basics.Point[float64]{X: x23, Y: y23})
				return
			}

			// Angle & Cusp Condition
			k = math.Atan2(y3-y2, x3-x2)
			da1 = math.Abs(k - math.Atan2(y2-y1, x2-x1))
			da2 = math.Abs(math.Atan2(y4-y3, x4-x3) - k)
			if da1 >= basics.Pi {
				da1 = 2*basics.Pi - da1
			}
			if da2 >= basics.Pi {
				da2 = 2*basics.Pi - da2
			}

			if da1+da2 < c.angleTolerance {
				// Finally we can stop the recursion
				c.points.Add(basics.Point[float64]{X: x23, Y: y23})
				return
			}

			if c.cuspLimit != 0.0 {
				if da1 > c.cuspLimit {
					c.points.Add(basics.Point[float64]{X: x2, Y: y2})
					return
				}

				if da2 > c.cuspLimit {
					c.points.Add(basics.Point[float64]{X: x3, Y: y3})
					return
				}
			}
		}
	}

	// Continue subdivision
	c.recursiveBezier(x1, y1, x12, y12, x123, y123, x1234, y1234, level+1, distanceToleranceSquare)
	c.recursiveBezier(x1234, y1234, x234, y234, x34, y34, x4, y4, level+1, distanceToleranceSquare)
}

// Curve4Points represents four control points for cubic Bezier curves
type Curve4Points struct {
	cp [8]float64 // x1, y1, x2, y2, x3, y3, x4, y4
}

// NewCurve4Points creates a new set of control points
func NewCurve4Points(x1, y1, x2, y2, x3, y3, x4, y4 float64) Curve4Points {
	return Curve4Points{
		cp: [8]float64{x1, y1, x2, y2, x3, y3, x4, y4},
	}
}

// Init initializes the control points
func (cp *Curve4Points) Init(x1, y1, x2, y2, x3, y3, x4, y4 float64) {
	cp.cp[0] = x1
	cp.cp[1] = y1
	cp.cp[2] = x2
	cp.cp[3] = y2
	cp.cp[4] = x3
	cp.cp[5] = y3
	cp.cp[6] = x4
	cp.cp[7] = y4
}

// At returns the coordinate at the specified index
func (cp *Curve4Points) At(i int) float64 {
	return cp.cp[i]
}

// Set sets the coordinate at the specified index
func (cp *Curve4Points) Set(i int, v float64) {
	cp.cp[i] = v
}

// CatromToBezier converts Catmull-Rom spline control points to Bezier curve control points
func CatromToBezier(x1, y1, x2, y2, x3, y3, x4, y4 float64) Curve4Points {
	// Trans. matrix Catmull-Rom to Bezier
	//
	//  0       1       0       0
	//  -1/6    1       1/6     0
	//  0       1/6     1       -1/6
	//  0       0       1       0
	//
	return NewCurve4Points(
		x2,
		y2,
		(-x1+6*x2+x3)/6,
		(-y1+6*y2+y3)/6,
		(x2+6*x3-x4)/6,
		(y2+6*y3-y4)/6,
		x3,
		y3,
	)
}

// CatromToBezierPoints converts Catmull-Rom spline control points to Bezier curve control points
func CatromToBezierPoints(cp Curve4Points) Curve4Points {
	return CatromToBezier(cp.At(0), cp.At(1), cp.At(2), cp.At(3), cp.At(4), cp.At(5), cp.At(6), cp.At(7))
}

// UBSplineToBezier converts uniform B-spline control points to Bezier curve control points
func UBSplineToBezier(x1, y1, x2, y2, x3, y3, x4, y4 float64) Curve4Points {
	// Trans. matrix Uniform BSpline to Bezier
	//
	//  1/6     4/6     1/6     0
	//  0       4/6     2/6     0
	//  0       2/6     4/6     0
	//  0       1/6     4/6     1/6
	//
	return NewCurve4Points(
		(x1+4*x2+x3)/6,
		(y1+4*y2+y3)/6,
		(4*x2+2*x3)/6,
		(4*y2+2*y3)/6,
		(2*x2+4*x3)/6,
		(2*y2+4*y3)/6,
		(x2+4*x3+x4)/6,
		(y2+4*y3+y4)/6,
	)
}

// UBSplineToBezierPoints converts uniform B-spline control points to Bezier curve control points
func UBSplineToBezierPoints(cp Curve4Points) Curve4Points {
	return UBSplineToBezier(cp.At(0), cp.At(1), cp.At(2), cp.At(3), cp.At(4), cp.At(5), cp.At(6), cp.At(7))
}

// HermiteToBezier converts Hermite spline control points to Bezier curve control points
func HermiteToBezier(x1, y1, x2, y2, x3, y3, x4, y4 float64) Curve4Points {
	// Trans. matrix Hermite to Bezier
	//
	//  1       0       0       0
	//  1       0       1/3     0
	//  0       1       0       -1/3
	//  0       1       0       0
	//
	return NewCurve4Points(
		x1,
		y1,
		(3*x1+x3)/3,
		(3*y1+y3)/3,
		(3*x2-x4)/3,
		(3*y2-y4)/3,
		x2,
		y2,
	)
}

// HermiteToBezierPoints converts Hermite spline control points to Bezier curve control points
func HermiteToBezierPoints(cp Curve4Points) Curve4Points {
	return HermiteToBezier(cp.At(0), cp.At(1), cp.At(2), cp.At(3), cp.At(4), cp.At(5), cp.At(6), cp.At(7))
}

// Curve3 is a unified interface for quadratic Bezier curves
type Curve3 struct {
	approximationMethod CurveApproximationMethod
	curveInc            *Curve3Inc
	curveDiv            *Curve3Div
}

// NewCurve3 creates a new quadratic curve with default div approximation
func NewCurve3() *Curve3 {
	return &Curve3{
		approximationMethod: CurveDiv,
		curveInc:            NewCurve3Inc(),
		curveDiv:            NewCurve3Div(),
	}
}

// NewCurve3WithPoints creates a new quadratic curve with initial points
func NewCurve3WithPoints(x1, y1, x2, y2, x3, y3 float64) *Curve3 {
	curve := NewCurve3()
	curve.Init(x1, y1, x2, y2, x3, y3)
	return curve
}

// Reset resets the curve
func (c *Curve3) Reset() {
	c.curveInc.Reset()
	c.curveDiv.Reset()
}

// Init initializes the curve with control points
func (c *Curve3) Init(x1, y1, x2, y2, x3, y3 float64) {
	if c.approximationMethod == CurveInc {
		c.curveInc.Init(x1, y1, x2, y2, x3, y3)
	} else {
		c.curveDiv.Init(x1, y1, x2, y2, x3, y3)
	}
}

// ApproximationMethod returns the current approximation method
func (c *Curve3) ApproximationMethod() CurveApproximationMethod {
	return c.approximationMethod
}

// SetApproximationMethod sets the approximation method
func (c *Curve3) SetApproximationMethod(v CurveApproximationMethod) {
	c.approximationMethod = v
}

// ApproximationScale returns the approximation scale
func (c *Curve3) ApproximationScale() float64 {
	return c.curveInc.ApproximationScale()
}

// SetApproximationScale sets the approximation scale
func (c *Curve3) SetApproximationScale(s float64) {
	c.curveInc.SetApproximationScale(s)
	c.curveDiv.SetApproximationScale(s)
}

// AngleTolerance returns the angle tolerance
func (c *Curve3) AngleTolerance() float64 {
	return c.curveDiv.AngleTolerance()
}

// SetAngleTolerance sets the angle tolerance
func (c *Curve3) SetAngleTolerance(a float64) {
	c.curveDiv.SetAngleTolerance(a)
}

// CuspLimit returns the cusp limit
func (c *Curve3) CuspLimit() float64 {
	return c.curveDiv.CuspLimit()
}

// SetCuspLimit sets the cusp limit
func (c *Curve3) SetCuspLimit(v float64) {
	c.curveDiv.SetCuspLimit(v)
}

// Rewind rewinds the curve iterator
func (c *Curve3) Rewind(pathID uint) {
	if c.approximationMethod == CurveInc {
		c.curveInc.Rewind(pathID)
	} else {
		c.curveDiv.Rewind(pathID)
	}
}

// Vertex returns the next vertex in the curve approximation
func (c *Curve3) Vertex() (x, y float64, cmd basics.PathCommand) {
	if c.approximationMethod == CurveInc {
		return c.curveInc.Vertex()
	}
	return c.curveDiv.Vertex()
}

// Curve4 is a unified interface for cubic Bezier curves
type Curve4 struct {
	approximationMethod CurveApproximationMethod
	curveInc            *Curve4Inc
	curveDiv            *Curve4Div
}

// NewCurve4 creates a new cubic curve with default div approximation
func NewCurve4() *Curve4 {
	return &Curve4{
		approximationMethod: CurveDiv,
		curveInc:            NewCurve4Inc(),
		curveDiv:            NewCurve4Div(),
	}
}

// NewCurve4WithPoints creates a new cubic curve with initial points
func NewCurve4WithPoints(x1, y1, x2, y2, x3, y3, x4, y4 float64) *Curve4 {
	curve := NewCurve4()
	curve.Init(x1, y1, x2, y2, x3, y3, x4, y4)
	return curve
}

// NewCurve4WithControlPoints creates a new cubic curve with Curve4Points
func NewCurve4WithControlPoints(cp Curve4Points) *Curve4 {
	curve := NewCurve4()
	curve.InitWithControlPoints(cp)
	return curve
}

// Reset resets the curve
func (c *Curve4) Reset() {
	c.curveInc.Reset()
	c.curveDiv.Reset()
}

// Init initializes the curve with control points
func (c *Curve4) Init(x1, y1, x2, y2, x3, y3, x4, y4 float64) {
	if c.approximationMethod == CurveInc {
		c.curveInc.Init(x1, y1, x2, y2, x3, y3, x4, y4)
	} else {
		c.curveDiv.Init(x1, y1, x2, y2, x3, y3, x4, y4)
	}
}

// InitWithControlPoints initializes the curve with Curve4Points
func (c *Curve4) InitWithControlPoints(cp Curve4Points) {
	c.Init(cp.At(0), cp.At(1), cp.At(2), cp.At(3), cp.At(4), cp.At(5), cp.At(6), cp.At(7))
}

// ApproximationMethod returns the current approximation method
func (c *Curve4) ApproximationMethod() CurveApproximationMethod {
	return c.approximationMethod
}

// SetApproximationMethod sets the approximation method
func (c *Curve4) SetApproximationMethod(v CurveApproximationMethod) {
	c.approximationMethod = v
}

// ApproximationScale returns the approximation scale
func (c *Curve4) ApproximationScale() float64 {
	return c.curveInc.ApproximationScale()
}

// SetApproximationScale sets the approximation scale
func (c *Curve4) SetApproximationScale(s float64) {
	c.curveInc.SetApproximationScale(s)
	c.curveDiv.SetApproximationScale(s)
}

// AngleTolerance returns the angle tolerance
func (c *Curve4) AngleTolerance() float64 {
	return c.curveDiv.AngleTolerance()
}

// SetAngleTolerance sets the angle tolerance
func (c *Curve4) SetAngleTolerance(a float64) {
	c.curveDiv.SetAngleTolerance(a)
}

// CuspLimit returns the cusp limit
func (c *Curve4) CuspLimit() float64 {
	return c.curveDiv.CuspLimit()
}

// SetCuspLimit sets the cusp limit
func (c *Curve4) SetCuspLimit(v float64) {
	c.curveDiv.SetCuspLimit(v)
}

// Rewind rewinds the curve iterator
func (c *Curve4) Rewind(pathID uint) {
	if c.approximationMethod == CurveInc {
		c.curveInc.Rewind(pathID)
	} else {
		c.curveDiv.Rewind(pathID)
	}
}

// Vertex returns the next vertex in the curve approximation
func (c *Curve4) Vertex() (x, y float64, cmd basics.PathCommand) {
	if c.approximationMethod == CurveInc {
		return c.curveInc.Vertex()
	}
	return c.curveDiv.Vertex()
}
