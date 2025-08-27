package color

type Space interface{ isColorSpace() }

// Spaces you support today:
type Linear struct{}

func (Linear) isColorSpace() {}

type SRGB struct{}

func (SRGB) isColorSpace() {}
