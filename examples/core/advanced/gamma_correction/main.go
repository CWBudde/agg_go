// Port of AGG C++ gamma_correction.cpp – anti-aliasing gamma demonstration.
//
// Shows how the anti-aliasing gamma affects thin ellipse rendering on a split
// dark/light background. Three slider controls adjust thickness, contrast and
// gamma. The image is rendered in a flipped work buffer and copied with
// y-flip, matching the C++ original's flip_y=true coordinate system.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	icolor "github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	sliderctrl "github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	isl "github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
)

const (
	frameWidth  = 400
	frameHeight = 320
)

// ---------------------------------------------------------------------------
// Rasterizer / scanline type alias
// ---------------------------------------------------------------------------

type rasType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

func newRasterizer() *rasType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
}

// ---------------------------------------------------------------------------
// Vertex source adapters
// ---------------------------------------------------------------------------

// ellipseVertexSource adapts shapes.Ellipse to conv.VertexSource.
type ellipseVertexSource struct {
	e *shapes.Ellipse
}

func (s *ellipseVertexSource) Rewind(id uint) { s.e.Rewind(uint32(id)) }
func (s *ellipseVertexSource) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = s.e.Vertex(&x, &y)
	return
}

// convVertexSourceAdapter wraps a conv.VertexSource into the rasterizer
// low-level interface (Rewind(uint32), Vertex(*x,*y) uint32).
type convRasAdapter struct{ src conv.VertexSource }

func (a *convRasAdapter) Rewind(id uint32) { a.src.Rewind(uint(id)) }
func (a *convRasAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// ctrlRasAdapter wraps a ctrl.Ctrl into the rasterizer low-level interface.
type ctrlRasAdapter struct{ ctrl ctrlbase.Ctrl[icolor.RGBA] }

func (a *ctrlRasAdapter) Rewind(id uint32) { a.ctrl.Rewind(uint(id)) }
func (a *ctrlRasAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// ---------------------------------------------------------------------------
// Rendering helpers
// ---------------------------------------------------------------------------

type renBase = renderer.RendererBase[*pixfmt.PixFmtRGBA32[icolor.Linear], icolor.RGBA8[icolor.Linear]]

func clampU8(v float64) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	return uint8(v*255.0 + 0.5)
}

func rgba8(r, g, b, a uint8) icolor.RGBA8[icolor.Linear] {
	return icolor.RGBA8[icolor.Linear]{R: r, G: g, B: b, A: a}
}

func copyFlipY(src, dst []uint8, w, h int) {
	stride := w * 4
	for y := 0; y < h; y++ {
		srcOff := (h - 1 - y) * stride
		dstOff := y * stride
		copy(dst[dstOff:dstOff+stride], src[srcOff:srcOff+stride])
	}
}

func renderCtrl(ras *rasType, sl *isl.ScanlineU8, rb *renBase, c ctrlbase.Ctrl[icolor.RGBA]) {
	for pathID := uint(0); pathID < c.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(&ctrlRasAdapter{ctrl: c}, uint32(pathID))
		col := c.Color(pathID)
		renscan.RenderScanlinesAASolid(ras, sl, rb, rgba8(clampU8(col.R), clampU8(col.G), clampU8(col.B), clampU8(col.A)))
	}
}

func renderSolid(ras *rasType, sl *isl.ScanlineU8, rb *renBase, col icolor.RGBA8[icolor.Linear]) {
	renscan.RenderScanlinesAASolid(ras, sl, rb, col)
}

// fillRect fills a rectangle in the work buffer directly (equivalent to copy_bar).
func fillRect(rb *renBase, x1, y1, x2, y2 int, col icolor.RGBA8[icolor.Linear]) {
	rb.CopyBar(x1, y1, x2, y2, col)
}

// renderStrokeEllipse strokes an ellipse.
func renderStrokeEllipse(ras *rasType, sl *isl.ScanlineU8, rb *renBase,
	cx, cy, rx, ry float64, steps uint32, strokeWidth float64, col icolor.RGBA8[icolor.Linear],
) {
	ell := shapes.NewEllipseWithParams(cx, cy, rx, ry, steps, false)
	stroke := conv.NewConvStroke(&ellipseVertexSource{e: ell})
	stroke.SetWidth(strokeWidth)
	ras.Reset()
	ras.AddPath(&convRasAdapter{src: stroke}, 0)
	renderSolid(ras, sl, rb, col)
}

// ---------------------------------------------------------------------------
// Demo
// ---------------------------------------------------------------------------

type demo struct {
	rx, ry    float64
	thickness *sliderctrl.SliderCtrl
	contrast  *sliderctrl.SliderCtrl
	gamma     *sliderctrl.SliderCtrl
}

func newDemo() *demo {
	// C++ slider positions (flip_y=true, so !flip_y=false):
	//   m_thickness(5, 5,    400-5, 11,    !flip_y)  → y: 5..11
	//   m_contrast (5, 5+15, 400-5, 11+15, !flip_y)  → y: 20..26
	//   m_gamma    (5, 5+30, 400-5, 11+30, !flip_y)  → y: 35..41
	// In work buffer (y=0 at bottom), these appear near the bottom.
	// After copyFlipY they appear near the bottom of the displayed image.
	thickness := sliderctrl.NewSliderCtrl(5, 5, frameWidth-5, 11, false)
	thickness.SetRange(0.0, 3.0)
	thickness.SetValue(1.0)
	thickness.SetLabel("Thickness=%3.2f")

	contrast := sliderctrl.NewSliderCtrl(5, 5+15, frameWidth-5, 11+15, false)
	contrast.SetRange(0.0, 1.0)
	contrast.SetValue(1.0)
	contrast.SetLabel("Contrast")

	gamma := sliderctrl.NewSliderCtrl(5, 5+30, frameWidth-5, 11+30, false)
	gamma.SetRange(0.5, 3.0)
	gamma.SetValue(1.0)
	gamma.SetLabel("Gamma=%3.2f")

	return &demo{
		rx:        float64(frameWidth) / 3.0,
		ry:        float64(frameHeight) / 3.0,
		thickness: thickness,
		contrast:  contrast,
		gamma:     gamma,
	}
}

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()

	// Work buffer: y=0 at bottom (flip_y=true convention), copied with y-flip.
	workBuf := make([]uint8, w*h*4)
	workRbuf := buffer.NewRenderingBufferU8WithData(workBuf, w, h, w*4)
	pf := pixfmt.NewPixFmtRGBA32Linear(workRbuf)
	rb := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[icolor.Linear], icolor.RGBA8[icolor.Linear]](pf)
	rb.Clear(rgba8(255, 255, 255, 255))

	thickness := d.thickness.Value()
	contrast := d.contrast.Value()
	g := d.gamma.Value()

	dark := contrast
	light := contrast
	darkVal := uint8((1.0 - dark) * 255.0)
	lightVal := uint8(light * 255.0)

	// Background (equivalent to C++ copy_bar calls):
	// Left half: dark grey
	fillRect(rb, 0, 0, w/2, h, rgba8(darkVal, darkVal, darkVal, 255))
	// Right half: light grey
	fillRect(rb, w/2+1, 0, w, h, rgba8(lightVal, lightVal, lightVal, 255))
	// Bottom half (in work-buffer coords, y > h/2 means upper visual area after flip):
	// C++: copy_bar(0, height/2+1, width, height, rgba(1.0, dark, dark))
	fillRect(rb, 0, h/2+1, w, h, rgba8(255, darkVal, darkVal, 255))

	ras := newRasterizer()
	sl := isl.NewScanlineU8()

	// Gamma power curve (in work-buffer: y=50 at bottom area, going up).
	// C++: x=(width-256)/2, y=50; for i: path.line_to(x+i, y + gp(v)*255)
	{
		ps := path.NewPathStorage()
		xStart := float64(w-256) / 2.0
		yStart := 50.0
		gp := func(v float64) float64 { return math.Pow(v, g) }
		for i := 0; i <= 255; i++ {
			v := float64(i) / 255.0
			gval := gp(v)
			px := xStart + float64(i)
			py := yStart + gval*255.0
			if i == 0 {
				ps.MoveTo(px, py)
			} else {
				ps.LineTo(px, py)
			}
		}
		stroke := conv.NewConvStroke(path.NewPathStorageVertexSourceAdapter(ps))
		stroke.SetWidth(2.0)
		ras.Reset()
		ras.AddPath(&convRasAdapter{src: stroke}, 0)
		renderSolid(ras, sl, rb, rgba8(80, 127, 80, 255))
	}

	cx := float64(w) / 2.0
	cy := float64(h) / 2.0

	renderStrokeEllipse(ras, sl, rb, cx, cy, d.rx, d.ry, 150, thickness, rgba8(255, 0, 0, 255))
	renderStrokeEllipse(ras, sl, rb, cx, cy, d.rx-5, d.ry-5, 150, thickness, rgba8(0, 255, 0, 255))
	renderStrokeEllipse(ras, sl, rb, cx, cy, d.rx-10, d.ry-10, 150, thickness, rgba8(0, 0, 255, 255))
	renderStrokeEllipse(ras, sl, rb, cx, cy, d.rx-15, d.ry-15, 150, thickness, rgba8(0, 0, 0, 255))
	renderStrokeEllipse(ras, sl, rb, cx, cy, d.rx-20, d.ry-20, 150, thickness, rgba8(255, 255, 255, 255))

	renderCtrl(ras, sl, rb, d.thickness)
	renderCtrl(ras, sl, rb, d.contrast)
	renderCtrl(ras, sl, rb, d.gamma)

	copyFlipY(workBuf, img.Data, w, h)
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if !btn.Left {
		return false
	}
	fx, fy := float64(x), float64(frameHeight-y)
	d.thickness.OnMouseButtonDown(fx, fy)
	d.contrast.OnMouseButtonDown(fx, fy)
	d.gamma.OnMouseButtonDown(fx, fy)
	// C++ always updates ellipse size on mouse down.
	d.rx = math.Abs(float64(frameWidth)/2 - float64(x))
	d.ry = math.Abs(float64(frameHeight)/2 - float64(y))
	return true
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(frameHeight-y)
	changed := d.thickness.OnMouseMove(fx, fy, btn.Left)
	if d.contrast.OnMouseMove(fx, fy, btn.Left) {
		changed = true
	}
	if d.gamma.OnMouseMove(fx, fy, btn.Left) {
		changed = true
	}
	if btn.Left {
		// C++ on_mouse_move calls on_mouse_button_down which always updates rx/ry.
		d.rx = math.Abs(float64(frameWidth)/2 - float64(x))
		d.ry = math.Abs(float64(frameHeight)/2 - float64(y))
		return true
	}
	return changed
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(frameHeight-y)
	changed := d.thickness.OnMouseButtonUp(fx, fy)
	if d.contrast.OnMouseButtonUp(fx, fy) {
		changed = true
	}
	if d.gamma.OnMouseButtonUp(fx, fy) {
		changed = true
	}
	return changed
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "AGG Example. Thin red ellipse",
		Width:  frameWidth,
		Height: frameHeight,
	}, newDemo())
}
