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
	"math"
	"math/rand"

	agg "agg_go"
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/demo/shapesdata"
	"agg_go/internal/pixfmt"
	"agg_go/internal/rasterizer"
	"agg_go/internal/renderer"
	renscan "agg_go/internal/renderer/scanline"
	"agg_go/internal/scanline"
)

// --- State ---

var (
	flash2ShapeIdx = 0

	flash2Shapes []shapesdata.RawShape
	flash2Colors []color.RGBA8[color.Linear] // 100 random colours
)

func setFlash2ShapeIdx(v int) { flash2ShapeIdx = v }

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

	// Pre-flatten all paths in screen coordinates.
	flatPaths := make([][]shapesdata.FlatVertex, len(shape.Paths))
	for i := range shape.Paths {
		flatPaths[i] = shapesdata.FlattenPath(&shape.Paths[i], sc, sc, tx, ty)
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

	// --- Stroke pass ---
	ras.AutoClose(true)

	// Re-use the same rasterizer with auto_close on.
	strokeColor := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 128}
	strokeW := math.Sqrt(sc)
	if strokeW < 0.5 {
		strokeW = 0.5
	}

	for i, p := range shape.Paths {
		if p.Line < 0 {
			continue
		}
		flat := flatPaths[i]
		if len(flat) == 0 {
			continue
		}
		ras.Reset()
		strokeFlatPath(ras, flat, strokeW)
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

	// Blit the rendered buffer into the canvas via the Agg2D context.
	// (The pixel data is already in img.Data which is the canvas buffer.)
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
// This mirrors C++ path_storage::invert_polygon:
//
//	commands are shifted by one position (the leading MoveTo shifts to the last vertex),
//	and the coordinate sequence is reversed.
//
// The resulting order is: LineTo(xN), LineTo(xN-1), …, LineTo(x1), MoveTo(x0).
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
	// reversed index into verts
	i := n - 1 - v.pos
	fv := v.verts[i]
	*x, *y = fv.X, fv.Y

	// Command shifting: the original MoveTo (at index 0) becomes the last vertex's command.
	// All other vertices keep PathCmdLineTo.
	var cmd uint32
	if v.pos == n-1 {
		// last emitted vertex → gets the original MoveTo command
		cmd = shapesdata.PathCmdMoveTo
	} else {
		cmd = shapesdata.PathCmdLineTo
	}
	v.pos++
	return cmd
}

// --- Minimal stroke expander for flat paths ---
// We stroke by expanding each segment by half-width on both sides.

func strokeFlatPath(ras *rasterizer.RasterizerScanlineAA[float64, rasterizer.DblConv, *rasterizer.RasterizerSlClip[float64, rasterizer.DblConv]], flat []shapesdata.FlatVertex, w float64) {
	if len(flat) < 2 {
		return
	}
	hw := w * 0.5
	for i := 1; i < len(flat); i++ {
		if flat[i].Cmd != shapesdata.PathCmdLineTo {
			continue
		}
		x1, y1 := flat[i-1].X, flat[i-1].Y
		x2, y2 := flat[i].X, flat[i].Y
		dx, dy := x2-x1, y2-y1
		d := math.Sqrt(dx*dx + dy*dy)
		if d < 1e-6 {
			continue
		}
		nx, ny := -dy/d*hw, dx/d*hw
		// Parallelogram around the segment
		ras.MoveToD(x1+nx, y1+ny)
		ras.LineToD(x2+nx, y2+ny)
		ras.LineToD(x2-nx, y2-ny)
		ras.LineToD(x1-nx, y1-ny)
	}
}
