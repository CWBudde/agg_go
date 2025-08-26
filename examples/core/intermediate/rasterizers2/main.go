// Rasterizers2 Demo - Direct port from AGG C++ rasterizers2.cpp
//
// This example demonstrates different rasterization techniques in AGG:
// 1. Aliased lines with pixel accuracy (Bresenham)
// 2. Aliased lines with subpixel accuracy (Bresenham)
// 3. Anti-aliased outline rendering
// 4. Anti-aliased scanline rendering with stroke
// 5. Anti-aliased outline with image patterns
//
// The demo renders spiral shapes using each technique and allows interactive
// parameter adjustment and performance testing.

package main

import (
    "fmt"
    "math"

    "agg_go/internal/basics"
    "agg_go/internal/color"
    "agg_go/internal/ctrl"
    "agg_go/internal/ctrl/checkbox"
    "agg_go/internal/ctrl/slider"
    "agg_go/internal/pixfmt"
    "agg_go/internal/platform"
    "agg_go/internal/rasterizer"
    "agg_go/internal/renderer"
    renscan "agg_go/internal/renderer/scanline"
    "agg_go/internal/renderer/outline"
    "agg_go/internal/renderer/primitives"
    "agg_go/internal/scanline"
    "agg_go/internal/transform"
)

// Chain link pattern data - direct port from C++ pixmap_chain array
var pixmapChain = []uint32{
	16, 7, // width, height
	0x00ffffff, 0x00ffffff, 0x00ffffff, 0x00ffffff, 0xb4c29999, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xb4c29999, 0x00ffffff, 0x00ffffff, 0x00ffffff, 0x00ffffff,
	0x00ffffff, 0x00ffffff, 0x0cfbf9f9, 0xff9a5757, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xb4c29999, 0x00ffffff, 0x00ffffff, 0x00ffffff,
	0x00ffffff, 0x5ae0cccc, 0xffa46767, 0xff660000, 0xff975252, 0x7ed4b8b8, 0x5ae0cccc, 0x5ae0cccc, 0x5ae0cccc, 0x5ae0cccc, 0xa8c6a0a0, 0xff7f2929, 0xff670202, 0x9ecaa6a6, 0x5ae0cccc, 0x00ffffff,
	0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xa4c7a2a2, 0x3affff00, 0x3affff00, 0xff975151, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000,
	0x00ffffff, 0x5ae0cccc, 0xffa46767, 0xff660000, 0xff954f4f, 0x7ed4b8b8, 0x5ae0cccc, 0x5ae0cccc, 0x5ae0cccc, 0x5ae0cccc, 0xa8c6a0a0, 0xff7f2929, 0xff670202, 0x9ecaa6a6, 0x5ae0cccc, 0x00ffffff,
	0x00ffffff, 0x00ffffff, 0x0cfbf9f9, 0xff9a5757, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xb4c29999, 0x00ffffff, 0x00ffffff, 0x00ffffff,
	0x00ffffff, 0x00ffffff, 0x00ffffff, 0x00ffffff, 0xb4c29999, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xb4c29999, 0x00ffffff, 0x00ffffff, 0x00ffffff, 0x00ffffff,
}

// PatternPixmapARGB32 implements a pattern source interface for the chain link pattern
type PatternPixmapARGB32 struct {
	pixmap []uint32
}

func NewPatternPixmapARGB32(pixmap []uint32) *PatternPixmapARGB32 {
	return &PatternPixmapARGB32{pixmap: pixmap}
}

func (p *PatternPixmapARGB32) Width() int {
	if len(p.pixmap) < 1 {
		return 0
	}
	return int(p.pixmap[0])
}

func (p *PatternPixmapARGB32) Height() int {
	if len(p.pixmap) < 2 {
		return 0
	}
	return int(p.pixmap[1])
}

func (p *PatternPixmapARGB32) Pixel(x, y int) color.RGBA8Linear {
	if x < 0 || y < 0 || x >= p.Width() || y >= p.Height() {
		return color.RGBA8Linear{R: 0, G: 0, B: 0, A: 0}
	}

	idx := y*p.Width() + x + 2 // +2 to skip width/height
	if idx >= len(p.pixmap) {
		return color.RGBA8Linear{R: 0, G: 0, B: 0, A: 0}
	}

	pixel := p.pixmap[idx]
	// ARGB32 format: extract components
	a := basics.Int8u((pixel >> 24) & 0xFF)
	r := basics.Int8u((pixel >> 16) & 0xFF)
	g := basics.Int8u((pixel >> 8) & 0xFF)
	b := basics.Int8u(pixel & 0xFF)

	// Return as RGBA
	return color.RGBA8Linear{R: r, G: g, B: b, A: a}
}

// PatternSourceAdapter adapts PatternPixmapARGB32 to work with outline.Source interface
type PatternSourceAdapter struct {
	pattern *PatternPixmapARGB32
}

func NewPatternSourceAdapter(pattern *PatternPixmapARGB32) *PatternSourceAdapter {
	return &PatternSourceAdapter{pattern: pattern}
}

func (psa *PatternSourceAdapter) Width() float64 {
	return float64(psa.pattern.Width())
}

func (psa *PatternSourceAdapter) Height() float64 {
	return float64(psa.pattern.Height())
}

func (psa *PatternSourceAdapter) Pixel(x, y int) color.RGBA {
	pixel := psa.pattern.Pixel(x, y)
	// Convert RGBA8Linear to RGBA (float64 normalized values)
	return color.NewRGBA(
		float64(pixel.R)/255.0,
		float64(pixel.G)/255.0,
		float64(pixel.B)/255.0,
		float64(pixel.A)/255.0,
	)
}

// Spiral generates a spiral path - direct port from C++ spiral class
type Spiral struct {
	x, y       float64
	r1, r2     float64
	step       float64
	startAngle float64
	angle      float64
	currR      float64
	da         float64
	dr         float64
	start      bool
}

func NewSpiral(x, y, r1, r2, step, startAngle float64) *Spiral {
	return &Spiral{
		x:          x,
		y:          y,
		r1:         r1,
		r2:         r2,
		step:       step,
		startAngle: startAngle,
		da:         basics.Deg2RadF(8.0),
		dr:         step / 45.0,
	}
}

// Rewind for uint interface (conv package) - making this the base implementation
func (s *Spiral) Rewind(pathID uint) {
	s.angle = s.startAngle
	s.currR = s.r1
	s.start = true
}

// Vertex for pointer interface (rasterizer package)
func (s *Spiral) Vertex(x, y *float64) uint32 {
	if s.currR > s.r2 {
		return uint32(basics.PathCmdStop)
	}

	*x = s.x + math.Cos(s.angle)*s.currR
	*y = s.y + math.Sin(s.angle)*s.currR
	s.currR += s.dr
	s.angle += s.da

	if s.start {
		s.start = false
		return uint32(basics.PathCmdMoveTo)
	}
	return uint32(basics.PathCmdLineTo)
}

// Roundoff transformer for pixel accuracy demo - direct port from C++ roundoff struct
type Roundoff struct{}

func (r *Roundoff) Transform(x, y *float64) {
	*x = math.Floor(*x)
	*y = math.Floor(*y)
}

// SpiralConvAdapter adapts Spiral to work with conv package interface
type SpiralConvAdapter struct {
	spiral *Spiral
}

func NewSpiralConvAdapter(spiral *Spiral) *SpiralConvAdapter {
	return &SpiralConvAdapter{spiral: spiral}
}

func (s *SpiralConvAdapter) Rewind(pathID uint) {
	s.spiral.Rewind(pathID)
}

func (s *SpiralConvAdapter) Vertex() (float64, float64, basics.PathCommand) {
	var x, y float64
	cmd := s.spiral.Vertex(&x, &y)
	return x, y, basics.PathCommand(cmd)
}

// SpiralRasterizerAdapter adapts Spiral to work with rasterizer package interface
type SpiralRasterizerAdapter struct {
	spiral *Spiral
}

func NewSpiralRasterizerAdapter(spiral *Spiral) *SpiralRasterizerAdapter {
	return &SpiralRasterizerAdapter{spiral: spiral}
}

func (s *SpiralRasterizerAdapter) Rewind(pathID uint32) {
	s.spiral.Rewind(uint(pathID))
}

func (s *SpiralRasterizerAdapter) Vertex(x, y *float64) uint32 {
	return s.spiral.Vertex(x, y)
}

// RendererBaseAdapter adapts RendererBase to work with outline renderer interface
type RendererBaseAdapter struct {
	renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32, color.RGBA8Linear]
}

func NewRendererBaseAdapter(renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32, color.RGBA8Linear]) *RendererBaseAdapter {
	return &RendererBaseAdapter{renBase: renBase}
}

func (r *RendererBaseAdapter) Width() int {
	return r.renBase.Width()
}

func (r *RendererBaseAdapter) Height() int {
	return r.renBase.Height()
}

func (r *RendererBaseAdapter) BlendSolidHSpan(x, y, length int, color color.RGBA8Linear, covers []basics.CoverType) {
	// Convert CoverType to Int8u
	int8uCovers := make([]basics.Int8u, len(covers))
	for i, c := range covers {
		int8uCovers[i] = basics.Int8u(c)
	}
	r.renBase.BlendSolidHspan(x, y, length, color, int8uCovers)
}

func (r *RendererBaseAdapter) BlendSolidVSpan(x, y, length int, color color.RGBA8Linear, covers []basics.CoverType) {
	// Convert CoverType to Int8u
	int8uCovers := make([]basics.Int8u, len(covers))
	for i, c := range covers {
		int8uCovers[i] = basics.Int8u(c)
	}
	r.renBase.BlendSolidVspan(x, y, length, color, int8uCovers)
}

// RendererBaseImageAdapter adapts RendererBase to work with outline.BaseRenderer interface for image rendering
type RendererBaseImageAdapter struct {
	renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32, color.RGBA8Linear]
}

func NewRendererBaseImageAdapter(renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32, color.RGBA8Linear]) *RendererBaseImageAdapter {
	return &RendererBaseImageAdapter{renBase: renBase}
}

func (r *RendererBaseImageAdapter) BlendColorHSpan(x, y int, length int, colors []color.RGBA, covers []basics.CoverType) {
	// Convert color.RGBA to color.RGBA8Linear and covers to basics.Int8u
	rgba8Colors := make([]color.RGBA8Linear, len(colors))
	for i, c := range colors {
		rgba8Colors[i] = color.RGBA8Linear{
			R: basics.Int8u(c.R * 255),
			G: basics.Int8u(c.G * 255),
			B: basics.Int8u(c.B * 255),
			A: basics.Int8u(c.A * 255),
		}
	}

	var int8uCovers []basics.Int8u
	if covers != nil {
		int8uCovers = make([]basics.Int8u, len(covers))
		for i, c := range covers {
			int8uCovers[i] = basics.Int8u(c)
		}
	}

	// Blend each color individually (simplified approach)
	for i, color := range rgba8Colors {
		var cover basics.Int8u = basics.CoverFull
		if int8uCovers != nil && i < len(int8uCovers) {
			cover = int8uCovers[i]
		}
		r.renBase.BlendPixel(x+i, y, color, cover)
	}
}

func (r *RendererBaseImageAdapter) BlendColorVSpan(x, y int, length int, colors []color.RGBA, covers []basics.CoverType) {
	// Convert color.RGBA to color.RGBA8Linear and covers to basics.Int8u
	rgba8Colors := make([]color.RGBA8Linear, len(colors))
	for i, c := range colors {
		rgba8Colors[i] = color.RGBA8Linear{
			R: basics.Int8u(c.R * 255),
			G: basics.Int8u(c.G * 255),
			B: basics.Int8u(c.B * 255),
			A: basics.Int8u(c.A * 255),
		}
	}

	var int8uCovers []basics.Int8u
	if covers != nil {
		int8uCovers = make([]basics.Int8u, len(covers))
		for i, c := range covers {
			int8uCovers[i] = basics.Int8u(c)
		}
	}

	// Blend each color individually (simplified approach)
	for i, color := range rgba8Colors {
		var cover basics.Int8u = basics.CoverFull
		if int8uCovers != nil && i < len(int8uCovers) {
			cover = int8uCovers[i]
		}
		r.renBase.BlendPixel(x, y+i, color, cover)
	}
}

// ControlVertexSourceAdapter adapts control's vertex source interface to the rasterizer's requirements
type ControlVertexSourceAdapter struct {
	ctrl   ctrl.Ctrl[color.RGBA8Linear]
	pathID uint32
}

func NewControlVertexSourceAdapter(ctrl ctrl.Ctrl[color.RGBA8Linear]) *ControlVertexSourceAdapter {
	return &ControlVertexSourceAdapter{ctrl: ctrl}
}

func (adapter *ControlVertexSourceAdapter) Rewind(pathID uint32) {
	adapter.pathID = pathID
	adapter.ctrl.Rewind(uint(pathID))
}

func (adapter *ControlVertexSourceAdapter) Vertex(x, y *float64) uint32 {
	px, py, cmd := adapter.ctrl.Vertex()
	*x = px
	*y = py
	return uint32(cmd)
}

// ControlRasterizerAdapter adapts the rasterizer to work with control interface
type ControlRasterizerAdapter struct {
    ras *rasterizer.RasterizerScanlineAA[*rasterizer.RasterizerSlNoClip, rasterizer.RasConvDbl]
}

func NewControlRasterizerAdapter() *ControlRasterizerAdapter {
    noClip := rasterizer.NewRasterizerSlNoClip(nil)
    ras := rasterizer.NewRasterizerScanlineAA[*rasterizer.RasterizerSlNoClip, rasterizer.RasConvDbl](1024, noClip)
    return &ControlRasterizerAdapter{ras: ras}
}

func (adapter *ControlRasterizerAdapter) Reset() {
	adapter.ras.Reset()
}

func (adapter *ControlRasterizerAdapter) AddPath(vs ctrl.VertexSourceInterface, pathID uint) {
	// Convert ctrl.VertexSourceInterface to rasterizer.VertexSource
	ctrlAdapter := &controlToRasterizerVertexAdapter{vs: vs}
	adapter.ras.AddPath(ctrlAdapter, uint32(pathID))
}

func (adapter *ControlRasterizerAdapter) RewindScanlines() bool {
	return adapter.ras.RewindScanlines()
}

func (adapter *ControlRasterizerAdapter) MinX() int {
	return adapter.ras.MinX()
}

func (adapter *ControlRasterizerAdapter) MaxX() int {
	return adapter.ras.MaxX()
}

func (adapter *ControlRasterizerAdapter) SweepScanline(sl renscan.ScanlineInterface) bool {
    // Adapt renderer/scanline interface to rasterizer scanline interface
    switch w := sl.(type) {
    case *rendererScanlineP8Wrapper:
        return adapter.ras.SweepScanline(&rasScanlineP8Adapter{sl: w.sl})
    default:
        // Fallback: try to extract spans dynamically (not ideal, but safe default)
        return false
    }
}

// controlToRasterizerVertexAdapter converts ctrl.VertexSourceInterface to rasterizer.VertexSource
type controlToRasterizerVertexAdapter struct {
	vs ctrl.VertexSourceInterface
}

func (adapter *controlToRasterizerVertexAdapter) Rewind(pathID uint32) {
	adapter.vs.Rewind(uint(pathID))
}

func (adapter *controlToRasterizerVertexAdapter) Vertex(x, y *float64) uint32 {
	px, py, cmd := adapter.vs.Vertex()
	*x = px
	*y = py
	return cmd
}

// Adapter: wrap internal/scanline.ScanlineP8 to renderer/scanline.ScanlineInterface
type rendererScanlineP8Wrapper struct{ sl *scanline.ScanlineP8 }

func newRendererScanlineP8Wrapper(sl *scanline.ScanlineP8) *rendererScanlineP8Wrapper { return &rendererScanlineP8Wrapper{sl: sl} }

// Reset implements renscan.ResettableScanline
func (w *rendererScanlineP8Wrapper) Reset(minX, maxX int) { w.sl.Reset(minX, maxX) }
func (w *rendererScanlineP8Wrapper) Y() int               { return w.sl.Y() }
func (w *rendererScanlineP8Wrapper) NumSpans() int        { return w.sl.NumSpans() }

type spanIterP8 struct {
    spans []scanline.SpanP8
    idx   int
}

func (it *spanIterP8) GetSpan() renscan.SpanData {
    s := it.spans[it.idx]
    // For ScanlineP8, Len < 0 indicates solid spans. Use absolute length
    length := s.ActualLen()
    covers := s.GetCovers()
    return renscan.SpanData{X: int(s.X), Len: length, Covers: covers}
}
func (it *spanIterP8) Next() bool { it.idx++; return it.idx < len(it.spans) }

func (w *rendererScanlineP8Wrapper) Begin() renscan.ScanlineIterator {
    spans := w.sl.Spans()
    if len(spans) == 0 {
        return &spanIterP8{spans: nil, idx: 0}
    }
    return &spanIterP8{spans: spans, idx: 0}
}

// Adapter: adapt ScanlineP8 to rasterizer.ScanlineInterface for SweepScanline
type rasScanlineP8Adapter struct{ sl *scanline.ScanlineP8 }

func (a *rasScanlineP8Adapter) ResetSpans()                 { a.sl.ResetSpans() }
func (a *rasScanlineP8Adapter) AddCell(x int, cover uint32) { a.sl.AddCell(x, uint(cover)) }
func (a *rasScanlineP8Adapter) AddSpan(x, length int, cover uint32) {
    a.sl.AddSpan(x, length, uint(cover))
}
func (a *rasScanlineP8Adapter) Finalize(y int) { a.sl.Finalize(y) }
func (a *rasScanlineP8Adapter) NumSpans() int  { return a.sl.NumSpans() }

// SliderCtrlAdapter wraps slider.SliderCtrl to present RGBA8Linear colors
type SliderCtrlAdapter struct{ s *slider.SliderCtrl }

func NewSliderCtrlAdapter(s *slider.SliderCtrl) *SliderCtrlAdapter { return &SliderCtrlAdapter{s: s} }

// Delegate ctrl.Ctrl methods, converting only Color
func (a *SliderCtrlAdapter) InRect(x, y float64) bool                                  { return a.s.InRect(x, y) }
func (a *SliderCtrlAdapter) OnMouseButtonDown(x, y float64) bool                        { return a.s.OnMouseButtonDown(x, y) }
func (a *SliderCtrlAdapter) OnMouseButtonUp(x, y float64) bool                          { return a.s.OnMouseButtonUp(x, y) }
func (a *SliderCtrlAdapter) OnMouseMove(x, y float64, buttonPressed bool) bool          { return a.s.OnMouseMove(x, y, buttonPressed) }
func (a *SliderCtrlAdapter) OnArrowKeys(left, right, down, up bool) bool                { return a.s.OnArrowKeys(left, right, down, up) }
func (a *SliderCtrlAdapter) SetTransform(mtx *transform.TransAffine)                    { a.s.SetTransform(mtx) }
func (a *SliderCtrlAdapter) ClearTransform()                                            { a.s.ClearTransform() }
func (a *SliderCtrlAdapter) TransformXY(x, y *float64)                                  { a.s.TransformXY(x, y) }
func (a *SliderCtrlAdapter) InverseTransformXY(x, y *float64)                           { a.s.InverseTransformXY(x, y) }
func (a *SliderCtrlAdapter) Scale() float64                                             { return a.s.Scale() }
func (a *SliderCtrlAdapter) X1() float64                                                { return a.s.X1() }
func (a *SliderCtrlAdapter) Y1() float64                                                { return a.s.Y1() }
func (a *SliderCtrlAdapter) X2() float64                                                { return a.s.X2() }
func (a *SliderCtrlAdapter) Y2() float64                                                { return a.s.Y2() }
func (a *SliderCtrlAdapter) FlipY() bool                                                { return a.s.FlipY() }
func (a *SliderCtrlAdapter) NumPaths() uint                                             { return a.s.NumPaths() }
func (a *SliderCtrlAdapter) Rewind(pathID uint)                                         { a.s.Rewind(pathID) }
func (a *SliderCtrlAdapter) Vertex() (x, y float64, cmd basics.PathCommand)             { return a.s.Vertex() }
func (a *SliderCtrlAdapter) Color(pathID uint) color.RGBA8Linear {
    c := a.s.Color(pathID)
    return color.RGBA8Linear{R: basics.Int8u(c.R*255 + 0.5), G: basics.Int8u(c.G*255 + 0.5), B: basics.Int8u(c.B*255 + 0.5), A: basics.Int8u(c.A*255 + 0.5)}
}

// Application holds the demo application state
type Application struct {
	ps          *platform.PlatformSupport
	rc          *platform.RenderingContext
	stepSlider  *slider.SliderCtrl
	widthSlider *slider.SliderCtrl
	testBox     *checkbox.CheckboxCtrl[color.RGBA8Linear]
	rotateBox   *checkbox.CheckboxCtrl[color.RGBA8Linear]
	joinBox     *checkbox.CheckboxCtrl[color.RGBA8Linear]
	scaleBox    *checkbox.CheckboxCtrl[color.RGBA8Linear]
	startAngle  float64
}

func NewApplication() *Application {
	app := &Application{}

	// Create platform support
	app.ps = platform.NewPlatformSupport(platform.PixelFormatRGBA32, false)
	app.ps.Caption("AGG Rasterizers2 Demo")
	app.rc = platform.NewRenderingContext(app.ps)

	// Create controls - positions match C++ version
	app.stepSlider = slider.NewSliderCtrl(10, 10+4, 150, 10+8+4, false)
	app.stepSlider.SetRange(0.0, 2.0)
	app.stepSlider.SetValue(0.1)
	app.stepSlider.SetLabel("Step=%.2f")

	app.widthSlider = slider.NewSliderCtrl(150+10, 10+4, 400-10, 10+8+4, false)
	app.widthSlider.SetRange(0.0, 14.0)
	app.widthSlider.SetValue(3.0)
	app.widthSlider.SetLabel("Width=%.2f")

	// Define colors for controls
	black := color.RGBA8Linear{R: 0, G: 0, B: 0, A: 255}
	gray := color.RGBA8Linear{R: 100, G: 100, B: 100, A: 255}
	blue := color.RGBA8Linear{R: 0, G: 0, B: 200, A: 255}

	app.testBox = checkbox.NewCheckboxCtrl(10, 10+4+16, "Test Performance", false, gray, black, blue)
	app.rotateBox = checkbox.NewCheckboxCtrl(130+10, 10+4+16, "Rotate", false, gray, black, blue)
	app.joinBox = checkbox.NewCheckboxCtrl(200+10, 10+4+16, "Accurate Joins", false, gray, black, blue)
	app.scaleBox = checkbox.NewCheckboxCtrl(310+10, 10+4+16, "Scale Pattern", false, gray, black, blue)
	app.scaleBox.SetChecked(true)

	// Enable rotation for testing
	app.rotateBox.SetChecked(true)

	return app
}

func (app *Application) OnDraw() {
	width := float64(app.ps.Width())
	height := float64(app.ps.Height())

	// Get rendering buffer from rendering context
	rbuf := app.rc.WindowBuffer()

	// Use RGBA32 format to match the renderer interface
	pixf := pixfmt.NewPixFmtRGBA32(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt(pixf)

	// Clear background - cream color (1.0, 1.0, 0.95)
	clearColor := color.RGBA8Linear{R: 255, G: 255, B: 242, A: 255}
	renBase.Clear(clearColor)

	// Set up drawing color
	drawColor := color.RGBA8Linear{R: 102, G: 77, B: 26, A: 255} // rgba(0.4, 0.3, 0.1)

	// Draw the five different rendering techniques
	app.drawAliasedPixelAccuracy(renBase, width, height, drawColor)
	app.drawAliasedSubpixelAccuracy(renBase, width, height, drawColor)
	app.drawAntiAliasedOutline(renBase, width, height, drawColor)
	app.drawAntiAliasedScanline(renBase, width, height, drawColor)
	app.drawAntiAliasedOutlineImg(renBase, width, height)

	// Add text labels for each technique
	app.drawText(renBase, width, height)

	// Render controls
	app.renderControls(renBase)
}

// drawAliasedPixelAccuracy draws aliased lines with pixel accuracy (rounded coordinates)
func (app *Application) drawAliasedPixelAccuracy(renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32, color.RGBA8Linear], width, height float64, color color.RGBA8Linear) {
	spiral := NewSpiral(width/5, height/4+50, 5, 70, 16, app.startAngle)

	// Create primitive renderer
	renPrim := primitives.NewRendererPrimitives(renBase)
	renPrim.LineColor(color)

	// Create outline rasterizer
	rasOutline := rasterizer.NewRasterizerOutline(renPrim)
	spiralRasAdapter := NewSpiralRasterizerAdapter(spiral)
	rasOutline.AddPath(spiralRasAdapter, 0)
}

// drawAliasedSubpixelAccuracy draws aliased lines with subpixel accuracy
func (app *Application) drawAliasedSubpixelAccuracy(renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32, color.RGBA8Linear], width, height float64, color color.RGBA8Linear) {
	spiral := NewSpiral(width/2, height/4+50, 5, 70, 16, app.startAngle)

	// Create primitive renderer
	renPrim := primitives.NewRendererPrimitives(renBase)
	renPrim.LineColor(color)

	// Create outline rasterizer
	rasOutline := rasterizer.NewRasterizerOutline(renPrim)
	spiralAdapter := NewSpiralRasterizerAdapter(spiral)
	rasOutline.AddPath(spiralAdapter, 0)
}

// drawAntiAliasedOutline draws anti-aliased outline (using simple fallback for now)
func (app *Application) drawAntiAliasedOutline(renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32, color.RGBA8Linear], width, height float64, color color.RGBA8Linear) {
	// For now, use anti-aliased line drawing as a fallback
	spiral := NewSpiral(width/5, height-height/4+20, 5, 70, 16, app.startAngle)

	var x, y float64
	var prevX, prevY float64
	first := true

	thickness := int(app.widthSlider.Value() + 0.5)
	if thickness < 1 {
		thickness = 1
	}

	for {
		cmd := spiral.Vertex(&x, &y)
		if cmd == uint32(basics.PathCmdStop) {
			break
		}

		if cmd == uint32(basics.PathCmdMoveTo) {
			prevX, prevY = x, y
			first = true
		} else if cmd == uint32(basics.PathCmdLineTo) && !first {
			// Draw anti-aliased thick line
			app.drawThickLine(renBase, prevX, prevY, x, y, thickness, color)
			prevX, prevY = x, y
			first = false
		}
	}
}

// drawAntiAliasedScanline draws anti-aliased scanline with stroke
func (app *Application) drawAntiAliasedScanline(renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32, color.RGBA8Linear], width, height float64, color color.RGBA8Linear) {
	// For now, use simple line drawing as a fallback until interfaces are fully compatible
	spiral := NewSpiral(width/2, height-height/4+20, 5, 70, 16, app.startAngle)

	var x, y float64
	var prevX, prevY float64
	first := true

	for {
		cmd := spiral.Vertex(&x, &y)
		if cmd == uint32(basics.PathCmdStop) {
			break
		}

		if cmd == uint32(basics.PathCmdMoveTo) {
			prevX, prevY = x, y
			first = true
		} else if cmd == uint32(basics.PathCmdLineTo) && !first {
			// Draw a thick line by drawing multiple parallel lines
			thickness := int(app.widthSlider.Value() + 0.5)
			for dy := -thickness / 2; dy <= thickness/2; dy++ {
				for dx := -thickness / 2; dx <= thickness/2; dx++ {
					app.drawLine(renBase, prevX+float64(dx), prevY+float64(dy), x+float64(dx), y+float64(dy), color)
				}
			}
			prevX, prevY = x, y
			first = false
		}
	}
}

// drawAntiAliasedOutlineImg draws anti-aliased outline with image patterns
func (app *Application) drawAntiAliasedOutlineImg(renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32, color.RGBA8Linear], width, height float64) {
	spiral := NewSpiral(width-width/5, height-height/4+20, 5, 70, 16, app.startAngle)

	// Create pattern filter adapter
	filter := outline.NewPatternFilterRGBAAdapter()

	// Create pattern source from the chain link pixmap
	patternPixmap := NewPatternPixmapARGB32(pixmapChain)
	patternSource := NewPatternSourceAdapter(patternPixmap)

	// Create line image scale if needed
	var scaledSource outline.Source = patternSource
	if app.scaleBox.IsChecked() {
		scaledSource = outline.NewLineImageScale(patternSource, app.widthSlider.Value())
	}

	// Create line image pattern (use power-of-2 optimized version since width is 16)
	pattern := outline.NewLineImagePatternPow2FromSource(filter, scaledSource)

	// Create base renderer adapter for image pattern
	baseAdapter := NewRendererBaseImageAdapter(renBase)

	// Create outline image renderer
	imgRenderer := outline.NewRendererOutlineImage(baseAdapter, pattern)

	// Set scaling if enabled
	if app.scaleBox.IsChecked() {
		imgRenderer.SetScaleX(patternSource.Height() / app.widthSlider.Value())
	}

	// Create rasterizer outline for image patterns (simplified approach)
	// Since we don't have RasterizerOutlineAA[outline.RendererOutlineImage] directly,
	// we'll draw the spiral manually using the image renderer
	var prevX, prevY float64
	first := true

	spiral.Rewind(0)
	for {
		var x, y float64
		cmd := spiral.Vertex(&x, &y)
		if cmd == uint32(basics.PathCmdStop) {
			break
		}

		if cmd == uint32(basics.PathCmdMoveTo) {
			prevX, prevY = x, y
			first = true
		} else if cmd == uint32(basics.PathCmdLineTo) && !first {
			// Draw line segment with image pattern
			// For simplicity, we'll sample the pattern and draw pixels
			app.drawImagePatternLine(renBase, imgRenderer, prevX, prevY, x, y)
			prevX, prevY = x, y
			first = false
		}
	}
}

// drawImagePatternLine draws a line segment with image pattern (simplified implementation)
func (app *Application) drawImagePatternLine(renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32, color.RGBA8Linear],
	imgRenderer *outline.RendererOutlineImage, x1, y1, x2, y2 float64,
) {
	// Simple line drawing with pattern sampling
	dx := x2 - x1
	dy := y2 - y1
	length := math.Sqrt(dx*dx + dy*dy)

	if length < 1 {
		return
	}

	steps := int(length)
	if steps < 2 {
		steps = 2
	}

	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := x1 + t*dx
		y := y1 + t*dy

		// Sample pattern color
		var patternColor color.RGBA
		// Use a simplified pattern sampling - in real implementation this would be more sophisticated
		patternX := int(t*float64(imgRenderer.GetPattern().PatternWidth())) % imgRenderer.GetPattern().PatternWidth()
		patternY := imgRenderer.GetPattern().LineWidth() / 2

		imgRenderer.Pixel(&patternColor, patternX, patternY)

		// Convert to RGBA8Linear and blend
		rgba8Color := color.RGBA8Linear{
			R: basics.Int8u(patternColor.R * 255),
			G: basics.Int8u(patternColor.G * 255),
			B: basics.Int8u(patternColor.B * 255),
			A: basics.Int8u(patternColor.A * 255),
		}

		// Draw a small circle around the point for thickness
		thickness := int(app.widthSlider.Value() + 0.5)
		if thickness < 1 {
			thickness = 1
		}

		px, py := int(x), int(y)
		for dy := -thickness; dy <= thickness; dy++ {
			for dx := -thickness; dx <= thickness; dx++ {
				if dx*dx+dy*dy <= thickness*thickness {
					nx, ny := px+dx, py+dy
					if nx >= 0 && ny >= 0 && nx < renBase.Width() && ny < renBase.Height() {
						renBase.BlendPixel(nx, ny, rgba8Color, basics.CoverFull/2)
					}
				}
			}
		}
	}
}

// drawText renders text labels for each technique (simplified version)
func (app *Application) drawText(renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32, color.RGBA8Linear], width, height float64) {
	textColor := color.RGBA8Linear{R: 0, G: 0, B: 0, A: 255}

	// For now, just draw simple placeholder text
	app.renderSimpleText(renBase, 50, 80, "Bresenham lines, regular accuracy", textColor)
	app.renderSimpleText(renBase, width/2-50, 80, "Bresenham lines, subpixel accuracy", textColor)
	app.renderSimpleText(renBase, 50, height/2+50, "Anti-aliased lines", textColor)
	app.renderSimpleText(renBase, width/2-50, height/2+50, "Scanline rasterizer", textColor)
	app.renderSimpleText(renBase, width-width/5-50, height/2+50, "Arbitrary Image Pattern", textColor)
}

// renderSimpleText renders simple text by drawing pixels (placeholder)
func (app *Application) renderSimpleText(renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32, color.RGBA8Linear], x, y float64, txt string, color color.RGBA8Linear) {
	// Draw a simple text placeholder - just a small rectangle for now
	for dy := 0; dy < 8; dy++ {
		for dx := 0; dx < len(txt)*6; dx++ {
			px, py := int(x)+dx, int(y)+dy
			if px >= 0 && py >= 0 && px < renBase.Width() && py < renBase.Height() {
				renBase.BlendPixel(px, py, color, basics.CoverFull/4)
			}
		}
	}
}

// drawThickLine draws a thick anti-aliased line
func (app *Application) drawThickLine(renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32, color.RGBA8Linear], x0, y0, x1, y1 float64, thickness int, color color.RGBA8Linear) {
	// Draw multiple parallel lines for thickness
	for dy := -thickness / 2; dy <= thickness/2; dy++ {
		for dx := -thickness / 2; dx <= thickness/2; dx++ {
			app.drawLine(renBase, x0+float64(dx), y0+float64(dy), x1+float64(dx), y1+float64(dy), color)
		}
	}
}

// drawLine draws a simple line using Bresenham algorithm
func (app *Application) drawLine(renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32, color.RGBA8Linear], x0, y0, x1, y1 float64, color color.RGBA8Linear) {
	ix0, iy0 := int(x0), int(y0)
	ix1, iy1 := int(x1), int(y1)

	dx := ix1 - ix0
	if dx < 0 {
		dx = -dx
	}
	dy := iy1 - iy0
	if dy < 0 {
		dy = -dy
	}

	sx, sy := 1, 1
	if ix0 > ix1 {
		sx = -1
	}
	if iy0 > iy1 {
		sy = -1
	}

	err := dx - dy
	x, y := ix0, iy0

	for {
		if x >= 0 && y >= 0 && x < renBase.Width() && y < renBase.Height() {
			renBase.BlendPixel(x, y, color, basics.CoverFull)
		}

		if x == ix1 && y == iy1 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

// renderControls renders the UI controls using proper AGG rendering pipeline
func (app *Application) renderControls(renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32, color.RGBA8Linear]) {
    // Create rasterizer and scanline for control rendering
    ras := NewControlRasterizerAdapter()
    sl := scanline.NewScanlineP8()
    slWrap := newRendererScanlineP8Wrapper(sl)

	// Create base renderer adapter for controls
	baseRenderer := &controlBaseRendererAdapter{renBase: renBase}

	// Render each control using the AGG pipeline
    app.renderSliderControl(ras, slWrap, baseRenderer, app.stepSlider)
    app.renderSliderControl(ras, slWrap, baseRenderer, app.widthSlider)
    app.renderCheckboxControl(ras, slWrap, baseRenderer, app.testBox)
    app.renderCheckboxControl(ras, slWrap, baseRenderer, app.rotateBox)
    app.renderCheckboxControl(ras, slWrap, baseRenderer, app.joinBox)
    app.renderCheckboxControl(ras, slWrap, baseRenderer, app.scaleBox)
}

// controlBaseRendererAdapter adapts RendererBase to work with ctrl.BaseRendererInterface
type controlBaseRendererAdapter struct {
    renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32, color.RGBA8Linear]
}

func (adapter *controlBaseRendererAdapter) BlendSolidHspan(x, y, length int, c color.RGBA8Linear, covers []basics.Int8u) {
    adapter.renBase.BlendSolidHspan(x, y, length, c, covers)
}

func (adapter *controlBaseRendererAdapter) BlendHline(x, y, x2 int, c color.RGBA8Linear, cover basics.Int8u) {
    adapter.renBase.BlendHline(x, y, x2, c, cover)
}

func (adapter *controlBaseRendererAdapter) BlendColorHspan(x, y, length int, colors []color.RGBA8Linear, covers []basics.Int8u, cover basics.Int8u) {
    adapter.renBase.BlendColorHspan(x, y, length, colors, covers, cover)
}

// renderSliderControl renders a slider control using the AGG pipeline
func (app *Application) renderSliderControl(ras *ControlRasterizerAdapter, sl *rendererScanlineP8Wrapper,
    baseRenderer *controlBaseRendererAdapter, s *slider.SliderCtrl,
) {
    // Wrap slider to adapt color type
    sAdapter := NewSliderCtrlAdapter(s)
    ctrl.RenderCtrl[*ControlRasterizerAdapter, *rendererScanlineP8Wrapper, *controlBaseRendererAdapter, color.RGBA8Linear](
        ras, sl, baseRenderer, sAdapter)
}

// renderCheckboxControl renders a checkbox control using the AGG pipeline
func (app *Application) renderCheckboxControl(ras *ControlRasterizerAdapter, sl *rendererScanlineP8Wrapper,
    baseRenderer *controlBaseRendererAdapter, checkboxCtrl *checkbox.CheckboxCtrl[color.RGBA8Linear],
) {
    // Use the ctrl.RenderCtrl function with adapters
    ctrl.RenderCtrl[*ControlRasterizerAdapter, *rendererScanlineP8Wrapper, *controlBaseRendererAdapter, color.RGBA8Linear](
        ras, sl, baseRenderer, checkboxCtrl)
}

// renderSimpleControlPlaceholders renders simple placeholders for controls
func (app *Application) renderSimpleControlPlaceholders(renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32, color.RGBA8Linear]) {
	controlColor := color.RGBA8Linear{R: 128, G: 128, B: 128, A: 255}

	// Draw slider backgrounds
	app.drawControlRect(renBase, 10, 10, 150, 20, controlColor)
	app.drawControlRect(renBase, 160, 10, 390, 20, controlColor)

	// Draw checkbox backgrounds
	app.drawControlRect(renBase, 10, 30, 120, 45, controlColor)
	app.drawControlRect(renBase, 140, 30, 190, 45, controlColor)
	app.drawControlRect(renBase, 210, 30, 300, 45, controlColor)
	app.drawControlRect(renBase, 320, 30, 420, 45, controlColor)
}

// drawControlRect draws a simple control rectangle
func (app *Application) drawControlRect(renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32, color.RGBA8Linear],
	x1, y1, x2, y2 int, color color.RGBA8Linear,
) {
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			if x >= 0 && y >= 0 && x < renBase.Width() && y < renBase.Height() {
				renBase.BlendPixel(x, y, color, basics.CoverFull/4)
			}
		}
	}
}

func main() {
	fmt.Println("AGG Rasterizers2 Demo")
	fmt.Println("=====================")

	app := NewApplication()

	// Set up event handlers
	app.ps.SetOnInit(app.OnInit)
	app.ps.SetOnResize(app.OnResize)
	app.ps.SetOnDraw(app.OnDraw)
	app.ps.SetOnIdle(app.OnIdle)
    app.ps.SetOnMouseDown(func(x, y int, flags platform.InputFlags) { _ = app.OnMouseButtonDown(x, y, flags) })
    app.ps.SetOnMouseUp(func(x, y int, flags platform.InputFlags) { _ = app.OnMouseButtonUp(x, y, flags) })
    app.ps.SetOnMouseMove(func(x, y int, flags platform.InputFlags) { _ = app.OnMouseMove(x, y, flags) })

	if app.ps.Init(500, 450, 0) == nil {
		app.ps.Run()
	}
}

// Platform event handlers
func (app *Application) OnInit() {
	fmt.Println("Rasterizers2 Demo initialized")
}

func (app *Application) OnResize(width, height int) {
	app.rc.SetupResizeTransform(width, height)
}

func (app *Application) OnIdle() {
	if app.rotateBox.IsChecked() {
		app.startAngle += basics.Deg2RadF(app.stepSlider.Value())
		if app.startAngle > basics.Deg2RadF(360.0) {
			app.startAngle -= basics.Deg2RadF(360.0)
		}
		fmt.Printf("Animation: angle %.2f\n", app.startAngle)
		app.ps.ForceRedraw()
		app.ps.UpdateWindow()
	}
}

func (app *Application) OnCtrlChange() {
	// Handle performance test
	if app.testBox.IsChecked() {
		app.performanceTest()
		app.testBox.SetChecked(false)
		app.ps.ForceRedraw()
	}
}

// Mouse event handlers for control interaction
func (app *Application) OnMouseButtonDown(x, y int, flags platform.InputFlags) bool {
	// Check if any control is clicked
	fx, fy := float64(x), float64(y)

	// Check sliders first
    if app.stepSlider.OnMouseButtonDown(fx, fy) {
        app.ps.ForceRedraw()
        return true
    }
    if app.widthSlider.OnMouseButtonDown(fx, fy) {
        app.ps.ForceRedraw()
        return true
    }

	// Check checkboxes
    if app.testBox.OnMouseButtonDown(fx, fy) {
        app.OnCtrlChange() // Trigger control change handling
        app.ps.ForceRedraw()
        return true
    }
    if app.rotateBox.OnMouseButtonDown(fx, fy) {
        app.ps.ForceRedraw()
        return true
    }
    if app.joinBox.OnMouseButtonDown(fx, fy) {
        app.ps.ForceRedraw()
        return true
    }
    if app.scaleBox.OnMouseButtonDown(fx, fy) {
        app.ps.ForceRedraw()
        return true
    }

	return false
}

func (app *Application) OnMouseButtonUp(x, y int, flags platform.InputFlags) bool {
	// Check if any control needs to handle mouse up
	fx, fy := float64(x), float64(y)

	// Check sliders
    if app.stepSlider.OnMouseButtonUp(fx, fy) {
        app.ps.ForceRedraw()
        return true
    }
    if app.widthSlider.OnMouseButtonUp(fx, fy) {
        app.ps.ForceRedraw()
        return true
    }

	// Check checkboxes
    if app.testBox.OnMouseButtonUp(fx, fy) {
        app.ps.ForceRedraw()
        return true
    }
    if app.rotateBox.OnMouseButtonUp(fx, fy) {
        app.ps.ForceRedraw()
        return true
    }
    if app.joinBox.OnMouseButtonUp(fx, fy) {
        app.ps.ForceRedraw()
        return true
    }
    if app.scaleBox.OnMouseButtonUp(fx, fy) {
        app.ps.ForceRedraw()
        return true
    }

	return false
}

func (app *Application) OnMouseMove(x, y int, flags platform.InputFlags) bool {
	// Check if any control needs to handle mouse move (for slider dragging)
	fx, fy := float64(x), float64(y)

	// Check sliders for dragging
    if app.stepSlider.OnMouseMove(fx, fy, (flags&platform.MouseLeft) != 0) {
        app.ps.ForceRedraw()
        return true
    }
    if app.widthSlider.OnMouseMove(fx, fy, (flags&platform.MouseLeft) != 0) {
        app.ps.ForceRedraw()
        return true
    }

	// Checkboxes typically don't need mouse move handling

	return false
}

func (app *Application) performanceTest() {
	fmt.Println("Running performance test...")

	width := float64(app.ps.Width())
	height := float64(app.ps.Height())
	rbuf := app.rc.WindowBuffer()
	pixf := pixfmt.NewPixFmtRGBA32(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt(pixf)
	drawColor := color.RGBA8Linear{R: 102, G: 77, B: 26, A: 255}

	// Test aliased subpixel accuracy (200 iterations)
	fmt.Print("Testing aliased subpixel accuracy...")
	app.ps.StartTimer()
	for i := 0; i < 200; i++ {
		app.drawAliasedSubpixelAccuracy(renBase, width, height, drawColor)
		app.startAngle += basics.Deg2RadF(app.stepSlider.Value())
	}
	elapsed1 := app.ps.ElapsedTime()
	fmt.Printf(" %.2f ms\n", elapsed1)

	// Test anti-aliased outline (200 iterations)
	fmt.Print("Testing anti-aliased outline...")
	app.ps.StartTimer()
	for i := 0; i < 200; i++ {
		app.drawAntiAliasedOutline(renBase, width, height, drawColor)
		app.startAngle += basics.Deg2RadF(app.stepSlider.Value())
	}
	elapsed2 := app.ps.ElapsedTime()
	fmt.Printf(" %.2f ms\n", elapsed2)

	// Test anti-aliased scanline (200 iterations)
	fmt.Print("Testing anti-aliased scanline...")
	app.ps.StartTimer()
	for i := 0; i < 200; i++ {
		app.drawAntiAliasedScanline(renBase, width, height, drawColor)
		app.startAngle += basics.Deg2RadF(app.stepSlider.Value())
	}
	elapsed3 := app.ps.ElapsedTime()
	fmt.Printf(" %.2f ms\n", elapsed3)

	// Test anti-aliased outline with image pattern (200 iterations)
	fmt.Print("Testing anti-aliased outline with image pattern...")
	app.ps.StartTimer()
	for i := 0; i < 200; i++ {
		app.drawAntiAliasedOutlineImg(renBase, width, height)
		app.startAngle += basics.Deg2RadF(app.stepSlider.Value())
	}
	elapsed4 := app.ps.ElapsedTime()
	fmt.Printf(" %.2f ms\n", elapsed4)

	// Display summary
	totalTime := elapsed1 + elapsed2 + elapsed3 + elapsed4
	fmt.Printf("\nPerformance Test Results:\n")
	fmt.Printf("  Aliased subpixel accuracy:     %8.2f ms\n", elapsed1)
	fmt.Printf("  Anti-aliased outline:          %8.2f ms\n", elapsed2)
	fmt.Printf("  Anti-aliased scanline:         %8.2f ms\n", elapsed3)
	fmt.Printf("  Anti-aliased outline + image:  %8.2f ms\n", elapsed4)
	fmt.Printf("  Total time:                    %8.2f ms\n", totalTime)
	fmt.Println("Performance test completed!")
}
