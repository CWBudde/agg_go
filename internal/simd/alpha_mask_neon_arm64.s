//go:build arm64 && !purego

#include "textflag.h"

DATA ·neonMaskR16+0(SB)/1, $0
DATA ·neonMaskR16+1(SB)/1, $3
DATA ·neonMaskR16+2(SB)/1, $6
DATA ·neonMaskR16+3(SB)/1, $9
DATA ·neonMaskR16+4(SB)/1, $0xff
DATA ·neonMaskR16+5(SB)/1, $0xff
DATA ·neonMaskR16+6(SB)/1, $0xff
DATA ·neonMaskR16+7(SB)/1, $0xff
DATA ·neonMaskR16+8(SB)/1, $0xff
DATA ·neonMaskR16+9(SB)/1, $0xff
DATA ·neonMaskR16+10(SB)/1, $0xff
DATA ·neonMaskR16+11(SB)/1, $0xff
DATA ·neonMaskR16+12(SB)/1, $0xff
DATA ·neonMaskR16+13(SB)/1, $0xff
DATA ·neonMaskR16+14(SB)/1, $0xff
DATA ·neonMaskR16+15(SB)/1, $0xff
GLOBL ·neonMaskR16(SB), RODATA|NOPTR, $16

DATA ·neonMaskG16+0(SB)/1, $1
DATA ·neonMaskG16+1(SB)/1, $4
DATA ·neonMaskG16+2(SB)/1, $7
DATA ·neonMaskG16+3(SB)/1, $10
DATA ·neonMaskG16+4(SB)/1, $0xff
DATA ·neonMaskG16+5(SB)/1, $0xff
DATA ·neonMaskG16+6(SB)/1, $0xff
DATA ·neonMaskG16+7(SB)/1, $0xff
DATA ·neonMaskG16+8(SB)/1, $0xff
DATA ·neonMaskG16+9(SB)/1, $0xff
DATA ·neonMaskG16+10(SB)/1, $0xff
DATA ·neonMaskG16+11(SB)/1, $0xff
DATA ·neonMaskG16+12(SB)/1, $0xff
DATA ·neonMaskG16+13(SB)/1, $0xff
DATA ·neonMaskG16+14(SB)/1, $0xff
DATA ·neonMaskG16+15(SB)/1, $0xff
GLOBL ·neonMaskG16(SB), RODATA|NOPTR, $16

DATA ·neonMaskB16+0(SB)/1, $2
DATA ·neonMaskB16+1(SB)/1, $5
DATA ·neonMaskB16+2(SB)/1, $8
DATA ·neonMaskB16+3(SB)/1, $11
DATA ·neonMaskB16+4(SB)/1, $0xff
DATA ·neonMaskB16+5(SB)/1, $0xff
DATA ·neonMaskB16+6(SB)/1, $0xff
DATA ·neonMaskB16+7(SB)/1, $0xff
DATA ·neonMaskB16+8(SB)/1, $0xff
DATA ·neonMaskB16+9(SB)/1, $0xff
DATA ·neonMaskB16+10(SB)/1, $0xff
DATA ·neonMaskB16+11(SB)/1, $0xff
DATA ·neonMaskB16+12(SB)/1, $0xff
DATA ·neonMaskB16+13(SB)/1, $0xff
DATA ·neonMaskB16+14(SB)/1, $0xff
DATA ·neonMaskB16+15(SB)/1, $0xff
GLOBL ·neonMaskB16(SB), RODATA|NOPTR, $16

// func copyMask1UNEONAsm(dst, src []byte, count int)
TEXT ·copyMask1UNEONAsm(SB), NOSPLIT, $0-56
	MOVD dst_base+0(FP), R0
	MOVD src_base+24(FP), R1
	MOVD count+48(FP), R2

	CMP  $1, R2
	BLT  neon_copy_done

	CMP  $16, R2
	BLT  neon_copy_tail

neon_copy_loop:
	VLD1 (R1), [V0.B16]
	VST1 [V0.B16], (R0)
	ADD  $16, R1
	ADD  $16, R0
	SUB  $16, R2
	CMP  $16, R2
	BGE  neon_copy_loop

neon_copy_tail:
	CBZ  R2, neon_copy_done

neon_copy_tail_loop:
	MOVBU (R1), R3
	MOVB  R3, (R0)
	ADD   $1, R1
	ADD   $1, R0
	SUB   $1, R2
	CBNZ  R2, neon_copy_tail_loop

neon_copy_done:
	RET

// func rgb24ToGrayU8NEONAsm(dst, src []byte, blocks int)
// Processes 8 pixels per block from two overlapping 16-byte loads:
// bytes [0:16] produce pixels 0..3, bytes [12:28] produce pixels 4..7.
TEXT ·rgb24ToGrayU8NEONAsm(SB), NOSPLIT, $0-56
	MOVD dst_base+0(FP), R0
	MOVD src_base+24(FP), R1
	MOVD blocks+48(FP), R2

	CMP  $1, R2
	BLT  neon_gray_done

	MOVD  $77, R20
	MOVD  $150, R21
	MOVD  $29, R22
	VDUP  R20, V20.B8
	VDUP  R21, V21.B8
	VDUP  R22, V22.B8

	MOVD  $·neonMaskR16(SB), R10
	MOVD  $·neonMaskG16(SB), R11
	MOVD  $·neonMaskB16(SB), R12
	VLD1  (R10), [V16.B16]
	VLD1  (R11), [V17.B16]
	VLD1  (R12), [V18.B16]

neon_gray_loop:
	VLD1  (R1), [V0.B16]
	ADD   $12, R1, R9
	VLD1  (R9), [V1.B16]

	// pixels 0..3 from V0
	VTBL  V16.B16, [V0.B16], V2.B16
	VTBL  V17.B16, [V0.B16], V3.B16
	VTBL  V18.B16, [V0.B16], V4.B16
	VPMULL V20.B8, V2.B8, V5.H8
	VPMULL V21.B8, V3.B8, V6.H8
	VPMULL V22.B8, V4.B8, V7.H8
	VADD  V6.H8, V5.H8, V5.H8
	VADD  V7.H8, V5.H8, V5.H8
	VUSHR $8, V5.H8, V5.H8
	VUZP1 V5.B16, V5.B16, V8.B16
	VMOV  V8.S[0], R3
	MOVW  R3, (R0)

	// pixels 4..7 from V1
	VTBL  V16.B16, [V1.B16], V9.B16
	VTBL  V17.B16, [V1.B16], V10.B16
	VTBL  V18.B16, [V1.B16], V11.B16
	VPMULL V20.B8, V9.B8, V12.H8
	VPMULL V21.B8, V10.B8, V13.H8
	VPMULL V22.B8, V11.B8, V14.H8
	VADD  V13.H8, V12.H8, V12.H8
	VADD  V14.H8, V12.H8, V12.H8
	VUSHR $8, V12.H8, V12.H8
	VUZP1 V12.B16, V12.B16, V15.B16
	VMOV  V15.S[0], R4
	MOVW  R4, 4(R0)

	ADD  $24, R1
	ADD  $8, R0
	SUB  $1, R2
	CBNZ R2, neon_gray_loop

neon_gray_done:
	RET
