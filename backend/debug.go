package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

// Opcode is a generic type definition for an opcode
type Opcode struct {
	Mnemonic string
	Operand1 string
	Operand2 string
	Addr     string
	Length   int
	Cycles   [1]int
}

func (op *Opcode) String() string {
	var buffer bytes.Buffer

	buffer.WriteString(op.Mnemonic)
	if op.Operand1 != "" {
		buffer.WriteString(" " + op.Operand1)
	}
	if op.Operand2 != "" {
		buffer.WriteString(" " + op.Operand2)
	}
	return buffer.String()
}

// DebugHarness contains all the necessary data for debugging the emulator
type DebugHarness struct {
	Unprefixed   map[byte]Opcode
	Cbprefixed   map[byte]Opcode
	ExercisedOps map[string]uint
}

// NewDebugHarness creates a new DebugHarness object
func NewDebugHarness() *DebugHarness {
	o := DebugHarness{}

	j := readOpcodesJSON("backend/opcodes.json")

	o.Unprefixed = map[byte]Opcode{}
	for _, v := range j["unprefixed"] {
		a, err := strconv.ParseUint(v.Addr, 0, 8)
		if err != nil {
			panic(err)
		}
		o.Unprefixed[byte(a)] = v
	}

	o.Cbprefixed = map[byte]Opcode{}
	for _, v := range j["cbprefixed"] {
		a, err := strconv.ParseUint(v.Addr, 0, 8)
		if err != nil {
			panic(err)
		}
		o.Cbprefixed[byte(a)] = v
	}

	o.ExercisedOps = make(map[string]uint)

	return &o
}

func (d *DebugHarness) PrintDebug(c *CPU) {
	var op Opcode
	if c.readMemory(c.PC) == 0xCB {
		op = d.Cbprefixed[c.readMemory(c.PC+1)]
	} else {
		op = d.Unprefixed[c.readMemory(c.PC)]
	}

	opStr := op.String()

	opStr = strings.Replace(opStr, "d8", fmt.Sprintf("0x%0.2X", c.readMemory(c.PC+1)), -1)
	opStr = strings.Replace(opStr, "a8", fmt.Sprintf("0x%0.2X", c.readMemory(c.PC+1)), -1)
	opStr = strings.Replace(opStr, "r8", fmt.Sprintf("0x%0.2X", c.readMemory(c.PC+1)), -1)
	opStr = strings.Replace(opStr, "d16", fmt.Sprintf("0x%0.2X%0.2X", c.readMemory(c.PC+2), c.readMemory(c.PC+1)), -1)
	opStr = strings.Replace(opStr, "a16", fmt.Sprintf("0x%0.2X%0.2X", c.readMemory(c.PC+2), c.readMemory(c.PC+1)), -1)
	opStr = strings.Replace(opStr, "(HL", fmt.Sprintf("(0x%0.4X", c.ReadHL()), -1)

	fmt.Printf("%20s | AF: 0x%0.4X | BC: 0x%0.4X | DE: 0x%0.4X | HL: 0x%0.4X | PC: 0x%0.4X\n",
		opStr, c.Readdouble(A, F), c.Readdouble(B, C), c.Readdouble(D, E), c.Readdouble(H, L), c.PC)
}

func readOpcodesJSON(filename string) map[string]map[string]Opcode {

	data, err0 := ioutil.ReadFile(filename)
	if err0 != nil {
		panic(err0)
	}

	var j map[string]map[string]Opcode
	err1 := json.Unmarshal(data, &j)
	if err1 != nil {
		panic(err1)
	}

	return j
}
