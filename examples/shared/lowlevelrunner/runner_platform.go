//go:build x11 || sdl2

package lowlevelrunner

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"strings"

	agg "github.com/MeKo-Christian/agg_go"
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
		fmt.Fprintf(os.Stderr, "lowlevelrunner: create backend: %v\n", err)
		os.Exit(1)
	}

	// Match C++ platform_support flip_y: negative stride means row 0 is at the
	// physical bottom of the buffer.  The platform blit reads img.Data
	// top-to-bottom without any row reversal, so the Y-axis is naturally flipped.
	stride := cfg.Width * 4
	if cfg.FlipY {
		stride = -stride
	}
	h := &handler{
		backend: backend,
		ps:      platform.NewPlatformSupport(platform.PixelFormatRGBA32, false),
		img:     agg.NewImage(make([]uint8, cfg.Width*cfg.Height*4), cfg.Width, cfg.Height, stride),
		cfg:     cfg,
		demo:    demo,
		running: true,
	}
	h.ps.Caption(cfg.Title)

	if setter, ok := backend.(platform.EventCallbackSetter); ok {
		setter.SetEventCallback(h)
	}
	if err := h.ps.Init(cfg.Width, cfg.Height, platform.WindowResize); err != nil {
		fmt.Fprintf(os.Stderr, "lowlevelrunner: platform support init: %v\n", err)
		os.Exit(1)
	}
	if err := backend.Init(cfg.Width, cfg.Height, platform.WindowResize); err != nil {
		fmt.Fprintf(os.Stderr, "lowlevelrunner: backend init: %v\n", err)
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
	img     *agg.Image
	cfg     Config
	demo    Demo
	running bool
}

func (h *handler) OnInit() {
	if initDemo, ok := h.demo.(InitHandler); ok {
		initDemo.OnInit()
	}
	h.backend.ForceRedraw()
}

func (h *handler) OnDraw() {
	h.demo.Render(h.img)
	h.blit()
}

func (h *handler) OnResize(width, height int) {
	stride := width * 4
	if h.cfg.FlipY {
		stride = -stride
	}
	h.img.Attach(make([]uint8, width*height*4), width, height, stride)
	h.backend.ForceRedraw()
}

func (h *handler) flipMouseY(y int) int {
	if h.cfg.FlipY {
		return h.img.Height() - 1 - y
	}
	return y
}

func (h *handler) OnMouseMove(x, y int, flags platform.InputFlags) {
	if md, ok := h.demo.(MouseHandler); ok {
		if md.OnMouseMove(x, h.flipMouseY(y), toButtons(flags)) {
			h.backend.ForceRedraw()
		}
	}
}

func (h *handler) OnMouseButtonDown(x, y int, flags platform.InputFlags) {
	if md, ok := h.demo.(MouseHandler); ok {
		if md.OnMouseDown(x, h.flipMouseY(y), toButtons(flags)) {
			h.backend.ForceRedraw()
		}
	}
}

func (h *handler) OnMouseButtonUp(x, y int, flags platform.InputFlags) {
	if md, ok := h.demo.(MouseHandler); ok {
		if md.OnMouseUp(x, h.flipMouseY(y), toButtons(flags)) {
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
	if idleDemo, ok := h.demo.(IdleHandler); ok {
		idleDemo.OnIdle()
	}

	animated := false
	if a, ok := h.demo.(Animated); ok {
		animated = a.IsAnimated()
	}
	if animated {
		h.backend.ForceRedraw()
		h.backend.Delay(16)
	}
}

// blit copies the raw image buffer into the platform window buffer and presents it.
// img.Data is always read sequentially top-to-bottom; when FlipY=true the image
// was created with negative stride so rbuf row 0 is physically at the bottom of
// img.Data — no additional row-reversal is needed here.
func (h *handler) blit() {
	winBuf := h.ps.WindowBuffer()
	src := h.img.Data
	dst := winBuf.Buf()

	srcStride := h.img.Width() * 4
	dstStride := winBuf.Stride()
	if dstStride < 0 {
		dstStride = -dstStride
	}
	for y := range winBuf.Height() {
		srcOff := y * srcStride
		dstOff := y * dstStride
		for x := range winBuf.Width() {
			srcIdx := srcOff + x*4
			dstIdx := dstOff + x*4
			dst[dstIdx] = src[srcIdx]
			dst[dstIdx+1] = src[srcIdx+1]
			dst[dstIdx+2] = src[srcIdx+2]
			dst[dstIdx+3] = 255
		}
	}
	_ = h.backend.UpdateWindow(winBuf)
}

func (h *handler) saveScreenshot() {
	filename := strings.ReplaceAll(strings.ToLower(h.cfg.Title), " ", "_") + ".png"
	goImg := image.NewRGBA(image.Rect(0, 0, h.img.Width(), h.img.Height()))
	srcStride := h.img.Width() * 4
	for y := range h.img.Height() {
		srcOff := y * srcStride
		dstOff := y * goImg.Stride
		for x := range h.img.Width() {
			srcIdx := srcOff + x*4
			dstIdx := dstOff + x*4
			goImg.Pix[dstIdx] = h.img.Data[srcIdx]
			goImg.Pix[dstIdx+1] = h.img.Data[srcIdx+1]
			goImg.Pix[dstIdx+2] = h.img.Data[srcIdx+2]
			goImg.Pix[dstIdx+3] = 255
		}
	}
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
