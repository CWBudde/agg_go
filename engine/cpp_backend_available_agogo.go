//go:build agogo && cgo

package engine

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

func newCPPBackendContextAvailable(width, height int) (Context, error) {
	return newCPPBackendContext(width, height)
}

func newCPPBackendContextForImageAvailable(img Image) (Context, error) {
	cppImg, err := unwrapCPPImage(img, CPP)
	if err != nil {
		return nil, err
	}
	return newCPPBackendContextForImage(cppImg)
}

func newCPPBackendImageAvailable(width, height int) (Image, error) {
	return newCPPBackendImage(width, height)
}

func newCPPBackendImageFromGoImageAvailable(src image.Image) (Image, error) {
	if src == nil {
		return nil, fmt.Errorf("image is nil")
	}
	bounds := src.Bounds()
	img, err := newCPPBackendImage(bounds.Dx(), bounds.Dy())
	if err != nil {
		return nil, err
	}
	pix, err := img.img.pixelView()
	if err != nil {
		return nil, err
	}
	for y := 0; y < bounds.Dy(); y++ {
		row := y * img.img.stride()
		for x := 0; x < bounds.Dx(); x++ {
			r, g, b, a := src.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			offset := row + x*4
			pix[offset+0] = uint8(r >> 8)
			pix[offset+1] = uint8(g >> 8)
			pix[offset+2] = uint8(b >> 8)
			pix[offset+3] = uint8(a >> 8)
		}
	}
	return img, nil
}

func newCPPBackendImageFromBufferAvailable(buf []byte, width, height, stride int) (Image, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("image dimensions must be positive, got %dx%d", width, height)
	}
	if stride == 0 {
		return nil, fmt.Errorf("stride must be non-zero")
	}
	minLen := abs(stride) * height
	if len(buf) < minLen {
		return nil, fmt.Errorf("buffer too small: len=%d need_at_least=%d", len(buf), minLen)
	}
	img, err := newCPPBackendImage(width, height)
	if err != nil {
		return nil, err
	}
	pix, err := img.img.pixelView()
	if err != nil {
		return nil, err
	}
	dstStride := img.img.stride()
	for y := 0; y < height; y++ {
		srcY := y
		if stride < 0 {
			srcY = height - 1 - y
		}
		copy(pix[y*dstStride:y*dstStride+width*4], buf[srcY*abs(stride):srcY*abs(stride)+width*4])
	}
	return img, nil
}

func loadCPPBackendImageFromFileAvailable(filename string) (Image, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	src, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return newCPPBackendImageFromGoImageAvailable(src)
}
