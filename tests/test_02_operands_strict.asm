; ═══════════════════════════════════════════════════════════════
; TEST FILE 02: OPERAND COUNT & TYPE VALIDATION
; 40 distinct cases for RSUB with operands, missing operands,
; too many operands, implicit accumulator violations, and
; malformed indexed addressing.
; ═══════════════════════════════════════════════════════════════
;
OPER   START  2000                         ;[OK] valid start
ALPHA  WORD   5                            ;[OK] define ALPHA
BETA   WORD   10                           ;[OK] define BETA
GAMMA  WORD   15                           ;[OK] define GAMMA
;
; --- CASES 01-06: RSUB with operands (zero-operand violation) ---
       RSUB   ALPHA                        ;[E01] RSUB takes 0 operands
       RSUB   BETA                         ;[E02] RSUB takes 0 operands
       RSUB   0000                         ;[E03] RSUB takes 0 operands
RET1   RSUB   GAMMA                        ;[E04] RSUB takes 0 operands (with label)
       RSUB   1000                         ;[E05] RSUB with absolute addr
       RSUB   ALPHA,X                      ;[E06] RSUB with indexed
;
; --- CASES 07-14: Missing operands on one-operand instructions ---
       LDA                                 ;[E07] LDA needs operand
       STA                                 ;[E08] STA needs operand
       ADD                                 ;[E09] ADD needs operand
       SUB                                 ;[E10] SUB needs operand
       COMP                                ;[E11] COMP needs operand
       LDX                                 ;[E12] LDX needs operand
       LDCH                                ;[E13] LDCH needs operand
       STCH                                ;[E14] STCH needs operand
;
; --- CASES 15-20: More missing operands ---
       MUL                                 ;[E15] MUL needs operand
       DIV                                 ;[E16] DIV needs operand
       AND                                 ;[E17] AND needs operand
       OR                                  ;[E18] OR needs operand
       TIX                                 ;[E19] TIX needs operand
       J                                   ;[E20] J needs operand
;
; --- CASES 21-24: Missing operands on jump instructions ---
       JEQ                                 ;[E21] JEQ needs operand
       JGT                                 ;[E22] JGT needs operand
       JLT                                 ;[E23] JLT needs operand
       JSUB                                ;[E24] JSUB needs operand
;
; --- CASES 25-30: Explicit accumulator (too many operands) ---
       ADD    A,ALPHA                      ;[E25] A is implicit, use ADD ALPHA
       SUB    A,BETA                       ;[E26] accumulator implicit
       COMP   A,GAMMA                      ;[E27] accumulator implicit
       LDA    A,ALPHA                      ;[E28] accumulator implicit
       AND    A,BETA                       ;[E29] accumulator implicit
       OR     A,GAMMA                      ;[E30] accumulator implicit
;
; --- CASES 31-38: Malformed indexed addressing ---
       LDCH   ALPHA,Y                      ;[E31] Y register invalid, only X
       STCH   ALPHA,A                      ;[E32] A register not for indexing
       ADD    ALPHA,B                      ;[E33] B register invalid
       LDA    ,X                           ;[E34] missing base operand with ,X
       SUB    ALPHA,Z                      ;[E35] Z register invalid
       COMP   ALPHA,L                      ;[E36] L register not for indexing
       MUL    ALPHA,S                      ;[E37] S register invalid
       DIV    ALPHA,T                      ;[E38] T register invalid
;
; --- CASES 39-42: I/O instruction operand issues ---
       RD                                  ;[E39] RD needs device operand
       WD                                  ;[E40] WD needs device operand
       TD                                  ;[E41] TD needs device operand
       STSW                                ;[E42] STSW needs operand
;
       END    OPER                         ;[OK]
