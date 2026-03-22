package integration

import (
	"testing"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

type singlePathColorStorage struct {
	color color.RGBA8[color.Linear]
}

func (s singlePathColorStorage) GetColor(index int) color.RGBA8[color.Linear] {
	return s.color
}

type singlePathIDStorage struct{}

func (singlePathIDStorage) GetPathID(index int) uint32 {
	return 0
}

func TestRenderAllPathsPreservesCloseFlag(t *testing.T) {
	const (
		w = 24
		h = 24
	)

	pathStorage := path.NewPathStorageStl()
	pathStorage.MoveTo(4, 4)
	pathStorage.LineTo(20, 4)
	pathStorage.LineTo(12, 20)
	pathStorage.ClosePolygon(basics.PathFlagsNone)

	correct := renderSinglePathViaRenderAllPaths(pathStorage, w, h)
	legacy := renderSinglePathViaLegacyLoop(pathStorage, w, h)

	correctR, correctG, correctB := px4(correct, w, 12, 10)
	legacyR, legacyG, legacyB := px4(legacy, w, 12, 10)

	if correctR == 255 && correctG == 255 && correctB == 255 {
		t.Fatalf("RenderAllPaths pixel remained background; expected filled triangle at (12,10)")
	}

	if legacyR != 255 || legacyG != 255 || legacyB != 255 {
		t.Fatalf("legacy loop unexpectedly filled pixel (12,10): got (%d,%d,%d)", legacyR, legacyG, legacyB)
	}
}

func renderSinglePathViaRenderAllPaths(pathStorage *path.PathStorageStl, w, h int) []uint8 {
	pixels := make([]uint8, w*h*4)
	rbuf := buffer.NewRenderingBufferU8WithData(pixels, w, h, w*4)
	pixf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt(pixf)
	renBase.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	ras.AutoClose(false)

	sl := scanline.NewScanlineU8()
	renSolid := renscan.NewRendererScanlineAASolidWithRenderer(renBase)

	pathVS := path.NewPathStorageStlVertexSourceAdapter(pathStorage)
	transVS := conv.NewConvTransform(pathVS, transform.NewTransAffine())
	rasVS := conv.NewRasterizerVertexSourceAdapter(transVS)

	renscan.RenderAllPaths(
		ras,
		sl,
		renSolid,
		rasVS,
		singlePathColorStorage{color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255}},
		singlePathIDStorage{},
		1,
	)

	return pixels
}

func renderSinglePathViaLegacyLoop(pathStorage *path.PathStorageStl, w, h int) []uint8 {
	pixels := make([]uint8, w*h*4)
	rbuf := buffer.NewRenderingBufferU8WithData(pixels, w, h, w*4)
	pixf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt(pixf)
	renBase.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	ras.AutoClose(false)

	sl := scanline.NewScanlineU8()
	pathStorage.Rewind(0)
	for {
		x, y, cmd := pathStorage.NextVertex()
		pathCmd := basics.PathCommand(cmd)
		if basics.IsStop(pathCmd) {
			break
		}
		if basics.IsMoveTo(pathCmd) {
			ras.AddVertex(x, y, uint32(basics.PathCmdMoveTo))
		} else if basics.IsLineTo(pathCmd) {
			ras.AddVertex(x, y, uint32(basics.PathCmdLineTo))
		}
	}

	renscan.RenderScanlinesAASolid(ras, sl, renBase, color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})
	return pixels
}
