package benchmark

import (
	"runtime"
	"testing"

	agg "github.com/MeKo-Christian/agg_go"
	blendcolordemo "github.com/MeKo-Christian/agg_go/internal/demo/blendcolor"
)

func BenchmarkBlendColorLUT(b *testing.B) {
	benchBlendColorScene(b, "800x600", 800, 600, blendcolordemo.Config{
		Method: 1,
		Radius: 15,
	})
}

func benchBlendColorScene(b *testing.B, name string, width, height int, cfg blendcolordemo.Config) {
	b.Helper()
	b.Run(name, func(b *testing.B) {
		ctx := agg.NewContext(width, height)

		b.ReportAllocs()
		b.SetBytes(int64(width * height * 4))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			blendcolordemo.Draw(ctx, &cfg)
		}

		runtime.KeepAlive(ctx)
	})
}
