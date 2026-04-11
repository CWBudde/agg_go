package agg

import (
	"math"

	intcolor "github.com/cwbudde/agg_go/internal/color"
)

type AlphaMask struct {
	Width  int
	Height int
	Pix    []uint8
}

// LinearGradientFill describes the two-color masked gradient used by the
// public layer-effect helpers.
type LinearGradientFill struct {
	Start   Color
	End     Color
	Angle   float64
	Scale   float64
	Reverse bool
}

// CheckerPatternFill describes the simple alternating-color checker fill used
// by the public layer-effect helpers.
type CheckerPatternFill struct {
	First  Color
	Second Color
	Scale  float64
}

// AlphaMaskFromRGBA extracts the alpha channel from an RGBA surface into a
// reusable mask object.
func AlphaMaskFromRGBA(surface []byte, width, height int) AlphaMask {
	mask := AlphaMask{
		Width:  width,
		Height: height,
		Pix:    make([]uint8, maxInt(0, width*height)),
	}
	for index := 0; index < len(mask.Pix); index++ {
		offset := index * 4
		if offset+3 >= len(surface) {
			break
		}
		mask.Pix[index] = surface[offset+3]
	}
	return mask
}

// NewAlphaMask allocates an empty mask with the requested dimensions.
func NewAlphaMask(width, height int) AlphaMask {
	return AlphaMask{
		Width:  width,
		Height: height,
		Pix:    make([]uint8, maxInt(0, width*height)),
	}
}

// Clone returns a copy of the mask and its backing storage.
func (mask AlphaMask) Clone() AlphaMask {
	cloned := NewAlphaMask(mask.Width, mask.Height)
	copy(cloned.Pix, mask.Pix)
	return cloned
}

// At returns the alpha value at x,y or 0 outside bounds.
func (mask AlphaMask) At(x, y int) uint8 {
	if x < 0 || y < 0 || x >= mask.Width || y >= mask.Height {
		return 0
	}
	return mask.Pix[y*mask.Width+x]
}

// SetPixel writes one alpha sample when x,y is inside the mask bounds.
func (mask *AlphaMask) SetPixel(x, y int, alpha uint8) {
	if x < 0 || y < 0 || x >= mask.Width || y >= mask.Height {
		return
	}
	mask.Pix[y*mask.Width+x] = alpha
}

// Fill assigns one alpha value to the whole mask.
func (mask *AlphaMask) Fill(alpha uint8) {
	for index := range mask.Pix {
		mask.Pix[index] = alpha
	}
}

// Shifted returns a translated copy of the mask. Overlapping writes keep the
// maximum alpha sample, matching AGG-style coverage composition.
func (mask AlphaMask) Shifted(dx, dy int) AlphaMask {
	shifted := NewAlphaMask(mask.Width, mask.Height)
	for index, alpha := range mask.Pix {
		if alpha == 0 {
			continue
		}
		x := index % mask.Width
		y := index / mask.Width
		dstX := x + dx
		dstY := y + dy
		if dstX < 0 || dstY < 0 || dstX >= mask.Width || dstY >= mask.Height {
			continue
		}
		dstIndex := dstY*mask.Width + dstX
		if alpha > shifted.Pix[dstIndex] {
			shifted.Pix[dstIndex] = alpha
		}
	}
	return shifted
}

// Dilated returns a box-filter expansion of the mask by radius pixels.
func (mask AlphaMask) Dilated(radius int) AlphaMask {
	if radius <= 0 {
		return mask.Clone()
	}
	dilated := NewAlphaMask(mask.Width, mask.Height)
	for y := 0; y < mask.Height; y++ {
		for x := 0; x < mask.Width; x++ {
			var maxAlpha uint8
			for sampleY := maxInt(0, y-radius); sampleY <= minInt(mask.Height-1, y+radius); sampleY++ {
				for sampleX := maxInt(0, x-radius); sampleX <= minInt(mask.Width-1, x+radius); sampleX++ {
					alpha := mask.At(sampleX, sampleY)
					if alpha > maxAlpha {
						maxAlpha = alpha
					}
				}
			}
			dilated.Pix[y*mask.Width+x] = maxAlpha
		}
	}
	return dilated
}

// Eroded returns a box-filter contraction of the mask by radius pixels.
func (mask AlphaMask) Eroded(radius int) AlphaMask {
	if radius <= 0 {
		return mask.Clone()
	}
	eroded := NewAlphaMask(mask.Width, mask.Height)
	for y := 0; y < mask.Height; y++ {
		for x := 0; x < mask.Width; x++ {
			minAlpha := uint8(255)
			for sampleY := y - radius; sampleY <= y+radius; sampleY++ {
				for sampleX := x - radius; sampleX <= x+radius; sampleX++ {
					if sampleX < 0 || sampleY < 0 || sampleX >= mask.Width || sampleY >= mask.Height {
						minAlpha = 0
						continue
					}
					alpha := mask.At(sampleX, sampleY)
					if alpha < minAlpha {
						minAlpha = alpha
					}
				}
			}
			eroded.Pix[y*mask.Width+x] = minAlpha
		}
	}
	return eroded
}

// Blurred returns a softened copy of the mask using stack blur.
func (mask AlphaMask) Blurred(radius int) AlphaMask {
	if radius <= 0 {
		return mask.Clone()
	}
	expanded := make([]uint8, len(mask.Pix)*4)
	for index, alpha := range mask.Pix {
		offset := index * 4
		expanded[offset] = alpha
		expanded[offset+1] = alpha
		expanded[offset+2] = alpha
		expanded[offset+3] = alpha
	}
	stackBlur := NewStackBlur[intcolor.Linear]()
	stackBlur.BlurRGBA8(expanded, mask.Width, mask.Height, mask.Width*4, radius)
	blurred := NewAlphaMask(mask.Width, mask.Height)
	for index := range blurred.Pix {
		blurred.Pix[index] = expanded[index*4+3]
	}
	return blurred
}

// Subtract subtracts the other mask's coverage from this mask, clamping at 0.
func (mask AlphaMask) Subtract(other AlphaMask) AlphaMask {
	out := NewAlphaMask(mask.Width, mask.Height)
	for y := 0; y < mask.Height; y++ {
		for x := 0; x < mask.Width; x++ {
			index := y*mask.Width + x
			value := int(mask.At(x, y)) - int(other.At(x, y))
			if value < 0 {
				value = 0
			}
			out.Pix[index] = uint8(value)
		}
	}
	return out
}

// Intersect multiplies the two masks' coverages per pixel. This is a coverage
// composition operation, not a geometric min() intersection.
func (mask AlphaMask) Intersect(other AlphaMask) AlphaMask {
	out := NewAlphaMask(mask.Width, mask.Height)
	for y := 0; y < mask.Height; y++ {
		for x := 0; x < mask.Width; x++ {
			out.Pix[y*mask.Width+x] = scaleAlpha(mask.At(x, y), other.At(x, y))
		}
	}
	return out
}

// Max keeps the stronger per-pixel coverage from either mask.
func (mask AlphaMask) Max(other AlphaMask) AlphaMask {
	out := NewAlphaMask(mask.Width, mask.Height)
	for y := 0; y < mask.Height; y++ {
		for x := 0; x < mask.Width; x++ {
			index := y*mask.Width + x
			left := mask.At(x, y)
			right := other.At(x, y)
			if left > right {
				out.Pix[index] = left
			} else {
				out.Pix[index] = right
			}
		}
	}
	return out
}

// AbsDiff returns the absolute per-pixel coverage difference between two masks.
func (mask AlphaMask) AbsDiff(other AlphaMask) AlphaMask {
	out := NewAlphaMask(mask.Width, mask.Height)
	for y := 0; y < mask.Height; y++ {
		for x := 0; x < mask.Width; x++ {
			index := y*mask.Width + x
			diff := int(mask.At(x, y)) - int(other.At(x, y))
			if diff < 0 {
				diff = -diff
			}
			out.Pix[index] = uint8(diff)
		}
	}
	return out
}

// RenderMaskedSolidRGBA returns an RGBA surface containing a solid color whose
// alpha is modulated by the supplied mask.
func RenderMaskedSolidRGBA(mask AlphaMask, color Color) []byte {
	src := CreateImageFromColor(mask.Width, mask.Height, color)
	return renderMaskedImageRGBA(mask, src)
}

// RenderMaskedLinearGradientRGBA returns a two-color RGBA gradient modulated by
// the supplied mask.
func RenderMaskedLinearGradientRGBA(mask AlphaMask, fill LinearGradientFill) []byte {
	surface := make([]byte, maxInt(0, mask.Width*mask.Height*4))
	centerX := float64(mask.Width-1) / 2
	centerY := float64(mask.Height-1) / 2
	maxAxis := math.Max(float64(mask.Width), float64(mask.Height))
	theta := fill.Angle * math.Pi / 180
	scale := math.Max(0.1, fill.Scale)
	span := math.Max(1, maxAxis*scale)
	cosTheta := math.Cos(theta)
	sinTheta := math.Sin(theta)
	x1 := centerX - cosTheta*span/2
	y1 := centerY + sinTheta*span/2
	x2 := centerX + cosTheta*span/2
	y2 := centerY - sinTheta*span/2
	start := fill.Start
	end := fill.End
	if fill.Reverse {
		start, end = end, start
	}
	denom := math.Max(1, (x2-x1)*(x2-x1)+(y2-y1)*(y2-y1))
	for y := 0; y < mask.Height; y++ {
		for x := 0; x < mask.Width; x++ {
			maskAlpha := mask.At(x, y)
			if maskAlpha == 0 {
				continue
			}
			projected := ((float64(x)-x1)*(x2-x1) + (float64(y)-y1)*(y2-y1)) / denom
			t := clampUnit(projected)
			color := start.Gradient(end, t)
			offset := (y*mask.Width + x) * 4
			surface[offset] = color.R
			surface[offset+1] = color.G
			surface[offset+2] = color.B
			surface[offset+3] = scaleAlpha(color.A, maskAlpha)
		}
	}
	return surface
}

// RenderMaskedCheckerPatternRGBA returns a checkerboard RGBA surface modulated
// by the supplied mask.
func RenderMaskedCheckerPatternRGBA(mask AlphaMask, fill CheckerPatternFill) []byte {
	src := CreateImage(mask.Width, mask.Height)
	ctx := NewContextForImage(src)
	ctx.Clear(fill.First)
	ctx.SetColor(fill.Second)
	tile := maxInt(1, int(math.Round(math.Max(1, fill.Scale)*4)))
	for y := 0; y < mask.Height; y += tile {
		for x := 0; x < mask.Width; x += tile {
			if ((x/tile)+(y/tile))%2 == 0 {
				continue
			}
			ctx.FillRectangle(float64(x), float64(y), float64(minInt(tile, mask.Width-x)), float64(minInt(tile, mask.Height-y)))
		}
	}
	return renderMaskedImageRGBA(mask, src)
}

func renderMaskedImageRGBA(mask AlphaMask, src *Image) []byte {
	dest := CreateImage(mask.Width, mask.Height)
	if src == nil || mask.Width <= 0 || mask.Height <= 0 {
		return dest.Data
	}
	width := minInt(mask.Width, src.Width())
	height := minInt(mask.Height, src.Height())
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			index := y*dest.Stride() + x*4
			srcIndex := y*src.Stride() + x*4
			copy(dest.Data[index:index+4], src.Data[srcIndex:srcIndex+4])
			dest.Data[index+3] = scaleAlpha(dest.Data[index+3], mask.At(x, y))
		}
	}
	return dest.Data
}

func (mask AlphaMask) alphaAt(index int) uint8 {
	if index < 0 || index >= len(mask.Pix) {
		return 0
	}
	return mask.Pix[index]
}

func scaleAlpha(alpha, maskAlpha uint8) uint8 {
	return uint8((uint16(alpha)*uint16(maskAlpha) + 127) / 255)
}

func clampUnit(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
