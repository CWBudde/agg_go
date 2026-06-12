package alphamask2

import (
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"testing"

	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	"github.com/cwbudde/agg_go/internal/path"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	"github.com/cwbudde/agg_go/internal/renderer/markers"
	outline "github.com/cwbudde/agg_go/internal/renderer/outline"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/shapes"
	"github.com/cwbudde/agg_go/internal/transform"
)

func loadPNG(t *testing.T, p string) *image.NRGBA {
	f, err := os.Open(p)
	if err != nil {
		t.Skipf("missing %s", p)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	b := img.Bounds()
	out := image.NewNRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			out.Set(x, y, img.At(x, y))
		}
	}
	return out
}

// TestStageAttribution renders the demo stage by stage and reports, for each
// pixel differing from the C++ reference, which stages touched it.
func TestStageAttribution(t *testing.T) {
	if os.Getenv("STAGE_ATTR") == "" {
		t.Skip("set STAGE_ATTR=1 to run")
	}
	const width, height = 512, 400
	ref := loadPNG(t, "../../../tests/visual/reference/cpp/examples/alpha_mask2.png")

	dst := make([]uint8, width*height*3)
	stageNames := []string{"clear+mask", "lion", "markers", "outline", "gradcircles"}
	touched := make([][]bool, len(stageNames))
	for i := range touched {
		touched[i] = make([]bool, width*height)
	}
	prev := make([]uint8, len(dst))

	snapshot := func(stage int) {
		for p := 0; p < width*height; p++ {
			if dst[p*3] != prev[p*3] || dst[p*3+1] != prev[p*3+1] || dst[p*3+2] != prev[p*3+2] {
				touched[stage][p] = true
			}
		}
		copy(prev, dst)
	}

	// Replicate RenderToBGR24 with stage snapshots.
	lionOnce.Do(initLion)
	rbuf := buffer.NewRenderingBufferU8WithData(dst, width, height, width*3)
	mainPixfLinear := pixfmt.NewPixFmtBGR24(rbuf)
	mainPixfLinearAdaptor := pixfmt.NewPixFmtRGBARendererAdaptor(mainPixfLinear)
	mainRbLinear := renderer.NewRendererBaseWithPixfmt(mainPixfLinearAdaptor)
	mainRbLinear.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	maskData := make([]uint8, width*height)
	maskBuf := buffer.NewRenderingBufferU8WithData(maskData, width, height, width)
	maskPixf := pixfmt.NewPixFmtSGray8(maskBuf)
	maskRb := renderer.NewRendererBaseWithPixfmt(maskPixf)
	maskRb.Clear(color.Gray8[color.SRGB]{V: 0, A: 255})

	ras := newRasterizer()
	sl := scanline.NewScanlineU8()
	rng := newClibcRand(1432)

	for i := 0; i < 10; i++ {
		ry := float64(rng.randN(100) + 20)
		rx := float64(rng.randN(100) + 20)
		y := float64(rng.randN(height))
		x := float64(rng.randN(width))
		ell := shapes.NewEllipseWithParams(x, y, rx, ry, 100, false)
		ras.Reset()
		ras.AddPath(&ellipseVS{e: ell}, 0)
		a := uint8(rng.randAnd(127) + 128)
		v := uint8(rng.randAnd(127) + 128)
		renscan.RenderScanlinesAASolid(ras, sl, maskRb, color.Gray8[color.SRGB]{V: v, A: a})
	}
	snapshot(0)

	mask := pixfmt.NewAMaskNoClipU8WithBuffer(maskBuf, 1, 0, pixfmt.OneComponentMaskU8{})
	amaskAdaptorLinear := pixfmt.NewPixFmtAMaskAdaptor(mainPixfLinearAdaptor, mask)
	rbAMaskLinear := renderer.NewRendererBaseWithPixfmt(amaskAdaptorLinear)

	mtx := transform.NewTransAffine()
	mtx.Multiply(transform.NewTransAffineTranslation(-lionBaseDX, -lionBaseDY))
	mtx.Multiply(transform.NewTransAffineScaling(1))
	mtx.Multiply(transform.NewTransAffineRotation(math.Pi))
	mtx.Multiply(transform.NewTransAffineSkewing(0, 0))
	mtx.Multiply(transform.NewTransAffineTranslation(float64(width)/2, float64(height)/2))

	pathVS := path.NewPathStorageStlVertexSourceAdapter(lionData.Path)
	transVS := conv.NewConvTransform(pathVS, mtx)
	rasVS := conv.NewRasterizerVertexSourceAdapter(transVS)
	renSolid := renscan.NewRendererScanlineAASolidWithRenderer(rbAMaskLinear)
	renscan.RenderAllPaths(ras, sl, renSolid, rasVS, &lionData, &lionData, lionData.NPaths)
	snapshot(1)

	// Per-object attribution: render markers/outline/gradcircles one object at a
	// time by replicating the loops with iteration-level snapshots.
	objOf := make([]string, width*height)
	objSnapshot := func(label string) {
		for p := 0; p < width*height; p++ {
			if dst[p*3] != prev[p*3] || dst[p*3+1] != prev[p*3+1] || dst[p*3+2] != prev[p*3+2] {
				touched[2][p] = true
				if objOf[p] != "" {
					objOf[p] += "|"
				}
				objOf[p] += label
			}
		}
		copy(prev, dst)
	}

	m := newTestMarkers(rbAMaskLinear)
	for i := 0; i < 50; i++ {
		lc := srgbaRandRTL(rng, 0x7F)
		fc := srgbaRandRTL(rng, 0x7F)
		m.setColors(lc, fc)
		y2 := rng.randN(height)
		x2 := rng.randN(width)
		y1 := rng.randN(height)
		x1 := rng.randN(width)
		m.line(x1, y1, x2, y2)
		objSnapshot(fmt.Sprintf("mline%d(%d,%d-%d,%d)", i, x1, y1, x2, y2))
		mt := rng.randN(m.endOfMarkers())
		radius := rng.randN(10) + 5
		y := rng.randN(height)
		x := rng.randN(width)
		m.marker(x, y, radius, mt)
		objSnapshot(fmt.Sprintf("marker%d-t%d-r%d(%d,%d)", i, mt, radius, x, y))
	}

	renderOutlineLinesAttr(rbAMaskLinear, rng, width, height, objSnapshot)

	renderGradientCircles(ras, sl, rbAMaskLinear, rng, width, height)
	snapshot(4)

	// Compare to the reference: ref row y = buffer row height-1-y, sRGB-encoded.
	_ = stageNames
	counts := map[string]int{}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if x < 165 && y > 378 {
				continue // slider region, drawn in main.go not here
			}
			s := ((height-1-y)*width + x) * 3
			c := color.ConvertRGB8LinearToSRGB(color.RGB8[color.Linear]{
				R: dst[s+2], G: dst[s+1], B: dst[s+0],
			})
			r := ref.NRGBAAt(x, y)
			if c.R != r.R || c.G != r.G || c.B != r.B {
				p := (height-1-y)*width + x
				key := objOf[p]
				if key == "" {
					key = "no-overlay-object"
				}
				counts[key]++
				fmt.Printf("(%3d,%3d) ref=%d,%d,%d gen=%d,%d,%d [%s]\n",
					x, y, r.R, r.G, r.B, c.R, c.G, c.B, key)
			}
		}
	}
	fmt.Println("=== counts per object ===")
	for k, v := range counts {
		fmt.Println(v, k)
	}
}

type testMarkers struct {
	m *markers.RendererMarkers[*renderer.RendererBase[*pixfmt.PixFmtAMaskAdaptor[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]], color.RGBA8[color.Linear]]
}

func newTestMarkers(rb *renderer.RendererBase[*pixfmt.PixFmtAMaskAdaptor[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]]) *testMarkers {
	return &testMarkers{m: markers.NewRendererMarkers(rb)}
}

func (t *testMarkers) setColors(lc, fc color.RGBA8[color.Linear]) {
	t.m.LineColor(lc)
	t.m.FillColor(fc)
}

func (t *testMarkers) line(x1, y1, x2, y2 int) {
	t.m.Line(t.m.Coord(float64(x1)), t.m.Coord(float64(y1)), t.m.Coord(float64(x2)), t.m.Coord(float64(y2)), false)
}

func (t *testMarkers) marker(x, y, r, mt int) {
	t.m.Marker(x, y, r, markers.MarkerType(mt))
}

func (t *testMarkers) endOfMarkers() int { return int(markers.EndOfMarkers) }

func renderOutlineLinesAttr(
	rbAMask *renderer.RendererBase[*pixfmt.PixFmtAMaskAdaptor[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]],
	rng *clibcRand,
	width, height int,
	objSnapshot func(string),
) {
	profile := outline.NewLineProfileAA()
	profile.Width(5.0)

	renOutline := outline.NewRendererOutlineAA[*outlineBaseAdapter[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](
		&outlineBaseAdapter[color.RGBA8[color.Linear]]{renBase: rbAMask},
		profile,
	)
	rasOutline := rasterizer.NewRasterizerOutlineAA[*outlineAAAdapter[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](
		&outlineAAAdapter[color.RGBA8[color.Linear]]{ren: renOutline},
	)
	rasOutline.SetRoundCap(true)

	for i := 0; i < 50; i++ {
		renOutline.Color(srgbaRandRTL(rng, 0x7F))
		y1 := rng.randN(height)
		x1 := rng.randN(width)
		rasOutline.MoveToD(float64(x1), float64(y1))
		y2 := rng.randN(height)
		x2 := rng.randN(width)
		rasOutline.LineToD(float64(x2), float64(y2))
		rasOutline.Render(false)
		objSnapshot(fmt.Sprintf("oline%d(%d,%d-%d,%d)", i, x1, y1, x2, y2))
	}
}
