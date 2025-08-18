package curves

import (
	"agg_go/internal/array"
)

// BSpline implements bi-cubic spline interpolation.
// This is a Go port of AGG's bspline class from agg_bspline.h/cpp.
//
// A very simple class of Bi-cubic Spline interpolation.
// First call Init(num, x[], y[]) or use NewBSplineFromPoints() where num is the number of source points,
// x, y are slices of X and Y values respectively. Here Y must be a function
// of X. It means that all the X-coordinates must be arranged in the ascending order.
// Then call Get(x) that calculates a value Y for the respective X.
// The class supports extrapolation, i.e. you can call Get(x) where x is
// outside the given X-range. Extrapolation is a simple linear function.
type BSpline struct {
	max     int
	num     int
	x       []float64                // X coordinates
	y       []float64                // Y coordinates
	am      *array.PodArray[float64] // Coefficient array
	lastIdx int                      // For optimization in GetStateful
}

// NewBSpline creates a new empty B-spline.
func NewBSpline() *BSpline {
	return &BSpline{
		max:     0,
		num:     0,
		x:       nil,
		y:       nil,
		am:      array.NewPodArray[float64](),
		lastIdx: -1,
	}
}

// NewBSplineWithCapacity creates a new B-spline with the specified maximum capacity.
func NewBSplineWithCapacity(maxPoints int) *BSpline {
	bs := NewBSpline()
	bs.Init(maxPoints)
	return bs
}

// NewBSplineFromPoints creates a new B-spline initialized with the given points.
// The x coordinates must be in ascending order.
func NewBSplineFromPoints(x, y []float64) *BSpline {
	bs := NewBSpline()
	bs.InitFromPoints(x, y)
	return bs
}

// Init initializes the B-spline with a maximum number of points.
// This allocates memory for the internal arrays.
func (bs *BSpline) Init(maxPoints int) {
	if maxPoints > 2 && maxPoints > bs.max {
		// Allocate space for coefficients and coordinate arrays
		// Layout: [coefficients][x coordinates][y coordinates]
		bs.am.Resize(maxPoints * 3)
		bs.max = maxPoints

		// Set up pointers to the different sections of the array
		data := bs.am.Data()
		bs.x = data[maxPoints : maxPoints*2]
		bs.y = data[maxPoints*2 : maxPoints*3]
	}
	bs.num = 0
	bs.lastIdx = -1
}

// AddPoint adds a control point to the spline.
// Points must be added in ascending order of X coordinates.
func (bs *BSpline) AddPoint(x, y float64) {
	if bs.num < bs.max {
		bs.x[bs.num] = x
		bs.y[bs.num] = y
		bs.num++
	}
}

// Prepare calculates the spline coefficients.
// This must be called after all points have been added and before calling Get().
func (bs *BSpline) Prepare() {
	if bs.num <= 2 {
		bs.lastIdx = -1
		return
	}

	// Initialize coefficient array to zero
	data := bs.am.Data()
	for k := 0; k < bs.num; k++ {
		data[k] = 0.0
	}

	// Create temporary arrays for the tridiagonal system
	n1 := 3 * bs.num
	temp := make([]float64, n1)
	for k := 0; k < n1; k++ {
		temp[k] = 0.0
	}

	// Set up pointers to different sections
	r := temp[bs.num:]   // Right diagonal
	s := temp[bs.num*2:] // Right-hand side

	n1 = bs.num - 1
	d := bs.x[1] - bs.x[0]
	e := (bs.y[1] - bs.y[0]) / d

	// Build the tridiagonal system
	for k := 1; k < n1; k++ {
		h := d
		d = bs.x[k+1] - bs.x[k]
		f := e
		e = (bs.y[k+1] - bs.y[k]) / d
		temp[k] = d / (d + h)          // Left diagonal
		r[k] = 1.0 - temp[k]           // Right diagonal
		s[k] = 6.0 * (e - f) / (h + d) // Right-hand side
	}

	// Forward elimination
	for k := 1; k < n1; k++ {
		p := 1.0 / (r[k]*temp[k-1] + 2.0)
		temp[k] *= -p
		s[k] = (s[k] - r[k]*s[k-1]) * p
	}

	// Back substitution
	data[n1] = 0.0
	temp[n1-1] = s[n1-1]
	data[n1-1] = temp[n1-1]

	for k, i := n1-2, 0; i < bs.num-2; i, k = i+1, k-1 {
		temp[k] = temp[k]*temp[k+1] + s[k]
		data[k] = temp[k]
	}

	bs.lastIdx = -1
}

// InitFromPoints initializes the B-spline from the given point arrays.
// The x coordinates must be in ascending order.
func (bs *BSpline) InitFromPoints(x, y []float64) {
	if len(x) != len(y) {
		panic("x and y slices must have the same length")
	}

	num := len(x)
	if num > 2 {
		bs.Init(num)
		for i := 0; i < num; i++ {
			bs.AddPoint(x[i], y[i])
		}
		bs.Prepare()
	}
	bs.lastIdx = -1
}

// bsearch performs binary search to find the interval containing x.
func (bs *BSpline) bsearch(x float64) int {
	if bs.num < 2 {
		return 0
	}

	i := 0
	j := bs.num - 1

	for j-i > 1 {
		k := (i + j) >> 1
		if x < bs.x[k] {
			j = k
		} else {
			i = k
		}
	}

	return i
}

// interpolation performs cubic spline interpolation at point x using interval i.
func (bs *BSpline) interpolation(x float64, i int) float64 {
	j := i + 1
	d := bs.x[i] - bs.x[j]
	h := x - bs.x[j]
	r := bs.x[i] - x
	p := d * d / 6.0

	coeffs := bs.am.Data()
	return (coeffs[j]*r*r*r+coeffs[i]*h*h*h)/6.0/d +
		((bs.y[j]-coeffs[j]*p)*r+(bs.y[i]-coeffs[i]*p)*h)/d
}

// extrapolationLeft performs linear extrapolation for x values to the left of the range.
func (bs *BSpline) extrapolationLeft(x float64) float64 {
	d := bs.x[1] - bs.x[0]
	coeffs := bs.am.Data()
	return (-d*coeffs[1]/6+(bs.y[1]-bs.y[0])/d)*(x-bs.x[0]) + bs.y[0]
}

// extrapolationRight performs linear extrapolation for x values to the right of the range.
func (bs *BSpline) extrapolationRight(x float64) float64 {
	d := bs.x[bs.num-1] - bs.x[bs.num-2]
	coeffs := bs.am.Data()
	return (d*coeffs[bs.num-2]/6+(bs.y[bs.num-1]-bs.y[bs.num-2])/d)*
		(x-bs.x[bs.num-1]) + bs.y[bs.num-1]
}

// Get calculates the interpolated Y value for the given X coordinate.
// This performs a full binary search each time.
func (bs *BSpline) Get(x float64) float64 {
	if bs.num <= 2 {
		return 0.0
	}

	// Extrapolation on the left
	if x < bs.x[0] {
		return bs.extrapolationLeft(x)
	}

	// Extrapolation on the right
	if x >= bs.x[bs.num-1] {
		return bs.extrapolationRight(x)
	}

	// Interpolation
	i := bs.bsearch(x)
	return bs.interpolation(x, i)
}

// GetStateful calculates the interpolated Y value using cached position information
// for better performance when calling Get multiple times with similar x values.
func (bs *BSpline) GetStateful(x float64) float64 {
	if bs.num <= 2 {
		return 0.0
	}

	// Extrapolation on the left
	if x < bs.x[0] {
		return bs.extrapolationLeft(x)
	}

	// Extrapolation on the right
	if x >= bs.x[bs.num-1] {
		return bs.extrapolationRight(x)
	}

	if bs.lastIdx >= 0 {
		// Check if x is not in current range
		if x < bs.x[bs.lastIdx] || x > bs.x[bs.lastIdx+1] {
			// Check optimization paths
			switch {
			case bs.lastIdx < bs.num-2 &&
				x >= bs.x[bs.lastIdx+1] &&
				x <= bs.x[bs.lastIdx+2]:
				// x is between next points (most probable)
				bs.lastIdx++
			case bs.lastIdx > 0 &&
				x >= bs.x[bs.lastIdx-1] &&
				x <= bs.x[bs.lastIdx]:
				// x is between previous points
				bs.lastIdx--
			default:
				// Perform full search
				bs.lastIdx = bs.bsearch(x)
			}
		}
		return bs.interpolation(x, bs.lastIdx)
	} else {
		// First call - perform full search
		bs.lastIdx = bs.bsearch(x)
		return bs.interpolation(x, bs.lastIdx)
	}
}

// NumPoints returns the number of control points in the spline.
func (bs *BSpline) NumPoints() int {
	return bs.num
}

// MaxPoints returns the maximum capacity of the spline.
func (bs *BSpline) MaxPoints() int {
	return bs.max
}

// Reset clears all points from the spline.
func (bs *BSpline) Reset() {
	bs.num = 0
	bs.lastIdx = -1
}
