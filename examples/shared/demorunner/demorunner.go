// Package demorunner provides the shared framework for all AGG demo applications.
//
// Each demo implements the Demo interface and calls Run. The behaviour depends
// on build tags:
//   - No tags (default): render once to a PNG file and exit.
//   - -tags x11: open an X11 window; S saves PNG, ESC quits.
//   - -tags sdl2: open an SDL2 window (preferred over X11 when both present).
//
// Optional interfaces (MouseHandler, KeyHandler) are detected at runtime via
// type assertions, so static demos only need to implement Render.
package demorunner

import agg "github.com/MeKo-Christian/agg_go"

// Config holds the configuration for a demo window / PNG output.
type Config struct {
	Title  string
	Width  int
	Height int
}

// Demo is the core interface every AGG demo must implement.
type Demo interface {
	// Render draws one complete frame into ctx.
	// ctx is pre-allocated to Config.Width × Config.Height and reused across
	// frames; the demo is responsible for clearing it each call.
	Render(ctx *agg.Context)
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
// screenshot) are handled by the demorunner itself and never forwarded.
// Return true if the frame must be redrawn after the event.
type KeyHandler interface {
	OnKey(key rune) bool
}

// Animated marks a demo that requires continuous redraws (e.g. animation).
// Static demos are only redrawn on expose / input events.
type Animated interface {
	IsAnimated() bool
}
