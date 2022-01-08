package backend

import "fmt"

type MBC1 struct {
	ramEnabled      bool
	selectedROMBank byte
	selectedRAMBank byte

	rom []byte
	ram []byte

	ROMMode bool
}

func NewMBC1(rom []byte, useRam, useBattery bool) *MBC1 {
	m := new(MBC1)

	m.selectedROMBank = 1

	headerSize := getROMSize(rom[0x0148])
	if len(rom) != headerSize {
		panic(fmt.Sprintf(
			"Actual rom has size %d but header says it has size %d, this is inconsistent",
			len(rom), headerSize))
	}
	m.rom = rom

	if useRam {
		m.ram = make([]byte, getRAMSize(rom[0x0149]))
	} else {
		m.ram = make([]byte, 0)
	}

	return m
}

func (m *MBC1) ReadMemory(address uint16) byte {

	if 0x0000 <= address && address < 0x4000 {
		return m.rom[address]
	}
	if 0x4000 <= address && address < 0x8000 {

		offset := uint32(address) - 0x4000
		bankAddress := (uint32(m.selectedROMBank) * 0x4000) + offset
		return m.rom[bankAddress]

	} else if 0xA000 <= address && address < 0xC000 {

		if m.ramEnabled && len(m.ram) > 0 {
			offset := uint32(address) - 0xA000
			bankAddress := (uint32(m.selectedRAMBank) * 0x2000) + offset
			return m.ram[bankAddress]
		}

		return 0xFF
	}

	panic(fmt.Sprintf("Got unexpected read address not handled by MBC %d", address))
}

func (m *MBC1) WriteMemory(address uint16, value byte) {

	if 0x0000 <= address && address < 0x2000 {

		m.ramEnabled = value&0xA == 0xA

	} else if 0x2000 <= address && address < 0x4000 {
		value &= 0x1F // mask off lower 5 bits
		if value%0x20 == 0 {
			value++
		}
		m.selectedROMBank &= 0x60
		m.selectedROMBank |= value
	} else if 0x4000 <= address && address < 0x6000 {
		value &= 0x60
		if m.ROMMode {
			// in ROM mode
			m.selectedROMBank &= 0x1F
			m.selectedROMBank |= value

			// in ROM mode only RAM bank 0 can be used
			m.selectedRAMBank = 0
		} else {
			// in RAM mode
			m.selectedRAMBank = value >> 5

			// in RAM mode, only ROM banks 0x01-0x1F can be used
			m.selectedROMBank &= 0x1F
		}

	} else if 0x6000 <= address && address < 0x8000 {

		m.ROMMode = value&1 == 0

	} else if 0xA000 <= address && address < 0xC000 {

		if m.ramEnabled && len(m.ram) > 0 {
			offset := uint32(address) - 0xA000
			bankAddress := (uint32(m.selectedRAMBank) * 0x2000) + offset
			m.ram[bankAddress] = value
		}

	} else {
		panic(fmt.Sprintf("Got unexpected write address not handled by MBC %d", address))
	}
}
