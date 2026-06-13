// Port of AGG C++ blur.cpp (flip_y = true).
//
// Renders an "a" glyph shape, applies stack/recursive blur as a shadow effect,
// then draws the shape on top.  A polygon control lets the shadow be
// perspective-distorted; slider and radio-button controls mirror the C++
// original.
//
// Rendering is done in a work buffer (y=0 at bottom, C++ coordinate frame)
// and copied with a y-flip into the output image, matching the
// flip_y=true platform_support convention.
package main

import (
	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/examples/shared/lowlevelrunner"
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	ctrlbase "github.com/cwbudde/agg_go/internal/ctrl"
	checkboxctrl "github.com/cwbudde/agg_go/internal/ctrl/checkbox"
	polygonctrl "github.com/cwbudde/agg_go/internal/ctrl/polygon"
	rboxctrl "github.com/cwbudde/agg_go/internal/ctrl/rbox"
	sliderctrl "github.com/cwbudde/agg_go/internal/ctrl/slider"
	"github.com/cwbudde/agg_go/internal/effects"
	"github.com/cwbudde/agg_go/internal/path"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/transform"
)

const (
	frameWidth  = 440
	frameHeight = 330
)

// Default control values matching the C++ demo.
const (
	defaultMethod   = 0 // 0=Stack blur, 1=Recursive blur, 2=Channels
	defaultRadius   = 15.0
	defaultChannelR = false
	defaultChannelG = true
	defaultChannelB = false
)

// ---------------------------------------------------------------------------
// Rasterizer type alias
// ---------------------------------------------------------------------------

type rasType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

func newRasterizer() *rasType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
}

// ---------------------------------------------------------------------------
// Vertex-source adapters
// ---------------------------------------------------------------------------

// pathStlVS adapts PathStorageStl to conv.VertexSource (Rewind/Vertex iterator style).
type pathStlVS struct{ ps *path.PathStorageStl }

func (v *pathStlVS) Rewind(id uint) { v.ps.Rewind(id) }
func (v *pathStlVS) Vertex() (x, y float64, cmd basics.PathCommand) {
	vx, vy, c := v.ps.NextVertex()
	return vx, vy, basics.PathCommand(c)
}

// convSourceVS bridges a conv.VertexSource to the rasterizer VertexSource
// interface (Rewind(uint32) / Vertex(*x,*y) uint32).
type convSourceVS struct{ src conv.VertexSource }

func (v *convSourceVS) Rewind(id uint32) { v.src.Rewind(uint(id)) }
func (v *convSourceVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := v.src.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// perspectiveVS applies a TransPerspective to a conv.VertexSource.
type perspectiveVS struct {
	src   conv.VertexSource
	persp *transform.TransPerspective
}

func (v *perspectiveVS) Rewind(id uint32) { v.src.Rewind(uint(id)) }
func (v *perspectiveVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := v.src.Vertex()
	*x, *y = vx, vy
	if basics.IsVertex(basics.PathCommand(cmd)) {
		v.persp.Transform(x, y)
	}
	return uint32(cmd)
}

// boundsVS adapts a conv.VertexSource to basics.VertexSource (for BoundingRectSingle).
type boundsVS struct{ src conv.VertexSource }

func (v *boundsVS) Rewind(id uint) { v.src.Rewind(id) }
func (v *boundsVS) Vertex() (x, y float64, cmd basics.PathCommand) {
	vx, vy, c := v.src.Vertex()
	return vx, vy, basics.PathCommand(c)
}

// ctrlVS adapts a ctrl widget to the rasterizer VertexSource interface.
type ctrlVS struct {
	ctrl ctrlbase.Ctrl[color.RGBA]
}

func (v *ctrlVS) Rewind(id uint32) { v.ctrl.Rewind(uint(id)) }
func (v *ctrlVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := v.ctrl.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// ---------------------------------------------------------------------------
// "a" glyph path (in glyph-local coordinates, before any transform)
// ---------------------------------------------------------------------------

func buildGlyphPath() *path.PathStorageStl {
	ps := path.NewPathStorageStl()

	// Outer contour
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
	ps.ClosePolygon(basics.PathFlagsCW)

	// Inner contour
	ps.MoveTo(28.47, 9.62)
	ps.LineTo(28.47, 26.66)
	ps.Curve3(21.09, 23.73, 18.95, 22.51)
	ps.Curve3(15.09, 20.36, 13.43, 18.02)
	ps.Curve3(11.77, 15.67, 11.77, 12.89)
	ps.Curve3(11.77, 9.38, 13.87, 7.06)
	ps.Curve3(15.97, 4.74, 18.70, 4.74)
	ps.Curve3(22.41, 4.74, 28.47, 9.62)
	ps.ClosePolygon(basics.PathFlagsCW)

	return ps
}

// ---------------------------------------------------------------------------
// Rendering helper
// ---------------------------------------------------------------------------

func renderPath(
	ras *rasType,
	sl *scanline.ScanlineP8,
	renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]],
	vs interface {
		Rewind(uint32)
		Vertex(*float64, *float64) uint32
	},
	fillColor color.RGBA8[color.Linear],
) {
	ras.Reset()
	ras.AddPath(vs, 0)
	renscan.RenderScanlinesAASolid(ras, sl, renBase, fillColor)
}

// ---------------------------------------------------------------------------
// Sub-region blur helpers
// ---------------------------------------------------------------------------

// clampInt clamps v to [lo, hi].
func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// blurSubRegion blurs the inclusive rectangle [x0,y0]..[x1,y1] in the work
// buffer. The work buffer uses positive stride (row 0 = top in y-up frame =
// bottom of screen).
//
// The bounds handling mirrors C++ pixfmt::attach: the rect is clipped to
// [0,w-1]x[0,h-1] and the attached sub-image covers (x2-x1)+1 by (y2-y1)+1
// pixels, i.e. x1/y1 are inclusive — not the exclusive upper bound a Go slice
// range would imply.
func blurSubRegion(
	workBuf []uint8,
	w, h int,
	x0, y0, x1, y1 int,
	radius float64,
	method int,
) {
	if radius <= 0 {
		return
	}
	x0 = clampInt(x0, 0, w-1)
	y0 = clampInt(y0, 0, h-1)
	x1 = clampInt(x1, 0, w-1)
	y1 = clampInt(y1, 0, h-1)
	if x0 > x1 || y0 > y1 {
		return
	}

	stride := w * 4
	rw := x1 - x0 + 1
	rh := y1 - y0 + 1

	// Extract sub-region into a 2D pixel slice.
	pixels := make([][]color.RGBA8[color.Linear], rh)
	for row := range rh {
		pixels[row] = make([]color.RGBA8[color.Linear], rw)
		for col := range rw {
			idx := (y0+row)*stride + (x0+col)*4
			pixels[row][col] = color.RGBA8[color.Linear]{
				R: workBuf[idx],
				G: workBuf[idx+1],
				B: workBuf[idx+2],
				A: workBuf[idx+3],
			}
		}
	}

	r := int(radius)
	if method == 0 {
		sb := effects.NewSimpleStackBlur()
		sb.Blur(pixels, r)
	} else {
		rb := effects.NewSimpleRecursiveBlur()
		rb.BlurHorizontal(pixels, radius)
		pixels = transposePixels(pixels)
		rb.BlurHorizontal(pixels, radius)
		pixels = transposePixels(pixels)
	}

	// Write blurred pixels back.
	for row := range rh {
		for col := range rw {
			idx := (y0+row)*stride + (x0+col)*4
			pix := pixels[row][col]
			workBuf[idx] = pix.R
			workBuf[idx+1] = pix.G
			workBuf[idx+2] = pix.B
			workBuf[idx+3] = pix.A
		}
	}
}

func transposePixels(pixels [][]color.RGBA8[color.Linear]) [][]color.RGBA8[color.Linear] {
	if len(pixels) == 0 {
		return pixels
	}
	h := len(pixels)
	w := len(pixels[0])
	out := make([][]color.RGBA8[color.Linear], w)
	for col := range w {
		out[col] = make([]color.RGBA8[color.Linear], h)
		for row := range h {
			out[col][row] = pixels[row][col]
		}
	}
	return out
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

// ---------------------------------------------------------------------------
// Control rendering helper
// ---------------------------------------------------------------------------

func renderCtrl(
	ras *rasType,
	sl *scanline.ScanlineP8,
	rb *renderer.RendererBase[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]],
	ctrl ctrlbase.Ctrl[color.RGBA],
) {
	clamp := func(v float64) uint8 {
		if v <= 0 {
			return 0
		}
		if v >= 1 {
			return 255
		}
		return uint8(v*255.0 + 0.5)
	}
	toLinear := func(c color.RGBA) color.RGBA8[color.Linear] {
		return color.RGBA8[color.Linear]{
			R: clamp(c.R), G: clamp(c.G), B: clamp(c.B), A: clamp(c.A),
		}
	}

	vs := &ctrlVS{ctrl: ctrl}
	for i := range ctrl.NumPaths() {
		ras.Reset()
		ras.AddPath(vs, uint32(i))
		renscan.RenderScanlinesAASolid(ras, sl, rb, toLinear(ctrl.Color(i)))
	}
}

// ---------------------------------------------------------------------------
// Demo
// ---------------------------------------------------------------------------

type demo struct{}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()

	// Work buffer: positive stride, y=0 at bottom (C++ y-up frame).
	workBuf := make([]uint8, w*h*4)
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(workBuf, w, h, w*4)
	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt(pf)
	renBase.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := newRasterizer()
	sl := scanline.NewScanlineP8()

	// ---------------------------------------------------------------------------
	// Build the "a" shape with C++ transform: scale(4.0) + translate(150, 100).
	// ---------------------------------------------------------------------------
	shapeMtx := transform.NewTransAffineScaling(4.0)
	shapeMtx.Multiply(transform.NewTransAffineTranslation(150, 100))

	glyphPS := buildGlyphPath()
	glyphXformed := conv.NewConvTransform(&pathStlVS{ps: glyphPS}, shapeMtx)
	glyphCurved := conv.NewConvCurve(glyphXformed)

	// Compute shape bounding rectangle (in work-buffer coords = C++ y-up frame).
	shapeBounds, ok := basics.BoundingRectSingle[float64](&boundsVS{src: glyphCurved}, 0)
	if !ok {
		// Fallback: use expected approximate bounds.
		shapeBounds = basics.Rect[float64]{X1: 164, Y1: 96, X2: 327, Y2: 284}
	}

	// ---------------------------------------------------------------------------
	// Shadow polygon control – default positions offset shape bounds by (+10, -10).
	// In y-up space, -10 in y means the shadow drops below the glyph visually.
	// ---------------------------------------------------------------------------
	shadowCtrl := polygonctrl.NewPolygonCtrl[color.RGBA](
		4, 5.0,
		color.NewRGBA(0, 0.3, 0.5, 0.3),
	)
	shadowCtrl.SetXn(0, shapeBounds.X1+10)
	shadowCtrl.SetYn(0, shapeBounds.Y1-10)
	shadowCtrl.SetXn(1, shapeBounds.X2+10)
	shadowCtrl.SetYn(1, shapeBounds.Y1-10)
	shadowCtrl.SetXn(2, shapeBounds.X2+10)
	shadowCtrl.SetYn(2, shapeBounds.Y2-10)
	shadowCtrl.SetXn(3, shapeBounds.X1+10)
	shadowCtrl.SetYn(3, shapeBounds.Y2-10)

	// ---------------------------------------------------------------------------
	// Method radio-button control (10,10 – 130,70, flipY=false = raw buffer y-up).
	// C++: m_method(10, 10, 130, 70, !flip_y=false)
	// ---------------------------------------------------------------------------
	methodCtrl := rboxctrl.NewDefaultRboxCtrl(10, 10, 130, 70, false)
	methodCtrl.SetTextSize(8, 0)
	methodCtrl.AddItem("Stack blur")
	methodCtrl.AddItem("Recursive blur")
	methodCtrl.AddItem("Channels")
	methodCtrl.SetCurItem(defaultMethod)

	// ---------------------------------------------------------------------------
	// Radius slider (140,14 – 430,22, flipY=false).
	// C++: m_radius(130+10, 10+4, 130+300, 10+8+4, !flip_y=false)
	// x2 = 130 + 300 = 430 (leaves a 10px right margin), NOT the window width
	// 440. Using 440 stretches the track and shifts the value knob right.
	// ---------------------------------------------------------------------------
	radiusCtrl := sliderctrl.NewSliderCtrl(140, 14, 430, 22, false)
	radiusCtrl.SetRange(0.0, 40.0)
	radiusCtrl.SetValue(defaultRadius)
	radiusCtrl.SetLabel("Blur Radius=%1.2f")

	// ---------------------------------------------------------------------------
	// Channel checkboxes (10,80 / 10,95 / 10,110, flipY=false).
	// C++: m_channel_r(10, 80, "Red", !flip_y=false)
	//      m_channel_g(10, 95, "Green", !flip_y=false)   ← default checked
	//      m_channel_b(10, 110, "Blue", !flip_y=false)
	// ---------------------------------------------------------------------------
	inactiveCol := color.NewRGBA(0.0, 0.0, 0.0, 1.0)
	textCol := color.NewRGBA(0.0, 0.0, 0.0, 1.0)
	activeCol := color.NewRGBA(0.4, 0.0, 0.0, 1.0)

	chanR := checkboxctrl.NewCheckboxCtrl[color.RGBA](10, 80, "Red", false, inactiveCol, textCol, activeCol)
	chanR.SetChecked(defaultChannelR)
	chanG := checkboxctrl.NewCheckboxCtrl[color.RGBA](10, 95, "Green", false, inactiveCol, textCol, activeCol)
	chanG.SetChecked(defaultChannelG)
	chanB := checkboxctrl.NewCheckboxCtrl[color.RGBA](10, 110, "Blue", false, inactiveCol, textCol, activeCol)
	chanB.SetChecked(defaultChannelB)

	// ---------------------------------------------------------------------------
	// 1. Render shadow via perspective transform of the shape onto the shadow polygon.
	// ---------------------------------------------------------------------------
	quad := [8]float64{
		shadowCtrl.Xn(0), shadowCtrl.Yn(0),
		shadowCtrl.Xn(1), shadowCtrl.Yn(1),
		shadowCtrl.Xn(2), shadowCtrl.Yn(2),
		shadowCtrl.Xn(3), shadowCtrl.Yn(3),
	}
	shadowPersp := transform.NewTransPerspectiveRectToQuad(
		shapeBounds.X1, shapeBounds.Y1,
		shapeBounds.X2, shapeBounds.Y2,
		quad,
	)

	// Re-build the curved path for shadow rendering (bounding rect consumed the iterator).
	glyphPS2 := buildGlyphPath()
	glyphXformed2 := conv.NewConvTransform(&pathStlVS{ps: glyphPS2}, shapeMtx)
	glyphCurved2 := conv.NewConvCurve(glyphXformed2)

	shadowVS := &perspectiveVS{src: glyphCurved2, persp: shadowPersp}
	// C++: agg::rgba(0.1, 0.1, 0.1) → rgba8 via uround(0.1*255)=uround(25.5)=26.
	renderPath(ras, sl, renBase, shadowVS,
		color.RGBA8[color.Linear]{R: 26, G: 26, B: 26, A: 255})

	// ---------------------------------------------------------------------------
	// 2. Compute shadow bounding box, expand by radius, and blur that sub-region.
	// ---------------------------------------------------------------------------
	// Re-build shadow vertex source for bounding rect computation.
	glyphPS3 := buildGlyphPath()
	glyphXformed3 := conv.NewConvTransform(&pathStlVS{ps: glyphPS3}, shapeMtx)
	glyphCurved3 := conv.NewConvCurve(glyphXformed3)
	shadowVS3 := &boundsVS{src: &perspectiveBoundsVS{src: glyphCurved3, persp: shadowPersp}}

	bboxR, validBbox := basics.BoundingRectSingle[float64](shadowVS3, 0)
	if validBbox {
		r := defaultRadius
		// Extend by radius (extra on x2/y2 side for recursive blur edge effects, matching C++).
		bx0 := int(bboxR.X1 - r)
		by0 := int(bboxR.Y1 - r)
		bx1 := int(bboxR.X2 + r*2)
		by1 := int(bboxR.Y2 + r*2)
		blurSubRegion(workBuf, w, h, bx0, by0, bx1, by1, r, defaultMethod)
	}

	// ---------------------------------------------------------------------------
	// 3. Render the actual shape on top of the blurred shadow.
	// ---------------------------------------------------------------------------
	glyphPS4 := buildGlyphPath()
	glyphXformed4 := conv.NewConvTransform(&pathStlVS{ps: glyphPS4}, shapeMtx)
	glyphCurved4 := conv.NewConvCurve(glyphXformed4)

	renderPath(ras, sl, renBase, &convSourceVS{src: glyphCurved4},
		color.RGBA8[color.Linear]{R: 153, G: 230, B: 179, A: 204})

	// ---------------------------------------------------------------------------
	// 4. Render controls.
	// ---------------------------------------------------------------------------
	renderCtrl(ras, sl, renBase, shadowCtrl)
	renderCtrl(ras, sl, renBase, methodCtrl)
	renderCtrl(ras, sl, renBase, radiusCtrl)
	renderCtrl(ras, sl, renBase, chanR)
	renderCtrl(ras, sl, renBase, chanG)
	renderCtrl(ras, sl, renBase, chanB)

	// ---------------------------------------------------------------------------
	// 5. Copy work buffer to output with y-flip (flip_y=true convention).
	// ---------------------------------------------------------------------------
	copyFlipY(workBuf, img.Data, w, h)
}

// perspectiveBoundsVS wraps a conv.VertexSource and applies TransPerspective
// so that BoundingRectSingle sees the post-perspective coordinates.
type perspectiveBoundsVS struct {
	src   conv.VertexSource
	persp *transform.TransPerspective
}

func (v *perspectiveBoundsVS) Rewind(id uint) { v.src.Rewind(id) }
func (v *perspectiveBoundsVS) Vertex() (x, y float64, cmd basics.PathCommand) {
	vx, vy, c := v.src.Vertex()
	if basics.IsVertex(c) {
		v.persp.Transform(&vx, &vy)
	}
	return vx, vy, c
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:                 "Blur",
		Width:                 frameWidth,
		Height:                frameHeight,
		EncodeLinearRGBToSRGB: true,
	}, &demo{})
}
