#include "textflag.h"

TEXT ·updateHashChain(SB), NOSPLIT, $0-56
    MOVQ memBase+0(FP), R15
    MOVL pos+8(FP), R14
    MOVL endPos+12(FP), R13
    MOVL windowOffset+16(FP), R12
    MOVL headOffset+20(FP), R11
    MOVL prevOffset+24(FP), R10
    MOVQ masks+32(FP), R9
    MOVL shift+40(FP), CX
    MOVL hash+44(FP), AX

    ADDQ R15, R12
    ADDQ R15, R11
    ADDQ R15, R10

    MOVQ R9, BX
    SHRQ $32, BX
    MOVL R9, R9

loop:
    CMPL R14, R13
    JGE done

    MOVBLZX (R12)(R14*1), DI

    SHLL CL, AX
    XORL DI, AX
    ANDL BX, AX

    MOVW (R11)(AX*2), SI

    MOVL R14, DI
    ANDL R9, DI
    MOVW SI, (R10)(DI*2)

    MOVW R14, (R11)(AX*2)

    INCL R14
    JMP loop

done:
    MOVL AX, ret+48(FP)
    RET
