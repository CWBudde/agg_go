// Float (128-bit, 4 x float32) text rendering for Agg2DFloat.
//
// This is the float twin of text.go. It mirrors the 8-bit Agg2D text pipeline
// method-for-method, swapping the solid fill color type to color.RGBA32 and the
// base renderer to the float baseRendererAdapter. The font engine, font cache
// manager, GSV stroke font, glyph layout, metrics, and bounds math are all
// color-agnostic and reused verbatim, as are the package-level helpers
// (quantizeRasterTextPhaseF26Dot6, positionedGlyphBounds, transformedPathSource,
// and the now-generic blendRasterGlyphBitmap).
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

// Font loads and configures a font for text rendering, mirroring 8-bit Agg2D.Font.
func (a *Agg2DFloat) Font(fileName string, height float64, bold, italic bool,
	cacheType FontCacheType, angle float64,
) error {
	if a.fontEngine == nil {
		engine, err := freetype.NewFontEngineFreetype(false, 32)
		if err != nil {
			return err
		}
		a.fontEngine = engine
		a.fontCacheManager = font.NewFontCacheManager(engine, 32)
	}

	effectiveCacheType := cacheType
	// Rotated text must use outline glyphs; raster glyph caches are screen-space
	// bitmaps that do not rotate cleanly (matches upstream AGG).
	if angle != 0.0 && cacheType == RasterFontCache {
		effectiveCacheType = VectorFontCache
	}

	a.textAngle = angle
	a.fontHeight = height
	a.fontCacheType = effectiveCacheType

	var renderingType freetype.GlyphRenderingType
	if effectiveCacheType == VectorFontCache {
		renderingType = freetype.GlyphRenderingOutline
	} else {
		renderingType = freetype.GlyphRenderingAAGray8
	}

	if a.fontEngine != nil {
		a.fontEngine.SetResolution(a.resolution)
		a.fontEngine.SetFlipY(a.flipText)
		if err := a.fontEngine.LoadFont(fileName, 0, renderingType, nil); err != nil {
			return err
		}

		a.fontEngine.SetHinting(a.textHints)
		a.fontEngine.SetForceAutohint(a.textForceAutohint)

		if effectiveCacheType == VectorFontCache {
			a.fontEngine.SetHeight(height)
		} else {
			a.fontEngine.SetHeight(a.WorldToScreenScalar(height))
		}
	}

	return nil
}

// FontGSV configures the built-in AGG GSV stroke-vector font as the active text
// backend. WASM-safe (no cgo, no font file), mirroring 8-bit Agg2D.FontGSV.
func (a *Agg2DFloat) FontGSV(height float64) {
	if a.gsvText == nil {
		a.gsvText = gsv.NewGSVText()
	}
	a.gsvText.SetFlip(a.rbuf.Stride() >= 0)
	a.gsvText.SetSize(height, 0) // width=0 → proportional
	a.fontHeight = height
	a.gsvFontMode = true
}

// SetResolution sets the font rendering resolution in DPI for FreeType text.
func (a *Agg2DFloat) SetResolution(dpi uint) {
	if dpi > 0 {
		a.resolution = dpi
	}
	if a.fontEngine != nil {
		a.fontEngine.SetResolution(dpi)
	}
}

// FontHeight returns the current font height.
func (a *Agg2DFloat) FontHeight() float64 { return a.fontHeight }

// GetAscender returns the configured font ascender in world units.
func (a *Agg2DFloat) GetAscender() float64 {
	if a.fontEngine != nil {
		return a.fontEngine.GetAscender()
	}
	return 0
}

// GetDescender returns the configured font descender in world units.
func (a *Agg2DFloat) GetDescender() float64 {
	if a.fontEngine != nil {
		return a.fontEngine.GetDescender()
	}
	return 0
}

// MeasureText returns width and height for the current font settings.
func (a *Agg2DFloat) MeasureText(text string) (width, height float64) {
	if a.fontCacheType == RasterFontCache {
		if width, height, _, _, ok := a.textRunMetrics(text); ok {
			return width, height
		}
	}
	width = a.TextWidth(text)
	ascent := a.GetAscender()
	descent := -a.GetDescender()
	if ascent <= 0 && descent <= 0 {
		return width, a.FontHeight()
	}
	return width, ascent + descent
}

// GetTextHeight returns the nominal height of the current font.
func (a *Agg2DFloat) GetTextHeight() float64 {
	_, height := a.MeasureText("X")
	if height > 0 {
		return height
	}
	return a.FontHeight()
}

// TextHints enables or disables font hinting.
func (a *Agg2DFloat) TextHints(hints bool) {
	a.textHints = hints
	if a.fontEngine != nil {
		a.fontEngine.SetHinting(hints)
	}
}

// TextForceAutohint enables or disables FreeType's auto-hinter for raster text.
func (a *Agg2DFloat) TextForceAutohint(force bool) {
	a.textForceAutohint = force
	if a.fontEngine != nil {
		a.fontEngine.SetForceAutohint(force)
	}
}

// GetTextHints returns whether text hinting is currently enabled.
func (a *Agg2DFloat) GetTextHints() bool { return a.textHints }

// ScreenToWorldScalar converts a screen scalar to world units (text metrics).
func (a *Agg2DFloat) ScreenToWorldScalar(scalar float64) float64 {
	x1, y1 := 0.0, 0.0
	x2, y2 := scalar, scalar
	a.ScreenToWorld(&x1, &y1)
	a.ScreenToWorld(&x2, &y2)
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx+dy*dy) / math.Sqrt(2.0)
}

func (a *Agg2DFloat) shapedRasterGlyphs(str string) ([]font.PositionedGlyph, bool) {
	if a.fontCacheType != RasterFontCache || a.fontEngine == nil || str == "" {
		return nil, false
	}
	return a.fontEngine.LayoutText(str)
}

func (a *Agg2DFloat) rasterPlacedGlyphBounds(glyphs []font.PositionedGlyph) (minX, minY, maxX, maxY float64, ok bool) {
	if a.fontEngine == nil {
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

		if !a.fontEngine.PrepareGlyphIndexSubpixel(placedGlyph.GlyphIndex, fracX, fracY) {
			currentX += placedGlyph.XAdvance
			currentY += placedGlyph.YAdvance
			continue
		}

		_, width, height, pitch, left, top, _ := a.fontEngine.CurrentBitmap()
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

func (a *Agg2DFloat) textRunBounds(str string) (minX, minY, maxX, maxY float64, ok bool) {
	fcm := a.fontCacheManager
	if fcm == nil || str == "" {
		return 0, 0, 0, 0, false
	}

	if glyphs, ok := a.shapedRasterGlyphs(str); ok {
		if minX, minY, maxX, maxY, ok := a.rasterPlacedGlyphBounds(glyphs); ok {
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

func (a *Agg2DFloat) textRunMetrics(str string) (width, height, ascent, descent float64, ok bool) {
	minX, minY, maxX, maxY, ok := a.textRunBounds(str)
	if !ok {
		return 0, 0, 0, 0, false
	}

	width = maxX - minX
	height = maxY - minY
	ascent = math.Max(0, -minY)
	descent = math.Max(0, maxY)

	if a.fontCacheType == RasterFontCache {
		width = a.ScreenToWorldScalar(width)
		height = a.ScreenToWorldScalar(height)
		ascent = a.ScreenToWorldScalar(ascent)
		descent = a.ScreenToWorldScalar(descent)
	}

	return width, height, ascent, descent, true
}

func (a *Agg2DFloat) screenToWorldOffset(offset float64) float64 {
	if offset == 0 {
		return 0
	}
	sign := 1.0
	if offset < 0 {
		sign = -1.0
		offset = -offset
	}
	return sign * a.ScreenToWorldScalar(offset)
}

// GetTextBounds reports the actual ink bounds of str relative to the baseline origin.
func (a *Agg2DFloat) GetTextBounds(str string) (x, y, width, height float64) {
	minX, minY, maxX, maxY, ok := a.textRunBounds(str)
	if !ok {
		return 0, 0, 0, 0
	}
	if a.fontCacheType == RasterFontCache {
		minX = a.screenToWorldOffset(minX)
		minY = a.screenToWorldOffset(minY)
		maxX = a.screenToWorldOffset(maxX)
		maxY = a.screenToWorldOffset(maxY)
	}
	return minX, minY, maxX - minX, maxY - minY
}

// textFillColor builds the float solid text color from the public fill color,
// applying master alpha exactly like the 8-bit renderScanlines.
func (a *Agg2DFloat) textFillColor() color.RGBA32[color.Linear] {
	c := colorToRGBA32(a.fillColor)
	if a.masterAlpha != 1.0 {
		c.A *= float32(a.masterAlpha)
	}
	return c
}

func (a *Agg2DFloat) renderShapedRasterMask(startX, startY float64, glyphs []font.PositionedGlyph) bool {
	if a.fontEngine == nil {
		return false
	}

	renderer := a.currentRenderer()
	if renderer == nil {
		return false
	}
	fillColor := a.textFillColor()
	currentX := startX
	currentY := startY
	rendered := false

	for _, placedGlyph := range glyphs {
		glyphX := currentX + placedGlyph.XOffset
		glyphY := currentY + placedGlyph.YOffset
		baseX, fracX := quantizeRasterTextPhaseF26Dot6(glyphX)
		baseY, _ := quantizeRasterTextPhaseF26Dot6(glyphY)
		fracY := quantizeRasterTextYPhaseF26Dot6(glyphY)

		if !a.fontEngine.PrepareGlyphIndexSubpixel(placedGlyph.GlyphIndex, fracX, fracY) {
			currentX += placedGlyph.XAdvance
			currentY += placedGlyph.YAdvance
			continue
		}

		data, width, height, pitch, left, top, pixelMode := a.fontEngine.CurrentBitmap()
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

// TextWidth calculates the width of the given text string in current units.
func (a *Agg2DFloat) TextWidth(str string) float64 {
	if a.gsvFontMode && a.gsvText != nil {
		return a.gsvText.MeasureText(str)
	}

	if a.fontCacheType == RasterFontCache {
		if width, _, _, _, ok := a.textRunMetrics(str); ok {
			return width
		}
	}

	fcm := a.fontCacheManager
	if fcm == nil {
		return 0.0
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
		x += glyph.AdvanceX
		y += glyph.AdvanceY
		first = false
		prevGlyphIndex = glyph.GlyphIndex
	}

	if a.fontCacheType == RasterFontCache {
		return a.ScreenToWorldScalar(x)
	}
	return x
}

// textGSV renders text using the built-in AGG GSV stroke-vector font.
func (a *Agg2DFloat) textGSV(x, y float64, str string, roundOff bool, dx, dy float64) {
	if a.gsvText == nil || str == "" {
		return
	}
	t := a.gsvText
	t.SetText(str)

	alignDx := 0.0
	alignDy := 0.0
	switch a.textAlignX {
	case AlignCenter:
		alignDx = -a.TextWidth(str) * 0.5
	case AlignRight:
		alignDx = -a.TextWidth(str)
	}
	switch a.textAlignY {
	case AlignCenter:
		alignDy = a.fontHeight * 0.5
	case AlignTop:
		alignDy = a.fontHeight
	}

	startX := x + alignDx + dx
	startY := y - alignDy + dy // GSV Y grows down; subtract to shift baseline
	if roundOff {
		startX = float64(int(startX))
		startY = float64(int(startY))
	}

	t.SetStartPoint(startX, startY)

	a.path.RemoveAll()
	t.Rewind(0)
	for {
		vx, vy, cmd := t.Vertex()
		switch cmd {
		case basics.PathCmdMoveTo:
			a.path.MoveTo(vx, vy)
		case basics.PathCmdLineTo:
			a.path.LineTo(vx, vy)
		case basics.PathCmdStop:
			goto vertexLoopDone
		default:
			// GSV does not emit end-poly or curve commands; ignore.
		}
	}
vertexLoopDone:

	pathAdapter := path.NewPathStorageStlVertexSourceAdapter(a.path)
	curvesAdapter := conv.NewConvCurve(pathAdapter)
	strokeAdapter := conv.NewConvStroke(curvesAdapter)
	strokeAdapter.SetWidth(a.fontHeight * 0.08) // ~8 % of glyph height
	strokeAdapter.SetLineCap(basics.RoundCap)
	strokeAdapter.SetLineJoin(basics.RoundJoin)

	transformedStroke := conv.NewConvTransform(strokeAdapter, a.transform)

	a.rasterizer.Reset()
	a.rasterizer.FillingRule(basics.FillNonZero)
	transformedStroke.Rewind(0)
	for {
		sx, sy, cmd := transformedStroke.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		a.rasterizer.AddVertex(sx, sy, uint32(cmd))
	}

	a.renderSolidFillWithColor(a.fillColor)
}

// Text renders text at the specified position, mirroring 8-bit Agg2D.Text.
func (a *Agg2DFloat) Text(x, y float64, str string, roundOff bool, dx, dy float64) {
	if a.gsvFontMode {
		a.textGSV(x, y, str, roundOff, dx, dy)
		return
	}

	fcm := a.fontCacheManager
	if fcm == nil || str == "" {
		return
	}

	shapedGlyphs, haveShapedGlyphs := a.shapedRasterGlyphs(str)

	alignDx := 0.0
	alignDy := 0.0

	if haveShapedGlyphs {
		minX, minY, maxX, maxY, ok := positionedGlyphBounds(shapedGlyphs, fcm)
		if ok {
			switch a.textAlignX {
			case AlignCenter:
				alignDx = -a.ScreenToWorldScalar((minX + maxX) * 0.5)
			case AlignRight:
				alignDx = -a.ScreenToWorldScalar(maxX)
			}
			switch a.textAlignY {
			case AlignCenter:
				alignDy = -a.ScreenToWorldScalar((minY + maxY) * 0.5)
			case AlignTop:
				alignDy = -a.ScreenToWorldScalar(minY)
			}
		}
	} else {
		switch a.textAlignX {
		case AlignCenter:
			alignDx = -a.TextWidth(str) * 0.5
		case AlignRight:
			alignDx = -a.TextWidth(str)
		}

		ascent := a.fontHeight
		glyph := fcm.Glyph(uint('H'))
		if glyph != nil {
			ascent = float64(glyph.Bounds.Y2 - glyph.Bounds.Y1)
		}

		if a.fontCacheType == RasterFontCache {
			ascent = a.ScreenToWorldScalar(ascent)
		}

		switch a.textAlignY {
		case AlignCenter:
			alignDy = -ascent * 0.5
		case AlignTop:
			alignDy = -ascent
		}

		if a.fontEngine != nil && a.fontEngine.GetFlipY() {
			alignDy = -alignDy
		}
	}

	startX := x + alignDx
	startY := y + alignDy

	if roundOff {
		startX = float64(int(startX))
		startY = float64(int(startY))
	}

	startX += dx
	startY += dy

	pathStorage := fcm.PathAdaptor()
	var textTransform *transform.TransAffine
	if a.textAngle != 0.0 {
		textTransform = transform.NewTransAffine()
		textTransform.Translate(-x, -y)
		textTransform.Rotate(a.textAngle)
		textTransform.Translate(x, y)
	}

	if a.fontCacheType == RasterFontCache {
		a.WorldToScreen(&startX, &startY)
	}

	currentX := startX
	currentY := startY

	if haveShapedGlyphs {
		if a.renderShapedRasterMask(currentX, currentY, shapedGlyphs) {
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

		fcm.InitEmbeddedAdaptors(glyph, currentX, currentY)

		switch glyph.DataType {
		case font.GlyphDataOutline:
			a.path.RemoveAll()
			if pathStorage != nil {
				if textTransform != nil {
					a.path.ConcatPath(&transformedPathSource{src: pathStorage, mtx: textTransform}, 0)
				} else {
					a.path.ConcatPath(pathStorage, 0)
				}
				a.DrawPath(FillAndStroke)
			}

		case font.GlyphDataGray8:
			if adaptor := fcm.Gray8Adaptor(); adaptor != nil {
				a.renderGlyphScanlines(adaptor, glyph)
			}

		// GlyphDataMono: Go extension (C++ agg2d.cpp handles only outline+gray8).
		case font.GlyphDataMono:
			if adaptor := fcm.MonoAdaptor(); adaptor != nil {
				a.renderGlyphScanlines(adaptor, glyph)
			}
		}

		currentX += glyph.AdvanceX
		currentY += glyph.AdvanceY
		prevGlyphIndex = glyph.GlyphIndex
		firstGlyph = false
	}
}

// renderGlyphScanlines renders a glyph using scanline data, mirroring the 8-bit flow.
func (a *Agg2DFloat) renderGlyphScanlines(adaptor font.SerializedScanlinesAdaptor, glyph *font.GlyphCache) {
	if a.scanline == nil || glyph == nil || adaptor == nil {
		return
	}
	a.renderScanlines(adaptor, a.scanline, glyph.DataType == font.GlyphDataMono)
}

// renderScanlines renders scanlines using the provided rasterizer and scanline
// adaptors with the current (master-alpha scaled) fill color.
func (a *Agg2DFloat) renderScanlines(ras renscan.RasterizerInterface, sl renscan.ScanlineInterface, mono bool) {
	renderer := a.currentRenderer()
	if renderer == nil {
		return
	}

	fillColor := a.textFillColor()
	if mono {
		renscan.RenderScanlinesBinSolid(ras, sl, renderer, fillColor)
		return
	}
	renscan.RenderScanlinesAASolid(ras, sl, renderer, fillColor)
}
