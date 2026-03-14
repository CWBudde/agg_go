package color

import "testing"

func TestNewRGBAFromRGBA8(t *testing.T) {
	got := NewRGBAFromRGBA8(128, 64, 192, 255)

	if got.R != 128.0/255.0 || got.G != 64.0/255.0 || got.B != 192.0/255.0 || got.A != 1.0 {
		t.Fatalf("NewRGBAFromRGBA8() = %+v", got)
	}
}

func TestNewRGBAFromGray8(t *testing.T) {
	got := NewRGBAFromGray8(128, 64)

	if got.R != 128.0/255.0 || got.G != 128.0/255.0 || got.B != 128.0/255.0 || got.A != 64.0/255.0 {
		t.Fatalf("NewRGBAFromGray8() = %+v", got)
	}
}

func TestNewRGBAFromGray16(t *testing.T) {
	got := NewRGBAFromGray16(32768, 49152)

	if got.R != 32768.0/65535.0 || got.G != 32768.0/65535.0 || got.B != 32768.0/65535.0 || got.A != 49152.0/65535.0 {
		t.Fatalf("NewRGBAFromGray16() = %+v", got)
	}
}
