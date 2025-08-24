// Package markers provides marker shape rendering functionality for the AGG Go port.
// This implements the renderer_markers<BaseRenderer> template class from AGG 2.6.
package markers

import (
	"agg_go/internal/basics"
	"agg_go/internal/primitives"
	primitives_pkg "agg_go/internal/renderer/primitives"
)

// RendererMarkers provides marker drawing operations on top of a base renderer.
// This is a port of AGG's renderer_markers<BaseRenderer> template class.
// It embeds RendererPrimitives to inherit all primitive drawing capabilities.
type RendererMarkers[BR primitives_pkg.BaseRenderer[C], C any] struct {
	*primitives_pkg.RendererPrimitives[BR, C]
}

// NewRendererMarkers creates a new marker renderer with the given base renderer.
func NewRendererMarkers[BR primitives_pkg.BaseRenderer[C], C any](ren BR) *RendererMarkers[BR, C] {
	return &RendererMarkers[BR, C]{
		RendererPrimitives: primitives_pkg.NewRendererPrimitives[BR, C](ren),
	}
}

// Visible tests if a marker with the given center and radius would be visible
// within the renderer's bounding clipping box.
func (rm *RendererMarkers[BR, C]) Visible(x, y, r int) bool {
	rc := basics.RectI{X1: x - r, Y1: y - r, X2: x + r, Y2: y + r}
	return rc.Clip(rm.Ren().BoundingClipBox())
}

// Square draws a solid square marker centered at (x, y) with radius r.
func (rm *RendererMarkers[BR, C]) Square(x, y, r int) {
	if rm.Visible(x, y, r) {
		if r > 0 {
			rm.OutlinedRectangle(x-r, y-r, x+r, y+r)
		} else {
			rm.Ren().BlendPixel(x, y, rm.GetFillColor(), basics.CoverFull)
		}
	}
}

// Diamond draws a solid diamond marker centered at (x, y) with radius r.
func (rm *RendererMarkers[BR, C]) Diamond(x, y, r int) {
	if rm.Visible(x, y, r) {
		if r > 0 {
			dy := -r
			dx := 0
			for dy <= 0 {
				rm.Ren().BlendPixel(x-dx, y+dy, rm.GetLineColor(), basics.CoverFull)
				rm.Ren().BlendPixel(x+dx, y+dy, rm.GetLineColor(), basics.CoverFull)
				rm.Ren().BlendPixel(x-dx, y-dy, rm.GetLineColor(), basics.CoverFull)
				rm.Ren().BlendPixel(x+dx, y-dy, rm.GetLineColor(), basics.CoverFull)

				if dx > 0 {
					rm.Ren().BlendHline(x-dx+1, y+dy, x+dx-1, rm.GetFillColor(), basics.CoverFull)
					rm.Ren().BlendHline(x-dx+1, y-dy, x+dx-1, rm.GetFillColor(), basics.CoverFull)
				}
				dy++
				dx++
			}
		} else {
			rm.Ren().BlendPixel(x, y, rm.GetFillColor(), basics.CoverFull)
		}
	}
}

// Circle draws a solid circle marker centered at (x, y) with radius r.
func (rm *RendererMarkers[BR, C]) Circle(x, y, r int) {
	if rm.Visible(x, y, r) {
		if r > 0 {
			rm.OutlinedEllipse(x, y, r, r)
		} else {
			rm.Ren().BlendPixel(x, y, rm.GetFillColor(), basics.CoverFull)
		}
	}
}

// CrossedCircle draws a circle with cross lines extending beyond the circle.
func (rm *RendererMarkers[BR, C]) CrossedCircle(x, y, r int) {
	if rm.Visible(x, y, r) {
		if r > 0 {
			rm.OutlinedEllipse(x, y, r, r)
			r6 := r + (r >> 1)
			if r <= 2 {
				r6++
			}
			r >>= 1
			rm.Ren().BlendHline(x-r6, y, x-r, rm.GetLineColor(), basics.CoverFull)
			rm.Ren().BlendHline(x+r, y, x+r6, rm.GetLineColor(), basics.CoverFull)
			rm.Ren().BlendVline(x, y-r6, y-r, rm.GetLineColor(), basics.CoverFull)
			rm.Ren().BlendVline(x, y+r, y+r6, rm.GetLineColor(), basics.CoverFull)
		} else {
			rm.Ren().BlendPixel(x, y, rm.GetFillColor(), basics.CoverFull)
		}
	}
}

// SemiEllipseLeft draws a left-facing semi-ellipse marker.
func (rm *RendererMarkers[BR, C]) SemiEllipseLeft(x, y, r int) {
	if rm.Visible(x, y, r) {
		if r > 0 {
			r8 := r * 4 / 5
			dy := -r
			dx := 0
			ei := primitives.NewEllipseBresenhamInterpolator(r*3/5, r+r8)
			for dy < r8 {
				dx += ei.Dx()
				dy += ei.Dy()

				rm.Ren().BlendPixel(x+dy, y+dx, rm.GetLineColor(), basics.CoverFull)
				rm.Ren().BlendPixel(x+dy, y-dx, rm.GetLineColor(), basics.CoverFull)

				if ei.Dy() != 0 && dx > 0 {
					rm.Ren().BlendVline(x+dy, y-dx+1, y+dx-1, rm.GetFillColor(), basics.CoverFull)
				}
				ei.Inc()
			}
			rm.Ren().BlendVline(x+dy, y-dx, y+dx, rm.GetLineColor(), basics.CoverFull)
		} else {
			rm.Ren().BlendPixel(x, y, rm.GetFillColor(), basics.CoverFull)
		}
	}
}

// SemiEllipseRight draws a right-facing semi-ellipse marker.
func (rm *RendererMarkers[BR, C]) SemiEllipseRight(x, y, r int) {
	if rm.Visible(x, y, r) {
		if r > 0 {
			r8 := r * 4 / 5
			dy := -r
			dx := 0
			ei := primitives.NewEllipseBresenhamInterpolator(r*3/5, r+r8)
			for dy < r8 {
				dx += ei.Dx()
				dy += ei.Dy()

				rm.Ren().BlendPixel(x-dy, y+dx, rm.GetLineColor(), basics.CoverFull)
				rm.Ren().BlendPixel(x-dy, y-dx, rm.GetLineColor(), basics.CoverFull)

				if ei.Dy() != 0 && dx > 0 {
					rm.Ren().BlendVline(x-dy, y-dx+1, y+dx-1, rm.GetFillColor(), basics.CoverFull)
				}
				ei.Inc()
			}
			rm.Ren().BlendVline(x-dy, y-dx, y+dx, rm.GetLineColor(), basics.CoverFull)
		} else {
			rm.Ren().BlendPixel(x, y, rm.GetFillColor(), basics.CoverFull)
		}
	}
}

// SemiEllipseUp draws an upward-facing semi-ellipse marker.
func (rm *RendererMarkers[BR, C]) SemiEllipseUp(x, y, r int) {
	if rm.Visible(x, y, r) {
		if r > 0 {
			r8 := r * 4 / 5
			dy := -r
			dx := 0
			ei := primitives.NewEllipseBresenhamInterpolator(r*3/5, r+r8)
			for dy < r8 {
				dx += ei.Dx()
				dy += ei.Dy()

				rm.Ren().BlendPixel(x+dx, y-dy, rm.GetLineColor(), basics.CoverFull)
				rm.Ren().BlendPixel(x-dx, y-dy, rm.GetLineColor(), basics.CoverFull)

				if ei.Dy() != 0 && dx > 0 {
					rm.Ren().BlendHline(x-dx+1, y-dy, x+dx-1, rm.GetFillColor(), basics.CoverFull)
				}
				ei.Inc()
			}
			rm.Ren().BlendHline(x-dx, y-dy-1, x+dx, rm.GetLineColor(), basics.CoverFull)
		} else {
			rm.Ren().BlendPixel(x, y, rm.GetFillColor(), basics.CoverFull)
		}
	}
}

// SemiEllipseDown draws a downward-facing semi-ellipse marker.
func (rm *RendererMarkers[BR, C]) SemiEllipseDown(x, y, r int) {
	if rm.Visible(x, y, r) {
		if r > 0 {
			r8 := r * 4 / 5
			dy := -r
			dx := 0
			ei := primitives.NewEllipseBresenhamInterpolator(r*3/5, r+r8)
			for dy < r8 {
				dx += ei.Dx()
				dy += ei.Dy()

				rm.Ren().BlendPixel(x+dx, y+dy, rm.GetLineColor(), basics.CoverFull)
				rm.Ren().BlendPixel(x-dx, y+dy, rm.GetLineColor(), basics.CoverFull)

				if ei.Dy() != 0 && dx > 0 {
					rm.Ren().BlendHline(x-dx+1, y+dy, x+dx-1, rm.GetFillColor(), basics.CoverFull)
				}
				ei.Inc()
			}
			rm.Ren().BlendHline(x-dx, y+dy+1, x+dx, rm.GetLineColor(), basics.CoverFull)
		} else {
			rm.Ren().BlendPixel(x, y, rm.GetFillColor(), basics.CoverFull)
		}
	}
}

// TriangleLeft draws a left-pointing triangle marker.
func (rm *RendererMarkers[BR, C]) TriangleLeft(x, y, r int) {
	if rm.Visible(x, y, r) {
		if r > 0 {
			dy := -r
			dx := 0
			flip := 0
			r6 := r * 3 / 5
			for dy < r6 {
				rm.Ren().BlendPixel(x+dy, y-dx, rm.GetLineColor(), basics.CoverFull)
				rm.Ren().BlendPixel(x+dy, y+dx, rm.GetLineColor(), basics.CoverFull)

				if dx > 0 {
					rm.Ren().BlendVline(x+dy, y-dx+1, y+dx-1, rm.GetFillColor(), basics.CoverFull)
				}
				dy++
				dx += flip
				flip ^= 1
			}
			rm.Ren().BlendVline(x+dy, y-dx, y+dx, rm.GetLineColor(), basics.CoverFull)
		} else {
			rm.Ren().BlendPixel(x, y, rm.GetFillColor(), basics.CoverFull)
		}
	}
}

// TriangleRight draws a right-pointing triangle marker.
func (rm *RendererMarkers[BR, C]) TriangleRight(x, y, r int) {
	if rm.Visible(x, y, r) {
		if r > 0 {
			dy := -r
			dx := 0
			flip := 0
			r6 := r * 3 / 5
			for dy < r6 {
				rm.Ren().BlendPixel(x-dy, y-dx, rm.GetLineColor(), basics.CoverFull)
				rm.Ren().BlendPixel(x-dy, y+dx, rm.GetLineColor(), basics.CoverFull)

				if dx > 0 {
					rm.Ren().BlendVline(x-dy, y-dx+1, y+dx-1, rm.GetFillColor(), basics.CoverFull)
				}
				dy++
				dx += flip
				flip ^= 1
			}
			rm.Ren().BlendVline(x-dy, y-dx, y+dx, rm.GetLineColor(), basics.CoverFull)
		} else {
			rm.Ren().BlendPixel(x, y, rm.GetFillColor(), basics.CoverFull)
		}
	}
}

// TriangleUp draws an upward-pointing triangle marker.
func (rm *RendererMarkers[BR, C]) TriangleUp(x, y, r int) {
	if rm.Visible(x, y, r) {
		if r > 0 {
			dy := -r
			dx := 0
			flip := 0
			r6 := r * 3 / 5
			for dy < r6 {
				rm.Ren().BlendPixel(x-dx, y-dy, rm.GetLineColor(), basics.CoverFull)
				rm.Ren().BlendPixel(x+dx, y-dy, rm.GetLineColor(), basics.CoverFull)

				if dx > 0 {
					rm.Ren().BlendHline(x-dx+1, y-dy, x+dx-1, rm.GetFillColor(), basics.CoverFull)
				}
				dy++
				dx += flip
				flip ^= 1
			}
			rm.Ren().BlendHline(x-dx, y-dy, x+dx, rm.GetLineColor(), basics.CoverFull)
		} else {
			rm.Ren().BlendPixel(x, y, rm.GetFillColor(), basics.CoverFull)
		}
	}
}

// TriangleDown draws a downward-pointing triangle marker.
func (rm *RendererMarkers[BR, C]) TriangleDown(x, y, r int) {
	if rm.Visible(x, y, r) {
		if r > 0 {
			dy := -r
			dx := 0
			flip := 0
			r6 := r * 3 / 5
			for dy < r6 {
				rm.Ren().BlendPixel(x-dx, y+dy, rm.GetLineColor(), basics.CoverFull)
				rm.Ren().BlendPixel(x+dx, y+dy, rm.GetLineColor(), basics.CoverFull)

				if dx > 0 {
					rm.Ren().BlendHline(x-dx+1, y+dy, x+dx-1, rm.GetFillColor(), basics.CoverFull)
				}
				dy++
				dx += flip
				flip ^= 1
			}
			rm.Ren().BlendHline(x-dx, y+dy, x+dx, rm.GetLineColor(), basics.CoverFull)
		} else {
			rm.Ren().BlendPixel(x, y, rm.GetFillColor(), basics.CoverFull)
		}
	}
}

// FourRays draws a four-ray (plus sign) marker with filled center.
func (rm *RendererMarkers[BR, C]) FourRays(x, y, r int) {
	if rm.Visible(x, y, r) {
		if r > 0 {
			dy := -r
			dx := 0
			flip := 0
			r3 := -(r / 3)
			for dy <= r3 {
				rm.Ren().BlendPixel(x-dx, y+dy, rm.GetLineColor(), basics.CoverFull)
				rm.Ren().BlendPixel(x+dx, y+dy, rm.GetLineColor(), basics.CoverFull)
				rm.Ren().BlendPixel(x-dx, y-dy, rm.GetLineColor(), basics.CoverFull)
				rm.Ren().BlendPixel(x+dx, y-dy, rm.GetLineColor(), basics.CoverFull)
				rm.Ren().BlendPixel(x+dy, y-dx, rm.GetLineColor(), basics.CoverFull)
				rm.Ren().BlendPixel(x+dy, y+dx, rm.GetLineColor(), basics.CoverFull)
				rm.Ren().BlendPixel(x-dy, y-dx, rm.GetLineColor(), basics.CoverFull)
				rm.Ren().BlendPixel(x-dy, y+dx, rm.GetLineColor(), basics.CoverFull)

				if dx > 0 {
					rm.Ren().BlendHline(x-dx+1, y+dy, x+dx-1, rm.GetFillColor(), basics.CoverFull)
					rm.Ren().BlendHline(x-dx+1, y-dy, x+dx-1, rm.GetFillColor(), basics.CoverFull)
					rm.Ren().BlendVline(x+dy, y-dx+1, y+dx-1, rm.GetFillColor(), basics.CoverFull)
					rm.Ren().BlendVline(x-dy, y-dx+1, y+dx-1, rm.GetFillColor(), basics.CoverFull)
				}
				dy++
				dx += flip
				flip ^= 1
			}
			rm.SolidRectangle(x+r3+1, y+r3+1, x-r3-1, y-r3-1)
		} else {
			rm.Ren().BlendPixel(x, y, rm.GetFillColor(), basics.CoverFull)
		}
	}
}

// Cross draws a simple cross (plus sign) marker.
func (rm *RendererMarkers[BR, C]) Cross(x, y, r int) {
	if rm.Visible(x, y, r) {
		if r > 0 {
			rm.Ren().BlendVline(x, y-r, y+r, rm.GetLineColor(), basics.CoverFull)
			rm.Ren().BlendHline(x-r, y, x+r, rm.GetLineColor(), basics.CoverFull)
		} else {
			rm.Ren().BlendPixel(x, y, rm.GetFillColor(), basics.CoverFull)
		}
	}
}

// X draws an X-shaped marker.
func (rm *RendererMarkers[BR, C]) X(x, y, r int) {
	if rm.Visible(x, y, r) {
		if r > 0 {
			dy := -r * 7 / 10
			for dy < 0 {
				rm.Ren().BlendPixel(x+dy, y+dy, rm.GetLineColor(), basics.CoverFull)
				rm.Ren().BlendPixel(x-dy, y+dy, rm.GetLineColor(), basics.CoverFull)
				rm.Ren().BlendPixel(x+dy, y-dy, rm.GetLineColor(), basics.CoverFull)
				rm.Ren().BlendPixel(x-dy, y-dy, rm.GetLineColor(), basics.CoverFull)
				dy++
			}
		}
		rm.Ren().BlendPixel(x, y, rm.GetFillColor(), basics.CoverFull)
	}
}

// Dash draws a horizontal dash marker.
func (rm *RendererMarkers[BR, C]) Dash(x, y, r int) {
	if rm.Visible(x, y, r) {
		if r > 0 {
			rm.Ren().BlendHline(x-r, y, x+r, rm.GetLineColor(), basics.CoverFull)
		} else {
			rm.Ren().BlendPixel(x, y, rm.GetFillColor(), basics.CoverFull)
		}
	}
}

// Dot draws a filled dot (solid ellipse) marker.
func (rm *RendererMarkers[BR, C]) Dot(x, y, r int) {
	if rm.Visible(x, y, r) {
		if r > 0 {
			rm.SolidEllipse(x, y, r, r)
		} else {
			rm.Ren().BlendPixel(x, y, rm.GetFillColor(), basics.CoverFull)
		}
	}
}

// Pixel draws a single pixel marker.
func (rm *RendererMarkers[BR, C]) Pixel(x, y, _ int) {
	rm.Ren().BlendPixel(x, y, rm.GetFillColor(), basics.CoverFull)
}

// Marker draws a single marker of the specified type at (x, y) with radius r.
func (rm *RendererMarkers[BR, C]) Marker(x, y, r int, markerType MarkerType) {
	switch markerType {
	case MarkerSquare:
		rm.Square(x, y, r)
	case MarkerDiamond:
		rm.Diamond(x, y, r)
	case MarkerCircle:
		rm.Circle(x, y, r)
	case MarkerCrossedCircle:
		rm.CrossedCircle(x, y, r)
	case MarkerSemiEllipseLeft:
		rm.SemiEllipseLeft(x, y, r)
	case MarkerSemiEllipseRight:
		rm.SemiEllipseRight(x, y, r)
	case MarkerSemiEllipseUp:
		rm.SemiEllipseUp(x, y, r)
	case MarkerSemiEllipseDown:
		rm.SemiEllipseDown(x, y, r)
	case MarkerTriangleLeft:
		rm.TriangleLeft(x, y, r)
	case MarkerTriangleRight:
		rm.TriangleRight(x, y, r)
	case MarkerTriangleUp:
		rm.TriangleUp(x, y, r)
	case MarkerTriangleDown:
		rm.TriangleDown(x, y, r)
	case MarkerFourRays:
		rm.FourRays(x, y, r)
	case MarkerCross:
		rm.Cross(x, y, r)
	case MarkerX:
		rm.X(x, y, r)
	case MarkerDash:
		rm.Dash(x, y, r)
	case MarkerDot:
		rm.Dot(x, y, r)
	case MarkerPixel:
		rm.Pixel(x, y, r)
	}
}

// Markers draws multiple markers of the same type with the same radius.
// Takes slices of x coordinates, y coordinates, a common radius, and marker type.
func (rm *RendererMarkers[BR, C]) Markers(x, y []int, r int, markerType MarkerType) {
	if len(x) != len(y) || len(x) == 0 {
		return
	}

	if r == 0 {
		for i := 0; i < len(x); i++ {
			rm.Ren().BlendPixel(x[i], y[i], rm.GetFillColor(), basics.CoverFull)
		}
		return
	}

	for i := 0; i < len(x); i++ {
		rm.Marker(x[i], y[i], r, markerType)
	}
}

// MarkersVarRadius draws multiple markers of the same type with varying radii.
// Takes slices of x coordinates, y coordinates, radii, and marker type.
func (rm *RendererMarkers[BR, C]) MarkersVarRadius(x, y, r []int, markerType MarkerType) {
	if len(x) != len(y) || len(x) != len(r) || len(x) == 0 {
		return
	}

	for i := 0; i < len(x); i++ {
		rm.Marker(x[i], y[i], r[i], markerType)
	}
}

// MarkersVarColor draws multiple markers with varying fill colors.
// Takes slices of x coordinates, y coordinates, radii, fill colors, and marker type.
func (rm *RendererMarkers[BR, C]) MarkersVarColor(x, y, r []int, fillColors []C, markerType MarkerType) {
	if len(x) != len(y) || len(x) != len(r) || len(x) != len(fillColors) || len(x) == 0 {
		return
	}

	for i := 0; i < len(x); i++ {
		rm.FillColor(fillColors[i])
		rm.Marker(x[i], y[i], r[i], markerType)
	}
}

// MarkersVarColorAndLine draws multiple markers with varying fill and line colors.
// Takes slices of x coordinates, y coordinates, radii, fill colors, line colors, and marker type.
func (rm *RendererMarkers[BR, C]) MarkersVarColorAndLine(x, y, r []int, fillColors, lineColors []C, markerType MarkerType) {
	if len(x) != len(y) || len(x) != len(r) || len(x) != len(fillColors) || len(x) != len(lineColors) || len(x) == 0 {
		return
	}

	for i := 0; i < len(x); i++ {
		rm.FillColor(fillColors[i])
		rm.LineColor(lineColors[i])
		rm.Marker(x[i], y[i], r[i], markerType)
	}
}
