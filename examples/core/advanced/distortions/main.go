// Port of AGG C++ distortions.cpp demo.
//
// Renders an image through wave/swirl distortion filters applied via a
// SpanInterpolatorAdaptor. Three rendering passes match the C++ original:
//  1. Distorted image fill inside a transformed ellipse.
//  2. Solid black outline of the same ellipse, shifted right.
//  3. Gradient circle with the same distortion applied.
//
// Controls are rendered to match the original layout:
//   - Angle and Scale sliders (top-left)
//   - Period and Amplitude sliders (top-center)
//   - Distortion type radio-button group (top-right)
//
// The image is rendered in a flipped work buffer and copied with y-flip,
// matching the C++ original's flip_y=true coordinate system.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	icol "github.com/MeKo-Christian/agg_go/internal/color"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	rboxctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/rbox"
	sliderctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	"github.com/MeKo-Christian/agg_go/internal/image"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
	"github.com/MeKo-Christian/agg_go/internal/span"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

// Canvas dimensions matching C++ agg_main:
//
//	init(rbuf_img(0).width() + 300, rbuf_img(0).height() + 40 + 20)
//
// The "spheres.bmp" image is 320×300, so window = 620×360.
const (
	srcImgW = 320
	srcImgH = 300
	canvasW = srcImgW + 300 // 620
	canvasH = srcImgH + 60  // 360
)

// g_gradient_colors is the color table from the original C++ demo (RGBA, 256 entries).
//
//nolint:gochecknoglobals // package-level var is intentional: matches C++ static array
var g_gradient_colors = [256 * 4]uint8{
	255, 255, 255, 255, 255, 255, 254, 255, 255, 255, 254, 255, 255, 255, 254, 255,
	255, 255, 253, 255, 255, 255, 253, 255, 255, 255, 252, 255, 255, 255, 251, 255,
	255, 255, 250, 255, 255, 255, 248, 255, 255, 255, 246, 255, 255, 255, 244, 255,
	255, 255, 241, 255, 255, 255, 238, 255, 255, 255, 235, 255, 255, 255, 231, 255,
	255, 255, 227, 255, 255, 255, 222, 255, 255, 255, 217, 255, 255, 255, 211, 255,
	255, 255, 206, 255, 255, 255, 200, 255, 255, 254, 194, 255, 255, 253, 188, 255,
	255, 252, 182, 255, 255, 250, 176, 255, 255, 249, 170, 255, 255, 247, 164, 255,
	255, 246, 158, 255, 255, 244, 152, 255, 254, 242, 146, 255, 254, 240, 141, 255,
	254, 238, 136, 255, 254, 236, 131, 255, 253, 234, 126, 255, 253, 232, 121, 255,
	253, 229, 116, 255, 252, 227, 112, 255, 252, 224, 108, 255, 251, 222, 104, 255,
	251, 219, 100, 255, 251, 216, 96, 255, 250, 214, 93, 255, 250, 211, 89, 255,
	249, 208, 86, 255, 249, 205, 83, 255, 248, 202, 80, 255, 247, 199, 77, 255,
	247, 196, 74, 255, 246, 193, 72, 255, 246, 190, 69, 255, 245, 187, 67, 255,
	244, 183, 64, 255, 244, 180, 62, 255, 243, 177, 60, 255, 242, 174, 58, 255,
	242, 170, 56, 255, 241, 167, 54, 255, 240, 164, 52, 255, 239, 161, 51, 255,
	239, 157, 49, 255, 238, 154, 47, 255, 237, 151, 46, 255, 236, 147, 44, 255,
	235, 144, 43, 255, 235, 141, 41, 255, 234, 138, 40, 255, 233, 134, 39, 255,
	232, 131, 37, 255, 231, 128, 36, 255, 230, 125, 35, 255, 229, 122, 34, 255,
	228, 119, 33, 255, 227, 116, 31, 255, 226, 113, 30, 255, 225, 110, 29, 255,
	224, 107, 28, 255, 223, 104, 27, 255, 222, 101, 26, 255, 221, 99, 25, 255,
	220, 96, 24, 255, 219, 93, 23, 255, 218, 91, 22, 255, 217, 88, 21, 255,
	216, 86, 20, 255, 215, 83, 19, 255, 214, 81, 18, 255, 213, 79, 17, 255,
	212, 77, 17, 255, 211, 74, 16, 255, 210, 72, 15, 255, 209, 70, 14, 255,
	207, 68, 13, 255, 206, 66, 13, 255, 205, 64, 12, 255, 204, 62, 11, 255,
	203, 60, 10, 255, 202, 58, 10, 255, 201, 56, 9, 255, 199, 55, 9, 255,
	198, 53, 8, 255, 197, 51, 7, 255, 196, 50, 7, 255, 195, 48, 6, 255,
	193, 46, 6, 255, 192, 45, 5, 255, 191, 43, 5, 255, 190, 42, 4, 255,
	188, 41, 4, 255, 187, 39, 3, 255, 186, 38, 3, 255, 185, 37, 2, 255,
	183, 35, 2, 255, 182, 34, 1, 255, 181, 33, 1, 255, 179, 32, 1, 255,
	178, 30, 0, 255, 177, 29, 0, 255, 175, 28, 0, 255, 174, 27, 0, 255,
	173, 26, 0, 255, 171, 25, 0, 255, 170, 24, 0, 255, 168, 23, 0, 255,
	167, 22, 0, 255, 165, 21, 0, 255, 164, 21, 0, 255, 163, 20, 0, 255,
	161, 19, 0, 255, 160, 18, 0, 255, 158, 17, 0, 255, 156, 17, 0, 255,
	155, 16, 0, 255, 153, 15, 0, 255, 152, 14, 0, 255, 150, 14, 0, 255,
	149, 13, 0, 255, 147, 12, 0, 255, 145, 12, 0, 255, 144, 11, 0, 255,
	142, 11, 0, 255, 140, 10, 0, 255, 139, 10, 0, 255, 137, 9, 0, 255,
	135, 9, 0, 255, 134, 8, 0, 255, 132, 8, 0, 255, 130, 7, 0, 255,
	128, 7, 0, 255, 126, 6, 0, 255, 125, 6, 0, 255, 123, 5, 0, 255,
	121, 5, 0, 255, 119, 4, 0, 255, 117, 4, 0, 255, 115, 4, 0, 255,
	113, 3, 0, 255, 111, 3, 0, 255, 109, 2, 0, 255, 107, 2, 0, 255,
	105, 2, 0, 255, 103, 1, 0, 255, 101, 1, 0, 255, 99, 1, 0, 255,
	97, 0, 0, 255, 95, 0, 0, 255, 93, 0, 0, 255, 91, 0, 0, 255,
	90, 0, 0, 255, 88, 0, 0, 255, 86, 0, 0, 255, 84, 0, 0, 255,
	82, 0, 0, 255, 80, 0, 0, 255, 78, 0, 0, 255, 77, 0, 0, 255,
	75, 0, 0, 255, 73, 0, 0, 255, 72, 0, 0, 255, 70, 0, 0, 255,
	68, 0, 0, 255, 67, 0, 0, 255, 65, 0, 0, 255, 64, 0, 0, 255,
	63, 0, 0, 255, 61, 0, 0, 255, 60, 0, 0, 255, 59, 0, 0, 255,
	58, 0, 0, 255, 57, 0, 0, 255, 56, 0, 0, 255, 55, 0, 0, 255,
	54, 0, 0, 255, 53, 0, 0, 255, 53, 0, 0, 255, 52, 0, 0, 255,
	52, 0, 0, 255, 51, 0, 0, 255, 51, 0, 0, 255, 51, 0, 0, 255,
	50, 0, 0, 255, 50, 0, 0, 255, 51, 0, 0, 255, 51, 0, 0, 255,
	51, 0, 0, 255, 51, 0, 0, 255, 52, 0, 0, 255, 52, 0, 0, 255,
	53, 0, 0, 255, 54, 1, 0, 255, 55, 2, 0, 255, 56, 3, 0, 255,
	57, 4, 0, 255, 58, 5, 0, 255, 59, 6, 0, 255, 60, 7, 0, 255,
	62, 8, 0, 255, 63, 9, 0, 255, 64, 11, 0, 255, 66, 12, 0, 255,
	68, 13, 0, 255, 69, 14, 0, 255, 71, 16, 0, 255, 73, 17, 0, 255,
	75, 18, 0, 255, 77, 20, 0, 255, 79, 21, 0, 255, 81, 23, 0, 255,
	83, 24, 0, 255, 85, 26, 0, 255, 87, 28, 0, 255, 90, 29, 0, 255,
	92, 31, 0, 255, 94, 33, 0, 255, 97, 34, 0, 255, 99, 36, 0, 255,
	102, 38, 0, 255, 104, 40, 0, 255, 107, 41, 0, 255, 109, 43, 0, 255,
	112, 45, 0, 255, 115, 47, 0, 255, 117, 49, 0, 255, 120, 51, 0, 255,
	123, 52, 0, 255, 126, 54, 0, 255, 128, 56, 0, 255, 131, 58, 0, 255,
	134, 60, 0, 255, 137, 62, 0, 255, 140, 64, 0, 255, 143, 66, 0, 255,
	145, 68, 0, 255, 148, 70, 0, 255, 151, 72, 0, 255, 154, 74, 0, 255,
}

// --- Distortion implementations ---

type distortionBase struct {
	cx, cy    float64
	period    float64
	amplitude float64 // stored as 1/userAmplitude (matches C++ periodic_distortion::amplitude)
	phase     float64
}

// calculateWave applies the wave distortion in-place (integer subpixel coords).
// Matches C++ calculate_wave free function.
func calculateWave(x, y *int, cx, cy, period, amplitude, phase float64) {
	xd := float64(*x)/float64(basics.PolySubpixelScale) - cx
	yd := float64(*y)/float64(basics.PolySubpixelScale) - cy
	d := math.Sqrt(xd*xd + yd*yd)
	if d > 1 {
		a := math.Cos(d/(16.0*period)-phase)*(1.0/(amplitude*d)) + 1.0
		*x = int((xd*a + cx) * float64(basics.PolySubpixelScale))
		*y = int((yd*a + cy) * float64(basics.PolySubpixelScale))
	}
}

// calculateSwirl applies the swirl distortion in-place.
// Matches C++ calculate_swirl free function.
func calculateSwirl(x, y *int, cx, cy, amplitude, phase float64) {
	xd := float64(*x)/float64(basics.PolySubpixelScale) - cx
	yd := float64(*y)/float64(basics.PolySubpixelScale) - cy
	a := (100.0 - math.Sqrt(xd*xd+yd*yd)) / 100.0 * (0.1 / -amplitude)
	sa := math.Sin(a - phase/25.0)
	ca := math.Cos(a - phase/25.0)
	*x = int((xd*ca - yd*sa + cx) * float64(basics.PolySubpixelScale))
	*y = int((xd*sa + yd*ca + cy) * float64(basics.PolySubpixelScale))
}

type distortionWave struct{ distortionBase }

func (d *distortionWave) Calculate(x, y *int) {
	calculateWave(x, y, d.cx, d.cy, d.period, d.amplitude, d.phase)
}

type distortionSwirl struct{ distortionBase }

func (d *distortionSwirl) Calculate(x, y *int) {
	calculateSwirl(x, y, d.cx, d.cy, d.amplitude, d.phase)
}

type distortionWaveSwirl struct{ distortionBase }

func (d *distortionWaveSwirl) Calculate(x, y *int) {
	calculateWave(x, y, d.cx, d.cy, d.period, d.amplitude, d.phase)
	calculateSwirl(x, y, d.cx, d.cy, d.amplitude, d.phase)
}

type distortionSwirlWave struct{ distortionBase }

func (d *distortionSwirlWave) Calculate(x, y *int) {
	calculateSwirl(x, y, d.cx, d.cy, d.amplitude, d.phase)
	calculateWave(x, y, d.cx, d.cy, d.period, d.amplitude, d.phase)
}

// makeDistortion constructs the distortion matching the rbox selection.
func makeDistortion(curItem int, db distortionBase) span.Distortion {
	switch curItem {
	case 1:
		return &distortionSwirl{db}
	case 2:
		return &distortionWaveSwirl{db}
	case 3:
		return &distortionSwirlWave{db}
	default:
		return &distortionWave{db}
	}
}

// --- Image source adapter for SpanImageFilterRGBABilinearClip ---

// imagePixFmt adapts a RenderingBufferU8 to the accessor pixel-format interface.
type imagePixFmt struct {
	rbuf *buffer.RenderingBufferU8
}

func (p imagePixFmt) Width() int    { return p.rbuf.Width() }
func (p imagePixFmt) Height() int   { return p.rbuf.Height() }
func (p imagePixFmt) PixWidth() int { return 4 }
func (p imagePixFmt) PixPtr(x, y int) []basics.Int8u {
	row := buffer.RowU8(p.rbuf, y)
	return row[x*4:]
}

// distortionsSource implements the RGBASourceInterface for the span generator.
type distortionsSource struct {
	accessor *image.ImageAccessorClip[imagePixFmt]
	ipf      *imagePixFmt
}

func (s *distortionsSource) Width() int                 { return s.ipf.Width() }
func (s *distortionsSource) Height() int                { return s.ipf.Height() }
func (s *distortionsSource) ColorType() string          { return "RGBA8" }
func (s *distortionsSource) OrderType() icol.ColorOrder { return icol.OrderRGBA }
func (s *distortionsSource) Span(x, y, length int) []basics.Int8u {
	return s.accessor.Span(x, y, length)
}
func (s *distortionsSource) NextX() []basics.Int8u { return s.accessor.NextX() }
func (s *distortionsSource) NextY() []basics.Int8u { return s.accessor.NextY() }
func (s *distortionsSource) RowPtr(y int) []basics.Int8u {
	return s.ipf.PixPtr(0, y)
}

// gradientColorFunc implements span.ColorFunction for the gradient LUT.
type gradientColorFunc struct {
	colors [256]icol.RGBA8[icol.Linear]
}

func (g *gradientColorFunc) Size() int { return 256 }
func (g *gradientColorFunc) ColorAt(index int) icol.RGBA8[icol.Linear] {
	return g.colors[index]
}

func buildGradientColors() *gradientColorFunc {
	f := &gradientColorFunc{}
	for i := range 256 {
		f.colors[i] = icol.RGBA8[icol.Linear]{
			R: g_gradient_colors[i*4+0],
			G: g_gradient_colors[i*4+1],
			B: g_gradient_colors[i*4+2],
			A: g_gradient_colors[i*4+3],
		}
	}
	return f
}

// --- Vertex source adapters ---

// transformedEllipseAdapter applies an affine transform to an ellipse and satisfies
// the rasterizer VertexSource interface (Rewind(uint32), Vertex(*float64,*float64) uint32).
type transformedEllipseAdapter struct {
	e   *shapes.Ellipse
	mtx *transform.TransAffine
}

func (a *transformedEllipseAdapter) Rewind(pathID uint32) { a.e.Rewind(pathID) }
func (a *transformedEllipseAdapter) Vertex(x, y *float64) uint32 {
	cmd := a.e.Vertex(x, y)
	if basics.IsVertex(basics.PathCommand(cmd)) {
		a.mtx.Transform(x, y)
	}
	return uint32(cmd)
}

// simpleVertexSource is the interface used by ctrl widgets.
type simpleVertexSource interface {
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
}

// ctrlVertexSourceAdapter adapts ctrl.Ctrl's Vertex() (uint-based) to the
// rasterizer AddPath interface (uint32-based).
type ctrlVertexSourceAdapter struct {
	src simpleVertexSource
}

func (a *ctrlVertexSourceAdapter) Rewind(pathID uint32) { a.src.Rewind(uint(pathID)) }
func (a *ctrlVertexSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

// --- Low-level rendering helpers ---

// rasScanlineAdapter adapts scanline.ScanlineU8 to rasterizer.ScanlineInterface.
// spanIter implements renscan.ScanlineIterator over scanline.Span slices.
// scanlineWrapper adapts scanline.ScanlineU8 to renscan.ScanlineInterface.
// rasAdapter wraps the rasterizer to satisfy renscan.RasterizerInterface.
// renderSolid renders the rasterizer using a solid color via renscan.RenderScanlinesAASolid.
func renderSolid(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	slw *scanline.ScanlineP8,
	rb *renderer.RendererBase[*pixfmt.PixFmtRGBA32Pre[icol.Linear], icol.RGBA8[icol.Linear]],
	color icol.RGBA8[icol.Linear],
) {
	renscan.RenderScanlinesAASolid(ras, slw, rb, color)
}

// renderImageSpan renders the image-filter pass via a manual loop (avoids interface wrapping
// issues with the imageSpanGen's concrete Generate signature).
func renderImageSpan(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	sl *scanline.ScanlineU8,
	rb *renderer.RendererBase[*pixfmt.PixFmtRGBA32Pre[icol.Linear], icol.RGBA8[icol.Linear]],
	alloc *span.SpanAllocator[icol.RGBA8[icol.Linear]],
	sg *span.SpanImageFilterRGBABilinearClip[*distortionsSource, *span.SpanInterpolatorAdaptor[*span.SpanInterpolatorLinear[*transform.TransAffine], span.Distortion]],
) {
	if !ras.RewindScanlines() {
		return
	}
	sl.Reset(ras.MinX(), ras.MaxX())
	for ras.SweepScanline(sl) {
		y := sl.Y()
		for _, sp := range sl.Spans() {
			if sp.Len > 0 {
				colors := alloc.Allocate(int(sp.Len))
				sg.Generate(colors[:int(sp.Len)], int(sp.X), y)
				rb.BlendColorHspan(int(sp.X), y, int(sp.Len), colors, sp.Covers, basics.CoverFull)
			}
		}
	}
}

// copyFlipY copies src to dst with vertical flip (y=0 at bottom → y=0 at top).
func copyFlipY(src, dst []uint8, w, h int) {
	stride := w * 4
	for y := range h {
		srcOff := (h - 1 - y) * stride
		dstOff := y * stride
		copy(dst[dstOff:dstOff+stride], src[srcOff:srcOff+stride])
	}
}

// --- Source image ---

// createTestImage builds a procedural source image to substitute for "spheres.bmp".
func createTestImage(w, h int) *agg.Image {
	img := agg.CreateImage(w, h)
	imgCtx := agg.NewContextForImage(img)
	imgCtx.Clear(agg.White)

	imgCtx.SetColor(agg.RGBA(0.8, 0.8, 0.8, 1.0))
	for i := 0; i < w; i += 20 {
		imgCtx.DrawLine(float64(i), 0, float64(i), float64(h))
	}
	for i := 0; i < h; i += 20 {
		imgCtx.DrawLine(0, float64(i), float64(w), float64(i))
	}
	imgCtx.SetColor(agg.Red)
	imgCtx.FillCircle(float64(w)/2, float64(h)/2, float64(w)/4)
	imgCtx.SetColor(agg.Blue)
	imgCtx.SetStrokeWidth(5.0)
	imgCtx.DrawRectangle(10, 10, float64(w-20), float64(h-20))
	imgCtx.SetColor(agg.Black)
	imgCtx.SetStrokeWidth(1.0)
	for i := -w; i < w; i += 4 {
		imgCtx.DrawLine(float64(i), 0, float64(i+w), float64(h))
	}
	return img
}

// --- Controls ---

// toRGBA8 converts an RGBA float color to RGBA8, clamping to [0, 255].
func toRGBA8(c icol.RGBA) icol.RGBA8[icol.Linear] {
	clamp := func(v float64) uint8 {
		switch {
		case v <= 0:
			return 0
		case v >= 1:
			return 255
		default:
			return uint8(v*255.0 + 0.5)
		}
	}
	return icol.RGBA8[icol.Linear]{
		R: clamp(c.R),
		G: clamp(c.G),
		B: clamp(c.B),
		A: clamp(c.A),
	}
}

// renderCtrl renders all paths of a control widget using low-level rendering.
func renderCtrl(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	slw *scanline.ScanlineP8,
	rb *renderer.RendererBase[*pixfmt.PixFmtRGBA32Pre[icol.Linear], icol.RGBA8[icol.Linear]],
	c ctrlbase.Ctrl[icol.RGBA],
) {
	for i := uint(0); i < c.NumPaths(); i++ {
		ras.Reset()
		ras.AddPath(&ctrlVertexSourceAdapter{src: c}, uint32(i))
		renderSolid(ras, slw, rb, toRGBA8(c.Color(i)))
	}
}

// --- Demo ---

type demo struct{}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()

	// Work buffer: y=0 at bottom (flip_y=true convention), copied with y-flip.
	workBuf := make([]uint8, w*h*4)

	// Fill work buffer white (C++: rb.clear(rgba(1,1,1))).
	for i := 0; i < len(workBuf); i += 4 {
		workBuf[i] = 255
		workBuf[i+1] = 255
		workBuf[i+2] = 255
		workBuf[i+3] = 255
	}

	// Default control values matching the C++ original.
	const (
		angleVal     = 20.0 // m_angle default
		scaleVal     = 1.0  // m_scale default
		amplitudeVal = 10.0 // m_amplitude default
		periodVal    = 1.0  // m_period default
		phase        = 0.0
		distType     = 0 // 0=Wave, 1=Swirl, 2=Wave-Swirl, 3=Swirl-Wave
	)

	// Source image dimensions match the C++ "spheres.bmp" default.
	imgW := float64(srcImgW)
	imgH := float64(srcImgH)

	// C++ on_init: m_center_x = rbuf_img(0).width()/2 + 10
	//              m_center_y = rbuf_img(0).height()/2 + 10 + 40
	centerX := imgW/2 + 10
	centerY := imgH/2 + 10 + 40

	// --- Low-level canvas objects ---
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(workBuf, w, h, w*4)
	pixFmt := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	rb := renderer.NewRendererBaseWithPixfmt(pixFmt)
	alloc := span.NewSpanAllocator[icol.RGBA8[icol.Linear]]()
	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()
	slw := scanline.NewScanlineP8()

	// --- Controls (C++ constructor positions, !flip_y = false) ---
	//   m_angle     (5,      5,    150,     12,    !flip_y)
	//   m_scale     (5,      5+15, 150,     12+15, !flip_y)
	//   m_period    (5+170,  5,    150+170, 12,    !flip_y)
	//   m_amplitude (5+170,  5+15, 150+170, 12+15, !flip_y)
	//   m_distortion(480,    5,    600,     90,    !flip_y)
	ctrlAngle := sliderctrl.NewSliderCtrl(5, 5, 150, 12, false)
	ctrlAngle.SetRange(-180.0, 180.0)
	ctrlAngle.SetValue(angleVal)
	ctrlAngle.SetLabel("Angle=%3.2f")

	ctrlScale := sliderctrl.NewSliderCtrl(5, 5+15, 150, 12+15, false)
	ctrlScale.SetRange(0.1, 5.0)
	ctrlScale.SetValue(scaleVal)
	ctrlScale.SetLabel("Scale=%3.2f")

	ctrlPeriod := sliderctrl.NewSliderCtrl(5+170, 5, 150+170, 12, false)
	ctrlPeriod.SetRange(0.1, 2.0)
	ctrlPeriod.SetValue(periodVal)
	ctrlPeriod.SetLabel("Period=%3.2f")

	ctrlAmplitude := sliderctrl.NewSliderCtrl(5+170, 5+15, 150+170, 12+15, false)
	ctrlAmplitude.SetRange(0.1, 40.0)
	ctrlAmplitude.SetValue(amplitudeVal)
	ctrlAmplitude.SetLabel("Amplitude=%3.2f")

	ctrlDistortion := rboxctrl.NewDefaultRboxCtrl(480, 5, 600, 90, false)
	ctrlDistortion.AddItem("Wave")
	ctrlDistortion.AddItem("Swirl")
	ctrlDistortion.AddItem("Wave-Swirl")
	ctrlDistortion.AddItem("Swirl-Wave")
	ctrlDistortion.SetCurItem(distType)

	// --- Build affine matrices (matching C++ on_draw) ---
	//
	// src_mtx: ellipse clip path (rotation only, no scale).
	srcMtx := transform.NewTransAffine()
	srcMtx.Multiply(transform.NewTransAffineTranslation(-imgW/2, -imgH/2))
	srcMtx.Multiply(transform.NewTransAffineRotation(angleVal * math.Pi / 180.0))
	srcMtx.Multiply(transform.NewTransAffineTranslation(imgW/2+10, imgH/2+10+40))

	// img_mtx: image interpolation (rotation + scale, then inverted).
	imgMtx := transform.NewTransAffine()
	imgMtx.Multiply(transform.NewTransAffineTranslation(-imgW/2, -imgH/2))
	imgMtx.Multiply(transform.NewTransAffineRotation(angleVal * math.Pi / 180.0))
	imgMtx.Multiply(transform.NewTransAffineScaling(scaleVal))
	imgMtx.Multiply(transform.NewTransAffineTranslation(imgW/2+10, imgH/2+10+40))
	imgMtx.Invert()

	// Map distortion center through img_mtx.
	cx, cy := centerX, centerY
	imgMtx.Transform(&cx, &cy)

	db := distortionBase{
		period:    periodVal,
		amplitude: 1.0 / amplitudeVal,
		phase:     phase,
		cx:        cx,
		cy:        cy,
	}
	dist := makeDistortion(distType, db)

	// --- Build image span generator ---
	srcImg := createTestImage(srcImgW, srcImgH)
	imgRbuf := buffer.NewRenderingBufferU8()
	imgRbuf.Attach(srcImg.Data, srcImg.Width(), srcImg.Height(), srcImg.Width()*4)
	ipf := &imagePixFmt{rbuf: imgRbuf}
	accessor := image.NewImageAccessorClip(ipf, []basics.Int8u{255, 255, 255, 255})
	source := &distortionsSource{accessor: accessor, ipf: ipf}

	li := span.NewSpanInterpolatorLinear(imgMtx, 8)
	interpolator := span.NewSpanInterpolatorAdaptor(li, dist)
	sg := span.NewSpanImageFilterRGBABilinearClipWithParams(source, icol.RGBA8[icol.Linear]{R: 255, G: 255, B: 255, A: 255}, interpolator)

	// --- Shared ellipse (C++: r=min(imgW,imgH), radius=r/2-20, 200 steps) ---
	r := imgW
	if imgH < r {
		r = imgH
	}
	ell := shapes.NewEllipseWithParams(imgW/2, imgH/2, r/2-20, r/2-20, 200, false)

	// --- Pass 1: Distorted image fill ---
	// C++: conv_transform<ellipse> tr(ell, src_mtx)
	//      ras.add_path(tr)
	//      render_scanlines_aa(ras, sl, rb, sa, sg)
	ras.Reset()
	ras.AddPath(&transformedEllipseAdapter{e: ell, mtx: srcMtx}, 0)
	renderImageSpan(ras, sl, rb, alloc, sg)

	// --- Pass 2: Solid black outline (ellipse shifted right) ---
	// C++: src_mtx *= trans_affine_translation(img_width - img_width/10, 0)
	//      ras.add_path(tr)   // tr still references src_mtx (now modified)
	//      render_scanlines_aa_solid(ras, sl, rb, srgba8(0,0,0))
	srcMtx2 := srcMtx
	srcMtx2.Multiply(transform.NewTransAffineTranslation(imgW-imgW/10, 0))
	ras.Reset()
	ell.Rewind(0)
	ras.AddPath(&transformedEllipseAdapter{e: ell, mtx: srcMtx2}, 0)
	renderSolid(ras, slw, rb, icol.RGBA8[icol.Linear]{R: 0, G: 0, B: 0, A: 255})

	// --- Pass 3: Gradient with distortion ---
	//
	// gr1_mtx: shapes the ellipse for the gradient pass.
	// gr2_mtx: inverted interpolation matrix for the gradient.
	// Distortion center is shifted right by (imgW - imgW/10).
	gr1Mtx := transform.NewTransAffine()
	gr1Mtx.Multiply(transform.NewTransAffineTranslation(-imgW/2, -imgH/2))
	gr1Mtx.Multiply(transform.NewTransAffineScaling(0.8))
	gr1Mtx.Multiply(transform.NewTransAffineRotation(angleVal * math.Pi / 180.0))
	gr1Mtx.Multiply(transform.NewTransAffineTranslation(imgW-imgW/10+imgW/2+10, imgH/2+10+40))

	gr2Mtx := transform.NewTransAffine()
	gr2Mtx.Multiply(transform.NewTransAffineRotation(angleVal * math.Pi / 180.0))
	gr2Mtx.Multiply(transform.NewTransAffineScaling(scaleVal))
	gr2Mtx.Multiply(transform.NewTransAffineTranslation(imgW-imgW/10+imgW/2+10+50, imgH/2+10+40+50))
	gr2Mtx.Invert()

	cx2, cy2 := centerX+imgW-imgW/10, centerY
	gr2Mtx.Transform(&cx2, &cy2)

	db2 := distortionBase{
		period:    periodVal,
		amplitude: 1.0 / amplitudeVal,
		phase:     phase,
		cx:        cx2,
		cy:        cy2,
	}

	// C++: interpolator.transformer(gr2_mtx); dist->center(cx, cy);
	// Update the base interpolator's matrix and distortion to match C++.
	li2 := span.NewSpanInterpolatorLinear(gr2Mtx, 8)
	interpolator.SetBase(li2)
	dist2 := makeDistortion(distType, db2)
	interpolator.SetDistortion(dist2)

	gradColors := buildGradientColors()
	gradSpan := span.NewSpanGradient(interpolator, span.GradientRadial{}, gradColors, 0, 180)

	ras.Reset()
	ell.Rewind(0)
	ras.AddPath(&transformedEllipseAdapter{e: ell, mtx: gr1Mtx}, 0)
	renscan.RenderScanlinesAA(ras, slw, rb, alloc, gradSpan)

	// --- Render controls (C++ render_ctrl calls at end of on_draw) ---
	// C++: render_ctrl(ras, sl, rb, m_angle); etc.
	renderCtrl(ras, slw, rb, ctrlAngle)
	renderCtrl(ras, slw, rb, ctrlScale)
	renderCtrl(ras, slw, rb, ctrlAmplitude)
	renderCtrl(ras, slw, rb, ctrlPeriod)
	renderCtrl(ras, slw, rb, ctrlDistortion)

	copyFlipY(workBuf, img.Data, w, h)
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Image and Gradient Distortions",
		Width:  canvasW,
		Height: canvasH,
	}, &demo{})
}
