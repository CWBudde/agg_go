// Port of AGG C++ conv_contour.cpp – "Contour Tool & Polygon Orientation".
package main

import (
	agg "agg_go"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

// --- State ---

var (
	contourWidth      = 0.0
	contourAutoDetect = true
	contourCloseMode  = 0 // 0=close, 1=close CW, 2=close CCW
)

// --- Drawing ---

func drawConvContourDemo() {
	a := ctx.GetAgg2D()
	a.ResetTransformations()

	// Build the "a" glyph path from the original conv_contour.cpp.
	ps := path.NewPathStorageStl()
	composePath(ps)

	// The C++ original used flip_y=true on its rendering buffer (330px tall window)
	// combined with scale(4) + translate(150,100). That is equivalent to the
	// affine matrix: x'=4*x+tx,  y'=-4*y+ty  (Y-axis negated).
	// We centre the glyph (path centroid ≈ (24, 22.5)) on the 800×600 canvas.
	// sx=4, shy=0, shx=0, sy=-4, tx=304, ty=390
	mtx := transform.NewTransAffineFromValues(4, 0, 0, -4, 304, 390)

	psAdapter := path.NewPathStorageStlVertexSourceAdapter(ps)
	trans := conv.NewConvTransform(psAdapter, mtx)
	curve := conv.NewConvCurve(trans)
	contour := conv.NewConvContour(curve)
	contour.Width(contourWidth)
	contour.AutoDetectOrientation(contourAutoDetect)

	// Feed the contour output vertices into the agg2d path and render.
	a.ResetPath()
	contour.Rewind(0)
	for {
		x, y, cmd := contour.Vertex()
		switch {
		case basics.IsStop(cmd):
			goto done
		case basics.IsMoveTo(cmd):
			a.MoveTo(x, y)
		case basics.IsEndPoly(cmd):
			if basics.IsClose(uint32(cmd)) {
				a.ClosePolygon()
			}
		case basics.IsVertex(cmd):
			a.LineTo(x, y)
		}
	}
done:
	a.FillColor(agg.Black)
	a.NoLine()
	a.DrawPath(agg.FillOnly)
}

// composePath builds the "a" glyph path from the original C++ example.
func composePath(ps *path.PathStorageStl) {
	var flag basics.PathFlag
	switch contourCloseMode {
	case 1:
		flag = basics.PathFlagsCW
	case 2:
		flag = basics.PathFlagsCCW
	default:
		flag = basics.PathFlagsNone
	}

	ps.RemoveAll()
	ps.MoveTo(28.47, 6.45)
	ps.Curve3(21.58, 1.12, 19.82, 0.29)
	ps.Curve3(17.19, -0.93, 14.21, -0.93)
	ps.Curve3(9.57, -0.93, 6.57, 2.25)
	ps.Curve3(3.56, 5.42, 3.56, 10.60)
	ps.Curve3(3.56, 13.87, 5.03, 16.26)
	ps.Curve3(7.03, 19.58, 11.99, 22.51)
	ps.Curve3(16.94, 25.44, 28.47, 29.64)
	ps.LineTo(28.47, 31.40)
	ps.Curve3(28.47, 38.09, 26.34, 40.58)
	ps.Curve3(24.22, 43.07, 20.17, 43.07)
	ps.Curve3(17.09, 43.07, 15.28, 41.41)
	ps.Curve3(13.43, 39.75, 13.43, 37.60)
	ps.LineTo(13.53, 34.77)
	ps.Curve3(13.53, 32.52, 12.38, 31.30)
	ps.Curve3(11.23, 30.08, 9.38, 30.08)
	ps.Curve3(7.57, 30.08, 6.42, 31.35)
	ps.Curve3(5.27, 32.62, 5.27, 34.81)
	ps.Curve3(5.27, 39.01, 9.57, 42.53)
	ps.Curve3(13.87, 46.04, 21.63, 46.04)
	ps.Curve3(27.59, 46.04, 31.40, 44.04)
	ps.Curve3(34.28, 42.53, 35.64, 39.31)
	ps.Curve3(36.52, 37.21, 36.52, 30.71)
	ps.LineTo(36.52, 15.53)
	ps.Curve3(36.52, 9.13, 36.77, 7.69)
	ps.Curve3(37.01, 6.25, 37.57, 5.76)
	ps.Curve3(38.13, 5.27, 38.87, 5.27)
	ps.Curve3(39.65, 5.27, 40.23, 5.62)
	ps.Curve3(41.26, 6.25, 44.19, 9.18)
	ps.LineTo(44.19, 6.45)
	ps.Curve3(38.72, -0.88, 33.74, -0.88)
	ps.Curve3(31.35, -0.88, 29.93, 0.78)
	ps.Curve3(28.52, 2.44, 28.47, 6.45)
	ps.ClosePolygon(flag)

	ps.MoveTo(28.47, 9.62)
	ps.LineTo(28.47, 26.66)
	ps.Curve3(21.09, 23.73, 18.95, 22.51)
	ps.Curve3(15.09, 20.36, 13.43, 18.02)
	ps.Curve3(11.77, 15.67, 11.77, 12.89)
	ps.Curve3(11.77, 9.38, 13.87, 7.06)
	ps.Curve3(15.97, 4.74, 18.70, 4.74)
	ps.Curve3(22.41, 4.74, 28.47, 9.62)
	ps.ClosePolygon(flag)
}
