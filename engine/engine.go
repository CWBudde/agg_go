package engine

import (
	"errors"
	"fmt"
	"image"

	agg "github.com/cwbudde/agg_go"
)

// Kind identifies the rendering engine implementation used by the facade.
type Kind string

const (
	// Port selects the native Go implementation in this repository.
	Port Kind = "port"
	// CPP selects the optional C++ AGG-backed engine.
	CPP Kind = "cpp"
	// AGoGo is an alias for the optional C++ engine name used in planning/docs.
	AGoGo Kind = CPP
)

// String returns the stable string form of the engine kind.
func (k Kind) String() string { return string(k) }

// Config selects the engine implementation used by facade constructors.
type Config struct {
	Kind Kind
}

// Context is the backend-neutral high-level drawing surface exposed by the
// engine facade.
type Context interface {
	Kind() Kind
	Width() int
	Height() int
	Clear(color agg.Color)
	SetColor(color agg.Color)
	SetFillColor(color agg.Color)
	SetStrokeColor(color agg.Color)
	SetLineWidth(width float64)
	SetLineCap(cap agg.LineCap)
	SetLineJoin(join agg.LineJoin)
	AddDash(dashLen, gapLen float64)
	RemoveAllDashes()
	DashStart(offset float64)
	GetDashStart() float64
	SetBlendMode(mode agg.BlendMode)
	GetBlendMode() agg.BlendMode
	FillEvenOdd(evenOdd bool)
	GetFillEvenOdd() bool
	SetLinearGradient(x1, y1, x2, y2 float64, c1, c2 agg.Color)
	SetLinearGradientWithProfile(x1, y1, x2, y2 float64, c1, c2 agg.Color, profile float64)
	SetRadialGradient(cx, cy, radius float64, c1, c2 agg.Color)
	SetRadialGradientWithProfile(cx, cy, radius float64, c1, c2 agg.Color, profile float64)
	SetStrokeLinearGradient(x1, y1, x2, y2 float64, c1, c2 agg.Color)
	SetStrokeLinearGradientWithProfile(x1, y1, x2, y2 float64, c1, c2 agg.Color, profile float64)
	SetStrokeRadialGradient(cx, cy, radius float64, c1, c2 agg.Color)
	SetStrokeRadialGradientWithProfile(cx, cy, radius float64, c1, c2 agg.Color, profile float64)
	GetFillGradientType() agg.GradientType
	GetStrokeGradientType() agg.GradientType
	BeginPath()
	MoveTo(x, y float64)
	LineTo(x, y float64)
	QuadTo(xCtrl, yCtrl, xTo, yTo float64)
	CubicTo(xCtrl1, yCtrl1, xCtrl2, yCtrl2, xTo, yTo float64)
	ClosePath()
	Fill()
	Stroke()
	DrawLine(x1, y1, x2, y2 float64)
	DrawRectangle(x, y, width, height float64)
	FillRectangle(x, y, width, height float64)
	DrawCircle(cx, cy, radius float64)
	FillCircle(cx, cy, radius float64)
	ClipBox(x1, y1, x2, y2 float64)
	GetClipBox() agg.RectD
	Translate(tx, ty float64)
	Rotate(angle float64)
	Scale(sx, sy float64)
	ResetTransform()
	DrawImage(img Image, x, y float64) error
	DrawImageScaled(img Image, x, y, width, height float64) error
	DrawImageRegion(img Image, srcX, srcY, srcW, srcH int, dstX, dstY, dstW, dstH float64) error
	DrawImageQuad(img Image, quad [8]float64) error
	DrawImageRegionQuad(img Image, srcX, srcY, srcW, srcH int, quad [8]float64) error
	LoadFont(fontFile string) error
	SetResolution(dpi uint)
	TextHints(hints bool)
	GetTextHints() bool
	SetTextAlignment(alignX, alignY agg.TextAlignment)
	DrawText(text string, x, y float64) error
	DrawTextAligned(text string, x, y float64, alignment agg.TextAlignment) error
	MeasureText(text string) (width, height float64)
	GetTextBounds(text string) (x, y, width, height float64)
	GetImage() Image
}

// Image is the backend-neutral raster image type used by the engine facade.
type Image interface {
	Kind() Kind
	Width() int
	Height() int
	Premultiply() error
	Demultiply() error
	ToGoImage() *image.RGBA
	ToStandardImage() (image.Image, error)
	SaveToPNG(filename string) error
	SaveToJPEG(filename string, quality int) error
}

// ErrUnavailable is returned when an engine was requested but is not available
// in the current build/runtime environment.
var ErrUnavailable = errors.New("engine unavailable")

// ErrEngineMismatch is returned when a resource from one engine is used with a
// different engine implementation.
var ErrEngineMismatch = errors.New("engine mismatch")

// UnavailableError describes why a specific engine kind is unavailable.
type UnavailableError struct {
	Kind   Kind
	Reason string
}

func (e *UnavailableError) Error() string {
	if e == nil {
		return ErrUnavailable.Error()
	}
	if e.Reason == "" {
		return fmt.Sprintf("%s: %s", ErrUnavailable, e.Kind)
	}
	return fmt.Sprintf("%s: %s (%s)", ErrUnavailable, e.Kind, e.Reason)
}

// Unwrap allows errors.Is(err, ErrUnavailable).
func (e *UnavailableError) Unwrap() error { return ErrUnavailable }

// EngineMismatchError describes an operation using resources from incompatible
// engine implementations.
type EngineMismatchError struct {
	ContextKind  Kind
	ResourceKind Kind
	ResourceType string
}

func (e *EngineMismatchError) Error() string {
	if e == nil {
		return ErrEngineMismatch.Error()
	}
	return fmt.Sprintf(
		"%s: context=%s resource=%s type=%s",
		ErrEngineMismatch,
		e.ContextKind,
		e.ResourceKind,
		e.ResourceType,
	)
}

// Unwrap allows errors.Is(err, ErrEngineMismatch).
func (e *EngineMismatchError) Unwrap() error { return ErrEngineMismatch }

// Available returns the engine kinds supported by the current build.
func Available() []Kind {
	kinds := []Kind{Port}
	if cppAvailable() {
		kinds = append(kinds, CPP)
	}
	return kinds
}

// DefaultKind returns the engine selected when Config.Kind is left unset.
func DefaultKind() Kind { return Port }

// NewContext creates a backend-neutral drawing context for the selected engine.
func NewContext(width, height int, cfg Config) (Context, error) {
	switch normalizeKind(cfg.Kind) {
	case Port:
		return newPortContext(width, height), nil
	case CPP:
		return newCPPContext(width, height)
	default:
		return nil, fmt.Errorf("unknown engine kind %q", cfg.Kind)
	}
}

// NewContextForImage creates a context attached to an existing engine image.
func NewContextForImage(img Image) (Context, error) {
	if img == nil {
		return nil, fmt.Errorf("image is nil")
	}
	switch img.Kind() {
	case Port:
		return newPortContextForImage(img)
	case CPP:
		return newCPPContextForImage(img)
	default:
		return nil, fmt.Errorf("unknown engine kind %q", img.Kind())
	}
}

// NewImage creates a blank engine image owned by the selected engine.
func NewImage(width, height int, cfg Config) (Image, error) {
	switch normalizeKind(cfg.Kind) {
	case Port:
		return newPortImage(width, height)
	case CPP:
		return newCPPImage(width, height)
	default:
		return nil, fmt.Errorf("unknown engine kind %q", cfg.Kind)
	}
}

// NewImageFromGoImage converts a standard library image into an engine image.
func NewImageFromGoImage(src image.Image, cfg Config) (Image, error) {
	switch normalizeKind(cfg.Kind) {
	case Port:
		return newPortImageFromGoImage(src)
	case CPP:
		return newCPPImageFromGoImage(src)
	default:
		return nil, fmt.Errorf("unknown engine kind %q", cfg.Kind)
	}
}

// NewImageFromBuffer creates an engine image attached to a caller-managed RGBA
// buffer.
func NewImageFromBuffer(buf []byte, width, height, stride int, cfg Config) (Image, error) {
	switch normalizeKind(cfg.Kind) {
	case Port:
		return newPortImageFromBuffer(buf, width, height, stride)
	case CPP:
		return newCPPImageFromBuffer(buf, width, height, stride)
	default:
		return nil, fmt.Errorf("unknown engine kind %q", cfg.Kind)
	}
}

// LoadImageFromFile loads an image file into an engine image.
func LoadImageFromFile(filename string, cfg Config) (Image, error) {
	switch normalizeKind(cfg.Kind) {
	case Port:
		return loadPortImageFromFile(filename)
	case CPP:
		return loadCPPImageFromFile(filename)
	default:
		return nil, fmt.Errorf("unknown engine kind %q", cfg.Kind)
	}
}

func normalizeKind(kind Kind) Kind {
	if kind == "" {
		return DefaultKind()
	}
	return kind
}
