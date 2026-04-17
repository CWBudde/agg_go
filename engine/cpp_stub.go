package engine

import (
	"image"
)

func cppAvailable() bool { return false }

func newCPPContext(_, _ int) (Context, error) {
	return nil, &UnavailableError{
		Kind:   CPP,
		Reason: "the C++ engine is not implemented in this repository yet",
	}
}

func newCPPImageFromGoImage(image.Image) (Image, error) {
	return nil, &UnavailableError{
		Kind:   CPP,
		Reason: "the C++ engine is not implemented in this repository yet",
	}
}

func loadCPPImageFromFile(string) (Image, error) {
	return nil, &UnavailableError{
		Kind:   CPP,
		Reason: "the C++ engine is not implemented in this repository yet",
	}
}
