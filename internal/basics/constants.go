package basics

import "math"

// Math functions for compatibility
func Sqrt(x float64) float64 {
	return math.Sqrt(x)
}

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

// Saturation functions for different types
// These replace the generic Saturation[T] type to avoid runtime type switches

// SaturationInt provides saturation clamping for int values
type SaturationInt struct {
	limit int
}

func NewSaturationInt(limit int) SaturationInt {
	return SaturationInt{limit: limit}
}

func (s SaturationInt) Apply(v int) int {
	if v > s.limit {
		return s.limit
	}
	return v
}

// IRound performs saturation-aware rounding for floating point values
func (s SaturationInt) IRound(v float64) int {
	limit := float64(s.limit)
	if v < -limit {
		return -s.limit
	}
	if v > limit {
		return s.limit
	}
	return IRound(v)
}

// SaturationInt32 provides saturation clamping for int32 values
type SaturationInt32 struct {
	limit int32
}

func NewSaturationInt32(limit int32) SaturationInt32 {
	return SaturationInt32{limit: limit}
}

func (s SaturationInt32) Apply(v int32) int32 {
	if v > s.limit {
		return s.limit
	}
	return v
}

// IRound performs saturation-aware rounding for floating point values
func (s SaturationInt32) IRound(v float64) int32 {
	limit := float64(s.limit)
	if v < -limit {
		return -s.limit
	}
	if v > limit {
		return s.limit
	}
	return int32(IRound(v))
}

// SaturationUint provides saturation clamping for uint values
type SaturationUint struct {
	limit uint
}

func NewSaturationUint(limit uint) SaturationUint {
	return SaturationUint{limit: limit}
}

func (s SaturationUint) Apply(v uint) uint {
	if v > s.limit {
		return s.limit
	}
	return v
}

// IRound performs saturation-aware rounding for floating point values
func (s SaturationUint) IRound(v float64) uint {
	if v < 0 {
		return 0
	}
	limit := float64(s.limit)
	if v > limit {
		return s.limit
	}
	return uint(IRound(v))
}

// SaturationUint32 provides saturation clamping for uint32 values
type SaturationUint32 struct {
	limit uint32
}

func NewSaturationUint32(limit uint32) SaturationUint32 {
	return SaturationUint32{limit: limit}
}

func (s SaturationUint32) Apply(v uint32) uint32 {
	if v > s.limit {
		return s.limit
	}
	return v
}

// IRound performs saturation-aware rounding for floating point values
func (s SaturationUint32) IRound(v float64) uint32 {
	if v < 0 {
		return 0
	}
	limit := float64(s.limit)
	if v > limit {
		return s.limit
	}
	return uint32(IRound(v))
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
	saturation := NewSaturationInt(limit)
	return saturation.IRound(v)
}
