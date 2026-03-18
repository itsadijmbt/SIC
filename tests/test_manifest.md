# SIC Assembler Test Suite — Expected Results Manifest

**Total Distinct Test Cases: 195+**
- test_01_lexical_syntax.asm: **43 errors** (lexical, syntax, memory)
- test_02_operands_strict.asm: **42 errors** (operand count/type)
- test_03_symbols_semantics.asm: **~20 errors** (4 Pass 1 duplicates + ~16 Pass 2 undefined symbols)
- test_04_directives_memory.asm: **24 errors** (directive bounds, memory)
- test_05_perfect_execution.asm: **0 errors** (74 valid instruction/directive cases)

---

## FILE 1: test_01_lexical_syntax.asm — 43 Errors

### Summary
Tests bad label names, unrecognized mnemonics, SIC/XE-only instructions, all BYTE
format violations, missing/wrong directive operands, invalid START addresses, missing
instruction operands, and END without operand.

| Category | Lines | Count | Error Type |
|----------|-------|-------|------------|
| Bad labels (digit start, illegal chars, >6 chars) | 12-16 | 5 | Lexical |
| Unrecognized mnemonics (MOVX, HALT, PUSH, POP) | 19-22 | 4 | Syntax |
| SIC/XE-only opcodes (LDB, STB, TIXR, CLEAR) | 25-28 | 4 | Syntax |
| Malformed BYTE (unclosed quote, empty, bad hex) | 31-36 | 6 | Syntax |
| More BYTE violations (no quotes, bad prefix, bare int) | 39-45 | 7 | Syntax |
| Missing directive operands (START/WORD/RESW/RESB) | 48-51 | 4 | Syntax |
| Wrong directive operand types (string/float for WORD/RESW/RESB) | 54-57 | 4 | Lexical |
| Bad START hex addresses (non-hex, negative, >32KB) | 60-63 | 4 | Lexical/Memory |
| Missing instruction operands (LDA/STA/ADD/COMP) | 66-69 | 4 | Syntax |
| END with no operand | 72 | 1 | Syntax |
| **Total** | | **43** | |

### Key Cases

| Line | Statement | Why It Fails |
|------|-----------|-------------|
| 12 | `1STLBL LDA ALPHA` | Label starts with digit. SIC labels must begin with letter. |
| 16 | `TOOLONGX LDA SIGMA` | Label `TOOLONGX` is 8 chars, exceeds SIC 6-char limit. |
| 19 | `VALID1 MOVX ALPHA` | `MOVX` is not a valid SIC opcode. |
| 25 | `LDB ALPHA` | `LDB` is SIC/XE only, not in standard SIC. |
| 31 | `BAD1 BYTE C'` | Unclosed character quote — missing closing `'`. |
| 32 | `BAD2 BYTE C''` | Empty character constant. |
| 34 | `BAD4 BYTE X'F'` | Odd number of hex digits (1). Must be even. |
| 41 | `BAD9 BYTE Z'FF'` | Prefix `Z` invalid. Only `C` and `X` allowed. |
| 44 | `BAD1B BYTE 12345` | Bare integer — BYTE requires `C'...'` or `X'...'`. |
| 48 | `START` (no operand) | START requires one hex address. |
| 54 | `BAD14 WORD HELLO` | WORD expects decimal integer, got string. |
| 60 | `PROG2 START ZZZZ` | `ZZZZ` is not valid hexadecimal. |
| 63 | `PROG5 START FFFFF` | 0xFFFFF exceeds SIC 32,768-byte memory. |
| 66 | `LDA` (no operand) | LDA requires exactly one memory operand. |
| 72 | `END` (no operand) | END requires label of first executable instruction. |

---

## FILE 2: test_02_operands_strict.asm — 42 Errors

### Summary
Tests RSUB with operands (6), missing operands on all one-operand instructions (18),
explicit accumulator violations (6), malformed indexed addressing with invalid registers (8),
and I/O instructions missing operands (4).

| Category | Lines | Count | Error Type |
|----------|-------|-------|------------|
| RSUB with operands (zero-operand violation) | 14-19 | 6 | Syntax |
| Missing operands: load/store/arithmetic | 22-29 | 8 | Syntax |
| Missing operands: math/logic/index | 32-37 | 6 | Syntax |
| Missing operands: jump instructions | 40-43 | 4 | Syntax |
| Explicit accumulator `ADD A,ALPHA` | 46-51 | 6 | Syntax |
| Invalid register in indexed addr (,Y ,A ,B ,Z ,L ,S ,T) | 54-61 | 8 | Syntax |
| Missing I/O operands (RD/WD/TD/STSW) | 64-67 | 4 | Syntax |
| **Total** | | **42** | |

### Key Cases

| Line | Statement | Why It Fails |
|------|-----------|-------------|
| 14 | `RSUB ALPHA` | RSUB takes exactly 0 operands. |
| 19 | `RSUB ALPHA,X` | RSUB with indexed — doubly wrong. |
| 22 | `LDA` (no operand) | LDA requires one memory operand. |
| 46 | `ADD A,ALPHA` | In SIC, Accumulator is implicit. Use `ADD ALPHA`. |
| 54 | `LDCH ALPHA,Y` | Only `,X` is valid for indexed addressing in SIC. |
| 57 | `LDA ,X` | Missing base operand — just `,X` with no symbol. |
| 64 | `RD` (no operand) | RD needs a device address operand. |

---

## FILE 3: test_03_symbols_semantics.asm — ~20 Errors

### Summary
Tests duplicate labels (Pass 1) and undefined symbols (Pass 2). Also tests valid forward
references that must resolve correctly, and case-sensitive symbol lookup.

**Critical behavior**: The assembler runs Pass 2 even when Pass 1 finds errors, skipping
lines that already errored in Pass 1. This allows catching undefined symbols alongside
duplicate label errors.

| Category | Pass | Count | Error Type |
|----------|------|-------|------------|
| Duplicate labels (LOOP, FIRST, DATA, EXIT) | 1 | 4 | Semantic |
| Undefined symbols (XYZZY, NOTHERE, UNDEF1-3, NOLBL, BADREF, NOSUB) | 2 | 8 | Semantic |
| Undefined indexed symbols (UNDFX, UNDFBF, NOBUF, NADA) | 2 | 4 | Semantic |
| Case mismatch (Alpha, aLPHA) | 2 | 2 | Semantic |
| Undefined symbols (GHOST, SPOOKY) | 2 | 2 | Semantic |
| **Total** | | **~20** | |

### Key Cases

| Line | Statement | Pass | Why It Fails |
|------|-----------|------|-------------|
| 12 | `LOOP STA BETA` | 1 | `LOOP` was already defined on line 11. Duplicate. |
| 13 | `FIRST ADD ALPHA` | 1 | `FIRST` already defined on line 8. |
| 20 | `LDA XYZZY` | 2 | `XYZZY` is never defined anywhere. |
| 42 | `LDA UNDFX,X` | 2 | `UNDFX` not defined — indexed mode doesn't help. |
| 49 | `LDA Alpha` | 2 | `Alpha` ≠ `ALPHA` — SYMTAB is case-sensitive. |

### Valid Forward References (Must NOT Error)
Lines 30-34 reference FWDA-FWDE which are defined later. Pass 2 resolves these via SYMTAB.

---

## FILE 4: test_04_directives_memory.asm — 24 Errors

### Summary
Tests WORD bounds checking (24-bit range), RESW/RESB with zero/negative/overflow values,
BYTE character and hex edge cases, and memory limit violations.

| Category | Lines | Count | Error Type |
|----------|-------|-------|------------|
| WORD exceeds 24-bit range | 10-11 | 2 | Semantic |
| WORD non-integer operands | 14-18 | 5 | Lexical/Syntax |
| RESW zero/negative/overflow | 23-26 | 4 | Semantic/Memory |
| RESB zero/negative/overflow | 31-34 | 4 | Semantic/Memory |
| BYTE empty C constant / unclosed | 41-42 | 2 | Syntax |
| BYTE odd hex / bad hex / empty X | 48-50 | 3 | Syntax |
| BYTE bad format (no prefix, bare string, wrong prefix) | 53-56 | 4 | Syntax |
| **Total** | | **24** | |

### Key Cases

| Line | Statement | Why It Fails |
|------|-----------|-------------|
| 10 | `W5 WORD 8388608` | Exceeds +24-bit max (8,388,607). |
| 11 | `W6 WORD -8388609` | Below -24-bit min (-8,388,608). |
| 14 | `W7 WORD NOTANUM` | Non-integer string for WORD. |
| 23 | `R3 RESW 0` | Zero count — not positive. |
| 26 | `R6 RESW 11000` | 11000×3 = 33,000 bytes > 32,768 limit. |
| 34 | `B6 RESB 40000` | 40,000 > 32,768 SIC memory limit. |
| 48 | `X4 BYTE X'F'` | 1 hex digit — odd count. |

---

## FILE 5: test_05_perfect_execution.asm — 0 Errors (74 Valid Cases)

### Expected Behavior
This file must assemble with **zero errors**. Every line is valid standard SIC.

### Complete Instruction Coverage (All 25 Opcodes)
| Mnemonic | Opcode | Used At | Notes |
|----------|--------|---------|-------|
| ADD | 18 | V10, V34 | Arithmetic addition |
| AND | 40 | V16 | Bitwise AND |
| COMP | 28 | V14, V19 | Compare accumulator |
| DIV | 24 | V13 | Integer division |
| J | 3C | V23, V25, V27, V29, V32 | Unconditional jump |
| JEQ | 30 | V20, V38, V42 | Jump if equal |
| JGT | 34 | V21 | Jump if greater |
| JLT | 38 | V22, V49 | Jump if less than |
| JSUB | 48 | V31 | Jump to subroutine |
| LDA | 00 | V02, V09, V15, V24, V26, V28, V33 | Load accumulator |
| LDCH | 50 | V07, V43, V46 | Load character |
| LDL | 08 | V06, V51 | Load link register |
| LDX | 04 | V04, V45 | Load index register |
| MUL | 20 | V12 | Multiply |
| OR | 44 | V17 | Bitwise OR |
| RD | D8 | V39 | Read from device |
| RSUB | 4C | V36, V52 | Return from subroutine |
| STA | 0C | V03, V18, V30, V35 | Store accumulator |
| STCH | 54 | V08, V40, V47 | Store character |
| STL | 14 | V01 | Store link register |
| STSW | E8 | V50 | Store status word |
| STX | 10 | V05 | Store index register |
| SUB | 1C | V11 | Subtraction |
| TD | E0 | V37, V41 | Test device |
| TIX | 2C | V48 | Test and increment index |
| WD | DC | V44 | Write to device |

### Complete Directive Coverage
| Directive | Usage | Cases |
|-----------|-------|-------|
| START | `ALLSIC START 1000` | Sets load address 0x1000 |
| END | `END FIRST` | References first executable |
| WORD | ZERO through MAXLEN | 8 constants (V53-V59, V67) |
| BYTE C | `C'Z'`, `C'EOF'` | V68, V71 |
| BYTE X | `X'F1'`, `X'05'` | V69, V70 |
| RESW | ALPHA through LENGTH | 7 reservations (V60-V66) |
| RESB | BUFFER, OUBUF | V72, V73 (4096 bytes each) |

### Indexed Addressing (5 cases)
V08, V40, V43, V46, V47 — all use `,X` suffix.

### Forward References (7+ cases)
V01 (RETADR), V02 (FIVE), V20 (EQPATH), V21 (GTPATH), V22 (LTPATH), V31 (PROC1), V32 (IOMAIN) — all defined later, resolved in Pass 2.

### Expected Object Program
```
H^ALLSIC^001000^{program_length}
T^001000^{len}^{machine code...}
T^{addr}^{len}^{more code...}
... (text records break at RESW/RESB boundaries)
E^001000
```

---

## Error Type Distribution

| Error Type | Count | Files |
|------------|-------|-------|
| Syntax Error | ~80 | test_01, test_02 |
| Semantic Error | ~30 | test_03, test_04 |
| Lexical Error | ~15 | test_01, test_04 |
| Memory Error | ~8 | test_01, test_04 |
| **Valid cases** | **74** | test_05 |
| **Total errors** | **~129** | All error files |

## Architecture Note: Two-Pass Error Collection
The assembler always runs both passes. Pass 1 catches structural errors (bad labels,
bad opcodes, bad directives). Pass 2 catches semantic errors (undefined symbols, forward
reference failures). Lines with Pass 1 errors are skipped in Pass 2 to avoid duplicate
error reports. This ensures comprehensive coverage: you see ALL errors in a single run.
