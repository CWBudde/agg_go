package main

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

// Port of AGG C++ rasterizer_compound.cpp.
//
// Web variant uses URL/JS parameters instead of AGG widgets.
var (
	compoundWidth  = 10.0
	compoundAlpha1 = 1.0
	compoundAlpha2 = 1.0
	compoundAlpha3 = 1.0
	compoundAlpha4 = 1.0
	compoundInvert = false
)

const (
	compoundRefW = 440.0
	compoundRefH = 330.0
)

func setCompoundWidth(v float64) { compoundWidth = v }
func setCompoundAlpha1(v float64) {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	compoundAlpha1 = v
}

func setCompoundAlpha2(v float64) {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	compoundAlpha2 = v
}

func setCompoundAlpha3(v float64) {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	compoundAlpha3 = v
}

func setCompoundAlpha4(v float64) {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	compoundAlpha4 = v
}
func setCompoundInvert(v bool) { compoundInvert = v }

type rcStyleHandler struct {
	styles []color.RGBA8[color.Linear]
}

func (h *rcStyleHandler) IsSolid(style int) bool { return true }
func (h *rcStyleHandler) Color(style int) color.RGBA8[color.Linear] {
	if style < 0 || style >= len(h.styles) {
		return color.RGBA8[color.Linear]{}
	}
	return h.styles[style]
}
func (h *rcStyleHandler) GenerateSpan(colors []color.RGBA8[color.Linear], x, y, length, style int) {}

type rcSLAdapter struct{ sl *scanline.ScanlineU8 }

func (a *rcSLAdapter) ResetSpans()                      { a.sl.ResetSpans() }
func (a *rcSLAdapter) AddCell(x int, c basics.Int8u)    { a.sl.AddCell(x, uint(c)) }
func (a *rcSLAdapter) AddSpan(x, l int, c basics.Int8u) { a.sl.AddSpan(x, l, uint(c)) }
func (a *rcSLAdapter) Finalize(y int)                   { a.sl.Finalize(y) }
func (a *rcSLAdapter) NumSpans() int                    { return a.sl.NumSpans() }

type rcConvVertexSource interface {
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
}

type rcConvVSAdapter struct {
	vs rcConvVertexSource
}

func (a *rcConvVSAdapter) Rewind(pathID uint32) {
	a.vs.Rewind(uint(pathID))
}

func (a *rcConvVSAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.vs.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

type rcEllipseVSAdapter struct {
	ell *shapes.Ellipse
}

func (a *rcEllipseVSAdapter) Rewind(pathID uint32) { a.ell.Rewind(pathID) }
func (a *rcEllipseVSAdapter) Vertex(x, y *float64) uint32 {
	return uint32(a.ell.Vertex(x, y))
}

type rcEllipseConvAdapter struct {
	ell *shapes.Ellipse
}

func (a *rcEllipseConvAdapter) Rewind(pathID uint) {
	a.ell.Rewind(uint32(pathID))
}

func (a *rcEllipseConvAdapter) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = a.ell.Vertex(&x, &y)
	return x, y, cmd
}

func composeCompoundPath(ps *path.PathStorageStl) {
	ps.RemoveAll()
	ps.MoveTo(28.47, 6.45)
	ps.Curve3(21.58, 1.12, 19.82, 0.29)
	ps.Curve3(17.19, -0.93, 14.21, -0.93)
	ps.Curve3(9.57, -0.93, 6.57, 2.25)
	ps.Curve3(3.56, 5.42, 3.56, 10.60)
	ps.Curve3(3.56, 13.87, 5.03, 16.26)
	ps.Curve3(7.03, 19.58, 11.99, 22.51)
	ps.Curve3(16.94, 25.44, 28.47, 29.64)
	ps.LineTo(28.47, 31.40)
	ps.Curve3(28.47, 38.09, 26.34, 40.58)
	ps.Curve3(24.22, 43.07, 20.17, 43.07)
	ps.Curve3(17.09, 43.07, 15.28, 41.41)
	ps.Curve3(13.43, 39.75, 13.43, 37.60)
	ps.LineTo(13.53, 34.77)
	ps.Curve3(13.53, 32.52, 12.38, 31.30)
	ps.Curve3(11.23, 30.08, 9.38, 30.08)
	ps.Curve3(7.57, 30.08, 6.42, 31.35)
	ps.Curve3(5.27, 32.62, 5.27, 34.81)
	ps.Curve3(5.27, 39.01, 9.57, 42.53)
	ps.Curve3(13.87, 46.04, 21.63, 46.04)
	ps.Curve3(27.59, 46.04, 31.40, 44.04)
	ps.Curve3(34.28, 42.53, 35.64, 39.31)
	ps.Curve3(36.52, 37.21, 36.52, 30.71)
	ps.LineTo(36.52, 15.53)
	ps.Curve3(36.52, 9.13, 36.77, 7.69)
	ps.Curve3(37.01, 6.25, 37.57, 5.76)
	ps.Curve3(38.13, 5.27, 38.87, 5.27)
	ps.Curve3(39.65, 5.27, 40.23, 5.62)
	ps.Curve3(41.26, 6.25, 44.19, 9.18)
	ps.LineTo(44.19, 6.45)
	ps.Curve3(38.72, -0.88, 33.74, -0.88)
	ps.Curve3(31.35, -0.88, 29.93, 0.78)
	ps.Curve3(28.52, 2.44, 28.47, 6.45)
	ps.ClosePolygon(basics.PathFlagsNone)

	ps.MoveTo(28.47, 9.62)
	ps.LineTo(28.47, 26.66)
	ps.Curve3(21.09, 23.73, 18.95, 22.51)
	ps.Curve3(15.09, 20.36, 13.43, 18.02)
	ps.Curve3(11.77, 15.67, 11.77, 12.89)
	ps.Curve3(11.77, 9.38, 13.87, 7.06)
	ps.Curve3(15.97, 4.74, 18.70, 4.74)
	ps.Curve3(22.41, 4.74, 28.47, 9.62)
	ps.ClosePolygon(basics.PathFlagsNone)
}

func drawRasterizerCompoundDemo() {
	img := ctx.GetImage()
	w, h := img.Width(), img.Height()
	wf, hf := float64(w), float64(h)

	// Scale factors from C++ reference canvas (440×330, flip_y=true) to WASM canvas.
	scaleX := wf / compoundRefW
	scaleY := hf / compoundRefH

	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, w, h, img.Stride())
	pf := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]](pf)

	// Horizontal background gradient: yellow -> cyan.
	gradient := make([]color.RGBA8[color.Linear], w)
	for x := 0; x < w; x++ {
		t := float64(x) / float64(w-1)
		gradient[x] = color.RGBA8[color.Linear]{
			R: uint8((1.0 - t) * 255),
			G: 255,
			B: uint8(t * 255),
			A: 255,
		}
	}
	for y := 0; y < h; y++ {
		renBase.CopyColorHspan(0, y, w, gradient)
	}

	// Two background triangles matching C++ y-up orientation mapped to y-down screen.
	// C++ y-up triangle 1: (0,0)→(w,0)→(w,h)  →  y-down: (0,h)→(w,h)→(w,0)
	// C++ y-up triangle 2: (0,0)→(0,h)→(w,0)  →  y-down: (0,h)→(0,0)→(w,h)
	bgRas := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	bgSl := scanline.NewScanlineU8()

	bgPath := path.NewPathStorageStl()
	bgPath.MoveTo(0, hf)
	bgPath.LineTo(wf, hf)
	bgPath.LineTo(wf, 0)
	bgPath.ClosePolygon(basics.PathFlagsNone)
	bgRas.Reset()
	bgRas.AddPath(&rcConvVSAdapter{vs: path.NewPathStorageStlVertexSourceAdapter(bgPath)}, 0)
	renscan.RenderScanlinesAASolid(bgRas, bgSl, renBase, color.RGBA8[color.Linear]{R: 0, G: 100, B: 0, A: 255})

	bgPath2 := path.NewPathStorageStl()
	bgPath2.MoveTo(0, hf)
	bgPath2.LineTo(0, 0)
	bgPath2.LineTo(wf, hf)
	bgPath2.ClosePolygon(basics.PathFlagsNone)
	bgRas.Reset()
	bgRas.AddPath(&rcConvVSAdapter{vs: path.NewPathStorageStlVertexSourceAdapter(bgPath2)}, 0)
	renscan.RenderScanlinesAASolid(bgRas, bgSl, renBase, color.RGBA8[color.Linear]{R: 0, G: 100, B: 100, A: 255})

	// Compose and transform glyph path.
	// C++ y-up: scale(4) + translate(150, 100), canvas height 330.
	// In WASM y-down with canvas scaling:
	//   x' = 4*scaleX*x + 150*scaleX
	//   y' = hf - (4*scaleY*y + 100*scaleY) = -4*scaleY*y + (hf - 100*scaleY)
	ps := path.NewPathStorageStl()
	composeCompoundPath(ps)
	psAdapter := path.NewPathStorageStlVertexSourceAdapter(ps)

	mtx := transform.NewTransAffineFromValues(4*scaleX, 0, 0, -4*scaleY, 150*scaleX, hf-100*scaleY)
	transPath := conv.NewConvTransform(psAdapter, mtx)
	curve := conv.NewConvCurve(transPath)
	stroke := conv.NewConvStroke(curve)
	stroke.SetWidth(compoundWidth * scaleX)

	// Ellipse: C++ y-up center (220, 180), rx=120, ry=10.
	// In WASM y-down: center (220*scaleX, hf-180*scaleY), rx=120*scaleX, ry=10*scaleY.
	ell := shapes.NewEllipseWithParams(220.0*scaleX, hf-180.0*scaleY, 120.0*scaleX, 10.0*scaleY, 128, false)
	ellTrans := conv.NewConvTransform(&rcEllipseConvAdapter{ell: ell}, transform.NewTransAffine())
	ellStroke := conv.NewConvStroke(ellTrans)
	ellStroke.SetWidth(compoundWidth * 0.5 * scaleX)

	styles := []color.RGBA8[color.Linear]{
		{R: 0, G: 0, B: 255, A: 255},   // 0
		{R: 143, G: 90, B: 6, A: 255},  // 1
		{R: 51, G: 0, B: 151, A: 255},  // 2
		{R: 255, G: 0, B: 108, A: 255}, // 3
	}
	styles[3].Opacity(compoundAlpha1)
	styles[2].Opacity(compoundAlpha2)
	styles[1].Opacity(compoundAlpha3)
	styles[0].Opacity(compoundAlpha4)
	for i := range styles {
		styles[i].Premultiply()
	}

	// Compound AA rasterizer render loop.
	rasc := rasterizer.NewRasterizerCompoundAA(&compoundNoClip{})
	if compoundInvert {
		rasc.LayerOrder(basics.LayerInverse)
	} else {
		rasc.LayerOrder(basics.LayerDirect)
	}

	rasc.Styles(3, -1)
	rasc.AddPath(&rcConvVSAdapter{vs: ellStroke}, 0)
	rasc.Styles(2, -1)
	rasc.AddPath(&rcConvVSAdapter{vs: ellTrans}, 0)
	rasc.Styles(1, -1)
	rasc.AddPath(&rcConvVSAdapter{vs: stroke}, 0)
	rasc.Styles(0, -1)
	rasc.AddPath(&rcConvVSAdapter{vs: curve}, 0)

	rasc.Sort()
	if !rasc.RewindScanlines() {
		return
	}

	minX := rasc.MinX()
	maxX := rasc.MaxX()
	slAA := scanline.NewScanlineU8()
	slAA.Reset(minX, maxX)
	adAA := &rcSLAdapter{sl: slAA}
	styleHandler := &rcStyleHandler{styles: styles}

	length := maxX - minX + 2
	if length < 0 {
		length = 0
	}
	colorSpan := make([]color.RGBA8[color.Linear], length*2)
	mixBuffer := colorSpan[length:]
	coverBuffer := make([]basics.Int8u, length)

	for {
		numStyles := rasc.SweepStyles()
		if numStyles == 0 {
			break
		}
		if numStyles == 1 {
			if rasc.SweepScanline(adAA, 0) {
				c := styleHandler.Color(int(rasc.Style(0)))
				renscan.RenderScanlineAASolid(slAA, renBase, c)
			}
		} else {
			slStart := rasc.ScanlineStart()
			slLen := int(rasc.ScanlineLength())
			if slLen == 0 {
				continue
			}
			// Zero mix and cover buffers for this scanline extent.
			for i := 0; i < slLen; i++ {
				mixBuffer[slStart-minX+i] = color.RGBA8[color.Linear]{}
				coverBuffer[slStart-minX+i] = 0
			}
			var slY int
			for i := uint32(0); i < numStyles; i++ {
				style := int(rasc.Style(i))
				if rasc.SweepScanline(adAA, int(i)) {
					slY = slAA.Y()
					c := styleHandler.Color(style)
					for _, sp := range slAA.Spans() {
						for j := 0; j < int(sp.Len); j++ {
							idx := int(sp.X) - minX + j
							cover := sp.Covers[j]
							dst := coverBuffer[idx]
							if int(dst)+int(cover) > basics.CoverFull {
								cover = basics.Int8u(basics.CoverFull) - dst
							}
							if cover > 0 {
								mixBuffer[idx].AddWithCover(c, cover)
								coverBuffer[idx] += cover
							}
						}
					}
				}
			}
			renBase.BlendColorHspan(slStart, slY, slLen, mixBuffer[slStart-minX:slStart-minX+slLen], nil, basics.CoverFull)
		}
	}
}
