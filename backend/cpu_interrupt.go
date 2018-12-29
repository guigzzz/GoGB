package backend

var (
	handlerAddresses = [5]uint16{0x40, 0x48, 0x50, 0x58, 0x60}
)

// Bit 0: V-Blank  Interrupt Request (INT 40h)  (1=Request)
// Bit 1: LCD STAT Interrupt Request (INT 48h)  (1=Request)
// Bit 2: Timer    Interrupt Request (INT 50h)  (1=Request)
// Bit 3: Serial   Interrupt Request (INT 58h)  (1=Request)
// Bit 4: Joypad   Interrupt Request (INT 60h)  (1=Request)

func (c *CPU) getInterruptRequestRegister() byte {
	return c.ram[0xFF0F]
}

func (c *CPU) getInterruptEnableRegister() byte {
	return c.ram[0xFFFF]
}

func (c *CPU) CheckAndHandleInterrupts() {

	if !c.IME {
		return
	}

	IFReg := c.getInterruptRequestRegister()
	IEReg := c.getInterruptEnableRegister()

	for n := uint16(0); n < 5; n++ {
		mask := byte(1 << n)
		if IFReg&IEReg&mask > 0 {
			c.IME = false

			c.ram[0xFF0F] &^= mask

			c.pushPC()
			c.PC = handlerAddresses[n]
			return
		}
	}
}
