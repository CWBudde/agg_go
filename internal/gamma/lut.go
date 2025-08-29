package gamma

import (
	"math"

	"agg_go/internal/basics"
)

// Unsigned is used to constrain LoResT/HiResT to unsigned integer-like types.
type Unsigned interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// GammaLUT is a Go-idiomatic equivalent of AGG's gamma_lut template.
// It does not use reflection or type switches; everything is resolved via generics.
type GammaLUT[LoResT Unsigned, HiResT Unsigned] struct {
	gamma      float64
	gammaShift uint
	gammaSize  int
	gammaMask  uint

	hiResShift uint
	hiResSize  int
	hiResMask  uint

	dir []HiResT // size = gammaSize
	inv []LoResT // size = hiResSize
}

// NewGammaLUTWithShifts directly matches AGG's template non-type params (GammaShift, HiResShift).
func NewGammaLUTWithShifts[LoResT Unsigned, HiResT Unsigned](gammaShift, hiResShift uint) *GammaLUT[LoResT, HiResT] {
	l := &GammaLUT[LoResT, HiResT]{
		gamma:      1.0,
		gammaShift: gammaShift,
		gammaSize:  1 << gammaShift,
		gammaMask:  (1 << gammaShift) - 1,

		hiResShift: hiResShift,
		hiResSize:  1 << hiResShift,
		hiResMask:  (1 << hiResShift) - 1,

		dir: make([]HiResT, 1<<gammaShift),
		inv: make([]LoResT, 1<<hiResShift),
	}

	// Identity mapping (exactly like AGG's default ctor).
	for i := 0; i < l.gammaSize; i++ {
		var scaled uint
		if l.hiResShift >= l.gammaShift {
			scaled = uint(i) << (l.hiResShift - l.gammaShift)
		} else {
			scaled = uint(i) >> (l.gammaShift - l.hiResShift)
		}
		l.dir[i] = HiResT(scaled)
	}
	for i := 0; i < l.hiResSize; i++ {
		var scaled uint
		if l.hiResShift >= l.gammaShift {
			scaled = uint(i) >> (l.hiResShift - l.gammaShift)
		} else {
			scaled = uint(i) << (l.gammaShift - l.hiResShift)
		}
		l.inv[i] = LoResT(scaled)
	}
	return l
}

// NewGammaLUT (defaults) matches AGG's template defaults: LoRes=int8u, HiRes=int8u, shifts=8/8.
func NewGammaLUT8() *GammaLUT[basics.Int8u, basics.Int8u] {
	return NewGammaLUTWithShifts[basics.Int8u, basics.Int8u](8, 8)
}

// NewGammaLUT16 matches the common LoRes=8, HiRes=16 flavor in AGG (used for hi-precision).
func NewGammaLUT16() *GammaLUT[basics.Int8u, basics.Int16u] {
	return NewGammaLUTWithShifts[basics.Int8u, basics.Int16u](8, 16)
}

// NewGammaLUT8WithGamma / NewGammaLUT16WithGamma convenience constructors.
func NewGammaLUT8WithGamma(g float64) *GammaLUT[basics.Int8u, basics.Int8u] {
	l := NewGammaLUT8()
	l.SetGamma(g)
	return l
}

func NewGammaLUT16WithGamma(g float64) *GammaLUT[basics.Int8u, basics.Int16u] {
	l := NewGammaLUT16()
	l.SetGamma(g)
	return l
}

// SetGamma rebuilds the tables using the power-law gamma, mirroring AGG.
func (l *GammaLUT[LoResT, HiResT]) SetGamma(g float64) {
	if !(g > 0) || math.IsNaN(g) || math.IsInf(g, 0) {
		g = 1.0
	}
	l.gamma = g

	gMax := float64(l.gammaMask)
	hMax := float64(l.hiResMask)

	// Direct table: y = pow(x, gamma) * hiResMask, with x in [0..1].
	for i := 0; i < l.gammaSize; i++ {
		x := float64(i) / gMax
		y := math.Pow(x, g) * hMax

		u := uint(basics.URound(y))
		if u > l.hiResMask {
			u = l.hiResMask
		}
		l.dir[i] = HiResT(u)
	}

	// Inverse table: y = pow(x, 1/gamma) * gammaMask, with x in [0..1].
	invG := 1.0 / g
	for i := 0; i < l.hiResSize; i++ {
		x := float64(i) / hMax
		y := math.Pow(x, invG) * gMax

		u := uint(basics.URound(y))
		if u > l.gammaMask {
			u = l.gammaMask
		}
		l.inv[i] = LoResT(u)
	}
}

func (l *GammaLUT[LoResT, HiResT]) Gamma() float64 {
	return l.gamma
}

func (l *GammaLUT[LoResT, HiResT]) Dir(v LoResT) HiResT {
	idx := int(uint(v) & l.gammaMask)
	return l.dir[idx]
}

func (l *GammaLUT[LoResT, HiResT]) Inv(v HiResT) LoResT {
	idx := int(uint(v) & l.hiResMask)
	return l.inv[idx]
}
