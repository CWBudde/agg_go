package timing

import (
	"os"
	"testing"
)

func TestShowTextDefaultsToTrue(t *testing.T) {
	old, hadOld := os.LookupEnv(TextEnv)
	t.Cleanup(func() {
		if hadOld {
			_ = os.Setenv(TextEnv, old)
		} else {
			_ = os.Unsetenv(TextEnv)
		}
	})
	_ = os.Unsetenv(TextEnv)

	if !ShowText() {
		t.Fatal("ShowText() = false with unset env, want true")
	}
}

func TestShowTextCanBeDisabled(t *testing.T) {
	for _, value := range []string{"0", "false", "no", "off"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv(TextEnv, value)
			if ShowText() {
				t.Fatalf("ShowText() = true for %q, want false", value)
			}
		})
	}
}

func TestShowTextCanBeEnabledExplicitly(t *testing.T) {
	for _, value := range []string{"", "1", "true", "yes", "on", "unexpected"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv(TextEnv, value)
			if !ShowText() {
				t.Fatalf("ShowText() = false for %q, want true", value)
			}
		})
	}
}
