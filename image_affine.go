package agg

import (
	"errors"
	"fmt"
	"math"
)

// ImageEdgeMode controls samples whose filter footprint leaves the source.
type ImageEdgeMode uint8

const (
	// ImageEdgeClamp extends the nearest edge pixel.
	ImageEdgeClamp ImageEdgeMode = iota
	// ImageEdgeTransparent treats pixels outside the source as transparent.
	ImageEdgeTransparent
)

// ImageTransformOptions configures DrawImageAffine.
type ImageTransformOptions struct {
	Filter           ImageFilter
	FilterRadius     float64
	Resample         ImageResample
	EdgeMode         ImageEdgeMode
	SourceAlpha      AlphaMode
	DestinationAlpha AlphaMode
	BlendMode        BlendMode
	Opacity          float64
	Clip             *Rect
	// SampleOffset shifts source sampling without moving the transformed
	// destination polygon. It is useful when adapting APIs with a different
	// pixel-center convention; the zero value preserves standard AGG behavior.
	SampleOffset Point
}

// DrawImageAffine maps a half-open source rectangle through an affine transform
// and draws it into a caller-owned RGBA image using AGG's filtered span path.
// Both images contain RGBA bytes and may use positive or negative strides whose
// absolute values are at least width*4. Clip is half-open and expressed in
// destination coordinates; it is intersected with the destination bounds.
//
// DrawImageAffine creates per-call AGG rendering state. Straight-alpha sources
// are premultiplied before filtering to prevent colored transparent texels from
// bleeding into visible edges; this requires a source-sized copy unless the
// source is opaque. Premultiplied destinations are converted through a
// destination-sized straight-alpha image. ImageEdgeTransparent additionally
// allocates a transparent border large enough for the filter footprint.
func DrawImageAffine(dst, src *Image, srcRect Rect, sourceToDestination *Transformations, opts ImageTransformOptions) error {
	if err := validateCompositeImage("affine destination", dst); err != nil {
		return err
	}
	if err := validateCompositeImage("affine source", src); err != nil {
		return err
	}
	if sourceToDestination == nil {
		return errors.New("agg: affine transform is nil")
	}
	for _, coefficient := range sourceToDestination.AffineMatrix {
		if math.IsNaN(coefficient) || math.IsInf(coefficient, 0) {
			return errors.New("agg: affine transform contains a non-finite coefficient")
		}
	}
	inverse := sourceToDestination.Clone()
	if !inverse.Invert() {
		return errors.New("agg: affine transform is not invertible")
	}
	if opts.SourceAlpha > AlphaPremultiplied || opts.DestinationAlpha > AlphaPremultiplied {
		return errors.New("agg: invalid affine alpha mode")
	}
	if opts.EdgeMode != ImageEdgeClamp && opts.EdgeMode != ImageEdgeTransparent {
		return fmt.Errorf("agg: invalid affine edge mode %d", opts.EdgeMode)
	}
	if opts.Filter < ImageFilterNoFilter || opts.Filter > ImageFilterLanczos {
		return fmt.Errorf("agg: invalid affine image filter %d", opts.Filter)
	}
	if opts.Resample < NoResample || opts.Resample > ResampleOnZoomOut {
		return fmt.Errorf("agg: invalid affine resample mode %d", opts.Resample)
	}
	if _, err := compositeOperation(opts.BlendMode); err != nil {
		return fmt.Errorf("agg: invalid affine blend mode: %w", err)
	}
	if opts.BlendMode == BlendDissolve {
		return errors.New("agg: affine dissolve requires coordinate-aware sampling and is not supported")
	}
	if math.IsNaN(opts.Opacity) || math.IsInf(opts.Opacity, 0) || opts.Opacity < 0 || opts.Opacity > 1 {
		return fmt.Errorf("agg: opacity must be finite and in [0,1], got %g", opts.Opacity)
	}
	if opts.Opacity == 0 || srcRect.X1 >= srcRect.X2 || srcRect.Y1 >= srcRect.Y2 {
		return nil
	}
	if srcRect.X1 < 0 || srcRect.Y1 < 0 || srcRect.X2 > src.Width() || srcRect.Y2 > src.Height() {
		return errors.New("agg: affine source rectangle is outside the source image")
	}
	if opts.FilterRadius < 0 || math.IsNaN(opts.FilterRadius) || math.IsInf(opts.FilterRadius, 0) {
		return errors.New("agg: affine filter radius must be finite and non-negative")
	}
	if !finitePoint(opts.SampleOffset) {
		return errors.New("agg: affine sample offset must be finite")
	}

	if byteSlicesOverlap(dst.Data, src.Data) {
		src = cloneImageBytes(src)
	}
	renderSource, renderRect, err := prepareAffineSource(src, srcRect, sourceToDestination, opts)
	if err != nil {
		return err
	}
	renderTarget := dst
	if opts.DestinationAlpha == AlphaPremultiplied {
		renderTarget = cloneImageBytes(dst)
		if err := renderTarget.Demultiply(); err != nil {
			return err
		}
	}

	ctx := NewContextForImage(renderTarget)
	if ctx == nil {
		return errors.New("agg: could not attach affine destination")
	}
	ctx.SetBlendMode(opts.BlendMode)
	ctx.agg2d.ImageBlendMode(BlendDst)
	ctx.agg2d.ImageBlendColorRGBA(255, 255, 255, uint8(math.Round(opts.Opacity*255)))
	ctx.SetImageFilter(opts.Filter)
	ctx.SetImageResample(opts.Resample)
	if opts.FilterRadius > 0 {
		ctx.agg2d.SetImageFilterRadius(opts.Filter, opts.FilterRadius)
	}
	clip, ok := affineDestinationClip(dst, opts.Clip)
	if !ok {
		return nil
	}
	// Agg2D's renderer clip uses integer coordinates while its rasterizer clipper
	// is intentionally a no-op. Keep the renderer clip non-degenerate (notably
	// for 1x1 targets) and clip the actual image path below.
	ctx.agg2d.ClipBox(float64(clip.X1), float64(clip.Y1), float64(clip.X2-1), float64(clip.Y2-1))
	x0, y0 := sourceToDestination.Transform(float64(srcRect.X1), float64(srcRect.Y1))
	x1, y1 := sourceToDestination.Transform(float64(srcRect.X2), float64(srcRect.Y1))
	// TransformImageParallelogram's third point corresponds to the source
	// rectangle's bottom-right corner (the fourth is derived by parallelogram
	// closure), despite the older high-level wrapper passing bottom-left.
	x3, y3 := sourceToDestination.Transform(float64(srcRect.X2), float64(srcRect.Y2))
	// Keep the rasterized path one pixel beyond the write clip. Clipping the
	// geometry exactly on a pixel boundary can turn a fully covered boundary
	// pixel into a partially covered one; the renderer clip still guarantees
	// that no writes escape the caller's half-open rectangle.
	pathClip := Rect{
		X1: 0,
		Y1: max(0, clip.Y1-2),
		X2: dst.Width(),
		Y2: min(dst.Height(), clip.Y2+2),
	}
	destinationPath := clipAffinePolygon([]affinePoint{
		{x: x0, y: y0},
		{x: x1, y: y1},
		{x: x3, y: y3},
		{x: x0 + x3 - x1, y: y0 + y3 - y1},
	}, pathClip)
	if len(destinationPath) < 3 {
		return nil
	}
	ctx.agg2d.ResetPath()
	ctx.agg2d.MoveTo(destinationPath[0].x, destinationPath[0].y)
	for _, point := range destinationPath[1:] {
		ctx.agg2d.LineTo(point.x, point.y)
	}
	ctx.agg2d.ClosePolygon()
	if err := ctx.agg2d.transformImagePathParallelogramFloat(
		renderSource,
		float64(renderRect.X1)+opts.SampleOffset.X,
		float64(renderRect.Y1)+opts.SampleOffset.Y,
		float64(renderRect.X2)+opts.SampleOffset.X,
		float64(renderRect.Y2)+opts.SampleOffset.Y,
		[]float64{x0, y0, x1, y1, x3, y3},
	); err != nil {
		return err
	}
	if opts.DestinationAlpha == AlphaPremultiplied {
		if err := renderTarget.Premultiply(); err != nil {
			return err
		}
		copy(dst.Data, renderTarget.Data)
	}
	return nil
}

func prepareAffineSource(src *Image, srcRect Rect, transform *Transformations, opts ImageTransformOptions) (*Image, Rect, error) {
	renderSource := src
	renderRect := srcRect
	needsPremultiply := opts.SourceAlpha == AlphaStraight && !affineImageOpaque(src)
	if opts.EdgeMode == ImageEdgeTransparent {
		padX, padY, err := affineTransparentPadding(transform, opts)
		if err != nil {
			return nil, Rect{}, err
		}
		renderSource, err = transparentBorderedImage(src, padX, padY)
		if err != nil {
			return nil, Rect{}, err
		}
		renderRect = Rect{
			X1: srcRect.X1 + padX,
			Y1: srcRect.Y1 + padY,
			X2: srcRect.X2 + padX,
			Y2: srcRect.Y2 + padY,
		}
	}
	if needsPremultiply {
		if renderSource == src {
			renderSource = cloneImageBytes(src)
		}
		if err := renderSource.Premultiply(); err != nil {
			return nil, Rect{}, err
		}
	}
	return renderSource, renderRect, nil
}

func affineImageOpaque(img *Image) bool {
	rowBytes := img.Width() * 4
	for y := 0; y < img.Height(); y++ {
		row := img.renBuf.Row(y)
		for offset := 3; offset < rowBytes; offset += 4 {
			if row[offset] != 255 {
				return false
			}
		}
	}
	return true
}

func affineTransparentPadding(transform *Transformations, opts ImageTransformOptions) (int, int, error) {
	if opts.Filter == ImageFilterNoFilter {
		return 1, 1, nil
	}
	radius := 4.0 // Largest built-in default support (Blackman/Sinc/Lanczos).
	if opts.FilterRadius > radius {
		radius = math.Ceil(opts.FilterRadius)
	}
	scaleX, scaleY := 1.0, 1.0
	inverse := transform.Clone()
	if !inverse.Invert() {
		return 0, 0, errors.New("agg: affine transform is not invertible")
	}
	m := inverse.AffineMatrix
	inverseScaleX := math.Hypot(m[0], m[2])
	inverseScaleY := math.Hypot(m[1], m[3])
	resample := opts.Resample == ResampleAlways ||
		(opts.Resample == ResampleOnZoomOut && (inverseScaleX > 1.125 || inverseScaleY > 1.125))
	if resample {
		scaleX, scaleY = inverseScaleX, inverseScaleY
		const scaleLimit = 200.0
		if scaleProduct := scaleX * scaleY; scaleProduct > scaleLimit {
			scaleX *= scaleLimit / scaleProduct
			scaleY *= scaleLimit / scaleProduct
		}
		scaleX = min(max(scaleX, 1), scaleLimit)
		scaleY = min(max(scaleY, 1), scaleLimit)
	}
	padXFloat := math.Ceil(radius*scaleX) + 2
	padYFloat := math.Ceil(radius*scaleY) + 2
	maxInt := int(^uint(0) >> 1)
	if padXFloat > float64(maxInt) || padYFloat > float64(maxInt) {
		return 0, 0, errors.New("agg: transparent affine filter footprint overflows int")
	}
	return int(padXFloat), int(padYFloat), nil
}

func transparentBorderedImage(src *Image, padX, padY int) (*Image, error) {
	maxInt := int(^uint(0) >> 1)
	if padX < 0 || padY < 0 || padX > (maxInt-src.Width())/2 || padY > (maxInt-src.Height())/2 {
		return nil, errors.New("agg: transparent affine border dimensions overflow int")
	}
	width := src.Width() + 2*padX
	height := src.Height() + 2*padY
	if width > maxInt/4 || (width != 0 && height > maxInt/(width*4)) {
		return nil, errors.New("agg: transparent affine image size overflows int")
	}
	stride := width * 4
	result := NewImage(make([]byte, height*stride), width, height, stride)
	rowBytes := src.Width() * 4
	for y := 0; y < src.Height(); y++ {
		sourceRow := src.renBuf.Row(y)
		destinationRow := result.renBuf.Row(y + padY)
		copy(destinationRow[padX*4:padX*4+rowBytes], sourceRow[:rowBytes])
	}
	return result, nil
}

func affineDestinationClip(dst *Image, requested *Rect) (Rect, bool) {
	clip := Rect{X2: dst.Width(), Y2: dst.Height()}
	if requested != nil {
		clip.X1 = max(clip.X1, requested.X1)
		clip.Y1 = max(clip.Y1, requested.Y1)
		clip.X2 = min(clip.X2, requested.X2)
		clip.Y2 = min(clip.Y2, requested.Y2)
	}
	return clip, clip.X1 < clip.X2 && clip.Y1 < clip.Y2
}

type affinePoint struct{ x, y float64 }

func clipAffinePolygon(polygon []affinePoint, clip Rect) []affinePoint {
	type clipEdge struct {
		inside    func(affinePoint) bool
		intersect func(affinePoint, affinePoint) affinePoint
	}
	x1, y1 := float64(clip.X1), float64(clip.Y1)
	x2, y2 := float64(clip.X2), float64(clip.Y2)
	edges := [...]clipEdge{
		{func(p affinePoint) bool { return p.x >= x1 }, func(a, b affinePoint) affinePoint {
			t := (x1 - a.x) / (b.x - a.x)
			return affinePoint{x: x1, y: a.y + t*(b.y-a.y)}
		}},
		{func(p affinePoint) bool { return p.x <= x2 }, func(a, b affinePoint) affinePoint {
			t := (x2 - a.x) / (b.x - a.x)
			return affinePoint{x: x2, y: a.y + t*(b.y-a.y)}
		}},
		{func(p affinePoint) bool { return p.y >= y1 }, func(a, b affinePoint) affinePoint {
			t := (y1 - a.y) / (b.y - a.y)
			return affinePoint{x: a.x + t*(b.x-a.x), y: y1}
		}},
		{func(p affinePoint) bool { return p.y <= y2 }, func(a, b affinePoint) affinePoint {
			t := (y2 - a.y) / (b.y - a.y)
			return affinePoint{x: a.x + t*(b.x-a.x), y: y2}
		}},
	}
	output := polygon
	for _, edge := range edges {
		input := output
		output = nil
		if len(input) == 0 {
			break
		}
		previous := input[len(input)-1]
		previousInside := edge.inside(previous)
		for _, current := range input {
			currentInside := edge.inside(current)
			if currentInside != previousInside {
				output = append(output, edge.intersect(previous, current))
			}
			if currentInside {
				output = append(output, current)
			}
			previous, previousInside = current, currentInside
		}
	}
	return output
}

func cloneImageBytes(src *Image) *Image {
	pixels := append([]byte(nil), src.Data...)
	return NewImage(pixels, src.Width(), src.Height(), src.Stride())
}
