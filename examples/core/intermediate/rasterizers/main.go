// Port of AGG C++ rasterizers.cpp.
//
// This standalone version renders the default frame to rasterizers_demo.ppm.
// Widget controls are represented by fixed defaults (gamma=0.5, alpha=1.0).
package main

import (
	"fmt"
	"os"

	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/gamma"
	"agg_go/internal/order"
	"agg_go/internal/path"
	"agg_go/internal/pixfmt"
	"agg_go/internal/pixfmt/blender"
	"agg_go/internal/rasterizer"
	"agg_go/internal/renderer"
	"agg_go/internal/scanline"
)

const (
	frameWidth  = 500
	frameHeight = 330
)

var (
	triX = [3]float64{100 + 120, 369 + 120, 143 + 120}
	triY = [3]float64{60, 170, 310}
)

type pathStorageAdapter struct {
	ps *path.PathStorageStl
}

func (a *pathStorageAdapter) Rewind(pathID uint32) {
	a.ps.Rewind(uint(pathID))
}

func (a *pathStorageAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ps.NextVertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

func renderSolidPath(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	sl *scanline.ScanlineU8,
	renBase *renderer.RendererBase[*pixfmt.PixFmtAlphaBlendRGBA[color.Linear, blender.BlenderRGBA8Pre[color.Linear, order.RGBA]], color.RGBA8[color.Linear]],
	vs rasterizer.VertexSource,
	col color.RGBA8[color.Linear],
) {
	ras.Reset()
	ras.AddPath(vs, 0)

	if !ras.RewindScanlines() {
		return
	}

	sl.Reset(ras.MinX(), ras.MaxX())
	for ras.SweepScanline(&rasScanlineAdapter{sl: sl}) {
		y := sl.Y()
		for _, spanData := range sl.Spans() {
			if spanData.Len > 0 {
				renBase.BlendSolidHspan(int(spanData.X), y, int(spanData.Len), col, spanData.Covers)
			}
		}
	}
}

type rasScanlineAdapter struct {
	sl *scanline.ScanlineU8
}

func (a *rasScanlineAdapter) ResetSpans()                 { a.sl.ResetSpans() }
func (a *rasScanlineAdapter) AddCell(x int, cover uint32) { a.sl.AddCell(x, uint(cover)) }
func (a *rasScanlineAdapter) AddSpan(x, length int, cover uint32) {
	a.sl.AddSpan(x, length, uint(cover))
}
func (a *rasScanlineAdapter) Finalize(y int) { a.sl.Finalize(y) }
func (a *rasScanlineAdapter) NumSpans() int  { return a.sl.NumSpans() }

func savePPM(filename string, imgData []uint8, width, height int) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "P6\n%d %d\n255\n", width, height); err != nil {
		return err
	}

	for i := 0; i < len(imgData); i += 4 {
		if _, err := f.Write([]byte{imgData[i], imgData[i+1], imgData[i+2]}); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	imgData := make([]uint8, frameWidth*frameHeight*4)
	rbuf := buffer.NewRenderingBufferU8WithData(imgData, frameWidth, frameHeight, frameWidth*4)

	pf := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtAlphaBlendRGBA[color.Linear, blender.BlenderRGBA8Pre[color.Linear, order.RGBA]], color.RGBA8[color.Linear]](pf)
	renBase.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	// Anti-aliased triangle (same defaults as C++ sample).
	pathAA := path.NewPathStorageStl()
	pathAA.MoveTo(triX[0], triY[0])
	pathAA.LineTo(triX[1], triY[1])
	pathAA.LineTo(triX[2], triY[2])
	pathAA.ClosePolygon(0)
	ras.SetGamma(gamma.NewGammaPower(0.5 * 2.0).Apply)
	renderSolidPath(
		ras,
		sl,
		renBase,
		&pathStorageAdapter{ps: pathAA},
		color.RGBA8[color.Linear]{R: 178, G: 127, B: 25, A: 255},
	)

	// Aliased triangle via threshold gamma.
	pathAliased := path.NewPathStorageStl()
	pathAliased.MoveTo(triX[0]-200, triY[0])
	pathAliased.LineTo(triX[1]-200, triY[1])
	pathAliased.LineTo(triX[2]-200, triY[2])
	pathAliased.ClosePolygon(0)
	ras.SetGamma(gamma.NewGammaThreshold(0.5).Apply)
	renderSolidPath(
		ras,
		sl,
		renBase,
		&pathStorageAdapter{ps: pathAliased},
		color.RGBA8[color.Linear]{R: 25, G: 127, B: 178, A: 255},
	)

	if err := savePPM("rasterizers_demo.ppm", imgData, frameWidth, frameHeight); err != nil {
		panic(err)
	}
	fmt.Println("rasterizers_demo.ppm")
}
