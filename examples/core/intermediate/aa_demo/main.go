// Port of AGG C++ aa_demo.cpp – anti-aliasing demonstration.
//
// Renders a triangle using the enlarged-pixel technique: each logical pixel
// in the rasterized triangle is drawn as a large square coloured by its AA
// coverage value, making the anti-aliasing algorithm visible.
// A slider controls the zoom factor (pixel size). The triangle vertices can
// be dragged with the mouse.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
	"github.com/MeKo-Christian/agg_go/internal/order"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt/blender"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
)

const (
	frameWidth  = 600
	frameHeight = 400
)

// pathSourceAdapter bridges PathStorageStl (uint Rewind) to the rasterizer
// VertexSource interface (uint32 Rewind + pointer-based Vertex).
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

// enlargedPixel stores a single collected pixel for deferred rendering.
type enlargedPixel struct {
	x, y float64
	col  agg.Color
}

// EnlargedRenderer implements renscan.RendererInterface.
// During the scanline sweep it collects one square per logical pixel;
// Flush then draws them all so the shared rasterizer is free for reuse.
type EnlargedRenderer struct {
	ctx       *agg.Context
	pixelSize float64
	col       agg.Color
	pixels    []enlargedPixel
}

func (r *EnlargedRenderer) Prepare() { r.pixels = r.pixels[:0] }

func (r *EnlargedRenderer) SetColor(c color.RGBA8[color.Linear]) {
	r.col = agg.NewColorRGBA8(c)
}

func (r *EnlargedRenderer) Render(sl renscan.ScanlineInterface) {
	y := sl.Y()
	it := sl.Begin()
	for i, n := 0, sl.NumSpans(); i < n; i++ {
		span := it.GetSpan()
		x := span.X
		numPix := span.Len
		covers := span.Covers
		if numPix < 0 {
			// Solid span: single cover value for all pixels.
			numPix = -numPix
			alpha := (uint16(covers[0]) * uint16(r.col.A)) >> 8
			c := agg.NewColor(r.col.R, r.col.G, r.col.B, uint8(alpha))
			for j := 0; j < numPix; j++ {
				r.pixels = append(r.pixels, enlargedPixel{
					x:   float64(x+j) * r.pixelSize,
					y:   float64(y) * r.pixelSize,
					col: c,
				})
			}
		} else {
			for j := 0; j < numPix; j++ {
				alpha := (uint16(covers[j]) * uint16(r.col.A)) >> 8
				r.pixels = append(r.pixels, enlargedPixel{
					x:   float64(x+j) * r.pixelSize,
					y:   float64(y) * r.pixelSize,
					col: agg.NewColor(r.col.R, r.col.G, r.col.B, uint8(alpha)),
				})
			}
		}
		if i < n-1 {
			it.Next()
		}
	}
}

// Flush draws all collected enlarged pixels to the canvas.
func (r *EnlargedRenderer) Flush() {
	for _, p := range r.pixels {
		r.ctx.SetColor(p.col)
		r.ctx.FillRectangle(p.x, p.y, r.pixelSize, r.pixelSize)
	}
}

// control is the interface required by renderControl.
type control interface {
	InRect(x, y float64) bool
	OnMouseButtonDown(x, y float64) bool
	OnMouseButtonUp(x, y float64) bool
	OnMouseMove(x, y float64, buttonPressed bool) bool
	NumPaths() uint
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
	Color(pathID uint) color.RGBA
}

type controlPathAdapter struct {
	rewindFn func(pathID uint)
	vertexFn func() (x, y float64, cmd basics.PathCommand)
}

func (a *controlPathAdapter) Rewind(pathID uint32) { a.rewindFn(uint(pathID)) }
func (a *controlPathAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.vertexFn()
	*x = vx
	*y = vy
	return uint32(cmd)
}

type rasScanlineAdapter struct{ sl *scanline.ScanlineU8 }

func (a *rasScanlineAdapter) ResetSpans()                 { a.sl.ResetSpans() }
func (a *rasScanlineAdapter) AddCell(x int, cover uint32) { a.sl.AddCell(x, uint(cover)) }
func (a *rasScanlineAdapter) AddSpan(x, length int, cover uint32) {
	a.sl.AddSpan(x, length, uint(cover))
}
func (a *rasScanlineAdapter) Finalize(y int) { a.sl.Finalize(y) }
func (a *rasScanlineAdapter) NumSpans() int  { return a.sl.NumSpans() }

func rgbaToRGBA8(c color.RGBA) color.RGBA8[color.Linear] {
	clamp := func(v float64) uint8 {
		if v <= 0 {
			return 0
		}
		if v >= 1 {
			return 255
		}
		return uint8(v*255 + 0.5)
	}
	return color.RGBA8[color.Linear]{R: clamp(c.R), G: clamp(c.G), B: clamp(c.B), A: clamp(c.A)}
}

func renderControl(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	sl *scanline.ScanlineU8,
	renBase *renderer.RendererBase[*pixfmt.PixFmtAlphaBlendRGBA[color.Linear, blender.BlenderRGBA8Pre[color.Linear, order.RGBA]], color.RGBA8[color.Linear]],
	ctrl control,
) {
	adapter := &controlPathAdapter{rewindFn: ctrl.Rewind, vertexFn: ctrl.Vertex}
	for pathID := uint(0); pathID < ctrl.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(adapter, uint32(pathID))
		col := rgbaToRGBA8(ctrl.Color(pathID))
		if !ras.RewindScanlines() {
			continue
		}
		sl.Reset(ras.MinX(), ras.MaxX())
		for ras.SweepScanline(&rasScanlineAdapter{sl: sl}) {
			y := sl.Y()
			for _, s := range sl.Spans() {
				if s.Len > 0 {
					renBase.BlendSolidHspan(int(s.X), y, int(s.Len), col, s.Covers)
				}
			}
		}
	}
}

// pointInTriangle reports whether (px, py) is inside the triangle.
func pointInTriangle(x1, y1, x2, y2, x3, y3, px, py float64) bool {
	cross := func(ax, ay, bx, by, cx, cy float64) float64 {
		return (ax-cx)*(by-cy) - (bx-cx)*(ay-cy)
	}
	b1 := cross(px, py, x1, y1, x2, y2) < 0
	b2 := cross(px, py, x2, y2, x3, y3) < 0
	b3 := cross(px, py, x3, y3, x1, y1) < 0
	return b1 == b2 && b2 == b3
}

type demo struct {
	x, y   [3]float64 // Triangle vertices
	dx, dy float64    // Mouse drag offset
	idx    int        // Dragged vertex index (3 = whole triangle, -1 = none)

	slider1 *slider.SliderCtrl
}

func newDemo() *demo {
	d := &demo{
		idx: -1,
		x:   [3]float64{57, 369, 143},
		y:   [3]float64{100, 170, 310},
	}
	// Slider at bottom of window, matching C++ m_slider1(80, 10, 600-10, 19, !flip_y)
	// with flip_y=true → y coords are from bottom → screen y = height - cpp_y.
	d.slider1 = slider.NewSliderCtrl(80, frameHeight-19, frameWidth-10, frameHeight-10, false)
	d.slider1.SetRange(8, 100)
	d.slider1.SetNumSteps(23)
	d.slider1.SetValue(32)
	d.slider1.SetLabel("Pixel size=%1.0f")
	return d
}

func (d *demo) Render(ctx *agg.Context) {
	ctx.Clear(agg.White)
	a := ctx.GetAgg2D()
	a.ResetTransformations()

	pixelSize := d.slider1.Value()

	// Build scaled triangle path (vertices / pixelSize).
	ps := path.NewPathStorageStl()
	ps.MoveTo(d.x[0]/pixelSize, d.y[0]/pixelSize)
	ps.LineTo(d.x[1]/pixelSize, d.y[1]/pixelSize)
	ps.LineTo(d.x[2]/pixelSize, d.y[2]/pixelSize)
	ps.ClosePolygon(basics.PathFlagsNone)

	// 1. Enlarged-pixel rendering: each rasterized pixel becomes a pixelSize×pixelSize square.
	enlargedRen := &EnlargedRenderer{ctx: ctx, pixelSize: pixelSize, col: agg.Black}
	ras := a.GetInternalRasterizer()
	ras.Reset()
	ras.AddPath(&pathSourceAdapter{ps: ps}, 0)
	a.ScanlineRender(ras, enlargedRen)
	enlargedRen.Flush()

	// 2. Actual-size solid black fill at the scaled coordinates
	//    (same as C++ render_scanlines_aa_solid after the enlarged pass).
	a.FillColor(agg.Black)
	a.NoLine()
	a.ResetPath()
	a.MoveTo(d.x[0]/pixelSize, d.y[0]/pixelSize)
	a.LineTo(d.x[1]/pixelSize, d.y[1]/pixelSize)
	a.LineTo(d.x[2]/pixelSize, d.y[2]/pixelSize)
	a.ClosePolygon()
	a.DrawPath(agg.FillOnly)

	// 3. Full-scale triangle outline in teal (matching C++ conv_stroke edges).
	teal := agg.NewColor(0, 150, 160, 200)
	a.LineColor(teal)
	a.NoFill()
	a.LineWidth(2.0)

	a.ResetPath()
	a.MoveTo(d.x[0], d.y[0])
	a.LineTo(d.x[1], d.y[1])
	a.DrawPath(agg.StrokeOnly)

	a.ResetPath()
	a.MoveTo(d.x[1], d.y[1])
	a.LineTo(d.x[2], d.y[2])
	a.DrawPath(agg.StrokeOnly)

	a.ResetPath()
	a.MoveTo(d.x[2], d.y[2])
	a.LineTo(d.x[0], d.y[0])
	a.DrawPath(agg.StrokeOnly)

	// 4. Render the slider control directly into the frame buffer.
	imgData := ctx.GetImage().Data
	rbuf := buffer.NewRenderingBufferU8WithData(imgData, frameWidth, frameHeight, frameWidth*4)
	pf := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt(pf)
	ctrlRas := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	ctrlSl := scanline.NewScanlineU8()
	renderControl(ctrlRas, ctrlSl, renBase, d.slider1)
}

func (d *demo) OnMouseDown(x, y int, btn demorunner.Buttons) bool {
	if !btn.Left {
		return false
	}
	fx, fy := float64(x), float64(y)

	if d.slider1.InRect(fx, fy) {
		return d.slider1.OnMouseButtonDown(fx, fy)
	}

	d.idx = -1
	for i := range 3 {
		dist := math.Sqrt((fx-d.x[i])*(fx-d.x[i]) + (fy-d.y[i])*(fy-d.y[i]))
		if dist < 10 {
			d.dx = fx - d.x[i]
			d.dy = fy - d.y[i]
			d.idx = i
			return true
		}
	}
	if pointInTriangle(d.x[0], d.y[0], d.x[1], d.y[1], d.x[2], d.y[2], fx, fy) {
		d.dx = fx - d.x[0]
		d.dy = fy - d.y[0]
		d.idx = 3
		return true
	}
	return false
}

func (d *demo) OnMouseMove(x, y int, btn demorunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	if d.slider1.OnMouseMove(fx, fy, btn.Left) {
		return true
	}
	if !btn.Left {
		d.idx = -1
		return false
	}
	if d.idx == 3 {
		dx := fx - d.dx
		dy := fy - d.dy
		d.x[1] -= d.x[0] - dx
		d.y[1] -= d.y[0] - dy
		d.x[2] -= d.x[0] - dx
		d.y[2] -= d.y[0] - dy
		d.x[0] = dx
		d.y[0] = dy
		return true
	}
	if d.idx >= 0 {
		d.x[d.idx] = fx - d.dx
		d.y[d.idx] = fy - d.dy
		return true
	}
	return false
}

func (d *demo) OnMouseUp(x, y int, btn demorunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	d.slider1.OnMouseButtonUp(fx, fy)
	d.idx = -1
	return false
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "AGG Example. Anti-Aliasing Demo",
		Width:  frameWidth,
		Height: frameHeight,
	}, newDemo())
}
