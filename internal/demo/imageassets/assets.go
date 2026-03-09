package imageassets

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"

	agg "github.com/MeKo-Christian/agg_go"
)

var (
	//go:embed agg.ppm
	aggPPM []byte
	//go:embed spheres.ppm
	spheresPPM []byte
)

// Agg returns AGG's original agg.ppm image as RGBA.
func Agg() (*agg.Image, error) {
	return decodePPM(aggPPM)
}

// Spheres returns AGG's original spheres.ppm image as RGBA.
func Spheres() (*agg.Image, error) {
	return decodePPM(spheresPPM)
}

func decodePPM(data []byte) (*agg.Image, error) {
	tokens := make([]string, 0, 4)
	i := 0
	for len(tokens) < 4 {
		for i < len(data) {
			b := data[i]
			if b == '#' {
				for i < len(data) && data[i] != '\n' {
					i++
				}
				continue
			}
			if b == ' ' || b == '\t' || b == '\r' || b == '\n' {
				i++
				continue
			}
			break
		}
		if i >= len(data) {
			return nil, fmt.Errorf("invalid PPM header")
		}
		start := i
		for i < len(data) {
			b := data[i]
			if b == ' ' || b == '\t' || b == '\r' || b == '\n' {
				break
			}
			i++
		}
		tokens = append(tokens, string(data[start:i]))
	}

	if tokens[0] != "P6" {
		return nil, fmt.Errorf("unsupported PPM magic: %q", tokens[0])
	}
	w, err := strconv.Atoi(tokens[1])
	if err != nil || w <= 0 {
		return nil, fmt.Errorf("invalid PPM width: %q", tokens[1])
	}
	h, err := strconv.Atoi(tokens[2])
	if err != nil || h <= 0 {
		return nil, fmt.Errorf("invalid PPM height: %q", tokens[2])
	}
	maxv, err := strconv.Atoi(tokens[3])
	if err != nil || maxv <= 0 || maxv > 255 {
		return nil, fmt.Errorf("invalid PPM max value: %q", tokens[3])
	}

	for i < len(data) && strings.ContainsRune(" \t\r\n", rune(data[i])) {
		i++
	}

	rgbLen := w * h * 3
	if i+rgbLen > len(data) {
		return nil, fmt.Errorf("PPM payload too short")
	}
	rgb := data[i : i+rgbLen]
	rgba := make([]byte, w*h*4)
	for p := 0; p < w*h; p++ {
		ri := p * 3
		oi := p * 4
		rgba[oi+0] = rgb[ri+0]
		rgba[oi+1] = rgb[ri+1]
		rgba[oi+2] = rgb[ri+2]
		rgba[oi+3] = 255
	}

	return agg.NewImage(rgba, w, h, w*4), nil
}
