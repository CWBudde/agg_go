//go:build agogo && cgo

package engine

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"math"
	"os"

	agg "github.com/cwbudde/agg_go"
)

type cppImage struct {
	img *cppNativeImage
}

type cppGradientKind int

const (
	cppGradientSolid cppGradientKind = iota
	cppGradientLinear
	cppGradientRadial
)

type cppGradientState struct {
	kind    cppGradientKind
	c1      agg.Color
	c2      agg.Color
	profile float64
	x1      float64
	y1      float64
	x2      float64
	y2      float64
	cx      float64
	cy      float64
	radius  float64
}

type cppContext struct {
	img                *cppImage
	path               *cppNativePath
	transform          *cppNativeMatrix
	font               *cppNativeFont
	transformDirty     bool
	currentX           float64
	currentY           float64
	subpathStartX      float64
	subpathStartY      float64
	hasCurrentPoint    bool
	fillColor          agg.Color
	strokeColor        agg.Color
	fillGradient       cppGradientState
	strokeGradient     cppGradientState
	lineWidth          float64
	lineCap            agg.LineCap
	lineJoin           agg.LineJoin
	dashes             []float32 // flat (dashLen, gapLen) pairs; empty strokes solid
	dashStart          float64
	blendMode          agg.BlendMode
	fillEvenOdd        bool
	fillGradientType   agg.GradientType
	strokeGradientType agg.GradientType
	textHints          bool
	textAlignX         agg.TextAlignment
	textAlignY         agg.TextAlignment
	textResolution     uint
	clipBox            agg.RectD
}

func newCPPBackendImage(width, height int) (*cppImage, error) {
	img, err := newCPPNativeImage(width, height)
	if err != nil {
		return nil, err
	}
	return &cppImage{img: img}, nil
}

func newCPPBackendContext(width, height int) (*cppContext, error) {
	img, err := newCPPBackendImage(width, height)
	if err != nil {
		return nil, err
	}
	return newCPPBackendContextForImage(img)
}

func newCPPBackendContextForImage(img *cppImage) (*cppContext, error) {
	if img == nil || img.img == nil {
		return nil, fmt.Errorf("image is nil")
	}
	path, err := newCPPNativePath()
	if err != nil {
		return nil, err
	}
	transform, err := newCPPNativeMatrix()
	if err != nil {
		_ = path.close()
		return nil, err
	}
	return &cppContext{
		img:                img,
		path:               path,
		transform:          transform,
		fillColor:          agg.Black,
		strokeColor:        agg.Black,
		fillGradient:       cppGradientState{kind: cppGradientSolid, c1: agg.Black, c2: agg.Black, profile: 1},
		strokeGradient:     cppGradientState{kind: cppGradientSolid, c1: agg.Black, c2: agg.Black, profile: 1},
		lineWidth:          1,
		lineCap:            agg.CapButt,
		lineJoin:           agg.JoinMiter,
		blendMode:          agg.BlendSrcOver,
		fillGradientType:   agg.SolidGradient,
		strokeGradientType: agg.SolidGradient,
		textAlignX:         agg.AlignLeft,
		textAlignY:         agg.AlignBottom,
		textResolution:     96,
		clipBox: agg.RectD{
			X1: 0,
			Y1: 0,
			X2: float64(img.Width()),
			Y2: float64(img.Height()),
		},
	}, nil
}

func (i *cppImage) Kind() Kind { return CPP }

func (i *cppImage) Width() int {
	if i == nil || i.img == nil {
		return 0
	}
	return i.img.width()
}

func (i *cppImage) Height() int {
	if i == nil || i.img == nil {
		return 0
	}
	return i.img.height()
}

func (i *cppImage) Premultiply() error {
	return &UnsupportedCapabilityError{Kind: CPP, Capability: CapabilityImageInterop, Operation: "Premultiply"}
}

func (i *cppImage) Demultiply() error {
	return &UnsupportedCapabilityError{Kind: CPP, Capability: CapabilityImageInterop, Operation: "Demultiply"}
}

func (i *cppImage) ToGoImage() *image.RGBA {
	if i == nil || i.img == nil {
		return nil
	}
	goImg, err := i.img.toGoImage()
	if err != nil {
		panic(err)
	}
	return goImg
}

func (i *cppImage) ToStandardImage() (image.Image, error) {
	if i == nil || i.img == nil {
		return nil, fmt.Errorf("image is nil")
	}
	return i.img.toGoImage()
}

func (i *cppImage) SaveToPNG(filename string) error {
	stdImg, err := i.ToStandardImage()
	if err != nil {
		return err
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, stdImg)
}

func (i *cppImage) SaveToJPEG(filename string, quality int) error {
	stdImg, err := i.ToStandardImage()
	if err != nil {
		return err
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, stdImg, &jpeg.Options{Quality: quality})
}

func (c *cppContext) Kind() Kind { return CPP }

func (c *cppContext) Width() int { return c.img.Width() }

func (c *cppContext) Height() int { return c.img.Height() }

func (c *cppContext) Clear(color agg.Color) {
	r, g, b, a := colorToRGBA8(color)
	c.must(c.img.img.clear(r, g, b, a))
}

func (c *cppContext) SetColor(color agg.Color) {
	c.SetFillColor(color)
	c.SetStrokeColor(color)
}

func (c *cppContext) SetFillColor(color agg.Color) {
	c.fillColor = color
	c.fillGradientType = agg.SolidGradient
	c.fillGradient = cppGradientState{kind: cppGradientSolid, c1: color, c2: color, profile: 1}
}

func (c *cppContext) SetStrokeColor(color agg.Color) {
	c.strokeColor = color
	c.strokeGradientType = agg.SolidGradient
	c.strokeGradient = cppGradientState{kind: cppGradientSolid, c1: color, c2: color, profile: 1}
}

func (c *cppContext) GetFillColor() agg.Color { return c.fillColor }

func (c *cppContext) GetStrokeColor() agg.Color { return c.strokeColor }

func (c *cppContext) SetLineWidth(width float64) { c.lineWidth = width }

func (c *cppContext) SetLineCap(cap agg.LineCap) { c.lineCap = cap }

func (c *cppContext) SetLineJoin(join agg.LineJoin) { c.lineJoin = join }

func (c *cppContext) GetLineWidth() float64 { return c.lineWidth }

func (c *cppContext) GetLineCap() agg.LineCap { return c.lineCap }

func (c *cppContext) GetLineJoin() agg.LineJoin { return c.lineJoin }

func (c *cppContext) AddDash(dashLen, gapLen float64) {
	c.dashes = append(c.dashes, float32(dashLen), float32(gapLen))
}

func (c *cppContext) RemoveAllDashes() {
	c.dashes = c.dashes[:0]
	c.dashStart = 0
}

func (c *cppContext) DashStart(offset float64) { c.dashStart = offset }

func (c *cppContext) GetDashStart() float64 { return c.dashStart }

func (c *cppContext) SetBlendMode(mode agg.BlendMode) { c.blendMode = mode }

func (c *cppContext) GetBlendMode() agg.BlendMode { return c.blendMode }

func (c *cppContext) FillEvenOdd(evenOdd bool) { c.fillEvenOdd = evenOdd }

func (c *cppContext) GetFillEvenOdd() bool { return c.fillEvenOdd }

func (c *cppContext) SetLinearGradientWithProfile(x1, y1, x2, y2 float64, c1, c2 agg.Color, profile float64) {
	c.fillGradientType = agg.LinearGradient
	c.fillGradient = cppGradientState{
		kind:    cppGradientLinear,
		c1:      c1,
		c2:      c2,
		profile: profile,
		x1:      x1,
		y1:      y1,
		x2:      x2,
		y2:      y2,
	}
}

func (c *cppContext) SetLinearGradient(x1, y1, x2, y2 float64, c1, c2 agg.Color) {
	c.SetLinearGradientWithProfile(x1, y1, x2, y2, c1, c2, 1)
}

func (c *cppContext) SetRadialGradient(cx, cy, radius float64, c1, c2 agg.Color) {
	c.SetRadialGradientWithProfile(cx, cy, radius, c1, c2, 1)
}

func (c *cppContext) SetRadialGradientWithProfile(cx, cy, radius float64, c1, c2 agg.Color, profile float64) {
	c.fillGradientType = agg.RadialGradient
	c.fillGradient = cppGradientState{
		kind:    cppGradientRadial,
		c1:      c1,
		c2:      c2,
		profile: profile,
		cx:      cx,
		cy:      cy,
		radius:  radius,
	}
}

func (c *cppContext) SetStrokeLinearGradient(x1, y1, x2, y2 float64, c1, c2 agg.Color) {
	c.SetStrokeLinearGradientWithProfile(x1, y1, x2, y2, c1, c2, 1)
}

func (c *cppContext) SetStrokeLinearGradientWithProfile(x1, y1, x2, y2 float64, c1, c2 agg.Color, profile float64) {
	c.strokeGradientType = agg.LinearGradient
	c.strokeGradient = cppGradientState{
		kind:    cppGradientLinear,
		c1:      c1,
		c2:      c2,
		profile: profile,
		x1:      x1,
		y1:      y1,
		x2:      x2,
		y2:      y2,
	}
}

func (c *cppContext) SetStrokeRadialGradient(cx, cy, radius float64, c1, c2 agg.Color) {
	c.SetStrokeRadialGradientWithProfile(cx, cy, radius, c1, c2, 1)
}

func (c *cppContext) SetStrokeRadialGradientWithProfile(cx, cy, radius float64, c1, c2 agg.Color, profile float64) {
	c.strokeGradientType = agg.RadialGradient
	c.strokeGradient = cppGradientState{
		kind:    cppGradientRadial,
		c1:      c1,
		c2:      c2,
		profile: profile,
		cx:      cx,
		cy:      cy,
		radius:  radius,
	}
}

func (c *cppContext) GetFillGradientType() agg.GradientType { return c.fillGradientType }

func (c *cppContext) GetStrokeGradientType() agg.GradientType { return c.strokeGradientType }

func (c *cppContext) BeginPath() {
	c.must(c.path.reset())
	c.hasCurrentPoint = false
}

func (c *cppContext) MoveTo(x, y float64) {
	c.must(c.path.moveTo(float32(x), float32(y)))
	c.currentX, c.currentY = x, y
	c.subpathStartX, c.subpathStartY = x, y
	c.hasCurrentPoint = true
}

func (c *cppContext) LineTo(x, y float64) {
	c.must(c.path.lineTo(float32(x), float32(y)))
	c.currentX, c.currentY = x, y
	c.hasCurrentPoint = true
}

func (c *cppContext) QuadTo(xCtrl, yCtrl, xTo, yTo float64) {
	startX, startY := c.currentPoint()
	const steps = 16
	for step := 1; step <= steps; step++ {
		t := float64(step) / float64(steps)
		mt := 1 - t
		x := mt*mt*startX + 2*mt*t*xCtrl + t*t*xTo
		y := mt*mt*startY + 2*mt*t*yCtrl + t*t*yTo
		c.LineTo(x, y)
	}
}

func (c *cppContext) CubicTo(xCtrl1, yCtrl1, xCtrl2, yCtrl2, xTo, yTo float64) {
	startX, startY := c.currentPoint()
	const steps = 24
	for step := 1; step <= steps; step++ {
		t := float64(step) / float64(steps)
		mt := 1 - t
		x := mt*mt*mt*startX + 3*mt*mt*t*xCtrl1 + 3*mt*t*t*xCtrl2 + t*t*t*xTo
		y := mt*mt*mt*startY + 3*mt*mt*t*yCtrl1 + 3*mt*t*t*yCtrl2 + t*t*t*yTo
		c.LineTo(x, y)
	}
}

func (c *cppContext) ClosePath() {
	c.must(c.path.closePath())
	if c.hasCurrentPoint {
		c.currentX, c.currentY = c.subpathStartX, c.subpathStartY
	}
}

func (c *cppContext) Fill() {
	working := c.mustTransformedPath()
	defer working.close()
	rule := cppNativeFillRuleNonZero
	if c.fillEvenOdd {
		rule = cppNativeFillRuleEvenOdd
	}
	if c.fillGradient.kind == cppGradientSolid {
		// Solid fill: render directly onto the destination through the comp-op
		// pixfmt so the blend mode is applied per span with anti-aliased coverage
		// (faithful to AGG's Agg2D). The earlier layer-then-composite path applied
		// the operator across the whole clip rectangle, so src/clear wiped the
		// untouched background.
		c.must(c.requireBlendMode("Fill"))
		r, g, b, a := colorToRGBA8(c.fillColor)
		c.must(fillCPPNativePathComp(c.img.img, working, rule, c.clipRectangle(), c.blendMode, r, g, b, a))
		return
	}
	// Gradient fill: rasterise the shape into a transparent layer, recolour by the
	// gradient, then composite. Gradient + non-src-over compositing is not yet a
	// faithful path (see docs/BACKENDS.md); the corpus only exercises gradients
	// with src-over.
	layer, err := newCPPNativeImage(c.Width(), c.Height())
	c.must(err)
	defer layer.close()
	c.must(layer.clear(0, 0, 0, 0))
	c.must(fillCPPNativePath(layer, working, rule, 255, 255, 255, 255))
	c.must(c.applyGradientToLayer(layer, c.fillGradient))
	c.must(c.compositeLayer(layer, "Fill"))
}

func (c *cppContext) Stroke() {
	working := c.mustTransformedPath()
	defer working.close()
	opts := defaultCPPNativeStrokeOptions()
	opts.Width = float32(c.lineWidth)
	opts.LineCap = mapLineCap(c.lineCap)
	opts.LineJoin = mapLineJoin(c.lineJoin)
	opts.Dashes = c.dashes
	opts.DashStart = float32(c.dashStart)
	if c.strokeGradient.kind == cppGradientSolid {
		// Solid stroke: render directly through the comp-op pixfmt (see Fill).
		c.must(c.requireBlendMode("Stroke"))
		r, g, b, a := colorToRGBA8(c.strokeColor)
		c.must(strokeCPPNativePathComp(c.img.img, working, opts, c.clipRectangle(), c.blendMode, r, g, b, a))
		return
	}
	layer, err := newCPPNativeImage(c.Width(), c.Height())
	c.must(err)
	defer layer.close()
	c.must(layer.clear(0, 0, 0, 0))
	c.must(strokeCPPNativePath(layer, working, opts, 255, 255, 255, 255))
	c.must(c.applyGradientToLayer(layer, c.strokeGradient))
	c.must(c.compositeLayer(layer, "Stroke"))
}

func (c *cppContext) DrawLine(x1, y1, x2, y2 float64) {
	c.BeginPath()
	c.MoveTo(x1, y1)
	c.LineTo(x2, y2)
	c.Stroke()
}

func (c *cppContext) DrawRectangle(x, y, width, height float64) {
	c.BeginPath()
	c.MoveTo(x, y)
	c.LineTo(x+width, y)
	c.LineTo(x+width, y+height)
	c.LineTo(x, y+height)
	c.ClosePath()
	c.Stroke()
}

func (c *cppContext) FillRectangle(x, y, width, height float64) {
	c.BeginPath()
	c.MoveTo(x, y)
	c.LineTo(x+width, y)
	c.LineTo(x+width, y+height)
	c.LineTo(x, y+height)
	c.ClosePath()
	c.Fill()
}

func (c *cppContext) DrawCircle(cx, cy, radius float64) {
	c.circlePath(cx, cy, radius)
	c.Stroke()
}

func (c *cppContext) FillCircle(cx, cy, radius float64) {
	c.circlePath(cx, cy, radius)
	c.Fill()
}

func (c *cppContext) ClipBox(x1, y1, x2, y2 float64) {
	c.clipBox = agg.RectD{X1: x1, Y1: y1, X2: x2, Y2: y2}
}

func (c *cppContext) GetClipBox() agg.RectD { return c.clipBox }

func (c *cppContext) Translate(tx, ty float64) {
	c.must(c.transform.translate(float32(tx), float32(ty)))
	c.transformDirty = true
}

func (c *cppContext) Rotate(angle float64) {
	c.must(c.transform.rotate(float32(angle)))
	c.transformDirty = true
}

func (c *cppContext) Scale(sx, sy float64) {
	c.must(c.transform.scale(float32(sx), float32(sy)))
	c.transformDirty = true
}

func (c *cppContext) ResetTransform() {
	c.must(c.transform.identity())
	c.transformDirty = false
}

// GetTransform returns the current cumulative affine transform as a
// backend-neutral matrix in AGG order (sx, shy, shx, sy, tx, ty), read back
// from the native matrix so it mirrors the port's GetTransform contract.
func (c *cppContext) GetTransform() agg.Transformations {
	m, err := c.transform.store()
	c.must(err)
	return agg.Transformations{AffineMatrix: m}
}

func (c *cppContext) DrawImage(img Image, x, y float64) error {
	if img == nil {
		return fmt.Errorf("image is nil")
	}
	return c.DrawImageRegion(img, 0, 0, img.Width(), img.Height(), x, y, float64(img.Width()), float64(img.Height()))
}

func (c *cppContext) DrawImageScaled(img Image, x, y, width, height float64) error {
	if img == nil {
		return fmt.Errorf("image is nil")
	}
	return c.DrawImageRegion(img, 0, 0, img.Width(), img.Height(), x, y, width, height)
}

func (c *cppContext) DrawImageRegion(img Image, srcX, srcY, srcW, srcH int, dstX, dstY, dstW, dstH float64) error {
	cppImg, err := unwrapCPPImage(img, c.Kind())
	if err != nil {
		return err
	}
	if c.transformDirty {
		return &UnsupportedCapabilityError{Kind: CPP, Capability: CapabilityImageDraw, Operation: "DrawImageRegion with active transform"}
	}
	if err := c.requireBlendMode("DrawImageRegion"); err != nil {
		return err
	}
	return c.img.img.compositeScaledFrom(
		cppImg.img,
		srcX,
		srcY,
		srcW,
		srcH,
		int(math.Round(dstX)),
		int(math.Round(dstY)),
		int(math.Round(dstW)),
		int(math.Round(dstH)),
		c.clipRectangle(),
		c.blendMode,
	)
}

func (c *cppContext) DrawImageQuad(img Image, quad [8]float64) error {
	if img == nil {
		return fmt.Errorf("image is nil")
	}
	return c.DrawImageRegionQuad(img, 0, 0, img.Width(), img.Height(), quad)
}

func (c *cppContext) DrawImageRegionQuad(img Image, srcX, srcY, srcW, srcH int, quad [8]float64) error {
	cppImg, err := unwrapCPPImage(img, c.Kind())
	if err != nil {
		return err
	}
	if err := c.requireBlendMode("DrawImageRegionQuad"); err != nil {
		return err
	}
	return c.img.img.compositeQuadFrom(
		cppImg.img,
		srcX,
		srcY,
		srcW,
		srcH,
		quad,
		c.clipRectangle(),
		c.blendMode,
	)
}

func (c *cppContext) LoadFont(fontFile string) error {
	if currentCPPNativeMetadata().Stub {
		return &UnsupportedCapabilityError{Kind: CPP, Capability: CapabilityText, Operation: "LoadFont"}
	}
	font, err := newCPPNativeFont(fontFile)
	if err != nil {
		return err
	}
	if err := font.setSize(12); err != nil {
		_ = font.close()
		return err
	}
	if err := font.setHinting(c.textHints); err != nil {
		_ = font.close()
		return err
	}
	if err := font.setFlipY(true); err != nil {
		_ = font.close()
		return err
	}
	if c.font != nil {
		_ = c.font.close()
	}
	c.font = font
	return nil
}

func (c *cppContext) SetResolution(dpi uint) { c.textResolution = dpi }

func (c *cppContext) TextHints(hints bool) {
	c.textHints = hints
	if c.font != nil {
		c.must(c.font.setHinting(hints))
	}
}

func (c *cppContext) GetTextHints() bool { return c.textHints }

func (c *cppContext) SetTextAlignment(alignX, alignY agg.TextAlignment) {
	c.textAlignX = alignX
	c.textAlignY = alignY
}

func (c *cppContext) DrawText(text string, x, y float64) error {
	if text == "" {
		return fmt.Errorf("text is empty")
	}
	if currentCPPNativeMetadata().Stub {
		return &UnsupportedCapabilityError{Kind: CPP, Capability: CapabilityText, Operation: "DrawText"}
	}
	if c.font == nil {
		return fmt.Errorf("font is not loaded")
	}
	layer, err := newCPPNativeImage(c.Width(), c.Height())
	if err != nil {
		return err
	}
	defer layer.close()
	if err := layer.clear(0, 0, 0, 0); err != nil {
		return err
	}
	drawX, drawY := c.textOrigin(text, x, y)
	r, g, b, a := colorToRGBA8(c.fillColor)
	if err := c.font.renderText(layer, text, float32(drawX), float32(drawY), r, g, b, a); err != nil {
		return err
	}
	return c.compositeLayer(layer, "DrawText")
}

func (c *cppContext) DrawTextAligned(text string, x, y float64, alignment agg.TextAlignment) error {
	if text == "" {
		return fmt.Errorf("text is empty")
	}
	width, _ := c.MeasureText(text)
	ax := x
	switch alignment {
	case agg.AlignCenter:
		ax = x - width/2
	case agg.AlignRight:
		ax = x - width
	}
	return c.DrawText(text, ax, y)
}

func (c *cppContext) MeasureText(text string) (width, height float64) {
	if c.font == nil || currentCPPNativeMetadata().Stub {
		return 0, 0
	}
	return c.font.textWidth(text), c.font.textHeight()
}

func (c *cppContext) GetTextBounds(text string) (x, y, width, height float64) {
	w, h := c.MeasureText(text)
	return 0, 0, w, h
}

func (c *cppContext) GetImage() Image { return c.img }

func (c *cppContext) compositeLayer(layer *cppNativeImage, operation string) error {
	if err := c.requireBlendMode(operation); err != nil {
		return err
	}
	return c.img.img.compositeFrom(layer, 0, 0, c.clipRectangle(), c.blendMode)
}

func (c *cppContext) applyGradientToLayer(layer *cppNativeImage, gradient cppGradientState) error {
	pixels, err := layer.pixelView()
	if err != nil {
		return err
	}
	stride := layer.stride()
	width := layer.width()
	height := layer.height()
	if stride == 0 || width == 0 || height == 0 {
		return nil
	}
	for y := 0; y < height; y++ {
		row := y * stride
		for x := 0; x < width; x++ {
			offset := row + x*4
			maskAlpha := pixels[offset+3]
			if maskAlpha == 0 {
				continue
			}
			t := gradient.sampleAt(float64(x)+0.5, float64(y)+0.5)
			color := gradient.c1.Gradient(gradient.c2, gradient.profileSample(t))
			pixels[offset+0] = color.R
			pixels[offset+1] = color.G
			pixels[offset+2] = color.B
			pixels[offset+3] = uint8((uint16(color.A) * uint16(maskAlpha)) / 255)
		}
	}
	return nil
}

func (c *cppContext) must(err error) {
	if err != nil {
		panic(err)
	}
}

func (c *cppContext) mustTransformedPath() *cppNativePath {
	path, err := c.path.transform(c.transform)
	if err != nil {
		panic(err)
	}
	return path
}

func (c *cppContext) currentPoint() (float64, float64) {
	if !c.hasCurrentPoint {
		panic("cppContext current point is undefined")
	}
	return c.currentX, c.currentY
}

func (c *cppContext) circlePath(cx, cy, radius float64) {
	c.BeginPath()
	const segments = 48
	for i := 0; i <= segments; i++ {
		angle := (2 * math.Pi * float64(i)) / float64(segments)
		x := cx + math.Cos(angle)*radius
		y := cy + math.Sin(angle)*radius
		if i == 0 {
			c.MoveTo(x, y)
		} else {
			c.LineTo(x, y)
		}
	}
	c.ClosePath()
}

func (c *cppContext) textOrigin(text string, x, y float64) (float64, float64) {
	width, height := c.MeasureText(text)
	outX := x
	switch c.textAlignX {
	case agg.AlignCenter:
		outX = x - width/2
	case agg.AlignRight:
		outX = x - width
	}
	outY := y
	switch c.textAlignY {
	case agg.AlignTop:
		outY = y + height
	case agg.AlignCenter:
		outY = y + height/2
	}
	return outX, outY
}

func (c *cppContext) clipRectangle() image.Rectangle {
	x1, y1 := c.clipBox.X1, c.clipBox.Y1
	x2, y2 := c.clipBox.X2, c.clipBox.Y2
	if x2 < x1 {
		x1, x2 = x2, x1
	}
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	return image.Rect(
		int(math.Floor(x1)),
		int(math.Floor(y1)),
		int(math.Ceil(x2)),
		int(math.Ceil(y2)),
	)
}

func (c *cppContext) requireBlendMode(operation string) error {
	switch c.blendMode {
	case agg.BlendAlpha, agg.BlendClear, agg.BlendSrc, agg.BlendDst, agg.BlendSrcOver:
		return nil
	default:
		return &UnsupportedCapabilityError{
			Kind:       CPP,
			Capability: CapabilityCompositing,
			Operation:  operation,
		}
	}
}

func (g cppGradientState) profileSample(t float64) float64 {
	if t <= 0 {
		return 0
	}
	if t >= 1 {
		return 1
	}
	profile := g.profile
	if profile <= 0 {
		profile = 0
	}
	start := 0.5 - profile*(127.0/255.0)
	end := 0.5 + profile*(127.0/255.0)
	if end <= start {
		end = start + (1.0 / 255.0)
	}
	switch {
	case t <= start:
		return 0
	case t >= end:
		return 1
	default:
		return (t - start) / (end - start)
	}
}

func (g cppGradientState) sampleAt(x, y float64) float64 {
	switch g.kind {
	case cppGradientLinear:
		dx := g.x2 - g.x1
		dy := g.y2 - g.y1
		lengthSq := dx*dx + dy*dy
		if lengthSq <= 1e-12 {
			return 0
		}
		return ((x-g.x1)*dx + (y-g.y1)*dy) / lengthSq
	case cppGradientRadial:
		if g.radius <= 1e-12 {
			return 1
		}
		return math.Hypot(x-g.cx, y-g.cy) / g.radius
	default:
		return 0
	}
}

func unwrapCPPImage(img Image, contextKind Kind) (*cppImage, error) {
	if img == nil {
		return nil, fmt.Errorf("image is nil")
	}
	cppImg, ok := img.(*cppImage)
	if !ok || cppImg == nil || cppImg.img == nil {
		return nil, &EngineMismatchError{
			ContextKind:  contextKind,
			ResourceKind: img.Kind(),
			ResourceType: "image",
		}
	}
	return cppImg, nil
}

func colorToRGBA8(color agg.Color) (r, g, b, a uint8) {
	return color.R, color.G, color.B, color.A
}

func mapLineCap(cap agg.LineCap) cppNativeLineCap {
	switch cap {
	case agg.CapRound:
		return cppNativeLineCapRound
	case agg.CapSquare:
		return cppNativeLineCapSquare
	default:
		return cppNativeLineCapButt
	}
}

func mapLineJoin(join agg.LineJoin) cppNativeLineJoin {
	switch join {
	case agg.JoinRound:
		return cppNativeLineJoinRound
	case agg.JoinBevel:
		return cppNativeLineJoinBevel
	default:
		return cppNativeLineJoinMiter
	}
}
