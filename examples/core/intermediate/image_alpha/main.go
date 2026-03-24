// Port of AGG C++ image_alpha.cpp – brightness-to-alpha image compositing.
//
// Loads spheres.bmp, draws random background ellipses, then composites the
// transformed image through a transformed ellipse while converting RGB
// brightness to output alpha via a 6-point spline LUT.
package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
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
func (p imagePixFmt) PixWidth() int { return 3 }
func (p imagePixFmt) PixPtr(x, y int) []basics.Int8u {
	row := buffer.RowU8(p.rbuf, y)
	return row[x*3:]
}

type imageClipSource struct {
	accessor *image.ImageAccessorClip[imagePixFmt]
	ipf      *imagePixFmt
}

func (s *imageClipSource) Width() int                  { return s.ipf.Width() }
func (s *imageClipSource) Height() int                 { return s.ipf.Height() }
func (s *imageClipSource) ColorType() string           { return "RGB8" }
func (s *imageClipSource) OrderType() color.ColorOrder { return color.OrderBGR }
func (s *imageClipSource) Span(x, y, l int) []basics.Int8u {
	return s.accessor.Span(x, y, l)
}
func (s *imageClipSource) NextX() []basics.Int8u { return s.accessor.NextX() }
func (s *imageClipSource) NextY() []basics.Int8u { return s.accessor.NextY() }
func (s *imageClipSource) RowPtr(y int) []basics.Int8u {
	return s.ipf.PixPtr(0, y)
}

// imgAlphaSpanGen wraps bilinear RGB sampling and converts brightness to alpha.
type imgAlphaSpanGen struct {
	inner *span.SpanImageFilterRGBBilinearClip[*imageClipSource, *span.SpanInterpolatorLinear[*transform.TransAffine]]
	lut   [256 * 3]uint8
}

func (g *imgAlphaSpanGen) Prepare() {}
func (g *imgAlphaSpanGen) Generate(colors []color.RGBA8[color.Linear], x, y, length int) {
	if length > len(colors) {
		length = len(colors)
	}
	tmp := make([]color.RGB8[color.Linear], length)
	g.inner.Generate(tmp, x, y)
	for i := 0; i < length; i++ {
		src := tmp[i]
		c := &colors[i]
		c.R = src.R
		c.G = src.G
		c.B = src.B
		sum := int(src.R) + int(src.G) + int(src.B) // 0..765
		idx := (sum * len(g.lut)) / (3 * 255)       // match C++ scaling
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

func newClibcRandSeed1() *clibcRand {
	return &clibcRand{
		state: [31]int32{
			-1726662223, 379960547, 1735697613, 1040273694, 1313901226,
			1627687941, -179304937, -2073333483, 1780058412, -1989503057,
			-615974602, 344556628, 939512070, -1249116260, 1507946756,
			-812545463, 154635395, 1388815473, -1926676823, 525320961,
			-1009028674, 968117788, -123449607, 1284210865, 435012392,
			-2017506339, -911064859, -370259173, 1132637927, 1398500161, -205601318,
		},
		fptr: 3,
		rptr: 0,
	}
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

func loadImageAsset(filename string) (*agg.Image, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	switch filepath.Ext(filename) {
	case ".bmp", ".BMP":
		return loadBMPImage(f)
	case ".ppm", ".PPM":
		return loadPPMImage(f)
	default:
		if img, err := loadBMPImage(f); err == nil {
			return img, nil
		}
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
		return loadPPMImage(f)
	}
}

func loadBMPImage(r io.ReadSeeker) (*agg.Image, error) {
	var fileHeader struct {
		Type      uint16
		Size      uint32
		Reserved1 uint16
		Reserved2 uint16
		OffBits   uint32
	}
	if err := binary.Read(r, binary.LittleEndian, &fileHeader); err != nil {
		return nil, err
	}
	if fileHeader.Type != 0x4D42 {
		return nil, errors.New("not a BMP file")
	}

	var infoHeader struct {
		Size          uint32
		Width         int32
		Height        int32
		Planes        uint16
		BitCount      uint16
		Compression   uint32
		SizeImage     uint32
		XPelsPerMeter int32
		YPelsPerMeter int32
		ClrUsed       uint32
		ClrImportant  uint32
	}
	if err := binary.Read(r, binary.LittleEndian, &infoHeader); err != nil {
		return nil, err
	}
	if infoHeader.Planes != 1 {
		return nil, fmt.Errorf("unsupported BMP planes: %d", infoHeader.Planes)
	}
	if infoHeader.Compression != 0 {
		return nil, fmt.Errorf("unsupported BMP compression: %d", infoHeader.Compression)
	}
	if infoHeader.BitCount != 24 && infoHeader.BitCount != 32 {
		return nil, fmt.Errorf("unsupported BMP bit depth: %d", infoHeader.BitCount)
	}

	width := int(infoHeader.Width)
	height := int(infoHeader.Height)
	if width <= 0 || height == 0 {
		return nil, fmt.Errorf("invalid BMP dimensions: %dx%d", width, height)
	}
	if height < 0 {
		height = -height
	}
	if _, err := r.Seek(int64(fileHeader.OffBits), io.SeekStart); err != nil {
		return nil, err
	}

	rowStride := ((width*int(infoHeader.BitCount) + 31) / 32) * 4
	rowData := make([]byte, rowStride)
	buf := make([]uint8, width*height*3)

	for y := 0; y < height; y++ {
		if _, err := io.ReadFull(r, rowData); err != nil {
			return nil, err
		}
		dstY := y
		for x := 0; x < width; x++ {
			src := x * int(infoHeader.BitCount) / 8
			dst := (dstY*width + x) * 3
			buf[dst+0] = rowData[src+0]
			buf[dst+1] = rowData[src+1]
			buf[dst+2] = rowData[src+2]
		}
	}

	return agg.NewImage(buf, width, height, width*3), nil
}

func loadPPMImage(r io.ReadSeeker) (*agg.Image, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if len(data) < 2 || data[0] != 'P' || data[1] != '6' {
		return nil, errors.New("unsupported PPM format")
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
			if b == ' ' || b == '\t' || b == '\r' || b == '\n' {
				i++
				continue
			}
			break
		}
		if i >= len(data) {
			return "", errors.New("unexpected end of PPM header")
		}
		start := i
		for i < len(data) {
			b := data[i]
			if b == ' ' || b == '\t' || b == '\r' || b == '\n' || b == '#' {
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
	width, err := strconv.Atoi(wTok)
	if err != nil || width <= 0 {
		return nil, fmt.Errorf("invalid PPM width: %q", wTok)
	}
	height, err := strconv.Atoi(hTok)
	if err != nil || height <= 0 {
		return nil, fmt.Errorf("invalid PPM height: %q", hTok)
	}
	maxVal, err := strconv.Atoi(maxTok)
	if err != nil || maxVal != 255 {
		return nil, fmt.Errorf("unsupported PPM max value: %q", maxTok)
	}

	for i < len(data) && (data[i] == ' ' || data[i] == '\t' || data[i] == '\r' || data[i] == '\n') {
		i++
	}
	if len(data)-i < width*height*3 {
		return nil, errors.New("PPM payload too short")
	}

	buf := make([]uint8, width*height*3)
	rgb := data[i:]
	for p := 0; p < width*height; p++ {
		src := p * 3
		dst := p * 3
		buf[dst+0] = rgb[src+0]
		buf[dst+1] = rgb[src+1]
		buf[dst+2] = rgb[src+2]
	}

	return agg.NewImage(buf, width, height, width*3), nil
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
	rng := newClibcRandSeed1()
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
	imgRbuf.Attach(d.srcImg.Data, d.srcImg.Width(), d.srcImg.Height(), -d.srcImg.Stride())
	ipf := imagePixFmt{rbuf: imgRbuf}
	accessor := image.NewImageAccessorClip(&ipf, []basics.Int8u{0, 0, 0, 0})
	src := &imageClipSource{accessor: accessor, ipf: &ipf}

	interp := span.NewSpanInterpolatorLinear[*transform.TransAffine](imgMtx, 8)
	sg := &imgAlphaSpanGen{inner: span.NewSpanImageFilterRGBBilinearClipWithParams(src, color.RGB8[color.Linear]{}, interp)}

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
	srcPath := filepath.Join("examples", "shared", "art", defaultImageName+".bmp")
	srcImg, err := loadImageAsset(srcPath)
	if err != nil {
		srcPath = filepath.Join("examples", "shared", "art", defaultImageName+".ppm")
		srcImg, err = loadImageAsset(srcPath)
		if err != nil {
			panic(err)
		}
	}

	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Image Alpha",
		Width:  320,
		Height: 300,
		FlipY:  true,
	}, &demo{srcImg: srcImg})
}
