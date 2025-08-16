package config

import (
	"strings"
	"testing"
)

func TestDefaultConfiguration(t *testing.T) {
	// Reset to default
	SetConfig(Config{
		DefaultRenderingBufferType: RenderingBufferStandard,
		IntegerOverrides:           nil,
	})

	cfg := GetConfig()
	if cfg.DefaultRenderingBufferType != RenderingBufferStandard {
		t.Errorf("Expected default rendering buffer type to be Standard, got %v", cfg.DefaultRenderingBufferType)
	}

	if cfg.IntegerOverrides != nil {
		t.Error("Expected no integer overrides by default")
	}
}

func TestSetDefaultRenderingBufferType(t *testing.T) {
	// Test setting to cached
	SetDefaultRenderingBufferType(RenderingBufferCached)
	if GetDefaultRenderingBufferType() != RenderingBufferCached {
		t.Error("Failed to set rendering buffer type to Cached")
	}

	// Test setting back to standard
	SetDefaultRenderingBufferType(RenderingBufferStandard)
	if GetDefaultRenderingBufferType() != RenderingBufferStandard {
		t.Error("Failed to set rendering buffer type to Standard")
	}
}

func TestGetCoordTypeDefault(t *testing.T) {
	// Reset to default
	SetConfig(Config{
		DefaultRenderingBufferType: RenderingBufferStandard,
		IntegerOverrides:           nil,
	})

	coordInfo := GetCoordType()
	if coordInfo.TypeName != "int32" {
		t.Errorf("Expected default coord type to be int32, got %s", coordInfo.TypeName)
	}
	if coordInfo.Size != 32 {
		t.Errorf("Expected default coord size to be 32, got %d", coordInfo.Size)
	}
	if coordInfo.IsFloat {
		t.Error("Expected default coord type to not be float")
	}
}

func TestGetCoordTypeCustom16Bit(t *testing.T) {
	SetConfig(Config{
		DefaultRenderingBufferType: RenderingBufferStandard,
		IntegerOverrides: &IntegerTypeOverrides{
			CustomCoordType: &CoordTypeConfig{
				Use16Bit: true,
			},
		},
	})

	coordInfo := GetCoordType()
	if coordInfo.TypeName != "int16" {
		t.Errorf("Expected coord type to be int16, got %s", coordInfo.TypeName)
	}
	if coordInfo.Size != 16 {
		t.Errorf("Expected coord size to be 16, got %d", coordInfo.Size)
	}
	if coordInfo.IsFloat {
		t.Error("Expected coord type to not be float")
	}
}

func TestGetCoordTypeCustomFloat(t *testing.T) {
	SetConfig(Config{
		DefaultRenderingBufferType: RenderingBufferStandard,
		IntegerOverrides: &IntegerTypeOverrides{
			CustomCoordType: &CoordTypeConfig{
				UseFloat: true,
			},
		},
	})

	coordInfo := GetCoordType()
	if coordInfo.TypeName != "float64" {
		t.Errorf("Expected coord type to be float64, got %s", coordInfo.TypeName)
	}
	if coordInfo.Size != 64 {
		t.Errorf("Expected coord size to be 64, got %d", coordInfo.Size)
	}
	if !coordInfo.IsFloat {
		t.Error("Expected coord type to be float")
	}
}

func TestGetInt64TypeDefault(t *testing.T) {
	// Reset to default
	SetConfig(Config{
		DefaultRenderingBufferType: RenderingBufferStandard,
		IntegerOverrides:           nil,
	})

	int64Info := GetInt64Type()
	if int64Info.TypeName != "int64" {
		t.Errorf("Expected int64 type to be int64, got %s", int64Info.TypeName)
	}
	if int64Info.ActualSize != 64 {
		t.Errorf("Expected int64 size to be 64, got %d", int64Info.ActualSize)
	}
	if int64Info.WillOverflow {
		t.Error("Expected int64 to not overflow")
	}
}

func TestGetInt64TypeOverride(t *testing.T) {
	SetConfig(Config{
		DefaultRenderingBufferType: RenderingBufferStandard,
		IntegerOverrides: &IntegerTypeOverrides{
			UseInt32AsInt64: true,
		},
	})

	int64Info := GetInt64Type()
	if int64Info.TypeName != "int32" {
		t.Errorf("Expected int64 type to be int32, got %s", int64Info.TypeName)
	}
	if int64Info.ActualSize != 32 {
		t.Errorf("Expected int64 size to be 32, got %d", int64Info.ActualSize)
	}
	if !int64Info.WillOverflow {
		t.Error("Expected int64 override to cause overflow")
	}
}

func TestValidateConfigNoWarnings(t *testing.T) {
	// Reset to default
	SetConfig(Config{
		DefaultRenderingBufferType: RenderingBufferStandard,
		IntegerOverrides:           nil,
	})

	warnings := ValidateConfig()
	if len(warnings) != 0 {
		t.Errorf("Expected no warnings for default config, got %d", len(warnings))
	}
}

func TestValidateConfigWithWarnings(t *testing.T) {
	SetConfig(Config{
		DefaultRenderingBufferType: RenderingBufferStandard,
		IntegerOverrides: &IntegerTypeOverrides{
			UseInt32AsInt64: true,
			CustomCoordType: &CoordTypeConfig{
				Use16Bit: true,
			},
		},
	})

	warnings := ValidateConfig()
	if len(warnings) != 2 {
		t.Errorf("Expected 2 warnings, got %d", len(warnings))
	}

	// Check for overflow warning
	foundOverflow := false
	foundPrecision := false
	for _, w := range warnings {
		if w.Type == WarningOverflow {
			foundOverflow = true
		}
		if w.Type == WarningPrecision {
			foundPrecision = true
		}
	}

	if !foundOverflow {
		t.Error("Expected overflow warning")
	}
	if !foundPrecision {
		t.Error("Expected precision warning")
	}
}

func TestValidateConfigFloatWarning(t *testing.T) {
	SetConfig(Config{
		DefaultRenderingBufferType: RenderingBufferStandard,
		IntegerOverrides: &IntegerTypeOverrides{
			CustomCoordType: &CoordTypeConfig{
				UseFloat: true,
			},
		},
	})

	warnings := ValidateConfig()
	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(warnings))
	}

	if warnings[0].Type != WarningPerformance {
		t.Errorf("Expected performance warning, got %v", warnings[0].Type)
	}
}

func TestWarningTypeString(t *testing.T) {
	tests := []struct {
		wt       WarningType
		expected string
	}{
		{WarningOverflow, "OVERFLOW"},
		{WarningPrecision, "PRECISION"},
		{WarningPerformance, "PERFORMANCE"},
		{WarningType(999), "UNKNOWN"},
	}

	for _, test := range tests {
		if test.wt.String() != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, test.wt.String())
		}
	}
}

func TestPrintableConfig(t *testing.T) {
	// Test default config
	SetConfig(Config{
		DefaultRenderingBufferType: RenderingBufferStandard,
		IntegerOverrides:           nil,
	})

	output := PrintableConfig()
	if !strings.Contains(output, "Standard") {
		t.Error("Expected printable config to mention Standard rendering buffer")
	}
	if !strings.Contains(output, "Standard Go types") {
		t.Error("Expected printable config to mention standard Go types")
	}

	// Test custom config
	SetConfig(Config{
		DefaultRenderingBufferType: RenderingBufferCached,
		IntegerOverrides: &IntegerTypeOverrides{
			UseInt32AsInt64: true,
			CustomCoordType: &CoordTypeConfig{
				UseFloat: true,
			},
		},
	})

	output = PrintableConfig()
	if !strings.Contains(output, "Cached") {
		t.Error("Expected printable config to mention Cached rendering buffer")
	}
	if !strings.Contains(output, "Custom overrides enabled") {
		t.Error("Expected printable config to mention custom overrides")
	}
	if !strings.Contains(output, "float64") {
		t.Error("Expected printable config to mention float64 coordinates")
	}
	if !strings.Contains(output, "Warnings") {
		t.Error("Expected printable config to include warnings section")
	}
}

func TestNewRenderingBufferDefaultType(t *testing.T) {
	// Test standard buffer creation
	SetDefaultRenderingBufferType(RenderingBufferStandard)
	buf := NewRenderingBuffer[uint8]()
	if buf == nil {
		t.Error("Failed to create standard rendering buffer")
	}

	// Test cached buffer creation
	SetDefaultRenderingBufferType(RenderingBufferCached)
	bufCached := NewRenderingBuffer[uint8]()
	if bufCached == nil {
		t.Error("Failed to create cached rendering buffer")
	}
}

func TestNewRenderingBufferWithDataDefaultType(t *testing.T) {
	data := make([]uint8, 1000)

	// Test standard buffer creation
	SetDefaultRenderingBufferType(RenderingBufferStandard)
	buf := NewRenderingBufferWithData(data, 100, 10, 100)
	if buf == nil {
		t.Error("Failed to create standard rendering buffer with data")
	}

	// Test cached buffer creation
	SetDefaultRenderingBufferType(RenderingBufferCached)
	bufCached := NewRenderingBufferWithData(data, 100, 10, 100)
	if bufCached == nil {
		t.Error("Failed to create cached rendering buffer with data")
	}
}
