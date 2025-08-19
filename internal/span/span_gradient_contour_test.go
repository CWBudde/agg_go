package span

import (
	"math"
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/path"
)

func TestNewGradientContour(t *testing.T) {
	gc := NewGradientContour()

	if gc == nil {
		t.Fatal("NewGradientContour returned nil")
	}

	if gc.frame != 10 {
		t.Errorf("Expected frame = 10, got %d", gc.frame)
	}

	if gc.d1 != 0.0 {
		t.Errorf("Expected d1 = 0.0, got %f", gc.d1)
	}

	if gc.d2 != 100.0 {
		t.Errorf("Expected d2 = 100.0, got %f", gc.d2)
	}
}

func TestNewGradientContourWithDistances(t *testing.T) {
	d1, d2 := 10.0, 200.0
	gc := NewGradientContourWithDistances(d1, d2)

	if gc == nil {
		t.Fatal("NewGradientContourWithDistances returned nil")
	}

	if gc.d1 != d1 {
		t.Errorf("Expected d1 = %f, got %f", d1, gc.d1)
	}

	if gc.d2 != d2 {
		t.Errorf("Expected d2 = %f, got %f", d2, gc.d2)
	}
}

func TestGradientContourProperties(t *testing.T) {
	gc := NewGradientContour()

	// Test setters and getters
	gc.SetD1(5.0)
	if gc.d1 != 5.0 {
		t.Errorf("Expected d1 = 5.0, got %f", gc.d1)
	}

	gc.SetD2(150.0)
	if gc.d2 != 150.0 {
		t.Errorf("Expected d2 = 150.0, got %f", gc.d2)
	}

	gc.SetFrame(20)
	if gc.Frame() != 20 {
		t.Errorf("Expected frame = 20, got %d", gc.Frame())
	}
}

func TestGradientContourCalculateEmptyBuffer(t *testing.T) {
	gc := NewGradientContour()

	// Without creating a contour, buffer should be nil
	result := gc.Calculate(100, 200, 1000)
	if result != 0 {
		t.Errorf("Expected Calculate to return 0 for empty buffer, got %d", result)
	}
}

func TestDistanceTransformAlgorithm(t *testing.T) {
	gc := NewGradientContour()

	// Test the 1D distance transform with a simple pattern
	length := 7
	spanf := []float32{0, math.MaxFloat32, math.MaxFloat32, 0, math.MaxFloat32, math.MaxFloat32, 0}
	spang := make([]float32, length+1)
	spanr := make([]float32, length)
	spann := make([]int, length)

	gc.dt(spanf, spang, spanr, spann, length)

	// Verify that distances are computed correctly
	// Expected pattern: distances to nearest zero
	expected := []float32{0, 1, 4, 0, 1, 4, 0}

	tolerance := float32(0.001)
	for i := 0; i < length; i++ {
		if math.Abs(float64(spanr[i]-expected[i])) > float64(tolerance) {
			t.Errorf("DT result[%d]: expected %f, got %f", i, expected[i], spanr[i])
		}
	}
}

func TestContourCreateWithNilPath(t *testing.T) {
	gc := NewGradientContour()

	result := gc.ContourCreate(nil)
	if result != nil {
		t.Error("Expected ContourCreate to return nil for nil path")
	}
}

func TestContourCreateWithSimplePath(t *testing.T) {
	gc := NewGradientContour()
	ps := path.NewPathStorage()

	// Create a simple rectangular path
	ps.MoveTo(10.0, 10.0)
	ps.LineTo(50.0, 10.0)
	ps.LineTo(50.0, 30.0)
	ps.LineTo(10.0, 30.0)
	ps.ClosePolygon(basics.PathFlagsClose)

	result := gc.ContourCreate(ps)

	if result == nil {
		t.Fatal("Expected ContourCreate to return non-nil buffer")
	}

	if gc.ContourWidth() <= 0 {
		t.Errorf("Expected positive width, got %d", gc.ContourWidth())
	}

	if gc.ContourHeight() <= 0 {
		t.Errorf("Expected positive height, got %d", gc.ContourHeight())
	}

	expectedSize := gc.ContourWidth() * gc.ContourHeight()
	if len(result) != expectedSize {
		t.Errorf("Expected buffer size %d, got %d", expectedSize, len(result))
	}
}

func TestContourCreateWithCircularPath(t *testing.T) {
	gc := NewGradientContour()
	ps := path.NewPathStorage()

	// Create a circular path using curve approximation
	centerX, centerY, radius := 25.0, 25.0, 15.0

	// Start at rightmost point
	ps.MoveTo(centerX+radius, centerY)

	// Create circle using 4 cubic bezier curves
	kappa := 4.0 * (math.Sqrt(2.0) - 1.0) / 3.0
	cp := kappa * radius

	// Right to top
	ps.Curve4(centerX+radius, centerY-cp, centerX+cp, centerY-radius, centerX, centerY-radius)
	// Top to left
	ps.Curve4(centerX-cp, centerY-radius, centerX-radius, centerY-cp, centerX-radius, centerY)
	// Left to bottom
	ps.Curve4(centerX-radius, centerY+cp, centerX-cp, centerY+radius, centerX, centerY+radius)
	// Bottom to right
	ps.Curve4(centerX+cp, centerY+radius, centerX+radius, centerY+cp, centerX+radius, centerY)
	ps.ClosePolygon(basics.PathFlagsClose)

	result := gc.ContourCreate(ps)

	if result == nil {
		t.Fatal("Expected ContourCreate to return non-nil buffer for circular path")
	}

	// Verify dimensions are reasonable
	if gc.ContourWidth() < 20 || gc.ContourHeight() < 20 {
		t.Errorf("Expected reasonable dimensions for circle, got %dx%d",
			gc.ContourWidth(), gc.ContourHeight())
	}
}

func TestGradientContourCalculateWithBuffer(t *testing.T) {
	gc := NewGradientContour()
	gc.SetD1(10.0)
	gc.SetD2(200.0)

	ps := path.NewPathStorage()
	ps.MoveTo(5.0, 5.0)
	ps.LineTo(15.0, 5.0)
	ps.LineTo(15.0, 15.0)
	ps.LineTo(5.0, 15.0)
	ps.ClosePolygon(basics.PathFlagsClose)

	buffer := gc.ContourCreate(ps)
	if buffer == nil {
		t.Fatal("Failed to create contour buffer")
	}

	// Test Calculate function with various coordinates
	testCases := []struct {
		x, y, d int
		name    string
	}{
		{0, 0, 1000, "origin"},
		{100 << GradientSubpixelShift, 100 << GradientSubpixelShift, 1000, "middle"},
		{-50 << GradientSubpixelShift, -50 << GradientSubpixelShift, 1000, "negative coords"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := gc.Calculate(tc.x, tc.y, tc.d)

			// Result should be within reasonable range based on d1 and d2
			minExpected := int(gc.d1) << GradientSubpixelShift
			maxExpected := int(gc.d2) << GradientSubpixelShift

			if result < minExpected || result > maxExpected {
				t.Errorf("Calculate result %d outside expected range [%d, %d]",
					result, minExpected, maxExpected)
			}
		})
	}
}

func TestPerformDistanceTransformUniformInput(t *testing.T) {
	gc := NewGradientContour()

	// Test with uniform white buffer (all pixels = 255)
	width, height := 10, 10
	bwBuffer := make([]uint8, width*height)
	for i := range bwBuffer {
		bwBuffer[i] = 255
	}

	result := gc.performDistanceTransform(bwBuffer, width, height)

	if result == nil {
		t.Fatal("Expected non-nil result from performDistanceTransform")
	}

	if len(result) != width*height {
		t.Errorf("Expected result length %d, got %d", width*height, len(result))
	}

	// All values should be the same (0) since input was uniform
	for i, val := range result {
		if val != 0 {
			t.Errorf("Expected uniform result value 0, got %d at index %d", val, i)
		}
	}
}

func TestPerformDistanceTransformSingleBlackPixel(t *testing.T) {
	gc := NewGradientContour()

	// Test with single black pixel in center
	width, height := 5, 5
	bwBuffer := make([]uint8, width*height)

	// Initialize to white
	for i := range bwBuffer {
		bwBuffer[i] = 255
	}

	// Set center pixel to black
	centerIndex := (height/2)*width + (width / 2)
	bwBuffer[centerIndex] = 0

	result := gc.performDistanceTransform(bwBuffer, width, height)

	if result == nil {
		t.Fatal("Expected non-nil result from performDistanceTransform")
	}

	// Center pixel should have minimum distance (darkest)
	centerValue := result[centerIndex]

	// Check that distances increase as we move away from center
	// Corner pixels should have maximum distance
	corners := []int{
		0,                    // top-left
		width - 1,            // top-right
		(height - 1) * width, // bottom-left
		height*width - 1,     // bottom-right
	}

	for _, corner := range corners {
		if result[corner] <= centerValue {
			t.Errorf("Expected corner distance %d > center distance %d",
				result[corner], centerValue)
		}
	}
}

func BenchmarkContourCreate(b *testing.B) {
	gc := NewGradientContour()
	ps := path.NewPathStorage()

	// Create a moderately complex path
	ps.MoveTo(0, 0)
	for i := 0; i < 20; i++ {
		angle := float64(i) * 2 * math.Pi / 20
		x := 25 + 20*math.Cos(angle)
		y := 25 + 20*math.Sin(angle)
		ps.LineTo(x, y)
	}
	ps.ClosePolygon(basics.PathFlagsClose)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		gc.ContourCreate(ps)
	}
}

func BenchmarkCalculate(b *testing.B) {
	gc := NewGradientContour()
	ps := path.NewPathStorage()

	ps.MoveTo(10.0, 10.0)
	ps.LineTo(30.0, 10.0)
	ps.LineTo(30.0, 30.0)
	ps.LineTo(10.0, 30.0)
	ps.ClosePolygon(basics.PathFlagsClose)

	gc.ContourCreate(ps)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		x := (i % 100) << GradientSubpixelShift
		y := ((i / 100) % 100) << GradientSubpixelShift
		gc.Calculate(x, y, 1000)
	}
}
