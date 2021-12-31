package backend

// CPU represents the current cpu state
type CPU struct {
	reg [8]byte
	SP  uint16 // stack pointer
	PC  uint16 // program counter
	ram []byte // 64 KB ram
	IME bool   // interrupt master enable

	mbc MBC // memory bank controller

	KeyPressedMap map[string]bool

	haltMode byte
	// 0 -> not halted,
	// 1 -> IME == true; stop executing until IE & IF > 0, then service interrupt
	// 2 -> IME == false; IE & IF == 0; stop executing until IE & IF > 0, then skip to next instruction
	// 3 -> IME == false; IE & IF > 0; HALT BUG

	cycleCounter uint64 // to count cycles

	debugger *DebugHarness
	logger   Logger

	apu APU
}

type ApuFactory func(c *CPU) APU

// NewCPU creates a new cpu struct
// also copies the bootrom into ram from 0x0000 to 0x00FF (256 bytes)
func NewCPU(rom []byte, debug bool, logger Logger, apuFactory ApuFactory) *CPU {
	c := new(CPU)

	c.ram = make([]byte, 1<<16)

	c.mbc = getMemoryControllerFrom(rom)

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

	c.KeyPressedMap = map[string]bool{
		"up": false, "down": false, "left": false, "right": false,
		"A": false, "B": false, "start": false, "select": false,
	}

	c.haltMode = 0

	if debug {
		c.debugger = NewDebugHarness()
	}

	if logger == nil {
		c.logger = NewPrintLogger()
	} else {
		c.logger = logger
	}

	if apuFactory == nil {
		c.apu = NewAPU(c)
	} else {
		c.apu = apuFactory(c)
	}

	return c
}

func (c *CPU) GetAPU() APU {
	return c.apu
}

// NewTestCPU creates a barebone CPU specifically for tests
// designed to be fast
func NewTestCPU() *CPU {
	c := new(CPU)
	c.ram = make([]byte, 1<<16)

	c.mbc = NewMBC0(make([]byte, 1<<15))

	return c
}

func (c *CPU) RunSync(allowance int) {
	var increment uint64
	for cycle := 0; cycle+int(increment) < allowance; cycle += int(increment) {
		if c.debugger != nil && c.haltMode == 0 {
			c.debugger.PrintDebug(c)
		}

		c.CheckAndHandleInterrupts()

		var pcIncrement int
		var cycleIncrement int
		if c.haltMode == 0 {
			pcIncrement, cycleIncrement = c.DecodeAndExecuteNext()
			c.PC += uint16(pcIncrement)
		}

		if c.haltMode > 0 {
			increment = 4
		} else {
			increment = uint64(cycleIncrement)
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
