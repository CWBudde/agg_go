package gpc

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"
)

var rng *rand.Rand

func init() {
	// Create a seeded random number generator for reproducible benchmarks
	rng = rand.New(rand.NewSource(42))
}

// Helper function to create a complex polygon with many vertices
func createComplexPolygon(numVertices int, radius float64) *GPCPolygon {
	polygon := NewGPCPolygon()
	contour := NewGPCVertexList(numVertices)

	for i := 0; i < numVertices; i++ {
		angle := 2 * math.Pi * float64(i) / float64(numVertices)
		// Add some randomness to make it more realistic
		r := radius + rng.Float64()*radius*0.1
		x := r * math.Cos(angle)
		y := r * math.Sin(angle)
		contour.AddVertex(x, y)
	}

	polygon.AddContour(contour, false)
	return polygon
}

// Helper function to create a star-shaped polygon
func createStarPolygon(numPoints int, outerRadius, innerRadius float64) *GPCPolygon {
	polygon := NewGPCPolygon()
	contour := NewGPCVertexList(numPoints * 2)

	for i := 0; i < numPoints; i++ {
		// Outer point
		outerAngle := 2 * math.Pi * float64(i) / float64(numPoints)
		contour.AddVertex(
			outerRadius*math.Cos(outerAngle),
			outerRadius*math.Sin(outerAngle),
		)

		// Inner point
		innerAngle := outerAngle + math.Pi/float64(numPoints)
		contour.AddVertex(
			innerRadius*math.Cos(innerAngle),
			innerRadius*math.Sin(innerAngle),
		)
	}

	polygon.AddContour(contour, false)
	return polygon
}

// Helper function to create a random polygon
func createRandomPolygon(numVertices int, bounds float64) *GPCPolygon {
	polygon := NewGPCPolygon()
	contour := NewGPCVertexList(numVertices)

	for i := 0; i < numVertices; i++ {
		x := (rng.Float64() - 0.5) * bounds * 2
		y := (rng.Float64() - 0.5) * bounds * 2
		contour.AddVertex(x, y)
	}

	polygon.AddContour(contour, false)
	return polygon
}

func BenchmarkGPCVertex_Equal(b *testing.B) {
	v1 := GPCVertex{1.234567, 2.345678}
	v2 := GPCVertex{1.234568, 2.345679}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v1.Equal(v2)
	}
}

func BenchmarkGPCVertexList_AddVertex(b *testing.B) {
	benchmarks := []struct {
		name     string
		capacity int
	}{
		{"Small", 10},
		{"Medium", 100},
		{"Large", 1000},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				vl := NewGPCVertexList(bm.capacity)
				for j := 0; j < bm.capacity; j++ {
					vl.AddVertex(float64(j), float64(j))
				}
			}
		})
	}
}

func BenchmarkGPCPolygon_AddContour(b *testing.B) {
	benchmarks := []struct {
		name        string
		numContours int
		numVertices int
	}{
		{"Few_Small", 10, 10},
		{"Few_Large", 10, 1000},
		{"Many_Small", 100, 10},
		{"Many_Large", 100, 100},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				polygon := NewGPCPolygon()
				for j := 0; j < bm.numContours; j++ {
					contour := createComplexPolygon(bm.numVertices, 100.0).Contours[0]
					polygon.AddContour(contour, false)
				}
			}
		})
	}
}

func BenchmarkGPCPolygon_Validate(b *testing.B) {
	benchmarks := []struct {
		name    string
		polygon *GPCPolygon
	}{
		{"Simple", createTrianglePolygon(0, 0, 1, 0, 0, 1)},
		{"Complex", createComplexPolygon(100, 50.0)},
		{"WithHoles", createPolygonWithHole()},
		{"Star", createStarPolygon(10, 50.0, 25.0)},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = bm.polygon.Validate()
			}
		})
	}
}

func BenchmarkWritePolygon(b *testing.B) {
	benchmarks := []struct {
		name    string
		polygon *GPCPolygon
	}{
		{"Triangle", createTrianglePolygon(0, 0, 1, 0, 0, 1)},
		{"Rectangle", createRectanglePolygon(0, 0, 10, 10)},
		{"Complex_10", createComplexPolygon(10, 50.0)},
		{"Complex_100", createComplexPolygon(100, 50.0)},
		{"Complex_1000", createComplexPolygon(1000, 50.0)},
		{"WithHoles", createPolygonWithHole()},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			var buf bytes.Buffer
			for i := 0; i < b.N; i++ {
				buf.Reset()
				_ = WritePolygon(&buf, bm.polygon, false)
			}
		})
	}
}

func BenchmarkPolygonClip(b *testing.B) {
	// Create test polygons of various complexities
	simple1 := createRectanglePolygon(0, 0, 10, 10)
	simple2 := createRectanglePolygon(5, 5, 15, 15)

	complex1 := createComplexPolygon(50, 50.0)
	complex2 := createComplexPolygon(50, 30.0)

	star1 := createStarPolygon(8, 40.0, 20.0)
	star2 := createStarPolygon(6, 35.0, 15.0)

	benchmarks := []struct {
		name      string
		operation GPCOp
		subject   *GPCPolygon
		clip      *GPCPolygon
	}{
		{"Simple_Union", GPCUnion, simple1, simple2},
		{"Simple_Intersection", GPCInt, simple1, simple2},
		{"Simple_Difference", GPCDiff, simple1, simple2},
		{"Simple_XOR", GPCXor, simple1, simple2},

		{"Complex_Union", GPCUnion, complex1, complex2},
		{"Complex_Intersection", GPCInt, complex1, complex2},
		{"Complex_Difference", GPCDiff, complex1, complex2},
		{"Complex_XOR", GPCXor, complex1, complex2},

		{"Star_Union", GPCUnion, star1, star2},
		{"Star_Intersection", GPCInt, star1, star2},
		{"Star_Difference", GPCDiff, star1, star2},
		{"Star_XOR", GPCXor, star1, star2},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = PolygonClip(bm.operation, bm.subject, bm.clip)
			}
		})
	}
}

func BenchmarkTristripClip(b *testing.B) {
	subject := createStarPolygon(8, 40.0, 20.0)
	clip := createComplexPolygon(20, 30.0)

	operations := []GPCOp{GPCUnion, GPCInt, GPCDiff, GPCXor}

	for _, op := range operations {
		b.Run(op.String(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = TristripClip(op, subject, clip)
			}
		})
	}
}

func BenchmarkPolygonToTristrip(b *testing.B) {
	benchmarks := []struct {
		name    string
		polygon *GPCPolygon
	}{
		{"Triangle", createTrianglePolygon(0, 0, 1, 0, 0, 1)},
		{"Rectangle", createRectanglePolygon(0, 0, 10, 10)},
		{"Complex_10", createComplexPolygon(10, 50.0)},
		{"Complex_50", createComplexPolygon(50, 50.0)},
		{"Complex_100", createComplexPolygon(100, 50.0)},
		{"Star_5", createStarPolygon(5, 40.0, 20.0)},
		{"Star_10", createStarPolygon(10, 40.0, 20.0)},
		{"WithHoles", createPolygonWithHole()},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = PolygonToTristrip(bm.polygon)
			}
		})
	}
}

func BenchmarkHelperFunctions(b *testing.B) {
	vertices := make([]GPCVertex, 100)
	for i := range vertices {
		angle := 2 * math.Pi * float64(i) / float64(len(vertices))
		vertices[i] = GPCVertex{
			X: 50.0 * math.Cos(angle),
			Y: 50.0 * math.Sin(angle),
		}
	}

	b.Run("eq", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = eq(1.234567, 1.234568)
		}
	})

	b.Run("isClockwise", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = isClockwise(vertices)
		}
	})

	contour := NewGPCVertexList(len(vertices))
	for _, v := range vertices {
		contour.AddVertex(v.X, v.Y)
	}

	b.Run("validateContourWinding", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = validateContourWinding(contour, true)
		}
	})
}

func BenchmarkMemoryAllocation(b *testing.B) {
	b.Run("VertexList_Creation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			vl := NewGPCVertexList(1000)
			for j := 0; j < 1000; j++ {
				vl.AddVertex(float64(j), float64(j))
			}
		}
	})

	b.Run("Polygon_Creation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			polygon := NewGPCPolygon()
			for j := 0; j < 10; j++ {
				contour := NewGPCVertexList(100)
				for k := 0; k < 100; k++ {
					contour.AddVertex(float64(k), float64(k))
				}
				polygon.AddContour(contour, false)
			}
		}
	})
}

func BenchmarkScalability(b *testing.B) {
	sizes := []int{10, 50, 100, 500, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("ComplexPolygon_%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = createComplexPolygon(size, 100.0)
			}
		})

		b.Run(fmt.Sprintf("ClipOperation_%d", size), func(b *testing.B) {
			subject := createComplexPolygon(size, 100.0)
			clip := createComplexPolygon(size/2, 50.0)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = PolygonClip(GPCUnion, subject, clip)
			}
		})
	}
}

func BenchmarkRandomizedOperations(b *testing.B) {
	// Create a set of random polygons for testing
	randomPolygons := make([]*GPCPolygon, 100)
	for i := range randomPolygons {
		numVertices := rng.Intn(50) + 3 // 3 to 52 vertices
		randomPolygons[i] = createRandomPolygon(numVertices, 100.0)
	}

	operations := []GPCOp{GPCUnion, GPCInt, GPCDiff, GPCXor}

	b.Run("RandomPairs", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			subject := randomPolygons[rng.Intn(len(randomPolygons))]
			clip := randomPolygons[rng.Intn(len(randomPolygons))]
			operation := operations[rng.Intn(len(operations))]

			_, _ = PolygonClip(operation, subject, clip)
		}
	})
}

// Benchmark to measure performance characteristics over time
func BenchmarkPerformanceRegression(b *testing.B) {
	// This benchmark can be used to detect performance regressions
	// by comparing against baseline measurements

	subject := createComplexPolygon(100, 100.0)
	clip := createStarPolygon(20, 80.0, 40.0)

	start := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = PolygonClip(GPCUnion, subject, clip)
	}

	duration := time.Since(start)

	// Log timing information for regression analysis
	b.Logf("Average operation time: %v", duration/time.Duration(b.N))
}

// Memory allocation benchmark
func BenchmarkMemoryUsage(b *testing.B) {
	b.Run("LargePolygon", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			polygon := createComplexPolygon(1000, 100.0)
			_ = polygon.Validate()
		}
	})

	b.Run("ManySmallPolygons", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			polygons := make([]*GPCPolygon, 100)
			for j := range polygons {
				polygons[j] = createComplexPolygon(10, 10.0)
			}
		}
	})
}
