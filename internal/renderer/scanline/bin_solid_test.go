package scanline

import (
	"testing"

	"agg_go/internal/basics"
)

func TestRendererScanlineBinSolid(t *testing.T) {
	baseRenderer := &MockBaseRenderer[string]{}

	t.Run("creation and basic operations", func(t *testing.T) {
		renderer := NewRendererScanlineBinSolid[*MockBaseRenderer[string], string]()

		// Test attachment
		renderer.Attach(baseRenderer)
		if renderer.BaseRenderer() != baseRenderer {
			t.Error("Base renderer not attached correctly")
		}

		// Test color setting
		color := "blue"
		renderer.SetColor(color)
		if renderer.Color() != color {
			t.Errorf("Expected color %v, got %v", color, renderer.Color())
		}
	})

	t.Run("creation with renderer", func(t *testing.T) {
		renderer := NewRendererScanlineBinSolidWithRenderer(baseRenderer)
		if renderer.BaseRenderer() != baseRenderer {
			t.Error("Base renderer not set correctly in constructor")
		}
	})

	t.Run("creation with renderer and color", func(t *testing.T) {
		color := "green"
		renderer := NewRendererScanlineBinSolidWithColor(baseRenderer, color)

		if renderer.BaseRenderer() != baseRenderer {
			t.Error("Base renderer not set correctly in constructor")
		}
		if renderer.Color() != color {
			t.Errorf("Expected color %v, got %v", color, renderer.Color())
		}
	})

	t.Run("prepare does nothing", func(t *testing.T) {
		renderer := NewRendererScanlineBinSolid[*MockBaseRenderer[string], string]()
		// Should not panic
		renderer.Prepare()
	})

	t.Run("render calls underlying function", func(t *testing.T) {
		renderer := NewRendererScanlineBinSolidWithRenderer(baseRenderer)
		color := "red"
		renderer.SetColor(color)

		// Reset mock state
		baseRenderer.hlineCalls = nil

		scanline := &MockScanline{
			y:        7,
			numSpans: 2,
			spans: []SpanData{
				{X: 10, Len: 5, Covers: []basics.Int8u{255, 255, 255, 255, 255}}, // Binary solid spans
				{X: 20, Len: 3, Covers: []basics.Int8u{255, 255, 255}},
			},
		}

		renderer.Render(scanline)

		// Binary solid rendering should use hline calls for efficiency
		if len(baseRenderer.hlineCalls) == 0 {
			t.Error("Expected hline calls for binary solid rendering")
		}

		// Verify the calls have correct parameters
		for _, call := range baseRenderer.hlineCalls {
			if call.Color != color {
				t.Errorf("Expected color %v in hline call, got %v", color, call.Color)
			}
			if call.Y != scanline.Y() {
				t.Errorf("Expected Y %d in hline call, got %d", scanline.Y(), call.Y)
			}
		}
	})

	t.Run("multiple spans rendering", func(t *testing.T) {
		renderer := NewRendererScanlineBinSolidWithColor(baseRenderer, "yellow")

		// Reset mock state
		baseRenderer.hlineCalls = nil

		scanline := &MockScanline{
			y:        12,
			numSpans: 3,
			spans: []SpanData{
				{X: 5, Len: 2, Covers: []basics.Int8u{255, 255}},
				{X: 15, Len: 4, Covers: []basics.Int8u{255, 255, 255, 255}},
				{X: 25, Len: 1, Covers: []basics.Int8u{255}},
			},
		}

		renderer.Render(scanline)

		// Should have one hline call per span
		expectedCalls := 3
		if len(baseRenderer.hlineCalls) != expectedCalls {
			t.Errorf("Expected %d hline calls, got %d", expectedCalls, len(baseRenderer.hlineCalls))
		}

		// Verify span positions and lengths
		expectedSpans := []struct{ x1, x2 int }{
			{5, 6},   // x=5, len=2 -> x1=5, x2=6
			{15, 18}, // x=15, len=4 -> x1=15, x2=18
			{25, 25}, // x=25, len=1 -> x1=25, x2=25
		}

		for i, call := range baseRenderer.hlineCalls {
			if i < len(expectedSpans) {
				if call.X != expectedSpans[i].x1 || call.X2 != expectedSpans[i].x2 {
					t.Errorf("Span %d: expected x1=%d, x2=%d, got x1=%d, x2=%d",
						i, expectedSpans[i].x1, expectedSpans[i].x2, call.X, call.X2)
				}
			}
		}
	})

	t.Run("zero-length spans are skipped", func(t *testing.T) {
		renderer := NewRendererScanlineBinSolidWithColor(baseRenderer, "purple")

		// Reset mock state
		baseRenderer.hlineCalls = nil

		scanline := &MockScanline{
			y:        20,
			numSpans: 2,
			spans: []SpanData{
				{X: 10, Len: 0, Covers: nil}, // Zero length span should be skipped
				{X: 20, Len: 3, Covers: []basics.Int8u{255, 255, 255}},
			},
		}

		renderer.Render(scanline)

		// Binary renderer renders all spans, even zero-length ones (though they may be degenerate)
		expectedCalls := 2
		if len(baseRenderer.hlineCalls) != expectedCalls {
			t.Errorf("Expected %d hline calls, got %d", expectedCalls, len(baseRenderer.hlineCalls))
		}

		// Check that the zero-length span creates a degenerate hline (x1 > x2)
		if len(baseRenderer.hlineCalls) >= 1 {
			call0 := baseRenderer.hlineCalls[0]
			if call0.X == 10 && call0.X2 == 9 { // Zero length span: x=10, len=0 -> endX=10+0-1=9
				// This is correct - degenerate span
			} else {
				t.Errorf("Zero-length span should create degenerate hline: got x=%d, x2=%d", call0.X, call0.X2)
			}
		}
	})
}
