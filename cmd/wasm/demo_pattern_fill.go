// Based on the original AGG example: pattern_fill.cpp
// Demonstrates tiled pattern filling: a small star pattern is drawn into an
// offscreen buffer and then used as a repeating fill for a larger star polygon.
package main

import (
	"math"

	agg "agg_go"
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	imageacc "agg_go/internal/image"
	"agg_go/internal/path"
	"agg_go/internal/pixfmt"
	"agg_go/internal/rasterizer"
	"agg_go/internal/renderer"
	"agg_go/internal/scanline"
	"agg_go/internal/span"
)

// --- Demo state ---

var (
	patFillPolygonAngle = 0.0  // -180..180
	patFillPolygonScale = 1.0  // 0.1..5.0
	patFillPatternAngle = 0.0  // -180..180
	patFillPatternSize  = 30.0 // 10..60
	patFillPolygonCX    = 0.0
	patFillPolygonCY    = 0.0

	patFillRbuf        *buffer.RenderingBufferU8
	patFillPixFmt      *pixfmt.PixFmtRGBA32Pre[color.Linear]
	patFillRenBase     *renderer.RendererBase[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]]
	patFillAlloc       *span.SpanAllocator[color.RGBA8[color.Linear]]
	patFillRas         *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]
	patFillSl          *scanline.ScanlineU8
	patFillPath        *path.PathStorageStl
	patFillInitialized bool
)

// patternPixFmt implements image.PixelFormat for an in-memory RGBA pattern buffer.
type patternPixFmt struct {
	data         []basics.Int8u
	w, h, stride int
}

func (p patternPixFmt) Width() int    { return p.w }
func (p patternPixFmt) Height() int   { return p.h }
func (p patternPixFmt) PixWidth() int { return 4 }
func (p patternPixFmt) PixPtr(x, y int) []basics.Int8u {
	if y < 0 || y >= p.h || x < 0 || x >= p.w {
		return p.data[0:]
	}
	return p.data[y*p.stride+x*4:]
}

// patternSource wraps an ImageAccessorWrap over patternPixFmt to satisfy
// span.RGBASourceInterface so it can be used with SpanPatternRGBA.
type patternSource struct {
	accessor *imageacc.ImageAccessorWrap[patternPixFmt, *imageacc.WrapModeReflectAutoPow2, *imageacc.WrapModeReflectAutoPow2]
	pf       patternPixFmt
}

func (s *patternSource) Width() int                  { return s.pf.w }
func (s *patternSource) Height() int                 { return s.pf.h }
func (s *patternSource) ColorType() string           { return "RGBA8" }
func (s *patternSource) OrderType() color.ColorOrder { return color.OrderRGBA }
func (s *patternSource) Span(x, y, length int) []basics.Int8u {
	return s.accessor.Span(x, y, length)
}
func (s *patternSource) NextX() []basics.Int8u { return s.accessor.NextX() }
func (s *patternSource) NextY() []basics.Int8u { return s.accessor.NextY() }
func (s *patternSource) RowPtr(y int) []basics.Int8u {
	return s.pf.PixPtr(0, y)
}

func initPatFillDemo() {
	if patFillInitialized {
		return
	}
	patFillRbuf = buffer.NewRenderingBufferU8()
	patFillPixFmt = pixfmt.NewPixFmtRGBA32PreLinear(patFillRbuf)
	patFillRenBase = renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]](patFillPixFmt)
	patFillAlloc = span.NewSpanAllocator[color.RGBA8[color.Linear]]()
	patFillRas = rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	patFillSl = scanline.NewScanlineU8()
	patFillPath = path.NewPathStorageStl()
	patFillPolygonCX = float64(width) / 2.0
	patFillPolygonCY = float64(height) / 2.0
	patFillInitialized = true
}

// generatePattern renders a small star shape into a patternPixFmt buffer.
func generatePattern(size int, patternAngleDeg float64) patternPixFmt {
	pf := patternPixFmt{
		w:      size,
		h:      size,
		stride: size * 4,
		data:   make([]basics.Int8u, size*size*4),
	}

	// Background: dark reddish, semi-transparent
	bgR, bgG, bgB, bgA := uint8(102), uint8(0), uint8(26), uint8(51)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			i := (y*size + x) * 4
			pf.data[i+0] = bgR // R
			pf.data[i+1] = bgG // G
			pf.data[i+2] = bgB // B
			pf.data[i+3] = bgA // A
		}
	}

	// Draw a mini star using the public API into a temporary agg.Image
	tmpImg := agg.CreateImage(size, size)
	tmpCtx := agg.NewContextForImage(tmpImg)

	// Fill background
	tmpCtx.SetColor(agg.RGBA(0.4, 0.0, 0.1, 0.2))
	tmpCtx.FillRectangle(0, 0, float64(size), float64(size))

	// Star fill
	tmpCtx.SetColor(agg.RGBA(0.43, 0.51, 0.20, 1.0))
	tmpCtx.MoveTo(0, 0) // will be overwritten by star
	cx := float64(size) / 2.0
	cy := float64(size) / 2.0
	r1 := float64(size)/2.0 - 1
	r2 := r1 / 2.5
	n := 6
	startRad := patternAngleDeg * math.Pi / 180.0

	// Build star path in tmpCtx
	for i := 0; i < n; i++ {
		a := math.Pi*2.0*float64(i)/float64(n) - math.Pi/2.0 + startRad
		var px, py float64
		if i&1 != 0 {
			px = cx + math.Cos(a)*r1
			py = cy + math.Sin(a)*r1
			tmpCtx.LineTo(px, py)
		} else {
			px = cx + math.Cos(a)*r2
			py = cy + math.Sin(a)*r2
			if i == 0 {
				tmpCtx.MoveTo(px, py)
			} else {
				tmpCtx.LineTo(px, py)
			}
		}
	}
	tmpCtx.ClosePath()
	tmpCtx.Fill()

	// Star outline
	tmpCtx.SetColor(agg.RGBA(0.0, 0.20, 0.31, 1.0))
	tmpCtx.SetLineWidth(float64(size) / 15.0)
	for i := 0; i < n; i++ {
		a := math.Pi*2.0*float64(i)/float64(n) - math.Pi/2.0 + startRad
		var px, py float64
		if i&1 != 0 {
			px = cx + math.Cos(a)*r1
			py = cy + math.Sin(a)*r1
			tmpCtx.LineTo(px, py)
		} else {
			px = cx + math.Cos(a)*r2
			py = cy + math.Sin(a)*r2
			if i == 0 {
				tmpCtx.MoveTo(px, py)
			} else {
				tmpCtx.LineTo(px, py)
			}
		}
	}
	tmpCtx.ClosePath()
	tmpCtx.Stroke()

	// Copy tmpImg RGBA data into pf
	copy(pf.data, tmpImg.Data)
	return pf
}

func drawPatternFillDemo() {
	initPatFillDemo()

	// Attach rendering target
	img := ctx.GetImage()
	patFillRbuf.Attach(img.Data, img.Width(), img.Height(), img.Width()*4)
	// Keep renderer clip box in sync after dynamic buffer attach.
	patFillRenBase.Attach(patFillPixFmt)

	size := int(patFillPatternSize)
	if size < 4 {
		size = 4
	}

	// Generate pattern
	pf := generatePattern(size, patFillPatternAngle)

	// Wrap mode for tiling
	wrapX := imageacc.NewWrapModeReflectAutoPow2(basics.Int32u(size))
	wrapY := imageacc.NewWrapModeReflectAutoPow2(basics.Int32u(size))
	accessor := imageacc.NewImageAccessorWrap[patternPixFmt, *imageacc.WrapModeReflectAutoPow2, *imageacc.WrapModeReflectAutoPow2](&pf, wrapX, wrapY)
	src := &patternSource{accessor: accessor, pf: pf}
	sg := span.NewSpanPatternRGBAWithParams[*patternSource](src, 0, 0)

	// Large star polygon
	polyAngleRad := patFillPolygonAngle * math.Pi / 180.0
	r := float64(width) / 3.0
	r1 := r - 8.0
	r2 := r1 / 1.45
	n := 14

	patFillPath.RemoveAll()
	for i := 0; i < n; i++ {
		a := math.Pi*2.0*float64(i)/float64(n) - math.Pi/2.0
		dx := math.Cos(a)
		dy := math.Sin(a)
		var px, py float64
		if i&1 != 0 {
			px = patFillPolygonCX + dx*r1
			py = patFillPolygonCY + dy*r1
			patFillPath.LineTo(px, py)
		} else {
			px = patFillPolygonCX + dx*r2
			py = patFillPolygonCY + dy*r2
			if i == 0 {
				patFillPath.MoveTo(px, py)
			} else {
				patFillPath.LineTo(px, py)
			}
		}
	}
	patFillPath.ClosePolygon(basics.PathFlagsClose)

	// Apply polygon rotation/scale
	rotated := path.NewPathStorageStl()
	patFillPath.Rewind(0)
	for {
		vx, vy, rawCmd := patFillPath.NextVertex()
		cmd := basics.PathCommand(rawCmd)
		if basics.IsStop(cmd) {
			break
		}
		// Rotate around polygon center
		dx := vx - patFillPolygonCX
		dy := vy - patFillPolygonCY
		rdx := dx*math.Cos(polyAngleRad) - dy*math.Sin(polyAngleRad)
		rdy := dx*math.Sin(polyAngleRad) + dy*math.Cos(polyAngleRad)
		vx = patFillPolygonCX + rdx*patFillPolygonScale
		vy = patFillPolygonCY + rdy*patFillPolygonScale
		switch {
		case basics.IsMoveTo(cmd):
			rotated.MoveTo(vx, vy)
		case basics.IsLineTo(cmd):
			rotated.LineTo(vx, vy)
		case basics.IsEndPoly(cmd):
			rotated.ClosePolygon(basics.PathFlagsClose)
		}
	}

	// Render with pattern fill
	patFillRas.Reset()
	patFillRas.ClipBox(0, 0, float64(width), float64(height))
	patFillRas.AddPath(&pathSourceAdapter{ps: rotated}, 0)

	if patFillRas.RewindScanlines() {
		patFillSl.Reset(patFillRas.MinX(), patFillRas.MaxX())
		for patFillRas.SweepScanline(&rasScanlineAdapter{sl: patFillSl}) {
			y := patFillSl.Y()
			for _, spanData := range patFillSl.Spans() {
				if spanData.Len > 0 {
					colors := patFillAlloc.Allocate(int(spanData.Len))
					sg.Generate(colors, int(spanData.X), y, uint(spanData.Len))
					patFillRenBase.BlendColorHspan(int(spanData.X), y, int(spanData.Len), colors, spanData.Covers, basics.CoverFull)
				}
			}
		}
	}
}

func setPatFillPolygonAngle(v float64) { patFillPolygonAngle = v }
func setPatFillPolygonScale(v float64) { patFillPolygonScale = v }
func setPatFillPatternAngle(v float64) { patFillPatternAngle = v }
func setPatFillPatternSize(v float64)  { patFillPatternSize = v }
