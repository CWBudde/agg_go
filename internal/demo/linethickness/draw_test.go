package linethickness

import (
	"testing"

	agg "github.com/MeKo-Christian/agg_go"
)

func TestStateClamp(t *testing.T) {
	st := State{Thickness: -2, Blur: 4}
	st.Clamp()
	if st.Thickness != 0 {
		t.Fatalf("Thickness = %v, want 0", st.Thickness)
	}
	if st.Blur != 2 {
		t.Fatalf("Blur = %v, want 2", st.Blur)
	}
}

func TestDraw(t *testing.T) {
	ctx := agg.NewContext(640, 480)
	Draw(ctx, DefaultState())
	if len(ctx.GetImage().Data) != 640*480*4 {
		t.Fatalf("unexpected image size: %d", len(ctx.GetImage().Data))
	}
}
