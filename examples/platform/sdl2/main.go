//go:build sdl2
// +build sdl2

// Package main is the SDL2 platform backend entry point for the interactive AGG demo.
package main

import (
	"fmt"
	"log"

	"github.com/MeKo-Christian/agg_go/examples/shared/platformdemo"
	"github.com/MeKo-Christian/agg_go/internal/platform"
)

func main() {
	fmt.Println("AGG Interactive Demo — SDL2")

	factory := platform.GetBackendFactory()
	backend, err := factory.CreateBackend(platform.BackendSDL2, platform.PixelFormatRGBA32, false)
	if err != nil {
		log.Fatalf("create SDL2 backend: %v", err)
	}

	if err := platformdemo.New(backend).Run(); err != nil {
		log.Fatalf("run: %v", err)
	}
}
