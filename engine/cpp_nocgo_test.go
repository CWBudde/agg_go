//go:build agogo && !cgo

package engine_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/cwbudde/agg_go/engine"
)

func TestCPPUnavailableReasonWithoutCGO(t *testing.T) {
	_, err := engine.NewContext(8, 8, engine.Config{Kind: engine.CPP})
	if err == nil {
		t.Fatal("expected unavailable error")
	}
	if !errors.Is(err, engine.ErrUnavailable) {
		t.Fatalf("expected ErrUnavailable, got %v", err)
	}

	var unavailable *engine.UnavailableError
	if !errors.As(err, &unavailable) {
		t.Fatalf("expected UnavailableError, got %T", err)
	}
	if !strings.Contains(unavailable.Reason, "agogo") || !strings.Contains(unavailable.Reason, "cgo") {
		t.Fatalf("unexpected unavailable reason: %q", unavailable.Reason)
	}
}

func TestCPPNotAdvertisedWithoutCGO(t *testing.T) {
	for _, kind := range engine.Available() {
		if kind == engine.CPP {
			t.Fatal("did not expect C++ engine to be advertised without cgo")
		}
	}
}
