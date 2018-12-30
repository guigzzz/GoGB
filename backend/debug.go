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

	o.ExercisedOps = make(map[string]uint)

	return o
}

// PrintDebug pretty prints debug information concerning the emulator status
func (d *DebugHarness) PrintDebug(c *CPU) {
	// b := c.ram[c.PC : c.PC+3]

	var op Opcode
	if c.readMemory(c.PC) == 0xCB {
		op = d.Cbprefixed[c.readMemory(c.PC+1)]
	} else {
		op = d.Unprefixed[c.readMemory(c.PC)]
	}

	// fmt.Sprintf("b: %6X, ", b),

	opStr := op.String()
	fmt.Println("Instruction:", opStr,
		fmt.Sprintf("[Length: %v, Cycles: %v]", op.Length, op.Cycles[0]))
	if op.Length == 2 {
		fmt.Printf("Value: 0x%0.2X\n", c.readMemory(c.PC+1))
	}
	if op.Length == 3 {
		fmt.Printf("Value: 0x%0.2X%0.2X\n",
			c.GetRAM()[c.PC+1], c.readMemory(c.PC+2))
	}
	fmt.Println("LY:", c.readMemory(0xFF44))
	fmt.Println(c.String())
}

func (d *DebugHarness) PrintDebugShort(c *CPU) {
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

	fmt.Printf("%s | PC: 0x%0.4X\n", opStr, c.PC)
}

func (d *DebugHarness) RecordNextExercisedOp(c *CPU) {
	var op Opcode
	if c.ram[c.PC] == 0xCB {
		op = d.Cbprefixed[c.ram[c.PC+1]]
	} else {
		op = d.Unprefixed[c.ram[c.PC]]
	}

	if _, ok := d.ExercisedOps[op.String()]; ok {
		d.ExercisedOps[op.String()]++
	} else {
		d.ExercisedOps[op.String()] = 1
	}
}

func (d *DebugHarness) GetExercicedOpSummary() {
	// for k, v := range d.ExercisedOps {
	// 	fmt.Println(k, v)
	// }
	fmt.Println("Instructions exerciced:")
	for _, v := range d.Unprefixed {
		if _, ok := d.ExercisedOps[v.String()]; ok {
			fmt.Println(v.String())
		}
	}
	// for _, v := range d.Unprefixed {
	// 	fmt.Println(v.String())
	// }
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
