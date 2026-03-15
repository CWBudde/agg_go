package gpctest

import (
	"fmt"
	"math"
	"time"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/demo/aggshapes"
	"github.com/MeKo-Christian/agg_go/internal/gpc"
	"github.com/MeKo-Christian/agg_go/internal/path"
)

type Config struct {
	Scene     int
	Operation int
	CenterX   float64
	CenterY   float64
}

const (
	referenceWidth  = 655.0
	referenceHeight = 520.0
)

type pt struct {
	x float64
	y float64
}

type contour []pt

func Draw(ctx *agg.Context, cfg Config) {
	w := float64(ctx.GetImage().Width())
	h := float64(ctx.GetImage().Height())
	frameOffX := (w - referenceWidth) * 0.5
	frameOffY := (h - referenceHeight) * 0.5
	if math.IsNaN(cfg.CenterX) || math.IsNaN(cfg.CenterY) {
		cfg.CenterX = w * 0.5
		cfg.CenterY = h * 0.5
	}

	cfg.Scene = clampInt(cfg.Scene, 0, 4)
	cfg.Operation = clampInt(cfg.Operation, 0, 5)
	cfg.CenterX = cfg.CenterX - frameOffX
	cfg.CenterY = referenceHeight - (cfg.CenterY - frameOffY)

	a, b := buildScene(cfg, referenceWidth, referenceHeight)
	a = transformContours(mirrorContoursY(a, referenceHeight), 0, 0, 1, 1, frameOffX, frameOffY)
	b = transformContours(mirrorContoursY(b, referenceHeight), 0, 0, 1, 1, frameOffX, frameOffY)

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()
	agg2d.ClearAll(agg.White)

	drawContours(agg2d, a, sceneAColor(cfg.Scene), sceneALineColor(cfg.Scene))
	drawContours(agg2d, b, sceneBColor(cfg.Scene), agg.Transparent)

	if cfg.Operation == 0 {
		return
	}

	clipStart := time.Now()
	subject, clip := toPolygon(a), toPolygon(b)
	if cfg.Operation == 5 {
		subject, clip = clip, subject
	}
	resultPoly, err := gpc.PolygonClip(mapOperation(cfg.Operation), subject, clip)
	clipTime := time.Since(clipStart)

	renderStart := time.Now()
	if err == nil {
		drawContours(
			agg2d,
			fromPolygon(resultPoly),
			agg.RGBA(0.5, 0.0, 0.0, 0.5),
			agg.RGBA(0.0, 0.0, 0.0, 0.7),
		)
	}
	renderTime := time.Since(renderStart)

	overlayStats(agg2d, frameOffX, frameOffY, resultPoly, err, clipTime, renderTime)
}

func buildScene(cfg Config, w, h float64) ([]contour, []contour) {
	switch cfg.Scene {
	case 0:
		return sceneSimple(cfg, w, h)
	case 1:
		return sceneClosedStroke(cfg, w, h)
	case 2:
		return sceneGBArrows(cfg, w, h)
	case 3:
		return sceneGBSpiral(cfg, w, h)
	case 4:
		return sceneSpiralGlyph(cfg)
	default:
		return sceneSimple(cfg, w, h)
	}
}

func sceneSimple(cfg Config, w, h float64) ([]contour, []contour) {
	x := cfg.CenterX - w*0.5 + 100
	y := cfg.CenterY - h*0.5 + 100

	ps1 := path.NewPathStorageStl()
	ps1.MoveTo(x+140, y+145)
	ps1.LineTo(x+225, y+44)
	ps1.LineTo(x+296, y+219)
	ps1.ClosePolygon(0)
	ps1.LineTo(x+226, y+289)
	ps1.LineTo(x+82, y+292)

	ps1.MoveTo(x+220, y+222)
	ps1.LineTo(x+363, y+249)
	ps1.LineTo(x+265, y+331)

	ps1.MoveTo(x+242, y+243)
	ps1.LineTo(x+268, y+309)
	ps1.LineTo(x+325, y+261)

	ps1.MoveTo(x+259, y+259)
	ps1.LineTo(x+273, y+288)
	ps1.LineTo(x+298, y+266)

	ps2 := path.NewPathStorageStl()
	ps2.MoveTo(132, 177)
	ps2.LineTo(573, 363)
	ps2.LineTo(451, 390)
	ps2.LineTo(454, 474)

	return pathToContours(ps1), pathToContours(ps2)
}

func sceneClosedStroke(cfg Config, w, h float64) ([]contour, []contour) {
	x := cfg.CenterX - w*0.5 + 100
	y := cfg.CenterY - h*0.5 + 100

	ps1 := path.NewPathStorageStl()
	ps1.MoveTo(x+140, y+145)
	ps1.LineTo(x+225, y+44)
	ps1.LineTo(x+296, y+219)
	ps1.ClosePolygon(0)
	ps1.LineTo(x+226, y+289)
	ps1.LineTo(x+82, y+292)

	ps1.MoveTo(x+170, y+222)
	ps1.LineTo(x+215, y+331)
	ps1.LineTo(x+313, y+249)
	ps1.ClosePolygon(basics.PathFlagsCCW)

	ps2 := path.NewPathStorageStl()
	ps2.MoveTo(132, 177)
	ps2.LineTo(573, 363)
	ps2.LineTo(451, 390)
	ps2.LineTo(454, 474)
	ps2.ClosePolygon(0)

	stroke := conv.NewConvStroke(path.NewPathStorageStlVertexSourceAdapter(ps2))
	stroke.SetWidth(10.0)

	return pathToContours(ps1), vertexSourceToContours(stroke)
}

func sceneGBArrows(cfg Config, w, h float64) ([]contour, []contour) {
	gbPoly := path.NewPathStorageStl()
	aggshapes.MakeGBPoly(gbPoly)
	arrows := path.NewPathStorageStl()
	aggshapes.MakeArrows(arrows)

	a := transformContours(pathToContours(gbPoly), -1150, -1150, 2.0, 2.0, 0, 0)
	b := transformContours(
		pathToContours(arrows),
		-1150,
		-1150,
		2.0,
		2.0,
		cfg.CenterX-w*0.5,
		cfg.CenterY-h*0.5,
	)
	return a, b
}

func sceneGBSpiral(cfg Config, w, h float64) ([]contour, []contour) {
	gbPoly := path.NewPathStorageStl()
	aggshapes.MakeGBPoly(gbPoly)

	spiralPath := buildSpiralPath(cfg.CenterX, cfg.CenterY, 10, 150, 30, 0.0)
	stroke := conv.NewConvStroke(path.NewPathStorageStlVertexSourceAdapter(spiralPath))
	stroke.SetWidth(15.0)

	return transformContours(pathToContours(gbPoly), -1150, -1150, 2.0, 2.0, 0, 0), vertexSourceToContours(stroke)
}

func sceneSpiralGlyph(cfg Config) ([]contour, []contour) {
	spiralPath := buildSpiralPath(cfg.CenterX, cfg.CenterY, 10, 150, 30, 0.0)
	stroke := conv.NewConvStroke(path.NewPathStorageStlVertexSourceAdapter(spiralPath))
	stroke.SetWidth(15.0)

	glyph := buildGlyphPath()
	curve := conv.NewConvCurve(path.NewPathStorageStlVertexSourceAdapter(glyph))

	return vertexSourceToContours(stroke), transformContours(vertexSourceToContours(curve), 0, 0, 4.0, 4.0, 220, 200)
}

func buildSpiralPath(cx, cy, r1, r2, step, startAngle float64) *path.PathStorageStl {
	ps := path.NewPathStorageStl()
	angle := startAngle
	radius := r1
	deltaAngle := 4.0 * math.Pi / 180.0
	deltaRadius := step / 90.0
	first := true
	for radius <= r2 {
		x := cx + math.Cos(angle)*radius
		y := cy + math.Sin(angle)*radius
		if first {
			ps.MoveTo(x, y)
			first = false
		} else {
			ps.LineTo(x, y)
		}
		radius += deltaRadius
		angle += deltaAngle
	}
	return ps
}

func buildGlyphPath() *path.PathStorageStl {
	ps := path.NewPathStorageStl()
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
	ps.ClosePolygon(0)

	ps.MoveTo(28.47, 9.62)
	ps.LineTo(28.47, 26.66)
	ps.Curve3(21.09, 23.73, 18.95, 22.51)
	ps.Curve3(15.09, 20.36, 13.43, 18.02)
	ps.Curve3(11.77, 15.67, 11.77, 12.89)
	ps.Curve3(11.77, 9.38, 13.87, 7.06)
	ps.Curve3(15.97, 4.74, 18.70, 4.74)
	ps.Curve3(22.41, 4.74, 28.47, 9.62)
	ps.ClosePolygon(0)
	return ps
}

func mapOperation(op int) gpc.GPCOp {
	switch op {
	case 1:
		return gpc.GPCUnion
	case 2:
		return gpc.GPCInt
	case 3:
		return gpc.GPCXor
	case 4:
		return gpc.GPCDiff
	case 5:
		return gpc.GPCDiff
	default:
		return gpc.GPCInt
	}
}

func mirrorContoursY(cs []contour, h float64) []contour {
	out := make([]contour, len(cs))
	for i := range cs {
		out[i] = make(contour, len(cs[i]))
		for j := range cs[i] {
			out[i][j] = pt{x: cs[i][j].x, y: h - cs[i][j].y}
		}
	}
	return out
}

func toPolygon(cs []contour) *gpc.GPCPolygon {
	p := gpc.NewGPCPolygon()
	for _, c := range cs {
		if len(c) < 3 {
			continue
		}
		vl := gpc.NewGPCVertexList(len(c))
		for _, q := range c {
			vl.AddVertex(q.x, q.y)
		}
		_ = p.AddContour(vl, false)
	}
	return p
}

func fromPolygon(p *gpc.GPCPolygon) []contour {
	if p == nil {
		return nil
	}
	out := make([]contour, 0, p.NumContours)
	for i := 0; i < p.NumContours; i++ {
		c, _, err := p.GetContour(i)
		if err != nil || c == nil || c.NumVertices < 3 {
			continue
		}
		cc := make(contour, 0, c.NumVertices)
		for _, v := range c.Vertices {
			cc = append(cc, pt{x: v.X, y: v.Y})
		}
		out = append(out, cc)
	}
	return out
}

func pathToContours(ps *path.PathStorageStl) []contour {
	return vertexSourceToContours(path.NewPathStorageStlVertexSourceAdapter(ps))
}

func vertexSourceToContours(vs interface {
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
}) []contour {
	var out []contour
	var cur contour

	vs.Rewind(0)
	for {
		x, y, cmd := vs.Vertex()
		switch {
		case basics.IsStop(cmd):
			if len(cur) >= 3 {
				out = append(out, closeIfNeeded(cur))
			}
			return out
		case basics.IsMoveTo(cmd):
			if len(cur) >= 3 {
				out = append(out, closeIfNeeded(cur))
			}
			cur = contour{{x: x, y: y}}
		case basics.IsVertex(cmd):
			cur = append(cur, pt{x: x, y: y})
		case basics.IsEndPoly(cmd):
			if len(cur) >= 3 {
				out = append(out, closeIfNeeded(cur))
			}
			cur = nil
		}
	}
}

func closeIfNeeded(c contour) contour {
	if len(c) < 2 {
		return c
	}
	first := c[0]
	last := c[len(c)-1]
	if math.Abs(first.x-last.x) > 1e-9 || math.Abs(first.y-last.y) > 1e-9 {
		c = append(c, first)
	}
	return c
}

func transformContours(cs []contour, tx1, ty1, sx, sy, tx2, ty2 float64) []contour {
	out := make([]contour, 0, len(cs))
	for _, c := range cs {
		cc := make(contour, 0, len(c))
		for _, p := range c {
			x := (p.x + tx1) * sx
			y := (p.y + ty1) * sy
			cc = append(cc, pt{x: x + tx2, y: y + ty2})
		}
		out = append(out, cc)
	}
	return out
}

func drawContours(a *agg.Agg2D, cs []contour, fill, line agg.Color) {
	for _, c := range cs {
		if len(c) < 3 {
			continue
		}
		a.ResetPath()
		a.MoveTo(c[0].x, c[0].y)
		for i := 1; i < len(c); i++ {
			a.LineTo(c[i].x, c[i].y)
		}
		a.ClosePolygon()
		a.FillColor(fill)
		if line.A > 0 {
			a.LineColor(line)
			a.LineWidth(1)
			a.DrawPath(agg.FillAndStroke)
		} else {
			a.NoLine()
			a.DrawPath(agg.FillOnly)
		}
	}
}

func overlayStats(a *agg.Agg2D, offX, offY float64, result *gpc.GPCPolygon, err error, clipTime, renderTime time.Duration) {
	contours, points := polygonStats(result)
	line1 := fmt.Sprintf("Contours: %d   Points: %d", contours, points)
	line2 := fmt.Sprintf(
		"GPC=%.3fms Render=%.3fms",
		float64(clipTime.Microseconds())/1000.0,
		float64(renderTime.Microseconds())/1000.0,
	)
	if err != nil {
		line2 = "GPC error: " + err.Error()
	}

	a.FontGSV(10)
	a.FillColor(agg.Black)
	a.NoLine()
	a.Text(250+offX, 15+offY, line1, false, 0, 0)
	a.Text(250+offX, 30+offY, line2, false, 0, 0)
}

func polygonStats(p *gpc.GPCPolygon) (int, int) {
	if p == nil {
		return 0, 0
	}
	points := 0
	for i := 0; i < p.NumContours; i++ {
		c, _, err := p.GetContour(i)
		if err != nil || c == nil {
			continue
		}
		points += c.NumVertices
	}
	return p.NumContours, points
}

func sceneAColor(scene int) agg.Color {
	if scene == 2 || scene == 3 {
		return agg.RGBA(0.5, 0.5, 0.0, 0.1)
	}
	return agg.RGBA(0.0, 0.0, 0.0, 0.1)
}

func sceneALineColor(scene int) agg.Color {
	if scene == 2 || scene == 3 {
		return agg.Black
	}
	return agg.Transparent
}

func sceneBColor(scene int) agg.Color {
	if scene == 2 || scene == 3 {
		return agg.RGBA(0.0, 0.5, 0.5, 0.1)
	}
	return agg.RGBA(0.0, 0.6, 0.0, 0.1)
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
