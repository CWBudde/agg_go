//go:build amd64 && !purego

#include "textflag.h"

// func fillRGBAAVX2Asm(dst []byte, pixel uint32, count int)
TEXT ·fillRGBAAVX2Asm(SB), NOSPLIT, $0-40
	MOVQ dst_base+0(FP), DI
	MOVL pixel+24(FP), AX
	MOVQ count+32(FP), CX

	TESTQ CX, CX
	JLE done

	MOVD AX, X0
	VPBROADCASTD X0, Y0

	CMPQ CX, $8
	JB tail

loop:
	VMOVDQU Y0, (DI)
	ADDQ $32, DI
	SUBQ $8, CX
	CMPQ CX, $8
	JGE loop

tail:
	TESTQ CX, CX
	JLE done

tail_loop:
	MOVL AX, (DI)
	ADDQ $4, DI
	DECQ CX
	JNZ tail_loop

done:
	VZEROUPPER
	RET
