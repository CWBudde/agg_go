// Package outline provides anti-aliased outline rendering functionality.
// This implements a port of AGG's renderer_outline_aa template class.
package outline

import (
	"agg_go/internal/basics"
	"agg_go/internal/primitives"
)

// BaseRendererInterface defines the interface for base renderers.
type BaseRendererInterface[C any] interface {
	Width() int
	Height() int
	BlendSolidHSpan(x, y, length int, color C, covers []basics.CoverType)
	BlendSolidVSpan(x, y, length int, color C, covers []basics.CoverType)
}

// ColorTypeInterface defines the interface for color types.
type ColorTypeInterface interface {
	// Add any color-specific methods needed
}

// RendererOutlineAA provides anti-aliased outline rendering.
// This is equivalent to AGG's renderer_outline_aa template class.
type RendererOutlineAA[BaseRenderer BaseRendererInterface[C], C any] struct {
	ren      BaseRenderer   // Base renderer
	profile  *LineProfileAA // Line profile
	color    C              // Current color
	clipBox  basics.RectI   // Clipping box
	clipping bool           // Clipping enabled
}

// NewRendererOutlineAA creates a new anti-aliased outline renderer.
func NewRendererOutlineAA[BaseRenderer BaseRendererInterface[C], C any](
	ren BaseRenderer, prof *LineProfileAA,
) *RendererOutlineAA[BaseRenderer, C] {
	return &RendererOutlineAA[BaseRenderer, C]{
		ren:      ren,
		profile:  prof,
		clipBox:  basics.RectI{X1: 0, Y1: 0, X2: 0, Y2: 0},
		clipping: false,
	}
}

// Attach attaches a base renderer.
func (r *RendererOutlineAA[BaseRenderer, C]) Attach(ren BaseRenderer) {
	r.ren = ren
}

// Color sets the drawing color.
func (r *RendererOutlineAA[BaseRenderer, C]) Color(c C) {
	r.color = c
}

// GetColor returns the current drawing color.
func (r *RendererOutlineAA[BaseRenderer, C]) GetColor() C {
	return r.color
}

// Profile sets the line profile.
func (r *RendererOutlineAA[BaseRenderer, C]) Profile(prof *LineProfileAA) {
	r.profile = prof
}

// GetProfile returns the current line profile.
func (r *RendererOutlineAA[BaseRenderer, C]) GetProfile() *LineProfileAA {
	return r.profile
}

// SubpixelWidth returns the subpixel width.
func (r *RendererOutlineAA[BaseRenderer, C]) SubpixelWidth() int {
	return r.profile.SubpixelWidth()
}

// ResetClipping disables clipping.
func (r *RendererOutlineAA[BaseRenderer, C]) ResetClipping() {
	r.clipping = false
}

// ClipBox sets the clipping box.
func (r *RendererOutlineAA[BaseRenderer, C]) ClipBox(x1, y1, x2, y2 float64) {
	r.clipBox.X1 = int(x1*primitives.LineSubpixelScale + 0.5)
	r.clipBox.Y1 = int(y1*primitives.LineSubpixelScale + 0.5)
	r.clipBox.X2 = int(x2*primitives.LineSubpixelScale + 0.5)
	r.clipBox.Y2 = int(y2*primitives.LineSubpixelScale + 0.5)
	r.clipping = true
}

// Cover returns coverage value for the given distance.
func (r *RendererOutlineAA[BaseRenderer, C]) Cover(d int) int {
	return int(r.profile.Value(d))
}

// BlendSolidHSpan renders a horizontal span.
func (r *RendererOutlineAA[BaseRenderer, C]) BlendSolidHSpan(x, y, length int, covers []basics.CoverType) {
	r.ren.BlendSolidHSpan(x, y, length, r.color, covers)
}

// BlendSolidVSpan renders a vertical span.
func (r *RendererOutlineAA[BaseRenderer, C]) BlendSolidVSpan(x, y, length int, covers []basics.CoverType) {
	r.ren.BlendSolidVSpan(x, y, length, r.color, covers)
}

// AccurateJoinOnly returns false (not needed for this implementation).
func (r *RendererOutlineAA[BaseRenderer, C]) AccurateJoinOnly() bool {
	return false
}

// SemidotHline renders a horizontal line segment for semidot operations.
func (r *RendererOutlineAA[BaseRenderer, C]) SemidotHline(
	cmp func(int) bool, xc1, yc1, xc2, yc2, x1, y1, x2 int,
) {
	covers := make([]basics.CoverType, MaxHalfWidth*2+4)
	p1 := 0
	x := x1 << primitives.LineSubpixelShift
	y := y1 << primitives.LineSubpixelShift
	w := r.SubpixelWidth()
	di := NewDistanceInterpolator0(xc1, yc1, xc2, yc2, x, y)
	x += primitives.LineSubpixelScale / 2
	y += primitives.LineSubpixelScale / 2

	x0 := x1
	dx := x - xc1
	dy := y - yc1

	for x1 <= x2 {
		d := int(basics.FastSqrt(uint32(dx*dx + dy*dy)))
		covers[p1] = 0
		if cmp(di.Dist()) && d <= w {
			covers[p1] = basics.CoverType(r.Cover(d))
		}
		p1++
		dx += primitives.LineSubpixelScale
		di.IncX()
		x1++
	}

	r.ren.BlendSolidHSpan(x0, y1, p1, r.color, covers[:p1])
}

// Semidot renders a semidot (half circle) shape.
func (r *RendererOutlineAA[BaseRenderer, C]) Semidot(
	cmp func(int) bool, xc1, yc1, xc2, yc2 int,
) {
	if r.clipping && basics.ClippingFlags(xc1, yc1, r.clipBox) != 0 {
		return
	}

	radius := ((r.SubpixelWidth() + primitives.LineSubpixelMask) >> primitives.LineSubpixelShift)
	if radius < 1 {
		radius = 1
	}

	// Use Bresenham ellipse algorithm for circle
	ei := primitives.NewEllipseBresenhamInterpolator(radius, radius)
	dx := 0
	dy := -radius
	dy0 := dy
	dx0 := dx
	x := xc1 >> primitives.LineSubpixelShift
	y := yc1 >> primitives.LineSubpixelShift

	for dy < 0 {
		dx += ei.Dx()
		dy += ei.Dy()

		if dy != dy0 {
			r.SemidotHline(cmp, xc1, yc1, xc2, yc2, x-dx0, y+dy0, x+dx0)
			r.SemidotHline(cmp, xc1, yc1, xc2, yc2, x-dx0, y-dy0, x+dx0)
		}
		dx0 = dx
		dy0 = dy
		ei.Inc()
	}
	r.SemidotHline(cmp, xc1, yc1, xc2, yc2, x-dx0, y+dy0, x+dx0)
}

// PieHline renders a horizontal line segment for pie operations.
func (r *RendererOutlineAA[BaseRenderer, C]) PieHline(
	xc, yc, xp1, yp1, xp2, yp2, xh1, yh1, xh2 int,
) {
	if r.clipping && basics.ClippingFlags(xc, yc, r.clipBox) != 0 {
		return
	}

	covers := make([]basics.CoverType, MaxHalfWidth*2+4)
	p1 := 0
	x := xh1 << primitives.LineSubpixelShift
	y := yh1 << primitives.LineSubpixelShift
	w := r.SubpixelWidth()

	di := NewDistanceInterpolator00(xc, yc, xp1, yp1, xp2, yp2, x, y)
	x += primitives.LineSubpixelScale / 2
	y += primitives.LineSubpixelScale / 2

	xh0 := xh1
	dx := x - xc
	dy := y - yc

	for xh1 <= xh2 {
		d := int(basics.FastSqrt(uint32(dx*dx + dy*dy)))
		covers[p1] = 0
		if di.Dist1() <= 0 && di.Dist2() > 0 && d <= w {
			covers[p1] = basics.CoverType(r.Cover(d))
		}
		p1++
		dx += primitives.LineSubpixelScale
		di.IncX()
		xh1++
	}

	r.ren.BlendSolidHSpan(xh0, yh1, p1, r.color, covers[:p1])
}

// Pie renders a pie segment.
func (r *RendererOutlineAA[BaseRenderer, C]) Pie(xc, yc, x1, y1, x2, y2 int) {
	radius := ((r.SubpixelWidth() + primitives.LineSubpixelMask) >> primitives.LineSubpixelShift)
	if radius < 1 {
		radius = 1
	}

	ei := primitives.NewEllipseBresenhamInterpolator(radius, radius)
	dx := 0
	dy := -radius
	dy0 := dy
	dx0 := dx
	x := xc >> primitives.LineSubpixelShift
	y := yc >> primitives.LineSubpixelShift

	for dy < 0 {
		dx += ei.Dx()
		dy += ei.Dy()

		if dy != dy0 {
			r.PieHline(xc, yc, x1, y1, x2, y2, x-dx0, y+dy0, x+dx0)
			r.PieHline(xc, yc, x1, y1, x2, y2, x-dx0, y-dy0, x+dx0)
		}
		dx0 = dx
		dy0 = dy
		ei.Inc()
	}
	r.PieHline(xc, yc, x1, y1, x2, y2, x-dx0, y+dy0, x+dx0)
}

// Line0NoClip renders a line without clipping (basic line).
func (r *RendererOutlineAA[BaseRenderer, C]) Line0NoClip(lp *primitives.LineParameters) {
	if lp.Len > primitives.LineMaxLength {
		lp1, lp2 := lp.Divide()
		r.Line0NoClip(&lp1)
		r.Line0NoClip(&lp2)
		return
	}

	li := NewLineInterpolatorAA0(r, lp)
	if li.Count() > 0 {
		if li.Vertical() {
			for li.StepVer() {
			}
		} else {
			for li.StepHor() {
			}
		}
	}
}

// Line0 renders a basic line with clipping.
func (r *RendererOutlineAA[BaseRenderer, C]) Line0(lp *primitives.LineParameters) {
	if r.clipping {
		x1 := lp.X1
		y1 := lp.Y1
		x2 := lp.X2
		y2 := lp.Y2
		flags := basics.ClipLineSegment(&x1, &y1, &x2, &y2, r.clipBox)
		if (flags & 4) == 0 {
			if flags != 0 {
				lp2 := primitives.NewLineParameters(x1, y1, x2, y2,
					int(basics.URound(basics.CalcDistance(float64(x1), float64(y1), float64(x2), float64(y2)))))
				r.Line0NoClip(&lp2)
			} else {
				r.Line0NoClip(lp)
			}
		}
	} else {
		r.Line0NoClip(lp)
	}
}

// Line1NoClip renders a line with start cap, without clipping.
func (r *RendererOutlineAA[BaseRenderer, C]) Line1NoClip(lp *primitives.LineParameters, sx, sy int) {
	if lp.Len > primitives.LineMaxLength {
		lp1, lp2 := lp.Divide()
		r.Line1NoClip(&lp1, (lp.X1+sx)>>1, (lp.Y1+sy)>>1)
		r.Line1NoClip(&lp2, lp1.X2+(lp1.Y2-lp1.Y1), lp1.Y2-(lp1.X2-lp1.X1))
		return
	}

	primitives.FixDegenerateBisectrixStart(lp, &sx, &sy)
	li := NewLineInterpolatorAA1(r, lp, sx, sy)
	if li.Vertical() {
		for li.StepVer() {
		}
	} else {
		for li.StepHor() {
		}
	}
}

// Line1 renders a line with start cap and clipping.
func (r *RendererOutlineAA[BaseRenderer, C]) Line1(lp *primitives.LineParameters, sx, sy int) {
	if r.clipping {
		x1 := lp.X1
		y1 := lp.Y1
		x2 := lp.X2
		y2 := lp.Y2
		flags := basics.ClipLineSegment(&x1, &y1, &x2, &y2, r.clipBox)
		if (flags & 4) == 0 {
			if flags != 0 {
				lp2 := primitives.NewLineParameters(x1, y1, x2, y2,
					int(basics.URound(basics.CalcDistance(float64(x1), float64(y1), float64(x2), float64(y2)))))
				if (flags & 1) != 0 {
					sx = x1 + (y2 - y1)
					sy = y1 - (x2 - x1)
				} else {
					for basics.Abs(sx-lp.X1)+basics.Abs(sy-lp.Y1) > lp2.Len {
						sx = (lp.X1 + sx) >> 1
						sy = (lp.Y1 + sy) >> 1
					}
				}
				r.Line1NoClip(&lp2, sx, sy)
			} else {
				r.Line1NoClip(lp, sx, sy)
			}
		}
	} else {
		r.Line1NoClip(lp, sx, sy)
	}
}

// Line2NoClip renders a line with end cap, without clipping.
func (r *RendererOutlineAA[BaseRenderer, C]) Line2NoClip(lp *primitives.LineParameters, ex, ey int) {
	if lp.Len > primitives.LineMaxLength {
		lp1, lp2 := lp.Divide()
		r.Line2NoClip(&lp1, lp1.X2+(lp1.Y2-lp1.Y1), lp1.Y2-(lp1.X2-lp1.X1))
		r.Line2NoClip(&lp2, (lp.X2+ex)>>1, (lp.Y2+ey)>>1)
		return
	}

	primitives.FixDegenerateBisectrixEnd(lp, &ex, &ey)
	li := NewLineInterpolatorAA2(r, lp, ex, ey)
	if li.Vertical() {
		for li.StepVer() {
		}
	} else {
		for li.StepHor() {
		}
	}
}

// Line2 renders a line with end cap and clipping.
func (r *RendererOutlineAA[BaseRenderer, C]) Line2(lp *primitives.LineParameters, ex, ey int) {
	if r.clipping {
		x1 := lp.X1
		y1 := lp.Y1
		x2 := lp.X2
		y2 := lp.Y2
		flags := basics.ClipLineSegment(&x1, &y1, &x2, &y2, r.clipBox)
		if (flags & 4) == 0 {
			if flags != 0 {
				lp2 := primitives.NewLineParameters(x1, y1, x2, y2,
					int(basics.URound(basics.CalcDistance(float64(x1), float64(y1), float64(x2), float64(y2)))))
				if (flags & 2) != 0 {
					ex = x2 + (y2 - y1)
					ey = y2 - (x2 - x1)
				} else {
					for basics.Abs(ex-lp.X2)+basics.Abs(ey-lp.Y2) > lp2.Len {
						ex = (lp.X2 + ex) >> 1
						ey = (lp.Y2 + ey) >> 1
					}
				}
				r.Line2NoClip(&lp2, ex, ey)
			} else {
				r.Line2NoClip(lp, ex, ey)
			}
		}
	} else {
		r.Line2NoClip(lp, ex, ey)
	}
}

// Line3NoClip renders a line with both start and end caps, without clipping.
func (r *RendererOutlineAA[BaseRenderer, C]) Line3NoClip(lp *primitives.LineParameters, sx, sy, ex, ey int) {
	if lp.Len > primitives.LineMaxLength {
		lp1, lp2 := lp.Divide()
		mx := lp1.X2 + (lp1.Y2 - lp1.Y1)
		my := lp1.Y2 - (lp1.X2 - lp1.X1)
		r.Line3NoClip(&lp1, (lp.X1+sx)>>1, (lp.Y1+sy)>>1, mx, my)
		r.Line3NoClip(&lp2, mx, my, (lp.X2+ex)>>1, (lp.Y2+ey)>>1)
		return
	}

	primitives.FixDegenerateBisectrixStart(lp, &sx, &sy)
	primitives.FixDegenerateBisectrixEnd(lp, &ex, &ey)
	li := NewLineInterpolatorAA3(r, lp, sx, sy, ex, ey)
	if li.Vertical() {
		for li.StepVer() {
		}
	} else {
		for li.StepHor() {
		}
	}
}

// Line3 renders a line with both start and end caps and clipping.
func (r *RendererOutlineAA[BaseRenderer, C]) Line3(lp *primitives.LineParameters, sx, sy, ex, ey int) {
	if r.clipping {
		x1 := lp.X1
		y1 := lp.Y1
		x2 := lp.X2
		y2 := lp.Y2
		flags := basics.ClipLineSegment(&x1, &y1, &x2, &y2, r.clipBox)
		if (flags & 4) == 0 {
			if flags != 0 {
				lp2 := primitives.NewLineParameters(x1, y1, x2, y2,
					int(basics.URound(basics.CalcDistance(float64(x1), float64(y1), float64(x2), float64(y2)))))
				if (flags & 1) != 0 {
					sx = x1 + (y2 - y1)
					sy = y1 - (x2 - x1)
				} else {
					for basics.Abs(sx-lp.X1)+basics.Abs(sy-lp.Y1) > lp2.Len {
						sx = (lp.X1 + sx) >> 1
						sy = (lp.Y1 + sy) >> 1
					}
				}
				if (flags & 2) != 0 {
					ex = x2 + (y2 - y1)
					ey = y2 - (x2 - x1)
				} else {
					for basics.Abs(ex-lp.X2)+basics.Abs(ey-lp.Y2) > lp2.Len {
						ex = (lp.X2 + ex) >> 1
						ey = (lp.Y2 + ey) >> 1
					}
				}
				r.Line3NoClip(&lp2, sx, sy, ex, ey)
			} else {
				r.Line3NoClip(lp, sx, sy, ex, ey)
			}
		}
	} else {
		r.Line3NoClip(lp, sx, sy, ex, ey)
	}
}
