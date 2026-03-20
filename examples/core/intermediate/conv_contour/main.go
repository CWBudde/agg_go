// Port of AGG C++ conv_contour.cpp – contour stroke generation.
//
// Demonstrates shrinking/expanding a path outline via conv_contour.
// The "a" glyph from the original demo is rendered at multiple contour
// widths to show the effect: negative = shrink, positive = expand.
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

const (
	width  = 440
	height = 330
)

// buildGlyphPath constructs the "a" glyph from the original AGG example.
func buildGlyphPath() *path.PathStorageStl {
	ps := path.NewPathStorageStl()

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
	ps.ClosePolygon(basics.PathFlagsNone)

	ps.MoveTo(28.47, 9.62)
	ps.LineTo(28.47, 26.66)
	ps.Curve3(21.09, 23.73, 18.95, 22.51)
	ps.Curve3(15.09, 20.36, 13.43, 18.02)
	ps.Curve3(11.77, 15.67, 11.77, 12.89)
	ps.Curve3(11.77, 9.38, 13.87, 7.06)
	ps.Curve3(15.97, 4.74, 18.70, 4.74)
	ps.Curve3(22.41, 4.74, 28.47, 9.62)
	ps.ClosePolygon(basics.PathFlagsNone)

	return ps
}

func renderContour(a *agg.Agg2D, ps *path.PathStorageStl, mtx *transform.TransAffine, w float64, fillColor agg.Color) {
	adapter := path.NewPathStorageStlVertexSourceAdapter(ps)
	trans := conv.NewConvTransform(adapter, mtx)
	curve := conv.NewConvCurve(trans)
	contour := conv.NewConvContour(curve)
	contour.Width(w)
	contour.AutoDetectOrientation(true)

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
	a.FillColor(fillColor)
	a.NoLine()
	a.DrawPath(agg.FillOnly)
}

type demo struct{}

func (d *demo) Render(img *agg.Image) {
	ctx := agg.NewContextForImage(img)
	ctx.Clear(agg.White)

	a := ctx.GetAgg2D()
	a.ResetTransformations()

	ps := buildGlyphPath()

	// Render at 4 positions with different contour widths.
	variants := []struct {
		tx, ty       float64
		scale        float64
		contourWidth float64
		color        agg.Color
	}{
		{100, 350, 4, 0, agg.Black},                       // original (no contour)
		{300, 350, 4, -3.0, agg.NewColor(0, 0, 150, 255)}, // shrunk
		{500, 350, 4, 3.0, agg.NewColor(0, 150, 0, 255)},  // expanded
		{700, 350, 4, 8.0, agg.NewColor(150, 0, 0, 200)},  // more expanded
	}

	for _, v := range variants {
		// Transform: scale + flip-Y + translate.
		mtx := transform.NewTransAffineFromValues(v.scale, 0, 0, -v.scale, v.tx, v.ty)
		renderContour(a, ps, mtx, v.contourWidth, v.color)
	}
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{Title: "Conv Contour", Width: width, Height: height}, &demo{})
}
