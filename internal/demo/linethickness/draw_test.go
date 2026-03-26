package linethickness

import (
	"testing"
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
	w, h := Width, Height
	buf := make([]uint8, w*h*4)
	Draw(buf, w, h, DefaultState())
	if len(buf) != w*h*4 {
		t.Fatalf("unexpected buffer size: %d", len(buf))
	}
}
