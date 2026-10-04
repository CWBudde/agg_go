package agg2d

import (
	"bufio"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// The expected buffers in testdata/cpporacle are rendered by the original AGG
// 2.6 agg2d.cpp (see testdata/cpporacle/agg2d_oracle.cpp). Each scene below is
// a line-by-line replay of the C++ scene of the same name.

const (
	oracleW = 40
	oracleH = 40
)

func oracleTriangle(g *Agg2D, flag DrawPathFlag) {
	g.ResetPath()
	g.MoveTo(4.3, 5.7)
	g.LineTo(35.6, 12.2)
	g.LineTo(13.1, 36.4)
	g.ClosePolygon()
	g.DrawPath(flag)
}

func oracleCurveShape(g *Agg2D, s float64) {
	g.ResetPath()
	g.MoveTo(0.5*s, 5.5*s)
	g.CubicCurveTo(0.5*s, 0.2*s, 6.2*s, 0.4*s, 5.8*s, 3.1*s)
	g.CubicCurveTo(5.4*s, 6.0*s, 2.0*s, 4.0*s, 0.5*s, 5.5*s)
	g.ClosePolygon()
	g.DrawPath(FillAndStroke)
}

var oracleScenes = []struct {
	name string
	fn   func(g *Agg2D)
}{
	{"master_alpha_fill", func(g *Agg2D) {
		g.SetMasterAlpha(0.5)
		g.NoLine()
		g.FillColor(Color{255, 0, 0, 255})
		oracleTriangle(g, FillOnly)
		g.FillColor(Color{0, 0, 255, 128})
		g.ResetPath()
		g.MoveTo(2.5, 22.25)
		g.LineTo(37.75, 22.25)
		g.LineTo(37.75, 30.5)
		g.LineTo(2.5, 30.5)
		g.ClosePolygon()
		g.DrawPath(FillOnly)
	}},
	{"master_alpha_stroke", func(g *Agg2D) {
		g.SetMasterAlpha(0.5)
		g.NoFill()
		g.LineColor(Color{0, 0, 255, 255})
		g.LineWidth(3.0)
		oracleTriangle(g, StrokeOnly)
	}},
	{"master_alpha_over1", func(g *Agg2D) {
		g.SetMasterAlpha(1.5)
		g.LineColor(Color{0, 0, 0, 200})
		g.FillColor(Color{0, 255, 0, 160})
		g.LineWidth(1.5)
		oracleTriangle(g, FillAndStroke)
	}},
	{"master_alpha_fill_with_line_color", func(g *Agg2D) {
		g.SetMasterAlpha(0.3)
		g.LineColor(Color{255, 0, 255, 255})
		oracleTriangle(g, FillWithLineColor)
	}},
	{"gamma_2_2", func(g *Agg2D) {
		g.SetAntiAliasGamma(2.2)
		g.FillColor(Color{255, 0, 0, 255})
		g.LineColor(Color{0, 0, 0, 255})
		g.LineWidth(1.25)
		oracleTriangle(g, FillAndStroke)
	}},
	{"gamma_alpha", func(g *Agg2D) {
		g.SetAntiAliasGamma(0.6)
		g.SetMasterAlpha(0.7)
		g.FillColor(Color{0, 0, 255, 255})
		g.LineColor(Color{0, 255, 255, 255})
		g.LineWidth(2.0)
		oracleTriangle(g, FillAndStroke)
	}},
	{"gamma_large", func(g *Agg2D) {
		g.SetAntiAliasGamma(4.5)
		g.FillColor(Color{0, 0, 0, 255})
		oracleTriangle(g, FillOnly)
	}},
	{"gradient_master_alpha", func(g *Agg2D) {
		g.SetMasterAlpha(0.5)
		g.NoLine()
		g.FillLinearGradient(3, 3, 37, 30, Color{0, 0, 255, 255}, Color{0, 0, 255, 40}, 1.0)
		oracleTriangle(g, FillOnly)
	}},
	{"approx_default", func(g *Agg2D) {
		g.FillColor(Color{0, 255, 255, 255})
		g.LineColor(Color{0, 0, 0, 255})
		g.LineWidth(3.0)
		oracleCurveShape(g, 6.0)
	}},
	{"approx_scaled", func(g *Agg2D) {
		g.Scale(6.0, 6.0)
		g.FillColor(Color{0, 255, 255, 255})
		g.LineColor(Color{0, 0, 0, 255})
		g.LineWidth(0.5)
		oracleCurveShape(g, 1.0)
	}},
	{"approx_after_reset", func(g *Agg2D) {
		g.Scale(0.25, 0.25)
		g.ResetTransformations()
		g.FillColor(Color{0, 255, 255, 255})
		g.LineColor(Color{0, 0, 0, 255})
		g.LineWidth(3.0)
		oracleCurveShape(g, 6.0)
	}},
	{"approx_skew", func(g *Agg2D) {
		g.Scale(0.5, 0.5)
		g.Skew(0.4, 0.1)
		g.Rotate(0.2)
		g.Translate(4.0, 2.0)
		g.FillColor(Color{0, 255, 255, 255})
		g.LineColor(Color{0, 0, 0, 255})
		g.LineWidth(4.0)
		oracleCurveShape(g, 10.0)
	}},
	{"guard_nofill_clear", func(g *Agg2D) {
		g.SetBlendMode(BlendClear)
		g.NoFill()
		g.NoLine()
		oracleTriangle(g, FillAndStroke)
		oracleTriangle(g, FillOnly)
		oracleTriangle(g, StrokeOnly)
		oracleTriangle(g, FillWithLineColor)
	}},
	{"guard_line_width", func(g *Agg2D) {
		g.FillColor(Color{255, 0, 0, 255})
		g.LineColor(Color{0, 0, 0, 255})
		g.LineWidth(-3.0)
		oracleTriangle(g, FillAndStroke)
		g.LineWidth(0.0)
		g.ResetPath()
		g.MoveTo(3, 36)
		g.LineTo(37, 30)
		g.DrawPath(StrokeOnly)
	}},
}

func TestCppOracleScenes(t *testing.T) {
	for _, sc := range oracleScenes {
		t.Run(sc.name, func(t *testing.T) {
			want, err := os.ReadFile(filepath.Join("testdata", "cpporacle", sc.name+".rgba"))
			if err != nil {
				t.Fatalf("read oracle: %v", err)
			}
			if len(want) != oracleW*oracleH*4 {
				t.Fatalf("oracle size %d, want %d", len(want), oracleW*oracleH*4)
			}

			buf := make([]uint8, oracleW*oracleH*4)
			g := NewAgg2D()
			g.Attach(buf, oracleW, oracleH, oracleW*4)
			g.ClearAll(Color{255, 255, 0, 255})
			sc.fn(g)

			compareOracleBuffers(t, buf, want)
		})
	}
}

func compareOracleBuffers(t *testing.T, got, want []uint8) {
	t.Helper()
	diffPixels, maxDiff := 0, 0
	var samples []string
	for i := 0; i < oracleW*oracleH; i++ {
		g := got[i*4 : i*4+4]
		w := want[i*4 : i*4+4]
		if g[0] == w[0] && g[1] == w[1] && g[2] == w[2] && g[3] == w[3] {
			continue
		}
		diffPixels++
		for c := 0; c < 4; c++ {
			d := int(g[c]) - int(w[c])
			if d < 0 {
				d = -d
			}
			if d > maxDiff {
				maxDiff = d
			}
		}
		if len(samples) < 8 {
			samples = append(samples, fmt.Sprintf("(%d,%d) got %v want %v", i%oracleW, i/oracleW, g, w))
		}
	}
	if diffPixels > 0 {
		t.Errorf("%d/%d pixels differ from C++ Agg2D (max channel diff %d); first: %s",
			diffPixels, oracleW*oracleH, maxDiff, strings.Join(samples, "; "))
	}
}

// TestCppOracleImageFilterLUT checks that ImageFilter() selects the same
// filter (kind and radius) as C++ Agg2D::imageFilter() (agg2d.cpp:1241) for
// every filter the C++ enum defines, by comparing the normalized weight LUT.
func TestCppOracleImageFilterLUT(t *testing.T) {
	filters := map[string]ImageFilter{
		"Bilinear":    Bilinear,
		"Hanning":     Hanning,
		"Hermite":     Hermite,
		"Quadric":     Quadric,
		"Bicubic":     Bicubic,
		"Catrom":      Catrom,
		"Spline16":    Spline16,
		"Spline36":    Spline36,
		"Blackman144": Blackman144,
	}

	f, err := os.Open(filepath.Join("testdata", "cpporacle", "image_filters.txt"))
	if err != nil {
		t.Fatalf("open oracle: %v", err)
	}
	defer f.Close()

	seen := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) != 5 {
			t.Fatalf("malformed oracle line %q", sc.Text())
		}
		name := fields[0]
		filter, ok := filters[name]
		if !ok {
			t.Fatalf("unknown filter %q in oracle", name)
		}
		radius, _ := strconv.ParseFloat(fields[1], 64)
		diameter, _ := strconv.Atoi(fields[2])
		start, _ := strconv.Atoi(fields[3])
		hash, _ := strconv.ParseUint(fields[4], 16, 64)

		g := NewAgg2D()
		g.ImageFilter(Bilinear)
		g.ImageFilter(filter)
		lut := g.imageFilterLUT
		if lut.Radius() != radius || lut.Diameter() != diameter || lut.Start() != start {
			t.Errorf("%s: radius/diameter/start = %v/%d/%d, want %v/%d/%d",
				name, lut.Radius(), lut.Diameter(), lut.Start(), radius, diameter, start)
			continue
		}
		h := fnv.New64a()
		for _, w := range lut.WeightArray()[:diameter*256] {
			_, _ = h.Write([]byte{byte(uint16(w)), byte(uint16(w) >> 8)})
		}
		if got := h.Sum64(); got != hash {
			t.Errorf("%s: weight LUT hash %016x, want %016x", name, got, hash)
		}
		seen++
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if seen != len(filters) {
		t.Fatalf("checked %d filters, want %d", seen, len(filters))
	}
}

// TestCPPOracleColorGradient checks Color.Gradient against Agg2D::Color::
// gradient (rgba8T::gradient) over the channel grid dumped by
// agg2d_oracle.cpp into color_gradient.txt.
func TestCPPOracleColorGradient(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "cpporacle", "color_gradient.txt"))
	if err != nil {
		t.Fatalf("read oracle: %v", err)
	}
	want, err := strconv.ParseUint(strings.TrimSpace(string(data)), 16, 64)
	if err != nil {
		t.Fatalf("parse oracle hash: %v", err)
	}

	values := []uint8{0, 1, 2, 51, 100, 127, 128, 200, 254, 255}
	const steps = 4096
	n := len(values)
	h := fnv.New64a()
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			a, b := values[i], values[j]
			ra, rb := values[n-1-i], values[n-1-j]
			c1 := Color{a, b, ra, rb}
			c2 := Color{b, ra, rb, a}
			for s := 0; s <= steps; s++ {
				r := c1.Gradient(c2, float64(s)/float64(steps))
				_, _ = h.Write(r[:])
			}
		}
	}
	if got := h.Sum64(); got != want {
		t.Fatalf("Color.Gradient hash %016x, want %016x", got, want)
	}
}
