package lowlevelrunner

import (
	"testing"

	agg "github.com/cwbudde/agg_go"
)

func TestOutputImage_EncodeLinearRGBToSRGB(t *testing.T) {
	img := agg.NewImage(make([]byte, 4), 1, 1, 4)
	img.Data[0] = 255
	img.Data[1] = 242
	img.Data[2] = 242
	img.Data[3] = 255

	goImg := outputImage(img, true)
	got := goImg.RGBAAt(0, 0)
	if got.R != 255 || got.G != 249 || got.B != 249 || got.A != 255 {
		t.Fatalf("encoded pixel = %v, want rgba(255,249,249,255)", got)
	}
}
