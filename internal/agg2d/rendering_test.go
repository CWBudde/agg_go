package agg2d

import (
	"testing"
)

// TestRenderingComponents verifies that rendering components are initialized
func TestRenderingComponents(t *testing.T) {
	agg2d := NewAgg2D()

	// Create a test buffer
	width, height := 100, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)

	// Attach buffer - this should initialize rendering components
	agg2d.Attach(buffer, width, height, stride)

	// Verify rendering components are initialized
	if agg2d.rasterizer == nil {
		t.Error("Expected rasterizer to be initialized")
	}

	if agg2d.scanline == nil {
		t.Error("Expected scanline to be initialized")
	}

	if agg2d.pixfmt == nil {
		t.Error("Expected pixfmt to be initialized")
	}

	// Test ClearAll functionality
	red := NewColorRGB(255, 0, 0)
	agg2d.ClearAll(red)

	// Verify buffer is filled with red
	for i := 0; i < len(buffer); i += 4 {
		if buffer[i] != 255 || buffer[i+1] != 0 || buffer[i+2] != 0 || buffer[i+3] != 255 {
			t.Errorf("Expected red pixel at offset %d, got RGBA(%d, %d, %d, %d)",
				i, buffer[i], buffer[i+1], buffer[i+2], buffer[i+3])
		}
	}
}

// TestBasicDrawing tests basic drawing functionality
func TestBasicDrawing(t *testing.T) {
	agg2d := NewAgg2D()

	// Create a test buffer
	width, height := 100, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)

	agg2d.Attach(buffer, width, height, stride)

	// Clear with white background
	agg2d.ClearAll(White)

	// Set fill color to blue
	agg2d.FillColor(Blue)

	// Draw a simple rectangle and verify the basic API path runs cleanly.
	agg2d.Rectangle(10, 10, 50, 50)
	agg2d.DrawPath(FillOnly)

	r, g, b, a := pixelAt(buffer, width, 20, 20)
	if r != 0 || g != 0 || b != 255 || a != 255 {
		t.Fatalf("inside rectangle pixel = (%d,%d,%d,%d), want solid blue", r, g, b, a)
	}

	r, g, b, a = pixelAt(buffer, width, 5, 5)
	if r != 255 || g != 255 || b != 255 || a != 255 {
		t.Fatalf("outside rectangle pixel = (%d,%d,%d,%d), want white background", r, g, b, a)
	}
}

// TestGradientAPI tests gradient API functionality
func TestGradientAPI(t *testing.T) {
	agg2d := NewAgg2D()

	// Create a test buffer
	width, height := 100, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)

	agg2d.Attach(buffer, width, height, stride)
	agg2d.ClearAll(White)

	// Set up a gradient and exercise the draw path end-to-end.
	agg2d.FillLinearGradient(0, 0, 100, 100, Red, Blue, 1.0)

	// Draw a shape with gradient
	agg2d.Rectangle(20, 20, 80, 80)
	agg2d.DrawPath(FillOnly)

	r1, g1, b1, a1 := pixelAt(buffer, width, 30, 30)
	r2, g2, b2, a2 := pixelAt(buffer, width, 70, 70)
	if a1 == 0 || a2 == 0 {
		t.Fatalf("expected gradient-filled pixels to be opaque enough, got a1=%d a2=%d", a1, a2)
	}
	if !(r1 > b1) {
		t.Fatalf("expected gradient near start to skew red, got (%d,%d,%d,%d)", r1, g1, b1, a1)
	}
	if !(b2 > r2) {
		t.Fatalf("expected gradient near end to skew blue, got (%d,%d,%d,%d)", r2, g2, b2, a2)
	}

	r, g, b, a := pixelAt(buffer, width, 10, 10)
	if r != 255 || g != 255 || b != 255 || a != 255 {
		t.Fatalf("outside gradient shape pixel = (%d,%d,%d,%d), want untouched white", r, g, b, a)
	}
}

// TestRenderingTransformations tests transformation functionality
func TestRenderingTransformations(t *testing.T) {
	agg2d := NewAgg2D()

	width, height := 80, 80
	buffer := make([]uint8, width*height*4)
	agg2d.Attach(buffer, width, height, width*4)
	agg2d.ClearAll(White)
	agg2d.FillColor(Black)

	// Test basic transformations
	agg2d.Translate(10, 20)
	agg2d.Rectangle(0, 0, 10, 10)
	agg2d.DrawPath(FillOnly)

	// Test transformation stack
	agg2d.PushTransform()
	agg2d.Scale(0.5, 0.5)

	success := agg2d.PopTransform()
	if !success {
		t.Error("Expected PopTransform to succeed")
	}

	r, g, b, a := pixelAt(buffer, width, 15, 25)
	if r != 0 || g != 0 || b != 0 || a != 255 {
		t.Fatalf("translated rectangle pixel = (%d,%d,%d,%d), want black", r, g, b, a)
	}

	r, g, b, a = pixelAt(buffer, width, 5, 5)
	if r != 255 || g != 255 || b != 255 || a != 255 {
		t.Fatalf("untranslated origin pixel = (%d,%d,%d,%d), want white", r, g, b, a)
	}

	// Test coordinate transformation
	x, y := 100.0, 200.0
	agg2d.WorldToScreen(&x, &y)
	if x == 100.0 && y == 200.0 {
		t.Fatal("expected WorldToScreen to apply the active transform")
	}
}

func TestClipBoxClipsFilledPaths(t *testing.T) {
	agg2d := NewAgg2D()

	width, height := 32, 32
	buffer := make([]uint8, width*height*4)
	agg2d.Attach(buffer, width, height, width*4)
	agg2d.ClearAll(White)
	agg2d.ClipBox(10, 10, 20, 20)
	agg2d.FillColor(Red)

	agg2d.Rectangle(4, 4, 24, 24)
	agg2d.DrawPath(FillOnly)

	r, g, b, a := pixelAt(buffer, width, 15, 15)
	if r != 255 || g != 0 || b != 0 || a != 255 {
		t.Fatalf("filled pixel inside clip = (%d,%d,%d,%d), want solid red", r, g, b, a)
	}

	r, g, b, a = pixelAt(buffer, width, 8, 15)
	if r != 255 || g != 255 || b != 255 || a != 255 {
		t.Fatalf("filled pixel outside clip = (%d,%d,%d,%d), want white background", r, g, b, a)
	}

	r, g, b, a = pixelAt(buffer, width, 15, 8)
	if r != 255 || g != 255 || b != 255 || a != 255 {
		t.Fatalf("filled pixel above clip = (%d,%d,%d,%d), want white background", r, g, b, a)
	}
}

func TestClipBoxClipsStrokedPaths(t *testing.T) {
	agg2d := NewAgg2D()

	width, height := 32, 32
	buffer := make([]uint8, width*height*4)
	agg2d.Attach(buffer, width, height, width*4)
	agg2d.ClearAll(White)
	agg2d.ClipBox(10, 10, 20, 20)
	agg2d.NoFill()
	agg2d.LineColor(Black)
	agg2d.LineWidth(1.0)

	agg2d.ResetPath()
	agg2d.MoveTo(4, 15)
	agg2d.LineTo(24, 15)
	agg2d.DrawPath(StrokeOnly)

	r, g, b, a := pixelAt(buffer, width, 15, 15)
	if !(r < 255 || g < 255 || b < 255) || a == 0 {
		t.Fatalf("stroked pixel inside clip = (%d,%d,%d,%d), want visible clipped stroke coverage", r, g, b, a)
	}

	r, g, b, a = pixelAt(buffer, width, 8, 15)
	if r != 255 || g != 255 || b != 255 || a != 255 {
		t.Fatalf("stroked pixel left of clip = (%d,%d,%d,%d), want white background", r, g, b, a)
	}

	r, g, b, a = pixelAt(buffer, width, 22, 15)
	if r != 255 || g != 255 || b != 255 || a != 255 {
		t.Fatalf("stroked pixel outside clipped stroke extent = (%d,%d,%d,%d), want white background", r, g, b, a)
	}
}

func TestBlendModesAffectRenderedPaths(t *testing.T) {
	tests := []struct {
		name string
		mode BlendMode
		want [4]uint8
	}{
		{name: "alpha", mode: BlendAlpha, want: [4]uint8{255, 0, 0, 255}},
		{name: "multiply", mode: BlendMultiply, want: [4]uint8{0, 0, 0, 255}},
		{name: "screen", mode: BlendScreen, want: [4]uint8{255, 255, 0, 255}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg2d := NewAgg2D()
			width, height := 16, 16
			buffer := make([]uint8, width*height*4)
			agg2d.Attach(buffer, width, height, width*4)
			agg2d.ClearAll(Color{0, 255, 0, 255})
			agg2d.SetBlendMode(tt.mode)
			agg2d.FillColor(Color{255, 0, 0, 255})

			agg2d.Rectangle(2, 2, 14, 14)
			agg2d.DrawPath(FillOnly)

			r, g, b, a := pixelAt(buffer, width, 8, 8)
			got := [4]uint8{r, g, b, a}
			if got != tt.want {
				t.Fatalf("blend mode %v pixel = %v, want %v", tt.mode, got, tt.want)
			}
		})
	}
}

func TestBlendDstPreservesDestinationForRenderedPaths(t *testing.T) {
	agg2d := NewAgg2D()

	width, height := 16, 16
	buffer := make([]uint8, width*height*4)
	agg2d.Attach(buffer, width, height, width*4)
	agg2d.ClearAll(Color{10, 20, 30, 255})
	agg2d.SetBlendMode(BlendDst)
	agg2d.FillColor(Color{255, 0, 0, 255})

	agg2d.Rectangle(2, 2, 14, 14)
	agg2d.DrawPath(FillOnly)

	r, g, b, a := pixelAt(buffer, width, 8, 8)
	if r != 10 || g != 20 || b != 30 || a != 255 {
		t.Fatalf("BlendDst pixel = (%d,%d,%d,%d), want unchanged destination", r, g, b, a)
	}
}
