// Package agg2d float image-convenience and export methods (L5/breadth).
// Thin float twins of the 8-bit Agg2D whole-image copy/blend convenience
// wrappers (image.go) and the PPM exporter (agg.go). The transfer primitives
// CopyImageFloat / BlendImageFloat (image_float.go) carry the actual pixel work;
// these add the float-destination + default-alpha spellings the 8-bit public
// surface exposes, plus a straight-RGB PPM writer over the attached float buffer.
package agg2d

import (
	"errors"
	"fmt"
	"os"

	"github.com/cwbudde/agg_go/internal/basics"
)

// CopyImageSimple copies the whole float source image to the destination at the
// given world coordinates without blending. Mirrors the 8-bit CopyImageSimple:
// the destination is transformed through the world matrix and truncated to
// integers before the rectangle-aligned copy.
func (a *Agg2DFloat) CopyImageSimple(img *ImageFloat, dstX, dstY float64) error {
	if img == nil {
		return errors.New("image is nil")
	}
	a.WorldToScreen(&dstX, &dstY)
	a.CopyImageFloat(img, int(dstX), int(dstY))
	return nil
}

// BlendImageSimple blends the whole float source image to the destination at the
// given world coordinates with the supplied alpha (0..255). Mirrors the 8-bit
// BlendImageSimple: world-transformed, integer-truncated destination.
func (a *Agg2DFloat) BlendImageSimple(img *ImageFloat, dstX, dstY float64, alpha uint) error {
	if img == nil {
		return errors.New("image is nil")
	}
	if alpha > 255 {
		alpha = 255
	}
	a.WorldToScreen(&dstX, &dstY)
	a.BlendImageFloat(img, int(dstX), int(dstY), basics.Int8u(alpha))
	return nil
}

// BlendImageDefaultAlpha blends the whole float source image at the given integer
// destination using the upstream default alpha of 255. Float twin of the 8-bit
// BlendImageDefaultAlpha convenience.
func (a *Agg2DFloat) BlendImageDefaultAlpha(img *ImageFloat, dstX, dstY int) {
	if img == nil {
		return
	}
	a.BlendImageFloat(img, dstX, dstY, 255)
}

// BlendImageSimpleDefaultAlpha blends the whole float source image to the world
// destination using the upstream default alpha of 255. Float twin of the 8-bit
// BlendImageSimpleDefaultAlpha convenience.
func (a *Agg2DFloat) BlendImageSimpleDefaultAlpha(img *ImageFloat, dstX, dstY float64) error {
	return a.BlendImageSimple(img, dstX, dstY, 255)
}

// SaveImagePPM writes the currently attached float buffer as a binary PPM file.
// Float twin of the 8-bit SaveImagePPM: the straight RGB channels are clamped
// and rounded to 8 bits (alpha is dropped, as PPM has no alpha channel).
func (a *Agg2DFloat) SaveImagePPM(filename string) error {
	if a.rbuf == nil {
		return fmt.Errorf("no attached float buffer")
	}
	width := a.rbuf.Width()
	height := a.rbuf.Height()
	if width <= 0 || height <= 0 {
		return fmt.Errorf("no attached float buffer")
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := fmt.Fprintf(file, "P6\n%d %d\n255\n", width, height); err != nil {
		return err
	}

	rgbRow := make([]byte, width*3)
	for y := 0; y < height; y++ {
		row := a.rbuf.Row(y)
		for x := 0; x < width; x++ {
			off := x * imageFloatChannels
			dst := x * 3
			if off+2 >= len(row) {
				return fmt.Errorf("attached buffer too small for %dx%d image", width, height)
			}
			rgbRow[dst] = roundToU8(row[off])
			rgbRow[dst+1] = roundToU8(row[off+1])
			rgbRow[dst+2] = roundToU8(row[off+2])
		}
		if _, err := file.Write(rgbRow); err != nil {
			return err
		}
	}

	return nil
}
