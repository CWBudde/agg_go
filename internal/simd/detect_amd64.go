//go:build amd64

package simd

import (
	"runtime"

	"golang.org/x/sys/cpu"
)

func detectFeaturesImpl() Features {
	return Features{
		HasSSE:       true,
		HasSSE2:      cpu.X86.HasSSE2,
		HasSSE3:      cpu.X86.HasSSE3,
		HasSSSE3:     cpu.X86.HasSSSE3,
		HasSSE41:     cpu.X86.HasSSE41,
		HasAVX:       cpu.X86.HasAVX,
		HasAVX2:      cpu.X86.HasAVX2,
		HasAVX512:    cpu.X86.HasAVX512,
		Architecture: runtime.GOARCH,
	}
}
