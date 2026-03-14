package outline

import (
	"testing"

	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/primitives"
)

func pixelHighResDiv(rows [][]color.RGBA, p *color.RGBA, x, y int) {
	if len(rows) == 0 || p == nil {
		*p = color.NewRGBA(0, 0, 0, 0)
		return
	}

	xLr := x >> primitives.LineSubpixelShift
	yLr := y >> primitives.LineSubpixelShift

	x &= primitives.LineSubpixelMask
	y &= primitives.LineSubpixelMask

	if yLr < 0 || yLr >= len(rows) || xLr < 0 || xLr >= len(rows[yLr]) {
		*p = color.NewRGBA(0, 0, 0, 0)
		return
	}

	var r, g, b, a int

	weight := (primitives.LineSubpixelScale - x) * (primitives.LineSubpixelScale - y)
	ptr := rows[yLr][xLr]
	r += weight * rgbaComponent(ptr.R)
	g += weight * rgbaComponent(ptr.G)
	b += weight * rgbaComponent(ptr.B)
	a += weight * rgbaComponent(ptr.A)

	if xLr+1 < len(rows[yLr]) {
		weight = x * (primitives.LineSubpixelScale - y)
		ptr = rows[yLr][xLr+1]
		r += weight * rgbaComponent(ptr.R)
		g += weight * rgbaComponent(ptr.G)
		b += weight * rgbaComponent(ptr.B)
		a += weight * rgbaComponent(ptr.A)
	}

	if yLr+1 < len(rows) && xLr < len(rows[yLr+1]) {
		weight = (primitives.LineSubpixelScale - x) * y
		ptr = rows[yLr+1][xLr]
		r += weight * rgbaComponent(ptr.R)
		g += weight * rgbaComponent(ptr.G)
		b += weight * rgbaComponent(ptr.B)
		a += weight * rgbaComponent(ptr.A)
	}

	if yLr+1 < len(rows) && xLr+1 < len(rows[yLr+1]) {
		weight = x * y
		ptr = rows[yLr+1][xLr+1]
		r += weight * rgbaComponent(ptr.R)
		g += weight * rgbaComponent(ptr.G)
		b += weight * rgbaComponent(ptr.B)
		a += weight * rgbaComponent(ptr.A)
	}

	shift := primitives.LineSubpixelShift * 2
	*p = color.NewRGBA(
		float64(r>>shift)/255.0,
		float64(g>>shift)/255.0,
		float64(b>>shift)/255.0,
		float64(a>>shift)/255.0,
	)
}

func makePatternFilterRows(width, height int) [][]color.RGBA {
	rows := make([][]color.RGBA, height)
	for y := range rows {
		rows[y] = make([]color.RGBA, width)
		for x := range rows[y] {
			rows[y][x] = color.NewRGBA(
				float64((x*17+y*3)&255)*rgba8ToFloat64,
				float64((x*5+y*11)&255)*rgba8ToFloat64,
				float64((x*13+y*7)&255)*rgba8ToFloat64,
				float64((x*19+y*23)&255)*rgba8ToFloat64,
			)
		}
	}
	return rows
}

func TestPatternFilterRGBAAdapterMatchesDivReference(t *testing.T) {
	rows := makePatternFilterRows(32, 16)
	filter := NewPatternFilterRGBAAdapter()
	var got, want color.RGBA

	for y := 0; y < len(rows)*primitives.LineSubpixelScale; y += 19 {
		for x := 0; x < len(rows[0])*primitives.LineSubpixelScale; x += 23 {
			filter.PixelHighRes(rows, &got, x, y)
			pixelHighResDiv(rows, &want, x, y)
			if rgbaComponent(got.R) != rgbaComponent(want.R) ||
				rgbaComponent(got.G) != rgbaComponent(want.G) ||
				rgbaComponent(got.B) != rgbaComponent(want.B) ||
				rgbaComponent(got.A) != rgbaComponent(want.A) {
				t.Fatalf("PixelHighRes mismatch at (%d,%d): got=%+v want=%+v", x, y, got, want)
			}
		}
	}
}

func BenchmarkPatternFilterRGBAAdapterPixelHighRes(b *testing.B) {
	rows := makePatternFilterRows(128, 32)
	filter := NewPatternFilterRGBAAdapter()
	coords := make([][2]int, 1024)
	for i := range coords {
		coords[i][0] = (i * 37) % ((len(rows[0]) - 1) << primitives.LineSubpixelShift)
		coords[i][1] = (i * 29) % ((len(rows) - 1) << primitives.LineSubpixelShift)
	}

	b.Run("mul_const", func(b *testing.B) {
		var out color.RGBA
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			xy := coords[i&1023]
			filter.PixelHighRes(rows, &out, xy[0], xy[1])
		}
		_ = out
	})

	b.Run("div_reference", func(b *testing.B) {
		var out color.RGBA
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			xy := coords[i&1023]
			pixelHighResDiv(rows, &out, xy[0], xy[1])
		}
		_ = out
	})
}
