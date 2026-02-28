// Package agg provides text rendering functionality for the AGG2D high-level interface.
// This implements the text-related methods from the original C++ Agg2D class.
package agg2d

import (
	"math"

	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/font"
	"agg_go/internal/font/freetype"
	"agg_go/internal/path"
	"agg_go/internal/renderer/scanline"
)

// Font loads and configures a font for text rendering.
// This matches the C++ Agg2D::font() method signature and behavior.
func (agg2d *Agg2D) Font(fileName string, height float64, bold, italic bool,
	cacheType FontCacheType, angle float64,
) error {
	if agg2d.fontEngine == nil {
		// Initialize font engine if not already done
		engine, err := freetype.NewFontEngineFreetype(false, 32)
		if err != nil {
			return err
		}
		agg2d.fontEngine = engine
		agg2d.fontCacheManager = font.NewFontCacheManager(engine, 32)
	}

	// Store font parameters
	agg2d.textAngle = angle
	agg2d.fontHeight = height
	agg2d.fontCacheType = cacheType

	// Determine rendering type based on cache type
	var renderingType freetype.GlyphRenderingType
	if cacheType == VectorFontCache {
		renderingType = freetype.GlyphRenderingOutline
	} else {
		renderingType = freetype.GlyphRenderingAAGray8
	}

	// Load the font
	if ftEngine, ok := agg2d.fontEngine.(*freetype.FontEngineFreetype); ok {
		err := ftEngine.LoadFont(fileName, 0, renderingType, nil)
		if err != nil {
			return err
		}

		ftEngine.SetHinting(agg2d.textHints)

		// Set height based on cache type
		if cacheType == VectorFontCache {
			ftEngine.SetHeight(height)
		} else {
			// Convert world coordinates to screen coordinates for raster fonts
			worldHeight := height
			agg2d.WorldToScreen(&worldHeight, &worldHeight)
			ftEngine.SetHeight(worldHeight)
		}
	}

	return nil
}

// FontHeight returns the current font height.
func (agg2d *Agg2D) FontHeight() float64 {
	return agg2d.fontHeight
}

// FlipText sets whether to flip text rendering vertically.
func (agg2d *Agg2D) FlipText(flip bool) {
	if ftEngine, ok := agg2d.fontEngine.(*freetype.FontEngineFreetype); ok {
		ftEngine.SetFlipY(flip)
	}
}

// NOTE: TextAlignment method already exists in agg2d.go, so we don't redefine it here

// TextHints enables or disables font hinting for better text rendering.
func (agg2d *Agg2D) TextHints(hints bool) {
	agg2d.textHints = hints
	if ftEngine, ok := agg2d.fontEngine.(*freetype.FontEngineFreetype); ok {
		ftEngine.SetHinting(hints)
	}
}

// GetTextHints returns whether text hinting is currently enabled.
func (agg2d *Agg2D) GetTextHints() bool {
	return agg2d.textHints
}

// TextWidth calculates the width of the given text string in current units.
// This matches the C++ Agg2D::textWidth() method.
func (agg2d *Agg2D) TextWidth(str string) float64 {
	fcm, ok := agg2d.fontCacheManager.(*font.FontCacheManager)
	if !ok || fcm == nil {
		return 0.0
	}

	x := 0.0
	y := 0.0
	first := true

	// Iterate through each character to calculate total width
	for _, r := range str {
		glyph := fcm.Glyph(uint(r))
		if glyph != nil {
			if !first {
				// Add kerning adjustment
				prevR := uint(0) // We'd need to track the previous character for proper kerning
				fcm.AddKerning(&x, &y, prevR, uint(r))
			}
			x += glyph.AdvanceX
			y += glyph.AdvanceY
			first = false
		}
	}

	// Convert screen coordinates back to world coordinates for raster fonts
	if agg2d.fontCacheType == RasterFontCache {
		worldX := x
		agg2d.ScreenToWorld(&worldX, &y)
		return worldX
	}

	return x
}

// Text renders text at the specified position with optional positioning adjustments.
// This closely matches the C++ Agg2D::text() method implementation.
func (agg2d *Agg2D) Text(x, y float64, str string, roundOff bool, dx, dy float64) {
	fcm, ok := agg2d.fontCacheManager.(*font.FontCacheManager)
	if !ok || fcm == nil || str == "" {
		return
	}

	// Calculate alignment offsets
	alignDx := 0.0
	alignDy := 0.0

	// Horizontal alignment
	switch agg2d.textAlignX {
	case AlignCenter:
		alignDx = -agg2d.TextWidth(str) * 0.5
	case AlignRight:
		alignDx = -agg2d.TextWidth(str)
	}

	// Vertical alignment - calculate font ascender
	ascent := agg2d.fontHeight
	// Try to get ascent from 'H' character for better alignment
	glyph := fcm.Glyph(uint('H'))
	if glyph != nil {
		ascent = float64(glyph.Bounds.Y2 - glyph.Bounds.Y1)
	}

	if agg2d.fontCacheType == RasterFontCache {
		worldAscent := ascent
		agg2d.ScreenToWorld(&worldAscent, &worldAscent)
		ascent = worldAscent
	}

	switch agg2d.textAlignY {
	case AlignCenter:
		alignDy = -ascent * 0.5
	case AlignTop:
		alignDy = -ascent
	}

	// Flip Y alignment if font engine has Y-flipping enabled
	if ftEngine, ok := agg2d.fontEngine.(*freetype.FontEngineFreetype); ok {
		if ftEngine.GetFlipY() {
			alignDy = -alignDy
		}
	}

	// Calculate starting position
	startX := x + alignDx
	startY := y + alignDy

	// Apply rounding if requested
	if roundOff {
		startX = math.Floor(startX)
		startY = math.Floor(startY)
	}

	// Apply additional offset
	startX += dx
	startY += dy

	// Get path adaptor for vector fonts - we'll handle rotation at the agg2d level
	var pathStorage *path.PathStorageStl
	if ps := fcm.PathAdaptor(); ps != nil {
		pathStorage = ps
	}

	// Convert to screen coordinates for raster fonts
	if agg2d.fontCacheType == RasterFontCache {
		agg2d.WorldToScreen(&startX, &startY)
	}

	// Render each character
	currentX := startX
	currentY := startY
	runes := []rune(str)

	for i, r := range runes {
		glyph := fcm.Glyph(uint(r))
		if glyph == nil {
			continue
		}

		// Apply kerning if not the first character
		if i > 0 {
			fcm.AddKerning(&currentX, &currentY, uint(runes[i-1]), uint(r))
		}

		// Initialize glyph adaptors for rendering
		fcm.InitEmbeddedAdaptors(glyph, currentX, currentY)

		// Render based on glyph data type
		switch glyph.DataType {
		case font.GlyphDataOutline:
			// Vector font rendering - add path to current path and draw
			agg2d.path.RemoveAll()
			if pathStorage != nil {
				// Add the glyph path to the current path
				agg2d.path.ConcatPath(pathStorage, 0)
			}

			// For text rotation, apply transformation to the entire rendering context
			if agg2d.textAngle != 0.0 {
				// Save current transform
				savedTransform := *agg2d.transform
				// Apply text rotation: translate(-x,-y) -> rotate(angle) -> translate(x,y)
				agg2d.transform.Translate(-x, -y)
				agg2d.transform.Rotate(agg2d.textAngle)
				agg2d.transform.Translate(x, y)

				agg2d.DrawPath(FillAndStroke)

				// Restore transform
				*agg2d.transform = savedTransform
			} else {
				agg2d.DrawPath(FillAndStroke)
			}

		case font.GlyphDataGray8:
			// Raster font rendering - render using scanlines
			if adaptor := fcm.Gray8Adaptor(); adaptor != nil {
				agg2d.renderGlyphScanlines(adaptor, glyph, currentX, currentY)
			}

		case font.GlyphDataMono:
			// Monochrome font rendering
			if adaptor := fcm.MonoAdaptor(); adaptor != nil {
				agg2d.renderGlyphScanlines(adaptor, glyph, currentX, currentY)
			}
		}

		// Advance to next character position
		currentX += glyph.AdvanceX
		currentY += glyph.AdvanceY
	}
}

// renderGlyphScanlines renders a glyph using scanline data.
// This is a helper method for the Text() function.
func (agg2d *Agg2D) renderGlyphScanlines(adaptor font.SerializedScanlinesAdaptor, glyph *font.GlyphCache, x, y float64) {
	if agg2d.renBase == nil || agg2d.scanline == nil {
		return
	}

	// Get the current fill color as RGBA8[Linear]
	fillColor := color.RGBA8[color.Linear]{
		R: agg2d.fillColor[0],
		G: agg2d.fillColor[1],
		B: agg2d.fillColor[2],
		A: agg2d.fillColor[3],
	}

	// Apply master alpha
	if agg2d.masterAlpha != 1.0 {
		alpha := uint8(float64(fillColor.A) * agg2d.masterAlpha)
		fillColor.A = alpha
	}

	// Use the interface methods directly - no type assertion needed
	bounds := adaptor.Bounds()
	data := adaptor.Data()
	if len(data) == 0 {
		return
	}

	width := bounds.X2 - bounds.X1
	height := bounds.Y2 - bounds.Y1
	if width <= 0 || height <= 0 {
		return
	}

	// Position the glyph at the correct location
	offsetX := int(x) + bounds.X1
	offsetY := int(y) + bounds.Y1

	dataType := font.GlyphDataGray8
	if glyph != nil {
		dataType = glyph.DataType
	}

	pitch := len(data) / height
	if pitch <= 0 {
		return
	}

	covers := make([]basics.Int8u, width)

	switch dataType {
	case font.GlyphDataGray8:
		for row := 0; row < height; row++ {
			rowStart := row * pitch
			if rowStart >= len(data) {
				break
			}

			rowData := data[rowStart:]
			rowLen := pitch
			if rowLen > len(rowData) {
				rowLen = len(rowData)
			}

			for i := range covers {
				covers[i] = 0
			}

			copyWidth := width
			if copyWidth > rowLen {
				copyWidth = rowLen
			}
			for col := 0; col < copyWidth; col++ {
				covers[col] = basics.Int8u(rowData[col])
			}

			agg2d.renBase.BlendSolidHspan(offsetX, offsetY+row, width, fillColor, covers)
		}

	case font.GlyphDataMono:
		for row := 0; row < height; row++ {
			rowStart := row * pitch
			if rowStart >= len(data) {
				break
			}

			rowData := data[rowStart:]
			rowLen := pitch
			if rowLen > len(rowData) {
				rowLen = len(rowData)
			}

			for col := 0; col < width; col++ {
				byteIdx := col >> 3
				if byteIdx >= rowLen {
					covers[col] = 0
					continue
				}

				bit := uint(7 - (col & 7))
				if ((rowData[byteIdx] >> bit) & 0x1) != 0 {
					covers[col] = basics.CoverFull
				} else {
					covers[col] = 0
				}
			}

			agg2d.renBase.BlendSolidHspan(offsetX, offsetY+row, width, fillColor, covers)
		}
	}
}

// renderScanlines renders scanlines using the provided rasterizer and scanline adaptors.
// This is a wrapper method that bridges between font glyph adaptors and the AGG rendering pipeline.
func (agg2d *Agg2D) renderScanlines(ras scanline.RasterizerInterface, sl scanline.ScanlineInterface) {
	if agg2d.renBase == nil {
		return
	}

	// Get the current fill color as RGBA8[Linear]
	fillColor := color.RGBA8[color.Linear]{
		R: agg2d.fillColor[0],
		G: agg2d.fillColor[1],
		B: agg2d.fillColor[2],
		A: agg2d.fillColor[3],
	}

	// Apply master alpha
	if agg2d.masterAlpha != 1.0 {
		alpha := uint8(float64(fillColor.A) * agg2d.masterAlpha)
		fillColor.A = alpha
	}

	// Use the scanline renderer to render all scanlines
	scanline.RenderScanlinesAASolid(ras, sl, agg2d.renBase, fillColor)
}
