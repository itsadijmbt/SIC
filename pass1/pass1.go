package pass1

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"sic-assembler/assembler"
	"strconv"
	"strings"
	"unicode"
)

// ═══════════════════════════════════════════════════════════════════════════════
// DATA STRUCTURES
// ═══════════════════════════════════════════════════════════════════════════════

type LineLocctr struct {
	LineNumber int
	Locctr     string
}

type RecordKey struct {
	RecordNumber int
	StartAddress string
}

// AssemblyResult holds the full output of a successful assembly
type AssemblyResult struct {
	Output        string
	Errors        []*assembler.AssemblyError
	Pass1Trace    []assembler.TraceEvent
	Pass2Trace    []assembler.TraceEvent
	SymbolTable   map[string]string
	ObjectProgram string
}

var locctrTable []LineLocctr
var progName string = " ADI'S"
var loadAddress string = "0000"

// ═══════════════════════════════════════════════════════════════════════════════
// CLEAR / RESET
// ═══════════════════════════════════════════════════════════════════════════════

func ClearTables() {
	locctrTable = nil
	assembler.Symtab = make(map[string]string)
	assembler.Opcodes = nil
	progName = " ADI'S"
	loadAddress = "0000"
}

// ═══════════════════════════════════════════════════════════════════════════════
// HEX HELPERS
// ═══════════════════════════════════════════════════════════════════════════════

func hexStringToInt(hexStr string) int {
	val, err := strconv.ParseInt(hexStr, 16, 64)
	if err != nil {
		return 0
	}
	return int(val)
}

func intToHexString(val int) string {
	return fmt.Sprintf("%04X", val)
}

func intToHexStringPass2TypeWord(val int) string {
	return fmt.Sprintf("%06X", val)
}

func intToHexStringPass2(val int) string {
	return fmt.Sprintf("%X", val)
}

func stringToHexNumber(str string) (int, error) {
	val, err := strconv.ParseInt(str, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid hex string '%s': %v", str, err)
	}
	return int(val), nil
}

func hexadecimalAdder(hex1 string, hex2 string) (string, error) {
	num1, err1 := stringToHexNumber(hex1)
	if err1 != nil {
		return "", err1
	}
	num2, err2 := stringToHexNumber(hex2)
	if err2 != nil {
		return "", err2
	}
	return intToHexStringPass2(num1 + num2), nil
}

// ═══════════════════════════════════════════════════════════════════════════════
// VALIDATION HELPERS
// ═══════════════════════════════════════════════════════════════════════════════

var validLabelRegex = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]{0,5}$`)
var validHexRegex = regexp.MustCompile(`^[0-9A-Fa-f]+$`)
var byteCharRegex = regexp.MustCompile(`^[Cc]'[^']*'$`)
var byteHexRegex = regexp.MustCompile(`^[Xx]'[0-9A-Fa-f]*'$`)

func isValidLabel(label string) bool {
	return validLabelRegex.MatchString(label)
}

func isValidHex(s string) bool {
	return validHexRegex.MatchString(s)
}

func hasIllegalChars(s string) bool {
	for _, ch := range s {
		if !unicode.IsPrint(ch) {
			return true
		}
	}
	return false
}

// ═══════════════════════════════════════════════════════════════════════════════
// MAIN ENTRY POINT
// ═══════════════════════════════════════════════════════════════════════════════

func RunPass1(filePath string) (string, error) {
	ClearTables()

	result := &AssemblyResult{
		Errors:      make([]*assembler.AssemblyError, 0),
		Pass1Trace:  make([]assembler.TraceEvent, 0),
		Pass2Trace:  make([]assembler.TraceEvent, 0),
		SymbolTable: make(map[string]string),
	}

	// ── PASS 1 ──
	err := runPass1Internal(filePath, result)
	if err != nil {
		return "", err
	}

	// Copy symtab snapshot
	for k, v := range assembler.Symtab {
		result.SymbolTable[k] = v
	}

	// ── Build set of lines that already have Pass 1 errors (avoid duplicates in Pass 2) ──
	pass1ErrorLines := make(map[int]bool)
	for _, e := range result.Errors {
		pass1ErrorLines[e.Line] = true
	}

	// ── PASS 2 — always run even if Pass 1 had errors, to catch undefined symbols ──
	var output bytes.Buffer
	err = runPass2Internal(filePath, &output, result, pass1ErrorLines)
	if err != nil {
		return "", err
	}
	if len(result.Errors) > 0 {
		return "", buildMultiError(result.Errors)
	}

	// ── OBJECT PROGRAM ──
	var objBuf bytes.Buffer
	HeaderAndTextRecord(&objBuf)

	// ── FORMAT FINAL OUTPUT ──
	var finalOut bytes.Buffer

	finalOut.WriteString("Symbol Table:\n")
	for label, address := range assembler.Symtab {
		finalOut.WriteString(fmt.Sprintf("  %-8s  %s\n", label, address))
	}
	finalOut.WriteString("\n")
	finalOut.WriteString(output.String())
	finalOut.WriteString("\n OBJECT PROGRAM: \n")
	finalOut.WriteString(objBuf.String())

	result.Output = finalOut.String()
	result.ObjectProgram = objBuf.String()

	return result.Output, nil
}

// RunPass1WithTrace returns the full AssemblyResult with trace events for the TUI
func RunPass1WithTrace(filePath string) (*AssemblyResult, error) {
	ClearTables()

	result := &AssemblyResult{
		Errors:      make([]*assembler.AssemblyError, 0),
		Pass1Trace:  make([]assembler.TraceEvent, 0),
		Pass2Trace:  make([]assembler.TraceEvent, 0),
		SymbolTable: make(map[string]string),
	}

	result.Pass1Trace = append(result.Pass1Trace, assembler.TraceEvent{
		Kind:   assembler.TracePass1Start,
		Detail: "Beginning Pass 1: Scanning source, assigning addresses, building Symbol Table (SYMTAB)",
	})

	err := runPass1Internal(filePath, result)
	if err != nil {
		return result, err
	}

	// Copy symtab snapshot
	for k, v := range assembler.Symtab {
		result.SymbolTable[k] = v
	}

	result.Pass1Trace = append(result.Pass1Trace, assembler.TraceEvent{
		Kind:   assembler.TracePass1End,
		Detail: fmt.Sprintf("Pass 1 complete. %d symbols defined. %d error(s).", len(assembler.Symtab), len(result.Errors)),
	})

	// ── Build set of lines that already have Pass 1 errors ──
	pass1ErrorLines := make(map[int]bool)
	for _, e := range result.Errors {
		pass1ErrorLines[e.Line] = true
	}

	// ── PASS 2 — always run to catch undefined symbols ──
	result.Pass2Trace = append(result.Pass2Trace, assembler.TraceEvent{
		Kind:   assembler.TracePass2Start,
		Detail: "Beginning Pass 2: Generating object code, resolving symbols via SYMTAB",
	})

	var output bytes.Buffer
	err = runPass2Internal(filePath, &output, result, pass1ErrorLines)
	if err != nil {
		return result, err
	}

	if len(result.Errors) > 0 {
		return result, buildMultiError(result.Errors)
	}

	var objBuf bytes.Buffer
	HeaderAndTextRecord(&objBuf)

	result.Pass2Trace = append(result.Pass2Trace, assembler.TraceEvent{
		Kind:   assembler.TracePass2End,
		Detail: "Pass 2 complete. Object program generated.",
	})

	var finalOut bytes.Buffer
	finalOut.WriteString("Symbol Table:\n")
	for label, address := range assembler.Symtab {
		finalOut.WriteString(fmt.Sprintf("  %-8s  %s\n", label, address))
	}
	finalOut.WriteString("\n")
	finalOut.WriteString(output.String())
	finalOut.WriteString("\n OBJECT PROGRAM: \n")
	finalOut.WriteString(objBuf.String())

	result.Output = finalOut.String()
	result.ObjectProgram = objBuf.String()

	return result, nil
}

func buildMultiError(errs []*assembler.AssemblyError) *MultiAssemblyError {
	return &MultiAssemblyError{Errors: errs}
}

type MultiAssemblyError struct {
	Errors []*assembler.AssemblyError
}

func (m *MultiAssemblyError) Error() string {
	var buf bytes.Buffer
	for _, e := range m.Errors {
		buf.WriteString(e.Error() + "\n")
	}
	return buf.String()
}

// ═══════════════════════════════════════════════════════════════════════════════
// PASS 1 INTERNALS
// ═══════════════════════════════════════════════════════════════════════════════

func runPass1Internal(filePath string, result *AssemblyResult) error {
	var locctr string = "0000"
	var linenumber int = 0

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("Error opening file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		linenumber++
		asmErr := processLinePass1(line, &locctr, linenumber, result)
		if asmErr != nil {
			result.Errors = append(result.Errors, asmErr)
		}
	}

	return nil
}

func registerOnSymtab(locctr *string, label string, line int, stmt string) *assembler.AssemblyError {
	if label == "" {
		return nil
	}
	if !isValidLabel(label) {
		return assembler.NewError(line, stmt, assembler.LexicalError,
			fmt.Sprintf("Invalid label '%s'. Labels must start with a letter, contain only alphanumerics, and be at most 6 characters.", label),
			"Rename the label to a valid SIC identifier (e.g., 'LOOP1', 'BUFR').",
		)
	}
	if _, exists := assembler.Symtab[label]; exists {
		return assembler.NewError(line, stmt, assembler.SemanticError,
			fmt.Sprintf("Duplicate symbol '%s' detected in the label field. This symbol was already defined earlier.", label),
			"Rename this label to a unique identifier that has not been used elsewhere in the program.",
		)
	}
	assembler.Symtab[label] = *locctr
	return nil
}

func processLinePass1(line string, locctr *string, linenumber int, result *AssemblyResult) *assembler.AssemblyError {
	origLine := line

	// Strip comments
	if commentIndex := strings.Index(line, ";"); commentIndex != -1 {
		line = line[:commentIndex]
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	// Check for non-printable characters
	if hasIllegalChars(line) {
		return assembler.NewError(linenumber, origLine, assembler.LexicalError,
			"Line contains non-printable or illegal characters.",
			"Remove any special/control characters from this line.",
		)
	}

	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}

	var label, ops, operand string

	// ── Parse fields ──
	// We need to handle various formats. The tricky case is distinguishing
	// "LABEL OP OPERAND" from "OP OPERAND" from "OP".
	// A token is a label if it's not an opcode/directive and we have 3+ tokens,
	// OR if it's 2 tokens and the second is RSUB (label RSUB).

	if len(parts) >= 4 {
		// Could be something like: LABEL ADD ALPHA , X (spaces around comma)
		// Try to rejoin operand parts
		label = parts[0]
		ops = strings.ToUpper(parts[1])
		operand = strings.Join(parts[2:], "")
		// Check for malformed indexed like "ALPHA , X" -> "ALPHA,X"
		operand = strings.ReplaceAll(operand, " ", "")
	} else if len(parts) == 3 {
		// Could be LABEL OP OPERAND or OP OPERAND,X (unlikely split)
		if _, isOp := assembler.Optable[strings.ToUpper(parts[0])]; isOp {
			// First token is an opcode: OP followed by maybe split operand
			ops = strings.ToUpper(parts[0])
			operand = parts[1] + parts[2]
			// Check if this is a malformed comma-X: "ADD ALPHA, X" where space split it
			if parts[1] == "," || strings.HasSuffix(parts[1], ",") || parts[2] == "X" {
				operand = strings.ReplaceAll(parts[1]+parts[2], " ", "")
			}
		} else if _, isPseudo := assembler.PseudoInstructions[strings.ToUpper(parts[0])]; isPseudo {
			ops = strings.ToUpper(parts[0])
			operand = strings.Join(parts[1:], " ")
		} else {
			label = parts[0]
			ops = strings.ToUpper(parts[1])
			operand = parts[2]
		}
	} else if len(parts) == 2 {
		upper0 := strings.ToUpper(parts[0])
		upper1 := strings.ToUpper(parts[1])

		if upper1 == "RSUB" {
			// LABEL RSUB
			label = parts[0]
			ops = "RSUB"
		} else if _, isOp := assembler.Optable[upper0]; isOp {
			ops = upper0
			operand = parts[1]
		} else if _, isPseudo := assembler.PseudoInstructions[upper0]; isPseudo {
			ops = upper0
			operand = parts[1]
		} else if upper0 == "RSUB" {
			// RSUB with an operand → error later
			ops = "RSUB"
			operand = parts[1]
		} else {
			// Assume LABEL OP (missing operand check later)
			label = parts[0]
			ops = upper1
		}
	} else if len(parts) == 1 {
		upper := strings.ToUpper(parts[0])
		ops = upper
	}

	ops = strings.ToUpper(ops)

	// ── Validate opcode/directive ──
	isOp := false
	if _, ok := assembler.Optable[ops]; ok {
		isOp = true
	}
	isPseudo := false
	if _, ok := assembler.PseudoInstructions[ops]; ok {
		isPseudo = true
	}

	if !isOp && !isPseudo {
		return assembler.NewError(linenumber, origLine, assembler.SyntaxError,
			fmt.Sprintf("Unrecognized mnemonic '%s'. This is not a valid SIC opcode or assembler directive.", ops),
			fmt.Sprintf("Use a valid SIC instruction (e.g., LDA, STA, ADD, J, JSUB, RSUB) or directive (START, END, BYTE, WORD, RESB, RESW)."),
		)
	}

	// ── HANDLE PSEUDO-INSTRUCTIONS ──
	if isPseudo {
		return processPseudoPass1(ops, operand, label, locctr, linenumber, origLine, result)
	}

	// ── HANDLE MACHINE INSTRUCTIONS ──
	return processMachinePass1(ops, operand, label, locctr, linenumber, origLine, result)
}

// ═══════════════════════════════════════════════════════════════════════════════
// PASS 1: PSEUDO-INSTRUCTION PROCESSING
// ═══════════════════════════════════════════════════════════════════════════════

func processPseudoPass1(ops, operand, label string, locctr *string, line int, stmt string, result *AssemblyResult) *assembler.AssemblyError {

	switch ops {
	case "START":
		operand = strings.TrimSpace(operand)
		if operand == "" {
			return assembler.NewError(line, stmt, assembler.SyntaxError,
				"START directive requires exactly one hexadecimal starting address.",
				"Provide a valid hex address, e.g., 'COPY START 1000'.",
			)
		}
		if !isValidHex(operand) {
			return assembler.NewError(line, stmt, assembler.LexicalError,
				fmt.Sprintf("Invalid hexadecimal address '%s' for START directive.", operand),
				"Use only valid hexadecimal digits (0-9, A-F), e.g., 'START 1000'.",
			)
		}
		addr := hexStringToInt(operand)
		if addr >= assembler.SICMemoryLimit {
			return assembler.NewError(line, stmt, assembler.MemoryError,
				fmt.Sprintf("START address 0x%s (%d) exceeds SIC memory limit of %d bytes.", operand, addr, assembler.SICMemoryLimit),
				fmt.Sprintf("Use a starting address less than %04X (%d).", assembler.SICMemoryLimit, assembler.SICMemoryLimit),
			)
		}
		*locctr = fmt.Sprintf("%04X", addr)
		progName = label
		loadAddress = *locctr
		locctrTable = append(locctrTable, LineLocctr{LineNumber: line, Locctr: *locctr})

		result.Pass1Trace = append(result.Pass1Trace, assembler.TraceEvent{
			Kind:    assembler.TracePass1Line,
			Line:    line,
			Locctr:  *locctr,
			Label:   label,
			Opcode:  "START",
			Operand: operand,
			Detail:  fmt.Sprintf("Program '%s' starts at address %s. LOCCTR initialized.", label, *locctr),
		})
		return nil

	case "END":
		if operand == "" {
			return assembler.NewError(line, stmt, assembler.SyntaxError,
				"END directive requires exactly one operand: the label of the first executable instruction.",
				"Provide a label, e.g., 'END FIRST'.",
			)
		}
		// END operand should be a label (validated in pass 2)
		if label != "" {
			if err := registerOnSymtab(locctr, label, line, stmt); err != nil {
				return err
			}
		}
		result.Pass1Trace = append(result.Pass1Trace, assembler.TraceEvent{
			Kind:    assembler.TracePass1Line,
			Line:    line,
			Locctr:  *locctr,
			Opcode:  "END",
			Operand: operand,
			Detail:  "END directive encountered. Pass 1 scanning terminates.",
		})
		return nil

	case "WORD":
		if err := registerOnSymtab(locctr, label, line, stmt); err != nil {
			return err
		}
		operand = strings.TrimSpace(operand)
		if operand == "" {
			return assembler.NewError(line, stmt, assembler.SyntaxError,
				"WORD directive requires exactly one decimal integer constant.",
				"Provide a decimal value, e.g., 'THREE WORD 3'.",
			)
		}
		val, parseErr := strconv.Atoi(operand)
		if parseErr != nil {
			return assembler.NewError(line, stmt, assembler.LexicalError,
				fmt.Sprintf("Invalid decimal integer '%s' for WORD directive.", operand),
				"WORD requires a valid decimal integer (e.g., 0, 5, -1).",
			)
		}
		if val > 8388607 || val < -8388608 {
			return assembler.NewError(line, stmt, assembler.SemanticError,
				fmt.Sprintf("WORD value %d exceeds the 24-bit (3-byte) range [-8388608, 8388607].", val),
				"Use a value within the 24-bit signed integer range.",
			)
		}

		currentLocctr := hexStringToInt(*locctr)
		currentLocctr += 3
		checkMemory(currentLocctr, line, stmt, result)
		*locctr = intToHexString(currentLocctr)
		locctrTable = append(locctrTable, LineLocctr{LineNumber: line, Locctr: *locctr})

		result.Pass1Trace = append(result.Pass1Trace, assembler.TraceEvent{
			Kind: assembler.TraceLocctrUpdate, Line: line, Locctr: *locctr, Label: label,
			Opcode: "WORD", Operand: operand,
			Detail: fmt.Sprintf("WORD reserves 3 bytes. LOCCTR → %s", *locctr),
		})
		if label != "" {
			result.Pass1Trace = append(result.Pass1Trace, assembler.TraceEvent{
				Kind: assembler.TraceSymbolAdded, Line: line, Label: label, Locctr: assembler.Symtab[label],
				Detail: fmt.Sprintf("Symbol '%s' added to SYMTAB at address %s", label, assembler.Symtab[label]),
			})
		}
		return nil

	case "RESW":
		if err := registerOnSymtab(locctr, label, line, stmt); err != nil {
			return err
		}
		operand = strings.TrimSpace(operand)
		if operand == "" {
			return assembler.NewError(line, stmt, assembler.SyntaxError,
				"RESW directive requires exactly one decimal integer specifying the number of words to reserve.",
				"Provide a count, e.g., 'BUFFER RESW 4'.",
			)
		}
		val, parseErr := strconv.Atoi(operand)
		if parseErr != nil {
			return assembler.NewError(line, stmt, assembler.LexicalError,
				fmt.Sprintf("Invalid decimal integer '%s' for RESW.", operand),
				"RESW requires a positive decimal integer (e.g., 4, 100).",
			)
		}
		if val <= 0 {
			return assembler.NewError(line, stmt, assembler.SemanticError,
				fmt.Sprintf("RESW count %d is not positive. Cannot reserve zero or negative words.", val),
				"Specify a positive integer for the number of words to reserve.",
			)
		}
		if val > 10922 { // 32768/3 ≈ 10922
			return assembler.NewError(line, stmt, assembler.MemoryError,
				fmt.Sprintf("RESW %d would reserve %d bytes, potentially exceeding SIC's 32KB memory.", val, val*3),
				"Reduce the reservation count to fit within SIC memory limits.",
			)
		}

		currentLocctr := hexStringToInt(*locctr)
		currentLocctr += 3 * val
		checkMemory(currentLocctr, line, stmt, result)
		*locctr = intToHexString(currentLocctr)
		locctrTable = append(locctrTable, LineLocctr{LineNumber: line, Locctr: *locctr})

		result.Pass1Trace = append(result.Pass1Trace, assembler.TraceEvent{
			Kind: assembler.TraceLocctrUpdate, Line: line, Locctr: *locctr, Label: label,
			Opcode: "RESW", Operand: operand,
			Detail: fmt.Sprintf("RESW %s reserves %d bytes (3×%s). LOCCTR → %s", operand, 3*val, operand, *locctr),
		})
		if label != "" {
			result.Pass1Trace = append(result.Pass1Trace, assembler.TraceEvent{
				Kind: assembler.TraceSymbolAdded, Line: line, Label: label, Locctr: assembler.Symtab[label],
				Detail: fmt.Sprintf("Symbol '%s' → %s", label, assembler.Symtab[label]),
			})
		}
		return nil

	case "RESB":
		if err := registerOnSymtab(locctr, label, line, stmt); err != nil {
			return err
		}
		operand = strings.TrimSpace(operand)
		if operand == "" {
			return assembler.NewError(line, stmt, assembler.SyntaxError,
				"RESB directive requires exactly one decimal integer specifying the number of bytes to reserve.",
				"Provide a count, e.g., 'BUFFER RESB 4096'.",
			)
		}
		val, parseErr := strconv.Atoi(operand)
		if parseErr != nil {
			return assembler.NewError(line, stmt, assembler.LexicalError,
				fmt.Sprintf("Invalid decimal integer '%s' for RESB.", operand),
				"RESB requires a positive decimal integer.",
			)
		}
		if val <= 0 {
			return assembler.NewError(line, stmt, assembler.SemanticError,
				fmt.Sprintf("RESB count %d is not positive.", val),
				"Specify a positive integer for the number of bytes to reserve.",
			)
		}
		if val > assembler.SICMemoryLimit {
			return assembler.NewError(line, stmt, assembler.MemoryError,
				fmt.Sprintf("RESB %d exceeds SIC's 32KB memory limit.", val),
				"Reduce the reservation count.",
			)
		}

		currentLocctr := hexStringToInt(*locctr)
		currentLocctr += val
		checkMemory(currentLocctr, line, stmt, result)
		*locctr = intToHexString(currentLocctr)
		locctrTable = append(locctrTable, LineLocctr{LineNumber: line, Locctr: *locctr})

		result.Pass1Trace = append(result.Pass1Trace, assembler.TraceEvent{
			Kind: assembler.TraceLocctrUpdate, Line: line, Locctr: *locctr, Label: label,
			Opcode: "RESB", Operand: operand,
			Detail: fmt.Sprintf("RESB reserves %d bytes. LOCCTR → %s", val, *locctr),
		})
		if label != "" {
			result.Pass1Trace = append(result.Pass1Trace, assembler.TraceEvent{
				Kind: assembler.TraceSymbolAdded, Line: line, Label: label, Locctr: assembler.Symtab[label],
				Detail: fmt.Sprintf("Symbol '%s' → %s", label, assembler.Symtab[label]),
			})
		}
		return nil

	case "BYTE":
		if err := registerOnSymtab(locctr, label, line, stmt); err != nil {
			return err
		}
		operand = strings.TrimSpace(operand)
		if operand == "" {
			return assembler.NewError(line, stmt, assembler.SyntaxError,
				"BYTE directive requires a character constant C'...' or hex constant X'...'.",
				"Use the format: BYTE C'EOF' or BYTE X'F1'.",
			)
		}

		currentLocctr := hexStringToInt(*locctr)

		if byteCharRegex.MatchString(operand) {
			// C'...'
			characters := operand[2 : len(operand)-1]
			if len(characters) == 0 {
				return assembler.NewError(line, stmt, assembler.SyntaxError,
					"BYTE C'...' constant must contain at least one character.",
					"Example: BYTE C'EOF'.",
				)
			}
			currentLocctr += len(characters)
		} else if byteHexRegex.MatchString(operand) {
			// X'...'
			hexValue := operand[2 : len(operand)-1]
			if len(hexValue) == 0 {
				return assembler.NewError(line, stmt, assembler.SyntaxError,
					"BYTE X'...' constant must contain at least one hex digit.",
					"Example: BYTE X'F1'.",
				)
			}
			if len(hexValue)%2 != 0 {
				return assembler.NewError(line, stmt, assembler.SyntaxError,
					fmt.Sprintf("BYTE X'%s' has an odd number of hex digits (%d). Hex constants must have an even number of digits.", hexValue, len(hexValue)),
					"Add a leading zero or correct the hex value (e.g., X'0F' instead of X'F').",
				)
			}
			currentLocctr += len(hexValue) / 2
		} else {
			return assembler.NewError(line, stmt, assembler.SyntaxError,
				fmt.Sprintf("Invalid BYTE operand '%s'. Must strictly match C'...' (character) or X'...' (hexadecimal) format.", operand),
				"Correct examples: BYTE C'EOF', BYTE X'F1'. Ensure matching quotes and valid prefix.",
			)
		}

		checkMemory(currentLocctr, line, stmt, result)
		*locctr = intToHexString(currentLocctr)
		locctrTable = append(locctrTable, LineLocctr{LineNumber: line, Locctr: *locctr})

		result.Pass1Trace = append(result.Pass1Trace, assembler.TraceEvent{
			Kind: assembler.TraceLocctrUpdate, Line: line, Locctr: *locctr, Label: label,
			Opcode: "BYTE", Operand: operand,
			Detail: fmt.Sprintf("BYTE constant. LOCCTR → %s", *locctr),
		})
		if label != "" {
			result.Pass1Trace = append(result.Pass1Trace, assembler.TraceEvent{
				Kind: assembler.TraceSymbolAdded, Line: line, Label: label, Locctr: assembler.Symtab[label],
				Detail: fmt.Sprintf("Symbol '%s' → %s", label, assembler.Symtab[label]),
			})
		}
		return nil
	}

	return nil
}

// ═══════════════════════════════════════════════════════════════════════════════
// PASS 1: MACHINE INSTRUCTION PROCESSING
// ═══════════════════════════════════════════════════════════════════════════════

func processMachinePass1(ops, operand, label string, locctr *string, line int, stmt string, result *AssemblyResult) *assembler.AssemblyError {

	// ── Operand count validation ──

	// Zero-operand instructions (RSUB)
	if assembler.ZeroOperandOps[ops] {
		if operand != "" {
			return assembler.NewError(line, stmt, assembler.SyntaxError,
				fmt.Sprintf("'%s' takes exactly 0 operands, but found operand '%s'.", ops, operand),
				fmt.Sprintf("Remove the operand. Correct usage: %s", ops),
			)
		}
	}

	// One-operand instructions
	if assembler.OneOperandOps[ops] {
		// Check for missing operand
		cleanedOperand := strings.TrimSuffix(strings.ToUpper(operand), ",X")
		cleanedOperand = strings.TrimSpace(cleanedOperand)

		if cleanedOperand == "" && !assembler.ZeroOperandOps[ops] {
			return assembler.NewError(line, stmt, assembler.SyntaxError,
				fmt.Sprintf("Missing operand for '%s'. This instruction requires exactly one memory operand.", ops),
				fmt.Sprintf("Provide a symbol or address, e.g., '%s ALPHA'.", ops),
			)
		}

		// Check for multiple operands (comma not followed by X → user tried explicit accumulator)
		if strings.Contains(operand, ",") {
			suffix := strings.ToUpper(strings.TrimSpace(operand[strings.Index(operand, ",")+1:]))
			if suffix != "X" {
				return assembler.NewError(line, stmt, assembler.SyntaxError,
					fmt.Sprintf("Invalid operand '%s' for '%s'. In standard SIC, the Accumulator (register A) is implicit.", operand, ops),
					fmt.Sprintf("Use a single memory operand. Example: '%s ALPHA'. The only valid comma usage is indexed addressing: '%s ALPHA,X'.", ops, ops),
				)
			}
		}
	}

	// ── Indexed addressing validation ──
	isIndexed := false
	if strings.HasSuffix(strings.ToUpper(operand), ",X") {
		isIndexed = true
		operand = strings.TrimSuffix(strings.ToUpper(operand), ",X")
		if !assembler.IndexableOps[ops] {
			return assembler.NewError(line, stmt, assembler.SyntaxError,
				fmt.Sprintf("Indexed addressing (,X) is not supported by '%s'.", ops),
				fmt.Sprintf("Remove ',X' from the operand."),
			)
		}
	}
	_ = isIndexed

	// ── Register symbol ──
	if err := registerOnSymtab(locctr, label, line, stmt); err != nil {
		return err
	}

	// ── Increment LOCCTR by 3 (all SIC machine instructions are 3 bytes) ──
	currentLocctr := hexStringToInt(*locctr)
	currentLocctr += 3
	checkMemory(currentLocctr, line, stmt, result)
	*locctr = intToHexString(currentLocctr)
	locctrTable = append(locctrTable, LineLocctr{LineNumber: line, Locctr: *locctr})

	result.Pass1Trace = append(result.Pass1Trace, assembler.TraceEvent{
		Kind: assembler.TracePass1Line, Line: line, Locctr: *locctr, Label: label,
		Opcode: ops, Operand: operand,
		Detail: fmt.Sprintf("Instruction '%s' occupies 3 bytes. LOCCTR → %s", ops, *locctr),
	})
	if label != "" {
		result.Pass1Trace = append(result.Pass1Trace, assembler.TraceEvent{
			Kind: assembler.TraceSymbolAdded, Line: line, Label: label, Locctr: assembler.Symtab[label],
			Detail: fmt.Sprintf("Symbol '%s' → %s", label, assembler.Symtab[label]),
		})
	}

	return nil
}

func checkMemory(addr int, line int, stmt string, result *AssemblyResult) {
	if addr > assembler.SICMemoryLimit {
		result.Errors = append(result.Errors, assembler.NewError(line, stmt, assembler.MemoryError,
			fmt.Sprintf("LOCCTR has reached 0x%04X (%d), exceeding SIC's 32,768-byte memory limit.", addr, addr),
			"Reduce program size or data reservations. SIC has only 32KB of memory (addresses 0x0000–0x7FFF).",
		))
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// PASS 2 INTERNALS
// ═══════════════════════════════════════════════════════════════════════════════

func runPass2Internal(filePath string, output *bytes.Buffer, result *AssemblyResult, skipLines map[int]bool) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("Error reopening file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineNum++
		// Skip lines that already had errors in Pass 1 (avoid duplicate errors)
		if skipLines != nil && skipLines[lineNum] {
			continue
		}
		asmErr := processLinePass2(line, lineNum, output, result)
		if asmErr != nil {
			result.Errors = append(result.Errors, asmErr)
		}
	}
	return nil
}

func processLinePass2(line string, lineNum int, output *bytes.Buffer, result *AssemblyResult) *assembler.AssemblyError {
	origLine := line

	if commentIndex := strings.Index(line, ";"); commentIndex != -1 {
		line = line[:commentIndex]
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}

	var label, ops, operand string

	if len(parts) >= 4 {
		label = parts[0]
		ops = strings.ToUpper(parts[1])
		operand = strings.Join(parts[2:], "")
		operand = strings.ReplaceAll(operand, " ", "")
	} else if len(parts) == 3 {
		if _, isOp := assembler.Optable[strings.ToUpper(parts[0])]; isOp {
			ops = strings.ToUpper(parts[0])
			operand = parts[1] + parts[2]
			if strings.HasSuffix(parts[1], ",") || parts[2] == "X" {
				operand = strings.ReplaceAll(parts[1]+parts[2], " ", "")
			}
		} else if _, isPseudo := assembler.PseudoInstructions[strings.ToUpper(parts[0])]; isPseudo {
			ops = strings.ToUpper(parts[0])
			operand = strings.Join(parts[1:], " ")
		} else {
			label = parts[0]
			ops = strings.ToUpper(parts[1])
			operand = parts[2]
		}
	} else if len(parts) == 2 {
		upper0 := strings.ToUpper(parts[0])
		upper1 := strings.ToUpper(parts[1])
		if upper1 == "RSUB" {
			label = parts[0]
			ops = "RSUB"
		} else if _, isOp := assembler.Optable[upper0]; isOp {
			ops = upper0
			operand = parts[1]
		} else if _, isPseudo := assembler.PseudoInstructions[upper0]; isPseudo {
			ops = upper0
			operand = parts[1]
		} else if upper0 == "RSUB" {
			ops = "RSUB"
			operand = parts[1]
		} else {
			label = parts[0]
			ops = upper1
		}
	} else if len(parts) == 1 {
		ops = strings.ToUpper(parts[0])
	}

	ops = strings.ToUpper(ops)
	_ = label

	// ── Handle pseudo-instructions ──
	if _, isPseudo := assembler.PseudoInstructions[ops]; isPseudo {
		if ops == "START" || ops == "END" {
			return nil
		}
	}

	// ── Index mode ──
	isIndexed := false
	if strings.HasSuffix(strings.ToUpper(operand), ",X") {
		isIndexed = true
		operand = strings.TrimSuffix(strings.ToUpper(operand), ",X")
	}

	// ── Generate object code ──

	switch ops {
	case "RSUB":
		opcode := assembler.Optable[ops] + "0000"
		output.WriteString(fmt.Sprintf("%s\t-\t%s\n", ops, opcode))
		assembler.Opcodes = append(assembler.Opcodes, opcode)
		result.Pass2Trace = append(result.Pass2Trace, assembler.TraceEvent{
			Kind: assembler.TraceOpcodeGenerate, Line: lineNum, Opcode: ops,
			Detail: fmt.Sprintf("RSUB → opcode %s + address 0000 = %s (return from subroutine)", assembler.Optable[ops], opcode),
		})
		return nil

	case "WORD":
		val, _ := strconv.Atoi(operand)
		wordValue := intToHexStringPass2TypeWord(val)
		output.WriteString(fmt.Sprintf("%s\t%s\t%s\n", ops, operand, wordValue))
		assembler.Opcodes = append(assembler.Opcodes, wordValue)
		result.Pass2Trace = append(result.Pass2Trace, assembler.TraceEvent{
			Kind: assembler.TraceOpcodeGenerate, Line: lineNum, Opcode: ops, Operand: operand,
			Detail: fmt.Sprintf("WORD %s → %s (24-bit integer constant)", operand, wordValue),
		})
		return nil

	case "RESW", "RESB":
		assembler.Opcodes = append(assembler.Opcodes, "xxxxxx")
		result.Pass2Trace = append(result.Pass2Trace, assembler.TraceEvent{
			Kind: assembler.TracePass2Line, Line: lineNum, Opcode: ops, Operand: operand,
			Detail: fmt.Sprintf("%s %s → no object code (storage reservation only)", ops, operand),
		})
		return nil

	case "BYTE":
		return processBytePass2(operand, lineNum, origLine, output, result)
	}

	// ── Machine instruction ──
	if _, exists := assembler.Optable[ops]; !exists {
		return assembler.NewError(lineNum, origLine, assembler.SyntaxError,
			fmt.Sprintf("Unknown operation '%s' in Pass 2.", ops),
			"This should have been caught in Pass 1.",
		)
	}

	// Look up operand in SYMTAB
	address, found := assembler.Symtab[operand]
	if !found {
		// Could be an absolute address (decimal or hex)
		if _, err := strconv.Atoi(operand); err == nil {
			val, _ := strconv.Atoi(operand)
			address = fmt.Sprintf("%04X", val)
			found = true
		} else if isValidHex(operand) {
			address = fmt.Sprintf("%04X", hexStringToInt(operand))
			found = true
		}
	}

	if !found {
		return assembler.NewError(lineNum, origLine, assembler.SemanticError,
			fmt.Sprintf("Undefined symbol '%s'. This label was never defined in the program.", operand),
			fmt.Sprintf("Define '%s' as a label on an instruction or data directive, or check for typos.", operand),
		)
	}

	result.Pass2Trace = append(result.Pass2Trace, assembler.TraceEvent{
		Kind: assembler.TraceForwardRef, Line: lineNum, Opcode: ops, Operand: operand,
		Detail: fmt.Sprintf("Symbol '%s' resolved to address %s via SYMTAB lookup", operand, address),
	})

	if isIndexed {
		addressHexString, hexErr := hexadecimalAdder(address, assembler.X_register)
		if hexErr != nil {
			return assembler.NewError(lineNum, origLine, assembler.SemanticError,
				fmt.Sprintf("Error computing indexed address for '%s': %v", operand, hexErr),
				"Check that the symbol resolves to a valid address.",
			)
		}

		firstBit := addressHexString[0:1]
		firstBitINT := hexStringToInt(firstBit)
		remainingBits := addressHexString[1:]

		if firstBitINT < 8 {
			firstBitINT += 8
			firstConvertedBit := intToHexStringPass2(firstBitINT)
			finalAddress := firstConvertedBit + remainingBits
			opcode := assembler.Optable[ops] + finalAddress
			output.WriteString(fmt.Sprintf("%s\t%s,X\t%s\n", ops, operand, opcode))
			assembler.Opcodes = append(assembler.Opcodes, opcode)
			result.Pass2Trace = append(result.Pass2Trace, assembler.TraceEvent{
				Kind: assembler.TraceOpcodeGenerate, Line: lineNum, Opcode: ops, Operand: operand + ",X",
				Detail: fmt.Sprintf("%s %s,X → opcode %s + indexed address %s = %s", ops, operand, assembler.Optable[ops], finalAddress, opcode),
			})
		} else {
			return assembler.NewError(lineNum, origLine, assembler.MemoryError,
				fmt.Sprintf("Cannot set indexed (X) bit: address %s is out of bounds (first nibble ≥ 8).", address),
				"The target address is too large for indexed addressing in SIC format.",
			)
		}
	} else {
		addressHexString, hexErr := hexadecimalAdder(address, "0000")
		if hexErr != nil {
			return assembler.NewError(lineNum, origLine, assembler.SemanticError,
				fmt.Sprintf("Error computing address for '%s': %v", operand, hexErr),
				"Check the symbol table entry.",
			)
		}

		// Pad to 4 hex digits
		for len(addressHexString) < 4 {
			addressHexString = "0" + addressHexString
		}

		firstBit := addressHexString[0:1]
		firstBitINT := hexStringToInt(firstBit)
		remainingBits := addressHexString[1:]

		if firstBitINT < 8 {
			opcode := assembler.Optable[ops] + firstBit + remainingBits
			output.WriteString(fmt.Sprintf("%s\t%s\t%s\n", ops, operand, opcode))
			assembler.Opcodes = append(assembler.Opcodes, opcode)
			result.Pass2Trace = append(result.Pass2Trace, assembler.TraceEvent{
				Kind: assembler.TraceOpcodeGenerate, Line: lineNum, Opcode: ops, Operand: operand,
				Detail: fmt.Sprintf("%s %s → opcode %s + address %s = %s", ops, operand, assembler.Optable[ops], addressHexString, opcode),
			})
		} else {
			return assembler.NewError(lineNum, origLine, assembler.MemoryError,
				fmt.Sprintf("Address %s is out of SIC addressing range (first nibble ≥ 8).", address),
				"The target address exceeds the 15-bit direct addressing limit of SIC.",
			)
		}
	}

	return nil
}

func processBytePass2(operand string, lineNum int, stmt string, output *bytes.Buffer, result *AssemblyResult) *assembler.AssemblyError {
	if len(operand) < 3 {
		return assembler.NewError(lineNum, stmt, assembler.SyntaxError,
			"Invalid BYTE operand in Pass 2.",
			"Use C'...' or X'...' format.",
		)
	}
	prefix := strings.ToUpper(operand[:1])

	if prefix == "C" {
		characters := operand[2 : len(operand)-1]
		var byteValues string
		for _, char := range characters {
			byteValues += fmt.Sprintf("%02X", char)
		}
		assembler.Opcodes = append(assembler.Opcodes, byteValues)
		output.WriteString(fmt.Sprintf("BYTE\t%s\t%s\n", operand, byteValues))
		result.Pass2Trace = append(result.Pass2Trace, assembler.TraceEvent{
			Kind: assembler.TraceOpcodeGenerate, Line: lineNum, Opcode: "BYTE", Operand: operand,
			Detail: fmt.Sprintf("BYTE C'%s' → ASCII hex %s", characters, byteValues),
		})
	} else if prefix == "X" {
		hexValue := operand[2 : len(operand)-1]
		assembler.Opcodes = append(assembler.Opcodes, hexValue)
		output.WriteString(fmt.Sprintf("BYTE\t%s\t%s\n", operand, hexValue))
		result.Pass2Trace = append(result.Pass2Trace, assembler.TraceEvent{
			Kind: assembler.TraceOpcodeGenerate, Line: lineNum, Opcode: "BYTE", Operand: operand,
			Detail: fmt.Sprintf("BYTE X'%s' → hex literal %s", hexValue, hexValue),
		})
	}
	return nil
}

// ═══════════════════════════════════════════════════════════════════════════════
// OBJECT PROGRAM GENERATION
// ═══════════════════════════════════════════════════════════════════════════════

func HeaderAndTextRecord(output *bytes.Buffer) {
	opCodes := assembler.Opcodes
	lineNum := 0

	record := make(map[RecordKey]string)
	space := 60
	recordNumber := 1

	if len(locctrTable) == 0 {
		return
	}

	startAddress := locctrTable[lineNum].Locctr

	for _, code := range opCodes {
		if lineNum >= len(locctrTable) {
			break
		}
		if space >= 6 && code != "xxxxxx" {
			key := RecordKey{recordNumber - 1, startAddress}
			record[key] += code
			lineNum++
			space -= 6
		} else if code == "xxxxxx" {
			recordNumber++
			lineNum++
			if lineNum < len(locctrTable) {
				startAddress = locctrTable[lineNum].Locctr
			}
			space = 60
		} else {
			recordNumber++
			if lineNum < len(locctrTable) {
				startAddress = locctrTable[lineNum].Locctr
			}
			key := RecordKey{recordNumber - 1, startAddress}
			record[key] = code
			lineNum++
			space = 60 - 6
		}
	}

	startAddrInt, _ := strconv.ParseInt(loadAddress, 16, 64)
	var lastAddr int64
	if len(locctrTable) > 0 {
		lastAddr, _ = strconv.ParseInt(locctrTable[len(locctrTable)-1].Locctr, 16, 64)
	}
	length := lastAddr - startAddrInt

	// Pad program name to 6 characters
	pName := progName
	for len(pName) < 6 {
		pName += " "
	}
	if len(pName) > 6 {
		pName = pName[:6]
	}

	output.WriteString(fmt.Sprintf("H^%-6s^%06X^%06X\n", pName, startAddrInt, length))

	for key, opcodes := range record {
		length := len(opcodes) / 2
		startAddr, _ := strconv.ParseInt(key.StartAddress, 16, 64)
		output.WriteString(fmt.Sprintf("T^%06X^%02X^%s\n", startAddr, length, opcodes))
	}

	firstExecAddr, _ := strconv.ParseInt(loadAddress, 16, 64)
	output.WriteString(fmt.Sprintf("E^%06X\n", firstExecAddr))
}
