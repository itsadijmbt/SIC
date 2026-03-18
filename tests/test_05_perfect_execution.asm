; ═══════════════════════════════════════════════════════════════
; TEST FILE 05: PERFECT EXECUTION — 100% VALID SIC PROGRAM
; Uses ALL 25 standard SIC opcodes, ALL 6 directives,
; indexed addressing, forward references, and comments.
; This must assemble with zero errors and produce correct
; H, T, E object records.
; ═══════════════════════════════════════════════════════════════
;
; PROGRAM: Complete file copy with I/O, arithmetic, logic, jumps.
;
; --- DIRECTIVES: START ---
ALLSIC START  1000                         ; Program loads at 0x1000
;
; ══════════════════════════════════════════
; SECTION 1: Load/Store instructions
; Tests: LDA, STA, LDX, STX, LDL, STL, LDCH, STCH
; ══════════════════════════════════════════
FIRST  STL    RETADR                       ; V01: STL - store link register
       LDA    FIVE                         ; V02: LDA - load accumulator
       STA    ALPHA                        ; V03: STA - store accumulator
       LDX    ZERO                         ; V04: LDX - load index register
       STX    INDEX                        ; V05: STX - store index register
       LDL    RETADR                       ; V06: LDL - load link register
       LDCH   CHRONE                       ; V07: LDCH - load character
       STCH   BUFFER,X                     ; V08: STCH - store char indexed
;
; ══════════════════════════════════════════
; SECTION 2: Arithmetic instructions
; Tests: ADD, SUB, MUL, DIV, COMP
; ══════════════════════════════════════════
       LDA    FIVE                         ; V09: load 5
       ADD    TEN                          ; V10: ADD - A = A + TEN
       SUB    THREE                        ; V11: SUB - A = A - THREE
       MUL    TWO                          ; V12: MUL - A = A * TWO
       DIV    FIVE                         ; V13: DIV - A = A / FIVE
       COMP   ZERO                         ; V14: COMP - compare A with 0
;
; ══════════════════════════════════════════
; SECTION 3: Logical instructions
; Tests: AND, OR
; ══════════════════════════════════════════
       LDA    MASK                         ; V15: load mask
       AND    ALPHA                        ; V16: AND - bitwise AND
       OR     BETA                         ; V17: OR - bitwise OR
       STA    RESULT                       ; V18: store result
;
; ══════════════════════════════════════════
; SECTION 4: Jump instructions
; Tests: J, JEQ, JGT, JLT, JSUB, RSUB
; ══════════════════════════════════════════
       COMP   ZERO                         ; V19: set condition code
       JEQ    EQPATH                       ; V20: JEQ - jump if equal
       JGT    GTPATH                       ; V21: JGT - jump if greater
       JLT    LTPATH                       ; V22: JLT - jump if less
       J      MERGE                        ; V23: J - unconditional jump
;
EQPATH LDA    ONE                          ; V24: equal path
       J      MERGE                        ; V25: jump to merge
;
GTPATH LDA    TWO                          ; V26: greater path
       J      MERGE                        ; V27: jump to merge
;
LTPATH LDA    THREE                        ; V28: less-than path
       J      MERGE                        ; V29: jump to merge
;
MERGE  STA    RESULT                       ; V30: merge point
       JSUB   PROC1                        ; V31: JSUB - jump to subroutine
       J      IOMAIN                       ; V32: continue to I/O section
;
; ══════════════════════════════════════════
; SECTION 5: Subroutine (tests RSUB)
; ══════════════════════════════════════════
PROC1  LDA    FIVE                         ; V33: subroutine body
       ADD    TEN                          ; V34: do some work
       STA    RESULT                       ; V35: store
       RSUB                                ; V36: RSUB - return from sub
;
; ══════════════════════════════════════════
; SECTION 6: I/O instructions
; Tests: TD, RD, WD
; ══════════════════════════════════════════
IOMAIN TD     INDEV                        ; V37: TD - test device
       JEQ    IOMAIN                       ; V38: loop if not ready
       RD     INDEV                        ; V39: RD - read from device
       STCH   BUFFER,X                     ; V40: store to buffer indexed
       TD     OUTDEV                       ; V41: test output device
       JEQ    WRWAIT                       ; V42: wait loop
WRWAIT LDCH   BUFFER,X                     ; V43: load from buffer indexed
       WD     OUTDEV                       ; V44: WD - write to device
;
; ══════════════════════════════════════════
; SECTION 7: Index and loop
; Tests: TIX with indexed addressing
; ══════════════════════════════════════════
       LDX    ZERO                         ; V45: reset index
LOOP   LDCH   BUFFER,X                     ; V46: load byte indexed
       STCH   OUBUF,X                      ; V47: store byte indexed
       TIX    MAXLEN                       ; V48: TIX - increment X, compare
       JLT    LOOP                         ; V49: loop if X < MAXLEN
;
; ══════════════════════════════════════════
; SECTION 8: Status register
; Tests: STSW
; ══════════════════════════════════════════
       STSW   STATUS                       ; V50: STSW - store status word
;
; ══════════════════════════════════════════
; SECTION 9: Finish
; ══════════════════════════════════════════
FINISH LDL    RETADR                       ; V51: restore link register
       RSUB                                ; V52: return
;
; ══════════════════════════════════════════
; DATA SECTION — Tests all data directives
; WORD, BYTE (C and X), RESW, RESB
; ══════════════════════════════════════════
ZERO   WORD   0                            ; V53: WORD constant
ONE    WORD   1                            ; V54: WORD constant
TWO    WORD   2                            ; V55: WORD constant
THREE  WORD   3                            ; V56: WORD constant
FIVE   WORD   5                            ; V57: WORD constant
TEN    WORD   10                           ; V58: WORD constant
MASK   WORD   255                          ; V59: WORD constant (0xFF)
ALPHA  RESW   1                            ; V60: RESW - 1 word
BETA   RESW   1                            ; V61: RESW - 1 word
RESULT RESW   1                            ; V62: RESW - 1 word
INDEX  RESW   1                            ; V63: RESW - 1 word
STATUS RESW   1                            ; V64: RESW - 1 word
RETADR RESW   1                            ; V65: RESW - 1 word
LENGTH RESW   1                            ; V66: RESW - 1 word
MAXLEN WORD   4096                         ; V67: WORD constant
CHRONE BYTE   C'Z'                         ; V68: BYTE character single
INDEV  BYTE   X'F1'                        ; V69: BYTE hex constant
OUTDEV BYTE   X'05'                        ; V70: BYTE hex constant
EOF    BYTE   C'EOF'                       ; V71: BYTE character multi
BUFFER RESB   4096                         ; V72: RESB - large buffer
OUBUF  RESB   4096                         ; V73: RESB - output buffer
;
; --- DIRECTIVE: END ---
       END    FIRST                        ; V74: END with first executable
