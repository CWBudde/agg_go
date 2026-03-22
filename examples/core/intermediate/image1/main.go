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
	"github.com/MeKo-Christian/agg_go/internal/conv"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	sliderctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
	"github.com/MeKo-Christian/agg_go/internal/span"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

const defaultImageName = "spheres"

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

type demo struct {
	srcImg *agg.Image
	angle  *sliderctrl.SliderCtrl
	scale  *sliderctrl.SliderCtrl
	w, h   int
}

func newDemo(srcImg *agg.Image) *demo {
	angle := sliderctrl.NewSliderCtrl(5, 5, 300, 12, false)
	scale := sliderctrl.NewSliderCtrl(5, 5+15, 300, 12+15, false)

	angle.SetLabel("Angle=%3.2f")
	scale.SetLabel("Scale=%3.2f")
	angle.SetRange(-180.0, 180.0)
	angle.SetValue(0.0)
	scale.SetRange(0.1, 5.0)
	scale.SetValue(1.0)

	return &demo{
		srcImg: srcImg,
		angle:  angle,
		scale:  scale,
		w:      srcImg.Width() + 20,
		h:      srcImg.Height() + 40 + 20,
	}
}

type ctrlVS struct {
	ctrl ctrlbase.Ctrl[color.RGBA]
}

func (a *ctrlVS) Rewind(id uint32) { a.ctrl.Rewind(uint(id)) }
func (a *ctrlVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

func clampU8(v float64) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	return uint8(v*255.0 + 0.5)
}

func renderCtrl(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	sl *scanline.ScanlineU8,
	renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]],
	ctrl ctrlbase.Ctrl[color.RGBA],
) {
	for pathID := uint(0); pathID < ctrl.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(&ctrlVS{ctrl: ctrl}, uint32(pathID))
		c := ctrl.Color(pathID)
		renscan.RenderScanlinesAASolid(ras, sl, renBase, color.RGBA8[color.Linear]{
			R: clampU8(c.R),
			G: clampU8(c.G),
			B: clampU8(c.B),
			A: clampU8(c.A),
		})
	}
}

// imagePixFmt wraps a RenderingBufferU8 and implements span.RGBASourceInterface.
type imagePixFmt struct {
	rbuf *buffer.RenderingBufferU8
}

func (p *imagePixFmt) Width() int    { return p.rbuf.Width() }
func (p *imagePixFmt) Height() int   { return p.rbuf.Height() }
func (p *imagePixFmt) PixWidth() int { return 4 }
func (p *imagePixFmt) PixPtr(x, y int) []basics.Int8u {
	row := buffer.RowU8(p.rbuf, y)
	return row[x*4:]
}

func (p *imagePixFmt) ColorType() string           { return "RGBA8" }
func (p *imagePixFmt) OrderType() color.ColorOrder { return color.OrderRGBA }
func (p *imagePixFmt) RowPtr(y int) []basics.Int8u { return p.PixPtr(0, y) }

// Stub methods to satisfy RGBASourceInterface
func (p *imagePixFmt) Span(x, y, length int) []basics.Int8u { return nil }
func (p *imagePixFmt) NextX() []basics.Int8u                { return nil }
func (p *imagePixFmt) NextY() []basics.Int8u                { return nil }

type ellipseVS struct {
	e *shapes.Ellipse
}

func (ev *ellipseVS) Rewind(id uint) { ev.e.Rewind(uint32(id)) }
func (ev *ellipseVS) Vertex() (float64, float64, basics.PathCommand) {
	var x, y float64
	cmd := ev.e.Vertex(&x, &y)
	return x, y, cmd
}

// spanGeneratorAdapter wraps SpanImageFilterRGBABilinearClip to satisfy renscan.SpanGeneratorInterface.
type spanGeneratorAdapter struct {
	sg *span.SpanImageFilterRGBABilinearClip[*imagePixFmt, *span.SpanInterpolatorLinear[*transform.TransAffine]]
}

func (a *spanGeneratorAdapter) Prepare() {}
func (a *spanGeneratorAdapter) Generate(colors []color.RGBA8[color.Linear], x, y, length int) {
	a.sg.Generate(colors[:length], x, y)
}

func (d *demo) Render(img *agg.Image) {
	if d.srcImg == nil {
		return
	}

	w, h := img.Width(), img.Height()

	dstRbuf := buffer.NewRenderingBufferU8WithData(img.Data, w, h, img.Stride())
	dstPixf := pixfmt.NewPixFmtRGBA32[color.Linear](dstRbuf)
	dstPixfPre := pixfmt.NewPixFmtRGBA32PreLinear(dstRbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]](dstPixf)
	renBasePre := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]](dstPixfPre)

	renBase.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	initW := float64(d.w)
	initH := float64(d.h)

	srcMtx := transform.NewTransAffine()
	srcMtx.Translate(-initW/2-10, -initH/2-20-10)
	srcMtx.Rotate(d.angle.Value() * math.Pi / 180.0)
	srcMtx.Scale(d.scale.Value())
	srcMtx.Translate(initW/2, initH/2+20)

	imgMtx := transform.NewTransAffine()
	imgMtx.Translate(-initW/2+10, -initH/2+20+10)
	imgMtx.Rotate(d.angle.Value() * math.Pi / 180.0)
	imgMtx.Scale(d.scale.Value())
	imgMtx.Translate(initW/2, initH/2+20)
	imgMtx.Invert()

	imgRbuf := buffer.NewRenderingBufferU8WithData(d.srcImg.Data, d.srcImg.Width(), d.srcImg.Height(), d.srcImg.Width()*4)
	ipf := &imagePixFmt{rbuf: imgRbuf}

	// Bilinear filtered image generator
	interp := span.NewSpanInterpolatorLinearDefault(imgMtx)
	// C++: rgba_pre(0, 0.4, 0, 0.5) => (0, 102, 0, 128)
	clipColor := color.RGBA8[color.Linear]{R: 0, G: 102, B: 0, A: 128}
	sg := span.NewSpanImageFilterRGBABilinearClipWithParams[*imagePixFmt, *span.SpanInterpolatorLinear[*transform.TransAffine]](ipf, clipColor, interp)
	sgAdapter := &spanGeneratorAdapter{sg: sg}

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()
	sa := span.NewSpanAllocator[color.RGBA8[color.Linear]]()

	r := initW
	if initH-60 < r {
		r = initH - 60
	}
	ell := shapes.NewEllipseWithParams(initW/2+10, initH/2+20+10, r/2+16, r/2+16, 200, false)
	tr := conv.NewConvTransform[conv.VertexSource, *transform.TransAffine](&ellipseVS{e: ell}, srcMtx)

	ras.Reset()
	ras.AddPath(conv.NewRasterizerVertexSourceAdapter(tr), 0)
	renscan.RenderScanlinesAA(ras, sl, renBasePre, sa, sgAdapter)

	renderCtrl(ras, sl, renBase, d.angle)
	renderCtrl(ras, sl, renBase, d.scale)
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	if btn.Left {
		if d.angle.OnMouseButtonDown(fx, fy) || d.scale.OnMouseButtonDown(fx, fy) {
			return true
		}
	}
	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	if d.angle.OnMouseMove(fx, fy, btn.Left) || d.scale.OnMouseMove(fx, fy, btn.Left) {
		return true
	}
	return false
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	if d.angle.OnMouseButtonUp(fx, fy) || d.scale.OnMouseButtonUp(fx, fy) {
		return true
	}
	return false
}

func main() {
	// Paths to look for spheres.ppm
	srcPath := filepath.Join("examples", "shared", "art", defaultImageName+".ppm")

	srcImg, err := loadPPMImage(srcPath)
	if err != nil {
		panic(err)
	}

	d := newDemo(srcImg)
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "AGG Example. Image Affine Transformations with filtering",
		Width:  d.w,
		Height: d.h,
		FlipY:  true,
	}, d)
}
