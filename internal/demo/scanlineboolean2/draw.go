package scanlineboolean2

import (
	"math"

	agg "agg_go"
	"agg_go/internal/basics"
	"agg_go/internal/demo/aggshapes"
	"agg_go/internal/gpc"
	"agg_go/internal/path"
)

type Config struct {
	Mode         int
	FillRule     int
	ScanlineType int
	Operation    int
	CenterX      float64
	CenterY      float64
}

type pt struct {
	x float64
	y float64
}

type contour []pt

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func Draw(ctx *agg.Context, cfg Config) {
	w := float64(ctx.GetImage().Width())
	h := float64(ctx.GetImage().Height())
	if cfg.CenterX == 0 && cfg.CenterY == 0 {
		cfg.CenterX = w / 2
		cfg.CenterY = h / 2
	}
	cfg.Mode = clampInt(cfg.Mode, 0, 4)
	cfg.FillRule = clampInt(cfg.FillRule, 0, 1)
	cfg.ScanlineType = clampInt(cfg.ScanlineType, 0, 2)
	cfg.Operation = clampInt(cfg.Operation, 0, 6)

	a, b := buildShapes(cfg, w, h)
	// AGG's original demo runs with flip_y=true; mirror geometry in Y to match
	// the reference orientation in this top-left-origin framebuffer.
	a = mirrorContoursY(a, h)
	b = mirrorContoursY(b, h)
	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()
	agg2d.ClearAll(agg.White)

	agg2d.FillEvenOdd(cfg.FillRule == 0)
	drawContours(agg2d, a, agg.RGBA(0.0, 0.0, 0.0, 0.1), agg.Transparent)
	drawContours(agg2d, b, agg.RGBA(0.0, 0.6, 0.0, 0.1), agg.Transparent)

	if cfg.Operation > 0 {
		pa, pb := toGPCPolygon(a), toGPCPolygon(b)
		op := mapOperation(cfg.Operation)
		if cfg.Operation == 6 {
			pa, pb = pb, pa
		}
		if out, err := gpc.PolygonClip(op, pa, pb); err == nil {
			result := fromGPCPolygon(out)
			drawContours(agg2d, result, agg.RGBA(0.5, 0.0, 0.0, 0.5), agg.RGBA(0, 0, 0, 0.55))
		}
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

func mapOperation(op int) gpc.GPCOp {
	switch op {
	case 1:
		return gpc.GPCUnion
	case 2:
		return gpc.GPCInt
	case 3, 4:
		return gpc.GPCXor
	case 5:
		return gpc.GPCDiff
	case 6:
		return gpc.GPCDiff
	default:
		return gpc.GPCInt
	}
}

func buildShapes(cfg Config, w, h float64) ([]contour, []contour) {
	switch cfg.Mode {
	case 0:
		return modeSimple(cfg, w, h)
	case 1:
		return modeClosedStroke(cfg, w, h)
	case 2:
		return modeGBArrows(cfg, w, h)
	case 3:
		return modeGBSpiral(cfg, w, h)
	case 4:
		return modeSpiralGlyph(cfg)
	default:
		return modeSimple(cfg, w, h)
	}
}

func modeSimple(cfg Config, w, h float64) ([]contour, []contour) {
	dx := cfg.CenterX - w/2 + 100
	dy := cfg.CenterY - h/2 + 100
	a := []contour{
		{{dx + 140, dy + 145}, {dx + 225, dy + 44}, {dx + 296, dy + 219}, {dx + 226, dy + 289}, {dx + 82, dy + 292}},
		{{dx + 220, dy + 222}, {dx + 363, dy + 249}, {dx + 265, dy + 331}},
		{{dx + 242, dy + 243}, {dx + 325, dy + 261}, {dx + 268, dy + 309}},
		{{dx + 259, dy + 259}, {dx + 273, dy + 288}, {dx + 298, dy + 266}},
	}
	b := []contour{{
		{132, 177}, {573, 363}, {451, 390}, {454, 474},
	}}
	return a, b
}

func modeClosedStroke(cfg Config, w, h float64) ([]contour, []contour) {
	a, _ := modeSimple(cfg, w, h)
	b := []contour{{
		{132, 177}, {573, 363}, {451, 390}, {454, 474},
	}}
	return a, b
}

func modeGBArrows(cfg Config, w, h float64) ([]contour, []contour) {
	psGB := path.NewPathStorageStl()
	aggshapes.MakeGBPoly(psGB)
	psAr := path.NewPathStorageStl()
	aggshapes.MakeArrows(psAr)

	a := transformContours(pathToContours(psGB), -1150, -1150, 2.0, 2.0, 0, 0)
	tx := cfg.CenterX - w/2
	ty := cfg.CenterY - h/2
	b := transformContours(pathToContours(psAr), -1150, -1150, 2.0, 2.0, tx, ty)
	return a, b
}

func modeGBSpiral(cfg Config, w, h float64) ([]contour, []contour) {
	psGB := path.NewPathStorageStl()
	aggshapes.MakeGBPoly(psGB)
	a := transformContours(pathToContours(psGB), -1150, -1150, 2.0, 2.0, 0, 0)
	b := []contour{buildSpiral(cfg.CenterX, cfg.CenterY, 10, 150, 400, 0)}
	return a, b
}

func modeSpiralGlyph(cfg Config) ([]contour, []contour) {
	a := []contour{buildSpiral(cfg.CenterX, cfg.CenterY, 10, 150, 400, 0)}
	b := []contour{buildGlyph(cfg.CenterX+10, cfg.CenterY+5, 4.0)}
	return a, b
}

func buildSpiral(cx, cy, r1, r2 float64, steps int, start float64) contour {
	pts := make(contour, 0, steps+1)
	for i := range steps {
		t := float64(i) / float64(steps-1)
		a := start + t*math.Pi*8.0
		r := r1 + (r2-r1)*t
		pts = append(pts, pt{x: cx + math.Cos(a)*r, y: cy + math.Sin(a)*r})
	}
	return pts
}

func buildGlyph(cx, cy, s float64) contour {
	return contour{
		{cx - 70*s/4, cy - 30*s/4},
		{cx - 20*s/4, cy - 80*s/4},
		{cx + 35*s/4, cy - 50*s/4},
		{cx + 55*s/4, cy + 10*s/4},
		{cx + 45*s/4, cy + 70*s/4},
		{cx + 10*s/4, cy + 90*s/4},
		{cx - 35*s/4, cy + 80*s/4},
		{cx - 60*s/4, cy + 40*s/4},
		{cx - 30*s/4, cy + 20*s/4},
		{cx + 5*s/4, cy + 30*s/4},
		{cx + 20*s/4, cy + 5*s/4},
		{cx + 2*s/4, cy - 20*s/4},
		{cx - 40*s/4, cy - 18*s/4},
	}
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

func toGPCPolygon(cs []contour) *gpc.GPCPolygon {
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

func fromGPCPolygon(p *gpc.GPCPolygon) []contour {
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
	var out []contour
	var cur contour
	ps.Rewind(0)
	for {
		x, y, cmd := ps.NextVertex()
		pc := basics.PathCommand(cmd)
		switch {
		case basics.IsStop(pc):
			if len(cur) >= 3 {
				out = append(out, closeIfNeeded(cur))
			}
			return out
		case basics.IsMoveTo(pc):
			if len(cur) >= 3 {
				out = append(out, closeIfNeeded(cur))
			}
			cur = contour{{x: x, y: y}}
		case basics.IsVertex(pc):
			cur = append(cur, pt{x: x, y: y})
		case basics.IsEndPoly(pc):
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
	f, l := c[0], c[len(c)-1]
	if math.Abs(f.x-l.x) > 1e-9 || math.Abs(f.y-l.y) > 1e-9 {
		c = append(c, f)
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
