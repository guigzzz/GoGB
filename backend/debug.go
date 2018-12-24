package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
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
	Unprefixed map[byte]Opcode
	Cbprefixed map[byte]Opcode
}

// NewDebugHarness creates a new DebugHarness object
func NewDebugHarness() DebugHarness {
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

	return o
}

// PrintDebug pretty prints debug information concerning the emulator status
func (d *DebugHarness) PrintDebug(c CPU) {
	// b := c.ram[c.PC : c.PC+3]

	var op Opcode
	if c.ram[c.PC] == 0xCB {
		op = d.Cbprefixed[c.ram[c.PC+1]]
	} else {
		op = d.Unprefixed[c.ram[c.PC]]
	}

	// fmt.Sprintf("b: %6X, ", b),

	fmt.Println("Instruction:", op.String(),
		fmt.Sprintf("[Length: %v, Cycles: %v]", op.Length, op.Cycles[0]))
	fmt.Println("LY:", c.ram[0xFF44])
	fmt.Println(c.String())
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
