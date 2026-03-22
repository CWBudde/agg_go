// Package lowlevelrunner provides a demo runner for examples that want direct
// access to the raw image buffer instead of the higher-level agg.Context API.
//
// This keeps the existing demorunner package intact for demos that still want
// the convenience layer, while opening a separate path for lower-level ports.
package lowlevelrunner

import agg "github.com/MeKo-Christian/agg_go"

// Config holds the configuration for a demo window / PNG output.
type Config struct {
	Title  string
	Width  int
	Height int
	// FlipY mirrors C++ platform_support's flip_y flag.
	// When true the runner flips mouse Y coordinates before forwarding
	// them to the demo (so the demo sees Y=0 at the bottom).
	FlipY bool
}

// Demo is the core interface every low-level demo must implement.
type Demo interface {
	// Render draws one complete frame into img.
	Render(img *agg.Image)
}

// InitHandler is an optional extension for demos that need one-time setup.
type InitHandler interface {
	OnInit()
}

// IdleHandler is an optional extension for demos that want to advance state
// between redraws.
type IdleHandler interface {
	OnIdle()
}

// Buttons reports which mouse buttons are currently held.
type Buttons struct {
	Left, Right, Middle bool
}

// MouseHandler is an optional extension for demos that respond to mouse input.
// Return true if the frame must be redrawn after the event.
type MouseHandler interface {
	OnMouseMove(x, y int, btn Buttons) bool
	OnMouseDown(x, y int, btn Buttons) bool
	OnMouseUp(x, y int, btn Buttons) bool
}

// KeyHandler is an optional extension for demos that respond to key presses.
// key is the printable rune (e.g. 'r', 'R'). Special keys (ESC, S for
// screenshot) are handled by the runner itself and never forwarded.
// Return true if the frame must be redrawn after the event.
type KeyHandler interface {
	OnKey(key rune) bool
}

// Animated marks a demo that requires continuous redraws (e.g. animation).
// Static demos are only redrawn on expose / input events.
type Animated interface {
	IsAnimated() bool
}
