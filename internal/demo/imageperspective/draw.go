package imageperspective

import (
	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/internal/demo/imageassets"
	"github.com/cwbudde/agg_go/internal/demo/quadwarp"
	imgacc "github.com/cwbudde/agg_go/internal/image"
)

type Config struct {
	Mode   int
	Quad   [4][2]float64
	Source *agg.Image
}

var cachedSpheres *agg.Image

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
		Normalize:          false,
		ForceParallelogram: forceParallelogram,
		ShowQuadFill:       true,
		ShowQuadOutline:    true,
		ShowHandles:        true,
		QuadFillColor:      agg.RGBA(0, 0.3, 0.5, 0.6),
		QuadLineColor:      agg.RGBA(0, 0, 0, 0.9),
	})
}
