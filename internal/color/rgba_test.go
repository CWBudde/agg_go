package color

import "testing"

func TestNewRGBAFromRGBA8(t *testing.T) {
	got := NewRGBAFromRGBA8(128, 64, 192, 255)

	if got.R != 128.0/255.0 || got.G != 64.0/255.0 || got.B != 192.0/255.0 || got.A != 1.0 {
		t.Fatalf("NewRGBAFromRGBA8() = %+v", got)
	}
}
