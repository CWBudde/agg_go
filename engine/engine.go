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
	SetBlendMode(mode agg.BlendMode)
	FillEvenOdd(evenOdd bool)
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
	Translate(tx, ty float64)
	Rotate(angle float64)
	Scale(sx, sy float64)
	ResetTransform()
	DrawImage(img Image, x, y float64) error
	DrawImageScaled(img Image, x, y, width, height float64) error
	DrawImageQuad(img Image, quad [8]float64) error
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
	SaveToPNG(filename string) error
}

// ErrUnavailable is returned when an engine was requested but is not available
// in the current build/runtime environment.
var ErrUnavailable = errors.New("engine unavailable")

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
