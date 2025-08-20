package pixfmt

import (
	"math"

	"agg_go/internal/basics"
)

// Numeric constraint for gamma LUT types
type Numeric interface {
	basics.Int8u | basics.Int16u | basics.Int32u | float32 | float64
}

// Gamma LUT constants for default 8-bit gamma tables
const (
	DefaultGammaShift = 8
	DefaultGammaSize  = 1 << DefaultGammaShift // 256
	DefaultGammaMask  = DefaultGammaSize - 1   // 255
	DefaultHiResShift = 8
	DefaultHiResSize  = 1 << DefaultHiResShift // 256
	DefaultHiResMask  = DefaultHiResSize - 1   // 255
)

// GammaLUT provides a high-performance gamma lookup table
// This is a generic implementation matching C++ AGG's gamma_lut template
type GammaLUT[LoResT, HiResT Numeric] struct {
	gamma      float64
	gammaShift uint
	gammaSize  uint
	gammaMask  uint
	hiResShift uint
	hiResSize  uint
	hiResMask  uint
	dirGamma   []HiResT
	invGamma   []LoResT
}

// NewGammaLUT creates a new gamma lookup table with default parameters
func NewGammaLUT[LoResT, HiResT Numeric]() *GammaLUT[LoResT, HiResT] {
	return NewGammaLUTWithShifts[LoResT, HiResT](DefaultGammaShift, DefaultHiResShift)
}

// NewGammaLUTWithShifts creates a new gamma lookup table with custom bit shifts
func NewGammaLUTWithShifts[LoResT, HiResT Numeric](gammaShift, hiResShift uint) *GammaLUT[LoResT, HiResT] {
	lut := &GammaLUT[LoResT, HiResT]{
		gamma:      1.0,
		gammaShift: gammaShift,
		gammaSize:  1 << gammaShift,
		gammaMask:  (1 << gammaShift) - 1,
		hiResShift: hiResShift,
		hiResSize:  1 << hiResShift,
		hiResMask:  (1 << hiResShift) - 1,
	}
	lut.dirGamma = make([]HiResT, lut.gammaSize)
	lut.invGamma = make([]LoResT, lut.hiResSize)

	// Initialize with identity mapping
	lut.initIdentity()
	return lut
}

// NewGammaLUTWithGamma creates a new gamma lookup table with specified gamma value
func NewGammaLUTWithGamma[LoResT, HiResT Numeric](gamma float64) *GammaLUT[LoResT, HiResT] {
	lut := NewGammaLUT[LoResT, HiResT]()
	lut.SetGamma(gamma)
	return lut
}

// initIdentity initializes the tables with identity mapping (gamma = 1.0)
func (lut *GammaLUT[LoResT, HiResT]) initIdentity() {
	// Direct gamma table: scale from low-res to hi-res
	for i := uint(0); i < lut.gammaSize; i++ {
		var scaled uint
		if lut.hiResShift >= lut.gammaShift {
			scaled = i << (lut.hiResShift - lut.gammaShift)
		} else {
			scaled = i >> (lut.gammaShift - lut.hiResShift)
		}

		// Convert to appropriate type
		switch any(&lut.dirGamma[i]).(type) {
		case *basics.Int8u:
			*any(&lut.dirGamma[i]).(*basics.Int8u) = basics.Int8u(scaled)
		case *basics.Int16u:
			*any(&lut.dirGamma[i]).(*basics.Int16u) = basics.Int16u(scaled)
		case *basics.Int32u:
			*any(&lut.dirGamma[i]).(*basics.Int32u) = basics.Int32u(scaled)
		case *float32:
			*any(&lut.dirGamma[i]).(*float32) = float32(scaled)
		case *float64:
			*any(&lut.dirGamma[i]).(*float64) = float64(scaled)
		}
	}

	// Inverse gamma table: scale from hi-res to low-res
	for i := uint(0); i < lut.hiResSize; i++ {
		var scaled uint
		if lut.hiResShift >= lut.gammaShift {
			scaled = i >> (lut.hiResShift - lut.gammaShift)
		} else {
			scaled = i << (lut.gammaShift - lut.hiResShift)
		}

		// Convert to appropriate type
		switch any(&lut.invGamma[i]).(type) {
		case *basics.Int8u:
			*any(&lut.invGamma[i]).(*basics.Int8u) = basics.Int8u(scaled)
		case *basics.Int16u:
			*any(&lut.invGamma[i]).(*basics.Int16u) = basics.Int16u(scaled)
		case *basics.Int32u:
			*any(&lut.invGamma[i]).(*basics.Int32u) = basics.Int32u(scaled)
		case *float32:
			*any(&lut.invGamma[i]).(*float32) = float32(scaled)
		case *float64:
			*any(&lut.invGamma[i]).(*float64) = float64(scaled)
		}
	}
}

// SetGamma sets the gamma value and rebuilds the lookup tables
func (lut *GammaLUT[LoResT, HiResT]) SetGamma(gamma float64) {
	lut.gamma = gamma

	// Build direct gamma table
	for i := uint(0); i < lut.gammaSize; i++ {
		normalized := float64(i) / float64(lut.gammaMask)
		corrected := math.Pow(normalized, gamma)
		value := corrected * float64(lut.hiResMask)
		rounded := basics.URound(value)

		// Convert to appropriate HiResT type
		switch any(&lut.dirGamma[i]).(type) {
		case *basics.Int8u:
			*any(&lut.dirGamma[i]).(*basics.Int8u) = basics.Int8u(rounded)
		case *basics.Int16u:
			*any(&lut.dirGamma[i]).(*basics.Int16u) = basics.Int16u(rounded)
		case *basics.Int32u:
			*any(&lut.dirGamma[i]).(*basics.Int32u) = basics.Int32u(rounded)
		case *float32:
			*any(&lut.dirGamma[i]).(*float32) = float32(rounded)
		case *float64:
			*any(&lut.dirGamma[i]).(*float64) = float64(rounded)
		}
	}

	// Build inverse gamma table
	invGamma := 1.0 / gamma
	for i := uint(0); i < lut.hiResSize; i++ {
		normalized := float64(i) / float64(lut.hiResMask)
		corrected := math.Pow(normalized, invGamma)
		value := corrected * float64(lut.gammaMask)
		rounded := basics.URound(value)

		// Convert to appropriate LoResT type
		switch any(&lut.invGamma[i]).(type) {
		case *basics.Int8u:
			*any(&lut.invGamma[i]).(*basics.Int8u) = basics.Int8u(rounded)
		case *basics.Int16u:
			*any(&lut.invGamma[i]).(*basics.Int16u) = basics.Int16u(rounded)
		case *basics.Int32u:
			*any(&lut.invGamma[i]).(*basics.Int32u) = basics.Int32u(rounded)
		case *float32:
			*any(&lut.invGamma[i]).(*float32) = float32(rounded)
		case *float64:
			*any(&lut.invGamma[i]).(*float64) = float64(rounded)
		}
	}
}

// Gamma returns the current gamma value
func (lut *GammaLUT[LoResT, HiResT]) Gamma() float64 {
	return lut.gamma
}

// Dir performs direct gamma correction lookup
func (lut *GammaLUT[LoResT, HiResT]) Dir(v LoResT) HiResT {
	idx := uint(v) & lut.gammaMask
	return lut.dirGamma[idx]
}

// Inv performs inverse gamma correction lookup
func (lut *GammaLUT[LoResT, HiResT]) Inv(v HiResT) LoResT {
	idx := uint(v) & lut.hiResMask
	return lut.invGamma[idx]
}

// Concrete 8-bit gamma LUT types
type (
	GammaLUT8  = GammaLUT[basics.Int8u, basics.Int8u]
	GammaLUT16 = GammaLUT[basics.Int8u, basics.Int16u]
)

// Constructor functions for common use cases
func NewGammaLUT8() *GammaLUT8 {
	return NewGammaLUT[basics.Int8u, basics.Int8u]()
}

func NewGammaLUT8WithGamma(gamma float64) *GammaLUT8 {
	return NewGammaLUTWithGamma[basics.Int8u, basics.Int8u](gamma)
}

func NewGammaLUT16() *GammaLUT16 {
	return NewGammaLUT[basics.Int8u, basics.Int16u]()
}

func NewGammaLUT16WithGamma(gamma float64) *GammaLUT16 {
	return NewGammaLUTWithGamma[basics.Int8u, basics.Int16u](gamma)
}

// Bridge to make GammaLUT8 compatible with the existing GammaLut interface
// This allows the new implementation to work with existing code

// AGGGammaLUT provides an improved gamma LUT implementation that replaces SimpleGammaLut
type AGGGammaLUT struct {
	lut *GammaLUT8
}

// NewAGGGammaLUT creates a new AGG-compatible gamma LUT with the specified gamma value
func NewAGGGammaLUT(gamma float64) *AGGGammaLUT {
	return &AGGGammaLUT{lut: NewGammaLUT8WithGamma(gamma)}
}

// Dir performs direct gamma correction (implements GammaLut interface)
func (agg *AGGGammaLUT) Dir(v basics.Int8u) basics.Int8u {
	return agg.lut.Dir(v)
}

// Inv performs inverse gamma correction (implements GammaLut interface)
func (agg *AGGGammaLUT) Inv(v basics.Int8u) basics.Int8u {
	return agg.lut.Inv(v)
}

// SetGamma updates the gamma value and rebuilds the tables
func (agg *AGGGammaLUT) SetGamma(gamma float64) {
	agg.lut.SetGamma(gamma)
}

// Gamma returns the current gamma value
func (agg *AGGGammaLUT) Gamma() float64 {
	return agg.lut.Gamma()
}

// sRGB lookup table base for all linear types
type SRGBLUTBase[LinearType Numeric] struct {
	dirTable [256]LinearType
	invTable [256]LinearType
}

// Dir performs sRGB to linear conversion
func (lut *SRGBLUTBase[LinearType]) Dir(v basics.Int8u) LinearType {
	return lut.dirTable[v]
}

// Inv performs linear to sRGB conversion using binary search
func (lut *SRGBLUTBase[LinearType]) Inv(v LinearType) basics.Int8u {
	// Unrolled binary search for optimal performance
	x := basics.Int8u(0)
	if v > lut.invTable[128] {
		x = 128
	}
	if v > lut.invTable[x+64] {
		x += 64
	}
	if v > lut.invTable[x+32] {
		x += 32
	}
	if v > lut.invTable[x+16] {
		x += 16
	}
	if v > lut.invTable[x+8] {
		x += 8
	}
	if v > lut.invTable[x+4] {
		x += 4
	}
	if v > lut.invTable[x+2] {
		x += 2
	}
	if v > lut.invTable[x+1] {
		x += 1
	}
	return x
}

// SRGBLUT provides sRGB conversion for specific linear types
type SRGBLUT[LinearType Numeric] struct {
	SRGBLUTBase[LinearType]
}

// SRGBLUTFloat specialization for float32
type SRGBLUTFloat struct {
	SRGBLUTBase[float32]
}

// NewSRGBLUTFloat creates a new sRGB LUT for float32 values
func NewSRGBLUTFloat() *SRGBLUTFloat {
	lut := &SRGBLUTFloat{}

	// Generate lookup tables
	lut.dirTable[0] = 0
	lut.invTable[0] = 0

	for i := 1; i <= 255; i++ {
		// Floating-point RGB is in range [0,1]
		normalized := float64(i) / 255.0
		lut.dirTable[i] = float32(SRGBToLinear(normalized))
		lut.invTable[i] = float32(SRGBToLinear((float64(i) - 0.5) / 255.0))
	}

	return lut
}

// SRGBLUT16 specialization for 16-bit values
type SRGBLUT16 struct {
	SRGBLUTBase[basics.Int16u]
}

// NewSRGBLUT16 creates a new sRGB LUT for 16-bit values
func NewSRGBLUT16() *SRGBLUT16 {
	lut := &SRGBLUT16{}

	// Generate lookup tables
	lut.dirTable[0] = 0
	lut.invTable[0] = 0

	for i := 1; i <= 255; i++ {
		// 16-bit RGB is in range [0,65535]
		normalized := float64(i) / 255.0
		lut.dirTable[i] = basics.Int16u(65535.0*SRGBToLinear(normalized) + 0.5)
		lut.invTable[i] = basics.Int16u(65535.0*SRGBToLinear((float64(i)-0.5)/255.0) + 0.5)
	}

	return lut
}

// SRGBLUT8 specialization for 8-bit values
type SRGBLUT8 struct {
	SRGBLUTBase[basics.Int8u]
}

// NewSRGBLUT8 creates a new sRGB LUT for 8-bit values
func NewSRGBLUT8() *SRGBLUT8 {
	lut := &SRGBLUT8{}

	// Generate lookup tables
	lut.dirTable[0] = 0
	lut.invTable[0] = 0

	for i := 1; i <= 255; i++ {
		// 8-bit RGB is handled with bidirectional lookup tables
		normalized := float64(i) / 255.0
		lut.dirTable[i] = basics.Int8u(255.0*SRGBToLinear(normalized) + 0.5)
		lut.invTable[i] = basics.Int8u(255.0*LinearToSRGB(normalized) + 0.5)
	}

	return lut
}

// Inv for 8-bit sRGB uses simple lookup instead of binary search
func (lut *SRGBLUT8) Inv(v basics.Int8u) basics.Int8u {
	return lut.invTable[v]
}

// sRGB conversion base for static methods
type SRGBConvBase[T Numeric] struct {
	lut SRGBLUT[T]
}

// Global sRGB conversion instances
var (
	srgbLUTFloat *SRGBLUTFloat
	srgbLUT16    *SRGBLUT16
	srgbLUT8     *SRGBLUT8
)

// Initialize sRGB LUTs on package load
func init() {
	srgbLUTFloat = NewSRGBLUTFloat()
	srgbLUT16 = NewSRGBLUT16()
	srgbLUT8 = NewSRGBLUT8()
}

// SRGBConv provides static conversion methods
type SRGBConv[T Numeric] struct{}

// SRGBConvFloat provides sRGB conversion for float32
type SRGBConvFloat struct{}

// RGBFromSRGB converts sRGB to linear float32
func (SRGBConvFloat) RGBFromSRGB(x basics.Int8u) float32 {
	return srgbLUTFloat.Dir(x)
}

// RGBToSRGB converts linear float32 to sRGB
func (SRGBConvFloat) RGBToSRGB(x float32) basics.Int8u {
	return srgbLUTFloat.Inv(x)
}

// AlphaFromSRGB converts sRGB alpha to linear float32
func (SRGBConvFloat) AlphaFromSRGB(x basics.Int8u) float32 {
	return float32(x) / 255.0
}

// AlphaToSRGB converts linear float32 alpha to sRGB
func (SRGBConvFloat) AlphaToSRGB(x float32) basics.Int8u {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 255
	}
	return basics.Int8u(0.5 + x*255)
}

// SRGBConv16 provides sRGB conversion for 16-bit values
type SRGBConv16 struct{}

// RGBFromSRGB converts sRGB to linear 16-bit
func (SRGBConv16) RGBFromSRGB(x basics.Int8u) basics.Int16u {
	return srgbLUT16.Dir(x)
}

// RGBToSRGB converts linear 16-bit to sRGB
func (SRGBConv16) RGBToSRGB(x basics.Int16u) basics.Int8u {
	return srgbLUT16.Inv(x)
}

// AlphaFromSRGB converts sRGB alpha to linear 16-bit
func (SRGBConv16) AlphaFromSRGB(x basics.Int8u) basics.Int16u {
	return (basics.Int16u(x) << 8) | basics.Int16u(x)
}

// AlphaToSRGB converts linear 16-bit alpha to sRGB
func (SRGBConv16) AlphaToSRGB(x basics.Int16u) basics.Int8u {
	return basics.Int8u(x >> 8)
}

// SRGBConv8 provides sRGB conversion for 8-bit values
type SRGBConv8 struct{}

// RGBFromSRGB converts sRGB to linear 8-bit
func (SRGBConv8) RGBFromSRGB(x basics.Int8u) basics.Int8u {
	return srgbLUT8.Dir(x)
}

// RGBToSRGB converts linear 8-bit to sRGB
func (SRGBConv8) RGBToSRGB(x basics.Int8u) basics.Int8u {
	return srgbLUT8.Inv(x)
}

// AlphaFromSRGB converts sRGB alpha to linear 8-bit (identity)
func (SRGBConv8) AlphaFromSRGB(x basics.Int8u) basics.Int8u {
	return x
}

// AlphaToSRGB converts linear 8-bit alpha to sRGB (identity)
func (SRGBConv8) AlphaToSRGB(x basics.Int8u) basics.Int8u {
	return x
}
