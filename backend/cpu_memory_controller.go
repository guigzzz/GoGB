package backend

var audioRegOrLookup = [32]byte{
	0x80, 0x3F, 0x00, 0xFF, 0xBF, // NR10-NR14
	0xFF, 0x3F, 0x00, 0xFF, 0xBF, // NR20-NR24
	0x7F, 0xFF, 0x9F, 0xFF, 0xBF, // NR30-NR34
	0xFF, 0xFF, 0x00, 0x00, 0xBF, // NR40-NR44
	0x00, 0x00, 0x70, // NR50-NR52
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // unused regs
}

func delegateReadToMBC(address uint16) bool {
	return 0x0000 <= address && address < 0x8000 ||
		0xA000 <= address && address < 0xC000
}

func delegateWriteToMBC(address uint16) bool {
	return 0x0000 <= address && address < 0x8000 ||
		0xA000 <= address && address < 0xC000
}

func (c *CPU) readMemory(address uint16) byte {

	if delegateReadToMBC(address) {

		return c.mbc.ReadMemory(address)

	} else if 0xFEA0 <= address && address < 0xFF00 {
		return 00
	} else if 0xFF10 <= address && address <= 0xFF2F {
		// audio regs
		or := audioRegOrLookup[address-0xFF10]
		return c.ram[address] | or
	}
	return c.ram[address]
}

func (c *CPU) writeMemory(address uint16, value byte) {

	if delegateWriteToMBC(address) {

		c.mbc.WriteMemory(address, value)

	} else if 0xFEA0 <= address && address < 0xFF00 {
		// ignore
	} else if 0xFF10 <= address && address <= 0xFF2F {
		oldValue := c.ram[address]
		// audio
		if address == 0xFF26 {
			// NR52 only top bits are writeable
			c.ram[address] = value & 0xF0

			// check if APU disabled. If yes, then clear all regs
			if value&0x80 == 0 {
				for i := 0xFF10; i <= 0xFF2F; i++ {
					c.ram[i] = 0
				}
			}

		} else {
			// ignore all writes if APU disabled
			isOn := c.ram[0xFF26]&0x80 > 0
			if isOn {
				c.ram[address] = value
			}
		}
		c.apu.AudioRegisterWriteCallback(address, oldValue, value)
	} else {
		if address == 0xFF02 && value == 0x81 {
			c.logger.Log(string(c.ram[0xFF01]))
		} else if address == 0xFF46 {
			c.DMA(value)
			c.ram[0xFF46] = value
		} else if address == 0xFF00 {
			c.ram[0xFF00] = 0b1100_0000 | (value & 0b11_0000) | c.readKeyPressed(value)
		} else if address == 0xFF04 {
			// when DIV is written
			// it is reset to 0
			c.ram[0xFF04] = 0
		} else {
			c.ram[address] = value
		}
	}
}

func (c *CPU) DMA(sourceAddress byte) {
	blockAddress := uint16(sourceAddress) << 8
	for i := uint16(0); i < 0xA0; i++ {
		c.ram[0xFE00+i] = c.ram[blockAddress+i]
	}
}

func (c *CPU) readKeyPressed(code byte) byte {
	regValue := byte(0xF)
	if code&0x20 == 0 { // 0b1101_1111
		if c.KeyPressedMap["start"] {
			regValue &^= 0x8
		}
		if c.KeyPressedMap["select"] {
			regValue &^= 0x4
		}
		if c.KeyPressedMap["B"] {
			regValue &^= 0x2
		}
		if c.KeyPressedMap["A"] {
			regValue &^= 0x1
		}
	}
	if code&0x10 == 0 { // 0b1110_1111
		if c.KeyPressedMap["down"] {
			regValue &^= 0x8
		}
		if c.KeyPressedMap["up"] {
			regValue &^= 0x4
		}
		if c.KeyPressedMap["left"] {
			regValue &^= 0x2
		}
		if c.KeyPressedMap["right"] {
			regValue &^= 0x1
		}
	}
	return regValue
}
