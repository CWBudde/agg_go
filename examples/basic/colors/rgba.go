// Package main demonstrates basic RGBA color handling in AGG Go.
package main

import (
	"fmt"
	"os"

	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/pixfmt"
)

func main() {
	fmt.Println("AGG Go - RGBA Color Example")
	fmt.Println("============================")

	// Demonstrate RGBA8 color operations
	demonstrateRGBA8Colors()

	// Demonstrate arithmetic operations
	demonstrateRGBA8Arithmetic()

	// Demonstrate color conversions
	demonstrateColorConversions()

	// Demonstrate blending operations
	demonstrateBlending()

	// Demonstrate premultiplication
	demonstratePremultiplication()

	// Demonstrate color orders
	demonstrateColorOrders()

	fmt.Println("\nRGBA example completed successfully!")
}

func demonstrateRGBA8Colors() {
	fmt.Println("\n1. RGBA8 Color Creation and Properties")
	fmt.Println("======================================")

	// Create different RGBA colors
	red := color.NewRGBA8[color.Linear](255, 0, 0, 255)
	green := color.NewRGBA8[color.Linear](0, 255, 0, 255)
	blue := color.NewRGBA8[color.Linear](0, 0, 255, 255)
	transparent := color.NewRGBA8[color.Linear](128, 128, 128, 0)
	semitransparent := color.NewRGBA8[color.Linear](200, 100, 50, 128)

	fmt.Printf("Red:             R=%3d, G=%3d, B=%3d, A=%3d - Opaque: %t, Transparent: %t\n",
		red.R, red.G, red.B, red.A, red.IsOpaque(), red.IsTransparent())
	fmt.Printf("Green:           R=%3d, G=%3d, B=%3d, A=%3d - Opaque: %t, Transparent: %t\n",
		green.R, green.G, green.B, green.A, green.IsOpaque(), green.IsTransparent())
	fmt.Printf("Blue:            R=%3d, G=%3d, B=%3d, A=%3d - Opaque: %t, Transparent: %t\n",
		blue.R, blue.G, blue.B, blue.A, blue.IsOpaque(), blue.IsTransparent())
	fmt.Printf("Transparent:     R=%3d, G=%3d, B=%3d, A=%3d - Opaque: %t, Transparent: %t\n",
		transparent.R, transparent.G, transparent.B, transparent.A, transparent.IsOpaque(), transparent.IsTransparent())
	fmt.Printf("Semi-transparent: R=%3d, G=%3d, B=%3d, A=%3d - Opaque: %t, Transparent: %t\n",
		semitransparent.R, semitransparent.G, semitransparent.B, semitransparent.A, semitransparent.IsOpaque(), semitransparent.IsTransparent())

	// Test opacity manipulation
	adjustable := color.NewRGBA8[color.Linear](100, 150, 200, 255)
	fmt.Printf("Original:        R=%3d, G=%3d, B=%3d, A=%3d (Opacity: %.2f)\n",
		adjustable.R, adjustable.G, adjustable.B, adjustable.A, adjustable.GetOpacity())

	adjustable.Opacity(0.5)
	fmt.Printf("50%% Opacity:     R=%3d, G=%3d, B=%3d, A=%3d (Opacity: %.2f)\n",
		adjustable.R, adjustable.G, adjustable.B, adjustable.A, adjustable.GetOpacity())

	adjustable.Opacity(0.75)
	fmt.Printf("75%% Opacity:     R=%3d, G=%3d, B=%3d, A=%3d (Opacity: %.2f)\n",
		adjustable.R, adjustable.G, adjustable.B, adjustable.A, adjustable.GetOpacity())
}

func demonstrateRGBA8Arithmetic() {
	fmt.Println("\n2. RGBA8 Arithmetic Operations")
	fmt.Println("==============================")

	// Test basic arithmetic functions
	a := basics.Int8u(100)
	b := basics.Int8u(150)
	alpha := basics.Int8u(128) // 50%

	multiply := color.RGBA8Multiply(a, b)
	lerp := color.RGBA8Lerp(a, b, alpha)
	prelerp := color.RGBA8Prelerp(a, b, alpha)

	fmt.Printf("RGBA8Multiply(%d, %d) = %d\n", a, b, multiply)
	fmt.Printf("RGBA8Lerp(%d, %d, %d) = %d (50%% blend: %d→%d)\n", a, b, alpha, lerp, a, b)
	fmt.Printf("RGBA8Prelerp(%d, %d, %d) = %d\n", a, b, alpha, prelerp)

	// Test color arithmetic
	color1 := color.NewRGBA8[color.Linear](100, 50, 200, 255)
	color2 := color.NewRGBA8[color.Linear](50, 150, 25, 128)

	sum := color1.Add(color2)
	scaled := color1.Scale(0.5)
	gradient := color1.Gradient(color2, 128) // 50% blend

	fmt.Printf("\nColor1:          R=%3d, G=%3d, B=%3d, A=%3d\n", color1.R, color1.G, color1.B, color1.A)
	fmt.Printf("Color2:          R=%3d, G=%3d, B=%3d, A=%3d\n", color2.R, color2.G, color2.B, color2.A)
	fmt.Printf("Add:             R=%3d, G=%3d, B=%3d, A=%3d\n", sum.R, sum.G, sum.B, sum.A)
	fmt.Printf("Scale(0.5):      R=%3d, G=%3d, B=%3d, A=%3d\n", scaled.R, scaled.G, scaled.B, scaled.A)
	fmt.Printf("Gradient(50%%):   R=%3d, G=%3d, B=%3d, A=%3d\n", gradient.R, gradient.G, gradient.B, gradient.A)
}

func demonstrateColorConversions() {
	fmt.Println("\n3. Color Conversion Examples")
	fmt.Println("============================")

	// Create a floating-point RGBA color
	rgba := color.NewRGBA(0.8, 0.4, 0.6, 0.9)
	fmt.Printf("Original RGBA:   R=%.2f, G=%.2f, B=%.2f, A=%.2f\n", rgba.R, rgba.G, rgba.B, rgba.A)

	// Convert to 8-bit
	rgba8 := color.ConvertFromRGBA[color.Linear](rgba)
	fmt.Printf("As RGBA8:        R=%3d, G=%3d, B=%3d, A=%3d\n", rgba8.R, rgba8.G, rgba8.B, rgba8.A)

	// Convert back
	rgbaBack := rgba8.ConvertToRGBA()
	fmt.Printf("Back to RGBA:    R=%.2f, G=%.2f, B=%.2f, A=%.2f\n", rgbaBack.R, rgbaBack.G, rgbaBack.B, rgbaBack.A)

	// Test colorspace types
	fmt.Println("\nColorspace Types:")
	linearRGBA := color.NewRGBA8[color.Linear](128, 64, 192, 255)
	srgbRGBA := color.NewRGBA8[color.SRGB](128, 64, 192, 255)

	fmt.Printf("Linear RGBA8:    R=%3d, G=%3d, B=%3d, A=%3d\n", linearRGBA.R, linearRGBA.G, linearRGBA.B, linearRGBA.A)
	fmt.Printf("sRGB RGBA8:      R=%3d, G=%3d, B=%3d, A=%3d\n", srgbRGBA.R, srgbRGBA.G, srgbRGBA.B, srgbRGBA.A)
}

func demonstrateBlending() {
	fmt.Println("\n4. Blending Operations")
	fmt.Println("======================")

	// Create different blenders
	normalBlender := pixfmt.BlenderRGBA8{}
	preBlender := pixfmt.BlenderRGBA8Pre{}

	// Test pixel buffer (RGBA format)
	dst := []basics.Int8u{100, 100, 100, 255}              // Gray background
	src := color.NewRGBA8[color.Linear](200, 150, 50, 128) // Orange with 50% alpha

	fmt.Printf("Background:      R=%3d, G=%3d, B=%3d, A=%3d\n", dst[0], dst[1], dst[2], dst[3])
	fmt.Printf("Source:          R=%3d, G=%3d, B=%3d, A=%3d (50%% alpha)\n", src.R, src.G, src.B, src.A)

	// Test normal blending
	dstNormal := make([]basics.Int8u, 4)
	copy(dstNormal, dst)
	normalBlender.BlendPix(dstNormal, src.R, src.G, src.B, src.A, 255)
	fmt.Printf("Normal blend:    R=%3d, G=%3d, B=%3d, A=%3d\n", dstNormal[0], dstNormal[1], dstNormal[2], dstNormal[3])

	// Test premultiplied blending
	dstPre := make([]basics.Int8u, 4)
	copy(dstPre, dst)
	// For premultiplied, we need to premultiply the source first
	premultSrc := src
	premultSrc.Premultiply()
	preBlender.BlendPix(dstPre, premultSrc.R, premultSrc.G, premultSrc.B, premultSrc.A, 255)
	fmt.Printf("Premult blend:   R=%3d, G=%3d, B=%3d, A=%3d\n", dstPre[0], dstPre[1], dstPre[2], dstPre[3])

	// Test blending with varying coverage
	fmt.Println("\nBlending with varying coverage:")
	coverages := []basics.Int8u{255, 200, 150, 100, 50, 0}
	for _, cover := range coverages {
		dstCover := make([]basics.Int8u, 4)
		copy(dstCover, dst)
		normalBlender.BlendPix(dstCover, src.R, src.G, src.B, src.A, cover)
		coveragePercent := float64(cover) / 255.0 * 100
		fmt.Printf("Coverage %3d%% :   R=%3d, G=%3d, B=%3d, A=%3d\n",
			int(coveragePercent), dstCover[0], dstCover[1], dstCover[2], dstCover[3])
	}
}

func demonstratePremultiplication() {
	fmt.Println("\n5. Premultiplication Examples")
	fmt.Println("=============================")

	// Create a semi-transparent color
	original := color.NewRGBA8[color.Linear](200, 100, 50, 128) // 50% alpha
	fmt.Printf("Original:        R=%3d, G=%3d, B=%3d, A=%3d\n", original.R, original.G, original.B, original.A)

	// Premultiply
	premult := original
	premult.Premultiply()
	fmt.Printf("Premultiplied:   R=%3d, G=%3d, B=%3d, A=%3d\n", premult.R, premult.G, premult.B, premult.A)

	// Demultiply back
	demult := premult
	demult.Demultiply()
	fmt.Printf("Demultiplied:    R=%3d, G=%3d, B=%3d, A=%3d\n", demult.R, demult.G, demult.B, demult.A)

	// Show the effect on different alpha values
	fmt.Println("\nPremultiplication with different alpha values:")
	alphas := []basics.Int8u{255, 200, 150, 100, 50, 0}
	for _, alpha := range alphas {
		test := color.NewRGBA8[color.Linear](200, 150, 100, alpha)
		original := test
		test.Premultiply()
		fmt.Printf("Alpha %3d: (%3d,%3d,%3d) → (%3d,%3d,%3d) [%.1f%% reduction]\n",
			alpha, original.R, original.G, original.B, test.R, test.G, test.B,
			(1.0-float64(test.R)/float64(original.R))*100)
	}
}

func demonstrateColorOrders() {
	fmt.Println("\n6. Color Order Types")
	fmt.Println("====================")

	// Show how different color orders work
	src := color.NewRGBA8[color.Linear](255, 128, 64, 200)
	fmt.Printf("Source Color:    R=%3d, G=%3d, B=%3d, A=%3d\n", src.R, src.G, src.B, src.A)

	// Test different blender types
	fmt.Println("\nBlender Types Available:")
	fmt.Println("- BlenderRGBA8:        Standard RGBA order")
	fmt.Println("- BlenderARGB8:        ARGB order")
	fmt.Println("- BlenderBGRA8:        BGRA order")
	fmt.Println("- BlenderABGR8:        ABGR order")
	fmt.Println("- BlenderRGBA8Pre:     Premultiplied RGBA")
	fmt.Println("- BlenderRGBA8Plain:   Plain (non-premult) RGBA")

	// Demonstrate color order differences
	rgbaOrder := color.OrderRGBA
	argbOrder := color.OrderARGB
	bgraOrder := color.OrderBGRA
	abgrOrder := color.OrderABGR

	fmt.Printf("\nColor Order Mappings:\n")
	fmt.Printf("RGBA: R=%d, G=%d, B=%d, A=%d\n", rgbaOrder.R, rgbaOrder.G, rgbaOrder.B, rgbaOrder.A)
	fmt.Printf("ARGB: A=%d, R=%d, G=%d, B=%d\n", argbOrder.A, argbOrder.R, argbOrder.G, argbOrder.B)
	fmt.Printf("BGRA: B=%d, G=%d, R=%d, A=%d\n", bgraOrder.B, bgraOrder.G, bgraOrder.R, bgraOrder.A)
	fmt.Printf("ABGR: A=%d, B=%d, G=%d, R=%d\n", abgrOrder.A, abgrOrder.B, abgrOrder.G, abgrOrder.R)

	fmt.Println("\nThis allows AGG to work with different pixel buffer formats")
	fmt.Println("commonly used in different graphics systems and platforms.")
}

// Check if this is being run as main
func init() {
	if len(os.Args) > 0 && os.Args[0] != "go" {
		// This is likely being run directly, not through go test
		// You can add any initialization here if needed
	}
}
