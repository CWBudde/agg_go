package engine

import (
	"fmt"
	"image"
)

const cppBuildTag = "agogo"

func newCPPContext(width, height int) (Context, error) {
	if !cppAvailable() {
		return nil, cppUnavailableError(cppUnavailableReason())
	}
	return newCPPBackendContextAvailable(width, height)
}

func newCPPContextForImage(img Image) (Context, error) {
	if !cppAvailable() {
		return nil, cppUnavailableError(cppUnavailableReason())
	}
	return newCPPBackendContextForImageAvailable(img)
}

func newCPPImage(width, height int) (Image, error) {
	if !cppAvailable() {
		return nil, cppUnavailableError(cppUnavailableReason())
	}
	return newCPPBackendImageAvailable(width, height)
}

func newCPPImageFromGoImage(src image.Image) (Image, error) {
	if !cppAvailable() {
		return nil, cppUnavailableError(cppUnavailableReason())
	}
	return newCPPBackendImageFromGoImageAvailable(src)
}

func newCPPImageFromBuffer(buf []byte, width, height, stride int) (Image, error) {
	if !cppAvailable() {
		return nil, cppUnavailableError(cppUnavailableReason())
	}
	return newCPPBackendImageFromBufferAvailable(buf, width, height, stride)
}

func loadCPPImageFromFile(filename string) (Image, error) {
	if !cppAvailable() {
		return nil, cppUnavailableError(cppUnavailableReason())
	}
	return loadCPPBackendImageFromFileAvailable(filename)
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
