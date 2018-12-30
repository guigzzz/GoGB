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

	if !c.IME || IF&IE == 0 {
		return
	}

	for n := uint16(0); n < 5; n++ {
		mask := byte(1 << n)
		if IF&IE&mask > 0 {
			c.IME = false

			c.ram[0xFF0F] &^= mask

			if c.haltMode == 0 || c.haltMode == 1 {
				// we are either not halted
				// or halted but will handle interrupt
				// either way PC points to next instruction
				c.pushPC()
				c.PC = handlerAddresses[n]
			}
			// if haltMode == 2 then don't handle interrupt,
			// skip to next instruction
			c.haltMode = 0

			return
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
