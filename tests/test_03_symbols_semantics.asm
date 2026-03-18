; ═══════════════════════════════════════════════════════════════
; TEST FILE 03: SYMBOLS & SEMANTIC ERRORS
; 35 distinct cases covering duplicate labels, undefined symbols,
; forward reference issues, END directive semantics.
; ═══════════════════════════════════════════════════════════════
;
SYM    START  1000                         ;[OK] valid start
;
; --- CASES 01-02: Valid symbol definitions (context) ---
FIRST  LDA    ALPHA                        ;[OK] defines FIRST, forward ref to ALPHA
       STA    BETA                         ;[OK] forward ref to BETA
;
; --- CASES 03-07: Duplicate label definitions ---
LOOP   LDA    ALPHA                        ;[OK] first definition of LOOP
LOOP   STA    BETA                         ;[E03] duplicate symbol LOOP
FIRST  ADD    ALPHA                        ;[E04] duplicate symbol FIRST (already line 8)
DATA   WORD   5                            ;[OK] first definition of DATA
DATA   WORD   10                           ;[E05] duplicate symbol DATA
EXIT   RSUB                                ;[OK] first definition of EXIT
EXIT   J      LOOP                         ;[E07] duplicate symbol EXIT
;
; --- CASES 08-15: Undefined symbols in operands ---
       LDA    XYZZY                        ;[E08] XYZZY never defined
       STA    NOTHERE                      ;[E09] NOTHERE never defined
       ADD    UNDEF1                       ;[E10] UNDEF1 never defined
       SUB    UNDEF2                       ;[E11] UNDEF2 never defined
       COMP   UNDEF3                       ;[E12] UNDEF3 never defined
       J      NOLBL                        ;[E13] NOLBL never defined
       JEQ    BADREF                       ;[E14] BADREF never defined
       JSUB   NOSUB                        ;[E15] NOSUB never defined
;
; --- CASES 16-20: Forward refs that ARE eventually defined (valid) ---
       LDA    FWDA                         ;[OK] forward ref, defined below
       STA    FWDB                         ;[OK] forward ref, defined below
       ADD    FWDC                         ;[OK] forward ref, defined below
       J      FWDD                         ;[OK] forward ref, defined below
       JSUB   FWDE                         ;[OK] forward ref, defined below
;
; --- Define the forward reference targets ---
FWDA   WORD   100                          ;[OK]
FWDB   WORD   200                          ;[OK]
FWDC   WORD   300                          ;[OK]
FWDD   LDA    ALPHA                        ;[OK]
FWDE   RSUB                               ;[OK]
;
; --- CASES 21-24: Indexed addressing with undefined symbols ---
       LDA    UNDFX,X                      ;[E21] UNDFX not defined
       STCH   UNDFBF,X                     ;[E22] UNDFBF not defined
       LDCH   NOBUF,X                      ;[E23] NOBUF not defined
       ADD    NADA,X                       ;[E24] NADA not defined
;
; --- CASES 25-28: Case sensitivity tests ---
ALPHA  WORD   42                           ;[OK] define ALPHA
alpha  WORD   43                           ;[OK or E25] lowercase label
       LDA    Alpha                        ;[E26] case mismatch - undefined
       STA    ALPHA                        ;[OK] exact match
       STA    aLPHA                        ;[E28] case mismatch - undefined
;
; --- CASES 29-30: Symbols with max length boundary ---
ABCDEF WORD   1                            ;[OK] exactly 6 chars
       LDA    ABCDEF                       ;[OK] references 6-char label
;
; --- CASES 31-35: Various END directive issues ---
BETA   WORD   99                           ;[OK] define BETA
; Note: only one END per file, so these cases are documented
; but only the last END will execute. Previous lines test
; the symbol lookup behavior.
       LDA    BETA                         ;[OK]
       LDA    GHOST                        ;[E33] undefined in pass 2
       J      SPOOKY                       ;[E34] undefined in pass 2
       RSUB                                ;[OK]
       END    SYM                          ;[OK] valid end
