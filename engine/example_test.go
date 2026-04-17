package engine_test

import (
	"fmt"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/engine"
)

func ExampleNewContext() {
	ctx, err := engine.NewContext(16, 16, engine.Config{})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	ctx.Clear(agg.White)
	ctx.SetColor(agg.Red)
	ctx.FillRectangle(4, 4, 8, 8)

	img := ctx.GetImage().ToGoImage()
	center := img.RGBAAt(8, 8)
	fmt.Println(ctx.Kind(), center.R > 200, center.G < 20, center.B < 20, center.A == 255)
	// Output:
	// port true true true true
}
