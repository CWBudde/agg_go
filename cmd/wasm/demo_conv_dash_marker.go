// Port of AGG C++ conv_dash_marker.cpp – dash/marker interactive demo.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
	"github.com/MeKo-Christian/agg_go/internal/vcgen"
)

// --- State ---

const (
	dashBaseWidth  = 500.0
	dashBaseHeight = 330.0
)

var (
	// Control points matching C++ constructor: m_x[i] = 57/369/143 + 100, m_y[i] = 60/170/310.
	dashX = [3]float64{157, 469, 243}
	dashY = [3]float64{60, 170, 310}

	dashWidth   = 3.0 // m_width default
	dashSmooth  = 1.0 // m_smooth default (range 0.0–2.0)
	dashClosed  = false
	dashCap     = 0     // 0=butt, 1=square, 2=round
	dashEvenOdd = false // m_even_odd

	dashIdx = -1
	dashDX  = 0.0
	dashDY  = 0.0
)

// --- Path builder ---

// buildDashPath creates the two-sub-path storage matching the C++ on_draw path.
func buildDashPath(mapX, mapY func(float64) float64) *path.PathStorageStl {
	cx := (dashX[0] + dashX[1] + dashX[2]) / 3
	cy := (dashY[0] + dashY[1] + dashY[2]) / 3

	ps := path.NewPathStorageStl()

	// Sub-path 1: P0 → P1 → centroid → P2
	ps.MoveTo(mapX(dashX[0]), mapY(dashY[0]))
	ps.LineTo(mapX(dashX[1]), mapY(dashY[1]))
	ps.LineTo(mapX(cx), mapY(cy))
	ps.LineTo(mapX(dashX[2]), mapY(dashY[2]))
	if dashClosed {
		ps.ClosePolygon(basics.PathFlagsNone)
	}

	// Sub-path 2: mid01 → mid12 → mid20
	ps.MoveTo(mapX((dashX[0]+dashX[1])/2), mapY((dashY[0]+dashY[1])/2))
	ps.LineTo(mapX((dashX[1]+dashX[2])/2), mapY((dashY[1]+dashY[2])/2))
	ps.LineTo(mapX((dashX[2]+dashX[0])/2), mapY((dashY[2]+dashY[0])/2))
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

// --- addPath helper: feeds a conv.VertexSource into agg2d via MoveTo/LineTo ---

func dashAddToPath(a *agg.Agg2D, src conv.VertexSource) {
	src.Rewind(0)
	for {
		x, y, cmd := src.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		switch {
		case basics.IsMoveTo(cmd):
			a.MoveTo(x, y)
		case basics.IsLineTo(cmd):
			a.LineTo(x, y)
		case basics.IsClosed(uint32(cmd)):
			a.ClosePolygon()
		}
	}
}

func fitDashFrame(w, h int) (scale, offX, offY float64) {
	sx := float64(w) / dashBaseWidth
	sy := float64(h) / dashBaseHeight
	scale = math.Min(sx, sy)
	if scale > 1.0 {
		scale = 1.0
	}
	if scale <= 0 {
		scale = 1.0
	}
	offX = (float64(w) - dashBaseWidth*scale) * 0.5
	offY = (float64(h) - dashBaseHeight*scale) * 0.5
	return scale, offX, offY
}

func dashMapPoint(scale, offX, offY, x, y float64) (float64, float64) {
	return offX + x*scale, offY + (dashBaseHeight-y)*scale
}

func dashUnmapPoint(scale, offX, offY, x, y float64) (float64, float64) {
	return (x - offX) / scale, dashBaseHeight - (y-offY)/scale
}

// --- Drawing ---

func drawDashDemo() {
	a := ctx.GetAgg2D()
	a.ResetTransformations()
	a.ClearAll(agg.White)

	scale, offX, offY := fitDashFrame(ctx.Width(), ctx.Height())
	mapX := func(x float64) float64 { return offX + x*scale }
	mapY := func(y float64) float64 { return offY + (dashBaseHeight-y)*scale }

	ps := buildDashPath(mapX, mapY)
	rawSrc := &pathToConvSource{ps: ps}

	a.FillEvenOdd(dashEvenOdd)

	// === Layer 1: raw fill (amber rgba(0.7, 0.5, 0.1, 0.5)) ===
	a.ResetPath()
	dashAddToPath(a, rawSrc)
	a.FillColor(agg.RGBA(0.7, 0.5, 0.1, 0.5))
	a.NoLine()
	a.DrawPath(agg.FillOnly)

	// === Layer 2: smooth poly fill (light blue rgba(0.1, 0.5, 0.7, 0.1)) ===
	smooth1 := conv.NewConvSmoothPoly1Curve(rawSrc)
	smooth1.SetSmoothValue(dashSmooth)
	a.ResetPath()
	dashAddToPath(a, smooth1)
	a.FillColor(agg.RGBA(0.1, 0.5, 0.7, 0.1))
	a.NoLine()
	a.DrawPath(agg.FillOnly)

	a.FillEvenOdd(false) // reset to non-zero for subsequent draws

	// === Layer 3: smooth poly stroke outline (green rgba(0.0, 0.6, 0.0, 0.8)) ===
	smooth2 := conv.NewConvSmoothPoly1Curve(rawSrc)
	smooth2.SetSmoothValue(dashSmooth)
	a.ResetPath()
	dashAddToPath(a, smooth2)
	a.LineColor(agg.RGBA(0, 0.6, 0, 0.8))
	a.LineWidth(max(1.0, scale))
	a.NoFill()
	a.DrawPath(agg.StrokeOnly)

	// === Layer 4: dashed smooth stroke + arrowhead markers (black) ===
	// Requires internal rasterizer to wire VCGenMarkersTerm → ConvMarker → Arrowhead.
	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Width()*4)
	pixFmt := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](pixFmt)
	sl := scanline.NewScanlineU8()
	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)

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
	ras.AddPath(&convToRasSource{src: stroke}, 0)
	ras.AddPath(&convToRasSource{src: arrow}, 0)

	black := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255}
	if ras.RewindScanlines() {
		sl.Reset(ras.MinX(), ras.MaxX())
		for ras.SweepScanline(&rasScanlineAdapter{sl: sl}) {
			y := sl.Y()
			for _, span := range sl.Spans() {
				if span.Len > 0 {
					renBase.BlendSolidHspan(int(span.X), y, int(span.Len), black, span.Covers)
				}
			}
		}
	}

	// === Handles ===
	for i := 0; i < 3; i++ {
		x, y := dashMapPoint(scale, offX, offY, dashX[i], dashY[i])
		drawHandle(x, y)
	}
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
	scale, offX, offY := fitDashFrame(ctx.Width(), ctx.Height())
	x, y = dashUnmapPoint(scale, offX, offY, x, y)
	dashIdx = -1
	// Hit-test individual control points first (radius 20 px, matching C++).
	for i := 0; i < 3; i++ {
		if math.Sqrt((x-dashX[i])*(x-dashX[i])+(y-dashY[i])*(y-dashY[i])) < 20 {
			dashDX = x - dashX[i]
			dashDY = y - dashY[i]
			dashIdx = i
			return true
		}
	}
	// Click inside the triangle → move all three points together.
	if dashPointInTriangle(dashX[0], dashY[0], dashX[1], dashY[1], dashX[2], dashY[2], x, y) {
		dashDX = x - dashX[0]
		dashDY = y - dashY[0]
		dashIdx = 3
		return true
	}
	return false
}

func handleDashMouseMove(x, y float64) bool {
	scale, offX, offY := fitDashFrame(ctx.Width(), ctx.Height())
	x, y = dashUnmapPoint(scale, offX, offY, x, y)
	if dashIdx == 3 {
		// Move whole polygon: new position of P0 is (x-dashDX, y-dashDY).
		dx := x - dashDX
		dy := y - dashDY
		dashX[1] -= dashX[0] - dx
		dashY[1] -= dashY[0] - dy
		dashX[2] -= dashX[0] - dx
		dashY[2] -= dashY[0] - dy
		dashX[0] = dx
		dashY[0] = dy
		return true
	}
	if dashIdx >= 0 {
		dashX[dashIdx] = x - dashDX
		dashY[dashIdx] = y - dashDY
		return true
	}
	return false
}

func handleDashMouseUp() {
	dashIdx = -1
}
