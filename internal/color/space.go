package color

// ColorSpace is a zero-cost compile-time marker interface.
// It prevents using arbitrary types as the CS type parameter.
type ColorSpace interface {
	isColorSpace()
}

// Spaces you support today:
type Linear struct{}

func (Linear) isColorSpace() {}

type SRGB struct{}

func (SRGB) isColorSpace() {}
