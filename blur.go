package agg

import (
	"github.com/cwbudde/agg_go/internal/blur"
	"github.com/cwbudde/agg_go/internal/color"
)

// ---------------------------------------------------------------------------
// Stack blur re-exports — thin aliases over internal/blur implementations.
//
// The full implementations live in internal/blur/ to keep the root package
// focused on the public API surface.  These Go 1.24 generic type aliases
// preserve full backward compatibility.
// ---------------------------------------------------------------------------

// StackBlur is the RGBA8 stack blur, mirroring C++ AGG's stack_blur<rgba8>.
type StackBlur[CS color.Space] = blur.StackBlurRGBA8[CS]

// NewStackBlur creates a new RGBA8 stack blur instance.
func NewStackBlur[CS color.Space]() *StackBlur[CS] {
	return blur.NewStackBlurRGBA8[CS]()
}

// StackBlurRGBA16 is the RGBA16 stack blur, mirroring C++ AGG's stack_blur<rgba16>.
type StackBlurRGBA16[CS color.Space] = blur.StackBlurRGBA16[CS]

// NewStackBlurRGBA16 creates a new RGBA16 stack blur instance.
func NewStackBlurRGBA16[CS color.Space]() *StackBlurRGBA16[CS] {
	return blur.NewStackBlurRGBA16[CS]()
}

// StackBlurRGB is the RGB8 stack blur, mirroring C++ AGG's stack_blur<rgb8>.
type StackBlurRGB[CS color.Space] = blur.StackBlurRGB8[CS]

// NewStackBlurRGB creates a new RGB8 stack blur instance.
func NewStackBlurRGB[CS color.Space]() *StackBlurRGB[CS] {
	return blur.NewStackBlurRGB8[CS]()
}

// StackBlurRGB16 is the RGB16 stack blur, mirroring C++ AGG's stack_blur<rgb16>.
type StackBlurRGB16[CS color.Space] = blur.StackBlurRGB16[CS]

// NewStackBlurRGB16 creates a new RGB16 stack blur instance.
func NewStackBlurRGB16[CS color.Space]() *StackBlurRGB16[CS] {
	return blur.NewStackBlurRGB16[CS]()
}

// StackBlurGray is the Gray8 stack blur, mirroring C++ AGG's stack_blur<gray8>.
type StackBlurGray[CS color.Space] = blur.StackBlurGray8[CS]

// NewStackBlurGray creates a new Gray8 stack blur instance.
func NewStackBlurGray[CS color.Space]() *StackBlurGray[CS] {
	return blur.NewStackBlurGray8[CS]()
}

// StackBlurGray16 is the Gray16 stack blur, mirroring C++ AGG's stack_blur<gray16>.
type StackBlurGray16[CS color.Space] = blur.StackBlurGray16[CS]

// NewStackBlurGray16 creates a new Gray16 stack blur instance.
func NewStackBlurGray16[CS color.Space]() *StackBlurGray16[CS] {
	return blur.NewStackBlurGray16[CS]()
}
