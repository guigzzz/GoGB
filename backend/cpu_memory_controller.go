package backend

import (
	"fmt"
)

func (c *CPU) readMemory(address uint16) byte {

	if c.mbc.DelegateReadToMBC(address) {

		return c.mbc.ReadMemory(address)

	} else if 0xFEA0 <= address && address < 0xFF00 {
		return 00
	}
	return c.ram[address]
}

func (c *CPU) writeMemory(address uint16, value byte) {

	if c.mbc.DelegateWriteToMBC(address) {

		c.mbc.WriteMemory(address, value)

	} else if 0xFEA0 <= address && address < 0xFF00 {
		// ignore
	} else {
		if address == 0xFF02 && value == 0x81 {
			fmt.Print(string(c.ram[0xFF01]))
		} else if address == 0xFF46 {
			c.DMA(value)
			c.ram[0xFF46] = value
		} else if address == 0xFF00 {
			c.ram[0xFF00] = c.readKeyPressed(value)
		} else if address == 0xFF07 {
			// timer
			newValue := value & 0x7
			oldValue := c.ram[address]
			c.handleTimer(newValue, oldValue)
			c.ram[address] = newValue | 0xF8

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
	c.KeyPressedMapLock.RLock()
	defer c.KeyPressedMapLock.RUnlock()

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

func (c *CPU) handleTimer(newValue, oldValue byte) {
	// Bit 2 - Timer Enable (0=Disable, 1=Enable)
	// Bits 1-0 - Main Frequency Divider
	// 00: 4096 Hz (Increase every 1024 clocks)
	// 01: 262144 Hz ( “ “ 16 clocks)
	// 10: 65536 Hz ( “ “ 64 clocks)
	// 11: 16386 Hz ( “ “ 256 clocks)

	if newValue&0x4 == 0 {
		return
	}

	switch newValue & 0x3 {
	case 0:
		c.timerPeriod = 1024
	case 1:
		c.timerPeriod = 16
	case 2:
		c.timerPeriod = 64
	case 3:
		c.timerPeriod = 256
	}
}
