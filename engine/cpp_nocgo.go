//go:build agogo && !cgo

package engine

func cppAvailable() bool { return false }

func cppUnavailableReason() string { return cppMissingCGOReason() }
