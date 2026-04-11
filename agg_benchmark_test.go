package agg_test

import (
	"testing"

	"github.com/cwbudde/agg_go"
)

func BenchmarkAAFill1MPixels(b *testing.B) {
	width := 1200
	height := 1200

	ctx := agg.NewContext(width, height)
	c := agg.NewColor(255, 0, 0, 128) // Red with 50% opacity

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx.Clear(agg.White)

		ctx.SetColor(c)
		ctx.BeginPath()
		ctx.MoveTo(100, 100)
		ctx.LineTo(1100, 100)
		ctx.LineTo(1100, 1100)
		ctx.LineTo(100, 1100)
		ctx.ClosePath()
		ctx.Fill()
	}
}
