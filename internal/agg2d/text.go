// Package agg provides text rendering functionality for the AGG2D high-level interface.
// This implements the text-related methods from the original C++ Agg2D class.
package agg2d

import (
	"math"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	"github.com/cwbudde/agg_go/internal/font"
	"github.com/cwbudde/agg_go/internal/font/freetype"
	"github.com/cwbudde/agg_go/internal/gsv"
	"github.com/cwbudde/agg_go/internal/path"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/transform"
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

	effectiveCacheType := cacheType
	// Rotated text must use outline glyphs. Raster glyph caches are screen-space
	// bitmaps and follow upstream AGG's limitation: they do not rotate cleanly.
	if angle != 0.0 && cacheType == RasterFontCache {
		effectiveCacheType = VectorFontCache
	}

	// Store font parameters
	agg2d.textAngle = angle
	agg2d.fontHeight = height
	agg2d.fontCacheType = effectiveCacheType

	// Determine rendering type based on cache type
	var renderingType freetype.GlyphRenderingType
	if effectiveCacheType == VectorFontCache {
		renderingType = freetype.GlyphRenderingOutline
	} else {
		renderingType = freetype.GlyphRenderingAAGray8
	}

	// Load the font
	if agg2d.fontEngine != nil {
		agg2d.fontEngine.SetResolution(agg2d.resolution)
		agg2d.fontEngine.SetFlipY(agg2d.flipText)
		err := agg2d.fontEngine.LoadFont(fileName, 0, renderingType, nil)
		if err != nil {
			return err
		}

		agg2d.fontEngine.SetHinting(agg2d.textHints)
		agg2d.fontEngine.SetForceAutohint(agg2d.textForceAutohint)

		// Set height based on cache type
		if effectiveCacheType == VectorFontCache {
			agg2d.fontEngine.SetHeight(height)
		} else {
			// Raster glyph caches are configured in screen units.
			agg2d.fontEngine.SetHeight(agg2d.WorldToScreenScalar(height))
		}
	}

	return nil
}

// FontGSV configures the built-in AGG GSV stroke-vector font as the active text
// backend.  This is a WASM-safe alternative to Font() because it uses no cgo
// and requires no font file.
//
// TODO(Path B): This is a temporary solution.  Replace with a pure-Go TTF
// engine (Path A) once one is integrated, and remove this method.
func (agg2d *Agg2D) FontGSV(height float64) {
	if agg2d.gsvText == nil {
		agg2d.gsvText = gsv.NewGSVText()
	}
	// GSV glyph data uses Y-up math coordinates. In a Y-down buffer (positive
	// stride) the height scalar must be negated so characters appear right-side
	// up. In a Y-up buffer (negative stride) the rendering buffer already
	// handles the flip, so no negation is needed.
	agg2d.gsvText.SetFlip(agg2d.rbuf.Stride() >= 0)
	agg2d.gsvText.SetSize(height, 0) // width=0 → proportional
	agg2d.fontHeight = height
	agg2d.gsvFontMode = true
}

// SetResolution sets the font rendering resolution in DPI for FreeType-backed text.
func (agg2d *Agg2D) SetResolution(dpi uint) {
	if dpi > 0 {
		agg2d.resolution = dpi
	}
	if agg2d.fontEngine != nil {
		agg2d.fontEngine.SetResolution(dpi)
	}
}

// FontHeight returns the current font height.
func (agg2d *Agg2D) FontHeight() float64 {
	return agg2d.fontHeight
}

// GetAscender returns the configured font ascender in world units.
func (agg2d *Agg2D) GetAscender() float64 {
	if agg2d.fontEngine != nil {
		return agg2d.fontEngine.GetAscender()
	}
	return 0
}

// GetDescender returns the configured font descender in world units.
func (agg2d *Agg2D) GetDescender() float64 {
	if agg2d.fontEngine != nil {
		return agg2d.fontEngine.GetDescender()
	}
	return 0
}

// MeasureText returns width and height for the current font settings.
func (agg2d *Agg2D) MeasureText(text string) (width, height float64) {
	if agg2d.fontCacheType == RasterFontCache {
		if width, height, _, _, ok := agg2d.textRunMetrics(text); ok {
			return width, height
		}
	}
	width = agg2d.TextWidth(text)
	ascent := agg2d.GetAscender()
	descent := -agg2d.GetDescender()
	if ascent <= 0 && descent <= 0 {
		return width, agg2d.FontHeight()
	}
	return width, ascent + descent
}

// GetTextHeight returns the nominal height of the current font.
func (agg2d *Agg2D) GetTextHeight() float64 {
	_, height := agg2d.MeasureText("X")
	if height > 0 {
		return height
	}
	return agg2d.FontHeight()
}

// FlipText sets whether to flip text rendering vertically.
func (agg2d *Agg2D) FlipText(flip bool) {
	agg2d.flipText = flip
	if agg2d.fontEngine != nil {
		agg2d.fontEngine.SetFlipY(flip)
	}
}

// NOTE: TextAlignment method already exists in agg2d.go, so we don't redefine it here

// TextHints enables or disables font hinting for better text rendering.
func (agg2d *Agg2D) TextHints(hints bool) {
	agg2d.textHints = hints
	if agg2d.fontEngine != nil {
		agg2d.fontEngine.SetHinting(hints)
	}
}

// TextForceAutohint enables or disables FreeType's auto-hinter for raster text.
func (agg2d *Agg2D) TextForceAutohint(force bool) {
	agg2d.textForceAutohint = force
	if agg2d.fontEngine != nil {
		agg2d.fontEngine.SetForceAutohint(force)
	}
}

// GetTextHints returns whether text hinting is currently enabled.
func (agg2d *Agg2D) GetTextHints() bool {
	return agg2d.textHints
}

func (agg2d *Agg2D) shapedRasterGlyphs(str string) ([]font.PositionedGlyph, bool) {
	if agg2d.fontCacheType != RasterFontCache || agg2d.fontEngine == nil || str == "" {
		return nil, false
	}
	return agg2d.fontEngine.LayoutText(str)
}

func quantizeRasterTextPhaseF26Dot6(v float64) (base int, frac float64) {
	quantized := math.Round(v*64.0) / 64.0
	baseFloat := math.Floor(quantized)
	return int(baseFloat), quantized - baseFloat
}

func quantizeRasterTextYPhaseF26Dot6(v float64) float64 {
	_, frac := quantizeRasterTextPhaseF26Dot6(-v)
	return frac
}

func (agg2d *Agg2D) rasterPlacedGlyphBounds(glyphs []font.PositionedGlyph) (minX, minY, maxX, maxY float64, ok bool) {
	if agg2d.fontEngine == nil {
		return 0, 0, 0, 0, false
	}

	currentX := 0.0
	currentY := 0.0
	minXi, minYi := math.MaxInt32, math.MaxInt32
	maxXi, maxYi := math.MinInt32, math.MinInt32

	for _, placedGlyph := range glyphs {
		glyphX := currentX + placedGlyph.XOffset
		glyphY := currentY + placedGlyph.YOffset
		baseX, fracX := quantizeRasterTextPhaseF26Dot6(glyphX)
		baseY, _ := quantizeRasterTextPhaseF26Dot6(glyphY)
		fracY := quantizeRasterTextYPhaseF26Dot6(glyphY)

		if !agg2d.fontEngine.PrepareGlyphIndexSubpixel(placedGlyph.GlyphIndex, fracX, fracY) {
			currentX += placedGlyph.XAdvance
			currentY += placedGlyph.YAdvance
			continue
		}

		_, width, height, pitch, left, top, _ := agg2d.fontEngine.CurrentBitmap()
		if width > 0 && height > 0 && pitch != 0 {
			dstX := baseX + left
			dstY := baseY - top + 1
			if dstX < minXi {
				minXi = dstX
			}
			if dstY < minYi {
				minYi = dstY
			}
			if dstX+width > maxXi {
				maxXi = dstX + width
			}
			if dstY+height > maxYi {
				maxYi = dstY + height
			}
			ok = true
		}

		currentX += placedGlyph.XAdvance
		currentY += placedGlyph.YAdvance
	}

	if !ok {
		return 0, 0, 0, 0, false
	}
	return float64(minXi), float64(minYi), float64(maxXi), float64(maxYi), true
}

func positionedGlyphBounds(glyphs []font.PositionedGlyph, fcm *font.FontCacheManager) (minX, minY, maxX, maxY float64, ok bool) {
	penX := 0.0
	penY := 0.0

	for _, placed := range glyphs {
		glyph := fcm.GlyphByIndex(placed.GlyphIndex)
		if glyph == nil {
			penX += placed.XAdvance
			penY += placed.YAdvance
			continue
		}

		x1 := penX + placed.XOffset + float64(glyph.Bounds.X1)
		x2 := penX + placed.XOffset + float64(glyph.Bounds.X2)
		y1 := penY + placed.YOffset + float64(glyph.Bounds.Y1)
		y2 := penY + placed.YOffset + float64(glyph.Bounds.Y2)
		if !ok {
			minX = math.Min(x1, x2)
			maxX = math.Max(x1, x2)
			minY = math.Min(y1, y2)
			maxY = math.Max(y1, y2)
			ok = true
		} else {
			minX = math.Min(minX, math.Min(x1, x2))
			maxX = math.Max(maxX, math.Max(x1, x2))
			minY = math.Min(minY, math.Min(y1, y2))
			maxY = math.Max(maxY, math.Max(y1, y2))
		}

		penX += placed.XAdvance
		penY += placed.YAdvance
	}
	return minX, minY, maxX, maxY, ok
}

func (agg2d *Agg2D) textRunBounds(str string) (minX, minY, maxX, maxY float64, ok bool) {
	fcm := agg2d.fontCacheManager
	if fcm == nil || str == "" {
		return 0, 0, 0, 0, false
	}

	if glyphs, ok := agg2d.shapedRasterGlyphs(str); ok {
		if minX, minY, maxX, maxY, ok := agg2d.rasterPlacedGlyphBounds(glyphs); ok {
			return minX, minY, maxX, maxY, true
		}
		return positionedGlyphBounds(glyphs, fcm)
	}

	x := 0.0
	y := 0.0
	first := true
	var prevGlyphIndex uint

	for _, r := range str {
		glyph := fcm.Glyph(uint(r))
		if glyph == nil {
			continue
		}
		if !first {
			fcm.AddKerning(&x, &y, prevGlyphIndex, glyph.GlyphIndex)
		}

		x1 := x + float64(glyph.Bounds.X1)
		x2 := x + float64(glyph.Bounds.X2)
		y1 := y + float64(glyph.Bounds.Y1)
		y2 := y + float64(glyph.Bounds.Y2)
		if !ok {
			minX = math.Min(x1, x2)
			maxX = math.Max(x1, x2)
			minY = math.Min(y1, y2)
			maxY = math.Max(y1, y2)
			ok = true
		} else {
			minX = math.Min(minX, math.Min(x1, x2))
			maxX = math.Max(maxX, math.Max(x1, x2))
			minY = math.Min(minY, math.Min(y1, y2))
			maxY = math.Max(maxY, math.Max(y1, y2))
		}

		x += glyph.AdvanceX
		y += glyph.AdvanceY
		prevGlyphIndex = glyph.GlyphIndex
		first = false
	}

	return minX, minY, maxX, maxY, ok
}

func (agg2d *Agg2D) textRunMetrics(str string) (width, height, ascent, descent float64, ok bool) {
	minX, minY, maxX, maxY, ok := agg2d.textRunBounds(str)
	if !ok {
		return 0, 0, 0, 0, false
	}

	width = maxX - minX
	height = maxY - minY
	ascent = math.Max(0, -minY)
	descent = math.Max(0, maxY)

	if agg2d.fontCacheType == RasterFontCache {
		width = agg2d.ScreenToWorldScalar(width)
		height = agg2d.ScreenToWorldScalar(height)
		ascent = agg2d.ScreenToWorldScalar(ascent)
		descent = agg2d.ScreenToWorldScalar(descent)
	}

	return width, height, ascent, descent, true
}

func (agg2d *Agg2D) screenToWorldOffset(offset float64) float64 {
	if offset == 0 {
		return 0
	}
	sign := 1.0
	if offset < 0 {
		sign = -1.0
		offset = -offset
	}
	return sign * agg2d.ScreenToWorldScalar(offset)
}

// GetTextBounds reports the actual ink bounds of str relative to the baseline
// origin. The returned x/y are offsets from the baseline point to the top-left
// corner of the bounds.
func (agg2d *Agg2D) GetTextBounds(str string) (x, y, width, height float64) {
	minX, minY, maxX, maxY, ok := agg2d.textRunBounds(str)
	if !ok {
		return 0, 0, 0, 0
	}
	if agg2d.fontCacheType == RasterFontCache {
		minX = agg2d.screenToWorldOffset(minX)
		minY = agg2d.screenToWorldOffset(minY)
		maxX = agg2d.screenToWorldOffset(maxX)
		maxY = agg2d.screenToWorldOffset(maxY)
	}
	return minX, minY, maxX - minX, maxY - minY
}

func preparedGlyphFromEngine(engine *freetype.FontEngineFreetype) *font.GlyphCache {
	if engine == nil {
		return nil
	}
	glyph := &font.GlyphCache{
		GlyphIndex: engine.GlyphIndex(),
		DataSize:   engine.DataSize(),
		DataType:   engine.DataType(),
		Bounds:     engine.Bounds(),
		AdvanceX:   engine.AdvanceX(),
		AdvanceY:   engine.AdvanceY(),
	}
	if glyph.DataSize > 0 {
		glyph.Data = make([]byte, glyph.DataSize)
		engine.WriteGlyphTo(glyph.Data)
	}
	return glyph
}

func (agg2d *Agg2D) renderShapedRasterMask(startX, startY float64, glyphs []font.PositionedGlyph) bool {
	if agg2d.fontEngine == nil {
		return false
	}

	renderer := agg2d.currentRenderer()
	if renderer == nil {
		return false
	}
	fillColor := color.RGBA8[color.Linear]{
		R: agg2d.fillColor[0],
		G: agg2d.fillColor[1],
		B: agg2d.fillColor[2],
		A: agg2d.fillColor[3],
	}
	currentX := startX
	currentY := startY
	rendered := false

	for _, placedGlyph := range glyphs {
		glyphX := currentX + placedGlyph.XOffset
		glyphY := currentY + placedGlyph.YOffset
		baseX, fracX := quantizeRasterTextPhaseF26Dot6(glyphX)
		baseY, _ := quantizeRasterTextPhaseF26Dot6(glyphY)
		fracY := quantizeRasterTextYPhaseF26Dot6(glyphY)

		if !agg2d.fontEngine.PrepareGlyphIndexSubpixel(placedGlyph.GlyphIndex, fracX, fracY) {
			currentX += placedGlyph.XAdvance
			currentY += placedGlyph.YAdvance
			continue
		}

		data, width, height, pitch, left, top, pixelMode := agg2d.fontEngine.CurrentBitmap()
		if len(data) == 0 || width <= 0 || height <= 0 || pitch == 0 {
			currentX += placedGlyph.XAdvance
			currentY += placedGlyph.YAdvance
			continue
		}

		dstX := baseX + left
		dstY := baseY - top + 1
		if blendRasterGlyphBitmap(renderer, fillColor, dstX, dstY, width, height, pitch, pixelMode, data) {
			rendered = true
		}

		currentX += placedGlyph.XAdvance
		currentY += placedGlyph.YAdvance
	}

	return rendered
}

func blendRasterGlyphBitmap(
	renderer *baseRendererAdapter[color.RGBA8[color.Linear]],
	fillColor color.RGBA8[color.Linear],
	dstX, dstY, width, height, pitch int,
	pixelMode uint8,
	data []byte,
) bool {
	rowStride := pitch
	if rowStride < 0 {
		rowStride = -rowStride
	}
	if rowStride <= 0 {
		return false
	}

	rendered := false
	covers := make([]basics.Int8u, width)
	for row := 0; row < height; row++ {
		srcRow := row
		if pitch < 0 {
			srcRow = height - 1 - row
		}
		srcOffset := srcRow * rowStride
		if srcOffset >= len(data) {
			continue
		}

		switch pixelMode {
		case 2: // FT_PIXEL_MODE_GRAY
			limit := width
			if srcOffset+limit > len(data) {
				limit = len(data) - srcOffset
			}
			if limit <= 0 {
				continue
			}
			for i := 0; i < limit; i++ {
				covers[i] = basics.Int8u(data[srcOffset+i])
			}
			for i := limit; i < width; i++ {
				covers[i] = 0
			}
		case 1: // FT_PIXEL_MODE_MONO
			clear(covers)
			for col := 0; col < width; col++ {
				byteIdx := srcOffset + (col >> 3)
				if byteIdx >= len(data) {
					break
				}
				if (data[byteIdx] & (1 << uint(7-(col&7)))) != 0 {
					covers[col] = 0xff
				}
			}
		default:
			continue
		}

		renderer.BlendSolidHspan(dstX, dstY+row, width, fillColor, covers)
		rendered = true
	}
	return rendered
}

// TextWidth calculates the width of the given text string in current units.
// This matches the C++ Agg2D::textWidth() method.
func (agg2d *Agg2D) TextWidth(str string) float64 {
	if agg2d.gsvFontMode && agg2d.gsvText != nil {
		return agg2d.gsvText.MeasureText(str)
	}

	if agg2d.fontCacheType == RasterFontCache {
		if width, _, _, _, ok := agg2d.textRunMetrics(str); ok {
			return width
		}
	}

	fcm := agg2d.fontCacheManager
	if fcm == nil {
		return 0.0
	}

	x := 0.0
	y := 0.0
	first := true
	var prevGlyphIndex uint

	// Iterate through each character to calculate total width.
	for _, r := range str {
		glyph := fcm.Glyph(uint(r))
		if glyph == nil {
			continue
		}
		if !first {
			// Kerning in FreeType is defined between glyph indices.
			fcm.AddKerning(&x, &y, prevGlyphIndex, glyph.GlyphIndex)
		}
		x += glyph.AdvanceX
		y += glyph.AdvanceY
		first = false
		prevGlyphIndex = glyph.GlyphIndex
	}

	if agg2d.fontCacheType == RasterFontCache {
		return agg2d.ScreenToWorldScalar(x)
	}
	return x
}

// textGSV renders text using the built-in AGG GSV stroke-vector font.
// The stroked glyph outlines are painted with the current fill color so that
// callers can simply set FillColor and call Text() without worrying about
// which underlying backend is active.
//
// TODO(Path B): Temporary GSV fallback — remove when a proper TTF engine
// (Path A) is integrated.
func (agg2d *Agg2D) textGSV(x, y float64, str string, roundOff bool, dx, dy float64) {
	if agg2d.gsvText == nil || str == "" {
		return
	}
	t := agg2d.gsvText
	t.SetText(str)

	// Alignment offsets (approximate using GSV TextWidth)
	alignDx := 0.0
	alignDy := 0.0
	switch agg2d.textAlignX {
	case AlignCenter:
		alignDx = -agg2d.TextWidth(str) * 0.5
	case AlignRight:
		alignDx = -agg2d.TextWidth(str)
	}
	switch agg2d.textAlignY {
	case AlignCenter:
		alignDy = agg2d.fontHeight * 0.5
	case AlignTop:
		alignDy = agg2d.fontHeight
	}

	startX := x + alignDx + dx
	startY := y - alignDy + dy // GSV Y grows down; subtract to shift baseline
	if roundOff {
		startX = float64(int(startX))
		startY = float64(int(startY))
	}

	t.SetStartPoint(startX, startY)

	// Collect all GSV vertices into agg2d.path so the standard transform
	// pipeline (convCurve → ConvTransform) picks them up automatically.
	agg2d.path.RemoveAll()
	t.Rewind(0)
	for {
		vx, vy, cmd := t.Vertex()
		switch cmd {
		case basics.PathCmdMoveTo:
			agg2d.path.MoveTo(vx, vy)
		case basics.PathCmdLineTo:
			agg2d.path.LineTo(vx, vy)
		case basics.PathCmdStop:
			goto vertexLoopDone
		default:
			// GSV does not emit end-poly or curve commands; ignore.
		}
	}
vertexLoopDone:

	// Stroke the skeleton paths to produce legible characters.
	// Use a thin fixed stroke width (1 px at current scale) so that the glyphs
	// look like the original AGG gsv_text examples rather than being as thick
	// as the document line width.
	// TODO(Path B): expose stroke width as a parameter of FontGSV().
	pathAdapter := path.NewPathStorageStlVertexSourceAdapter(agg2d.path)
	curvesAdapter := conv.NewConvCurve(pathAdapter)
	strokeAdapter := conv.NewConvStroke(curvesAdapter)
	strokeAdapter.SetWidth(agg2d.fontHeight * 0.08) // ~8 % of glyph height
	strokeAdapter.SetLineCap(basics.RoundCap)
	strokeAdapter.SetLineJoin(basics.RoundJoin)

	// Apply the global affine transform so text respects Viewport(), Rotate(), etc.
	transformedStroke := conv.NewConvTransform(strokeAdapter, agg2d.transform)

	agg2d.rasterizer.Reset()
	agg2d.rasterizer.FillingRule(basics.FillNonZero)
	transformedStroke.Rewind(0)
	for {
		x, y, cmd := transformedStroke.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		agg2d.rasterizer.AddVertex(x, y, uint32(cmd))
	}

	// Paint using fill color (mirrors FreeType path: fill color = text color).
	agg2d.renderSolidFillWithColor(agg2d.fillColor)
}

// Text renders text at the specified position with optional positioning adjustments.
// This closely matches the C++ Agg2D::text() method implementation.
func (agg2d *Agg2D) Text(x, y float64, str string, roundOff bool, dx, dy float64) {
	// TODO(Path B): Route through GSV when no FreeType font is loaded.
	if agg2d.gsvFontMode {
		agg2d.textGSV(x, y, str, roundOff, dx, dy)
		return
	}

	fcm := agg2d.fontCacheManager
	if fcm == nil || str == "" {
		return
	}

	shapedGlyphs, haveShapedGlyphs := agg2d.shapedRasterGlyphs(str)

	// Calculate alignment offsets
	alignDx := 0.0
	alignDy := 0.0

	if haveShapedGlyphs {
		minX, minY, maxX, maxY, ok := positionedGlyphBounds(shapedGlyphs, fcm)
		if ok {
			switch agg2d.textAlignX {
			case AlignCenter:
				alignDx = -agg2d.ScreenToWorldScalar((minX + maxX) * 0.5)
			case AlignRight:
				alignDx = -agg2d.ScreenToWorldScalar(maxX)
			}
			switch agg2d.textAlignY {
			case AlignCenter:
				alignDy = -agg2d.ScreenToWorldScalar((minY + maxY) * 0.5)
			case AlignTop:
				alignDy = -agg2d.ScreenToWorldScalar(minY)
			}
		}
	} else {
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
			ascent = agg2d.ScreenToWorldScalar(ascent)
		}

		switch agg2d.textAlignY {
		case AlignCenter:
			alignDy = -ascent * 0.5
		case AlignTop:
			alignDy = -ascent
		}

		// Flip Y alignment if font engine has Y-flipping enabled
		if agg2d.fontEngine != nil && agg2d.fontEngine.GetFlipY() {
			alignDy = -alignDy
		}
	}

	// Calculate starting position
	startX := x + alignDx
	startY := y + alignDy

	// Apply rounding if requested (matches C++ int() truncation semantics)
	if roundOff {
		startX = float64(int(startX))
		startY = float64(int(startY))
	}

	// Apply additional offset
	startX += dx
	startY += dy

	pathStorage := fcm.PathAdaptor()
	var textTransform *transform.TransAffine
	if agg2d.textAngle != 0.0 {
		textTransform = transform.NewTransAffine()
		textTransform.Translate(-x, -y)
		textTransform.Rotate(agg2d.textAngle)
		textTransform.Translate(x, y)
	}

	// Convert to screen coordinates for raster fonts
	if agg2d.fontCacheType == RasterFontCache {
		agg2d.WorldToScreen(&startX, &startY)
	}

	// Render each character
	currentX := startX
	currentY := startY

	if haveShapedGlyphs {
		if agg2d.renderShapedRasterMask(currentX, currentY, shapedGlyphs) {
			return
		}
	}

	firstGlyph := true
	var prevGlyphIndex uint
	var glyph *font.GlyphCache

	for _, r := range str {
		glyph = fcm.Glyph(uint(r))
		if glyph == nil {
			continue
		}

		if !firstGlyph {
			fcm.AddKerning(&currentX, &currentY, prevGlyphIndex, glyph.GlyphIndex)
		}

		// Initialize glyph adaptors for rendering.
		fcm.InitEmbeddedAdaptors(glyph, currentX, currentY)

		switch glyph.DataType {
		case font.GlyphDataOutline:
			agg2d.path.RemoveAll()
			if pathStorage != nil {
				if textTransform != nil {
					agg2d.path.ConcatPath(&transformedPathSource{src: pathStorage, mtx: textTransform}, 0)
				} else {
					agg2d.path.ConcatPath(pathStorage, 0)
				}
				agg2d.DrawPath(FillAndStroke)
			}

		case font.GlyphDataGray8:
			if adaptor := fcm.Gray8Adaptor(); adaptor != nil {
				agg2d.renderGlyphScanlines(adaptor, glyph, currentX, currentY)
			}

		// GlyphDataMono: Go extension — C++ agg2d.cpp text() only handles outline and
		// gray8; mono is rendered here for completeness when a font engine is configured
		// for binary (non-AA) rasterization.
		case font.GlyphDataMono:
			if adaptor := fcm.MonoAdaptor(); adaptor != nil {
				agg2d.renderGlyphScanlines(adaptor, glyph, currentX, currentY)
			}
		}

		currentX += glyph.AdvanceX
		currentY += glyph.AdvanceY
		prevGlyphIndex = glyph.GlyphIndex
		firstGlyph = false
	}
}

// transformedPathSource applies an affine transform while iterating a path source.
type transformedPathSource struct {
	src path.VertexSource
	mtx *transform.TransAffine
}

func (t *transformedPathSource) Rewind(pathID uint) {
	if t.src != nil {
		t.src.Rewind(pathID)
	}
}

func (t *transformedPathSource) NextVertex() (x, y float64, cmd uint32) {
	if t.src == nil {
		return 0, 0, uint32(basics.PathCmdStop)
	}

	x, y, cmd = t.src.NextVertex()
	if t.mtx != nil && basics.IsVertex(basics.PathCommand(cmd)) {
		t.mtx.Transform(&x, &y)
	}
	return x, y, cmd
}

// renderGlyphScanlines renders a glyph using scanline data.
// This mirrors AGG2D's render(gray8_adaptor/mono_adaptor, scanline) flow.
func (agg2d *Agg2D) renderGlyphScanlines(adaptor font.SerializedScanlinesAdaptor, glyph *font.GlyphCache, x, y float64) {
	if agg2d.scanline == nil || glyph == nil || adaptor == nil {
		return
	}
	agg2d.renderScanlines(adaptor, agg2d.scanline, glyph.DataType == font.GlyphDataMono)
}

// renderScanlines renders scanlines using the provided rasterizer and scanline adaptors.
func (agg2d *Agg2D) renderScanlines(ras renscan.RasterizerInterface, sl renscan.ScanlineInterface, mono bool) {
	renderer := agg2d.currentRenderer()
	if renderer == nil {
		return
	}

	fillColor := color.RGBA8[color.Linear]{
		R: agg2d.fillColor[0],
		G: agg2d.fillColor[1],
		B: agg2d.fillColor[2],
		A: agg2d.fillColor[3],
	}
	if agg2d.masterAlpha != 1.0 {
		alpha := uint8(float64(fillColor.A) * agg2d.masterAlpha)
		fillColor.A = alpha
	}

	if mono {
		renscan.RenderScanlinesBinSolid(ras, sl, renderer, fillColor)
		return
	}
	renscan.RenderScanlinesAASolid(ras, sl, renderer, fillColor)
}
