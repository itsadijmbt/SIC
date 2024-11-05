package pass1

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"sic-assembler/assembler"
	"strconv"
	"strings"
)

// /************************************************

type LineLocctr struct {
	LineNumber int
	Locctr     string
}

type RecordKey struct {
	RecordNumber int
	StartAddress string
}

var locctrTable []LineLocctr

var progName string = " ADI'S"

var loadAddress string = "0000"

// dont forget to empty these
//************************************************//

func ClearTables() {
	locctrTable = nil
	assembler.Symtab = make(map[string]string)
	assembler.Opcodes = nil
	progName = " ADI'S"
	loadAddress = "0000"
}

func hexStringToInt(hexStr string) int {
	val, err := strconv.ParseInt(hexStr, 16, 64)
	if err != nil {
		fmt.Println("Error parsing hex:", err)
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
		return 0, fmt.Errorf("Error parsing hex string '%s': %v", str, err)
	}
	return int(val), nil
}

func hexadecimalAdder(hex1 string, hex2 string) (string, error) {

	num1, err1 := stringToHexNumber(hex1)
	if err1 != nil {
		return "", fmt.Errorf("Error converting hex1: %v", err1)
	}

	num2, err2 := stringToHexNumber(hex2)
	if err2 != nil {
		return "", fmt.Errorf("Error converting hex2: %v", err2)
	}

	sum := num1 + num2

	hexSum := intToHexStringPass2(sum)

	return hexSum, nil
}

func RunPass1(filePath string) (string, error) {
	var output bytes.Buffer
	var locctr string = "0000"
	var linenumber int = 0

	file, err := os.Open(filePath)
	if err != nil {
		output.WriteString(fmt.Sprintf("Error opening file: %v\n", err))
		return output.String(), err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		err := processLine(line, &locctr, &linenumber)
		if err != nil {
			output.WriteString(fmt.Sprintf("Error processing line %d: %v\n", linenumber, err))
			return output.String(), err
		}
	}

	output.WriteString("\n\nSymbol Table:\n")
	for label, address := range assembler.Symtab {
		output.WriteString(fmt.Sprintf("Label: %s Address: %s\n", label, address))
	}
	output.WriteString("\n")
	file.Close()
	file, err = os.Open(filePath)
	if err != nil {
		output.WriteString(fmt.Sprintf("Error reopening file: %v\n", err))
		return output.String(), err
	}
	defer file.Close()

	scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		err := processLineForPass2(line, &output)
		if err != nil {
			output.WriteString(fmt.Sprintf("Error processing line %d: %v\n", linenumber, err))
			return output.String(), err
		}
	}

	// output.WriteString("\n")
	// for _, add := range locctrTable {
	// 	output.WriteString(fmt.Sprintf("%d -  %s\n", add.LineNumber, add.Locctr))
	// }

	output.WriteString("\n OBJECT PROGRAM: \n")
	HeaderAndTextRecord(&output)

	ClearTables()
	return output.String(), nil
}

func registerOnSymtab(locctr *string, label string) error {
	if label != "" {
		if _, exists := assembler.Symtab[label]; exists {
			return fmt.Errorf("duplicate label found for label '%s'", label)
		} else {
			assembler.Symtab[label] = *locctr
		}
	}
	return nil
}

func processLine(line string, locctr *string, linenumber *int) error {
	// Remove comments from the line

	*linenumber++
	if commentIndex := strings.Index(line, ";"); commentIndex != -1 {
		line = line[:commentIndex]
	}

	line = strings.TrimSpace(line)
	parts := strings.Fields(line)

	if len(parts) == 0 || line == "" {
		return nil
	}

	var label, ops, operand string

	if len(parts) == 3 {
		label = parts[0]
		ops = parts[1]
		operand = parts[2]

	} else if len(parts) == 2 {

		if parts[1] == "RSUB" {
			ops = parts[1]
			if parts[0] != " " {
				label = parts[0]
			}
			operand = "0000"
		} else {

			ops = parts[0]
			operand = parts[1]
		}

	} else if len(parts) == 1 {
		ops = parts[0]
		if ops == "RSUB" {
			operand = "0000"
		}
	}

	if _, isPseudoCode := assembler.PseudoInstructions[ops]; isPseudoCode {

		if ops == "START" {
			*locctr = operand
			progName = label
			loadAddress = operand
			fmt.Printf("START found. Setting LOCCTR to: %s\n\n", *locctr)
			locctrTable = append(locctrTable, LineLocctr{LineNumber: *linenumber, Locctr: *locctr})
			return nil
		}

		// Handle END pseudo-instruction
		if ops == "END" {
			err := registerOnSymtab(locctr, label)
			if err != nil {
				return fmt.Errorf("Error at line %d: %v", *linenumber, err)
			}
			fmt.Println("End of assembler directives\n")
			return nil
		}
	}
	if strings.HasSuffix(operand, ",X") {
		operand = strings.TrimSuffix(operand, ",X")
	}

	err := registerOnSymtab(locctr, label)
	if err != nil {
		return fmt.Errorf("Error at line %d: %v", *linenumber, err)
	}

	if _, isValid := assembler.Optable[ops]; isValid {
		currentLocctr := hexStringToInt(*locctr)
		currentLocctr += 3
		*locctr = intToHexString(currentLocctr)
		locctrTable = append(locctrTable, LineLocctr{LineNumber: *linenumber, Locctr: *locctr})
		// fmt.Printf("Valid machine instruction: %s. Incrementing LOCCTR to: %s\n\n", ops, *locctr)

	} else if _, isValidPseudo := assembler.PseudoInstructions[ops]; isValidPseudo {

		if ops == "RESW" {
			operandInt, err := strconv.Atoi(operand)
			if err != nil {
				return fmt.Errorf("Error parsing operand '%s' at line %d: %v", operand, *linenumber, err)
			}

			currentLocctr := hexStringToInt(*locctr)
			currentLocctr += 3 * operandInt
			*locctr = intToHexString(currentLocctr)
			locctrTable = append(locctrTable, LineLocctr{LineNumber: *linenumber, Locctr: *locctr})
			// fmt.Printf("RESW found, incrementing LOCCTR by %d (hex: %s). New LOCCTR: %s\n", 3*operandInt, intToHexString(3*operandInt), *locctr)

		} else if ops == "WORD" {
			currentLocctr := hexStringToInt(*locctr)
			currentLocctr += 3
			*locctr = intToHexString(currentLocctr)
			locctrTable = append(locctrTable, LineLocctr{LineNumber: *linenumber, Locctr: *locctr})
			// fmt.Printf("WORD instruction found. LOCCTR incremented to: %s\n", *locctr)

		} else if ops == "RESB" {
			operandInt, err := strconv.Atoi(operand)
			if err != nil {
				return fmt.Errorf("Error parsing operand '%s' at line %d: %v", operand, *linenumber, err)
			}
			currentLocctr := hexStringToInt(*locctr)
			currentLocctr += operandInt
			*locctr = intToHexString(currentLocctr)
			locctrTable = append(locctrTable, LineLocctr{LineNumber: *linenumber, Locctr: *locctr})

		} else if ops == "BYTE" {
			currentLocctr := hexStringToInt(*locctr)

			if operand[:1] == "C" || operand[:1] == "c" {

				characters := operand[2 : len(operand)-1]
				var byteValues string
				for _, char := range characters {
					byteValues += fmt.Sprintf("%02X", char)
				}

				// fmt.Println("BYTE (Character Constant):", byteValues)
				currentLocctr += len(characters)

			} else if operand[:1] == "X" || operand[:1] == "x" {

				hexValue := operand[2 : len(operand)-1]

				if len(hexValue)%2 != 0 {

					return fmt.Errorf("Error: Incorrect size hex in operand '%s' at line %d", operand, *linenumber)
				}

				currentLocctr += len(hexValue) / 2
			}
			*locctr = intToHexString(currentLocctr)
			locctrTable = append(locctrTable, LineLocctr{LineNumber: *linenumber, Locctr: *locctr})
			// fmt.Printf("BYTE instruction found. LOCCTR incremented to: %s\n", *locctr)

		} else {
			return fmt.Errorf("Unknown pseudo-instruction: %s at line %d", ops, *linenumber)
		}
	} else {
		return fmt.Errorf("Unknown operation: %s at line %d", ops, *linenumber)
	}
	return nil
}

func processLineForPass2(line string, output *bytes.Buffer) error {

	if commentIndex := strings.Index(line, ";"); commentIndex != -1 {
		line = line[:commentIndex]
	}

	line = strings.TrimSpace(line)
	parts := strings.Fields(line)

	if len(parts) == 0 || line == "" {
		return nil
	}

	var ops, operand string

	if len(parts) == 3 {
		ops = parts[1]
		operand = parts[2]
	} else if len(parts) == 2 {
		if parts[1] == "RSUB" {
			ops = parts[1]
			operand = "0000"
		} else {
			ops = parts[0]
			operand = parts[1]
		}
	} else if len(parts) == 1 {
		ops = parts[0]
		if ops == "RSUB" {
			operand = "0000"
		}
	}

	if _, isPseudoCode := assembler.PseudoInstructions[ops]; isPseudoCode {
		if ops == "START" {
			fmt.Printf("START found\n\n")
			return nil
		}
		if ops == "END" {
			fmt.Println("End\n\n")
			return nil
		}
	}

	isIndexed := false
	if strings.HasSuffix(strings.ToUpper(operand), ",X") {
		isIndexed = true
		operand = strings.TrimSuffix(strings.ToUpper(operand), ",X")
	}

	if address, isValidLabel := assembler.Symtab[operand]; isValidLabel {

		switch ops {
		case "RSUB":

			opcode := assembler.Optable[ops] + "0000"
			output.WriteString(fmt.Sprintf("%s\t%s - \t%s\n", ops, operand, opcode))

			assembler.Opcodes = append(assembler.Opcodes, opcode)
			return nil

		case "WORD":

			wordValue := intToHexStringPass2TypeWord(hexStringToInt(operand))
			output.WriteString(fmt.Sprintf("%s\t%s - \t%s\n", ops, operand, wordValue))

			assembler.Opcodes = append(assembler.Opcodes, wordValue)
			return nil

		case "RESW", "RESB":
			assembler.Opcodes = append(assembler.Opcodes, "xxxxxx")
			return nil

		case "BYTE":
			if operand[:1] == "C" || operand[:1] == "c" {

				characters := operand[2 : len(operand)-1]
				var byteValues string
				for _, char := range characters {
					byteValues += fmt.Sprintf("%02X", char)
				}

				// fmt.Println("BYTE (Character Constant):", byteValues)
				assembler.Opcodes = append(assembler.Opcodes, byteValues)

			} else if operand[:1] == "X" || operand[:1] == "x" {

				hexValue := operand[2 : len(operand)-1]

				if len(hexValue)%2 != 0 {

					return fmt.Errorf("Error: Incorrect size hex in operand '%s'", operand)
				}
				assembler.Opcodes = append(assembler.Opcodes, hexValue)
				output.WriteString(fmt.Sprintf("%s\t%s - \t%s\n", ops, operand, hexValue))
			}

			return nil

		default:

			if _, exists := assembler.Optable[ops]; exists {

				if isIndexed {

					addressHexString, err := hexadecimalAdder(address, assembler.X_register)
					if err != nil {
						return fmt.Errorf("Error adding hex values: %v", err)
					}

					firstBit := addressHexString[0:1]
					firstBitINT := hexStringToInt(firstBit)
					remainingBits := addressHexString[1:]

					if firstBitINT < 8 {
						firstBitINT += 8
						firstConvertedBit := intToHexStringPass2(firstBitINT)
						finalUpdatedAddress := firstConvertedBit + remainingBits

						opcode := assembler.Optable[ops] + finalUpdatedAddress
						fmt.Println(ops, " \t", operand, " - \t", opcode)
						assembler.Opcodes = append(assembler.Opcodes, opcode)
					} else {
						return fmt.Errorf("OUT OF BOUNDS: Cannot set X bit for address %s", address)
					}

				} else {

					addressHexString, err := hexadecimalAdder(address, "0000")
					if err != nil {
						return fmt.Errorf("Error adding hex values: %v", err)
					}
					firstBit := addressHexString[0:1]
					firstBitINT := hexStringToInt(firstBit)
					remainingBits := addressHexString[1:]

					if firstBitINT < 8 {
						opcode := assembler.Optable[ops] + firstBit + remainingBits
						output.WriteString(fmt.Sprintf("%s\t%s - \t%s\n", ops, operand, opcode))

						assembler.Opcodes = append(assembler.Opcodes, opcode)
					} else {
						return fmt.Errorf("OUT OF BOUNDS: Cannot set X bit for address %s", address)
					}
				}

			} else {

				return fmt.Errorf("INVALID OPERATION: %s", ops)
			}
		}

	} else {
		switch ops {
		case "RSUB":

			opcode := assembler.Optable[ops] + "0000"
			output.WriteString(fmt.Sprintf("%s\t%s - \t%s\n", ops, operand, opcode))

			assembler.Opcodes = append(assembler.Opcodes, opcode)
			return nil

		case "WORD":

			wordValue := intToHexStringPass2TypeWord(hexStringToInt(operand))
			output.WriteString(fmt.Sprintf("%s\t%s - \t%s\n", ops, operand, wordValue))

			assembler.Opcodes = append(assembler.Opcodes, wordValue)
			return nil

		case "RESW", "RESB":

			assembler.Opcodes = append(assembler.Opcodes, "xxxxxx")
			return nil

		case "BYTE":

			if operand[:1] == "C" || operand[:1] == "c" {

				characters := operand[2 : len(operand)-1]
				var byteValues string
				for _, char := range characters {
					byteValues += fmt.Sprintf("%02X", char)
				}

				// fmt.Println("BYTE (Character Constant):", byteValues)
				assembler.Opcodes = append(assembler.Opcodes, byteValues)

			} else if operand[:1] == "X" || operand[:1] == "x" {

				hexValue := operand[2 : len(operand)-1]

				if len(hexValue)%2 != 0 {

					return fmt.Errorf("Error: Incorrect size hex in operand '%s'", operand)
				}
				assembler.Opcodes = append(assembler.Opcodes, hexValue)
				output.WriteString(fmt.Sprintf("%s\t%s - \t%s\n", ops, operand, hexValue))
			}

			return nil

		default:
			return fmt.Errorf("INVALID OPERATION: %s", ops)
		}
	}
	return nil
}

func HeaderAndTextRecord(output *bytes.Buffer) {
	opCodes := assembler.Opcodes
	lineNum := 0

	record := make(map[RecordKey]string)
	space := 60

	recordNumber := 1
	startAddress := locctrTable[lineNum].Locctr

	for _, code := range opCodes {
		if space >= 6 && code != "xxxxxx" {
			key := RecordKey{recordNumber - 1, startAddress}
			record[key] += code
			lineNum++
			space -= 6
		} else if code == "xxxxxx" {
			recordNumber++
			lineNum++
			startAddress = locctrTable[lineNum].Locctr
			space = 60
		} else {
			recordNumber++
			startAddress = locctrTable[lineNum].Locctr
			key := RecordKey{recordNumber - 1, startAddress}
			record[key] = code
			lineNum++
			space = 60 - 6
		}
	}

	startAddrInt, _ := strconv.ParseInt(loadAddress, 16, 64)
	lastAddrInt, _ := strconv.ParseInt(locctrTable[len(locctrTable)-1].Locctr, 16, 64)
	length := lastAddrInt - startAddrInt

	output.WriteString(fmt.Sprintf("\nH %s 00%s %06X\n", progName, loadAddress, length))

	for key, opcodes := range record {
		length := len(opcodes) / 2
		output.WriteString(fmt.Sprintf("T00%s^%02X^%s\n", key.StartAddress, length, opcodes))
	}

	output.WriteString(fmt.Sprintf("E 00%s\n", loadAddress))
}
