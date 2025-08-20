package scanline

import (
	"testing"
)

func TestRendererScanlineAASolid(t *testing.T) {
	baseRenderer := &MockBaseRenderer{}

	t.Run("creation and basic operations", func(t *testing.T) {
		renderer := NewRendererScanlineAASolid[*MockBaseRenderer]()

		// Test attachment
		renderer.Attach(baseRenderer)
		if renderer.BaseRenderer() != baseRenderer {
			t.Error("Base renderer not attached correctly")
		}

		// Test color setting
		color := "red"
		renderer.SetColor(color)
		if renderer.Color() != color {
			t.Errorf("Expected color %v, got %v", color, renderer.Color())
		}
	})

	t.Run("creation with renderer", func(t *testing.T) {
		renderer := NewRendererScanlineAASolidWithRenderer(baseRenderer)
		if renderer.BaseRenderer() != baseRenderer {
			t.Error("Base renderer not set correctly in constructor")
		}
	})

	t.Run("prepare does nothing", func(t *testing.T) {
		renderer := NewRendererScanlineAASolid[*MockBaseRenderer]()
		// Should not panic
		renderer.Prepare()
	})

	t.Run("render calls underlying function", func(t *testing.T) {
		renderer := NewRendererScanlineAASolidWithRenderer(baseRenderer)
		color := "blue"
		renderer.SetColor(color)

		scanline := &MockScanline{
			y:        5,
			numSpans: 1,
			spans: []SpanData{
				{X: 10, Len: 5, Covers: []uint8{255, 200, 150, 100, 50}},
			},
		}

		// Clear previous calls
		baseRenderer.solidHspanCalls = nil

		renderer.Render(scanline)

		if len(baseRenderer.solidHspanCalls) != 1 {
			t.Errorf("Expected 1 solid hspan call, got %d", len(baseRenderer.solidHspanCalls))
		}

		call := baseRenderer.solidHspanCalls[0]
		if call.Color != color {
			t.Errorf("Expected color %v, got %v", color, call.Color)
		}
	})
}
