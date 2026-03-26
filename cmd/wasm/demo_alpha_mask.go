// Port of AGG C++ alpha_mask.cpp – alpha-masked lion rendering.
//
// Generates a grayscale alpha mask from random ellipses, then renders the
// lion through it so only the mask's bright regions show the lion colours.
package main

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	liondemo "github.com/MeKo-Christian/agg_go/internal/demo/lion"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

var (
	amAlphaMaskBuf *buffer.RenderingBuffer[uint8]
	amAlphaMask    *pixfmt.AMaskNoClipU8
	amLionAngle    = 0.0
	amLionScale    = 1.0
	amLionSkewX    = 0.0
	amLionSkewY    = 0.0
)

func generateAlphaMask(w, h int) {
	if amAlphaMaskBuf == nil || amAlphaMaskBuf.Width() != w || amAlphaMaskBuf.Height() != h {
		data := make([]uint8, w*h)
		amAlphaMaskBuf = buffer.NewRenderingBufferWithData[uint8](data, w, h, w)
	}

	// Create a grayscale pixel format for the mask buffer
	maskPixf := pixfmt.NewPixFmtSGray8(amAlphaMaskBuf)
	maskRb := renderer.NewRendererBaseWithPixfmt(maskPixf)

	maskRb.Clear(color.Gray8[color.SRGB]{V: 0, A: 255})

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineP8()

	for i := 0; i < 10; i++ {
		cx := float64(rand.Intn(w))
		cy := float64(rand.Intn(h))
		rx := float64(rand.Intn(100) + 20)
		ry := float64(rand.Intn(100) + 20)

		ras.Reset()
		ell := shapes.NewEllipseWithParams(cx, cy, rx, ry, 100, false)
		ell.Rewind(0)
		for {
			var x, y float64
			cmd := ell.Vertex(&x, &y)
			if basics.IsStop(cmd) {
				break
			}
			ras.AddVertex(x, y, uint32(cmd))
		}

		c := uint8(rand.Intn(256))
		opacity := uint8(rand.Intn(256))
		gray := color.Gray8[color.SRGB]{V: c, A: opacity}

		renscan.RenderScanlinesAASolid(ras, sl, maskRb, gray)
	}

	// Create the alpha mask
	maskFunc := pixfmt.OneComponentMaskU8{}
	amAlphaMask = pixfmt.NewAMaskNoClipU8WithBuffer(amAlphaMaskBuf, 1, 0, maskFunc)
}

func drawAlphaMaskDemo() {
	img := ctx.GetImage()
	w, h := img.Width(), img.Height()
	if amAlphaMask == nil {
		generateAlphaMask(w, h)
	}

	if lionData == nil {
		ld := liondemo.Parse()
		lionData = &ld
	}

	// Attach the image buffer.
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, w, h, img.Stride())

	// White background.
	imgPixf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	mainRb := renderer.NewRendererBaseWithPixfmt(imgPixf)
	mainRb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	// Draw checkered background using direct pixel rendering.
	checkColor := color.RGBA8[color.Linear]{R: 0xdf, G: 0xdf, B: 0xdf, A: 0xff}
	for y := 0; y < h; y += 8 {
		for x := ((y >> 3) & 1) << 3; x < w; x += 16 {
			x2 := x + 8
			if x2 > w {
				x2 = w
			}
			y2 := y + 8
			if y2 > h {
				y2 = h
			}
			mainRb.CopyBar(x, y, x2-1, y2-1, checkColor)
		}
	}

	// Build transform for the lion.
	// Compute bounding box from lion path data.
	minX, minY := 1e9, 1e9
	maxX, maxY := -1e9, -1e9
	for idx := uint(0); idx < lionData.Path.TotalVertices(); idx++ {
		x, y, cmd := lionData.Path.Vertex(idx)
		pathCmd := basics.PathCommand(cmd)
		if basics.IsMoveTo(pathCmd) || basics.IsLineTo(pathCmd) {
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x > maxX {
				maxX = x
			}
			if y > maxY {
				maxY = y
			}
		}
	}
	baseDX := (maxX - minX) / 2.0
	baseDY := (maxY - minY) / 2.0

	mtx := transform.NewTransAffine()
	mtx.Multiply(transform.NewTransAffineTranslation(-baseDX, -baseDY))
	mtx.Multiply(transform.NewTransAffineScaling(amLionScale))
	mtx.Multiply(transform.NewTransAffineRotation(amLionAngle + math.Pi))
	mtx.Multiply(transform.NewTransAffineSkewing(amLionSkewX/1000.0, amLionSkewY/1000.0))
	mtx.Multiply(transform.NewTransAffineTranslation(float64(w)/2, float64(h)/2))

	// Render lion through the alpha mask.
	amaskAdaptor := pixfmt.NewPixFmtAMaskAdaptor(imgPixf, amAlphaMask)
	rbAMask := renderer.NewRendererBaseWithPixfmt(amaskAdaptor)

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineP8()

	pathVS := path.NewPathStorageStlVertexSourceAdapter(lionData.Path)
	transVS := conv.NewConvTransform(pathVS, mtx)
	rasVS := conv.NewRasterizerVertexSourceAdapter(transVS)
	renSolid := renscan.NewRendererScanlineAASolidWithRenderer(rbAMask)
	renscan.RenderAllPaths(ras, sl, renSolid, rasVS, lionData, lionData, lionData.NPaths)

	applyLinearToSRGB(img)

	logStatus(fmt.Sprintf("Alpha Mask Demo: Scale=%.2f, Angle=%.2f", amLionScale, amLionAngle))
}

func handleAlphaMaskMouseDown(x, y float64, flags int) bool {
	w, h := ctx.GetImage().Width(), ctx.GetImage().Height()
	dx := x - float64(w)/2
	dy := y - float64(h)/2
	amLionAngle = math.Atan2(dy, dx)
	amLionScale = math.Sqrt(dy*dy+dx*dx) / 100.0
	return true
}

func handleAlphaMaskRightMouseDown(x, y float64) bool {
	amLionSkewX = x
	amLionSkewY = y
	return true
}
