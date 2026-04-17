package engine

import (
	"errors"
	"fmt"
)

// Capability identifies a high-level feature that an engine implementation may
// support through the facade.
type Capability string

const (
	CapabilitySolidStyle   Capability = "solid_style"
	CapabilityPath         Capability = "path"
	CapabilityTransforms   Capability = "transforms"
	CapabilityClipBox      Capability = "clip_box"
	CapabilityCompositing  Capability = "compositing"
	CapabilityImageDraw    Capability = "image_draw"
	CapabilityImageExport  Capability = "image_export"
	CapabilityImageInterop Capability = "image_interop"
	CapabilityGradients    Capability = "gradients"
	CapabilityText         Capability = "text"
	CapabilityDashedStroke Capability = "dashed_stroke"
)

var portCapabilities = []Capability{
	CapabilitySolidStyle,
	CapabilityPath,
	CapabilityTransforms,
	CapabilityClipBox,
	CapabilityCompositing,
	CapabilityImageDraw,
	CapabilityImageExport,
	CapabilityImageInterop,
	CapabilityGradients,
	CapabilityText,
}

// ErrUnsupportedCapability is returned when a known engine does not implement a
// requested facade capability.
var ErrUnsupportedCapability = errors.New("engine capability unsupported")

// UnsupportedCapabilityError describes an unsupported high-level facade
// capability on a specific engine.
type UnsupportedCapabilityError struct {
	Kind       Kind
	Capability Capability
	Operation  string
}

func (e *UnsupportedCapabilityError) Error() string {
	if e == nil {
		return ErrUnsupportedCapability.Error()
	}
	if e.Operation == "" {
		return fmt.Sprintf("%s: %s lacks %s", ErrUnsupportedCapability, e.Kind, e.Capability)
	}
	return fmt.Sprintf("%s: %s lacks %s for %s", ErrUnsupportedCapability, e.Kind, e.Capability, e.Operation)
}

// Unwrap allows errors.Is(err, ErrUnsupportedCapability).
func (e *UnsupportedCapabilityError) Unwrap() error { return ErrUnsupportedCapability }

// Capabilities returns the currently supported high-level facade capabilities
// for the selected engine kind.
func Capabilities(kind Kind) ([]Capability, error) {
	switch normalizeKind(kind) {
	case Port:
		out := make([]Capability, len(portCapabilities))
		copy(out, portCapabilities)
		return out, nil
	case CPP:
		if !cppAvailable() {
			return nil, &UnavailableError{
				Kind:   CPP,
				Reason: "the C++ engine is not implemented in this repository yet",
			}
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown engine kind %q", kind)
	}
}

// Supports reports whether the selected engine kind supports a specific facade
// capability.
func Supports(kind Kind, capability Capability) bool {
	caps, err := Capabilities(kind)
	if err != nil {
		return false
	}
	for _, cap := range caps {
		if cap == capability {
			return true
		}
	}
	return false
}

// RequireCapability returns a typed error when a known engine lacks a requested
// facade capability.
func RequireCapability(kind Kind, capability Capability, operation string) error {
	caps, err := Capabilities(kind)
	if err != nil {
		return err
	}
	for _, cap := range caps {
		if cap == capability {
			return nil
		}
	}
	return &UnsupportedCapabilityError{
		Kind:       normalizeKind(kind),
		Capability: capability,
		Operation:  operation,
	}
}
