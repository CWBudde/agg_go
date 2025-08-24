package scanline

import (
	"testing"

	"agg_go/internal/basics"
)

func TestRendererScanlineBin(t *testing.T) {
	baseRenderer := &MockBaseRenderer[string]{}
	spanAllocator := &MockSpanAllocator[string]{}
	spanGenerator := &MockSpanGenerator[string]{}

	t.Run("creation", func(t *testing.T) {
		renderer := NewRendererScanlineBin[*MockBaseRenderer[string], *MockSpanAllocator[string], *MockSpanGenerator[string], string]()

		if renderer == nil {
			t.Fatal("NewRendererScanlineBin returned nil")
		}
	})

	t.Run("creation with components", func(t *testing.T) {
		renderer := NewRendererScanlineBinWithComponents(baseRenderer, spanAllocator, spanGenerator)

		if renderer.BaseRenderer() != baseRenderer {
			t.Error("Base renderer not set correctly")
		}
		if renderer.SpanAllocator() != spanAllocator {
			t.Error("Span allocator not set correctly")
		}
		if renderer.SpanGenerator() != spanGenerator {
			t.Error("Span generator not set correctly")
		}
	})

	t.Run("attach all components", func(t *testing.T) {
		renderer := NewRendererScanlineBin[*MockBaseRenderer[string], *MockSpanAllocator[string], *MockSpanGenerator[string], string]()

		renderer.Attach(baseRenderer, spanAllocator, spanGenerator)

		if renderer.BaseRenderer() != baseRenderer {
			t.Error("Base renderer not attached correctly")
		}
		if renderer.SpanAllocator() != spanAllocator {
			t.Error("Span allocator not attached correctly")
		}
		if renderer.SpanGenerator() != spanGenerator {
			t.Error("Span generator not attached correctly")
		}
	})

	t.Run("attach individual components", func(t *testing.T) {
		renderer := NewRendererScanlineBin[*MockBaseRenderer[string], *MockSpanAllocator[string], *MockSpanGenerator[string], string]()

		renderer.AttachBaseRenderer(baseRenderer)
		renderer.AttachSpanAllocator(spanAllocator)
		renderer.AttachSpanGenerator(spanGenerator)

		if renderer.BaseRenderer() != baseRenderer {
			t.Error("Base renderer not attached correctly")
		}
		if renderer.SpanAllocator() != spanAllocator {
			t.Error("Span allocator not attached correctly")
		}
		if renderer.SpanGenerator() != spanGenerator {
			t.Error("Span generator not attached correctly")
		}
	})

	t.Run("prepare calls span generator", func(t *testing.T) {
		renderer := NewRendererScanlineBinWithComponents(baseRenderer, spanAllocator, spanGenerator)

		// Reset mock state
		spanGenerator.prepareCalled = false

		renderer.Prepare()

		if !spanGenerator.prepareCalled {
			t.Error("Prepare() should call span generator's Prepare()")
		}
	})

	t.Run("render calls RenderScanlineBin", func(t *testing.T) {
		renderer := NewRendererScanlineBinWithComponents(baseRenderer, spanAllocator, spanGenerator)

		// Reset mock state
		baseRenderer.colorHspanCalls = nil
		spanAllocator.allocations = nil
		spanGenerator.generateCalls = nil

		scanline := &MockScanline{
			y:        15,
			numSpans: 1,
			spans: []SpanData{
				{X: 8, Len: 4, Covers: []basics.Int8u{255, 255, 255, 255}}, // Binary rendering uses solid coverage
			},
		}

		renderer.Render(scanline)

		// Verify that the underlying RenderScanlineBin function was called
		// by checking if the span allocator and generator were used
		if len(spanAllocator.allocations) == 0 {
			t.Error("Render should have called span allocator")
		}
		if len(spanGenerator.generateCalls) == 0 {
			t.Error("Render should have called span generator")
		}
		if len(baseRenderer.colorHspanCalls) == 0 {
			t.Error("Render should have called base renderer")
		}
	})

	t.Run("getters return correct values", func(t *testing.T) {
		renderer := NewRendererScanlineBinWithComponents(baseRenderer, spanAllocator, spanGenerator)

		if renderer.BaseRenderer() != baseRenderer {
			t.Error("BaseRenderer() returns wrong value")
		}
		if renderer.SpanAllocator() != spanAllocator {
			t.Error("SpanAllocator() returns wrong value")
		}
		if renderer.SpanGenerator() != spanGenerator {
			t.Error("SpanGenerator() returns wrong value")
		}
	})
}
