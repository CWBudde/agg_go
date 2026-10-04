//go:build agogo && !cgo

package engine

import "fmt"

func cppAvailable() bool { return false }

func cppUnavailableReason() string {
	return fmt.Sprintf("the %q build tag is enabled, but cgo is disabled", cppBuildTag)
}
