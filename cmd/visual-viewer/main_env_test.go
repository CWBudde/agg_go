package main

import (
	"slices"
	"testing"

	"github.com/cwbudde/agg_go/internal/demo/timing"
)

func TestDemoCommandEnvDisablesTimingText(t *testing.T) {
	env := demoCommandEnv([]string{"PATH=/bin"})
	if !slices.Contains(env, "PATH=/bin") {
		t.Fatalf("demoCommandEnv dropped existing env: %v", env)
	}
	if !slices.Contains(env, timing.TextEnv+"=0") {
		t.Fatalf("demoCommandEnv missing timing suppression flag %q in %v", timing.TextEnv+"=0", env)
	}
}
