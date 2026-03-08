package patternresample

import (
	"math"

	agg "agg_go"
	"agg_go/internal/demo/imageassets"
	"agg_go/internal/demo/quadwarp"
	imgacc "agg_go/internal/image"
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
	if (mode == 1 || mode == 4 || mode == 5) && blur > 1.0 {
		src = blurApprox(src, blur)
	}

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

func blurApprox(src *agg.Image, blur float64) *agg.Image {
	if src == nil || blur <= 1.0 {
		return src
	}
	iterations := int(math.Round((blur - 1.0) * 4.0))
	if iterations < 1 {
		iterations = 1
	}
	w := src.Width()
	h := src.Height()
	cur := append([]byte(nil), src.Data...)
	tmp := make([]byte, len(cur))

	for it := 0; it < iterations; it++ {
		for y := 0; y < h; y++ {
			y0 := maxInt(y-1, 0)
			y1 := minInt(y+1, h-1)
			for x := 0; x < w; x++ {
				x0 := maxInt(x-1, 0)
				x1 := minInt(x+1, w-1)
				var rs, gs, bs int
				n := 0
				for yy := y0; yy <= y1; yy++ {
					for xx := x0; xx <= x1; xx++ {
						i := (yy*w + xx) * 4
						rs += int(cur[i+0])
						gs += int(cur[i+1])
						bs += int(cur[i+2])
						n++
					}
				}
				o := (y*w + x) * 4
				tmp[o+0] = uint8(rs / n)
				tmp[o+1] = uint8(gs / n)
				tmp[o+2] = uint8(bs / n)
				tmp[o+3] = cur[o+3]
			}
		}
		cur, tmp = tmp, cur
	}
	return agg.NewImage(cur, w, h, w*4)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
