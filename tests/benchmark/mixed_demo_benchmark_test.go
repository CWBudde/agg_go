package benchmark

import (
	"runtime"
	"testing"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/demo/benchmarkmix"
)

func BenchmarkMixedDemoScene(b *testing.B) {
	for _, tc := range []struct {
		name          string
		width, height int
	}{
		{name: "800x600", width: 800, height: 600},
		{name: "1280x960", width: 1280, height: 960},
	} {
		b.Run(tc.name, func(b *testing.B) {
			ctx := agg.NewContext(tc.width, tc.height)
			scene := benchmarkmix.New(tc.width, tc.height)

			b.ReportAllocs()
			b.SetBytes(int64(tc.width * tc.height * 4))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := scene.Draw(ctx); err != nil {
					b.Fatalf("scene draw failed: %v", err)
				}
			}

			runtime.KeepAlive(ctx)
			runtime.KeepAlive(scene)
		})
	}
}

func BenchmarkMixedDemoComponents(b *testing.B) {
	const (
		width  = 1280
		height = 960
	)

	scene := benchmarkmix.New(width, height)
	ctx := agg.NewContext(width, height)

	b.Run("BlendColorTile", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			scene.DrawBlendColorTile()
		}
	})

	b.Run("PatternResampleTile", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			scene.DrawPatternResampleTile()
		}
	})

	b.Run("PolygonClipTile", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			scene.DrawGPCTile()
		}
	})

	b.Run("LinePatternsTile", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			scene.DrawLinePatternsTile()
		}
	})

	b.Run("GraphTile", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			scene.DrawGraphTile()
		}
	})

	b.Run("FilterGraphTile", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			scene.DrawFilterGraphTile()
		}
	})

	b.Run("OverlayAndComposite", func(b *testing.B) {
		scene.DrawBlendColorTile()
		scene.DrawPatternResampleTile()
		scene.DrawGPCTile()
		scene.DrawLinePatternsTile()
		scene.DrawGraphTile()
		scene.DrawFilterGraphTile()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			scene.DrawOverlay(ctx)
		}
	})
}
