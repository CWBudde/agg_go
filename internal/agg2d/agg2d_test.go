package agg2d

import (
	"math"
	"testing"

	"agg_go/internal/basics"
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

func TestInitializeDashingPreservesStrokeAttributes(t *testing.T) {
	ctx := NewAgg2D()

	ctx.LineWidth(7.0)
	ctx.LineCap(CapSquare)
	ctx.LineJoin(JoinBevel)
	ctx.MiterLimit(6.5)
	ctx.InnerMiterLimit(1.75)
	ctx.Shorten(2.25)
	ctx.ApproximationScale(3.5)
	ctx.AddDash(10, 5)

	if ctx.convDash == nil {
		t.Fatal("expected AddDash to initialize convDash")
	}
	if got := ctx.convStroke.Width(); got != 7.0 {
		t.Fatalf("stroke width = %f, want 7.0", got)
	}
	if got := ctx.convStroke.LineCap(); got != basics.SquareCap {
		t.Fatalf("line cap = %v, want %v", got, basics.SquareCap)
	}
	if got := ctx.convStroke.LineJoin(); got != basics.BevelJoin {
		t.Fatalf("line join = %v, want %v", got, basics.BevelJoin)
	}
	if got := ctx.convStroke.MiterLimit(); math.Abs(got-6.5) > 1e-10 {
		t.Fatalf("miter limit = %f, want 6.5", got)
	}
	if got := ctx.convStroke.InnerMiterLimit(); math.Abs(got-1.75) > 1e-10 {
		t.Fatalf("inner miter limit = %f, want 1.75", got)
	}
	if got := ctx.convStroke.Shorten(); math.Abs(got-2.25) > 1e-10 {
		t.Fatalf("shorten = %f, want 2.25", got)
	}
	if got := ctx.convStroke.ApproximationScale(); math.Abs(got-3.5) > 1e-10 {
		t.Fatalf("approximation scale = %f, want 3.5", got)
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

	if got := ctx.path.TotalVertices(); got <= 8 {
		t.Fatalf("expected curve and arc commands to add vertices, got %d", got)
	}

	wantPrefix := []struct {
		x, y float64
		cmd  basics.PathCommand
	}{
		{10, 20, basics.PathCmdMoveTo},
		{30, 40, basics.PathCmdLineTo},
		{35, 45, basics.PathCmdMoveTo},
		{45, 55, basics.PathCmdLineTo},
		{50, 55, basics.PathCmdLineTo},
		{55, 55, basics.PathCmdLineTo},
		{55, 60, basics.PathCmdLineTo},
		{55, 65, basics.PathCmdLineTo},
	}
	for i, want := range wantPrefix {
		x, y, cmd := ctx.path.Vertex(uint(i))
		if x != want.x || y != want.y || basics.PathCommand(cmd) != want.cmd {
			t.Fatalf("vertex %d: got (%v,%v,%v), want (%v,%v,%v)", i, x, y, basics.PathCommand(cmd), want.x, want.y, want.cmd)
		}
	}

	lastX, lastY, lastCmd := ctx.path.Vertex(ctx.path.TotalVertices() - 1)
	if !basics.IsEndPoly(basics.PathCommand(lastCmd)) {
		t.Fatalf("expected ClosePolygon to append EndPoly, got %v at (%v,%v)", basics.PathCommand(lastCmd), lastX, lastY)
	}
	if !ctx.hasLastCtrl {
		t.Fatal("expected curve commands to leave control-point tracking enabled")
	}
}

// TestAddEllipse verifies ellipse addition
func TestAddEllipse(t *testing.T) {
	ctx := NewAgg2D()

	ctx.AddEllipse(100, 100, 50, 30, CW)
	cwArea := signedPathArea(ctx.path)
	if got := ctx.path.TotalVertices(); got < 8 {
		t.Fatalf("expected ellipse path to contain multiple vertices, got %d", got)
	}
	if cwArea >= 0 {
		t.Fatalf("expected CW ellipse to have negative signed area, got %v", cwArea)
	}
	if _, _, cmd := ctx.path.Vertex(ctx.path.TotalVertices() - 1); !basics.IsEndPoly(basics.PathCommand(cmd)) {
		t.Fatalf("expected CW ellipse path to end with EndPoly, got %v", basics.PathCommand(cmd))
	}

	ctx.ResetPath()
	ctx.AddEllipse(200, 200, 75, 75, CCW)
	ccwArea := signedPathArea(ctx.path)
	if ccwArea <= 0 {
		t.Fatalf("expected CCW ellipse to have positive signed area, got %v", ccwArea)
	}
	if got := math.Abs(ccwArea); got == 0 {
		t.Fatal("expected CCW ellipse to enclose non-zero area")
	}
}

func signedPathArea(ps interface {
	TotalVertices() uint
	Vertex(idx uint) (x, y float64, cmd uint32)
},
) float64 {
	points := make([][2]float64, 0, ps.TotalVertices())
	for i := uint(0); i < ps.TotalVertices(); i++ {
		x, y, cmd := ps.Vertex(i)
		if basics.IsVertex(basics.PathCommand(cmd)) {
			points = append(points, [2]float64{x, y})
		}
	}
	if len(points) < 3 {
		return 0
	}

	area := 0.0
	for i := range points {
		j := (i + 1) % len(points)
		area += points[i][0]*points[j][1] - points[j][0]*points[i][1]
	}
	return area / 2
}

// hasNonWhiteIn checks that at least one pixel in the bounding box is not white (255,255,255,255)
// or zero (0,0,0,0). This confirms that rendering actually produced visible output.
// pixelAt is defined in text_rendering_bitmap_test.go.
func hasNonWhiteIn(buf []uint8, width, x1, y1, x2, y2 int) bool {
	for y := y1; y < y2; y++ {
		for x := x1; x < x2; x++ {
			r, g, b, a := pixelAt(buf, width, x, y)
			if a != 0 && (r != 255 || g != 255 || b != 255 || a != 255) {
				return true
			}
		}
	}
	return false
}

// TestDrawPath verifies path drawing produces visible pixel output for each draw mode.
func TestDrawPath(t *testing.T) {
	width, height := 100, 100
	stride := width * 4

	drawAndCheck := func(t *testing.T, mode DrawPathFlag) []uint8 {
		t.Helper()
		ctx := NewAgg2D()
		buf := make([]uint8, height*stride)
		// Fill with white so we can detect rendered pixels.
		for i := 0; i < len(buf); i += 4 {
			buf[i], buf[i+1], buf[i+2], buf[i+3] = 255, 255, 255, 255
		}
		ctx.Attach(buf, width, height, stride)
		ctx.FillColor(NewColorRGB(255, 0, 0))
		ctx.LineColor(NewColorRGB(0, 0, 255))
		ctx.LineWidth(2.0)
		ctx.ResetPath()
		ctx.MoveTo(10, 10)
		ctx.LineTo(50, 10)
		ctx.LineTo(50, 50)
		ctx.LineTo(10, 50)
		ctx.ClosePolygon()
		ctx.DrawPath(mode)
		return buf
	}

	t.Run("FillOnly", func(t *testing.T) {
		buf := drawAndCheck(t, FillOnly)
		r, g, b, a := pixelAt(buf, width, 30, 30)
		if r != 255 || g != 0 || b != 0 || a != 255 {
			t.Fatalf("FillOnly interior pixel = (%d,%d,%d,%d), want solid red", r, g, b, a)
		}
	})

	t.Run("StrokeOnly", func(t *testing.T) {
		buf := drawAndCheck(t, StrokeOnly)
		if !hasNonWhiteIn(buf, width, 9, 9, 12, 12) {
			t.Fatal("StrokeOnly should produce visible stroke pixels near corner")
		}
		// Interior should remain white (no fill).
		r, g, b, a := pixelAt(buf, width, 30, 30)
		if r != 255 || g != 255 || b != 255 || a != 255 {
			t.Fatalf("StrokeOnly interior pixel = (%d,%d,%d,%d), want white", r, g, b, a)
		}
	})

	t.Run("FillAndStroke", func(t *testing.T) {
		buf := drawAndCheck(t, FillAndStroke)
		// Interior should be filled.
		if !hasNonWhiteIn(buf, width, 25, 25, 35, 35) {
			t.Fatal("FillAndStroke should produce visible fill pixels")
		}
	})

	t.Run("FillWithLineColor", func(t *testing.T) {
		buf := drawAndCheck(t, FillWithLineColor)
		r, g, b, a := pixelAt(buf, width, 30, 30)
		// Should be filled with line color (blue).
		if b == 255 && r == 255 && g == 255 {
			t.Fatalf("FillWithLineColor interior should not be white, got (%d,%d,%d,%d)", r, g, b, a)
		}
		_ = a
	})
}

// TestBasicShapes verifies that each shape primitive produces visible pixels.
func TestBasicShapes(t *testing.T) {
	width, height := 300, 300
	stride := width * 4

	type shapeCase struct {
		name string
		draw func(ctx *Agg2D)
		// Bounding box to check for non-white pixels.
		x1, y1, x2, y2 int
	}

	cases := []shapeCase{
		{"Line", func(c *Agg2D) { c.Line(10, 10, 50, 50) }, 10, 10, 50, 50},
		{"Triangle", func(c *Agg2D) { c.Triangle(60, 10, 80, 10, 70, 30) }, 60, 10, 81, 31},
		{"Rectangle", func(c *Agg2D) { c.Rectangle(90, 10, 130, 50) }, 90, 10, 131, 51},
		{"RoundedRect", func(c *Agg2D) { c.RoundedRect(10, 60, 50, 100, 5) }, 10, 60, 51, 101},
		{"RoundedRectVar", func(c *Agg2D) { c.RoundedRectVariableRadii(60, 60, 100, 100, 5, 10, 15, 20) }, 60, 60, 101, 101},
		{"Ellipse", func(c *Agg2D) { c.Ellipse(160, 80, 20, 30) }, 140, 50, 181, 111},
		{"Arc", func(c *Agg2D) { c.Arc(50, 170, 30, 30, 0, Pi/2) }, 45, 165, 81, 201},
		{"Star", func(c *Agg2D) { c.Star(120, 170, 15, 25, 0, 5) }, 95, 145, 146, 196},
		{"Curve", func(c *Agg2D) { c.Curve(10, 230, 30, 210, 50, 230) }, 10, 210, 51, 231},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := NewAgg2D()
			buf := make([]uint8, height*stride)
			for i := 0; i < len(buf); i += 4 {
				buf[i], buf[i+1], buf[i+2], buf[i+3] = 255, 255, 255, 255
			}
			ctx.Attach(buf, width, height, stride)
			ctx.FillColor(NewColorRGB(255, 0, 0))
			ctx.LineColor(NewColorRGB(0, 0, 255))
			ctx.LineWidth(1.0)
			tc.draw(ctx)
			if !hasNonWhiteIn(buf, width, tc.x1, tc.y1, tc.x2, tc.y2) {
				t.Fatalf("%s did not produce visible pixels in (%d,%d)-(%d,%d)", tc.name, tc.x1, tc.y1, tc.x2, tc.y2)
			}
		})
	}
}

// TestClipBox verifies clipping functionality
func TestClipBox(t *testing.T) {
	ctx := NewAgg2D()

	// Set clip box
	ctx.ClipBox(10, 20, 100, 200)

	// Get clip box and verify
	x1, y1, x2, y2 := ctx.GetClipBox()
	expected := RectD{10, 20, 100, 200}
	actual := RectD{x1, y1, x2, y2}

	if actual != expected {
		t.Errorf("Expected clip box %v, got %v", expected, actual)
	}
}

func TestFillEvenOddUpdatesRasterizerImmediately(t *testing.T) {
	ctx := NewAgg2D()

	width, height := 10, 10
	stride := width * 4
	buffer := make([]uint8, height*stride)
	ctx.Attach(buffer, width, height, stride)

	ctx.FillEvenOdd(true)
	if ctx.rasterizer.GetFillingRule() != basics.FillEvenOdd {
		t.Fatalf("expected rasterizer fill rule %v, got %v", basics.FillEvenOdd, ctx.rasterizer.GetFillingRule())
	}

	ctx.FillEvenOdd(false)
	if ctx.rasterizer.GetFillingRule() != basics.FillNonZero {
		t.Fatalf("expected rasterizer fill rule %v, got %v", basics.FillNonZero, ctx.rasterizer.GetFillingRule())
	}
}

func TestClipBoxPropagatesToRendererCopyOps(t *testing.T) {
	ctx := NewAgg2D()

	width, height := 6, 6
	stride := width * 4
	dst := make([]uint8, height*stride)
	ctx.Attach(dst, width, height, stride)

	srcW, srcH := 3, 3
	src := make([]uint8, srcW*srcH*4)
	for i := 0; i < len(src); i += 4 {
		src[i+0] = 255
		src[i+1] = 0
		src[i+2] = 0
		src[i+3] = 255
	}
	img := NewImage(src, srcW, srcH, srcW*4)

	ctx.ClipBox(1, 1, 2, 2)
	if err := ctx.CopyImageSimple(img, 0, 0); err != nil {
		t.Fatalf("CopyImageSimple failed: %v", err)
	}

	redPixels := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			offset := y*stride + x*4
			r, g, b, a := dst[offset], dst[offset+1], dst[offset+2], dst[offset+3]
			inClip := x >= 1 && x <= 2 && y >= 1 && y <= 2

			if inClip {
				if r != 255 || g != 0 || b != 0 || a != 255 {
					t.Fatalf("pixel (%d,%d) expected red in clip box, got rgba=(%d,%d,%d,%d)", x, y, r, g, b, a)
				}
				redPixels++
				continue
			}
			if r != 0 || g != 0 || b != 0 || a != 0 {
				t.Fatalf("pixel (%d,%d) expected untouched outside clip box, got rgba=(%d,%d,%d,%d)", x, y, r, g, b, a)
			}
		}
	}
	if redPixels != 4 {
		t.Fatalf("expected 4 copied red pixels in clip region, got %d", redPixels)
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

func TestClearAllIgnoresClipBox(t *testing.T) {
	ctx := NewAgg2D()

	width, height := 4, 4
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx.Attach(buffer, width, height, stride)
	ctx.ClipBox(1, 1, 2, 2)
	ctx.ClearAllRGBA(10, 20, 30, 40)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			offset := y*stride + x*4
			if buffer[offset] != 10 || buffer[offset+1] != 20 || buffer[offset+2] != 30 || buffer[offset+3] != 40 {
				t.Fatalf("pixel (%d,%d) = (%d,%d,%d,%d), want (10,20,30,40)", x, y,
					buffer[offset], buffer[offset+1], buffer[offset+2], buffer[offset+3])
			}
		}
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
