package agg

// ContextFloat is the float-precision counterpart of Context: a high-level
// drawing facade over a float (128-bit) backing image and an Agg2DFloat renderer.
type ContextFloat struct {
	agg2d     *Agg2DFloat
	image     *ImageFloat
	width     int
	height    int
	lineWidth float64
}

// NewContextFloat allocates a new float image buffer and attaches a fresh
// Agg2DFloat renderer to it.
func NewContextFloat(width, height int) *ContextFloat {
	img := NewImageFloat(width, height)
	a := NewAgg2DFloat()
	a.AttachImage(img)

	ctx := &ContextFloat{
		agg2d:     a,
		image:     img,
		width:     width,
		height:    height,
		lineWidth: 1.0,
	}
	ctx.SetColor(Black)
	ctx.agg2d.LineWidth(ctx.lineWidth)
	return ctx
}

// NewContextFloatForImage creates a ContextFloat that renders into an existing
// float image.
func NewContextFloatForImage(img *ImageFloat) *ContextFloat {
	if img == nil {
		return nil
	}
	a := NewAgg2DFloat()
	a.AttachImage(img)
	ctx := &ContextFloat{
		agg2d:     a,
		image:     img,
		width:     img.Width(),
		height:    img.Height(),
		lineWidth: 1.0,
	}
	ctx.SetColor(Black)
	ctx.agg2d.LineWidth(ctx.lineWidth)
	return ctx
}

// Width returns the context width in pixels.
func (ctx *ContextFloat) Width() int { return ctx.width }

// Height returns the context height in pixels.
func (ctx *ContextFloat) Height() int { return ctx.height }

// GetImage returns the backing float image (shares memory with the context).
func (ctx *ContextFloat) GetImage() *ImageFloat { return ctx.image }

// GetAgg2D exposes the underlying float renderer for advanced operations.
func (ctx *ContextFloat) GetAgg2D() *Agg2DFloat { return ctx.agg2d }

// Clear fills the entire image with color.
func (ctx *ContextFloat) Clear(color Color) { ctx.agg2d.ClearAll(color) }

// SetColor sets both fill and stroke colors.
func (ctx *ContextFloat) SetColor(color Color) {
	ctx.agg2d.FillColor(color)
	ctx.agg2d.LineColor(color)
}

// SetLineWidth sets the stroke width for subsequent stroked operations.
func (ctx *ContextFloat) SetLineWidth(width float64) {
	ctx.lineWidth = width
	ctx.agg2d.LineWidth(width)
}

// DrawLine strokes a line immediately.
func (ctx *ContextFloat) DrawLine(x1, y1, x2, y2 float64) { ctx.agg2d.Line(x1, y1, x2, y2) }

// DrawRectangle strokes a rectangle immediately.
func (ctx *ContextFloat) DrawRectangle(x, y, width, height float64) {
	ctx.agg2d.ResetPath()
	ctx.agg2d.MoveTo(x, y)
	ctx.agg2d.LineTo(x+width, y)
	ctx.agg2d.LineTo(x+width, y+height)
	ctx.agg2d.LineTo(x, y+height)
	ctx.agg2d.ClosePolygon()
	ctx.agg2d.DrawPath(StrokeOnly)
}

// FillRectangle fills a rectangle immediately.
func (ctx *ContextFloat) FillRectangle(x, y, width, height float64) {
	ctx.agg2d.ResetPath()
	ctx.agg2d.MoveTo(x, y)
	ctx.agg2d.LineTo(x+width, y)
	ctx.agg2d.LineTo(x+width, y+height)
	ctx.agg2d.LineTo(x, y+height)
	ctx.agg2d.ClosePolygon()
	ctx.agg2d.DrawPath(FillOnly)
}

// DrawCircle strokes a circle immediately.
func (ctx *ContextFloat) DrawCircle(cx, cy, radius float64) { ctx.agg2d.DrawCircle(cx, cy, radius) }

// FillCircle fills a circle immediately.
func (ctx *ContextFloat) FillCircle(cx, cy, radius float64) { ctx.agg2d.FillCircle(cx, cy, radius) }
