package main

import (
	"image"
	"image/png"
	"os"

	agg "github.com/MeKo-Christian/agg_go"
	alphamask2 "github.com/MeKo-Christian/agg_go/internal/demo/alphamask2"
)

func main() {
	const (
		width  = 512
		height = 400
	)

	img := agg.NewImage(make([]uint8, width*height*4), width, height, width*4)
	work := make([]uint8, width*height*3)
	alphamask2.RenderToBGR24(work, width, height, alphamask2.Config{
		NumEllipses: 10,
		Scale:       1,
	})

	for y := 0; y < height; y++ {
		srcOff := (height - 1 - y) * width * 3
		dstOff := y * width * 4
		for x := 0; x < width; x++ {
			s := srcOff + x*3
			d := dstOff + x*4
			img.Data[d] = work[s+2]
			img.Data[d+1] = work[s+1]
			img.Data[d+2] = work[s]
			img.Data[d+3] = 255
		}
	}

	goImg := image.NewRGBA(image.Rect(0, 0, width, height))
	copy(goImg.Pix, img.Data)

	f, err := os.Create("/tmp/alpha_mask2_go_current.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := png.Encode(f, goImg); err != nil {
		panic(err)
	}
}
