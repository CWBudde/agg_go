// Package main ports AGG's rasterizer_compound.cpp demo.
package main

import (
	"fmt"

	agg "agg_go"
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/conv"
	"agg_go/internal/path"
	"agg_go/internal/rasterizer"
	"agg_go/internal/scanline"
	"agg_go/internal/shapes"
	"agg_go/internal/transform"
)

type compoundNoClip struct{ x1, y1 float64 }

func (c *compoundNoClip) ResetClipping()                 {}
func (c *compoundNoClip) ClipBox(x1, y1, x2, y2 float64) {}
func (c *compoundNoClip) MoveTo(x, y float64)            { c.x1, c.y1 = x, y }
func (c *compoundNoClip) LineTo(outline *rasterizer.RasterizerCellsAAStyled, x, y float64) {
	outline.Line(
		basics.IRound(c.x1*basics.PolySubpixelScale), basics.IRound(c.y1*basics.PolySubpixelScale),
		basics.IRound(x*basics.PolySubpixelScale), basics.IRound(y*basics.PolySubpixelScale),
	)
	c.x1, c.y1 = x, y
}

type rcStyleHandler struct {
	styles []color.RGBA8[color.Linear]
}

func (h *rcStyleHandler) IsSolid(style int) bool { return true }
func (h *rcStyleHandler) Color(style int) color.RGBA8[color.Linear] {
	if style < 0 || style >= len(h.styles) {
		return color.RGBA8[color.Linear]{}
	}
	return h.styles[style]
}
func (h *rcStyleHandler) GenerateSpan(colors []color.RGBA8[color.Linear], x, y, length, style int) {}

type rcSLAdapter struct{ sl *scanline.ScanlineU8 }

func (a *rcSLAdapter) ResetSpans()                      { a.sl.ResetSpans() }
func (a *rcSLAdapter) AddCell(x int, c basics.Int8u)    { a.sl.AddCell(x, uint(c)) }
func (a *rcSLAdapter) AddSpan(x, l int, c basics.Int8u) { a.sl.AddSpan(x, l, uint(c)) }
func (a *rcSLAdapter) Finalize(y int)                   { a.sl.Finalize(y) }
func (a *rcSLAdapter) NumSpans() int                    { return a.sl.NumSpans() }

type rcConvVertexSource interface {
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
}

type rcConvVSAdapter struct {
	vs rcConvVertexSource
}

func (a *rcConvVSAdapter) Rewind(pathID uint32) {
	a.vs.Rewind(uint(pathID))
}

func (a *rcConvVSAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.vs.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

type rcEllipseVSAdapter struct {
	ell *shapes.Ellipse
}

func (a *rcEllipseVSAdapter) Rewind(pathID uint32) { a.ell.Rewind(pathID) }
func (a *rcEllipseVSAdapter) Vertex(x, y *float64) uint32 {
	return uint32(a.ell.Vertex(x, y))
}

type rcEllipseConvAdapter struct {
	ell *shapes.Ellipse
}

func (a *rcEllipseConvAdapter) Rewind(pathID uint) {
	a.ell.Rewind(uint32(pathID))
}

func (a *rcEllipseConvAdapter) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = a.ell.Vertex(&x, &y)
	return x, y, cmd
}

func composeCompoundPath(ps *path.PathStorageStl) {
	ps.RemoveAll()
	ps.MoveTo(28.47, 6.45)
	ps.Curve3(21.58, 1.12, 19.82, 0.29)
	ps.Curve3(17.19, -0.93, 14.21, -0.93)
	ps.Curve3(9.57, -0.93, 6.57, 2.25)
	ps.Curve3(3.56, 5.42, 3.56, 10.60)
	ps.Curve3(3.56, 13.87, 5.03, 16.26)
	ps.Curve3(7.03, 19.58, 11.99, 22.51)
	ps.Curve3(16.94, 25.44, 28.47, 29.64)
	ps.LineTo(28.47, 31.40)
	ps.Curve3(28.47, 38.09, 26.34, 40.58)
	ps.Curve3(24.22, 43.07, 20.17, 43.07)
	ps.Curve3(17.09, 43.07, 15.28, 41.41)
	ps.Curve3(13.43, 39.75, 13.43, 37.60)
	ps.LineTo(13.53, 34.77)
	ps.Curve3(13.53, 32.52, 12.38, 31.30)
	ps.Curve3(11.23, 30.08, 9.38, 30.08)
	ps.Curve3(7.57, 30.08, 6.42, 31.35)
	ps.Curve3(5.27, 32.62, 5.27, 34.81)
	ps.Curve3(5.27, 39.01, 9.57, 42.53)
	ps.Curve3(13.87, 46.04, 21.63, 46.04)
	ps.Curve3(27.59, 46.04, 31.40, 44.04)
	ps.Curve3(34.28, 42.53, 35.64, 39.31)
	ps.Curve3(36.52, 37.21, 36.52, 30.71)
	ps.LineTo(36.52, 15.53)
	ps.Curve3(36.52, 9.13, 36.77, 7.69)
	ps.Curve3(37.01, 6.25, 37.57, 5.76)
	ps.Curve3(38.13, 5.27, 38.87, 5.27)
	ps.Curve3(39.65, 5.27, 40.23, 5.62)
	ps.Curve3(41.26, 6.25, 44.19, 9.18)
	ps.LineTo(44.19, 6.45)
	ps.Curve3(38.72, -0.88, 33.74, -0.88)
	ps.Curve3(31.35, -0.88, 29.93, 0.78)
	ps.Curve3(28.52, 2.44, 28.47, 6.45)
	ps.ClosePolygon(basics.PathFlagsNone)

	ps.MoveTo(28.47, 9.62)
	ps.LineTo(28.47, 26.66)
	ps.Curve3(21.09, 23.73, 18.95, 22.51)
	ps.Curve3(15.09, 20.36, 13.43, 18.02)
	ps.Curve3(11.77, 15.67, 11.77, 12.89)
	ps.Curve3(11.77, 9.38, 13.87, 7.06)
	ps.Curve3(15.97, 4.74, 18.70, 4.74)
	ps.Curve3(22.41, 4.74, 28.47, 9.62)
	ps.ClosePolygon(basics.PathFlagsNone)
}

func main() {
	const (
		w      = 440
		h      = 330
		width  = 10.0
		alpha1 = 1.0
		alpha2 = 1.0
		alpha3 = 1.0
		alpha4 = 1.0
		out    = "rasterizer_compound.png"
	)

	ctx := agg.NewContext(w, h)
	a := ctx.GetAgg2D()

	img := ctx.GetImage()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			t := float64(x) / float64(w-1)
			i := (y*w + x) * 4
			img.Data[i+0] = uint8((1.0 - t) * 255)
			img.Data[i+1] = 255
			img.Data[i+2] = uint8(t * 255)
			img.Data[i+3] = 255
		}
	}

	ctx.SetColor(agg.NewColor(0, 100, 0, 255))
	a.ResetPath()
	a.MoveTo(0, 0)
	a.LineTo(w, 0)
	a.LineTo(w, h)
	a.ClosePolygon()
	a.DrawPath(agg.FillOnly)

	ctx.SetColor(agg.NewColor(0, 100, 100, 255))
	a.ResetPath()
	a.MoveTo(0, 0)
	a.LineTo(0, h)
	a.LineTo(w, 0)
	a.ClosePolygon()
	a.DrawPath(agg.FillOnly)

	ps := path.NewPathStorageStl()
	composeCompoundPath(ps)
	psAdapter := path.NewPathStorageStlVertexSourceAdapter(ps)
	mtx := transform.NewTransAffine()
	mtx.Multiply(transform.NewTransAffineScaling(4.0))
	mtx.Multiply(transform.NewTransAffineTranslation(150, 100))
	transPath := conv.NewConvTransform(psAdapter, mtx)
	curve := conv.NewConvCurve(transPath)
	stroke := conv.NewConvStroke(curve)
	stroke.SetWidth(width)

	ell := shapes.NewEllipseWithParams(220.0, 180.0, 120.0, 10.0, 128, false)
	ellStroke := conv.NewConvStroke(&rcEllipseConvAdapter{ell: ell})
	ellStroke.SetWidth(width / 2.0)

	styles := []color.RGBA8[color.Linear]{
		{R: 0, G: 0, B: 255, A: 255},
		{R: 143, G: 90, B: 6, A: 255},
		{R: 51, G: 0, B: 151, A: 255},
		{R: 255, G: 0, B: 108, A: 255},
	}
	styles[3].Opacity(alpha1)
	styles[2].Opacity(alpha2)
	styles[1].Opacity(alpha3)
	styles[0].Opacity(alpha4)
	for i := range styles {
		styles[i].Premultiply()
	}

	rasc := rasterizer.NewRasterizerCompoundAA(&compoundNoClip{})
	rasc.LayerOrder(basics.LayerDirect)
	rasc.Styles(3, -1)
	rasc.AddPath(&rcConvVSAdapter{vs: ellStroke}, 0)
	rasc.Styles(2, -1)
	rasc.AddPath(&rcEllipseVSAdapter{ell: ell}, 0)
	rasc.Styles(1, -1)
	rasc.AddPath(&rcConvVSAdapter{vs: stroke}, 0)
	rasc.Styles(0, -1)
	rasc.AddPath(&rcConvVSAdapter{vs: curve}, 0)

	rasc.Sort()
	if !rasc.RewindScanlines() {
		return
	}

	minX := rasc.MinX()
	maxX := rasc.MaxX()
	slAA := scanline.NewScanlineU8()
	slBin := scanline.NewScanlineU8()
	slAA.Reset(minX, maxX)
	slBin.Reset(minX, maxX)
	adAA := &rcSLAdapter{sl: slAA}
	adBin := &rcSLAdapter{sl: slBin}
	styleHandler := &rcStyleHandler{styles: styles}

	length := maxX - minX + 2
	colorSpan := make([]color.RGBA8[color.Linear], length*2)
	mixBuffer := colorSpan[length:]

	for {
		numStyles := rasc.SweepStyles()
		if numStyles == 0 {
			break
		}
		if numStyles == 1 {
			if rasc.SweepScanline(adAA, 0) {
				c := styleHandler.Color(int(rasc.Style(0)))
				y := slAA.Y()
				for _, sp := range slAA.Spans() {
					for j := 0; j < int(sp.Len); j++ {
						x := int(sp.X) + j
						i := (y*w + x) * 4
						cover := float64(sp.Covers[j]) / 255.0
						inv := 1.0 - cover
						img.Data[i+0] = uint8(float64(c.R)*cover + float64(img.Data[i+0])*inv)
						img.Data[i+1] = uint8(float64(c.G)*cover + float64(img.Data[i+1])*inv)
						img.Data[i+2] = uint8(float64(c.B)*cover + float64(img.Data[i+2])*inv)
						img.Data[i+3] = 255
					}
				}
			}
		} else if rasc.SweepScanline(adBin, -1) {
			y := slBin.Y()
			for _, sp := range slBin.Spans() {
				for j := 0; j < int(sp.Len); j++ {
					mixBuffer[int(sp.X)-minX+j] = color.RGBA8[color.Linear]{}
				}
			}
			for i := uint32(0); i < numStyles; i++ {
				style := int(rasc.Style(i))
				if rasc.SweepScanline(adAA, int(i)) {
					c := styleHandler.Color(style)
					for _, sp := range slAA.Spans() {
						for j := 0; j < int(sp.Len); j++ {
							ptr := &mixBuffer[int(sp.X)-minX+j]
							ptr.AddWithCover(c, sp.Covers[j])
						}
					}
				}
			}
			for _, sp := range slBin.Spans() {
				for j := 0; j < int(sp.Len); j++ {
					x := int(sp.X) + j
					i := (y*w + x) * 4
					c := mixBuffer[int(sp.X)-minX+j]
					img.Data[i+0] = uint8(c.R)
					img.Data[i+1] = uint8(c.G)
					img.Data[i+2] = uint8(c.B)
					img.Data[i+3] = 255
				}
			}
		}
	}

	if err := img.SaveToPNG(out); err != nil {
		fmt.Printf("error writing %s: %v\n", out, err)
		return
	}
	fmt.Printf("wrote %s (%dx%d)\n", out, w, h)
}
