package backend

import "fmt"

type MBC1 struct {
	RamEnabled      bool
	SelectedROMBank byte
	SelectedRAMBank byte

	Rom []byte
	Ram []byte

	ROMMode bool
}

func NewMBC1(rom []byte, useRam, useBattery bool) *MBC1 {
	m := new(MBC1)

	m.SelectedROMBank = 1

	headerSize := getROMSize(rom[0x0148])
	if len(rom) != headerSize {
		panic(fmt.Sprintf(
			"Actual rom has size %d but header says it has size %d, this is inconsistent",
			len(rom), headerSize))
	}
	m.Rom = rom

	if useRam {
		m.Ram = make([]byte, getRAMSize(rom[0x0149]))
	} else {
		m.Ram = make([]byte, 0)
	}

	return m
}

func (m *MBC1) DelegateReadToMBC(address uint16) bool {
	return 0x0000 <= address && address < 0x8000 ||
		0xA000 <= address && address < 0xC000
}

func (m *MBC1) DelegateWriteToMBC(address uint16) bool {
	return 0x0000 <= address && address < 0x8000 ||
		0xA000 <= address && address < 0xC000
}

func (m *MBC1) ReadMemory(address uint16) byte {

	if 0x0000 <= address && address < 0x4000 {
		return m.Rom[address]
	}
	if 0x4000 <= address && address < 0x8000 {

		offset := uint32(address) - 0x4000
		bankAddress := (uint32(m.SelectedROMBank) * 0x4000) + offset
		return m.Rom[bankAddress]

	} else if 0xA000 <= address && address < 0xC000 {

		if m.RamEnabled && len(m.Ram) > 0 {
			offset := uint32(address) - 0xA000
			bankAddress := (uint32(m.SelectedRAMBank) * 0x2000) + offset
			return m.Ram[bankAddress]
		}

		return 0xFF
	}

	panic(fmt.Sprintf("Got unexpected read address not handled by MBC %d", address))
}

func (m *MBC1) WriteMemory(address uint16, value byte) {

	if 0x0000 <= address && address < 0x2000 {

		m.RamEnabled = value&0xA == 0xA

	} else if 0x2000 <= address && address < 0x4000 {
		value &= 0x1F // mask off lower 5 bits
		if value%0x20 == 0 {
			value++
		}
		m.SelectedROMBank &= 0x60
		m.SelectedROMBank |= value
	} else if 0x4000 <= address && address < 0x6000 {
		value &= 0x60
		if m.ROMMode {
			// in ROM mode
			m.SelectedROMBank &= 0x1F
			m.SelectedROMBank |= value

			// in ROM mode only RAM bank 0 can be used
			m.SelectedRAMBank = 0
		} else {
			// in RAM mode
			m.SelectedRAMBank = value >> 5

			// in RAM mode, only ROM banks 0x01-0x1F can be used
			m.SelectedROMBank &= 0x1F
		}

	} else if 0x6000 <= address && address < 0x8000 {

		m.ROMMode = value&1 == 0

	} else if 0xA000 <= address && address < 0xC000 {

		if m.RamEnabled && len(m.Ram) > 0 {
			offset := uint32(address) - 0xA000
			bankAddress := (uint32(m.SelectedRAMBank) * 0x2000) + offset
			m.Ram[bankAddress] = value
		}

	} else {
		panic(fmt.Sprintf("Got unexpected write address not handled by MBC %d", address))
	}
}
