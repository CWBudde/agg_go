package patternresample

import (
	"math"
	"sync"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/demo/imageassets"
	"github.com/MeKo-Christian/agg_go/internal/demo/quadwarp"
	imgacc "github.com/MeKo-Christian/agg_go/internal/image"
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

	applyGammaInvLUT(ctx.GetImage(), gamma)
}

func gammaCacheKey(gamma float64) int {
	return int(math.Round(gamma * 1000))
}

func gammaAdjustedSource(gamma float64) *agg.Image {
	if cachedAgg == nil {
		return nil
	}
	if gamma <= 0 || math.Abs(gamma-1.0) < 1e-9 {
		return cachedAgg
	}

	key := gammaCacheKey(gamma)
	if cached, ok := gammaImageCache.Load(key); ok {
		return cached.(*agg.Image)
	}

	src := quadwarp.CopyWithGammaDir(cachedAgg, gamma)
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
