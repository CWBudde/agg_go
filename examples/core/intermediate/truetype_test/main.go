// Go-idiomatic equivalent of AGG's truetype_test.cpp.
//
// This standalone demo focuses on the classic rendering showcase rather than the
// original interactive controls. It renders the same long TrueType paragraph in
// three modes: anti-aliased gray8, weighted outline, and monochrome raster.
// FreeType is required for full output; without it the demo renders an
// explanatory fallback note.
package main

import (
	"os"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/font"
	"github.com/MeKo-Christian/agg_go/internal/font/freetype"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

const demoText = "Anti-Grain Geometry is designed as a set of loosely coupled algorithms and class templates united with a common idea, so that all the components can be easily combined. Also, the template based design allows you to replace any part of the library without the necessity to modify a single byte in the existing code. AGG is designed keeping in mind extensibility and flexibility. Basically I just wanted to create a toolkit that would allow me and anyone else to add new fancy algorithms very easily."

const (
	width       = 960
	height      = 620
	margin      = 18
	panelGap    = 14
	panelTop    = 70
	panelBottom = 28
	fontSize    = 18.0
	lineGap     = 4.0
)

type panel struct {
	title       string
	rendering   freetype.GlyphRenderingType
	outlineW    float64
	hinting     bool
	panelX1     int
	panelY1     int
	panelX2     int
	panelY2     int
	description string
}

type demo struct{}

func (d *demo) Render(img *agg.Image) {
	ctx := agg.NewContextForImage(img)
	ctx.Clear(agg.RGBA(1, 1, 1, 1))

	rbuf := buffer.NewRenderingBufferU8WithData(img.Data, img.Width(), img.Height(), img.Width()*4)
	pf := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]](pf)
	renBase.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	a := ctx.GetAgg2D()
	a.ResetTransformations()
	drawLayout(a)

	fontPath := findTrueTypeFont()
	engine, err := freetype.NewFontEngineFreetype(false, 32)
	if err != nil || fontPath == "" {
		drawFallback(a, err, fontPath == "")
		return
	}
	defer func() { _ = engine.Close() }()

	for _, p := range buildPanels() {
		if err := renderPanel(renBase, engine, p, fontPath); err != nil {
			drawPanelError(a, p, err.Error())
		}
	}

	drawTitles(a)
}

func buildPanels() []panel {
	panelW := (width - margin*2 - panelGap*2) / 3
	y2 := height - panelBottom
	return []panel{
		{
			title:       "Gray 8",
			description: "Anti-aliased bitmap glyphs",
			rendering:   freetype.GlyphRenderingAAGray8,
			hinting:     true,
			panelX1:     margin,
			panelY1:     panelTop,
			panelX2:     margin + panelW,
			panelY2:     y2,
		},
		{
			title:       "Outline",
			description: "Vector outlines with contour weight",
			rendering:   freetype.GlyphRenderingOutline,
			outlineW:    -fontSize * 0.06,
			hinting:     false,
			panelX1:     margin + panelW + panelGap,
			panelY1:     panelTop,
			panelX2:     margin + panelW*2 + panelGap,
			panelY2:     y2,
		},
		{
			title:       "Mono",
			description: "1-bit bitmap rasterization",
			rendering:   freetype.GlyphRenderingMono,
			hinting:     true,
			panelX1:     margin + (panelW+panelGap)*2,
			panelY1:     panelTop,
			panelX2:     margin + panelW*3 + panelGap*2,
			panelY2:     y2,
		},
	}
}

func renderPanel(
	renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]],
	engine *freetype.FontEngineFreetype,
	p panel,
	fontPath string,
) error {
	if err := engine.LoadFont(fontPath, 0, p.rendering, nil); err != nil {
		return err
	}
	engine.SetHinting(p.hinting)
	engine.SetHeight(fontSize)
	engine.SetWidth(0.0)
	engine.SetFlipY(true)

	mtx := transform.NewTransAffine()
	mtx.Rotate(basics.Deg2RadF(-4.0))
	engine.SetTransform(mtx)

	fcm := font.NewFontCacheManager(engine, 32)

	x := float64(p.panelX1 + 10)
	y := float64(p.panelY1) + fontSize + 8
	lineAdvance := fontSize + lineGap

	var prevGlyphIndex uint
	firstGlyph := true

	for _, r := range demoText {
		glyph := fcm.Glyph(uint(r))
		if glyph == nil {
			continue
		}

		if !firstGlyph {
			fcm.AddKerning(&x, &y, prevGlyphIndex, glyph.GlyphIndex)
		}
		if x+glyph.AdvanceX >= float64(p.panelX2-8) {
			x = float64(p.panelX1 + 10)
			y += lineAdvance
		}
		if y >= float64(p.panelY2-8) {
			break
		}

		fcm.InitEmbeddedAdaptors(glyph, x, y)
		switch glyph.DataType {
		case font.GlyphDataOutline:
			if err := renderOutlineGlyph(renBase, fcm.PathAdaptor(), p.outlineW); err != nil {
				return err
			}
		case font.GlyphDataGray8:
			renderGrayGlyph(renBase, glyph, x, y)
		case font.GlyphDataMono:
			renderMonoGlyph(renBase, glyph, x, y)
		}

		x += glyph.AdvanceX
		y += glyph.AdvanceY
		prevGlyphIndex = glyph.GlyphIndex
		firstGlyph = false
	}

	return nil
}

func renderOutlineGlyph(
	renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]],
	ps *path.PathStorageStl,
	outlineW float64,
) error {
	curves := conv.NewConvCurve(path.NewPathStorageStlVertexSourceAdapter(ps))
	curves.SetApproximationScale(2.0)

	var src conv.VertexSource = curves
	if outlineW != 0 {
		contour := conv.NewConvContour(curves)
		contour.Width(outlineW)
		contour.AutoDetectOrientation(false)
		src = contour
	}

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	ras.AddPath(&convToRasAdapter{src: src}, 0)

	sl := scanline.NewScanlineU8()
	renscan.RenderScanlinesAASolid(
		ras,
		sl,
		renBase,
		color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255},
	)
	return nil
}

func renderGrayGlyph(
	renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]],
	glyph *font.GlyphCache,
	x, y float64,
) {
	sl := scanline.NewScanlineU8()
	ras := &glyphBitmapRasterizer{
		data:     glyph.Data,
		bounds:   glyph.Bounds,
		dataType: glyph.DataType,
		pitch:    max(1, glyph.Bounds.X2-glyph.Bounds.X1),
		offsetX:  basics.IRound(x),
		offsetY:  basics.IRound(y),
	}
	renscan.RenderScanlinesAASolid(
		ras,
		sl,
		renBase,
		color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255},
	)
}

func renderMonoGlyph(
	renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]],
	glyph *font.GlyphCache,
	x, y float64,
) {
	width := glyph.Bounds.X2 - glyph.Bounds.X1
	sl := scanline.NewScanlineU8()
	ras := &glyphBitmapRasterizer{
		data:     glyph.Data,
		bounds:   glyph.Bounds,
		dataType: glyph.DataType,
		pitch:    max(1, (width+7)>>3),
		offsetX:  basics.IRound(x),
		offsetY:  basics.IRound(y),
	}
	renscan.RenderScanlinesBinSolid(
		ras,
		sl,
		renBase,
		color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255},
	)
}

func drawLayout(a *agg.Agg2D) {
	a.LineColor(agg.NewColor(210, 210, 210, 255))
	a.LineWidth(1.0)
	a.FillColor(agg.NewColor(250, 250, 248, 255))
	for _, p := range buildPanels() {
		a.Rectangle(float64(p.panelX1), float64(p.panelY1), float64(p.panelX2), float64(p.panelY2))
	}
}

func drawTitles(a *agg.Agg2D) {
	a.FontGSV(16)
	a.FillColor(agg.Black)
	a.NoLine()
	a.Text(18, 24, "TrueType Test", false, 0, 0)
	a.FontGSV(10)
	a.Text(18, 42, "Classic standalone FreeType demo: Gray8, Outline, Mono", false, 0, 0)
	for _, p := range buildPanels() {
		a.FontGSV(12)
		a.Text(float64(p.panelX1+8), float64(p.panelY1-18), p.title, false, 0, 0)
		a.FontGSV(9)
		a.Text(float64(p.panelX1+8), float64(p.panelY1-6), p.description, false, 0, 0)
	}
}

func drawFallback(a *agg.Agg2D, fontErr error, missingFont bool) {
	drawTitles(a)
	a.FontGSV(12)
	a.FillColor(agg.Black)
	a.Text(18, 88, "FreeType TrueType rendering is unavailable in this build.", false, 0, 0)
	if missingFont {
		a.Text(18, 108, "No suitable .ttf font was found in common system locations.", false, 0, 0)
		return
	}
	if fontErr != nil {
		a.Text(18, 108, fontErr.Error(), false, 0, 0)
	}
}

func drawPanelError(a *agg.Agg2D, p panel, msg string) {
	a.FontGSV(9)
	a.FillColor(agg.Black)
	a.Text(float64(p.panelX1+8), float64(p.panelY1+18), msg, false, 0, 0)
}

func findTrueTypeFont() string {
	candidates := []string{
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/usr/share/fonts/TTF/DejaVuSans.ttf",
		"/usr/share/fonts/truetype/liberation2/LiberationSans-Regular.ttf",
		"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf",
		"/usr/share/fonts/truetype/noto/NotoSans-Regular.ttf",
		"/System/Library/Fonts/Supplemental/Arial.ttf",
		"/System/Library/Fonts/Arial.ttf",
		"C:\\Windows\\Fonts\\arial.ttf",
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

type convToRasAdapter struct {
	src conv.VertexSource
}

func (a *convToRasAdapter) Rewind(pathID uint32) {
	a.src.Rewind(uint(pathID))
}

func (a *convToRasAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

type glyphBitmapRasterizer struct {
	data     []byte
	bounds   basics.Rect[int]
	dataType font.GlyphDataType
	pitch    int
	offsetX  int
	offsetY  int
	row      int
}

func (r *glyphBitmapRasterizer) RewindScanlines() bool {
	r.row = 0
	return len(r.data) > 0 && (r.bounds.X2-r.bounds.X1) > 0 && (r.bounds.Y2-r.bounds.Y1) > 0
}

func (r *glyphBitmapRasterizer) MinX() int { return r.bounds.X1 + r.offsetX }
func (r *glyphBitmapRasterizer) MaxX() int { return r.bounds.X2 + r.offsetX - 1 }

func (r *glyphBitmapRasterizer) SweepScanline(sl renscan.ScanlineInterface) bool {
	slU8, ok := sl.(*scanline.ScanlineU8)
	if !ok || slU8 == nil {
		return false
	}

	width := r.bounds.X2 - r.bounds.X1
	height := r.bounds.Y2 - r.bounds.Y1

	for r.row < height {
		row := r.row
		r.row++

		rowStart := row * r.pitch
		if rowStart >= len(r.data) {
			continue
		}
		rowEnd := rowStart + r.pitch
		if rowEnd > len(r.data) {
			rowEnd = len(r.data)
		}
		rowData := r.data[rowStart:rowEnd]

		slU8.ResetSpans()
		scanY := r.bounds.Y1 + r.offsetY + row
		baseX := r.bounds.X1 + r.offsetX

		if r.dataType == font.GlyphDataMono {
			runStart := -1
			for col := 0; col < width; col++ {
				byteIdx := col >> 3
				bitSet := false
				if byteIdx < len(rowData) {
					bit := uint(7 - (col & 7))
					bitSet = ((rowData[byteIdx] >> bit) & 0x1) != 0
				}
				if bitSet {
					if runStart < 0 {
						runStart = col
					}
					continue
				}
				if runStart >= 0 {
					slU8.AddSpan(baseX+runStart, col-runStart, uint(basics.CoverFull))
					runStart = -1
				}
			}
			if runStart >= 0 {
				slU8.AddSpan(baseX+runStart, width-runStart, uint(basics.CoverFull))
			}
		} else {
			runStart := -1
			covers := make([]basics.Int8u, 0, width)
			flush := func() {
				if runStart >= 0 && len(covers) > 0 {
					slU8.AddCells(baseX+runStart, len(covers), covers)
				}
				runStart = -1
				covers = covers[:0]
			}

			for col := 0; col < width; col++ {
				var cov basics.Int8u
				if col < len(rowData) {
					cov = basics.Int8u(rowData[col])
				}
				if cov == 0 {
					flush()
					continue
				}
				if runStart < 0 {
					runStart = col
				}
				covers = append(covers, cov)
			}
			flush()
		}

		if slU8.NumSpans() > 0 {
			slU8.Finalize(scanY)
			return true
		}
	}

	return false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "TrueType Test",
		Width:  width,
		Height: height,
	}, &demo{})
}
