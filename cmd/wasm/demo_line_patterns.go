package main

import (
	"math"
	"strconv"
	"strings"

	"github.com/MeKo-Christian/agg_go/internal/demo/linepatterns"
)

// Port of AGG C++ line_patterns.cpp (web variant).
const (
	linePatternPaddingMin      = 32.0
	linePatternPaddingMax      = 72.0
	linePatternControlHitDist  = 20.0
	linePatternCurveHitDist    = 14.0
	linePatternCurveSampleStep = 48
)

const (
	linePatternDragModeNone = iota
	linePatternDragModePoint
	linePatternDragModeCurve
)

var (
	linePatternScaleX = 1.0
	linePatternStartX = 0.0

	linePatternCurves            []linepatterns.Curve
	linePatternCurvesInitialized bool
	linePatternCurvesCustomized  bool
	linePatternCanvasW           int
	linePatternCanvasH           int

	linePatternDragMode      = linePatternDragModeNone
	linePatternSelectedCurve = -1
	linePatternSelectedPoint = -1
	linePatternDragLastX     float64
	linePatternDragLastY     float64
)

func setLinePatternScaleX(v float64) {
	if v < 0.2 {
		v = 0.2
	}
	if v > 3.0 {
		v = 3.0
	}
	linePatternScaleX = v
}

func setLinePatternStartX(v float64) {
	if v < 0 {
		v = 0
	}
	if v > 10 {
		v = 10
	}
	linePatternStartX = v
}

func linePatternPadding(w, h int) float64 {
	pad := math.Min(float64(w), float64(h)) * 0.08
	if pad < linePatternPaddingMin {
		return linePatternPaddingMin
	}
	if pad > linePatternPaddingMax {
		return linePatternPaddingMax
	}
	return pad
}

func stretchLinePatternCurves(w, h int) []linepatterns.Curve {
	curves := linepatterns.DefaultCurves()
	if len(curves) == 0 || w <= 0 || h <= 0 {
		return curves
	}

	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64
	for _, c := range curves {
		points := [8]float64{c.X1, c.Y1, c.X2, c.Y2, c.X3, c.Y3, c.X4, c.Y4}
		for i := 0; i < len(points); i += 2 {
			x := points[i]
			y := points[i+1]
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
			if y < minY {
				minY = y
			}
			if y > maxY {
				maxY = y
			}
		}
	}

	pad := linePatternPadding(w, h)
	availW := math.Max(1, float64(w)-2*pad)
	availH := math.Max(1, float64(h)-2*pad)
	srcW := math.Max(1, maxX-minX)
	srcH := math.Max(1, maxY-minY)
	scale := math.Min(availW/srcW, availH/srcH)
	offX := (float64(w)-srcW*scale)*0.5 - minX*scale
	offY := (float64(h)-srcH*scale)*0.5 - minY*scale

	out := make([]linepatterns.Curve, len(curves))
	for i, c := range curves {
		out[i] = linepatterns.Curve{
			X1: offX + c.X1*scale, Y1: offY + c.Y1*scale,
			X2: offX + c.X2*scale, Y2: offY + c.Y2*scale,
			X3: offX + c.X3*scale, Y3: offY + c.Y3*scale,
			X4: offX + c.X4*scale, Y4: offY + c.Y4*scale,
		}
	}
	return out
}

func ensureLinePatternCurves() {
	w, h := ctx.Width(), ctx.Height()
	if !linePatternCurvesInitialized ||
		(!linePatternCurvesCustomized && (linePatternCanvasW != w || linePatternCanvasH != h)) {
		linePatternCurves = stretchLinePatternCurves(w, h)
		linePatternCurvesInitialized = true
		linePatternCanvasW = w
		linePatternCanvasH = h
	}
}

func linePatternControlPoint(c *linepatterns.Curve, idx int) (*float64, *float64) {
	switch idx {
	case 0:
		return &c.X1, &c.Y1
	case 1:
		return &c.X2, &c.Y2
	case 2:
		return &c.X3, &c.Y3
	default:
		return &c.X4, &c.Y4
	}
}

func linePatternClamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func clampLinePatternPoint(x, y float64) (float64, float64) {
	if ctx == nil {
		return x, y
	}
	pad := math.Max(8, linePatternPadding(ctx.Width(), ctx.Height())*0.25)
	maxX := math.Max(pad, float64(ctx.Width())-pad)
	maxY := math.Max(pad, float64(ctx.Height())-pad)
	return linePatternClamp(x, pad, maxX), linePatternClamp(y, pad, maxY)
}

func cubicPoint(c linepatterns.Curve, t float64) (float64, float64) {
	u := 1 - t
	tt := t * t
	uu := u * u
	uuu := uu * u
	ttt := tt * t
	x := uuu*c.X1 + 3*uu*t*c.X2 + 3*u*tt*c.X3 + ttt*c.X4
	y := uuu*c.Y1 + 3*uu*t*c.Y2 + 3*u*tt*c.Y3 + ttt*c.Y4
	return x, y
}

func linePatternDistanceSquared(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return dx*dx + dy*dy
}

func linePatternSegmentDistanceSquared(px, py, ax, ay, bx, by float64) float64 {
	abx := bx - ax
	aby := by - ay
	den := abx*abx + aby*aby
	if den <= 0 {
		return linePatternDistanceSquared(px, py, ax, ay)
	}
	t := ((px-ax)*abx + (py-ay)*aby) / den
	t = linePatternClamp(t, 0, 1)
	cx := ax + abx*t
	cy := ay + aby*t
	return linePatternDistanceSquared(px, py, cx, cy)
}

func linePatternNearestCurve(x, y float64) (idx int, dist2 float64) {
	idx = -1
	dist2 = math.MaxFloat64
	for i, c := range linePatternCurves {
		prevX, prevY := cubicPoint(c, 0)
		best := math.MaxFloat64
		for step := 1; step <= linePatternCurveSampleStep; step++ {
			t := float64(step) / float64(linePatternCurveSampleStep)
			curX, curY := cubicPoint(c, t)
			d := linePatternSegmentDistanceSquared(x, y, prevX, prevY, curX, curY)
			if d < best {
				best = d
			}
			prevX, prevY = curX, curY
		}
		if best < dist2 {
			dist2 = best
			idx = i
		}
	}
	return idx, dist2
}

func encodeLinePatternCurves() string {
	ensureLinePatternCurves()
	parts := make([]string, len(linePatternCurves))
	for i, c := range linePatternCurves {
		values := []string{
			strconv.FormatFloat(c.X1, 'f', 1, 64),
			strconv.FormatFloat(c.Y1, 'f', 1, 64),
			strconv.FormatFloat(c.X2, 'f', 1, 64),
			strconv.FormatFloat(c.Y2, 'f', 1, 64),
			strconv.FormatFloat(c.X3, 'f', 1, 64),
			strconv.FormatFloat(c.Y3, 'f', 1, 64),
			strconv.FormatFloat(c.X4, 'f', 1, 64),
			strconv.FormatFloat(c.Y4, 'f', 1, 64),
		}
		parts[i] = strings.Join(values, ",")
	}
	return strings.Join(parts, ";")
}

func setLinePatternCurvesEncoded(encoded string) bool {
	base := linepatterns.DefaultCurves()
	chunks := strings.Split(encoded, ";")
	if len(chunks) != len(base) {
		return false
	}

	curves := make([]linepatterns.Curve, len(base))
	for i, chunk := range chunks {
		fields := strings.Split(chunk, ",")
		if len(fields) != 8 {
			return false
		}
		values := make([]float64, 8)
		for j, field := range fields {
			v, err := strconv.ParseFloat(field, 64)
			if err != nil {
				return false
			}
			values[j] = v
		}
		curves[i] = linepatterns.Curve{
			X1: values[0], Y1: values[1],
			X2: values[2], Y2: values[3],
			X3: values[4], Y3: values[5],
			X4: values[6], Y4: values[7],
		}
	}

	linePatternCurves = curves
	linePatternCurvesInitialized = true
	linePatternCurvesCustomized = true
	if ctx != nil {
		linePatternCanvasW = ctx.Width()
		linePatternCanvasH = ctx.Height()
	}
	return true
}

func handleLinePatternsMouseDown(x, y float64) bool {
	ensureLinePatternCurves()

	bestCurve := -1
	bestPoint := -1
	bestDist2 := linePatternControlHitDist * linePatternControlHitDist
	for i := range linePatternCurves {
		for j := 0; j < 4; j++ {
			px, py := linePatternControlPoint(&linePatternCurves[i], j)
			d2 := linePatternDistanceSquared(x, y, *px, *py)
			if d2 <= bestDist2 {
				bestDist2 = d2
				bestCurve = i
				bestPoint = j
			}
		}
	}
	if bestCurve >= 0 {
		linePatternDragMode = linePatternDragModePoint
		linePatternSelectedCurve = bestCurve
		linePatternSelectedPoint = bestPoint
		linePatternCurvesCustomized = true
		return true
	}

	curveIdx, dist2 := linePatternNearestCurve(x, y)
	if curveIdx >= 0 && dist2 <= linePatternCurveHitDist*linePatternCurveHitDist {
		linePatternDragMode = linePatternDragModeCurve
		linePatternSelectedCurve = curveIdx
		linePatternSelectedPoint = -1
		linePatternDragLastX = x
		linePatternDragLastY = y
		linePatternCurvesCustomized = true
		return true
	}

	return false
}

func handleLinePatternsMouseMove(x, y float64) bool {
	if linePatternDragMode == linePatternDragModeNone || linePatternSelectedCurve < 0 {
		return false
	}

	switch linePatternDragMode {
	case linePatternDragModePoint:
		px, py := linePatternControlPoint(&linePatternCurves[linePatternSelectedCurve], linePatternSelectedPoint)
		*px, *py = clampLinePatternPoint(x, y)
		return true
	case linePatternDragModeCurve:
		dx := x - linePatternDragLastX
		dy := y - linePatternDragLastY
		c := &linePatternCurves[linePatternSelectedCurve]
		for i := 0; i < 4; i++ {
			px, py := linePatternControlPoint(c, i)
			*px += dx
			*py += dy
			*px, *py = clampLinePatternPoint(*px, *py)
		}
		linePatternDragLastX = x
		linePatternDragLastY = y
		return true
	}

	return false
}

func handleLinePatternsMouseUp() {
	linePatternDragMode = linePatternDragModeNone
	linePatternSelectedCurve = -1
	linePatternSelectedPoint = -1
}

func drawLinePatternsDemo() {
	ensureLinePatternCurves()
	linepatterns.DrawCurves(ctx.GetImage(), linePatternScaleX, linePatternStartX, linePatternCurves)
}
