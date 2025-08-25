// Package agg2d blend modes for AGG2D high-level interface.
// This file contains blend mode constants and functionality that match the C++ AGG2D interface.
package agg2d

// BlendMode represents the different blending modes available in AGG2D
// These values match the C++ AGG2D BlendMode enum
const (
	// Standard alpha blending (default)
	BlendAlpha BlendMode = iota

	// Porter-Duff compositing operations
	BlendClear   // Clear destination
	BlendSrc     // Source replaces destination
	BlendDst     // Destination unchanged
	BlendSrcOver // Source over destination (standard alpha blend)
	BlendDstOver // Destination over source
	BlendSrcIn   // Source in destination
	BlendDstIn   // Destination in source
	BlendSrcOut  // Source out destination
	BlendDstOut  // Destination out source
	BlendSrcAtop // Source atop destination
	BlendDstAtop // Destination atop source
	BlendXor     // Source XOR destination

	// Additional blending modes
	BlendAdd        // Additive blending
	BlendMultiply   // Multiply blending
	BlendScreen     // Screen blending
	BlendOverlay    // Overlay blending
	BlendDarken     // Darken blending
	BlendLighten    // Lighten blending
	BlendColorDodge // Color dodge blending
	BlendColorBurn  // Color burn blending
	BlendHardLight  // Hard light blending
	BlendSoftLight  // Soft light blending
	BlendDifference // Difference blending
	BlendExclusion  // Exclusion blending
)

// BlendModeString returns a string representation of the blend mode
func BlendModeString(mode BlendMode) string {
	switch mode {
	case BlendAlpha:
		return "Alpha"
	case BlendClear:
		return "Clear"
	case BlendSrc:
		return "Src"
	case BlendDst:
		return "Dst"
	case BlendSrcOver:
		return "SrcOver"
	case BlendDstOver:
		return "DstOver"
	case BlendSrcIn:
		return "SrcIn"
	case BlendDstIn:
		return "DstIn"
	case BlendSrcOut:
		return "SrcOut"
	case BlendDstOut:
		return "DstOut"
	case BlendSrcAtop:
		return "SrcAtop"
	case BlendDstAtop:
		return "DstAtop"
	case BlendXor:
		return "Xor"
	case BlendAdd:
		return "Add"
	case BlendMultiply:
		return "Multiply"
	case BlendScreen:
		return "Screen"
	case BlendOverlay:
		return "Overlay"
	case BlendDarken:
		return "Darken"
	case BlendLighten:
		return "Lighten"
	case BlendColorDodge:
		return "ColorDodge"
	case BlendColorBurn:
		return "ColorBurn"
	case BlendHardLight:
		return "HardLight"
	case BlendSoftLight:
		return "SoftLight"
	case BlendDifference:
		return "Difference"
	case BlendExclusion:
		return "Exclusion"
	default:
		return "Unknown"
	}
}

// SetBlendMode sets the general blending mode.
// This matches the C++ Agg2D::blendMode(BlendMode m) method.
func (agg2d *Agg2D) SetBlendMode(mode BlendMode) {
	agg2d.blendMode = mode
	// TODO: Apply blend mode to pixel format and renderers
	agg2d.updateBlendMode()
}

// GetBlendMode returns the current general blending mode.
// This matches the C++ Agg2D::blendMode() const method.
func (agg2d *Agg2D) GetBlendMode() BlendMode {
	return agg2d.blendMode
}

// SetImageBlendMode sets the image blending mode.
// This matches the C++ Agg2D::imageBlendMode(BlendMode m) method.
func (agg2d *Agg2D) SetImageBlendMode(mode BlendMode) {
	agg2d.imageBlendMode = mode
	// TODO: Apply image blend mode to image operations
}

// GetImageBlendMode returns the current image blending mode.
// This matches the C++ Agg2D::imageBlendMode() const method.
func (agg2d *Agg2D) GetImageBlendMode() BlendMode {
	return agg2d.imageBlendMode
}

// SetImageBlendColor sets the image blend color.
// This matches the C++ Agg2D::imageBlendColor(Color c) method.
func (agg2d *Agg2D) SetImageBlendColor(c Color) {
	agg2d.imageBlendColor = c
}

// SetImageBlendColorRGBA sets the image blend color using RGBA values.
// This matches the C++ Agg2D::imageBlendColor(unsigned r, g, b, a) method.
func (agg2d *Agg2D) SetImageBlendColorRGBA(r, g, b, a uint8) {
	agg2d.imageBlendColor = Color{r, g, b, a}
}

// GetImageBlendColor returns the current image blend color.
// This matches the C++ Agg2D::imageBlendColor() const method.
func (agg2d *Agg2D) GetImageBlendColor() Color {
	return agg2d.imageBlendColor
}

// updateBlendMode applies the current blend mode to the rendering pipeline
func (agg2d *Agg2D) updateBlendMode() {
	// TODO: This needs to be implemented based on the actual pixel format and renderer interfaces
	// The implementation will vary depending on whether we're using:
	// - PixFormatRGBA32 with alpha blending
	// - PixFormatComp with custom blending
	// - PixFormatPre with premultiplied alpha
	// - PixFormatCompPre with premultiplied custom blending

	// For now, this is a placeholder that needs to be connected to the actual
	// pixel format and renderer system once those interfaces are finalized
}

// IsPorterDuffMode returns true if the blend mode is a Porter-Duff compositing mode
func IsPorterDuffMode(mode BlendMode) bool {
	return mode >= BlendClear && mode <= BlendXor
}

// IsExtendedBlendMode returns true if the blend mode is an extended blending mode
func IsExtendedBlendMode(mode BlendMode) bool {
	return mode >= BlendAdd && mode <= BlendExclusion
}

// RequiresPremultipliedAlpha returns true if the blend mode requires premultiplied alpha
func RequiresPremultipliedAlpha(mode BlendMode) bool {
	// Most Porter-Duff modes work better with premultiplied alpha
	return IsPorterDuffMode(mode) || mode == BlendAlpha
}

// GetDefaultBlendMode returns the default blend mode (alpha blending)
func GetDefaultBlendMode() BlendMode {
	return BlendAlpha
}

// ValidateBlendMode returns true if the blend mode is valid
func ValidateBlendMode(mode BlendMode) bool {
	return mode >= BlendAlpha && mode <= BlendExclusion
}
