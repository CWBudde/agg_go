package image

import (
	"math"

	"agg_go/internal/array"
	"agg_go/internal/basics"
)

// Image filter scale constants
const (
	ImageFilterShift = 14                    // Filter precision shift
	ImageFilterScale = 1 << ImageFilterShift // Filter scale factor
	ImageFilterMask  = ImageFilterScale - 1  // Filter mask
)

// Image subpixel scale constants
const (
	ImageSubpixelShift = 8                       // Subpixel precision shift
	ImageSubpixelScale = 1 << ImageSubpixelShift // Subpixel scale factor
	ImageSubpixelMask  = ImageSubpixelScale - 1  // Subpixel mask
)

// FilterFunction interface for filter weight calculation
type FilterFunction interface {
	Radius() float64
	CalcWeight(x float64) float64
}

// ImageFilterLUT represents a lookup table for image filtering weights
type ImageFilterLUT struct {
	radius      float64
	diameter    int
	start       int
	weightArray *array.PodArray[int16]
}

// NewImageFilterLUT creates a new image filter lookup table
func NewImageFilterLUT() *ImageFilterLUT {
	return &ImageFilterLUT{
		radius:      0,
		diameter:    0,
		start:       0,
		weightArray: array.NewPodArray[int16](),
	}
}

// NewImageFilterLUTWithFilter creates a new image filter LUT with the given filter
func NewImageFilterLUTWithFilter(filter FilterFunction, normalization bool) *ImageFilterLUT {
	lut := NewImageFilterLUT()
	lut.Calculate(filter, normalization)
	return lut
}

// Calculate computes the filter weights for the given filter function
func (lut *ImageFilterLUT) Calculate(filter FilterFunction, normalization bool) {
	r := filter.Radius()
	lut.reallocLUT(r)

	pivot := lut.diameter << (ImageSubpixelShift - 1)
	for i := 0; i < pivot; i++ {
		x := float64(i) / float64(ImageSubpixelScale)
		y := filter.CalcWeight(x)
		weight := int16(basics.IRound(y * ImageFilterScale))
		lut.weightArray.Set(pivot+i, weight)
		lut.weightArray.Set(pivot-i, weight)
	}

	end := (lut.diameter << ImageSubpixelShift) - 1
	lut.weightArray.Set(0, lut.weightArray.At(end))

	if normalization {
		lut.Normalize()
	}
}

// reallocLUT reallocates the weight array for the given radius
func (lut *ImageFilterLUT) reallocLUT(radius float64) {
	lut.radius = radius
	lut.diameter = int(math.Ceil(radius)) * 2
	lut.start = -int(lut.diameter/2 - 1)
	size := lut.diameter << ImageSubpixelShift
	if size > lut.weightArray.Size() {
		lut.weightArray.Resize(size)
	}
}

// Normalize normalizes the integer filter weights to ensure sum equals 1.0
func (lut *ImageFilterLUT) Normalize() {
	flip := 1

	for i := 0; i < ImageSubpixelScale; i++ {
		for {
			sum := 0
			for j := 0; j < lut.diameter; j++ {
				sum += int(lut.weightArray.At(j*ImageSubpixelScale + i))
			}

			if sum == ImageFilterScale {
				break
			}

			k := float64(ImageFilterScale) / float64(sum)
			sum = 0
			for j := 0; j < lut.diameter; j++ {
				idx := j*ImageSubpixelScale + i
				oldWeight := int(lut.weightArray.At(idx))
				newWeight := basics.IRound(float64(oldWeight) * k)
				lut.weightArray.Set(idx, int16(newWeight))
				sum += newWeight
			}

			sum -= ImageFilterScale
			inc := 1
			if sum > 0 {
				inc = -1
			}

			for j := 0; j < lut.diameter && sum != 0; j++ {
				flip ^= 1
				idx := lut.diameter/2 + j/2
				if flip == 0 {
					idx = lut.diameter/2 - j/2
				}
				weightIdx := idx*ImageSubpixelScale + i
				v := int(lut.weightArray.At(weightIdx))
				if v < ImageFilterScale {
					lut.weightArray.Set(weightIdx, int16(v+inc))
					sum += inc
				}
			}
		}
	}

	pivot := lut.diameter << (ImageSubpixelShift - 1)
	for i := 0; i < pivot; i++ {
		lut.weightArray.Set(pivot+i, lut.weightArray.At(pivot-i))
	}
	end := (lut.diameter << ImageSubpixelShift) - 1
	lut.weightArray.Set(0, lut.weightArray.At(end))
}

// Radius returns the filter radius
func (lut *ImageFilterLUT) Radius() float64 {
	return lut.radius
}

// Diameter returns the filter diameter
func (lut *ImageFilterLUT) Diameter() int {
	return lut.diameter
}

// Start returns the filter start position
func (lut *ImageFilterLUT) Start() int {
	return lut.start
}

// WeightArray returns the weight array data
func (lut *ImageFilterLUT) WeightArray() []int16 {
	return lut.weightArray.Data()
}

// ImageFilter is a generic filter that holds a filter function and LUT
type ImageFilter[F FilterFunction] struct {
	*ImageFilterLUT
	filterFunction F
}

// NewImageFilter creates a new image filter with the given filter function
func NewImageFilter[F FilterFunction](filter F) *ImageFilter[F] {
	imgFilter := &ImageFilter[F]{
		ImageFilterLUT: NewImageFilterLUT(),
		filterFunction: filter,
	}
	imgFilter.Calculate(filter, true)
	return imgFilter
}

// Filter Functions Implementation

// BilinearFilter implements bilinear interpolation filter
type BilinearFilter struct{}

func (BilinearFilter) Radius() float64              { return 1.0 }
func (BilinearFilter) CalcWeight(x float64) float64 { return 1.0 - x }

// HanningFilter implements Hanning window filter
type HanningFilter struct{}

func (HanningFilter) Radius() float64 { return 1.0 }
func (HanningFilter) CalcWeight(x float64) float64 {
	return 0.5 + 0.5*math.Cos(basics.Pi*x)
}

// HammingFilter implements Hamming window filter
type HammingFilter struct{}

func (HammingFilter) Radius() float64 { return 1.0 }
func (HammingFilter) CalcWeight(x float64) float64 {
	return 0.54 + 0.46*math.Cos(basics.Pi*x)
}

// HermiteFilter implements Hermite interpolation filter
type HermiteFilter struct{}

func (HermiteFilter) Radius() float64 { return 1.0 }
func (HermiteFilter) CalcWeight(x float64) float64 {
	return (2.0*x-3.0)*x*x + 1.0
}

// QuadricFilter implements quadric B-spline filter
type QuadricFilter struct{}

func (QuadricFilter) Radius() float64 { return 1.5 }
func (QuadricFilter) CalcWeight(x float64) float64 {
	if x < 0.5 {
		return 0.75 - x*x
	}
	if x < 1.5 {
		t := x - 1.5
		return 0.5 * t * t
	}
	return 0.0
}

// BicubicFilter implements bicubic interpolation filter
type BicubicFilter struct{}

func (BicubicFilter) Radius() float64 { return 2.0 }
func (BicubicFilter) CalcWeight(x float64) float64 {
	pow3 := func(x float64) float64 {
		if x <= 0.0 {
			return 0.0
		}
		return x * x * x
	}

	return (1.0 / 6.0) * (pow3(x+2) - 4*pow3(x+1) + 6*pow3(x) - 4*pow3(x-1))
}

// KaiserFilter implements Kaiser window filter
type KaiserFilter struct {
	a       float64
	i0a     float64
	epsilon float64
}

// NewKaiserFilter creates a new Kaiser filter with the given beta parameter
func NewKaiserFilter(b float64) *KaiserFilter {
	if b == 0 {
		b = 6.33
	}
	filter := &KaiserFilter{
		a:       b,
		epsilon: 1e-12,
	}
	filter.i0a = 1.0 / filter.besselI0(b)
	return filter
}

func (k *KaiserFilter) Radius() float64 { return 1.0 }
func (k *KaiserFilter) CalcWeight(x float64) float64 {
	return k.besselI0(k.a*math.Sqrt(1.0-x*x)) * k.i0a
}

func (k *KaiserFilter) besselI0(x float64) float64 {
	sum := 1.0
	y := x * x / 4.0
	t := y

	for i := 2; t > k.epsilon; i++ {
		sum += t
		t *= y / (float64(i) * float64(i))
	}
	return sum
}

// CatromFilter implements Catmull-Rom spline filter
type CatromFilter struct{}

func (CatromFilter) Radius() float64 { return 2.0 }
func (CatromFilter) CalcWeight(x float64) float64 {
	if x < 1.0 {
		return 0.5 * (2.0 + x*x*(-5.0+x*3.0))
	}
	if x < 2.0 {
		return 0.5 * (4.0 + x*(-8.0+x*(5.0-x)))
	}
	return 0.0
}

// MitchellFilter implements Mitchell-Netravali filter
type MitchellFilter struct {
	p0, p2, p3     float64
	q0, q1, q2, q3 float64
}

// NewMitchellFilter creates a new Mitchell filter with the given B and C parameters
func NewMitchellFilter(b, c float64) *MitchellFilter {
	if b == 0 && c == 0 {
		b = 1.0 / 3.0
		c = 1.0 / 3.0
	}
	return &MitchellFilter{
		p0: (6.0 - 2.0*b) / 6.0,
		p2: (-18.0 + 12.0*b + 6.0*c) / 6.0,
		p3: (12.0 - 9.0*b - 6.0*c) / 6.0,
		q0: (8.0*b + 24.0*c) / 6.0,
		q1: (-12.0*b - 48.0*c) / 6.0,
		q2: (6.0*b + 30.0*c) / 6.0,
		q3: (-b - 6.0*c) / 6.0,
	}
}

func (m *MitchellFilter) Radius() float64 { return 2.0 }
func (m *MitchellFilter) CalcWeight(x float64) float64 {
	if x < 1.0 {
		return m.p0 + x*x*(m.p2+x*m.p3)
	}
	if x < 2.0 {
		return m.q0 + x*(m.q1+x*(m.q2+x*m.q3))
	}
	return 0.0
}

// Spline16Filter implements 16-point spline filter
type Spline16Filter struct{}

func (Spline16Filter) Radius() float64 { return 2.0 }
func (Spline16Filter) CalcWeight(x float64) float64 {
	if x < 1.0 {
		return ((x-9.0/5.0)*x-1.0/5.0)*x + 1.0
	}
	return ((-1.0/3.0*(x-1)+4.0/5.0)*(x-1) - 7.0/15.0) * (x - 1)
}

// Spline36Filter implements 36-point spline filter
type Spline36Filter struct{}

func (Spline36Filter) Radius() float64 { return 3.0 }
func (Spline36Filter) CalcWeight(x float64) float64 {
	if x < 1.0 {
		return ((13.0/11.0*x-453.0/209.0)*x-3.0/209.0)*x + 1.0
	}
	if x < 2.0 {
		return ((-6.0/11.0*(x-1)+270.0/209.0)*(x-1) - 156.0/209.0) * (x - 1)
	}
	return ((1.0/11.0*(x-2)-45.0/209.0)*(x-2) + 26.0/209.0) * (x - 2)
}

// GaussianFilter implements Gaussian filter
type GaussianFilter struct{}

func (GaussianFilter) Radius() float64 { return 2.0 }
func (GaussianFilter) CalcWeight(x float64) float64 {
	return math.Exp(-2.0*x*x) * math.Sqrt(2.0/basics.Pi)
}

// BesselFilter implements Bessel filter
type BesselFilter struct{}

func (BesselFilter) Radius() float64 { return 3.2383 }
func (BesselFilter) CalcWeight(x float64) float64 {
	if x == 0.0 {
		return basics.Pi / 4.0
	}
	return basics.BesselJ(basics.Pi*x, 1) / (2.0 * x)
}

// SincFilter implements windowed sinc filter
type SincFilter struct {
	radius float64
}

// NewSincFilter creates a new sinc filter with the given radius
func NewSincFilter(r float64) *SincFilter {
	if r < 2.0 {
		r = 2.0
	}
	return &SincFilter{radius: r}
}

func (s *SincFilter) Radius() float64 { return s.radius }
func (s *SincFilter) CalcWeight(x float64) float64 {
	if x == 0.0 {
		return 1.0
	}
	x *= basics.Pi
	return math.Sin(x) / x
}

// LanczosFilter implements Lanczos windowed sinc filter
type LanczosFilter struct {
	radius float64
}

// NewLanczosFilter creates a new Lanczos filter with the given radius
func NewLanczosFilter(r float64) *LanczosFilter {
	if r < 2.0 {
		r = 2.0
	}
	return &LanczosFilter{radius: r}
}

func (l *LanczosFilter) Radius() float64 { return l.radius }
func (l *LanczosFilter) CalcWeight(x float64) float64 {
	if x == 0.0 {
		return 1.0
	}
	if x > l.radius {
		return 0.0
	}
	x *= basics.Pi
	xr := x / l.radius
	return (math.Sin(x) / x) * (math.Sin(xr) / xr)
}

// BlackmanFilter implements Blackman windowed sinc filter
type BlackmanFilter struct {
	radius float64
}

// NewBlackmanFilter creates a new Blackman filter with the given radius
func NewBlackmanFilter(r float64) *BlackmanFilter {
	if r < 2.0 {
		r = 2.0
	}
	return &BlackmanFilter{radius: r}
}

func (b *BlackmanFilter) Radius() float64 { return b.radius }
func (b *BlackmanFilter) CalcWeight(x float64) float64 {
	if x == 0.0 {
		return 1.0
	}
	if x > b.radius {
		return 0.0
	}
	x *= basics.Pi
	xr := x / b.radius
	return (math.Sin(x) / x) * (0.42 + 0.5*math.Cos(xr) + 0.08*math.Cos(2*xr))
}

// Pre-defined filter variants
type Sinc36Filter struct{ *SincFilter }
type Sinc64Filter struct{ *SincFilter }
type Sinc100Filter struct{ *SincFilter }
type Sinc144Filter struct{ *SincFilter }
type Sinc196Filter struct{ *SincFilter }
type Sinc256Filter struct{ *SincFilter }

func NewSinc36Filter() *Sinc36Filter   { return &Sinc36Filter{NewSincFilter(3.0)} }
func NewSinc64Filter() *Sinc64Filter   { return &Sinc64Filter{NewSincFilter(4.0)} }
func NewSinc100Filter() *Sinc100Filter { return &Sinc100Filter{NewSincFilter(5.0)} }
func NewSinc144Filter() *Sinc144Filter { return &Sinc144Filter{NewSincFilter(6.0)} }
func NewSinc196Filter() *Sinc196Filter { return &Sinc196Filter{NewSincFilter(7.0)} }
func NewSinc256Filter() *Sinc256Filter { return &Sinc256Filter{NewSincFilter(8.0)} }

type Lanczos36Filter struct{ *LanczosFilter }
type Lanczos64Filter struct{ *LanczosFilter }
type Lanczos100Filter struct{ *LanczosFilter }
type Lanczos144Filter struct{ *LanczosFilter }
type Lanczos196Filter struct{ *LanczosFilter }
type Lanczos256Filter struct{ *LanczosFilter }

func NewLanczos36Filter() *Lanczos36Filter   { return &Lanczos36Filter{NewLanczosFilter(3.0)} }
func NewLanczos64Filter() *Lanczos64Filter   { return &Lanczos64Filter{NewLanczosFilter(4.0)} }
func NewLanczos100Filter() *Lanczos100Filter { return &Lanczos100Filter{NewLanczosFilter(5.0)} }
func NewLanczos144Filter() *Lanczos144Filter { return &Lanczos144Filter{NewLanczosFilter(6.0)} }
func NewLanczos196Filter() *Lanczos196Filter { return &Lanczos196Filter{NewLanczosFilter(7.0)} }
func NewLanczos256Filter() *Lanczos256Filter { return &Lanczos256Filter{NewLanczosFilter(8.0)} }

type Blackman36Filter struct{ *BlackmanFilter }
type Blackman64Filter struct{ *BlackmanFilter }
type Blackman100Filter struct{ *BlackmanFilter }
type Blackman144Filter struct{ *BlackmanFilter }
type Blackman196Filter struct{ *BlackmanFilter }
type Blackman256Filter struct{ *BlackmanFilter }

func NewBlackman36Filter() *Blackman36Filter   { return &Blackman36Filter{NewBlackmanFilter(3.0)} }
func NewBlackman64Filter() *Blackman64Filter   { return &Blackman64Filter{NewBlackmanFilter(4.0)} }
func NewBlackman100Filter() *Blackman100Filter { return &Blackman100Filter{NewBlackmanFilter(5.0)} }
func NewBlackman144Filter() *Blackman144Filter { return &Blackman144Filter{NewBlackmanFilter(6.0)} }
func NewBlackman196Filter() *Blackman196Filter { return &Blackman196Filter{NewBlackmanFilter(7.0)} }
func NewBlackman256Filter() *Blackman256Filter { return &Blackman256Filter{NewBlackmanFilter(8.0)} }
