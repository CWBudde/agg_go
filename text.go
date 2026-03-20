package agg

import (
	"errors"

	ia "github.com/MeKo-Christian/agg_go/internal/agg2d"
)

// FontCacheType defines font caching modes (re-exported from internal).
type FontCacheType = ia.FontCacheType

const (
	// RasterFontCache selects raster glyph caching.
	RasterFontCache FontCacheType = ia.RasterFontCache
	// VectorFontCache selects vector-outline glyph caching.
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

// SetResolution sets the font rendering resolution in DPI.
func (ctx *Context) SetResolution(dpi uint) { ctx.agg2d.impl.SetResolution(dpi) }

// GetAscender returns the current font ascender in world units.
func (ctx *Context) GetAscender() float64 { return ctx.agg2d.impl.GetAscender() }

// GetDescender returns the current font descender in world units.
func (ctx *Context) GetDescender() float64 { return ctx.agg2d.impl.GetDescender() }

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
	if text == "" {
		return errors.New("text is empty")
	}
	ctx.agg2d.impl.Text(x, y, text, true, 0, 0)
	return nil
}

// DrawTextAligned renders text aligned relative to (x,y).
func (ctx *Context) DrawTextAligned(text string, x, y float64, alignment TextAlignment) error {
	if text == "" {
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
	if text == "" {
		return errors.New("text is empty")
	}
	ctx.agg2d.impl.Text(x, y, text, true, 0, 0)
	return nil
}

// MeasureText returns width and height of the text with current font settings.
func (ctx *Context) MeasureText(text string) (width, height float64) {
	width = ctx.agg2d.impl.TextWidth(text)
	ascent := ctx.GetAscender()
	descent := -ctx.GetDescender()
	if ascent <= 0 && descent <= 0 {
		return width, ctx.agg2d.impl.FontHeight()
	}
	return width, ascent + descent
}

// GetTextWidth returns the width of the text.
func (ctx *Context) GetTextWidth(text string) float64 { return ctx.agg2d.impl.TextWidth(text) }

// GetTextHeight returns the nominal text height.
func (ctx *Context) GetTextHeight() float64 {
	ascent := ctx.GetAscender()
	descent := -ctx.GetDescender()
	if ascent > 0 || descent > 0 {
		return ascent + descent
	}
	return ctx.agg2d.impl.FontHeight()
}

// GetTextBounds returns a simple bounds box for the text.
func (ctx *Context) GetTextBounds(text string) (x, y, width, height float64) {
	w, h := ctx.MeasureText(text)
	return 0, 0, w, h
}

// DrawTextOnPath placeholder until path integration is implemented.
func (ctx *Context) DrawTextOnPath(text string, curved bool) error {
	if text == "" {
		return errors.New("text is empty")
	}
	return errors.New("text on path not yet implemented - requires path integration")
}

// SetTextRotation rotates subsequent text by angle (radians). Use ResetTextRotation to restore.
func (ctx *Context) SetTextRotation(angle float64) { ctx.PushTransform(); ctx.Rotate(angle) }

// ResetTextRotation restores the previous transform.
func (ctx *Context) ResetTextRotation() { ctx.PopTransform() }

// SetBold is currently a placeholder for future font-style handling.
func (ctx *Context) SetBold(bold bool) {}

// SetItalic applies a simple skew-based italic transform when italic is true.
func (ctx *Context) SetItalic(italic bool) {
	if italic {
		ctx.Skew(0.2, 0)
	}
}

// SetUnderline is currently a placeholder for future underline support.
func (ctx *Context) SetUnderline(u bool) {}

// DrawTextCentered draws text centered on x.
func (ctx *Context) DrawTextCentered(text string, x, y float64) error {
	return ctx.DrawTextAligned(text, x, y, AlignCenter)
}

// DrawTextRight draws text right-aligned to x.
func (ctx *Context) DrawTextRight(text string, x, y float64) error {
	return ctx.DrawTextAligned(text, x, y, AlignRight)
}

// DrawTextLeft draws text left-aligned to x.
func (ctx *Context) DrawTextLeft(text string, x, y float64) error {
	return ctx.DrawTextAligned(text, x, y, AlignLeft)
}

// DrawTextLines draws multiple lines with a fixed line advance.
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

// DrawTextWrapped wraps text to maxWidth and renders the resulting lines.
func (ctx *Context) DrawTextWrapped(text string, x, y, maxWidth, lineHeight float64) error {
	if text == "" {
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
			if cur != "" {
				words = append(words, cur)
				cur = ""
			}
		} else {
			cur += string(ch)
		}
	}
	if cur != "" {
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
		if test != "" {
			test += " "
		}
		test += w
		if ctx.GetTextWidth(test) <= maxWidth {
			line = test
		} else {
			if line != "" {
				lines = append(lines, line)
			}
			line = w
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}
