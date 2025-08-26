// Package agg provides text rendering functionality for 2D graphics.
// This file wires public text APIs to internal/agg2d font and text rendering.
package agg

import (
	"errors"

	ia "agg_go/internal/agg2d"
)

// FontCacheType defines font caching modes (re-exported from internal).
type FontCacheType = ia.FontCacheType

const (
	RasterFontCache FontCacheType = ia.RasterFontCache
	VectorFontCache FontCacheType = ia.VectorFontCache
)

// Font loads a font with full configuration.
func (ctx *Context) Font(fileName string, height float64, bold, italic bool, cacheType FontCacheType, angle float64) error {
	return ctx.agg2d.impl.Font(fileName, height, bold, italic, cacheType, angle)
}

// LoadFont loads a font from a file with default settings.
func (ctx *Context) LoadFont(fontFile string) error {
	return ctx.Font(fontFile, 12.0, false, false, RasterFontCache, 0.0)
}

// FontHeight returns the current font height.
func (ctx *Context) FontHeight() float64 { return ctx.agg2d.impl.FontHeight() }

// FlipText flips text vertically.
func (ctx *Context) FlipText(flip bool) { ctx.agg2d.impl.FlipText(flip) }

// TextHints enables/disables font hinting.
func (ctx *Context) TextHints(hints bool) { ctx.agg2d.impl.TextHints(hints) }

// GetTextHints returns current hinting state.
func (ctx *Context) GetTextHints() bool { return ctx.agg2d.impl.GetTextHints() }

// SetTextAlignment configures horizontal and vertical alignment for text.
func (ctx *Context) SetTextAlignment(alignX, alignY TextAlignment) {
	ctx.agg2d.impl.TextAlignment(int(alignX), int(alignY))
}

// DrawText renders text at the specified position.
func (ctx *Context) DrawText(text string, x, y float64) error {
	if len(text) == 0 {
		return errors.New("text is empty")
	}
	ctx.agg2d.impl.Text(x, y, text, true, 0, 0)
	return nil
}

// DrawTextAligned renders text aligned relative to (x,y).
func (ctx *Context) DrawTextAligned(text string, x, y float64, alignment TextAlignment) error {
	if len(text) == 0 {
		return errors.New("text is empty")
	}

	width, _ := ctx.MeasureText(text)
	ax := x
	switch alignment {
	case AlignCenter:
		ax = x - width/2
	case AlignRight:
		ax = x - width
	}
	ctx.agg2d.impl.Text(ax, y, text, true, 0, 0)
	return nil
}

// FillText renders filled text (same as DrawText for AGG path-based rendering).
func (ctx *Context) FillText(text string, x, y float64) error { return ctx.DrawText(text, x, y) }

// StrokeText renders outlined text (uses current stroke settings).
func (ctx *Context) StrokeText(text string, x, y float64) error {
	if len(text) == 0 {
		return errors.New("text is empty")
	}
	ctx.agg2d.impl.Text(x, y, text, true, 0, 0)
	return nil
}

// MeasureText returns width and height of the text with current font settings.
func (ctx *Context) MeasureText(text string) (width, height float64) {
	return ctx.agg2d.impl.TextWidth(text), ctx.agg2d.impl.FontHeight()
}

// GetTextWidth returns the width of the text.
func (ctx *Context) GetTextWidth(text string) float64 { return ctx.agg2d.impl.TextWidth(text) }

// GetTextHeight returns the nominal text height.
func (ctx *Context) GetTextHeight() float64 { return ctx.agg2d.impl.FontHeight() }

// GetTextBounds returns a simple bounds box for the text.
func (ctx *Context) GetTextBounds(text string) (x, y, width, height float64) {
	w := ctx.agg2d.impl.TextWidth(text)
	h := ctx.agg2d.impl.FontHeight()
	return 0, 0, w, h
}

// DrawTextOnPath placeholder until path integration is implemented.
func (ctx *Context) DrawTextOnPath(text string, curved bool) error {
	if len(text) == 0 {
		return errors.New("text is empty")
	}
	return errors.New("text on path not yet implemented - requires path integration")
}

// SetTextRotation rotates subsequent text by angle (radians). Use ResetTextRotation to restore.
func (ctx *Context) SetTextRotation(angle float64) { ctx.PushTransform(); ctx.Rotate(angle) }

// ResetTextRotation restores the previous transform.
func (ctx *Context) ResetTextRotation() { ctx.PopTransform() }

// Styling placeholders
func (ctx *Context) SetBold(bold bool) {}

func (ctx *Context) SetItalic(italic bool) {
	if italic {
		ctx.Skew(0.2, 0)
	}
}
func (ctx *Context) SetUnderline(u bool) {}

// Alignment helpers
func (ctx *Context) DrawTextCentered(text string, x, y float64) error {
	return ctx.DrawTextAligned(text, x, y, AlignCenter)
}

func (ctx *Context) DrawTextRight(text string, x, y float64) error {
	return ctx.DrawTextAligned(text, x, y, AlignRight)
}

func (ctx *Context) DrawTextLeft(text string, x, y float64) error {
	return ctx.DrawTextAligned(text, x, y, AlignLeft)
}

// Multi-line and wrapping
func (ctx *Context) DrawTextLines(lines []string, x, y, lineHeight float64) error {
	if len(lines) == 0 {
		return errors.New("no lines provided")
	}
	cy := y
	for _, line := range lines {
		if err := ctx.DrawText(line, x, cy); err != nil {
			return err
		}
		cy += lineHeight
	}
	return nil
}

func (ctx *Context) DrawTextWrapped(text string, x, y, maxWidth, lineHeight float64) error {
	if len(text) == 0 {
		return errors.New("text is empty")
	}
	words := splitWords(text)
	lines := wrapWords(ctx, words, maxWidth)
	return ctx.DrawTextLines(lines, x, y, lineHeight)
}

// Helpers for wrapping
func splitWords(text string) []string {
	words := make([]string, 0)
	cur := ""
	for _, ch := range text {
		if ch == ' ' || ch == '\n' || ch == '\t' {
			if len(cur) > 0 {
				words = append(words, cur)
				cur = ""
			}
		} else {
			cur += string(ch)
		}
	}
	if len(cur) > 0 {
		words = append(words, cur)
	}
	return words
}

func wrapWords(ctx *Context, words []string, maxWidth float64) []string {
	if len(words) == 0 {
		return []string{}
	}
	lines := make([]string, 0)
	line := ""
	for _, w := range words {
		test := line
		if len(test) > 0 {
			test += " "
		}
		test += w
		if ctx.GetTextWidth(test) <= maxWidth {
			line = test
		} else {
			if len(line) > 0 {
				lines = append(lines, line)
			}
			line = w
		}
	}
	if len(line) > 0 {
		lines = append(lines, line)
	}
	return lines
}
