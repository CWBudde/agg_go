//go:build !agogo

package engine

func cppAvailable() bool { return false }

func cppUnavailableReason() string { return cppMissingTagReason() }
