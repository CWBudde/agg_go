package lowlevelrunner

import (
	"image"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/internal/color"
)

func outputImage(img *agg.Image, encodeLinearRGBToSRGB bool) *image.RGBA {
	goImg := img.ToGoImage()
	if goImg == nil || !encodeLinearRGBToSRGB {
		return goImg
	}
	for i := 0; i+3 < len(goImg.Pix); i += 4 {
		c := color.ConvertToSRGBFromLinear(color.RGBA8[color.Linear]{
			R: goImg.Pix[i],
			G: goImg.Pix[i+1],
			B: goImg.Pix[i+2],
			A: goImg.Pix[i+3],
		})
		goImg.Pix[i] = c.R
		goImg.Pix[i+1] = c.G
		goImg.Pix[i+2] = c.B
		goImg.Pix[i+3] = c.A
	}
	return goImg
}
