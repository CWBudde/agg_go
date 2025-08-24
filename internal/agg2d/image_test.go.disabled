// Package agg provides AGG2D image operations test suite.
package agg

import (
	"testing"
)

// TestTransformImage tests the primary TransformImage method.
func TestTransformImage(t *testing.T) {
	// Create test context
	agg2d := NewAgg2D()

	// Create test buffer and attach
	buf := make([]uint8, 800*600*4) // RGBA buffer
	agg2d.Attach(buf, 800, 600, 800*4)

	// Create test image
	imgBuf := make([]uint8, 100*100*4) // Small test image
	img := NewImage(imgBuf, 100, 100, 100*4)

	// Test basic transformation
	err := agg2d.TransformImage(img, 0, 0, 100, 100, 10, 10, 110, 110)
	if err != nil {
		t.Errorf("TransformImage failed: %v", err)
	}

	// Test with nil image
	err = agg2d.TransformImage(nil, 0, 0, 100, 100, 10, 10, 110, 110)
	if err == nil {
		t.Error("Expected error for nil image")
	}
}

// TestTransformImageSimple tests the simplified TransformImage method.
func TestTransformImageSimple(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]uint8, 400*300*4)
	agg2d.Attach(buf, 400, 300, 400*4)

	imgBuf := make([]uint8, 50*50*4)
	img := NewImage(imgBuf, 50, 50, 50*4)

	err := agg2d.TransformImageSimple(img, 20, 20, 120, 120)
	if err != nil {
		t.Errorf("TransformImageSimple failed: %v", err)
	}

	// Test with nil image
	err = agg2d.TransformImageSimple(nil, 20, 20, 120, 120)
	if err == nil {
		t.Error("Expected error for nil image")
	}
}

// TestTransformImageParallelogram tests parallelogram transformation.
func TestTransformImageParallelogram(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]uint8, 400*300*4)
	agg2d.Attach(buf, 400, 300, 400*4)

	imgBuf := make([]uint8, 60*60*4)
	img := NewImage(imgBuf, 60, 60, 60*4)

	// Define a simple parallelogram (x1,y1, x2,y2, x3,y3)
	parallelogram := []float64{10, 10, 70, 20, 60, 70}

	err := agg2d.TransformImageParallelogram(img, 0, 0, 60, 60, parallelogram)
	if err != nil {
		t.Errorf("TransformImageParallelogram failed: %v", err)
	}

	// Test with invalid parallelogram
	invalidParallelogram := []float64{10, 10, 70} // Wrong size
	err = agg2d.TransformImageParallelogram(img, 0, 0, 60, 60, invalidParallelogram)
	if err == nil {
		t.Error("Expected error for invalid parallelogram")
	}
}

// TestTransformImageParallelogramSimple tests simple parallelogram transformation.
func TestTransformImageParallelogramSimple(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]uint8, 400*300*4)
	agg2d.Attach(buf, 400, 300, 400*4)

	imgBuf := make([]uint8, 40*40*4)
	img := NewImage(imgBuf, 40, 40, 40*4)

	parallelogram := []float64{5, 5, 45, 15, 35, 45}

	err := agg2d.TransformImageParallelogramSimple(img, parallelogram)
	if err != nil {
		t.Errorf("TransformImageParallelogramSimple failed: %v", err)
	}
}

// TestTransformImagePath tests path-based image transformation.
func TestTransformImagePath(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]uint8, 400*300*4)
	agg2d.Attach(buf, 400, 300, 400*4)

	imgBuf := make([]uint8, 30*30*4)
	img := NewImage(imgBuf, 30, 30, 30*4)

	// Create a path first
	agg2d.ResetPath()
	agg2d.MoveTo(50, 50)
	agg2d.LineTo(150, 60)
	agg2d.LineTo(140, 160)
	agg2d.LineTo(40, 150)
	agg2d.ClosePolygon()

	err := agg2d.TransformImagePath(img, 0, 0, 30, 30, 50, 50, 150, 150)
	if err != nil {
		t.Errorf("TransformImagePath failed: %v", err)
	}
}

// TestTransformImagePathSimple tests simple path-based image transformation.
func TestTransformImagePathSimple(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]uint8, 400*300*4)
	agg2d.Attach(buf, 400, 300, 400*4)

	imgBuf := make([]uint8, 35*35*4)
	img := NewImage(imgBuf, 35, 35, 35*4)

	// Create a circular path
	agg2d.ResetPath()
	agg2d.Ellipse(100, 100, 75, 75)

	err := agg2d.TransformImagePathSimple(img, 60, 60, 140, 140)
	if err != nil {
		t.Errorf("TransformImagePathSimple failed: %v", err)
	}
}

// TestTransformImagePathParallelogram tests parallelogram with path transformation.
func TestTransformImagePathParallelogram(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]uint8, 400*300*4)
	agg2d.Attach(buf, 400, 300, 400*4)

	imgBuf := make([]uint8, 25*25*4)
	img := NewImage(imgBuf, 25, 25, 25*4)

	// Create a triangular path
	agg2d.ResetPath()
	agg2d.MoveTo(100, 50)
	agg2d.LineTo(150, 120)
	agg2d.LineTo(50, 120)
	agg2d.ClosePolygon()

	parallelogram := []float64{75, 60, 125, 70, 115, 110}

	err := agg2d.TransformImagePathParallelogram(img, 0, 0, 25, 25, parallelogram)
	if err != nil {
		t.Errorf("TransformImagePathParallelogram failed: %v", err)
	}
}

// TestTransformImagePathParallelogramSimple tests simple parallelogram with path.
func TestTransformImagePathParallelogramSimple(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]uint8, 400*300*4)
	agg2d.Attach(buf, 400, 300, 400*4)

	imgBuf := make([]uint8, 20*20*4)
	img := NewImage(imgBuf, 20, 20, 20*4)

	// Create a rounded rectangle path
	agg2d.ResetPath()
	agg2d.RoundedRect(80, 80, 160, 140, 15)

	parallelogram := []float64{85, 85, 155, 90, 150, 135}

	err := agg2d.TransformImagePathParallelogramSimple(img, parallelogram)
	if err != nil {
		t.Errorf("TransformImagePathParallelogramSimple failed: %v", err)
	}
}

// TestBlendImage tests image blending functionality.
func TestBlendImage(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]uint8, 300*200*4)
	agg2d.Attach(buf, 300, 200, 300*4)

	imgBuf := make([]uint8, 50*50*4)
	img := NewImage(imgBuf, 50, 50, 50*4)

	// Test blending with various alpha values
	err := agg2d.BlendImage(img, 0, 0, 50, 50, 100, 100, 255)
	if err != nil {
		t.Errorf("BlendImage failed: %v", err)
	}

	err = agg2d.BlendImage(img, 10, 10, 40, 40, 150, 100, 128)
	if err != nil {
		t.Errorf("BlendImage with partial alpha failed: %v", err)
	}

	// Test alpha clamping
	err = agg2d.BlendImage(img, 0, 0, 50, 50, 50, 50, 300) // Alpha > 255
	if err != nil {
		t.Errorf("BlendImage with clamped alpha failed: %v", err)
	}

	// Test with nil image
	err = agg2d.BlendImage(nil, 0, 0, 50, 50, 100, 100, 255)
	if err == nil {
		t.Error("Expected error for nil image in BlendImage")
	}
}

// TestBlendImageSimple tests simple image blending.
func TestBlendImageSimple(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]uint8, 200*150*4)
	agg2d.Attach(buf, 200, 150, 200*4)

	imgBuf := make([]uint8, 40*40*4)
	img := NewImage(imgBuf, 40, 40, 40*4)

	err := agg2d.BlendImageSimple(img, 80, 55, 200)
	if err != nil {
		t.Errorf("BlendImageSimple failed: %v", err)
	}

	// Test with nil image
	err = agg2d.BlendImageSimple(nil, 80, 55, 200)
	if err == nil {
		t.Error("Expected error for nil image in BlendImageSimple")
	}
}

// TestCopyImage tests image copying functionality.
func TestCopyImage(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]uint8, 250*180*4)
	agg2d.Attach(buf, 250, 180, 250*4)

	imgBuf := make([]uint8, 60*45*4)
	img := NewImage(imgBuf, 60, 45, 60*4)

	err := agg2d.CopyImage(img, 0, 0, 60, 45, 95, 67)
	if err != nil {
		t.Errorf("CopyImage failed: %v", err)
	}

	err = agg2d.CopyImage(img, 10, 5, 50, 40, 150, 90)
	if err != nil {
		t.Errorf("CopyImage with subregion failed: %v", err)
	}

	// Test with nil image
	err = agg2d.CopyImage(nil, 0, 0, 60, 45, 95, 67)
	if err == nil {
		t.Error("Expected error for nil image in CopyImage")
	}
}

// TestCopyImageSimple tests simple image copying.
func TestCopyImageSimple(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]uint8, 180*120*4)
	agg2d.Attach(buf, 180, 120, 180*4)

	imgBuf := make([]uint8, 30*30*4)
	img := NewImage(imgBuf, 30, 30, 30*4)

	err := agg2d.CopyImageSimple(img, 75, 45)
	if err != nil {
		t.Errorf("CopyImageSimple failed: %v", err)
	}

	// Test with nil image
	err = agg2d.CopyImageSimple(nil, 75, 45)
	if err == nil {
		t.Error("Expected error for nil image in CopyImageSimple")
	}
}

// TestImagePremultiplyDemultiply tests alpha channel operations.
func TestImagePremultiplyDemultiply(t *testing.T) {
	imgBuf := make([]uint8, 20*20*4)
	img := NewImage(imgBuf, 20, 20, 20*4)

	// Test premultiplication
	err := img.Premultiply()
	if err != nil {
		t.Errorf("Premultiply failed: %v", err)
	}

	// Test demultiplication
	err = img.Demultiply()
	if err != nil {
		t.Errorf("Demultiply failed: %v", err)
	}

	// Test with nil buffer
	nilImg := &Image{renBuf: nil}
	err = nilImg.Premultiply()
	if err == nil {
		t.Error("Expected error for nil buffer in Premultiply")
	}

	err = nilImg.Demultiply()
	if err == nil {
		t.Error("Expected error for nil buffer in Demultiply")
	}
}

// TestImageFilterModes tests different image filter settings.
func TestImageFilterModes(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]uint8, 200*150*4)
	agg2d.Attach(buf, 200, 150, 200*4)

	imgBuf := make([]uint8, 40*30*4)
	img := NewImage(imgBuf, 40, 30, 40*4)

	// Test with different filter modes
	filterModes := []ImageFilter{
		NoFilter, Bilinear, Hanning, Hermite, Quadric,
		Bicubic, Catrom, Spline16, Spline36, Blackman144,
	}

	for _, filter := range filterModes {
		agg2d.ImageFilter(filter)

		err := agg2d.TransformImageSimple(img, 50, 50, 90, 80)
		if err != nil {
			t.Errorf("TransformImageSimple failed with filter %d: %v", filter, err)
		}
	}
}

// TestImageResampleModes tests different image resample settings.
func TestImageResampleModes(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]uint8, 300*200*4)
	agg2d.Attach(buf, 300, 200, 300*4)

	imgBuf := make([]uint8, 50*50*4)
	img := NewImage(imgBuf, 50, 50, 50*4)

	// Test with different resample modes
	resampleModes := []ImageResample{
		NoResample, ResampleAlways, ResampleOnZoomOut,
	}

	for _, resample := range resampleModes {
		agg2d.ImageResample(resample)

		err := agg2d.TransformImageSimple(img, 60, 60, 160, 160)
		if err != nil {
			t.Errorf("TransformImageSimple failed with resample %d: %v", resample, err)
		}
	}
}

// TestImageTransformationWithWorldTransform tests image operations with world transforms.
func TestImageTransformationWithWorldTransform(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]uint8, 400*300*4)
	agg2d.Attach(buf, 400, 300, 400*4)

	imgBuf := make([]uint8, 32*32*4)
	img := NewImage(imgBuf, 32, 32, 32*4)

	// Apply world transformations
	agg2d.Rotate(0.5) // 0.5 radians
	agg2d.Scale(1.5, 1.2)
	agg2d.Translate(50, 30)

	err := agg2d.TransformImageSimple(img, 100, 100, 132, 132)
	if err != nil {
		t.Errorf("TransformImageSimple with world transform failed: %v", err)
	}

	// Reset transformations and try again
	agg2d.ResetTransformations()

	err = agg2d.TransformImageSimple(img, 200, 150, 232, 182)
	if err != nil {
		t.Errorf("TransformImageSimple after reset failed: %v", err)
	}
}

// TestRenderImageInternalMethod tests the internal renderImage method.
func TestRenderImageInternalMethod(t *testing.T) {
	agg2d := NewAgg2D()
	buf := make([]uint8, 300*200*4)
	agg2d.Attach(buf, 300, 200, 300*4)

	imgBuf := make([]uint8, 25*25*4)
	img := NewImage(imgBuf, 25, 25, 25*4)

	// Test valid parallelogram
	parallelogram := []float64{50, 50, 75, 55, 70, 75}
	err := agg2d.renderImage(img, 0, 0, 25, 25, parallelogram)
	if err != nil {
		t.Errorf("renderImage failed: %v", err)
	}

	// Test invalid parallelogram (wrong length)
	invalidParallelogram := []float64{50, 50, 75, 55}
	err = agg2d.renderImage(img, 0, 0, 25, 25, invalidParallelogram)
	if err == nil {
		t.Error("Expected error for invalid parallelogram length")
	}

	// Test nil image
	err = agg2d.renderImage(nil, 0, 0, 25, 25, parallelogram)
	if err == nil {
		t.Error("Expected error for nil image")
	}

	// Test image with nil buffer
	emptyImg := &Image{renBuf: nil}
	err = agg2d.renderImage(emptyImg, 0, 0, 25, 25, parallelogram)
	if err == nil {
		t.Error("Expected error for image with nil buffer")
	}
}

// BenchmarkTransformImage benchmarks the image transformation performance.
func BenchmarkTransformImage(b *testing.B) {
	agg2d := NewAgg2D()
	buf := make([]uint8, 800*600*4)
	agg2d.Attach(buf, 800, 600, 800*4)

	imgBuf := make([]uint8, 100*100*4)
	img := NewImage(imgBuf, 100, 100, 100*4)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agg2d.TransformImage(img, 0, 0, 100, 100, 50, 50, 150, 150)
	}
}

// BenchmarkBlendImage benchmarks the image blending performance.
func BenchmarkBlendImage(b *testing.B) {
	agg2d := NewAgg2D()
	buf := make([]uint8, 400*300*4)
	agg2d.Attach(buf, 400, 300, 400*4)

	imgBuf := make([]uint8, 64*64*4)
	img := NewImage(imgBuf, 64, 64, 64*4)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agg2d.BlendImage(img, 0, 0, 64, 64, 100, 100, 128)
	}
}
