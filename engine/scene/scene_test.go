package scene_test

import (
	"errors"
	"testing"

	"github.com/cwbudde/agg_go/engine"
	"github.com/cwbudde/agg_go/engine/scene"
)

func TestAllReturnsStableNonEmptyCorpus(t *testing.T) {
	a := scene.All()
	if len(a) == 0 {
		t.Fatal("corpus is empty")
	}
	b := scene.All()
	for i := range a {
		if a[i].Name != b[i].Name {
			t.Fatalf("corpus order not stable at %d: %q vs %q", i, a[i].Name, b[i].Name)
		}
	}
	// Names must be unique.
	seen := map[string]bool{}
	for _, s := range a {
		if seen[s.Name] {
			t.Fatalf("duplicate scene name %q", s.Name)
		}
		seen[s.Name] = true
		if s.Width <= 0 || s.Height <= 0 {
			t.Fatalf("scene %q has non-positive size %dx%d", s.Name, s.Width, s.Height)
		}
		if s.Draw == nil {
			t.Fatalf("scene %q has nil Draw", s.Name)
		}
	}
}

func TestByNameAndFilter(t *testing.T) {
	if _, ok := scene.ByName("solid_fill_stroke"); !ok {
		t.Fatal("expected solid_fill_stroke in corpus")
	}
	if _, ok := scene.ByName("does_not_exist"); ok {
		t.Fatal("unexpected scene found")
	}
	if got := len(scene.Filter("")); got != len(scene.All()) {
		t.Fatalf("Filter(\"\") = %d scenes, want %d", got, len(scene.All()))
	}
	comp := scene.Filter("compositing")
	if len(comp) == 0 {
		t.Fatal("expected compositing scenes")
	}
	for _, s := range comp {
		if got, want := s.Name[:11], "compositing"; got != want {
			t.Fatalf("Filter returned non-matching scene %q", s.Name)
		}
	}
}

// TestPortRendersEntireCorpus renders every scene through the always-available
// Port engine and asserts it produces an image of the declared size. Scenes
// whose assets are unavailable (e.g. no font) are skipped, not failed.
func TestPortRendersEntireCorpus(t *testing.T) {
	assets, err := scene.BuildAssets(engine.Port)
	if err != nil {
		t.Fatalf("BuildAssets(Port): %v", err)
	}
	for _, s := range scene.All() {
		t.Run(s.Name, func(t *testing.T) {
			if !s.SupportedBy(engine.Port) {
				t.Skipf("port lacks caps %v", s.Caps)
			}
			ctx, err := engine.NewContext(s.Width, s.Height, engine.Config{Kind: engine.Port})
			if err != nil {
				t.Fatalf("NewContext: %v", err)
			}
			if err := s.Draw(ctx, assets); err != nil {
				if errors.Is(err, scene.ErrAssetUnavailable) {
					t.Skipf("asset unavailable: %v", err)
				}
				t.Fatalf("Draw: %v", err)
			}
			img := ctx.GetImage().ToGoImage()
			if got := img.Bounds().Dx(); got != s.Width {
				t.Fatalf("width = %d, want %d", got, s.Width)
			}
			if got := img.Bounds().Dy(); got != s.Height {
				t.Fatalf("height = %d, want %d", got, s.Height)
			}
		})
	}
}
