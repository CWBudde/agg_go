package patternresample

import (
	"bytes"
	"testing"

	agg "github.com/cwbudde/agg_go"
)

func TestGammaAdjustedSourceUsesAGGFlipYImageOrientation(t *testing.T) {
	oldCachedAgg := cachedAgg
	oldGammaCache := gammaImageCache
	t.Cleanup(func() {
		cachedAgg = oldCachedAgg
		gammaImageCache = oldGammaCache
	})

	cachedAgg = agg.NewImage([]byte{
		10, 20, 30, 255,
		40, 50, 60, 255,
		70, 80, 90, 255,
		100, 110, 120, 255,
	}, 2, 2, 8)
	gammaImageCache.Clear()

	got := gammaAdjustedSource(1.0)
	want := []byte{
		70, 80, 90, 255,
		100, 110, 120, 255,
		10, 20, 30, 255,
		40, 50, 60, 255,
	}
	if got == nil {
		t.Fatal("gammaAdjustedSource(1) returned nil")
	}
	if !bytes.Equal(got.Data, want) {
		t.Fatalf("gammaAdjustedSource(1) data = %v, want AGG flip_y source orientation %v", got.Data, want)
	}
	if got == cachedAgg {
		t.Fatalf("gammaAdjustedSource(1) returned cachedAgg directly; want orientation-adjusted copy")
	}
}

func TestQuadToolColorMatchesPatternResampleSource(t *testing.T) {
	got := toRawAggColor(quadToolColor())
	if got.R != 0 || got.G != 77 || got.B != 128 || got.A != 26 {
		t.Fatalf("pattern_resample quad tool color = rgba(%d,%d,%d,%d), want C++ rgba(0,0.3,0.5,0.1) rounded to rgba(0,77,128,26)",
			got.R, got.G, got.B, got.A)
	}
}
