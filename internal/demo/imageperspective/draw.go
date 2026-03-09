package imageperspective

import (
	agg "agg_go"
	"github.com/MeKo-Christian/agg_go/internal/demo/imageassets"
	"github.com/MeKo-Christian/agg_go/internal/demo/quadwarp"
	imgacc "github.com/MeKo-Christian/agg_go/internal/image"
)

type Config struct {
	Mode int
	Quad [4][2]float64
}

var cachedSpheres *agg.Image

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

	quadwarp.Draw(ctx, quadwarp.Config{
		CanvasWidth:        ctx.GetImage().Width(),
		CanvasHeight:       ctx.GetImage().Height(),
		Source:             cachedSpheres,
		SourceRect:         [4]float64{0, 0, float64(cachedSpheres.Width()), float64(cachedSpheres.Height())},
		Quad:               cfg.Quad,
		Transform:          transformMode,
		Interpolator:       interpMode,
		Sampling:           sampling,
		SourceMode:         quadwarp.SourceClone,
		FilterKernel:       imgacc.BilinearFilter{},
		Normalize:          false,
		ForceParallelogram: forceParallelogram,
		ShowQuadFill:       true,
		ShowQuadOutline:    true,
		ShowHandles:        true,
		QuadFillColor:      agg.RGBA(0, 0.3, 0.5, 0.16),
		QuadLineColor:      agg.RGBA(0, 0.25, 0.35, 0.9),
	})
}
