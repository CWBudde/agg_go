package benchmark

import (
	"runtime"
	"testing"

	"github.com/MeKo-Christian/agg_go/internal/agg2d"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
)

func BenchmarkLionFull(b *testing.B) {
	benchLion(b, "800x600", 800, 600)
}

func BenchmarkLionSmall(b *testing.B) {
	benchLion(b, "400x300", 400, 300)
}

func benchLion(b *testing.B, name string, width, height int) {
	b.Helper()
	b.Run(name, func(b *testing.B) {
		stride := width * 4
		buffer := make([]uint8, height*stride)

		ctx := agg2d.NewAgg2D()
		ctx.Attach(buffer, width, height, stride)

		// Pre-parse lion data outside the benchmark loop
		lionPaths := liondemo.Parse()

		b.ReportAllocs()
		b.SetBytes(int64(len(buffer)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			ctx.ClearAll(benchWhite)
			renderLion(ctx, lionPaths, 1.2, 250.0, 100.0)
		}

		runtime.KeepAlive(buffer)
	})
}

func renderLion(ctx *agg2d.Agg2D, paths []liondemo.Path, scale, offsetX, offsetY float64) {
	for _, lp := range paths {
		ctx.FillColor(agg2d.Color{lp.Color.R, lp.Color.G, lp.Color.B, lp.Color.A})
		ctx.NoLine()
		ctx.ResetPath()
		lp.Path.Rewind(0)
		for {
			x, y, cmd := lp.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			tx, ty := x*scale+offsetX, y*scale+offsetY
			if basics.IsMoveTo(basics.PathCommand(cmd)) {
				ctx.MoveTo(tx, ty)
			} else if basics.IsLineTo(basics.PathCommand(cmd)) {
				ctx.LineTo(tx, ty)
			}
		}
		ctx.ClosePolygon()
		ctx.DrawPath(agg2d.FillOnly)
	}
}
