package backend

// CPU represents the current cpu state
type CPU struct {
	reg [8]byte
	SP  uint16 // stack pointer
	PC  uint16 // program counter
	IME bool   // interrupt master enable

	haltMode byte
	// 0 -> not halted,
	// 1 -> IME == true; stop executing until IE & IF > 0, then service interrupt
	// 2 -> IME == false; IE & IF == 0; stop executing until IE & IF > 0, then skip to next instruction
	// 3 -> IME == false; IE & IF > 0; HALT BUG

	cycleCounter uint64 // to count cycles

	mmu *MMU
	apu *APU

	debugger *DebugHarness
}

// NewCPU creates a new cpu struct
// also copies the bootrom into ram from 0x0000 to 0x00FF (256 bytes)
func NewCPU(debug bool, apu *APU, mmu *MMU) *CPU {
	c := new(CPU)

	c.apu = apu
	c.mmu = mmu

	c.writeMemory(0xFF40, 0x91)
	c.writeMemory(0xFF47, 0xFC)
	c.writeMemory(0xFF48, 0xFF)
	c.writeMemory(0xFF49, 0xFF)

	c.PC = 0x100
	c.SP = 0xFFFE

	c.Writedouble(A, F, 0x01B0)
	c.Writedouble(B, C, 0x0013)
	c.Writedouble(D, E, 0x00D8)
	c.Writedouble(H, L, 0x014D)

	c.haltMode = 0

	if debug {
		c.debugger = NewDebugHarness()
	}

	return c
}

func (c *CPU) readMemory(address uint16) byte {
	return c.mmu.readMemory(address)
}

func (c *CPU) writeMemory(address uint16, value byte) {
	c.mmu.writeMemory(address, value)
}

func (c *CPU) RunSync(allowance int) {
	var increment uint64
	for cycle := 0; cycle+int(increment) < allowance; cycle += int(increment) {
		if c.debugger != nil && c.haltMode == 0 {
			c.debugger.PrintDebug(c)
		}

		c.CheckAndHandleInterrupts()

		if c.haltMode == 0 {
			pcIncrement, cycleIncrement := c.DecodeAndExecuteNext()
			c.PC += uint16(pcIncrement)
			increment = uint64(cycleIncrement)
		} else {
			increment = 4
		}

		for i := 0; i < int(increment); i++ {
			c.apu.StepAPU()
		}

		for inc := 0; inc < int(increment); inc += 4 {
			c.cycleCounter += 4
			c.checkForTimerIncrementAndInterrupt()
		}
	}
}

// DecodeAndExecuteNext fetches next instruction from memory stored at PC
func (c *CPU) DecodeAndExecuteNext() (pcIncrement, cycleIncrement int) {
	op := c.readMemory(c.PC)
	oprow := (op & 0xF0) >> 4

	switch {
	case oprow <= 3:
		second := c.readMemory(c.PC + 1)
		third := c.readMemory(c.PC + 2)
		// various instructions
		return c.DecodeVariousUpper(op, second, third)
	case oprow <= 7:
		// various memory instructions
		return 1, c.DecodeMem(op)
	case oprow <= 11:
		// various ALU inctructions
		return 1, c.DecodeArith(op)
	default:
		second := c.readMemory(c.PC + 1)
		third := c.readMemory(c.PC + 2)
		// various instructions
		return c.DecodeVariousLower(op, second, third)
	}
}
