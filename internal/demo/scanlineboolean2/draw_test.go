package scanlineboolean2

import (
	"image/color"
	"math"
	"testing"

	agg "github.com/cwbudde/agg_go"
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
	cfg.CenterX -= frameOffX
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

// TestDefaultSceneSpanCountMatchesCpp pins the output span count of the default
// reference scene (Great Britain + Spiral, non-zero, scanline_u, AND) to the
// value produced by the upstream C++ scanline_boolean2 demo. A mismatch means
// the rasterized shapes diverged from AGG (e.g. the GB-poly data regressed); the
// count is exact because the whole scanline_u8 + AND pipeline is span-preserving.
func TestDefaultSceneSpanCountMatchesCpp(t *testing.T) {
	const w, h = 655.0, 520.0
	cfg := Config{Mode: 3, FillRule: 1, ScanlineType: 1, Operation: 2, CenterX: w * 0.5, CenterY: h * 0.5}

	// Mirror Draw()'s shape construction (zero frame offset at 655x520, no Y mirror).
	a, b := buildShapes(cfg, w, h)

	ctx := agg.NewContext(int(w), int(h))
	_, _, numSpans := combineAndRender(ctx.GetImage(), cfg, a, b)

	const want = 1031
	if numSpans != want {
		t.Fatalf("default scene num_spans = %d, want %d (C++ oracle)", numSpans, want)
	}
}

func TestDrawUsesFlipYBufferWithoutManualSceneMirror(t *testing.T) {
	const width, height = 655, 520
	img := agg.NewImage(make([]uint8, width*height*4), width, height, -width*4)
	ctx := agg.NewContextForImage(img)

	Draw(ctx, Config{
		Mode:         3,
		FillRule:     1,
		ScanlineType: 1,
		Operation:    0,
		CenterX:      width * 0.5,
		CenterY:      height * 0.5,
	})

	got := img.ToGoImage()
	if got == nil {
		t.Fatal("ToGoImage returned nil")
	}
	northernScotland := got.RGBAAt(275, 45)
	if nearWhite(northernScotland) {
		t.Fatalf("northern GB sample is white, scene appears vertically mirrored: %v", northernScotland)
	}
}

func TestBlendPixelUsesImageStride(t *testing.T) {
	const width, height = 3, 2
	img := agg.NewImage(make([]uint8, width*height*4), width, height, -width*4)

	blendPixel(img, 1, 0, colorDef{r: 1, g: 0, b: 0, a: 1}, 255)

	wantOff := (height-1)*width*4 + 1*4
	if got := img.Data[wantOff : wantOff+4]; got[0] != 255 || got[1] != 0 || got[2] != 0 || got[3] != 255 {
		t.Fatalf("logical row 0 pixel not written through negative stride, got RGBA=%v", got)
	}
	wrongOff := 1 * 4
	if got := img.Data[wrongOff : wrongOff+4]; got[0] != 0 || got[1] != 0 || got[2] != 0 || got[3] != 0 {
		t.Fatalf("blendPixel wrote through raw positive offset despite negative stride, got RGBA=%v", got)
	}
}

func nearWhite(c color.RGBA) bool {
	return c.R > 252 && c.G > 252 && c.B > 252
}
