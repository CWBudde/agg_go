package imageresample

import (
	"math"
	"sync"
	"time"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/demo/imageassets"
	"github.com/MeKo-Christian/agg_go/internal/demo/quadwarp"
	imgacc "github.com/MeKo-Christian/agg_go/internal/image"
)

type Config struct {
	Mode int
	Blur float64
	Quad [4][2]float64
}

var (
	cachedSpheres *agg.Image
	once          sync.Once
)

func flippedVerticalCopy(src *agg.Image) *agg.Image {
	if src == nil {
		return nil
	}
	w, h := src.Width(), src.Height()
	rowBytes := w * 4
	buf := make([]byte, len(src.Data))
	for y := 0; y < h; y++ {
		srcOff := (h - 1 - y) * rowBytes
		dstOff := y * rowBytes
		copy(buf[dstOff:dstOff+rowBytes], src.Data[srcOff:srcOff+rowBytes])
	}
	return agg.NewImage(buf, w, h, rowBytes)
}

func DrawTimed(ctx *agg.Context, cfg Config) time.Duration {
	start := time.Now()
	Draw(ctx, cfg)
	return time.Since(start)
}

func Draw(ctx *agg.Context, cfg Config) {
	once.Do(func() {
		img, err := imageassets.Spheres()
		if err == nil {
			cachedSpheres = img
		}
	})
	if ctx == nil || cachedSpheres == nil {
		return
	}

	source := flippedVerticalCopy(cachedSpheres)

	mode := cfg.Mode
	if mode < 0 {
		mode = 0
	}
	if mode > 5 {
		mode = 5
	}

	blur := cfg.Blur
	if blur < 0.5 {
		blur = 0.5
	}
	if blur > 5.0 {
		blur = 5.0
	}

	ctx.Clear(agg.White)

	transformMode := quadwarp.TransformPerspective
	interpMode := quadwarp.InterpolatorPerspectiveLerp
	sampling := quadwarp.SampleResample
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
		Source:             source,
		SourceRect:         [4]float64{0, 0, float64(source.Width()), float64(source.Height())},
		Quad:               cfg.Quad,
		Transform:          transformMode,
		Interpolator:       interpMode,
		Sampling:           sampling,
		SourceMode:         quadwarp.SourceClone,
		FilterKernel:       imgacc.BilinearFilter{},
		Normalize:          true,
		Blur:               blur,
		ForceParallelogram: forceParallelogram,
		ShowQuadFill:       true,
		ShowQuadOutline:    true,
		ShowHandles:        true,
		QuadFillColor:      agg.RGBA(0, 0.3, 0.5, 0.5),
		QuadLineColor:      agg.RGBA(0, 0.2, 0.3, 0.9),
	})
}

func RotateQuad90(quad *[4][2]float64) {
	if quad == nil {
		return
	}
	cx := (quad[0][0] + quad[1][0] + quad[2][0] + quad[3][0]) / 4.0
	cy := (quad[0][1] + quad[1][1] + quad[2][1] + quad[3][1]) / 4.0
	s, c := math.Sincos(math.Pi / 2.0)
	for i := range quad {
		dx := quad[i][0] - cx
		dy := quad[i][1] - cy
		quad[i][0] = cx + dx*c - dy*s
		quad[i][1] = cy + dx*s + dy*c
	}
}
