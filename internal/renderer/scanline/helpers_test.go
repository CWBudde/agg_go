package scanline

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
)

func TestRenderScanlines(t *testing.T) {
	t.Run("early return if rewind fails", func(t *testing.T) {
		rasterizer := &MockRasterizer{rewindResult: false}
		scanline := &MockScanline{}
		renderer := &MockRenderer[string]{}

		RenderScanlines(rasterizer, scanline, renderer)

		if renderer.prepareCalled {
			t.Error("Prepare should not be called if RewindScanlines returns false")
		}
		if len(renderer.renderCalls) > 0 {
			t.Error("Render should not be called if RewindScanlines returns false")
		}
	})

	t.Run("successful rendering", func(t *testing.T) {
		rasterizer := &MockRasterizer{
			rewindResult: true,
			sweepResults: []bool{true, true, false}, // Two successful sweeps, then done
			minX:         10,
			maxX:         50,
		}
		scanline := &MockScanline{}
		renderer := &MockRenderer[string]{}

		RenderScanlines(rasterizer, scanline, renderer)

		if !renderer.prepareCalled {
			t.Error("Prepare should be called")
		}
		if len(renderer.renderCalls) != 2 {
			t.Errorf("Expected 2 render calls, got %d", len(renderer.renderCalls))
		}
	})

	t.Run("scanline reset if supported", func(t *testing.T) {
		rasterizer := &MockRasterizer{
			rewindResult: true,
			sweepResults: []bool{false}, // No sweeps needed
			minX:         5,
			maxX:         15,
		}

		// Use a scanline that supports reset tracking
		resetScanline := &MockResettableScanline{}

		renderer := &MockRenderer[string]{}

		RenderScanlines(rasterizer, resetScanline, renderer)

		if !resetScanline.ResetCalled {
			t.Error("Reset should be called on resettable scanline")
		}
		if resetScanline.ResetMinX != 5 || resetScanline.ResetMaxX != 15 {
			t.Errorf("Reset called with wrong bounds: got (%d, %d), expected (5, 15)",
				resetScanline.ResetMinX, resetScanline.ResetMaxX)
		}
	})
}

func TestRenderAllPaths(t *testing.T) {
	t.Run("render multiple paths with colors", func(t *testing.T) {
		rasterizer := &MockRasterizer{rewindResult: true}
		scanline := &MockScanline{}
		renderer := &MockRenderer[string]{}
		vertexSource := &MockVertexSource{}

		colorStorage := &MockPathColorStorage[string]{
			colors:     []string{"red", "blue", "green"},
			defaultVal: "default",
		}
		pathIdStorage := &MockPathIdStorage{
			pathIds: []int{1, 2, 3},
		}

		RenderAllPaths(rasterizer, scanline, renderer, vertexSource,
			colorStorage, pathIdStorage, 3)

		// Verify that the renderer had colors set
		// Note: This is a simplified test as the full functionality would require
		// more complex mock implementations
		if renderer.color != "green" { // Last color set
			t.Errorf("Expected final color to be 'green', got %v", renderer.color)
		}
	})

	t.Run("zero paths", func(t *testing.T) {
		rasterizer := &MockRasterizer{rewindResult: true}
		scanline := &MockScanline{}
		renderer := &MockRenderer[string]{}
		vertexSource := &MockVertexSource{}
		colorStorage := &MockPathColorStorage[string]{}
		pathIdStorage := &MockPathIdStorage{}

		RenderAllPaths(rasterizer, scanline, renderer, vertexSource,
			colorStorage, pathIdStorage, 0)

		// Should not have called renderer
		if len(renderer.renderCalls) > 0 {
			t.Error("Should not have rendered any paths for numPaths=0")
		}
	})
}

func TestRenderScanlinesCompound(t *testing.T) {
	t.Run("early return if rewind fails", func(t *testing.T) {
		rasterizer := &MockCompoundRasterizer{
			MockRasterizer: MockRasterizer{rewindResult: false},
		}
		scanlineAA := &MockScanline{}
		scanlineBin := &MockScanline{}
		baseRenderer := &MockBaseRenderer[color.RGBA8[color.Linear]]{}
		spanAllocator := &MockSpanAllocator[color.RGBA8[color.Linear]]{}
		styleHandler := &MockStyleHandler[color.RGBA8[color.Linear]]{}

		RenderScanlinesCompound[color.RGBA8[color.Linear], *color.RGBA8[color.Linear]](rasterizer, scanlineAA, scanlineBin,
			baseRenderer, spanAllocator, styleHandler)

		if len(spanAllocator.allocations) > 0 {
			t.Error("Should not allocate anything if RewindScanlines returns false")
		}
	})

	t.Run("single style solid rendering", func(t *testing.T) {
		rasterizer := &MockCompoundRasterizer{
			MockRasterizer: MockRasterizer{
				rewindResult: true,
				minX:         0,
				maxX:         10,
			},
			sweepStylesResults: []int{1, 0},      // One style, then done
			styleResults:       [][]bool{{true}}, // Style 0 has one successful sweep
			styles:             []int{100},
		}

		scanlineAA := &MockScanline{
			y:        5,
			numSpans: 1,
			spans:    []SpanData{{X: 2, Len: 3, Covers: []basics.Int8u{255, 200, 150}}},
		}
		scanlineBin := &MockScanline{}
		baseRenderer := &MockBaseRenderer[color.RGBA8[color.Linear]]{}
		spanAllocator := &MockSpanAllocator[color.RGBA8[color.Linear]]{}
		styleHandler := &MockStyleHandler[color.RGBA8[color.Linear]]{
			solidFlags: []bool{true}, // Style 100 is solid
			colors:     []color.RGBA8[color.Linear]{{R: 255, G: 0, B: 0, A: 255}},
		}

		RenderScanlinesCompound[color.RGBA8[color.Linear], *color.RGBA8[color.Linear]](rasterizer, scanlineAA, scanlineBin,
			baseRenderer, spanAllocator, styleHandler)

		// Should have allocated the main color span buffer
		if len(spanAllocator.allocations) == 0 {
			t.Error("Expected span allocation for compound rendering")
		}

		// Should have called solid hspan blending
		if len(baseRenderer.solidHspanCalls) == 0 {
			t.Error("Expected solid hspan calls for solid style")
		}
	})

	t.Run("single style generated rendering", func(t *testing.T) {
		rasterizer := &MockCompoundRasterizer{
			MockRasterizer: MockRasterizer{
				rewindResult: true,
				sweepResults: []bool{true}, // One sweep for style sweeping
				minX:         0,
				maxX:         10,
			},
			sweepStylesResults: []int{1, 0},      // One style, then done
			styleResults:       [][]bool{{true}}, // Style 0 has one successful sweep
			styles:             []int{200},
		}

		scanlineAA := &MockScanline{
			y:        8,
			numSpans: 1,
			spans:    []SpanData{{X: 5, Len: 2, Covers: []basics.Int8u{255, 128}}},
		}
		scanlineBin := &MockScanline{}
		baseRenderer := &MockBaseRenderer[color.RGBA8[color.Linear]]{}
		spanAllocator := &MockSpanAllocator[color.RGBA8[color.Linear]]{}
		styleHandler := &MockStyleHandler[color.RGBA8[color.Linear]]{
			solidFlags: []bool{false}, // Style ID 200 is not solid - but need to check index 0
		}

		RenderScanlinesCompound[color.RGBA8[color.Linear], *color.RGBA8[color.Linear]](rasterizer, scanlineAA, scanlineBin,
			baseRenderer, spanAllocator, styleHandler)

		// Should have allocated spans
		if len(spanAllocator.allocations) == 0 {
			t.Error("Expected span allocations for compound rendering")
		}

		// Should have called span generation
		if len(styleHandler.generateCalls) == 0 {
			t.Errorf("Expected span generation calls for non-solid style. SweepStyles calls: %v, StyleResults: %v",
				rasterizer.sweepStylesResults, rasterizer.styleResults)
		}

		// Should have called color hspan blending
		if len(baseRenderer.colorHspanCalls) == 0 {
			t.Error("Expected color hspan calls for generated style")
		}
	})

	t.Run("multiple styles rendering", func(t *testing.T) {
		rasterizer := &MockCompoundRasterizer{
			MockRasterizer: MockRasterizer{
				rewindResult: true,
				sweepResults: []bool{true, true, true}, // Binary sweep + style sweeps
				minX:         0,
				maxX:         5,
			},
			sweepStylesResults: []int{2, 0}, // Two styles, then done
			styleResults: [][]bool{
				{true}, // Style 0 has one successful sweep
				{true}, // Style 1 has one successful sweep
			},
			styles: []int{300, 301},
		}

		scanlineAA := &MockScanline{
			y:        10,
			numSpans: 1,
			spans:    []SpanData{{X: 1, Len: 3, Covers: []basics.Int8u{255, 200, 150}}},
		}
		scanlineBin := &MockScanline{
			y:        10,
			numSpans: 1,
			spans:    []SpanData{{X: 1, Len: 3, Covers: []basics.Int8u{255, 255, 255}}},
		}
		baseRenderer := &MockBaseRenderer[color.RGBA8[color.Linear]]{}
		spanAllocator := &MockSpanAllocator[color.RGBA8[color.Linear]]{}
		styleHandler := &MockStyleHandler[color.RGBA8[color.Linear]]{
			solidFlags: []bool{true, true}, // Both styles are solid
			colors:     []color.RGBA8[color.Linear]{{R: 255, G: 0, B: 0, A: 255}, {R: 0, G: 0, B: 255, A: 255}},
		}

		RenderScanlinesCompound[color.RGBA8[color.Linear], *color.RGBA8[color.Linear]](rasterizer, scanlineAA, scanlineBin,
			baseRenderer, spanAllocator, styleHandler)

		// Should process multiple styles and emit final result
		if len(spanAllocator.allocations) == 0 {
			t.Error("Expected span allocation for multiple styles")
		}

		// Should have final color hspan call for the blended result
		if len(baseRenderer.colorHspanCalls) == 0 {
			t.Error("Expected final color hspan call for multiple styles")
		}
	})
}

// Test AddWithCover method directly on color types (used by compound rendering)

// Test RGBA8 AddWithCover method behavior for edge cases

func TestRGBA8_AddWithCover_FullOpaque(t *testing.T) {
	dest := color.RGBA8[color.Linear]{R: 50, G: 75, B: 100, A: 200}
	src := color.RGBA8[color.Linear]{R: 100, G: 150, B: 200, A: 255} // Fully opaque

	dest.AddWithCover(src, basics.CoverFull)

	// With full cover and fully opaque source, dest should become src
	if dest != src {
		t.Errorf("Expected dest to be replaced by fully opaque src, got %+v, want %+v", dest, src)
	}
}

func TestRGBA8_AddWithCover_ZeroCover(t *testing.T) {
	original := color.RGBA8[color.Linear]{R: 50, G: 75, B: 100, A: 200}
	dest := original
	src := color.RGBA8[color.Linear]{R: 100, G: 150, B: 200, A: 255}

	dest.AddWithCover(src, 0)

	// With zero cover, dest should remain unchanged
	if dest != original {
		t.Errorf("Expected dest to remain unchanged with zero cover, got %+v, want %+v", dest, original)
	}
}

func TestRGBA8_AddWithCover_Saturation(t *testing.T) {
	dest := color.RGBA8[color.Linear]{R: 200, G: 200, B: 200, A: 200}
	src := color.RGBA8[color.Linear]{R: 100, G: 150, B: 200, A: 100} // Not fully opaque

	dest.AddWithCover(src, basics.CoverFull)

	// Components should saturate at 255
	if dest.R != 255 || dest.G != 255 || dest.B != 255 {
		t.Errorf("Expected RGB components to saturate at 255, got %+v", dest)
	}
	// Alpha should also saturate
	if dest.A != 255 {
		t.Errorf("Expected alpha to saturate at 255, got %d", dest.A)
	}
}

// Test RGBA32 AddWithCover saturation at 1.0
func TestRGBA32_AddWithCover_Saturation(t *testing.T) {
	dest := color.RGBA32[color.Linear]{R: 0.8, G: 0.8, B: 0.8, A: 0.8}
	src := color.RGBA32[color.Linear]{R: 0.4, G: 0.6, B: 0.8, A: 0.5} // Not fully opaque

	dest.AddWithCover(src, basics.CoverFull)

	// Components should saturate at 1.0
	if dest.R != 1.0 || dest.G > 1.0 || dest.B > 1.0 {
		t.Errorf("Expected RGB components to saturate at 1.0, got %+v", dest)
	}
	// Alpha should also saturate
	if dest.A > 1.0 {
		t.Errorf("Expected alpha to be <= 1.0, got %f", dest.A)
	}
}

// Test Gray8 AddWithCover behavior
func TestGray8_AddWithCover_CallsAdd(t *testing.T) {
	original := color.Gray8[color.Linear]{V: 50, A: 128}
	dest := original
	src := color.Gray8[color.Linear]{V: 100, A: 255}
	cover := basics.Int8u(127)

	// Call AddWithCover
	dest.AddWithCover(src, cover)

	// Call Add directly for comparison
	expected := original
	expected.Add(src, cover)

	if dest != expected {
		t.Errorf("AddWithCover should behave same as Add, got %+v, want %+v", dest, expected)
	}
}

// Test coverage accumulation in compound rendering
func TestRenderCompoundSolidStyle_CoverageAccumulation(t *testing.T) {
	// Create test data
	span := SpanData{
		X:      2,
		Len:    3,
		Covers: []basics.Int8u{100, 200, 150}, // Coverage values that would exceed 255 when accumulated
	}

	styleHandler := &MockStyleHandler[color.RGBA8[color.Linear]]{
		colors: []color.RGBA8[color.Linear]{{R: 255, G: 0, B: 0, A: 255}}, // Red color
	}

	mixBuffer := make([]color.RGBA8[color.Linear], 10)
	coverBuffer := make([]basics.Int8u, 10)
	minX := 0

	// First call - should add coverage normally
	renderCompoundSolidStyle(span, styleHandler, 0, mixBuffer, coverBuffer, minX)

	// Check that coverage was accumulated
	expectedCoverages := []basics.Int8u{100, 200, 150}
	for i, expectedCover := range expectedCoverages {
		if coverBuffer[span.X+i] != expectedCover {
			t.Errorf("Coverage[%d] = %d, want %d", span.X+i, coverBuffer[span.X+i], expectedCover)
		}
	}

	// Second call with same span - should clamp coverage to not exceed CoverFull
	renderCompoundSolidStyle(span, styleHandler, 0, mixBuffer, coverBuffer, minX)

	// Check that coverage was accumulated correctly, clamped where needed
	expectedFinalCoverages := []basics.Int8u{200, 255, 255} // 100+100, min(200+200,255), min(150+150,255)
	for i, expectedCover := range expectedFinalCoverages {
		if coverBuffer[span.X+i] != expectedCover {
			t.Errorf("Coverage[%d] = %d, want %d", span.X+i, coverBuffer[span.X+i], expectedCover)
		}
	}
}

func TestRenderCompoundGeneratedStyle_CoverageAccumulation(t *testing.T) {
	// Create test data
	span := SpanData{
		X:      1,
		Len:    2,
		Covers: []basics.Int8u{128, 200}, // Coverage values
	}

	sl := &MockScanline{y: 5}
	styleHandler := &MockStyleHandler[color.RGBA8[color.Linear]]{}
	alloc := &MockSpanAllocator[color.RGBA8[color.Linear]]{}

	mixBuffer := make([]color.RGBA8[color.Linear], 10)
	coverBuffer := make([]basics.Int8u, 10)
	colorSpan := make([]color.RGBA8[color.Linear], 10)
	minX := 0

	// First call - should add coverage normally
	renderCompoundGeneratedStyle(span, sl, styleHandler, 0, colorSpan, mixBuffer, coverBuffer, minX, alloc)

	// Check that coverage was accumulated
	expectedCoverages := []basics.Int8u{128, 200}
	for i, expectedCover := range expectedCoverages {
		if coverBuffer[span.X+i] != expectedCover {
			t.Errorf("Coverage[%d] = %d, want %d", span.X+i, coverBuffer[span.X+i], expectedCover)
		}
	}

	// Second call with overlapping coverage - should clamp to not exceed CoverFull
	renderCompoundGeneratedStyle(span, sl, styleHandler, 0, colorSpan, mixBuffer, coverBuffer, minX, alloc)

	// Check clamping: 128+128=256->255, 200+200=400->255
	for i := range expectedCoverages {
		if coverBuffer[span.X+i] != basics.CoverFull {
			t.Errorf("Coverage[%d] = %d, want %d (CoverFull)", span.X+i, coverBuffer[span.X+i], basics.CoverFull)
		}
	}
}

func TestRenderCompoundSolidStyle_ZeroCoverageSkipped(t *testing.T) {
	// Test that zero coverage pixels are skipped
	span := SpanData{
		X:      0,
		Len:    3,
		Covers: []basics.Int8u{100, 0, 150}, // Middle pixel has zero coverage
	}

	styleHandler := &MockStyleHandler[color.RGBA8[color.Linear]]{
		colors: []color.RGBA8[color.Linear]{{R: 128, G: 128, B: 128, A: 255}}, // Gray color
	}

	initialColor := color.RGBA8[color.Linear]{R: 64, G: 64, B: 64, A: 128}
	mixBuffer := []color.RGBA8[color.Linear]{initialColor, initialColor, initialColor}
	coverBuffer := []basics.Int8u{200, 200, 100} // Pre-existing high coverage
	minX := 0

	// Call the function
	renderCompoundSolidStyle(span, styleHandler, 0, mixBuffer, coverBuffer, minX)

	// Check that the zero-coverage pixel was skipped and buffer unchanged
	if mixBuffer[1] != initialColor {
		t.Errorf("Zero-coverage pixel should not be modified, got %v", mixBuffer[1])
	}
	if coverBuffer[1] != 200 {
		t.Errorf("Zero-coverage pixel coverage should remain unchanged, got %d", coverBuffer[1])
	}

	// Check that non-zero coverage pixels were processed
	if coverBuffer[0] != 255 { // 200+100 clamped to 255
		t.Errorf("Expected coverage[0] to be clamped to 255, got %d", coverBuffer[0])
	}
	if coverBuffer[2] != 250 { // 100+150 = 250
		t.Errorf("Expected coverage[2] to be 250, got %d", coverBuffer[2])
	}
}
