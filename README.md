# SIC Assembler TUI Visualizer v3.0

A comprehensive Terminal User Interface (TUI) visualizer for the standard Beck's SIC (Simplified Instructional Computer) two-pass assembler algorithm. Features robust error handling, strict instruction set validation, and educational pass-by-pass visualization.

## Features

### 1. Comprehensive Line-by-Line Error Handling
- Structured error reports with **Line Number**, **Error Type**, **Cause**, and **Fix**
- Error categories: `Syntax Error`, `Semantic Error`, `Lexical Error`, `Memory Error`
- Covers: duplicate symbols, undefined symbols, invalid opcodes, malformed operands, memory overflow

### 2. Strict Instruction Set Validation
- All 25 standard SIC opcodes validated (LDA, STA, ADD, SUB, COMP, J, JEQ, JGT, JLT, JSUB, RSUB, etc.)
- All 6 assembler directives enforced (START, END, BYTE, WORD, RESB, RESW)
- Label validation (alphanumeric, 1-6 chars, starts with letter)
- Memory bounds checking (32KB SIC limit)

### 3. Operand Count & Type Validation
- **RSUB**: Enforced as zero-operand (errors on `RSUB ALPHA`)
- **One-operand instructions**: Missing operand detection, implicit accumulator enforcement
- **Indexed addressing**: `,X` suffix validation, only on supported instructions
- **Directive-specific**: START (hex address), END (label), BYTE (C'...' or X'...'), WORD/RESB/RESW (decimal integer)

### 4. Internal Mechanics Visualization
- **Pass 1 Trace**: Step-by-step LOCCTR updates and SYMTAB population
- **Pass 2 Trace**: Opcode generation, SYMTAB lookups, forward reference resolution
- **Toggleable views**: Switch between Output, Pass 1, Pass 2, and Error views

## Keyboard Shortcuts

| Key | Action |
|-----------|-------------------------------|
| `Tab` | Navigate UI elements |
| `Enter` | Activate / Run |
| `Ctrl+R` | Run assembler |
| `Ctrl+L` | Clear output |
| `Ctrl+T` | Cycle theme |
| `Ctrl+H` | Help screen |
| `Ctrl+Q` | Quit |
| `F1` | Output view |
| `F2` | Pass 1 trace |
| `F3` | Pass 2 trace |
| `F4` | Error report |
| `Alt+1-4` | Alternative view switching |

## Building & Running

```bash
# Initialize and download dependencies
go mod tidy

# Build
go build -o sic-assembler .

# Run
./sic-assembler
```

Then enter a `.asm` file path and press `Ctrl+R` or click **▸ run**.

## Test Files

- `test_valid.asm` – Standard SIC copy program (should assemble successfully)
- `test_errors.asm` – Intentionally broken file to demonstrate error handling

## Project Structure

```
sic-assembler/
├── main.go              # TUI application (themes, views, layout)
├── assembler/
│   └── assembler.go     # Opcode tables, error types, trace events
├── pass1/
│   └── pass1.go         # Two-pass assembler engine with validation
├── test_valid.asm       # Sample valid SIC program
├── test_errors.asm      # Sample program with intentional errors
├── go.mod
└── README.md
```

## Themes

5 built-in themes: **Noir**, **Ocean**, **Forest**, **Sunset**, **Minimal**. Cycle with `Ctrl+T`.

## Architecture

The assembler implements the standard two-pass algorithm from Leland Beck's "System Software":

**Pass 1** scans source code to assign addresses (LOCCTR) and build the Symbol Table (SYMTAB).

**Pass 2** re-scans to generate object code by looking up opcodes in OPTAB and resolving symbols through SYMTAB, producing Header (H), Text (T), and End (E) records.
