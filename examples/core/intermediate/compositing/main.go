// Port of AGG C++ compositing.cpp – compositing modes with a checkerboard
// background, an image blend, a shaded circle, and a rounded-rect source shape.
package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	icol "github.com/MeKo-Christian/agg_go/internal/color"
	ctrlbase "github.com/MeKo-Christian/agg_go/internal/ctrl"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/rbox"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/slider"
)

const (
	frameWidth  = 600
	frameHeight = 400
)

type controlVertexSourceAdapter struct {
	ctrl ctrlbase.Ctrl[icol.RGBA]
}

func (a *controlVertexSourceAdapter) Rewind(pathID uint32) {
	a.ctrl.Rewind(uint(pathID))
}

func (a *controlVertexSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

type demo struct {
	srcAlpha *slider.SliderCtrl
	dstAlpha *slider.SliderCtrl
	compOp   *rbox.RboxCtrl[icol.RGBA]
	bgImg    *agg.Image
}

func srgba8(r, g, b, a uint8) agg.Color {
	c := icol.ConvertRGBA8SRGBToLinear(icol.RGBA8[icol.SRGB]{
		R: r,
		G: g,
		B: b,
		A: a,
	})
	return agg.NewColor(c.R, c.G, c.B, c.A)
}

func toAggColor(c icol.RGBA) agg.Color {
	clamp := func(v float64) uint8 {
		switch {
		case v <= 0:
			return 0
		case v >= 1:
			return 255
		default:
			return uint8(v*255.0 + 0.5)
		}
	}
	return agg.NewColor(clamp(c.R), clamp(c.G), clamp(c.B), clamp(c.A))
}

func compositingAssetPath() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join("examples", "shared", "art", "compositing.bmp")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "shared", "art", "compositing.bmp")
}

func newDemo() *demo {
	srcAlpha := slider.NewSliderCtrl(5, 5, 400, 11, false)
	srcAlpha.SetLabel("Src Alpha=%.2f")
	srcAlpha.SetRange(0, 1)
	srcAlpha.SetValue(1)

	dstAlpha := slider.NewSliderCtrl(5, 20, 400, 26, false)
	dstAlpha.SetLabel("Dst Alpha=%.2f")
	dstAlpha.SetRange(0, 1)
	dstAlpha.SetValue(1)

	compOp := rbox.NewDefaultRboxCtrl(420, 5, 590, 340, false)
	compOp.SetTextSize(6.8, 0)
	compOp.AddItem("clear")
	compOp.AddItem("src")
	compOp.AddItem("dst")
	compOp.AddItem("src-over")
	compOp.AddItem("dst-over")
	compOp.AddItem("src-in")
	compOp.AddItem("dst-in")
	compOp.AddItem("src-out")
	compOp.AddItem("dst-out")
	compOp.AddItem("src-atop")
	compOp.AddItem("dst-atop")
	compOp.AddItem("xor")
	compOp.AddItem("plus")
	compOp.AddItem("multiply")
	compOp.AddItem("screen")
	compOp.AddItem("overlay")
	compOp.AddItem("darken")
	compOp.AddItem("lighten")
	compOp.AddItem("color-dodge")
	compOp.AddItem("color-burn")
	compOp.AddItem("hard-light")
	compOp.AddItem("soft-light")
	compOp.AddItem("difference")
	compOp.AddItem("exclusion")
	compOp.SetCurItem(3)

	bgImg, err := loadBMPImage(compositingAssetPath(), true)
	if err != nil {
		panic(fmt.Errorf("load compositing background image: %w", err))
	}

	return &demo{
		srcAlpha: srcAlpha,
		dstAlpha: dstAlpha,
		compOp:   compOp,
		bgImg:    bgImg,
	}
}

func blendModeForIndex(idx int) agg.BlendMode {
	modes := []agg.BlendMode{
		agg.BlendClear,
		agg.BlendSrc,
		agg.BlendDst,
		agg.BlendSrcOver,
		agg.BlendDstOver,
		agg.BlendSrcIn,
		agg.BlendDstIn,
		agg.BlendSrcOut,
		agg.BlendDstOut,
		agg.BlendSrcAtop,
		agg.BlendDstAtop,
		agg.BlendXor,
		agg.BlendAdd,
		agg.BlendMultiply,
		agg.BlendScreen,
		agg.BlendOverlay,
		agg.BlendDarken,
		agg.BlendLighten,
		agg.BlendColorDodge,
		agg.BlendColorBurn,
		agg.BlendHardLight,
		agg.BlendSoftLight,
		agg.BlendDifference,
		agg.BlendExclusion,
	}
	if idx < 0 || idx >= len(modes) {
		return agg.BlendSrcOver
	}
	return modes[idx]
}

func loadBMPImage(filename string, flipY bool) (*agg.Image, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var fileHeader struct {
		Type      uint16
		Size      uint32
		Reserved1 uint16
		Reserved2 uint16
		OffBits   uint32
	}
	if err := binary.Read(f, binary.LittleEndian, &fileHeader); err != nil {
		return nil, err
	}
	if fileHeader.Type != 0x4D42 {
		return nil, errors.New("not a BMP file")
	}

	var infoHeader struct {
		Size          uint32
		Width         int32
		Height        int32
		Planes        uint16
		BitCount      uint16
		Compression   uint32
		SizeImage     uint32
		XPelsPerMeter int32
		YPelsPerMeter int32
		ClrUsed       uint32
		ClrImportant  uint32
	}
	if err := binary.Read(f, binary.LittleEndian, &infoHeader); err != nil {
		return nil, err
	}
	if infoHeader.Planes != 1 {
		return nil, fmt.Errorf("unsupported BMP planes: %d", infoHeader.Planes)
	}
	if infoHeader.Compression != 0 {
		return nil, fmt.Errorf("unsupported BMP compression: %d", infoHeader.Compression)
	}
	if infoHeader.BitCount != 24 && infoHeader.BitCount != 32 {
		return nil, fmt.Errorf("unsupported BMP bit depth: %d", infoHeader.BitCount)
	}

	width := int(infoHeader.Width)
	height := int(infoHeader.Height)
	if width <= 0 || height == 0 {
		return nil, fmt.Errorf("invalid BMP dimensions: %dx%d", width, height)
	}
	if height < 0 {
		height = -height
	}
	if _, err := f.Seek(int64(fileHeader.OffBits), io.SeekStart); err != nil {
		return nil, err
	}

	rowStride := ((width*int(infoHeader.BitCount) + 31) / 32) * 4
	rowData := make([]byte, rowStride)
	buf := make([]uint8, width*height*4)

	for y := 0; y < height; y++ {
		if _, err := io.ReadFull(f, rowData); err != nil {
			return nil, err
		}
		dstY := y
		flipVertical := infoHeader.Height > 0
		if flipVertical != flipY {
			dstY = height - 1 - y
		}
		for x := 0; x < width; x++ {
			src := x * int(infoHeader.BitCount) / 8
			dst := (dstY*width + x) * 4

			srgb := icol.RGBA8[icol.SRGB]{
				R: rowData[src+2],
				G: rowData[src+1],
				B: rowData[src+0],
				A: 255,
			}
			if infoHeader.BitCount == 32 {
				srgb.A = rowData[src+3]
			}

			linear := icol.ConvertRGBA8SRGBToLinear(srgb)
			buf[dst+0] = linear.R
			buf[dst+1] = linear.G
			buf[dst+2] = linear.B
			buf[dst+3] = linear.A
		}
	}

	return agg.NewImage(buf, width, height, width*4), nil
}

func drawCheckerboard(a *agg.Agg2D, width, height int) {
	a.BlendMode(agg.BlendAlpha)
	a.FillColor(agg.White)
	a.NoLine()
	a.ResetPath()
	a.MoveTo(0, 0)
	a.LineTo(float64(width), 0)
	a.LineTo(float64(width), float64(height))
	a.LineTo(0, float64(height))
	a.ClosePolygon()
	a.DrawPath(agg.FillOnly)

	for y := 0; y < height; y += 8 {
		xStart := 0
		if ((y >> 3) & 1) == 1 {
			xStart = 8
		}
		for x := xStart; x < width; x += 16 {
			a.FillColor(srgba8(0xDF, 0xDF, 0xDF, 255))
			a.NoLine()
			a.ResetPath()
			a.MoveTo(float64(x), float64(y))
			a.LineTo(float64(x+7), float64(y))
			a.LineTo(float64(x+7), float64(y+7))
			a.LineTo(float64(x), float64(y+7))
			a.ClosePolygon()
			a.DrawPath(agg.FillOnly)
		}
	}
}

func circle(a *agg.Agg2D, c1, c2 agg.Color, x1, y1, x2, y2, shadowAlpha float64) {
	cx := (x1 + x2) * 0.5
	cy := (y1 + y2) * 0.5
	r := math.Hypot(x2-x1, y2-y1) * 0.5

	a.BlendMode(agg.BlendAlpha)
	a.FillColor(agg.RGBA(0.6, 0.6, 0.6, 0.7*shadowAlpha))
	a.NoLine()
	a.FillCircle(cx+5, cy-3, r)

	a.FillLinearGradient(x1, y1, x2, y2, c1, c2, 1.0)
	a.NoLine()
	a.FillCircle(cx, cy, r)
}

func srcShape(a *agg.Agg2D, c1, c2 agg.Color, x1, y1, x2, y2 float64) {
	a.FillLinearGradient(x1, y1, x2, y2, c1, c2, 1.0)
	a.NoLine()
	a.RoundedRect(x1, y1, x2, y2, 40)
	a.DrawPath(agg.FillOnly)
}

func renderCtrl(a *agg.Agg2D, c ctrlbase.Ctrl[icol.RGBA]) {
	ras := a.GetInternalRasterizer()
	adapter := &controlVertexSourceAdapter{ctrl: c}
	for pathID := uint(0); pathID < c.NumPaths(); pathID++ {
		ras.Reset()
		ras.AddPath(adapter, uint32(pathID))
		a.RenderRasterizerWithColor(toAggColor(c.Color(pathID)))
	}
}

func (d *demo) Render(img *agg.Image) {
	ctx := agg.NewContextForImage(img)
	if ctx == nil {
		return
	}

	a := ctx.GetAgg2D()
	a.ResetTransformations()
	a.BlendMode(agg.BlendSrcOver)
	ctx.Clear(agg.White)

	drawCheckerboard(a, frameWidth, frameHeight)

	if d.bgImg != nil {
		_ = a.BlendImageSimple(d.bgImg, 250, 180, uint(d.dstAlpha.Value()*255.0+0.5))
	}

	circle(
		a,
		srgba8(0xFD, 0xF0, 0x6F, uint8(d.dstAlpha.Value()*255)),
		srgba8(0xFE, 0x9F, 0x34, uint8(d.dstAlpha.Value()*255)),
		70*3, 100+24*3, 37*3, 100+79*3,
		d.dstAlpha.Value(),
	)

	a.BlendMode(blendModeForIndex(d.compOp.CurItem()))
	srcShape(
		a,
		srgba8(0x7F, 0xC1, 0xFF, uint8(d.srcAlpha.Value()*255)),
		srgba8(0x05, 0x00, 0x5F, uint8(d.srcAlpha.Value()*255)),
		300+50, 100+24*3, 107+50, 100+79*3,
	)

	a.BlendMode(agg.BlendSrcOver)
	renderCtrl(a, d.srcAlpha)
	renderCtrl(a, d.dstAlpha)
	renderCtrl(a, d.compOp)
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	if btn.Left {
		if d.srcAlpha.OnMouseButtonDown(fx, fy) || d.dstAlpha.OnMouseButtonDown(fx, fy) || d.compOp.OnMouseButtonDown(fx, fy) {
			return true
		}
	}
	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	if d.srcAlpha.OnMouseMove(fx, fy, btn.Left) || d.dstAlpha.OnMouseMove(fx, fy, btn.Left) || d.compOp.OnMouseMove(fx, fy, btn.Left) {
		return true
	}
	return false
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	fx, fy := float64(x), float64(y)
	if d.srcAlpha.OnMouseButtonUp(fx, fy) || d.dstAlpha.OnMouseButtonUp(fx, fy) || d.compOp.OnMouseButtonUp(fx, fy) {
		return true
	}
	return false
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:                 "AGG Example. Compositing Modes",
		Width:                 frameWidth,
		Height:                frameHeight,
		FlipY:                 true,
		EncodeLinearRGBToSRGB: true,
	}, newDemo())
}
