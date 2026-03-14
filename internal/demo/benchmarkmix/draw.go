package benchmarkmix

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	blendcolordemo "github.com/MeKo-Christian/agg_go/internal/demo/blendcolor"
	"github.com/MeKo-Christian/agg_go/internal/demo/gpctest"
	"github.com/MeKo-Christian/agg_go/internal/demo/graphtest"
	"github.com/MeKo-Christian/agg_go/internal/demo/imageassets"
	"github.com/MeKo-Christian/agg_go/internal/demo/imagefltrgraph"
	"github.com/MeKo-Christian/agg_go/internal/demo/linepatterns"
	"github.com/MeKo-Christian/agg_go/internal/demo/patternresample"
)

type tile struct {
	img *agg.Image
	ctx *agg.Context
	x   float64
	y   float64
	w   int
	h   int
}

type Scene struct {
	width  int
	height int

	blendTile   tile
	patternTile tile
	gpcTile     tile
	lineTile    tile
	graphTile   tile
	filterTile  tile

	graph       *graphtest.Graph
	filterState imagefltrgraph.State
	aggImage    *agg.Image
	spheres     *agg.Image
	backdrop    *agg.Image
	overlay     *agg.Image
}

func New(width, height int) *Scene {
	scene := &Scene{
		width:  width,
		height: height,
		graph:  graphtest.NewGraph(220, 180),
	}
	scene.filterState = imagefltrgraph.DefaultState()
	scene.filterState.Enabled = [16]bool{}
	scene.filterState.Enabled[0] = true
	scene.filterState.Enabled[4] = true
	scene.filterState.Enabled[9] = true
	scene.filterState.Enabled[15] = true
	scene.filterState.Radius = 4.5

	scene.aggImage, _ = imageassets.Agg()
	scene.spheres, _ = imageassets.Spheres()
	scene.initTiles()
	scene.initLayers()

	return scene
}

func (s *Scene) initTiles() {
	const (
		margin = 18.0
		gap    = 14.0
		header = 118.0
	)

	tileW := int(math.Max(180, math.Floor((float64(s.width)-2*margin-2*gap)/3.0)))
	tileH := int(math.Max(160, math.Floor((float64(s.height)-header-2*margin-gap)/2.0)))

	s.blendTile = newTile(margin, header, tileW, tileH)
	s.patternTile = newTile(margin+float64(tileW)+gap, header, tileW, tileH)
	s.gpcTile = newTile(margin+2*float64(tileW)+2*gap, header, tileW, tileH)

	row2Y := header + float64(tileH) + gap
	s.lineTile = newTile(margin, row2Y, tileW, tileH)
	s.graphTile = newTile(margin+float64(tileW)+gap, row2Y, tileW, tileH)
	s.filterTile = newTile(margin+2*float64(tileW)+2*gap, row2Y, tileW, tileH)
}

func newTile(x, y float64, w, h int) tile {
	buf := make([]uint8, w*h*4)
	img := agg.NewImage(buf, w, h, w*4)
	return tile{
		img: img,
		ctx: agg.NewContextForImage(img),
		x:   x,
		y:   y,
		w:   w,
		h:   h,
	}
}

func (s *Scene) Draw(ctx *agg.Context) error {
	if ctx == nil {
		return nil
	}

	s.DrawBlendColorTile()
	s.DrawPatternResampleTile()
	s.DrawGPCTile()
	s.DrawLinePatternsTile()
	s.DrawGraphTile()
	s.DrawFilterGraphTile()

	s.drawBackdrop(ctx)
	return s.DrawOverlay(ctx)
}

func (s *Scene) DrawBlendColorTile() {
	cfg := blendcolordemo.Config{
		Method: 1,
		Radius: 14,
		Quad: [8]float64{
			float64(s.blendTile.w) * 0.16, float64(s.blendTile.h) * 0.18,
			float64(s.blendTile.w) * 0.87, float64(s.blendTile.h) * 0.10,
			float64(s.blendTile.w) * 0.76, float64(s.blendTile.h) * 0.85,
			float64(s.blendTile.w) * 0.10, float64(s.blendTile.h) * 0.76,
		},
	}
	blendcolordemo.Draw(s.blendTile.ctx, &cfg)
}

func (s *Scene) DrawPatternResampleTile() {
	cfg := patternresample.Config{
		Mode:  5,
		Gamma: 1.9,
		Blur:  1.35,
		Quad: [4][2]float64{
			{float64(s.patternTile.w) * 0.10, float64(s.patternTile.h) * 0.20},
			{float64(s.patternTile.w) * 0.92, float64(s.patternTile.h) * 0.08},
			{float64(s.patternTile.w) * 0.82, float64(s.patternTile.h) * 0.90},
			{float64(s.patternTile.w) * 0.06, float64(s.patternTile.h) * 0.84},
		},
	}
	patternresample.Draw(s.patternTile.ctx, cfg)
}

func (s *Scene) DrawGPCTile() {
	gpctest.Draw(s.gpcTile.ctx, gpctest.Config{
		Scene:     4,
		Operation: 2,
		CenterX:   float64(s.gpcTile.w) * 0.52,
		CenterY:   float64(s.gpcTile.h) * 0.52,
	})
}

func (s *Scene) DrawLinePatternsTile() {
	linepatterns.Draw(s.lineTile.img, 1.55, 3.0)
}

func (s *Scene) DrawGraphTile() {
	graphtest.Draw(s.graphTile.ctx, s.graph, graphtest.Config{
		Mode:        1,
		Width:       1.4,
		Translucent: true,
		DrawNodes:   true,
		DrawEdges:   true,
	})
}

func (s *Scene) DrawFilterGraphTile() {
	imagefltrgraph.Draw(s.filterTile.ctx, s.filterState)
}

func (s *Scene) DrawOverlay(ctx *agg.Context) error {
	if ctx == nil {
		return nil
	}

	a := ctx.GetAgg2D()
	a.ResetTransformations()
	a.ClipBox(0, 0, float64(s.width), float64(s.height))
	if err := a.CopyImageSimple(s.blendTile.img, s.blendTile.x, s.blendTile.y); err != nil {
		return err
	}
	if err := a.CopyImageSimple(s.patternTile.img, s.patternTile.x, s.patternTile.y); err != nil {
		return err
	}
	if err := a.CopyImageSimple(s.gpcTile.img, s.gpcTile.x, s.gpcTile.y); err != nil {
		return err
	}
	if err := a.CopyImageSimple(s.lineTile.img, s.lineTile.x, s.lineTile.y); err != nil {
		return err
	}
	if err := a.CopyImageSimple(s.graphTile.img, s.graphTile.x, s.graphTile.y); err != nil {
		return err
	}
	if err := a.CopyImageSimple(s.filterTile.img, s.filterTile.x, s.filterTile.y); err != nil {
		return err
	}
	if s.overlay != nil {
		if err := a.BlendImageSimple(s.overlay, 0, 0, 255); err != nil {
			return err
		}
	}
	return nil
}

func (s *Scene) drawBackdrop(ctx *agg.Context) {
	if s.backdrop != nil {
		ctx.GetAgg2D().CopyImageSimple(s.backdrop, 0, 0)
		return
	}
	ctx.Clear(agg.RGBA(0.985, 0.985, 0.97, 1.0))
	a := ctx.GetAgg2D()
	a.ResetTransformations()
	a.FillLinearGradient(0, 0, 0, float64(s.height),
		agg.RGBA(0.97, 0.975, 0.985, 1.0),
		agg.RGBA(0.91, 0.93, 0.95, 1.0), 1.0)
	a.NoLine()
	a.Rectangle(0, 0, float64(s.width), float64(s.height))
}

func (s *Scene) initLayers() {
	s.backdrop = newLayerImage(s.width, s.height)
	if s.backdrop != nil {
		s.renderBackdropLayer(agg.NewContextForImage(s.backdrop))
	}

	s.overlay = newLayerImage(s.width, s.height)
	if s.overlay != nil {
		s.renderOverlayLayer(agg.NewContextForImage(s.overlay))
	}
}

func newLayerImage(width, height int) *agg.Image {
	buf := make([]uint8, width*height*4)
	return agg.NewImage(buf, width, height, width*4)
}

func (s *Scene) renderBackdropLayer(ctx *agg.Context) {
	if ctx == nil {
		return
	}
	ctx.Clear(agg.RGBA(0.985, 0.985, 0.97, 1.0))
	a := ctx.GetAgg2D()
	a.ResetTransformations()
	a.FillLinearGradient(0, 0, 0, float64(s.height),
		agg.RGBA(0.97, 0.975, 0.985, 1.0),
		agg.RGBA(0.91, 0.93, 0.95, 1.0), 1.0)
	a.NoLine()
	a.Rectangle(0, 0, float64(s.width), float64(s.height))
}

func (s *Scene) renderOverlayLayer(ctx *agg.Context) {
	if ctx == nil {
		return
	}
	ctx.Clear(agg.RGBA(0, 0, 0, 0))
	if err := s.drawOverlayDecor(ctx); err != nil {
		return
	}
	s.drawTileBorder(ctx, s.blendTile)
	s.drawTileBorder(ctx, s.patternTile)
	s.drawTileBorder(ctx, s.gpcTile)
	s.drawTileBorder(ctx, s.lineTile)
	s.drawTileBorder(ctx, s.graphTile)
	s.drawTileBorder(ctx, s.filterTile)
}

func (s *Scene) drawOverlayDecor(ctx *agg.Context) error {
	a := ctx.GetAgg2D()
	a.ResetTransformations()
	a.ClipBox(0, 0, float64(s.width), float64(s.height))
	a.FillLinearGradient(0, 0, float64(s.width), 0,
		agg.RGBA(0.92, 0.96, 1.0, 0.90),
		agg.RGBA(0.97, 0.93, 0.90, 0.86), 1.0)
	headerBottom := 94.0
	a.NoLine()
	a.Rectangle(0, 0, float64(s.width), headerBottom)

	a.FillRadialGradient(float64(s.width)*0.14, 58, 86,
		agg.RGBA(0.18, 0.42, 0.78, 0.42),
		agg.RGBA(0.18, 0.42, 0.78, 0.02), 1.0)
	a.Ellipse(float64(s.width)*0.14, 58, 92, 56)
	a.FillRadialGradient(float64(s.width)*0.85, 34, 74,
		agg.RGBA(0.90, 0.48, 0.16, 0.24),
		agg.RGBA(0.90, 0.48, 0.16, 0.01), 1.0)
	a.Ellipse(float64(s.width)*0.85, 34, 78, 40)

	a.NoFill()
	a.LineColor(agg.RGBA(0.06, 0.15, 0.22, 0.88))
	a.LineWidth(3.0)
	a.AddDash(10, 7)
	a.DashStart(2)
	a.ResetPath()
	a.MoveTo(18, headerBottom-10)
	for i := 0; i < 6; i++ {
		x1 := float64(s.width) * (0.16 + float64(i)*0.13)
		y1 := 24.0 + float64(i%2)*22.0
		x2 := x1 + float64(s.width)*0.065
		y2 := 80.0 - float64(i%3)*9.0
		a.QuadricCurveTo(x1, y1, x2, y2)
	}
	a.DrawPath(agg.StrokeOnly)
	a.NoDashes()

	a.FontGSV(18)
	a.FillColor(agg.RGBA(0.08, 0.10, 0.14, 1.0))
	a.Text(22, 28, "AGG Mixed Benchmark Scene", false, 0, 0)
	a.FontGSV(9.5)
	a.FillColor(agg.RGBA(0.20, 0.24, 0.28, 0.95))
	a.Text(24, 52, "tiles: masks, LUTs, warped resampling, polygon clipping, line patterns, graphs, filter curves", false, 0, 0)

	if s.spheres != nil {
		a.ImageFilter(agg.Bilinear)
		a.ImageResample(agg.ResampleAlways)
		if err := a.TransformImageSimple(s.spheres,
			float64(s.width)-210, 8,
			float64(s.width)-88, 84); err != nil {
			return err
		}
	}
	if s.aggImage != nil {
		parallelogram := []float64{
			float64(s.width) - 330, 16,
			float64(s.width) - 240, 28,
			float64(s.width) - 350, 82,
		}
		if err := a.TransformImageParallelogramSimple(s.aggImage, parallelogram); err != nil {
			return err
		}
	}

	a.ImageResample(agg.NoResample)
	a.ResetTransformations()
	a.ClipBox(0, 0, float64(s.width), float64(s.height))
	return nil
}

func (s *Scene) drawTileBorder(ctx *agg.Context, t tile) {
	a := ctx.GetAgg2D()
	x1, y1 := t.x-4, t.y-4
	x2, y2 := t.x+float64(t.w)+4, t.y+float64(t.h)+4

	a.FillColor(agg.RGBA(1.0, 1.0, 1.0, 0.65))
	a.LineColor(agg.RGBA(0.14, 0.18, 0.24, 0.32))
	a.LineWidth(1.2)
	a.RoundedRect(x1, y1, x2, y2, 8)
	a.DrawPath(agg.FillAndStroke)
}
