package patternresample

import (
	agg "agg_go"
	"github.com/MeKo-Christian/agg_go/internal/demo/imageassets"
	"github.com/MeKo-Christian/agg_go/internal/demo/quadwarp"
	imgacc "github.com/MeKo-Christian/agg_go/internal/image"
)

type Config struct {
	Mode  int
	Gamma float64
	Blur  float64
	Quad  [4][2]float64
}

var cachedAgg *agg.Image

func Draw(ctx *agg.Context, cfg Config) {
	if cachedAgg == nil {
		img, err := imageassets.Agg()
		if err != nil {
			return
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

	src := quadwarp.CopyWithGammaDir(cachedAgg, gamma)

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

	quadwarp.Draw(ctx, quadwarp.Config{
		CanvasWidth:        ctx.GetImage().Width(),
		CanvasHeight:       ctx.GetImage().Height(),
		Source:             src,
		SourceRect:         [4]float64{-150, -150, 150, 150},
		Quad:               cfg.Quad,
		Transform:          transformMode,
		Interpolator:       interpMode,
		Sampling:           sampling,
		SourceMode:         quadwarp.SourceWrapReflect,
		FilterKernel:       imgacc.HanningFilter{},
		Normalize:          true,
		Blur:               blur,
		ForceParallelogram: forceParallelogram,
		ShowQuadFill:       true,
		ShowQuadOutline:    true,
		ShowHandles:        true,
		QuadFillColor:      agg.RGBA(0, 0.3, 0.5, 0.12),
		QuadLineColor:      agg.RGBA(0, 0.25, 0.35, 0.9),
	})

	quadwarp.ApplyGammaInv(ctx.GetImage(), gamma)
}
