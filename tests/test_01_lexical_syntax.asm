; ═══════════════════════════════════════════════════════════════
; TEST FILE 01: LEXICAL & SYNTAX ERRORS
; 42 distinct error cases covering illegal characters, malformed
; labels, invalid hex, unclosed quotes, bad mnemonics, spacing.
; Lines marked ;[OK] are valid context lines. All others are errors.
; ═══════════════════════════════════════════════════════════════
;
; --- CASE 01: Valid START (context) ---
COPY   START  1000                         ;[OK] valid start
;
; --- CASES 02-06: Bad label names ---
1STLBL LDA    ALPHA                        ;[E02] label starts with digit
A@B    STA    BETA                         ;[E03] label contains @
LAB#EL LDA    GAMMA                        ;[E04] label contains #
L.ABEL LDA    DELTA                        ;[E05] label contains .
TOOLONGX LDA  SIGMA                        ;[E06] label exceeds 6 chars
;
; --- CASES 07-10: Completely unrecognized mnemonics ---
VALID1 MOVX   ALPHA                        ;[E07] MOVX is not a SIC opcode
       HALT                                ;[E08] HALT is not a SIC opcode
       PUSH   ALPHA                        ;[E09] PUSH is not SIC
       POP    ALPHA                        ;[E10] POP is not SIC
;
; --- CASES 11-14: SIC/XE instructions used on SIC (invalid) ---
       LDB    ALPHA                        ;[E11] LDB is SIC/XE only
       STB    ALPHA                        ;[E12] STB is SIC/XE only
       TIXR   ALPHA                        ;[E13] TIXR is SIC/XE
       CLEAR  ALPHA                        ;[E14] CLEAR is SIC/XE
;
; --- CASES 15-20: Malformed BYTE directives ---
BAD1   BYTE   C'                           ;[E15] unclosed C quote
BAD2   BYTE   C''                          ;[E16] empty C constant
BAD3   BYTE   X'GG'                        ;[E17] non-hex chars in X
BAD4   BYTE   X'F'                         ;[E18] odd number of hex digits
BAD5   BYTE   X''                          ;[E19] empty X constant
BAD6   BYTE   HELLO                        ;[E20] missing C or X prefix
;
; --- CASES 21-27: More BYTE format violations ---
BAD7   BYTE   C                            ;[E21] C without quotes
BAD8   BYTE   X                            ;[E22] X without quotes
BAD9   BYTE   Z'FF'                        ;[E23] invalid prefix Z
BAD10  BYTE   D'99'                        ;[E24] invalid prefix D
BAD1A  BYTE   'EOF'                        ;[E25] missing C/X prefix
BAD1B  BYTE   12345                        ;[E26] bare number, no C/X format
BAD1C  BYTE   X'GGFF'                      ;[E27] G is not valid hex
;
; --- CASES 28-31: Missing operands on directives ---
       START                               ;[E28] START with no address
BAD11  WORD                                ;[E29] WORD with no value
BAD12  RESW                                ;[E30] RESW with no count
BAD13  RESB                                ;[E31] RESB with no count
;
; --- CASES 29-32: Wrong operand types on directives ---
BAD14  WORD   HELLO                        ;[E29] WORD expects integer not string
BAD15  WORD   3.14                         ;[E30] WORD expects integer not float
BAD16  RESW   ABC                          ;[E31] RESW expects integer
BAD17  RESB   XYZ                          ;[E32] RESB expects integer
;
; --- CASES 33-36: Hex address issues for START ---
PROG2  START  ZZZZ                         ;[E33] non-hex address for START
PROG3  START  12GH                         ;[E34] contains non-hex chars
PROG4  START  -100                         ;[E35] negative not valid hex
PROG5  START  FFFFF                        ;[E36] exceeds 32KB limit
;
; --- CASES 37-40: Misc lexical issues ---
       LDA                                 ;[E37] missing operand
       STA                                 ;[E38] missing operand
       ADD                                 ;[E39] missing operand
       COMP                                ;[E40] missing operand
;
; --- CASES 41-42: END directive issues ---
       END                                 ;[E41] END with no operand
       END    COPY                         ;[E42] valid END (context cleanup)
