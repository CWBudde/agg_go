// Port of AGG C++ image1.cpp – affine-transformed image fill inside an ellipse.
//
// Loads spheres.ppm and fills a transformed ellipse with bilinear-filtered image
// samples using the same default transform setup as the original demo
// (angle=0, scale=1).
package main

import (
	"bytes"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strconv"

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

const defaultImageName = "spheres"

// imagePixFmt wraps a RenderingBufferU8 and implements image.PixelFormat.
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

// imageClipSource implements span.RGBASourceInterface for image access.
type imageClipSource struct {
	accessor *image.ImageAccessorClip[imagePixFmt]
	ipf      *imagePixFmt
}

func (s *imageClipSource) Width() int                  { return s.ipf.Width() }
func (s *imageClipSource) Height() int                 { return s.ipf.Height() }
func (s *imageClipSource) ColorType() string           { return "RGBA8" }
func (s *imageClipSource) OrderType() color.ColorOrder { return color.OrderRGBA }
func (s *imageClipSource) Span(x, y, l int) []basics.Int8u {
	return s.accessor.Span(x, y, l)
}
func (s *imageClipSource) NextX() []basics.Int8u { return s.accessor.NextX() }
func (s *imageClipSource) NextY() []basics.Int8u { return s.accessor.NextY() }
func (s *imageClipSource) RowPtr(y int) []basics.Int8u {
	return s.ipf.PixPtr(0, y)
}

// spanGenAdapter wraps SpanImageFilterRGBABilinearClip for the render loop.
type spanGenAdapter struct {
	sg *span.SpanImageFilterRGBABilinearClip[*imageClipSource, *span.SpanInterpolatorLinear[*transform.TransAffine]]
}

func (a *spanGenAdapter) Prepare() {}
func (a *spanGenAdapter) Generate(colors []color.RGBA8[color.Linear], x, y, length int) {
	if length > len(colors) {
		length = len(colors)
	}
	a.sg.Generate(colors[:length], x, y)
}

// rasScanlineAdapter adapts ScanlineU8 to rasterizer.ScanlineInterface.
type rasScanlineAdapter struct{ sl *scanline.ScanlineU8 }

func (a *rasScanlineAdapter) ResetSpans()                { a.sl.ResetSpans() }
func (a *rasScanlineAdapter) AddCell(x int, c uint32)    { a.sl.AddCell(x, uint(c)) }
func (a *rasScanlineAdapter) AddSpan(x, l int, c uint32) { a.sl.AddSpan(x, l, uint(c)) }
func (a *rasScanlineAdapter) Finalize(y int)             { a.sl.Finalize(y) }
func (a *rasScanlineAdapter) NumSpans() int              { return a.sl.NumSpans() }

// pathSourceAdapter bridges PathStorageStl to rasterizer VertexSource.
type pathSourceAdapter struct{ ps *path.PathStorageStl }

func (a *pathSourceAdapter) Rewind(id uint32) { a.ps.Rewind(uint(id)) }
func (a *pathSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ps.NextVertex()
	*x = vx
	*y = vy
	return cmd
}

func loadPPMImage(filename string) (*agg.Image, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if len(data) < 2 || data[0] != 'P' || data[1] != '6' {
		return nil, errors.New("unsupported ppm format: expected P6")
	}

	i := 2
	readToken := func() (string, error) {
		for i < len(data) {
			b := data[i]
			if b == '#' {
				for i < len(data) && data[i] != '\n' {
					i++
				}
				continue
			}
			if bytes.IndexByte([]byte{' ', '\t', '\n', '\r'}, b) >= 0 {
				i++
				continue
			}
			break
		}
		if i >= len(data) {
			return "", errors.New("unexpected end of ppm header")
		}
		start := i
		for i < len(data) {
			b := data[i]
			if bytes.IndexByte([]byte{' ', '\t', '\n', '\r', '#'}, b) >= 0 {
				break
			}
			i++
		}
		return string(data[start:i]), nil
	}

	wTok, err := readToken()
	if err != nil {
		return nil, err
	}
	hTok, err := readToken()
	if err != nil {
		return nil, err
	}
	maxTok, err := readToken()
	if err != nil {
		return nil, err
	}
	w, err := strconv.Atoi(wTok)
	if err != nil || w <= 0 {
		return nil, errors.New("invalid ppm width")
	}
	h, err := strconv.Atoi(hTok)
	if err != nil || h <= 0 {
		return nil, errors.New("invalid ppm height")
	}
	maxV, err := strconv.Atoi(maxTok)
	if err != nil || maxV != 255 {
		return nil, errors.New("unsupported ppm max value")
	}

	for i < len(data) && (data[i] == ' ' || data[i] == '\t' || data[i] == '\n' || data[i] == '\r') {
		i++
	}
	rgb := data[i:]
	if len(rgb) < w*h*3 {
		return nil, errors.New("ppm pixel data too short")
	}

	buf := make([]uint8, w*h*4)
	for p := 0; p < w*h; p++ {
		buf[p*4+0] = rgb[p*3+0]
		buf[p*4+1] = rgb[p*3+1]
		buf[p*4+2] = rgb[p*3+2]
		buf[p*4+3] = 255
	}

	return agg.NewImage(buf, w, h, w*4), nil
}

func buildTransformedEllipsePath(canvasW, canvasH int, mtx *transform.TransAffine) *path.PathStorageStl {
	initialW := float64(canvasW)
	initialH := float64(canvasH)
	r := initialW
	if initialH-60.0 < r {
		r = initialH - 60.0
	}
	rx := r*0.5 + 16.0
	ry := r*0.5 + 16.0
	cx := initialW*0.5 + 10.0
	cy := initialH*0.5 + 30.0

	ps := path.NewPathStorageStl()
	const steps = 200
	for i := 0; i <= steps; i++ {
		a := 2 * math.Pi * float64(i) / float64(steps)
		x := cx + rx*math.Cos(a)
		y := cy + ry*math.Sin(a)
		mtx.Transform(&x, &y)
		if i == 0 {
			ps.MoveTo(x, y)
		} else {
			ps.LineTo(x, y)
		}
	}
	ps.ClosePolygon(basics.PathFlagsNone)
	return ps
}

type demo struct {
	srcImg *agg.Image
}

func (d *demo) Render(img *agg.Image) {
	if d.srcImg == nil {
		return
	}

	canvasW := img.Width()
	canvasH := img.Height()

	ctx := agg.NewContextForImage(img)
	ctx.Clear(agg.RGBA(1, 1, 1, 1))

	dstRbuf := buffer.NewRenderingBufferWithData[uint8](img.Data, img.Width(), img.Height(), img.Width()*4)
	dstPixf := pixfmt.NewPixFmtRGBA32Pre[color.Linear](dstRbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]](dstPixf)
	alloc := span.NewSpanAllocator[color.RGBA8[color.Linear]]()

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	angle := 0.0
	scale := 1.0
	initialW := float64(canvasW)
	initialH := float64(canvasH)

	srcMtx := transform.NewTransAffine()
	srcMtx.Translate(-initialW*0.5-10.0, -initialH*0.5-30.0)
	srcMtx.Rotate(angle * math.Pi / 180.0)
	srcMtx.Scale(scale)
	srcMtx.Translate(initialW*0.5, initialH*0.5+20.0)

	imgMtx := transform.NewTransAffine()
	imgMtx.Translate(-initialW*0.5+10.0, -initialH*0.5+30.0)
	imgMtx.Rotate(angle * math.Pi / 180.0)
	imgMtx.Scale(scale)
	imgMtx.Translate(initialW*0.5, initialH*0.5+20.0)
	imgMtx.Invert()

	imgRbuf := buffer.NewRenderingBufferU8()
	imgRbuf.Attach(d.srcImg.Data, d.srcImg.Width(), d.srcImg.Height(), d.srcImg.Width()*4)
	ipf := imagePixFmt{rbuf: imgRbuf}

	// C++ clip color: rgba_pre(0, 0.4, 0, 0.5)
	accessor := image.NewImageAccessorClip(&ipf, []basics.Int8u{0, 102, 0, 128})
	src := &imageClipSource{accessor: accessor, ipf: &ipf}

	bgRGBA := color.RGBA8[color.Linear]{R: 0, G: 102, B: 0, A: 128}
	interp := span.NewSpanInterpolatorLinear[*transform.TransAffine](imgMtx, 8)
	sg := span.NewSpanImageFilterRGBABilinearClipWithParams(src, bgRGBA, interp)
	sgAdp := &spanGenAdapter{sg: sg}

	clipPath := buildTransformedEllipsePath(canvasW, canvasH, srcMtx)

	ras.Reset()
	ras.ClipBox(0, 0, float64(canvasW), float64(canvasH))
	ras.AddPath(&pathSourceAdapter{ps: clipPath}, 0)

	if ras.RewindScanlines() {
		sl.Reset(ras.MinX(), ras.MaxX())
		for ras.SweepScanline(&rasScanlineAdapter{sl: sl}) {
			y := sl.Y()
			for _, spanData := range sl.Spans() {
				if spanData.Len <= 0 {
					continue
				}
				colors := alloc.Allocate(int(spanData.Len))
				sgAdp.Generate(colors, int(spanData.X), y, int(spanData.Len))
				renBase.BlendColorHspan(int(spanData.X), y, int(spanData.Len), colors, spanData.Covers, basics.CoverFull)
			}
		}
	}
}

func main() {
	srcPath := filepath.Join("examples", "shared", "art", defaultImageName+".ppm")
	srcImg, err := loadPPMImage(srcPath)
	if err != nil {
		panic(err)
	}

	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Image1",
		Width:  340,
		Height: 360,
	}, &demo{srcImg: srcImg})
}
