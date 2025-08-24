package scanline

import (
	"testing"

	"agg_go/internal/basics"
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
		baseRenderer := &MockBaseRenderer[string]{}
		spanAllocator := &MockSpanAllocator[string]{}
		styleHandler := &MockStyleHandler[string]{}

		RenderScanlinesCompound(rasterizer, scanlineAA, scanlineBin,
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
		baseRenderer := &MockBaseRenderer[string]{}
		spanAllocator := &MockSpanAllocator[string]{}
		styleHandler := &MockStyleHandler[string]{
			solidFlags: []bool{true}, // Style 100 is solid
			colors:     []string{"red"},
		}

		RenderScanlinesCompound(rasterizer, scanlineAA, scanlineBin,
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
		baseRenderer := &MockBaseRenderer[string]{}
		spanAllocator := &MockSpanAllocator[string]{}
		styleHandler := &MockStyleHandler[string]{
			solidFlags: []bool{false}, // Style ID 200 is not solid - but need to check index 0
		}

		RenderScanlinesCompound(rasterizer, scanlineAA, scanlineBin,
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
		baseRenderer := &MockBaseRenderer[string]{}
		spanAllocator := &MockSpanAllocator[string]{}
		styleHandler := &MockStyleHandler[string]{
			solidFlags: []bool{true, true}, // Both styles are solid
			colors:     []string{"red", "blue"},
		}

		RenderScanlinesCompound(rasterizer, scanlineAA, scanlineBin,
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
