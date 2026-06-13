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

	"github.com/cwbudde/agg_go/internal/curves"
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
		if line == "" {
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
// approxScale is the device-space approximation scale, exactly mirroring AGG's
// pipeline order path_storage → conv_curve(approximation_scale) → conv_transform:
// curves are subdivided in WORLD space (raw control points) with the given
// approximation scale, and only the resulting polyline points are transformed
// to device coordinates. For a uniform scale this is mathematically identical
// to subdividing the already-transformed bezier to a ~0.5px device tolerance.
func FlattenPath(p *RawPath, sx, sy, tx, ty, approxScale float64) []FlatVertex {
	if len(p.Vertices) == 0 {
		return nil
	}
	result := make([]FlatVertex, 0, len(p.Vertices)*2)

	// First vertex is always MoveTo
	v0 := p.Vertices[0]
	result = append(result, FlatVertex{X: v0.X*sx + tx, Y: v0.Y*sy + ty, Cmd: PathCmdMoveTo})

	curve := curves.NewCurve3Div()
	curve.SetApproximationScale(approxScale)

	x0, y0 := v0.X, v0.Y // world-space current point
	for i := 1; i < len(p.Vertices); i++ {
		v := p.Vertices[i]
		if v.IsCurve {
			// Quadratic bezier in WORLD space from (x0,y0) via control
			// (v.CX,v.CY) to (v.X,v.Y), subdivided faithfully (agg::curve3_div).
			curve.Reset()
			curve.Init(x0, y0, v.CX, v.CY, v.X, v.Y)
			curve.Rewind(0)
			first := true
			for {
				px, py, cmd := curve.Vertex()
				if cmd == 0 { // PathCmdStop
					break
				}
				if first {
					// Skip the curve's start point; it duplicates the
					// previous segment's endpoint already in result.
					first = false
					continue
				}
				result = append(result, FlatVertex{X: px*sx + tx, Y: py*sy + ty, Cmd: PathCmdLineTo})
			}
			x0, y0 = v.X, v.Y
		} else {
			result = append(result, FlatVertex{X: v.X*sx + tx, Y: v.Y*sy + ty, Cmd: PathCmdLineTo})
			x0, y0 = v.X, v.Y
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

// LoadShapes parses the embedded shapes.txt and returns all shapes.
func LoadShapes() []RawShape {
	return ParseShapes(ShapesTxt)
}
