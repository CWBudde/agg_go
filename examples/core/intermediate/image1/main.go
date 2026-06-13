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

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/examples/shared/lowlevelrunner"
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	ctrlbase "github.com/cwbudde/agg_go/internal/ctrl"
	sliderctrl "github.com/cwbudde/agg_go/internal/ctrl/slider"
	"github.com/cwbudde/agg_go/internal/image"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/shapes"
	"github.com/cwbudde/agg_go/internal/span"
	"github.com/cwbudde/agg_go/internal/transform"
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

	// C++ platform_support::load_img converts the sRGB image file into the
	// linear window pixel format (color_type = rgba8 = linear under AGG_BGR24).
	// Decode each channel sRGB->linear so all filtering happens in linear space.
	buf := make([]uint8, w*h*4)
	for p := 0; p < w*h; p++ {
		lin := color.ConvertRGB8SRGBToLinear(color.RGB8[color.SRGB]{
			R: rgb[p*3+0],
			G: rgb[p*3+1],
			B: rgb[p*3+2],
		})
		buf[p*4+0] = lin.R
		buf[p*4+1] = lin.G
		buf[p*4+2] = lin.B
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

// displayRGBA8 converts a control's float RGBA into linear bytes. C++ slider
// colors are rgba() floats that map directly to linear rgba8 (color_type) via
// *255; the whole frame is sRGB-encoded at save time.
func displayRGBA8(c color.RGBA) color.RGBA8[color.Linear] {
	return color.RGBA8[color.Linear]{
		R: clampU8(c.R),
		G: clampU8(c.G),
		B: clampU8(c.B),
		A: clampU8(c.A),
	}
}

// rgba8Pre premultiplies a straight float RGBA into linear bytes, matching
// C++ agg::rgba_pre(r,g,b,a).
func rgba8Pre(r, g, b, a float64) color.RGBA8[color.Linear] {
	return color.RGBA8[color.Linear]{
		R: clampU8(r * a),
		G: clampU8(g * a),
		B: clampU8(b * a),
		A: clampU8(a),
	}
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
		renscan.RenderScanlinesAASolid(ras, sl, renBase, displayRGBA8(ctrl.Color(pathID)))
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

// rgbBilinearClipSpanGenerator ports the RGB variant used by image1.cpp:
// span_image_filter_rgb_bilinear_clip<pixfmt, interpolator_type>.
type rgbBilinearClipSpanGenerator struct {
	src    *imagePixFmt
	back   color.RGBA8[color.Linear]
	interp *span.SpanInterpolatorLinear[*transform.TransAffine]
}

func (g *rgbBilinearClipSpanGenerator) Prepare() {}

func (g *rgbBilinearClipSpanGenerator) Generate(colors []color.RGBA8[color.Linear], x, y, length int) {
	if length > len(colors) {
		length = len(colors)
	}
	if length <= 0 {
		return
	}

	g.interp.Begin(float64(x)+0.5, float64(y)+0.5, length)

	backR := int(g.back.R)
	backG := int(g.back.G)
	backB := int(g.back.B)
	backA := int(g.back.A)
	maxX := g.src.Width() - 1
	maxY := g.src.Height() - 1

	for i := 0; i < length; i++ {
		xHr, yHr := g.interp.Coordinates()
		xHr -= image.ImageSubpixelScale / 2
		yHr -= image.ImageSubpixelScale / 2

		xLr := xHr >> image.ImageSubpixelShift
		yLr := yHr >> image.ImageSubpixelShift
		var fg [3]int
		srcAlpha := 0

		if xLr >= 0 && yLr >= 0 && xLr < maxX && yLr < maxY {
			xHr &= image.ImageSubpixelMask
			yHr &= image.ImageSubpixelMask

			row := g.src.RowPtr(yLr)
			off := xLr * 4
			weight := (image.ImageSubpixelScale - xHr) * (image.ImageSubpixelScale - yHr)
			fg[0] += weight * int(row[off+0])
			fg[1] += weight * int(row[off+1])
			fg[2] += weight * int(row[off+2])

			weight = xHr * (image.ImageSubpixelScale - yHr)
			fg[0] += weight * int(row[off+4])
			fg[1] += weight * int(row[off+5])
			fg[2] += weight * int(row[off+6])

			row = g.src.RowPtr(yLr + 1)
			weight = (image.ImageSubpixelScale - xHr) * yHr
			fg[0] += weight * int(row[off+0])
			fg[1] += weight * int(row[off+1])
			fg[2] += weight * int(row[off+2])

			weight = xHr * yHr
			fg[0] += weight * int(row[off+4])
			fg[1] += weight * int(row[off+5])
			fg[2] += weight * int(row[off+6])

			fg[0] >>= image.ImageSubpixelShift * 2
			fg[1] >>= image.ImageSubpixelShift * 2
			fg[2] >>= image.ImageSubpixelShift * 2
			srcAlpha = 255
		} else if xLr < -1 || yLr < -1 || xLr > maxX || yLr > maxY {
			fg[0] = backR
			fg[1] = backG
			fg[2] = backB
			srcAlpha = backA
		} else {
			xHr &= image.ImageSubpixelMask
			yHr &= image.ImageSubpixelMask

			sample := func(sampleX, sampleY, weight int) {
				if sampleX >= 0 && sampleY >= 0 && sampleX <= maxX && sampleY <= maxY {
					row := g.src.RowPtr(sampleY)
					off := sampleX * 4
					fg[0] += weight * int(row[off+0])
					fg[1] += weight * int(row[off+1])
					fg[2] += weight * int(row[off+2])
					srcAlpha += weight * 255
					return
				}
				fg[0] += backR * weight
				fg[1] += backG * weight
				fg[2] += backB * weight
				srcAlpha += backA * weight
			}

			sample(xLr, yLr, (image.ImageSubpixelScale-xHr)*(image.ImageSubpixelScale-yHr))
			sample(xLr+1, yLr, xHr*(image.ImageSubpixelScale-yHr))
			sample(xLr, yLr+1, (image.ImageSubpixelScale-xHr)*yHr)
			sample(xLr+1, yLr+1, xHr*yHr)

			fg[0] >>= image.ImageSubpixelShift * 2
			fg[1] >>= image.ImageSubpixelShift * 2
			fg[2] >>= image.ImageSubpixelShift * 2
			srcAlpha >>= image.ImageSubpixelShift * 2
		}

		colors[i] = color.RGBA8[color.Linear]{
			R: basics.Int8u(fg[0]),
			G: basics.Int8u(fg[1]),
			B: basics.Int8u(fg[2]),
			A: basics.Int8u(srcAlpha),
		}
		g.interp.Next()
	}
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
	imgMtx.Translate(initW*0.5, initH*0.5+20)
	imgMtx.Invert()

	imgRbuf := buffer.NewRenderingBufferU8WithData(d.srcImg.Data, d.srcImg.Width(), d.srcImg.Height(), -d.srcImg.Stride())
	ipf := &imagePixFmt{rbuf: imgRbuf}

	// Bilinear filtered image generator
	interp := span.NewSpanInterpolatorLinearDefault(imgMtx)
	clipColor := rgba8Pre(0, 0.4, 0, 0.5)
	sgAdapter := &rgbBilinearClipSpanGenerator{src: ipf, back: clipColor, interp: interp}

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
	lowlevelrunner.Run(runnerConfig(d), d)
}

func runnerConfig(d *demo) lowlevelrunner.Config {
	return lowlevelrunner.Config{
		Title:                 "AGG Example. Image Affine Transformations with filtering",
		Width:                 d.w,
		Height:                d.h,
		FlipY:                 true,
		EncodeLinearRGBToSRGB: true,
	}
}
