package platform

import "testing"

type testBackendFactory struct {
	created []BackendType
}

func (f *testBackendFactory) CreateBackend(backendType BackendType, format PixelFormat, flipY bool) (PlatformBackend, error) {
	f.created = append(f.created, backendType)
	return NewMockBackend(format, flipY), nil
}

func (f *testBackendFactory) GetAvailableBackends() []BackendType {
	return []BackendType{BackendMock, BackendMacOS}
}

func (f *testBackendFactory) GetDefaultBackend() BackendType {
	return BackendMacOS
}

func TestSetBackendFactory(t *testing.T) {
	orig := GetBackendFactory()
	factory := &testBackendFactory{}
	SetBackendFactory(factory)
	t.Cleanup(func() { SetBackendFactory(orig) })

	if got := GetBackendFactory(); got != factory {
		t.Fatalf("GetBackendFactory() did not return the injected factory")
	}

	backend, err := GetBackendFactory().CreateBackend(BackendMacOS, PixelFormatRGBA32, true)
	if err != nil {
		t.Fatalf("CreateBackend() returned error: %v", err)
	}
	if backend == nil {
		t.Fatal("CreateBackend() returned nil backend")
	}
	if len(factory.created) != 1 || factory.created[0] != BackendMacOS {
		t.Fatalf("factory CreateBackend calls = %v, want [BackendMacOS]", factory.created)
	}
}

func TestMockBackendSurfaceAndHandleContracts(t *testing.T) {
	backend := NewMockBackend(PixelFormatRGBA32, false)

	surface, err := backend.CreateImageSurface(7, 5)
	if err != nil {
		t.Fatalf("CreateImageSurface() error: %v", err)
	}
	if !surface.IsValid() {
		t.Fatal("mock image surface should be valid")
	}
	if surface.GetWidth() != 7 || surface.GetHeight() != 5 {
		t.Fatalf("surface size = %dx%d, want 7x5", surface.GetWidth(), surface.GetHeight())
	}
	if got := len(surface.GetData()); got != 7*5*4 {
		t.Fatalf("surface data len = %d, want %d", got, 7*5*4)
	}

	handle := backend.GetNativeHandle()
	if handle == nil || !handle.IsValid() {
		t.Fatal("mock native handle should be valid")
	}
	if got := handle.GetType(); got != "mock" {
		t.Fatalf("native handle type = %q, want %q", got, "mock")
	}

	if got := backend.Run(); got != 0 {
		t.Fatalf("Run() = %d, want 0", got)
	}

	backend.Delay(5)
}

func TestPlatformSupportLoadSavePPMAndPNG(t *testing.T) {
	formats := []string{".ppm", ".png"}

	for _, ext := range formats {
		t.Run(ext, func(t *testing.T) {
			ps := NewPlatformSupport(PixelFormatRGBA32, false)
			if !ps.CreateImage(0, 2, 2) {
				t.Fatal("CreateImage() failed")
			}

			buf := ps.ImageBuffer(0).Buf()
			copy(buf, []uint8{
				255, 0, 0, 255,
				0, 255, 0, 255,
				0, 0, 255, 255,
				255, 255, 255, 255,
			})

			filename := t.TempDir() + "/roundtrip" + ext
			if !ps.SaveImage(0, filename) {
				t.Fatalf("SaveImage(%q) failed", filename)
			}

			loaded := NewPlatformSupport(PixelFormatRGBA32, false)
			if !loaded.LoadImage(0, filename) {
				t.Fatalf("LoadImage(%q) failed", filename)
			}

			got := loaded.ImageBuffer(0)
			if got.Width() != 2 || got.Height() != 2 {
				t.Fatalf("loaded image size = %dx%d, want 2x2", got.Width(), got.Height())
			}
			if got.Buf() == nil || len(got.Buf()) != len(buf) {
				t.Fatalf("loaded image buffer len = %d, want %d", len(got.Buf()), len(buf))
			}

			// Check one pixel that should survive the round trip exactly.
			if px := got.Buf()[:4]; px[0] != 255 || px[1] != 0 || px[2] != 0 {
				t.Fatalf("first pixel after %s round trip = %v, want red", ext, px)
			}
		})
	}
}

func TestSupportedImageExtensions(t *testing.T) {
	ps := NewPlatformSupport(PixelFormatRGB24, false)
	got := ps.SupportedImageExtensions()
	want := []string{".bmp", ".ppm", ".png"}
	if len(got) != len(want) {
		t.Fatalf("SupportedImageExtensions len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("SupportedImageExtensions[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
