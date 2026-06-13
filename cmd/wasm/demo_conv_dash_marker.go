// Port of AGG C++ conv_dash_marker.cpp – dash/marker interactive demo.
package main

import (
	"math"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	"github.com/cwbudde/agg_go/internal/path"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/shapes"
	"github.com/cwbudde/agg_go/internal/vcgen"
)

// --- State ---

// dashFrameHeight is the native C++ frame height. Control points are stored in
// screen space (Y-down), so the original Y-up frame coordinates are flipped once
// via dashFrameHeight - y at init. The scene is rendered at native scale in the
// top-left of the canvas, matching the standalone conv_dash_marker port.
const dashFrameHeight = 330.0

var (
	// Control points in screen coordinates (Y-down), initialised by dashInit.
	dashPts [3]basics.PointD

	dashWidth   = 3.0 // m_width default
	dashSmooth  = 1.0 // m_smooth default (range 0.0–2.0)
	dashClosed  = false
	dashCap     = 0     // 0=butt, 1=square, 2=round
	dashEvenOdd = false // m_even_odd

	dashIdx  = -1
	dashDrag basics.PointD // grab offset (cursor − handle) captured on mouse-down
)

// dashInit seeds control points from the original C++ constructor values
// (m_x = {157, 469, 243}, m_y = {60, 170, 310} in a 500×330 Y-up frame). The
// native coordinates are used 1:1; only Y is flipped once to screen space so the
// scene and the browser's Y-down mouse coordinates share a single frame.
func dashInit() {
	flipY := func(x, y float64) basics.PointD {
		return basics.PointD{X: x, Y: dashFrameHeight - y}
	}

	dashPts[0] = flipY(157, 60)
	dashPts[1] = flipY(469, 170)
	dashPts[2] = flipY(243, 310)
}

func init() {
	dashInit()
}

// --- Path builder ---

// buildDashPath creates the two-sub-path storage matching the C++ on_draw path.
// Coordinates are in screen space (Y-down).
func buildDashPath() *path.PathStorageStl {
	p0, p1, p2 := dashPts[0], dashPts[1], dashPts[2]
	cx := (p0.X + p1.X + p2.X) / 3
	cy := (p0.Y + p1.Y + p2.Y) / 3

	ps := path.NewPathStorageStl()

	// Sub-path 1: P0 → P1 → centroid → P2
	ps.MoveTo(p0.X, p0.Y)
	ps.LineTo(p1.X, p1.Y)
	ps.LineTo(cx, cy)
	ps.LineTo(p2.X, p2.Y)
	if dashClosed {
		ps.ClosePolygon(basics.PathFlagsNone)
	}

	// Sub-path 2: mid01 → mid12 → mid20
	ps.MoveTo((p0.X+p1.X)/2, (p0.Y+p1.Y)/2)
	ps.LineTo((p1.X+p2.X)/2, (p1.Y+p2.Y)/2)
	ps.LineTo((p2.X+p0.X)/2, (p2.Y+p0.Y)/2)
	if dashClosed {
		ps.ClosePolygon(basics.PathFlagsNone)
	}

	return ps
}

// --- Vertex source adapters ---

// pathToConvSource adapts path.PathStorageStl (NextVertex uint32) → conv.VertexSource.
type pathToConvSource struct{ ps *path.PathStorageStl }

func (a *pathToConvSource) Rewind(pathID uint) { a.ps.Rewind(pathID) }
func (a *pathToConvSource) Vertex() (x, y float64, cmd basics.PathCommand) {
	vx, vy, c := a.ps.NextVertex()
	return vx, vy, basics.PathCommand(c)
}

// convToRasSource adapts conv.VertexSource → rasterizer.VertexSource.
type convToRasSource struct{ src conv.VertexSource }

func (a *convToRasSource) Rewind(pathID uint32) { a.src.Rewind(uint(pathID)) }
func (a *convToRasSource) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// arrowheadShapes adapts shapes.Arrowhead → conv.MarkerShapes.
type arrowheadShapes struct{ ah *shapes.Arrowhead }

func (a *arrowheadShapes) Rewind(shapeIndex uint) { a.ah.Rewind(uint32(shapeIndex)) }
func (a *arrowheadShapes) Vertex() (x, y float64, cmd basics.PathCommand) {
	var vx, vy float64
	c := a.ah.Vertex(&vx, &vy)
	return vx, vy, c
}

// --- Drawing ---

func drawDashDemo() {
	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Stride())
	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](pf)
	renBase.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	if dashEvenOdd {
		ras.FillingRule(basics.FillEvenOdd)
	} else {
		ras.FillingRule(basics.FillNonZero)
	}
	sl := scanline.NewScanlineU8()

	renderSolidColor := func(col color.RGBA8[color.Linear]) {
		if !ras.RewindScanlines() {
			return
		}
		sl.Reset(ras.MinX(), ras.MaxX())
		for ras.SweepScanline(sl) {
			y := sl.Y()
			for _, span := range sl.Spans() {
				if span.Len > 0 {
					renBase.BlendSolidHspan(int(span.X), y, int(span.Len), col, span.Covers)
				}
			}
		}
	}

	ps := buildDashPath()
	rawSrc := &pathToConvSource{ps: ps}

	// === Layer 1: raw fill (amber rgba(0.7, 0.5, 0.1, 0.5)) ===
	ras.Reset()
	ras.AddPath(&convToRasSource{src: rawSrc}, 0)
	renderSolidColor(color.RGBA8[color.Linear]{R: 178, G: 127, B: 25, A: 127})

	// === Layer 2: smooth poly fill (light blue rgba(0.1, 0.5, 0.7, 0.1)) ===
	smooth1 := conv.NewConvSmoothPoly1Curve(rawSrc)
	smooth1.SetSmoothValue(dashSmooth)
	ras.Reset()
	ras.AddPath(&convToRasSource{src: smooth1}, 0)
	renderSolidColor(color.RGBA8[color.Linear]{R: 25, G: 127, B: 178, A: 25})

	// === Layer 3: smooth poly stroke outline (green rgba(0.0, 0.6, 0.0, 0.8)) ===
	smooth2 := conv.NewConvSmoothPoly1(rawSrc)
	smooth2.SetSmoothValue(dashSmooth)
	smoothOutline := conv.NewConvStroke(smooth2)
	smoothOutline.SetWidth(1.0)
	ras.Reset()
	ras.AddPath(&convToRasSource{src: smoothOutline}, 0)
	renderSolidColor(color.RGBA8[color.Linear]{R: 0, G: 153, B: 0, A: 204})

	// === Layer 4: dashed smooth stroke + arrowhead markers (black) ===

	// Smooth + curve-flatten source for dashing
	curve := conv.NewConvSmoothPoly1Curve(rawSrc)
	curve.SetSmoothValue(dashSmooth)

	// Markers terminal (collects start/end positions for arrowhead placement)
	markers := vcgen.NewVCGenMarkersTerm()

	// Dash on smooth curve, feeding marker positions to markers terminal
	dash := conv.NewConvDashWithMarkers(curve, markers)
	dash.AddDash(20, 5)
	dash.AddDash(5, 5)
	dash.AddDash(5, 5)
	dash.DashStart(10)

	// Stroke the dash
	stroke := conv.NewConvStroke(dash)
	stroke.SetWidth(dashWidth)
	switch dashCap {
	case 1:
		stroke.SetLineCap(basics.SquareCap)
	case 2:
		stroke.SetLineCap(basics.RoundCap)
	default:
		stroke.SetLineCap(basics.ButtCap)
	}

	// Arrowhead geometry (k = pow(width, 0.7) as in C++)
	k := math.Pow(dashWidth, 0.7)
	ah := shapes.NewArrowhead()
	ah.Head(4*k, 4*k, 3*k, 2*k)
	if !dashClosed {
		ah.Tail(1*k, 1.5*k, 3*k, 5*k)
	}

	// ConvMarker places the arrowhead at each line endpoint recorded by markers.
	arrow := conv.NewConvMarker(markers, &arrowheadShapes{ah: ah})

	// Add stroked dash path and arrowhead markers to rasterizer.
	ras.Reset()
	ras.AddPath(&convToRasSource{src: stroke}, 0)
	ras.AddPath(&convToRasSource{src: arrow}, 0)
	renderSolidColor(color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})

	// === Handles ===
	for i := range 3 {
		drawHandle(dashPts[i].X, dashPts[i].Y)
	}

	applyLinearToSRGB(img)
}

// --- Mouse handlers ---

// dashPointInTriangle returns true if (px, py) is inside the triangle.
func dashPointInTriangle(ax, ay, bx, by, cx, cy, px, py float64) bool {
	d1 := (px-bx)*(ay-by) - (ax-bx)*(py-by)
	d2 := (px-cx)*(by-cy) - (bx-cx)*(py-cy)
	d3 := (px-ax)*(cy-ay) - (cx-ax)*(py-ay)
	hasNeg := (d1 < 0) || (d2 < 0) || (d3 < 0)
	hasPos := (d1 > 0) || (d2 > 0) || (d3 > 0)
	return !hasNeg || !hasPos
}

func handleDashMouseDown(x, y float64) bool {
	dashIdx = -1
	// Hit-test individual control points first (radius 20 px).
	for i := 0; i < 3; i++ {
		if math.Hypot(x-dashPts[i].X, y-dashPts[i].Y) < 20 {
			dashDrag = basics.PointD{X: x - dashPts[i].X, Y: y - dashPts[i].Y}
			dashIdx = i
			return true
		}
	}
	// Click inside the triangle → move all three points together.
	if dashPointInTriangle(dashPts[0].X, dashPts[0].Y, dashPts[1].X, dashPts[1].Y, dashPts[2].X, dashPts[2].Y, x, y) {
		dashDrag = basics.PointD{X: x - dashPts[0].X, Y: y - dashPts[0].Y}
		dashIdx = 3
		return true
	}
	return false
}

func handleDashMouseMove(x, y float64) bool {
	if dashIdx == 3 {
		// Move whole polygon: new position of P0 is cursor − grab offset.
		dx := x - dashDrag.X
		dy := y - dashDrag.Y
		dashPts[1].X -= dashPts[0].X - dx
		dashPts[1].Y -= dashPts[0].Y - dy
		dashPts[2].X -= dashPts[0].X - dx
		dashPts[2].Y -= dashPts[0].Y - dy
		dashPts[0].X = dx
		dashPts[0].Y = dy
		return true
	}
	if dashIdx >= 0 {
		dashPts[dashIdx] = basics.PointD{X: x - dashDrag.X, Y: y - dashDrag.Y}
		return true
	}
	return false
}

func handleDashMouseUp() {
	dashIdx = -1
}
