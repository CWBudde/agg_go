// Package outline provides anti-aliased outline rendering functionality.
// This implements a port of AGG's agg_renderer_outline_image.h for image-based outline rendering.
package outline

import (
	"math"

	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/primitives"
)

// Source interface defines methods needed by LineImageScale.
type Source interface {
	Width() float64
	Height() float64
	Pixel(x, y int) color.RGBA
}

// LineImageScale scales a source image to a specified height for line patterns.
// This is equivalent to AGG's line_image_scale template class.
type LineImageScale struct {
	source   Source
	height   float64
	scale    float64
	scaleInv float64
}

// NewLineImageScale creates a new line image scaler.
func NewLineImageScale(src Source, height float64) *LineImageScale {
	srcHeight := src.Height()
	if srcHeight == 0 {
		srcHeight = 1
	}

	return &LineImageScale{
		source:   src,
		height:   height,
		scale:    srcHeight / height,
		scaleInv: height / srcHeight,
	}
}

// Width returns the scaled width (same as source).
func (lis *LineImageScale) Width() float64 {
	return lis.source.Width()
}

// Height returns the target height.
func (lis *LineImageScale) Height() float64 {
	return lis.height
}

// Pixel returns the scaled pixel at the specified coordinates.
func (lis *LineImageScale) Pixel(x, y int) color.RGBA {
	if lis.scale < 1.0 {
		// Interpolate between nearest source pixels
		srcY := (float64(y)+0.5)*lis.scale - 0.5
		h := int(lis.source.Height()) - 1
		y1 := int(math.Floor(srcY))
		y2 := y1 + 1

		var pix1, pix2 color.RGBA
		if y1 < 0 {
			pix1 = color.NewRGBA(0, 0, 0, 0) // no_color equivalent
		} else {
			pix1 = lis.source.Pixel(x, y1)
		}

		if y2 > h {
			pix2 = color.NewRGBA(0, 0, 0, 0) // no_color equivalent
		} else {
			pix2 = lis.source.Pixel(x, y2)
		}

		// Linear interpolation
		t := srcY - float64(y1)
		return color.NewRGBA(
			pix1.R*(1-t)+pix2.R*t,
			pix1.G*(1-t)+pix2.G*t,
			pix1.B*(1-t)+pix2.B*t,
			pix1.A*(1-t)+pix2.A*t,
		)
	} else {
		// Average source pixels between y and y+1
		srcY1 := (float64(y)+0.5)*lis.scale - 0.5
		srcY2 := srcY1 + lis.scale
		h := int(lis.source.Height()) - 1
		y1 := int(math.Floor(srcY1))
		y2 := int(math.Floor(srcY2))

		c := color.NewRGBA(0, 0, 0, 0)
		weight := 0.0

		if y1 >= 0 {
			pix := lis.source.Pixel(x, y1)
			w := float64(y1+1) - srcY1
			c.R += pix.R * w
			c.G += pix.G * w
			c.B += pix.B * w
			c.A += pix.A * w
			weight += w
		}

		for y := y1 + 1; y < y2; y++ {
			if y <= h {
				pix := lis.source.Pixel(x, y)
				c.R += pix.R
				c.G += pix.G
				c.B += pix.B
				c.A += pix.A
				weight += 1.0
			}
		}

		if y2 <= h {
			pix := lis.source.Pixel(x, y2)
			w := srcY2 - float64(y2)
			c.R += pix.R * w
			c.G += pix.G * w
			c.B += pix.B * w
			c.A += pix.A * w
			weight += w
		}

		if weight > 0 {
			c.R *= lis.scaleInv
			c.G *= lis.scaleInv
			c.B *= lis.scaleInv
			c.A *= lis.scaleInv
		}

		return c
	}
}

// Filter interface defines methods needed by LineImagePattern.
type Filter interface {
	Dilation() int
	PixelHighRes(rows [][]color.RGBA, p *color.RGBA, x, y int)
}

// LineImagePattern creates a line pattern from a source image with filtering.
// This is equivalent to AGG's line_image_pattern template class.
type LineImagePattern struct {
	filter       Filter
	dilation     int
	dilationHR   int
	data         []color.RGBA
	buf          *buffer.RowPtrCache[color.RGBA]
	width        int
	height       int
	widthHR      int
	halfHeightHR int
	offsetYHR    int
}

// NewLineImagePattern creates a new line image pattern.
func NewLineImagePattern(filter Filter) *LineImagePattern {
	return &LineImagePattern{
		filter:     filter,
		dilation:   filter.Dilation() + 1,
		dilationHR: (filter.Dilation() + 1) << primitives.LineSubpixelShift,
		buf:        buffer.NewRowPtrCache[color.RGBA](),
	}
}

// NewLineImagePatternFromSource creates a pattern from a source image.
func NewLineImagePatternFromSource(filter Filter, src Source) *LineImagePattern {
	pattern := NewLineImagePattern(filter)
	pattern.Create(src)
	return pattern
}

// Create creates the pattern from a source image.
func (lip *LineImagePattern) Create(src Source) {
	lip.height = int(math.Ceil(src.Height()))
	lip.width = int(math.Ceil(src.Width()))
	lip.widthHR = int(math.Round(src.Width() * primitives.LineSubpixelScale))
	lip.halfHeightHR = int(math.Round(src.Height() * primitives.LineSubpixelScale / 2))
	lip.offsetYHR = lip.dilationHR + lip.halfHeightHR - primitives.LineSubpixelScale/2
	lip.halfHeightHR += primitives.LineSubpixelScale / 2

	totalSize := (lip.width + lip.dilation*2) * (lip.height + lip.dilation*2)
	lip.data = make([]color.RGBA, totalSize)

	lip.buf.Attach(lip.data, lip.width+lip.dilation*2, lip.height+lip.dilation*2, lip.width+lip.dilation*2)

	// Copy source data to center of buffer
	for y := 0; y < lip.height; y++ {
		row := lip.buf.RowPtr(y + lip.dilation)
		if row != nil && len(row) > lip.dilation {
			for x := 0; x < lip.width; x++ {
				if lip.dilation+x < len(row) {
					row[lip.dilation+x] = src.Pixel(x, y)
				}
			}
		}
	}

	// Fill dilation areas with transparent pixels
	for y := 0; y < lip.dilation; y++ {
		// Top and bottom dilation
		topRow := lip.buf.RowPtr(lip.dilation - y - 1)
		bottomRow := lip.buf.RowPtr(lip.dilation + lip.height + y)

		if topRow != nil && bottomRow != nil {
			for x := lip.dilation; x < lip.dilation+lip.width; x++ {
				if x < len(topRow) && x < len(bottomRow) {
					topRow[x] = color.NewRGBA(0, 0, 0, 0)    // no_color
					bottomRow[x] = color.NewRGBA(0, 0, 0, 0) // no_color
				}
			}
		}
	}

	// Fill left and right dilation
	totalHeight := lip.height + lip.dilation*2
	for y := 0; y < totalHeight; y++ {
		row := lip.buf.RowPtr(y)
		if row == nil || len(row) == 0 {
			continue
		}

		// Copy edge pixels for side dilation
		for x := 0; x < lip.dilation; x++ {
			// Left side
			if lip.dilation < len(row) {
				leftIdx := lip.dilation - x - 1
				srcIdx := lip.dilation
				if leftIdx >= 0 && leftIdx < len(row) && srcIdx < len(row) {
					row[leftIdx] = row[srcIdx]
				}
			}

			// Right side
			rightIdx := lip.dilation + lip.width + x
			srcIdx := lip.dilation + lip.width - 1
			if rightIdx < len(row) && srcIdx >= 0 && srcIdx < len(row) {
				row[rightIdx] = row[srcIdx]
			}
		}
	}
}

// PatternWidth returns the pattern width in high resolution.
func (lip *LineImagePattern) PatternWidth() int {
	return lip.widthHR
}

// LineWidth returns the line width in high resolution.
func (lip *LineImagePattern) LineWidth() int {
	return lip.halfHeightHR
}

// Width returns the pattern height as a double.
func (lip *LineImagePattern) Width() float64 {
	return float64(lip.height)
}

// Pixel gets a filtered pixel from the pattern.
func (lip *LineImagePattern) Pixel(p *color.RGBA, x, y int) {
	lip.filter.PixelHighRes(lip.buf.Rows(), p, x%lip.widthHR+lip.dilationHR, y+lip.offsetYHR)
}

// GetFilter returns the filter being used.
func (lip *LineImagePattern) GetFilter() Filter {
	return lip.filter
}

// LineImagePatternPow2 is a power-of-2 optimized version of LineImagePattern.
// This is equivalent to AGG's line_image_pattern_pow2 template class.
type LineImagePatternPow2 struct {
	*LineImagePattern
	mask int
}

// NewLineImagePatternPow2 creates a new power-of-2 optimized pattern.
func NewLineImagePatternPow2(filter Filter) *LineImagePatternPow2 {
	return &LineImagePatternPow2{
		LineImagePattern: NewLineImagePattern(filter),
		mask:             primitives.LineSubpixelMask,
	}
}

// NewLineImagePatternPow2FromSource creates a pattern from a source image.
func NewLineImagePatternPow2FromSource(filter Filter, src Source) *LineImagePatternPow2 {
	pattern := NewLineImagePatternPow2(filter)
	pattern.Create(src)
	return pattern
}

// Create creates the pattern and optimizes the mask for power-of-2.
func (lipp2 *LineImagePatternPow2) Create(src Source) {
	lipp2.LineImagePattern.Create(src)

	// Calculate power-of-2 mask
	lipp2.mask = 1
	for lipp2.mask < lipp2.width {
		lipp2.mask <<= 1
		lipp2.mask |= 1
	}
	lipp2.mask <<= primitives.LineSubpixelShift - 1
	lipp2.mask |= primitives.LineSubpixelMask
	lipp2.widthHR = lipp2.mask + 1
}

// Pixel gets a filtered pixel with power-of-2 optimization.
func (lipp2 *LineImagePatternPow2) Pixel(p *color.RGBA, x, y int) {
	lipp2.filter.PixelHighRes(lipp2.buf.Rows(), p, (x&lipp2.mask)+lipp2.dilationHR, y+lipp2.offsetYHR)
}

// DistanceInterpolator4 provides advanced distance interpolation for image-based rendering.
// This is equivalent to AGG's distance_interpolator4 class.
type DistanceInterpolator4 struct {
	dx        int
	dy        int
	dxStart   int
	dyStart   int
	dxPict    int
	dyPict    int
	dxEnd     int
	dyEnd     int
	dist      int
	distStart int
	distPict  int
	distEnd   int
	len       int
}

// NewDistanceInterpolator4 creates a new distance interpolator.
func NewDistanceInterpolator4(x1, y1, x2, y2, sx, sy, ex, ey, length int, scale float64, x, y int) *DistanceInterpolator4 {
	di := &DistanceInterpolator4{}

	di.dx = x2 - x1
	di.dy = y2 - y1
	di.dxStart = primitives.LineMR(sx) - primitives.LineMR(x1)
	di.dyStart = primitives.LineMR(sy) - primitives.LineMR(y1)
	di.dxEnd = primitives.LineMR(ex) - primitives.LineMR(x2)
	di.dyEnd = primitives.LineMR(ey) - primitives.LineMR(y2)

	di.dist = int(math.Round(float64(x+primitives.LineSubpixelScale/2-x2)*float64(di.dy) -
		float64(y+primitives.LineSubpixelScale/2-y2)*float64(di.dx)))

	di.distStart = (primitives.LineMR(x+primitives.LineSubpixelScale/2)-primitives.LineMR(sx))*di.dyStart -
		(primitives.LineMR(y+primitives.LineSubpixelScale/2)-primitives.LineMR(sy))*di.dxStart

	di.distEnd = (primitives.LineMR(x+primitives.LineSubpixelScale/2)-primitives.LineMR(ex))*di.dyEnd -
		(primitives.LineMR(y+primitives.LineSubpixelScale/2)-primitives.LineMR(ey))*di.dxEnd

	di.len = int(math.Round(float64(length) / scale))

	d := float64(length) * scale
	dx := int(math.Round((float64(x2-x1) * float64(1<<primitives.LineSubpixelShift)) / d))
	dy := int(math.Round((float64(y2-y1) * float64(1<<primitives.LineSubpixelShift)) / d))
	di.dxPict = -dy
	di.dyPict = dx
	di.distPict = ((x+primitives.LineSubpixelScale/2-(x1-dy))*di.dyPict -
		(y+primitives.LineSubpixelScale/2-(y1+dx))*di.dxPict) >> primitives.LineSubpixelShift

	di.dx <<= primitives.LineSubpixelShift
	di.dy <<= primitives.LineSubpixelShift
	di.dxStart <<= primitives.LineMRSubpixelShift
	di.dyStart <<= primitives.LineMRSubpixelShift
	di.dxEnd <<= primitives.LineMRSubpixelShift
	di.dyEnd <<= primitives.LineMRSubpixelShift

	return di
}

// IncX increments X and updates distances.
func (di *DistanceInterpolator4) IncX() {
	di.dist += di.dy
	di.distStart += di.dyStart
	di.distPict += di.dyPict
	di.distEnd += di.dyEnd
}

// DecX decrements X and updates distances.
func (di *DistanceInterpolator4) DecX() {
	di.dist -= di.dy
	di.distStart -= di.dyStart
	di.distPict -= di.dyPict
	di.distEnd -= di.dyEnd
}

// IncY increments Y and updates distances.
func (di *DistanceInterpolator4) IncY() {
	di.dist -= di.dx
	di.distStart -= di.dxStart
	di.distPict -= di.dxPict
	di.distEnd -= di.dxEnd
}

// DecY decrements Y and updates distances.
func (di *DistanceInterpolator4) DecY() {
	di.dist += di.dx
	di.distStart += di.dxStart
	di.distPict += di.dxPict
	di.distEnd += di.dxEnd
}

// IncXWithDY increments X with Y delta and updates distances.
func (di *DistanceInterpolator4) IncXWithDY(dy int) {
	di.dist += di.dy
	di.distStart += di.dyStart
	di.distPict += di.dyPict
	di.distEnd += di.dyEnd
	if dy > 0 {
		di.dist -= di.dx
		di.distStart -= di.dxStart
		di.distPict -= di.dxPict
		di.distEnd -= di.dxEnd
	}
	if dy < 0 {
		di.dist += di.dx
		di.distStart += di.dxStart
		di.distPict += di.dxPict
		di.distEnd += di.dxEnd
	}
}

// DecXWithDY decrements X with Y delta and updates distances.
func (di *DistanceInterpolator4) DecXWithDY(dy int) {
	di.dist -= di.dy
	di.distStart -= di.dyStart
	di.distPict -= di.dyPict
	di.distEnd -= di.dyEnd
	if dy > 0 {
		di.dist -= di.dx
		di.distStart -= di.dxStart
		di.distPict -= di.dxPict
		di.distEnd -= di.dxEnd
	}
	if dy < 0 {
		di.dist += di.dx
		di.distStart += di.dxStart
		di.distPict += di.dxPict
		di.distEnd += di.dxEnd
	}
}

// IncYWithDX increments Y with X delta and updates distances.
func (di *DistanceInterpolator4) IncYWithDX(dx int) {
	di.dist -= di.dx
	di.distStart -= di.dxStart
	di.distPict -= di.dxPict
	di.distEnd -= di.dxEnd
	if dx > 0 {
		di.dist += di.dy
		di.distStart += di.dyStart
		di.distPict += di.dyPict
		di.distEnd += di.dyEnd
	}
	if dx < 0 {
		di.dist -= di.dy
		di.distStart -= di.dyStart
		di.distPict -= di.dyPict
		di.distEnd -= di.dyEnd
	}
}

// DecYWithDX decrements Y with X delta and updates distances.
func (di *DistanceInterpolator4) DecYWithDX(dx int) {
	di.dist += di.dx
	di.distStart += di.dxStart
	di.distPict += di.dxPict
	di.distEnd += di.dxEnd
	if dx > 0 {
		di.dist += di.dy
		di.distStart += di.dyStart
		di.distPict += di.dyPict
		di.distEnd += di.dyEnd
	}
	if dx < 0 {
		di.dist -= di.dy
		di.distStart -= di.dyStart
		di.distPict -= di.dyPict
		di.distEnd -= di.dyEnd
	}
}

// Dist returns the current distance.
func (di *DistanceInterpolator4) Dist() int {
	return di.dist
}

// DistStart returns the start distance.
func (di *DistanceInterpolator4) DistStart() int {
	return di.distStart
}

// DistPict returns the picture distance.
func (di *DistanceInterpolator4) DistPict() int {
	return di.distPict
}

// DistEnd returns the end distance.
func (di *DistanceInterpolator4) DistEnd() int {
	return di.distEnd
}

// DX returns the X delta.
func (di *DistanceInterpolator4) DX() int {
	return di.dx
}

// DY returns the Y delta.
func (di *DistanceInterpolator4) DY() int {
	return di.dy
}

// DXStart returns the start X delta.
func (di *DistanceInterpolator4) DXStart() int {
	return di.dxStart
}

// DYStart returns the start Y delta.
func (di *DistanceInterpolator4) DYStart() int {
	return di.dyStart
}

// DXPict returns the picture X delta.
func (di *DistanceInterpolator4) DXPict() int {
	return di.dxPict
}

// DYPict returns the picture Y delta.
func (di *DistanceInterpolator4) DYPict() int {
	return di.dyPict
}

// DXEnd returns the end X delta.
func (di *DistanceInterpolator4) DXEnd() int {
	return di.dxEnd
}

// DYEnd returns the end Y delta.
func (di *DistanceInterpolator4) DYEnd() int {
	return di.dyEnd
}

// Len returns the length.
func (di *DistanceInterpolator4) Len() int {
	return di.len
}

// MaxHalfWidthImage defines the maximum half width for image interpolation.
const MaxHalfWidthImage = 64

// ImageRenderer interface defines methods required by LineInterpolatorImage.
type ImageRenderer interface {
	SubpixelWidth() int
	PatternWidth() int
	Pixel(p *color.RGBA, x, y int)
	BlendColorVSpan(x, y int, length int, colors []color.RGBA)
	BlendColorHSpan(x, y int, length int, colors []color.RGBA)
}

// LineInterpolatorImage implements line interpolation for image-based patterns.
// This is equivalent to AGG's line_interpolator_image template class.
type LineInterpolatorImage struct {
	lp        *primitives.LineParameters
	li        *primitives.Dda2LineInterpolator
	di        *DistanceInterpolator4
	ren       ImageRenderer
	x         int
	y         int
	oldX      int
	oldY      int
	count     int
	width     int
	maxExtent int
	start     int
	step      int
	distPos   [MaxHalfWidthImage + 1]int
	colors    [MaxHalfWidthImage*2 + 4]color.RGBA
}

// NewLineInterpolatorImage creates a new image line interpolator.
func NewLineInterpolatorImage(ren ImageRenderer, lp *primitives.LineParameters,
	sx, sy, ex, ey, patternStart int, scaleX float64,
) *LineInterpolatorImage {
	li := &LineInterpolatorImage{
		lp:  lp,
		ren: ren,
		x:   lp.X1 >> primitives.LineSubpixelShift,
		y:   lp.Y1 >> primitives.LineSubpixelShift,
	}

	// Create line interpolator
	var liParam int
	if lp.Vertical {
		liParam = primitives.LineDBLHR(lp.X2 - lp.X1)
	} else {
		liParam = primitives.LineDBLHR(lp.Y2 - lp.Y1)
	}

	var count int
	if lp.Vertical {
		count = int(math.Abs(float64(lp.Y2 - lp.Y1)))
	} else {
		count = int(math.Abs(float64(lp.X2-lp.X1))) + 1
	}

	li.li = primitives.NewDda2LineInterpolator(0, liParam, count)

	// Create distance interpolator
	li.di = NewDistanceInterpolator4(lp.X1, lp.Y1, lp.X2, lp.Y2, sx, sy, ex, ey, lp.Len, scaleX,
		lp.X1&^primitives.LineSubpixelMask, lp.Y1&^primitives.LineSubpixelMask)

	li.oldX = li.x
	li.oldY = li.y

	if lp.Vertical {
		li.count = int(math.Abs(float64((lp.Y2 >> primitives.LineSubpixelShift) - li.y)))
	} else {
		li.count = int(math.Abs(float64((lp.X2 >> primitives.LineSubpixelShift) - li.x)))
	}

	li.width = ren.SubpixelWidth()
	li.maxExtent = (li.width + primitives.LineSubpixelScale) >> primitives.LineSubpixelShift
	li.start = patternStart + (li.maxExtent+2)*ren.PatternWidth()
	li.step = 0

	// Build distance array
	subLi := primitives.NewDda2LineInterpolator(0, 0, lp.Len)
	if lp.Vertical {
		subLi = primitives.NewDda2LineInterpolator(0, lp.DY<<primitives.LineSubpixelShift, lp.Len)
	} else {
		subLi = primitives.NewDda2LineInterpolator(0, lp.DX<<primitives.LineSubpixelShift, lp.Len)
	}

	stop := li.width + primitives.LineSubpixelScale*2
	for i := 0; i < MaxHalfWidthImage; i++ {
		li.distPos[i] = subLi.Y()
		if li.distPos[i] >= stop {
			break
		}
		subLi.Inc()
	}
	li.distPos[MaxHalfWidthImage] = 0x7FFF0000

	// Pre-position for rendering
	li.preposition(lp)

	return li
}

// preposition prepares the interpolator for rendering.
func (lii *LineInterpolatorImage) preposition(lp *primitives.LineParameters) {
	var dist1Start, dist2Start int
	npix := 1

	if lp.Vertical {
		for {
			lii.li.DecInc()
			lii.y -= lp.Inc
			lii.x = (lii.lp.X1 + lii.li.Y()) >> primitives.LineSubpixelShift

			if lp.Inc > 0 {
				lii.di.DecYWithDX(lii.x - lii.oldX)
			} else {
				lii.di.IncYWithDX(lii.x - lii.oldX)
			}

			lii.oldX = lii.x
			dist1Start = lii.di.DistStart()
			dist2Start = dist1Start

			dx := 0
			if dist1Start < 0 {
				npix++
			}
			for {
				dist1Start += lii.di.DYStart()
				dist2Start -= lii.di.DYStart()
				if dist1Start < 0 {
					npix++
				}
				if dist2Start < 0 {
					npix++
				}
				dx++
				if lii.distPos[dx] > lii.width {
					break
				}
			}
			if npix == 0 {
				break
			}
			npix = 0
			lii.step--
			if lii.step < -lii.maxExtent {
				break
			}
		}
	} else {
		for {
			lii.li.DecInc()
			lii.x -= lp.Inc
			lii.y = (lii.lp.Y1 + lii.li.Y()) >> primitives.LineSubpixelShift

			if lp.Inc > 0 {
				lii.di.DecXWithDY(lii.y - lii.oldY)
			} else {
				lii.di.IncXWithDY(lii.y - lii.oldY)
			}

			lii.oldY = lii.y
			dist1Start = lii.di.DistStart()
			dist2Start = dist1Start

			dy := 0
			if dist1Start < 0 {
				npix++
			}
			for {
				dist1Start -= lii.di.DXStart()
				dist2Start += lii.di.DXStart()
				if dist1Start < 0 {
					npix++
				}
				if dist2Start < 0 {
					npix++
				}
				dy++
				if lii.distPos[dy] > lii.width {
					break
				}
			}
			if npix == 0 {
				break
			}
			npix = 0
			lii.step--
			if lii.step < -lii.maxExtent {
				break
			}
		}
	}
	lii.li.AdjustForward()
	lii.step -= lii.maxExtent
}

// StepHor performs horizontal stepping.
func (lii *LineInterpolatorImage) StepHor() bool {
	lii.li.Inc()
	lii.x += lii.lp.Inc
	lii.y = (lii.lp.Y1 + lii.li.Y()) >> primitives.LineSubpixelShift

	if lii.lp.Inc > 0 {
		lii.di.IncXWithDY(lii.y - lii.oldY)
	} else {
		lii.di.DecXWithDY(lii.y - lii.oldY)
	}

	lii.oldY = lii.y

	s1 := lii.di.Dist() / lii.lp.Len
	s2 := -s1

	if lii.lp.Inc < 0 {
		s1 = -s1
	}

	distStart := lii.di.DistStart()
	distPict := lii.di.DistPict() + lii.start
	distEnd := lii.di.DistEnd()

	p0Index := MaxHalfWidthImage + 2
	p1Index := p0Index

	npix := 0
	lii.colors[p1Index] = color.NewRGBA(0, 0, 0, 0) // clear
	if distEnd > 0 {
		if distStart <= 0 {
			lii.ren.Pixel(&lii.colors[p1Index], distPict, s2)
		}
		npix++
	}
	p1Index++

	dy := 1
	for {
		dist := lii.distPos[dy] - s1
		if dist > lii.width {
			break
		}
		distStart -= lii.di.DXStart()
		distPict -= lii.di.DXPict()
		distEnd -= lii.di.DXEnd()
		lii.colors[p1Index] = color.NewRGBA(0, 0, 0, 0) // clear
		if distEnd > 0 && distStart <= 0 {
			if lii.lp.Inc > 0 {
				dist = -dist
			}
			lii.ren.Pixel(&lii.colors[p1Index], distPict, s2-dist)
			npix++
		}
		p1Index++
		dy++
	}

	dy = 1
	distStart = lii.di.DistStart()
	distPict = lii.di.DistPict() + lii.start
	distEnd = lii.di.DistEnd()
	for {
		dist := lii.distPos[dy] + s1
		if dist > lii.width {
			break
		}
		distStart += lii.di.DXStart()
		distPict += lii.di.DXPict()
		distEnd += lii.di.DXEnd()
		p0Index--
		lii.colors[p0Index] = color.NewRGBA(0, 0, 0, 0) // clear
		if distEnd > 0 && distStart <= 0 {
			if lii.lp.Inc > 0 {
				dist = -dist
			}
			lii.ren.Pixel(&lii.colors[p0Index], distPict, s2+dist)
			npix++
		}
		dy++
	}

	spanLength := p1Index - p0Index
	if spanLength > 0 {
		lii.ren.BlendColorVSpan(lii.x, lii.y-dy+1, spanLength, lii.colors[p0Index:p1Index])
	}

	lii.step++
	return npix > 0 && lii.step < lii.count
}

// StepVer performs vertical stepping.
func (lii *LineInterpolatorImage) StepVer() bool {
	lii.li.Inc()
	lii.y += lii.lp.Inc
	lii.x = (lii.lp.X1 + lii.li.Y()) >> primitives.LineSubpixelShift

	if lii.lp.Inc > 0 {
		lii.di.IncYWithDX(lii.x - lii.oldX)
	} else {
		lii.di.DecYWithDX(lii.x - lii.oldX)
	}

	lii.oldX = lii.x

	s1 := lii.di.Dist() / lii.lp.Len
	s2 := -s1

	if lii.lp.Inc > 0 {
		s1 = -s1
	}

	distStart := lii.di.DistStart()
	distPict := lii.di.DistPict() + lii.start
	distEnd := lii.di.DistEnd()

	p0Index := MaxHalfWidthImage + 2
	p1Index := p0Index

	npix := 0
	lii.colors[p1Index] = color.NewRGBA(0, 0, 0, 0) // clear
	if distEnd > 0 {
		if distStart <= 0 {
			lii.ren.Pixel(&lii.colors[p1Index], distPict, s2)
		}
		npix++
	}
	p1Index++

	dx := 1
	for {
		dist := lii.distPos[dx] - s1
		if dist > lii.width {
			break
		}
		distStart += lii.di.DYStart()
		distPict += lii.di.DYPict()
		distEnd += lii.di.DYEnd()
		lii.colors[p1Index] = color.NewRGBA(0, 0, 0, 0) // clear
		if distEnd > 0 && distStart <= 0 {
			if lii.lp.Inc > 0 {
				dist = -dist
			}
			lii.ren.Pixel(&lii.colors[p1Index], distPict, s2+dist)
			npix++
		}
		p1Index++
		dx++
	}

	dx = 1
	distStart = lii.di.DistStart()
	distPict = lii.di.DistPict() + lii.start
	distEnd = lii.di.DistEnd()
	for {
		dist := lii.distPos[dx] + s1
		if dist > lii.width {
			break
		}
		distStart -= lii.di.DYStart()
		distPict -= lii.di.DYPict()
		distEnd -= lii.di.DYEnd()
		p0Index--
		lii.colors[p0Index] = color.NewRGBA(0, 0, 0, 0) // clear
		if distEnd > 0 && distStart <= 0 {
			if lii.lp.Inc > 0 {
				dist = -dist
			}
			lii.ren.Pixel(&lii.colors[p0Index], distPict, s2-dist)
			npix++
		}
		dx++
	}

	spanLength := p1Index - p0Index
	if spanLength > 0 {
		lii.ren.BlendColorHSpan(lii.x-dx+1, lii.y, spanLength, lii.colors[p0Index:p1Index])
	}

	lii.step++
	return npix > 0 && lii.step < lii.count
}

// PatternEnd returns the pattern end position.
func (lii *LineInterpolatorImage) PatternEnd() int {
	return lii.start + lii.di.Len()
}

// Vertical returns true if the line is vertical.
func (lii *LineInterpolatorImage) Vertical() bool {
	return lii.lp.Vertical
}

// Width returns the line width.
func (lii *LineInterpolatorImage) Width() int {
	return lii.width
}

// Count returns the step count.
func (lii *LineInterpolatorImage) Count() int {
	return lii.count
}

// Pattern interface defines methods for image patterns.
type Pattern interface {
	LineWidth() int
	PatternWidth() int
	Pixel(p *color.RGBA, x, y int)
}

// BaseRenderer interface defines methods required by RendererOutlineImage.
type BaseRenderer interface {
	BlendColorHSpan(x, y int, length int, colors []color.RGBA, covers []basics.CoverType)
	BlendColorVSpan(x, y int, length int, colors []color.RGBA, covers []basics.CoverType)
}

// RendererOutlineImage renders image-based outlines.
// This is equivalent to AGG's renderer_outline_image template class.
type RendererOutlineImage struct {
	ren      BaseRenderer
	pattern  Pattern
	start    int
	scaleX   float64
	clipBox  basics.RectI
	clipping bool
}

// NewRendererOutlineImage creates a new image outline renderer.
func NewRendererOutlineImage(ren BaseRenderer, pattern Pattern) *RendererOutlineImage {
	return &RendererOutlineImage{
		ren:      ren,
		pattern:  pattern,
		start:    0,
		scaleX:   1.0,
		clipBox:  basics.RectI{X1: 0, Y1: 0, X2: 0, Y2: 0},
		clipping: false,
	}
}

// Attach attaches a new base renderer.
func (roi *RendererOutlineImage) Attach(ren BaseRenderer) {
	roi.ren = ren
}

// SetPattern sets the image pattern.
func (roi *RendererOutlineImage) SetPattern(pattern Pattern) {
	roi.pattern = pattern
}

// GetPattern returns the current pattern.
func (roi *RendererOutlineImage) GetPattern() Pattern {
	return roi.pattern
}

// ResetClipping disables clipping.
func (roi *RendererOutlineImage) ResetClipping() {
	roi.clipping = false
}

// ClipBox sets the clipping rectangle.
func (roi *RendererOutlineImage) ClipBox(x1, y1, x2, y2 float64) {
	roi.clipBox.X1 = primitives.LineCoordSatConv(x1)
	roi.clipBox.Y1 = primitives.LineCoordSatConv(y1)
	roi.clipBox.X2 = primitives.LineCoordSatConv(x2)
	roi.clipBox.Y2 = primitives.LineCoordSatConv(y2)
	roi.clipping = true
}

// SetScaleX sets the X scale factor.
func (roi *RendererOutlineImage) SetScaleX(s float64) {
	roi.scaleX = s
}

// ScaleX returns the current X scale factor.
func (roi *RendererOutlineImage) ScaleX() float64 {
	return roi.scaleX
}

// SetStartX sets the starting X position.
func (roi *RendererOutlineImage) SetStartX(s float64) {
	roi.start = int(math.Round(s * primitives.LineSubpixelScale))
}

// StartX returns the starting X position.
func (roi *RendererOutlineImage) StartX() float64 {
	return float64(roi.start) / primitives.LineSubpixelScale
}

// SubpixelWidth returns the subpixel line width.
func (roi *RendererOutlineImage) SubpixelWidth() int {
	return roi.pattern.LineWidth()
}

// PatternWidth returns the pattern width.
func (roi *RendererOutlineImage) PatternWidth() int {
	return roi.pattern.PatternWidth()
}

// Width returns the line width.
func (roi *RendererOutlineImage) Width() float64 {
	return float64(roi.SubpixelWidth()) / primitives.LineSubpixelScale
}

// Pixel gets a pixel from the pattern.
func (roi *RendererOutlineImage) Pixel(p *color.RGBA, x, y int) {
	roi.pattern.Pixel(p, x, y)
}

// BlendColorHSpan blends a horizontal color span.
func (roi *RendererOutlineImage) BlendColorHSpan(x, y int, length int, colors []color.RGBA) {
	roi.ren.BlendColorHSpan(x, y, length, colors, nil)
}

// BlendColorVSpan blends a vertical color span.
func (roi *RendererOutlineImage) BlendColorVSpan(x, y int, length int, colors []color.RGBA) {
	roi.ren.BlendColorVSpan(x, y, length, colors, nil)
}

// AccurateJoinOnly returns true for accurate joins.
func (roi *RendererOutlineImage) AccurateJoinOnly() bool {
	return true
}

// Semidot draws a semidot (not implemented for image renderer).
func (roi *RendererOutlineImage) Semidot(cmp func(int) bool, xc, yc, xp, yp int) {
	// Not implemented for image renderer
}

// Pie draws a pie segment (not implemented for image renderer).
func (roi *RendererOutlineImage) Pie(xc, yc, x1, y1, x2, y2 int) {
	// Not implemented for image renderer
}

// Line0 draws a 0-width line (not implemented for image renderer).
func (roi *RendererOutlineImage) Line0(lp *primitives.LineParameters) {
	// Not implemented for image renderer
}

// Line1 draws a 1-pixel line (not implemented for image renderer).
func (roi *RendererOutlineImage) Line1(lp *primitives.LineParameters, sx, sy int) {
	// Not implemented for image renderer
}

// Line2 draws a 2-pixel line (not implemented for image renderer).
func (roi *RendererOutlineImage) Line2(lp *primitives.LineParameters, ex, ey int) {
	// Not implemented for image renderer
}

// Line3NoClip draws a line without clipping.
func (roi *RendererOutlineImage) Line3NoClip(lp *primitives.LineParameters, sx, sy, ex, ey int) {
	if lp.Len > primitives.LineMaxLength {
		lp1, lp2 := lp.Divide()
		mx := lp1.X2 + (lp1.Y2 - lp1.Y1)
		my := lp1.Y2 - (lp1.X2 - lp1.X1)
		roi.Line3NoClip(&lp1, (lp.X1+sx)>>1, (lp.Y1+sy)>>1, mx, my)
		roi.Line3NoClip(&lp2, mx, my, (lp.X2+ex)>>1, (lp.Y2+ey)>>1)
		return
	}

	// Fix degenerate bisectrix
	primitives.FixDegenerateBisectrixStart(lp, &sx, &sy)
	primitives.FixDegenerateBisectrixEnd(lp, &ex, &ey)

	li := NewLineInterpolatorImage(roi, lp, sx, sy, ex, ey, roi.start, roi.scaleX)
	if li.Vertical() {
		for li.StepVer() {
		}
	} else {
		for li.StepHor() {
		}
	}
	roi.start += int(math.Round(float64(lp.Len) / roi.scaleX))
}

// Line3 draws a line with optional clipping.
func (roi *RendererOutlineImage) Line3(lp *primitives.LineParameters, sx, sy, ex, ey int) {
	if roi.clipping {
		x1, y1, x2, y2 := lp.X1, lp.Y1, lp.X2, lp.Y2
		flags := basics.ClipLineSegment(&x1, &y1, &x2, &y2, roi.clipBox)
		start := roi.start

		if (flags & 4) == 0 {
			if flags != 0 {
				lp2 := primitives.NewLineParameters(x1, y1, x2, y2,
					int(math.Round(basics.CalcDistance(float64(x1), float64(y1), float64(x2), float64(y2)))))

				if (flags & 1) != 0 {
					roi.start += int(math.Round(basics.CalcDistance(float64(lp.X1), float64(lp.Y1), float64(x1), float64(y1)) / roi.scaleX))
					sx = x1 + (y2 - y1)
					sy = y1 - (x2 - x1)
				} else {
					for int(math.Abs(float64(sx-lp.X1)))+int(math.Abs(float64(sy-lp.Y1))) > lp2.Len {
						sx = (lp.X1 + sx) >> 1
						sy = (lp.Y1 + sy) >> 1
					}
				}

				if (flags & 2) != 0 {
					ex = x2 + (y2 - y1)
					ey = y2 - (x2 - x1)
				} else {
					for int(math.Abs(float64(ex-lp.X2)))+int(math.Abs(float64(ey-lp.Y2))) > lp2.Len {
						ex = (lp.X2 + ex) >> 1
						ey = (lp.Y2 + ey) >> 1
					}
				}
				roi.Line3NoClip(&lp2, sx, sy, ex, ey)
			} else {
				roi.Line3NoClip(lp, sx, sy, ex, ey)
			}
		}
		roi.start = start + int(math.Round(float64(lp.Len)/roi.scaleX))
	} else {
		roi.Line3NoClip(lp, sx, sy, ex, ey)
	}
}
