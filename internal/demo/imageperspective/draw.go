package imageperspective

import (
	agg "github.com/cwbudde/agg_go"
	icol "github.com/cwbudde/agg_go/internal/color"
	ctrlbase "github.com/cwbudde/agg_go/internal/ctrl"
	polygonctrl "github.com/cwbudde/agg_go/internal/ctrl/polygon"
	"github.com/cwbudde/agg_go/internal/demo/imageassets"
	"github.com/cwbudde/agg_go/internal/demo/quadwarp"
	imgacc "github.com/cwbudde/agg_go/internal/image"
	"github.com/cwbudde/agg_go/internal/rasterizer"
)

type Config struct {
	Mode   int
	Quad   [4][2]float64
	Source *agg.Image
}

var cachedSpheres *agg.Image

type ctrlVertexSourceAdapter struct {
	ctrl ctrlbase.Ctrl[icol.RGBA]
}

func (a *ctrlVertexSourceAdapter) Rewind(pathID uint32) {
	a.ctrl.Rewind(uint(pathID))
}

func (a *ctrlVertexSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

func toDisplayAggColor(c icol.RGBA) agg.Color {
	clamp := func(v float64) uint8 {
		switch {
		case v <= 0:
			return 0
		case v >= 1:
			return 255
		default:
			return uint8(v*255.0 + 0.5)
		}
	}

	srgb := icol.ConvertToSRGBFromLinear(icol.RGBA8[icol.Linear]{
		R: clamp(c.R),
		G: clamp(c.G),
		B: clamp(c.B),
		A: clamp(c.A),
	})
	return agg.NewColor(srgb.R, srgb.G, srgb.B, srgb.A)
}

func renderCtrl(
	a *agg.Agg2D,
	ras interface {
		Reset()
		AddPath(vs rasterizer.VertexSource, pathID uint32)
	},
	c ctrlbase.Ctrl[icol.RGBA],
) {
	for pathID := uint(0); pathID < c.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(&ctrlVertexSourceAdapter{ctrl: c}, uint32(pathID))
		a.RenderRasterizerWithColor(toDisplayAggColor(c.Color(pathID)))
	}
}

func renderQuadTool(ctx *agg.Context, quad [4][2]float64) {
	tool := polygonctrl.NewPolygonCtrl[icol.RGBA](4, 5.0, icol.NewRGBA(0, 0.3, 0.5, 0.6))
	tool.SetClose(true)
	tool.SetLineWidth(1.0)
	for i, pt := range quad {
		tool.SetXn(uint(i), pt[0])
		tool.SetYn(uint(i), pt[1])
	}
	a := ctx.GetAgg2D()
	renderCtrl(a, a.GetInternalRasterizer(), tool)
}

func flippedVerticalCopy(src *agg.Image) *agg.Image {
	if src == nil {
		return nil
	}
	goImg := src.ToGoImage()
	if goImg == nil {
		return nil
	}

	w, h := goImg.Bounds().Dx(), goImg.Bounds().Dy()
	buf := make([]byte, len(goImg.Pix))
	rowBytes := w * 4
	for y := 0; y < h; y++ {
		srcOff := (h - 1 - y) * goImg.Stride
		dstOff := y * rowBytes
		copy(buf[dstOff:dstOff+rowBytes], goImg.Pix[srcOff:srcOff+rowBytes])
	}

	return agg.NewImage(buf, w, h, rowBytes)
}

func Draw(ctx *agg.Context, cfg Config) {
	if cachedSpheres == nil {
		img, err := imageassets.Spheres()
		if err != nil {
			return
		}
		cachedSpheres = img
	}

	mode := cfg.Mode
	if mode < 0 {
		mode = 0
	}
	if mode > 2 {
		mode = 2
	}

	ctx.Clear(agg.White)

	source := cfg.Source
	if source == nil {
		source = cachedSpheres
	}
	source = flippedVerticalCopy(source)

	transformMode := quadwarp.TransformPerspective
	interpMode := quadwarp.InterpolatorTrans
	sampling := quadwarp.SampleFilter2x2
	forceParallelogram := false
	switch mode {
	case 0:
		transformMode = quadwarp.TransformAffine
		interpMode = quadwarp.InterpolatorLinear
		sampling = quadwarp.SampleNearest
		forceParallelogram = true
	case 1:
		transformMode = quadwarp.TransformBilinear
		interpMode = quadwarp.InterpolatorLinear
		sampling = quadwarp.SampleFilter2x2
	}

	quad := cfg.Quad
	if forceParallelogram {
		quad[3][0] = quad[0][0] + (quad[2][0] - quad[1][0])
		quad[3][1] = quad[0][1] + (quad[2][1] - quad[1][1])
	}
	renderQuadTool(ctx, quad)

	quadwarp.Draw(ctx, quadwarp.Config{
		CanvasWidth:        ctx.GetImage().Width(),
		CanvasHeight:       ctx.GetImage().Height(),
		Source:             source,
		SourceRect:         [4]float64{0, 0, float64(source.Width()), float64(source.Height())},
		Quad:               quad,
		Transform:          transformMode,
		Interpolator:       interpMode,
		Sampling:           sampling,
		SourceMode:         quadwarp.SourceClone,
		FilterKernel:       imgacc.BilinearFilter{},
		Normalize:          false,
		ForceParallelogram: forceParallelogram,
	})
}
