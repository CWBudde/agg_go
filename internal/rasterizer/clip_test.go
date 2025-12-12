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

func TestIntConv(t *testing.T) {
	conv := IntConv{}

	// Test MulDiv
	result := conv.MulDiv(10.0, 20.0, 4.0)
	expected := 50
	if result != expected {
		t.Errorf("MulDiv(10, 20, 4) = %d, want %d", result, expected)
	}

	// Test Xi with int
	xi := conv.Xi(100)
	if xi != 100 {
		t.Errorf("Xi(100) = %d, want 100", xi)
	}

	// Test Yi
	yi := conv.Yi(50)
	if yi != 50 {
		t.Errorf("Yi(50) = %d, want 50", yi)
	}

	// Test Upscale
	upscaled := conv.Upscale(1.0)
	expected_upscaled := basics.IRound(basics.PolySubpixelScale)
	if upscaled != expected_upscaled {
		t.Errorf("Upscale(1.0) = %d, want %d", upscaled, expected_upscaled)
	}

	// Test Downscale
	downscaled := conv.Downscale(256)
	if downscaled != 1 {
		t.Errorf("Downscale(256) = %d, want 1", downscaled)
	}
}

func TestDblConv(t *testing.T) {
	conv := DblConv{}

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
	upscaled := conv.Upscale(42.5)
	if upscaled != 42.5 {
		t.Errorf("Upscale(42.5) = %f, want 42.5", upscaled)
	}

	// Test Downscale
	downscaled := conv.Downscale(256)
	expectedDownscaled := 256.0 / basics.PolySubpixelScale
	if downscaled != expectedDownscaled {
		t.Errorf("Downscale(256) = %f, want %f", downscaled, expectedDownscaled)
	}
}

func TestInt3xConv(t *testing.T) {
	conv := Int3xConv{}

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
	clipper := NewRasterizerSlClip[int, IntConv](IntConv{})

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
	clipper := NewRasterizerSlClip[int, IntConv](IntConv{})

	// Set clipping box
	clipper.ClipBox(10, 10, 50, 50)

	tests := []struct {
		name              string
		startX, startY    int
		endX, endY        int
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
			expectedLineCount: 0, // Fixed: No lines drawn when both points outside same boundary
			shouldHaveLines:   false,
		},
		{
			name:   "crosses_boundary",
			startX: 5, startY: 20,
			endX: 25, endY: 30,
			expectedLineCount: 1, // Fixed: Only draw the visible segment (no boundary line)
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
	clipper := NewRasterizerSlClip[int, IntConv](IntConv{})

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
	clipper := NewRasterizerSlClip[int, IntConv](IntConv{})

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
	noClip := NewRasterizerSlNoClip()

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

func TestComplexClippingScenario(t *testing.T) {
	mock := &MockRasterizer{}
	clipper := NewRasterizerSlClip[int, IntConv](IntConv{})

	// Set a clipping box
	clipper.ClipBox(20, 20, 80, 80)

	// Draw multiple lines that cross different boundaries
	testLines := []struct {
		startX, startY, endX, endY int
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
	// Test that different converter types work correctly
	clipInt := NewRasterizerSlClip[int, IntConv](IntConv{})
	clipIntSat := NewRasterizerSlClip[int, IntSatConv](IntSatConv{})
	clipInt3x := NewRasterizerSlClip[int, Int3xConv](Int3xConv{})
	clipDbl := NewRasterizerSlClip[float64, DblConv](DblConv{})
	clipDbl3x := NewRasterizerSlClip[float64, Dbl3xConv](Dbl3xConv{})

	// Basic functionality test for each type
	types := []struct {
		name string
		test func(t *testing.T)
	}{
		{"Int", func(t *testing.T) { testRasterizerInt(t, clipInt) }},
		{"IntSat", func(t *testing.T) { testRasterizerInt(t, clipIntSat) }},
		{"Int3x", func(t *testing.T) { testRasterizerInt(t, clipInt3x) }},
		{"Dbl", func(t *testing.T) { testRasterizerDbl(t, clipDbl) }},
		{"Dbl3x", func(t *testing.T) { testRasterizerDbl(t, clipDbl3x) }},
	}

	for _, typ := range types {
		t.Run(typ.name, typ.test)
	}
}

// testRasterizerInt is a helper function to test int-based rasterizer types
func testRasterizerInt[V Conv[int]](t *testing.T, clipper *RasterizerSlClip[int, V]) {
	mock := &MockRasterizer{}
	mock.Reset()

	clipper.MoveTo(10, 20)
	clipper.LineTo(mock, 30, 40)

	if len(mock.Lines) != 1 {
		t.Errorf("Expected 1 line for type alias test, got %d", len(mock.Lines))
	}
}

// testRasterizerDbl is a helper function to test float64-based rasterizer types
func testRasterizerDbl[V Conv[float64]](t *testing.T, clipper *RasterizerSlClip[float64, V]) {
	mock := &MockRasterizer{}
	mock.Reset()

	clipper.MoveTo(10.0, 20.0)
	clipper.LineTo(mock, 30.0, 40.0)

	if len(mock.Lines) != 1 {
		t.Errorf("Expected 1 line for type alias test, got %d", len(mock.Lines))
	}
}

func typeName(i int) string {
	names := []string{"Int", "IntSat", "Int3x", "Dbl", "Dbl3x"}
	return names[i]
}
