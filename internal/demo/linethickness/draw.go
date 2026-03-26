// Package linethickness ports AGG's line_thickness.cpp demo.
//
// Rendering uses the low-level AGG pipeline with the same coordinates as the
// C++ source (flip_y=true convention: y=0 at bottom, y=479 at top).
// The caller is responsible for copying the work buffer to the output image
// with a y-flip if needed.
package linethickness

import (
	"math"
	"time"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/effects"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
)

const (
	Width  = 640
	Height = 480
)

// State holds the interactive control values for the demo.
type State struct {
	Thickness float64
	Blur      float64
	Mono      bool
	Invert    bool
}

// DefaultState returns the C++ default control values.
func DefaultState() State {
	return State{
		Thickness: 1.0,
		Blur:      1.5,
		Mono:      true,
		Invert:    false,
	}
}

// Clamp clamps the state values to their valid ranges.
func (s *State) Clamp() {
	if s.Thickness < 0 {
		s.Thickness = 0
	}
	if s.Thickness > 5 {
		s.Thickness = 5
	}
	if s.Blur < 0 {
		s.Blur = 0
	}
	if s.Blur > 2 {
		s.Blur = 2
	}
}

var lastBlurMS float64

// LastBlurMS returns the time in milliseconds taken by the last blur pass.
func LastBlurMS() float64 { return lastBlurMS }

// Colors returns the foreground and background colors for the given state,
// matching C++ line_thickness.cpp's color selection logic.
func Colors(st State) (fg, bg color.RGBA8[color.Linear]) {
	// C++: clr1 = mono ? rgba(1,1,1) : rgba(1,0,1)
	//      clr2 = mono ? rgba(0,0,0) : rgba(0,1,0)
	//      foreground = invert ? clr1 : clr2
	//      background = invert ? clr2 : clr1
	var clr1, clr2 color.RGBA8[color.Linear]
	if st.Mono {
		clr1 = color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255}
		clr2 = color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255}
	} else {
		clr1 = color.RGBA8[color.Linear]{R: 255, G: 0, B: 255, A: 255}
		clr2 = color.RGBA8[color.Linear]{R: 0, G: 255, B: 0, A: 255}
	}
	if st.Invert {
		return clr1, clr2
	}
	return clr2, clr1
}

// pathVS adapts PathStorageStl to conv.VertexSource.
type pathVS struct{ ps *path.PathStorageStl }

func (v *pathVS) Rewind(id uint) { v.ps.Rewind(id) }
func (v *pathVS) Vertex() (x, y float64, cmd basics.PathCommand) {
	vx, vy, c := v.ps.NextVertex()
	return vx, vy, basics.PathCommand(c)
}

// rasVS adapts conv.VertexSource to rasterizer.VertexSource.
type rasVS struct{ src conv.VertexSource }

func (v *rasVS) Rewind(id uint32) { v.src.Rewind(uint(id)) }
func (v *rasVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := v.src.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// pixFmtAdapter implements effects.PixFmtInterface for a raw RGBA byte slice
// with positive stride (y=0 at row 0 of the slice).
type pixFmtAdapter struct {
	buf  []uint8
	w, h int
}

func (p *pixFmtAdapter) Width() int  { return p.w }
func (p *pixFmtAdapter) Height() int { return p.h }

func (p *pixFmtAdapter) GetPixel(x, y int) color.RGBA8[color.Linear] {
	if x < 0 || y < 0 || x >= p.w || y >= p.h {
		return color.RGBA8[color.Linear]{}
	}
	i := (y*p.w + x) * 4
	return color.RGBA8[color.Linear]{R: p.buf[i], G: p.buf[i+1], B: p.buf[i+2], A: p.buf[i+3]}
}

func (p *pixFmtAdapter) CopyPixel(x, y int, c color.RGBA8[color.Linear]) {
	if x < 0 || y < 0 || x >= p.w || y >= p.h {
		return
	}
	i := (y*p.w + x) * 4
	p.buf[i], p.buf[i+1], p.buf[i+2], p.buf[i+3] = c.R, c.G, c.B, c.A
}

// Draw renders the line_thickness scene into workBuf using exact C++ coordinates.
//
// workBuf is a positive-stride RGBA buffer of size w*h*4, where row 0 is the
// bottom of the logical frame (y-up / flip_y=true convention).  The caller
// must copy workBuf to the output image with a y-flip after this call.
func Draw(workBuf []uint8, w, h int, st State) {
	st.Clamp()

	fg, bg := Colors(st)

	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(workBuf, w, h, w*4)
	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt(pf)
	renBase.Clear(bg)

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	ps := path.NewPathStorageStl()
	psvs := &pathVS{ps: ps}
	pg := conv.NewConvStroke(psvs)
	pgRas := &rasVS{src: pg}

	// C++: for (int i = 0; i < 20; ++i) { pg.width(...); ps.move_to(20+30*i, 310); ps.line_to(40+30*i, 460); }
	for i := range 20 {
		pg.SetWidth(st.Thickness * 0.3 * float64(i+1))
		ps.RemoveAll()
		ps.MoveTo(float64(20+30*i), 310)
		ps.LineTo(float64(40+30*i), 460)
		ras.Reset()
		ras.AddPath(pgRas, 0)
		renscan.RenderScanlinesAASolid(ras, sl, renBase, fg)
	}

	// C++: for (int i = 0; i < 40; ++i) { pg.width(...); ps.move_to(320+20*sin(...), 180+20*cos(...)); ps.line_to(320+100*sin(...), 180+100*cos(...)); }
	for i := range 40 {
		ang := float64(i) * math.Pi / 20.0
		pg.SetWidth(st.Thickness)
		ps.RemoveAll()
		ps.MoveTo(320+20*math.Sin(ang), 180+20*math.Cos(ang))
		ps.LineTo(320+100*math.Sin(ang), 180+100*math.Cos(ang))
		ras.Reset()
		ras.AddPath(pgRas, 0)
		renscan.RenderScanlinesAASolid(ras, sl, renBase, fg)
	}

	// C++: agg::apply_slight_blur(ren, m_slider2.value())
	if st.Blur > 0 {
		start := time.Now()
		effects.ApplySlightBlurFull(&pixFmtAdapter{buf: workBuf, w: w, h: h}, st.Blur)
		lastBlurMS = time.Since(start).Seconds() * 1000
	} else {
		lastBlurMS = 0
	}
}
