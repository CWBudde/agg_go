package visual

import (
	"path/filepath"
	"testing"

	agg "github.com/cwbudde/agg_go"
)

// Float Agg2D image-transform visual/demo hook (PLAN.md §4.7).
//
// Renders an affine-scaled image transform through the public FLOAT path
// (Agg2DFloat.TransformImageSimple -> ImageFloat.ToRGBA) and the SAME transform
// through the 8-bit path (Agg2D.TransformImageSimple -> Image.ToGoImage), then
// asserts agreement within a documented tolerance. The 8-bit render is the
// oracle, so no new reference PNGs are needed.
//
// The source and background are fully OPAQUE: premultiplied == straight for
// opaque pixels, so the float ToRGBA / 8-bit ToGoImage boundary is identity and
// the comparison is apples-to-apples.

func fillRampSource8(w, h int) []uint8 {
	data := make([]uint8, w*h*4)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			o := (y*w + x) * 4
			data[o+0] = uint8(20 + x*200/(w-1))
			data[o+1] = uint8(30 + y*200/(h-1))
			data[o+2] = 128
			data[o+3] = 255
		}
	}
	return data
}

func TestFloatImageTransformVisualParity(t *testing.T) {
	const srcW, srcH = 16, 16
	const w, h = 96, 96

	srcData := fillRampSource8(srcW, srcH)

	// 8-bit oracle: transform into a target buffer, then read it back as RGBA.
	src8 := agg.NewImage(srcData, srcW, srcH, srcW*4)
	dst8Buf := make([]uint8, w*h*4)
	a8 := agg.NewAgg2D()
	a8.Attach(dst8Buf, w, h, w*4)
	a8.ClearAll(agg.Color{R: 255, G: 255, B: 255, A: 255})
	if err := a8.TransformImageSimple(src8, 16, 16, 80, 80); err != nil {
		t.Fatalf("8bit TransformImageSimple: %v", err)
	}
	img8 := agg.NewImage(dst8Buf, w, h, w*4).ToGoImage()

	// Float path: build a float source with the same straight data.
	srcF := agg.NewImageFloat(srcW, srcH)
	for y := 0; y < srcH; y++ {
		for x := 0; x < srcW; x++ {
			o := (y*srcW + x) * 4
			srcF.SetPixelFloat(x, y,
				float32(srcData[o+0])/255, float32(srcData[o+1])/255,
				float32(srcData[o+2])/255, float32(srcData[o+3])/255)
		}
	}
	dstF := agg.NewImageFloat(w, h)
	aF := agg.NewAgg2DFloat()
	aF.AttachImage(dstF)
	aF.ClearAll(agg.Color{R: 255, G: 255, B: 255, A: 255})
	if err := aF.TransformImageSimple(srcF, 16, 16, 80, 80); err != nil {
		t.Fatalf("float TransformImageSimple: %v", err)
	}
	imgF := dstF.ToRGBA()

	out := filepath.Join("output", "float_image_transform.png")
	if err := savePNG(out, imgF.Pix, w, h); err != nil {
		t.Fatalf("failed to save float render: %v", err)
	}
	t.Logf("float image-transform render written to %s", out)

	// Sample interior points of the transformed region; skip edges where AA
	// differs. Tolerance mirrors the cross-precision envelope (bilinear ~4).
	const tol = 4
	for _, p := range [][2]int{{32, 32}, {48, 48}, {64, 64}, {40, 56}, {56, 40}} {
		if d := maxRGBADiff(imgF, img8, p[0], p[1]); d > tol {
			f := imgF.RGBAAt(p[0], p[1])
			e := img8.RGBAAt(p[0], p[1])
			t.Errorf("transform mismatch at (%d,%d): float=%v 8bit=%v maxdiff=%d (tol=%d)",
				p[0], p[1], f, e, d, tol)
		}
	}
}
