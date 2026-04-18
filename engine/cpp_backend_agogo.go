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

type cppContext struct {
	img                *cppImage
	path               *cppNativePath
	transform          *cppNativeMatrix
	transformDirty     bool
	currentX           float64
	currentY           float64
	subpathStartX      float64
	subpathStartY      float64
	hasCurrentPoint    bool
	fillColor          agg.Color
	strokeColor        agg.Color
	lineWidth          float64
	lineCap            agg.LineCap
	lineJoin           agg.LineJoin
	blendMode          agg.BlendMode
	fillEvenOdd        bool
	fillGradientType   agg.GradientType
	strokeGradientType agg.GradientType
	textHints          bool
	textAlignX         agg.TextAlignment
	textAlignY         agg.TextAlignment
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
		lineWidth:          1,
		lineCap:            agg.CapButt,
		lineJoin:           agg.JoinMiter,
		blendMode:          agg.BlendSrcOver,
		fillGradientType:   agg.SolidGradient,
		strokeGradientType: agg.SolidGradient,
		textAlignX:         agg.AlignLeft,
		textAlignY:         agg.AlignBottom,
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
}

func (c *cppContext) SetStrokeColor(color agg.Color) {
	c.strokeColor = color
	c.strokeGradientType = agg.SolidGradient
}

func (c *cppContext) SetLineWidth(width float64) { c.lineWidth = width }

func (c *cppContext) SetLineCap(cap agg.LineCap) { c.lineCap = cap }

func (c *cppContext) SetLineJoin(join agg.LineJoin) { c.lineJoin = join }

func (c *cppContext) SetBlendMode(mode agg.BlendMode) { c.blendMode = mode }

func (c *cppContext) GetBlendMode() agg.BlendMode { return c.blendMode }

func (c *cppContext) FillEvenOdd(evenOdd bool) { c.fillEvenOdd = evenOdd }

func (c *cppContext) GetFillEvenOdd() bool { return c.fillEvenOdd }

func (c *cppContext) SetLinearGradient(_, _, _, _ float64, _, _ agg.Color) {
	c.fillGradientType = agg.LinearGradient
}

func (c *cppContext) SetLinearGradientWithProfile(_, _, _, _ float64, _, _ agg.Color, _ float64) {
	c.fillGradientType = agg.LinearGradient
}

func (c *cppContext) SetRadialGradient(_, _, _ float64, _, _ agg.Color) {
	c.fillGradientType = agg.RadialGradient
}

func (c *cppContext) SetRadialGradientWithProfile(_, _, _ float64, _, _ agg.Color, _ float64) {
	c.fillGradientType = agg.RadialGradient
}

func (c *cppContext) SetStrokeLinearGradient(_, _, _, _ float64, _, _ agg.Color) {
	c.strokeGradientType = agg.LinearGradient
}

func (c *cppContext) SetStrokeLinearGradientWithProfile(_, _, _, _ float64, _, _ agg.Color, _ float64) {
	c.strokeGradientType = agg.LinearGradient
}

func (c *cppContext) SetStrokeRadialGradient(_, _, _ float64, _, _ agg.Color) {
	c.strokeGradientType = agg.RadialGradient
}

func (c *cppContext) SetStrokeRadialGradientWithProfile(_, _, _ float64, _, _ agg.Color, _ float64) {
	c.strokeGradientType = agg.RadialGradient
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
	layer, err := newCPPNativeImage(c.Width(), c.Height())
	c.must(err)
	defer layer.close()
	c.must(layer.clear(0, 0, 0, 0))
	r, g, b, a := colorToRGBA8(c.fillColor)
	rule := cppNativeFillRuleNonZero
	if c.fillEvenOdd {
		rule = cppNativeFillRuleEvenOdd
	}
	c.must(fillCPPNativePath(layer, working, rule, r, g, b, a))
	c.must(c.compositeLayer(layer, "Fill"))
}

func (c *cppContext) Stroke() {
	working := c.mustTransformedPath()
	defer working.close()
	layer, err := newCPPNativeImage(c.Width(), c.Height())
	c.must(err)
	defer layer.close()
	c.must(layer.clear(0, 0, 0, 0))
	r, g, b, a := colorToRGBA8(c.strokeColor)
	opts := defaultCPPNativeStrokeOptions()
	opts.Width = float32(c.lineWidth)
	opts.LineCap = mapLineCap(c.lineCap)
	opts.LineJoin = mapLineJoin(c.lineJoin)
	c.must(strokeCPPNativePath(layer, working, opts, r, g, b, a))
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

func (c *cppContext) LoadFont(string) error {
	return &UnsupportedCapabilityError{Kind: CPP, Capability: CapabilityText, Operation: "LoadFont"}
}

func (c *cppContext) SetResolution(uint) {}

func (c *cppContext) TextHints(hints bool) { c.textHints = hints }

func (c *cppContext) GetTextHints() bool { return c.textHints }

func (c *cppContext) SetTextAlignment(alignX, alignY agg.TextAlignment) {
	c.textAlignX = alignX
	c.textAlignY = alignY
}

func (c *cppContext) DrawText(string, float64, float64) error {
	return &UnsupportedCapabilityError{Kind: CPP, Capability: CapabilityText, Operation: "DrawText"}
}

func (c *cppContext) DrawTextAligned(string, float64, float64, agg.TextAlignment) error {
	return &UnsupportedCapabilityError{Kind: CPP, Capability: CapabilityText, Operation: "DrawTextAligned"}
}

func (c *cppContext) MeasureText(string) (width, height float64) { return 0, 0 }

func (c *cppContext) GetTextBounds(string) (x, y, width, height float64) { return 0, 0, 0, 0 }

func (c *cppContext) GetImage() Image { return c.img }

func (c *cppContext) compositeLayer(layer *cppNativeImage, operation string) error {
	if err := c.requireBlendMode(operation); err != nil {
		return err
	}
	return c.img.img.compositeFrom(layer, 0, 0, c.clipRectangle(), c.blendMode)
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
