// Port of AGG C++ image_transforms.cpp – image fill inside a star polygon.
//
// Renders a star polygon filled with an affine-transformed image (spheres).
// The polygon and image transforms are independent. Default: no rotation,
// scale=1.0, example=0 (polygon fixed, image fixed).
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

const (
	defaultImageName = "spheres"
)

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

func buildStarIT(cx, cy float64, w, h int) *path.PathStorageStl {
	r := float64(w)
	if float64(h) < r {
		r = float64(h)
	}
	r1 := r/3 - 8
	r2 := r1 / 1.45
	nr := 14

	ps := path.NewPathStorageStl()
	for i := 0; i < nr; i++ {
		a := math.Pi*2.0*float64(i)/float64(nr) - math.Pi*0.5
		dx := math.Cos(a)
		dy := math.Sin(a)
		if i&1 != 0 {
			ps.LineTo(cx+dx*r1, cy+dy*r1)
		} else {
			if i == 0 {
				ps.MoveTo(cx+dx*r2, cy+dy*r2)
			} else {
				ps.LineTo(cx+dx*r2, cy+dy*r2)
			}
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

	ctx := agg.NewContextForImage(img)
	canvasW := img.Width()
	canvasH := img.Height()

	ctx.Clear(agg.RGBA(1.0, 1.0, 1.0, 1.0))

	dstImg := img
	dstRbuf := buffer.NewRenderingBufferWithData[uint8](dstImg.Data, dstImg.Width(), dstImg.Height(), dstImg.Width()*4)
	dstPixf := pixfmt.NewPixFmtRGBA32Pre[color.Linear](dstRbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]](dstPixf)
	alloc := span.NewSpanAllocator[color.RGBA8[color.Linear]]()

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	cx, cy := float64(canvasW)/2, float64(canvasH)/2

	// C++ default example=0 => identity image transform.
	imgMtx := transform.NewTransAffine()

	// Image source.
	imgRbuf := buffer.NewRenderingBufferU8()
	imgRbuf.Attach(d.srcImg.Data, d.srcImg.Width(), d.srcImg.Height(), d.srcImg.Width()*4)
	ipf := imagePixFmt{rbuf: imgRbuf}
	accessor := image.NewImageAccessorClip(&ipf, []basics.Int8u{0, 0, 0, 0})
	src := &imageClipSource{accessor: accessor, ipf: &ipf}

	bgRGBA := color.RGBA8[color.Linear]{}
	interp := span.NewSpanInterpolatorLinear[*transform.TransAffine](imgMtx, 8)
	sg := span.NewSpanImageFilterRGBABilinearClipWithParams(src, bgRGBA, interp)
	sgAdp := &spanGenAdapter{sg: sg}

	// Rasterize the star polygon.
	ps := buildStarIT(cx, cy, canvasW, canvasH)
	ras.Reset()
	ras.ClipBox(0, 0, float64(canvasW), float64(canvasH))
	ras.AddPath(&pathSourceAdapter{ps: ps}, 0)

	if ras.RewindScanlines() {
		sl.Reset(ras.MinX(), ras.MaxX())
		for ras.SweepScanline(sl) {
			y := sl.Y()
			for _, spanData := range sl.Spans() {
				if spanData.Len > 0 {
					colors := alloc.Allocate(int(spanData.Len))
					sgAdp.Generate(colors, int(spanData.X), y, int(spanData.Len))
					renBase.BlendColorHspan(int(spanData.X), y, int(spanData.Len), colors, spanData.Covers, basics.CoverFull)
				}
			}
		}
	}

	// Marker for image center (same visual helper as C++).
	a := ctx.GetAgg2D()
	a.ResetTransformations()
	a.FillColor(agg.RGBA(0.7, 0.8, 0.0, 1.0))
	a.NoLine()
	a.FillCircle(cx, cy, 5.0)
	a.FillColor(agg.Black)
	a.NoLine()
	a.FillCircle(cx, cy, 2.0)
}

func main() {
	srcPath := filepath.Join("examples", "shared", "art", defaultImageName+".ppm")
	srcImg, err := loadPPMImage(srcPath)
	if err != nil {
		panic(err)
	}

	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Image Transforms",
		Width:  srcImg.Width(),
		Height: srcImg.Height(),
	}, &demo{srcImg: srcImg})
}
