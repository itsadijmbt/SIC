; ===== SIC Error Test File =====
; This file contains intentional errors to test the assembler's error handling.

COPY   START  GHIJ
FIRST  STL    RETADR
FIRST  LDA    ALPHA
       MOVX   BETA
       ADD
       RSUB   ALPHA
       LDA    A, ALPHA
       BYTE   C''
       BYTE   X'F'
       BYTE   HELLO
       WORD   ABC
       RESW   -5
       RESB   99999
TOOLNG LDA    BUFFER
       ADD    UNDEF
123BAD LDA    ALPHA
       END    FIRST
