; ═══════════════════════════════════════════════════════════════
; TEST FILE 04: DIRECTIVE VALIDATION & MEMORY ERRORS
; 40 distinct cases covering START address validation, WORD
; bounds, RESB/RESW negative/overflow, BYTE edge cases,
; and memory limit (32,768 byte) violations.
; ═══════════════════════════════════════════════════════════════
;
; --- CASE 01: Valid START ---
MEM    START  0000                         ;[OK] start at address 0
;
; --- CASES 02-07: WORD directive edge cases ---
W1     WORD   0                            ;[OK] zero is valid
W2     WORD   -1                           ;[OK] negative within range
W3     WORD   8388607                      ;[OK] max 24-bit positive
W4     WORD   -8388608                     ;[OK] min 24-bit negative
W5     WORD   8388608                      ;[E05] exceeds +24-bit max
W6     WORD   -8388609                     ;[E06] exceeds -24-bit min
;
; --- CASES 08-12: More WORD errors ---
W7     WORD   NOTANUM                      ;[E08] non-integer
W8     WORD   3.14159                      ;[E09] float not integer
W9     WORD   0xFF                         ;[E10] hex format not allowed for WORD
W10    WORD   1000000000                   ;[E11] way over 24-bit limit
W11    WORD                                ;[E12] missing operand entirely
;
; --- CASES 13-18: RESW edge cases ---
R1     RESW   1                            ;[OK] reserve 1 word (3 bytes)
R2     RESW   100                          ;[OK] reserve 100 words
R3     RESW   0                            ;[E15] zero not positive
R4     RESW   -1                           ;[E16] negative
R5     RESW   -100                         ;[E17] large negative
R6     RESW   11000                        ;[E18] 11000*3=33000 > 32768
;
; --- CASES 19-24: RESB edge cases ---
B1     RESB   1                            ;[OK] reserve 1 byte
B2     RESB   100                          ;[OK] reserve 100 bytes
B3     RESB   0                            ;[E21] zero not positive
B4     RESB   -1                           ;[E22] negative
B5     RESB   -500                         ;[E23] large negative
B6     RESB   40000                        ;[E24] exceeds 32KB
;
; --- CASES 25-30: BYTE character constant edge cases ---
C1     BYTE   C'A'                         ;[OK] single char
C2     BYTE   C'EOF'                       ;[OK] three chars
C3     BYTE   C'HELLO WORLD'               ;[OK] spaces in constant
C4     BYTE   C'ABCDEFGHIJKLMNOPQRST'      ;[OK] long string (20 bytes)
C5     BYTE   C''                          ;[E29] empty char constant
C6     BYTE   C'                           ;[E30] unclosed quote
;
; --- CASES 31-36: BYTE hex constant edge cases ---
X1     BYTE   X'00'                        ;[OK] zero byte
X2     BYTE   X'FF'                        ;[OK] max byte
X3     BYTE   X'AABB'                      ;[OK] two bytes
X4     BYTE   X'F'                         ;[E34] odd hex digits
X5     BYTE   X'FFF'                       ;[E35] odd hex digits (3)
X6     BYTE   X'GG'                        ;[E36] non-hex chars
;
; --- CASES 37-40: More BYTE format violations ---
X7     BYTE   X''                          ;[E37] empty hex constant
X8     BYTE   NOQUOTE                      ;[E38] no C/X prefix or quotes
X9     BYTE   'EOF'                        ;[E39] missing C/X prefix
X10    BYTE   D'123'                       ;[E40] invalid D prefix
;
; Note: START address overflow cases are in test_01
;
       END    MEM                          ;[OK]
