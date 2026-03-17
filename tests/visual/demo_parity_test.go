package visual

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/MeKo-Christian/agg_go/tests/visual/framework"
)

func updateRequested() bool { return os.Getenv("UPDATE_VISUAL") != "" }

const (
	cppReferenceDir = "reference/cpp/examples"
	goReferenceDir  = "reference/go/examples"
	cppDiffDir      = "diffs/cpp_demo_parity"
)

type demoConfig struct {
	name string
	dir  string
}

var demoConfigs = []demoConfig{
	{name: "aa_demo", dir: "examples/core/intermediate/aa_demo"},
	{name: "alpha_mask", dir: "examples/core/intermediate/alpha_mask"},
	{name: "alpha_mask2", dir: "examples/core/intermediate/alpha_mask2"},
	{name: "alpha_mask3", dir: "examples/core/intermediate/alpha_mask3"},
	{name: "bezier_div", dir: "examples/core/intermediate/bezier_div"},
	{name: "blend_color", dir: "examples/core/intermediate/blend_color"},
	{name: "blur", dir: "examples/core/intermediate/blur"},
	{name: "bspline", dir: "examples/core/intermediate/bspline"},
	{name: "circles", dir: "examples/core/basic/circles"},
	{name: "component_rendering", dir: "examples/core/basic/component_rendering"},
	{name: "compositing", dir: "examples/core/intermediate/compositing"},
	{name: "compositing2", dir: "examples/core/intermediate/compositing2"},
	{name: "conv_contour", dir: "examples/core/intermediate/conv_contour"},
	{name: "conv_dash_marker", dir: "examples/core/intermediate/conv_dash_marker"},
	{name: "conv_stroke", dir: "examples/core/intermediate/conv_stroke"},
	{name: "distortions", dir: "examples/core/advanced/distortions"},
	{name: "flash_rasterizer", dir: "examples/core/advanced/flash_rasterizer"},
	{name: "flash_rasterizer2", dir: "examples/core/intermediate/flash_rasterizer2"},
	{name: "gamma_correction", dir: "examples/core/advanced/gamma_correction"},
	{name: "gamma_ctrl", dir: "examples/core/advanced/gamma_ctrl"},
	{name: "gamma_tuner", dir: "examples/core/advanced/gamma_tuner"},
	{name: "gouraud", dir: "examples/core/intermediate/gouraud"},
	{name: "gouraud_mesh", dir: "examples/core/intermediate/gouraud_mesh"},
	{name: "gradient_focal", dir: "examples/core/intermediate/gradient_focal"},
	{name: "gradients", dir: "examples/core/intermediate/gradients"},
	{name: "gradients_contour", dir: "examples/core/intermediate/gradients_contour"},
	{name: "graph_test", dir: "examples/core/intermediate/graph_test"},
	{name: "idea", dir: "examples/core/intermediate/idea"},
	{name: "image1", dir: "examples/core/intermediate/image1"},
	{name: "image_alpha", dir: "examples/core/intermediate/image_alpha"},
	{name: "image_filters", dir: "examples/core/intermediate/image_filters"},
	{name: "image_filters2", dir: "examples/core/intermediate/image_filters2"},
	{name: "image_fltr_graph", dir: "examples/core/intermediate/image_fltr_graph"},
	{name: "image_perspective", dir: "examples/core/intermediate/image_perspective"},
	{name: "image_resample", dir: "examples/core/intermediate/image_resample"},
	{name: "line_patterns", dir: "examples/core/intermediate/line_patterns"},
	{name: "line_patterns_clip", dir: "examples/core/intermediate/line_patterns_clip"},
	{name: "lion", dir: "examples/core/intermediate/lion"},
	{name: "lion_lens", dir: "examples/core/intermediate/lion_lens"},
	{name: "lion_outline", dir: "examples/core/intermediate/lion_outline"},
}

// TestDemoDimensions checks that each Go demo produces an image whose
// dimensions match the corresponding C++ reference PNG.  This is a
// prerequisite for the full visual-content comparison.
func TestUpdateGoReferences(t *testing.T) {
	if !updateRequested() {
		t.Skip("set UPDATE_VISUAL=1 to regenerate Go reference images")
	}
	regenerateGoReferenceImages(t, findProjectRoot(t))
}

func TestDemoDimensions(t *testing.T) {
	cppFiles := mustListPNGs(t, cppReferenceDir)
	goFiles := listPNGs(goReferenceDir)

	if len(goFiles) == 0 {
		t.Skip("no Go reference images found; run with UPDATE_VISUAL=1 to generate them")
	}

	for _, name := range sortedKeys(cppFiles) {
		name := name
		t.Run(strings.TrimSuffix(name, ".png"), func(t *testing.T) {
			t.Parallel()
			goPath, ok := goFiles[name]
			if !ok {
				t.Skipf("no Go reference for %s", name)
			}
			cppImg, err := framework.LoadImage(cppFiles[name])
			if err != nil {
				t.Fatalf("load cpp reference: %v", err)
			}
			goImg, err := framework.LoadImage(goPath)
			if err != nil {
				t.Fatalf("load go reference: %v", err)
			}
			if cppImg.Bounds().Size() != goImg.Bounds().Size() {
				t.Fatalf("dimension mismatch: cpp=%v go=%v",
					cppImg.Bounds().Size(), goImg.Bounds().Size())
			}
		})
	}
}

func TestGoPortMatchesCPPExamples(t *testing.T) {
	cppFiles := mustListPNGs(t, cppReferenceDir)
	goFiles := listPNGs(goReferenceDir)

	cppNames := sortedKeys(cppFiles)
	goNames := sortedKeys(goFiles)

	if !slices.Equal(cppNames, goNames) {
		if !updateRequested() {
			t.Skipf("Go reference images incomplete or missing; run with UPDATE_VISUAL=1 to generate them\ncpp-only: %s\ngo-only: %s",
				strings.Join(diffNames(cppNames, goNames), ", "),
				strings.Join(diffNames(goNames, cppNames), ", "))
		}
		t.Fatalf("reference set mismatch after update\ncpp-only: %s\ngo-only: %s",
			strings.Join(diffNames(cppNames, goNames), ", "),
			strings.Join(diffNames(goNames, cppNames), ", "))
	}
	if len(cppNames) == 0 {
		t.Skip("no reference images found; run with -update to generate Go references")
	}

	options := framework.ComparisonOptions{
		ExactMatch:        false,
		Tolerance:         10,
		MaxDifferentRatio: 0.01,
		GenerateDiffImage: true,
		IgnoreAlpha:       false,
	}

	for _, name := range cppNames {
		name := name
		t.Run(strings.TrimSuffix(name, ".png"), func(t *testing.T) {
			t.Parallel()
			cppImg, err := framework.LoadImage(cppFiles[name])
			if err != nil {
				t.Fatalf("load cpp reference: %v", err)
			}
			goImg, err := framework.LoadImage(goFiles[name])
			if err != nil {
				t.Fatalf("load go reference: %v", err)
			}

			if cppImg.Bounds().Size() != goImg.Bounds().Size() {
				t.Fatalf("dimension mismatch: cpp=%v go=%v", cppImg.Bounds().Size(), goImg.Bounds().Size())
			}

			result := framework.CompareImages(cppImg, goImg, options)
			if result.Passed {
				return
			}

			diffPath := filepath.Join(cppDiffDir, name)
			if result.DiffImage != nil {
				if err := framework.SaveDiffImage(result.DiffImage, diffPath); err != nil {
					t.Fatalf("save diff image: %v", err)
				}
			}

			t.Fatalf("images differ: ratio=%.4f different_pixels=%d/%d max_diff=%d avg_diff=%.2f diff=%s",
				result.DifferentRatio,
				result.DifferentPixels,
				result.TotalPixels,
				result.MaxDifference,
				result.AverageDifference,
				diffPath,
			)
		})
	}
}

func regenerateGoReferenceImages(t *testing.T, repoRoot string) {
	t.Helper()

	outDir := filepath.Join(repoRoot, "tests", "visual", goReferenceDir)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("create go reference dir: %v", err)
	}
	existing, err := filepath.Glob(filepath.Join(outDir, "*.png"))
	if err != nil {
		t.Fatalf("glob go references: %v", err)
	}
	for _, path := range existing {
		if err := os.Remove(path); err != nil {
			t.Fatalf("remove stale go reference %s: %v", path, err)
		}
	}

	for i, demo := range demoConfigs {
		fmt.Fprintf(os.Stderr, "[%d/%d] generating %s ...\n", i+1, len(demoConfigs), demo.name)
		if err := regenerateSingleGoReference(repoRoot, outDir, demo); err != nil {
			t.Fatalf("generate %s: %v", demo.name, err)
		}
	}
}

func regenerateSingleGoReference(repoRoot, outDir string, demo demoConfig) error {
	if err := tryGenerateFromDir(outDir, demo, filepath.Join(repoRoot, demo.dir), []string{"go", "run", "."}); err == nil {
		return nil
	}
	return tryGenerateFromDir(outDir, demo, repoRoot, []string{"go", "run", "./" + demo.dir})
}

func tryGenerateFromDir(outDir string, demo demoConfig, runDir string, args []string) error {
	stamp, err := os.CreateTemp(runDir, ".demo-stamp-*")
	if err != nil {
		return err
	}
	stampPath := stamp.Name()
	_ = stamp.Close()
	defer os.Remove(stampPath)

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = runDir
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("run %s in %s failed: %w\n%s", strings.Join(args, " "), runDir, err, strings.TrimSpace(string(output)))
	}

	generated, err := findGeneratedPNG(runDir, stampPath)
	if err != nil {
		return fmt.Errorf("find generated png after %s in %s: %w\n%s", strings.Join(args, " "), runDir, err, strings.TrimSpace(string(output)))
	}
	defer os.Remove(generated)

	dstPath := filepath.Join(outDir, demo.name+".png")
	return copyFile(generated, dstPath)
}

func findGeneratedPNG(runDir, stampPath string) (string, error) {
	entries, err := os.ReadDir(runDir)
	if err != nil {
		return "", err
	}

	var found []string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".png" {
			continue
		}
		path := filepath.Join(runDir, entry.Name())
		info, err := os.Stat(path)
		if err != nil {
			return "", err
		}
		stampInfo, err := os.Stat(stampPath)
		if err != nil {
			return "", err
		}
		if info.ModTime().After(stampInfo.ModTime()) {
			found = append(found, path)
		}
	}
	if len(found) == 0 {
		return "", fmt.Errorf("no generated png found")
	}
	slices.Sort(found)
	return found[0], nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func listPNGs(dir string) map[string]string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	files := make(map[string]string, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".png" {
			continue
		}
		files[entry.Name()] = filepath.Join(dir, entry.Name())
	}
	return files
}

func mustListPNGs(t *testing.T, dir string) map[string]string {
	t.Helper()
	files := listPNGs(dir)
	if files == nil {
		t.Fatalf("read dir %s: directory not found or unreadable", dir)
	}
	return files
}

func findProjectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find project root")
		}
		dir = parent
	}
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func diffNames(left, right []string) []string {
	rightSet := make(map[string]struct{}, len(right))
	for _, name := range right {
		rightSet[name] = struct{}{}
	}

	var diff []string
	for _, name := range left {
		if _, ok := rightSet[name]; !ok {
			diff = append(diff, name)
		}
	}
	if len(diff) == 0 {
		return []string{"<none>"}
	}
	return diff
}

func TestDemoParityThresholdIsTight(t *testing.T) {
	opts := framework.ComparisonOptions{
		ExactMatch:        false,
		Tolerance:         10,
		MaxDifferentRatio: 0.01,
	}
	if opts.Tolerance > 16 || opts.MaxDifferentRatio > 0.02 {
		t.Fatalf("demo parity threshold too loose: %+v", opts)
	}
	if opts.MaxDifferentRatio <= 0 {
		t.Fatalf("demo parity ratio must be bounded: %+v", opts)
	}
	if fmt.Sprintf("%.2f", opts.MaxDifferentRatio*100) != "1.00" {
		t.Fatalf("unexpected ratio formatting: %.4f", opts.MaxDifferentRatio)
	}
}
