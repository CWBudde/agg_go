package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
)

type demo struct{}

func (d *demo) Render(img *agg.Image) {
	ctx := agg.NewContextForImage(img)
	ctx.Clear(agg.White)

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()
	agg2d.ResetPath()

	scale := 1.2
	offsetX, offsetY := 250.0, 100.0
	ld := liondemo.Parse()
	for i := 0; i < ld.NPaths; i++ {
		agg2d.FillColor(agg.NewColor(ld.Colors[i].R, ld.Colors[i].G, ld.Colors[i].B, 255))
		agg2d.NoLine()
		agg2d.ResetPath()
		ld.Path.Rewind(ld.PathIdx[i])
		for {
			x, y, cmd := ld.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}

			tx, ty := x*scale+offsetX, y*scale+offsetY
			if basics.IsMoveTo(basics.PathCommand(cmd)) {
				agg2d.MoveTo(tx, ty)
			} else if basics.IsLineTo(basics.PathCommand(cmd)) {
				agg2d.LineTo(tx, ty)
			}
		}
		agg2d.ClosePolygon()
		agg2d.DrawPath(agg.FillOnly)
	}
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Lion",
		Width:  512,
		Height: 400,
	}, &demo{})
}
