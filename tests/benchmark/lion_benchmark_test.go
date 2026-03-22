package benchmark

import (
	"runtime"
	"testing"

	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/transform"
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
		pixels := make([]uint8, height*stride)
		rbuf := buffer.NewRenderingBufferU8WithData(pixels, width, height, stride)
		pixf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
		renBase := renderer.NewRendererBaseWithPixfmt(pixf)
		ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
			rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
		)
		sl := scanline.NewScanlineU8()
		renSolid := renscan.NewRendererScanlineAASolidWithRenderer(renBase)

		// Pre-parse lion data outside the benchmark loop
		ld := liondemo.Parse()
		pathVS := path.NewPathStorageStlVertexSourceAdapter(ld.Path)
		mtx := transform.NewTransAffine()
		mtx.Multiply(transform.NewTransAffineScaling(1.2))
		mtx.Multiply(transform.NewTransAffineTranslation(250.0, 100.0))
		transVS := conv.NewConvTransform(pathVS, mtx)
		rasVS := conv.NewRasterizerVertexSourceAdapter(transVS)

		b.ReportAllocs()
		b.SetBytes(int64(len(pixels)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			renBase.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})
			renscan.RenderAllPaths(ras, sl, renSolid, rasVS, &ld, &ld, ld.NPaths)
		}

		runtime.KeepAlive(pixels)
	})
}
