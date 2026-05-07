package patternresample

import (
	"math"
	"sync"
	"time"

	agg "github.com/cwbudde/agg_go"
	icol "github.com/cwbudde/agg_go/internal/color"
	ctrlbase "github.com/cwbudde/agg_go/internal/ctrl"
	polygonctrl "github.com/cwbudde/agg_go/internal/ctrl/polygon"
	"github.com/cwbudde/agg_go/internal/demo/imageassets"
	"github.com/cwbudde/agg_go/internal/demo/quadwarp"
	imgacc "github.com/cwbudde/agg_go/internal/image"
	"github.com/cwbudde/agg_go/internal/rasterizer"
)

const rgbaByteScale = 1.0 / 255.0

type Config struct {
	Mode  int
	Gamma float64
	Blur  float64
	Quad  [4][2]float64
}

var (
	cachedAgg       *agg.Image
	gammaImageCache sync.Map
)

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

func toRawAggColor(c icol.RGBA) agg.Color {
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

	return agg.NewColor(clamp(c.R), clamp(c.G), clamp(c.B), clamp(c.A))
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
		a.RenderRasterizerWithColor(toRawAggColor(c.Color(pathID)))
	}
}

func quadToolColor() icol.RGBA {
	return icol.NewRGBA(0, 0.3, 0.5, 0.1)
}

func renderQuadTool(ctx *agg.Context, quad [4][2]float64) {
	tool := polygonctrl.NewPolygonCtrl[icol.RGBA](4, 5.0, quadToolColor())
	tool.SetClose(true)
	for i, pt := range quad {
		tool.SetXn(uint(i), pt[0])
		tool.SetYn(uint(i), pt[1])
	}
	a := ctx.GetAgg2D()
	renderCtrl(a, a.GetInternalRasterizer(), tool)
}

func Draw(ctx *agg.Context, cfg Config) {
	_ = draw(ctx, cfg)
}

// DrawTimed renders the demo and returns the time spent in the core image
// resampling pass, matching the original AGG timer label more closely.
func DrawTimed(ctx *agg.Context, cfg Config) time.Duration {
	return draw(ctx, cfg)
}

func draw(ctx *agg.Context, cfg Config) time.Duration {
	if cachedAgg == nil {
		img, err := imageassets.Agg()
		if err != nil {
			return 0
		}
		cachedAgg = img
	}

	mode := cfg.Mode
	if mode < 0 {
		mode = 0
	}
	if mode > 5 {
		mode = 5
	}
	gamma := cfg.Gamma
	if gamma < 0.5 {
		gamma = 0.5
	}
	if gamma > 3.0 {
		gamma = 3.0
	}
	blur := cfg.Blur
	if blur < 0.5 {
		blur = 0.5
	}
	if blur > 2.0 {
		blur = 2.0
	}

	src := gammaAdjustedSource(gamma)

	ctx.Clear(agg.White)

	transformMode := quadwarp.TransformPerspective
	interpMode := quadwarp.InterpolatorLinearSubdiv
	sampling := quadwarp.SampleFilter2x2
	forceParallelogram := false

	switch mode {
	case 0:
		transformMode = quadwarp.TransformAffine
		interpMode = quadwarp.InterpolatorLinear
		sampling = quadwarp.SampleFilter2x2
		forceParallelogram = true
	case 1:
		transformMode = quadwarp.TransformAffine
		interpMode = quadwarp.InterpolatorLinear
		sampling = quadwarp.SampleResample
		forceParallelogram = true
	case 2:
		transformMode = quadwarp.TransformPerspective
		interpMode = quadwarp.InterpolatorLinearSubdiv
		sampling = quadwarp.SampleFilter2x2
	case 3:
		transformMode = quadwarp.TransformPerspective
		interpMode = quadwarp.InterpolatorTrans
		sampling = quadwarp.SampleFilter2x2
	case 4:
		transformMode = quadwarp.TransformPerspective
		interpMode = quadwarp.InterpolatorPerspectiveLerp
		sampling = quadwarp.SampleResample
	case 5:
		transformMode = quadwarp.TransformPerspective
		interpMode = quadwarp.InterpolatorPerspectiveExact
		sampling = quadwarp.SampleResample
	}

	quad := cfg.Quad
	if forceParallelogram {
		quad[3][0] = quad[0][0] + (quad[2][0] - quad[1][0])
		quad[3][1] = quad[0][1] + (quad[2][1] - quad[1][1])
	}
	renderQuadTool(ctx, quad)

	start := time.Now()
	quadwarp.Draw(ctx, quadwarp.Config{
		CanvasWidth:        ctx.GetImage().Width(),
		CanvasHeight:       ctx.GetImage().Height(),
		Source:             src,
		SourceRect:         [4]float64{-150, -150, 150, 150},
		Quad:               quad,
		Transform:          transformMode,
		Interpolator:       interpMode,
		Sampling:           sampling,
		SourceMode:         quadwarp.SourceWrapReflect,
		FilterKernel:       imgacc.HanningFilter{},
		Normalize:          true,
		Blur:               blur,
		ForceParallelogram: forceParallelogram,
	})
	elapsed := time.Since(start)

	applyGammaInvLUT(ctx.GetImage(), gamma)
	return elapsed
}

func gammaCacheKey(gamma float64) int {
	return int(math.Round(gamma * 1000))
}

func gammaAdjustedSource(gamma float64) *agg.Image {
	if cachedAgg == nil {
		return nil
	}
	base := quadwarp.CopyFlippedVertical(cachedAgg)
	if base == nil {
		return nil
	}
	if gamma <= 0 || math.Abs(gamma-1.0) < 1e-9 {
		return base
	}

	key := gammaCacheKey(gamma)
	if cached, ok := gammaImageCache.Load(key); ok {
		return cached.(*agg.Image)
	}

	src := quadwarp.CopyWithGammaDir(base, gamma)
	actual, _ := gammaImageCache.LoadOrStore(key, src)
	return actual.(*agg.Image)
}

func applyGammaInvLUT(img *agg.Image, gamma float64) {
	if img == nil || gamma <= 0 || math.Abs(gamma-1.0) < 1e-9 {
		return
	}

	inv := 1.0 / gamma
	var lut [256]byte
	for i := range lut {
		v := math.Pow(float64(i)*rgbaByteScale, inv)
		lut[i] = byte(v*255.0 + 0.5)
	}

	for i := 0; i+3 < len(img.Data); i += 4 {
		img.Data[i] = lut[img.Data[i]]
		img.Data[i+1] = lut[img.Data[i+1]]
		img.Data[i+2] = lut[img.Data[i+2]]
	}
}
