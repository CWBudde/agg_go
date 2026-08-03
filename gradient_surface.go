package agg

import (
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/span"
	"github.com/cwbudde/agg_go/internal/transform"
)

// GradientShape identifies a full-surface AGG gradient distance function.
type GradientShape uint8

const (
	GradientShapeLinear GradientShape = iota
	GradientShapeRadial
	GradientShapeAngular
	GradientShapeDiamond
	GradientShapeReflected
)

// GradientLUT is an immutable, caller-sized RGBA gradient lookup table.
type GradientLUT struct {
	colors []Color
}

// NewGradientLUT builds a stable, arbitrary-stop lookup table. Positions are
// clamped to [0,1]; at duplicate positions the last stop is authoritative.
func NewGradientLUT(stops []GradientStop, size int) (*GradientLUT, error) {
	if size < 2 {
		return nil, fmt.Errorf("agg: gradient LUT size must be at least 2, got %d", size)
	}
	if len(stops) == 0 {
		return nil, errors.New("agg: gradient LUT requires at least one stop")
	}
	normalized := append([]GradientStop(nil), stops...)
	for i := range normalized {
		if math.IsNaN(normalized[i].Position) || math.IsInf(normalized[i].Position, 0) {
			return nil, fmt.Errorf("agg: gradient stop %d has non-finite position", i)
		}
		normalized[i].Position = math.Max(0, math.Min(1, normalized[i].Position))
	}
	sort.SliceStable(normalized, func(i, j int) bool { return normalized[i].Position < normalized[j].Position })

	colors := make([]Color, size)
	if len(normalized) == 1 {
		for i := range colors {
			colors[i] = normalized[0].Color
		}
		return &GradientLUT{colors: colors}, nil
	}
	for i := range colors {
		position := float64(i) / float64(size-1)
		if position < normalized[0].Position {
			colors[i] = normalized[0].Color
			continue
		}
		lo := sort.Search(len(normalized), func(index int) bool { return normalized[index].Position > position }) - 1
		if lo < 0 {
			lo = 0
		}
		if lo >= len(normalized)-1 {
			colors[i] = normalized[len(normalized)-1].Color
			continue
		}
		hi := lo + 1
		segment := normalized[hi].Position - normalized[lo].Position
		fraction := 0.0
		if segment > 0 {
			fraction = (position - normalized[lo].Position) / segment
		}
		colors[i] = normalized[lo].Color.Gradient(normalized[hi].Color, fraction)
	}
	return &GradientLUT{colors: colors}, nil
}

// Len returns the number of entries in the lookup table.
func (lut *GradientLUT) Len() int {
	if lut == nil {
		return 0
	}
	return len(lut.colors)
}

// At returns one LUT entry and panics for an out-of-range index, like a slice.
func (lut *GradientLUT) At(index int) Color { return lut.colors[index] }

// GradientRenderOptions configures RenderGradient.
type GradientRenderOptions struct {
	Shape   GradientShape
	Start   Point
	End     Point
	Reverse bool
	Dither  bool
	Clip    *Rect
}

// RenderGradient overwrites RGBA pixels in dst through AGG span-gradient
// generators. Start is the origin for every shape. End sets the axis and
// distance for linear, radial, diamond, and reflected gradients; for angular
// gradients its direction is the zero-angle ray and its distance sets the
// period scale. Clip is half-open in destination coordinates. Destination
// pixels and row padding outside Clip are left unchanged.
func RenderGradient(dst *Image, lut *GradientLUT, opts GradientRenderOptions) error {
	if err := validateCompositeImage("gradient destination", dst); err != nil {
		return err
	}
	if lut == nil || len(lut.colors) < 2 {
		return errors.New("agg: gradient LUT is nil or too small")
	}
	if opts.Shape > GradientShapeReflected {
		return fmt.Errorf("agg: invalid gradient shape %d", opts.Shape)
	}
	if !finitePoint(opts.Start) || !finitePoint(opts.End) {
		return errors.New("agg: gradient endpoints must be finite")
	}
	clip := Rect{X1: 0, Y1: 0, X2: dst.Width(), Y2: dst.Height()}
	if opts.Clip != nil {
		clip.X1 = max(clip.X1, opts.Clip.X1)
		clip.Y1 = max(clip.Y1, opts.Clip.Y1)
		clip.X2 = min(clip.X2, opts.Clip.X2)
		clip.Y2 = min(clip.Y2, opts.Clip.Y2)
	}
	if clip.X1 >= clip.X2 || clip.Y1 >= clip.Y2 {
		return nil
	}
	table := make([]color.RGBA8[color.Linear], len(lut.colors))
	for i := range lut.colors {
		index := i
		if opts.Reverse {
			index = len(lut.colors) - 1 - i
		}
		entry := lut.colors[index]
		table[i] = color.RGBA8[color.Linear]{R: entry.R, G: entry.G, B: entry.B, A: entry.A}
	}
	return renderGradientSpans(dst, table, opts, clip)
}

func finitePoint(point Point) bool {
	return !math.IsNaN(point.X) && !math.IsInf(point.X, 0) && !math.IsNaN(point.Y) && !math.IsInf(point.Y, 0)
}

func renderGradientSpans(dst *Image, table []color.RGBA8[color.Linear], opts GradientRenderOptions, clip Rect) error {
	dx, dy := opts.End.X-opts.Start.X, opts.End.Y-opts.Start.Y
	length := math.Hypot(dx, dy)
	if length < 1 {
		length = 1
	}
	matrix := transform.NewTransAffine()
	angle := math.Atan2(dy, dx)
	if opts.Shape == GradientShapeDiamond {
		angle += math.Pi / 4
	}
	matrix.Rotate(angle)
	// Span interpolators sample destination pixel centres. Offset the gradient
	// origin by half a pixel so public Start/End retain pixel-coordinate
	// semantics (Start maps exactly to LUT index zero).
	matrix.Translate(opts.Start.X+0.5, opts.Start.Y+0.5)
	matrix.Invert()
	interpolator := span.NewSpanInterpolatorLinearDefault(matrix)
	colors := span.NewGradientPrebuiltColorRGBA8[color.Linear](table)
	d2 := length
	if opts.Shape == GradientShapeDiamond {
		d2 /= math.Sqrt2
	}
	line := make([]color.RGBA8[color.Linear], clip.Width())
	switch opts.Shape {
	case GradientShapeLinear:
		generator := span.NewSpanGradientWithSubpixelShift(interpolator, span.GradientLinearX{}, colors, 0, d2, interpolator.SubpixelShift())
		writeGradientRows(dst, generator, line, clip, opts.Dither)
	case GradientShapeRadial:
		generator := span.NewSpanGradientWithSubpixelShift(interpolator, span.GradientRadialDouble{}, colors, 0, d2, interpolator.SubpixelShift())
		writeGradientRows(dst, generator, line, clip, opts.Dither)
	case GradientShapeAngular:
		generator := span.NewSpanGradientWithSubpixelShift(interpolator, span.GradientAngular{}, colors, 0, d2, interpolator.SubpixelShift())
		writeGradientRows(dst, generator, line, clip, opts.Dither)
	case GradientShapeDiamond:
		generator := span.NewSpanGradientWithSubpixelShift(interpolator, span.GradientDiamond{}, colors, 0, d2, interpolator.SubpixelShift())
		writeGradientRows(dst, generator, line, clip, opts.Dither)
	case GradientShapeReflected:
		generator := span.NewSpanGradientWithSubpixelShift(interpolator, span.GradientReflected{}, colors, 0, d2, interpolator.SubpixelShift())
		writeGradientRows(dst, generator, line, clip, opts.Dither)
	}
	return nil
}

type gradientRowGenerator interface {
	Generate(colors []color.RGBA8[color.Linear], x, y, length int)
}

func writeGradientRows(dst *Image, generator gradientRowGenerator, line []color.RGBA8[color.Linear], clip Rect, dither bool) {
	for y := clip.Y1; y < clip.Y2; y++ {
		generator.Generate(line, clip.X1, y, len(line))
		row := dst.renBuf.Row(y)
		for i, c := range line {
			x := clip.X1 + i
			r, g, b := c.R, c.G, c.B
			if dither {
				noise := uint32(x*1103515245 ^ y*12345)
				jitter := (float64((noise>>24)&7)/255 - 0.014) * 0.25
				r = ditherChannel(r, jitter)
				g = ditherChannel(g, jitter)
				b = ditherChannel(b, jitter)
			}
			offset := x * 4
			row[offset], row[offset+1], row[offset+2], row[offset+3] = r, g, b, c.A
		}
	}
}

func ditherChannel(value uint8, jitter float64) uint8 {
	adjusted := math.Max(0, math.Min(1, float64(value)/255+jitter))
	return uint8(math.Round(adjusted * 255))
}
