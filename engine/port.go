package engine

import (
	"fmt"
	"image"

	agg "github.com/cwbudde/agg_go"
)

type portContext struct {
	ctx *agg.Context
}

type portImage struct {
	img *agg.Image
}

func newPortContext(width, height int) Context {
	return &portContext{ctx: agg.NewContext(width, height)}
}

func newPortContextForImage(img Image) (Context, error) {
	portImg, err := unwrapPortImage(img, Port)
	if err != nil {
		return nil, err
	}
	return &portContext{ctx: agg.NewContextForImage(portImg)}, nil
}

func newPortImage(width, height int) (Image, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("image dimensions must be positive, got %dx%d", width, height)
	}
	stride := width * 4
	return &portImage{
		img: agg.NewImage(make([]byte, height*stride), width, height, stride),
	}, nil
}

func newPortImageFromGoImage(src image.Image) (Image, error) {
	img, err := agg.NewImageFromStandardImage(src)
	if err != nil {
		return nil, err
	}
	return &portImage{img: img}, nil
}

func newPortImageFromBuffer(buf []byte, width, height, stride int) (Image, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("image dimensions must be positive, got %dx%d", width, height)
	}
	if stride == 0 {
		return nil, fmt.Errorf("stride must be non-zero")
	}
	minLen := abs(stride) * height
	if len(buf) < minLen {
		return nil, fmt.Errorf("buffer too small: len=%d need_at_least=%d", len(buf), minLen)
	}
	return &portImage{img: agg.NewImage(buf, width, height, stride)}, nil
}

func loadPortImageFromFile(filename string) (Image, error) {
	img, err := agg.LoadImageFromFile(filename)
	if err != nil {
		return nil, err
	}
	return &portImage{img: img}, nil
}

func (c *portContext) Kind() Kind { return Port }

func (c *portContext) Width() int { return c.ctx.Width() }

func (c *portContext) Height() int { return c.ctx.Height() }

func (c *portContext) Clear(color agg.Color) { c.ctx.Clear(color) }

func (c *portContext) SetColor(color agg.Color) { c.ctx.SetColor(color) }

func (c *portContext) SetFillColor(color agg.Color) { c.ctx.SetFillColor(color) }

func (c *portContext) SetStrokeColor(color agg.Color) { c.ctx.SetStrokeColor(color) }

func (c *portContext) SetLineWidth(width float64) { c.ctx.SetLineWidth(width) }

func (c *portContext) SetLineCap(lineCap agg.LineCap) { c.ctx.SetLineCap(lineCap) }

func (c *portContext) SetLineJoin(join agg.LineJoin) { c.ctx.SetLineJoin(join) }

func (c *portContext) SetBlendMode(mode agg.BlendMode) { c.ctx.SetBlendMode(mode) }

func (c *portContext) GetBlendMode() agg.BlendMode { return c.ctx.GetBlendMode() }

func (c *portContext) FillEvenOdd(evenOdd bool) { c.ctx.GetAgg2D().FillEvenOdd(evenOdd) }

func (c *portContext) GetFillEvenOdd() bool { return c.ctx.GetAgg2D().GetFillEvenOdd() }

func (c *portContext) SetLinearGradient(x1, y1, x2, y2 float64, c1, c2 agg.Color) {
	c.ctx.SetLinearGradient(x1, y1, x2, y2, c1, c2)
}

func (c *portContext) SetLinearGradientWithProfile(x1, y1, x2, y2 float64, c1, c2 agg.Color, profile float64) {
	c.ctx.SetLinearGradientWithProfile(x1, y1, x2, y2, c1, c2, profile)
}

func (c *portContext) SetRadialGradient(cx, cy, radius float64, c1, c2 agg.Color) {
	c.ctx.SetRadialGradient(cx, cy, radius, c1, c2)
}

func (c *portContext) SetRadialGradientWithProfile(cx, cy, radius float64, c1, c2 agg.Color, profile float64) {
	c.ctx.SetRadialGradientWithProfile(cx, cy, radius, c1, c2, profile)
}

func (c *portContext) SetStrokeLinearGradient(x1, y1, x2, y2 float64, c1, c2 agg.Color) {
	c.ctx.SetStrokeLinearGradient(x1, y1, x2, y2, c1, c2)
}

func (c *portContext) SetStrokeLinearGradientWithProfile(x1, y1, x2, y2 float64, c1, c2 agg.Color, profile float64) {
	c.ctx.SetStrokeLinearGradientWithProfile(x1, y1, x2, y2, c1, c2, profile)
}

func (c *portContext) SetStrokeRadialGradient(cx, cy, radius float64, c1, c2 agg.Color) {
	c.ctx.SetStrokeRadialGradient(cx, cy, radius, c1, c2)
}

func (c *portContext) SetStrokeRadialGradientWithProfile(cx, cy, radius float64, c1, c2 agg.Color, profile float64) {
	c.ctx.SetStrokeRadialGradientWithProfile(cx, cy, radius, c1, c2, profile)
}

func (c *portContext) GetFillGradientType() agg.GradientType { return c.ctx.GetFillGradientType() }

func (c *portContext) GetStrokeGradientType() agg.GradientType { return c.ctx.GetStrokeGradientType() }

func (c *portContext) BeginPath() { c.ctx.BeginPath() }

func (c *portContext) MoveTo(x, y float64) { c.ctx.MoveTo(x, y) }

func (c *portContext) LineTo(x, y float64) { c.ctx.LineTo(x, y) }

func (c *portContext) QuadTo(xCtrl, yCtrl, xTo, yTo float64) {
	c.ctx.QuadricCurveTo(xCtrl, yCtrl, xTo, yTo)
}

func (c *portContext) CubicTo(xCtrl1, yCtrl1, xCtrl2, yCtrl2, xTo, yTo float64) {
	c.ctx.CubicCurveTo(xCtrl1, yCtrl1, xCtrl2, yCtrl2, xTo, yTo)
}

func (c *portContext) ClosePath() { c.ctx.ClosePath() }

func (c *portContext) Fill() { c.ctx.Fill() }

func (c *portContext) Stroke() { c.ctx.Stroke() }

func (c *portContext) DrawLine(x1, y1, x2, y2 float64) { c.ctx.DrawLine(x1, y1, x2, y2) }

func (c *portContext) DrawRectangle(x, y, width, height float64) {
	c.ctx.DrawRectangle(x, y, width, height)
}

func (c *portContext) FillRectangle(x, y, width, height float64) {
	c.ctx.FillRectangle(x, y, width, height)
}

func (c *portContext) DrawCircle(cx, cy, radius float64) { c.ctx.DrawCircle(cx, cy, radius) }

func (c *portContext) FillCircle(cx, cy, radius float64) { c.ctx.FillCircle(cx, cy, radius) }

func (c *portContext) ClipBox(x1, y1, x2, y2 float64) { c.ctx.GetAgg2D().ClipBox(x1, y1, x2, y2) }

func (c *portContext) GetClipBox() agg.RectD { return c.ctx.GetAgg2D().GetClipBoxRect() }

func (c *portContext) Translate(tx, ty float64) { c.ctx.Translate(tx, ty) }

func (c *portContext) Rotate(angle float64) { c.ctx.Rotate(angle) }

func (c *portContext) Scale(sx, sy float64) { c.ctx.Scale(sx, sy) }

func (c *portContext) ResetTransform() { c.ctx.ResetTransform() }

func (c *portContext) DrawImage(img Image, x, y float64) error {
	portImg, err := unwrapPortImage(img, c.Kind())
	if err != nil {
		return err
	}
	return c.ctx.DrawImage(portImg, x, y)
}

func (c *portContext) DrawImageScaled(img Image, x, y, width, height float64) error {
	portImg, err := unwrapPortImage(img, c.Kind())
	if err != nil {
		return err
	}
	return c.ctx.DrawImageScaled(portImg, x, y, width, height)
}

func (c *portContext) DrawImageRegion(img Image, srcX, srcY, srcW, srcH int, dstX, dstY, dstW, dstH float64) error {
	portImg, err := unwrapPortImage(img, c.Kind())
	if err != nil {
		return err
	}
	return c.ctx.DrawImageRegion(portImg, srcX, srcY, srcW, srcH, dstX, dstY, dstW, dstH)
}

func (c *portContext) DrawImageQuad(img Image, quad [8]float64) error {
	portImg, err := unwrapPortImage(img, c.Kind())
	if err != nil {
		return err
	}
	return c.ctx.DrawImageQuad(portImg, quad)
}

func (c *portContext) DrawImageRegionQuad(img Image, srcX, srcY, srcW, srcH int, quad [8]float64) error {
	portImg, err := unwrapPortImage(img, c.Kind())
	if err != nil {
		return err
	}
	return c.ctx.DrawImageRegionQuad(portImg, srcX, srcY, srcW, srcH, quad)
}

func (c *portContext) LoadFont(fontFile string) error { return c.ctx.LoadFont(fontFile) }

func (c *portContext) SetResolution(dpi uint) { c.ctx.SetResolution(dpi) }

func (c *portContext) TextHints(hints bool) { c.ctx.TextHints(hints) }

func (c *portContext) GetTextHints() bool { return c.ctx.GetTextHints() }

func (c *portContext) SetTextAlignment(alignX, alignY agg.TextAlignment) {
	c.ctx.SetTextAlignment(alignX, alignY)
}

func (c *portContext) DrawText(text string, x, y float64) error { return c.ctx.DrawText(text, x, y) }

func (c *portContext) DrawTextAligned(text string, x, y float64, alignment agg.TextAlignment) error {
	return c.ctx.DrawTextAligned(text, x, y, alignment)
}

func (c *portContext) MeasureText(text string) (width, height float64) {
	return c.ctx.MeasureText(text)
}

func (c *portContext) GetTextBounds(text string) (x, y, width, height float64) {
	return c.ctx.GetTextBounds(text)
}

func (c *portContext) GetImage() Image {
	return &portImage{img: c.ctx.GetImage()}
}

func (i *portImage) Kind() Kind { return Port }

func (i *portImage) Width() int { return i.img.Width() }

func (i *portImage) Height() int { return i.img.Height() }

func (i *portImage) Premultiply() error { return i.img.Premultiply() }

func (i *portImage) Demultiply() error { return i.img.Demultiply() }

func (i *portImage) ToGoImage() *image.RGBA { return i.img.ToGoImage() }

func (i *portImage) ToStandardImage() (image.Image, error) { return i.img.ToStandardImage() }

func (i *portImage) SaveToPNG(filename string) error { return i.img.SaveToPNG(filename) }

func (i *portImage) SaveToJPEG(filename string, quality int) error {
	return i.img.SaveToJPEG(filename, quality)
}

func unwrapPortImage(img Image, contextKind Kind) (*agg.Image, error) {
	if img == nil {
		return nil, fmt.Errorf("image is nil")
	}
	portImg, ok := img.(*portImage)
	if !ok || portImg == nil || portImg.img == nil {
		return nil, &EngineMismatchError{
			ContextKind:  contextKind,
			ResourceKind: img.Kind(),
			ResourceType: "image",
		}
	}
	return portImg.img, nil
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
