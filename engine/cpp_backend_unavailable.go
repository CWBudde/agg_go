//go:build !agogo || !cgo

package engine

import "image"

func newCPPBackendContextAvailable(int, int) (Context, error) {
	return nil, cppUnavailableError(cppUnavailableReason())
}

func newCPPBackendContextForImageAvailable(Image) (Context, error) {
	return nil, cppUnavailableError(cppUnavailableReason())
}

func newCPPBackendImageAvailable(int, int) (Image, error) {
	return nil, cppUnavailableError(cppUnavailableReason())
}

func newCPPBackendImageFromGoImageAvailable(image.Image) (Image, error) {
	return nil, cppUnavailableError(cppUnavailableReason())
}

func newCPPBackendImageFromBufferAvailable([]byte, int, int, int) (Image, error) {
	return nil, cppUnavailableError(cppUnavailableReason())
}

func loadCPPBackendImageFromFileAvailable(string) (Image, error) {
	return nil, cppUnavailableError(cppUnavailableReason())
}
