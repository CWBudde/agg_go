package main

import (
	agg "agg_go"
)

func drawAADemo() {
	ctx.SetColor(agg.Black)
	ctx.SetLineWidth(1.0)

	// Draw lines with sub-pixel increments
	for i := 0; i < 20; i++ {
		offset := float64(i) / 10.0
		ctx.DrawLine(50+offset, 50+float64(i)*20, 250+offset, 70+float64(i)*20)
	}

	// Draw a circle with sub-pixel movement
	for i := 0; i < 10; i++ {
		offset := float64(i) / 5.0
		ctx.SetColor(agg.RGBA(0.1, 0.5, 0.8, 0.3))
		ctx.FillCircle(400+offset*10, 300+offset*10, 50)
		ctx.SetColor(agg.Black)
		ctx.DrawCircle(400+offset*10, 300+offset*10, 50)
	}
}
