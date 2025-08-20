package agg

import (
	"testing"
)

// TestNewAgg2D verifies that NewAgg2D creates a properly initialized context
func TestNewAgg2D(t *testing.T) {
	ctx := NewAgg2D()

	if ctx == nil {
		t.Fatal("NewAgg2D should not return nil")
	}

	// Verify default values
	if ctx.lineWidth != 1.0 {
		t.Errorf("Expected default line width 1.0, got %v", ctx.lineWidth)
	}

	if ctx.lineCap != CapRound {
		t.Errorf("Expected default line cap CapRound, got %v", ctx.lineCap)
	}

	if ctx.lineJoin != JoinRound {
		t.Errorf("Expected default line join JoinRound, got %v", ctx.lineJoin)
	}

	if ctx.fillColor != White {
		t.Errorf("Expected default fill color White, got %v", ctx.fillColor)
	}

	if ctx.lineColor != Black {
		t.Errorf("Expected default line color Black, got %v", ctx.lineColor)
	}
}

// TestAttach verifies buffer attachment
func TestAttach(t *testing.T) {
	ctx := NewAgg2D()

	// Create a test buffer (RGBA format, 100x100)
	width, height := 100, 100
	stride := width * 4 // 4 bytes per pixel for RGBA
	buffer := make([]uint8, height*stride)

	ctx.Attach(buffer, width, height, stride)

	// Verify buffer attachment
	if ctx.rbuf == nil {
		t.Error("Buffer should be attached")
	}

	if ctx.rbuf.Width() != width {
		t.Errorf("Expected width %d, got %d", width, ctx.rbuf.Width())
	}

	if ctx.rbuf.Height() != height {
		t.Errorf("Expected height %d, got %d", height, ctx.rbuf.Height())
	}

	if ctx.rbuf.Stride() != stride {
		t.Errorf("Expected stride %d, got %d", stride, ctx.rbuf.Stride())
	}
}

// TestColorSetters verifies color setting methods
func TestColorSetters(t *testing.T) {
	ctx := NewAgg2D()

	// Test FillColor
	red := NewColorRGB(255, 0, 0)
	ctx.FillColor(red)
	if ctx.fillColor != red {
		t.Errorf("Expected fill color %v, got %v", red, ctx.fillColor)
	}

	// Test LineColor
	blue := NewColorRGB(0, 0, 255)
	ctx.LineColor(blue)
	if ctx.lineColor != blue {
		t.Errorf("Expected line color %v, got %v", blue, ctx.lineColor)
	}

	// Test RGBA variants
	ctx.FillColorRGBA(128, 64, 32, 255)
	expected := NewColor(128, 64, 32, 255)
	if ctx.fillColor != expected {
		t.Errorf("Expected fill color %v, got %v", expected, ctx.fillColor)
	}

	ctx.LineColorRGBA(64, 128, 192, 128)
	expected = NewColor(64, 128, 192, 128)
	if ctx.lineColor != expected {
		t.Errorf("Expected line color %v, got %v", expected, ctx.lineColor)
	}
}

// TestLineAttributes verifies line attribute setters
func TestLineAttributes(t *testing.T) {
	ctx := NewAgg2D()

	// Test LineWidth
	ctx.LineWidth(2.5)
	if ctx.lineWidth != 2.5 {
		t.Errorf("Expected line width 2.5, got %v", ctx.lineWidth)
	}

	// Test LineCap
	ctx.LineCap(CapSquare)
	if ctx.lineCap != CapSquare {
		t.Errorf("Expected line cap CapSquare, got %v", ctx.lineCap)
	}

	// Test LineJoin
	ctx.LineJoin(JoinBevel)
	if ctx.lineJoin != JoinBevel {
		t.Errorf("Expected line join JoinBevel, got %v", ctx.lineJoin)
	}
}

// TestTransformations verifies transformation methods
func TestTransformations(t *testing.T) {
	ctx := NewAgg2D()

	// Create test buffer for transformations to work
	width, height := 100, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)
	ctx.Attach(buffer, width, height, stride)

	// Test coordinate transformation
	x, y := 10.0, 20.0
	originalX, originalY := x, y

	ctx.WorldToScreen(&x, &y)
	// For identity transform, coordinates should remain the same
	if x != originalX || y != originalY {
		t.Errorf("Expected coordinates (%v, %v), got (%v, %v)", originalX, originalY, x, y)
	}

	ctx.ScreenToWorld(&x, &y)
	// Should get back to original coordinates
	if x != originalX || y != originalY {
		t.Errorf("Expected coordinates (%v, %v), got (%v, %v)", originalX, originalY, x, y)
	}
}

// TestPathCommands verifies path manipulation commands
func TestPathCommands(t *testing.T) {
	ctx := NewAgg2D()

	// Test basic path commands - these should not panic
	ctx.ResetPath()
	ctx.MoveTo(10, 20)
	ctx.LineTo(30, 40)
	ctx.MoveRel(5, 5)
	ctx.LineRel(10, 10)
	ctx.HorLineTo(50)
	ctx.HorLineRel(5)
	ctx.VerLineTo(60)
	ctx.VerLineRel(5)

	// Test curve commands
	ctx.QuadricCurveTo(70, 80, 90, 100)
	ctx.QuadricCurveRel(10, 10, 20, 20)
	ctx.CubicCurveTo(110, 120, 130, 140, 150, 160)
	ctx.CubicCurveRel(10, 10, 20, 20, 30, 30)

	// Test arc commands
	ctx.ArcTo(50, 50, 0, false, true, 100, 100)
	ctx.ArcRel(25, 25, 45, true, false, 50, 50)

	// Test close polygon
	ctx.ClosePolygon()

	// If we get here without panicking, the basic path commands work
}

// TestAddEllipse verifies ellipse addition
func TestAddEllipse(t *testing.T) {
	ctx := NewAgg2D()

	// Test ellipse addition - should not panic
	ctx.AddEllipse(100, 100, 50, 30, CW)
	ctx.AddEllipse(200, 200, 75, 75, CCW)

	// If we get here without panicking, ellipse addition works
}

// TestDrawPath verifies path drawing
func TestDrawPath(t *testing.T) {
	ctx := NewAgg2D()

	// Create test buffer
	width, height := 100, 100
	stride := width * 4
	buffer := make([]uint8, height*stride)
	ctx.Attach(buffer, width, height, stride)

	// Create a simple path
	ctx.ResetPath()
	ctx.MoveTo(10, 10)
	ctx.LineTo(50, 10)
	ctx.LineTo(50, 50)
	ctx.LineTo(10, 50)
	ctx.ClosePolygon()

	// Test different draw modes - should not panic
	ctx.DrawPath(FillOnly)
	ctx.DrawPath(StrokeOnly)
	ctx.DrawPath(FillAndStroke)
	ctx.DrawPath(FillWithLineColor)

	// Test no-transform version
	ctx.DrawPathNoTransform(FillOnly)

	// If we get here without panicking, path drawing works
}

// TestBasicShapes verifies basic shape drawing
func TestBasicShapes(t *testing.T) {
	ctx := NewAgg2D()

	// Create test buffer
	width, height := 200, 200
	stride := width * 4
	buffer := make([]uint8, height*stride)
	ctx.Attach(buffer, width, height, stride)

	// Test all basic shapes - should not panic
	ctx.Line(10, 10, 50, 50)
	ctx.Triangle(60, 10, 80, 10, 70, 30)
	ctx.Rectangle(90, 10, 130, 50)
	ctx.RoundedRect(10, 60, 50, 100, 5)
	ctx.RoundedRectVariableRadii(60, 60, 100, 100, 5, 10, 15, 20)
	ctx.Ellipse(140, 80, 20, 30)
	ctx.Ellipse(170, 80, 15, 15) // Circle is just an ellipse with equal radii

	// Test arc
	ctx.Arc(50, 150, 30, 30, 0, Pi/2)

	// Test star
	ctx.Star(100, 150, 15, 25, 0, 5)

	// Test curve
	ctx.Curve(10, 180, 30, 160, 50, 180)

	// Test polygon
	points := []float64{100, 180, 120, 160, 140, 180, 130, 200, 110, 200}
	ctx.Polygon(points, 5)

	// Test polyline
	ctx.Polyline(points, 5)

	// If we get here without panicking, all basic shapes work
}

// TestClipBox verifies clipping functionality
func TestClipBox(t *testing.T) {
	ctx := NewAgg2D()

	// Set clip box
	ctx.ClipBox(10, 20, 100, 200)

	// Get clip box and verify
	clipBox := ctx.GetClipBox()
	expected := RectD{10, 20, 100, 200}

	if clipBox != expected {
		t.Errorf("Expected clip box %v, got %v", expected, clipBox)
	}
}

// TestClearAll verifies buffer clearing
func TestClearAll(t *testing.T) {
	ctx := NewAgg2D()

	// Create test buffer
	width, height := 10, 10
	stride := width * 4
	buffer := make([]uint8, height*stride)

	// Fill with non-zero values first
	for i := range buffer {
		buffer[i] = 128
	}

	ctx.Attach(buffer, width, height, stride)

	// Clear with red
	red := NewColorRGB(255, 0, 0)
	ctx.ClearAll(red)

	// Verify first pixel is red (RGBA format)
	if buffer[0] != 255 || buffer[1] != 0 || buffer[2] != 0 || buffer[3] != 255 {
		t.Errorf("Expected first pixel to be red (255,0,0,255), got (%d,%d,%d,%d)",
			buffer[0], buffer[1], buffer[2], buffer[3])
	}

	// Test RGBA variant
	ctx.ClearAllRGBA(0, 255, 0, 128)

	// Verify first pixel is green with alpha
	if buffer[0] != 0 || buffer[1] != 255 || buffer[2] != 0 || buffer[3] != 128 {
		t.Errorf("Expected first pixel to be green with alpha (0,255,0,128), got (%d,%d,%d,%d)",
			buffer[0], buffer[1], buffer[2], buffer[3])
	}
}

// TestImageFilter verifies image filter settings
func TestImageFilter(t *testing.T) {
	ctx := NewAgg2D()

	ctx.ImageFilter(Bilinear)
	if ctx.imageFilter != Bilinear {
		t.Errorf("Expected image filter Bilinear, got %v", ctx.imageFilter)
	}

	ctx.ImageFilter(NoFilter)
	if ctx.imageFilter != NoFilter {
		t.Errorf("Expected image filter NoFilter, got %v", ctx.imageFilter)
	}
}

// TestImageResample verifies image resample settings
func TestImageResample(t *testing.T) {
	ctx := NewAgg2D()

	ctx.ImageResample(NoResample)
	if ctx.imageResample != NoResample {
		t.Errorf("Expected image resample NoResample, got %v", ctx.imageResample)
	}
}

// TestTextAlignment verifies text alignment settings
func TestTextAlignment(t *testing.T) {
	ctx := NewAgg2D()

	ctx.TextAlignment(AlignCenter, AlignTop)

	if ctx.textAlignX != AlignCenter {
		t.Errorf("Expected text align X AlignCenter, got %v", ctx.textAlignX)
	}

	if ctx.textAlignY != AlignTop {
		t.Errorf("Expected text align Y AlignTop, got %v", ctx.textAlignY)
	}
}

// BenchmarkPathOperations benchmarks basic path operations
func BenchmarkPathOperations(b *testing.B) {
	ctx := NewAgg2D()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.ResetPath()
		ctx.MoveTo(float64(i%100), float64(i%100))
		ctx.LineTo(float64(i%100+50), float64(i%100+50))
		ctx.ClosePolygon()
	}
}

// BenchmarkShapeDrawing benchmarks shape drawing
func BenchmarkShapeDrawing(b *testing.B) {
	ctx := NewAgg2D()

	// Create test buffer
	width, height := 200, 200
	stride := width * 4
	buffer := make([]uint8, height*stride)
	ctx.Attach(buffer, width, height, stride)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := float64(i % 150)
		y := float64(i % 150)
		ctx.Ellipse(x, y, 10, 10) // Circle is just an ellipse with equal radii
	}
}
