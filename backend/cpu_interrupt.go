package backend

var (
	handlerAddresses = [5]uint16{0x40, 0x48, 0x50, 0x58, 0x60}
)

// Bit 0: V-Blank  Interrupt Request (INT 40h)  (1=Request)
// Bit 1: LCD STAT Interrupt Request (INT 48h)  (1=Request)
// Bit 2: Timer    Interrupt Request (INT 50h)  (1=Request)
// Bit 3: Serial   Interrupt Request (INT 58h)  (1=Request)
// Bit 4: Joypad   Interrupt Request (INT 60h)  (1=Request)

func (c *CPU) getInterruptRegisters() (byte, byte) {
	return c.ram[0xFF0F], c.ram[0xFFFF]
}

func (c *CPU) CheckAndHandleInterrupts() {

	IF, IE := c.getInterruptRegisters()

	if c.haltMode == 2 && IF&IE > 0 {
		c.haltMode = 0
		return
	}

	if !c.IME || IF&IE == 0 {
		return
	}

	for n := uint16(0); n < 5; n++ {
		mask := byte(1 << n)
		if IF&IE&mask > 0 {
			c.IME = false

			c.ram[0xFF0F] &^= mask

			// we are either not halted
			// or halted but will handle interrupt (i.e. mode 1)
			// either way PC points to next instruction
			c.pushPC()
			c.PC = handlerAddresses[n]

			// remove halted status
			c.haltMode = 0

			return
		}
	}
}

func (c *CPU) checkForTimerIncrementAndInterrupt(cycleIncrement uint64) {

	c.ram[0xFF04] = byte(c.cycleCounter >> 8) // div

	modulo := c.cycleCounter & (c.timerPeriod - 1)

	if c.timerPeriod == 0 {
		c.ram[0xFF05] = 0
		return
	} else if cycleIncrement < c.timerPeriod-modulo {
		return
	}

	if c.ram[0xFF05] == 0xFF {

		// write TMA into TIMA
		c.ram[0xFF05] = c.ram[0xFF06]

		// write to IF to signal interrupt
		c.ram[0xFF0F] |= 0x4
	} else {
		c.ram[0xFF05]++
	}
	if 2*c.timerPeriod-modulo < cycleIncrement {
		if c.ram[0xFF05] == 0xFF {
			// write TMA into TIMA
			c.ram[0xFF05] = c.ram[0xFF06]

			// write to IF to signal interrupt
			c.ram[0xFF0F] |= 0x4
		} else {
			c.ram[0xFF05]++
		}
	}
}

func (c *CPU) halt() {
	if c.IME {
		c.haltMode = 1
	} else {
		IF, IE := c.getInterruptRegisters()
		if IF&IE == 0 {
			c.haltMode = 2
		} else {
			// halt bug
			c.haltMode = 3
		}
	}
}
