package scanlineboolean2

import (
	"math"
	"testing"

	agg "github.com/MeKo-Christian/agg_go"
)

func TestCombineAndRenderProducesResult(t *testing.T) {
	ctx := agg.NewContext(800, 600)
	img := ctx.GetImage()
	cfg := Config{
		Mode:         3,
		FillRule:     1,
		ScanlineType: 1,
		Operation:    2,
		CenterX:      400,
		CenterY:      300,
	}

	frameOffX := (800.0 - referenceWidth) * 0.5
	frameOffY := (600.0 - referenceHeight) * 0.5
	cfg.CenterX = cfg.CenterX - frameOffX
	cfg.CenterY = referenceHeight - (cfg.CenterY - frameOffY)

	a, b := buildShapes(cfg, referenceWidth, referenceHeight)
	a = transformContours(mirrorContoursY(a, referenceHeight), 0, 0, 1, 1, frameOffX, frameOffY)
	b = transformContours(mirrorContoursY(b, referenceHeight), 0, 0, 1, 1, frameOffX, frameOffY)

	_, _, numSpans := combineAndRender(img, cfg, a, b)
	if numSpans == 0 {
		t.Fatal("combineAndRender returned zero spans")
	}

	hasResultPixel := false
	for i := 0; i+2 < len(img.Data); i += 4 {
		r, g, b := img.Data[i], img.Data[i+1], img.Data[i+2]
		if r > g && r > b && math.Abs(float64(r)-float64(g)) > 10 {
			hasResultPixel = true
			break
		}
	}
	if !hasResultPixel {
		t.Fatal("combineAndRender did not produce any result-colored pixels")
	}
}
