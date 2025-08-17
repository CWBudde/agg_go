package basics

import (
	"math"
	"testing"
)

func TestFastSqrt(t *testing.T) {
	tests := []struct {
		input     uint32
		expected  uint32
		tolerance uint32
	}{
		{0, 0, 0},
		{1, 1, 1},
		{4, 2, 1},
		{9, 3, 1},
		{16, 4, 1},
		{25, 5, 1},
		{36, 6, 1},
		{49, 7, 1},
		{64, 8, 1},
		{81, 9, 1},
		{100, 10, 1},
		{10000, 100, 2},
		{1000000, 1000, 10},
	}

	for _, tt := range tests {
		result := FastSqrt(tt.input)
		diff := uint32(0)
		if result > tt.expected {
			diff = result - tt.expected
		} else {
			diff = tt.expected - result
		}

		if diff > tt.tolerance {
			t.Errorf("FastSqrt(%d) = %d, want %d (±%d), diff = %d",
				tt.input, result, tt.expected, tt.tolerance, diff)
		}
	}
}

func TestFastSqrtAccuracy(t *testing.T) {
	// Test accuracy against standard math.Sqrt for various values
	testValues := []uint32{1, 4, 16, 64, 256, 1024, 4096, 16384, 65536, 262144, 1048576}

	for _, val := range testValues {
		fastResult := FastSqrt(val)
		mathResult := uint32(math.Sqrt(float64(val)))

		// Allow for small error due to approximation
		diff := uint32(0)
		if fastResult > mathResult {
			diff = fastResult - mathResult
		} else {
			diff = mathResult - fastResult
		}

		// For most values, the error should be very small
		maxError := uint32(math.Max(1, float64(mathResult)/100)) // 1% error or at least 1
		if diff > maxError {
			t.Errorf("FastSqrt(%d) = %d, math.Sqrt = %d, diff = %d (max allowed: %d)",
				val, fastResult, mathResult, diff, maxError)
		}
	}
}

func TestBesj(t *testing.T) {
	tests := []struct {
		input     float64
		expected  float64
		tolerance float64
	}{
		{0.0, 1.0, 1e-10},        // J0(0) = 1
		{1.0, 0.7651976, 1e-4},   // J0(1) ≈ 0.7651976
		{2.0, 0.2238908, 1e-4},   // J0(2) ≈ 0.2238908
		{-1.0, 0.7651976, 1e-4},  // J0(-1) = J0(1) (even function)
		{5.0, -0.1775968, 2e-2},  // J0(5) ≈ -0.1775968 (larger tolerance for asymptotic approx)
		{10.0, -0.2459358, 1e-2}, // J0(10) ≈ -0.2459358 (larger tolerance)
	}

	for _, tt := range tests {
		result := Besj(tt.input)
		diff := math.Abs(result - tt.expected)

		if diff > tt.tolerance {
			t.Errorf("Besj(%f) = %f, want %f (±%e), diff = %e",
				tt.input, result, tt.expected, tt.tolerance, diff)
		}
	}
}

func TestBesjSeriesExpansion(t *testing.T) {
	// Test the series expansion for small values (< 3.0)
	smallValues := []float64{0.1, 0.5, 1.0, 2.0, 2.9}

	for _, x := range smallValues {
		result := Besj(x)

		// The result should be reasonable for small values
		// J0(x) oscillates between approximately -0.4 and 1.0 for small x
		if result > 1.1 || result < -0.5 {
			t.Errorf("Besj(%f) = %f, seems out of reasonable range for small values", x, result)
		}
	}
}

func TestBesjAsymptoticApproximation(t *testing.T) {
	// Test the asymptotic approximation for large values (>= 3.0)
	largeValues := []float64{3.0, 5.0, 10.0, 20.0}

	for _, x := range largeValues {
		result := Besj(x)

		// For large x, |J0(x)| should be roughly bounded by sqrt(2/(pi*x))
		bound := math.Sqrt(2.0 / (math.Pi * x))

		if math.Abs(result) > bound*1.5 { // Allow some tolerance
			t.Errorf("Besj(%f) = %f, exceeds expected bound %f for large values",
				x, result, bound*1.5)
		}
	}
}

func TestGeometryFunctions(t *testing.T) {
	t.Run("CrossProduct", func(t *testing.T) {
		// Test cross product with known values - using AGG formula
		result := CrossProduct(0, 0, 1, 0, 0, 1)
		expected := -1.0 // AGG's cross product formula gives -1 for this case
		if math.Abs(result-expected) > 1e-10 {
			t.Errorf("CrossProduct(0,0,1,0,0,1) = %f, want %f", result, expected)
		}

		// Test with collinear points (should be 0)
		result = CrossProduct(0, 0, 2, 0, 1, 0)
		if math.Abs(result) > 1e-10 {
			t.Errorf("CrossProduct of collinear points should be 0, got %f", result)
		}
	})

	t.Run("CalcDistance", func(t *testing.T) {
		// Test distance calculation
		result := CalcDistance(0, 0, 3, 4)
		expected := 5.0
		if math.Abs(result-expected) > 1e-10 {
			t.Errorf("CalcDistance(0,0,3,4) = %f, want %f", result, expected)
		}

		// Test zero distance
		result = CalcDistance(1, 1, 1, 1)
		if math.Abs(result) > 1e-10 {
			t.Errorf("CalcDistance of same point should be 0, got %f", result)
		}
	})

	t.Run("CalcSqDistance", func(t *testing.T) {
		// Test squared distance calculation
		result := CalcSqDistance(0, 0, 3, 4)
		expected := 25.0
		if math.Abs(result-expected) > 1e-10 {
			t.Errorf("CalcSqDistance(0,0,3,4) = %f, want %f", result, expected)
		}
	})

	t.Run("PointInTriangle", func(t *testing.T) {
		// Test point inside triangle
		result := PointInTriangle(0, 0, 3, 0, 1.5, 3, 1.5, 1)
		if !result {
			t.Error("Point (1.5,1) should be inside triangle")
		}

		// Test point outside triangle
		result = PointInTriangle(0, 0, 3, 0, 1.5, 3, 5, 5)
		if result {
			t.Error("Point (5,5) should be outside triangle")
		}
	})

	t.Run("CalcTriangleArea", func(t *testing.T) {
		// Test triangle area calculation
		result := CalcTriangleArea(0, 0, 3, 0, 0, 4)
		expected := 6.0
		if math.Abs(result-expected) > 1e-10 {
			t.Errorf("CalcTriangleArea(0,0,3,0,0,4) = %f, want %f", result, expected)
		}
	})
}

func TestCalcPolygonArea(t *testing.T) {
	t.Run("Square", func(t *testing.T) {
		vertices := []PointD{
			{X: 0, Y: 0},
			{X: 2, Y: 0},
			{X: 2, Y: 2},
			{X: 0, Y: 2},
		}
		result := CalcPolygonArea(vertices)
		expected := 4.0
		if math.Abs(result-expected) > 1e-10 {
			t.Errorf("CalcPolygonArea(square) = %f, want %f", result, expected)
		}
	})

	t.Run("Triangle", func(t *testing.T) {
		vertices := []PointD{
			{X: 0, Y: 0},
			{X: 3, Y: 0},
			{X: 0, Y: 4},
		}
		result := CalcPolygonArea(vertices)
		expected := 6.0
		if math.Abs(result-expected) > 1e-10 {
			t.Errorf("CalcPolygonArea(triangle) = %f, want %f", result, expected)
		}
	})

	t.Run("Empty polygon", func(t *testing.T) {
		vertices := []PointD{}
		result := CalcPolygonArea(vertices)
		if result != 0 {
			t.Errorf("CalcPolygonArea(empty) = %f, want 0", result)
		}
	})

	t.Run("Too few vertices", func(t *testing.T) {
		vertices := []PointD{{X: 0, Y: 0}, {X: 1, Y: 1}}
		result := CalcPolygonArea(vertices)
		if result != 0 {
			t.Errorf("CalcPolygonArea(2 vertices) = %f, want 0", result)
		}
	})
}

func TestCalcSegmentPointSqDistance(t *testing.T) {
	tests := []struct {
		name     string
		x1, y1   float64
		x2, y2   float64
		x, y     float64
		expected float64
	}{
		{
			name: "Point on segment",
			x1:   0, y1: 0, x2: 4, y2: 0,
			x: 2, y: 0,
			expected: 0,
		},
		{
			name: "Point perpendicular to segment",
			x1:   0, y1: 0, x2: 4, y2: 0,
			x: 2, y: 3,
			expected: 9, // 3^2
		},
		{
			name: "Point before segment start",
			x1:   2, y1: 2, x2: 4, y2: 2,
			x: 0, y: 2,
			expected: 4, // distance to (2,2)
		},
		{
			name: "Point after segment end",
			x1:   0, y1: 0, x2: 2, y2: 0,
			x: 4, y: 0,
			expected: 4, // distance to (2,0)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalcSegmentPointSqDistance(tt.x1, tt.y1, tt.x2, tt.y2, tt.x, tt.y)
			if math.Abs(result-tt.expected) > 1e-10 {
				t.Errorf("CalcSegmentPointSqDistance() = %f, want %f", result, tt.expected)
			}
		})
	}
}

func TestCalcIntersection(t *testing.T) {
	t.Run("Intersecting lines", func(t *testing.T) {
		// Two lines that intersect at (2, 2)
		x, y, ok := CalcIntersection(0, 0, 4, 4, 0, 4, 4, 0)
		if !ok {
			t.Error("Lines should intersect")
		}
		expectedX, expectedY := 2.0, 2.0
		if math.Abs(x-expectedX) > 1e-10 || math.Abs(y-expectedY) > 1e-10 {
			t.Errorf("Intersection point = (%f, %f), want (%f, %f)", x, y, expectedX, expectedY)
		}
	})

	t.Run("Parallel lines", func(t *testing.T) {
		// Two parallel lines
		_, _, ok := CalcIntersection(0, 0, 2, 0, 0, 1, 2, 1)
		if ok {
			t.Error("Parallel lines should not intersect")
		}
	})
}

func TestCalcLinePointDistance(t *testing.T) {
	t.Run("Horizontal segment", func(t *testing.T) {
		d := CalcLinePointDistance(0, 0, 4, 0, 2, 3)
		if math.Abs(d-3) > 1e-12 {
			t.Errorf("expected 3, got %f", d)
		}
		d = CalcLinePointDistance(0, 0, 4, 0, 2, 0)
		if math.Abs(d-0) > 1e-12 {
			t.Errorf("expected 0, got %f", d)
		}
	})

	t.Run("Vertical segment", func(t *testing.T) {
		d := CalcLinePointDistance(0, 0, 0, 5, 3, 2)
		if math.Abs(d-3) > 1e-12 {
			t.Errorf("expected 3, got %f", d)
		}
	})

	t.Run("Degenerate segment", func(t *testing.T) {
		d := CalcLinePointDistance(1, 1, 1, 1, 4, 5)
		if math.Abs(d-5) > 1e-12 {
			t.Errorf("expected 5, got %f", d)
		}
	})
}

func TestCalcSegmentPointU(t *testing.T) {
	t.Run("Typical cases", func(t *testing.T) {
		u := CalcSegmentPointU(0, 0, 4, 0, 2, 0)
		if math.Abs(u-0.5) > 1e-12 {
			t.Errorf("expected 0.5, got %f", u)
		}
		u = CalcSegmentPointU(0, 0, 4, 0, -1, 0)
		if math.Abs(u-(-0.25)) > 1e-12 {
			t.Errorf("expected -0.25, got %f", u)
		}
		u = CalcSegmentPointU(0, 0, 4, 0, 5, 0)
		if math.Abs(u-1.25) > 1e-12 {
			t.Errorf("expected 1.25, got %f", u)
		}
	})

	t.Run("Degenerate segment", func(t *testing.T) {
		u := CalcSegmentPointU(1, 1, 1, 1, 2, 3)
		if u != 0 {
			t.Errorf("expected 0 for degenerate segment, got %f", u)
		}
	})
}

func TestIntersectionExists(t *testing.T) {
	t.Run("Crossing segments", func(t *testing.T) {
		if !IntersectionExists(0, 0, 4, 4, 0, 4, 4, 0) {
			t.Error("expected true for crossing segments")
		}
	})

	t.Run("Disjoint colinear segments", func(t *testing.T) {
		if IntersectionExists(0, 0, 1, 0, 2, 0, 3, 0) {
			t.Error("expected false for disjoint colinear segments (den=0)")
		}
	})

	t.Run("Endpoint touch", func(t *testing.T) {
		if !IntersectionExists(0, 0, 2, 0, 2, 0, 2, 2) {
			t.Error("expected true for touching at endpoint")
		}
	})

	t.Run("Parallel", func(t *testing.T) {
		if IntersectionExists(0, 0, 2, 0, 0, 1, 2, 1) {
			t.Error("expected false for parallel segments")
		}
	})
}

func TestCalcOrthogonal(t *testing.T) {
	t.Run("Axis-aligned", func(t *testing.T) {
		x, y := CalcOrthogonal(2, 0, 0, 1, 0)
		if math.Abs(x-0) > 1e-12 || math.Abs(y-(-2)) > 1e-12 {
			t.Errorf("expected (0,-2), got (%f,%f)", x, y)
		}
		x, y = CalcOrthogonal(3, 0, 0, 0, 2)
		if math.Abs(x-3) > 1e-12 || math.Abs(y-0) > 1e-12 {
			t.Errorf("expected (3,0), got (%f,%f)", x, y)
		}
	})

	t.Run("Perpendicular and magnitude", func(t *testing.T) {
		tth := 2.0
		x1, y1, x2, y2 := 1.0, 1.0, 4.0, 5.0
		ox, oy := CalcOrthogonal(tth, x1, y1, x2, y2)
		// Perpendicular: dot((dx,dy),(ox,oy)) ≈ 0
		dx, dy := x2-x1, y2-y1
		dot := dx*ox + dy*oy
		if math.Abs(dot) > 1e-10 {
			t.Errorf("expected perpendicular vector, dot=%e", dot)
		}
		// Magnitude ≈ thickness
		mag := math.Hypot(ox, oy)
		if math.Abs(mag-tth) > 1e-10 {
			t.Errorf("expected magnitude %f, got %f", tth, mag)
		}
	})
}

// Benchmark tests
func BenchmarkFastSqrt(b *testing.B) {
	values := []uint32{16, 64, 256, 1024, 4096, 16384, 65536}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FastSqrt(values[i%len(values)])
	}
}

func BenchmarkMathSqrt(b *testing.B) {
	values := []uint32{16, 64, 256, 1024, 4096, 16384, 65536}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		math.Sqrt(float64(values[i%len(values)]))
	}
}

func BenchmarkBesj(b *testing.B) {
	values := []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Besj(values[i%len(values)])
	}
}

func BenchmarkCalcDistance(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalcDistance(0, 0, float64(i%100), float64((i+1)%100))
	}
}

func BenchmarkCalcPolygonArea(b *testing.B) {
	vertices := []PointD{
		{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 5, Y: 15}, {X: 0, Y: 10},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalcPolygonArea(vertices)
	}
}

func TestLookupTables(t *testing.T) {
    t.Run("ElderBitTable correctness", func(t *testing.T) {
        for i := 0; i < 256; i++ {
            var expected uint32
            if i == 0 {
                expected = 0
            } else {
                // floor(log2(i))
                v := i
                for (v >> 1) > 0 {
                    expected++
                    v >>= 1
                }
            }
            if gElderBitTable[i] != expected {
                t.Fatalf("gElderBitTable[%d]=%d, expected %d", i, gElderBitTable[i], expected)
            }
        }
    })

    t.Run("SqrtTable sanity", func(t *testing.T) {
        if len(gSqrtTable) != 1024 {
            t.Fatalf("gSqrtTable length=%d, expected 1024", len(gSqrtTable))
        }
        // Non-decreasing and within 16-bit range (matches AGG's int16u g_sqrt_table[1024])
        prev := uint16(0)
        for idx, v := range gSqrtTable {
            if v < prev {
                t.Fatalf("gSqrtTable not non-decreasing at %d: %d < %d", idx, v, prev)
            }
            // Values should be valid uint16 (no need to check > 65535 since v is uint16)
            prev = v
        }
        // First value should be 0, last should be 65504 (from AGG)
        if gSqrtTable[0] != 0 {
            t.Fatalf("gSqrtTable[0]=%d, expected 0", gSqrtTable[0])
        }
        if gSqrtTable[len(gSqrtTable)-1] != 65504 {
            t.Fatalf("gSqrtTable[1023]=%d, expected 65504", gSqrtTable[len(gSqrtTable)-1])
        }
    })
}
