package color

// Space is the compile-time colorspace marker used by generic color types.
type Space interface {
	isColorSpace()
}

// Linear marks colors stored in linear-light space.
type Linear struct{}

func (Linear) isColorSpace() {}

// SRGB marks colors stored in the standard sRGB transfer space.
type SRGB struct{}

func (SRGB) isColorSpace() {}
