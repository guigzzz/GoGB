package backend

func (c *CPU) readMemory(address uint16) byte {
	if 0x4000 <= address && address < 0x8000 {

		offset := uint32(address) - 0x4000
		bankAddress := (uint32(c.selectedROMBank) * 0x4000) + offset
		return c.cartridgeROM[bankAddress]

	} else if 0xA000 <= address && address < 0xC000 {

		if c.selectedRAMBank == 0 {
			return c.ram[address]
		}

		offset := uint32(address) - 0xA000
		bankAddress := (uint32(c.selectedRAMBank) * 0x2000) + offset
		return c.cartridgeRAM[bankAddress]
	}
	return c.ram[address]
}

func (c *CPU) writeMemory(address uint16, value byte) {

	if 0x0000 <= address && address < 0x8000 {
		c.handleBankSwitching(address, value)

	} else if 0xA000 <= address && address < 0xC000 {
		if c.selectedRAMBank == 0 {
			c.ram[address] = value
		}

		offset := uint32(address) - 0xA000
		bankAddress := (uint32(c.selectedRAMBank) * 0x2000) + offset
		c.cartridgeRAM[bankAddress] = value

	} else {
		c.ram[address] = value
	}
}

func (c *CPU) handleBankSwitching(address uint16, value byte) {
	if 0x0000 <= address && address < 0x2000 {
		if value&0xA > 0 {
			c.cartrideRAMEnabled = true
		} else {
			c.cartrideRAMEnabled = false
		}
	} else if 0x2000 <= address && address < 0x4000 {
		value &= 0x1F // mask off lower 5 bits
		if value%0x20 == 0 {
			value++
		}
		c.selectedROMBank &= 0x60
		c.selectedROMBank |= value
	} else if 0x4000 <= address && address < 0x6000 {
		value &= 0x60
		if c.ROMMode {
			// in ROM mode
			c.selectedROMBank &= 0x1F
			c.selectedROMBank |= value
		} else {
			// in RAM mode
			c.selectedRAMBank = value >> 5
		}

	} else if 0x6000 <= address && address < 0x8000 {
		if value&1 == 0 {
			c.ROMMode = true
		} else {
			c.ROMMode = false
		}
	}
}
