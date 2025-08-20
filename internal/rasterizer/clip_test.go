package rasterizer

import (
	"testing"

	"agg_go/internal/basics"
)

// MockRasterizer for testing
type MockRasterizer struct {
	Lines []Line
}

type Line struct {
	X1, Y1, X2, Y2 int
}

func (m *MockRasterizer) Line(x1, y1, x2, y2 int) {
	m.Lines = append(m.Lines, Line{X1: x1, Y1: y1, X2: x2, Y2: y2})
}

func (m *MockRasterizer) Reset() {
	m.Lines = nil
}

func TestRasConvInt(t *testing.T) {
	conv := RasConvInt{}

	// Test MulDiv
	result := conv.MulDiv(10.0, 20.0, 4.0)
	expected := 50.0
	if result != expected {
		t.Errorf("MulDiv(10, 20, 4) = %f, want %f", result, expected)
	}

	// Test Xi with int
	xi := conv.Xi(100)
	if xi != 100 {
		t.Errorf("Xi(100) = %d, want 100", xi)
	}

	// Test Xi with float64
	xi = conv.Xi(100.7)
	if xi != 101 {
		t.Errorf("Xi(100.7) = %d, want 101", xi)
	}

	// Test Yi
	yi := conv.Yi(50)
	if yi != 50 {
		t.Errorf("Yi(50) = %d, want 50", yi)
	}

	// Test Upscale
	upscaled := conv.Upscale(1.0).(int)
	expected_upscaled := basics.IRound(basics.PolySubpixelScale)
	if upscaled != expected_upscaled {
		t.Errorf("Upscale(1.0) = %d, want %d", upscaled, expected_upscaled)
	}

	// Test Downscale
	downscaled := conv.Downscale(256).(int)
	if downscaled != 256 {
		t.Errorf("Downscale(256) = %d, want 256", downscaled)
	}
}

func TestRasConvDbl(t *testing.T) {
	conv := RasConvDbl{}

	// Test MulDiv
	result := conv.MulDiv(10.0, 20.0, 4.0)
	expected := 50.0
	if result != expected {
		t.Errorf("MulDiv(10, 20, 4) = %f, want %f", result, expected)
	}

	// Test Xi
	xi := conv.Xi(1.0)
	expectedXi := basics.IRound(basics.PolySubpixelScale)
	if xi != expectedXi {
		t.Errorf("Xi(1.0) = %d, want %d", xi, expectedXi)
	}

	// Test Yi
	yi := conv.Yi(1.0)
	expectedYi := basics.IRound(basics.PolySubpixelScale)
	if yi != expectedYi {
		t.Errorf("Yi(1.0) = %d, want %d", yi, expectedYi)
	}

	// Test Upscale (pass-through for double)
	upscaled := conv.Upscale(42.5).(float64)
	if upscaled != 42.5 {
		t.Errorf("Upscale(42.5) = %f, want 42.5", upscaled)
	}

	// Test Downscale
	downscaled := conv.Downscale(256).(float64)
	expectedDownscaled := 256.0 / basics.PolySubpixelScale
	if downscaled != expectedDownscaled {
		t.Errorf("Downscale(256) = %f, want %f", downscaled, expectedDownscaled)
	}
}

func TestRasConvInt3x(t *testing.T) {
	conv := RasConvInt3x{}

	// Test Xi with 3x scaling
	xi := conv.Xi(100)
	if xi != 300 {
		t.Errorf("Xi(100) = %d, want 300", xi)
	}

	// Test Yi (no scaling)
	yi := conv.Yi(50)
	if yi != 50 {
		t.Errorf("Yi(50) = %d, want 50", yi)
	}
}

func TestRasterizerSlClipBasic(t *testing.T) {
	mock := &MockRasterizer{}
	clipper := NewRasterizerSlClip[RasConvInt]()

	// Test without clipping
	clipper.MoveTo(10, 20)
	clipper.LineTo(mock, 30, 40)

	if len(mock.Lines) != 1 {
		t.Errorf("Expected 1 line without clipping, got %d", len(mock.Lines))
		return
	}

	line := mock.Lines[0]
	if line.X1 != 10 || line.Y1 != 20 || line.X2 != 30 || line.Y2 != 40 {
		t.Errorf("Line without clipping: got (%d,%d)-(%d,%d), want (10,20)-(30,40)",
			line.X1, line.Y1, line.X2, line.Y2)
	}
}

func TestRasterizerSlClipWithClipping(t *testing.T) {
	mock := &MockRasterizer{}
	clipper := NewRasterizerSlClip[RasConvInt]()

	// Set clipping box
	clipper.ClipBox(10, 10, 50, 50)

	tests := []struct {
		name              string
		startX, startY    float64
		endX, endY        float64
		expectedLineCount int
		shouldHaveLines   bool
	}{
		{
			name:   "fully_visible",
			startX: 20, startY: 20,
			endX: 30, endY: 30,
			expectedLineCount: 1,
			shouldHaveLines:   true,
		},
		{
			name:   "fully_outside_left",
			startX: 0, startY: 20,
			endX: 5, endY: 30,
			expectedLineCount: 1, // AGG draws boundary line when both points are outside same boundary
			shouldHaveLines:   true,
		},
		{
			name:   "crosses_boundary",
			startX: 5, startY: 20,
			endX: 25, endY: 30,
			expectedLineCount: 2, // AGG draws boundary segment + visible segment for boundary crossings
			shouldHaveLines:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.Reset()
			clipper.MoveTo(tt.startX, tt.startY)
			clipper.LineTo(mock, tt.endX, tt.endY)

			if len(mock.Lines) != tt.expectedLineCount {
				t.Errorf("Expected %d lines, got %d", tt.expectedLineCount, len(mock.Lines))
			}

			if tt.shouldHaveLines && len(mock.Lines) > 0 {
				line := mock.Lines[0]
				// Basic sanity check - clipped line should be within bounds
				if line.X1 < 0 || line.X2 < 0 || line.Y1 < 0 || line.Y2 < 0 {
					t.Errorf("Clipped line has negative coordinates: (%d,%d)-(%d,%d)",
						line.X1, line.Y1, line.X2, line.Y2)
				}
			}
		})
	}
}

func TestRasterizerSlClipResetClipping(t *testing.T) {
	mock := &MockRasterizer{}
	clipper := NewRasterizerSlClip[RasConvInt]()

	// Set clipping and then reset it
	clipper.ClipBox(10, 10, 50, 50)
	clipper.ResetClipping()

	// Draw line that would be clipped if clipping was enabled
	clipper.MoveTo(0, 20)
	clipper.LineTo(mock, 5, 30)

	if len(mock.Lines) != 1 {
		t.Errorf("Expected 1 line after reset clipping, got %d", len(mock.Lines))
		return
	}

	line := mock.Lines[0]
	if line.X1 != 0 || line.Y1 != 20 || line.X2 != 5 || line.Y2 != 30 {
		t.Errorf("Line after reset: got (%d,%d)-(%d,%d), want (0,20)-(5,30)",
			line.X1, line.Y1, line.X2, line.Y2)
	}
}

func TestRasterizerSlClipNormalization(t *testing.T) {
	_ = &MockRasterizer{} // not used in this test
	clipper := NewRasterizerSlClip[RasConvInt]()

	// Set clipping box with reversed coordinates
	clipper.ClipBox(50, 50, 10, 10)

	// Check that the clip box was normalized
	if clipper.clipBox.X1 != 10 || clipper.clipBox.Y1 != 10 ||
		clipper.clipBox.X2 != 50 || clipper.clipBox.Y2 != 50 {
		t.Errorf("Clip box not normalized: got (%f,%f)-(%f,%f), want (10,10)-(50,50)",
			clipper.clipBox.X1, clipper.clipBox.Y1, clipper.clipBox.X2, clipper.clipBox.Y2)
	}
}

func TestRasterizerSlNoClip(t *testing.T) {
	mock := &MockRasterizer{}
	noClip := NewRasterizerSlNoClip(mock)

	// Test basic functionality
	noClip.MoveTo(10, 20)
	noClip.LineTo(mock, 30, 40)

	if len(mock.Lines) != 1 {
		t.Errorf("Expected 1 line, got %d", len(mock.Lines))
		return
	}

	line := mock.Lines[0]
	if line.X1 != 10 || line.Y1 != 20 || line.X2 != 30 || line.Y2 != 40 {
		t.Errorf("No-clip line: got (%d,%d)-(%d,%d), want (10,20)-(30,40)",
			line.X1, line.Y1, line.X2, line.Y2)
	}

	// Test that clipping methods are no-ops
	noClip.ResetClipping()         // should not panic
	noClip.ClipBox(0, 0, 100, 100) // should not affect anything

	// Draw another line to ensure no clipping occurred
	mock.Reset()
	noClip.MoveTo(-10, -20)
	noClip.LineTo(mock, 110, 120)

	if len(mock.Lines) != 1 {
		t.Errorf("Expected 1 line after no-clip operations, got %d", len(mock.Lines))
		return
	}

	line = mock.Lines[0]
	if line.X1 != -10 || line.Y1 != -20 || line.X2 != 110 || line.Y2 != 120 {
		t.Errorf("No-clip line after ops: got (%d,%d)-(%d,%d), want (-10,-20)-(110,120)",
			line.X1, line.Y1, line.X2, line.Y2)
	}
}

func TestAllConverterTypes(t *testing.T) {
	converters := []struct {
		name string
		test func(t *testing.T)
	}{
		{"RasConvInt", func(t *testing.T) { testConverterType[RasConvInt](t) }},
		{"RasConvIntSat", func(t *testing.T) { testConverterType[RasConvIntSat](t) }},
		{"RasConvInt3x", func(t *testing.T) { testConverterType[RasConvInt3x](t) }},
		{"RasConvDbl", func(t *testing.T) { testConverterType[RasConvDbl](t) }},
		{"RasConvDbl3x", func(t *testing.T) { testConverterType[RasConvDbl3x](t) }},
	}

	for _, conv := range converters {
		t.Run(conv.name, conv.test)
	}
}

func testConverterType[Conv any](t *testing.T) {
	mock := &MockRasterizer{}
	clipper := NewRasterizerSlClip[Conv]()

	// Basic test: draw a line without clipping
	clipper.MoveTo(10, 20)
	clipper.LineTo(mock, 30, 40)

	if len(mock.Lines) != 1 {
		t.Errorf("Expected 1 line for converter type, got %d", len(mock.Lines))
		return
	}

	// Test with clipping enabled
	mock.Reset()
	clipper.ClipBox(15, 15, 35, 35)
	clipper.MoveTo(10, 20)
	clipper.LineTo(mock, 30, 40)

	// Should have at least one line (could be clipped)
	if len(mock.Lines) == 0 {
		t.Errorf("Expected at least one line with clipping enabled")
	}
}

func TestComplexClippingScenario(t *testing.T) {
	mock := &MockRasterizer{}
	clipper := NewRasterizerSlClip[RasConvInt]()

	// Set a clipping box
	clipper.ClipBox(20, 20, 80, 80)

	// Draw multiple lines that cross different boundaries
	testLines := []struct {
		startX, startY, endX, endY float64
		description                string
	}{
		{10, 50, 90, 50, "horizontal line crossing both boundaries"},
		{50, 10, 50, 90, "vertical line crossing both boundaries"},
		{10, 10, 90, 90, "diagonal line crossing all boundaries"},
		{30, 30, 70, 70, "fully visible line"},
		{0, 0, 10, 10, "fully outside line"},
	}

	totalLines := 0
	for _, line := range testLines {
		t.Run(line.description, func(t *testing.T) {
			beforeCount := len(mock.Lines)
			clipper.MoveTo(line.startX, line.startY)
			clipper.LineTo(mock, line.endX, line.endY)
			afterCount := len(mock.Lines)

			linesAdded := afterCount - beforeCount
			totalLines += linesAdded

			// Log the result for debugging
			t.Logf("%s: added %d lines", line.description, linesAdded)
		})
	}

	t.Logf("Total lines drawn: %d", totalLines)

	// Verify that some lines were drawn (not all should be clipped)
	if totalLines == 0 {
		t.Errorf("No lines were drawn, clipping might be too aggressive")
	}
}

func TestTypeAliases(t *testing.T) {
	// Test that type aliases work correctly
	mock := &MockRasterizer{}
	var clipInt = NewRasterizerSlClip[RasConvInt]()
	var clipIntSat = NewRasterizerSlClip[RasConvIntSat]()
	var clipInt3x = NewRasterizerSlClip[RasConvInt3x]()
	var clipDbl = NewRasterizerSlClip[RasConvDbl]()
	var clipDbl3x = NewRasterizerSlClip[RasConvDbl3x]()

	// Basic functionality test for each type

	rasterizers := []interface{}{
		clipInt, clipIntSat, clipInt3x, clipDbl, clipDbl3x,
	}

	for i, rast := range rasterizers {
		t.Run(typeName(i), func(t *testing.T) {
			mock.Reset()

			// Use type switch to call methods (since they have the same interface)
			switch r := rast.(type) {
			case *RasterizerSlClip[RasConvInt]:
				r.MoveTo(10, 20)
				r.LineTo(mock, 30, 40)
			case *RasterizerSlClip[RasConvIntSat]:
				r.MoveTo(10, 20)
				r.LineTo(mock, 30, 40)
			case *RasterizerSlClip[RasConvInt3x]:
				r.MoveTo(10, 20)
				r.LineTo(mock, 30, 40)
			case *RasterizerSlClip[RasConvDbl]:
				r.MoveTo(10, 20)
				r.LineTo(mock, 30, 40)
			case *RasterizerSlClip[RasConvDbl3x]:
				r.MoveTo(10, 20)
				r.LineTo(mock, 30, 40)
			}

			if len(mock.Lines) != 1 {
				t.Errorf("Expected 1 line for type alias test, got %d", len(mock.Lines))
			}
		})
	}
}

func typeName(i int) string {
	names := []string{"Int", "IntSat", "Int3x", "Dbl", "Dbl3x"}
	return names[i]
}
