// Package blendcolor implements the blend_color.cpp AGG demo.
// It renders a letter "a" glyph with perspective-distorted shadow,
// blurred with stack blur, and composited onto the main canvas using
// either BlendFromColor (single color) or BlendFromLUT (gradient LUT).
package blendcolor

import (
	"fmt"
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/effects"
	"github.com/MeKo-Christian/agg_go/internal/gsv"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

// Config holds the configuration for the blend color demo.
type Config struct {
	Method int        // 0: Single Color, 1: Color LUT
	Radius float64    // Blur radius 0..40
	Quad   [8]float64 // Shadow quad corners (x0,y0,x1,y1,x2,y2,x3,y3), zero = use shape bounds
}

// Result holds the output state from a Draw call.
type Result struct {
	Quad        [8]float64    // The (potentially initialized) quad corners
	ShapeBounds [4]float64    // x1, y1, x2, y2 of the curved shape
	ElapsedText string        // Formatted timing text (caller provides elapsed time)
}

// pathStorageAdapter adapts path.PathStorageStl to conv.VertexSource.
// PathStorageStl.NextVertex returns (x, y float64, cmd uint32) but
// conv.VertexSource requires Vertex() (x, y float64, cmd basics.PathCommand).
type pathStorageAdapter struct {
	ps *path.PathStorageStl
}

func (a *pathStorageAdapter) Rewind(pathID uint) {
	a.ps.Rewind(pathID)
}

func (a *pathStorageAdapter) Vertex() (x, y float64, cmd basics.PathCommand) {
	vx, vy, rawCmd := a.ps.NextVertex()
	return vx, vy, basics.PathCommand(rawCmd)
}

// rasPathAdapter adapts conv.VertexSource to rasterizer.VertexSource.
type rasPathAdapter struct {
	vs conv.VertexSource
}

func (a *rasPathAdapter) Rewind(pathID uint32) {
	a.vs.Rewind(uint(pathID))
}

func (a *rasPathAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.vs.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

// scanlineAdapter adapts scanline.ScanlineP8 to the rasterizer's ScanlineInterface.
type scanlineAdapter struct {
	sl *scanline.ScanlineP8
}

func (a *scanlineAdapter) ResetSpans()                 { a.sl.ResetSpans() }
func (a *scanlineAdapter) AddCell(x int, cover uint32) { a.sl.AddCell(x, uint(cover)) }
func (a *scanlineAdapter) AddSpan(x, length int, cover uint32) {
	a.sl.AddSpan(x, length, uint(cover))
}
func (a *scanlineAdapter) Finalize(y int) { a.sl.Finalize(y) }
func (a *scanlineAdapter) NumSpans() int  { return a.sl.NumSpans() }

// buildGlyphPath creates the letter "a" glyph path, scales it 4x,
// and translates to (150, 100). Returns the path storage.
func buildGlyphPath() *path.PathStorageStl {
	p := path.NewPathStorageStl()

	p.MoveTo(28.47, 6.45)
	p.Curve3(21.58, 1.12, 19.82, 0.29)
	p.Curve3(17.19, -0.93, 14.21, -0.93)
	p.Curve3(9.57, -0.93, 6.57, 2.25)
	p.Curve3(3.56, 5.42, 3.56, 10.60)
	p.Curve3(3.56, 13.87, 5.03, 16.26)
	p.Curve3(7.03, 19.58, 11.99, 22.51)
	p.Curve3(16.94, 25.44, 28.47, 29.64)
	p.LineTo(28.47, 31.40)
	p.Curve3(28.47, 38.09, 26.34, 40.58)
	p.Curve3(24.22, 43.07, 20.17, 43.07)
	p.Curve3(17.09, 43.07, 15.28, 41.41)
	p.Curve3(13.43, 39.75, 13.43, 37.60)
	p.LineTo(13.53, 34.77)
	p.Curve3(13.53, 32.52, 12.38, 31.30)
	p.Curve3(11.23, 30.08, 9.38, 30.08)
	p.Curve3(7.57, 30.08, 6.42, 31.35)
	p.Curve3(5.27, 32.62, 5.27, 34.81)
	p.Curve3(5.27, 39.01, 9.57, 42.53)
	p.Curve3(13.87, 46.04, 21.63, 46.04)
	p.Curve3(27.59, 46.04, 31.40, 44.04)
	p.Curve3(34.28, 42.53, 35.64, 39.31)
	p.Curve3(36.52, 37.21, 36.52, 30.71)
	p.LineTo(36.52, 15.53)
	p.Curve3(36.52, 9.13, 36.77, 7.69)
	p.Curve3(37.01, 6.25, 37.57, 5.76)
	p.Curve3(38.13, 5.27, 38.87, 5.27)
	p.Curve3(39.65, 5.27, 40.23, 5.62)
	p.Curve3(41.26, 6.25, 44.19, 9.18)
	p.LineTo(44.19, 6.45)
	p.Curve3(38.72, -0.88, 33.74, -0.88)
	p.Curve3(31.35, -0.88, 29.93, 0.78)
	p.Curve3(28.52, 2.44, 28.47, 6.45)
	p.ClosePolygon(basics.PathFlagsNone)

	p.MoveTo(28.47, 9.62)
	p.LineTo(28.47, 26.66)
	p.Curve3(21.09, 23.73, 18.95, 22.51)
	p.Curve3(15.09, 20.36, 13.43, 18.02)
	p.Curve3(11.77, 15.67, 11.77, 12.89)
	p.Curve3(11.77, 9.38, 13.87, 7.06)
	p.Curve3(15.97, 4.74, 18.70, 4.74)
	p.Curve3(22.41, 4.74, 28.47, 9.62)
	p.ClosePolygon(basics.PathFlagsNone)

	// Apply scale(4.0) then translate(150, 100) to all vertices in-place.
	mtx := transform.NewTransAffine().Scale(4.0).Translate(150, 100)
	n := p.TotalVertices()
	for i := range n {
		x, y, cmd := p.Vertex(i)
		if basics.IsVertex(basics.PathCommand(cmd)) {
			mtx.Transform(&x, &y)
			p.ModifyVertex(i, x, y)
		}
	}

	return p
}

// gradientColors is the 256-entry gradient color LUT from blend_color.cpp.
// Each entry is {R, G, B} — alpha is computed separately.
var gradientColorsRGB = [256][3]uint8{
	{255, 255, 255}, {255, 255, 254}, {255, 255, 254}, {255, 255, 254},
	{255, 255, 253}, {255, 255, 253}, {255, 255, 252}, {255, 255, 251},
	{255, 255, 250}, {255, 255, 248}, {255, 255, 246}, {255, 255, 244},
	{255, 255, 241}, {255, 255, 238}, {255, 255, 235}, {255, 255, 231},
	{255, 255, 227}, {255, 255, 222}, {255, 255, 217}, {255, 255, 211},
	{255, 255, 206}, {255, 255, 200}, {255, 254, 194}, {255, 253, 188},
	{255, 252, 182}, {255, 250, 176}, {255, 249, 170}, {255, 247, 164},
	{255, 246, 158}, {255, 244, 152}, {254, 242, 146}, {254, 240, 141},
	{254, 238, 136}, {254, 236, 131}, {253, 234, 126}, {253, 232, 121},
	{253, 229, 116}, {252, 227, 112}, {252, 224, 108}, {251, 222, 104},
	{251, 219, 100}, {251, 216, 96}, {250, 214, 93}, {250, 211, 89},
	{249, 208, 86}, {249, 205, 83}, {248, 202, 80}, {247, 199, 77},
	{247, 196, 74}, {246, 193, 72}, {246, 190, 69}, {245, 187, 67},
	{244, 183, 64}, {244, 180, 62}, {243, 177, 60}, {242, 174, 58},
	{242, 170, 56}, {241, 167, 54}, {240, 164, 52}, {239, 161, 51},
	{239, 157, 49}, {238, 154, 47}, {237, 151, 46}, {236, 147, 44},
	{235, 144, 43}, {235, 141, 41}, {234, 138, 40}, {233, 134, 39},
	{232, 131, 37}, {231, 128, 36}, {230, 125, 35}, {229, 122, 34},
	{228, 119, 33}, {227, 116, 31}, {226, 113, 30}, {225, 110, 29},
	{224, 107, 28}, {223, 104, 27}, {222, 101, 26}, {221, 99, 25},
	{220, 96, 24}, {219, 93, 23}, {218, 91, 22}, {217, 88, 21},
	{216, 86, 20}, {215, 83, 19}, {214, 81, 18}, {213, 79, 17},
	{212, 77, 17}, {211, 74, 16}, {210, 72, 15}, {209, 70, 14},
	{207, 68, 13}, {206, 66, 13}, {205, 64, 12}, {204, 62, 11},
	{203, 60, 10}, {202, 58, 10}, {201, 56, 9}, {199, 55, 9},
	{198, 53, 8}, {197, 51, 7}, {196, 50, 7}, {195, 48, 6},
	{193, 46, 6}, {192, 45, 5}, {191, 43, 5}, {190, 42, 4},
	{188, 41, 4}, {187, 39, 3}, {186, 38, 3}, {185, 37, 2},
	{183, 35, 2}, {182, 34, 1}, {181, 33, 1}, {179, 32, 1},
	{178, 30, 0}, {177, 29, 0}, {175, 28, 0}, {174, 27, 0},
	{173, 26, 0}, {171, 25, 0}, {170, 24, 0}, {168, 23, 0},
	{167, 22, 0}, {165, 21, 0}, {164, 21, 0}, {163, 20, 0},
	{161, 19, 0}, {160, 18, 0}, {158, 17, 0}, {156, 17, 0},
	{155, 16, 0}, {153, 15, 0}, {152, 14, 0}, {150, 14, 0},
	{149, 13, 0}, {147, 12, 0}, {145, 12, 0}, {144, 11, 0},
	{142, 11, 0}, {140, 10, 0}, {139, 10, 0}, {137, 9, 0},
	{135, 9, 0}, {134, 8, 0}, {132, 8, 0}, {130, 7, 0},
	{128, 7, 0}, {126, 6, 0}, {125, 6, 0}, {123, 5, 0},
	{121, 5, 0}, {119, 4, 0}, {117, 4, 0}, {115, 4, 0},
	{113, 3, 0}, {111, 3, 0}, {109, 2, 0}, {107, 2, 0},
	{105, 2, 0}, {103, 1, 0}, {101, 1, 0}, {99, 1, 0},
	{97, 0, 0}, {95, 0, 0}, {93, 0, 0}, {91, 0, 0},
	{90, 0, 0}, {88, 0, 0}, {86, 0, 0}, {84, 0, 0},
	{82, 0, 0}, {80, 0, 0}, {78, 0, 0}, {77, 0, 0},
	{75, 0, 0}, {73, 0, 0}, {72, 0, 0}, {70, 0, 0},
	{68, 0, 0}, {67, 0, 0}, {65, 0, 0}, {64, 0, 0},
	{63, 0, 0}, {61, 0, 0}, {60, 0, 0}, {59, 0, 0},
	{58, 0, 0}, {57, 0, 0}, {56, 0, 0}, {55, 0, 0},
	{54, 0, 0}, {53, 0, 0}, {53, 0, 0}, {52, 0, 0},
	{52, 0, 0}, {51, 0, 0}, {51, 0, 0}, {51, 0, 0},
	{50, 0, 0}, {50, 0, 0}, {51, 0, 0}, {51, 0, 0},
	{51, 0, 0}, {51, 0, 0}, {52, 0, 0}, {52, 0, 0},
	{53, 0, 0}, {54, 1, 0}, {55, 2, 0}, {56, 3, 0},
	{57, 4, 0}, {58, 5, 0}, {59, 6, 0}, {60, 7, 0},
	{62, 8, 0}, {63, 9, 0}, {64, 11, 0}, {66, 12, 0},
	{68, 13, 0}, {69, 14, 0}, {71, 16, 0}, {73, 17, 0},
	{75, 18, 0}, {77, 20, 0}, {79, 21, 0}, {81, 23, 0},
	{83, 24, 0}, {85, 26, 0}, {87, 28, 0}, {90, 29, 0},
	{92, 31, 0}, {94, 33, 0}, {97, 34, 0}, {99, 36, 0},
	{102, 38, 0}, {104, 40, 0}, {107, 41, 0}, {109, 43, 0},
	{112, 45, 0}, {115, 47, 0}, {117, 49, 0}, {120, 51, 0},
	{123, 52, 0}, {126, 54, 0}, {128, 56, 0}, {131, 58, 0},
	{134, 60, 0}, {137, 62, 0}, {140, 64, 0}, {143, 66, 0},
	{145, 68, 0}, {148, 70, 0}, {151, 72, 0}, {154, 74, 0},
}

// buildColorLUT builds the 256-entry RGBA color LUT from the gradient data.
// Alpha is: i*4 for i<64, else 255 (matching C++ line 426).
func buildColorLUT() []color.RGBA8[color.SRGB] {
	lut := make([]color.RGBA8[color.SRGB], 256)
	for i := range 256 {
		rgb := gradientColorsRGB[i]
		a := uint8(255)
		if i <= 63 {
			a = uint8(min(i*4, 255))
		}
		lut[i] = color.RGBA8[color.SRGB]{R: basics.Int8u(rgb[0]), G: basics.Int8u(rgb[1]), B: basics.Int8u(rgb[2]), A: basics.Int8u(a)}
	}
	return lut
}

// Draw renders the blend color demo onto the given context.
// Returns the (potentially initialized) quad and shape bounds so
// the caller can track the draggable quad across frames.
func Draw(ctx *agg.Context, cfg *Config) Result {
	if ctx == nil {
		return Result{}
	}

	// Build the glyph path (letter "a", scaled 4x, translated to (150,100)).
	glyphPath := buildGlyphPath()

	// Wrap path with ConvCurve to flatten curve3 segments.
	shape := conv.NewConvCurve(&pathStorageAdapter{ps: glyphPath})

	// Compute bounding rect of the curved shape.
	shapeBounds, ok := basics.BoundingRectSingle[float64](shape, 0)
	if !ok {
		return Result{}
	}

	// Initialize quad to shape bounds if all zeros.
	quad := cfg.Quad
	allZero := true
	for _, v := range quad {
		if v != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		quad = [8]float64{
			shapeBounds.X1, shapeBounds.Y1, // top-left
			shapeBounds.X2, shapeBounds.Y1, // top-right
			shapeBounds.X2, shapeBounds.Y2, // bottom-right
			shapeBounds.X1, shapeBounds.Y2, // bottom-left
		}
	}

	// Get canvas dimensions and pixel data.
	outImg := ctx.GetImage()
	w, h := outImg.Width(), outImg.Height()

	outRbuf := buffer.NewRenderingBufferU8()
	outRbuf.Attach(outImg.Data, w, h, w*4)
	outPixFmt := pixfmt.NewPixFmtRGBA32PreLinear(outRbuf)
	renBase := renderer.NewRendererBaseWithPixfmt(outPixFmt)
	renBase.Clear(color.RGBA8[color.Linear]{R: 255, G: 242, B: 242, A: 255})

	// Create gray8 buffer for shadow rendering.
	grayBuf := make([]basics.Int8u, w*h)
	grayRbuf := buffer.NewRenderingBufferU8()
	grayRbuf.Attach(grayBuf, w, h, w)
	grayPixFmt := pixfmt.NewPixFmtSGray8(grayRbuf)
	grayRendBase := renderer.NewRendererBaseWithPixfmt(grayPixFmt)
	grayRendBase.Clear(color.Gray8[color.SRGB]{V: 0, A: 255})

	// Build perspective transform from shape bounds to quad.
	shadowPersp := transform.NewTransPerspectiveRectToQuad(
		shapeBounds.X1, shapeBounds.Y1,
		shapeBounds.X2, shapeBounds.Y2,
		quad,
	)

	// Apply perspective transform to the curved shape.
	shadowTrans := conv.NewConvTransform(shape, shadowPersp)

	// Create rasterizer and scanline.
	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	ras.ClipBox(0, 0, float64(w), float64(h))
	sl := scanline.NewScanlineP8()

	// Render shadow into gray8 buffer.
	ras.AddPath(&rasPathAdapter{vs: shadowTrans}, 0)
	slAdapter := &scanlineAdapter{sl: sl}
	if ras.RewindScanlines() {
		sl.Reset(ras.MinX(), ras.MaxX())
		grayColor := color.Gray8[color.SRGB]{V: 255, A: 255}
		for ras.SweepScanline(slAdapter) {
			renderScanlineGray8(sl, grayRendBase, grayColor)
		}
	}

	// Compute bounding box of the transformed shadow and extend by blur radius.
	bbox, bboxOk := basics.BoundingRectSingle[float64](shadowTrans, 0)
	if !bboxOk {
		bbox = basics.Rect[float64]{X1: 0, Y1: 0, X2: float64(w), Y2: float64(h)}
	}

	bbox.X1 -= cfg.Radius
	bbox.Y1 -= cfg.Radius
	bbox.X2 += cfg.Radius
	bbox.Y2 += cfg.Radius

	canvasRect := basics.Rect[float64]{X1: 0, Y1: 0, X2: float64(w), Y2: float64(h)}
	if bbox.Clip(canvasRect) {
		// Apply stack blur to the gray8 buffer.
		blurR := int(math.Round(cfg.Radius))
		if blurR > 0 {
			effects.StackBlurGray8(grayPixFmt, blurR, blurR)
		}

		// Blend the shadow onto the main RGBA canvas.
		// The shadow is rendered at correct canvas coordinates in the full gray8
		// buffer, so we use dx=0, dy=0 with a srcRect limiting the source to the
		// bbox region. (The C++ version uses a sub-view pixfmt attached to bbox,
		// then passes dx=bbox.X1, dy=bbox.Y1. We achieve the same effect with
		// a source rect and zero offset.)
		srcRect := &basics.RectI{
			X1: int(bbox.X1), Y1: int(bbox.Y1),
			X2: int(bbox.X2), Y2: int(bbox.Y2),
		}
		if cfg.Method == 0 {
			// Single color method: green shadow.
			greenColor := color.RGBA8[color.Linear]{R: 0, G: 100, B: 0, A: 255}
			renBase.BlendFromColor(grayPixFmt, greenColor, srcRect, 0, 0, 255)
		} else {
			// Color LUT method: gradient shadow.
			colorLUT := buildColorLUT()
			// Convert SRGB LUT to Linear for the linear renderer.
			linearLUT := make([]color.RGBA8[color.Linear], len(colorLUT))
			for i, c := range colorLUT {
				linearLUT[i] = color.RGBA8[color.Linear](c)
			}
			renBase.BlendFromLUT(grayPixFmt, linearLUT, srcRect, 0, 0, 255)
		}
	}

	// Render timing text using GsvText + ConvStroke.
	t := gsv.NewGSVText()
	t.SetFlip(true) // Our canvas is Y-down; GsvText font data is Y-up.
	t.SetSize(10.0, 0)
	t.SetStartPoint(140.0, 30.0)
	t.SetText(fmt.Sprintf("Blend Color Demo (radius=%.1f)", cfg.Radius))

	st := conv.NewConvStroke(t)
	st.SetWidth(1.5)

	ras.Reset()
	ras.AddPath(&rasPathAdapter{vs: st}, 0)
	if ras.RewindScanlines() {
		sl.Reset(ras.MinX(), ras.MaxX())
		textColor := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255}
		for ras.SweepScanline(slAdapter) {
			renderScanlineRGBA(sl, renBase, textColor)
		}
	}

	return Result{
		Quad:        quad,
		ShapeBounds: [4]float64{shapeBounds.X1, shapeBounds.Y1, shapeBounds.X2, shapeBounds.Y2},
	}
}

// renderScanlineGray8 renders a single scanline with solid gray color.
// This is an inline version of RenderScanlineAASolid for gray8 renderer.
func renderScanlineGray8(
	sl *scanline.ScanlineP8,
	ren *renderer.RendererBase[*pixfmt.PixFmtSGray8, color.Gray8[color.SRGB]],
	c color.Gray8[color.SRGB],
) {
	y := sl.Y()
	spans := sl.Begin()

	for _, span := range spans {
		x := int(span.X)
		length := int(span.Len)

		if length > 0 {
			ren.BlendSolidHspan(x, y, length, c, span.Covers)
		} else {
			endX := x - length - 1
			cover := span.Covers[0]
			ren.BlendHline(x, y, endX, c, cover)
		}
	}
}

// renderScanlineRGBA renders a single scanline with solid RGBA color.
func renderScanlineRGBA(
	sl *scanline.ScanlineP8,
	ren *renderer.RendererBase[*pixfmt.PixFmtRGBA32Pre[color.Linear], color.RGBA8[color.Linear]],
	c color.RGBA8[color.Linear],
) {
	y := sl.Y()
	spans := sl.Begin()

	for _, span := range spans {
		x := int(span.X)
		length := int(span.Len)

		if length > 0 {
			ren.BlendSolidHspan(x, y, length, c, span.Covers)
		} else {
			endX := x - length - 1
			cover := span.Covers[0]
			ren.BlendHline(x, y, endX, c, cover)
		}
	}
}
