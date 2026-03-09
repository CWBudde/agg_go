//go:build x11 || sdl2

package demorunner

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"strings"

	agg "agg_go"
	"github.com/MeKo-Christian/agg_go/internal/platform"
)

// Run opens a platform window (SDL2 preferred, X11 fallback) and runs the
// demo event loop:
//   - Escape / window-close exits.
//   - S saves a PNG screenshot.
//   - Mouse and key events are forwarded to MouseHandler / KeyHandler.
func Run(cfg Config, demo Demo) {
	factory := platform.GetBackendFactory()
	backend, err := factory.CreateBackend(
		factory.GetDefaultBackend(),
		platform.PixelFormatRGBA32,
		false,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "demorunner: create backend: %v\n", err)
		os.Exit(1)
	}

	h := &handler{
		backend: backend,
		ps:      platform.NewPlatformSupport(platform.PixelFormatRGBA32, false),
		ctx:     agg.NewContext(cfg.Width, cfg.Height),
		cfg:     cfg,
		demo:    demo,
		running: true,
	}
	h.ps.Caption(cfg.Title)

	if setter, ok := backend.(platform.EventCallbackSetter); ok {
		setter.SetEventCallback(h)
	}
	if err := h.ps.Init(cfg.Width, cfg.Height, platform.WindowResize); err != nil {
		fmt.Fprintf(os.Stderr, "demorunner: platform support init: %v\n", err)
		os.Exit(1)
	}
	if err := backend.Init(cfg.Width, cfg.Height, platform.WindowResize); err != nil {
		fmt.Fprintf(os.Stderr, "demorunner: backend init: %v\n", err)
		os.Exit(1)
	}
	defer backend.Destroy()

	for h.running {
		if !backend.PollEvents() {
			break
		}
		h.onIdle()
	}
}

// handler bridges platform events to the Demo interface.
type handler struct {
	platform.BaseEventHandler
	backend platform.PlatformBackend
	ps      *platform.PlatformSupport
	ctx     *agg.Context
	cfg     Config
	demo    Demo
	running bool
}

func (h *handler) OnDraw() {
	h.demo.Render(h.ctx)
	h.blit()
}

func (h *handler) OnResize(width, height int) {
	h.ctx = agg.NewContext(width, height)
	h.backend.ForceRedraw()
}

func (h *handler) OnMouseMove(x, y int, flags platform.InputFlags) {
	if md, ok := h.demo.(MouseHandler); ok {
		if md.OnMouseMove(x, y, toButtons(flags)) {
			h.backend.ForceRedraw()
		}
	}
}

func (h *handler) OnMouseButtonDown(x, y int, flags platform.InputFlags) {
	if md, ok := h.demo.(MouseHandler); ok {
		if md.OnMouseDown(x, y, toButtons(flags)) {
			h.backend.ForceRedraw()
		}
	}
}

func (h *handler) OnMouseButtonUp(x, y int, flags platform.InputFlags) {
	if md, ok := h.demo.(MouseHandler); ok {
		if md.OnMouseUp(x, y, toButtons(flags)) {
			h.backend.ForceRedraw()
		}
	}
}

func (h *handler) OnKey(_ int, _ int, key platform.KeyCode, _ platform.InputFlags) {
	switch key {
	case platform.KeyEscape:
		h.running = false
	case platform.KeyCode('s'), platform.KeyCode('S'):
		h.saveScreenshot()
	default:
		if kd, ok := h.demo.(KeyHandler); ok {
			if kd.OnKey(rune(key)) {
				h.backend.ForceRedraw()
			}
		}
	}
}

func (h *handler) OnDestroy() { h.running = false }

func (h *handler) onIdle() {
	animated := false
	if a, ok := h.demo.(Animated); ok {
		animated = a.IsAnimated()
	}
	if animated {
		h.backend.ForceRedraw()
		h.backend.Delay(16)
	}
}

// blit copies the agg.Context pixel buffer into the platform window buffer
// and presents it.
func (h *handler) blit() {
	img := h.ctx.GetImage()
	winBuf := h.ps.WindowBuffer()
	src := img.Data
	dst := winBuf.Buf()

	w := winBuf.Width()
	if len(src) == len(dst) {
		copy(dst, src)
	} else {
		// Stride mismatch: copy row by row.
		srcStride := img.Width() * 4
		dstStride := winBuf.Stride()
		if dstStride < 0 {
			dstStride = -dstStride
		}
		for y := range winBuf.Height() {
			copy(dst[y*dstStride:y*dstStride+w*4], src[y*srcStride:y*srcStride+w*4])
		}
	}
	_ = h.backend.UpdateWindow(winBuf)
}

func (h *handler) saveScreenshot() {
	filename := strings.ReplaceAll(strings.ToLower(h.cfg.Title), " ", "_") + ".png"
	img := h.ctx.GetImage()
	goImg := image.NewRGBA(image.Rect(0, 0, img.Width(), img.Height()))
	copy(goImg.Pix, img.Data)
	f, err := os.Create(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "screenshot: %v\n", err)
		return
	}
	defer f.Close()
	if err := png.Encode(f, goImg); err != nil {
		fmt.Fprintf(os.Stderr, "screenshot: encode: %v\n", err)
		return
	}
	fmt.Printf("screenshot saved to %s\n", filename)
}

func toButtons(flags platform.InputFlags) Buttons {
	return Buttons{
		Left:   flags.HasMouseLeft(),
		Right:  flags.HasMouseRight(),
		Middle: false, // platform.InputFlags has no middle-button flag
	}
}
