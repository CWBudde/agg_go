package rasterizer_test

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/conv"
	aggpath "github.com/cwbudde/agg_go/internal/path"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/scanline"
)

func TestRasterizerScanlineAACircleMatchesMatplotlibAGG(t *testing.T) {
	path := aggpath.NewPathStorage()
	addMatplotlibCircleMarkerPath(path, 100.25, 100.75, 21.269446210866192)

	curve := conv.NewConvCurve(aggpath.NewPathStorageVertexSourceAdapter(path))
	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	ras.ClipBox(0, 0, 200, 200)
	ras.AddPath(rasterizerVertexSource{src: curve}, 0)

	if !ras.RewindScanlines() {
		t.Fatal("expected rasterized scanlines")
	}

	sl := scanline.NewScanlineU8()
	sl.Reset(ras.MinX(), ras.MaxX())

	var got []string
	for ras.SweepScanline(sl) {
		for _, span := range sl.Begin() {
			got = append(got, formatSpan(sl.Y(), span))
		}
	}

	want := []string{
		"y=90 x=96 len=9 45 131 183 220 224 207 157 94 10",
		"y=91 x=94 len=13 47 193 255 255 255 255 255 255 255 255 239 124 5",
		"y=92 x=92 len=16 1 118 251 255 255 255 255 255 255 255 255 255 255 255 207 33",
		"y=93 x=92 len=17 134 255 255 255 255 255 255 255 255 255 255 255 255 255 255 228 34",
		"y=94 x=91 len=19 84 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 207 5",
		"y=95 x=90 len=20 19 234 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 124",
		"y=96 x=90 len=21 120 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 238 9",
		"y=97 x=89 len=22 1 222 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 94",
		"y=98 x=89 len=22 29 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 156",
		"y=99 x=89 len=22 79 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 206",
		"y=100 x=89 len=22 96 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 224",
		"y=101 x=89 len=22 92 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 220",
		"y=102 x=89 len=22 55 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 182",
		"y=103 x=89 len=22 9 251 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 130",
		"y=104 x=90 len=21 173 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 44",
		"y=105 x=90 len=20 65 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 193",
		"y=106 x=91 len=19 169 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 251 47",
		"y=107 x=91 len=18 21 225 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 117",
		"y=108 x=92 len=17 36 225 255 255 255 255 255 255 255 255 255 255 255 255 255 133 1",
		"y=109 x=93 len=14 21 169 255 255 255 255 255 255 255 255 255 255 234 84",
		"y=110 x=95 len=11 65 172 251 255 255 255 255 255 222 119 19",
		"y=111 x=97 len=7 8 54 92 96 78 28 1",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("scanline coverage mismatch\nwant:\n%s\n\ngot:\n%s", strings.Join(want, "\n"), strings.Join(got, "\n"))
	}
}

type rasterizerVertexSource struct {
	src *conv.ConvCurve
}

func (v rasterizerVertexSource) Rewind(pathID uint32) {
	v.src.Rewind(uint(pathID))
}

func (v rasterizerVertexSource) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := v.src.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

func formatSpan(y int, span scanline.Span) string {
	var b strings.Builder
	fmt.Fprintf(&b, "y=%d x=%d len=%d", y, span.X, span.Len)
	for i := 0; i < int(span.Len); i++ {
		fmt.Fprintf(&b, " %d", span.Covers[i])
	}
	return b.String()
}

func addMatplotlibCircleMarkerPath(path *aggpath.PathStorage, cx, cy, scale float64) {
	const segments = 8
	const control = 0.2652031
	r := 0.5 * scale
	delta := 2 * math.Pi / segments
	point := func(theta float64) (float64, float64) {
		return cx + r*math.Cos(theta), cy + r*math.Sin(theta)
	}
	tangent := func(theta float64) (float64, float64) {
		return -math.Sin(theta), math.Cos(theta)
	}

	theta0 := -math.Pi / 2
	x0, y0 := point(theta0)
	path.MoveTo(x0, y0)
	for i := 0; i < segments; i++ {
		theta1 := theta0 + delta
		x1, y1 := point(theta1)
		tx0, ty0 := tangent(theta0)
		tx1, ty1 := tangent(theta1)
		path.Curve4(
			x0+control*r*tx0,
			y0+control*r*ty0,
			x1-control*r*tx1,
			y1-control*r*ty1,
			x1,
			y1,
		)
		x0, y0 = x1, y1
		theta0 = theta1
	}
	path.ClosePolygon(basics.PathFlagsNone)
}
