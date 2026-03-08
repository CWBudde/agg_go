package patternperspective

import (
	agg "agg_go"
	"agg_go/internal/demo/imageassets"
	"agg_go/internal/demo/quadwarp"
	imgacc "agg_go/internal/image"
)

type Config struct {
	Mode int
	Quad [4][2]float64
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
	if mode > 2 {
		mode = 2
	}

	ctx.Clear(agg.White)

	transformMode := quadwarp.TransformPerspective
	interpMode := quadwarp.InterpolatorLinearSubdiv
	forceParallelogram := false
	switch mode {
	case 0:
		transformMode = quadwarp.TransformAffine
		interpMode = quadwarp.InterpolatorLinear
		forceParallelogram = true
	case 1:
		transformMode = quadwarp.TransformBilinear
		interpMode = quadwarp.InterpolatorLinear
	}

	quadwarp.Draw(ctx, quadwarp.Config{
		CanvasWidth:        ctx.GetImage().Width(),
		CanvasHeight:       ctx.GetImage().Height(),
		Source:             cachedAgg,
		SourceRect:         [4]float64{-150, -150, 150, 150},
		Quad:               cfg.Quad,
		Transform:          transformMode,
		Interpolator:       interpMode,
		Sampling:           quadwarp.SampleFilter2x2,
		SourceMode:         quadwarp.SourceWrapReflect,
		FilterKernel:       imgacc.HanningFilter{},
		Normalize:          true,
		ForceParallelogram: forceParallelogram,
		ShowQuadFill:       true,
		ShowQuadOutline:    true,
		ShowHandles:        true,
		QuadFillColor:      agg.RGBA(0, 0.3, 0.5, 0.16),
		QuadLineColor:      agg.RGBA(0, 0.25, 0.35, 0.9),
	})
}
