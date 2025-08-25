// Rasterizers2 Demo - Direct port from AGG C++ rasterizers2.cpp
//
// This example demonstrates different rasterization techniques in AGG:
// 1. Aliased lines with pixel accuracy (Bresenham)
// 2. Aliased lines with subpixel accuracy (Bresenham)  
// 3. Anti-aliased outline rendering
// 4. Anti-aliased scanline rendering with stroke
// 5. Anti-aliased outline with image patterns
//
// The demo renders spiral shapes using each technique and allows interactive
// parameter adjustment and performance testing.

package main

import (
	"fmt"
	"math"

	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/ctrl/checkbox"
	"agg_go/internal/ctrl/slider"
	"agg_go/internal/pixfmt"
	"agg_go/internal/platform"
	"agg_go/internal/renderer"
)

// Chain link pattern data - direct port from C++ pixmap_chain array
var pixmapChain = []uint32{
	16, 7, // width, height
	0x00ffffff, 0x00ffffff, 0x00ffffff, 0x00ffffff, 0xb4c29999, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xb4c29999, 0x00ffffff, 0x00ffffff, 0x00ffffff, 0x00ffffff,
	0x00ffffff, 0x00ffffff, 0x0cfbf9f9, 0xff9a5757, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xb4c29999, 0x00ffffff, 0x00ffffff, 0x00ffffff,
	0x00ffffff, 0x5ae0cccc, 0xffa46767, 0xff660000, 0xff975252, 0x7ed4b8b8, 0x5ae0cccc, 0x5ae0cccc, 0x5ae0cccc, 0x5ae0cccc, 0xa8c6a0a0, 0xff7f2929, 0xff670202, 0x9ecaa6a6, 0x5ae0cccc, 0x00ffffff,
	0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xa4c7a2a2, 0x3affff00, 0x3affff00, 0xff975151, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000,
	0x00ffffff, 0x5ae0cccc, 0xffa46767, 0xff660000, 0xff954f4f, 0x7ed4b8b8, 0x5ae0cccc, 0x5ae0cccc, 0x5ae0cccc, 0x5ae0cccc, 0xa8c6a0a0, 0xff7f2929, 0xff670202, 0x9ecaa6a6, 0x5ae0cccc, 0x00ffffff,
	0x00ffffff, 0x00ffffff, 0x0cfbf9f9, 0xff9a5757, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xb4c29999, 0x00ffffff, 0x00ffffff, 0x00ffffff,
	0x00ffffff, 0x00ffffff, 0x00ffffff, 0x00ffffff, 0xb4c29999, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xb4c29999, 0x00ffffff, 0x00ffffff, 0x00ffffff, 0x00ffffff,
}

// PatternPixmapARGB32 implements a pattern source interface for the chain link pattern
type PatternPixmapARGB32 struct {
	pixmap []uint32
}

func NewPatternPixmapARGB32(pixmap []uint32) *PatternPixmapARGB32 {
	return &PatternPixmapARGB32{pixmap: pixmap}
}

func (p *PatternPixmapARGB32) Width() int {
	if len(p.pixmap) < 1 {
		return 0
	}
	return int(p.pixmap[0])
}

func (p *PatternPixmapARGB32) Height() int {
	if len(p.pixmap) < 2 {
		return 0
	}
	return int(p.pixmap[1])
}

func (p *PatternPixmapARGB32) Pixel(x, y int) color.RGBA8Linear {
	if x < 0 || y < 0 || x >= p.Width() || y >= p.Height() {
		return color.RGBA8Linear{R: 0, G: 0, B: 0, A: 0}
	}

	idx := y*p.Width() + x + 2 // +2 to skip width/height
	if idx >= len(p.pixmap) {
		return color.RGBA8Linear{R: 0, G: 0, B: 0, A: 0}
	}

	pixel := p.pixmap[idx]
	// ARGB32 format: extract components
	a := basics.Int8u((pixel >> 24) & 0xFF)
	r := basics.Int8u((pixel >> 16) & 0xFF)
	g := basics.Int8u((pixel >> 8) & 0xFF)
	b := basics.Int8u(pixel & 0xFF)

	// Return as RGBA
	return color.RGBA8Linear{R: r, G: g, B: b, A: a}
}

// Spiral generates a spiral path - direct port from C++ spiral class
type Spiral struct {
	x, y        float64
	r1, r2      float64
	step        float64
	startAngle  float64
	angle       float64
	currR       float64
	da          float64
	dr          float64
	start       bool
}

func NewSpiral(x, y, r1, r2, step, startAngle float64) *Spiral {
	return &Spiral{
		x:          x,
		y:          y,
		r1:         r1,
		r2:         r2,
		step:       step,
		startAngle: startAngle,
		da:         basics.Deg2RadF(8.0),
		dr:         step / 45.0,
	}
}

func (s *Spiral) Rewind(pathID uint) {
	s.angle = s.startAngle
	s.currR = s.r1
	s.start = true
}

func (s *Spiral) Vertex(x, y *float64) uint32 {
	if s.currR > s.r2 {
		return uint32(basics.PathCmdStop)
	}

	*x = s.x + math.Cos(s.angle)*s.currR
	*y = s.y + math.Sin(s.angle)*s.currR
	s.currR += s.dr
	s.angle += s.da

	if s.start {
		s.start = false
		return uint32(basics.PathCmdMoveTo)
	}
	return uint32(basics.PathCmdLineTo)
}

// Roundoff transformer for pixel accuracy demo - direct port from C++ roundoff struct
type Roundoff struct{}

func (r *Roundoff) Transform(x, y *float64) {
	*x = math.Floor(*x)
	*y = math.Floor(*y)
}

// Application holds the demo application state
type Application struct {
	ps         *platform.PlatformSupport
	rc         *platform.RenderingContext
	stepSlider *slider.SliderCtrl
	widthSlider *slider.SliderCtrl
	testBox    *checkbox.CheckboxCtrl[color.RGBA8Linear]
	rotateBox  *checkbox.CheckboxCtrl[color.RGBA8Linear]
	joinBox    *checkbox.CheckboxCtrl[color.RGBA8Linear]
	scaleBox   *checkbox.CheckboxCtrl[color.RGBA8Linear]
	startAngle float64
}

func NewApplication() *Application {
	app := &Application{}

	// Create platform support
	app.ps = platform.NewPlatformSupport(platform.PixelFormatRGBA32, false)
	app.ps.Caption("AGG Rasterizers2 Demo")
	app.rc = platform.NewRenderingContext(app.ps)

	// Create controls - positions match C++ version
	app.stepSlider = slider.NewSliderCtrl(10, 10+4, 150, 10+8+4, false)
	app.stepSlider.SetRange(0.0, 2.0)
	app.stepSlider.SetValue(0.1)
	app.stepSlider.SetLabel("Step=%.2f")

	app.widthSlider = slider.NewSliderCtrl(150+10, 10+4, 400-10, 10+8+4, false)
	app.widthSlider.SetRange(0.0, 14.0)
	app.widthSlider.SetValue(3.0)
	app.widthSlider.SetLabel("Width=%.2f")

	// Define colors for controls
	black := color.RGBA8Linear{R: 0, G: 0, B: 0, A: 255}
	gray := color.RGBA8Linear{R: 100, G: 100, B: 100, A: 255}
	blue := color.RGBA8Linear{R: 0, G: 0, B: 200, A: 255}

	app.testBox = checkbox.NewCheckboxCtrl(10, 10+4+16, "Test Performance", false, gray, black, blue)
	app.rotateBox = checkbox.NewCheckboxCtrl(130+10, 10+4+16, "Rotate", false, gray, black, blue)
	app.joinBox = checkbox.NewCheckboxCtrl(200+10, 10+4+16, "Accurate Joins", false, gray, black, blue)
	app.scaleBox = checkbox.NewCheckboxCtrl(310+10, 10+4+16, "Scale Pattern", false, gray, black, blue)
	app.scaleBox.SetChecked(true)
	
	// Enable rotation for testing
	app.rotateBox.SetChecked(true)

	return app
}

func (app *Application) OnDraw() {
	width := float64(app.ps.Width())
	height := float64(app.ps.Height())
	
	// Get rendering buffer from rendering context
	rbuf := app.rc.WindowBuffer()
	
	// Use RGBA32 format to match the renderer interface 
	pixf := pixfmt.NewPixFmtRGBA32(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt(pixf)
	
	// Clear background - cream color (1.0, 1.0, 0.95)
	clearColor := color.RGBA8Linear{R: 255, G: 255, B: 242, A: 255}
	renBase.Clear(clearColor)

	// Set up color for drawing
	drawColor := color.RGBA8Linear{R: 102, G: 77, B: 26, A: 255} // rgba(0.4, 0.3, 0.1)

	// For now, draw simple test spirals to show the basic structure
	// These represent the five different rendering techniques from the original
	
	// 1. Aliased pixel accuracy (top-left)
	app.drawSimpleSpiral(renBase, width/5, height/4+50, drawColor, "pixel")
	
	// 2. Aliased subpixel accuracy (top-center)
	app.drawSimpleSpiral(renBase, width/2, height/4+50, drawColor, "subpixel")
	
	// 3. Anti-aliased outline (bottom-left)
	app.drawSimpleSpiral(renBase, width/5, height-height/4+20, drawColor, "outline")
	
	// 4. Anti-aliased scanline (bottom-center)
	app.drawSimpleSpiral(renBase, width/2, height-height/4+20, drawColor, "scanline")
	
	// 5. Anti-aliased with pattern (bottom-right) 
	app.drawSimpleSpiral(renBase, width-width/5, height-height/4+20, drawColor, "pattern")
	
	fmt.Printf("Drew 5 test spirals, angle %.2f\n", app.startAngle)

	// TODO: Add text rendering for labels
}

// drawSimpleSpiral draws a simple spiral using pixel plotting
func (app *Application) drawSimpleSpiral(renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32, color.RGBA8Linear], centerX, centerY float64, color color.RGBA8Linear, technique string) {
	// Create spiral with different parameters based on technique
	r1, r2, step := 5.0, 70.0, 16.0
	if technique == "pattern" {
		// Make pattern spiral slightly different for variety
		r1, r2, step = 3.0, 60.0, 12.0
	}
	
	spiral := NewSpiral(centerX, centerY, r1, r2, step, app.startAngle)
	
	// Different rendering styles for each technique (simplified)
	var x, y float64
	spiral.Rewind(0)
	
	switch technique {
	case "pixel":
		// Rounded coordinates (pixel accuracy)
		for {
			cmd := spiral.Vertex(&x, &y)
			if cmd == uint32(basics.PathCmdStop) {
				break
			}
			if cmd == uint32(basics.PathCmdMoveTo) || cmd == uint32(basics.PathCmdLineTo) {
				px, py := int(math.Floor(x)), int(math.Floor(y))
				if px >= 0 && py >= 0 && px < renBase.Width() && py < renBase.Height() {
					renBase.BlendPixel(px, py, color, basics.CoverFull)
				}
			}
		}
		
	case "subpixel":
		// Subpixel accuracy (no rounding)
		for {
			cmd := spiral.Vertex(&x, &y)
			if cmd == uint32(basics.PathCmdStop) {
				break
			}
			if cmd == uint32(basics.PathCmdMoveTo) || cmd == uint32(basics.PathCmdLineTo) {
				px, py := int(x), int(y)
				if px >= 0 && py >= 0 && px < renBase.Width() && py < renBase.Height() {
					renBase.BlendPixel(px, py, color, basics.CoverFull)
				}
			}
		}
		
	case "outline":
		// Anti-aliased (thicker points)
		for {
			cmd := spiral.Vertex(&x, &y)
			if cmd == uint32(basics.PathCmdStop) {
				break
			}
			if cmd == uint32(basics.PathCmdMoveTo) || cmd == uint32(basics.PathCmdLineTo) {
				px, py := int(x), int(y)
				// Draw thicker points for "outline" effect
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						if px+dx >= 0 && py+dy >= 0 && px+dx < renBase.Width() && py+dy < renBase.Height() {
							renBase.BlendPixel(px+dx, py+dy, color, basics.CoverFull/2)
						}
					}
				}
			}
		}
		
	default: // "scanline" and "pattern"
		// Regular points
		for {
			cmd := spiral.Vertex(&x, &y)
			if cmd == uint32(basics.PathCmdStop) {
				break
			}
			if cmd == uint32(basics.PathCmdMoveTo) || cmd == uint32(basics.PathCmdLineTo) {
				px, py := int(x), int(y)
				if px >= 0 && py >= 0 && px < renBase.Width() && py < renBase.Height() {
					renBase.BlendPixel(px, py, color, basics.CoverFull)
				}
			}
		}
	}
}

/*
// TODO: These rendering methods need interface adaptation work
// 1. Draw aliased lines with pixel accuracy (rounded coordinates)
func (app *Application) drawAliasedPixelAccuracy(renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32, color.RGBA8Linear], width, height float64, color color.RGBA8Linear) {
	spiral := NewSpiral(width/5, height/4+50, 5, 70, 16, app.startAngle)
	roundoff := &Roundoff{}
	trans := conv.NewConvTransform(spiral, roundoff)
	
	// Create primitive renderer
	renPrim := primitives.NewRendererPrimitives(renBase)
	renPrim.LineColor(color)
	
	// Create outline rasterizer  
	rasOutline := rasterizer.NewRasterizerOutline(renPrim)
	rasOutline.AddPath(trans, 0)
}

// 2. Draw aliased lines with subpixel accuracy  
func (app *Application) drawAliasedSubpixelAccuracy(renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32, color.RGBA8Linear], width, height float64, color color.RGBA8Linear) {
	spiral := NewSpiral(width/2, height/4+50, 5, 70, 16, app.startAngle)
	
	// Create primitive renderer
	renPrim := primitives.NewRendererPrimitives(renBase)
	renPrim.LineColor(color)
	
	// Create outline rasterizer
	rasOutline := rasterizer.NewRasterizerOutline(renPrim)
	rasOutline.AddPath(spiral, 0)
}

// 3. Draw anti-aliased outline
func (app *Application) drawAntiAliasedOutline(renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32, color.RGBA8Linear], width, height float64, color color.RGBA8Linear) {
	spiral := NewSpiral(width/5, height-height/4+20, 5, 70, 16, app.startAngle)
	
	// Create line profile and outline renderer
	lineProfile := outline.NewLineProfileAA()
	lineProfile.SetWidth(app.widthSlider.Value())
	
	renOutlineAA := outline.NewRendererOutlineAA(renBase, lineProfile)
	renOutlineAA.SetColor(color)
	
	// Create anti-aliased outline rasterizer
	rasOutlineAA := rasterizer.NewRasterizerOutlineAA(renOutlineAA)
	// TODO: Set line join mode based on checkbox
	rasOutlineAA.AddPath(spiral, 0)
}

// 4. Draw anti-aliased scanline with stroke
func (app *Application) drawAntiAliasedScanline(renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32, color.RGBA8Linear], width, height float64, color color.RGBA8Linear) {
	spiral := NewSpiral(width/2, height-height/4+20, 5, 70, 16, app.startAngle)
	
	// Create stroke converter
	stroke := conv.NewConvStroke(spiral)
	stroke.SetWidth(app.widthSlider.Value())
	stroke.SetLineCap(basics.RoundCap)
	
	// Create scanline rasterizer and renderer
	rasScanline := rasterizer.NewRasterizerScanlineAA()
	sl := scanline.NewScanlineP8()
	renScanline := scanline.NewRendererScanlineAASolid(renBase)
	renScanline.SetColor(color)
	
	rasScanline.AddPath(stroke, 0)
	scanline.RenderScanlines(rasScanline, sl, renScanline)
}

// 5. Draw anti-aliased outline with image pattern
func (app *Application) drawAntiAliasedOutlineImg(renBase *renderer.RendererBase[*pixfmt.PixFmtRGBA32, color.RGBA8Linear], width, height float64) {
	spiral := NewSpiral(width-width/5, height-height/4+20, 5, 70, 16, app.startAngle)
	
	// TODO: Implement image pattern rendering
	// This requires PatternFilter, LineImagePattern, and RendererOutlineImage
	fmt.Println("Image pattern rendering not yet implemented")
}
*/

func main() {
	fmt.Println("AGG Rasterizers2 Demo")
	fmt.Println("=====================")
	
	app := NewApplication()

	// Set up event handlers
	app.ps.SetOnInit(app.OnInit)
	app.ps.SetOnResize(app.OnResize)
	app.ps.SetOnDraw(app.OnDraw)
	app.ps.SetOnIdle(app.OnIdle)

	if app.ps.Init(500, 450, 0) == nil {
		app.ps.Run()
	}
}

// Platform event handlers
func (app *Application) OnInit() {
	fmt.Println("Rasterizers2 Demo initialized")
}

func (app *Application) OnResize(width, height int) {
	app.rc.SetupResizeTransform(width, height)
}

func (app *Application) OnIdle() {
	if app.rotateBox.IsChecked() {
		app.startAngle += basics.Deg2RadF(app.stepSlider.Value())
		if app.startAngle > basics.Deg2RadF(360.0) {
			app.startAngle -= basics.Deg2RadF(360.0)
		}
		fmt.Printf("Animation: angle %.2f\n", app.startAngle)
		app.ps.ForceRedraw()
		app.ps.UpdateWindow()
	}
}

func (app *Application) OnCtrlChange() {
	// Handle performance test
	if app.testBox.IsChecked() {
		app.performanceTest()
		app.testBox.SetChecked(false)
		app.ps.ForceRedraw()
	}
}

func (app *Application) performanceTest() {
	fmt.Println("Running performance test...")
	// TODO: Implement performance testing
}