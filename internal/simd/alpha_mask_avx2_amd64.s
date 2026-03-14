//go:build amd64 && !purego

#include "textflag.h"

DATA ·maskExtractRG32+0(SB)/8, $0x0A09070604030100
DATA ·maskExtractRG32+8(SB)/8, $0x8080808080808080
DATA ·maskExtractRG32+16(SB)/8, $0x0A09070604030100
DATA ·maskExtractRG32+24(SB)/8, $0x8080808080808080
GLOBL ·maskExtractRG32(SB), RODATA|NOPTR, $32

DATA ·maskExtractGB32+0(SB)/8, $0x0B0A080705040201
DATA ·maskExtractGB32+8(SB)/8, $0x8080808080808080
DATA ·maskExtractGB32+16(SB)/8, $0x0B0A080705040201
DATA ·maskExtractGB32+24(SB)/8, $0x8080808080808080
GLOBL ·maskExtractGB32(SB), RODATA|NOPTR, $32

DATA ·weightRGBase32+0(SB)/8, $0x334D334D334D334D
DATA ·weightRGBase32+8(SB)/8, $0x334D334D334D334D
DATA ·weightRGBase32+16(SB)/8, $0x334D334D334D334D
DATA ·weightRGBase32+24(SB)/8, $0x334D334D334D334D
GLOBL ·weightRGBase32(SB), RODATA|NOPTR, $32

DATA ·weightRGExtra32+0(SB)/8, $0x6300630063006300
DATA ·weightRGExtra32+8(SB)/8, $0x6300630063006300
DATA ·weightRGExtra32+16(SB)/8, $0x6300630063006300
DATA ·weightRGExtra32+24(SB)/8, $0x6300630063006300
GLOBL ·weightRGExtra32(SB), RODATA|NOPTR, $32

DATA ·weightGBBlue32+0(SB)/8, $0x1D001D001D001D00
DATA ·weightGBBlue32+8(SB)/8, $0x1D001D001D001D00
DATA ·weightGBBlue32+16(SB)/8, $0x1D001D001D001D00
DATA ·weightGBBlue32+24(SB)/8, $0x1D001D001D001D00
GLOBL ·weightGBBlue32(SB), RODATA|NOPTR, $32

DATA ·grayPack8Mask16+0(SB)/8, $0x0B0A090803020100
DATA ·grayPack8Mask16+8(SB)/8, $0x8080808080808080
GLOBL ·grayPack8Mask16(SB), RODATA|NOPTR, $16

// func copyMask1U8AVX2Asm(dst, src []byte, count int)
TEXT ·copyMask1U8AVX2Asm(SB), NOSPLIT, $0-56
	MOVQ dst_base+0(FP), DI
	MOVQ src_base+24(FP), SI
	MOVQ count+48(FP), CX

	TESTQ CX, CX
	JLE   avx_copy_done

	CMPQ CX, $32
	JB   avx_copy_tail16

avx_copy_loop32:
	VMOVDQU (SI), Y0
	VMOVDQU Y0, (DI)
	ADDQ   $32, SI
	ADDQ   $32, DI
	SUBQ   $32, CX
	CMPQ   CX, $32
	JGE    avx_copy_loop32

avx_copy_tail16:
	CMPQ CX, $16
	JB   avx_copy_tail

	VMOVDQU (SI), X0
	VMOVDQU X0, (DI)
	ADDQ   $16, SI
	ADDQ   $16, DI
	SUBQ   $16, CX

avx_copy_tail:
	TESTQ CX, CX
	JLE   avx_copy_done

avx_copy_tail_loop:
	MOVBLZX (SI), AX
	MOVB    AL, (DI)
	INCQ    SI
	INCQ    DI
	DECQ    CX
	JNZ     avx_copy_tail_loop

avx_copy_done:
	VZEROUPPER
	RET

// func rgb24ToGrayU8AVX2Asm(dst, src []byte, blocks int)
// Processes 8 RGB24 pixels per block by loading two overlapping 16-byte chunks:
// bytes [0:16] become lane 0 (pixels 0..3), bytes [12:28] become lane 1 (pixels 4..7).
TEXT ·rgb24ToGrayU8AVX2Asm(SB), NOSPLIT, $0-56
	MOVQ dst_base+0(FP), DI
	MOVQ src_base+24(FP), SI
	MOVQ blocks+48(FP), CX

	TESTQ CX, CX
	JLE   avx_gray_done

	VMOVDQU ·maskExtractRG32(SB), Y10
	VMOVDQU ·maskExtractGB32(SB), Y11
	VMOVDQU ·weightRGBase32(SB), Y12
	VMOVDQU ·weightRGExtra32(SB), Y13
	VMOVDQU ·weightGBBlue32(SB), Y14

avx_gray_loop:
	VMOVDQU      (SI), X0
	VMOVDQU      12(SI), X1
	VINSERTI128  $1, X1, Y0, Y0
	VMOVDQU      Y0, Y1
	VMOVDQU      Y0, Y2
	VMOVDQU      Y0, Y3
	VPSHUFB      Y10, Y1, Y1
	VPSHUFB      Y10, Y2, Y2
	VPSHUFB      Y11, Y3, Y3
	VPMADDUBSW   Y12, Y1, Y1
	VPMADDUBSW   Y13, Y2, Y2
	VPMADDUBSW   Y14, Y3, Y3
	VPADDW       Y2, Y1, Y1
	VPADDW       Y3, Y1, Y1
	VPSRLW       $8, Y1, Y1
	VEXTRACTI128 $1, Y1, X2
	VPACKUSWB    X2, X1, X1
	VPSHUFB      ·grayPack8Mask16(SB), X1, X1
	MOVQ         X1, (DI)

	ADDQ $24, SI
	ADDQ $8, DI
	DECQ CX
	JNZ  avx_gray_loop

avx_gray_done:
	VZEROUPPER
	RET
