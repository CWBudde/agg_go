// Package agg provides text rendering functionality for the AGG2D high-level interface.
// This implements the text-related methods from the original C++ Agg2D class.
package agg2d

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/font"
	"github.com/MeKo-Christian/agg_go/internal/font/freetype"
	"github.com/MeKo-Christian/agg_go/internal/gsv"
	"github.com/MeKo-Christian/agg_go/internal/path"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/transform"
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
	if agg2d.fontEngine != nil {
		agg2d.fontEngine.SetResolution(agg2d.resolution)
		agg2d.fontEngine.SetFlipY(agg2d.flipText)
		err := agg2d.fontEngine.LoadFont(fileName, 0, renderingType, nil)
		if err != nil {
			return err
		}

		agg2d.fontEngine.SetHinting(agg2d.textHints)

		// Set height based on cache type
		if cacheType == VectorFontCache {
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
	// In standard screen coordinates (Y increases downward) GSV must flip its
	// Y axis so that characters are rendered right-side up.
	// C++ equivalent: agg2d.flipText(true) when the render buffer is NOT flipped.
	agg2d.gsvText.SetFlip(true)
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

// GetTextHints returns whether text hinting is currently enabled.
func (agg2d *Agg2D) GetTextHints() bool {
	return agg2d.textHints
}

// TextWidth calculates the width of the given text string in current units.
// This matches the C++ Agg2D::textWidth() method.
func (agg2d *Agg2D) TextWidth(str string) float64 {
	if agg2d.gsvFontMode && agg2d.gsvText != nil {
		return agg2d.gsvText.MeasureText(str)
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
	firstGlyph := true
	var prevGlyphIndex uint

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

type glyphBitmapRasterizer struct {
	data     []byte
	bounds   basics.Rect[int]
	dataType font.GlyphDataType
	pitch    int
	offsetX  int
	offsetY  int
	row      int
}

func newGlyphBitmapRasterizer(adaptor font.SerializedScanlinesAdaptor, dataType font.GlyphDataType, x, y float64) *glyphBitmapRasterizer {
	if adaptor == nil {
		return nil
	}

	bounds := adaptor.Bounds()
	width := bounds.X2 - bounds.X1
	height := bounds.Y2 - bounds.Y1
	data := adaptor.Data()
	if width <= 0 || height <= 0 || len(data) == 0 {
		return nil
	}

	pitch := len(data) / height
	switch dataType {
	case font.GlyphDataMono:
		minPitch := (width + 7) >> 3
		if pitch < minPitch {
			pitch = minPitch
		}
	default:
		if pitch < width {
			pitch = width
		}
	}
	if pitch <= 0 {
		return nil
	}

	return &glyphBitmapRasterizer{
		data:     data,
		bounds:   bounds,
		dataType: dataType,
		pitch:    pitch,
		offsetX:  basics.IRound(x),
		offsetY:  basics.IRound(y),
	}
}

func (r *glyphBitmapRasterizer) RewindScanlines() bool {
	r.row = 0
	return len(r.data) > 0 && (r.bounds.X2-r.bounds.X1) > 0 && (r.bounds.Y2-r.bounds.Y1) > 0
}

func (r *glyphBitmapRasterizer) MinX() int {
	return r.bounds.X1 + r.offsetX
}

func (r *glyphBitmapRasterizer) MaxX() int {
	return r.bounds.X2 + r.offsetX - 1
}

func (r *glyphBitmapRasterizer) SweepScanline(sl renscan.ScanlineInterface) bool {
	w, ok := sl.(*scanlineWrapper)
	if !ok || w.sl == nil {
		return false
	}

	width := r.bounds.X2 - r.bounds.X1
	height := r.bounds.Y2 - r.bounds.Y1

	for r.row < height {
		row := r.row
		r.row++

		rowStart := row * r.pitch
		if rowStart >= len(r.data) {
			continue
		}
		rowEnd := rowStart + r.pitch
		if rowEnd > len(r.data) {
			rowEnd = len(r.data)
		}
		rowData := r.data[rowStart:rowEnd]

		w.sl.ResetSpans()
		scanY := r.bounds.Y1 + r.offsetY + row
		baseX := r.bounds.X1 + r.offsetX

		switch r.dataType {
		case font.GlyphDataMono:
			runStart := -1
			for col := 0; col < width; col++ {
				byteIdx := col >> 3
				bitSet := false
				if byteIdx < len(rowData) {
					bit := uint(7 - (col & 7))
					bitSet = ((rowData[byteIdx] >> bit) & 0x1) != 0
				}

				if bitSet {
					if runStart < 0 {
						runStart = col
					}
					continue
				}
				if runStart >= 0 {
					w.sl.AddSpan(baseX+runStart, col-runStart, uint(basics.CoverFull))
					runStart = -1
				}
			}
			if runStart >= 0 {
				w.sl.AddSpan(baseX+runStart, width-runStart, uint(basics.CoverFull))
			}

		default:
			runStart := -1
			covers := make([]basics.Int8u, 0, width)
			flush := func() {
				if runStart >= 0 && len(covers) > 0 {
					w.sl.AddCells(baseX+runStart, len(covers), covers)
				}
				runStart = -1
				covers = covers[:0]
			}

			for col := 0; col < width; col++ {
				var cov basics.Int8u
				if col < len(rowData) {
					cov = basics.Int8u(rowData[col])
				}
				if cov == 0 {
					flush()
					continue
				}
				if runStart < 0 {
					runStart = col
				}
				covers = append(covers, cov)
			}
			flush()
		}

		if w.sl.NumSpans() > 0 {
			w.sl.Finalize(scanY)
			return true
		}
	}

	return false
}

// renderGlyphScanlines renders a glyph using scanline data.
// This mirrors AGG2D's render(gray8_adaptor/mono_adaptor, scanline) flow.
func (agg2d *Agg2D) renderGlyphScanlines(adaptor font.SerializedScanlinesAdaptor, glyph *font.GlyphCache, x, y float64) {
	if agg2d.scanline == nil || glyph == nil {
		return
	}

	ras := newGlyphBitmapRasterizer(adaptor, glyph.DataType, x, y)
	if ras == nil {
		return
	}

	agg2d.renderScanlines(ras, &agg2d.slAdapter, glyph.DataType == font.GlyphDataMono)
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
