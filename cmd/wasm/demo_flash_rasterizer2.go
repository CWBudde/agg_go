// Port of AGG's flash_rasterizer2.cpp.
//
// Alternative Flash rasterization method: decomposes a compound shape into
// separate sub-shapes per fill style.  For each style index, paths whose
// left-fill matches are added forward; paths whose right-fill matches are
// added reversed (inverted polygon winding).  A clipping rasterizer is used
// so the spurious edge from the clipper origin is safely discarded.
//
// Controls (HTML/URL):
//
//	fr2Shape  (0–23): which shape frame to display
package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	agg "agg_go"
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/conv"
	"agg_go/internal/demo/shapesdata"
	"agg_go/internal/gsv"
	"agg_go/internal/pixfmt"
	"agg_go/internal/rasterizer"
	"agg_go/internal/renderer"
	renscan "agg_go/internal/renderer/scanline"
	"agg_go/internal/scanline"
)

// --- State ---

var (
	flash2ShapeIdx = 0
	flash2Zoom     = 1.0  // zoom factor (centered on flash2ZoomX/Y)
	flash2ZoomX    = 0.0  // zoom center X (canvas coords)
	flash2ZoomY    = 0.0  // zoom center Y (canvas coords)

	flash2Shapes []shapesdata.RawShape
	flash2Colors []color.RGBA8[color.Linear] // 100 random colours
)

func setFlash2ShapeIdx(v int) { flash2ShapeIdx = v }

// applyFlash2Wheel applies mouse-wheel zoom centered at (mx, my).
// deltaY > 0 means zoom out, < 0 means zoom in (standard browser convention).
func applyFlash2Wheel(mx, my, deltaY float64) {
	factor := 1.0 / 1.1
	if deltaY < 0 {
		factor = 1.1
	}
	// Translate so (mx,my) is origin, scale, translate back.
	flash2ZoomX = mx - (mx-flash2ZoomX)*factor
	flash2ZoomY = my - (my-flash2ZoomY)*factor
	flash2Zoom *= factor
}

// --- Initialisation ---

func initFlash2() {
	if flash2Shapes != nil {
		return
	}
	flash2Shapes = shapesdata.LoadShapes()

	rng := rand.New(rand.NewSource(42))
	flash2Colors = make([]color.RGBA8[color.Linear], 100)
	for i := range flash2Colors {
		flash2Colors[i] = color.RGBA8[color.Linear]{
			R: uint8(rng.Intn(256)),
			G: uint8(rng.Intn(256)),
			B: uint8(rng.Intn(256)),
			A: 230,
		}
		flash2Colors[i].Premultiply() // C++ does srgba8(...).premultiply()
	}
}

// --- Rendering ---

func drawFlashRasterizer2Demo() {
	initFlash2()

	ctx.GetAgg2D().ResetTransformations()

	if len(flash2Shapes) == 0 {
		return
	}
	idx := flash2ShapeIdx
	if idx < 0 {
		idx = 0
	}
	if idx >= len(flash2Shapes) {
		idx = len(flash2Shapes) - 1
	}
	shape := &flash2Shapes[idx]

	if len(shape.Paths) == 0 {
		return
	}

	// Viewport: fit shape bounding rect into canvas with preserve-aspect-ratio = meet (centred).
	bx1, by1, bx2, by2 := shape.BoundingRect()
	worldW := bx2 - bx1
	worldH := by2 - by1
	if worldW <= 0 || worldH <= 0 {
		return
	}
	cW := float64(width)
	cH := float64(height)
	scaleX := cW / worldW
	scaleY := cH / worldH
	sc := scaleX
	if scaleY < sc {
		sc = scaleY
	}
	ox := (cW - worldW*sc) / 2
	oy := (cH - worldH*sc) / 2
	// Affine: x' = (x - bx1)*sc + ox,  y' = (y - by1)*sc + oy
	// simplified: x' = x*sc + (ox - bx1*sc)
	tx := ox - bx1*sc
	ty := oy - by1*sc

	// Apply interactive zoom: translate+scale around zoom center.
	sc *= flash2Zoom
	tx = tx*flash2Zoom + flash2ZoomX
	ty = ty*flash2Zoom + flash2ZoomY

	// Pre-flatten all paths in screen coordinates.
	flatPaths := make([][]shapesdata.FlatVertex, len(shape.Paths))
	for i := range shape.Paths {
		flatPaths[i] = shapesdata.FlattenPath(&shape.Paths[i], sc, sc, tx, ty, 1.0)
	}

	// Set up raw renderer pipeline (bypass Agg2D for direct scanline access).
	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Width()*4)
	pixFmt := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](pixFmt)
	renBase.ClipBox(0, 0, img.Width(), img.Height())
	renBase.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 242, A: 255})

	// Clipping rasterizer (matches C++ rasterizer_scanline_aa<rasterizer_sl_clip_dbl>).
	clipper := rasterizer.NewRasterizerSlClip[float64, rasterizer.DblConv](rasterizer.DblConv{})
	ras := rasterizer.NewRasterizerScanlineAA[float64, rasterizer.DblConv, *rasterizer.RasterizerSlClip[float64, rasterizer.DblConv]](
		rasterizer.DblConv{}, clipper,
	)
	ras.ClipBox(0, 0, float64(img.Width()), float64(img.Height()))
	ras.AutoClose(false)

	sl := scanline.NewScanlineU8()
	slRas := &rasScanlineAdapter{sl: sl}

	// --- Fill pass (flash2 method) ---
	tFillStart := time.Now()
	for s := shape.MinStyle; s <= shape.MaxStyle; s++ {
		ras.Reset()
		for i, p := range shape.Paths {
			if p.LeftFill == p.RightFill {
				continue
			}
			flat := flatPaths[i]
			if len(flat) == 0 {
				continue
			}
			if p.LeftFill == s {
				// Forward: add path vertices as-is.
				vs := &flatVertexSource{verts: flat, pos: 0}
				ras.AddPath(vs, 0)
			}
			if p.RightFill == s {
				// Inverted: reversed vertices with shifted commands.
				vs := &invertedFlatVS{verts: flat, pos: 0}
				ras.AddPath(vs, 0)
			}
		}
		if !ras.RewindScanlines() {
			continue
		}
		sl.Reset(ras.MinX(), ras.MaxX())
		c := styleColor(s)
		for ras.SweepScanline(slRas) {
			renscan.RenderScanlineAASolid(
				&scanlineWrapperU8{sl: sl},
				renBase,
				c,
			)
		}
	}

	tFill := time.Since(tFillStart)

	// --- Stroke pass (using conv_stroke with round joins/caps, matching C++) ---
	tStrokeStart := time.Now()
	ras.AutoClose(true)

	strokeColor := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 128}
	strokeW := math.Sqrt(sc)
	if strokeW < 0.5 {
		strokeW = 0.5
	}

	flatSrc := &flatConvVS{}
	stroke := conv.NewConvStroke(flatSrc)
	stroke.SetWidth(strokeW)
	stroke.SetLineJoin(basics.RoundJoin)
	stroke.SetLineCap(basics.RoundCap)

	strokeRasVS := &convStrokeRasVS{stroke: stroke}

	for i, p := range shape.Paths {
		if p.Line < 0 {
			continue
		}
		flat := flatPaths[i]
		if len(flat) == 0 {
			continue
		}
		ras.Reset()
		flatSrc.verts = flat
		ras.AddPath(strokeRasVS, 0)
		if !ras.RewindScanlines() {
			continue
		}
		sl.Reset(ras.MinX(), ras.MaxX())
		for ras.SweepScanline(slRas) {
			renscan.RenderScanlineAASolid(
				&scanlineWrapperU8{sl: sl},
				renBase,
				strokeColor,
			)
		}
	}

	tStroke := time.Since(tStrokeStart)
	tTotal := tFill + tStroke

	// --- Text overlay (timing info, matching C++ gsv_text output) ---
	ras.AutoClose(true)
	tfillMs := float64(tFill.Microseconds()) / 1000.0
	tstrokeMs := float64(tStroke.Microseconds()) / 1000.0
	ttotalMs := float64(tTotal.Microseconds()) / 1000.0
	fillFPS, strokeFPS, totalFPS := 0, 0, 0
	if tfillMs > 0 {
		fillFPS = int(1000.0 / tfillMs)
	}
	if tstrokeMs > 0 {
		strokeFPS = int(1000.0 / tstrokeMs)
	}
	if ttotalMs > 0 {
		totalFPS = int(1000.0 / ttotalMs)
	}

	txt := fmt.Sprintf("Fill=%.2fms (%dFPS) Stroke=%.2fms (%dFPS) Total=%.2fms (%dFPS)",
		tfillMs, fillFPS, tstrokeMs, strokeFPS, ttotalMs, totalFPS)

	t := gsv.NewGSVText()
	t.SetSize(8.0, 0)
	t.SetFlip(true)
	t.SetStartPoint(10.0, 20.0)
	t.SetText(txt)

	ts := gsv.NewGSVTextOutline(t)
	ts.SetWidth(1.6)

	textRasVS := &convVertexSourceRasVS{src: ts}
	ras.Reset()
	ras.AddPath(textRasVS, 0)
	if ras.RewindScanlines() {
		sl.Reset(ras.MinX(), ras.MaxX())
		textColor := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255}
		for ras.SweepScanline(slRas) {
			renscan.RenderScanlineAASolid(
				&scanlineWrapperU8{sl: sl},
				renBase,
				textColor,
			)
		}
	}

	_ = agg.RGBA(0, 0, 0, 0) // keep agg import live
}

// styleColor returns the random colour for a given style index.
func styleColor(s int) color.RGBA8[color.Linear] {
	if s < 0 || s >= len(flash2Colors) {
		return color.RGBA8[color.Linear]{R: 200, G: 200, B: 200, A: 200}
	}
	return flash2Colors[s]
}

// --- Flat vertex sources ---

// flatVertexSource iterates FlatVertex slices forward.
type flatVertexSource struct {
	verts []shapesdata.FlatVertex
	pos   int
}

func (v *flatVertexSource) Rewind(_ uint32) { v.pos = 0 }
func (v *flatVertexSource) Vertex(x, y *float64) uint32 {
	if v.pos >= len(v.verts) {
		return uint32(basics.PathCmdStop)
	}
	fv := v.verts[v.pos]
	v.pos++
	*x, *y = fv.X, fv.Y
	return fv.Cmd
}

// invertedFlatVS iterates FlatVertex slices with polygon winding inverted.
// This mirrors C++ path_storage::invert_polygon exactly:
//
//  1. Commands are shifted left by one position.
//  2. The original first command (MoveTo) moves to the last position.
//  3. Coordinates are reversed.
//
// Result: LineTo(pN), LineTo(pN-1), …, LineTo(p1), MoveTo(p0).
//
// The first vertex being LineTo (not MoveTo) is intentional — with auto_close(false),
// it draws a stitching edge from the previous sub-path's endpoint, which is how the
// C++ flash_rasterizer2 demo achieves proper compound-shape fill stitching.
type invertedFlatVS struct {
	verts []shapesdata.FlatVertex
	pos   int
}

func (v *invertedFlatVS) Rewind(_ uint32) { v.pos = 0 }
func (v *invertedFlatVS) Vertex(x, y *float64) uint32 {
	n := len(v.verts)
	if v.pos >= n {
		return uint32(basics.PathCmdStop)
	}
	// Reversed coordinate index.
	fv := v.verts[n-1-v.pos]
	*x, *y = fv.X, fv.Y

	// Shifted command: cmd[pos] = original_cmd[pos+1], except last gets original_cmd[0].
	// Since flat verts are [MoveTo, LineTo, …, LineTo], this yields
	// [LineTo, LineTo, …, LineTo, MoveTo].
	var cmd uint32
	if v.pos < n-1 {
		cmd = v.verts[v.pos+1].Cmd
	} else {
		cmd = v.verts[0].Cmd // MoveTo
	}
	v.pos++
	return cmd
}

// --- conv.VertexSource adapter for flat vertices (feeds into ConvStroke) ---

// flatConvVS implements conv.VertexSource for pre-flattened vertex slices.
type flatConvVS struct {
	verts []shapesdata.FlatVertex
	pos   int
}

func (v *flatConvVS) Rewind(_ uint) { v.pos = 0 }
func (v *flatConvVS) Vertex() (x, y float64, cmd basics.PathCommand) {
	if v.pos >= len(v.verts) {
		return 0, 0, basics.PathCmdStop
	}
	fv := v.verts[v.pos]
	v.pos++
	return fv.X, fv.Y, basics.PathCommand(fv.Cmd)
}

// convStrokeRasVS adapts conv.ConvStroke to the rasterizer's VertexSource interface.
type convStrokeRasVS struct {
	stroke *conv.ConvStroke
}

func (a *convStrokeRasVS) Rewind(pathID uint32) { a.stroke.Rewind(uint(pathID)) }
func (a *convStrokeRasVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.stroke.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// convVertexSourceRasVS adapts any conv.VertexSource to the rasterizer's VertexSource interface.
type convVertexSourceRasVS struct {
	src conv.VertexSource
}

func (a *convVertexSourceRasVS) Rewind(pathID uint32) { a.src.Rewind(uint(pathID)) }
func (a *convVertexSourceRasVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}
