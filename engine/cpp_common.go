package engine

import (
	"fmt"
	"image"
)

const cppBuildTag = "agogo"

func newCPPContext(_, _ int) (Context, error) {
	return nil, cppUnavailableError(cppUnavailableReason())
}

func newCPPContextForImage(Image) (Context, error) {
	return nil, cppUnavailableError(cppUnavailableReason())
}

func newCPPImage(_, _ int) (Image, error) {
	return nil, cppUnavailableError(cppUnavailableReason())
}

func newCPPImageFromGoImage(image.Image) (Image, error) {
	return nil, cppUnavailableError(cppUnavailableReason())
}

func newCPPImageFromBuffer([]byte, int, int, int) (Image, error) {
	return nil, cppUnavailableError(cppUnavailableReason())
}

func loadCPPImageFromFile(string) (Image, error) {
	return nil, cppUnavailableError(cppUnavailableReason())
}

func cppUnavailableError(reason string) error {
	return &UnavailableError{
		Kind:   CPP,
		Reason: reason,
	}
}

func cppMissingTagReason() string {
	return fmt.Sprintf("the C++ engine is disabled in this build; rebuild with the %q build tag", cppBuildTag)
}

func cppMissingCGOReason() string {
	return fmt.Sprintf("the %q build tag is enabled, but cgo is disabled", cppBuildTag)
}

func cppBridgeNotImplementedReason() string {
	return fmt.Sprintf("the %q build tag is enabled, but the in-repo C++ engine bridge is not implemented yet", cppBuildTag)
}
