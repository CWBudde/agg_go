package integration

import "agg_go/internal/agg2d"

func addRectPath(ctx *agg2d.Agg2D, x1, y1, x2, y2 float64) {
	ctx.MoveTo(x1, y1)
	ctx.LineTo(x2, y1)
	ctx.LineTo(x2, y2)
	ctx.LineTo(x1, y2)
	ctx.ClosePolygon()
}

func drawFilledRectPath(ctx *agg2d.Agg2D, x1, y1, x2, y2 float64) {
	ctx.ResetPath()
	addRectPath(ctx, x1, y1, x2, y2)
	ctx.DrawPath(agg2d.FillOnly)
}
