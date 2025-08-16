// Package config provides configuration definitions for the AGG rendering library.
// This package allows customization of basic types and rendering buffer behavior
// to match different performance and memory requirements.
package config

import (
	"agg_go/internal/buffer"
)

// RenderingBufferType represents the type of rendering buffer to use by default.
type RenderingBufferType int

const (
	// RenderingBufferStandard uses the standard row_accessor equivalent.
	// Provides cheaper creation and destruction with no memory allocations.
	// Best for general use cases where buffer creation/destruction is frequent.
	RenderingBufferStandard RenderingBufferType = iota

	// RenderingBufferCached uses the row_ptr_cache equivalent.
	// Provides faster access for massive pixel operations like blur and image filtering.
	// Uses more memory to cache row pointers but significantly faster for bulk operations.
	RenderingBufferCached
)

// Config holds the global configuration for the AGG library.
type Config struct {
	// DefaultRenderingBufferType specifies which rendering buffer implementation
	// to use for default pixel format typedefs and high-level API.
	DefaultRenderingBufferType RenderingBufferType

	// IntegerOverrides allows overriding basic integer types.
	// This can be useful for platforms with different integer size requirements.
	// If nil, standard Go types are used.
	IntegerOverrides *IntegerTypeOverrides
}

// IntegerTypeOverrides allows customization of basic integer types.
// This is equivalent to AGG's compile-time type redefinition mechanism.
type IntegerTypeOverrides struct {
	// UseInt32AsInt64 forces 64-bit operations to use 32-bit integers.
	// This matches AGG's fallback for compilers without 64-bit support.
	// Will result in overflow in 16-bit-per-component operations but prevents crashes.
	UseInt32AsInt64 bool

	// CustomCoordType allows overriding the default coordinate type.
	// If nil, uses the default int32.
	CustomCoordType *CoordTypeConfig
}

// CoordTypeConfig defines custom coordinate type configuration.
type CoordTypeConfig struct {
	// Use16Bit forces coordinates to use 16-bit integers instead of 32-bit.
	// Reduces memory usage but limits coordinate range to ±32,767.
	Use16Bit bool

	// UseFloat forces coordinates to use floating-point instead of integers.
	// Provides sub-pixel precision but may be slower on some platforms.
	UseFloat bool
}

// Global configuration instance
var globalConfig = Config{
	DefaultRenderingBufferType: RenderingBufferStandard,
	IntegerOverrides:           nil,
}

// SetConfig updates the global configuration.
// This should be called before any AGG operations for consistent behavior.
func SetConfig(cfg Config) {
	globalConfig = cfg
}

// GetConfig returns the current global configuration.
func GetConfig() Config {
	return globalConfig
}

// SetDefaultRenderingBufferType sets the default rendering buffer type.
func SetDefaultRenderingBufferType(bufType RenderingBufferType) {
	globalConfig.DefaultRenderingBufferType = bufType
}

// GetDefaultRenderingBufferType returns the current default rendering buffer type.
func GetDefaultRenderingBufferType() RenderingBufferType {
	return globalConfig.DefaultRenderingBufferType
}

// NewRenderingBuffer creates a rendering buffer using the configured default type.
func NewRenderingBuffer[T any]() interface{} {
	switch globalConfig.DefaultRenderingBufferType {
	case RenderingBufferCached:
		return buffer.NewRenderingBufferCache[T]()
	default:
		return buffer.NewRenderingBuffer[T]()
	}
}

// NewRenderingBufferWithData creates a rendering buffer with data using the configured default type.
func NewRenderingBufferWithData[T any](buf []T, width, height, stride int) interface{} {
	switch globalConfig.DefaultRenderingBufferType {
	case RenderingBufferCached:
		rbc := buffer.NewRenderingBufferCache[T]()
		rbc.Attach(buf, width, height, stride)
		return rbc
	default:
		return buffer.NewRenderingBufferWithData(buf, width, height, stride)
	}
}

// GetCoordType returns the configured coordinate type.
// This allows runtime selection of coordinate precision vs. performance trade-offs.
func GetCoordType() CoordTypeInfo {
	if globalConfig.IntegerOverrides == nil || globalConfig.IntegerOverrides.CustomCoordType == nil {
		return CoordTypeInfo{
			TypeName: "int32",
			Size:     32,
			IsFloat:  false,
		}
	}

	cfg := globalConfig.IntegerOverrides.CustomCoordType
	if cfg.UseFloat {
		return CoordTypeInfo{
			TypeName: "float64",
			Size:     64,
			IsFloat:  true,
		}
	}

	if cfg.Use16Bit {
		return CoordTypeInfo{
			TypeName: "int16",
			Size:     16,
			IsFloat:  false,
		}
	}

	return CoordTypeInfo{
		TypeName: "int32",
		Size:     32,
		IsFloat:  false,
	}
}

// CoordTypeInfo provides information about the configured coordinate type.
type CoordTypeInfo struct {
	TypeName string // Name of the Go type being used
	Size     int    // Size in bits
	IsFloat  bool   // Whether it's a floating-point type
}

// GetInt64Type returns the appropriate 64-bit integer type based on configuration.
func GetInt64Type() Int64TypeInfo {
	if globalConfig.IntegerOverrides != nil && globalConfig.IntegerOverrides.UseInt32AsInt64 {
		return Int64TypeInfo{
			TypeName:     "int32",
			ActualSize:   32,
			WillOverflow: true,
		}
	}

	return Int64TypeInfo{
		TypeName:     "int64",
		ActualSize:   64,
		WillOverflow: false,
	}
}

// Int64TypeInfo provides information about the configured 64-bit integer type.
type Int64TypeInfo struct {
	TypeName     string // Name of the Go type being used
	ActualSize   int    // Actual size in bits
	WillOverflow bool   // Whether overflow may occur in 16-bit-per-component operations
}

// ValidateConfig checks if the current configuration is valid and warns about potential issues.
func ValidateConfig() []ConfigWarning {
	var warnings []ConfigWarning

	if globalConfig.IntegerOverrides != nil {
		if globalConfig.IntegerOverrides.UseInt32AsInt64 {
			warnings = append(warnings, ConfigWarning{
				Type:    WarningOverflow,
				Message: "Using 32-bit integers for 64-bit operations may cause overflow in 16-bit-per-component image processing",
			})
		}

		if cfg := globalConfig.IntegerOverrides.CustomCoordType; cfg != nil {
			if cfg.Use16Bit {
				warnings = append(warnings, ConfigWarning{
					Type:    WarningPrecision,
					Message: "Using 16-bit coordinates limits coordinate range to ±32,767",
				})
			}
			if cfg.UseFloat {
				warnings = append(warnings, ConfigWarning{
					Type:    WarningPerformance,
					Message: "Using floating-point coordinates may reduce performance on some platforms",
				})
			}
		}
	}

	return warnings
}

// ConfigWarning represents a configuration warning.
type ConfigWarning struct {
	Type    WarningType
	Message string
}

// WarningType represents the type of configuration warning.
type WarningType int

const (
	WarningOverflow WarningType = iota
	WarningPrecision
	WarningPerformance
)

// String returns a string representation of the warning type.
func (wt WarningType) String() string {
	switch wt {
	case WarningOverflow:
		return "OVERFLOW"
	case WarningPrecision:
		return "PRECISION"
	case WarningPerformance:
		return "PERFORMANCE"
	default:
		return "UNKNOWN"
	}
}

// PrintableConfig returns a human-readable representation of the current configuration.
func PrintableConfig() string {
	cfg := globalConfig
	result := "AGG Configuration:\n"

	// Rendering buffer type
	switch cfg.DefaultRenderingBufferType {
	case RenderingBufferStandard:
		result += "  Rendering Buffer: Standard (row_accessor equivalent)\n"
	case RenderingBufferCached:
		result += "  Rendering Buffer: Cached (row_ptr_cache equivalent)\n"
	}

	// Integer overrides
	if cfg.IntegerOverrides == nil {
		result += "  Integer Types: Standard Go types\n"
	} else {
		result += "  Integer Types: Custom overrides enabled\n"
		if cfg.IntegerOverrides.UseInt32AsInt64 {
			result += "    - Using int32 for 64-bit operations (overflow possible)\n"
		}
		if coordCfg := cfg.IntegerOverrides.CustomCoordType; coordCfg != nil {
			if coordCfg.UseFloat {
				result += "    - Coordinates: float64\n"
			} else if coordCfg.Use16Bit {
				result += "    - Coordinates: int16\n"
			} else {
				result += "    - Coordinates: int32 (default)\n"
			}
		}
	}

	// Warnings
	warnings := ValidateConfig()
	if len(warnings) > 0 {
		result += "  Warnings:\n"
		for _, w := range warnings {
			result += "    - " + w.Type.String() + ": " + w.Message + "\n"
		}
	}

	return result
}
