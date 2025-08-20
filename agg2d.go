// Package agg provides the AGG2D high-level interface.
// This is a Go port of the C++ Agg2D class from AGG 2.6.
package agg

import (
	"fmt"
	"os"

	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/conv"
	"agg_go/internal/path"
	"agg_go/internal/pixfmt"
	"agg_go/internal/rasterizer"
	renscan "agg_go/internal/renderer/scanline"
	"agg_go/internal/scanline"
	"agg_go/internal/transform"
)

// pathVertexSourceAdapter adapts path.VertexSource to conv.VertexSource
type pathVertexSourceAdapter struct {
	path *path.PathStorageStl
}

func (p *pathVertexSourceAdapter) Rewind(pathID uint) {
	p.path.Rewind(pathID)
}

func (p *pathVertexSourceAdapter) Vertex() (x, y float64, cmd basics.PathCommand) {
	x, y, cmdUint32 := p.path.NextVertex()
	return x, y, basics.PathCommand(cmdUint32)
}

// BlendMode defines pixel blending operations.
type BlendMode int

const (
	BlendAlpha      BlendMode = iota // Standard alpha blending
	BlendClear                       // Clear destination
	BlendSrc                         // Copy source
	BlendDst                         // Keep destination
	BlendSrcOver                     // Source over destination
	BlendDstOver                     // Destination over source
	BlendSrcIn                       // Source in destination
	BlendDstIn                       // Destination in source
	BlendSrcOut                      // Source out destination
	BlendDstOut                      // Destination out source
	BlendSrcAtop                     // Source atop destination
	BlendDstAtop                     // Destination atop source
	BlendXor                         // XOR operation
	BlendAdd                         // Additive blending
	BlendMultiply                    // Multiply blending
	BlendScreen                      // Screen blending
	BlendOverlay                     // Overlay blending
	BlendDarken                      // Darken blending
	BlendLighten                     // Lighten blending
	BlendColorDodge                  // Color dodge blending
	BlendColorBurn                   // Color burn blending
	BlendHardLight                   // Hard light blending
	BlendSoftLight                   // Soft light blending
	BlendDifference                  // Difference blending
	BlendExclusion                   // Exclusion blending
)

// DrawPathFlag defines how paths should be rendered.
type DrawPathFlag int

const (
	FillOnly          DrawPathFlag = iota // Fill path only
	StrokeOnly                            // Stroke path only
	FillAndStroke                         // Both fill and stroke
	FillWithLineColor                     // Fill with line color
)

// Direction defines path winding direction.
type Direction int

const (
	CW  Direction = iota // Clockwise
	CCW                  // Counter-clockwise
)

// RectD represents a floating-point rectangle.
type RectD struct {
	X1, Y1, X2, Y2 float64
}

// baseColorType is a minimal ColorType for renderer.RendererBase.
// It satisfies the ColorTypeInterface by providing a NoColor value.
type baseColorType struct{}

func (baseColorType) NoColor() interface{} { return color.RGBA8[color.Linear]{} }

// rasVSAdapter adapts a conv.VertexSource to rasterizer.VertexSource
type rasVSAdapter struct{ src conv.VertexSource }

// Rewind adapts conv's Rewind(uint) to rasterizer's Rewind(uint32)
func (a rasVSAdapter) Rewind(pathID uint32) { a.src.Rewind(uint(pathID)) }

// Vertex adapts conv's Vertex() to rasterizer's Vertex(*x,*y) uint32 signature
func (a rasVSAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// baseRendererAdapter adapts pixfmt.PixFmtRGBA32 to the scanline BaseRendererInterface
type baseRendererAdapter struct{ pf *pixfmt.PixFmtRGBA32 }

func (b *baseRendererAdapter) BlendSolidHspan(x, y, length int, c interface{}, covers []basics.Int8u) {
	if col, ok := c.(color.RGBA8[color.Linear]); ok {
		b.pf.BlendSolidHspan(x, y, length, col, covers)
	}
}

func (b *baseRendererAdapter) BlendHline(x, y, x2 int, c interface{}, cover basics.Int8u) {
	if col, ok := c.(color.RGBA8[color.Linear]); ok {
		b.pf.BlendHline(x, y, x2, col, cover)
	}
}

func (b *baseRendererAdapter) BlendColorHspan(x, y, length int, colors []interface{}, covers []basics.Int8u, cover basics.Int8u) {
	// Fallback implementation: blend pixel-by-pixel
	for i := 0; i < length && i < len(colors); i++ {
		col, ok := colors[i].(color.RGBA8[color.Linear])
		if !ok {
			continue
		}
		cvr := cover
		if covers != nil && i < len(covers) {
			cvr = covers[i]
		}
		b.pf.BlendPixel(x+i, y, col, cvr)
	}
}

// scanlineWrapper adapts internal/scanline.ScanlineU8 to renderer/scanline.ScanlineInterface
type scanlineWrapper struct{ sl *scanline.ScanlineU8 }

// Reset implements ResettableScanline
func (w *scanlineWrapper) Reset(minX, maxX int) { w.sl.Reset(minX, maxX) }
func (w *scanlineWrapper) Y() int               { return w.sl.Y() }
func (w *scanlineWrapper) NumSpans() int        { return w.sl.NumSpans() }

// spanIter implements renderer/scanline.ScanlineIterator over our scanline spans
type spanIter struct {
	spans []scanline.Span
	idx   int
}

func (it *spanIter) GetSpan() renscan.SpanData {
	s := it.spans[it.idx]
	return renscan.SpanData{X: int(s.X), Len: int(s.Len), Covers: s.Covers}
}
func (it *spanIter) Next() bool { it.idx++; return it.idx < len(it.spans) }

func (w *scanlineWrapper) Begin() renscan.ScanlineIterator {
	spans := w.sl.Spans()
	if len(spans) == 0 {
		return &spanIter{spans: nil, idx: 0}
	}
	return &spanIter{spans: spans, idx: 0}
}

// rasterizerAdapter adapts internal rasterizer to renderer/scanline.RasterizerInterface
type rasterizerAdapter struct {
	ras *rasterizer.RasterizerScanlineAA[*rasterizer.RasterizerSlNoClip, rasterizer.RasConvDbl]
}

func (r rasterizerAdapter) RewindScanlines() bool { return r.ras.RewindScanlines() }
func (r rasterizerAdapter) MinX() int             { return r.ras.MinX() }
func (r rasterizerAdapter) MaxX() int             { return r.ras.MaxX() }

// rasScanlineAdapter adapts scanline.ScanlineU8 to rasterizer.ScanlineInterface
type rasScanlineAdapter struct{ sl *scanline.ScanlineU8 }

func (a *rasScanlineAdapter) ResetSpans()                 { a.sl.ResetSpans() }
func (a *rasScanlineAdapter) AddCell(x int, cover uint32) { a.sl.AddCell(x, uint(cover)) }
func (a *rasScanlineAdapter) AddSpan(x, length int, cover uint32) {
	a.sl.AddSpan(x, length, uint(cover))
}
func (a *rasScanlineAdapter) Finalize(y int) { a.sl.Finalize(y) }
func (a *rasScanlineAdapter) NumSpans() int  { return a.sl.NumSpans() }

func (r rasterizerAdapter) SweepScanline(sl renscan.ScanlineInterface) bool {
	if w, ok := sl.(*scanlineWrapper); ok {
		return r.ras.SweepScanline(&rasScanlineAdapter{sl: w.sl})
	}
	return false
}

// NewAgg2D creates a new AGG2D rendering context.
// This matches the C++ Agg2D constructor.
func NewAgg2D() *Agg2D {
	agg2d := &Agg2D{
		rbuf:               buffer.NewRenderingBuffer[uint8](),
		clipBox:            struct{ X1, Y1, X2, Y2 float64 }{0, 0, 0, 0},
		blendMode:          int(BlendAlpha),
		imageBlendMode:     int(BlendDst),
		imageBlendColor:    NewColor(0, 0, 0, 255),
		masterAlpha:        1.0,
		antiAliasGamma:     1.0,
		fillColor:          White,
		lineColor:          Black,
		fillGradientFlag:   Solid,
		lineGradientFlag:   Solid,
		fillGradientD1:     0.0,
		lineGradientD1:     0.0,
		fillGradientD2:     100.0,
		lineGradientD2:     100.0,
		textAngle:          0.0,
		textAlignX:         AlignLeft,
		textAlignY:         AlignBottom,
		textHints:          true,
		fontHeight:         0.0,
		fontAscent:         0.0,
		fontDescent:        0.0,
		fontCacheType:      RasterFontCache,
		imageFilter:        Bilinear,
		imageResample:      NoResample,
		lineWidth:          1.0,
		lineCap:            CapRound,
		lineJoin:           JoinRound,
		evenOddFlag:        false,
		path:               path.NewPathStorageStl(),
		transform:          transform.NewTransAffine(),
		fillGradientMatrix: transform.NewTransAffine(),
		lineGradientMatrix: transform.NewTransAffine(),
		scanline:           scanline.NewScanlineU8(),
	}

	// Initialize converters
	pathAdapter := &pathVertexSourceAdapter{path: agg2d.path}
	agg2d.convCurve = conv.NewConvCurve(pathAdapter)
	agg2d.convStroke = conv.NewConvStroke(agg2d.convCurve)

	// Set default line cap and join
	agg2d.LineCap(agg2d.lineCap)
	agg2d.LineJoin(agg2d.lineJoin)

	return agg2d
}

// Attach attaches a rendering buffer to the AGG2D context.
// This matches the C++ Agg2D::attach method.
func (agg2d *Agg2D) Attach(buf []uint8, width, height, stride int) {
	agg2d.rbuf.Attach(buf, width, height, stride)

	// Reset clipping and transformations
	agg2d.ResetTransformations()
	agg2d.LineWidth(1.0)
	agg2d.LineColor(Black)
	agg2d.FillColor(White)
	agg2d.TextAlignment(AlignLeft, AlignBottom)
	agg2d.ClipBox(0, 0, float64(width), float64(height))
	agg2d.LineCap(CapRound)
	agg2d.LineJoin(JoinRound)
	agg2d.ImageFilter(Bilinear)
	agg2d.ImageResample(NoResample)
	agg2d.masterAlpha = 1.0
	agg2d.antiAliasGamma = 1.0
	agg2d.blendMode = int(BlendAlpha)

	// Initialize rendering pipeline
	agg2d.initializeRendering()
}

// AttachImage attaches an AGG2D image to the context.
func (agg2d *Agg2D) AttachImage(img *Image) {
	buf := img.renBuf.Buf()
	agg2d.Attach(buf, img.Width(), img.Height(), img.renBuf.Stride())
}

// SaveImagePPM saves the current image to a PPM file.
// This is a utility function for testing and examples.
func (agg2d *Agg2D) SaveImagePPM(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	width := agg2d.rbuf.Width()
	height := agg2d.rbuf.Height()
	buf := agg2d.rbuf.Buf()
	stride := agg2d.rbuf.Stride()

	// Write PPM header (P3 format for ASCII RGB)
	_, err = fmt.Fprintf(file, "P3\n%d %d\n255\n", width, height)
	if err != nil {
		return fmt.Errorf("failed to write PPM header: %w", err)
	}

	// Write pixel data
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			offset := y*stride + x*4
			if offset+3 < len(buf) {
				r, g, b := buf[offset], buf[offset+1], buf[offset+2]
				_, err = fmt.Fprintf(file, "%d %d %d ", r, g, b)
				if err != nil {
					return fmt.Errorf("failed to write pixel data: %w", err)
				}
			}
		}
		_, err = fmt.Fprintln(file)
		if err != nil {
			return fmt.Errorf("failed to write newline: %w", err)
		}
	}

	return nil
}

// ClipBox sets the clipping rectangle.
func (agg2d *Agg2D) ClipBox(x1, y1, x2, y2 float64) {
	agg2d.clipBox = struct{ X1, Y1, X2, Y2 float64 }{x1, y1, x2, y2}
	if agg2d.rasterizer != nil {
		agg2d.rasterizer.ClipBox(x1, y1, x2, y2)
	}
}

// GetClipBox returns the current clipping rectangle.
func (agg2d *Agg2D) GetClipBox() RectD {
	return RectD{agg2d.clipBox.X1, agg2d.clipBox.Y1, agg2d.clipBox.X2, agg2d.clipBox.Y2}
}

// ClearAll fills the entire buffer with the specified color.
func (agg2d *Agg2D) ClearAll(c Color) {
	// Simple implementation - fill the entire buffer
	buf := agg2d.rbuf.Buf()
	width := agg2d.rbuf.Width()
	height := agg2d.rbuf.Height()
	stride := agg2d.rbuf.Stride()

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			offset := y*stride + x*4
			if offset+3 < len(buf) {
				buf[offset] = c.R
				buf[offset+1] = c.G
				buf[offset+2] = c.B
				buf[offset+3] = c.A
			}
		}
	}
}

// ClearAllRGBA fills the entire buffer with the specified RGBA color.
func (agg2d *Agg2D) ClearAllRGBA(r, g, b, a uint8) {
	agg2d.ClearAll(NewColor(r, g, b, a))
}

// WorldToScreen transforms world coordinates to screen coordinates.
func (agg2d *Agg2D) WorldToScreen(x, y *float64) {
	agg2d.transform.Transform(x, y)
}

// ScreenToWorld transforms screen coordinates to world coordinates.
func (agg2d *Agg2D) ScreenToWorld(x, y *float64) {
	agg2d.transform.InverseTransform(x, y)
}

// FillColor sets the fill color.
func (agg2d *Agg2D) FillColor(c Color) {
	agg2d.fillColor = c
	agg2d.fillGradientFlag = Solid
}

// FillColorRGBA sets the fill color from RGBA values.
func (agg2d *Agg2D) FillColorRGBA(r, g, b, a uint8) {
	agg2d.FillColor(NewColor(r, g, b, a))
}

// LineColor sets the line color.
func (agg2d *Agg2D) LineColor(c Color) {
	agg2d.lineColor = c
	agg2d.lineGradientFlag = Solid
}

// LineColorRGBA sets the line color from RGBA values.
func (agg2d *Agg2D) LineColorRGBA(r, g, b, a uint8) {
	agg2d.LineColor(NewColor(r, g, b, a))
}

// LineWidth sets the line width.
func (agg2d *Agg2D) LineWidth(w float64) {
	agg2d.lineWidth = w
	agg2d.convStroke.SetWidth(w)
}

// LineCap sets the line cap style.
func (agg2d *Agg2D) LineCap(cap LineCap) {
	agg2d.lineCap = cap
	switch cap {
	case CapButt:
		agg2d.convStroke.SetLineCap(basics.ButtCap)
	case CapSquare:
		agg2d.convStroke.SetLineCap(basics.SquareCap)
	case CapRound:
		agg2d.convStroke.SetLineCap(basics.RoundCap)
	}
}

// LineJoin sets the line join style.
func (agg2d *Agg2D) LineJoin(join LineJoin) {
	agg2d.lineJoin = join
	switch join {
	case JoinMiter:
		agg2d.convStroke.SetLineJoin(basics.MiterJoin)
	case JoinRound:
		agg2d.convStroke.SetLineJoin(basics.RoundJoin)
	case JoinBevel:
		agg2d.convStroke.SetLineJoin(basics.BevelJoin)
	}
}

// ResetTransformations resets the transformation matrix to identity.
func (agg2d *Agg2D) ResetTransformations() {
	agg2d.transform.Reset()
}

// ImageFilter sets the image filtering mode.
func (agg2d *Agg2D) ImageFilter(f ImageFilter) {
	agg2d.imageFilter = f
}

// ImageResample sets the image resampling mode.
func (agg2d *Agg2D) ImageResample(r ImageResample) {
	agg2d.imageResample = r
}

// TextAlignment sets the text alignment.
func (agg2d *Agg2D) TextAlignment(alignX, alignY TextAlignment) {
	agg2d.textAlignX = alignX
	agg2d.textAlignY = alignY
}

// Rendering methods - core AGG2D rendering pipeline implementation

// initializeRendering sets up the rendering pipeline when the buffer is attached
func (agg2d *Agg2D) initializeRendering() {
	if agg2d.rbuf == nil {
		return
	}

	// Create rendering buffer U8 from generic buffer
	rbufU8 := &buffer.RenderingBufferU8{}
	rbufU8.Attach(agg2d.rbuf.Buf(), agg2d.rbuf.Width(), agg2d.rbuf.Height(), agg2d.rbuf.Stride())

	// Create pixel format (RGBA32)
	pf := pixfmt.NewPixFmtRGBA32(rbufU8)
	agg2d.pixfmt = pf

	// Minimal renderer chain: BaseRendererAdapter -> RendererScanlineAASolid
	rb := &baseRendererAdapter{pf: pf}
	agg2d.renBase = rb

	rs := renscan.NewRendererScanlineAASolidWithRenderer[*baseRendererAdapter](rb)
	agg2d.renSolid = rs

	// Initialize rasterizer
	agg2d.rasterizer = rasterizer.NewRasterizerScanlineAA[*rasterizer.RasterizerSlNoClip, rasterizer.RasConvDbl](1024) // 1024 cell block limit

	// Initialize rasterizer clip to current clipBox
	agg2d.rasterizer.ClipBox(agg2d.clipBox.X1, agg2d.clipBox.Y1, agg2d.clipBox.X2, agg2d.clipBox.Y2)
}

// renderFill renders the current path with fill color
func (agg2d *Agg2D) renderFill() {
	if agg2d.rasterizer == nil || agg2d.renSolid == nil || agg2d.scanline == nil {
		return
	}

	// Reset rasterizer and add path
	agg2d.rasterizer.Reset()

	// Set fill color on the renderer
	fillColor := color.NewRGBA8[color.Linear](
		agg2d.fillColor.R,
		agg2d.fillColor.G,
		agg2d.fillColor.B,
		agg2d.fillColor.A,
	)
	if setter, ok := agg2d.renSolid.(renscan.ColorSetter); ok {
		setter.SetColor(fillColor)
	}

	// Add the transformed path to the rasterizer using curve converter
	if agg2d.convCurve != nil {
		agg2d.addPathToRasterizer(agg2d.convCurve)
	} else {
		// Fallback to raw path adapter
		pathAdapter := &pathVertexSourceAdapter{path: agg2d.path}
		agg2d.addPathToRasterizer(pathAdapter)
	}

	// Render scanlines
	agg2d.renderScanlines()
}

// renderStroke renders the current path with stroke
func (agg2d *Agg2D) renderStroke() {
	if agg2d.rasterizer == nil || agg2d.renSolid == nil || agg2d.scanline == nil || agg2d.convStroke == nil {
		return
	}

	// Reset rasterizer
	agg2d.rasterizer.Reset()

	// Set stroke color on the renderer
	lineColor := color.NewRGBA8[color.Linear](
		agg2d.lineColor.R,
		agg2d.lineColor.G,
		agg2d.lineColor.B,
		agg2d.lineColor.A,
	)
	if setter, ok := agg2d.renSolid.(renscan.ColorSetter); ok {
		setter.SetColor(lineColor)
	}

	// Add the stroked path to the rasterizer
	agg2d.addPathToRasterizer(agg2d.convStroke)

	// Render scanlines
	agg2d.renderScanlines()
}

// addPathToRasterizer adds a vertex source to the rasterizer with transformation
func (agg2d *Agg2D) addPathToRasterizer(vs conv.VertexSource) {
	if agg2d.rasterizer == nil {
		return
	}

	// Apply current transformation to the vertex source
	xform := conv.NewConvTransform(vs, agg2d.transform)

	// Apply fill rule to rasterizer (even-odd vs non-zero)
	agg2d.applyFillRuleToRasterizer(agg2d.rasterizer)

	// Add transformed path
	agg2d.rasterizer.AddPath(rasVSAdapter{src: xform}, 0)
}

// renderScanlines renders the rasterized scanlines using the current renderer
func (agg2d *Agg2D) renderScanlines() {
	// Use generic scanline rendering helper with adapters
	if ren, ok := agg2d.renSolid.(renscan.RendererInterface); ok {
		rr := rasterizerAdapter{ras: agg2d.rasterizer}
		sw := &scanlineWrapper{sl: agg2d.scanline}
		renscan.RenderScanlines(rr, sw, ren)
	}
}

// renderFillWithLineColor renders the current path filled with line color
func (agg2d *Agg2D) renderFillWithLineColor() {
	// Save current fill color and temporarily use line color
	originalFillColor := agg2d.fillColor
	agg2d.fillColor = agg2d.lineColor
	agg2d.renderFill()
	agg2d.fillColor = originalFillColor
}
