// Package outline provides anti-aliased outline rendering functionality.
// This implements a port of AGG's agg_renderer_outline_aa.h distance interpolators.
package outline

import (
	"agg_go/internal/primitives"
)

// DistanceInterpolator0 provides basic distance interpolation for anti-aliased lines.
// This is equivalent to AGG's distance_interpolator0 class.
type DistanceInterpolator0 struct {
	dx   int // x delta
	dy   int // y delta
	dist int // current distance
}

// NewDistanceInterpolator0 creates a new basic distance interpolator.
func NewDistanceInterpolator0(x1, y1, x2, y2, x, y int) *DistanceInterpolator0 {
	d := &DistanceInterpolator0{}
	d.dx = primitives.LineMR(x2) - primitives.LineMR(x1)
	d.dy = primitives.LineMR(y2) - primitives.LineMR(y1)
	d.dist = (primitives.LineMR(x+primitives.LineSubpixelScale/2)-primitives.LineMR(x2))*d.dy -
		(primitives.LineMR(y+primitives.LineSubpixelScale/2)-primitives.LineMR(y2))*d.dx

	d.dx <<= primitives.LineMRSubpixelShift
	d.dy <<= primitives.LineMRSubpixelShift

	return d
}

// IncX increments x position and updates distance.
func (d *DistanceInterpolator0) IncX() {
	d.dist += d.dy
}

// Dist returns the current distance.
func (d *DistanceInterpolator0) Dist() int {
	return d.dist
}

// DistanceInterpolator00 provides dual distance interpolation for pie segments.
// This is equivalent to AGG's distance_interpolator00 class.
type DistanceInterpolator00 struct {
	dx1   int // x delta 1
	dy1   int // y delta 1
	dx2   int // x delta 2
	dy2   int // y delta 2
	dist1 int // current distance 1
	dist2 int // current distance 2
}

// NewDistanceInterpolator00 creates a new dual distance interpolator.
func NewDistanceInterpolator00(xc, yc, x1, y1, x2, y2, x, y int) *DistanceInterpolator00 {
	d := &DistanceInterpolator00{}
	d.dx1 = primitives.LineMR(x1) - primitives.LineMR(xc)
	d.dy1 = primitives.LineMR(y1) - primitives.LineMR(yc)
	d.dx2 = primitives.LineMR(x2) - primitives.LineMR(xc)
	d.dy2 = primitives.LineMR(y2) - primitives.LineMR(yc)
	d.dist1 = (primitives.LineMR(x+primitives.LineSubpixelScale/2)-primitives.LineMR(x1))*d.dy1 -
		(primitives.LineMR(y+primitives.LineSubpixelScale/2)-primitives.LineMR(y1))*d.dx1
	d.dist2 = (primitives.LineMR(x+primitives.LineSubpixelScale/2)-primitives.LineMR(x2))*d.dy2 -
		(primitives.LineMR(y+primitives.LineSubpixelScale/2)-primitives.LineMR(y2))*d.dx2

	d.dx1 <<= primitives.LineMRSubpixelShift
	d.dy1 <<= primitives.LineMRSubpixelShift
	d.dx2 <<= primitives.LineMRSubpixelShift
	d.dy2 <<= primitives.LineMRSubpixelShift

	return d
}

// IncX increments x position and updates both distances.
func (d *DistanceInterpolator00) IncX() {
	d.dist1 += d.dy1
	d.dist2 += d.dy2
}

// Dist1 returns the first distance.
func (d *DistanceInterpolator00) Dist1() int {
	return d.dist1
}

// Dist2 returns the second distance.
func (d *DistanceInterpolator00) Dist2() int {
	return d.dist2
}

// DistanceInterpolator1 provides enhanced distance interpolation with directional movement.
// This is equivalent to AGG's distance_interpolator1 class.
type DistanceInterpolator1 struct {
	dx   int // x delta
	dy   int // y delta
	dist int // current distance
}

// NewDistanceInterpolator1 creates a new enhanced distance interpolator.
func NewDistanceInterpolator1(x1, y1, x2, y2, x, y int) *DistanceInterpolator1 {
	d := &DistanceInterpolator1{}
	d.dx = x2 - x1
	d.dy = y2 - y1
	d.dist = int(float64(x+primitives.LineSubpixelScale/2-x2)*float64(d.dy) -
		float64(y+primitives.LineSubpixelScale/2-y2)*float64(d.dx) + 0.5)

	d.dx <<= primitives.LineSubpixelShift
	d.dy <<= primitives.LineSubpixelShift

	return d
}

// IncX increments x position.
func (d *DistanceInterpolator1) IncX() {
	d.dist += d.dy
}

// DecX decrements x position.
func (d *DistanceInterpolator1) DecX() {
	d.dist -= d.dy
}

// IncY increments y position.
func (d *DistanceInterpolator1) IncY() {
	d.dist -= d.dx
}

// DecY decrements y position.
func (d *DistanceInterpolator1) DecY() {
	d.dist += d.dx
}

// IncXWithDY increments x position with y delta.
func (d *DistanceInterpolator1) IncXWithDY(dy int) {
	d.dist += d.dy
	if dy > 0 {
		d.dist -= d.dx
	}
	if dy < 0 {
		d.dist += d.dx
	}
}

// DecXWithDY decrements x position with y delta.
func (d *DistanceInterpolator1) DecXWithDY(dy int) {
	d.dist -= d.dy
	if dy > 0 {
		d.dist -= d.dx
	}
	if dy < 0 {
		d.dist += d.dx
	}
}

// IncYWithDX increments y position with x delta.
func (d *DistanceInterpolator1) IncYWithDX(dx int) {
	d.dist -= d.dx
	if dx > 0 {
		d.dist += d.dy
	}
	if dx < 0 {
		d.dist -= d.dy
	}
}

// DecYWithDX decrements y position with x delta.
func (d *DistanceInterpolator1) DecYWithDX(dx int) {
	d.dist += d.dx
	if dx > 0 {
		d.dist += d.dy
	}
	if dx < 0 {
		d.dist -= d.dy
	}
}

// Dist returns the current distance.
func (d *DistanceInterpolator1) Dist() int {
	return d.dist
}

// DX returns the x delta.
func (d *DistanceInterpolator1) DX() int {
	return d.dx
}

// DY returns the y delta.
func (d *DistanceInterpolator1) DY() int {
	return d.dy
}

// DistanceInterpolator2 provides distance interpolation with start distance tracking.
// This is equivalent to AGG's distance_interpolator2 class.
type DistanceInterpolator2 struct {
	dx        int // x delta
	dy        int // y delta
	dxStart   int // start x delta
	dyStart   int // start y delta
	dist      int // current distance
	distStart int // start distance
}

// NewDistanceInterpolator2Start creates a new start-tracking distance interpolator.
func NewDistanceInterpolator2Start(x1, y1, x2, y2, sx, sy, x, y int) *DistanceInterpolator2 {
	d := &DistanceInterpolator2{}
	d.dx = x2 - x1
	d.dy = y2 - y1
	d.dxStart = primitives.LineMR(sx) - primitives.LineMR(x1)
	d.dyStart = primitives.LineMR(sy) - primitives.LineMR(y1)

	d.dist = int(float64(x+primitives.LineSubpixelScale/2-x2)*float64(d.dy) -
		float64(y+primitives.LineSubpixelScale/2-y2)*float64(d.dx) + 0.5)

	d.distStart = (primitives.LineMR(x+primitives.LineSubpixelScale/2)-primitives.LineMR(sx))*d.dyStart -
		(primitives.LineMR(y+primitives.LineSubpixelScale/2)-primitives.LineMR(sy))*d.dxStart

	d.dx <<= primitives.LineSubpixelShift
	d.dy <<= primitives.LineSubpixelShift
	d.dxStart <<= primitives.LineMRSubpixelShift
	d.dyStart <<= primitives.LineMRSubpixelShift

	return d
}

// NewDistanceInterpolator2End creates a new end-tracking distance interpolator.
func NewDistanceInterpolator2End(x1, y1, x2, y2, ex, ey, x, y int) *DistanceInterpolator2 {
	d := &DistanceInterpolator2{}
	d.dx = x2 - x1
	d.dy = y2 - y1
	d.dxStart = primitives.LineMR(ex) - primitives.LineMR(x2)
	d.dyStart = primitives.LineMR(ey) - primitives.LineMR(y2)

	d.dist = int(float64(x+primitives.LineSubpixelScale/2-x2)*float64(d.dy) -
		float64(y+primitives.LineSubpixelScale/2-y2)*float64(d.dx) + 0.5)

	d.distStart = (primitives.LineMR(x+primitives.LineSubpixelScale/2)-primitives.LineMR(ex))*d.dyStart -
		(primitives.LineMR(y+primitives.LineSubpixelScale/2)-primitives.LineMR(ey))*d.dxStart

	d.dx <<= primitives.LineSubpixelShift
	d.dy <<= primitives.LineSubpixelShift
	d.dxStart <<= primitives.LineMRSubpixelShift
	d.dyStart <<= primitives.LineMRSubpixelShift

	return d
}

// IncX increments x position.
func (d *DistanceInterpolator2) IncX() {
	d.dist += d.dy
	d.distStart += d.dyStart
}

// DecX decrements x position.
func (d *DistanceInterpolator2) DecX() {
	d.dist -= d.dy
	d.distStart -= d.dyStart
}

// IncY increments y position.
func (d *DistanceInterpolator2) IncY() {
	d.dist -= d.dx
	d.distStart -= d.dxStart
}

// DecY decrements y position.
func (d *DistanceInterpolator2) DecY() {
	d.dist += d.dx
	d.distStart += d.dxStart
}

// IncXWithDY increments x position with y delta.
func (d *DistanceInterpolator2) IncXWithDY(dy int) {
	d.dist += d.dy
	d.distStart += d.dyStart
	if dy > 0 {
		d.dist -= d.dx
		d.distStart -= d.dxStart
	}
	if dy < 0 {
		d.dist += d.dx
		d.distStart += d.dxStart
	}
}

// DecXWithDY decrements x position with y delta.
func (d *DistanceInterpolator2) DecXWithDY(dy int) {
	d.dist -= d.dy
	d.distStart -= d.dyStart
	if dy > 0 {
		d.dist -= d.dx
		d.distStart -= d.dxStart
	}
	if dy < 0 {
		d.dist += d.dx
		d.distStart += d.dxStart
	}
}

// IncYWithDX increments y position with x delta.
func (d *DistanceInterpolator2) IncYWithDX(dx int) {
	d.dist -= d.dx
	d.distStart -= d.dxStart
	if dx > 0 {
		d.dist += d.dy
		d.distStart += d.dyStart
	}
	if dx < 0 {
		d.dist -= d.dy
		d.distStart -= d.dyStart
	}
}

// DecYWithDX decrements y position with x delta.
func (d *DistanceInterpolator2) DecYWithDX(dx int) {
	d.dist += d.dx
	d.distStart += d.dxStart
	if dx > 0 {
		d.dist += d.dy
		d.distStart += d.dyStart
	}
	if dx < 0 {
		d.dist -= d.dy
		d.distStart -= d.dyStart
	}
}

// Dist returns the current distance.
func (d *DistanceInterpolator2) Dist() int {
	return d.dist
}

// DistStart returns the start distance.
func (d *DistanceInterpolator2) DistStart() int {
	return d.distStart
}

// DistEnd returns the end distance (alias for start).
func (d *DistanceInterpolator2) DistEnd() int {
	return d.distStart
}

// DX returns the x delta.
func (d *DistanceInterpolator2) DX() int {
	return d.dx
}

// DY returns the y delta.
func (d *DistanceInterpolator2) DY() int {
	return d.dy
}

// DXStart returns the start x delta.
func (d *DistanceInterpolator2) DXStart() int {
	return d.dxStart
}

// DYStart returns the start y delta.
func (d *DistanceInterpolator2) DYStart() int {
	return d.dyStart
}

// DXEnd returns the end x delta (alias for start).
func (d *DistanceInterpolator2) DXEnd() int {
	return d.dxStart
}

// DYEnd returns the end y delta (alias for start).
func (d *DistanceInterpolator2) DYEnd() int {
	return d.dyStart
}

// DistanceInterpolator3 provides distance interpolation with both start and end tracking.
// This is equivalent to AGG's distance_interpolator3 class.
type DistanceInterpolator3 struct {
	dx        int // x delta
	dy        int // y delta
	dxStart   int // start x delta
	dyStart   int // start y delta
	dxEnd     int // end x delta
	dyEnd     int // end y delta
	dist      int // current distance
	distStart int // start distance
	distEnd   int // end distance
}

// NewDistanceInterpolator3 creates a new full-tracking distance interpolator.
func NewDistanceInterpolator3(x1, y1, x2, y2, sx, sy, ex, ey, x, y int) *DistanceInterpolator3 {
	d := &DistanceInterpolator3{}
	d.dx = x2 - x1
	d.dy = y2 - y1
	d.dxStart = primitives.LineMR(sx) - primitives.LineMR(x1)
	d.dyStart = primitives.LineMR(sy) - primitives.LineMR(y1)
	d.dxEnd = primitives.LineMR(ex) - primitives.LineMR(x2)
	d.dyEnd = primitives.LineMR(ey) - primitives.LineMR(y2)

	d.dist = int(float64(x+primitives.LineSubpixelScale/2-x2)*float64(d.dy) -
		float64(y+primitives.LineSubpixelScale/2-y2)*float64(d.dx) + 0.5)

	d.distStart = (primitives.LineMR(x+primitives.LineSubpixelScale/2)-primitives.LineMR(sx))*d.dyStart -
		(primitives.LineMR(y+primitives.LineSubpixelScale/2)-primitives.LineMR(sy))*d.dxStart

	d.distEnd = (primitives.LineMR(x+primitives.LineSubpixelScale/2)-primitives.LineMR(ex))*d.dyEnd -
		(primitives.LineMR(y+primitives.LineSubpixelScale/2)-primitives.LineMR(ey))*d.dxEnd

	d.dx <<= primitives.LineSubpixelShift
	d.dy <<= primitives.LineSubpixelShift
	d.dxStart <<= primitives.LineMRSubpixelShift
	d.dyStart <<= primitives.LineMRSubpixelShift
	d.dxEnd <<= primitives.LineMRSubpixelShift
	d.dyEnd <<= primitives.LineMRSubpixelShift

	return d
}

// IncX increments x position.
func (d *DistanceInterpolator3) IncX() {
	d.dist += d.dy
	d.distStart += d.dyStart
	d.distEnd += d.dyEnd
}

// DecX decrements x position.
func (d *DistanceInterpolator3) DecX() {
	d.dist -= d.dy
	d.distStart -= d.dyStart
	d.distEnd -= d.dyEnd
}

// IncY increments y position.
func (d *DistanceInterpolator3) IncY() {
	d.dist -= d.dx
	d.distStart -= d.dxStart
	d.distEnd -= d.dxEnd
}

// DecY decrements y position.
func (d *DistanceInterpolator3) DecY() {
	d.dist += d.dx
	d.distStart += d.dxStart
	d.distEnd += d.dxEnd
}

// IncXWithDY increments x position with y delta.
func (d *DistanceInterpolator3) IncXWithDY(dy int) {
	d.dist += d.dy
	d.distStart += d.dyStart
	d.distEnd += d.dyEnd
	if dy > 0 {
		d.dist -= d.dx
		d.distStart -= d.dxStart
		d.distEnd -= d.dxEnd
	}
	if dy < 0 {
		d.dist += d.dx
		d.distStart += d.dxStart
		d.distEnd += d.dxEnd
	}
}

// DecXWithDY decrements x position with y delta.
func (d *DistanceInterpolator3) DecXWithDY(dy int) {
	d.dist -= d.dy
	d.distStart -= d.dyStart
	d.distEnd -= d.dyEnd
	if dy > 0 {
		d.dist -= d.dx
		d.distStart -= d.dxStart
		d.distEnd -= d.dxEnd
	}
	if dy < 0 {
		d.dist += d.dx
		d.distStart += d.dxStart
		d.distEnd += d.dxEnd
	}
}

// IncYWithDX increments y position with x delta.
func (d *DistanceInterpolator3) IncYWithDX(dx int) {
	d.dist -= d.dx
	d.distStart -= d.dxStart
	d.distEnd -= d.dxEnd
	if dx > 0 {
		d.dist += d.dy
		d.distStart += d.dyStart
		d.distEnd += d.dyEnd
	}
	if dx < 0 {
		d.dist -= d.dy
		d.distStart -= d.dyStart
		d.distEnd -= d.dyEnd
	}
}

// DecYWithDX decrements y position with x delta.
func (d *DistanceInterpolator3) DecYWithDX(dx int) {
	d.dist += d.dx
	d.distStart += d.dxStart
	d.distEnd += d.dxEnd
	if dx > 0 {
		d.dist += d.dy
		d.distStart += d.dyStart
		d.distEnd += d.dyEnd
	}
	if dx < 0 {
		d.dist -= d.dy
		d.distStart -= d.dyStart
		d.distEnd -= d.dyEnd
	}
}

// Dist returns the current distance.
func (d *DistanceInterpolator3) Dist() int {
	return d.dist
}

// DistStart returns the start distance.
func (d *DistanceInterpolator3) DistStart() int {
	return d.distStart
}

// DistEnd returns the end distance.
func (d *DistanceInterpolator3) DistEnd() int {
	return d.distEnd
}

// DX returns the x delta.
func (d *DistanceInterpolator3) DX() int {
	return d.dx
}

// DY returns the y delta.
func (d *DistanceInterpolator3) DY() int {
	return d.dy
}

// DXStart returns the start x delta.
func (d *DistanceInterpolator3) DXStart() int {
	return d.dxStart
}

// DYStart returns the start y delta.
func (d *DistanceInterpolator3) DYStart() int {
	return d.dyStart
}

// DXEnd returns the end x delta.
func (d *DistanceInterpolator3) DXEnd() int {
	return d.dxEnd
}

// DYEnd returns the end y delta.
func (d *DistanceInterpolator3) DYEnd() int {
	return d.dyEnd
}
