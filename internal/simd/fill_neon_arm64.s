//go:build arm64 && !purego

#include "textflag.h"

// func fillRGBANEONAsm(dst []byte, pixel uint32, count int)
//
// Fills count tightly-packed 4-byte RGBA pixels in dst with the 32-bit pixel
// value. Uses NEON to store 4 pixels (16 bytes) per iteration.
//
// ABI0 stack layout ($0-40):
//   dst_base  +0(FP)   8 bytes
//   dst_len   +8(FP)   8 bytes
//   dst_cap  +16(FP)   8 bytes
//   pixel    +24(FP)   4 bytes (uint32)
//   count    +32(FP)   8 bytes (int)
TEXT ·fillRGBANEONAsm(SB), NOSPLIT, $0-40
	MOVD  dst_base+0(FP), R0   // R0 = dst pointer
	MOVW  pixel+24(FP), R1     // R1 = 32-bit RGBA pixel
	MOVD  count+32(FP), R2     // R2 = pixel count

	CMP   $1, R2
	BLT   done

	// Broadcast pixel to all 4 lanes of V0 (4x uint32 = 16 bytes)
	VDUP  R1, V0.S4

	CMP   $4, R2
	BLT   tail

loop:
	VST1  [V0.S4], (R0)
	ADD   $16, R0
	SUB   $4, R2
	CMP   $4, R2
	BGE   loop

tail:
	CBZ   R2, done

tail_loop:
	MOVW  R1, (R0)
	ADD   $4, R0
	SUB   $1, R2
	CBNZ  R2, tail_loop

done:
	RET
