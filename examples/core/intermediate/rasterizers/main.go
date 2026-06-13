// Port of AGG C++ rasterizers.cpp.
//
// This standalone version renders the default frame to a PNG via demorunner.
// Widget controls are represented by fixed defaults (gamma=0.5, alpha=1.0).
package main

import (
	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/examples/shared/lowlevelrunner"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	ctrlpkg "github.com/cwbudde/agg_go/internal/ctrl"
	"github.com/cwbudde/agg_go/internal/ctrl/checkbox"
	"github.com/cwbudde/agg_go/internal/ctrl/slider"
	"github.com/cwbudde/agg_go/internal/gamma"
	"github.com/cwbudde/agg_go/internal/path"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
)

const (
	frameWidth  = 500
	frameHeight = 330
)

// Type aliases for readability. C++ rasterizers.cpp uses AGG_BGR24, i.e. the
// plain (non-premultiplied) pixfmt_bgr24 for BOTH the scene and the controls
// (render_ctrl). We mirror that with a single plain linear RGBA base so that
// translucent control colours (the slider knobs use alpha 0.4/0.6) blend with
// straight alpha exactly like the C++ reference.
type (
	colorType   = color.RGBA8[color.Linear]
	pixFmtPlain = *pixfmt.PixFmtRGBA32[color.Linear]
	renBaseT    = *renderer.RendererBase[pixFmtPlain, colorType]
	rasType     = *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]
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

type controlPathAdapter struct {
	ctrl ctrlpkg.Ctrl[color.RGBA]
}

func (a *controlPathAdapter) Rewind(pathID uint32) {
	a.ctrl.Rewind(uint(pathID))
}

func (a *controlPathAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

func renderSolidPath(
	ras rasType,
	sl *scanline.ScanlineP8,
	renBase renBaseT,
	vs rasterizer.VertexSource,
	col colorType,
) {
	ras.Reset()
	ras.AddPath(vs, 0)

	if !ras.RewindScanlines() {
		return
	}

	sl.Reset(ras.MinX(), ras.MaxX())
	for ras.SweepScanline(sl) {
		y := sl.Y()
		for _, spanData := range sl.Spans() {
			if spanData.Len > 0 {
				renBase.BlendSolidHspan(int(spanData.X), y, int(spanData.Len), col, spanData.Covers)
			} else {
				renBase.BlendHline(int(spanData.X), y, int(spanData.X)-int(spanData.Len)-1, col, spanData.Covers[0])
			}
		}
	}
}

func renderAliasedPath(
	ras rasType,
	sl *scanline.ScanlineBin,
	renBase renBaseT,
	vs rasterizer.VertexSource,
	col colorType,
) {
	ras.Reset()
	ras.AddPath(vs, 0)

	if !ras.RewindScanlines() {
		return
	}

	sl.Reset(ras.MinX(), ras.MaxX())
	for ras.SweepScanline(sl) {
		renscan.RenderScanlineBinSolid(sl, renBase, col)
	}
}

func rgbaToRGBA8(c color.RGBA) color.RGBA8[color.Linear] {
	clamp := func(v float64) uint8 {
		if v <= 0 {
			return 0
		}
		if v >= 1 {
			return 255
		}
		return uint8(v*255 + 0.5)
	}
	return color.RGBA8[color.Linear]{
		R: clamp(c.R),
		G: clamp(c.G),
		B: clamp(c.B),
		A: clamp(c.A),
	}
}

func renderControl(
	ras rasType,
	sl *scanline.ScanlineP8,
	renBase renBaseT,
	ctrl ctrlpkg.Ctrl[color.RGBA],
) {
	adapter := &controlPathAdapter{ctrl: ctrl}
	for pathID := uint(0); pathID < ctrl.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(adapter, uint32(pathID))
		if !ras.RewindScanlines() {
			continue
		}
		sl.Reset(ras.MinX(), ras.MaxX())
		col := rgbaToRGBA8(ctrl.Color(pathID))
		renscan.RenderScanlinesAASolid(ras, sl, renBase, col)
	}
}

type demo struct{}

func (d *demo) Render(img *agg.Image) {
	imgData := img.Data
	rbuf := buffer.NewRenderingBufferU8WithData(imgData, frameWidth, frameHeight, img.Stride())

	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt(pf)
	renBase.Clear(colorType{R: 255, G: 255, B: 255, A: 255})

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineP8()
	slBin := scanline.NewScanlineBin()
	rasCtrl := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	slCtrl := scanline.NewScanlineP8()

	// Controls mirror the C++ members; their values drive the scene exactly as
	// in rasterizers.cpp (gamma feeds the rasterizer gamma, alpha the fill).
	gammaSlider := slider.NewSliderCtrl(140, 14, 280, 22, false)
	gammaSlider.SetRange(0.0, 1.0)
	gammaSlider.SetValue(0.5)
	gammaSlider.SetLabel("Gamma=%1.2f")

	alphaSlider := slider.NewSliderCtrl(290, 14, 490, 22, false)
	alphaSlider.SetRange(0.0, 1.0)
	alphaSlider.SetValue(1.0)
	alphaSlider.SetLabel("Alpha=%1.2f")

	testPerf := checkbox.NewDefaultCheckboxCtrl(140, 30, "Test Performance", false)
	testPerf.SetChecked(false)

	gammaValue := gammaSlider.Value()
	alphaValue := alphaSlider.Value()

	// Anti-aliased triangle: agg::rgba(0.7, 0.5, 0.1, m_alpha.value()).
	pathAA := path.NewPathStorageStl()
	pathAA.MoveTo(triX[0], triY[0])
	pathAA.LineTo(triX[1], triY[1])
	pathAA.LineTo(triX[2], triY[2])
	pathAA.ClosePolygon(0)
	ras.SetGamma(gamma.NewGammaPower(gammaValue * 2.0).Apply)
	renderSolidPath(
		ras,
		sl,
		renBase,
		&pathStorageAdapter{ps: pathAA},
		rgbaToRGBA8(color.NewRGBA(0.7, 0.5, 0.1, alphaValue)),
	)

	// Aliased triangle via threshold gamma: agg::rgba(0.1, 0.5, 0.7, m_alpha.value()).
	pathAliased := path.NewPathStorageStl()
	pathAliased.MoveTo(triX[0]-200, triY[0])
	pathAliased.LineTo(triX[1]-200, triY[1])
	pathAliased.LineTo(triX[2]-200, triY[2])
	pathAliased.ClosePolygon(0)
	ras.SetGamma(gamma.NewGammaThreshold(gammaValue).Apply)
	renderAliasedPath(
		ras,
		slBin,
		renBase,
		&pathStorageAdapter{ps: pathAliased},
		rgbaToRGBA8(color.NewRGBA(0.1, 0.5, 0.7, alphaValue)),
	)

	renderControl(rasCtrl, slCtrl, renBase, gammaSlider)
	renderControl(rasCtrl, slCtrl, renBase, alphaSlider)
	renderControl(rasCtrl, slCtrl, renBase, testPerf)
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:                 "Rasterizers",
		Width:                 frameWidth,
		Height:                frameHeight,
		FlipY:                 true,
		EncodeLinearRGBToSRGB: true,
	}, &demo{})
}
