package agg

import (
	"errors"
	"fmt"
	"math"
	"unsafe"

	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/order"
	"github.com/cwbudde/agg_go/internal/pixfmt/blender"
)

// AlphaMode describes how RGB channels are stored relative to alpha.
//
// CompositeImage requires source and destination to use the same mode. This
// keeps the public surface small while still exposing AGG's straight- and
// premultiplied-alpha compositors.
type AlphaMode uint8

const (
	// AlphaStraight stores RGB independently from alpha.
	AlphaStraight AlphaMode = iota
	// AlphaPremultiplied stores RGB multiplied by alpha.
	AlphaPremultiplied
)

// DissolveSeedFunc returns a deterministic random value for a destination
// pixel. CompositeImage passes document/destination coordinates, so clipping a
// render does not change its dissolve pattern.
type DissolveSeedFunc func(x, y int) uint32

// CompositeOptions configures CompositeImage.
type CompositeOptions struct {
	BlendMode    BlendMode
	Opacity      float64
	AlphaMode    AlphaMode
	Mask         *AlphaMask
	MaskOrigin   PointI
	Clip         *Rect
	DissolveSeed DissolveSeedFunc
}

// DefaultDissolveSeed is the stable coordinate hash used when CompositeOptions
// does not provide a DissolveSeed function. Its output is part of the public
// rendering contract and will not change between releases.
func DefaultDissolveSeed(x, y int) uint32 {
	return blender.DissolveSeed(x, y)
}

// CompositeImage composites a half-open source rectangle onto dst.
//
// srcRect is expressed in source coordinates and dstOrigin is the destination
// position corresponding to srcRect's top-left corner. Clip and MaskOrigin are
// expressed in destination coordinates. Empty or fully clipped rectangles are
// successful no-ops.
//
// Both images must contain RGBA bytes, have a stride whose absolute value is at
// least width*4, and use the common alpha convention selected by opts.AlphaMode.
// CompositeImage writes only the clipped destination pixels and normally does
// not allocate. If source and destination storage overlap, it snapshots only
// the clipped source rectangle before writing.
func CompositeImage(dst, src *Image, srcRect Rect, dstOrigin PointI, opts CompositeOptions) error {
	if err := validateCompositeImage("destination", dst); err != nil {
		return err
	}
	if err := validateCompositeImage("source", src); err != nil {
		return err
	}
	if opts.AlphaMode != AlphaStraight && opts.AlphaMode != AlphaPremultiplied {
		return fmt.Errorf("agg: invalid alpha mode %d", opts.AlphaMode)
	}
	if math.IsNaN(opts.Opacity) || math.IsInf(opts.Opacity, 0) || opts.Opacity < 0 || opts.Opacity > 1 {
		return fmt.Errorf("agg: opacity must be finite and in [0,1], got %g", opts.Opacity)
	}
	if opts.Mask != nil {
		if err := validateCompositeMask(opts.Mask); err != nil {
			return err
		}
	}

	op, err := compositeOperation(opts.BlendMode)
	if err != nil {
		return err
	}

	region, ok := clipCompositeRegion(dst, src, srcRect, dstOrigin, opts.Clip)
	if !ok || opts.Opacity == 0 {
		return nil
	}

	input := compositeInput{img: src, x: region.srcX, y: region.srcY}
	if byteSlicesOverlap(dst.Data, src.Data) {
		input, err = snapshotCompositeInput(src, region)
		if err != nil {
			return err
		}
	}

	if opts.BlendMode == BlendDissolve {
		op = blender.CompOpSrcOver
	}
	if opts.AlphaMode == AlphaPremultiplied {
		compositor := blender.NewCompositeBlenderPre[color.Linear, order.RGBA](op)
		compositeRegion(dst, compositor, input, region, opts)
		return nil
	}

	compositor := blender.NewCompositeBlenderPlain[color.Linear, order.RGBA](op)
	if opts.Opacity == 1 && opts.Mask == nil && opts.BlendMode != BlendDissolve {
		compositeStraightRows(dst, compositor, input, region)
		return nil
	}
	compositeRegion(dst, compositor, input, region, opts)
	return nil
}

func compositeStraightRows(dst *Image, compositor blender.CompositeBlenderPlain[color.Linear, order.RGBA], input compositeInput, region compositeRegionBounds) {
	for rowIndex := 0; rowIndex < region.height; rowIndex++ {
		srcRow := input.row(input.y + rowIndex)
		dstRow := dst.renBuf.Row(region.dstY + rowIndex)
		srcOffset := input.x * 4
		dstOffset := region.dstX * 4
		r, g, b, a := srcRow[srcOffset], srcRow[srcOffset+1], srcRow[srcOffset+2], srcRow[srcOffset+3]
		uniform := true
		for column := 1; column < region.width; column++ {
			offset := srcOffset + column*4
			if srcRow[offset] != r || srcRow[offset+1] != g || srcRow[offset+2] != b || srcRow[offset+3] != a {
				uniform = false
				break
			}
		}
		if uniform {
			compositor.BlendSolidSpanStraight(dstRow[dstOffset:dstOffset+region.width*4], r, g, b, a, nil, region.width)
			continue
		}
		for column := 0; column < region.width; column++ {
			offset := srcOffset + column*4
			compositor.BlendPixFloat(dstRow[dstOffset+column*4:dstOffset+column*4+4], srcRow[offset], srcRow[offset+1], srcRow[offset+2], srcRow[offset+3], 1)
		}
	}
}

type compositeRegionBounds struct {
	srcX, srcY int
	dstX, dstY int
	width      int
	height     int
}

type compositeInput struct {
	img    *Image
	pixels []byte
	stride int
	x, y   int
}

func (input compositeInput) row(y int) []byte {
	if input.pixels != nil {
		start := y * input.stride
		return input.pixels[start : start+input.stride]
	}
	return input.img.renBuf.Row(y)
}

type compositePixelWriter interface {
	BlendPixFloat(dst []uint8, r, g, b, a uint8, coverage float64)
}

func compositeRegion[Compositor compositePixelWriter](dst *Image, compositor Compositor, input compositeInput, region compositeRegionBounds, opts CompositeOptions) {
	seed := opts.DissolveSeed
	if seed == nil {
		seed = DefaultDissolveSeed
	}
	for rowIndex := 0; rowIndex < region.height; rowIndex++ {
		srcRow := input.row(input.y + rowIndex)
		dstRow := dst.renBuf.Row(region.dstY + rowIndex)
		for column := 0; column < region.width; column++ {
			srcOffset := (input.x + column) * 4
			dstOffset := (region.dstX + column) * 4
			dstX := region.dstX + column
			dstY := region.dstY + rowIndex
			coverage := opts.Opacity
			if opts.Mask != nil {
				coverage *= float64(opts.Mask.At(dstX-opts.MaskOrigin.X, dstY-opts.MaskOrigin.Y)) / 255
			}

			r, g, b, a := srcRow[srcOffset], srcRow[srcOffset+1], srcRow[srcOffset+2], srcRow[srcOffset+3]
			if opts.BlendMode == BlendDissolve {
				effectiveAlpha := float64(a) / 255 * coverage
				if !blender.DissolveAccept(effectiveAlpha, seed(dstX, dstY)) {
					continue
				}
				if opts.AlphaMode == AlphaPremultiplied {
					r = dissolveStraightChannel(r, a)
					g = dissolveStraightChannel(g, a)
					b = dissolveStraightChannel(b, a)
				}
				a = 255
				coverage = 1
			}
			compositor.BlendPixFloat(dstRow[dstOffset:dstOffset+4], r, g, b, a, coverage)
		}
	}
}

func dissolveStraightChannel(channel, alpha uint8) uint8 {
	if alpha == 0 {
		return 0
	}
	value := (int(channel)*255 + int(alpha)/2) / int(alpha)
	return uint8(min(value, 255))
}

func validateCompositeImage(name string, img *Image) error {
	if img == nil {
		return fmt.Errorf("agg: %s image is nil", name)
	}
	if img.width < 0 || img.height < 0 {
		return fmt.Errorf("agg: %s image has negative dimensions %dx%d", name, img.width, img.height)
	}
	if img.width == 0 || img.height == 0 {
		if len(img.Data) != 0 && img.renBuf == nil {
			return fmt.Errorf("agg: %s image has no rendering buffer", name)
		}
		return nil
	}
	if img.renBuf == nil {
		return fmt.Errorf("agg: %s image has no rendering buffer", name)
	}
	if img.width > int(^uint(0)>>1)/4 {
		return fmt.Errorf("agg: %s image row size overflows int", name)
	}
	rowBytes := img.width * 4
	stride := img.Stride()
	absStride := stride
	if absStride < 0 {
		if absStride == -int(^uint(0)>>1)-1 {
			return fmt.Errorf("agg: %s image stride overflows int", name)
		}
		absStride = -absStride
	}
	if absStride < rowBytes {
		return fmt.Errorf("agg: %s image stride %d is smaller than RGBA row size %d", name, stride, rowBytes)
	}
	if img.height-1 > (int(^uint(0)>>1)-rowBytes)/absStride {
		return fmt.Errorf("agg: %s image buffer size overflows int", name)
	}
	required := (img.height-1)*absStride + rowBytes
	if len(img.Data) < required {
		return fmt.Errorf("agg: %s image buffer has %d bytes, need at least %d", name, len(img.Data), required)
	}
	return nil
}

func validateCompositeMask(mask *AlphaMask) error {
	if mask.Width < 0 || mask.Height < 0 {
		return fmt.Errorf("agg: alpha mask has negative dimensions %dx%d", mask.Width, mask.Height)
	}
	if mask.Width != 0 && mask.Height > int(^uint(0)>>1)/mask.Width {
		return errors.New("agg: alpha mask size overflows int")
	}
	required := mask.Width * mask.Height
	if len(mask.Pix) < required {
		return fmt.Errorf("agg: alpha mask has %d samples, need at least %d", len(mask.Pix), required)
	}
	return nil
}

func clipCompositeRegion(dst, src *Image, srcRect Rect, dstOrigin PointI, clip *Rect) (compositeRegionBounds, bool) {
	if srcRect.X2 <= srcRect.X1 || srcRect.Y2 <= srcRect.Y1 {
		return compositeRegionBounds{}, false
	}

	region := compositeRegionBounds{
		srcX: srcRect.X1, srcY: srcRect.Y1,
		dstX: dstOrigin.X, dstY: dstOrigin.Y,
		width: srcRect.X2 - srcRect.X1, height: srcRect.Y2 - srcRect.Y1,
	}
	clipCompositeLeft(&region, -region.srcX)
	clipCompositeTop(&region, -region.srcY)
	region.width = min(region.width, src.width-region.srcX)
	region.height = min(region.height, src.height-region.srcY)

	clipCompositeLeft(&region, -region.dstX)
	clipCompositeTop(&region, -region.dstY)
	region.width = min(region.width, dst.width-region.dstX)
	region.height = min(region.height, dst.height-region.dstY)

	if clip != nil {
		clipCompositeLeft(&region, clip.X1-region.dstX)
		clipCompositeTop(&region, clip.Y1-region.dstY)
		region.width = min(region.width, clip.X2-region.dstX)
		region.height = min(region.height, clip.Y2-region.dstY)
	}
	return region, region.width > 0 && region.height > 0
}

func clipCompositeLeft(region *compositeRegionBounds, amount int) {
	if amount <= 0 {
		return
	}
	region.srcX += amount
	region.dstX += amount
	region.width -= amount
}

func clipCompositeTop(region *compositeRegionBounds, amount int) {
	if amount <= 0 {
		return
	}
	region.srcY += amount
	region.dstY += amount
	region.height -= amount
}

func byteSlicesOverlap(first, second []byte) bool {
	if len(first) == 0 || len(second) == 0 {
		return false
	}
	firstStart := uintptr(unsafe.Pointer(&first[0]))
	secondStart := uintptr(unsafe.Pointer(&second[0]))
	return firstStart < secondStart+uintptr(len(second)) && secondStart < firstStart+uintptr(len(first))
}

func snapshotCompositeInput(src *Image, region compositeRegionBounds) (compositeInput, error) {
	if region.width > int(^uint(0)>>1)/4 || region.height > int(^uint(0)>>1)/(region.width*4) {
		return compositeInput{}, errors.New("agg: composite snapshot size overflows int")
	}
	stride := region.width * 4
	pixels := make([]byte, stride*region.height)
	for y := 0; y < region.height; y++ {
		srcRow := src.renBuf.Row(region.srcY + y)
		start := region.srcX * 4
		copy(pixels[y*stride:(y+1)*stride], srcRow[start:start+stride])
	}
	return compositeInput{pixels: pixels, stride: stride}, nil
}

func compositeOperation(mode BlendMode) (blender.CompOp, error) {
	switch mode {
	case BlendAlpha, BlendSrcOver:
		return blender.CompOpSrcOver, nil
	case BlendClear:
		return blender.CompOpClear, nil
	case BlendSrc:
		return blender.CompOpSrc, nil
	case BlendDst:
		return blender.CompOpDst, nil
	case BlendDstOver:
		return blender.CompOpDstOver, nil
	case BlendSrcIn:
		return blender.CompOpSrcIn, nil
	case BlendDstIn:
		return blender.CompOpDstIn, nil
	case BlendSrcOut:
		return blender.CompOpSrcOut, nil
	case BlendDstOut:
		return blender.CompOpDstOut, nil
	case BlendSrcAtop:
		return blender.CompOpSrcAtop, nil
	case BlendDstAtop:
		return blender.CompOpDstAtop, nil
	case BlendXor:
		return blender.CompOpXor, nil
	case BlendAdd:
		return blender.CompOpPlus, nil
	case BlendMultiply:
		return blender.CompOpMultiply, nil
	case BlendScreen:
		return blender.CompOpScreen, nil
	case BlendOverlay:
		return blender.CompOpOverlay, nil
	case BlendDarken:
		return blender.CompOpDarken, nil
	case BlendLighten:
		return blender.CompOpLighten, nil
	case BlendColorDodge:
		return blender.CompOpColorDodge, nil
	case BlendColorBurn:
		return blender.CompOpColorBurn, nil
	case BlendHardLight:
		return blender.CompOpHardLight, nil
	case BlendSoftLight:
		return blender.CompOpSoftLight, nil
	case BlendDifference:
		return blender.CompOpDifference, nil
	case BlendExclusion:
		return blender.CompOpExclusion, nil
	case BlendDissolve:
		return blender.CompOpDissolve, nil
	case BlendLinearBurn:
		return blender.CompOpLinearBurn, nil
	case BlendDarkerColor:
		return blender.CompOpDarkerColor, nil
	case BlendLinearDodge:
		return blender.CompOpPlus, nil
	case BlendLighterColor:
		return blender.CompOpLighterColor, nil
	case BlendVividLight:
		return blender.CompOpVividLight, nil
	case BlendLinearLight:
		return blender.CompOpLinearLight, nil
	case BlendPinLight:
		return blender.CompOpPinLight, nil
	case BlendHardMix:
		return blender.CompOpHardMix, nil
	case BlendSubtract:
		return blender.CompOpSubtract, nil
	case BlendDivide:
		return blender.CompOpDivide, nil
	case BlendHue:
		return blender.CompOpHue, nil
	case BlendSaturation:
		return blender.CompOpSaturation, nil
	case BlendColor:
		return blender.CompOpColor, nil
	case BlendLuminosity:
		return blender.CompOpLuminosity, nil
	case BlendColorBurnPhotoshop:
		return blender.CompOpColorBurnPhotoshop, nil
	case BlendSoftLightPhotoshop:
		return blender.CompOpSoftLightPhotoshop, nil
	default:
		return 0, fmt.Errorf("agg: unsupported blend mode %d", mode)
	}
}
