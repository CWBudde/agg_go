package markers

import (
	"testing"

	"agg_go/internal/basics"
)

// MockRenderer implements the BaseRenderer interface for testing
type MockRenderer struct {
	width, height int
	pixels        map[[2]int]MockColor
	clipBox       basics.RectI
}

type MockColor struct {
	R, G, B, A uint8
}

func NewMockRenderer(width, height int) *MockRenderer {
	return &MockRenderer{
		width:   width,
		height:  height,
		pixels:  make(map[[2]int]MockColor),
		clipBox: basics.RectI{X1: 0, Y1: 0, X2: width, Y2: height},
	}
}

func (m *MockRenderer) BlendPixel(x, y int, c MockColor, cover basics.Int8u) {
	m.pixels[[2]int{x, y}] = c
}

func (m *MockRenderer) BlendHline(x1, y, x2 int, c MockColor, cover basics.Int8u) {
	for x := x1; x <= x2; x++ {
		m.pixels[[2]int{x, y}] = c
	}
}

func (m *MockRenderer) BlendVline(x, y1, y2 int, c MockColor, cover basics.Int8u) {
	for y := y1; y <= y2; y++ {
		m.pixels[[2]int{x, y}] = c
	}
}

func (m *MockRenderer) BlendBar(x1, y1, x2, y2 int, c MockColor, cover basics.Int8u) {
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			m.pixels[[2]int{x, y}] = c
		}
	}
}

func (m *MockRenderer) BoundingClipBox() basics.RectI {
	return m.clipBox
}

func (m *MockRenderer) GetPixelCount() int {
	return len(m.pixels)
}

func (m *MockRenderer) HasPixel(x, y int) bool {
	_, exists := m.pixels[[2]int{x, y}]
	return exists
}

// Test basic marker renderer creation
func TestNewRendererMarkersSimple(t *testing.T) {
	mockRenderer := NewMockRenderer(100, 100)
	markerRenderer := NewRendererMarkers[*MockRenderer, MockColor](mockRenderer)

	if markerRenderer == nil {
		t.Fatal("Failed to create marker renderer")
	}
}

// Test visibility function
func TestVisibleSimple(t *testing.T) {
	mockRenderer := NewMockRenderer(100, 100)
	markerRenderer := NewRendererMarkers[*MockRenderer, MockColor](mockRenderer)

	// Test visible marker
	if !markerRenderer.Visible(50, 50, 5) {
		t.Error("Marker at center should be visible")
	}

	// Test marker outside bounds
	if markerRenderer.Visible(-10, -10, 5) {
		t.Error("Marker outside bounds should not be visible")
	}

	// Test marker at edge
	if !markerRenderer.Visible(95, 95, 5) {
		t.Error("Marker at edge should be visible")
	}
}

// Test square marker
func TestSquareMarkerSimple(t *testing.T) {
	mockRenderer := NewMockRenderer(100, 100)
	markerRenderer := NewRendererMarkers[*MockRenderer, MockColor](mockRenderer)

	// Set colors
	fillColor := MockColor{R: 255, G: 0, B: 0, A: 255}
	lineColor := MockColor{R: 0, G: 255, B: 0, A: 255}
	markerRenderer.FillColor(fillColor)
	markerRenderer.LineColor(lineColor)

	// Draw a square marker
	markerRenderer.Square(50, 50, 5)

	// Check that some pixels were drawn
	if mockRenderer.GetPixelCount() == 0 {
		t.Error("Square marker should have drawn pixels")
	}
}

// Test diamond marker
func TestDiamondMarkerSimple(t *testing.T) {
	mockRenderer := NewMockRenderer(100, 100)
	markerRenderer := NewRendererMarkers[*MockRenderer, MockColor](mockRenderer)

	// Set colors
	fillColor := MockColor{R: 255, G: 0, B: 0, A: 255}
	lineColor := MockColor{R: 0, G: 255, B: 0, A: 255}
	markerRenderer.FillColor(fillColor)
	markerRenderer.LineColor(lineColor)

	// Draw a diamond marker
	markerRenderer.Diamond(50, 50, 5)

	// Check that some pixels were drawn
	if mockRenderer.GetPixelCount() == 0 {
		t.Error("Diamond marker should have drawn pixels")
	}
}

// Test circle marker
func TestCircleMarkerSimple(t *testing.T) {
	mockRenderer := NewMockRenderer(100, 100)
	markerRenderer := NewRendererMarkers[*MockRenderer, MockColor](mockRenderer)

	// Set colors
	fillColor := MockColor{R: 255, G: 0, B: 0, A: 255}
	lineColor := MockColor{R: 0, G: 255, B: 0, A: 255}
	markerRenderer.FillColor(fillColor)
	markerRenderer.LineColor(lineColor)

	// Draw a circle marker
	markerRenderer.Circle(50, 50, 5)

	// Check that some pixels were drawn
	if mockRenderer.GetPixelCount() == 0 {
		t.Error("Circle marker should have drawn pixels")
	}
}

// Test all marker types through generic interface
func TestAllMarkerTypesSimple(t *testing.T) {
	mockRenderer := NewMockRenderer(200, 200)
	markerRenderer := NewRendererMarkers[*MockRenderer, MockColor](mockRenderer)

	// Set colors
	fillColor := MockColor{R: 255, G: 0, B: 0, A: 255}
	lineColor := MockColor{R: 0, G: 255, B: 0, A: 255}
	markerRenderer.FillColor(fillColor)
	markerRenderer.LineColor(lineColor)

	// Test all marker types
	markerTypes := []MarkerType{
		MarkerSquare, MarkerDiamond, MarkerCircle, MarkerCrossedCircle,
		MarkerSemiEllipseLeft, MarkerSemiEllipseRight, MarkerSemiEllipseUp, MarkerSemiEllipseDown,
		MarkerTriangleLeft, MarkerTriangleRight, MarkerTriangleUp, MarkerTriangleDown,
		MarkerFourRays, MarkerCross, MarkerX, MarkerDash, MarkerDot, MarkerPixel,
	}

	for i, markerType := range markerTypes {
		x := 20 + (i%10)*15
		y := 20 + (i/10)*15
		markerRenderer.Marker(x, y, 5, markerType)
	}

	// Check that pixels were drawn
	if mockRenderer.GetPixelCount() == 0 {
		t.Error("Markers should have drawn pixels")
	}
}

// Test zero radius markers
func TestZeroRadiusMarkersSimple(t *testing.T) {
	mockRenderer := NewMockRenderer(100, 100)
	markerRenderer := NewRendererMarkers[*MockRenderer, MockColor](mockRenderer)

	// Set colors
	fillColor := MockColor{R: 255, G: 0, B: 0, A: 255}
	lineColor := MockColor{R: 0, G: 255, B: 0, A: 255}
	markerRenderer.FillColor(fillColor)
	markerRenderer.LineColor(lineColor)

	// Test zero radius (should draw single pixels)
	markerRenderer.Square(50, 50, 0)
	markerRenderer.Circle(51, 50, 0)
	markerRenderer.Diamond(52, 50, 0)

	// Check that at least 3 pixels were drawn
	if mockRenderer.GetPixelCount() < 3 {
		t.Error("Zero radius markers should have drawn at least 3 pixels")
	}
}

// Test batch marker rendering
func TestBatchMarkersSimple(t *testing.T) {
	mockRenderer := NewMockRenderer(100, 100)
	markerRenderer := NewRendererMarkers[*MockRenderer, MockColor](mockRenderer)

	// Set colors
	fillColor := MockColor{R: 255, G: 0, B: 0, A: 255}
	lineColor := MockColor{R: 0, G: 255, B: 0, A: 255}
	markerRenderer.FillColor(fillColor)
	markerRenderer.LineColor(lineColor)

	// Test markers with same radius
	x := []int{10, 20, 30, 40}
	y := []int{10, 20, 30, 40}
	markerRenderer.Markers(x, y, 3, MarkerSquare)

	// Check that pixels were drawn
	if mockRenderer.GetPixelCount() == 0 {
		t.Error("Batch markers should have drawn pixels")
	}
}

// Test marker type string representation
func TestMarkerTypeStringSimple(t *testing.T) {
	tests := []struct {
		markerType MarkerType
		expected   string
	}{
		{MarkerSquare, "square"},
		{MarkerCircle, "circle"},
		{MarkerDiamond, "diamond"},
		{MarkerCross, "cross"},
		{MarkerX, "x"},
		{MarkerPixel, "pixel"},
		{MarkerType(999), "unknown"},
	}

	for _, test := range tests {
		result := test.markerType.String()
		if result != test.expected {
			t.Errorf("Expected %s, got %s for marker type %d", test.expected, result, test.markerType)
		}
	}
}

// Test invalid batch inputs
func TestInvalidBatchInputsSimple(t *testing.T) {
	mockRenderer := NewMockRenderer(100, 100)
	markerRenderer := NewRendererMarkers[*MockRenderer, MockColor](mockRenderer)

	// Set colors
	fillColor := MockColor{R: 255, G: 0, B: 0, A: 255}
	markerRenderer.FillColor(fillColor)

	// Test mismatched slice lengths
	x := []int{10, 20}
	y := []int{10}                                // Different length
	markerRenderer.Markers(x, y, 3, MarkerSquare) // Should handle gracefully

	// Test empty slices
	emptyX := []int{}
	emptyY := []int{}
	markerRenderer.Markers(emptyX, emptyY, 3, MarkerSquare) // Should handle gracefully

	// If we get here without panic, the test passes
}
