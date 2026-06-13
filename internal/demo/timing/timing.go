// Package timing centralizes demo timing-overlay flags.
package timing

import (
	"os"
	"strings"
)

// TextEnv controls whether demos render volatile timing text into their frame.
// It defaults to enabled; set it to 0, false, no, or off for parity PNG output.
const TextEnv = "AGG_GO_DEMO_TIMING_TEXT"

// ShowText reports whether demos should draw timing/benchmark text overlays.
func ShowText() bool {
	value, ok := os.LookupEnv(TextEnv)
	if !ok {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}
