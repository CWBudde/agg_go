package basics

import (
	"math"
)

// LineCap represents line cap styles
type LineCap int

const (
	ButtCap LineCap = iota
	SquareCap
	RoundCap
)

// LineJoin represents line join styles
type LineJoin int

const (
	MiterJoin LineJoin = iota
	MiterJoinRevert
	RoundJoin
	BevelJoin
	MiterJoinRound
)

// InnerJoin represents inner join styles
type InnerJoin int

const (
	InnerBevel InnerJoin = iota
	InnerMiter
	InnerJag
	InnerRound
)

// VertexConsumer interface for consuming stroke vertices
type VertexConsumer interface {
	Add(x, y float64)
	RemoveAll()
}

// MathStroke performs stroke geometry calculations
type MathStroke struct {
	width           float64
	widthAbs        float64
	widthEps        float64
	widthSign       int
	miterLimit      float64
	innerMiterLimit float64
	approxScale     float64
	lineCap         LineCap
	lineJoin        LineJoin
	innerJoin       InnerJoin
}

// NewMathStroke creates a new stroke calculator
func NewMathStroke() *MathStroke {
	ms := &MathStroke{
		width:           0.5,
		widthAbs:        0.5,
		widthEps:        0.5 / 1024.0,
		widthSign:       1,
		miterLimit:      4.0,
		innerMiterLimit: 1.01,
		approxScale:     1.0,
		lineCap:         ButtCap,
		lineJoin:        MiterJoin,
		innerJoin:       InnerMiter,
	}
	return ms
}

// SetLineCap sets the line cap style
func (ms *MathStroke) SetLineCap(lc LineCap) {
	ms.lineCap = lc
}

// LineCap returns the current line cap style
func (ms *MathStroke) LineCap() LineCap {
	return ms.lineCap
}

// SetLineJoin sets the line join style
func (ms *MathStroke) SetLineJoin(lj LineJoin) {
	ms.lineJoin = lj
}

// LineJoin returns the current line join style
func (ms *MathStroke) LineJoin() LineJoin {
	return ms.lineJoin
}

// SetInnerJoin sets the inner join style
func (ms *MathStroke) SetInnerJoin(ij InnerJoin) {
	ms.innerJoin = ij
}

// InnerJoin returns the current inner join style
func (ms *MathStroke) InnerJoin() InnerJoin {
	return ms.innerJoin
}

// SetWidth sets the stroke width
func (ms *MathStroke) SetWidth(w float64) {
	ms.width = w * 0.5
	if ms.width < 0 {
		ms.widthAbs = -ms.width
		ms.widthSign = -1
	} else {
		ms.widthAbs = ms.width
		ms.widthSign = 1
	}
	ms.widthEps = ms.width / 1024.0
}

// Width returns the stroke width (full width, not radius)
func (ms *MathStroke) Width() float64 {
	return ms.width * 2.0
}

// SetMiterLimit sets the miter limit
func (ms *MathStroke) SetMiterLimit(ml float64) {
	ms.miterLimit = ml
}

// MiterLimit returns the current miter limit
func (ms *MathStroke) MiterLimit() float64 {
	return ms.miterLimit
}

// SetMiterLimitTheta sets the miter limit from an angle
func (ms *MathStroke) SetMiterLimitTheta(t float64) {
	ms.miterLimit = 1.0 / math.Sin(t*0.5)
}

// SetInnerMiterLimit sets the inner miter limit
func (ms *MathStroke) SetInnerMiterLimit(ml float64) {
	ms.innerMiterLimit = ml
}

// InnerMiterLimit returns the current inner miter limit
func (ms *MathStroke) InnerMiterLimit() float64 {
	return ms.innerMiterLimit
}

// SetApproximationScale sets the approximation scale
func (ms *MathStroke) SetApproximationScale(as float64) {
	ms.approxScale = as
}

// ApproximationScale returns the current approximation scale
func (ms *MathStroke) ApproximationScale() float64 {
	return ms.approxScale
}

// addVertex is a helper to add vertices to consumer
func (ms *MathStroke) addVertex(vc VertexConsumer, x, y float64) {
	vc.Add(x, y)
}

// calcArc calculates arc vertices for round joins/caps
func (ms *MathStroke) calcArc(vc VertexConsumer, x, y, dx1, dy1, dx2, dy2 float64) {
	a1 := math.Atan2(dy1*float64(ms.widthSign), dx1*float64(ms.widthSign))
	a2 := math.Atan2(dy2*float64(ms.widthSign), dx2*float64(ms.widthSign))

	da := math.Acos(ms.widthAbs/(ms.widthAbs+0.125/ms.approxScale)) * 2

	ms.addVertex(vc, x+dx1, y+dy1)

	if ms.widthSign > 0 {
		if a1 > a2 {
			a2 += 2 * Pi
		}
		n := int((a2 - a1) / da)
		da = (a2 - a1) / float64(n+1)
		a1 += da
		for i := 0; i < n; i++ {
			ms.addVertex(vc, x+math.Cos(a1)*ms.width, y+math.Sin(a1)*ms.width)
			a1 += da
		}
	} else {
		if a1 < a2 {
			a2 -= 2 * Pi
		}
		n := int((a1 - a2) / da)
		da = (a1 - a2) / float64(n+1)
		a1 -= da
		for i := 0; i < n; i++ {
			ms.addVertex(vc, x+math.Cos(a1)*ms.width, y+math.Sin(a1)*ms.width)
			a1 -= da
		}
	}
	ms.addVertex(vc, x+dx2, y+dy2)
}

// calcMiter calculates miter join with fallback to other join types
func (ms *MathStroke) calcMiter(vc VertexConsumer, v0, v1, v2 VertexDist,
	dx1, dy1, dx2, dy2 float64, lj LineJoin, mlimit, dbevel float64,
) {
	xi := v1.X
	yi := v1.Y
	di := 1.0
	lim := ms.widthAbs * mlimit
	miterLimitExceeded := true // Assume the worst
	intersectionFailed := true // Assume the worst

	// Try to calculate intersection
	var intersectionOK bool
	xi, yi, intersectionOK = CalcIntersection(v0.X+dx1, v0.Y-dy1,
		v1.X+dx1, v1.Y-dy1,
		v1.X+dx2, v1.Y-dy2,
		v2.X+dx2, v2.Y-dy2)

	if intersectionOK {
		// Calculation of the intersection succeeded
		di = CalcDistance(v1.X, v1.Y, xi, yi)
		if di <= lim {
			// Inside the miter limit
			ms.addVertex(vc, xi, yi)
			miterLimitExceeded = false
		}
		intersectionFailed = false
	} else {
		// Intersection failed, check if segments are collinear
		x2 := v1.X + dx1
		y2 := v1.Y - dy1
		if (CrossProduct(v0.X, v0.Y, v1.X, v1.Y, x2, y2) < 0.0) ==
			(CrossProduct(v1.X, v1.Y, v2.X, v2.Y, x2, y2) < 0.0) {
			// Segments continue in straight line
			ms.addVertex(vc, v1.X+dx1, v1.Y-dy1)
			miterLimitExceeded = false
		}
	}

	if miterLimitExceeded {
		// Handle miter limit exceeded
		switch lj {
		case MiterJoinRevert:
			// Simple bevel join for compatibility
			ms.addVertex(vc, v1.X+dx1, v1.Y-dy1)
			ms.addVertex(vc, v1.X+dx2, v1.Y-dy2)

		case MiterJoinRound:
			ms.calcArc(vc, v1.X, v1.Y, dx1, -dy1, dx2, -dy2)

		default:
			// Calculate beveled miter
			if intersectionFailed {
				mlimit *= float64(ms.widthSign)
				ms.addVertex(vc, v1.X+dx1+dy1*mlimit, v1.Y-dy1+dx1*mlimit)
				ms.addVertex(vc, v1.X+dx2-dy2*mlimit, v1.Y-dy2-dx2*mlimit)
			} else {
				x1 := v1.X + dx1
				y1 := v1.Y - dy1
				x2 := v1.X + dx2
				y2 := v1.Y - dy2
				di = (lim - dbevel) / (di - dbevel)
				ms.addVertex(vc, x1+(xi-x1)*di, y1+(yi-y1)*di)
				ms.addVertex(vc, x2+(xi-x2)*di, y2+(yi-y2)*di)
			}
		}
	}
}

// VertexFilter represents a vertex type that can validate itself against another vertex.
type VertexFilter interface {
	// Validate checks if this vertex should be kept when the given vertex is being added.
	// Returns true if the vertex meets the criteria, false if it should be filtered.
	Validate(val VertexFilter) bool
}

// VertexDist represents a vertex with distance information for stroke calculations
type VertexDist struct {
	X, Y float64 // Vertex coordinates
	Dist float64 // Distance to the next vertex
}

// Validate implements the VertexFilter interface for VertexDist.
func (v VertexDist) Validate(val VertexFilter) bool {
	other, ok := val.(VertexDist)
	if !ok {
		return false
	}

	distance := CalcDistance(v.X, v.Y, other.X, other.Y)
	return distance > VertexDistEpsilon
}

// CalculateDistance calculates and sets the distance to another vertex.
func (v *VertexDist) CalculateDistance(other VertexDist) {
	v.Dist = CalcDistance(v.X, v.Y, other.X, other.Y)
	if v.Dist <= VertexDistEpsilon {
		v.Dist = 1.0 / VertexDistEpsilon
	}
}

// CalcCap calculates line cap vertices
func (ms *MathStroke) CalcCap(vc VertexConsumer, v0, v1 VertexDist, length float64) {
	vc.RemoveAll()

	dx1 := (v1.Y - v0.Y) / length
	dy1 := (v1.X - v0.X) / length
	dx2 := 0.0
	dy2 := 0.0

	dx1 *= ms.width
	dy1 *= ms.width

	if ms.lineCap != RoundCap {
		if ms.lineCap == SquareCap {
			dx2 = dy1 * float64(ms.widthSign)
			dy2 = dx1 * float64(ms.widthSign)
		}
		ms.addVertex(vc, v0.X+dx1-dx2, v0.Y-dy1-dy2)
		ms.addVertex(vc, v0.X-dx1-dx2, v0.Y+dy1-dy2)
	} else {
		// Round cap
		da := math.Acos(ms.widthAbs/(ms.widthAbs+0.125/ms.approxScale)) * 2
		n := int(Pi / da)
		da = Pi / float64(n+1)

		ms.addVertex(vc, v0.X+dx1, v0.Y-dy1)

		if ms.widthSign > 0 {
			a1 := math.Atan2(dy1, -dx1)
			a1 += da
			for i := 0; i < n; i++ {
				ms.addVertex(vc, v0.X+math.Cos(a1)*ms.width, v0.Y+math.Sin(a1)*ms.width)
				a1 += da
			}
		} else {
			a1 := math.Atan2(-dy1, dx1)
			a1 -= da
			for i := 0; i < n; i++ {
				ms.addVertex(vc, v0.X+math.Cos(a1)*ms.width, v0.Y+math.Sin(a1)*ms.width)
				a1 -= da
			}
		}
		ms.addVertex(vc, v0.X-dx1, v0.Y+dy1)
	}
}

// CalcJoin calculates line join vertices
func (ms *MathStroke) CalcJoin(vc VertexConsumer, v0, v1, v2 VertexDist, len1, len2 float64) {
	dx1 := ms.width * (v1.Y - v0.Y) / len1
	dy1 := ms.width * (v1.X - v0.X) / len1
	dx2 := ms.width * (v2.Y - v1.Y) / len2
	dy2 := ms.width * (v2.X - v1.X) / len2

	vc.RemoveAll()

	cp := CrossProduct(v0.X, v0.Y, v1.X, v1.Y, v2.X, v2.Y)
	if cp != 0 && (cp > 0) == (ms.width > 0) {
		// Inner join
		limit := len1
		if len2 < len1 {
			limit = len2
		}
		limit /= ms.widthAbs
		if limit < ms.innerMiterLimit {
			limit = ms.innerMiterLimit
		}

		switch ms.innerJoin {
		case InnerMiter:
			ms.calcMiter(vc, v0, v1, v2, dx1, dy1, dx2, dy2, MiterJoinRevert, limit, 0)

		case InnerJag, InnerRound:
			cp := (dx1-dx2)*(dx1-dx2) + (dy1-dy2)*(dy1-dy2)
			if cp < len1*len1 && cp < len2*len2 {
				ms.calcMiter(vc, v0, v1, v2, dx1, dy1, dx2, dy2, MiterJoinRevert, limit, 0)
			} else {
				if ms.innerJoin == InnerJag {
					ms.addVertex(vc, v1.X+dx1, v1.Y-dy1)
					ms.addVertex(vc, v1.X, v1.Y)
					ms.addVertex(vc, v1.X+dx2, v1.Y-dy2)
				} else {
					ms.addVertex(vc, v1.X+dx1, v1.Y-dy1)
					ms.addVertex(vc, v1.X, v1.Y)
					ms.calcArc(vc, v1.X, v1.Y, dx2, -dy2, dx1, -dy1)
					ms.addVertex(vc, v1.X, v1.Y)
					ms.addVertex(vc, v1.X+dx2, v1.Y-dy2)
				}
			}

		default: // InnerBevel
			ms.addVertex(vc, v1.X+dx1, v1.Y-dy1)
			ms.addVertex(vc, v1.X+dx2, v1.Y-dy2)
		}
	} else {
		// Outer join
		dx := (dx1 + dx2) / 2
		dy := (dy1 + dy2) / 2
		dbevel := math.Sqrt(dx*dx + dy*dy)

		// Optimization for nearly collinear segments
		if ms.lineJoin == RoundJoin || ms.lineJoin == BevelJoin {
			if ms.approxScale*(ms.widthAbs-dbevel) < ms.widthEps {
				var intersectOK bool
				dx, dy, intersectOK = CalcIntersection(v0.X+dx1, v0.Y-dy1,
					v1.X+dx1, v1.Y-dy1,
					v1.X+dx2, v1.Y-dy2,
					v2.X+dx2, v2.Y-dy2)
				if intersectOK {
					ms.addVertex(vc, dx, dy)
				} else {
					ms.addVertex(vc, v1.X+dx1, v1.Y-dy1)
				}
				return
			}
		}

		switch ms.lineJoin {
		case MiterJoin, MiterJoinRevert, MiterJoinRound:
			ms.calcMiter(vc, v0, v1, v2, dx1, dy1, dx2, dy2, ms.lineJoin, ms.miterLimit, dbevel)

		case RoundJoin:
			ms.calcArc(vc, v1.X, v1.Y, dx1, -dy1, dx2, -dy2)

		default: // BevelJoin
			ms.addVertex(vc, v1.X+dx1, v1.Y-dy1)
			ms.addVertex(vc, v1.X+dx2, v1.Y-dy2)
		}
	}
}
