package basics

import "math"

// Cross product calculation
func CrossProduct(x1, y1, x2, y2, x, y float64) float64 {
	return (x-x2)*(y2-y1) - (y-y2)*(x2-x1)
}

// Distance calculations
func CalcDistance(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}

func CalcSqDistance(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return dx*dx + dy*dy
}

// Point in triangle test
func PointInTriangle(x1, y1, x2, y2, x3, y3, x, y float64) bool {
	cp1 := CrossProduct(x1, y1, x2, y2, x, y)
	cp2 := CrossProduct(x2, y2, x3, y3, x, y)
	cp3 := CrossProduct(x3, y3, x1, y1, x, y)
	return (cp1*cp2 >= 0) && (cp2*cp3 >= 0)
}

// Line point distance calculation
func CalcLinePointDistance(x1, y1, x2, y2, x, y float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	d := math.Sqrt(dx*dx + dy*dy)
	if d < VertexDistEpsilon {
		return CalcDistance(x1, y1, x, y)
	}
	return math.Abs((x-x1)*dy-(y-y1)*dx) / d
}

// Segment point u parameter calculation
func CalcSegmentPointU(x1, y1, x2, y2, x, y float64) float64 {
	dx := x2 - x1
	dy := y2 - y1

	if dx == 0 && dy == 0 {
		return 0
	}

	pdx := x - x1
	pdy := y - y1

	return (pdx*dx + pdy*dy) / (dx*dx + dy*dy)
}

// Segment point squared distance calculation
func CalcSegmentPointSqDistance(x1, y1, x2, y2, x, y float64) float64 {
	u := CalcSegmentPointU(x1, y1, x2, y2, x, y)

	if u <= 0 {
		return CalcSqDistance(x, y, x1, y1)
	}
	if u >= 1 {
		return CalcSqDistance(x, y, x2, y2)
	}

	ix := x1 + u*(x2-x1)
	iy := y1 + u*(y2-y1)
	return CalcSqDistance(x, y, ix, iy)
}

// Line intersection calculation
func CalcIntersection(ax, ay, bx, by, cx, cy, dx, dy float64) (x, y float64, ok bool) {
	num := (ay-cy)*(dx-cx) - (ax-cx)*(dy-cy)
	den := (bx-ax)*(dy-cy) - (by-ay)*(dx-cx)

	if math.Abs(den) < IntersectionEpsilon {
		return 0, 0, false
	}

	r := num / den
	x = ax + r*(bx-ax)
	y = ay + r*(by-ay)
	return x, y, true
}

// Check if intersection exists
func IntersectionExists(ax, ay, bx, by, cx, cy, dx, dy float64) bool {
	num := (ay-cy)*(dx-cx) - (ax-cx)*(dy-cy)
	den := (bx-ax)*(dy-cy) - (by-ay)*(dx-cx)

	if math.Abs(den) < IntersectionEpsilon {
		return false
	}

	r := num / den
	s := ((ay-cy)*(bx-ax) - (ax-cx)*(by-ay)) / den

	return r >= 0 && r <= 1 && s >= 0 && s <= 1
}

// Calculate orthogonal vector
func CalcOrthogonal(thickness, x1, y1, x2, y2 float64) (x, y float64) {
	dx := x2 - x1
	dy := y2 - y1
	d := math.Sqrt(dx*dx + dy*dy)

	if d < VertexDistEpsilon {
		return 0, 0
	}

	x = thickness * dy / d
	y = -thickness * dx / d
	return x, y
}

// Dilate triangle
func DilateTriangle(x1, y1, x2, y2, x3, y3, d float64) (x1o, y1o, x2o, y2o, x3o, y3o float64) {
	dx1, dy1 := CalcOrthogonal(d, x1, y1, x2, y2)
	dx2, dy2 := CalcOrthogonal(d, x2, y2, x3, y3)
	dx3, dy3 := CalcOrthogonal(d, x3, y3, x1, y1)

	x1o = x1 + dx1 + dx3
	y1o = y1 + dy1 + dy3
	x2o = x2 + dx1 + dx2
	y2o = y2 + dy1 + dy2
	x3o = x3 + dx2 + dx3
	y3o = y3 + dy2 + dy3

	return
}

// Calculate triangle area
func CalcTriangleArea(x1, y1, x2, y2, x3, y3 float64) float64 {
	return math.Abs((x1*(y2-y3) + x2*(y3-y1) + x3*(y1-y2)) * 0.5)
}

// Calculate polygon area using the shoelace formula
func CalcPolygonArea[T ~float64](vertices []Point[T]) float64 {
	if len(vertices) < 3 {
		return 0
	}

	area := 0.0
	n := len(vertices)

	for i := 0; i < n; i++ {
		j := (i + 1) % n
		area += float64(vertices[i].X * vertices[j].Y)
		area -= float64(vertices[j].X * vertices[i].Y)
	}

	return math.Abs(area) * 0.5
}

// Fast sqrt lookup table (1024 entries for fast square root approximation)
var gSqrtTable = [1024]uint32{
	0, 16, 22, 27, 32, 35, 39, 42, 45, 48, 50, 53, 55, 57, 59, 61,
	64, 65, 67, 69, 71, 73, 75, 76, 78, 80, 81, 83, 84, 86, 87, 89,
	90, 91, 93, 94, 96, 97, 98, 99, 101, 102, 103, 104, 106, 107, 108, 109,
	110, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126,
	128, 128, 129, 130, 131, 132, 133, 134, 135, 136, 137, 138, 139, 140, 141, 142,
	143, 144, 144, 145, 146, 147, 148, 149, 150, 150, 151, 152, 153, 154, 155, 155,
	156, 157, 158, 159, 160, 160, 161, 162, 163, 164, 164, 165, 166, 167, 167, 168,
	169, 170, 170, 171, 172, 173, 173, 174, 175, 176, 176, 177, 178, 178, 179, 180,
	181, 181, 182, 183, 183, 184, 185, 185, 186, 187, 187, 188, 189, 189, 190, 191,
	192, 192, 193, 193, 194, 195, 195, 196, 197, 197, 198, 199, 199, 200, 201, 201,
	202, 203, 203, 204, 204, 205, 206, 206, 207, 208, 208, 209, 209, 210, 211, 211,
	212, 212, 213, 214, 214, 215, 215, 216, 217, 217, 218, 218, 219, 219, 220, 221,
	221, 222, 222, 223, 224, 224, 225, 225, 226, 226, 227, 227, 228, 229, 229, 230,
	230, 231, 231, 232, 232, 233, 234, 234, 235, 235, 236, 236, 237, 237, 238, 238,
	239, 240, 240, 241, 241, 242, 242, 243, 243, 244, 244, 245, 245, 246, 246, 247,
	247, 248, 248, 249, 249, 250, 250, 251, 251, 252, 252, 253, 253, 254, 254, 255,
}

// Elder bit table for fast computation
var gElderBitTable = [256]uint32{
	0, 0, 1, 1, 2, 2, 2, 2, 3, 3, 3, 3, 3, 3, 3, 3,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5,
	5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5,
	6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6,
	6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6,
	6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6,
	6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6,
	7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
	7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
	7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
	7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
	7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
	7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
	7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
	7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
}

// FastSqrt provides fast square root approximation using lookup tables
func FastSqrt(val uint32) uint32 {
	if val == 0 {
		return 0
	}

	var t uint32

	if val >= 0x10000 {
		if val >= 0x1000000 {
			if val >= 0x10000000 {
				if val >= 0x40000000 {
					t = gSqrtTable[val>>24] << 8
				} else {
					t = gSqrtTable[val>>22] << 7
				}
			} else {
				if val >= 0x4000000 {
					t = gSqrtTable[val>>20] << 6
				} else {
					t = gSqrtTable[val>>18] << 5
				}
			}
		} else {
			if val >= 0x100000 {
				if val >= 0x400000 {
					t = gSqrtTable[val>>16] << 4
				} else {
					t = gSqrtTable[val>>14] << 3
				}
			} else {
				if val >= 0x40000 {
					t = gSqrtTable[val>>12] << 2
				} else {
					t = gSqrtTable[val>>10] << 1
				}
			}
		}
	} else {
		if val >= 0x100 {
			if val >= 0x1000 {
				if val >= 0x4000 {
					t = gSqrtTable[val>>8]
				} else {
					t = gSqrtTable[val>>6] >> 1
				}
			} else {
				if val >= 0x400 {
					t = gSqrtTable[val>>4] >> 2
				} else {
					t = gSqrtTable[val>>2] >> 3
				}
			}
		} else {
			if val >= 0x10 {
				if val >= 0x40 {
					t = gSqrtTable[val] >> 4
				} else {
					t = gSqrtTable[val<<2] >> 5
				}
			} else {
				if val >= 0x4 {
					t = gSqrtTable[val<<4] >> 6
				} else {
					t = gSqrtTable[val<<6] >> 7
				}
			}
		}
	}

	// Newton-Raphson refinement
	t = (val/t + t) >> 1
	t = (val/t + t) >> 1

	return t
}

// Besj calculates the Bessel function of the first kind (simplified implementation)
func Besj(x float64) float64 {
	if x == 0.0 {
		return 1.0
	}

	// Use series expansion for small values
	if math.Abs(x) < 3.0 {
		x2 := x * x / 4.0
		term := 1.0
		sum := 1.0

		for i := 1; i < 50; i++ {
			term *= -x2 / (float64(i) * float64(i))
			sum += term

			if math.Abs(term) < 1e-12 {
				break
			}
		}

		return sum
	}

	// For larger values, use asymptotic approximation
	return math.Sqrt(2.0/(math.Pi*x)) * math.Cos(x-math.Pi/4.0)
}
