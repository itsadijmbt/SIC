package assembler

import "fmt"

// ═══════════════════════════════════════════════════════════════════════════════
// OPCODE TABLE - Standard Beck's SIC Architecture
// ═══════════════════════════════════════════════════════════════════════════════

var Optable = map[string]string{
	"ADD":  "18",
	"AND":  "40",
	"COMP": "28",
	"DIV":  "24",
	"J":    "3C",
	"JEQ":  "30",
	"JGT":  "34",
	"JLT":  "38",
	"JSUB": "48",
	"LDA":  "00",
	"LDCH": "50",
	"LDL":  "08",
	"LDX":  "04",
	"MUL":  "20",
	"OR":   "44",
	"RD":   "D8",
	"RSUB": "4C",
	"STA":  "0C",
	"STCH": "54",
	"STL":  "14",
	"STSW": "E8",
	"STX":  "10",
	"SUB":  "1C",
	"TD":   "E0",
	"TIX":  "2C",
	"WD":   "DC",
}

// ZeroOperandOps - instructions that take exactly 0 operands
var ZeroOperandOps = map[string]bool{
	"RSUB": true,
}

// OneOperandOps - instructions that take exactly 1 memory operand (accumulator is implicit)
var OneOperandOps = map[string]bool{
	"ADD": true, "AND": true, "COMP": true, "DIV": true,
	"J": true, "JEQ": true, "JGT": true, "JLT": true,
	"JSUB": true, "LDA": true, "LDCH": true, "LDL": true,
	"LDX": true, "MUL": true, "OR": true, "RD": true,
	"STA": true, "STCH": true, "STL": true, "STSW": true,
	"STX": true, "SUB": true, "TD": true, "TIX": true,
	"WD": true,
}

// IndexableOps - instructions that support indexed addressing (,X)
var IndexableOps = map[string]bool{
	"ADD": true, "AND": true, "COMP": true, "DIV": true,
	"J": true, "JEQ": true, "JGT": true, "JLT": true,
	"JSUB": true, "LDA": true, "LDCH": true, "LDL": true,
	"LDX": true, "MUL": true, "OR": true, "RD": true,
	"STA": true, "STCH": true, "STL": true, "STSW": true,
	"STX": true, "SUB": true, "TD": true, "TIX": true,
	"WD": true,
}

var PseudoInstructions = map[string]bool{
	"START": true,
	"END":   true,
	"RESW":  true,
	"WORD":  true,
	"BYTE":  true,
	"RESB":  true,
}

var Symtab = map[string]string{}
var Opcodes = []string{}
var X_register string = "00000"

// SIC memory limit: 2^15 = 32768 bytes
const SICMemoryLimit = 32768

// ═══════════════════════════════════════════════════════════════════════════════
// STRUCTURED ERROR TYPES
// ═══════════════════════════════════════════════════════════════════════════════

type ErrorType string

const (
	SyntaxError   ErrorType = "Syntax Error"
	SemanticError ErrorType = "Semantic Error"
	LexicalError  ErrorType = "Lexical Error"
	MemoryError   ErrorType = "Memory Error"
)

// AssemblyError is a structured, educational error for the TUI
type AssemblyError struct {
	Line      int
	Statement string
	Type      ErrorType
	Cause     string
	Fix       string
}

func (e *AssemblyError) Error() string {
	return fmt.Sprintf("[Line %d] %s: %s", e.Line, e.Type, e.Cause)
}

func NewError(line int, stmt string, errType ErrorType, cause, fix string) *AssemblyError {
	return &AssemblyError{
		Line:      line,
		Statement: stmt,
		Type:      errType,
		Cause:     cause,
		Fix:       fix,
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// PASS TRACE EVENTS (for educational TUI visualization)
// ═══════════════════════════════════════════════════════════════════════════════

type TraceEventKind string

const (
	TracePass1Start     TraceEventKind = "PASS1_START"
	TracePass1Line      TraceEventKind = "PASS1_LINE"
	TraceSymbolAdded    TraceEventKind = "SYMBOL_ADDED"
	TraceLocctrUpdate   TraceEventKind = "LOCCTR_UPDATE"
	TracePass1End       TraceEventKind = "PASS1_END"
	TracePass2Start     TraceEventKind = "PASS2_START"
	TracePass2Line      TraceEventKind = "PASS2_LINE"
	TraceOpcodeGenerate TraceEventKind = "OPCODE_GEN"
	TraceForwardRef     TraceEventKind = "FORWARD_REF"
	TraceObjectRecord   TraceEventKind = "OBJECT_RECORD"
	TracePass2End       TraceEventKind = "PASS2_END"
)

type TraceEvent struct {
	Kind    TraceEventKind
	Line    int
	Locctr  string
	Label   string
	Opcode  string
	Operand string
	Detail  string
}
