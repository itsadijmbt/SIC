package assembler

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
