package basics

import (
	"testing"
)

func TestClippingFlags(t *testing.T) {
	clipBox := Rect[float64]{X1: 10, Y1: 20, X2: 50, Y2: 60}

	tests := []struct {
		name     string
		x, y     float64
		expected uint32
	}{
		{"center", 30, 40, 0},
		{"left", 5, 40, ClippingFlagsX1Clipped},
		{"right", 55, 40, ClippingFlagsX2Clipped},
		{"bottom", 30, 15, ClippingFlagsY1Clipped},
		{"top", 30, 65, ClippingFlagsY2Clipped},
		{"bottom-left", 5, 15, ClippingFlagsX1Clipped | ClippingFlagsY1Clipped},
		{"top-right", 55, 65, ClippingFlagsX2Clipped | ClippingFlagsY2Clipped},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ClippingFlags(tt.x, tt.y, clipBox)
			if result != tt.expected {
				t.Errorf("ClippingFlags(%f, %f) = %d, want %d", tt.x, tt.y, result, tt.expected)
			}
		})
	}
}

func TestClippingFlagsX(t *testing.T) {
	clipBox := Rect[float64]{X1: 10, Y1: 20, X2: 50, Y2: 60}

	tests := []struct {
		name     string
		x        float64
		expected uint32
	}{
		{"center", 30, 0},
		{"left", 5, ClippingFlagsX1Clipped},
		{"right", 55, ClippingFlagsX2Clipped},
		{"edge-left", 10, 0},
		{"edge-right", 50, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ClippingFlagsX(tt.x, clipBox)
			if result != tt.expected {
				t.Errorf("ClippingFlagsX(%f) = %d, want %d", tt.x, result, tt.expected)
			}
		})
	}
}

func TestClippingFlagsY(t *testing.T) {
	clipBox := Rect[float64]{X1: 10, Y1: 20, X2: 50, Y2: 60}

	tests := []struct {
		name     string
		y        float64
		expected uint32
	}{
		{"center", 40, 0},
		{"bottom", 15, ClippingFlagsY1Clipped},
		{"top", 65, ClippingFlagsY2Clipped},
		{"edge-bottom", 20, 0},
		{"edge-top", 60, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ClippingFlagsY(tt.y, clipBox)
			if result != tt.expected {
				t.Errorf("ClippingFlagsY(%f) = %d, want %d", tt.y, result, tt.expected)
			}
		})
	}
}

func TestClipLiangBarsky(t *testing.T) {
	clipBox := Rect[float64]{X1: 10, Y1: 20, X2: 50, Y2: 60}

	tests := []struct {
		name               string
		x1, y1, x2, y2     float64
		expectedPointCount uint32
		expectedPoints     []float64 // [x1, y1, x2, y2]
	}{
		{
			name: "fully_visible",
			x1:   20, y1: 30, x2: 40, y2: 50,
			expectedPointCount: 0, // Fully visible means no clipping needed
		},
		{
			name: "fully_outside_left",
			x1:   0, y1: 30, x2: 5, y2: 50,
			expectedPointCount: 0,
		},
		{
			name: "fully_outside_right",
			x1:   60, y1: 30, x2: 70, y2: 50,
			expectedPointCount: 0,
		},
		{
			name: "crosses_left_boundary",
			x1:   5, y1: 30, x2: 25, y2: 40,
			expectedPointCount: 2,
			expectedPoints:     []float64{10, 32.5, 25, 40},
		},
		{
			name: "crosses_right_boundary",
			x1:   25, y1: 30, x2: 55, y2: 50,
			expectedPointCount: 2,
			expectedPoints:     []float64{25, 30, 50, 46.67}, // approximately
		},
		{
			name: "horizontal_line_clipped",
			x1:   5, y1: 30, x2: 55, y2: 30,
			expectedPointCount: 2,
			expectedPoints:     []float64{10, 30, 50, 30},
		},
		{
			name: "vertical_line_clipped",
			x1:   30, y1: 15, x2: 30, y2: 65,
			expectedPointCount: 2,
			expectedPoints:     []float64{30, 20, 30, 60},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := make([]float64, 4)
			y := make([]float64, 4)

			pointCount := ClipLiangBarsky(tt.x1, tt.y1, tt.x2, tt.y2, clipBox, x, y)

			if pointCount != tt.expectedPointCount {
				t.Errorf("ClipLiangBarsky returned %d points, want %d", pointCount, tt.expectedPointCount)
				return
			}

			if pointCount > 0 && len(tt.expectedPoints) >= 4 {
				tolerance := 0.1
				for i := uint32(0); i < pointCount; i++ {
					expectedX := tt.expectedPoints[i*2]
					expectedY := tt.expectedPoints[i*2+1]

					if abs(x[i]-expectedX) > tolerance {
						t.Errorf("Point %d X: got %f, want %f (tolerance %f)", i, x[i], expectedX, tolerance)
					}
					if abs(y[i]-expectedY) > tolerance {
						t.Errorf("Point %d Y: got %f, want %f (tolerance %f)", i, y[i], expectedY, tolerance)
					}
				}
			}
		})
	}
}

func TestClipMovePoint(t *testing.T) {
	clipBox := Rect[float64]{X1: 10, Y1: 20, X2: 50, Y2: 60}

	tests := []struct {
		name           string
		x1, y1, x2, y2 float64
		flags          uint32
		expectedX      float64
		expectedY      float64
		expectedResult bool
	}{
		{
			name: "clip_to_left_boundary",
			x1:   5, y1: 30, x2: 25, y2: 50,
			flags:          ClippingFlagsX1Clipped,
			expectedX:      10,
			expectedY:      35, // interpolated Y
			expectedResult: true,
		},
		{
			name: "clip_to_right_boundary",
			x1:   25, y1: 30, x2: 55, y2: 50,
			flags:          ClippingFlagsX2Clipped,
			expectedX:      50,
			expectedY:      46.67, // interpolated Y (approximately)
			expectedResult: true,
		},
		{
			name: "vertical_line_cannot_clip_x",
			x1:   30, y1: 15, x2: 30, y2: 65,
			flags:          ClippingFlagsX1Clipped,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y := tt.x1, tt.y1
			result := ClipMovePoint(tt.x1, tt.y1, tt.x2, tt.y2, clipBox, &x, &y, tt.flags)

			if result != tt.expectedResult {
				t.Errorf("ClipMovePoint returned %v, want %v", result, tt.expectedResult)
				return
			}

			if result {
				tolerance := 0.1
				if abs(x-tt.expectedX) > tolerance {
					t.Errorf("Clipped X: got %f, want %f (tolerance %f)", x, tt.expectedX, tolerance)
				}
				if abs(y-tt.expectedY) > tolerance {
					t.Errorf("Clipped Y: got %f, want %f (tolerance %f)", y, tt.expectedY, tolerance)
				}
			}
		})
	}
}

func TestClipLineSegment(t *testing.T) {
	clipBox := Rect[float64]{X1: 10, Y1: 20, X2: 50, Y2: 60}

	tests := []struct {
		name           string
		x1, y1, x2, y2 float64
		expectedResult uint32
		expectedX1     float64
		expectedY1     float64
		expectedX2     float64
		expectedY2     float64
	}{
		{
			name: "fully_visible",
			x1:   20, y1: 30, x2: 40, y2: 50,
			expectedResult: 0, // no clipping needed
			expectedX1:     20,
			expectedY1:     30,
			expectedX2:     40,
			expectedY2:     50,
		},
		{
			name: "fully_clipped",
			x1:   0, y1: 5, x2: 5, y2: 10,
			expectedResult: 4, // fully clipped
		},
		{
			name: "first_point_clipped",
			x1:   5, y1: 30, x2: 25, y2: 40,
			expectedResult: 1, // first point moved
			expectedX1:     10,
			expectedY1:     32.5,
			expectedX2:     25,
			expectedY2:     40,
		},
		{
			name: "second_point_clipped",
			x1:   25, y1: 30, x2: 55, y2: 50,
			expectedResult: 2, // second point moved
			expectedX1:     25,
			expectedY1:     30,
			expectedX2:     50,
			expectedY2:     46.67,
		},
		{
			name: "both_points_clipped",
			x1:   5, y1: 30, x2: 55, y2: 50,
			expectedResult: 3, // both points moved
			expectedX1:     10,
			expectedY1:     32.5,
			expectedX2:     50,
			expectedY2:     46.67,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x1, y1, x2, y2 := tt.x1, tt.y1, tt.x2, tt.y2
			result := ClipLineSegment(&x1, &y1, &x2, &y2, clipBox)

			if result != tt.expectedResult {
				t.Errorf("ClipLineSegment returned %d, want %d", result, tt.expectedResult)
				return
			}

			if result < 4 { // not fully clipped
				tolerance := 0.1
				if abs(x1-tt.expectedX1) > tolerance {
					t.Errorf("Clipped X1: got %f, want %f (tolerance %f)", x1, tt.expectedX1, tolerance)
				}
				if abs(y1-tt.expectedY1) > tolerance {
					t.Errorf("Clipped Y1: got %f, want %f (tolerance %f)", y1, tt.expectedY1, tolerance)
				}
				if abs(x2-tt.expectedX2) > tolerance {
					t.Errorf("Clipped X2: got %f, want %f (tolerance %f)", x2, tt.expectedX2, tolerance)
				}
				if abs(y2-tt.expectedY2) > tolerance {
					t.Errorf("Clipped Y2: got %f, want %f (tolerance %f)", y2, tt.expectedY2, tolerance)
				}
			}
		})
	}
}

func TestIntegerClipping(t *testing.T) {
	clipBox := Rect[int]{X1: 10, Y1: 20, X2: 50, Y2: 60}

	tests := []struct {
		name     string
		x, y     int
		expected uint32
	}{
		{"center", 30, 40, 0},
		{"left", 5, 40, ClippingFlagsX1Clipped},
		{"right", 55, 40, ClippingFlagsX2Clipped},
		{"bottom", 30, 15, ClippingFlagsY1Clipped},
		{"top", 30, 65, ClippingFlagsY2Clipped},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ClippingFlags(tt.x, tt.y, clipBox)
			if result != tt.expected {
				t.Errorf("ClippingFlags(%d, %d) = %d, want %d", tt.x, tt.y, result, tt.expected)
			}
		})
	}
}

// Helper function for floating point comparison
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
