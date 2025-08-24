// Package agg2d provides the internal AGG2D high-level interface implementation.
// This is a Go port of the C++ Agg2D class from AGG 2.6.
package agg2d

import (
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

// Core types that need to be imported from the parent package
// These will be defined in the public API
type (
	Color          = [4]uint8 // Temporary - will be resolved when we set up proper imports
	BlendMode      = int
	Gradient       = int
	LineCap        = int
	LineJoin       = int
	TextAlignment  = int
	FontCacheType  = int
	ImageFilter    = int
	ImageResample  = int
	ViewportOption = int
)

// Image represents a raster image that can be used as a rendering target.
// This matches the C++ Agg2D::Image structure.
type Image struct {
	renBuf *buffer.RenderingBuffer[uint8]
	Data   []uint8 // Raw pixel data (RGBA format)
	width  int     // Width in pixels
	height int     // Height in pixels
}

// Width returns the image width.
func (img *Image) Width() int {
	return img.width
}

// Height returns the image height.
func (img *Image) Height() int {
	return img.height
}

// Agg2D is the main high-level rendering interface.
// This matches the C++ Agg2D class from the original AGG library.
type Agg2D struct {
	// Rendering buffer
	rbuf *buffer.RenderingBuffer[uint8]

	// Clip box
	clipBox struct{ X1, Y1, X2, Y2 float64 } // RectD equivalent

	// Blend modes
	blendMode       BlendMode
	imageBlendMode  BlendMode
	imageBlendColor Color

	// Scanline and rasterizer
	scanline   *scanline.ScanlineU8
	rasterizer *rasterizer.RasterizerScanlineAA[*rasterizer.RasterizerSlNoClip, rasterizer.RasConvDbl]

	// Rendering components (using interfaces for compatibility)
	pixfmt   interface{} // RGBA pixel format - will be *pixfmt.PixFmtRGBA32
	renBase  interface{} // Base renderer
	renSolid interface{} // Solid color renderer

	// Master alpha and anti-aliasing gamma
	masterAlpha    float64
	antiAliasGamma float64

	// Fill and line colors
	fillColor Color
	lineColor Color

	// Gradients
	fillGradient       [256]Color
	lineGradient       [256]Color
	fillGradientFlag   Gradient
	lineGradientFlag   Gradient
	fillGradientMatrix *transform.TransAffine
	lineGradientMatrix *transform.TransAffine
	fillGradientD1     float64
	lineGradientD1     float64
	fillGradientD2     float64
	lineGradientD2     float64

	// Line attributes
	lineCap   LineCap
	lineJoin  LineJoin
	lineWidth float64

	// Text attributes
	textAngle        float64
	textAlignX       TextAlignment
	textAlignY       TextAlignment
	textHints        bool
	fontHeight       float64
	fontAscent       float64
	fontDescent      float64
	fontCacheType    FontCacheType
	fontEngine       interface{} // FontEngine interface - actual type depends on build tags
	fontCacheManager interface{} // FontCacheManager - manages glyph caching

	// Image filtering
	imageFilter   ImageFilter
	imageResample ImageResample

	// Fill mode
	evenOddFlag bool

	// Path and transformation
	path           *path.PathStorageStl
	transform      *transform.TransAffine
	transformStack interface{} // Optional transform stack for push/pop operations (disabled for now)

	// Converters
	convCurve  *conv.ConvCurve
	convDash   *conv.ConvDash // Optional dash converter (nil when not using dashes)
	convStroke *conv.ConvStroke
}

// TransformStack is defined in transform.go

// Simplified adapters - will be improved later
type pathVertexSourceAdapter struct {
	path *path.PathStorageStl
}

func (p *pathVertexSourceAdapter) Rewind(pathID uint) {
	p.path.Rewind(pathID)
}

// Simplified stub - proper implementation needs interface matching
func (p *pathVertexSourceAdapter) Vertex() (float64, float64, basics.PathCommand) {
	// This is a simplified stub
	return 0, 0, 0 // Using 0 as stop command for now
}

// pixFmtAdapter adapts pixfmt to renderer interfaces
type pixFmtAdapter struct {
	pf *pixfmt.PixFmtRGBA32
}

func (p *pixFmtAdapter) Width() int  { return p.pf.Width() }
func (p *pixFmtAdapter) Height() int { return p.pf.Height() }

// RowPtr method doesn't exist in PixFmtRGBA32, simplified for now
func (p *pixFmtAdapter) RowPtr(y int) []uint8 { return nil }

func (p *pixFmtAdapter) BlendPixel(x, y int, c interface{}, cover basics.Int8u) {
	if col, ok := c.(color.RGBA8[color.Linear]); ok {
		p.pf.BlendPixel(x, y, col, cover)
	}
}

func (p *pixFmtAdapter) CopyPixel(x, y int, c interface{}) {
	if col, ok := c.(color.RGBA8[color.Linear]); ok {
		p.pf.CopyPixel(x, y, col)
	}
}

func (p *pixFmtAdapter) BlendHline(x, y, x2 int, c interface{}, cover basics.Int8u) {
	if col, ok := c.(color.RGBA8[color.Linear]); ok {
		p.pf.BlendHline(x, y, x2, col, cover)
	}
}

// baseRendererAdapter adapts renderer base functionality
type baseRendererAdapter struct {
	pf *pixfmt.PixFmtRGBA32
}

func (b *baseRendererAdapter) Width() int  { return b.pf.Width() }
func (b *baseRendererAdapter) Height() int { return b.pf.Height() }

func (b *baseRendererAdapter) Clear(c interface{}) {
	if col, ok := c.(color.RGBA8[color.Linear]); ok {
		b.pf.Clear(col)
	}
}

func (b *baseRendererAdapter) CopyPixel(x, y int, c interface{}) {
	if col, ok := c.(color.RGBA8[color.Linear]); ok {
		b.pf.CopyPixel(x, y, col)
	}
}

func (b *baseRendererAdapter) BlendPixel(x, y int, c interface{}, cover basics.Int8u) {
	if col, ok := c.(color.RGBA8[color.Linear]); ok {
		b.pf.BlendPixel(x, y, col, cover)
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

// Constants (will need to be resolved with proper imports)
const (
	BlendAlpha          = 0
	BlendDst            = 3
	Solid               = 0
	AlignLeft           = 0
	AlignBottom         = 0
	RasterFontCache     = 0
	ImageFilterBilinear = 0
	NoResample          = 0
	CapRound            = 1
	JoinRound           = 1
)

var (
	Black = Color{0, 0, 0, 255}
	White = Color{255, 255, 255, 255}
)

func NewColor(r, g, b, a uint8) Color {
	return Color{r, g, b, a}
}

// NewAgg2D creates a new AGG2D rendering context.
// This matches the C++ Agg2D constructor.
func NewAgg2D() *Agg2D {
	agg2d := &Agg2D{
		rbuf:               buffer.NewRenderingBuffer[uint8](),
		clipBox:            struct{ X1, Y1, X2, Y2 float64 }{0, 0, 0, 0},
		blendMode:          BlendAlpha,
		imageBlendMode:     BlendDst,
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
		imageFilter:        ImageFilterBilinear,
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

	// Initialize converters - simplified for now
	// pathAdapter := &pathVertexSourceAdapter{path: agg2d.path}
	// agg2d.convCurve = conv.NewConvCurve(pathAdapter)
	// agg2d.convStroke = conv.NewConvStroke(agg2d.convCurve)
	// TODO: Fix interface mismatches

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
	agg2d.ImageFilter(ImageFilterBilinear)
	agg2d.ImageResample(NoResample)
	agg2d.masterAlpha = 1.0
	agg2d.antiAliasGamma = 1.0
	agg2d.blendMode = BlendAlpha

	// Initialize rendering pipeline
	agg2d.initializeRendering()
}

// initializeRendering sets up the rendering pipeline
func (agg2d *Agg2D) initializeRendering() {
	// Initialize pixel format with the attached buffer
	width := agg2d.rbuf.Width()
	height := agg2d.rbuf.Height()

	if width > 0 && height > 0 {
		// Create pixel format
		pf := pixfmt.NewPixFmtRGBA32(agg2d.rbuf)
		agg2d.pixfmt = &pixFmtAdapter{pf: pf}
		agg2d.renBase = &baseRendererAdapter{pf: pf}

		// Initialize rasterizer - simplified for now
		// if agg2d.rasterizer == nil {
		//     clipper := rasterizer.NewRasterizerSlNoClip()
		//     agg2d.rasterizer = rasterizer.NewRasterizerScanlineAA(clipper)
		// }
		// TODO: Fix rasterizer initialization

		// Reset rasterizer clip box - commented out until rasterizer is initialized
		// agg2d.rasterizer.Reset()
		// agg2d.rasterizer.ClipBox(0, 0, float64(width), float64(height))
	}
}

// Core rendering methods needed by Context

// ClipBox sets the clipping rectangle.
func (agg2d *Agg2D) ClipBox(x1, y1, x2, y2 float64) {
	agg2d.clipBox.X1 = x1
	agg2d.clipBox.Y1 = y1
	agg2d.clipBox.X2 = x2
	agg2d.clipBox.Y2 = y2

	// if agg2d.rasterizer != nil {
	//     agg2d.rasterizer.ClipBox(x1, y1, x2, y2)
	// }
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
				buf[offset] = c[0]   // R
				buf[offset+1] = c[1] // G
				buf[offset+2] = c[2] // B
				buf[offset+3] = c[3] // A
			}
		}
	}
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

// LineColor sets the line color.
func (agg2d *Agg2D) LineColor(c Color) {
	agg2d.lineColor = c
	agg2d.lineGradientFlag = Solid
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
	case 0: // CapButt
		agg2d.convStroke.SetLineCap(0) // basics.ButtCap
	case 2: // CapSquare
		agg2d.convStroke.SetLineCap(2) // basics.SquareCap
	case 1: // CapRound
		agg2d.convStroke.SetLineCap(1) // basics.RoundCap
	}
}

// LineJoin sets the line join style.
func (agg2d *Agg2D) LineJoin(join LineJoin) {
	agg2d.lineJoin = join
	switch join {
	case 0: // JoinMiter
		agg2d.convStroke.SetLineJoin(0) // basics.MiterJoin
	case 1: // JoinRound
		agg2d.convStroke.SetLineJoin(1) // basics.RoundJoin
	case 2: // JoinBevel
		agg2d.convStroke.SetLineJoin(2) // basics.BevelJoin
	}
}

// ResetTransformations resets the transformation matrix to identity.
func (agg2d *Agg2D) ResetTransformations() {
	if agg2d.transform != nil {
		agg2d.transform.Reset()
	}
}

// ImageFilter sets the image filtering method.
func (agg2d *Agg2D) ImageFilter(f ImageFilter) {
	agg2d.imageFilter = f
}

// ImageResample sets the image resampling method.
func (agg2d *Agg2D) ImageResample(r ImageResample) {
	agg2d.imageResample = r
}

// TextAlignment sets text alignment.
func (agg2d *Agg2D) TextAlignment(alignX, alignY TextAlignment) {
	agg2d.textAlignX = alignX
	agg2d.textAlignY = alignY
}
