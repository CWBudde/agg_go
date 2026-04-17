//go:build agogo && cgo

package cppbridge

import (
	"strings"
	"testing"
)

func TestCurrentMetadataReportsStubBridge(t *testing.T) {
	meta := CurrentMetadata()
	if !meta.Stub {
		t.Fatal("expected agogo bridge to report stub mode")
	}
	if meta.BuildID == "" {
		t.Fatal("expected bridge build id to be set")
	}
}

func TestProbeReturnsStubFailure(t *testing.T) {
	err := Probe()
	if err == nil {
		t.Fatal("expected stub bridge probe to fail")
	}
	if !strings.Contains(err.Error(), "stub") {
		t.Fatalf("expected stub failure message, got %q", err.Error())
	}
}
