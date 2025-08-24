package gamma

import (
	"math"
)

// GammaFunction represents a gamma correction function interface
type GammaFunction interface {
	Apply(x float64) float64
}

// GammaNone provides no gamma correction (identity function)
type GammaNone struct{}

// NewGammaNone creates a new identity gamma function
func NewGammaNone() *GammaNone {
	return &GammaNone{}
}

// Apply returns the input value unchanged
func (g *GammaNone) Apply(x float64) float64 {
	return x
}

// GammaPower provides power-based gamma correction
type GammaPower struct {
	gamma float64
}

// NewGammaPower creates a new power gamma function
func NewGammaPower(gamma float64) *GammaPower {
	return &GammaPower{gamma: gamma}
}

// SetGamma sets the gamma value
func (g *GammaPower) SetGamma(gamma float64) {
	g.gamma = gamma
}

// Gamma returns the current gamma value
func (g *GammaPower) Gamma() float64 {
	return g.gamma
}

// Apply applies power gamma correction
func (g *GammaPower) Apply(x float64) float64 {
	return math.Pow(x, g.gamma)
}

// GammaThreshold provides threshold-based binary gamma correction
type GammaThreshold struct {
	threshold float64
}

// NewGammaThreshold creates a new threshold gamma function
func NewGammaThreshold(threshold float64) *GammaThreshold {
	return &GammaThreshold{threshold: threshold}
}

// SetThreshold sets the threshold value
func (g *GammaThreshold) SetThreshold(threshold float64) {
	g.threshold = threshold
}

// Threshold returns the current threshold value
func (g *GammaThreshold) Threshold() float64 {
	return g.threshold
}

// Apply applies threshold gamma correction (binary output)
func (g *GammaThreshold) Apply(x float64) float64 {
	if x < g.threshold {
		return 0.0
	}
	return 1.0
}

// GammaLinear provides linear interpolation between start and end points
type GammaLinear struct {
	start float64
	end   float64
}

// NewGammaLinear creates a new linear gamma function
func NewGammaLinear(start, end float64) *GammaLinear {
	return &GammaLinear{start: start, end: end}
}

// Set sets both start and end values
func (g *GammaLinear) Set(start, end float64) {
	g.start = start
	g.end = end
}

// SetStart sets the start value
func (g *GammaLinear) SetStart(start float64) {
	g.start = start
}

// SetEnd sets the end value
func (g *GammaLinear) SetEnd(end float64) {
	g.end = end
}

// Start returns the start value
func (g *GammaLinear) Start() float64 {
	return g.start
}

// End returns the end value
func (g *GammaLinear) End() float64 {
	return g.end
}

// Apply applies linear gamma correction
func (g *GammaLinear) Apply(x float64) float64 {
	if x < g.start {
		return 0.0
	}
	if x > g.end {
		return 1.0
	}
	return (x - g.start) / (g.end - g.start)
}

// GammaMultiply provides multiplication-based gamma correction with clamping
type GammaMultiply struct {
	multiplier float64
}

// NewGammaMultiply creates a new multiply gamma function
func NewGammaMultiply(multiplier float64) *GammaMultiply {
	return &GammaMultiply{multiplier: multiplier}
}

// SetValue sets the multiplier value
func (g *GammaMultiply) SetValue(multiplier float64) {
	g.multiplier = multiplier
}

// Value returns the current multiplier value
func (g *GammaMultiply) Value() float64 {
	return g.multiplier
}

// Apply applies multiplication gamma correction with clamping
func (g *GammaMultiply) Apply(x float64) float64 {
	y := x * g.multiplier
	if y > 1.0 {
		return 1.0
	}
	return y
}

// sRGB conversion functions (reference implementation)
// These match the C++ AGG sRGB_to_linear and linear_to_sRGB functions

// SRGBToLinear converts sRGB value to linear RGB
func SRGBToLinear(x float64) float64 {
	if x <= 0.04045 {
		return x / 12.92
	}
	return math.Pow((x+0.055)/1.055, 2.4)
}

// LinearToSRGB converts linear RGB value to sRGB
func LinearToSRGB(x float64) float64 {
	if x <= 0.0031308 {
		return x * 12.92
	}
	return 1.055*math.Pow(x, 1.0/2.4) - 0.055
}
