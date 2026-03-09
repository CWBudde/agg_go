//go:build x11
// +build x11

// Package main is the X11 platform backend entry point for the interactive AGG demo.
package main

import (
	"fmt"
	"log"

	"github.com/MeKo-Christian/agg_go/examples/shared/platformdemo"
	"github.com/MeKo-Christian/agg_go/internal/platform"
)

func main() {
	fmt.Println("AGG Interactive Demo — X11")

	factory := platform.GetBackendFactory()
	backend, err := factory.CreateBackend(platform.BackendX11, platform.PixelFormatRGBA32, false)
	if err != nil {
		log.Fatalf("create X11 backend: %v", err)
	}

	if err := platformdemo.New(backend).Run(); err != nil {
		log.Fatalf("run: %v", err)
	}
}
