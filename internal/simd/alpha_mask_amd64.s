//go:build amd64 && !purego

#include "textflag.h"

DATA ·maskExtractRG16(SB)/8, $0x0A09070604030100
DATA ·maskExtractRG16+8(SB)/8, $0x8080808080808080
GLOBL ·maskExtractRG16(SB), RODATA|NOPTR, $16

DATA ·maskExtractGB16(SB)/8, $0x0B0A080705040201
DATA ·maskExtractGB16+8(SB)/8, $0x8080808080808080
GLOBL ·maskExtractGB16(SB), RODATA|NOPTR, $16

DATA ·weightRGBase16(SB)/8, $0x334D334D334D334D
DATA ·weightRGBase16+8(SB)/8, $0x334D334D334D334D
GLOBL ·weightRGBase16(SB), RODATA|NOPTR, $16

DATA ·weightRGExtra16(SB)/8, $0x6300630063006300
DATA ·weightRGExtra16+8(SB)/8, $0x6300630063006300
GLOBL ·weightRGExtra16(SB), RODATA|NOPTR, $16

DATA ·weightGBBlue16(SB)/8, $0x1D001D001D001D00
DATA ·weightGBBlue16+8(SB)/8, $0x1D001D001D001D00
GLOBL ·weightGBBlue16(SB), RODATA|NOPTR, $16

// func copyMask1U8SSE41Asm(dst, src []byte, count int)
TEXT ·copyMask1U8SSE41Asm(SB), NOSPLIT, $0-56
	MOVQ dst_base+0(FP), DI
	MOVQ src_base+24(FP), SI
	MOVQ count+48(FP), CX

	TESTQ CX, CX
	JLE   copy_done

	CMPQ CX, $16
	JB   copy_tail

copy_loop:
	MOVOU (SI), X0
	MOVOU X0, (DI)
	ADDQ  $16, SI
	ADDQ  $16, DI
	SUBQ  $16, CX
	CMPQ  CX, $16
	JGE   copy_loop

copy_tail:
	TESTQ CX, CX
	JLE   copy_done

copy_tail_loop:
	MOVBLZX (SI), AX
	MOVB    AL, (DI)
	INCQ    SI
	INCQ    DI
	DECQ    CX
	JNZ     copy_tail_loop

copy_done:
	RET

// func rgb24ToGrayU8SSE41Asm(dst, src []byte, blocks int)
// Processes 4 RGB24 pixels per block. The caller guarantees each block can
// safely read 16 bytes from src.
TEXT ·rgb24ToGrayU8SSE41Asm(SB), NOSPLIT, $0-56
	MOVQ dst_base+0(FP), DI
	MOVQ src_base+24(FP), SI
	MOVQ blocks+48(FP), CX

	TESTQ CX, CX
	JLE   gray_done

	MOVOU ·maskExtractRG16(SB), X10
	MOVOU ·maskExtractGB16(SB), X11
	MOVOU ·weightRGBase16(SB), X12
	MOVOU ·weightRGExtra16(SB), X13
	MOVOU ·weightGBBlue16(SB), X14

gray_loop:
	MOVOU  (SI), X0
	MOVOU  X0, X1
	MOVOU  X0, X2
	MOVOU  X0, X3
	PSHUFB X10, X1
	PSHUFB X10, X2
	PSHUFB X11, X3
	PMADDUBSW X12, X1
	PMADDUBSW X13, X2
	PMADDUBSW X14, X3
	PADDW  X2, X1
	PADDW  X3, X1
	PSRLW  $8, X1
	PACKUSWB X1, X1
	MOVD   X1, (DI)

	ADDQ $12, SI
	ADDQ $4, DI
	DECQ CX
	JNZ  gray_loop

gray_done:
	RET
