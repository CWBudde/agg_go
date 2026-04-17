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

func newPortImageFromGoImage(src image.Image) (Image, error) {
	img, err := agg.NewImageFromStandardImage(src)
	if err != nil {
		return nil, err
	}
	return &portImage{img: img}, nil
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

func (c *portContext) SetLineCap(cap agg.LineCap) { c.ctx.SetLineCap(cap) }

func (c *portContext) SetLineJoin(join agg.LineJoin) { c.ctx.SetLineJoin(join) }

func (c *portContext) SetBlendMode(mode agg.BlendMode) { c.ctx.SetBlendMode(mode) }

func (c *portContext) FillEvenOdd(evenOdd bool) { c.ctx.GetAgg2D().FillEvenOdd(evenOdd) }

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

func (c *portContext) DrawImageQuad(img Image, quad [8]float64) error {
	portImg, err := unwrapPortImage(img, c.Kind())
	if err != nil {
		return err
	}
	return c.ctx.DrawImageQuad(portImg, quad)
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

func (i *portImage) SaveToPNG(filename string) error { return i.img.SaveToPNG(filename) }

func unwrapPortImage(img Image, contextKind Kind) (*agg.Image, error) {
	if img == nil {
		return nil, fmt.Errorf("image is nil")
	}
	portImg, ok := img.(*portImage)
	if !ok || portImg == nil || portImg.img == nil {
		return nil, fmt.Errorf("image engine mismatch: context=%s image=%s", contextKind, img.Kind())
	}
	return portImg.img, nil
}
