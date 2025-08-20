// Package transform provides viewport transformation functionality for AGG.
// This implements a port of AGG's trans_viewport class.
package transform

import (
	"encoding/binary"
	"math"
)

// Compile-time interface checks
var _ Transformer = (*TransViewport)(nil)
var _ InverseTransformer = (*TransViewport)(nil)

// AspectRatio defines how aspect ratio is preserved during viewport transformations.
type AspectRatio int

const (
	// AspectRatioStretch stretches to fit the entire viewport, ignoring aspect ratio
	AspectRatioStretch AspectRatio = iota
	// AspectRatioMeet scales to fit entirely within viewport, preserving aspect ratio
	AspectRatioMeet
	// AspectRatioSlice scales to fill entire viewport, preserving aspect ratio (may crop)
	AspectRatioSlice
)

// TransViewport represents a viewport transformation system for converting
// between world coordinates and device (screen) coordinates.
// It provides orthogonal conversions with optional aspect ratio preservation.
type TransViewport struct {
	// Original viewport bounds
	worldX1, worldY1, worldX2, worldY2     float64
	deviceX1, deviceY1, deviceX2, deviceY2 float64

	// Aspect ratio and alignment settings
	aspect AspectRatio
	alignX float64 // 0.0 = left/bottom, 0.5 = center, 1.0 = right/top
	alignY float64

	// Calculated transformation parameters
	wx1, wy1, wx2, wy2 float64 // Actual world bounds after aspect ratio adjustment
	dx1, dy1           float64 // Device viewport top-left
	kx, ky             float64 // Scale factors
	isValid            bool    // Whether the transformation is valid
}

// NewTransViewport creates a new viewport transformation with default settings.
// Initially maps world coordinates (0,0)-(1,1) to device coordinates (0,0)-(1,1).
func NewTransViewport() *TransViewport {
	v := &TransViewport{
		worldX1:  0.0,
		worldY1:  0.0,
		worldX2:  1.0,
		worldY2:  1.0,
		deviceX1: 0.0,
		deviceY1: 0.0,
		deviceX2: 1.0,
		deviceY2: 1.0,
		aspect:   AspectRatioStretch,
		alignX:   0.5,
		alignY:   0.5,
		wx1:      0.0,
		wy1:      0.0,
		wx2:      1.0,
		wy2:      1.0,
		dx1:      0.0,
		dy1:      0.0,
		kx:       1.0,
		ky:       1.0,
		isValid:  true,
	}
	return v
}

// PreserveAspectRatio sets the aspect ratio preservation mode and alignment.
// alignX and alignY specify how to align the preserved aspect ratio (0.0 to 1.0).
// 0.0 = left/bottom, 0.5 = center, 1.0 = right/top
func (v *TransViewport) PreserveAspectRatio(alignX, alignY float64, aspect AspectRatio) {
	v.alignX = alignX
	v.alignY = alignY
	v.aspect = aspect
	v.update()
}

// DeviceViewport sets the device (screen) coordinate bounds.
func (v *TransViewport) DeviceViewport(x1, y1, x2, y2 float64) {
	v.deviceX1 = x1
	v.deviceY1 = y1
	v.deviceX2 = x2
	v.deviceY2 = y2
	v.update()
}

// WorldViewport sets the world coordinate bounds.
func (v *TransViewport) WorldViewport(x1, y1, x2, y2 float64) {
	v.worldX1 = x1
	v.worldY1 = y1
	v.worldX2 = x2
	v.worldY2 = y2
	v.update()
}

// GetDeviceViewport returns the device coordinate bounds.
func (v *TransViewport) GetDeviceViewport() (x1, y1, x2, y2 float64) {
	return v.deviceX1, v.deviceY1, v.deviceX2, v.deviceY2
}

// GetWorldViewport returns the world coordinate bounds.
func (v *TransViewport) GetWorldViewport() (x1, y1, x2, y2 float64) {
	return v.worldX1, v.worldY1, v.worldX2, v.worldY2
}

// GetWorldViewportActual returns the actual world bounds after aspect ratio adjustment.
func (v *TransViewport) GetWorldViewportActual() (x1, y1, x2, y2 float64) {
	return v.wx1, v.wy1, v.wx2, v.wy2
}

// IsValid returns whether the transformation is valid.
func (v *TransViewport) IsValid() bool {
	return v.isValid
}

// AlignX returns the horizontal alignment factor.
func (v *TransViewport) AlignX() float64 {
	return v.alignX
}

// AlignY returns the vertical alignment factor.
func (v *TransViewport) AlignY() float64 {
	return v.alignY
}

// AspectRatio returns the current aspect ratio preservation mode.
func (v *TransViewport) AspectRatio() AspectRatio {
	return v.aspect
}

// Transform converts world coordinates to device coordinates.
func (v *TransViewport) Transform(x, y *float64) {
	*x = (*x-v.wx1)*v.kx + v.dx1
	*y = (*y-v.wy1)*v.ky + v.dy1
}

// TransformScaleOnly applies only the scaling transformation.
func (v *TransViewport) TransformScaleOnly(x, y *float64) {
	*x *= v.kx
	*y *= v.ky
}

// InverseTransform converts device coordinates to world coordinates.
func (v *TransViewport) InverseTransform(x, y *float64) {
	*x = (*x-v.dx1)/v.kx + v.wx1
	*y = (*y-v.dy1)/v.ky + v.wy1
}

// InverseTransformScaleOnly applies only the inverse scaling transformation.
func (v *TransViewport) InverseTransformScaleOnly(x, y *float64) {
	*x /= v.kx
	*y /= v.ky
}

// update recalculates the transformation parameters based on current settings.
// This is called automatically when viewport bounds or aspect ratio settings change.
func (v *TransViewport) update() {
	const epsilon = 1e-30

	// Check for invalid/degenerate viewports
	if math.Abs(v.worldX1-v.worldX2) < epsilon ||
		math.Abs(v.worldY1-v.worldY2) < epsilon ||
		math.Abs(v.deviceX1-v.deviceX2) < epsilon ||
		math.Abs(v.deviceY1-v.deviceY2) < epsilon {

		// Set to safe defaults for invalid viewport
		v.wx1 = v.worldX1
		v.wy1 = v.worldY1
		v.wx2 = v.worldX1 + 1.0
		v.wy2 = v.worldY1 + 1.0
		v.dx1 = v.deviceX1
		v.dy1 = v.deviceY1
		v.kx = 1.0
		v.ky = 1.0
		v.isValid = false
		return
	}

	// Start with original bounds
	worldX1 := v.worldX1
	worldY1 := v.worldY1
	worldX2 := v.worldX2
	worldY2 := v.worldY2
	deviceX1 := v.deviceX1
	deviceY1 := v.deviceY1
	deviceX2 := v.deviceX2
	deviceY2 := v.deviceY2

	// Handle aspect ratio preservation
	if v.aspect != AspectRatioStretch {
		// Calculate initial scale factors
		v.kx = (deviceX2 - deviceX1) / (worldX2 - worldX1)
		v.ky = (deviceY2 - deviceY1) / (worldY2 - worldY1)

		// Determine which dimension to adjust based on aspect ratio mode
		var d float64
		if (v.aspect == AspectRatioMeet) == (v.kx < v.ky) {
			// Adjust Y dimension to match X scale
			d = (worldY2 - worldY1) * v.ky / v.kx
			worldY1 += (worldY2 - worldY1 - d) * v.alignY
			worldY2 = worldY1 + d
		} else {
			// Adjust X dimension to match Y scale
			d = (worldX2 - worldX1) * v.kx / v.ky
			worldX1 += (worldX2 - worldX1 - d) * v.alignX
			worldX2 = worldX1 + d
		}
	}

	// Store the final calculated values
	v.wx1 = worldX1
	v.wy1 = worldY1
	v.wx2 = worldX2
	v.wy2 = worldY2
	v.dx1 = deviceX1
	v.dy1 = deviceY1
	v.kx = (deviceX2 - deviceX1) / (worldX2 - worldX1)
	v.ky = (deviceY2 - deviceY1) / (worldY2 - worldY1)
	v.isValid = true
}

// DeviceDX returns the X offset in device coordinates.
func (v *TransViewport) DeviceDX() float64 {
	return v.dx1 - v.wx1*v.kx
}

// DeviceDY returns the Y offset in device coordinates.
func (v *TransViewport) DeviceDY() float64 {
	return v.dy1 - v.wy1*v.ky
}

// ScaleX returns the X scale factor.
func (v *TransViewport) ScaleX() float64 {
	return v.kx
}

// ScaleY returns the Y scale factor.
func (v *TransViewport) ScaleY() float64 {
	return v.ky
}

// Scale returns the average scale factor.
func (v *TransViewport) Scale() float64 {
	return (v.kx + v.ky) * 0.5
}

// ToAffine converts the viewport transformation to an affine transformation matrix.
func (v *TransViewport) ToAffine() *TransAffine {
	mtx := NewTransAffineTranslation(-v.wx1, -v.wy1)
	mtx.Multiply(NewTransAffineScalingXY(v.kx, v.ky))
	mtx.Multiply(NewTransAffineTranslation(v.dx1, v.dy1))
	return mtx
}

// ToAffineScaleOnly converts only the scaling part to an affine transformation matrix.
func (v *TransViewport) ToAffineScaleOnly() *TransAffine {
	return NewTransAffineScalingXY(v.kx, v.ky)
}

// ByteSize returns the number of bytes required to serialize this viewport.
func (v *TransViewport) ByteSize() int {
	// 12 float64 values (8 bytes each) + 1 int (aspect ratio) + 1 bool + padding
	return 12*8 + 4 + 1 + 3 // 100 bytes total with padding
}

// Serialize writes the viewport data to a byte slice.
// The slice must be at least ByteSize() bytes long.
func (v *TransViewport) Serialize(data []byte) error {
	if len(data) < v.ByteSize() {
		return &SerializationError{"insufficient buffer size"}
	}

	offset := 0

	// Serialize float64 values
	floats := []float64{
		v.worldX1, v.worldY1, v.worldX2, v.worldY2,
		v.deviceX1, v.deviceY1, v.deviceX2, v.deviceY2,
		v.alignX, v.alignY, v.kx, v.ky,
	}

	for _, f := range floats {
		binary.LittleEndian.PutUint64(data[offset:], math.Float64bits(f))
		offset += 8
	}

	// Serialize aspect ratio (as int32)
	binary.LittleEndian.PutUint32(data[offset:], uint32(v.aspect))
	offset += 4

	// Serialize bool
	if v.isValid {
		data[offset] = 1
	} else {
		data[offset] = 0
	}

	return nil
}

// Deserialize reads viewport data from a byte slice.
func (v *TransViewport) Deserialize(data []byte) error {
	if len(data) < v.ByteSize() {
		return &SerializationError{"insufficient data size"}
	}

	offset := 0

	// Deserialize float64 values
	floatPtrs := []*float64{
		&v.worldX1, &v.worldY1, &v.worldX2, &v.worldY2,
		&v.deviceX1, &v.deviceY1, &v.deviceX2, &v.deviceY2,
		&v.alignX, &v.alignY, &v.kx, &v.ky,
	}

	for _, ptr := range floatPtrs {
		bits := binary.LittleEndian.Uint64(data[offset:])
		*ptr = math.Float64frombits(bits)
		offset += 8
	}

	// Deserialize aspect ratio
	v.aspect = AspectRatio(binary.LittleEndian.Uint32(data[offset:]))
	offset += 4

	// Deserialize bool
	v.isValid = data[offset] != 0

	// Recalculate derived values
	v.update()

	return nil
}

// SerializationError represents an error during serialization/deserialization.
type SerializationError struct {
	Message string
}

func (e *SerializationError) Error() string {
	return "viewport serialization error: " + e.Message
}
