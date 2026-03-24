// Port of AGG C++ image_alpha.cpp – brightness-to-alpha image compositing.
//
// Loads spheres.ppm, draws random background ellipses, then composites the
// transformed image through a transformed ellipse while converting RGB
// brightness to output alpha via a 6-point spline LUT.
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
	"github.com/MeKo-Christian/agg_go/internal/ctrl/spline"
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

// imgAlphaSpanGen wraps bilinear and converts brightness to alpha.
type imgAlphaSpanGen struct {
	inner *span.SpanImageFilterRGBABilinearClip[*imageClipSource, *span.SpanInterpolatorLinear[*transform.TransAffine]]
	lut   [256 * 3]uint8
}

func (g *imgAlphaSpanGen) Prepare() {}
func (g *imgAlphaSpanGen) Generate(colors []color.RGBA8[color.Linear], x, y, length int) {
	if length > len(colors) {
		length = len(colors)
	}
	g.inner.Generate(colors[:length], x, y)
	for i := 0; i < length; i++ {
		c := &colors[i]
		sum := int(c.R) + int(c.G) + int(c.B) // 0..765
		idx := (sum * len(g.lut)) / (3 * 255) // match C++ scaling
		if idx >= len(g.lut) {
			idx = len(g.lut) - 1
		}
		c.A = g.lut[idx]
	}
}

// rasScanlineAdapter adapts ScanlineU8 to rasterizer.ScanlineInterface.
// pathSourceAdapter bridges PathStorageStl to rasterizer VertexSource.
type pathSourceAdapter struct{ ps *path.PathStorageStl }

func (a *pathSourceAdapter) Rewind(id uint32) { a.ps.Rewind(uint(id)) }
func (a *pathSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ps.NextVertex()
	*x = vx
	*y = vy
	return cmd
}

type ctrlPathSource interface {
	NumPaths() uint
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
	Color(pathID uint) color.RGBA8[color.Linear]
}

type ctrlPathAdapter struct {
	ctrl ctrlPathSource
}

func (a *ctrlPathAdapter) Rewind(pathID uint32) {
	a.ctrl.Rewind(uint(pathID))
}

func (a *ctrlPathAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

func toAggColor(c color.RGBA8[color.Linear]) agg.Color {
	clamp := func(v basics.Int8u) uint8 {
		return uint8(v)
	}
	return agg.NewColor(clamp(c.R), clamp(c.G), clamp(c.B), clamp(c.A))
}

type clibcRand struct {
	state [31]int32
	fptr  int
	rptr  int
}

func newClibcRandSeed(seed int32) *clibcRand {
	if seed == 0 {
		seed = 1
	}

	r := &clibcRand{}
	r.state[0] = seed
	for i := 1; i < len(r.state); i++ {
		next := (16807 * int64(r.state[i-1])) % 2147483647
		r.state[i] = int32(next)
	}
	r.fptr = 3
	r.rptr = 0
	for i := 0; i < 310; i++ {
		r.next()
	}
	return r
}

func (r *clibcRand) next() int32 {
	r.state[r.fptr] += r.state[r.rptr]
	result := int32(uint32(r.state[r.fptr]) >> 1)
	r.fptr++
	if r.fptr >= len(r.state) {
		r.fptr = 0
	}
	r.rptr++
	if r.rptr >= len(r.state) {
		r.rptr = 0
	}
	return result
}

func (r *clibcRand) randN(n int) int {
	return int(r.next()) % n
}

func renderCtrl(ctx *agg.Context, ctrl ctrlPathSource) {
	a := ctx.GetAgg2D()
	ras := a.GetInternalRasterizer()
	adapter := &ctrlPathAdapter{ctrl: ctrl}

	for pathID := uint(0); pathID < ctrl.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(adapter, uint32(pathID))
		a.RenderRasterizerWithColor(toAggColor(ctrl.Color(pathID)))
	}
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

func buildTransformedEllipsePath(w, h int, mtx *transform.TransAffine) *path.PathStorageStl {
	cx := float64(w) * 0.5
	cy := float64(h) * 0.5
	rx := float64(w) / 1.9
	ry := float64(h) / 1.9

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
	a := ctx.GetAgg2D()
	a.ResetTransformations()
	ctx.Clear(agg.RGBA(1, 1, 1, 1))

	// C++ on_init uses rand() without an explicit srand(), so match its
	// default seed and generator instead of Go's PRNG.
	rng := newClibcRandSeed(1)
	for i := 0; i < 50; i++ {
		x := float64(rng.randN(canvasW))
		y := float64(rng.randN(canvasH))
		rx := float64(rng.randN(60) + 10)
		ry := float64(rng.randN(60) + 10)
		a.FillColor(agg.NewColor(
			uint8(rng.randN(256)),
			uint8(rng.randN(256)),
			uint8(rng.randN(256)),
			uint8(rng.randN(256)),
		))
		a.NoLine()
		a.Ellipse(x, y, rx, ry)
	}

	dstRbuf := buffer.NewRenderingBufferWithData[uint8](img.Data, img.Width(), img.Height(), img.Stride())
	dstPixf := pixfmt.NewPixFmtRGBA32Pre[color.Linear](dstRbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]](dstPixf)
	alloc := span.NewSpanAllocator[color.RGBA8[color.Linear]]()

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	angle := 10.0 * math.Pi / 180.0
	srcMtx := transform.NewTransAffine()
	srcMtx.Translate(-float64(canvasW)*0.5, -float64(canvasH)*0.5)
	srcMtx.Rotate(angle)
	srcMtx.Translate(float64(canvasW)*0.5, float64(canvasH)*0.5)

	imgMtx := transform.NewTransAffine()
	imgMtx.Translate(-float64(canvasW)*0.5, -float64(canvasH)*0.5)
	imgMtx.Rotate(angle)
	imgMtx.Translate(float64(canvasW)*0.5, float64(canvasH)*0.5)
	imgMtx.Invert()

	imgRbuf := buffer.NewRenderingBufferU8()
	imgRbuf.Attach(d.srcImg.Data, d.srcImg.Width(), d.srcImg.Height(), d.srcImg.Stride())
	ipf := imagePixFmt{rbuf: imgRbuf}
	accessor := image.NewImageAccessorClip(&ipf, []basics.Int8u{0, 0, 0, 0})
	src := &imageClipSource{accessor: accessor, ipf: &ipf}

	interp := span.NewSpanInterpolatorLinear[*transform.TransAffine](imgMtx, 8)
	innerSG := span.NewSpanImageFilterRGBABilinearClipWithParams(src, color.RGBA8[color.Linear]{}, interp)
	sg := &imgAlphaSpanGen{inner: innerSG}

	alphaCtrl := spline.NewSplineCtrl[color.RGBA8[color.Linear]](2, 2, 200, 30, 6, false)
	alphaCtrl.SetBackgroundColor(color.NewRGBA8[color.Linear](255, 255, 230, 255))
	alphaCtrl.SetBorderColor(color.NewRGBA8[color.Linear](0, 0, 0, 255))
	alphaCtrl.SetCurveColor(color.NewRGBA8[color.Linear](0, 0, 0, 255))
	alphaCtrl.SetInactivePointColor(color.NewRGBA8[color.Linear](0, 0, 0, 255))
	alphaCtrl.SetActivePointColor(color.NewRGBA8[color.Linear](255, 0, 0, 255))
	alphaCtrl.SetValue(0, 1.0)
	alphaCtrl.SetValue(1, 1.0)
	alphaCtrl.SetValue(2, 1.0)
	alphaCtrl.SetValue(3, 0.5)
	alphaCtrl.SetValue(4, 0.5)
	alphaCtrl.SetValue(5, 1.0)
	for i := range sg.lut {
		t := float64(i) / float64(len(sg.lut))
		sg.lut[i] = uint8(alphaCtrl.Value(t)*255.0 + 0.5)
	}

	clipPath := buildTransformedEllipsePath(canvasW, canvasH, srcMtx)

	ras.Reset()
	ras.ClipBox(0, 0, float64(canvasW), float64(canvasH))
	ras.AddPath(&pathSourceAdapter{ps: clipPath}, 0)

	if ras.RewindScanlines() {
		sl.Reset(ras.MinX(), ras.MaxX())
		for ras.SweepScanline(sl) {
			y := sl.Y()
			for _, spanData := range sl.Spans() {
				if spanData.Len <= 0 {
					continue
				}
				colors := alloc.Allocate(int(spanData.Len))
				sg.Generate(colors, int(spanData.X), y, int(spanData.Len))
				renBase.BlendColorHspan(int(spanData.X), y, int(spanData.Len), colors, spanData.Covers, basics.CoverFull)
			}
		}
	}

	renderCtrl(ctx, alphaCtrl)
}

func main() {
	srcPath := filepath.Join("examples", "shared", "art", defaultImageName+".ppm")
	srcImg, err := loadPPMImage(srcPath)
	if err != nil {
		panic(err)
	}

	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Image Alpha",
		Width:  320,
		Height: 300,
		FlipY:  true,
	}, &demo{srcImg: srcImg})
}
