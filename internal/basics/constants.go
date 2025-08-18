package basics

import "math"

// Cover scale enumeration for anti-aliasing
const (
	CoverShift = 8
	CoverSize  = 1 << CoverShift
	CoverMask  = CoverSize - 1
	CoverNone  = 0
	CoverFull  = CoverMask
)

// Poly subpixel scale enumeration
const (
	PolySubpixelShift = 8
	PolySubpixelScale = 1 << PolySubpixelShift
	PolySubpixelMask  = PolySubpixelScale - 1
)

// Filling rule enumeration
type FillingRule int

const (
	FillNonZero FillingRule = iota
	FillEvenOdd
)

// Layer order enumeration for compound rasterization
type LayerOrder int

const (
	LayerUnsorted LayerOrder = iota // Unsorted layers
	LayerDirect                     // Direct layer order
	LayerInverse                    // Inverse layer order
)

// Mathematical constants
const (
	Pi      = math.Pi
	Deg2Rad = Pi / 180.0
	Rad2Deg = 180.0 / Pi
)

// Vertex distance epsilon for geometric calculations
const (
	VertexDistEpsilon   = 1e-14
	IntersectionEpsilon = 1.0e-30
)

// Conversion functions
func Deg2RadF(deg float64) float64 {
	return deg * Deg2Rad
}

func Rad2DegF(rad float64) float64 {
	return rad * Rad2Deg
}

// Rounding functions (from AGG's platform-specific optimizations)
func IRound(v float64) int {
	if v >= 0.0 {
		return int(v + 0.5)
	}
	return int(v - 0.5)
}

func URound(v float64) uint32 {
	if v >= 0.0 {
		return uint32(v + 0.5)
	}
	return 0
}

func IFloor(v float64) int {
	return int(math.Floor(v))
}

func UFloor(v float64) uint32 {
	if v >= 0.0 {
		return uint32(math.Floor(v))
	}
	return 0
}

func ICeil(v float64) int {
	return int(math.Ceil(v))
}

func UCeil(v float64) uint32 {
	if v >= 0.0 {
		return uint32(math.Ceil(v))
	}
	return 0
}

// Saturation template equivalent using generics
type Saturation[T ~int | ~int32 | ~uint | ~uint32] struct {
	limit T
}

func NewSaturation[T ~int | ~int32 | ~uint | ~uint32](limit T) Saturation[T] {
	return Saturation[T]{limit: limit}
}

func (s Saturation[T]) Apply(v T) T {
	if v > s.limit {
		return s.limit
	}
	return v
}

// IRound performs saturation-aware rounding for floating point values
func (s Saturation[T]) IRound(v float64) T {
	limit := float64(s.limit)

	// Use type switch to determine if T is unsigned
	var zero T
	switch any(zero).(type) {
	case uint, uint32:
		// T is unsigned, clamp negative values to 0
		if v < 0 {
			return 0
		}
	default:
		// T is signed (int, int32), use original AGG behavior
		if v < -limit {
			return -s.limit
		}
	}

	if v > limit {
		return s.limit
	}
	return T(IRound(v))
}

// MulOne template equivalent with shift-based multiplication
type MulOne[T ~int | ~int32 | ~uint | ~uint32] struct {
	shift int
}

func NewMulOne[T ~int | ~int32 | ~uint | ~uint32](shift int) MulOne[T] {
	return MulOne[T]{shift: shift}
}

func (m MulOne[T]) Apply(v T) T {
	return v >> m.shift
}

// Mul performs multiplication with shift optimization
func (m MulOne[T]) Mul(a, b T) T {
	q := a*b + (1 << (m.shift - 1))
	return (q + (q >> m.shift)) >> m.shift
}

// Abs returns the absolute value of x.
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// SaturationIRound performs saturation-aware rounding with given limit.
func SaturationIRound(v float64, limit int) int {
	saturation := NewSaturation(limit)
	return saturation.IRound(v)
}
