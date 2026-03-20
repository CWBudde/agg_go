// Port of AGG C++ distortions.cpp demo.
//
// Renders a test image through a wave distortion filter using the
// SpanInterpolatorAdaptor with a custom Distortion implementation.
// The default state uses the "wave" distortion type at phase=0.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/image"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/span"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

// --- Distortion implementations ---

type distortionBase struct {
	cx, cy    float64
	period    float64
	amplitude float64
	phase     float64
}

type distortionWave struct {
	distortionBase
}

func (d *distortionWave) Calculate(x, y *int) {
	xd := float64(*x)/float64(basics.PolySubpixelScale) - d.cx
	yd := float64(*y)/float64(basics.PolySubpixelScale) - d.cy
	dist := math.Sqrt(xd*xd + yd*yd)
	if dist > 1 {
		a := math.Cos(dist/(16.0*d.period)-d.phase)*(1.0/(d.amplitude*dist)) + 1.0
		*x = int((xd*a + d.cx) * float64(basics.PolySubpixelScale))
		*y = int((yd*a + d.cy) * float64(basics.PolySubpixelScale))
	}
}

// imagePixFmt adapts a RenderingBufferU8 to the pixel-format interface
// needed by ImageAccessorClip.
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

func (s *distortionsSource) Width() int                  { return s.ipf.Width() }
func (s *distortionsSource) Height() int                 { return s.ipf.Height() }
func (s *distortionsSource) ColorType() string           { return "RGBA8" }
func (s *distortionsSource) OrderType() color.ColorOrder { return color.OrderRGBA }

func (s *distortionsSource) Span(x, y, length int) []basics.Int8u {
	return s.accessor.Span(x, y, length)
}
func (s *distortionsSource) NextX() []basics.Int8u { return s.accessor.NextX() }
func (s *distortionsSource) NextY() []basics.Int8u { return s.accessor.NextY() }
func (s *distortionsSource) RowPtr(y int) []basics.Int8u {
	return s.ipf.PixPtr(0, y)
}

// spanGeneratorAdapter bridges the span generator's signature to what the
// manual rendering loop expects.
type spanGeneratorAdapter struct {
	sg *span.SpanImageFilterRGBABilinearClip[*distortionsSource, *span.SpanInterpolatorAdaptor[*span.SpanInterpolatorLinear[*transform.TransAffine], span.Distortion]]
}

func (a *spanGeneratorAdapter) Prepare() {}
func (a *spanGeneratorAdapter) Generate(colors []color.RGBA8[color.Linear], x, y, length int) {
	if length > len(colors) {
		length = len(colors)
	}
	a.sg.Generate(colors[:length], x, y)
}

// pathSourceAdapter adapts PathStorageStl to the rasterizer VertexSource interface.
type pathSourceAdapter struct {
	ps *path.PathStorageStl
}

func (a *pathSourceAdapter) Rewind(pathID uint32) { a.ps.Rewind(uint(pathID)) }
func (a *pathSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ps.NextVertex()
	*x = vx
	*y = vy
	return cmd
}

// rasScanlineAdapter adapts ScanlineU8 to the rasterizer's ScanlineInterface.
type rasScanlineAdapter struct {
	sl *scanline.ScanlineU8
}

func (a *rasScanlineAdapter) ResetSpans()                 { a.sl.ResetSpans() }
func (a *rasScanlineAdapter) AddCell(x int, cover uint32) { a.sl.AddCell(x, uint(cover)) }
func (a *rasScanlineAdapter) AddSpan(x, length int, cover uint32) {
	a.sl.AddSpan(x, length, uint(cover))
}
func (a *rasScanlineAdapter) Finalize(y int) { a.sl.Finalize(y) }
func (a *rasScanlineAdapter) NumSpans() int  { return a.sl.NumSpans() }

// createTestImage builds a small procedural image (grid + shapes) used as the
// distortion source.
func createTestImage(w, h int) *agg.Image {
	img := agg.CreateImage(w, h)
	imgCtx := agg.NewContextForImage(img)
	imgCtx.Clear(agg.White)

	// Grid
	imgCtx.SetColor(agg.RGBA(0.8, 0.8, 0.8, 1.0))
	for i := 0; i < w; i += 20 {
		imgCtx.DrawLine(float64(i), 0, float64(i), float64(h))
	}
	for i := 0; i < h; i += 20 {
		imgCtx.DrawLine(0, float64(i), float64(w), float64(i))
	}
	// Red circle
	imgCtx.SetColor(agg.Red)
	imgCtx.FillCircle(float64(w)/2, float64(h)/2, float64(w)/4)
	// Blue border
	imgCtx.SetColor(agg.Blue)
	imgCtx.SetStrokeWidth(5.0)
	imgCtx.DrawRectangle(10, 10, float64(w-20), float64(h-20))
	// Diagonal lines (high-frequency pattern)
	imgCtx.SetColor(agg.Black)
	imgCtx.SetStrokeWidth(1.0)
	for i := -w; i < w; i += 4 {
		imgCtx.DrawLine(float64(i), 0, float64(i+w), float64(h))
	}
	return img
}

type demo struct{}

func (d *demo) Render(img *agg.Image) {
	const (
		canvasW, canvasH = 620, 360
		// Default control values matching WASM demo defaults
		centerX   = 350.0
		centerY   = 265.0
		phase     = 0.0
		angle     = 20.0
		scale     = 1.0
		amplitude = 10.0
		period    = 1.0
	)

	// Clear to white.
	for i := 0; i < len(img.Data); i += 4 {
		img.Data[i] = 255
		img.Data[i+1] = 255
		img.Data[i+2] = 255
		img.Data[i+3] = 255
	}

	testImage := createTestImage(canvasW/2, canvasH/2)
	imgW, imgH := float64(testImage.Width()), float64(testImage.Height())

	// Attach canvas
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Width()*4)

	pixFmt := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]](pixFmt)
	alloc := span.NewSpanAllocator[color.RGBA8[color.Linear]]()
	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()
	ps := path.NewPathStorageStl()

	// Build image-to-canvas affine matrix
	imgMtx := transform.NewTransAffine()
	imgMtx.Translate(-imgW/2, -imgH/2)
	imgMtx.Rotate(angle * math.Pi / 180.0)
	imgMtx.Scale(scale)
	imgMtx.Translate(imgW/2+10, imgH/2+50)
	imgMtx.Invert()

	// Map center point through matrix for distortion origin
	cx, cy := centerX, centerY
	imgMtx.Transform(&cx, &cy)

	dist := &distortionWave{distortionBase{
		cx:        cx,
		cy:        cy,
		period:    period,
		amplitude: 1.0 / amplitude,
		phase:     phase,
	}}

	// Interpolator with distortion adaptor
	li := span.NewSpanInterpolatorLinear[*transform.TransAffine](imgMtx, 8)
	interpolator := span.NewSpanInterpolatorAdaptor[*span.SpanInterpolatorLinear[*transform.TransAffine], span.Distortion](li, dist)

	// Source image buffer
	imgRbuf := buffer.NewRenderingBufferU8()
	imgRbuf.Attach(testImage.Data, testImage.Width(), testImage.Height(), testImage.Width()*4)
	ipf := imagePixFmt{rbuf: imgRbuf}

	accessor := image.NewImageAccessorClip(&ipf, []basics.Int8u{255, 255, 255, 255})
	source := &distortionsSource{accessor: accessor, ipf: &ipf}

	sg := span.NewSpanImageFilterRGBABilinearClipWithParams(source, color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255}, interpolator)
	adapterSG := &spanGeneratorAdapter{sg: sg}

	// Build a circular clip path
	r := imgW
	if imgH < r {
		r = imgH
	}
	ps.RemoveAll()
	const numPoints = 100
	for i := 0; i < numPoints; i++ {
		a := 2.0 * math.Pi * float64(i) / float64(numPoints)
		x := imgW/2 + (r/2-20)*math.Cos(a)
		y := imgH/2 + (r/2-20)*math.Sin(a)
		if i == 0 {
			ps.MoveTo(x, y)
		} else {
			ps.LineTo(x, y)
		}
	}
	ps.ClosePolygon(basics.PathFlagsNone)

	ras.Reset()
	ras.AddPath(&pathSourceAdapter{ps: ps}, 0)

	if ras.RewindScanlines() {
		sl.Reset(ras.MinX(), ras.MaxX())
		for ras.SweepScanline(&rasScanlineAdapter{sl: sl}) {
			y := sl.Y()
			for _, spanData := range sl.Spans() {
				if spanData.Len > 0 {
					colors := alloc.Allocate(int(spanData.Len))
					adapterSG.Generate(colors, int(spanData.X), y, int(spanData.Len))
					renBase.BlendColorHspan(int(spanData.X), y, int(spanData.Len), colors, spanData.Covers, basics.CoverFull)
				}
			}
		}
	}
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Distortions",
		Width:  620,
		Height: 360,
	}, &demo{})
}
