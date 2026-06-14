package scene

import (
	"fmt"
	"math"
	"os"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/engine"
)

// canvas is the fixed output size for every corpus scene.
const canvasW, canvasH = 256, 256

// corpus is the static, deterministically-ordered scene list. It is a plain
// slice (no init-based registration) so iteration order is stable across runs.
var corpus = []Scene{
	{
		Name:   "solid_fill_stroke",
		Width:  canvasW,
		Height: canvasH,
		Caps:   []engine.Capability{engine.CapabilitySolidStyle, engine.CapabilityPath},
		Draw: func(ctx engine.Context, _ *Assets) error {
			ctx.Clear(agg.White)

			ctx.SetFillColor(agg.NewColorRGB(200, 30, 30))
			ctx.FillRectangle(30, 30, 90, 70)

			ctx.SetStrokeColor(agg.NewColorRGB(30, 60, 200))
			ctx.SetLineWidth(6)
			ctx.SetLineCap(agg.CapRound)
			ctx.SetLineJoin(agg.JoinBevel)
			ctx.DrawRectangle(150, 40, 70, 70)

			// Open stroked polyline exercising caps and joins.
			ctx.SetStrokeColor(agg.NewColorRGB(20, 130, 30))
			ctx.SetLineWidth(8)
			ctx.SetLineCap(agg.CapSquare)
			ctx.SetLineJoin(agg.JoinMiter)
			ctx.BeginPath()
			ctx.MoveTo(40, 200)
			ctx.LineTo(110, 140)
			ctx.LineTo(170, 210)
			ctx.LineTo(220, 150)
			ctx.Stroke()
			return nil
		},
	},
	{
		Name:   "dashed_stroke",
		Width:  canvasW,
		Height: canvasH,
		Caps:   []engine.Capability{engine.CapabilitySolidStyle, engine.CapabilityPath, engine.CapabilityDashedStroke},
		Draw: func(ctx engine.Context, _ *Assets) error {
			ctx.Clear(agg.White)
			// No transform: dash lengths are in device space on both backends, so
			// the dash segmentation lines up (the port applies dashes in user
			// space, the CPP backend on the pre-transformed device path).
			ctx.SetStrokeColor(agg.NewColorRGB(30, 60, 200))
			ctx.SetLineWidth(5)
			ctx.SetLineCap(agg.CapButt)
			ctx.SetLineJoin(agg.JoinMiter)
			ctx.AddDash(18, 10)
			ctx.AddDash(6, 10)
			ctx.DashStart(4)
			ctx.BeginPath()
			ctx.MoveTo(30, 60)
			ctx.LineTo(226, 60)
			ctx.LineTo(30, 130)
			ctx.LineTo(226, 130)
			ctx.Stroke()

			// A second, dash-reset solid stroke confirms RemoveAllDashes returns to
			// solid stroking on both backends.
			ctx.RemoveAllDashes()
			ctx.SetStrokeColor(agg.NewColorRGB(200, 40, 40))
			ctx.SetLineWidth(4)
			ctx.BeginPath()
			ctx.MoveTo(30, 200)
			ctx.LineTo(226, 200)
			ctx.Stroke()
			return nil
		},
	},
	{
		Name:   "path_nonzero",
		Width:  canvasW,
		Height: canvasH,
		Caps:   []engine.Capability{engine.CapabilitySolidStyle, engine.CapabilityPath},
		Draw: func(ctx engine.Context, _ *Assets) error {
			ctx.Clear(agg.White)
			ctx.FillEvenOdd(false)
			ctx.SetFillColor(agg.NewColorRGB(40, 120, 200))
			drawPentagram(ctx)
			ctx.Fill()
			return nil
		},
	},
	{
		Name:   "path_evenodd",
		Width:  canvasW,
		Height: canvasH,
		Caps:   []engine.Capability{engine.CapabilitySolidStyle, engine.CapabilityPath},
		Draw: func(ctx engine.Context, _ *Assets) error {
			ctx.Clear(agg.White)
			ctx.FillEvenOdd(true)
			ctx.SetFillColor(agg.NewColorRGB(40, 120, 200))
			drawPentagram(ctx)
			ctx.Fill()
			return nil
		},
	},
	{
		Name:   "clip_box",
		Width:  canvasW,
		Height: canvasH,
		Caps:   []engine.Capability{engine.CapabilityClipBox},
		Draw: func(ctx engine.Context, _ *Assets) error {
			ctx.Clear(agg.White)
			ctx.ClipBox(60, 60, 196, 196)
			// A fill that overflows the clip box on every side.
			ctx.SetFillColor(agg.NewColorRGB(210, 80, 30))
			ctx.FillRectangle(20, 20, 220, 220)
			return nil
		},
	},
	{
		Name:   "gradient_linear",
		Width:  canvasW,
		Height: canvasH,
		Caps:   []engine.Capability{engine.CapabilityGradients},
		Draw: func(ctx engine.Context, _ *Assets) error {
			ctx.Clear(agg.White)
			ctx.SetLinearGradient(40, 40, 216, 216,
				agg.NewColorRGB(220, 40, 40), agg.NewColorRGB(40, 60, 220))
			ctx.FillRectangle(40, 40, 176, 176)
			return nil
		},
	},
	{
		Name:   "gradient_radial",
		Width:  canvasW,
		Height: canvasH,
		Caps:   []engine.Capability{engine.CapabilityGradients},
		Draw: func(ctx engine.Context, _ *Assets) error {
			ctx.Clear(agg.White)
			ctx.SetRadialGradient(128, 128, 90,
				agg.NewColorRGB(240, 220, 40), agg.NewColorRGB(40, 40, 120))
			ctx.FillCircle(128, 128, 90)
			return nil
		},
	},
	{
		Name:   "compositing_srcover",
		Width:  canvasW,
		Height: canvasH,
		Caps:   []engine.Capability{engine.CapabilityCompositing},
		Draw: func(ctx engine.Context, _ *Assets) error {
			ctx.Clear(agg.White)
			ctx.SetBlendMode(agg.BlendSrcOver)
			ctx.SetFillColor(agg.NewColor(220, 40, 40, 160))
			ctx.FillRectangle(50, 50, 110, 110)
			ctx.SetFillColor(agg.NewColor(40, 60, 220, 160))
			ctx.FillRectangle(110, 110, 110, 110)
			return nil
		},
	},
	{
		Name:   "compositing_src",
		Width:  canvasW,
		Height: canvasH,
		Caps:   []engine.Capability{engine.CapabilityCompositing},
		Draw: func(ctx engine.Context, _ *Assets) error {
			ctx.Clear(agg.White)
			ctx.SetFillColor(agg.NewColorRGB(220, 40, 40))
			ctx.FillRectangle(50, 50, 110, 110)
			ctx.SetBlendMode(agg.BlendSrc)
			ctx.SetFillColor(agg.NewColor(40, 60, 220, 160))
			ctx.FillRectangle(110, 110, 110, 110)
			return nil
		},
	},
	{
		Name:   "compositing_clear",
		Width:  canvasW,
		Height: canvasH,
		Caps:   []engine.Capability{engine.CapabilityCompositing},
		Draw: func(ctx engine.Context, _ *Assets) error {
			ctx.Clear(agg.White)
			ctx.SetFillColor(agg.NewColorRGB(40, 130, 60))
			ctx.FillRectangle(40, 40, 176, 176)
			// Knock a hole out of the filled region.
			ctx.SetBlendMode(agg.BlendClear)
			ctx.SetFillColor(agg.NewColor(0, 0, 0, 255))
			ctx.FillRectangle(100, 100, 56, 56)
			return nil
		},
	},
	{
		Name:   "compositing_gradient",
		Width:  canvasW,
		Height: canvasH,
		Caps:   []engine.Capability{engine.CapabilityGradients, engine.CapabilityCompositing},
		Draw: func(ctx engine.Context, _ *Assets) error {
			ctx.Clear(agg.White)
			// Opaque background block that must survive outside the gradient shape:
			// a non-src-over blend that wiped the whole clip rectangle would erase it.
			ctx.SetFillColor(agg.NewColorRGB(40, 130, 60))
			ctx.FillRectangle(30, 30, 150, 150)
			// Gradient fill under a non-src-over blend. Src replaces the destination
			// only within the shape's coverage; the background outside the circle is
			// left untouched.
			ctx.SetBlendMode(agg.BlendSrc)
			ctx.SetLinearGradient(96, 96, 220, 220,
				agg.NewColor(220, 40, 40, 160), agg.NewColor(40, 60, 220, 160))
			ctx.FillCircle(140, 140, 70)
			return nil
		},
	},
	{
		// Separable (non-Porter-Duff) blend mode on the vector path. Multiply
		// darkens where the two opaque rectangles overlap and where the second
		// rectangle covers the white background, exercising AGG's comp_op_multiply.
		Name:   "compositing_multiply",
		Width:  canvasW,
		Height: canvasH,
		Caps:   []engine.Capability{engine.CapabilityCompositing},
		Draw: func(ctx engine.Context, _ *Assets) error {
			ctx.Clear(agg.White)
			ctx.SetFillColor(agg.NewColorRGB(200, 120, 40))
			ctx.FillRectangle(50, 50, 110, 110)
			ctx.SetBlendMode(agg.BlendMultiply)
			ctx.SetFillColor(agg.NewColorRGB(120, 200, 160))
			ctx.FillRectangle(110, 110, 110, 110)
			return nil
		},
	},
	{
		// Porter-Duff xor on the vector path: a translucent source over the opaque
		// white canvas reduces to dst*(1-Sa) (Da is 1), exercising comp_op_xor.
		Name:   "compositing_xor",
		Width:  canvasW,
		Height: canvasH,
		Caps:   []engine.Capability{engine.CapabilityCompositing},
		Draw: func(ctx engine.Context, _ *Assets) error {
			ctx.Clear(agg.White)
			ctx.SetFillColor(agg.NewColorRGB(40, 130, 60))
			ctx.FillRectangle(40, 40, 176, 176)
			ctx.SetBlendMode(agg.BlendXor)
			ctx.SetFillColor(agg.NewColor(40, 60, 220, 160))
			ctx.FillCircle(128, 128, 70)
			return nil
		},
	},
	{
		Name:   "image_scaled",
		Width:  canvasW,
		Height: canvasH,
		Caps:   []engine.Capability{engine.CapabilityImageDraw},
		Draw: func(ctx engine.Context, assets *Assets) error {
			if assets == nil || assets.Source == nil {
				return ErrAssetUnavailable
			}
			ctx.Clear(agg.White)
			// Scaled image draw with no active transform: supported by both
			// backends, so it exercises cross-backend image sampling directly.
			return ctx.DrawImageScaled(assets.Source, 32, 32, 192, 192)
		},
	},
	{
		Name:   "image_affine",
		Width:  canvasW,
		Height: canvasH,
		Caps:   []engine.Capability{engine.CapabilityImageDraw, engine.CapabilityTransforms},
		Draw: func(ctx engine.Context, assets *Assets) error {
			if assets == nil || assets.Source == nil {
				return ErrAssetUnavailable
			}
			ctx.Clear(agg.White)
			// Rotate/scale about the canvas centre, then draw the source scaled.
			// DrawImageScaled (not DrawImageRegion) is used so the CPP backend,
			// which rejects DrawImageRegion under an active transform, can run it.
			ctx.Translate(128, 128)
			ctx.Rotate(20 * math.Pi / 180)
			ctx.Scale(1.6, 1.6)
			ctx.Translate(-64, -64)
			return ctx.DrawImageScaled(assets.Source, 0, 0, 128, 128)
		},
	},
	{
		Name:   "text_basic",
		Width:  canvasW,
		Height: canvasH,
		Caps:   []engine.Capability{engine.CapabilityText},
		Draw: func(ctx engine.Context, _ *Assets) error {
			font := resolveFont()
			if font == "" {
				return ErrAssetUnavailable
			}
			ctx.Clear(agg.White)
			if err := ctx.LoadFont(font); err != nil {
				// A backend without font support (e.g. the port built without
				// the freetype tag) makes this scene a skip, not a failure.
				return fmt.Errorf("%w: load font %s: %v", ErrAssetUnavailable, font, err)
			}
			ctx.SetResolution(72)
			ctx.SetColor(agg.NewColorRGB(20, 20, 20))
			return ctx.DrawText("Agg 123", 30, 140)
		},
	},
}

// drawPentagram emits a self-intersecting five-point star path centred on the
// canvas. Filling it with the non-zero rule produces a solid star; the even-odd
// rule leaves the central pentagon empty.
func drawPentagram(ctx engine.Context) {
	const (
		cx, cy = 128.0, 128.0
		r      = 100.0
	)
	ctx.BeginPath()
	// Connect every second vertex of a regular pentagon to self-intersect.
	for i := range 5 {
		// Start at the top (-90°) and step by 144° (2 vertices) each time.
		angle := -math.Pi/2 + float64(i)*4*math.Pi/5
		x := cx + r*math.Cos(angle)
		y := cy + r*math.Sin(angle)
		if i == 0 {
			ctx.MoveTo(x, y)
		} else {
			ctx.LineTo(x, y)
		}
	}
	ctx.ClosePath()
}

// fontCandidates lists font files probed for the text scene, after the
// AGG_TEST_FONT override. The list mirrors the conventions already used by the
// agg2d text tests and the aggreal engine test.
var fontCandidates = []string{
	"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
	"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf",
	"/usr/share/fonts/TTF/DejaVuSans.ttf",
	"/usr/share/fonts/dejavu-sans-fonts/DejaVuSans.ttf",
	"/usr/share/fonts/truetype/noto/NotoSans-Regular.ttf",
}

// resolveFont returns a usable font file path, honouring the AGG_TEST_FONT
// environment override, or "" when none is found.
func resolveFont() string {
	if p := os.Getenv("AGG_TEST_FONT"); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p
		}
		return ""
	}
	for _, p := range fontCandidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
