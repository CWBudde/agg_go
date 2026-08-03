// Package agg2d blend modes for AGG2D high-level interface.
// This file contains blend mode constants and functionality that match the C++ AGG2D interface.
package agg2d

import (
	"github.com/cwbudde/agg_go/internal/pixfmt/blender"
)

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

	// Photoshop-compatible blending modes. These are appended deliberately:
	// the numeric values of every pre-existing mode are part of the public API.
	BlendDissolve
	BlendLinearBurn
	BlendDarkerColor
	BlendLinearDodge
	BlendLighterColor
	BlendVividLight
	BlendLinearLight
	BlendPinLight
	BlendHardMix
	BlendSubtract
	BlendDivide
	BlendHue
	BlendSaturation
	BlendColor
	BlendLuminosity
	// BlendColorBurnPhotoshop is Photoshop's/W3C's straight-colour Color Burn.
	// BlendColorBurn retains its historical AGG output for compatibility.
	BlendColorBurnPhotoshop
	// BlendSoftLightPhotoshop is the straight-colour Photoshop/W3C formula.
	// BlendSoftLight retains its historical AGG output for compatibility.
	BlendSoftLightPhotoshop
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
	case BlendDissolve:
		return "Dissolve"
	case BlendLinearBurn:
		return "LinearBurn"
	case BlendDarkerColor:
		return "DarkerColor"
	case BlendLinearDodge:
		return "LinearDodge"
	case BlendLighterColor:
		return "LighterColor"
	case BlendVividLight:
		return "VividLight"
	case BlendLinearLight:
		return "LinearLight"
	case BlendPinLight:
		return "PinLight"
	case BlendHardMix:
		return "HardMix"
	case BlendSubtract:
		return "Subtract"
	case BlendDivide:
		return "Divide"
	case BlendHue:
		return "Hue"
	case BlendSaturation:
		return "Saturation"
	case BlendColor:
		return "Color"
	case BlendLuminosity:
		return "Luminosity"
	case BlendColorBurnPhotoshop:
		return "ColorBurnPhotoshop"
	case BlendSoftLightPhotoshop:
		return "SoftLightPhotoshop"
	default:
		return "Unknown"
	}
}

// SetBlendMode sets the general blending mode.
// This matches the C++ Agg2D::blendMode(BlendMode m) method.
func (agg2d *Agg2D) SetBlendMode(mode BlendMode) {
	agg2d.blendMode = mode
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
	agg2d.SetImageBlendColor(Color{r, g, b, a})
}

// GetImageBlendColor returns the current image blend color.
// This matches the C++ Agg2D::imageBlendColor() const method.
func (agg2d *Agg2D) GetImageBlendColor() Color {
	return agg2d.imageBlendColor
}

// blendModeToCompOp converts AGG2D BlendMode to pixfmt CompOp
func blendModeToCompOp(mode BlendMode) blender.CompOp {
	switch mode {
	case BlendAlpha, BlendSrcOver:
		return blender.CompOpSrcOver
	case BlendClear:
		return blender.CompOpClear
	case BlendSrc:
		return blender.CompOpSrc
	case BlendDst:
		return blender.CompOpDst
	case BlendDstOver:
		return blender.CompOpDstOver
	case BlendSrcIn:
		return blender.CompOpSrcIn
	case BlendDstIn:
		return blender.CompOpDstIn
	case BlendSrcOut:
		return blender.CompOpSrcOut
	case BlendDstOut:
		return blender.CompOpDstOut
	case BlendSrcAtop:
		return blender.CompOpSrcAtop
	case BlendDstAtop:
		return blender.CompOpDstAtop
	case BlendXor:
		return blender.CompOpXor
	case BlendAdd:
		return blender.CompOpPlus
	case BlendMultiply:
		return blender.CompOpMultiply
	case BlendScreen:
		return blender.CompOpScreen
	case BlendOverlay:
		return blender.CompOpOverlay
	case BlendDarken:
		return blender.CompOpDarken
	case BlendLighten:
		return blender.CompOpLighten
	case BlendColorDodge:
		return blender.CompOpColorDodge
	case BlendColorBurn:
		return blender.CompOpColorBurn
	case BlendHardLight:
		return blender.CompOpHardLight
	case BlendSoftLight:
		return blender.CompOpSoftLight
	case BlendDifference:
		return blender.CompOpDifference
	case BlendExclusion:
		return blender.CompOpExclusion
	case BlendDissolve:
		return blender.CompOpDissolve
	case BlendLinearBurn:
		return blender.CompOpLinearBurn
	case BlendDarkerColor:
		return blender.CompOpDarkerColor
	case BlendLinearDodge:
		return blender.CompOpPlus
	case BlendLighterColor:
		return blender.CompOpLighterColor
	case BlendVividLight:
		return blender.CompOpVividLight
	case BlendLinearLight:
		return blender.CompOpLinearLight
	case BlendPinLight:
		return blender.CompOpPinLight
	case BlendHardMix:
		return blender.CompOpHardMix
	case BlendSubtract:
		return blender.CompOpSubtract
	case BlendDivide:
		return blender.CompOpDivide
	case BlendHue:
		return blender.CompOpHue
	case BlendSaturation:
		return blender.CompOpSaturation
	case BlendColor:
		return blender.CompOpColor
	case BlendLuminosity:
		return blender.CompOpLuminosity
	case BlendColorBurnPhotoshop:
		return blender.CompOpColorBurnPhotoshop
	case BlendSoftLightPhotoshop:
		return blender.CompOpSoftLightPhotoshop
	default:
		return blender.CompOpSrcOver // Default to SrcOver for unsupported modes
	}
}

// updateBlendMode applies the current blend mode to the rendering pipeline
func (agg2d *Agg2D) updateBlendMode() {
	// Update the composite operation on the composite pixel format
	if agg2d.pixfmtComp != nil {
		compOp := blendModeToCompOp(agg2d.blendMode)
		agg2d.pixfmtComp.SetCompOp(compOp)
		if agg2d.pixfmtCompPre != nil {
			agg2d.pixfmtCompPre.SetCompOp(compOp)
		}
	}
}

// IsPorterDuffMode returns true if the blend mode is a Porter-Duff compositing mode
func IsPorterDuffMode(mode BlendMode) bool {
	return mode >= BlendClear && mode <= BlendXor
}

// IsExtendedBlendMode returns true if the blend mode is an extended blending mode
func IsExtendedBlendMode(mode BlendMode) bool {
	return mode >= BlendAdd && mode <= BlendSoftLightPhotoshop
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
	return mode >= BlendAlpha && mode <= BlendSoftLightPhotoshop
}
