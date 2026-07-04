//go:build amd64 && !noasm

// Ultra-Safe & Fast Longest Match (Go ASM AMD64 ABI0 compliant)
// Prevents out-of-bounds panics by strictly limiting comparison to s.lookahead.
#include "textflag.h"

// func _longest_match_asm(m *Module, state_off int32, cur_match int32) int32
// Stack frame: 40 bytes for local variables, 20 bytes for args/return.
TEXT ·_longest_match_asm(SB), NOSPLIT, $40-32
    // Загружаем аргументы (выравнивание 8 байт)
    MOVQ m+0(FP), AX            // AX = Module pointer
    MOVL state_off+8(FP), CX    // CX = state offset (v0)
    MOVL cur_match+16(FP), R10  // R10 = cur_match index

    // Базовый указатель на Wasm память: m.memory.ptr
    MOVQ 48(AX), R8             // R8 = memory_ptr

    // s = memory_ptr + state_offset
    MOVL CX, CX
    LEAQ (R8)(CX*1), R9         // R9 = s (absolute state address)
    MOVQ R9, 0(SP)              // stack[0] = s_ptr (spill)

    // ВЫЧИСЛЯЕМ s.limit = strstart > w_size - 262 ? strstart - (w_size - 262) : 0
    MOVL 108(R9), AX            // AX = s.strstart
    MOVL 44(R9), BX             // BX = s.w_size
    SUBL $262, BX               // BX = w_size - 262
    CMPL AX, BX
    JLE limit_zero
    SUBL BX, AX                 // AX = limit
    JMP limit_ok
limit_zero:
    XORL AX, AX                 // limit = 0
limit_ok:
    MOVL AX, 16(SP)             // stack[16] = limit (отсечка окна)

    // Проверка: если cur_match <= limit, то выходим сразу
    CMPL R10, AX
    JLE mismatch_early

    // ВЫЧИСЛЯЕМ s.max_compare_len = min(nice_match, lookahead)
    MOVL 124(R9), AX            // AX = s.nice_match
    MOVL 116(R9), CX            // CX = s.lookahead
    CMPL AX, CX
    JLE use_nice_match
    MOVL CX, AX                 // Использовать lookahead, если он меньше
use_nice_match:
    // Ограничиваем сверху константой MAX_MATCH (258)
    CMPL AX, $258
    JLE max_len_ok
    MOVL $258, AX
max_len_ok:
    MOVL AX, 8(SP)              // stack[8] = max_compare_len (лимит сравнения)

    // s.prev_base -> на стек в stack[24]
    MOVL 64(R9), AX             // s.prev_offset
    MOVL AX, AX
    LEAQ (R8)(AX*1), AX
    MOVQ AX, 24(SP)             // stack[24] = prev_base

    // Настраиваем регистры для цикла
    MOVL 120(R9), R11           // R11 = best_len (рекорд)
    MOVL 128(R9), DX            // DX = max_chain (counter)

    // window_base в SI
    MOVL 56(R9), AX
    MOVL AX, AX
    LEAQ (R8)(AX*1), SI         // SI = window_base

    // scan_ptr в DI (window_base + strstart)
    MOVL 108(R9), AX
    MOVL AX, AX
    LEAQ (SI)(AX*1), DI         // DI = scan_ptr

chain_loop:
    // match_ptr в AX (window_base + cur_match)
    MOVL R10, R10
    LEAQ (SI)(R10*1), AX        // AX = match_ptr

    // Проверка window[cur_match + best_len] == window[strstart + best_len]
    MOVL R11, CX                // CX = best_len
    MOVB (DI)(CX*1), R8B        // R8B — временный байт
    MOVB (AX)(CX*1), R9B        // R9B — временный байт
    CMPB R8B, R9B
    JNE next_match

    // Проверка первых двух байт
    MOVW (DI), CX
    MOVW (AX), R8
    CMPW CX, R8
    JNE next_match

    // --- 64-BIT XOR SCAN (Ограниченный по lookahead) ---
    XORQ CX, CX                 // CX = счетчик длины
    
    // Вычисляем безопасный лимит для 8-байтового сравнения
    MOVL 8(SP), R8              // R8 = max_compare_len
    SUBL $8, R8                 // R8 = max_compare_len - 8
    JS compare_tail             // Если лимит < 8, идем сразу в побайтовый цикл
compare_8:
    CMPQ CX, R8
    JGE compare_tail            // Достигли безопасной границы - переходим на хвост
    MOVQ (DI)(CX*1), R9
    MOVQ (AX)(CX*1), R15
    XORQ R9, R15
    JNZ found_diff
    ADDQ $8, CX
    JMP compare_8

found_diff:
    BSFQ R15, R15
    SHRQ $3, R15
    ADDQ R15, CX
    JMP check_improvement

compare_tail:
    CMPL CX, 8(SP)              // Достигли ли s.max_compare_len?
    JGE check_improvement
    MOVB (DI)(CX*1), R8B
    MOVB (AX)(CX*1), R9B
    CMPB R8B, R9B
    JNE check_improvement
    INCQ CX
    JMP compare_tail

check_improvement:
    // Проверяем, лучше ли результат
    CMPL CX, R11
    JLE next_match

    // Новый рекорд!
    MOVL CX, R11                // best_len = CX
    MOVQ 0(SP), R9              // Восстанавливаем s_ptr
    MOVL R10, 112(R9)           // s.match_start = cur_match
    
    // nice_match check
    CMPL R11, 8(SP)             // nice_match (или lookahead)
    JGE done

next_match:
    // max_chain--
    DECL DX
    JZ done

    // cur_match = prev[cur_match & w_mask]
    MOVQ 0(SP), R9              // R9 = s_ptr
    MOVL 52(R9), AX             // AX = w_mask
    ANDL R10, AX
    MOVQ 24(SP), CX             // CX = prev_base
    MOVW (CX)(AX*2), R10        // Читаем uint16 из prev
    ANDL $0xFFFF, R10

    // cur_match > limit?
    CMPL R10, 16(SP)            // limit на стеке
    JG chain_loop

done:
    // Финализация
    MOVQ 0(SP), R9              // Восстанавливаем s_ptr
    MOVL R11, 120(R9)           // s.best_len = R11
    MOVL R11, AX                // !!! ВАЖНО: Возвращаем результат в AX для ABIInternal!
    MOVL R11, ret+16(FP)        // Дублируем на стек для ABI0 совместимости
    RET

mismatch_early:
    MOVQ 0(SP), R9
    MOVL 120(R9), AX            // Возвращаем текущий best_len
    MOVL AX, ret+16(FP)         // Возвращаем результат по смещению 16!
    RET
