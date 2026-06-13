// Command engine-compare renders the backend-neutral scene corpus through every
// available engine and writes per-scene <scene>_<engine>.png images plus, when
// both the port and cpp engines produce output, an amplified <scene>_diff.png
// and a one-line difference summary.
//
// It is built untagged. In the default build only the port engine is available,
// so it renders port-only and prints a notice. Built with -tags "agogo aggreal"
// (and "freetype" for the text scene) the cpp engine becomes available and the
// tool emits side-by-side port/cpp/diff outputs for manual parity inspection.
//
//	go run ./cmd/engine-compare -out ./engine-compare-out
//	go run -tags "agogo aggreal freetype" ./cmd/engine-compare -out /tmp/ec -scene gradient
package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/cwbudde/agg_go/engine"
	"github.com/cwbudde/agg_go/engine/scene"
	"github.com/cwbudde/agg_go/internal/imgdiff"
	"github.com/cwbudde/agg_go/tests/visual/framework"
)

func main() {
	out := flag.String("out", "engine-compare-out", "output directory")
	sceneFilter := flag.String("scene", "", "render only scenes whose name contains this substring")
	enginesFlag := flag.String("engines", "port,cpp", "comma-separated engines to render (intersected with availability)")
	factor := flag.Int("factor", 8, "diff amplification factor")
	flag.Parse()

	if *factor <= 0 {
		fmt.Fprintln(os.Stderr, "-factor must be positive")
		os.Exit(2)
	}

	available := engine.Available()
	requested := parseEngines(*enginesFlag)
	var kinds []engine.Kind
	for _, k := range requested {
		if slices.Contains(available, k) {
			kinds = append(kinds, k)
		} else {
			fmt.Printf("notice: engine %q not available in this build; skipping\n", k)
		}
	}
	if len(kinds) == 0 {
		fmt.Fprintln(os.Stderr, "no requested engines are available")
		os.Exit(1)
	}

	// Build assets once per engine kind (image scenes need a per-engine source).
	assets := make(map[engine.Kind]*scene.Assets, len(kinds))
	for _, k := range kinds {
		a, err := scene.BuildAssets(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "build assets for %s: %v\n", k, err)
			os.Exit(1)
		}
		assets[k] = a
	}

	scenes := scene.Filter(*sceneFilter)
	if len(scenes) == 0 {
		fmt.Fprintf(os.Stderr, "no scenes match filter %q\n", *sceneFilter)
		os.Exit(1)
	}

	for _, s := range scenes {
		rendered := make(map[engine.Kind]image.Image)
		for _, k := range kinds {
			if !s.SupportedBy(k) {
				fmt.Printf("scene=%s engine=%s skipped (unsupported caps %v)\n", s.Name, k, s.Caps)
				continue
			}
			img, err := renderScene(s, k, assets[k])
			if err != nil {
				if errors.Is(err, scene.ErrAssetUnavailable) || errors.Is(err, engine.ErrUnsupportedCapability) {
					fmt.Printf("scene=%s engine=%s skipped (%v)\n", s.Name, k, err)
					continue
				}
				fmt.Fprintf(os.Stderr, "scene=%s engine=%s render error: %v\n", s.Name, k, err)
				os.Exit(1)
			}
			path := filepath.Join(*out, fmt.Sprintf("%s_%s.png", s.Name, k))
			if err := framework.SaveImage(img, path); err != nil {
				fmt.Fprintf(os.Stderr, "save %s: %v\n", path, err)
				os.Exit(1)
			}
			rendered[k] = img
		}

		// Emit a diff + stats when both port and cpp produced output.
		port, hasPort := rendered[engine.Port]
		cpp, hasCPP := rendered[engine.CPP]
		if hasPort && hasCPP {
			emitDiff(*out, s, port, cpp, *factor)
		} else {
			fmt.Printf("scene=%s size=%dx%d engines=%s (no cross-engine diff)\n",
				s.Name, s.Width, s.Height, renderedKinds(rendered))
		}
	}
}

func renderScene(s scene.Scene, kind engine.Kind, assets *scene.Assets) (image.Image, error) {
	ctx, err := engine.NewContext(s.Width, s.Height, engine.Config{Kind: kind})
	if err != nil {
		return nil, err
	}
	if err := s.Draw(ctx, assets); err != nil {
		return nil, err
	}
	return ctx.GetImage().ToGoImage(), nil
}

func emitDiff(out string, s scene.Scene, port, cpp image.Image, factor int) {
	stats, err := imgdiff.Analyze(port, cpp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scene=%s analyze: %v\n", s.Name, err)
		return
	}
	diff, err := imgdiff.Amplify(port, cpp, factor)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scene=%s amplify: %v\n", s.Name, err)
		return
	}
	diffPath := filepath.Join(out, s.Name+"_diff.png")
	if err := framework.SaveImage(diff, diffPath); err != nil {
		fmt.Fprintf(os.Stderr, "save %s: %v\n", diffPath, err)
		return
	}
	fmt.Printf("scene=%s size=%dx%d diff_pixels=%d/%d ratio=%.6f max_diff=%d avg_diff=%.4f rmse=%.4f\n",
		s.Name, stats.Width, stats.Height, stats.DifferentPixels, stats.TotalPixels,
		stats.Ratio(), stats.MaxDiff, stats.AverageDiff, stats.RMSE)
}

func parseEngines(s string) []engine.Kind {
	var out []engine.Kind
	for part := range strings.SplitSeq(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, engine.Kind(part))
		}
	}
	return out
}

func renderedKinds(m map[engine.Kind]image.Image) string {
	var names []string
	for k := range m {
		names = append(names, string(k))
	}
	slices.Sort(names)
	return strings.Join(names, ",")
}
