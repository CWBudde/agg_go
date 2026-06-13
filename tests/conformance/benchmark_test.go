package conformance

import (
	"testing"

	"github.com/cwbudde/agg_go/engine"
	"github.com/cwbudde/agg_go/engine/scene"
)

// BenchmarkCorpusRender measures the full per-render cost (context creation +
// draw + image readback) of each scene through every available engine. Default
// `go test -bench` runs the port/* sub-benchmarks; building with
// -tags "agogo aggreal" adds the cpp/* sub-benchmarks via engine.Available().
//
// Note: cpp/* timings include cgo call overhead and are not directly comparable
// to port/* timings.
func BenchmarkCorpusRender(b *testing.B) {
	for _, kind := range engine.Available() {
		assets, err := scene.BuildAssets(kind)
		if err != nil {
			b.Fatalf("BuildAssets(%s): %v", kind, err)
		}
		for _, s := range scene.All() {
			if !s.SupportedBy(kind) {
				continue
			}
			b.Run(string(kind)+"/"+s.Name, func(b *testing.B) {
				// Probe once so an unavailable asset or unsupported capability
				// skips instead of failing b.N times.
				if _, err := render(s, kind, assets); err != nil {
					if skip, msg := skipReason(err); skip {
						b.Skip(msg)
					}
					b.Fatalf("probe render: %v", err)
				}
				b.ReportAllocs()
				b.ResetTimer()
				for range b.N {
					ctx, err := engine.NewContext(s.Width, s.Height, engine.Config{Kind: kind})
					if err != nil {
						b.Fatalf("NewContext: %v", err)
					}
					if err := s.Draw(ctx, assets); err != nil {
						b.Fatalf("Draw: %v", err)
					}
					_ = ctx.GetImage()
				}
			})
		}
	}
}
