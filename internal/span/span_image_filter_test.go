package span

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/image"
	"agg_go/internal/transform"
)

// MockSource implements SourceInterface for testing
type MockSource struct {
	width  int
	height int
}

func (ms MockSource) Width() int  { return ms.width }
func (ms MockSource) Height() int { return ms.height }

// TestSpanImageFilter_Construction tests basic construction and initialization
func TestSpanImageFilter_Construction(t *testing.T) {
	// Test default construction
	filter := NewSpanImageFilter[MockSource, *SpanInterpolatorLinear[*transform.TransAffine]]()

	if filter == nil {
		t.Fatal("Expected filter to be created")
	}

	// Check default values
	if filter.FilterDxDbl() != 0.5 {
		t.Errorf("Expected default dx to be 0.5, got %f", filter.FilterDxDbl())
	}

	if filter.FilterDyDbl() != 0.5 {
		t.Errorf("Expected default dy to be 0.5, got %f", filter.FilterDyDbl())
	}

	expectedDxInt := image.ImageSubpixelScale / 2
	if filter.FilterDxInt() != expectedDxInt {
		t.Errorf("Expected default dx int to be %d, got %d", expectedDxInt, filter.FilterDxInt())
	}

	expectedDyInt := image.ImageSubpixelScale / 2
	if filter.FilterDyInt() != expectedDyInt {
		t.Errorf("Expected default dy int to be %d, got %d", expectedDyInt, filter.FilterDyInt())
	}
}

// TestSpanImageFilter_WithParams tests construction with parameters
func TestSpanImageFilter_WithParams(t *testing.T) {
	source := MockSource{width: 100, height: 100}
	affine := transform.NewTransAffine()
	interpolator := NewSpanInterpolatorLinear(affine, image.ImageSubpixelShift)
	imageFilter := image.NewImageFilterLUT()

	filter := NewSpanImageFilterWithParams(source, interpolator, imageFilter)

	if filter == nil {
		t.Fatal("Expected filter to be created")
	}

	if filter.Source().Width() != source.Width() || filter.Source().Height() != source.Height() {
		t.Error("Expected source to be attached")
	}

	if filter.Interpolator() != interpolator {
		t.Error("Expected interpolator to be attached")
	}

	if filter.Filter() != imageFilter {
		t.Error("Expected image filter to be attached")
	}
}

// TestSpanImageFilter_FilterOffset tests filter offset functionality
func TestSpanImageFilter_FilterOffset(t *testing.T) {
	filter := NewSpanImageFilter[MockSource, *SpanInterpolatorLinear[*transform.TransAffine]]()

	// Test setting custom offset
	dx, dy := 0.25, 0.75
	filter.FilterOffset(dx, dy)

	if filter.FilterDxDbl() != dx {
		t.Errorf("Expected dx to be %f, got %f", dx, filter.FilterDxDbl())
	}

	if filter.FilterDyDbl() != dy {
		t.Errorf("Expected dy to be %f, got %f", dy, filter.FilterDyDbl())
	}

	expectedDxInt := basics.IRound(dx * float64(image.ImageSubpixelScale))
	if filter.FilterDxInt() != expectedDxInt {
		t.Errorf("Expected dx int to be %d, got %d", expectedDxInt, filter.FilterDxInt())
	}

	expectedDyInt := basics.IRound(dy * float64(image.ImageSubpixelScale))
	if filter.FilterDyInt() != expectedDyInt {
		t.Errorf("Expected dy int to be %d, got %d", expectedDyInt, filter.FilterDyInt())
	}
}

// TestSpanImageFilter_FilterOffsetUniform tests uniform filter offset
func TestSpanImageFilter_FilterOffsetUniform(t *testing.T) {
	filter := NewSpanImageFilter[MockSource, *SpanInterpolatorLinear[*transform.TransAffine]]()

	offset := 0.3
	filter.FilterOffsetUniform(offset)

	if filter.FilterDxDbl() != offset {
		t.Errorf("Expected dx to be %f, got %f", offset, filter.FilterDxDbl())
	}

	if filter.FilterDyDbl() != offset {
		t.Errorf("Expected dy to be %f, got %f", offset, filter.FilterDyDbl())
	}
}

// TestSpanImageFilter_Attach tests source attachment
func TestSpanImageFilter_Attach(t *testing.T) {
	filter := NewSpanImageFilter[MockSource, *SpanInterpolatorLinear[*transform.TransAffine]]()

	source := MockSource{width: 200, height: 150}
	filter.Attach(source)

	if filter.Source().Width() != source.Width() || filter.Source().Height() != source.Height() {
		t.Error("Expected source to be attached")
	}

	if filter.Source().Width() != 200 {
		t.Errorf("Expected source width to be 200, got %d", filter.Source().Width())
	}

	if filter.Source().Height() != 150 {
		t.Errorf("Expected source height to be 150, got %d", filter.Source().Height())
	}
}

// TestSpanImageFilter_Prepare tests base prepare method
func TestSpanImageFilter_Prepare(t *testing.T) {
	filter := NewSpanImageFilter[MockSource, *SpanInterpolatorLinear[*transform.TransAffine]]()

	// Base prepare should not panic and should be a no-op
	filter.Prepare()

	// Should be able to call multiple times
	filter.Prepare()
	filter.Prepare()
}

// TestSpanImageResampleAffine_Construction tests affine resampling construction
func TestSpanImageResampleAffine_Construction(t *testing.T) {
	// Test default construction
	filter := NewSpanImageResampleAffine[MockSource]()

	if filter == nil {
		t.Fatal("Expected filter to be created")
	}

	// Check default values
	if filter.ScaleLimit() != 200 {
		t.Errorf("Expected default scale limit to be 200, got %d", filter.ScaleLimit())
	}

	if filter.BlurX() != 1.0 {
		t.Errorf("Expected default blur X to be 1.0, got %f", filter.BlurX())
	}

	if filter.BlurY() != 1.0 {
		t.Errorf("Expected default blur Y to be 1.0, got %f", filter.BlurY())
	}
}

// TestSpanImageResampleAffine_WithParams tests affine construction with parameters
func TestSpanImageResampleAffine_WithParams(t *testing.T) {
	source := MockSource{width: 100, height: 100}
	affine := transform.NewTransAffine()
	interpolator := NewSpanInterpolatorLinear(affine, image.ImageSubpixelShift)
	imageFilter := image.NewImageFilterLUTWithFilter(image.BilinearFilter{}, true)

	filter := NewSpanImageResampleAffineWithParams(source, interpolator, imageFilter)

	if filter == nil {
		t.Fatal("Expected filter to be created")
	}

	if filter.Source() != source {
		t.Error("Expected source to be attached")
	}

	if filter.Interpolator() != interpolator {
		t.Error("Expected interpolator to be attached")
	}

	if filter.Filter() != imageFilter {
		t.Error("Expected image filter to be attached")
	}
}

// TestSpanImageResampleAffine_ScaleLimit tests scale limit functionality
func TestSpanImageResampleAffine_ScaleLimit(t *testing.T) {
	filter := NewSpanImageResampleAffine[MockSource]()

	// Test setting custom scale limit
	filter.SetScaleLimit(150)

	if filter.ScaleLimit() != 150 {
		t.Errorf("Expected scale limit to be 150, got %d", filter.ScaleLimit())
	}
}

// TestSpanImageResampleAffine_Blur tests blur factor functionality
func TestSpanImageResampleAffine_Blur(t *testing.T) {
	filter := NewSpanImageResampleAffine[MockSource]()

	// Test setting individual blur factors
	filter.SetBlurX(1.5)
	filter.SetBlurY(2.0)

	if filter.BlurX() != 1.5 {
		t.Errorf("Expected blur X to be 1.5, got %f", filter.BlurX())
	}

	if filter.BlurY() != 2.0 {
		t.Errorf("Expected blur Y to be 2.0, got %f", filter.BlurY())
	}

	// Test setting uniform blur
	filter.Blur(1.25)

	if filter.BlurX() != 1.25 {
		t.Errorf("Expected blur X to be 1.25, got %f", filter.BlurX())
	}

	if filter.BlurY() != 1.25 {
		t.Errorf("Expected blur Y to be 1.25, got %f", filter.BlurY())
	}
}

// TestSpanImageResampleAffine_Prepare tests affine preparation
func TestSpanImageResampleAffine_Prepare(t *testing.T) {
	filter := NewSpanImageResampleAffine[MockSource]()

	// Test with no interpolator (should not panic)
	filter.Prepare()

	// Test with interpolator but no transformer
	affine := transform.NewTransAffine()
	interpolator := NewSpanInterpolatorLinear(affine, image.ImageSubpixelShift)
	filter.SetInterpolator(interpolator)
	filter.Prepare()

	// Verify scaling factors are calculated
	if filter.RX() == 0 && filter.RY() == 0 {
		t.Error("Expected non-zero scaling factors after prepare")
	}
}

// TestSpanImageResampleAffine_PrepareWithScaling tests affine preparation with scaling
func TestSpanImageResampleAffine_PrepareWithScaling(t *testing.T) {
	filter := NewSpanImageResampleAffine[MockSource]()

	// Create a scaled affine transformation
	affine := transform.NewTransAffine()
	affine.ScaleXY(2.0, 1.5) // Scale by 2x in X, 1.5x in Y

	interpolator := NewSpanInterpolatorLinear(affine, image.ImageSubpixelShift)
	filter.SetInterpolator(interpolator)

	filter.Prepare()

	// Check that scaling factors are calculated
	if filter.RX() == 0 {
		t.Error("Expected non-zero RX after prepare with scaling")
	}

	if filter.RY() == 0 {
		t.Error("Expected non-zero RY after prepare with scaling")
	}

	if filter.RXInv() == 0 {
		t.Error("Expected non-zero RXInv after prepare with scaling")
	}

	if filter.RYInv() == 0 {
		t.Error("Expected non-zero RYInv after prepare with scaling")
	}
}

// TestSpanImageResample_Construction tests general resampling construction
func TestSpanImageResample_Construction(t *testing.T) {
	// Test default construction
	filter := NewSpanImageResample[MockSource, *SpanInterpolatorLinear[*transform.TransAffine]]()

	if filter == nil {
		t.Fatal("Expected filter to be created")
	}

	// Check default values
	if filter.ScaleLimit() != 20 {
		t.Errorf("Expected default scale limit to be 20, got %d", filter.ScaleLimit())
	}

	expectedBlur := 1.0 // image.ImageSubpixelScale / image.ImageSubpixelScale
	if filter.BlurX() != expectedBlur {
		t.Errorf("Expected default blur X to be %f, got %f", expectedBlur, filter.BlurX())
	}

	if filter.BlurY() != expectedBlur {
		t.Errorf("Expected default blur Y to be %f, got %f", expectedBlur, filter.BlurY())
	}
}

// TestSpanImageResample_WithParams tests general resampling with parameters
func TestSpanImageResample_WithParams(t *testing.T) {
	source := MockSource{width: 100, height: 100}
	affine := transform.NewTransAffine()
	interpolator := NewSpanInterpolatorLinear(affine, image.ImageSubpixelShift)
	imageFilter := image.NewImageFilterLUTWithFilter(image.BicubicFilter{}, true)

	filter := NewSpanImageResampleWithParams(source, interpolator, imageFilter)

	if filter == nil {
		t.Fatal("Expected filter to be created")
	}

	if filter.Source() != source {
		t.Error("Expected source to be attached")
	}

	if filter.Interpolator() != interpolator {
		t.Error("Expected interpolator to be attached")
	}

	if filter.Filter() != imageFilter {
		t.Error("Expected image filter to be attached")
	}
}

// TestSpanImageResample_ScaleLimit tests general scale limit functionality
func TestSpanImageResample_ScaleLimit(t *testing.T) {
	filter := NewSpanImageResample[MockSource, *SpanInterpolatorLinear[*transform.TransAffine]]()

	// Test setting custom scale limit
	filter.SetScaleLimit(30)

	if filter.ScaleLimit() != 30 {
		t.Errorf("Expected scale limit to be 30, got %d", filter.ScaleLimit())
	}
}

// TestSpanImageResample_Blur tests general blur functionality
func TestSpanImageResample_Blur(t *testing.T) {
	filter := NewSpanImageResample[MockSource, *SpanInterpolatorLinear[*transform.TransAffine]]()

	// Test setting individual blur factors
	filter.SetBlurX(0.8)
	filter.SetBlurY(1.2)

	if absDiff(filter.BlurX()-0.8) > 0.01 {
		t.Errorf("Expected blur X to be approximately 0.8, got %f", filter.BlurX())
	}

	if absDiff(filter.BlurY()-1.2) > 0.01 {
		t.Errorf("Expected blur Y to be approximately 1.2, got %f", filter.BlurY())
	}

	// Test setting uniform blur
	filter.Blur(0.9)

	if absDiff(filter.BlurX()-0.9) > 0.01 {
		t.Errorf("Expected blur X to be approximately 0.9, got %f", filter.BlurX())
	}

	if absDiff(filter.BlurY()-0.9) > 0.01 {
		t.Errorf("Expected blur Y to be approximately 0.9, got %f", filter.BlurY())
	}
}

// TestSpanImageResample_AdjustScale tests scale adjustment functionality
func TestSpanImageResample_AdjustScale(t *testing.T) {
	filter := NewSpanImageResample[MockSource, *SpanInterpolatorLinear[*transform.TransAffine]]()

	// Test scale adjustment with values below minimum
	rx, ry := 10, 20
	filter.AdjustScale(&rx, &ry)

	if rx < image.ImageSubpixelScale {
		t.Errorf("Expected rx to be at least %d, got %d", image.ImageSubpixelScale, rx)
	}

	if ry < image.ImageSubpixelScale {
		t.Errorf("Expected ry to be at least %d, got %d", image.ImageSubpixelScale, ry)
	}

	// Test scale adjustment with values above maximum
	filter.SetScaleLimit(10)
	rx, ry = image.ImageSubpixelScale*50, image.ImageSubpixelScale*60
	filter.AdjustScale(&rx, &ry)

	maxScale := image.ImageSubpixelScale * 10
	if rx > maxScale {
		t.Errorf("Expected rx to be at most %d, got %d", maxScale, rx)
	}

	if ry > maxScale {
		t.Errorf("Expected ry to be at most %d, got %d", maxScale, ry)
	}
}

// TestSpanImageResample_AdjustScaleWithBlur tests scale adjustment with blur factors
func TestSpanImageResample_AdjustScaleWithBlur(t *testing.T) {
	filter := NewSpanImageResample[MockSource, *SpanInterpolatorLinear[*transform.TransAffine]]()

	// Set blur factors
	filter.SetBlurX(0.5)
	filter.SetBlurY(2.0)

	// Start with normal scale values
	rx, ry := image.ImageSubpixelScale*2, image.ImageSubpixelScale*2
	originalRx, originalRy := rx, ry

	filter.AdjustScale(&rx, &ry)

	// Blur should affect the scale values
	if rx == originalRx {
		t.Error("Expected rx to be modified by blur factor")
	}

	if ry == originalRy {
		t.Error("Expected ry to be modified by blur factor")
	}

	// Ensure minimum scale is still enforced after blur
	if rx < image.ImageSubpixelScale {
		t.Errorf("Expected rx to be at least %d after blur, got %d", image.ImageSubpixelScale, rx)
	}

	if ry < image.ImageSubpixelScale {
		t.Errorf("Expected ry to be at least %d after blur, got %d", image.ImageSubpixelScale, ry)
	}
}

// TestSpanImageFilter_FilterIntegration tests integration with different filter types
func TestSpanImageFilter_FilterIntegration(t *testing.T) {
	filter := NewSpanImageFilter[MockSource, *SpanInterpolatorLinear[*transform.TransAffine]]()

	// Test with different filter types
	testCases := []struct {
		name       string
		filterFunc image.FilterFunction
	}{
		{"Bilinear", image.BilinearFilter{}},
		{"Bicubic", image.BicubicFilter{}},
		{"Hermite", image.HermiteFilter{}},
		{"Hanning", image.HanningFilter{}},
		{"Hamming", image.HammingFilter{}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			imageFilter := image.NewImageFilterLUTWithFilter(tc.filterFunc, true)
			filter.SetFilter(imageFilter)

			if filter.Filter() != imageFilter {
				t.Errorf("Failed to set %s filter", tc.name)
			}

			if filter.Filter().Radius() != tc.filterFunc.Radius() {
				t.Errorf("Filter radius mismatch for %s: expected %f, got %f",
					tc.name, tc.filterFunc.Radius(), filter.Filter().Radius())
			}
		})
	}
}

// Helper function for float comparison
func absDiff(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
