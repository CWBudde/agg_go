// Package shapesdata provides embedded Flash compound-shape data from shapes.txt.
//
// The file uses a simple text format:
//
//	=...  – begins a new shape frame
//	P left right line ax ay – new sub-path with fill/line style IDs and start point
//	C cx cy ax ay           – quadratic Bezier (Curve3) to (ax,ay) via control (cx,cy)
//	L ax ay                 – LineTo (ax,ay)
//	<...                    – EndPath marker (ignored, path ends at next P / !)
//	!...                    – ends the current shape frame
package shapesdata

import (
	_ "embed"
	"math"
	"strconv"
	"strings"
)

//go:embed shapes.txt
var ShapesTxt []byte

// RawVertex holds a single vertex command from shapes.txt.
type RawVertex struct {
	X, Y    float64 // endpoint
	CX, CY  float64 // control point (only valid for Curve3)
	IsCurve bool    // true = Curve3, false = LineTo (first vertex is always MoveTo)
}

// RawPath is one sub-path within a shape, with its style IDs and flattened vertices.
type RawPath struct {
	LeftFill  int // fill style index for the left side (-1 = none)
	RightFill int // fill style index for the right side (-1 = none)
	Line      int // line style index (-1 = none)
	Vertices  []RawVertex
}

// RawShape is one complete shape (frame) parsed from shapes.txt.
type RawShape struct {
	Paths    []RawPath
	MinStyle int
	MaxStyle int
}

// ParseShapes parses all shapes from shapes.txt data.
func ParseShapes(data []byte) []RawShape {
	lines := strings.Split(string(data), "\n")

	var shapes []RawShape
	var cur *RawShape
	var curPath *RawPath

	for _, raw := range lines {
		line := strings.TrimRight(raw, "\r")
		if len(line) == 0 {
			continue
		}
		switch line[0] {
		case '=':
			// Begin new shape
			if cur != nil {
				shapes = append(shapes, *cur)
			}
			cur = &RawShape{MinStyle: math.MaxInt32, MaxStyle: math.MinInt32}
			curPath = nil

		case '!':
			// End of shape
			if cur != nil {
				shapes = append(shapes, *cur)
			}
			cur = nil
			curPath = nil

		case 'P':
			// Path left right line ax ay
			if cur == nil {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) < 6 {
				continue
			}
			lf := parseInt(fields[1])
			rf := parseInt(fields[2])
			ln := parseInt(fields[3])
			ax := parseFloat(fields[4])
			ay := parseFloat(fields[5])

			cur.Paths = append(cur.Paths, RawPath{
				LeftFill:  lf,
				RightFill: rf,
				Line:      ln,
				Vertices:  []RawVertex{{X: ax, Y: ay}},
			})
			curPath = &cur.Paths[len(cur.Paths)-1]

			updateStyleRange(cur, lf, rf)

		case 'C':
			// Curve cx cy ax ay
			if curPath == nil {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) < 5 {
				continue
			}
			curPath.Vertices = append(curPath.Vertices, RawVertex{
				CX:      parseFloat(fields[1]),
				CY:      parseFloat(fields[2]),
				X:       parseFloat(fields[3]),
				Y:       parseFloat(fields[4]),
				IsCurve: true,
			})

		case 'L':
			// Line ax ay
			if curPath == nil {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) < 3 {
				continue
			}
			curPath.Vertices = append(curPath.Vertices, RawVertex{
				X: parseFloat(fields[1]),
				Y: parseFloat(fields[2]),
			})

		case '<':
			// EndPath – no action needed (path already complete)
		}
	}

	if cur != nil {
		shapes = append(shapes, *cur)
	}

	return shapes
}

func updateStyleRange(s *RawShape, lf, rf int) {
	if lf >= 0 {
		if lf < s.MinStyle {
			s.MinStyle = lf
		}
		if lf > s.MaxStyle {
			s.MaxStyle = lf
		}
	}
	if rf >= 0 {
		if rf < s.MinStyle {
			s.MinStyle = rf
		}
		if rf > s.MaxStyle {
			s.MaxStyle = rf
		}
	}
}

func parseInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// BoundingRect computes the conservative bounding box of all paths in the shape.
// For Curve3 it uses the control point as a conservative extension.
func (s *RawShape) BoundingRect() (x1, y1, x2, y2 float64) {
	x1, y1 = math.MaxFloat64, math.MaxFloat64
	x2, y2 = -math.MaxFloat64, -math.MaxFloat64
	for _, p := range s.Paths {
		for _, v := range p.Vertices {
			expand(&x1, &y1, &x2, &y2, v.X, v.Y)
			if v.IsCurve {
				expand(&x1, &y1, &x2, &y2, v.CX, v.CY)
			}
		}
	}
	return
}

func expand(x1, y1, x2, y2 *float64, x, y float64) {
	if x < *x1 {
		*x1 = x
	}
	if y < *y1 {
		*y1 = y
	}
	if x > *x2 {
		*x2 = x
	}
	if y > *y2 {
		*y2 = y
	}
}

// FlatVertex is a flattened (no curves) vertex with screen coordinates.
type FlatVertex struct {
	X, Y float64
	Cmd  uint32 // PathCmdMoveTo or PathCmdLineTo
}

// FlattenPath flattens a RawPath (applies quadratic bezier subdivision and affine) into FlatVertex slices.
// The affine is specified as [a, b, c, d, e, f] where x' = a*x + c*y + e, y' = b*x + d*y + f.
// approxScale controls curve subdivision quality (use the viewport scale for zoom-aware flattening).
func FlattenPath(p *RawPath, sx, sy, tx, ty, approxScale float64) []FlatVertex {
	if len(p.Vertices) == 0 {
		return nil
	}
	result := make([]FlatVertex, 0, len(p.Vertices)*2)

	// First vertex is always MoveTo
	v0 := p.Vertices[0]
	x0, y0 := v0.X*sx+tx, v0.Y*sy+ty
	result = append(result, FlatVertex{X: x0, Y: y0, Cmd: PathCmdMoveTo})

	for i := 1; i < len(p.Vertices); i++ {
		v := p.Vertices[i]
		if v.IsCurve {
			// Quadratic bezier from (x0,y0) via control (cx,cy) to (ax,ay)
			cx, cy := v.CX*sx+tx, v.CY*sy+ty
			ax, ay := v.X*sx+tx, v.Y*sy+ty
			subdivideCurve3(&result, x0, y0, cx, cy, ax, ay, approxScale)
			x0, y0 = ax, ay
		} else {
			ax, ay := v.X*sx+tx, v.Y*sy+ty
			result = append(result, FlatVertex{X: ax, Y: ay, Cmd: PathCmdLineTo})
			x0, y0 = ax, ay
		}
	}
	return result
}

// PathCmdMoveTo and PathCmdLineTo match the AGG basics constants.
// We redefine them here to avoid importing the full basics package in the data package.
const (
	PathCmdMoveTo uint32 = 1
	PathCmdLineTo uint32 = 2
	PathCmdStop   uint32 = 0
)

// subdivideCurve3 recursively flattens a quadratic bezier and appends LineTo vertices.
func subdivideCurve3(out *[]FlatVertex, x1, y1, cx, cy, x2, y2, scale float64) {
	// Compute midpoints
	x12 := (x1 + cx) * 0.5
	y12 := (y1 + cy) * 0.5
	x23 := (cx + x2) * 0.5
	y23 := (cy + y2) * 0.5
	x123 := (x12 + x23) * 0.5
	y123 := (y12 + y23) * 0.5

	// Distance from midpoint to chord
	dx := x2 - x1
	dy := y2 - y1
	d := math.Abs((cx-x2)*dy - (cy-y2)*dx)

	if d*d <= 0.5*scale*scale*(dx*dx+dy*dy) {
		// Flat enough
		*out = append(*out, FlatVertex{X: x2, Y: y2, Cmd: PathCmdLineTo})
		return
	}

	subdivideCurve3(out, x1, y1, x12, y12, x123, y123, scale)
	subdivideCurve3(out, x123, y123, x23, y23, x2, y2, scale)
}

// LoadShapes parses the embedded shapes.txt and returns all shapes.
func LoadShapes() []RawShape {
	return ParseShapes(ShapesTxt)
}
