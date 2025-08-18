// Package outline provides anti-aliased outline rendering functionality.
// This implements a port of AGG's line_profile_aa class.
package outline

import (
	"agg_go/internal/array"
	"agg_go/internal/basics"
	"agg_go/internal/primitives"
)

// Subpixel scale constants for line profiles
const (
	SubpixelShift = primitives.LineSubpixelShift // 8
	SubpixelScale = primitives.LineSubpixelScale // 256
	SubpixelMask  = primitives.LineSubpixelMask  // 255
)

// AA scale constants
const (
	AAShift = 8
	AAScale = 1 << AAShift // 256
	AAMask  = AAScale - 1  // 255
)

// ValueType represents the profile value type
type ValueType = basics.Int8u

// GammaFunction interface for gamma correction functions
type GammaFunction interface {
	Call(x float64) float64
}

// LineProfileAA provides anti-aliased line profile generation.
// This is equivalent to AGG's line_profile_aa class.
type LineProfileAA struct {
	profile       array.PodArray[ValueType] // Profile array
	gamma         [AAScale]ValueType        // Gamma table
	subpixelWidth int                       // Subpixel width
	minWidth      float64                   // Minimum width
	smootherWidth float64                   // Smoother width
}

// NewLineProfileAA creates a new line profile with default gamma.
func NewLineProfileAA() *LineProfileAA {
	lp := &LineProfileAA{
		subpixelWidth: 0,
		minWidth:      1.0,
		smootherWidth: 1.0,
	}

	// Initialize linear gamma
	for i := 0; i < AAScale; i++ {
		lp.gamma[i] = ValueType(i)
	}

	return lp
}

// NewLineProfileAAWithGamma creates a new line profile with width and gamma function.
func NewLineProfileAAWithGamma(w float64, gammaFn GammaFunction) *LineProfileAA {
	lp := NewLineProfileAA()
	lp.SetGamma(gammaFn)
	lp.Width(w)
	return lp
}

// MinWidth sets the minimum width.
func (lp *LineProfileAA) MinWidth(w float64) {
	lp.minWidth = w
}

// SmootherWidth sets the smoother width.
func (lp *LineProfileAA) SmootherWidth(w float64) {
	lp.smootherWidth = w
}

// SetGamma sets the gamma correction function.
func (lp *LineProfileAA) SetGamma(gammaFn GammaFunction) {
	for i := 0; i < AAScale; i++ {
		val := gammaFn.Call(float64(i)/AAMask) * AAMask
		lp.gamma[i] = ValueType(basics.URound(val))
	}
}

// Width sets the line width and generates the profile.
func (lp *LineProfileAA) Width(w float64) {
	if w < 0.0 {
		w = 0.0
	}

	if w < lp.smootherWidth {
		w += w
	} else {
		w += lp.smootherWidth
	}

	w *= 0.5

	w -= lp.smootherWidth
	s := lp.smootherWidth
	if w < 0.0 {
		s += w
		w = 0.0
	}
	lp.set(w, s)
}

// profileSlice allocates and returns a pointer to the profile array.
func (lp *LineProfileAA) profileSlice(w float64) []ValueType {
	lp.subpixelWidth = int(basics.URound(w * SubpixelScale))
	size := lp.subpixelWidth + SubpixelScale*6
	if size > lp.profile.Size() {
		lp.profile.Resize(size)
	}
	return lp.profile.Data()
}

// set generates the line profile with center and smoother widths.
func (lp *LineProfileAA) set(centerWidth, smootherWidth float64) {
	baseVal := 1.0
	if centerWidth == 0.0 {
		centerWidth = 1.0 / SubpixelScale
	}
	if smootherWidth == 0.0 {
		smootherWidth = 1.0 / SubpixelScale
	}

	width := centerWidth + smootherWidth
	if width < lp.minWidth {
		k := width / lp.minWidth
		baseVal *= k
		centerWidth /= k
		smootherWidth /= k
	}

	ch := lp.profileSlice(centerWidth + smootherWidth)

	subpixelCenterWidth := int(centerWidth * SubpixelScale)
	subpixelSmootherWidth := int(smootherWidth * SubpixelScale)

	chCenter := SubpixelScale * 2
	chSmoother := chCenter + subpixelCenterWidth

	// Fill center region with full intensity
	val := lp.gamma[int(baseVal*AAMask)]
	for i := 0; i < subpixelCenterWidth; i++ {
		ch[chCenter+i] = val
	}

	// Fill smoother region with gradient
	for i := 0; i < subpixelSmootherWidth; i++ {
		intensity := baseVal - baseVal*(float64(i)/float64(subpixelSmootherWidth))
		ch[chSmoother+i] = lp.gamma[int(intensity*AAMask)]
	}

	// Fill remaining region with zero
	nSmoother := lp.ProfileSize() - subpixelSmootherWidth - subpixelCenterWidth - SubpixelScale*2
	val = lp.gamma[0]
	for i := 0; i < nSmoother; i++ {
		ch[chSmoother+subpixelSmootherWidth+i] = val
	}

	// Mirror the profile to negative side
	for i := 0; i < SubpixelScale*2; i++ {
		ch[chCenter-1-i] = ch[chCenter+i]
	}
}

// ProfileSize returns the size of the profile array.
func (lp *LineProfileAA) ProfileSize() int {
	return lp.profile.Size()
}

// SubpixelWidth returns the subpixel width.
func (lp *LineProfileAA) SubpixelWidth() int {
	return lp.subpixelWidth
}

// GetMinWidth returns the minimum width.
func (lp *LineProfileAA) GetMinWidth() float64 {
	return lp.minWidth
}

// GetSmootherWidth returns the smoother width.
func (lp *LineProfileAA) GetSmootherWidth() float64 {
	return lp.smootherWidth
}

// Value returns the profile value at the given distance.
func (lp *LineProfileAA) Value(dist int) ValueType {
	idx := dist + SubpixelScale*2
	if idx < 0 || idx >= lp.profile.Size() {
		return 0
	}
	return lp.profile.Data()[idx]
}
